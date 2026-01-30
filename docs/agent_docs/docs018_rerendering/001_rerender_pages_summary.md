# Re-render Site Pages Agent

## Purpose
Re-assembles all deployed pages with current components without regenerating content. Use when:
- Component templates are updated (head, header, footer)
- CSS stylesheet links are added/changed
- Navigation structure changes
- Branding updates across all pages

## Agent: `rerender-pages`

**Input:**
- `site_id` or `domain` (looks up site if only domain provided)

**Output:**
```json
{
  "rerender_result": {
    "success": true,
    "pages_rendered": 5,
    "pages": [...]
  },
  "deploy_result": { "iterations": [...] }
}
```

## What It Does
1. Loads pages from `pages` table
2. Loads section HTML from `page_components.rendered_html`
3. Strips any existing page wrapper (DOCTYPE, head, body)
4. Applies current `head` component (includes CSS link)
5. Injects header/footer via `InjectHeader`/`InjectFooter`
6. Commits each page via git adapter

## Deployment

1. Register action in Go:
```go
// action_registry.go
"rerender_site_pages": RerenderSitePagesAction,
```

2. Apply SQL:
```bash
psql -f 11_rerender_pages_agent.sql
```

## Trigger

**Via agent definition:**
```bash
./trigger_rerender_agent.sh
```

**Via inline workflow (no DB entry needed):**
```bash
./trigger_rerender_pages.sh
```

## Files
| File | Purpose |
|------|---------|
| `10_rerender_pages_action.go` | Action implementation |
| `11_rerender_pages_agent.sql` | Agent definition |
| `trigger_rerender_agent.sh` | Trigger via agent type |
| `trigger_rerender_pages.sh` | Trigger via inline workflow |