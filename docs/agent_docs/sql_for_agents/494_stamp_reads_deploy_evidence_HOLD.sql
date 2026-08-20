-- 494_stamp_reads_deploy_evidence_HOLD.sql
--
-- bugs_open/315 / RFC_038. Turns ON the deploy-evidence guard for the three
-- agents that stamp pages.deployed_at immediately after a git_commit whose
-- result they currently discard.
--
-- ############################################################################
-- ##  _HOLD — DO NOT APPLY UNTIL THE CHASSIS IMAGE CARRYING                  ##
-- ##  `deploy_result_field` HAS ROLLED. THIS IS A HARD ORDERING              ##
-- ##  CONSTRAINT, AND IT RUNS THE OPPOSITE WAY TO THE USUAL SEED RULE.       ##
-- ############################################################################
--
-- WHY. `UpdatePageStatusInputSpec` is `StrictConfig: true`
-- (v3_site_actions.go). An unrecognised config key does not no-op — it FAILS
-- VALIDATION. So applying this against a chassis whose binary does not declare
-- `deploy_result_field` breaks `update_page_status` on every page of all three
-- agents at once. Config is live immediately; Go is not. Hence the hold.
--
-- The `_HOLD` suffix is load-bearing: the runner's SIDECAR_RE
-- (`_[A-Z][A-Z0-9_]*\.sql$`) excludes this file from `--apply` while STILL
-- listing it under "Sidecars (hand-run only)", so it is held back visibly
-- rather than silently. A banner alone would not hold it — the runner does not
-- read comments.
--
-- APPLY IT BY HAND, in this order, and the person doing it will not be me:
--
--   1. Confirm the running chassis carries the key. Ask the ARTEFACT, not git:
--        kubectl -n ai-persona-system get pods -l app=agent-chassis \
--          -o jsonpath='{range .items[*]}{.spec.containers[0].image}{"\n"}{end}'
--        kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 \
--          | grep -m1 'build provenance'
--        git merge-base --is-ancestor daaa7541b <the stamped sha> && echo SAFE-TO-ARM
--
--      ⚠⚠ THE GATE COMMIT IS NOW daaa7541b, AND THE TWO EARLIER ANSWERS WERE
--      BOTH WRONG. This migration was armed at 06:49:49Z on 2026-08-20 with
--      the f0dd97c71 precondition MET, and it BROKE THE FLEET'S ENTIRE
--      PAGE-PUBLISHING PATH for 33 minutes: `deploy_result_field` had been
--      declared on RenderComponentInputSpec instead of UpdatePageStatusInputSpec
--      (same file, forty lines apart), and because the reader's spec is
--      StrictConfig:true an undeclared key is a HARD VALIDATION FAILURE for the
--      whole workflow. 8 items failed, 123 page_rerender items queued and did
--      not drain. See bugs_open/336; service was restored by running this
--      file's own ROLLBACK. daaa7541b moves the declaration to the right spec.
--
--      THE LESSON FOR WHOEVER ARMS THIS: the earlier preconditions were about
--      the READER shipping. They cannot see a declaration on the wrong spec.
--      So check the LIST, inside the RIGHT struct, at the commit the running
--      binary was built from:
--        git show <stamp>:platform/orchestration/actions/v3_site_actions.go \
--          | awk '/^var UpdatePageStatusInputSpec/,/^}/' | grep deploy_result_field
--      A binary grep for the literal CANNOT see this and WILL mislead you: the
--      string is in the chassis three times over (the reader and two zap calls),
--      so it reads PRESENT even when the declaration is on the wrong spec.
--
--      AND AFTER ARMING, THE FIRST QUERY IS 'WHAT DID I BREAK?', NOT 'DID IT
--      WORK?':
--        SELECT count(*) FROM orchestration_states WHERE error ILIKE '%deploy_result_field%';
--        SELECT status, count(*) FROM site_work_items WHERE item_type='page_rerender' GROUP BY status;
--      A zero in `count(pages.content_hash)` looks the same whether nothing has
--      run or nothing CAN run. That is how the 33 minutes happened.
--
--      ⚠ THE REQUIRED COMMIT IS f0dd97c71, NOT 086f9b7b7. 086f9b7b7 declared the
--      key and SHIPPED IN v1.0.1316 (17:13Z 2026-08-19) — but the council gate's
--      round-2 review (corr 377167cd, prior_art_librarian, gating) found its
--      resolver unsafe: it borrowed a "unique-or-nothing" guarantee from
--      datahelpers.ExtractFields that RFC_029 Phase 1 does not actually provide
--      (Phase 1 still resolves conflicts to a shallowest-first winner; Phase 2
--      flips them to refusal and has not shipped). Against v1.0.1316 an
--      ambiguous subtree would fingerprint the page from an ARBITRARY
--      git_commit, which is silently and permanently wrong. f0dd97c71 makes the
--      resolver collect candidates itself and REFUSE on conflict.
--      ARMING THIS AGAINST v1.0.1316 IS THE ONE THING THAT MAKES IT DANGEROUS.
--
--      An EMPTY provenance grep means "not in log range", NOT "unstamped" — it
--      is a startup line and it scrolls on a busy service. The git-adapter is
--      quiet and usually still has it; the chassis usually does not. Do NOT
--      fall back to `grep -a <sha> /proc/1/exe`: measured 2026-08-19, that
--      returned ABSENT for a commit the image demonstrably carried while
--      returning PRESENT for a 40-zero control (Go's internal tables).
--
--   2. Then, scoped so it cannot sweep other threads' pending files (the
--      assignment MUST be on the same line as the command):
--        mkdir -p /tmp/mig494 && cp docs/agent_docs/sql_for_agents/494_*.sql /tmp/mig494/
--        MIGRATIONS_DIR=/tmp/mig494 ./scripts/migration/run-migrations.sh --no-probe
--        kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--          psql -U clients_user -d clients_db < docs/agent_docs/sql_for_agents/494_stamp_reads_deploy_evidence_HOLD.sql
--      (a _HOLD file is a sidecar, so the runner will not apply it; pipe it.)
--      ⚠ DO NOT bother with --record-only: the runner REFUSES a sidecar
--      ("'..._HOLD.sql' is an UPPERCASE-suffixed sidecar ... recording one is
--      meaningless"). Measured 2026-08-19 — an earlier version of this header
--      told you to do it and it does not work. That is harmless: a sidecar
--      never appears in Pending, so the runner cannot double-apply it, and the
--      already-applied guard above (RAISE '494: already applied') catches a
--      human re-run. Record the apply in the lane's NOTES instead.
--
-- WHAT TURNING IT ON DOES. For these three steps the action will, before
-- stamping: refuse the stamp when the deploy step reported it skipped, and on
-- success record the sha256 of the bytes committed into pages.content_hash.
-- With no evidence resolvable it stamps anyway and writes a
-- DEPLOY_EVIDENCE_UNREADABLE row — expected while the GIT-ADAPTER image is
-- older than RFC_038, since chassis and adapter are separate images.
--
-- FIELD NAMES DIFFER PER AGENT AND THAT IS THE POINT: the 19 live git_commit
-- steps carry NINE distinct output_field values. section-editor's is
-- `git_result`, not `deploy_result`. Getting one wrong does not fail loudly —
-- it resolves nothing and fails open.
--
-- ROLLBACK: 494_stamp_reads_deploy_evidence_HOLD_ROLLBACK.sql (removes the key,
-- restoring today's behaviour byte for byte).

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('page-rerender',   '494_stamp_reads_deploy_evidence: pre-update');
SELECT snapshot_agent('report-builder',  '494_stamp_reads_deploy_evidence: pre-update');
SELECT snapshot_agent('section-editor',  '494_stamp_reads_deploy_evidence: pre-update');

-- Already-applied arm (the runner reads a RAISE containing 'already').
DO $$
DECLARE done int;
BEGIN
    SELECT count(*) INTO done FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND ( (type='page-rerender'  AND default_config->'workflow'->'steps'->'update_status'->'config'      ? 'deploy_result_field')
          OR (type='report-builder' AND default_config->'workflow'->'steps'->'update_status'->'config'      ? 'deploy_result_field')
          OR (type='section-editor' AND default_config->'workflow'->'steps'->'update_page_status'->'config' ? 'deploy_result_field') );
    IF done = 3 THEN
        RAISE EXCEPTION '494: already applied — all three steps already name a deploy_result_field';
    END IF;
END $$;

-- COUNTED NEEDLE-GATE. Each stamping step must still be preceded by the
-- git_commit step whose output_field we are about to name, with that exact
-- name. Anything else and the premise has moved — abort rather than wire a
-- guard to a field nothing writes.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND ( (type='page-rerender'
              AND default_config->'workflow'->'steps'->'deploy_page'->>'output_field' = 'deploy_result'
              AND default_config->'workflow'->'steps'->'deploy_page'->>'next_step'    = 'update_status')
          OR (type='report-builder'
              AND default_config->'workflow'->'steps'->'deploy_page'->>'output_field' = 'deploy_result'
              AND default_config->'workflow'->'steps'->'deploy_page'->>'next_step'    = 'update_status')
          OR (type='section-editor'
              AND default_config->'workflow'->'steps'->'deploy_page'->>'output_field' = 'git_result'
              AND default_config->'workflow'->'steps'->'deploy_page'->>'next_step'    = 'update_page_status') );
    IF n <> 3 THEN
        RAISE EXCEPTION
          '494 needle-gate: expected 3 agents whose git_commit step feeds the stamp under the named output_field, found % — re-derive against the live workflow', n;
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,update_status,config,deploy_result_field}', '"deploy_result"'),
       updated_at = NOW()
 WHERE type IN ('page-rerender','report-builder')
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,update_page_status,config,deploy_result_field}', '"git_result"'),
       updated_at = NOW()
 WHERE type = 'section-editor'
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

DO $$
DECLARE bad int;
BEGIN
    SELECT count(*) INTO bad FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND ( (type IN ('page-rerender','report-builder')
              AND default_config->'workflow'->'steps'->'update_status'->'config'->>'deploy_result_field'
                  IS DISTINCT FROM 'deploy_result')
          OR (type='section-editor'
              AND default_config->'workflow'->'steps'->'update_page_status'->'config'->>'deploy_result_field'
                  IS DISTINCT FROM 'git_result') );
    IF bad <> 0 THEN
        RAISE EXCEPTION '494 post-verify: % row(s) did not take the deploy_result_field', bad;
    END IF;
END $$;

COMMIT;
