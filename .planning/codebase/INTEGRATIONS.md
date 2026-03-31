# External Integrations

**Analysis Date:** 2026-03-31

## APIs & External Services

**Academic Content API:**
- arXiv API (`https://export.arxiv.org/api/query`) - paper search/fetch and metadata retrieval in `internal/client/arxiv/config.go` and `internal/client/arxiv/client.go`.
  - SDK/Client: custom HTTP client in `internal/client/arxiv/client.go` (uses stdlib `net/http`).
  - Auth: none detected in request construction (`internal/client/arxiv/client.go`).

**Messaging/Notification API:**
- Telegram Bot API (`https://api.telegram.org`) - sends startup/shutdown and paper notification messages in `internal/client/telegram/config.go`, `internal/client/telegram/client.go`, and `internal/app/app.go`.
  - SDK/Client: custom HTTP client in `internal/client/telegram/client.go` (uses stdlib `net/http`).
  - Auth: bot token passed in URL path; env source is `TELEGRAM_TOKEN` in `internal/app/config.go`.

**Internal Service API:**
- gRPC ArxivService - internal/external consumers can call `Search` over TCP address configured via `GRPC_ADDRESS` in `internal/app/config.go`; service contract in `proto/arxiv/v1/arxiv.proto`; server implementation in `internal/server/grpc/arxiv.go`.
  - SDK/Client: `google.golang.org/grpc` and generated code in `internal/gen/arxiv/v1/arxiv_grpc.pb.go`.
  - Auth: Not detected (no interceptors or auth middleware in `internal/app/app.go`).

## Data Storage

**Databases:**
- Not detected. No SQL/NoSQL client usage in `internal/**/*.go`; no DB dependency in `go.mod`.
  - Connection: Not applicable.
  - Client: Not applicable.

**File Storage:**
- Local filesystem cache for downloaded PDFs under `.cache/pdfs` configured in `internal/client/arxiv/config.go` and written in `internal/client/arxiv/client.go`.

**Caching:**
- Local file-based cache only (PDF existence check via `os.Stat` in `internal/client/arxiv/client.go`).
- No Redis/Memcached/external cache provider detected in `go.mod` or `internal/**/*.go`.

## Authentication & Identity

**Auth Provider:**
- Custom token-based outbound auth only for Telegram API.
  - Implementation: Bot token injected from env config (`TELEGRAM_TOKEN` in `internal/app/config.go`) and used to build endpoint URL in `internal/client/telegram/client.go`.

## Monitoring & Observability

**Error Tracking:**
- None detected (no Sentry/Rollbar/etc. dependency in `go.mod`).

**Logs:**
- Structured logging with Zap (`go.uber.org/zap`) initialized in `main/main.go` and used across modules (`internal/app/app.go`, `internal/client/arxiv/client.go`, `internal/client/telegram/client.go`, `internal/server/grpc/arxiv.go`).

## CI/CD & Deployment

**Hosting:**
- Container deployment path is defined by `Dockerfile` (multi-stage Go build to `scratch`) and local orchestration in `docker-compose.yml`.

**CI Pipeline:**
- Not detected (no `.github/workflows/*` or alternative CI configs found in scanned tree).

## Environment Configuration

**Required env vars:**
- `GROQ_API_KEY` (declared required in `internal/app/config.go`; current code path does not consume it in `internal/**/*.go`).
- `TELEGRAM_TOKEN` (required in `internal/app/config.go`; used in `internal/client/telegram/client.go`).
- `TELEGRAM_CHAT_ID` (required in `internal/app/config.go`; used in `internal/app/app.go` and `internal/cron/arxiv_fetcher.go`).

**Optional env vars:**
- `ADDRESS` default `:8080` in `internal/app/config.go`.
- `GRPC_ADDRESS` default `:9090` in `internal/app/config.go`.

**Secrets location:**
- Runtime environment variables loaded via `envconfig` in `main/main.go`.
- `.env` is referenced by `docker-compose.yml`; `.env.example` exists in repository root.

## Webhooks & Callbacks

**Incoming:**
- None detected. HTTP server exposes only `/health` in `internal/app/app.go`; gRPC service provides request/response RPC endpoints in `proto/arxiv/v1/arxiv.proto`.

**Outgoing:**
- HTTP GET requests to arXiv API in `internal/client/arxiv/client.go`.
- HTTP POST requests to Telegram `sendMessage` endpoint in `internal/client/telegram/client.go`.

---

*Integration audit: 2026-03-31*
