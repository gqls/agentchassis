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

## Turn 40 — 2026-07-14 — I2.4 LIVE and PROVEN end-to-end (incl. idempotence); I2.5's CSS live

**The deploy already carried I2.4 + I2.5** — verified by grepping the RUNNING POD's
binary, not git (v1.0.1118 was built from the local filesystem AFTER the bulk commit
swept the code in). Note the grep gotcha: `sprite-list>li.sprite-b-` is ABSENT from
the binary and that is CORRECT — I2.5 composes that selector at runtime via
`joinScoped`, so only the fragments (`ul.sprite-list`, `.sprite-bullets ul`,
`>li.sprite-b-`) exist as literals. Grep for what the code actually contains.

**SAFETY FIRST (the standing warning):** before triggering any discovery I checked
that all 30 robot-hands imagery plan rows already have assets → `unfulfilled_imagery_plan`
can emit no `needs_imagery` → no image lands → **no `image_landed` scoped rerender →
no article bodies blanked.** I also ran design-discovery-agent (checks only) rather
than the full improvement-loop, so none of the loop's other machinery could fire a
scoped rerender. Verify this before ANY discovery run until the article-body fix lands.

**I2.4 PROVEN, the full chain, in order:**
1. Discovery pass → `sprite_css_missing` FIRED, finding + work item with
   `reason: "missing"` (no stamp on the plan row), handler asset-deployer, attempt 0.
2. Item sat at `detected` — discovery items are NOT auto-triaged (standalone check run
   has no triage step). Promoted to `triaged`, attempt_count 0.
3. Dispatch → asset-deployer → `emit_sprite_css` → COMPLETE first attempt.
4. `/assets/css/sprites.css` re-emitted: **12 `.sprite-bullets` container rules live**
   (I2.5's CSS), per-list scope intact, 3,179 bytes.
5. Plan row STAMPED: `{format: 2, signature: "3x3:check,gauge,…", emitted_at, sheet_path}`.
6. **IDEMPOTENCE PROVEN (the one that mattered):** re-ran discovery → the other three
   checks reported findings, `sprite_css_missing` did NOT, and **zero** new items. It
   fires once, fixes the gap, stamps, goes silent. Had the stamp write failed, it would
   have re-committed an identical stylesheet on every pass forever.
7. No regression on the gate page — the four distinct glyphs still render.
Also observed: a second discovery run while the item was still open created NO
duplicate — the `(site_id, item_key)` partial-unique dedup held.

**I2 status:** I2.0–I2.4 ✅ COMPLETE AND LIVE. I2.5 = CSS live; only the
`sprite-bullets` class on article-body's wrapper remains, still ⛔ BLOCKED on
`HANDOFF_2026-07-14_article_body_json_envelope.md` (the wrapper isn't in the deployed
markup on 14/16 article pages). I2 closes the moment that repair lands.

## Turn 41 — 2026-07-15 — I2.5 LANDED + PROVEN on the gate page → I2 COMPLETE

**Container opt-in proven end-to-end on the live site.** To land I2.5's class I had
to repair ONE article-body page (robot-hands has no healthy article page — all are
leak/blanked). Chose the sprite gate page itself (friction-calculator guide), which
was ALSO a JSON-leak page — so this both closes I2 AND removes a visible leak.

**What I did (scoped pilot; fleet-wide repair stays in the article-body handoff):**
1. Extracted the article HTML from the never-parsed envelope in
   `content_data.result`. NOTE the envelope is NOT valid JSON — HTML attribute
   quotes are BARE (`class="sprite-list"`, not `\"`), so the writer stored a naive
   concatenation. Extraction = string surgery (substring after `"content": "`), not
   a JSON parse. **The envelope is also TRUNCATED mid-word ("…it is a sympt")** —
   pre-existing (the live page already showed it cut off); recovery from these rows
   is therefore PARTIAL. Added that to the article-body handoff.
2. Set `content_data.content` (the schema/template field) from the extracted HTML;
   kept result/type as evidence.
3. **I2.5:** added `sprite-bullets` to the GLOBAL article-body template's
   `.article-body__content` wrapper (one row, all sites; inert without sprites.css
   → only robot-hands themes). Backup `bak_cc_articlebody_20260715`.
4. Set the component's `rendered_html` to the template-rendered HTML (removes the
   leak, adds the wrapper). Backup `bak_pc_gate_articlebody_20260715`.
5. Assemble-ONLY page_rerender (no `spec.reason`) — the SAFE path (no section
   re-render, no re-resolve, hero/CTA untouched). Verified hero+CTA healthy first.

**PROVEN on the served page (headless render):**
- The "Coefficient of Friction" factors list — plain LLM content, **NO class of its
  own** — now shows sprite glyphs on every item, themed purely by the wrapper. That
  IS the container opt-in working: generated content themes itself, zero markup.
- The Safety Factor list keeps its explicit info/gauge/warning glyphs — per-item
  overrides (0,2,3) still beat the container default (0,1,3). Both scopes coexist.
- JSON leak GONE (article starts clean at its h2); all 5 h2s intact; nav/footer
  `<ul>`s correctly NOT themed (they're outside the wrapper).

**I2 STATUS: ✅ COMPLETE.** I2.0–I2.5 all done and live. Sprite sheets, bullets,
fulfilment check, container house-style — the whole phase. Acceptance met.

**OPEN DESIGN NOTE (raised to user):** with the container opt-in, EVERY content list
gets the DEFAULT glyph = check (cell 0). On a factors/checklist it reads fine; on a
neutral/reference list a check is a touch assertive. Question for the user: keep
check as the universal default, or make the container default a more neutral glyph
(e.g. arrow, cell 6) and reserve check for explicit `sprite-b-check`? Not a blocker.

**Fleet reminder unchanged:** the article-body JSON-envelope defect still affects
13 other pages (9 blanked, 4 leaking) — `HANDOFF_2026-07-14_article_body_json_envelope.md`.
This turn repaired ONLY the gate page. And landing an image on the unrepaired pages
still blanks them.

## Turn 42 — 2026-07-15 — USER CHOSE arrow default; implemented (Go), rides next deploy & self-heals

**Design decision (shown a live side-by-side, chose arrow):** the container opt-in's
DEFAULT list bullet = **arrow** (neutral marker), not check. `check` stays reachable
as an explicit `sprite-b-check`. Rationale: the container themes EVERY content list,
so the fallback should be neutral; a check reads as affirmation.

**Implemented in `buildSpriteCSS`:** default resolved by NAME via new const
`spriteDefaultBulletGlyph = "arrow"` (falls back to cell 0 if the sheet lacks it —
robust to per-site vocabularies), used for the `>li::before` default rule. Per-glyph
overrides unchanged. `imageryplan.SpriteCSSFormat` bumped **2→3**; tests updated
(default now asserts arrow `0px -40px`; added an assert that explicit `sprite-b-check`
still maps to cell 0). Build + all sprite tests green.

**No live action now, by design:** the deployed binary is format 2 (check) and the
plan row is stamped format 2, so `sprite_css_missing` sees no mismatch → no churn.
On the NEXT deploy (format 3), the check finds stamped-2 ≠ code-3 → re-emits the
arrow CSS → re-stamps format 3. The gate page's content lists flip check→arrow with
no page edit (CSS-only); the Safety Factor list keeps info/gauge/warning. This is the
same self-healing cycle I2.4 already proved — so the arrow rollout IS the next live
confirmation of the format-version mechanism. Watch for it.

**I2 remains COMPLETE** — this is a house-style refinement within it, not new scope.

## Turn 43 — 2026-07-16 — Doc sync + status read-out; arrow default confirmed NOT-yet-live

**Housekeeping turn (no code).** Created `READOUT_2026-07-16_imagery_status.md` (a
spoken-word status briefing for reading aloud) and brought the workstream docs
current: HANDOFF (WHERE-WE-ARE header + refreshed READ FIRST pointing at the
separate image-landing thread + refreshed Next actions + doc map), RUNBOOK (B10/B11
closed, added B12 arrow-default gate + B13 content-loss pointer), PLAN decisions
already carry D9/D10.

**Live-state check (so the docs are accurate):** prod is now `v1.0.1123`. The
served `sprites.css` default bullet is STILL `0px 0px` (check) and the plan-row
stamp is STILL format 2 — so the **arrow default (format 3) is committed but NOT
yet live.** It self-heals on the next discovery pass once the format-3 binary is
running (sprite_css_missing sees stamped-2 ≠ code-3 → re-emit arrow). Verify with:
`curl -s https://robot-hands.com/assets/css/sprites.css | grep -o 'li::before[^}]*background-position:[0-9px -]*}' | head -1`
→ arrow = `0px -40px`.

**Article-body / image-landing trap: now a SEPARATE active thread** (user driving
it). Wrote `../aaa_fails_to_mend/004_HANDOFF_image_landing_blanks_article_body.md`.
Re-verified current state while writing it: 13 pages broken (9 blanked, 4 leak;
down from 14 since the Turn 41 gate-page repair). The escalate-not-blank GUARD is
committed in source but ABSENT from prod `v1.0.1123` (has `escalateRerenderToWriter`
but not `missingRequiredLLMFields`) — so **the trap is still LIVE in prod**; the
standing "don't land images on affected pages" rule holds. `ParseLLMJSON` writer
repair exists but its test fails on 14 fixtures (some envelopes truncated →
unrecoverable). All cross-referenced from the 004 handoff.

**I2 remains COMPLETE.** Next imagery phase is I3. Overall priority = the content-
loss fix (separate thread), not more imagery.

<!-- Append new turns below this line. Format: ## Turn N — date — one-line summary -->

## Turn 44 — 2026-07-16 — Live re-verification: trap CLOSED, testbed CLEAR → I3 unblocked; detection-query defect found

**Session goal:** resume the workstream (I2 done → I3 next). Before touching
anything, re-verified live state — and it has MOVED since Turn 43, in our favour.

**1. The image-landing guard is now LIVE in prod (contradicts Turn 43).** Turn 43
recorded the guard ABSENT from `v1.0.1123`. This turn the RUNNING pod
(`agent-chassis-5f8ddf649f-dgss6`, started 08:44Z today) greps `2 / 1 / 4` for
`missingRequiredLLMFields` / `"escalating page to writer instead of blanking"` /
`escalateRerenderToWriter` — the FULL guard, matching 004 §3's own criterion. So
the tag `v1.0.1123` was **rebuilt in place** between Turn 43 and now (consistent
with the build-from-local-filesystem practice — same tag, new binary; the pod is
ground truth, never the tag string). **The trap is closed.**

**2. robot-hands testbed is fully clear.** All 3 article-body pages healthy; all
**30** imagery plan rows fulfilled (each has ≥1 active asset) — so nothing will
fire a `needs_imagery` → no image landing can even be triggered on the testbed.
I3 acceptance work can land images here safely.

**3. Content-loss recovery has advanced (separate thread) BUT a detection-query
defect masks 3 pages.** Robust length-independent detector
(`rendered_html ~ 'article-body__content[^>]*></div>'`) gives the true fleet state:
**4 BLANKED, 0 JSON LEAK, 13 healthy** (was 9+4 at Turn 43's snapshot). The 004/006
handoffs' `length(rendered_html)=1326` test UNDER-reports, because **I2.5's
`sprite-bullets` class made a fresh blank shell 1341 bytes** (`1326 + len(" sprite-bullets")`).
Three finetuning/gamesdesign pages re-blanked at 22:23–23:17 on 2026-07-15 —
BEFORE the 08:44Z guard pod, so NOT a guard failure. Corrected 004 §5 (dated note +
new detector + current 4-page list) so the separate thread doesn't under-count.

**4. Arrow default still not live** (as Turn 43): served CSS default `0px 0px`,
plan stamp format 2; HEAD `9752bc68d` carries `spriteDefaultBulletGlyph="arrow"` +
`SpriteCSSFormat=3`. Self-heals on the format-3 deploy. Unchanged.

**Docs updated:** 004 handoff (detection correction), imagery HANDOFF (WHERE-WE-ARE
re-verified + READ FIRST downgraded from STANDING WARNING to "trap CLOSED / residual
notes"), this entry. **Kicked off I3 code-surface reconnaissance** (Explore agent
mapping assets schema, custom_image_id precedent, ImagePurposes, generation entry
points, resolver, content-entity tables, card components, discovery-check pattern) to
produce a concrete build plan before running any migration.

## Turn 45 — 2026-07-16 — Phase I3 BUILT (I3.1–I3.4): entity link live, card derivation + resolver + check committed

**Recon findings that shaped the design (full map in the Turn-44 recon):**
- `assets` has NO entity link; the `affiliate_products.custom_image_id` FK is
  schema-only — **zero Go callers**. I3 is the first real content→asset link.
- **Why listing cards are empty, precisely:** `content-listing` declares
  `articles: {source: "query.blog_posts", on_missing: skip_section}` — and
  `blog_posts` was NOT in `queryresolve.Resolve`'s vocabulary (only
  pages_where_type / pages_under_section / section_index_for / products), and
  NO resolver projected an image field. Two gaps, both closed this turn.
- No WebP encoder exists anywhere in the codebase; `OptimizeImageForWeb` only
  fit-resizes (no crop). Heroes 1600×900 and cards 800×450 are both 16:9, so
  cover-crop degrades to pure downscale on the common path.
- Fleet article convention: `page_type='blog-post'` (9 pages on robot-hands).
- `imageryplan.go`'s alias table already names I3 as the designed seam
  (product_screenshot→hero "until Lane B lands literal keys").

**User decisions (AskUserQuestion): D11** cards = 800×450 JPG q82 now, WebP
deferred to I7 (no encoder; budget ≤60KB met). **D12** `card` is a derived
PURPOSE (I1 favicon/og_card pattern, `origin_asset_id` lineage), NOT a
chk_kind kind — amends D3's batch list with the same logic that kept `chart`
out. Migration approved and run.

**Built:**
1. **I3.1 (LIVE):** `SQL_2026-07-16_assets_entity_link.sql` — `entity_type` +
   `entity_id` on assets, partial lookup index + unique active
   (site,entity,purpose). Applied, verified.
2. **I3.2:** `derive_card_asset` action — mirrors derive_brand_head_assets:
   page-scope plan hero (site-scope fallback) → S3 download →
   `storage.CoverCropResize` (NEW exact-size cover-crop helper,
   image_processing.go, unit-tested incl. centre-crop proof) → JPG q82 →
   `sendGitCommitRequest` commits `/assets/images/card-<page>.jpg` →
   entity-linked upsert (NOT best-effort — the link is what resolvers read).
   `card` purpose in ImagePurposes. asset-deployer `content_card` mode
   chained after sprite_css (SQL applied; explicit Strategy-0 paths per the
   I2.1 lesson).
3. **I3.3:** queryresolve — `blog_posts` base (delegates to pages_where_type
   `blog-post`) + shared `pageImageProjection`/`pageImageJoins` fragments:
   every page-listing item now carries `image` = entity card → plan hero →
   "". Card wins the moment it exists (LEFT JOIN preference), heroes serve
   as day-one fallback so listings aren't imageless while cards derive.
4. **I3.4:** `content_image_missing` discovery check + registration SQL
   (applied; warn-and-skip until deploy). Two anti-churn gates: a real
   query.blog_posts consumer must exist, and the page must be derivable
   (own hero or site brand hero) — otherwise the handler would complete
   `derived:false` and the check would re-emit forever (the sprite_css
   stamp lesson, solved here by the entity link BEING the stamp).

**Verification:** `go build` clean (platform/internal/cmd); vet clean except
the known pre-existing unreachable-code note; storage, discovery_checks,
actions, imageryplan tests all pass. (`go build ./...` also trips over a
pre-existing package clash in docs/…/traffic_probe/deploy_setup — another
workstream's scratch, not code.)

**Not done / deferred:** card landings don't auto-re-render listing pages
(eventual consistency via next rebuild; revisit if it bites). News cards =
I5, product cards = I6, on the same entity columns. WebP = I7.

**Next: deploy → RUNBOOK B14** (one discovery pass fires BOTH
sprite_css_missing format-3 [the B12 arrow] AND content_image_missing → ~9
cards; then needs_page re-render of learning-center-hub; then the A3 gate).

## Turn 46 — 2026-07-16 — v1.0.1125 DEPLOYED: I3 mechanism acceptance MET LIVE; D13 (per-article generation) decided + built

**B14 executed end-to-end on v1.0.1125** (binary verified by pod grep: all I3
symbols present). Sequence + results:
1. Pre-flight: 0/30 imagery rows unfulfilled; zombies cleared.
2. improvement-loop triggered (kcat, correlation b20b7688). **Both checks
   proved live in ONE pass:** content_image_missing emitted EXACTLY its 9
   gated items; sprite_css_missing re-emitted on the format-2≠3 mismatch.
3. **Live finding #1 — dispatch priority is ASC** (`load_work_item_actions.go:589`
   `ORDER BY wi.priority ASC`): LOWER number = dispatched SOONER; comments in
   other checks confirm the convention ("30 // high"). Items at 65/70 are
   correctly "background" — but they sat behind ~50 stale 35–60 items, so for
   the watched run I bumped the 10 fresh items to priority 5. NOTE: this
   workstream's old "needs_page@99" habit actually meant "run LAST" — harmless
   in an empty queue, wrong mental model. Recorded here to kill it.
4. **All 10 items completed first-attempt** (~4 min once front-of-queue):
   9 cards derived + entity-linked + committed; sprite stamp → format 3;
   **served CSS default = `0px -40px` (ARROW) — B12 CLOSED.**
5. **Live finding #2 — card budget miss:** q82 on the dense photographic hero
   → 64,097B > the ≤60KB budget (D8). Fixed: card quality 82→78
   (`ImagePurposes`, committed, rides next deploy).
6. learning-center-hub re-rendered via a manual needs_page@5 (one dispatch
   zombie cleared en route; the item also showed the known completion-race
   timestamp oddity — completed fine). **`query.blog_posts` RESOLVED IN
   PRODUCTION:** content_data.articles = 9 items, each with ITS OWN
   `/assets/images/card-<page>.jpg`; rendered HTML carries all 9 distinct
   refs; **served page shows all 9 (HTTP 200).** I3 mechanism acceptance MET.

**Live finding #3 → USER DECISION D13.** All 9 cards are byte-identical
(64,097B): no blog-post page has a plan hero, so every card derived from
`hero_canonical` (proven via origin_asset_id lineage). Options put to user:
(a) planner emits article heroes (Lane A) vs (b) generate from article
content + style guide (Lane B proper). **User chose (b), and confirmed
today's cards should re-derive after the next deploy.**

**D13 BUILT (same session, all tests green):**
- `imageryplan.ContentHeroKey(page)` — literal key `content_hero_<page>`;
  SQL mirrors as `'content_hero_' || replace(name,'-','_')` (pinned by test +
  comments both sides).
- `check_content_image_missing` v2 — TWO-MODE emitter with a pure decision
  fn (`contentImageAction`, table-tested): no plan/content hero → GENERATE
  (item_type **needs_imagery**, BuildSpec shape → image-build-handler's
  generic call_imagery_gen path, prompt = title+meta_description subject
  only, style guide layers at generation; scope/scope_ref set so
  flag_rebuild re-renders the article page on landing); source exists but
  card missing OR **card.origin_asset_id ≠ current preferred source** →
  DERIVE. Origin-staleness = the user's "re-derive at next deploy" for free:
  pass 1 generates 9 heroes, pass 2 re-derives 9 cards at q78, pass 3 silent.
- Preference order unified in all three consumers (check, derive_card_asset,
  plan_sections.ensureAssets): **plan page hero → content hero → site brand
  hero** — so the article page renders the same image family as its card.
- Site hero deliberately no longer suppresses generation (D13's point) but
  remains the deriver's last-resort source.

**Safety note:** the parallel thread has now recovered ALL 17 article-body
instances (004 updated; root cause was writer max_tokens truncation) — image
landings from D13 generation are safe fleet-wide, and the escalate-not-blank
guard is live regardless.

**Next deploy carries:** q78 cards + D13 generation. Post-deploy: trigger
discovery (or let the loop cycle) → watch 9 generations (SDXL — B5 budget
note) → next pass re-derives 9 cards → re-render learning-center-hub →
**A3 EYEBALL GATE: 9 visually DISTINCT per-article cards.**

## Turn 47 — 2026-07-17 — D13 ran end-to-end on v1.0.1128 (mechanism PROVEN) but the GATE FAILED → two fix handoffs

**The convergence machinery worked exactly as designed:**
- 9 content heroes generated (SDXL, first-attempt, all committed + serving);
  triage stranded the items in `detected` AGAIN (second time this week) —
  hand-promoted 9; the improvement-loop triage-stranding is now a confirmed
  recurring platform nit.
- Origin-staleness re-derived all 9 cards from their own heroes (sizes went
  nine-identical-64,097B → nine distinct 37–73KB, q78 applied).
- Served learning-center-hub shows 9 distinct card refs. NOTE learned: card
  PATHS are stable, so re-derived bytes go live WITHOUT a page re-render.
  Also: my monitor's 10-min zombie-clear restart-looped a legitimately-long
  needs_page build — don't blanket-clear while a page build is in flight.

**USER GATE: FAILED — on quality and on site-level context.** Imagery
failures: style inconsistency across the 9 (photo / studio-with-pseudo-text /
line-art; at least one in colour — style-guide adherence weak on SDXL, no
reference anchoring on that path); click-through: only 2 of 9 match (1
mismatch = cycle-time article's flag_rebuild re-render never landed; 6 of 9
listed URLs are 404 — scaffold pages + never-deployed /blog/ duplicates ⇒
query.blog_posts needs an eligibility filter); only one surface (hub) has
cards. **User direction: redesign the card style around the SMALL FORMAT'S
limitations.** Site-level failures (NOT imagery): blue-brochure regression
(evidence: css_themes row + served styles.css are STILL tool-portal-dark —
the damage is component-level: header/footer regenerated non-theme-aware
2026-07-16 14:18 by the improvement loop; hardcoded_section_colors handler
stripped legit dark backgrounds; generic_theme check misjudges the site;
"37 pages missing header/footer" unresolved×2), yellow-on-white hub text,
learning-center nav/URL sprawl, tools broken (nav→content page;
/tools/matchmatrix/index.html 404; description-vs-tool = misdirected_cta
class, overlaps the experience_loop workstream), dead Load More button.

**HANDOFFS CREATED (start fresh chats from these):**
- `HANDOFF_2026-07-17_i3_imagery_gate_fixes.md` (imagery: style redesign,
  blog_posts filter, rollout to other card surfaces).
- `../robot_hands/HANDOFF_2026-07-17_robot_hands_site_fixes.md` (site: theme
  restoration R1 — the B7 state is INTACT, do NOT re-swap the FK; nav/IA;
  404 page rows; tools w/ experience_loop; load-more).
This workstream PAUSES I3 polish until those land.

## Turn 48 — 2026-07-17 — D14 decided & BUILT (content_hero kind, flat duotone, eligibility filter); awaiting another thread's rollout

Fix chat for the imagery half of the Turn-47 gate failure. **D14 with user:**
flat duotone editorial illustration for content heroes/cards (small-format-
first, per the user's direction), carried by a new `content_hero` KIND routed
to **Banana** (the only provider honouring reference anchors; SDXL's free-text
adherence is what failed), with the style guide gaining a **per-kind override
map** (`kinds`) that replaces direction/avoid/anchors WHOLESALE for its kind —
partial merging would let the photographic base voice ("cartoonish rendering"
in avoid, photographic reference anchors) contaminate the flat style.
Regen budget: **pilot 3, eyeball with user, then the rest**.

**Built & committed `4e35c8064`** (all green, pathspec commit):
- `check_content_image_missing.go`: Row.Kind → `content_hero`; sweep gains the
  F2.1 eligibility predicate `deployed_at IS NOT NULL AND
  jsonb_array_length(sections) > 0` (robot-hands 9→3 listed articles; the 6
  excluded are R6's build-or-retire rows in the site handoff).
- `queryresolve.go`: `blog_posts` base passes `listedOnly=true` into
  `resolvePagesWhereType` — same predicate, lockstep comments both sides.
  Fleet impact checked: only 3 consumers of query.blog_posts (robot-hands,
  dartsonline — 4 never-built `planned` rows, listing goes honestly empty —
  and idea.uk — no blog-post rows).
- `imagery_style_guide.go`: `Kinds` map + `directionForKind` override branch,
  new `avoidForKind`/`referenceKeysForKind` (override keys flow ungated; guide-
  level keys keep the flat-vector gate). Tests extended (`TestStyleGuideKindOverrides`).
- `generate_image_actions.go`: kindDefaults entry (NEUTRAL negative prompt —
  the visual language comes from the guide override, so a photographic-content-
  hero site isn't fought by a baked-in "no photorealism"); avoid/anchor call
  sites go through the per-kind accessors.
- `dynamic_adapter.go`: `content_hero` → banana. `url_helpers.go`:
  ImagePurposes `content_hero` {1600, 900, 85, jpg} (same geometry as hero —
  BuildSpec sets purpose from kind, so the kind rename changes the stored
  purpose; all content-hero consumers look up by asset_key, verified).

**Style guide superseded** → row `361f2ed7` (from `439329c4`, I1 seed), adds
`kinds.content_hero`; SQL artifact
`SQL_2026-07-17_d14_style_guide_content_hero_override.sql`. Inert until the
new binary lands (old struct has no `kinds` field; guide-level fields
unchanged).

**Rollout is another thread's** (v1.0.1132 claimed in their uncommitted
makefile bump; my commit rides HEAD). Coordination lesson re-learned live:
my makefile tag-bump edit was rejected → the tag had been claimed under me
minutes later. Coverage check run before any dispatch: no open imagery items
on robot-hands; D13 residue = 1 failed grip-force re-render, 2 scaffold
re-renders in needs_human_review (will be mooted by F2.1/R6), site-chat items
untouched.

**NEXT (blocked on rollout):** pod-verify `strings /app/agent-chassis | grep
-c content_hero` (and the image-generator-adapter binary) → pilot: supersede
the 3 LIVE articles' content heroes (tool-grip-force-friction /
tool-gripper-payload / tool-gripper-cycle-time — the cycle-time one also
clears its F2 stale-hero mismatch on re-render) → hand-promote stranded
`detected` items → eyeball 3 with user → release remainder → re-gate on
learning-center-hub → then F3 surfaces (featured_article / tool-list).

## Turn 49 — 2026-07-17/18 — D14 PILOT PASSED live: 3 flat-duotone heroes + cards, F2.1 filter proven on the served hub

**A stale-adapter trap, caught by the pod-grep and worth remembering.** The
rollout thread shipped v1.0.1134 and the chassis binary carried my markers —
but the **image-generator-adapter binary at the same tag was pre-I2 source**
(no `sprite_sheet`, no `content_hero`; `dispatching to provider` present, so
it was a real but OLD build). Had I trusted the tag, the pilot would have
silently generated on SDXL and "failed" the new style for the wrong reason.
Proven stale three ways: pod grep, a `docker run` grep of the local image,
and a `--no-cache` rebuild from the HEAD archive (which DID contain the
routing) — i.e. the shipped image never came from that source. Rebuilt as
**v1.0.1135** (chassis + adapter), pod-verified both.
**Fixed at the root:** `quick-agent-update` now builds/pushes/deploys/
restarts the **image-generator-adapter alongside the chassis** (commit
c0ef457a1) — a chassis-only release was restarting adapter pods onto a
retagged-but-stale image, and the two share the kind vocabulary.

**Pilot ran end-to-end, all three generations first-attempt:**
- Runtime proof of routing, from the adapter's own log: `"kind":"content_hero",
  "provider":"banana","model_id":"gemini-3-pro-image-preview"` ×3.
- Items stranded in `detected` AGAIN (third time this week) — hand-promoted.
  The **derive** items landed as `unresolved` (not `detected`) this time and
  needed the same promotion; watch for both statuses.
- **User gate on the 3 heroes: 2 on-style, 1 drift** — grip-force and payload
  came back charcoal-ground flat duotone; **cycle-time came back on a WHITE
  ground**. Ruling: tighten `avoid`, re-roll cycle-time only, accept payload's
  blue-heavy variant. Avoid gained "white background, pale background, light
  background, bright full-bleed colour field" (row `1c51bafb`, supersedes
  `361f2ed7`; `SQL_2026-07-18_d14_avoid_tighten_white_grounds.sql`). The
  re-roll came back correctly dark — **so the avoid list, not the medium
  description, is the lever for ground-colour drift on Banana.**
- Cards re-derived from all 3 (origin_asset_id lineage now points at ACTIVE
  content heroes) and are **22.1KB / 26.4KB / 23.6KB — comfortably inside the
  ≤60KB D8 budget**, where the D13 photographic set ran 37–73KB (one over).
  Flat colour compressing better at q78 is exactly the D14 small-format bet,
  now measured.

**Served state on learning-center-hub (the re-gate):** the listing shows
**exactly 3 articles, 3 distinct on-style cards, and every click-through
resolves 200 to an article showing ITS OWN content hero** — F2's mismatch
(cycle-time) and all six 404 links are gone, the latter because the F2.1
eligibility filter now excludes the scaffold/never-built rows. F1 (style
consistency) and F2 (click-through) are closed for this surface.

**Landmine confirmed live:** dispatch is one-site-at-a-time against a
fleet-wide pool — robot-hands items sat `triaged@5` for ~10 min while
leopardess/vetcomparison held the loop. Priority 5 does not jump the queue
across sites; it only orders within one. Budget as wall-clock, not failure.

**STILL OPEN:** F3 rollout to the other surfaces (featured_article,
product-card-with-cta, news-listing, info-card-grid, tool-list) — untouched;
the 6 excluded blog-post rows remain R6's build-or-retire call in the site
handoff; RUNBOOK B5 (formal budget sign-off) still open — this pilot spent 4
Banana generations.

## Turn 50 — 2026-07-18 — F3 surveyed: not shovel-ready; council summary filed

**Council gate written up from the submitting side** —
`../fixloop_eg_dartsonline/SUMMARY_council_first_external_catch_2026-07-18.md`
(commit 40f547c6d): the three real defects it caught on D14, why the
evidentiary objections cost extra rounds (the gate reviews the PLAN, not the
repo — front-load your evidence), and why this thread committed without a
`Council-Reviewed:` trailer (verdict was revise, and 098 buckets false
trailers on purpose).

**F3 surveyed before touching code — and the handoff's suggested order was
wrong.** Live fleet usage per surface: info-card-grid 15 pages/7 sites,
news-listing 7/5, tool-list 5/3, content-listing 3/3 (done), and
**featured_article + product-card-with-cta on ZERO pages fleet-wide**. The
handoff proposed starting at featured_article; it is used nowhere and its
resolver base is unimplemented, so that would have been speculative work.

The finding that matters: **no decision-free F3 code remains.**
- `tool-list` is the closest to ready — it is query-fed and the resolver
  ALREADY hands it an `image` (the shared `pageImageProjection`), so the
  template is the only missing piece — but **0 of 38 tool pages fleet-wide
  have any image at all** (no card, no plan hero), so the slot would render
  nothing. Making it real costs ~38 generations plus extending
  `content_image_missing` beyond `page_type='blog-post'` (its consumer gate is
  `query.blog_posts`-specific and would need per-page-type logic). That is the
  B5 budget call, still formally open.
- `info-card-grid` has the widest reach by far but is not query-fed and has no
  `<img>` at all — a design decision about whether info/category cards want
  imagery, and from which entity.
- news-listing is I5; product cards are I6.

**Landmine recorded:** `content_components` is a GLOBAL library (no `site_id`;
`forked_from` for forks). Editing `tool-list`'s template is a fleet-wide change
to every adopting site, not a per-site tweak — and a component-write guard is
being built by another thread right now, so hand-editing component HTML in the
DB is the wrong move this week.

**Next actor:** F3 needs an owner decision before code — see the costed table
in the handoff's F3 section.

## Turn 51 — 2026-07-18 — F3 tool surface SHIPPED; and a CORRECTION: the "stale adapter" was my measurement error

**CORRECTION FIRST, because Turn 49 is wrong and other threads may have read it.**
Turn 49 (and the first version of the 016b §9 entry, and the handoff's trap #1)
claimed the image-generator-adapter shipped **stale** at v1.0.1134 — "proven
three independent ways". **That was FALSE.** All three proofs greped for
`content_hero` / `sprite_sheet`, and those literals are simply **not retained**
by the Dockerfile's `CGO_ENABLED=0 go build -a -installsuffix cgo` on
`golang:1.24-alpine`, though they survive a plain host `go build`. Measured:

| marker | shipped image | host `go build` |
|---|---|---|
| `content_hero` | **0** | 1 |
| `sprite_sheet` | **0** | 1 |
| `infographic` (same `case` clause!) | 1 | 1 |
| `"dispatching to provider"` (log) | 1 | 1 |
| `"reference images will be IGNORED"` (log) | 1 | 1 |

The clincher: the shipped binary contains `"reference images will be IGNORED"`,
which was added to that file **later** than `content_hero` — a stale binary
cannot have the newer string and miss the older one. So the adapter was current
all along, the D14 routing would very likely have worked on v1.0.1134, and the
rebuild I did was unnecessary. 016b §9 has been rewritten with the real and
better lesson: **a pod-grep is a POSITIVE test only — a miss proves nothing
until you show the marker survives a known-good build.** Pick log-message
strings as markers, never `case` values. The `quick-agent-update` change
(release the adapter with the chassis) is kept — it is sound practice — but its
stated justification was a measurement error and now says so.

**F3 tool-page surface SHIPPED (owner funded the ~33-page rollout).**
- `content_image_missing` now iterates a **surface table** rather than
  hardcoding `blog-post`: each entry carries the page type, the consumer LIKE
  that proves the site lists that type, the eligibility predicate and the
  prompt's subject noun (commit `8b804bc27`). Tool pages differ in **all four**,
  which is why a widened `IN` list would not have worked: they are listed by
  `query.pages_where_type:tool`, their substance is the committed
  `/tools/<name>/` bundle so `sections` is legitimately empty (the article rule
  would exclude **20 of the fleet's 33** deployed tool pages — hence the new
  `queryresolve.DeployedPageEligibilitySQL`), and "Article header image" is the
  wrong phrase for a calculator. The per-pass cap now spans surfaces so a site
  with both articles and tools cannot spend double.
- `tool-list` got its image slot (migration `170`, applied): the resolver was
  **already** handing it an `image` per row (shared `pageImageProjection`) —
  only the template never rendered it. `{{if .image}}`-guarded `.tl-card-media`
  at 16/9 `object-fit:cover`; imageless cards keep today's icon treatment, so
  it is inert on sites without tool imagery. Guarded exact-match replace in the
  `151` house style, full pre-edit template in a doc_note **and** a
  `component_versions` snapshot.
- Released **v1.0.1136** via `quick-agent-update` — which now ships both
  services, exercising my own fix. Chassis pod-verified on markers that do
  survive (`contentImageSurfaces`, and the eligibility SQL fragment).

**Landmine for the next actor:** `content_components` is a GLOBAL library
(no `site_id`), so `170` changed `tool-list` for all 3 adopting sites at once.

## Turn 52 — 2026-07-19 — F3 surface #2 LIVE: robot-hands' tool directory has card imagery

3 tool content heroes generated (Banana, flat duotone, on-style — the tightened
`avoid` held: dark grounds, no white drift), 3 cards derived, and the live
`/index.html` tool directory now renders them. Served: `card-tool-*.jpg` at
**22.3 / 27.1 / 36.4 KB**, all inside the ≤60KB budget.

**Exactly 3 of robot-hands' 5 tool pages emitted** — `DeployedPageEligibilitySQL`
correctly excluded the 2 that never deployed, and the 2 imageless cards render
with no media div at all (the `{{if .image}}` guard), so nothing regressed.

**THE MECHANISM FINDING — a page_rerender does NOT pick up a component template
change, and this is not obvious.** `page-rerender`'s workflow branches on
**`input_data.spec.reason`**:

```
check_rerender_mode: condition
  spec.reason IN ('image_landed','section_data_resolved','cta_links_stale')
      -> rerender_sections   (re-renders from the TEMPLATE + freshly resolved fields)
  else -> render_page        (assembles the STORED page_components.rendered_html)
```

I queued a rerender with no `reason`, it completed green, the page redeployed —
and the template change was still invisible, because `page_components.rendered_html`
is a **stored rendered artifact** (this instance's was frozen at 07-18 03:26).
Re-queuing the identical item with `"reason":"image_landed"` re-rendered the
sections and the imagery appeared immediately. **If a component template edit
"doesn't take", check the reason field before you doubt the edit.** Same family
as the known landmine that `site_components.rendered_html` is a rendered artifact
regenerated from its template.

**Also confirmed live:** the manual kcat trigger is **flaky via
`kubectl run -i --rm`** — two fires produced NOTHING on the topic (verified by
consuming the topic tail: no design-discovery message at all), which reads
exactly like a broken consumer. The stdin attach races the container start.
Reliable form, used for every subsequent fire:
`kubectl run <pod> --restart=Never --attach=false --env="PAYLOAD=$JSON" --command -- sh -c 'printf "%s" "$PAYLOAD" | kcat -P ...'`
Note `--rm` is rejected with `--attach=false`; delete the pod afterwards.
**Before concluding "the dispatch is broken", consume the topic and check the
message was ever produced.**

**Remaining F3 spend (owner funded ~33 fleet-wide, ~5 used):** gamesdesign.co.uk
(9 tool pages) and idea.uk (1) also carry `tool-list`, so their next discovery
pass emits for their deployed tool pages, capped at 10/site/pass. The other 7
sites have tool pages but **no `tool-list` component**, so the per-surface
consumer gate correctly keeps them from spending anything.

## Turn 53 — 2026-07-19 — checked whether the tool rollout was safe to let drain. It is not: `content_hero` is unstyled off robot-hands, and the rollout figures were wrong

The handoff's next action #2 was "let the funded tool rollout drain". Before
firing anything I checked what would actually emit. Two findings, both live-
verified, both filed as **`/bugs_open/027`**.

**FINDING 1 — the rollout figures in my own handoff were wrong in both
directions.** I wrote (B16.2 / next-actions §2) that gamesdesign.co.uk (9) and
idea.uk (1) draw on the funded budget and "the other 7 sites with tool pages have
no `tool-list`, so the consumer gate spends nothing on them". Against the live DB,
running the check's own gate query:

```sql
SELECT s.domain,
       COUNT(*) FILTER (WHERE cc.input_schema::text LIKE '%query.pages_where_type:tool%') AS tool_list_consumers
  FROM page_components pc
  JOIN content_components cc ON cc.id = pc.component_id
  JOIN pages p ON p.id = pc.page_id
  JOIN sites s ON s.id = p.site_id
 WHERE p.status IN ('active','deployed') GROUP BY 1;
```
`finetuning.uk` (1 consumer) and `leopardessconsulting.co.uk` (2) also pass the
gate, with **5 deployed tool pages each** — 10 generations nobody had counted.
And `idea.uk` spends **nothing**: its single tool page has `deployed_at IS NULL`,
which the tool surface's `DeployedPageEligibilitySQL` excludes.

So the real pending exposure is **19 generations across 3 sites**, not 10 across
2. I had asserted the smaller figure from a survey rather than from the gate
query, which is the mistake — the check's own predicate is the only thing that
answers "what will emit", and it was a two-minute query away.

**FINDING 2 — and the reason that exposure matters: `content_hero` has no
defined style on any site but robot-hands.** D14 gave the kind a per-kind
override map (`imagery_style_guide.kinds.content_hero`). What it did not do is
add `content_hero` to `directionAppliesToKind` — so when a site has no override,
`generate_image_actions.go:415` falls through to the free-text
`design_intent.imagery_direction`, which is written to describe the site's
*photographic* house style. That is the contamination class the function's own
doc comment was written to prevent.

Live: **only robot-hands.com has an `imagery_style_guide` row at all.** The three
sites queued to generate have none, so all 19 would take whatever free text their
`design_intent` holds — including leopardess's, which describes *two* styles and
expects a human to pick one.

**Why this reads as done and isn't:** the five-place checklist for a new kind is
written in the handoff's own Mechanisms section and says a new kind must be added
to BOTH gating functions. D14 satisfied one of them via the override map, and
robot-hands — the only site exercising the kind — has the override, so every
observation of `content_hero` to date has been of the branch that works. A fix
covering one branch of a two-branch router reads as done; this is that shape
again.

**Not on fire, but armed.** `scheduled_tasks` has no discovery/improvement-loop
entry (12 enabled tasks: health checks, reapers, build-pipeline-trigger, feed
refresh), so passes are fired by hand — but the check is registered on
`design-discovery-agent` by `type`, so it runs on *every* site's pass. Any
concurrent session running a routine sweep on those three sites trips it.
Leopardess had 123 discovery items in the last two days.

**Not verified:** I have not run a generation on the three sites, so the output
harm is inferred from the code path plus the D13 gate precedent, not observed.
The code path, the missing style guides, the 19-page exposure and the corrected
figures are all verified and quoted in 027.

---

## 2026-07-19 — bug 011 R1: provider routing (session "bugfix thread3")

**Committed `6896ce22e`** on `085_debug_and_feature_loops`. Built
`image-generator-adapter:v1.0.1137`; **`agent-chassis:v1.0.1137` NOT built** —
stopped mid-task, see missteps. Nothing deployed. Bug 011 stays OPEN per the
`/bugs_closed/` bar (fixed AND live); R1 is fixed in code only.

**What the defect actually was.** Not "hero is on the wrong model" — that is the
instance. The mechanism is `generateImage`'s hand-maintained `switch data.Kind`
whose `default:` selected Stability *silently*. `content_hero` fell through it
(shipped mis-routed, found at the D13 card gate); `hero` fell through it (shipped
a gibberish diagram as leopardess' homepage). The switch's own comment already
blamed itself in writing. Adding `hero` to the case list would have fixed
instance three and left instance four.

**Shipped:** `internal/adapters/imagegenerator/routing.go` — `kindProviderRouting`
map + pure `routeProvider()`, so the routed set is *enumerable data* and an
unrouted non-empty kind is detected and logged by name (`UNROUTED KIND`, listing
the valid set). Empty kind deliberately does not warn (documented legacy path; a
warning that always fires is one nobody reads). `hero` added. Per-site
`provider` field on `imagery_style_guide` (guide-level + per-kind, mirroring
`avoidForKind`'s override-wins-even-when-empty contract) → `provider_hint` →
adapter. Six routing tests + `TestProviderForKind`.

**Verified, not asserted** (both were council escalate-conditions):
`ImageRequestData{}` is constructed nowhere — the adapter only unmarshals from
Kafka JSON; only `topic_manager.go` (declaration), `generate_image_actions.go`
(sole producer) and the adapter touch the topic. `GenerateImageAction` is the
exclusive path. `imagery_style_guide` JSON has one reader
(`getImageryStyleGuideForSite`) plus one seed file — no UI, no other service.

**Rejected design, with evidence.** Keyword-inferring the provider from
`design_intent.imagery_direction` — the obvious reading of R1 — was checked
against all 11 live values first and fails on ≥3: `9ec3b9ee` = "Minimal
photography. Prefer abstract geometric constructions…", `1244516d` = "Photography
and illustration should be minimal…". Both contain "photography" while intending
the opposite; a substring match misroutes them *silently*, i.e. reproduces the
bug class being fixed.

**Council:** `e996bf0a-4cdd-40fa-8ff0-1f1a76c3d181`, three rounds, final
**REVISE — not approved**, so no `Council-Reviewed:` trailer (that would register
as MISMATCH in the 098 report). `bug_historian` went high→medium and accepts the
mechanism fix's shape; residual is that `UnmigratedKind` is a log line, not an
`agent_error_log` row or work item. **Not done here, and the trap for whoever
does it:** the adapter has *no DB handle at all*, so persisting from there is an
architectural change; and relocating detection to the action layer (which has a
DB) would put a second copy of the routing table on a **separately-imaged
service** — the dedup-index↔Go-list drift class. Correct shape is *adapter
reports → action persists*. Written up in `bugs_open/011` §6.

**UNVERIFIED, owner-facing:** no cost or latency parity established between
Stability and Banana for the fleet's largest kind. No billing data was available
and none is claimed. Adapter HTTP timeout (120s) was tuned around SDXL's 30-60s.

### Missteps this session (the point of this file)

1. **I did not file the diagnosis loop before asserting a durable mechanism
   claim.** CLAUDE.md's "Diagnosis before debugging" was **corrected on
   2026-07-19 to the opposite of what I acted on** — it is now the DEFAULT for
   exactly what I was doing (a mechanism, a structural property, a fleet-wide
   behaviour change) — and I was working from the session-start copy. Filed
   retroactively as `f6e6a732-d83a-49db-b0e6-4d555249a5f8`. The corrected text
   anticipates the failure precisely: *"Confidence is not a signal."*
2. **I mis-diagnosed a queued council run as a dropped dispatch.** Round 2
   returned no `orchestration_state_audit` rows for ~4 minutes, so I concluded
   the spawn was dropped and resubmitted — it had actually been **queued ~80
   minutes** and ran at 15:52. That cost a **duplicate council run** (real
   credits). *Lesson:* absence of audit rows means queued OR dropped; check
   `diagnosis_artifacts` for a `fix_plan` on the correlation before resubmitting,
   and check cluster load — the pod-restart-drop rule (~300s) does not cover a
   backlog.
3. **I wrote a §9 pattern without first grepping `/bugs_closed/`** (new
   2026-07-19) or the diagnosis queue. Another thread had filed `bugs_open/027`
   hours earlier — the *same mechanism* in the sibling function
   `directionAppliesToKind` (`default: return true`) in the same file. Corroborates
   the pattern, but I should have cross-referenced rather than risked forking a
   second account of it. There is also an existing **"five-place checklist for a
   new kind"** in `HANDOFF_imagery_best_in_class.md` I did not know about — that
   checklist and my routing table are the same idea; whoever next touches kinds
   should reconcile them.
4. **I did not create/update these working docs as I went** — the standing-five
   directive (CLAUDE.md, 2026-07-18, cadence 2026-07-19) says that is part of
   doing the work, not finishing it. Written up only after the owner stopped me.

### Landmines for the next thread

- `agent-chassis` and `image-generator-adapter` must ship **together**: the action
  layer sends `provider_hint`, the adapter consumes it. Shipping one is a half-live
  change. (See §9 "One image tag, two services, different vintages".)
- I built 1137 via `make build-<svc> IMAGE_TAG=v1.0.1137` **without editing the
  makefile**, because another session was actively building 1136 (both images,
  timestamped 4 minutes before my commit — so 1136 does *not* contain this fix).
  CLAUDE.md says bump line ~16; a CLI override achieves the same
  no-stale-cache guarantee without mutating shared state mid-flight. Flagging the
  divergence rather than hiding it.
- `bugs_open/027` and the `content_hero` diagnosis-queue item are live in the same
  two files. Re-read both before editing `generate_image_actions.go` or
  `imagery_style_guide.go`.

### Turn 53 (cont.) — B16.3 executed; B16.1's chosen source does not exist; and two wrong turns of mine

**B16.3 — owner chose "write the three style guides and run without my approval".**
Done and applied live: `SQL_2026-07-19_style_guides_three_sites.sql`. Each site got a
guide with a `kinds.content_hero` override, palettes taken from
`design_intent.palette.reference_values` where pinned (gamesdesign, leopardess) and
from `colour_mood` prose where not (finetuning — marked as such in the row's `notes`).

**Anchors were decided by LOOKING, not assuming** — the workstream's own recurring
lesson:
- **leopardess → `["hero_home"]`.** Downloaded and viewed it: flat antique-gold
  linework on near-black, no text, generous negative space, Banana-generated. Exactly
  the "same hand as the logo" its design_intent asks for.
- **finetuning → `[]`.** Its `hero` turned out to be a teal/charcoal mark on a **pale
  grey ground** — anchoring to it would have imported the very white background D14's
  `avoid` had to be tightened to exclude on robot-hands. Had I assumed "it's their
  hero, anchor to it", I would have shipped that.
- **gamesdesign → `[]`.** All SDXL photographic; and its homepage's own
  `/assets/images/hero.jpg` **404s** (301 → 404) — noted, not chased, not mine.

**WRONG TURN 1 — I diagnosed a header problem, retracted it, and the retraction was
the wrong half.** Fire 1 used a minimal header set (`action`, `message_type`,
`sender_agent_*` only) and produced no orchestration. I concluded headers were
missing and re-fired with the full 033-script set (fire 2). Still nothing after ten
minutes, so I concluded the headers were irrelevant and the real cause was consumer
lag. **Both halves were half-right, and I asserted each too early.** Final state:
fire 1 (minimal) NEVER ran — no orchestration_states row for it, ever. Fires 2 and 3
both COMPLETED. So the headers *did* matter; fire 2 was simply queued behind a
backlog, which made a correct fix look like a failed one. **The lesson is not about
headers — it is that I read a delayed success as a failure and rediagnosed on top of
it.** A6.1's own example elides the header list as `...`; the working set is in
`033_rerender_pages_trigger.sh`.

**WRONG TURN 2 — I called the consumer "stalled" on 75 seconds of evidence.**
`generic-requests-group` on `system.agent.generic.requests` showed CURRENT-OFFSET
frozen at 93404 across three samples while LOG-END-OFFSET grew 93445 → 93490 (lag
43 → 86). That observation was real and is worth knowing — the chassis processes this
topic behind long orchestrations, and a council-gate run (13 LLM seats) can hold it up
for minutes — but **"stalled" was too strong: it drained on its own** and all queued
work ran. Useful command, since nothing in the runbook had it:
```bash
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
  /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group generic-requests-group
```
**Before diagnosing "dead dispatch", check this lag.** A frozen offset with a growing
end means queued, not broken — sample it over minutes, not seconds.

**Result of the pass:** exactly **9 `needs_imagery` / `kind=content_hero` items on
gamesdesign.co.uk**, one per deployed tool page — the count this turn's analysis
predicted. They emit at status `detected`, which is the spend control point: nothing
generates until promoted. Two promoted as a pilot (ehp-calculator, jump-physics)
before releasing the other seven. Dedup held across the repeated fires: 2 passes ran,
9 items, not 18.

**B16.1 — the owner chose "reuse the linked page's card image", and the source it
names does not exist.** Before building anything I read what the 15 live grids
actually contain (86 cards):

| | cards |
|---|---|
| total | 86 |
| no `link_url` at all | 14 |
| linked | 72 |
| …whose URL resolves to no page (matching on `url`, `url+/index.html`, `/name`) | **41** |
| …whose target has a **card asset** | **0** |
| …whose target has a plan hero only | 23 |
| …whose target is a real page with no image at all | 8 |

And **7 of the 15 grids have fewer distinct destinations than cards** —
`idea.uk/report` has **6 cards pointing at 1 URL**, `leopardess/who-we-help` 6 at 2,
and three grids (`ai-agent-orchestration/services`, `gaswholesalers/how-it-works`,
`leopardess/how-we-work`) link nowhere at all.

So implementing the decision literally yields: **zero** cards with a purpose-built card
crop, at most 23 of 86 showing a heavy full-size plan hero, several grids showing the
**same image repeated 4–6 times**, and 41 cards linking to nothing. The premise the
decision rested on — "reuse the card assets we already derive, no new spend" — does not
hold, because card derivation only ever runs for LISTED article and tool pages, and
these cards point at ordinary nav pages (`/about`, `/contact`, `/products`). Taking it
back to the owner rather than building a known-rejected outcome (byte-identical cards
are what failed the D13 gate).

**Incidentally surfaced, not chased:** 41 of 72 card links resolve to no page. That
overlaps `/bugs_open/023` (CTA/link integrity — dead controls, findings dying at
`needs_human_review`) and belongs to that workstream, not this one.

### Turn 53 (cont. 2) — re-pilot PASSED; the truncation fix is proven end-to-end

Re-drove the two pilot items after superseding their assets (A6.5 shape) with the
terse directions in place. **The fix is proven at both ends:**

`assets.origin_prompt` now reads:
> `flat duotone editorial illustration. bold simple silhouette, minimal detail. colour
> palette: cyan #00bcd4 on near-black #121212, light grey accents. Header image for a
> web-based tool represe…`

— the cyan reaches the model, where before the prompt stopped dead at
`colour palette: near-black ground (#121212).`

And the output matches: **both images came back cyan #00bcd4 on near-black, flat
duotone, and — the axis that actually failed the D13 gate — CONSISTENT WITH EACH
OTHER.** Both are also free of lettering, where v1 carried "HP"/"EHP"/"ARMOR" and
"g"/"v".

**Two things I am NOT claiming:**
- **The text disappearance is not cleanly attributable.** I changed two variables at
  once — the terse direction AND a strengthened `avoid` ("words, labels, captions"
  added). n=2. Whether `avoid` even reaches Banana as a negative prompt is still
  UNVERIFIED and remains the next check; if it does not, the text improvement came
  from the shorter positive prompt alone and is luck.
- **Ground drift is not fully solved.** The ehp-calculator image carries pale
  near-white diagonal bands at left and right, despite `avoid` listing "white
  background, pale background". Same family as the D14 white-ground drift, not as
  severe. jump-physics has large flat cyan fills — in-palette, arguably fine.

Sizes 497KB / 406KB — these are HEROES; the ≤60KB budget applies to the derived
cards, so no budget concern.

**Remaining 7 released** to triaged after the pilot passed, per the owner's standing
"write the guides and run" decision. Expect 9 content heroes on gamesdesign, then
cards to derive by origin-staleness on the next pass, then the tool listing needs a
re-render with **`reason='image_landed'`** (A6.2 — the step that bites).

**Still to do after they land:** eyeball the 9 as a set (the D13 gate was about
consistency ACROSS the set, which 2 images cannot prove); then finetuning.uk (5) and
leopardessconsulting.co.uk (5), whose guides are written but whose sweeps have not
been fired.

---

## 2026-07-20 — bugs_open/028 (avoid lists inert): confirmed end-to-end, fixed, council round 1 lost to a harness bug

**Thread:** bugfix-028. Bug file carries the full account (`/bugs_open/028`, §7 is the fix).

### Diagnosis — deliberately NOT sent to the diagnosis loop, and why

CLAUDE.md's default is to file before asserting a durable cause. I judged this one
self-evidencing and skipped the loop. The reasoning, so it can be challenged: `avoid`
has **exactly three references in the entire Go tree** — `imagery_style_guide.go:128`
(emptiness check), `:194`, `:196` (the accessor) — and its only consumer is
`generate_image_actions.go:333` → `negativePrompt` → `banana/provider.go:105`, a
`logger.Debug`. The consumer set is *enumerable and enumerated*; there is no room for
the cause to be elsewhere. That is the narrow condition the guide allows.

What I did instead of the loop was **verify live**, which the original filing had not:

```sql
-- every sampled image was served by the discarding provider
SELECT a.asset_key, a.origin_model FROM assets a JOIN sites s ON s.id=a.site_id
WHERE s.domain LIKE '%gamesdesign%' AND a.created_at > '2026-07-18';
-- → 11/11 rows: banana/gemini-3-pro-image-preview
```

and read `assets.origin_prompt` for `content_hero_tool_xp_curve_designer`: medium,
mood and palette present, **zero terms** of the site's 240-char avoid list. That
upgraded the filing's "verified in code" to proven in prod.

### The fix, and the one contestable decision

`banana/provider.go` now folds `NegativePrompt` into the positive prompt
(`foldNegativeIntoPrompt`) and logs at Info; `provider.Request.NegativePrompt`'s
contract was tightened (it had *licensed* the drop — "providers that don't support
negative prompts log and ignore" — so Banana was obeying it); 7 unit tests added to a
package that had none.

**Placement was the real decision: provider, not action layer.** Three reasons, in
`/bugs_open/028` §7. Short form: negation in the positive prompt helps Gemini and
*hurts* SDXL (CLIP can't negate), so the action layer would need the routing decision,
which lives only in `routing.go` — and a second hand-maintained provider list is the
drift class `routing.go`'s header exists to prevent, and that queued item `5db192c5`
is already filed against. It also keeps the fold downstream of
`maxImageryDirectionInPrompt = 200`, so it cannot evict the palette (027 §4b).

**Cost accepted, recorded so nobody trips on it:** `origin_prompt` is written by the
ACTION layer, so it will never show the avoid terms — fix or no fix. §6 of the bug file
told the next thread to verify exactly there, so **§6 is now corrected**; verify from
the adapter's Info log (`grep "folded NegativePrompt into positive prompt"`) instead.
Rejected adding `FinalPrompt` to `provider.Result`: unconsumed surface, and actually
recording it means repointing `origin_prompt_field` in 4+ workflows (107) and changing
what the column means for every historical row.

### Misstep / correction to our own record

The D14 lesson "ground colour is fixed via `avoid`, not `medium`" (in
`HANDOFF_imagery_best_in_class.md`, the imagery memory, and RUNBOOK A6.5) is
**disproven, not merely doubted** — those generations were Banana-routed, so the
`avoid` edit reached nothing. The mechanism is worth naming because it will recur: a
free config edit plus a re-roll of a nondeterministic generator, improvement credited
to the edit, n=1. `SQL_2026-07-18_d14_avoid_tighten_white_grounds.sql` is that edit.
**Not corrected in those three docs by this thread** (imagery workstream's to make);
flagged in `/bugs_open/028` §4 so it is not lost. Pattern written up in 016b §9
("A field the backend has no parameter for gets accepted and discarded").

### Council gate — round 1 VOID (harness), round 2 pending

`SUBMISSION_CORR=d35844da-f533-42da-b096-4f82cc2839bc`. Round 1
(orch `5438f851`) reached `complete_invalid` with **`error` empty**, `status` COMPLETED,
and no `council_report`. Cause was only in `collected_data.__step_error`:

```
reviewer output at "review_debug_historian.result" does not match the review schema:
json: cannot unmarshal string into Go struct field .objections.edit of type int
```

`councilReview.Objections[].Edit` is a plain `int` (`diagnose_council_decide_action.go:56`)
under a strict unmarshal, so one seat writing `"3"` instead of `3` discarded **every**
seat's review after all had been paid for. **Filed as `/bugs_open/036`.** Distinct from
019 (truncation → invalid JSON, salvageable); this is well-formed JSON with a wrong type,
so 019's bracket-closing salvage cannot help. Did not fix it: `diagnose_council_decide_action.go`
was dirty in the shared tree (another session on the 019 work), and two edits to one file
cannot be separated by a pathspec commit.

Resubmitted with `RESUBMIT_CORR` so the trail accumulates, **consolidated to one edit
entry per file** (round 1 had 5 edits, 3 naming the same file — a plausible pull toward
answering "which edit" with a string). Workaround, not a fix; noted as such in 035 §4.

**Trap for the next thread:** a voided council round records itself as `COMPLETED` with
an empty `error`. Check for the `council_report` row — **absence of the report is the
symptom**, not presence of an error.

### Council round 2: REVISE — and it caught me committing this bug's own failure mode

Verdict `revise` on `d35844da`, `decided_by: objection from editquality`. 7 approve,
3 object (editquality, bug_historian, guardian), **`abstained: 5`** (relevance filter;
the memory's "verify abstained: 0" trap applies to APPROVED rounds, not this one).
**No trailer claimed on any commit** — REVISE never earns one.

The Banana fix drew no objection from any seat. `constitution`: *"a genuine translation
rather than a workaround, and the rejected alternative is explicitly stated with three
concrete reasons"*. `reuse_agent`: *"the rejected alternative is the one that would
actually have produced a second, competing implementation"*. The objections were all on
**edit 2, the interface contract, being comment-only**.

Four challenges, all answered by checking (details + table in `/bugs_open/028` §7b):
SDXL's negative-prompt use **verified** at `stability/provider.go:185–201` (weighted
`api.TextPrompt` — true negative conditioning); **exactly two** `Provider` implementers;
nothing coupled to the old wording, **but the same "log and ignore" licence sat one field
down on `ReferenceImageURIs`** — now rewritten, and the useful distinction recorded: that
discard is fine *because it is loud* (Warn in two layers vs Debug in one). `reuse_agent`
was right that `endsWithSentenceBoundary` already exists in
`generate_image_actions.go:1161`, but importing `platform/orchestration/actions` into a
provider adapter inverts the dependency direction — duplication kept, deliberately.

**The misstep worth recording.** I wrote in the submission that three documents still
assert the disproven "avoid fixes ground colour" lesson and that fixing them was out of
scope. **False — and I had opened none of them.** All three were already corrected before
I started (RUNBOOK A6.5 carries a dated 07-19 correction; the memory carries its own;
HANDOFF had been reframed to *"Unverified and next: whether the Banana path sends `avoid`
as a negative prompt at all"*). editquality filed it as a *missing* item, which is what
sent me to read them. So while fixing a bug that exists because someone asserted a
mechanism instead of checking it, I asserted a document state instead of checking it. The
council caught it; nothing else would have. Corrected in place in 028 §4.

**Left open, deliberately, as an owner call:** editquality + bug_historian (both medium)
hold that a comment cannot stop the next provider repeating the discard —
*"documentation-as-guard has already failed once on this exact field"* — and bug_historian
routes it explicitly to a human while saying it must not block the Banana fix. Three
options in 028 §7b (conformance test needing an injectable client; a capability method on
the shared interface; or accept the comment given only two implementers).

---

## 2026-07-20 ~11:55Z — provider routing (011 R1) PROVEN end-to-end

Resumed the provider-routing thread from `HANDOFF_2026-07-20_provider_routing_011.md`.
The handoff's one missing piece — no real generation had exercised the new routing —
resolved itself before the session started: **eight assets generated on dartsonline.com
(`5fe8785b`) 10:41–10:56Z, 1 `hero` + 7 `icon`, all
`origin_model = 'banana/gemini-3-pro-image-preview'`**. The adapter's own decision is in
its log at `dynamic_adapter.go:569`: `"msg":"generateImage: dispatching to provider",
"kind":"hero","provider":"banana"`. No `UNROUTED KIND` on either replica. Evidence
recorded in `bugs_open/011` §6 (dated update block).

**Misstep worth keeping: the adapter runs TWO replicas** (`-lmp5j`, `-pl6jc`) and all
traffic went through the second. My first log grep hit the first replica, returned
nothing, and for a moment read exactly like "the generations never touched the adapter" —
the same false-negative shape as the strings/`case`-values trap in the handoff. Grep all
replicas of a service before concluding from an empty log.

**Observation handed to the 028 thread, not acted on here:** the hero dispatch carried
`has_negative_prompt:true`, and the running adapter predates `32f2d51e2` (pods up since
07:35, no restart), so that negative prompt still reached Banana's discard path. 028's
fix remains inert until the next image roll — this generation is evidence of the gap, not
a regression.

Concurrent-thread note: while this session was verifying, another thread answered the
cost question in `bugs_open/011` §6 and the resume handoff §4.5 — **owner APPROVED the
routing change**; ~14× per hero but ~+$5/month absolute. Their working-tree edits and
mine share those two files, so whichever session commits first carries the other's
section as a named passenger (CLAUDE.md same-file rule).

Next per the handoff: §4 item 1, the council residual — persist `UnmigratedKind` via
adapter response → action, never by duplicating the routing table in the action layer.

## 2026-07-20 ~12:15Z — 011 §4 residual BUILT: reported_conditions, adapter → chassis → agent_error_log

The council residual (bug_historian, corr e996bf0a: "detection living only in process
logs still depends on someone tailing the right pod") is implemented and submitted as
round 4 on the same correlation (orchestration `e16bfb00-172f-441e-8f3b-753d3df00e31`).
Shape, exactly as the handoff prescribed: the adapter DETECTS against the routing table
its own binary ships and reports a `reported_conditions` list in its success response;
the chassis coordinator PERSISTS each entry to `agent_error_log` (severity `warning`,
`resolved=false`, reporter identity from `response.Headers.Sender`) at
`handleCompleteResponse` — the one point every complete response crosses, so no
workflow-branch coverage gap. Neither service predicts the other's config: the entire
coupling is the field name.

- `routing.go` `reportedConditions()` (pure) covers all three warn-then-forget flags:
  `UNROUTED_IMAGE_KIND`, `UNRECOGNISED_PROVIDER_HINT`, `REFERENCE_ANCHORS_DROPPED`.
- `agent_error_log.go` `parseReportedConditions()` (pure, junk-safe, capped at 10/response)
  + `persistReportedConditions()` reusing `buildErrorEntry`/`logAgentError`.
- Split-roll safe in both directions: old chassis ignores the field, old adapter never
  sends it. Healthy path = absent field, one nil map-lookup per response fleet-wide.
- Tests: 4 new in `routing_test.go` (11 total pass), 4 in new
  `agent_error_log_test.go` — the latter run via `git archive HEAD` + overlay because
  `orchestration_test.go` is broken AT HEAD (pre-existing: `NewSagaCoordinator` grew a
  `storage.Client` param, old external test never updated — not this thread's, noted here
  so nobody re-diagnoses it).
- Checked before claiming (the 028 lesson): `reported_conditions` appears in NO seed,
  config or live `agent_definitions` row; `warning` is already an in-use severity in
  `agent_error_log` (`error`/`warning`/`fatal`).

Uncommitted pending the council verdict; commit will carry the trailer only if APPROVED.

## 2026-07-20 ~14:40Z — round 4 REVISE (decided by editquality), all objections answered by CHECKING; round 5 submitted

Round 4 on corr `e996bf0a` came back REVISE (10 reviewers, 6 abstained). Every
objection was answerable with evidence rather than argument, and two were genuine
catches of my submission (not the code):

- **editquality (medium, the deciding objection):** my sketch changed `generateImage`'s
  signature but never showed the call-site update — "the plan may not compile as
  described". The call site WAS updated in the working tree; I compressed it out of the
  sketch. Lesson repeated from the runbook trap: **on a resubmit, the sketch is the
  reviewable artifact — show the real hunks, not prose summaries of them.**
- **editquality (low):** could `handleCompleteResponse` run twice for one physical
  response (Kafka redelivery) and duplicate rows? CHECKED: no — it is reachable only
  after `ProcessResponse`'s atomic claim (`coordinator.go:226` "This single operation
  prevents duplicate processing"; redelivery → `awaitedReq == nil` → "Duplicate or
  orphaned" → return). At most one persist per physical response.
- **reuse_agent (medium):** did anyone search for an EXISTING adapter→chassis
  non-fatal-condition convention before minting `reported_conditions`? Fair — I had
  searched the DB-placement question, not the convention question. CHECKED: none
  exists; "warnings"/"issues"/"conditions" hits are all local action result maps or
  workflow branching config; `ResponseBody` = {Success, Headers, Body, Error} with
  `ErrorInfo.Details` on the error path only. First convention, not a second.
- **guardian (medium, "resolves to approve if checks clean"):** sole-crossing-point and
  no-other-emitter both CHECKED clean (one call site at `coordinator.go:289`; the
  `:1632` lookalike is `executeLocalAction` = local actions, never on the wire; key
  emitted nowhere else in repo or live agent_definitions).
- **debug_historian (approve-with-objection):** the plan must name its own deploy
  verification — now committed in the risks section: discriminating pod-binary greps
  for strings the change CREATES (`UNROUTED_IMAGE_KIND`, `REFERENCE_ANCHORS_DROPPED`,
  the chassis persist log literal), on BOTH replicas of BOTH services, plus positive
  control, plus a live-fire `agent_error_log` row check.
- **tooling_provenance (low):** on landing, write a doc_note recording the contract
  against `5db192c5` — which stays OPEN (it covers the wider unmatched-case family:
  `directionAppliesToKind`, per-kind accessors).
- **guidelines (approve)** flagged a GUIDELINE-GAP side-task for some future thread:
  DECLARED CONTRACTS has no clause for chassis-parsed generic response fields
  (headers already work this way, undeclared); recommend an explicit exemption clause.

No code changes were needed — round 4's objections were all about evidence and sketch
completeness. Round 5 resubmitted on the same correlation with the checks inline.

## 2026-07-20 ~15:05Z — round 5 REJECTED (guardian hard veto — and it was RIGHT); gate built; round 6 submitted

Round 5 flipped editquality and reuse_agent to approve but drew a **guardian hard
veto**, and the veto is a genuine catch, not process friction: my unconditional
`persistReportedConditions` call in `handleCompleteResponse` made the new channel
**fleet-wide and ungoverned** — any future adapter emitting the key (deliberately or
by copy-paste) would silently start writing `agent_error_log` rows with zero review.
"A foundational-plumbing change wearing a one-adapter bug fix's clothes." My own
DECLARED-CONTRACTS side-task note was correctly read as an admission.

Built the guardian's contained alternative (a) verbatim:
- `conditionReportingAgentTypes = {"image-generator": true}` — an explicit allowlist
  in the CONSUMER; adding a sender is one reviewed line. Unsanctioned sender emitting
  the field → loud named Warn, zero rows. Non-sanctioned pipelines: provably one map
  lookup, nothing else.
- Alternative (b) (gate via `agent_definitions.output_contract`) rejected with stated
  reasons: config lookup in the hottest path; the config re-seed clobber landmine
  would make the gate's ground truth mutable outside code review; the in-code list is
  the direct expression of "whom the chassis believes". Swap seam if architecture
  review disagrees: `senderMayReportConditions`.
- Also fixed bug_historian's medium (the second real catch): parse now returns a
  `malformed` flag — present-but-not-a-list and all-junk-list warn "the reporting
  contract broke; conditions were lost, fix the emitter" instead of reading as
  healthy. Empty list stays benign. Test-pinned.

Six tests now in `agent_error_log_test.go` (incl. the gate), all passing via the
archive-overlay run; builds clean. Round 6 submitted, orchestration `6e2f0018`.

**Lesson for the file:** the council's value this round was exactly what CLAUDE.md's
diagnosis section claims — the wrong shape FELT like the principled one (generic
mechanism, no special-casing), and it took an adversarial seat to name the cost:
generic-by-default in shared plumbing is an architecture decision, and "scoped by
default, reviewed per addition" is the platform-safe shape.

## 2026-07-20 ~16:00Z — round 6 VOIDED by bugs_open/019; round 7 submitted

Round 6 (`6e2f0018`) produced NO verdict: `review_editquality` truncated at
`max_tokens=8000` and the round routed to `complete_invalid` — bugs_open/019's exact
mechanism, still live because its fix (a3b606798) is inert until the next image roll.
Reproduction #3 logged in the 019 case file (committed `58a7c7a8d`). Not worked
around; resubmitted as round 7 (`0430544f`) with the rationale cut to ~60% (the round
history lives on the correlation trail, no need to restate it) and the sketches kept
full per round 4's demand. Monitor now also watches for the void shape, not just the
verdict, after the round-6 monitor burned its 40 minutes timing out on silence.

## Turn 54 — 2026-07-20 — the owner said "think hard again before committing", and that caught a real flaw the council also missed

Council trail for the 027 fix (correlation 0a07f5ed): R1 REVISE (guardian +
debug_historian, verification completeness — answered with fleet enumeration);
one run VOIDED by bugs_open/019 (edit-quality hit max_tokens at 8000; evidence
added to 019's file with a four-point size series); R2 REVISE (editquality caught
my prose claiming a rune guard my sketch didn't contain); R3 REVISE (editquality:
my test never called the function I was changing; reuse_agent: I was hand-rolling
a second truncation while TruncateString sat in datahelpers — checked it, and it
has THE SAME byte-slice bug, so the answer became extract SafeCut and fix both);
R4 REVISE but all-low (8 approve).

**Then the owner asked me to re-think before committing, and the re-think found
what five council rounds had not.** I had never simulated the truncation backoff
— I'd verified the palette SURVIVES and hand-waved what happens to the rest.
Simulated the exact Go logic against all 8 live fleet directions:

- **My round-2 claim to the guardian was FALSE.** "robot-hands gains the full
  palette and loses two mood adjectives" — actually, under the R2–R4 plan it
  keeps ONLY the palette. The backoff cuts at the FIRST '. ' past minKeep
  (`strings.Index`, generate_image_actions.go:1145); with a >100-char palette
  clause leading, that boundary is the END OF THE PALETTE, and medium + mood are
  discarded even though the medium would fit.
- **The bug is latent today and my reorder ARMS it.** Palette-last, the first
  sentence boundary lands early (short medium first) so the first-sentence branch
  almost never fires. Palette-first, it always fires. Fix is one word — Index →
  LastIndex (keep every complete sentence that fits) — plus a fixture that fails
  under Index and passes under LastIndex (robot-hands' own 122-char palette
  clause is the distinguishing shape).
- **An honest trade the earlier rounds under-stated:** all four BASE voices
  (304–398 chars) cannot fit palette AND prose in 200 regardless of ordering or
  backoff. Today they truncate to prose-without-colours; under the plan,
  colours-without-prose. That is a value choice and is now argued openly in the
  submission's risks (colour is the pinned, owner-approved field; the WARN makes
  every over-cap site visible; config shortening is live-immediate).

Round 5 submitted with the correction, the one-word backoff fix, the
distinguishing test, and the round-4 lows answered with data (TruncateString: 31
call sites in 7 files, all log/preview truncations).

**The lesson, plainly: I verified the property I was optimising for and never
simulated the mechanism I was feeding it through.** Nine council seats reviewed
R4 and none simulated it either — reviewers judge the plan against the rationale,
and my rationale contained a confident wrong number. A 40-line simulation against
live data found it in two minutes. Simulate the machine you are about to feed,
not just the property you care about.

## 2026-07-20 ~18:10Z — round 7 REVISE (guardian veto RESOLVED); code went LIVE via someone else's sweep; round 8 submitted

Three things happened at once, and the middle one matters most.

**1. Round 7 = REVISE, but the veto lifted.** Guardian APPROVED — the sender allowlist
answered its round-5 objection. 10 of 12 seats approve. Two mediums left, both real:
- **bug_historian** (again, and again right): the round-5 fix caught "the whole field
  is broken" but a MIXED list still silently dropped junk entries while the survivors
  made the response look healthy. **Same silence, one level down.** Fixed: parse
  returns a `skipped` count, caller warns "partly UNPARSEABLE — some entries were
  dropped and are lost; the surviving ones are not the whole picture". New test
  pins 3-valid + 2-junk → 3 parsed, skipped=2. Committed `8ec9e2ab8`.
- **prior_art_librarian** (DORMANT-MACHINERY): had I checked `diagnosis_artifacts`,
  which has a similar shape? Checked rather than argued — it holds ONLY fix-loop
  artifacts (fix_plan 111, council_report 122, bundle 59, escalation 5), no adapter
  writes it, and of {severity, resolved, work_item_id, site_id} it has only site_id,
  so no `resolved=false` query and no analogue of `idx_error_log_unresolved`. Also
  proved `logAgentError`/`buildErrorEntry` predate this work: both `c24aa7411`,
  2026-03-12, via `git log -S`.

**2. MY UNCOMMITTED CODE WENT TO PRODUCTION IN ANOTHER SESSION'S SWEEP.** `bca5d8255`
("v1.0.1140 - sweep. includes imagegenerator changes, coordinator.go amend") took my
then-uncommitted working tree — routing.go, dynamic_adapter.go, agent_error_log.go,
coordinator.go, tests, NOTES, the submission JSON — into its build, and both services
rolled at ~17:58Z. **This is exactly the hazard CLAUDE.md documents** ("your
uncommitted work is not safe, and this practice does not make it safe"), experienced
from the losing side: I was holding the code back pending a verdict, and holding it
back is precisely what exposed it. Nothing was lost and forward-only holds, but the
lesson is sharper than the doc's: *waiting for a verdict is not a reason to leave work
uncommitted* — commit it narrowly and let the trailer, not the working tree, carry the
review status.
Verified live against running binaries, both replicas of both services, greping
strings this change CREATES (never `case` values): chassis `UNSANCTIONED sender`=1,
`Persisted adapter-reported conditions`=1, `the reporting contract broke`=1 (control
`site provider preference applied`=1); adapter `UNROUTED_IMAGE_KIND`=1,
`REFERENCE_ANCHORS_DROPPED`=1 on `-6df8q` AND `-drwlg` (control `UNROUTED KIND`=1).
**No Council-Reviewed trailer claimed — the verdict is REVISE, and a deploy is not an
approval.** Disclosed unprompted in the round-8 rationale so reviewers know they are
reviewing code that already serves traffic.

**3. Round 8 submitted** (`49512359`) with both objections answered by checks.
The gate is what makes the accidental early ship tolerable: every non-sanctioned
pipeline is provably unaffected, which is the property the guardian insisted on.

> **CORRECTED 2026-07-20 (owner, same evening):** the entry above attributes the
> v1.0.1140 sweep (`bca5d8255`) to "another session". **It was the owner's own manual
> commit, not an agent session.** I inferred an agent because the message read like the
> sweep pattern CLAUDE.md warns about, and I stated that inference as fact — the
> `[INFERRED]` marker the working-docs rules now require is exactly what was missing.
> What caught it: the owner said so directly.
>
> **The lesson survives the correction, and generalises:** it does not matter *who*
> holds the broom. A working tree shared by many sessions AND a human is mutable state;
> "I'll commit once the council approves" leaves finished code exposed to anyone's next
> `git add`. Commit narrowly when coherent; let the trailer, not the tree, carry review
> status. What is now WRONG in the entry above is only the actor — not the mechanism,
> not the exposure, and not the remedy.

### 028 PROVEN END-TO-END and CLOSED — plus the next measurement it hands us

Owner waived the `bugs_open/020` tool-imagery HOLD for one generation. Superseded
`content_hero_tool_xp_curve_designer` (the §3 near-white-ground asset), inserted ONE
`triaged` `needs_imagery` item cloned from the completed row (spec verbatim — do not
hand-roll it), and watched the adapter.

**18:38:53Z, `image-generator-adapter-…-drwlg`:**
```
Banana: folded NegativePrompt into positive prompt as a prohibition clause
kind=content_hero  prompt_len_before=517  prompt_len_after=905
negative_prompt = text, watermark, signature, low quality, blurry, distorted,
  photorealism, … numerals, … white background, pale background, bright full-bleed colour field
```
Both halves present — `kindDefaults["content_hero"]` **and** the site's
`kinds.content_hero.avoid` — including `numerals` and `white background`, the two terms
whose violation caused the filing. **028 moved to `/bugs_closed/`.**

**The §6 trade-off is now measured, not predicted.** Same asset's stored
`assets.origin_prompt` = **515 chars, containing neither term**, while the model received
**905 chars containing both**. Anyone verifying via `origin_prompt` would conclude the fix
is dead. **Adapter log only.**

**THE NEXT MEASUREMENT, and it is ours, not the bugfix thread's.** Closing 028 proves the
avoid terms are *delivered*. It proves nothing about whether Gemini *obeys* them, and n=1
cannot: 5 of the original 9 complied by luck, which is exactly what hid the defect for a
release. The honest prior is discouraging — this same asset had `near-black #121212` in its
POSITIVE prompt on 07-19 and still came back pale. **So: generate 5+ content heroes on
gamesdesign and COUNT violations against the list** (white/pale grounds, numerals,
lettering). If the violation rate is materially below 4-in-9, the prohibition clause is
earning its place; if it is not, the wording needs tuning (it is a first attempt) or the
constraint needs a different instrument. Either way that is a real finding about how our
imagery direction reaches the model, and it is now cheap to run.

**Getting one generation to run cost ~18 minutes and crossed two open bugs.** The item sat
at `triaged` looking perfectly dispatchable. `bugs_open/029` was genuinely present (2
`build-pipeline-trigger` hung at `spawn_dispatch`; applied its documented recovery, 2
cancelled) — **and it was not the blocker.** `bugs_open/030` was: one partition, one
consumer, `LAG` 51→67, consumer offset frozen at 95919 while the chassis chewed a council
run. **I misread that frozen offset as a dead consumer and nearly filed a fleet-wide
outage.** It cleared by itself when the offset jumped 95919→95935 — 16 messages at once,
because draining here is a jump, not a smooth advance. Both landmines are recorded in
`bugs_open/030` § Landmines. Practical rule for this workstream: **a triaged imagery item
that will not dispatch is almost always 030; check `LAG` before touching anything.**

---

### 2026-07-21 — §4b fix is LIVE in v1.0.1144, verified in the pod; landing gate still untouched (session "bugfix 027")

**The roll happened.** `1191cecdb` (the §4b palette-first + LastIndex-backoff fix) is an
ancestor of the `f9dfa0205` v1.0.1144 build commit, and v1.0.1144 is the tag deployed to
the running `agent-chassis` and `image-generator-adapter` pods (up ~46m at check). So the
bug file's "INERT UNTIL AN IMAGE ROLL" is now stale — corrected in place there.

**Verified in the running binary, not the tag** (016b §9 / A6.3 — log-literal grep, with
controls):
- chassis pod `agent-chassis-59c675c4f-pxr9f`: `Imagery direction TRUNCATED before
  generation` → **1** (the marker `1191cecdb` created); positive control `Prepended
  imagery direction` → 1; negative control (a nonsense literal) → 0.
- adapter pod `image-generator-adapter-5fb5dbf7c5-7qj9p`: `folded NegativePrompt into
  positive prompt` → **1** (the 028 marker); positive control `Banana image generation
  succeeded` → 1. So 028's fold is also live in this same roll — consistent with its
  already-CLOSED status; nothing to do there.

**Composition proof on robot-hands' real fields — the palette now survives (no credits
spent).** robot-hands' `kinds.content_hero` composes to 233 chars (over the 200 cap; it is
the only over-cap site — the three seeded guides are 139–147 and never truncate). Running
the live logic by hand on the actual DB values:
- `palette` = `deep charcoal ground, electric blue (#0080FF) flat shapes and linework,
  light grey secondary accents only` (105 chars);
- new order `colour palette: <palette>. <medium>. <mood>`; `LastIndex(". ")` within 200
  lands after `medium`, so the truncated direction =
  `colour palette: deep charcoal ground, electric blue (#0080FF) flat shapes and linework,
  light grey secondary accents only. flat duotone editorial illustration.` — **palette
  intact, `electric blue (#0080FF)` retained**, only `mood` dropped, and `truncated=true`
  so the WARN fires. Under the pre-fix order (palette LAST) this same cut would have kept
  medium+mood and lost the colour. This matches the shipped unit fixtures; it is the same
  result against production data.

**Base-voice trade re-measured, and it is now "acceptable", not a live break.** Base-guide
composed lengths (guide-level medium/mood/palette, the fallback for photographic kinds with
no per-kind override): robot-hands 398, gamesdesign 352, leopardess 305, finetuning 304 —
all over 200. But palette-first means every one of them now **keeps the palette** and clips
only trailing prose. The config trim (shorten the palette glosses so prose also fits) is an
enhancement, not a fix for a live defect — the colour is no longer the casualty.

**Landing gate: STILL UNTOUCHED.** Nothing has regenerated since the deploy. The
/bugs_closed/ bar (fixed AND proven on a page) is not met: it needs robot-hands' 3 ARTICLE
content heroes regenerated vs D13 + ≥5 observed generations. That spends real credits on a
live production site and the 3 TOOL heroes stay behind the bugs 020 owner hold — so it is
an owner-go step, not something to fire unilaterally. Left as the next decision.

**Dispatch state at check** (for whoever runs the gate): across the four sites,
`needs_content_image` failed ×9, `needs_imagery` failed ×1, `wont_fix` ×3 — pre-existing,
not created by this session. Per the 030 rule above, check `LAG` before firing anything.

---

### 2026-07-21 (later) — §5(a) structural fix BUILT (option i, the exclude form), `5e19fd3cb`

Owner picked the minimal §5(a) route: exclude `content_hero` from the photographic
free-text fallback rather than invent a fleet-default flat-illustration direction (that
would be option ii, and it needs a brand decision that is the owner's).

**The trap I nearly stepped in, and the reason the fix is two changes not one.** §1's
root cause is stated as "D14 added content_hero to `directionForKind` but not to
`directionAppliesToKind`" — an *asymmetry between two sibling functions*. Reading only
that line, the obvious fix is one line in `directionAppliesToKind`. But tracing every
path a photographic direction can reach a flat `content_hero` shows **three**, and that
one-line fix closes only two:
1. free-text `design_intent.imagery_direction` fallback (guide-LESS site) — gated by
   `directionAppliesToKind`. ✓ closed by the exclusion.
2. `directionForKind`'s `default:` case = the photographic **base voice**
   `composeDirection(g.Medium, g.Mood, g.Palette)` (guide site, no `content_hero`
   override). ✗ NOT closed by touching `directionAppliesToKind` — it is a different
   function. This is the twin hole; fixing only #1 would have re-created §1's exact
   asymmetry, just moved by one function.
3. `referenceKeysForKind`'s guide-level anchor fallback — also gated by
   `directionAppliesToKind` (line 226). ✓ closed as a side effect.

So the fix moves `content_hero` into the flat-vector branch of **both** functions:
excluded from `directionAppliesToKind`, and palette-only (like icon/sprite_sheet) in
`directionForKind`'s switch. No current site is affected — all four with a guide carry a
`content_hero` override and hit the override path (line 171) in both. It is future-proofing
for the next site that has a guide but forgets the override, or has no guide at all.

**Verified**: `go build` + `go vet` clean on the package; `TestDirectionAppliesToKind`
(new) and the extended `TestStyleGuideDirectionForKind` green, alongside the existing
§4b fixtures. Inert until an image roll (Go change). **Not council-reviewed** — it is
council-gate-eligible (fleet-wide generation behaviour); left as an offered next step
rather than spending a council run unprompted.

---

### 2026-07-24 — 027 residue clearance (session "bugfix 027"): §5(a) LIVE, base voices trimmed, landing gate FIRED with owner go

Owner asked for 027 to be researched and fixed; approved BOTH credit spends (article
landing gate + queued-item drain). Who-owns check first: nobody else on 027 (last bug
commit 07-21; today's imagery activity was the 020 hold-lift, a different bug).

- **§5(a) (`5e19fd3cb`) is LIVE in v1.0.1155 — by ancestry, not symbol grep.** The
  diff creates no unique retained literal and §6/A6.3 forbid the `content_hero` grep,
  so: `reconcile_superseded_reviews` (first added `8fd1e3bfc`, 2026-07-22, a strict
  descendant of `5e19fd3cb`) is in the running pod binary ×3, WARN literal positive
  control ×1, absent-string negative control 0. Build point ≥ 07-22 ⟹ contains the
  07-21 fix. Changed branch stays latent (all guide sites carry overrides) —
  unit-pinned only, recorded as such.
- **The stale §4b "avoid UNVERIFIED" paragraph corrected in the bug file**: it became
  `bugs_closed/028` (fold-into-positive-prompt fix `32f2d51e2`, live v1.0.1140,
  proven end-to-end). Traced the full chain again this session to be sure:
  action layer assembles `negative_prompt` (kind defaults + input constraints +
  guide avoid) → adapter → Banana provider folds it into the positive prompt
  (`banana/provider.go:121-134`); Stability keeps true `Weight:-1.0`. Delivered ≠
  obeyed: lettering can still appear (Gemini).
- **Base voices under the cap** (`SQL_2026-07-24_base_voices_under_cap.sql`): the
  promised trim of the three authored guides' ROOT voices — 304/352/305 → 196/189/190
  (palette-first order). Needle-gated on current `mood` text (all three hit UPDATE 1);
  backup `site_specs_imagery_guide_backup_20260724` (6 rows). robot-hands untouched
  (base 398 + 233 override both the owner's / gate testbed).
- **Landing gate fired**: superseded the 3 ACTIVE article guide-hero assets
  (grip_force/payload/cycle_time `_guide`), fired design-discovery on robot-hands
  (detached kcat, produce VERIFIED on the topic — A6.1 gotcha respected). The 3
  queued `image_url_404` items (pages referencing never-deployed guide hero JPGs)
  are the symptom this regeneration resolves. Judging vs D13: distinct, on-style
  (electric blue #0080FF retained — the §4b proof), click-through matching, cards
  ≤60KB. Truncation WARN EXPECTED for robot-hands (233-char override, mood tail
  drops, palette survives).
- **Council retro-review of `5e19fd3cb` submitted** (corr `7388a068`), per the
  2026-07-24 strengthened-advisory ruling. Verdict pending at time of writing.

---

### 2026-07-24 — 011 residual round 9: over-cap truncation counted; owed items 3+4 done (bugfix-011 session)

Ownership check before touching anything: who-owns → imagery (last 011 activity
2026-07-20/21), no open work items in the territory, no round-9 submission
existed, no uncommitted edits on the files. Clear.

**Round 8 read in full** (council_report 2026-07-20 19:15Z): **11 approve, 1
objection** — bug_historian (medium): `parseReportedConditions` capped via a
bare `break`, over-cap WELL-FORMED entries dropped with `skipped=0` — the same
silence rounds 5/7 cured, re-entering at the cap boundary; plus a named MISSING
test (length > cap, all well-formed, truncation counted distinctly). Note the
committed code already had a caller-side truncation warn (8ec9e2ab8 re-derived
it from the raw list) — but the round-8 SKETCH showed the bare break. Same
lesson as relojistas 07-24: **sketches must be FINAL-state**; the reviewers can
only see the sketch.

**Fix, committed `2f4fc0596`:** `parseReportedConditions` returns a 4th value
`truncated` (entries never examined because the cap fired — indexed loop,
`truncated = len(list) - i` at the break); `persistReportedConditions` warns on
it distinctly (truncated/persisted/cap) and the caller-side re-derivation is
DELETED. truncated deliberately NOT folded into skipped — different remedy
(raise cap / report less, vs fix the emitter). The distinction is now
absent ≠ malformed ≠ partly-dropped ≠ over-cap. New
`TestParseReportedConditionsOverCapTruncationIsCountedDistinctly` (15
well-formed → 10/0/5) is the historian's missing test verbatim; capped + mixed
tests now pin both counts in both directions. 8/8 green — **against
`git archive HEAD` + the two files**, because the tree's
`orchestration_test.go` (external test package) has a pre-existing
`NewSagaCoordinator` signature break at clean HEAD, unrelated.

**Round 9 submitted** with `RESUBMIT_CORR=e996bf0a` (submission JSON committed:
`submission_011_residual_round9.json`, final-state hunks extracted
programmatically from the committed files, not retyped). Verdict pending at
write time.

**Owed item 3 done:** doc_note `5842d7ed` (subject pipeline/image-generation,
source bugfix-011, source_item_id `5db192c5`) records the contract: allowlist
load-bearing, the four distinct parse states, scope = only the
routing-observability member of the unmatched-case family. NOTE: item
`5db192c5` is now status **complete** (the bug file's "stays open" is stale —
it completed sometime after 2026-07-20).

**Owed item 4 done:** DECLARED CONTRACTS exemption clause (the guidelines seat
asked twice) — `PATCH_fix_proposer_011_guidelines_declared_contracts_exemption.sql`
applied (snapshot `f9d90a2d`), mirrored to council-gate via 099 (drift was
exactly `review_guidelines`, verified 1 occurrence on both councils).

**Live-fire proof in flight:** the 021-style harness turned out simpler than
planned — no scratch agent_definition needed. The existing `image-generator`
agent IS the 1-step harness (workflow: generate → complete), so the kcat
`orchestrate` envelope targets it directly with
`input_data:{kind:"scratch_unrouted_011", prompt:"a plain flat mid-grey
square…"}`, NO site_id (skips direction enrichment and provider hint — cleanest
possible path to the adapter). Pre-flight: pods 68 min up (300s rule ok);
chassis binary carries "Persisted adapter-reported conditions" + "UNSANCTIONED
sender" =1 each; both adapter replicas carry UNROUTED_IMAGE_KIND=1; my new
"TRUNCATED at the per-response cap" string =0 on the live chassis (correct —
2f4fc0596 not yet rolled, good negative control). Envelope published ~21:15Z
(orch name `scratch-011-livefire-211539`); row not yet visible at +2 min —
queued, not dropped (the 30-min dispatch trap; do NOT resubmit).

### 2026-07-25 — 027 landing gate PASSED; bug CLOSED to bugs_closed/

Continuation of yesterday's session (same thread, overnight pause). What the
morning found and did:

- **Council verdict on `5e19fd3cb`: APPROVED** (corr `7388a068`, council_report
  20:45 UTC 07-24). Trap for the log: the latest doc_notes council-gate row was
  ANOTHER thread's REVISE (corr `45664479`, bugfix-015) — resolve verdicts by
  correlation in `diagnosis_artifacts`, never by "latest note".
- **Discovery had emitted 4 needs_imagery items** at 20:46 UTC (the 3 superseded
  article guide-heroes + a bonus: `tool-matchmatrix` had no hero of its own).
  They sat `detected` overnight — promotion is a manual gate. Promoted all 4 to
  `triaged` priority 5 at 08:35 UTC; dispatch picked them up within ~20 min.
- **Generations 08:56–09:03 UTC**, one per ephemeral `agent-image-generator-*`
  pod, all `banana/gemini-3-pro-image-preview`, all `active`. NOTE the log
  location trap: the direction lines are NOT in the long-lived `agent-chassis`
  pod (generic orchestration only) and NOT in `image-generator-adapter` — they
  are in the ephemeral spawned `agent-image-generator-*` pods, which garbage-
  collect quickly. Grep them FAST or lose the evidence.
- **Delivery proof (adapter-log layer)**: each pod logged
  `Imagery direction TRUNCATED before generation` (`direction_len:233, cap:200`)
  then `Prepended imagery direction` `source:"+style_guide"`, `truncated:true`,
  `direction_preview` = `colour palette: deep charcoal ground, electric blue
  (#0080FF) flat shapes and linework, light grey s…` — palette survives first,
  mood tail drops. §4b's palette-first order observed working on real fleet
  generations for the first time.
- **D13: PASS ×4** (viewed full-size): flat duotone, charcoal ground, #0080FF +
  light grey only, no invented accents, no lettering, subjects distinct and
  click-through-matching. Count: 4 fresh + 3 tool heroes of 07-18 = 7 ≥ 5.
- **Live pages**: all four JPGs 200 on robot-hands.com (63–116KB, precedent band
  58–93KB; ≤60KB is the CARD budget and nothing cards these heroes —
  learning-center.html checked). Three guide pages embed them as hero
  `background-image`. All four `image_url_404` items marked complete with an
  evidence note.
- **CLOSED**: `git mv` → `bugs_closed/027_…`; closure block in the file holds
  the full evidence chain. Residuals recorded there (truncation-signal
  persistence, TruncateString sweep, §5(a) latent branch) — follow-ups, not
  reproductions.

### 2026-07-25 — 011 CLOSED to bugs_closed/: round 9 APPROVED, live-fire landed, R2/R3/R4 migrated (session "bugfix 011")

Ownership check first: `who-owns.py 011` → OWNED by imagery (35 mentions, 86
commits/14d). Read `HANDOFF_2026-07-20_provider_routing_011.md` and this file's
07-24 entry before touching anything. No open work items in the territory (the
15 open `needs_imagery` rows are today's webdesign.co.uk icons, unrelated), and
`git status` clean on all five 011 code files — no other session mid-edit.

**Nothing was built today.** Everything owed had either landed or was in flight
when the 07-24 session stopped; the work was checking whether it arrived.

- **Round 9: APPROVED.** `diagnosis_artifacts` corr `e996bf0a`, `council_report`
  20:33:34Z on 07-24: `decided_by: "all reviewers approve"`, 8 reviewers,
  **0 objections**. `abstained:8` is the 16-seat relevance filter, not silence —
  the standing reading trap. Resolve by correlation, never "latest note" (the
  07-25 morning entry above records another thread's REVISE sitting on top).
- **`2f4fc0596` is now LIVE** — it was inert at commit time on 07-24. Chassis pod
  `agent-chassis-774877f4c6-zjh4t` (88 min up): `"TRUNCATED at the per-response
  cap"` = 1 (the string the change CREATES), `"Persisted adapter-reported
  conditions"` = 1 as positive control.
- **Live-fire proof FIRED** — §7's owed item 2, deferred since 07-20 because it
  costs a real generation. The 07-24 envelope landed:
  `agent_error_log` row `UNROUTED_IMAGE_KIND` / warning / image-generator /
  **2026-07-24 20:45:57Z**, context `{"kind":"scratch_unrouted_011",
  "provider":"stability","routed_kinds":[7 kinds]}`. Published ~20:15Z → landed
  20:45Z: **exactly the ~30-minute dispatch latency**, and the 07-24 session was
  right not to resubmit at +2 minutes.
- **The strongest R1 proof turned out to be a zero.** `SELECT count(*) … FROM
  assets WHERE origin_model ILIKE '%stability%' AND status='active'` → 60, of
  which **0 created on or after 2026-07-18**; newest is 2026-07-17. Seven days,
  every site, every kind, every dispatch path: nothing has reached the weaker
  provider. Two single-generation proofs (07-20 dartsonline, 07-25 robot-hands)
  could not say that; one aggregate does.

**What I found that nobody had checked since 07-18: the bug's own exhibit is
still live.** `https://leopardessconsulting.co.uk/assets/images/hero.jpg` → 200,
143,819 bytes. **Downloaded it and looked at it** — a gold/charcoal flowchart in
which every label is gibberish, exactly as §1 describes. Referenced by ONE live
page, `how-it-works.html:300`, as the hero `background-image` behind a 50–60%
black gradient; the other five pages checked don't reference it. Asset row:
`asset_key='hero'`, `stability/…`, `active`, created 2026-07-17 22:46Z — the day
*before* the fix. **A stale artefact, not a reproduction**: regenerating it needs
no code, because `hero` routes to Banana now. Deliberately NOT regenerated —
that is a production action on another workstream's client site plus a page
re-render, so it is flagged to the owner in `README_where_we_are.md` and written
into `features_open/022` as the sweep's first confirmed target.

**R2/R3/R4 migrated rather than dropped.** They were never routing defects, and
keeping them in the bug file would have held a fixed bug open indefinitely.
`features_open/022` = the legibility guard, carrying the 60-asset legacy corpus
and the constraint set the council imposed (adapter reports → chassis persists;
the sender allowlist IS the review; absent ≠ malformed ≠ partly-dropped ≠
over-cap). `features_open/023` = evidence-based figures + the SVG boundary; it
records that **only 8 sites have an `evidence_base` at all** (8 current rows,
17 all-versions), which is the design question that feature turns on and which
the bug file never measured.

**A structural gap worth naming:** the council verdict (20:33Z) post-dates the
commit it approved (`2f4fc0596`, 20:10Z UTC). Forward-only means no amend, so no
`Council-Reviewed:` trailer can ever be attached, and `098` will bucket it
UNREVIEWED. That is a **false negative, not a missed review** — any commit made
while its own council round runs is unattributable by design. Recorded in the
closed bug's §8.2 with the correlation to quote.

**Closed**: `git mv` → `bugs_closed/011_…`; `016b` §10 row and its three
`bugs_open/011` cross-refs updated to `bugs_closed`.

## 2026-08-11 — cross-lane notice from bugfix 214: what changed under your dispatch path (written by the `bugfix_214_imagery_scope_ref` lane, per the owner ruling 2026-07-29 §3 — consumers must be TOLD, not merely measured)

`WriteSitePlanAction` now canonicalises `site_plan_imagery.scope_ref` against the
plan's own page names at write time (live on chassis `v1.0.1283`, 2026-08-10;
`bugs_open/214`, register `IMG-070`). Three things change about the guarantee your
pipeline reads, none of them yet exercised in production — no site with renamed
(`-index`) pages has replanned since the roll:

1. **Pass-1 membership of your emitter can shift once, per site, on the first
   replan-with-renames.** `imageryplan.LoadCurrentPlan` orders by `scope_ref` with
   `MaxPerPass=20`; canonicalising refs re-sorts them. Only plans over 20
   page+section rows can notice: measured 2026-08-11 that is **robot-hands.com
   (27), fundamentallyai.com (23), vonc.com (22)**. Overflow is picked up by later
   passes, so this is transient reordering, not loss.
2. **`needs_imagery` ItemKeys minted from pre-fix refs drift on that same replan**
   — the key embeds the ref, so a rewritten ref mints a NEW key while the old open
   item still carries the old one. Live today: mortgagecalculator.co.uk's deferred
   `about`/`contact` items name refs its current plan spells `about-index`/
   `contact-index` (that site's whole queue is deferred and its assets were never
   generated, so nothing is stuck *now*). Whether to re-key, cancel or leave open
   items when their site replans is **your lane's call** — this note is so the
   choice is made knowingly, not discovered as a dedup mystery.
3. **Human-approved imagery locks now transfer across a rename for one
   generation** — `transferImageryLocks` gained a canonical fallback that fires
   only when the exact `(plan_id, scope, scope_ref, key)` match finds nothing.

Full decision record: `doc_notes` id `0633aa2f-cdf6-4d3f-afdf-9496ee694af1`
(`subject_type='action'`, `subject_key='write_site_plan'`). Questions or
objections → `docs024_key_docs_latest/bugfix_214_imagery_scope_ref/`, or the
council trail on corr `46a50b4c`.

---

## 2026-08-24 — note from the `bugs_open/382` lane: 011's routing table changed, and so did image-build-handler's variant branch

Written into this lane's log rather than only into my own, because this lane owns
`internal/adapters/imagegenerator/routing.go` and `image-build-handler`, and I have changed
both. Nothing is asked of you; this is so the change is not a surprise the next time this lane
picks up.

**What was wrong.** 011's routing table exempted one branch by design: an EMPTY `kind` routed to
Stability and reported nothing, because *"legacy callers that predate the field are a documented,
deliberate Stability path, not an oversight"*. That premise was never measured. It is false.
**As of 2026-08-24, 15 hero assets on 5 sites had been generated by SDXL after 011's fix** (latest
2026-08-11), **none** of those sites carried the sanctioned `provider:"stability"` opt-out, and
**0** live agent definitions pinned a Stability model.

**Where the empty kind came from.** `image-build-handler.call_variant_gen` — the only handler of
`unfulfilled_hero_variant`, i.e. every per-page hero — forwarded no `kind` and no `site_id`, while
carrying a `default_kind: "hero"` config key that is read by nothing (a `call_agent` step's callee
receives `input_mapping` and nothing else). Migration `390` found this class on 2026-08-11, fixed
`call_hero_gen`/`call_logo_gen`, and then stated in its own blast-radius paragraph that
`call_variant_gen` *"already forward[s] kind"*. It did not.

**What changed.**
- **Migration `586`** (APPLIED + RECORDED 2026-08-24, live on apply): `call_variant_gen` gains
  `"kind?": "input_data.spec.purpose"` and `"site_id": "site_record.site_id"`; all **three** dead
  `default_kind` keys deleted. **Note the second half — hero VARIANTS now reach
  `getImageryStyleGuideForSite` for the first time**, so per-kind overrides, `avoid` terms,
  reference anchors and `design_intent.imagery_direction` apply to them. That is this lane's own
  D14 machinery becoming reachable on a path it never covered; expect variant output to change.
- **Commit `da21ae20f`** (inert until the image-generator adapter rolls): `routeProvider` sends an
  absent `kind` to **Banana** and sets a new `MissingKind` flag, with a `MISSING_IMAGE_KIND`
  condition joining the existing three in `reportedConditions`. `UnmigratedKind` is untouched, and
  an explicit `provider:"stability"` still wins — pinned by a test, since that escape hatch is
  what the change leans on.
- `TestRouteProviderEmptyKindIsLegacyNotUnmigrated` **was replaced**, not extended: it asserted
  the defect. Its old body is quoted in the successor's comment so the inversion reads as a
  correction rather than a weakened test.
- 016b §9's *"Getting the guard right"* bullet, which prescribed this exemption to every future
  session, carries a dated correction.

**Still open, and it is this lane's territory if it wants it:** `pageflow-builder` and
`site-work-orchestrator` each carry `generate_hero_image` and `call_logo_generation` with **no
kind at all** — four steps. They cannot be fixed the way `call_variant_gen` was, because
`input_mapping` resolves data paths and not literals, and their collected_data holds no field
whose value is the string `hero` or `logo`. Their reachability is **[UNMEASURED]** beyond a
1-day `orchestration_states` window. The code change makes them harmless; it does not make them
correct.

Full account: `bugs_open/382` §7–§8 and
`docs/agent_docs/docs024_key_docs_latest/bugfix_382_empty_kind_routing/`.
Council-Submitted: `e53f57ae-3bb1-442c-8e7b-742a1c2bb0ad`.

> **UPDATE 2026-08-24, later the same day — the entry above says "inert until the image-generator
> adapter rolls". IT HAS ROLLED. `da21ae20f` is LIVE.** Written as an addition rather than an edit
> to the paragraph, but flagged loudly here because an "inert until the roll" line is one of this
> estate's recorded traps in its own right: it makes the correct next action look premature, and a
> detector once sat switched off for nine days after its blocker cleared for exactly that reason.
>
> Proof, per SERVICE and at the artefact: `v1.0.1334`, both `image-generator-adapter` replicas on
> ONE digest `sha256:d7a1d219…`, started 2026-08-24 15:39:36Z; the service's own provenance line
> says `70fd163c2`; `git merge-base --is-ancestor da21ae20f 70fd163c2` → YES with both controls
> behaving; and `MISSING_IMAGE_KIND` is present in `/proc/1/exe` with a method control present and
> a fake needle absent. **Label trap for whoever repeats this: `-l app=image-generator` returns
> NOTHING — the deployment is `image-generator-adapter`.**
>
> `bugs_closed/382` is CLOSED. The residual left on this lane's doorstep is unchanged and is in
> §10c of that file: `pageflow-builder` and `site-work-orchestrator` still carry four image steps
> with no `kind`, they cannot be fixed in config (`input_mapping` resolves paths, not literals),
> and their reachability is still UNMEASURED beyond a 1-day window. They are now harmless — an
> absent kind gets Banana — and LOUD, so if they ever run, `MISSING_IMAGE_KIND` will say so with
> the prompt's opening words attached.

---

## 2026-09-04 — lane picked up after 11 days idle; the 382 residual is now MEASURED, and three other things fell out of measuring it

Lane state on arrival: last own-work commit `07f3a3966` (2026-07-25, closing 011).
The only two commits since are other lanes writing *into* this log — 214 (08-11)
and 382 (08-24). `scripts/who-owns.py imagery` shows the two imagery-numbered bug
files owned elsewhere (114 lane active 09-03, 214 lane closed 08-11). Nothing was
in flight here, so I took it.

### 1. The residual 382 left on this lane's doorstep: ANSWERED

382 §10c handed this lane *"`pageflow-builder` / `site-work-orchestrator` — 4 steps
with no `kind`, reachability **UNMEASURED** beyond 1 day"*. `orchestration_states` is
still a 1-day window (`[MEASURED 2026-09-04]` oldest row 2026-09-03 11:47Z), so that
route is as closed as it was. But **`assets.origin_model` is durable and is the
fingerprint of the very path in question** — until 2026-08-24 a kind-less request was
routed to Stability, so every traversal of these steps left an SDXL row behind.

**The four steps still exist, unchanged** `[MEASURED 2026-09-04]`, read from the live
rows (note `input_mapping` is under `config`, not at the step root — my first census
returned four blank columns because I read the wrong path):

```sql
SELECT a.type, s.key, s.value->'config'->'input_mapping' ? 'kind' AS kind,
       s.value->'config'->'input_mapping' ? 'kind?' AS optkind
  FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s
 WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
   AND s.value->>'action'='call_agent'
   AND s.value->'config'->>'agent_type' IN ('image-generator','image-build-handler');
```
→ `pageflow-builder.{call_logo_generation,generate_hero_image}` and
`site-work-orchestrator.{call_logo_generation,generate_hero_image}` are f/f;
`image-build-handler`'s four are all f/**t** (390 + 586 did their job).

**The provider census `[MEASURED 2026-09-04]`, every generated asset since 2026-07-18**
(`derived-from-hero`/`derived-from-logo` excluded — they are card derivations, not
generations; enumerating `origin_model` rather than regexing it is what showed me that,
after a first pass where 291 of 606 rows in one week matched neither arm of my regex):

| week | SDXL | banana | total gen |
|---|---|---|---|
| 07-13 → 08-02 | **0** | 130 | 154 |
| 08-03 | 10 | 64 | 92 |
| 08-10 | 6 | 74 | 110 |
| 08-17 | **0** | 76 | 101 |
| 08-24 | **0** | 315 | 606 |
| 08-31 → 09-04 | **0** | 151 | 226 |

**SDXL generation ceased on 2026-08-11 and has not resumed in 24 days**, across
**1,046** generated assets on **37** sites — a demand control an order of magnitude
larger than the 16 SDXL assets it is being compared against.

**Control that the SDXL rows were the silent fallback and not a sanctioned pin:**
`[MEASURED 2026-09-04]` **14** sites carry a current `imagery_style_guide` spec (aspect
`imagery_style_guide`, its own row — *not* nested under `design_intent`, which is where
I first looked and got a false zero of 0 guides); **1** sets `provider` at all
(idea.uk → `banana`); **0** pin stability. So none of the 16 was opted in.

**Verdict: reachability is no longer UNMEASURED. The four kind-less steps have not
traversed in 24 days under heavy demand.** That is not "unreachable" — it is a
measured quiet period with a stated instrument, and the instrument (`origin_model`)
is durable rather than windowed, so it can be re-run at any future date.

### 2. …and the corollary nobody had noticed: the 382 fix has NEVER fired in production

The SDXL stop is **2026-08-11**. `da21ae20f` — the commit that routes a kind-less
request to Banana and raises `MISSING_IMAGE_KIND` — rolled **2026-08-24 15:39Z**,
**thirteen days later**. The traffic it was built to catch had already stopped.

What stopped it was **migration 390** (*"found 2026-08-11"*), which put `kind?` on
`image-build-handler.call_hero_gen`/`call_logo_gen`. So the 16 SDXL assets were
**image-build-handler's**, not these four steps'; 390 closed that source, and the
routing fix inherited a quiet estate.

This is consistent with, and sharpens, 382 §10c's own disclaimer (*"the new branch has
not been observed executing in production, and cannot be until a caller omits `kind`
again"*). It now has a date attached: **no caller has omitted `kind` since 2026-08-11**.

⚠ **So the zero I report in §1 rests on the SDXL census, NOT on the absence of
`MISSING_IMAGE_KIND` rows** — and it must, because that detector has no positive
control available in-window (see §4). Two independent instruments would have been
better; I have one durable one and one I cannot vouch for, and I am not treating the
second as evidence.

### 3. Migration 390 is APPLIED but UNRECORDED — a clean hole in `schema_migrations`

Found while dating the SDXL stop. `[MEASURED 2026-09-04]` `schema_migrations` holds
**565** rows. `385`–`389` and `391`–`395` are all present, all applied 2026-08-11.
**`390` is absent under every spelling** (`LIKE '%390%'`, `ILIKE '%forward_kind%'`,
`ILIKE '%legacy_image%'` → only 586 matches). Its effect *is* live, so it ran.

Consequence, per CLAUDE.md's migration practice (*"`--apply` takes EVERY pending
file"*): a runner pass today would list 390 as pending and re-apply it. **Re-application
looks harmless** — it is an `UPDATE` followed by a `RAISE EXCEPTION` assertion
(line 58) that the post-state is exactly 1 row with `kind?` on both branches, which
already holds — but it is a surprise waiting for whoever next runs the runner, and the
hole is the kind of thing that reads as "390 was never applied" to a future reader.
**[UNVERIFIED]** why the row was never written; I did not run the runner to find out.

Also worth recording because it is this lane's own file: 390 line 27 asserts
*"(call_imagery_gen, call_variant_gen) already forward kind"*. That was false for
`call_variant_gen` — 382 caught it on 08-24 and 586 fixed it. The claim and its
refutation are now both in the record.

### 4. The 011 closure's headline proof has EXPIRED — and this log is one of the places citing it

011 was closed partly on a live-fire: an `agent_error_log` row
`UNROUTED_IMAGE_KIND`/image-generator/**2026-07-24 20:45:57Z**, context
`scratch_unrouted_011`. **That row no longer exists.** `[MEASURED 2026-09-04]`
`agent_error_log`'s oldest surviving row is **2026-07-24 23:30:20Z** — two and a half
hours *after* the cited one — and retention is **resolved > 14 days, unresolved > 30
days** (`database-cleanup` arm 1, migration `465`). At 42 days old it was reaped.

Nothing about 011's closure is wrong; the *evidence* is simply no longer re-runnable.
Anyone re-checking it from the bug file, from this log, or from the auto-memory will
query for that row, find nothing, and have to decide whether that means the fix was
never proven. Landmine appended (`LANDMINES.md`), footprint `agent_error_log`.

### Missteps this session, all caught before anything was published

1. **A false zero from a wrong JSON path, twice.** `s.value->'input_mapping'` (should
   be `->'config'->'input_mapping'`) returned four blank columns, and
   `data->'imagery_style_guide'` under `design_intent` returned "0 sites have a guide"
   when the true answer is 14. Both would have read as findings. What caught the second
   was insisting on a control that *had* to be non-zero before believing the zero
   beside it — the memory's own rule, and it earned its place again.
2. **A regex census that silently dropped half its population.** `origin_model ~*
   'banana|gemini'` matched 315 of 606 rows in one week and I nearly reported the gap
   as unexplained. Enumerating the distinct values first showed the remainder was
   `derived-from-hero`/`derived-from-logo` — correctly excluded, but by luck, not by
   design. **Enumerate the values before writing the bucket predicate.**
3. **I read `error_type` and `created_at` off `agent_error_log` from memory**; the
   columns are `error_code` and `occurred_at`. Schema first, as CLAUDE.md says.

> **Provenance footnote, same session (2026-09-04).** The `agent_error_log`-retention
> landmine described in §4 above is **committed under another lane's message**:
> `d2e1763d8` ("the finetuning lane caught a defect in MY recommendation…", the
> `infographics` lane, 14:34). Between my append and my commit, three other sessions
> committed `LANDMINES.md`, and mine rode along as a same-file passenger — the exact
> class CLAUDE.md documents, running in the direction that catches *you* rather than
> the one you guard against. Nothing was lost and forward-only holds, so this is a
> note, not a repair. Recorded because `git log` on that entry names a lane that did
> not write it, and the ledger's own convention is that the `added:` line is the
> authority on authorship, not the commit. Entry verified present at HEAD; doc_notes
> synced; verifier corr `8c5d0f5f`.

> **CORRECTED 2026-09-04, same session, before anyone acted on it — my own demand-control
> figure was wrong in §1 and §2 above, and in the commit message `9029482ec` which
> forward-only forbids me amending.**
>
> I published **"1,046 generated assets across 37 sites"**. The exact figure is
> **1,025 generated assets across 39 sites** `[MEASURED 2026-09-04]`:
> ```sql
> SELECT count(*), count(DISTINCT site_id)
>   FROM assets WHERE origin_type='generated'
>    AND created_at > (SELECT max(created_at) FROM assets WHERE origin_model ~* 'stab|sdxl');
> ```
> **Two errors, one cause — I built the figure by summing the DISPLAY table instead of
> re-deriving it with the claim's own predicate.** (1) I added whole weekly buckets,
> including the `08-10` bucket, which *contains* the last SDXL asset — so ~110 assets
> generated **before** the stop were counted as being after it. (2) The site count `37`
> came from a different query bounded at the **08-24 roll**, not the **08-11 stop**; I
> carried it across without noticing the windows differed.
>
> **The conclusion is unchanged and if anything strengthened** — 1,025 generations on 39
> sites with zero SDXL is the same overwhelming demand control, over a *correctly* dated
> window. But the number was wrong in five places, and it was wrong in the direction that
> flatters the argument, which is the direction to distrust in your own writing.
>
> **What caught it:** re-deriving the figure to answer a *different* question — whether my
> population matched the one `bugs_closed/382` §10d's own standing check uses
> (`origin_type='generated'`; it does, 1,290 of 1,293 rows since 07-18). I would not have
> caught it by re-reading, because the arithmetic on the table I had written was correct.
> **The cheap check, and it is the one I skipped: a headline figure gets re-derived by a
> query carrying the claim's own predicate — never summed off a table built to display
> something else.** A bucket boundary is not a rounding error; it silently reassigns
> everything on the wrong side of the event you are dating.
> Logged in `WRONG_CALLS.md`.

> **CORRECTED 2026-09-04, second correction this session — §4 above states the retention
> rule from the MIGRATION FILES, and the live rule is different in kind. I wrote an entry
> about stale evidence and sourced it from a stale artefact.**
>
> §4 says *"resolved > 14 days, unresolved > 30 days (`sql_for_agents/465`)"*. That is the
> **superseded** text. Migration **`567`** replaced the arm. `[MEASURED 2026-09-04]`, read
> from the live row — which is where it lives, and I never opened it:
> ```sql
> SELECT substr(pre_query, position('DELETE FROM agent_error_log' in pre_query), 1400)
>   FROM scheduled_tasks WHERE name='database-cleanup';
> ```
> **The live rule:** a row is deleted at **30 days ONLY IF `split_part(error_code,':',1)`
> is on a 16-code allow-list**; everything else survives to **365 days**. **`resolved`
> plays no part** — the sweep's own comment reads *"`resolved` does NOT shorten a row any
> more (it used to halve it to 14 days, which was backwards)"*.
>
> **This strengthens the finding in two places and I want both on the record:**
> 1. **`UNROUTED_IMAGE_KIND` IS on the allow-list.** So 011's proof row was not merely
>    caught by a general sweep — it was reaped **by name** at 30 days. §4's conclusion is
>    right and now has its exact mechanism.
> 2. **`MISSING_IMAGE_KIND` is NOT on the list**, so such a row would live **365 days**.
>    **Therefore §2's caveat was too harsh on itself**: the absence of `MISSING_IMAGE_KIND`
>    over the 11 days since the roll *is* meaningful, because a row would still be there.
>    The SDXL census and the condition-code absence are now **two independent instruments
>    agreeing**, not one instrument and one shrug. The "no positive control" point still
>    stands narrowly — the only `reported_conditions` row ever written carried the one code
>    that *is* on the reap list — but it no longer undercuts the absence.
>
> **What caught it:** the landmine-verifier returned `NEEDS_HUMAN_REVIEW` on my entry over
> a 465-vs-466 numbering discrepancy. The discrepancy was real and *understated* the
> problem — both files are superseded. I had two migration files agreeing with each other
> and treated that as corroboration; two stale sources agree exactly as loudly as two
> live ones.
>
> **The cheap check: retention, schedules and sweeps live in `scheduled_tasks`, not in the
> migration that last edited them.** This is [[seed-sql-is-history-live-row-is-fact]],
> which is in my own loaded memory index, committed while writing about stale evidence.
> And note what it now costs a reader: **the retention window is PER-CODE**, so "is my
> evidence still there?" cannot be answered from the row's age — you must check whether
> its code is on the list.
