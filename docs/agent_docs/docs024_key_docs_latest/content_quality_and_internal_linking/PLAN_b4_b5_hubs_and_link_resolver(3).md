# PLAN — B4/B5 hub links + internal-link-resolver agent

The remaining internal-linking work after the phantom fixes (Step 1, Layer 1b, Layer 2). Principle throughout: resolve from the real `pages` set, or omit; intent-aware destination choice lives in a dedicated agent.

> **Status (2026-06-11).** Both halves are written and ready for the batch deploy. B4/B5: `section_index_for.go` (verb), `b4_b5_hub_links_schema.sql`, `b4_b5_hub_links_template_gate.sql` — option (c) below. Step 3: `resolve_internal_links_action.go` (incl. in-Go `unresolved_cta` HITL emission mirroring `createDeferredItems`), `internal_link_resolver_agent.sql` (row modelled on `research-agent`), `page_content_writer_link_resolver_wiring.sql` (spawn/call/select/iterate via `snapshot_agent` + chained `jsonb_set`, with an `extract_fields` fallback chain making resolver failure byte-identical to today), and `check_phantom_internal_links.go` rerouted (`page_component` → `page-build-handler`/`content`, mirroring `check_empty_sections` — a rebuild re-runs build-time resolution; `page-rebuild` is a batch agent requiring pre-flagged pages, not a per-finding handler). Deploy per RUNBOOK Part 2; the design rationale below stands as written.

---

## B4/B5 — "Browse All X" hub links (empty href)

### Problem (verified)
The three list components source their Browse-All href from unpopulated/inconsistent specs:
- `tool-list`: `cta_url` ← `site_specs.identity.tools_index_url` (no fallback)
- `game-list_pre_037`: `cta_url` ← `site_specs.identity.games_index_url` (fallback `/games/index.html`)
- `guide-list_pre_037`: `cta_url` ← `site_specs.navigation.guides_index_url` (no fallback)

The `*_index_url` keys are absent from the specs → empty. Cards already resolve correctly via `query.pages_where_type:<type>`. Real hubs exist: `tools-index` `/tools/index.html`, `guides-index`, `games-index`.

Two things to verify (do not assume):
1. **Spec-path resolver missing-leaf behaviour.** `game-list` has fallback `/games/index.html` (a real page) yet rendered `href=""`. Likely `site_specs.identity.games_index_url` reports the `identity` aspect as found with a missing leaf, so the fallback never fires. Confirm in `resolveSpecPath` (sourceResolver): a missing leaf under a present aspect should return `(nil,false)` so `on_missing`/`fallback` governs.
2. **List template gating.** Confirm the list templates render the Browse-All button gated (`{{if .cta_url}}`) or unconditionally (`href="{{.cta_url}}"`). If ungated, gate it (same `{{if .cta_url}}` pattern as the hero/header fix) so an absent hub renders no button rather than `href=""`.

### Options
- (a) Populate the `*_index_url` specs from real hubs. **Rejected** — per-site data to maintain; re-introduces the inconsistent-source brittleness.
- (b) `source: pages.<hub-name>` per component (e.g. `pages.tools-index`). Zero new code (uses the fixed resolver), but bakes the `<area>-index` naming convention into each schema; breaks if a site names its hub differently.
- (c) **Recommended.** A new `queryresolve` verb returning the hub URL, sourced as `query.section_index_for:<type>`. Cards and hub link both flow through `queryresolve`; no naming or spec baked in; derives the hub from real page relationships.

### Design (option c)
New verb in `queryresolve.Resolve` switch: `section_index_for` (arg = type, e.g. `tool`/`guide`/`game`).
- Resolve the hub as the `section-index` page sharing the area with pages of that type:
  - Primary: `SELECT url FROM pages WHERE site_id=$1 AND page_type='section-index' AND site_area_id = (SELECT site_area_id FROM pages WHERE site_id=$1 AND page_type=$2 AND site_area_id IS NOT NULL LIMIT 1) LIMIT 1`.
  - Fallback (site_area_id unset): match by the common URL prefix of the type's pages (e.g. `/tools/`) → the `section-index` page at `/<prefix>/index.html`. Reuse the `pages_under_section` url_prefix fallback approach already in `queryresolve`.
- Return a single value (URL string). Note: existing verbs return lists; this returns one scalar — keep the `interface{}` return and have the resolver/source layer accept a scalar for a `url` field. Confirm how the field-resolution loop consumes a `query.*` result for a non-array field (the card `items` is an array; `cta_url` is a scalar) — this is the one integration point to get right.
- Component schema change (3 components): `cta_url.source` → `query.section_index_for:tool|guide|game`, `on_missing: skip_field`, drop the inconsistent spec sources. Snapshot first.
- Then gate the list template Browse-All button on `cta_url` (per verify #2).

### Result
Browse-All resolves to the real hub or omits — no `href=""`. Consistent across tools/guides/games. Spec-independent.

---

## Step 3 — internal-link-resolver agent

### Why
Step 1/Layer 1b deliberately leave hero/CTA/header destinations as correct-or-absent (no phantom, but the button may be dropped). Restoring an intent-appropriate destination is a distinct responsibility, and internal links appear in **any** component (a guide links to its tools, in-body prose, related blocks), not just hero/nav. This is its own agent (every agent is an orchestrator).

### Responsibility (single)
Given a page's context and the real pages, resolve/validate intended internal link destinations and choose intent-appropriate targets; emit a signal when it genuinely can't.

### Workflow (thin; complexity in Go actions)
- Spawned by the content/build path (page_component link findings from `check_phantom_internal_links` route here; also invocable during the page-content flow).
- Reuse: `prepare_link_context`'s `available_pages` (the valid-target list), `queryresolve` (intent→hub, incl. the new `section_index_for`), the fixed `sourceResolver` `pages` case (ref→URL), `datahelpers.PageURLSet` (validate against real pages).
- For a CTA: choose the destination from page role + real hubs (homepage → primary hub; a guide → its related tool or the guides hub; a tool page → the relevant hub), write the resolved `cta_url`/`secondary_cta_url` into section data. Never emit a URL not in the real pages set.
- Responds on the parent's responses topic (not its own).

### unresolved_cta signal (build-time detectability)
When a section has CTA text but no resolvable destination, emit an `unresolved_cta` finding (page, section, intent text) at resolution time — the deploy gate can't see a correctly-dropped button (no fingerprint in rendered HTML), so this is the only place the absence is detectable before deploy. Routes to the improvement loop / HITL per severity.

### Sequencing with the audit + sweep
Build this agent before re-enabling `improvement-sweep` and before enabling `phantom_internal_links`, so page_component findings have a handler and the sweep clears rather than accumulates.

---

## Order
1. B4/B5 (smaller; finishes write-path phantoms): verify resolveSpecPath + list gating, add `section_index_for` verb, repoint 3 schemas + gate templates, re-render, dry-run.
2. internal-link-resolver agent (Step 3).
3. Enable `phantom_internal_links`; then re-enable `improvement-sweep`.
