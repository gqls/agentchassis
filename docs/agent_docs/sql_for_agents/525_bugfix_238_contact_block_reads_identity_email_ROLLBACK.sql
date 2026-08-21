-- FILE: docs/agent_docs/sql_for_agents/525_bugfix_238_contact_block_reads_identity_email_ROLLBACK.sql
--
-- Revert bugs_open/238 §11.11: put `contact-block`'s `contact_email` source back
-- to `site_specs.contact.email`. Run BY HAND — the runner never applies a
-- ROLLBACK sidecar.
--
-- ⚠ WHAT REVERTING ACTUALLY DOES, stated plainly because it is not symmetrical
-- with applying. `site_specs.contact.email` resolves NOWHERE, on any site, and
-- never has — that is the whole finding. So this does not restore a previous
-- working state; it restores a BROKEN one, deliberately. Reach for it only if
-- the owner's 2026-08-21 ruling ("yes, that email should appear on contact
-- pages") is withdrawn.
--
-- ⚠ AND IT DOES NOT UNPUBLISH ANYTHING. Any page re-rendered while the forward
-- migration was live has the address in its `page_components.content_data` and
-- in its deployed HTML. Reverting the schema stops FUTURE resolves; it does not
-- remove the value from the six rows or the served pages. To actually withdraw a
-- published address you must also clear the key from `content_data` and
-- re-render each affected page — and note the PBP-039 carry will re-supply a
-- stored value on the next build if you clear only one of the two. Doing the
-- schema revert alone and calling the address withdrawn is the mistake this
-- paragraph exists to prevent.
--
-- The affected pages, so the second half is not left as an exercise:
--   leopardessconsulting.co.uk — contact, ai-readiness-quiz,
--     case-study-data-pipeline-companies-house, case-study-tool-generation-pipeline
--   robot-hands.com — contact
--   gamesdesign.co.uk — contact-index  (never resolved; nothing published)

\set ON_ERROR_STOP on

BEGIN;

DO $$
DECLARE
    v_src text;
BEGIN
    SELECT input_schema->'fields'->'contact_email'->>'source' INTO v_src
      FROM content_components
     WHERE id = '4ebdccb5-6eb0-4b03-9fe1-0fa5cc315fb4';
    IF v_src IS DISTINCT FROM 'site_specs.identity.email' THEN
        RAISE EXCEPTION '238/525 ROLLBACK: source is % (want site_specs.identity.email) — the forward migration is not applied, or something else moved it', COALESCE(v_src, '(absent)');
    END IF;
END $$;

UPDATE content_components
   SET input_schema = jsonb_set(
           input_schema,
           '{fields,contact_email,source}',
           '"site_specs.contact.email"'::jsonb,
           false),
       updated_at = now()
 WHERE id = '4ebdccb5-6eb0-4b03-9fe1-0fa5cc315fb4';

DO $$
DECLARE
    v_src text;
BEGIN
    SELECT input_schema->'fields'->'contact_email'->>'source' INTO v_src
      FROM content_components
     WHERE id = '4ebdccb5-6eb0-4b03-9fe1-0fa5cc315fb4';
    IF v_src IS DISTINCT FROM 'site_specs.contact.email' THEN
        RAISE EXCEPTION '238/525 ROLLBACK verify FAILED: source is %', COALESCE(v_src, '(absent)');
    END IF;
    RAISE NOTICE '238/525 ROLLBACK: contact_email points at site_specs.contact.email again — which resolves NOWHERE. Pages already re-rendered keep the published address; see this file''s header.';
END $$;

COMMIT;

-- ---------------------------------------------------------------------------
-- Census of what is still published, if you are withdrawing the address:
--
--   SELECT s.domain, p.name, pc.content_data->>'contact_email'
--     FROM page_components pc
--     JOIN pages p ON p.id = pc.page_id
--     JOIN sites s ON s.id = p.site_id
--    WHERE pc.build_status = 'deployed'
--      AND pc.content_data ? 'contact_email';
