-- 657_selector_ranks_sites_by_loadable_work_VERIFY.sql
--
-- Re-runnable check that the 657 contract HOLDS at the live artefact. Run it after
-- applying 657, and alongside the daily 584 VERIFY habit (dispatch_throughput RUNBOOK).
--
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - \
--     < docs/agent_docs/sql_for_agents/657_selector_ranks_sites_by_loadable_work_VERIFY.sql
--
-- Exit 0 = the contract holds; a RAISE names the failing assertion.
--
-- ⚠ BEFORE 657 is applied this file FAILS BY DESIGN on assertion 1 — that failure is the
-- mutation proof that the check can fire. Do not "fix" it by widening the md5 list.
--
-- Assertions:
--   1. exactly one active build-pipeline-trigger row, and its find_dispatchable_site
--      query is byte-identical to 657's text (md5) — drift refused, not diagnosed.
--   2. the selector's window mirrors the loader's ordering VERBATIM
--      (load_work_item_actions.go:789: ORDER BY wi.priority ASC, wi.created_at ASC).
--      If the loader's ORDER BY ever changes, the Go contract test
--      (load_work_items_ordering_contract_test.go) fails first and sends its editor
--      here; this arm is the DB-side half of the same lockstep.
--   3. the K agreement with 658: build-dispatch-loop > load_items > max_items resolves
--      to a positive int on exactly one active row (the path 657 reads live; if this
--      breaks, the selector silently degrades to K=1 — safe but untruthful).
--   4. the stored text EXECUTEs against the live schema and returns a (site_id, domain)
--      pick — reported as a NOTICE with the current pin census (informational; pin
--      status is DYNAMIC, date any figure you quote from it).

DO $verify$
DECLARE
    -- 2026-09-02: 674 (D4 stage B, council 8f4bb57d APPROVED) inserted the spend-governor
    -- clause 'AND governor_admits(wi.item_type)' into this selector — the md5 moved WITH a
    -- reviewed change, which is the ONE sanctioned reason to touch this constant (the
    -- header's do-not-widen warning is about drift, not about lockstep with an approved
    -- edit; council 8f4bb57d r3, guardian advisory, mandated exactly this same-sitting edit).
    v_new_md5 CONSTANT text := 'fcbe8821a2a56512911955735796460e';
    v_rows    int;
    v_q       text;
    v_k       int;
    v_krows   int;
    v_site    text;
    v_domain  text;
    v_pinned  int;
    v_waiting int;
BEGIN
    -- 1. one active row, exact text
    SELECT count(*) INTO v_rows
      FROM agent_definitions ad
     WHERE ad.type='build-pipeline-trigger' AND ad.is_active
       AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '657 VERIFY 1/4: expected exactly 1 active build-pipeline-trigger row, found %', v_rows;
    END IF;

    SELECT s.value->'config'->>'query' INTO v_q
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.type='build-pipeline-trigger' AND s.key='find_dispatchable_site'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
    IF v_q IS NULL THEN
        RAISE EXCEPTION '657 VERIFY 1/4: find_dispatchable_site has no config.query';
    END IF;
    IF md5(v_q) <> v_new_md5 THEN
        RAISE EXCEPTION '657 VERIFY 1/4: live selector text md5 % <> expected % — 657 not applied, rolled back, or drifted', md5(v_q), v_new_md5;
    END IF;

    -- 2. window mirrors the loader ordering
    IF position('ORDER BY e.priority ASC, e.created_at ASC' in v_q) = 0 THEN
        RAISE EXCEPTION '657 VERIFY 2/4: window ordering no longer mirrors load_work_item_actions.go:789';
    END IF;

    -- 3. the K agreement
    SELECT count(*) INTO v_krows
      FROM agent_definitions ad
     WHERE ad.type='build-dispatch-loop' AND ad.is_active
       AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
    IF v_krows <> 1 THEN
        RAISE EXCEPTION '657 VERIFY 3/4: expected exactly 1 active build-dispatch-loop row, found % (K would come from an arbitrary one)', v_krows;
    END IF;
    -- Selected the way THE RUNTIME selects the row (loadAgentDefinition,
    -- processor.go:371-389: version DESC; updated_at is degenerate — council ecf2e542).
    SELECT (ad.default_config->'workflow'->'steps'->'load_items'->'config'->>'max_items')::int
      INTO v_k
      FROM agent_definitions ad
     WHERE ad.type='build-dispatch-loop' AND ad.is_active
       AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
     ORDER BY ad.version DESC LIMIT 1;
    IF v_k IS NULL OR v_k < 1 THEN
        RAISE EXCEPTION '657 VERIFY 3/4: load_items.max_items does not resolve to a positive int (got %) — selector is silently running at K=1', v_k;
    END IF;

    -- 4. execution probe + informational census
    EXECUTE v_q INTO v_site, v_domain;

    WITH elig AS (
      SELECT wi.id, wi.site_id, wi.created_at, wi.priority
      FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
      WHERE (s.locked_at IS NULL OR wi.id = ANY(COALESCE(s.lock_except_item_ids, ARRAY[]::uuid[])))
        AND wi.status IN ('triaged','approved') AND wi.attempt_count < wi.max_attempts
        AND (wi.retry_after IS NULL OR wi.retry_after <= now())
        AND (COALESCE(wi.approval_mode,'auto')='auto' OR wi.status='approved')
        AND (wi.depends_on IS NULL OR NOT EXISTS (
              SELECT 1 FROM unnest(wi.depends_on) dep_id
              WHERE dep_id NOT IN (SELECT id FROM site_work_items
                                   WHERE site_id=wi.site_id AND status IN ('complete','verified'))))),
    ranked AS (
      SELECT e.site_id,
        row_number() OVER (PARTITION BY e.site_id ORDER BY e.created_at, e.priority, e.id) age_rank,
        row_number() OVER (PARTITION BY e.site_id ORDER BY e.priority, e.created_at) load_rank
      FROM elig e)
    SELECT count(DISTINCT site_id),
           count(DISTINCT site_id) FILTER (WHERE age_rank=1 AND load_rank > v_k)
      INTO v_waiting, v_pinned
      FROM ranked;

    RAISE NOTICE '657 VERIFY OK: K=%; next pick=% (%); census: % sites hold eligible work, % pinned (oldest row outside top-K). Pins no longer freeze the age order; quote figures WITH this timestamp: %',
        v_k, COALESCE(v_domain,'<no eligible site>'), COALESCE(v_site,'-'), v_waiting, v_pinned, now();
END;
$verify$;
