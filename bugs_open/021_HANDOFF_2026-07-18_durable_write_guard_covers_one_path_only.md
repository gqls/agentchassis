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
