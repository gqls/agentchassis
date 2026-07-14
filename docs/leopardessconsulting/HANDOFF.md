# HANDOFF — leopardessconsulting.co.uk rebuild

**Purpose.** Start a fresh chat from exactly here. Read this top-to-bottom first; it is
the single source of truth for state. The deeper detail lives in the four companion
docs, but this file is enough to resume without them.

**Last updated:** 2026-07-14 (turn 15)
**Branch:** `083_imagery`
**Site:** `leopardessconsulting.co.uk` · `site_id = 4851f6fc-71cf-4160-a270-e03d6d3e0732`

---

## ⚡ CURRENT PUNCH-LIST (owner site review, 2026-07-14) — start here

The owner reviewed the live site and raised the items below. Root causes and prior-art
fixes are from two research passes (see RUNNING_NOTES turn 15). Status as of turn 15:

| # | Issue | Status | Fix / root cause |
|---|---|---|---|
| 1 | Nav sometimes blue, sometimes black; footer navy | **FIXED & verified** | Header/footer baked into each page at assemble time; header slot was empty; footer navy = collection `color_palette.primary=#1a1a2e`. Set primary→`#0D0D0D`; triggered `rerender-pages` (`refresh_site_components:true, force_rerender:true`) — re-rendered all 3 slots + re-assembled every page. 27/30 gold header (3 empty pages excepted). NOTE: leftover `#1a1a2e`/`#0f3460` in page HTML are **dead CSS fallbacks** `var(--color-primary,#1a1a2e)` — the variable is defined, so they never render. Don't chase them. |
| 2 | Nav cluttered / blank "For Leaders" in nav | **FIXED** | Header nav reads `site_nav_items` (primary group), NOT `pages.in_header`. Rebuilt to 9 items. |
| 3 | Card links 404 (how-we-work, who-we-help, use-cases) | **FIXED** | `info-card-grid` rendered `<a href>` ungated. Gated template `{{if .link_url}}`; stripped 6 phantom links; repointed use-cases quiz link. Backstop: enable `phantom_internal_links` + `broken_nav_links` discovery checks (currently OFF). |
| 4 | About invented stats (30 Clients Served / 2,767 Awards Won) | **FIXED (made true, not removed)** | Labels were LLM-fallbacks. Set to honest+true (30 yrs / 8 sites / 2,767 records). Clean *removal* needs gating the shared `content-block-about` template. |
| 5 | **Missing images site-wide** | **OPEN — biggest item; NOT blocked** | **A6 Banana routing IS deployed** (running `image-generator-adapter v1.0.1114`; the turn-8 "not deployed" note was stale). Root cause: **leopardess has NO `site_plan` and 0 `site_plan_imagery` rows.** Two routes — see §9 below. Per-card/section images need **Phase I3 (content-imagery lane), which is NOT built** — that's a structural limit, not a config miss. |
| 6 | blog.html broken (empty "min read/Read more", no posts/images) | **OPEN (less broken than it looks)** | Listing already renders 5 posts with working links. Real defects: card `image=""` (Phase I3 gap, `rebuild_blog_listing_action.go:186` hardcodes `""`) + empty excerpts (blog posts have empty `pages.meta_description`) + blank read_time. Fix excerpts: populate `meta_description`, re-run `rebuild_blog_listing`, reassemble `/blog.html`. Copy robot-hands' blog. |
| 7 | use-cases claims we do things we don't (LinkedIn enrichment, doc-watching agents, Slack/PagerDuty) | **OPEN** | Reframe as "could do", per the AUDIT rule. Not yet edited. Content in the `use-cases-list` rendered_html (content_data NULL — LLM-authored into rendered_html). |
| 8 | favicon.png 404 at `/assets/images/favicon.png` | **FIXED** | The head (`render_site_components_action.go:399,403`) hardcodes `/assets/images/favicon.png`; we'd only committed `.ico`. Committed `brand/favicon.png` (=favicon-180) to that path via git-adapter → now 200. (`derive_brand_head_assets` can't self-serve it: it needs the logo `url` to be an S3 handle, but ours is a git path.) |
| 9 | 3 zero-section pages (ai-readiness-quiz, for-engineering-leaders, llm-cost-calculator-guide) | **OPEN** | Blank pages; need content rebuild or removal. |
| 10 | **Voice still reads LLM-written** | **PROMPT WRITTEN** | Owner: "think hard about a prompt." → `specs/VOICE_REWRITE_PROMPT.md`. Apply it page-by-page (substance rewrite, not repolish). |

Backups made this turn: `bak_stylecoll_leo_20260713`, `bak_site_nav_items_leo_20260713`,
`bak_infocardgrid_20260713`, `bak_pages_leopardess_20260713`.

---

## 0. The one-paragraph version

We are rebuilding leopardessconsulting.co.uk (a real, live site the agentchassis platform
built for itself) to be honest, well-branded, and genuinely useful. The old site was full
of fabrications (invented staff, "8 departments", fake client case studies, capabilities
that don't exist). The engineering it describes is largely real; the framing was not.
We have: audited every claim against code/DB (`AUDIT_verified_facts.md`), rewritten the
specs and page content for a sceptical non-specialist business buyer, deployed a real
leopardess logo + favicon + OG card, forked a WCAG-validated light-reading/dark-chrome
palette, removed the fabrications, and got the main pages live. What remains is finishing
uniformity across secondary pages, an SEO/meta polish, and the bigger build-out (tools,
guides, news, charts, illustrated content).

**Governing rule:** no claim ships without a row in `AUDIT_verified_facts.md`. Verify by
artifact (curl/DB/screenshot), never by a "complete" status.

---

## 1. Companion docs (read as needed)

| Doc | What it is |
|---|---|
| `AUDIT_verified_facts.md` | The evidence base. Every C/U/D/P/X finding, what's true, what was removed. **Read before writing any claim.** |
| `PLAN_leopardess_rebuild.md` | Phases L0–L9, owner decisions A1–A10, the chart plan (L7). |
| `RUNNING_NOTES.md` | Turn-by-turn log. Newest turn at the bottom. **Update every turn.** |
| `RUNBOOK.md` | Human tasks (H1–H10) and operator procedures (O1–O7) + landmines. |
| `REPLICATION_in_chassis.md` | How everything done here maps to in-chassis operations (owner requirement: replicable without external tools). |
| `scripts/` | Reusable, documented SQL + kcat scripts (see §5). |
| `brand/` | Logo masters, favicons, OG card, hero image. |
| `specs/` | The rewritten site_specs JSON (identity, voice, design_intent, portfolio) + backups. |

---

## 2. Owner decisions locked (don't re-litigate)

- **A1** No invented staff. About is first-person, **unnamed** ("Founder and engineer").
  Real background: 30 yrs, worldsoccernews.com (~12M unique users/mo peak, reported in
  **New Media Age**; "third busiest sports site" allowed as hedged recollection). Bumble
  left out.
- **A2** Audience = sceptical, commercially-sharp, non-specialist business buyer. NOT
  CTO-only. Technical depth is available one click down, not the register of every page.
- **A3** Logo = stylised leopardess head, profile. **Done** (owner chose candidate c2).
- **A4** Dark chrome, light reading surfaces. **Done** (palette forked + deployed).
- **A5/A7** Chart component = Go emits static SVG (data-first, simple/extractable), JS
  progressively enhances. Honour imagery decisions D1 (code renders data, LLM never
  touches values) and D3 (chart is a Lane-B asset, not a site_plan_imagery kind). **Not
  started** — this is L7.
- **A6** Route logo/illustration/infographic → Banana (honours reference images).
  **DEPLOYED** — running `image-generator-adapter v1.0.1114` (commit `49d67e82`, v1.0.1103,
  2026-07-10). Verified: robot-hands icon/logo/sprite assets show `origin_model=banana/…`.
  The earlier "not deployed" note (turn 8) was stale.
- **A8** Data-sovereignty pitch = a capability built *with* a client (pilot-scoped),
  never a standing isolation/residency guarantee.
- **A10** Two-tone gold: bright `#C8A951` on dark chrome only; bronze `#836E32` for links
  on light (bright gold fails WCAG AA on light — 2.1:1).

**Palette (forked, leopardess-owned, WCAG-validated):** bg `#FAF8F4`, text `#1A1A1A`,
accent/link `#836E32`, primary `#0D0D0D`, header/footer/hero bg `#0D0D0D`, card `#FFFFFF`,
cta bg `#C8A951` + cta text `#0D0D0D`, footer text `#B9B3A6`, hover `#A8843C`.

---

## 3. What is DONE and live (verified by artifact)

- **Logo/favicon/apple-touch/OG card** — live at `/assets/images/logo.png`, `/favicon.ico`,
  `/apple-touch-icon.png`, `/assets/images/og-card.png`. Byte-verified.
- **Palette** — forked (`palettes`/`css_themes`/`style_collections` all leopardess-owned;
  seed `3196d966` untouched, still dresses 3 other sites). Deployed `styles.css` matches
  the validated palette exactly. Light reading surface, dark chrome, gold CTAs.
- **Header** — forked (`header-leopardess`): charcoal bg, **gold** "Get Started" button
  with dark text, gold nav hover, logo shown. Wired through the forked collection + the
  `site_components` header slot. Verified in a screenshot.
- **Footer** — already correct (uses `--color-footer-bg`/`--color-footer-text`; no fork
  needed).
- **Specs** — identity/voice/design_intent/portfolio rewritten (source_agent
  `operator-rebuild`, pinned). All fabrications removed.
- **Content, main pages** — index, about, services, who-we-help, case-studies,
  how-we-work + main nav re-rendered. Homepage: hero → "What we have built" (3) →
  "What we might build with you" (3) → CTA. Stats panel removed (see §4). Hero AI image
  replaced with a subtle on-brand one. **Fabrication sweep across all content_data =
  CLEAN.**
- **`sites.tagline`, `logo_url`** — updated.

---

## 4. Sharp edges discovered (so you don't re-learn them)

1. **`needs_rebuild` is inert** — nothing scans `pages`; you must insert a `site_work_items`
   row. (RUNBOOK O3.)
2. **Rerender pipeline is a minefield.** `rerender-site`'s sequential page loop STALLS on
   a lost child response (only ~1/30 pages rendered before it hung). Drive pages directly
   with `page-rerender` instead (scripts in §5).
3. **`page-rerender` only regenerates section HTML when `spec.reason='section_data_resolved'`
   AND `spec.page_name` (the page `name`, not url) is set.** Without the reason it just
   re-assembles stored HTML (fine for embedding a new header/footer). `rerender_page_sections`
   hard-requires `spec.page_name`.
4. **Section resolvers OVERRIDE content_data on every render.** The hero `background_image`
   is auto-resolved to `/assets/images/hero.jpg` (`plan_sections_action.go:1338`) — you
   can't remove it via content_data, you change the FILE. `source:"static"` schema fields
   (e.g. system-stats suffixes `%/ms/+/x`) are re-applied from the schema `fallback` every
   render and can't be overridden per-instance; the component is shared and re-links on
   rerender, so forking a section component does NOT stick.
5. **Shared components** (grids, header, system-stats) back many sites — never edit their
   templates/schema. Fork + repoint, but note (4): a *section* fork gets re-linked by
   `save_page_sections`; a *site_components* fork (header) sticks because it's wired via
   the collection.
6. **The `analyze_design` LLM invents a palette** unless design_intent uses
   `palette.reference_values` with prescriptive guidance. It reads that exact path, not
   `color_scheme`. Ours is now structured correctly.
7. **Image assets: `assets.url` is a throwaway presigned URL; pages serve the durable git
   path** via `DeployedWebPath`. Don't chase presigned URLs — see the "HOW IMAGE SERVING
   ACTUALLY WORKS" box in `PLAN_imagery_best_in_class.md`. The pipeline works.
8. **Back up before ANY change**, including component forks (I once forked without a backup
   and briefly feared corrupting a shared row — I hadn't, but the rule stands). Backups
   use the repo's `bak_*` table convention.
9. `kubectl exec ... <<HEREDOC` needs `-i` or it silently runs nothing. Prefer
   `kubectl cp` a .sql file then `psql -f`.

---

## 5. Operator scripts (reusable, in `scripts/`)

| Script | Use |
|---|---|
| `rerender_pages.sh <site> <domain> <page_name>…` | Regenerate section HTML from content_data + deploy (sets `spec.reason=section_data_resolved`). Use after editing a page's content. |
| `reassemble_pages.sh <site> <domain> <page_name>…` | Re-assemble page (embed current header/footer + deploy) WITHOUT touching sections. Use after a header/footer/palette change. |
| `commit_brand_assets.sh <domain> <brand_dir>` | Commit specific pre-approved images (logo/favicon/OG) straight to the git-adapter. |
| `deploy_brand_asset.sh …` | Single-asset variant. |
| `L3_fork_palette.sql` | The palette fork (reference). |
| `L5_homepage.sql`, `L5_pages.sql`, `L5_faq_hero.sql`, `L5_casestudies.sql` | The content rewrites (reference). |

**Trigger any agent** by producing to Kafka `system.agent.generic.requests` — copy the
`kcat` block from `082_submit_domain_unified.sh`; only `agent_type` + `input_data` change.
**DB:** `kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c "<SQL>"`.
**Verify a deploy:** `curl` the page (GitHub Actions → B2 → Cloudflare Worker, ~30–90s),
never trust orchestration status alone.

---

## 6. NEXT ACTIONS (in priority order)

1. **[✅ done — 27/30] Site uniformity.** All content pages carry the forked gold header/
   logo; the palette applies site-wide via the stylesheet. The 3 not done
   (`ai-readiness-quiz`, `for-engineering-leaders`, `guides/llm-cost-calculator-guide`)
   have **zero sections** — they need a content REBUILD (see #5/L8), not a re-assembly.
   KEY LESSON: use `reassemble_pages.sh` (ASSEMBLE mode + page_id) for header/footer
   changes — `section_data_resolved` SKIPS unchanged pages and silently won't update them.
2. **Per-page `<title>` + og:image meta.** Still partly stale (old marketing titles).
   Comes from page metadata / the head slot mechanism, not the section content. SEO/social
   polish.
3. **Secondary-page copy (L5 cont.).** engagement-model, for-engineering-teams,
   for-engineering-leaders, how-it-works, our-approach, technical-architecture still carry
   CTO-register copy; rewrite for A2. Merge near-duplicate pages rather than restyle each.
4. **Head `theme-color` meta** is still `#1a1a2e` — set to `#0D0D0D`.
5. **The build-out the brief actually asks for** (beyond fixing what's broken): tools,
   illustrated guides, news surface, games, "AI working frontend" demos (label simulations
   honestly), infographics. Tool library already has reusable interactive components
   (`tool-ai-agent-roi-estimator`, quizzes, calculators) — deploy/adopt rather than
   rebuild. News feed pipeline is real and running — surface it.
6. **L7 — the chart component** (A5/A7). The one genuinely-new build. Start from real DB
   numbers (2,767 / 4,672 / 8 / 75,061), Go emits static SVG, JS enhances. See PLAN §5.

## 7. Open owner questions (RUNBOOK H-items)

- **H7** worldsoccernews.com "third busiest" — no citable source; published as hedged
  recollection + the New Media Age-reported 12M figure. Owner may still supply a source.
- Nothing else blocking. Contact details confirmed real (H3). Engagement shape = pilot-first
  ladder (H6).

## 8. Platform issues surfaced (not this site's to fix, but logged)

- No deterministic contrast gate on palette specialised slots (WCAG primitives exist in
  `color_util.go` but aren't called there). We validated by hand.
- A suffix-free "big numbers" stats component would be a good addition (the shared
  `system-stats` forces `%/ms/+/x`).
- `system-stats` and other section components can't be per-site customised (rerender
  re-links to the canonical component).
- **`contact-block` can never pass its own validation** (turn 17): its schema has
  `email_placeholder` `source:"static"` `fallback:"jane@company.com"`; static fallbacks
  re-apply on every render, and `validate_page_content`'s email check rejects any email
  that isn't the site contact → every page-build containing contact-block fails
  (`0 blockers, 1 errors`). This killed the quiz rebuild twice and likely yesterday's 4
  failed `content_data_backfill` items. Fleet-wide fix = change the shared fallback to `""`
  (+ gate in template) or teach the validator that placeholder-attribute emails aren't
  contact claims. Per-page workaround: drop contact-block from `pages.sections`.
- **spawn→call stall, node-local Kafka dial timeouts** (turn 17): two spawned
  image-generator pods initialised, sent their init response, then their consumers looped
  on `dial tcp 10.20.99.93:9092: i/o timeout` (broker healthy, other nodes fine) — the
  orchestrations sat AWAITING_RESPONSES at the spawn step forever (the workflow's
  `timeout_seconds: 300` did not fire). Remedy: delete the idle pods, re-fire; consider a
  liveness bail-out when the request consumer can't dial for N minutes.

## 9. IMAGERY playbook (from research 2026-07-14) — how to give leopardess images

**A6 Banana routing is DEPLOYED** (`image-generator-adapter v1.0.1114`; commit `49d67e82`,
v1.0.1103). logo/illustration/infographic/icon/sprite_sheet → Banana (honours reference
images); hero → SDXL. No deploy needed. **Root problem: leopardess has NO `site_plan` and
0 `site_plan_imagery` rows** — nothing has ever emitted imagery work items for it.

**⚠️ Route A as originally written is NOT content-safe on this site (turn 17 code trace).**
With `"scope":"page"`, image-build-handler's terminal `flag_page_image_rebuild` step emits
`needs_page` → page-build-handler, which has **no skip-content branch** — it ALWAYS calls
page-content-writer when sections are ready, and every `source:"llm"` field is regenerated
(`plan_sections_action.go:1124`). On robot-hands that was fine (LLM copy anyway); here it
would clobber the hand-fixed copy.

**Route A-safe (used turn 17): generate scope-less, wire manually, rerender no-LLM.**
1. Fire the inline spec **without `scope`** (flag step no-ops; asset still generated,
   stored, git-deployed). Omit `brand_update` unless you want `sites.content_data.hero_url`
   (a SITE-WIDE fallback) overwritten:
```json
{"action":"orchestrate","config":{"agent_type":"image-build-handler"},
 "input_data":{"site_id":"4851f6fc-71cf-4160-a270-e03d6d3e0732",
   "domain":"leopardessconsulting.co.uk","item_type":"needs_imagery",
   "spec":{"key":"hero_<page>","kind":"hero","prompt":"<prompt>","purpose":"hero",
           "asset_key":"hero_<page>","scope_ref":"<page>"}}}
```
2. AFTER the asset row is `active` (never before — design-discovery-agent runs
   `unfulfilled_imagery_plan`; a fulfilled plan emits nothing), insert `site_plans`
   (is_current) + `site_plan_imagery` rows (`scope='page'`, `scope_ref=<page name>`,
   `kind='hero'`, `key=<asset_key>`).
3. Fire `page-rerender` with `spec.reason='image_landed'`, `spec.page_name`, top-level
   `page_id` — `rerender_page_sections` is no-LLM by design; the hero resolves from the
   plan rows.
**PRECONDITION for step 3 (`rerender_page_sections_action.go:169`):** every
page_component on that page must have non-empty `content_data`, or the WHOLE page is
escalated to the content writer (`content_data_backfill`) and rewritten. Only `contact`
still has empty content_data (use-cases was backfilled turn 17). Check first:
`SELECT slot_name FROM page_components pc JOIN pages p ON p.id=pc.page_id WHERE p.name='<page>' AND p.site_id='…' AND (pc.content_data IS NULL OR pc.content_data::text IN ('{}','null'));`

Also: only pages whose hero component declares an image field can show a hero at all —
`index`, `who-we-help`, `how-we-work` use `hero` (`background_image` ✓); `about`/`services`
use `hero-about`/`hero-services` (NO image field — needs component work first).
image-build-handler generates → asset-deployer commits to `/assets/images/<asset_key>.<ext>`.
**Verify `assets.url` is `/assets/images/…`, NOT a presigned `s3…?X-Amz-…` URL.**
For brand consistency, pass the logo as a reference image (Banana kinds only).

**Route B — systematic (build-site-planner re-plan).** Writes the plan + imagery rows +
auto-emits `needs_imagery`. BUT ⚠️ **a full re-plan may re-run content generation and could
overwrite the carefully-fixed copy** — do NOT fire `build-site-planner` blind on this site.
If used, scope it and verify content isn't clobbered. Route A is the safe default.

**Per-card / per-section images = Phase I3 (content-imagery lane) — NOT built.** Cards get
the page-hero fallback (`imageRoleAliases`) or empty. A section image is wired via a
`site_plan_imagery` row `scope='section'`, `scope_ref='<page>:<ordering>'` joined to an asset
by kind (`plan_sections_action.go:231-260`). Building I3 is real new work.

**Recommended order for a session on imagery:** (1) hero images per page via Route A
(safe, immediate, big visual win); (2) the blog card images + the info-card-grid card images
need I3 — defer or build I3; (3) fix `design_intent.avoid` first if it still bans leopard
imagery (RUNBOOK O5 landmine — value unconfirmed, check it).

**Blog quick win (no I3 needed):** populate `pages.meta_description` on the blog-post pages,
re-run `rebuild_blog_listing` for the site, reassemble `/blog.html` → excerpts fill in.
Card thumbnails still need I3.

Key files: `internal/adapters/imagegenerator/dynamic_adapter.go`,
`platform/orchestration/actions/{write_site_plan_action.go, emit_imagery_items_action.go,
plan_sections_action.go, image build/deploy actions}`, `imageryplan/imageryplan.go`,
`discovery_checks/{check_unfulfilled_imagery_plan.go, check_empty_blog.go}`.
