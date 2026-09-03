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

---

# 8. WORKED 2026-09-03 by the `bugs_open/445` thread — the mechanism underneath §3

**Taken at an owner ruling relayed via the `designblog.co.uk` lane: "a thread has taken bug 445",
which puts BOTH the detector work and the missing archetype's design on this thread.**

Lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_445_layout_fit/`.

**§1-§7 above are not retracted.** §3 is correct that the library has no archetype for this shape,
and §2's two rule-outs both survived re-testing (see §8f). What follows is the layer beneath: *why
nobody found out until an owner looked at three finished sites.*

## 8a. The estate could not SEE a library gap of this shape, by construction

`queueLayoutCandidateReview` was reached only when `resolveLayoutByTags` set `IsFallback` (no
layout scores above zero **anywhere in the library**) or `IsSchemeMismatch`. But the category
(0.75), description (≤0.90) and same-scheme (0.50) bonuses are added to `total` **independently of
any tag matching**, so `total > 0` is satisfiable with `tagScore == 0`.

`[MEASURED 2026-09-03]` **Four live sites are recorded by that code as `tags 0.00` AND lineage
`layout_source: library_match`** — a layout matching *none* of the site's tags, recorded as a
successful library match:

| site | layout | the system's own recorded reasoning |
|---|---|---|
| webdesign.uk | brochure-bold | `score 0.75 (tags 0.00)` — exactly `lmCategoryMatchBonus`, alone |
| farmerinsurance.uk | industry-hub | `score 0.90 (tags 0.00)` |
| garden-tools.uk | industry-hub | `score 0.90 (tags 0.00)` |
| vetcomparison.uk | industry-hub | `score 0.90 (tags 0.00)` |

`[MEASURED 2026-09-03]` **Exactly TWO `needs_new_layout_candidate` items exist across 63,007 work
items ever written** (29,657 live ∪ 33,350 `site_work_items_archive`, 2026-02-22 → 2026-09-03) —
robot-hands.com 2026-07-08 (`wont_fix`) and ai-agent-orchestration.com 2026-08-12. **Both carry
`reason: "fallback — no classification tags"`: the degenerate no-tags-at-all arm.** So the number
of times this mechanism has assessed the library and reported it short is **zero**.

## 8b. The remedy was specified in April 2026 and never built

`docs/agent_docs/sql_for_agents/103_site_design_planner.sql:175-215` defines the
`resolved_composition` lineage contract, including three things that did not exist:

| specified | reality on 2026-09-03 |
|---|---|
| `layout_match_score` — *"(float 0-1) — tag-overlap score for chosen layout"* (**normalised**) | **0 of 33** rows carry the key |
| a threshold — worked example: *"scored utility-tool=0.82 above threshold 0.5"* | no threshold anywhere in the code |
| `layout_source: "needs_new_layout_candidate"` | **never written**: 31 `library_match`, 2 `library_fallback` |

Absent a normalised score there was no quantity a threshold could be applied to, so the signal was
left firing on the only thing available — total-is-zero — which is not a measure of fit.

## 8c. §4's concentration has a mechanism, and it is four strings

`[MEASURED 2026-09-03]` The layout library's canonical tag vocabulary is **101** terms. Across 33
composed sites the classifier emitted **216** distinct terms. The overlap is **28**, and four of
them decide nearly everything: `interactive-platform` (18 sites), `tool-portal` (15),
`editorial-publication` (12), `interactive` (12). **188 of 216 (87%) match no layout at all.**

Hence the identical scores §1 noticed: the seven `magazine-grid` sites all scored **3.05 (tags
2.30)** to 2dp because 2.3026 = `log(1+18/2)` — one tag present in exactly two layouts — plus the
0.75 category bonus. They are not similar sites; they share **one attractor tag** and it addresses
**7-10%** of each site's declared identity.

**Method note, because the number is load-bearing:** the scorer was re-implemented independently in
Python and **reproduces the system's own recorded score exactly on 29 of 30 scored sites** (the one
miss, gamedesign.uk, had its classification refreshed after composition). That replication is the
control for every figure in §8.

## 8d. The classifier was told to rely on a mechanism that did not exist — *and shown an empty list*

The live prompt said, verbatim: *"If no existing tag fits this site well, coin a new one using the
same style — an unmatched tag will trigger a library-growth review work item rather than silently
fail."* It did not.

**And the `portfolio_positioning` lane then found the half I had missed, which is the mechanical
cause of the 87%.** `read_layout_taxonomy` fetched the live vocabulary into `collected_data`, but
`classify_and_extract`'s `input_fields` allow-list did not name `layout_taxonomy`, so it was
**dropped at the template boundary**. Verified by me at the rendered artefact in
`llm_call_log.prompt_rendered`:

```
Current library tags (match these when they describe this site — each match is an overlap point for the matcher):
null

The library currently has <no value> active layouts. If no existing tag fits this site well, coin a new one...
```

**`null`, and `<no value>`.** The model was told to match against an empty list and then told to
coin. The 87% is not carelessness; it is the only thing available to it.

**Two independent broken links in one loop, and either alone would have been survivable:** link 1
(theirs) produced the coined vocabulary; **link 2 (this file's) is why nobody noticed for months**,
because 87% of tags vanishing is exactly the condition the review item existed to report.

## 8e. Fixed

| what | where | state |
|---|---|---|
| `layout_taxonomy` reaches the prompt | migration **734**, `portfolio_positioning` | **APPLIED 2026-09-03 11:39:14Z** (their work, their finding) |
| Fit measured, recorded, and the signal fired on it | commit **`76db94fc7`** | committed; **inert until the next chassis roll** |
| Prompt stops promising a mechanism, stops offering layout names as tag examples, prefers FORM over industry words | migration **735** | **APPLIED + verified at the live row 2026-09-03** |

`76db94fc7` computes `TagCoverage` (103's normalised `layout_match_score`), records it with the
threshold in force plus matched/unmatched terms and runner-up, and widens the predicate to
`fallback | scheme mismatch | weak tag fit` at a 0.50 cut taken from the measured distribution's
**widest empty band (38%–62%)**. **Selection is unchanged — no site's layout moves.**
Council: `Council-Submitted: 34d57f60-7013-4fa4-8106-e8d8e5e29887`.

**Pre-registered disconfirmation, so the threshold can be contradicted rather than defended:** 734
should raise coverage fleet-wide from 11:39Z. If compositions land inside the empty 38%–62% band,
the cut was a 33-site artefact and must be re-derived. `lineage.layout_fit.threshold` records the
cut per row so the re-derivation stays honest.

## 8f. §2's rule-outs re-tested, and one extended

A peer lane re-raised §2's first rule-out with a 12-site census (8 × `magazine-grid`, 4 ×
`affiliate-hub` carrying a layout name in `industry_tags`). **§2 stands, and I checked the path §2
did not:** the description bonus tests for `" magazine-grid "` **or** `" magazine grid "`, and
magazine-grid's description opens *"Publication layout with featured article…"*. Zero contribution
on **all three** paths (own tags, category, description). Same for `affiliate-hub`. It remains a
real prompt-hygiene problem — fixed in 735 — and it remains inert in the scorer.

## 8g. The archetype (§6 candidate 1) — still owed, now evidence-led, and partly disconfirmed

The 7-site cluster's shared **unmatchable** terms are `content-hub` (3 of 7) and `interactive-tools`
(2 of 7). §3 named the missing archetype "content hub with embedded tools" from three examples;
the sites' own emitted vocabulary says the same thing from the other end, and supplies the tag path
that would make a new layout *reachable* — **9 of 18 existing layouts have zero sites emitting any
of their tags**, so an 18th drawn without one becomes the tenth unreachable layout.

**Simulated before recommending.** Adding a hypothetical `content-hub-tools` layout breaks the
cluster up — websitepromotion 8%→28%, relojistas 10%→22%, homegarden 8%→17%, advertise 7%→16% —
**but designblog.co.uk (7%→6%) and apis.uk (9%→8%) still win on a single tag.** For those two it is
a different near-miss, not a fit. **Do not close 445 on "archetype drawn, problem solved":** the
site that started this is one of the two the archetype does not rescue. The tag set is
load-bearing and was invented by the simulator, so the real archetype's tags must be chosen by
simulating candidates against the live fleet — including the 17 unbuilt remakes
(`portfolio_positioning/CONTRIB_2026-09-03_seventeen_remaining_remakes_for_tag_simulation.md`),
since fitting only to 33 built sites is how you produce a tag set that looks good and helps nobody.

## 8h. Still open

1. **The archetype itself** (§8g) — this thread's, per the owner ruling.
2. **A fleet fit sweep.** `76db94fc7` fires at composition only. `[MEASURED 2026-09-03]`
   `SelectStyleCollectionAction` (`v3_site_actions.go:67`) points a site at an existing style
   collection and writes **no `resolved_composition` at all**, and render-time `theme_id`/
   `theme_name` overrides persist nothing — only 33 of 38 sites with a layout have a lineage row.
   A sweep must key on `sites → style_collections → css_themes → layouts` and ask "is this site
   well served by the layout it *has*", independent of how it arrived. Planned behind
   `internal/cronchecks` (owner decision 2026-09-03), answering the open `RFC_024` rather than
   becoming the tenth copy of an un-harnessed cron check.
3. **`RFC_037` is complementary, not a substitute.** Its fix makes the classifier *say* something
   different; this one makes the difference *reachable*. Perfectly differentiated tags still match
   nothing if the vocabulary cannot reach a layout. Recorded as an addendum in that RFC.
4. **§4's "9 never chosen" is now split:** 5 correctly unused (never score above zero for any site
   — `affiliate-hub`, `comparison-aggregator`, `docs-sidebar`, `media-grid`, `portfolio-kinetic`)
   and 4 reachable-but-losing, of which `soft-editorial` scores above zero on 27 of 33 sites **on
   the same-scheme bonus alone, zero tags**.
