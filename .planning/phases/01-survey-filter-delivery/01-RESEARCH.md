# Phase 1: Survey Filter Delivery - Research

**Researched:** 2026-03-31  
**Domain:** Go filtering logic in arXiv → Telegram cron pipeline  
**Confidence:** MEDIUM

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

### the agent's Discretion
- None — all gray areas resolved.

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| FILT-01 | Paper is eligible only if its arXiv category is in the configured topic list (`cs.AI`, `cs.CL`) | Use `paper.Categories` (parsed in `internal/client/arxiv/client.go`) and match against `topicList` in `internal/cron/arxiv_fetcher.go`. |
| FILT-02 | Paper is eligible only if any survey keyword matches case-insensitive in title or abstract | Implement normalization + `strings.Contains` check over `paper.Title` and `paper.Abstract`. |
| FILT-03 | Survey keyword list is fixed and includes provided phrases | Define a fixed `surveyKeywords` slice (constants) in `internal/cron` package; include required phrases. |
| FILT-04 | Only eligible papers are sent to the Telegram channel | Gate `sendNotification` with eligibility check inside `FetchPapers`. |
| FILT-05 | Message format for eligible papers remains unchanged from current output | Preserve `formatPaper` and SendHTML; filtering happens before sending. |
</phase_requirements>

## Project Constraints (from CLAUDE.md)
- Go service; keep integration within existing cron pipeline.
- Use functional options for client constructors and keep interfaces defined at point of use.
- Logging via zap `SugaredLogger` (`Infow`, `Warnw`, `Errorw`).
- Use `wrappers.RateLimiter.Do(ctx, fn)` for notification sends.
- `arxiv.Paper.PublishedAt` is `time.Time` parsed at the API boundary.
- Proto stubs live in `internal/gen/`; use `buf generate` when `.proto` changes.

## Summary

Filtering should be added inside the existing cron fetch loop (`internal/cron/arxiv_fetcher.go`) before any Telegram send. The paper metadata already contains a normalized title and abstract, plus a list of all categories, so eligibility can be determined without changing upstream parsing or the notification formatter.

Use a small, pure eligibility function that checks category membership against `topicList` and performs case-insensitive substring matching over title/abstract with a light normalization layer to allow spacing and hyphen variants (e.g., normalize hyphens to spaces and collapse whitespace). Title-only matches are allowed if the abstract is empty. Keep `formatPaper` unchanged to satisfy formatting constraints.

**Primary recommendation:** Implement `isEligibleSurvey(paper arxiv.Paper, topics []string, keywords []string) bool` in `internal/cron` and gate `sendNotification` on it, using normalized case-insensitive matching with hyphen/space normalization.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib (`strings`, optional `regexp`) | Go 1.26 | Case-insensitive matching and normalization | Already used throughout codebase; no new dependencies. |
| `go.uber.org/zap` | v1.27.1 | Logging in cron fetcher | Existing logging pattern in `internal/cron/arxiv_fetcher.go`. |
| `internal/client/arxiv` | local | Paper metadata source | Provides `Paper.Title`, `Paper.Abstract`, `Paper.Categories`. |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `internal/wrappers` | local | Rate-limited Telegram sends | Keep existing `RateLimiter.Do` usage. |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `strings.Contains` on normalized strings | `regexp` per keyword | Regex adds complexity; normalization + substring is sufficient and faster. |

**Installation:** No new dependencies required.

## Architecture Patterns

### Recommended Project Structure
```
internal/
  cron/
    arxiv_fetcher.go    # add eligibility gate
    survey_filter.go    # (optional) isolated filter helpers
```

### Pattern 1: Pure Eligibility Helper
**What:** A pure function that takes `Paper` and config slices, returns `true/false`.  
**When to use:** For all eligibility checks so filtering is testable without Telegram or network dependencies.  
**Example:**
```go
// Source: internal/cron/arxiv_fetcher.go + internal/client/arxiv/paper.go
func isEligibleSurvey(p arxiv.Paper, topics, keywords []string) bool {
	if !hasAnyCategory(p.Categories, topics) {
		return false
	}
	text := normalizeForMatch(p.Title)
	if p.Abstract != "" {
		text += " " + normalizeForMatch(p.Abstract)
	}
	for _, kw := range keywords {
		if strings.Contains(text, normalizeForMatch(kw)) {
			return true
		}
	}
	return false
}
```

### Anti-Patterns to Avoid
- **Filtering after sending:** Must gate before `SendHTML` to satisfy FILT-04.
- **Primary-category-only checks:** Use `paper.Categories` (all categories) per D-04.
- **Changing formatter:** `formatPaper` must remain unchanged per FILT-05.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Complex NLP or stemming | Custom tokenizer/stemmer | Fixed keyword list + normalization | Requirement is simple substring match, not semantic search. |

**Key insight:** Requirements explicitly call for fixed, case-insensitive substring matching; anything more complex risks scope creep and regressions.

## Common Pitfalls

### Pitfall 1: Missing Hyphen/Spacing Variants
**What goes wrong:** `"state of the art"` fails to match `"state-of-the-art"`.  
**Why it happens:** Direct `strings.Contains` on raw text.  
**How to avoid:** Normalize by lowercasing, replacing hyphens with spaces, and collapsing whitespace before matching.  
**Warning signs:** Eligible-looking papers not delivered when hyphenated phrases exist.

### Pitfall 2: Category Mismatch
**What goes wrong:** Papers with secondary category `cs.AI` are dropped.  
**Why it happens:** Using only the fetch topic or primary category instead of full list.  
**How to avoid:** Check `paper.Categories` for any match against `topicList`.  
**Warning signs:** Papers fetched for a topic but filtered out without keyword issues.

### Pitfall 3: Abstract Empty Handling
**What goes wrong:** Papers with empty abstract are always excluded.  
**Why it happens:** Assuming abstract always present.  
**How to avoid:** Allow title-only match if `Abstract` is empty.  
**Warning signs:** Empty-abstract papers never delivered even with survey titles.

## Code Examples

### Normalization Helper
```go
// Source: internal/client/arxiv/client.go (cleanText pattern) + internal/cron/arxiv_fetcher.go
func normalizeForMatch(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "-", " ")
	return strings.Join(strings.Fields(s), " ")
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| No filtering | Survey + category filtering | Phase 1 | Telegram receives only survey/review papers in configured categories. |

**Deprecated/outdated:**
- None for this phase.

## Open Questions

1. **Exact keyword list for FILT-03**
   - What we know: Must include provided phrases (e.g., "survey", "review", "state of the art", "taxonomy").
   - What's unclear: Full canonical list and any variants beyond those examples.
   - Recommendation: Confirm the full list before implementation to avoid rework.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | Build/tests | ✓ | go1.26.1 | — |

**Missing dependencies with no fallback:**
- None.

**Missing dependencies with fallback:**
- None.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` (stdlib) |
| Config file | none |
| Quick run command | `go test ./...` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| FILT-01 | Categories in configured list only | unit | `go test ./internal/cron -run TestSurveyFilterCategories -v` | ❌ Wave 0 |
| FILT-02 | Case-insensitive keyword match in title/abstract | unit | `go test ./internal/cron -run TestSurveyFilterKeywords -v` | ❌ Wave 0 |
| FILT-03 | Fixed keyword list includes required phrases | unit | `go test ./internal/cron -run TestSurveyKeywordList -v` | ❌ Wave 0 |
| FILT-04 | Only eligible papers sent | unit/integration | `go test ./internal/cron -run TestFetchPapersFiltersBeforeSend -v` | ❌ Wave 0 |
| FILT-05 | Format unchanged | unit | `go test ./internal/cron -run TestFormatPaperUnchanged -v` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/cron -run TestSurveyFilter.* -v`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/cron/survey_filter_test.go` — eligibility + keyword list coverage
- [ ] `internal/cron/arxiv_fetcher_test.go` — ensure filter gates send

## Sources

### Primary (HIGH confidence)
- `internal/cron/arxiv_fetcher.go` — current fetch loop, topic list, formatting
- `internal/client/arxiv/client.go` — category parsing and text normalization
- `internal/client/arxiv/paper.go` — paper fields used by filter
- `.planning/phases/01-survey-filter-delivery/01-CONTEXT.md` — locked decisions
- `.planning/REQUIREMENTS.md` — FILT-01..FILT-05 definitions
- `CLAUDE.md` — project conventions and constraints

### Secondary (MEDIUM confidence)
- None.

### Tertiary (LOW confidence)
- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — derived from repo stack and go.mod.
- Architecture: MEDIUM — based on existing cron structure; no new patterns required.
- Pitfalls: MEDIUM — inferred from requirements and current code path.

**Research date:** 2026-03-31  
**Valid until:** 2026-04-30
