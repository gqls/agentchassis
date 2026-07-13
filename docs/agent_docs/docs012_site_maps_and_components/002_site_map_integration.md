Good thinking. Currently the strategist outputs `global.navigation` as just labels `["Home", "About", "Services", "Contact"]` but doesn't map them to URLs or page names. The HTML developer needs to know how pages link together.

Let me look at what we have and propose a sitemap structure:I see the issue. The strategist outputs `navigation: ["Home", "About", "Services", "Contact"]` as labels only, with no mapping to page URLs. The HTML developer has to guess that "About" → `about.html`.

**Proposed Solution: Add explicit sitemap to strategist output**

```json
{
  "pages": [...],
  "sitemap": [
    {"label": "Home", "page": "index", "url": "/index.html", "in_header": true, "in_footer": true},
    {"label": "About Us", "page": "about", "url": "/about.html", "in_header": true, "in_footer": true},
    {"label": "Services", "page": "services", "url": "/services.html", "in_header": true, "in_footer": true},
    {"label": "Contact", "page": "contact", "url": "/contact.html", "in_header": true, "in_footer": true},
    {"label": "Privacy Policy", "page": "privacy", "url": "/privacy.html", "in_header": false, "in_footer": true}
  ],
  "global": {
    "brand_tone": "professional"
  }
}
```

This gives the HTML developer:
- Exact URLs for each page
- Which links go in header nav vs footer
- Mapping between display labels and page files

Let me create the updated strategist prompt:Now the html-developer needs to use the sitemap. Let me check the current html-developer configuration:The html-developer uses the `generate_html` action which builds the prompt in Go code. Let me check how the prompt uses navigation:Now I understand. The `buildFullHTMLPrompt` function needs to include sitemap information for navigation. Let me create a patch for `html_actions.go`:Now let me check what fields are passed to the html-developer in the multipage loop:Found it. The html-developer receives `["page_content", "current_page", "input_data"]` but NOT `page_plan` (which contains the sitemap).

Let me create a complete SQL update that:
1. Updates the strategist with sitemap output
2. Updates the loop to pass `page_plan` to html-developer## Summary: Sitemap-Enabled Navigation

### Data Flow

```
chief-strategist
    ↓ outputs: {pages: [...], sitemap: [...], global: {...}}
    
multipage-website-builder (loop)
    ↓ passes to html-developer: page_plan (contains sitemap)
    
html-developer
    ↓ extracts sitemap from page_plan.plan_data.sitemap
    ↓ includes in prompt: "Header nav: Home -> /index.html | About -> /about.html"
    
Generated HTML
    ↓ has correct <nav> links using exact URLs from sitemap
```

### Changes Required

| File | Change |
|------|--------|
| **sitemap_navigation_complete.sql** | Updates chief-strategist prompt to output sitemap; Updates multipage-website-builder to pass `page_plan` to html-developer; Changes `iterate_over` to `pages` |
| **html_actions_sitemap_patch.go** | Adds `extractSitemapInfo()` function; Updates `buildFullHTMLPrompt()` to include navigation URLs in prompt |
| **transform_actions_patched.go** | Accepts `pages` OR `sections` in validation |
| **unified_extractor_patched.go** | Supports both `{{.domain}}` and `{{.input_data.domain}}` access |

### Sitemap Structure

```json
{
  "sitemap": [
    {"label": "Home", "page": "index", "url": "/index.html", "in_header": true, "in_footer": true},
    {"label": "About Us", "page": "about", "url": "/about.html", "in_header": true, "in_footer": true},
    {"label": "Services", "page": "services", "url": "/services.html", "in_header": true, "in_footer": true},
    {"label": "Contact", "page": "contact", "url": "/contact.html", "in_header": true, "in_footer": true},
    {"label": "Privacy", "page": "privacy", "url": "/privacy.html", "in_header": false, "in_footer": true}
  ]
}
```

### Key Points

1. **html-developer now receives `page_plan`** - This gives it access to the sitemap
2. **Sitemap defines exact URLs** - No guessing needed
3. **in_header/in_footer flags** - Control where each link appears
4. **Page/label separation** - "About Us" label can link to `about.html` page
