-- 709: remove the four DEAD `sites.content_data.<purpose>_url` keys the
-- pre-v1.0.1326 StoreAssetAction poisoned and nothing reads — the residual
-- migration 562's header explicitly deferred ("SCOPE — hero_url ONLY").
-- bugs_open/114, bugfix_114_imagery_wiring lane, 2026-09-02 resumption.
-- REVISED same day after council round 1 (corr 3b568104): split into its own
-- submission; pre-mutation counted gate + per-arm assertions added
-- (debug_historian); the "no reader" claim VERIFIED, not asserted
-- (prior_art_librarian, bug_historian) — the queries and results are below.
--
-- THE FOUR KEYS, with the one distinct value each carries fleet-wide
-- [MEASURED 2026-09-02; 27 key-pairs across 18 distinct sites, arm-by-arm
-- SELECT preview run and the site names recorded in the lane NOTES]:
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
-- active asset keyed `icon`, `content_hero`, `illustration` or `sprite_sheet`.
-- HTTP probes: idea.uk icon.jpg 404, gamesdesign.co.uk content_hero.jpg 404,
-- apis.uk illustration.jpg 404.
--
-- "NOTHING READS THEM" — VERIFIED, with the queries, not asserted
-- [MEASURED 2026-09-02, after the council round that rightly objected to the
-- asserted version; the missingkey=zero hazard the bug_historian named is real
-- and is exactly why the template surface had to be swept, not sampled]:
--   * Go: the resolver's content_data fallback handles ONLY hero_url and
--     logo_url (plan_sections_action.go:653-694; the :206 map lists logo_url,
--     the :667 block hero_url). A repo grep for the four literals returns
--     comments and this migration family.
--   * Live templates, ALL FOUR literals, whole fleet:
--       SELECT count(*) FILTER (WHERE html_template LIKE '%icon_url%'),
--              count(*) FILTER (WHERE html_template LIKE '%content_hero_url%'),
--              count(*) FILTER (WHERE html_template LIKE '%sprite_sheet_url%'),
--              count(*) FILTER (WHERE html_template LIKE '%illustration_url%')
--         FROM content_components WHERE is_active;
--       -> 0 | 0 | 0 | 1
--     The single illustration_url hit is `brief-explanation`'s
--     `{{.illustration_url}}` — a FIELD rendered from the COMPONENT's resolved
--     content_data, declared `source: site_assets.illustration` and filled
--     from plan/asset tables; a template never reads sites.content_data
--     directly, and the only bridge (the resolver fallback) excludes this key.
--   * Field sources: 0 active input_schema fields declare a source naming any
--     of the four keys through content_data (LATERAL jsonb_each over
--     fields/properties, source LIKE '%content_data%' AND the key names).
--   * The council's own read-only check agreed: 0 live component templates
--     directly reference the three unaudited keys (round-1 report, corr
--     3b568104).
--   * `logo_url` (26 sites, 1 value) is NOT touched: 26/26 canonical-backed —
--     the identical-values signature is legitimate there. `hero_url` was
--     repaired by 562 (APPROVED, corr 4145fcdc) and is NOT touched.
--
-- ⚠ KEY NAMES ARE PINNED TO THE CURRENT SEED SHAPE (debug_historian's replay
-- note, from the 051 landmine): the arms match exact key names + exact
-- literals. If an EARLIER migration in a replay chain ever renames one of
-- these keys, the delete becomes a silent no-op indistinguishable from
-- "already clean" — the pre-gate below turns that into a loud count mismatch
-- instead, which is the reason it asserts exact per-key counts rather than
-- tolerating any partial state.
--
-- WHY NOW IS SAFE (the re-poisoning question): since v1.0.1326 (IMG-072, live;
-- fleet v1.0.1354) store_asset writes a site-wide `<purpose>_url` only for a
-- canonical store (asset_key == purpose) or a caller declaring
-- update_site_brand_assets, and writes the DEPLOYER'S OWN path. A future
-- canonical icon/illustration store re-creating its key with a real value is
-- legitimate, not re-poisoning — and such a site is excluded by every arm's
-- no-canonical-asset predicate anyway.
--
-- ROLLBACK: 709_remove_dead_purpose_url_keys_ROLLBACK.sql (sibling file)
-- restores the four keys from the per-row backup table without reasoning from
-- it under pressure.
--
-- RE-RUN BEHAVIOUR: the pre-gate accepts exactly two worlds — the recorded
-- census (16/6/4/1) or fully clean (0/0/0/0, a no-op re-run). ANY other
-- combination aborts: a partial state means the world moved under this file
-- (a key renamed, a site added or repaired by hand) and the census must be
-- re-derived, not deleted around.
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

-- PRE-MUTATION COUNTED GATE + THE ARMS + PER-ARM POST-CONDITIONS, one block,
-- so the counts the gate asserts are the counts the arms are then held to.
DO $$
DECLARE
    n_icon int; n_ch int; n_illus int; n_sprite int;
    touched int;
BEGIN
    -- The gate: each key's live occurrence (poisoned literal + no canonical
    -- asset) must be exactly the recorded census figure, or exactly zero.
    SELECT count(*) INTO n_icon FROM sites s
     WHERE s.content_data->>'icon_url' = '/assets/images/icon.jpg'
       AND NOT EXISTS (SELECT 1 FROM assets a WHERE a.site_id = s.id
                        AND a.asset_key = 'icon' AND a.status = 'active');
    SELECT count(*) INTO n_ch FROM sites s
     WHERE s.content_data->>'content_hero_url' = '/assets/images/content_hero.jpg'
       AND NOT EXISTS (SELECT 1 FROM assets a WHERE a.site_id = s.id
                        AND a.asset_key = 'content_hero' AND a.status = 'active');
    SELECT count(*) INTO n_illus FROM sites s
     WHERE s.content_data->>'illustration_url' = '/assets/images/illustration.jpg'
       AND NOT EXISTS (SELECT 1 FROM assets a WHERE a.site_id = s.id
                        AND a.asset_key = 'illustration' AND a.status = 'active');
    SELECT count(*) INTO n_sprite FROM sites s
     WHERE s.content_data->>'sprite_sheet_url' = '/assets/images/sprite_sheet.jpg'
       AND NOT EXISTS (SELECT 1 FROM assets a WHERE a.site_id = s.id
                        AND a.asset_key = 'sprite_sheet' AND a.status = 'active');

    IF NOT ((n_icon, n_ch, n_illus, n_sprite) = (16, 6, 4, 1)
         OR (n_icon, n_ch, n_illus, n_sprite) = (0, 0, 0, 0)) THEN
        RAISE EXCEPTION
          '709 pre-gate: live counts icon=% content_hero=% illustration=% sprite_sheet=% match neither the recorded census (16/6/4/1, counted 2026-09-02) nor fully-clean (0/0/0/0). The world moved under this file — re-derive the census before applying; do not widen this gate.',
          n_icon, n_ch, n_illus, n_sprite;
    END IF;

    -- The arms. Each is held to the count the gate just measured.
    UPDATE sites s SET content_data = s.content_data - 'icon_url'
     WHERE s.content_data->>'icon_url' = '/assets/images/icon.jpg'
       AND NOT EXISTS (SELECT 1 FROM assets a WHERE a.site_id = s.id
                        AND a.asset_key = 'icon' AND a.status = 'active');
    GET DIAGNOSTICS touched = ROW_COUNT;
    IF touched <> n_icon THEN
        RAISE EXCEPTION '709 arm icon_url: touched % rows, gate counted % — predicate drift mid-transaction', touched, n_icon;
    END IF;

    UPDATE sites s SET content_data = s.content_data - 'content_hero_url'
     WHERE s.content_data->>'content_hero_url' = '/assets/images/content_hero.jpg'
       AND NOT EXISTS (SELECT 1 FROM assets a WHERE a.site_id = s.id
                        AND a.asset_key = 'content_hero' AND a.status = 'active');
    GET DIAGNOSTICS touched = ROW_COUNT;
    IF touched <> n_ch THEN
        RAISE EXCEPTION '709 arm content_hero_url: touched % rows, gate counted %', touched, n_ch;
    END IF;

    UPDATE sites s SET content_data = s.content_data - 'illustration_url'
     WHERE s.content_data->>'illustration_url' = '/assets/images/illustration.jpg'
       AND NOT EXISTS (SELECT 1 FROM assets a WHERE a.site_id = s.id
                        AND a.asset_key = 'illustration' AND a.status = 'active');
    GET DIAGNOSTICS touched = ROW_COUNT;
    IF touched <> n_illus THEN
        RAISE EXCEPTION '709 arm illustration_url: touched % rows, gate counted %', touched, n_illus;
    END IF;

    UPDATE sites s SET content_data = s.content_data - 'sprite_sheet_url'
     WHERE s.content_data->>'sprite_sheet_url' = '/assets/images/sprite_sheet.jpg'
       AND NOT EXISTS (SELECT 1 FROM assets a WHERE a.site_id = s.id
                        AND a.asset_key = 'sprite_sheet' AND a.status = 'active');
    GET DIAGNOSTICS touched = ROW_COUNT;
    IF touched <> n_sprite THEN
        RAISE EXCEPTION '709 arm sprite_sheet_url: touched % rows, gate counted %', touched, n_sprite;
    END IF;

    RAISE NOTICE '709 arms OK: removed icon=% content_hero=% illustration=% sprite_sheet=%',
        n_icon, n_ch, n_illus, n_sprite;
END $$;

-- FINAL VERIFY — DO/RAISE (a SELECT cannot stop a COMMIT): nothing poisoned
-- survives, and the backup count is reported with its expectation.
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
        RAISE EXCEPTION '709 verify: % poisoned key(s) survived — roll back and re-derive', remaining;
    END IF;

    SELECT count(*) INTO backed_up FROM bak_site_dead_purpose_urls_20260902;
    RAISE NOTICE '709 OK: 0 poisoned keys remain; % row(s) held in bak_site_dead_purpose_urls_20260902 (18 expected on a 2026-09-02 first run; unchanged on a re-run)', backed_up;
END $$;

COMMIT;
