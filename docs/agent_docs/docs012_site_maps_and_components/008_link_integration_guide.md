# Link Management Integration Guide

## How It Fits With multipage-website-builder

### Current Flow (Before Integration)

```
intake-orchestrator
    └─► multipage-website-builder
            ├─► chief-strategist ──────► page_plan (pages[], sitemap[])
            ├─► content-creator (loop) ─► page_content
            ├─► html-developer (loop) ──► page_html
            ├─► assemble_multipage_site ► site_files
            └─► deployer ────────────────► github
```

**Problem**: Navigation URLs guessed by LLM, no persistence, no change tracking.

### New Flow (With Integration)

```
intake-orchestrator
    └─► multipage-website-builder
            │
            ├─► [NEW] ensure_site_record ──► site_id (in DB)
            │
            ├─► chief-strategist ──────────► page_plan
            │
            ├─► [NEW] sync_pages_to_db ────► pages in DB, navigation cached
            │
            ├─► content-creator (loop) ────► page_content
            │
            ├─► html-developer (loop)
            │       │
            │       ├─► [CHANGED] receives navigation from DB
            │       └─► [NEW] after render: extract links, sync to DB
            │
            ├─► assemble_multipage_site
            │       │
            │       └─► [CHANGED] reads navigation from DB
            │
            ├─► [NEW] sync_components_to_db ► page_components tracked
            │
            └─► deployer ──────────────────► github + update timestamps
```

---

## Minimal Changes for MVP

### 1. New Action: `ensure_site_record`

Creates or updates site record in database:

```go
// In coordinator.go or new file site_db_actions.go

func EnsureSiteRecordAction(ctx context.Context, params ActionParams) (interface{}, error) {
    domain := extractDomain(params.CollectedData)
    
    // Get or create site
    var siteID uuid.UUID
    err := db.QueryRowContext(ctx, `
        INSERT INTO sites (domain, name, network_id)
        VALUES ($1, $1, $2)
        ON CONFLICT (domain) DO UPDATE SET updated_at = NOW()
        RETURNING id
    `, domain, defaultNetworkID).Scan(&siteID)
    
    return map[string]interface{}{
        "site_id": siteID,
        "domain":  domain,
    }, err
}
```

### 2. New Action: `sync_pages_to_db`

After strategist, sync pages to database and build navigation:

```go
func SyncPagesToDB(ctx context.Context, params ActionParams) (interface{}, error) {
    siteID := params.CollectedData["site_id"].(uuid.UUID)
    pagePlan := extractPagePlan(params.CollectedData)
    
    for _, page := range pagePlan.Pages {
        _, err := db.ExecContext(ctx, `
            INSERT INTO pages (site_id, name, url, title, nav_label, nav_order, in_header, in_footer)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            ON CONFLICT (site_id, name) DO UPDATE SET
                url = EXCLUDED.url,
                title = EXCLUDED.title,
                nav_label = EXCLUDED.nav_label,
                updated_at = NOW()
        `, siteID, page.Name, page.URL, page.Title, 
           page.NavLabel, page.NavOrder, page.InHeader, page.InFooter)
    }
    
    // Navigation cache auto-invalidated by trigger
    // Get fresh navigation
    var navStructure json.RawMessage
    db.QueryRowContext(ctx, `SELECT get_current_navigation($1, 'header')`, siteID).Scan(&navStructure)
    
    return map[string]interface{}{
        "navigation": navStructure,
        "pages_synced": len(pagePlan.Pages),
    }, nil
}
```

### 3. Modified: `extractSitemapInfo` in html_actions.go

Add database lookup before falling back to context:

```go
func extractSitemapInfo(context map[string]interface{}) string {
    // PRIORITY 1: Check for pre-built navigation from DB
    if nav, ok := context["navigation"].(map[string]interface{}); ok {
        return formatNavigationFromStructure(nav)
    }
    
    // PRIORITY 2: Check for link_data.navigation (from link-manager if used)
    if linkData, ok := context["link_data"].(map[string]interface{}); ok {
        if nav, ok := linkData["navigation"].(map[string]interface{}); ok {
            return formatNavigationFromStructure(nav)
        }
    }
    
    // PRIORITY 3: Fall back to sitemap in context (existing behavior)
    // ... existing code ...
}

func formatNavigationFromStructure(nav map[string]interface{}) string {
    items, ok := nav["items"].([]interface{})
    if !ok {
        return ""
    }
    
    var headerNav []string
    for _, item := range items {
        if itemMap, ok := item.(map[string]interface{}); ok {
            label, _ := itemMap["label"].(string)
            url, _ := itemMap["url"].(string)
            if label != "" && url != "" {
                headerNav = append(headerNav, fmt.Sprintf("%s -> %s", label, url))
            }
        }
    }
    
    if len(headerNav) == 0 {
        return ""
    }
    
    var result strings.Builder
    result.WriteString("NAVIGATION (use these EXACT relative URLs):\n")
    result.WriteString("Header navigation: ")
    result.WriteString(strings.Join(headerNav, " | "))
    result.WriteString("\n")
    
    return result.String()
}
```

### 4. New Action: `extract_and_sync_links`

After HTML is rendered, extract links:

```go
func ExtractAndSyncLinks(ctx context.Context, params ActionParams) (interface{}, error) {
    pageID := params.CollectedData["current_page_id"].(uuid.UUID)
    siteID := params.CollectedData["site_id"].(uuid.UUID)
    htmlContent := params.CollectedData["page_html"].(string)
    
    // Parse HTML for links
    doc, _ := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
    
    // Clear existing links for this page
    db.ExecContext(ctx, `DELETE FROM link_registry WHERE source_page_id = $1`, pageID)
    
    // Extract and insert new links
    var linksExtracted int
    doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
        href, _ := s.Attr("href")
        anchorText := strings.TrimSpace(s.Text())
        rel, _ := s.Attr("rel")
        
        scope := classifyScope(href, siteID)
        linkType := classifyLinkType(s) // based on parent element, classes, etc.
        
        // Resolve internal links
        var targetPageID *uuid.UUID
        if scope == "page" {
            targetPageID = resolvePageID(ctx, db, siteID, href)
        }
        
        db.ExecContext(ctx, `
            INSERT INTO link_registry (
                source_page_id, source_site_id, target_url, target_page_id,
                scope, link_type, anchor_text, rel_attr
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        `, pageID, siteID, href, targetPageID, scope, linkType, anchorText, rel)
        
        linksExtracted++
    })
    
    return map[string]interface{}{
        "links_extracted": linksExtracted,
    }, nil
}

func classifyScope(href string, currentSiteID uuid.UUID) string {
    if strings.HasPrefix(href, "#") {
        return "internal" // anchor
    }
    if strings.HasPrefix(href, "/") {
        return "page" // relative, same site
    }
    if strings.HasPrefix(href, "http") {
        // Check if it's our site or another site in network
        // For now, treat as external
        return "external"
    }
    return "page" // default to relative
}
```

---

## Updated Workflow Definition

```sql
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow}',
    '{
        "start_step": "ensure_site",
        "steps": {
            "ensure_site": {
                "action": "ensure_site_record",
                "config": {},
                "output_field": "site_record",
                "next_step": "spawn_strategist",
                "description": "Create or get site record in database"
            },
            "spawn_strategist": {
                "action": "spawn_agent",
                "config": {"agent_type": "chief-strategist", "role": "strategist"},
                "next_step": "spawn_content_creator",
                "output_field": "strategist_info"
            },
            "spawn_content_creator": {
                "action": "spawn_agent",
                "config": {"agent_type": "content-creator", "role": "writer"},
                "next_step": "spawn_html_developer",
                "output_field": "writer_info"
            },
            "spawn_html_developer": {
                "action": "spawn_agent",
                "config": {"agent_type": "html-developer", "role": "developer"},
                "next_step": "spawn_deployer",
                "output_field": "developer_info"
            },
            "spawn_deployer": {
                "action": "spawn_agent",
                "config": {"agent_type": "deployer-agent", "role": "deployer"},
                "next_step": "call_strategist",
                "output_field": "deployer_info"
            },
            "call_strategist": {
                "action": "call_agent",
                "config": {
                    "agent_type": "chief-strategist",
                    "target_role": "strategist",
                    "input_fields": ["input_data", "site_record"],
                    "timeout_seconds": 120
                },
                "next_step": "sync_pages",
                "output_field": "page_plan"
            },
            "sync_pages": {
                "action": "sync_pages_to_db",
                "config": {},
                "output_field": "db_sync",
                "next_step": "generate_pages_loop",
                "description": "Sync pages to database, build navigation"
            },
            "generate_pages_loop": {
                "action": "loop",
                "config": {
                    "iterate_over": "page_plan.plan_data.pages",
                    "loop_var": "current_page",
                    "max_iterations": 10,
                    "substeps": {
                        "generate_content": {
                            "action": "call_agent",
                            "config": {
                                "agent_type": "content-creator",
                                "target_role": "writer",
                                "input_fields": ["current_page", "input_data", "page_plan"],
                                "timeout_seconds": 180
                            },
                            "next_step": "create_html",
                            "output_field": "page_content"
                        },
                        "create_html": {
                            "action": "call_agent",
                            "config": {
                                "agent_type": "html-developer",
                                "target_role": "developer",
                                "input_fields": ["page_content", "current_page", "input_data", "db_sync"],
                                "timeout_seconds": 180
                            },
                            "next_step": "extract_links",
                            "output_field": "page_html"
                        },
                        "extract_links": {
                            "action": "extract_and_sync_links",
                            "config": {},
                            "output_field": "link_sync",
                            "description": "Extract links from HTML, sync to registry"
                        }
                    }
                },
                "next_step": "assemble_site",
                "output_field": "all_pages"
            },
            "assemble_site": {
                "action": "assemble_multipage_site",
                "config": {
                    "pages_field": "all_pages",
                    "include_sitemap_xml": true,
                    "include_robots_txt": true
                },
                "next_step": "deploy",
                "output_field": "site_files"
            },
            "deploy": {
                "action": "call_agent",
                "config": {
                    "agent_type": "deployer-agent",
                    "target_role": "deployer",
                    "input_fields": ["site_files", "input_data", "site_record"],
                    "timeout_seconds": 180
                },
                "next_step": "update_timestamps",
                "output_field": "deployment_result"
            },
            "update_timestamps": {
                "action": "update_site_timestamps",
                "config": {},
                "next_step": "complete",
                "description": "Update last_built_at, last_deployed_at"
            },
            "complete": {
                "action": "complete_workflow"
            }
        }
    }'::jsonb
)
WHERE type = 'multipage-website-builder';
```

---

## Data Flow

```
1. ensure_site_record
   Input:  domain from input_data
   Output: site_id (UUID)
   DB:     sites table (insert/update)

2. chief-strategist  
   Input:  input_data, site_record
   Output: page_plan {pages: [...], sitemap: [...]}
   DB:     none (LLM only)

3. sync_pages_to_db
   Input:  site_id, page_plan
   Output: navigation structure
   DB:     pages table (insert/update)
          navigation_structures (rebuilt via trigger)

4. html-developer (per page)
   Input:  page_content, current_page, db_sync.navigation
   Output: page_html
   DB:     none (but uses navigation from DB)

5. extract_and_sync_links (per page)
   Input:  page_html, site_id, current_page
   Output: links_extracted count
   DB:     link_registry (delete old, insert new)

6. assemble_multipage_site
   Input:  all_pages
   Output: site_files
   DB:     can read navigation_structures for sitemap.xml

7. deployer
   Input:  site_files, site_record
   Output: deployment_result
   DB:     none (GitHub only)

8. update_site_timestamps
   Input:  site_id
   Output: none
   DB:     sites.last_built_at, last_deployed_at
```

---

## What This Enables

### Now (MVP)
- ✅ Correct relative URLs in navigation
- ✅ Pages tracked in database
- ✅ Links indexed for queries
- ✅ Navigation cached, auto-invalidated

### Soon (Phase 2)
- Component instance tracking (page_components)
- Change detection via content_hash
- Patch updates: only re-render changed components

### Later (Phase 3+)
- Semantic relationships via existing `relationships` table
- Cross-site linking within networks
- Affiliate link management
- Link health monitoring
- Topic extraction for vector search

---

## Files to Create/Modify

| File                                | Change                   | Priority   |
|-------------------------------------|--------------------------|------------|
| `link_management_migration.sql`     | New tables               | Now        |
| `site_db_actions.go`                | New actions for DB sync  | Now        |
| `html_actions.go`                   | Add DB navigation lookup | Now        |
| `multipage_actions.go`              | Add link extraction      | Now        |
| `017_multipage_website_builder.sql` | Updated workflow         | Now        |
| `coordinator.go`                    | Register new actions     | Now        |

--



===============


Looking at the patterns in your existing actions, I'll create the Go implementation files. Let me start with the site database actions:Now I have a clear picture of the patterns. Let me create the site database actions file:

