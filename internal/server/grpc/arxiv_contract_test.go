package grpc

import (
	"testing"

	arxivv1 "github.com/KushnerykPavel/go-rag-arxiv/internal/gen/arxiv/v1"
)

func TestArxivServiceContract(t *testing.T) {
	explicitHandlerMethods := map[string]struct{}{
		"Search": {},
		"Ask":    {},
	}

	for _, method := range arxivv1.ArxivService_ServiceDesc.Methods {
		if _, ok := explicitHandlerMethods[method.MethodName]; !ok {
			t.Fatalf("proto declares unimplemented RPC method: %s", method.MethodName)
		}
	}
}
