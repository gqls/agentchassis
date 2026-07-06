-- NNN_seed_vertical_exemplar_researcher.sql  (§B4 — the exemplar-research hop)
--
-- NEW RELAY HOP between classification and strategy: find the vertical's best
-- existing sites, read them shallowly, and distil WHY they succeed — reasons,
-- not copies — into spec aspect `vertical_landscape` for the strategist and
-- planner to consume. Design recorded in RUNBOOK_builder_route.md §B4.
--
-- REUSE-ONLY: every action already exists (read_site_spec, execute_llm_prompt,
-- firecrawl_crawl, format_crawl_for_analysis, write_site_spec,
-- create_work_item, complete_workflow). Config key shapes mirrored VERBATIM
-- from on-disk siblings (build-briefing-agent; site-adoption-agent's crawl).
-- DELIBERATE differences from adoption's crawl, per the agreed budget
-- (3 exemplars, shallow): limit 6, markdown only, only_main_content true,
-- max_discovery_depth 1 (adoption: 30/rawHtml/false/4 — one site read deeply;
-- this hop reads THREE sites lightly).
-- Known v1 limitation (accepted): one bad exemplar URL fails the item; the
-- immune system retries. The selection prompt forbids the customer's own
-- domain and requires live https URLs to reduce that risk.
-- Completion uses the PREFERRED result contract (result_from) — valid now the
-- Option A image (v1.0.1092) is deployed.
--
-- Donor for category/agent_category/status: domain-research-classifier
-- (closest sibling). UNIQUE(type,version) holds; REVERT soft-deletes.

BEGIN;

-- AMENDED post-incident (see NNN_fix_researcher_spawn_columns.sql): the
-- spawner's getAgentDefinition consumes image_repository/image_tag/command/
-- resources/health_config/env_vars/idle_timeout_seconds and gates on
-- is_active=true. Copy them from the donor or the pod boots as generic via
-- the image's default entrypoint.
INSERT INTO agent_definitions
  (type, display_name, description, category, agent_category, status, version, default_config,
   image_repository, image_tag, command, resources, health_config, env_vars,
   idle_timeout_seconds, is_active)
SELECT
  'vertical-exemplar-researcher',
  'Vertical Exemplar Researcher',
  'Relay hop between classification and strategy: selects the vertical''s three best exemplar sites, crawls each shallowly, and synthesises WHY they succeed (positioning, content, design, tools, trust) into site_specs aspect=vertical_landscape. Reasons, not copies. RUNBOOK_builder_route.md §B4.',
  d.category,
  d.agent_category,
  d.status,
  1,
  $wf$
{
  "workflow": {
    "start_step": "read_specs",
    "steps": {
      "read_specs": {
        "action": "read_site_spec",
        "config": { "site_id": "input_data.site_id" },
        "next_step": "select_exemplars",
        "description": "Load identity + classification (incl. competitors_found) from site_specs",
        "output_field": "site_specs"
      },
      "select_exemplars": {
        "action": "execute_llm_prompt",
        "config": {
          "ai_service": {
            "model": "claude-sonnet-4-6",
            "provider": "anthropic",
            "max_tokens": 1500,
            "api_key_env_var": "ANTHROPIC_API_KEY"
          },
          "input_fields": ["input_data", "site_specs"],
          "output_format": "json",
          "prompt_template": "You select exemplar websites for competitive research.\n\nDomain being built: {{.input_data.domain}}\n\n## Research Data (identity + classification for this domain)\n{{.site_specs}}\n\n## Task\nPick the THREE best EXISTING websites in this domain's vertical with the same objectives — the sites a person in this niche would call the best. Prefer sites named in identity.competitors_found when they are genuinely strong; otherwise use well-known leaders of the vertical.\n\nRules:\n- Real, currently-live sites; full https:// URLs to their front page\n- NOT the customer's own domain ({{.input_data.domain}})\n- Same site_type/objectives as the classification where possible\n- Three DIFFERENT organisations\n\nReturn ONLY valid JSON:\n{\n  \"exemplar_1\": {\"url\": \"https://...\", \"why\": \"one line\"},\n  \"exemplar_2\": {\"url\": \"https://...\", \"why\": \"one line\"},\n  \"exemplar_3\": {\"url\": \"https://...\", \"why\": \"one line\"}\n}"
        },
        "next_step": "crawl_exemplar_1",
        "description": "LLM selects three exemplar sites of the vertical (flat keys for dotted-path reads)",
        "output_field": "selected_exemplars"
      },
      "crawl_exemplar_1": {
        "action": "firecrawl_crawl",
        "config": {
          "url_field": "selected_exemplars.result.exemplar_1.url",
          "scrape_config": { "limit": 6, "formats": ["markdown"], "max_age": 600000, "only_main_content": true, "max_discovery_depth": 1 }
        },
        "next_step": "format_exemplar_1",
        "description": "Shallow crawl of exemplar 1 (front page + direct links)",
        "output_field": "crawl_1"
      },
      "format_exemplar_1": {
        "action": "format_crawl_for_analysis",
        "config": { "crawl_field": "crawl_1", "summary_chars_per_page": 400 },
        "next_step": "crawl_exemplar_2",
        "description": "Format exemplar 1 pages for the synthesis prompt",
        "output_field": "formatted_1"
      },
      "crawl_exemplar_2": {
        "action": "firecrawl_crawl",
        "config": {
          "url_field": "selected_exemplars.result.exemplar_2.url",
          "scrape_config": { "limit": 6, "formats": ["markdown"], "max_age": 600000, "only_main_content": true, "max_discovery_depth": 1 }
        },
        "next_step": "format_exemplar_2",
        "description": "Shallow crawl of exemplar 2",
        "output_field": "crawl_2"
      },
      "format_exemplar_2": {
        "action": "format_crawl_for_analysis",
        "config": { "crawl_field": "crawl_2", "summary_chars_per_page": 400 },
        "next_step": "crawl_exemplar_3",
        "description": "Format exemplar 2 pages",
        "output_field": "formatted_2"
      },
      "crawl_exemplar_3": {
        "action": "firecrawl_crawl",
        "config": {
          "url_field": "selected_exemplars.result.exemplar_3.url",
          "scrape_config": { "limit": 6, "formats": ["markdown"], "max_age": 600000, "only_main_content": true, "max_discovery_depth": 1 }
        },
        "next_step": "format_exemplar_3",
        "description": "Shallow crawl of exemplar 3",
        "output_field": "crawl_3"
      },
      "format_exemplar_3": {
        "action": "format_crawl_for_analysis",
        "config": { "crawl_field": "crawl_3", "summary_chars_per_page": 400 },
        "next_step": "synthesise",
        "description": "Format exemplar 3 pages",
        "output_field": "formatted_3"
      },
      "synthesise": {
        "action": "execute_llm_prompt",
        "config": {
          "ai_service": {
            "model": "claude-sonnet-4-6",
            "provider": "anthropic",
            "max_tokens": 8000,
            "api_key_env_var": "ANTHROPIC_API_KEY"
          },
          "input_fields": ["input_data", "site_specs", "selected_exemplars", "formatted_1", "formatted_2", "formatted_3"],
          "output_format": "json",
          "prompt_template": "You are a competitive analyst. Three exemplar websites of a vertical were crawled. Work out WHY they succeed and what a NEW site in this vertical should learn. Reasons, NOT copies — never reproduce their text or layout; extract the underlying causes of success.\n\nDomain being built: {{.input_data.domain}}\n\n## This domain's identity + classification\n{{.site_specs}}\n\n## Selected exemplars (and why)\n{{.selected_exemplars}}\n\n## Exemplar 1 content\n{{.formatted_1}}\n\n## Exemplar 2 content\n{{.formatted_2}}\n\n## Exemplar 3 content\n{{.formatted_3}}\n\n## Task\nFor EACH exemplar: its positioning, its success factors (why it works for its users), notable tools/interactive features, and trust signals. Then across all three: the shared content patterns (types, depth, cadence), design patterns, and tool patterns. Finally the lessons for OUR site: what to adopt (proven in this vertical), what to adapt (right idea, do it our way), what to avoid, and the differentiation opportunity none of them covers.\n\nReturn ONLY valid JSON:\n{\n  \"exemplars_analyzed\": [\n    {\"url\": \"...\", \"positioning\": \"...\", \"success_factors\": [\"...\"], \"notable_tools\": [\"...\"], \"trust_signals\": [\"...\"]}\n  ],\n  \"patterns\": {\"content\": \"...\", \"design\": \"...\", \"tools\": \"...\"},\n  \"lessons\": {\"adopt\": [\"...\"], \"adapt\": [\"...\"], \"avoid\": [\"...\"]},\n  \"differentiation_opportunity\": \"...\",\n  \"confidence\": 0.0\n}"
        },
        "next_step": "write_landscape_spec",
        "description": "Synthesise WHY the exemplars succeed + lessons for our site",
        "output_field": "vertical_synthesis"
      },
      "write_landscape_spec": {
        "action": "write_site_spec",
        "config": {
          "aspect": "vertical_landscape",
          "source": "vertical-exemplar-researcher",
          "site_id": "input_data.site_id",
          "spec_data": "vertical_synthesis.result",
          "source_agent": "vertical-exemplar-researcher",
          "source_item_id": "input_data.work_item_id"
        },
        "next_step": "create_next_item",
        "description": "Persist the landscape synthesis to site_specs",
        "output_field": "landscape_written"
      },
      "create_next_item": {
        "action": "create_work_item",
        "config": {
          "source": "vertical-exemplar-researcher",
          "site_id": "input_data.site_id",
          "summary": "Strategy needed after vertical exemplar research",
          "priority": 8,
          "severity": "high",
          "item_type": "needs_strategy",
          "item_domain": "build",
          "handler_agent": "domain-strategist",
          "item_key_prefix": "strategy"
        },
        "next_step": "complete",
        "description": "Chain to the strategist (the hop this agent was inserted before)",
        "output_field": "next_item_created"
      },
      "complete": {
        "action": "complete_workflow",
        "config": { "result_from": "landscape_written" },
        "description": "Vertical research complete"
      }
    }
  }
}
$wf$::jsonb
  , d.image_repository, d.image_tag, d.command, d.resources, d.health_config,
    d.env_vars, d.idle_timeout_seconds, true
FROM agent_definitions d
WHERE d.type = 'page-build-handler'  -- spawn-proven donor for the infra columns
  AND COALESCE(d.is_snapshot, false) = false AND d.deleted_at IS NULL
ON CONFLICT (type, version) DO NOTHING;
-- NOTE: donor changed classifier→page-build-handler so ALL copied columns come
-- from a row the dispatcher demonstrably spawns; category/agent_category/status
-- now also inherit from it (equivalent values). ON CONFLICT makes re-runs
-- harmless (the 2026-07-06 duplicate-key on re-application was expected: the
-- agent already existed; the live row was corrected by
-- NNN_fix_researcher_spawn_columns.sql, snapshot 139362d5, image v1.0.1094).

-- verify: one live row, workflow parses
SELECT type, version, status,
       jsonb_object_keys(default_config->'workflow'->'steps') AS steps
FROM agent_definitions
WHERE type = 'vertical-exemplar-researcher'
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

COMMIT;

-- ── REVERT ──────────────────────────────────────────────────────────────────
-- BEGIN;
-- UPDATE agent_definitions SET deleted_at = now()
-- WHERE type = 'vertical-exemplar-researcher' AND COALESCE(is_snapshot,false)=false;
-- COMMIT;
