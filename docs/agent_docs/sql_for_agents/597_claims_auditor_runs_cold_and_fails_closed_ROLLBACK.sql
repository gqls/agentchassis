-- ROLLBACK for 597_claims_auditor_runs_cold_and_fails_closed.sql
--
-- Anchored reverse edits (the 464 rollback shape): restores the opt-in gate, the silent
-- error route, the single-arm prompt and the 3500 cap, and removes the receipt steps.
-- Aborts rather than guess if any anchor is not found exactly once — so it cannot damage
-- an edit another migration has made since. Restores the DEFECT bugs_open/380 describes;
-- roll back only to recover from a 597 misfire, not to keep.

BEGIN;

SELECT snapshot_agent('claims-auditor', '597_ROLLBACK: pre-revert');

DO $do$
DECLARE
  cfg jsonb; steps jsonb; prompt text; q text; n int;
  repl1 text := $R$## VERIFIED FACT REGISTER
{{if .evidence_facts.facts_text}}{{.evidence_facts.facts_text}}{{else}}(EMPTY — no verified facts are registered for this business, and no operating history is attested. Every assertion of fact about this business is therefore unsupported. Report, at severity high, every first-person present-tense or habitual statement of PRACTICE — that we test, trial, buy, purchase, weigh, measure, record, inspect, receive or are sent products or samples, visit, interview, survey, garden or cook ourselves — and every description of a review or assessment method stated as something this business DOES, in body copy and in FAQ answers alike. Statements that manufacturers, brands or suppliers send this business anything are the same class. Claims of scale, coverage, clients, team, premises or track record are severity medium. Soft self-evidencing verbs (we look at, we consider, we compare) are severity low. Do NOT report: descriptions of what the page contains or explains; conditional or aspirational framings (could, aim to, plan to, will); negated statements and honest disclosures ("we have not tested", "where we have not used a tool directly, we say so"); quoted third-party speech; generic industry statements not about this business.){{end}}$R$;
  anchor1 text := $A$## VERIFIED FACT REGISTER
{{.evidence_facts.facts_text}}$A$;
  repl2 text := 'ALLOWED ENTITIES: {{if .evidence_facts.allowed_entities}}{{.evidence_facts.allowed_entities}}{{else}}(none){{end}}';
  anchor2 text := 'ALLOWED ENTITIES: {{.evidence_facts.allowed_entities}}';
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
   FOR UPDATE;

  IF (cfg #> '{workflow,steps}') ? 'check_opted_in' THEN
    RAISE EXCEPTION '597 ROLLBACK: check_opted_in already present — 597 not applied';
  END IF;

  -- Prompt back to one arm.
  prompt := cfg #>> '{workflow,steps,run_claims_audit,config,prompt}';
  n := (length(prompt) - length(replace(prompt, repl1, ''))) / length(repl1);
  IF n <> 1 THEN RAISE EXCEPTION '597 ROLLBACK: cold arm found % times, expected 1', n; END IF;
  n := (length(prompt) - length(replace(prompt, repl2, ''))) / length(repl2);
  IF n <> 1 THEN RAISE EXCEPTION '597 ROLLBACK: entities guard found % times, expected 1', n; END IF;
  prompt := replace(replace(prompt, repl1, anchor1), repl2, anchor2);

  -- Cap back to 3500.
  q := cfg #>> '{workflow,steps,load_page_text,config,query}';
  n := (length(q) - length(replace(q, ', 12000) AS page_text', ''))) / length(', 12000) AS page_text');
  IF n <> 1 THEN RAISE EXCEPTION '597 ROLLBACK: cap anchor found % times, expected 1', n; END IF;
  q := replace(q, ', 12000) AS page_text', ', 3500) AS page_text');

  -- Structure back: receipt steps out, gate back in, pointers restored.
  steps := (cfg #> '{workflow,steps}')
             - 'compose_receipt_clean' - 'compose_receipt_findings'
             - 'write_receipt_clean' - 'write_receipt_findings';
  steps := steps || $J$
{
  "check_opted_in": {
    "action": "conditional_branch",
    "config": {
      "condition": "evidence_facts.facts_text",
      "then_step": "load_page_text",
      "else_step": "complete"
    },
    "description": "Skip entirely when the site has no evidence base"
  }
}
$J$::jsonb;
  steps := jsonb_set(steps, '{load_evidence_facts,next_step}', '"check_opted_in"');
  steps := jsonb_set(steps, '{load_evidence_facts,config,error_step}', '"complete"');
  steps := jsonb_set(steps, '{check_findings,config,else_step}', '"complete"');
  steps := jsonb_set(steps, '{request_claims_review,next_step}', '"complete"');
  steps := steps #- '{request_claims_review,config,recurrence_expected}';
  steps := jsonb_set(steps, '{run_claims_audit,config,prompt}', to_jsonb(prompt));
  steps := jsonb_set(steps, '{load_page_text,config,query}', to_jsonb(q));

  UPDATE agent_definitions
     SET default_config = jsonb_set(cfg, '{workflow,steps}', steps),
         updated_at = now()
   WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '597 ROLLBACK: updated % rows', n; END IF;

  RAISE NOTICE '597 ROLLBACK OK: gate restored, receipts removed, prompt and cap reverted.';
END $do$;

COMMIT;
