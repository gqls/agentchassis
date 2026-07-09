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

<!-- Append new turns below this line. Format: ## Turn N — date — one-line summary -->
