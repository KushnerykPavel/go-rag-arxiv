# Phase 1: Survey Filter Delivery - Research

**Researched:** 2026-03-31
**Domain:** Go cron pipeline filtering (arXiv → Telegram)
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
## Implementation Decisions

### Keyword Matching
- **D-01:** Use case-insensitive substring matching across title and abstract.
- **D-02:** Title-only match is allowed when abstract is missing/empty.
- **D-03:** Multi-word phrases allow flexible spacing/hyphen variants (e.g., "state of the art" matches "state-of-the-art").

### Category Matching
- **D-04:** A paper is eligible if **any** category matches the configured topic list (not just the primary category).

### Claude's Discretion
### the agent's Discretion
- None — all gray areas resolved.

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| FILT-01 | Paper is eligible only if its arXiv category is in the configured topic list (`cs.AI`, `cs.CL`) | Use `topicList` and match against `paper.Categories` before sending. |
| FILT-02 | Paper is eligible only if any survey keyword matches case-insensitive in title or abstract | Add eligibility function using case-insensitive substring match across title + abstract. |
| FILT-03 | Survey keyword list is fixed and includes the provided phrases (e.g., "survey", "review", "state of the art", "taxonomy") | Create fixed keyword list constant and normalize spacing/hyphen variants. |
| FILT-04 | Only eligible papers are sent to the Telegram channel | Apply eligibility check before `sendNotification` in `FetchPapers`. |
| FILT-05 | Message format for eligible papers remains unchanged from current output | Preserve `formatPaper` and `sendNotification` flow unchanged; only filter before calling it. |
</phase_requirements>

## Summary

The current cron fetcher (`internal/cron/arxiv_fetcher.go`) iterates `topicList`, fetches papers per topic, and sends each paper to Telegram using `formatPaper(...)` with no filtering. Eligibility should be evaluated in `FetchPapers` before calling `sendNotification`, ensuring only eligible papers are sent and formatting remains untouched.

Eligibility is a conjunction of category match and keyword match. Category match must check **any** value in `paper.Categories` against the configured topic list (not just the primary category). Keyword match must be case-insensitive across title + abstract, allow title-only when abstract is empty, and support flexible spacing/hyphen variants for multi-word phrases like "state of the art". Existing parsing already provides `Paper.Abstract` and `Paper.Categories`, so no client changes are required.

**Primary recommendation:** Add a small, testable `isEligiblePaper(paper arxiv.Paper) bool` helper in the cron fetcher package and apply it in the inner loop before `sendNotification`.

## Project Constraints (from CLAUDE.md)

- Use Go 1.26 (project baseline in `go.mod`).
- Keep integration in existing cron pipeline (`internal/cron/arxiv_fetcher.go`).
- Preserve existing message formatting (`formatPaper`) and notification flow.
- Follow existing conventions: functional options, interfaces defined at point of use, zap SugaredLogger, log-and-continue on per-item failures.
- Environment config via `envconfig` with prefix `producer`.
- Do not edit generated protobuf stubs (`internal/gen/`); regenerate with `buf generate` if proto changes.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | 1.26.1 | Service runtime | Baseline toolchain (`go.mod`, local `go version`). |
| `strings` (stdlib) | — | Case-insensitive matching and normalization | Sufficient for required rules. |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `go.uber.org/zap` | v1.27.1 | Logging | Use existing `SugaredLogger` for eligibility diagnostics if needed. |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `strings` matching | `regexp` | Unnecessary complexity; substring + normalization meets requirements. |

**Installation:** No new dependencies required.

## Architecture Patterns

### Recommended Project Structure
```
internal/
  cron/
    arxiv_fetcher.go   # eligibility check inserted here
```

### Pattern 1: Filter Before Notification
**What:** Compute eligibility and skip send for non-eligible papers.
**When to use:** Any gating logic where output formatting must remain unchanged.
**Example:**
```go
// Source: internal/cron/arxiv_fetcher.go
for _, paper := range papers {
	if !isEligiblePaper(paper) {
		continue
	}
	f.sendNotification(ctx, topic, paper)
}
```

### Pattern 2: Normalize Text for Flexible Phrase Matching
**What:** Lowercase + replace hyphens with spaces + collapse whitespace before substring checks.
**When to use:** Required for multi-word phrases with flexible spacing/hyphen variants (D-03).
**Example:**
```go
// Source: new helper in internal/cron/arxiv_fetcher.go
func normalizeText(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "-", " ")
	return strings.Join(strings.Fields(s), " ")
}
```

### Anti-Patterns to Avoid
- **Filtering by `topic` only:** The fetch topic is not sufficient for D-04; use `paper.Categories`.
- **Modifying `formatPaper`:** Violates FILT-05; keep output formatting unchanged.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Text normalization | Custom tokenizer | `strings.ToLower`, `strings.ReplaceAll`, `strings.Fields` | Simple, predictable behavior that matches requirements. |

**Key insight:** The eligibility rules are narrow; avoid complex NLP or regex to keep behavior deterministic.

## Common Pitfalls

### Pitfall 1: Missing Abstract Handling
**What goes wrong:** Keyword matching fails when abstract is empty.
**Why it happens:** Abstract may be empty or missing; D-02 requires title-only match in that case.
**How to avoid:** Treat empty abstract as title-only input.
**Warning signs:** Eligible survey titles not delivered.

### Pitfall 2: Hyphen/Whitespace Variants Not Matched
**What goes wrong:** "state-of-the-art" doesn't match "state of the art".
**Why it happens:** Direct substring match on raw text.
**How to avoid:** Normalize hyphens to spaces and collapse whitespace for both text and keywords.
**Warning signs:** Known survey phrases missing in notifications.

### Pitfall 3: Category Matching Too Narrow
**What goes wrong:** Only primary category checked; cross-listed papers dropped.
**Why it happens:** Assuming fetch topic equals eligibility.
**How to avoid:** Check `paper.Categories` for any match with `topicList`.
**Warning signs:** Papers missing despite having target category in secondary list.

## Code Examples

Verified patterns from existing codebase:

### Send HTML Notification
```go
// Source: internal/cron/arxiv_fetcher.go
func (f *ArxivFetcher) sendNotification(ctx context.Context, topic string, paper arxiv.Paper) {
	err := f.limiter.Do(ctx, func(ctx context.Context) error {
		return f.notifier.SendHTML(ctx, f.chatID, formatPaper(topic, paper))
	})
	if err != nil {
		f.l.Errorw("failed to send paper notification",
			"paper_id", paper.ArxivID,
			zap.Error(err),
		)
	}
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Send all papers per topic | Filter by category + survey keywords before send | Phase 1 | Reduces noise; keeps format unchanged. |

**Deprecated/outdated:**
- None — filtering is new behavior for this phase.

## Open Questions

1. **Should skipped papers be logged (counts only) or stay silent?**
   - What we know: Current fetcher logs errors only, not per-paper decisions.
   - What's unclear: Desired verbosity for eligibility filtering.
   - Recommendation: Default to minimal logging (optional debug-level counts) to avoid log noise.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | Build/test | ✓ | 1.26.1 | — |

**Missing dependencies with no fallback:**
- None.

**Missing dependencies with fallback:**
- None.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` (stdlib) |
| Config file | none — Go defaults |
| Quick run command | `go test ./internal/cron -run TestArxivFetcherEligibility -count=1` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|--------------|
| FILT-01 | Category list intersection required | unit | `go test ./internal/cron -run TestEligibilityCategory -count=1` | ❌ Wave 0 |
| FILT-02 | Keyword match in title/abstract, case-insensitive | unit | `go test ./internal/cron -run TestEligibilityKeywords -count=1` | ❌ Wave 0 |
| FILT-03 | Fixed keyword list includes phrases + spacing/hyphen variants | unit | `go test ./internal/cron -run TestEligibilityKeywordVariants -count=1` | ❌ Wave 0 |
| FILT-04 | Only eligible papers sent | unit | `go test ./internal/cron -run TestFetchSkipsIneligible -count=1` | ❌ Wave 0 |
| FILT-05 | Format unchanged | unit | `go test ./internal/cron -run TestFormatPaperUnchanged -count=1` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/cron -run TestArxivFetcherEligibility -count=1`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/cron/arxiv_fetcher_test.go` — eligibility + fetch behavior tests

## Sources

### Primary (HIGH confidence)
- `internal/cron/arxiv_fetcher.go` — fetch loop, topic list, send/format flow
- `internal/client/arxiv/paper.go` — paper fields used for matching
- `internal/client/arxiv/client.go` — category + abstract parsing
- `go.mod` — toolchain and dependency versions
- `CLAUDE.md` — project constraints and conventions

### Secondary (MEDIUM confidence)
- None.

### Tertiary (LOW confidence)
- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - versions pinned in `go.mod`, toolchain verified locally.
- Architecture: HIGH - pipeline structure is clear in `internal/cron/arxiv_fetcher.go`.
- Pitfalls: MEDIUM - based on known edge cases for string/category matching.

**Research date:** 2026-03-31
**Valid until:** 2026-04-30
