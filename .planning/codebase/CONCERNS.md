# Codebase Concerns

**Analysis Date:** 2026-03-31

## Tech Debt

**Hardcoded scheduling and topics:**
- Issue: Fetch topics and schedule are hardcoded rather than configurable.
- Files: `internal/cron/arxiv_fetcher.go`, `internal/app/app.go`
- Impact: Requires code changes for topic/schedule updates; makes ops changes risky.
- Fix approach: Move topics and schedule cron string into config/env and validate at startup.

**No pagination for arXiv daily fetch:**
- Issue: The daily fetch requests a single page with `MaxResults` and does not paginate.
- Files: `internal/cron/arxiv_fetcher.go`, `internal/client/arxiv/client.go`
- Impact: Results are truncated when the topic exceeds `maxResultsCap`, silently missing papers.
- Fix approach: Implement pagination with `start` and loop until fewer than `maxResults` results.

**Unbounded PDF cache growth:**
- Issue: Downloaded PDFs are cached without eviction or size limits.
- Files: `internal/client/arxiv/client.go`, `internal/client/arxiv/config.go`
- Impact: Disk usage grows over time; cache may fill disk on long-running deployments.
- Fix approach: Add TTL/size-based eviction or externalize cache to managed storage.

## Known Bugs

**Rate limiter uses minutes instead of seconds:**
- Symptoms: With `rps=1`, limiter allows ~1 request per minute, not per second.
- Files: `internal/wrappers/ratelimit.go`
- Trigger: Any call to `NewRateLimiter(1)` or similar.
- Workaround: None in code; adjust limiter math to `rate.Every(time.Second / time.Duration(rps))`.

**Server error does not shut down app:**
- Symptoms: If HTTP or gRPC server goroutine returns an error, the other continues running and app does not exit.
- Files: `internal/app/app.go`
- Trigger: gRPC listener failure or HTTP server error.
- Workaround: Manually terminate process; use errgroup context and cancel on first error.

**Telegram client ignores HTTP status codes:**
- Symptoms: Non-200 responses can result in JSON decode errors or misleading errors.
- Files: `internal/client/telegram/client.go`
- Trigger: Telegram API errors or network issues returning non-JSON responses.
- Workaround: Add status checks and handle non-JSON error bodies.

## Security Considerations

**Unauthenticated HTTP and gRPC endpoints:**
- Risk: Public access to search/ask endpoints without auth or rate limiting.
- Files: `internal/app/app.go`, `internal/server/grpc/arxiv.go`
- Current mitigation: None in code.
- Recommendations: Add TLS termination, authentication, and rate limiting at the server or gateway.

**Unescaped HTML in Telegram messages:**
- Risk: arXiv data can include characters that break HTML formatting or inject markup.
- Files: `internal/cron/arxiv_fetcher.go`
- Current mitigation: None.
- Recommendations: Escape HTML entities before formatting (title/authors/URL).

## Performance Bottlenecks

**Serial notifications for large result sets:**
- Problem: Each paper triggers a separate Telegram API call, processed serially.
- Files: `internal/cron/arxiv_fetcher.go`, `internal/wrappers/ratelimit.go`
- Cause: Per-paper send loop plus strict limiter.
- Improvement path: Batch papers into fewer messages and fix limiter math.

## Fragile Areas

**Time window based on last 24 hours:**
- Files: `internal/cron/arxiv_fetcher.go`, `internal/app/app.go`
- Why fragile: If the job is delayed or the service is down for more than 24 hours, papers are skipped.
- Safe modification: Persist last successful fetch timestamp and query from that point.
- Test coverage: No tests for time window behavior.

**Ask query expects raw arXiv syntax:**
- Files: `internal/rag/ask_pipeline.go`, `internal/server/grpc/arxiv.go`
- Why fragile: Natural-language queries may return zero results, causing user-visible errors.
- Safe modification: Add query normalization or a keyword-to-arXiv query builder.
- Test coverage: No tests for query normalization or fallback strategies.

## Scaling Limits

**Hard cap on fetch size:**
- Current capacity: `maxResultsCap = 2000` per request.
- Limit: High-volume categories can exceed the cap in a day.
- Scaling path: Paginate or partition by time slices.
- Files: `internal/client/arxiv/config.go`, `internal/client/arxiv/client.go`

## Dependencies at Risk

**Go toolchain version set to a future release:**
- Risk: `go 1.26` may be unavailable in build environments, blocking builds.
- Impact: CI/CD and local builds fail unless toolchain exists.
- Migration plan: Pin to a released Go version and verify compatibility.
- Files: `go.mod`

## Missing Critical Features

**No retry/backoff for Groq or Telegram API failures:**
- Problem: Transient upstream issues cause immediate request failures.
- Blocks: Reliable daily notifications and ask responses during API outages.
- Files: `internal/client/groq/client.go`, `internal/client/telegram/client.go`

## Test Coverage Gaps

**Network-only arXiv test without assertions:**
- What's not tested: Client behavior on errors, pagination, and response parsing.
- Files: `internal/client/arxiv/client_test.go`
- Risk: Flaky tests and undetected parsing regressions.
- Priority: High

**No tests for cron, rate limiting, and notification formatting:**
- What's not tested: Fetch scheduling, limiter behavior, and HTML escaping.
- Files: `internal/cron/arxiv_fetcher.go`, `internal/wrappers/ratelimit.go`, `internal/client/telegram/client.go`
- Risk: Production failures are not caught in CI.
- Priority: Medium

---

*Concerns audit: 2026-03-31*
