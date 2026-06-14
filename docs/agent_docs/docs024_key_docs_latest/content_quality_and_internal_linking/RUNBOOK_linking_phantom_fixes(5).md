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

## 4. Measure (post-deploy AUDIT query against rendered HTML — read-only; not a dry run)
Expected after step 3:
- `site_component` `/contact.html` `/privacy.html` `/terms.html` — **gone** (header CTA → real contact page; no legal links).
- `empty_internal_href` on the three lists — **gone as empties** (button hidden by gate; hub link comes in step 5).
- Hero `/contact.html` `/services.html` — **likely still present** (stored data; see §2). These rows are step 5's worklist. If they're unexpectedly gone, `rerender-pages` re-resolves sources — note it in the running notes and step 5 shrinks to the one-page resolver test.

## 5. Rebuild the pages needing source re-resolution (081c path: work item → build-dispatch-loop → page-build-handler)

Handler facts (from its definition, 2026-06-12):
- It does **not** branch on `item_type` — dispatch metadata only. It reads `input_data.spec.page_id` / `spec.page_name` (`page_name` is also used by `save_sections`/`update_status`: **mandatory in spec**), `spec.mode` (recreate ⇒ load adoption crawl, preserves original copy), `spec.suggestion` (→ writer `rewrite_guidance`).
- Its `call_content_writer` maps `"section_plan": "section_plan"` — so THIS path feeds the resolver (`sections?` resolves). It validates (our non-blocking phantom warnings don't fail it), saves sections, sets page status, spawns `page-rerender`, deploys via git — one commit per page.
- Dedup: partial unique `(site_id, item_key)` on non-terminal statuses — `ON CONFLICT DO NOTHING` is safe; a completed item's key can be reused.
- `pipeline='build'` is the dispatch-proven combination (gap planner); `status='triaged'` (the triager is disabled).

Preview the worklist (join `pc.component_id = cc.id`, confirmed from check_empty_sections):
```sql
SELECT DISTINCT p.name, p.url, cc.function
FROM pages p
JOIN page_components pc ON pc.page_id = p.id
JOIN content_components cc ON cc.id = pc.component_id
WHERE p.site_id = (SELECT id FROM sites WHERE domain='gamesdesign.co.uk')
  AND cc.function IN ('hero','call-to-action','tool-list','game-list','guide-list')
ORDER BY p.name;
```

**First: ONE item (index) — doubles as the resolver end-to-end check:**
```sql
INSERT INTO site_work_items (
  site_id, source, pipeline, item_type, severity, summary, spec,
  page_id, priority, handler_agent, status, created_by, item_key
)
SELECT p.site_id, 'runbook', 'build', 'link_resolution_rebuild', 'medium',
       'Rebuild ' || p.name || ': re-resolve internal links (linking batch §5)',
       jsonb_build_object(
         'page_id', p.id, 'page_name', p.name, 'mode', 'recreate',
         'suggestion', 'Preserve the existing copy; this rebuild exists to re-resolve internal link destinations (hero CTA / Browse-All).',
         'reason', 'linking batch §5'),
       p.id, 60, 'page-build-handler', 'triaged', 'runbook-linking',
       'linking_rebuild_' || p.name
FROM pages p
WHERE p.site_id = (SELECT id FROM sites WHERE domain='gamesdesign.co.uk')
  AND p.name = 'index'
ON CONFLICT DO NOTHING;
```
Watch: item `claimed` → writer orchestration `resolve_links → select_sections → process_sections_loop` with `sections_for_render.sections_ready` = plan count → resolver log `resolve_internal_links: augmented CTA sections` `hub_count=5` `unresolved=0` → hero buttons → top two hubs → git commit → item `complete`. `unresolved_cta` count stays 0.

Watch SQL (run A/B repeatedly; C/D after items complete; E is ground truth):
```sql
SITE='gamesdesign.co.uk';  -- (psql var; or inline the site_id)

-- A) progress: triaged → claimed → complete
SELECT status, count(*), max(updated_at) AS latest
FROM site_work_items
WHERE created_by='runbook-linking'
  AND site_id=(SELECT id FROM sites WHERE domain=:'SITE')
GROUP BY status ORDER BY status;

-- B) resolver distress signal — must stay 0
SELECT count(*) AS unresolved_cta
FROM site_work_items
WHERE item_type='unresolved_cta'
  AND site_id=(SELECT id FROM sites WHERE domain=:'SITE');

-- C) resolver ran — its output nests inside the WRITER's collected_data, not its
--    own orchestration; the reliable confirmation is the pod log line, not SQL:
--      kubectl -n ai-persona-system logs -l app=internal-link-resolver --tail=300 | grep augmented
--    expect per page:  resolve_internal_links: augmented CTA sections  hub_count=5 unresolved=0

-- D) writer iterated augmented (not fallback): for_render vs plan_count.
--    section_plan arrives via input_mapping, so it nests under input_data.
SELECT orchestration_id, current_step, status,
       jsonb_array_length(collected_data->'sections_for_render'->'sections_ready')        AS for_render,
       jsonb_array_length(collected_data->'input_data'->'section_plan'->'sections_ready') AS plan_count
FROM orchestration_states
WHERE collected_data ? 'sections_for_render' AND updated_at > now() - interval '60 minutes'
ORDER BY updated_at DESC LIMIT 30;
-- for_render == plan_count ⇒ resolver augmented + loop consumed it. for_render>0 alone
-- already rules out a fallback-to-empty (that would give 0).

-- E) GROUND TRUTH — re-run the §4 audit query. Hero /contact.html+/services.html
--    and list empty_internal_href rows REPLACED by /tools|games|guides/index.html.
```
D's path is verified (2026-06-14: `for_render=2`); `section_plan` nests under `input_data` (it arrives via `input_mapping`). The resolver's own output does NOT surface as a top-level `link_resolution` key — it comes back inside the writer's `resolved_links.response`, so confirm the resolver ran via the **pod log line** (C), not SQL. **E (the §4 audit) is ground truth** — trust it over C/D if they disagree.

**Then the rest (dedup skips index):** same INSERT with the `p.name='index'` line replaced by
```sql
  AND EXISTS (SELECT 1 FROM page_components pc
              JOIN content_components cc ON cc.id = pc.component_id
              WHERE pc.page_id = p.id
                AND cc.function IN ('hero','call-to-action','tool-list','game-list','guide-list'))
```
Track: `SELECT status, count(*) FROM site_work_items WHERE created_by='runbook-linking' GROUP BY 1;` until all `complete`. Any `needs_human_review` = content validation blocked it (placeholders/contamination, not phantom warnings) — inspect that page's `validation_result`.

**Count note:** this "rest" INSERT matches every page that HAS a hero/CTA/list component, which is MORE than the pages the §4 audit flagged (the audit only catches pages whose *rendered* output carried a phantom; a page with a clean stored render still has the component). Observed 2026-06-13: 21 items, vs 11 pages with visible phantoms. Both are correct rebuilds (every matched page legitimately has CTA components and gets real hub destinations) — but it's more `recreate` content churn and a larger verification surface than §4 implied. To scope to just the proven-phantom set instead: `DELETE FROM site_work_items WHERE created_by='runbook-linking' AND status='triaged' AND page_id NOT IN (<§4 page_ids>)` before the dispatch loop claims them.

## 6. Done-criteria for the batch
Audit query (§4): zero phantom rows, zero `empty_internal_href`, on BOTH surfaces. Browse-All buttons link `/tools/index.html` `/games/index.html` `/guides/index.html`. `unresolved_cta` = 0.

## 7. Later, deliberately

### 7a. Enabling `phantom_internal_links` WITHOUT the sweep (observe-only)
Enabling the check ≠ enabling autonomous remediation. The check self-registers (`init()→Register`) but only runs when a discovery agent's `checks` array names it. Its findings land in `site_work_items` as `status='detected'` — and `detected` is NOT a claimable status (dispatch loops claim `triaged`/`approved`), so findings sit visible and inert until something triages them. `improvement-sweep` (the triager) stays disabled. So you get the audit signal with zero auto-action.

PREREQUISITE (one-line code fix, ship first): the check currently emits `page_component` findings with `pipeline='content'` (`routeBySurface`), but the dispatch-proven pipeline for `page-build-handler` is `'build'` (the §5 inserts used `'build'` and were claimed fine). Change the `page_component` case to `"build"` so that if/when a loop does triage these, they're claimable. Until then it only affects observation, not correctness.

Reads needed before editing (don't guess the array owner):
```sql
SELECT type, default_config #> '{workflow,steps}' IS NOT NULL AS has_wf
FROM agent_definitions
WHERE (default_config::text LIKE '%discovery%' OR default_config::text LIKE '%"checks"%')
  AND is_active = true;
```
Then snapshot the owning agent (`snapshot_agent(...)`), `jsonb_set`-append `"phantom_internal_links"` to its checks array, and run that discovery agent once via kcat (the §3 orchestrate pattern, `agent_type` = the discovery agent) against gamesdesign. Inspect: `SELECT item_type, surface, count(*) FROM site_work_items WHERE created_by=<discovery agent> AND status='detected' GROUP BY 1,2;` — on the post-§5 site this should be ~empty (the proof the batch worked end-to-end through the real check, not just the manual audit query).

### 7b. Re-enable `improvement-sweep` (only after 7a observed clean)
Separate, later decision. Watch the first triage cycle's work items before letting it run unattended.

### 7c. Readopt gamedesign.uk → gamesdesign.co.uk
The from-scratch acceptance + fresh content-quality baseline — only after §6 passes, so any readopt failure is attributable. During readopt watch: per-page resolver log lines; `unresolved_cta` stays 0 on the NEW site_id; §4 audit query clean on the virgin build. EXPECTED recurrences (adopt-path, untouched): brand-suffix card titles, tool-flavoured guide copy, empty tool descriptions, footer metadata — next package's input, not linking regressions.

## Rollback
- Templates/schema: restore from `content_components_bak_cta0610` / `_navfix_0610` / `_hubfix_0610`.
- Writer wiring: `SELECT revert_agent('page-content-writer');`
- Code: roll back the chassis image tag.
