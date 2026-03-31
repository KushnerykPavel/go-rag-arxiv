# Technology Stack

**Analysis Date:** 2026-03-31

## Languages

**Primary:**
- Go 1.26 - application code and build target in `go.mod`, `main/main.go`, and `internal/**/*.go`.

**Secondary:**
- Protocol Buffers (proto3) - gRPC contract definitions in `proto/arxiv/v1/arxiv.proto`.
- YAML - protobuf generation/lint configuration in `buf.gen.yaml` and `buf.yaml`; container orchestration config in `docker-compose.yml`.
- Dockerfile syntax - container build/runtime definition in `Dockerfile`.

## Runtime

**Environment:**
- Go runtime, target toolchain `go 1.26` from `go.mod`.
- Linux container runtime for production image (`FROM scratch`) in `Dockerfile`.

**Package Manager:**
- Go Modules via `go.mod`/`go.sum`.
- Lockfile: present (`go.sum`).

## Frameworks

**Core:**
- `github.com/go-chi/chi/v5` (`go.mod`) - HTTP routing and middleware wiring in `internal/app/app.go`.
- `google.golang.org/grpc` (`go.mod`) - gRPC server and generated protobuf service stubs in `internal/app/app.go`, `internal/server/grpc/arxiv.go`, `internal/gen/arxiv/v1/arxiv_grpc.pb.go`.
- `github.com/go-co-op/gocron/v2` (`go.mod`) - scheduled daily jobs in `internal/app/app.go`.

**Testing:**
- Go standard `testing` package - test file `internal/client/arxiv/client_test.go`.

**Build/Dev:**
- `buf` (configured in `buf.yaml` and `buf.gen.yaml`) - protobuf linting and Go/gRPC stub generation to `internal/gen/`.
- Docker multi-stage build (`Dockerfile`) - reproducible binary build and minimal runtime image.

## Key Dependencies

**Critical:**
- `github.com/kelseyhightower/envconfig` (`go.mod`) - environment-driven config loading in `main/main.go` and struct tags in `internal/app/config.go`.
- `go.uber.org/zap` (`go.mod`) - structured logging setup in `main/main.go` and app/client modules such as `internal/app/app.go`.
- `golang.org/x/sync/errgroup` (`go.mod`) - coordinated goroutine lifecycle in `internal/app/app.go`.
- `golang.org/x/time/rate` (`go.mod`) - rate limiting wrapper in `internal/wrappers/ratelimit.go`.

**Infrastructure:**
- `google.golang.org/protobuf` (`go.mod`) - protobuf runtime used by generated code in `internal/gen/arxiv/v1/arxiv.pb.go`.
- Standard library `net/http` - outbound API clients and inbound health endpoint in `internal/client/arxiv/client.go`, `internal/client/telegram/client.go`, and `internal/app/app.go`.

## Configuration

**Environment:**
- App config is loaded via `envconfig.Process("arxiv-rag-go", &cfg)` in `main/main.go`.
- Required/optional variables are declared in `internal/app/config.go`:
- `ADDRESS` (default `:8080`)
- `GRPC_ADDRESS` (default `:9090`)
- `GROQ_API_KEY` (required by config; no current call site found in `internal/**/*.go`)
- `TELEGRAM_TOKEN` (required)
- `TELEGRAM_CHAT_ID` (required)
- `.env.example` is present at repository root; contents intentionally not read.
- Docker Compose references an env file (`env_file: .env`) and port mapping in `docker-compose.yml`.

**Build:**
- Protobuf build config files: `buf.yaml`, `buf.gen.yaml`.
- Container build config file: `Dockerfile`.
- Go module/build metadata: `go.mod`, `go.sum`.

## Platform Requirements

**Development:**
- Go 1.26 toolchain (`go.mod`).
- `buf` CLI for lint/generation workflows referenced in `CLAUDE.md` and configured in `buf.yaml`/`buf.gen.yaml`.
- Network access required for outbound HTTPS integrations used in `internal/client/arxiv/client.go` and `internal/client/telegram/client.go`.

**Production:**
- Container-capable environment that can run the image built from `Dockerfile`.
- HTTPS trust store is required and embedded by copying CA certs in `Dockerfile`.
- Exposed HTTP health port and gRPC port configured by `ADDRESS`/`GRPC_ADDRESS` in `internal/app/config.go`.

---

*Stack analysis: 2026-03-31*
