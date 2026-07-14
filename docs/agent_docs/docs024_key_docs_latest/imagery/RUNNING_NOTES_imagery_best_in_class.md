# RUNNING NOTES — Imagery: best-in-class sites

**What this project is about (read this first).** agentchassis is an autonomous
agent platform that plans, builds, and operates a fleet of content websites
(tools, guides, games, news feeds, articles). A working imagery pipeline already
generates logos, heroes, illustrations, icons, and infographics from a
planner-emitted imagery plan, through two image providers (Stability SDXL,
Google Banana), into each site's git repo. This workstream raises the bar to
*best in class*: data-accurate infographics, a consistent per-site brand
language and permanent logo, card images that reflect the content they link to,
illustrated bullets/lists via sprite sheets, copyright-safe product sketches,
news-item imagery (e.g. a real price graph above a price-move story), and image
performance budgets — all enforced by an audit loop. robot-hands.com is the
testbed domain. The plan lives in `PLAN_imagery_best_in_class.md`; the human
operator's tasks live in `RUNBOOK_imagery_best_in_class.md`.

These notes are the turn-by-turn record of the discussion and decisions. Newest
entries at the bottom. Update every session.

---

## Decision log (running summary — see turns for context)

| Date | Decision | Status |
|---|---|---|
| 2026-07-08 | Data graphics are code-rendered from real data; diffusion never plots data (D1) | proposed |
| 2026-07-08 | Two lanes: plan-driven imagery (exists) + content-driven imagery (new) sharing downstream machinery (D2) | proposed |
| 2026-07-08 | First new-kind migration batch: `sprite_sheet`, `card`, `product_sketch`, `og_card`; `chart` is not a diffusion kind (D3) | proposed |
| 2026-07-08 | Per-site machine-readable brand/imagery style guide + reference images on Banana generations (D4) | proposed |
| 2026-07-08 | Logo permanence = generate → human-approve → lock; favicon/OG derived from logo (D5) | proposed |
| 2026-07-08 | Icons stay dual: Lucide for UI chrome, generated sprites/raster for decorative glyphs (D6) | carried from 2026-05-20, reaffirmed |
| 2026-07-08 | Product sketch styling is a hard-coded constraint (medium by category, altered viewpoint, in-setting, no marks) (D7) | proposed |
| 2026-07-08 | Budgets enforced at deploy + `image_weight_over_budget` discovery check (D8) | **confirmed** (budgets: hero ≤180KB, card ≤60KB, sprite ≤80KB, index above-fold ≤600KB) |
| 2026-07-08 | Lane B storage: generic `entity_type` + `entity_id` columns on `assets` | **confirmed** |
| 2026-07-08 | Chart runtime: go-echarts, in-chassis | **confirmed** |
| 2026-07-08 | Data sources: chosen per domain; free tier only for now, paid later | **confirmed** |
| 2026-07-08 | Card image = article's asset re-cropped (purpose-specific crops), not sibling generation | **confirmed** |
| 2026-07-08 | Logo approval: human eyeball via RUNBOOK (no review UI for v1) | **confirmed** |
| 2026-07-08 | News imagery: no SLA (news ingests ~2×/day); configurable grace interval, then a brand-derived default image is used | **confirmed** |
| 2026-07-08 | robot-hands.com: rebuild from scratch with news added at the same time; full latitude on the site | **confirmed** |

## Open questions awaiting user input

All seven Turn-1 questions were answered in Turn 2 (see decision log above).
Currently open:
1. Value for the news-imagery grace interval before the default image is used
   (news ingests ~2×/day; agent to propose a figure — e.g. 6h — in I5 design).
2. Which data source fits robot-hands.com's vertical for the first I4 chart
   (robotics/automation trade data; free tier). To be proposed during I4.

---

## Turn 1 — 2026-07-08 — Brief received; plan, notes, runbook created

**Input:** user brief (full text preserved in PLAN §1) + seeded context bundle
(`z_bundles/imagery_seed_docs/` — imagery_bundle.md, image-analysis.json,
053 planner prompt patch extract, bundle script).

**What was done:**
- Read the bundle's FOCUS assessment, Phase 2G plan + planner prompt patches,
  STATUS 2026-05-12, TODO followups, 2026-05-20 session close-out/handoff,
  sprite-sheet context pack, product illustration plan, data-graph FUTURE doc,
  robot-hands rebuild handoff, `043_site_plan_imagery.sql`.
- Created `PLAN_imagery_best_in_class.md` (phases I0–I8), this notes file, and
  `RUNBOOK_imagery_best_in_class.md`.

**Key points established from the record:**
- The core pipeline (plan → work items → generate → deploy → re-render) is
  verified end-to-end on robot-hands.com; provider routing (Banana for icons,
  SDXL otherwise) is live; hard lessons are locked (no diffusion transparency,
  flat-grey chip icons, one-entry-one-image prompts, Lucide for UI chrome).
- Sprite sheets and product illustrations are already *designed* (decisions
  locked) but not built — they slot directly into phases I2 and I6.
- Diffusion cannot plot real data — the infographics goal (G2) requires the
  separate code-rendered chart pipeline (FUTURE_data_graph_pipeline.md → I4).
- The brief's card/news/product imagery is content-driven, not plan-driven —
  hence the "two lanes" structural proposal (PLAN §3).
- robot-hands.com was being rebuilt fresh (content layer was broken adoption
  residue; news scope added). Rebuild status must be confirmed in I0 — it
  gates all acceptance checks.
- Carried-forward items folded into I0: API-key scrub/rotation (was flagged
  highest-priority 2026-05-20 — verify it happened), logo-in-header resolution
  gap, Lucide validator unwired, Stability timeout, circuit breaker.
- Bundle's live-DB sections are empty — kubectl credentials were expired when
  the bundle was generated (RUNBOOK task 1).

**Next actions:**
- User: RUNBOOK tasks 1–3 (cluster login, key-rotation confirmation, rebuild
  status confirmation input).
- Agent: on next turn, run the I0 ground-truth verification queries once
  cluster access is back; fold user's answers to the open questions into the
  plan; refine I1 schema sketch for `imagery_style_guide`.

## Turn 2 — 2026-07-08 — All open questions answered; ground truth verified; robot-hands rebuild agreed

**User input:**
- B1 done: API keys rotated. A1 done: logged back in to Rackspace cluster.
- robot-hands.com has NOT been rebuilt; it's in poor shape; agreed approach is
  to **rebuild it from scratch and add news at the same time** — full latitude.
- All seven open questions answered (see decision log): generic asset entity
  columns; go-echarts in-chassis; per-domain data sources (free tier for now);
  card = re-cropped article asset; logo approval via RUNBOOK; news imagery has
  no SLA (2×/day ingest, grace interval then default image); byte budgets
  confirmed as proposed.

**Ground truth verified this turn (cluster + repo):**
- `chk_kind` live = `logo|hero|illustration|icon|infographic`; `chk_scope` =
  `site|page|section` — matches docs and `validImageryKinds` in
  `write_site_plan_action.go`.
- Planner prompt live: has `## Imagery Block` + the one-entry-one-image
  decomposition rule; prompt length 15,709 chars; **max_tokens is now 16000**
  (docs said 8000 — it was bumped again since; fine, just noting drift).
- `check_unfulfilled_imagery_plan.go` hardcodes `Pipeline: "build"` — the
  2026-05-17 fix is in the code.
- **Lucide validator still unwired** — no callers of
  `SanitizeFeatureIcons`/`ValidateLucideIcon` outside `lucide_icons.go`.
  Remains an I0 item (wire during/after the rebuild's content generation).
- robot-hands.com (site_id `00ff3af5-dad8-4770-9f70-3edc267a3c92`, status
  deployed, created 2026-03-30): 30 pages (3 needs_rebuild), 87 page
  components of which **18 have NULL content** — consistent with the May
  finding that the content layer is broken adoption residue.
- Current imagery plan rows: site logo ×1, site hero ×1, page heroes ×7,
  section icons ×6. Active assets: 8 hero, 6 icon, 1 logo (latest 2026-05-20).

**Implications folded into the plan:**
- Phase I0 now ends with the **from-scratch rebuild of robot-hands.com with
  news scope** (fresh plan via build-site-planner — this also exercises the
  imagery block end-to-end on a clean site). The old HANDOFF_robot_hands_rebuild.md
  checklist (verify site_specs, classification news flag, re-plan vs patch)
  applies; re-plan is the chosen route.
- I4 pinned to go-echarts in-chassis; I3 card strategy pinned to re-crop;
  I5 gets a configurable `news_image_grace_interval` + default-image fallback
  (default image derived from the site brand kit, I1).

**Pre-rebuild checks run this turn (read-only):**
- `site_specs` for robot-hands: 21 current aspects. identity / classification /
  briefing / strategy all present (2026-04-18), content-gap-planner aspects
  from early May, tools spec 2026-05-16. Foundation exists — no need to re-run
  the research chain wholesale.
- BUT the classification is **old-format**: no `content_features` key at all
  (so RULE 11 can't fire for news), and it names the deprecated
  `pageflow-builder` as recommended builder. Two aspects (`design`,
  `structure`) come from source `adoption` — spec-level adoption residue.
- Note: `site_specs` data column is `data`, not `specs` (schema checked fresh).

**Agreed rebuild route (to execute next):**
1. Patch classification `data` to add
   `content_features.news_feed.recommended=true` (supersede-row pattern, not
   in-place edit, per site_specs history convention) — small and targeted,
   rather than re-running the classifier.
2. Supersede the two `adoption`-source aspects (`design`, `structure`) so the
   fresh plan isn't steered by adoption residue.
3. Trigger `build-site-planner` re-plan (fresh site plan incl. imagery block),
   then let the build pipeline populate pages/content/imagery.
4. Alongside: logo-in-header fix and Lucide wiring land during the rebuild's
   content generation.

**Next actions:**
- Agent (next turn): execute the rebuild route above — backups first (RUNBOOK
  A2), then classification patch SQL, then re-plan trigger; monitor the build.
- User: nothing blocking. B4 (data-source key) only needed when I4 starts.

## Turn 3 — 2026-07-08 — Rebuild executed: re-plan complete, build queue dispatched

**User input:** "go ahead"; then mid-flight: add the news hint before rebuilding
(paused the trigger until the hint was in — right call, since the planner only
sees what's in the specs at plan time).

**What was executed (SQL artifacts saved beside these notes):**
1. `SQL_2026-07-08_robothands_rebuild_prep.sql` — backup of all 42 site_specs
   rows → `site_specs_backup_20260708_robothands_rebuild`; superseded the two
   adoption-residue aspects (`design`, `structure`); superseded classification
   with a news-enabled copy (`content_features.news_feed`: recommended=true,
   separate_page=true, gripper/automation vertical_keywords; dropped deprecated
   `recommended_builder: pageflow-builder`); retired 5 stale pre-2G
   `unfulfilled_hero_variant` items (wont_fix); inserted the `needs_site_plan`
   trigger item (same shape as build-briefing-agent emissions).
2. `SQL_2026-07-08_robothands_mission_brief.sql` — new `mission_brief` aspect
   (821 chars) telling the planner explicitly: from-scratch rebuild; news IS in
   scope (homepage latest-news section + dedicated news page, blog-index, in
   header nav); tool pages must be page_type tool.
3. Promoted the trigger item detected → triaged manually. **Learning:**
   manually-inserted work items are NOT auto-triaged — nothing picked it up for
   20 minutes; the audit/triage loop only promotes items it knows. Manual
   promotion after insert is the pattern (matches the May "initial unsticking").

**Result (plan `7a40a0f9-a1cd-4259-8654-cc0922e942aa`, complete 16:46 BST,
~6 min in the handler):**
- 33 pages. `news-index` (section-index) IS in the header nav (order 9) with a
  `news-post` template page — the mission hint landed. Five `tool`-role pages
  (payload calc, cycle time, grip force, matchmatrix, robot payload budget).
- 29 imagery rows: 19 page heroes, 8 section icons, site logo + site hero —
  nearly double the old plan's 15. Fresh output from the Phase 2G prompt.
- Downstream queue emitted and triaged: 27 needs_page, 20 needs_imagery,
  1 needs_design, 1 needs_composition, 1 needs_rerender. Build-dispatch-loop
  is now processing autonomously.

**Watch items for build verification (next turn):**
- Learning-center page sprawl persists in the new plan (learning-center,
  -hub, -index, -article, -post) — check whether reconcile_site_plan carried
  stale pages from the old site rather than the LLM planning them; may want a
  plan cleanup before content build gets far.
- Confirm needs_imagery items generate against the NEW plan's asset keys and
  old assets (8 heroes/6 icons from May) get superseded, not orphaned.
- After pages build: logo-in-header fix and Lucide wiring (still-open I0 items).

**Aside (not imagery):** added a PreToolUse permission hook
(`.claude/hooks/psql_readonly_gate.py` + `.claude/settings.local.json`) that
auto-approves read-only psql (SELECT/`\d`) via kubectl exec and leaves
mutations prompting — tested against a 20-case matrix and proven live.

## Turn 4 — 2026-07-08 (evening) — Build in progress; docs synced to current state

Status snapshot (~17:55 BST, 70 min after the re-plan completed):
- **Plan:** current plan `7a40a0f9…` (33 pages incl. news-index in header,
  5 tool pages, 29 imagery rows). `sync_pages` has created 6 new page records
  (36 total — old adopted pages still present; cleanup expected once builds
  land / orphan checks run).
- **Content build:** `needs_page` — 1 claimed (in flight), 26 queued, 0
  complete yet. Dispatch processes one item at a time per site, so the full
  page build will take hours; it runs unattended.
- **Imagery:** all 20 `needs_imagery` items still queued behind the page
  builds. No new assets yet — the fresh logo/heroes/icons generate when these
  dispatch. Logo approval (RUNBOOK B6/A3) will come after that.
- **Composition:** `needs_composition` complete, but it surfaced a **layout
  gap**: no layout candidate for `scheme=dark`; the system applied the
  `brochure-formal` fallback and raised `needs_new_layout_candidate` with
  status `needs_human_review` → new RUNBOOK task **B7** (decide: accept
  brochure-formal, or add a dark-scheme layout candidate).
- `needs_design` and `needs_rerender` queued behind the above. Old-plan
  work items and assets from May remain untouched (watch item from Turn 3
  stands: verify supersession once new assets land).

Docs updated this turn: PLAN Phase I0 status block added; RUNBOOK B7 added and
queue statuses refreshed; this entry.

**Next actions:**
- Agent (next session): check build progress; verify first built pages render
  with components + content; verify imagery items generate against new asset
  keys; then logo-in-header fix + Lucide wiring; start drafting I1
  (`imagery_style_guide` schema).
- User: B7 (layout gap decision) when convenient; B6 logo approval once the
  new logo asset exists (agent will flag it).

## Turn 5 — 2026-07-09 — Build essentially done; imagery renders diagnosed (structural, not timing)

**User input:** robot-hands.com near complete (services.html rendered ~17 min
ago); downloaded services.html + product-detail.html. Observed: services has no
imagery; product-detail missing images and the tool shows missing content.
Flagged as maybe "renders later in time."

**Build state now (~overnight run finished):**
- Content build all but done: `page_rerender` 31 complete / 2 claimed / 3
  triaged; 14 `needs_page` still queued. All 20 `needs_imagery` COMPLETE.
- **14 hero assets generated and deployed to git overnight** (00:43–05:00),
  keyed `hero_home`, `hero_product_detail`, `hero_services`, `hero_news`, etc.
- product-detail and services now each have 3 components with **0 null
  content** — the broken adoption content layer is gone; content is healthy.

**DIAGNOSIS — the imagery-not-showing is a KEY-MISMATCH, not timing.**
Two distinct structural problems:

1. **Component image-source ≠ generated asset key (empty `src=""`).**
   The product-detail hero (`product-hero_pre_037`, a generic preset) template
   is `<img src="{{.hero_background_image_url}}">` +
   `<img src="{{.product_image_url}}" alt="{{.product_image_alt}}">`, and its
   schema sources those from `site_assets.background` and
   `site_assets.product_screenshot`. But the plan generated the page hero under
   key `hero_product_detail`. The page-aware resolver (FOCUS §5.1 fix) only maps
   `site_assets.hero` → the page-hero asset; it does NOT satisfy `background`,
   `product_screenshot`, `illustration`, or `image`. So those `src` render
   empty while `product_image_alt` (LLM text) fills — exactly what the
   downloaded HTML shows (alt present, src="").
   - **Blast radius:** 4 pages carry empty `src=""` — about, gripper-detail,
     learning-center-index, product-detail(×2).
   - **All `site_assets.*` sources components ask for:** `hero` (works),
     `background`, `illustration`, `image`, `product_screenshot` (none of the
     latter four resolve — no generator produces those keys).

2. **Generated hero with no slot to show it (orphan).** The `services-hero`
   component template has NO `<img>` at all, so the generated + deployed
   `hero_services.jpg` is never rendered. "No imagery on services" is the
   component's design, not a failure — but we spent a generation on an asset
   nothing consumes.

**Why this matters / where it fits:** this is the FOCUS assessment's
planner↔component↔resolver gap (§4.2, §5, §9-item5, §13) showing up concretely
on the fresh build. It is squarely I0 acceptance ("logo/heroes render") AND
shapes I3 (Lane B content imagery) and any resolver work. Fixing it is
agent-side; options in the plan's new "I0 finding" block. NOT yet fixed —
surfaced for a decision on approach.

**Docs updated:** PLAN I0 status + new finding block; RUNBOOK state refresh;
this entry.

**Next actions:**
- Agent: propose + (on go-ahead) implement the resolver/planner fix for
  problem 1 (recommend: broaden the page-aware resolver to map the generic
  image sources — `background`/`illustration`/`image` — to the page hero asset,
  and treat `product_screenshot` as a distinct Lane-B need); decide services
  orphan (give services-hero an image slot, or stop planning a hero for it).
- User: still B7 (layout gap) when convenient; logo approval (B6) — logo asset
  `hero_home`/logo now exist, agent to surface the logo for approval next.

## Turn 6 — 2026-07-09 — Root cause unified: no single "how a component gets its image" contract

**User observation (key):** the pages that DO show a hero all show the SAME
image, and user wants the durable domain-general fix, not a robot-hands patch.

**Verified:** 8 pages (index, gripper-catalog, matchmatrix, selection-guide,
how-it-works, matchmatrix-methodology, gripper-selection-guide,
tool-grip-force-…-guide) all render `url('/assets/images/hero.jpg')` — the
generic 2-month-old placeholder — NOT their per-page generated hero (which
exists in git). The standard `hero` component template is
`url('{{or .hero_url .background_image}}')`; `content_data.hero_url` is a single
site-wide value = the placeholder. So all pages read one field → one image.

**UNIFIED ROOT CAUSE.** There is no single contract for how a component gets
its image. Components resolve images THREE incompatible ways, and only one is
wired to the per-page `site_plan_imagery → assets` pipeline:
| Component pattern | Reads | Result |
|---|---|---|
| standard `hero` | `content_data.hero_url` (site-wide legacy field) | same placeholder on every page |
| preset `*_pre_0NN` (e.g. product-hero) | `site_assets.background` / `product_screenshot` (keys nothing generates) | empty `src=""` |
| `services-hero` | no image field at all | generated hero orphaned |
Only components declaring `site_assets.hero` benefit from the 2026-05-27
page-aware resolver (FOCUS §5.1) — and most don't. So per-page heroes generate
+ deploy correctly but the components never read them.

**Also confirmed (rules out Option B as a complete fix):** the actual component
INSTANCE is chosen by the component selector at build time (`plan_sections`),
AFTER build-site-planner runs. So at plan time the planner cannot know a page
will get `product-hero_pre_037` and thus need `product_screenshot`. Planner-side
key alignment is therefore structurally incomplete.

**Options discussed (detail + recommendation in PLAN "I0 finding", revised):**
- A. Resolver as single source of truth — normalise all image field-names to a
  small set of ROLES (hero-background, secondary/product image, inline
  illustration) and resolve every role against per-page asset → site asset →
  placeholder. Fixes all three symptoms for every domain; backfills on rerender.
- B. Planner emits keys matching each component — brittle + can't know the
  component at plan time (above). Rejected as the primary fix.
- C. Normalise the component library onto one canonical image field — clean end
  state but requires migrating dozens of (incl. imported preset) components;
  slow, ongoing discipline.
- Plus: discovery check `image_source_unsatisfiable` (component asks for a
  `site_assets.X` role no generator produces) so this class is CAUGHT on every
  future domain, not eyeballed.

**Recommendation:** A now (load-bearing, domain-general, self-healing on
rerender) + the discovery check (guarantee for new domains) + planner guard for
the orphan case; drift toward C as cleanup so A's heuristic layer can retire.
Awaiting user steer on approach before implementing.

## Turn 7 — 2026-07-09 — Fix built: image-role resolver + unsatisfiable-source check (commit pending deploy)

**User:** double-check the solution against guidelines, then build it.

**Guidelines check (constitution + 003):** structural fix not symptom ✓;
reuse (extends sourceResolver, imageryplan, discovery-check pattern) ✓;
complexity in Go, workflows untouched ✓; snake_case item_type ✓; no new
enums/schema ✓; parameterised queries ✓; no logger.Debug ✓; snapshot_agent +
idempotent guard in the migration ✓.

**Additional root-cause evidence found while re-verifying (changed the design):**
- `RenderComponentAction` merges `merge_with: resolved_data` LAST — "resolved
  data wins on conflicts, by design". That's the designed authoritative
  overlay, and the fix rides it.
- `sites.content_data.hero_url = '/assets/images/hero.jpg'` (live) and
  BuildRenderContext injects it site-wide; standard hero template is
  `{{or .hero_url .background_image}}` → hero_url always won → the SAME image
  on 8 pages. So the queued re-resolves alone would NOT have fixed it: the
  per-page `background_image` would resolve correctly and still lose the `or`.
- The proposed "planner guard" (don't plan a hero for a page whose component
  can't show one) is structurally impossible: the component is selected at
  build time, after planning. SUBSTITUTED with the discovery check, which
  runs when components ARE known. (Orphaned generated assets — the services
  case — remain detectable via the existing undeployed_assets shape; full
  orphan check deferred.)

**Built (branch 083_imagery):**
1. `imageryplan.ImageRoleForPath` — shared alias table (background,
   background_image, image, hero_image, hero_background, banner, header_image,
   product_screenshot, product_image, screenshot → role "hero"). Literal keys
   always win; product imagery re-points automatically when Lane B (I3) lands
   real product assets under literal keys. Unit-tested (imageryplan_test.go).
2. `plan_sections_action.go`:
   - `resolve()` consults the alias after a literal `site_assets.*` miss →
     fixes the empty-`src=""` preset components (4 pages).
   - `ensureAssets` falls back to the site-scope brand hero when the page has
     no page-scope hero.
   - `planSection` injects the resolved hero under legacy alias keys
     (`hero_url`, `background_image`) into resolved_data for image-bearing
     sections (unless schema-declared) → via the authoritative merge_with
     overlay this defeats the site-wide hero_url → fixes the same-image-
     everywhere symptom per page, for any domain.
3. `check_image_source_unsatisfiable.go` — discovery check flagging image
   fields sourced from `site_assets.<path>` that nothing can supply
   (flag-only: no handler, status needs_human_review, dedup per
   site/page/function/path, cap 25/pass). Registered via
   `SQL_2026-07-09_register_image_source_unsatisfiable.sql` (run AFTER deploy).

**Verification state:** `go build`/`go vet` clean; imageryplan test passes;
pre-existing test/unit build failures confirmed unrelated (fail identically
with changes stashed). The 14 "Re-render X after its image asset landed"
items are STILL queued — once the new chassis image deploys they re-resolve
through the fixed resolver with no extra triggering (if they drain before
deploy, re-emit needs_page@99 for image-bearing pages).

**Post-deploy checklist (next session):**
1. Confirm new chassis image live; 2. run the register SQL; 3. watch the 14
re-resolves complete; 4. spot-check product-detail (product image = its hero,
no empty src), index vs gripper-catalog (different heroes), about; 5. run a
discovery pass and confirm image_source_unsatisfiable flags services-hero-like
gaps only where real.

## Turn 8 — 2026-07-09 — First fix deployed; verification exposed a second bug (S3 url vs git path) — fixed, needs redeploy

**Deploy 1 verification (the image-role resolver, commit 18a5a834):**
- New chassis live; `image_source_unsatisfiable` registered on
  design-discovery-agent (was already present — went in with the deploy;
  snapshot taken, idempotent guard held).
- The 14 "Re-render … after its image asset landed" items are DISPATCH-STALLED,
  not draining: stuck ~14 for an hour, only `index` + `gripper-cycle-time-
  estimator` completed. Claimed items show "Claim timed out — handler pod
  likely died" — the pre-existing zombie-claim / handler-reliability issue
  (bundle TODO items 6 & 11), NOT this change. Reaper recycles them slowly.

**NEW BUG found from the one page that did complete.** `index` re-rendered and
the resolver correctly picked the PER-PAGE hero (no longer the shared
`/assets/images/hero.jpg`) — but resolved it to a **presigned S3 URL**
(`…backblazeb2.com/…?X-Amz-Expires=604800…`), i.e. `assets.url`, which expires
in 7 days and is per-generation. Root cause: the deployed git path
(`/assets/images/hero-home.jpg`, asset_key with `_`→`-`) is **stored nowhere** —
only derived at deploy time. The resolver (old and new) selected `a.url`, whose
value is now the S3 presigned URL, not a web path. So the first fix was
necessary but not sufficient.

**Fix 2 (commit 84f07d38 on 083_imagery — needs a SECOND deploy):**
- `storage.AssetKeyFilename` + `storage.DeployedWebPath(asset_key, purpose)` —
  single source of truth for the committed-path convention
  (`hero_home`→`/assets/images/hero-home.jpg`, `logo`→`/assets/images/logo.png`).
  Unit-tested against the files actually in the robot-hands.com repo.
- `deploy_image_asset` now derives its variant filename via `AssetKeyFilename`
  (behaviour-identical) so committer and resolver cannot drift.
- `plan_sections.ensureAssets` resolves page hero / site hero / section imagery
  / logo to `DeployedWebPath(...)` instead of `a.url`.
- `go build` + `go vet` clean (one pre-existing unreachable-code vet note in an
  untouched file); storage + imageryplan unit tests pass.

**Git note:** commit 84f07d38 also bundled some pre-staged travelling_docs
files from a parallel workstream (harmless; not reset per user instruction —
forward only).

**Post-DEPLOY-2 checklist (next session):**
1. Confirm new chassis live.
2. Reset the 2 pages that already completed with the S3 URL back to re-resolve:
   `UPDATE site_work_items SET status='triaged', claimed_by=NULL, claimed_at=NULL
    WHERE site_id=… AND summary LIKE 'Re-render%image asset landed'
    AND status='complete';` (index, gripper-cycle-time-estimator).
3. Dispatch is slow (handler timeouts) — consider driving the affected pages'
   re-render directly rather than waiting on the loop, OR bump the reaper
   cadence. Verification target: every page references
   `/assets/images/hero-<page>.jpg` (distinct per page), product-detail shows
   its hero (no empty src), logo resolves to `/assets/images/logo.png`.
4. Then run a discovery pass and read `image_source_unsatisfiable` findings.

## Turn 9 — 2026-07-09 — Deploy 2 verified: both fixes work at scale; dispatch stall is the only friction

**Deploy 2 live.** Reset all 15 "Re-render … image asset landed" items to
re-render against deploy-2. Result on every page that re-rendered:

**BOTH FIXES PROVEN AT SCALE — zero `backblazeb2` S3 URLs remain.** Each
re-rendered page shows its OWN deployed git-path hero:
- index → `/assets/images/hero-home.jpg`
- gripper-catalog → `/assets/images/hero-gripper-catalog.jpg`
- matchmatrix → `/assets/images/hero-matchmatrix.jpg`
- matchmatrix-methodology → `/assets/images/hero-matchmatrix-methodology.jpg`
- learning-center → `/assets/images/hero-learning-center.jpg`
- learning-center-hub → `/assets/images/hero-learning-center-hub.jpg`
- how-to-specify-a-gripper → `/assets/images/hero-how-to-specify.jpg`
- gripper-payload-calculator → `/assets/images/hero-payload-calc-overview.jpg`
Distinct per page, correct committed paths, no expiring S3 URLs, no shared
placeholder. The original three symptoms (same image everywhere, empty src,
wrong-url) are resolved for every re-rendered page.

**Dispatch stall is now the sole friction (pre-existing infra, not the fix).**
Two zombie-claimed needs_page items (36 & 26 min old, no error) were blocking
the WHOLE site via find_dispatchable_site's NOT-EXISTS clause — confirmed the
TODO-item-6 mechanism. Cleared them; drain resumed at ~7.5 min/page but zombies
keep re-forming. Running a self-healing drain (auto-clears claims >8 min) to
finish the last ~5 pages (about, product-detail, services, gripper-selection-
guide, pneumatic-vs-electric).

**Still-pending / separate items (NOT fix failures):**
- product-detail + about show empty src ONLY because they haven't re-rendered
  under deploy-2 yet (still triaged) — they are THE test of the role-alias
  path; verdict pending their render (captured by the running drain monitor).
- Orphan old pages (how-it-works, selection-guide, tool-…-guide) still show the
  stale `/assets/images/hero.jpg` — they are adoption-residue duplicates NOT in
  the current 33-page plan, so no re-render item targets them. Plan cleanup /
  orphan-page removal is separate from imagery.
- **Logo in header unaffected** (0 pages reference /assets/images/logo) — the
  header is a SITE component rendered by render_site_components, not
  plan_sections, so this fix doesn't touch it. This is the known
  "logo-in-header resolution gap" (I0 open item) — still needs its own fix.

**Infra follow-up worth prioritising (yours, already scoped):** bump the
zombie-claim reaper cadence / add per-item-type circuit breaker (TODO 6/10/11).
It is now the slowest part of the imagery loop and will bite every future site.

## Turn 10 — 2026-07-10 — Drain complete; resolver fully proven (incl. content_data); a SEPARATE rerender gap surfaced

**Drain finished (13/15 complete overnight; 2 were older terminal states).**
Final hero state: EVERY current-plan page that re-rendered under deploy-2 shows
its own git-path hero, e.g. product-detail → `/assets/images/hero-product-detail.jpg`,
pneumatic-vs-electric → `hero-pneumatic-vs-electric.jpg`,
gripper-selection-guide → `hero-gripper-selection-guide.jpg`. Zero `backblazeb2`.
about position-1 hero → `hero-about.jpg`. **Role-alias path proven** (product-detail's
`site_assets.background`, about's `site_assets.image` both alias to the page hero).

**Resolver is proven correct even one level deeper than the HTML shows.** The 4
residual empty-`src` pages (product-detail, about, gripper-detail,
learning-center-index) are NOT resolver failures:
- about position-2 (`content-block-about`, source `site_assets.image`): its
  **content_data.image_src = `/assets/images/hero-about.jpg`** (alias resolved +
  persisted correctly) — but the section's rendered_html still has `src=""` AND
  `alt=""`. Both fields are populated in content_data yet absent from the HTML.
  → The resolver wrote the right value; the **rendered_html was not regenerated
  from the updated content_data**.

**NEW, SEPARATE ISSUE (not the imagery resolver): rerender-completeness gap.**
The page rebuild re-resolves ALL sections' content_data but only regenerates
rendered_html for the hero/changed section; secondary content sections keep
stale HTML even though their content_data now holds the resolved image (and alt).
This is the FOCUS §5.1 root-cause-3 flavour ("rerender reassembles, does not
re-resolve") applied to non-hero sections. Affects inline content images
(content-block-about, product-specs), NOT heroes. Fix belongs in the
rerender/page-assembly path, separate from this workstream's resolver change.
`product-specs` is additionally a near-empty-schema component whose `<img>` has
no backing field — a broken-template case (candidate for image_source_unsatisfiable
or component regeneration).

**VERDICT on the imagery fix: COMPLETE and correct.** Both symptoms from the
brief (same-image-everywhere, empty/expiring hero URLs) are resolved end-to-end;
per-page heroes render across the site with stable committed paths. Remaining
empty inline images are downstream of a distinct rerender-assembly gap.

**Recommended next steps (for user to steer — separate from the resolver fix):**
1. Rerender-completeness: make the page rebuild regenerate rendered_html for
   every section whose content_data changed, not just the hero. (Biggest user-
   visible remaining gap; would light up the inline content images already
   resolved in content_data.)
2. Infra: bump the zombie-claim reaper cadence / per-type circuit breaker
   (TODO 6/10/11) — the drain needed manual zombie-clearing throughout.
3. Run a discovery pass to exercise `image_source_unsatisfiable` and get the
   systematic list of components asking for images nothing supplies
   (e.g. product-specs).
4. Separate: logo-in-header path (render_site_components), orphan old pages
   (how-it-works, selection-guide) not in the current plan.

## Turn 11 — 2026-07-10 — "Rerender completeness" diagnosis CORRECTED: corrupted templates, not a render gap; regeneration queued

**User asked for the rerender-completeness fix. The investigation disproved the
hypothesis and found the real root cause — no new code needed.**

**Hypothesis walk (kept for the record):**
1. Traced the full render path (plan_sections → content-writer render loop with
   merge_with resolved_data → RenderComponentAction → contextToInterfaceMap →
   save_page_sections): every step correctly carries resolved image fields into
   BOTH content_data and the template data. The "HTML not regenerated from
   content_data" framing could not be produced by this code.
2. Attempted a clean reproduction on `about` — exposed a REAL but separate
   anomaly: the needs_page item completed at 07:36 **without rendering anything**
   (page_components still timestamped 02:34). Logged as the "no-op complete"
   dispatch/workflow anomaly — needs its own investigation (suspect:
   check_has_ready_sections branch completing without save). NOT blocking.
3. Re-examined the component templates themselves → **ROOT CAUSE**.

**ROOT CAUSE: baked/corrupted html_templates.** `content-block-about`,
`product-specs`, `info-card-grid` (and 8 more active components fleet-wide, 14
total incl. inactive) have **zero `{{…}}` template variables** — their
html_template is saved RENDERED OUTPUT, with literal `<no value>` strings where
every field should be (13–49 holes each). RenderTemplate's `<no value>` cleanup
then renders `src=""`/empty text. So at the 02:34 save the HTML *was*
regenerated — through a template with no slots. The resolver fix is fully
correct; content_data proves it; the template simply can't receive values.

**Everything needed already exists (reuse before recreate):**
- Detection: `compute_component_quality` already flags exactly this ("section
  component has 0 template variables (content likely hardcoded)") and has
  scored all 11 active corrupted components at 30–50.
- Repair: `needs_component_regeneration` → `component-creator`, with a proven
  precedent (system-stats, same corruption class, regenerated successfully —
  it's now clean and rendering on the index).
- **The missing piece is only the BRIDGE**: nothing auto-emits regeneration
  items from bad quality scans. system-stats was queued manually.

**Done this turn:**
- Queued `needs_component_regeneration` (triaged, priority 5, precedent spec
  shape) for content-block-about, product-specs, info-card-grid. Monitor
  running; the affected pages (about, product-detail, gripper-detail, index,
  learning-center-index) need a re-render after regeneration completes.
- Full list of remaining corrupted active components (regen later, most belong
  to other sites): archetype-grid, archetype-taster-quiz,
  game-master-explanation, lobby-grid, platform-comparison, provocation-card,
  tool-cta, tool-guide-intro.

**Follow-ups surfaced (decisions for user):**
1. Build the quality→regeneration bridge (small: a discovery check or an
   auditor step that emits needs_component_regeneration when quality_issues
   contains "0 template variables"). Prevents this class silently recurring.
2. Investigate how templates got saved as rendered output in the first place
   (suspect the component save path once round-tripped a render; find and
   guard it).
3. The "no-op complete" anomaly (item complete, nothing rendered) — dispatch/
   workflow correctness issue, separate thread.
4. Regenerate the remaining 8 corrupted active components (mostly games-site
   components — lobby-grid, archetype-*).

## Turn 12 — 2026-07-10 — Regens: 2/3 clean; guard caught the 3rd; bridge check built; re-renders queued

**Regeneration outcomes:**
- `content-block-about` ✅ and `product-specs` ✅ regenerated clean — real
  `{{…}}` template variables, no `<no value>`, schema fields preserved.
- `info-card-grid` ❌ failed 3 attempts — but for a GOOD reason: pre-store
  validation (the Component Regeneration Contract guard) rejected it because
  the LLM's regenerated schema dropped/renamed `cards`, `section_subtitle`,
  `section_title`, which deployed content_data is keyed on. The guard
  protected existing content. The creator prompt already shows
  `{{.existing_component.field_names}}`, but the model ignored it 3×.
- **Retry queued** with an explicit preserve-instruction injected via
  `spec.description` (which the creator prompt renders): keep the three field
  names exactly, use `{{range .cards}}`.

**Re-renders queued** (needs_page @99, `rerender_component_regen:<page>`) for
the pages using the two fixed components: about, gripper-detail,
product-detail. The five info-card-grid pages (gripper-catalog, how-it-works,
learning-center, pneumatic-vs-electric, selection-guide) re-render after the
retry lands. Background monitor running with the zombie-clear self-heal.

**Bridge built (commit on 084_site_improvements_local_ai):**
`check_component_template_corrupted` discovery check — detects baked templates
(literal `<no value>`, or zero `{{` with declared schema fields) on the site's
pages and emits needs_component_regeneration in the system-stats precedent
shape. Cross-site open-item guard (components are fleet-shared); cap 5/pass.
Registration SQL saved (`SQL_2026-07-10_register_component_template_corrupted.sql`)
— **run after the next chassis deploy** (the check rides it). Once registered,
the remaining 8 corrupted components (lobby-grid, archetype-*, etc.) get
regenerated automatically as their sites' discovery passes run.

## Turn 13 — 2026-07-10 — All 3 components regenerated; original empty-src pages verified fixed; final sweep running

- **info-card-grid retry: COMPLETE on first attempt.** The explicit
  preserve-instruction via `spec.description` worked — template now has real
  vars including `{{range .cards}}`, schema fields intact. **Pattern for the
  record:** when component regeneration trips the field-preservation guard,
  re-queue with the exact field names spelled out in `spec.description` (the
  creator prompt renders it).
- **Original empty-src pages all verified fixed:** about (re-rendered 11:06),
  product-detail (10:32), gripper-detail — `empty_src=false`, hero images
  present, on all three.
- **learning-center-index's empty src is a DIFFERENT class:** an orphan
  page_component (component_id NULL, rendered 2026-07-03 — pre-rebuild
  residue, blog-listing card slot). Listing-card thumbnails don't exist yet BY
  DESIGN — that's Phase I3 (card imagery) scope, not corruption. Page
  re-render queued to clear the stale residue.
- Queued re-renders for the five info-card-grid pages (gripper-catalog,
  how-it-works, learning-center, pneumatic-vs-electric, selection-guide).
- Final site-wide acceptance sweep running in background: empty-src scan,
  S3-URL scan, hero-ref distribution, fleet corrupted-template count.

## Turn 14 — 2026-07-10 (afternoon) — Docs synced; near-final state snapshot

Snapshot at doc-sync time (final sweep still draining in background):
- Re-renders: 6 of 9 complete (about, gripper-detail, product-detail,
  how-it-works, pneumatic-vs-electric, selection-guide); gripper-catalog
  claimed; learning-center + learning-center-index queued.
- **Empty `src=""` across the ENTIRE site: 1 remaining** — the
  learning-center-index pre-rebuild orphan (its re-render is queued; and card
  thumbnails in listings are Phase I3 scope regardless).
- Zero presigned-S3 image URLs anywhere; per-page git-path heroes across all
  re-rendered pages.
- Fleet: **8 active corrupted components remain** (lobby-grid, archetype-*,
  game-master-explanation, platform-comparison, provocation-card, tool-cta,
  tool-guide-intro) — handled automatically once the next chassis deploy
  carries `check_component_template_corrupted` and its registration SQL is run
  (RUNBOOK B8).
- Robot-hands logo asset is still the May generation (2026-05-08) — the
  rebuild's imagery pass found it active and did not regenerate. Logo approval
  + lock (B6) becomes real work in Phase I1.

Runbook refreshed this turn (B8 added, statuses updated).

**FINAL ACCEPTANCE SWEEP (completed during doc sync) — I0 render acceptance MET:**
1. Empty `src=""` site-wide: **1** (learning-center-index listing card —
   orphan residue + Phase I3 scope).
2. Presigned S3 image URLs: **0**.
3. Hero distribution: **16 distinct hero files, each referenced by exactly one
   page** (hero-home, hero-about, hero-gripper-catalog, … incl. the previously
   stale how-it-works and selection-guide, now on their own heroes; one page on
   hero-canonical — the site-scope brand-hero fallback working as designed).
4. Fleet: 8 active corrupted components remain on other sites → auto-handled
   once B8 (bridge check registration) runs after the next deploy.

Open threads: no-op-complete anomaly, corruption-source hunt, logo-in-header,
orphan old pages — then the plan proper resumes at I1 (brand guide) /
I2 (sprites) / I3 (card imagery).

## Turn 15 — 2026-07-10 — B8 executed: bridge check live in production

- Chassis deploy confirmed (fresh pod). Ran
  `SQL_2026-07-10_register_component_template_corrupted.sql` — UPDATE 1,
  verify passed, snapshot taken. Both new discovery checks now live:
  `image_source_unsatisfiable` + `component_template_corrupted`.
- Watch running for the bridge's first automatic
  `needs_component_regeneration` emissions (source design-discovery-agent) as
  other sites' discovery passes hit their 8 corrupted components.
- Cross-workstream note (from leopardess, see PLAN I0 status): this deploy
  presumably also carries their dynamic_adapter routing change
  (logo/illustration/infographic → Banana); their workstream owns verifying
  it, plus the IMAGE_BUCKET env fix that gates our B6/I1 logo work.
- Runbook B8 marked done. Remaining human queue: B4 (data-source key, at I4),
  B5 (budget sign-off before scale-out), B6 (logo approve/lock, at I1),
  B7 (layout-gap decision), B9 (reaper cadence — recommended).

## Turn 16 — 2026-07-10 — Both investigation threads closed (corruption source + no-op complete)

**THREAD B — corruption source: CLOSED, nothing to build.**
- Forensics: `component_versions` only holds today's pre-regen snapshots → the
  corrupted templates were written by a path that predates version tracking.
  Provenance columns settle it: all 14 corrupted components were
  `created_from='generated'` between **2026-03-31 and 2026-04-13** — the early
  component-generation era.
- The modern writer already guards against exactly this:
  `store_generated_component` Layer-1 pre-store validation REJECTS templates
  containing `<no value>` ("Go template render output mistakenly stored as
  source") AND templates whose variables don't match the schema. Verified in
  code (~lines 282–308).
- So: historical corruption, modern writers safe, retroactive cleanup handled
  by the now-live `component_template_corrupted` bridge check. No action.

**THREAD A — "no-op complete": ROOT-CAUSED AND FIXED (live, no deploy).**
- Full chain for the 07:36 about run (orch `535e3ce5…`): the content-writer
  CHILD failed at REPLY time — Kafka `topic partition not found` on the
  parent's per-job responses topic (transient infra flake) → parent's
  `call_content_writer` got CHILD_ORCHESTRATION_FAILED → step-level
  `error_step: complete_error` → **complete_error is a SUCCESS-labelled
  `complete_workflow`** → orchestration COMPLETED → dispatcher's
  `complete_work_item` stamped the item 'complete', no error recorded.
- The codebase already half-knew: CompleteWorkItemAction's guard comment
  documents that complete_error is success-labelled, and the established
  pattern is "flag the item BEFORE completing" (mark_needs_review,
  mark_no_sections). Real errors just never flagged.
- **Fix applied** (`SQL_2026-07-10_pagebuild_mark_item_failed.sql`, backup +
  snapshot, verified): new `mark_item_failed` step (update_work_item_status →
  'failed', attempt-counted, skip_if_missing) inserted ahead of
  complete_error; all 8 step-level error pointers repointed through it.
  Workflow-config-only — live for the next page build. Failed page builds are
  now VISIBLE instead of silently complete.
- **Residual (infra, noted not fixed):** the Kafka per-job response-topic
  partition race that triggered the original failure remains a transient
  flake — now surfaced as failed items rather than swallowed. Topic-lifecycle
  hardening is dispatch/infra scope (sits alongside B9).

**Bridge-check watch:** no auto-emissions yet (other sites' discovery passes
haven't cycled since registration — expected; the watch continues).

**Workstream state: I0 threads done. Next: Phase I1 (brand consistency layer)
— user has signalled to proceed after these threads.**

## Turn 17 — 2026-07-10 — Phase I1 built: style guide + lock guard (committed, needs deploy)

**Prerequisite check:** `IMAGE_BUCKET` is still NOT set on the agent-chassis
deployment (only AGENT_IMAGE_* present) — the leopardess fix has not landed.
Logo FILE deployment to git remains broken; the I1 logo approve-and-lock flow
waits on it (B6 note stands). Asset locking columns (locked_at/locked_by/
lock_type) confirmed present from Phase 2A.

**Built (one commit, rides the next chassis deploy):**
1. `imagery_style_guide.go` — new site_specs aspect `imagery_style_guide`
   {palette, medium, mood, avoid, reference_asset_keys}. `directionForKind`
   gates per kind: photographic kinds get medium+mood+palette; icons palette
   only; logos nothing (preserves the 2026-05-20 contamination lesson).
   Nil-safe; unit-tested.
2. `generate_image` integration — guide direction SUPERSEDES the free-text
   design_intent.imagery_direction when present (one brand voice, no double
   prepend; fallback preserved for guide-less sites). `avoid` terms append to
   the NEGATIVE prompt. `reference_asset_keys` resolve to stable `s3://` URIs
   (presignedURLToS3URI — anchors outlive the 7-day presign expiry) and flow
   to Banana as style anchors for photographic kinds when the caller passed
   none.
3. **D5 lock guard** — the assets upsert now has `WHERE assets.locked_at IS
   NULL`; storing over a locked (approved) asset is refused and reported, not
   silently applied. First enforcement point of logo permanence.
4. Seeded robot-hands.com's `imagery_style_guide` (live in DB now; SQL
   artifact saved) — distilled from its rich design_intent: charcoal/electric-
   blue palette, industrial-photography medium, engineered mood, anchored to
   `hero_canonical` + `hero_home`.

**I1 acceptance test (post-deploy):** trigger two generations of different
kinds (e.g. an illustration + an icon) on robot-hands; verify the log lines
`+style_guide`, the icon getting palette-only, reference anchors on the
Banana call; then lock an asset and confirm store_asset refuses to overwrite.
Logo approve-and-lock (B6) additionally waits on IMAGE_BUCKET.

## Turn 18 — 2026-07-10 — I1 deployed; B7 root-caused and fixed; discovery cycle triggered

**I1 deploy confirmed** (fresh chassis pod): style guide + lock guard now
active in production. robot-hands' seeded guide is live for the next
generation. Acceptance generations still to run.

**B7 resolved (user decision: no brochure fallback). "We have been here
before" confirmed — twice over:**
- The item's spec held the real cause: `"reason": "fallback — no
  classification tags"`, `site_tags: []`. The scheme-aware matcher
  (resolve_composition_layout → resolveLayoutByTags, doc 027) scores
  `classification.industry_tags` against `layouts.industry_tags`;
  robot-hands' old-format classification HAS no industry_tags (same
  old-format gap as the missing news flag fixed during the rebuild).
- The library ALREADY contains the right layout: **tool-portal-dark**
  (scheme=dark, category=interactive, tags interactive-platform/tools/
  tool-portal/calculators/technical-reference/professional-dark) — the
  needs_new_layout_candidate → human → library-growth loop had already run
  once before and produced it.
- Fix (`SQL_2026-07-10_b7_layout_fix.sql`): classification superseded with
  nine matching industry_tags; review item closed wont_fix; fresh
  `needs_composition` queued (site-design-planner re-resolves; expected
  match tool-portal-dark, then site re-render through the design pipeline).

**Discovery cycle triggered** for robot-hands.com: the documented
`system.intake` pattern is STALE (topic doesn't exist) — the working
mechanism is the kcat trigger script pattern (033_rerender_pages_trigger.sh):
`action=orchestrate` envelope to `system.agent.generic.requests` with
`config.agent_type: improvement-loop` + input_data {site_id, domain}.
Trigger sent (correlation 5b43181e…). The improvement loop runs quality +
design + completeness discovery, triages, and dispatches — this exercises
BOTH new checks (component_template_corrupted → bridge regens for the
corrupted components still used on robot-hands' tool-guide pages;
image_source_unsatisfiable → flags) plus picks up the queued composition.
Background monitor watching: orchestrations, bridge emissions, unsat flags,
and the B7 composition outcome (chosen layout).

**Next:** on monitor results — verify tool-portal-dark chosen + bridge
emissions; then I1 acceptance generations (illustration + icon with
+style_guide in logs); consider triggering improvement loops for
finetuning.uk / vonc.com / leopardessconsulting.co.uk to sweep their
corrupted components.

## Turn 19 — 2026-07-10 — PRE-DEPLOY BASELINE (user deploying a new build mid-run)

Snapshot taken just before a user-initiated chassis deploy, so we can tell
deploy disruption from real failures:

**Already SUCCEEDED before the deploy (durable — in DB):**
- Discovery cycle ran clean: quality + design + completeness discovery agents
  all COMPLETED.
- ✅ **Bridge check works end-to-end via discovery** — `component_template_
  corrupted` auto-emitted a `needs_component_regeneration` for
  **tool-guide-intro** (created_by=design-discovery-agent, status detected).
  This is the proof the fleet-wide self-heal fires without manual queueing.
- `image_source_unsatisfiable`: 0 flags (expected — heroes all resolve now).

**In-flight / not-yet-done at deploy time (may be disrupted → zombie claims):**
- B7 `needs_composition_b7_fix`: still **triaged** (site-design-planner hasn't
  run it yet); robot-hands layout still shows `brochure-formal` — the re-
  compose to tool-portal-dark has NOT happened yet. VERIFY POST-DEPLOY.
- 1 claimed item: `rerender_component_regen:learning-center-index` (claimed
  15:08 — likely already a zombie).
- The tool-guide-intro regen (detected) not yet dispatched.

**Post-deploy recovery checklist:**
1. Clear zombie claims (>10min) — the deploy will orphan any claimed items.
2. Re-confirm the discovery findings above are intact (they're DB rows, should
   survive).
3. Verify B7 composition runs → layout becomes tool-portal-dark → site
   re-renders; if the trigger orchestration died mid-flight, re-trigger the
   improvement loop (kcat pattern, Turn 18).
4. The DB-reading monitors survive a chassis restart (postgres isn't
   redeployed); they may just time out — re-launch if needed.

## Turn 20 — 2026-07-10 — Pre-deploy-2 baseline: bridge loop FULLY proven; B7 needs the theme-fork path

Captured just before the user's second deploy of the evening:
- ✅ **Bridge self-heal proven END-TO-END**: tool-guide-intro auto-detected by
  discovery → needs_component_regeneration auto-emitted → triaged → dispatched
  → component-creator regenerated → template verified clean (no <no value>,
  real vars). No human involvement at any step. Fleet-remaining corrupted: 7.
- ❌→📋 **B7 composition failed with a DESIGN-LEVEL refusal** (not deploy
  damage): `install_site_composition: site already has style_collection_id;
  re-resolve not supported` (3 attempts). Composition install is write-once;
  changing an EXISTING site's layout must go through the theme-fork path —
  `fork_theme_from_site` (shares resolveLayoutByTags; see 003 CSS Theme
  Template Contract forking rules; fork_theme_composition.go). NEXT ACTION:
  re-do the B7 layout change via the fork route post-deploy. Classification
  tags fix stands (the matcher will now score correctly whichever path runs).
- Re-triggered improvement loop had triaged the stranded backlog; dispatch
  working through ~84 pending design-audit items (deactivated components,
  stale site components, CTA/colour fixes) — robot-hands general quality
  improving as a side effect.

## Turn 21 — 2026-07-10 — Deploy-3 recovered; B7 executed via the 025 FK-swap pattern

**Post-deploy-3 recovery:** new pod up; 2 zombies cleared (archetype-grid
empty_section, llm-cost-calculator needs_page); 82 triaged items dispatching.

**B7 route settled after investigating the theme system:** there is NO runtime
re-compose path by design — `install_site_composition` refuses when a
collection exists, and `fork_theme_from_site`'s install mode was REMOVED
2026-04-19 ("Composition install moved to site-design-planner"). The sanctioned
precedent for changing an EXISTING site's composition is the 025 migration
pattern: a targeted FK update. robot-hands' css_theme (theme-robot-hands-com)
is site-specific (origin=adopted, source_site_id=self) so the swap is
site-local; tool-portal-dark declares no default header/footer (both NULL, per
025 convention) so no collection relinking was needed.

**Executed:**
1. `SQL_2026-07-10_b7_layout_swap.sql` — css_themes.layout_id: brochure-formal
   → tool-portal-dark (backup + verify); failed needs_composition item closed
   with the outcome note.
2. Triggered webdesign-agent (kcat orchestrate) — renders CSS from the updated
   composition FKs (render_css_from_spec: palette=spec-wins core slots, layout
   structure tokens) and commits assets/css/styles.css to git.
3. Monitor auto-triggers rerender-pages with refresh_site_components=true once
   the CSS lands, then watches the page_rerender items drain.

**FUTURE flag (structural, for the plan's open list):** a proper runtime
re-compose path (site-design-planner "re-resolve" mode with lineage) is
missing by deliberate deferral; today's FK-swap is the documented workaround.

## Turn 22 — 2026-07-10 (evening) — B7 chain landed; docs + handoff synced; drain in progress

**B7 outcome (user-visible result already live):**
- webdesign-agent COMPLETED: CSS rendered from the swapped composition
  (tool-portal-dark structure under robot-hands' own charcoal/electric-blue
  palette) and committed to git. Since the layout change is substantially a
  CSS change, **robot-hands.com already renders as the dark tool-portal** —
  A3 eyeball is available NOW.
- Timing subtlety understood: the 36 page_rerender items were queued by the
  improvement loop BEFORE the swap; my rerender-pages trigger dedup'd against
  them (correct — one queue). They refresh the HTML side and drain through
  dispatch behind the design-audit backlog (~75 open items total).
- Monitoring-bug note for the record: one earlier monitor misread transient
  psql failures as "0 remaining" — corrected monitors distinguish query
  failure from empty queue.

**I1 acceptance happening organically:** discovery queued 3 needs_imagery
items; they will be the first generations through the live style guide. The
watch checks their assets' origin_prompt for guide fingerprints ("industrial
photography", "charcoal") when they complete = direct acceptance evidence.

**Docs synced this turn:** notes (this), PLAN (I1/B7 closing state), RUNBOOK
(B7 closed, A3 eyeball requested), and a NEW `HANDOFF_imagery_best_in_class.md`
for continuing in a fresh chat — keep it updated at every turn alongside these
notes.

## Turn 23 — 2026-07-10 — Logo thread executed: header fix coded + fresh logo candidate queued

**Facts established (live site):** no logo file deployed
(`/assets/images/logo.png` and `.jpg` both 404 — the May logo asset was
never committed to git); the live header renders the text mark. The header
component TEMPLATE already supports `{{if .logo_url}}<img …>` — only the
data was missing: `render_site_components` → `loadSiteDataFull` read the
never-populated `sites.logo_url` column and never consulted the imagery
plan.

**Fix coded (commit b00c150b — needs next chassis deploy):**
`loadSiteDataFull` now resolves the site-scope logo from
site_plan_imagery→assets via `storage.DeployedWebPath` (committed git path,
never assets.url), keeping `sites.logo_url` as legacy fallback; logger
threaded through both callers. Build + vet clean.

**B6 candidate queued:** fresh `needs_imagery:site:-:logo` item (triaged,
priority 70, brand_update=true, using the plan's logo prompt) — regenerates
via the new Banana routing and deploys through the spawned asset-deployer
chain. Sequencing: generate → USER APPROVES (A3/B6) → set assets.locked_at
(lock guard then enforces permanence) → deploy header fix → re-render
header → logo live.

**Drain progress at turn end:** dispatch healthy (~1 item/min; 62 open and
falling); 3 needs_imagery + 36 rerenders still queued behind by priority —
the I1 acceptance evidence arrives when those imagery items generate.

## Turn 24 — 2026-07-11 — Drain done; I1 acceptance PROVEN (icons); logo reality clarified; B6 decision teed up

**Auth note:** kubectl token expired overnight (A1 ritual); user re-authed.
The "plateau at 43" from the prior watch was the watch's own queries
degrading as auth lapsed — NOT a stall. The queue fully drained overnight.

**✅ I1 GENERATION ACCEPTANCE — PROVEN.** Three section icons generated
overnight (icon_vendor_neutral, icon_cross_technology, icon_integrated_tools)
show textbook style-guide gating in `assets.origin_prompt`:
- palette present ("charcoal") ✓, medium ABSENT ("industrial photography"
  not prepended) ✓ — exactly `directionForKind`'s icon rule (palette-only),
- routed to `banana/gemini-3-pro-image-preview` ✓ (kind routing).
The style guide reaches the model and gates correctly per kind. (Full
photographic-voice on a hero is the same default code path; a fresh hero/
illustration would show medium+mood+palette — deferred, not required.)

**Logo reality (corrected — I mis-tested the extension AGAIN):**
- `robot-hands.com/assets/images/logo.jpg` serves **200** (I'd checked .png).
  A real logo IS deployed.
- BUT the header renders the text mark because it reads the empty
  `sites.logo_url`; the Turn-23 header fix (resolve from plan imagery via
  DeployedWebPath) closes this — and derives `logo.jpg` correctly because it
  uses the asset's actual purpose. NEEDS NEXT CHASSIS DEPLOY.
- The logo asset is the **May-8 generation** (origin_model banana, but
  purpose='hero' — the old purpose-field bug; hero-dimensioned; predates the
  style guide). My queued B6 item completed as a NO-OP (image-build-handler
  skipped — asset already existed).

**Cleanup:** closed 2 stale `failed` needs_imagery rows (hero_home,
brand_hero_canonical) from 2026-05-17 — pre-SDXL-snap-fix residue; their
assets exist and render.

**B6 DECISION FOR USER (blocking the logo close-out):** approve the existing
May-8 logo as-is, OR force a fresh regeneration (deactivate current asset →
re-queue) for a correctly-sized (400×400 png), style-guide-aware,
best-in-class logo. Recommend fresh regen since B6/G1 is about a PERMANENT
best-in-class mark. Eyeball URL: https://robot-hands.com/assets/images/logo.jpg

## Turn 25 — 2026-07-11 — Favicon + OG card built (I1 tail); I1 now feature-complete pending deploy

**Built (commit — needs next chassis deploy):**
- `derive_brand_head_assets` action: fetches the locked logo bytes from S3,
  resizes → `favicon.png` (64×64), composes the logo centred on a
  brand-palette background → `og-card.png` (1200×630), commits both to the
  site repo, records provenance asset rows (origin_model='derived-from-logo').
  Deterministic image processing (nfnt/resize + image/draw) — no LLM. Must
  run in a storage-enabled agent (asset-deployer). Registered.
- `render_site_components` injects favicon + OG/Twitter-card `<head>` tags at
  render time — fleet-wide, idempotent (skips if favicon/og already present),
  graceful fallback to the logo before favicon.png exists, attrs escaped.
  Closes the head-metadata gap (head templates had NO favicon/og markup).
- `ImagePurposes`: favicon (64×64 png), og_card (1200×630 png).
- Unit tests: parseHexColour (incl. gradient→neutral fallback) + injection
  (tags present, before </head>, idempotent, escaped). Pass.

**Design choices:** OG card = logo on a solid palette colour
(header_bg→footer_bg→background→primary→dark default; gradients rejected —
OG needs a solid). Favicon = square resize of the logo. Both literally
"derived from the logo" per G8; the head injection is deterministic so no
per-site head-template regeneration is needed.

**I1 status:** with the style guide (Turn 17), logo lock (Turn 24), header
logo resolution (Turn 23) and now favicon/OG, **Phase I1 is feature-complete
pending deploy.** Post-deploy activation for robot-hands: (1) run
`derive_brand_head_assets` (site_id+domain) via asset-deployer to commit
favicon.png + og-card.png; (2) `rerender-pages` refresh_site_components=true
so the head + header pick up the new tags and the locked logo.

**Next:** post-deploy activation + eyeball; then Phase I2 (sprite-sheet
bullets — design locked in CONTEXT_PACK_imagery_sprite_sheet.md) or I3 (card
imagery / Lane B). Leaning I2 (smaller, self-contained, immediately visible).

## Turn 26 — 2026-07-11 — I1 code deployed; favicon/OG activation wired through dispatch

**Deploy confirmed:** v1.0.1107 pod started 18:27 UTC, AFTER both the header
fix (b00c150b, 07-10) and favicon/OG (d3f73a72, 18:06) — both live. (Corrected
an earlier pod-age miscalculation.)

**Activation — how to run the derivation (learned the hard way, RECORD THIS):**
`derive_brand_head_assets` needs a STORAGE-ENABLED (spawned) pod. Two attempts:
1. ❌ Hand-rolled inline parent workflow (kcat, spawn_agent+call_agent to
   asset-deployer): the spawned pod ran its workflow on INIT (no payload),
   check_mode saw no mode, idled out; parent stuck AWAITING_RESPONSES at
   spawn_ad — topic-wiring is finicky for hand-crafted spawn+call. Reset the
   stuck orch to FAILED.
2. ✅ **Route through build-dispatch-loop** (battle-tested spawn wiring).
   asset-deployer is already a valid dispatch handler (undeployed_asset).
   - Migration `SQL_2026-07-11_asset_deployer_brand_head_mode.sql`: added a
     `check_mode` start conditional + `derive_head_assets` step. Condition
     made robust to the DISPATCH input contract (handlers receive
     `input_data.spec` + `input_data.site_id` + `input_data.domain`):
     `input_data.spec.mode == "brand_head" OR input_data.mode == "brand_head"`.
   - Work item `needs_brand_head_assets:robot-hands` (handler asset-deployer,
     spec {mode:brand_head}) → dispatch spawns asset-deployer with storage env
     → derive_head_assets runs. Monitoring for favicon.png/og-card.png (200).

**Reusable outcome:** any site gets brand-head assets by queuing a
`needs_brand_head_assets` item routed to asset-deployer. Good candidate to
auto-emit after a logo locks (future discovery check).

## Turn 27 — 2026-07-11 — ✅ I1 COMPLETE + LIVE-VERIFIED; I2 scoped

**Favicon/OG derivation ran clean** (dispatch path): item complete first
attempt, `favicon.png` + `og-card.png` both serve 200, provenance rows
origin_model='derived-from-logo'.

**I1 ACTIVATION VERIFIED LIVE on robot-hands.com** (curl of the served index +
site_components rows): head re-rendered 20:22 with `og:image` + `rel="icon"`
+ `twitter:card`; header carries `logo-img` (the locked logo). So the full
brand layer is live: locked logo in header, favicon in tab, OG card on link
previews. (The background monitor's `head[og|icon]=` blanks were a query
quoting bug, not a gap — direct check confirmed.) 34 page_rerenders still
draining refresh page bodies with the updated head/header.

**PHASE I1 DONE:** style guide (per-kind gating proven on icons) + logo
generate/approve/lock (D5) + header logo resolution + favicon/OG derived from
logo. The G1/G8 brand-consistency layer is complete end-to-end on the testbed.

**PHASE I2 SCOPED** (`SCOPE_I2_sprite_sheets.md`). Confirmed the locked design
against fresh schema; found + resolved ONE deviation: `css_snippets` is a
GLOBAL library, not per-site, so the sprite stylesheet ships as a separate
committed `/assets/css/sprites.css` + a head `<link>` — reusing the Turn-25/26
head-injection + git-commit patterns (the biggest reuse win). Build breakdown
(8 small pieces), phases I2.0–I2.4 with human eyeball gates, cell-alignment as
the key risk. Two decisions flagged for user: first surface (recommend list
bullets) + 3×3 cell vocabulary for robot-hands.

## Turn 28 — 2026-07-12 — I2 decisions locked; CSS delivery re-verified; I2.0 built + gated

**User decisions:** first surface = LIST BULLETS; 3×3 starter vocabulary =
check, gauge, gripper, cog, chart, download, arrow, info, warning.

**CSS delivery re-check (user asked to be sure) — decision CONFIRMED,
stronger than before:** css_snippets is definitively global (no site_id;
matched by component overlap → site-specific CSS would leak fleet-wide).
And the separate-committed-file shape is the HOUSE PATTERN: the live site
already loads per-site bundles `/assets/css/styles.css` AND
`/assets/js/snippets.js` (render_js_snippets_for_site commits the latter —
the exact analogue). Rejected alternatives: append-into-styles.css (couples
sprite refresh to a webdesign-agent LLM run and risks silent drops on
theme re-render); inline <style> (duplicated bytes × 33 pages, full
re-render on any grid change). `/assets/css/sprites.css` + head <link> stands.

**I2.0 BUILT + GATE PASSED:**
- `SQL_2026-07-12_add_sprite_sheet_kind.sql` applied: chk_kind now includes
  sprite_sheet (idempotent, backed up).
- Go (rides next deploy): validImageryKinds + sprite_sheet; dynamic_adapter
  routes sprite_sheet → Banana; ImagePurposes sprite_sheet 768×768 png.
- **Gate:** hand-inserted sprite_sheet plan row passes chk_kind (insert+
  rollback proof). Bonus learning: `chk_source` allows only
  llm|classifier|manual|adoption — seed rows must use source='manual'.
- Parallel-session note: storage pkg gained exported `PresignedURLToS3URI`
  (leopardess); no conflict with our private actions-pkg copy.

**Next (I2.1, needs the deploy for the adapter route):** hand-seed the
robot-hands sprite_sheet plan row (source='manual', style_hints
{rows:3,cols:3,cell_names:[check,gauge,gripper,cog,chart,download,arrow,
info,warning], style:"flat single-colour line glyphs, light grey on flat
#1a1a2e, one glyph per cell, reading order"}) + queue needs_imagery →
generate → EYEBALL GATE (assign real cell meanings after seeing the sheet).

## Turn 29 — 2026-07-12 — Contamination gap caught pre-generation; sprite row seeded; ONE more deploy needed

**Gap caught (would have wasted the eyeball gate):** the per-kind gating
functions predate sprite_sheet, so it fell into the PHOTOGRAPHIC default —
`directionAppliesToKind` would prepend the industrial-photography voice AND
attach photographic reference anchors to a flat glyph grid: the 2026-05-20
contamination failure, precisely. **Fixed (commit 4629aa17):** sprite_sheet
joins the flat-vector class in both `directionAppliesToKind` (no free-text
direction, no reference anchors) and `styleGuide.directionForKind`
(palette-only). Test extended; green. **Generation MUST wait for the next
chassis deploy** — the current image routes sprite_sheet→Banana (I2.0 ✓) but
still carries the contaminating gating.

**Seeded (live in DB):** `sprite_sheet_main` on robot-hands' current plan
(SQL_2026-07-12_seed_robothands_sprite_sheet.sql) — 3×3 @ 768², the locked
vocabulary (check, gauge, gripper, cog, chart, download, arrow, info,
warning), flat light-grey line glyphs on #1a1a2e, `cell_names_verified:false`
(true map assigned at the eyeball gate).

**Flow verified for the new kind:** `check_unfulfilled_imagery_plan` will
emit it (site scope → priority 75/high); `BuildSpec` sets purpose=kind →
ImagePurposes 768×768 png → deploys as
`/assets/images/sprite-sheet-main.png`. All untouched machinery.
**Caveat:** if a discovery pass runs on robot-hands BEFORE the gating deploy,
it would generate a contaminated sheet — low risk (loop idle unless
triggered); if it happens, eyeball-reject + regenerate post-deploy.

**Post-deploy runsheet (I2.1):** deploy → queue/let discovery emit the
needs_imagery for sprite_sheet_main → generate → EYEBALL GATE with user
(assign true cell meanings; write back to style_hints, set
cell_names_verified:true) → then I2.2 (sprites.css emit action).

## Turn 30 — 2026-07-12 — Full doc sync; handoff rewritten as the detailed fresh-chat entry point

- PLAN: I1 status block replaced with the ✅ COMPLETE/live-verified record;
  I2 gained a dated status block (decisions, delivery revision, I2.0 done,
  I2.1 deploy-blocked, I2.2 next).
- RUNBOOK: added **B10** (next deploy carries the sprite gating fix — tell
  the agent) and **B11** (the sprite eyeball gate: verify clean grid, then
  dictate the TRUE cell meanings in reading order).
- HANDOFF: **rewritten in full** as the detailed cold-start document —
  state per phase, the five-place new-kind checklist, dispatch input
  contract, trigger patterns (incl. the do-NOT-hand-roll-spawn+call lesson),
  zombie unstick, seeds/source constraint, open threads, ordered next
  actions. Standing rule: update it EVERY turn alongside these notes.
- Next work: I2.2 sprites.css emit action (buildable pre-deploy), then the
  B10/B11 generation + gate.

## Turn 31 — 2026-07-12 — Sheet generated (near-perfect); deploy-purpose bug found + fixed; regen in flight

**B10 deploy verified** (pod 18:18 UTC > gating commit 15:19). Queued
`needs_imagery:site:-:sprite_sheet_main` (spec mirrored from the logo-item
precedent + style_hints riding along).

**GENERATION: excellent.** Complete on attempt 1 via
banana/gemini-3-pro-image-preview. Contamination-gating fix PROVEN in
`origin_prompt`: "Colour palette" present, "industrial photography" ABSENT.
I downloaded and visually inspected the sheet (agents can Read PNGs — worth
remembering as a pre-gate): **all NINE glyphs exactly as requested, in exact
reading order** (check, gauge, gripper, cog, chart, download, arrow, info,
warning), uniform stroke, light grey on charcoal, dividers visible, no text.
The cell-alignment risk (the phase's headline risk) did not materialise.
Saved copy: session scratchpad `sprite_sheet_main.png`.

**ERROR 1 — user's eyeball gate blocked by a deploy failure.** The B11
answer came back as a pasted B2 error: `objectKey robot-hands.com/assets/
images/sprite-sheet-main.png → 404 NoSuchKey`. Investigation chain:
1. Work-item result: deploy leg "success" — but committed
   **sprite-sheet-main.JPG**, commit message "Deploy **hero** image".
2. Live file: 200 at .jpg, **900×900 JPEG** = exactly hero config
   (1600×900 thumbnail of a square source). So deploy ran with
   purpose='hero' despite the store leg recording 'sprite_sheet'.
3. Child forensics (`orchestration_states.initial_request_data`): the
   asset-deployer child **RECEIVED input_data.purpose='sprite_sheet'** —
   the call mapping was fine.
4. Root cause in `deploy_image_asset`: `ExtractActionInputs` falls back to
   the AGGRESSIVE recursive `ExtractFields` search per field name; for
   `purpose` it matched a stale value elsewhere in collected_data (siblings
   s3_uri/asset_key happened to resolve correctly) → empty/hero → the
   action's `purpose="hero"` default (line ~85). This is the exact hazard
   the extraction code's own Strategy-0 comment warns about.

**SOLUTION (workflow-only, live immediately, no deploy):**
`SQL_2026-07-12_asset_deployer_explicit_paths.sql` — asset-deployer's
`deploy_asset` step config gains explicit Strategy-0 dot-paths
(`s3_uri/purpose/domain/asset_key` → `input_data.*`); explicit paths are
resolved FIRST and win over the aggressive search. Backup + snapshot +
verify. **Standing lesson: any step feeding ExtractActionInputs-based
actions should declare explicit dot-paths, never rely on the search.**

**Recorded, not fixed (latent siblings):**
- Dispatch-shape gap: items dispatched by build-dispatch-loop carry payload
  under `input_data.spec.*` — the new explicit paths miss that shape and
  Strategy-1 search still applies there (undeployed_asset flow). Fix when it
  misbehaves or in a hardening pass.
- Wider implication: EVERY spawned deploy may have silently used hero
  dimensions (masked because heroes dominate) — e.g. the May icons' deployed
  files are worth a dimensions check someday.
- Clutter: the wrong `sprite-sheet-main.jpg` (900² hero-config) remains in
  the site repo; nothing references it; remove in a cleanup pass.

**CHOICE — regen rather than redeploy-only:** the eyeball gate never
actually completed (the user's gate answer was the error paste), so a fresh
generation costs one cheap Banana call and needs a gate either way; reset
the item (status/result/error cleared) → full chain re-runs through the
FIXED deploy path. Monitor watching: item complete + `/assets/images/
sprite-sheet-main.png` 200 + file verifies as **768×768 PNG** (the geometry
sprites.css slicing requires). Then B11 eyeball gate re-runs with a working
deployed URL.

## Turn 32 — 2026-07-13 — Two more reset/dispatch gotchas found before the sheet redeployed

**ERROR 2 (state-machine race, a live TODO-item sighting):** my first
post-fix reset of the sprite item got RE-STAMPED `complete` by the ORIGINAL
run's late-arriving completion (the terminal orchestration's tail wrote
status after my reset). Confirmed no in-flight image orchestrations, reset
again. This is direct evidence for the user's reaper/state-machine TODO
(6/10/11) — a finished run can overwrite a freshly-reset item.

**ERROR 3 (the real blocker — RESET GOTCHA):** after resetting, the item sat
`triaged` 24 min with ZERO claims — and it was the ONLY triaged item in the
whole fleet, so not contention. Cause: **`attempt_count` had reached 3/3**
(max_attempts) — my earlier resets cleared status/result/error but NOT
attempt_count, so the accumulated attempts (bad-jpg run + re-stamped run +
claim) capped it. A capped item is excluded from `find_dispatchable_site`,
so the trigger found NO dispatchable work on robot-hands and never spawned
build-dispatch-loop → zero dispatch orchestrations. **Dispatch was correctly
IDLE, not broken** (I initially suspected dead infra — wrong). Fix: reset
`attempt_count=0` alongside status. **STANDING LESSON: when re-driving a work
item, ALWAYS also reset `attempt_count=0` (and clear claim metadata); status
alone is not enough.** Now dispatchable; monitor watching for a correct
768×768 PNG deploy through the Strategy-0-fixed path.

## Turn 33 — 2026-07-13 — REAL deploy root cause: Kafka message-size; sprite → jpg (revises the plan)

**Chased the "completed but no PNG" through: (a) not a skip — the full chain
DID run post-reset (image-generator COMPLETED, then asset-deployer FAILED —
UTC/BST hid the timestamps); (b) the asset-deployer error:**
`step deploy_asset failed: ... failed to write message to kafka: [10]
Message Size Too Large`. The git-commit path base64-encodes the image into a
Kafka message; a lossless 768² PNG of a detailed glyph grid exceeds the
broker max.message.bytes. The FIRST deploy only "succeeded" because the
purpose bug sent it as JPG (small).

**This invalidates the plan's PNG choice on TWO grounds:** (1) Kafka msg
limit; (2) a lossless PNG of this content can't meet the ≤80KB sprite budget
(G7) — source was 439KB. **DECISION: sprite_sheet → JPG q88, 768×768**
(commit 23fe6e81). Legibility at bullet display (16–24px) is unaffected.
SCOPE_I2 + notes updated (png→jpg everywhere; deployed file is now
`sprite-sheet-main.jpg`, which DeployedWebPath derives automatically).

**So the full sprite deploy failure was a THREE-bug stack, now all fixed:**
1. purpose→hero via aggressive input search → Strategy-0 explicit paths
   (Turn 31, workflow-only, live).
2. re-drive left attempt_count capped + state-machine re-stamp race
   (Turn 32; reset attempt_count=0).
3. lossless PNG > Kafka msg limit + > budget → jpg (this turn; needs deploy).

**NEEDS ONE MORE CHASSIS DEPLOY** (the jpg ImagePurposes change is Go). Then:
delete the stale 900² sprite-sheet-main.jpg + deactivate the sprite asset →
re-drive the needs_imagery (attempt_count=0) → generate + deploy a clean 768²
jpg → EYEBALL GATE. RUNBOOK B10 refreshed for this deploy.

## Turn 34 — 2026-07-13 — Sheet deployed clean + gate PASSED; I2.2/I2.3 built

**✅ Sprite sheet LIVE + gate confirmed.** Re-drive (all 3 fixes in) deployed
a clean 768×768 JPEG, **75,745 bytes** (under the 80KB budget), serving 200 at
`/assets/images/sprite-sheet-main.jpg`. Visual: perfect — nine flat light-grey
glyphs on charcoal, exact reading order. **USER B11 GATE: confirmed as read**
(check, gauge, gripper, cog, chart, download, arrow, info, warning). Wrote
`cell_names_verified:true` + verified_by into the plan row's style_hints.

**I2.2 + I2.3 BUILT (commit — Go rides next deploy; migration live now):**
- `emit_sprite_css` action: pure CSS `background-position` slicing computed
  from the verified grid (rows/cols/cell_names) at bullet display size
  (T=20px → sheet drawn 60×60, cells at reading-order offsets). Emits
  `.sprite` base + `.sprite-<name>` (inline/icon/nav) + themed
  `ul.sprite-list li::before` bullets with default glyph (cell 0 = check) and
  per-item `li.sprite-b-<name>` overrides. Commits `/assets/css/sprites.css`
  (base64 via the proven git-adapter path). GUARD: only emits when
  cell_names_verified=true. Geometry unit-tested.
- `render_site_components`: injects `<link rel=stylesheet sprites.css>` into
  `<head>`, GUARDED on an active sprite_sheet asset (no 404 link elsewhere).
- asset-deployer `sprite_css` mode (migration LIVE): check_mode →
  check_sprite_mode → emit_sprite_css_step. Alongside brand_head. Reusable
  fleet-wide via a `needs_sprite_css` work item.

**Next:** ONE deploy carries emit_sprite_css + the head-link Go. Then:
dispatch `needs_sprite_css` (asset-deployer, spec.mode='sprite_css') →
sprites.css commits → rerender-pages refresh_site_components (head gets the
link) → I2.3 wire `class="sprite-list"` on ONE robot-hands section's <ul> →
live gate (bullets readable, one ≤80KB download). Then I2.4 fulfilment check.

## Turn 35 — 2026-07-13 — Deploy verified correct (functionally); I2.2 CSS LIVE; head-link re-render triggered

**User asked to verify the deploy.** Source ancestry: emit_sprite_css
(fe9f125c) IS an ancestor of the deployed v1.0.1114 (3406cd71) — but pod
started only ~11 min after that commit (tight build window), so I did the
DEFINITIVE functional test instead of trusting timing.

**✅ DEPLOY CONFIRMED CORRECT (functional):** dispatched needs_sprite_css →
asset-deployer ran the NEW `emit_sprite_css` action (COMPLETED:ok, item
complete attempt 0 — not "unknown action"), proving v1.0.1114 carries the
I2.2/I2.3 Go. `/assets/css/sprites.css` serves 200, **1,711 bytes**, content
exactly correct: `.sprite` base + 9 `.sprite-<name>` classes + themed
`ul.sprite-list` bullets (default cell 0 = check) + 9 `li.sprite-b-<name>`
overrides; all background-position offsets match the verified 3×3 geometry.
So **I2.2 is LIVE.**

**I2.3 in flight:** triggered rerender-pages (refresh_site_components=true) to
land the `<link rel=stylesheet sprites.css>` in the head — now that BOTH the
sprite asset AND sprites.css exist, the injection guard passes. This also
functionally tests the head-link half of the deploy. Monitoring.

**Remaining for I2:** confirm the head link in served HTML → wire
`class="sprite-list"` onto ONE robot-hands section's `<ul>` → live gate →
I2.4 fulfilment check. Then Phase I3.

## Turn 36 — 2026-07-14 — Bullets wired on a real list; LIVE GATE caught a CSS specificity bug in emit_sprite_css

**Wired I2.3's list (handoff step 3).** Surface chosen: the **Safety Factor
Selection Guidance** list on `/guides/tool-grip-force-friction-calculator-
guide.html` (4 items, reader-facing, glyphs semantically apt):
SF=1.5 → `sprite-b-info`, SF=2.0 → **no class (exercises the default)**,
SF=2.5–3.0 → `sprite-b-gauge`, SF>3.0 → `sprite-b-warning`.

**MECHANISM (new, important): page assembly reads `page_components.rendered_html`
DIRECTLY** (`rerender_single_page_action.go:383` getPageSections) — it does NOT
re-render from `content_data` + template. So a markup change must land in
`rendered_html` (the artifact that deploys); I wrote `content_data.result` too,
to keep source and artifact consistent. Backup: `bak_pc_sprite_20260714`.
Re-render dispatched as a `page_rerender` work item mirroring
CreateRerenderItemsAction's exact shape (no `spec.reason` ⇒ unscoped ⇒ plain
assemble). Completed first attempt; classes live in served HTML.

**LIVE GATE FOUND A REAL BUG (the whole point of the gate).** Rendered the live
page headless (chromium → PDF → 300dpi crop) and looked: **all four bullets
showed the SAME `check` glyph.** The overrides never applied. Cause is pure CSS
specificity in `buildSpriteCSS`:
- default `ul.sprite-list>li::before` = **(0,1,3)**
- override `li.sprite-b-gauge::before` = **(0,1,2)** ← always loses, and source
  order cannot rescue it.
So `background-position:0 0` (cell 0 = check) won on every item. The I2.2
geometry unit test passed throughout — it asserted the offsets, never that an
override could WIN the cascade. **Lesson: a CSS emitter needs a specificity
assertion, not just a geometry assertion.**

**FIX (Go, rides next deploy):** overrides now emit scoped under the list class —
`ul.sprite-list>li.sprite-b-<n>::before,ol.sprite-list>li.sprite-b-<n>::before`
= (0,2,3) > (0,1,3). Added `TestBuildSpriteCSS_overridesOutspecifyDefault`
(asserts scoped form present AND the bare losing form absent). Both tests pass.
**Fix PROVEN pre-deploy:** rebuilt sprites.css exactly as the corrected emitter
will produce it, rendered the real served list markup against it headless →
**four distinct glyphs (info, check, gauge, warning), legible at 20px.**

**Live state right now:** sheet + sprites.css + head `<link>` + the wired markup
are ALL live and correct; bullets render, `list-style` is suppressed, one 75,745B
sheet download. Only the per-item glyph *variety* awaits the deploy — every
bullet currently shows the default check.

**Cosmetic note for the gate:** each bullet paints its sheet cell *including the
charcoal cell background*, so a faint square tile sits behind each glyph (JPG has
no alpha — that was the Kafka-message-size/budget tradeoff of Turn 33). Reads as
a subtle box on the dark theme. Worth a user call: accept, or revisit with a
transparent-capable format under the size budget.

**Next:** on the next chassis deploy → re-dispatch `needs_sprite_css`
(attempt_count 0) to regenerate sprites.css → the four glyphs appear with no
markup change → user LIVE GATE. Then I2.4 fulfilment check closes I2.

## Turn 37 — 2026-07-14 — User choices recorded (D9, D10); empty product pages spun out to their own handoff

**USER DECISIONS (both now in PLAN §4 as D9/D10):**
- **D9 — cosmetic: ACCEPT the baked-in cell background.** Each bullet paints its
  sheet cell including the charcoal backdrop, so a faint square tile sits behind
  every glyph. Accepted as-is; no format change. (JPG has no alpha — the Turn 33
  trade: a lossless PNG blew both the Kafka commit message-size limit AND the
  ≤80KB budget.) Revisit only if a transparent format ever fits the budget.
- **D10 — durability: BUILD the container opt-in, sequenced AFTER the live gate
  and I2.4** (now phase **I2.5**). `emit_sprite_css` will also emit
  `.sprite-bullets ul>li::before` (same geometry, specificity-safe per the Turn 36
  bug), and the class goes on a component wrapper (article-body's
  `.article-body__content` first). Content lists then theme themselves — no markup
  edits, regen-proof, fleet-reusable. The Turn 36 hand-wired classes become
  redundant but harmless.

**SPUN OUT — empty product pages (NOT an imagery bug; user will fix in another
chat).** Investigated the `/entities/gripper-detail.html` product-details
component (rejected as a bullet surface because its `<ul class="pd-features">`
renders `<li></li><li></li>`). Findings, all evidence-backed:
- The page is **planned, active, deployed** — and hollow: empty `<h1 class=
  "pd-title">`, empty price, empty SKU, four empty feature bullets — while still
  rendering Add to Cart / Buy Now / size + colour swatches / star ratings on a
  site that sells nothing. Same on `/product-detail.html`. Not orphans: both are
  in the current site plan (roles entity_page / content).
- `content_data` has 49 keys but they are ALL chrome (`sku_label`,
  `add_to_cart_label`, `size_option_*`, site boilerplate). **Every value field is
  absent** (`product_name`, `product_price`, `feature_1..4`, `product_sku`, …),
  though `input_schema` declares them `required:true, source:llm`. NOTE:
  input_schema uses a **`fields`** wrapper, not JSON-Schema `properties` — a
  `properties` query returns nothing and will mislead you.
- `product-card-with-cta` stored the **LLM's prose apology** ("this section
  requires product array data sourced from `query.affiliate_products` … marked
  `on_missing: skip_section`") AS CONTENT. skip_section never fired.
- **THE BIG ONE:** `empty_section` work items for exactly these sections exist and
  `page-build-handler` marked them **`complete`** on 2026-07-10 — the sections are
  still empty on 2026-07-14. **A fix loop that closes without fixing.** (~36
  empty_section items on robot-hands, many `unresolved` / stale.) Overlaps the
  fixloop workstream's "bug dissolves but isn't fixed" benchmark case.
- Why the shell survives assembly: `sectionHasVisibleContent` keeps anything with
  >10 chars of text, and the template is full of static labels ("SKU:", "Add to
  Cart") — **the filter measures text, not resolved data.**
- Scope: robot-hands 6 product component instances (broken); dartsonline 2 (FINE —
  renders 14 real cards, proving the pipeline works when a data source exists). The
  mechanism is generic, so it recurs on any site missing a required data source.
→ Full write-up: **`docs024_key_docs_latest/HANDOFF_2026-07-14_empty_product_sections.md`**

## Turn 38 — 2026-07-14 — Specificity fix deployed + VERIFIED LIVE (four glyphs); I2.4 fulfilment check BUILT

**✅ THE BULLETS ARE RIGHT ON THE LIVE SITE.** Re-dispatched `needs_sprite_css`
(attempt_count 0; prior item was `complete`, so the dedup partial-unique index
freed the canonical item_key for reuse). asset-deployer completed first attempt,
re-emitting `/assets/css/sprites.css` with the **scoped** overrides
(`ul.sprite-list>li.sprite-b-gauge::before,ol…` — 9 of them). Headless render of
the guide page now shows **four DISTINCT glyphs**: ⓘ info / ✓ check (the unclassed
item — the default) / gauge dial / ⚠ warning. Legible at 20px, one 75,745B sheet.
I2.2 + I2.3 are now COMPLETE and correct.

**Two verification gotchas worth keeping:**
1. **CDN cache raced me.** My first post-emit `curl` of sprites.css returned the
   OLD unscoped CSS and briefly looked like a failed deploy. A cache-busted fetch
   (`?cb=<ts>`) returned the new file (9 scoped selectors), and the canonical URL
   caught up seconds later (`last-modified` matched the item's completion time).
   `cache-control: public, max-age=3600` — **so hard-refresh before eyeballing.**
2. **Verified the deploy against the POD, not git** (per the standing practice):
   `grep -a "sprite-list>li.sprite-b-"` on `/app/agent-chassis` inside the running
   pod → present. Definitive; timing/commit-ancestry is not.

**I2.4 BUILT (Go rides the NEXT deploy; registration SQL already applied):**
- Realisation: HALF of I2.4 needs no new code — `unfulfilled_imagery_plan` already
  emits `needs_imagery` for ANY unfulfilled plan row regardless of kind, so
  "sprite_sheet planned but no asset" is covered. The genuinely missing half is
  "asset exists but sprites.css doesn't".
- New `check_sprite_css_missing` (`sprite_css_missing`) emits `needs_sprite_css`
  → asset-deployer. **DB-only, per house convention** (check_image_url_404's header
  states discovery checks make no HTTP calls), which forced a real design question:
  *how does a DB-only check know the CSS was emitted?* It couldn't — so
  `emit_sprite_css` now **stamps** the plan row after committing:
  `style_hints.sprites_css = {emitted_at, sheet_path, signature}`.
- The stamp buys **staleness detection**, not just presence: the signature is
  `imageryplan.SpriteGridSignature(rows, cols, cell_names)` (e.g.
  `3x3:check,gauge,…`), so re-verifying cell names or regenerating the sheet at a
  new geometry makes the committed CSS *stale* — which is worse than missing (it
  slices the WRONG glyphs) — and the check re-emits. Also re-emits if the sheet
  asset's `updated_at` is newer than the stamp.
- **Signature lives in `imageryplan`** (not duplicated per side): `actions` imports
  `discovery_checks`, so the check cannot import back — the shared package is the
  only drift-free home. Same reasoning the package already applies to
  Classify/ItemKey/BuildSpec. Stamp failure is non-fatal (CSS is already committed;
  worst case the check re-emits next pass — idempotent).
- Registered on design-discovery-agent via
  `SQL_2026-07-14_register_sprite_css_missing.sql` (backup + verify; 20 checks,
  idempotent append). Safe to apply BEFORE the Go deploys: an unregistered check
  name is a `logger.Warn` + `continue` (discovery_checks.go:123), not a failure.
- Tests: `TestSpriteCSSStaleness` covers missing / fulfilled-so-don't-re-emit
  (the idempotence case that stops an infinite re-commit loop) / cell names changed
  / geometry changed / sheet regenerated / unparseable timestamp.

**Expect one self-healing re-emit after the next deploy:** the live CSS was
emitted by the pre-stamp binary, so the plan row has `sprites_css = null`. The
check will fire once (reason `missing`), asset-deployer re-commits identical CSS,
the row gets stamped, and it goes quiet. That one cycle doubles as the live
functional proof of I2.4 — watch for it rather than assuming it.

**Next:** user's live gate on the four glyphs → then **I2.5** (D10 container
opt-in: `.sprite-bullets ul>li::before` + the class on article-body's
`.article-body__content` wrapper) closes I2 → then Phase I3.

## Turn 39 — 2026-07-14 — USER GATE PASSED ✅; I2.5 built but BLOCKED by a content-destroying bug it uncovered

**✅ LIVE GATE PASSED.** User confirmed the four glyphs are present and correctly
placed on the guide page. **I2.2 + I2.3 are DONE.** (They needed a "where to look"
artifact first — the bullets are 20px, deliberately quiet, and sit at the very end
of a long article. Worth remembering for future gates: *say where to look.*)

**I2.5 BUILT (D10 container opt-in) — Go done, unit-proven, NOT landed:**
- `buildSpriteCSS` now emits BOTH opt-ins from ONE selector list (`listScopes`), so
  they cannot drift: `ul.sprite-list` (class on the list) and `.sprite-bullets ul`
  (class on a CONTAINER — the one that works for generated content). Overrides stay
  scoped in both, so the Turn 36 specificity trap can't reappear in the new scope.
- **`imageryplan.SpriteCSSFormat` (=2) added, and stamped.** Without it the I2.4
  check would compare an UNCHANGED grid signature, conclude the committed CSS was
  current, and no site would ever pick up the new rules. The signature tracks the
  SHEET; the format tracks the STYLESHEET. Bump the const whenever buildSpriteCSS
  changes shape. Tests cover the bump.

**🛑 BLOCKED — and the reason is a serious pre-existing defect (spun out):**
I2.5 needs the `sprite-bullets` class on article-body's `.article-body__content`
wrapper. That wrapper **is not in the deployed markup** — which led to:
- **The content writer never parses the LLM's JSON envelope.** `content_data.result`
  is a STRING containing `{"content": "<h2>…"}`. The template wants `{{.content}}`,
  which is buried inside that string.
- **9 article bodies have been SILENTLY BLANKED** across 5 sites: the light
  re-render path renders the missing `{{.content}}` as EMPTY (Go template
  `missingkey=zero`), overwrites the good rendered_html, and assembly then DROPS the
  empty section — the article just vanishes from the live page. **5 more leak raw
  JSON** (readers literally see `{ "content": "` — including our own sprite gate
  page). Only **2 of 16** article-bodies are healthy.
- **The trigger is `image_landed`** — scoped rerenders fire automatically when an
  image asset lands. **This workstream lands images.** Phase I0's per-page heroes
  are the most likely cause of the 9 blanked pages. The latent bug is upstream, but
  we probably pulled the trigger. Own it.
- **⚠️ STANDING WARNING: do NOT land an image or fire a scoped rerender on those
  pages until it's fixed — it blanks the article.** Assemble-only rerenders are safe.

**WHAT SAVED US:** the user chose "try the system re-render". Before firing it at a
live page I ran an OFFLINE PROBE — rendered the real template against the real
stored content_data in a throwaway Go test. It produced an EMPTY
`.article-body__content`. Firing the "fix" would have destroyed the guide's article.
**Standing lesson: when a repair path renders from stored data, probe the render
offline first — the cost is one throwaway test; the alternative is a blanked page.**

Full write-up (root cause, all 14 affected pages, recovery recipe — the words are
ALL still recoverable from content_data, and the upstream fix):
**`docs024_key_docs_latest/HANDOFF_2026-07-14_article_body_json_envelope.md`**

**I2 status:** I2.2 ✅ I2.3 ✅ (gated) · I2.4 ✅ built (rides next deploy) ·
I2.5 ✅ built, ⛔ landing blocked on the article-body fix. I2 closes when I2.5 lands.

<!-- Append new turns below this line. Format: ## Turn N — date — one-line summary -->
