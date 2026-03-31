---
phase: 02-rag-ask-delivery
plan: 01
subsystem: api
tags: [grpc, proto, buf, contract, rag]
requires:
  - phase: 01-contract-baseline
    provides: "Search-only baseline and contract drift tests used as foundation for Ask reintroduction"
provides:
  - "Ask RPC declared in ArxivService transport contract"
  - "AskRequest/AskResponse/Citation protobuf messages for answer plus structured citations"
  - "Contract tests guarding Ask presence and response field shape"
affects: [phase-2-rag-ask-delivery, grpc, proto]
tech-stack:
  added: []
  patterns: ["contract-first rpc evolution", "descriptor-based transport contract assertions"]
key-files:
  created: [internal/server/grpc/arxiv_ask_test.go]
  modified: [internal/server/grpc/arxiv_contract_test.go, proto/arxiv/v1/arxiv.proto, internal/gen/arxiv/v1/arxiv.pb.go, internal/gen/arxiv/v1/arxiv_grpc.pb.go]
key-decisions:
  - "Used descriptor-level assertions so Ask contract tests fail before proto regeneration and pass after generated code updates."
patterns-established:
  - "Any new RPC surface is introduced with failing transport tests before proto/stub regeneration."
requirements-completed: [RAG-01]
duration: 11min
completed: 2026-03-31
---

# Phase 2 Plan 1: Ask Contract Reintroduction Summary

**Reintroduced Ask as a generated gRPC transport surface with explicit answer and citations schema guarded by contract tests**

## Performance

- **Duration:** 11 min
- **Started:** 2026-03-31T16:00:50Z
- **Completed:** 2026-03-31T16:11:24Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- Added failing Ask transport contract tests that asserted service method presence and AskResponse field shape.
- Extended proto contract with `Ask`, `AskRequest`, `AskResponse`, and `Citation`.
- Regenerated protobuf and gRPC stubs with `buf generate` and validated targeted and package grpc tests.

## Task Commits

1. **Task 1: Add failing Ask transport contract tests (RAG-01)** - `76f3654` (test)
2. **Task 2: Add Ask proto/messages and regenerate stubs (RAG-01)** - `100cc6d` (feat)

## Files Created/Modified
- `internal/server/grpc/arxiv_ask_test.go` - Ask descriptor/field contract tests.
- `internal/server/grpc/arxiv_contract_test.go` - Contract allowlist updated for `Ask`.
- `proto/arxiv/v1/arxiv.proto` - Ask RPC plus request/response/citation message definitions.
- `internal/gen/arxiv/v1/arxiv.pb.go` - Regenerated protobuf message bindings.
- `internal/gen/arxiv/v1/arxiv_grpc.pb.go` - Regenerated gRPC service/client bindings including Ask.

## Decisions Made
- Kept Ask contract verification focused on generated descriptor data to avoid runtime implementation coupling at this stage.

## Deviations from Plan
None - plan executed exactly as written.

## Issues Encountered
- Go tests required elevated execution due sandbox restrictions on the Go build cache path.
- Parallel git operations briefly contended on `.git/index.lock`; resolved by retrying commits serially.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Transport contract for `Ask` is stable and test-guarded.
- Runtime Ask implementation can proceed against generated server/client interfaces.

## Self-Check: PASSED
- Found summary artifact: `.planning/phases/02-rag-ask-delivery/02-rag-ask-delivery-01-SUMMARY.md`
- Found task commit: `76f3654`
- Found task commit: `100cc6d`

---
*Phase: 02-rag-ask-delivery*
*Completed: 2026-03-31*
