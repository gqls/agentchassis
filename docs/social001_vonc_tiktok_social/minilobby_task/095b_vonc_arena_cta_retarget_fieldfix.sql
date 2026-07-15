-- 095b_vonc_arena_cta_retarget_fieldfix.sql — correct 095's field-name miss
-- Created 2026-07-15. Run AFTER 094 (Arena deployed) and 095.
--
-- BUG in 095: it matched the call-to-action TEXT field as `primary_cta_text`,
-- but the live call-to-action component names that field `primary_cta` (only
-- the URL field is `primary_cta_url`). So 095's `primary_cta_text ILIKE
-- '%arena%'` never matched, catalyst's call-to-action ("Enter the Arena" ->
-- still /tools/gauntlet/index.html) was left misdirected, and 095's verify
-- passed blind to it (it checked the same wrong field name). 095 only caught
-- the guide page's hero (cta_text), which is correct but wasn't the target.
--
-- Also handles two arena-naming CTAs 095 never scoped:
--   - tool-arena-interface-guide call-to-action: primary_cta "Enter the Arena",
--     empty primary_cta_url -> the Arena.
--   - provocations-index provocations-archive-list: cta_label "Enter today's
--     Arena", circular cta_url /index.html -> the Arena (was the
--     cta_names_unknown_destination finding; now the Arena exists).
--
-- Match is on the copy itself (text ILIKE '%arena%'), not row ids. Idempotent.
-- Reversal: _vonc_095b_backup_20260715_content.

BEGIN;

-- Pre-flight: the Arena must be deployed.
DO $$
DECLARE ok INT;
BEGIN
  SELECT COUNT(*) INTO ok FROM pages
  WHERE site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
    AND url='/tools/arena/index.html' AND build_status='deployed';
  IF ok <> 1 THEN
    RAISE EXCEPTION 'pre-flight failed: /tools/arena/index.html not deployed';
  END IF;
END $$;

CREATE TABLE _vonc_095b_backup_20260715_content AS
  SELECT pc.id, pc.page_id, pc.slot_name, pc.content_data
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN content_components cc ON cc.id = pc.component_id
  WHERE p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
    AND ( (cc.function IN ('hero','call-to-action')
           AND pc.content_data->>'primary_cta' ILIKE '%arena%')
       OR (cc.function = 'provocations-archive-list'
           AND pc.content_data->>'cta_label' ILIKE '%arena%') );

-- call-to-action: primary_cta (text) names the Arena -> primary_cta_url
UPDATE page_components pc
SET content_data = jsonb_set(pc.content_data, '{primary_cta_url}', '"/tools/arena/index.html"'),
    updated_at = NOW()
FROM pages p, content_components cc
WHERE p.id = pc.page_id AND cc.id = pc.component_id
  AND p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
  AND cc.function IN ('hero','call-to-action')
  AND pc.content_data->>'primary_cta' ILIKE '%arena%';

-- provocations-archive-list: cta_label names the Arena -> cta_url
UPDATE page_components pc
SET content_data = jsonb_set(pc.content_data, '{cta_url}', '"/tools/arena/index.html"'),
    updated_at = NOW()
FROM pages p, content_components cc
WHERE p.id = pc.page_id AND cc.id = pc.component_id
  AND p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
  AND cc.function = 'provocations-archive-list'
  AND pc.content_data->>'cta_label' ILIKE '%arena%';

-- Verify: no arena-naming CTA (by the CORRECT field names) points elsewhere.
DO $$
DECLARE mismatched INT; retargeted INT;
BEGIN
  SELECT COUNT(*) INTO mismatched
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN content_components cc ON cc.id = pc.component_id
  WHERE p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
    AND ( (cc.function IN ('hero','call-to-action')
           AND pc.content_data->>'primary_cta' ILIKE '%arena%'
           AND pc.content_data->>'primary_cta_url' IS DISTINCT FROM '/tools/arena/index.html')
       OR (cc.function = 'provocations-archive-list'
           AND pc.content_data->>'cta_label' ILIKE '%arena%'
           AND pc.content_data->>'cta_url' IS DISTINCT FROM '/tools/arena/index.html') );
  SELECT COUNT(*) INTO retargeted FROM _vonc_095b_backup_20260715_content;
  IF mismatched <> 0 THEN
    RAISE EXCEPTION 'verify failed: % arena-naming CTA slots still point elsewhere', mismatched;
  END IF;
  RAISE NOTICE 'verified: % component rows retargeted to the Arena', retargeted;
END $$;

COMMIT;

-- Follow with a cta_links_stale page_rerender for catalyst, provocations-index,
-- and tool-arena-interface-guide so the change reaches deployed HTML.
