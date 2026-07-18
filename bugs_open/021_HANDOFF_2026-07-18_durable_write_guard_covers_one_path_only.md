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

## References

- `bugs_open/012` — the incident and the shipped guard.
- `016b` §9 "LLM truncation persisted as a successful artifact" — the pattern,
  the calibration story, and the counter-examples that killed the first check.
- `platform/orchestration/actions/component_write_guard.go` — the shape to reuse.
- Council report: `diagnosis_artifacts` where
  `correlation_id='e8827490-764a-4c90-b4db-72e358f9be87'` and `kind='council_report'`.
