-- 0NN_fix_proposer_v7_code_lookup.sql — F2.3b(c): the council's CODE-shaped
-- verify tier. PATCH seed: targeted jsonb updates against the LIVE fix-proposer
-- def (v6+), NOT a full re-seed. Idempotent (every UPDATE is guarded).
--
-- v7 (2026-07-17): reviewers gain `code_checks` beside their SQL `checks` —
-- structured code questions (kind: symbol | content | ls) answered from the
-- code_symbols index by the new diagnose_code_lookup action (fixed SQL,
-- reviewer input only as bind parameters; runs in-chassis, no repo token).
-- Born from run ca064df2 (2026-07-17): the bug-historian's materially-correct
-- objection ("do other LLM provider adapters exist?") was un-resolvable by the
-- SQL-only verify step; three revise rounds burned; the router escalated. The
-- answer was one index query away: code_symbols has BOTH GenerateText
-- implementations (anthropic.go + ollama.go).
--
-- ██ DEPLOY SEQUENCING ██ — apply ONLY AFTER the chassis image carrying the
-- diagnose_code_lookup action is live (committed 2026-07-17, first image AFTER
-- v1.0.1128 — verify: strings /app/agent-chassis | grep -c diagnose_code_lookup
-- in the RUNNING POD). A v7 workflow on an older binary fails at the unknown
-- action; the v6 workflow on the new binary is harmless. Order: image → this
-- file → fire.
--
-- Wiring: run_checks (SQL tier) → code_lookup (code tier) → repropose, which
-- renders both result blocks. All three reviewer prompts learn the code_checks
-- schema. council_decide is untouched (it reads verdicts only).

BEGIN;

-- 1. New step: code_lookup ---------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,code_lookup}',
      $step$
      {
        "action": "diagnose_code_lookup",
        "config": {
          "code_check_fields": [
            "review_editquality.result.code_checks",
            "review_bug_historian.result.code_checks",
            "review_reuse_agent.result.code_checks",
            "review_guidelines.result.code_checks",
            "review_tooling_provenance.result.code_checks",
            "review_guardian.result.code_checks"
          ],
          "max_checks": 8,
          "row_cap": 40,
          "excerpt_chars": 400
        },
        "next_step": "repropose",
        "description": "F2.3b(c): answer the reviewers' code-shaped questions (code_checks) from the code_symbols index; results feed the repropose beside the SQL check_results.",
        "output_field": "code_lookup_results"
      }
      $step$::jsonb
    ),
    updated_at = now()
WHERE type = 'fix-proposer'
  AND default_config #> '{workflow,steps,code_lookup}' IS NULL;

-- 2. Route the revise path THROUGH it: run_checks → code_lookup (was repropose)
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,run_checks,next_step}',
      '"code_lookup"'
    ),
    updated_at = now()
WHERE type = 'fix-proposer'
  AND default_config #>> '{workflow,steps,run_checks,next_step}' = 'repropose';

-- 3. repropose consumes the new results field --------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,repropose,config,input_fields}',
      (default_config #> '{workflow,steps,repropose,config,input_fields}')
        || '["code_lookup_results"]'::jsonb
    ),
    updated_at = now()
WHERE type = 'fix-proposer'
  AND NOT (default_config #> '{workflow,steps,repropose,config,input_fields}')
        @> '["code_lookup_results"]'::jsonb;

-- 4. repropose prompt: render the code answers beside the SQL answers --------
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,repropose,config,prompt_template}',
      to_jsonb(
        replace(
          default_config #>> '{workflow,steps,repropose,config,prompt_template}',
          $a$## Verification results (the reviewers' own read-only queries, now answered)
{{.check_results.results_text}}$a$,
          $b$## Verification results (the reviewers' own read-only queries, now answered)
{{.check_results.results_text}}

## Code lookup results (the reviewers' code questions, answered from the code_symbols index)
{{.code_lookup_results.results_text}}

These settle CODE-shaped objections (other implementations, symbol locations,
call sites). If a lookup shows the mechanism exists in MORE places than the
plan covers, WIDEN the plan to cover them or name the residual explicitly —
do not resubmit a plan a lookup has just shown to be partial.$b$
        )
      )
    ),
    updated_at = now()
WHERE type = 'fix-proposer'
  AND (default_config #>> '{workflow,steps,repropose,config,prompt_template}')
        NOT LIKE '%code_lookup_results%';

-- 5+6. Every reviewer seat learns the code_checks schema AND gets the guidance
-- paragraph — looped over ALL seats so the set survives roster growth (v8 added
-- reuse_agent, guidelines, tooling_provenance to the original three; a future
-- seat is picked up by re-running this idempotent block). Per-seat guarded:
-- the schema replace only fires where the anchor is still present and
-- code_checks absent; the guidance append only where it is absent.
DO $do$
DECLARE
  seat text;
  seats text[] := ARRAY[
    'review_editquality','review_bug_historian','review_reuse_agent',
    'review_guidelines','review_tooling_provenance','review_guardian'
  ];
  tmpl text;
  path text[];
  schema_anchor text := '"checks": [{"sql": "SELECT ...", "why": "what this settles"}]';
  schema_repl  text := '"checks": [{"sql": "SELECT ...", "why": "what this settles"}], "code_checks": [{"kind": "symbol|content|ls", "query": "pattern", "why": "what this settles"}]';
  guidance text :=
E'\n\nCODE QUESTIONS (code_checks): when your verdict hinges on a fact about the\n'
'CODEBASE — does another implementation of this mechanism exist? which files\n'
'carry symbol X? does anything reference Y? — attach a code_checks entry rather\n'
'than objecting blind. kind "symbol" matches symbol names, "content" searches\n'
'source bodies, "ls" lists indexed paths under a prefix; all are answered from\n'
'the code_symbols index next round. SQL checks cannot see the codebase;\n'
'code_checks cannot see the database — use each for its own tier.';
BEGIN
  FOREACH seat IN ARRAY seats LOOP
    path := ARRAY['workflow','steps',seat,'config','prompt_template'];
    -- only act on seats that actually exist in this def
    IF (SELECT default_config #> path FROM agent_definitions WHERE type='fix-proposer') IS NULL THEN
      CONTINUE;
    END IF;
    SELECT default_config #>> path INTO tmpl FROM agent_definitions WHERE type='fix-proposer';
    IF tmpl NOT LIKE '%code_checks%' AND position(schema_anchor in tmpl) > 0 THEN
      tmpl := replace(tmpl, schema_anchor, schema_repl);
    END IF;
    IF tmpl NOT LIKE '%CODE QUESTIONS (code_checks)%' THEN
      tmpl := tmpl || guidance;
    END IF;
    UPDATE agent_definitions
      SET default_config = jsonb_set(default_config, path, to_jsonb(tmpl)),
          updated_at = now()
    WHERE type='fix-proposer';
  END LOOP;
END
$do$;

COMMIT;

-- Verify (run after applying):
--   SELECT default_config #>> '{workflow,steps,run_checks,next_step}',       -- code_lookup
--          default_config #> '{workflow,steps,code_lookup}' IS NOT NULL,     -- t
--          (default_config #> '{workflow,steps,repropose,config,input_fields}')
--            @> '["code_lookup_results"]'::jsonb                             -- t
--   FROM agent_definitions WHERE type='fix-proposer';
-- Re-grade: re-run 091 on e505f70f (clear its council_reports first for a fair
-- round count) — expect the historian's adapter question to arrive as a
-- code_check, the answer (anthropic.go + ollama.go) to reach repropose, and
-- the plan to widen instead of exhausting.
