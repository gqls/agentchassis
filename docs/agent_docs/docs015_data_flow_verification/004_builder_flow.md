intake-orchestrator:
fetch_available_builders → spawn_classifier → spawn_briefer
→ call_classifier → hitl_confirm_type → fetch_questionnaire
→ call_briefer → hitl_review_brief → spawn_builder → call_builder

site-classifier:
Single LLM call → outputs ONE site_type from [landing, content, portfolio, brochure]
→ ONE recommended_builder

site-planner:
Load components from DB → load style collections → LLM plans pages
→ each page gets name, title, sections[] (component names)

pageflow-builder:
ensure_site_record → call_site_planner → sync_pages → populate_nav
→ image generation → style collection → build_pages_loop
(loop: write_content → review → assemble → git_commit → save_sections)