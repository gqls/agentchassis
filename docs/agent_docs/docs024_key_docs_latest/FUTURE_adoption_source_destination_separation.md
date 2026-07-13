# Adoption Flow — Source vs Destination Domain

**Date:** 2026-04-20
**Status:** Future work. Variants A-D captured. Option 1 sketched for
immediate implementation.

---

## Problem

Today the adoption flow conflates two concepts into a single `site_id`:

- **Target** — the URL being crawled for reference
- **Destination** — the site being built

If you submit `https://competitor-example.com` as adoption input, the
system crawls it AND builds the new site as `competitor-example.com`.
There's no way to say "crawl this site as inspiration, build my own
site called `my-new-site.com`."

This matters for two legitimate workflows:
1. **Mirroring your own sites** — e.g. you have a working design on
   `site-a.com` and want `site-b.com` to look the same (complete copy).
2. **Competitor-informed builds** — crawl a competitor's site, take the
   best parts, build a destination that isn't a clone.

---

## Variants considered

### Variant A — reference-only adoption

Crawl target X purely for design inspiration. Build destination Y from
scratch. No content from X carries over. Y has its own brief, mission,
pages.

**Specs that carry over:** `design_reference` (fingerprint)
**Specs that do not carry over:** identity, content_direction, pages, archetype
**Use case:** "My site should look like Stripe's but it's a different business"

### Variant B — design + structure adoption

Crawl X, carry over its design fingerprint AND its page structure
(number of pages, nav organisation, section types). Content is rewritten
for Y's brief.

**Specs that carry over:** design_reference, archetype, pages (structure)
**Specs that do not carry over:** identity, content (tone/voice is up to Y)
**Use case:** "Our competitor has a great structure for this vertical;
build us a site with the same pages but our content"

### Variant C — full clone with substitution

Crawl X, reproduce design, structure, AND content voice/tone on Y.
Rewrite references to X-specific entities (company name, services).

**Specs that carry over:** design_reference, archetype, pages,
content_direction (voice/tone guide)
**Specs that do not carry over:** identity (destination has its own)
**Use case:** "I own site-a.com and want site-b.com to be a clone with
different branding"

### Variant D — multi-source competitor analysis

Crawl multiple targets (X₁, X₂, X₃), extract patterns, use as aggregated
input to Y's planning. Not a mirror of any single source.

**Specs that carry over:** an aggregate `competitor_landscape` aspect
with patterns across sources
**Use case:** "Build us a site informed by these three competitors but
don't copy any one of them"

---

## Options for implementing

### Option 1 — Parameterise inputs (SIMPLEST, IMMEDIATE)

**What it is:** add a `destination_domain` input alongside the existing
`url` / `target_url`. `ensure_site_record` uses destination; `crawl_site`
uses target.

**Schema change:** none
**Migration:** none
**Reversible:** yes (remove input, fall back to default where dest=target)

**Limitations:**
- Same target crawled multiple times = repeated work
- No visibility of "who adopted from whom" in admin UI
- Target never appears as a first-class site record (good if crawling
  competitors you shouldn't store; bad if you want provenance)

**Good fit for:** all four variants if you're happy to re-crawl each
time. No-frills, unblocks the immediate "build my own site styled like
this other one" use case today.

### Option 2 — Two-site model with `source_site_id`

**What it is:** new column `sites.source_site_id UUID REFERENCES sites(id)`.
Source sites exist in `sites` but flagged (via new column or
`status = 'adoption_source'`) as reference-only — never deployed, never
get pages built or work items created against them (besides crawl).

**Schema change:** one migration, nullable column
**Reversible:** yes (drop column, no data loss)

**Benefits:**
- Clean provenance
- Same crawl reusable across multiple destinations
- Downstream agents can query `source_site_id` to read reference specs
- Admin UI can show "built referencing site X"

**Limitations:**
- Source sites fill the admin UI with entries that don't correspond to
  real deployments (fixed by filtering on `status`)
- Need to decide what "refresh" means for a source (re-crawl and replace
  specs vs version-and-keep-old)

**Good fit for:** Variants B, C, D when crawling-once and building-many
becomes common.

### Option 3 — Reference library (separate table)

**What it is:** new table `adoption_references` (url, fingerprint,
content_samples, archetype, created_at). Independent from `sites`. A
site can reference an `adoption_reference_id` via a new nullable column
on `sites`.

**Schema change:** one migration creating new table + one FK column on `sites`
**Reversible:** yes (new structures, old path still works)

**Benefits:**
- Cleanest conceptual separation
- References can be browsed as a library ("what competitor looks do we
  have on file?")
- No confusion between source-sites and destination-sites

**Limitations:**
- Biggest change: new CRUD, new admin UI surface, adoption-crawler becomes
  its own agent
- Needs thought on "what if the reference site structure evolves?"

**Good fit for:** Variant D especially, and as the end-state if adoption
crawls become a regular curated activity.

---

## Recommended approach: stepwise

### Phase 1 — ship Option 1 (immediate)

Scope: Variant A + C (you said "some of my own sites want a complete
copy, some competitor sites want the best bits"). Variant B and D fall
out for free once the destination_domain plumbing exists.

**Agent changes (single SQL patch to `site-adoption-agent` definition):**

Add `target_url` and `destination_domain` as input fields; keep `url` and
`domain` as fallbacks for backward compatibility. Modify workflow:

- `crawl_site` config: `url_field: input_data.target_url` (with fallback
  to `input_data.url`)
- `ensure_site_record` config: pass `destination_domain` explicitly to
  override what `extractDomainFromInput` would default to

**Go changes:**

1. `extractDomainFromInput` in `ensure_site_record_action.go` — add a
   new helper `extractDestinationDomain` that prefers
   `input_data.destination_domain` over the default domain extraction.
   Leave `extractDomainFromInput` untouched (used by many other agents).

2. `apply_adoption_plan_action.go` — add a variant selector:
   - Read `input_data.adoption_variant` (values: `reference`, `structure`,
     `clone`, `analysis`; default `clone` for backward compat)
   - Branch spec-writing based on variant:
     - `reference`: write only `design_reference` + `design_intent`
     - `structure`: + archetype + pages
     - `clone`: + content_direction (the current behaviour)
     - `analysis`: nothing written as primary specs; store as
       `competitor_landscape` aspect to be merged later

3. Optionally: `crawl_site` step gains a `target_url` resolution helper so
   if someone passes just `target_url` the URL is normalised.

**Input shape:**

```json
{
  "url": "https://example-reference.com",   // legacy, still accepted
  "target_url": "https://example-reference.com",
  "destination_domain": "my-new-site.com",
  "adoption_variant": "clone"
}
```

**Backward compat:** if `destination_domain` and `target_url` are not
set, fall back to old behaviour (url=target=destination). Existing
submissions keep working.

**Effort estimate:** 1 SQL patch, 2-3 Go action files changed, ~100
lines total. Under a day.

### Phase 2 — add provenance (Option 2)

When Phase 1 is in use and the need for "crawl once, build many" or
"show me where this site's design came from" emerges, add the
`sites.source_site_id` column and modify the crawl to create/reuse a
source-site row.

### Phase 3 — reference library (Option 3)

When competitor analysis becomes a curated activity — someone
periodically crawling a set of inspirations independent of any specific
destination build — move to the dedicated table. Source sites migrate
into adoption_references.

---

## Phase 1 — concrete change list (for implementation when ready)

### File 1: `agent_definitions` UPDATE for `site-adoption-agent`

In the workflow JSON, change two steps:

**`crawl_site`** — from `"url_field": "input_data.url"` to
`"url_field": "input_data.target_url"` with a Go-side fallback reading
`input_data.url` if `target_url` is empty.

**`ensure_site_record`** — add a new config field
`domain_override_field: "input_data.destination_domain"`. If present,
`EnsureSiteRecordAction` uses this value instead of `extractDomainFromInput`.

### File 2: `ensure_site_record_action.go`

Add logic:

```go
// Check for explicit destination_domain override first
if override, ok := params.StepConfig.Config["domain_override_field"].(string); 
   ok && override != "" {
    if explicitDomain := datahelpers.ExtractNestedFieldString(
        params.CollectedData, override); explicitDomain != "" {
        logger.Info("EnsureSiteRecordAction: using explicit destination domain",
            zap.String("destination", explicitDomain))
        domain = explicitDomain
    } else {
        // Fall through to extractDomainFromInput
    }
}
if domain == "" {
    domain = extractDomainFromInput(params.CollectedData, params.Logger)
}
```

### File 3: `apply_adoption_plan_action.go`

Read `adoption_variant` from `input_data.adoption_variant`. Default
`"clone"` for backward compat. Gate spec writes on variant:

```go
variant := datahelpers.ExtractNestedFieldString(
    params.CollectedData, "input_data.adoption_variant")
if variant == "" {
    variant = "clone"
}

// Always write
writeDesignReference(...)
writeDesignIntent(...)

// Variant-gated
switch variant {
case "clone":
    writeIdentity(...)    // current behaviour
    writeContentDirection(...)
    writeArchetype(...)
    writePages(...)
case "structure":
    writeArchetype(...)
    writePages(...)
    // NO content_direction, NO identity
case "reference":
    // Nothing more
case "analysis":
    writeCompetitorLandscape(...)  // new, stored as aggregate aspect
    // Spec is informational; does not drive a build on its own
}
```

### File 4 (optional): new workflow step `choose_destination_brief`

For the `clone` variant, destination needs its own identity/mission. We
can either:

(a) Require the caller to pass `input_data.destination_brief` as input
    (simplest)
(b) Add a step that runs the briefing agent for the destination domain
    before `apply_plan` (more work, more flexible)

Start with (a).

### File 5: new `firecrawl_crawl` step input handling

In the Go code for `FirecrawlCrawlAction`, handle the case where
`url_field` resolves to empty:

```go
url := datahelpers.ExtractNestedFieldString(params.CollectedData, urlField)
if url == "" && urlField != "input_data.url" {
    // Fallback to legacy field
    url = datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.url")
}
```

---

## Open questions to answer when picking this up

1. **Destination's identity**: does it come from the caller
   (`destination_brief` in input), or do we run briefing-agent as a
   pre-step? Phase 1 says caller-provided; that keeps the change small.

2. **Variant persistence**: should `adoption_variant` be stored on the
   destination site (so we know how it was adopted)? If yes, new column
   `sites.adopted_via` text, or store in `sites.settings` jsonb.

3. **Re-crawl strategy**: if the target site changes, does the
   destination get updated? Out of scope for Phase 1; assume "crawl once
   at build time."

4. **Variant D (analysis) storage format**: competitor_landscape spec
   structure is TBD. Aggregate of palette_moods, archetype_labels,
   common_structural_patterns? Defer until first use case arrives.

5. **What if source and destination are in different industries?** The
   crawled archetype says "industry-X" but destination is industry-Y.
   Currently nothing enforces consistency; should we add a warning? Or
   trust the user?

---

## Risks

- **Data bleed:** if spec-write code isn't properly gated on variant,
  competitor content could leak into a destination. Mitigate with unit
  tests of `apply_adoption_plan_action.go` that assert variant gating.

- **Domain confusion:** if `destination_domain` is typoed, we could
  create a new site record for the typo'd domain (e.g. `my-new-siet.com`).
  Add validation: require a confirmed domain (DNS check or whitelist)
  OR require an explicit `site_id` of an existing record.

- **Content_direction from competitor:** variant `clone` copies content
  voice. If the destination is legally required to differ (regulatory,
  trademark), the clone variant is inappropriate. Mark this as a user
  responsibility; provide docs.

---

## Relationship to existing work

- Complements the composable theme migration (doc 025). If adoption
  creates a `design_reference`, it could feed directly into palette +
  layout + typography selection for the destination.

- Depends on the fingerprint extraction pipeline being healthy (see
  `FOCUS_design_and_styling_adoption_*` docs — there are still open
  issues with `enrich_fingerprint_with_css` path bug and `analyze_site`
  JSON truncation). Phase 1 can ship without these fully resolved,
  because Variant A (reference-only) doesn't depend on a perfect
  fingerprint — even a partial one is useful.

- Conflicts with nothing in the current Track 2 work. Ready when you are.
