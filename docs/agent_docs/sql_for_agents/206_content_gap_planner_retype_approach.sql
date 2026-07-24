-- 206_content_gap_planner_retype_approach.sql
-- ----------------------------------------------------------------------------
-- bugs_open/015 candidate 1 (LLM half): give content-gap-planner the
-- retype_existing approach its own advice already refers to.
--
-- Since v1.0.1144 MissingNewsPageCheck tells this planner to "RE-TYPE that
-- page" when a stranded page already occupies the news role — but the prompt
-- offered only add_to_page | new_page | update_spec | not_actionable, and
-- apply_gap_plan had no branch for anything else, so the advice could only
-- dead-end (unknown approach -> applied=false) or be mis-mapped onto
-- add_to_page (page builds but stays mistyped) or new_page (duplicate page —
-- the exact outcome the advice warns against).
--
-- *** DO NOT APPLY until a chassis image containing applyRetypeExisting is
-- LIVE. *** On the old binary a retype_existing plan hits the
-- unknown-approach branch and the gap item silently stops progressing.
-- Check the running pod first:
--   kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
--     'strings /app/agent-chassis | grep -c applyRetypeExisting'
--   (>0 means the new binary; also grep a known-old symbol as positive control)
--
-- Mechanics: whole-template jsonb_set (the 071 pattern), based on the LIVE
-- template captured 2026-07-24 (md5 c86623bec9455d745e9e6e03119d6ba5). The
-- UPDATE's WHERE asserts that md5, so if another thread has since changed
-- the prompt this is a 0-row no-op instead of a clobber — on 0 rows, STOP,
-- re-read the live template and re-derive this file's $PROMPT$ from it.
-- ----------------------------------------------------------------------------

BEGIN;

-- Backup current prompt (071's backup table).
CREATE TABLE IF NOT EXISTS agent_definition_prompt_backups (
    backed_up_at timestamptz DEFAULT now(),
    type         text,
    step         text,
    prompt       text
);
INSERT INTO agent_definition_prompt_backups (type, step, prompt)
SELECT 'content-gap-planner', 'plan_gaps',
       default_config #>> '{workflow,steps,plan_gaps,config,prompt_template}'
FROM agent_definitions
WHERE type = 'content-gap-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Apply: the live template plus approach E and its JSON schema entry.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_gaps,config,prompt_template}',
        to_jsonb($PROMPT$You are a site architect planning how to fill content gaps.

## Site
Domain: {{.site_record.domain}}
{{if .site_specs.specs.identity.company_name}}Company: {{.site_specs.specs.identity.company_name}}{{end}}
{{if .site_specs.specs.identity.industry}}Industry: {{.site_specs.specs.identity.industry}}{{end}}
{{if .site_specs.specs.identity.target_audience}}Audience: {{.site_specs.specs.identity.target_audience}}{{end}}

{{if .site_specs.specs.content_direction}}## Content Direction
Voice: {{.site_specs.specs.content_direction.voice}}
Emphasis: {{.site_specs.specs.content_direction.emphasis}}{{end}}

## Existing Pages
{{if .existing_pages.pages}}{{range .existing_pages.pages}}- {{.name}} ({{.page_type}}): {{.title}}
{{end}}{{else}}No pages loaded.{{end}}

## Available Section Components
{{if .available_components}}{{range .available_components}}- {{.name}} ({{.function}})
{{end}}{{end}}

## Content Gap to Address
Description: {{.input_data.spec.description}}
{{if .input_data.spec.suggestion}}Audit suggestion: {{.input_data.spec.suggestion}}{{end}}
Original category: {{.input_data.spec.category}}

## Your Task
Decide how to address this content gap. Choose ONE of these approaches:

### A) Add to existing page
If the gap can be filled by adding a section to an existing page (e.g. add FAQ section to the services page), recommend this. Only use section components from the Available list above.

### B) Create new page
If the gap needs its own page (e.g. a dedicated FAQ page, a pricing page), recommend creating one. Specify the page name, title, purpose, and sections from the Available list.

### C) Update site spec
If the gap is about missing metadata (target audience, tone definition), recommend a spec update.

### D) Not actionable
If the gap is too vague, already covered, or not worth addressing, say so.

### E) Re-type existing page
ONLY available when the gap description lists stranded candidate pages (nav-linked pages with no sections that can never build). If one of those candidates is clearly the page the gap asks for — created under the wrong page_type — choose this instead of creating a duplicate. page_name MUST be one of the listed candidate names verbatim; the executor refuses any other page, and the target page_type comes from the work item itself, not from you. Give the sections the re-typed page needs, from the Available list.

### Section selection rules
- Structured components such as `faq` and `pricing` are COMPLETE content surfaces — they hold their own content. Do NOT also add a `generic-text-block` covering the same material on the same page. Use a `generic-text-block` only for narrative content that a structured component does not already cover (e.g. a short intro clearly distinct from the FAQ items themselves).
- Use ONLY component function names from the Available Section Components list. Use the function name (the value in parentheses), never a display name or title.

Return ONLY valid JSON:
{
  "approach": "add_to_page" | "new_page" | "update_spec" | "not_actionable" | "retype_existing",
  "reasoning": "Why this approach",
  "add_to_page": {
    "page_name": "existing page name",
    "add_sections": ["faq", "call-to-action"],
    "content_guidance": "What the new sections should cover"
  },
  "new_page": {
    "name": "kebab-case-name",
    "title": "Page Title | Company Name",
    "page_type": "content",
    "purpose": "What this page covers and why",
    "sections": ["hero", "<choose content sections by purpose>", "call-to-action"],
    "nav_label": "Nav Label",
    "in_header": true,
    "in_footer": true
  },
  "update_spec": {
    "aspect": "identity",
    "field": "target_audience",
    "suggested_value": "UK SMEs looking for practical AI solutions"
  },
  "not_actionable": {
    "reason": "Why this gap doesnt need addressing"
  },
  "retype_existing": {
    "page_name": "name of one listed candidate page, verbatim",
    "sections": ["hero", "<sections the re-typed page needs>", "call-to-action"]
  }
}$PROMPT$::text),
        false
    ),
    updated_at = NOW()
WHERE type = 'content-gap-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND md5(default_config #>> '{workflow,steps,plan_gaps,config,prompt_template}')
      = 'c86623bec9455d745e9e6e03119d6ba5';

-- Verify (both = t; 0 rows means the md5 gate refused — see header).
SELECT
  ((default_config #>> '{workflow,steps,plan_gaps,config,prompt_template}') LIKE '%### E) Re-type existing page%') AS has_approach_e,
  ((default_config #>> '{workflow,steps,plan_gaps,config,prompt_template}') LIKE '%"retype_existing"%') AS has_schema_entry
FROM agent_definitions
WHERE type = 'content-gap-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

COMMIT;

-- ----------------------------------------------------------------------------
-- ROLLBACK (if needed): restore the previous prompt from the backup table.
--   UPDATE agent_definitions
--   SET default_config = jsonb_set(default_config,
--         '{workflow,steps,plan_gaps,config,prompt_template}',
--         to_jsonb((SELECT prompt FROM agent_definition_prompt_backups
--                   WHERE type='content-gap-planner' AND step='plan_gaps'
--                   ORDER BY backed_up_at DESC LIMIT 1)), false),
--       updated_at = NOW()
--   WHERE type = 'content-gap-planner' AND is_active
--     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
-- ----------------------------------------------------------------------------
