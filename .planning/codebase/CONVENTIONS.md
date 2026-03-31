# Coding Conventions

**Analysis Date:** 2026-03-31

## Naming Patterns

**Files:**
- Use `snake_case.go` for multi-word filenames, e.g. `internal/cron/arxiv_fetcher.go`, `internal/wrappers/ratelimit.go`.
- Use `_test.go` suffix for tests, e.g. `internal/app/config_test.go`, `internal/server/grpc/arxiv_ask_test.go`.
- Generated protobuf files use `*.pb.go` and `*_grpc.pb.go`, e.g. `internal/gen/arxiv/v1/arxiv.pb.go`.

**Functions:**
- Exported functions/types are `PascalCase` (`NewArxivHandler`, `FetchPapers`, `SendMarkdown`) in `internal/server/grpc/arxiv.go`, `internal/client/arxiv/client.go`, `internal/client/telegram/client.go`.
- Unexported helpers are `camelCase` (`fetchPapers`, `doGet`, `parseSingleEntry`) in `internal/client/arxiv/client.go`.

**Variables:**
- Short `camelCase` locals for small scopes (`cfg`, `errGrp`, `srv`, `lis`) in `internal/app/app.go`.
- Receiver names are single-letter tied to type (`a *App`, `c *Client`, `f *ArxivFetcher`, `h *ArxivHandler`) across `internal/app/app.go`, `internal/client/arxiv/client.go`, `internal/cron/arxiv_fetcher.go`, `internal/server/grpc/arxiv.go`.

**Types:**
- Exported domain/service types in `PascalCase` (`App`, `Client`, `Paper`, `ArxivFetcher`, `ArxivHandler`) in `internal/app/app.go`, `internal/client/arxiv/paper.go`, `internal/cron/arxiv_fetcher.go`, `internal/server/grpc/arxiv.go`.
- Internal dependency seams are unexported interfaces named by capability (`fetcher`, `notifier`, `paperFetcher`, `askService`) in `internal/cron/arxiv_fetcher.go`, `internal/server/grpc/arxiv.go`.

## Code Style

**Formatting:**
- Tool used: `gofmt` (standard Go formatting in all Go files such as `main/main.go`, `internal/client/arxiv/client.go`).
- Key settings: No repo-level formatter config detected; use `go fmt ./...` from `CLAUDE.md`.

**Linting:**
- Tool used: Not detected (no `.golangci.*` or equivalent config files found).
- Key rules:
- Inline lint suppression exists for `forbidigo` in `main/main.go` where `log.Printf` is intentionally used.

## Import Organization

**Order:**
1. Standard library imports (`context`, `fmt`, `net/http`) in files like `internal/app/app.go`, `internal/client/telegram/client.go`.
2. Third-party dependencies (`go.uber.org/zap`, `google.golang.org/grpc`) in `internal/app/app.go`, `internal/server/grpc/arxiv.go`.
3. Internal module imports (`github.com/KushnerykPavel/go-rag-arxiv/...`) in `internal/app/app.go`, `internal/server/grpc/arxiv.go`.

**Path Aliases:**
- Use aliases where clarity or name conflicts exist:
- `arxivv1` for protobuf package in `internal/app/app.go`, `internal/server/grpc/arxiv.go`.
- `grpcserver` for handler package in `internal/app/app.go`.

## Error Handling

**Patterns:**
- Wrap errors with context and `%w` (`fmt.Errorf("creating scheduler: %w", err)`) in `internal/app/app.go`, `internal/client/arxiv/client.go`, `internal/client/telegram/client.go`, `internal/wrappers/ratelimit.go`.
- Use sentinel errors for domain classification (`ErrAskInvalidInput`, `ErrAskRateLimited`) in `internal/rag/ask_pipeline.go`.
- Map domain errors to transport errors in gRPC handlers using `status.Error`/`status.Errorf` and `codes.*` in `internal/server/grpc/arxiv.go`.
- Use `errors.Is`/`errors.As`/`errors.Join` to classify upstream errors in `internal/rag/ask_pipeline.go`.
- In background loops, log and continue on per-item failures (`ArxivFetcher.FetchPapers`) in `internal/cron/arxiv_fetcher.go`.
- Use selective ignores for non-critical operations (`_ = telegramClient.SendMarkdown(...)`, `_ = w.Write(...)`) in `internal/app/app.go`.

## Logging

**Framework:** `go.uber.org/zap` (Sugared logger) in `main/main.go`.

**Patterns:**
- Logger is constructed at entrypoint and injected (`main/main.go` -> `app.New` -> internal clients).
- Add scoped context via `.With(...)` during construction in `internal/client/groq/client.go`, `internal/client/telegram/client.go`, `internal/cron/arxiv_fetcher.go`, `internal/server/grpc/arxiv.go`.
- Use structured `Infow`, `Warnw`, `Errorw` for operational events in `internal/app/app.go`, `internal/client/arxiv/client.go`, `internal/cron/arxiv_fetcher.go`.
- Fatal logging only at top-level for unrecoverable startup failures in `main/main.go`.

## Comments

**When to Comment:**
- Exported types/functions use doc comments (e.g., `Client`, `FetchParams`, `RateLimiter`) in `internal/client/arxiv/client.go`, `internal/wrappers/ratelimit.go`.
- Behavior-specific clarifications appear near tricky logic (e.g., arXiv query escaping, version suffix stripping) in `internal/client/arxiv/client.go`.

**JSDoc/TSDoc:**
- Not applicable (Go codebase).

## Function Design

**Size:** Keep functions focused on one responsibility; most exported methods are short and layered with helpers (e.g., `FetchPapers` delegates to `fetchPapers`) in `internal/client/arxiv/client.go`.

**Parameters:** Use explicit input structs for multi-parameter calls (`FetchParams`, `AskRequest`) in `internal/client/arxiv/client.go`, `internal/rag/ask_pipeline.go`.

**Return Values:** Prefer `(value, error)` and return early on error; use `nil, error` for failure paths in `internal/server/grpc/arxiv.go`, `internal/client/arxiv/client.go`.

## Module Design

**Exports:** Keep public API small; internal helpers unexported within package (e.g., `parseResponse`, `downloadFile`) in `internal/client/arxiv/client.go`.

**Barrel Files:** Not detected.

---

*Convention analysis: 2026-03-31*
