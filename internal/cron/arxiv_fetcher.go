package cron

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/KushnerykPavel/go-rag-arxiv/internal/client/arxiv"
	"github.com/KushnerykPavel/go-rag-arxiv/internal/wrappers"
)

var topicList = []string{
	"cs.AI",
	"cs.CL",
}

type fetcher interface {
	FetchPapers(ctx context.Context, params arxiv.FetchParams) ([]arxiv.Paper, error)
}

type notifier interface {
	SendHTML(ctx context.Context, chatID int64, html string) error
}

type ArxivFetcher struct {
	arxivClient fetcher
	notifier    notifier
	chatID      int64
	l           *zap.SugaredLogger
	limiter     *wrappers.RateLimiter
}

func NewArxivFetcher(
	arxivClient fetcher,
	notifier notifier,
	chatID int64,
	l *zap.SugaredLogger,
	limiter *wrappers.RateLimiter,
) *ArxivFetcher {
	return &ArxivFetcher{
		arxivClient: arxivClient,
		notifier:    notifier,
		chatID:      chatID,
		l:           l.With("module", "cron", "fetcher", "arxiv"),
		limiter:     limiter,
	}
}

func (f *ArxivFetcher) FetchPapers(ctx context.Context) {
	dateFrom := time.Now().Add(-24 * time.Hour).UTC().Format(arxiv.TimeFormat)
	dateTo := time.Now().UTC().Format(arxiv.TimeFormat)

	for _, topic := range topicList {
		params := arxiv.FetchParams{
			SearchCategory: topic,
			MaxResults:     2000,
			FromDate:       dateFrom,
			ToDate:         dateTo,
		}

		papers, err := f.arxivClient.FetchPapers(ctx, params)
		if err != nil {
			f.l.Errorw("failed to fetch papers",
				"topic", topic,
				"from", dateFrom,
				"to", dateTo,
				zap.Error(err),
			)
			continue
		}

		for _, paper := range papers {
			f.sendNotification(ctx, topic, paper)
		}
	}
}

func (f *ArxivFetcher) sendNotification(ctx context.Context, topic string, paper arxiv.Paper) {
	err := f.limiter.Do(ctx, func(ctx context.Context) error {
		return f.notifier.SendHTML(ctx, f.chatID, formatPaper(topic, paper))
	})
	if err != nil {
		f.l.Errorw("failed to send paper notification",
			"paper_id", paper.ArxivID,
			zap.Error(err),
		)
	}
}

func formatPaper(topic string, p arxiv.Paper) string {
	return fmt.Sprintf(
		"<b>[%s]</b> %s\n\n<b>Authors:</b> %s\n<b>Published:</b> %s\n<a href=\"%s\">PDF</a>",
		topic,
		p.Title,
		strings.Join(p.Authors, ", "),
		p.PublishedAt.Format(time.DateOnly),
		p.PDFURL,
	)
}
