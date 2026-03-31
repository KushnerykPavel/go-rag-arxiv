# Codebase Concerns

**Analysis Date:** 2026-03-31

## Tech Debt

**Service contract and implementation drift (`Ask` RPC declared but not implemented):**
- Issue: The gRPC contract exposes `Ask`, but the server handler only implements `Search`.
- Files: `proto/arxiv/v1/arxiv.proto`, `internal/server/grpc/arxiv.go`, `internal/gen/arxiv/v1/arxiv_grpc.pb.go`
- Impact: Clients can compile against `Ask` and receive `Unimplemented` at runtime, creating a broken public API surface.
- Fix approach: Implement `Ask` in `internal/server/grpc/arxiv.go` with retrieval+LLM orchestration, or remove `Ask` from `proto/arxiv/v1/arxiv.proto` until implemented.

**Required configuration not used in runtime path:**
- Issue: `GROQ_API_KEY` is mandatory in config but no code path uses it.
- Files: `internal/app/config.go`, `main/main.go`, `internal/app/app.go`
- Impact: Startup can fail for a credential that has no functional effect, increasing operational friction and misconfiguration noise.
- Fix approach: Either wire Groq usage into runtime (e.g., `Ask` pipeline) or make the field optional until consumed.

**Application wiring is centralized and tightly coupled:**
- Issue: `App.Run` creates scheduler, HTTP server, gRPC server, arXiv client, Telegram client, and cron job directly in one method.
- Files: `internal/app/app.go`
- Impact: Feature changes require broad edits in a single file; unit testing lifecycle/error branches is difficult.
- Fix approach: Split construction into focused modules (`transport`, `jobs`, `clients`) and inject interfaces into `App`.

## Known Bugs

**Rate limiter behavior does not match API contract/comments:**
- Symptoms: `NewRateLimiter(1)` enforces about one call per minute, while comments say “rps per second”.
- Files: `internal/wrappers/ratelimit.go`
- Trigger: Any call path using `wrappers.NewRateLimiter(1)` such as notifications in `internal/app/app.go`.
- Workaround: Pass a larger value experimentally, but behavior remains non-obvious and unit-mismatched.

**“Application stopped” notification is sent with canceled context:**
- Symptoms: Stop notification can fail or be dropped after shutdown starts.
- Files: `internal/app/app.go`
- Trigger: Normal shutdown flow after `<-ctx.Done()`; send is done with `ctx` that is already canceled.
- Workaround: Send stop notification with a fresh timeout context before final cancellation, or tolerate omission explicitly.

## Security Considerations

**No authentication/authorization on exposed HTTP and gRPC endpoints:**
- Risk: Any network client that can reach the process can invoke `/health` and gRPC `Search`.
- Files: `internal/app/app.go`, `internal/server/grpc/arxiv.go`
- Current mitigation: None detected in middleware/interceptors.
- Recommendations: Restrict bind interfaces, add authn/authz middleware/interceptors, and enforce network policy.

**No transport security configured for gRPC listener:**
- Risk: gRPC traffic is served without TLS by default.
- Files: `internal/app/app.go`
- Current mitigation: None detected (`grpc.NewServer()` without credentials).
- Recommendations: Use `credentials.NewServerTLSFromFile` (or mTLS) and rotate certs via secret management.

**Unescaped HTML content forwarded to Telegram messages:**
- Risk: Title/author text from external feed is interpolated into HTML payload; malformed or crafted content can alter rendering/phishing surface in chat.
- Files: `internal/cron/arxiv_fetcher.go`
- Current mitigation: None detected (raw `fmt.Sprintf` into HTML body).
- Recommendations: Escape all dynamic fields before `SendHTML`, or switch to plain text/Markdown with strict escaping.

## Performance Bottlenecks

**Notification fan-out throughput is effectively capped and can backlog for days:**
- Problem: Fetcher can request up to 2000 papers per topic and send one message per paper through a limiter that currently behaves as ~1/min.
- Files: `internal/cron/arxiv_fetcher.go`, `internal/wrappers/ratelimit.go`, `internal/app/app.go`
- Cause: High `MaxResults`, per-paper sends, and limiter interval mismatch.
- Improvement path: Fix limiter units, batch notifications, cap per-run sends, and persist checkpoint state for incremental processing.

**Single-request XML body is fully loaded into memory:**
- Problem: Entire API response body is read using `io.ReadAll`.
- Files: `internal/client/arxiv/client.go`
- Cause: Full-buffer parsing path.
- Improvement path: Stream parse with `xml.Decoder` and apply entry-level limits.

## Fragile Areas

**Runtime behavior depends on external networks during tests:**
- Files: `internal/client/arxiv/client_test.go`
- Why fragile: Test hits live arXiv API and has no deterministic fixtures or assertions.
- Safe modification: Introduce `httptest.Server`-based tests using `WithBaseURL`/`WithHTTPClient` and assert parsed outputs.
- Test coverage: Only one test file exists (`internal/client/arxiv/client_test.go`).

**Error handling path can terminate process abruptly:**
- Files: `main/main.go`
- Why fragile: `log.Fatal` and second-signal `os.Exit(1)` bypass normal cleanup semantics.
- Safe modification: Return errors through a controlled shutdown path and prefer context cancellation over forced exit.
- Test coverage: No tests for signal/shutdown behavior.

## Scaling Limits

**Single-process scheduler and notifier pipeline:**
- Current capacity: One in-process scheduler, one notifier flow, one limiter instance.
- Limit: Horizontal scaling duplicates daily jobs and can cause duplicate notifications.
- Scaling path: Use distributed job coordination/locking and idempotency keys per paper/topic/day.

**Local PDF cache only:**
- Current capacity: Cache rooted at local filesystem path `.cache/pdfs`.
- Limit: Multi-instance deployments do not share cache and repeat downloads.
- Scaling path: Move cache/index to shared object storage and deduplicate by arXiv ID.

## Dependencies at Risk

**Pinned generated protobuf code versions are older than module runtime versions:**
- Risk: Generated files indicate `protoc-gen-go v1.34.2`, while `go.mod` pulls newer protobuf runtime; drift can cause regeneration diffs and subtle incompatibility risk.
- Impact: Regenerating stubs in CI/dev may produce noisy churn or API differences.
- Migration plan: Standardize proto toolchain version in build tooling and regenerate `internal/gen/arxiv/v1/*` deterministically.

## Missing Critical Features

**RAG answer path is not implemented despite API and env contract:**
- Problem: API advertises `Ask`, and config requires `GROQ_API_KEY`, but no runtime LLM answer flow exists.
- Blocks: End-to-end “question answering over retrieved papers” cannot be delivered from current server implementation.

## Test Coverage Gaps

**Transport layer untested:**
- What's not tested: HTTP server (`/health`) and gRPC handler behavior (`Search` validation, error mapping, response transformation).
- Files: `internal/app/app.go`, `internal/server/grpc/arxiv.go`
- Risk: Regressions in API behavior and startup/shutdown handling can ship unnoticed.
- Priority: High

**Scheduler/cron behavior untested:**
- What's not tested: Daily job registration, topic iteration, retry/error behavior, and notifier invocation patterns.
- Files: `internal/app/app.go`, `internal/cron/arxiv_fetcher.go`
- Risk: Silent data loss or notification floods.
- Priority: High

**Client resilience paths untested:**
- What's not tested: Retry backoff, non-200 response handling, PDF cache behavior, and cancellation behavior.
- Files: `internal/client/arxiv/client.go`, `internal/client/telegram/client.go`
- Risk: Production failures under transient network errors are hard to predict.
- Priority: High

---

*Concerns audit: 2026-03-31*
