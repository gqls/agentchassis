# HANDOFF — leopardessconsulting.co.uk rebuild

**Purpose.** Start a fresh chat from exactly here. Read this top-to-bottom first; it is
the single source of truth for state. The deeper detail lives in the four companion
docs, but this file is enough to resume without them.

**Last updated:** 2026-07-16 (turn 17, extended)
**Branch:** `085_debug_and_feature_loops`
**Site:** `leopardessconsulting.co.uk` · `site_id = 4851f6fc-71cf-4160-a270-e03d6d3e0732`
**Plain-language status to show someone:** `SUMMARY_where_we_are.md` (this file is the technical resume point).

---

## ⚡ CURRENT PUNCH-LIST (owner site review, 2026-07-14) — start here

The owner reviewed the live site and raised the items below. Status as of **turn 17**.

| # | Issue | Status | Fix / root cause |
|---|---|---|---|
| 1 | Nav sometimes blue, sometimes black; footer navy | **FIXED & verified** | Header/footer baked into each page at assemble time; footer navy = collection `color_palette.primary=#1a1a2e`. Set primary→`#0D0D0D`; `rerender-pages` (`refresh_site_components:true`). NOTE: leftover `#1a1a2e`/`#0f3460` in page HTML are usually **dead CSS fallbacks** `var(--color-primary,#1a1a2e)` — BUT turn 17 found a *live* navy gradient hardcoded in **stale rendered_html** on use-cases (no `var()`, it really rendered). Distinguish: `var(...)` = dead; bare hex in a `background:` = real, fix by re-rendering the section from its current template. |
| 2 | Nav cluttered / blank "For Leaders" in nav | **FIXED (header + footer)** | Header nav reads `site_nav_items` **primary** group. Turn 17: the FOOTER reads `site_nav_items` **primary+utility+legal** (`render_site_components_action.go:98`) — `pages.in_footer` is DEPRECATED and does nothing. "For Leaders" survived in the footer of every page until its `utility` nav row was DELETED (`bak_navitem_fel_20260714`). |
| 3 | Card links 404 (how-we-work, who-we-help, use-cases) | **FIXED** | `info-card-grid` rendered `<a href>` ungated → gated `{{if .link_url}}`. Turn 17 found 2 MORE phantom `/tools/tool-ai-readiness-quiz.html` links (use-cases hero + CTA) that turn 15 missed; gone now. Backstop: enable `phantom_internal_links` + `broken_nav_links` discovery checks (still OFF). |
| 4 | About invented stats (30 Clients Served / 2,767 Awards Won) | **FIXED (made true, not removed)** | Labels were LLM-fallbacks. Set to honest+true (30 yrs / 8 sites / 2,767 records). |
| 5 | **Missing images site-wide** | **PARTIALLY DONE — 2 per-page heroes LIVE (turn 18)** | who-we-help + how-we-work carry on-brand Banana-generated heroes (`kind:'illustration'` — NOT `'hero'`, which routes to SDXL and produced garbage), wired via manual `site_plans`/`site_plan_imagery` rows (`scripts/wire_heroes.sql`) + `image_landed` rerenders. **Review every generated image by eye before wiring** (both of these passed; the SDXL one didn't). Remaining: index has its hand-chosen hero; about/services heroes need component work (their hero components have no image field); per-card/section images = Phase I3 (being built in the imagery workstream — content_card plumbing appearing in `url_helpers.go`). |
| 6 | blog.html broken (empty "min read/Read more", no posts/images) | **FIXED (excerpts + read times live)** | Root cause was empty `pages.meta_description` on the 2 `/blog/` posts. Populated → `rerender-pages` → `rebuild_blog_listing`. Live: 5 cards, real excerpts, real read times. Card **thumbnails** remain blank (Phase I3 gap, `rebuild_blog_listing_action.go:186` hardcodes `image:""`). |
| 7 | use-cases claims we do things we don't | **FIXED & verified live** | Was worse than reported: 5 fabricated case studies with invented clients + invented results. The rewritten `portfolio` spec already held 5 honest "Not yet done for a client" use_cases — made the spec the source of truth (mirrored `status`→`client`), backfilled content_data on all 3 slots, re-rendered. Live: 0 fabrications, 0 phantom links, 0 navy. |
| 8 | favicon.png 404 | **FIXED** | Head hardcodes `/assets/images/favicon.png`; only `.ico` was committed. Committed the png → 200. |
| 9 | Empty pages | **✅ FULLY CLOSED (turn 18)** | **ai-readiness-quiz LIVE** — take 6 built in a healthy cluster window (54KB interactive quiz, 3 components, integrity-checked: 0 invented emails, 0 banned claims, links real). llm-cost-calculator-guide rebuilt turn 17; for-engineering-leaders archived turn 17 (+ our-approach & for-engineering-teams archived turn 18 by merge decision A11). 7 empty case-study pages remain `archived`, unlinked → harmless. |
| 10 | **Voice still reads LLM-written** | **DONE for the pages that needed it — verified live** | `specs/VOICE_REWRITE_PROMPT.md`. **4 pages rewritten & verified live:** services (hero+CTA — killed the banned triad "observability, fault isolation, cost controls" that appeared twice; fixed a circular CTA self-link), how-it-works (duplicated text block → honest "What it does not do" limits section), our-approach (hero triad + title-case heading), contact (CTO copy + empty-cd landmine + the 4th phantom quiz link). **Sitewide check: 0 banned-triad occurrences, 0 phantom quiz links.** **Method:** the pipeline path (page-content-writer + `rewrite_guidance`) is the SAME spawn→child path broken by §8, so hand-edit `content_data` + fire a no-LLM `section_data_resolved` rerender (guard: a slot with EMPTY content_data escalates the whole page to the writer — populate first, as done for contact). **Already fine, don't redo:** how-it-works body, engagement-model, who-we-help, for-engineering-teams, services middle sections. **Remaining:** technical-architecture (low priority) + the page-MERGE decision (§6.3). |

Backups (turn 17): `bak_pages_contentdata_leo_20260714`, `bak_usecases_leo_20260714`,
`bak_portfolio_spec_leo_20260714`, `bak_blogmeta_leo_20260714`, `bak_fel_page_leo_20260714`,
`bak_navitem_fel_20260714`, `bak_contactblock_20260714`; voice pass:
`bak_services_leo_20260715`, `bak_howitworks_leo_20260715`, `bak_ourapproach_leo_20260715`,
`bak_contact_leo_20260715`.

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
- **A11** (2026-07-16) Near-duplicate pages MERGED: keep `how-it-works` + `technical-architecture`
  ("Architecture"); `our-approach` + `for-engineering-leaders` + `for-engineering-teams` are
  archived. Don't resurrect them; new content goes on the survivors.

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

**THE ONE OPEN IN-FLIGHT ITEM: `ai-readiness-quiz` is still blank.** Its content path is clean
(validation passes; the `contact-block` validator bug is fixed fleet-wide) — it is blocked ONLY
by the infra flake in §8 (5 build attempts, all died on lost-child-response / reaper, never on
content). Its `sections` are `["hero","ai-readiness-quiz","contact-block"]`; the interactive
quiz component exists and is active. **To resume: re-fire page-build-handler for it when the
cluster is healthy, and watch the spawned child land its response** (see
`docs/HANDOFF_spawn_lost_child_response.md` — that's the separate thread for the actual bug).
Check state: `SELECT count(*) FROM page_components pc JOIN pages p ON p.id=pc.page_id WHERE p.name='ai-readiness-quiz' AND p.site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732';`
It's linked from the footer + use-cases CTA, so it must not stay blank.

1. **[✅ done] Site uniformity / header + footer.** KEY LESSON: use `reassemble_pages.sh`
   (ASSEMBLE mode + page_id) for header/footer changes — `section_data_resolved` SKIPS
   unchanged pages and silently won't update them. For a NAV change you must also re-render
   the slot first: `rerender-pages` with `spec.refresh_site_components:true`. And when the
   `page_rerender` queue won't drain (build-dispatch-loop stalls — happens), `reassemble_pages.sh`
   drives `page-rerender` directly, bypassing `site_work_items` entirely.
2. **[✅ mostly done] VOICE REWRITE (punch #10).** 4 pages done & verified live this session:
   services, how-it-works, our-approach, contact (all via hand-edit content_data + no-LLM
   `section_data_resolved` rerender — the reliable path; the pipeline route is infra-blocked).
   **Remaining:** technical-architecture (defensibly technical, low priority); the rest of the
   pages are already in good voice — DON'T redo how-it-works body, engagement-model,
   who-we-help, for-engineering-teams, or the services middle sections.
3. **[✅ done — owner decision A11, executed turn 18] PAGE MERGE.** Kept `how-it-works`
   (canonical, body-linked from all 15 pages) + `technical-architecture` (the "one click down"
   technical page, nav label "Architecture", hero now "The architecture, in detail").
   ARCHIVED `our-approach` + `for-engineering-teams` (same six claims, zero inbound body links);
   their utility nav rows deleted. The two unique for-engineering-teams items were folded into
   technical-architecture's features **rewritten to verified facts** — the original "Depth
   Across Business Functions" contained the audited-out "eight departments" fabrication; it's
   now "over 150 agent definitions" (DB-verified 156). TA's duplicate differentiators section
   deleted. All verified live. Do NOT un-archive or re-polish the merged pages.
4. **Imagery (punch #5).** Blocked on the infra flake + the hero→SDXL routing gap, both in
   §8/§9. `scripts/wire_heroes.sql` is written and guarded — only run it against an asset you
   have *looked at* (the one SDXL hero generated was unusable; nothing is wired).
5. **Per-page `<title>` + og:image meta** (partly stale marketing titles) and **there is no
   `sitemap.xml` at all** (0 URLs) — SEO/social polish.
6. **The build-out the brief actually asks for**: tools, illustrated guides, news surface,
   "AI working frontend" demos (label simulations honestly), infographics. Tool library
   already has reusable interactive components — deploy/adopt rather than rebuild. News feed
   pipeline is real and running — surface it.
7. **L7 — the chart component** (A5/A7). The one genuinely-new build. Start from real DB
   numbers (2,767 / 4,672 / 8 / 75,061), Go emits static SVG, JS enhances. See PLAN §5.

**Done, don't redo:** theme-color is already `#0D0D0D` live. `for-engineering-leaders` is
archived (do NOT "fix" it — deliberate merge-duplicate removal). All 4 phantom
`/tools/tool-ai-readiness-quiz.html` links are gone (use-cases ×3 + contact ×1). The banned
triad "observability, fault isolation, cost controls" is gone sitewide.

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
  liveness bail-out when the request consumer can't dial for N minutes. **Full write-up for a
  separate thread: `docs/HANDOFF_spawn_lost_child_response.md`.**
- **No fact-verification layer exists anywhere in the platform** (turn 18, owner-confirmed
  gap). Prompt rules say "never invent" (leaky), `validate_page_content` checks form not truth
  (emails are the sole fact-shaped check — and the only fabrication class ever caught),
  discovery checks are structural, `content-quality-auditor` is tone-only, and the identity
  spec's `evidence_base` key has zero code consumers. The audit is manual
  (AUDIT_verified_facts.md). **Owner decision: build it — spec for a NEW thread at
  `docs/agent_docs/docs024_key_docs_latest/claims_verification/SPEC_claims_verification.md`**
  (leopardess is the pilot site; its audit doc transcribes into the evidence base, and its
  shipped fabrications are the benchmark corpus).

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
page_component on that page must have non-empty **object** `content_data`, or the WHOLE page
is escalated to the content writer (`content_data_backfill`) and rewritten.
**⚠️ The escalation has TWO branches (learned turn 18 the hard way — clobbered who-we-help
once, nearly twice):**
(a) `len(contentData)==0` after `json.Unmarshal` into a map — so a valid JSON **array** or
scalar ALSO escalates (unmarshal fails silently → nil map), not just NULL/`{}`;
(b) content_data present but **missing ANY schema-required `source:"llm"` field**
(`missingRequiredLLMFields` — e.g. info-card-grid requires `section_eyebrow` +
`section_subtitle` on top of the obvious `cards`/`section_title`).
Pre-check BOTH before any section rerender:
`SELECT slot_name FROM page_components pc JOIN pages p ON p.id=pc.page_id WHERE p.name='<page>' AND p.site_id='…' AND (pc.content_data IS NULL OR jsonb_typeof(pc.content_data::jsonb) <> 'object' OR pc.content_data::jsonb = '{}'::jsonb);`
plus compare `jsonb_object_keys(content_data)` against each slot's
`input_schema->'fields'` where `source='llm' AND required=true`.
If an escalation DOES fire: it emits a `content_data_backfill` needs_page item that survives
orchestration death (resets to 'triaged' when the reaper kills its claimer) — close it
manually (`status='complete'`) or it re-clobbers on a later dispatch. Also know what a writer
rebuild does: it preserved the honest voice (pinned specs held) but fabricated four case-study
titles and a phantom link on who-we-help — review everything it writes before deploying.

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
