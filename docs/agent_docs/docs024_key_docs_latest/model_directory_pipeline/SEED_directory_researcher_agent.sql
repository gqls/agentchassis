-- SEED — directory-researcher agent (model directory pipeline, Phase B)
--
-- APPLY ONLY AFTER an image carrying `verify_and_register_directory_claims`
-- is deployed and pod-verified (CLAUDE.md: image first, then seeds):
--
--   kubectl -n ai-persona-system exec <agent-chassis-pod> -- \
--     sh -c 'strings /app/agent-chassis | grep -c verify_and_register_directory_claims'
--
-- What it does: given a research question, it searches the web, scrapes the
-- best sources, has an LLM extract ATOMIC candidate claims about AI models —
-- each with a number/fact and a VERBATIM quote — then hands them to
-- `verify_and_register_directory_claims`, which re-fetches every cited URL
-- and only registers candidates whose quote is actually present. The model
-- proposes; the string comparison disposes. Rejects go to human review,
-- never into the register. Directly modelled on evidence-researcher
-- (claims_verification/SEED_evidence_researcher.sql) — this agent needs no
-- site-resolution step (ensure_site_record) because the registry it writes
-- to is global, not per-site.
--
-- Invocation (spawn/call via the generic entry point, or a scheduled_tasks
-- inline workflow — see SEED_directory_scheduled_tasks.sql):
--   input_data: { research_query }
--   research_query examples:
--     "AI models released in the last 30 days pricing and context window"
--     "Anthropic Claude Sonnet 4.6 pricing per million tokens"
--
-- Reuses the live research primitives (web_search / prepare_urls /
-- batch_webscrape — same config shapes as research-agent / evidence-researcher)
-- and the V5 verifier (verifyCitationLive, reused unchanged by
-- directory_claims.go). ai_service is set at STEP level only (bugs_open/009:
-- a root ai_service shadows step config).

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status, is_active,
    image_repository, image_tag, input_contract, output_contract, default_config
)
SELECT
    'directory-researcher',
    'Directory Researcher',
    'Model directory pipeline acquisition lane: researches AI models on the open web, extracts atomic candidate claims (price, context window, licence, ...) with verbatim quotes, and registers only those whose cited source demonstrably contains the quote (verify_and_register_directory_claims) into the cross-site directory_entities/directory_claims registry. Unverifiable candidates terminate at human review.',
    'analyst', 'analyst', 'active', true,
    'docker.io/aqls/agent-chassis',
    (SELECT image_tag FROM agent_definitions WHERE type = 'claims-auditor'),
    '{"required": ["research_query"]}'::jsonb,
    '{"produces": {"registration": "verified directory claims added to directory_entities/directory_claims; rejects raised for human review"}}'::jsonb,
    $cfg${
  "workflow": {
    "start_step": "search_web",
    "processing_mode": "orchestrator",
    "timeout_seconds": 600,
    "steps": {
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
          "prefer_domains": ["openai.com", "anthropic.com", "ai.google.dev", "deepmind.google", "mistral.ai", "meta.ai", "huggingface.co", "x.ai", "cohere.com", "openrouter.ai"],
          "exclude_domains": ["pinterest.com", "facebook.com", "twitter.com", "instagram.com", "reddit.com"]
        },
        "next_step": "scrape_pages",
        "output_field": "prepared_urls",
        "description": "Pick the strongest sources, preferring the model owner's own docs/pricing pages over aggregators"
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
          "prompt_template": "You are extracting ATOMIC, CITABLE claims about AI MODELS for a public model directory. Research question: {{.input_data.research_query}}\n\nSCRAPED SOURCES (each has a url):\n{{.scrape_results}}\n\nExtract up to 15 candidate claims, one per (model, fact) pair. Each MUST be atomic and verifiable:\n- name the specific model (entity_slug: lowercase-hyphenated, e.g. 'anthropic-claude-sonnet-4-6'; entity_name: display name; entity_owner: the company that runs/owns it);\n- field: one of price_input_per_mtok, price_output_per_mtok, context_window, license, modality, release_date (or another short snake_case field name if the source states a different concrete, checkable spec);\n- a VERBATIM quote copied EXACTLY from ONE source's text that states it — do not paraphrase, do not stitch two sentences together, do not normalise numbers or units;\n- the url of that exact source page (from the material above — never invent or shorten one).\n\nEvery claim will be machine-checked: the url will be re-fetched and rejected unless your quote appears in it verbatim. A paraphrased quote fails. A summary fails. When a source page does not literally contain a sentence stating the fact, DO NOT include that claim.\n\nPrefer the model owner's own pricing/docs pages over third-party aggregators or reviews.\n\nOptionally, on the FIRST claim you emit for a given model, also include entity_summary (one plain sentence: what the model is / does), entity_links (object: docs, weights [only if genuinely open-weight], video_urls [array, only if you found a genuine demo/review video URL in the material]), and entity_attributes (object: modality as an array e.g. [\"text\",\"image\"], category e.g. \"frontier-llm\"/\"open-weight\"/\"coding\" — a coarse filing tag, not a claim). Omit any of these you don't have real material for rather than guessing.\n\nReturn ONLY a JSON array (no commentary). Each element:\n{\"entity_kind\": \"model\", \"entity_slug\": \"...\", \"entity_name\": \"...\", \"entity_owner\": \"...\", \"entity_summary\": \"...\" (optional), \"entity_links\": {...} (optional), \"entity_attributes\": {...} (optional), \"field\": \"...\", \"value\": \"...\", \"unit\": \"...\", \"quote\": \"verbatim sentence from the source\", \"url\": \"...\", \"publisher\": \"...\", \"title\": \"...\", \"published\": \"YYYY-MM or YYYY if shown\", \"staleness_days\": <suggest: 30 for prices, 200 for specs like context_window/modality, 400 for license/structural facts>}\n\nIf nothing extractable meets the bar, return []."
        },
        "next_step": "verify_and_register",
        "output_field": "candidate_claims",
        "description": "Extract atomic model claims with verbatim quotes — the checkable form"
      },
      "verify_and_register": {
        "action": "verify_and_register_directory_claims",
        "config": {
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
WHERE NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type = 'directory-researcher' AND deleted_at IS NULL)
RETURNING type, status;

-- ── Post-apply verification ────────────────────────────────────────────────
-- 1. Row exists:
--    SELECT type, status FROM agent_definitions WHERE type='directory-researcher';
-- 2. Smoke run (spawn/call via generic entry) with input_data
--    {"research_query": "Anthropic Claude Sonnet 4.6 pricing per million tokens"}, then:
--    SELECT jsonb_pretty(collected_data->'registration') FROM orchestration_states WHERE ...;
--    Expect: registered ids for claims whose quotes verify; rejected list for
--    the rest; a directory_citation_unverified work item iff rejected is non-empty.
-- 3. The registered claims appear in directory_entities/directory_claims:
--    SELECT de.slug, dc.field, dc.value, dc.status, dc.citation->>'url'
--    FROM directory_claims dc JOIN directory_entities de ON de.id = dc.entity_id
--    WHERE dc.is_current ORDER BY de.slug, dc.field;
--
-- ── Rollback ────────────────────────────────────────────────────────────────
--    UPDATE agent_definitions SET status='disabled', is_active=false
--    WHERE type='directory-researcher';
-- Facts already registered are data, versioned in directory_claims history;
-- prune by superseding the row (never edit in place).
