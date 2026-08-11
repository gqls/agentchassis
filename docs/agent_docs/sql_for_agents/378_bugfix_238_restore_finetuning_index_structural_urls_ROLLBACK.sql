-- FILE: docs/agent_docs/sql_for_agents/378_bugfix_238_restore_finetuning_index_structural_urls_ROLLBACK.sql
--
-- Undo of 378. Run BY HAND, deliberately — the migration runner never applies a
-- ROLLBACK sidecar.
--
-- Restores the row to the 47-key state bugs_open/238 left it in (five empty
-- <img src="">, five vanished card links, no section CTA) and removes the two
-- seeded site_specs aspects. Only do this if the restored content is judged
-- WRONG — the pre-repair state is the bug, not a safe harbour.
--
-- NOT VALID AFTER A REBUILD. If the page has been rebuilt since 378 applied,
-- page_components.id will have changed (save_page_sections DELETEs and
-- re-INSERTs) and this file's WHERE clause silently matches nothing. Re-derive
-- the row id first:
--   SELECT pc.id FROM page_components pc JOIN pages p ON p.id = pc.page_id
--    WHERE p.id = 'a716cacc-eec2-4aa6-a08b-7e6732506f41'
--      AND pc.slot_name = 'case-studies-grid';

\set ON_ERROR_STOP on

BEGIN;

DO $$
DECLARE
    v_keys int;
BEGIN
    SELECT (SELECT count(*) FROM jsonb_object_keys(pc.content_data))
      INTO v_keys
      FROM page_components pc
     WHERE pc.id = 'e20e474f-2e22-4a60-b56d-3cc2fd0f1de7';

    IF NOT FOUND THEN
        RAISE EXCEPTION '238 rollback: row e20e474f-... not found — the page was rebuilt; re-derive the row id (see header)';
    END IF;
    IF v_keys <> 58 THEN
        RAISE EXCEPTION '238 rollback: key count is % (expected 58) — 378 is not the state being undone; stop and re-measure', v_keys;
    END IF;
END $$;

UPDATE page_components
   SET content_data = content_data
                      - 'card1_image_url' - 'card2_image_url' - 'card3_image_url'
                      - 'card4_image_url' - 'card5_image_url'
                      - 'card1_link_url'  - 'card2_link_url'  - 'card3_link_url'
                      - 'card4_link_url'  - 'card5_link_url'
                      - 'cta_link_url',
       updated_at = now()
 WHERE id = 'e20e474f-2e22-4a60-b56d-3cc2fd0f1de7';

DELETE FROM site_specs
 WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
   AND aspect IN ('case_studies', 'pages')
   AND created_by = 'bugfix-238';

DO $$
DECLARE
    v_keys int;
BEGIN
    SELECT (SELECT count(*) FROM jsonb_object_keys(pc.content_data))
      INTO v_keys
      FROM page_components pc
     WHERE pc.id = 'e20e474f-2e22-4a60-b56d-3cc2fd0f1de7';
    IF v_keys <> 47 THEN
        RAISE EXCEPTION '238 rollback: key count is % after undo, expected 47 — aborting', v_keys;
    END IF;
    RAISE NOTICE '238 rollback: row back to 47 keys, seeded aspects removed. The page still serves the REPAIRED html until it is re-rendered.';
END $$;

COMMIT;
