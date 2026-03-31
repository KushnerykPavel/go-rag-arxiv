---
phase: 1
slug: contract-baseline
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-31
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 1-01-01 | 01 | 0 | APC-01 | integration/transport | `go test ./internal/server/grpc -run TestArxivServiceContract -count=1` | ❌ W0 | ⬜ pending |
| 1-01-02 | 01 | 0 | APC-02 | unit | `go test ./internal/app -run TestConfigValidation -count=1` | ❌ W0 | ⬜ pending |
| 1-01-03 | 01 | 0 | APC-02 | startup/integration | `go test ./internal/app -run TestStartupValidation -count=1` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/server/grpc/arxiv_contract_test.go` — contract alignment test stubs for APC-01
- [ ] `internal/app/config_test.go` — config required/optional behavior tests for APC-02
- [ ] `internal/app/startup_validation_test.go` — startup validation message tests for APC-02

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| None | None | None | All phase behaviors should be automated |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
