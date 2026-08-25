# HANDOFF 2026-08-25 — continue here

**Lane:** `copy_quality_two_stage` (stage 2 = the `copy-editor` editorial pass).
**Supersedes `HANDOFF_2026-08-23_continue_here.md`.**

> ## ▶ START HERE, IN THIS ORDER
> 1. **`SUMMARY_2026-08-23_the_editor_ships_and_the_brief_defect_is_closed.md`** — 5 minutes, plain
>    prose. Still current on everything except the wiring, which landed after it.
> 2. **This file's "Next work"**. Everything above it is context.
>
> **One-line state:** stage 2 proposes, its approved edits ship to live pages, and it is now
> **DISPATCHABLE** — the routing went live on `v1.0.1337` today. It has its **first user outside this
> lane**. Nothing is in flight from me. **Two things need a human, and one is a defect in my own gate.**

## ⚠ ADDENDUM 2026-08-25 (review pass, same day, later) — read BEFORE "Next work"

**0. (added late 08-25) A SECOND owner escalation arrived the same day, WITH instructions, and the
first two are DONE.** From his homegarden.uk review (canonical:
`loanzy_uk_example_site/OWNER_REVIEW_2026-08-25_homegarden_and_what_it_says_about_every_site.md`):
the machinery must "up their game a lot"; refresh context BEFORE proposing fixes; audit EVERY prompt
in DB and code against "is it encouraging AI styles of writing". The refresh is done —
**`REFRESH_2026-08-25_deep_context_the_accumulated_copy_discussion.md`** is now the lane's
assembled context and the first read for any new session, ahead of the summary below. The audit is
scoped and censused — **`PLAN_2026-08-25_prompt_audit.md`** — and **phase 1 is DONE**
(`AUDIT_prompts/PHASE1_2026-08-25_findings.md`): the rendered writer prompt demonstrates ~64
negation constructions + "plainly" ×14 per call, and **the about-page premise was INSTRUCTED** —
migration 223's remedy clause in the writer template ("say what we DO … we name our sources and
their dates … say plainly that we can still be wrong"), textual match measured, causation
`[INFERRED]` with its test named. **Phase 2 verdicts 1 of N done** (`AUDIT_prompts/PHASE2_2026-08-25_verdicts_writer_template_and_house_voice.md`):
writer template TEACHES-AI via three "say this instead" clauses (method = sourcing narration;
"we cannot tell you X" as the substitute for the banned word "honest"; values/approach filler for
empty testimonial slots); house voice NEUTRAL in content, TEACHES-AI in form (17 demonstrations of
the construction it bans — OWNER DECISION, his approved text). Fix shapes recorded, nothing applied.
**(i) The causal test RAN 2026-08-25** (`AUDIT_prompts/EXPERIMENT_2026-08-25_about_section_replay.md`,
pre-registered, 12 offline replays): Finding 2 PART-REFUTED — the writer clauses are secondary
(10→6); **the PLANNER'S page title ("…Editorial Approach and What We Will Not Do") is the primary
premise carrier** (→1, heading gone 3/3); the drafted replacement instruction earned nothing over
deletion; register tells unchanged in every arm (separate fault, fed by demonstrations).
**(late 08-25) THE OWNER RULED ON ALL SIX** (`OWNER_RULINGS_2026-08-25_six_decisions_on_the_copy_machinery.md`
— his words + execution record). **SHIPPED same day: migrations 627/628/629** (writer substitutes
deleted with every ban + tool mandate kept; house voice form-rewrite 17→0 demos; planner stops
demonstrating and planning unfillable social-proof slots), applied + backed up + rollbacks +
recorded + verified at the live rows; council `Council-Submitted: 6a0f8b99`. **Next executable
steps: (a) the planner PREMISE fix candidate awaits his yes (rulings doc, ruling-1 research —
mission prohibitions become page premises despite the existing guard; ~6 of 23 about pages
affected); (b) ruling 4 scoping — stakes-split of the uncertainty device over the BRIEFS
(12–31 demos/site, the biggest untouched layer) + propagate the best-in-class Build standard
beyond the classifier (0 of 51 specs carry it; research wiring is a never-built TODO);
(c) canary: the next greenfield build's about page + re-run CQ-032 on a fresh rendered prompt;
(d) phase 2 verdicts 2 of N: briefs → llm_guidance → copy-editor; (e) delete the writer's dead
rules 16–17 in a follow-up migration.** His sharpest new datum: the
about.html PREMISE is wrong (14/17 methodology headings), his phrase list is a SAMPLE, and fixes
that remove named strings while keeping the page's shape will not satisfy it.

**1. This handoff MISSED AN OWNER ESCALATION that landed nine minutes before it was committed**
(`e0da73a1b` at 10:54:36 vs this file's `28965069a` at 11:03:53 — checked in git, both today).
`CONTRIB_2026-08-25_OWNER_ESCALATION_finetuning_pages_fail_the_would_a_person_say_this_test_after_a_maximal_seed.md`
carries an owner verdict on two fully framework-written finetuning.uk pages built AFTER a maximal
seed: they fail his *"would a person actually say this"* test, and his instruction, verbatim, is
that this lane's machinery *"will need to substantially improve"*. **Treat it as item 0 of "Next
work".** It also reframes item 5: the escalation's sharpest finding is that the owner's tell class
is WIDER than any enumerable pattern list (the methodical scaffold, the performed-candour beat) —
their own checklist scored the rejected section CLEAN — so a regex/tell-count instrument is now
demonstrated insufficient as an acceptance test for whatever "substantially improve" becomes.

**2. The fleet-mix caveat has EXPIRED, in the good direction.** All **62** agent-chassis pods are on
`v1.0.1337` `[MEASURED 2026-08-25, this pass]`. Only the new router is live; the "a 1336 pod still
files `tone_shift`" branch is gone.

**3. `tone` findings are live traffic, not a theoretical trickle.** One fired 2026-08-24 10:13Z —
item `7f73a84c`, filed as `tone_shift` (old router, pre-roll), resolved `wont_fix` 21 minutes later.
"No `needs_copy_edit` row yet" still holds (0 rows, re-checked this pass), but the first dispatched
run (item 2) could land any day, not eventually.

**4. Register CQ-030's index row was stale** — it still said the router half was "inert until a
roll" after the roll. Corrected visibly this pass (both `content-quality.md` and the index; council
seats read register status as ground truth).

**5. Two traps for whoever fixes item 1**, found reading the gate this pass:
- `load_page_text()` returns `content_data::text`, i.e. **JSON-encoded** text — an href appears
  there as `href=\"…\"`, not `href="…"`. The volume arm already compensates (line 459 searches the
  escaped form); a page-scope declared-set check that searches the plain form against `page_text`
  reads every declared link as absent, and the noise this fix exists to remove comes straight back.
- `load_page_text()` **excludes the edited component entirely**, so "page scope after the edit" =
  other components' text + **this component's unedited fields (post-edit `before_data` minus the
  proposal's keys)** + the edited fields' `after` values. Miss the middle term and a required link
  sitting in an unedited field of the edited component fails by construction — the same defect one
  scope out. (Residual edge, note only: item-mode grades each edit against its own component, so an
  edit that MOVES a declared link between two components of one proposal fails on the losing side.)

Everything else in this file re-verified clean this pass: both proposals still parked at
`needs_human_review` (unchanged since 08-24), migration/register/council ids all resolve, the
CONTRIB telling finetuning their proposal fails the gate exists in their directory, and the item-1
defect is real and exactly as described (`gate_stage2_edit.py` — the `required_links` check sits
inside the per-field loop, lines 412–416 under the loop at 369).

## What is true as of 2026-08-25

- **The wiring is LIVE** (register **CQ-030**, council **APPROVED** `c1931fa1-5a98-4874-9730-b9ef3519c0d4`).
  Audit `tone` findings file **`needs_copy_edit`** at **`copy-editor`** instead of `tone_shift` at
  `page-build-handler`, which regenerates the page.
  - Half 1: migration **579**, applied 08-24 — `copy-editor`'s dispatched entry path.
  - Half 2: the router, live on `v1.0.1337` today.
  - **Verified at the binary with a full control set** on a persistent pod: `content_rewrite` = 3
    (probe works), `needs_copy_edit` = **2** (change is in), `tone_shift` = **0** (a genuine
    REMOVED-string control — the strongest kind, because my change deleted that literal).
- ~~⚠ **The fleet is MIXED: 56 pods on `v1.0.1336`, 40 on `v1.0.1337`.** Both routers are live at once,
  so a `tone` finding handled on a 1336 pod still files `tone_shift`. Re-check before concluding
  anything about routing behaviour.~~ **EXPIRED same day — fleet converged on `v1.0.1337` (all 62
  pods) by the afternoon review pass; see the addendum above.**
- **No `needs_copy_edit` row exists yet** — `tone` is genuinely that rare, which was the whole safety
  argument. For contrast, **63 `content_rewrite` items were filed since 08-24**: routing THAT category
  would put ~63/day into a queue with no working surface. **Do not extend the route to it.**
- **`bugs_closed/327`** (the brief defect) is closed and artefact-verified. Owner ruled ship-as-is,
  no gate; the three fragment briefs stay with their own lanes.

## ⚠ TWO THINGS NEED A HUMAN

1. **Two proposals are parked at `needs_human_review`:**
   - `b0dea48e` — ours, `ai-agent-orchestration.com/index`, 3 edits. Cuts the CTA a **second** time
     (496 → 245) and flags a **figure conflict**: 175 "Agent Definitions" vs 170 "Agent Types" across
     sections. Gate-passed when filed; ~~**re-grade before acting** (that page rebuilds daily)~~
     **RE-GRADED 2026-08-25 against the rebuilt page: edits 1–2 now FAIL** (both would WRITE figures
     `170`/`14` the current page no longer carries — the proposal standardised the conflict against
     the 08-24 page, and the page has moved; edit 2 also drops two prose URLs now unique to its
     field). **Edit 3 (the CTA cut) still PASSES. Approve edit 3 at most; edits 1–2 need
     re-proposing.** All three stored component ids were dead again (fourth dangling-id occurrence);
     the gate re-resolved by slot. Evidence: NOTES 2026-08-25 (later).
   - `8003c51a` — **`finetuning_uk_service`'s**, `finetuning.uk/your-own-model`, 2 edits. **FAILS the
     gate.** Told them (CONTRIB of 2026-08-25 in their directory). Not ours to approve.
2. **My gate has a defect, found by grading their proposal** — see "next work" item 1.

## Next work, in the order that closes doors

1. ✅ **DONE 2026-08-25 (same day, review-pass session).** ~~FIX THE GATE'S DECLARED-LINK CHECK — it
   runs PER FIELD.~~ `declared_link_verdicts()` now grades the declared set ONCE per grade at PAGE
   scope after the edit (merged component, unedited fields included, plus every other component).
   Four verdicts: kept / added (ok) / **REMOVED (fails)** / gap (pre-existing page-level absence —
   ⚠ reported, never failed). The FAIL arm is control-proven direct-logic (the prose-URL pattern —
   a proposal mutation cannot delete a link from another component's row) and **mutation-proven**:
   regress the helper to read `others_json` un-escaped and the self-test exits CONTROL FAILED.
   On `8003c51a` the three noise lines are gone, edit 1 is credited with ADDING `/contact.html`,
   and the real structural failures stand alone. Evidence: NOTES 2026-08-25 (later).
2. **The first DISPATCHED run has not happened.** When a `tone` finding appears, verify it end to end:
   a `needs_copy_edit` row at `copy-editor`, then a run that reaches `run_copy_edit` via the
   **dispatched** branch (`load_work_item`, not `echo_page_ref`).
   `SELECT collected_data->'page_ref'->>'page_id' FROM orchestration_states WHERE correlation_id='…'`.
3. ✅ **BOUNDED 2026-08-25 (review-pass session).** ~~Convergence on a diffuse page is UNTESTED…
   Bound it (one run per page per period) before anyone widens the route.~~ Built as a **structural**
   bound rather than the sketched clock (a period leaves N un-reviewed proposals representable;
   pending-review does not): `pendingCopyEditForPage` in `write_audit_findings_action.go` withholds a
   new `needs_copy_edit` while the page has an open one (any producer) or an un-reviewed
   `copy_edit_proposed`, and reports `items_skipped_pending_proposal`. Drains when the human acts —
   D2's rate limiter. Council **APPROVED** `754dcffd` (round 2, 4 advisories all checked — NOTES
   2026-08-25 evening; round 1 REVISE corrected a false premise of mine: the anti-churn brake
   already rate-limits same-key re-files, so the bound's true gap is cross-type + cross-source
   only). Wiring mutation-proven; **Go, so inert until the next chassis roll** — verify at the
   binary then: probe literal `items_copy_edit_bound_unevaluated`, present-control `content_rewrite`.
   Convergence itself (does repeated proposing ever go quiet?) remains unmeasured; the bound makes it
   safe to find out slowly.
4. **The narrow sibling** — three lanes have asked (`277`, `301`/`083`, `323`; 999 + 160 + 98 items as
   of 2026-08-20, live **and** archive). Specced in
   `DESIGN_2026-08-20_the_narrow_sibling_one_component_one_defect.md`. Not this lane's to build.
5. **The form-versus-phrase question is still the honest limit on everything this lane claims.**
   Phrase transfer is proven (one tagline, 1,369 prompts → 409 responses). Whether FORM transfers is
   untested. ⚠ **Two preconditions now, both from `apis_uk_bees_homepage`:** control for whether the
   section plan assigns per-section subjects (or you measure the plan's emptiness instead), and note
   that **a concrete, on-topic exemplar is LIFTED AS CONTENT rather than imitated** — so transfer is a
   function of the exemplar's character, not a rate. That corrected a figure of mine that had already
   fed an owner decision (`WRONG_CALLS` 2026-08-24).

## The apply path — two scripts, and the traps are in their headers

- **`scripts/fire-copy-editor.sh <domain> <page>`** — fire one stage-2 run (proposal side).
- **`scripts/fire-section-edit.sh <work_item_id>`** — ship ONE approved edit. Sequential.

⚠ **`client_id` is interpolated UNQUOTED as a SCHEMA NAME** (`spawn_actions.go:2315`,
`INSERT INTO client_%s.agent_instances`). A hyphenated one you invent for tracing dies as
`syntax error at or near "-"` — **and reads like a platform fault**. Use `demo_client` / `system`.
`section-editor` spawns its deployer **second**, so the run dies *before* the edit is attempted,
leaving the item claimed and the page untouched.
⚠ **Resolve `page_component_id` by `(page_id, slot_name)` at DISPATCH time.** Pages rebuild on a
schedule and a rebuild REPLACES the row: ids filed on 08-21 were dead by 08-22. Three occurrences.
⚠ **`complete` is not proof** — `check_edit_skipped` routes a gated REFUSAL to `complete` too. Verify
`content_data` changed.

## Standing cautions (fresh first)

- **Filter by YOUR correlation, never "the most recent row".** Logged 08-21, repeated 08-23 with three
  edits in flight. The check now lives in the script.
- **An empty result from a wrong path looks exactly like a real absence.** Three instances now:
  `max_tokens` at the wrong config path; `grep -c '^--- FAIL'` returning 0 on a **build** failure;
  `site_work_items` alone reporting 29 of 999 because the archive is bigger than the live table.
  **Demand a positive control from the same query.**
- **Probe the binary by IMAGE, never `-l app=agent-chassis`** (2 pods of ~96), one literal at a time —
  and use a **long-lived** pod: a dynamic-agent pod was reaped mid-`grep` (exit 137). Dispatch
  readiness is the **deployment** rollout, not pod churn, which never quiesces.
- **`llm_call_log` lags the orchestration by minutes.** An empty result straight after a run reads
  like an instrumentation outage.

## The five living docs

- **PLAN** — untouched. **NOTES** — evidence log, 08-25 tail. **README_where_we_are** — the owner's
  plain-prose log. **SUMMARY series** — 08-12 · 08-14 · 08-15 · 08-17 · 08-19 · **08-23 (newest)**.
  A new summary is **not** due yet: the five headings would repeat 08-23's until the first dispatched
  run lands. **this HANDOFF.**

**Tooling:** `gate_stage2_edit.py` (⚠ item 1) · `audit_writer_brief.py` · `count_negation_tells.py` ·
the two `scripts/fire-*.sh`. **Platform code owned:** `content_direction` derivation in
`site_spec_actions.go`, `datahelpers/format_content_direction.go`, the `tone` arm of
`write_audit_findings_action.go`, and their tests (all mutation-proven).
**Migrations:** `447`, `462`, `579` (all applied; each has a `_ROLLBACK`).
