-- SQL — dartsonline.com identity truth reset (owner decision D4, 2026-07-29)
--
-- WHY: the current `identity` aspect carries another company's facts. Verified
-- 2026-07-29 against the live row:
--   contact.email    = sales@darts.com                       (darts.com — a different company)
--   contact.phone    = (800) 526-1920                        (same)
--   contact.address  = 13010 NE David Cir, Portland, US      (same)
--   contact.location = Portland, Oregon, USA                 (same)
--   services[1]      = "Stocking major brands such as Red Dragon, Winmau, Harrows,
--                       Shot, Unicorn, Target, and Mission"  (no such relationships exist)
--   social_profiles.facebook = facebook.com/dartsonlineau    (the AUSTRALIAN company's page)
--   about_summary    = openly derived from "the related Australian entity"
-- `briefing.gaps` already admitted the borrowing; nobody acted on it. An affiliate
-- network verifies a site before accepting it, so this had to go first.
--
-- The real details come from the `sites` row (email darts@contactforsales.com,
-- phone 07934 524 911) — the ones the owner actually submitted the domain with.
--
-- SHAPE NOTE (bug 072 class, verified in code 2026-07-29): the contact WRITER nests
-- under `contact` (sync_site_identity_action.go:103-110) while one READER looks for
-- FLAT keys (validate_page_content.go:1281-1288 reads identity.data->>'email' and
-- ->>'contact_email'). So write BOTH shapes. sites.email is first in that COALESCE
-- and is already correct, so this is belt-and-braces, not the load-bearing fix.
--
-- ORDERING IS FORCED: idx_site_specs_current is UNIQUE on (site_id, aspect)
-- WHERE is_current — supersede must commit before the insert.
--
-- COPY RAILS this establishes (binding on every page rebuilt afterwards):
--   MAY say: specialist darts site; spec-first buying guides; setup help; UK online-only.
--   MAY NOT say: "we stock/carry"; named brand relationships; any address;
--                founding history; shipping/warehouse claims.

BEGIN;

-- 1. Backup (the whole aspect history, not just the current row)
CREATE TABLE IF NOT EXISTS bak_darts_identity_20260729 AS
SELECT * FROM site_specs
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND aspect = 'identity';

-- 2. Supersede the current row
UPDATE site_specs
SET is_current = false, superseded_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND aspect = 'identity'
  AND is_current = true;

-- 3. Insert the corrected row
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current, created_by, notes)
VALUES (
  '5fe8785b-223d-41a3-88ee-c07187622381',
  'identity',
  jsonb_build_object(
    'company_name', 'Darts Online',
    'tagline', 'Spec-first darts buying guides and setup advice',
    'industry', 'Darts Equipment & Advice',
    'sub_industry', 'Specialist darts guidance and product curation',
    -- flat keys for the validate_page_content reader (bug 072 shape)
    'email', 'darts@contactforsales.com',
    'phone', '07934 524 911',
    'contact_email', 'darts@contactforsales.com',
    'contact_phone', '07934 524 911',
    -- nested object for sync_site_identity_action
    'contact', jsonb_build_object(
      'email', 'darts@contactforsales.com',
      'phone', '07934 524 911',
      'location', 'United Kingdom (online only)'
    ),
    'about_summary', 'Darts Online is a UK-based specialist darts website. It publishes '
      || 'spec-first buying guides — tungsten percentage, barrel weight and profile, shaft '
      || 'length, flight shape — and helps players choose a setup that suits how they throw. '
      || 'It is an online-only publication: it does not hold stock, run a warehouse, or have '
      || 'a retail premises.',
    'services', jsonb_build_array(
      jsonb_build_object(
        'name', 'Buying guides',
        'description', 'Spec-first guides to barrels, shafts, flights and boards, written for '
          || 'both first-set buyers and club players'
      ),
      jsonb_build_object(
        'name', 'Setup guidance',
        'description', 'Help choosing a barrel, shaft and flight combination as a complete setup'
      ),
      jsonb_build_object(
        'name', 'Darts news and analysis',
        'description', 'Tournament news gathered from published sources, with gear-led analysis'
      )
    ),
    'target_audience', 'Darts players at all skill levels — from casual pub players buying '
      || 'their first set, to committed club and league competitors seeking specific tungsten '
      || 'weights and custom setups',
    'unique_selling_points', jsonb_build_array(
      'Specialist darts focus — not a general sporting goods site',
      'Spec-first guidance: weights, tungsten percentages and barrel profiles stated plainly',
      'Written for both first-set buyers and club players without patronising either',
      'Memorable .com address'
    ),
    'key_people', '[]'::jsonb,
    'social_profiles', '{}'::jsonb,
    'competitors_found', jsonb_build_array('dartscorner.com', 'darts.com', 'dartsonline.com.au'),
    'has_existing_site', true,
    'existing_site_quality', 'in_development'
  ),
  'authored',
  'dartsonline-traffic-workstream',
  true,
  'dartsonline-traffic-workstream',
  'Truth reset per owner decision D4 (2026-07-29). REMOVED: darts.com contact details '
    || '(Portland address, US phone, sales@darts.com), the brand-stocking service entry, the '
    || 'dartsonline.com.au Facebook profile, and the pricing/EOFY and brand-portfolio USPs — '
    || 'all belonged to other companies. Previous rows preserved in bak_darts_identity_20260729 '
    || 'and in the aspect history. has_existing_site flipped false->true: the site IS live now.'
);

COMMIT;

-- 4. Verify
SELECT is_current, source_agent, created_at,
       data->'contact'->>'email'  AS nested_email,
       data->>'email'             AS flat_email,
       data->'contact' ? 'address' AS still_has_address,
       data::text ILIKE '%Portland%'   AS mentions_portland,
       data::text ILIKE '%Red Dragon%' AS mentions_stocking,
       data::text ILIKE '%dartsonlineau%' AS mentions_au_facebook
FROM site_specs
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND aspect = 'identity'
ORDER BY created_at DESC;
