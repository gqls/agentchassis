# NOTES — bugs_open/029 tool-suggester phantom links (append-only, newest at bottom)

## 2026-07-21 — session start, diagnosis

- Picked up `/bugs_open/029_..._tool_suggester_writes_phantom_tool_links.md` (NOT the
  same-numbered `..._hung_spawns_...` — 029 is an ambiguous number; `who-owns.py 029` warns and
  its "OWNED" verdict conflates the two files. This one had no active owner.).
- Read related bugs 023 (CTA label/url pairing), 049 (chrome staleness + planned-but-unbuilt
  pages). 049 explicitly lists 029 as related-but-distinct ("another route to a 404, from page
  content rather than chrome"). Confirmed 029 is its own defect.
- Design doc `005_tool_pipeline.md` names the culprit step: `create_cross_links` →
  `create_tool_cross_link_items.go`, run after `create_items_loop`.

### What the DB showed (all queries in RUNBOOK)

- Fleet sweep R1: **0 of 24** constructed `/tools/{function}.html` URLs resolve to a real page.
- The one built tool (`tool-process-automation-scorer`) deployed at
  `/tools/process-automation-scorer/index.html` — item points at `/tools/tool-process-
  automation-scorer.html`. So "tool was never built" is NOT the whole story; the URL is wrong
  even when the tool exists.
- R3: deployed tool URLs use three incompatible shapes. No deterministic function→url map.

### What the code showed (Explore agent + direct reads)

- `create_tool_cross_link_items.go:142` fabricates the URL from suggestion-time data; guards
  only check the *related* page exists, never the *tool's* page.
- `deploy_tool_action.go` / `create_tool_component_action.go` create the page first, then emit
  follow-on items with the real `page_id`/url — the pattern to reuse.
- `add_tool` spec carries `related_pages` (R4) via `spec_data: current_suggestion` — the build
  path already has what it needs.
- Downstream: internal-link-resolver only does CTA fields; `validate_page_content.go` detects
  the in-body phantom but files it `warning` (non-blocking) → deploys anyway.

### MISSTEP / correction caught this session

- The Explore agent read `k8s/bk_agent_definitions_backup.sql:214` and reported page-build-
  handler does NOT thread `spec.suggestion` to the writer. That contradicted the live evidence
  (the phantom link IS in leopardess's rendered services page). Checked the **live** row (R5):
  it DOES map `rewrite_guidance? → input_data.spec.suggestion`. The backup is stale (predates
  migration 072). Lesson (again): verify agent config against the live row, never the k8s backup.
- The original 2026-07-19 handoff said the diagnosis item "never dispatched — no orchestration
  row". Today it is `status='complete'`. That was queue latency (~30 min), not a drop — same
  trap as the council queue. No durable verdict landed in `doc_notes`, so primary evidence stands.

### Decision

Fix at the build path (candidate 1), not the emitter (runs too early) or the consumer (needs a
deferral mechanism). See PLAN. Next: implement, route through council gate (platform change),
build + roll + verify.
