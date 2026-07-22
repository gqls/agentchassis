-- R7c (2026-07-22) — fix bug_open/043 point (b) at its ROOT: the system-stats
-- component's input_schema hands out junk unit suffixes as fallbacks.
--
-- WHY R7b DID NOT HOLD. gripper-detail's stat suffixes are field source=static with
-- fallback "%","ms","+","x" on the SHARED component
-- content_components.id=fdd92ad4-521a-4602-89cf-7ee1a66c10f1. An empty suffix resolves
-- to the fallback AND persists it, so R7b's suffix="" reverted to "%" within one render.
-- (Values/labels/descriptions are source=llm with no fallback, which is why R7's values
-- and R7b's descriptions held and only the suffixes reverted.) This is the structural
-- source of 043 point (b) and it is fleet-wide: all five system-stats consumers render
-- the placeholder — ai-agent-orchestration renders "1,000sms" and vonc "14,203%", which
-- proves the fallback is leaking, not an intended unit.
--
-- FIX. A stat has no unit unless one is specified, so the correct default is "" not "%".
--   1. Set the four fallbacks to "" on the shared component. Because the field is
--      source=static (deterministic, not LLM) and the OTHER four pages carry their
--      suffix persisted NON-empty, this changes NO current live page — only the default
--      for a genuinely-empty suffix. Verified blast radius before applying: 5 pages /
--      4 sites, every one non-empty.
--   2. Clear robot-hands gripper-detail's four persisted "%/ms/+/x" back to empty, so the
--      static resolve lands on the new empty fallback and renders no suffix.
--   3. Re-render gripper-detail.
-- Residual (NOT touched — other owners' sites): the four non-robot-hands rows still carry
-- the junk suffix persisted; each needs a clear + re-render, and the intended unit is the
-- site owner's call. Listed in bugs_open/043 Update 2026-07-22 (b).
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < THIS_FILE

\set site '00ff3af5-dad8-4770-9f70-3edc267a3c92'
\set comp 'fdd92ad4-521a-4602-89cf-7ee1a66c10f1'

BEGIN;

-- 1. Root: junk unit fallbacks -> empty (shared component; zero live-page impact)
UPDATE content_components
   SET input_schema = jsonb_set(jsonb_set(jsonb_set(jsonb_set(
         input_schema,
         '{fields,stat1_suffix,fallback}', '""'::jsonb),
         '{fields,stat2_suffix,fallback}', '""'::jsonb),
         '{fields,stat3_suffix,fallback}', '""'::jsonb),
         '{fields,stat4_suffix,fallback}', '""'::jsonb),
       updated_at = now()
 WHERE id = :'comp';

\echo '--- fallbacks after fix (all four must be empty) ---'
SELECT key, value->>'fallback' AS fallback
FROM content_components, jsonb_each(input_schema->'fields')
WHERE id = :'comp' AND key ~ 'stat[0-9]_suffix' ORDER BY key;

-- 2. robot-hands: clear the persisted junk so the resolve uses the new empty fallback
UPDATE page_components pc
   SET content_data = pc.content_data
       || '{"stat1_suffix":"","stat2_suffix":"","stat3_suffix":"","stat4_suffix":""}'::jsonb,
       updated_at = now()
 WHERE pc.page_id = (SELECT id FROM pages WHERE site_id = :'site' AND name='gripper-detail')
   AND pc.slot_name = 'system-stats';

-- 3. Re-render gripper-detail
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, status, created_by, pipeline,
   priority, triaged_at, handler_agent, item_key, spec)
SELECT
  p.site_id, 'robot-hands-r7c-suffix-root-fix', 'page_rerender', 'medium',
  'Rerender gripper-detail — stat suffixes now resolve to empty (043(b) root fallback fix)',
  'triaged', 'session-2026-07-22-robot-hands', 'build',
  20, now(), 'page-rerender',
  'page_rerender_' || p.name || '_r7csuffix_' || p.site_id::text,
  jsonb_build_object('domain','robot-hands.com','reason','cta_links_stale',
                     'page_id',p.id,'page_name',p.name,'filename',ltrim(p.url,'/'))
FROM pages p
WHERE p.site_id = :'site' AND p.name = 'gripper-detail';

\echo '--- queued ---'
SELECT status, handler_agent, count(*) FROM site_work_items
WHERE source='robot-hands-r7c-suffix-root-fix' GROUP BY 1,2;

COMMIT;
