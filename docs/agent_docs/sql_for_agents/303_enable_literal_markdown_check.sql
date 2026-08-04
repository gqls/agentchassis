-- 303_enable_literal_markdown_check.sql — enable the literal_markdown
-- discovery check on quality-discovery-agent (bugs_open/184: markdown syntax
-- reaching live pages as literal asterisks/backticks/hashes).
--
-- Emits literal_markdown items (one per page, item_key
-- literal_markdown:<page_id>), status detected, handler page-content-writer
-- (auto-repair — a definite mechanical defect, the placeholder_contact
-- routing). Dual-surface scan (content_data + rendered_html, the
-- unverified_claims/093 precedent). Locked components skipped.
--
-- ORDER: apply AFTER the chassis image carrying check_literal_markdown.go is
-- live (image -> seed). Since bugs_open/149 B4 an unregistered check name
-- FAILS the run_discovery_checks step loudly rather than being skipped.

BEGIN;

SELECT snapshot_agent('quality-discovery-agent', '303_enable_literal_markdown_check: pre-update');

DO $$
DECLARE
  checks jsonb;
BEGIN
  SELECT default_config #> '{workflow,steps,run_checks,config,checks}' INTO checks
  FROM agent_definitions
  WHERE type = 'quality-discovery-agent' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF checks IS NULL THEN
    RAISE EXCEPTION '303: no active quality-discovery-agent with run_checks.config.checks';
  END IF;
  IF checks ? 'literal_markdown' THEN
    RAISE EXCEPTION '303: literal_markdown already enabled';
  END IF;

  UPDATE agent_definitions
  SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        checks || '["literal_markdown"]'::jsonb)
  WHERE type = 'quality-discovery-agent' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
END $$;

INSERT INTO doc_notes (id, subject_type, subject_key, body, categories, source, created_by)
VALUES (
  gen_random_uuid(), 'pipeline', 'discovery',
  '## literal_markdown check enabled on quality discovery
Observed: LLM writers emitted markdown (**bold**, `code spans`) into text-typed content_data fields; text/template renders them verbatim — 3 live rows on 3 sites, silent to every existing check (bugs_open/184).
Fix: literal_markdown discovery check appended to quality-discovery-agent run_checks (migration 303, image-first ordering). Items: literal_markdown, detected, handler page-content-writer. Companion: migration 304 hardens the writer prompt (STRICT RULE 9).
Categories: fix, guard-rail',
  '["fix","guard-rail"]'::jsonb,
  'migration', '303_enable_literal_markdown_check'
);

COMMIT;
