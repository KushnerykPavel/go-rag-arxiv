---
phase: 01
slug: survey-filter-delivery
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-31
---

# Phase 01 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `go test ./internal/cron -run TestSurvey -v` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/cron -run TestSurvey -v`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 01-01-01 | 01 | 1 | FILT-01 | unit | `go test ./internal/cron -run TestSurveyFilterCategories -v` | ✅ | ⬜ pending |
| 01-01-01 | 01 | 1 | FILT-02 | unit | `go test ./internal/cron -run TestSurveyFilterKeywords -v` | ✅ | ⬜ pending |
| 01-01-01 | 01 | 1 | FILT-03 | unit | `go test ./internal/cron -run TestSurveyKeywordList -v` | ✅ | ⬜ pending |
| 01-01-01 | 01 | 1 | FILT-04 | integration | `go test ./internal/cron -run TestFetchPapersFiltersBeforeSend -v` | ✅ | ⬜ pending |
| 01-01-01 | 01 | 1 | FILT-05 | unit | `go test ./internal/cron -run TestFormatPaperUnchanged -v` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

All phase behaviors have automated verification.

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
