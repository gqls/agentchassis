-- SQL — dartsonline.com BRIEFING truth reset (owner decision D4, 2026-07-29)
--
-- Companion to SQL_2026-07-29_identity_truth_reset.sql. The `briefing` aspect is the
-- worse of the two because it is what the page writer actually renders an About page
-- from. Verified live 2026-07-29, the current row contained:
--
--   about_us      : "…stocking a deep catalogue … from all major brands — Red Dragon,
--                    Winmau, Harrows, Target, Mission, Shot, and Unicorn … we carry the
--                    range … Based in Portland, Oregon, we serve players at every level…"
--   headquarters  : "13010 NE David Cir, Portland, Oregon 97230, United States"
--   contact_email : "sales@darts.com"
--   contact_phone : "(800) 526-1920"
--   services[].description: the seven brands again, as "a deep multi-brand catalogue"
--
-- The same row's own `gaps` array ALREADY recorded that all of this was borrowed:
--   "contact_phone — number sourced from associated Portland operation; not confirmed"
--   "confirmed brand partnership or stock relationship details beyond signal-level inference"
-- So the research was honest about its uncertainty and the build shipped the claims
-- anyway. That gap between "recorded as unverified" and "rendered as fact" is the
-- transferable lesson; see NOTES_dartsonline_traffic.md.
--
-- What replaces it: the same voice (enthusiast-to-enthusiast, second person, contractions
-- — content_direction is untouched and still good), making only claims that are true of a
-- UK online-only specialist darts publication.

BEGIN;

CREATE TABLE IF NOT EXISTS bak_darts_briefing_20260729 AS
SELECT * FROM site_specs
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND aspect = 'briefing';

UPDATE site_specs
SET is_current = false, superseded_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND aspect = 'briefing'
  AND is_current = true;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current, created_by, notes)
SELECT
  '5fe8785b-223d-41a3-88ee-c07187622381',
  'briefing',
  -- start from the superseded row so nothing unrelated is lost, then overwrite the
  -- fabricated keys and drop the ones that have no truthful replacement
  (b.data
    - 'headquarters'
    - 'location')
  || jsonb_build_object(
    'contact_email', 'darts@contactforsales.com',
    'contact_phone', '07934 524 911',
    'tagline', 'Spec-first darts buying guides and setup advice',
    'about_us', 'Darts Online is for players who want to know exactly what they''re throwing '
      || 'and why. We publish spec-first buying guides — tungsten percentage, barrel weight '
      || 'and profile, shaft length, flight shape — and we explain what each one actually '
      || 'changes when the dart leaves your hand. Whether you''re picking up your first set '
      || 'for the pub or fine-tuning a setup for league night, you''ll find the specs stated '
      || 'plainly and the trade-offs explained without the sales patter. We''re not a general '
      || 'sporting goods site that happens to mention darts. Darts is all we do.',
    'services', jsonb_build_array(
      jsonb_build_object(
        'name', 'Buying guides',
        'description', 'Spec-first guides to barrels, shafts, flights and boards. We state '
          || 'weights, tungsten percentages and profiles plainly, and explain what each '
          || 'choice changes about your throw.'
      ),
      jsonb_build_object(
        'name', 'Setup guidance',
        'description', 'Barrel, shaft and flight work as one system. Our guidance treats '
          || 'them that way, so you can pick a combination that suits your grip and release.'
      ),
      jsonb_build_object(
        'name', 'Darts news and analysis',
        'description', 'Tournament news gathered from published darts sources, with '
          || 'gear-led analysis that connects what happened on stage to what you throw.'
      )
    ),
    'honesty_rails', jsonb_build_array(
      'Never claim to stock, carry, hold or ship products — this site holds no inventory',
      'Never name a brand relationship, partnership or stockist arrangement',
      'Never state a business address, headquarters or founding history',
      'Never make delivery, returns or warehouse claims',
      'UK online-only; contact is darts@contactforsales.com / 07934 524 911'
    )
  ),
  'authored',
  'dartsonline-traffic-workstream',
  true,
  'dartsonline-traffic-workstream',
  'Truth reset per owner decision D4 (2026-07-29). REMOVED from about_us/services: the '
    || 'seven named brands, "stocking"/"we carry" claims, and the Portland HQ; REMOVED keys: '
    || 'headquarters, location; contact replaced with the real submission details. The row''s '
    || 'own `gaps` array had already flagged every one of these as unverified. Added '
    || 'honesty_rails so a future writer/planner reading this aspect sees the constraints '
    || 'inline rather than having to find this note. Prior rows in bak_darts_briefing_20260729.'
FROM site_specs b
WHERE b.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND b.aspect = 'briefing'
  AND b.is_current = false
ORDER BY b.created_at DESC
LIMIT 1;

COMMIT;

-- Verify: new row clean, old row preserved and still dirty
SELECT is_current, source_agent, created_at::date,
       data->>'contact_email' AS email,
       data ? 'headquarters'  AS has_hq,
       data::text ILIKE '%Portland%'    AS portland,
       data::text ILIKE '%Red Dragon%'  AS brands,
       data::text ~* '(we |)(stock|carry)(ing|)' AS stock_words
FROM site_specs
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND aspect = 'briefing'
ORDER BY created_at DESC;
