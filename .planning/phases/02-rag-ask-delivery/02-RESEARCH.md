# Phase 2: RAG Ask Delivery - Research

**Researched:** 2026-03-31
**Domain:** Go gRPC RAG pipeline over arXiv retrieval + LLM generation
**Confidence:** MEDIUM-HIGH

## User Constraints

### Locked Decisions
No `CONTEXT.md` found in `.planning/phases/02-rag-ask-delivery/`. No additional locked user decisions were provided beyond roadmap/requirements.

### Claude's Discretion
Implementation details for Ask contract shape, retrieval-to-generation orchestration, citation formatting, and downstream error mapping.

### Deferred Ideas (OUT OF SCOPE)
No deferred ideas documented for this phase.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| RAG-01 | User can call `Ask` over gRPC with a natural-language question and receive an answer payload. | Add `Ask` RPC to proto + generated stubs + handler implementation with strict request validation and non-empty response contract. |
| RAG-02 | `Ask` retrieves relevant arXiv papers before generating the response. | Enforce retrieval-first orchestration (`FetchPapersWithQuery`) before any LLM call; fail request if retrieval is empty/error. |
| RAG-03 | `Ask` returns citations (paper identifiers/titles/links) for answer traceability. | Build citations deterministically from retrieved `arxiv.Paper` metadata and return as structured fields in `AskResponse`. |
| APC-03 | gRPC errors are mapped to clear, stable status codes for validation and downstream failures. | Adopt explicit status mapping matrix (`InvalidArgument`, `DeadlineExceeded`, `Unavailable`, `ResourceExhausted`, `Internal`) with tests per failure mode. |
</phase_requirements>

## Summary

Phase 2 should be implemented as a transport-safe retrieval-augmented pipeline, not as a direct LLM call. The current repo has only `Search` in proto and server implementation; `Ask` does not exist yet in contract or runtime. Because generated stubs are source-of-truth at runtime, `Ask` requires coordinated updates in `proto/arxiv/v1/arxiv.proto`, `buf generate`, handler code, and contract tests.

Grounding must be guaranteed by control flow: `Ask` must retrieve papers first, construct bounded context from retrieval results, then call generation. Citations should not be post-hoc guessed by the model. They should be assembled from retrieved papers and returned as structured output (ID/title/link) to satisfy traceability even when model output quality varies.

For APC-03, plan a stable error taxonomy at handler boundary and keep it deterministic across retries/backends. Map input errors to `codes.InvalidArgument`, context timeouts to `codes.DeadlineExceeded`, upstream overload/rate limits to `codes.ResourceExhausted` or `codes.Unavailable`, and unknown downstream failures to `codes.Internal`.

**Primary recommendation:** Implement `Ask` as `validate -> retrieve -> build prompt context -> generate -> attach deterministic citations -> map errors`.

## Project Constraints (from CLAUDE.md)

- Preserve stack and patterns: Go + gRPC + existing client abstractions; avoid large architectural refactors.
- Keep brownfield compatibility: existing `Search` behavior and scheduler notification flow must remain stable.
- Use functional options for new clients/config (`Option func(*Config)` pattern).
- Define interfaces at point of use (small local interfaces in consuming package).
- Do not edit generated files in `internal/gen/`; regenerate via `buf generate`.
- Keep `arxiv.Paper.PublishedAt` as `time.Time` at API boundary.
- Prefer wrapped errors with context (`fmt.Errorf("...: %w", err)`).
- Keep deterministic test coverage growing, especially transport/error behavior.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `google.golang.org/grpc` | `v1.79.2` | gRPC server, status and code semantics | Already integrated; official status/codes model for stable RPC contracts |
| `buf` | `v1.50.0` (CLI installed) | Proto lint + code generation to `internal/gen` | Existing project workflow for contract/stub alignment |
| Internal `internal/client/arxiv` | in-repo | Retrieval source for grounding | Already production path for paper discovery |
| Groq Chat Completions API (`/openai/v1/chat/completions`) | API docs current as crawled | Answer generation over retrieved context | Existing config direction (`GROQ_API_KEY`) and compatible chat API |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `google.golang.org/grpc/status` | `v1.79.2` | Construct stable RPC errors | Every handler error exit path |
| `google.golang.org/grpc/codes` | `v1.79.2` | Canonical code set | Error mapping matrix enforcement |
| Go stdlib `net/http`, `encoding/json`, `context` | Go `1.26.1` | Downstream API call, payload encoding, deadlines | Groq client implementation and timeout propagation |
| `go.uber.org/zap` | `v1.27.1` | Structured logs for request tracing | Ask request lifecycle and failure diagnostics |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Direct `net/http` Groq client | Third-party Go SDK | Fewer lines, but adds dependency surface and version churn; direct HTTP is enough for one endpoint |
| Immediate vector DB introduction | Query-time arXiv retrieval (current) | Vector DB may improve semantic retrieval later, but is scope creep for this phase and adds ops burden |
| Model-only citation text | Structured citations from retrieved papers | Model-only citations are less reliable for traceability requirements |

**Installation:**
```bash
# No new mandatory dependency required for Phase 2 baseline.
# If proto changes:
buf lint
buf generate
```

**Version verification:**
- Local runtime verified: `go version go1.26.1`, `buf 1.50.0`, `protoc 3.21.12`.
- `grpc`/`codes`/`status` docs verified on pkg.go.dev at module `v1.79.2` (published 2026-03-06).

## Architecture Patterns

### Recommended Project Structure
```text
internal/
├── server/grpc/          # gRPC handlers and transport error mapping
├── client/arxiv/         # retrieval source (existing)
├── client/groq/          # new LLM client (thin HTTP wrapper)
└── rag/                  # orchestration: retrieve -> context -> generate -> citations
```

### Pattern 1: Retrieval-First Ask Pipeline
**What:** Handler must perform retrieval before generation and fail if retrieval fails/empty.
**When to use:** Every `Ask` call.
**Example:**
```go
// Source: https://grpc.io/docs/guides/status-codes/ + local architecture
papers, err := h.fetcher.FetchPapersWithQuery(ctx, req.Query, arxiv.FetchParams{MaxResults: int(limit)})
if err != nil {
	return nil, status.Errorf(codes.Unavailable, "retrieval failed: %v", err)
}
if len(papers) == 0 {
	return nil, status.Error(codes.NotFound, "no papers matched query")
}
```

### Pattern 2: Deterministic Citation Construction
**What:** Construct citation objects from retrieved papers (not from generated text parsing).
**When to use:** Always before returning `AskResponse`.
**Example:**
```go
// Source: https://info.arxiv.org/help/api/user-manual.html
citations := make([]*arxivv1.Citation, 0, len(papers))
for _, p := range papers {
	citations = append(citations, &arxivv1.Citation{
		Id:    p.ArxivID,
		Title: p.Title,
		Url:   p.PDFURL,
	})
}
```

### Pattern 3: Stable Error Mapping Boundary
**What:** Convert downstream HTTP/context errors into stable gRPC codes in one place.
**When to use:** At gRPC transport boundary.
**Example:**
```go
// Source: https://pkg.go.dev/google.golang.org/grpc/status
if errors.Is(err, context.DeadlineExceeded) {
	return nil, status.Error(codes.DeadlineExceeded, "ask deadline exceeded")
}
if isRateLimit(err) {
	return nil, status.Error(codes.ResourceExhausted, "generation rate limited")
}
return nil, status.Errorf(codes.Internal, "generation failed: %v", err)
```

### Anti-Patterns to Avoid
- **Generate before retrieval:** breaks grounding requirement (`RAG-02`).
- **Free-form citation text only:** weak traceability and brittle parsing.
- **Inconsistent code selection by call site:** violates `APC-03`; mapping must be centralized.
- **Mutating generated protobuf files manually:** creates contract drift; always regenerate with `buf`.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| gRPC error wire format | Custom error envelope in payload | `status.Error`/`status.Errorf` + `codes` | Canonical transport semantics and client interoperability |
| Proto stubs | Manual interface/stub edits in `internal/gen` | `buf generate` pipeline | Prevents contract drift and regeneration conflicts |
| arXiv metadata parsing in Ask path | Ad-hoc scraping/parsing in handler | Existing `internal/client/arxiv` client | Reuses tested retrieval/parsing logic |
| Citation provenance | Post-process LLM text with regex | Structured citations from retrieved paper structs | Deterministic traceability and simpler tests |

**Key insight:** Keep `Ask` mostly orchestration logic; reuse existing retrieval and transport primitives to minimize brownfield risk.

## Common Pitfalls

### Pitfall 1: Hallucinated answers without retrieval grounding
**What goes wrong:** LLM produces plausible text with no source linkage.
**Why it happens:** Retrieval step made optional or only best-effort.
**How to avoid:** Hard gate generation on non-empty retrieval result and include retrieval context in prompt.
**Warning signs:** `Ask` succeeds even when retrieval fails.

### Pitfall 2: Unstable gRPC status behavior
**What goes wrong:** Same failure returns different status codes across paths.
**Why it happens:** Error mapping scattered across handler/client layers.
**How to avoid:** Single mapping function in gRPC layer with explicit matrix and tests.
**Warning signs:** flaky transport tests or status changes after refactors.

### Pitfall 3: Oversized context causing downstream failures
**What goes wrong:** LLM request fails, times out, or truncates heavily.
**Why it happens:** Dumping full abstracts for too many papers.
**How to avoid:** Bound retrieved count and per-paper excerpt length before generation.
**Warning signs:** frequent 413/422/timeout responses from generation backend.

### Pitfall 4: Citation mismatch with generated claims
**What goes wrong:** Answer references paper not present in citations.
**Why it happens:** Citation extraction from model text rather than retrieval set.
**How to avoid:** Build citations from retrieval set only; optionally constrain model to cite by index.
**Warning signs:** QA finds links/IDs that were never retrieved.

## Code Examples

Verified patterns from official sources:

### gRPC status return from handler
```go
// Source: https://pkg.go.dev/google.golang.org/grpc/status
if req.Query == "" {
	return nil, status.Error(codes.InvalidArgument, "query is required")
}
```

### arXiv query and result controls
```text
GET /api/query?search_query=all:electron&start=0&max_results=10&sortBy=submittedDate&sortOrder=descending
```
Source: https://info.arxiv.org/help/api/user-manual.html

### Groq chat completion endpoint contract
```text
POST https://api.groq.com/openai/v1/chat/completions
Authorization: Bearer $GROQ_API_KEY
```
Source: https://console.groq.com/docs/api-reference

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Search-only gRPC API | Retrieval + generation Ask API with structured citations | Phase 2 scope (2026-03-31) | Enables answer generation while preserving source traceability |
| Ad-hoc downstream errors | Explicit gRPC status taxonomy | gRPC docs and APC-03 requirement | Predictable client behavior and deterministic tests |
| Free-form response without provenance | Grounded response + citation objects | RAG best practice + RAG-03 | Better trust/debuggability |

**Deprecated/outdated:**
- Manual edits in generated gRPC code under `internal/gen/` are outdated workflow; use `buf generate`.

## Open Questions

1. **Ask proto response shape finalization**
   - What we know: Needs non-empty answer + citations (RAG-01/03).
   - What's unclear: Whether response should include intermediate retrieved papers or only citations.
   - Recommendation: Keep minimal v1 shape (`answer`, `citations[]`) and defer richer debug fields.

2. **Empty retrieval semantics**
   - What we know: Grounding requirement forbids pure model-only answer.
   - What's unclear: Should zero results be `NotFound` vs empty answer with OK.
   - Recommendation: Use `codes.NotFound` to keep semantics explicit and testable.

3. **Model selection policy**
   - What we know: Groq model IDs are dynamic and externally managed.
   - What's unclear: Which model should be defaulted in env/config for latency/quality balance.
   - Recommendation: Add explicit config key for model ID and default to a documented production model at implementation time.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | Build/tests | ✓ | `go1.26.1` | — |
| `buf` CLI | Proto lint/generate after `Ask` RPC addition | ✓ | `1.50.0` | `protoc` direct generation (less aligned with project workflow) |
| `protoc` | Underlying proto compiler | ✓ | `3.21.12` | Install via package manager if missing |
| Groq API key env var | Live generation in Ask path | ✗ (unset) | — | Use mocked generation client in tests |
| External arXiv API network | Live retrieval path | Unknown in offline CI | — | Use mocked `paperFetcher` in transport tests |

**Missing dependencies with no fallback:**
- None for planning and test-first implementation.

**Missing dependencies with fallback:**
- Groq credentials for local/manual live calls; fallback is deterministic mocks for automated tests.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` (stdlib), Go `1.26.1` |
| Config file | none |
| Quick run command | `go test ./internal/server/grpc -count=1` |
| Full suite command | `go test ./... -count=1` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| RAG-01 | `Ask` returns non-empty answer for valid request | unit/transport | `go test ./internal/server/grpc -run TestAskSuccess -count=1` | ❌ Wave 0 |
| RAG-02 | `Ask` performs retrieval before generation | unit | `go test ./internal/server/grpc -run TestAskRetrievalFirst -count=1` | ❌ Wave 0 |
| RAG-03 | `Ask` returns citations with id/title/link | unit | `go test ./internal/server/grpc -run TestAskCitations -count=1` | ❌ Wave 0 |
| APC-03 | Stable status codes for invalid/downstream failures | unit/table-driven | `go test ./internal/server/grpc -run TestAskErrorMapping -count=1` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/server/grpc -count=1`
- **Per wave merge:** `go test ./... -count=1`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/server/grpc/arxiv_ask_test.go` - Ask success/grounding/citations/error mapping coverage
- [ ] `internal/server/grpc/testdata/` fixtures - deterministic paper + model outputs
- [ ] Mock generation client interface in `internal/server/grpc` or `internal/rag` for deterministic tests

## Sources

### Primary (HIGH confidence)
- Local codebase files (`proto/arxiv/v1/arxiv.proto`, `internal/server/grpc/arxiv.go`, `internal/app/app.go`, `go.mod`) - current implementation and constraints.
- gRPC Status Codes guide - https://grpc.io/docs/guides/status-codes/ (canonical code semantics).
- gRPC Go `status` package docs - https://pkg.go.dev/google.golang.org/grpc/status (handler error construction API).
- gRPC Go `codes` package docs - https://pkg.go.dev/google.golang.org/grpc/codes (canonical code set for Go).
- arXiv API User Manual - https://info.arxiv.org/help/api/user-manual.html (query/paging/sorting and Atom response structure).
- Groq API Reference - https://console.groq.com/docs/api-reference (chat completions endpoint, request parameters incl. citation options).

### Secondary (MEDIUM confidence)
- Groq API Error Codes - https://console.groq.com/docs/errors (HTTP error taxonomy used for downstream-to-gRPC mapping strategy).
- gRPC Error Handling guide - https://grpc.io/docs/guides/error/ (library-generated vs app-generated status behavior).

### Tertiary (LOW confidence)
- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - grounded in local code + official gRPC/arXiv/Groq docs.
- Architecture: MEDIUM-HIGH - retrieval-first/citation patterns are well-supported but final Ask proto shape is still a project decision.
- Pitfalls: MEDIUM - based on mixed official docs and applied RAG implementation experience.

**Research date:** 2026-03-31
**Valid until:** 2026-04-30
