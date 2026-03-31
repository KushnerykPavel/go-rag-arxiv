# Architecture

**Analysis Date:** 2026-03-31

## Pattern Overview

**Overall:** Layered modular monolith with composition-root orchestration.

**Key Characteristics:**
- Single binary entrypoint in `main/main.go` that wires config, logging, lifecycle, and app execution.
- Central application orchestrator in `internal/app/app.go` that composes HTTP, gRPC, scheduler, and external clients.
- Interface-driven boundaries in `internal/cron/arxiv_fetcher.go` and `internal/server/grpc/arxiv.go` to decouple orchestration from concrete clients.

## Layers

**Composition Root / Bootstrap:**
- Purpose: Process environment configuration, initialize logger, own OS signal lifecycle, start the app.
- Location: `main/main.go`
- Contains: `main()`, signal handler, `envconfig.Process` wiring to `app.Config`.
- Depends on: `internal/app`, `github.com/kelseyhightower/envconfig`, `go.uber.org/zap`.
- Used by: Process startup (`go run ./main` or built container entrypoint from `Dockerfile`).

**Application Orchestration Layer:**
- Purpose: Compose runtime components and supervise HTTP, gRPC, scheduler, and graceful shutdown.
- Location: `internal/app/app.go`, `internal/app/config.go`
- Contains: `App.Run`, scheduler job registration, HTTP `/health`, gRPC server registration.
- Depends on: `internal/client/arxiv`, `internal/client/telegram`, `internal/cron`, `internal/server/grpc`, `internal/wrappers`, `internal/gen/arxiv/v1`.
- Used by: `main/main.go`.

**Transport Layer (Inbound):**
- Purpose: Accept inbound network traffic and map transport models to domain/client operations.
- Location: `internal/server/grpc/arxiv.go`, `proto/arxiv/v1/arxiv.proto`, `internal/gen/arxiv/v1/arxiv.pb.go`, `internal/gen/arxiv/v1/arxiv_grpc.pb.go`
- Contains: gRPC handler (`ArxivHandler`) and protobuf contracts (`ArxivService`, `SearchRequest`, `SearchResponse`, etc.).
- Depends on: `internal/client/arxiv` for fetching papers, generated protobuf package in `internal/gen/arxiv/v1`.
- Used by: gRPC server setup in `internal/app/app.go`.

**Job/Workflow Layer:**
- Purpose: Run scheduled arXiv fetch workflows and fan out notifications.
- Location: `internal/cron/arxiv_fetcher.go`
- Contains: `ArxivFetcher`, fixed topic list, per-paper notification loop.
- Depends on: `internal/client/arxiv` types, notifier interface (implemented by telegram client), `internal/wrappers/ratelimit.go`.
- Used by: Scheduler registration in `internal/app/app.go`.

**Outbound Client Layer (External APIs):**
- Purpose: Encapsulate HTTP communication with external services.
- Location: `internal/client/arxiv/client.go`, `internal/client/arxiv/config.go`, `internal/client/arxiv/paper.go`, `internal/client/telegram/client.go`, `internal/client/telegram/config.go`
- Contains: arXiv query/download logic, XML parsing, Telegram message dispatch.
- Depends on: stdlib HTTP/XML/time, zap logging.
- Used by: `internal/app/app.go`, `internal/server/grpc/arxiv.go`, `internal/cron/arxiv_fetcher.go`.

**Shared Utility Layer:**
- Purpose: Cross-cutting helper utilities used by workflows.
- Location: `internal/wrappers/ratelimit.go`
- Contains: `RateLimiter` wrapper over `golang.org/x/time/rate`.
- Depends on: `golang.org/x/time/rate`.
- Used by: `internal/cron/arxiv_fetcher.go` via `limiter.Do`.

## Data Flow

**Scheduled Fetch + Notification Flow:**

1. `internal/app/app.go` creates `gocron` scheduler and registers `arxivFetcher.FetchPapers`.
2. `internal/cron/arxiv_fetcher.go` computes date window and topic filters, then calls `fetcher.FetchPapers`.
3. `internal/client/arxiv/client.go` builds arXiv query URL, executes HTTP request, parses Atom XML into `arxiv.Paper`.
4. `internal/cron/arxiv_fetcher.go` sends each paper through rate-limited notifier calls.
5. `internal/client/telegram/client.go` posts formatted HTML notifications to Telegram Bot API.

**gRPC Search Flow:**

1. `internal/app/app.go` starts `grpc.NewServer` and registers `ArxivServiceServer` from `internal/server/grpc/arxiv.go`.
2. Incoming `Search` RPC hits `ArxivHandler.Search` in `internal/server/grpc/arxiv.go`.
3. Handler validates request and calls `FetchPapersWithQuery` on injected fetcher (arXiv client).
4. `internal/client/arxiv/client.go` fetches and parses data from arXiv API.
5. Handler maps `arxiv.Paper` to protobuf `arxivv1.Paper` and returns `SearchResponse`.

**State Management:**
- Primarily stateless request/job processing with ephemeral in-memory state.
- Explicit mutable state exists in `internal/client/arxiv/client.go` (`lastRequestAt` + mutex) for client-side rate limiting.
- Runtime services (HTTP server, gRPC server, scheduler) are lifecycle-managed in `internal/app/app.go` and tied to context cancellation from `main/main.go`.

## Key Abstractions

**Application Container (`App`):**
- Purpose: Runtime assembly and coordinated startup/shutdown.
- Examples: `internal/app/app.go`, `internal/app/config.go`
- Pattern: Composition root with explicit dependency construction.

**Arxiv Paper Model (`Paper`):**
- Purpose: Internal canonical representation of paper metadata shared across cron and gRPC paths.
- Examples: `internal/client/arxiv/paper.go`, mapped in `internal/server/grpc/arxiv.go`
- Pattern: Plain data struct translated to transport DTOs.

**Port Interfaces for Inversion:**
- Purpose: Decouple workflow/transport code from concrete clients.
- Examples: `fetcher` and `notifier` in `internal/cron/arxiv_fetcher.go`, `paperFetcher` in `internal/server/grpc/arxiv.go`
- Pattern: Small local interfaces consumed via constructor injection.

**Generated Contract Boundary:**
- Purpose: Define and enforce gRPC API surface.
- Examples: `proto/arxiv/v1/arxiv.proto`, generated outputs in `internal/gen/arxiv/v1/*.go`
- Pattern: Proto-first API with generated server/client stubs.

## Entry Points

**Process Entry Point:**
- Location: `main/main.go`
- Triggers: Program execution (`./app`, `go run ./main`, container `ENTRYPOINT` from `Dockerfile`).
- Responsibilities: Load env config into `app.Config`, create logger, handle SIGTERM/SIGINT, call `App.Run`.

**HTTP Health Endpoint:**
- Location: route registration in `internal/app/app.go`
- Triggers: HTTP GET `/health` on configured `Config.Address`.
- Responsibilities: Liveness-style JSON response `{"status":"ok"}`.

**gRPC Service Entry Point:**
- Location: server setup in `internal/app/app.go`, handler in `internal/server/grpc/arxiv.go`
- Triggers: gRPC requests to `arxiv.v1.ArxivService/Search` (declared in `proto/arxiv/v1/arxiv.proto`).
- Responsibilities: Validate request, execute arXiv search, map to protobuf response.

**Scheduler Entry Point:**
- Location: job registration in `internal/app/app.go`
- Triggers: Cron expression `0 5 * * *`.
- Responsibilities: Run daily `ArxivFetcher.FetchPapers` workflow from `internal/cron/arxiv_fetcher.go`.

## Error Handling

**Strategy:** Fail-fast during startup/wiring; continue-on-error within recurring fetch loops.

**Patterns:**
- Error wrapping with context (`fmt.Errorf("...: %w", err)`) across `main/main.go`, `internal/app/app.go`, `internal/client/arxiv/client.go`, `internal/client/telegram/client.go`.
- Transport-level error mapping in gRPC handler (`status.Error`, `codes.InvalidArgument`, `codes.Internal`) in `internal/server/grpc/arxiv.go`.
- Best-effort shutdown/send operations (ignored errors on startup/stop Telegram notifications) in `internal/app/app.go`.

## Cross-Cutting Concerns

**Logging:** Structured logging with `zap.SugaredLogger` passed through constructors and augmented with scoped fields in `main/main.go`, `internal/app/app.go`, `internal/cron/arxiv_fetcher.go`, `internal/client/telegram/client.go`, `internal/server/grpc/arxiv.go`.
**Validation:** Input validation at API boundaries (`req.Query` and limit normalization in `internal/server/grpc/arxiv.go`); constructor validation for wrappers (`NewRateLimiter` in `internal/wrappers/ratelimit.go`).
**Authentication:** Token-based Telegram auth injected from env-backed config (`internal/app/config.go`) and used when constructing bot API endpoint in `internal/client/telegram/client.go`.

---

*Architecture analysis: 2026-03-31*
