https://claude.ai/chat/a36b6fe1-efa1-4d53-b30e-768ab6c9bf68

┌─────────────────┐
│  strategist     │ → Generates build plan (JSON)
└────────┬────────┘
↓
┌─────────────────┐
│   architect     │ → Generates template with {{placeholders}}
└────────┬────────┘
↓
┌─────────────────┐
│     writer      │ → Generates content (JSON)
└────────┬────────┘
↓
┌─────────────────┐
│ html_assembler  │ → ✅ Combines template + content = assembled HTML
└────────┬────────┘
↓
┌─────────────────┐
│ wrap_multipage  │ → ✅ Takes HTML, creates index/about/contact
└────────┬────────┘
↓
┌─────────────────┐
│    deployer     │ → Commits all 3 files to Git
└─────────────────┘



CollectedData at wrap_multipage:
{
"build_plan": {...},
"template_data": {
"assemble_template": {
"stitched_html_template": "<html>...{{brand_name}}...</html>",
"content_requirements": {...}
}
},
"content_data": {
"generate_content": {
"result": "{\"meta\":{...}, \"sections\":{...}}"
}
},
"final_html": {
"assemble_html": {
"final_html": "<html>...Acme Corp...</html>"  // ✅ Assembled!
}
}
}


Data Flow Details
Step 1: Strategist
Input: domain, objective, model
Output: build_plan

{
"model": "AIDA",
"sections": ["header", "hero", "features", "social_proof", ...],
"section_guidance": {...}
}

Step 2: Architect
Input: build_plan, input_data
Output: template_data

{
"assemble_template": {
"stitched_html_template": "<html><body>{{brand_name}}...</body></html>",
"content_requirements": {
"component_header_0": {
"brand_name": "string",
"cta_text": "string"
}
}
}
}

Step 3: Writer
Input: template_data, build_plan, input_data
Output: content_data

{
"generate_content": {
"result": "{
\"sections\": {
\"component_header_0\": {
\"brand_name\": \"Acme Corp\",
\"cta_text\": \"Get Started\"
}
}
}"
}
}

Step 4: HTML Assembler (NEW!)
Input: template_data, content_data, input_data
Process:

Takes stitched_html_template
Takes content_json from content_data
Replaces all {{placeholders}} with actual values
Returns fully assembled HTML

{
"assemble_html": {
"final_html": "<html><body><header>Acme Corp ... Get Started</header>...</body></html>"
}
}

Step 5: Wrap Multipage
Input: final_html, input_data
Process:

Extracts final_html.assemble_html.final_html → index.html
Generates about.html with brand info
Generates contact.html with email
Wraps all into files map

Output: site_files

{
"files": {
"index.html": "<html>...</html>",
"about.html": "<html>...</html>",
"contact.html": "<html>...</html>"
},
"file_count": 3,
"pages": ["index.html", "about.html", "contact.html"]
}

Step 6: Deployer
Input: site_files, input_data
Process:

Extracts domain from input_data
Extracts files from site_files.files
Commits to Git: domain/index.html, domain/about.html, domain/contact.html

Output: Deployment result

{
"status": "committed",
"repo": "sites",
"domain": "test-multipage.com",
"file_count": 3
}

Agent Definitions Required
The workflow requires these 5 agents to be spawned:

site-strategist - Generates build plan (AIDA model)
landing-page-architect - Assembles template from component library
content-writer - Generates content JSON via LLM
html-assembler - Combines template + content → HTML
site-deployer - Commits to Git repository
