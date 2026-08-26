-- ROLLBACK for 654 — remove the per-site header declaration from
-- site_specs.data->'chrome' (header_slots and max_header_items only).
--
-- WHEN YOU WOULD RUN THIS: the seeded header is wrong for the site, or the roll it
-- depends on was reverted, or `nav_declaration_source` came back `invalid` (this
-- file wrote a shape the reader could not use).
--
-- NOTE THE ASYMMETRY WITH APPLYING. Removing the key is ALWAYS safe: the reader
-- falls straight back to the fleet tier default, which is what every other site
-- uses. So this file needs no hold and no binary probe.
--
-- ⚠ IT DOES NOT UNDO THE NAV ITSELF. `site_nav_items` is derived and was rebuilt
-- when the declaration was applied; removing the declaration does not rebuild it
-- back. Trigger nav-updater for the site afterwards, or the header stays as the
-- declaration left it while the declaration is gone — which is the most confusing
-- of the available states:
--   {"action":"process","agent_type":"nav-updater","data":{"domain":"<domain>"}}
-- and then wait for the re-render items to drain before reading the served page.

BEGIN;

UPDATE site_specs sp
   SET data = jsonb_set(sp.data, '{chrome}',
         (sp.data->'chrome') - 'header_slots' - 'max_header_items', true),
       updated_at = NOW()
  FROM sites s
 WHERE s.id = sp.site_id
   AND s.domain = 'ai-agent-orchestration.com'
   AND sp.aspect = 'site_config' AND sp.is_current
   AND sp.data->'chrome' ? 'header_slots';

DO $$
DECLARE still boolean; chrome_kept boolean;
BEGIN
    SELECT (sp.data->'chrome' ? 'header_slots'), (sp.data ? 'chrome')
      INTO still, chrome_kept
      FROM site_specs sp JOIN sites s ON s.id = sp.site_id
     WHERE s.domain = 'ai-agent-orchestration.com'
       AND sp.aspect = 'site_config' AND sp.is_current;
    IF still THEN
        RAISE EXCEPTION '654_ROLLBACK: chrome.header_slots is still present';
    END IF;
    -- The `-` operators remove two keys from `chrome`; `chrome` itself, and the
    -- header CTA keys other sites keep there, must survive.
    IF NOT chrome_kept THEN
        RAISE EXCEPTION '654_ROLLBACK: the chrome object is gone — the removal took more than its two keys';
    END IF;
END $$;

COMMIT;
