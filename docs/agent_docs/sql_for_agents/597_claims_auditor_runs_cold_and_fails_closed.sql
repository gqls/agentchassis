-- 597_claims_auditor_runs_cold_and_fails_closed.sql
--
-- bugs_open/380 slice S2. The claims-auditor's opt-in gate inverted the safety property it
-- guarded: `check_opted_in` branched on `evidence_facts.facts_text` — a string_agg over
-- facts[], which is NULL for a site with NO register AND for a site with an EMPTY one — and
-- its else arm was `complete`, so the sites with nothing attested (the ones whose writer had
-- to invent to fill a page) got no audit at all, recorded as success. 33 of 48 sites
-- [MEASURED 2026-08-24]. The auditor itself is an orphan: no seed file (migration 350 records
-- that), no schedule, ONE llm_call_log row ever (2026-07-18, returned []), and its
-- request_claims_review step has never fired (item_key LIKE 'claims_llm%' = 0 rows).
-- ⚠ THOSE TWO FIGURES ARE PINNED TO BEFORE THIS FIX, because proving the fix populated the
-- tables cited as empty (caught by the loanzy_uk_example_site lane within the hour):
--   SELECT count(*) FROM llm_call_log WHERE agent_type='claims-auditor'
--     AND created_at < '2026-08-24 16:00+00';                              -- 1
--   SELECT count(*) FROM site_work_items WHERE item_key LIKE 'claims_llm%'
--     AND created_at < '2026-08-24 16:00+00';                              -- 0
-- Re-run them UNPINNED and you measure the fix, not the bug.
--
-- The cold-audit posture this enables was designed 2026-07-20 and never built
-- (claims_verification/PLAN_2026-07-20_gaswholesalers_second_site.md §3): treat every
-- operational assertion as unsupported until someone attests it. Owner rulings applied:
-- RFC_017 (audits fail CLOSED — a DB error now FAILS the run rather than completing it),
-- RFC_023 (making a failure visible is a fix, ordinary gate), owner 2026-08-24 on
-- bugs_open/380 ("aspirations stated as present-tense practice… say only what is sourced").
-- Owner decision 2026-08-24 (this lane's plan D1): NO empty register is minted anywhere —
-- absence IS the cold posture, and this migration is what makes that true for the auditor.
--
-- WHAT CHANGES, all on the live claims-auditor row (its config exists nowhere in git):
--   1. check_opted_in is DELETED (not repointed — a skip branch that exists will be pointed
--      somewhere again); load_evidence_facts.next_step -> load_page_text.
--   2. load_evidence_facts.config.error_step ('complete') is REMOVED: a query failure now
--      fails the workflow (status FAILED + agent_error_log via the standard router), which
--      the immune-system sweep can see. Completing on error was the same silent skip.
--   3. The audit prompt's register section becomes a two-arm branch: the roster as today, or
--      a cold-register instruction that names the practice-claims class (the garden-tools /
--      gaswholesalers defect) with an explicit do-not-report list (could-framed, negations,
--      quoted speech, industry statements) so honesty is not punished. ALLOWED ENTITIES is
--      nil-guarded (it would render "<no value>" on a register-less site).
--   4. load_page_text's per-page cap 3500 -> 12000 chars: garden-tools' how-we-assess is
--      ~9k chars of text; at 3500 the audit reads the top third and a clean verdict on the
--      rest is "not looking", not "nothing there".
--   5. Every run now writes ONE doc_notes receipt row (subject 'pipeline'/'claims-audit',
--      categories audit-ran | audit-findings) — the RFC_022/RFC_024 convention: a MISSING
--      row means the audit did not run, so coverage becomes a query instead of an inference.
--      Findings still file the claims_unverified work item exactly as before (HITL-terminal;
--      the two-strike revalidator contract on that step is unchanged — migration 572's
--      reasoning — and recurrence_expected: false is now stated explicitly).
--
-- The bare {{.evidence_facts.facts_text}} render appears TWICE conceptually (roster header +
-- ALLOWED ENTITIES line is a different variable) — each anchor below is asserted at its own
-- expected count before any write. Template balance ({{if}}/{{end}}) is asserted unchanged+2/+2.
--
-- Apply: psql -f, then record. DB config: LIVE IMMEDIATELY on the next dispatch.
-- Companion: 597_claims_auditor_runs_cold_and_fails_closed_ROLLBACK.sql (snapshot restore).
-- Seed regenerated in the same commit: claims_verification/SEED_claims_auditor.sql.

BEGIN;

-- ---------------------------------------------------------------------------
-- Guard 1: exactly one active, non-snapshot claims-auditor row.
-- ---------------------------------------------------------------------------
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '597 REFUSED: expected exactly 1 active claims-auditor row, found %', n;
  END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Guard 2: pre-state. The marker that drives everything is check_opted_in's
-- presence; re-running after apply raises 'already applied'. Drift in any of
-- the pointers this migration rewrites aborts before the snapshot.
-- ---------------------------------------------------------------------------
DO $$
DECLARE cfg jsonb; prompt text; q text; n int;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF NOT (cfg #> '{workflow,steps}') ? 'check_opted_in' THEN
    RAISE EXCEPTION '597: already applied (check_opted_in absent)';
  END IF;
  IF cfg #>> '{workflow,steps,load_evidence_facts,next_step}' IS DISTINCT FROM 'check_opted_in' THEN
    RAISE EXCEPTION '597: drift — load_evidence_facts.next_step is %',
      cfg #>> '{workflow,steps,load_evidence_facts,next_step}';
  END IF;
  IF cfg #>> '{workflow,steps,check_opted_in,config,else_step}' IS DISTINCT FROM 'complete' THEN
    RAISE EXCEPTION '597: drift — check_opted_in.config.else_step is %',
      cfg #>> '{workflow,steps,check_opted_in,config,else_step}';
  END IF;
  IF cfg #>> '{workflow,steps,load_evidence_facts,config,error_step}' IS DISTINCT FROM 'complete' THEN
    RAISE EXCEPTION '597: drift — load_evidence_facts.config.error_step is %',
      cfg #>> '{workflow,steps,load_evidence_facts,config,error_step}';
  END IF;
  IF cfg #>> '{workflow,steps,check_findings,config,else_step}' IS DISTINCT FROM 'complete' THEN
    RAISE EXCEPTION '597: drift — check_findings.config.else_step is %',
      cfg #>> '{workflow,steps,check_findings,config,else_step}';
  END IF;
  IF cfg #>> '{workflow,steps,request_claims_review,next_step}' IS DISTINCT FROM 'complete' THEN
    RAISE EXCEPTION '597: drift — request_claims_review.next_step is %',
      cfg #>> '{workflow,steps,request_claims_review,next_step}';
  END IF;
  IF (cfg #> '{workflow,steps}') ? 'compose_receipt_clean' THEN
    RAISE EXCEPTION '597: drift — compose_receipt_clean already exists';
  END IF;

  prompt := cfg #>> '{workflow,steps,run_claims_audit,config,prompt}';
  IF prompt IS NULL THEN RAISE EXCEPTION '597: run_claims_audit.config.prompt not found'; END IF;

  n := (length(prompt) - length(replace(prompt, $A$## VERIFIED FACT REGISTER
{{.evidence_facts.facts_text}}$A$, ''))) / length($A$## VERIFIED FACT REGISTER
{{.evidence_facts.facts_text}}$A$);
  IF n <> 1 THEN RAISE EXCEPTION '597: register anchor found % times, expected 1', n; END IF;

  n := (length(prompt) - length(replace(prompt, 'ALLOWED ENTITIES: {{.evidence_facts.allowed_entities}}', '')))
       / length('ALLOWED ENTITIES: {{.evidence_facts.allowed_entities}}');
  IF n <> 1 THEN RAISE EXCEPTION '597: allowed-entities anchor found % times, expected 1', n; END IF;

  q := cfg #>> '{workflow,steps,load_page_text,config,query}';
  IF q IS NULL THEN RAISE EXCEPTION '597: load_page_text.config.query not found'; END IF;
  n := (length(q) - length(replace(q, ', 3500) AS page_text', ''))) / length(', 3500) AS page_text');
  IF n <> 1 THEN RAISE EXCEPTION '597: page-cap anchor found % times, expected 1', n; END IF;

  PERFORM snapshot_agent('claims-auditor',
                         '597_claims_auditor_runs_cold_and_fails_closed.sql: pre-update');
END $$;

-- ---------------------------------------------------------------------------
-- Edit 1: structure. Delete the gate, repoint the chain through the receipt
-- steps, remove the silent error route, declare recurrence, add 4 new steps.
-- Built SEQUENTIALLY on one steps object — a parallel jsonb_set over the
-- original config would clobber its own sibling edits.
-- ---------------------------------------------------------------------------
DO $do1$
DECLARE cfg jsonb; steps jsonb; n int;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
   FOR UPDATE;

  steps := (cfg #> '{workflow,steps}') - 'check_opted_in';
  steps := steps || $J$
{
  "compose_receipt_clean": {
    "action": "query_database",
    "config": {
      "query": "SELECT 'claims audit ran clean for ' || s.domain || ' (register facts: ' || COALESCE((SELECT jsonb_array_length(COALESCE(ss.data->'facts','[]'::jsonb))::text FROM site_specs ss WHERE ss.site_id = s.id AND ss.aspect = 'evidence_base' AND ss.is_current), 'none') || ')' AS doc_note_body FROM sites s WHERE s.id = $1",
      "params": ["site_record.site_id"],
      "output_format": "object"
    },
    "next_step": "write_receipt_clean",
    "output_field": "receipt",
    "description": "Compose the per-run receipt body (clean pass). One receipt per run, clean or not (RFC_024): a MISSING doc_notes row means the audit did not run."
  },
  "compose_receipt_findings": {
    "action": "query_database",
    "config": {
      "query": "SELECT 'claims audit for ' || s.domain || ' filed unsupported-assertion findings for human review (work item claims_llm)' AS doc_note_body FROM sites s WHERE s.id = $1",
      "params": ["site_record.site_id"],
      "output_format": "object"
    },
    "next_step": "write_receipt_findings",
    "output_field": "receipt",
    "description": "Compose the per-run receipt body (findings filed)."
  },
  "write_receipt_clean": {
    "action": "append_doc_note",
    "config": {
      "subject_type": "pipeline",
      "subject_key": "claims-audit",
      "note_body_field": "receipt.doc_note_body",
      "note_categories": ["claims-audit", "audit-ran"],
      "note_site_id_field": "site_record.site_id",
      "created_by": "claims-auditor"
    },
    "next_step": "complete",
    "description": "Durable per-run receipt (doc_notes). Findings are work items; this is coverage."
  },
  "write_receipt_findings": {
    "action": "append_doc_note",
    "config": {
      "subject_type": "pipeline",
      "subject_key": "claims-audit",
      "note_body_field": "receipt.doc_note_body",
      "note_categories": ["claims-audit", "audit-findings"],
      "note_site_id_field": "site_record.site_id",
      "created_by": "claims-auditor"
    },
    "next_step": "complete",
    "description": "Durable per-run receipt (doc_notes), findings variant."
  }
}
$J$::jsonb;

  steps := jsonb_set(steps, '{load_evidence_facts,next_step}', '"load_page_text"');
  steps := steps #- '{load_evidence_facts,config,error_step}'
                 #- '{load_evidence_facts,error_step}';
  steps := jsonb_set(steps, '{check_findings,config,else_step}', '"compose_receipt_clean"');
  steps := jsonb_set(steps, '{request_claims_review,next_step}', '"compose_receipt_findings"');
  steps := jsonb_set(steps, '{request_claims_review,config,recurrence_expected}', 'false');

  UPDATE agent_definitions
     SET default_config = jsonb_set(cfg, '{workflow,steps}', steps),
         updated_at = now()
   WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND (default_config #> '{workflow,steps}') ? 'check_opted_in';
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '597 edit1: updated % rows, expected exactly 1', n; END IF;
END $do1$;

-- ---------------------------------------------------------------------------
-- Edit 2: the prompt's two-arm register section + nil-guarded entities line.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  prompt text; newprompt text; n int;
  anchor1 text := $A$## VERIFIED FACT REGISTER
{{.evidence_facts.facts_text}}$A$;
  repl1 text := $R$## VERIFIED FACT REGISTER
{{if .evidence_facts.facts_text}}{{.evidence_facts.facts_text}}{{else}}(EMPTY — no verified facts are registered for this business, and no operating history is attested. Every assertion of fact about this business is therefore unsupported. Report, at severity high, every first-person present-tense or habitual statement of PRACTICE — that we test, trial, buy, purchase, weigh, measure, record, inspect, receive or are sent products or samples, visit, interview, survey, garden or cook ourselves — and every description of a review or assessment method stated as something this business DOES, in body copy and in FAQ answers alike. Statements that manufacturers, brands or suppliers send this business anything are the same class. Claims of scale, coverage, clients, team, premises or track record are severity medium. Soft self-evidencing verbs (we look at, we consider, we compare) are severity low. Do NOT report: descriptions of what the page contains or explains; conditional or aspirational framings (could, aim to, plan to, will); negated statements and honest disclosures ("we have not tested", "where we have not used a tool directly, we say so"); quoted third-party speech; generic industry statements not about this business.){{end}}$R$;
  anchor2 text := 'ALLOWED ENTITIES: {{.evidence_facts.allowed_entities}}';
  repl2 text := 'ALLOWED ENTITIES: {{if .evidence_facts.allowed_entities}}{{.evidence_facts.allowed_entities}}{{else}}(none){{end}}';
  ifs_before int; ends_before int; ifs_after int; ends_after int;
BEGIN
  SELECT default_config #>> '{workflow,steps,run_claims_audit,config,prompt}' INTO prompt
    FROM agent_definitions WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  ifs_before  := (length(prompt) - length(replace(prompt, '{{if ', ''))) / length('{{if ');
  ends_before := (length(prompt) - length(replace(prompt, '{{end}}', ''))) / length('{{end}}');

  n := (length(prompt) - length(replace(prompt, anchor1, ''))) / length(anchor1);
  IF n <> 1 THEN RAISE EXCEPTION '597 edit2: anchor1 found % times', n; END IF;
  n := (length(prompt) - length(replace(prompt, anchor2, ''))) / length(anchor2);
  IF n <> 1 THEN RAISE EXCEPTION '597 edit2: anchor2 found % times', n; END IF;

  newprompt := replace(prompt, anchor1, repl1);
  newprompt := replace(newprompt, anchor2, repl2);

  IF length(newprompt) <> length(prompt) + (length(repl1) - length(anchor1)) + (length(repl2) - length(anchor2)) THEN
    RAISE EXCEPTION '597 edit2: unexpected length delta (% vs %)',
      length(newprompt) - length(prompt),
      (length(repl1) - length(anchor1)) + (length(repl2) - length(anchor2));
  END IF;

  -- Template balance: exactly two new {{if}} and two new {{end}}.
  ifs_after  := (length(newprompt) - length(replace(newprompt, '{{if ', ''))) / length('{{if ');
  ends_after := (length(newprompt) - length(replace(newprompt, '{{end}}', ''))) / length('{{end}}');
  IF ifs_after <> ifs_before + 2 OR ends_after <> ends_before + 2 THEN
    RAISE EXCEPTION '597 edit2: template balance broken (ifs % -> %, ends % -> %)',
      ifs_before, ifs_after, ends_before, ends_after;
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,run_claims_audit,config,prompt}', to_jsonb(newprompt), false),
         updated_at = now()
   WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '597 edit2: updated % rows', n; END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Edit 3: page-text cap 3500 -> 12000 (garden-tools' how-we-assess is ~9k chars).
-- ---------------------------------------------------------------------------
DO $$
DECLARE q text; n int;
BEGIN
  SELECT default_config #>> '{workflow,steps,load_page_text,config,query}' INTO q
    FROM agent_definitions WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  n := (length(q) - length(replace(q, ', 3500) AS page_text', ''))) / length(', 3500) AS page_text');
  IF n <> 1 THEN RAISE EXCEPTION '597 edit3: cap anchor found % times', n; END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,load_page_text,config,query}',
           to_jsonb(replace(q, ', 3500) AS page_text', ', 12000) AS page_text')), false),
         updated_at = now()
   WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '597 edit3: updated % rows', n; END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Verify: DO/RAISE (a verify block of bare SELECTs verifies nothing).
-- ---------------------------------------------------------------------------
DO $$
DECLARE cfg jsonb; prompt text; n int;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF (cfg #> '{workflow,steps}') ? 'check_opted_in' THEN
    RAISE EXCEPTION '597 VERIFY: check_opted_in survived';
  END IF;
  IF (cfg #> '{workflow,steps,load_evidence_facts}') ? 'error_step'
     OR (cfg #> '{workflow,steps,load_evidence_facts,config}') ? 'error_step' THEN
    RAISE EXCEPTION '597 VERIFY: load_evidence_facts still carries an error_step';
  END IF;
  IF cfg #>> '{workflow,steps,load_evidence_facts,next_step}' IS DISTINCT FROM 'load_page_text' THEN
    RAISE EXCEPTION '597 VERIFY: load_evidence_facts.next_step is %',
      cfg #>> '{workflow,steps,load_evidence_facts,next_step}';
  END IF;
  IF cfg #>> '{workflow,steps,request_claims_review,config,recurrence_expected}' IS DISTINCT FROM 'false' THEN
    RAISE EXCEPTION '597 VERIFY: recurrence_expected not declared false';
  END IF;

  -- Every pointer on THIS agent must name an existing step (a typo''d pointer is ""
  -- -> silent completion, which is the defect class this migration exists to end).
  SELECT count(*) INTO n
  FROM jsonb_each(cfg->'workflow'->'steps') AS step(k,v),
       LATERAL (VALUES (v->>'next_step'), (v->>'error_step'),
                       (v->'config'->>'then_step'), (v->'config'->>'else_step')) AS tgt(target)
  WHERE tgt.target IS NOT NULL AND tgt.target <> ''
    AND NOT (cfg #> '{workflow,steps}') ? tgt.target;
  IF n <> 0 THEN RAISE EXCEPTION '597 VERIFY: % dangling pointers on claims-auditor', n; END IF;

  prompt := cfg #>> '{workflow,steps,run_claims_audit,config,prompt}';
  IF position('no verified facts are registered for this business' in prompt) = 0 THEN
    RAISE EXCEPTION '597 VERIFY: cold arm missing from prompt';
  END IF;
  IF position('{{if .evidence_facts.facts_text}}' in prompt) = 0 THEN
    RAISE EXCEPTION '597 VERIFY: register branch missing';
  END IF;
  IF position(', 12000) AS page_text' in cfg #>> '{workflow,steps,load_page_text,config,query}') = 0 THEN
    RAISE EXCEPTION '597 VERIFY: page cap not raised';
  END IF;
  IF NOT ((cfg #> '{workflow,steps}') ? 'compose_receipt_clean'
      AND (cfg #> '{workflow,steps}') ? 'compose_receipt_findings'
      AND (cfg #> '{workflow,steps}') ? 'write_receipt_clean'
      AND (cfg #> '{workflow,steps}') ? 'write_receipt_findings') THEN
    RAISE EXCEPTION '597 VERIFY: receipt steps missing';
  END IF;

  RAISE NOTICE '597 OK: gate deleted, error route removed, cold arm live, cap 12000, receipts wired.';
END $$;

COMMIT;
