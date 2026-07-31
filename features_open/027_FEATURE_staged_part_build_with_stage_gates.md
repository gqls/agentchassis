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

---

## Progress — 2026-07-31: the first machinery built, and it is not a stage gate

**Scope reminder:** the ladder was cut from eight gates to three by owner ruling (D8, lane
PLAN). Machinery gets built for the claim-before-the-build (S1), verification through the
visitor's real gesture (S6), and **every check proven able to fail** (the S2 *discipline*, not
the S2 harness). The other five stages are unfunded, not disproved.

**What now exists** (`docs024_key_docs_latest/staged_component_build/scripts/`, register
**TL-036**):

- **`try_fence.go`** — run a candidate criteria fence against a live URL and see every verdict,
  offline, before it is published. It separates profile-gated skips (legitimate) from
  type-not-implemented skips (a defect, because upstream an unknown type is *skipped, not
  failed*, and an all-skipped set reads as PASS plus a 7-day cooldown — this feature's D3), and
  it asserts its own arithmetic rather than trusting its own report.
- **`prove_fence_can_fail.go`** — the S2 discipline discharged: a green local baseline, then one
  mutation at a time, and **exit 1 if any check has no mutant at all**. A green with a hole in
  it is unreachable.

Both call `RunChecksAction.Execute` from `internal/adapters/browserrunner` — the fleet's own
evaluator, not a second implementation of the same switch.

**Why this and not a stage gate first.** P1a (the three-way naming contract) outranked
everything because a mismatch makes a fired run *skip and read clean*. That check is built and
run. Its one remaining BROKEN-B case — a tool with a PLAN and no fence, so its acceptance run
started and asserted nothing — is now **closed**: `tool-review-council-simulator` has an
18-check fence live in `doc_plans`, verified 36/36 on the live page across both profiles, with
17 of 17 mutants caught and all 18 checks watched red.

**The finding worth carrying into the remaining stages.** The mutation prover **refuted one of
the checks it was written to validate, on its first run** — a check named for a guarantee it
could not test (`threshold-lever-updates-live`, which passed the mutant that killed the
slider's `input` listener, because Playwright's `fill()` also fires `change`). Generalised:
**a check's ID is an assertion, and it is the one part of a check that nothing validates.** The
evaluator will faithfully run the wrong test under the right name for ever. So the S2 rule
should be stated as *mutate against the NAME*, not against the code — every one of this lane's
four self-inflicted instances of the class had correct code and a lying name.

**Still open on this feature**, in order:

1. **S6 for components** — dispatch a component's fence to `browser-runner-adapter` as
   `tool-acceptance-agent` does for tools. Wiring, not construction; the harnesses above now
   let the fence be written and proven before the dispatch exists.
2. **The Go gate for `subject_type='component'` has never been exercised** (register DOC-068).
   The DB half is live (migration 273 applied 07-31); the Go half is in the binary by build date
   only, and `doc_plans` holds 0 component rows. Needs a dispatch, not a query —
   `load_doc_context` takes `subject_type` from **step config** (`load_doc_context_action.go:37-43`).
3. **`tool-arena-interface`** — the last BROKEN-A case: a tool component with no page under any
   name. Not a rename; a decision about whether it should exist. Needs a human.
4. **`has_visible_area` checks are owed to every fence** once `bugs_open/157` closes. The type
   is live in the running binary but reports 0 for whole-number axes, so it currently accuses
   correct elements of being invisible — deliberately omitted from the first fence, and recorded
   there as an omission rather than left to look like an oversight.

### Correction and completion, same day (2026-07-31, later)

Two items in the list above are now resolved differently from how they were written.

**S6 is PROVEN for a tool, end to end, in the cluster** — not just offline. The fence was
dispatched via `tool_acceptance_run.sh` and came back `complete` in **18s: 22 passed / 0
failed / 14 intentional profile-gated skips**, with every skip verified as
`not run on profile mobile` and none as `not implemented`. That is the stage whose absence
cost `teaser-reveal-panel` five rounds, now firing and asserting on a real tool.

**But the FIRST dispatch failed, and the reason is a constraint this feature must carry:**
`runDeadline` is **120 seconds for the whole request** — every url x every profile. An
oversized fence surfaces as `browser open failed … context deadline exceeded`, which reads as
infrastructure. Measured: **36 evaluations = 10.6s locally (x3) but FAILED at 133s
in-cluster**; ~3-5s per evaluation there against ~0.3s locally. Gating the 14
profile-independent checks to desktop gave 22 evaluations, 18s, and lost no assertion.
**So a stage gate must be sized for the pod, and the S2 rule needs a companion: an offline
harness proves a check is CORRECT, never that it FITS.**

**Item 3 above is WITHDRAWN as written.** `tool-arena-interface` is **not** an orphan, and the
check that said so is the reason it was believed. It is live, deployed and serving on vonc.com
under a page named `tool-arena` (`/tools/arena/index.html`, `build_status=deployed`,
redeployed 2026-07-31 12:45); the component's own markup is present in the served page. The
check had concluded "no page at all" from "no page under the two names I guessed" — and its URL
guess assumed a `<name>.html` filename convention that this site does not use. **The decision
in front of a human is therefore not "should this component exist" but "which of the two names
should move"** — the page-rename side being the safer one, since `function` is the naming
contract that `page_components.slot_name`, cross-links and `RekeyTravellingDocs`
(`features_open/028`) all key on. `CHECK_naming_contract.sh` now asks placement before
concluding absence, and prints that remedy.

**The transferable rule for every future gate in this ladder**, since this is now the second
distinct way a gate has misled us in two days: a gate must not be able to state a conclusion
wider than its measurement. "No page under the names I tried" is not "no page"; "passes on my
machine" is not "passes"; "a PLAN row exists" is not "a fence exists". All three were shipped
by this lane, and all three were caught by running the thing rather than reading it.

### P1a CLOSED, 2026-07-31 evening — and the naming check's value is now demonstrated, not argued

`CHECK_naming_contract.sh` returns **PASS**: BROKEN A **0**, BROKEN B **0**, across 30
canonical tool components (12 testable now / 10 authoring backlog / 8 neither; reconciled).
That is the first pass since it was written, and it retires the item that "jumps the queue".

The last case was the arena page rename, and it contributed two things beyond closing the item.

**1. A rename's blast radius is every equality join on the renamed value, not just the
consumers you can name.** `pages.name` had a second consumer keying on it:
`check_sectionless_pages` joins `site_plan_pages spp ON spp.name = p.name`. Renaming
`pages.name` alone would have silently removed the page from that detector's population — it
qualified and was actively reported. **A stage gate that fixes addressability must not blind a
different gate**, and nothing in the platform would have announced the trade. Both name-side
rows moved together and the detector's own join was re-run afterwards. Landmine recorded.

**2. The naming defect was MASKING a substantive one, which is the strongest evidence this
feature has produced for its own ordering.** The instant the arena page became resolvable, its
existing fence — generated 2026-07-14, never once executed — **failed honestly**: it asserts
`interaction: fill #take-input` and the served page has no `id` attributes and no form control
at all. The run had always died on page resolution before reaching that assertion.

So the argument for P1a is no longer *"a mismatch makes a run skip and read clean"* as a
prediction. It is measured: **of the two BROKEN cases, one (BROKEN B) was a tool asserting
nothing, and the other (BROKEN A) was a tool whose contract and markup have disagreed for
seventeen days with no way for anyone to find out.** Addressability is not hygiene before the
real gates — it is the precondition that decides whether any gate's verdict means anything.

The remedy for the arena is a design decision belonging to the `gauntlet_dead_cta` lane
(CONTRIB filed in their directory, pointer appended to their cold-start §4). **Deliberately not
fired from here**: a failing verdict files an `improve_tool` item routed to an automated
`tool-improver` against an `owned` page, and letting a one-shot fixer choose between "the fence
is stale" and "the tool is incomplete" is the wrong way to resolve it.
