# Register — debugging

74 concepts, consolidated from 120 raw extractions across units U01, U02, U03,
U04, U05, U06, U07, U08, U09, U10, U11, U12, U13, U14, U15, U16, U17a, U18,
U19, U20, U21, U23, U24a, U24b, U24c, U24d, U25, U26.

(Note: the raw cluster input file for this bucket was mechanically duplicated
in full — every block appeared exactly twice, byte-for-byte identical. The 120
raw-extraction figure above and every merge below is counted from the
de-duplicated content; no information was lost by discarding the second copy.)

### DBG-001 — The 016/016b debugging guide: assumption checklist + durable invariants methodology
- **status:** deployed
- **status-evidence:** Live canonical version is `016_debugging_guide_v2_58_consolidated.md` (docs024_key_docs_latest), superseding a v2_31→v2_57 lineage and continuing in 016b (v1–v8+, latest seen `016b_debugging_guide_7_3_(7).md`, 2026-07-09/10). Earliest ancestors (`old/older1/016_debugging_guide.md`, `_v2_april26`, `WM/016_debugging_guide_v2_44`) are archived/superseded drafts.
- **what:** A continuously-maintained, versioned operational runbook pairing a numbered "assumption checklist" (process-discipline items each tied to a real dated defect — e.g. per-action `_field` conventions, `input_mapping` required-by-default, `error_preview` before grepping logs, deploy≠migration-ran, config authoritative only at its runtime read-path) with a "Specific Failure Patterns" catalogue (§9, ~53 subsections in the live version) and a "durable invariants" section (016b) distilling the core doctrine: trust the rendered artefact not the status; `completed_at` is orchestration END not write instant; 0 rows is not decisive until the query is cleared; a config key read on a different path than it's set is a silent no-op; only `save_page_sections` writes `page_components`. Nearly every unit that touched docs024/docs019 independently extracted large fragments of this same guide at different version snapshots, plus a "wrong-turns" log convention recording falsified hypotheses so later passes don't re-walk them.
- **sources:** 016 §0/§9 (numbered-core unit); `016b_debugging_guide_7_3_(7).md` (travelling_docs unit); `docs/016b_debugging_guide_merged(3).md` (vonc unit); `old/older1/016_debugging_guide_v2_april26.md` (archives, superseded); `WM/016_debugging_guide_v2_44.md` (archives, superseded)
- **relations:** virtually every other entry in this register cites or extends this guide; "wrong-turns" log methodology; problem-category taxonomy (DBG-072)
- **verify-later:** current head version of 016b; whether the guide has forked across chats again (v5 fixed one prior fork)

### DBG-002 — Agent is a DB row: trust default_config over prose; two possible definition-source DBs
- **status:** deployed
- **status-evidence:** 016 §6.0, build-dispatch-loop description-vs-config contradiction example.
- **what:** Agents are rows in `agent_definitions`; the executable behaviour lives in `default_config.workflow`, not the human-readable description column, and the two can contradict each other (trust the config). `agent_definitions` may physically be read from `templates_db` or `clients_db` depending on which pod/deployment is running — confirm which copy a given running pod actually loads before patching.
- **sources:** 016 §6.0 (numbered-core unit)
- **relations:** agent_definitions clients_db vs templates_db (DBG-024); hand-applied migrations (DBG-010)
- **verify-later:** which DB each deployment reads definitions from

### DBG-003 — LLM step config field-path shadowing (ai_service resolution order; max_tokens/temperature dead paths)
- **status:** partial
- **status-evidence:** 016 §6.6 dated the bug live as of 2026-05-18 (22 of ~60 agents shadowed; structural per-field fix "planned"); a later FYI addendum (2026-07-10) confirms max_tokens specifically was DEAD CONFIG on two more agents and fixed them directly (not the general resolver).
- **what:** `ExecuteLLMPromptAction`/`execute_llm_prompt` resolves `ai_service` top-level → step-level → StepConfig and stops at the first match, so a top-level `ai_service` silently shadows every step-level override (including per-step model swaps). `max_tokens` is only read from the agent's top-level config or from inside the step's `ai_service` block — never from a step-config-root sibling — so misplaced `max_tokens` silently falls back to a hardcoded ~2048 output tokens (tell: `output_tokens` exactly 2048, or a truncated JSON verdict). `temperature` is read only from `default_config.temperature` top-level; `llm_call_log.temperature` was universally NULL. The general structural per-field fallback fix was "planned" but not confirmed shipped; individual agents have been patched one at a time as the bug recurred.
- **sources:** 016 §6.6 (numbered-core unit); `FYI_from_fixloop_2026-07-10_verdict_prompt_symptom_check.md#second-addendum` (travelling_docs unit)
- **relations:** silent no-op config-path heuristic (DBG-001); model swap functions
- **verify-later:** whether per-field resolution shipped generally; `llm_call_log.temperature` now populated; remaining workflows with root-level max_tokens

### DBG-004 — Timeout chain ordering contract (claim > call_handler > handler workflow)
- **status:** deployed
- **status-evidence:** 016 §7 states current values (claim 30m > call_handler 20m > handler workflow) and both mis-ordering failure modes; an earlier archived pair of drafts states the same requirement with call_handler bumped 900s→1200s between drafts, not independently re-verified against the current consolidated guide.
- **what:** Three nested timeouts must strictly decrease in scope — claim timeout (30m) must exceed dispatch's call_handler timeout (20m) must exceed the handler's own workflow timeout — or one of two failure modes occurs: a claim reset mid-work causing duplicate handling, or dispatch marking an item failed (orphaned completion) while the handler is still working with nothing listening for the eventual response. Idle monitor has a 3600s fallback; K8s ActiveDeadlineSeconds is a 24h ceiling.
- **sources:** 016 §7 (numbered-core unit); `old/older1/016_debugging_guide.md#"7. Timeout Chain"`; `old/older1/016_debugging_guide_v2_april26.md#"7. Timeout Chain"` (archives unit)
- **relations:** claim-lease-too-short reproducible timeouts (v2_49 sub-case b); claimed-item-timeout evidence gate (DBG-005); parent-timeout vs child-HITL race (DBG-063, analogous nesting requirement)
- **verify-later:** current values across dispatch/handlers vs the consolidated guide

### DBG-005 — Claimed-item-timeout evidence-based auto-completion (false-positive family, v1→v2 fix)
- **status:** partial
- **status-evidence:** v2 gate (`migration_claimed_item_timeout_evidence_v2`) live since 2026-06-04, described in one incident as "working as intended" (correctly resets a 0-component claim to failed rather than false-completing it); but two confirmed dated production false positives bracket it — gaswholesalers (~47 min early auto-complete) and gamesdesign homepage (`build_status='deployed'`+stamped, zero `page_components`, no committed file, root-caused to this exact mechanism).
- **what:** The scheduled task recovering stuck claimed work items distinguishes "work actually finished but the response was lost" from "handler died" using per-item-type evidence: `needs_content_page` requires `page_components` rows for that page updated after the claim (not the untrustworthy `build_status='deployed'` flag); `page_rerender` requires `page.deployed_at` after claim; `needs_design` keeps a caveated site-level check; `needs_rerender` is deliberately excluded (site-level, retry cheap). The original (v1) evidence was far looser — any page on the site updated since claim, using `updated_at` not `deployed_at` — producing confirmed false-completions including a homepage marked "Auto-completed: work verified done despite lost response" with zero rendered components. v2 tightens this to `p.id=wi.page_id AND deployed_at>claimed_at`, but the fix's presence/scope in the live `pre_query` has not been independently reconfirmed since.
- **sources:** `docs/agent_docs/sql_for_tables/018_site_work_items.sql#migration_claimed_item_timeout_evidence_v2` (sql_tables_components unit); 016 §9 claimed-item-timeout entries (numbered-core unit); `idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md` + `idea.uk/016_debugging_guide_v2_32(1).md` (idea_uk unit); `running_notes_14(20)`/`CATALOGUE_gamesdesign_post_sync_fix_defects` (adoption_content_quality archive unit — homepage "A4" incident)
- **relations:** silent-completion family (DBG-007); timeout chain (DBG-004); `build_status` flag distrust; work queue lifecycle
- **verify-later:** live `pre_query` text of claimed-item-timeout vs the v2 migration; debugging guide §9 current entry

### DBG-006 — jsonb && operator class bug (silent CSS-snippet failure vs hard JS failure)
- **status:** deployed
- **status-evidence:** 016 §9: css path "silently failing the entire time"; JS analog fixed in the same change set (May 2026).
- **what:** Postgres has no `jsonb&&jsonb` overlap operator; `applies_to && $1::jsonb` errored on every call, silently swallowed by a `logger.Warn`-and-return-`""` handler, so `css_snippets` never reached any deployed `styles.css` for months. Fixed with `EXISTS` + `jsonb_array_elements_text`. Wider lesson: silent-failure loaders paired with graceful consumers hide months-old breakage — prefer hard failure when the data is supposed to be there.
- **sources:** 016 §9 jsonb && entry (numbered-core unit)
- **relations:** best-effort-needs-monitoring; audit grep pattern for other `&&` uses
- **verify-later:** `loadComponentCSSSnippets` fixed in place

### DBG-007 — Silent-completion / "trust the artefact, not the status" anti-pattern family
- **status:** deployed
- **status-evidence:** Recurring across nearly every unit; 016b names it the first durable invariant; a leopardess-unit incident found seven work items all reporting 'complete' with "a success status masked a no-op for two weeks."
- **what:** The platform's single most recurring failure shape: a work item, commit, or orchestration reports success while the underlying work didn't happen. Named mechanisms include: a result-contract stub silently replacing real child output (fixed 2026-06-18); a content-regression guard's error masked by an `error_step` literally named `complete_error`; a pod dying mid-flight leaving a `complete` status with a non-empty `error`; "git committed the file" re-committing stale already-stored components; zero-planned-sections completing as success. The companion doctrine: verify against `page_components` timestamps and live HTML/artefact, never against status columns; `completed_at` is the orchestration END, not the write instant (trace child orchestrations by `page_id` in `collected_data`); intermediate signals (work-item names, pod snapshots, mid-flight tables) lie.
- **sources:** 016/016b recurring entries + traps 1–3 (numbered-core, travelling_docs units); `docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-index(4).md` + `HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md` (leopardess_social unit, "ten empty builds")
- **relations:** zero-planned-sections (DBG-013); `save_page_sections` content-regression guard (DBG-008); claimed-item-timeout (DBG-005); gamesdesign result-stub (DBG-050); loose dispatch item-status semantics (DBG-029)
- **verify-later:** framework preventions specified but only partially shipped — planner invariant (every planned page ≥1 section), fail-loud zero-sections path, auditor rules for planned-but-linked pages, post-deploy URL presence checks

### DBG-008 — save_page_sections: sole page_components writer, HTML-fallback bug, and content-regression guard laundered to false success
- **status:** partial
- **status-evidence:** The tool-pages HTML-fallback bug was fixed and shipped 2026-05-28 (end-to-end verification "honest-open"); the guard-masking fix is flagged but not confirmed applied; a separate later investigation walked through four falsified theories (dispatcher starvation, concurrent-deploy pod-cycling, claim-lease timeout, caller/callee timeout mismatch) before confirming the guard-laundering mechanism, then course-corrected to flag a second, still-unconfirmed candidate mechanism (`page_components` row-locking via an `auto_lock_on_deploy` trigger).
- **what:** `save_page_sections` is the sole writer of `page_components` (DELETE+INSERT, with a history row written but `source_item_id` left NULL on the overwrite path). Its HTML fallback originally extracted only bare `<section>` blocks, so `<div class="tool-page">` tool/game-page HTML was silently discarded (n_rendered=0, no file ever committed) — fixed with a whole-fragment-as-one-section fallback guarded against full documents. Separately, its content-regression guard (refuses to overwrite existing deployed content with much-shorter new content) is a real, correct safety check, but its error return was silently laundered by `complete_error` — itself a SUCCESS-labelled `complete_workflow` step — so a legitimate guard refusal reads as a clean success while the committed file stays stale. Two distinct real bugs follow regardless of which mechanism actually fires on a given stuck page: (1) the guard's refusal shouldn't route through `complete_error`; (2) deploy shouldn't proceed (re-render + git-commit) after a zero-row save.
- **sources:** 016 §9 tool-pages-never-deploy entry + guard entry (numbered-core unit); `content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(16).md` (adoption_content_quality archive unit)
- **relations:** silent-completion family (DBG-007); deployed→needs_rebuild flip; `auto_lock_on_deploy` trigger (unconfirmed second mechanism)
- **verify-later:** patched `save_page_sections` deployed to all three callers; which mechanism (regression guard vs component lock) actually fires; `auto_lock_on_deploy()` trigger function body; `page_build_handler_save_failure_visible.sql` application state

### DBG-009 — Presign/awaited-loop O(K²) state bloat, fixed by batch adapter calls
- **status:** deployed
- **status-evidence:** Marked DONE + CONFIRMED IN PROD 2026-06-09 (migration 110); full launcher path completed in ~26s post-fix vs a retired loop still at iter_9 nine minutes in pre-fix.
- **what:** Every step transition in an awaited loop re-persists the ENTIRE orchestration state (expanded workflow + collected_data + processing history), so a K-iteration awaited loop over cheap per-item adapter calls costs O(K²) and geometrically slows (iter_0-4 ~2-3s, iter_8 ~100s, then Kafka I/O timeouts) while a GPU bills throughout. The structural fix, not tuning: replace the per-item loop with one batch adapter call (`prepare_object_urls` returning all URLs in one reply) — one await, one persist, no flatten step. General platform lesson: awaited loops over cheap local operations are an anti-pattern; batch at the adapter. A related fix landed alongside: `configOrInput` now coerces numeric config scalars (`expiry_minutes` 3000 was silently dropped by a `.(string)` type assertion — see DBG-026).
- **sources:** 016 §9 presign entries (numbered-core unit); `NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-(3)`; `110_training_launcher_batch_presign(2).sql` (finetuning unit)
- **relations:** send-before-register await race (DBG-023); `loop_complete` convention; configOrInput coercion (DBG-026)
- **verify-later:** orchestration state persistence cost in coordinator generally; whether other awaited loops exist at risk; training-launcher def shape

### DBG-010 — Hand-applied agent-def/launcher migrations have no ledger; re-running an earlier one reverts later ones
- **status:** deployed
- **status-evidence:** Directly observed and reversed 2026-06-09 ("re-running 109 silently REVERTED both [110 and 111]"); recovery required re-applying 110 then 111; "RUNBOOK 2d state check" added as a live procedural safeguard across at least three units.
- **what:** `agent_definitions`/launcher-def jsonb migrations are hand-applied with no `schema_migrations`-style ledger or runner — the live definition's SHAPE is the only "did it run" truth. A migration is idempotent only against its OWN prior application, never against LATER migrations that mutated the same object/subtree, so "re-running an earlier one to make sure" silently reverts everything applied after it. Standing mitigation: a per-migration state-check query (RUNBOOK step "2d") after every deploy and before any launch, and using the sanctioned `snapshot_agent()`/`revert_agent()` helpers (not hand-rolled CREATE TABLE backups, which collide with the existing `agent_definitions_backup`) for recovery. This exact incident is the direct motivating case for the NEW:migration-governance proposal (see register/migration-governance.md MIGG-001).
- **sources:** 016 §9 re-running-idempotent-migration entry (numbered-core unit); `NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-(7),(3)` (finetuning unit); `archive_april_26/016_debugging_guide_v2_47(1).md#"§9 Re-running an idempotent migration"` (archives unit, recovered into live v2_58 after being dropped from the v2_48–v2_57 branch)
- **relations:** register/migration-governance.md MIGG-001; wrong-binary/CrashLoop incident (same "shipped what I thought?" family, DBG-011); model swap/revert functions
- **verify-later:** RUNBOOK 2d query vs live launcher def; whether a migration runner/ledger was ever built

### DBG-011 — CrashLoop `exec: "./X": no such file` — image lacks the binary (build/packaging fault, not config)
- **status:** deployed
- **status-evidence:** Concrete instance 2026-06-14: thunder-adapter:v1.0.1063 actually shipped the analyser-adapter binary (Dockerfile overwritten, tag shared a digest) — the third deploy-regression of this exact shape in a row; pod CrashLoopBackOff'd ~31h; guide entry recovered into v2_58 after being dropped from an intermediate branch (v2_48–v2_57).
- **what:** `exec ./X: no such file or directory` at pod startup means the running image does not contain binary X — always a build/packaging fault, never a runtime config problem. Diagnosis: `docker run --rm --entrypoint ls <image> -la /app` to inspect contents, `docker inspect .Config.Entrypoint`/`.RepoDigests` to catch tag-collisions (two tags sharing one digest). Fix: restore the correct Dockerfile and push a FRESH tag — never re-push the poisoned one. The recurring structural gap named across three independent incidents is "no guard between built and running"; prescribed guards are a pre-push `docker run --entrypoint ls` check or a CI step that fails the build if the expected binary is absent.
- **sources:** 016 §9 CrashLoop entry (numbered-core unit); `working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-14-8` (finetuning unit); `archive_april_26/016_debugging_guide_v2_47(1).md#"§9 CrashLoop exec"` (archives unit)
- **relations:** hand-applied migrations lesson (DBG-010, same "shipped what I thought?" family); deploy≠migration-ran checklist item; adapter-vs-chassis deployment drift (DBG-068)
- **verify-later:** whether a CI image-content guard was ever added for thunder-adapter/analyser-adapter Dockerfiles

### DBG-012 — Open problem: nav-updater never spawns
- **status:** unknown
- **status-evidence:** 103 "Active Problem" — definition exists/active, topics exist, dispatch generic, yet no pod ever appeared; all `nav_drift` items claim-timeout; investigation was open at handoff 2026-04-12.
- **what:** `nav_drift` work items route to the nav-updater agent via the standard generic dynamic dispatch, but no nav-updater pod has ever actually started, and every `nav_drift` item exhausts its claim timeout instead. Distinct from the nav-link-fixer path (which does work). Never resolved within the material seen.
- **sources:** 103#Active Problem (numbered-core unit)
- **relations:** dispatch loop; missing-handler pattern (distinct: here the definition DOES exist)
- **verify-later:** whether resolved since; nav_drift item outcomes

### DBG-013 — Zero-planned-sections silent no-op success (planning gap + complete_error anti-pattern)
- **status:** partial
- **status-evidence:** 016b v5 + 2026-07-06 amendment confirms the route from a workflow dump and that a `pages.sections` UPDATE changes behaviour, but guards/planner-invariant fixes are listed as prevention only; a sibling incident (guide-skinner-box) shipped one layer of the fix stack (`check_sectionless_pages.go`) as "next, optional... cleanliness" at time of writing (a different unit's account later reports it enabled 2026-07-10).
- **what:** A page reaching page-build with zero planned sections (`pages.sections=[]`) is not a build failure but a planning gap — no sections were ever produced by the planner-to-pages pipeline (`site_specs` site_plan aspect → `pages.sections` is the read order; `site_plan_sections` is NOT read by builds). page-build-handler's zero-ready-sections branch is literally a `complete_workflow` step named `complete_error` — "an error path implemented as a successful completion," diagnostically signalled by a result containing only `site_record`. A page can 404 or stay stale for weeks while several work items complete clean. Fix stack (partially applied): a read-time fallback copying a same-role sibling's section layout when both known section sources are empty; a discovery check (`check_sectionless_pages.go`) that detects and retriggers stuck sectionless pages; a workflow-level fix routing the genuinely-unrecoverable case to a flagged (non-complete) state; planner invariant that every planned page have ≥1 section; rerender warn; auditor rules; dynamic-list component vocabulary for archive pages.
- **sources:** 016b#Page-build-completes-having-built-nothing + amendment (numbered-core/travelling_docs units); `adoption/running_notes_15_skinner_box_and_adoption_sections(5).md` Parts 1–8 (adoption_content_quality archive unit)
- **relations:** silent-completion family (DBG-007); content↔template key-contract drift (DBG-014); dormant `checkEmptyPageSections` discovery-check machinery
- **verify-later:** `complete_error` branch fixed?; planner invariant added?; `check_sectionless_pages.go` enablement state; whether the workflow-level flagged-state fix ever shipped

### DBG-014 — Content↔template key-contract drift (system-stats class)
- **status:** partial
- **status-evidence:** TRIAGED 2026-06-24, remedy un-applied at that point ("full content rebuild... then re-check"); fleet-wide (usage_count 22); the component itself was later regenerated to q100 in a separate (vonc) arc.
- **what:** A populated-but-visually-blank section is a content↔template KEY-CONTRACT problem, not a generation failure: component-creator rewrote a template mid-flight (e.g. `eyebrow_label`/`section_headline`/`stat1_value`...) while stored `content_data` kept the OLD key names (`eyebrow`/`heading`/`stat_1_number`...) sharing ZERO keys with the new placeholders, so every placeholder rendered empty and the (correct) visible-content filter dropped the band entirely. Diagnostic heuristic: diff the two key sets directly rather than treating it as a generation bug. Structural need: a component schema change should trigger dependent content rebuilds (`page_rerender` reuses the mis-keyed `content_data` and does NOT fix it — only a full content rebuild does).
- **sources:** 016b Part 5 + wrong-turn #4 (numbered-core unit); `docs/016b_debugging_guide_merged(3).md#open-threads` (vonc unit)
- **relations:** schema-template-drift tag; zero-planned-sections (DBG-013); shared library field guard (later prevention)
- **verify-later:** whether schema-change→dependent-rebuild triggering exists (`markPagesPendingRebuild` covers regen only; mid-build rewrites unclear)

### DBG-015 — agent_error_log / llm_call_log / http_request_log as the primary forensic sources (read before pod stdout)
- **status:** deployed
- **status-evidence:** Repeatedly promoted across units as the "first read"; a 2026-07-08 incident settled a tool-generation blocker in one query via `agent_error_log`; schema confirmed via `docs/agent_docs/sql_for_tables/022_agent_error_log.sql`.
- **what:** Persistent DB logs outlive ephemeral pod stdout and should be read first: `agent_error_log` (`orchestration_id` [TEXT, not uuid], step_name, action, error_message, error_code, severity, JSONB context, resolution tracking) captures every reported error and survives pod rotation/reap; `llm_call_log` records every LLM call including failures with `error_message`; `http_request_log` (mirrors the `llm_call_log` pattern) records every outbound HTTP call (caller identity, method/url/domain, status/bytes/latency) for operational visibility and per-domain rate-limit tracking. Companion cautions: `current_step` from polling is a sample not an attribution (a 120s poll blamed the wrong step); a terminal step's `success_message` can name the wrong phase; pod logs vanish on rotation/rollout and zap JSON must be grepped by message string not `field=value` (`logger.Debug` is invisible in production; house rule is `logger.Info`).
- **sources:** 016#hunting-for-logs + 016b#Verifying-a-deploy (numbered-core unit); `016b_debugging_guide_7_3_(7).md#agent-error-log-entry` (travelling_docs unit); `docs/agent_docs/sql_for_tables/022_agent_error_log.sql`, `#026_http_request_log.sql` (sql_tables_components unit)
- **relations:** pod-rotation log loss (assumption checklist item); silent-completion family (DBG-007); database cleanup retention policy
- **verify-later:** `agent_error_log` schema (`orchestration_id` type); HTTP client wrapper writing rows; writers in chassis error paths generally

### DBG-016 — SQL/DB template-and-data surgery method: needle-gate discipline + Postgres pitfall catalogue
- **status:** deployed
- **status-evidence:** Applied consistently across every production template/prompt/config mutation seen (W1–W3e, slice4a/4b, the 019 migrations, the vonc marker/ghost-row fixes); pitfalls repeatedly promoted into 016b guide versions with the refinement "count expectations mechanically from the dump, never from memory."
- **what:** The dominant safe method for mutating production data stored as text/jsonb (templates, layouts, prompts, docs): dump the current bytes and take a shell/CREATE-TABLE-AS backup first; run a read-only "needle-gate" that asserts every expected needle as a LIKE boolean PLUS a mechanically-derived occurrence count (never recalled from memory — a mismatch means real drift OR a bad expectation, stop); apply a guarded, idempotent UPDATE (exact-string nested `replace()` or anchored/backreferenced `regexp_replace`, gated on a pre-state marker so re-runs are 0-row no-ops); check `RETURNING` booleans as immediate post-conditions; ship a separate verify file and a value-agnostic rollback file. Catalogued Postgres pitfalls baked into the method: bounded regex quantifiers `{m,n}` cap at 255 (use `strpos`/`substr` instead); `substring(... from '(pattern)')` returns only the FIRST capture group; LIKE treats a needle's literal `%` as a wildcard (use `position()`); naive `background:\s*#`-style regexes miss gradient-embedded hex colours; `replace()` is literal-byte and silently no-ops on a missed anchor while still reporting UPDATE 1; CASE doesn't short-circuit sub-SELECTs (use DO blocks + RAISE); NAMEDATALEN 63 truncates long backup-table names; Go text/template rendering of stored prompts needs `funcMap` helpers for literal `{{…}}`; an aborted transaction is sticky and ignores everything but ROLLBACK — migration files should open with a defensive ROLLBACK and run via `psql -f`/`\i`, never pasted into interactive psql; `\set ON_ERROR_STOP on` for dependent statements; a 0-rows result from one's own verification query is not decisive either, until cleared.
- **sources:** 016b#SQL-verification-pitfalls (numbered-core unit); `running_notes_scheme_to_components(55).md#Sr/Ss/St/Sv/Sw` (idea_uk_section_data unit); `RUNBOOK_scheme_to_components(18).md` W1–W3e (content_quality2 unit); `016b_debugging_guide_7_3_(7).md#postgres-regex-entry` (travelling_docs unit); `FOCUS_directory_builder_and_list_components.md#gotchas` (adoption unit); `docs/016b_debugging_guide_merged(3).md#sql-verification-pitfalls` (vonc unit)
- **relations:** register/sql-change-management.md SQLC-001 (same method, extracted independently under the proposed new category); marker/attribute REPLACE anchoring (DBG-071); migration governance (DBG-010, register/migration-governance.md MIGG-001)
- **verify-later:** n/a — practice; instances cited are the verification

### DBG-017 — sites.status is informational; never scope blast-radius by status='active'
- **status:** deployed
- **status-evidence:** 016b v4 entry with a silently-dropped-site incident as the motivating case.
- **what:** `UpdateSiteStatusAction`'s real vocabulary is draft/building/review/published/deployed/archived/error; 'active' is a legacy hand-written value that nothing in dispatch actually filters on (dispatch keys on `site_work_items` instead). The corollary discipline: always enumerate `GROUP BY status` before writing any blast-radius query, and check `pg_proc`/`pg_trigger` before adding helper functions (a shared `set_updated_at` trigger already exists).
- **sources:** 016b#sites.status (numbered-core unit)
- **relations:** assumed-status-values trap (DBG-051, the same lesson on `pages.status`)
- **verify-later:** —

### DBG-018 — Kafka trigger payload discipline: multi-line kcat bodies silently mis-route
- **status:** deployed
- **status-evidence:** Documented as a permanent ops pattern in 016_debugging_guide_v2 §9 (2026-04-23); independently re-struck in a later incident (run 464102f4, rev 45) where a pretty-printed body piped through `kcat -P` became one message per line and the chassis married the wrong headers to a neighbouring message's body.
- **what:** `kcat -P` is strictly line-delimited: a multi-line (pretty-printed or heredoc) JSON trigger body becomes multiple separate messages, and the chassis can pair your intended headers with a NEIGHBOURING message's body — observed as a correlation id "completing after 0 steps" holding a scheduler no-op's body. The fix is universal: compact all Kafka trigger bodies to a single line (`<<<'{...flat json...}'` here-strings or `jq -nc`), and scripts should refuse/detect multi-line bodies. A related manual-trigger pattern: build the payload with `psql jsonb_build_object` and pipe straight to kcat with standard headers, used to trigger handlers directly when dispatch itself is blocked.
- **sources:** `FOCUS_finetuning_flywheel_and_service(13).md#2.4f`, `FOCUS_dispatch_diagnostic(4).md#Workarounds` (docs024_focus_handoff unit); `RUNNING_NOTES_travelling_docs(39).md#rev45-run1`, `RUNBOOK_travelling_docs(38).md#new-durable-rules` (travelling_docs unit)
- **relations:** env-prefix trap (DBG-036, sibling shell/trigger-script gotcha); CLI/ops data-transfer pitfalls (DBG-025)
- **verify-later:** the stale-buffer wrinkle in the chassis consumer flagged but never followed up

### DBG-019 — Discovery-checks list maintenance and the workflow-replace landmine
- **status:** deployed
- **status-evidence:** "Closed — investigation found no overwriter" (2026-04-19); the safe jsonb `||` append pattern recommended; the latent risk (`updateAgentWorkflow`) explicitly logged as "currently safe because nothing fires it."
- **what:** A suspected "checks keep falling off discovery agents" turned out to be a red herring — a stale in-code example being manually copy-pasted as a full-array SQL replace, not an overwriter process. The safe pattern going forward is a jsonb array append, not a whole-array replace. A latent structural risk was logged for the future: `updateAgentWorkflow` performs a `jsonb_set` of the ENTIRE workflow subtree, so once an automated improvement-proposal generator ships, a partial proposal will silently erase the rest of a workflow unless the write path is converted to a deep-merge.
- **sources:** `HANDOFF_2026-04-19_component_linking_news_template_discovery_checks.md#3,#4` (docs024_focus_handoff unit)
- **relations:** `improvement_proposals` (empty table); `ApproveImprovementAction`; SQL surgery method (DBG-016)
- **verify-later:** `updateAgentWorkflow` (context line ~61056); stale comment cleanup

### DBG-020 — Deployed-binary-predates-disk failure class
- **status:** deployed
- **status-evidence:** "Fork RESOLVED: extraction sound … the on-disk code cannot produce the observed escalation → deployed predates disk; the skip_field fix exists and never shipped."
- **what:** A named diagnosis class where observed runtime behaviour contradicts a correct reading of the on-disk code because the running pod's image predates the working copy — the fix exists in the repo and simply never shipped. Diagnostic: compare `git log -1 -- <file>` against the running pod's image build/age. Sibling lessons from the same investigation: verify the running image actually contains an edit before debugging it further ("success path silent" reading trap), and prefer a forward test (one clean build, then read both the render and the stored data) over forensic reconstruction across overlapping rebuild windows.
- **sources:** `running_notes_scheme_to_components(55).md#Tl/#Tm/#Tt`; `RUNBOOK_scheme_to_components(50).md#W7-FINDINGS` (idea_uk_section_data unit)
- **relations:** chassis deploy model (build from filesystem, decoupled from commits); plan_sections deferral instance
- **verify-later:** 016b guide entry for this class

### DBG-021 — LLM API shape disciplines (server-tool injection, per-model thinking shapes, long-call timeouts)
- **status:** deployed
- **status-evidence:** Items 24–27 added to the 016 debugging guide from a 2026-06-04 validation run described as "three API bugs found and fixed during validation."
- **what:** Standing disciplines earned from live Anthropic API breakage: a hosted server tool may auto-inject its own documented dependency, so declaring it yourself collides (web_search v2 injects code_execution; the 400 error names the conflict); the same capability has different wire formats across model generations (newer Opus-class models take adaptive thinking + `output_config.effort`, Sonnet 4.6 takes a manual `budget_tokens`, and Opus additionally rejects non-default temperature/top_p) so helper code must branch per model; long agentic calls (high effort + many tool round-trips) can send no HTTP headers for minutes, so client timeouts must be sized for the worst-case step (180s→900s, with streaming as the durable long-term answer); and the request shape should always be confirmed from live docs before coding, especially right after a model bump — a remembered shape is a guess, and every failed round-trip costs real spend.
- **sources:** `idea.uk/016_debugging_guide_v2_32(1).md` (items 24–27); `idea.uk/DEVELOPMENT_RUNBOOK(3).md#A1`; `idea.uk/running_notes(63).md` (idea_uk unit)
- **relations:** engine upgrade; model-infrastructure; LLM step config shadowing (DBG-003, different bug, same model-generation API surface domain)
- **verify-later:** `engine.go` `usesAdaptiveThinking` + client timeout values

### DBG-022 — Operator/assistant division-of-labour and DB-change safety conventions
- **status:** deployed
- **status-evidence:** Standing-rules blocks repeated verbatim across at least three handoff documents (HANDOFF_2026-06-09, HANDOFF_2026-06-15 §0, HANDOFF_page_pipeline §0-1).
- **what:** The working covenant under which these sessions ran: the assistant reads code and writes deliverables with no direct cluster/DB access; a human operator runs all SQL/kubectl/builds. Safety conventions layered on top: snapshot before any DB change (`snapshot_agent()`/`revert_agent()` house helpers for agent rows; short-named, in-transaction `CREATE TABLE <t>_bak_<tag> AS SELECT` for data, with rollback documented); a fresh `\d` before any SQL; every template `replace()` verified by an UPDATE-count check plus a flag flip (a stuck flag is often a whitespace mismatch silently no-op'd); `check_linking_sql_applied.sql` as an idempotent "which SQLs are already in" orientation step; workflow (jsonb) changes are DB-only and immediate, Go changes need an image roll + `image_tag` bump; tags of co-deployed agents must move together or a lagging resolver tag becomes a permanent silent fallback; never roll the chassis image while a rebuild batch is draining.
- **sources:** `HANDOFF_2026-06-15(2).md#0`; `HANDOFF_page_pipeline(11).md#0-1`; `check_linking_sql_applied.sql`; `RUNBOOK_linking_phantom_fixes(7).md` (content_quality_linking unit)
- **relations:** SQL surgery method (DBG-016); guide-meta doctrine (DBG-001)
- **verify-later:** `snapshot_agent`/`revert_agent` function definitions; `agent_backups` table behaviour

### DBG-023 — Send-before-register await race (preRegisterAwaitedRequest fix)
- **status:** deployed
- **status-evidence:** Fix confirmed live 2026-06-09 ("Race fix works [verified-log]... claimed:true"); logged as the fourth cause of stuck-'waiting' awaited requests in the 016 debugging guide §9.
- **what:** A local dispatch action (`dispatch_thunder_prepare_object_url` and similarly-shaped clones) produced its adapter request and returned `await_response:true` BEFORE the coordinator had inserted the corresponding `awaited_requests` row; a fast (~1s) adapter reply could beat the insert, so `ClaimAwaitedRequest` (WHERE status='waiting') found nothing, the reply was silently dropped, and the timeout handler re-dispatched the same request forever with fresh request_ids (RetryVersion pinned at 0). `spawn_agent`/`call_agent` never race this way because they call `preRegisterAwaitedRequest` first (register-before-send, `ON CONFLICT DO NOTHING`). Fix: call the same helper in the affected dispatch actions before `ProduceWithValidation`, guarded on `params.DB != nil`; caveat — the helper hardcodes a 120s timeout that overrides step config, and the per-request timeout goroutine is skipped (a background expiry sweep is the net that catches it). "Moving stall point" (the symptom relocates rather than disappearing) is the diagnostic tell for this race class.
- **sources:** `working/phase5/HANDOFF_2026-06-06_checkpoint_upload_loop_await_race(2).md`; `NOTES_phase5_training_launcher_running(45).md#update-2026-06-06,#2026-06-08,#2026-06-09` (finetuning unit); `docubundle/.../HANDOFF_2026-06-06_checkpoint_upload_loop_await_race.md`; `CONTEXT_PACK_thunder_checkpoint_race.md` (archive_finetuning unit)
- **relations:** O(K²) loop bloat (DBG-009, found immediately after this fix); `awaited_requests` machinery; reply-topic own-vs-parent derivation (DBG-069, same subsystem)
- **verify-later:** `preRegisterAwaitedRequest` call sites across `thunder_prepare_object_url_dispatch.go` and any batch/resume dispatches

### DBG-024 — agent_definitions source-of-truth is clients_db, not templates_db, for the rich (versioned) schema
- **status:** deployed
- **status-evidence:** Corrected same-day 2026-06-03: "templates_db.agent_definitions has the OLD schema… holds only the 8 original website-builder agents… PIN (corrected): for the flywheel-C agent_definitions, always read AND patch clients_db" — a pin that was itself first applied to the wrong DB, then reversed within the same day.
- **what:** `agent_definitions` exists physically in BOTH `clients_db` and `templates_db`. The architecture doc's "source of truth is templates_db" statement refers only to the legacy website-builder catalogue (old schema, no version column). The chassis definition-loader (filters `is_active`/`is_snapshot`, `ORDER BY version`) can only run against clients_db's rich schema, so all modern/flywheel-C definitions live there — meaning the clients_db copy of a definition can silently diverge from what a stale architecture doc claims, and doc claims about "the" source of truth need runtime confirmation, not just a document read.
- **sources:** `working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-03`; `103_call_data_preparer_optional_inputs.sql`, `104_provisioner_output_fields_and_launcher_mapping.sql` (finetuning unit)
- **relations:** agent-is-a-DB-row (DBG-002); hand-applied migrations (DBG-010); documentation-system stale doc line (`002_system_architecture.md`)
- **verify-later:** chassis definition-loader query; `002_system_architecture.md` wording

### DBG-025 — CLI/ops data-transfer pitfalls (COPY-vs-psql jsonb escaping, kubectl exec/cp truncation, tnr scp nesting)
- **status:** deployed
- **status-evidence:** Both bugs validated + corrected command shipped 2026-05-06; independently observed truncation cases (1716/1958 rows with "next reader: unexpected EOF").
- **what:** A cluster of verified transfer traps beyond the kcat heredoc issue (DBG-018): `COPY … TO STDOUT` is not JSON-safe for jsonb columns (double escape layers) — use `psql -tAXc` with a plain SELECT for JSONL instead; `kubectl exec -i` without fully-consumed stdin sporadically truncates stdout; `kubectl cp` truncates large files silently (use `exec cat > local` instead); `tnr scp` of directories nests `{dest}/{source_basename}/` on both source and destination.
- **sources:** `working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted_PATCH_2026-05-06.md`; `FOCUS_finetuning_flywheel_and_service(25).md#14`; `HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons(4)` (finetuning unit)
- **relations:** Kafka trigger payload discipline (DBG-018); dataset pull path
- **verify-later:** `01_pull_dataset_from_postgres.sh` uses the corrected form

### DBG-026 — configOrInput numeric config coercion (expiry_minutes silently dropped by a .(string) assertion)
- **status:** deployed
- **status-evidence:** Fixed and verified 2026-06-09: "expiry_minutes override — FIXED… configOrInput read config via Config[name].(string), so the JSON-number 3000 failed the assertion → fell through → adapter default." Logged in debug guide v2_43.
- **what:** The shared `configOrInput` helper type-asserted every config value to string, so JSON-number config (`expiry_minutes:3000`, `timeout_seconds`, etc.) silently failed the assertion and fell through to hardcoded adapter defaults — presigned upload URLs came back valid for 24h instead of the intended 50h, with no error anywhere. Fixed with a `coerceConfigScalar` helper handling string/float64/json.Number/int/bool. Class lesson: any shared config-reading helper must coerce scalar types, and a numeric setting is only proven "applied" by observing its actual effect (e.g. reading `X-Amz-Expires` off the generated URL), never by trusting the config value alone.
- **sources:** `working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-(4),(5)` (finetuning unit)
- **relations:** O(K²) loop replacement (DBG-009, same change window); launcher dispatch family
- **verify-later:** `coerceConfigScalar` in `thunder_ssh_exec_dispatch.go`

### DBG-027 — Scheduler-fired chassis-resident agents report owner_agent_type='generic' (observability gotcha)
- **status:** deployed
- **status-evidence:** STATUS 2026-05-12 architectural follow-ups: "filtering orchestration_states by owner_agent_type MISSES top-level chassis-resident workflows, which are owned by 'generic'."
- **what:** Scheduler-fired agents that run inside the generic chassis process (thunder-reaper, build-pipeline-trigger, etc.) have `orchestration_states.owner_agent_type='generic'` rather than their semantic type; the real agent type lives at `collected_data->'config'->>'agent_type'`, and `orchestration_name` follows a `sched-<task>-<ts>` convention. Any monitoring/filtering query must key on those fields instead of `owner_agent_type`. A related unresolved cosmetic anomaly: a stale non-DB `agent_config` stub (an old reaper-style no-op) persists in message envelopes across redeploys even while the full WorkflowPlan correctly executes — the source of the cached representation was never found.
- **sources:** `working/flywheel_docs/STATUS_thunder_adapter_2026-05-12(1).md#6`; `NOTES_phase5_training_launcher_running(45).md#stub-source-narrowed` (finetuning unit)
- **relations:** monitor testing; guide meta (DBG-001)
- **verify-later:** where the stale `agent_config` envelope field loads from

### DBG-028 — Kafka topic-creation race self-heal (transient "Topic not yet on broker")
- **status:** deployed
- **status-evidence:** 2026-06-06: "Transient Topic not yet on broker for the launcher .responses topic self-healed on attempt 2 (topic-creation race) — normal."
- **what:** Per-spawn child Kafka topics (`job.<id>.requests`, per-agent responses topics) are created on demand, and a first-publish race against broker propagation produces a transient failure that a retry resolves on its own — this should not be chased as a real fault. The distinguishing signature versus a genuinely missing topic (e.g. Strimzi auto-create disabled, which fails every attempt) is exactly this self-heal-on-retry behaviour.
- **sources:** `working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-06`; `FOCUS_adapter_design(3).md#required-cluster-resources` (finetuning unit)
- **relations:** adapter deployment essentials (KafkaTopic CRDs); Kafka per-spawn partition race (DBG-045, a different, non-self-healing Kafka defect)
- **verify-later:** topic auto-creation settings for spawned-agent topics vs adapter topics

### DBG-029 — Loose dispatch item-status semantics (complete ≠ done)
- **status:** aspirational
- **status-evidence:** "worth a pass when convenient" (RUNBOOK Part E Hygiene); seven dated sightings logged in running notes.
- **what:** A documented defect class in the dispatch loop's work-item bookkeeping, observed seven separate times: items marked 'complete' at dispatch time while the child orchestration is still running or later fails; the child's full error text landing in the `error` column of an item whose status still reads 'complete'; status transitions that don't bump `updated_at`; a shared batch-claim timestamp across items with different eventual fates; a parent's fire-and-forget topic lifecycle polluting child completions with "topic partition not found" noise. Operational rule derived: never trust item status as proof of work — verify the artefact, and treat `agent_error_log` (by `occurred_at`) as outranking status. The fix itself is parked as hygiene, not applied.
- **sources:** `NOTES(43).md §9i,§9l,§9m,§9aa,§9ac,§9ax,§9bd`; `RUNBOOK(49).md Step 9 + Part E` (content_quality2 unit)
- **relations:** silent-completion family (DBG-007); F2 methodology (DBG-030, discriminator ordering); auto-escalation
- **verify-later:** build-dispatch-loop status handling; whether items get failure statuses on child errors

### DBG-030 — F2 tiered guard-verification methodology (unit → integration → live keep/reject fixtures)
- **status:** deployed
- **status-evidence:** "F2 COMPLETE: Tier 1 unit ✔, 3a preservation ✔ (×2 regens, md5-verified template change), 3b reject ✔ (live firing, three-level visibility, zero mutation)."
- **what:** A verification pattern for proving a guard/fix works without touching live shared components: Tier 1 deterministic unit tests of the guard logic (including the real incident's own rename case); Tier 2 a DB-backed reject-path test (folded into Tier 3 where no harness existed); Tier 3 end-to-end tests on throwaway `zzz-*` components — a KEEP fixture proving preservation-by-instruction (template md5 changes while tracked fields hold) and an intentionally INACTIVE REJECT fixture that exploits the store-vs-loader `is_active` divergence to force a rename and observe the guard fire live with zero mutation. Also codifies a discriminator ordering for evidence quality: `agent_error_log` > pod logs > work-item status, never the reverse.
- **sources:** `NOTES(43).md §9f,§9h,§9k-§9o`; `RUNBOOK(30)` family Step F2 tiers (content_quality2 unit)
- **relations:** loose dispatch item-status semantics (DBG-029); F1 guard; F4
- **verify-later:** `store_generated_component_guard_test.go`; zzz fixtures fully cleaned up

### DBG-031 — R6c artifact-forensics method: cache-busted, metric-consistent comparisons
- **status:** deployed
- **status-evidence:** "md5sum: gd.html == gd2.html… ONE artifact all along. OWNED: my stale-cache story AND the earlier '4-of-8 mis-assembled' reading were metric artifacts."
- **what:** Lessons drawn from a "blank page" false trail: only compare live artifacts using identical metrics (a data-component inventory vs a class-based grep counted different things and manufactured a false mis-assembly story); md5 the actual fetched bytes before concluding a page is stale-cached; distinguish 404 / 200-empty / 200-styled-but-invisible with curl size + response headers rather than eyeballing; visually-blank does not mean content-missing (a fallback-CSS-variables case had content present but dark-on-dark). The eventual real cause (a theming issue, not assembly or deploy) reshaped the whole investigation thread.
- **sources:** `NOTES(43).md §9af-§9al`; `RUNBOOK(49).md Part B` (content_quality2 unit)
- **relations:** assembly membership model; needle-gate pattern (same mechanical-counting ethos, DBG-016)
- **verify-later:** n/a — method; instances cited

### DBG-032 — Code-ahead-of-DB schema drift (SQLSTATE 42703, latent until first caller)
- **status:** deployed
- **status-evidence:** Root-caused 2026-07-08 (`create_tool_component` referencing missing `content_components` provenance columns, latent ~2 months since 2026-05-16); fix applied and proven 2026-07-09.
- **what:** A binary can reference new columns that were deployed before their migration ran, and nothing fails until the rare code path that touches them is actually called. Detection: the failing INSERT's own error text names the missing column/migration; comparing last-successful-call latency distinguishes a long-latent drift from a fresh regression. Fix pattern: mirror the column types dynamically from the table the code says it copies from (`format_type`/`pg_attribute` + `ADD COLUMN IF NOT EXISTS`), additive/nullable/idempotent. The specific root cause here was a migration file that existed but sat parked in a docs folder, never renumbered into the actual migrations path — exactly the mechanism by which a deploy can skip a migration. Standing pre-deploy check: grep the diff for new column names and assert each exists in production before shipping the binary.
- **sources:** `016b_debugging_guide_7_3_(7).md#schema-drift-entry`; `HANDOFF_2026-07-08…md#3`; `RUNNING_NOTES_travelling_docs(39).md#rev29,#rev30` (travelling_docs unit)
- **relations:** migrations system; migration governance (DBG-010, register/migration-governance.md MIGG-001); `content_components` provenance columns (migration 133)
- **verify-later:** `sql_for_agents/133_add_component_provenance.sql` vs the docs019 design copy

### DBG-033 — Prompt-template rendering resolvers differ by output_format (text vs json vs action-config)
- **status:** deployed
- **status-evidence:** Root cause of a missing first auto-PLAN (016b rev 32); the same class of bug was independently caught in Task-4 templates before it fired; re-proven as "THE RULE (proven by this run, not assumed)" in a separate incident (migration 134, 2026-07-09).
- **what:** `execute_llm_prompt` hands its downstream prompt template a BARE string when `output_format:text` (`{{.X}}`), but a MAP when `output_format:json` (live form needs `{{.X.result | toJSON}}`); action CONFIG field paths are resolved by a completely different mechanism and keep the `.result` suffix. Reaching an unverified nested key directly is fragile — dump whole objects with `| toJSON` instead of guessing. A render-time template error fires before any tokens are spent, and with error containment upstream, the whole workflow can "succeed" while the specific step's product is silently missing — the reading rule is: a normal terminal status plus a missing downstream artefact means a contained step failure, not a healthy run.
- **sources:** `016b_debugging_guide_7_3_(7).md#template-entry`; `RUNNING_NOTES_travelling_docs(39).md#rev32`; `HANDOFF_2026-07-09_recreation_and_chassis_1_.md#2` (travelling_docs unit); `134_fix_prompt_template_field_paths.sql` (sql_for_agents unit)
- **relations:** docs-never-fail containment masking effect; `error_step`-inside-config rule (DBG-055); call metadata/response convention
- **verify-later:** template data shaping in `ai_actions.go` by output_format; `ExtractActionInputs` / template renderer code

### DBG-034 — EXECUTING_STEP frozen forever means the worker pod died (OOMKill triage), superseding stall/leak hypotheses
- **status:** deployed
- **status-evidence:** 016b v8 rewrite 2026-07-09; explicitly supersedes an earlier v5-era "error containment does not protect against a HANG" entry and a slow-leak hypothesis, both walked back on the evidence trail kept in the RUNBOOK.
- **what:** `orchestration_states` rows are written BY the worker process itself, so a dead pod (OOMKill exit 137, eviction, panic) writes nothing further — the row freezes at `EXECUTING_STEP` and `since_s` measures time since the crash, not a live hang. Correct triage order: RESTARTS column → `describe pod` Last State → `logs --previous` (capture crash logs IMMEDIATELY, since a ReplicaSet replacement erases them). Only after ruling out a dead pod should a suspected genuinely-stalled dependency be probed with a bounded call (`curl -m 5`). This specific incident walked through three wrong hypotheses in sequence (stall → missing context deadline → slow memory leak) before the real cause (the chunkContent infinite loop, DBG-035) was found — each correction was documented rather than silently discarded, per the "wrong-turns" convention.
- **sources:** `016b_debugging_guide_7_3_(7).md#executing-step-entry`; `RUNBOOK_travelling_docs(38).md#superseded-incident-block`; `RUNNING_NOTES_travelling_docs(39).md#rev34,#rev35,#rev36` (travelling_docs unit)
- **relations:** chunkContent infinite loop (DBG-035, the eventual answer); containment-limit corollary; guide-meta wrong-turns convention (DBG-001)
- **verify-later:** n/a — operational triage pattern

### DBG-035 — chunkContent() infinite loop — the OOM root cause across a four-migration bypass/fix/reenable arc
- **status:** deployed
- **status-evidence:** "THE OOM ANSWER (closes the incident chain for good)" confirmed 2026-07-10; fix deployed v1.0.1104; proof run 05d1fc97 with 0 pod restarts; four migrations (135/139/140/141) trace the whole arc.
- **what:** `chunkContent()` in `rag_actions.go` never terminated on content longer than `chunk_size`: at the final chunk, `start = end - overlap` steps BACKWARDS instead of forward, so the same tail is appended forever, producing ~2Gi of duplicate chunks in seconds and OOMKilling the pod (both observed OOMKills were PLAN-sized bodies passing through `index_plan`). Content ≤1000 chars returned early, which hid the bug for weeks. The incident was worked as a reversible arc: 135 bypassed `index_plan` (keeping ground truth in Postgres via `write_plan` while sacrificing only derived indexing) alongside a hygiene embedding-timeout deadline (139 re-enabled, prematurely — reoccurrence disproved the first hypothesis), 140 re-bypassed with the confirmed root cause recorded, and 141 re-enabled after the real Go fix (a final-chunk break plus a forward-progress guard) shipped with four regression tests using a 30s timeout that specifically catches loop regressions. Durable class rule: a content-below-threshold early return can hide a non-terminating path elsewhere in the same function; "a proof run is a probe — fire proofs early" (the proof run here found the real cause within the hour).
- **sources:** `RUNBOOK_travelling_docs(38).md#task-6`; `RUNNING_NOTES_travelling_docs(39).md#v1.0.1103-proof-run,#fix-140-141` (travelling_docs unit); `135_bypass_index_plan_until_embed_timeout.sql`, `139_reenable_index_plan.sql`, `140_rebypass_index_plan_chunk_loop.sql`, `141_reenable_index_plan_after_chunk_fix.sql` (sql_for_agents unit)
- **relations:** EXECUTING_STEP/OOMKill triage (DBG-034, the diagnostic path that led here); tool_docs indexing (unblocked by the fix); reversible SQL bypass pattern
- **verify-later:** `rag_actions_chunk_test.go` presence; deployed image ≥ fix commit

### DBG-036 — Env-prefix trap: a shell VAR=x on its own line (or before `;`) never reaches the child process
- **status:** deployed
- **status-evidence:** Cost two 3b.4 runs and one 085 run in one unit; independently logged as debugging-guide item #26 by a different unit ("shell variables never reach child processes without export, die with the session").
- **what:** Shell variables set on their own line, or terminated by `;` before the command, are not exported to the child process a script then invokes — so triggers silently run with default values instead of the intended override, with no error anywhere. Correct forms: a same-line prefix (`VAR=x command`) or an explicit `export`. The durable mitigation adopted was to make scripts print an explicit banner of the effective values actually in force as a load-bearing tell (e.g. "Subject: NONE — will SKIP") rather than trust that a preceding assignment worked.
- **sources:** `RUNNING_NOTES_travelling_docs(39).md#rev19,#rev33`; `RUNBOOK_travelling_docs(38).md#8` (travelling_docs unit)
- **relations:** Probe debugging-guide entries #24-28 (DBG-049, independently documents the same class as item #26); kcat single-line trigger discipline (DBG-018)
- **verify-later:** n/a — operational pattern

### DBG-037 — Two failure envelopes: a COMPLETED parent orchestration does not mean the child succeeded
- **status:** deployed
- **status-evidence:** Observed live in the 3a arc's first two runs; independently adopted as platform behaviour by at least three separate units, each naming the same two code paths (`sendWorkflowResponse` vs `notifyParentOfFailure`).
- **what:** A mid-run child failure is reported via `sendWorkflowResponse` with header `status:"complete"` but the real failure sitting in the response BODY (`body.status:"failed"`) — the parent forwards this and itself ends up COMPLETED with a non-empty `error` column, i.e. a forwarded child failure disguised as success at the header level. A failed-to-START child instead uses `notifyParentOfFailure` with `status:"error_unrecoverable"` / `CHILD_ORCHESTRATION_FAILED`. Any consumer of child results must check the body, never the header status alone; which of the two shapes appears tells you WHERE in the child's lifecycle it died.
- **sources:** `016b_debugging_guide_7_3_(7).md#failure-envelopes-entry`; `RUNNING_NOTES_travelling_docs(39).md#rev11,#rev12` (travelling_docs unit); `archive_april_26/016b_debugging_guide_7(4).md#"Two failure envelopes"` (archives unit); `docs019/RUNBOOK_builder_route(21).md#B5` (docs019_runbooks unit)
- **relations:** agent_error_log first-read discipline (DBG-015); silent-completion family (DBG-007); spawn-consumed columns lesson (DBG-058, carries this as a sibling gotcha)
- **verify-later:** `sendWorkflowResponse` / `notifyParentOfFailure` implementations, current form

### DBG-038 — Pod label `agent-type` (hyphen) vs log field `agent_type` (underscore) — silent zero-match selectors
- **status:** deployed
- **status-evidence:** Stated as a settled operational rule with a named failure mode in two independent units, each with a proven working command vs a zero-match selector.
- **what:** Kubernetes pod LABELS use the key `agent-type` (hyphen) while structured log JSON fields use `agent_type` (underscore); using the underscore form in a `kubectl logs -l` selector silently matches zero pods with no error. Separately, because a correct type-wide selector spans ALL live pods of that type (the idle reaper only recycles after 3600s), a tail can contain residue from an earlier, unrelated run — every line must be attributed by orchestration id / pod / timestamp before being read as current. Some older trigger scripts (082/083c) were found still carrying the incorrect underscore form.
- **sources:** `016b_debugging_guide_7_3_(7).md#label-entry`; `RUNNING_NOTES_travelling_docs(39).md#rev11,#rev13` (travelling_docs unit); `archive_april_26/016b_debugging_guide_7(4).md#"Pod label key is agent-type..."` (archives unit)
- **relations:** 0-rows rule (DBG-039); guide-meta (DBG-001)
- **verify-later:** grep trigger scripts 082/083c for the underscore `agent_type=` selector

### DBG-039 — 0-rows rule: a zero-row result is not decisive until the query AND the run's completion are both ruled in
- **status:** deployed
- **status-evidence:** Closed via the "state-dump substitute" technique 2026-07-06; codified in a 016b anchorless entry and RUNBOOK §7.
- **what:** A `0 rows` result only becomes decisive once you've confirmed both that the query itself is correct AND that the run in question actually completed (a run that died upstream also produces 0 rows for entirely unrelated reasons). When a step's NON-firing is itself the intended success condition, closing the loop needs three things together: a COMPLETED child orchestration, the step's explicit skip log line, and the 0-count. Skip log lines have only a 3600s capture window before the idle reaper recycles the pod; past that window, a post-completion state dump (ProcessingHistory showing the step executed, reached a terminal status, and produced a 0-count) is the accepted substitute evidence. A related hygiene rule: when following a documented placeholder pattern, replace it INCLUDING the angle brackets.
- **sources:** `016b_debugging_guide_7_3_(7).md#anchorless-entry`; `RUNBOOK_travelling_docs(38).md#7,#stage-3`; `RUNNING_NOTES_travelling_docs(39).md#rev16` (travelling_docs unit)
- **relations:** pod label attribution (DBG-038); agent_error_log first-read (DBG-015); `persist_diagnosis_note` gate proof
- **verify-later:** idle-reaper timeout value (3600s), confirm still current

### DBG-040 — Untracked-file deploy trap: verify a deploy by commit ancestry, not by tag or commit message
- **status:** deployed
- **status-evidence:** Struck twice independently (a Tier-2 checker missed two releases; `check_tool_acceptance_due` missed v1.0.1111); rules banked in a HANDOFF and the durable-rules list.
- **what:** `git commit -a` only commits already-tracked, modified files — a brand-new untracked (`??`) file can silently miss any number of release commits while its sibling changes ship normally, with no error. Guards: run `git status` for `??` files before every release (or commit new files as soon as they're written); verify a deploy actually carries a given commit by ANCESTRY (`git merge-base --is-ancestor <commit> <release>`), not by trusting a commit message; this repo also reuses version tags, so pod-start-time vs commit-time is what actually settles what a tag contains. A related safe-failure companion: unknown discovery-check names warn-and-skip rather than error, so wiring a check in by migration before its binary is deployed is safe.
- **sources:** `HANDOFF_2026-07-10…md#T8,T11,#4`; `RUNNING_NOTES_travelling_docs(39).md#stage-5-live,#v1.0.1111`; `README_summary_paragraph2_for_discussion.md` (travelling_docs unit)
- **relations:** continuous sweep gate; migrations-before-binary safety; chassis build/deploy practice (build from filesystem, decoupled from commits)
- **verify-later:** n/a — operational pattern

### DBG-041 — Convergence inertness: `[]map[string]interface{}` vs `[]interface{}` type-assertion bug
- **status:** deployed
- **status-evidence:** "A clean re-adoption proved `reconcilePlanWithRealised` has never run for any site… `query_database` returns []map[string]interface{} — a type that does NOT satisfy that assertion in Go… Fix… accepts both… plus a count log" (2026-06-05, verified fixed same day).
- **what:** The entire doc-029 Phase-1 convergence feature was dead since its original deploy because `ValidateSitePlanAction` asserted `ev.([]interface{})` on `QueryDatabaseAction`'s output, which is actually typed `[]map[string]interface{}` in Go — the assertion always failed, `existingPages` stayed empty, and reconcile silently early-returned every time with no error. A canonical instance of the "silent empty input" failure class; the fix pairs a proper type switch with an "existing pages loaded for convergence" count log so emptiness can never again fail silently. Also documented in the same fix: `QueryDatabaseAction` stringifies jsonb columns, so sections arrive as JSON strings needing an explicit `json.Unmarshal`.
- **sources:** `FOCUS_adoption_faithfulness_via_locks(5).md#2026-06-05-correction`; `running_notes_14(25)#part-14l` (adoption unit)
- **relations:** union-clobber fix (downstream of it); silent-completion family (DBG-007, same "silent empty input" pattern)
- **verify-later:** `v3_site_actions.go` type switch; 016 debugging guide v2_31+ entry

### DBG-042 — Defect-catalogue discipline: enumerate by root cause, then read-pin-confirm-fix
- **status:** deployed
- **status-evidence:** "Enumerate every observed defect as a separate item before fixing, so distinct causes are not conflated… causes marked 'tentative' have NOT been pinned" — practised consistently across defect Families A–J with per-item verification states.
- **what:** A post-deployment audit methodology: walk the live site, catalogue every observed defect grouped by root-cause FAMILY rather than surface symptom (e.g. deployment gaps, silent-fallback links, list-component content, section-data gaps, content quality, guides duplication, design fidelity, hygiene, unknowns, dispatch throughput), then work each family as its own thread — read the responsible action's code, pin the cause, confirm it against live data, and only then fix. Paired reading-discipline rules: `site_plan_pages` is the authoritative plan output; confirm a run actually completed before diagnosing it; teardown must be scoped by `site_id` never by domain; a matching `updated_at` is not proof of authorship; a hardcoded `site_id` is stale after any teardown (always resolve via a domain subquery); an empty LEFT JOIN means a wrong anchor id, not a genuinely missing link.
- **sources:** `CATALOGUE_gamesdesign_post_sync_fix_defects(9).md`; `HANDOFF_2026-05-25#reading-discipline`; `running_notes_14(25)#principles` (adoption unit)
- **relations:** guide-meta (DBG-001); F2 methodology (DBG-030)
- **verify-later:** n/a — methodology

### DBG-043 — Kafka consumer-group recovery: restart-to-rejoin, park-at-latest, never replay-from-earliest
- **status:** deployed
- **status-evidence:** "Resolution: chassis restart re-established group membership; a fresh trigger produced orchestration… Kafka recovery: restart-to-rejoin + park at latest + one fresh trigger, rather than replay-from-earliest (which would spawn stale adoptions)."
- **what:** After a topic wipe destroyed `__consumer_offsets`/group membership, the chassis logged a clean-looking consumer setup but was not actually joined to the group — triggers produced messages that nobody consumed (a site row was created by the trigger path with no corresponding orchestration row). Diagnostic: `kafka-consumer-groups --describe` showing empty membership means "not consuming" regardless of how healthy the producer side looks. Recovery doctrine: restart the chassis to force rejoin, park consumption at latest, and send exactly one fresh trigger — a `--reset-offsets --to-earliest` replay was tried and was a mistake, since it risked spawning duplicate stale adoptions from every historical message. Principle: a trigger printing an ID proves production happened, never that anything consumed it.
- **sources:** `running_notes_14(25)#part-4` (adoption unit)
- **relations:** orchestration creation writes a state row at creation (its absence proves "never consumed"); scheduler tick races during DB cleanup (separate, noted as noise)
- **verify-later:** n/a — operational doctrine

### DBG-044 — Manual work-item insertion as an operational rebuild lever
- **status:** deployed
- **status-evidence:** "Operational fact confirmed: a manually-inserted needs_page / needs_content_page work item IS claimed by build-dispatch-loop (status triaged → claimed → complete), so single-page (re)builds can be hand-triggered" (2026-06-06); verified end-to-end for three separate pages.
- **what:** Canonical hand-trigger shapes for unsticking a single page: to re-render existing components, insert a `needs_page` item with spec `{reason, page_name}`; to regenerate content, insert `needs_content_page` with spec `{mode:'recreate', source:'adoption', page_name, page_type}`. Both use `handler_agent='page-build-handler'`, `status='triaged'`, `ON CONFLICT DO NOTHING`, with ids resolved via a domain subquery. This is a verified, reusable escape hatch, though "complete" afterwards is still not proof on its own — the standard positive-evidence verification discipline still applies.
- **sources:** `HANDOFF_2026-06-06#key-references`; `GUIDE_deploy_from_context_packs.md#C`; `RUNBOOK(2)#4` (adoption unit)
- **relations:** dispatch pipeline; silent-completion family (DBG-007, "complete alone is not proof")
- **verify-later:** n/a — operational recipe

### DBG-045 — Kafka per-spawn response-topic partition race (adapter reply lost to wrong partition)
- **status:** partial
- **status-evidence:** Did not reproduce on a second run in one incident (2026-05-11); a later account (2026-07-12) reports it downgraded from silent-success to a visible failed item via the `mark_item_failed` fix, i.e. still occurring but no longer silently masked.
- **what:** An adapter (observed with kafka-go's LeastBytes balancer) occasionally writes its response to partition 1 of a per-spawn topic that only has partition 0, and Kafka rejects it ("topic partition not found") — so the underlying work (e.g. a git commit) succeeds but the orchestration times out or reports failure. Root cause is suspected stale partition metadata for just-created single-partition topics; it has never been structurally fixed, only downgraded from a silent success/failure mismatch to a visible failed item via the `mark_item_failed` pattern. The same race is suspected to have killed a content-writer reply, producing a "no-op complete" anomaly, and is suspected to affect any adapter writing to per-spawn topics (webscrape, image-generator), not just git-adapter.
- **sources:** `ANALYSIS_phase_2f_two_defects.md#Defect-2`; `RUNNING_NOTES_imagery_best_in_class.md#Turn-16`; `HANDOFF_imagery_best_in_class.md#Open-threads` (imagery unit); `imagery/old/ANALYSIS_phase_2f_two_defects(1).md#defect-2` (archive_classic unit)
- **relations:** `mark_item_failed` error-honesty fix; a separate, sibling parked defect (chassis response consumer-group race, `ANALYSIS_chassis_response_consumer_group_race.md`); Kafka topic-creation race self-heal (DBG-028, a DIFFERENT, self-healing Kafka defect)
- **verify-later:** `platform/kafka/producer.go` Balancer; topic_manager per-spawn partition count; adapter logs for "topic partition not found"

### DBG-046 — Work-item re-drive and zombie-claim operational semantics
- **status:** partial
- **status-evidence:** Standing lessons recorded across Turns 31–32 (2026-07-12/13); "the zombie-claim dispatch stall was the single biggest time cost of the 2026-07-09/10 verification" (still open); recurrence independently logged in a separate unit (2026-07-10/12) as "stuck-claim/zombie-handler dispatch noise... survived across the v1.0.1107 deploy," root cause not established.
- **what:** Hard-won dispatch mechanics: a claimed item stuck for more than ~10 minutes blocks its ENTIRE site via `find_dispatchable_site`'s NOT-EXISTS clause (worked around with a standing unstick UPDATE; the real fix — reaper cadence + a per-item-type circuit breaker — remains a TODO); re-driving a stuck item requires resetting `attempt_count=0` and clearing claim metadata, not just flipping status back (items that have exhausted attempts are silently excluded from dispatch, so it can look dead while actually correctly idle); a just-finished orchestration's tail can re-stamp a freshly-reset item complete (a state-machine race); manually-inserted items are NOT auto-triaged (must be inserted directly as triaged); dedup is a partial unique index on (site_id, item_key) over non-terminal statuses only, which makes resets awkward. Historical/adjacent gaps: dispatch once didn't claim triaged imagery items sitting behind page work; fairness/observability gaps (outer ORDER BY, the trigger not writing orchestration_states) remain open.
- **sources:** `RUNNING_NOTES_imagery_best_in_class.md#Turn-32`; `HANDOFF_imagery_best_in_class.md#Mechanisms`; `TODO_imagery_followups.md#6/#8/#9/#10`; `RUNBOOK_imagery_best_in_class.md#B9` (imagery unit); `docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#0`, `RUNNING_NOTES_minilobby_task.md#2026-07-12`, `087_dispatch_work_items_vonc.sh` (leopardess_social unit)
- **relations:** `mark_item_failed`; claim metadata not cleared on failed items; scheduler-and-tasks; leopardess O4 unstick procedure (same class)
- **verify-later:** `find_dispatchable_site` SQL; reaper cadence; `idx_swi_dedup` definition; claim/spawn/call sequence in build-dispatch-loop; agent_error_log around stuck claims

### DBG-047 — Pipeline field as a soft routing label (needs_imagery items excluded by pipeline='design')
- **status:** deployed
- **status-evidence:** "`check_unfulfilled_imagery_plan.go` hardcodes Pipeline: 'build' — the 2026-05-17 fix is in the code" (verified 2026-07-08), with a further dispatcher-filter loosening scoped alongside it.
- **what:** Discovery checks running under design-discovery-agent inherited a `pipeline='design'` label, which build-dispatch-loop's `item_pipeline` filter silently excluded, so `needs_imagery` items required manual UPDATEs before they'd dispatch at all. Two-part fix: checks now write `Pipeline:"build"` at source (the pipeline value should reflect the destination handler, not the emitting agent), and the dispatcher's filter was removed so any future mismatched emission still dispatches regardless. The field survives afterward as a soft routing label kept for a possible future multi-pipeline dispatcher.
- **sources:** `TODO_imagery_followups.md#7`; `RUNNING_NOTES_imagery_best_in_class.md#Turn-2` (imagery unit)
- **relations:** work-item dispatch semantics; design-discovery-agent context
- **verify-later:** build-dispatch-loop `load_items` config; Pipeline literal in imagery checks

### DBG-048 — Early pipeline-failure triage priorities dropped by deeper root-cause diagnosis
- **status:** abandoned
- **status-evidence:** The 2026-04-14 report's P3 (vonc.com raw CSS), P4 (stale-item process gap) and P5 (timeout tuning) do not appear at all in the following day's v3 report's P1-P10 list.
- **what:** A first-pass symptom-level triage of 57 stuck work items proposed three priorities that were superseded within a day by deeper diagnosis identifying concretely-fixed root causes that hadn't originally been named: rate-limit errors misclassified as non-transient (1,869 occurrences), `load_page_record` lacking a `page_id` fallback, and later audit-finding routing/classification bugs. Kept as a case study in how first-pass symptom triage can misdirect effort versus root-cause investigation.
- **sources:** `old/older1/105_dispatch-pipeline-failures-report.md#"Priority Fixes"`; `old/older1/105_dispatch-pipeline-failures-report_v3.md#"Priority Fixes"` (archives unit)
- **relations:** plan_sections pre-check evolution; three-way audit-finding classification
- **verify-later:** current state of vonc.com's about page (the raw-CSS-serving bug)

### DBG-049 — Probe-project debugging-guide entries #24-#28
- **status:** deployed
- **status-evidence:** running notes 2026-06-12 "Debug guide updated… 016_debugging_guide_v2_46" and 2026-06-13(g) "Debug guide v2_48."
- **what:** Five checklist entries earned during field work on the traffic-probe project, each with a dated real instance: #24 a config/workflow file is only authoritative at its actual runtime read-path (a stale `agentchassis/.git/workflows/deploy-to-b2.yml` nearly produced a never-firing Action); #25 prove the harness delivered the intended input before debugging the system under test (a non-expanding `dash` `$'…'` literalized a field to the string "$value"); #26 shell variables never reach child processes without export and die with the shell session — read state back from the artifact, not from `echo $KEY`; #27 never invent an interface — compiling standalone is not the same as satisfying the real DiscoveryCheck signature (this fixed a "backend_unreachable" bug); #28 `agent_definitions` is UNIQUE(type,version) and has two similarly-named category columns (this fixed an agent INSERT bug). Plus operator-handover lessons: explicit file manifests with a loud `go vet`/build check, flat-shipped workflows (the delivery channel rejects dot-directories), and `git branch -M main` before the first push.
- **sources:** `traffic_probe_running_notes(28).md#2026-06-12,#2026-06-13-g`; `traffic_probe_runbook(13).md#3.5-3.6` (traffic_probe unit); `traffic_probe_running_notes(27).md#2026-06-12-debug-guide,#2026-06-13-g` (archive_traffic_probe unit)
- **relations:** env-prefix trap (DBG-036, same class as #26); untracked-file deploy trap (DBG-040, adjacent class); guide-meta (DBG-001)
- **verify-later:** `016_debugging_guide_v2_48.md` entries #24-#28, current form

### DBG-050 — gamesdesign index/homepage silent-staleness: result-contract stub (output_field vs output_fields) root cause
- **status:** deployed
- **status-evidence:** Root cause pinned and fixed 2026-06-18 ("`resolveResultSpec`... treats singular as FLATTEN"); reached only after explicitly-superseded intermediate hypotheses in two independent write-ups (pod-dying-mid-flight → content-regression-guard-masked-error → the eventual generalized stub mechanism; and separately, a per-section max_tokens cap → a recreate-mode discriminator → the same eventual mechanism).
- **what:** The gamesdesign index/homepage page repeatedly "completed" rebuilds while never actually updating, diagnosed across several sessions with multiple wrong hypotheses discarded in turn before landing on the real cause: a January chassis regression made `SagaCoordinator.extractWorkflowResult` honour only the PLURAL `output_fields` key, while `page-content-writer` declares the SINGULAR `output_field`, so the compiled page collapsed into an oversized state-dump skip path that reported success while the live page never updated. Fixed with `resolveResultSpec` (`result_spec.go`, new), which treats a singular `output_field` declaration as a FLATTEN instruction rather than silently ignoring it. This specific bug's diagnosis arc — the reversals and the eventual fix — became the canonical worked example baked into the diagnosis-fix-loop's own verdict prompt as a ground-truth benchmark.
- **sources:** `archive_april_26/016_debugging_guide_v2_49.md`,`v2_49(1,2).md#§9` (archives unit); `NOTES_running_synthesis_v2(36).md` 2026-06-14/17; `NOTES_running_synthesis_principles(59)` 2026-06-13/14 (docs019_running_notes unit)
- **relations:** content-regression guard laundering (DBG-008, a candidate/related mechanism ruled out for THIS incident but real elsewhere); gpu-provisioner output_fields flattening (DBG-070, same bug CLASS on a different agent); diagnosis loop; silent-completion family (DBG-007)
- **verify-later:** confirm `platform/orchestration/result_spec.go` (`resolveResultSpec`) is present in the current codebase

### DBG-051 — Assumed-status-values trap: never assume a status column's vocabulary from naming convention
- **status:** deployed
- **status-evidence:** Formalized as a Section 9 addendum + Section 0 checklist candidate in `016_debugging_guide_addenda.md`.
- **what:** General lesson learned the hard way: always run `SELECT DISTINCT status FROM <table>` before writing logic that assumes plausible-sounding status values exist. Concrete instance: `pages.status` uses the literal value `'active'` exclusively, platform-wide — other plausible-sounding values simply never occur in the data.
- **sources:** `js_snippets_news_gaswholesalers/016_debugging_guide_addenda.md#Assumed-status-values-trap`; `old/design_actions_status_filter_fix.md` (docs024_small_dirs unit)
- **relations:** sites.status is informational (DBG-017, the same lesson on a sibling table)
- **verify-later:** grep/inspect `SELECT DISTINCT status FROM <table>` usage; `pages.status` values

### DBG-052 — "Renders empty" diagnostic method: a data-binding diagnosis, not a template diagnosis
- **status:** deployed
- **status-evidence:** Formalized into `016_debugging_guide_addenda.md` Section 9 headline entry + Section 0 checklist item #16.
- **what:** A reusable 5-step method for "a component renders its structural shell but no repeated content": (1) check `page_components` for orphaning; (2) confirm `input_schema` expectations; (3) check whether the structured data exists anywhere at all; (4) count rendered shells; (5) compare actual sections against the site plan for duplicate/stale slots. Core lesson: an empty shell means the template DID run — the bug is in data binding — so never trigger a rebuild before completing this walk, since a rebuild alone won't fix a data-binding gap.
- **sources:** `js_snippets_news_gaswholesalers/016_debugging_guide_addenda.md#empty-shells`; `old/empty_faqs.md` (docs024_small_dirs unit)
- **relations:** FAQ duplicate content-surface bug; rendered_html snapshot-not-view (DBG-053); isolated build test methodology (DBG-054)
- **verify-later:** n/a — process/design record

### DBG-053 — rendered_html is a snapshot, not a live view (stale render after a content_components migration)
- **status:** deployed
- **status-evidence:** Verified via a diagnostic query comparing `template_has_script_src` vs `rendered_has_script_src` on live gaswholesalers data.
- **what:** A migration to `content_components.html_template` does not retroactively affect already-built pages, because `page-rerender` reads `page_components.rendered_html` — frozen output from the last writer run — and never re-pulls from the live template. General principle: `rendered_html` is a snapshot, not a live view; any migration touching `content_components` must also either update affected pages' snapshots directly or explicitly trigger a rebuild for them.
- **sources:** `js_snippets_news_gaswholesalers/016_debugging_guide_addenda.md#Migration-updates`; `old/findings_and_plan_news_visual.md` (docs024_small_dirs unit)
- **relations:** files_field deploy bug; "Renders empty" diagnostic method (DBG-052)
- **verify-later:** grep/inspect `content_components.html_template`; `page-rerender`; `page_components.rendered_html`

### DBG-054 — Isolated build test methodology: throwaway test-page pattern for pipeline-layer attribution
- **status:** deployed
- **status-evidence:** Used successfully to prove the content writer was NOT the bug in a FAQ investigation.
- **what:** A reusable diagnostic technique: create a throwaway page (kept out of nav) with a deliberately minimal/isolated section list, drive it through the full production build path unmodified, then read out `page_components` to conclusively attribute a bug to one specific pipeline layer rather than guessing. Used to prove the FAQ writer worked correctly in isolation, redirecting the investigation to a different layer.
- **sources:** `js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md#The-page-content-creation-flow`; `old/page_content_creation_flow.md` (docs024_small_dirs unit)
- **relations:** FAQ duplicate content-surface bug; "Renders empty" diagnostic method (DBG-052)
- **verify-later:** n/a — method

### DBG-055 — error_step must live inside step.Config, not at step level (silently ignored otherwise)
- **status:** deployed
- **status-evidence:** Adopted cross-thread from another unit's notes (001 §16 finding); independently re-derived and banked as a "durable rule" in a later incident (migration 128, 2026-07-10) with 131/132 retroactively moving ten inert step-level `error_step`s into config.
- **what:** The chassis workflow coordinator reads `step.Config["error_step"]` only — a step-LEVEL `error_step` key (i.e. placed as a sibling of config rather than inside it) is silently ignored, with dormant instances of this misplacement existing in tool-agent workflows. A routing target that names a non-existent step fails the whole workflow, so `error_step` targets should be derived from the step's own `next_step` rather than guessed. Correct-while-touching policy: migrate any old inert step-level key to inside config whenever that workflow is next edited for any reason. Separately (same unit): spawned agent pods are reaped ~3600s after going idle, so post-completion evidence should come from the orchestration state's ProcessingHistory dump rather than pod logs.
- **sources:** `docs019/RUNBOOK_diagnosis_fix_loop(9).md#collision-surface` (docs019_runbooks unit); `128_fix_load_runtime_error_step_target.sql`, `127_diagnose_load_runtime_error_step.sql`, `131_tool_generator_plan_writing.sql`, `132_fix_agents_note_writing.sql` (sql_for_agents unit)
- **relations:** prompt-template rendering resolvers (DBG-033); 0-rows/agent_error_log discipline (DBG-015, DBG-039)
- **verify-later:** coordinator error-routing code, current form; error_step placement across remaining tool agent workflows

### DBG-056 — Stage-by-stage rebuild verification and the false-complete rule
- **status:** deployed
- **status-evidence:** A five-stage (A–E) verification method with per-stage SQL, stated as: "status='complete' is only meaningful together with Stage C showing changed components; complete + unchanged components = the old false-complete."
- **what:** A verification method for confirming a page rebuild genuinely happened, staged A through E: A — the writer delivered a flat result to the parent (sections_metadata path check); B — the save was attempted and, if blocked, logged loudly (`agent_error_log`); C — components ACTUALLY changed (content_hash/updated_at fingerprint vs a captured baseline); D — the work item completed on a REAL save (meaningful only together with C); E — deploy. Baseline fingerprints must be captured BEFORE triggering, an existing stuck work item should be re-opened rather than fabricating a new one, and a triage table maps each stopping stage to its likely cause.
- **sources:** `docs019/RUNBOOK_gamesdesign_index_rebuild.md#2,#5,#7` (docs019_runbooks unit)
- **relations:** content-regression guard (DBG-008); silent-completion family (DBG-007)
- **verify-later:** `page_components` fingerprint queries; `site_work_items` re-open pattern

### DBG-057 — Code-retrieval corpus staleness masquerading as a retrieval-quality problem
- **status:** deployed
- **status-evidence:** "corpus check result: the index is the blocker... the index is of a YEAR-OLD tree" (2026-07-02/03).
- **what:** After the diagnosis loop was measured to gain nothing from code retrieval, a follow-up investigation found the underlying `code_symbols` index itself had been built from a year-old stale checkout of the default branch (stale since 2025-07-14) — a corpus freshness problem masquerading as a retrieval-quality problem. This led to a reindexing effort, a ref-pinning strategy, and ultimately a decision to migrate the code-indexer's analysis step onto the already-proven `analyse_repo_local` path instead of maintaining the separate stale index.
- **sources:** `NOTES_running_synthesis_v4(39).md` headers 2026-07-02/03, DECISIONS section (docs019_running_notes unit)
- **relations:** B4a embedding-quality finding; code-context retrieval infrastructure
- **verify-later:** current freshness of the deployed `code_symbols` index; whether the analyse-step migration was applied

### DBG-058 — Spawn-consumed columns lesson: seeds must copy image/infra columns from a live donor row
- **status:** deployed
- **status-evidence:** Documented incident + fix in `NNN_fix_researcher_spawn_columns`; carried forward as a standing guard in `HANDOFF_builder_thread` ("seeds must copy image columns from a live donor (the amended seed does)").
- **what:** `getAgentDefinition` SELECTs `image_repository`/`image_tag`/`command`/`resources`/`health_config`/`env_vars`/`idle_timeout_seconds` and gates on `is_active=true`; a hand-written seed populating only `default_config` leaves `command` NULL, so the image's default entrypoint boots the GENERIC chassis service, which never reads the injected AGENT_TYPE env var — the dispatcher's call goes unheard and the item sits claimed forever with no visible error. Fix and standing rule: copy the spawn-consumed infrastructure columns from a proven donor row (deliberately NOT capabilities/topics/default_config, which should stay agent-specific). A related, separate gotcha carried alongside: `image_tag` DEFAULT 'latest' pointed at an ancient build (the Makefile now pins IMAGE_TAG explicitly).
- **sources:** `NNN_fix_researcher_spawn_columns.sql`; `HANDOFF_builder_thread.md#2,#5`; `HANDOFF_fixloop_thread(8).md#3` (docs019_design_plans unit)
- **relations:** workflow-in-default_config lesson; index-orchestrator spawn wrapper; two failure envelopes (DBG-037, carried as sibling gotcha in the same source)
- **verify-later:** guidelines 001 New Agent checklist line (flagged as a residual gap)

### DBG-059 — orchestration_state_audit: a temporary attachable trigger for diagnosing state races
- **status:** deployed
- **status-evidence:** Create-trigger + analysis queries (`time_since_prev` via LAG, `pg_backend_pid`, `application_name`) with an explicit "Remove trigger when done investigating" teardown step.
- **what:** A temporary, attachable audit table plus an AFTER UPDATE trigger that captures every version/status/current_step transition on `orchestration_states` — used to diagnose state races and stuck orchestrations, then explicitly removed afterward. Distinct from the permanent logs (`agent_error_log` etc.); also swept by the database-cleanup scheduled task (keeps only the last 100k rows) if left attached.
- **sources:** `docs/agent_docs/sql_for_tables/010_orchestration_state_audit.sql`; `020_scheduled_tasks.sql#database-cleanup` (sql_tables_components unit)
- **relations:** agent_error_log (DBG-015); database cleanup
- **verify-later:** whether the trigger is currently attached anywhere

### DBG-060 — Message-flow logging / observability plan (never fully built)
- **status:** aspirational
- **status-evidence:** An early README's Week-2 objective: "MessageFlowLogger… Track every message through the system with database persistence"; a problem statement repeats the same desire; only zap logging plus `orchestration_states.processing_history` was ever evidenced as actually built.
- **what:** An early-platform aspiration to persist every send/receive event, agent creation, and topic-routing decision to the database for replay/debugging. A dedicated message-flow store never materialized; the closest existing analogues are zap application logs and the `processing_history` field inside `orchestration_states`.
- **sources:** `docs001_flow_general/README.002.agent_orchestration1.philosophy.md`; `docs002_hitl_parallel/README.0100.state_of_play_for_creating_website.md` (legacy_docs_a unit)
- **relations:** guide-meta (DBG-001, its successor observability practices); `processed_messages` table (exists — see env reset runbook, DBG-061)
- **verify-later:** `processed_messages` table's actual purpose; whether any message-audit table now exists

### DBG-061 — Orchestration environment reset runbook (clean-slate test-cycle procedure)
- **status:** deployed
- **status-evidence:** The identical script repeated verbatim across five or more early docs.
- **what:** The standard clean-slate procedure used across the early platform's test cycles: scale agent-chassis to 0, TRUNCATE `processed_messages`/`orchestration_states`/`pending_requests`, delete spawned jobs, delete all `job.*` topics and bootstrap topics, reset all consumer-group offsets to earliest, then scale back up. Also documents the persistence surface of that era: `processed_messages` (dedupe), `orchestration_states`, `pending_requests` tables; `job.*` and `system.agent.*` topics; `spawned-by=orchestrator` job labels.
- **sources:** `docs001_flow_general/README.095d.mycurrentinputmessagebeforechanging.md`; `README.096d.robotics_startmessage.md`; `docs004_website_capture_project/initial_messages/initial_messages.txt` (legacy_docs_a unit)
- **relations:** message-flow logging plan (DBG-060); stateless-agents concept (what gets truncated)
- **verify-later:** whether `pending_requests`/`processed_messages` tables are still present in the current schema

### DBG-062 — Early message-routing failure-mode catalogue (case studies behind the core architecture)
- **status:** deployed
- **status-evidence:** Each bug has a traced fix document from the early platform: nested-vs-flat input_data mismatch, verbose child responses breaking aggregation, silent root completion, duplicate second response to own topic ("poisoned pill" crash-loop), responses_topic dropped in header parsing, missing `in_response_to_request_id`, fire-and-forget spawn ignoring init responses.
- **what:** The canon of early failure modes that directly shaped the platform's core conventions: every major architectural convention (data normalisation, reply-to storage, perspective transformation, single completion path, await semantics) exists because it was the concrete fix to one of these seven traced production bugs. Valuable primarily as diagnostic priors — most modern debugging-guide failure patterns are descendants of one of these.
- **sources:** `docs001_flow_general/README.011.flow2.md`; `README.016.flow5.md`; `README.4.2.lifespanofresponsestopic.md`; `README.023.flow12.await_response.md`; `README.012.flow3.md` (legacy_docs_a unit)
- **relations:** all core system-architecture concepts; guide-meta (DBG-001, the 016b successor of this same debugging lineage)
- **verify-later:** none — historical lessons

### DBG-063 — Parent-timeout vs child-HITL race
- **status:** deployed
- **status-evidence:** docs014 log trace: "pageflow-builder times out (5 min)... content-reviewer: Cleaned up expired awaited requests count=1. The fix is to increase the parent's timeout."
- **what:** A failure class where a parent's `call_agent` timeout fires before the child's human-in-the-loop request can be answered; the parent then retries with a null body, and the child's awaited request gets cleaned up as expired, permanently losing the pause point. Fix: parent timeouts must always exceed the child's HITL timeout window, not just its normal processing timeout.
- **sources:** `docs014_research_agent/001_human_in_the_loop_response_flow.md#Why-There-Were-No-Awaited-Requests` (legacy_docs_b unit)
- **relations:** stale orchestration sweeper; HITL protocol; timeout chain ordering (DBG-004, a different but analogous nesting requirement)
- **verify-later:** current `call_agent` timeout_seconds vs HITL timeout defaults

### DBG-064 — Orchestration debug log taxonomy (early ancestor of the formal debugging guide)
- **status:** superseded
- **status-evidence:** Raw notes listing grep targets ("DEBUGaa: What have I done with CollectedData", "The Golden Search: grep -B 5 -A 30 generate_html") plus a real database-lock incident (idle-in-transaction blocking INSERT INTO sites).
- **what:** The earliest debugging playbook: canonical log messages for action-execution flow, LLM calls, data extraction, and CollectedData tracking, with matching kubectl grep recipes, plus `pg_stat_activity` lock triage and `pg_terminate_backend` for idle-in-transaction blockers. Directly ancestral to the formal 016/016b debugging guides.
- **sources:** `docs006_workflow_builder/010_debugging.md` (legacy_docs_b unit)
- **relations:** guide-meta (DBG-001); data-path problem
- **verify-later:** whether DEBUGaa-style markers remain anywhere in the current codebase

### DBG-065 — Mode A / Mode B broken-template taxonomy, repair/regeneration routing, and the pre-extraction JS-shell class
- **status:** deployed
- **status-evidence:** Code delivered 2026-06-22/23 (`checkBrokenTemplateSlots`, `repair_template_slots`); gauntlet-interface repaired under Mode A, archetype-result-card regenerated to q100 under Mode B; a related earlier-vintage class (js-not-extracted shells) was root-caused 2026-06-29, with the cosmetic-extraction backlog item still open 2026-07-09.
- **what:** Two distinct broken-template failure modes in the component library: Mode A — `<no value>FIELD</no>` — a render output was stored as source with field names surviving as fallback text, repairable in place by string substitution (`repair_template_slots`); Mode B — a bare `<no value>` with no field-name tag — the template was rendered against an empty context and the cleaned output stored back, so the field names are irretrievably lost and the only fix is `needs_component_regeneration` routing to component-creator. `repair_template_slots` detects Mode B (absence of `</no>` tags) and returns `needs_regeneration` rather than attempting a doomed repair; `checkBrokenTemplateSlots` is the discovery check that surfaces both modes. A related, earlier-vintage class: `provocation-card`, `lobby-grid` (and `brief-explanation`) were stored through a pre-`separateInlineJS` code path — raw inline `<script>` still in `html_template`, empty `js_content`/schema, `<no value>` placeholders — so their JS bundle was never produced and built-in interactivity never deployed; one of them was additionally truncated mid-generation (missing `</script>`), which shipped and swallowed the page footer when live.
- **sources:** `docs/RUNBOOK_vonc_session(1).md#structural-findings`; `RUNNING_NOTES_vonc(36).md#two-broken-template-failure-modes`; `RUNBOOK_vonc_migrations(14).md#step-1`; `RUNBOOK_phase2_provocation_js(29).md#extraction-bug-findings`; `RUNNING_NOTES_vonc(36).md#2026-06-29,#2026-07-02`; `RUNNING_NOTES_vonc_v2(28).md#2026-07-07-gate-passed` (vonc unit)
- **relations:** store-path validation hardening (rejects `<no value>` at the write gate); component regeneration in place
- **verify-later:** `check_component_standards.go`; `fix_component_template_action.go` `repairNoValueSlots`; `content_components.js_content`/`html_template` for provocation-card 6163ff14 and lobby-grid 9304f14d (still raw inline?)

### DBG-066 — Snapshot-shadowing defect: version+1000 snapshot rows outrank the active row in naive loaders
- **status:** superseded
- **status-evidence:** Root-caused 2026-05-11; patched in `processor.go::loadAgentDefinition` and `spawn_actions.go::getAgentDefinition` with an `is_active=true AND (is_snapshot IS NULL OR is_snapshot=false)` filter.
- **what:** The model-swap/rollback mechanism (`snapshot_agent()`) creates snapshot rows at `version+1000, is_snapshot=true`; any loader doing a naive `ORDER BY version DESC LIMIT 1` without an explicit `is_snapshot` filter reads version 1001 (a stale pre-migration snapshot) ahead of the true active row at version 1, silently shipping stale/pre-migration workflow config despite the DB otherwise being in the correct state. Structural and latent since launch; it only surfaced once a downstream phase (2F) first depended on a value that actually differed between the active and snapshot rows. Fixed by adding the snapshot filter to both known loaders.
- **sources:** `imagery/old/ANALYSIS_phase_2f_two_defects(1).md#defect-1`; `PLAN_imagery_loop_closure(9).md#2f` (archive_classic unit)
- **relations:** replacement fix locations (`processor.go`/`spawn_actions.go`); `021_model_swap_and_rollback.sql` `snapshot_agent()`
- **verify-later:** grep "FROM agent_definitions" *.go for any other loader missing the `is_snapshot` filter

### DBG-067 — Secret hygiene: image-provider API keys logged in plaintext, rotation repeatedly deferred
- **status:** partial
- **status-evidence:** "SECURITY (highest): scrub + rotate STABILITY_API_KEY and BANANA_API_KEY (plaintext in logs; Banana on paid tier)"; a later TODO repeats "STILL OPEN, STILL HIGHEST PRIORITY (do not let slide)."
- **what:** Image-generation provider API keys (Stability AI, Banana — a paid tier) were found being logged in plaintext; the standing highest-priority remediation is to scrub existing logs and rotate both keys. This item was repeatedly carried forward across multiple imagery-project sessions without ever being closed out.
- **sources:** `imagery/old/HANDOFF_robot_hands_rebuild(2).md#carried-forward`; `TODO_imagery_followups(15).md#security` (archive_classic unit)
- **relations:** image generation pipeline; adapter deployment; storage secrets
- **verify-later:** whether adapter logging of STABILITY_API_KEY/BANANA_API_KEY was fixed; secret rotation status in personae-default-secrets

### DBG-068 — Adapter-vs-chassis deployment drift (separate K8s resources, separate binaries)
- **status:** partial
- **status-evidence:** "Adapter deployment vs chassis deployment (2026-05-14)": the image-generator adapter's `dynamic_adapter.go` is a separate K8s resource from the chassis's `generate_image_actions.go` — a chassis rebuild+rollout did not refresh the adapter binary, and an expected new log line wasn't found on adapter pods.
- **what:** The image-generator adapter and the agent-chassis are deployed as physically distinct Kubernetes resources, so rebuilding and rolling out the chassis does not refresh the adapter's binary — action-layer changes made in the chassis (e.g. per-kind cfg_scale/negative_prompt tuning) can remain silently inactive at the adapter indefinitely. Recommendation (not confirmed applied): explicitly document which deployment carries which binary, and add the adapter to the standard rebuild/rollout sequence.
- **sources:** `imagery/old/PLAN_imagery_loop_closure(9).md#known-issues`; `HANDOFF_robot_hands_rebuild(2).md#carried-forward` (archive_classic unit)
- **relations:** image request shape; Stability timeout 30s→120s side-fix; multi-cluster dispatch; CrashLoop binary-mismatch family (DBG-011, same "which binary is actually running" concern)
- **verify-later:** image-generator-adapter deployment vs chassis image tag; rollout sequence in the Makefile

### DBG-069 — Launcher reply-topic own-vs-parent derivation (Decision D4)
- **status:** deployed
- **status-evidence:** "D4 CONFIRMED live … the adapter's reply went to system.agent.generic.responses — the agent's own ExecutionContext.ResponsesTopic" (2026-06-02).
- **what:** An intermediate adapter reply must be routed to the agent's OWN `ExecutionContext.ResponsesTopic` (seeded from `__my_responses_topic__`), never `__parent_responses_topic__` (reserved for the final child→parent notification only). An inherited handoff document had this backwards; the correct pattern (own-topic) was already used successfully by provision/decommission. The same bug class independently bit `dispatch_thunder_ssh_get_status` (cloned from ssh_exec) and was fixed to prefer `execCtx.ResponsesTopic`; a latent instance of the same mistake was flagged as possibly still present if ssh_exec dispatch is ever fired top-level.
- **sources:** `phase5/NOTES_phase5_training_launcher_running(39).md#3(D4),#6,#10`; `docubundle/.../STATUS_thunder_adapter_2026-06_04.md#update-2026-06-04` (archive_finetuning unit)
- **relations:** send-before-register await race (DBG-023, same subsystem); superseded 2026-05-24 handoff claims
- **verify-later:** `thunder_prepare_object_url_dispatch.go`, `thunder_ssh_exec_dispatch.go`, `thunder_ssh_get_status_dispatch.go`; coordinator `determineResponsesTopic`

### DBG-070 — gpu-provisioner output-shape flattening (output_fields plural vs output_field singular)
- **status:** deployed
- **status-evidence:** "extractWorkflowResult … reads output_fields — PLURAL only. The gpu-provisioner complete uses output_field (SINGULAR) … falls to the fallback branch" (2026-06-03); fixed by migration 104.
- **what:** `call_launcher` failed with "provisioning_result.provisioning_id not found" because gpu-provisioner's `complete` step declared the SINGULAR `output_field` key, which `extractWorkflowResult` never reads (only the plural `output_fields`), so its result surfaced step-name-keyed instead of field-keyed. Migration 104 fixed the provisioner's `complete` step to use plural `output_fields:["dispatch_provision"]` and re-pointed the launcher's input mapping to `provisioning_result.dispatch_provision.provisioning_id`. A proper chassis-side fix (making `extractWorkflowResult` also honour the singular form) was explicitly considered and vetoed in favour of making the non-compliant agent conform to the existing convention.
- **sources:** `phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-03-15:47,17:4x,17:5x` (archive_finetuning unit)
- **relations:** gamesdesign result-contract stub (DBG-050, the SAME underlying `output_field`/`output_fields` bug class on a different agent); launcher input-mapping contract; same singular bug flagged as latent in thunder-reaper
- **verify-later:** `extractWorkflowResult`; `agent_definitions` gpu-provisioner (0bf9fa8a); migration 104

### DBG-071 — Marker/attribute REPLACE anchoring and `hidden`-vs-author-CSS landmines in stored HTML
- **status:** deployed
- **status-evidence:** The marker-anchoring bug was introduced twice independently (provocation-card, lobby-grid), fixed via `fix_marker_selector.sql` with RETURNING checks (still_broken=f ×4) and redeployed 2026-07-04; the hidden-attribute ghost-row fix was verified end-to-end 2026-07-08 (rendered_len 7455→7671; live grep confirms 2 instances); both landed in the guide and as component-creator requirements.
- **what:** Two related landmines when hand-editing stored component HTML by SQL: (1) replacing the bare string `data-component="X"` to add an attribute ALSO hits that section's own inline `querySelector('[data-component="X"]')`, corrupting the selector into a malformed two-attribute string and throwing a SyntaxError that kills the section's cosmetic IIFE (loaders are unaffected) — the rule is to anchor marker REPLACEs on the OPENING TAG (followed by more attributes) and separately revert only the in-selector copy (followed by `]`), or better, emit markers at generation time; (2) the `hidden` attribute is only a UA-stylesheet `display:none` and loses to ANY author `display` rule on the same element, so a hidden clone-template row inside a `display:grid` container renders as a visible ghost row — fixed with a more specific author rule `[data-…-template]{display:none}` applied to both the template and its mobile media-query copy, with prevention requiring component-creator to always emit the hiding rule alongside a `hidden` attribute on clone templates. General rule for multi-line block edits: dump → edit offline → UPDATE the full text → verify by length delta, never a multi-line SQL REPLACE of nested markup.
- **sources:** `docs/fix_marker_selector.sql`; `RUNNING_NOTES_vonc_v2(28).md#2026-07-04-marker-replace-broke,#2026-07-08`; `docs/016b_debugging_guide_merged(3).md#data-runtime-fill-marker-anchoring,#hidden-attribute-loses` (vonc unit); `HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#8`; `docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-archive-list.md#2026-07-08` (leopardess_social unit)
- **relations:** SQL surgery method (DBG-016); generation-time guards (the actual prevention); clone-template list pattern
- **verify-later:** component-creator prompt includes the hiding-rule requirement?

### DBG-072 — Problem-category taxonomy for component/tool defects
- **status:** deployed
- **status-evidence:** In active use across every NOTES file in one unit ("Categories:" lines); seed set defined 2026-06-29, extended organically in practice.
- **what:** A shared, greppable vocabulary for tagging every incident so patterns roll up into the global debugging guide: css-variable-mismatch, empty-shell/mode-b-template, broken-template-slots, content-vs-runtime-mismatch, detool-on-rebuild, js-not-extracted, js-bundle-stale, schema-template-drift, plus organically-added tags (assembly-drop, planning-gap, silent-noop-success, cta-graph, css-specificity, method-correction). When a category recurs across multiple tools it graduates to a global pattern with a systemic fix — exactly how the empty-shell and visible-content-filter issues surfaced.
- **sources:** `docs/social001_vonc_tiktok_social/tool_docs/TOOL_DOCS_convention(3).md#Problem-category-taxonomy`; `NOTES_provocation-card(12).md` (leopardess_social unit)
- **relations:** guide-meta (DBG-001); per-tool travelling docs
- **verify-later:** 016b entries fed from these categories

### DBG-073 — Workflow monitoring REST endpoints (built but apparently unused in practice)
- **status:** unknown
- **status-evidence:** An architecture doc lists GET /monitor/workflows, /monitor/workflow/{id}, /monitor/stuck?hours=n, /monitor/metrics as built ("Each agent exposes monitoring endpoints"), but no later document in the same unit's material ever actually uses them — operational debugging goes through psql/db-inspector instead.
- **what:** A documented per-agent HTTP monitoring API over orchestration state: list active workflows per client, inspect a specific workflow's execution path/state, find workflows stuck for N hours, and aggregate metrics — complemented by per-step execution_path timing records and execution_metadata counters in the state row. Never observed in actual use anywhere else in the corpus.
- **sources:** `docs/architecture/004-agent-chassis-architecture.md#monitoring-and-observability` (misc_dirs unit)
- **relations:** database-backed workflow state; kcat/db-inspector runbook (DBG-074, the practice that actually survives); guide-meta (DBG-001)
- **verify-later:** whether /monitor routes exist in the current chassis HTTP server code

### DBG-074 — kcat + db-inspector operational runbook (early ops playbook)
- **status:** deployed
- **status-evidence:** Working command logs with real pasted outputs (correlation IDs returned, "0 rows" failure cases) for triggering and tracing workflows in the live cluster.
- **what:** The early ops playbook still in practical use: scale deployments up/down; inject workflow-start messages via kcat from an in-cluster pod with the full required header set; fetch the latest correlation_id from orchestrator_state; watch progress with the db-inspector tool (-watch mode); trace a specific agent by finding its spawned instance ID, then grep the shared chassis pod logs (agents don't get dedicated pods); check consumer-group lag, response topics, ServiceAccount job-creation rights, and Kubernetes events for spawned jobs.
- **sources:** `docs/basic_usage/001basic_usage.txt`; `docs/basic_usage/004_debugging` (misc_dirs unit)
- **relations:** agent spawning; website-builder group; workflow monitoring REST endpoints (DBG-073, the unused alternative); guide-meta (DBG-001)
- **verify-later:** tools/db-inspector, tools/kafka-producer still exist; whether the runbook matches the current namespace/topics
