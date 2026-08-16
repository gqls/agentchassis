-- SQL_2026-08-16 — two doc_notes rows for RFC_029 Phase 1 (staged_component_build lane)
--
-- WHY: the council round on the Phase 1 submission (corr 75091072-…, run ae2a88a7,
-- REVISE) had two seats say the same thing from different angles — tooling_provenance:
-- "no doc_notes entry records that Phase 1 shipped, what the WARN instrument is, or when
-- the window opens"; prior_art_librarian: "every claim traces to RFC_029 §9, which no
-- seat can see" (owner rulings live in markdown; a known gate landmine). Both are met by
-- writing the record INTO the DB, keyed so a seat's ILIKE '%RFC_029%' and a subject
-- lookup both find it. Follows the precedent of the PAY-009 lane's evidence note
-- (decision / council-submission-4ac1fe52, 2026-08-11).
--
-- Data rows only (no agent config, no DDL) — hand-applied, not a numbered migration.
-- Idempotent: each INSERT is fenced on its own subject_key + a body marker.
-- Apply:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
--     -v ON_ERROR_STOP=1 -f - < docs/agent_docs/docs024_key_docs_latest/staged_component_build/SQL_2026-08-16_doc_notes_rfc029_phase1_and_council_evidence.sql

BEGIN;

-- Row A: the ruling and what shipped, readable in-DB.
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
SELECT 'decision', 'RFC_029',
$note$RFC_029 — the aggressive recursive (whole-tree) search may resolve, never guess. OWNER-DELEGATED RULING 2026-08-15 (§9), Phase 1 IMPLEMENTED + REVISED (§10) — written to doc_notes 2026-08-16 so council seats can read it (owner rulings in markdown are invisible to seats).

WHAT THE MECHANISM IS: when an action input field has no explicit mapping, datahelpers' resolver (platform/orchestration/datahelpers/unified_extractor.go findFieldRecursive, reached from extractSingleField / ExtractActionInputs) walks the run's whole collected_data for any key with that name. Until 2026-08-15 it took the FIRST match a randomised Go map iteration met — a per-run coin flip when two values share the name (two production incidents: bugs 248 and 213).

THE RULING (§9, verbatim key lines): "no field at all is better than a wrong field." Chosen: UNIQUE-OR-NOTHING — "the search collects ALL matches (same depth cap as today). If every match carries the same value, it resolves — deterministically, shallowest path first … If the candidates CONFLICT, it resolves NOTHING and logs a WARN naming the field and every candidate path."
D1 findFieldRecursive → collect-all / unique-or-nothing, sorted-key traversal.
D2 "Instrument first, refuse second (two chassis builds)." Phase 1: conflicts STILL RESOLVE (stable shallowest winner) + WARN. Phase 2 flips conflicts to refusal. "Precondition for the flip: zero conflict WARNs observed [over 48h minimum, a week preferred], or every observed field/caller pair given an explicit mapping first."
D3 opt-in `!` strict marker (mirror of `?`): explicit resolution only, loud failure, never the search. Default OFF. First adopter image-build-handler asset_id (migration 417_HOLD). Second named adopter build-dispatch-loop (402) NOT adopted — DATED CORRECTION §10.3: that mapping is shared by 636+ item types; `!` there hard-fails every non-asset dispatch.
D4 arm budget on the inner chain (extractSingleField: floor 5, ceiling 8, resolver_arm_budget_test.go), descriptive arm names (direct-path, input-data-prefix, input-data-map, whole-tree-search, alias); migration 402's "Strategy 4" miscitation carries a dated correction.

PHASE 1 SHIPPED: commits 927e12bd9 (test repairs), 1806371ef (implementation), 6b0736eed (notes). ROLLED: chassis v1.0.1303, binary stamped 5e075a6f9 (descendant of 1806371ef), probed 2026-08-16.

COUNCIL ROUND 1 (corr 75091072-9d65-433e-8a30-84719dc3f30f, run ae2a88a7): REVISE — reuse_agent (HIGH, gating): the two WARNs were plain log lines and pod log retention is ~90s, so the 48h+ window could not be read after the fact. Correct. REVISION (2026-08-16, same lane): every occurrence of both WARNs is ALSO persisted to agent_error_log via a registered sink (datahelpers.SetResolverFindingRecorder, nil = log-only, registered by the chassis in agentbase.initializeComponents where the pool + pod identity live; thin wrapper over orchestration.LogAgentError → agenterrors.Write). ERROR CODES: RESOLVER_CONFLICTING_CANDIDATES (context: field, candidate_paths, winner_path, phase) and RESOLVER_MAPPING_BYPASSED (context: field, reference, resolved_type); severity 'warning'; action 'input-resolver'; pod-level attribution only (orchestration_id empty BY DESIGN — the resolver cannot know it; each row's context says so under identity_scope).

THE OBSERVATION WINDOW OPENS AT THE ROLL OF THE REVISED BUILD (not Phase 1's) and is read from rows:
  SELECT error_code, context->>'field', count(*), min(occurred_at), max(occurred_at) FROM agent_error_log WHERE error_code IN ('RESOLVER_CONFLICTING_CANDIDATES','RESOLVER_MAPPING_BYPASSED') GROUP BY 1,2 ORDER BY 3 DESC;
Phase 2 (conflicts resolve NOTHING) is its own council-gated task and proceeds only on D2's precondition. Disconfirmation (§9): "a substantial population of conflict WARNs whose lucky winner is load-bearing" → Phase 2 does not proceed on schedule; pairs get explicit mappings first; §9 gets a dated correction.

Full text: docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_029_the_aggressive_recursive_search_has_no_boundary_for_an_unmapped_field.md (§9 ruling, §10 implementation + §10.3 correction + §10.4 revision). Concept register CTS-060. Lane: staged_component_build.$note$,
'["council-evidence","rfc","rfc-029","observation-window","resolver"]'::jsonb,
'staged_component_build lane, session 2026-08-16', 'RFC_029 Phase 1 revision prep'
WHERE NOT EXISTS (
  SELECT 1 FROM doc_notes WHERE subject_type='decision' AND subject_key='RFC_029'
    AND body LIKE 'RFC_029 — the aggressive recursive (whole-tree) search may resolve, never guess.%'
);

-- Row B: this round's evidence, with the queries that produced it (PAY-009 precedent).
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
SELECT 'decision', 'council-submission-75091072',
$note$EVIDENCE NOTE for council correlation 75091072-9d65-433e-8a30-84719dc3f30f (RFC_029 Phase 1 → revision, round 2) — dated 2026-08-16, written by the submitting session so seats can INSPECT the checks rather than take them on report. Each item states the check output and the query/command that produced it.

[1] Round-1 verdict read: REVISE, decided by a gating HIGH objection from reuse_agent (run ae2a88a7, completed 2026-08-15 14:10Z). SELECT metadata->>'decision', body FROM diagnosis_artifacts WHERE correlation_id='75091072-9d65-433e-8a30-84719dc3f30f' AND kind='council_report' ORDER BY created_at DESC LIMIT 1;

[2] The objection is correct and is FIXED IN CODE by this round: both Phase 1 WARNs (unified_extractor.go "aggressive search: conflicting candidates"; action_inputs.go "aggressive search: explicit single-segment mapping bypassed") now also call recordResolverFinding(...) → the chassis-registered recorder → orchestration.LogAgentError → agenterrors.Write (the ONE writer, RFC_012). Rows: error_code RESOLVER_CONFLICTING_CANDIDATES / RESOLVER_MAPPING_BYPASSED, severity 'warning', action 'input-resolver'. Nil recorder = log-only = previous behaviour (default-OFF).

[3] Migration ledger checked (tooling_provenance): SELECT filename, applied_at FROM schema_migrations WHERE filename LIKE '41%' → 417_brief_fidelity_auditor_speaks_the_routers_vocabulary.sql applied 2026-08-15 17:46Z (another lane); 417_image_build_handler_asset_id_goes_strict_HOLD.sql NOT present (unclaimed, held on purpose). The ledger keys on the full filename, so the shared number does not collide.

[4] Two-active-rows trap (debug_historian): SELECT count(*), array_agg(version) FROM agent_definitions WHERE type='image-build-handler' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL → 1 row, version {1}. NOT APPLICABLE.

[5] snapshot_agent overload (debug_historian): SELECT p.oid::regprocedure, (regexp_matches(pg_get_functiondef(p.oid),'INSERT INTO\s+(\w+)','g'))[1] FROM pg_proc p WHERE proname='snapshot_agent' → snapshot_agent(text) writes agent_definitions; snapshot_agent(text,text) writes agent_definitions_backup. Migration 417 calls the two-arg form → pre-image lands in agent_definitions_backup; header now says so and gives the has-old-key check.

[6] Binary provenance (debug_historian's pod-verification ask): agent-chassis deploy image docker.io/aqls/agent-chassis:v1.0.1303; kubectl exec <pod> -- grep -aq <sha> /proc/1/exe → 5e075a6f9 PRESENT, bc4cd65e7 (HEAD, must be absent) absent; git merge-base --is-ancestor 1806371ef 5e075a6f9 → true. Phase 1 log-only is LIVE; the `!` parser is live; the revision's rows are NOT live until a chassis image ≥ its commit rolls. Recipe for next time: logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance' (startup line, scrolls) then the /proc/1/exe probe with a two-way control.

[7] Live sample (tiny, stated as such): kubectl logs -l app=agent-chassis --tail=2000 (2 pods) → resolver INFO lines present ("Trying aggressive search" 1, "Found via aggressive search" 1), WARNs of either kind 0. Consistent with §9's premise; not evidence of it — which is why the rows exist.

[8] editquality's two "missing D4" items: SHIPPED in 1806371ef — git show 1806371ef --stat lists unified_extractor.go (arm renames: direct-path, input-data-prefix, input-data-map, whole-tree-search, alias) and docs/agent_docs/sql_for_agents/402_build_dispatch_loop_maps_asset_id_top_level.sql (dated CORRECTION block). The round-1 plan failed to LIST them; round 2 names them.

[9] guardian's "winner changes now for every consumer": that is RFC_029 §9 D2's explicit owner-delegated choice — Phase 1 = determinism + instrument, conflicts still resolve (stable shallowest-first instead of a coin flip); refusal is Phase 2. Text in doc_notes decision/RFC_029 (row written today).

[10] Tests: go test ./platform/orchestration/datahelpers/ ./platform/agentbase/ ./platform/orchestration/agenterrors/ → ok ×3 from git archive HEAD + this task's files; both recordResolverFinding call sites mutation-proven (removing either fails its test); TestResolverArmBudget / TestInnerChainArmBudget unchanged (10/15, 5/8).$note$,
'["council-evidence","rfc-029","migration"]'::jsonb,
'staged_component_build lane, session 2026-08-16', 'council submission 75091072 round 2 prep'
WHERE NOT EXISTS (
  SELECT 1 FROM doc_notes WHERE subject_type='decision' AND subject_key='council-submission-75091072'
    AND body LIKE 'EVIDENCE NOTE for council correlation 75091072-%round 2)%'
);

DO $$
DECLARE a int; b int;
BEGIN
  SELECT count(*) INTO a FROM doc_notes WHERE subject_type='decision' AND subject_key='RFC_029';
  SELECT count(*) INTO b FROM doc_notes WHERE subject_type='decision' AND subject_key='council-submission-75091072';
  IF a < 1 OR b < 1 THEN
    RAISE EXCEPTION 'SQL_2026-08-16 doc_notes: expected both rows, got RFC_029=% council-submission-75091072=%', a, b;
  END IF;
  RAISE NOTICE 'doc_notes rows present: RFC_029=% council-submission-75091072=%', a, b;
END $$;

COMMIT;
