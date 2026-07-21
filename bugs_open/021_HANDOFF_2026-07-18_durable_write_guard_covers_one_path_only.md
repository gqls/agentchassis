# HANDOFF — the completeness guard covers ONE durable-write path; the same shape is unguarded elsewhere

**Filed:** 2026-07-18, by the thread that built the `bugs_open/012` guard.
**Severity:** latent, destructive-class. Nothing is currently known to be broken —
this is the *mechanism* left generic after the instance was fixed.
**Status:** OPEN. Not an outage. Needs a decision on scope before code.

---

## Why this file exists

`bugs_open/012` was fixed at the path with the proven incident:
`update_component_html` now refuses to overwrite `content_components.html_template`
with a replacement bearing truncation's fingerprints
(`platform/orchestration/actions/component_write_guard.go`, commit `cc7bcc881`).

When that fix went through the council gate
(`SUBMISSION_CORR=e8827490-764a-4c90-b4db-72e358f9be87`, verdict REVISE), the
**bug_historian** seat objected — correctly — that this is the platform's own
recurring shape:

> "This is precisely pattern-instance-7's shape (Go template `missingkey=zero`:
> one call site patched, mechanism left generic) recurring at the level of
> 'whole durable-artifact overwrite with no completeness check'. […] a human
> should confirm there is a follow-up work item to extend the same
> comparative-guard shape […] before the next incident lands there instead."

That seat also noted the platform has *already* paid for this twice — "rendered_html
tools deleted twice independently, the page-assembler visible-content filter
needing two independent fixes". This file is that follow-up, so the objection
does not evaporate with the council run.

## The mechanism, stated generically

**Any path that overwrites a whole durable artifact from LLM output, without
comparing the replacement to the row it replaces, can persist a truncated
fragment and report success.** The guard now closes exactly one such path.

## Specifically still exposed (named by the reviewer)

| Target | Why it matters |
|---|---|
| `page_components.rendered_html` | The field named in the original rendered-artifact incident — raw HTML/JS tools silently deleted by rebuild. Interactive game/tool pages store the WHOLE tool as a section's `rendered_html` (~18 KB), so a truncated write here loses the tool. |
| `pages.rendered_header` / `rendered_footer` / `rendered_head` | Whole-artifact chrome, same overwrite shape. |
| Any future action writing `html_template` / `rendered_html` directly | Does not route through `UpdateComponentHTMLAction`, so it inherits no guard. |

Note the asymmetry that makes this worth doing: `content_components` has
`component_versions` (recovery was possible for 012 *because* of it). Confirm
whether the paths above have an equivalent snapshot before assuming a bad write
is recoverable there.

## Before writing any code — the calibration rule

Do **not** port the three checks blind. They were tuned against real history and
an earlier, obvious-looking check had to be deleted because it refused
legitimate rewrites (016b §9 has the counter-examples). The transferable rules:

1. **Truncation cannot grow an artifact.** Gate structural checks on the
   replacement being no larger than what it replaces.
2. **Every check must be comparative**, never absolute — otherwise it blocks
   repairs on exactly the already-broken artifacts that most need them.
3. **Simulate against real history before shipping.** For components that was
   `component_versions` (29 transitions). Find the equivalent table for the
   target path and run the same simulation; if there is no history table, that
   absence is itself a finding.
4. A guard that refuses good work gets switched off, and then it protects nothing.

## Open questions for the fixing thread

- Is one shared comparative guard right, or does `rendered_html` need different
  signals? It is post-substitution render output, not a template — a size-collapse
  rule may behave differently when content_data varies between renders.
- `save_page_sections` is the only writer of `page_components` (DELETE+INSERT, per
  016b durable invariants). A DELETE+INSERT has no "current row" at insert time in
  the way an UPDATE does — the guard would need to read the prior row first.
- Should the guard live at the action layer (as now) or lower, so any writer
  inherits it? The action layer was chosen for 012 because it was where the
  incident was; a lower seam covers more but touches more.

---

> **PROGRESS 2026-07-21 — INSTANCE 1 diagnosed + fixed (committed, inert until an
> image roll). Stays OPEN until it ships. `durable_write_guard` workstream
> (`docs024_key_docs_latest/durable_write_guard/`), commit `ba702c8c6`.**
>
> **Corrects the scope this file asked for.** The two named surfaces are
> non-exposures, verified live 2026-07-21:
> - `pages.rendered_*` — **no Go writer anywhere** (exhaustive grep) and **0 of 301
>   pages** carry a value. Dormant columns; chrome lives in
>   `site_components.rendered_html`, assembled deterministically. A guard here
>   protects a write that never happens.
> - `page_components.rendered_html` — a **derived render** of the guarded+versioned
>   `content_components.html_template` (for every tool, `rendered_html` length ≈
>   `html_template` length, 98–122%). Recovered by re-render. An agent census of
>   all 19 writers (both tables) found **none** takes a whole LLM completion into
>   it as an *overwrite*: they are deterministic renders (a comparative guard would
>   fire on legitimate output) or deterministic CSS transforms (it would block
>   legitimate fixes). So the `componentRegressionIssues` *comparative* shape does
>   NOT belong here.
>
> **The real hole is at BIRTH, not overwrite.** `create_tool_component` and
> `store_generated_component` INSERT a whole LLM artifact with no completeness
> check strong enough to see a *tail* cut — `HasToolDocHeader` only proves the
> top-of-`<script>` sentinels, and `scoreComponent.TemplateClosed` balances
> `<section>` only. That is the 012 shape at birth, and it is **live**: it is how
> `bugs_open/046`'s 8 tools were born serving broken JavaScript on 6 domains
> (7 of 8 with no intact prior version). Because there is no prior row at birth,
> the fix is an **absolute** structural check, not the comparative guard this file
> imagined.
>
> **Shipped (owner scope decision: Phase 1 + section gate):**
> - `hasUnbalancedStructuralTags` — the absolute 5-pair tag-imbalance signal
>   (`component_write_guard.go`), calibrated 0-over-fire fleet-wide per 046;
>   ends-mid-token excluded (tool-safe but +36 FP on non-tools).
> - `create_tool_component` gates birth on `componentTemplateValid(_, "tool")` —
>   the SAME predicate the schema loader applies, so a tool that would be dropped
>   at LOAD can't be BORN. Hard fail → `agent_error_log`
>   (`tool_birth_truncation_blocked`) → `needs_human_review`.
> - `store_generated_component` rejects an unbalanced `<script>/<style>/…` section.
> - Tests incl. the 046 "cut `<script>` after a closed `</section>`" discriminator.
>
> **Coordination:** this is the PREVENTION half of the stop-new/clean-old split;
> `bugs_open/046` (`truncation_casualties_046`) owns the SWEEP of the 8 existing
> casualties and its own PLAN explicitly routed the write guard here. Non-competing.
> Verify-after-ship steps are in the workstream PLAN (fault-inject a tail-cut tool
> → not created, item `needs_human_review`, refusal logged).

---

## INSTANCE 2 — the completion-verification framework, same shape (added 2026-07-19, reasoning-dataset thread)

Filing here rather than as a separate bug, because it is this file's pattern
exactly and a second account would drift.

The completion gate (`complete_work_item_verification.go`) consults a per-item_type
verifier before stamping `complete`, and routes a failed verification into the
attempt machinery. Good mechanism, built properly, and:

**`RegisterVerifier` has been called exactly ONCE in the entire codebase** —
`RegisterVerifier("empty_section", VerifyEmptySectionResolved)`
(`check_empty_sections.go:38`). There are ~50 item types with discovery checks.

Live evidence, 2026-07-19:

```
site_work_items status='complete'                     : 4,570
  ...carrying a result._verification record           :     5
site_work_items ever status='verified'                :     9  (all empty_section,
                                                                none since 07-14)
```

So ~49 item types complete on the handler saga's self-report — the precise shape
this file describes: *"one call site patched, mechanism left generic."* The
mechanism is even opt-in by construction (`verifiers.go:47-51`: *"Called from
init() in check files"*), so it silently stays at one unless an author
remembers.

**Why it is the same bug and not a feature request:** the gate exists because a
saga can "succeed" without touching the defect (robot-hands' gripper-detail
sections, complete on 2026-07-10, still empty on 2026-07-14). Registering it for
one item type fixes that one incident and leaves the class open, which is what
this file was opened to stop happening again.

**Suggested shape** (declined by the reasoning-dataset thread as out of its
scope — it is read-only for `platform/`): make the gap *fail loud* rather than
documented. A `VerifierCoverage()` helper plus a test asserting every registered
check either has a verifier or appears on an explicit known-gap list with a
reason, so adding a check without a verifier breaks the build. A council
`bug_historian` seat reviewing that proposal noted the helper must not iterate
only the check registry — an item_type created by a path that never registered a
discovery check would be invisible to the coverage guard, under-reporting the
very gap it exists to expose.

> **PROGRESS 2026-07-20 — work_item_completion_integrity thread. INSTANCE 2 stays
> OPEN; two of its three parts are committed but inert.** Commit `08b35ccc4`.
>
> **Corrects this section's diagnosis.** INSTANCE 2 (and `verifiers.go`) attribute the
> gap to the mechanism being opt-in — *"stays at one unless an author remembers"* —
> which aims the fix at discipline. That is incomplete. The contract passed only
> `(ctx, db, spec, logger)`, and measured over all 5,514 live items: 2,370 specs carry
> `page_id`, 310 `component_id`, **9** `site_id`. For a site-aggregate type — this
> file's own example `hardcoded_section_colors` files ONE item per site and its
> predicate needs the site_id — a verifier was **unwritable**, however willing the
> author. So `submission_B`'s proposal to register exactly that verifier was
> unimplementable as specced. `VerifyTarget{ItemID,SiteID,PageID,ItemType,Spec}` now
> carries the identity; site_id is NOT NULL so it costs nothing.
>
> **The "suggested shape" is built**, and it answers the bug_historian objection this
> section records. `verifier_coverage_test.go`: all 69 live item types must be either
> verified or classified (mechanical / creation / judgement / no_target) with a
> reason, or the build fails. The objection was righter than stated — item types are
> string literals inside each check's `Run` and so are **not enumerable at runtime at
> all**, and the highest-volume types (`cta_improvement` 313, `needs_content_planning`
> 387, `spacing_fix` 116) come from paths with no discovery check. The denominator is
> therefore live-DB-sourced, refresh query in the workstream RUNBOOK; its weakness — a
> ratchet, not a sensor — is stated in the file.
>
> **Coverage is still 1 verifier, deliberately.** A `page_rerender` verifier (1,849 of
> 4,644 completions) was written, tested and **held**: re-running the misdirected-CTA
> predicate over a whole page is stricter than the handler's remit — the rerender only
> rewrites CTA fields in `ctaFieldNames` components, and a prose misdirect is
> *deliberately* left for two-strike escalation — so it would have marked
> correctly-handled items unresolved and stranded them in `failed`. Logged in
> `WRONG_CALLS.md`; the finding is the guard's gap entry for `page_rerender`. Writing
> it needs the rendered component mapped back to its spec section's
> `component.function`. **Do that before re-attempting it.**
>
> Docs: `docs024_key_docs_latest/work_item_completion_integrity/`.

**Related, filed separately because it is a distinct live defect rather than a
coverage gap:** `bugs_open/032` — the one registered verifier reports
`Resolved: true` when its target row is *absent*, so a rebuild that silently
deleted a component reads as a successful fix.

## References

- `bugs_open/012` — the incident and the shipped guard.
- `bugs_open/032` — instance-2's sibling defect (verifier false-positive on a
  deleted target).
- `016b` §9 "LLM truncation persisted as a successful artifact" — the pattern,
  the calibration story, and the counter-examples that killed the first check.
- `platform/orchestration/actions/component_write_guard.go` — the shape to reuse.
- Council report: `diagnosis_artifacts` where
  `correlation_id='e8827490-764a-4c90-b4db-72e358f9be87'` and `kind='council_report'`.
