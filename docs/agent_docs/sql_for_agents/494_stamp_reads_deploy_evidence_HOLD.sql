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
--        git merge-base --is-ancestor 086f9b7b7 <the stamped sha> && echo CARRIES-THE-KEY
--      086f9b7b7 is the commit that declared `deploy_result_field`. An EMPTY
--      grep means "not in log range", NOT "unstamped" — the provenance line is
--      a startup line and scrolls.
--
--   2. Then, scoped so it cannot sweep other threads' pending files (the
--      assignment MUST be on the same line as the command):
--        mkdir -p /tmp/mig494 && cp docs/agent_docs/sql_for_agents/494_*.sql /tmp/mig494/
--        MIGRATIONS_DIR=/tmp/mig494 ./scripts/migration/run-migrations.sh --no-probe
--        kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--          psql -U clients_user -d clients_db < docs/agent_docs/sql_for_agents/494_stamp_reads_deploy_evidence_HOLD.sql
--      (a _HOLD file is a sidecar, so the runner will not apply it; pipe it.)
--      Then record it: ./scripts/migration/run-migrations.sh --record-only \
--          docs/agent_docs/sql_for_agents/494_stamp_reads_deploy_evidence_HOLD.sql \
--          --note "held for the chassis roll; applied by hand after verifying the image"
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
