-- bugs_open/027 §4b follow-up — bring the three authored BASE VOICES under the
-- 200-char direction cap, so photographic kinds keep palette AND prose.
--
-- WHY. The §4b code fix (1191cecdb, live v1.0.1144) composes the palette FIRST
-- (imagery_style_guide.go:136), so an over-cap direction now loses its prose
-- tail (medium/mood) instead of its colours. Correct trade, but the four base
-- voices composed to 304–398 chars (measured live 2026-07-24), so every
-- photographic-kind generation on those sites trips the truncation WARN and
-- flips from prose-without-colours to colours-without-prose. The recorded
-- remedy (bugs_open/027 fix trail; imagery README_where_we_are 2026-07-21):
-- shorten the palette glosses in config — "I'll do the three I wrote;
-- robot-hands' is yours." This is that migration. robot-hands (398) is the
-- owner's guide and is DELIBERATELY untouched, as is its 233-char
-- content_hero override (a passed-gate testbed).
--
-- Scope: ROOT medium/mood/palette only. The kinds.content_hero overrides are
-- already under cap (139–147, SQL_2026-07-19_style_guides_terse_directions.sql)
-- and are not touched; neither is any `avoid` list.
--
-- Composed lengths after this migration (palette-first order, verified below):
--   finetuning.uk 196 · gamesdesign.co.uk 189 · leopardessconsulting.co.uk 190
--
-- Each UPDATE is needle-gated on the row's current `mood` text: if another
-- session has already changed the voice, the needle misses, the UPDATE reports
-- UPDATE 0, and this migration must be re-derived rather than applied blind.
--
-- Applied: 2026-07-24.

\set ON_ERROR_STOP on
BEGIN;

CREATE TABLE IF NOT EXISTS site_specs_imagery_guide_backup_20260724 AS
SELECT * FROM site_specs WHERE aspect = 'imagery_style_guide';

-- finetuning.uk — was 304. Accent first; systems/data-flow meaning kept.
UPDATE site_specs SET
  data = data || $j$
  {
    "medium": "abstract geometric illustration of network patterns and data flow",
    "mood": "calm, precise, trustworthy",
    "palette": "electric teal on deep navy and charcoal, cool blue accents, warm off-white highlights"
  }
  $j$::jsonb,
  updated_at = now()
 WHERE aspect='imagery_style_guide' AND is_current=true
   AND site_id = (SELECT id FROM sites WHERE domain='finetuning.uk')
   AND data->>'mood' = 'calm, competent, purposeful, precise, trustworthy';

-- gamesdesign.co.uk — was 352. Single cyan accent first with its hex; HUD/schematic kept.
UPDATE site_specs SET
  data = data || $j$
  {
    "medium": "flat technical diagram illustration, schematic HUD linework",
    "mood": "precise, utilitarian, restrained",
    "palette": "single cyan accent #00bcd4 on near-black #121212, light grey #e0e0e0 secondary"
  }
  $j$::jsonb,
  updated_at = now()
 WHERE aspect='imagery_style_guide' AND is_current=true
   AND site_id = (SELECT id FROM sites WHERE domain='gamesdesign.co.uk')
   AND data->>'mood' = 'precise, systems-oriented, utilitarian, confident, restrained';

-- leopardessconsulting.co.uk — was 305. Gold-on-charcoal first; same-hand-as-logo kept.
UPDATE site_specs SET
  data = data || $j$
  {
    "medium": "flat diagrammatic illustration, drawn in the same hand as the logo",
    "mood": "explanatory, calm, editorial",
    "palette": "antique gold #836E32 on near-black #0D0D0D, warm off-white #FAF8F4 sparingly"
  }
  $j$::jsonb,
  updated_at = now()
 WHERE aspect='imagery_style_guide' AND is_current=true
   AND site_id = (SELECT id FROM sites WHERE domain='leopardessconsulting.co.uk')
   AND data->>'mood' = 'explanatory, calm, precise, editorial, quietly authoritative';

-- Verify: composed BASE voice per site in the LIVE order (palette FIRST since
-- §4b, imagery_style_guide.go:136) against the 200 cap. robot-hands stays over
-- by design (owner's guide).
SELECT s.domain,
       length('colour palette: ' || (sd.data->>'palette') || '. ' ||
              (sd.data->>'medium') || '. ' || (sd.data->>'mood')) AS composed_len,
       CASE WHEN length('colour palette: ' || (sd.data->>'palette') || '. ' ||
              (sd.data->>'medium') || '. ' || (sd.data->>'mood')) <= 200
            THEN 'under cap' ELSE 'OVER CAP' END AS verdict
  FROM site_specs sd JOIN sites s ON s.id=sd.site_id
 WHERE sd.aspect='imagery_style_guide' AND sd.is_current=true
 ORDER BY s.domain;

COMMIT;
