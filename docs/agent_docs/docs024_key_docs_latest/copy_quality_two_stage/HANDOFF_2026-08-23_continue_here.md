> # ⚠ SUPERSEDED 2026-08-25 — read `HANDOFF_2026-08-25_continue_here.md` first.
> The wiring it lists as "committed, inert until a roll" is now **LIVE on v1.0.1337**, stage 2 has a
> user outside this lane, and a defect has been found in `gate_stage2_edit.py`. Kept for the apply
> record and the council history.

# HANDOFF 2026-08-23 — continue here

**Lane:** `copy_quality_two_stage`. **Supersedes `HANDOFF_2026-08-20_continue_here.md`**, which is
wrong about the two things that matter most: `bugs_open/327` is **closed**, and run 4's proposal is
**applied and live**.

> ## ▶ START HERE, IN THIS ORDER
> 1. **`SUMMARY_2026-08-23_the_editor_ships_and_the_brief_defect_is_closed.md`** — 5 minutes, plain
>    prose, the whole state.
> 2. **This file's "Next work"**. Everything else here is context you can skip until you need it.
>
> **One-line state:** both halves of stage 2 now work end to end — it proposes, and its approved
> edits ship to live pages; the brief defect is fixed, live, artefact-verified and **closed**; the
> writer-side half belongs to another lane; **nothing is in flight and nothing is waiting on the
> owner.**

## What is true as of 2026-08-23

- **`bugs_closed/327`** (slug `a_partial_spec_write_silently_shrinks_the_brief_the_writer_reads` —
  the number is AMBIGUOUS, another lane filed a different 327). Fixed, live on `v1.0.1319`+, and
  **verified on two real writes by two independent producers** — `loanzy.uk` (operator) and `apis.uk`
  (classifier), every key with content reaching the brief.
- **Run 4's three edits are LIVE** on `ai-agent-orchestration.com/index`, owner-approved 08-21,
  applied 08-23: `differentiators-section` 3,286 → 3,028 chars (7 items intact), `call-to-action`
  subheadline 733 → 496, `latest-news` 340 → 255. Confirmed on the served page, not just the DB.
- **The 3-edit budget is self-correcting, not a ceiling** (CQ-024's verify-later, answered): run 4
  chose three components with **zero overlap** against run 3, and reported in `page_judgement` that
  the remaining fault needs a structural merge rather than editing.
- **Owner rulings, 2026-08-21:** ship `327` as-is with no gate (this answers the escalated `compliance`
  HIGH — do not reopen it); the three fragment briefs stay with their own lanes; run 4 approved in full.
- **The writer-side gate is `bugfix_305_negation_gate`'s**, built on our diagnosis with
  `audit_writer_brief.py` as its specification. Item 6 (scheduling the brief detector) is theirs now.

## ⚠ UPDATED 2026-08-24 — THE WIRING IS DONE (both halves), and item 1 below is superseded

**Register CQ-030.** Stage 2 is now dispatchable, and audit `tone` findings route to it.

- **Half 1 — migration `579`, APPLIED AND LIVE.** `copy-editor` gained a dispatched entry path.
  A claimed handler receives only `work_item_id`, while `load_page_target` bound
  `input_data.page_id`, and a single dual-source step is impossible because `QueryDatabaseAction`
  **errors** on a param path resolving to nil. So: branch on entry (truthy on
  `input_data.work_item_id`), converge both paths on `page_ref.page_id`. ⚠ **This changed a LIVE
  agent's entry** — the hand-fired path was re-proven immediately after applying, and must be
  re-proven after any further change to that branch, because both entries now share one
  `load_page_target` binding.
- **Half 2 — the router, committed, `Council-Submitted: c1931fa1-5a98-4874-9730-b9ef3519c0d4`,
  inert until a roll.** `tone` files `needs_copy_edit` at `copy-editor` instead of `tone_shift` at
  `page-build-handler`. The reason is an incident, not a preference: page-build-handler
  REGENERATES, and one `tone_shift` cost finetuning.uk's homepage 11 non-llm URL keys, five empty
  `<img src="">` and six controls (`bugs_open/238`).
- **The safety argument is VOLUME and it is measured:** `tone_shift` **33** lifetime (live+archive)
  as of 2026-08-24 vs **1,893** `content_rewrite`, into a review queue holding **1,079** items with
  no working surface (`bugs_open/033`, still open). ~1/week cannot flood it. **Do not extend this
  to the high-volume categories** — the code comment says so in terms.
- **D2 is untouched:** copy-editor cannot write to a page (447 RAISEs), so this routes an
  auto-PROPOSAL, never an auto-edit.

**What remains before this is finished:**
1. **A roll**, then **the first DISPATCHED run** — none has happened. Verify it behaves as a
   hand-fired one does.
2. ⚠ **Convergence is untested.** Run 5 (2026-08-24, hand-fired) chose 2 of 3 components run 4 had
   just edited. Checked: **not oscillation** — it cut further on restatement that survived, and
   found a new fault class (a 175/170 figure conflict between sections). But repeated runs keep
   proposing on a diffuse page, so **unbounded auto-dispatch there is untested**.
3. **Run 5's proposal `b0dea48e` is PARKED** at `needs_human_review` and needs the owner: it cuts
   the CTA a second time (496 → 245) and flags the figure conflict.

## Next work, in the order that closes doors

1. **Dispatch — the last unbuilt half, and now the only one.** Nothing routes work to `copy-editor`;
   every run has been hand-fired. Wiring `content-quality-auditor` findings to it is the
   `css-patch-agent` shape the PLAN cites. ⚠ **It is also a change of safety posture, not a
   configuration** (owner decision D2): a queue at volume cannot keep one human approval per item.
   The containable shape is **approval gating the DISPATCH rather than the typing** — a human
   releases a batch, the agent proposes, the gate grades, then it applies. A new
   `(item_type, handler)` pair is held for a human canary regardless.
2. **The narrow sibling** — three lanes have asked (`277`, `301`/`083`, `323`; demand 999 + 160 + 98
   items as of 2026-08-20, counted across live **and** archive). Specced in
   `DESIGN_2026-08-20_the_narrow_sibling_one_component_one_defect.md`. **Not this lane's to build**
   unless someone says so; whoever takes it owns it.
3. **The form-versus-phrase question is still open** and is the honest limit on everything this lane
   has claimed. Phrase transfer is proven (one tagline, 1,369 prompts → 409 responses). Whether the
   *form* of a brief shapes output independently is untested, and `305`'s gate will produce the
   corpus. **State the refutation condition before running it** — this lane published one answer of
   that shape and withdrew it within the hour.
4. **`bugs_open/033`** (another lane's) still gates ROUTINE operation; it does not gate one-off runs.

## The two scripts that are now the lane's apply path

- **`scripts/fire-copy-editor.sh <domain> <page>`** — fire one stage-2 run. Guards: deployment
  rollout, endpoint health, and no non-terminal proposal already parked on that page.
- **`scripts/fire-section-edit.sh <work_item_id>`** — ship ONE approved edit. Sequential: several
  edits on one page race on the render and deploy.

Both headers carry the traps. The three that cost real time:

- ⚠ **`client_id` is interpolated UNQUOTED as a SCHEMA NAME** (`spawn_actions.go:2315`,
  `INSERT INTO client_%s.agent_instances`). A hyphenated one you invented for traceability dies as
  `syntax error at or near "-"` (SQLSTATE 42601) — **and it reads like a platform fault**. Use
  `demo_client` / `system`. Worse, `section-editor` spawns its deployer as its **second** step, so
  the run dies **before** the edit is attempted, leaving the item claimed and the page untouched.
- ⚠ **Resolve `page_component_id` by `(page_id, slot_name)` at DISPATCH time, never from the item.**
  That page is re-rendered daily (~14:43) and a rerender REPLACES the row: every id filed on 08-21
  was dead by 08-22. Third time this has bitten the lane. The gate re-resolves too, loudly.
- ⚠ **`complete` is not proof** — `check_edit_skipped` routes a lock- or decision-gated REFUSAL to
  `complete` as well. Check that `content_data` actually changed.

## Standing cautions (fresh first)

- **Watch YOUR correlation, never "the most recent row".** I made this mistake on 08-21, wrote it
  into `WRONG_CALLS`, and **made it again on 08-23** with three edits in flight — the watcher
  reported the previous edit's `COMPLETED` as mine. The check now lives in the script rather than in
  a document. If a watcher's SQL has `ORDER BY … DESC LIMIT 1` and not your identifier, it is
  watching the fleet.
- **An empty result from a wrong path is indistinguishable from a real absence.** `max_tokens` lives
  at `config.ai_service.max_tokens` and the 3-edit budget is **prose in the prompt**, not a
  `max_edits` key — two wrong paths returned empty and I briefly read that as the budget being gone.
  Demand a positive control from the same query.
- **`llm_call_log` lags the orchestration by minutes.** An empty result straight after a run reads
  exactly like an instrumentation outage. Ask whether the table is receiving anything at all.
- **Do not diff two briefs** — they are rendered in sorted order only since `v1.0.1319`, and the
  three fragment sites still carry pre-fix text. Compare label presence and phrase position.

## The five living docs

- **PLAN** — untouched; nothing in the plan changed.
- **NOTES** — evidence log; the 08-23 tail covers the apply and both failure causes.
- **README_where_we_are** — the owner's log; 08-23 entry is the plain-prose version of this.
- **SUMMARY series** — 08-12 · 08-14 · 08-15 · 08-17 · 08-19 · **08-23 (newest — start there)**.
- **this HANDOFF.**

**Tooling:** `gate_stage2_edit.py` (grades one proposal; `--self-test` must fail) ·
`audit_writer_brief.py` (`--self-test`, `--fleet`, `--transfer`) · `count_negation_tells.py` ·
`loanandmortgagecalculator_couk/gate_page_links.py` · the two `scripts/fire-*.sh` above.
**Platform code owned:** the `content_direction` derivation in `site_spec_actions.go`,
`datahelpers/format_content_direction.go`, `site_spec_formatted_from_merged_test.go` (6 tests, all
mutation-proven). **Migrations:** `447`, `462`.
