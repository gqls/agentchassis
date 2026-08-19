-- 473 — bug 184: make a page-rerender the mechanical repair for literal_markdown.
--
-- Two config changes on the live page-rerender row:
--   1. check_rerender_mode.condition gains OR spec.reason == 'literal_markdown',
--      so a literal_markdown work item (whose spec carries that reason as of the
--      184 fix-2 check change, commit 763bb5d55) takes the sections branch
--      instead of the assemble-only else branch.
--   2. rerender_sections.config.strip_literal_markdown = true — the opt-in flag
--      (default OFF in code) that makes RerenderPageSectionsAction pass each
--      stored section's content_data through
--      datahelpers.StripLiteralMarkdownFromContentData before it feeds both the
--      render context and the persisted mergedContent. The action reads it from
--      params.StepConfig.Config, which is exactly where the engine delivers this
--      path (coordinator.go:1696 StepConfig: step) — the same source the step's
--      existing keys (reason/page_name/target_site_id) are read from in
--      production.
--
-- SCOPE (stated per the bugfix_277 lane's CONTRIB of 2026-08-19, measured, and
-- correct): this migration repairs literal_markdown on GENERIC pages only. On
-- OWNED pages (rebuild_policy='owned') the sections branch this opens reaches
-- save_page_sections' owned-page guard and is REFUSED — by design, and the
-- refusal is BY CONSTRUCTION: this migration is precisely what moves the
-- population off the assemble-only branch onto the branch that calls the save.
-- The owned-page residual is bugs_open/301's (route repairs through the owning
-- pipeline), NOT evidence this migration failed. At apply time the verify
-- block RAISES NOTICE with the open items' generic/owned split so the residual
-- is a recorded expectation. (pages.rebuild_policy is mutable — an at-apply
-- count is honest; any RETROSPECTIVE split must use the run's own error text,
-- error LIKE '%rebuild_policy=owned%', not the column.) Every no-write path an
-- owned page can take (guard refusal, resolve-miss skip, no-content_data
-- escalation) still ends honestly: VerifyLiteralMarkdownResolved gates every
-- completion and refuses a page that still scans dirty — nothing false-greens.
--
-- ORDERING: safe to apply BEFORE the image carrying the 184 fix-2 code —
--   - the extra OR clause references a spec.reason value nothing emits until
--     the re-routed check ships, so no dispatch can reach it early;
--   - strip_literal_markdown is unread by the old binary.
-- The check re-route ships only WITH the strip hook (both at HEAD since
-- 757a0890a + 763bb5d55, one image), so there is no window in which items
-- dispatch to an unequipped handler.
--
-- Idempotent: the UPDATE is needle-gated on the condition not already carrying
-- the literal; a re-run is a 0-row no-op and the verify still passes on final
-- state. Verify is DO/RAISE (a SELECT-only verify cannot stop the COMMIT —
-- RFC_006 landmine). A drifted anchor cannot silently no-op: the UPDATE's WHERE
-- requires the exact template_changed literal, so drift → 0 rows → the DO block
-- RAISES because the literal never appears.
--
-- Backup: snapshot_agent() (the estate's standard pre-update idiom), NOT a
-- bespoke side-table — council round 1 (corr 060bcc0a) reuse objection, upheld.
-- The rollback is a surgical inverse (removes exactly what this added), so
-- intervening migrations on the same row survive a rollback untouched.
--
-- NULL-DIRECTION ANALYSIS of the verify (council r3, editquality/debug_historian
-- — the jsonb <>-vs-NULL landmine): every comparison below fails LOUD on an
-- absent key, never green. `cond` is explicitly guarded (`IF cond IS NULL OR
-- cond NOT LIKE ...` → RAISE); the flag check is a POSITIVE-presence EXISTS
-- (`(#> path)::text = 'true'` — an absent path yields NULL, NULL = 'true' is
-- not true, the row is not selected, NOT EXISTS → RAISE). The landmine's trap
-- is negative-form verifies (`<> 'bad'` passing on NULL); none is used here.
-- jsonb_set with a missing PARENT path returns its input unchanged — that
-- silent no-op is also caught, because the final-state check above would then
-- find no flag and RAISE.

BEGIN;

SELECT snapshot_agent('page-rerender',
                      '473_literal_markdown_mechanical_repair.sql: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           default_config,
           '{workflow,steps,check_rerender_mode,config,condition}',
           to_jsonb(replace(
             default_config #>> '{workflow,steps,check_rerender_mode,config,condition}',
             'input_data.spec.reason == ''template_changed''',
             'input_data.spec.reason == ''template_changed'' OR input_data.spec.reason == ''literal_markdown'''
           ))
         ),
         '{workflow,steps,rerender_sections,config,strip_literal_markdown}',
         'true'::jsonb
       ),
       updated_at = now()
 WHERE type = 'page-rerender' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #>> '{workflow,steps,check_rerender_mode,config,condition}'
       NOT LIKE '%literal_markdown%'
   -- anchor: refuse to fire if the live literal this replace targets has moved
   AND default_config #>> '{workflow,steps,check_rerender_mode,config,condition}'
       LIKE '%input_data.spec.reason == ''template_changed''%';

DO $$
DECLARE cond text;
BEGIN
  SELECT default_config #>> '{workflow,steps,check_rerender_mode,config,condition}'
    INTO cond
    FROM agent_definitions
   WHERE type = 'page-rerender' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF cond IS NULL OR cond NOT LIKE '%literal_markdown%' THEN
    RAISE EXCEPTION '473 FAILED: check_rerender_mode.condition does not carry literal_markdown — the template_changed anchor has moved; read the live condition and re-anchor. Live: %', cond;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM agent_definitions
     WHERE type = 'page-rerender' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND (default_config #> '{workflow,steps,rerender_sections,config,strip_literal_markdown}')::text = 'true'
  ) THEN
    RAISE EXCEPTION '473 FAILED: rerender_sections.config.strip_literal_markdown is not true';
  END IF;

  -- the other four reasons must have survived the string surgery
  IF cond NOT LIKE '%image_landed%' OR cond NOT LIKE '%section_data_resolved%'
     OR cond NOT LIKE '%cta_links_stale%' OR cond NOT LIKE '%template_changed%' THEN
    RAISE EXCEPTION '473 FAILED: an existing reason clause was damaged. Live: %', cond;
  END IF;

  RAISE NOTICE '473 OK: condition = %', cond;
END $$;

-- Scope record (see SCOPE in the header): the open literal_markdown population
-- by page ownership AT APPLY TIME. The owned count is the EXPECTED residual
-- this migration does not repair (bugs_open/301's remit).
DO $$
DECLARE n_generic int; n_owned int; n_orphan int;
BEGIN
  SELECT count(*) FILTER (WHERE COALESCE(p.rebuild_policy,'generic') <> 'owned' AND p.id IS NOT NULL),
         count(*) FILTER (WHERE p.rebuild_policy = 'owned'),
         count(*) FILTER (WHERE p.id IS NULL)
    INTO n_generic, n_owned, n_orphan
    FROM site_work_items swi LEFT JOIN pages p ON p.id = swi.page_id
   WHERE swi.item_type = 'literal_markdown'
     AND swi.status NOT IN ('complete','cancelled','rejected');
  RAISE NOTICE '473 SCOPE at apply: open literal_markdown items — % generic (repairable via this route), % owned (EXPECTED residual, bugs_open/301), % with no page row', n_generic, n_owned, n_orphan;
END $$;

COMMIT;
