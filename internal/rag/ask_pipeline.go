package rag

import (
	"context"
	"errors"

	"github.com/KushnerykPavel/go-rag-arxiv/internal/client/arxiv"
)

var (
	ErrAskInvalidInput       = errors.New("ask invalid input")
	ErrAskEmptyRetrieval     = errors.New("ask retrieval returned no papers")
	ErrAskUpstreamUnavailable = errors.New("ask upstream unavailable")
	ErrAskRateLimited        = errors.New("ask rate limited")
)

type AskRequest struct {
	Query string
	Limit int32
}

type Citation struct {
	ID    string
	Title string
	URL   string
}

type AskResult struct {
	Answer    string
	Citations []Citation
}

type Retriever interface {
	FetchPapersWithQuery(ctx context.Context, query string, params arxiv.FetchParams) ([]arxiv.Paper, error)
}

type Generator interface {
	GenerateAnswer(ctx context.Context, query string, papers []arxiv.Paper) (string, error)
}

type AskService struct {
	retriever Retriever
	generator Generator
}

func NewAskService(retriever Retriever, generator Generator) *AskService {
	return &AskService{
		retriever: retriever,
		generator: generator,
	}
}
