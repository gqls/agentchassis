-- 095_vonc_arena_cta_retarget.sql — point the Arena CTAs at the real Arena
-- Created 2026-07-14. Run ONLY AFTER the Arena page from 094 is DEPLOYED
-- (the pre-flight block below hard-fails otherwise).
--
-- Context: the two Arena-naming CTAs ("Enter the Arena" / "Enter today's
-- Arena") were misdirected (circular -> /index.html). The cta_links_stale
-- rerender pass gave them the Gauntlet as an interim REAL destination; that
-- interim value is valid and non-excluded, so the recompute's exception rule
-- correctly refuses to change it again. Pointing copy that names the Arena AT
-- the Arena is therefore a hand retarget — matched by the copy itself (any
-- vonc hero/call-to-action CTA whose text says "arena"), not by row ids.
--
-- After this: dispatch a cta_links_stale rerender for the affected pages —
-- the exception rule keeps /tools/arena/index.html (real, non-excluded), and
-- the re-run misdirected_cta check goes quiet ("arena" now token-matches the
-- Arena page, and the href agrees).
--
-- Reversal: _vonc_095_backup_20260714_content.

BEGIN;

-- ── Pre-flight: the Arena page must be deployed ─────────────────────────────
DO $$
DECLARE ok INT;
BEGIN
  SELECT COUNT(*) INTO ok FROM pages
  WHERE site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
    AND url = '/tools/arena/index.html'
    AND build_status = 'deployed';
  IF ok <> 1 THEN
    RAISE EXCEPTION 'pre-flight failed: /tools/arena/index.html not deployed yet — run 094''s pipeline first';
  END IF;
END $$;

CREATE TABLE _vonc_095_backup_20260714_content AS
  SELECT pc.id, pc.page_id, pc.slot_name, pc.content_data
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN content_components cc ON cc.id = pc.component_id
  WHERE p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
    AND cc.function IN ('hero', 'call-to-action')
    AND (pc.content_data->>'cta_text' ILIKE '%arena%'
      OR pc.content_data->>'primary_cta_text' ILIKE '%arena%'
      OR pc.content_data->>'secondary_cta_text' ILIKE '%arena%'
      OR pc.content_data->>'secondary_cta' ILIKE '%arena%');

-- ── Retarget each URL slot whose companion text names the Arena ─────────────
UPDATE page_components pc
SET content_data = jsonb_set(pc.content_data, '{cta_url}', '"/tools/arena/index.html"'),
    updated_at = NOW()
FROM pages p, content_components cc
WHERE p.id = pc.page_id AND cc.id = pc.component_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
  AND cc.function = 'hero'
  AND pc.content_data->>'cta_text' ILIKE '%arena%';

UPDATE page_components pc
SET content_data = jsonb_set(pc.content_data, '{primary_cta_url}', '"/tools/arena/index.html"'),
    updated_at = NOW()
FROM pages p, content_components cc
WHERE p.id = pc.page_id AND cc.id = pc.component_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
  AND cc.function = 'call-to-action'
  AND pc.content_data->>'primary_cta_text' ILIKE '%arena%';

UPDATE page_components pc
SET content_data = jsonb_set(pc.content_data, '{secondary_cta_url}', '"/tools/arena/index.html"'),
    updated_at = NOW()
FROM pages p, content_components cc
WHERE p.id = pc.page_id AND cc.id = pc.component_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
  AND cc.function IN ('hero', 'call-to-action')
  AND (pc.content_data->>'secondary_cta_text' ILIKE '%arena%'
    OR pc.content_data->>'secondary_cta' ILIKE '%arena%');

-- ── Verify: every arena-naming CTA slot now points at the Arena ─────────────
DO $$
DECLARE mismatched INT; retargeted INT;
BEGIN
  SELECT COUNT(*) INTO mismatched
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN content_components cc ON cc.id = pc.component_id
  WHERE p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
    AND cc.function IN ('hero', 'call-to-action')
    AND ((pc.content_data->>'cta_text' ILIKE '%arena%'
          AND pc.content_data->>'cta_url' IS DISTINCT FROM '/tools/arena/index.html')
      OR (pc.content_data->>'primary_cta_text' ILIKE '%arena%'
          AND pc.content_data->>'primary_cta_url' IS DISTINCT FROM '/tools/arena/index.html')
      OR ((pc.content_data->>'secondary_cta_text' ILIKE '%arena%'
           OR pc.content_data->>'secondary_cta' ILIKE '%arena%')
          AND pc.content_data->>'secondary_cta_url' IS DISTINCT FROM '/tools/arena/index.html'));
  SELECT COUNT(*) INTO retargeted FROM _vonc_095_backup_20260714_content;
  IF mismatched <> 0 THEN
    RAISE EXCEPTION 'verify failed: % arena-naming CTA slots still point elsewhere', mismatched;
  END IF;
  IF retargeted = 0 THEN
    RAISE NOTICE '095: no arena-naming CTAs found in content_data — check cta text field names against live rows';
  ELSE
    RAISE NOTICE 'verified: % component rows retargeted to /tools/arena/index.html', retargeted;
  END IF;
END $$;

COMMIT;

-- NOTE: follow with a cta_links_stale page_rerender for the affected pages,
-- then re-run discovery — misdirected_cta must report zero arena findings.
