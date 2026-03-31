# Requirements: go-rag-arxiv

**Defined:** 2026-03-31
**Core Value:** Deliver relevant arXiv paper discovery and notifications reliably, with a clear path to answer-generation over retrieved papers.

## v1 Requirements

### RAG Answering

- [ ] **RAG-01**: User can call `Ask` over gRPC with a natural-language question and receive an answer payload.
- [ ] **RAG-02**: `Ask` retrieves relevant arXiv papers before generating the response.
- [ ] **RAG-03**: `Ask` returns citations (paper identifiers/titles/links) for answer traceability.

### API Contract & Config

- [ ] **APC-01**: gRPC proto contract and server implementation remain aligned (no declared-but-unimplemented public RPCs).
- [ ] **APC-02**: Runtime env configuration matches actual usage (no required secrets that are unused).
- [ ] **APC-03**: gRPC errors are mapped to clear, stable status codes for validation and downstream failures.

### Reliability & Testing

- [ ] **REL-01**: Transport-layer behavior is covered by deterministic tests (`Search`, `Ask`, validation, and error mapping).
- [ ] **REL-02**: Scheduler/notification workflow is covered by deterministic tests for execution and failure handling.
- [ ] **REL-03**: External client behavior (arXiv/Telegram) has deterministic tests for retry/non-200/error paths.

### Security & Operations

- [ ] **OPS-01**: gRPC traffic can be served with TLS in production-ready configuration.
- [ ] **OPS-02**: Service access controls are defined for exposed endpoints (bind policy and/or auth middleware/interceptor).
- [ ] **OPS-03**: Notification formatting safely escapes dynamic content before delivery.

## v2 Requirements

### Product Expansion

- **EXP-01**: Add user-facing filtering/preferences for scheduled topics and alert formats.
- **EXP-02**: Add distributed coordination for scheduler execution in multi-instance deployments.
- **EXP-03**: Move PDF cache/index from local disk to shared storage for horizontal scaling.

## Out of Scope

| Feature | Reason |
|---------|--------|
| Web frontend/dashboard | Current milestone focuses on backend correctness and API capabilities |
| Multi-tenant accounts and RBAC | Not required for current single-tenant Telegram workflow |
| Mobile clients | Not needed before backend API and RAG path are stable |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| RAG-01 | Phase TBD | Pending |
| RAG-02 | Phase TBD | Pending |
| RAG-03 | Phase TBD | Pending |
| APC-01 | Phase TBD | Pending |
| APC-02 | Phase TBD | Pending |
| APC-03 | Phase TBD | Pending |
| REL-01 | Phase TBD | Pending |
| REL-02 | Phase TBD | Pending |
| REL-03 | Phase TBD | Pending |
| OPS-01 | Phase TBD | Pending |
| OPS-02 | Phase TBD | Pending |
| OPS-03 | Phase TBD | Pending |

**Coverage:**
- v1 requirements: 12 total
- Mapped to phases: 0
- Unmapped: 12 ⚠️

---
*Requirements defined: 2026-03-31*
*Last updated: 2026-03-31 after initial definition*
