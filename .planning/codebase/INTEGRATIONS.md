# External Integrations

**Analysis Date:** 2026-03-31

## APIs & External Services

**Research & Content:**
- arXiv API - paper search and PDF retrieval (`internal/client/arxiv/client.go`, `internal/client/arxiv/config.go`)
  - SDK/Client: custom HTTP client in `internal/client/arxiv/client.go`
  - Auth: none

**LLM Generation:**
- Groq OpenAI-compatible API - answer generation (`internal/client/groq/client.go`, `internal/client/groq/config.go`)
  - SDK/Client: custom HTTP client in `internal/client/groq/client.go`
  - Auth: `GROQ_API_KEY` (`internal/app/config.go`, `.env.example`)

**Notifications:**
- Telegram Bot API - operational messages and digests (`internal/client/telegram/client.go`, `internal/cron/arxiv_fetcher.go`)
  - SDK/Client: custom HTTP client in `internal/client/telegram/client.go`
  - Auth: `TELEGRAM_TOKEN` (`internal/app/config.go`, `.env.example`)

## Data Storage

**Databases:**
- Not detected

**File Storage:**
- Local filesystem cache for PDFs at `.cache/pdfs` (`internal/client/arxiv/config.go`, `internal/client/arxiv/client.go`)

**Caching:**
- None (beyond local PDF cache)

## Authentication & Identity

**Auth Provider:**
- Not detected (no user auth; service uses API keys for Groq and Telegram)

## Monitoring & Observability

**Error Tracking:**
- None detected

**Logs:**
- Zap structured logging to stdout (`main/main.go`)

## CI/CD & Deployment

**Hosting:**
- Docker container (`Dockerfile`, `docker-compose.yml`)

**CI Pipeline:**
- None detected

## Environment Configuration

**Required env vars:**
- `GROQ_API_KEY` (`internal/app/config.go`, `.env.example`)
- `TELEGRAM_TOKEN` (`internal/app/config.go`, `.env.example`)
- `TELEGRAM_CHAT_ID` (`internal/app/config.go`, `.env.example`)
- `ADDRESS` (`internal/app/config.go`, `.env.example`)
- `GRPC_ADDRESS` (`internal/app/config.go`)

**Secrets location:**
- Runtime environment variables (example file `.env.example`)

## Webhooks & Callbacks

**Incoming:**
- None detected

**Outgoing:**
- None detected

---

*Integration audit: 2026-03-31*
