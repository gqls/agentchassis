-- 560_aiao_bind_case_study_card_images.sql
--
-- ai-agent-orchestration.com — point the `case-studies-grid` cards at the images
-- that now exist. Closes the imagery half of the owner's ask; the carousel half
-- shipped as 559 and these are the cards it scrolls.
--
-- THE STATE THIS FIXES, measured 2026-08-22 before generation:
--
--   index                           5 × <img src="">        (no cardN_image_url key at all)
--   enterprise-reference-deployment 5 × /assets/images/case-study-*.png, ALL HTTP 404
--
-- ⚠ THE HANDOFF SAID "there is no `cardN_image_url` key at all". That is true of
-- `index` ONLY. `enterprise-reference-deployment` HAS all five keys and they point
-- at files that have never existed. Two different faults wearing one symptom, and
-- a fix aimed only at the missing-key half would have left five 404s in place.
--
-- ⚠ AND THE OLD URLs COULD NEVER HAVE COME GOOD, whatever was generated: they end
-- `.png`, and this purpose serves `.jpg`. The served path is DERIVED, not stored —
-- `storage.DeployedWebPath(asset_key, purpose)` → `DeployedAssetPath` →
-- `AssetKeyFilename(assetKey, ext)`, which lowercases nothing but does
-- `strings.ReplaceAll(assetKey, "_", "-")` and appends the extension from
-- `ImagePurposes[purpose]`. For `content_hero` that extension is **jpg**
-- (`url_helpers.go:363`). Confirmed at the artefact rather than reasoned about:
-- every `<key>.jpg` returns 200 and every `<key>.png` returns 404.
--
-- WHY `content_hero` IS THE RIGHT PURPOSE for a card image: its geometry is
-- 1600×900 and the comment at `url_helpers.go:360-363` says the card derivation
-- cover-crops it to 800×450. `icon` (240×240) would have been the wrong shape.
--
-- HOW THE IMAGES WERE MADE. Nine `needs_imagery` work items → `image-build-handler`
-- (the live path: **62** completions fleet-wide as of 2026-08-22, most recent the day before this file).
-- **NOT** `image-url-404-handler` or `image-source-unsatisfiable-handler`: both are
-- live and both look like the obvious owner of these rows, but their workflows are
-- `query_database → create_work_item → checkpoint_for_review` — they TRIAGE, they do
-- not generate, and the only site they had ever run against as of 2026-08-22 now has zero `<img>`
-- tags in any component. Routing these rows there would most likely have stripped
-- the cards.
--
-- THE PROMPTS ARE THE FRAMEWORK'S OWN WORDS. Each prompt is the card's existing
-- `cardN_image_alt` — prose the pipeline itself wrote, describing the intended
-- diagram — plus a house style clause taken from `design_intent.imagery_direction`
-- as migration 458 left it ("technical illustrations and architectural diagrams are
-- the default … clean vector work in the dark palette"). Per the owner's ruling of
-- 2026-08-06 the framework writes the content, not the session; reusing the alt
-- text is what keeps that true here.
--
-- ⚠ IMAGERY POLICY (458) IS SATISFIED BY SUBJECT, NOT BY LUCK. Every prompt asks
-- for an abstract architectural diagram. None depicts a person, so the ruling's one
-- hard rule — never present a photographed person as a member of this company —
-- cannot be engaged. The `.member-icon` slot the ruling warns about belongs to
-- `departments-grid`/`leadership-team` and is NOT touched here.
--
-- NINE IMAGES FOR TEN SLOTS: `index` card4 and `enterprise-reference-deployment`
-- card5 are the same subject (distributed tracing across a hierarchical topology),
-- so they share one asset rather than generating two near-identical diagrams.
--
-- NO PRE-SIGNED URLs. Every value written here is a stable site-relative
-- `/assets/images/…` path. The site's **9** `hero`/`content_hero` rows in `assets` (as of 2026-08-22) hold
-- expiring Backblaze links (`X-Amz-Expires=604800`, stamped 2026-08-11, lapsed
-- 2026-08-18); `assets.url` must never be copied into content.
--
-- DOES NOT RE-RENDER. Propagate with a page-scoped `template_changed` rerender
-- (RUNBOOK R8), then verify every src over HTTP.
--
-- ROLLBACK: 560_aiao_bind_case_study_card_images_ROLLBACK.sql

BEGIN;

INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '560_aiao_bind_case_study_card_images', 'page_components', pc.id::text,
       jsonb_build_object('content_data', pc.content_data),
       'pre-560 content_data for ' || p.name || '/' || pc.slot_name
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND pc.slot_name='case-studies-grid';

UPDATE page_components pc
SET content_data = coalesce(pc.content_data,'{}'::jsonb) || jsonb_build_object(
      'card1_image_url','/assets/images/case-study-department-orchestration.jpg',
      'card2_image_url','/assets/images/case-study-blast-radius-containment.jpg',
      'card3_image_url','/assets/images/case-study-kafka-consumer-groups.jpg',
      'card4_image_url','/assets/images/case-study-distributed-tracing.jpg',
      'card5_image_url','/assets/images/case-study-ordered-task-chains.jpg'),
    updated_at = now()
FROM pages p
WHERE p.id=pc.page_id AND p.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND p.name='index' AND pc.slot_name='case-studies-grid';

UPDATE page_components pc
SET content_data = coalesce(pc.content_data,'{}'::jsonb) || jsonb_build_object(
      'card1_image_url','/assets/images/case-study-financial-pipeline.jpg',
      'card2_image_url','/assets/images/case-study-multi-region-dispatch.jpg',
      'card3_image_url','/assets/images/case-study-healthcare-audit.jpg',
      'card4_image_url','/assets/images/case-study-web-automation-scaling.jpg',
      'card5_image_url','/assets/images/case-study-distributed-tracing.jpg'),
    updated_at = now()
FROM pages p
WHERE p.id=pc.page_id AND p.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND p.name='enterprise-reference-deployment' AND pc.slot_name='case-studies-grid';

DO $$
DECLARE
  backed_up int; bound int; pngs int; presigned int; alts int;
BEGIN
  SELECT count(*) INTO backed_up FROM migration_backups
   WHERE migration_name='560_aiao_bind_case_study_card_images';
  IF backed_up <> 2 THEN
    RAISE EXCEPTION '560: expected 2 backup rows, wrote %', backed_up;
  END IF;

  -- Ten slots, all bound.
  SELECT count(*) INTO bound
    FROM page_components pc JOIN pages p ON p.id=pc.page_id,
         LATERAL jsonb_object_keys(pc.content_data) k
   WHERE p.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
     AND pc.slot_name='case-studies-grid' AND k ~ '^card[1-5]_image_url$'
     AND pc.content_data->>k <> '';
  IF bound <> 10 THEN
    RAISE EXCEPTION '560: expected 10 bound card image urls, found %', bound;
  END IF;

  -- No `.png` may survive: that extension is what this purpose does NOT serve.
  SELECT count(*) INTO pngs
    FROM page_components pc JOIN pages p ON p.id=pc.page_id,
         LATERAL jsonb_object_keys(pc.content_data) k
   WHERE p.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
     AND pc.slot_name='case-studies-grid' AND k ~ '^card[1-5]_image_url$'
     AND pc.content_data->>k LIKE '%.png';
  IF pngs <> 0 THEN
    RAISE EXCEPTION '560: % card url(s) still end .png — content_hero serves .jpg, so those are 404s', pngs;
  END IF;

  -- Never an expiring link in content.
  SELECT count(*) INTO presigned
    FROM page_components pc JOIN pages p ON p.id=pc.page_id,
         LATERAL jsonb_object_keys(pc.content_data) k
   WHERE p.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
     AND pc.slot_name='case-studies-grid' AND k ~ '^card[1-5]_image_url$'
     AND (pc.content_data->>k ILIKE '%X-Amz-%' OR pc.content_data->>k ILIKE 'http%');
  IF presigned <> 0 THEN
    RAISE EXCEPTION '560: % card url(s) are absolute/pre-signed; only site-relative /assets/images paths may be published', presigned;
  END IF;

  -- The alt text is the accessibility surface AND was the prompt: it must survive.
  SELECT count(*) INTO alts
    FROM page_components pc JOIN pages p ON p.id=pc.page_id,
         LATERAL jsonb_object_keys(pc.content_data) k
   WHERE p.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
     AND pc.slot_name='case-studies-grid' AND k ~ '^card[1-5]_image_alt$'
     AND coalesce(pc.content_data->>k,'') <> '';
  IF alts <> 10 THEN
    RAISE EXCEPTION '560: expected 10 non-empty card image alts, found % — the bind must not disturb them', alts;
  END IF;

  RAISE NOTICE '560 OK: 10 card slots bound to 9 stable /assets/images/*.jpg paths, 0 .png, 0 absolute urls, 10 alts intact. Re-render both pages, then verify every src over HTTP.';
END $$;

COMMIT;
