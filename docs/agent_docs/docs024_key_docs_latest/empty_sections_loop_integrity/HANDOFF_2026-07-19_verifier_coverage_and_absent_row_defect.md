# HANDOFF — the completion verifier: one live defect, one coverage gap, one reviewed plan

**To:** the `empty_sections_loop_integrity` thread — you built the completion
gate and the only registered verifier, so this is yours.
**From:** the reasoning-dataset thread, 2026-07-19. We are read-only for
`platform/` and are not implementing any of this.
**Cost already paid:** two council-gate rounds. The objections are answered below
so you do not re-spend them.

---

## TL;DR

Your completion-verification framework is good and it is doing almost nothing,
because `RegisterVerifier` has been called **once**. While proposing to extend
it, the council found a **defect in the one verifier that exists**: it reports a
*deleted* component as a *successfully fixed* one.

Three artefacts, in priority order:

| # | what | where | who decides |
|---|---|---|---|
| 1 | **Live defect** — verifier reads a deleted target as success | `bugs_open/032` | you — one branch, conservative fix drafted |
| 2 | **Coverage gap** — 1 of ~50 item types verified | `bugs_open/021` §INSTANCE 2 | you — needs a policy call |
| 3 | **A reviewed plan** for both, twice through the council | `reasoning_dataset/submission_B_register_more_item_verifiers.json` | you — take, adapt or bin it |

---

## 1. The live defect (`bugs_open/032`) — read this one first

`VerifyEmptySectionResolved` (`check_empty_sections.go:205`):

```go
	if err == sql.ErrNoRows {
		// Component removed — nothing left to be empty.
		return VerifyResult{Resolved: true, Detail: "component no longer exists"}, nil
	}
```

A missing `page_components` row is equally the signature of a rebuild silently
deleting the component — the failure class `021` and `012` document, that this
platform has paid for at least twice. So a content-loss incident is recorded as a
**verified fix**, by the very mechanism that exists to stop `complete` being
taken on trust.

Found by the council's `bug_historian` seat, in its own words:

> *"the detection layer improves, but the new check adopts the same blind spot
> that caused the original content-loss incidents."*

**Conservative fix** (full rationale and verification queries in `032`): return an
**error**, not a verdict. Your gate already fails OPEN on verifier error, so item
flow is unchanged — the item still completes — but `result._verification` records
that verification *could not be made* instead of asserting success. A silent
false positive becomes a visible unknown.

**Your call, and the better answer if you want it:** if the page still expects
that component (a `plan_sections` entry, a slot reference), absence is *not*
ambiguous — it is deletion, and `Resolved: false` is correct. Bigger change; the
error-return is the safe floor and does not preclude it. Do **not** return
`Resolved: false` unconditionally — a legitimately removed component would burn
an attempt and strand a fine item in `failed`.

## 2. The coverage gap (`bugs_open/021` §INSTANCE 2)

Filed as an instance on 021 rather than a new bug, because it is exactly that
file's pattern — *one call site patched, mechanism left generic*.

```
RegisterVerifier calls in the entire codebase : 1   (check_empty_sections.go:38)
item types with a discovery check             : ~50
site_work_items status='complete'             : 4,570   (live 2026-07-19)
  ...with a result._verification record       :     5
ever status='verified'                        :     9   (all empty_section,
                                                         none since 2026-07-14)
```

The mechanism is opt-in by construction (`verifiers.go:47-51`: *"Called from
init() in check files"*), so it stays at one unless an author remembers. That is
a policy question for you, and there are two shapes:

- **Register more verifiers** — direct, and each one is real coverage.
- **Make the gap fail loud** — a `VerifierCoverage()` helper plus a test that
  breaks the build when a check has no verifier and is not on an explicit
  known-gap list with a reason. A council seat reviewing this warned it must not
  iterate *only* the check registry: an item_type created by a path that never
  registered a discovery check would be invisible to the guard, under-reporting
  the very gap it exists to expose. Worth heeding — that is the same
  one-layer-up blindness as defect 1.

## 3. The plan, and what the council did to it

`reasoning_dataset/submission_B_register_more_item_verifiers.json`
(`SUBMISSION_CORR=66dbd0dd-de5f-4f50-acd3-f5f3d817dbd9`, two rounds, both REVISE).

**Take it as a starting point, not a finished plan.** It proposes: fix defect 1 at
source, register verifiers for `hardcoded_section_colors` (180 items / 121
complete) and `phantom_internal_link` (63 / 37), extract each check's predicate
into a shared pure function so detection and verification cannot drift, add a
test for the extracted predicate, and add the coverage guard.

**Objections still outstanding — do not re-discover these:**

- **The `phantom_internal_link` edit is a stub.** `editquality` called it *"a stub
  dressed as an edit"* — comments describing intent, no query, no return, while
  the other edits carry real logic. Correct. Whoever implements needs to read
  `check_phantom_internal_links.go` and write the real shared predicate.
- **`VerifierCoverage()` sees too little** — see §2 above.
- **The known-gap allowlist is sketched, not enumerated.** ~47 entries needed; the
  test cannot compile until it is populated and accurate.
- **An internal contradiction in the plan's own evidence**: it cites 9 items at
  `status='verified'` while asserting no Go code sets that status. Both are true —
  no Go writer exists (grep finds only `business_intel.businesses.verification_status`,
  a different table), so the 9 are historical or hand-set — but the plan never
  reconciles them. Worth establishing properly before anyone relies on `verified`
  meaning anything.
- **`image_url_404` was dropped** and should stay dropped: 4 items, 0 completions,
  and it was the only candidate adding an outbound HTTP call to the completion
  path.

**Scope note the plan holds deliberately, and we suggest you keep:** it does not
stamp `status='verified'`. No Go code sets that on `site_work_items` today, and
adding a terminal status touches the `workItemTerminalStatuses` / `idx_swi_dedup`
lockstep that has caused fleet-wide 42P10 before. Evidence lands in
`result._verification`, which is additive.

## Why a dataset thread cares (declare the interest)

We wanted `complete` to mean something checkable, because a self-reported
completion is a poor training label. That motivation is why we looked; it is not
why this matters. Defect 1 masks content loss in production whether or not anyone
ever builds a dataset — that is the reason to fix it, and the reason we are
handing it to you rather than arguing for our own use case.

## Reading the council reports yourself

```sql
SELECT r->>'reviewer', r->>'verdict', jsonb_pretty(r->'objections')
FROM diagnosis_artifacts da, jsonb_array_elements(da.body::jsonb->'reviews') r
WHERE da.kind='council_report'
  AND da.correlation_id='66dbd0dd-de5f-4f50-acd3-f5f3d817dbd9'
ORDER BY da.created_at;
```

Also: `016b` §9 now carries the transferable pattern (*a verifier that treats a
missing target as success cannot distinguish repair from deletion*), so nobody
re-walks it.
