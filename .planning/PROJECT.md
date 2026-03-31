# Arxiv Survey Filter

## What This Is

A filter layer on the existing arXiv fetch → Telegram pipeline that only forwards survey/review-style papers in the specified arXiv categories. It is designed for internal use to reduce noise by sending only higher-level survey content about AI (and related topics as defined by the category list).

## Core Value

Only survey/review articles from the chosen arXiv categories reach the Telegram channel.

## Requirements

### Validated

- ✓ Scheduled arXiv fetching by category list exists and runs daily — existing
- ✓ Telegram notifications are sent from the fetch pipeline — existing
- ✓ Topic/category list is configurable in code — existing
- ✓ Filter papers by a fixed `SURVEY_KEYWORDS` list, matching case-insensitive in both title and abstract — Validated in Phase 01: Survey Filter Delivery
- ✓ Only include papers whose arXiv category is in the configured topic list (currently `cs.AI`, `cs.CL`) — Validated in Phase 01: Survey Filter Delivery
- ✓ Send to Telegram only the papers that pass the survey keyword filter — Validated in Phase 01: Survey Filter Delivery
- ✓ Keep existing fetch schedule and output formatting unchanged for matching papers — Validated in Phase 01: Survey Filter Delivery

### Active

(None)

### Out of Scope

- Cryptography category expansion (e.g., `cs.CR`) — not requested for v1
- Non-survey papers in the same categories — explicitly excluded
- New sources beyond arXiv — arXiv API only
- UI/dashboard for managing filters — not required for v1

## Context

There is an existing Go service that fetches arXiv papers on a cron schedule and sends messages to Telegram. The fetching logic lives in `func (f *ArxivFetcher) FetchPapers(ctx context.Context)` and already filters by a topic/category list. This project adds a survey-only filter using a predefined keyword list and applies it to both title and abstract before sending to Telegram.

## Constraints

- **Source**: arXiv API — existing integration must remain the input source
- **Runtime**: Go service with existing cron job — must integrate with current pipeline
- **Filter Logic**: Keyword list is fixed for v1; match is case-insensitive over title + abstract
- **Categories**: Only those in the current topic list (`cs.AI`, `cs.CL`)

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Use a fixed `SURVEY_KEYWORDS` list | Simple, predictable filter for survey/review content | Implemented in Phase 01 |
| Match keywords against both title and abstract | Avoid missing surveys that omit keywords in title | Implemented in Phase 01 |
| Filter only within configured arXiv categories | Preserve existing category scope | Implemented in Phase 01 |

## Current State

Phase 01 complete — survey-only eligibility gate and regression coverage added to the cron fetch pipeline.

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `$gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `$gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-03-31 after Phase 01 completion*
