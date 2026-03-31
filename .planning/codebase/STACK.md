# Technology Stack

**Analysis Date:** 2026-03-31

## Languages

**Primary:**
- Go 1.26 - Application source in `main/main.go` and `internal/**/*.go`

**Secondary:**
- Protocol Buffers - Service contracts in `proto/` with generated code in `internal/gen/arxiv/v1/`

## Runtime

**Environment:**
- Go toolchain 1.26 (`go.mod`, `Dockerfile`)

**Package Manager:**
- Go modules - `go.mod` / `go.sum`
- Lockfile: present (`go.sum`)

## Frameworks

**Core:**
- `github.com/go-chi/chi/v5` v5.2.5 - HTTP routing (`internal/app/app.go`)
- `google.golang.org/grpc` v1.79.2 - gRPC server (`internal/app/app.go`, `internal/server/grpc/arxiv.go`)
- `github.com/go-co-op/gocron/v2` v2.19.1 - scheduled jobs (`internal/app/app.go`, `internal/cron/arxiv_fetcher.go`)

**Testing:**
- Go standard library `testing` - tests in `internal/**/*_test.go`

**Build/Dev:**
- Buf (v2 config) - Protobuf lint/breaking/generation (`buf.yaml`, `buf.gen.yaml`)
- `protoc-gen-go`, `protoc-gen-go-grpc` - code generation (`buf.gen.yaml`)

## Key Dependencies

**Critical:**
- `github.com/kelseyhightower/envconfig` v1.4.0 - runtime config from env vars (`main/main.go`, `internal/app/config.go`)
- `go.uber.org/zap` v1.27.1 - structured logging (`main/main.go`, `internal/app/app.go`)
- `golang.org/x/sync` v0.19.0 - errgroup for concurrency (`internal/app/app.go`)
- `golang.org/x/time` v0.14.0 - rate limiting utilities (`internal/wrappers/ratelimit.go`)

**Infrastructure:**
- `google.golang.org/protobuf` v1.36.11 - protobuf runtime (`internal/gen/arxiv/v1/arxiv.pb.go`)

## Configuration

**Environment:**
- Env vars via `envconfig.Process("arxiv-rag-go", &cfg)` (`main/main.go`)
- Example env file: `.env.example` (used by `docker-compose.yml` via `env_file: .env`)
- Required runtime keys validated in `internal/app/config.go`

**Build:**
- Docker multi-stage build (`Dockerfile`)
- Protobuf build configs (`buf.yaml`, `buf.gen.yaml`)

## Platform Requirements

**Development:**
- Go 1.26 toolchain
- Buf + protobuf plugins if regenerating gRPC code (`buf.gen.yaml`)

**Production:**
- Container image built from `Dockerfile` (scratch + CA certs)

---

*Stack analysis: 2026-03-31*
