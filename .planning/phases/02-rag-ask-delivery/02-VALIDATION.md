---
phase: 2
slug: rag-ask-delivery
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-31
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~45 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 90 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 2-01-01 | 01 | 0 | RAG-01 | transport/contract | `go test ./internal/server/grpc -run TestAskContract -count=1` | ❌ W0 | ⬜ pending |
| 2-01-02 | 01 | 0 | RAG-02 | integration/orchestration | `go test ./internal/server/grpc -run TestAskRetrievalFirst -count=1` | ❌ W0 | ⬜ pending |
| 2-01-03 | 01 | 0 | RAG-03 | unit/response-shape | `go test ./internal/server/grpc -run TestAskCitations -count=1` | ❌ W0 | ⬜ pending |
| 2-02-01 | 02 | 0 | APC-03 | transport/error-mapping | `go test ./internal/server/grpc -run TestAskStatusMapping -count=1` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/server/grpc/arxiv_ask_test.go` — Ask contract/grounding/citations/error mapping tests
- [ ] `internal/client/groq/client_test.go` — Groq client deterministic tests (success/timeout/rate-limit)
- [ ] `internal/rag/ask_pipeline_test.go` — retrieval-first orchestration tests

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| End-to-end live Groq response quality check | RAG-01 | Depends on live external model behavior | Run service with valid `GROQ_API_KEY`; invoke Ask with known query; confirm non-empty answer and citations from returned sources |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 90s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
