-- FILE: docs/agent_docs/sql_for_agents/431_bugfix_285_restore_learn_ai_builders_content_first_from_archive.sql
--
-- bugs_open/285 (tool-improver rewrote a SHARED template) — the ONE live casualty.
--
-- WHAT HAPPENED (evidence in the bug file's 2026-08-15 evening contribution):
--   On 2026-08-14 18:48Z the tool-improver, fixing asset-formatter, (1) wrote
--   Asset Path Formatter markup into the shared `ported-page` template and
--   (2) — because its load_tool resolved the shared component with LIMIT 1 —
--   filed its DELIVERY (`section_edit`, content_edit, field_updates {}) at an
--   ARBITRARY placement: webdesign.co.uk `learn-ai-builders-content-first`,
--   a learn article. The section-editor re-rendered that page's `ported-page`
--   slot from the poisoned template against the ported stub content_data at
--   18:51:51Z, and the page deployed. Since then the LIVE page has served an
--   EMPTY <article> plus a fabricated "Related Downloads" list of three
--   non-existent files (content-first-checklist.pdf, ai-content-brief-template.docx,
--   prompt-library.md). Nothing detected it; the bug file said "zero pages ever
--   served the bad markup" — that claim was checked at the HEAD of the row, and
--   the poison's head IS the wrapper CSS.
--
-- THE RESTORE SOURCE IS PROVEN, NOT ASSUMED: migration 357's archive trigger
--   banked the row it overwrote — page_component_history
--   ab400131-2a41-434b-bd95-d44c9f064a32 (3,781 chars, op=overwrite, unstamped)
--   — and sha256(that html) == content_data->>'sha256' of the placement
--   (a2d9fa85…), the webdesignport provenance stamp of the ported body. Byte-exact.
--
-- WHAT THIS DOES: (a) restores page_components.rendered_html from that archive
--   row (the 357 trigger archives the poisoned version as it goes — nothing is
--   lost either way); clears rendered_html_digest, because ported bytes are not
--   reproducible from content_data and must not read as machine_made;
--   (b) queues a REASON-LESS page_rerender — assemble-only (chrome + stored
--   rendered_html), no LLM, no section re-render. A `reason` of
--   section_data_resolved/image_landed would route through
--   rerender_page_sections, which for THIS component (input_schema declares
--   `body` required, source llm; the stub has no body) escalates the whole page
--   to a needs_page LLM rebuild — the exact chain that ran 154 rerenders and 73
--   refused rebuilds on 2026-08-08. Do not add a reason.
--
-- GUARDED: refuses unless the row still carries the poison fingerprint AND the
--   archive row's sha matches the placement's provenance sha AND no page_rerender
--   is open on the page. A replay is a no-op (ON CONFLICT + fingerprint gate).
--
-- ROLLBACK: the poisoned html is archived by the trigger at restore time —
--   `SELECT rendered_html FROM page_component_history WHERE component_id =
--    'ff0404b0-f52a-41db-b04a-bc563c2a3a4f' AND op='overwrite' ORDER BY created_at DESC LIMIT 1`.
--   Nobody should want it back.

\set ON_ERROR_STOP on

BEGIN;

DO $$
DECLARE
    v_current      text;
    v_archived     text;
    v_stub_sha     text;
    v_inflight     int;
BEGIN
    SELECT pc.rendered_html, pc.content_data->>'sha256'
      INTO v_current, v_stub_sha
      FROM page_components pc
     WHERE pc.id = 'ff0404b0-f52a-41db-b04a-bc563c2a3a4f'
       AND pc.page_id = '53709f51-9323-4eb9-88d5-3e4f1a38fee9';
    IF NOT FOUND THEN
        RAISE EXCEPTION '285/431: placement ff0404b0… on page 53709f51… not found — re-derive before restoring';
    END IF;
    IF v_current NOT LIKE '%portedPageAssetList%' THEN
        RAISE EXCEPTION '285/431: the row no longer carries the 2026-08-14 poison fingerprint (portedPageAssetList) — already restored or changed by someone else; read it before acting';
    END IF;

    SELECT rendered_html INTO v_archived
      FROM page_component_history
     WHERE id = 'ab400131-2a41-434b-bd95-d44c9f064a32'
       AND component_id = 'ff0404b0-f52a-41db-b04a-bc563c2a3a4f'
       AND op = 'overwrite';
    IF v_archived IS NULL OR length(v_archived) = 0 THEN
        RAISE EXCEPTION '285/431: archive row ab400131… missing or empty — do not restore from a guess';
    END IF;
    IF encode(sha256(convert_to(v_archived, 'UTF8')), 'hex') <> v_stub_sha THEN
        RAISE EXCEPTION '285/431: archived html sha % does not match the placement provenance sha % — wrong archive row',
            encode(sha256(convert_to(v_archived, 'UTF8')), 'hex'), v_stub_sha;
    END IF;

    SELECT count(*) INTO v_inflight
      FROM site_work_items
     WHERE page_id = '53709f51-9323-4eb9-88d5-3e4f1a38fee9'
       AND item_type = 'page_rerender'
       AND status NOT IN ('complete', 'cancelled', 'rejected', 'wont_fix', 'failed');
    IF v_inflight > 0 THEN
        RAISE EXCEPTION '285/431: % page_rerender item(s) already open on this page — read them before adding another', v_inflight;
    END IF;

    UPDATE page_components
       SET rendered_html        = v_archived,
           rendered_html_digest = NULL,          -- ported bytes: unstamped by design (357 header)
           build_status         = 'deployed',
           updated_at           = now()
     WHERE id = 'ff0404b0-f52a-41db-b04a-bc563c2a3a4f'
       AND rendered_html LIKE '%portedPageAssetList%';
    IF NOT FOUND THEN
        RAISE EXCEPTION '285/431: guarded UPDATE matched no row';
    END IF;
    RAISE NOTICE '285/431: restored % chars over the % char poison', length(v_archived), length(v_current);
END $$;

-- The deploy. Reason-less on purpose — see header.
INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary, spec,
    page_id, priority, handler_agent, status, created_by, item_key
) VALUES (
    '6b49db8e-d447-4467-8277-4f3018af9897',
    'bugfix-285',
    'build',
    'page_rerender',
    'high',
    'Re-deploy learn-ai-builders-content-first after 431 restored its ported body from the 357 archive (bugs_open/285: the tool-improver''s mis-targeted delivery emptied the article and seeded fake download links on 2026-08-14)',
    jsonb_build_object(
        'domain',    'webdesign.co.uk',
        'page_id',   '53709f51-9323-4eb9-88d5-3e4f1a38fee9',
        'page_name', 'learn-ai-builders-content-first',
        'filename',  'learn/ai-builders/content-first.html'
        -- NO reason: assemble-only. See header.
    ),
    '53709f51-9323-4eb9-88d5-3e4f1a38fee9',
    20,
    'page-rerender',
    'triaged',
    'bugfix-285',
    'page_rerender:learn-ai-builders-content-first:285-archive-restore'
)
ON CONFLICT DO NOTHING;

DO $$
DECLARE v_id uuid;
BEGIN
    SELECT id INTO v_id FROM site_work_items
     WHERE item_key = 'page_rerender:learn-ai-builders-content-first:285-archive-restore';
    IF v_id IS NULL THEN
        RAISE EXCEPTION '285/431: the page_rerender item was not created — investigate before assuming it is queued';
    END IF;
    RAISE NOTICE '285/431: queued page_rerender % — verify at the SERVED page, not at this row', v_id;
END $$;

INSERT INTO schema_migrations (filename, applied_at)
VALUES ('431_bugfix_285_restore_learn_ai_builders_content_first_from_archive.sql', now())
ON CONFLICT DO NOTHING;

COMMIT;

-- ---------------------------------------------------------------------------
-- WATCH:
--   SELECT status, attempt_count, updated_at FROM site_work_items
--    WHERE item_key = 'page_rerender:learn-ai-builders-content-first:285-archive-restore';
--
-- VERIFY AT THE ARTEFACT (complete ≠ served):
--   curl -sL https://webdesign.co.uk/learn/ai-builders/content-first.html > /tmp/cf.html
--   grep -c 'portedPageAssetList'          /tmp/cf.html   # want 0 (was 1)
--   grep -c 'content-first-checklist.pdf'  /tmp/cf.html   # want 0 (was 1)
--   grep -c 'class="article-content"'      /tmp/cf.html   # want ≥1 (the ported article wrapper)
-- ---------------------------------------------------------------------------
