-- 657_selector_ranks_sites_by_loadable_work.sql
--
-- bugs_open/413: the dispatch selector and the item loader disagree on ordering, so one
-- pinned item freezes its site's age and starves every younger site of trigger dispatch.
--
-- ── THE CONTRACT THIS FILE ENFORCES ──
--
-- A selector must represent a container only by work its drainer will actually take:
-- same eligibility filters, same ordering, same window. Migration 285 established the
-- eligibility half ("the selector's job is to AGREE with the loader"); this file extends
-- the agreement to ORDERING and WINDOW, which 285 left open and 284's header wrongly
-- called bounded ("at most ceil(backlog/5) batches" — refuted by 413: the bound assumed
-- no better-priority inflow; with inflow, the pin never comes within reach of the top-5
-- and the bound is infinite).
--
-- ── WHAT IT DOES ──
--
--   build-pipeline-trigger > find_dispatchable_site (config.query — NOT `pre_query`;
--   633's trap) is rewritten. OLD: rank sites by the age of their single globally-oldest
--   eligible row (`ORDER BY wi.created_at ASC, wi.priority ASC, wi.id ASC LIMIT 1`) —
--   an old worst-priority row wins selection for its site for ever while the loader
--   (priority-major, top-max_items) never takes it. NEW: rank sites by
--   min(created_at) over each site's top-K eligible rows under THE LOADER'S OWN ordering
--   (`priority ASC, created_at ASC` — load_work_item_actions.go:789), where K is read
--   LIVE from build-dispatch-loop > load_items > max_items. The pin becomes
--   unrepresentable: a site's claim to age is exactly the work its next pick will drain.
--
--   K is dynamic BY AGREEMENT with migration 658 (Phase 3, batch 5→8): 658's header
--   defers its selector lockstep to this file, and this file reads 658's knob live, so
--   neither owes the other an edit when the knob moves again. If the K path ever fails
--   to resolve, COALESCE degrades to K=1 — still pin-free (the site is represented by
--   the one row the loader takes first), never a fleet dispatch stop.
--
--   Eligibility clauses are byte-identical to the previous text — including the
--   CROSS-SITE spelling of the lock exception. ⚠⚠ Do NOT "DRY" it against the per-site
--   Go fragment: work_items_common.go:851-870 explains why the two spellings must stay
--   different. The output shape (one row: site_id::text, domain) is unchanged — the
--   check_has_site / spawn_dispatch steps depend on it (284's header §32-35).
--
-- ── WHY _HOLD (ordering, not safety) ──
--
--   The dispatch_throughput lane's 24h post-B read lands ~09:00Z 2026-08-27 and Phase 3
--   (658) applies ~09:30Z. This file changes site-selection behaviour, so applying it
--   earlier destroys the attribution between ruling B, batch 8 and the selector rework.
--
--   DO NOT APPLY BEFORE 12:00Z 2026-08-27, and only after BOTH boundaries above are
--   stamped (agreed with the dispatch_throughput session 2026-08-26 ~20:4xZ). If the
--   24h read fails its gate, re-coordinate with that lane first. After applying:
--   stamp the apply time in bugs_open/413 and tell the dispatch_throughput lane so they
--   cut their measurement windows on it.
--
--   Numbers record creation order, not apply order: 658 applies BEFORE 657 by design.
--
-- ── APPLY (by hand; the runner's SIDECAR_RE never auto-applies a _HOLD) ──
--
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - \
--     < docs/agent_docs/sql_for_agents/657_selector_ranks_sites_by_loadable_work.sql
--
--   Then run the VERIFY (657_selector_ranks_sites_by_loadable_work_VERIFY.sql) —
--   it must PASS after apply and FAILS BY DESIGN before it (its md5 arm is the proof it
--   can fire). Acceptance meter over the following hours: RUNBOOK §"Per-site starvation
--   floor" — no site with eligible work goes > ~1h unserved while pinned rows exist
--   elsewhere.
--
-- ── REVISION (council round 1, corr ecf2e542, gating objection from debug_historian) ──
--
--   The K subquery originally selected the build-dispatch-loop row with
--   ORDER BY updated_at DESC. WRONG: the runtime loader (`loadAgentDefinition`,
--   platform/messaging/processor.go:371-389, and `loadAgentDefinitionForAction`,
--   ai_actions.go:1400-1412) selects `ORDER BY version DESC LIMIT 1`, and
--   `agent_definitions.updated_at` is documented DEGENERATE (199/200 live rows share
--   one microsecond) — so under a duplicate-active-row shape the old form could read
--   K from a row the runtime never loads. The K subquery now mirrors the loader's
--   selection rule VERBATIM (same filters, ORDER BY version DESC). Additionally the
--   preflight now asserts exactly ONE active row per type and the UPDATE is scoped by
--   the captured row id, not by type — so a duplicate-active build-pipeline-trigger
--   row is refused loudly instead of md5-checking one row while the runtime executes
--   the other (debug_historian's second objection).
--
-- ── ROUND-2 ADVISORY ACTED ON (guardian, medium — approved verdict, same corr) ──
--
--   The bare ::int cast on max_items would THROW on a malformed (non-numeric) value,
--   halting dispatch FLEET-WIDE on the next fire — a larger blast radius than the bug
--   being fixed, and 658's write-side jsonb_typeof guard only constrains 658's own
--   writer. The K expression is now total: a regex guard (~ '^[0-9]+$') turns any
--   malformed value into NULL, COALESCE turns NULL into 1, and GREATEST(...,1) turns a
--   literal 0 into 1 (K=0 would silently select no site ever). Every failure direction
--   lands on K>=1, which is pin-free and cannot stop dispatch. Note the deliberate
--   divergence from the loader's own failure mode (GetIntField falls back to 50 on type
--   mismatch, load_work_item_actions.go:699): the selector falling to K=1 while the
--   loader runs 50 keeps K <= M, which is the safe direction of the contract.
--
-- ROLLBACK: 657_selector_ranks_sites_by_loadable_work_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('build-pipeline-trigger', '657_selector_ranks_sites_by_loadable_work.sql: pre-update');

DO $mig$
DECLARE
    -- md5 of the query text this file replaces (read from the live row 2026-08-26
    -- ~20:2xZ; last written by 633). A drifted row is refused, not blindly overwritten.
    v_old_md5 CONSTANT text := 'd6f98acdb5aec385d5eb4077eac530fc';
    v_new     CONSTANT text := $q$WITH k AS (SELECT GREATEST(COALESCE((SELECT CASE WHEN ad.default_config->'workflow'->'steps'->'load_items'->'config'->>'max_items' ~ '^[0-9]+$' THEN (ad.default_config->'workflow'->'steps'->'load_items'->'config'->>'max_items')::int END FROM agent_definitions ad WHERE ad.type = 'build-dispatch-loop' AND ad.is_active AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL ORDER BY ad.version DESC LIMIT 1), 1), 1) AS n), elig AS (SELECT wi.id, wi.site_id, wi.created_at, wi.priority FROM site_work_items wi JOIN sites s ON s.id = wi.site_id WHERE (s.locked_at IS NULL OR wi.id = ANY(COALESCE(s.lock_except_item_ids, ARRAY[]::uuid[]))) AND wi.status IN ('triaged', 'approved') AND wi.attempt_count < wi.max_attempts AND (wi.retry_after IS NULL OR wi.retry_after <= NOW()) AND (COALESCE(wi.approval_mode, 'auto') = 'auto' OR wi.status = 'approved') AND (wi.depends_on IS NULL OR NOT EXISTS (SELECT 1 FROM unnest(wi.depends_on) dep_id WHERE dep_id NOT IN (SELECT id FROM site_work_items WHERE site_id = wi.site_id AND status IN ('complete', 'verified')))) AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND active.status = 'claimed')), win AS (SELECT e.site_id, e.created_at, row_number() OVER (PARTITION BY e.site_id ORDER BY e.priority ASC, e.created_at ASC) AS load_rank FROM elig e) SELECT w.site_id::text, s.domain FROM win w JOIN sites s ON s.id = w.site_id, k WHERE w.load_rank <= k.n GROUP BY w.site_id, s.domain ORDER BY MIN(w.created_at) ASC, w.site_id ASC LIMIT 1$q$;
    v_q      text;
    v_k      int;
    v_n      int;
    v_cnt    int;
    v_row_id uuid;
BEGIN
    -- PRECONDITION 0: exactly ONE active row per type. Four agent types carry TWO
    -- active rows (LANDMINES), the runtime resolves them by version DESC, and
    -- updated_at is degenerate — a type-scoped md5 check against a duplicate shape
    -- could pass on one row while the runtime executes the other (council ecf2e542,
    -- debug_historian). Refuse loudly; a duplicate here means re-derive by hand.
    SELECT count(*) INTO v_cnt FROM agent_definitions ad
     WHERE ad.type='build-pipeline-trigger' AND ad.is_active
       AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
    IF v_cnt <> 1 THEN
        RAISE EXCEPTION '657: % active build-pipeline-trigger rows (expected 1) — duplicate-active-row shape; the runtime loads only the highest version; STOP and re-derive', v_cnt;
    END IF;
    SELECT count(*) INTO v_cnt FROM agent_definitions ad
     WHERE ad.type='build-dispatch-loop' AND ad.is_active
       AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
    IF v_cnt <> 1 THEN
        RAISE EXCEPTION '657: % active build-dispatch-loop rows (expected 1) — K would be read from an ambiguous shape; STOP and re-derive', v_cnt;
    END IF;

    -- PRECONDITION 1: the step exists and carries exactly the text this file expects.
    -- Capture the row id so the UPDATE below is pinned to the row the md5 was checked
    -- on, never re-resolved by type.
    SELECT ad.id, s.value->'config'->>'query' INTO v_row_id, v_q
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.type='build-pipeline-trigger' AND s.key='find_dispatchable_site'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;

    IF v_q IS NULL THEN
        RAISE EXCEPTION '657: find_dispatchable_site has no config.query — the step shape has changed; STOP and re-read it';
    END IF;
    IF md5(v_q) <> v_old_md5 THEN
        IF position('load_rank' in v_q) > 0 THEN
            RAISE EXCEPTION '657: already applied (query already ranks by load_rank)';
        END IF;
        RAISE EXCEPTION '657: query text has drifted since this file was written (md5 %, expected %) — STOP, re-read the live text and re-derive', md5(v_q), v_old_md5;
    END IF;

    -- PRECONDITION 2: the K path must resolve on the live loader row, selected the way
    -- THE RUNTIME selects it (loadAgentDefinition, processor.go:371-389: version DESC).
    -- COALESCE(...,1) in the query makes a later breakage degrade safely, but APPLYING
    -- against a path that already fails means the agreement with 658 is broken — stop.
    SELECT (ad.default_config->'workflow'->'steps'->'load_items'->'config'->>'max_items')::int
      INTO v_k
      FROM agent_definitions ad
     WHERE ad.type='build-dispatch-loop' AND ad.is_active
       AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
     ORDER BY ad.version DESC LIMIT 1;
    IF v_k IS NULL OR v_k < 1 THEN
        RAISE EXCEPTION '657: build-dispatch-loop load_items.max_items does not resolve to a positive int (got %) — the K agreement with 658 is broken; STOP', v_k;
    END IF;

    UPDATE agent_definitions ad
       SET default_config = jsonb_set(
             ad.default_config,
             '{workflow,steps,find_dispatchable_site,config,query}',
             to_jsonb(v_new),
             false)
     WHERE ad.id = v_row_id;
    GET DIAGNOSTICS v_n = ROW_COUNT;
    IF v_n <> 1 THEN
        RAISE EXCEPTION '657: expected to update exactly 1 build-pipeline-trigger row (id %), updated %', v_row_id, v_n;
    END IF;
END;
$mig$;

-- ---------------------------------------------------------------------------
-- GUARDS — read back and assert; DO/RAISE because a block of SELECTs cannot
-- stop a COMMIT. The last guard EXECUTEs the stored text against the live
-- schema: a query that does not parse or run must not be committed.
-- ---------------------------------------------------------------------------
DO $guard$
DECLARE
    v_q      text;
    v_frag   text;
    v_site   text;
    v_domain text;
BEGIN
    SELECT s.value->'config'->>'query' INTO v_q
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.type='build-pipeline-trigger' AND s.key='find_dispatchable_site'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;

    IF position('load_rank' in COALESCE(v_q,'')) = 0 THEN
        RAISE EXCEPTION '657 GUARD: the new query is not in place (no load_rank window)';
    END IF;
    -- The window must mirror the loader's ordering VERBATIM (load_work_item_actions.go:789).
    IF position('ORDER BY e.priority ASC, e.created_at ASC' in v_q) = 0 THEN
        RAISE EXCEPTION '657 GUARD: the window does not mirror the loader ordering (priority ASC, created_at ASC)';
    END IF;
    -- No eligibility clause may be lost — each widens dispatch if dropped. The four
    -- OR-bearing fragments are pinned WITH their wrapping parens (hardened 2026-08-27,
    -- from the deferred_work_item_park lane's CONTRIB): AND binds tighter than OR, so a
    -- paren drop widens dispatch WITHOUT dropping anything — measured 1,104 -> 15,683
    -- admitted rows on fragment 1 — and a bare-presence test cannot see it. A substring
    -- still cannot prove the parens BALANCE; the leading '(' catches the realistic edit
    -- (wholesale drop), and the VERIFY's md5 arm pins the live text byte-exactly.
    FOREACH v_frag IN ARRAY ARRAY[
        '(s.locked_at IS NULL OR wi.id = ANY(COALESCE(s.lock_except_item_ids, ARRAY[]::uuid[])))',
        'wi.status IN (''triaged'', ''approved'')',
        'wi.attempt_count < wi.max_attempts',
        '(wi.retry_after IS NULL OR wi.retry_after <= NOW())',
        '(COALESCE(wi.approval_mode, ''auto'') = ''auto'' OR wi.status = ''approved'')',
        '(wi.depends_on IS NULL OR NOT EXISTS',
        'active.status = ''claimed'''
    ] LOOP
        IF position(v_frag in v_q) = 0 THEN
            RAISE EXCEPTION '657 GUARD: eligibility clause LOST from the new query: %', v_frag;
        END IF;
    END LOOP;

    -- EXECUTION PROBE: the stored text must run as-is (it takes no parameters) and
    -- return at most one (site_id, domain) row. INTO takes the first row; zero
    -- eligible sites yields NULLs, which is a valid quiet-fleet result.
    EXECUTE v_q INTO v_site, v_domain;
    RAISE NOTICE '657 probe: next pick under the new ordering = % (%)',
        COALESCE(v_domain, '<no eligible site>'), COALESCE(v_site, '-');

    RAISE NOTICE '657 OK: selector now ranks sites by min(created_at) over their top-K loadable rows; K read live from build-dispatch-loop.';
END;
$guard$;

COMMIT;
