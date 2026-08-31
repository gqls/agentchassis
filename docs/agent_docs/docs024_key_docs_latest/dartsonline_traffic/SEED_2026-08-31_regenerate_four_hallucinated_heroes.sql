-- SEED_2026-08-31_regenerate_four_hallucinated_heroes.sql
--
-- OWNER ASK (2026-08-31): the hero images on several pages have "hallucinated what darts and
-- dartboards and other objects look like". Regenerate them through the framework on the current
-- model setup.
--
-- WHAT THE MEASUREMENT ACTUALLY FOUND, because it is NOT what the ask assumed and it narrows the
-- job from "loads of pages" to four [MEASURED 2026-08-31]:
--
--   * Everything the site currently SERVES from an asset row is already Banana
--     (banana/gemini-3-pro-image-preview) and is GOOD. Verified by eye:
--       /assets/images/hero.jpg                    — real bristle board, correct wire spider,
--                                                    correct bull colours, knurled barrel. Accurate.
--       /assets/images/content-hero-grip-styles.jpg — four barrels, four genuinely distinct and
--                                                    correct grip patterns. Accurate.
--   * The BAD images are FOUR deployed files with NO ACTIVE ASSET ROW AT ALL — July-era artefacts
--     left in the repo after their asset rows were superseded:
--       hero-home.jpg         — no numbers, invented segment geometry, objects that are not darts
--       hero-new-arrivals.jpg — concentric rings not segments, FEATHERED FLIGHTS (those are
--                               archery arrows), blue segments (a board is red/green/black/cream)
--       hero-guides.jpg       — 4-panel; feathered flights throughout, garbled numbers
--       hero-sale.jpg         — same cohort, not individually inspected
--   * So the discriminator is mechanical and worth keeping: HAS an active asset row -> good;
--     NO asset row -> July SDXL-era leftover. `SELECT ... FROM assets WHERE url = '/assets/images/<f>'`
--     answers it without opening the image.
--
--   * The 8 `stability/stable-diffusion-xl-1024-v1-0` asset rows on this site are a RED HERRING:
--     their `url` is a SIGNED S3 link that expired 7 days after 2026-07-06 (X-Amz-Expires=604800),
--     none is referenced by any served page, and regenerating them would change nothing a visitor
--     sees. Left alone deliberately.
--
-- WHY REGENERATING NOW GETS THE GOOD MODEL — checked, not assumed:
--   bugs_closed/382 (fixed 2026-08-24, commit da21ae20f) flipped the routing default so an image
--   request with NO kind gets Banana instead of SDXL. PROVEN LIVE at the artefact today: the
--   running image-generator-adapter (v1.0.1349) reports build provenance ef06af0e0, and
--   `git merge-base --is-ancestor da21ae20f ef06af0e0` succeeds — with a control commit 5 ahead of
--   the build correctly NOT an ancestor, so the test could have failed.
--   Belt and braces anyway: every spec below carries `kind: 'hero'` EXPLICITLY, so routing decides
--   on the kind table (hero -> banana) and never needs the missing-kind default. The absent kind
--   was 382's actual cause.
--
-- WHY OPERATOR-INSERTED: both framework emitters skip plan rows that already have an active asset
-- (imageryplan.HasActiveAsset), so a re-request cannot come from a sweep. Same reason the agritec
-- lane hand-inserted its 17 on 2026-08-26; this seed follows that file's shape.
--
-- WHY IT IS SAFE: StoreAssetAction UPSERTs on (site_id, asset_key) WHERE status='active', so the
-- rows update in place with the same served paths and there is no window with a missing asset.
-- deploy_image_asset derives the filename from (asset_key, purpose) and REFUSES a caller-supplied
-- path, so hero_home necessarily lands back on /assets/images/hero-home.jpg — overwriting the bad
-- file rather than orphaning a second one. All four assets checked unlocked 2026-08-31.
--
-- The prompt is taken FROM THE PLAN ROW, never retyped here — site_plan_imagery is the authority
-- (RUNBOOK §12, the pages.sections lesson). If a prompt needs improving, amend the plan row and
-- re-run; do not edit the derived spec.
--
-- Site: dartsonline.com  5fe8785b-223d-41a3-88ee-c07187622381

BEGIN;

-- Pre-state assertions. Each names what it protects; a failure aborts before any insert.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_plan_imagery spi
    JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current
   WHERE sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
     AND spi.kind = 'hero'
     AND spi.key IN ('hero_home','hero_new_arrivals','hero_guides','hero_sale');
  IF n <> 4 THEN
    RAISE EXCEPTION 'expected 4 current hero plan rows, found %, - the plan is the authority and the specs are built from it', n;
  END IF;

  SELECT count(*) INTO n FROM assets
   WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND status = 'active'
     AND asset_key IN ('hero_home','hero_new_arrivals','hero_guides','hero_sale')
     AND locked_at IS NOT NULL;
  IF n <> 0 THEN
    RAISE EXCEPTION 'l% of the four target assets are LOCKED; StoreAssetAction refuses a locked row and the run would report success while changing nothing', n;
  END IF;

  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND item_type = 'needs_imagery'
     AND status NOT IN ('complete','cancelled','rejected','failed')
     AND spec->>'asset_key' IN ('hero_home','hero_new_arrivals','hero_guides','hero_sale');
  IF n <> 0 THEN
    RAISE EXCEPTION '% open needs_imagery item(s) already target these keys - inserting would violate idx_swi_dedup', n;
  END IF;
END $$;

INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, summary, spec, priority,
   handler_agent, status, created_by, item_key, approval_mode)
SELECT
  sp.site_id,
  'operator',
  'build',
  'needs_imagery',
  'medium',
  'Regenerate hallucinated hero on the current model setup (owner 2026-08-31): ' || spi.key,
  jsonb_build_object(
    'key',               spi.key,
    'kind',              'hero',                 -- EXPLICIT: absent kind was bugs_closed/382's cause
    'check',             'emit_imagery_items',
    'scope',             spi.scope,
    'scope_ref',         spi.scope_ref,
    'prompt',            spi.prompt,             -- from the plan row, the authority; never retyped
    'purpose',           'hero',
    'asset_key',         spi.key,
    'style_hints',       jsonb_build_object('aspect_ratio', '16:9'),
    'brand_update',      false,
    'original_pipeline', 'build'
  ),
  40,
  'image-build-handler',
  'triaged',
  'dartsonline-traffic-2026-08-31',
  'needs_imagery:' || spi.scope || ':' || COALESCE(spi.scope_ref,'-') || ':' || spi.key,
  'auto'
FROM site_plan_imagery spi
JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current
WHERE sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND spi.kind = 'hero'
  AND spi.key IN ('hero_home','hero_new_arrivals','hero_guides','hero_sale');

SELECT item_key, status, spec->>'kind' AS kind, spec->>'asset_key' AS asset_key
FROM site_work_items
WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND created_by='dartsonline-traffic-2026-08-31'
ORDER BY item_key;

COMMIT;
