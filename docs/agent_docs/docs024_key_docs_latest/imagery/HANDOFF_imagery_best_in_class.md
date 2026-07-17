# HANDOFF — Imagery best-in-class workstream (start a new chat from here)

**Last updated: 2026-07-16. UPDATE THIS DOCUMENT EVERY WORKING TURN, alongside
the running notes — it is the single entry point for a fresh session.**

## WHERE WE ARE (2026-07-16, Turn 46) — start here
- **I0, I1, I2 are ✅ COMPLETE AND LIVE** (incl. B12: served CSS default is the
  ARROW, self-healed on the v1.0.1125 pass). Read-out for a status briefing:
  `READOUT_2026-07-16_imagery_status.md`.
- **I3 mechanism acceptance MET LIVE on v1.0.1125 (Turn 46):** both checks fired
  in one discovery pass; 9 cards derived + entity-linked + committed;
  `query.blog_posts` resolved 9 articles with per-article images; the served
  `learning-center-hub.html` shows all 9 `card-*.jpg` (HTTP 200).
- **D13 (user 2026-07-16): per-article content-hero GENERATION is BUILT, rides
  the next deploy.** All 9 first-run cards were byte-identical (every blog-post
  fell back to `hero_canonical` — planner emits no article heroes). The check is
  now a two-mode emitter (generate content hero via image-build-handler's generic
  needs_imagery path / derive-or-re-derive the card on ORIGIN-STALENESS), and the
  preference order plan-hero → content-hero (`ContentHeroKey`) → site-hero is
  unified across check, deriver, and page renderer. Also riding: card q78 (first
  run hit 64,097B > the ≤60KB budget at q82). **Post-deploy the fleet converges
  itself:** pass 1 generates ~9 article heroes (SDXL — B5 budget note), pass 2
  re-derives the 9 cards from them, pass 3 silent. Then the A3 gate: 9 visually
  DISTINCT cards.
- **Dispatch priority is ASC — lower number = sooner** (`ORDER BY wi.priority ASC`;
  "30 // high" in check comments). The old "needs_page@99" habit meant "run LAST".
  Front-of-queue nudge for a watched run = set priority 5.
- **The image-landing trap is CLOSED and the content-loss thread has RECOVERED
  ALL 17 article-body instances** (004 updated by that thread; root cause was
  writer max_tokens truncation) — D13's image landings are safe fleet-wide.

## ✅ READ FIRST — image-landing trap CLOSED (guard live); residual notes
**History:** landing an image fired a scoped re-render (`image_landed`) that BLANKED
the article body on any page whose content was a never-parsed JSON envelope. The
**escalate-to-writer guard is now LIVE in prod** (re-verified in the running pod this
turn — the 004 handoff's own criterion), so a scoped re-render on such a page now
escalates to the writer instead of blanking. **Residuals to respect:**
- **The testbed (robot-hands.com) is safe** — its article-body pages are healthy and
  every imagery row is fulfilled, so no image landing will even fire. Proceed with I3
  here.
- **Do not assume every OTHER site is safe to land images on yet:** the writer-side
  envelope repair (`ParseLLMJSON`) still fails on some fixtures, so an escalation may
  partially- or not-regenerate. And **4 pages are still blanked** (finetuning×3,
  gamesdesign×1) awaiting recovery in the separate thread.
- **Detection-query fix (found this turn):** the 004/006 `length=1326` test now
  UNDER-reports — I2.5's `sprite-bullets` class made a fresh blank shell 1341 bytes.
  Use `rendered_html ~ 'article-body__content[^>]*></div>'` instead. 004 §5 corrected.
Full write-up: **`../aaa_fails_to_mend/004_HANDOFF_image_landing_blanks_article_body.md`**
Root cause + recovery: **`../HANDOFF_2026-07-14_article_body_json_envelope.md`**

## Turn 38–39: I2.2 + I2.3 are DONE, LIVE AND USER-GATED
Sprite bullets render **four distinct glyphs** (ⓘ info / ✓ check / gauge / ⚠
warning) on a real list — `/guides/tool-grip-force-friction-calculator-guide.html`,
the "Safety Factor Selection Guidance" list. The Turn 36 CSS-specificity bug is
fixed, deployed (verified by grepping the running POD's binary, not git), and
sprites.css re-emitted with scoped overrides. One 75,745B sheet; ≤80KB budget met.
**Hard-refresh before eyeballing — sprites.css is cached `max-age=3600`.**

## ✅ I2 IS COMPLETE (Turn 41)
I2.0–I2.5 all done and live. Sprite sheet, per-page glyph bullets (`ul.sprite-list`),
the `sprite_css_missing` fulfilment check (proven idempotent on the live fleet), and
the D10 **container house-style** (`.sprite-bullets` on the article-body wrapper —
content lists theme themselves with no markup). Proven on the friction-calculator
guide: an unclassed LLM-content list now shows sprite glyphs, the Safety Factor list
keeps its explicit info/gauge/warning, and the old JSON leak on that page is gone.

**Default glyph = arrow (user chose it 2026-07-15).** The container default is a
neutral `arrow`, not `check`; check is explicit-only (`sprite-b-check`). Implemented
in `buildSpriteCSS` (const `spriteDefaultBulletGlyph`), `SpriteCSSFormat` bumped 2→3.
It rides the next deploy and self-heals: `sprite_css_missing` sees the format
mismatch, re-emits, re-stamps. **On the next deploy, verify the gate page's content
lists flipped check→arrow** (CSS-only, no page edit) — that doubles as live proof of
the format-version half of I2.4.

**Next phase: I3** (content-linked card imagery / Lane B). See the ordered actions
at the bottom. First, though, the highest-value open item is NOT imagery — it's the
article-body content loss (below), which this workstream's image-landings trigger.

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
2. `READOUT_2026-07-16_imagery_status.md` — spoken-word status briefing (what
   we've done / where we are / where we're going). For reading aloud.
3. `PLAN_imagery_best_in_class.md` — goals G1–G9, user-confirmed decisions
   D1–D10, phases I0–I8 with dated status blocks. The map.
4. `RUNNING_NOTES_imagery_best_in_class.md` — Turns 1–43, every diagnosis
   with evidence. Append a turn each session.
5. `RUNBOOK_imagery_best_in_class.md` — the human's task queue (A-rituals,
   B-items). Done: B1–B3, B6–B8, B10–B12. Open: B4 (data-source key, at I4),
   B5 (budget sign-off), B9 (reaper cadence), B12 (verify arrow default after
   deploy), B13 (content-loss fixes — separate threads).
6. `SCOPE_I2_sprite_sheets.md` — I2 implementation scope (png→jpg revised).
7. `SHOWCASE_*.md` (3 files) — shareable summaries (one-pager / narrative /
   technical w/ diagrams). Refresh stats before reuse.
8. `SQL_2026-07-*.sql` — every migration/seed run, each with backup+verify.

## State of the world (historical detail — CURRENT state is at the top of this doc)

**As of 2026-07-16: I0, I1, I2 are all ✅ COMPLETE AND LIVE.** The blocks below
are the detailed record (kept for the hard-won bug-stacks); read the top of the
doc for the current one-screen state. Phase I3 is next; the highest-priority OPEN
item is the article-body content-loss fix, spun out to its own thread (see the
READ FIRST warning and the Spun-out section).

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

**Phase I2 (sprite-sheet bullets): ✅ COMPLETE AND LIVE (2026-07-15/16).** The
detail below traces how it got there (I2.0→I2.5); the top-of-doc summary is the
current state. Remaining loose end: the arrow-default swap is committed but not
yet live — self-heals on the next discovery pass after the format-3 deploy.
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

## Next actions, in order (updated 2026-07-16 Turn 46 — D13 built, deploy-gated)
Turn 45's B14 sequence RAN and passed (mechanism acceptance met live). What's
staged behind the NEXT deploy is D13 (per-article generation) + q78 cards:

1. **After the deploy, trigger a discovery pass for robot-hands.com**
   (improvement-loop kcat pattern, notes Turn 18) or let the loop cycle.
   `content_image_missing` now emits ~9 **needs_imagery GENERATION** items
   (content heroes, SDXL — expect real API spend, B5) — watch them complete;
   each landing also re-renders its article page via flag_rebuild (safe: guard
   live, all article bodies healthy fleet-wide).
2. **Second discovery pass** (or next loop cycle): the check sees each card's
   `origin_asset_id` ≠ its new content hero → re-emits 9 DERIVE items → cards
   re-cut at q78 from the per-article heroes. Third pass: silent (convergence
   proof, same shape as I2.4's fire-once-then-quiet).
3. **Re-render `learning-center-hub`** (needs_page, priority 5 — sections must
   RE-RESOLVE; assemble-only won't refresh the items). Precedent item shape:
   Turn 46 used item_key `needs_page:learning-center-hub:i3_cards`.
4. **A3 gate (user eyeball):** `/learning-center-hub.html` shows **9 visually
   DISTINCT per-article cards**; click-through shows the same image family on
   the article page (its content hero); `card-*.jpg` ≤60KB now (q78). The
   `learning-center-index` orphan slot clears with a listing rebuild.
5. **Then Phase I4** (data graphics — go-echarts, real series; needs RUNBOOK B4
   data-source key) or extend Lane B to news (I5) / products (I6) on the same
   entity columns.
Dispatch reminders: priority is ASC (5 = front, 99 = LAST); clear zombie claims
>10 min (they block the whole site); watched runs may need both.

**How I2 closed (2026-07-15, for the record):** to land I2.5's `sprite-bullets`
class the article-body wrapper had to exist, but robot-hands had no healthy article
page. I repaired ONE page (the friction-calculator gate page, which was also a
JSON-leak page) by extracting its article from the unparsed envelope into
`content_data.content`, added the class to the GLOBAL article-body template, and
did an assemble-only re-render. Result proven live: an unclassed content list themes
itself; the Safety Factor list keeps its explicit glyphs; the leak is gone. The
container opt-in (`.sprite-bullets ul`) exists because content `<ul>`s never carry
classes (LLM HTML into `{{.content}}`), so a per-list class is a hand-edit that
regeneration would wipe.

## Spun out of this workstream — NOT imagery, and the current top priority
Three pre-existing platform bugs surfaced during I2, all the same shape (a
background process reporting success while silently failing / losing content). Each
has a full handoff; the user is driving the article-body/image-landing pair in a
separate chat. **Do not chase these here — but heed the image-landing hazard.**
- **Image landing blanks the article body** (the one you're fixing separately):
  landing an image fires a scoped re-render that renders the never-parsed
  article-body envelope EMPTY and overwrites the good HTML → 9 pages already blanked,
  4 more leaking, across 5 sites. **STANDING HAZARD for THIS workstream: our image
  landings are the trigger — do not land an image on an affected page until the
  guard deploys.** → `../aaa_fails_to_mend/004_HANDOFF_image_landing_blanks_article_body.md`
  and root cause `../HANDOFF_2026-07-14_article_body_json_envelope.md`.
- **Product pages ship empty** and the `empty_section` fix-loop marked them
  `complete` without filling them (a loop closing without fixing). →
  `../HANDOFF_2026-07-14_empty_product_sections.md`.
