-- ============================================================================
-- 224_grounded_explainer_agent.sql — a generic HIGH-ATTENTION content lane
--
-- WHAT THIS IS FOR (owner direction, 2026-07-26)
--   Some pages must not be written from a model's memory. Anything explaining
--   law, regulation, safety, medicine, tax, or any other domain where a
--   confident wrong sentence does real harm needs its facts fetched, quoted and
--   re-checked before a word of prose is written — and then read by a human
--   before it ships.
--
--   The normal content path is the opposite shape: research is optional, it
--   feeds one prompt as loose context, and the writer's output goes straight to
--   the page. That is right for most of the fleet and wrong for this class.
--
--   THE OCCASION: oufe.com needs explainers on UK restructuring mechanism. Those
--   are statements of statute. Writing them from memory is exactly the failure
--   the site now promises it does not commit, so the site could not have its
--   most important content until this lane existed.
--
-- WHY THE ATTENTION IS STRUCTURAL, NOT ADVISORY
--   "Be careful" in a prompt is not a control. Every step below removes a way to
--   be careless:
--     * facts are acquired by SEARCH, not recall;
--     * each candidate must carry a VERBATIM quote from one named source;
--     * verify_and_register_citations RE-FETCHES every url and discards any
--       claim whose quote is not literally present — the model proposes, the
--       fetcher disposes;
--     * the composer is handed ONLY the surviving claims and told what it may
--       not assert;
--     * a separate grounding audit re-reads the draft against that same set and
--       lists every sentence it cannot trace;
--     * the run ENDS at needs_human_review. It never publishes. There is no
--       config flag to make it publish.
--
--   The last one is the load-bearing one. An automated content lane that can
--   publish will eventually publish something wrong at 3am; one that cannot,
--   cannot.
--
-- GENERIC BY DESIGN
--   Nothing here is oufe-specific. Any site, any topic:
--     input_data: {domain, topic, research_query, page_name?, audience?,
--                  prefer_domains?[], must_not_assert?}
--   `prefer_domains` is the one dial that matters per use: point it at the
--   primary publishers for the field (legislation.gov.uk and judiciary.uk for UK
--   law; nice.org.uk for clinical; hmrc for tax) so the acquisition step reaches
--   for the statute rather than for somebody's blog about the statute.
--
-- REUSE, NOT REINVENTION
--   Steps 2-6 are the proven V5 acquisition chain from SEED_evidence_researcher
--   (same actions, same config shapes). This adds the two steps V5 has no reason
--   to have — compose and audit — plus the human gate. Facts it registers become
--   writer-usable via V2 and re-verified via V4 like any other, so a page written
--   this way keeps its evidence current afterwards.
--
-- PREREQUISITE (checked before applying, not assumed)
--   The image must carry verify_and_register_citations:
--     kubectl exec -n ai-persona-system <chassis-pod> -- \
--       sh -c 'strings /app/agent-chassis | grep -c verify_and_register_citations'
-- ============================================================================

\set ON_ERROR_STOP on

-- NOTE: the column is display_name, NOT name (a plain `name` column does not
-- exist on agent_definitions; the first attempt at this insert failed on it).
INSERT INTO agent_definitions (
  id, type, display_name, description, category, status, is_active, default_config
)
SELECT
  gen_random_uuid(),
  'grounded-explainer',
  'Grounded explainer writer',
  'High-attention content lane for pages whose facts must not come from model memory (law, regulation, safety, clinical, tax). Researches the open web, extracts atomic claims with verbatim quotes, machine-verifies each quote by re-fetching its source, composes the page from the surviving claims only, audits its own draft for ungrounded assertions, and terminates at human review. It cannot publish.',
  'content',
  'experimental',
  true,
  $cfg${
  "workflow": {
    "start_step": "ensure_site_record",
    "processing_mode": "orchestrator",
    "timeout_seconds": 900,
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
        "description": "Acquire candidate sources by SEARCH — never from recall"
      },

      "prepare_urls": {
        "action": "prepare_urls",
        "config": {
          "max_scrapes": 5,
          "max_snippets": 6,
          "prefer_domains": [".gov", ".gov.uk", "legislation.gov.uk", "judiciary.uk", "nationalarchives.gov.uk", "bailii.org", ".org", ".edu", ".ac.uk"],
          "exclude_domains": ["pinterest.com", "facebook.com", "twitter.com", "instagram.com", "youtube.com", "reddit.com", "quora.com", "medium.com"]
        },
        "next_step": "scrape_pages",
        "output_field": "prepared_urls",
        "description": "Prefer primary publishers over commentary about them"
      },

      "scrape_pages": {
        "action": "batch_webscrape",
        "config": {
          "urls_field": "prepared_urls.urls_to_scrape",
          "scrape_config": {"only_main_content": true, "capture_screenshot": false}
        },
        "next_step": "extract_claims",
        "output_field": "scrape_results",
        "timeout_seconds": 180,
        "description": "Fetch the sources themselves"
      },

      "extract_claims": {
        "action": "execute_llm_prompt",
        "config": {
          "ai_service": {"model": "claude-sonnet-5", "provider": "anthropic", "max_tokens": 8000, "api_key_env_var": "ANTHROPIC_API_KEY"},
          "input_fields": ["scrape_results", "prepared_urls", "input_data"],
          "output_format": "json",
          "error_step": "complete_no_sources",
          "prompt_template": "You are extracting ATOMIC, CITABLE claims for an evidence register, to ground an explainer on: {{.input_data.topic}}\n\nResearch question: {{.input_data.research_query}}\n\nSCRAPED SOURCES (each has a url):\n{{.scrape_results}}\n\nExtract up to 14 candidate claims. Each MUST be atomic and verifiable:\n- a specific, dated factual statement, a defined term, a stated condition, or a specific figure — never a vague characterisation (\"is widely used\" is NOT a claim);\n- a VERBATIM quote copied EXACTLY from ONE source's text that states it — do not paraphrase, do not stitch two sentences together, do not normalise numbers or punctuation;\n- the url of that exact source page, taken from the material above — never invent, shorten or reconstruct a url.\n\nFor an explainer about rules or law, the claims that matter most are the ones a reader could get WRONG: the exact condition that must be satisfied, the threshold, who decides, what the alternative is. Prefer those over background colour.\n\nEvery claim will be machine-checked: the url is re-fetched and the claim is DISCARDED unless your quote appears in it verbatim. A paraphrase fails. A summary fails. If a source does not literally contain a sentence stating the fact, DO NOT include that claim — omitting it is correct behaviour and costs nothing, because the composer is told to write around gaps.\n\nPrefer the primary instrument (the statute, the judgment, the regulator's own page) over any article describing it; where only second-hand is available set \"secondhand\": true.\n\nReturn ONLY a JSON array (no commentary). Each element:\n{\"claim\": \"plain statement of the fact\", \"value\": <number if it carries one>, \"unit\": \"...\", \"quote\": \"verbatim sentence from the source\", \"url\": \"...\", \"publisher\": \"...\", \"title\": \"...\", \"published\": \"YYYY-MM or YYYY if shown\", \"secondhand\": false, \"staleness_days\": <200 prices/market, 400 annual statistics, 800 structural/statutory>, \"writer_line\": \"how copy should state it, with {value} where a number goes and the source named\"}\n\nIf nothing meets the bar, return []."
        },
        "next_step": "verify_and_register",
        "output_field": "candidate_claims",
        "description": "Candidates must arrive in checkable form or not at all"
      },

      "verify_and_register": {
        "action": "verify_and_register_citations",
        "config": {
          "site_id": "site_record.site_id",
          "candidates": "candidate_claims.result"
        },
        "next_step": "compose_explainer",
        "output_field": "registration",
        "description": "THE DISPOSAL STEP: re-fetch every url, keep only quotes that are really there"
      },

      "compose_explainer": {
        "action": "execute_llm_prompt",
        "config": {
          "ai_service": {"model": "claude-sonnet-5", "provider": "anthropic", "max_tokens": 12000, "api_key_env_var": "ANTHROPIC_API_KEY"},
          "input_fields": ["input_data", "registration", "site_record"],
          "output_format": "json",
          "prompt_template": "Write an explainer for {{.input_data.domain}} on: {{.input_data.topic}}\n\nAudience: {{if .input_data.audience}}{{.input_data.audience}}{{else}}an intelligent reader who is new to this subject and wants the mechanism, not a summary{{end}}\n\n## The ONLY externally-sourced facts you may assert\n\nThese survived machine verification — each quote was found in the live source:\n\n{{.registration}}\n\n## What you may write, and what you may not\n\nYou may explain MECHANISM freely: how a process works, what follows from what, why a party would do one thing rather than another, and worked arithmetic using figures you invent and label as hypothetical. A mechanism explained with openly hypothetical numbers cannot be wrong about anybody, and it is the part of this page that transfers to every situation the reader will meet.\n\nYou may NOT assert any of the following unless it appears in the verified list above:\n- a legal condition, threshold, percentage, time limit, or definition;\n- what a court, regulator or statute requires, permits or has decided;\n- any fact about a real named organisation, case or transaction;\n- any statistic, quantity or date.\n\nWhere you need one of those and it is not in the list, DO NOT reach for what you think you remember. Write the sentence so it does not need the fact, or say plainly that we have not verified it and name the kind of document it would come from. \"We have not verified the current threshold against the statute\" is publishable. A remembered number is not.\n\n{{if .input_data.must_not_assert}}Additional prohibition for this site: {{.input_data.must_not_assert}}\n{{end}}\n\n## Tone\n\nPlain, precise, unhurried. Explain a term the first time it appears and then use it properly — the reader wants the real vocabulary. No hype, no urgency, nothing that reads like marketing. Never imply the reader should act on this.\n\n## Honesty about this page itself\n\nThe page must include, in its own voice and not as boilerplate, that this is an explanation which can be wrong, that our reading of a source can be wrong, and that a reader should check anything that matters against the primary document. Do NOT claim the page is verified, authoritative, complete or reliable.\n\nReturn ONLY this JSON:\n{\"heading\": \"page heading\", \"content\": \"the explainer as clean semantic HTML — <p>, <h3>, <ul>, <strong> only; no inline styles, no scripts, no images\", \"sources_used\": [\"url\", ...], \"gaps\": [\"each fact you needed and could not verify, stated plainly\"], \"hypothetical_figures_used\": true|false}"
        },
        "next_step": "audit_grounding",
        "output_field": "draft",
        "description": "Compose from the surviving claims only"
      },

      "audit_grounding": {
        "action": "execute_llm_prompt",
        "config": {
          "ai_service": {"model": "claude-sonnet-5", "provider": "anthropic", "max_tokens": 6000, "api_key_env_var": "ANTHROPIC_API_KEY"},
          "input_fields": ["draft", "registration", "input_data"],
          "output_format": "json",
          "prompt_template": "You are auditing a draft explainer for ungrounded assertions. You did not write it and you have no stake in it shipping.\n\n## The verified facts that were available to the writer\n{{.registration}}\n\n## The draft\n{{.draft}}\n\nGo sentence by sentence. For each assertion, decide which it is:\n- GROUNDED — traceable to a verified fact above;\n- MECHANISM — an explanation of how something works, or arithmetic on openly hypothetical figures, which needs no external source and is fine;\n- HEDGED — explicitly marked as unverified or unknown, which is fine;\n- UNGROUNDED — a legal condition, threshold, definition, decision, statistic, date, or claim about a real named organisation that does NOT appear in the verified list.\n\nUNGROUNDED is the only category that matters. Be strict about it: a confident sentence about what a statute requires is ungrounded even when it sounds obviously true, because sounding true is exactly how a wrong one survives. Do not give the benefit of the doubt.\n\nAlso flag separately any sentence that claims the page or the site is accurate, verified, complete, authoritative, or can be relied upon — those are claims about us that no check can support.\n\nReturn ONLY:\n{\"ungrounded\": [{\"sentence\": \"...\", \"why\": \"...\"}], \"reliability_overclaims\": [\"...\"], \"verdict\": \"clean|needs_revision\", \"notes\": \"...\"}"
        },
        "next_step": "create_review_item",
        "output_field": "grounding_audit",
        "description": "Independent re-read: every sentence the draft cannot support"
      },

      "create_review_item": {
        "action": "create_work_item",
        "config": {
          "site_id": "site_record.site_id",
          "item_type": "grounded_draft_review",
          "status": "needs_human_review",
          "handler_agent": "human-review",
          "priority": 20,
          "severity": "medium",
          "summary_template": "Grounded explainer draft ready for review: {{.input_data.topic}}",
          "spec_fields": ["draft", "grounding_audit", "registration", "input_data"]
        },
        "next_step": "complete",
        "output_field": "review_item",
        "description": "THE GATE. The run ends here by design — this agent cannot publish."
      },

      "complete_no_sources": {
        "action": "complete_workflow",
        "config": {"output_fields": ["search_results", "prepared_urls"]},
        "description": "Nothing citable was found. Ending with no draft is the correct outcome, not a failure to route around."
      },

      "complete": {
        "action": "complete_workflow",
        "config": {"output_fields": ["registration", "draft", "grounding_audit", "review_item"]}
      }
    }
  }
}$cfg$::jsonb
WHERE NOT EXISTS (
  SELECT 1 FROM agent_definitions WHERE type = 'grounded-explainer' AND deleted_at IS NULL
);

-- Verify
--   SELECT type, status, is_active,
--          jsonb_array_length(jsonb_path_query_array(default_config,'$.workflow.steps.keyvalue().key')) AS steps
--     FROM agent_definitions WHERE type='grounded-explainer' AND deleted_at IS NULL;
--
-- Invoke (generic — any site, any topic):
--   ./docs/agent_docs/docs024_key_docs_latest/oufe/TRIGGER_grounded_explainer.sh \
--       <domain> "<topic>" "<research query>" ["<audience>"]
--
-- Read the result:
--   SELECT summary, spec->'grounding_audit' FROM site_work_items
--    WHERE item_type='grounded_draft_review' ORDER BY created_at DESC LIMIT 1;
