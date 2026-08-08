# PLAN — 117: the chrome staleness reference points at the wrong table

**Started** 2026-08-07. **Lane owner:** this thread (session `521bfaa9`).
**Bug:** `bugs_open/117_HANDOFF_2026-07-27_site_chrome_is_a_stored_artefact_no_page_rerender_regenerates.md`

## Why this bug, and why now

Picked from `bugs_open/` as the next case with **no active thread and no owning
workstream**. Checked three ways on 2026-08-07:

- `scripts/who-owns.py 117` → no owning workstream; last commit touching the bug
  file was `db14421e7`, 2026-07-27.
- Grep of all 27 session transcripts modified in the previous 5 hours: only this
  session mentions `bugs_open/117`. Four other sessions mention `site_components`
  incidentally; **none** mentions `check_integrity`, `StaleSiteComponents`,
  `stale_site_components`, `deactivated_site_components` or
  `render_site_components_action`.
- Re-checked 30 minutes later before starting work — same answer.

## What the bug says, and what I found instead

117 as filed is a **coupling gap**: chrome is pre-rendered into
`site_components.rendered_html` and served verbatim, and nothing causes a chrome
rebuild when the thing it renders from changes. That framing is still correct.

Its fix candidate 2 ("stamp provenance and detect drift") is marked
`[UNMEASURED] fleet-wide — run it before designing anything`. I ran it, and the
answer changed the shape of the fix:

**A drift detector already exists, is live, is firing — and compares the wrong
two timestamps.** `StaleSiteComponentsCheck`
(`platform/orchestration/actions/discovery_checks/check_integrity.go:306-375`,
check name `stale_site_components`) compares `site_components.updated_at`
against `MAX(page_components.updated_at)` for the site, threshold 24h. Chrome is
not rendered from `page_components`. The reference point is independent of the
subject, so the check is wrong in **both** directions at once.

So this is not "build a detector that does not exist". It is "**an existing,
draining detector answers a different question from the one its name and its
work items claim**" — which is worse, because its output is trusted.

## Decisions

**D1 — the deliverable is a correct staleness REFERENCE, not a new queue.**
The detect→rebuild path already works end to end (`needs_rerender` →
`rerender-pages`, 7 complete per slot, most recent 2026-08-06). Adding a second
producer or a second item_type would hit the `UNIQUE (site_id, item_key)` dedup
trap that `check_integrity.go`'s own `deactivated_pin_` comment documents, and
would reproduce the shape of the open `bugs_open/213`. Fix the predicate; keep
the pipe.

**D2 — a wider timestamp is NOT a better signal.** [MEASURED, see NOTES]
`GREATEST(content_components.updated_at, site_nav_items.updated_at,
sites.updated_at)` marks essentially every row stale, because `sites.updated_at`
churns for reasons unrelated to chrome. Rejected before it was written up.

**D3 — timestamps cannot be the whole answer, because one live writer does not
set one.** `fixTemplateColors`
(`platform/orchestration/actions/fix_harcoded_colours_action.go:180`) does
`UPDATE content_components SET html_template = $1 WHERE id = $2` — no
`updated_at` — and its selection query (same file, ~:145-160) explicitly targets
chrome via `EXISTS (SELECT 1 FROM site_components sc WHERE sc.site_id=$1 AND
sc.component_id=cc.id)`. A chrome template edit by that writer is invisible to
every timestamp-based detector, including a corrected one.

**D4 — therefore the primary mechanism should answer "would a re-render change
anything?", not "is one timestamp older than another".** That points at a
render-input fingerprint stamped on the `site_components` row at render time and
recomputed by the check. It is immune to (a) writers that forget `updated_at`,
(b) timestamp bumps that change no output, and (c) unrelated churn.
**NOT YET DESIGNED IN DETAIL — this is where the work stopped.** See the handoff.

**D5 — this is council-gate scope, not RFC scope [ASSUMED, wants a second
opinion].** Under the owner ruling of 2026-07-29 (§1), an RFC is needed when a
change alters what a *shared mechanism guarantees*. Correcting a check's
predicate changes which rows it emits, not the contract of the work-item seam.
But a new provenance column on `site_components` is a shared-schema addition and
may pull it over the line. Decide before submitting.

> **RESOLVED 2026-08-08: council gate, not RFC.** The 2026-07-29 ruling §1
> narrows architecture scope to changes that alter what a shared mechanism
> *guarantees*; a nullable provenance column read by nothing but the check that
> stamps it is additive-and-inert. The 2026-08-02 ruling §1 covers the item-key
> convergence half (one producer, key shape stated in the register entry). The
> consumers that must be *told* (2026-07-29 §3) are named in D7.6 below.

**D6 — the real render inputs, per slot, with file:line (the question the fable
agent was cut off answering).** All three slots render from ONE `RenderContext`
built once per site in `RenderSiteComponentsAction`; the slots differ only in
*which template* renders it (and one head-only injection). The inputs, verified
2026-08-08 by reading the whole of `render_site_components_action.go`:

| input | tables | where consumed |
|---|---|---|
| site identity (domain, name, company_name, tagline, email, phone, logo_text, logo_url) | `sites` | `loadSiteDataFull` `render_site_components_action.go:349-431`; also `injectBrandHeadTags` :439 (head) |
| palette + theme CSS | `style_collections.color_palette`, `css_themes.css_content` via `sites.style_collection_id` | same query :356-371 |
| plan logo (overrides `sites.logo_url`) | `site_plan_imagery` × `site_plans(is_current)` × `assets(status='active')` | :409-428 |
| nav (header primary; footer primary+utility+legal; legal links; quick links) | `site_nav_items` × `site_nav_groups`, filtered by fetchability (`pages` via `ChromeLinkPolicy`) | `GetNavItems` calls :103-104, :119, :202; `nav_tables.go:324-395`, policy filter `nav_tables.go:175-238` |
| services column | `pages` direct (status, NeverDeployed, name filters, in_header/in_footer, LIMIT 6) | `buildServicesHTML` :1161-1204 (footer) |
| header CTA | contact nav item, else fallback ranking over interactive pages/hubs, gated by `ChromeLinkPolicy` | :148-195 |
| copyright year | `time.Now().Year()` | :107-108 |
| component template + input_schema | `content_components` via `site_components.component_id` (or library default via `ResolveChromeComponent`) | :705-770 |
| schema-fill values | **`site_specs`** (`site_specs.*` and `config.*` sources — `resolveConfigPath` hard-codes aspects `site_config`, `identity`, `design_intent`, `plan_sections_action.go:691-702`), **`assets`** (`site_assets.*`), **`pages`** (`pages.*`) | :782-857 via `sourceResolver` |
| sprite CSS link | `assets` (purpose='sprite_sheet', active) count | :913-923 (head only) |

[MEASURED 2026-08-08] the schema-fill gap is NOT marginal: live chrome
components declare `config.analytics.gtm_container_id` (nearly every head/
header — the GTM estate), `config.color_scheme.*`, `config.chrome.
compliance_lines`, `site_specs.identity.company_name`, `site_assets.logo`,
`pages.contact`. A fingerprint that cannot see `site_specs` misses live,
high-value drift. All of these resolve from three stores: `site_specs`,
`assets`, `pages` — all SQL-reachable.

**D7 — the mechanism (question 2 answered).** A render-inputs fingerprint,
stamped and recomputed by ONE shared SQL expression.

1. **Where it lives:** new nullable column `site_components.render_inputs jsonb`
   (migration 334). NOT a `content_data` key: `content_data_envelope_guard.go:115`
   documents "site_components has NO automated content_data writer at all" as a
   structural property bug 190's guard relies on — a stamp there would silently
   invalidate another mechanism's stated invariant.
2. **What is hashed:** a jsonb of NAMED per-input digests — `{v, year, template,
   identity, style, specs, nav, services, brand_assets, plan_logo}` — computed
   entirely in SQL from the live rows in D6's table. `specs` scopes to the three
   aspects `resolveConfigPath` itself hard-codes (pinned by test). Whole-object
   jsonb comparison (`IS DISTINCT FROM`) detects drift; `v` bumps force a
   fleet restamp deliberately.
3. **Why one SQL expression:** the checker lives in `discovery_checks`, which
   cannot import `actions` (dependency runs the other way — asserted by
   `site_component_lock_guard_test.go:51`), so it can never call `GetNavItems`.
   Two hand-maintained copies are the drift class this estate keeps paying for.
   The fragment lives in `datahelpers` (both packages already import it) —
   exactly the move `content_data_envelope_guard.go:119` prophesies ("would move
   to datahelpers unchanged the day a caller outside this package needs it").
   Both sides run the SAME string, so stamp and recompute agree by construction.
4. **Stamping:** in the SAME guarded UPDATE that stores `rendered_html`
   (`render_site_components_action.go:928`) — atomic, no window, lock-guarded.
5. **Absent stamp = stale (policy b), deliberately.** All 51 unlocked rows are
   unstamped at rollout → each of the 19 sites fires ONE `stale_chrome` item on
   its first post-roll discovery pass → one-time bounded baseline drain that
   also heals oufe.com/footer (the known false negative) and the 3
   `component_id IS NULL` rows (which also fire via an explicit clause). The
   rejected alternative — stamping current state as baseline — would declare
   oufe's known-stale footer fresh. The drain REPLACES a detector that today
   fires 33 false positives *continuously*.
6. **Consumers told (2026-07-29 §3):** `rerender-pages` (handler — contract
   unchanged: same item_type `needs_rerender`, same spec key
   `refresh_site_components: true`, confirmed satisfiable: its
   `check_refresh_components` step force-rerenders all three slots);
   `site_components` schema readers (new nullable column, no `SELECT *` scan
   breakage found); bug 190's census comment (untouched — that is WHY the
   column, see 1); the old `stale_sc_<slot>` item keys (retired: one site-level
   `stale_chrome` key replaces three per-slot keys — 2026-08-02 ruling §1
   satisfied by naming producer + key shape in the register entry; open old-key
   items drain normally, no collision).
7. **Locked rows are skipped** via the shared writable predicate
   (`pageComponentAgentWritableSQL` delegating to a new
   `datahelpers.AgentWritableSQLFor`) — 6 of 57 rows today. A locked stale slot
   is deliberately unmonitored by this check: the render path cannot satisfy the
   item (069 owns the surface), so firing would churn items to `unresolved`.
8. **Declared coverage boundaries, each measured:** pages-fallback nav (0 of 19
   chrome sites lack nav rows today); CTA fallback candidate ranking (the
   harmful direction — a dead CTA — is covered by 191's `markStaleChromeLinkSlot`);
   `site_specs` aspects beyond the resolver's three (only `identity` observed
   in live chrome sources, and it IS covered); `site_assets.*` beyond
   logo/sprite purposes. `year` is included on purpose: stale copyright is real
   staleness; cost is one fleet restamp every 1 January.

**D8 — candidate predicate run against live data before proposing (the R4/misstep-2
rule).** [MEASURED 2026-08-08, scratchpad `fingerprint_candidate.sql`]: two runs
identical (deterministic — every aggregate carries ORDER BY); per-key variance
across 57 rows: identity 19/19 sites distinct, specs 19, style 16, nav 18,
services 17, template 10 (library components shared across sites — correct),
brand_assets 3 (few sites hold logo/sprite assets); template digest NULL only on
the 3 `component_id IS NULL` rows. Neither ~0% nor ~100%, and the digests vary
where the underlying data varies.

## Phasing

1. ~~Confirm the bug is still live~~ — DONE, see NOTES.
2. ~~Measure the detector against the real signal~~ — DONE, cross-tab in NOTES.
3. ~~File the 090 diagnosis~~ — DONE, ran to completion; verdict not retrievable.
4. ~~Design the fingerprint~~ — DONE 2026-08-08, D6–D8 above.
5. **Fate of the existing check (question 3): the predicate is REPLACED in
   place** — same registered name `stale_site_components`, same item_type
   `needs_rerender`, same handler, same `refresh_site_components: true` spec
   key; item_key consolidates `stale_sc_<slot>` → one site-level `stale_chrome`
   (one firing = one site rebuild instead of up to three). **Induction test that
   it can still fire:** the rollout itself is the induction — 19 sites carry no
   stamp, so the first post-roll discovery pass MUST fire 19 items, then go
   quiet as rebuilds restamp (goes-quiet proves convergence; a check that stays
   loud or was never loud is broken in one of the two opposite ways). Second-
   order induction available any time after: change any covered input on a
   canary (e.g. a utility nav item), expect exactly one item.
6. Council gate, then commit with `Council-Submitted:`, then migration 334
   (scoped apply), then build + roll, then pod-verify with positive AND negative
   grep controls (positive: `render_inputs` in the chassis binary; negative: the
   retired `stale_sc_` item-key literal, expect 0).

## Corrections to this plan

None yet.
