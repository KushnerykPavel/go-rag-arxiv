# Phase 1: Survey Filter Delivery - Context

**Gathered:** 2026-03-31
**Status:** Ready for planning

<domain>
## Phase Boundary

Add survey-only filtering to the existing arXiv → Telegram pipeline; no new capabilities beyond filtering behavior.

</domain>

<decisions>
## Implementation Decisions

### Keyword Matching
- **D-01:** Use case-insensitive substring matching across title and abstract.
- **D-02:** Title-only match is allowed when abstract is missing/empty.
- **D-03:** Multi-word phrases allow flexible spacing/hyphen variants (e.g., "state of the art" matches "state-of-the-art").

### Category Matching
- **D-04:** A paper is eligible if **any** category matches the configured topic list (not just the primary category).

### the agent's Discretion
- None — all gray areas resolved.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase Definition
- `.planning/ROADMAP.md` — Phase 1 goal, requirements, success criteria
- `.planning/REQUIREMENTS.md` — FILT-01..FILT-05 requirement details
- `.planning/PROJECT.md` — constraints and context for the pipeline

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/cron/arxiv_fetcher.go`: Current fetch loop + Telegram send flow; primary integration point for filtering.
- `internal/client/arxiv/client.go`: arXiv API client and paper model parsing.
- `internal/client/arxiv/paper.go`: Paper fields (title, abstract, categories) used for matching.

### Established Patterns
- Logging via zap SugaredLogger (`Infow`, `Warnw`, `Errorw`) inside cron fetcher.
- Error handling: log-and-continue for per-paper failures in fetch loop.

### Integration Points
- Apply filter within `func (f *ArxivFetcher) FetchPapers(ctx context.Context)` before Telegram send.

</code_context>

<specifics>
## Specific Ideas

No specific UI or output changes; formatting must remain unchanged.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---
*Phase: 01-survey-filter-delivery*
*Context gathered: 2026-03-31*
