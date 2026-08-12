# HANDOFF — leopardessconsulting.co.uk rebuild

**Purpose.** Start a fresh chat from exactly here. Read this top-to-bottom first; it is
the single source of truth for state. The deeper detail lives in the four companion
docs, but this file is enough to resume without them.

**Last updated:** 2026-08-12 — **the frontier is now §11, not §10.** If you already know
this file, read §11 first: it covers 07-30 → 08-12, and it is the section that changes
what you should do next. §10 is still accurate for 07-29/30 but is no longer the front.
**Branch:** `087_towards_multiple_domains`
**Site:** `leopardessconsulting.co.uk` · `site_id = 4851f6fc-71cf-4160-a270-e03d6d3e0732`
**Plain-language status to show someone:** `SUMMARY_where_we_are.md` (this file is the technical resume point).

---

## 🔴 READ FIRST (2026-08-12) — this site is now inside the fleet's automated loops, and they have already changed it

**The premise most of this file was written under has changed.** Between 2026-07-30 and
2026-08-12 this site was worked on **twice by automated fleet machinery, not by this
lane** — once deliberately (the owner picked leopardess for the D3 programme's first
supervised improvement-loop run on 08-08) and once at much larger scale on 08-11/12
(≈190 orchestrations: 97 `page-rerender`, 22 `asset-deployer`, 10 `internal-linker`,
**5 `page-content-writer` + 5 `page-build-handler`**, 2 `tool-improver`, 2
`section-editor`, 1 `webdesign-agent`). Details and figures: §11.

Three consequences, all measured today, all in §11:

1. **Hand-authored `content_data` is again at risk here — for a new reason.** Not
   `bugs_open/001` (that is genuinely closed). The live mechanism is
   `bugs_open/238`: a regeneration **replaces** `content_data` wholesale and drops
   keys the generator never emits, and gated templates then hide the loss in silence.
   **It has already fired on `/services.html`** — see §11.3.
2. **Two of this lane's 2026-07-31 deliverables are no longer live** (the second
   carousel; the six service images). Nobody reported it, because nothing on this
   platform reports it: the template renders less, without erroring.
3. **A verified-false claim came back onto a live page** after this lane removed it.

**So the working rule for this site has changed.** "I hand-authored it and verified it
live" is no longer a durable state — it is a state with a **date**. Before you build on
anything in §10 or earlier, re-check it at the artefact. The `[MEASURED 2026-08-12]`
table in §11.4 tells you what was still true today so you don't re-check it twice.

---

## 🔴 READ FIRST — the blocker that made content work provisional

> **CORRECTED 2026-07-30: this whole box is HISTORICAL. `bugs_open/001` closed
> 2026-07-20** — fixed in `c41e9ddbc` + migration 173, live on chassis `v1.0.1138`/
> planner `v1.0.1139`, verified twice on dartsonline (proven as a genuine rescue: a
> deployed page kept its real section against an LLM re-plan proposing to swap it, and
> refused an injected section on a sibling page). A re-plan can no longer silently
> re-compose or drop a `build_status='deployed'` page. **The "undefined shelf life"
> warning below no longer applies** — hand-fixed `page_components` content is durable
> again, the same as everywhere else on the platform. Left open by 001, but narrower and
> not currently active on leopardess specifically: `bugs_open/037/038/039/040/050/051`
> (see the closed file's own table for exactly what each one covers). This correction
> matters because two full days of imagery + a 4-article blog series were hand-authored
> directly into `page_components` on 2026-07-29/30 (§10) on the understanding that this
> platform no longer clobbers deployed pages — if you find evidence that's wrong, that is
> now itself a NEW bug, not a recurrence of 001.

**Original 2026-07-18 text, kept for the record:** Another session's `build-site-planner`
re-plan keeps clobbering this site's pages. Filed with fresh evidence at `bugs_open/001`.
It hit twice in 24h:
- **index, 07-17 14:14** — rebuilt 4→6 sections, re-adding *fabricated* "Functional Areas:
  150+" and invented case-study titles. Restored by hand; plans pruned.
- **services, 07-18 07:50** — rebuilt; invented a link to
  `/tools/tool-monitoring-coverage-gap-finder.html` (**404**) that the owner clicked. **Services
  is STILL in its clobbered state** (`page_components.updated_at = 2026-07-18 07:50`, verified)
  — my earlier v2 copy on that page is gone.

It is not only content *loss*, it is fabrication *injection*, and it defeats human review.
~~Anything you hand-fix in `page_components` has an undefined shelf life until 001 is fixed.~~

**What survives a clobber:** heroes wired through `site_plans` / `site_plan_imagery` rows.
That is why imagery work is durable here and copy work is not — prefer imagery until 001 lands.
~~(superseded — 001 is closed; both are durable now)~~

### 🔧 2026-07-19 — a second, quieter clobber path on this site, now closed (not by your thread)

A bugfix thread working `bugs_open/002` error C changed **this site's
`site_specs.site_plan` aspect** — telling you so it isn't a surprise. Migration
`176_leopardess_aspect_generic_text_block_fix.sql`, applied and ledger-recorded.

**Why it mattered here specifically.** This site has a current `site_plans` row with
**zero `site_plan_sections` rows**, so the section-list authority is the **aspect**, not
the table. 16 pages carry a deployed `generic-text-block` (added 2026-07-18 at the
`page_components` level, never written back up). On 14 that is harmless — the page is
absent from the aspect, or its aspect entry is `"sections": null`, so the aspect misses
and `pages.sections` governs. On **index** and **case-studies** the aspect held a real
array *omitting* the block, so the next rebuild would have **silently deleted live
copy**: "The whole thing on one page / Every figure here comes from our own database…"
and "The three systems, as a route map…". Both verified live before the fix. The
migration aligns the aspect **up** to what is deployed — it makes no editorial choice.
**If you want those blocks gone, remove them from the aspect AND `pages.sections`
together**, or the mismatch simply comes back.

Two things checked while there, both reassuring: the `who-we-help` drift that was filed
on 07-17 had already self-resolved in the 07-18 rebuild, and the `case-studies-grid` it
restored came back carrying the **audited honest framing** ("not client pitches dressed
up as outcomes") — no fabrication was resurrected. Fleet-wide section drift is now 0
pages; the sweep query is RUNBOOK §5c-bis in `empty_sections_loop_integrity/`.

This is the same *class* as 001 but a different *path* — 001 is the re-planner rewriting
built pages; this was an authority source that never learned about a hand-added section.
~~Fixing it does not close 001, and your "undefined shelf life" warning above still stands.~~
(001 closed 2026-07-20 regardless, see the correction at the top of this file.)

---

## ⚡ PUNCH-LIST 1 (owner review 2026-07-14) — all items now closed or superseded

Status as of **turn 25**. Item 5 (imagery) is superseded by punch-list 2 below.

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

## ⚡ PUNCH-LIST 2 (owner review 2026-07-18) — the current one

| # | Item | Status |
|---|---|---|
| 1 | Tools not linked from nav | **FIXED** — a rebuild had stripped them; a `tools` nav group existed that renders in NEITHER header nor footer. 4 working tools now in the `utility` group (footer). |
| 2 | Blank page behind "Monitoring Coverage Gap Finder" (services) | **ROOT-CAUSED, not fixed** — a phantom link *invented by the re-plan*. See the red box at the top; it is `bugs_open/001`, not a missing page. |
| 3 | Hero image has unreadable text | **FIXED — `/assets/images/hero.jpg` replaced 2026-07-29** (clean Banana hero, same file, so all 14 consuming pages fixed at once; see RUNNING_NOTES same date). ~~The garbled file is still the site-wide fallback~~ |
| 4 | "trust" / "honest" / "earns its keep" overused | **Config done, copy partly done.** Banned in `voice_gate` + `banned_language` (bak_voice_words_20260718). Homepage instances rewritten. The rest are in the 25 `voice_tells` items. |
| 5 | Want infographics showing system strengths | **DONE ×4** — see the imagery table below. |
| 6 | Want more imagery / graphics / better design | **7 heroes + 4 infographics live as of 2026-07-29** (index, who-we-help, how-we-work, about, services, contact, use-cases). Only blog + the 4 tool pages have nothing now. |
| 7 | Want more hero images | **7 live** (index, who-we-help, how-we-work, about, services, contact, use-cases — all added 2026-07-29). ~~about/services need a shared-component change first~~ **CORRECTED 2026-07-29: `about-hero`/`services-hero` already had the image field when checked — someone else added it fleet-wide since this row was written. Only `use-cases-hero` genuinely lacked it; added the same guarded pattern, 2-site blast radius measured first.** See RUNNING_NOTES same date for the real remaining gap this uncovered (the schema's `input_schema` field, not the template, gates whether an image actually resolves). |

Plan for these: `PLAN_imagery_and_design_2026-07-18.md`.

## ⚡ PUNCH-LIST 3 (owner review 2026-07-19) — broken buttons

| # | Item | Status |
|---|---|---|
| 1 | 4 buttons on the tool pages broken/meaningless: *Start Ranking Free*, *See How It Works*, *Start the Guide*, *Visit the Tool* | **✅ FIXED & EXTINCT FLEET-WIDE (2026-07-20).** leopardess tool pages cleaned (P4, verified live); fleet completed + `tool-guide-intro` fixed at component level by migration 179; the structural fix (schema-derived pairing) runs as a live observe-stage. **This grew into its own workstream — resume there, not here: `docs/agent_docs/docs024_key_docs_latest/cta_link_integrity/HANDOFF.md`.** Bug: `bugs_open/023` (+ siblings 033/039/045). |

> The block below is the ORIGINAL 2026-07-19 diagnosis, kept for the record. It is
> superseded by the CTA workstream HANDOFF above — figures here (75 ungated, 34/21/13 items)
> are the 07-19 snapshot and have since moved; do not act on them, use the workstream doc.

**Summary of the diagnosis** (full detail in the plan — do not re-derive):
Four buttons, four *different* mechanisms, each defeating a different check. `href=""`
(warning-only, never blocks); `href="#guide-start"` with no such id (fragment scope is
skipped by every check); a fabricated external host (external scope skipped by every check,
zero HTTP checks exist in `platform/`); and a frozen `source:static` label from a *different
tool* (`bayesian-ranking-hero-tool_pre_037` — a `_pre_037` backup row that is the sole live
component for its function).

Two things matter beyond this site:
- **It is fleet-wide.** 51 dead/suspect controls across 7 of 11 sites; **75 of 89 URL-bound
  CTA anchors in the component library are ungated**, so a missing address renders a dead
  button instead of no button — violating the platform's own LNK-005 invariant.
- **One of the four was correctly detected on 2026-07-17** and filed at
  `needs_human_review`, which nothing consumes (21 `unresolved_cta` + 13
  `cta_names_unknown_destination` open here, oldest 2026-07-13). The delivery gap is a
  separate fix from the detection gap.

⚠️ **Fix at component/schema level, NOT in `page_components`** — `bugs_open/001` gives
anything written there an undefined shelf life on this site.
⚠️ **Do not run the experience loop on this yet** — it is a detection loop and these are
already correctly described in 34 unread items. Build the handler first.
✅ **Owner-approved 2026-07-19 (P4.1): 301 `leopardessconsulting.com` → `.co.uk`, path
preserved.** Owner/DNS action, not blocked on any code work; fixes one of the four buttons
immediately. ⚠️ It makes a *fabricated* URL resolve — not the defect fixed; the field still
invents on the next build.

**Provenance of the fabricated hostnames (corrected 2026-07-19 after owner challenge):** the
model has no knowledge of the owner's domain estate. `leopardessconsulting.com` is just the
obvious `.com` variant of the site name (owner ownership = coincidence).
`leopardess.contactforsales.com` is a transform of the site's own identity-spec contact email
`leopardess@contactforsales.com` (`@`→`.`). **6 sites share `contactforsales.com` as their
contact domain, so any can produce this.** Yields a deterministic check — plan step P1.5.

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
| `tools/ai-vendor-trust-checklist/` | The built tool: plan doc, template, JS, render harness, fence check. **Live** — see the §10.3 correction. |
| `HANDOFF_vendor_trust_checklist_tool.md` | ⚠️ **HISTORICAL.** The brief for a tool that was built the same evening. Do not action it. |

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
| **Added 2026-07-30/31 — read the header of each, the gotchas are in-file** | |
| `refresh_owned_page_chrome.sh` | Chrome refresh for a **protected** page: lifts the no-rebuild flag, rebuilds, **restores the flag even if interrupted**. The only safe way to reach the 5 tool pages. |
| `reconcile_footer_nav.sh` | The safe two-step footer/nav reconcile. ⚠️ Use this, **never the nav-updater agent** — it deletes every nav link under `/tools/`, `/blog/`, `/guides/` and reports success (§11.1). |
| `orchestrate_safe.sh` | Kafka dispatch that **confirms it sent**. The old inline publish quietly lost ~4 in 5 attempts while discarding the output, so a lost publish looked identical to a slow one. |
| `rerender_page_safe.sh` | Single-page rerender that takes the **page_id** (a page *name* is silently rejected — one existing script has the same fault). |
| `commit_tool_asset.sh` | Commit a tool's JS bundle / assets to the git-adapter. |
| `tool_acceptance_run.sh` | Fire browser acceptance at ONE tool (S6 of the staged ladder). **Rewritten 2026-08-11** — raises the work item rather than running inline; see §11.6 for why the old route failed silently. |
| `prove_component_template_identity.go` | Renders every live instance of a shared component through old + new templates and asserts byte-identity, **with the control that the new arm differs**. Use before touching any shared component. |
| `L9_services_carousels.sql`, `L9_info_card_grid_template_with_carousel.html` | The 07-31 services work (reference — and see §11.3, parts of it are no longer live). |

**Trigger any agent** by producing to Kafka `system.agent.generic.requests` — copy the
`kcat` block from `082_submit_domain_unified.sh`; only `agent_type` + `input_data` change.
**DB:** `kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c "<SQL>"`.
**Verify a deploy:** `curl` the page (GitHub Actions → B2 → Cloudflare Worker, ~30–90s),
never trust orchestration status alone.

---

## 6. NEXT ACTIONS (in priority order)

**STATE AS OF 2026-07-18 (all figures checked live, not carried forward):**

*Imagery — the big change this week.* Four infographics are live, all generated through
`kind:'infographic'` → Banana → **`gemini-3-pro-image-preview`**, all reviewed by eye, all
carrying full descriptive `alt` text:

| Page | Graphic | Hero |
|---|---|---|
| index | `infographic-what-we-build.jpg` | `hero-home.jpg` ✅ |
| how-it-works | `infographic-how-a-job-runs.jpg` | `hero.jpg` ⚠️ (the garbled one) |
| case-studies | `infographic-leopardess-line.jpg` (the Beck transit map) | — |
| technical-architecture | `infographic-architecture.jpg` | `hero.jpg` ⚠️ |
| who-we-help / how-we-work | — | good Banana heroes ✅ |
| **about, services, use-cases, contact, blog** | **— none** | **— none** |
| engagement-model, faq, careers, insights | — | `hero.jpg` ⚠️ |

> **CORRECTED 2026-07-18:** an earlier version of `bugs_open/011` claimed *"generated images
> cannot render readable text"* and proposed building an SVG renderer. **Wrong.** The
> capability was already wired: `infographic` has always routed to Gemini, which renders text
> well. The real bug is narrower — `kind:'hero'` falls through to Stability/SDXL, which cannot.
> Caught by the owner showing working Gemini infographics. Pattern recorded in 016b §9
> ("read the dispatch table, not the output").

*Two new checking layers are LIVE in the deployed chassis* (verified in-pod:
`strings /app/agent-chassis | grep -c voice_tells` → 7):
- **claims** (`unverified_claims`) — banned-claim + unregistered-number scan vs the
  `evidence_base` spec. Spec: `docs024_key_docs_latest/claims_verification/`.
- **voice_tells** — LLM-tells scan vs the site's `voice_gate` (banned phrases incl. the owner's
  2026-07-18 ban on *trust / honest / earns its keep*, em-dash + triad density, sentence
  length, contractions). Engine `platform/orchestration/datahelpers/voicetells.go`;
  CLI `cmd/voicescan`. Both are HITL-terminal — they never rewrite.

*Owner review queue* (live counts): **25** `voice_tells`, 21 `unresolved_cta`, 17
`needs_section_data`, 17 `image_source_unsatisfiable`, 13 `cta_names_unknown_destination`,
11 `content_rewrite`, 6 `needs_content_page`. The `voice_tells` list **is** the v1→v2 rewrite
worklist. ⚠️ Some `needs_content_page` / `content_rewrite` items are from `content-gap-planner`
and propose **fabricated** content (it invented a whole financial-services case study with
metrics) — read before actioning; several are deliberately parked at needs_human_review.

*Tools:* 4 of 5 work (ai-agent-roi-estimator, password-entropy,
tool-agent-complexity-estimator, process-automation-scorer — the last three are
self-contained inline JS). **llm-cost-calculator is broken**: references
`bayesian-ranking-hero-tool.js` (wrong tool) while its own bundle 404s. All 4 working tools are
now linked from the footer. Diagnosis recorded on `needs_diagnosis:leo-tools-runtime`.

**✅ `ai-readiness-quiz` is FIXED** (turn 21, verified live 2026-07-18: 54,118 bytes, 3
components). It was blocked by the shared `contact-block` schema whose `email_placeholder`
fallback was `jane@company.com` — the email validator read that as a hallucinated contact and
failed *every* build of *any* page using that component, fleet-wide. Fallback changed to
"Enter your email address"; the page then built. Do not re-open this.

**NEXT ACTIONS (in priority order)**

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
5. **[✅ done turn 18] Titles / meta / sitemap.** 4 A2-violating titles fixed; honest
   meta_descriptions on all 12 key pages that had none; **sitemap.xml LIVE** (27 URLs from the
   pages table via git-adapter — no platform generator exists; a deploy-time `generate_sitemap`
   action would fix this fleet-wide, all sibling sites lack one). robots.txt is
   Cloudflare-managed (search allowed; AI-training bots blocked) — fine as is.
   Remaining polish: og:image per page (currently site-wide og-card).
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
`index`, `who-we-help`, `how-we-work` use `hero` (`background_image` ✓). ~~`about`/`services`
use `hero-about`/`hero-services` (NO image field — needs component work first).~~
**STALE, corrected 2026-07-29: their real component names are `about-hero`/`services-hero`
(this row's names were the CSS class, not the component), and both already declare
`background_image`/`hero_url` in their template — someone added it fleet-wide (12/6 site
consumers) sometime between 07-18 and 07-29. `use-cases-hero` genuinely lacked it and got
the same guarded pattern added, 2-site blast radius measured first. The REAL remaining gate
is a level down: `plan_sections_action.go`'s `sectionHasImageField` only auto-writes
`background_image` when the component's `input_schema` declares an image-typed field —
none of these four did, so having the template guard was not sufficient by itself. See
RUNNING_NOTES 2026-07-29 for the full trace and why the fix was scoped to leopardess's
`content_data` directly rather than the shared schema (a schema edit would auto-populate a
fallback image on every consuming site's next rerender, not merely sit inert until opted
into).** All 4 now have heroes on leopardess (§10). image-build-handler generates →
asset-deployer commits to `/assets/images/<asset_key>.<ext>`.
**Verify `assets.url` is `/assets/images/…`, NOT a presigned `s3…?X-Amz-…` URL.**
For brand consistency, pass the logo as a reference image (Banana kinds only).

**Route B — systematic (build-site-planner re-plan).** Writes the plan + imagery rows +
auto-emits `needs_imagery`. BUT ⚠️ **a full re-plan may re-run content generation and could
overwrite the carefully-fixed copy** — do NOT fire `build-site-planner` blind on this site.
If used, scope it and verify content isn't clobbered. Route A is the safe default.

**Per-card / per-section images = Phase I3.** ~~content-imagery lane — NOT built~~
**CORRECTED 2026-07-29: it IS built** — `derive_card_asset_action.go` (registered,
`asset-deployer`'s `content_card` mode), cover-crops the page's hero to an 800×450 card,
no LLM. Built by the `imagery` workstream (`docs024_key_docs_latest/imagery/`,
"I3.2 ✅ built"); actively being hardened today by `bugfix_131_og_card`
(`bugs_open/143`, latent lock-check race, not live). Never fired for leopardess (0
`purpose='card'` assets) — a triggering gap, not a missing capability. See
RUNNING_NOTES 2026-07-29 before building anything here; check who owns 143 first.
Cards get the page-hero fallback (`imageRoleAliases`) or empty until it's fired. A
section image is wired via a
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

---

## 10. STATE AS OF 2026-07-30 — start here if you already know the rest of this file

Two sessions' worth of work happened on 2026-07-29 after §0–9 above were written (turn 25,
2026-07-18). This section is the actual frontier. Everything below is verified live at
time of writing, not carried forward from a status field.

### 10.1 — Missing/broken images: closed out, five pages remain untouched by choice

The owner's original ask ("fix the missing images") turned into a full pass:

- **The real defect:** `/blog.html` served six `<img src="">` — empty `src`, not a dead
  link. Root cause: `rebuild_blog_listing_action.go:218` hardcodes `image:""` for every
  article (no per-article imagery on this platform), and the shared `content-listing` /
  `category-listing` templates rendered the `<img>` unconditionally. **Fixed**: both
  templates now guard with `{{if .image}}`. Blast radius measured first (only
  `robot-hands.com` shares these components, and its 3 articles all have real images, so
  its output is byte-identical). Commit `679f33685`.
- **`hero.jpg` replaced.** Was AI-generated gibberish-text garbage, live as the fallback
  on **14 pages**. Regenerated via Route A (`kind:'hero'`, now routed to Banana
  platform-wide since `bugs_closed/011` — no more SDXL fallback trap). Same filename, same
  asset_key as purpose, so the new file overwrote the old one and every consuming page
  updated with zero rerenders. Origin model `banana/gemini-3-pro-image-preview`.
- **4 new page heroes**: about, services, contact, use-cases (see the §9 correction above
  for the schema-field gate this uncovered and why the fix was scoped to leopardess's
  `content_data`, not the shared component schema).
- **Expired asset URLs.** Found that `deploy_image_asset` never rewrites `assets.url` off
  its presigned S3 URL unless called with `asset_id` — a known RUNBOOK landmine, but
  apparently never actually exercised by any caller: **every** active hero/infographic row
  here carried a presigned URL, including ones generated minutes earlier the same session.
  Confirmed a REAL (not just theoretical) risk: `derive_brand_head_assets_action.go` and
  `derive_card_asset_action.go`'s `findCardSourceHero` both fetch `assets.url` directly and
  would 401 past the 7-day window — the latter is exactly the mechanism §10.3 below needs.
  Filed **`bugs_open/152`** (platform-wide, unowned). Fixed here: retired one orphaned
  wrong-provider row (`hero_case_studies`, wrong-routing SDXL leftover from turn 17, wired
  nowhere), rewrote `assets.url` to the verified-200 local path for the other 12.
- **Contributed to `bugs_open/128`** (that image-URL check's third blind spot: `src=""`
  has no URL to probe, so its proposed HTTP-check fix would confirm a broken image as `200
  OK`).

**Still not fixed, by choice, not oversight:** blog index + the 4 tool pages carry no
image. Not part of what was asked; noted as a remaining gap, not actioned.

Full account with every command and verification: `RUNNING_NOTES.md`, 2026-07-29 entries
(two separate dated sections — the first is the blog fix, the second is the rest of the
imagery work). SQL for anything DB-config here is NOT saved to `scripts/` (these were
template/data edits, not page inserts) — read RUNNING_NOTES for the exact statements if
you need to reproduce or extend the pattern.

### 10.2 — A 4-part researched feature series is live: "Can You Trust AI With Your Data?"

Owner asked for a thoroughly researched, honestly-argued feature article on trusting AI
with data, cited sources with their opening lines quoted, multiple industry angles,
charts/tools considered, real statistics, both sides argued. Published as one pillar +
three industry deep-dives, all cross-linked, all live:

| Page | url | Words | What it covers |
|---|---|---|---|
| Pillar | `/blog/can-you-trust-ai-with-your-data.html` | ~3,100 | The KPMG trust paradox, an industry tour, both sides, a vendor-evaluation checklist built around Anthropic's own certifications |
| Healthcare | `/blog/ai-data-trust-in-healthcare.html` | ~1,370 | Patient trust fell 52%→44% in 2 years; what's proven to rebuild it |
| Financial services | `/blog/ai-data-trust-in-financial-services.html` | ~1,080 | Adoption outpacing governance; EU AI Act deadline (2 Aug 2026) |
| Hiring/HR | `/blog/ai-data-trust-in-hiring-and-hr.html` | ~1,010 | The widest trust gap found anywhere: 70% of hiring managers trust AI hiring decisions, 8% of job seekers call it fair |

**Every statistic is attributed to a named study with a sample size where available** —
KPMG/University of Melbourne (n=48,000+), Pew Research 2026, Edelman 2026, Cisco 2025
(n=2,600), Reach3/Rival 2026 (n=1,043), Deloitte 2026, Cambridge Judge 2026, McKinsey 2026
(n≈500), IBM's 2025 Cost of a Data Breach Report, Salesforce (n=6,058), Thomson Reuters,
Greenhouse 2025 (n=4,136), Gartner (n=2,918), ResumeBuilder (n=948), Dice (n=319), and
Anthropic's own published trust-centre certifications. Two numbers from a first-pass
search were caught and corrected against primary sources before publishing (a mis-cited
84%→78% Salesforce figure; an unattributed hiring stat traced to a specific Dice report).
**Do not add a statistic to this series without checking it against a named, dated
source the same way** — that discipline is the entire point of the series, and it is
what a `voice_tells`/`unverified_claims`-style scan on this content should hold it to if
one is ever run against it.

**Content is hand-authored HTML, inserted directly as `page_components.content_data`,
NOT run through page-content-writer/page-build-handler.** This matches the site's
established practice (the voice-rewrite pass in §6 item 2 used the same hand-edit route)
and was the deliberate choice here specifically because an automated writer has no way to
enforce "every number traces to a real source" — that discipline lives in the human doing
the research, not in a prompt. Before every insert, both escalation-guard branches were
checked clean (no missing required `content_data`, no missing required `source:"llm"`
schema fields) so nothing ever risked escalating to an LLM rewrite on render.

**5 charts, built as hand-coded inline `<svg>` inside the article HTML — deliberately
NOT using the shared `evidence-chart` component.** That component resolves from
`site_specs.evidence_base`, which this site uses specifically for FIRST-PARTY,
re-queryable facts (`source.sql` / `source.artifact` — something rechecked against this
site's own DB or code). Third-party survey citations have no such backing and do not
belong in that register without blurring the distinction the claims-verification work
exists to draw. If you want an evidence-chart-backed chart on this content later, it
would need to be about LEOPARDESS'S OWN figures, not a cited survey number.

**Platform behaviour worth knowing before writing another linked series here:** a page's
`<a href="...">` to a page that doesn't exist YET gets silently stripped from
`rendered_html` at render time (confirmed: `content_data` kept it, `rendered_html` did
not). Not diagnosed to root cause or filed — the practical fix is mechanical: publish leaf
pages before the hub, or budget one extra re-render of the hub once all targets exist.
That is what was done here (published the 3 industry pages, then re-rendered the pillar).

**Blog listing extended, not rebuilt**: cloned one existing card's markup programmatically
and prepended the 4 new entries to both `content_data.articles` and `rendered_html`
directly, rather than firing the full `rerender-pages` workflow (`rebuild_blog_listing` is
one step in that workflow, but it also runs `get_pages_for_rerender`/
`create_rerender_items`/`render_site_components` — a far wider blast radius than "add 4
list entries" on a site that has been clobbered by a wide rebuild before, historically).
Blog listing now shows 10 cards.

**Known gap, not fixed:** the 4 new pages are not yet in `sitemap.xml` — a separate,
undiagnosed generation mechanism (§6 item 5 already notes no platform sitemap generator
exists fleet-wide; this is that same gap, now affecting new pages too).

Full account: `RUNNING_NOTES.md`, 2026-07-29 "a 4-part feature series" entry. Insert SQL
for all 4 pages + the blog-listing update saved at
`scripts/L8_article_trust_{pillar,healthcare,financial,hiring}.sql` and
`scripts/L8_blog_listing_add_trust_series.sql` — read these before writing a 5th article
in this style; they are the working, verified pattern (dollar-quoted `content_data`/
`rendered_html`, escalation-guard pre-checks, the meta_description `%%`-escaping trap
noted in RUNNING_NOTES — do not sprinkle literal `%%` into an f-string SQL literal, it is
not a printf format string and will land in the DB verbatim).

### 10.3 — ~~Considered and deliberately deferred~~ **BUILT AND LIVE — corrected 2026-08-12**

> **CORRECTED 2026-08-12: the tool was built the same evening this section was written,
> and it is live.** `/tools/ai-vendor-trust-checklist.html` — **HTTP 200 today**, twelve
> checks in four sections, three plain-English verdicts, an N/A path that drops the
> denominator to eleven, entirely client-side. Commit `0bfdf5b2e` (2026-07-30), built up
> the S0–S7 staged ladder with each check written *before* the thing it checks. It is in
> the footer of every page (see §11.1). **`HANDOFF_vendor_trust_checklist_tool.md` is
> therefore a HISTORICAL brief, not a live to-do** — do not "pick it up".
>
> Two things that fell out of building it, both of which outlived the tool:
> - **It found `bugs_open/157`** — `has_visible_area` measures 0 on any axis that is a
>   whole number, so the acceptance run called 24px tick-boxes invisible while the same
>   run had successfully *clicked* them. The tool was deliberately **left alone** rather
>   than nudged to make the check go green. **157 is now CLOSED** (fixed, live in
>   `v1.0.1216`, proven with a negative control — `bugs_closed/157_…`).
> - **It found the tool-naming gate**: browser acceptance could only find a tool when
>   three separate names agreed, which excluded 6 of 22 tools across 5 sites *including
>   this site's own ROI estimator and cost calculator*. **That gate is now gone** — see
>   §11.6.

Original 2026-07-30 text, kept for the record: assessed as a good fit for the pillar
article's "what trustworthy actually looks like" section and technically straightforward
(deterministic client-side scoring, same shape as the site's existing calculators). **Not
built** — it's a separate feature (new JS, new page, needs browser testing per this
platform's own UI-work standard), not a content task, and was scoped rather than rushed.
**Full standalone build handoff, self-contained, for a fresh thread:
`HANDOFF_vendor_trust_checklist_tool.md`** (same directory as this file).

### 10.4 — ~~What a fresh thread should actually do next~~ **SUPERSEDED by §11.7 (2026-08-12)**

> **This list is kept for the record. Item 1 is DONE (§10.3). Item 2 is partly OTBE — the
> tool pages now carry card imagery, added by an automated pass on 08-11, though see
> §11.3 for what else that pass did. Items 3–5 still stand. Read §11.7 for the current
> ordering, which is different because two 07-31 deliverables need restoring first.**

1. ~~If picking up the tool: go straight to `HANDOFF_vendor_trust_checklist_tool.md`, it is
   self-contained.~~ **DONE 2026-07-30 — built, live, footer-linked. See §10.3.**
2. If picking up general leopardess work: the blog index + 4 tool pages still have no
   image (§10.1) — lowest-risk next imagery item, same Route A recipe already proven
   working 6 times over on this site (see RUNNING_NOTES for the exact recipe and prompts).
3. `sitemap.xml` doesn't include the 4 new trust-series pages (§10.2) — either find and fix
   the (currently unlocated) generation mechanism, or hand-add 4 `<url>` entries the same
   way the original 27-page sitemap was produced (§6 item 5 — check RUNNING_NOTES/git
   history around the original sitemap work for how that was actually deployed, since it
   is git-adapter-committed, not DB-driven).
4. `bugs_open/152` (this session's finding: `assets.url` never gets rewritten off its
   presigned form) is unowned and platform-wide — worth a diagnosis-loop run before anyone
   builds more on the card-derivation path in §10.3's tool or elsewhere, since it will
   recur on every future image generation, here and fleet-wide.
5. The remaining sections of the article series (retail, government, legal, SMB) got
   folded into the pillar as supporting evidence rather than becoming standalone pieces —
   if the owner wants any of those as their own deep-dive, the research is already done
   and cited in the pillar; it would need only expansion, not a fresh literature search.

---

## 11. STATE AS OF 2026-08-12 — the frontier

Written by a session asked only to bring this file up to date. **Everything below marked
`[MEASURED 2026-08-12]` was checked against the live DB or the served page today**, not
carried forward from a status field or from another doc. Where I could not establish a
cause I say so with `[UNVERIFIED]` rather than guessing — the attributions are the weakest
part of this section and the measurements are the strong part.

### 11.1 — What this lane shipped after §10 was written (2026-07-30 → 07-31)

**The vendor trust checklist tool** — see the §10.3 correction. Built, live, and the two
platform defects it found are both resolved (`bugs_closed/157`; the naming gate, §11.6).

**It was then linked from the footer of all 34 pages, and the route there is the reusable
part** (`a5b969f92`, `49639b68c`):

- ⚠️ **Do NOT use the nav-updater agent on this site.** It rebuilds navigation from
  scratch and **discards any link whose target lives in a sub-folder**. Every tool page
  here is under `/tools/`, so it would have deleted all the tool links and put none back,
  **on a run that reports success.** Measured blast radius at the time: 16 links across 7
  sites are in that position. Landmine committed in `be9ce5314`.
- The safe route is two smaller steps: refresh the shared header/footer from the nav list,
  then rebuild each page in **ASSEMBLE** mode (reuses existing content, swaps only the
  chrome). The other mode regenerates sections from stored data and **escalates any page
  with a missing required field to the LLM writer** — 5 of 34 pages were in that state,
  so the wrong mode would have had 5 pages rewritten.
- The 5 tool pages are deliberately protected from rebuild, so the ordinary refresh cannot
  reach them at all. `scripts/refresh_owned_page_chrome.sh` lifts the protection, rebuilds,
  and **restores it even if interrupted** — an unprotected tool page can be overwritten by
  anyone else's rebuild.

**`/services.html`: both blocks made carousels, and six links that were never links**
(`2e694bbe3`, 2026-07-31). Full account in `RUNNING_NOTES.md` under that date; the parts
that matter to a future thread:

- The six card links were not broken, they were **not links**. All six pointed under
  `/services/`, which has never existed here. `datahelpers.RepairPageLinks`
  (`link_repair.go:139`) strips an anchor whose target is not a real `pages.url` **and
  keeps the inner text** — so the label shipped as dead prose. That is correct, deliberate
  behaviour; it is what the safety net looks like when it fires.
- Three further instances of the invented `/tools/tool-monitoring-coverage-gap-finder.html`
  were still live **as prose**, which `RepairPageLinks` cannot touch and the phantom-link
  detector cannot see. Removed. **Still 0 today** `[MEASURED 2026-08-12]`.
- The `call-to-action` had labels but no `*_cta_url`, so the template drew **no buttons at
  all** — a page that looks deliberately button-less. Fixed.
- Block B's carousel was added as an **opt-in `carousel: true` flag on the shared
  `info-card-grid`** (18 instances, 9 sites), not a fork — a section-component fork does
  not survive rerender. Byte-identity proven 18/18 with a control that all 18 *differ*
  when the flag is set. **⚠ That flag is gone as of today — §11.3.**
- **`90,790` orchestration state records was a false claim**: a point-in-time count of a
  table pruned hourly at 24h, published as a cumulative total with a durability promise.
  Live count that day: 2,364. The claims layer had caught it on 07-26 and filed it at
  `needs_human_review`; **nothing drained the queue**, so it stayed live 5 more days.

**New operator scripts from these two days** (all in `scripts/`, all documented in-file):
`refresh_owned_page_chrome.sh`, `reconcile_footer_nav.sh`, `orchestrate_safe.sh`,
`rerender_page_safe.sh`, `commit_tool_asset.sh`, `tool_acceptance_run.sh`,
`prove_component_template_identity.go` (the Go template byte-identity harness),
`L9_services_carousels.sql`, `L9_info_card_grid_template_with_carousel.html`.

### 11.2 — What happened to this site WITHOUT this lane (2026-08-08 → 08-12)

**08-08, ~17:12Z — the D3 programme's first supervised improvement-loop run, fired at
leopardess because the owner picked it** (option A of four). Account, with its pre-flight,
in `docs/agent_docs/docs024_key_docs_latest/bugfix_179_deploy_path_override/NOTES_deploy_path_override.md`.

- Reached `COMPLETED` in 9 minutes. **Drained 10 items, minted 68, deployed 0 pages.**
  Open items went 189 → 248. `pages.deployed_at`, `updated_at` and `last_built_at` each
  moved on **zero** rows here (positive control: the same query showed 614 pages fleet-wide
  and a deploy inside the run window on another site — the column moves, it did not move
  here).
- That lane's own conclusion is worth carrying: **"the repairs that ran were sane; the
  loop that runs them does not terminate."** At ~68 minted per ~10 drained, a per-site
  supervised run cannot converge.
- **Its pre-flight cleared the risk this lane had recorded** — that an improvement sweep
  would clobber the hand-corrected specs. No step in the improvement chain has a
  spec-write action (positive control: 9 live agents fleet-wide *do* have
  `write_site_spec`, so the filter can match), and `content-gap-planner` is reachable by
  nothing in that chain. The July-era hazard of fabrications baked into `rendered_html`
  against NULL `content_data` **had also expired** — every `page_component` here has
  `content_data`.
- A real finding it surfaced: `acceptance_run` on `tool-process-automation-scorer`
  returned **7 passed / 2 failed** — `submit-shows-error` failing at *both* `@desktop`
  and `@mobile`, `fix_cycles_spent: 0`. Nobody has picked that up.

**08-11 17:47Z → 08-12 02:07Z — a much larger unattended pass, undocumented in any lane's
notes I could find** `[MEASURED 2026-08-12]`. By `owner_agent_type`: 97 `page-rerender`
(+6 FAILED), 32 `build-dispatch-loop`, 22 `asset-deployer`, 10 `internal-linker`, **5
`page-content-writer` + 5 `page-build-handler`**, 2 `tool-improver`, 2 `section-editor`,
1 `webdesign-agent`, 1 `site-asset-renderer`, 1 `tool-auditor`, 1 `image-build-handler`.
Work items completed in the same window: 220 `page_rerender`, 24 `undeployed_asset`, 20
`needs_internal_links`, 8 `needs_rerender`, 7 `needs_imagery`, 6 `needs_content_image`,
5 `improve_tool`, 4 `section_edit`, 3 `audit_tool`, 2 `generic_theme`.

The writer targeted `llm-cost-calculator`, **`case-studies`**,
`case-study-automated-intelligence-pipeline`, `ai-agent-roi-estimator` and
`tool-automation-savings-estimator`. **`/case-studies.html` came through untouched** —
its four components still read `updated_at = 2026-07-18 15:27:18` — so the audited honest
framing survived. That is luck plus a failed run, not a guarantee.

### 11.3 — ⚠ Two 07-31 deliverables are NO LONGER LIVE, and a false claim came back

All three found by checking the artefact, not by any alert. Nothing on this platform
reported them, because in each case the template renders **less** rather than erroring.

**(a) The six service images do not render. `/services.html` has exactly ONE `<img>` tag
on the whole page** `[MEASURED 2026-08-12]`. All six `teaser-reveal-panel` items carry
`image_url` as an **empty string**, and item 6 has lost `image_url`, `image_alt` and
`open_label` **entirely**. The template gates on the field, so six missing images render
as no error, no empty box, nothing. This is the exact damage class of **`bugs_open/238`**
(a regeneration replaces `content_data` wholesale and drops resolver-sourced `_url` keys;
gated templates hide it). 238's fixes were committed 2026-08-11 and are **inert until the
chassis image next rolls**. *Which* handler emptied these is `[UNVERIFIED]` — the
components' `updated_at` is `2026-08-11 18:15:23`, inside the rerender wave.

**(b) Block B is no longer a carousel.** The `carousel` key is **absent** from the
`info-card-grid`'s `content_data` — it now holds only `cards`, `section_title`,
`section_eyebrow`, `section_subtitle` — and the served page carries no `data-hcc-*`
markup `[MEASURED 2026-08-12]`. Block A's carousel (`trp__track`) **does** still work, and
`snippets.js` is still 13,781 bytes / 2 snippets, so the JS half is intact: this is purely
the lost opt-in flag. Same class as (a).

**(c) The six card links were repointed by the automated linker, and one of them 404s.**
This lane set 3 × `/case-studies.html`, 2 × `/technical-architecture.html`, 1 ×
`/how-it-works.html`, all verified 200 first. Live today: `/blog/hierarchical-multi-agent-orchestration-explained.html`,
**`/case-study-automated-intelligence-pipeline.html`**, `/case-studies.html`,
`/how-it-works.html`, `/technical-architecture.html`, `/use-cases.html`. The second one
**returns 404** `[MEASURED 2026-08-12]`: its `pages` row was created 2026-08-11 16:21 with
`build_status='planned'` and was never deployed. **`RepairPageLinks` cannot save you here**
— the row exists, so the link is "real" by its test, and the visitor gets the 404 that the
07-31 fix existed to eliminate. Only `/services.html` carries it.

**(d) A verified-false provider claim is back on the live page.** `/services.html` item 5
(`model-routing`) now reads *"A workflow step can call Claude, Gemini, **Mistral** or
another provider…"*. On 2026-07-31 this lane drove `OpenAI|Mistral|xAI|Perplexity` to **0**
on this page; today it is **1** `[MEASURED 2026-08-12]`, in a component updated
2026-08-11 18:15:23. `platform/aiservice/factory.go:23-33` supports exactly **anthropic,
ollama, gemini**; `openai` returns "not yet implemented" and everything else hits
`default: unsupported AI provider`. Two other components still carry a Mistral claim:
`/guides/llm-cost-calculator-guide.html` (**deployed and live**) and
`/for-engineering-teams.html` (archived, harmless).

### 11.4 — What I checked and found STILL GOOD, so you don't check it twice

| Thing | `[MEASURED 2026-08-12]` |
|---|---|
| The 4-part trust series | All 4 pages 200. **All 5 inline SVG charts present** (2 on the pillar, 1 on each industry page). KPMG / Edelman / Cisco / Deloitte / Salesforce citations all still in the pillar |
| `/case-studies.html` | Components untouched since 2026-07-18; the honest framing survives verbatim — *"Those eight are ours. They are not clients, and we would rather say so"* |
| The `gap-finder` phantom | **0** occurrences on `/services.html` |
| The `90,790` false claim | **0** occurrences; replaced by a durably-phrased "more than 2,000 orchestration state records" |
| The hand-authored card copy | Intact on `/services.html` — the 07-31 prose was not rewritten |
| Tools and guides | **7 tools + 5 guides, all deployed, all 200.** All 7 are in the footer |
| Site size | **48 pages, 33 deployed.** Newest deploy 2026-08-12 02:07:13Z |
| `snippets.js` | 13,781 bytes, still serving both snippets |

New since §10: `/tools/automation-savings-estimator/index.html` (created 08-08) and five
`/guides/*` pages, four of them deployed on 08-11. `/tools/llm-cost-calculator.html` no
longer references the wrong `bayesian-ranking-hero-tool.js` bundle — it now carries 4
inline `<script>` blocks, 9 inputs and 4 buttons. **Whether it actually computes is
`[UNVERIFIED]`** — not driven in a browser this session; §6's "llm-cost-calculator is
broken" should be treated as superseded-but-unproven, not as fixed.

### 11.5 — Figures on the site that are stale or false RIGHT NOW

The 07-31 re-measurement pass only covered `/services.html`. **It was never swept across
the rest of the site, and `/case-studies.html` is still publishing the old numbers**
`[MEASURED 2026-08-12]`:

| Live on `/case-studies.html` | Actual today | Note |
|---|---|---|
| "143 agent definitions, 56 of them active" | **193 definitions, 187 active** | badly understated |
| "75,061 orchestration state records" | `orchestration_states` holds **5,997 rows** | **the same defect class as `90,790`** — this table is pruned, so *no* cumulative total belongs on a page. Rephrase it durably or drop it |
| "Eight live sites… Those eight are ours" | the `sites` table holds **40 rows** | ⚠ the 07-31 audit counted **15** by its own definition of "live". **Re-derive the definition before republishing any number here** — 40 is a row count, not a claim |

The claims layer already knows about some of this: **3 `claims_unverified` items sit at
`needs_human_review`**. That is the same queue nobody drained for five days in July.

### 11.6 — Two platform changes since §10 that directly affect this site

**The tool-naming gate is gone (owner decision 2026-08-11).** `url_field` is live on the
acceptance `request_run` step (migration 384) and is checked **before** the name lookup, so
a tool page is now resolved **by component placement** and its URL carried in
`spec.page_url`. This site's ROI estimator and cost calculator are testable without being
renamed. `scripts/tool_acceptance_run.sh` was rewritten to match (`3a91684bd`,
`585e37dad`) — it now raises the same `site_work_items` row the due-sweep raises, so
`build-dispatch-loop` spawns a dedicated pod. **The old kcat route ran the agent inline on
a standing chassis pod that deliberately has no storage client, so the *vision* half of
acceptance failed silently on every manual run while the check half read green.** Read the
header of that script before firing it; three things must still line up and two fail
quietly.

**This site is 1 of only 2 sites fleet-wide that has opted into `voice_gate`** (14
patterns; `oufe` has 10) — the mechanism is real and code-enforced, the adoption is the
defect. The `fleet_copy_quality` lane's 2026-08-12 contribution notes two things owed
here: **33 `voice_tells` items sitting unactioned**, and **~12 leopardess pages still
carrying the banned word "honest"** (owner's 2026-07-18 ban, in this site's own spec for
25 days). Its explicit ask is *"your site, your voice"* — the measurement and the method
are in
`docs/agent_docs/docs024_key_docs_latest/fleet_copy_quality/CONTRIB_2026-08-12_the_honest_ban_and_the_voice_gate_nobody_opted_into.md`.
⚠ Its trap is worth repeating: a find-and-replace that removes the target word 100% of the
time can still be wrong 100% of the time — assert on **shape** (double commas, dangling
articles) afterwards, not just on the word being gone.

### 11.7 — What to do next, in order (supersedes §10.4)

1. **Restore the three `/services.html` regressions** (§11.3). Together they are one
   session's work and they are the only items here where the site is *worse* than it was:
   re-point the 404 card link; restore `carousel: true`; restore the six `image_url`
   values; fix the Mistral claim (and the live `/guides/llm-cost-calculator-guide.html`
   copy of it). ⚠ **Do this with the escalation guard pre-checked, as §10.2 describes, and
   expect it to be undone again** until `bugs_open/238`'s fix rolls — so verify at the
   served page afterwards, and re-verify after the next fleet roll.
2. **Sweep the false and stale figures off `/case-studies.html`** (§11.5), and while you
   are there decide whether any cumulative claim should exist at all against a pruned
   table. Then drain the 3 `claims_unverified` items rather than leaving them.
3. **Take the voice work** (§11.6): ~12 pages carrying "honest", 33 `voice_tells` items.
   This is the largest owner-visible quality item outstanding and the other lane is
   waiting on this site's owner-lane to do it.
4. **`bugs_open/248` (the `undeployed_asset` slug — the number is shared with an unrelated
   CTA bug, so resolve by slug) is firing on this site.** 15 asset rows here carry the
   literal placeholder URL `/assets/images/input-data.asset-key.jpg` — six `icon_service_*`
   rows from 07-31 and six `content_hero_tool_*` rows from 08-08 — so several distinct
   images resolve to one file. Fleet-wide it is **76 rows across 12 sites, 2026-01-28 →
   2026-08-11** `[MEASURED 2026-08-12]`, i.e. still producing new rows. The bug is filed,
   CONFIRMED by a `090` run, and OPEN; **do not re-file it**, contribute the leopardess
   numbers to it. This is the most likely cause of §11.3(a) but I did **not** establish the
   link `[UNVERIFIED]`.
5. **`bugs_open/152` is still open and still unowned** — `assets.url` never gets rewritten
   off its presigned S3 form. Two rows here have already regressed to presigned URLs
   (`hero_case_studies`, recreated 08-08 after the 07-29 session retired it, and
   `content_hero_tool_automation_savings_estimator`, 08-11), so this recurs on every
   image generation exactly as §10.4 predicted.
6. **The `tool-process-automation-scorer` acceptance failure** (§11.2): 2 of 9 checks
   failing at both viewports, `fix_cycles_spent: 0`, unlooked-at since 08-08.
7. §10.4 items 3 and 5 still stand unchanged: the trust-series pages are still missing from
   `sitemap.xml`, and the four unwritten industry deep-dives are still fully researched
   inside the pillar if the owner wants them.

**The standing question this section raises for the owner, which is not mine to answer:**
this site is now both a hand-curated showcase *and* a target for the fleet's automated
improvement loops, and §11.3 is what that combination costs. Either the loops are told to
leave it alone, or this lane accepts that verification here has a shelf life measured in
days and re-checks on a schedule. Right now it is neither, which is why two deliverables
sat broken for eleven days with nobody noticing.
