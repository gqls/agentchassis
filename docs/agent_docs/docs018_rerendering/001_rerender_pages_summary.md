# Re-render Site Pages Action

## Purpose
Re-assembles all deployed pages with current components without regenerating content. Use when:
- Component templates are updated (head, header, footer)
- CSS stylesheet links are added/changed
- Navigation structure changes
- Branding updates across all pages

## Action: `rerender_site_pages`

**Input:**
- `site_id` or `domain` (looks up site if only domain provided)
- `include_statuses`: page statuses to include (default: `["deployed", "active"]`)

**Output:**
```json
{
  "success": true,
  "site_id": "uuid",
  "domain": "example.com",
  "pages_rendered": 5,
  "pages": [
    {"page_id": "...", "title": "...", "name": "...", "slug": "index", "filename": "index.html", "html": "..."}
  ]
}
```

## What It Does
1. Loads pages from `pages` table
2. Loads section HTML from `page_components.rendered_html`
3. Strips any existing page wrapper (DOCTYPE, head, body)
4. Applies current `head` component (includes CSS link)
5. Injects header/footer via `InjectHeader`/`InjectFooter`
6. Returns assembled pages ready for git commit

## Trigger Script
```bash
./trigger_rerender_pages.sh
```

Sends inline workflow to `system.agent.generic.requests`:
1. `rerender_site_pages` → load and reassemble
2. `loop` → for each page, `git_commit` → git adapter

## Registration
```go
// action_registry.go
"rerender_site_pages": RerenderSitePagesAction,
```

## Files
- `10_rerender_pages_action.go` - Action implementation
- `trigger_rerender_pages.sh` - CLI trigger script