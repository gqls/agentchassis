-- 212_tool_crosslink_action_notes.sql — travelling NOTES for the three actions
-- changed by bugs_open/029, so the next thread to touch them inherits the
-- decision instead of rediscovering it.
--
-- Answers the council gate's tooling_provenance objection on round 1 of
-- submission 745f9dfd: "three workflows edited without evidence the travelling
-- PLAN/NOTES for those subjects were consulted or will be updated."
--
-- Two grounded corrections to that objection's framing, both checked live
-- 2026-07-26 rather than assumed:
--
--   1. `agent` is not a subject_type in this table. The whole vocabulary is
--      pipeline (239), experience (56), tool (39), action (4) — so an
--      agent-keyed note would be unreadable by every existing consumer. The
--      travelling-doc unit for a Go action is subject_type='action',
--      subject_key=<action name>, which is what this file writes.
--
--   2. There were no prior decisions to consult for these three subjects:
--      SELECT * FROM doc_notes WHERE subject_type='action'
--        AND subject_key IN ('deploy_tool_to_site','create_tool_component',
--                            'create_tool_cross_link_items');
--      returns 0 rows, and doc_plans has no rows for them either. "Consulted,
--      found nothing" is the honest answer — and after this file the next
--      thread will not have that problem.
--
-- 211 already wrote the pipeline-level note (subject_type='pipeline',
-- subject_key='build'). This adds the per-action detail that a pipeline note is
-- the wrong altitude for.
--
-- Idempotent: each insert is guarded on its own subject_key + this file's
-- created_by, so a re-run adds nothing (the trap 211 fell into — it was applied
-- twice and left two identical pipeline notes).
--
-- ROLLBACK: DELETE FROM doc_notes WHERE created_by = '212_tool_crosslink_action_notes';

BEGIN;

INSERT INTO doc_notes (id, subject_type, subject_key, body, categories, source, created_by)
SELECT gen_random_uuid(), 'action', 'create_tool_cross_link_items',
  '2026-07-26 (bugs_open/029): this action is NO LONGER the emitter. It used to construct the tool page URL as /tools/{function}.html at SUGGESTION time and bake it into both the rewrite instruction and its acceptance_test — 0 of 27 emitted items across 4 sites resolved to a real page, including tools that WERE built, because pages.url has three incompatible shapes (prefix stripped, prefix kept, CanonicalisePage /index.html) and none is derivable from the function name.
The file now holds emitToolCrossLinkItems, a shared emitter that TAKES a real pages.url and refuses anything that is not an absolute path, called from deploy_tool_to_site and create_tool_component. The action itself stays REGISTERED but fail-safe (resolves the tool to a real page via page_components -> content_components.function, emits nothing when there is none) because an unregistered action named in config invalidates the whole workflow (bugs_closed/017) and config can be restored from a stale backup.
Its workflow step (tool-suggester.create_cross_links) was deleted by migration 211. Do not re-add it.
Categories: fix',
  '["fix"]'::jsonb, 'migration', '212_tool_crosslink_action_notes'
WHERE NOT EXISTS (
  SELECT 1 FROM doc_notes WHERE subject_type='action'
    AND subject_key='create_tool_cross_link_items'
    AND created_by='212_tool_crosslink_action_notes');

INSERT INTO doc_notes (id, subject_type, subject_key, body, categories, source, created_by)
SELECT gen_random_uuid(), 'action', 'deploy_tool_to_site',
  '2026-07-26 (bugs_open/029): now emits the tool''s cross-link content_rewrite items itself, after the page row and its needs_content_page item exist, using the pageURL it just wrote (never a constructed one). Reads related_pages from the add_tool spec — new Optional input, wired in step config by migration 211 as "related_pages": "input_data.spec.related_pages", with a direct input_data.spec read as the rollout-order fallback.
The already-deployed EARLY RETURN also emits now, resolving the URL from the page row: re-running the deployer is the supported way to backfill cross-links for a tool deployed before this fix. Dedup (item_key tool_crosslink:{function}:{page}:{site}, inserted through the central insertWorkItem helper) makes the repeat harmless.
Returns cross_links_added in its result map.
Categories: fix',
  '["fix"]'::jsonb, 'migration', '212_tool_crosslink_action_notes'
WHERE NOT EXISTS (
  SELECT 1 FROM doc_notes WHERE subject_type='action'
    AND subject_key='deploy_tool_to_site'
    AND created_by='212_tool_crosslink_action_notes');

INSERT INTO doc_notes (id, subject_type, subject_key, body, categories, source, created_by)
SELECT gen_random_uuid(), 'action', 'create_tool_component',
  '2026-07-26 (bugs_open/029): now emits the tool''s cross-link content_rewrite items itself, after the page row and its needs_content_page item exist, using the pageURL CanonicalisePage just produced. This path''s URL shape (/tools/<function>/index.html) is one of the three the old suggestion-time constructor got wrong. Reads related_pages from the add_tool spec — new Optional input, wired in step config by migration 211.
The emitted items are GATED: emitted immediately only if the tool page is already deployed/needs_rebuild, otherwise depends_on the open needs_content_page item for that page; if there is no open item, or it failed terminally, nothing is emitted. A tool page that never deploys therefore leaves no dead link. Declined emits are recorded in agent_error_log under error_code tool_crosslink_not_emitted:* — check there before concluding cross-linking is broken.
Returns cross_links_added in its result map.
Categories: fix',
  '["fix"]'::jsonb, 'migration', '212_tool_crosslink_action_notes'
WHERE NOT EXISTS (
  SELECT 1 FROM doc_notes WHERE subject_type='action'
    AND subject_key='create_tool_component'
    AND created_by='212_tool_crosslink_action_notes');

DO $$
DECLARE
  n int;
BEGIN
  SELECT count(*) INTO n FROM doc_notes
  WHERE created_by = '212_tool_crosslink_action_notes';
  IF n <> 3 THEN
    RAISE EXCEPTION '212 GUARD: expected 3 action notes, found %', n;
  END IF;
END $$;

COMMIT;

-- Verify
SELECT subject_type, subject_key, left(body, 70) AS head, created_at
FROM doc_notes WHERE created_by = '212_tool_crosslink_action_notes' ORDER BY subject_key;
