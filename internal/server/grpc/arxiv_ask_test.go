package grpc

import (
	"testing"

	arxivv1 "github.com/KushnerykPavel/go-rag-arxiv/internal/gen/arxiv/v1"
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
