this doc is a bit out of date but still has merit
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
    
    -- Currency
    is_current      boolean NOT NULL DEFAULT true,
    created_at      timestamptz NOT NULL DEFAULT now(),
    superseded_at   timestamptz,
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

-- Recent history queries
CREATE INDEX idx_site_specs_history
ON site_specs (site_id, aspect, created_at DESC);
```

### Every row is a complete record

The `write_site_spec` Go action enforces this. Callers can pass partial updates — the action deep-merges them over the current value and stores the full result:

```go
func WriteSiteSpec(siteID uuid.UUID, aspect string, update jsonb, source string, ...) {
    current := query("SELECT data FROM site_specs WHERE site_id=$1 AND aspect=$2 AND is_current=true")
    merged := deepMerge(current, update)
    // Mark old row as is_current = false, superseded_at = now()
    // Insert new row with merged (complete) value, is_current = true
}
```

This means any row can be pruned without losing data. No rebase problem. No incremental update reconstruction.

### Pruning

```sql
-- Keep last 5 versions per site per aspect, delete older
DELETE FROM site_specs
WHERE id NOT IN (
    SELECT id FROM (
        SELECT id, ROW_NUMBER() OVER (
            PARTITION BY site_id, aspect ORDER BY created_at DESC
        ) AS rn
        FROM site_specs
    ) ranked WHERE rn <= 5
);
```

Safe to run at any time. Every remaining row is self-contained.

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

---

## Page Content History

### Table: `page_component_history`

Tracks previous content_data values for page components. Used for fast structured rollback of recent changes.

```sql
CREATE TABLE page_component_history (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    component_id    uuid REFERENCES page_components(id) ON DELETE SET NULL,
    page_id         uuid NOT NULL REFERENCES pages(id),
    site_id         uuid NOT NULL REFERENCES sites(id),
    content_data    jsonb NOT NULL,
    
    source          text NOT NULL,
    source_item_id  uuid,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_pch_component ON page_component_history (component_id, created_at DESC);
CREATE INDEX idx_pch_site ON page_component_history (site_id, created_at DESC);
CREATE INDEX idx_pch_source ON page_component_history (source_item_id)
WHERE source_item_id IS NOT NULL;
```

Notes:
- `component_id` uses `ON DELETE SET NULL` so history survives page deletion.
- Each row is a complete copy of the previous content_data, not a diff.
- Before any content_data write to page_components, the current value is copied to this table.

### Pruning

```sql
-- Keep last 5 versions per component, delete older
DELETE FROM page_component_history
WHERE id NOT IN (
    SELECT id FROM (
        SELECT id, ROW_NUMBER() OVER (
            PARTITION BY component_id ORDER BY created_at DESC
        ) AS rn
        FROM page_component_history
    ) ranked WHERE rn <= 5
);
```

### Removing legacy columns

The `content_snapshot` and `schema_snapshot` columns on `page_components` are not in use and are replaced by this table. Migration to drop them:

```sql
ALTER TABLE page_components DROP COLUMN IF EXISTS content_snapshot;
ALTER TABLE page_components DROP COLUMN IF EXISTS schema_snapshot;
```

---

## Data History Strategy

Git is the long-term history store. The DB tables are a hot cache of recent structured data for fast rollback. Anything older than the prune window requires git checkout or adoption-style recovery.

| Storage | What it holds | How far back | Prune policy |
|---|---|---|---|
| `site_specs` (DB) | Current + recent history per aspect | Last 5 per aspect | Hard delete older |
| `page_component_history` (DB) | Recent content changes per component | Last 5 per component | Hard delete older |
| `.site-spec.json` (git) | Complete spec at milestone moments | Forever | Git GC |
| Deployed files (git) | Rendered HTML/CSS/images at every commit | Forever | Git GC |
| S3 archive (future) | Deep structured history before pruning | If needed | Configurable |

No `site_spec_snapshots` table. Git handles milestone snapshots. No need for a DB copy of data that's already in git.

If 5 versions per component proves insufficient, an S3 archival step can dump older rows as compressed JSON before pruning. Not built initially — only if the need arises.

---

## Snapshots

### What a snapshot is

A `.site-spec.json` file committed to the site's git repo. Contains all current aspects from site_specs assembled into one JSON document. Committed with a message including the milestone name.

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

### Snapshot-agent

A spawned agent, triggered by `needs_snapshot` work items. Independent logs, identifiable in monitoring, extensible workflow.

```
needs_snapshot | handler: snapshot-agent | spec: {milestone: "plan_complete"}
```

Workflow:
```
start: read_specs
  → action: read_all_current_specs  (SELECT from site_specs for site_id)
  → next: commit_to_git
commit_to_git:
  → action: commit_spec_file        (.site-spec.json, message "spec: {milestone}")
  → next: complete
```

Two steps. Read DB, write git. Extensible — add steps to upload to S3, send notifications, etc.

### What triggers snapshots

All triggers create work items. No inline snapshot calls in the dispatch loop.

**Automated (side-effects engine):** After certain item types complete, create a `needs_snapshot` item:

```go
var postSnapshotTriggers = map[string]string{
    "needs_site_plan":      "plan_complete",
    "needs_design":         "design_complete",
    "needs_site_adoption":  "adoption_complete",
    "needs_brand_update":   "post_rebrand",
}
```

**Build-complete:** When the dispatch loop detects no more pending items for a site, it creates a `needs_snapshot` with `item_key: "snapshot:build_complete:{batch_id}"`. The item_key dedup index prevents re-creation on subsequent checks.

**Human-initiated:** Add a work item to the queue:

```
needs_snapshot | spec: {milestone: "pre_restructure", notes: "Before blog changes"}
```

For pre-change protection, the human creates paired items:

```
needs_snapshot          | priority 1 | spec: {milestone: "pre_restructure"}
needs_section_expansion | priority 5 | depends_on: [snapshot_id]
```

**Agent-initiated:** Agents creating significant changes (improvement agent proposing rebrand, etc.) pair their change items with a snapshot item, same as humans.

The dispatch loop has no "before" triggers. Pre-change snapshots are the responsibility of whoever creates the change — human, agent, or improvement agent. "After" snapshots are automated by the side-effects engine.

---

## Rollback

### Spec rollback

Reads `.site-spec.json` from git history, diffs against current site_specs, writes changed aspects back, creates work items:

```go
func RollbackToMilestone(siteID uuid.UUID, milestone string) {
    // 1. Find git commit with message "spec: {milestone}"
    commitHash := gitLog(repo, "--grep=spec: " + milestone, ".site-spec.json")
    
    // 2. Read .site-spec.json from that commit
    specJSON := gitShow(repo, commitHash + ":.site-spec.json")
    
    // 3. Diff against current site_specs
    current := getCurrentSpecs(siteID)
    changed := diff(specJSON, current)
    
    // 4. Write changed aspects as new current rows (source = 'rollback')
    // 5. Create work items for rebuilds:
    //    - identity/tone changed → needs_content_rewrite per page
    //    - visual_direction changed → needs_design
    //    - image_guidance changed → needs_logo, needs_hero_image
    //    - structure changed → needs_site_plan (re-plan)
}
```

### Content rollback

"This page content was wrong" is separate from spec rollback. The spec didn't change — only the output was wrong.

```sql
-- Find previous content for a specific component
SELECT content_data FROM page_component_history
WHERE component_id = $1 ORDER BY created_at DESC LIMIT 1;

-- Or: undo everything from a specific work item
SELECT component_id, content_data FROM page_component_history
WHERE source_item_id = $1;
```

Then create `needs_page_rerender` work items with the old content_data. Normal queue processing.

---

## Fork / Copy

Create a new site starting from another site's spec state.

### Direction in build_queue

```json
{
  "fork_from": "site-uuid-or-domain",
  "fork_time": "2026-02-20T14:00:00Z",
  "copy_aspects": ["tone", "visual_direction", "design"],
  "override": {
    "identity": {"name": "NewSite", "audience_primary": "different audience"}
  }
}
```

### Seed action

1. Create site record for new domain
2. Read source site's specs from git snapshot at `fork_time`, or from current site_specs if no time specified
3. Copy specified aspects (or all if `copy_aspects` absent)
4. Apply overrides
5. Write as new site_specs rows with `source = 'fork'`
6. Seed work items — typically `needs_site_plan` (specs already populated)

### Network siblings

```json
{"fork_from": "design.co.uk", "copy_aspects": ["strategy", "tone", "visual_direction"],
 "override": {"identity": {"name": "WebDesign.uk", "audience_primary": "UK startups"}}}
```

Network-wide brand changes: INSERT work items per site, each processes independently:

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

```json
{"adopt_from": "mortgagecalculator.co.uk"}
```

### What the adoption agent does

1. **Crawl:** Spawns site-scraper in full-site mode — all pages, CSS, assets, metadata
2. **Parse design:** Extract color palette, typography, spacing from CSS
3. **Map components:** LLM analysis maps scraped sections to our component library
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

A content writer receiving `mode: "recreate"` with `existing_content` uses that content instead of generating from scratch. A webdesign-agent receiving `adopt_from.color_palette` matches those colors. The spec is just richer than usual.

### Adopt then improve

Phase 1: Adopt — recreate the site as-is. All items have `mode: "recreate"`.

Phase 2: Improvement agent compares against what a fresh classifier would recommend. Writes items for weak copy, missing sections, dated UX, thin SEO.

---

## Database Backup

Independent of the agent system. The backup process must work even when the cluster is in trouble.

### Backup script

A K8s CronJob (every 6 hours, configurable):

1. `pg_dump` the clients_db
2. Compress (gzip)
3. Upload to Backblaze S3
4. Retain N most recent, prune older
5. Write status record to S3 metadata

Standalone shell script (~30 lines), not an agent.

### Verification

Separate less-frequent CronJob (daily/weekly): download backup → restore to temp DB → validation queries → drop temp. Also standalone.

### Disaster recovery

After DB restore from backup:

1. DB backup restores most data to a point in time (e.g. 3am)
2. `spec-recovery-agent` scans all site git repos for `.site-spec.json`
3. For each site: compare git spec timestamp against DB spec timestamps
4. If git spec is newer: write spec rows with `source = 'recovery'`
5. Diff against deployed state → create work items for anything stale

Or: treat a git checkout as an adoption source — the recovery agent runs the adoption flow on the deployed files.

---

## Domain Queue

### Table: `build_queue`

```sql
CREATE TABLE build_queue (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    domain          text UNIQUE NOT NULL,
    direction       jsonb,
    status          text NOT NULL DEFAULT 'queued',
    batch_id        uuid,
    priority        integer DEFAULT 100,
    created_at      timestamptz DEFAULT now(),
    updated_at      timestamptz DEFAULT now()
);
```

Direction handles the full spectrum:

| Domain | Direction | First work item |
|---|---|---|
| `vonc.com` | `null` | `needs_domain_research` |
| `finetuning.uk` | `{"objective": "AI fine-tuning"}` | `needs_domain_research` (with hint) |
| `wykefarm.co.uk` | `{full brief, brief_complete: true}` | `needs_site_plan` (skip research + briefing) |
| `mortgagecalculator.co.uk` | `{"adopt_from": "mortgagecalculator.co.uk"}` | `needs_site_adoption` |
| `webdesign.uk` | `{"fork_from": "design.co.uk", ...}` | `needs_site_plan` (specs pre-populated) |

### Seeding

`seed_build_queue` action (Go): take N from queue → `ensure_site_record` → examine direction → write initial specs if present → insert appropriate first work item → mark `seeded`.

Pacing controlled by batch size. Add 5, they seed, orchestrators spawn. Check logs. Add more.

---

## Initial Build: Work Item Chain

When a domain is seeded with minimal or no direction:

```
needs_domain_research   | priority 1  | handler: domain-research-classifier | source: seed
needs_briefing          | priority 2  | handler: briefing-agent             | source: seed    | depends: [research]
needs_site_plan         | priority 3  | handler: site-planner               | source: seed    | depends: [briefing]
```

### Handler outputs → site_specs

- **domain-research-classifier:** identity, strategy, tone, visual_direction, image_guidance
- **briefing-agent:** identity (refined), tone (refined)
- **site-planner:** structure

### Planner writes downstream items

```
needs_logo              | priority 5  | handler: image-generator      | source: planner
needs_hero_image        | priority 5  | handler: image-generator      | source: planner
needs_design            | priority 8  | handler: webdesign-agent      | source: planner | depends: [logo, hero]
needs_content_page (/)       | priority 10 | handler: page-content-writer  | source: planner | depends: [design]
needs_content_page (/about)  | priority 10 | handler: page-content-writer  | source: planner | depends: [design]
needs_content_page (/etc)    | priority 10 | handler: page-content-writer  | source: planner | depends: [design]
needs_nav_setup         | priority 12 | handler: nav-agent            | source: planner
needs_sitemap           | priority 15 | handler: sitemap-agent        | source: planner
```

Also creates page records in `pages`. Priority controls ordering. Dependencies add hard constraints.

### End state

```
finetuning.uk:
  /index.html, /about.html, /services.html, /contact.html
  /assets/css/styles.css, /assets/images/logo.png, /assets/images/hero.jpg
  /sitemap.xml
  /.site-spec.json (milestone: build_complete)
  
  site_specs: identity, strategy, tone, visual_direction, design, structure
  site_work_items: ~12 items, all status = 'complete'
```

---

## Approval Model

```sql
ALTER TABLE site_work_items ADD COLUMN approval_mode text DEFAULT 'auto';
```

**auto:** Handler runs immediately when eligible. Default for all items.

**hitl:** Status moves to `pending_review`. Human approves, status moves to `approved`, orchestrator picks it up.

**eval:** Evaluation agent reviews handler output before marking complete. Future — column exists from day one.

### Override levels

1. **Per-item:** `approval_mode` on individual work items
2. **Per-item-type:** Site setting maps item types to approval modes
3. **Per-site:** Site-level override
4. **System default:** auto

---

## Expansion

Same queue, same dispatch. Different sources write items.

| Source | Trigger | Example items |
|---|---|---|
| **Manual** | API/CLI/dashboard | `needs_tool_page`, `needs_section: blog` |
| **Re-plan** | Planner in additive mode | New page items |
| **Improvement agent** | Periodic scan | `needs_faq_page`, `needs_social_proof` |
| **Content feed** | External data | `needs_article` |
| **Side effects** | Handler completes | `needs_nav_update`, `needs_sitemap` |
| **Marketing agent** | Campaign analysis | `needs_landing_page`, `needs_schema_markup` |
| **Discovery** | Maintenance scan | `stale_content`, `broken_link` |

### Side effects: deterministic follow-on items

Go rules engine (not LLM), runs after each handler completes:

| Trigger | Side-effect items |
|---|---|
| New page created | `needs_nav_update` + `needs_sitemap` |
| Page deleted | `needs_redirect` + `needs_nav_update` + `needs_sitemap` |
| CSS changed | `needs_rerender` for affected pages |
| Content with links changed | `needs_link_check` |
| Certain item types (plan, design, etc.) | `needs_snapshot` (automated milestone) |

---

## Marketing: SEM, Outbound, and Growth

### Marketing as work items

| Item Type | Handler |
|---|---|
| `needs_sem_campaign` | `sem-campaign-agent` |
| `needs_landing_page` | `landing-page-builder` |
| `needs_email_sequence` | `email-sequence-agent` |
| `needs_social_content` | `social-content-agent` |
| `needs_schema_markup` | `schema-markup-agent` |
| `needs_ad_copy` | `ad-copy-agent` |
| `sem_performance_check` | `sem-analyst-agent` |

### OpenClaw as marketing adapter

```
Our agents                    OpenClaw adapter                  External platforms
──────────                    ───────────────                   ──────────────────
sem-campaign-agent
  → structured campaign spec  openclaw-adapter
                                → translates to OpenClaw task
                                → OpenClaw executes:
                                  → Google Ads, Meta, LinkedIn
                                → returns results
  ← campaign IDs, metrics
```

The `openclaw-adapter` is an adapter service (like the GitHub adapter), not an agent. Self-hosted — data stays on our infrastructure.

### SEM campaign lifecycle

```
Research → Setup (hitl) → Launch (via OpenClaw) → Monitor (automated, work items for optimization)
```

### Marketing discovery

`marketing-discovery-agent` runs periodically: GBP? Schema markup? Page 2 rankings? Competitor ads? Each finding → work item.

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
| `content-reviewer` | Eval agent (future `approval_mode: eval`) |
| `page-rerender` | Handler for `needs_rerender` |

### New agents

| Agent | Handler for | Phase | Complexity |
|---|---|---|---|
| `domain-research-classifier` | `needs_domain_research` | 0 | Medium |
| `site-planner` (handler mode) | `needs_site_plan` | 0 | Medium |
| `snapshot-agent` | `needs_snapshot` | 0 | Low |
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
| `write_site_spec` | Any handler writing specs (deep-merge, complete rows) | 0 |
| `read_site_spec` | Any handler reading specs | 0 |
| `seed_build_queue` | Reads build_queue, creates site + first work items | 0 |
| `write_plan_as_work_items` | site-planner handler mode | 0 |
| `check_approval_mode` | Dispatch loop, before handler call | 0 |
| `save_component_history` | Before any content_data write | 0 |
| `commit_spec_snapshot` | snapshot-agent workflow | 0 |
| `check_side_effects` | Dispatch loop, after mark_complete | 1 |
| `rollback_specs` | CLI/API triggered | 1 |
| `fork_specs` | Seed action for fork-type direction | 1 |

---

## Phases

### Phase 0 — Work item build pipeline

The minimum to build a site via work items.

1. `site_specs` table migration
2. `build_queue` table migration
3. `page_component_history` table migration
4. `approval_mode` column on `site_work_items`
5. `page_spec` column on `pages`
6. Drop `content_snapshot` / `schema_snapshot` from `page_components`
7. `write_site_spec` and `read_site_spec` Go actions
8. `save_component_history` Go action (called before content_data writes)
9. `seed_build_queue` Go action (handles null, direction, brief_complete, adopt_from, fork_from)
10. `domain-research-classifier` agent definition (new agent, not modifying existing)
11. `briefing-agent` adapted as handler (receives work item context)
12. `site-planner` handler mode + `write_plan_as_work_items` action
13. `snapshot-agent` definition + `commit_spec_snapshot` action
14. `check_approval_mode` in dispatch loop
15. Dispatch loop handles `depends_on` ordering
16. Side-effects engine: snapshot triggers after plan/design/build-complete
17. Test: seed finetuning.uk → full build via work items

### Phase 1 — Infrastructure, adoption, and resilience

18. `check_side_effects` rules engine (nav, sitemap, link, snapshot triggers)
19. `nav-agent`
20. `sitemap-agent`
21. `site-adoption-agent` (scrape → parse → map components → write specs + items)
22. `area_spec` column on `site_areas`
23. `rollback_specs` and `fork_specs` Go actions
24. `spec-recovery-agent`
25. DB backup CronJob (pg_dump → S3, standalone script)
26. DB backup verification CronJob
27. Backup monitoring (stale backup alerts)
28. `improvement-agent` v1 (structural gaps only)
29. Content, links, SEO discovery agents
30. Pruning CronJobs for site_specs and page_component_history

### Phase 2 — Tools and advanced content

31. `tool-builder` agent
32. `schema-markup-agent`
33. Section-rewriter, legal-updater
34. `improvement-agent` v2 (SEO gaps, conversion optimization)
35. Content feed pipeline
36. S3 archival for deep history (if prune window proves insufficient)

### Phase 3 — Marketing

37. `openclaw-adapter` service
38. `sem-campaign-agent` + `ad-copy-agent`
39. `landing-page-builder`
40. `sem-analyst-agent` (monitoring + optimization loop)
41. `email-sequence-agent` + email adapter
42. `social-content-agent`
43. `marketing-discovery-agent`

### Phase 4 — Autonomous growth

44. Auto-approval based on eval agent scores
45. Improvement agent on schedule, auto-approve if high-confidence
46. Marketing feedback loop: campaign metrics influence content priorities
47. Cross-site learning: patterns from one site suggested to similar sites

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
                              (writes structure spec)                │
                                                    ◄──writes items──┘
                            needs_snapshot          ──dispatch──►  snapshot-agent (plan_complete)
                            needs_logo              ──dispatch──►  image-generator
                            needs_hero_image        ──dispatch──►  image-generator
                            needs_design            ──dispatch──►  webdesign-agent
                              (writes design spec)
                            needs_snapshot          ──dispatch──►  snapshot-agent (design_complete)
                            needs_content_page × N  ──dispatch──►  page-content-writer
                              (reads specs, saves history)          │
                                                    ◄──side effects──┘
                            needs_nav_setup         ──dispatch──►  nav-agent
                            needs_sitemap           ──dispatch──►  sitemap-agent
                            needs_snapshot          ──dispatch──►  snapshot-agent (build_complete)

                            ═══ site is live ═══

adopt existing   ──seed──►  needs_site_adoption    ──dispatch──►  site-adoption-agent
                              (writes all specs from scrape)        │
                                                    ◄──writes items──┘
                            needs_design            ──dispatch──►  webdesign-agent
                            needs_content_page × N  ──dispatch──►  page-content-writer (recreate)
                            needs_tool_page         ──dispatch──►  tool-builder (recreate)
                            needs_snapshot          ──dispatch──►  snapshot-agent (adoption_complete)

fork from site   ──seed──►  (specs pre-populated from source)
                            needs_site_plan         ──dispatch──►  site-planner

improvement      ──scan──►  needs_snapshot          ──dispatch──►  snapshot-agent (pre_changes)
                            improvement_suggestion  ──hitl──►     site-planner (additive)
                            needs_content_page      ──dispatch──►  page-content-writer
                            needs_tool_page         ──dispatch──►  tool-builder

discovery        ──scan──►  stale_content           ──dispatch──►  section-rewriter
                            broken_link             ──dispatch──►  redirect-manager
                            missing_meta            ──dispatch──►  seo-content-agent

marketing        ──scan──►  needs_sem_campaign      ──hitl──►     sem-campaign-agent
                            needs_ad_copy           ──dispatch──►  ad-copy-agent
                            needs_landing_page      ──dispatch──►  landing-page-builder
                                                    ◄──openclaw──►  Google Ads / Meta / etc.

rollback         ──query──► (specs from git .site-spec.json → write to site_specs)
                            needs_design            ──dispatch──►  webdesign-agent
                            needs_content_rewrite   ──dispatch──►  page-content-writer

recovery         ──scan──►  (specs from git → site_specs, source: recovery)
                            (work items for anything stale)
```

---

## What the pageflow-builder path remains for

The existing `intake-orchestrator → site-classifier → briefing-agent → pageflow-builder` pipeline continues to work unchanged. Useful for:

- Domains with rich direction where a monolithic build is simpler
- Testing and comparison: build same domain both ways
- Existing sites in that pipeline — no migration needed

The work-item pipeline becomes the default path over time.

