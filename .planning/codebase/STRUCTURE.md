# Codebase Structure

**Analysis Date:** 2026-03-31

## Directory Layout

```text
go-rag-arxiv/
├── main/                  # Executable entrypoint package
├── internal/
│   ├── app/               # Application composition root and runtime config
│   ├── client/
│   │   ├── arxiv/         # arXiv API client, models, and options
│   │   └── telegram/      # Telegram Bot API client and options
│   ├── cron/              # Scheduled workflow jobs
│   ├── server/grpc/       # gRPC transport handlers
│   ├── wrappers/          # Shared utility wrappers (rate limiting)
│   └── gen/arxiv/v1/      # Generated protobuf/go-grpc code
├── proto/arxiv/v1/        # Source protobuf contracts
├── .planning/codebase/    # Mapper output documents
├── Dockerfile             # Multi-stage container build
├── docker-compose.yml     # Local service run definition
├── go.mod                 # Go module and dependency definitions
├── buf.yaml               # Proto lint/breaking config
└── buf.gen.yaml           # Proto generation config targeting internal/gen
```

## Directory Purposes

**`main/`:**
- Purpose: Host executable startup package.
- Contains: `main/main.go`.
- Key files: `main/main.go` (env config, logger, signal handling, `app.Run` invocation).

**`internal/app/`:**
- Purpose: Central runtime assembly and lifecycle coordination.
- Contains: App container, environment-backed config structs.
- Key files: `internal/app/app.go`, `internal/app/config.go`.

**`internal/client/arxiv/`:**
- Purpose: External arXiv integration and parsing.
- Contains: HTTP client, option pattern config, XML parsing helpers, `Paper` model.
- Key files: `internal/client/arxiv/client.go`, `internal/client/arxiv/config.go`, `internal/client/arxiv/paper.go`, `internal/client/arxiv/client_test.go`.

**`internal/client/telegram/`:**
- Purpose: Outbound Telegram notifications.
- Contains: HTTP bot client and option pattern config.
- Key files: `internal/client/telegram/client.go`, `internal/client/telegram/config.go`.

**`internal/cron/`:**
- Purpose: Scheduled business workflow.
- Contains: `ArxivFetcher` job that queries papers and sends notifications.
- Key files: `internal/cron/arxiv_fetcher.go`.

**`internal/server/grpc/`:**
- Purpose: Inbound gRPC request handling.
- Contains: `ArxivHandler` implementation for service methods.
- Key files: `internal/server/grpc/arxiv.go`.

**`internal/gen/arxiv/v1/`:**
- Purpose: Generated code from protobuf definitions.
- Contains: gRPC and protobuf generated files.
- Key files: `internal/gen/arxiv/v1/arxiv.pb.go`, `internal/gen/arxiv/v1/arxiv_grpc.pb.go`.

**`proto/arxiv/v1/`:**
- Purpose: Source-of-truth API contract definitions.
- Contains: Service and message declarations.
- Key files: `proto/arxiv/v1/arxiv.proto`.

## Key File Locations

**Entry Points:**
- `main/main.go`: Process bootstrap and lifecycle root.
- `internal/app/app.go`: Runtime entry for HTTP server, gRPC server, scheduler.

**Configuration:**
- `internal/app/config.go`: App runtime env-config schema (`ADDRESS`, `GRPC_ADDRESS`, Telegram token/chat ID, Groq key).
- `buf.yaml`: Proto lint and breaking-change policy.
- `buf.gen.yaml`: Proto generation targets (`internal/gen`).
- `Dockerfile`: Build/runtime packaging.
- `docker-compose.yml`: Container runtime wiring via `.env` file reference.
- `.env.example`: Example environment variable shape.
- `.env`: Present (environment configuration; do not commit secrets).

**Core Logic:**
- `internal/cron/arxiv_fetcher.go`: Scheduled fetch + notify loop.
- `internal/server/grpc/arxiv.go`: gRPC `Search` method behavior.
- `internal/client/arxiv/client.go`: arXiv query execution, XML parsing, PDF download/cache.
- `internal/client/telegram/client.go`: Telegram message delivery.
- `internal/wrappers/ratelimit.go`: Reusable throttling wrapper.

**Testing:**
- `internal/client/arxiv/client_test.go`: Current test location/pattern.

## Naming Conventions

**Files:**
- Lowercase snake_case for multiword files: `arxiv_fetcher.go`.
- Package-aligned simple names for core files: `client.go`, `config.go`, `app.go`.
- Generated protobuf files follow plugin convention: `*.pb.go`, `*_grpc.pb.go` in `internal/gen/arxiv/v1/`.

**Directories:**
- Domain-first package grouping under `internal/` (`app`, `client`, `cron`, `server`, `wrappers`, `gen`).
- Transport subtype as nested directory: `internal/server/grpc/`.
- API versioning in contract paths: `proto/arxiv/v1/` and mirrored `internal/gen/arxiv/v1/`.

## Where to Add New Code

**New Feature:**
- Primary code: Add orchestration in `internal/app/`; put feature logic in focused package under `internal/` (for example `internal/cron/` for scheduled jobs or `internal/server/grpc/` for RPC handlers).
- Tests: Co-locate tests next to implementation as `*_test.go` (existing precedent: `internal/client/arxiv/client_test.go`).

**New Component/Module:**
- Implementation: Use `internal/<component>/` with package-local `config.go` + `client.go` or equivalent constructor-centric file pattern, mirroring `internal/client/arxiv/` and `internal/client/telegram/`.

**Utilities:**
- Shared helpers: Place reusable wrappers in `internal/wrappers/` when cross-package and non-domain-specific.

## Special Directories

**`internal/gen/`:**
- Purpose: Generated protobuf and gRPC stubs emitted by Buf/protoc tooling.
- Generated: Yes (configured in `buf.gen.yaml`).
- Committed: Yes (present in repository under `internal/gen/arxiv/v1/`).

**`vendor/`:**
- Purpose: Vendored module dependencies.
- Generated: Yes (from `go mod vendor` workflow).
- Committed: No (ignored via `.gitignore` entry `vendor/`).

**`.planning/`:**
- Purpose: Planning and mapper artifacts used by GSD workflow.
- Generated: Yes (workflow-generated documents and plans).
- Committed: Project-specific process choice; currently present in workspace.

---

*Structure analysis: 2026-03-31*
