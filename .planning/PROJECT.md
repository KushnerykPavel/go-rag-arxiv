# go-rag-arxiv

## What This Is

go-rag-arxiv is a Go service that fetches recent arXiv papers on a schedule, exposes paper search over gRPC, and sends paper notifications to Telegram. It is aimed at users who want a lightweight backend for research paper discovery and alerting. The codebase is currently a brownfield baseline with retrieval/search implemented and answer-generation (RAG `Ask`) still pending.

## Core Value

Deliver relevant arXiv paper discovery and notifications reliably, with a clear path to answer-generation over retrieved papers.

## Requirements

### Validated

- ✓ Daily paper fetch workflow runs via scheduler and topic queries — existing in `internal/app/app.go` and `internal/cron/arxiv_fetcher.go`
- ✓ Telegram paper notifications are sent from fetched results — existing in `internal/client/telegram/client.go`
- ✓ gRPC paper search API is implemented (`Search`) — existing in `internal/server/grpc/arxiv.go`
- ✓ Health check endpoint is exposed over HTTP (`/health`) — existing in `internal/app/app.go`
- ✓ Local PDF caching/downloading is implemented in arXiv client — existing in `internal/client/arxiv/client.go`
- ✓ Public gRPC contract is aligned to implemented runtime behavior (no declared unimplemented RPCs) — validated in Phase 1 (`APC-01`)
- ✓ Runtime config validation now requires only actively used keys with explicit errors — validated in Phase 1 (`APC-02`)

### Active

- [ ] Implement `Ask` RPC end-to-end using retrieval + LLM answer generation
- [ ] Harden transport/security posture for gRPC and service exposure
- [ ] Add deterministic tests for transport, scheduler, and client resilience paths

### Out of Scope

- Public frontend/web UI — current project scope is backend service only
- Multi-tenant user/account management — not required for current Telegram-centric workflow

## Context

- Runtime stack is Go 1.26 with gRPC, chi, gocron, envconfig, and zap.
- The architecture is a layered modular monolith with a composition-root `App.Run` that wires clients, scheduler, and servers.
- Brownfield mapping completed on 2026-03-31 under `.planning/codebase/` and highlights key gaps: unimplemented `Ask` RPC, config/contract drift, weak security defaults, and limited test coverage.
- The repository already uses git and protobuf generation (`buf`) with generated stubs in `internal/gen/arxiv/v1/`.

## Constraints

- **Tech stack**: Go + gRPC + existing client abstractions — preserve current patterns to reduce refactor risk.
- **Brownfield compatibility**: Existing `Search` API and scheduled notification behavior must remain stable while adding new features.
- **Operational simplicity**: Service should remain straightforward to run in containerized environments with env-based configuration.
- **Quality**: New work should increase deterministic test coverage, especially around transport and scheduled workflows.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Initialize as brownfield using inferred validated requirements | Existing code already delivers core retrieval/notification capabilities | ✓ Good |
| Prioritize `Ask` pipeline as first major active capability | gRPC contract and env config already signal this direction; currently missing in runtime | — Pending |
| Keep architecture modular monolith for now | Current size/scope does not justify distributed split; focus on correctness and reliability first | ✓ Good |

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
*Last updated: 2026-03-31 after Phase 1 completion*
