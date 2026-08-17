# RFC 035 — loop expansion rewrites ANY reference-shaped sibling-output config value (WFA-017): ratify, narrow, or revert

## STATUS: OPEN — filed 2026-08-17 by the bugfix-287 lane, at the council's direction (guardian VETO + architecture seat `needs_rfc`, round 1, corr `cba35b35`). The code is COMMITTED (`0ed96c7eb`) and will ride the next chassis roll unless the owner directs a forward revert first.

## 1. What happened, plainly

Fixing `bugs_open/287` (spawn_record slug — dispatch loops storing the spawn record as a work
item's `result`), the fixing lane shipped two halves:

- **Config half** — RFC_029 `!` strict markers + an error route on the three dispatch agents'
  `mark_complete` (migrations `448` applied / `452_HOLD` gated). The council's guardian seat
  explicitly endorsed this half ("a contained, agent-scoped fix … should proceed").
- **Go half (THIS RFC's subject)** — `prefixConfigStepReferences` (coordinator.go), the loop
  expansion's config rewriter, stopped enumerating: instead of suffixing sibling-output
  references only for an 8-key allow-list + `input_mapping`, a generic pass now rewrites ANY
  top-level string config value that is *reference-shaped*
  (`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z0-9_]+)*$`) and whose first dotted segment names a
  sibling `output_field`. Registered as **WFA-017**; mutation-proven tests
  (`loop_config_reference_suffixing_test.go`).

The guardian vetoed the Go half on SCOPE: a fleet-wide change to shared loop-expansion
behaviour, justified as closing a class of future bugs, riding a 3-agent bug fix — "exactly
'architecture change dressed as a point fix'". The architecture seat: `needs_rfc`, split from
the migrations. Per the 2026-07-28 owner ruling (a scope veto is not answered by resubmitting;
record it where the change lives, route the seam here, let a human break it) — this file.

## 2. The question for the owner

**Should loop expansion's data-reference rewriting be enumerated (allow-list) or shape-based
(generic)?** Three options, costed in §5. This is a real convention choice: the generic pass
makes "a loop-substep config string that names a sibling output" MEAN a reference, fleet-wide,
for every future workflow author.

## 3. Evidence base (all measured, 2026-08-17; queries in `bugfix_287_spawn_record/RUNBOOK`)

- The allow-list's own comment says *"IMPORTANT: Any config key that references step outputs
  must be listed here"* — and three actions had already broken the promise unnoticed:
  `complete_work_item` (`result`, `work_item_id`, `commit_sha`), `mark_maintenance_complete`
  (`result_field`), plus the `condition` strings. That gap is one of 287's two doors.
- Census of every live loop substep: **22 config strings across 7 agents** have a sibling-
  output first segment through a non-allow-listed key; **all 22 are read-references**; zero
  literals, zero write-destinations; the only expression-shaped values are 3 `condition`
  strings, which the shape gate excludes by construction.
- Nested sibling-output strings exist ONLY inside `input_fields` ARRAYS (Strategy-1 field
  names) — the generic pass is top-level-string-only and a test pins the arrays untouched.
- Behaviour delta today ≈ **path, not value**: `propagateIterationOutputs` runs before every
  substep (coordinator.go:1355), so the base key a dotted reference resolves today holds the
  same value as the suffixed key the generic pass makes it resolve. (This same fact is why
  the strict markers alone close 287 — the veto is right about that, and 287 §11 says so.)

## 4. The seats' substantive concerns, kept whole

1. **Guardian:** a snapshot census "only proves today's config is safe — any future or
   currently-unindexed agent config matching the shape gets silently reprefixed too". True:
   the census is a point-in-time survey; the generic pass is a standing convention.
2. **Tooling-provenance:** `referenceShapedConfigValue` is a SECOND definition of "this
   string is a reference", beside the resolver's own context-dependent classification
   (`IsDottedPathReference`, Strategy-4 dotless references, spec-Default literals — the
   CTS-059 landmine: a dotless string means VALUE on a spec-defaulted field and REFERENCE on
   every other one). Two classifiers for one concept is the drift class this estate reviews for.
3. **Editquality:** the shape gate bounds what gets REWRITTEN (first-segment collision), but
   the pass EXAMINES every top-level string; an enum-like value (`mode: fast`) that happens
   to collide with a sibling output name would be rewritten. Censused zero today; not zero by
   construction.
4. **Reuse:** the estate now has this pass AND `input_mapping`'s own allow-list gate
   unreconciled — two gates for one problem, mirroring a known pattern.

## 5. Options, costed

**(1) Ratify WFA-017 as-is (generic, on by default).**
Cost: the concerns in §4 stand; the mitigation is the register entry's LANDMINE line ("a
loop-substep literal that exactly names a sibling output_field is now a reference") plus the
census query for audits. Benefit: the 287 door class is closed for every future action; the
allow-list promise (already broken three times) stops existing; conventions-over-enumeration
is how `input_mapping` already behaves.

**(2) Narrow: keep the pass but make it opt-in per key (`*_from`-style) or per action** —
i.e. revert to enumeration with `result`/`work_item_id`/`commit_sha`/`result_field` added.
Cost: restores the exact promise that was broken unnoticed three times, and the next action
author must know the list exists (they demonstrably do not); the 2026-08-02 §2 ruling's
reasoning ("a comment is not a control on a tree this many sessions share") cuts against it.
Benefit: zero risk to unexamined configs; smallest possible surface.

**(3) Revert the Go half (forward commit) and rely on strict markers + propagation.**
Cost: `!` on loop agents then resolves the BASE key via the per-substep propagation
side-channel — correct today (287 §11's measured correction), but coupled to
`setLoopVariable`'s silent early-returns, and every future loop action with an un-listed
reference key silently re-enters the whole-tree search until someone marks it strict.
Benefit: shared mechanism untouched; 287 still closed (the migrations do that alone).

The filing lane's recommendation is **(1)**, with (2) as the acceptable fallback; the
evidence for "the allow-list is a promise nobody keeps" is the strongest single fact in the
file. But the guardian is right that this is a convention decision, not a bug fix, and it is
the owner's to make.

## 6. Interactions

- **Migration `452_HOLD` does NOT depend on this ruling** (guardian's own note): strict-on-base
  works on the current binary. Its gate text now says both readings.
- If (3) is chosen, revert forward (`git revert` of the coordinator.go hunk + test file),
  keep WFA-017's register entry with status REVERTED (the register records what existed).
- RFC_029 Phase 2's triage reads these same resolver rows; its lane was notified via
  RFC_029 §10.6b either way.

## Sources

`bugs_open/287` §9a/§11 · council report corr `cba35b35` (verdict REJECTED, guardian veto;
`diagnosis_artifacts` kind=`council_report`) · `0ed96c7eb` (the committed change + tests) ·
WFA-017 (`register/workflow-authoring.md`) · CTS-059/CTS-060 · the 2026-07-28 scope-veto
ruling and `bugs_closed/124` / BLD-019 precedents (code-stays-pending-human, "Live ≠ approved").
