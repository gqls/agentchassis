# RUNBOOK — internal-linking fix batch (gamesdesign.co.uk)

**State (2026-06-12):** chassis image with all Go changes DEPLOYED. Agent row + writer wiring applied 06-11 (verified). DB template/schema SQLs: run `check_linking_sql_applied.sql` to see which are in — don't rely on memory.

```bash
SITE_ID=$(psql -Atc "SELECT id FROM sites WHERE domain='gamesdesign.co.uk'")   # re-resolve after any teardown
```

---

## 0. Orient
1. `check_linking_sql_applied.sql` — four blocks, expected values in comments. Any `f` ⇒ run that one SQL file, re-check.
2. Tags moved together (a lagging resolver tag = old-image pods = permanent silent fallback):
```sql
SELECT type, image_tag FROM agent_definitions
WHERE type IN ('internal-link-resolver','page-content-writer','research-agent');
```

## 1. DB steps (only those the check shows missing), in order
1. `step1_hero_cta_phantom_fix.sql`
2. `layer1b_header_footer_phantom_fix.sql`
3. `b4_b5_hub_links_schema.sql`   (image must be live first — it is)
4. `b4_b5_hub_links_template_gate.sql`
Each file ends with its own verification SELECT; every `UPDATE` on a template `replace()` must report `UPDATE 1` **and** the flag must flip — a flag stuck at `f` means the match string differed (whitespace) and the replace no-op'd. Fix the string; don't assume.

## 2. Render vs rebuild — what fixes what
**Re-render** (`page-rerender`, `rerender-pages`) re-applies templates to component data **stored at last build**. **Rebuild** (work item → `build-dispatch-loop` → `page-build-handler`) re-runs the full write path: `plan_sections` source resolution → writer (→ `resolve_links` sub-agent) → render.

Consequences:
- Header/footer: **re-render suffices** — the new Go builds their data fresh at render time (real contact page, `GetNavItems` legal links).
- Hero/CTA pages: **rebuild required** — stored data still holds the fabricated `/contact.html`; the gate sees a value and renders the phantom on a mere re-render.
- List pages (Browse-All): re-render only *hides* the button (stored `cta_url` empty + new gate); **rebuild** resolves the hub URL via `query.section_index_for`.
- Do **NOT** use the `page-rebuild` agent for this: its writer call doesn't map `section_plan` (v2_33 gap) — sectionless-rebuild risk.

## 3. Site-wide re-render (header/footer + gates) — known machinery, 081d pattern
```bash
SITE_ID=$(psql -Atc "SELECT id FROM sites WHERE domain='gamesdesign.co.uk'")
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid); ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid); MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

kubectl -n kafka run -i --rm kcat-rerender-gd-$(date +%s) \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID -H orchestration_id=$ORCHESTRATION_ID \
  -H request_id=$REQUEST_ID -H message_id=$MESSAGE_ID \
  -H message_type=request -H client_id=demo_client \
  -H action=orchestrate -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"rerender-pages"},"input_data":{"site_id":"${SITE_ID}","domain":"gamesdesign.co.uk","refresh_site_components":true}}
JSON
```
(Single-page spot check: same pattern with `agent_type: page-rerender` + `page_id` — 081b.)

## 4. Measure (dry-run audit SQL against rendered HTML)
Expected after step 3:
- `site_component` `/contact.html` `/privacy.html` `/terms.html` — **gone** (header CTA → real contact page; no legal links).
- `empty_internal_href` on the three lists — **gone as empties** (button hidden by gate; hub link comes in step 5).
- Hero `/contact.html` `/services.html` — **likely still present** (stored data; see §2). These rows are step 5's worklist. If they're unexpectedly gone, `rerender-pages` re-resolves sources — note it in the running notes and step 5 shrinks to the one-page resolver test.

## 5. Rebuild the pages needing source re-resolution (081c path: work item → build-dispatch-loop → page-build-handler)
Worklist (confirm join column names with `\d page_components` first):
```sql
SELECT DISTINCT p.id, p.name, p.url
FROM pages p
JOIN page_components pc ON pc.page_id = p.id
JOIN content_components cc ON cc.id = pc.component_id
WHERE p.site_id = :'site_id'
  AND cc.function IN ('hero','call-to-action','tool-list','game-list','guide-list');
```
Before the INSERT, two confirmations (081c stopped exactly here):
1. `\d site_work_items` — column names (`handler_agent`, `spec`, `item_key`, …).
2. `SELECT jsonb_pretty(default_config->'workflow') FROM agent_definitions WHERE type='page-build-handler';` — which `item_type`s it branches on and what it reads from `spec` for an EXISTING page (the `empty_section` items from `check_empty_sections` are the proven existing-page analog; `needs_content_page` is the gap-planner NEW-page shape).

INSERT skeleton (finalize `item_type`/`spec` from confirmation 2 — paste the workflow back and I'll pin it):
```sql
INSERT INTO site_work_items (
  site_id, source, pipeline, item_type, severity, summary,
  spec, priority, status, created_by, item_key, handler_agent
) VALUES (
  :'site_id', 'runbook', 'content', '<from confirmation 2>', 'medium',
  'Rebuild <page>: re-resolve internal links post linking-fix batch',
  '{"page_id":"<id>","page_name":"<name>","reason":"linking batch §5"}'::jsonb,
  100, 'triaged', 'runbook-linking', 'linking_rebuild_<page_name>', 'page-build-handler'
)
ON CONFLICT DO NOTHING;
```
`status='triaged'` because `build-dispatch-loop` claims `triaged`/`approved` and `improvement-sweep` (the triager) is disabled.

**First item doubles as the resolver end-to-end check.** Insert ONE (the index page), watch:
- writer orchestration: `resolve_links → select_sections → process_sections_loop`, `sections_for_render.sections_ready` count = plan count;
- resolver log: `resolve_internal_links: augmented CTA sections` `hub_count=5` `unresolved=0`;
- rendered hero: both buttons → top two content hubs;
- `SELECT count(*) FROM site_work_items WHERE item_type='unresolved_cta' AND site_id=:'site_id';` → 0.
Then insert the rest of the worklist.

## 6. Done-criteria for the batch
Dry-run: zero phantom rows, zero `empty_internal_href`, on BOTH surfaces. Browse-All buttons link `/tools/index.html` `/games/index.html` `/guides/index.html`. `unresolved_cta` = 0.

## 7. Later, deliberately (unchanged)
1. Enable `phantom_internal_links` in the discovery checks array (routing targets `nav-link-fixer`, `page-build-handler` both exist); watch one sweep's work items; then re-enable `improvement-sweep`.
2. Readopt gamedesign.uk → gamesdesign.co.uk as the from-scratch acceptance + fresh content-quality baseline — only after §6 passes, so any readopt failure is attributable. During readopt watch: per-page resolver log lines; `unresolved_cta` stays 0 on the NEW site_id; dry-run clean on the virgin build. EXPECTED recurrences (adopt-path, untouched): brand-suffix card titles, tool-flavoured guide copy, empty tool descriptions, footer metadata — next package's input, not linking regressions.

## Rollback
- Templates/schema: restore from `content_components_bak_cta0610` / `_navfix_0610` / `_hubfix_0610`.
- Writer wiring: `SELECT revert_agent('page-content-writer');`
- Code: roll back the chassis image tag.
