-- 097_provocations_archive_arena_cta.sql — make "Enter today's Arena" reach the Arena
-- Created 2026-07-15. Run AFTER 094 (Arena deployed).
--
-- Why 095b's provocations retarget did not stick: provocations-archive-list's
-- cta_url is source: "static" (fallback /index.html, semantic "the day's
-- provocation page"). A static field re-applies on every render, so the
-- content_data edit 095b wrote was clobbered back to /index.html on the
-- cta_links_stale rerender — the exact landmine 091/096 address.
--
-- The label is llm-authored "Enter today's Arena"; the Arena is now a real page.
-- Copy and destination must agree (the workstream's whole point), and the
-- close-the-loop misdirected_cta check will otherwise correctly flag this. Fix:
-- flip cta_url source static -> renderer (so stored content_data wins), then
-- author /tools/arena/index.html into the one vonc instance. Fallback /index.html
-- is kept as the safety net. provocations-archive-list is a vonc-only component
-- (1 instance), so blast radius is that instance.
--
-- Reversal: _vonc_097_backup_20260715 (component + content).

BEGIN;

CREATE TABLE _vonc_097_backup_20260715_component AS
  SELECT * FROM content_components WHERE function = 'provocations-archive-list';
CREATE TABLE _vonc_097_backup_20260715_content AS
  SELECT pc.id, pc.page_id, pc.slot_name, pc.content_data
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN content_components cc ON cc.id = pc.component_id
  WHERE p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
    AND cc.function = 'provocations-archive-list';

-- 1. Schema: cta_url static -> renderer (content_data becomes authoritative).
UPDATE content_components
SET input_schema = jsonb_set(input_schema, '{fields,cta_url,source}', '"renderer"'),
    updated_at = NOW()
WHERE function = 'provocations-archive-list'
  AND input_schema->'fields'->'cta_url'->>'source' = 'static';

-- 2. Author the Arena URL into the vonc instance's content_data.
UPDATE page_components pc
SET content_data = jsonb_set(pc.content_data, '{cta_url}', '"/tools/arena/index.html"'),
    updated_at = NOW()
FROM pages p, content_components cc
WHERE p.id = pc.page_id AND cc.id = pc.component_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
  AND cc.function = 'provocations-archive-list'
  AND pc.content_data->>'cta_label' ILIKE '%arena%';

-- Verify.
DO $$
DECLARE src TEXT; url TEXT;
BEGIN
  SELECT input_schema->'fields'->'cta_url'->>'source' INTO src
  FROM content_components WHERE function = 'provocations-archive-list';
  SELECT pc.content_data->>'cta_url' INTO url
  FROM page_components pc JOIN pages p ON p.id = pc.page_id JOIN content_components cc ON cc.id = pc.component_id
  WHERE p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND cc.function = 'provocations-archive-list'
    AND pc.content_data->>'cta_label' ILIKE '%arena%';
  IF src <> 'renderer' THEN RAISE EXCEPTION 'verify failed: cta_url source = % (want renderer)', src; END IF;
  IF url IS DISTINCT FROM '/tools/arena/index.html' THEN
    RAISE EXCEPTION 'verify failed: content_data cta_url = % (want /tools/arena/index.html)', url;
  END IF;
  RAISE NOTICE 'verified: provocations-archive-list cta_url renderer-sourced, instance points at the Arena';
END $$;

COMMIT;

-- Follow with a cta_links_stale page_rerender for provocations-index; the
-- renderer source now yields to content_data, so /tools/arena/index.html sticks.
