# HANDOFF — news editorial features, 2026-08-21. START HERE.

Written at a context checkpoint. Everything below is measured unless marked
otherwise. **Read §1 and §9 before doing anything.**

Fleet state at handoff: chassis `v1.0.1322`, pods started 2026-08-21 16:54Z.
**Nothing in this lane was ever pending a roll** — all of its work is DB
migrations and docs, which are live on apply. The build changes nothing here.

---

## 1. What this lane is, in one paragraph

Editorial feature articles: a page about one current news story, built from the
site's **own feed** where several channels cover the same story, with the
background charted from **cited, registered facts**. Not a headline list. The
pattern is registered as **`NEWS-020`** and is deliberately built from existing
components — nothing new was created to ship it.

Companion lane for how these pages *look*:
`docs024_key_docs_latest/editorial_design_uplift/`.

## 2. What is live

| URL | site | notes |
|---|---|---|
| `/insights/index.html` | robot-hands.com | the hub, **in the top nav**, lists both features |
| `/insights/robot-demand-step-change.html` | robot-hands.com | feature 1 (IFR installations) |
| `/insights/electric-vs-pneumatic-economics.html` | robot-hands.com | feature 3 (compressed air) |
| `/insights/darts-calendar-density.html` | dartsonline.com | feature 2 (PDC calendar) |

All four verified at the served artefact. Migrations: `491`–`495`, `498`, `499`
(features + substrate), `496` (legible-ink repoint), `497` (meta descriptions).

**dartsonline has NO hub yet** — one feature does not need one. Its second
feature is the trigger; the peer lane owns that site's nav (see §7).

## 3. The recipe for the next feature — proven three times

1. **Find a cluster** in the site's own feed: one story on **≥2 channels**.
   ```sql
   SELECT source_title, topics, relevance_score FROM content_feed_items
    WHERE site_id='<id>' AND status='relevant' AND created_at > now() - interval '10 days'
    ORDER BY created_at DESC LIMIT 25;
   ```
2. **State the premise, not the topic.** Design D's test: *if this were false,
   would the story change?* The premise must be falsifiable against **one**
   thing, and that thing becomes the chart. Keep the **dissenting** article — it
   has been the most valuable item in all three clusters.
3. **Verify every figure verbatim from a PRIMARY**, in-session. `pdftotext` on a
   PDF beats a vendor page restating it. Where a source gives a range, register
   the end that **favours the side you argue against**.
4. **Register the substrate** (a `_facts.sql` migration): a `series` fact needs
   `observations[]` where each observation carries **its own** `source.citation`,
   never inherited. Metrics need `value`, `tolerance`, `context_terms`,
   `writer_line`.
5. **File the hero image** through the framework, never a hand-upload — the key
   convention is `content_hero_<page_name_underscored>`
   (`imageryplan.ContentHeroKey`), item shape copied from
   `check_content_image_missing`, handler `image-build-handler`. **File at
   `triaged`, not `detected`**: `improvement-sweep` is DISABLED, and
   `detected-item-promoter` (900s) may be slow; `build-pipeline-trigger` (60s)
   claims `triaged`.
6. **Seed the page** (a `_feature.sql` migration): six sections, all existing
   components, rows locked `permanent` with `rendered_html` written in the same
   statement, `rebuild_policy='owned'`, slot names matching `pages.sections`
   entry-for-entry. Render locally first with the harness (§8).
7. **claimscan**, then **deploy after the hero asset is active**, then verify at
   the served page, then **render-audit with the stylesheet control first**.
8. **Add it to the hub** and stamp `published_page_id` on the cluster's items.

Worked examples to copy: `sql_for_agents/498` (best verify block) and `499`.

## 4. Open work, precise state

| item | state |
|---|---|
| **Fable / `features_open/035`** | **BLOCKED.** Component-hierarchy plan, owner wants **Fable specifically**, four capacity failures. **Do not substitute a model — ask.** Brief and reading list: design lane PLAN §2 + this lane's NOTES. |
| **Cobot feature (#4)** | **PARKED, ready.** Cluster real, premise good ("coverage reads as a takeover; ~1 in 10 installations says otherwise"). Needs ONE thing: a **primary** source for the 2024 cobot share. IFR's own page yields only 2023 (10.5% of 541,302). The 11.9%/64,542 figures are **not in any indexed IFR source** — negative result recorded so nobody re-walks it. |
| **`published_page_id` reader** | Data exists (15 rows, 2 pages) and **nothing reads it**. The `analysis_url` addition to `newsJSONItem` in `render_news_section_action.go` + the two news components' JS is designed, not built. Go change → council gate → inert until a roll. |
| **`bugs_open/349`** | Filed by this lane. 090 verdict **CONFIRMED**. 20 orphan page rows on 12 live sites that are "wanted live" and serve 404. **No root cause established** — do not read the verdict as one. |
| **`bugs_open/198`** | Not ours, but this lane contributed heavily: the corrected two-clause ink check, its first fleet run (5 stale tokens / 4 sites, verified), and the restore-procedure additions. **Not fixing those five** — see §9. |
| **`bugs_open/296`** | The 1.00:1 white-on-white shared CTA button. Contributed severity + cause distinctions. Owned by the brochure lane. |
| **Design uplift Phase A2** | `compute_component_quality` on the editorial components: **never run.** |

## 5. Rollout candidates

Sites with a live feed AND an evidence base: **dartsonline** (done),
**mortgagecalculator**, **fundamentallyai**, **ai-agent-orchestration**,
**relojistas**, **webdesign**. `gaswholesalers` has a feed but **no evidence
base**, so its charts would fail closed — that is a prerequisite, not a blocker
to note in passing.

## 6. Owner rulings in force

- **Lifecycle RATIFIED 2026-08-20:** indefinite retention at one stable URL;
  retirement is **deliberate de-listing, never deletion**; **refresh cadence is
  per FACT, not per page**; tracker/feature/explainer tiers.
- **Hero default:** image + semi-transparent overlay, ahead of gradient-only.
  Needs no new component — the live `hero` template's image branch already emits
  the overlay.
- **Everything goes through the framework** — no hand-built pages, no
  hand-uploaded assets.

## 7. Peer coordination (live, and it works)

`dartsonline_traffic` (owns dartsonline nav + sitemap) and the `320`/`339` lane
have each found real defects in this lane's work. **Tell them before creating a
dartsonline page** (rolling growth budget), and send page names for one nav
rebuild + sitemap regeneration per batch.

Between us we established: the `--absent` reconcile mode, the owned-page filter
narrowing (§9), the two-clause ink check, and the restore-procedure rules on 198.

## 8. Tools

- **Local render harness**: `scratchpad/build/render.go` — replicates
  `executeGoTemplate` (text/template, `missingkey=zero`, the six-function
  funcmap, `<no value>` stripped). Render every section locally *before* seeding.
- **Deploy**: assemble-only `page-rerender` — kcat `orchestrate` envelope with
  `spec:{page_name}` and **NO `spec.reason`**. Sending a reason runs the section
  rerender and can escalate the page to the content writer.
- **Audit**: `python3 scripts/render_audit.py <url>...` — **always check the
  served stylesheet size first** (§9).
- **De-listing**: RUNBOOK §10.

## 9. Traps — every one of these was paid for

1. **A render audit against a broken stylesheet produces false PASSES.**
   Unapplied CSS cannot fail contrast. `curl -sL <domain>/assets/css/styles.css |
   wc -c` **before** believing any audit; healthy is 25–27 KB on these sites.
   This invalidated a conclusion already committed.
2. **A status code is not an artefact.** A `200` was returned while the body was
   9 bytes reading `Not found`. Assert on **bytes plus a known string**.
3. **`COMPLETED` is not success.** A render audit returned
   `status=COMPLETED, current_step=complete_error` with a TIMEOUT inside. Read
   `__step_error`.
4. **Never set `in_footer=true` on a feature article** — it appends the article
   to every page's footer, growing linearly. Articles take **no** nav membership;
   the hub is the listing.
5. **`meta_description` is public copy.** A premise-test sentence from the design
   doc shipped as one. The `≤160` guard catches length only; **tone is caught
   only by reading it as a stranger.**
6. **A sample is not a census.** "~30 of 31" from five spot-checks was really 19;
   "42 rows" for bug 349 was really 20 once two never-built *sites* were split
   out.
7. **`-ink == --color-text` alone is NOT staleness** — it must also differ from
   its own source token, or you condemn a correct token on a palette that uses
   one ink for two roles.
8. **A restore is not a repair**: seeding a theme row reinstates whatever the blob
   encoded, including a pre-2026-08-14 derivation.
9. **The news render path already detects same-story clusters and DISCARDS them**
   (`queryresolve/news_items.go`). Reuse the heuristic, never that call site.
10. **`reconcile_footer_nav.sh` skips only `owned AND verbatim`** now (narrowed
    2026-08-21, 174 owned / 3 verbatim fleet-wide). Assemble mode works fine on
    ordinary owned pages — proven by six successful dispatches.

## 10. What NOT to do

- Do not substitute a model for Fable on `035`.
- Do not build the cobot feature on second-hand figures.
- Do not regenerate the four sites' palettes from `198`'s list — those are
  **suspects**, the confirming test is a fresh render, and the remedy changes
  every derived value on a site at once.
- Do not narrow `PageWantedLivePredicateFor` (bug 349) without architecture
  review — `LoadSitePagesAction` and `check_orphan_pages` both depend on it.
- Do not hand-patch nav rows; file `nav_drift` and let the derivation run.
