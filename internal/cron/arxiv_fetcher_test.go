package cron

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/KushnerykPavel/go-rag-arxiv/internal/client/arxiv"
	"github.com/KushnerykPavel/go-rag-arxiv/internal/wrappers"
)

type fakeFetcher struct {
	papersByCategory map[string][]arxiv.Paper
}

func (f fakeFetcher) FetchPapers(ctx context.Context, params arxiv.FetchParams) ([]arxiv.Paper, error) {
	if papers, ok := f.papersByCategory[params.SearchCategory]; ok {
		return papers, nil
	}
	return nil, nil
}

type fakeNotifier struct {
	calls int
	htmls []string
}

func (f *fakeNotifier) SendHTML(ctx context.Context, chatID int64, html string) error {
	f.calls++
	f.htmls = append(f.htmls, html)
	return nil
}

func TestFetchPapersFiltersBeforeSend(t *testing.T) {
	limiter, err := wrappers.NewRateLimiter(1)
	if err != nil {
		t.Fatalf("NewRateLimiter() error = %v", err)
	}

	eligiblePaper := arxiv.Paper{
		Title:       "Survey of AI",
		Abstract:    "overview",
		Categories:  []string{"cs.AI"},
		Authors:     []string{"Ada"},
		PublishedAt: time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC),
		PDFURL:      "https://arxiv.org/pdf/1234.56789.pdf",
	}
	ineligiblePaper := arxiv.Paper{
		Title:      "On Transformers",
		Abstract:   "method",
		Categories: []string{"cs.AI"},
	}

	fetcher := fakeFetcher{
		papersByCategory: map[string][]arxiv.Paper{
			"cs.AI": {eligiblePaper, ineligiblePaper},
			"cs.CL": {},
		},
	}
	notifier := &fakeNotifier{}

	arxivFetcher := NewArxivFetcher(fetcher, notifier, 12345, zap.NewNop().Sugar(), limiter)
	arxivFetcher.FetchPapers(context.Background())

	if notifier.calls != 1 {
		t.Fatalf("SendHTML calls = %d, want 1", notifier.calls)
	}

	if len(notifier.htmls) != 1 {
		t.Fatalf("SendHTML html count = %d, want 1", len(notifier.htmls))
	}

	wantHTML := formatPaper("cs.AI", eligiblePaper)
	if notifier.htmls[0] != wantHTML {
		t.Fatalf("SendHTML html = %q, want %q", notifier.htmls[0], wantHTML)
	}
}

func TestFormatPaperUnchanged(t *testing.T) {
	paper := arxiv.Paper{
		Title:       "Survey of AI",
		Authors:     []string{"Ada Lovelace", "Alan Turing"},
		PublishedAt: time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC),
		PDFURL:      "https://arxiv.org/pdf/1234.56789.pdf",
	}

	want := "<b>[cs.AI]</b> Survey of AI\n\n<b>Authors:</b> Ada Lovelace, Alan Turing\n<b>Published:</b> 2026-03-30\n<a href=\"https://arxiv.org/pdf/1234.56789.pdf\">PDF</a>"
	if got := formatPaper("cs.AI", paper); got != want {
		t.Fatalf("formatPaper() = %q, want %q", got, want)
	}
}
