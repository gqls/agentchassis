-- 633_site_lock_exception_honoured_HOLD.sql
--
-- The CONFIG half of the site-lock exception list. bugs_open/396.
--
-- ⚠⚠ _HOLD, AND THE HOLD IS NOT BUREAUCRACY — APPLYING THIS EARLY UNLOCKS SITES.
--
-- `find_dispatchable_site` selects a SITE, not an item. This file teaches it to
-- select a LOCKED site when one excepted item is dispatchable. The very next step
-- (`build-dispatch-loop > load_items` → `LoadWorkItemsAction`) then loads every
-- dispatchable item on that site — because that loader has NEVER checked
-- `sites.locked_at`. Applied before the binary that carries the loader's
-- `honour_site_lock` arm, this file turns a full site hold into no hold at all,
-- on exactly the sites somebody has deliberately locked.
--
-- ── DO NOT APPLY UNTIL BOTH ARE TRUE ──
--
--   1. `sites.lock_except_item_ids` exists (migration 632, safe to apply any time).
--   2. The running chassis carries `honour_site_lock` in LoadWorkItemsAction.
--      PROVE IT AT THE ARTEFACT, not from git:
--
--        kubectl -n ai-persona-system logs -l app=agent-chassis --tail=3000 \
--          | grep -m1 'build provenance'
--        git merge-base --is-ancestor <the 396 loader commit> <that stamp>
--
--      ⚠ the provenance line is a STARTUP line and scrolls; an empty result means
--      "not in range", NOT "unstamped". Fall back to the binary probe, and run it
--      with a present-control AND an absent-control in the same command:
--
--        kubectl -n ai-persona-system exec <chassis-pod> -- \
--          grep -aq "honour_site_lock" /proc/1/exe   # must be PRESENT
--        kubectl -n ai-persona-system exec <chassis-pod> -- \
--          grep -aq "zzzNotARealSymbol396zzz" /proc/1/exe  # must be ABSENT
--
--      Never `strings` — it is absent from the debian-slim images and behind the
--      customary 2>/dev/null its failure is indistinguishable from "not stamped".
--      Check EVERY replica: a release can straddle revisions.
--
-- ── WHAT IT DOES ──
--
--   (a) build-pipeline-trigger > find_dispatchable_site (config.query — NOT
--       `pre_query`; the step is a `query_database` action and the SQL lives
--       under `query`. Reading the wrong key is how the first draft of this file
--       patched nothing and reported success).
--         WHERE s.locked_at IS NULL
--       becomes
--         WHERE (s.locked_at IS NULL OR wi.id = ANY(COALESCE(s.lock_except_item_ids, ARRAY[]::uuid[])))
--
--   (b) build-dispatch-loop > load_items (config) gains `honour_site_lock: true`
--       — the opt-in field whose unsafe default is OFF (owner ruling 2026-08-02,
--       RFC_010 §2). It is set HERE, on the caller, so a reviewer of the caller
--       sees the decision. `site-work-orchestrator`'s two `load_work_items` steps
--       are deliberately NOT opted in: that flow is human-initiated and gates on
--       nothing today, and changing it is a separate decision.
--
-- Both edits are in ONE file on purpose: (a) without (b) is the unlock described
-- above, and (b) without (a) is inert. They must land together.
--
-- ROLLBACK: 633_site_lock_exception_honoured_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('build-pipeline-trigger', '633_site_lock_exception_honoured_HOLD.sql: pre-update');
SELECT snapshot_agent('build-dispatch-loop',    '633_site_lock_exception_honoured_HOLD.sql: pre-update');

DO $mig$
DECLARE
    v_old_anchor CONSTANT text := 'WHERE s.locked_at IS NULL AND wi.status IN';
    v_new_clause CONSTANT text := 'WHERE (s.locked_at IS NULL OR wi.id = ANY(COALESCE(s.lock_except_item_ids, ARRAY[]::uuid[]))) AND wi.status IN';
    v_q          text;
    v_n          int;
BEGIN
    -- PRECONDITION 1: the column must exist, or the new query references nothing.
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                    WHERE table_name='sites' AND column_name='lock_except_item_ids') THEN
        RAISE EXCEPTION '633: sites.lock_except_item_ids is missing — apply 632 first';
    END IF;

    -- PRECONDITION 2: the anchor must be present EXACTLY once, or the step has
    -- drifted since this file was written and a blind replace would corrupt it.
    SELECT s.value->'config'->>'query' INTO v_q
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.type='build-pipeline-trigger' AND s.key='find_dispatchable_site'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;

    IF v_q IS NULL THEN
        RAISE EXCEPTION '633: find_dispatchable_site has no config.query — the step shape has changed; STOP and re-read it';
    END IF;
    IF position(v_old_anchor in v_q) = 0 THEN
        IF position('lock_except_item_ids' in v_q) > 0 THEN
            RAISE EXCEPTION '633: already applied (query already names lock_except_item_ids)';
        END IF;
        RAISE EXCEPTION '633: anchor % not found in find_dispatchable_site.query — it has drifted; STOP', v_old_anchor;
    END IF;

    -- (a) the site gate
    UPDATE agent_definitions ad
       SET default_config = jsonb_set(
             ad.default_config,
             '{workflow,steps,find_dispatchable_site,config,query}',
             to_jsonb(replace(v_q, v_old_anchor, v_new_clause)))
     WHERE ad.type='build-pipeline-trigger'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
    GET DIAGNOSTICS v_n = ROW_COUNT;
    IF v_n <> 1 THEN
        RAISE EXCEPTION '633: expected to update exactly 1 build-pipeline-trigger row, updated %', v_n;
    END IF;

    -- (b) the loader opt-in, on the caller
    UPDATE agent_definitions ad
       SET default_config = jsonb_set(
             ad.default_config,
             '{workflow,steps,load_items,config,honour_site_lock}',
             'true'::jsonb)
     WHERE ad.type='build-dispatch-loop'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
       AND ad.default_config->'workflow'->'steps'->'load_items' IS NOT NULL;
    GET DIAGNOSTICS v_n = ROW_COUNT;
    IF v_n <> 1 THEN
        RAISE EXCEPTION '633: expected to update exactly 1 build-dispatch-loop row, updated %', v_n;
    END IF;
END;
$mig$;

-- ---------------------------------------------------------------------------
-- GUARDS — read back with a query this file does not contain, and assert BOTH
-- halves plus a negative control. DO/RAISE, because ON_ERROR_STOP ignores a
-- non-empty SELECT result and a verify block of SELECTs cannot stop a COMMIT.
-- ---------------------------------------------------------------------------
DO $guard$
DECLARE
    v_gate    text;
    v_optin   jsonb;
    v_others  bigint;
BEGIN
    SELECT s.value->'config'->>'query' INTO v_gate
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.type='build-pipeline-trigger' AND s.key='find_dispatchable_site'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;

    IF position('lock_except_item_ids' in COALESCE(v_gate,'')) = 0 THEN
        RAISE EXCEPTION '633 GUARD: the site gate does not name lock_except_item_ids after the update';
    END IF;
    IF position('s.locked_at IS NULL' in COALESCE(v_gate,'')) = 0 THEN
        RAISE EXCEPTION '633 GUARD: the site gate LOST its locked_at test — a locked site with no exceptions would now dispatch';
    END IF;

    SELECT s.value->'config'->'honour_site_lock' INTO v_optin
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.type='build-dispatch-loop' AND s.key='load_items'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
    IF v_optin IS DISTINCT FROM 'true'::jsonb THEN
        RAISE EXCEPTION '633 GUARD: build-dispatch-loop>load_items did not get honour_site_lock=true (got %)', v_optin;
    END IF;

    -- NEGATIVE CONTROL: site-work-orchestrator's two load_work_items steps are
    -- deliberately NOT opted in. If they are, this file did more than it says.
    SELECT count(*) INTO v_others
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.type='site-work-orchestrator'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
       AND s.value->'config' ? 'honour_site_lock';
    IF v_others <> 0 THEN
        RAISE EXCEPTION '633 GUARD: % site-work-orchestrator step(s) gained honour_site_lock — that flow is human-initiated and is deliberately out of scope', v_others;
    END IF;

    RAISE NOTICE '633 OK: site gate honours the exception list and KEPT its locked_at test; load_items opted in; site-work-orchestrator untouched.';
END;
$guard$;

COMMIT;
