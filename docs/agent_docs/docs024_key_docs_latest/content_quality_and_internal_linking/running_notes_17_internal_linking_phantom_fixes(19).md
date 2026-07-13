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

## index reproducibly stale — concurrent-deploy theory FALSIFIED for index (2026-06-15)
- games-index retry SUCCEEDED: rendered games/index.html shows all 5 cards, Browse All Games → /games/index.html, clean hero. That straggler DONE.
- index `_retry2` = `complete`, attempt_count=1, SAME error "Claim timed out — handler pod likely died", and user CONFIRMS no production deploy at 12:41 06-15. So index has now failed identically 3×: 06-13 (attempt-0, clean-but-stale), 06-14, 06-15 — across quiet windows, no deploy.
- CORRECTION: the concurrent-deploy explanation does NOT hold for index. Reproducible same-page failure with no deploy ⇒ index-specific. Leading hypothesis: index's build DURATION exceeds the page-build-handler claim lease/heartbeat — lease expires, dispatcher treats claim as dead, terminal-exit still stamps `complete` (silent-completion). Index is the ONLY page with all three lists (tool-list+game-list+guide-list) + hero ⇒ most plan_sections resolution + most query.section_index_for calls + longest writer run ⇒ the one page that exceeds the lease while smaller pages finish inside it. (Same root also feeds the slow-whole-batch concern.)
- DO NOT re-queue index a 4th time blind. NEXT: read index's retry orchestration — current_step + seconds_elapsed (stuck partway near the timeout value = duration-vs-lease confirmed); and whether the writer/resolver child even spawned. Queries handed over. Deceideafter reading: raise page-build-handler claim lease/timeout (idle_timeout_seconds=180 on the row; check the claim-lease separately) OR cut per-section cost.
- Debug guide v2_49 corrected: the complete-but-stale entry now splits sub-case (a) concurrent-deploy vs (b) reproducible same-page duration-exceeds-lease; changelog updated.
- Ties to SPEED-UP TODO: if index is slow because per-section resolution is serial, that's both why it times out AND part of why the batch takes hours — investigate together.

## index deep-dive — root cause NOT timeout/deploy; it's save_page_sections writing nothing (2026-06-15)

### How we got here (theories raised and FALSIFIED in order)
1. "Load — 21 rebuilds at once starved the dispatcher." FALSIFIED: user confirms ample capacity, minimal load.
2. "Concurrent production deploy cycled the pods mid-flight." FALSIFIED for index: user deployed nothing at the 12:41/13:00 retry window; `kubectl get pods` shows agent-chassis up 2d2h, 0 restarts (only unrelated CrashLoopBackOff on thunder-adapter + a CreateContainerConfigError on analyser-adapter — neither in our path).
3. "index's build DURATION exceeds the page-build-handler claim lease." FALSIFIED: index's handler orchestrations (06a75284, 34a1b79d) ran ~740s & ~766s and reached current_step=complete / status=COMPLETED — well within timeout_seconds=1800. They did NOT die partway.
4. "Caller-side call_agent timeout < callee runtime (the agent_error_log call_dispatch 'timed out after 3 retries')." Partially real as a STATUS artifact — the work item shows complete + error 'Claim timed out — handler pod likely died' because a caller stopped waiting while the child finished — BUT this is noise, not the defect: the child DID finish and DID deploy.

### What is CONFIRMED
- Index builds, renders, and DEPLOYS successfully. git-adapter log: committed `/index.html` (+ /tools/assets/game-list.js), success:true, SHA c489e52a, 13:00:40. Handler query: saved=t, status_updated=t, deployed=t, validated=t on both runs.
- YET no page_components row for index was written: ALL FIVE components (hero, tool-list, guide-list, game-list, system-stats) are frozen at created_at=updated_at=2026-06-06 16:59:53. Nothing rewritten on 06-15.
- The page row WAS touched (pages.updated_at / last_built_at / deployed_at = 06-15 13:43) — so the handler ran and stamped the page, but save_page_sections wrote zero component rows.
- Contrast proves it's index-specific: about-index hero updated 06-14 20:53, phantom cta_url GONE. Hero-rewrite works on other pages; on index every component is stale.
- resolved_links.response is entirely NULL on index's writer orch (19ecda59) — augmented/unresolved/changed all empty.

### SCHEMA CORRECTION (carry forward)
- There is NO `page_components.resolved_data` column. Hero/section data lives in `page_components.content_data` (jsonb). The page row columns include content_data, sections (jsonb array of slot names), build_status, last_built_at, deployed_at, site_area_id, page_spec, suppressed_sections, built_from_plan_version. Several earlier queries (and possibly resolver design assumptions) used `resolved_data` — WRONG. Must reconcile where the resolver actually writes vs content_data before any fix.
- index.pages.sections = ["hero","tool-list","guide-list","game-list","system-stats"] (5 — matches writer plan_sections=5). system-stats is a section index has that the simpler pages don't.

### CORRECTED root-cause picture (leading hypothesis, NOT yet fully proven)
- The committed index.html is STALE, not freshly rebuilt: its hero carries the 06-06 phantom CTAs (/contact.html, /services.html with button text) AND its list Browse-All buttons are href="" (the OLD ungated state). That's exactly the 06-06 stored components re-rendered and re-committed. So fact "git committed index.html" is NOT "new content deployed" — it's stale content re-deployed.
- Mechanism: save_page_sections writes ZERO rows for index (silent early-return), the deploy step re-renders the unchanged 06-06 components, git commits the stale file, handler reports success. Reproduces every run because the input (06-06 components) + the failing parse are identical each time.
- This is the SAME FAMILY as v2_49 §A1: save_page_sections' HTML fallback saveSectionsExtractFromHTML extracts only <section>…</section> blocks and returns early ("no sections found") writing nothing when it can't parse the expected shape. There it hit tool pages (<div> not <section>); here it may be hitting index for a different reason (5-section set / system-stats / mode:recreate+preserve-copy producing HTML that doesn't parse into the expected per-section shape, or a sections_metadata mapping gap).
- WHY index and not about-index: NOT YET PROVEN. Candidate factors unique to index: most sections (5), only page with system-stats, the recreate/preserve-copy path on a page whose adopted HTML differs structurally. Could also be that save's section-count or metadata handling fails when the writer returns the multi-section index page.

### Still OPEN (what to nail before any fix)
- Read save_page_sections_action.go AS DEPLOYED (current upload, not the dump) to see: (a) the exact early-return condition that yields zero rows, (b) whether it parses writer HTML by <section> blocks or by sections_metadata, (c) how mode:recreate + "preserve copy" interacts. The A1 entry says a single-fragment fallback fix was proposed — confirm whether that shipped and whether it covers a 5-section page.
- Confirm whether deploy_page/page-rerender reads from page_components (stored) or from the writer's in-flight output. If from page_components, the stale-commit story is fully consistent.
- Determine the actual divergence index vs about-index: dump the writer's produced HTML for index's 13:43 run (if retained) and see whether save received parseable sections. Look at why save wrote 0 rows specifically.
- Reconcile the resolver write target (content_data vs the non-existent resolved_data) — if the resolver ever wrote to resolved_data, it wrote nowhere.

### Status
- games-index: DONE (re-rendered, 5 cards, Browse All → /games/index.html, clean hero).
- index: NOT fixed; do NOT re-queue again (reproduces identically). Defect is save_page_sections writing nothing for index → stale re-deploy. Diagnosis-first; fix deferred until the OPEN items above are read.
- Theories corrected in 016_debugging_guide_v2_49 already split deploy-collision vs duration; BOTH are now superseded for index by "save writes nothing / stale re-deploy" — guide needs a follow-up correction once the save mechanism is confirmed (NOT yet written, pending the save_page_sections read).

## index — save_page_sections read CONFIRMS the mechanism (2026-06-15, save_page_sections_action.go)
- Read the deployed action. Structure: primary path (structured sections_metadata, no parsing) → HTML-parsing fallback → enrich → **content-regression guard** → snapshot-to-history → DELETE page_components → re-INSERT.
- The A1 single-fragment fallback DID ship (saveSectionsExtractFromHTML lines ~512-547 store the whole fragment as one section when no <section> blocks match). So "no <section> → zero rows" is NOT index's failure.
- **Content-regression guard (lines 223-259) is the leading mechanism.** Sums stripped-text length of existing build_status='deployed' components; if existingTextLen>200 AND new stripped text < existingTextLen/4 → `return nil, fmt.Errorf("content regression blocked…")`. That's an ERROR return.
- Resolves the earlier contradiction (saved=t + empty saved_page_id + COMPLETED): the three returns differ by page_id — success(end)=has page_id+sections_saved=N; "no sections found"=has page_id+0; **regression guard=`return nil,err`=NO page_id**. My query showed EMPTY saved_page_id ⇒ matches the guard's error return, not the other two.
- Why COMPLETED despite an error: page-build-handler's `save_sections` step has `error_step: complete_error`, and complete_error IS a complete_workflow (success exit, "Content writer skipped…"). So the guard's error is LAUNDERED into success — silent-completion, same family as v2_33's complete_error==complete_workflow. Then deploy re-renders the stale 06-06 components and git commits the stale file.
- Why index only: the guard ARMS only at existingTextLen>200 and BLOCKS only at new<existing/4. index is the text-heaviest page (hero + tool-list + guide-list + game-list + system-stats) → biggest existing text → trips when the recreate/preserve writer returns thin content. Single-list pages don't have enough existing text to arm/block, so they save normally. Consistent with about-index saving fine.
- NOT yet 100% confirmed: needs the chassis log line `CONTENT REGRESSION BLOCKED` (with existing_text_chars/new_text_chars) from index's 13:43 run to nail it outright. The file logs every branch, so the log will name the path.

### Two distinct bugs identified (the guard ITSELF is correct — it should refuse to wipe)
1. error_step: complete_error launders the guard's legitimate refusal into a `complete` work item → page silently stays stale, operator sees success. FIX direction: give save_sections a distinct error_step (mark_needs_review / fail_work_item), not complete_error.
2. deploy proceeds after a no-write save → re-renders + commits stale components (the "git committed ≠ new content" trap). FIX direction: gate deploy on sections_saved>0 (or don't report rebuild done).
- SEPARATE OPEN CONTENT question: WHY does index's writer return thin content on mode:recreate + "preserve existing copy"? The silent-completion just hides this. Needs the writer's produced HTML for index's run.

### Logging observation (re user's "add logs where failures completed silently")
- save_page_sections ALREADY logs each branch (metadata field check / Using structured metadata path / Using HTML parsing fallback / CONTENT REGRESSION BLOCKED / No sections found / Complete). The silent part is NOT missing logs in the action — it's that (a) the success-exits return `"success": true`, and (b) the HANDLER's complete_error converts the guard error to a success workflow. So the logging gap to address is at the workflow/handler layer: a regression-blocked or zero-save outcome must surface as a non-terminal status + a distinct work-item error, not `complete`. Adding a WARN is not enough; the status is what misleads.

### Debug guide updated → 016_debugging_guide_v2_49.md rev2
- Added §9 "git committed ≠ new content" (+ the component-timestamp decisive check; content_data not resolved_data).
- Added §9 "save_page_sections regression guard → complete_error masks as success" (mechanism, page_id signature, why-index, the CONTENT REGRESSION BLOCKED log grep, the two fixes).
- Changelog rev2 line added.

## FIX written: make save_page_sections failure visible (2026-06-15)
- existing_text_chars for index = 23041 (>>200) → the content-regression guard is ARMED on index. Blocks if new stripped text < ~5760 (23041/4). Half-confirmed from data; the block itself needs the live run / DB-after signature.
- _diag1 item inserted (item_key linking_rebuild_index_diag1) to reproduce + capture. Logs rolled (the 13:43 run's stdout is gone), so confirm from DB after: _diag1 lands `complete` with empty result page_id AND index components still frozen 06-06 = regression-guard signature reproduced.
- DELIVERABLE: page_build_handler_save_failure_visible.sql — routes save_sections error to a NEW `mark_save_failed` step (fail_work_item → status_override needs_human_review, save-specific message) instead of `error_step: complete_error` (which is a complete_workflow success exit = the laundering). Reuses the exact pattern mark_needs_review already uses for the validate path. Snapshot + revert_agent included.
  - Chose needs_human_review (non-terminal, holds for a human, dedup index blocks silent same-key re-run) over failed (terminal, re-runnable) — a regression block is not a transient; re-queue reproduces it.
- OPEN PREREQUISITE before applying (documented in the SQL): save_sections has error_step in TWO places — step-level `steps.save_sections.error_step` AND `steps.save_sections.config.error_step`, both = complete_error. The migration repoints the step-level one; UNCERTAIN which the engine reads for routing (step.ErrorStep vs config["error_step"]). MUST read the chassis engine's error_step resolution before relying on it; safest is to set BOTH once confirmed they don't conflict. Verification query in the SQL surfaces both.
- SCOPE of this fix = visibility only. Does NOT fix (a) deploy proceeding after a no-write save (separate: gate deploy_page on sections_saved>0), or (b) WHY the recreate/preserve writer returns thin content for index (separate content investigation, still the real open root cause that the silent-completion was hiding).

## COURSE-CORRECTION: a second mechanism — page_components LOCKING (2026-06-15)
- `\d page_components` revealed a locking subsystem I hadn't accounted for: columns locked_at, locked_by, lock_type (CHECK: permanent|timed|review), lock_expires_at; indexes idx_page_components_locked, idx_page_components_timed_lock; and a TRIGGER `trigger_auto_lock_on_deploy BEFORE UPDATE ... EXECUTE FUNCTION auto_lock_on_deploy()`.
- NEW HYPOTHESIS (at least as plausible as the regression guard): index's components are build_status='deployed' frozen at 06-06; if auto_lock_on_deploy set a PERMANENT lock on them at first deploy, save_page_sections' DELETE+reinsert could be blocked/skipped by the lock → components never update → deploy re-renders stale → success reported. Same VISIBLE symptom, DIFFERENT mechanism + different fix (lock lifecycle, not error_step routing).
- This DIRECTLY engages user's point "other pages are probably >200 chars too": if the regression guard is ARMED on every page (all >200), the guard is NOT the discriminator for why ONLY index is stuck — something page-specific is. A per-component lock could be that discriminator. So the guard may be a red herring (or a defense that's never reached because the lock stops the write first).
- HONESTY: walked back guard-only confidence. CONFIRMED: guard armed on index (23041 chars); earlier empty-page_id consistent with guard's error return. NOT confirmed: that the guard (vs the lock) is what actually fires on index. Need the lock query to discriminate.
- DECISIVE NEXT QUERY (run first, before chasing the regression log):
    SELECT pc.slot_name, pc.build_status, pc.locked_at, pc.locked_by, pc.lock_type, pc.lock_expires_at, pc.updated_at
    FROM page_components pc JOIN pages p ON p.id=pc.page_id
    WHERE p.site_id=(SELECT id FROM sites WHERE domain='gamesdesign.co.uk') AND p.name='index' ORDER BY pc.position;
  Compare to about-index (which DID update 06-14): if index components are locked (esp permanent) and about-index's are not → LOCK is the cause. If both same / index unlocked → lock not the differentiator, regression guard remains leading.
- ALSO READ: the auto_lock_on_deploy function body (what lock_type it sets, on what condition) and whether save_page_sections checks/clears locks before DELETE (the action read earlier did a plain `DELETE FROM page_components WHERE page_id=$1` with no lock handling visible — if a trigger or FK/permission blocks delete of locked rows, the DELETE could error or no-op).
- _diag1 still `claimed` at 15:23 — DB-signature recheck is premature; re-run the (qualified) component-timestamp + result-page_id queries once it reaches terminal status. Qualified query: SELECT pc.slot_name, pc.updated_at FROM page_components pc JOIN pages p ... (updated_at was ambiguous unqualified).
- The save-failure-visible migration remains correct REGARDLESS of mechanism (a blocked save should surface, whether blocked by guard or lock) — but the ROOT-CAUSE fix depends on the lock query outcome. Hold the migration's deeper companion fixes until mechanism is pinned.

## Lock hypothesis DISPROVED — regression guard back as leading; question narrows to "why thin content" (2026-06-15)
- Lock query: index's 5 components have locked_at/locked_by/lock_type/lock_expires_at ALL NULL (despite build_status=deployed). So trigger_auto_lock_on_deploy did not lock them; the save is NOT blocked by a lock. Locking eliminated as the mechanism.
- Therefore the discriminator for "why only index" is NOT the lock and NOT the guard's arming (all pages likely >200) — it must be the guard's OTHER condition: new content < existing/4. i.e. index's writer returned content thin enough to trip the ratio; the pages that updated fine returned content that cleared it.
- So the real root cause collapses to a CONTENT question: why does index's mode:recreate + "preserve existing copy" writer run produce thin output when other pages don't? Leading structural suspicion (TO TEST, not assert): the recreate/preserve path may not be loading the existing rich content_data into the writer prompt (or loads from an empty field), so it regenerates sparse content. index's stored hero content_data is rich (headline/subheadline/CTAs/bg image), so "preserve" should keep it — if the output is thin, preserve isn't preserving.
- DECISIVE READS (no logs needed, use _diag1 once terminal):
  1. _diag1 terminal signature: components still 06-06 + complete + empty result_page_id = guard fired/reproduced. (vs components flip to 06-15 + populated page_id = save SUCCEEDED, which would mean guard didn't block and we rethink again.)
  2. Writer output length for _diag1's writer child: length(collected_data->'page_content'->'response'->>'page_html'); small = thin-content confirmed → investigate why recreate produces thin output (writer-input question: does it load existing content_data on recreate?).
- CONFIRMED ELIMINATED so far for index: load, concurrent-deploy, claim-lease, caller-timeout, component-lock. CONFIRMED: builds+deploys successfully, components frozen 06-06, guard armed (23041 chars). LEADING: regression guard blocking a thin recreate-writer output; pending _diag1 terminal signature + writer output length.
- save-failure-visible migration: still correct and worth applying (surfaces the block as needs_human_review regardless of why content is thin) — but it's the VISIBILITY fix; the ROOT fix is making the recreate writer preserve/produce full content. Engine error_step-resolution prerequisite still stands before applying the migration.

## _diag1 REPRODUCED the signature; "thin content" NOT yet proven (2026-06-15)
- _diag1: status=complete, empty result_page_id, components STILL frozen 06-06, no fresh git commit, handler orch cd73eea6 ran ~612s COMPLETED. Defect reproduced exactly: rebuild runs, completes, deploys nothing new, work item shows success. Confirms the silent-completion signature (regression-guard-via-complete_error remains the leading mechanism — empty page_id matches the guard's `return nil,err`).
- CAUTION (do not over-read): the writer-output length query returned NULL produced_html_len for both cd73eea6 and the writer child 472eed7d (which shows sections_ready=5). NULL = the path collected_data->'page_content'->'response'->>'page_html' does not exist, NOT "short". So "thin content trips the ratio" is STILL UNPROVEN — I have not actually measured what the writer produced. sections_ready=5 confirms the writer RAN, not how much text.
- USER'S GATE (correct, do this BEFORE more writer digging): confirm the guard's arming across pages — are all pages >200 existing chars, or only some? Query added (existing_text_chars per page + /4 threshold + last_component_update). Reading:
  - If all pages >>200 AND the others still updated (last_component_update 06-14/15) while index is 06-06 → discriminator is that index's NEW content is uniquely thin (others cleared the ratio) → points at recreate/preserve writer producing less for index.
  - If the updated pages are <200 → guard simply wasn't armed on them; cleaner story.
- NEXT (find the real writer-output key before measuring): jsonb_object_keys(collected_data->'page_content'->'response') on 472eed7d; if null, widen to collected_data->'page_content' keys / sections_for_render. Measure produced length only once the correct key is found — do NOT read null as small.
- ELIMINATED for index: load, concurrent-deploy, claim-lease, caller-timeout, component-lock (all NULL lock fields). CONFIRMED: builds+deploys success, components frozen 06-06, guard armed on index (23041), _diag1 reproduces signature. LEADING: guard blocking; thin-content cause PENDING the per-page 200-char comparison + a correctly-keyed writer-output measurement.

## Per-page text comparison — CONFOUND identified, guard mechanism refined (2026-06-15)
- Per-page existing_text_chars (deployed components), sorted desc, with last_component_update:
  - HIGH-text, all stuck at 06-06: index 23041, game-jelly-invaders 20700, game-p2p-networking 20196, tool-progression-architect 16678, game-economy-simulator 15072, game-auto-battler 14544, tool-ehp 13478, tool-ttk 13376, tool-drop-rate 13040, tool-lanchester 13029, tool-jump-physics 11848 (+ guide-economy-basics 3408 outlier).
  - LOW-text, all updated 06-14/15: guides-index 7570, tools-index 7178, games-index 6938, about-index 4516, guide-rng-design 3993, game-pathfinding 3618, guide-skinner-box 3414, guide-p2p-architecture 3203, guide-fairness-in-rng 2940, contact-index 1914.
- CONFOUND (corrects my earlier framing): the high-text pages are stuck at 06-06 because they were NEVER IN THE §5 REBUILD BATCH (the batch was the 21 content pages: index, *-index hubs, guide-* detail pages — NOT the tool-* / game-* DETAIL pages). Their 06-06 staleness ≠ a rebuild failure; nothing tried to rebuild them. So I must NOT read "high text ↔ stale" as the mechanism — the real variable there is "was it rebuilt".
- THE CLEAN COMPARISON: among pages that WERE rebuilt, index (23041) is the ONLY high-text one; every other rebuilt page is <7600 and updated fine. So the guard hypothesis SURVIVES but for a sharper reason: the guard's block threshold = existing/4 SCALES with existing text. about-index threshold=1129 (trivially cleared); index threshold=5760. A normal regenerated hero+lists page (~4-5k text) would CLEAR about-index's bar but FAIL index's. Coherent + specific: same-sized regenerated output passes everywhere except the one page whose existing text is 3-5x larger.
- STILL UNMEASURED (do not assert): writer-output query returned 0 ROWS — collected_data->'page_content'->'response' has NO keys for 472eed7d. So I have NOT measured new_text_chars; the "~4-5k clears about-index not index" is arithmetic-plausible, not confirmed. 472eed7d had sections_ready=5 so it's a relevant orch but its HTML is stored elsewhere (or it's not the writer child).
- DECISIVE CONFIRMATION (either):
  1. The CONTENT REGRESSION BLOCKED log line for _diag1's 15:23 run (recent — may still be in chassis logs): kubectl ... logs --tail=5000 | grep -iE 'CONTENT REGRESSION|SavePageSections.*index'. Gives existing_text_chars + new_text_chars directly → ends the inference.
  2. If rolled: locate 472eed7d's actual output key — jsonb_object_keys(collected_data) — find where produced HTML lives, measure its stripped length, compare to 5760.
- NET: guard-blocking-a-regenerated-page is now the coherent leading mechanism with a scaling-threshold rationale; index is the only rebuilt page large enough to trip it. PENDING one decisive number (new_text_chars) to confirm vs assert.
