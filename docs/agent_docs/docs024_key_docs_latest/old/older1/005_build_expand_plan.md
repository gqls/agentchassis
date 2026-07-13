# 005 — Site Lifecycle: Build, Expand, Market

The unified pipeline for taking a domain from zero to a fully-built, continuously-improving, actively-marketed website.

---

## Core Principle

A new domain build is just a set of work items. So is adding a blog section. So is fixing a broken link. So is launching a Google Ads campaign. So is adopting an existing site. The same queue, same orchestrator, same dispatch loop handles all of them. What changes is: what writes the items, and what handlers exist to process them.

---

## Site Specification System

### The problem it solves

"What should this site be?" gets answered by many sources at different times: the classifier researching a domain, a human providing direction, the briefing agent collecting answers, an adoption agent scraping an existing site, HITL corrections, improvement agents. All of this accumulates as the site's specification — its identity, strategy, tone, visual direction, design, and structure.

Previously this data was scattered across `sites.content_data` (a growing JSONB blob), `collected_data` in orchestration runs (transient), and nowhere persistent at all (lost after orchestration completes). The spec system gives it a proper home.

### Table: `site_specs`

```sql
CREATE TABLE site_specs (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id         uuid NOT NULL REFERENCES sites(id),
    aspect          text NOT NULL,
    data            jsonb NOT NULL,
    
    -- Provenance
    source          text NOT NULL,      -- 'classifier', 'adoption', 'hitl', 
                                        -- 'planner', 'improvement', 'seed', 'manual',
                                        -- 'rollback', 'fork', 'recovery'
    source_agent    text,               -- agent type that wrote this
    source_item_id  uuid,               -- work item that caused this write
    notes           text,               -- human-readable reason
    
    -- Milestones and history
    milestone       text,               -- null for most rows; named for key moments
    is_current      boolean NOT NULL DEFAULT true,
    created_at      timestamptz NOT NULL DEFAULT now(),
    superseded_at   timestamptz,
    superseded_by   uuid REFERENCES site_specs(id),
    
    created_by      text NOT NULL
);

-- One current spec per site per aspect
CREATE UNIQUE INDEX idx_site_specs_current 
ON site_specs (site_id, aspect) 
WHERE is_current = true;

-- Fast lookup: all current specs for a site
CREATE INDEX idx_site_specs_lookup 
ON site_specs (site_id) 
WHERE is_current = true;

-- History queries
CREATE INDEX idx_site_specs_history 
ON site_specs (site_id, aspect, created_at DESC);

-- Find milestones
CREATE INDEX idx_site_specs_milestones
ON site_specs (site_id, milestone)
WHERE milestone IS NOT NULL;
```

### Aspects

Open-ended — new aspects can be added by writing a row with a new name. No migration needed.

| Aspect | Written by | Read by |
|---|---|---|
| `identity` | classifier, adoption, HITL, seed | planner, content writers, briefing |
| `strategy` | classifier, HITL | planner, improvement agent |
| `tone` | classifier, HITL | content writers |
| `visual_direction` | classifier, HITL | webdesign-agent, planner |
| `design` | webdesign-agent | content writers (context), rerender |
| `image_guidance` | classifier | planner, image-generator |
| `structure` | planner, adoption | improvement agent, nav-agent |
| `marketing` | marketing agents | SEM, email, social agents |
| `adoption_source` | adoption agent | reference only — raw scraped data |

### Area and page level

Not in site_specs. Simple JSONB columns on existing tables:

- `site_areas.area_spec` — area-level tone, visual_direction overrides. Most areas: null.
- `pages.page_spec` — content hints, existing content (for adoption), content_direction for rewrites.

When a content writer needs the tone for a page in the /blog area: check area_spec for tone override → if null, read site's tone from site_specs. Two queries at most, simple Go fallback logic. Not a cascade engine.

### How agents use it

**Writing:** After the classifier runs, it calls a `write_site_spec` action:

```go
WriteSiteSpec(siteID, "identity", identityData, "classifier", "domain-research-classifier", workItemID)
```

This marks the previous `identity` row as `is_current = false`, inserts the new row as `is_current = true`.

**Reading:** A content writer starting work loads what it needs:

```sql
SELECT aspect, data FROM site_specs 
WHERE site_id = $1 AND aspect IN ('identity', 'tone') AND is_current = true;
```

Single indexed query, returns 2 small rows.

### Milestones

Most spec writes have `milestone = null`. Key moments get labeled:

| Milestone | When set | Example |
|---|---|---|
| `initial_research` | Classifier completes | First identity, strategy, visual_direction |
| `initial_plan` | Planner completes | Structure, image_guidance finalized |
| `post_build` | All initial build items complete | Design, assets, content all done |
| `adoption_complete` | Adoption agent finishes | Full site recreated |
| `rebrand_q2` | Manual/HITL brand change | Identity, visual_direction updated |

Milestones make rollback targets easy to find: "show me milestones" → "roll back to initial_plan."

---

## Rollback

A spec rollback creates work items. It doesn't undo deployments directly.

### How it works

1. Query specs as they were at a milestone (or timestamp):

```sql
SELECT DISTINCT ON (aspect) aspect, data
FROM site_specs
WHERE site_id = $1 
  AND created_at <= (
    SELECT MAX(created_at) FROM site_specs 
    WHERE site_id = $1 AND milestone = 'initial_plan'
  )
ORDER BY aspect, created_at DESC;
```

2. Diff against current specs — which aspects changed?

3. Write new site_specs rows with the old data, `source = 'rollback'`

4. Create work items based on what changed:
    - identity/tone changed → `needs_content_rewrite` per affected page
    - visual_direction changed → `needs_design` (regenerate CSS)
    - image_guidance changed → `needs_logo`, `needs_hero_image`
    - structure changed → `needs_site_plan` (re-plan, then downstream items)

The handlers don't know it's a rollback. They see "generate CSS matching this visual_direction" and do their job. Each produces a new git commit. The site progressively updates.

### Content rollback (separate concern)

"This page content was wrong, change it back" is not a spec rollback. The spec (tone, identity, direction) didn't change — only the output was wrong. This uses `page_components.content_snapshot` — take the snapshot from a previous version, write a work item: `needs_page_rerender` with the old content_data. The rerender handler produces the page from stored content. Spec stays untouched.

---

## Fork / Copy

Create a new site starting from another site's spec state.

### Full fork

```
direction: {
  "fork_from": "site-uuid-or-domain",
  "fork_time": "2026-02-20T14:00:00Z",  -- optional, defaults to current
  "override": {
    "identity": {"name": "NewSite", "audience_primary": "different audience"}
  }
}
```

The seed action:
1. Creates site record for new domain
2. Copies spec rows from source site (at fork_time if specified)
3. Applies overrides
4. Writes new rows with `source = 'fork'`
5. Seeds work items — typically `needs_site_plan` (specs already populated, skip research + briefing)

### Selective fork

```
direction: {
  "fork_from": "site-uuid-or-domain",
  "copy_aspects": ["tone", "visual_direction", "design"],
  "override": {
    "identity": {"name": "MortgageCompare", "type": "mortgage comparison site"}
  }
}
```

Only copies the specified aspects. Identity, strategy, structure come from the override or from a fresh classifier run.

### Network siblings

design.co.uk, webdesign.uk, website-design.com targeting different markets:

```
-- First site built normally
-- Subsequent sites fork with selective overrides:
direction: {
  "fork_from": "design.co.uk",
  "copy_aspects": ["strategy", "tone", "visual_direction"],
  "override": {
    "identity": {"name": "WebDesign.uk", "audience_primary": "UK startups"}
  }
}
```

Network-wide brand changes: INSERT work items per site, each site processes independently:

```sql
INSERT INTO site_work_items (site_id, item_type, handler_agent, spec, source, ...)
SELECT s.id, 'needs_brand_update', 'brand-update-agent', 
       '{"new_identity": {...}, "reason": "network rebrand"}'::jsonb, 'manual'
FROM sites s WHERE s.network_id = $1;
```

No cascade engine. Just work items.

---

## Site Adoption

Take an existing live site and bring it into the system.

### Direction

```
direction: {"adopt_from": "mortgagecalculator.co.uk"}
```

### What the adoption agent does

1. **Crawl:** Spawns site-scraper in full-site mode — all pages, CSS, assets, metadata
2. **Parse design:** Extract color palette, typography, spacing from CSS
3. **Map components:** LLM analysis maps scraped sections to our component library ("this hero looks like our hero component, this grid is services-grid")
4. **Write specs:** Rows for identity, design, visual_direction, structure, adoption_source
5. **Write page specs:** `page_spec` on each page record — sections with existing content
6. **Write work items:**

```
needs_design          | spec: {adopt_from: design data}        | handler: webdesign-agent
needs_content_page ×N | spec: {mode: "recreate", page_spec: …} | handler: page-content-writer
needs_tool_page       | spec: {mode: "recreate", tool_spec: …} | handler: tool-builder
needs_assets          | spec: {adopt_from: asset list}         | handler: asset-deployer
```

### Handlers don't know they're adopting

A content writer receiving `mode: "recreate"` with `existing_content` in its page_spec uses that content instead of generating from scratch. A webdesign-agent receiving `adopt_from.color_palette` matches those colors. The spec is just richer than usual.

### Adopt then improve

Phase 1: Adopt — recreate the site as-is. All items have `mode: "recreate"`.

Phase 2: Improvement agent compares the adopted site against what a fresh classifier would recommend. Writes items:
- "Hero copy is weak — rewrite"
- "No social proof — add testimonials section"
- "Calculator UX is dated — rebuild with modern patterns"
- "SEO metadata is thin — enhance titles and descriptions"

Same queue, same dispatch, different specs.

---

## Git Spec Snapshots

At milestone moments, a `.site-spec.json` is committed to the site's git repo. This is a lightweight checkpoint that doubles as disaster recovery.

### What it is

One JSON file per site, committed alongside deployed HTML/CSS/images:

```json
{
  "milestone": "post_build",
  "timestamp": "2026-02-24T10:00:00Z",
  "site_id": "uuid",
  "identity": { ... },
  "strategy": { ... },
  "tone": { ... },
  "visual_direction": { ... },
  "design": { ... },
  "structure": { ... }
}
```

### When it's written

Only at milestones — maybe 3-5 times in a site's first week, then rarely. One Go function called by whatever action writes the milestone:

```go
func CommitSpecSnapshot(siteID uuid.UUID, milestone string) {
    // 1. SELECT aspect, data FROM site_specs WHERE site_id = $1 AND is_current = true
    // 2. Marshal to JSON
    // 3. Git commit as .site-spec.json with message "spec: {milestone}"
}
```

One query, one file, one commit.

### Rules

- The DB is always the source of truth. If DB and git disagree, the DB wins.
- The git snapshot is an output, not a source. Agents never read from it during normal operation.
- Its purpose is recovery and human inspection — `git log .site-spec.json` shows milestone history.

---

## Database Backup

Independent of the agent system. The backup process must work even when the cluster is in trouble.

### Backup script

A K8s CronJob (runs every 6 hours, configurable):

1. `pg_dump` the clients_db
2. Compress (gzip)
3. Upload to Backblaze S3 (we already have the S3 adapter)
4. Retain N most recent backups, prune older ones
5. Write a status record (timestamp, size, success/failure) to S3 metadata

This is a standalone shell script (~30 lines), not an agent. It depends on nothing in the agent system.

### Verification

A separate, less-frequent CronJob (daily or weekly):

1. Download most recent backup from S3
2. Restore to a temporary database
3. Run validation queries (row counts, table existence, basic integrity)
4. Drop temporary database
5. Write verification result to S3 metadata

Also standalone. Also ~30 lines.

### Monitoring agent

A lightweight agent (or simple CronJob) that checks:

- Backup age: if most recent successful backup is older than threshold → `needs_backup_repair` work item
- Backup size: sudden large change → alert
- Verification status: last verify failed → alert

This is the only part that touches the agent system, and it's optional.

### Disaster recovery with git spec snapshots

After a DB restore from backup:

1. DB backup restores most data to a point in time (e.g. 3am)
2. `spec-recovery-agent` scans all site git repos for `.site-spec.json`
3. For each site: compare git spec timestamp against DB spec timestamps
4. If git spec is newer: insert spec rows with `source = 'recovery'`
5. Diff against deployed state (HTML/CSS in git) → create work items for anything stale

The gap between backup timestamp and now gets filled by the git snapshots. Pages, page_components, orchestration state are lost for that gap — but the specs tell the system what each site should be, and work items bring it back in line.

---

## Domain Queue

### Table: `build_queue`

```sql
CREATE TABLE build_queue (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    domain          text UNIQUE NOT NULL,
    direction       jsonb,          -- null to full brief or adoption/fork spec
    status          text NOT NULL DEFAULT 'queued',  -- queued, seeded, building, complete, failed
    batch_id        uuid,           -- groups domains for paced processing
    priority        integer DEFAULT 100,
    created_at      timestamptz DEFAULT now(),
    updated_at      timestamptz DEFAULT now()
);
```

The `direction` column handles the full spectrum:

| Domain | Direction | First work item |
|---|---|---|
| `vonc.com` | `null` | `needs_domain_research` |
| `finetuning.uk` | `{"objective": "AI fine-tuning"}` | `needs_domain_research` (with hint) |
| `wykefarm.co.uk` | `{full brief, brief_complete: true}` | `needs_site_plan` (skip research + briefing) |
| `mortgagecalculator.co.uk` | `{"adopt_from": "mortgagecalculator.co.uk"}` | `needs_site_adoption` |
| `webdesign.uk` | `{"fork_from": "design.co.uk", "copy_aspects": [...]}` | `needs_site_plan` (specs pre-populated) |

### Seeding

A `seed_build_queue` action (Go) processes queued domains:

1. Take N domains from `build_queue WHERE status = 'queued' ORDER BY priority, created_at LIMIT N`
2. For each: `ensure_site_record` → examine direction → insert appropriate first work item → mark `seeded`
3. If direction contains initial spec data, write it to `site_specs` with `source = 'seed'`
4. The batch scheduler or manual trigger kicks site-work-orchestrator per site

Pacing is controlled by batch size and frequency. Add 5 domains, they seed, orchestrators spawn. Check logs. Add more when ready.

---

## Initial Build: Work Item Chain

When a domain is seeded with minimal or no direction:

```
needs_domain_research   | priority 1  | handler: domain-research-classifier | source: seed
needs_briefing          | priority 2  | handler: briefing-agent             | source: seed    | depends: [research]
needs_site_plan         | priority 3  | handler: site-planner               | source: seed    | depends: [briefing]
```

Each handler completes and the next item becomes eligible (via `depends_on`). No HITL pause by default.

### Handler outputs → site_specs

Each handler writes to site_specs as part of completing:

- **domain-research-classifier:** identity, strategy, tone, visual_direction, image_guidance (milestone: `initial_research`)
- **briefing-agent:** identity (refined), tone (refined)
- **site-planner:** structure (milestone: `initial_plan`)

### Planner writes downstream items

The planner runs in handler mode: receives `site_id`, `domain`, reads specs from DB. Writes work items directly to `site_work_items`:

```
needs_logo              | priority 5  | handler: image-generator      | source: planner
needs_hero_image        | priority 5  | handler: image-generator      | source: planner
needs_design            | priority 8  | handler: webdesign-agent      | source: planner | depends: [logo, hero]
needs_content_page (/)  | priority 10 | handler: page-content-writer  | source: planner | depends: [design]
needs_content_page (/about)    | priority 10 | handler: page-content-writer  | source: planner | depends: [design]
needs_content_page (/services) | priority 10 | handler: page-content-writer  | source: planner | depends: [design]
needs_content_page (/contact)  | priority 10 | handler: page-content-writer  | source: planner | depends: [design]
needs_nav_setup         | priority 12 | handler: nav-agent            | source: planner
needs_sitemap           | priority 15 | handler: sitemap-agent        | source: planner
```

Also creates page records in the `pages` table. Content writer work items reference these via `page_id`.

Priority numbers control ordering. Lower = first. Dependencies add hard ordering on top.

### Design before content

Webdesign-agent depends on logo and hero (CSS theme references brand colors from generation). Content pages depend on design (assembled with correct chrome).

### End state

```
finetuning.uk:
  /index.html, /about.html, /services.html, /contact.html
  /assets/css/styles.css, /assets/images/logo.png, /assets/images/hero.jpg
  /sitemap.xml
  /.site-spec.json (milestone: post_build)
  
  site_specs: identity, strategy, tone, visual_direction, design, structure
  site_work_items: ~12 items, all status = 'complete'
```

Each item = one git commit = live via GitHub Actions → S3.

---

## Approval Model

### Built in from day one, defaulting to auto

```sql
ALTER TABLE site_work_items ADD COLUMN approval_mode text DEFAULT 'auto';
-- values: 'auto', 'hitl', 'eval'
```

**auto:** Handler runs immediately when eligible. Default for all items.

**hitl:** Status moves to `pending_review` instead of being dispatched. Human approves, status moves to `approved`, orchestrator picks it up.

**eval:** Evaluation agent reviews handler output before marking complete. If eval fails, item goes to `needs_revision`. Future — column exists from day one, implementation in Phase 4.

### Override levels

1. **Per-item:** `approval_mode` on individual work items
2. **Per-item-type:** Site setting maps item types to approval modes
3. **Per-site:** Site-level override (e.g. wykefarm = hitl for everything initially)
4. **System default:** auto

The dispatch loop checks before calling the handler: if `approval_mode = 'hitl'` and `status != 'approved'`, skip to next item.

---

## Expansion

Expansion looks exactly like initial build — different sources write items to the same queue.

### Sources of expansion work

| Source | Trigger | Example items written |
|---|---|---|
| **Manual** | API/CLI/dashboard | `needs_tool_page`, `needs_section: blog` |
| **Re-plan** | Planner runs in additive mode | New page items, new section items |
| **Improvement agent** | Periodic scan | `needs_faq_page`, `needs_social_proof_section` |
| **Content feed** | External data source | `needs_article: "New FCA regulation"` |
| **Side effects** | Handler completes | `needs_nav_update`, `needs_sitemap`, `needs_link_check` |
| **Marketing agent** | Campaign analysis | `needs_landing_page`, `needs_schema_markup` |
| **Discovery** | Maintenance scan | `stale_content`, `broken_link`, `missing_meta` |

### Structural expansion: "add a blog section"

```
needs_section_expansion | spec: {section_type: "blog", reason: "SEO freshness"}
                        | handler: site-planner (additive mode)
                        | approval_mode: hitl
```

Planner reads existing pages from DB, doesn't touch them, writes new items.

### Tool pages: "add a mortgage calculator"

```
needs_tool_page | spec: {page_name: "mortgage-calculator", tool_type: "calculator",
                         description: "Monthly repayment calculator with LTV and stamp duty"}
                | handler: tool-builder
                | approval_mode: hitl
```

Tool-builder: research → generate HTML + JS → assemble with chrome → git commit → side effects.

### Side effects: deterministic follow-on items

Go rules engine (not LLM), runs after each handler completes:

| Trigger | Side-effect items |
|---|---|
| New page created | `needs_nav_update` + `needs_sitemap` |
| Page deleted | `needs_redirect` + `needs_nav_update` + `needs_sitemap` |
| CSS changed | `needs_rerender` for affected pages |
| Content with links changed | `needs_link_check` |

Side-effect items: `source: 'side_effect'`, `approval_mode: 'auto'`.

### Improvement agent: autonomous growth suggestions

Periodic scan comparing the site against competitors, SEO gaps, structural gaps, missing conversion elements. Each finding → work item with `source: 'improvement'`, `approval_mode: 'hitl'`.

---

## Marketing: SEM, Outbound, and Growth

Marketing is part of the site lifecycle. The same work-item model extends to marketing tasks.

### Marketing as work items

| Item Type | Handler | Example |
|---|---|---|
| `needs_sem_campaign` | `sem-campaign-agent` | Google Ads campaign for finetuning.uk |
| `needs_landing_page` | `landing-page-builder` | /free-audit for the ad campaign |
| `needs_email_sequence` | `email-sequence-agent` | Welcome sequence for new leads |
| `needs_social_content` | `social-content-agent` | LinkedIn posts for new blog |
| `needs_schema_markup` | `schema-markup-agent` | LocalBusiness + Service schema |
| `needs_ad_copy` | `ad-copy-agent` | Headlines + descriptions for Google Ads |
| `sem_performance_check` | `sem-analyst-agent` | Weekly metrics, suggest adjustments |

### OpenClaw as marketing adapter

Rather than building direct integrations to Google Ads, Meta, Mailchimp, etc., OpenClaw acts as a marketing execution adapter.

```
Our agents                    OpenClaw adapter                  External platforms
──────────                    ───────────────                   ──────────────────
sem-campaign-agent            
  → structured campaign spec  openclaw-adapter
                                → translates to OpenClaw task
                                → OpenClaw executes:
                                  → Google Ads API              Google Ads
                                  → Meta Ads API                Meta/Facebook
                                  → LinkedIn Ads API            LinkedIn
                                → returns results
  ← campaign IDs, metrics    
```

The `openclaw-adapter` is an adapter service (like the GitHub adapter), not an agent. It receives structured requests from marketing agents via Kafka/HTTP, translates to OpenClaw-native instructions, and returns structured results.

Advantages: OpenClaw already handles auth, rate limits, platform quirks, browser automation. Adding a new platform = adding an OpenClaw skill, not building a new adapter. Self-hosted — data stays on our infrastructure.

### SEM campaign lifecycle

```
Phase 1: Research → competitor ads, keyword opportunities, budget recommendation
Phase 2: Setup (hitl) → ad copy, landing page, campaign structure
Phase 3: Launch → create campaign via OpenClaw adapter, conversion tracking
Phase 4: Monitor → daily performance checks, work items for optimization
```

### Marketing discovery

A `marketing-discovery-agent` runs periodically: Google Business Profile? Schema markup? Pages close to page 1? Competitors running ads? Each finding → work item → handler → done.

---

## Agent Inventory

### Existing (unchanged)

| Agent | Role in new pipeline |
|---|---|
| `research-agent` | Spawned by classifier, planner, content writers, improvement agent |
| `page-content-writer` | Handler for `needs_content_page` |
| `image-generator` | Handler for `needs_logo`, `needs_hero_image` |
| `webdesign-agent` | Handler for `needs_design`, `missing_css` |
| `asset-deployer` | Handler for `undeployed_asset` |
| `deployer-agent` | Git commit + deploy (used by handlers internally) |
| `content-reviewer` | Eval agent for content quality (future `approval_mode: eval`) |
| `page-rerender` | Handler for `needs_rerender` |

### New agents

| Agent | Handler for | Phase | Complexity |
|---|---|---|---|
| `domain-research-classifier` | `needs_domain_research` | 0 | Medium |
| `site-planner` (handler mode) | `needs_site_plan` | 0 | Medium |
| `site-adoption-agent` | `needs_site_adoption` | 1 | High |
| `nav-agent` | `needs_nav_update`, `needs_nav_setup` | 1 | Low |
| `sitemap-agent` | `needs_sitemap` | 1 | Low |
| `spec-recovery-agent` | Disaster recovery from git snapshots | 1 | Low |
| `tool-builder` | `needs_tool_page` | 2 | High |
| `improvement-agent` | Writes items, doesn't handle them | 2 | Medium |
| `schema-markup-agent` | `needs_schema_markup` | 2 | Low |
| `sem-campaign-agent` | `needs_sem_campaign` | 3 | Medium |
| `ad-copy-agent` | `needs_ad_copy` | 3 | Low |
| `landing-page-builder` | `needs_landing_page` | 3 | Medium |
| `sem-analyst-agent` | `sem_performance_check` | 3 | Medium |
| `email-sequence-agent` | `needs_email_sequence` | 3 | Medium |
| `social-content-agent` | `needs_social_content` | 3 | Low |
| `marketing-discovery-agent` | Writes items, doesn't handle them | 3 | Medium |

### New adapters

| Adapter | Connects to | Phase |
|---|---|---|
| `openclaw-adapter` | OpenClaw → Google Ads, Meta, LinkedIn, email, social | 3 |

### New Go actions

| Action | Used by | Phase |
|---|---|---|
| `seed_build_queue` | Reads build_queue, creates site + first work items | 0 |
| `write_plan_as_work_items` | site-planner handler mode | 0 |
| `write_site_spec` | Any handler writing specs | 0 |
| `read_site_spec` | Any handler reading specs | 0 |
| `check_approval_mode` | Dispatch loop, before handler call | 0 |
| `commit_spec_snapshot` | Called at milestone moments | 0 |
| `check_side_effects` | Dispatch loop, after mark_complete | 1 |
| `rollback_specs` | CLI/API triggered | 1 |
| `fork_specs` | Seed action for fork-type direction | 1 |

---

## Phases

### Phase 0 — Work item build pipeline

The minimum to build a site via work items.

1. `site_specs` table migration
2. `build_queue` table migration
3. `approval_mode` column on `site_work_items`
4. `page_spec` column on `pages`
5. `write_site_spec` and `read_site_spec` Go actions
6. `commit_spec_snapshot` Go action
7. `seed_build_queue` Go action (handles null, direction, brief_complete, adopt_from, fork_from)
8. `domain-research-classifier` agent definition (new agent, not modifying existing)
9. `briefing-agent` adapted as handler (receives work item context)
10. `site-planner` handler mode + `write_plan_as_work_items` action
11. `check_approval_mode` in dispatch loop
12. Dispatch loop handles `depends_on` ordering
13. Test: seed finetuning.uk → full build via work items

### Phase 1 — Infrastructure, adoption, and resilience

14. `check_side_effects` rules engine in dispatch loop
15. `nav-agent` (extracted from existing action)
16. `sitemap-agent`
17. `site-adoption-agent` (scrape → parse → map components → write specs + items)
18. `area_spec` column on `site_areas`
19. `rollback_specs` and `fork_specs` Go actions
20. `spec-recovery-agent`
21. DB backup CronJob (pg_dump → S3, standalone script)
22. DB backup verification CronJob
23. Backup monitoring (health checks, stale backup alerts)
24. `improvement-agent` v1 (structural gaps only)
25. Content, links, SEO discovery agents

### Phase 2 — Tools and advanced content

26. `tool-builder` agent
27. `schema-markup-agent`
28. Section-rewriter, legal-updater
29. `improvement-agent` v2 (SEO gaps, conversion optimization)
30. Content feed pipeline

### Phase 3 — Marketing

31. `openclaw-adapter` service
32. `sem-campaign-agent` + `ad-copy-agent`
33. `landing-page-builder`
34. `sem-analyst-agent` (monitoring + optimization loop)
35. `email-sequence-agent` + email adapter
36. `social-content-agent`
37. `marketing-discovery-agent`

### Phase 4 — Autonomous growth

38. Auto-approval based on eval agent scores
39. Improvement agent on schedule, auto-approve if high-confidence
40. Marketing feedback loop: campaign metrics influence content priorities
41. Cross-site learning: patterns from one site suggested to similar sites

---

## Data Flow: Complete Lifecycle

```
Build queue                         Work items                          Handlers
──────────                          ──────────                          ────────

domain added     ──seed──►  needs_domain_research  ──dispatch──►  domain-research-classifier
                              (writes site_specs)                    │
                            needs_briefing          ──dispatch──►  briefing-agent
                              (refines site_specs)                   │
                            needs_site_plan         ──dispatch──►  site-planner
                              (writes structure spec + milestone)    │
                                                    ◄──writes items──┘
                            needs_logo              ──dispatch──►  image-generator
                            needs_hero_image        ──dispatch──►  image-generator
                            needs_design            ──dispatch──►  webdesign-agent
                              (writes design spec)
                            needs_content_page × N  ──dispatch──►  page-content-writer
                              (reads specs for tone, identity)      │
                                                    ◄──side effects──┘
                            needs_nav_setup         ──dispatch──►  nav-agent
                            needs_sitemap           ──dispatch──►  sitemap-agent
                              (milestone: post_build → git spec snapshot)

                            ═══ site is live ═══

adopt existing   ──seed──►  needs_site_adoption    ──dispatch──►  site-adoption-agent
                              (writes all specs from scrape)        │
                                                    ◄──writes items──┘
                            needs_design            ──dispatch──►  webdesign-agent
                            needs_content_page × N  ──dispatch──►  page-content-writer (recreate)
                            needs_tool_page         ──dispatch──►  tool-builder (recreate)

fork from site   ──seed──►  (specs pre-populated from source)
                            needs_site_plan         ──dispatch──►  site-planner

improvement      ──scan──►  improvement_suggestion  ──hitl──►     site-planner (additive)
                            needs_content_page      ──dispatch──►  page-content-writer
                            needs_tool_page         ──dispatch──►  tool-builder

discovery        ──scan──►  stale_content           ──dispatch──►  section-rewriter
                            broken_link             ──dispatch──►  redirect-manager
                            missing_meta            ──dispatch──►  seo-content-agent

marketing        ──scan──►  needs_sem_campaign      ──hitl──►     sem-campaign-agent
                            needs_ad_copy           ──dispatch──►  ad-copy-agent
                            needs_landing_page      ──dispatch──►  landing-page-builder
                                                    ◄──openclaw──►  Google Ads / Meta / etc.

rollback         ──query──► (old specs restored, source: rollback)
                            needs_design            ──dispatch──►  webdesign-agent
                            needs_content_rewrite   ──dispatch──►  page-content-writer

recovery         ──scan──►  (specs from git snapshots, source: recovery)
                            (work items for anything stale)
```

Everything flows through the same queue.

---

## What the pageflow-builder path remains for

The existing `intake-orchestrator → site-classifier → briefing-agent → pageflow-builder` pipeline continues to work unchanged. Useful for:

- Domains with rich direction where a monolithic build is simpler
- Testing and comparison: build same domain both ways
- Existing sites in that pipeline — no migration needed

The work-item pipeline becomes the default path over time.
