# Requirements: Arxiv Survey Filter

**Defined:** 2026-03-31
**Core Value:** Only survey/review articles from the chosen arXiv categories reach the Telegram channel.

## v1 Requirements

### Filtering

- [ ] **FILT-01**: Paper is eligible only if its arXiv category is in the configured topic list (`cs.AI`, `cs.CL`)
- [ ] **FILT-02**: Paper is eligible only if any survey keyword matches case-insensitive in title or abstract
- [ ] **FILT-03**: Survey keyword list is fixed and includes the provided phrases (e.g., "survey", "review", "state of the art", "taxonomy")
- [ ] **FILT-04**: Only eligible papers are sent to the Telegram channel
- [ ] **FILT-05**: Message format for eligible papers remains unchanged from current output

## v2 Requirements

(None yet — no deferred scope defined)

## Out of Scope

| Feature | Reason |
|---------|--------|
| Cryptography category expansion (e.g., `cs.CR`) | Not requested for v1 |
| Non-survey papers in target categories | Explicitly excluded |
| Additional sources beyond arXiv | arXiv API only |
| UI/dashboard for managing filters | Not required for v1 |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| FILT-01 | Phase 1 | Pending |
| FILT-02 | Phase 1 | Pending |
| FILT-03 | Phase 1 | Pending |
| FILT-04 | Phase 1 | Pending |
| FILT-05 | Phase 1 | Pending |

**Coverage:**
- v1 requirements: 5 total
- Mapped to phases: 5
- Unmapped: 0 ✓

---
*Requirements defined: 2026-03-31*
*Last updated: 2026-03-31 after initial definition*
