# Testing Patterns

**Analysis Date:** 2026-03-31

## Test Framework

**Runner:**
- Go standard test runner (`go test`) from Go toolchain aligned to `go 1.26` in `go.mod`.
- Config: No dedicated test config file detected (`jest.config.*`, `vitest.config.*`, or Go-specific third-party config not present).

**Assertion Library:**
- Go standard `testing` package only (`testing.T` in `internal/client/arxiv/client_test.go`).
- No `testify`, `gomega`, or other assertion helper library imports detected.

**Run Commands:**
```bash
go test ./...                                  # Run all tests
go test ./... -run TestClient -count=1         # Run specific test without cache
go test ./... -cover                           # Package-level coverage summary
```

## Test File Organization

**Location:**
- Co-located tests next to implementation package.
- Current example: `internal/client/arxiv/client_test.go` sits beside `internal/client/arxiv/client.go`.

**Naming:**
- Test files follow Go convention: `*_test.go` (`internal/client/arxiv/client_test.go`).
- Test functions use `TestXxx` naming (`TestClient` in `internal/client/arxiv/client_test.go`).

**Structure:**
```text
internal/
  client/
    arxiv/
      client.go
      client_test.go
```

## Test Structure

**Suite Organization:**
```typescript
// Go example from `internal/client/arxiv/client_test.go`
func TestClient(t *testing.T) {
    l := zap.NewNop().Sugar()
    client := arxiv.NewClient(l)
    papers, _ := client.FetchPapers(context.Background(), arxiv.FetchParams{
        SearchCategory: "cs.AI",
        FromDate:       time.Now().UTC().Add(-24 * time.Hour).Format(arxiv.TimeFormat),
        ToDate:         time.Now().UTC().Format(arxiv.TimeFormat),
    })

    for _, paper := range papers {
        fmt.Printf("ID: %s ...\n", paper.ArxivID)
    }
}
```

**Patterns:**
- Setup pattern:
- Instantiate a no-op logger (`zap.NewNop().Sugar()`) to avoid noisy logs in tests in `internal/client/arxiv/client_test.go`.
- Construct real client via production constructor (`arxiv.NewClient`) in `internal/client/arxiv/client_test.go`.
- Teardown pattern:
- No explicit teardown or cleanup currently used.
- Assertion pattern:
- No assertions currently used (`papers, _ := ...` ignores errors; loop prints values) in `internal/client/arxiv/client_test.go`.

## Mocking

**Framework:** Not detected in current tests.

**Patterns:**
```typescript
// No mocking pattern is currently implemented in test files.
// Existing production code uses interfaces that can support manual fakes:
// `fetcher` and `notifier` in `internal/cron/arxiv_fetcher.go`
// `paperFetcher` in `internal/server/grpc/arxiv.go`
```

**What to Mock:**
- External network boundaries should be mocked in new tests:
- arXiv HTTP interactions in `internal/client/arxiv/client.go` (use `WithHTTPClient` from `internal/client/arxiv/config.go`).
- Telegram Bot API calls in `internal/client/telegram/client.go` (use `WithHTTPClient` from `internal/client/telegram/config.go`).
- Scheduler-triggered notifier/fetcher dependencies in `internal/cron/arxiv_fetcher.go` (replace interfaces with fakes).

**What NOT to Mock:**
- Pure transformation helpers in `internal/client/arxiv/client.go` (`extractArxivID`, `cleanText`, `toProtoPaper` indirectly via handler output).
- Value structs/config parsing logic in `internal/app/config.go` where direct deterministic checks are feasible.

## Fixtures and Factories

**Test Data:**
```typescript
// Current pattern uses dynamic time window values:
arxiv.FetchParams{
    SearchCategory: "cs.AI",
    FromDate: time.Now().UTC().Add(-24 * time.Hour).Format(arxiv.TimeFormat),
    ToDate:   time.Now().UTC().Format(arxiv.TimeFormat),
}
```

**Location:**
- No shared fixture/factory package detected.
- Inline setup only in `internal/client/arxiv/client_test.go`.

## Coverage

**Requirements:** None enforced by config or CI files in this repository.

**Observed current state (from `go test ./... -cover`):**
- `internal/client/arxiv`: 46.7%
- `internal/app`: 0.0%
- `internal/client/telegram`: 0.0%
- `internal/cron`: 0.0%
- `internal/server/grpc`: 0.0%
- `internal/wrappers`: 0.0%

**View Coverage:**
```bash
go test ./... -cover
go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out
```

## Test Types

**Unit Tests:**
- Minimal unit coverage exists only for `internal/client/arxiv` package (`internal/client/arxiv/client_test.go`).
- Current test behaves like a smoke/integration probe because it performs real fetch calls through the live client path.

**Integration Tests:**
- No dedicated integration test suite folder or tagging strategy detected.
- Existing `TestClient` in `internal/client/arxiv/client_test.go` effectively exercises external arXiv API behavior.

**E2E Tests:**
- Not used (no E2E framework, harness, or workflow files detected).

## Common Patterns

**Async Testing:**
```typescript
// No async-specific testing helpers or patterns are currently implemented.
// Production code uses context cancellation and goroutines in:
// `main/main.go` and `internal/app/app.go`.
```

**Error Testing:**
```typescript
// No explicit error-path assertions are currently implemented.
// New tests should check returned errors instead of discarding them:
papers, err := client.FetchPapers(ctx, params)
if err != nil {
    t.Fatalf("FetchPapers failed: %v", err)
}
```

---

*Testing analysis: 2026-03-31*
