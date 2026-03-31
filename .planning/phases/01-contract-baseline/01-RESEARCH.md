# Phase 1: Contract Baseline - Research

**Researched:** 2026-03-31
**Domain:** Go gRPC contract/runtime alignment and env configuration validation
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
No CONTEXT.md found for this phase.

### Claude's Discretion
No CONTEXT.md found for this phase.

### Deferred Ideas (OUT OF SCOPE)
No CONTEXT.md found for this phase.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| APC-01 | gRPC proto contract and server implementation remain aligned (no declared-but-unimplemented public RPCs). | Contract baseline should expose only implemented public RPCs; remove or implement any proto-declared RPC that currently falls through to generated `Unimplemented...` defaults. |
| APC-02 | Runtime env configuration matches actual usage (no required secrets that are unused). | Config validation should require only runtime-used values; startup error messages should reference actual required keys and why they are needed. |
</phase_requirements>

## Summary

Phase 1 is primarily a correctness and operability phase, not a feature phase. The current codebase has contract drift: `proto/arxiv/v1/arxiv.proto` declares `Ask`, while `internal/server/grpc/arxiv.go` only implements `Search` and inherits default `Ask` behavior from `UnimplementedArxivServiceServer` (runtime `codes.Unimplemented`). This violates APC-01 because the public contract advertises behavior that does not exist.

There is also configuration drift: `internal/app/config.go` marks `GROQ_API_KEY` as required, but the value is not consumed by any runtime path in Phase 1 scope. This violates APC-02 and causes unnecessary startup failures for operators. In addition, startup validation currently relies only on `envconfig` tag failures, so messaging can be technically correct but semantically unclear about runtime necessity.

**Primary recommendation:** For Phase 1, make the public API behavior-first (`Search` only until `Ask` exists), and make config requirements usage-driven (remove unused required secrets and enforce explicit validation for truly required runtime keys).

## Project Constraints (from CLAUDE.md)

- Use Go toolchain and existing architecture conventions; preserve brownfield compatibility.
- Keep existing `Search` API and scheduled notification behavior stable.
- Use existing command/tooling conventions:
  - `go test ./...`
  - `go fmt ./...`
  - `go vet ./...`
  - `buf lint`
  - `buf generate`
- Do not edit generated protobuf stubs manually (`internal/gen/` is generated output).
- Keep constructor/option and interface-at-point-of-use patterns already established in the repository.
- Maintain env-based operational simplicity for service startup.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | 1.26.1 (installed), `go 1.26` (module) | Runtime and build toolchain | Native project language and test/build pipeline baseline |
| `google.golang.org/grpc` | v1.79.2 | gRPC server runtime and status mapping | Canonical Go gRPC stack already wired in app startup |
| `google.golang.org/protobuf` | v1.36.11 | Proto runtime/types | Required by generated transport models |
| `github.com/kelseyhightower/envconfig` | v1.4.0 | Environment loading into `app.Config` | Existing config contract mechanism used by `main/main.go` |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `buf` CLI | 1.50.0 (installed) | Proto lint + codegen | Any time proto RPC surface is changed |
| `go.uber.org/zap` | v1.27.1 | Structured startup/runtime logging | Validation and operator-visible error context |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Removing unimplemented RPC from proto | Keep RPC and return `Unimplemented` intentionally | Preserves forward placeholder, but violates APC-01 and misleads callers |
| `envconfig` + explicit validation | Hand-rolled map/env parsing | More control, but unnecessary complexity and divergence from existing stack |

**Installation:**
```bash
go mod download
```

**Version verification:** Verified against local project/toolchain state (`go.mod`, `go version`, `buf --version`) on 2026-03-31.

## Architecture Patterns

### Recommended Project Structure
```text
main/
  main.go                    # env load + lifecycle + signal handling
internal/app/
  config.go                  # config schema + validation rules
  app.go                     # composition root and service registration
internal/server/grpc/
  arxiv.go                   # RPC handlers for declared public API
proto/arxiv/v1/
  arxiv.proto                # source-of-truth public contract
internal/gen/arxiv/v1/
  *.pb.go, *_grpc.pb.go      # generated artifacts from proto
```

### Pattern 1: Behavior-First API Contract
**What:** Proto service methods should only include RPCs with non-placeholder runtime behavior.
**When to use:** Anytime transport contracts are ahead of implementation.
**Example:**
```proto
service ArxivService {
  rpc Search(SearchRequest) returns (SearchResponse);
}
```
Source: `proto/arxiv/v1/arxiv.proto` (phase recommendation for APC-01 alignment)

### Pattern 2: Usage-Driven Required Config
**What:** Mark env keys required only when a runtime code path consumes them in current phase behavior.
**When to use:** Startup config schema changes and operator runbook updates.
**Example:**
```go
type Config struct {
  Address     string `envconfig:"ADDRESS" default:":8080"`
  GRPCAddress string `envconfig:"GRPC_ADDRESS" default:":9090"`
  TelegramConfig
}
```
Source: `internal/app/config.go` (phase recommendation for APC-02 alignment)

### Anti-Patterns to Avoid
- **Proto placeholders in public service definition:** Declaring RPCs before implementation causes runtime `Unimplemented` drift.
- **Required-but-unused secret keys:** Forces irrelevant operator setup and obscures true runtime requirements.
- **Validation by tags only:** Produces generic missing-var failures without clear runtime intent.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Proto/server contract diffing | Custom parser for `.proto` vs handlers | `buf lint` + generated server interface constraints + focused transport tests | Lower maintenance and aligned with project tooling |
| Env parsing framework | Custom os.Getenv parser/validator engine | Existing `envconfig` + explicit `Validate()` in `app.Config` | Keeps compatibility and improves clarity with minimal change |
| gRPC status semantics | Manual numeric code handling | `status.Error` + `codes.*` | Stable, idiomatic transport behavior |

**Key insight:** The repo already has the right primitives; Phase 1 needs stricter alignment policy, not new infrastructure.

## Common Pitfalls

### Pitfall 1: “Implemented enough” assumption from embedding `Unimplemented...`
**What goes wrong:** Service compiles while still exposing declared RPCs that always fail with `codes.Unimplemented`.
**Why it happens:** Generated embedding hides missing explicit method implementations.
**How to avoid:** Treat any proto RPC without explicit handler method as contract drift for this phase.
**Warning signs:** Handler type has `UnimplementedArxivServiceServer` but no corresponding method for every proto RPC.

### Pitfall 2: Required env tags drifting ahead of feature rollout
**What goes wrong:** Startup fails on keys not needed for active runtime behavior.
**Why it happens:** Config schema prepared for future phases but enforced immediately.
**How to avoid:** Gate “required” by active usage; add explicit validation and phase-scoped requirements.
**Warning signs:** `required:"true"` fields with zero call sites in runtime paths.

### Pitfall 3: Validation errors that are syntactically correct but operationally vague
**What goes wrong:** Operators see missing env names without clear “required for what” context.
**Why it happens:** Framework default errors are not wrapped with runtime intent.
**How to avoid:** Add config validation step with contextual messages.
**Warning signs:** Startup fatal exits directly on raw `envconfig` error strings.

## Code Examples

Verified patterns from current codebase:

### gRPC input validation + status mapping
```go
if req.Query == "" {
  return nil, status.Error(codes.InvalidArgument, "query is required")
}
```
Source: `internal/server/grpc/arxiv.go`

### Startup env loading as single entry point
```go
var cfg app.Config
err := envconfig.Process("arxiv-rag-go", &cfg)
if err != nil {
  log.Fatal(err.Error())
}
```
Source: `main/main.go`

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Contract-first placeholder RPC (`Ask` declared before implementation) | Behavior-first contract (declare only implemented RPCs) | Phase 1 target (2026-03-31 planning) | Eliminates declared-vs-runtime drift for callers |
| Future-feature required secrets in baseline startup | Usage-driven required config keys | Phase 1 target (2026-03-31 planning) | Reduces operator friction and false startup blockers |

**Deprecated/outdated:**
- Proto-declared public RPCs without concrete server behavior in same phase.
- Required secrets with no runtime call site in active feature set.

## Open Questions

1. **Should `Ask` be removed from proto in Phase 1 or minimally implemented behind real behavior?**
   - What we know: Phase 2 owns `Ask` delivery; current `Ask` behavior is generated default `Unimplemented`.
   - What's unclear: Whether external clients already depend on proto containing `Ask`.
   - Recommendation: Default to removing `Ask` for APC-01; if compatibility is required, explicitly document and accept temporary APC-01 exception (not preferred).

2. **Should env prefix in docs (`PRODUCER_*`) be corrected in this phase?**
   - What we know: Runtime prefix in code is `arxiv-rag-go`; CLAUDE header table references `PRODUCER_*`.
   - What's unclear: Whether this table is historical or currently used by operators.
   - Recommendation: Update docs in same phase if touched by config contract work to avoid operator confusion.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | Build/test/validation | ✓ | go1.26.1 | — |
| buf CLI | Proto lint/codegen after contract edit | ✓ | 1.50.0 | `protoc` direct workflow (less preferred) |
| protoc | Codegen backend | ✓ | 3.21.12 | Use buf-managed generation |
| grpcurl | Manual RPC surface checks | ✓ | dev build | Use Go integration tests instead |
| Docker | Container startup verification (optional) | ✓ | 28.4.0 | Local `go run ./main` |

**Missing dependencies with no fallback:**
- None.

**Missing dependencies with fallback:**
- None.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` package (stdlib) |
| Config file | none |
| Quick run command | `go test ./...` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| APC-01 | Public gRPC methods in proto match implemented server methods for exposed service | integration/transport | `go test ./internal/server/grpc -run TestArxivServiceContract -count=1` | ❌ Wave 0 |
| APC-02 | Startup validation requires only runtime-used env keys and returns clear missing-key errors | unit | `go test ./internal/app -run TestConfigValidation -count=1` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./...`
- **Per wave merge:** `go test ./...`
- **Phase gate:** `go test ./...` plus `buf lint` and `buf generate` clean

### Wave 0 Gaps
- [ ] `internal/server/grpc/arxiv_contract_test.go` — verifies proto/service method alignment for APC-01.
- [ ] `internal/app/config_test.go` — validates required/optional env behavior and human-readable errors for APC-02.
- [ ] `internal/app/startup_validation_test.go` or equivalent — verifies startup path surfaces validation messages.

## Sources

### Primary (HIGH confidence)
- `proto/arxiv/v1/arxiv.proto` - declared public RPC surface (`Search`, `Ask`).
- `internal/server/grpc/arxiv.go` - implemented server methods and status mapping.
- `internal/gen/arxiv/v1/arxiv_grpc.pb.go` - generated default `Unimplemented` behavior for missing RPC implementations.
- `internal/app/config.go` - env requirement schema (`required:"true"` fields).
- `main/main.go` - env loading entry point and startup failure behavior.
- `.planning/REQUIREMENTS.md` - requirement definitions APC-01/APC-02.
- `.planning/ROADMAP.md` - phase goal, dependencies, and success criteria.
- `CLAUDE.md` - project constraints, commands, and architecture conventions.

### Secondary (MEDIUM confidence)
- `go.mod` - dependency versions used by the repository.
- Local toolchain checks (`go version`, `buf --version`, `protoc --version`, `docker --version`, `grpcurl -version`).

### Tertiary (LOW confidence)
- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - directly verified from `go.mod` and installed tool versions.
- Architecture: HIGH - derived from current implementation layout and startup wiring.
- Pitfalls: HIGH - directly evidenced by current proto/handler/config mismatches.

**Research date:** 2026-03-31
**Valid until:** 2026-04-30
