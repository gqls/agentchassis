# Link Constraints Implementation Summary

## Problem Statement

Content writers (LLMs) hallucinate links to pages that don't exist, resulting in:
- Broken internal links (404 errors)
- Poor user experience
- SEO penalties
- Inconsistent navigation

Example: LLM writes `<a href="/services/ai-consulting.html">AI Consulting</a>` when no such page exists.

## Solution

A dedicated action in the page-content-writer workflow that prepares link context before content generation.

**Key Principle:** Single-purpose output. The content writer ONLY produces content. It does NOT suggest new pages or flag opportunities - that's handled by a separate maintenance process.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    page-content-writer workflow                  │
│                                                                 │
│  spawn_research_agent                                           │
│         │                                                       │
│         ▼                                                       │
│  ┌─────────────────────────┐                                    │
│  │ prepare_link_context    │  ◄── NEW ACTION                    │
│  │                         │                                    │
│  │ Input: db_sync.pages    │                                    │
│  │ Output: link_context    │                                    │
│  │   • available_pages     │                                    │
│  │   • link_constraint_text│                                    │
│  └─────────────────────────┘                                    │
│         │                                                       │
│         ▼                                                       │
│  load_page_components                                           │
│         │                                                       │
│         ▼                                                       │
│  build_render_context                                           │
│         │                                                       │
│         ▼                                                       │
│  process_sections_loop                                          │
│         │                                                       │
│         ├── generate_section_content (execute_llm_prompt)       │
│         │      │                                                │
│         │      │  Prompt includes:                              │
│         │      │  {{.link_context.link_constraint_text}}        │
│         │      │                                                │
│         │      └── LLM generates with link awareness            │
│         │                                                       │
│         └── ... next section                                    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Files

| File | Type | Purpose |
|------|------|---------|
| `21_prepare_link_context_action.go` | Action | Prepares link context, register as `prepare_link_context` |
| `22_page_content_writer_link_context.sql` | SQL | Updates workflow to use the action |

## Implementation Details

### 1. Action: prepare_link_context

**Registered as:** `"prepare_link_context": PrepareLinkContextAction`

**Config:**
```json
{
  "enabled": true,
  "pages_field": "db_sync.pages",
  "max_links_per_section": 3
}
```

**Output:**
```json
{
  "enabled": true,
  "available_pages": [
    {"url": "/index.html", "title": "Home", "name": "index"},
    {"url": "/about.html", "title": "About Us", "name": "about"},
    {"url": "/services.html", "title": "Our Services", "name": "services"},
    {"url": "/contact.html", "title": "Contact", "name": "contact"}
  ],
  "link_constraint_text": "## Internal Links\n\nWhen creating internal links, ONLY link to these pages:\n\n- /index.html (Home)\n- /about.html (About Us)\n...",
  "page_count": 4
}
```

### 2. Workflow Update

The page-content-writer workflow gains a new step:

```
spawn_research_agent
    │
    ▼
prepare_link_context  ◄── NEW
    │
    ▼
load_page_components
    │
    ... rest of workflow
```

### 3. Prompt Template Update

The LLM prompt for section generation includes the constraint text:

```
{{if .link_context.link_constraint_text}}
{{.link_context.link_constraint_text}}

---
{{end}}

Generate content for the {{.current_section.name}} section...
```

## Data Flow

```
1. site-planner creates page plan
       │
       ▼
2. sync_pages_to_db stores pages in DB, returns db_sync.pages
       │
       ▼
3. page-content-writer receives db_sync in input_mapping
       │
       ▼
4. prepare_link_context action runs
       │
       ├── Extracts pages from db_sync.pages
       │
       ├── Builds link_constraint_text
       │
       └── Stores in link_context output_field
       │
       ▼
5. process_sections_loop runs with link_context available
       │
       ├── Prompt template includes {{.link_context.link_constraint_text}}
       │
       └── LLM generates content with link awareness
       │
       ▼
6. content-reviewer validates (catches any violations)
```

## Validation Layer

Even with constraints, the LLM may occasionally violate them. The `content-reviewer` agent has `validate_page_content` action that:

1. Extracts all `href` values from generated HTML
2. Checks each against pages table
3. Flags broken links as errors
4. Forces HITL review if violations found

This provides defense-in-depth.

## What This Does NOT Do

- **Does NOT suggest new pages** - That's a maintenance workflow concern
- **Does NOT auto-insert links** - LLM decides where links make sense
- **Does NOT analyze content gaps** - Separate link-analyzer process
- **Does NOT handle external links** - Only internal site links
- **Does NOT modify execute_llm_prompt** - Stays in page-content-writer scope

## Deployment Steps

```bash
# 1. Copy action to actions package
cp 21_prepare_link_context_action.go platform/orchestration/actions/

# 2. Register action in action_registry.go
#    Add: "prepare_link_context": PrepareLinkContextAction

# 3. Rebuild
make build-agent-chassis
make push-agent-chassis

# 4. Update page-content-writer workflow
psql -f 22_page_content_writer_link_context.sql

# 5. Update prompt template to include {{.link_context.link_constraint_text}}
#    (in content_components or agent definition prompt)

# 6. Restart pods
make update-agent-images
```

## Testing

1. **Unit test:** Call `PrepareLinkContextAction` with mock db_sync, verify output
2. **Integration test:** Run page-content-writer, check link_context in collected_data
3. **E2E test:** Build a site, verify no broken internal links in output

## Configuration Reference

```go
// Action config
type PrepareLinkContextConfig struct {
    // Enable/disable link context preparation
    Enabled bool `json:"enabled"`
    
    // Path to pages array in collected_data
    PagesField string `json:"pages_field"`
    
    // Maximum internal links per content section
    MaxLinksPerSection int `json:"max_links_per_section"`
}

// Action output
type LinkContext struct {
    Enabled             bool       `json:"enabled"`
    AvailablePages      []PageInfo `json:"available_pages"`
    LinkConstraintText  string     `json:"link_constraint_text"`
    PageCount           int        `json:"page_count"`
}
```

## Troubleshooting

| Issue | Cause | Solution |
|-------|-------|----------|
| No constraints in prompt | link_context not in input_fields | Add `link_context` to LLM step's input_fields |
| Empty pages list | db_sync not passed to content-writer | Verify input_mapping in pageflow-builder |
| LLM still hallucinates | Prompt template missing constraint | Add `{{.link_context.link_constraint_text}}` |
| Valid links flagged | Page status not 'deployed' | Check pages table status column |
| Action not found | Not registered | Add to action_registry.go |