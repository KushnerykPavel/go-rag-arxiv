# Architecture

**Analysis Date:** 2026-03-31

## Pattern Overview

**Overall:** Layered Go service with adapter-style clients and a centralized application orchestrator.

**Key Characteristics:**
- Single process hosting HTTP health, gRPC API, and scheduled jobs.
- External integrations encapsulated behind client interfaces.
- Business logic concentrated in a small service layer.

## Layers

**Application Orchestrator:**
- Purpose: Compose dependencies, start servers/scheduler, manage shutdown.
- Location: `internal/app/app.go`
- Contains: wiring, cron schedule, server startup, lifecycle control.
- Depends on: `internal/client/*`, `internal/cron`, `internal/rag`, `internal/server/grpc`, `internal/wrappers`
- Used by: `main/main.go`

**Transport (gRPC):**
- Purpose: Expose API endpoints and translate between proto and domain models.
- Location: `internal/server/grpc/arxiv.go`
- Contains: gRPC handlers, validation, error mapping, proto conversion.
- Depends on: `internal/rag`, `internal/client/arxiv`, `internal/gen/arxiv/v1`
- Used by: `internal/app/app.go`

**Domain / Service:**
- Purpose: RAG ask pipeline (retrieval + generation) and error semantics.
- Location: `internal/rag/ask_pipeline.go`
- Contains: request validation, retrieval orchestration, context building, error classification.
- Depends on: `internal/client/arxiv` (via interface), generator interface.
- Used by: `internal/app/app.go`, `internal/server/grpc/arxiv.go`

**Scheduled Jobs:**
- Purpose: Periodic arXiv fetch and notification flow.
- Location: `internal/cron/arxiv_fetcher.go`
- Contains: topic list, fetch loop, notification formatting.
- Depends on: `internal/client/arxiv`, `internal/client/telegram`, `internal/wrappers`
- Used by: `internal/app/app.go`

**External Clients (Adapters):**
- Purpose: Integrate external APIs (arXiv, Groq, Telegram).
- Location: `internal/client/arxiv/*`, `internal/client/groq/*`, `internal/client/telegram/*`
- Contains: HTTP clients, request/response handling, config options.
- Depends on: stdlib HTTP and config helpers.
- Used by: `internal/app/app.go`, `internal/cron/arxiv_fetcher.go`, `internal/rag/ask_pipeline.go`

**Shared Utilities:**
- Purpose: Cross-cutting helpers (rate limiting).
- Location: `internal/wrappers/ratelimit.go`
- Contains: rate limiter wrapper with context-aware execution.
- Depends on: `golang.org/x/time/rate`
- Used by: `internal/app/app.go`, `internal/cron/arxiv_fetcher.go`

**Generated Code:**
- Purpose: gRPC/proto bindings.
- Location: `internal/gen/arxiv/v1/arxiv.pb.go`, `internal/gen/arxiv/v1/arxiv_grpc.pb.go`
- Contains: protobuf types and gRPC server interfaces.
- Depends on: `proto/arxiv/v1/arxiv.proto`
- Used by: `internal/server/grpc/arxiv.go`, `internal/app/app.go`

## Data Flow

**gRPC Ask Flow:**

1. `main/main.go` builds config/logger and runs `internal/app/app.go`.
2. `internal/app/app.go` starts gRPC server and registers `internal/server/grpc/arxiv.go`.
3. `internal/server/grpc/arxiv.go` validates request and calls `internal/rag/ask_pipeline.go`.
4. `internal/rag/ask_pipeline.go` calls `internal/client/arxiv/client.go` to retrieve papers.
5. `internal/rag/ask_pipeline.go` calls `internal/client/groq/client.go` to generate answer.
6. `internal/server/grpc/arxiv.go` maps result to proto and returns.

**gRPC Search Flow:**

1. `internal/server/grpc/arxiv.go` validates request.
2. `internal/server/grpc/arxiv.go` calls `internal/client/arxiv/client.go`.
3. Results are mapped to proto and returned.

**Scheduled Fetch Flow:**

1. `internal/app/app.go` schedules `internal/cron/arxiv_fetcher.go` via gocron.
2. `internal/cron/arxiv_fetcher.go` queries `internal/client/arxiv/client.go` per topic.
3. `internal/cron/arxiv_fetcher.go` formats notifications and sends via `internal/client/telegram/client.go`.
4. Rate limiting enforced via `internal/wrappers/ratelimit.go`.

**State Management:**
- Stateless services with in-memory config and logger only.
- arXiv PDF downloads cached on disk under `.cache/pdfs` in `internal/client/arxiv/client.go`.

## Key Abstractions

**AskService:**
- Purpose: Retrieval + generation orchestration for RAG.
- Examples: `internal/rag/ask_pipeline.go`
- Pattern: Interface-based dependencies (`Retriever`, `Generator`) for testability.

**ArxivHandler:**
- Purpose: Transport adapter for gRPC.
- Examples: `internal/server/grpc/arxiv.go`
- Pattern: Thin handler that validates and delegates to services/clients.

**Clients (Adapters):**
- Purpose: External API integration.
- Examples: `internal/client/arxiv/client.go`, `internal/client/groq/client.go`, `internal/client/telegram/client.go`
- Pattern: Configurable HTTP clients with option functions.

## Entry Points

**Binary Main:**
- Location: `main/main.go`
- Triggers: Process start.
- Responsibilities: Load config, init logger, run app, handle signals.

**gRPC Server:**
- Location: `internal/app/app.go`
- Triggers: `grpc.NewServer()` and `Serve` on `Config.GRPCAddress`.
- Responsibilities: Register `ArxivService` and serve requests.

**HTTP Health Server:**
- Location: `internal/app/app.go`
- Triggers: `http.Server` on `Config.Address`.
- Responsibilities: Serve `/health` endpoint.

**Scheduled Job:**
- Location: `internal/app/app.go`
- Triggers: gocron job `0 5 * * *`.
- Responsibilities: Run `internal/cron/arxiv_fetcher.go`.

## Error Handling

**Strategy:** Propagate errors with context and map to gRPC status codes at transport boundary.

**Patterns:**
- Domain errors in `internal/rag/ask_pipeline.go` mapped in `internal/server/grpc/arxiv.go`.
- Startup validation in `internal/app/config.go` with explicit missing-config errors.

## Cross-Cutting Concerns

**Logging:** zap `SugaredLogger` used across modules, configured in `main/main.go`.
**Validation:** Config validation in `internal/app/config.go`, request validation in `internal/server/grpc/arxiv.go`.
**Authentication:** Not implemented; no auth middleware or tokens in handlers.

---

*Architecture analysis: 2026-03-31*
