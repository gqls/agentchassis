# EXTRACTION U24d — docs/_archive adoption/ + content_quality_and_internal_linking/ (gap-fill)
Extracted 2026-07-13. Files in scope: 74 (adoption/ 35, content_quality_and_internal_linking/ 39). Concepts found: 21.

## Method note
This unit is a gap-fill: every file here is an **archived duplicate or earlier draft** of material also covered live under `docs/agent_docs/docs024_key_docs_latest/adoption/` (U09) and `.../content_quality_and_internal_linking/` (U05). Per-file, each archive file was diffed (`diff -q`) against its plausible live counterpart(s) (matched by base filename, stripping the `(N)` suffix). Two patterns dominated:
1. **Byte-identical duplicates** (cloud-sync-style `(N)` copies, or a `package_module/output_contexts` "context pack" bundle that literally concatenates copies of other docs for LLM-context assembly). These carry no unique content; treated `family-delta` or `skipped-generated` with zero new concepts, but every file is still listed in Coverage per the audit-trail rule.
2. **Genuine earlier drafts** of a still-live doc, where the live version is a later dated revision (e.g. `FOCUS_content_quality.md` 2026-06-09 archived vs `FOCUS_content_quality(2).md` 2026-06-10 live) or a later part of the same growing running-notes log. These were read in full and diffed against the live continuation to surface exactly what got **corrected, rejected, or superseded** — this is the archive's unique value and where nearly all concepts below come from.

The three "running_notes_NN" series (14, 15, 17) are not independent documents — each `(N)` is a fuller snapshot of the *same* cumulative session log (verified: `(1)` is a byte-for-byt prefix of `(20)` for running_notes_14). So only the highest-numbered file in each archived series was read in full (`family-latest` within this archive's slice); the rest are confirmed-subset duplicates.

## Coverage
| file | treatment |
|---|---|
| adoption/CATALOGUE_gamesdesign_post_sync_fix_defects.md | full |
| adoption/CATALOGUE_gamesdesign_post_sync_fix_defects(1).md | family-delta (intermediate draft, superseded by (4)) |
| adoption/CATALOGUE_gamesdesign_post_sync_fix_defects(2).md | family-delta (intermediate draft) |
| adoption/CATALOGUE_gamesdesign_post_sync_fix_defects(3).md | family-delta (intermediate draft; diffed against (4)) |
| adoption/CATALOGUE_gamesdesign_post_sync_fix_defects(4).md | full |
| adoption/docubundle/CONTEXT_PACK_adoption_skinner_box.md | family-delta (byte-identical to live `adoption/CONTEXT_PACK_adoption_skinner_box.md`) |
| adoption/docubundle/GUIDE_deploy_from_context_packs(1).md | full |
| adoption/docubundle/analyser(2).go | header-scan |
| adoption/docubundle/assembler(2).go | header-scan |
| adoption/old2/FOCUS_llm_reliability_for_component_generation.md | family-delta (byte-identical to live top-level `FOCUS_llm_reliability_for_component_generation.md`) |
| adoption/running_notes_14_sync_fix_and_adoption_rerun.md | family-delta (early snapshot; subset of (20)) |
| adoption/running_notes_14_sync_fix_and_adoption_rerun(1).md | family-delta (subset of (20)) |
| adoption/running_notes_14_sync_fix_and_adoption_rerun(2).md | family-delta (subset of (20)) |
| adoption/running_notes_14_sync_fix_and_adoption_rerun(3).md | family-delta (subset of (20)) |
| adoption/running_notes_14_sync_fix_and_adoption_rerun(4).md | family-delta (subset of (20)) |
| adoption/running_notes_14_sync_fix_and_adoption_rerun(5).md | family-delta (subset of (20)) |
| adoption/running_notes_14_sync_fix_and_adoption_rerun(6).md | family-delta (subset of (20)) |
| adoption/running_notes_14_sync_fix_and_adoption_rerun(8).md | family-delta (subset of (20)) |
| adoption/running_notes_14_sync_fix_and_adoption_rerun(9).md | family-delta (subset of (20)) |
| adoption/running_notes_14_sync_fix_and_adoption_rerun(10).md | family-delta (subset of (20)) |
| adoption/running_notes_14_sync_fix_and_adoption_rerun(11).md | family-delta (subset of (20)) |
| adoption/running_notes_14_sync_fix_and_adoption_rerun(12).md | family-delta (subset of (20)) |
| adoption/running_notes_14_sync_fix_and_adoption_rerun(14).md | family-delta (subset of (20)) |
| adoption/running_notes_14_sync_fix_and_adoption_rerun(15).md | family-delta (subset of (20)) |
| adoption/running_notes_14_sync_fix_and_adoption_rerun(16).md | family-delta (subset of (20)) |
| adoption/running_notes_14_sync_fix_and_adoption_rerun(17).md | family-delta (subset of (20)) |
| adoption/running_notes_14_sync_fix_and_adoption_rerun(18).md | family-delta (subset of (20)) |
| adoption/running_notes_14_sync_fix_and_adoption_rerun(19).md | family-delta (subset of (20)) |
| adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md | full |
| adoption/running_notes_15_skinner_box_and_adoption_sections.md | family-delta (subset of (5)) |
| adoption/running_notes_15_skinner_box_and_adoption_sections(1).md | family-delta (subset of (5)) |
| adoption/running_notes_15_skinner_box_and_adoption_sections(2).md | family-delta (subset of (5)) |
| adoption/running_notes_15_skinner_box_and_adoption_sections(3).md | family-delta (subset of (5)) |
| adoption/running_notes_15_skinner_box_and_adoption_sections(4).md | family-delta (subset of (5)) |
| adoption/running_notes_15_skinner_box_and_adoption_sections(5).md | full |
| content_quality_and_internal_linking/FOCUS_content_quality.md | full |
| content_quality_and_internal_linking/FOCUS_internal_linking.md | full |
| content_quality_and_internal_linking/PLAN_b4_b5_hubs_and_link_resolver(1).md | family-delta (byte-identical to live `PLAN_b4_b5_hubs_and_link_resolver.md`; read fully, rejected-options captured) |
| content_quality_and_internal_linking/PLAN_pathfinding_missing_game(1).md | family-delta (byte-identical to live `PLAN_pathfinding_missing_game.md`) |
| content_quality_and_internal_linking/RUNBOOK_linking_phantom_fixes(1).md | family-delta (byte-identical to live `RUNBOOK_linking_phantom_fixes(0).md`) |
| content_quality_and_internal_linking/package_module/output_contexts/001_development_guide(3).md | skipped-generated (context-pack bundle copy of spine doc `001_development_guide`, owned by another unit) |
| content_quality_and_internal_linking/package_module/output_contexts/002_system_architecture.md | skipped-generated (byte-identical to live top-level `002_system_architecture.md`) |
| content_quality_and_internal_linking/package_module/output_contexts/016_debugging_guide_v2.md | skipped-generated (earlier bundled copy of spine doc `016_debugging_guide`) |
| content_quality_and_internal_linking/package_module/output_contexts/019_tool_library.md | skipped-generated (bundled copy of spine doc `019_tool_library`) |
| content_quality_and_internal_linking/package_module/output_contexts/020_tool_lifecycle.md | skipped-generated (bundled copy of spine doc `020_tool_lifecycle`) |
| content_quality_and_internal_linking/package_module/output_contexts/026_component_regeneration_flow(1).md | skipped-generated (byte-identical to live `026_component_regeneration_flow(1).md`) |
| content_quality_and_internal_linking/package_module/output_contexts/CATALOGUE_gamesdesign_post_sync_fix_defects(9).md | family-delta (byte-identical to live `adoption/CATALOGUE_gamesdesign_post_sync_fix_defects(9).md`) |
| content_quality_and_internal_linking/package_module/output_contexts/FOCUS_content_quality.md | family-delta (byte-identical to this unit's own `FOCUS_content_quality.md`) |
| content_quality_and_internal_linking/package_module/output_contexts/FOCUS_internal_linking.md | family-delta (byte-identical to this unit's own `FOCUS_internal_linking.md`) |
| content_quality_and_internal_linking/package_module/output_contexts/FOCUS_llm_reliability_for_component_generation.md | family-delta (byte-identical to live top-level copy) |
| content_quality_and_internal_linking/package_module/output_contexts/FOCUS_page_build_handler_silent_completion.md | family-delta (byte-identical to live top-level ORIGINAL, i.e. without the 2026-06-09 update present in live `adoption/FOCUS_page_build_handler_silent_completion(1).md`) |
| content_quality_and_internal_linking/package_module/output_contexts/HANDOFF_2026-06-02_hero_resolver_and_section_data_reconciler.md | family-delta (byte-identical to live top-level copy) |
| content_quality_and_internal_linking/package_module/output_contexts/HANDOFF_2026-06-06_guide_list_and_skinner_box.md | family-delta (byte-identical to live `adoption/HANDOFF_2026-06-06_guide_list_and_skinner_box.md`) |
| content_quality_and_internal_linking/package_module/output_contexts/HANDOFF_2026-06-09_sections_durability_and_content_quality.md | family-delta (byte-identical to live `adoption/HANDOFF_2026-06-09_sections_durability_and_content_quality.md`, earlier than live content_quality (1)/(2)) |
| content_quality_and_internal_linking/package_module/output_contexts/RUNBOOK_section_sectionless_durability.md | family-delta (byte-identical to live `adoption/RUNBOOK_section_sectionless_durability.md`) |
| content_quality_and_internal_linking/package_module/output_contexts/running_notes_15_skinner_box_and_adoption_sections(9).md | family-delta (byte-identical to live `adoption/running_notes_15_skinner_box_and_adoption_sections(9).md`) |
| content_quality_and_internal_linking/package_module/output_contexts/running_notes_16_content_quality_and_internal_linking.md | family-delta (byte-identical to live `running_notes_16_content_quality_and_internal_linking.md`) |
| content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes.md | family-delta (subset of (16)) |
| content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(1).md | family-delta (subset of (16)) |
| content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(2).md | family-delta (subset of (16)) |
| content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(3).md | family-delta (subset of (16)) |
| content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(4).md | family-delta (subset of (16)) |
| content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(5).md | family-delta (subset of (16)) |
| content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(6).md | family-delta (subset of (16)) |
| content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(7).md | family-delta (subset of (16)) |
| content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(8).md | family-delta (subset of (16)) |
| content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(9).md | family-delta (subset of (16)) |
| content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(10).md | family-delta (subset of (16)) |
| content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(11).md | family-delta (subset of (16)) |
| content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(12).md | family-delta (subset of (16)) |
| content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(13).md | family-delta (subset of (16)) |
| content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(14).md | family-delta (subset of (16)) |
| content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(15).md | family-delta (subset of (16)) |
| content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(16).md | full |

## Concepts

### Context-pack code-gathering tooling: analyser/assembler vs package_*.sh
- **category:** NEW:context-pack-tooling
- **status-signal:** superseded
- **status-evidence:** The archived `GUIDE_deploy_from_context_packs(1).md` (adoption/docubundle) devotes half its length to a considered trade-off between the directory-walking `package_page_build_debug.sh` script ("broad, thorough, large... 110 files when the task really turns on about 14... caught `registry.go`") and the call-graph-based `analyser.go`/`assembler.go` pair ("leaner... but currently blind to wiring files" like `registry.go`, because registration happens via an init/registry mechanism the call graph never reaches). The live rewritten guide (`docubundle/GUIDE_deploy_from_context_packs.md`) drops this entire discussion — it restructures around a deploy-mechanism reference (A–F) and per-project quick reference, never mentioning the analyser/assembler tool or its registry.go blind spot at all.
- **what:** Two competing tools for assembling an LLM chat's working context from a Go repo: (1) `package_*.sh`, which concatenates whole hand-picked directories plus a live DB/pod capture into one text bundle; (2) `analyser.go` (structural JSON index of the repo) + `assembler.go` (pulls only named functions plus their call-graph neighbourhood into a tight bundle, given a `-scope`/`-task`/`-constitution` spec). The archived guide is careful to flag that the assembler's call-graph approach misses non-call wiring (e.g. `init()`-based registry.go registration) that the script's brute-force directory walk happens to catch.
- **sources:** adoption/docubundle/GUIDE_deploy_from_context_packs(1).md; live docubundle/GUIDE_deploy_from_context_packs.md; adoption/docubundle/analyser(2).go; adoption/docubundle/assembler(2).go
- **relations:** the analyser/assembler pair itself lives on with a proper home under `docs019_documentation_audit_autonomous_build_and_operate/_archive/thin_slice_run/` (`thin_slice_constitution.md`) — so the *tool* isn't abandoned, but this guide's discussion of when/why to reach for it over the script was dropped from the doc lineage.
- **verify-later:** whether analyser.go/assembler.go are still invoked anywhere in practice, or whether package_*.sh fully displaced them for chassis debugging work.

### Chassis deploy-mechanism reference (targets A–F)
- **category:** NEW:deploy-mechanics-reference
- **status-signal:** deployed
- **status-evidence:** live `docubundle/GUIDE_deploy_from_context_packs.md` names six distinct deploy mechanisms (A: chassis image rebuild+rollout, B: DB/SQL migration, C: work-item insert, D: orchestration `orchestrate` trigger via kcat, E: generated static site via git→GitHub Actions→Backblaze, F: idea.uk standalone binary) and a per-project quick reference mapping each named task to its mechanism(s).
- **what:** A structured taxonomy of "what shipping a change actually means" per target: the agent-chassis Kubernetes image is a different deploy surface from the sites it builds (Backblaze-hosted static output) which is different again from the idea.uk box (file-based, cPanel, no k8s/DB). The archived draft only had this half-formed (a looser walkthrough focused on one task, skinner-box); the live version generalized it into the reusable A–F reference.
- **sources:** adoption/docubundle/GUIDE_deploy_from_context_packs(1).md; live docubundle/GUIDE_deploy_from_context_packs.md
- **relations:** adapters (033/035), deployment-github (034), storage-architecture (032)
- **verify-later:** confirm the A–F reference still matches current `makefile.txt`/`kustomization.yaml` targets.

### Defect-cataloguing discipline (enumerate-before-fixing)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** `CATALOGUE_gamesdesign_post_sync_fix_defects.md` states its purpose explicitly: "Enumerate every observed defect as a *separate* item before fixing, so distinct causes are not conflated into one rolling investigation," with causes marked "tentative" until confirmed by reading source. Later revisions (`(4)`) show the discipline paying off: defects graduate from `[NEW]`/`[PARKED]` through `[FIX SHIPPED — PARTIALLY VERIFIED]` to `[VERIFIED CLOSED]` with a pinned, source-read cause replacing the original tentative one.
- **what:** A working method for a real adoption-run defect sweep: group symptoms into lettered families by shared mechanism (A deployment gaps, B link fallbacks, C list-component content, D section-data gaps, E content quality, F guide duplication, G design fidelity, H hygiene, I open unknowns, J dispatch throughput), triage by root cause not symptom, and forbid shipping a fix from a "tentative" cause without first reading the responsible action.
- **sources:** adoption/CATALOGUE_gamesdesign_post_sync_fix_defects.md; adoption/CATALOGUE_gamesdesign_post_sync_fix_defects(4).md
- **relations:** running-notes debugging-log convention (below), silent-completion family
- **verify-later:** whether this catalogue format was formalised anywhere beyond this one adoption run.

### SyncPagesToDBAction / WriteSitePlanAction canonicalisation divergence — Option 1 rejected
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** running_notes_14(20) Part 2: "**Option 1 (single source of truth):** sync reads identity from `site_plan_pages`... **Decision: Option 2**... Corrected an earlier framing that called Option 1 'the structural one' — Option 2 is the structural fix here; Option 1 is coupling."
- **what:** Two canonicalisation surfaces disagreed — `WriteSitePlanAction` ran `ValidateRoles + CanonicalisePage` (producing correct `section-index` hubs in `site_plan_pages`), while `SyncPagesToDBAction` ran `CanonicalisePage` alone on raw `page_plan` (producing flat `pages` rows), and `upsertPage`'s `ON CONFLICT` then overwrote the correct row with the flat one. Option 1 (make sync read the already-validated `site_plan_pages`) was rejected because `pageflow-builder` (confirmed active) and two other callers invoke sync with no plan ever written, so Option 1 would silently break them. The shipped fix (Option 2) runs `ValidateRoles` inside sync too, unifying the pipeline across all five callers.
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Part 1–3
- **relations:** `pageflow-builder` deprecation (decoupled from this fix, tracked separately), guide page_type restructuring
- **verify-later:** `SyncPagesToDBAction`/`site_db_actions.go` current state; whether `pageflow-builder` was ever actually deprecated.

### "Type guides as `guide`" — falsified as a quick companion fix, later built properly as a structural fix
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** running_notes_14(20) Part 2: "**Falsified the 'type guides as `guide`' companion.** ... source guides live flat at `/blog/guide-rng-design.html`, while a `guide` role would *nest* them... typing `guide` would be *less* faithful... Left as an open product decision; did NOT ship the wrong patch." Then Part 14 (2026-06-04), same session-log lineage: `migration_retype_guides_to_guide.sql` + `migration_guides_url_to_canonical.sql` were written, applied, and guides were deliberately moved to `/guides/<slug>/index.html` — the exact "less faithful" move earlier rejected, now chosen deliberately once `guide` became a first-class page_type with its own canonicalisation rule.
- **what:** Two-stage decision on how adopted "guide" content should be typed/URL'd. First pass: rejected retyping guides as `guide` as a *quick fix* for a de-prefixing side-effect, because it would misplace the URL relative to the untouched source. Second pass, as a *deliberate structural project*: added `guide` to the page_type enum, re-typed the 5 real guide-* pages, added the classifier's default_config guidance, and migrated their URLs from `/blog/guide-*.html` to `/guides/<slug>/index.html` — closing the exact gap the earlier rejection had flagged as "an open product decision."
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Parts 2, 13, 14, 14a–14h
- **relations:** SyncPagesToDBAction canonicalisation divergence; bare-guide-duplicate defect (below); adoption-faithfulness-via-locks (below)
- **verify-later:** `pages.page_type` enum and `page_canonical.go`'s `guide` case in current code; whether `build-site-planner`'s vocabulary was ever updated to include `guide` for new adoptions.

### Bare-guide duplicate pages — root cause: planner ignores adopted state (prompt-rule gap, not wiring)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes_14(20) Part 14e: "DECISIVE (llm_call_log plan_site)... `saw_guide_pages=t`, `prompt_says_no_existing=f`, `planned_bare_in_response=t`... So the planner WAS given the adopted guides and emitted `economy-basics` anyway → PROMPT-RULE gap... NOT a wiring/status gap." Cleanup migration (`migration_cleanup_bare_guide_duplicates.sql`) applied and confirmed durable (Part 14f: "current-plan bare-name query returns 0 rows").
- **what:** `build-site-planner` re-invents a differently-slugged sibling page (`economy-basics`) for a topic already adopted under a prefixed name (`guide-economy-basics`), because its "never duplicate an existing page" prompt rule only named games/tools examples and didn't generalise to the `guide-` prefix pattern. This is a fresh, concretely-diagnosed instance of the previously-documented `FOCUS_planner_ignores_adopted_state.md` mechanism (2026-05-19). A durable Go-level fix (deterministic topic-stem collision guard in `validate_site_plan`/`write_site_plan`, reusing `CanonicalisePage`'s prefix-stripping) was recommended but not shipped in this arc; only the data cleanup + an optional prompt-rule stopgap were delivered.
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Parts 14c–14g
- **relations:** adoption-faithfulness-via-locks convergence (below); "type guides as guide" (above)
- **verify-later:** `FOCUS_planner_ignores_adopted_state.md`; whether the Go-level topic-stem guard was ever built.

### Adoption-faithfulness-via-locks convergence — confirmed INERT
- **category:** site-plan-and-reconciler
- **status-signal:** abandoned
- **status-evidence:** running_notes_14(20) Part 14h: "TRUE root cause: `reconcilePlanWithRealised` gates on `rm[\"adoption_locked\"]`; the live `load_existing_pages` query does NOT emit `adoption_locked` ... lockedPages always empty -> reconcile ALWAYS no-ops." And: "`FOCUS_adoption_faithfulness_via_locks.md` status — convergence 'Inert until 054 + write_site_plan land.' ... LIVE STATE: lock tables have ONLY `locked_at`/`locked_by` — NO `lock_type`/`lock_expires_at` -> 053 NOT applied... 054 NOT applied."
- **what:** A designed subsystem meant to make adoption re-plans faithful to already-realised (locked) pages — schema migration 053 (lock_type/lock_expires_at columns), migration 054 (`load_existing_pages` emits `adoption_locked`), and `write_site_plan` locking logic — was found, on live inspection, to be entirely unapplied. The one piece that *was* built (`reconcilePlanWithRealised`'s convergence check in `v3_site_actions.go`) silently no-ops because its input is never populated. This directly explains two other defects in the same arc (the bare-guide duplicates, and 5 guide pages never being unioned into the plan).
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Part 14h; (references live `FOCUS_adoption_faithfulness_via_locks.md`, `031_locks(3).md`)
- **relations:** bare-guide duplicate pages; sync/write-site-plan divergence
- **verify-later:** whether migrations 053/054 have since been applied; current state of `write_site_plan_action.go`'s `transferDirectiveLocks`.

### Deployed→needs_rebuild ON CONFLICT flip — pre-design stand-in later completed properly (Option B)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** running_notes_14(20) Part 8: "the flip is a pre-design *stand-in* for 're-sync invalidates deployed pages'... It over-fires (every deployed page, every sync) and mis-fires on pre-plan deploys (tools)... **Option B shipped**: COALESCE fill-if-null; removed the `deployed→needs_rebuild` CASE branch... Drift now flows through the reconciler's `decideEmit`."
- **what:** `upsertPage`'s `ON CONFLICT` branch that flipped any `deployed` page back to `needs_rebuild` on every sync was a workaround for a never-shipped design: `029`/`030` intended `built_from_plan_version` to be stamped at build time and drift detected by the reconciler, but the stamp was "explicitly deferred" per `HANDOFF_2026-05-07` #5 ("User explicitly OK'd this"). The investigation confirmed the flip should be completed as originally designed rather than patched around (rejecting a narrower "Option A: exclude tool/game from the flip" as entrenching the workaround) — shipped as the deploy-time stamp in `UpdatePageStatusAction` + COALESCE fill-if-null in `upsertPage`.
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Part 8; CATALOGUE_gamesdesign_post_sync_fix_defects(4).md A1
- **relations:** A1 tool/game deploy-gap root cause (below)
- **verify-later:** `v3_site_actions.go` `UpdatePageStatusAction`, `site_db_actions.go` `upsertPage` current state.

### A1 — tool/game pages never deployed: root-cause hypothesis evolution
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** Original catalogue (base, 2026-05-28): "*Tentative area:* the gap is between 'page row built/complete in DB' and 'file committed to git'. Could be the rerender/deploy path skipping nested child pages, the tool-recreation handler not producing a deployable artefact, or a git-path mismatch." Catalogue `(4)` (2026-06-04): "*Cause pinned (two coordinated root causes):* 1. Parser... `saveSectionsExtractFromHTML` extracts only `<section>` blocks, but `tool-recreation-handler`'s prompt emits `<div class="tool-page">` (no `<section>`)... 2. Flip churn: `upsertPage`'s ON CONFLICT flipped `deployed → needs_rebuild`."
- **what:** The site's actual interactive product (tools/games) never deployed a file despite `pages` rows and `complete` work items. The three original hypotheses (deploy-path bug, handler artefact-production bug, git-path mismatch) were all superseded by two pinned, source-confirmed causes: an HTML-fragment parser that only recognises `<section>` blocks (tool output uses `<div class="tool-page">`, so it silently extracted zero sections), plus the ON CONFLICT flip churn (above). Fix: single-fragment fallback in the parser + the Option B stamp/flip removal. Verified end-to-end on a subsequent adoption run (all 5 games + tools deployed with working links).
- **sources:** adoption/CATALOGUE_gamesdesign_post_sync_fix_defects.md; adoption/CATALOGUE_gamesdesign_post_sync_fix_defects(4).md; running_notes_14(20) Parts 7–10
- **relations:** deployed→needs_rebuild flip (above); dispatch throughput bottleneck (Family J, below)
- **verify-later:** `save_page_sections_action.go` current parser logic.

### A4/homepage-missing-file — root cause hypothesis evolution to "auto-complete on lost response"
- **category:** debugging
- **status-signal:** superseded
- **status-evidence:** running_notes_14(20) Part 10 (2026-06-03): "Candidate cause (NOT concluded): ...silent non-fast-forward [git race]. Alternatives not ruled out: empty-assembly skip... or a path bug." CATALOGUE(3)→(4) diff (2026-06-04): "*Cause (updated... git race RULED OUT)*: ...empty-assembly case" superseded again by "*Cause PINNED*: ...work item is `complete` with error 'Auto-completed: work verified done despite lost response.'" Running_notes Part 11 confirms: "Root cause: the homepage's content build was dispatched, the handler's response was lost... and the recovery path optimistically auto-completed the work item without verifying the artifact."
- **what:** The homepage (`index`) was `build_status='deployed'`+`stamped` in the DB with zero rendered components and no committed file — three successive hypotheses (git-commit race, empty-assembly/planner-vs-composition gap, and finally the pinned cause) were tested and discarded in turn before landing on: a scheduled task's SQL `pre_query` (`claimed-item-timeout`) auto-completed a claimed work item using loose evidence ("any page on the site updated since claim") after the handler's response was lost to a pod death, without checking that *this* page actually produced components.
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Parts 10–12; CATALOGUE_gamesdesign_post_sync_fix_defects(3).md vs (4)
- **relations:** claimed-item-timeout evidence-check reliability mini-project (below); silent-completion family (page-build-handler, save_page_sections)
- **verify-later:** `migration_claimed_item_timeout_evidence_v2.sql` application state; `v3_site_actions_optionB.patch` deployment.

### claimed-item-timeout evidence-gated completion + reset (Lever A/C) — avoided building a duplicate watchdog
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** running_notes_14(20) Part 12: "'Auto-completed...' is set by the **`claimed-item-timeout` scheduled task's SQL `pre_query`**, NOT a Go reaper... `migration_claimed_item_timeout_evidence_check.sql` ... is essentially Option A + Lever C, already authored... the FOCUS_dispatch `reset_stale_claims` watchdog is redundant; do NOT build it." Part 12 addenda confirm the v2 migration (page_components-based evidence, not the untrustworthy `build_status='deployed'` flag) applied and verified live, plus the companion `pageHasComponents` deploy-guard (Option B) delivered.
- **what:** A `claimed-item-timeout` scheduled task's `pre_query` already implements both (a) evidence-gated auto-completion of stuck claims (only complete if the specific artefact shows positive evidence) and (b) a stale-claim reset-to-`triaged`/`failed` after a timeout with attempt counting. Mid-investigation, an agent nearly built a brand-new "reset stale claims" watchdog before discovering this — a documented reuse-over-build catch. The evidence signal itself evolved further: from trusting `pages.deployed_at`/`build_status='deployed'` (provably lying, per the homepage case) to checking `page_components.updated_at > claimed_at` directly.
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Parts 12–12 addenda
- **relations:** A4 homepage root cause (above); sectionless-page durability stack (below); dispatch throughput (Family J)
- **verify-later:** current `claimed-item-timeout` pre_query SQL in `scheduled_tasks`.

### Sectionless-page silent completion (guide-skinner-box) + durability stack
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** running_notes_15(5) Part 1: "`complete_error` is a `complete_workflow` (SUCCESS) with message *'Content writer skipped — page has no sections defined.'* — the silent-success smell." Part 8: "Wrote `discovery_checks/check_sectionless_pages.go` (new check...)" with an explicit 4-layer "durability stack" logged, item 3 of which is marked "(next, optional for this gap; cleanliness)" i.e. not yet done at time of writing.
- **what:** A page reaching page-build with zero planned sections (`pages.sections=[]`) causes `check_has_ready_sections` to route to `complete_error`, which is a SUCCESS-labelled `complete_workflow` — so a genuinely broken page is marked `complete` and never retried. Root-caused (after correctly ruling out the convergence-union code, confirmed correct) to "the gap is reconciliation: nothing repairs a page in-plan with zero sections." Fix stack: (1) a read-time fallback in `load_page_sections_from_spec_action.go` that copies a same-role sibling's section layout ("skeleton only, not content") when both known sources are empty; (2) a new discovery check `check_sectionless_pages.go` that detects and retriggers stuck sectionless pages (chosen over patching the existing but **dormant** `checkEmptyPageSections`, see below); (3) a workflow-level fix so the genuinely-unrecoverable case routes to a flagged state instead of `complete` — logged as not yet shipped; (4) the broader positive-evidence-completion mini-project (shared with A4).
- **sources:** adoption/running_notes_15_skinner_box_and_adoption_sections(5).md Parts 1–8
- **relations:** dormant discovery-check machinery (below); A4 auto-complete-on-lost-response; FOCUS_page_build_handler_silent_completion.md
- **verify-later:** whether S2 (workflow-level flagged-state fix) ever shipped; `check_sectionless_pages.go` enablement state.

### Dormant discovery-check machinery (`checkEmptyPageSections` / `validate_component_standards`)
- **category:** improvement-loop
- **status-signal:** abandoned
- **status-evidence:** running_notes_15(5) Part 6: "**RESULT (decisive):** ... `validate_component_standards` (its wrapper) is **not enabled in any** discovery agent (`enables_vcs=f` for all three)... The empty-page detector is **dormant code**, not a buggy check."
- **what:** A pre-existing check (`checkEmptyPageSections`, inside `ComponentStandardsCheck`/`validate_component_standards`) already targets exactly "page with no rendered sections," but was never added to any discovery agent's `checks` config array, so it has literally never fired in production — its 11 historical `needs_content_page` items were all traced to adoption-run/manual sources, none to this check. It was also found to be scoped too narrowly (`deployed`/`active` only, missing `planned`) and to recover by re-emitting a still-empty spec (would loop, not repair) — reasons a *new* dedicated check (`check_sectionless_pages.go`) was written instead of extending the dormant one.
- **sources:** adoption/running_notes_15_skinner_box_and_adoption_sections(5).md Parts 6–8
- **relations:** sectionless-page silent completion (above)
- **verify-later:** `discovery_checks/` registry current contents; whether `validate_component_standards` has since been enabled anywhere.

### Hero/CTA link fabrication in `sourceResolver.resolve` — the "310–318 hardcoded fallback nav" hypothesis superseded
- **category:** link-management
- **status-signal:** superseded
- **status-evidence:** Archived `FOCUS_internal_linking.md` (2026-06-09): "**Defect (lines 310–318 of `multipage_actions.go`):** when nav resolution returns empty, `AssembleMultipageSiteAction` injects a **hardcoded fallback nav**... This generic brochure default is a primary source of the phantom `/services.html`." Live `FOCUS_internal_linking(1).md` (2026-06-10): "**Nav is already real-page-derived; the brochure fallback was not the live path.**... The header/footer phantoms... came from **hardcoded `ContentData` in `render_site_components_action.go`**... not from the `multipage_actions.go` 310–318 fallback nav... (**Correction to the earlier note that blamed 310–318**.)"
- **what:** The initial (2026-06-09) diagnosis of site-wide phantom links blamed a specific hardcoded fallback-nav code path (`multipage_actions.go:310-318`) as the likely root cause. The next day's investigation, grounded in reads of `render_site_components_action.go`, corrected this: nav was already correctly real-page-derived, and the actual mechanism was (a) `sourceResolver.resolve`'s `"pages"` case *fabricating* a URL (`"/"+path+".html"`) and returning `found=true` for any non-existent page (so schema `on_missing`/`fallback` never fired), plus (b) separately, hardcoded `ContentData` literals for header/footer CTAs and legal links.
- **sources:** content_quality_and_internal_linking/FOCUS_internal_linking.md (archived, 2026-06-09); live FOCUS_internal_linking(1).md (2026-06-10); running_notes_17(16) "Decisive findings"
- **relations:** component-template-fixer CTA-reuse assumption; link_registry hypothesis (below)
- **verify-later:** `plan_sections_action.go` `sourceResolver.resolve` current "pages" case; `render_site_components_action.go` ContentData construction.

### `component-template-fixer` CTA-handling reuse assumption — corrected, replaced by dedicated agent
- **category:** content-quality
- **status-signal:** superseded
- **status-evidence:** Archived `FOCUS_content_quality.md` (2026-06-09): "The plan notes `component-template-fixer` 'already handles CTA fixes' — verify and extend rather than build new." Live `FOCUS_content_quality(2).md` (2026-06-10): "`component-template-fixer` exists but **explicitly punts on CTAs** (`cta_improvement`/`cta` → `fixed:false, action:'needs_review'`)... So the PLAN's 'already handles CTA fixes' was wrong; there was no CTA resolver to reuse — hence the dedicated `internal-link-resolver` (Step 3)."
- **what:** `PLAN_design-note-recommendation-specialists.md` asserted `component-template-fixer` already had CTA-fix handling that could be reused/extended for the hero-CTA phantom-link defect. Verification against the live agent's actual routing table found it explicitly declines CTA improvements, routing them to `needs_review` instead of fixing them. This wrong assumption, once corrected, directly motivated building a new dedicated agent (`internal-link-resolver`, see below) rather than extending the wrong one.
- **sources:** content_quality_and_internal_linking/FOCUS_content_quality.md (archived); live FOCUS_content_quality(2).md; running_notes_17(16)
- **relations:** internal-link-resolver agent (below); identity-advisor/sites.approval_mode (below)
- **verify-later:** `component-template-fixer`'s current action set.

### `identity-advisor` agent and `sites.approval_mode` gate — proposed, confirmed never built
- **category:** content-quality
- **status-signal:** abandoned
- **status-evidence:** Archived `FOCUS_content_quality.md` (2026-06-09) lists them as PROPOSED pieces needing verification ("Before relying on `identity-advisor` / `component-template-fixer` / `sites.approval_mode`, confirm each exists"). Live `FOCUS_content_quality(2).md`/`FOCUS_internal_linking(1).md` (2026-06-10) confirm: "`identity-advisor` does **not** exist. `sites.approval_mode` does **not** exist. The three-way `finding_type` classification and those specialists are PROPOSED, not built."
- **what:** `PLAN_design-note-recommendation-specialists.md`'s three-way finding-routing design (bug / gap / recommendation) named `identity-advisor` as the specialist for contact/email findings and `sites.approval_mode` as the gate for whether recommendation-type findings auto-apply. Neither was ever implemented — a clean case of a documented plan whose specific pieces were checked against the live schema/agent_definitions and found absent.
- **sources:** content_quality_and_internal_linking/FOCUS_content_quality.md (archived) and live FOCUS_content_quality(2).md; running_notes_17(16)
- **relations:** component-template-fixer CTA-reuse assumption (above)
- **verify-later:** re-check `agent_definitions` and the `sites` table for these names in case they were built later.

### `link_registry` as a phantom-link validation substrate — considered, found unusable, abandoned
- **category:** link-management
- **status-signal:** abandoned
- **status-evidence:** Archived `FOCUS_internal_linking.md` (2026-06-09): "**Link inventory.** `ExtractAndSyncLinksAction`... syncs them per page into `link_registry`... A per-page link inventory already exists — the natural substrate for a broken/phantom-link discovery check." Live `FOCUS_internal_linking(1).md` (2026-06-10): "**`link_registry` only records, never validates.** It *has* a `target_page_id` column + FK to `pages`, but `syncLinksToDB` never populates it. And `extract_and_sync_links` is wired into **no live workflow**, so `link_registry` is empty. It is not a usable substrate today."
- **what:** The internal-linking investigation initially proposed reusing the existing `link_registry` table/action as the base for a new phantom-link discovery check. Follow-up code reading found the table permanently empty in practice (the populating column is never written, and the syncing action isn't wired into any live workflow) — so the check that was actually built (`check_phantom_internal_links.go`) instead scans `rendered_html` directly via new shared helpers (`datahelpers/links.go`).
- **sources:** content_quality_and_internal_linking/FOCUS_internal_linking.md (archived); live FOCUS_internal_linking(1).md; running_notes_17(16) "Decisive findings"
- **relations:** hero/CTA link fabrication (above)
- **verify-later:** whether `ExtractAndSyncLinksAction`/`link_registry` were ever wired up subsequently.

### Hub "Browse All X" link resolution — rejected design options
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** `PLAN_b4_b5_hubs_and_link_resolver(1).md`: "(a) Populate the `*_index_url` specs from real hubs. **Rejected** — per-site data to maintain; re-introduces the inconsistent-source brittleness. (b) `source: pages.<hub-name>` per component... bakes the `<area>-index` naming convention into each schema... (c) **Recommended.** A new `queryresolve` verb... `query.section_index_for:<type>`." Shipped per running_notes_17(16): "`section_index_for.go` — new `queryresolve` verb... B4/B5 — done."
- **what:** For the empty-href "Browse All Tools/Games/Guides" defect, two options were explicitly weighed and rejected in the design doc before settling on a new `queryresolve` verb: manually populating `*_index_url` site_specs (rejected as brittle, per-site maintenance) and a per-component `pages.<hub-name>` source (rejected as baking a naming convention into every schema). The chosen option — `query.section_index_for:<type>`, resolving the hub via shared `site_area_id`/URL-prefix — shipped and was confirmed working in deployed HTML.
- **sources:** content_quality_and_internal_linking/PLAN_b4_b5_hubs_and_link_resolver(1).md; running_notes_17(16)
- **relations:** hero/CTA link fabrication; internal-link-resolver agent
- **verify-later:** `queryresolve.go` `section_index_for` case in current code.

### `internal-link-resolver` agent (Step 3) — dedicated intent-aware internal-link resolution
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** running_notes_17(16): "**Step 3 — completed (2026-06-11), all deliverables written**... Agent row (`internal_link_resolver_agent.sql`)... Writer wiring (`page_content_writer_link_resolver_wiring.sql`)... `unresolved_cta`: emitted in-Go... `status needs_human_review`." Confirmed live end-to-end: "Query D (corrected paths) on both completed rebuilds: `for_render=2`, `plan_count=2`, EQUAL ⇒ resolver augmented sections + writer loop consumed them."
- **what:** A new sub-agent of `page-content-writer` (modelled on `research-agent`, no persistence) that, at build time, resolves hero/CTA link destinations to intent-appropriate real pages (excluding the page's own hub, about/contact/legal) rather than a fixed contact page, validates every candidate against `datahelpers.PageURLSet`, and emits an `unresolved_cta` HITL signal when no destination can be found — the only place a "correctly dropped" (absent) button is detectable, since the deploy gate can't see an absence. Replaces the abandoned assumption that `component-template-fixer` already handled this.
- **sources:** content_quality_and_internal_linking/PLAN_b4_b5_hubs_and_link_resolver(1).md; running_notes_17(16) "Step 3" sections
- **relations:** component-template-fixer CTA-reuse assumption (superseded); identity-advisor/approval_mode (abandoned); hero/CTA fabrication fix
- **verify-later:** `internal_link_resolver_agent.sql`, `resolve_internal_links_action.go` current deployment/image-tag state.

### `save_page_sections` content-regression guard laundered into false success — theories falsified in sequence, course-corrected to a second mechanism
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** running_notes_17(16) "index deep-dive": four theories tried and explicitly falsified in order — (1) "Load — 21 rebuilds... starved the dispatcher. FALSIFIED"; (2) "Concurrent production deploy cycled the pods mid-flight. FALSIFIED for index"; (3) "index's build DURATION exceeds the... claim lease. FALSIFIED"; (4) caller-timeout theory, "partially real as a STATUS artifact... but this is noise, not the defect." Landed on: "**Content-regression guard... is the leading mechanism.**" Then a further course-correction: "**COURSE-CORRECTION: a second mechanism — page_components LOCKING**... NEW HYPOTHESIS (at least as plausible as the regression guard)."
- **what:** The homepage (`index`) repeatedly failed to rebuild despite the work item showing `complete` and git successfully committing a file — the committed file was stale (unchanged since 2026-06-06). Root cause hunt discarded four increasingly specific theories (load, concurrent deploy, claim-lease timeout, caller/callee timeout mismatch) before finding `save_page_sections_action.go`'s **content-regression guard** — a real safety check (refuses to overwrite existing deployed content with much-shorter new content) whose error return was silently laundered into `complete_error`, itself a SUCCESS-labelled `complete_workflow`. Before fully confirming this, the investigation surfaced a *second* candidate mechanism discovered via schema inspection — a `page_components` row-locking subsystem with an `auto_lock_on_deploy` trigger — and explicitly walked back single-mechanism confidence pending a discriminating query. Two distinct, independently-real bugs were named regardless of which mechanism fires: (1) the guard's legitimate refusal shouldn't route through `complete_error`; (2) deploy shouldn't proceed (re-render + git-commit) after a zero-row save.
- **sources:** content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(16).md, sections "index reproducibly stale" through "COURSE-CORRECTION"
- **relations:** A4 auto-complete-on-lost-response; sectionless-page silent completion; page-build-handler silent-completion family generally
- **verify-later:** which mechanism (regression guard vs component lock) actually fires on `index`; `page_build_handler_save_failure_visible.sql` application state; `auto_lock_on_deploy()` trigger function body.

### Dispatch throughput bottleneck (Family J) — one-site-per-tick, NOT-EXISTS-blocked
- **category:** scheduler-and-tasks
- **status-signal:** unknown
- **status-evidence:** CATALOGUE_gamesdesign_post_sync_fix_defects(4).md Family J: "the dispatcher is **one-site-per-tick** (selection `LIMIT 1`... processes ~5 items then exits) and **NOT-EXISTS-blocked** (a `NOT EXISTS` clause excludes a site *entirely* while any of its items is `status='claimed'`... line 276)... Standard manual unstick for now... **To investigate in the separate thread.**"
- **what:** Multi-tool/multi-game adoption sites drain over hours, appearing stalled, because the build-dispatch mechanism processes one site per scheduler tick and blocks an entire site's queue while any single item on it is claimed (no bounded concurrency, no per-item exclusion). A dead handler leaving a stale claim freezes the whole site until a reaper resets it. Explicitly spun out as a separate, not-yet-investigated thread rather than fixed within this arc; running_notes_17(16) later notes it's still an open TODO ("SPEED UP the rebuild pipeline... Not yet investigated").
- **sources:** adoption/CATALOGUE_gamesdesign_post_sync_fix_defects(4).md Family J; running_notes_14(20) Part 9; running_notes_17(16) "Missing-game... + speed TODO"
- **relations:** claimed-item-timeout evidence-check reliability mini-project; A1 tool/game deploy gap
- **verify-later:** `build-pipeline-trigger` dispatcher current selection logic (`LIMIT 1`, NOT EXISTS clause, line ~276 at time of writing).

### `improvement-sweep` scheduled task — deliberately disabled pending consumer readiness
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** running_notes_17(16): "Operational: `improvement-sweep` scheduled_task is **disabled** (`enabled=f`, last completed 2026-05-08), intentionally paused during core build... Before re-enabling: have the `phantom_internal_links` check enabled AND both handler agents in place (`nav-link-fixer` exists; `internal-link-resolver` is Step 3), so resuming the sweep clears findings rather than accumulating them." Later resolved to a specific enablement gate (§7) confirming per-finding routing survives `run_discovery_checks_action.go`'s pipeline stamping, so the check could finally be enabled "observe-only" without turning the sweep back on.
- **what:** The discover→triage→fix improvement loop's top-level scheduler is deliberately kept off while core build work is in flight, on the explicit policy that a discovery check should only be enabled once its handler agent actually exists — otherwise findings accumulate unconsumed rather than clearing. This is a recorded operational policy (not a bug) governing when automation is safe to turn back on.
- **sources:** content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(16).md, "Policy settled" and "§7" sections
- **relations:** dormant discovery-check machinery; internal-link-resolver agent
- **verify-later:** current `enabled` state of `improvement-sweep` in `scheduled_tasks`.
