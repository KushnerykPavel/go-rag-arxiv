package rag

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/KushnerykPavel/go-rag-arxiv/internal/client/arxiv"
)

var (
	ErrAskInvalidInput        = errors.New("ask invalid input")
	ErrAskEmptyRetrieval      = errors.New("ask retrieval returned no papers")
	ErrAskUpstreamUnavailable = errors.New("ask upstream unavailable")
	ErrAskRateLimited         = errors.New("ask rate limited")
)

const (
	defaultAskLimit        = 5
	maxAskLimit            = 20
	maxContextAbstractRune = 600
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
	GenerateAnswer(ctx context.Context, query string, contextBlock string) (string, error)
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

func (s *AskService) Ask(ctx context.Context, req AskRequest) (AskResult, error) {
	query := strings.TrimSpace(req.Query)
	if query == "" {
		return AskResult{}, fmt.Errorf("%w: query is required", ErrAskInvalidInput)
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultAskLimit
	}
	if limit > maxAskLimit {
		return AskResult{}, fmt.Errorf("%w: limit must be <= %d", ErrAskInvalidInput, maxAskLimit)
	}

	papers, err := s.retriever.FetchPapersWithQuery(ctx, query, arxiv.FetchParams{
		MaxResults: int(limit),
	})
	if err != nil {
		return AskResult{}, fmt.Errorf("%w: retrieval failed: %v", ErrAskUpstreamUnavailable, err)
	}
	if len(papers) == 0 {
		return AskResult{}, ErrAskEmptyRetrieval
	}

	contextBlock := buildContextBlock(papers)
	answer, err := s.generator.GenerateAnswer(ctx, query, contextBlock)
	if err != nil {
		return AskResult{}, classifyGenerationError(err)
	}

	trimmedAnswer := strings.TrimSpace(answer)
	if trimmedAnswer == "" {
		return AskResult{}, fmt.Errorf("%w: generation returned empty answer", ErrAskUpstreamUnavailable)
	}

	return AskResult{
		Answer:    trimmedAnswer,
		Citations: buildCitations(papers),
	}, nil
}

type statusCoder interface {
	StatusCode() int
}

func classifyGenerationError(err error) error {
	var codeErr statusCoder
	if errors.As(err, &codeErr) {
		switch codeErr.StatusCode() {
		case 429:
			return fmt.Errorf("%w: %v", ErrAskRateLimited, err)
		case 502, 503, 504:
			return fmt.Errorf("%w: %v", ErrAskUpstreamUnavailable, err)
		}
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return fmt.Errorf("%w: %v", ErrAskUpstreamUnavailable, err)
	}

	return fmt.Errorf("%w: %v", ErrAskUpstreamUnavailable, err)
}

func buildContextBlock(papers []arxiv.Paper) string {
	var b strings.Builder
	for i, p := range papers {
		abstract := []rune(strings.TrimSpace(p.Abstract))
		if len(abstract) > maxContextAbstractRune {
			abstract = append(abstract[:maxContextAbstractRune], []rune("...")...)
		}

		b.WriteString(fmt.Sprintf(
			"Paper %d\nID: %s\nTitle: %s\nURL: %s\nAbstract: %s\n\n",
			i+1,
			p.ArxivID,
			strings.TrimSpace(p.Title),
			strings.TrimSpace(p.PDFURL),
			string(abstract),
		))
	}
	return strings.TrimSpace(b.String())
}

func buildCitations(papers []arxiv.Paper) []Citation {
	citations := make([]Citation, 0, len(papers))
	for _, p := range papers {
		citations = append(citations, Citation{
			ID:    p.ArxivID,
			Title: p.Title,
			URL:   p.PDFURL,
		})
	}
	return citations
}
