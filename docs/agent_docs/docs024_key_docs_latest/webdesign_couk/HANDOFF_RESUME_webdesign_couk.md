# HANDOFF — webdesign.co.uk (cold start)

**Written 2026-07-26.** Read this first, then `NOTES_webdesign_couk.md` for the
misstep log. Everything below was verified live, not inferred.

> **PHASE 2 STARTS HERE →
> [`HANDOFF_2026-07-27_phase2_uk_authority.md`](HANDOFF_2026-07-27_phase2_uk_authority.md)**
> The owner's 2026-07-27 direction: UK focus, copy rewrite ordered by popularity,
> third-party UK tool directory, renewed AI focus, a news section, added
> gradually over a month. **It opens with the one thing that must be decided
> first — we have no popularity data at all, and the site is one day old.**
> This document remains the reference for how the site is BUILT.

---

## In one paragraph

`webdesign.co.uk` is a chassis-managed site that merges two of the owner's
hand-built static sites — `website-design.com` (55 tools, 23 articles) and
`websitedesign.com` (10 tools, 10 guides) — into one home, in the **warm
minimalist** design, carrying every feature except `websitedesign.com`'s
client-side LLM builder. **It is live and complete**: 98 pages, all serving 200.
What remains is browser QA of the interactive tools and a few small follow-ups.

---

## Key identifiers

| thing | value |
|---|---|
| domain | `webdesign.co.uk` |
| site_id | `6b49db8e-d447-4467-8277-4f3018af9897` |
| submission corr | `85878c43-5b8a-4a49-8256-470b798f3bea` |
| deploy repo | `gqls/sites`, dir `webdesign.co.uk/` (B2 + Cloudflare, no config needed) |
| sources | `~/projects/sites/website-design.com`, `~/projects/sites/websitedesign.com` (untouched, still live) |
| workstream dir | `docs/agent_docs/docs024_key_docs_latest/webdesign_couk/` |
| port tool | `cmd/webdesignport` (`harvest` \| `transform` \| `import` \| `verify`) |

**Content**: 63 tools + 31 learn pages + about = 95 catalogued; + 2 generated
indexes = 97 owned pages; + 1 chassis-composed home = **98 pages**.

---

## What is live (verified by curl, 2026-07-26)

- `/`, `/tools/index.html`, `/learn/index.html`, `/about/index.html` → 200
- All 94 tool/article pages → 200
- Tool engine JS (e.g. `/tools/bayesian-rank/bayes.js`) → 200
- Header JS `/tools/assets/webdesign-couk-header.js` → 200
- `port-compat.css`, `search.json` (95 entries), `sitemap.xml` (98 urls),
  `robots.txt`, `404.html` → 200
- All 12 generated images incl. `/assets/images/hero-home.jpg` → 200
- Chrome (sticky header, Tools/Learn/About nav, search box, footer) on every page
- Zero occurrences of the old Swiss blue `#0055ff`; zero dark-mode blocks

---

## How it is built — the five things you must know

**1. The design is pinned, and the pin held.**
`site_specs` aspect `design_intent` carries `palette.reference_values` (8 slots)
and `typography.reference_values`. Without it, `webdesign-agent` invents a
palette on *every* run. `SQL_p3_design_intent_pin.sql`. Verify:
```sql
SELECT data->'palette'->'reference_values' FROM site_specs ss
JOIN sites s ON s.id=ss.site_id
WHERE s.domain='webdesign.co.uk' AND aspect='design_intent' AND is_current;
```
Expect `background #f9f8f6`, `primary #5c6b5d`. **A pin governs colour values,
not component selection** — see "the hero lesson" below.

**2. Chrome is forked per-site.** Three `content_components`
(`webdesign-couk-head` / `-header` / `-footer`) bound to slots by **explicit**
`site_components` rows. Never rely on the global fallback: there are five active
`site-header` rows fleet-wide and the lookup is `ORDER BY name LIMIT 1`.
`SQL_p5_chrome_forks.sql`. The header carries the search engine as its
`js_content` **and its own `<script src>` tag** — the assembler injects none.

**3. The 97 content pages are `rebuild_policy='owned'`.** Their HTML lives in
`page_components.rendered_html`, written by `webdesignport import`. Generic
writers refuse to touch them (`save_page_sections` hard-refuses owned pages).
`in_header/in_footer=false` except the three top-level ones.

**4. The port is reproducible, and `catalogue.json` is generated.**
Only `catalogue_additions.json`, `colour_map.json` and `overrides.json` are
hand-edited. `harvest` fails if any source page is neither catalogued nor
explicitly dropped with a reason — nothing can be lost silently.

**5. The indexes are generated from the manifest, not authored.** Both source
indexes were already stale (51 of 55 tools, 13 of 23 articles linked), so
porting them would have carried the omissions over.

---

## To change content: the loop

```bash
# 1. edit port/overrides.json (or colour_map.json / catalogue_additions.json)
/path/to/wdport harvest      # only if the catalogue changed
/path/to/wdport transform    # -> build/output/webdesign_couk (gitignored)

# 2. push static assets (idempotent; skips unchanged)
bash scripts/webdesign_publish_assets.sh

# 3. import changed pages only (sha256-keyed)
kubectl -n ai-persona-system port-forward svc/postgres-clients 5433:5432 &
PW=$(kubectl -n ai-persona-system get secret postgres-clients-secret -o jsonpath='{.data.POSTGRES_PASSWORD}' | base64 -d)
go run ./cmd/webdesignport import --dsn "postgres://clients_user:${PW}@localhost:5433/clients_db?sslmode=disable" --dry-run
```

Then queue re-renders. **Budget 3.5 hours for a full-site pass.**

---

## THE OPERATIONAL FACT THAT COST A DAY

**A site-wide rerender of ~98 pages takes ~3.5 hours and shows NOTHING for the
first ~20 minutes.**

- Items created → **first claim 20m40s later** (documented publish→run-start
  latency).
- Then ~2.1 min per page, single-flight per site → 3h28m for 98.
- **Priority is absolute across item types on one site.** Setting
  `page_rerender` to priority 5 starved priority-90 imagery for the entire 3.5
  hours, then imagery resumed 18 seconds after the last page finished.

I read that starvation — which my own priority change caused — as evidence that
`page-rerender` was broken, and filed it as `bugs_open/003` spawn loss. **It was
not a bug at all.** Full account in NOTES and `WRONG_CALLS.md`.

**Before ever concluding "dispatch is broken", run this:**
```sql
SELECT item_type, min(created_at), min(claimed_at), max(completed_at), count(*)
FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
WHERE s.domain='webdesign.co.uk' GROUP BY 1 ORDER BY 1;
```
Items claimed *before yours existed* say nothing about yours. `attempt_count=0`
means *not yet tried*, which is indistinguishable from *never will be* until the
documented window has actually elapsed.

---

## The hero lesson (worth internalising)

The planner built exactly the one page the mission brief allowed — and chose a
**dark full-bleed hero** for it, painting `rgba(0,0,0,0.5)` over a background
image. That contradicted the brief *and* `design_intent.avoid`.

The pin was not at fault and could not have helped: the darkness was a literal
`rgba()` in the component's own template, drawn from no palette. `avoid` is prose
the planner's component choice never reads.

**So: reviewing the planner's page LIST is not enough — review its SECTION list
too.** I checked the former and not the latter. Fixed by `SQL_p6_two_column_hero.sql`
(a per-site `webdesign-couk-hero`, forked rather than editing the shared `hero`
that six other sites use).

---

## Open items

| # | item | notes |
|---|---|---|
| 1 | **Browser QA of the tier-1 tools** (~16 pages) — *but see below first* | canvas / file / clipboard / IndexedDB tools need a real browser: `micro-cms`, `bg-remover`, `image-optimizer`, `meme-generator`, `favicon-maker`, `animated-favicon`, `white-balance`, `mesh-gradient`, `golden-ratio`, `social-card`, `blob-maker`, `pasteboard`, `mind-map`, `logic-architect`, `vibe-equalizer`, `blueprint-compiler`. Tiers are in the manifest (`qa_tier`). |
| 2 | ~~Confirm the hero swap landed~~ | **DONE 2026-07-26.** Live and verified: two-column, light, both CTAs, zero dark overlay in the hero section. Note the mechanism in NOTES — no rerender path re-renders a section whose COMPONENT changed; the template had to be executed directly. |
| 3 | **`webdesignport verify` is still a stub** | The gates exist as ad-hoc greps in NOTES/RUNBOOK. Worth implementing: structural diff, colour gates, link closure, live 200s. |
| 4 | **`[OWNER-CHECK]` Cloudflare purge token** | Confirm the deploy Action's `CF_API_TOKEN` can see this zone. Symptom of failure is only a stale cache; check for a null `ZONE_ID` in the Action log. |
| 5 | **`bugs_open/084`** (rewritten 2026-07-26) | **A browser-verification tier already exists and is live** — Tier 4 drives the deployed page in real headless Chromium (`internal/adapters/browserrunner/`, v1.0.1167): real clicks, post-interaction assertions, console-error capture. My first version of 084 wrongly claimed nothing did this. The real gap is *coverage*: Tier 4 gates on `component_level='tool'` + a criteria fence, so **none of this site's 97 owned pages is ever browser-tested**; and `asset_loads` checks a script path is *mentioned*, never that it *loads*. |
| 6 | **Council review** | The platform-code footprint here is `cmd/` and docs only, so the gate is not required. If any `platform/` change emerges from 084, it should go through. |
| 7 | **12 `undeployed_asset` items** (new 2026-07-27) | The immune system claims the generated images were never deployed. **All 12 return 200** (`/assets/images/hero-home.jpg` etc.), so this is either a detector false positive or a DB deploy flag never set. Establish which — do not guess. |
| 8 | **Old domains** | `website-design.com` and `websitedesign.com` are untouched and still live. Owner deferred the redirect/canonical decision. |

---

## The best available next move

**Wire the ported tools into the existing Tier-4 browser runner instead of
clicking them by hand.** The machinery is built, live and proven
(`internal/adapters/browserrunner/run_checks_action.go`, v1.0.1167 — Playwright
Chromium, real `fill`/`click`/`select`, post-interaction DOM assertions,
`console.error` capture, desktop + mobile). It simply never looks at pages like
ours, because `check_tool_acceptance_due.go:51` gates on
`cc.component_level = 'tool'` and ours are `'section'`.

Widening that predicate — or letting an owned page carry acceptance criteria in
`content_data` — converts open item 1 from ~16 manual browser sessions into a
recurring automated check, and would cover the whole fleet's owned pages at the
same time. See `bugs_open/084` fix candidate 3.

Read `travelling_docs/OVERVIEW_self_verifying_tools.md` (the Tier 0/1/2/4 ladder)
before touching any of it. **Do not build a browser tier — there is one.**

---

## Landmines specific to this site

- **`content_components.name` is NOT NULL with no default** — every doc example
  omits it and the insert fails.
- **`build_status='planned'` on an owned page is a mistake** —
  `write_build_items` sweeps planned pages into the generic pipeline. Import
  writes `'deployed'`.
- **Data-modifying CTEs**: the `INSERT` must select **from the `UPDATE`'s own
  `RETURNING`**, or the supersede has no ordering guarantee and you collide with
  `idx_site_specs_current`. This is why `robot_hands/SQL_..._r1b` is shaped as it is.
- **`read` with `IFS=$'\t'` collapses tab runs** — an empty field shifts every
  later field left. The asset publisher uses `|` with an explicit `-` placeholder.
- **base64 as argv blows ARG_MAX** near 1MB; the publisher pipes via `--input -`.
- **The airlock watcher reads its allowlist once per cycle** — append, *wait one
  interval*, then flip, or the release races and the item is re-parked.
- **Counts are substituted, never typed** (`{{TOOL_COUNT}}`). A hand-typed count
  produced an invented statistic twice, once reaching eight live specs.

---

## Files that matter

```
docs/agent_docs/docs024_key_docs_latest/webdesign_couk/
  PLAN_2026-07-25_webdesign_couk.md      decisions + 4 dated corrections
  NOTES_webdesign_couk.md                the misstep log — read this
  RUNBOOK_webdesign_couk.md              commands with their gotchas
  README_where_we_are.md                 owner's plain-prose log
  HANDOFF_RESUME_webdesign_couk.md       this file
  SQL_p3_design_intent_pin.sql           the palette pin
  SQL_p4_fix_tool_count.sql              64 -> 63 across 8 specs
  SQL_p5_chrome_forks.sql                head/header/footer + slot binding
  SQL_p6_two_column_hero.sql             the hero fix
  port/                                  catalogue, colour map, overrides, site_assets
  scripts/watch_park_webdesign.sh        the cascade airlock

cmd/webdesignport/                       the port tool (+ checkScriptParity gate)
scripts/webdesign_publish_assets.sh      static asset publisher
docs/agent_docs/sql_for_agents/208_webdesign_ported_page_component.sql
bugs_open/084_..._javascript_with_no_signal.md
```
