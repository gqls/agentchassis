# 360 — a section edit RESURRECTS a `removed` page_component, and the page publicly serves both tools

**Filed 2026-08-22** by the `webdesign_tool_rebuilds` lane, which took the damage. Instance
REPAIRED same morning (all four pages re-proven at the served bytes, 11:0xZ); the CLASS is live
and fleet-wide until the writers below are fixed.

> **On the 090 loop, stated plainly per the 2026-07-31 owner ruling.** This file asserts a
> cross-cutting cause in shared infra and did NOT go through the diagnosis loop. The substitution:
> every arm of the mechanism was OBSERVED live rather than inferred — the four status flips are
> timestamped inside the four section-editor claim windows (13:30–13:37Z 08-21, each bracketing its
> write to the second), the byte deltas match the transform exactly (+11 chars = one
> code-span→`<code>` conversion, md5 pairs recorded), the two code paths are quoted below from the
> current tree, and the public damage was confirmed cache-busted at the served bytes on both days.
> Nothing below rests on a grep hit whose function was not read; both deciding functions are quoted.
> A 090 run would also have raced the two ACTIVE lanes already holding the adjacent files (277, 356).

## Symptom

A page whose old component was retired (`page_components.build_status='removed'`, the documented
assembly-excluded tombstone) starts serving that component AGAIN — stacked above/below its
replacement — with no human action, no error, and every involved item reporting `complete`.

## Measured instance (webdesign.co.uk, site `6b49db8e-d447-4467-8277-4f3018af9897`)

- 2026-08-21 13:19:02Z: a quality-discovery sweep files 7 `literal_markdown` items (the 277 lane's
  canary for the new `rendered_html_transform` route — bugs_open/277 §5, migrations 499 + 513).
- Four target RETIRED ported slots (pages tool-grid-generator, tool-json-cleaner,
  tool-noise-generator, tool-text-extractor). Section-editor completes each; each tombstone flips
  `removed`→`approved` inside its claim window (e.g. grid: write 13:30:49, claim 13:30:17→13:30:56).
- Afternoon sweep rerenders assemble the pages with both slots (completes 15:19–16:29Z); 'approved'
  becomes 'deployed'. **All four pages publicly served two stacked tools for ~19 h** (measured at
  the served bytes, cache-busted: e.g. grid 24,545 B with `class="ported-page"`=1, vs 18,114 B
  single-tool after repair).
- Repair 2026-08-22: guarded re-retire (one txn, DO/RAISE asserts, post-commit re-read), four
  corrective assemble-only rerenders, serve-grade PASS ×4. Evidence trail:
  `docs/agent_docs/docs024_key_docs_latest/webdesign_tool_rebuilds/NOTES_…` entry 2026-08-22 11:06Z.

## Root cause — two defects in series, either alone is insufficient

1. **The filer scans tombstones.** `check_literal_markdown.go` Run (:391-398) and verifier (:311-)
   select `page_components` with ONLY `pc.locked_at IS NULL` — no `build_status` predicate. A
   `removed` slot is not on the page (assembly excludes it — `rerender_single_page_action.go:843`),
   but it IS on this check's audit surface, so a finding there routes a repair at content nobody
   serves. (`bugs_open/356` §6-B line 158 already names this check's missing PAGE-axis filter;
   this is the COMPONENT-axis sibling.)
2. **The writer promotes unconditionally.** `section_editor_actions.go` `updatePageComponentAfterEdit`
   (~:1436 and :1445) and `updatePageComponentSwap` (~:1472) run
   `UPDATE page_components SET … build_status='approved' … WHERE id=$1 AND <lock predicate>`.
   The only guard is `pageComponentAgentWritableSQL` → `datahelpers.AgentWritableSQLFor`
   (chrome_render_inputs.go:91), which tests LOCKS only. An edit to a `removed` row therefore
   un-deletes it. The target-resolution queries (:1225, :1303) don't exclude `removed` either.

Why nothing detected it: `check_page_component_status_drift` treats `removed` and `deployed` as
known statuses; the transient `approved` (also its worked 2026-07-09 example!) lived only minutes
before a sweep rerender normalised it to `deployed`.

## Blast radius

Fleet-wide, not site-scoped: migration 513 enables the transform on the ONE live section-editor
definition, and the same write helpers serve every `section_edit` producer. A SECOND producer with
the same hole one step earlier: **migration 486's `create_section_edit_delivery`** INSERT selects
placements `WHERE pc.component_id=$1 AND p.rebuild_policy='owned'` — no build_status filter — so
one template fix to the shared Ported Page component would file section_edits at EVERY owned
placement, including all 27 tombstones on webdesign.co.uk, resurrecting them in one batch.
Any workflow that uses `build_status='removed'` as a tombstone (every rebuild/replacement lane) is
exposed on every retire.

## Fix candidates, ordered by what closes the door

1. **Write path (makes the bad state unrepresentable for EVERY filer):** the section-editor's
   update helpers and target resolution refuse rows with `build_status='removed'` (add
   `AND COALESCE(build_status,'pending') <> 'removed'` to the UPDATE WHEREs; RowsAffected=0 already
   has a skip path — `errComponentLocked` — give tombstones their own typed refusal so the skip is
   attributable). An automated editor has no business writing a row the page does not contain.
2. **Filer scoping:** `check_literal_markdown` Run + verifier exclude `removed` rows. Note the
   verifier consequence is CORRECT: an item filed pre-retire then completes vacuously post-retire
   (the page genuinely no longer shows the defect).
3. **The 486 arm:** same predicate in `create_section_edit_delivery`'s INSERT…SELECT (config
   migration, live immediately — the cheapest of the three and the one guarding a BATCH resurrect).
4. (Optional detector) a page with >1 non-`removed` slot where one is `ported-page` is site-shaped;
   the general invariant "no automated write raises build_status FROM removed" is better enforced
   at (1) than detected after the fact.

Coordination: (2) sits in the 277 lane's active route (their canary claim "7/7 repaired and proven
at the served bytes" checked markdown absence, not slot count — corrected via CONTRIB into their
lane dir); the posture registry from `bugs_open/356` (`page_lifecycle_posture_test.go`, commit
`24d0bc251`) declares this check's PAGE-axis posture — a fix to (2) should update that registry
entry in the same commit or CI will disagree.

## How to verify a fix

Recreate the arm harmlessly: pick any `removed` slot, plant a code span in its `rendered_html`
(`UPDATE … SET rendered_html = replace(...)` on a copy/test row), run the check → it must NOT file;
dispatch a section_edit at the row directly → the write must refuse and the item must complete with
the typed skip, and the row must still read `removed` after. Then the motivating case: re-run the
2026-08-21 13:19Z canary shape against webdesign.co.uk and confirm zero items name tombstoned slots.

## Fix candidate (1) BUILT — 2026-08-22, this lane

The door-close half is committed: `section_editor_actions.go` gains an advisory tombstone gate
(skip-result `{tombstoned:true}`, mirroring the lock gate) plus the race-free predicate
`COALESCE(build_status,'pending') <> 'removed'` on all three page_components UPDATEs (shared const
`pageComponentNotRemovedSQL`); zero-rows still surfaces via the existing skip-convertible sentinel.
Mutation-proven captured-SQL test (`section_editor_tombstone_guard_test.go`) — predicate deleted
from the swap statement alone failed exactly that case. Proven against `git archive HEAD` + the two
changed files (the tree carries other lanes' WIP). Council `Council-Submitted:
4007ce96-4cc7-4d52-bdfe-cd00deeeca89`. **Go — INERT until the next chassis roll**; the defect stays
reproducible in production until then, so this file stays OPEN and the tool-rebuilds RUNBOOK's
post-retire re-read rule stays in force. Candidates (2) filer scoping and (3) the 486 INSERT
predicate remain with the 277/283 lanes (CONTRIBs delivered).
