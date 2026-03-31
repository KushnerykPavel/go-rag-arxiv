package grpc

import (
	"context"
	"errors"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/KushnerykPavel/go-rag-arxiv/internal/client/arxiv"
	arxivv1 "github.com/KushnerykPavel/go-rag-arxiv/internal/gen/arxiv/v1"
	"github.com/KushnerykPavel/go-rag-arxiv/internal/rag"
)

type paperFetcher interface {
	FetchPapersWithQuery(ctx context.Context, query string, params arxiv.FetchParams) ([]arxiv.Paper, error)
}

type askService interface {
	Ask(ctx context.Context, req rag.AskRequest) (rag.AskResult, error)
}

// ArxivHandler implements arxivv1.ArxivServiceServer.
type ArxivHandler struct {
	arxivv1.UnimplementedArxivServiceServer
	fetcher paperFetcher
	asker   askService
	l       *zap.SugaredLogger
}

func NewArxivHandler(fetcher paperFetcher, asker askService, l *zap.SugaredLogger) *ArxivHandler {
	return &ArxivHandler{
		fetcher: fetcher,
		asker:   asker,
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

func (h *ArxivHandler) Ask(ctx context.Context, req *arxivv1.AskRequest) (*arxivv1.AskResponse, error) {
	query := strings.TrimSpace(req.GetQuery())
	if query == "" {
		return nil, status.Error(codes.InvalidArgument, "query is required")
	}

	result, err := h.asker.Ask(ctx, rag.AskRequest{
		Query: query,
		Limit: req.GetLimit(),
	})
	if err != nil {
		code := mapAskErrorCode(err)
		h.l.Errorw("ask failed", "query", req.GetQuery(), "code", code.String(), "error", err)
		return nil, status.Errorf(code, "ask failed: %v", err)
	}

	resp := &arxivv1.AskResponse{
		Answer:    result.Answer,
		Citations: make([]*arxivv1.Citation, 0, len(result.Citations)),
	}
	for _, c := range result.Citations {
		resp.Citations = append(resp.Citations, &arxivv1.Citation{
			Id:    c.ID,
			Title: c.Title,
			Url:   c.URL,
		})
	}

	return resp, nil
}

func mapAskErrorCode(err error) codes.Code {
	switch {
	case errors.Is(err, rag.ErrAskInvalidInput):
		return codes.InvalidArgument
	case errors.Is(err, rag.ErrAskEmptyRetrieval):
		return codes.NotFound
	case errors.Is(err, context.DeadlineExceeded):
		return codes.DeadlineExceeded
	case errors.Is(err, rag.ErrAskRateLimited):
		return codes.ResourceExhausted
	case errors.Is(err, rag.ErrAskUpstreamUnavailable):
		return codes.Unavailable
	default:
		return codes.Internal
	}
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
