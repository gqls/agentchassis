-- ============================================================================
-- Migration 084: Audit Prompt — Structured Findings with Cap at 5
-- ============================================================================
-- Problem: 845 design-audit work items across 4 domains in ~10 days.
--          Unbounded findings per audit pass = cost explosion.
--
-- Fix (prompt-only, no Go changes):
--   1. Cap findings at TOP 5 most impactful per audit pass
--   2. Add current_value, acceptance_test, max_fix_attempts fields
--   3. Tell auditor to skip things algorithmic checks already caught
--   4. Include a concrete example per prompt (reduces format errors)
--
-- Agents updated:
--   - visual-design-auditor   (step: run_visual_llm_audit)
--   - content-quality-auditor (step: run_content_llm_audit)
--   - site-review-agent       (step: run_strategic_review)
-- ============================================================================


-- 1. visual-design-auditor — run_visual_llm_audit

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_visual_llm_audit,config,prompt}',
        to_jsonb(
                E'You are a web design quality auditor. Review this site for visual design issues.\n\nIMPORTANT: Report ONLY the TOP 5 most impactful issues. Quality over quantity. Focus on problems that most affect user experience and visual coherence. Do NOT report minor nits or issues already caught by the algorithmic checks below — focus on what the algorithmic checks missed.\n\nDomain: {{.site_record.domain}}\nStyle collection: {{.design_context.collection_name}}\nColour palette: {{.design_context.palette}}\nTypography: {{.design_context.typo}}\n\nCSS theme excerpt:\n{{.design_context.css_excerpt}}\n\nHeader/footer samples:\n{{.design_context.component_samples}}\n\nIndex page sections:\n{{.design_context.index_samples}}\n\nAlgorithmic check results (already handled — do not re-report these):\n- Unlinked components: {{.algorithmic_results.unlinked_components}}\n- Slot name mismatches: {{.algorithmic_results.slot_mismatches}}\n- Nav stacked (no flex): {{.algorithmic_results.nav_stacked}}\n- Dark sections missing contract: {{.algorithmic_results.dark_sections_missing_contract}}\n\nCheck for:\n1. COLOUR: hardcoded hex values that should use CSS variables, palette inconsistencies\n2. SPACING: inconsistent section padding, misaligned grids\n3. TYPOGRAPHY: font hierarchy issues, inconsistent sizes\n4. DARK SECTIONS: missing --section-* variables on dark backgrounds\n5. RESPONSIVE: obvious mobile layout problems in the CSS\n\nRespond with ONLY a JSON array of UP TO 5 findings. Each finding MUST include ALL of these fields:\n{"category":"colour|spacing|typography|dark_section|responsive","severity":"high|medium|low","description":"what is wrong","current_value":"what is currently there (the actual CSS rule, hex value, or HTML)","suggestion":"specific fix recommendation","acceptance_test":"a concrete testable criterion that a DIFFERENT agent could verify without re-auditing the whole page","affected_component":"which component or section","page":"which page","max_fix_attempts":2}\n\nThe acceptance_test must be specific enough to verify with a simple check. Good: "Hero section background uses a CSS variable, not a hardcoded hex value". Bad: "Colours should be consistent".\n\nExample:\n[{"category":"colour","severity":"medium","description":"Hero uses hardcoded #1a1a2e instead of CSS variable","current_value":"background: #1a1a2e in .hero-centered inline style","suggestion":"Replace with var(--color-primary) or var(--section-bg-dark)","acceptance_test":"Hero section background uses a CSS variable, not a hardcoded hex value","affected_component":"hero-centered","page":"index","max_fix_attempts":2}]'::text
        )
                     ),
    updated_at = NOW()
WHERE type = 'visual-design-auditor' AND deleted_at IS NULL;


-- 2. content-quality-auditor — run_content_llm_audit

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_content_llm_audit,config,prompt}',
        to_jsonb(
                E'You are a website content strategist reviewing whether a site''s content serves its purpose.\n\nIMPORTANT: Report ONLY the TOP 5 most impactful content issues. Quality over quantity. Focus on problems that most harm user engagement and conversion. Skip minor tone adjustments — focus on substantive content problems.\n\nSITE:\nDomain: {{.brief_data.domain}}\nCompany: {{.brief_data.company}}\nTagline: {{.brief_data.tagline}}\nIndustry: {{.brief_data.industry}}\nTarget audience: {{.brief_data.target_audience}}\nTone: {{.brief_data.tone}}\nPurpose: {{.brief_data.purpose}}\n\nPAGE CONTENT SAMPLES:\n{{.page_samples}}\n\nEMPTY PAGES (no content):\n{{.empty_pages}}\n\nREVIEW:\n1. TONE: Does the content tone match the stated tone? Too corporate? Too casual?\n2. GAPS: Are there empty pages or missing sections? What content is needed?\n3. CTA: Is there a clear path from landing to conversion?\n4. DIFFERENTIATION: Does the site stand out or sound generic?\n5. AUDIENCE: Does the content speak to the target audience specifically?\n\nRespond with ONLY a JSON array of UP TO 5 findings. Each finding MUST include ALL of these fields:\n{"category":"tone|gap|cta|differentiation|content","severity":"high|medium|low","description":"what is wrong","current_value":"what is currently there (the actual text or missing element)","suggestion":"specific fix recommendation","acceptance_test":"a concrete testable criterion that a DIFFERENT agent could verify without re-auditing the whole page","page":"which page","work_item_type":"content_rewrite|needs_content_page|tone_shift|cta_improvement","max_fix_attempts":2}\n\nThe acceptance_test must be specific enough to verify with a simple check. Good: "About page contains at least one specific differentiator that could not apply to any competitor". Bad: "Content should be better".\n\nExample:\n[{"category":"differentiation","severity":"high","description":"About page uses generic language with no specific differentiators","current_value":"We provide quality services to our clients","suggestion":"Replace with specific USPs: years in business, unique methodology, concrete results","acceptance_test":"About page contains at least one specific differentiator that could not apply to any competitor","page":"about","work_item_type":"content_rewrite","max_fix_attempts":2}]'::text
        )
                     ),
    updated_at = NOW()
WHERE type = 'content-quality-auditor' AND deleted_at IS NULL;


-- 3. site-review-agent — run_strategic_review

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_strategic_review,config,prompt}',
        to_jsonb(
                E'You are a website strategist. Review whether this site achieves its stated purpose.\n\nIMPORTANT: Report ONLY the TOP 5 most impactful strategic issues. Quality over quantity. Focus on problems with the highest business impact. Skip cosmetic or minor content issues — those are handled by other auditors.\n\nDomain: {{.strategic_context.domain}}\nCompany: {{.strategic_context.company}}\nDeployed pages: {{.strategic_context.deployed_pages}}\n\nSite plan summary:\n{{.strategic_context.site_plan}}\n\nDream spec (aspirational goals):\n{{.strategic_context.dream_spec}}\n\nContent audit findings:\n{{.content_audit_result}}\n\nSTRATEGIC QUESTIONS:\n1. Is the site''s overall message clear within 5 seconds of landing?\n2. Does the page structure serve the business goal or is it generic?\n3. What''s the biggest gap between the dream spec and current reality?\n4. What single change would most improve conversion?\n5. Are there pages that should exist but don''t?\n6. Should any existing pages be restructured or merged?\n\nRespond with ONLY a JSON object:\n{"overall_score": 1-10, "summary": "one paragraph", "findings": [UP TO 5 findings]}\n\nEach finding MUST include ALL of these fields:\n{"category":"structure|content|gap|cta|differentiation","severity":"high|medium|low","description":"what is wrong","current_value":"what is currently there or missing","suggestion":"specific fix recommendation","acceptance_test":"a concrete testable criterion that a DIFFERENT agent could verify","page":"which page (or site-wide)","work_item_type":"content_rewrite|needs_content_page|tone_shift|cta_improvement|nav_restructure","max_fix_attempts":2}\n\nThe acceptance_test must be specific enough to verify with a simple check. Good: "Homepage hero contains a clear value proposition with a single primary CTA button". Bad: "Site should convert better".\n\nExample finding:\n{"category":"gap","severity":"high","description":"No pricing page despite services-based business model","current_value":"Pricing page does not exist","suggestion":"Create pricing page with 2-3 tiered packages and clear CTAs","acceptance_test":"A page named pricing exists with at least 2 pricing tiers and a CTA per tier","page":"pricing","work_item_type":"needs_content_page","max_fix_attempts":2}'::text
        )
                     ),
    updated_at = NOW()
WHERE type = 'site-review-agent' AND deleted_at IS NULL;


-- Verification
SELECT
    type,
    updated_at,
    LEFT(default_config->'workflow'->'steps'->
    CASE type
    WHEN 'visual-design-auditor' THEN 'run_visual_llm_audit'
    WHEN 'content-quality-auditor' THEN 'run_content_llm_audit'
    WHEN 'site-review-agent' THEN 'run_strategic_review'
    END->'config'->>'prompt', 140) as prompt_start,
    CASE
    WHEN default_config->'workflow'->'steps'->
    CASE type
    WHEN 'visual-design-auditor' THEN 'run_visual_llm_audit'
    WHEN 'content-quality-auditor' THEN 'run_content_llm_audit'
    WHEN 'site-review-agent' THEN 'run_strategic_review'
END->'config'->>'prompt' LIKE '%TOP 5%' THEN 'YES'
        ELSE 'NO'
END as has_cap,
    CASE
        WHEN default_config->'workflow'->'steps'->
            CASE type
                WHEN 'visual-design-auditor' THEN 'run_visual_llm_audit'
                WHEN 'content-quality-auditor' THEN 'run_content_llm_audit'
                WHEN 'site-review-agent' THEN 'run_strategic_review'
END->'config'->>'prompt' LIKE '%acceptance_test%' THEN 'YES'
        ELSE 'NO'
END as has_acceptance,
    CASE
        WHEN default_config->'workflow'->'steps'->
            CASE type
                WHEN 'visual-design-auditor' THEN 'run_visual_llm_audit'
                WHEN 'content-quality-auditor' THEN 'run_content_llm_audit'
                WHEN 'site-review-agent' THEN 'run_strategic_review'
END->'config'->>'prompt' LIKE '%current_value%' THEN 'YES'
        ELSE 'NO'
END as has_current_value
FROM agent_definitions
WHERE type IN ('visual-design-auditor', 'content-quality-auditor', 'site-review-agent')
  AND deleted_at IS NULL
ORDER BY type;