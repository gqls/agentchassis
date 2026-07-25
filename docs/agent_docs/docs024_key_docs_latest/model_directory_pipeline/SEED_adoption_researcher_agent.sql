-- SEED — adoption-researcher agent (Phase E: the company AI-agent adoption
-- tracker, second half of the owner's original brief)
--
-- NO IMAGE ROLL REQUIRED. Every action this agent uses already ships:
-- web_search, prepare_urls, batch_webscrape, execute_llm_prompt, and
-- verify_and_register_directory_claims — whose candidate handling has been
-- kind-agnostic since Phase B (directory_claims.go:121 reads entity_kind
-- from the candidate and only DEFAULTS it to "model"). So the acquisition
-- lane for kind='company' can run, and fill the register, before the
-- read/publish/discovery generalisation reaches production. That ordering is
-- deliberate: when the image does roll, the tracker has data on day one
-- instead of waiting a week for its first research sweep.
--
-- It is a SIBLING of directory-researcher, not a parameterisation of it. The
-- workflow shape is identical, but the extraction prompt is the whole value
-- of these agents and the two prompts have almost nothing in common: one
-- extracts a price from a vendor's pricing table, the other extracts a
-- claimed business result from a case study written to flatter the vendor.
-- The second needs different instructions about what NOT to believe.
--
-- WHAT IT MUST NOT DO — the reason the prompt is long. Adoption reporting is
-- the most fabrication-prone material this platform has touched:
-- bugs_open/043 (invented statistics in site copy) and bugs_open/061 (an LLM
-- inventing an entire price table from a prompt's example) both came from
-- models filling a plausible shape with plausible numbers. A vendor case
-- study saying "customers see up to 40% faster resolution" is NOT a claim
-- that a named company measured 40%. So: the ROI figure and HOW IT WAS
-- MEASURED are separate fields, and an unmeasured claim must be recorded as
-- unmeasured rather than dropped or dressed up — "the company stated X and
-- did not say how it was measured" is the honest and more useful fact.
--
-- The verifier is unchanged and is what makes this safe: every candidate's
-- URL is re-fetched and the quote must appear verbatim, or it never enters
-- the register (it becomes a directory_citation_unverified item for review).
--
-- TIMEOUT GOTCHA (bugs_open/062): timeout_seconds must live INSIDE the
-- step's "config" object — models.Step has no step-level timeout field, so a
-- step-level value is silently dropped and the await runs at the 180s
-- default. scrape_config.formats=["markdown"] keeps the batch reply inside
-- Kafka's max message size (062's root cause).
--
-- Invocation:
--   input_data: { research_query }
--   examples:
--     "companies deploying AI agents in production 2026 measured results"
--     "Model Context Protocol MCP enterprise adoption announcements"
--     "customer service AI agent deployment cost savings case study 2026"

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status, is_active,
    image_repository, image_tag, input_contract, output_contract, default_config
)
SELECT
    'adoption-researcher',
    'AI Adoption Researcher',
    'Adoption tracker acquisition lane: researches which organisations are deploying AI agents and with what stated result, extracts atomic candidate claims (framework used, rollout scope, claimed ROI, HOW that ROI was measured) with verbatim quotes, and registers only those whose cited source demonstrably contains the quote (verify_and_register_directory_claims) into directory_entities/directory_claims as kind=company / kind=protocol. Unverifiable candidates terminate at human review.',
    'analyst', 'analyst', 'active', true,
    'docker.io/aqls/agent-chassis',
    (SELECT image_tag FROM agent_definitions WHERE type = 'claims-auditor'),
    '{"required": ["research_query"]}'::jsonb,
    '{"produces": {"registration": "verified company/protocol adoption claims added to the directory register; rejects raised for human review"}}'::jsonb,
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
          "exclude_domains": ["pinterest.com", "facebook.com", "twitter.com", "instagram.com", "reddit.com"]
        },
        "next_step": "scrape_pages",
        "output_field": "prepared_urls",
        "description": "Pick the strongest sources. NOTE the deliberate absence of prefer_domains: unlike the model directory, where the vendor's own pricing page is the most authoritative source, here the vendor's own case study is the most INTERESTED one. Engineering blogs, earnings calls, regulatory filings and trade press are all legitimate, and preferring vendor domains would systematically bias the register towards marketing claims."
      },
      "scrape_pages": {
        "action": "batch_webscrape",
        "config": {
          "urls_field": "prepared_urls.urls_to_scrape",
          "scrape_config": {"only_main_content": true, "capture_screenshot": false, "formats": ["markdown"]},
          "timeout_seconds": 240
        },
        "next_step": "extract_claims",
        "output_field": "scrape_results",
        "description": "Scrape the selected sources"
      },
      "extract_claims": {
        "action": "execute_llm_prompt",
        "config": {
          "ai_service": {"model": "claude-sonnet-4-6", "provider": "anthropic", "max_tokens": 8000, "api_key_env_var": "ANTHROPIC_API_KEY"},
          "input_fields": ["scrape_results", "prepared_urls", "input_data"],
          "output_format": "json",
          "prompt_template": "You are extracting ATOMIC, CITABLE claims about ORGANISATIONS DEPLOYING AI AGENTS, and about AGENT COMMUNICATION PROTOCOLS, for a public adoption tracker. Research question: {{.input_data.research_query}}\n\nSCRAPED SOURCES (each has a url):\n{{.scrape_results}}\n\nExtract up to 15 candidate claims, one per (organisation, fact) pair. Each MUST be atomic and verifiable:\n- name the specific organisation (entity_kind: \"company\"; entity_slug: lowercase-hyphenated, e.g. 'intuit'; entity_name: display name; entity_owner: the parent group if the source names one, else the same as entity_name) — or, for a protocol rather than an organisation, entity_kind: \"protocol\" with entity_owner set to its steward/publisher;\n- field: one of agent_framework, use_case, rollout_scope, agent_count, roi_claimed, roi_basis, deployment_date, protocol_adopted, protocol_spec_version, protocol_governance (or another short snake_case field name if the source states a different concrete, checkable fact);\n- a VERBATIM quote copied EXACTLY from ONE source's text that states it — do not paraphrase, do not stitch two sentences together, do not normalise numbers, percentages, currencies or units;\n- the url of that exact source page (from the material above — never invent or shorten one).\n\nEvery claim will be machine-checked: the url will be re-fetched and rejected unless your quote appears in it verbatim. A paraphrased quote fails. A summary fails. If a source page does not literally contain a sentence stating the fact, DO NOT include that claim.\n\nTHE THINGS THAT MATTER MOST HERE, because this material is written to persuade:\n1. A RESULT AND ITS MEASUREMENT ARE TWO CLAIMS. If a source says a company cut handling time 40%, emit roi_claimed with that quote. Emit roi_basis ONLY if the source separately says how it was measured (A/B test, before/after over a stated period, an internal estimate, a survey). If the source gives a number and no method, emit roi_claimed alone — do NOT invent a basis, and do NOT drop the claim. 'Stated, method not given' is the honest record and is exactly what a reader of this directory needs to know.\n2. WHO IS SPEAKING. A vendor's marketing page saying its customers see gains is NOT a named company reporting a measured result. Only extract a claim about company X when the source names X. Anonymous customers ('a Fortune 500 retailer') are not entities — skip them.\n3. GENERIC MARKET STATISTICS ARE NOT ADOPTION FACTS. 'Analysts expect 40% of enterprises to deploy agents by 2027' is a forecast about nobody. Skip it.\n4. A PILOT IS NOT A ROLLOUT. Record rollout_scope in the source's own words (pilot / one team / one department / company-wide) rather than upgrading it.\n5. NEVER COPY A NUMBER FROM THESE INSTRUCTIONS. The example figures above are illustrations of shape, not data.\n\nOptionally, on the FIRST claim you emit for a given entity, also include entity_summary (one plain sentence: what the organisation does, or what the protocol is for), entity_links (object: docs [the source's own page about the deployment or the protocol spec], video_urls [array, only if you found a genuine talk/demo URL in the material]), and entity_attributes (object: sector e.g. \"financial-services\", region, and for protocols steward/status — coarse filing tags, not claims). Omit any you don't have real material for rather than guessing.\n\nReturn ONLY a JSON array (no commentary). Each element:\n{\"entity_kind\": \"company\"|\"protocol\", \"entity_slug\": \"...\", \"entity_name\": \"...\", \"entity_owner\": \"...\", \"entity_summary\": \"...\" (optional), \"entity_links\": {...} (optional), \"entity_attributes\": {...} (optional), \"field\": \"...\", \"value\": \"...\", \"unit\": \"...\", \"quote\": \"verbatim sentence from the source\", \"url\": \"...\", \"publisher\": \"...\", \"title\": \"...\", \"published\": \"YYYY-MM or YYYY if shown\", \"staleness_days\": <suggest: 400 for a dated announcement or a measured result, which does not become false with time; 200 for rollout_scope and agent_count, which do drift; 200 for protocol_spec_version>}\n\nIf nothing extractable meets the bar, return []."
        },
        "next_step": "verify_and_register",
        "output_field": "candidate_claims",
        "description": "Extract atomic adoption claims with verbatim quotes — the checkable form"
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
WHERE NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type = 'adoption-researcher' AND deleted_at IS NULL)
RETURNING type, status;

-- ── Post-apply verification ────────────────────────────────────────────────
-- 1. Row exists:
--      SELECT type, status FROM agent_definitions WHERE type='adoption-researcher';
-- 2. Smoke run — dispatch via the generic entry with the FULL ten-header kcat
--    set (RUNBOOK: a partial header set silently produces no run):
--      input_data {"research_query": "companies deploying AI agents in production 2026 measured results"}
-- 3. What good looks like — company entities with BOTH a result and its
--    basis where the source gave one, and roi_claimed alone where it didn't:
--      SELECT de.name, dc.field, dc.value, dc.status, dc.citation->>'url'
--      FROM directory_claims dc JOIN directory_entities de ON de.id = dc.entity_id
--      WHERE de.kind = 'company' AND dc.is_current ORDER BY de.name, dc.field;
-- 4. What to CHECK rather than assume, first run: that rejects are landing as
--    directory_citation_unverified rather than everything sailing through —
--    an all-verified first run on marketing material would be more suspicious
--    than a mixed one.
--      SELECT status, count(*) FROM directory_claims WHERE is_current GROUP BY 1;
--      SELECT count(*) FROM site_work_items WHERE item_type='directory_citation_unverified';
