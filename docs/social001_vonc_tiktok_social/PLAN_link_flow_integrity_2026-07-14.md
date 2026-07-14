# Plan: Fleet-generic link/flow integrity + vonc.com repair + Arena build

*Status 2026-07-14: drafted locally, handed to Ultraplan for remote refinement
(https://claude.ai/code/session_01Bpbm91nJC7bYze7QLnZTAv). This file is the
local record of the draft as handed off; the refined plan supersedes it once
approved and teleported back.*

## Context

vonc.com ("Spark", site_id `9ec3b9ee-5b08-461b-b4f8-9e1e03579c74`) looks broken in two owner-flagged ways: CTAs that promise one thing and link another ("Enter the Gauntlet" → `/contact.html`), and "tools/games" that read like destinations but don't exist ("the Arena"). The owner wants this fixed **generically** — so any new domain with incomplete multipage tools or bad links gets caught and fixed by the platform, not by hand.

Investigation (2026-07-14, verified against live DB + code) found:

1. **Root cause — schema-level CTA misdirection, fleet-wide.** `hero.cta_url`/`secondary_cta_url` and `call_to_action.primary_cta_url`/`secondary_cta_url` carry `"source": "pages.contact"` / `"pages.services"` in `content_components.input_schema` — a literal name lookup (`plan_sections_action.go`, `sourceResolver.resolve`, `pages` case ~line 420) with zero copy-awareness, written into `resolvedData` **unconditionally on every render**. This is the URL-field twin of the 090 stat-labels bug: no `content_data` edit can win. Proof: migration 089 pointed vonc's CTAs at the Gauntlet; the next render silently reverted all of them to `/contact.html`. Fleet impact: 75/75 `call-to-action` and 68/69 `hero` CTA instances across 8/10 sites → `/contact.html`.
2. **The existing mitigation (`internal-link-resolver`) is crippled**: (i) only runs on full content-writer rebuilds, never rerenders; (ii) only picks `page_type='section-index'` hubs — it can never choose a `tool`/`game` page, so "Enter the Gauntlet" could never resolve to the Gauntlet; (iii) on a miss it silently keeps the phantom value. vonc has exactly one hub, so it mostly missed.
3. **No discovery check sees any of this.** `phantom_internal_links` (audit-time backstop, already written) is registered in Go but **enabled on no agent** (verified live). It also only catches hrefs with *no* page behind them — a CTA landing on a real-but-wrong page passes. Nothing checks copy-vs-target coherence or page-group completeness.
4. **vonc concrete defects**: 19 misdirected CTAs across 10 pages, 2 true 404s on `/about.html` (`/how-it-works`, `/how-it-works/the-gauntlet`), 1 circular CTA ("Enter today's Arena" → `/index.html`), both tool pages (`in_header=true`) absent from `site_nav_items`. Gauntlet and Quiz are real deployed tools; **"Arena" has no page anywhere** — a v3-deferred concept (spec: competitive mode — provocations, reactions Genius/Delusional/Suspicious/Based/Cursed, remix chains, duels) that leaked into copy as a destination. A previous session explicitly parked the CTA retarget "until the real arena exists".

Owner decisions (2026-07-14): full generalized fix (checks + root cause + vonc repair), fix-and-broaden the existing resolver (not a new query-verb source), and **build Arena as a real page/tool**.

## Verified mechanics to reuse (do not reinvent)

- **Discovery checks**: one self-registering file each in `platform/orchestration/actions/discovery_checks/` (`func init() { Register(...) }`); the named import in `platform/orchestration/actions/discovery_checks.go:21` fires every `init()` — no import edits needed. Enabled via `agent_definitions.default_config → {workflow,steps,run_checks,config,checks}` (verified step name `run_checks`); unregistered names warn+skip, so SQL can land before the image.
- **Link helpers**: `platform/orchestration/datahelpers/links.go` — `ExtractHrefs`, `ClassifyLinkScope`, `NormalizePagePath`, `PageURLSet`. All new link logic must use these (shared with deploy gate + phantom check).
- **Work items**: checks emit `WorkItemSpec` → `insertWorkItem` (`load_work_item_actions.go` ~953–1072; `item_key` dedup, two-strike rule) → `build-dispatch-loop` dispatches dynamically via `current_item.handler_agent`. Discovery is manual-trigger (`scripts/initial_messages/170_work_item_flow_build/075_trigger_discovery.sh <domain> completeness`); improvement scheduler off since 2026-05-02.
- **Same-package reuse**: `resolve_internal_links_action.go`, `rerender_page_sections_action.go`, `render_site_components_action.go` are all `package actions` — `chooseCTATargets`/`loadContentHubs`/`ctaFieldNames`/`setCTAField` are directly callable across them.
- **Migration convention** (090 + b4/b5 precedent): numbered SQL in `docs/social001_vonc_tiktok_social/minilobby_task/`, header comment (root cause, blast radius, reversal), `BEGIN/COMMIT`, `CREATE TABLE _backup_... AS SELECT`, `jsonb_set` touching only `source` (preserve type/on_missing/guidance), live `content_data` repair, `DO $$` verification block that hard-fails.
- **Tool creation** (Gauntlet-class, zero bespoke Go): `tool-generator` → `create_tool_component` (LLM-generated widget, `component_level='tool'`, `created_from='generated'`, tool-doc sentinel header enforced). Known gaps to plan around: **TP-002** (never enqueues final render — dispatch manually), **TP-004** (`reconcile_site_plan` tool route commented out — don't rely on it), **TL-001** (a generic full page rebuild clobbers the widget row — future edits must use section-editor targeted path).
- **Landmines** (do not trip): vonc `provocation-card`/`lobby-grid` are deliberate runtime-fill shells — their blank `content_data` is correct, LEAVE ALONE. After any vonc `reconcile_site_plan`, park the re-emitted `needs_page:provocation` item back to `detected`. Deploy: bump IMAGE_TAG (makefile line 16), verify symbols in pod via `grep -ac <symbol> /proc/1/exe`, image before seeds, no orchestration fires within ~5 min of pod start. Work on branch `085_debug_and_feature_loops`, forward-only.

---

## Workstream 1 — Root-cause fix (Go, one chassis image)

### 1a. Broaden `chooseCTATargets` to interactive pages
`resolve_internal_links_action.go`: add a `loadInteractivePages` loader (`SELECT name, title, url, nav_order FROM pages WHERE site_id=$1 AND page_type IN ('tool','game') AND status IN ('active','deployed')` — mirror `loadContentHubs`). Ranking rule v2: interactive pages first (by nav_order), then `section-index` hubs; keep `areasExcludedFromCTA` and self-exclusion. Result: on vonc, primary → Gauntlet, secondary → Quiz; sites without tools behave exactly as today.

### 1b. Copy–destination coherence hint
The resolver runs *before* copy generation (`resolve_links → select_sections → process_sections_loop`), so instead of matching copy to URLs after the fact, make copy follow the URL: when `setCTAField` writes a URL, also write `cta_target_title` (the target page's title/name) into `resolved_data`. Then update `page-content-writer`'s section guidance (agent config SQL, not Go) so CTA text is written *for the actual destination*. Guard against the known "prompt seams dropping spec intent" trap — verify the guidance actually reaches the section prompt.

### 1c. CTA recompute on rerender (the repair path for already-deployed pages)
`rerender_page_sections_action.go` (~lines 202–274, the per-section loop): if the section's component function is in `ctaFieldNames`, recompute CTA URLs via `loadContentHubs`+`loadInteractivePages`+`chooseCTATargets` into the section's resolved data — same pattern as the existing `query.*`/image re-resolution. Add a new `spec.reason` value `"cta_links_stale"`: stamp it in `create_rerender_items_action.go` and add it to the `page-rerender` workflow's `check_rerender_mode` OR-list (agent config SQL). This gives every site a **safe, no-full-rebuild, no-widget-clobber** repair: emit `page_rerender` with `reason: cta_links_stale`.
**Exception rule**: recompute must NOT override an explicitly authored URL — if `content_data` already holds a cta_url that resolves to a real page (validated via `PageURLSet`), keep it. Only replace values that are phantom or land in `areasExcludedFromCTA`. This preserves hand-authored links (e.g. vonc's archetype CTAs after WS3).

### 1d. Header CTA validation (small hardening)
`render_site_components_action.go` (~142–158): the header `cta_url` is "the nav item labelled Contact". Validate it against the real-page set; if absent, fall back to `chooseCTATargets` over hubs. (The 3 header components with `pages.contact` schema source are all `is_active=false` — no schema migration needed for them.)

### 1e. New discovery check: `misdirected_cta`
New file `platform/orchestration/actions/discovery_checks/check_misdirected_cta.go`, template: `check_phantom_internal_links.go`. Add a `datahelpers` helper extracting `(anchor text, href)` pairs (extend `links.go`; `ExtractHrefs` alone drops the text). Scan deployed `page_components.rendered_html` for anchors inside CTA-bearing components. Two deterministic findings:
- **`misdirected_cta`**: anchor text token-matches a real page's name/title (prefer `tool`/`game` pages; normalize case/stopwords) but href points elsewhere. Work item → `page_rerender` items with `reason: cta_links_stale` (handler `page-rerender`), spec includes `suggested_target`.
- **`cta_names_unknown_destination`**: anchor text names no real page AND href lands in an excluded area or on the page itself → `needs_human_review` (this is the Arena case: a product decision, not auto-fixable).
Both use `item_key` dedup keyed on page+href+text.

### 1f. New discovery check: `incomplete_page_group`
New file `check_incomplete_page_group.go`. For each `parent_section` group in the current `site_plan_pages` (plus `page_type IN ('tool','game')` pages as an implicit "interactive" group), if ≥1 sibling is deployed and ≥1 is missing/`planned`/stale (reuse `reconcile_site_plan_action.go`'s `decideEmit` logic), emit one finding per gap: `needs_page` → `page-build-handler` for ordinary content roles; `needs_human_review` for `tool`/`game` roles (TP-004: the generic builder would produce a widget-less prose page). This is the generic "incomplete multipage tool/game" detector.

**Tests**: unit tests for the anchor-text extractor, token matcher, ranking rule, and rerender-exception rule, following existing check test conventions.

---

## Workstream 2 — SQL migrations (after image; `docs/social001_vonc_tiktok_social/minilobby_task/`, next numbers 091+)

1. **091 — CTA source flip (fleet)**: on active `hero` + `call_to_action` components, remove/inert the `pages.contact`/`pages.services` source on the four URL fields (keep `type`, `on_missing: skip_field`, guidance) so the resolver + authored `content_data` are the only writers. Exact inert form (drop key vs `"renderer"`) decided by reading the field-loop's absent-source handling in `plan_sections_action.go` (~1101–1296). Backup tables + `DO $$` verification per 090 convention. Safe: only bites on next render, and 1c's exception rule preserves valid authored URLs.
2. **092 — enable checks**: append `phantom_internal_links`, `misdirected_cta`, `incomplete_page_group` to `completeness-discovery-agent`'s `run_checks` checks array (idempotent append pattern from `074`/`090` migrations). Add the `cta_links_stale` reason to `page-rerender`'s `check_rerender_mode` conditional and the `page-content-writer` guidance tweak (1b).

---

## Workstream 3 — vonc.com repair (data; after 091, before/independent of Arena)

1. **093 — CTA retargets** (090 pattern: backups + content_data updates + verification):
   - 19 misdirected CTAs on 10 pages: hero `cta_url` + call-to-action `primary_cta_url` → `/tools/gauntlet/index.html` where copy names the Gauntlet; keep quiz secondaries.
   - `/about.html`: `content-block-about.cta_url` `/how-it-works` → real page (likely `/archetypes.html` given copy); `platform-comparison.cta_url` `/how-it-works/the-gauntlet` → `/tools/gauntlet/index.html`.
   - "Enter the Arena"/"Enter today's Arena" CTAs: → `/tools/arena/index.html` once WS4 lands (or Gauntlet interim if Arena slips).
2. **Nav**: add Gauntlet + Quiz (+ Arena later) to `site_nav_items` — prefer dispatching `orphan_pages`' `nav_drift` → `nav-updater` path; if the enabled `orphan_pages` check doesn't flag the `in_header=true` tools on the verification discovery run, fix its query as part of WS1 instead of hand-inserting.
3. Rerender affected pages (`page_rerender`, `reason: cta_links_stale`), then `curl` each.
4. Do NOT touch `provocation-card`/`lobby-grid` blank fields (runtime-fill shells).

---

## Workstream 4 — Arena (new tool page on vonc)

Platform constraint: no backend/user accounts exist — Arena v1 is a **client-side tool page like Gauntlet** (the only pattern the platform supports). Community features (rooms, duels, live votes) stay v3.

1. Add `site_plan_pages` row (`tool-arena`, role `tool`, current vonc plan) — 088-style SQL.
2. Generate `arena-interface` via the `tool-generator`/`create_tool_component` path (or an 083/084-style manual dispatch script). v1 spec, drawn from the concept docs (`002e_concept_spark(6).md` Arena Mechanics): today's provocation display, take-filing UI (localStorage), the five Arena Reactions (Genius/Delusional/Suspicious/Based/Cursed) on sample takes, a simple remix-chain visual. Reuse `archetype-result-card` pairing if natural.
3. TP-002: manually dispatch the render/deploy work item; TL-001: record in running notes that this page must never receive a generic full rebuild.
4. Retarget the 2 Arena CTAs (WS3.1) and add nav entry.
5. Owner eyeball of the page before calling it done.

---

## Sequencing

1. WS1 Go changes → chassis image (bump IMAGE_TAG, verify symbols in pod) — branch `085_debug_and_feature_loops`.
2. WS2 migrations (091, 092).
3. WS3 vonc repair (093 + rerenders).
4. WS4 Arena.
5. Verification (below). Docs: update vonc workstream running notes + memory per convention.

## Verification

- **Unit**: `go test ./platform/orchestration/...` for new helpers/checks.
- **Fleet safety**: after 091, confirm an untouched business site (ai-agent-orchestration `/about.html`) still serves its persisted CTAs with unchanged DB rows (the 090 verification pattern).
- **vonc end-to-end**: trigger `075_trigger_discovery.sh vonc.com completeness` **before** WS3 → expect `misdirected_cta` (19), `phantom_internal_links` (2), `cta_names_unknown_destination` (Arena), `nav_drift` (2 tools) findings — this proves the generic detection on a real broken site. Then run WS3/WS4, re-trigger discovery → expect **zero** of those findings (the generic loop closed). Park `needs_page:provocation` if reconcile re-emits it.
- **Browser/curl**: every previously-broken CTA on vonc resolves 200 to the page its copy names; Gauntlet/Quiz/Arena present in rendered nav; Arena page interactive (JS present, tool-doc header, not a prose page).
- **Regression guard**: the new checks run clean (zero findings) on robot-hands.com — after its own CTA retarget items are reviewed — proving no false-positive storm fleet-wide.
