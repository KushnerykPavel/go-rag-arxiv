# Roadmap: Arxiv Survey Filter

## Overview

Deliver a single, reliable filter layer that only forwards survey/review papers from the configured arXiv categories to Telegram, while preserving existing schedule and message formatting.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Survey Filter Delivery** - Survey-only filtering in the existing arXiv → Telegram pipeline

## Phase Details

### Phase 1: Survey Filter Delivery
**Goal**: Only survey/review papers from the configured categories are delivered to Telegram with unchanged formatting
**Depends on**: Nothing (first phase)
**Requirements**: FILT-01, FILT-02, FILT-03, FILT-04, FILT-05
**Success Criteria** (what must be TRUE):
  1. Only papers whose arXiv category is in the configured list (`cs.AI`, `cs.CL`) are eligible.
  2. A paper is eligible only if a survey keyword matches case-insensitive in title or abstract.
  3. The survey keyword list is fixed and includes the provided phrases (e.g., "survey", "review", "state of the art", "taxonomy").
  4. Telegram receives only eligible papers.
  5. Message formatting for eligible papers is unchanged from current output.
**Plans**: 1 plans

Plans:
- [ ] 01-survey-filter-delivery-01-PLAN.md — Add survey eligibility filter with fixed keywords and tests in cron fetch flow

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 1.1 → 1.2 → 2

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Survey Filter Delivery | 0/TBD | Not started | - |
