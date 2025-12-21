# Component-Based Headers - Architecture

## Concept

Instead of letting the LLM generate headers (which produces inconsistent results), we:

1. **Store tested header templates** in `content_components` table
2. **Group components into style collections** - a bundle that defines a site's look
3. **Link sites to style collections** - each site uses a specific collection
4. **Render headers from templates** - inject site-specific data (logo, nav, colors)

## Database Structure

```
┌─────────────────────┐      ┌──────────────────────┐
│       sites         │      │  style_collections   │
├─────────────────────┤      ├──────────────────────┤
│ id                  │      │ id                   │
│ domain              │      │ name                 │
│ style_collection_id │─────>│ header_component_id  │──┐
│ style_overrides     │      │ header_home_id       │  │
└─────────────────────┘      │ footer_component_id  │  │
                             │ css_theme_id         │  │
                             │ color_palette        │  │
                             └──────────────────────┘  │
                                                       │
                             ┌──────────────────────┐  │
                             │ content_components   │<─┘
                             ├──────────────────────┤
                             │ id                   │
                             │ name                 │
                             │ html_template        │
                             │ input_schema         │
                             │ category: "header"   │
                             └──────────────────────┘
```

## Example Header Component

Stored in `content_components`:

```html
<!-- name: header-professional-dark -->
<header class="site-header">
    <div class="header-container">
        <a href="/index.html" class="logo">
            <span class="logo-text">{{logo_text}}</span>
        </a>
        <nav class="main-nav">
            <ul>
                {{#each nav_items}}
                <li><a href="{{this.url}}"{{#if this.is_active}} class="active"{{/if}}>{{this.label}}</a></li>
                {{/each}}
            </ul>
        </nav>
    </div>
</header>
<style>
.site-header { background: {{primary_color}}; ... }
.main-nav a:hover { color: {{accent_color}}; }
</style>
```

With input schema:
```json
{
  "logo_text": "string",
  "primary_color": "#1a1a2e",
  "accent_color": "#16a085", 
  "nav_items": [{"label": "string", "url": "string", "is_active": false}]
}
```

## Style Collections

A collection bundles related components:

| Collection | Header | Footer | Colors |
|------------|--------|--------|--------|
| professional-dark | header-professional-dark | footer-standard | navy/teal |
| minimal-light | header-minimal-light | footer-minimal | white/blue |
| creative-bold | header-creative | footer-creative | custom |

## Rendering Flow

```
1. Site build starts
   │
   ├─> Get site's style_collection_id
   │
   ├─> Load header component from collection
   │   (or use header_home_id for index page)
   │
   ├─> Build render input:
   │   - logo_text: from domain
   │   - nav_items: from db_sync.navigation
   │   - colors: from collection.color_palette + site.style_overrides
   │
   ├─> Render template with data
   │
   └─> Inject into page HTML (replacing LLM header)
```

## Site-Specific Overrides

Sites can override collection colors without creating a new collection:

```sql
UPDATE sites 
SET style_overrides = '{
    "color_palette": {
        "primary": "#2d3436",
        "accent": "#00cec9"
    }
}'::jsonb
WHERE domain = 'specialclient.com';
```

## Benefits

| Before (LLM) | After (Components) |
|--------------|-------------------|
| Different nav labels each page | Same labels everywhere |
| Different layouts/styles | Consistent design |
| Broken mobile menus | Tested, working JS |
| Random colors | Controlled palette |
| Hard to update | Update once, affects all sites |

## Files

| File | Purpose |
|------|---------|
| `style_collections_schema.sql` | Database tables and example components |
| `component_renderer.go` | Go code to load and render components |

## Implementation Steps

### Phase 1: Schema
1. Run `style_collections_schema.sql` to create tables
2. Insert header component templates
3. Create initial style collections

### Phase 2: Integration
1. Add `component_renderer.go` to actions
2. Update `AssembleMultipageSiteAction` to use `RenderHeaderForSite`
3. Update `EnsureSiteRecordAction` to assign default style collection

### Phase 3: Site Assignment
1. When creating a site, assign a style collection based on:
    - Industry vertical from classification
    - Explicit selection in brief
    - Default: "professional-dark"

## Future Extensions

- **Footer components** - Same pattern for consistent footers
- **Section components** - Hero, testimonials, features sections
- **Theme inheritance** - Collections can extend base collections
- **A/B testing** - Assign different collections to test variants
- **User customization** - Let users pick collection in brief questionnaire