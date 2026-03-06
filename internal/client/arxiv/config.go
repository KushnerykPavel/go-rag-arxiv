package arxiv

import (
	"net/http"
	"time"
)

const (
	SortBySubmittedDate   = "submittedDate"
	SortByLastUpdatedDate = "lastUpdatedDate"
	SortByRelevance       = "relevance"

	SortOrderAscending  = "ascending"
	SortOrderDescending = "descending"

	defaultBaseURL                = "https://export.arxiv.org/api/query"
	defaultRateLimitDelay         = 3 * time.Second
	defaultTimeout                = 30 * time.Second
	defaultMaxResults             = 100
	defaultSearchCategory         = "cs.AI"
	defaultPDFCacheDir            = ".cache/pdfs"
	defaultDownloadMaxRetries     = 3
	defaultDownloadRetryDelayBase = 5 * time.Second

	maxResultsCap = 2000
)

// Config holds configuration for the arXiv API client.
type Config struct {
	baseURL                string
	rateLimitDelay         time.Duration
	timeout                time.Duration
	maxResults             int
	searchCategory         string
	pdfCacheDir            string
	downloadMaxRetries     int
	downloadRetryDelayBase time.Duration
	httpClient             *http.Client
}

// Option configures the arXiv API client.
type Option func(*Config)

func WithBaseURL(url string) Option {
	return func(c *Config) { c.baseURL = url }
}

func WithRateLimitDelay(d time.Duration) Option {
	return func(c *Config) { c.rateLimitDelay = d }
}

func WithTimeout(d time.Duration) Option {
	return func(c *Config) { c.timeout = d }
}

func WithMaxResults(n int) Option {
	return func(c *Config) { c.maxResults = n }
}

func WithSearchCategory(category string) Option {
	return func(c *Config) { c.searchCategory = category }
}

func WithPDFCacheDir(dir string) Option {
	return func(c *Config) { c.pdfCacheDir = dir }
}

func WithDownloadMaxRetries(n int) Option {
	return func(c *Config) { c.downloadMaxRetries = n }
}

func WithDownloadRetryDelayBase(d time.Duration) Option {
	return func(c *Config) { c.downloadRetryDelayBase = d }
}

func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) { c.httpClient = client }
}

func defaultConfig() Config {
	return Config{
		baseURL:                defaultBaseURL,
		rateLimitDelay:         defaultRateLimitDelay,
		timeout:                defaultTimeout,
		maxResults:             defaultMaxResults,
		searchCategory:         defaultSearchCategory,
		pdfCacheDir:            defaultPDFCacheDir,
		downloadMaxRetries:     defaultDownloadMaxRetries,
		downloadRetryDelayBase: defaultDownloadRetryDelayBase,
	}
}
