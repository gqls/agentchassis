# Integrated Link Architecture

## Principle: Links Live in Components, Registry is an Index

The key insight is that **links are not separate entities** - they exist WITHIN components. What we need is:

1. **Source tracking**: Which component contains which links
2. **Index/Registry**: Queryable view of all links across sites
3. **Semantic relationships**: Use existing `relationships` table
4. **Navigation structures**: Aggregation of component-embedded links

This avoids duplication and keeps components as the source of truth.

---

## Part 1: What We Already Have

### Content Components (existing)
```sql
content_components:
  - id, name, html_template
  - function (e.g., 'navigation-header', 'cta-primary', 'related-articles')
  - semantic_tags (jsonb - can include link-related tags)
  - input_schema (defines what data the component needs)
```

### Relationships (existing, empty)
```sql
relationships:
  - source_entity_id, source_entity_type
  - target_entity_id, target_entity_type  
  - relationship_type (e.g., 'pillar_to_cluster', 'related_content', 'next_in_series')
  - properties (jsonb - relevance_score, auto_generated, etc.)
  - status ('active', 'archived')
```

This is PERFECT for semantic links between pages!

---

## Part 2: What We Need to Add

### 2.1 Client/Network Hierarchy

```sql
-- Clients (links to auth-service)
CREATE TABLE clients (
                         id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                         external_id VARCHAR(255) UNIQUE, -- ID from auth-service
                         name VARCHAR(255) NOT NULL,
                         settings JSONB DEFAULT '{}',
                         created_at TIMESTAMP DEFAULT NOW(),
                         updated_at TIMESTAMP DEFAULT NOW()
);

-- Networks belong to clients
CREATE TABLE networks (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          client_id UUID REFERENCES clients(id) ON DELETE CASCADE,
                          name VARCHAR(255) NOT NULL,
                          slug VARCHAR(100) NOT NULL,
                          description TEXT,
                          settings JSONB DEFAULT '{}', -- network-wide affiliate config, etc.
                          created_at TIMESTAMP DEFAULT NOW(),
                          updated_at TIMESTAMP DEFAULT NOW(),

                          UNIQUE(client_id, slug)
);

-- Sites belong to networks
CREATE TABLE sites (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       network_id UUID REFERENCES networks(id) ON DELETE CASCADE,
                       domain VARCHAR(255) NOT NULL UNIQUE,
                       name VARCHAR(255),
                       brand_dna JSONB DEFAULT '{}', -- visual identity, voice, invariants
                       github_repo VARCHAR(500),
                       github_branch VARCHAR(100) DEFAULT 'main',
                       settings JSONB DEFAULT '{}',
                       status VARCHAR(50) DEFAULT 'active',
                       last_built_at TIMESTAMP,
                       last_deployed_at TIMESTAMP,
                       created_at TIMESTAMP DEFAULT NOW(),
                       updated_at TIMESTAMP DEFAULT NOW()
);

-- Index for cross-network queries (within same client)
CREATE INDEX idx_sites_client ON sites(network_id);
```

### 2.2 Flows (Multi-Track Journeys)

From the multitrack architecture doc - flows are first-class:

```sql
-- User journey flows within a site
CREATE TABLE site_flows (
                            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                            site_id UUID REFERENCES sites(id) ON DELETE CASCADE,
                            name VARCHAR(255) NOT NULL,
                            slug VARCHAR(100) NOT NULL,
                            audience_segment VARCHAR(255),
                            narrative_arc JSONB, -- stages with voice/tone parameters
                            entry_points TEXT[],
                            success_metric TEXT,
                            voice_parameters JSONB DEFAULT '{}', -- formality, technical_depth, etc.
                            is_default BOOLEAN DEFAULT false,
                            created_at TIMESTAMP DEFAULT NOW(),
                            updated_at TIMESTAMP DEFAULT NOW(),

                            UNIQUE(site_id, slug)
);
```

### 2.3 Pages (with Flow Membership)

```sql
-- Pages belong to sites, can be in multiple flows
CREATE TABLE pages (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       site_id UUID REFERENCES sites(id) ON DELETE CASCADE,
                       name VARCHAR(255) NOT NULL, -- slug: "about", "services/consulting"
                       url VARCHAR(500) NOT NULL, -- "/about.html", "/services/consulting.html"
                       title VARCHAR(500),
                       page_type VARCHAR(50), -- index, content, product, legal, landing
                       status VARCHAR(50) DEFAULT 'active',
                       content_hash VARCHAR(64), -- for change detection
                       meta_description TEXT,
                       topics TEXT[], -- for semantic queries, future vector indexing
                       last_built_at TIMESTAMP,
                       expires_at TIMESTAMP, -- for campaign pages
                       created_at TIMESTAMP DEFAULT NOW(),
                       updated_at TIMESTAMP DEFAULT NOW(),

                       UNIQUE(site_id, name)
);

-- Pages in flows (many-to-many, with flow-specific context)
CREATE TABLE flow_pages (
                            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                            flow_id UUID REFERENCES site_flows(id) ON DELETE CASCADE,
                            page_id UUID REFERENCES pages(id) ON DELETE CASCADE,
                            stage_in_narrative VARCHAR(100), -- "awareness", "consideration", "conversion"
                            sequence_order INTEGER,
                            context_overrides JSONB DEFAULT '{}', -- voice_formality, urgency, etc.

                            UNIQUE(flow_id, page_id)
);
```

### 2.4 Page Components (Instances of Components on Pages)

This is the bridge between `content_components` (templates) and actual page content:

```sql
-- Component instances on pages
CREATE TABLE page_components (
                                 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                 page_id UUID REFERENCES pages(id) ON DELETE CASCADE,
                                 component_id UUID REFERENCES content_components(id), -- template used
                                 position INTEGER NOT NULL, -- order on page
                                 slot_name VARCHAR(100), -- if nested in parent component
                                 parent_component_instance_id UUID REFERENCES page_components(id), -- for nesting

    -- Rendered content
                                 rendered_html TEXT, -- actual HTML with data filled in
                                 content_data JSONB, -- data that was passed to template
                                 content_hash VARCHAR(64), -- for change detection

    -- Semantic addressing
                                 data_path VARCHAR(500), -- e.g., "page.section[2].grid.slot[left]"
                                 data_uuid UUID DEFAULT gen_random_uuid(), -- unique ID for editing

                                 created_at TIMESTAMP DEFAULT NOW(),
                                 updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_page_components_page ON page_components(page_id);
CREATE INDEX idx_page_components_template ON page_components(component_id);
```

### 2.5 Link Registry (Index, Not Source)

Links are extracted FROM page_components, not stored separately. This table is an INDEX for querying:

```sql
-- Link index (derived from page_components, kept in sync)
CREATE TABLE link_registry (
                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Source (where the link appears)
                               source_component_instance_id UUID REFERENCES page_components(id) ON DELETE CASCADE,
                               source_page_id UUID REFERENCES pages(id) ON DELETE CASCADE,
                               source_site_id UUID REFERENCES sites(id) ON DELETE CASCADE,

    -- Target
                               target_url VARCHAR(1000) NOT NULL,
                               target_page_id UUID REFERENCES pages(id), -- if internal, resolved
                               target_site_id UUID REFERENCES sites(id), -- if cross-site internal

    -- Classification
                               scope VARCHAR(50) NOT NULL, -- internal, page, site, network, external
                               link_type VARCHAR(50) NOT NULL, -- navigation, content, semantic, affiliate, reference

    -- Metadata
                               anchor_text VARCHAR(500),
                               rel_attr VARCHAR(100), -- nofollow, sponsored, ugc

    -- For affiliates
                               affiliate_provider VARCHAR(100),
                               affiliate_tag VARCHAR(255),

    -- Status & health
                               status VARCHAR(50) DEFAULT 'active',
                               last_validated_at TIMESTAMP,
                               validation_result VARCHAR(50), -- ok, broken, timeout, redirect

                               created_at TIMESTAMP DEFAULT NOW(),
                               updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX idx_links_source_component ON link_registry(source_component_instance_id);
CREATE INDEX idx_links_source_page ON link_registry(source_page_id);
CREATE INDEX idx_links_target_page ON link_registry(target_page_id);
CREATE INDEX idx_links_type ON link_registry(link_type);
CREATE INDEX idx_links_scope ON link_registry(scope);
CREATE INDEX idx_links_broken ON link_registry(validation_result)
    WHERE validation_result != 'ok';
```

### 2.6 Navigation Structures (Aggregated from Components)

Navigation is site-level, but references page_components:

```sql
-- Pre-computed navigation structures
CREATE TABLE navigation_structures (
                                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                       site_id UUID REFERENCES sites(id) ON DELETE CASCADE,
                                       nav_type VARCHAR(50) NOT NULL, -- header, footer, mobile, sidebar

    -- The structure (references pages)
                                       structure JSONB NOT NULL,
    /* Example:
    {
      "items": [
        {"page_id": "uuid", "label": "Home", "url": "/index.html", "children": []},
        {"page_id": "uuid", "label": "Services", "url": "/services.html", "children": [
          {"page_id": "uuid", "label": "Consulting", "url": "/services/consulting.html"}
        ]}
      ]
    }
    */

    -- Versioning
                                       version INTEGER DEFAULT 1,
                                       is_current BOOLEAN DEFAULT true,

                                       created_at TIMESTAMP DEFAULT NOW(),

                                       UNIQUE(site_id, nav_type, version)
);
```

### 2.7 Redirects

```sql
CREATE TABLE redirects (
                           id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                           site_id UUID REFERENCES sites(id) ON DELETE CASCADE,
                           from_url VARCHAR(500) NOT NULL,
                           to_url VARCHAR(500) NOT NULL,
                           redirect_type INTEGER DEFAULT 301, -- 301, 302, 307, 410
                           reason VARCHAR(255),
                           source_page_id UUID REFERENCES pages(id), -- if we know which page moved
                           hit_count INTEGER DEFAULT 0,
                           created_at TIMESTAMP DEFAULT NOW(),
                           expires_at TIMESTAMP,

                           UNIQUE(site_id, from_url)
);
```

---

## Part 3: Using the Existing `relationships` Table

The `relationships` table is PERFECT for semantic links. We use it directly:

```sql
-- Example: Link blog post to its pillar page
INSERT INTO relationships (
    source_entity_id, source_entity_type,
    target_entity_id, target_entity_type,
    relationship_type,
    properties
) VALUES (
             'page-uuid-123', 'page',
             'pillar-page-uuid', 'page',
             'cluster_to_pillar',
             '{"relevance_score": 0.85, "auto_generated": true}'::jsonb
         );

-- Example: Link pages in same topic cluster
INSERT INTO relationships (
    source_entity_id, source_entity_type,
    target_entity_id, target_entity_type,
    relationship_type,
    properties
) VALUES (
             'page-uuid-123', 'page',
             'page-uuid-456', 'page',
             'related_content',
             '{"cluster_id": "topic-cluster-uuid", "relevance_score": 0.72}'::jsonb
         );

-- Example: Cross-site relationship (within network)
INSERT INTO relationships (
    source_entity_id, source_entity_type,
    target_entity_id, target_entity_type,
    relationship_type,
    properties
) VALUES (
             'page-uuid-123', 'page',
             'other-site-page-uuid', 'page',
             'cross_site_reference',
             '{"source_site": "site-uuid-1", "target_site": "site-uuid-2"}'::jsonb
         );
```

### Querying Semantic Relationships

```sql
-- Find all cluster pages for a pillar
SELECT p.* FROM pages p
                    JOIN relationships r ON r.source_entity_id = p.id::text
WHERE r.target_entity_id = 'pillar-page-uuid'
  AND r.relationship_type = 'cluster_to_pillar'
  AND r.status = 'active';

-- Find related content for a page
SELECT p.*, r.properties->>'relevance_score' as relevance
FROM pages p
    JOIN relationships r ON r.target_entity_id = p.id::text
WHERE r.source_entity_id = 'current-page-uuid'
  AND r.relationship_type = 'related_content'
  AND r.status = 'active'
ORDER BY (r.properties->>'relevance_score')::float DESC
    LIMIT 5;
```

---

## Part 4: Component Change Detection & Patching

### How Patch Updates Work

```
1. Component template changes (content_components)
   └─► Find all page_components using that template
   └─► Re-render those instances only
   └─► Update rendered_html and content_hash
   └─► Extract links, update link_registry
   └─► Rebuild only affected pages

2. Component instance data changes (page_components.content_data)
   └─► Re-render that single instance
   └─► Update rendered_html and content_hash
   └─► Extract links, update link_registry
   └─► Rebuild only that page

3. Navigation change (page added/removed/renamed)
   └─► Update navigation_structures
   └─► Find all page_components with function='navigation-*'
   └─► Re-render those components on ALL pages
   └─► Or: navigation components pull from navigation_structures at render time
```

### Content Hash for Change Detection

```sql
-- Function to detect what changed
CREATE OR REPLACE FUNCTION detect_component_changes(p_page_id UUID)
RETURNS TABLE(component_instance_id UUID, change_type TEXT) AS $$
BEGIN
RETURN QUERY
SELECT
    pc.id,
    CASE
        WHEN pc.content_hash != md5(pc.rendered_html) THEN 'content_modified'
        WHEN pc.rendered_html IS NULL THEN 'needs_render'
        ELSE 'unchanged'
        END as change_type
FROM page_components pc
WHERE pc.page_id = p_page_id;
END;
$$ LANGUAGE plpgsql;
```

---

## Part 5: Link Extraction from Components

When a component is rendered, we extract links:

```go
// ExtractLinksFromHTML parses rendered HTML and extracts all links
func ExtractLinksFromHTML(html string, sourceComponent PageComponent) []LinkRegistryEntry {
doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
var links []LinkRegistryEntry

doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
href, _ := s.Attr("href")
anchorText := s.Text()
rel, _ := s.Attr("rel")

link := LinkRegistryEntry{
SourceComponentInstanceID: sourceComponent.ID,
SourcePageID:              sourceComponent.PageID,
SourceSiteID:              sourceComponent.SiteID,
TargetURL:                 href,
AnchorText:                anchorText,
RelAttr:                   rel,
Scope:                     classifyLinkScope(href, sourceComponent.SiteID),
LinkType:                  classifyLinkType(sourceComponent.Function),
}

// Resolve internal links
if link.Scope == "page" || link.Scope == "site" {
link.TargetPageID = resolvePageID(href, sourceComponent.SiteID)
}

links = append(links, link)
})

return links
}

// classifyLinkType based on component function
func classifyLinkType(componentFunction string) string {
switch {
case strings.Contains(componentFunction, "navigation"):
return "navigation"
case strings.Contains(componentFunction, "cta"):
return "content"
case strings.Contains(componentFunction, "related"):
return "semantic"
case strings.Contains(componentFunction, "affiliate"):
return "affiliate"
default:
return "content"
}
}
```

### Sync Link Registry After Render

```go
// SyncLinksForComponent updates link_registry after component render
func SyncLinksForComponent(ctx context.Context, db *sql.DB, componentInstanceID uuid.UUID, newLinks []LinkRegistryEntry) error {
tx, _ := db.BeginTx(ctx, nil)
defer tx.Rollback()

// Delete old links for this component
_, err := tx.ExecContext(ctx,
`DELETE FROM link_registry WHERE source_component_instance_id = $1`,
componentInstanceID)
if err != nil {
return err
}

// Insert new links
for _, link := range newLinks {
_, err := tx.ExecContext(ctx, `
            INSERT INTO link_registry (
                source_component_instance_id, source_page_id, source_site_id,
                target_url, target_page_id, target_site_id,
                scope, link_type, anchor_text, rel_attr,
                status
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'active')
        `, link.SourceComponentInstanceID, link.SourcePageID, link.SourceSiteID,
link.TargetURL, link.TargetPageID, link.TargetSiteID,
link.Scope, link.LinkType, link.AnchorText, link.RelAttr)
if err != nil {
return err
}
}

return tx.Commit()
}
```

---

## Part 6: Navigation Agent Integration

The Navigation Agent doesn't store links separately - it builds navigation from pages:

```go
// BuildNavigationStructure creates nav from pages table
func BuildNavigationStructure(ctx context.Context, db *sql.DB, siteID uuid.UUID) (*NavigationStructure, error) {
// Get all active pages with nav metadata
rows, err := db.QueryContext(ctx, `
        SELECT id, name, url, title, page_type,
               COALESCE((SELECT properties->>'nav_label' FROM page_metadata WHERE page_id = p.id), title) as nav_label,
               COALESCE((SELECT (properties->>'nav_order')::int FROM page_metadata WHERE page_id = p.id), 999) as nav_order,
               COALESCE((SELECT (properties->>'in_header')::boolean FROM page_metadata WHERE page_id = p.id), true) as in_header,
               COALESCE((SELECT (properties->>'in_footer')::boolean FROM page_metadata WHERE page_id = p.id), true) as in_footer
        FROM pages p
        WHERE site_id = $1 AND status = 'active'
        ORDER BY nav_order
    `, siteID)

// Build structure...
// Store in navigation_structures table
// Return for use in component rendering
}
```

When rendering a navigation component, it receives the pre-built structure:

```go
// RenderNavigationComponent uses pre-built nav structure
func RenderNavigationComponent(template string, navStructure NavigationStructure, currentPageID uuid.UUID) string {
data := map[string]interface{}{
"items":       navStructure.Items,
"currentPage": currentPageID,
}
return renderTemplate(template, data)
}
```

---

## Part 7: MVP Implementation Path

### Phase 1: Database Foundation (Now)

```sql
-- Run these migrations
1. clients (if not exists)
2. networks  
3. sites
4. site_flows
5. pages
6. flow_pages
7. page_components
8. link_registry
9. navigation_structures
10. redirects

-- relationships table already exists, just use it
```

### Phase 2: Basic Link Resolution (Now)

For multipage-website-builder to work:

1. **When strategist outputs pages**: Insert into `pages` table
2. **When building navigation**: Query `pages`, build structure, store in `navigation_structures`
3. **When html-developer renders**:
    - Receive nav structure as input
    - Use relative URLs from pages table
    - After render: extract links, store in `link_registry`

### Phase 3: Component Tracking (Soon)

1. **Track component instances**: Insert into `page_components` when assembling page
2. **Store rendered HTML**: For patch updates
3. **Content hash**: For change detection

### Phase 4: Semantic Relationships (Later)

1. **Use relationships table**: For pillar↔cluster, related content
2. **Topic extraction**: Populate `pages.topics`
3. **Auto-suggest**: Query relationships for "related content" component

### Phase 5: Cross-Network & Affiliates (Later)

1. **Network-level queries**: Find pages across sites in same network
2. **Affiliate configuration**: Per-network settings
3. **Cross-site link tracking**: Same client, different networks

---

## Part 8: Integration with Existing Code

### In html_actions.go

The `extractSitemapInfo` function we already modified would change to:

```go
// extractNavigationFromDB gets pre-built navigation
func extractNavigationFromDB(ctx context.Context, db *sql.DB, siteID uuid.UUID) (*NavigationStructure, error) {
var structure NavigationStructure
err := db.QueryRowContext(ctx, `
        SELECT structure FROM navigation_structures
        WHERE site_id = $1 AND nav_type = 'header' AND is_current = true
    `, siteID).Scan(&structure)
return &structure, err
}

// Fallback: if not in DB, build from context (current behavior)
func extractSitemapInfo(context map[string]interface{}) string {
// First try DB
if siteID := extractSiteID(context); siteID != uuid.Nil {
if nav, err := extractNavigationFromDB(ctx, db, siteID); err == nil {
return formatNavigationForPrompt(nav)
}
}

// Fallback to context-based (existing code)
// ... existing implementation ...
}
```

### In multipage_actions.go

When assembling site:

```go
// After building all pages, sync to database
func syncSiteToDatabase(ctx context.Context, db *sql.DB, siteID uuid.UUID, pages []PageResult) error {
for _, page := range pages {
// Upsert page
pageID, err := upsertPage(ctx, db, siteID, page)

// Track components
for _, component := range page.Components {
componentInstanceID, err := upsertPageComponent(ctx, db, pageID, component)

// Extract and sync links
links := ExtractLinksFromHTML(component.RenderedHTML, component)
err = SyncLinksForComponent(ctx, db, componentInstanceID, links)
}
}

// Build navigation structure
nav, err := BuildNavigationStructure(ctx, db, siteID)
err = upsertNavigationStructure(ctx, db, siteID, "header", nav)

return nil
}
```

---

## Summary

**Integration approach:**
- Links live IN components (page_components.rendered_html)
- link_registry is an INDEX, synced after render
- relationships table handles semantic links (already exists!)
- navigation_structures aggregates page links for nav components

**Key tables:**
| Table | Purpose | Exists? |
|-------|---------|---------|
| content_components | Templates | ✅ Yes |
| relationships | Semantic links | ✅ Yes (empty) |
| clients | Multi-tenant | ❌ Add |
| networks | Site groups | ❌ Add |
| sites | Domains | ❌ Add |
| site_flows | User journeys | ❌ Add |
| pages | URLs | ❌ Add |
| flow_pages | Page↔Flow | ❌ Add |
| page_components | Instances | ❌ Add |
| link_registry | Link index | ❌ Add |
| navigation_structures | Nav cache | ❌ Add |
| redirects | URL redirects | ❌ Add |

**Why this works:**
1. Components remain source of truth for content (including links)
2. Link registry enables queries without parsing HTML every time
3. Existing relationships table handles semantic connections
4. Patch updates work: change component → re-extract links → update registry
5. Client isolation via client_id → network_id → site_id chain