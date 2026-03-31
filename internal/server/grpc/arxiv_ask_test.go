package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/KushnerykPavel/go-rag-arxiv/internal/rag"
	arxivv1 "github.com/KushnerykPavel/go-rag-arxiv/internal/gen/arxiv/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestAskContract(t *testing.T) {
	t.Run("service descriptor includes Ask", func(t *testing.T) {
		hasAsk := false
		for _, method := range arxivv1.ArxivService_ServiceDesc.Methods {
			if method.MethodName == "Ask" {
				hasAsk = true
				break
			}
		}
		if !hasAsk {
			t.Fatalf("expected ArxivService_ServiceDesc to include Ask method")
		}
	})

	t.Run("AskResponse includes answer and repeated citations", func(t *testing.T) {
		msg := arxivv1.File_arxiv_v1_arxiv_proto.Messages().ByName("AskResponse")
		if msg == nil {
			t.Fatalf("expected AskResponse message in proto file descriptor")
		}

		answer := msg.Fields().ByName("answer")
		if answer == nil {
			t.Fatalf("expected AskResponse.answer field")
		}
		if answer.Kind() != protoreflect.StringKind {
			t.Fatalf("expected AskResponse.answer to be string, got %s", answer.Kind())
		}

		citations := msg.Fields().ByName("citations")
		if citations == nil {
			t.Fatalf("expected AskResponse.citations field")
		}
		if !citations.IsList() {
			t.Fatalf("expected AskResponse.citations to be repeated")
		}
		if citations.Kind() != protoreflect.MessageKind {
			t.Fatalf("expected AskResponse.citations to be message list, got %s", citations.Kind())
		}
		if citations.Message().Name() != "Citation" {
			t.Fatalf("expected AskResponse.citations element type Citation, got %s", citations.Message().Name())
		}
	})
}

type fakeAskService struct {
	askFn func(ctx context.Context, req rag.AskRequest) (rag.AskResult, error)
}

func (f fakeAskService) Ask(ctx context.Context, req rag.AskRequest) (rag.AskResult, error) {
	return f.askFn(ctx, req)
}

func TestAskSuccess(t *testing.T) {
	handler := NewArxivHandler(nil, fakeAskService{
		askFn: func(_ context.Context, req rag.AskRequest) (rag.AskResult, error) {
			if req.Query != "What is new in RAG?" || req.Limit != 3 {
				t.Fatalf("unexpected ask request: %+v", req)
			}
			return rag.AskResult{
				Answer: "RAG improved retrieval quality.",
				Citations: []rag.Citation{
					{ID: "1234.5678", Title: "RAG paper", URL: "https://arxiv.org/abs/1234.5678"},
				},
			}, nil
		},
	}, zap.NewNop().Sugar())

	resp, err := handler.Ask(context.Background(), &arxivv1.AskRequest{
		Query: "What is new in RAG?",
		Limit: 3,
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if resp.GetAnswer() == "" {
		t.Fatalf("expected non-empty answer")
	}
	if len(resp.GetCitations()) == 0 {
		t.Fatalf("expected at least one citation")
	}
}

func TestAskStatusMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		req         *arxivv1.AskRequest
		askErr      error
		wantCalled  bool
		wantCode    codes.Code
		wantMessage string
	}{
		{
			name:        "invalid input",
			req:         &arxivv1.AskRequest{Query: ""},
			wantCalled:  false,
			wantCode:    codes.InvalidArgument,
			wantMessage: "query is required",
		},
		{
			name:        "empty retrieval",
			req:         &arxivv1.AskRequest{Query: "none"},
			askErr:      rag.ErrAskEmptyRetrieval,
			wantCalled:  true,
			wantCode:    codes.NotFound,
			wantMessage: "ask failed: ask retrieval returned no papers",
		},
		{
			name:        "generation timeout",
			req:         &arxivv1.AskRequest{Query: "timeout"},
			askErr:      context.DeadlineExceeded,
			wantCalled:  true,
			wantCode:    codes.DeadlineExceeded,
			wantMessage: "ask failed: context deadline exceeded",
		},
		{
			name:        "upstream unavailable",
			req:         &arxivv1.AskRequest{Query: "unavailable"},
			askErr:      rag.ErrAskUpstreamUnavailable,
			wantCalled:  true,
			wantCode:    codes.Unavailable,
			wantMessage: "ask failed: ask upstream unavailable",
		},
		{
			name:        "rate limited",
			req:         &arxivv1.AskRequest{Query: "429"},
			askErr:      rag.ErrAskRateLimited,
			wantCalled:  true,
			wantCode:    codes.ResourceExhausted,
			wantMessage: "ask failed: ask rate limited",
		},
		{
			name:        "internal fallback",
			req:         &arxivv1.AskRequest{Query: "boom"},
			askErr:      errors.New("boom"),
			wantCalled:  true,
			wantCode:    codes.Internal,
			wantMessage: "ask failed: boom",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			called := false
			handler := NewArxivHandler(nil, fakeAskService{
				askFn: func(_ context.Context, req rag.AskRequest) (rag.AskResult, error) {
					called = true
					return rag.AskResult{}, tt.askErr
				},
			}, zap.NewNop().Sugar())

			_, err := handler.Ask(context.Background(), tt.req)
			if err == nil {
				t.Fatalf("expected error")
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("expected grpc status error, got %T", err)
			}
			if st.Code() != tt.wantCode {
				t.Fatalf("unexpected code: got=%s want=%s", st.Code(), tt.wantCode)
			}
			if st.Message() != tt.wantMessage {
				t.Fatalf("unexpected message: got=%q want=%q", st.Message(), tt.wantMessage)
			}
			if called != tt.wantCalled {
				t.Fatalf("unexpected asker call: called=%v want=%v", called, tt.wantCalled)
			}
		})
	}
}
