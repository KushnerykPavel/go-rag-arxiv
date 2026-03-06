package arxiv

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

const TimeFormat = "20060102"

// FetchParams configures paper search parameters.
type FetchParams struct {
	MaxResults     int
	Start          int
	SortBy         string // submittedDate, lastUpdatedDate, relevance
	SortOrder      string // ascending, descending
	FromDate       string // YYYYMMDD format
	ToDate         string // YYYYMMDD format
	SearchCategory string
}

// Client communicates with the arXiv API.
type Client struct {
	cfg        Config
	log        *zap.SugaredLogger
	httpClient *http.Client

	mu            sync.Mutex
	lastRequestAt time.Time
}

// NewClient creates a new arXiv API client.
func NewClient(log *zap.SugaredLogger, opts ...Option) *Client {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	httpClient := cfg.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.timeout}
	}

	return &Client{
		cfg:        cfg,
		log:        log,
		httpClient: httpClient,
	}
}

// FetchPapers retrieves papers for the configured search category.
func (c *Client) FetchPapers(ctx context.Context, params FetchParams) ([]Paper, error) {
	query := "cat:" + params.SearchCategory

	if params.FromDate != "" || params.ToDate != "" {
		dateFrom := "*"
		if params.FromDate != "" {
			dateFrom = params.FromDate + "0000"
		}
		dateTo := "*"
		if params.ToDate != "" {
			dateTo = params.ToDate + "2359"
		}
		query += fmt.Sprintf(" AND submittedDate:[%s TO %s]", dateFrom, dateTo)
	}

	return c.fetchPapers(ctx, query, params)
}

// FetchPapersWithQuery retrieves papers using a custom arXiv search query.
//
// Example queries:
//   - "cat:cs.AI AND submittedDate:[20240101 TO *]"
//   - "au:LeCun AND cat:cs.AI"
//   - "ti:transformer AND cat:cs.AI"
func (c *Client) FetchPapersWithQuery(ctx context.Context, query string, params FetchParams) ([]Paper, error) {
	return c.fetchPapers(ctx, query, params)
}

// FetchPaperByID retrieves a single paper by its arXiv ID (e.g. "2507.17748v1").
// Returns nil without error if the paper is not found.
func (c *Client) FetchPaperByID(ctx context.Context, arxivID string) (*Paper, error) {
	// Strip optional version suffix: "2507.17748v1" → "2507.17748".
	cleanID, _, _ := strings.Cut(arxivID, "v")

	v := url.Values{}
	v.Set("id_list", cleanID)
	v.Set("max_results", "1")
	reqURL := c.cfg.baseURL + "?" + v.Encode()

	body, err := c.doGet(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("fetching paper %s: %w", arxivID, err)
	}

	papers, err := parseResponse(body)
	if err != nil {
		return nil, fmt.Errorf("parsing response for paper %s: %w", arxivID, err)
	}

	if len(papers) == 0 {
		c.log.Warnw("paper not found", "arxiv_id", arxivID)
		return nil, nil
	}

	return &papers[0], nil
}

// DownloadPDF downloads the PDF for a paper into the local cache directory.
// Returns the file path of the downloaded PDF. If forceDownload is false and
// a cached copy exists, the cached path is returned without re-downloading.
func (c *Client) DownloadPDF(ctx context.Context, paper Paper, forceDownload bool) (string, error) {
	if paper.PDFURL == "" {
		return "", fmt.Errorf("no PDF URL for paper %s", paper.ArxivID)
	}

	pdfPath := c.pdfPath(paper.ArxivID)

	if !forceDownload {
		if _, err := os.Stat(pdfPath); err == nil {
			c.log.Infow("using cached PDF", "file", filepath.Base(pdfPath))
			return pdfPath, nil
		}
	}

	if err := c.downloadWithRetry(ctx, paper.PDFURL, pdfPath); err != nil {
		return "", err
	}

	return pdfPath, nil
}

// --- internal helpers ---

func (c *Client) fetchPapers(ctx context.Context, searchQuery string, params FetchParams) ([]Paper, error) {
	maxResults := params.MaxResults
	if maxResults <= 0 {
		maxResults = c.cfg.maxResults
	}
	if maxResults > maxResultsCap {
		maxResults = maxResultsCap
	}

	sortBy := params.SortBy
	if sortBy == "" {
		sortBy = SortBySubmittedDate
	}
	sortOrder := params.SortOrder
	if sortOrder == "" {
		sortOrder = SortOrderDescending
	}

	reqURL := c.buildSearchURL(searchQuery, params.Start, maxResults, sortBy, sortOrder)

	c.log.Infow("fetching papers from arXiv",
		"query", searchQuery,
		"max_results", maxResults,
	)

	body, err := c.doGet(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("fetching papers: %w", err)
	}

	papers, err := parseResponse(body)
	if err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	c.log.Infow("fetched papers", "count", len(papers))
	return papers, nil
}

// buildSearchURL constructs the arXiv API query URL, keeping :, +, [, ]
// characters unescaped in search_query as required by the arXiv API.
func (c *Client) buildSearchURL(searchQuery string, start, maxResults int, sortBy, sortOrder string) string {
	v := url.Values{}
	v.Set("start", strconv.Itoa(start))
	v.Set("max_results", strconv.Itoa(maxResults))
	v.Set("sortBy", sortBy)
	v.Set("sortOrder", sortOrder)

	escapedQuery := strings.ReplaceAll(searchQuery, " ", "+")
	return fmt.Sprintf("%s?search_query=%s&%s", c.cfg.baseURL, escapedQuery, v.Encode())
}

func (c *Client) doGet(ctx context.Context, reqURL string) ([]byte, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from arXiv API", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return body, nil
}

// waitForRateLimit enforces the minimum delay between requests recommended by arXiv (3s).
func (c *Client) waitForRateLimit(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.lastRequestAt.IsZero() {
		if wait := c.cfg.rateLimitDelay - time.Since(c.lastRequestAt); wait > 0 {
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	c.lastRequestAt = time.Now()
	return nil
}

func (c *Client) pdfPath(arxivID string) string {
	safe := strings.ReplaceAll(arxivID, "/", "_") + ".pdf"
	return filepath.Join(c.cfg.pdfCacheDir, safe)
}

func (c *Client) downloadWithRetry(ctx context.Context, pdfURL, destPath string) error {
	if err := os.MkdirAll(c.cfg.pdfCacheDir, 0o755); err != nil {
		return fmt.Errorf("creating PDF cache directory: %w", err)
	}

	c.log.Infow("downloading PDF", "url", pdfURL)

	if err := c.waitForRateLimit(ctx); err != nil {
		return err
	}

	var lastErr error
	for attempt := range c.cfg.downloadMaxRetries {
		if err := c.downloadFile(ctx, pdfURL, destPath); err != nil {
			lastErr = err
			if attempt < c.cfg.downloadMaxRetries-1 {
				wait := c.cfg.downloadRetryDelayBase * time.Duration(attempt+1)
				c.log.Warnw("PDF download failed, retrying",
					"attempt", attempt+1,
					"max_retries", c.cfg.downloadMaxRetries,
					"error", err,
					"retry_in", wait,
				)
				select {
				case <-time.After(wait):
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		} else {
			c.log.Infow("PDF downloaded", "file", filepath.Base(destPath))
			return nil
		}
	}

	_ = os.Remove(destPath)
	return fmt.Errorf("PDF download failed after %d attempts: %w", c.cfg.downloadMaxRetries, lastErr)
}

func (c *Client) downloadFile(ctx context.Context, fileURL, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

// --- XML parsing ---

type atomFeed struct {
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	ID         string         `xml:"id"`
	Title      string         `xml:"title"`
	Summary    string         `xml:"summary"`
	Published  string         `xml:"published"`
	Authors    []atomAuthor   `xml:"author"`
	Categories []atomCategory `xml:"category"`
	Links      []atomLink     `xml:"link"`
}

type atomAuthor struct {
	Name string `xml:"name"`
}

type atomCategory struct {
	Term string `xml:"term,attr"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Type string `xml:"type,attr"`
}

func parseResponse(data []byte) ([]Paper, error) {
	var feed atomFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("unmarshaling arXiv XML: %w", err)
	}

	papers := make([]Paper, 0, len(feed.Entries))
	for _, entry := range feed.Entries {
		if paper, ok := parseSingleEntry(entry); ok {
			papers = append(papers, paper)
		}
	}
	return papers, nil
}

func parseSingleEntry(entry atomEntry) (Paper, bool) {
	arxivID := extractArxivID(entry.ID)
	if arxivID == "" {
		return Paper{}, false
	}

	authors := make([]string, 0, len(entry.Authors))
	for _, a := range entry.Authors {
		if name := strings.TrimSpace(a.Name); name != "" {
			authors = append(authors, name)
		}
	}

	categories := make([]string, 0, len(entry.Categories))
	for _, cat := range entry.Categories {
		if cat.Term != "" {
			categories = append(categories, cat.Term)
		}
	}

	publishedAt, _ := time.Parse(time.RFC3339, strings.TrimSpace(entry.Published))

	return Paper{
		ArxivID:     arxivID,
		Title:       cleanText(entry.Title),
		Authors:     authors,
		Abstract:    cleanText(entry.Summary),
		PublishedAt: publishedAt,
		Categories:  categories,
		PDFURL:      extractPDFURL(entry.Links),
	}, true
}

func extractArxivID(idURL string) string {
	idx := strings.LastIndex(idURL, "/")
	if idx == -1 || idx == len(idURL)-1 {
		return ""
	}
	return idURL[idx+1:]
}

func extractPDFURL(links []atomLink) string {
	for _, link := range links {
		if link.Type == "application/pdf" {
			u := link.Href
			if strings.HasPrefix(u, "http://arxiv.org/") {
				u = strings.Replace(u, "http://arxiv.org/", "https://arxiv.org/", 1)
			}
			return u
		}
	}
	return ""
}

func cleanText(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
