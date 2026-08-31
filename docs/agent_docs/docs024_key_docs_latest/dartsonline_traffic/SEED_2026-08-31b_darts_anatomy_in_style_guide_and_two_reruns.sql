-- SEED_2026-08-31b_darts_anatomy_in_style_guide_and_two_reruns.sql
--
-- WHY THIS EXISTS. The four heroes regenerated at 12:35Z today
-- (SEED_2026-08-31_regenerate_four_hallucinated_heroes.sql) all landed, in place, on the
-- current model. I then LOOKED at all four served files [MEASURED 2026-08-31 13:0xZ]:
--
--   hero-new-arrivals.jpg  GOOD. Flat four-wing moulded flights, correct brass knurled
--                          barrels, black stems. The archery-fletching defect the owner
--                          named is genuinely gone here.
--   hero-sale.jpg          GOOD. Macro of a steel barrel, correct machined grip rings,
--                          plausible "24g" weight stamp.
--   hero-home.jpg          STILL WRONG in one respect: the bull is ENTIRELY RED. On a real
--                          board the inner bull is red and the outer bull ring around it is
--                          GREEN. Everything else (wire spider, black/cream segments,
--                          knurled barrels) is right.
--   hero-guides.jpg        STILL WRONG in two respects: (a) the dart carries a FEATHERED,
--                          fletched flight — visible barbs and quill, i.e. an archery arrow,
--                          which is the exact defect the owner complained about; (b) the
--                          number clockwise of 20 renders as "7" where a real board reads
--                          "1". The rest of the board is correct (colours, rings, bull,
--                          spider, and the other nine legible numbers are in true sequence).
--
-- So 2 of 4 are clean and 2 carry residual anatomy errors. Both live pages:
-- hero-home.jpg on "/" and hero-guides.jpg on "/guides/index.html" (both confirmed served).
--
-- THE CAUSE IS IN THE PROMPTS, AND IT IS LEGIBLE:
--   * hero_home's plan prompt says the barrels catch light "against the deep black and red
--     of the board". It names TWO board colours and neither is green or cream. The model did
--     what it was told and produced a board with no green anywhere — including the bull.
--     (Same prompt asks for darts "embedded in the triple-twenty segment"; the image put
--     them in the bull, so brief adherence is imperfect independently of colour.)
--   * hero_guides' plan prompt says only "a single dart". Nothing in it, and nothing in the
--     style guide, says what a dart flight is. A feather is then a free choice, and the
--     image's vintage desk-lamp staging makes a feather a plausible one for the model.
--
-- THE STRUCTURAL GAP THIS CLOSES. The site's imagery_style_guide is rich, but every clause
-- in it governs COMPOSITION, PALETTE or COMMERCIAL CLAIM — grounds, collages, logos,
-- packaging, faces. Not one clause asserts what the product this site sells actually LOOKS
-- LIKE. A negative style guide that never states subject anatomy leaves the model free to
-- hallucinate the subject, which is precisely the owner's complaint, and re-rolling the
-- image does not fix it because nothing in the request ever said otherwise.
--
-- ⚠⚠ WHERE THE CLAUSE HAS TO GO, AND WHY THE OBVIOUS PLACE IS WRONG [MEASURED 2026-08-31].
-- Do NOT add this to the guide-level `avoid`. A per-kind override REPLACES the guide-level
-- field WHOLESALE, verified at the code, not the comment:
--
--     platform/orchestration/actions/imagery_style_guide.go, avoidForKind:
--         if o, ok := g.Kinds[kind]; ok { return o.Avoid }
--         return g.Avoid
--
-- This site HAS a `kinds.hero` override. So for kind='hero' the entire guide-level avoid
-- list — "text, lettering, numerals or watermarks", "multi-panel collages", "manufacturer
-- logos", all of it — is DROPPED, and a hero is governed by the override's three clauses
-- alone. The site-level `mood` is dropped for heroes too (directionForKind composes
-- o.Medium/o.Mood/o.Palette and the hero override declares no mood).
-- A clause added at guide level would therefore have changed NOTHING for heroes, while
-- looking exactly like a fix — and the next re-roll would have "confirmed" it.
-- So the anatomy clauses go in `kinds.hero` and `kinds.content_hero`, appended to what is
-- already there rather than replacing it.
--
-- WHY THE PROMPTS CARRY THE POSITIVE HALF AND THE GUIDE CARRIES ONLY PROHIBITIONS:
--   * The composed style DIRECTION is capped at maxImageryDirectionInPrompt = 200 chars
--     (generate_image_actions.go:1098) and the hero direction already spends ~129 of it, so
--     anatomy prose in `medium` would be truncated and could evict the palette
--     (bugs_open/027 §4b). `medium` is therefore left untouched.
--   * The `avoid` path has NO cap: banana/provider.go folds it into the positive prompt
--     AFTER the direction cap has been applied — "nothing this function adds can evict
--     anything". So prohibitions are safe to lengthen.
--   * But that same file warns a folded prohibition "is a softer instrument than SDXL's true
--     negative conditioning" and Gemini "honours it imperfectly". So the reliable, positive
--     statement of anatomy goes in the plan-row PROMPT, which is uncapped and positive, and
--     the guide carries the durable negative backstop for every future hero on this site.
--   * Prohibitions are kept PURELY negative. A "must not contain" list that embeds positive
--     facts ("the outer bull is green") risks the model reading the fact as the thing to
--     avoid. Positive facts live in the prompt only.
--
-- Per SEED_2026-08-31's own rule and RUNBOOK §12: the plan row is the authority, so the
-- prompt is amended THERE and the spec is rebuilt FROM the live row, never retyped.
--
-- Site: dartsonline.com  5fe8785b-223d-41a3-88ee-c07187622381
-- Plan: 0fb05b75-04f4-4f4c-8890-c34d6a71012c (current)

BEGIN;

-- ── 0. backup of exactly the rows this file rewrites ──────────────────────────
CREATE TABLE IF NOT EXISTS bak_20260831b_site_plan_imagery AS
SELECT spi.* FROM site_plan_imagery spi
JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current
WHERE sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND spi.key IN ('hero_home','hero_guides');

CREATE TABLE IF NOT EXISTS bak_20260831b_site_specs AS
SELECT * FROM site_specs
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND aspect = 'imagery_style_guide' AND is_current;

-- ── 1. pre-state assertions. Each names what it protects. ─────────────────────
DO $$
DECLARE n int; g jsonb;
BEGIN
  SELECT count(*) INTO n FROM site_plan_imagery spi
    JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current
   WHERE sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
     AND spi.key IN ('hero_home','hero_guides') AND spi.locked_at IS NULL;
  IF n <> 2 THEN
    RAISE EXCEPTION 'expected 2 unlocked target plan rows, found % - a locked row would be silently un-amendable', n;
  END IF;

  SELECT data INTO g FROM site_specs
   WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
     AND aspect = 'imagery_style_guide' AND is_current;
  IF g IS NULL THEN
    RAISE EXCEPTION 'no current imagery_style_guide - this file APPENDS to the live row and must not invent one';
  END IF;
  -- The override keys must exist, because jsonb_set on a missing path is a SILENT NO-OP.
  -- Without this assertion the whole point of the file could fail while reporting success.
  IF g #> '{kinds,hero,avoid}' IS NULL THEN
    RAISE EXCEPTION 'kinds.hero.avoid absent - jsonb_set would no-op and the hero clause would never ship';
  END IF;
  IF g #> '{kinds,content_hero,avoid}' IS NULL THEN
    RAISE EXCEPTION 'kinds.content_hero.avoid absent - jsonb_set would no-op and the content_hero clause would never ship';
  END IF;

  SELECT count(*) INTO n FROM assets
   WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND status = 'active'
     AND asset_key IN ('hero_home','hero_guides') AND locked_at IS NOT NULL;
  IF n <> 0 THEN
    RAISE EXCEPTION '% target asset(s) LOCKED; StoreAssetAction refuses a locked row and the run would report success while changing nothing', n;
  END IF;

  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND item_type = 'needs_imagery'
     AND status NOT IN ('complete','cancelled','rejected','failed')
     AND spec->>'asset_key' IN ('hero_home','hero_guides');
  IF n <> 0 THEN
    RAISE EXCEPTION '% open needs_imagery item(s) already target these keys - inserting would violate idx_swi_dedup', n;
  END IF;
END $$;

-- ── 2. the anatomy clauses, appended to the two overrides that govern them ────
-- Built FROM the live row by jsonb_set, so nothing already in the guide is retyped and
-- nothing can be lost in transcription. Captured first because idx_site_specs_current is a
-- UNIQUE partial index: two current rows cannot coexist even inside one transaction.
CREATE TEMP TABLE _new_guide ON COMMIT DROP AS
SELECT jsonb_set(
         jsonb_set(
           data,
           '{kinds,hero,avoid}',
           to_jsonb((data #>> '{kinds,hero,avoid}') ||
             '; feathered or fletched flights of any kind — archery fletching on a dart; an all-red bull with no green outer ring; dartboard numbers out of their true sequence, repeated, invented or garbled; board segments coloured blue, purple, orange or anything other than black and cream; concentric rings in place of radial segments; darts drawn as arrows')
         ),
         '{kinds,content_hero,avoid}',
         to_jsonb((data #>> '{kinds,content_hero,avoid}') ||
           '; feathered or fletched flights of any kind — archery fletching on a dart; an all-red bull with no green outer ring; dartboard numbers out of their true sequence, repeated, invented or garbled; board segments coloured blue, purple, orange or anything other than black and cream; concentric rings in place of radial segments; darts drawn as arrows')
       ) AS data
FROM site_specs
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND aspect = 'imagery_style_guide' AND is_current;

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND aspect = 'imagery_style_guide' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, notes)
SELECT
  '5fe8785b-223d-41a3-88ee-c07187622381',
  'imagery_style_guide',
  data,
  'hand_authored',
  'dartsonline-traffic-2026-08-31b',
  'dartsonline-traffic-2026-08-31b',
  'Appends darts ANATOMY prohibitions to kinds.hero.avoid and kinds.content_hero.avoid, '
  || 'derived from the live row by jsonb_set (nothing retyped). Written after LOOKING at all '
  || 'four heroes regenerated 2026-08-31 12:35Z: hero-new-arrivals and hero-sale came back '
  || 'correct, hero-home came back with an all-red bull (no green outer ring) and hero-guides '
  || 'with a feathered archery flight and a "7" where the board reads "1". Every clause added '
  || 'names a defect observed in this site''s OWN output, not a generic caution. The clauses go '
  || 'in the per-kind overrides because avoidForKind returns the override INSTEAD of the '
  || 'guide-level avoid when one exists (imagery_style_guide.go), so a guide-level edit would '
  || 'have been inert for heroes while looking like a fix.'
FROM _new_guide;

-- ── 3. the two prompts, amended at the plan (the authority) ───────────────────
-- hero_home: "the deep black and red of the board" is the clause that produced a board with
-- no green in it. Replaced with the board's real colours, and the anatomy stated positively.
UPDATE site_plan_imagery spi
SET prompt =
'A dramatic, high-energy close-up of three tungsten darts embedded in the treble twenty of a professional sisal dartboard, shot with shallow depth of field and warm directional lighting from the left. The steel barrels catch the light with metallic brilliance against the black and cream sisal segments and the red and green of the treble ring. Moody dark background fading to black, conveying precision and competition. Physical accuracy matters: each dart is a steel point, a knurled tungsten barrel, a slim shaft and a flat moulded plastic flight, never a feather; the board face is black and cream radial segments divided by a steel wire spider, the doubles and trebles rings alternate red and green, and the bull is a small red centre inside a green outer ring.'
FROM site_plans sp
WHERE sp.id = spi.plan_id AND sp.is_current
  AND sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND spi.key = 'hero_home';

-- hero_guides: the original never said what a dart is, so a feather was a free choice.
-- The number sequence is stated because a garbled ring was the July defect and a "7" for "1"
-- is the residue of it.
UPDATE site_plan_imagery spi
SET prompt =
'A contemplative overhead shot of a dartboard with a single dart and a measuring tape laid across it, alongside a notebook with hand-drawn diagrams of dart trajectories. Warm desk-lamp lighting, educational and thoughtful mood. Dark wood surface. Physical accuracy matters: the dart is a steel point, a knurled tungsten barrel, a slim shaft and a flat moulded plastic flight, never a feather and never an archery arrow; the board face is black and cream radial segments divided by a steel wire spider, the doubles and trebles rings alternate red and green, the bull is a small red centre inside a green outer ring, and the numbers around the ring read 20, 1, 18, 4, 13, 6, 10, 15, 2, 17, 3, 19, 7, 16, 8, 11, 14, 9, 12, 5 clockwise from the top.'
FROM site_plans sp
WHERE sp.id = spi.plan_id AND sp.is_current
  AND sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND spi.key = 'hero_guides';

-- ── 4. re-request the two, from the amended plan rows ─────────────────────────
-- Same shape as SEED_2026-08-31: kind carried EXPLICITLY (the absent kind was
-- bugs_closed/382's cause), prompt taken from the plan row, filed at 'triaged' because
-- 'detected' items do not drain on this site.
INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, summary, spec, priority,
   handler_agent, status, created_by, item_key, approval_mode)
SELECT
  sp.site_id,
  'operator',
  'build',
  'needs_imagery',
  'medium',
  'Regenerate with darts anatomy stated (owner 2026-08-31, second pass): ' || spi.key,
  jsonb_build_object(
    'key',               spi.key,
    'kind',              'hero',
    'check',             'emit_imagery_items',
    'scope',             spi.scope,
    'scope_ref',         spi.scope_ref,
    'prompt',            spi.prompt,
    'purpose',           'hero',
    'asset_key',         spi.key,
    'style_hints',       jsonb_build_object('aspect_ratio', '16:9'),
    'brand_update',      false,
    'original_pipeline', 'build'
  ),
  40,
  'image-build-handler',
  'triaged',
  'dartsonline-traffic-2026-08-31b',
  'needs_imagery:' || spi.scope || ':' || COALESCE(spi.scope_ref,'-') || ':' || spi.key,
  'auto'
FROM site_plan_imagery spi
JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current
WHERE sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND spi.key IN ('hero_home','hero_guides');

-- ── 5. post-state assertions: prove the edit actually landed ──────────────────
DO $$
DECLARE n int; g jsonb;
BEGIN
  SELECT data INTO g FROM site_specs
   WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
     AND aspect = 'imagery_style_guide' AND is_current;
  IF (g #>> '{kinds,hero,avoid}') NOT LIKE '%fletched flights%' THEN
    RAISE EXCEPTION 'hero anatomy clause absent after write - jsonb_set no-opped';
  END IF;
  IF (g #>> '{kinds,content_hero,avoid}') NOT LIKE '%fletched flights%' THEN
    RAISE EXCEPTION 'content_hero anatomy clause absent after write - jsonb_set no-opped';
  END IF;
  -- the pre-existing clauses must SURVIVE: this file appends, it does not replace
  IF (g #>> '{kinds,hero,avoid}') NOT LIKE '%catalogue shot%' THEN
    RAISE EXCEPTION 'pre-existing hero avoid clauses lost - this file must append, not replace';
  END IF;
  IF (g #>> '{avoid}') NOT LIKE '%identifiable faces%' THEN
    RAISE EXCEPTION 'guide-level avoid lost';
  END IF;

  SELECT count(*) INTO n FROM site_specs
   WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
     AND aspect = 'imagery_style_guide' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION 'expected exactly 1 current style guide, found %', n; END IF;

  SELECT count(*) INTO n FROM site_plan_imagery spi
    JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current
   WHERE sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
     AND spi.key IN ('hero_home','hero_guides')
     AND spi.prompt LIKE '%never a feather%';
  IF n <> 2 THEN RAISE EXCEPTION 'expected 2 amended prompts, found %', n; END IF;

  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
     AND created_by = 'dartsonline-traffic-2026-08-31b';
  IF n <> 2 THEN RAISE EXCEPTION 'expected 2 new work items, found %', n; END IF;
END $$;

SELECT item_key, status, spec->>'asset_key' AS asset_key, left(spec->>'prompt', 60) AS prompt_head
FROM site_work_items
WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND created_by='dartsonline-traffic-2026-08-31b'
ORDER BY item_key;

COMMIT;
