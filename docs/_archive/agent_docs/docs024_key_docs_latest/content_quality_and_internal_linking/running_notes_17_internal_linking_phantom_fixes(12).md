# Running notes 17 — internal linking phantom fixes (2026-06-10)

Site: gamesdesign.co.uk. Work: eliminate phantom/broken internal links at source, add a detection net, leave intent-aware resolution for a dedicated agent. Companion: `FOCUS_internal_linking.md`, `FOCUS_content_quality.md`. Deploy runbook: `RUNBOOK_linking_phantom_fixes.md`. Remaining: `PLAN_b4_b5_hubs_and_link_resolver.md`.

## Policy settled
- Phantom/missing internal link = **loud but non-blocking**. Not a deploy stopper; the improvement loop resolves it. Deploy gate flags phantoms as warnings, not errors.
- Prevention is layered: stop producing them at each write path (Layer 1), detect at the deploy gate + post-deploy audit (Layer 2/3), restore intent-correct destinations via a dedicated agent (Step 3).

## Decisive findings (verified against code/data, not assumed)
- `sourceResolver.resolve` (`plan_sections_action.go`) `case "pages"` **fabricated** `/<path>.html` and returned `found=true` for any non-existent page → the phantom generator behind hero/CTA. `on_missing` never fired; schema fallbacks were dead code.
- `hero` + `call-to-action` are the **only** components using a `pages.*` source (`pages.contact`, `pages.services`). Tight blast radius.
- Header/footer phantoms (`/contact.html`, `/privacy.html`, `/terms.html`) were **hardcoded `ContentData`** in `render_site_components_action.go` (`cta_url`, `legal_links`) — NOT the `multipage_actions` 310–318 fallback nav (nav is already real-page-derived via `GetNavItems`), and NOT in the templates. So `nav-link-fixer` (which only rewrites `#slug` anchors in templates) can't reach them.
- `link_registry` has a `target_page_id` column + FK but `syncLinksToDB` never populates it; `extract_and_sync_links` is wired into no live workflow → `link_registry` empty. Audit reads `rendered_html`.
- `validate_page_content` already had `validateInternalLinks`, but emitted missing targets as one non-blocking warning (phantom + planned lumped) and never inspected `site_components`. Its `normalizePagePath` (lowercase + append `.html`) disagreed with the audit's normalisation.
- `component-template-fixer` exists but **punts on CTAs** (`cta_improvement`/`cta` → `needs_review`). `identity-advisor` and `sites.approval_mode` do **not** exist.
- B4/B5: the `*_index_url` specs are **absent** (the `identity` spec has tone/contact/services but no `*_index_url` keys); `game-list` even has a real fallback `/games/index.html` that still didn't apply (spec-path resolver / template gating to verify). Real hubs exist: `tools-index`, `guides-index`, `games-index`.
- Operational: `improvement-sweep` scheduled_task is **disabled** (`enabled=f`, last completed 2026-05-08), intentionally paused during core build.

## Shipped/written this session (files in outputs)
- `plan_sections_action.go` — `resolve` `pages` case: real URL or `(nil,false)`, no fabrication. (Patch.)
- `step1_hero_cta_phantom_fix.sql` — `hero`/`call-to-action`: `on_missing: skip_field`, fallbacks removed, templates gated on url.
- `datahelpers/links.go` — canonical `ExtractHrefs`/`ClassifyLinkScope`/`IsAssetPath`/`NormalizePagePath`/`PageURLSet`.
- `validate_page_content.go` — gate now uses datahelpers; `phantom_link` + `empty_internal_href`, non-blocking.
- `check_phantom_internal_links.go` — post-deploy audit on datahelpers; routes by surface (site_component→`nav-link-fixer`, page_component→`internal-link-resolver`). Inert until enabled.
- `render_site_components_action.go` + `layer1b_header_footer_phantom_fix.sql` — header/footer phantoms fixed at source; `legal_links` from `GetNavItems(NavGroupLegal)`, `cta_url` from real contact page, header CTA gated, footer legal data-driven.

## B4/B5 — done (ready for batch deploy)
- `section_index_for.go` — new `queryresolve` verb (new file in the package; only existing-code edit is one switch case). Shared-area lookup then URL-prefix fallback; returns the hub URL or `nil` (never `""`).
- `b4_b5_hub_links_schema.sql` — repoints `tool-list`/`game-list_pre_037`/`guide-list_pre_037` `cta_url` to `query.section_index_for:<type>`; drops `game-list`'s dead `/games/index.html` fallback. (Confirmed: that fallback never fired — the field had no `on_missing`, so it defaulted to `skip_field`, which ignores `fallback`; and the list templates render the Browse-All anchor ungated, hence `href=""`.)
- `b4_b5_hub_links_template_gate.sql` — gates the three Browse-All anchors on `{{if .cta_url}}` (correct-or-absent for hub-less sites).

## Step 3 — resolver core done; agent + wiring next
- Decision (confirmed): build-time resolution via `internal-link-resolver` spawned as a sub-agent of `page-content-writer` (like `research-agent`). No persistence; a post-deploy phantom → page rebuild re-runs resolution. `merge_with` is single-source, so the resolver augments each hero/`call-to-action` section's `resolved_data` in Go; the render loop is unchanged.
- `resolve_internal_links_action.go` — section-augmenting, component-aware (`hero`: `cta_url`/`secondary_cta_url`; `call-to-action`: `primary_cta_url`/`secondary_cta_url`), validates every target via `datahelpers.PageURLSet`, returns augmented `sections_ready` + `unresolved`. v1 rule: top content hubs by `nav_order`, excluding about/contact/legal and the page's own hub.

### Guideline audit (001/003) — fixes applied to the resolver action
- Was reading the `sections` array with the literal key via `ExtractNestedField` → would silently run on empty. Fixed: resolve the config PATH (`params.StepConfig.Config["sections"]`), then `ExtractNestedField`.
- `sections` was in `ActionInputSpec` → `current_page.sections` collision risk. Fixed: removed from the spec; read from the config path. Spec keeps scalars only (`site_id`, `page_type`, `page_name`).
- Own-hub exclusion was keyed on the section name (`"hero"`), never matched. Fixed: added `page_name` and excluded `hub.Name == page_name`.
- Logged the pattern in `016_debugging_guide_v2_45` (§9 + §0 #15).

### Agent-modeling facts (from `research-agent` row + 003)
- `agent_definitions` has NO `processing_mode` column; the workflow lives in `default_config.workflow`, with `processing_mode`/`timeout_seconds` inside `default_config`. `task_workflow`/`orchestrator_workflow` are null for called sub-agents.
- 003 requires `agent_category` (use `specialist`), `input_contract`, `output_contract`, `image_repository`/`image_tag`. Topics templated (`system.agent.{type}.process`, etc.).
- Rebuild trigger (003 `content_direction` + arch table): set `pages.build_status = 'needs_rebuild'`; the `page-rebuild` specialist picks it up. So the check's `page_component` finding should set `needs_rebuild`, not route to the resolver directly.

## Pre-deploy review (2026-06-11)
- Found the Step 1 `plan_sections_action.go` fix was MISSING from outputs while the current uploaded file still had the fabrication (line 334-335). Recreated as a full drop-in copy of the current upload + the one-block fix (returns `(nil,false)` + Info log instead of fabricating `/<path>.html`). Verified: one-block diff, balanced, `zap`/`r.logger`/`r.siteID` already present in the file.
- Verified `section_index_for.go` against the live `queryresolve` package: signature/parse/`req.SiteID uuid.UUID`/`*sql.DB` match the siblings, and the `status IN ('active','deployed')` filter is exactly `resolvePagesWhereType`'s proven filter.
- Verified the resolver action's seams: `datahelpers.ExtractNestedField` exists and is exported; `stringOrEmpty` is package `actions` (defined in `plan_sections_action.go`); no name collisions for the new helpers; the validation `PageURLSet` filter is identical to the deploy gate's (`NOT IN ('deleted','archived')`), and hub candidates (`IN ('active','deployed')`) are a subset — the resolver cannot emit a URL the gate flags.
- Verified my `render_site_components_action.go` / `validate_page_content.go` outputs are the current uploads + localized intended edits only.
- Noted intermediate state: layer1b SQL is applied but its Go isn't deployed, so header/footer phantoms PERSIST until the batch image rolls and site components re-render (old Go still injects `cta_url=/contact.html` and the hardcoded legal links into the now-gated/data-driven templates — values present, so gates pass). Expected; clears at RUNBOOK step 3.
- `check_phantom_internal_links.go` still routes `page_component` → `internal-link-resolver` (line 138); the agreed change to a rebuild route is pending `page-rebuild`'s input contract. Inert until the check is enabled, so not a batch blocker.

## Step 3 — completed (2026-06-11), all deliverables written
Grounded in the real rows (`page-rebuild`, `page-content-writer`) + existing patterns:
- `page-rebuild`'s contract ("pages must be PRE-flagged needs_rebuild"; batch maintenance agent) ruled it out as a per-finding handler. The per-item rebuild path the improvement loop already uses is **`page-build-handler`** (`check_empty_sections` routes there, pipeline `content`) — so `routeBySurface` now: `page_component` → `page-build-handler`/`content`/35, `site_component` → `nav-link-fixer`/`build`/40 (mirrors `check_broken_nav_links`). Check header + fix text updated.
- `pages.status`: only `'active'` exists (22 rows). The `IN ('active','deployed')` filters match the proven sibling verb; vestigial second term kept for consistency.
- **Agent row** (`internal_link_resolver_agent.sql`): modelled on `research-agent` — workflow in `default_config.workflow` (`resolve_links → complete`, thin; logic in the Go action), `processing_mode: task` inside config, `agent_category: specialist`, contracts, templated topics, NOT EXISTS-guarded. `image_tag` to be set to the batch tag.
- **Writer wiring** (`page_content_writer_link_resolver_wiring.sql`): `snapshot_agent()` first (v2_40), then chained `jsonb_set`: + `spawn_link_resolver` (after `spawn_research_agent`), + `resolve_links` (call_agent after `build_render_context`, `error_step` falls through), + `select_sections` (`extract_fields` FALLBACK CHAIN — verified `ExtractFieldsAction` copies raw values, arrays included, and supports map-of-arrays), loop `iterate_over` → `sections_for_render.sections_ready`. Regression-safe: resolver failure ⇒ byte-identical to today (incl. the rebuild flow where `section_plan` is unmapped — pre-existing v2_33 territory, unchanged). `input_mapping` uses `?` optional suffixes for all but `site_id`.
- **`unresolved_cta`**: emitted in-Go from `ResolveInternalLinksAction`, mirroring `createDeferredItems` (same package; reuses `sanitiseSectionKey`): one item per affected section, `item_type unresolved_cta`, `status needs_human_review` (a rebuild can't fix "no hub exists" — HITL, not a handler), `ON CONFLICT DO NOTHING` dedup, insert failure logged never returned.
- All Go balance-checked; all embedded SQL JSON validated.

## Deployment (2026-06-11)
- Applied pre-image (verified): `internal_link_resolver_agent.sql` (one agent row; the 2-row verify output was `jsonb_object_keys` fanning out the two steps, not a duplicate) and `page_content_writer_link_resolver_wiring.sql` (all 7 verification columns correct).
- Gap property noted: wiring live before the new image is SAFE — the resolver call errors on the old image, `error_step` → `select_sections` → fallback chain → behaviour identical to today. Resolver-call errors in logs during the gap are expected noise.
- Baseline pre-roll: `internal-link-resolver` + `page-content-writer` both `image_tag v1.0.1060`; `unresolved_cta` count 0. Post-roll check: tags must bump TOGETHER (a lagging resolver tag = old image pods = permanent silent fallback).
- Batch deploying now; verification + readopt sequence in RUNBOOK Part 3.

## Readopt decision
Readopt gamedesign.uk → gamesdesign.co.uk is the right next step, SECOND: first the targeted Part 3.1 checks on the existing site (six code changes + four SQLs shipped together; the per-layer checks make any failure attributable), then the readopt as the from-scratch acceptance test + the fresh baseline for the content-quality package. Evidence the virgin path resolves hubs: planning creates all pages rows before writing (writer receives `db_sync.pages`), and the original build's list cards resolved through the same status filter the hub queries use. Expected on readopt: the open content-quality defects recur (adopt-path, untouched) — input to the next package, not linking regressions.


1. Deploy the batch per `RUNBOOK_linking_phantom_fixes.md` (Parts 1+2), re-render, run the dry-run.
2. Later, deliberately: enable `phantom_internal_links` in the discovery checks array, watch one sweep, then re-enable `improvement-sweep`.
3. Then back to the content-quality items (FOCUS_content_quality).

## Re-render mechanics learned (2026-06-12, from 081b/081d/081c)
- Three real trigger mechanisms: `page-rerender` (single page, kcat orchestrate to `system.agent.generic.requests`), `rerender-pages` (whole site, `refresh_site_components: true`), and work item → `build-dispatch-loop` → `page-build-handler` (the production path to the dynamic handler; a direct orchestrate does NOT reach it). `build-dispatch-loop` claims `status IN ('triaged','approved')` — with `improvement-sweep` (the triager) disabled, manual items must be inserted as `'triaged'`.
- DECISIVE distinction: **re-render re-applies templates to component data stored at last build; only a rebuild re-runs `plan_sections` source resolution + the writer (and now `resolve_links`).** Consequence: hero pages' stored data still carries the fabricated `/contact.html`, so a re-render passes the new gate and the phantom SURVIVES re-render — heroes and the Browse-All hub URLs need the rebuild path. Header/footer are the exception (their data is built fresh in Go at render time — layer1b).
- `page-rebuild` agent stays OFF the menu for this (its writer call lacks the `section_plan` mapping — v2_33 gap; sectionless risk).
- RUNBOOK rewritten pragmatically around this: orient (`check_linking_sql_applied.sql`) → missing SQLs → site-wide rerender (081d pattern) → dry-run measure → work-item rebuilds for hero/CTA/list pages (first item = resolver end-to-end check) → done-criteria. Open before the §5 INSERTs: `\d site_work_items` + page-build-handler's workflow (which `item_type` it expects for an existing-page rebuild; `empty_section` is the proven analog).

## page-build-handler contract (2026-06-12, from its definition + \d site_work_items)
- Handler does NOT branch on item_type; reads `spec.page_id`/`spec.page_name` (page_name also feeds save_sections/update_status — MANDATORY in spec), `spec.mode` ('recreate' loads the adoption crawl, preserving copy), `spec.suggestion` → writer `rewrite_guidance`. Its `call_content_writer` maps `section_plan` — so the work-item path is the one that feeds the resolver (`sections?` resolves). It deploys per page (spawns page-rerender + git commit).
- Dedup is a partial unique index `(site_id, item_key)` over non-terminal statuses — `ON CONFLICT DO NOTHING` is correct, and a completed item's key is reusable for a future rebuild.
- RUNBOOK §5 finalized: item_type `link_resolution_rebuild`, pipeline `build` (the dispatch-PROVEN combination via the gap planner), status `triaged`, handler_agent `page-build-handler`, spec {page_id, page_name, mode: recreate, suggestion, reason}; index-first then the worklist (hero/call-to-action/tool-list/game-list/guide-list pages; join `pc.component_id = cc.id` confirmed from check_empty_sections).
- WATCH ITEMS (new, non-blocking):
  1. `update_page_status` in this handler sets `status: "deployed"` — yet all 22 pages read status='active'. Either it writes build_status or this path never ran here. If pages flip to 'deployed' after §5: queryresolve filters include it (fine); `ensurePages` in plan_sections filters status='active' ONLY — pages.* lookups would then miss flipped pages (only hero/cta contact/services use pages.*, both non-existent, so currently harmless — but a latent trap).
  2. This path maps NO `db_sync` to the writer → `prepare_link_context`/`available_pages` get nothing → the LLM's internal-linking constraint text is empty here. Pre-existing, resolver-independent (it queries the DB directly). Candidate fix later: map db_sync or make prepare_link_context load pages itself.
  3. My audit check emits page_component items with pipeline='content' (mirroring check_empty_sections); the dispatch-PROVEN pipeline for page-build-handler is 'build'. Before ENABLING the check: confirm build-dispatch-loop claims 'content' items too, or align the check to 'build'.

## §3 re-render done + §4 dry-run baseline (2026-06-13)
- site_id this cycle: `e33263f4-74f8-494f-b191-546845dbbddf` (re-resolve each cycle; teardown changes it).
- All four template/schema SQLs verified applied (step1 was the last in; six `UPDATE 1`s, no silent no-op; `step1_snapshot=t`; both CTA components `skip_field`/`fallbacks_gone=t`/`has_and_gate=t`, literals gone). CTA schemas read `pages.contact`/`pages.services` + `on_missing:skip_field` — correct ONLY because the `plan_sections` Go fix is live (those lookups now return nil → field omitted → gate renders no button).
- §3 site-wide re-render via `rerender-pages` + `refresh_site_components:true`: orchestration `a2644e72` COMPLETED (~16s). Two operator slips, both harmless: the runbook's `psql`-on-desktop SITE_ID line doesn't work from the desktop (psql is cluster-side); and a space-prefixed `SITE_ID=" e33…"` run (`639bfc6d`) produced ZERO orchestration_states rows — never created a workflow, wrote nothing, nothing to undo (malformed site_id rejected at ingestion). Real site_id confirmed `e33263f4…`.
- §4 dry-run audit (manual, mirrors the phantom_internal_links check's normalisation) — 28 findings, ALL `page_component`, splitting exactly as predicted:
  - **0 `site_component` rows** — header/footer phantoms (/contact.html, /privacy.html, /terms.html) GONE. Layer1b Go re-render confirmed working (the layer a re-render CAN fix).
  - Hero `/contact.html` + `/services.html` pair on all 11 content pages (stored resolved_data from last build) + 4 `empty_internal_href` on list slots (game-list/guide-list/tool-list). These need SOURCE re-resolution → §5's job; presence is correct, not a regression.
  - Neither surprise fired (site-component phantoms didn't persist; hero phantoms didn't vanish) ⇒ §5 runs in FULL, not the collapsed single-page version.
- §5 worklist (the distinct `loc` values needing rebuild): index, tools-index, games-index, guides-index, about-index, contact-index, guide-economy-basics, guide-fairness-in-rng, guide-p2p-architecture, guide-rng-design, guide-skinner-box (11 pages). about-index/contact-index heroes get real hub CTAs too (own-hub exclusion is by area — about/contact already excluded as candidates regardless).
- RUNBOOK fix owed (deferred, not yet applied): replace the `psql`-on-desktop SITE_ID line with a manual-id form + guard against leading/trailing space in SITE_ID.

## §5 launched (2026-06-13) — 21 items queued
- Ran index-only INSERT then the all-pages INSERT → `INSERT 0 21` (not 11). EXPECTED-larger: the all-pages filter matches any page with a hero/CTA/list component; §4's 11 were only pages whose RENDERED output carried a phantom. All 21 are valid rebuilds (every matched page has CTA components → gets real hub destinations); cost is more recreate-mode content churn + 21-page verification surface. Chose to let all 21 run (all correct). Scope-down option recorded in runbook if churn matters.
- Watch SQL added to runbook §5 (A progress / B unresolved_cta=0 / C resolver augment result / D writer iterated-augmented-not-fallback / E §4 dry-run = ground truth). CAVEAT: C/D assume resolver output at `collected_data.link_resolution` and writer's at `collected_data.sections_for_render` — UNVERIFIED nesting; if null but pages clean, adjust path, trust E. Resolver is task-mode sub-agent; if C empty, confirm via pod log `resolve_internal_links: augmented CTA sections`.

## §5 progress + terminology/query fixes (2026-06-14)
- §5 healthy mid-run: 3 complete, 1 claimed, 18 triaged; `unresolved_cta=0`. Query D returned `for_render=2` ⇒ select_sections ran and the loop consumed the resolver's augmented sections (a fallback-to-empty would give 0) — wiring confirmed LIVE.
- Fixed my own errors, not system issues:
  - "dry run" was a misnomer for the §4/§6 audit query (it's a read-only POST-DEPLOY audit of already-deployed HTML; §3/§5 write). Renamed throughout the runbook to "audit query".
  - Watch query C errored ("array length of a scalar") and D's plan_count was null — both my wrong collected_data paths. Corrected: `section_plan` nests under `input_data` (arrives via input_mapping); the resolver's output is NOT a top-level `link_resolution` key (it returns inside the writer's `resolved_links.response`), so confirm the resolver ran via the POD LOG line, not SQL. E (the §4 audit) is ground truth.
- §7 rewritten: enabling `phantom_internal_links` ≠ enabling the sweep. Findings land `status='detected'` which is NOT claimable (loops claim triaged/approved), so the check can run observe-only with improvement-sweep still disabled. PREREQUISITE before enabling: fix `routeBySurface` page_component pipeline 'content'→'build' (dispatch-proven pipeline for page-build-handler; §5 inserts proved 'build' is claimed). Reads-first to find the discovery agent owning the checks array (don't guess) before the jsonb_set append + snapshot.
- STILL OWED (runbook): psql-on-desktop SITE_ID line + leading/trailing-space guard. And the routeBySurface pipeline fix is now a real code edit to make (uploaded check still has 'content').

## Wiring confirmed + §7 home/pipeline corrected (2026-06-14)
- Query D (corrected paths) on both completed rebuilds: `for_render=2`, `plan_count=2`, EQUAL ⇒ resolver augmented sections + writer loop consumed them (not fallback). Wiring confirmed end-to-end.
- Discovery-agent enumeration (12 agents). Discovery checks live in domain-scoped agents via `run_discovery_checks` `config.checks` + `check_domain`:
  - quality-discovery-agent: [broken_nav_links, placeholder_contact, generic_theme], domain=build
  - design-discovery-agent: design checks, domain=design
  - completeness-discovery-agent: [..., empty_sections, unlinked_page_components, unresolved_sections, ...], domain=content
- §7 CORRECTIONS:
  - Home for phantom_internal_links = **completeness-discovery-agent** (content-integrity family; check_domain=content), NOT quality.
  - REVERSED my earlier watch-item #3: `pipeline="content"` for page_component findings is CORRECT (completeness agent is content-domain). Do NOT change content→build. The §5 manual inserts used 'build' only because they hit page-build-handler directly.
- NEW open question (the real one): `run_discovery_checks` passes `check_domain` to all checks, but my check sets its OWN per-surface WorkItemSpec.Pipeline in routeBySurface. Must confirm (code read of run_discovery_checks + a sibling check's WorkItemSpec) whether per-finding Pipeline/handler override SURVIVES or gets STAMPED by check_domain. If stamped: site_component→nav-link-fixer/build route is lost (all become content) → move per-surface routing out of the check. Gate before enabling.
- improvement-loop confirmed as the sweep orchestrator (notify_scheduler updates scheduled_tasks 'improvement-sweep'); it spawns the discovery agents + build-dispatch-loop + triage_detected_items. Enabling a check WITHOUT this loop running = findings stay 'detected' (unclaimable). Safe observe-only.

## §7 gate RESOLVED — enable as-is (2026-06-14, run_discovery_checks_action.go)
- Per-finding routing SURVIVES: insertWorkItem maps `pipeline: wi.Pipeline` + `handlerAgent: wi.HandlerAgent` straight from each returned WorkItemSpec. Config `check_pipeline` only sets dctx.Pipeline (a shared default the check never reads). So routeBySurface values write verbatim (site_component→nav-link-fixer/build, page_component→page-build-handler/content). check_phantom_internal_links needs NO code change.
- Config key is `check_pipeline` (not check_domain as I'd been calling it); only sets the unused dctx.Pipeline default — moot for my check.
- Pattern confirmed correct: checks RETURN WorkItemSpecs; the action inserts them in its own tx. My check reading via dctx.DB (not dctx.TX) is right for auditing live rows.
- Net: zero code changes outstanding for the linking work. §7 enable is purely the snapshot + jsonb_set append to completeness-discovery-agent's run_checks.config.checks, then a one-shot kcat run, observe-only (findings stay 'detected', unclaimable, sweep off).
- DONE (runbook hygiene, 2026-06-14): added a Conventions note (shell vs psql var syntax; the literal site_id; the leading/trailing-space = silent-no-op warning); replaced both `psql -Atc` desktop lines (psql isn't on the desktop) with a literal SITE_ID; fixed the watch block's shell-syntax `SITE='...'`/`:'SITE'` to inline `(SELECT id FROM sites …)`. No runnable block now mixes shell and psql syntax.

## §5 stragglers + observed defects (2026-06-14)
RESULT: 21 complete, 1 failed (22 items = 1 index + 21 rest dedup-skipping index). BUT verification found TWO pages actually unrebuilt, both from the same transient.

### Two stragglers — same transient cause (claim timeout / handler pod died)
- `games-index`: status=failed, attempt_count=3, error "Claim timed out (attempts exhausted)".
- `index`: TWO rows both status=`complete` but errors "Claim timed out — handler pod likely died" (attempt 2) and a clean attempt-0 row (06-13). Despite `complete`, index's STORED render still carries hero `/contact.html`+`/services.html` AND all three list Browse-All buttons still `href=""` → index was NEVER successfully rebuilt. This is the silent-`complete` failure mode (terminal status ≠ work done) — trust the rendered HTML, not the status.
- Cause is NOT page-specific AND NOT load: same "claim timed out / pod died" hit index (x1), games-index (x3). [CORRECTION 2026-06-14, per user] Plenty of capacity, minimal load — so NOT dispatcher/pod-resource starvation. Likely cause: pods killed mid-flight by a CONCURRENT production deploy elsewhere (chassis rollout / rollout restart cycles agent pods → kills in-flight page-build-handler jobs → presents as "claim timed out / pod died"). OPERATIONAL RULE: don't roll the chassis image while a rebuild batch is draining. FIX: re-queued both as `_retry2` (terminal-status rows excluded from dedup index, fresh keys insert). WATCH: retries should land cleanly now (no concurrent deploy); if they land it confirms the deploy-collision theory; if they die again with no concurrent deploy THEN look at claim-timeout/capacity.
- unresolved_cta still 0 throughout (failures were infra, not "no hub to resolve to").

### Linking fixes CONFIRMED in deployed HTML (the 20 that landed)
- Game/guide/tool heroes: headline+subheadline, NO buttons (Step 1 gate `{{if and .cta_text .cta_url}}` correct when hero has no cta_text). Phantoms gone.
- games hub page-body `game-list`: "Browse All Games" → `/games/index.html` (real hub, via query.section_index_for); cards → real `/games/<slug>/index.html`. B4/B5 confirmed.
- Header/footer clean across rebuilt pages.

### NEW defect (NOT linking) — pathfinding game page has no game
- `games/pathfinding/index.html`: hero + "How A* Actually Works" prose only. NO interactive surface — no <canvas>, no game <script>. The games-hub card promises the interactive sim; the page doesn't contain it. Appears isolated to this one game (others not yet checked). Adopt-path/build-integrity issue (site-adoption-agent is meant to preserve interactive surfaces even when prose is minimal). OUT OF SCOPE for linking; log + investigate separately. TODO: check whether other games retain their interactive surface (pathfinding-only vs pattern); determine if crawl missed the JS or the rebuild dropped it.

### Content-quality observations (adopt-path, parked, EXPECTED to recur on readopt)
- Title carries brand suffix "- GameDesign.uk" / "| GameDesign.uk" (note: also the OLD-domain casing GameDesign.uk, not gamesdesign.co.uk — brand-name source is stale/wrong-domain).
- Card titles carry "- GameDesign.uk" suffix (e.g. "Auto-Battler Prototype - GameDesign.uk").
- Footer brand-tagline empty (`<p></p>`), footer contact empty (`<a href="mailto:"></a>`, empty phone) — identity.contact email/phone/address null.
- games hub meta description empty.
- LEAD for the contact/identity fix: finetune.uk has WORKING contact details — inspect its site_specs identity/contact + content_data to see the shape that populates correctly, and compare to gamesdesign's nulls. (Likely the next package's first task.)

## Missing-game (pathfinding) + speed TODO + doc updates (2026-06-14)
- PLAN written: PLAN_pathfinding_missing_game.md. CONFIRMED same problem class as the May–June "adopted interactive page has no widget" investigation (PLAN_tool_widget_clobber.md / HANDOFF_2026-05-26_tool_routing_fix_deployed.md / 016 addendum). Decisive NEW fact: the OTHER four games DO still have their interactive surface (user-confirmed) → NOT the systemic all-games recreation-loss; isolated to pathfinding. Leading candidate = clobber (pathfinding was a §5 rebuild target; save_page_sections delete-and-reinsert drops widget rows not render_action-preserved). Plan is diagnosis-first: scoped queries (pathfinding vs working sibling) Q1 page-dump / Q2 twin / Q3 page_component_history clobber evidence / Q4 snippets false-negative; also the v2_49 A1 `<div>`-not-`<section>` parser-miss candidate. Fix decided by outcome; preferred durable fix = preserve component_level IN ('tool','game') in findPreservedComponentIDs so NO content/linking rebuild can ever drop a widget (also protects the working games + the readopt). DO NOT bulk-re-trigger before pinning the candidate.
- DEBUG GUIDE updated → 016_debugging_guide_v2_49.md (from uploaded v2_48): added §9 "two rebuild routes" + §9 "complete-but-stale / pod died mid-flight (concurrent deploy)"; changelog corrects the "v2_33 gap" mis-citation; cross-linked the A1 tool-recreation entry to the pathfinding plan. NOTE: v2_48 is a separate lineage from my earlier standalone v2_45 — the "two rebuild routes" entry was re-added here so it lives in the canonical guide.
- TODO (user, soon): SPEED UP the rebuild pipeline. Takes MANY HOURS; should be 1–2h even allowing for the slow dispatcher tick + container boot. Not yet investigated. Candidate angles to explore: dispatcher tick interval; per-item container cold-boot vs a warm/long-lived handler; sequential vs parallel claim; whether 21 items serialise. (Separate workstream from linking close-out.)
