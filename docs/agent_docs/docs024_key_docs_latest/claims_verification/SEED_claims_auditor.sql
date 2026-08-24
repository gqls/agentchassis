-- SEED - claims-auditor agent (V3 judgement lane of the claims-verification layer)
--
-- THE FIRST SEED THIS AGENT HAS EVER HAD. It was hand-built as a live row on 2026-07-17
-- and existed nowhere in git (migration 350 records that; register CLM-006 carries the
-- provenance warning). This file was REGENERATED FROM THE LIVE ROW immediately after
-- migrations 597 and 601 applied (2026-08-24), which is the row's authoritative state:
--   597: cold-register audit (bugs_open/380) - opt-in gate DELETED, DB errors FAIL the
--        run, page cap 12000, per-run doc_notes receipts, recurrence declared.
--   601: page text stripped PER COMPONENT in a deterministic order (Postgres greedy-first
--        regex was eating cross-component text; the owner's own sentence never reached
--        the model).
-- Any later migration that touches this agent MUST regenerate this file in the same
-- commit (the 464 seed-correction rule). The live row is the source of truth; this seed
-- exists so a fresh database can be stood up and so reviewers can read the workflow.
--
-- ai_service is set at STEP level only (bugs_open/009: a root ai_service shadows step
-- config). image_tag is copied from a live sibling at insert time.
--
-- Idempotent: WHERE NOT EXISTS on an active claims-auditor row.

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status, is_active,
    image_repository, image_tag, default_config
)
SELECT
    'claims-auditor',
    'Claims Auditor',
    'V3 of the claims-verification layer (judgement lane): one LLM pass per site per audit that extracts prose assertions of fact and classifies them supported / could-framed / unsupported against the site''s evidence_base register. With NO register (or no facts) it runs COLD (bugs_open/380, migration 597): every assertion of fact about the business is unsupported, and first-person present-tense practice claims are the class it reports at severity high. Findings terminate at human review (claims_unverified, item_key claims_llm*) - never rewrites. Every run writes one doc_notes receipt (pipeline/claims-audit). A query failure FAILS the run (RFC_017).',
    'analyst', 'analyst', 'active', true,
    'docker.io/aqls/agent-chassis',
    (SELECT image_tag FROM agent_definitions
      WHERE type='page-content-writer' AND is_active
        AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL),
    $cfg${
  "workflow": {
    "steps": {
      "complete": {
        "action": "complete_workflow",
        "config": {
          "output_fields": [
            "claims_audit"
          ]
        }
      },
      "check_findings": {
        "action": "conditional_branch",
        "config": {
          "condition": "claims_audit.result",
          "else_step": "compose_receipt_clean",
          "then_step": "request_claims_review"
        },
        "description": "No findings, no item — clean passes leave no noise"
      },
      "load_page_text": {
        "action": "query_database",
        "config": {
          "query": "SELECT p.name, LEFT(regexp_replace(string_agg(regexp_replace(regexp_replace(regexp_replace(pc.rendered_html, '<style[^>]*?>.*?</style>', ' ', 'gi'), '<script[^>]*?>.*?</script>', ' ', 'gi'), '<[^>]*>', ' ', 'g'), ' ' ORDER BY pc.position, pc.slot_name), '\\s+', ' ', 'g'), 12000) AS page_text FROM pages p JOIN page_components pc ON pc.page_id = p.id WHERE p.site_id = $1 AND p.build_status IN ('deployed','active') AND pc.rendered_html IS NOT NULL AND pc.rendered_html <> '' AND pc.locked_at IS NULL GROUP BY p.id, p.name ORDER BY p.name",
          "params": [
            "site_record.site_id"
          ],
          "output_format": "rows"
        },
        "next_step": "run_claims_audit",
        "description": "Tag-stripped text of all deployed unlocked pages, capped per page",
        "output_field": "page_texts"
      },
      "run_claims_audit": {
        "action": "execute_llm_prompt",
        "config": {
          "prompt": "You are a claims auditor. Your job is factual verification against an evidence register — not writing quality.\n\nEvery ASSERTION OF FACT about this business in the page text below must be supported by an entry in the verified fact register. An assertion of fact claims something exists, happened, or is true of THIS business: numbers and statistics, track record (\"we built/verified/served…\"), named client or partner relationships, capabilities presented as deployed today.\n\nNOT assertions (ignore these): offers honestly framed as possibilities (\"we could build…\"), generic industry statements not about this business, navigation and UI text, interactive tool interface copy.\n\nVerdicts:\n- supported: a register fact covers the SPECIFIC claim wording. A fact supports its wording, not its topic — \"records verified against Companies House\" is supported; \"handles dissolved companies\" would not be, even though the register has a Companies House fact.\n- could-framed: honestly presented as an offer or possibility. Fine — do not report.\n- unsupported: nothing in the register covers it. Includes: invented clients, staff, or awards; invented track record; capability claims beyond the register; business numbers with no matching fact; a TRUE registered number under a FALSE label (a verified-records count presented as \"Awards Won\").\n\nNamed entities: the copy may assert relationships only with entities in the ALLOWED ENTITIES list or named inside a register fact. Any other named client, partner, award body, or publication is unsupported.\n\nReport ONLY unsupported assertions — at most 12, worst first. Respond with ONLY a JSON array (no commentary):\n[{\"page\":\"…\",\"assertion\":\"exact quoted text from the page\",\"why_unsupported\":\"one sentence\",\"nearest_fact_id\":\"closest register id or null\",\"severity\":\"high|medium|low\",\"suggestion\":\"the honest fix: cite the right fact, reframe as could-do, or delete\"}]\nseverity: high = invented entities, people, or track record; medium = capability or number overstatement; low = wording drift from a real fact.\nIf every assertion is supported or could-framed, respond with exactly: []\n\n## VERIFIED FACT REGISTER\n{{if .evidence_facts.facts_text}}{{.evidence_facts.facts_text}}{{else}}(EMPTY — no verified facts are registered for this business, and no operating history is attested. Every assertion of fact about this business is therefore unsupported. Report, at severity high, every first-person present-tense or habitual statement of PRACTICE — that we test, trial, buy, purchase, weigh, measure, record, inspect, receive or are sent products or samples, visit, interview, survey, garden or cook ourselves — and every description of a review or assessment method stated as something this business DOES, in body copy and in FAQ answers alike. Statements that manufacturers, brands or suppliers send this business anything are the same class. Claims of scale, coverage, clients, team, premises or track record are severity medium. Soft self-evidencing verbs (we look at, we consider, we compare) are severity low. Do NOT report: descriptions of what the page contains or explains; conditional or aspirational framings (could, aim to, plan to, will); negated statements and honest disclosures (\"we have not tested\", \"where we have not used a tool directly, we say so\"); quoted third-party speech; generic industry statements not about this business.){{end}}\n\nALLOWED ENTITIES: {{if .evidence_facts.allowed_entities}}{{.evidence_facts.allowed_entities}}{{else}}(none){{end}}\n\n## PAGE TEXT\n{{.page_texts}}",
          "ai_service": {
            "model": "claude-sonnet-4-6",
            "provider": "anthropic",
            "max_tokens": 8000,
            "api_key_env_var": "ANTHROPIC_API_KEY"
          },
          "input_fields": [
            "evidence_facts",
            "page_texts",
            "site_record"
          ]
        },
        "next_step": "check_findings",
        "description": "One LLM pass: extract assertions, classify vs the register, report unsupported only",
        "output_field": "claims_audit"
      },
      "ensure_site_record": {
        "action": "ensure_site_record",
        "config": {
          "store_brief_in_content_data": false
        },
        "next_step": "load_evidence_facts",
        "description": "Resolve site record",
        "output_field": "site_record"
      },
      "load_evidence_facts": {
        "action": "query_database",
        "config": {
          "query": "SELECT (SELECT string_agg('- [' || (f.value->>'id') || '] ' || (f.value->>'claim') || CASE WHEN f.value ? 'value' THEN ' [value: ' || (f.value->>'value') || COALESCE(', tolerance: ' || (f.value->>'tolerance'), '') || ']' ELSE '' END, E'\\n') FROM jsonb_array_elements(ss.data->'facts') f) AS facts_text, (SELECT string_agg(e.value #>> '{}', ', ') FROM jsonb_array_elements(ss.data->'allowed_entities') e) AS allowed_entities FROM site_specs ss WHERE ss.site_id = $1 AND ss.aspect = 'evidence_base' AND ss.is_current = true",
          "params": [
            "site_record.site_id"
          ],
          "output_format": "object"
        },
        "next_step": "load_page_text",
        "description": "Load the evidence base fact register (empty when site not opted in)",
        "output_field": "evidence_facts"
      },
      "write_receipt_clean": {
        "action": "append_doc_note",
        "config": {
          "created_by": "claims-auditor",
          "subject_key": "claims-audit",
          "subject_type": "pipeline",
          "note_body_field": "receipt.doc_note_body",
          "note_categories": [
            "claims-audit",
            "audit-ran"
          ],
          "note_site_id_field": "site_record.site_id"
        },
        "next_step": "complete",
        "description": "Durable per-run receipt (doc_notes). Findings are work items; this is coverage."
      },
      "compose_receipt_clean": {
        "action": "query_database",
        "config": {
          "query": "SELECT 'claims audit ran clean for ' || s.domain || ' (register facts: ' || COALESCE((SELECT jsonb_array_length(COALESCE(ss.data->'facts','[]'::jsonb))::text FROM site_specs ss WHERE ss.site_id = s.id AND ss.aspect = 'evidence_base' AND ss.is_current), 'none') || ')' AS doc_note_body FROM sites s WHERE s.id = $1",
          "params": [
            "site_record.site_id"
          ],
          "output_format": "object"
        },
        "next_step": "write_receipt_clean",
        "description": "Compose the per-run receipt body (clean pass). One receipt per run, clean or not (RFC_024): a MISSING doc_notes row means the audit did not run.",
        "output_field": "receipt"
      },
      "request_claims_review": {
        "action": "create_work_item",
        "config": {
          "status": "needs_human_review",
          "site_id": "site_record.site_id",
          "summary": "Claims audit: unsupported prose assertions need a human ruling",
          "severity": "high",
          "item_type": "claims_unverified",
          "spec_data": "claims_audit",
          "handler_agent": "human-review",
          "item_pipeline": "content",
          "item_key_prefix": "claims_llm",
          "recurrence_expected": false
        },
        "next_step": "compose_receipt_findings",
        "description": "HITL-terminal: findings become one needs_human_review item per pass; no auto-fix, ever",
        "output_field": "review_item"
      },
      "write_receipt_findings": {
        "action": "append_doc_note",
        "config": {
          "created_by": "claims-auditor",
          "subject_key": "claims-audit",
          "subject_type": "pipeline",
          "note_body_field": "receipt.doc_note_body",
          "note_categories": [
            "claims-audit",
            "audit-findings"
          ],
          "note_site_id_field": "site_record.site_id"
        },
        "next_step": "complete",
        "description": "Durable per-run receipt (doc_notes), findings variant."
      },
      "compose_receipt_findings": {
        "action": "query_database",
        "config": {
          "query": "SELECT 'claims audit for ' || s.domain || ' filed unsupported-assertion findings for human review (work item claims_llm)' AS doc_note_body FROM sites s WHERE s.id = $1",
          "params": [
            "site_record.site_id"
          ],
          "output_format": "object"
        },
        "next_step": "write_receipt_findings",
        "description": "Compose the per-run receipt body (findings filed).",
        "output_field": "receipt"
      }
    },
    "start_step": "ensure_site_record",
    "processing_mode": "orchestrator",
    "timeout_seconds": 300
  }
}$cfg$::jsonb
WHERE NOT EXISTS (
  SELECT 1 FROM agent_definitions
   WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);
