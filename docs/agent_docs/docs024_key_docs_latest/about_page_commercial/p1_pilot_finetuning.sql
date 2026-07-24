-- ============================================================================
-- p1_pilot_finetuning.sql — Phase 1 pilot: arm finetuning.uk's about page with
-- the about-commercial-block (BUILT-BY LINE ONLY; for-sale and advertise stay
-- gated OFF — no Afternic listing confirmed, advertise.co.uk not built yet;
-- honesty rails in RUNBOOK "Honesty preconditions").
--
-- PRE-REQ: 202_about_commercial_block_component.sql applied (sole candidate).
-- Steps here: (1) write 'commercial' aspect (raw FACTS, supersede-then-insert,
-- mirrors HandleUpdateSiteSpec)  (2) append the section to the site_plan
-- aspect's about page (this site's authoritative store — it has NO site_plans
-- rows)  (3) mirror pages.sections cache  (4) flag ONLY the about page
-- needs_rebuild. Dispatch is a separate script (p1_trigger_rebuild.sh).
-- Idempotency: every step no-ops or aborts if already applied.
-- ============================================================================

BEGIN;

-- ── (1) commercial aspect: raw facts, built-by only ──
UPDATE site_specs SET is_current=false, superseded_at=NOW()
WHERE site_id=(SELECT id FROM sites WHERE domain='finetuning.uk')
  AND aspect='commercial' AND is_current=true;

INSERT INTO site_specs (site_id, aspect, data, source, created_by, is_current)
VALUES (
  (SELECT id FROM sites WHERE domain='finetuning.uk'),
  'commercial',
  '{
    "class": "portfolio",
    "tier": "2",
    "domain": "finetuning.uk",
    "for_sale_requested": false,
    "advertising_active": false,
    "inventory_open": false,
    "built_by_url": "https://fundamentallyai.com"
  }'::jsonb,
  'pilot-manual', 'uk', true
);

-- ── (2) append section to the site_plan aspect's about page ──
DO $$
DECLARE
  sid uuid := (SELECT id FROM sites WHERE domain='finetuning.uk');
  cur_id uuid;
  cur_data jsonb;
  about_idx int;
  about_sections jsonb;
BEGIN
  SELECT id, data INTO cur_id, cur_data
  FROM site_specs
  WHERE site_id=sid AND aspect='site_plan' AND is_current=true;
  IF cur_id IS NULL THEN
    RAISE EXCEPTION 'no current site_plan aspect for finetuning.uk — store layout changed, re-scout';
  END IF;

  SELECT (ord-1)::int, pg->'sections' INTO about_idx, about_sections
  FROM jsonb_array_elements(cur_data->'pages') WITH ORDINALITY arr(pg, ord)
  WHERE pg->>'name'='about';
  IF about_idx IS NULL THEN
    RAISE EXCEPTION 'no about page in finetuning.uk site_plan aspect';
  END IF;

  IF about_sections ? 'about-commercial-block' THEN
    RAISE NOTICE 'site_plan aspect already carries about-commercial-block — skipping aspect edit';
  ELSE
    -- supersede-then-insert a new current site_plan row with the section appended
    UPDATE site_specs SET is_current=false, superseded_at=NOW() WHERE id=cur_id;
    INSERT INTO site_specs (site_id, aspect, data, source, created_by, is_current)
    VALUES (
      sid, 'site_plan',
      jsonb_set(cur_data,
                ARRAY['pages', about_idx::text, 'sections'],
                about_sections || '"about-commercial-block"'::jsonb),
      'pilot-manual', 'uk', true
    );
  END IF;
END $$;

-- ── (3) mirror the pages.sections cache ──
UPDATE pages
SET sections = sections || '"about-commercial-block"'::jsonb
WHERE id='c0c68034-469f-420c-90bd-d3c0fc0e13d2'  -- finetuning.uk /about.html
  AND NOT (sections ? 'about-commercial-block');

-- ── (4) flag ONLY the about page ──
UPDATE pages SET build_status='needs_rebuild'
WHERE id='c0c68034-469f-420c-90bd-d3c0fc0e13d2';

-- ── verify before COMMIT ──
SELECT (SELECT data FROM site_specs
        WHERE site_id=(SELECT id FROM sites WHERE domain='finetuning.uk')
          AND aspect='commercial' AND is_current) AS commercial_facts,
       (SELECT pg->'sections' FROM site_specs ss,
             jsonb_array_elements(ss.data->'pages') pg
        WHERE ss.site_id=(SELECT id FROM sites WHERE domain='finetuning.uk')
          AND ss.aspect='site_plan' AND ss.is_current AND pg->>'name'='about') AS plan_about_sections,
       (SELECT sections FROM pages WHERE id='c0c68034-469f-420c-90bd-d3c0fc0e13d2') AS cache_sections,
       (SELECT build_status FROM pages WHERE id='c0c68034-469f-420c-90bd-d3c0fc0e13d2') AS about_status;

COMMIT;
