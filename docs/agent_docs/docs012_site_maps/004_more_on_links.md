# Link Management Architecture: Multi-Network Scale

## Context

This document explores the architecture needed for link management at scale:
- Multiple networks of sites (not just sites, but networks of sites)
- 1000s of sites, 10000s+ pages
- Variable update frequencies (small edits to full rebuilds)
- Component-level updates (replace affiliate blocks across many pages)
- GitHub as current source of truth, database for metadata/links
- Future: Vector storage for semantic retrieval and manipulation

The goal is to start simple but not design into a corner.

---

## Part 1: Entity Hierarchy

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           ORGANIZATION                                   │
│                     (top-level owner/account)                            │
└─────────────────────────────────┬───────────────────────────────────────┘
                                  │
              ┌───────────────────┼───────────────────┐
              │                   │                   │
              ▼                   ▼                   ▼
       ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
       │   NETWORK    │    │   NETWORK    │    │   NETWORK    │
       │  "Recipes"   │    │  "Finance"   │    │  "Reviews"   │
       └──────┬───────┘    └──────┬───────┘    └──────┬───────┘
              │                   │                   │
    ┌─────────┼─────────┐        ...                 ...
    │         │         │
    ▼         ▼         ▼
┌───────┐ ┌───────┐ ┌───────┐
│ SITE  │ │ SITE  │ │ SITE  │
│italian│ │mexican│ │dessert│
│recipes│ │recipes│ │recipes│
└───┬───┘ └───┬───┘ └───┬───┘
    │         │         │
    ▼         ▼         ▼
 [pages]   [pages]   [pages]
```

### Why Networks Matter for Links

Sites within a network can:
- Cross-link freely (we control both ends)
- Share navigation elements ("Our other recipe sites")
- Share affiliate configurations
- Have coordinated updates (change affiliate across network)

Sites across networks might:
- Occasionally cross-link (recipes → kitchen tools review site)
- Share affiliate providers but different configurations
- Be updated independently

---

## Part 2: Content Decomposition

### The Component Model

Pages aren't monolithic - they're composed of components:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              PAGE                                        │
│                         /pasta-recipes                                   │
├─────────────────────────────────────────────────────────────────────────┤
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │ COMPONENT: header-nav                                              │ │
│  │ Type: navigation | Links: [home, recipes, about, contact]         │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │ COMPONENT: hero                                                    │ │
│  │ Type: content | Links: none                                        │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │ COMPONENT: recipe-grid                                             │ │
│  │ Type: content | Links: [/spaghetti, /lasagna, /carbonara]         │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │ COMPONENT: affiliate-cta                                           │ │
│  │ Type: affiliate | Links: [amazon.com/pasta-maker?tag=xxx]         │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │ COMPONENT: related-articles                                        │ │
│  │ Type: semantic | Links: [/italian-cooking, /pasta-sauces]         │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │ COMPONENT: footer-nav                                              │ │
│  │ Type: navigation | Links: [privacy, terms, sitemap]               │ │
│  └────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘
```

### Why Components Matter for Updates

**Scenario: Change affiliate provider from Amazon to Walmart**

Without components:
- Find all pages with affiliate links
- Regenerate entire page
- Rebuild and deploy

With components:
- Query: all components where type=affiliate AND links contain "amazon.com"
- Update just those components
- Patch pages (replace component, keep rest)
- Deploy changed files only

**Scenario: Update navigation across entire network**

Without components:
- Regenerate all pages on all sites
- Full rebuild

With components:
- Update navigation component template
- Query: all components where type=navigation
- Re-render navigation components only
- Patch pages

---

## Part 3: Link Classification (Refined)

```
LINK
├── scope
│   ├── internal (same page anchor)
│   ├── page (same site, different page)
│   ├── site (same network, different site)
│   ├── network (different network, we control)
│   └── external (we don't control)
│
├── type
│   ├── navigation (structural, site-wide)
│   ├── content (editorial, in-page)
│   ├── semantic (topic relationship)
│   ├── affiliate (tracked, revenue)
│   └── reference (citation, source)
│
├── lifecycle
│   ├── permanent (unlikely to change)
│   ├── campaign (planned expiration)
│   ├── dynamic (changes based on context)
│   └── conditional (A/B test, personalization)
│
└── status
    ├── active
    ├── scheduled (future activation)
    ├── expired (no longer valid)
    ├── redirected (points elsewhere now)
    └── broken (target doesn't exist)
```

### Link Resolution at Different Scopes

```
When rendering a page, links resolve differently:

INTERNAL: #section-name
  → Anchor in same page
  → No validation needed (component handles)

PAGE: /about.html
  → Same site
  → Validate in site's link registry

SITE: https://mexican-recipes.com/tacos.html
  → Same network, different site
  → Resolve via network link index
  → Can verify we control target

NETWORK: https://kitchen-tools-review.com/pasta-makers
  → Different network, but ours
  → Resolve via organization link index
  → Can verify we control target

EXTERNAL: https://amazon.com/...
  → External
  → Cannot guarantee availability
  → Should monitor health
```

---

## Part 4: Database Schema (Expanded)

```sql
-- ============================================================================
-- ORGANIZATION & NETWORK STRUCTURE
-- ============================================================================

CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE networks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    settings JSONB DEFAULT '{}', -- network-wide affiliate config, etc.
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(organization_id, slug)
);

CREATE TABLE sites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    network_id UUID REFERENCES networks(id),
    domain VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    github_repo VARCHAR(500),
    github_branch VARCHAR(100) DEFAULT 'main',
    settings JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'active', -- active, paused, archived
    last_built_at TIMESTAMP,
    last_deployed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(domain)
);

-- ============================================================================
-- PAGE & COMPONENT STRUCTURE
-- ============================================================================

CREATE TABLE pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID REFERENCES sites(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL, -- slug, e.g., "about", "recipes/pasta"
    url VARCHAR(500) NOT NULL, -- full path, e.g., "/about.html"
    title VARCHAR(500),
    meta_description TEXT,
    page_type VARCHAR(50), -- index, content, product, legal, etc.
    status VARCHAR(50) DEFAULT 'active', -- active, draft, archived, redirected
    content_hash VARCHAR(64), -- for change detection
    topics TEXT[], -- for semantic queries
    last_content_update TIMESTAMP,
    last_built_at TIMESTAMP,
    expires_at TIMESTAMP, -- for campaign pages
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(site_id, name)
);

CREATE TABLE page_components (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID REFERENCES pages(id) ON DELETE CASCADE,
    component_type VARCHAR(100) NOT NULL, -- header-nav, hero, cta, affiliate-block, etc.
    position INTEGER NOT NULL, -- order on page
    content JSONB, -- structured content data
    content_hash VARCHAR(64), -- for change detection
    template_id UUID, -- reference to component template if using one
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Index for finding components by type across sites
CREATE INDEX idx_components_type ON page_components(component_type);

-- ============================================================================
-- LINK REGISTRY
-- ============================================================================

CREATE TABLE links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Source (where the link appears)
    source_component_id UUID REFERENCES page_components(id) ON DELETE CASCADE,
    source_page_id UUID REFERENCES pages(id) ON DELETE CASCADE,
    source_site_id UUID REFERENCES sites(id) ON DELETE CASCADE,
    
    -- Target (where the link points)
    target_url VARCHAR(1000) NOT NULL,
    target_page_id UUID REFERENCES pages(id), -- if internal and resolvable
    target_site_id UUID REFERENCES sites(id), -- if cross-site and resolvable
    
    -- Classification
    scope VARCHAR(50) NOT NULL, -- internal, page, site, network, external
    link_type VARCHAR(50) NOT NULL, -- navigation, content, semantic, affiliate, reference
    
    -- Metadata
    anchor_text VARCHAR(500),
    title_attr VARCHAR(500),
    rel_attr VARCHAR(100), -- nofollow, sponsored, etc.
    
    -- Tracking (for affiliates)
    affiliate_provider VARCHAR(100),
    affiliate_tag VARCHAR(255),
    
    -- Status
    status VARCHAR(50) DEFAULT 'active',
    last_checked_at TIMESTAMP,
    last_check_result VARCHAR(50), -- ok, timeout, 404, 500, etc.
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX idx_links_source_page ON links(source_page_id);
CREATE INDEX idx_links_source_site ON links(source_site_id);
CREATE INDEX idx_links_target_page ON links(target_page_id);
CREATE INDEX idx_links_type ON links(link_type);
CREATE INDEX idx_links_scope ON links(scope);
CREATE INDEX idx_links_affiliate ON links(affiliate_provider) WHERE affiliate_provider IS NOT NULL;

-- ============================================================================
-- NAVIGATION STRUCTURES (site-level, pre-computed)
-- ============================================================================

CREATE TABLE navigation_structures (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID REFERENCES sites(id) ON DELETE CASCADE,
    nav_type VARCHAR(50) NOT NULL, -- header, footer, mobile, breadcrumb
    structure JSONB NOT NULL, -- the navigation tree
    version INTEGER DEFAULT 1,
    is_current BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(site_id, nav_type, version)
);

-- ============================================================================
-- REDIRECTS
-- ============================================================================

CREATE TABLE redirects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID REFERENCES sites(id) ON DELETE CASCADE,
    from_url VARCHAR(500) NOT NULL,
    to_url VARCHAR(500) NOT NULL,
    redirect_type INTEGER DEFAULT 301, -- 301, 302, 307, 410
    reason VARCHAR(255),
    source_page_id UUID REFERENCES pages(id), -- original page if known
    hit_count INTEGER DEFAULT 0, -- track usage
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    
    UNIQUE(site_id, from_url)
);

-- ============================================================================
-- SEMANTIC RELATIONSHIPS (for topic clusters)
-- ============================================================================

CREATE TABLE topic_clusters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID REFERENCES sites(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    pillar_page_id UUID REFERENCES pages(id),
    keywords TEXT[],
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE semantic_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_page_id UUID REFERENCES pages(id) ON DELETE CASCADE,
    target_page_id UUID REFERENCES pages(id) ON DELETE CASCADE,
    relationship_type VARCHAR(50), -- pillar_to_cluster, cluster_to_pillar, related, series
    relevance_score FLOAT,
    cluster_id UUID REFERENCES topic_clusters(id),
    auto_generated BOOLEAN DEFAULT true,
    approved BOOLEAN DEFAULT false, -- human review
    created_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(source_page_id, target_page_id, relationship_type)
);

-- ============================================================================
-- AFFILIATE CONFIGURATION
-- ============================================================================

CREATE TABLE affiliate_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id),
    name VARCHAR(100) NOT NULL, -- amazon, walmart, etc.
    base_url VARCHAR(500),
    tag_parameter VARCHAR(100), -- tag, irclickid, etc.
    default_tag VARCHAR(255),
    settings JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE affiliate_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID REFERENCES affiliate_providers(id),
    network_id UUID REFERENCES networks(id), -- network-specific tag
    site_id UUID REFERENCES sites(id), -- site-specific tag (optional)
    product_id VARCHAR(255), -- external product identifier
    product_url VARCHAR(1000),
    tracking_tag VARCHAR(255),
    display_name VARCHAR(255),
    last_checked_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- ============================================================================
-- AUDIT LOG
-- ============================================================================

CREATE TABLE link_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL, -- page, link, redirect, navigation
    entity_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL, -- created, updated, deleted, redirected
    old_value JSONB,
    new_value JSONB,
    performed_by VARCHAR(255),
    reason TEXT,
    performed_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_audit_entity ON link_audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_time ON link_audit_log(performed_at);

-- ============================================================================
-- FUTURE: VECTOR STORAGE PREPARATION
-- ============================================================================
-- When we add vector storage, we'll need:
-- 1. Embeddings for pages (topic extraction)
-- 2. Embeddings for components (semantic matching)
-- 3. Similarity queries for related content
--
-- This could be:
-- - pgvector extension in PostgreSQL
-- - External vector DB (Pinecone, Weaviate, etc.)
-- - Hybrid (metadata in Postgres, vectors elsewhere)
--
-- The pages.topics TEXT[] field is a placeholder for extracted topics
-- that will eventually be vectorized.

-- Example pgvector preparation (uncomment when ready):
-- CREATE EXTENSION IF NOT EXISTS vector;
-- ALTER TABLE pages ADD COLUMN embedding vector(1536);
-- ALTER TABLE page_components ADD COLUMN embedding vector(1536);
-- CREATE INDEX idx_pages_embedding ON pages USING ivfflat (embedding vector_cosine_ops);
```

---

## Part 5: Update Operations

### Operation Types

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        UPDATE OPERATIONS                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  COMPONENT-LEVEL                                                        │
│  ├── Replace affiliate block with new provider                          │
│  ├── Update CTA text/link across pages                                  │
│  └── Regenerate navigation components                                   │
│                                                                         │
│  PAGE-LEVEL                                                             │
│  ├── Add new page                                                       │
│  ├── Edit page content                                                  │
│  ├── Rename page (URL change)                                           │
│  ├── Delete page                                                        │
│  └── Merge pages                                                        │
│                                                                         │
│  SITE-LEVEL                                                             │
│  ├── Rebuild navigation                                                 │
│  ├── Regenerate sitemap                                                 │
│  ├── Update site-wide settings                                          │
│  └── Full site rebuild                                                  │
│                                                                         │
│  NETWORK-LEVEL                                                          │
│  ├── Update cross-site links                                            │
│  ├── Change affiliate provider for network                              │
│  └── Propagate template changes                                         │
│                                                                         │
│  ORGANIZATION-LEVEL                                                     │
│  ├── Cross-network link updates                                         │
│  └── Global affiliate changes                                           │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Update Flow

```
UPDATE REQUEST
     │
     ▼
┌──────────────┐
│ DETERMINE    │
│ SCOPE        │
│ (component/  │
│  page/site/  │
│  network)    │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ FIND         │
│ AFFECTED     │
│ ENTITIES     │
│ (SQL query)  │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ COMPUTE      │
│ LINK         │
│ CHANGES      │
│ (what breaks,│
│  what's new) │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ GENERATE     │
│ REDIRECTS    │
│ (if URLs     │
│  changed)    │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ UPDATE       │
│ COMPONENTS   │
│ (patch, not  │
│  regenerate) │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ REBUILD      │
│ AFFECTED     │
│ PAGES        │
│ (assemble    │
│  components) │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ UPDATE       │
│ NAVIGATION   │
│ (if pages    │
│  added/      │
│  removed)    │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ COMMIT TO    │
│ GITHUB       │
│ (changed     │
│  files only) │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ UPDATE       │
│ DATABASE     │
│ (links,      │
│  hashes,     │
│  timestamps) │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ DEPLOY       │
│ (if auto-    │
│  deploy on)  │
└──────────────┘
```

---

## Part 6: Agent Architecture (Revised)

### Multi-Level Agent Group

```
┌─────────────────────────────────────────────────────────────────────────┐
│              ORGANIZATION LINK COORDINATOR                               │
│              (handles cross-network operations)                          │
├─────────────────────────────────────────────────────────────────────────┤
│  • Cross-network link management                                        │
│  • Organization-wide affiliate configuration                            │
│  • Global link health monitoring                                        │
└─────────────────────────────────┬───────────────────────────────────────┘
                                  │
          ┌───────────────────────┼───────────────────────┐
          │                       │                       │
          ▼                       ▼                       ▼
┌─────────────────────┐ ┌─────────────────────┐ ┌─────────────────────┐
│  NETWORK LINK       │ │  NETWORK LINK       │ │  NETWORK LINK       │
│  ORCHESTRATOR       │ │  ORCHESTRATOR       │ │  ORCHESTRATOR       │
│  (recipes network)  │ │  (finance network)  │ │  (reviews network)  │
├─────────────────────┤ ├─────────────────────┤ ├─────────────────────┤
│  • Cross-site links │ │                     │ │                     │
│  • Network nav      │ │                     │ │                     │
│  • Affiliate config │ │                     │ │                     │
└─────────┬───────────┘ └─────────────────────┘ └─────────────────────┘
          │
    ┌─────┴─────┐
    │           │
    ▼           ▼
┌─────────┐ ┌─────────┐
│ SITE    │ │ SITE    │     (per-site link management)
│ LINK    │ │ LINK    │
│ MANAGER │ │ MANAGER │
└────┬────┘ └────┬────┘
     │           │
     ▼           ▼
  [agents]    [agents]
```

### Site-Level Link Manager (Detail)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    SITE LINK MANAGER                                     │
│                    (orchestrates site link operations)                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  INPUT: Site ID + Operation (build/update/patch)                        │
│  OUTPUT: Updated pages, link registry, navigation, SEO artifacts        │
│                                                                         │
└───────────────────────────────┬─────────────────────────────────────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        │                       │                       │
        ▼                       ▼                       ▼
┌───────────────┐      ┌───────────────┐      ┌───────────────┐
│  PAGE LINK    │      │  NAVIGATION   │      │  SEO LINK     │
│  RESOLVER     │      │  BUILDER      │      │  GENERATOR    │
├───────────────┤      ├───────────────┤      ├───────────────┤
│ • Resolve all │      │ • Header nav  │      │ • Sitemap.xml │
│   links in    │      │ • Footer nav  │      │ • Robots.txt  │
│   page to     │      │ • Mobile menu │      │ • Canonical   │
│   URLs        │      │ • Breadcrumbs │      │ • Schema.org  │
│ • Validate    │      │               │      │               │
│   targets     │      │               │      │               │
└───────────────┘      └───────────────┘      └───────────────┘
        │                       │                       │
        │                       │                       │
        ▼                       ▼                       ▼
┌───────────────┐      ┌───────────────┐      ┌───────────────┐
│  AFFILIATE    │      │  SEMANTIC     │      │  LIFECYCLE    │
│  LINK         │      │  LINK         │      │  MANAGER      │
│  RESOLVER     │      │  ANALYZER     │      │               │
├───────────────┤      ├───────────────┤      ├───────────────┤
│ • Apply       │      │ • Topic       │      │ • Redirects   │
│   tracking    │      │   extraction  │      │ • Expiration  │
│   tags        │      │ • Related     │      │ • Archiving   │
│ • Provider    │      │   content     │      │ • Audit log   │
│   lookup      │      │ • Cluster map │      │               │
└───────────────┘      └───────────────┘      └───────────────┘
```

---

## Part 7: MVP vs Future

### MVP (Phase 1)

**Goal**: Fix current navigation problem, lay database foundation

```
Implement:
├── Database schema (full schema, but only use subset)
├── Site Link Manager (orchestrator)
│   ├── Navigation Builder (header/footer)
│   ├── Page Link Resolver (internal links)
│   └── SEO Link Generator (sitemap, robots)
├── Basic operations:
│   ├── Full site build
│   ├── Add page
│   └── Delete page (with redirect)
└── Integration:
    ├── Read from strategist output
    ├── Write to database
    └── Output to GitHub
```

**Not in MVP**:
- Network-level coordination
- Affiliate link management
- Semantic link analysis
- Component-level patching
- Vector storage

### Phase 2: Persistence & Updates

```
Add:
├── Page rename with redirect cascade
├── Component tracking
├── Change detection (content hashes)
├── Incremental updates (patch changed components)
└── Audit logging
```

### Phase 3: Network Operations

```
Add:
├── Network Link Orchestrator
├── Cross-site link management
├── Network navigation ("Our other sites")
└── Network-wide updates
```

### Phase 4: Affiliates

```
Add:
├── Affiliate provider configuration
├── Affiliate link resolution
├── Tracking tag management
├── Link rotation (A/B testing)
└── Conversion tracking integration
```

### Phase 5: Semantic Intelligence

```
Add:
├── Topic extraction (AI)
├── Vector storage integration
├── Cluster detection
├── Related content suggestions
├── Gap analysis
└── Semantic link auto-generation
```

---

## Part 8: Key Design Decisions

### 1. Component as First-Class Entity

By tracking components separately from pages, we enable:
- Component-level updates without page regeneration
- Template changes that propagate to all instances
- Fine-grained affiliate management
- Better change tracking

**Decision**: Components have their own table and identity.

### 2. Links Reference Components, Not Just Pages

A link exists within a component. When we update a component:
- We know exactly which links are affected
- We can validate just those links
- We can track link origin precisely

**Decision**: Links have source_component_id, not just source_page_id.

### 3. Multi-Scope Link Classification

Links have explicit scope (internal, page, site, network, external).
This enables:
- Different validation strategies per scope
- Different monitoring requirements
- Clear upgrade path when external becomes internal

**Decision**: scope is required, validated enum.

### 4. Database + GitHub Hybrid

- **GitHub**: Source of truth for content (HTML files)
- **Database**: Source of truth for metadata (links, structure, config)

When building:
1. Read structure from database
2. Generate content
3. Write files to GitHub
4. Update database with new hashes/timestamps

**Decision**: Neither is complete without the other.

### 5. Prepare for Vectors, Don't Implement Yet

The schema includes topics TEXT[] on pages.
When we add vector storage:
- Extract topics using AI
- Store in topics field
- Generate embeddings
- Store in vector DB (or pgvector)
- Query for semantic relationships

**Decision**: Schema is vector-ready, implementation is deferred.

---

## Part 9: Migration Path

### From Current State

Currently:
- Links generated at build time
- No persistence
- Navigation in strategist output (sitemap field)

Migration:
1. Deploy database schema
2. On each build, write to database (not just generate)
3. Read navigation from database (not sitemap field)
4. Gradually add more link types

### Backward Compatibility

- Strategist can still output sitemap for fallback
- If database empty, use sitemap
- If database populated, prefer database
- Gradual transition, not big bang

---

## Part 10: Open Questions

1. **GitHub as source vs database as source**
    - For now: generate and write both, database is metadata
    - Future: could store content in database, generate to GitHub
    - Or: parse from GitHub to populate database

2. **When to regenerate vs patch**
    - Simple: always regenerate page when any component changes
    - Optimal: patch HTML with new component
    - Trade-off: complexity vs efficiency
    - Start with regenerate, optimize later

3. **Cross-network visibility**
    - Can networks see each other's pages?
    - For linking, probably yes
    - For editing, probably no
    - Need permission model eventually

4. **External link monitoring**
    - How often to check?
    - What to do when broken?
    - Separate service or agent?
    - Probably separate scheduled job

5. **Vector storage choice**
    - pgvector (in PostgreSQL)
    - Pinecone (managed, scalable)
    - Weaviate (open source)
    - Depends on scale and query patterns
    - Defer decision until Phase 5

---

## Summary

The key insight is that links exist at multiple levels (component, page, site, network) and have different types (navigation, content, semantic, affiliate). A single agent can't handle all of this, but a well-designed agent group can.

Starting with:
- Navigation Builder
- Page Link Resolver
- SEO Link Generator
- Database persistence

Gives us a working MVP that solves the immediate problem while setting up for:
- Component-level updates
- Network operations
- Affiliate management
- Semantic intelligence
- Vector-based content relationships

The database schema is designed to support all of this from day one, even though we won't use all tables immediately.


----

This is helpful context. You're describing a content management platform at scale, not just a website builder. Let me think through the architectural implications.## Key Architectural Decisions

### 1. Entity Hierarchy
```
Organization → Network → Site → Page → Component → Link
```
Each level can have its own operations and cross-level linking is tracked explicitly.

### 2. Components as First-Class Entities

Instead of treating pages as monolithic, we track components separately. This enables:
- "Update all affiliate blocks across 500 pages" without regenerating full pages
- "Change CTA text network-wide" as a single operation
- Template changes that propagate

### 3. Link Scope Classification

Every link has an explicit scope:
- **internal**: anchor within page
- **page**: same site
- **site**: same network, different site
- **network**: different network, we control
- **external**: we don't control

This determines validation strategy, monitoring requirements, and update cascades.

### 4. Database + GitHub Hybrid

- **Database**: Metadata, structure, links, relationships
- **GitHub**: Content files (source of truth for deployed HTML)

Neither is complete without the other. Build reads from both, writes to both.

### 5. MVP That Doesn't Box Us In

Start with:
| Now | Later |
|-----|-------|
| Navigation Builder | Network Coordinator |
| Page Link Resolver | Affiliate Manager |
| SEO Generator | Semantic Analyzer |
| Full database schema | Vector storage |
| Full site builds | Component patching |

The schema supports everything from day one, even if we only use a subset initially.

### 6. Open Questions

A few things I flagged that need your input:

1. **Scraping existing sites**: You mentioned pulling down sites. Should we design for parsing HTML back into components? That's a different data flow.

2. **Update granularity**: Start with "regenerate full page when anything changes" or invest in "patch only changed component"?

3. **Permission model**: Can different users/systems edit different networks? Do networks see each other's link registries?

4. **Vector storage**: pgvector (simple, in Postgres) vs external (Pinecone/Weaviate - more scalable)?

Would you like to discuss any of these areas further, or shall we start implementing the MVP schema and the first agents (Navigation Builder, Page Link Resolver)?