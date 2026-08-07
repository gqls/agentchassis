-- 331_literal_markdown_claim_timeout_exclusion_ROLLBACK.sql
--
-- Reverses 331: removes 'literal_markdown' from the claimed-item-timeout sweep's
-- item_type exclusion list (scheduled_tasks.pre_query).
--
-- WHY THIS FILE EXISTS AS A SEPARATE ARTEFACT: the council's debug_historian seat
-- objected (corr f14a8b64-4f71-4915-88d0-9587db845052, low) that the needle-gate
-- discipline calls for verify and rollback as DISTINCT artefacts, not inline
-- post-verify DO blocks bundled with the forward migration. Correct — 211's
-- *_ROLLBACK.sql is the house precedent. UPPERCASE-suffixed sidecars are excluded
-- from run-migrations.sh by SIDECAR_RE and are run BY HAND, deliberately.
--
-- WHEN YOU WOULD RUN THIS: only if excluding literal_markdown from the 15-minute
-- auto-complete turns out to strand items rather than protect them — i.e. if
-- VerifyLiteralMarkdownResolved is wrong often enough that items pile up claimed.
-- Note that rolling back re-arms the bypass: the sweep will then auto-complete a
-- literal_markdown item on handler-orchestration evidence alone, walking past the
-- verifier. That is the state bugs_open/201 symptom 2 is about, so prefer fixing or
-- unregistering the verifier over restoring the bypass.
--
-- Same shape as the forward migration: assert before, replace, assert both
-- directions after. A verify block of plain SELECTs cannot stop a COMMIT.

\set ON_ERROR_STOP on

BEGIN;

DO $pre$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM scheduled_tasks
     WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'', ''dead_fragment_link'', ''literal_markdown'')%';
    IF v_rows <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 scheduled_task carrying the 8-entry list 331 installed, found % — the live pre_query has drifted since; read it before rolling back', v_rows;
    END IF;
END $pre$;

UPDATE scheduled_tasks
   SET pre_query = replace(
         pre_query,
         $new$item_type NOT IN ('truncated_component', 'hardcoded_section_colors', 'empty_section', 'orphan_element_refs', 'content_duplication', 'page_canonical_collision', 'dead_fragment_link', 'literal_markdown')$new$,
         $old$item_type NOT IN ('truncated_component', 'hardcoded_section_colors', 'empty_section', 'orphan_element_refs', 'content_duplication', 'page_canonical_collision', 'dead_fragment_link')$old$),
       updated_at = NOW()
 WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'', ''dead_fragment_link'', ''literal_markdown'')%';

DO $verify$
DECLARE v_after int; v_new int;
BEGIN
    -- The 7-entry list is back. The closing paren is load-bearing: it matches only
    -- a list that ENDS at dead_fragment_link.
    SELECT count(*) INTO v_after FROM scheduled_tasks
     WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'', ''dead_fragment_link'')%';
    IF v_after <> 1 THEN
        RAISE EXCEPTION 'rollback did not restore the 7-entry list: % rows carry it', v_after;
    END IF;

    SELECT count(*) INTO v_new FROM scheduled_tasks WHERE pre_query LIKE '%''literal_markdown''%';
    IF v_new <> 0 THEN
        RAISE EXCEPTION 'literal_markdown still present in % row(s) after rollback', v_new;
    END IF;

    RAISE NOTICE 'literal_markdown removed from the claim-timeout exclusion; the 15-minute auto-complete can once again bypass VerifyLiteralMarkdownResolved';
END $verify$;

COMMIT;

-- AFTER RUNNING THIS BY HAND, tell the ledger, or the next --apply will try 331 again:
--   ./scripts/migration/run-migrations.sh --record-only 331_literal_markdown_claim_timeout_exclusion.sql \
--     --note 'rolled back by 331_..._ROLLBACK.sql on <date> because <why>'
-- (run-migrations.sh does not un-record; deleting the schema_migrations row is the
--  other option and is the one that makes a re-apply possible.)
