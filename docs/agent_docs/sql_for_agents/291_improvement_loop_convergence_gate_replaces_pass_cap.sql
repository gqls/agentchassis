-- 291 — improvement-loop: convergence gate replaces the 3-pass cap (bugs_open/171)
--
-- vigilant_designer_offer_analysis programme, Phase A0.2 (PLAN 2026-08-02).
--
-- WHAT WAS WRONG (171, both halves):
--   1. A site at its lifetime 3-audit cap short-circuited to notify_scheduler_clean and
--      reported "No issues found — site is clean". A capped site is not clean; it is
--      unexamined.
--   2. BOTH skip paths bypassed triage_findings, so a capped site's `detected` pile was
--      never promoted by anyone — the cap silently disabled the drain along with the
--      audit, though promotion costs one UPDATE and the cap only ever protected LLM spend.
--
-- THE REPLACEMENT. A site is fingerprinted (rendered page components + composed palette
-- + chrome). The audit chain runs when the fingerprint changed OR a 14-day cooldown has
-- expired; otherwise the LLM/discovery chain is skipped — but PROMOTION NOW RUNS ON
-- EVERY PATH (skip paths jump straight to triage_findings; the single-owner rule of
-- migration 286 is preserved because both paths converge on the ONE existing step).
-- Three audits at an UNCHANGED fingerprint means fixes are not landing: the loop stops
-- spending there and files ONE deferred capability_gap roadmap row instead of reporting
-- clean. `audit_state` (fingerprint, audit_due, not_converging) lands in collected_data
-- and in both terminal outputs, so a skipped audit is separable from a clean one in
-- orchestration_states — 171's other requirement.
--
-- FINGERPRINT NOTE. pages/page_components.content_hash was the obvious input and it is
-- a DEAD COLUMN — populated on 0 of 1,183 live rows (measured 2026-08-02, the
-- usage_count lesson again). The function hashes md5(rendered_html) instead.
--
-- jsonb_set NOTE. jsonb_set only creates the LAST path element; writing
-- {maintenance_profile,last_audit} into settings lacking maintenance_profile would
-- silently no-op. record_audit_pass therefore materialises the parent first.
--
-- WHAT THIS DOES NOT TOUCH: get_audit_pass_count/increment_audit_pass stay defined
-- (history; sole caller was this workflow); improvement-sweep stays enabled=false
-- (G1 is the owner's separate call); the audit chain's internal error edges are
-- unchanged — note call_completeness_discovery.error_step still jumps to
-- triage_findings, so a completeness error skips the LLM auditors AND skips
-- record_audit_pass, meaning an errored chain retries next sweep. Deliberate: an
-- errored audit should not consume a cooldown slot.

SELECT snapshot_agent('improvement-loop', '291_improvement_loop_convergence_gate_replaces_pass_cap.sql: pre-update');

BEGIN;

-- ── STEP 1 — the fingerprint function ─────────────────────────────────────
CREATE OR REPLACE FUNCTION site_audit_fingerprint(p_site_id uuid) RETURNS text AS $fn$
    SELECT md5(
        COALESCE((SELECT string_agg(md5(COALESCE(pc.rendered_html, '')), '|'
                                    ORDER BY pc.page_id, pc.position, pc.id)
                    FROM page_components pc
                    JOIN pages p ON p.id = pc.page_id
                   WHERE p.site_id = p_site_id
                     AND pc.build_status <> 'removed'), '')
        || '§' ||
        COALESCE((SELECT pal.colours::text
                    FROM site_specs ss
                    JOIN palettes pal ON pal.id = (ss.data->>'palette_id')::uuid
                   WHERE ss.site_id = p_site_id
                     AND ss.aspect = 'resolved_composition'
                     AND ss.is_current), '')
        || '§' ||
        COALESCE((SELECT string_agg(md5(COALESCE(sc.rendered_html, '')), '|'
                                    ORDER BY sc.slot_name)
                    FROM site_components sc
                   WHERE sc.site_id = p_site_id), '')
    );
$fn$ LANGUAGE sql STABLE;

-- ── STEP 2 — remove the cap steps ─────────────────────────────────────────
UPDATE agent_definitions
SET default_config = default_config
        #- '{workflow,steps,load_pass_count}'
        #- '{workflow,steps,check_audit_pass_limit}'
        #- '{workflow,steps,increment_audit_pass}',
    updated_at = now()
WHERE type = 'improvement-loop'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── STEP 3 — add the five convergence-gate steps ──────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps}',
        (default_config #> '{workflow,steps}') || $steps$
{
  "load_audit_state": {
    "action": "query_database",
    "config": {
      "query": "SELECT fp AS fingerprint, not_conv AS not_converging, CASE WHEN not_conv THEN false ELSE (fp_changed OR cooldown_expired) END AS audit_due FROM (SELECT fp, (NOT fp_changed AND passes >= 3) AS not_conv, fp_changed, cooldown_expired FROM (SELECT site_audit_fingerprint($1) AS fp, COALESCE(s.settings#>>'{maintenance_profile,last_audit,fingerprint}','') <> site_audit_fingerprint($1) AS fp_changed, ((s.settings#>>'{maintenance_profile,last_audit,at}') IS NULL OR (s.settings#>>'{maintenance_profile,last_audit,at}')::timestamptz < now() - interval '14 days') AS cooldown_expired, COALESCE((s.settings#>>'{maintenance_profile,last_audit,passes_at_fingerprint}')::int, 0) AS passes FROM sites s WHERE s.id = $1) a) b",
      "params": ["site_record.site_id"],
      "output_format": "object"
    },
    "next_step": "check_audit_due",
    "description": "Fingerprint the site and decide: audit due (changed or 14d cooldown), skip (unchanged, in cooldown), or not-converging (3 audits at one fingerprint)",
    "output_field": "audit_state"
  },
  "check_audit_due": {
    "action": "conditional",
    "config": {
      "condition": "audit_state.audit_due == true",
      "then_step": "spawn_quality_discovery",
      "else_step": "check_not_converging"
    },
    "description": "Audit only when the site changed or cooldown expired; promotion runs regardless (171)"
  },
  "check_not_converging": {
    "action": "conditional",
    "config": {
      "condition": "audit_state.not_converging == true",
      "then_step": "record_not_converging",
      "else_step": "triage_findings"
    },
    "description": "Three audits at one fingerprint: fixes are not landing — surface it instead of spending again"
  },
  "record_not_converging": {
    "action": "create_work_item",
    "config": {
      "spec": {
        "reason": "three audit passes at an unchanged site fingerprint with findings still open — fixes are not landing; human attention needed",
        "capability": "audit_not_converging"
      },
      "source": "improvement-loop",
      "site_id": "site_record.site_id",
      "summary": "Audit not converging: 3 passes at one fingerprint, findings still open",
      "priority": 200,
      "severity": "low",
      "status": "deferred",
      "item_type": "capability_gap",
      "item_pipeline": "build",
      "handler_agent": "",
      "item_key_prefix": "capability_gap_audit_not_converging",
      "recurrence_expected": true
    },
    "next_step": "triage_findings",
    "error_step": "triage_findings",
    "description": "One deferred roadmap row per site (dedup on item_key); never dispatched — the remit.go capability_gap convention",
    "output_field": "not_converging_recorded"
  },
  "record_audit_pass": {
    "action": "query_database",
    "config": {
      "query": "UPDATE sites SET settings = jsonb_set(jsonb_set(COALESCE(settings, '{}'::jsonb), '{maintenance_profile}', COALESCE(settings->'maintenance_profile', '{}'::jsonb), true), '{maintenance_profile,last_audit}', jsonb_build_object('fingerprint', $2::text, 'at', now(), 'passes_at_fingerprint', CASE WHEN settings#>>'{maintenance_profile,last_audit,fingerprint}' = $2::text THEN COALESCE((settings#>>'{maintenance_profile,last_audit,passes_at_fingerprint}')::int, 0) + 1 ELSE 1 END), true) WHERE id = $1",
      "params": ["site_record.site_id", "audit_state.fingerprint"],
      "output_format": "object"
    },
    "next_step": "triage_findings",
    "description": "Record the audited fingerprint; count consecutive audits at an unchanged fingerprint",
    "output_field": "audit_pass_recorded"
  }
}
$steps$::jsonb),
    updated_at = now()
WHERE type = 'improvement-loop'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── STEP 4 — repoint the surviving edges ──────────────────────────────────
-- enrich_news_feed carries its error edge INSIDE config (a fifth edge shape the
-- 288 checker did not cover — both are repointed and both are guarded below).
UPDATE agent_definitions
SET default_config =
      jsonb_set(
        jsonb_set(
          jsonb_set(
            jsonb_set(
              jsonb_set(default_config,
                '{workflow,steps,enrich_news_feed,next_step}', '"load_audit_state"'),
              '{workflow,steps,enrich_news_feed,config,error_step}', '"load_audit_state"'),
            '{workflow,steps,triage_findings,next_step}', '"check_has_findings"'),
          '{workflow,steps,call_site_review,next_step}', '"record_audit_pass"'),
        '{workflow,steps,call_site_review,error_step}', '"record_audit_pass"'),
    updated_at = now()
WHERE type = 'improvement-loop'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── STEP 5 — honest terminal outputs (skipped ≠ clean) ────────────────────
UPDATE agent_definitions
SET default_config =
      jsonb_set(
        jsonb_set(
          jsonb_set(default_config,
            '{workflow,steps,complete_clean,config,output_fields}',
            '["quality_result","design_result","completeness_result","triage_result","audit_state"]'),
          '{workflow,steps,complete_clean,config,success_message}',
          '"No dispatchable findings after promotion — audit_state says whether the audit ran, was skipped in cooldown, or is not converging"'),
        '{workflow,steps,complete,config,output_fields}',
        '["completeness_result","design_audit_result","design_result","dispatch_result","news_feed_enrichment","quality_result","site_review_result","triage_result","audit_state"]'),
    updated_at = now()
WHERE type = 'improvement-loop'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── STEP 6 — ENFORCING GUARD ──────────────────────────────────────────────
DO $$
DECLARE
    steps jsonb;
    dangling text;
    triage_owners int;
    probe_site uuid;
    fp1 text;
    fp2 text;
    r record;
BEGIN
    SELECT default_config #> '{workflow,steps}' INTO steps
    FROM agent_definitions
    WHERE type = 'improvement-loop'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF steps IS NULL THEN
        RAISE EXCEPTION '291: live improvement-loop row not found';
    END IF;

    -- (i) the cap steps are gone, the gate steps exist
    IF steps ? 'load_pass_count' OR steps ? 'check_audit_pass_limit' OR steps ? 'increment_audit_pass' THEN
        RAISE EXCEPTION '291: a cap step survived: %', (SELECT string_agg(k, ', ') FROM jsonb_object_keys(steps) k WHERE k IN ('load_pass_count','check_audit_pass_limit','increment_audit_pass'));
    END IF;
    IF NOT (steps ? 'load_audit_state' AND steps ? 'check_audit_due' AND steps ? 'check_not_converging'
            AND steps ? 'record_not_converging' AND steps ? 'record_audit_pass') THEN
        RAISE EXCEPTION '291: a gate step is missing';
    END IF;

    -- (ii) no dangling edges, across ALL FIVE edge shapes incl. config.error_step
    SELECT string_agg(t.step_name || ' -> ' || t.target, ', ') INTO dangling
    FROM (
        SELECT step.key AS step_name, tgt.target AS target
        FROM jsonb_each(steps) AS step,
             LATERAL (VALUES (step.value->>'next_step'),
                             (step.value->>'error_step'),
                             (step.value->'config'->>'error_step'),
                             (step.value->'config'->>'then_step'),
                             (step.value->'config'->>'else_step')) AS tgt(target)
        WHERE tgt.target IS NOT NULL AND tgt.target <> ''
          AND NOT steps ? tgt.target
    ) t;
    IF dangling IS NOT NULL THEN
        RAISE EXCEPTION '291: dangling step reference(s): %', dangling;
    END IF;

    -- (iii) single-owner: exactly one triage_detected_items step fleet-wide (286)
    SELECT count(*) INTO triage_owners
    FROM agent_definitions ad,
         jsonb_each(ad.default_config->'workflow'->'steps') AS step
    WHERE ad.is_active AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL
      AND step.value->>'action' = 'triage_detected_items';
    IF triage_owners <> 1 THEN
        RAISE EXCEPTION '291: triage_detected_items owner count is % (must be exactly 1 — migration 286)', triage_owners;
    END IF;

    -- (iv) the fingerprint function is live, stable, and the gate query returns
    --      non-null booleans (the NULL-boolean trap is mechanical here, not a review comment)
    SELECT id INTO probe_site FROM sites WHERE status IN ('deployed','active') ORDER BY created_at LIMIT 1;
    IF probe_site IS NOT NULL THEN
        fp1 := site_audit_fingerprint(probe_site);
        fp2 := site_audit_fingerprint(probe_site);
        IF fp1 IS NULL OR length(fp1) <> 32 OR fp1 <> fp2 THEN
            RAISE EXCEPTION '291: site_audit_fingerprint unstable or malformed for %: % / %', probe_site, fp1, fp2;
        END IF;
        SELECT fp AS fingerprint, not_conv AS not_converging,
               CASE WHEN not_conv THEN false ELSE (fp_changed OR cooldown_expired) END AS audit_due
        INTO r
        FROM (SELECT fp, (NOT fp_changed AND passes >= 3) AS not_conv, fp_changed, cooldown_expired
              FROM (SELECT site_audit_fingerprint(probe_site) AS fp,
                           COALESCE(s.settings#>>'{maintenance_profile,last_audit,fingerprint}','') <> site_audit_fingerprint(probe_site) AS fp_changed,
                           ((s.settings#>>'{maintenance_profile,last_audit,at}') IS NULL
                             OR (s.settings#>>'{maintenance_profile,last_audit,at}')::timestamptz < now() - interval '14 days') AS cooldown_expired,
                           COALESCE((s.settings#>>'{maintenance_profile,last_audit,passes_at_fingerprint}')::int, 0) AS passes
                    FROM sites s WHERE s.id = probe_site) a) b;
        IF r.audit_due IS NULL OR r.not_converging IS NULL THEN
            RAISE EXCEPTION '291: gate query produced NULL booleans for %', probe_site;
        END IF;
        RAISE NOTICE '291: probe site % — fingerprint %, audit_due %, not_converging %', probe_site, fp1, r.audit_due, r.not_converging;
    ELSE
        RAISE NOTICE '291: no deployed/active site to probe — function shape asserted only';
    END IF;

    RAISE NOTICE '291: convergence gate live; promotion now runs on every path';
END $$;

COMMIT;

-- ── ROLLBACK ──
-- snapshot_agent() wrote a before-image row; restore the workflow from it
-- (no restore_* helper function exists — verified against pg_proc 2026-08-02):
--   UPDATE agent_definitions live
--   SET default_config = snap.default_config, updated_at = now()
--   FROM (SELECT default_config FROM agent_definitions
--          WHERE type='improvement-loop' AND COALESCE(is_snapshot,false)
--            AND notes LIKE '291_%' ORDER BY created_at DESC LIMIT 1) snap
--   WHERE live.type='improvement-loop' AND live.is_active
--     AND COALESCE(live.is_snapshot,false)=false AND live.deleted_at IS NULL;
--   (check the snapshot row's actual marker column — notes vs description — before running)
-- DROP FUNCTION IF EXISTS site_audit_fingerprint(uuid);   -- optional; harmless if left
