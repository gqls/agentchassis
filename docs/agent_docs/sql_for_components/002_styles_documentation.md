That's a good direction. Let me look at the content_components structure to see what we have:This is a good approach. We have the schema already. Let me propose a design using the existing tables:## Summary: Component-Based Headers

The idea is to use the existing `content_components` table to store **tested, working header templates** that get rendered with site-specific data.

### The Structure

```
Style Collection (e.g., "professional-dark")
├── Header Component (for internal pages)
├── Header Home Component (optional variant for home)
├── Footer Component
├── CSS Theme
└── Color Palette

Site
├── Links to: Style Collection
└── Can override: Colors, logo
```

### How It Works

1. **Header templates** stored in `content_components` with Handlebars-style placeholders:
   ```html
   <a class="logo">{{logo_text}}</a>
   {{#each nav_items}}
   <li><a href="{{this.url}}">{{this.label}}</a></li>
   {{/each}}
   ```

2. **Style collections** bundle header + footer + colors together - sites link to a collection

3. **At build time**, we:
    - Get site's style collection
    - Load the header component
    - Render it with: logo from domain, nav from `db_sync.navigation`, colors from collection
    - Inject into page (replacing whatever the LLM generated)

### Benefits Over Current Approach

| Current (Code-based) | Component-based |
|---------------------|-----------------|
| Header HTML hardcoded in Go | Header stored in DB, easy to edit |
| One style fits all | Multiple style options per collection |
| Changes require deploy | Changes are instant (DB update) |
| Can't A/B test | Can assign different collections |

### What You'd Need

**Immediate (works now):**
- The `multipage_actions_fixes.go` already does header injection with code-based templates

**For component system:**
1. Run `style_collections_schema.sql` to create tables
2. Add `component_renderer.go` to actions
3. Update assembly to call `RenderHeaderForSite` instead of `generateConsistentHeader`

The component approach is more maintainable long-term, especially when you want to offer different header styles (minimal, bold, centered logo, etc.) or let clients customize their look.

