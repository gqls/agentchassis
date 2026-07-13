# STATUS — Affiliate sites: where we are, what we want, where to start

A companion to `STATUS_imagery_2026-05-12.md`. Written so that the affiliate
work doesn't disappear off the radar while imagery is the focus. **This is
not the active workstream right now** — it's a holding doc for the thinking,
ready to pick up.

---

## The big picture

The longer-term goal is sites like:

- **Boxing tickets** — events with dates, venues, line-ups. Affiliate links
  to ticketing platforms. Calendar / schedule centrality.
- **Darts gear** — traditional product affiliate. Reviews, comparisons,
  buying guides. Player history infographics as content draw.
- **Many others** in similar shapes (event-based, gear-based, lead-gen).

The unifying feature: each site contains products or events that come from
an external source (an affiliate program, a feed, manual curation), gets
editorial enrichment from us, and earns revenue when visitors click through.

A working affiliate site needs:

1. **Products / events data** in the database
2. **Rendering** that turns those rows into pages
3. **Editorial content** around them (reviews, guides, comparisons)
4. **Imagery** — product photos, illustrations, decorative graphics
5. **Calendars, schedules, infographics** for event-shaped verticals
6. **Affiliate link management** with disclosure and tracking

We have pieces of each. We don't have any of it end-to-end.

---

## What's already there

Surveyed 2026-05-12. More exists than I'd initially assumed.

### Database

- **`affiliate_products`** table — full schema in place. Cached fields
  (cached_name, cached_description, cached_price, cached_image_url, etc.)
  from the affiliate feed, custom fields (custom_name, custom_description,
  custom_pros, custom_cons, custom_verdict, custom_rating, custom_image_id
  FK→assets) for editorial overlay. Zero rows today.
- **`affiliate_programs`** table — exists. Empty.
- **`link_registry`** — exists, partially built per doc
  024_link_management_v2.md. Tracks individual affiliate link instances
  with provider, tag, requires_disclosure flag.
- **`assets`** with `asset_key` (Phase 2B) — supports per-product image
  storage via `asset_key = 'product_<external_id>'` pattern.

### Components (the rendering shell)

Five product-shaped components in `content_components`:

- `product-card-with-cta` (content category, 4438-char template,
  633-char schema, `source: query.affiliate_products`)
- `product-grid` (content, 3888/624)
- `product-hero_pre_037` (custom, 7402/3470, quality 100)
- `product-details_pre_037` (custom, 9250/5991, quality 100)
- `product-specs` (custom, 10862/2 — schema essentially empty)

The `product-card-with-cta` schema declares the array source as
`query.affiliate_products` — the component expects a resolver to query
the table. So rendering isn't waiting for new components; it's waiting
for the resolver that populates them. See "what's missing" below.

### Adjacent agent patterns

The `med-*` agent family (med-url-discoverer, med-price-collector,
med-url-mapper, med-url-discover-orchestrator, med-url-map-orchestrator)
exists for veterinary medicine pricing ingestion. It's a working model
for "scrape an external source, populate rows in our database, keep
fresh". Conceptually adjacent to "scrape an affiliate feed, populate
affiliate_products". Worth studying before designing the affiliate
feed ingester.

### Imagery infrastructure

Hero and logo generation works end-to-end. Variant pipeline works. The
product illustration plan (`PLAN_product_illustration.md`) is drafted
but parked behind the rendering question. When products do start
rendering, illustrations slot in via the `custom_image_id` field on
each product row.

### Disclosure & compliance

Per doc 024: `link_registry` has `requires_disclosure` for FTC/ASA
compliance markers, plus `affiliate_provider`, `affiliate_tag`. The
content writer agent has hooks for affiliate-specific prompt
instructions (disclosure placement, link format). Not yet exercised.

---

## What's missing

The shortest list of structural gaps before any affiliate site can ship:

1. **No affiliate program is wired up.** No Awin / Amazon / SkimLinks /
   Impact / Rakuten / SaleStock / whatever integration. No place to put
   credentials. No ingestion job to populate `affiliate_products`.
2. **No resolver for `query.affiliate_products`.** Component declares
   the source; nothing reads it. The render context build process
   ignores it. Needs an action like `resolve_component_data_sources`
   that reads the input_schema, sees `source: query.X`, and runs the
   right query.
3. **No page → component selection logic for product components.** When
   would `product-card-with-cta` appear on a page? Today this is
   planner-driven via the strategist's "components" array. We'd need
   the strategist (or its successor) to know when to recommend a
   product component.
4. **`product-specs` schema is effectively empty** (2 chars). The
   component exists but isn't usable until somebody fills in its
   input_schema. Worth checking against the template to see what fields
   are referenced.
5. **No editorial enrichment pipeline.** Today, when a product arrives
   via an affiliate feed, nothing turns `cached_name` into a polished
   `custom_name` or fills in `custom_pros`/`custom_cons`. Needs an LLM
   step.
6. **No calendar/schedule infrastructure** for event-based verticals
   (boxing tickets, fight cards, tournament dates). Probably needs its
   own table(s) plus a `event-card`-shaped component.
7. **No infographic generation** for player history, league standings,
   stat visualisations. Probably an LLM-generated SVG path or a
   chart-rendering action — open question.

---

## Three vertical shapes worth distinguishing

Affiliate sites aren't one shape. The pipeline you build first should
match the vertical.

### Shape A: Pure-product affiliate (darts gear, kitchen knives, etc.)

- Products come from an affiliate API
- Editorial overlay (reviews, comparisons)
- Renders as cards / grids / detail pages
- Imagery: product photos (from feed) + possibly illustrations (ours)
- This is what the existing 5 product components target

**Closest to working today.** Resolver for `query.affiliate_products` +
one ingestion job + some manual editorial = a v0 affiliate site.

### Shape B: Event/ticket affiliate (boxing, concerts, sports fixtures)

- Events come from a different kind of feed (Ticketmaster, StubHub,
  league APIs, sometimes RSS)
- Date / venue / line-up dominate
- Affiliate link is to "buy tickets for this event"
- Calendar / schedule view is central
- Editorial overlay (preview articles, predictions, after-reports)
- Imagery: event poster, fighter / artist photos, venue shots, results
  infographics

**Furthest from working today.** Needs event-shaped tables, calendar
components, and feed integrations that aren't in the codebase yet.

### Shape C: Lead-generation (mortgages, insurance, broadband deals)

- "Products" are services, not physical goods
- Visitor fills in a form; lead is sold to providers
- Comparison tables central; calculators very prominent
- Imagery: less product-shot, more reassuring/branded; iconography
- Tool components (calculators) already exist in the library

**Different shape, partly working.** The tool-calculator library
already handles mortgages, stamp duty, equity release, etc. The
lead-capture / lead-sale piece is the new bit.

---

## Layered approach when we start

Whichever vertical we pick, the same structural layers apply.

### Layer 1 — Get one product onto one page

Minimum viable.

- Manually insert one `affiliate_products` row
- Add a page-component that uses `product-card-with-cta`
- Build the resolver for `query.affiliate_products`
- See it render with `cached_image_url` as the image
- This validates the components, surfaces wiring gaps, costs one day

### Layer 2 — Ingestion + multiple products + editorial

- Wire one affiliate program (probably easiest: a CSV import for
  testing, then an API for real)
- Editorial enrichment step (LLM populates custom_name from cached_name,
  drafts custom_short_description, etc.)
- Page strategy: when does the planner add product components vs not?

### Layer 3 — Imagery story

Now there's something to display, illustrations become concrete. The
`PLAN_product_illustration.md` plan ships here. The resolver from
Layer 1 gets the precedence rule (`custom_image_id` first, else
`cached_image_url`).

### Layer 4+ — Verticals other than pure-product

Event-shaped sites (boxing tickets, fixture-based) need event tables
and calendar components. Lead-gen sites need form handling and lead
routing. Each is its own project.

---

## Open questions for the eventual decision

When we come back to this, these are the things to settle:

1. **Which vertical first?** Pure-product (Shape A) ships fastest given
   existing components. Event-based (Shape B) feels more interesting
   long-term but is most work. Lead-gen (Shape C) overlaps with
   existing tool infrastructure but adds a sales pipeline. Probably
   start with Shape A as the shortest path to a working site, then
   tackle one of the others.
2. **What does "looks good" mean?** Is the visual standard driven by
   product photos themselves (so stock + scraped imagery matters), by
   surrounding editorial content (hero imagery + decorative graphics),
   or by both? Different answers shape what we build.
3. **What's the legal margin for product imagery?** Reviewing
   copyrighted product photos is generally fine under fair dealing
   when accompanied by substantive editorial. Plain replication is
   risky. Illustration is safer but more work. Stock photos are safe
   but generic. This was the original driver of the illustration plan
   — worth keeping in view.
4. **What affiliate program(s)?** Awin and Amazon Associates are the
   obvious starting points in the UK. Each has its own API shape and
   compliance rules. Either works; pick one.
5. **Manual curation or feed-driven?** A site with 20 hand-picked
   products is very different from one with 5,000 auto-ingested
   products. Different rate limits, different review patterns,
   different editorial cost. Worth deciding before building either.

---

## References for when we pick this up

- `PLAN_product_illustration.md` — the illustration pipeline. Plugs
  into Layer 3 as a precedence rule in the resolver.
- `024_link_management_v2.md` — `link_registry` design, affiliate
  disclosure requirements.
- `FOCUS_imagery_assessment.md` section 3.2 — `affiliate_products` and
  `custom_image_id` notes.
- The `med-*` agent family in `bk_agent_definitions_backup.sql` — the
  closest existing pattern for "scrape external feed, populate our
  rows, keep fresh".
- Component audit results from 2026-05-12 (in chat transcript) — the
  five product components and what their schemas declare.

---

## When to come back to this

When imagery is at a clean stopping point (probably end of Phase 2G).
The first concrete affiliate work — building the
`query.affiliate_products` resolver — depends on the resolver design,
which interacts with how the planner emits page-component plans. That's
broadly settled but worth a fresh look when we get there.

Until then: not lost, not actively progressing.
