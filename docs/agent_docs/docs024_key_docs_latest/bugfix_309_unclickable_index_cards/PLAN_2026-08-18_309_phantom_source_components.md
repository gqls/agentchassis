# PLAN — bugs_open/309: unclickable index cards, and the phantom-source component class behind it

**Started 2026-08-18** by the session named "bugfix bugs_open/297" (297 turned out closed;
309 was the next unowned open bug — filed by the 279/284/290 thread, whose handoff
explicitly leaves it: "read the 090 verdict (df8ca3a1), then fix the unclickable index").

## What the bug is (measured, this session, 2026-08-18 ~19:25 UTC)

`fundamentallyai.com/platform-log/index.html` renders 6 `<article class="bl-card">`
cards with ZERO `<a>` elements — re-confirmed at the served page this session (HTTP
200, 32,594 B, 6 cards, 0 anchors each). Bug still valid.

## Root cause (first-hand, evidence inline; 090 df8ca3a1 in flight as cross-check)

The chain, each link verified against live DB + code this session:

1. The page's listing section is `blog-listing`, resolved to component
   `blog-listing_pre_037` (id `4b097683-5b95-47dd-a598-8d6829296a75`,
   `created_from='generated'`, born 2026-04-08). Its input_schema declares
   `post1_url…post6_url` as `{"type":"url","source":"site_specs.blog.postN_url","required":true}`
   and `cta_url` as `site_specs.blog.archive_url`.
2. **No site in the fleet has ever had a `site_specs` aspect named `blog`**
   (`SELECT domain FROM site_specs WHERE aspect='blog'` → 0 rows, all history).
   The declared source has been unresolvable since the component's birth.
3. `plan_sections_action.go` resolve(): literal spec path misses (`resolveSpecPath`,
   aspect absent), alias chain misses (aliases cover identity/contact facts only),
   the 238 carry misses (the stored row never had a URL to carry). `on_missing`
   defaults to `skip_field` even for required fields (line ~2093), so all six URLs
   are omitted and the section builds anyway.
4. The template gates each anchor on `{{if .postN_url}}` → six complete-looking,
   link-less cards. Silent at every stage.

## The class, not the case (census 2026-08-18, queries in RUNBOOK)

- **10 phantom aspects across 11 active components** declare `site_specs.<aspect>`
  sources for aspects that exist on NO site: blog, categories, inventory, legal,
  nav (16 fields), pricing, product, search, social, social_proof.
- Live usage of those components: `blog-listing_pre_037` (2 pages: the 309 page +
  leopardessconsulting.co.uk/blog), `testimonials` + `social_proof` (3
  page_components on gaswholesalers.com). The other 8 are dormant in the selector —
  latent 309s.
- **Second latent class, same shape:** 7 `query.*` names declared in active
  component schemas that `queryresolve.Resolve`'s switch does not know
  (affiliate_products, category, category_posts, comparison_filter_types,
  comparison_results, featured_post, bare `pages`). An unknown query name errors at
  plan time and the field goes through the same silent skip_field path.
- The store gate (`store_generated_component_action.go`) validates template shape,
  truncation, and schema dialect — **it never validates that a declared source is
  in the resolver's vocabulary or points at a store that can exist.** That is the
  door this class walked through.

## Decisions

- **D1 (framework fix first):** the durable fix is a source-vocabulary validation
  at component birth, not a data patch to one page. Preferring the mechanism that
  makes the bad state unrepresentable (order-fix-candidates rule).
- **D2 (case repair uses existing machinery):** `content-listing` (manual,
  2025-11-28) is the framework-native article listing — `articles` from
  `query.blog_posts` (required, on_missing=skip_section, range loop). The worked
  case is repaired by swapping the section to it and rebuilding, NOT by patching
  URLs into content_data (which the next regeneration would destroy — bugfix 238's
  lesson) and NOT by hand-editing rendered_html (framework rule, owner 2026-08-04).
  `query.blog_posts` lists this site's 8 active deployed blog-post pages; the
  archived `ai-readiness-checker-guide` is excluded naturally, which also resolves
  the bug file's card-4 concern.
- **D3 (scope):** phase 1 = guard + case repair on the 309 page. leopardess/blog and
  gaswholesalers are the same class but different serving states — verify before
  touching (leopardess's stored rendered_html for the same component is NOT the
  bl-card markup at all; something else rewrote it — do not assume it is broken the
  same way).

## Phases

1. **Research + census** — DONE (this file, NOTES).
2. **090 cross-check** — verdict for df8ca3a1 pending dispatch latency; read before
   asserting the root cause in the bug file as diagnosed. My census stands as
   first-hand verification either way (the 2026-07-31 owner ruling's named escape
   hatch, stated here: the mechanism was verified at every link against live DB and
   code, and an independent 090 with the same symptom is already in flight from the
   filing thread — a second run would be a duplicate round).
3. **Framework guard** — validate field sources in `store_generated_component_action.go`:
   - unknown source prefix → reject (message names the vocabulary);
   - `query.<name>` not answerable by queryresolve → reject; export the known-name
     check from the queryresolve package so the vocabulary lives in ONE place;
   - `site_specs.<aspect>` where the aspect has never been written for any site →
     reject, message lists the real aspects and points listings at `query.*`.
   Council-gate the change (platform code). Register the check in the concept
   register in the same commit (ordering-exemption condition 2).
4. **Case repair** — swap platform-log-index's section `blog-listing` →
   `content-listing`, rebuild the page through the pipeline, verify at the served
   page (all six cards → anchors resolving 200; archived guide absent).
5. **Class follow-ups** — file/attach findings for gaswholesalers + the 7 unknown
   query names + the 8 dormant phantom components (census report in the bug file;
   separate work items where warranted — not silently widened into this fix).
