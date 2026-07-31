# 027 FEATURE — build a site part step by step, with a travelling doc and a gate per stage

**Raised:** 2026-07-30, owner, to the `fundamentallyai.com 4` session
(`brochure_component_library` lane), after the `teaser-reveal-panel` carousel took five
rounds and the fifth found a bug present since the first.
> **SCOPE CUT 2026-07-31 (owner ruling, PLAN D8): EIGHT GATES → THREE.** The eight-stage
> ladder stays as a **checklist**, not machinery. Built machinery is limited to: the claim
> written before the build; verification through the visitor's real gesture; and every
> check proven able to fail (a mutant counts only if the artefact provably changed).
> Decisive reason: **two of the eight gates were wrong on first contact with reality and
> one would have BLOCKED a correct build** — see `SUMMARY_2026-07-31_we_cut_the_ladder_down.md`.
> **First item is now `CHECK_naming_contract.sh`, not the substrate work** — it currently
> FAILS with 2 tools carrying a fence that can never run, one of them ours.

**Status:** **ADOPTED 2026-07-30 — OWNED by the `brochure_component_library` lane**
(owner: *"This provenance and ladder project is now this lane's project"*). Designed,
nothing built yet; the blocking unknown is resolved. ~~Owner wants the work done in a
SEPARATE thread.~~ Superseded — the lane that produced the evidence owns it.

**Before routing work at this feature, read the lane's docs — it has an owner and a
plan.** Cold-start: `docs024_key_docs_latest/staged_component_build/` (standing five,
created 2026-07-30). The design brief is
`PROPOSAL_2026-07-30_step_by_step_build_with_stage_gates.md`; decisions D1–D7 and the
phasing are in `PLAN_2026-07-30_staged_component_build.md`; **`RUNBOOK_…` carries the
DDL that unblocks P1 and the pod-grep that must precede any gate.** Contribute into
those rather than competing.

## Resolved since filing (2026-07-30)

- **The `[UNVERIFIED]` first action is DONE.** `doc_plans`/`doc_notes` do **not** fit a
  component today: both carry a CHECK on `subject_type` and neither allows `component`.
  The fix is one additive migration extending two constraints, with a four-times
  precedent (163, 184, 218, **270** — the template). Normal council-gate scope, not an
  RFC. **Trap:** the `doc_notes` re-add must keep `'landmine'` or it orphans 57 live
  rows another thread wrote.
- **Good news in the same read:** `doc_notes` has `site_id` and `doc_plans` does not,
  which is exactly the split needed — **the PLAN is the fleet-wide contract, the NOTES
  are the per-site verdicts.** No column has to be added.
- **A hazard that constrains every gate here, filed to `LANDMINES.md`:** an unknown
  check type is **skipped, not failed**, and an all-skipped set reads as PASS + a 7-day
  cooldown. For a ladder that is corrosive, since stage N's pass licenses N+1 — so a
  gate that cannot evaluate its question must be **inconclusive** (PLAN D3). Live now:
  `has_visible_area` (TL-034) is committed and not rolled, so the most useful new check
  type is currently the one that would silently skip.
- **`features_open/015` deliberately NOT adopted.** Accepted decomposition (PROPOSED,
  owner's call): **015 = rung vocabulary · 027 = gate mechanism · 026 = missing
  instrument.** Composable, not merged — so this lane proceeds without owning 015.

## The owner's framing (2026-07-30)

> *"if we are to build more and more complicated components we need to do it step by
> step and follow the doc traveller idea for each small part of a site build … We could
> for instance have a set of build tools and acceptance checks — perhaps a bit like the
> checkers that we have now but are responsible for checking a particular stage in
> development — some may even be created dynamically at the start or at different
> stages of the project."*

## Why this is its own feature

The `teaser-reveal-panel` build was careful: hazards named in advance, a 24-check render
harness run before any DB write, every check proven non-vacuous by mutation, every
change verified against the served page rather than a `complete` status. It still
shipped a component whose JavaScript **never ran client-side at all**, from the first
commit, for four rounds, until the owner clicked it.

The cause was not weak checks. Every check was sound about what it measured. **They all
measured static markup or forced DOM state; not one ever fired a real click.** What was
missing was not rigour — it was a *stage*.

There is no such thing as a build stage for a site part today: a component either exists
or it doesn't, a page is either deployed or it isn't, and everything between (shape
right · contract sound · template renders · registered · placed durably · serves ·
**operates** · still operates after a roll) lives in a session's head.

## The shape proposed

One PLAN + one NOTES stream per small part, in `doc_plans`/`doc_notes` — the same
travelling-docs machinery TL-017 already runs **for tools and not for components** —
with the PLAN carrying a ```criteria fence, and an eight-stage ladder (S0 shape → S7
regress) where each stage has one question and one gate that can go red.

**The key reuse finding:** `interaction` + `text_matches` in `browser-runner-adapter`
already does what the hand-rolled real-click test did, and was proven end to end on
2026-07-29 by the `smart-contrast` pilot (11/11 checks, real Chromium, two profiles,
asserting arithmetic against known answers). The missing stage is **not new
construction** — it is pointing a proven mechanism at components instead of only tools.

## Next actions (P1 — the lane's own plan)

~~1. Run the one query this proposal could not.~~ **DONE — see above.**
~~2. Create the standing five.~~ **DONE 2026-07-30.**

1. **Take the `subject_type='component'` migration through the council gate**, then
   apply. DDL is staged in the lane's RUNBOOK, **deliberately not numbered in
   `sql_for_agents/`** — the runner takes every pending file in a directory, so an
   unreviewed `272_*.sql` could be swept in by an unrelated session's `--apply`.
2. **Give `teaser-reveal-panel` a PLAN + criteria fence**, NOTES backfilled from
   `brochure_component_library/NOTES_…`. Chosen because its five-round history is fully
   written down, so nothing has to be reconstructed.
3. **Make S6 real** — dispatch a component's fence to `browser-runner-adapter` as
   `tool-acceptance-agent` does. Trusted only once a deliberately broken component makes
   it go red.

**Still open, and genuinely the owner's to decide** (unchanged from filing): **who fires
the stages** — G5, discovery passes are manual-fire and the improvement loop is ruled
stopped, so a ladder with no trigger is inert; and **whether a gate may refuse** — a
blocking gate is a guarantee change under the 2026-07-29 ruling and goes to architecture
review, a reporting gate is additive.

## Inputs to read first (don't re-derive)

- The PROPOSAL above — Part 1 is the evidenced carousel provenance, Part 2 the design.
- `webdesign_tools_repair/REPORT_2026-07-29_concepts_for_a_working_tools_chain.md` —
  the tools chain with its five wiring gaps (G1–G5) measured against the live system.
- `travelling_docs/OVERVIEW_self_verifying_tools.md` + `RUNBOOK_travelling_docs(38).md`
  §0 (the Stage 0–6 tracker) — the existing step-by-step build documentation.
- `brochure_component_library/components/README.md` — the per-component acceptance
  checklist this generalises.
- `bugs_open/149` — the cautionary evidence on proliferating checkers: 22 discovery
  handler agents, only 2 running `validate_page_content`, six checks in no agent at all.

## Do not build two of these

- **`features_open/026`** (render the page and check it before it ships) — its Phase 3,
  `browser-runner-adapter` on the deploy path, is a sibling of this feature's S6 stage.
  026 is page-and-palette scoped; this is part-and-interaction scoped. **The dispatch
  should be shared.**
- **`features_open/015`** (staged site maturity ladder) — the same idea one altitude up:
  sites climbing named rungs with per-rung promotion criteria. Also REQUESTED and
  undesigned. These are probably **one design with two altitudes**; decide that early.
- **`bugs_open/151`** candidate (3), a post-build fact-repetition census, is already
  stage-gate shaped and is the only 151 candidate that protects the nine deployed sites.
  It may be cheapest to build as this ladder's first content-stage gate.

## Cross-links

`features_open/015`, `017`, `026`; `bugs_open/149`, `151`; register TL-008, TL-012,
TL-016, TL-017, TL-033, DOC-003, DOC-010, CLC-012;
`docs024_key_docs_latest/brochure_component_library/` (the lane that produced the evidence).
