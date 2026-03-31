# Phase 1: Survey Filter Delivery - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-31
**Phase:** 1-Survey Filter Delivery
**Areas discussed:** Keyword Matching, Category Matching, Missing Metadata, Edge Phrases

---

## Keyword Matching

| Option | Description | Selected |
|--------|-------------|----------|
| Substring | Case-insensitive substring match in title/abstract; simplest and catches variants | ✓ |
| Whole-word only | Match on word boundaries; reduces false positives but may miss variants | |
| Exact phrase only | For multi-word phrases, require exact phrase match; stricter | |

**User's choice:** Substring
**Notes:** Case-insensitive substring matching across title and abstract.

---

## Category Matching

| Option | Description | Selected |
|--------|-------------|----------|
| Any category match | Paper is eligible if any category matches configured list | ✓ |
| Primary category only | Only primary category counts | |

**User's choice:** Any category match
**Notes:** Use any category in the paper's category list.

---

## Missing Metadata

| Option | Description | Selected |
|--------|-------------|----------|
| Title-only match | If abstract missing/empty, match on title only | ✓ |
| Drop the paper | Skip if abstract missing | |

**User's choice:** Title-only match
**Notes:** Allow title-only matching if abstract missing.

---

## Edge Phrases

| Option | Description | Selected |
|--------|-------------|----------|
| Flexible spacing/hyphen variants | Allow flexible spacing/hyphen variants for multi-word phrases | ✓ |
| Exact phrase only | Require exact phrase match | |

**User's choice:** Flexible spacing/hyphen variants
**Notes:** "state of the art" should match "state-of-the-art".

---

## the agent's Discretion

None.

## Deferred Ideas

None.
