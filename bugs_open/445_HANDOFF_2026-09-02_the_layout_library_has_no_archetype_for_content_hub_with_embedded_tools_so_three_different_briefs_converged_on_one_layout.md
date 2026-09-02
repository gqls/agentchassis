# 445 — the layout library has no archetype for "content hub with embedded tools", so three different remake briefs converged on one layout

**Filed:** 2026-09-02, `site_design_planner` lane, in response to an owner critique
routed via the `designblog.co.uk` session: "the design is exactly the same as all
the other sites, it should be different." Full critique:
`docs/agent_docs/docs024_key_docs_latest/designblog_couk/CRITIQUE_2026-09-02_owner_site_review.md`.

**090 substitute** (owner ruling 2026-07-31, structural claim): no 090 run.
Substituted first-hand verification below — every count is a query a reader can
re-run, and the matcher's own scoring code (`fork_theme_composition.go`) was read
line-by-line, not inferred from behaviour.

## 1. The claim, measured directly

The three sibling "portfolio positioning" remakes named in the critique
(`designblog.co.uk`, `advertise.co.uk`, `websitepromotion.co.uk` — live 2026-09-02,
same lane, same day) **all resolved to the identical layout, `magazine-grid`**:

```sql
SELECT s.domain, l.name FROM sites s
JOIN style_collections sc ON sc.id=s.style_collection_id
JOIN css_themes t ON t.id=sc.css_theme_id LEFT JOIN layouts l ON l.id=t.layout_id
WHERE s.domain IN ('designblog.co.uk','advertise.co.uk','websitepromotion.co.uk');
-- websitepromotion.co.uk | magazine-grid
-- designblog.co.uk       | magazine-grid
-- advertise.co.uk        | magazine-grid
```

18 more remakes are queued in the same programme, so this concentrates rather
than dilutes without a fix.

## 2. Ruled out first: this is NOT primarily a matcher bug or a classifier-prompt artefact

Two more specific hypotheses were checked and set aside, because the evidence
argues against them being the main cause (recorded so nobody re-walks them):

- **Not the classifier literally naming the layout as a tag.** All three sites'
  `industry_tags` contain the string `"magazine-grid"` verbatim — traced to the
  classifier's own prompt, which lists `"magazine-grid"` as a *worked example* of
  a "site-shape tag" alongside `"tool-portal"`/`"affiliate-hub"`/etc. This looked
  like the smoking gun at first. It isn't: the `magazine-grid` layout row's own
  `industry_tags` (`publication, news, blog, opinion, long-form, editorial`) do
  **not** contain the string `"magazine-grid"` — `resolveLayoutByTags` never
  matches a site tag against a layout's own name, only against its `industry_tags`/
  `category`/`description`, so this particular tag contributes nothing to the
  score. A real prompt-hygiene issue (the classifier's few-shot examples double
  as the layout taxonomy's names, which invites exactly this kind of coincidental
  near-match), but not the mechanism that produced this result. Not filed
  separately — recorded here so the false lead isn't rediscovered.
- **Not a matcher that ignores differentiating signal.** All three sites' full
  tag sets (10 tags each) are genuinely different — SEO/marketing education vs.
  design publication vs. advertising/media — and the matcher did return different
  5-candidate shortlists per site (`resolved_composition.lineage.layout_candidates`),
  proving it evaluated the full library each time, not a cache or a fixed default.

## 3. What actually decided it: `category='editorial'` + description-word overlap, and there is exactly ONE such layout

Read `resolveLayoutByTags` directly (`fork_theme_composition.go:198-240`). Score
= tag-overlap (IDF-weighted) + a **category bonus** (fires when the layout's own
`category` matches the site's classified category, or appears in its tags) + a
**description-word bonus** (site tags matching words in the layout's prose
description) + a same-scheme bonus.

```sql
SELECT name, category, industry_tags FROM layouts WHERE is_active ORDER BY category;
```
Of 18 active layouts, exactly **one** carries `category='editorial'` with tags
fit for a professional/B2B publication: `magazine-grid`
(`publication, news, blog, opinion, long-form, editorial`). The only other
`editorial`-category layout, `soft-editorial`, is tagged for a different register
entirely (`wellness, lifestyle, bakery, artisan, personal-brand`) — a lifestyle
blog, not a professional content hub. All three sites are classified with
`category`/tags reading as professional editorial content
(`editorial`, `editorial-blog`, `editorial-hub`, `content-hub`, `content-platform`),
which correctly and heavily favours `magazine-grid` over `soft-editorial` on tag
fit — this is the matcher working as designed, not misfiring.

**The actual gap: none of the 18 layouts is built for what these three sites
structurally are — a content hub whose core offering is a set of embedded
interactive tools, presented with editorial framing.** The library forces a binary
choice between the editorial layouts (which have no tool-forward treatment) and
the `tool-portal-*`/`tool-first-landing`/`utility-tool` layouts (which have no
publication/content-hub framing). For a genuinely mixed shape — and per the
critique, the whole "portfolio positioning" programme is remaking sites of
roughly this shape, 18 more queued — there is one defensible answer today, and
three different briefs found it independently.

## 4. Secondary, fleet-wide finding: real concentration exists independent of this specific gap

`[MEASURED 2026-09-02]`, 37 deployed sites, grouped by chosen layout:

| layout | sites |
|---|---|
| `tool-portal-light` | 13 |
| `magazine-grid` | 8 |
| `brochure-formal` | 6 |
| `industry-hub` | 3 |
| `tool-portal-dark` | 3 |
| `social-lobby` | 1 |
| `brochure-bold` | 1 |
| `high-energy` | 1 |
| `ecommerce-storefront` | 1 |

**9 of 18 active layouts have never been chosen for any live site**
(`docs-sidebar`, `portfolio-kinetic`, `soft-editorial`, `technical-precise`,
`comparison-aggregator`, `affiliate-hub`, `tool-first-landing`, `utility-tool`,
`media-grid`). Three layouts account for **27 of 37 sites (73%)**. This is a
real, separate signal from §3 — some of it will be genuine fleet composition
(this estate does build a lot of tool-portal and brochure sites), some of it may
be the same "closest-available-archetype" pressure playing out across other
shapes too. **Not fully attributed here** — flagged as a fleet-wide pattern
worth its own look, not claimed as the same mechanism as §3.

## 5. What is explicitly OUT of this mechanism's scope, checked so it isn't guessed at

The critique also named identical top/bottom nav across sites. Checked: all
three siblings' `style_collections.header_component_id`/`footer_component_id`
are **NULL**:

```sql
SELECT s.domain, sc.header_component_id, sc.footer_component_id
FROM sites s JOIN style_collections sc ON sc.id=s.style_collection_id
WHERE s.domain IN ('designblog.co.uk','advertise.co.uk','websitepromotion.co.uk');
-- all three: header/footer NULL
```
Header/footer selection is a **separate mechanism**
(`link_site_components_action.go`, "Link site_components to content_components
from style collection" — its own doc comment: "Without this linkage,
renderAndStoreSiteComponent falls through to a hardcoded fallback that ignores
the style collection's templates"). `site-design-planner`'s `install_site_composition`
does not populate these FKs for these three sites, so the NULL-fallback path is
what's actually serving their chrome. The library itself has some header/footer
variety (`content_components` function names: `header-docs`, `header-with-categories`,
`header-with-search`, `header-with-cart-or-nav`, `header-minimal-tool`, 3×
`site-header` generic) — **not measured here** whether that variety is reachable
for these sites, since the linkage step never runs for them. Routed to the
`components` thread per the critique's own routing table (§3 of the critique
doc) — not investigated further from this file.

## 6. Fix candidates

1. **Add a layout archetype for the "content hub with embedded tools" shape** —
   the direct fix for §3, and the one the queued 18 remakes would benefit from
   most. A real design task (structure, CSS, header treatment for a tools
   showcase inside an editorial frame), not a config change.
2. **Diversify the classifier's few-shot "site-shape tag" examples** so they
   don't double as the layout taxonomy's own vocabulary (§2's near-miss) — cheap,
   doesn't fix §3 on its own (ruled out as the live mechanism), but is a real
   prompt-hygiene issue worth fixing on its own merits: an LLM given layout names
   as examples of "shape" will tend to echo them back, narrowing the tag space
   over time as more sites get classified this way.
3. **Investigate the 9 never-chosen layouts** — establish for each whether it's
   correctly unused (no site of that shape has been built) or reachable-but-losing
   (a real candidate that the weighted matcher keeps passing over). Not done
   here; a separate, larger piece of work than this file's scope.

Not recommended: treat §4's concentration as proof of a matcher defect without
per-layout attribution — it may simply reflect what kinds of sites this fleet
actually builds.

## 7. What this does NOT claim

- Does not claim the matcher is broken. §3 shows it picked correctly, given what
  the library actually offers.
- Does not claim the classifier-prompt issue (§2) caused this specific result —
  ruled out directly, kept in the file as a documented near-miss so it isn't
  re-investigated as if it were live.
- Does not touch chrome/header-footer selection (§5) — different mechanism,
  different owning thread, already routed by the critique.
