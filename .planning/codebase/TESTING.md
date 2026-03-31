# Testing Patterns

**Analysis Date:** 2026-03-31

## Test Framework

**Runner:**
- Go standard `testing` package (used in all tests such as `internal/app/config_test.go`, `internal/server/grpc/arxiv_ask_test.go`).
- Config: Not detected (no `go test` config files).

**Assertion Library:**
- Standard `testing` assertions via `t.Fatalf`, `t.Fatal` in `internal/app/config_test.go`, `internal/server/grpc/arxiv_ask_test.go`.

**Run Commands:**
```bash
go test ./...          # Run all tests (documented in `CLAUDE.md`)
```

## Test File Organization

**Location:**
- Co-located with implementation packages under `internal/...` (e.g., `internal/app/config_test.go`, `internal/server/grpc/arxiv_ask_test.go`).

**Naming:**
- `_test.go` suffix (e.g., `internal/client/arxiv/client_test.go`).

**Structure:**
```
internal/app/*_test.go
internal/client/arxiv/*_test.go
internal/server/grpc/*_test.go
```

## Test Structure

**Suite Organization:**
```go
tests := []struct {
	name    string
	cfg     Config
	wantErr string
}{
	{name: "passes when groq and telegram values are set", cfg: Config{GroqAPIKey: "groq-key"}},
}

for _, tt := range tests {
	tt := tt
	t.Run(tt.name, func(t *testing.T) {
		t.Parallel()
		err := tt.cfg.Validate()
		if tt.wantErr == "" && err != nil {
			t.Fatalf("Validate() error = %v, want nil", err)
		}
	})
}
```

**Patterns:**
- Table-driven subtests with `t.Run` and per-case parallelization in `internal/app/config_test.go`.
- Non-parallel subtests when shared state is used in `internal/server/grpc/arxiv_ask_test.go`.
- Validation of gRPC status codes and messages using `status.FromError` in `internal/server/grpc/arxiv_ask_test.go`.

## Mocking

**Framework:** None (manual fakes).

**Patterns:**
```go
type fakeAskService struct {
	askFn func(ctx context.Context, req rag.AskRequest) (rag.AskResult, error)
}

func (f fakeAskService) Ask(ctx context.Context, req rag.AskRequest) (rag.AskResult, error) {
	return f.askFn(ctx, req)
}
```

**What to Mock:**
- External service interfaces defined at point-of-use (e.g., `askService` in `internal/server/grpc/arxiv.go`, mocked by `fakeAskService` in `internal/server/grpc/arxiv_ask_test.go`).

**What NOT to Mock:**
- Pure config validation logic in `internal/app/config.go` is tested directly in `internal/app/config_test.go`.

## Fixtures and Factories

**Test Data:**
```go
cfg := Config{
	GroqAPIKey: "groq-key",
	TelegramConfig: TelegramConfig{
		Token:  "token",
		ChatID: 12345,
	},
}
```

**Location:**
- Inline within tests; no shared fixture directory detected in `internal/...`.

## Coverage

**Requirements:** Not enforced (no coverage tooling/config detected).

**View Coverage:**
```bash
Not detected
```

## Test Types

**Unit Tests:**
- Config validation (`internal/app/config_test.go`).
- gRPC handler error mapping (`internal/server/grpc/arxiv_ask_test.go`).
- gRPC contract checks on service descriptors (`internal/server/grpc/arxiv_contract_test.go`).

**Integration Tests:**
- Direct arXiv API calls without assertions in `internal/client/arxiv/client_test.go`.

**E2E Tests:**
- Not used.

## Common Patterns

**Async Testing:**
```go
t.Parallel()
```

**Error Testing:**
```go
if err == nil {
	t.Fatalf("expected error")
}
```

---

*Testing analysis: 2026-03-31*
