-- SEED — evidence-researcher agent (V5 acquisition — SPEC_V5_researched_citations §3b)
--
-- APPLY ONLY AFTER an image carrying `verify_and_register_citations` is deployed
-- and pod-verified (CLAUDE.md: image first, then seeds):
--
--   kubectl -n ai-persona-system exec <agent-chassis-pod> -- \
--     sh -c 'strings /app/agent-chassis | grep -c verify_and_register_citations'
--
-- What it does: given a research question and a site, it searches the web,
-- scrapes the best sources, has an LLM extract ATOMIC candidate claims — each
-- with a number and a VERBATIM quote — then hands them to
-- `verify_and_register_citations`, which re-fetches every cited URL and only
-- registers candidates whose quote is actually present. The model proposes;
-- the string comparison disposes. Rejects go to human review, never into the
-- register.
--
-- Invocation (spawn/call via the generic entry point, or a dispatch handler):
--   input_data: { site_id, domain, research_query }
--   research_query example: "global LNG trade volume 2024 statistics"
--
-- Reuses the live research primitives (web_search / prepare_urls /
-- batch_webscrape — same config shapes as research-agent) and the V5 verifier.
-- ai_service is set at STEP level only (bugs_open/009: a root ai_service
-- shadows step config).

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status, is_active,
    image_repository, image_tag, input_contract, output_contract, default_config
)
SELECT
    'evidence-researcher',
    'Evidence Researcher',
    'V5 acquisition lane of the claims-verification layer: researches a question on the open web, extracts atomic candidate claims with verbatim quotes, and registers only those whose cited source demonstrably contains the quote (verify_and_register_citations). Unverifiable candidates terminate at human review. Facts it registers become writer-usable via V2 and re-verified via V4.',
    'analyst', 'analyst', 'active', true,
    'docker.io/aqls/agent-chassis',
    (SELECT image_tag FROM agent_definitions WHERE type = 'claims-auditor'),
    '{"required": ["site_id", "domain", "research_query"]}'::jsonb,
    '{"produces": {"registration": "verified citation facts added to the site evidence_base; rejects raised for human review"}}'::jsonb,
    $cfg${
  "workflow": {
    "start_step": "ensure_site_record",
    "processing_mode": "orchestrator",
    "timeout_seconds": 600,
    "steps": {
      "ensure_site_record": {
        "action": "ensure_site_record",
        "config": {"store_brief_in_content_data": false},
        "next_step": "search_web",
        "output_field": "site_record",
        "description": "Resolve site record"
      },
      "search_web": {
        "action": "web_search",
        "config": {"query_from": "input_data.research_query", "num_results": 10},
        "next_step": "prepare_urls",
        "output_field": "search_results",
        "description": "Search the open web for candidate sources"
      },
      "prepare_urls": {
        "action": "prepare_urls",
        "config": {
          "max_scrapes": 4,
          "max_snippets": 5,
          "prefer_domains": [".gov", ".org", ".edu", "iea.org", "igu.org", "reuters.com", "bbc.com", "ft.com", "spglobal.com"],
          "exclude_domains": ["pinterest.com", "facebook.com", "twitter.com", "instagram.com", "youtube.com", "reddit.com"]
        },
        "next_step": "scrape_pages",
        "output_field": "prepared_urls",
        "description": "Pick the strongest sources, preferring primary publishers over aggregators"
      },
      "scrape_pages": {
        "action": "batch_webscrape",
        "config": {
          "urls_field": "prepared_urls.urls_to_scrape",
          "scrape_config": {"only_main_content": true, "capture_screenshot": false}
        },
        "next_step": "extract_claims",
        "output_field": "scrape_results",
        "timeout_seconds": 120,
        "description": "Scrape the selected sources"
      },
      "extract_claims": {
        "action": "execute_llm_prompt",
        "config": {
          "ai_service": {"model": "claude-sonnet-4-6", "provider": "anthropic", "max_tokens": 8000, "api_key_env_var": "ANTHROPIC_API_KEY"},
          "input_fields": ["scrape_results", "prepared_urls", "input_data"],
          "output_format": "json",
          "prompt_template": "You are extracting ATOMIC, CITABLE claims for an evidence register. Research question: {{.input_data.research_query}}\n\nSCRAPED SOURCES (each has a url):\n{{.scrape_results}}\n\nExtract up to 10 candidate claims. Each MUST be atomic and verifiable:\n- a specific figure or a specific, dated factual statement — never a vague trend (\"grew strongly\" is NOT a claim);\n- a VERBATIM quote copied EXACTLY from ONE source's text that states it — do not paraphrase, do not stitch two sentences together, do not normalise numbers;\n- the url of that exact source page (from the material above — never invent or shorten one).\n\nEvery claim will be machine-checked: the url will be re-fetched and rejected unless your quote appears in it verbatim. A paraphrased quote fails. A summary fails. When a source page does not literally contain a sentence stating the fact, DO NOT include that claim.\n\nPrefer primary sources (the report itself) over articles quoting it; when only second-hand is available, set \"secondhand\": true.\n\nReturn ONLY a JSON array (no commentary). Each element:\n{\"claim\": \"plain statement of the fact\", \"value\": <number if the claim carries one>, \"unit\": \"...\", \"quote\": \"verbatim sentence from the source\", \"url\": \"...\", \"publisher\": \"...\", \"title\": \"...\", \"published\": \"YYYY-MM or YYYY if shown\", \"secondhand\": false, \"staleness_days\": <suggest: 200 for prices/market figures, 400 for annual statistics, 800 for structural facts>, \"writer_line\": \"how site copy should state it, with {value} where the number goes and the source named, e.g. 'global LNG trade reached {value} million tonnes in 2024 (IGU, World LNG Report 2025)'\"}\n\nIf nothing extractable meets the bar, return []."
        },
        "next_step": "verify_and_register",
        "output_field": "candidate_claims",
        "description": "Extract atomic claims with verbatim quotes — the checkable form"
      },
      "verify_and_register": {
        "action": "verify_and_register_citations",
        "config": {
          "site_id": "site_record.site_id",
          "candidates": "candidate_claims.result"
        },
        "next_step": "complete",
        "output_field": "registration",
        "description": "The disposal step: re-fetch every cited url; register only quotes that are really there; rejects to human review"
      },
      "complete": {
        "action": "complete_workflow",
        "config": {"output_fields": ["registration"]}
      }
    }
  }
}$cfg$::jsonb
WHERE NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type = 'evidence-researcher' AND deleted_at IS NULL)
RETURNING type, status;

-- ── Post-apply verification ────────────────────────────────────────────────
-- 1. Row exists:
--    SELECT type, status FROM agent_definitions WHERE type='evidence-researcher';
-- 2. Smoke run (spawn/call via generic entry, RUNBOOK §7 shape) with
--    input_data {site_id, domain, research_query} on a REAL question, then:
--    SELECT jsonb_pretty(collected_data->'registration') FROM orchestration_states WHERE ...;
--    Expect: registered ids for claims whose quotes verify; rejected list for
--    the rest; a citation_unverified work item iff rejected is non-empty.
-- 3. The registered facts appear in the site's evidence_base with
--    source.citation.{url,quote,accessed} and verified_at = today; the V4
--    freshness pass then re-verifies them daily.
--
-- ── Rollback ───────────────────────────────────────────────────────────────
--    UPDATE agent_definitions SET status='disabled', is_active=false
--    WHERE type='evidence-researcher';
-- Facts already registered are data, versioned in site_specs history; prune by
-- superseding the evidence_base row (never edit in place).
