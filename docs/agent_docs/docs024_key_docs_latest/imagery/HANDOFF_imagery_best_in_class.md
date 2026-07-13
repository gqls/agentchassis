# HANDOFF — Imagery best-in-class workstream (start a new chat from here)

**Last updated: 2026-07-12 (Turn 30). UPDATE THIS DOCUMENT EVERY WORKING
TURN, alongside the running notes — it is the single entry point for a fresh
session.**

## What this project is (fresh-reader paragraph)

agentchassis is an autonomous agent platform that plans, builds, and operates
a fleet of content websites (tools, guides, games, news feeds, articles).
One generic Go runtime executes declarative workflows stored as JSONB in
Postgres (`agent_definitions`); agents cooperate over Kafka; all work flows
through `site_work_items` (discovery checks find problems → dispatcher →
handler agents → git-backed deploy). The **imagery best-in-class workstream**
(started 2026-07-08) raises fleet visual quality: brand consistency,
data-accurate infographics, card imagery, sprite-sheet bullets, product
sketches, news imagery, performance budgets, audit loop.
**Testbed: robot-hands.com**, site_id `00ff3af5-dad8-4770-9f70-3edc267a3c92`.
Working branch: `084_site_improvements_local_ai` (user makes bulk commits
that sweep in-progress edits — check `git log` before assuming anything is
uncommitted; forward-only, never reset).

## Document map (all in this directory)
1. **THIS FILE** — state + mechanisms + next actions. Start here.
2. `PLAN_imagery_best_in_class.md` — goals G1–G9, user-confirmed decisions
   D1–D8, phases I0–I8 with dated status blocks. The map.
3. `RUNNING_NOTES_imagery_best_in_class.md` — Turns 1–29, every diagnosis
   with evidence. Append a turn each session.
4. `RUNBOOK_imagery_best_in_class.md` — the human's task queue (A-rituals,
   B-items). Open: B4 (data-source key, at I4), B5 (budget sign-off), B9
   (reaper cadence), **B10 (next deploy carries the sprite gating fix)**,
   **B11 (sprite eyeball gate, after B10)**.
5. `SCOPE_I2_sprite_sheets.md` — current phase's implementation scope.
6. `SHOWCASE_*.md` (3 files) — shareable summaries (one-pager / narrative /
   technical w/ diagrams). Refresh stats before reuse.
7. `SQL_2026-07-*.sql` — every migration/seed run, each with backup+verify.

## State of the world (2026-07-12, Turn 30)

**Phase I0 (testbed rebuild + render acceptance): ✅ COMPLETE.**
33-page rebuild w/ live news (9 sources, latest-news on index); 16 distinct
per-page git-path heroes, zero expiring URLs; layout = tool-portal-dark
(B7: classification industry_tags fix + 025-pattern `css_themes.layout_id`
swap — there is NO runtime re-compose path, by design); corrupted-template
class self-healing (bridge check live; 10/14 healed, 4 remaining fix
themselves on other sites' discovery passes).

**Phase I1 (brand consistency): ✅ COMPLETE, LIVE-VERIFIED on served HTML.**
- `imagery_style_guide` site-spec drives every generation; per-kind gating
  (photographic kinds: medium+mood+palette; icon/sprite_sheet: palette only;
  logo: nothing) — PROVEN on real generations (icons carried palette, not
  medium). `avoid` → negative prompt; `reference_asset_keys` → stable s3://
  anchors for Banana.
- Logo: user-approved, LOCKED (`assets.locked_at`, lock_type=permanent);
  store-guard refuses overwrites (D5). Header resolves it from plan imagery
  (`logo-img` live). Favicon + OG card DERIVED from the logo
  (`derive_brand_head_assets`; favicon.png/og-card.png serve 200; og:image +
  twitter:card injected into every head at render time).

**Phase I2 (sprite-sheet bullets): ⏳ IN PROGRESS — the active phase.**
- Decisions locked: first surface = LIST BULLETS; 3×3 vocabulary = check,
  gauge, gripper, cog, chart, download, arrow, info, warning.
- Delivery (verified twice): separate committed `/assets/css/sprites.css` +
  head `<link>` — css_snippets is a GLOBAL library (no site_id), and the
  per-site committed bundle is the house pattern (cf. /assets/js/snippets.js).
- I2.0 ✅: chk_kind + validImageryKinds include `sprite_sheet`; adapter
  routes it → Banana; ImagePurposes 768×768 png; insert-gate passed.
- I2.1 ⏳ REGEN IN FLIGHT (Turn 31). B10 deploy done; gating fix live and
  PROVEN (first sheet's origin_prompt: palette-only, no photographic voice).
  First generation was near-perfect (all 9 glyphs, exact reading order) but
  its DEPLOY committed 900×900 sprite-sheet-main.JPG — hero config — because
  `deploy_image_asset`'s ExtractActionInputs aggressive search matched a
  stale `purpose` (child HAD received purpose='sprite_sheet'; verified via
  initial_request_data). FIXED workflow-only:
  `SQL_2026-07-12_asset_deployer_explicit_paths.sql` adds explicit
  Strategy-0 `input_data.*` paths to deploy_asset. **Standing lesson: give
  ExtractActionInputs actions explicit dot-paths; never trust the search.**
  Item reset → full chain re-running; monitor expects item complete +
  sprite-sheet-main.png 200 + 768×768 PNG; then the B11 eyeball gate
  (user says true cell meanings → write back, cell_names_verified:true).
  Latent gaps recorded: dispatch-shape inputs (`input_data.spec.*`) not
  covered by the explicit paths; historical spawned deploys may have used
  hero dims silently (check May icons' file dims someday); stale
  sprite-sheet-main.jpg clutter in the site repo.
- I2.2 (NEXT BUILDABLE NOW): `sprites.css` emit — pure string compute from
  style_hints grid + fixed 768² dims, committed via the git-adapter (mirror
  the asset-deployer `brand_head` mode / needs_brand_head_assets pattern).
  Rules per cell: `.sprite-<name>{background-image:url(/assets/images/
  sprite-sheet-main.png);background-position:-<c*W>px -<r*H>px;width/height;
  background-size}` + a `li::before` bullet helper. ONLY emit for cells with
  verified names (cell_names_verified:true after the eyeball gate).
- I2.3: head `<link rel="stylesheet" href="/assets/css/sprites.css">`
  (extend `injectBrandHeadTags`, guard on sprite asset existing) + wire
  bullets on ONE robot-hands section. I2.4+: fulfilment discovery check;
  later per-section sheets / vision auto-verify.

**Phases I3–I8: not started.** I3 (card imagery / Lane B: generic
entity_type+entity_id on assets — user-confirmed D2) is next after I2.
I4 charts = go-echarts in-chassis, per-domain free data sources. All
decisions D1–D8 user-confirmed (see PLAN §4/§8).

## Mechanisms a fresh session must know (hard-won)

- **DB access:** `kubectl exec -n ai-persona-system postgres-clients-0 --
  psql -U clients_user -d clients_db`. A PreToolUse hook auto-approves
  read-only SELECT/\d through this exact form; mutations prompt the user.
  Auth expires ~daily → runbook A1 (user re-login), symptoms: "server has
  asked for the client to provide credentials" on EVERY call.
- **Manual agent trigger** (`system.intake` is STALE — do not use): kcat pod
  → `system.agent.generic.requests`, header `action=orchestrate`, body
  `{"headers":{...},"config":{"agent_type":"<type>"},"input_data":{...}}`.
  Full working example: notes Turn 18; script precedent
  `docs/agent_docs/sql_for_agents/033_rerender_pages_trigger.sh`. Known-good
  for: improvement-loop (full discovery cycle), webdesign-agent (CSS
  re-render+deploy), rerender-pages (+refresh_site_components:true).
  **Do NOT hand-roll spawn_agent+call_agent inline workflows** — the spawned
  child runs its workflow on INIT and idles before your call arrives
  (Turn 26). Route work to spawned handlers via WORK ITEMS + dispatch.
- **Dispatch input contract** (handlers invoked by build-dispatch-loop
  receive): `input_data.spec` (the item's spec), `input_data.site_id`,
  `input_data.domain`, `input_data.item_type`. Write step conditions against
  these (e.g. asset-deployer's check_mode: `input_data.spec.mode ==
  "brand_head" OR input_data.mode == "brand_head"`).
- **Manually-inserted work items are NOT auto-triaged** — insert with
  status='triaged', triaged_at=now(). Dedup: partial unique (site_id,
  item_key) over non-terminal statuses. `site_plan_imagery.chk_source`
  allows only llm|classifier|manual|adoption → seeds use 'manual'.
- **RE-DRIVING a work item (Turn 32 lesson):** ALWAYS reset
  `attempt_count=0` alongside `status='triaged'` and clearing claim metadata.
  At `attempt_count>=max_attempts` the item is EXCLUDED from
  find_dispatchable_site → sits triaged forever, and if it's the site's only
  candidate the trigger idles (looks like dead dispatch; it isn't). Also
  beware a just-finished orchestration's tail re-stamping a freshly-reset
  item back to complete (state-machine race) — verify no in-flight
  orchestrations for the item before/after resetting.
- **Zombie claims:** a claimed item stuck >~10 min blocks its ENTIRE site
  from dispatch. Standing unstick:
  `UPDATE site_work_items SET status='triaged', claimed_by=NULL,
  claimed_at=NULL, updated_at=now() WHERE status='claimed' AND claimed_at <
  now() - interval '10 minutes';` Real fix = B9 (user's TODO 6/10/11).
- **Component regeneration:** queue `needs_component_regeneration` →
  component-creator; spec {function, component_id, section_type, ...,
  description}. `description` renders into the creator prompt — use it to
  DEMAND exact schema-field preservation if the pre-store guard rejects.
- **Deploys:** user-driven (git → GitHub Actions → docker → k8s). Go changes
  are inert until then. Post-deploy ritual: clear zombies; re-trigger any
  loop the restart interrupted (improvement-loop items strand in 'detected'
  if triage hadn't run — re-trigger the loop, don't hand-promote dozens).
- **Paths/conventions:** deployed asset path = `storage.DeployedWebPath(
  asset_key, purpose)` (NEVER assets.url — expiring presigned);
  `imageryplan.ImageRoleForPath` = shared image-role aliases; kind→provider
  routing in `dynamic_adapter.go` (flat kinds → Banana, photographic →
  SDXL); per-kind prompt gating in `directionAppliesToKind` +
  `styleGuide.directionForKind` — ANY NEW KIND must be added to BOTH gating
  functions AND chk_kind AND validImageryKinds AND (usually) the adapter
  switch + ImagePurposes. That five-place checklist is the I2.0 lesson.

## Open threads (parked, non-blocking)
- 4 corrupted components remain (archetype-taster-quiz, lobby-grid,
  provocation-card, tool-cta) — self-heal on their sites' next discovery
  passes; forceable via improvement-loop trigger per site.
- Kafka per-job response-topic partition race — transient; now surfaces as
  failed items (mark_item_failed fix) instead of silent successes.
- No runtime re-compose path (layout changes = 025 FK-swap pattern);
  build a site-design-planner re-resolve mode if this becomes routine.
- learning-center-index orphan component row (pre-rebuild residue) — clears
  with I3 card imagery or a listing rebuild.
- Old orphan pages (how-it-works, selection-guide, learning-center sprawl)
  not in the current plan — cleanup pass someday.
- image_source_unsatisfiable check live but has produced 0 flags (heroes all
  resolve now) — expected.

## Next actions, in order
1. **Build I2.2 now** (doesn't need the deploy): sprites.css emit action +
   `needs_sprite_css`-style work-item route (mirror brand_head mode), gated
   on cell_names_verified:true.
2. **On the user's next deploy (B10):** let/queue the needs_imagery for
   `sprite_sheet_main` → generate via Banana → **B11 eyeball gate with the
   user** → write the TRUE cell→name map into style_hints, set
   cell_names_verified:true.
3. I2.2 emit runs → sprites.css committed → I2.3 head link + wire bullets on
   one robot-hands section → live gate (readable at bullet size, one ≤80KB
   download).
4. I2.4 fulfilment discovery check; then Phase I3 (card imagery / Lane B).
