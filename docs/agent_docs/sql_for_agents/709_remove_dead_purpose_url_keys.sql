-- 709: remove the four DEAD `sites.content_data.<purpose>_url` keys the
-- pre-v1.0.1326 StoreAssetAction poisoned and nothing reads — the residual
-- migration 562's header explicitly deferred ("SCOPE — hero_url ONLY").
-- bugs_open/114, bugfix_114_imagery_wiring lane, 2026-09-02 resumption.
--
-- THE FOUR KEYS, with the one distinct value each carries fleet-wide
-- [MEASURED 2026-09-02; 562's census had the same four with 3 illustration
-- sites — the 4th, apis.uk, was poisoned pre-roll on roll day, see NOTES]:
--
--   icon_url          16 sites   /assets/images/icon.jpg
--   content_hero_url   6 sites   /assets/images/content_hero.jpg
--   illustration_url   4 sites   /assets/images/illustration.jpg
--   sprite_sheet_url   1 site    /assets/images/sprite_sheet.jpg
--
-- One distinct value per key across every site is the signature of a value
-- with no site and no asset input (562's argument). Two of the four literals
-- (content_hero.jpg, sprite_sheet.jpg) keep the UNDERSCORE, which the deployer
-- structurally cannot produce (DeployedWebPath hyphenates); the other two are
-- producible only by a canonical store of that purpose, and ZERO sites hold an
-- active asset keyed `icon`, `content_hero`, `illustration` or `sprite_sheet`
-- [MEASURED 2026-09-02: EXISTS(assets, asset_key = replace(k,'_url','')) = 0
-- for all 27 site-key pairs]. HTTP probes: idea.uk icon.jpg 404,
-- gamesdesign.co.uk content_hero.jpg 404, apis.uk illustration.jpg 404.
--
-- WHY DELETION IS BEHAVIOUR-NEUTRAL — the analysis 562 asked for, done:
--   * No Go reader. The resolver's content_data fallback handles ONLY
--     `hero_url` and `logo_url` (plan_sections_action.go:653-694; the map at
--     :206 lists logo_url alone, hero_url is the block at :667). grep for the
--     other four key literals across platform/ internal/ returns comments and
--     this family of migrations only.
--   * The one template consumer that LOOKS like a reader is not one:
--     `brief-explanation.html_template`'s `{{.illustration_url}}` is a FIELD
--     whose declared source is `site_assets.illustration` — resolved from
--     site_plan_imagery/assets by kind, never from sites.content_data. The
--     inline_guide_imagery lane's 2026-09-02 census of image-field sources
--     agrees (1 field sourced site_assets.illustration, none sourced from
--     content_data keys).
--   * `logo_url` (26 sites, 1 value) is NOT touched: 26/26 sites hold an
--     active canonical `logo` asset — the identical-values signature is
--     legitimate there because every canonical logo deploys to the same path.
--   * `hero_url` was repaired by 562 and is NOT touched.
--
-- WHY NOW IS SAFE (the re-poisoning question): since v1.0.1326 (IMG-072, live;
-- fleet now v1.0.1354) store_asset writes a site-wide `<purpose>_url` only for
-- a canonical store (asset_key == purpose) or a caller declaring
-- update_site_brand_assets, and writes the DEPLOYER'S OWN path. A future
-- canonical icon/illustration store re-creating its key with a real value is
-- legitimate, not re-poisoning.
--
-- IDEMPOTENT: re-running matches 0 rows (the WHERE demands the poisoned
-- literal, which the first run removed).
--
-- HOW TO RUN:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
--       -f - < docs/agent_docs/sql_for_agents/709_remove_dead_purpose_url_keys.sql

BEGIN;

-- Backup the exact rows being changed (562's pattern — not the whole table).
CREATE TABLE IF NOT EXISTS bak_site_dead_purpose_urls_20260902 AS
SELECT id, domain, content_data, now() AS backed_up_at FROM sites WHERE false;

INSERT INTO bak_site_dead_purpose_urls_20260902
SELECT s.id, s.domain, s.content_data, now()
  FROM sites s
 WHERE (   s.content_data->>'icon_url'         = '/assets/images/icon.jpg'
        OR s.content_data->>'content_hero_url' = '/assets/images/content_hero.jpg'
        OR s.content_data->>'illustration_url' = '/assets/images/illustration.jpg'
        OR s.content_data->>'sprite_sheet_url' = '/assets/images/sprite_sheet.jpg')
   AND NOT EXISTS (SELECT 1 FROM bak_site_dead_purpose_urls_20260902 b WHERE b.id = s.id);

-- One arm per key. Each removes the key ONLY where it still carries the
-- poisoned literal AND the site holds no active canonical asset of that
-- purpose — so a site that legitimately re-acquires the key (post-IMG-072
-- canonical store) can never match.
UPDATE sites s SET content_data = s.content_data - 'icon_url'
 WHERE s.content_data->>'icon_url' = '/assets/images/icon.jpg'
   AND NOT EXISTS (SELECT 1 FROM assets a
                    WHERE a.site_id = s.id AND a.asset_key = 'icon' AND a.status = 'active');

UPDATE sites s SET content_data = s.content_data - 'content_hero_url'
 WHERE s.content_data->>'content_hero_url' = '/assets/images/content_hero.jpg'
   AND NOT EXISTS (SELECT 1 FROM assets a
                    WHERE a.site_id = s.id AND a.asset_key = 'content_hero' AND a.status = 'active');

UPDATE sites s SET content_data = s.content_data - 'illustration_url'
 WHERE s.content_data->>'illustration_url' = '/assets/images/illustration.jpg'
   AND NOT EXISTS (SELECT 1 FROM assets a
                    WHERE a.site_id = s.id AND a.asset_key = 'illustration' AND a.status = 'active');

UPDATE sites s SET content_data = s.content_data - 'sprite_sheet_url'
 WHERE s.content_data->>'sprite_sheet_url' = '/assets/images/sprite_sheet.jpg'
   AND NOT EXISTS (SELECT 1 FROM assets a
                    WHERE a.site_id = s.id AND a.asset_key = 'sprite_sheet' AND a.status = 'active');

-- VERIFY — DO/RAISE, never bare SELECTs (a non-empty result cannot stop a
-- COMMIT under ON_ERROR_STOP).
DO $$
DECLARE remaining int; backed_up int;
BEGIN
    SELECT count(*) INTO remaining
      FROM sites s, LATERAL jsonb_each_text(s.content_data) AS e(k, v)
     WHERE jsonb_typeof(s.content_data) = 'object'
       AND ((e.k = 'icon_url'         AND e.v = '/assets/images/icon.jpg')
         OR (e.k = 'content_hero_url' AND e.v = '/assets/images/content_hero.jpg')
         OR (e.k = 'illustration_url' AND e.v = '/assets/images/illustration.jpg')
         OR (e.k = 'sprite_sheet_url' AND e.v = '/assets/images/sprite_sheet.jpg'))
       AND NOT EXISTS (SELECT 1 FROM assets a
                        WHERE a.site_id = s.id
                          AND a.asset_key = replace(e.k, '_url', '')
                          AND a.status = 'active');
    IF remaining <> 0 THEN
        RAISE EXCEPTION '709 verify: % poisoned key(s) survived the four arms — an arm''s literal or predicate no longer matches the data; roll back and re-derive', remaining;
    END IF;

    -- The backup count is REPORTED, not asserted: the INSERT's predicate is
    -- the union of the arms' predicates, and asserting a minimum would falsely
    -- fail a run against a world where the population had already been
    -- cleaned. Previewed as SELECTs 2026-09-02 (the 562 lesson — read the
    -- names the WHERE returns before applying): 27 key-pairs across 18
    -- distinct sites, so 18 backup rows expected on a first run, 0 new on a
    -- re-run.
    SELECT count(*) INTO backed_up FROM bak_site_dead_purpose_urls_20260902;
    RAISE NOTICE '709 OK: 0 poisoned keys remain; % row(s) held in bak_site_dead_purpose_urls_20260902', backed_up;
END $$;

COMMIT;
