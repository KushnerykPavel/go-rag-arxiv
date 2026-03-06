package arxiv_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/KushnerykPavel/go-rag-arxiv/internal/client/arxiv"
)

func TestClient(t *testing.T) {
	l := zap.NewNop().Sugar()
	client := arxiv.NewClient(l)
	papers, _ := client.FetchPapers(context.Background(), arxiv.FetchParams{
		SearchCategory: "cs.AI",
		FromDate:       time.Now().UTC().Add(-24 * time.Hour).Format(arxiv.TimeFormat),
		ToDate:         time.Now().UTC().Format(arxiv.TimeFormat),
	})

	for _, paper := range papers {
		fmt.Printf("ID: %s\nTitle: %s\nAuthors: %v\nPublished: %s\nSummary: %s\nAbstract: %s\n\n",
			paper.ArxivID, paper.Title, paper.Authors, paper.PublishedAt, paper.PDFURL, paper.Abstract)
	}
}
