# 283 — CONTINUE HERE (2026-08-19). Judged pipeline BUILT (Go + 486_HOLD); building it found `bugs_open/324` — 32/69 mechanical conversions dangle a binding, 14 SERVING. Repair BUILT. Everything waits on the next roll.

**Supersedes `283_CONTINUE_HERE_2026-08-18.md`.** Case file §14 is the record; design =
`bugfix_283_component_instance_scope/PLAN_2026-08-18_judged_pipeline.md` (+ §9 addendum);
the new defect = **`bugs_open/324`** (mechanism, census, classes, live-broken list).
Council round 6 submitted 2026-08-19 (same correlation `07635a2f…`) — READ THE VERDICT before
writing any `Council-Reviewed:`; the commit carries `Council-Submitted:`.

## Do next, in order (all gated on a roll carrying the 2026-08-19 commit)

1. **Verify the roll at the DIGEST** (RUNBOOK §1) AND
   `git merge-base --is-ancestor <the 2026-08-19 283 commit> <pod revision>`.
2. **Read the round-6 verdict** (RUNBOOK §7). REVISE → answer with measurements.
3. **Apply migration 486** (`sql_for_agents/486_judged_instance_scope_pipeline_HOLD.sql` —
   rename away from _HOLD first; its verify block PREPARE-compiles and asserts
   tolerate_truncation absent). Then **seed 487** (same renaming; expects ~69 items, ≥5 high).
   ⚠ 486 BEFORE 487: without the judged branch a repair refusal completes as a quiet no-op.
4. **Drain the repair batch** under a monitor. Expect ~27 `fixed:true` (each auto-filing
   page-scoped `template_changed` rerenders; all 14 serving-broken placements are generic
   pages, so delivery is the proven path) + 5 → `needs_human_review` (judged pool) + ~37
   no-ops. Done-check: `go run ./cmd/instanceaudit <converted-export> --bindings` exit 0.
   Spot-check repaired served pages **with a binding check** (no bare declared-id literal in
   the served script), not only ids/tokens — that blindness is how 324 shipped.
5. **Then the judged sequence** (PLAN §5 + addendum): owed steps (b2_verify rebaseline;
   oracle lockstep — LMC CONTRIB veto window filed 08-18, silence = no objection), canary
   `loans-standard-calc`, the 22, then the 2 generic tools + the 5 repair-refusals.
   The canary also proves the one [UNVERIFIED] link: empty-field_updates section_edit
   re-renders + deploys an LMC owned page.

## ⚠ Traps

- **A pre-2026-08-19 `scope_component_instance` green trail is NOT evidence the tool works**
  (LANDMINES, new entry). The batch's own spot-checks were green on 14 dead tools.
- **SQL migration numbers are a live race** — 484/485 were taken by two lanes in the minutes
  between listing and writing (renumbered to 486/487). Re-check the directory at apply time.
- **The judged LLM step must never gain `tolerate_truncation`** — 486's verify asserts it.
- **4 automation-savings placements are live-broken until the judged pipeline runs** — if it
  slips, snapshot rollback of 3 rows + rerender is the contained alternative (owner call,
  named in 324).
- The tree carries other sessions' WIP (broken `datahelpers` test build at the time of
  writing) — test via `git archive HEAD` overlay (RUNBOOK §3).
