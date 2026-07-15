-- 156: product-spec-refresher agent (drives refresh_product_specs)
--
-- Context (2026-07-15, empty_sections_loop_integrity Session 8): closes the
-- named gap from Session 6 — the first gripper specs were researched by hand
-- with no reusable way to re-verify them or add a manufacturer. This agent
-- re-scrapes each product's own source_url via Firecrawl and refreshes the
-- specs with a strictly-grounded LLM extraction (factual-only, refuse-if-
-- absent, closed key set). It does NOT web-search/discover — a human adds a
-- new manufacturer by inserting a products stub (name + slug + category +
-- content_data.source_url); this agent then fills the specs. Discovery stays
-- a human judgement (that's where wrong-product fabrication risk lives);
-- extraction+write is automated and reproducible with provenance.
--
-- processing_mode: task — directly orchestratable via kcat with
-- {site_id, category?}. Requires the chassis image that registers
-- refresh_product_specs, and FIRECRAWL_API_KEY + OLLAMA reachable on the pod
-- (both confirmed present on agent-chassis 2026-07-15).
--
-- MANUAL posture: no scheduled_tasks row is created — fired by hand, matching
-- the workstream's "more awareness before more autonomy" stance. Add a
-- scheduled_tasks entry later if periodic re-verification is wanted.
--
-- Verify after applying:
--   SELECT type, status FROM agent_definitions WHERE type='product-spec-refresher';
--   -- then trigger (RUNBOOK §8) and inspect products.content_data->>'verified_date'

BEGIN;

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    is_active, capabilities, image_repository, image_tag, default_config
) VALUES (
    'product-spec-refresher',
    'Product Spec Refresher',
    'Re-scrapes each product source_url and refreshes specs via grounded LLM extraction (factual-only, provenance-stamped). Refresh, not discovery.',
    'specialist',
    'specialist',
    'active',
    true,
    '["scraping", "product-data", "grounded-extraction"]'::jsonb,
    'docker.io/aqls/agent-chassis',
    'latest',
    '{
      "workflow": {
        "start_step": "ensure_site_record",
        "processing_mode": "task",
        "timeout_seconds": 900,
        "steps": {
          "ensure_site_record": {
            "action": "ensure_site_record",
            "config": { "input_fields": ["site_id", "domain"] },
            "next_step": "refresh_specs",
            "description": "Resolve site record from site_id or domain",
            "output_field": "site_record"
          },
          "refresh_specs": {
            "action": "refresh_product_specs",
            "config": {
              "site_id": "site_record.site_id",
              "category?": "input_data.category",
              "limit?": "input_data.limit"
            },
            "next_step": "complete",
            "description": "Firecrawl each product source_url, ground-extract specs, refresh rows",
            "output_field": "refresh_result"
          },
          "complete": {
            "action": "complete_workflow",
            "config": { "output_fields": ["refresh_result"] },
            "description": "Refresh pass complete"
          }
        }
      }
    }'::jsonb
);

COMMIT;
