-- 654_site_declared_header_slots_seed_HOLD.sql
--
-- bugs_open/407. Seeds the FIRST per-site header declaration, now that a site can
-- have one: `site_specs` aspect `site_config`, key
-- `chrome.header_slots` — an ordered list of page names, read by
-- populate_nav_tables.
--
-- > ⚠ CORRECTED before this shipped. The first draft wrote to
-- > `sites.settings->'nav'`, on a rationale asserting "there is NO site_specs
-- > table". That was FALSE — read off a `\dt site*` listing truncated at twenty
-- > rows. The council's reuse_agent and prior_art_librarian seats both caught it,
-- > and the measurement settles it: `site_config.chrome` ALREADY carries per-site
-- > header config (`header_cta_url`/`header_cta_label` on oufe.com,
-- > `compliance_lines` on two more, `[MEASURED 2026-08-26]`). Header slots belong
-- > beside header CTAs, not in a second table.
--
-- ############################################################################
-- ##  _HOLD — DO NOT APPLY UNTIL A CHASSIS IMAGE CARRYING                    ##
-- ##  platform/orchestration/actions/nav_declaration.go HAS ROLLED.          ##
-- ############################################################################
--
-- ⚠ THE HOLD IS FOR A DIFFERENT REASON FROM THE USUAL ONE, AND SAYING SO MATTERS
-- BECAUSE IT CHANGES WHAT AN EARLY APPLY COSTS. Migration 648's hold exists
-- because a checks-array name the binary does not know FAILS A WHOLE STEP and
-- discards every earlier check's findings. Nothing like that happens here: the old
-- binary never reads `chrome.header_slots`, so applying this early is INERT, not
-- fatal.
--
-- The hold protects the MEASUREMENT instead. This bug's before/after comparison is
-- a mechanism replay (§6 of the bug file) diffed against the stored nav, and it
-- only means anything if no nav rebuild ran between the seed and the roll. Apply
-- early and the first rebuild after the roll gives a post-fix number with no
-- matching pre-fix one — which is how a fix ends up unverifiable while looking
-- finished.
--
-- WHY THIS SITE, AND WHY ONLY THIS SITE.
-- ai-agent-orchestration.com is the case where the OLD remedy is provably
-- exhausted. Three site-specific page names were hardcoded into the fleet-wide
-- tier-2 map to fix this once (populate_nav_tables_action.go, the tier2 comment),
-- and `[MEASURED 2026-08-26 at the SERVED page]` that site's header holds exactly
-- eight links — index, services, about, tools, contact, case-studies, blog,
-- model-directory — so ONE of those three names is present and its two siblings
-- are not. The workaround is in the source today and the pages it was written for
-- are still missing.
--
-- finetuning.uk is deliberately NOT seeded here even though it is the site the
-- owner raised. He named four pages he was happy to displace; the served header
-- has changed since (his offer page is now in, `approach` is out), so his current
-- ordering is an OPERATOR DECISION and inventing one in a migration would be this
-- estate writing the site's mind for it — which is the defect, wearing a nicer hat.
-- A commented template is below for whoever takes that decision with him.
--
-- APPLY IT BY HAND, in this order:
--
--   1. Confirm the running chassis carries the reader. There is no capability
--      registry for this one (unlike a discovery check), so probe the binary with
--      a POSITIVE and a NEGATIVE control in the same exec — never `strings`, which
--      is absent from these images:
--
--        for p in $(kubectl -n ai-persona-system get pods -l app=agent-chassis \
--                     -o name | cut -d/ -f2 | head -5); do
--          echo "== $p"
--          kubectl -n ai-persona-system exec "$p" -- \
--            grep -ac 'header_slots' /proc/1/exe                 # expect >= 1
--          kubectl -n ai-persona-system exec "$p" -- \
--            grep -ac 'header_slots_NOTREAL' /proc/1/exe         # expect 0 (control)
--        done
--
--      A control that comes out PRESENT means the probe matches everything and
--      proves nothing (the 40-zeros trap).
--
--   2. Take the PRE-FIX reading before applying, or the after has nothing to be
--      compared with. The §6 population plus the mechanism replay, both in
--      docs024_key_docs_latest/bugfix_407_site_declares_its_own_header/RUNBOOK.
--      `[MEASURED 2026-08-26]` the replay split the six absent pages 5/1: five
--      excluded by tier/cap, one (idea.uk `report`) barred by neverPrimaryTypes.
--
--   3. Apply this file, then trigger a nav rebuild for the seeded site — the
--      declaration does nothing until populate_nav_tables next runs:
--        {"action":"process","agent_type":"nav-updater","data":{"domain":"ai-agent-orchestration.com"}}
--
--   4. Ask WHAT DID I BREAK before what worked. The step returns the declaration's
--      own outcome, so read that rather than guessing:
--        SELECT collected_data->'refresh_nav_tables'->>'nav_declaration_source',
--               collected_data->'refresh_nav_tables'->'declared_missing',
--               collected_data->'refresh_nav_tables'->'declared_ineligible'
--          FROM orchestration_states
--         WHERE created_at > now() - interval '30 minutes'
--           AND collected_data ? 'refresh_nav_tables'
--         ORDER BY created_at DESC LIMIT 3;
--      `nav_declaration_source = 'invalid'` means this file wrote a shape the
--      reader could not use — roll back rather than leaving it half-read.
--
--   5. ⚠ VERIFY AT THE SERVED PAGE, NOT AT THE NAV TABLES. nav-updater's LAST step
--      only FILES re-render items, so the tables can be right for an hour while
--      every served header is stale (52 stale items on finetuning.uk, 2026-08-25).
--      Wait for those to drain, then:
--        curl -s https://ai-agent-orchestration.com/ \
--          | tr -d '\n' | grep -oE '<nav[^>]*>.*</nav>' | head -1 \
--          | grep -c 'href="/adoption-tracker.html"'      # expect 1
--      Anchor on <nav> or the chrome's own class, NEVER on a bare <header> tag: a
--      rendered page contains several, because components carry their own, and
--      matching the wrong one produced a 420-byte "header" with no links in it
--      during this bug's own investigation.
--
--   Rollback is 654_..._HOLD_ROLLBACK.sql. Removing the key is always safe — the
--   reader falls straight back to the fleet tier default — so it needs no hold.

-- A stray ROLLBACK first, deliberately. A previous hand-apply of a _HOLD file that
-- aborted leaves the psql session in a failed transaction, and every statement
-- after that returns "current transaction is aborted" — which reads like this
-- file's own failure. Harmless when there is nothing to roll back; it prints a
-- warning and continues. The council's debug_historian seat asked for it.
ROLLBACK;

BEGIN;

-- Idempotency FIRST, snapshot second. A replay of an already-applied migration
-- must not write a decoy "pre-update" row that actually holds the post-update
-- state (LANDMINES; and the shape 526/541 both carry, contributed 2026-08-26).
-- There is no snapshot_agent call here at all — this touches `sites`, not
-- `agent_definitions` — and the rollback is a one-key delete, so the pre-image is
-- reconstructible without one.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM site_specs sp JOIN sites s ON s.id = sp.site_id
                WHERE s.domain = 'ai-agent-orchestration.com'
                  AND sp.aspect = 'site_config' AND sp.is_current
                  AND sp.data->'chrome' ? 'header_slots') THEN
        RAISE EXCEPTION '654: already applied — ai-agent-orchestration.com already declares chrome.header_slots';
    END IF;
END $$;

-- COUNTED NEEDLE-GATE. Two premises this seed rests on, both re-checked rather
-- than assumed: the site exists exactly once, and the pages it names exist. A slot
-- naming a page that is not there is REPORTED at runtime rather than fatal — but a
-- migration that seeds three such names has misunderstood the site, and should
-- abort rather than leave an operator reading declared_missing and wondering.
DO $$
DECLARE n_site int; n_pages int;
BEGIN
    SELECT count(*) INTO n_site FROM sites WHERE domain = 'ai-agent-orchestration.com';
    IF n_site <> 1 THEN
        RAISE EXCEPTION '654 needle-gate: expected exactly 1 site row for ai-agent-orchestration.com, found %', n_site;
    END IF;

    SELECT count(*) INTO n_pages
      FROM pages p JOIN sites s ON s.id = p.site_id
     WHERE s.domain = 'ai-agent-orchestration.com'
       AND lower(p.name) IN ('index','services','tools','model-directory',
                             'protocol-tracker','adoption-tracker','news-index','about','contact')
       AND p.status IN ('active','deployed','pending');
    IF n_pages <> 9 THEN
        RAISE EXCEPTION '654 needle-gate: expected all 9 declared page names to exist in scope, found % — re-derive the slot list against the live pages before seeding it', n_pages;
    END IF;
END $$;

-- UPDATE THE CURRENT ROW IN PLACE rather than superseding it.
--
-- The house write path (WriteSiteSpecAction) marks the old row is_current=false
-- and inserts a merged successor. That is right for an AGENT writing a spec it
-- derived. It is wrong here: this is an operator declaration added to an existing
-- site_config, and a supersede would fork the row's history for a key nothing else
-- touches. `jsonb_set` with create_missing writes ONE key at ONE path and cannot
-- disturb `analytics`, `locale`, or the `chrome` keys already there.
--
-- ⚠ jsonb_set is STRICT: if `data` were NULL the whole expression returns NULL and
-- the row would be emptied. `data` is NOT NULL on this table, so that cannot
-- happen — stated because the guard is the schema, not this file, and a reader
-- must not have to re-derive it.
UPDATE site_specs sp
   SET data = jsonb_set(
         jsonb_set(
           COALESCE(sp.data, '{}'::jsonb),
           '{chrome}', COALESCE(sp.data->'chrome', '{}'::jsonb), true),
         '{chrome,header_slots}',
         '["index","services","tools","model-directory","protocol-tracker","adoption-tracker","news-index","about","contact"]'::jsonb,
         true),
       updated_at = NOW()
  FROM sites s
 WHERE s.id = sp.site_id
   AND s.domain = 'ai-agent-orchestration.com'
   AND sp.aspect = 'site_config' AND sp.is_current;

UPDATE site_specs sp
   SET data = jsonb_set(sp.data, '{chrome,max_header_items}', '9'::jsonb, true),
       updated_at = NOW()
  FROM sites s
 WHERE s.id = sp.site_id
   AND s.domain = 'ai-agent-orchestration.com'
   AND sp.aspect = 'site_config' AND sp.is_current;

-- VERIFY AS A RAISE, NEVER AS A SELECT. ON_ERROR_STOP does not fire on a non-empty
-- result, so a verify block made of SELECTs cannot stop the COMMIT and the
-- migration would report success having written the wrong thing.
DO $$
DECLARE slots int; cap int; rows_current int; sibling_keys int;
BEGIN
    SELECT count(*) INTO rows_current
      FROM site_specs sp JOIN sites s ON s.id = sp.site_id
     WHERE s.domain = 'ai-agent-orchestration.com'
       AND sp.aspect = 'site_config' AND sp.is_current;
    IF rows_current <> 1 THEN
        RAISE EXCEPTION '654 verify: expected exactly 1 current site_config row, found % — the partial unique index should make this impossible', rows_current;
    END IF;

    SELECT jsonb_array_length(sp.data->'chrome'->'header_slots'),
           (sp.data->'chrome'->>'max_header_items')::int,
           (SELECT count(*) FROM jsonb_object_keys(sp.data) k WHERE k <> 'chrome')
      INTO slots, cap, sibling_keys
      FROM site_specs sp JOIN sites s ON s.id = sp.site_id
     WHERE s.domain = 'ai-agent-orchestration.com'
       AND sp.aspect = 'site_config' AND sp.is_current;

    IF slots <> 9 THEN
        RAISE EXCEPTION '654 verify: chrome.header_slots is % long, expected 9', slots;
    END IF;
    IF cap <> 9 THEN
        RAISE EXCEPTION '654 verify: chrome.max_header_items is %, expected 9 — a nine-slot list under an eight-slot cap would truncate the site''s own declaration', cap;
    END IF;
    -- A jsonb_set that wrote the wrong PATH would satisfy both checks above while
    -- having replaced `data` wholesale. Containment cannot tell an add from a
    -- replace, so this asks about the siblings directly: analytics and locale are
    -- present on this row and must still be.
    IF sibling_keys < 1 THEN
        RAISE EXCEPTION '654 verify: the site_config row has no keys besides chrome — the write REPLACED where it should have set one path';
    END IF;
END $$;

COMMIT;

-- ---------------------------------------------------------------------------
-- TEMPLATE for the next site, to be filled in WITH the operator, never guessed.
-- The list is the header the site wants, in order, by pages.name.
--
--   UPDATE site_specs sp
--      SET data = jsonb_set(
--            jsonb_set(sp.data, '{chrome}', COALESCE(sp.data->'chrome','{}'::jsonb), true),
--            '{chrome,header_slots}',
--            '["index","services","your-own-model","..."]'::jsonb, true)
--     FROM sites s
--    WHERE s.id = sp.site_id AND s.domain = '<domain>'
--      AND sp.aspect = 'site_config' AND sp.is_current;
--
-- ⚠ A SITE WITH NO CURRENT site_config ROW GETS NOTHING FROM THAT UPDATE — it
-- matches nothing and reports UPDATE 0, which reads exactly like success. Check
-- first, and INSERT the row if it is absent:
--   SELECT count(*) FROM site_specs sp JOIN sites s ON s.id=sp.site_id
--    WHERE s.domain='<domain>' AND sp.aspect='site_config' AND sp.is_current;
-- [MEASURED 2026-08-26] 33 of 51 sites have one, so 18 do not.
--
-- The slot list may name a page the fleet default would never promote (a tool
-- page, a page under /tools/) — that is the point of the feature. It may NOT
-- usefully name a system page (404, sitemap, robots) or a legal page: those stay
-- ineligible and are reported as such, because they are correctness rather than
-- preference.
