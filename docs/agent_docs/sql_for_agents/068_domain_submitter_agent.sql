-- ============================================================================
-- domain-submitter agent
--
-- Receives: domain, email (optional), phone (optional)
-- Does: creates site record, stores contact info, creates needs_domain_research
-- work item. The dispatch loop then picks it up and routes to the classifier.
--
-- No new Go code. Uses existing actions:
--   ensure_site_record, update_site_content, create_work_item, complete_workflow
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, agent_category, status,
    image_repository, image_tag, category,
    default_config, input_contract, output_contract, domain_tags
) VALUES (
             'domain-submitter',
             'Domain Submitter',
             'Entry point for new domain submissions. Creates site record, stores contact info, creates needs_domain_research work item for the classifier. Minimal input required — just the domain name.',
             'coordinator', 'active',
             'docker.io/aqls/agent-chassis', 'v1.0.837', 'coordinator',
             '{"workflow":{"start_step":"ensure_site_record","processing_mode":"orchestrator","timeout_seconds":60,"steps":{
                 "ensure_site_record":{
                     "action":"ensure_site_record",
                     "config":{"store_brief_in_content_data":false},
                     "next_step":"store_contact_info",
                     "description":"Create or find site record for this domain",
                     "output_field":"site_record"
                 },
                 "store_contact_info":{
                     "action":"update_site_content",
                     "config":{
                         "merge":true,
                         "sync_columns":true,
                         "content_field":"input_data",
                         "site_id_field":"site_record.site_id"
                     },
                     "next_step":"create_research_item",
                     "description":"Store email and phone on the site record",
                     "output_field":"contact_stored"
                 },
                 "create_research_item":{
                     "action":"create_work_item",
                     "config":{
                         "site_id":"site_record.site_id",
                         "item_type":"needs_domain_research",
                         "handler_agent":"domain-research-classifier",
                         "item_domain":"build",
                         "source":"domain-submitter",
                         "priority":5,
                         "severity":"high",
                         "summary":"Research and classify domain",
                         "item_key_prefix":"research"
                     },
                     "next_step":"complete",
                     "description":"Create the first work item — classifier will research and classify",
                     "output_field":"research_item"
                 },
                 "complete":{
                     "action":"complete_workflow",
                     "config":{"output_fields":["site_record","research_item"]},
                     "description":"Domain submitted — classifier will pick it up"
                 }
             }}}'::jsonb,
             '{"required": ["domain"], "optional": ["email", "phone", "objective"]}'::jsonb,
             '{"produces": {"site_record": "site record with site_id", "research_item": "needs_domain_research work item"}}'::jsonb,
             '["intake", "submission"]'::jsonb
         ) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description,
    image_tag = EXCLUDED.image_tag,
    updated_at = NOW();

-- Verify
SELECT type, display_name, status
FROM agent_definitions WHERE type = 'domain-submitter' AND deleted_at IS NULL;

-- Add store_submission_spec step to domain-submitter
-- Goes between store_contact_info and create_research_item

-- Change store_contact_info to route to the new step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_contact_info,next_step}',
        '"store_submission_spec"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'domain-submitter' AND deleted_at IS NULL;

-- Add the new step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_submission_spec}',
        '{
            "action": "write_site_spec",
            "config": {
                "site_id": "site_record.site_id",
                "spec_data": "input_data",
                "aspect": "submission",
                "source": "domain-submitter",
                "source_agent": "domain-submitter"
            },
            "next_step": "create_research_item",
            "description": "Store full submission data in site_specs for classifier to read",
            "output_field": "submission_stored"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'domain-submitter' AND deleted_at IS NULL;

-- Verify the chain: ensure_site_record → store_contact_info → store_submission_spec → create_research_item → complete
SELECT
    default_config->'workflow'->'steps'->'ensure_site_record'->>'next_step' as step1_next,
    default_config->'workflow'->'steps'->'store_contact_info'->>'next_step' as step2_next,
    default_config->'workflow'->'steps'->'store_submission_spec'->>'next_step' as step3_next,
    default_config->'workflow'->'steps'->'create_research_item'->>'next_step' as step4_next
FROM agent_definitions
WHERE type = 'domain-submitter' AND deleted_at IS NULL;


The `input_data` in the Kafka message is already arbitrary JSON. The domain-submitter just needs to persist whatever's in it. The approach:

The submitter writes everything it receives to `site_specs` as a `submission` aspect. The classifier reads it alongside its web research. No new actions needed — just add a `write_site_spec` step to the domain-submitter workflow.

```json
// Simple submission (what we have now)
{"domain": "dartsonline.com", "email": "darts@contactforsales.com"}

// Rich submission (future)
{
    "domain": "dartsonline.com",
    "email": "darts@contactforsales.com",
    "phone": "+44 7934 524 911",
    "hints": {
        "purpose": "we want to sell leads to ticket providers",
        "not_a": "we are NOT a darts shop or equipment seller",
        "must_have": ["event calendar", "player profiles", "news feed"],
        "nice_to_have": ["live scores", "betting odds comparison"],
        "tone": "enthusiastic but knowledgeable, like a pub commentator who knows stats"
    },
    "fixed_content": {
        "about": "DartsOnline was founded in 2024 by...",
        "contact_address": "123 Oche Lane, London"
    },
    "brand": {
        "logo_url": "s3://bucket/submitted-logos/dartsonline.png",
        "colours": {"primary": "#dc2626", "accent": "#fbbf24"},
        "existing_site": "https://old.dartsonline.com"
    },
    "assets": [
        {"purpose": "logo", "url": "s3://bucket/submitted/logo.png"},
        {"purpose": "hero", "url": "s3://bucket/submitted/hero.jpg"}
    ]
}
```

The domain-submitter stores the full `input_data` as a `submission` aspect in `site_specs`. The classifier's prompt already loads all specs via `read_site_spec` — it would see the submission alongside identity, classification etc. We'd add a line to the classifier prompt: "If a submission aspect exists, treat its hints as strong guidance — the human knows what they want."

For images, the assets array would be processed by adding a step in the domain-submitter that calls `store_asset` for each submitted asset. The URIs get stored in the assets table and the classifier/planner can reference them (skipping image generation for assets that are already provided).

The change to `domain-submitter` is small — one additional step:

```json
"store_submission_spec": {
    "action": "write_site_spec",
    "config": {
        "site_id": "site_record.site_id",
        "spec_data": "input_data",
        "aspect": "submission",
        "source": "domain-submitter",
        "source_agent": "domain-submitter"
    },
    "next_step": "create_research_item",
    "description": "Store full submission data in site_specs for classifier to read"
}
```

This means everything submitted — hints, fixed content, brand preferences, asset references — is versioned in `site_specs` and available to every downstream agent. The classifier reads it and adjusts its output accordingly. The planner reads it and respects fixed content. The content writer reads it and uses provided copy instead of generating.

No new Go code. The step uses `write_site_spec` which already exists. The richness of the submission grows over time as we add more fields — the pipeline just reads whatever's there.

