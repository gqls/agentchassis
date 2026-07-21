# PLAN — bugs_open/046: sweep the truncation casualties, and make the class visible

**Started:** 2026-07-21 · bugfix-046 thread
**Bug:** `bugs_open/046_HANDOFF_2026-07-20_truncation_casualties_never_swept_live_tools_serve_broken_js.md`

## The problem, re-grounded (2026-07-21)

`bugs_closed/012` fixed the *cause* of component truncation and restored **one**
component by hand. Nobody swept for the others. Census re-run against the live DB
2026-07-21 — **still exactly 9 damaged active components** (8 `tool` + 1
`section`), unchanged from the 07-20 filing:

- The 5-pair unterminated-tag predicate (`<script>/<style>/<section>/<div>/<fieldset>`,
  open>close on any) catches **exactly these 9** fleet-wide — 0 over-fire.
- The `ends-mid-token` heuristic (used by `toolTemplateValid`) would add **36**
  false positives fleet-wide, so the sweep must NOT use it. Tag-imbalance alone
  is the precise signal.

Per-component facts (2026-07-21):

| component | level | on deployed page? | intact prior version? |
|---|---|---|---|
| tool-grip-force-friction-calculator-robot-hands-com | tool | 1 page_component (page not `deployed`-status, but live on host) | **yes — v2, 23,526 chars, balanced, ends `</script>`** |
| tool-llm-cost-calculator-…-finetuning-uk | tool | 2 | 1 version, none intact |
| tool-arena-interface-vonc-com | tool | 1 | none |
| tool-drop-rate-tuner-gamesdesign-co-uk | tool | 1 | none |
| tool-llm-cost-calculator-ai-agent-orchestration-com | tool | 1 | none |
| tool-process-automation-scorer-leopardessconsulting-co-uk | tool | 1 | none |
| tool-archetype-clash-calculator-vonc-com | tool | 0 (unplaced) | none |
| tool-llm-cost-calculator-…-leopardessconsulting-co-uk | tool | 0 (unplaced) | none |
| archetype-taster-quiz | section | 0 (unplaced) | — |

## Constraints that shape the fix (why not "just regenerate them all")

- **`needs_tool_recreation` → tool-recreation-handler can FABRICATE data**
  (`bugs_open/020`: it invented a 2,100-practice directory and shipped it live).
  So an automated sweep must **not** auto-route this class to a regenerator.
- **The re-render *delivery* pipeline is buggy and owned.** `bugs_open/024`
  (travelling-docs, actively worked 2026-07-21) — a template fix does not reach
  the live page; defect 6 (reason-less `page_rerender` collision) still open.
  Restoring a template fixes the *source*; the *live page* only updates when a
  re-render runs through 024's pipeline. Do **not** compete on that path.
- **7 of 8 tools have no intact prior version** — restore is unavailable; they
  need regeneration (LLM, heavy, and fabrication-guarded per 020).

## What this thread does (safe, verifiable, non-competing)

1. **Detection (the durable structural fix — bug candidate 2).** New discovery
   check `truncated_component`. Sweeps `content_components` used on a site's pages
   for the 5-pair imbalance; emits **detect-and-surface** items
   (`needs_human_review`, NO handler — the `dead_controls` pattern, safe under
   020). Spec carries `intact_version_available` + `intact_version_number` so
   triage (restore vs regenerate) is immediate. Reuses the
   `check_dead_controls` / `check_component_template_corrupted` machinery. Pure
   predicate is unit-tested against the census signatures.
   - **Not tool-scoped** (bug requirement): joins `content_components` at any
     level, so the `section` component is covered too.
   - Enabled on `completeness-discovery-agent` (alongside `dead_controls` +
     `component_template_corrupted`) via an **image-first** seed.
2. **Cheap repair (bug candidate 1, first step).** Restore
   `tool-grip-force-friction-calculator-robot-hands-com`'s template from its
   intact v2 — DB, live, no LLM. Census 9 → 8. Live-page delivery pending a
   re-render (024).
3. **Surface + document the other 7** as the tracked repair backlog. Do NOT
   auto-trigger recreation (020). The check will raise them once live; documented
   here meanwhile.

## Out of scope for this thread (handed off, not competing)

- Live-page **re-render delivery** for any component → `bugs_open/024`.
- **Render-surface write guard** → `bugs_open/021` (owned).
- **LLM regeneration** of the 7 no-intact-version tools → owner/next thread
  decides per-item (restore impossible; recreation is 020-guarded).

## Verify

Bug's own query must drop from 9 rows after the grip-force restore:
```sql
SELECT cc.name FROM content_components cc
WHERE cc.is_active AND length(cc.html_template) >= 100
  AND (SELECT count(*) FROM regexp_matches(lower(cc.html_template),'<script','g'))
    > (SELECT count(*) FROM regexp_matches(lower(cc.html_template),'</script>','g'));
```
Check + test: `go test ./platform/orchestration/actions/discovery_checks/`.
