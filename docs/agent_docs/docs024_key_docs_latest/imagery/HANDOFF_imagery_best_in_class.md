# HANDOFF — Imagery best-in-class workstream (start a new chat from here)

**Last updated: 2026-07-14 (Turn 36). UPDATE THIS DOCUMENT EVERY WORKING
TURN, alongside the running notes — it is the single entry point for a fresh
session.**

## READ FIRST (Turn 38): I2.2 + I2.3 are DONE AND CORRECT ON THE LIVE SITE
Sprite bullets render **four distinct glyphs** (ⓘ info / ✓ check / gauge / ⚠
warning) on a real list — `/guides/tool-grip-force-friction-calculator-guide.html`,
the "Safety Factor Selection Guidance" list. The Turn 36 CSS-specificity bug is
fixed, deployed (verified by grepping the running POD's binary, not git), and
sprites.css re-emitted with scoped overrides. One 75,745B sheet; ≤80KB budget met.
**Hard-refresh before eyeballing — sprites.css is cached `max-age=3600`.**

**Open:** the user's live gate on those four glyphs.
**Built, awaiting the next deploy:** I2.4 (`sprite_css_missing` check + the
`emit_sprite_css` fulfilment stamp). Its registration SQL is ALREADY applied — an
unregistered check name is just a warn+skip, so it activates by itself on deploy.
Expect exactly ONE self-healing re-emit then (the live CSS predates the stamp, so
the row has `sprites_css = null` → check fires once → identical CSS re-committed →
row stamped → quiet). That cycle IS the live proof of I2.4 — watch for it.
**Then:** I2.5 (D10 container opt-in) closes I2.

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
3. `RUNNING_NOTES_imagery_best_in_class.md` — Turns 1–34, every diagnosis
   with evidence. Append a turn each session.
4. `RUNBOOK_imagery_best_in_class.md` — the human's task queue (A-rituals,
   B-items). Done: B1–B3, B6, B7, B8, B10, B11. Open: B4 (data-source key, at
   I4), B5 (budget sign-off), B9 (reaper cadence — the one that keeps biting).
   The next deploy carries the I2.2/I2.3 Go (emit_sprite_css + head link).
5. `SCOPE_I2_sprite_sheets.md` — current phase's implementation scope
   (png→jpg revised).
6. `SHOWCASE_*.md` (3 files) — shareable summaries (one-pager / narrative /
   technical w/ diagrams). Refresh stats before reuse.
7. `SQL_2026-07-*.sql` — every migration/seed run, each with backup+verify.

## State of the world (2026-07-13, Turn 34)

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

**Phase I2 (sprite-sheet bullets): ⏳ IN PROGRESS — code built, ONE deploy +
finish sequence away from live. This is the active phase.**
- Decisions locked: first surface = LIST BULLETS; 3×3 vocabulary = check,
  gauge, gripper, cog, chart, download, arrow, info, warning.
- Delivery (verified twice): separate committed `/assets/css/sprites.css` +
  head `<link>` — css_snippets is a GLOBAL library (no site_id), and the
  per-site committed bundle is the house pattern (cf. /assets/js/snippets.js).
- I2.0 ✅ (live): chk_kind + validImageryKinds include `sprite_sheet`;
  adapter routes it → Banana; ImagePurposes `sprite_sheet` = **768×768 JPG
  q88** (revised from PNG — see the three-bug stack below); insert-gate passed.
- I2.1 ✅ DONE + GATE PASSED (Turn 34): the sheet is LIVE at
  `/assets/images/sprite-sheet-main.jpg` (768×768 JPEG, 75,745 bytes, under
  the 80KB budget), the glyphs are perfect, and the USER CONFIRMED the cell
  map at the B11 gate. `style_hints.cell_names_verified=true` written into the
  plan row. **Getting a clean deploy took a THREE-BUG STACK, all now fixed —
  the generation itself was flawless first try (cell-alignment risk never
  materialised):**
    1. purpose→hero via ExtractActionInputs' aggressive recursive search →
       fixed workflow-only (`SQL_2026-07-12_asset_deployer_explicit_paths.sql`,
       explicit Strategy-0 `input_data.*` paths on deploy_asset). LIVE.
    2. re-drive left attempt_count capped (3/3 → item excluded from
       find_dispatchable_site → sits triaged; looked like dead dispatch, was
       correctly idle) + a state-machine completion race re-stamped a reset
       item. Fixed by resetting attempt_count=0.
    3. lossless 768² PNG exceeded the Kafka git-commit message-size limit
       ("Message Size Too Large") AND the ≤80KB budget → switched to JPG q88
       (Go, live). SCOPE_I2 revised png→jpg.
- I2.2 ✅ LIVE (deployed v1.0.1114; sprites.css serves 200, 1,711B. NOTE: the
  per-item override selectors it emits are specificity-broken until the next
  deploy carries the Turn 36 fix — see READ FIRST): `emit_sprite_css` — pure CSS
  background-position slicing from the verified grid at bullet size (T=20px →
  sheet drawn 60×60, cells at reading-order offsets). Emits `.sprite` base +
  `.sprite-<name>` (inline/icon/nav) + themed `ul.sprite-list li::before`
  bullets (default glyph = cell 0 = check) + per-item `li.sprite-b-<name>`
  overrides. Commits `/assets/css/sprites.css` (base64 via sendGitCommitRequest).
  GUARDED on cell_names_verified=true. Geometry unit-tested.
- I2.3 ✅ LIVE (head `<link>` verified on served HTML; one real list wired —
  Turn 36): `injectBrandHeadTags` now adds
  `<link rel=stylesheet href=/assets/css/sprites.css>`, GUARDED on an active
  sprite_sheet asset existing. asset-deployer gained a `sprite_css` mode
  (migration `SQL_2026-07-13_asset_deployer_sprite_css_mode.sql`, LIVE):
  check_mode → check_sprite_mode → emit_sprite_css_step. Reusable fleet-wide
  via a `needs_sprite_css` work item (spec.mode='sprite_css').
- **FINISH SEQUENCE (after the next chassis deploy):** dispatch a
  `needs_sprite_css` item (handler asset-deployer, spec {mode:'sprite_css'},
  status triaged, attempt_count 0) → sprites.css commits → rerender-pages with
  refresh_site_components=true (head gets the link) → wire
  `class="sprite-list"` onto ONE robot-hands section's `<ul>` (section-editor
  or a targeted content edit) → LIVE GATE (bullets readable at ~20px, one
  ≤80KB download). Then I2.4: fulfilment discovery check (sheet planned but no
  asset → needs_imagery; asset but no sprites.css → needs_sprite_css); later
  per-section sheets / vision auto-verify.
- Latent gaps recorded (parked): asset-deployer's explicit deploy paths cover
  `input_data.*` (image-build-handler shape) but NOT `input_data.spec.*`
  (dispatch shape) — fix if the undeployed_asset flow misbehaves; historical
  spawned deploys may have silently used hero dims (check the May icons'
  file dims someday); a stale 900² `sprite-sheet-main.jpg` was overwritten by
  the correct one (no longer an issue).

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
- **Page assembly reads `page_components.rendered_html` DIRECTLY** (Turn 36,
  `rerender_single_page_action.go:383` getPageSections) — it does NOT re-render
  from `content_data` + template. To change a section's markup you must write
  `rendered_html` (the artifact that deploys); write `content_data` too so
  source and artifact stay consistent. Re-render by inserting a `page_rerender`
  work item shaped exactly like CreateRerenderItemsAction
  (`create_rerender_items_action.go:192`: handler 'page-rerender', priority 80,
  spec {page_id,page_name,filename,domain}; filename = url minus leading '/').
  Omit `spec.reason` ⇒ unscoped ⇒ plain assemble (no section regeneration).
- **A CSS emitter needs a SPECIFICITY assertion, not just a geometry one**
  (Turn 36): emit_sprite_css's offsets were all correct and unit-tested, yet
  every bullet rendered the default glyph because the override selector was less
  specific than the default rule. Geometry tests can't see a cascade loss —
  only a live render (or a specificity test) can.
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
- **git-commit path has a Kafka message-size limit** (Turn 33): images are
  base64'd into a Kafka message. A lossless PNG of a detailed image can
  exceed the broker max ("Message Size Too Large") — prefer JPG for anything
  visually dense, which also serves the ≤80KB budgets. Text files (CSS/JS)
  commit fine as base64.
- **Committing a per-site TEXT file** (sprites.css, snippets.js): action
  returns/commits a `files` map `{path: {content: base64(text),
  encoding: "base64"}}` via `sendGitCommitRequest`; head `<link>`/`<script>`
  is injected by render_site_components. No storage client needed (DB + Kafka
  producer only).
- **B11-style human gates on generated imagery:** agents can Read PNG/JPG
  files directly — DOWNLOAD the asset (curl its assets.url, which is the live
  presigned S3 URL) and inspect it yourself BEFORE asking the user, so the
  gate question is precise. Give the user the deployed web URL to open.

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

## Next actions, in order (updated Turn 38)
I2.2 + I2.3 are DONE and live-correct (four distinct glyphs on the guide page).
I2.4 is BUILT and its SQL is applied; it activates on the next deploy. Remaining:
1. **USER LIVE GATE** on the four glyphs:
   `/guides/tool-grip-force-friction-calculator-guide.html` → "Safety Factor
   Selection Guidance". Hard-refresh (CSS is cached 1h). Criteria: readable at
   ~20px, one ≤80KB download, on-brand. (D9 already accepted the faint charcoal
   tile behind each glyph — that is expected, not a defect.)
2. **On the next deploy — WATCH I2.4 prove itself** (don't assume): the
   `sprite_css_missing` check should fire ONCE on robot-hands (the plan row has
   no `sprites_css` stamp yet), emit `needs_sprite_css`, asset-deployer re-commits
   an identical sprites.css, and the row gets stamped — then it must go SILENT on
   subsequent passes. If it keeps emitting every pass, the stamp write is broken
   (that is the idempotence bug the tests guard against). Verify:
   `SELECT style_hints->'sprites_css' FROM site_plan_imagery …` is non-null.
3. **I2.5 — container opt-in (D10, user-approved).** Make sprite bullets a house
   style instead of a per-list opt-in: `emit_sprite_css` also emits
   `.sprite-bullets ul>li::before` (same geometry; keep it specificity-safe — the
   Turn 36 lesson), and the `sprite-bullets` class goes on a **component wrapper**,
   starting with article-body's `.article-body__content`. Then every content list
   themes itself. Re-emit CSS + re-render; the Turn 36 hand-wired classes become
   redundant (harmless). NOTE: `article-body` is a GLOBAL component (other sites
   use it) — the class is inert on sites with no sprites.css, so it is safe
   fleet-wide, but say so in the change. Closes I2.
   **WHY (the durability finding):** `ul.sprite-list` needs the class ON the
   `<ul>`, but article bodies are LLM-generated HTML dropped into `{{.content}}`
   — **content `<ul>`s never carry classes.** The Turn 36 wiring is a targeted
   edit to `rendered_html` + `content_data.result` that a content regeneration
   would wipe. robot-hands' only template-owned `<ul>`s are `pd-features`
   (renders EMPTY — see the product-page handoff below) and a JS-filled
   `formula-steps`. A container opt-in is regen-proof and fleet-reusable.
4. **Then Phase I3** (content-linked card imagery / Lane B): generic
   `entity_type`+`entity_id` columns on `assets` (user-confirmed D2), a
   `needs_content_image` work item + handler composing prompts from the
   content item + style guide, card components resolving the entity's image.

## Spun out of this workstream (Turn 36) — not ours to fix
While choosing a list to wire, the `/entities/gripper-detail.html` product page
turned out to be a **hollow shell**: planned, live, `deployed` — and every product
value (name, price, SKU, features) empty, behind full Add-to-Cart furniture on a
site that sells nothing. Worse, the `empty_section` check DID fire and
`page-build-handler` marked the items **`complete`** while the sections stayed
empty (a fix loop closing without fixing). Written up in full, with evidence and
a plan, at **`docs024_key_docs_latest/HANDOFF_2026-07-14_empty_product_sections.md`**
— the user will start that fix in a separate chat. Do not chase it here.
