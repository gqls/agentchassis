Here's how the page-content-writer currently works

page-build-handler receives work item (e.g. needs_content_page)
→ ensure_site_record (loads site from DB)
→ load_page_record (gets page with sections list)
→ plan_sections (checks data readiness per section)
→ spawn page-content-writer
→ call page-content-writer with input_mapping:
site_record, current_page, section_plan,
reviewed_brief (from content_data)

Inside page-content-writer:
1. load_site_specs → reads all site_specs (identity, content_direction, etc.)
2. prepare_link_context → builds internal linking data
3. load_page_components → loads component DEFINITIONS from content_components
   (template, input_schema, category, description)
4. build_render_context → assembles company name, tone, email, phone etc.
5. process_sections_loop → for each section:
   a. check_render_mode → needs LLM or just template?
   b. if LLM: generate_content → execute_llm_prompt with big prompt
   c. render_section → renders LLM output into component template

The generate_content step's prompt includes:

Company context (from render_context)
Contact info
Content direction (from site_specs)
Section requirements: component name, function, purpose (from the component DEFINITION)
Data schema required
Research findings (if any)

--


What's missing: The prompt has no awareness of page_components.content_brief. The section requirements come from the component definition (what a "hero" section generally is), not from any admin-edited instructions for this specific instance.

new flow:

Admin edits brief in dashboard → saved to page_components.content_brief
Admin clicks "Regenerate" → content_rewrite work item created
↓
page-build-handler picks up content_rewrite
→ spawns page-content-writer
→ page-content-writer runs:
1. load_page_section_components → loads definitions + content_briefs
2. process_sections_loop → for each section:
- current_section.content_brief is present (from the DB query)
- generate_content prompt includes "## Admin Content Brief" block
- LLM follows the brief's purpose/tone/guidance
3. Sections without briefs → prompt is identical to before