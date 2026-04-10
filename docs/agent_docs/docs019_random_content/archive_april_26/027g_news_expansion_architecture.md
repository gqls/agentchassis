# 026g — News Expansion Architecture

Future expansion path for the news feed pipeline. Covers: JSON-based homepage news, dedicated insights section with nav, research-driven analysis articles, event timeline database, and visualization.

This document captures architectural decisions made now so the current build doesn't block the expansion. Nothing here needs building immediately — it's the roadmap that the current implementation is designed to accommodate.

---

## Three Tiers of News

The news system has three tiers, each building on the previous:

| Tier | What | Where | Update frequency | Status |
|------|------|-------|-----------------|--------|
| 1. Homepage snippets | Title + summary + link to source | Homepage section | Every 6 hours | Building now |
| 2. Insights section | Curated, rewritten articles with own pages | `/insights/` with nav link | Daily | Future |
| 3. Research analysis | Multi-source analysis with timelines, graphs, infographics | `/insights/` (rich pages) | Daily/weekly | Future |

Each tier is independently valuable. Tier 1 works without Tier 2. Tier 2 works without Tier 3. The architecture supports all three without rework.

---

## Tier 1: Homepage News via JSON

### Why JSON instead of page rerender

The homepage has 6-7 sections. News is one of them. Rerendering the entire page every 6 hours to update 6 news cards is disproportionate — the rerender pipeline involves assembling all sections, committing the full HTML to git, triggering GitHub Actions, and pushing to S3.

A JSON file decouples news updates from page structure:

```
/data/latest-news.json   ← small file (~2KB), updated every cycle
/index.html              ← stable, changes only on actual content/design edits
```

Benefits:
- Updates are tiny git commits (one JSON file vs full HTML page)
- Can increase update frequency without cost (hourly if needed)
- Homepage HTML is stable — design changes and news changes are independent
- The insights listing page can use the same pattern (`/data/news-listing.json`)
- Works with CDN caching — short TTL on `/data/`, long TTL on HTML

### JSON file shape

```json
{
    "headline": "What's Moving in Wholesale Gas",
    "subheadline": "Market news and price updates for energy professionals",
    "items": [
        {
            "title": "UK Gas Prices Have Nearly Doubled This Week",
            "summary": "UK wholesale gas prices have nearly doubled following escalation...",
            "url": "https://finance.yahoo.com/news/uk-gas-prices-nearly-doubled",
            "source": "Yahoo Finance",
            "date": "2h ago"
        }
    ],
    "insights_url": "/insights/",
    "insights_label": "More insights →",
    "updated_at": "2026-03-28T18:00:00Z"
}
```

When no insights section exists, `insights_url` is empty and the link doesn't render. When you build Tier 2, the link appears automatically.

### Homepage component template

The `latest-news` component becomes client-side rendered with a noscript fallback:

```html
<section data-component="latest-news" class="latest-news-section section-padding">
  <div class="container">
    <h2 class="section-heading" id="news-headline">Latest News</h2>
    <p class="section-subheadline" id="news-subheadline"></p>
    <div class="news-grid" id="news-container">
      <noscript>
        <p class="news-empty">Visit <a href="/insights/">our insights page</a> 
        for the latest news.</p>
      </noscript>
    </div>
    <div id="news-footer"></div>
  </div>
</section>
<script>
fetch('/data/latest-news.json')
  .then(r => r.json())
  .then(data => {
    if (data.headline) 
      document.getElementById('news-headline').textContent = data.headline;
    if (data.subheadline) 
      document.getElementById('news-subheadline').textContent = data.subheadline;
    const container = document.getElementById('news-container');
    if (data.items && data.items.length > 0) {
      container.innerHTML = data.items.map(item => `
        <article class="news-card">
          <div class="news-card-content">
            <h3 class="news-card-title">
              <a href="${item.url}" target="_blank" rel="noopener">${item.title}</a>
            </h3>
            ${item.summary ? `<p class="news-card-summary">${item.summary}</p>` : ''}
            <div class="news-card-meta">
              ${item.source ? `<span class="news-source">${item.source}</span>` : ''}
              ${item.date ? `<time class="news-date">${item.date}</time>` : ''}
            </div>
          </div>
        </article>
      `).join('');
    }
    if (data.insights_url) {
      document.getElementById('news-footer').innerHTML = 
        `<div class="news-section-footer">
          <a href="${data.insights_url}" class="news-more-link">
            ${data.insights_label || 'More insights →'}
          </a>
        </div>`;
    }
  })
  .catch(() => {});
</script>
```

The `noscript` fallback ensures search engines and users without JS still get a sensible experience. The fetch URL (`/data/latest-news.json`) is relative, so it works on any domain.

### render_news_section changes

The action changes output from "render HTML, upsert page_component" to "build JSON, commit to git":

1. Query `content_feed_items` (same query as now — relevant items, max 6, within 72h)
2. Read headline/subheadline from `page_components.content_data` (set by content writer at build time)
3. Check if insights page exists (for the link)
4. Format as JSON with relative dates
5. Commit `/data/latest-news.json` to git via the existing `git_commit` action pattern
6. S3 deploy happens via GitHub Actions (already configured)

The page_component row still exists (for the webdesign agent to see `latest-news` in the component list and generate CSS). Its `rendered_html` contains the JS template above, not server-rendered items. This is set once at build time and doesn't change.

---

## Tier 2: Insights Section

### Structure

```
/insights/index.html              ← listing page (all articles)
/insights/uk-gas-prices-surge.html ← individual rewritten article
/insights/iran-conflict-impact.html
```

### Page types

| Page type | URL pattern | Template |
|-----------|------------|----------|
| `news-index` | `/insights/index.html` | Card grid listing, similar to blog-index |
| `news-post` | `/insights/{slug}.html` | Article page with content, sources cited |

These follow the same pattern as `blog-index` and `blog-post`. The `link_constraints.go` URL builder needs one addition:

```go
case "news-post": → /insights/{name}.html
```

### Navigation

A "News" or "Insights" entry in the header nav, pointing to `/insights/`. Added by the planner if `news_feed.recommended` is true, or by a discovery check if the section exists but the nav entry doesn't.

### Listing page

`rebuild_news_listing` action (mirrors `rebuild_blog_listing`):
- Queries `pages WHERE page_type = 'news-post' AND status = 'active' ORDER BY created_at DESC`
- Loads news-listing component template
- Renders card grid with title, summary, date, read-more link
- Upserts page_component on the news-index page

Could also use the JSON pattern: `/data/news-listing.json` with client-side rendering on the listing page.

### Article rewriter (simple version)

The Tier 2 article rewriter takes a high-scoring feed item and produces a rewritten page:

1. Load the feed item (title, summary, source_url)
2. Load the site spec (identity, content_direction, audience)
3. Optionally `web_fetch` the source article for more context
4. LLM produces: headline, body (500-800 words), framed for site's audience
5. Create a `pages` row (type: `news-post`, URL: `/insights/{slug}.html`)
6. Save as `page_components` with rendered HTML
7. Update `content_feed_items.published_page_id`
8. Commit to git

Which items get rewritten? Configurable threshold — score ≥ 70 by default. Items with score 50-69 appear as link-out snippets on the homepage. Items ≥ 70 get full treatment.

---

## Tier 3: Research Analysis Pipeline

### What it produces

Not a rewrite of a single article. A daily (or per-event) analysis piece that:

- Clusters the day's news by theme/story
- Identifies the salient points in each cluster
- Researches the history and context behind each point
- Draws on the site's own event timeline for continuity
- Produces analysis with timeline graphs, infographics, and illustrated concepts
- Writes 1000-2000 word articles structured as: summary, background, timeline, implications

### Multi-agent pipeline

```
Daily analysis cycle:

1. news-analyst agent
   → loads today's relevant items from content_feed_items
   → LLM clusters items into 3-5 story themes
   → for each cluster, spawns a story-researcher

2. story-researcher agent (one per cluster)
   → identifies key claims/events in the cluster
   → web_search + web_fetch for background context
   → queries event_timeline for prior related events
   → produces research brief:
     {events, timeline, key_players, context, data_points}

3. analysis-writer agent (one per cluster)
   → receives research brief + feed items + site spec
   → LLM writes structured analysis:
     - Executive summary (2-3 sentences)
     - What happened (the news itself)
     - Background and context (from research)
     - Timeline of events (from event_timeline)
     - Implications for [site's audience]
   → requests visualizations (data series, timelines, relationships)

4. visualization-renderer (per visualization request)
   → data series → SVG line/bar chart
   → event list → SVG timeline
   → relationship data → SVG concept diagram
   → all static SVG embedded in HTML, no JS needed

5. page-publisher
   → creates news-post page at /insights/{slug}.html
   → writes events to event_timeline (for future analysis)
   → updates news-index listing
   → commits to git
```

### Event timeline database

The core enabler for continuity across analysis pieces. Each analysis cycle both reads from and writes to this table:

```sql
CREATE TABLE event_timeline (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    
    -- What happened
    subject TEXT NOT NULL,           -- "Iran-US conflict", "UK wholesale gas prices"
    event_date DATE NOT NULL,
    event_type TEXT,                 -- "price_movement", "policy_change", "conflict",
                                    -- "supply_disruption", "regulatory", "market_shift"
    headline TEXT NOT NULL,
    description TEXT,
    
    -- Structured data for graphs
    data_points JSONB,              -- {"gas_price_gbp": 136.30, "oil_price_usd": 94.30,
                                    --  "change_pct": 5.05}
    
    -- Provenance
    source_feed_item_ids UUID[],    -- which feed items reported this
    source_urls TEXT[],             -- original source URLs
    analysis_page_id UUID,          -- the insights page that covers this event
    
    -- Taxonomy
    tags TEXT[],                    -- ["energy", "geopolitics", "prices"]
    related_event_ids UUID[],       -- links to related events in the timeline
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by TEXT                  -- agent that created the entry
);

CREATE INDEX idx_et_site_subject ON event_timeline(site_id, subject, event_date DESC);
CREATE INDEX idx_et_site_date ON event_timeline(site_id, event_date DESC);
CREATE INDEX idx_et_tags ON event_timeline USING GIN(tags);
```

Over time, this builds a rich history per subject:

```
UK wholesale gas prices timeline:
  2026-02-28: Conflict begins. Gas at 98 GBp/thm.
  2026-03-05: Strait of Hormuz disrupted. Gas jumps to 115 (+17%).
  2026-03-12: Qatar LNG terminal strikes. Gas hits 128 (+31%).
  2026-03-20: Emergency reserves released. Gas stabilises at 130.
  2026-03-27: Diplomatic pause announced. Gas at 136 (+38%).
```

Each new analysis piece draws on this timeline and extends it. The tenth article about gas prices has much richer context than the first — "this is the fifth supply disruption in four weeks, and the largest sustained price increase since the 2022 European energy crisis."

### Visualization approach

All visualizations are static SVG embedded in the HTML page. No JavaScript charting libraries on the client. This fits the static site model and works on every device/browser.

The visualization-renderer agent (or action) takes structured data and produces SVG:

**Timeline chart:**
```
Input: [{date: "2026-02-28", value: 98, label: "Conflict begins"}, ...]
Output: SVG with axis, points, connecting lines, event annotations
```

**Price chart:**
```
Input: [{date: "2026-02-28", gas_gbp: 98, oil_usd: 68}, ...]
Output: SVG dual-axis line chart with legend
```

**Concept/relationship diagram:**
```
Input: {nodes: ["Iran", "Qatar", "UK", "LNG"], edges: [{from, to, label}]}
Output: SVG network diagram
```

These could be generated by a Go action using SVG templates (deterministic, fast) or by an LLM that outputs SVG (more creative but less consistent). The Go template approach is more reliable for data-driven charts; the LLM approach might work better for concept illustrations.

The SVG is stored in `page_components.rendered_html` as part of the article content. No separate asset files needed.

---

## Compatibility Check

Does the current design (Tier 1 build) block Tiers 2 or 3?

| Current design element | Compatible with Tier 2? | Compatible with Tier 3? | Notes |
|----------------------|------------------------|------------------------|-------|
| `content_sources` table | Yes | Yes | Unchanged — feeds items to all tiers |
| `content_feed_items` table | Yes | Yes | `published_page_id` links to insights pages. `topics` enables clustering. |
| `content_feed_items.status` lifecycle | Yes | Yes | Extends: `relevant → rewriting → published` or `relevant → analyzing → published` |
| Feed-triage agent | Yes | Yes | Triage evolves to rewriter (Tier 2), analysis input selector (Tier 3) |
| Homepage `latest-news` (JSON approach) | Yes | Yes | JSON file is independent of insights pages. `insights_url` link appears when pages exist. |
| `render_news_section` action | Yes | Yes | Outputs JSON instead of HTML. Checks for insights page. |
| Discovery checks | Yes | Yes | `stale_news_section` works regardless of tier. New checks can be added for stale analysis. |
| CSS snippets approach | Yes | Yes | News-post pages get CSS from snippets just like news-grid does |
| Component template in `content_components` | Yes | Yes | JS-based template for Tier 1. Article templates for Tier 2/3 are separate components. |
| `evaluate_news_feed` vertical map | Yes | Yes | Same recommendation feeds all tiers |
| `page_type` in pages table | Yes | Yes | Add `news-index` and `news-post` when needed |
| Planner integration | Yes | Yes | Planner adds homepage section (Tier 1) and nav link (Tier 2) |

**No blocking issues identified.** All three tiers build on the same data foundation (`content_feed_items`, `content_sources`) and add progressively more processing on top.

---

## Build Order

### Now (this session)

- Tier 1 with JSON approach for homepage
- `latest-news` component template (JS-based with noscript fallback)
- `render_news_section` outputs JSON to `/data/latest-news.json`
- `insights_url` slot in JSON (empty for now, links when Tier 2 exists)

### Next phase

- Tier 2: article rewriter (simple version — single item → rewritten page)
- `news-index` and `news-post` page types
- `rebuild_news_listing` action
- Navigation entry for insights
- Triage threshold config: which items get rewritten vs link-out only

### Later phase

- Tier 3: research analysis pipeline
- `event_timeline` table
- `news-analyst`, `story-researcher`, `analysis-writer` agents
- Visualization renderer (SVG timelines, charts, diagrams)
- Cross-analysis continuity (each piece builds on prior timeline)

### Each tier is independently deployable and testable. Don't build Tier 2 until Tier 1 is running well. Don't build Tier 3 until Tier 2 has proven the article quality is good enough.

---

## Resolved Decisions

43. **JSON file for homepage news, not page rerender.** Updating 6 news cards doesn't justify rerendering the entire homepage. A `/data/latest-news.json` file is committed to git and pushed to S3 — tiny commits, fast deploys, decoupled from page design changes. The homepage HTML contains a JS snippet that fetches and renders the JSON. Noscript fallback links to the insights page.

44. **Three-tier expansion path.** Tier 1 (homepage snippets with source links) → Tier 2 (rewritten articles at /insights/) → Tier 3 (research analysis with timelines and infographics). Each tier builds on the previous without rework. The current build (Tier 1) is designed to accommodate all three.

45. **Event timeline for research continuity.** An `event_timeline` table tracks structured events per subject per site. Each analysis cycle reads history and writes new events. Over time, the system builds institutional knowledge — the tenth article about a topic has much richer context than the first. This is the core differentiator for Tier 3.

46. **Static SVG for visualizations, not JS charting.** Timeline graphs, price charts, and concept diagrams are embedded SVG in the HTML. Fits the static site model, works everywhere, no client-side dependencies. Generated by Go templates (for data-driven charts) or LLM (for concept illustrations).

47. **Content writer generates headline, render action preserves it.** The content writer creates the section at build time with an LLM-generated headline tailored to the site. `render_news_section` reads the headline from `page_components.content_data` and includes it in the JSON output. The triage rewriter (Tier 2) can update the subheadline with contextual text ("Iran conflict drives gas prices to 18-month high") while preserving the stable headline.
