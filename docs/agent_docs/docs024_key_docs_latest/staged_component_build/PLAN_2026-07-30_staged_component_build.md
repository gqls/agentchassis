# PLAN — staged component build with a gate per stage

**Lane adopted 2026-07-30** by `brochure_component_library` on owner direction:
*"This provenance and ladder project is now this lane's project."* Anchor feature:
`features_open/027`. The design argument and the evidence live in
`PROPOSAL_2026-07-30_step_by_step_build_with_stage_gates.md`; **this file holds
decisions and their reasons**, and corrections to the proposal are recorded here
rather than silently edited into it.

## What we are trying to do

Make a component build a sequence of small, individually-verified steps instead of
one leap, so that a more complicated component is *more stages of the same size*
rather than a bigger risk. Each small part carries its own travelling doc — the same
`doc_plans`/`doc_notes` machinery tools already use — and each stage has one question
and one gate that is capable of failing.

The originating evidence: `teaser-reveal-panel` took five rounds, and Round 5 found a
bug present since Round 1 — JavaScript that had never once run client-side. Four
rounds of honest, non-vacuous, served-artefact verification passed straight through
it, because every check read static markup or forced DOM state and none ever fired a
real click. **The gap was a missing stage, not missing rigour.**

## Decisions taken, and why

**D1 — Adopt 027's S0–S7 numbering; do not invent a parallel vocabulary.**
The tools lane independently sketched a four-stage version (skeleton → one real
behaviour → the rest → polish) and then deferred to S0–S7 as a superset. Two lanes
using one numbering is worth more than either lane's preferred names. *Reason:* a
forked vocabulary is the exact drift class the council reviews for.

**D2 — Every gate is validation; none is judgement.**
Forced by the owner's validation-versus-judgement correction to the tools lane. A gate
answers a closed question with a fixed rule and the same answer every time. *"Is this
component any good?"* is judgement and belongs to a reviewer seat. **The ladder has no
aesthetic gate and that is deliberate**, not an omission — conflating the two yields a
gate that drifts per component and a judgement boxed into a checklist.

**D3 — `skip` is not a pass, and this is the ladder's load-bearing correctness
requirement.** The Tier-4 runner skips an unknown check type
(`default: skip(ch.ID, ch.Type+" not implemented")`), and G4 means an all-skipped
result set reads as PASS plus a 7-day cooldown. For a *ladder* that is worse than for
a checklist, because stage N's pass licenses stage N+1. A stage that cannot evaluate
its question is **inconclusive**. *Reason:* discovered live — `has_visible_area` is
committed and not rolled, so the newest and most useful check type is currently the
one that would silently skip.

**D4 — The PLAN is the fleet-wide contract; the NOTES are the per-site verdicts.**
Not a preference — the schema already says so. `doc_notes` has `site_id`, `doc_plans`
does not. A component's template is fleet-shared (one `content_components` row serves
11 sites for `info-card-grid`) so a site-less PLAN is correct; S4–S7 are per-site
facts and land in NOTES.

**D5 — ~~Stage the enabling migration in the RUNBOOK, not in `sql_for_agents/`.~~**
~~The migration runner takes *every* pending file in a directory, so an unreviewed
`272_*.sql` could be swept in by an unrelated session's `--apply`. It gets a number
when it goes to the council gate.~~

> **CORRECTED 2026-07-30, same session — D5 WAS WRONG, and what caught it was reading
> the enforcement points instead of trusting my own read of the schema.**
> Two things I had not looked at made it wrong:
>
> 1. **The migration was never the whole change.** `subject_type` has a **second
>    enforcement point in Go** — `validDocSubjectTypes`
>    (`platform/orchestration/actions/doc_subjects_common.go`), which gates
>    `write_doc_plan`, `append_doc_note`, `load_doc_context` and
>    `persist_diagnosis_note`. Shipping the DDL alone would have reproduced
>    **`bugs_open/064` for the third time**: migration 163 missed the
>    `persist_diagnosis_note` gate, migration 184 moved the DB CHECKs *only* and left
>    its own seeded action docs unreachable through every doc action. The file's own
>    comment states the rule — *a value the DB accepts but a Go gate rejects is a split
>    contract; move both together.*
> 2. **The migration MUST be numbered, because a test parses it.**
>    `TestValidDocSubjectTypes_LockstepWithMigrationCheck` finds the newest **numbered**
>    `.sql` under `sql_for_agents/` that recreates `doc_plans_subject_type_check` and
>    fails if its ARRAY differs from the Go list. So withholding the number does not
>    protect anything — it **reddens HEAD** for every other session the moment the Go
>    edit lands. D5 as written was unbuildable.
>
> **Replaced by D5′: the Go edit and the numbered migration land in ONE commit, and the
> migration is not applied until an image carries the Go half.** The residual risk D5
> was worried about — another session's `--apply` sweeping the file in early — is real
> but *inert here*, because nothing writes component docs yet, so a widened CHECK ahead
> of the image has no effect. Shipped as `273_doc_subjects_component.sql` +
> `doc_subjects_common.go`, commit `c659e312b`, council correlation
> `e5673868-7c5b-489c-931a-7ba59b959b91`. **The lesson is the one this lane exists to
> make mechanical: I costed a change by reading one enforcement point and calling it
> "the smallest possible platform change". There were two.**

**D6 — Do not take ownership of `features_open/015`.**
The tools lane's decomposition (015 = rung vocabulary, 027 = gate mechanism, 026 =
missing instrument) makes the three composable rather than merged, which means this
lane can proceed without owning the site-scale ladder. Recorded as PROPOSED — whether
015 stays a separate thread is the owner's call.

**D7 — Prove a check type in the running binary before authoring a gate against it,
using a LONG marker.** Go compiles short string literals to immediate comparisons that
never reach rodata, so `grep -ac "selector_count"` returns 0 on a binary that fully
supports it. A negative from a short marker is worthless.

## Phasing

**P0 — adoption and design (this session).** Standing five created; the blocking
unknown resolved; the proposal updated with what the review found. **Done.**

**P1 — make a component documentable. SUBMITTED 2026-07-30, awaiting verdict.**
Both halves written, tested and committed (`c659e312b`); council correlation
`e5673868-7c5b-489c-931a-7ba59b959b91`; **migration 273 NOT applied** — image first.
Mutation-proven rather than merely green: with the Go half alone the lockstep test
fails naming 184's exact failure mode, and with both it passes.
Remaining in P1, in order: verdict → build/roll an image carrying the Go half →
pod-grep to prove it shipped (a roll is not evidence) → apply 273 → then one real
component (`teaser-reveal-panel`, because its history is fully written down, so nothing
has to be reconstructed) gets a PLAN with a criteria fence and its NOTES backfilled
from `NOTES_brochure_component_library.md`. Gate: the fence exists, passes the ten-rule
validator, and every criterion has been watched to pass by hand.

**P1a — the three-way naming contract check. NEW, and it jumps the queue** (2026-07-30,
on the first forward run's measured recommendation). Assert
`doc_plans.subject_key == pages.name == content_components.function` for every subject
that has a fence, and report the mismatches. **One query.** It is first because a
mismatch makes a fired run *skip and read as clean* — so every other stage's verdict is
untrustworthy until this passes, and it already has a known population of **6 of 22
hosted tools**. It also needs nothing from the blocked migration.

**P2 — make S6 real.** Dispatch a component's fence to `browser-runner-adapter` the
way `tool-acceptance-agent` does for tools. This is the stage whose absence cost five
rounds, and it is wiring rather than construction — the mechanism was proven end to
end on 2026-07-29 by the `smart-contrast` pilot. Gate: a deliberately broken
component makes it go red.

**P3 — the remaining gates**, cheapest first, each with its mutation.

**P4 — only then** stage-scoped dynamic generation of gates, and the anti-vacuous
verdict rule (D3), which changes a shared guarantee and therefore goes to the gate and
plausibly an RFC.

## What would make this lane wrong

Stated up front so it is falsifiable rather than defended later:

- ~~**If nothing fires the stages, the ladder is worthless.** This is the tools lane's
  G5, and it is the most likely way this fails.~~
  > **ANSWERED BY MEASUREMENT 2026-07-30, and the answer reframes the risk rather than
  > clearing it** (`REPLY_2026-07-30_vendor_trust_checklist_build.md` Q3). Firing by hand
  > is **cheap**: S6 end to end was **one script, 48 seconds** wall-clock, correlation
  > `dc952633`. So the trigger was never the binding constraint. **Addressability is.**
  >
  > Three values must be equal or a fired run quietly does nothing —
  > `doc_plans.subject_key == pages.name == content_components.function`. `load_docs`
  > keys on `spec.function`; a mismatch yields an empty fence and `request_browser_run`
  > **SKIPS with `needs_criteria`**: honest, but not a failure either, so **it reads as a
  > clean run that asserted nothing.** Measured fleet-wide: **6 of 22 hosted tools cannot
  > be acceptance-tested at all** until renamed, across five sites — with an honest
  > denominator note that including the three non-tool riders would read 9 of 25 and
  > "flatters the problem".
  >
  > Their conclusion, adopted: *"a ladder whose stages CAN be fired but silently resolve
  > to nothing is worse than one nobody fires, because it produces green."* **So the
  > highest-value single thing to build is not a trigger — it is the check that asserts
  > the three-way naming contract.** One query, and it would have found six broken tools
  > before anyone fired anything. It is now P1's first item.
- **If gates proliferate into dead config, it is a net negative.** `bugs_open/149` is
  the measured precedent: 22 discovery handler agents, only 2 running
  `validate_page_content`, six registered checks in no agent and zero items ever.
- **If the claim that stages would have caught Round 5 is wrong**, the whole argument
  weakens. It is marked `[INFERRED]` in the proposal and stays marked until a
  deliberately broken component is caught by an S6 gate nobody tuned for it.
