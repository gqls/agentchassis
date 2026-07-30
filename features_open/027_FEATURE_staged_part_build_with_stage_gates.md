# 027 FEATURE — build a site part step by step, with a travelling doc and a gate per stage

**Raised:** 2026-07-30, owner, to the `fundamentallyai.com 4` session
(`brochure_component_library` lane), after the `teaser-reveal-panel` carousel took five
rounds and the fifth found a bug present since the first.
**Status:** PROPOSED — designed as a document, nothing built.
**Owner wants the work done in a SEPARATE thread.** This entry is the anchor.

**The proposal document is the brief — read it first, do not re-derive it:**
`docs/agent_docs/docs024_key_docs_latest/staged_component_build/PROPOSAL_2026-07-30_step_by_step_build_with_stage_gates.md`

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

## First actions for the separate thread

1. **Run the one query this proposal could not:** does `doc_plans` fit a component's
   needs without schema change? Marked `[UNVERIFIED]` in the proposal and it gates the
   whole design.
2. Create the standing five under `docs024_key_docs_latest/staged_component_build/`
   (the PROPOSAL is already there; PLAN/RUNBOOK/NOTES/README_where_we_are next).
3. Decide the six open questions in the proposal's final section rather than inheriting
   them — especially **who fires the stages** (G5: discovery passes are manual-fire and
   the improvement loop is ruled stopped, so a ladder with no trigger is inert) and
   **whether a gate may refuse** (a blocking gate is a guarantee change → architecture
   review under the 2026-07-29 ruling; a reporting gate is additive).

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
