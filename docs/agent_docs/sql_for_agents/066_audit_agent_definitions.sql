-- ============================================================================
-- Audit Agent Hierarchy:
--
-- design-audit-agent (top-level)
--   → visual-design-auditor (group agent)
--       Uses: check actions for colour, spacing, typography, dark sections
--       Plus: LLM call for holistic visual assessment
--
--   → content-quality-auditor (group agent)
--       Used by site-review-agent too
--       Uses: check actions for tone, gaps, CTA
--       Plus: LLM call for content assessment
--
-- site-review-agent (top-level)
--   → content-quality-auditor (reused)
--   → strategic alignment LLM call
--
-- Pattern: group agents start with algorithmic checks (Go actions),
-- then make ONE LLM call for subjective assessment, then write findings.
-- Each can grow independently — an algorithmic check can be promoted
-- to its own agent when it needs to call vision AI or research agents.
--
-- Pre-flight (Step 0):
--   Existing: ensure_site_record, query_database, execute_llm_prompt,
--             spawn_agent, call_agent, complete_workflow
--   New: write_audit_findings (Go action, provided separately)
--   No other new Go code for the agent definitions.
-- ============================================================================


-- visual-design-auditor
-- content-quality-auditor
-- design-audit-agent
-- site-review-agent
-- page-build-handler

-- ============================================================================
-- 1. visual-design-auditor
-- Group agent: loads design context, runs algorithmic colour/spacing checks
-- via a single query_database step, then does one LLM call for holistic
-- visual assessment. Writes findings via write_audit_findings.
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, agent_category, status,
    image_repository, image_tag, category,
    default_config, input_contract, output_contract, domain_tags
) VALUES (
             'visual-design-auditor',
             'Visual Design Auditor',
             'Group auditor for visual design quality. Loads style collection and rendered HTML samples, runs algorithmic checks for colour consistency and spacing, then makes one LLM call for holistic visual assessment. Produces findings as work items.',
             'analyst', 'active',
             'docker.io/aqls/agent-chassis', 'v1.0.831', 'analyst',
             '{"workflow":{"start_step":"ensure_site_record","processing_mode":"orchestrator","timeout_seconds":300,"steps":{
                 "ensure_site_record":{
                     "action":"ensure_site_record",
                     "config":{"store_brief_in_content_data":false},
                     "next_step":"load_design_context",
                     "output_field":"site_record"
                 },
                 "load_design_context":{
                     "action":"query_database",
                     "config":{
                         "query":"SELECT sc.name as collection_name, sc.color_palette::text as palette, sc.typography::text as typo, LEFT(ct.css_content, 2000) as css_excerpt, (SELECT string_agg(slot_name || '':'' || LEFT(rendered_html, 800), ''|||'') FROM site_components WHERE site_id = s.id) as component_samples, (SELECT string_agg(LEFT(pc.rendered_html, 600), ''|||'') FROM page_components pc JOIN pages p ON pc.page_id = p.id WHERE p.site_id = s.id AND p.name = ''index'' AND pc.rendered_html IS NOT NULL LIMIT 5) as index_samples FROM sites s LEFT JOIN style_collections sc ON s.style_collection_id = sc.id LEFT JOIN css_themes ct ON sc.css_theme_id = ct.id WHERE s.id = $1",
                         "params":["site_record.site_id"],
                         "output_format":"object"
                     },
                     "next_step":"run_algorithmic_checks",
                     "output_field":"design_context"
                 },
                 "run_algorithmic_checks":{
                     "action":"query_database",
                     "config":{
                         "query":"SELECT (SELECT COUNT(*) FROM site_components WHERE site_id = $1 AND component_id IS NULL AND slot_name IN (''header'',''footer'',''head'')) as unlinked_components, (SELECT COUNT(*) FROM page_components pc JOIN pages p ON pc.page_id = p.id WHERE p.site_id = $1 AND pc.rendered_html LIKE ''%data-component=%'' AND pc.slot_name IS NOT NULL AND pc.slot_name != substring(pc.rendered_html from ''data-component=\"([^\"]*)\"'')) as slot_mismatches, (SELECT CASE WHEN rendered_html NOT LIKE ''%display: flex%'' AND rendered_html NOT LIKE ''%display:flex%'' AND rendered_html LIKE ''%<ul%'' THEN 1 ELSE 0 END FROM site_components WHERE site_id = $1 AND slot_name = ''header'') as nav_stacked, (SELECT COUNT(*) FROM page_components pc JOIN pages p ON pc.page_id = p.id WHERE p.site_id = $1 AND pc.rendered_html LIKE ''%background:%'' AND pc.rendered_html NOT LIKE ''%--section-text%'' AND pc.rendered_html LIKE ''%color: #fff%'') as dark_sections_missing_contract",
                         "params":["site_record.site_id"],
                         "output_format":"object"
                     },
                     "next_step":"run_visual_llm_audit",
                     "output_field":"algorithmic_results"
                 },
                 "run_visual_llm_audit":{
                     "action":"execute_llm_prompt",
                     "config":{
                         "prompt":"You are a web design quality auditor. Review this site for visual design issues.\n\nDomain: {{.site_record.domain}}\nStyle collection: {{.design_context.collection_name}}\nColour palette: {{.design_context.palette}}\nTypography: {{.design_context.typo}}\n\nCSS theme excerpt:\n{{.design_context.css_excerpt}}\n\nHeader/footer samples:\n{{.design_context.component_samples}}\n\nIndex page sections:\n{{.design_context.index_samples}}\n\nAlgorithmic check results:\n- Unlinked components: {{.algorithmic_results.unlinked_components}}\n- Slot name mismatches: {{.algorithmic_results.slot_mismatches}}\n- Nav stacked (no flex): {{.algorithmic_results.nav_stacked}}\n- Dark sections missing contract: {{.algorithmic_results.dark_sections_missing_contract}}\n\nCheck for:\n1. COLOUR: hardcoded hex values that should use CSS variables, palette inconsistencies\n2. SPACING: inconsistent section padding, misaligned grids\n3. TYPOGRAPHY: font hierarchy issues, inconsistent sizes\n4. DARK SECTIONS: missing --section-* variables on dark backgrounds\n5. RESPONSIVE: obvious mobile layout problems in the CSS\n\nRespond with ONLY a JSON array of findings. Each: {\"category\":\"colour|spacing|typography|dark_section|responsive\",\"severity\":\"high|medium|low\",\"description\":\"...\",\"suggestion\":\"...\",\"affected_component\":\"...\",\"page\":\"...\"}",
                         "input_fields":["site_record","design_context","algorithmic_results"],
                         "ai_service":{"provider":"anthropic","model":"claude-sonnet-4-5-20250514","max_tokens":4000}
                     },
                     "next_step":"write_findings",
                     "output_field":"visual_audit"
                 },
                 "write_findings":{
                     "action":"write_audit_findings",
                     "config":{
                         "findings_field":"visual_audit.result",
                         "site_id":"site_record.site_id",
                         "audit_source":"visual-design-audit"
                     },
                     "next_step":"complete",
                     "output_field":"findings_written"
                 },
                 "complete":{
                     "action":"complete_workflow",
                     "config":{"output_fields":["algorithmic_results","visual_audit","findings_written"]}
                 }
             }}}'::jsonb,
             '{"required": ["site_id", "domain"]}'::jsonb,
             '{"produces": {"findings_written": "count of work items created", "algorithmic_results": "structural check counts"}}'::jsonb,
             '["audit", "design", "visual"]'::jsonb
         ) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description,
    image_tag = EXCLUDED.image_tag,
    updated_at = NOW();


-- ============================================================================
-- 2. content-quality-auditor
-- Group agent: loads brief and page content, runs one LLM call for
-- tone, content gaps, CTA effectiveness, differentiation.
-- Reusable by both design-audit-agent and site-review-agent.
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, agent_category, status,
    image_repository, image_tag, category,
    default_config, input_contract, output_contract, domain_tags
) VALUES (
             'content-quality-auditor',
             'Content Quality Auditor',
             'Group auditor for content quality. Loads the site brief, page content samples, and target audience. Makes one LLM call to assess tone alignment, content gaps, CTA effectiveness, and differentiation. Produces findings as work items.',
             'analyst', 'active',
             'docker.io/aqls/agent-chassis', 'v1.0.831', 'analyst',
             '{"workflow":{"start_step":"ensure_site_record","processing_mode":"orchestrator","timeout_seconds":300,"steps":{
                 "ensure_site_record":{
                     "action":"ensure_site_record",
                     "config":{"store_brief_in_content_data":false},
                     "next_step":"load_brief",
                     "output_field":"site_record"
                 },
                 "load_brief":{
                     "action":"query_database",
                     "config":{
                         "query":"SELECT s.domain, COALESCE(s.company_name, s.domain) as company, COALESCE(s.tagline, '''') as tagline, COALESCE(s.content_data->>''industry'', '''') as industry, COALESCE(ss.spec_data->>''target_audience'', '''') as target_audience, COALESCE(ss.spec_data->>''tone'', '''') as tone, COALESCE(ss.spec_data->>''purpose'', '''') as purpose, COALESCE(ss.spec_data->>''key_messages'', '''') as key_messages FROM sites s LEFT JOIN site_specs ss ON ss.site_id = s.id AND ss.spec_type = ''site_plan'' WHERE s.id = $1 ORDER BY ss.created_at DESC LIMIT 1",
                         "params":["site_record.site_id"],
                         "output_format":"object"
                     },
                     "next_step":"load_page_content",
                     "output_field":"brief_data"
                 },
                 "load_page_content":{
                     "action":"query_database",
                     "config":{
                         "query":"SELECT p.name, LEFT(string_agg(pc.rendered_html, '' ''), 1000) as content_sample FROM pages p JOIN page_components pc ON pc.page_id = p.id WHERE p.site_id = $1 AND p.name IN (''index'', ''about'', ''services'', ''contact'') AND pc.rendered_html IS NOT NULL GROUP BY p.name ORDER BY p.name",
                         "params":["site_record.site_id"],
                         "output_format":"rows"
                     },
                     "next_step":"check_empty_pages",
                     "output_field":"page_samples"
                 },
                 "check_empty_pages":{
                     "action":"query_database",
                     "config":{
                         "query":"SELECT p.name FROM pages p LEFT JOIN page_components pc ON pc.page_id = p.id AND pc.rendered_html IS NOT NULL AND pc.rendered_html != '''' WHERE p.site_id = $1 AND p.build_status IN (''deployed'', ''active'') GROUP BY p.name HAVING COUNT(pc.id) = 0",
                         "params":["site_record.site_id"],
                         "output_format":"rows"
                     },
                     "next_step":"run_content_llm_audit",
                     "output_field":"empty_pages"
                 },
                 "run_content_llm_audit":{
                     "action":"execute_llm_prompt",
                     "config":{
                         "prompt":"You are a website content strategist reviewing whether a site''s content serves its purpose.\n\nSITE:\nDomain: {{.brief_data.domain}}\nCompany: {{.brief_data.company}}\nTagline: {{.brief_data.tagline}}\nIndustry: {{.brief_data.industry}}\nTarget audience: {{.brief_data.target_audience}}\nTone: {{.brief_data.tone}}\nPurpose: {{.brief_data.purpose}}\n\nPAGE CONTENT SAMPLES:\n{{.page_samples}}\n\nEMPTY PAGES (no content):\n{{.empty_pages}}\n\nREVIEW:\n1. TONE: Does the content tone match the stated tone? Too corporate? Too casual?\n2. GAPS: Are there empty pages or missing sections? What content is needed?\n3. CTA: Is there a clear path from landing to conversion?\n4. DIFFERENTIATION: Does the site stand out or sound generic?\n5. AUDIENCE: Does the content speak to the target audience specifically?\n\nRespond with ONLY a JSON array of findings. Each: {\"category\":\"tone|gap|cta|differentiation|content\",\"severity\":\"high|medium|low\",\"description\":\"...\",\"suggestion\":\"...\",\"page\":\"...\",\"work_item_type\":\"content_rewrite|needs_content_page|tone_shift|cta_improvement\"}",
                         "input_fields":["brief_data","page_samples","empty_pages"],
                         "ai_service":{"provider":"anthropic","model":"claude-sonnet-4-5-20250514","max_tokens":4000}
                     },
                     "next_step":"write_findings",
                     "output_field":"content_audit"
                 },
                 "write_findings":{
                     "action":"write_audit_findings",
                     "config":{
                         "findings_field":"content_audit.result",
                         "site_id":"site_record.site_id",
                         "audit_source":"content-quality-audit"
                     },
                     "next_step":"complete",
                     "output_field":"findings_written"
                 },
                 "complete":{
                     "action":"complete_workflow",
                     "config":{"output_fields":["content_audit","findings_written","empty_pages"]}
                 }
             }}}'::jsonb,
             '{"required": ["site_id", "domain"]}'::jsonb,
             '{"produces": {"findings_written": "count of work items created", "empty_pages": "list of pages with no content"}}'::jsonb,
             '["audit", "content", "review"]'::jsonb
         ) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description,
    image_tag = EXCLUDED.image_tag,
    updated_at = NOW();


-- ============================================================================
-- 3. design-audit-agent (top-level orchestrator)
-- Spawns and calls visual-design-auditor and content-quality-auditor.
-- Then triages detected items.
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, agent_category, status,
    image_repository, image_tag, category,
    default_config, input_contract, output_contract, domain_tags
) VALUES (
             'design-audit-agent',
             'Design Audit Agent',
             'Top-level design audit orchestrator. Spawns visual-design-auditor for CSS/layout/colour checks and content-quality-auditor for tone/gaps/CTA checks. Aggregates findings and triages work items.',
             'analyst', 'active',
             'docker.io/aqls/agent-chassis', 'v1.0.831', 'analyst',
             '{"workflow":{"start_step":"ensure_site_record","processing_mode":"orchestrator","timeout_seconds":600,"steps":{
                 "ensure_site_record":{
                     "action":"ensure_site_record",
                     "config":{"store_brief_in_content_data":false},
                     "next_step":"spawn_visual_auditor",
                     "output_field":"site_record"
                 },
                 "spawn_visual_auditor":{
                     "action":"spawn_agent",
                     "config":{"role":"visual_auditor","agent_type":"visual-design-auditor"},
                     "next_step":"call_visual_auditor",
                     "output_field":"visual_auditor_agent"
                 },
                 "call_visual_auditor":{
                     "action":"call_agent",
                     "config":{
                         "target_role":"visual_auditor",
                         "input_mapping":{"site_id":"site_record.site_id","domain":"site_record.domain"},
                         "timeout_seconds":300
                     },
                     "next_step":"spawn_content_auditor",
                     "error_step":"spawn_content_auditor",
                     "output_field":"visual_audit_result"
                 },
                 "spawn_content_auditor":{
                     "action":"spawn_agent",
                     "config":{"role":"content_auditor","agent_type":"content-quality-auditor"},
                     "next_step":"call_content_auditor",
                     "output_field":"content_auditor_agent"
                 },
                 "call_content_auditor":{
                     "action":"call_agent",
                     "config":{
                         "target_role":"content_auditor",
                         "input_mapping":{"site_id":"site_record.site_id","domain":"site_record.domain"},
                         "timeout_seconds":300
                     },
                     "next_step":"triage",
                     "error_step":"triage",
                     "output_field":"content_audit_result"
                 },
                 "triage":{
                     "action":"triage_detected_items",
                     "config":{"site_id":"site_record.site_id","target_domain":"build"},
                     "next_step":"complete",
                     "output_field":"triage_result"
                 },
                 "complete":{
                     "action":"complete_workflow",
                     "config":{"output_fields":["visual_audit_result","content_audit_result","triage_result"]}
                 }
             }}}'::jsonb,
             '{"required": ["site_id", "domain"]}'::jsonb,
             '{"produces": {"visual_audit_result": "visual design findings", "content_audit_result": "content quality findings", "triage_result": "items promoted to triaged"}}'::jsonb,
             '["audit", "design", "orchestrator"]'::jsonb
         ) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description,
    image_tag = EXCLUDED.image_tag,
    updated_at = NOW();


-- ============================================================================
-- 4. site-review-agent (top-level strategic review)
-- Loads brief/dream_spec, calls content-quality-auditor, then does its own
-- strategic alignment LLM call for higher-level questions.
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, agent_category, status,
    image_repository, image_tag, category,
    default_config, input_contract, output_contract, domain_tags
) VALUES (
             'site-review-agent',
             'Site Review Agent',
             'Strategic site review. Compares current site against original brief and dream spec. Asks: is the site achieving its purpose? Calls content-quality-auditor for content assessment, then runs its own strategic alignment review. Produces work items for content rewrites, new pages, and tone shifts.',
             'analyst', 'active',
             'docker.io/aqls/agent-chassis', 'v1.0.831', 'analyst',
             '{"workflow":{"start_step":"ensure_site_record","processing_mode":"orchestrator","timeout_seconds":600,"steps":{
                 "ensure_site_record":{
                     "action":"ensure_site_record",
                     "config":{"store_brief_in_content_data":false},
                     "next_step":"load_strategic_context",
                     "output_field":"site_record"
                 },
                 "load_strategic_context":{
                     "action":"query_database",
                     "config":{
                         "query":"SELECT s.domain, COALESCE(s.company_name, s.domain) as company, s.content_data->>''dream_spec'' as dream_spec, COALESCE(ss.spec_data::text, ''{}''::text) as site_plan, (SELECT COUNT(*) FROM pages WHERE site_id = s.id AND build_status = ''deployed'') as deployed_pages, (SELECT COUNT(*) FROM site_work_items WHERE site_id = s.id AND status = ''complete'') as completed_items FROM sites s LEFT JOIN site_specs ss ON ss.site_id = s.id AND ss.spec_type = ''site_plan'' WHERE s.id = $1 ORDER BY ss.created_at DESC LIMIT 1",
                         "params":["site_record.site_id"],
                         "output_format":"object"
                     },
                     "next_step":"spawn_content_auditor",
                     "output_field":"strategic_context"
                 },
                 "spawn_content_auditor":{
                     "action":"spawn_agent",
                     "config":{"role":"content_auditor","agent_type":"content-quality-auditor"},
                     "next_step":"call_content_auditor",
                     "output_field":"content_auditor_agent"
                 },
                 "call_content_auditor":{
                     "action":"call_agent",
                     "config":{
                         "target_role":"content_auditor",
                         "input_mapping":{"site_id":"site_record.site_id","domain":"site_record.domain"},
                         "timeout_seconds":300
                     },
                     "next_step":"run_strategic_review",
                     "error_step":"run_strategic_review",
                     "output_field":"content_audit_result"
                 },
                 "run_strategic_review":{
                     "action":"execute_llm_prompt",
                     "config":{
                         "prompt":"You are a website strategist. Review whether this site achieves its stated purpose.\n\nDomain: {{.strategic_context.domain}}\nCompany: {{.strategic_context.company}}\nDeployed pages: {{.strategic_context.deployed_pages}}\n\nSite plan summary:\n{{.strategic_context.site_plan}}\n\nDream spec (aspirational goals):\n{{.strategic_context.dream_spec}}\n\nContent audit findings:\n{{.content_audit_result}}\n\nSTRATEGIC QUESTIONS:\n1. Is the site''s overall message clear within 5 seconds of landing?\n2. Does the page structure serve the business goal or is it generic?\n3. What''s the biggest gap between the dream spec and current reality?\n4. What single change would most improve conversion?\n5. Are there pages that should exist but don''t?\n6. Should any existing pages be restructured or merged?\n\nRespond with ONLY a JSON object: {\"overall_score\": 1-10, \"summary\": \"one paragraph\", \"findings\": [{\"category\":\"structure|content|gap|cta|differentiation\",\"severity\":\"high|medium|low\",\"description\":\"...\",\"suggestion\":\"...\",\"page\":\"...\",\"work_item_type\":\"content_rewrite|needs_content_page|tone_shift|cta_improvement|nav_restructure\"}]}",
                         "input_fields":["strategic_context","content_audit_result"],
                         "ai_service":{"provider":"anthropic","model":"claude-sonnet-4-5-20250514","max_tokens":4000}
                     },
                     "next_step":"write_strategic_findings",
                     "output_field":"strategic_review"
                 },
                 "write_strategic_findings":{
                     "action":"write_audit_findings",
                     "config":{
                         "findings_field":"strategic_review.result",
                         "site_id":"site_record.site_id",
                         "audit_source":"site-review"
                     },
                     "next_step":"triage",
                     "output_field":"strategic_findings_written"
                 },
                 "triage":{
                     "action":"triage_detected_items",
                     "config":{"site_id":"site_record.site_id","target_domain":"build"},
                     "next_step":"complete",
                     "output_field":"triage_result"
                 },
                 "complete":{
                     "action":"complete_workflow",
                     "config":{"output_fields":["content_audit_result","strategic_review","strategic_findings_written","triage_result"]}
                 }
             }}}'::jsonb,
             '{"required": ["site_id", "domain"]}'::jsonb,
             '{"produces": {"strategic_review": "overall score and findings", "triage_result": "items promoted to triaged"}}'::jsonb,
             '["review", "strategy", "orchestrator"]'::jsonb
         ) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description,
    image_tag = EXCLUDED.image_tag,
    updated_at = NOW();


-- ============================================================================
-- Verify all agents created
-- ============================================================================
SELECT type, display_name, agent_category, status
FROM agent_definitions
WHERE type IN (
               'visual-design-auditor',
               'content-quality-auditor',
               'design-audit-agent',
               'site-review-agent',
               'page-build-handler'
    )
  AND deleted_at IS NULL
ORDER BY type;