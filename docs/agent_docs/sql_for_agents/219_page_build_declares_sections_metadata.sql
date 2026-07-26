-- 219_page_build_declares_sections_metadata.sql
-- Council gate objection (2026-07-26, submission 569241fb, seat bug_historian,
-- MEDIUM): validate_page_content's stat audit (check 9) returned silently when
-- sections_metadata was absent, so on a page that DID need checking there was
-- nothing to distinguish "checked, no fabrication found" from "could not check".
-- The objection is correct and it is the exact shape bugs_open/043 is about.
--
-- The Go half now emits a `stat_audit_unavailable` WARNING when the step
-- declares it expects sections — but only when it declares it, because this
-- gate has four callers and only page-build-handler builds sections:
-- tool-recreation, report-builder and content-reviewer pass a bare HTML blob
-- and must stay silent. Guessing from the payload shape would either
-- false-positive on those three or go quiet on the one that matters, so the
-- expectation is DECLARED here rather than inferred.
--
-- Warning, never error: a missing input is an operational fault, not a content
-- defect, and blocking a deploy on it would punish the page for the pipeline.
--
-- Idempotent: sets one key to the same value on re-run.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('page-build-handler', '219: pre-update (require_sections_metadata)');

DO $d$
DECLARE
    n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type='page-build-handler' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND default_config #> '{workflow,steps,validate_content}' IS NOT NULL;
    IF n < 1 THEN
        RAISE EXCEPTION '219: page-build-handler has no validate_content step — the gate moved, re-derive';
    END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
             '{workflow,steps,validate_content,config,require_sections_metadata}',
             'true'::jsonb, true),
           updated_at = now()
     WHERE type='page-build-handler' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    SELECT count(*) INTO n FROM agent_definitions
    WHERE type='page-build-handler' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND (default_config #>> '{workflow,steps,validate_content,config,require_sections_metadata}') = 'true';
    IF n < 1 THEN
        RAISE EXCEPTION '219: the declaration did not land';
    END IF;
    RAISE NOTICE '219: page-build-handler declares require_sections_metadata — a skipped stat audit is now visible';
END $d$;

INSERT INTO schema_migrations (filename, notes)
VALUES ('219_page_build_declares_sections_metadata.sql',
        'Council objection 569241fb (bug_historian, medium): the stat audit skipped silently when sections_metadata was absent, so "could not check" looked like "checked and clean". page-build-handler now declares it expects sections; the Go half warns (never errors) when the declaration holds and the metadata does not arrive. The other three callers of this gate do not declare it and stay silent.')
ON CONFLICT (filename) DO NOTHING;

COMMIT;
