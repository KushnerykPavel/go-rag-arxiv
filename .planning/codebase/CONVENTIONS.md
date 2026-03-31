# Coding Conventions

**Analysis Date:** 2026-03-31

## Naming Patterns

**Files:**
- Use `snake_case.go` for multi-word filenames in implementation code, e.g. `internal/cron/arxiv_fetcher.go` and `internal/wrappers/ratelimit.go`.
- Use `_test.go` suffix for test files, e.g. `internal/client/arxiv/client_test.go`.
- Generated protobuf files use `*.pb.go` / `*_grpc.pb.go`, e.g. `internal/gen/arxiv/v1/arxiv.pb.go`.

**Functions:**
- Use exported `PascalCase` for public API (`NewClient`, `FetchPapers`, `SendHTML`) in `internal/client/arxiv/client.go` and `internal/client/telegram/client.go`.
- Use unexported `camelCase` for internal helpers (`fetchPapers`, `doGet`, `parseSingleEntry`) in `internal/client/arxiv/client.go`.
- Constructor naming follows `New<Type>` (`New`, `NewArxivFetcher`, `NewArxivHandler`) in `internal/app/app.go`, `internal/cron/arxiv_fetcher.go`, and `internal/server/grpc/arxiv.go`.

**Variables:**
- Local variables use short `camelCase` names (`cfg`, `errGrp`, `srv`, `lis`) in `internal/app/app.go`.
- Receivers use single-letter names tied to type (`a *App`, `c *Client`, `f *ArxivFetcher`, `h *ArxivHandler`) across `internal/app/app.go`, `internal/client/arxiv/client.go`, `internal/cron/arxiv_fetcher.go`, and `internal/server/grpc/arxiv.go`.

**Types:**
- Exported domain/service types use `PascalCase` (`App`, `Client`, `Paper`, `ArxivFetcher`, `ArxivHandler`) in `internal/app/app.go`, `internal/client/arxiv/paper.go`, and related packages.
- Internal dependency seams are interfaces named by capability (`fetcher`, `notifier`, `paperFetcher`) in `internal/cron/arxiv_fetcher.go` and `internal/server/grpc/arxiv.go`.
- Options pattern uses `type Option func(*Config)` in `internal/client/arxiv/config.go` and `internal/client/telegram/config.go`.

## Code Style

**Formatting:**
- Tool used: `gofmt` style (tabs, canonical import formatting) is the effective formatter across all Go files such as `main/main.go` and `internal/client/arxiv/client.go`.
- Key settings: No repo-level formatter config file detected; rely on standard Go formatting (`go.mod` with `go 1.26`).

**Linting:**
- Tool used: Not detected (`.eslintrc*`, `.prettierrc*`, `eslint.config.*`, `biome.json` are absent).
- Key rules:
- Explicit lint suppression exists for `forbidigo` only where `log.Printf` is intentionally used in signal/logger-sync paths in `main/main.go`.
- Prefer wrapping errors with `%w` and contextual messages (`fmt.Errorf("...: %w", err)`) in `internal/app/app.go`, `internal/client/arxiv/client.go`, `internal/client/telegram/client.go`, and `internal/wrappers/ratelimit.go`.

## Import Organization

**Order:**
1. Standard library imports first (`context`, `fmt`, `net/http`, etc.) in files such as `internal/client/arxiv/client.go`.
2. Third-party imports second (`go.uber.org/zap`, `google.golang.org/grpc`) in `main/main.go` and `internal/server/grpc/arxiv.go`.
3. Local module imports last (`github.com/KushnerykPavel/go-rag-arxiv/...`) in `internal/app/app.go` and other internal packages.

**Path Aliases:**
- Explicit aliases are used when names would collide or to increase clarity:
- `arxivv1` for protobuf package in `internal/app/app.go` and `internal/server/grpc/arxiv.go`.
- `grpcserver` alias for internal gRPC handler package in `internal/app/app.go`.

## Error Handling

**Patterns:**
- Return errors with context and wrapping (`%w`) at boundaries:
- Scheduler/server/client setup in `internal/app/app.go`.
- HTTP and XML operations in `internal/client/arxiv/client.go`.
- Telegram API request/response flow in `internal/client/telegram/client.go`.
- Convert invalid client input to gRPC status errors (`codes.InvalidArgument`) and backend failures to `codes.Internal` in `internal/server/grpc/arxiv.go`.
- For background/loop operations, log and continue instead of failing entire job (`FetchPapers` loop and notification sending) in `internal/cron/arxiv_fetcher.go`.
- Selective ignore pattern (`_ =`) is used for non-critical operations (startup/shutdown Telegram notifications, health-write return) in `internal/app/app.go`.

## Logging

**Framework:** `go.uber.org/zap` (`*zap.SugaredLogger`) across `main/main.go`, `internal/app/app.go`, `internal/client/arxiv/client.go`, `internal/client/telegram/client.go`, `internal/cron/arxiv_fetcher.go`, and `internal/server/grpc/arxiv.go`.

**Patterns:**
- Inject logger dependency from top-level (`main/main.go` -> `app.New` -> constructors in internal packages).
- Attach structured context with `.With(...)` during client/handler/fetcher construction:
- `internal/client/telegram/client.go`
- `internal/cron/arxiv_fetcher.go`
- `internal/server/grpc/arxiv.go`
- Use structured fields through `Infow/Errorw/Warnw` in service code.
- Use fatal logging for unrecoverable startup failure in `main/main.go`.

## Comments

**When to Comment:**
- Public exported types/functions are documented with concise doc comments (`Client`, `FetchParams`, `DownloadPDF`, `RateLimiter`) in `internal/client/arxiv/client.go` and `internal/wrappers/ratelimit.go`.
- Complex protocol/API behavior is explained with targeted comments (arXiv query escaping and version stripping) in `internal/client/arxiv/client.go`.

**JSDoc/TSDoc:**
- Not applicable for this Go repository.
- Go doc comments are used for exported symbols in `internal/client/arxiv/*.go`, `internal/client/telegram/*.go`, and `internal/wrappers/ratelimit.go`.

## Function Design

**Size:** Functions are mostly compact and single-purpose; notable larger orchestrator function is `(*App).Run` in `internal/app/app.go`, which coordinates scheduler, HTTP, gRPC, and shutdown.

**Parameters:** `context.Context` is passed explicitly to I/O and long-running calls (`Run`, `FetchPapers`, HTTP requests, notifier calls) in `internal/app/app.go`, `internal/client/arxiv/client.go`, and `internal/cron/arxiv_fetcher.go`.

**Return Values:**
- Business and infra operations usually return `(value, error)` (`FetchPapers`, `FetchPaperByID`, `Search`) in `internal/client/arxiv/client.go` and `internal/server/grpc/arxiv.go`.
- Fire-and-forget scheduler task uses no return and handles errors via logging inside method (`ArxivFetcher.FetchPapers`) in `internal/cron/arxiv_fetcher.go`.

## Module Design

**Exports:** Keep package internals private by default; export only constructors, DTOs, and service entry methods:
- Exported API in `internal/client/arxiv/client.go` and `internal/client/telegram/client.go`
- Unexported helpers for parsing/network internals in `internal/client/arxiv/client.go`

**Barrel Files:** Not used. Packages are imported directly by path (e.g. `github.com/KushnerykPavel/go-rag-arxiv/internal/client/arxiv`) throughout `internal/app/app.go` and other files.

---

*Convention analysis: 2026-03-31*
