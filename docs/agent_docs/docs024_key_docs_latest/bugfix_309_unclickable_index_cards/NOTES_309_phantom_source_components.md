# NOTES — bugfix 309 (append-only, newest at the bottom)

## 2026-08-18 — session "bugfix bugs_open/297" picks up 309

- Asked to take 297; it is CLOSED (fixed+live 2026-08-17, mig 453, bugs_closed/).
  Swept 298–310 with `who-owns.py` + live-transcript greps: 298–308 owned by active
  lanes; 310's file is untracked in the tree (its filing session f7646672 still
  active on it); 309 filed by the 279/284/290 thread whose session (fca1cedb) fired
  a 090 (corr `df8ca3a1-9cca-474a-88fb-19577e088080`), wrote its handoff, and
  ENDED. 309 is the pick.
- **Bug re-validated at the served page** (~19:25 UTC): 200, 32,594 B, 6 bl-cards,
  0 anchors in each. Matches the filing exactly.
- 090 dispatch not yet visible in orchestration_states (payload query, 0 rows) and
  0 diagnosis_artifacts — consistent with the measured ~29 min publish→start
  latency; did NOT re-fire.
- `bl-card` appears in NO Go file — the markup is `content_components` data.
  The page's listing pc `79d769e4…` → component `blog-listing_pre_037`
  (`4b097683…`), `created_from='generated'`, born 2026-04-08, template stores
  anchors as `{{if .postN_url}}<a href="{{.postN_url}}"…>{{end}}`.
- `content_data` for that pc holds all six posts' titles/dates/excerpts/images and
  **no `postN_url` key at all** (65 keys, none `_url` except image URLs). So the
  template's `{{if}}` ate every anchor. Renderer behaved exactly as told.
- Schema declares `postN_url` `required:true`, `source: site_specs.blog.postN_url`.
  **`SELECT … FROM site_specs WHERE aspect='blog'` → 0 rows fleet-wide, all
  history.** The source has never been resolvable for any site since the component
  was born.
- `plan_sections_action.go`: `on_missing` defaults to `skip_field` (line ~2093) —
  required + skip_field ⇒ field omitted, section builds, structural miss recorded
  (`STRUCTURAL_KEY_CARRY_MISS` in agent_error_log — 28 rows fleet-wide, none for
  this page; the page's last builds may predate the mechanism or have gone through
  the merge-path rerender. Not load-bearing for the fix; the 090 may say).
- **The bug file's control does not control what it seems to**: the working
  `mortgagecalculator.co.uk/investor/index.html` uses a DIFFERENT component
  (`tool-list`), not blog-listing. So "the component is capable" was true of the
  card idiom, not of this code path. `blog-listing_pre_037` has plausibly never
  produced a working link anywhere.
- Second consumer `leopardessconsulting.co.uk/blog`: its pc points at the same
  component but its stored rendered_html (8,712 B) contains **zero bl-cards** — a
  plain link list; some other writer replaced it. Do not assume broken-the-same-way;
  verify before touching. [UNVERIFIED what serves live on that URL]
- **Census** (queries in RUNBOOK): 10 phantom `site_specs.<aspect>` vocabularies
  across 11 active components (blog, categories, inventory, legal, nav ×16 fields,
  pricing, product, search, social, social_proof — aspect exists on NO site).
  Live exposure: this page, leopardess/blog, and 3 pcs on gaswholesalers.com
  (testimonials, social_proof). 8 components dormant. Plus 7 declared `query.*`
  names the resolver's switch does not know (affiliate_products, category,
  category_posts, comparison_filter_types, comparison_results, featured_post,
  bare `pages`) — same silent-skip fate at plan time.
- The store gate (`store_generated_component_action.go`) has checks 1–4 (HTML
  structure, unclosed style, empty schema, legacy dialect) — and no source
  validation. That is the class gap. `recordValidationRejection` already exists as
  the feedback channel for refused generations.
- `content-listing` (manual, 2025-11-28, active) is the correct article listing:
  `articles` ← `query.blog_posts`, required, `on_missing: skip_section`, range
  loop. fundamentallyai has 9 `page_type='blog-post'` pages (8 active+deployed,
  1 archived) — `query.blog_posts` returns exactly the right 8 and drops the
  archived card-4 target naturally.

### Missteps so far

- My opening `ls bugs_open/ | grep -i 297` returned empty and I nearly concluded
  the file was missing; a combined `ls` of both dirs found it in `bugs_closed/`.
  The first command's empty output remains unexplained — treat a surprising empty
  grep as suspect and re-ask differently (grep-silent traps are a known family).
