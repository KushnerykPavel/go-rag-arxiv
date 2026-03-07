package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/KushnerykPavel/go-rag-arxiv/internal/client/arxiv"
	arxivv1 "github.com/KushnerykPavel/go-rag-arxiv/internal/gen/arxiv/v1"
)

type paperFetcher interface {
	FetchPapersWithQuery(ctx context.Context, query string, params arxiv.FetchParams) ([]arxiv.Paper, error)
}

// ArxivHandler implements arxivv1.ArxivServiceServer.
type ArxivHandler struct {
	arxivv1.UnimplementedArxivServiceServer
	fetcher paperFetcher
	l       *zap.SugaredLogger
}

func NewArxivHandler(fetcher paperFetcher, l *zap.SugaredLogger) *ArxivHandler {
	return &ArxivHandler{
		fetcher: fetcher,
		l:       l.With("server", "grpc", "handler", "arxiv"),
	}
}

func (h *ArxivHandler) Search(ctx context.Context, req *arxivv1.SearchRequest) (*arxivv1.SearchResponse, error) {
	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "query is required")
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	papers, err := h.fetcher.FetchPapersWithQuery(ctx, req.Query, arxiv.FetchParams{
		MaxResults: int(limit),
	})
	if err != nil {
		h.l.Errorw("search failed", "query", req.Query, "error", err)
		return nil, status.Errorf(codes.Internal, "search failed: %v", err)
	}

	resp := &arxivv1.SearchResponse{
		Papers: make([]*arxivv1.Paper, 0, len(papers)),
	}
	for _, p := range papers {
		resp.Papers = append(resp.Papers, toProtoPaper(p))
	}
	return resp, nil
}

func toProtoPaper(p arxiv.Paper) *arxivv1.Paper {
	return &arxivv1.Paper{
		Id:      p.ArxivID,
		Title:   p.Title,
		Summary: p.Abstract,
		Url:     p.PDFURL,
		Authors: p.Authors,
	}
}
