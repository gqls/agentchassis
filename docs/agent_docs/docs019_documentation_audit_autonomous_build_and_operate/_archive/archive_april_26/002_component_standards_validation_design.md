# Component Standards Validation — Discovery Check Design

## Overview

A new discovery check `validate_component_standards` that audits all components
used by a site against the contracts in 003b. Runs as part of the improvement
loop's `run_discovery_checks` action. Produces work items for each violation,
routed to the appropriate handler agent.

This replaces manual auditing with automated detection that runs every cycle.

## Check: `validate_component_standards`

File: `actions/discovery_checks/check_component_standards.go`

Implements the `DiscoveryCheck` interface:

```go
type ComponentStandardsCheck struct{}

func (c *ComponentStandardsCheck) Name() string { return "validate_component_standards" }

func (c *ComponentStandardsCheck) Run(dctx DiscoveryCheckContext) (*DiscoveryResult, error) {
    var result DiscoveryResult

    // Sub-checks run in order, each appends findings and work items
    checkUnlinkedSiteComponents(dctx, &result)
    checkMissingDataComponent(dctx, &result)
    checkSlotNameMismatch(dctx, &result)
    checkMissingSiteMetadata(dctx, &result)
    checkMissingAssetRefs(dctx, &result)
    checkNavLayout(dctx, &result)
    checkUnwantedElements(dctx, &result)
    checkDarkSectionContract(dctx, &result)  // existing, but consolidate here
    checkEmptyPageSections(dctx, &result)

    return &result, nil
}

func init() {
    Register("validate_component_standards", &ComponentStandardsCheck{})
}
```

## Sub-checks

### 1. `checkUnlinkedSiteComponents`

**What:** site_components rows with NULL component_id for header/footer/head.

**Query:**
```sql
SELECT sc.slot_name
FROM site_components sc
WHERE sc.site_id = $1
  AND sc.slot_name IN ('header', 'footer', 'head')
  AND sc.component_id IS NULL
```

**Work item:**
```json
{
    "item_type": "unlinked_site_component",
    "handler_agent": "site-component-linker",
    "severity": "high",
    "priority": 5,
    "spec": { "slot_name": "header" },
    "summary": "Header component not linked to template — using fallback rendering"
}
```

**Handler: `site-component-linker`** (new, simple workflow)
- Loads style_collection for the site
- Reads header_component_id / footer_component_id from collection
- Updates site_components.component_id
- Clears rendered_html to force re-render
- Creates needs_rerender work item

### 2. `checkMissingDataComponent`

**What:** page_components with rendered_html that doesn't have a `data-component` attribute.

**Query:**
```sql
SELECT pc.id, pc.slot_name, p.name as page_name
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
WHERE p.site_id = $1
  AND pc.rendered_html IS NOT NULL
  AND pc.rendered_html != ''
  AND pc.rendered_html NOT LIKE '%data-component=%'
```

**Work item:**
```json
{
    "item_type": "missing_data_component_attr",
    "handler_agent": "component-template-fixer",
    "severity": "medium",
    "spec": { "fix_type": "add_data_component", "page_name": "index", "slot_name": "hero" }
}
```

**Handler:** `component-template-fixer` — injects `data-component="{slot_name}"` on the root element.

### 3. `checkSlotNameMismatch`

**What:** page_components where slot_name doesn't match the data-component attribute in rendered_html.

**Query:**
```sql
SELECT pc.id, pc.slot_name,
       substring(pc.rendered_html from 'data-component="([^"]*)"') as data_component,
       p.name as page_name
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
WHERE p.site_id = $1
  AND pc.slot_name IS NOT NULL
  AND pc.rendered_html LIKE '%data-component=%'
  AND pc.slot_name != substring(pc.rendered_html from 'data-component="([^"]*)"')
```

**Work item:**
```json
{
    "item_type": "slot_name_mismatch",
    "handler_agent": "component-template-fixer",
    "severity": "medium",
    "spec": {
        "fix_type": "align_slot_name",
        "page_component_id": "fc5c9dac-...",
        "current_slot_name": "contact-form",
        "data_component": "contact-info"
    }
}
```

**Handler:** `component-template-fixer` — updates slot_name to match data-component.

### 4. `checkMissingSiteMetadata`

**What:** Sites with empty company_name, tagline, logo_url, email.

**Query:**
```sql
SELECT
    s.company_name IS NULL OR s.company_name = '' as missing_company,
    s.tagline IS NULL OR s.tagline = '' as missing_tagline,
    s.logo_url IS NULL OR s.logo_url = '' as missing_logo,
    s.email IS NULL OR s.email = '' as missing_email
FROM sites s
WHERE s.id = $1
```

**Work item:**
```json
{
    "item_type": "missing_site_metadata",
    "handler_agent": "site-metadata-fixer",
    "severity": "high",
    "priority": 3,
    "spec": {
        "missing_fields": ["company_name", "tagline", "email"],
        "derive_from": "content_data"
    }
}
```

**Handler: `site-metadata-fixer`** (new)
- Reads content_data.response.pages[0].title to derive company name
- Reads content_data.logo_url for logo
- Derives email from domain (info@{domain})
- Optionally calls LLM for tagline from domain + page titles
- Updates sites table
- Creates needs_rerender work item

### 5. `checkMissingAssetRefs`

**What:** Sites with logo_url/hero_url set but not appearing in rendered header/hero HTML.

**Logic:**
```go
func checkMissingAssetRefs(dctx DiscoveryCheckContext, result *DiscoveryResult) {
    var logoURL string
    dctx.DB.QueryRow(`SELECT COALESCE(logo_url,'') FROM sites WHERE id=$1`, dctx.SiteID).Scan(&logoURL)

    if logoURL != "" {
        var headerHTML string
        dctx.DB.QueryRow(`SELECT COALESCE(rendered_html,'') FROM site_components
            WHERE site_id=$1 AND slot_name='header'`, dctx.SiteID).Scan(&headerHTML)

        if !strings.Contains(headerHTML, "<img") {
            // Logo URL exists but header has no <img> tag
            result.AddWorkItem(WorkItemSpec{
                ItemType:     "missing_logo_in_header",
                HandlerAgent: "site-component-linker",  // relink + rerender fixes this
                Severity:     "high",
                Priority:     5,
                Summary:      "Logo URL set but header doesn't render <img> tag",
            })
        }
    }
}
```

**Handler:** `site-component-linker` — the root cause is usually unlinked site_components.
If components ARE linked but logo still missing, escalate to `component-template-fixer`.

### 6. `checkNavLayout`

**What:** Rendered header HTML where nav `<ul>` has no flex CSS.

**Logic:**
```go
func checkNavLayout(dctx DiscoveryCheckContext, result *DiscoveryResult) {
    var headerHTML string
    dctx.DB.QueryRow(`SELECT COALESCE(rendered_html,'') FROM site_components
        WHERE site_id=$1 AND slot_name='header'`, dctx.SiteID).Scan(&headerHTML)

    // Check if nav list has display:flex in the component's <style> block
    hasNavFlex := strings.Contains(headerHTML, "display: flex") ||
                  strings.Contains(headerHTML, "display:flex")
    hasNavList := strings.Contains(headerHTML, "<ul")

    if hasNavList && !hasNavFlex {
        result.AddWorkItem(WorkItemSpec{
            ItemType:     "stacked_nav",
            HandlerAgent: "component-template-fixer",
            Severity:     "high",
            Priority:     5,
            Spec: map[string]interface{}{
                "fix_type": "inject_nav_flex_css",
                "slot_name": "header",
            },
            Summary: "Header nav list has no flex CSS — links stack vertically",
        })
    }
}
```

### 7. `checkUnwantedElements`

**What:** Search icon in header when site doesn't have search.

**Logic:**
```go
func checkUnwantedElements(dctx DiscoveryCheckContext, result *DiscoveryResult) {
    var headerHTML string
    dctx.DB.QueryRow(`SELECT COALESCE(rendered_html,'') FROM site_components
        WHERE site_id=$1 AND slot_name='header'`, dctx.SiteID).Scan(&headerHTML)

    if strings.Contains(headerHTML, "search-toggle") {
        result.AddWorkItem(WorkItemSpec{
            ItemType:     "unwanted_nav_element",
            HandlerAgent: "component-template-fixer",
            Severity:     "low",
            Priority:     50,
            Spec: map[string]interface{}{
                "fix_type": "remove_element",
                "slot_name": "header",
                "pattern":   "search-toggle",
            },
            Summary: "Header contains search icon — site has no search",
        })
    }
}
```

### 8. `checkDarkSectionContract`

**What:** Dark section components missing --section-* CSS variables. (Already exists as
separate check — consolidate here or keep separate and cross-reference.)

### 9. `checkEmptyPageSections`

**What:** Deployed pages with no rendered page_components.

**Query:**
```sql
SELECT p.name, p.id
FROM pages p
LEFT JOIN page_components pc ON pc.page_id = p.id
    AND pc.rendered_html IS NOT NULL AND pc.rendered_html != ''
WHERE p.site_id = $1
  AND p.build_status IN ('deployed', 'active')
GROUP BY p.id, p.name
HAVING COUNT(pc.id) = 0
```

**Work item:**
```json
{
    "item_type": "needs_content_page",
    "handler_agent": "page-content-writer",
    "severity": "high",
    "priority": 50,
    "page_id": "<page_uuid>",
    "spec": { "page_name": "blog" }
}
```

---

## New Handler Agents Needed

### 1. `site-component-linker`

Simple workflow: load site → get style collection → set component_ids → clear rendered_html → complete.

```json
{
    "type": "site-component-linker",
    "workflow": {
        "start_step": "ensure_site_record",
        "steps": {
            "ensure_site_record": {
                "action": "ensure_site_record",
                "next_step": "link_components",
                "output_field": "site_record"
            },
            "link_components": {
                "action": "link_site_components",
                "config": {
                    "site_id_field": "site_record.site_id"
                },
                "next_step": "create_rerender_item",
                "output_field": "link_result"
            },
            "create_rerender_item": {
                "action": "query_database",
                "config": {
                    "query": "INSERT INTO site_work_items (site_id, source, domain, item_type, severity, summary, priority, handler_agent, status, created_by, spec) VALUES ($1::uuid, 'side_effect', 'build', 'needs_rerender', 'medium', 'Rerender after component linkage fix', 99, 'rerender-pages', 'detected', 'site-component-linker', '{\"refresh_site_components\": true}'::jsonb) ON CONFLICT DO NOTHING",
                    "params_from": ["site_record.site_id"]
                },
                "next_step": "complete",
                "output_field": "rerender_item"
            },
            "complete": {
                "action": "complete_workflow",
                "config": { "output_fields": ["link_result"] }
            }
        }
    }
}
```

**New Go action: `link_site_components`**
```go
func LinkSiteComponentsAction(ctx context.Context, params ActionParams) (interface{}, error) {
    // 1. Get site_id
    // 2. Load style_collection for site
    // 3. For each slot (header, footer, head):
    //    - Get component_id from style collection (header_component_id etc)
    //    - UPDATE site_components SET component_id = ?, rendered_html = NULL, build_status = 'pending'
    // 4. Return count of linked components
}
```

### 2. `site-metadata-fixer`

Workflow: load site → extract metadata from content_data → optionally call LLM for tagline → update sites table → create rerender item → complete.

### 3. `component-template-fixer`

Already described in fix_05. Routes on `spec.fix_type`:
- `inject_nav_flex_css`
- `remove_element`
- `add_data_component`
- `align_slot_name`

---

## Registration

Add to the discovery checks registry:

```go
// In check_component_standards.go init():
func init() {
    Register("validate_component_standards", &ComponentStandardsCheck{})
}
```

Enable in the improvement loop's config:

```json
{
    "checks": [
        "hardcoded_section_colors",
        "forced_text_colors",
        "validate_component_standards"
    ]
}
```

---

## Ordering

The `validate_component_standards` check should run BEFORE other design checks
(hardcoded_colors, forced_text_colors) because those checks operate on
rendered HTML which may be wrong if the component isn't linked properly.
Fix the structural issues first, then audit the content.

Suggested check order:
1. `validate_component_standards` — structural correctness
2. `hardcoded_section_colors` — CSS variable compliance
3. `forced_text_colors` — text color inheritance
4. Other content/design checks
