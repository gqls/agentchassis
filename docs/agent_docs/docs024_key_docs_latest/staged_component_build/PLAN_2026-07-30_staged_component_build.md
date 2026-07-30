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

**D5 — Stage the enabling migration in the RUNBOOK, not in `sql_for_agents/`.**
The migration runner takes *every* pending file in a directory, so an unreviewed
`272_*.sql` could be swept in by an unrelated session's `--apply`. It gets a number
when it goes to the council gate. *Reason:* a staged artefact that another session can
apply by accident is not staged, it is armed.

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

**P1 — make a component documentable.** The `subject_type='component'` migration
through the council gate, then one real component (`teaser-reveal-panel`, because its
history is fully written down) gets a PLAN with a criteria fence and its NOTES
backfilled from `NOTES_brochure_component_library.md`. Gate: the fence exists, passes
the ten-rule validator, and every criterion has been watched to pass by hand.

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

- **If nothing fires the stages, the ladder is worthless.** This is the tools lane's
  G5, and it is the most likely way this fails. Discovery passes are manual-fire and
  the improvement loop is stopped by owner ruling. **A ladder with no trigger is a
  mechanism rotting unexercised**, which is the cost the owner has already ruled
  against paying. P2 must name who fires it.
- **If gates proliferate into dead config, it is a net negative.** `bugs_open/149` is
  the measured precedent: 22 discovery handler agents, only 2 running
  `validate_page_content`, six registered checks in no agent and zero items ever.
- **If the claim that stages would have caught Round 5 is wrong**, the whole argument
  weakens. It is marked `[INFERRED]` in the proposal and stays marked until a
  deliberately broken component is caught by an S6 gate nobody tuned for it.
