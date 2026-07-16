# RUNBOOK — fleet link/flow integrity + vonc repair + Arena (2026-07-14)

> **STATUS 2026-07-16: EXECUTED. Migrations 091–097b applied live, Arena
> deployed, all verified. Current handoff / summary / next steps:
> `HANDOFF_link_integrity_arena_2026-07-16.md`. This RUNBOOK is now the
> historical operational sequence; read the HANDOFF first.**

Operational sequence for the work shipped on branch `085_debug_and_feature_loops`
(plan: owner-approved 2026-07-14). Everything below runs against the live
cluster/DB — none of it happened in the authoring session. **Order is
load-bearing**; each step's gate is stated.

Companion migrations (this directory): `091` … `097b`.
Go changes (one chassis image): `resolve_internal_links_action.go` (interactive-
first `chooseCTATargets`, `cta_target_title`, `ctaExcludedDestination`),
`rerender_page_sections_action.go` (`applyCTARecompute`, gated on
`reason == "cta_links_stale"`), `create_rerender_items_action.go` (reason
stamp), `render_site_components_action.go` (header CTA validation + fallback),
`discovery_checks/check_misdirected_cta.go` + `check_incomplete_page_group.go`
(new), `check_orphan_pages.go` (nav-flagged tool pages now considered),
`datahelpers/links.go` (`ExtractAnchors`).

## 1. Chassis image

1. Bump `IMAGE_TAG` (makefile line 16, currently `v1.0.1114` → next).
2. Build + roll the pod. Image BEFORE seeds; no orchestration fires within
   ~5 min of pod start.
3. Verify symbols in the running binary (verify by artifact, not deploy logs):
   ```
   kubectl -n ai-persona-system exec <pod> -- sh -c \
     'for s in loadInteractivePages applyCTARecompute misdirected_cta incomplete_page_group ExtractAnchors; do
        echo -n "$s: "; grep -ac $s /proc/1/exe; done'
   ```
   Every count must be ≥ 1.

## 2. Migrations 091 + 092 (fleet)

Gate: image live (091 is safe earlier — it only bites on next render — but
keep it simple: image first, then SQL).

1. `091_cta_url_schema_source_flip.sql` — hard-fails if 0 fields flip or any
   `pages.*` CTA source remains.
2. `092_enable_link_checks_and_cta_rerender.sql` — enables
   `phantom_internal_links` / `misdirected_cta` / `incomplete_page_group` on
   completeness-discovery-agent, adds `cta_links_stale` to page-rerender's
   `check_rerender_mode`, appends the `*_target_title` writer guidance.
    * 1b follow-up check: after the next content-writer run on any site,
      confirm the guidance text actually reached the section prompt (the known
      "prompt seams dropping spec intent" trap) — look for `_target_title` in
      the writer's prompt log.

**Fleet safety gate** (the 090 pattern) before touching vonc: on an untouched
business site (ai-agent-orchestration), `curl /about.html` — persisted CTAs
unchanged, DB rows unchanged. Then dispatch one plain `image_landed` rerender
anywhere and diff CTA hrefs before/after: must be byte-identical (proves the
1c gate).

## 3. vonc discovery — BEFORE repair (proves generic detection)

```
scripts/initial_messages/170_work_item_flow_build/075_trigger_discovery.sh vonc.com completeness
```
Expected findings on the broken site (investigation figures, 2026-07-14):
- `misdirected_cta`: ~10 page items covering the 19 `/contact.html` CTAs
- `phantom_internal_links`: 2 (`/how-it-works`, `/how-it-works/the-gauntlet`)
- `cta_names_unknown_destination`: the Arena CTAs (circular/homepage)
- `nav_drift`: 1 site item covering the 2 nav-flagged tool pages

If any class is missing, STOP and diagnose before repairing — the point is
that the generic loop catches all of it.

Park rule: if any pass re-emits `needs_page:provocation`, set it back to
`'detected'` (deliberate — Spark pipeline's page).

## 4. vonc repair

1. `093_vonc_prose_link_404_fixes.sql` (the two /how-it-works phantoms; the
   19 CTA misdirections are deliberately NOT hand-fixed — see file header).
2. Dispatch the `nav_drift` item → nav-updater (adds Gauntlet + Quiz to
   `site_nav_items`, re-renders site components — the new header-CTA
   validation also runs here).
3. Promote + dispatch the `misdirected_cta` → `page_rerender` items
   (status `detected` → `triaged`, then a dispatch pass, e.g.
   `087_dispatch_work_items_vonc.sh`):
   ```sql
   UPDATE site_work_items SET status='triaged'
   WHERE site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
     AND item_key LIKE 'misdirected_cta:%' AND status='detected';
   ```
   The `cta_links_stale` rerender recomputes hero/call-to-action targets:
   primary → `/tools/gauntlet/index.html`, secondary → `/tools/quiz/index.html`
   (interim Gauntlet for the Arena CTAs). Also rerender `/about.html` if it is
   not in the misdirected set, so 093's content_data lands in HTML.
4. `curl` every previously-broken CTA target → 200, and each lands on the
   page its copy names. Do NOT touch `provocation-card` / `lobby-grid` blank
   fields (runtime-fill shells).

## 5. Arena (WS4)

1. `094_vonc_arena_tool.sql` — plan row + `add_tool` item + PLAN doc. The
   `incomplete_page_group` finding for `tool-arena` that appears at the next
   discovery pass is EXPECTED (live proof of the check); it clears when the
   page deploys.
2. Dispatch the `add_tool_novel:tool-arena-interface` item → tool-generator.
3. **TP-002**: manually dispatch the render/deploy for the new page (the
   generator does not enqueue it). 083/084-style scripts in
   `scripts/initial_messages/210_vonc_trigger/` are the template.
4. **TP-004**: if anything routes `tool-arena` to page-build-handler, kill it.
5. `095_vonc_arena_cta_retarget.sql` (pre-flight enforces the page is
   deployed), then a `cta_links_stale` rerender for the affected pages.
6. **TL-001** (standing): `/tools/arena/index.html` must never receive a
   generic full rebuild — section-editor targeted path only.
7. Owner eyeball of the page before calling it done.

## 6. Close the loop

1. Re-trigger vonc completeness discovery → expect ZERO of: misdirected_cta,
   phantom_internal_links, cta_names_unknown_destination, nav_drift,
   incomplete_page_group. (Park `needs_page:provocation` if re-emitted.)
2. Browser: Gauntlet/Quiz/Arena in rendered nav; Arena interactive (JS
   present, tool-doc sentinel header in the component, not a prose page).
3. Regression guard: run the same discovery on robot-hands.com — after its
   own CTA retarget items are reviewed, the new checks must run clean (no
   false-positive storm fleet-wide).
4. Unit tests (already green in the authoring session):
   `go test ./platform/orchestration/...` — pre-existing failure in
   `platform/orchestration/orchestration_test.go` (`NewSagaCoordinator`
   signature drift) is unrelated and predates this work.
