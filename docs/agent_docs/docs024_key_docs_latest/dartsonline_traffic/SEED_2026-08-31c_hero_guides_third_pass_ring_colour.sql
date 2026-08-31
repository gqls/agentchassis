-- SEED_2026-08-31c_hero_guides_third_pass_ring_colour.sql
--
-- ONE image, one axis. hero_guides only.
--
-- WHERE THE SECOND PASS GOT TO [MEASURED 2026-08-31 13:16Z, by eye + by pixel count]:
--   FIXED  the feathered archery flight — the dart now carries a flat translucent moulded
--          flight. This was the owner's actual complaint and it is resolved.
--   FIXED  the number ring: it now reads 20, 1, 18 clockwise. The previous file rendered a
--          "7" in the 1 position.
--   LOST   the board's GREEN. Counted over saturated pixels (S>0.35, V>0.20, hue 70–170°):
--            hero-guides v1  2,867 green px (1.9%)   ← had correct red/green rings
--            hero-guides v2      0 green px (0.0%)   ← rings render red/orange only
--          A board whose doubles and trebles rings carry no green is inaccurate in the same
--          way a feathered flight is, so this is not a cosmetic preference.
--
-- WHY THE PIXEL COUNT AND NOT MY EYE: the scene is lit by a warm desk lamp on dark wood, and
-- green under tungsten reads olive. "I cannot see any green" and "there is no green" are
-- different claims and only the second is checkable. The count could have come out either
-- way — it returned 1.9% on the previous file of the same subject under the same staging,
-- which is what makes 0.0% evidence rather than an impression.
--
-- WHAT CHANGES, AND WHAT DELIBERATELY DOES NOT. Only the colour clause is strengthened. The
-- flight and number-sequence clauses are left EXACTLY as they are, because they are the two
-- things that just started working and a re-roll can lose them (this run is itself the third
-- roll of this image; each one has traded one axis for another). If this pass regresses
-- either, the second-pass file is restorable — every version's S3 object is recorded:
--   v1 (feather, "7", green rings)  s3://personae-prod-uk001-images/images/system/20260831/83f3e56a-3cb0-4229-a950-89e30c8fdbec.png
--   v2 (good flight + "1", no green) s3://personae-prod-uk001-images/images/system/20260831/337869e4-c8d2-4722-82d7-4015d09b695c.png
--
-- The style guide is NOT touched. Its anatomy clauses went in at 13:13Z
-- (SEED_2026-08-31b) and already forbid an all-red bull; banana/provider.go's own comment
-- says a folded prohibition is "a softer instrument… honoured imperfectly", and that
-- lengthening it further on the evidence of ONE image is exactly the move that file warns
-- against ("verify by counting violations across a set, never on one image"). The positive
-- prompt is the instrument that has actually moved this image twice.
--
-- Site: dartsonline.com  5fe8785b-223d-41a3-88ee-c07187622381

BEGIN;

DO $$
DECLARE n int; p text;
BEGIN
  SELECT spi.prompt INTO p FROM site_plan_imagery spi
    JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current
   WHERE sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND spi.key = 'hero_guides';
  IF p IS NULL THEN
    RAISE EXCEPTION 'no current hero_guides plan row';
  END IF;
  -- The two clauses that are WORKING must be present before this file rewrites the prompt,
  -- so that a rewrite can never silently drop what the second pass bought.
  IF p NOT LIKE '%never a feather%' THEN
    RAISE EXCEPTION 'flight clause absent - the second pass prompt is not in place, do not overwrite';
  END IF;
  IF p NOT LIKE '%20, 1, 18, 4, 13, 6, 10, 15, 2, 17, 3, 19, 7, 16, 8, 11, 14, 9, 12, 5%' THEN
    RAISE EXCEPTION 'number-sequence clause absent - the second pass prompt is not in place, do not overwrite';
  END IF;

  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND item_type = 'needs_imagery'
     AND status NOT IN ('complete','cancelled','rejected','failed')
     AND spec->>'asset_key' = 'hero_guides';
  IF n <> 0 THEN
    RAISE EXCEPTION '% open needs_imagery item(s) already target hero_guides - would violate idx_swi_dedup', n;
  END IF;

  SELECT count(*) INTO n FROM assets
   WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND status = 'active'
     AND asset_key = 'hero_guides' AND locked_at IS NOT NULL;
  IF n <> 0 THEN
    RAISE EXCEPTION 'hero_guides asset is LOCKED; the run would report success and change nothing';
  END IF;
END $$;

-- The colour clause is moved OUT of the trailing accuracy sentence and into the scene
-- description, where the lighting is set — the previous version stated the colours only in
-- a list at the end, after "warm desk-lamp lighting" had already framed the palette as warm.
UPDATE site_plan_imagery spi
SET prompt =
'A contemplative overhead shot of a dartboard with a single dart and a measuring tape laid across it, alongside a notebook with hand-drawn diagrams of dart trajectories. Warm desk-lamp lighting on a dark wood surface, but the board itself is in full colour and its scoring rings read clearly: the doubles ring at the edge and the trebles ring halfway in are made of alternating BRIGHT RED and BRIGHT GREEN segments, and both colours are plainly visible, the green as green and not as olive or brown. Educational and thoughtful mood. Physical accuracy matters: the dart is a steel point, a knurled tungsten barrel, a slim shaft and a flat moulded plastic flight, never a feather and never an archery arrow; the board face is black and cream radial segments divided by a steel wire spider, the bull is a small red centre inside a green outer ring, and the numbers around the ring read 20, 1, 18, 4, 13, 6, 10, 15, 2, 17, 3, 19, 7, 16, 8, 11, 14, 9, 12, 5 clockwise from the top.'
FROM site_plans sp
WHERE sp.id = spi.plan_id AND sp.is_current
  AND sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND spi.key = 'hero_guides';

INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, summary, spec, priority,
   handler_agent, status, created_by, item_key, approval_mode)
SELECT
  sp.site_id, 'operator', 'build', 'needs_imagery', 'medium',
  'Regenerate hero_guides, third pass: board rings must carry visible green (owner 2026-08-31)',
  jsonb_build_object(
    'key', spi.key, 'kind', 'hero', 'check', 'emit_imagery_items',
    'scope', spi.scope, 'scope_ref', spi.scope_ref, 'prompt', spi.prompt,
    'purpose', 'hero', 'asset_key', spi.key,
    'style_hints', jsonb_build_object('aspect_ratio', '16:9'),
    'brand_update', false, 'original_pipeline', 'build'
  ),
  40, 'image-build-handler', 'triaged', 'dartsonline-traffic-2026-08-31c',
  'needs_imagery:' || spi.scope || ':' || COALESCE(spi.scope_ref,'-') || ':' || spi.key,
  'auto'
FROM site_plan_imagery spi
JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current
WHERE sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND spi.key = 'hero_guides';

DO $$
DECLARE n int; p text;
BEGIN
  SELECT spi.prompt INTO p FROM site_plan_imagery spi
    JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current
   WHERE sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND spi.key = 'hero_guides';
  -- the two working clauses must SURVIVE this rewrite
  IF p NOT LIKE '%never a feather%' THEN RAISE EXCEPTION 'flight clause lost by this rewrite'; END IF;
  IF p NOT LIKE '%20, 1, 18, 4, 13, 6, 10, 15, 2, 17, 3, 19, 7, 16, 8, 11, 14, 9, 12, 5%' THEN
    RAISE EXCEPTION 'number-sequence clause lost by this rewrite';
  END IF;
  IF p NOT LIKE '%BRIGHT GREEN%' THEN RAISE EXCEPTION 'colour clause did not land'; END IF;

  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
     AND created_by = 'dartsonline-traffic-2026-08-31c';
  IF n <> 1 THEN RAISE EXCEPTION 'expected 1 new work item, found %', n; END IF;
END $$;

COMMIT;
