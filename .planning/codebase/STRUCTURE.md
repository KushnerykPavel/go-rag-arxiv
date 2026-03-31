# Codebase Structure

**Analysis Date:** 2026-03-31

## Directory Layout

```
[project-root]/
├── main/             # Binary entrypoint
├── internal/         # Application code (non-exported)
├── proto/            # Protobuf definitions
├── vendor/           # Vendored dependencies
├── go.mod            # Go module definition
├── go.sum            # Go module checksums
├── Dockerfile        # Container build
├── docker-compose.yml# Local orchestration
├── buf.yaml          # Protobuf tooling config
└── buf.gen.yaml      # Protobuf generation config
```

## Directory Purposes

**main:**
- Purpose: Application entrypoint and process lifecycle.
- Contains: `main/main.go`
- Key files: `main/main.go`

**internal:**
- Purpose: All application modules and adapters.
- Contains: app wiring, server handlers, domain logic, clients, cron jobs, wrappers, generated code.
- Key files: `internal/app/app.go`, `internal/server/grpc/arxiv.go`, `internal/rag/ask_pipeline.go`

**internal/app:**
- Purpose: Dependency wiring and lifecycle orchestration.
- Contains: `internal/app/app.go`, `internal/app/config.go`
- Key files: `internal/app/app.go`, `internal/app/config.go`

**internal/server/grpc:**
- Purpose: gRPC transport handlers and proto mapping.
- Contains: `internal/server/grpc/arxiv.go`
- Key files: `internal/server/grpc/arxiv.go`

**internal/rag:**
- Purpose: RAG ask pipeline and error semantics.
- Contains: `internal/rag/ask_pipeline.go`
- Key files: `internal/rag/ask_pipeline.go`

**internal/client:**
- Purpose: External API clients (arXiv, Groq, Telegram).
- Contains: `internal/client/arxiv/*`, `internal/client/groq/*`, `internal/client/telegram/*`
- Key files: `internal/client/arxiv/client.go`, `internal/client/groq/client.go`, `internal/client/telegram/client.go`

**internal/cron:**
- Purpose: Scheduled jobs.
- Contains: `internal/cron/arxiv_fetcher.go`
- Key files: `internal/cron/arxiv_fetcher.go`

**internal/wrappers:**
- Purpose: Shared utilities and wrappers.
- Contains: `internal/wrappers/ratelimit.go`
- Key files: `internal/wrappers/ratelimit.go`

**internal/gen:**
- Purpose: Generated protobuf and gRPC code.
- Contains: `internal/gen/arxiv/v1/arxiv.pb.go`, `internal/gen/arxiv/v1/arxiv_grpc.pb.go`
- Key files: `internal/gen/arxiv/v1/arxiv.pb.go`, `internal/gen/arxiv/v1/arxiv_grpc.pb.go`

**proto:**
- Purpose: Protobuf source definitions.
- Contains: `proto/arxiv/v1/arxiv.proto`
- Key files: `proto/arxiv/v1/arxiv.proto`

## Key File Locations

**Entry Points:**
- `main/main.go`: Process entrypoint.

**Configuration:**
- `internal/app/config.go`: Runtime config and validation.
- `buf.yaml`: Protobuf configuration.
- `buf.gen.yaml`: Protobuf generation configuration.
- `docker-compose.yml`: Local service orchestration.

**Core Logic:**
- `internal/app/app.go`: Application wiring and lifecycle.
- `internal/rag/ask_pipeline.go`: RAG ask pipeline.
- `internal/server/grpc/arxiv.go`: gRPC API handlers.

**Testing:**
- `internal/app/config_test.go`
- `internal/app/startup_validation_test.go`
- `internal/client/arxiv/client_test.go`
- `internal/server/grpc/arxiv_ask_test.go`
- `internal/server/grpc/arxiv_contract_test.go`

## Naming Conventions

**Files:**
- Go source files use `lower_snake_case.go` (e.g., `internal/cron/arxiv_fetcher.go`).
- Test files use `*_test.go` (e.g., `internal/client/arxiv/client_test.go`).
- Protobuf files use `lower_snake_case.proto` (e.g., `proto/arxiv/v1/arxiv.proto`).

**Directories:**
- Go packages use lowercase directory names (e.g., `internal/rag`, `internal/server/grpc`).

## Where to Add New Code

**New Feature:**
- Primary code: `internal/rag` for domain logic, `internal/server/grpc` for API exposure.
- Tests: co-located `*_test.go` alongside new files (e.g., `internal/rag/new_feature_test.go`).

**New Component/Module:**
- Implementation: `internal/<module>` with an exported constructor and internal interfaces.

**Utilities:**
- Shared helpers: `internal/wrappers`

## Special Directories

**vendor:**
- Purpose: Vendored Go dependencies.
- Generated: Yes.
- Committed: Yes.

**internal/gen:**
- Purpose: Generated protobuf and gRPC code.
- Generated: Yes.
- Committed: Yes.

---

*Structure analysis: 2026-03-31*
