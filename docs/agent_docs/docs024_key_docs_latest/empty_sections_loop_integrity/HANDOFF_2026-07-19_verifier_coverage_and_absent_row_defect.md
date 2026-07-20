# POINTER — a live defect in the `empty_section` verifier you built

**2026-07-19, reasoning-dataset thread. This is a pointer, not the substance.**

`VerifyEmptySectionResolved` (`check_empty_sections.go:205`) — the verifier this
workstream built, and still the only one registered on the platform — reports
**success when its target row is absent**:

```go
	if err == sql.ErrNoRows {
		// Component removed — nothing left to be empty.
		return VerifyResult{Resolved: true, Detail: "component no longer exists"}, nil
	}
```

A missing `page_components` row is equally the signature of a rebuild silently
deleting the component. So a content-loss incident is recorded as a *verified
fix* — by the mechanism built to stop `complete` being taken on trust. Found by
the council gate's `bug_historian` seat while reviewing a plan to copy this
branch to two more item types.

**Where the substance lives:**

- **`bugs_open/032`** — the case: evidence, conservative fix (return an error so
  the gate fails open and records "could not verify" instead of asserting
  success), the stronger option, and verification queries.
- **`work_item_completion_integrity/HANDOFF_2026-07-19_verifier_absent_row_defect_and_coverage.md`**
  — the full handoff. That thread owns `CompleteWorkItemAction` and had already
  named this gap in its own PLAN, so it holds the primary.
- **`bugs_open/021` §INSTANCE 2** — the coverage half: `RegisterVerifier` has been
  called once, for ~50 item types.
- **`016b` §9** — the transferable pattern (*a verifier that treats a missing
  target as success cannot distinguish repair from deletion*).

Flagged here because you built the verifier and `empty_section` is the item type
carrying the defect — you may hold context on whether a removed component *should*
read as resolved that neither the council nor we have. If so, say so on `032`;
the conservative fix is deliberately reversible.
