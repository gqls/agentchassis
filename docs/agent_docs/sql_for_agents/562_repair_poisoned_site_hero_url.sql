-- 562: repair the site-wide `sites.content_data.hero_url` values that
-- StoreAssetAction poisoned, on the sites where the value names a file that does
-- not exist.
--
-- WHY. Until v1.0.1326, StoreAssetAction wrote `content_data.<purpose>_url` on
-- EVERY asset store, deriving the value from `storage.BuildAssetPaths(purpose)` —
-- the PURPOSE alone — while the deployer commits under the ASSET KEY
-- (`storage.DeployedAssetPath`). So every page-scoped hero generation re-stamped
-- the site-wide default with `/assets/images/hero.jpg`, a path that exists only on
-- sites that happen to hold a canonical `hero` asset. `bugs_open/114`, register
-- IMG-072, council corr 3c0560f3 (APPROVED).
--
-- `ensureAssets` reads this value as its LAST resort (plan_sections_action.go, the
-- content_data fallback after the four page-scoped routes), so on a page whose own
-- routes miss it is what the hero renders as.
--
-- ORDERING — THE GATE MUST BE LIVE FIRST, AND IT IS. This file is deliberately
-- held until the fix ships, because on an older binary the next hero generation
-- simply re-poisons every row this repairs. That is not hypothetical: it is what
-- happened to fundamentallyai.com's hand-repair of 2026-07-29, which had reverted
-- by 2026-08-22. The guard below refuses to apply unless the marker asserting a
-- fixed fleet is set, so "remember to check the version" is not left to a human.
--
-- Verified before writing this file (2026-08-22, after the v1.0.1326 roll):
-- both agent-chassis replicas carry the new literals in /proc/1/exe, with a
-- present-control and an absent-control in the same pass.
--
-- WHAT IT DOES, per site, and it is deliberately NARROW:
--   * a site holding an ACTIVE asset keyed `hero`  -> LEFT ALONE. The value is
--     correct there; 7 sites are in this class and each serves 200.
--   * a site with no canonical hero but an ACTIVE `hero_home` -> repointed to
--     `/assets/images/hero-home.jpg`, which is `DeployedWebPath('hero_home','hero')`
--     and is the same repair fundamentallyai was given by hand.
--   * a site with neither -> the key is REMOVED. `ensureAssets` skips an absent or
--     empty value and falls through to the component's own declared fallback, which
--     is strictly better than naming a 404.
--
-- MEASURED AT THE WIRE, 2026-08-22, not inferred from the DB (the point of the
-- repair is what a visitor gets):
--   hero-home=200 / hero=404 on apis.uk, dartsonline.com, fundamentallyai.com,
--   lendzy.co.uk, loanzy.uk, relojistas.com, remortgagecalculator.uk,
--   vetcomparison.uk, vonc.com  -> the 9 repointed.
--   loancalculator.co.uk: both 404 and no hero_home asset -> key removed.
--
-- TWO SITES WHERE THE WIRE PROBE COULD NOT CONFIRM AN IMPROVEMENT, and they are
-- INCLUDED anyway — deliberately, with the reason stated, because an earlier draft
-- of this header claimed they were "excluded by the WHERE clause's own logic" and
-- that was simply FALSE. Both hold a `hero_home` asset, so ARM 1 catches them. The
-- claim was caught by previewing the arms against the live DB instead of trusting
-- the comment; it is left recorded here because a header asserting an exclusion the
-- SQL does not implement is worse than no header at all.
--   * noted.co.uk serves /assets/images/hero.jpg 200 despite holding NO asset row
--     keyed `hero` — the file is in the repo without a row behind it. Repointing
--     changes a currently-working reference, which is why it needed a decision
--     rather than a default: it is included because `hero_home` is the asset the
--     site actually owns, and a reference backed only by an orphan file breaks the
--     day anything reconciles files against rows. The orphan itself is recorded,
--     not acted on.
--   * webdesign.uk returns 302 for BOTH paths, so the probe cannot say which
--     resolves and no measurement here is decisive. Included because the DB is
--     unambiguous — a `hero_home` asset, no canonical `hero` — so the repair aligns
--     the record with the assets the site owns rather than with a path nothing
--     backs. If that site's 302 later resolves to a working hero.jpg, this was a
--     no-op on a value nothing reads; if it resolves to a 404, it was the fix.
--   Neither is carved out by name. A hardcoded name list goes stale silently, and
--   the honest predicate — "is the current value actually served?" — is an HTTP
--   question SQL cannot ask, which is itself why the wire probe is recorded above
--   rather than encoded below.
--
-- SCOPE — hero_url ONLY. `content_data` also carries poisoned `icon_url`,
-- `content_hero_url`, `illustration_url` and `sprite_sheet_url`, all with exactly
-- one distinct value fleet-wide and all naming files the deployer cannot produce.
-- They are NOT touched here: no Go code reads them, 0 deployed page_components
-- reference their literals, and `illustration_url` HAS a live template consumer
-- (`brief-explanation`.html_template), so removing it needs its own analysis rather
-- than riding a hero repair. They are dead values waiting for a template change —
-- recorded in bugs_open/114 as the residual.
--
-- IDEMPOTENT: re-running touches 0 rows (the WHERE clause excludes rows already
-- carrying the repaired value).
--
-- HOW TO RUN (the marker is a GUC, not a psql -v variable — psql refuses a dotted
-- -v name and does not substitute :vars inside dollar-quoted bodies):
--
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
--       -c "SET imagery114.gate_is_live = 'v1.0.1326+'" \
--       -f - < docs/agent_docs/sql_for_agents/562_repair_poisoned_site_hero_url.sql
--
-- -c and -f run in the SAME session, in that order, so the plain SET (not SET
-- LOCAL) is still in force inside the transaction below.

BEGIN;

-- Refuse on an unfixed fleet. A repair applied before the gate is live is undone
-- by the next generation, which is the exact failure this file exists to end.
DO $$
BEGIN
    IF current_setting('imagery114.gate_is_live', true) IS DISTINCT FROM 'v1.0.1326+' THEN
        RAISE EXCEPTION
            'REFUSING: imagery114.gate_is_live is not set to ''v1.0.1326+''. This repair is only durable on a chassis carrying IMG-072 (store_asset''s site-wide brand-state gate). On an older binary the next hero generation re-poisons every row repaired here — see fundamentallyai.com, hand-repaired 2026-07-29 and reverted by 2026-08-22. Verify the fleet first: kubectl -n ai-persona-system exec <chassis-pod> -- grep -aqF "Left site-wide content_data untouched" /proc/1/exe';
    END IF;
END $$;

-- Backup the exact rows being changed (not the whole table).
CREATE TABLE IF NOT EXISTS bak_site_hero_url_20260822 AS
SELECT id, domain, content_data, now() AS backed_up_at FROM sites WHERE false;

INSERT INTO bak_site_hero_url_20260822
SELECT s.id, s.domain, s.content_data, now()
  FROM sites s
 WHERE s.content_data->>'hero_url' = '/assets/images/hero.jpg'
   AND NOT EXISTS (SELECT 1 FROM assets a
                    WHERE a.site_id = s.id AND a.asset_key = 'hero' AND a.status = 'active')
   AND NOT EXISTS (SELECT 1 FROM bak_site_hero_url_20260822 b WHERE b.id = s.id);

-- ARM 1 — repoint to the homepage hero where one actually exists.
UPDATE sites s
   SET content_data = jsonb_set(s.content_data, '{hero_url}', '"/assets/images/hero-home.jpg"'::jsonb)
 WHERE s.content_data->>'hero_url' = '/assets/images/hero.jpg'
   AND NOT EXISTS (SELECT 1 FROM assets a
                    WHERE a.site_id = s.id AND a.asset_key = 'hero' AND a.status = 'active')
   AND EXISTS (SELECT 1 FROM assets a
                WHERE a.site_id = s.id AND a.asset_key = 'hero_home' AND a.status = 'active');

-- ARM 2 — no canonical hero and no homepage hero: remove the key rather than
-- leave it naming a 404. ensureAssets skips an absent value.
UPDATE sites s
   SET content_data = s.content_data - 'hero_url'
 WHERE s.content_data->>'hero_url' = '/assets/images/hero.jpg'
   AND NOT EXISTS (SELECT 1 FROM assets a
                    WHERE a.site_id = s.id AND a.asset_key = 'hero' AND a.status = 'active')
   AND NOT EXISTS (SELECT 1 FROM assets a
                    WHERE a.site_id = s.id AND a.asset_key = 'hero_home' AND a.status = 'active');

-- VERIFY — a DO/RAISE block, never bare SELECTs. ON_ERROR_STOP ignores a non-empty
-- result set, so a verify block built from SELECTs cannot stop the COMMIT.
DO $$
DECLARE
    still_broken int;
    repointed    int;
    kept         int;
BEGIN
    -- No site may still name hero.jpg without holding the asset that backs it.
    SELECT count(*) INTO still_broken
      FROM sites s
     WHERE s.content_data->>'hero_url' = '/assets/images/hero.jpg'
       AND NOT EXISTS (SELECT 1 FROM assets a
                        WHERE a.site_id = s.id AND a.asset_key = 'hero' AND a.status = 'active');

    SELECT count(*) INTO repointed
      FROM sites WHERE content_data->>'hero_url' = '/assets/images/hero-home.jpg';

    SELECT count(*) INTO kept
      FROM sites s
     WHERE s.content_data->>'hero_url' = '/assets/images/hero.jpg'
       AND EXISTS (SELECT 1 FROM assets a
                    WHERE a.site_id = s.id AND a.asset_key = 'hero' AND a.status = 'active');

    IF still_broken > 0 THEN
        RAISE EXCEPTION 'VERIFY FAILED: % site(s) still name /assets/images/hero.jpg with no canonical hero asset', still_broken;
    END IF;

    -- A positive control on the untouched class: the 7 sites that legitimately
    -- carry hero.jpg must STILL carry it. A repair that quietly emptied them
    -- would satisfy the check above, which is why this one exists.
    IF kept = 0 THEN
        RAISE EXCEPTION 'VERIFY FAILED: no site retains /assets/images/hero.jpg — the arms have over-reached and stripped the sites where the value is correct';
    END IF;

    RAISE NOTICE 'repair OK: % site(s) now point at hero-home.jpg, % correctly retain hero.jpg, 0 broken', repointed, kept;
END $$;

COMMIT;
