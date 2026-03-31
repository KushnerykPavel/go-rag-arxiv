# Roadmap: go-rag-arxiv

## Overview

This roadmap delivers the missing `Ask` capability, aligns API/config contracts with runtime behavior, hardens service security posture, and adds deterministic reliability tests across transport, scheduler, and external clients while preserving existing `Search` and notification behavior.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Contract Baseline** - Align public RPC/config contracts with implemented runtime behavior.
- [ ] **Phase 2: RAG Ask Delivery** - Deliver end-to-end `Ask` answering with retrieval, citations, and stable error mapping.
- [ ] **Phase 3: Security Hardening** - Enable secure production transport and safer endpoint/message handling.
- [ ] **Phase 4: Transport Determinism** - Add deterministic transport tests for gRPC behavior and validation paths.
- [ ] **Phase 5: Workflow Resilience Tests** - Add deterministic tests for scheduler and external client failure handling.

## Phase Details

### Phase 1: Contract Baseline
**Goal**: Service contracts match real runtime behavior so callers and operators see no drift between declared and implemented capabilities.
**Depends on**: Nothing (first phase)
**Requirements**: APC-01, APC-02
**Success Criteria** (what must be TRUE):
  1. Operator can start the service without providing unused required secrets.
  2. gRPC API surface exposed by the server matches the proto contract without declared-but-unimplemented public RPCs.
  3. Configuration validation errors clearly identify missing values that are actually required at runtime.
**Plans**: 2 plans
Plans:
- [ ] 01-contract-baseline-01-PLAN.md — Align proto and generated gRPC contract to implemented runtime RPC surface (APC-01).
- [ ] 01-contract-baseline-02-PLAN.md — Align runtime config requirements and startup validation with actual usage (APC-02).

### Phase 2: RAG Ask Delivery
**Goal**: Users can ask natural-language questions over gRPC and receive grounded answers with paper citations.
**Depends on**: Phase 1
**Requirements**: RAG-01, RAG-02, RAG-03, APC-03
**Success Criteria** (what must be TRUE):
  1. User can call `Ask` over gRPC and receive a non-empty answer payload for valid requests.
  2. `Ask` responses are grounded in retrieved arXiv papers rather than returning an answer without retrieval context.
  3. User receives citations in `Ask` output including paper identifiers/titles/links.
  4. User receives stable gRPC status codes for invalid input and downstream retrieval/generation failures.
**Plans**: TBD

### Phase 3: Security Hardening
**Goal**: Service can be run in a production-ready secure mode for transport and endpoint/message handling.
**Depends on**: Phase 2
**Requirements**: OPS-01, OPS-02, OPS-03
**Success Criteria** (what must be TRUE):
  1. Operator can run gRPC with TLS enabled and clients can connect using that secure transport.
  2. Exposed endpoints enforce defined access controls (bind policy and/or auth guard) rather than unrestricted access by default.
  3. Telegram notifications render dynamic paper content safely without formatting injection side effects.
**Plans**: TBD

### Phase 4: Transport Determinism
**Goal**: gRPC transport behavior is verifiable through deterministic automated tests.
**Depends on**: Phase 2
**Requirements**: REL-01
**Success Criteria** (what must be TRUE):
  1. Automated tests deterministically validate `Search` transport behavior and response expectations.
  2. Automated tests deterministically validate `Ask` transport behavior and response expectations.
  3. Automated tests deterministically validate request validation and gRPC status-code mapping paths.
**Plans**: TBD

### Phase 5: Workflow Resilience Tests
**Goal**: Scheduled workflows and external client integrations are covered by deterministic resilience tests.
**Depends on**: Phase 4
**Requirements**: REL-02, REL-03
**Success Criteria** (what must be TRUE):
  1. Scheduler tests deterministically verify fetch-and-notify execution flow on successful runs.
  2. Scheduler tests deterministically verify failure handling without nondeterministic timing flakiness.
  3. arXiv and Telegram client tests deterministically verify retry/non-200/error-path behavior.
**Plans**: TBD

## Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Contract Baseline | 0/TBD | Not started | - |
| 2. RAG Ask Delivery | 0/TBD | Not started | - |
| 3. Security Hardening | 0/TBD | Not started | - |
| 4. Transport Determinism | 0/TBD | Not started | - |
| 5. Workflow Resilience Tests | 0/TBD | Not started | - |
