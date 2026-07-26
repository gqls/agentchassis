# PLAN — durable-write completeness guard (bugs_open/021 INSTANCE 1)

**Started:** 2026-07-21. Branch `085_debug_and_feature_loops`.
**Bug:** `bugs_open/021_HANDOFF_2026-07-18_durable_write_guard_covers_one_path_only.md`
(INSTANCE 1 only — the write guard. INSTANCE 2, verifier coverage, is OWNED by
the `work_item_completion_integrity` thread; not touched here.)

## The originating brief, and its correction

The handoff asked to extend the `componentRegressionIssues` guard (which protects
`content_components.html_template` on `update_component_html`) to two more
surfaces: `page_components.rendered_html` and `pages.rendered_header/footer/head`.

> **CORRECTION 2026-07-21 — the brief's two named surfaces are non-exposures, and
> the guard shape it asks for (comparative) is the wrong shape for the real hole.**
> Kept here, not silently edited away, because the brief's *generic mechanism* is
> still right and its instinct (raised by the council bug_historian seat) still
> pointed at a real live defect — just not at the surfaces it named.

Evidence for the correction (all 2026-07-21, live):
- `pages.rendered_*`: **no Go writer anywhere** (exhaustive grep) and **0/301
  pages** carry any value. Dormant columns. Chrome lives in
  `site_components.rendered_html`, assembled deterministically. A guard here
  protects a write that never happens.
- `page_components.rendered_html`: a **derived render** of the guarded+versioned
  `content_components.html_template`. Live check: for every tool, `rendered_html`
  length ≈ `html_template` length (98–122%). The platform's own code says so
  (`update_component_html_action.go:249-253`). A truncated `rendered_html` is
  recovered by re-render. An agent census of all 19 writers (both tables) found
  **none** takes a whole LLM completion into it as an overwrite — the writers are
  deterministic renders (guard would fire on legit output) or deterministic CSS
  transforms (guard would block legit fixes).

## The real defect (the mechanism, correctly located)

The 012 shape — *a whole LLM artifact persisted with no completeness check* —
survives at the tool **BIRTH** write, not at any overwrite:

- `create_tool_component_action.go:105` gates birth on `HasToolDocHeader` only,
  which checks the doc-header sentinels at the TOP of the `<script>` and **cannot
  see a tail cut**. The whole `html_content` is then written to
  `content_components.html_template` + `page_components.rendered_html`.
- **Live proof:** `bugs_open/046` — 8 tools + 1 section serve unterminated
  JavaScript on 6 live domains right now, 7 of 8 with no intact prior version =
  born broken, straight through this gate.
- The correct absolute check already exists and is live: `toolTemplateValid`
  (`plan_sections_action.go:1046`, reuses `balancedPairs`+`endsCleanly`,
  calibrated on all 27 tools, 0 FP) — but it is wired only into the schema-LOAD
  path, not the birth WRITE.

## Design

**Principle: same predicate at both seams.** A tool that `componentTemplateValid`
would drop at load must not be *born*. So gate the birth write with the identical
function.

**Phase 1 — tool birth gate (the fix).**
In `create_tool_component_action.go`, after the `HasToolDocHeader` gate and before
the `content_components` INSERT, add:
```go
if !componentTemplateValid(htmlContent, "tool") { // == toolTemplateValid
    // hard fail: work item -> needs_human_review via the workflow error_step,
    // reason carried; record to agent_error_log like recordComponentWriteRejection.
    return nil, fmt.Errorf("refusing to persist truncated tool %q: <script>/<style>/… unbalanced or ends mid-token — the generation was cut", componentName)
}
```
- Safe: `create_tool_component` only ever handles tools; `toolTemplateValid` is
  calibrated 0-FP on the tool population (the 46-plan's 36-FP warning is about the
  *fleet-wide* population, not tools).
- Reuses the live, proven predicate — no new thresholds, no new simulation.
- Record the refusal (reuse `recordComponentWriteRejection` /
  `agent_error_log`, error_code e.g. `tool_birth_truncation_blocked`).
- **DEPLOYMENT ORDER caveat inherited:** `create_tool_component` already documents
  that the tool-generator prompt (teaching the doc header) must ship before/with
  the gate binary. Same ordering holds — but this gate is purely additive to an
  already-present `HasToolDocHeader` fail, so it introduces no *new* ordering risk.

**Phase 2 (owner's call) — section birth gate.**
`store_generated_component`'s `scoreComponent().TemplateClosed` checks `<section>`
balance only. Strengthen with the **5-pair tag-imbalance** predicate (tag-balance
ALONE — endsCleanly over-fires by 36 fleet-wide on non-tool components). Smaller,
separate, evidence is one `section` casualty (unplaced), so lower priority.

## Calibration rules carried from the brief + 046 (do not re-derive)
1. Truncation cannot grow an artifact (comparative guards only — N/A at birth,
   where there is no prior row, so birth uses ABSOLUTE structural checks).
2. Tag-imbalance (`<script>/<style>/<section>/<div>/<fieldset>` open>close) is the
   0-over-fire fleet-wide signal. `ends-mid-token` is tool-safe but +36 FP
   fleet-wide — never apply it to a non-tool population.
3. A guard that refuses good work gets switched off. The tool gate reuses a
   predicate already proven against the live tool population.

## Non-goals / coordination
- NOT extending the comparative guard to rendered_html/pages.rendered_* (refuted).
- NOT sweeping the existing 9 casualties — that is `bugs_open/046`
  (`truncation_casualties_046`), which explicitly routed the write guard here.
- NOT touching the re-render *delivery* path — that is `bugs_open/024`
  (travelling-docs, actively worked).

## Verify after coding (Phase 1)
1. Unit-test `componentTemplateValid("<truncated tool>", "tool") == false` and a
   whole tool `== true`, against a 046 census signature + a healthy 27-tool sample.
2. Fault-injection (per the "verify the failing branch" rule): drive
   `create_tool_component` with a tail-cut tool HTML (valid header, unbalanced
   `<script>`) → component NOT created, work item `needs_human_review`, refusal in
   `agent_error_log`. A healthy generation must still be created.
3. Pod-grep a literal the change CREATED (e.g. the new error_code), plus a
   positive control — never a string it merely uses.

## Status
- Investigation COMPLETE, evidence-grounded, docs written.
- Owner scope decision 2026-07-21: **Phase 1 + Phase 2.**
- **CODE COMMITTED `ba702c8c6`** (component_write_guard.go, create_tool_component_action.go,
  store_generated_component_action.go, + test). Build + tests green. `bugs_open/021`
  updated with the corrected scope + fix.
- **NOW LIVE in v1.0.1146** — rode the owner's sweep build `fe2ba5e52` (a descendant
  of my code commit). Pod-verified: all 3 created literals present + positive
  control.
- **BEHAVIOURALLY VERIFIED 2026-07-23 on v1.0.1149** (fault-injection via a scratch
  one-step `create_tool_component` agent; LLM bypassed). Tail-cut tool → gate
  fired, `tool_birth_truncation_blocked` logged, 0 created; healthy tool → gate
  passed (failed later at site-load), 0 created. Scratch fixtures removed, leak
  check 0. Evidence in NOTES (2026-07-23). **INSTANCE 1 is DONE** — the 021 file
  stays in /bugs_open/ only for INSTANCE 2 (not ours).
- NOTE (NOTES §misstep): the planned `toolTemplateValid` refactor in
  `plan_sections_action.go` was DROPPED to avoid a same-file passenger (another
  session's 041/044 WIP in that file); cosmetic only, no functional loss.
- ~~NEXT: owner-sanctioned build + roll, then fault-injection verify; optional
  advisory council gate.~~ **ALL THREE DONE** — rolled (v1.0.1146→1159),
  fault-injected 07-23 and again 07-25, council APPROVED r1 on 07-25. Struck
  rather than deleted so the plan still shows what it was waiting on.

---

## CLOSED 2026-07-25 — scope grew to INSTANCE 2, and the whole bug is now shut

> **CORRECTION to this plan's own framing, line 5.** It says INSTANCE 2 "is OWNED
> by the `work_item_completion_integrity` thread; not touched here." That held for
> four days and then stopped being true: that workstream went dormant on 07-20
> leaving a documented next step, and on 07-24 and 07-25 this lane did the work —
> **into their docs, never forked.** Ownership did not change; the labour did.
> Recording it because a plan that still claims "not touched here" would send the
> next reader to the wrong file for half the outcome.

**Decisions taken at closure, and why:**

1. **Closed on behavioural evidence, not on the council verdict.** The gate is
   advisory. Both branches were induced on the live binary (v1.0.1159); that is
   the bar. The verdict (APPROVED r1) landed 10 minutes later and changed nothing
   — had it come back REVISE, the objections would have become a follow-up commit,
   not a reopening.
2. **No `Council-Reviewed:` trailer, deliberately.** `34adb171c` predates its
   verdict by a day and forward-only forbids an amend, so the pair is a permanent
   `098` false negative. Written into the bug file so the coverage gap reads as
   explained rather than careless. *Generalises:* submit before or alongside the
   commit, never after — a verdict that post-dates its commit can never be joined
   to it.
3. **Verifier coverage stays at 3 of 77, by design.** The alternative — write
   verifiers to raise the number — is the exact regression the held `page_rerender`
   verifier demonstrated (six passing tests, wrong contract, would have stranded
   1,849 items). The coverage map is the build-enforced backlog and names its own
   next candidate. **No separate feature file**: a second list would drift from the
   one the build enforces.
4. **The residue was split out, not buried.** `bugs_open/077` — the detector's
   predicate is wider than its handler's remit. Filed on a *corrected* premise:
   the 07-24 note called it churn, measurement showed `idx_swi_dedup` bounds it to
   one open item per site. A design call, explicitly left with the check's owner
   rather than picked here.

**What this plan got right, and what it got wrong.** Right: the instinct to verify
the brief's named surfaces against live data before building anything — that
correction (the two named surfaces are non-exposures) is the single most valuable
thing in the file, and it saved building a comparative guard at a seam where it
would have blocked legitimate work. Wrong: the phrase *"INSTANCE 2 … not touched
here"* aged into a false statement within days, and the "Verify after coding"
section stopped at Phase 1 — Phase 2's and INSTANCE 2's verification steps were
never planned here, so when they came due on 07-25 they had to be designed on the
spot. **A plan's verification section should cover every phase it authorises**,
not just the one being built that day.

**Full closing record:** `bugs_closed/021_HANDOFF_2026-07-18_…md`, the 2026-07-25
NOTES entry (including a collected MISSTEPS section), and
`SUMMARY_2026-07-26_durable_write_guard.md`.
