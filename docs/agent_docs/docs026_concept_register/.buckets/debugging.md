
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Debugging assumption checklist (28-item process discipline)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §0, each item tied to a dated real defect through 2026-06-13
- **what:** The distilled pre-change checks: per-action _field conventions; input_mapping required-by-default; error_preview before log grepping; partial DB rows; SQL-immediate-vs-Go-deployed; sibling functions as canonical pattern; token budgets vs structured output; set -u in trigger scripts; jq slurpfile nulls; manual triggers to isolate dispatch-vs-handler; parent/child orchestration rows; `?` placement; \d before SQL; refire-before-refactor; pod-rotation log loss; don't change evidence-proven values; deploy ≠ migration ran; interface widening breaks all importers; prompt_rendered proves input not output; updated_at is not authorship; re-resolve site_id after teardown (zero-row LEFT JOIN = wrong anchor); check design docs for deliberate deferral; output_fields plural; config authoritative only at its runtime read-path; prove the harness delivered input (dash vs bash); env vars vs shell locals + stale deployed copies; read the interface definition; agent_definitions UNIQUE(type,version).
- **sources:** 016 §0 items 1–28 + 016_additions
- **relations:** everything in §9; 016b durable invariants
- **verify-later:** —

<!-- SOURCE: U01_docs024_numbered_core.md -->
### An agent is a DB row; trust default_config over prose; two possible definition sources
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §6.0 with build-dispatch-loop description-vs-config contradiction example
- **what:** Agents live in agent_definitions.default_config.workflow — grepping Go finds actions, not agents. Descriptions can contradict configs (trust the config). agent_definitions may be read from templates_db or clients_db depending on pod — confirm which copy the running pod loads before patching.
- **sources:** 016 §6.0
- **relations:** orchestration state; snapshot conventions
- **verify-later:** which DB each deployment reads definitions from

<!-- SOURCE: U01_docs024_numbered_core.md -->
### LLM step config shadowing (ai_service resolution order; dead temperature paths)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016 §6.6: bug live as of 2026-05-18 (22 of ~60 agents shadowed); structural per-field fix "planned"
- **what:** ExecuteLLMPromptAction resolves ai_service top-level → step-level → StepConfig and stops at first match, so a top-level ai_service shadows every step override (incl. doc-023-style per-step model swaps); max_tokens falls to hardcoded 2048 (tell: output_tokens exactly 2048); step.config.max_tokens sibling is never read; temperature is read ONLY from default_config.temperature top-level (all other locations dead) and llm_call_log.temperature was universally NULL. Fix path: per-field fallback chain + raise floor to 8000 + log sent values.
- **sources:** 016 §6.6
- **relations:** model swap functions; __sent_* write-backs (001(5) suggests later capture landed)
- **verify-later:** whether per-field resolution shipped; llm_call_log temperature now populated

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Timeout chain ordering (claim > call_handler > handler workflow)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §7 with current values and both mis-ordering failure modes
- **what:** claim timeout (30m) must exceed dispatch call_handler (20m) must exceed handler workflow timeouts; otherwise duplicate handlers (claim reset mid-work) or orphaned completions (dispatch gave up early). Idle monitor 3600s fallback; K8s ActiveDeadline 24h ceiling.
- **sources:** 016 §7
- **relations:** claim-lease-too-short reproducible timeouts (v2_49 sub-case b)
- **verify-later:** current values across dispatch/handlers

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Reaper false-positive completions (claimed-item-timeout evidence checks too loose)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016 §9 with confirmed gaswholesalers instance (auto-complete 47 min before the real commit); fix described as small SQL, not marked applied
- **what:** The "verified done despite lost response" branch auto-completes on ANY page updated on the site (not the target page, and updated_at not deployed_at) — treat empty-result + 'Auto-completed' items as untrusted. Correct evidence: p.id = wi.page_id AND deployed_at > claimed_at; needs_rerender/needs_design shouldn't auto-complete this way. Sibling issue: orchestration engine doesn't enforce awaited_requests timeout_at (spawn-handler hangs until reapers paper over it).
- **sources:** 016 §9 claimed-item-timeout + spawn_handler-hang entries
- **relations:** silent-completion family; timeout chain
- **verify-later:** claimed-item-timeout pre_query current form

<!-- SOURCE: U01_docs024_numbered_core.md -->
### jsonb && operator class bug (silent CSS-snippet failure vs hard JS failure)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §9: css path "silently failing the entire time"; JS analog fixed in same change set (May 2026)
- **what:** Postgres has no jsonb&&jsonb overlap operator; `applies_to && $1::jsonb` errored forever, swallowed by a logger.Warn-return-"" handler, so css_snippets never reached any deployed styles.css. Fix: EXISTS + jsonb_array_elements_text. Wider lesson: silent-failure loaders + graceful consumers hide months-old breakage — prefer hard failure when the data is supposed to be there.
- **sources:** 016 §9 jsonb && entry
- **relations:** best-effort-needs-monitoring; audit grep pattern for other && uses
- **verify-later:** loadComponentCSSSnippets fixed in place

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Silent-completion family (trust the artefact, not the status)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016/016b recurring entries; 016b names it the first durable invariant
- **what:** Work items report complete while the work didn't happen, via: result-contract stub (fixed 06-18); content-regression guard error masked by error_step complete_error; pod dying mid-flight (complete with non-empty error); "git committed the file" re-committing stale stored components; zero-planned-sections completing as success. Verify against page_components timestamps + live HTML. Companion rules: completed_at is orchestration END not write instant (trace child orchestrations by page_id in collected_data — trap part 3); intermediate signals (work-item names, pod snapshots, mid-flight tables) lie (trap part 2).
- **sources:** 016 §9 several entries + traps 1–3; 016b invariants
- **relations:** workflow result contract; zero-planned-sections
- **verify-later:** —

<!-- SOURCE: U01_docs024_numbered_core.md -->
### save_page_sections is the sole page_components writer; its section-regex fallback and content-regression guard
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016 §9 tool-pages-never-deploy (fix shipped 2026-05-28, end-to-end verification honest-open); guard masking fix flagged not confirmed applied
- **what:** save_page_sections DELETE+INSERTs page_components (history row written; source_item_id NULL on overwrite path — gap). Its HTML fallback extracted only `<section>` blocks, so `<div class="tool-page">` tool HTML was silently discarded (all tool/game pages n_rendered=0, rerender skips, no file ever committed); fixed by whole-fragment-as-one-section fallback (guarded against full documents). The content-regression guard (new text < existing/4) protects prose but returned errors that complete_error converted to success. Deferred sections' instances are dropped on save (carry-forward pending, cousin of the interactive clobber).
- **sources:** 016 §9 tool-pages entry + guard entry; 016b Part 5/regenerated-section entries
- **relations:** de-tool hazard fix layers; deployed→needs_rebuild flip
- **verify-later:** patched save_page_sections deployed to all three callers

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Presign loop collapse: batch adapter calls over awaited-loop iterations (O(K²) state bloat)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §9 marked DONE + CONFIRMED IN PROD 2026-06-09 (migration 110; 26s vs never-finishing)
- **what:** Every step transition re-persists the whole orchestration state (expanded workflow + collected_data + history), so a K-iteration awaited loop is O(K²) and geometrically slows; the structural fix is one batch adapter call (prepare_object_urls returning all URLs in one reply) — deleting both the race class and the bloat class. Related fix: configOrInput now coerces numeric config scalars (expiry_minutes 3000 was silently dropped by a .(string) assertion).
- **sources:** 016 §9 presign entries
- **relations:** loop mechanisms; envelope race
- **verify-later:** training-launcher def shape (2d state check)

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Hand-applied agent-def migrations have no ledger; re-running an earlier one reverts later ones
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §9 (2026-06-09): re-running 109 silently reverted 110+111; runbook corrected
- **what:** Agent-definition jsonb migrations are hand-applied with no runner/ledger — the live def SHAPE is the only source of truth. A migration is idempotent only vs its own prior application, never vs later migrations on the same object; recover from doubt by checking state, never by re-running. Per-migration state checks (runbook 2d) after every deploy.
- **sources:** 016 §9 re-running-idempotent-migration entry
- **relations:** backup discipline; deploy≠migration checklist item
- **verify-later:** —

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Image doesn't contain the binary (CrashLoop exec not-found ⇒ build/packaging fault)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §9 (2026-06-14): thunder-adapter tag shipped the analyser binary (overwritten Dockerfile, shared digest); third deploy-regression in a row
- **what:** `exec ./X: no such file` means the image lacks X — inspect image contents (docker run ls /app; Image-ID vs Image digest tells tag collisions), restore the Dockerfile, push a FRESH tag (never re-push the poisoned one). Guard: pre-push ls /app or a CI binary-name assertion — "no guard between built and running" is the recurring gap.
- **sources:** 016 §9 CrashLoop entry
- **relations:** deploy≠migration; stale-artifact family (checklist 24/26)
- **verify-later:** CI assertion added?

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Open problem: nav-updater never spawns
- **category:** debugging
- **status-signal:** unknown
- **status-evidence:** 103 "Active Problem" — definition exists/active, topics exist, dispatch generic, yet no pod ever appeared; all nav_drift items claim-timeout
- **what:** nav_drift items route to nav-updater via the generic dynamic dispatch, but no nav-updater pod has ever started and items exhaust claim timeouts. Investigation was open at handoff (2026-04-12); distinct from the nav-link-fixer path.
- **sources:** 103#Active Problem
- **relations:** dispatch loop; missing-handler pattern (different: def exists)
- **verify-later:** whether resolved since; nav_drift item outcomes

<!-- SOURCE: U01_docs024_numbered_core.md -->
### 016b durable invariants + wrong-turns log as a debugging methodology
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v1–v5 changelog; wrong-turns section explicitly kept "so the next pass doesn't re-walk them"
- **what:** Vol. 2 distils the paying-off heuristics (trust artefact; completed_at ≠ write instant; config-key-path no-ops; who writes page_components; 0 rows not decisive; negative inference needs mechanism checked in ALL cases; reuse before rebuild) and logs false leads per arc with the heuristic each violates. Also fixes doc process: the guide had forked across chats; v5 is the explicit merge point.
- **sources:** 016b#Orientation, #Durable invariants, #Wrong turns
- **relations:** 016 §0; travelling docs
- **verify-later:** —

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Zero-planned-sections silent no-op success (planning gap + complete_error anti-pattern)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016b v5 + 2026-07-06 amendment: route confirmed from workflow dump; pages.sections UPDATE proven to change behaviour; guards/planner-invariant fixes listed as prevention, not applied
- **what:** A linked-everywhere page 404'd for two weeks while seven work items completed clean: planner emitted the page with no sections; page-build-handler's zero-ready branch is literally a complete_workflow step named complete_error ("an error path implemented as a successful completion" — diagnostic signature: result contains only site_record); rerender skips no-component pages quietly. Section sources in order: site_specs site_plan aspect → pages.sections; site_plan_sections table is NOT read by builds. Fixes: planner invariant (every page ≥1 section), fail-loud zero-planned guard, rerender warn, auditor rules (active+linked+planned; post-deploy URL HEAD), dynamic-list component vocabulary for archive pages.
- **sources:** 016b#Page build completes having built nothing + amendment
- **relations:** silent-noop-success/planning-gap tags; section-index vocabulary
- **verify-later:** complete_error branch fixed?; planner invariant added?

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Content↔template key-contract drift (system-stats class)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016b Part 5 TRIAGED 2026-06-24, remedy un-applied; fleet-wide (usage_count 22)
- **what:** component-creator rewrote a template AFTER pages were built; stored content_data keys share ZERO keys with the new template placeholders → renders text-empty → visible-content filter correctly drops the section. Remedy: full content rebuild (not page_rerender, which reuses mis-keyed content_data); structural need: component schema changes must trigger dependent rebuilds, or fix writer↔input_schema binding. Diagnostic: diff the two key sets directly (a populated-but-unrendered section is a key-contract check, not a generation failure).
- **sources:** 016b Part 5 + wrong-turn #4
- **relations:** schema-template-drift tag; component regeneration rerender items
- **verify-later:** schema-change-triggers-rebuild mechanism

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Log tables before pod stdout (agent_error_log, llm_call_log as forensic sources)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 "hunting for logs" section; pod-rotation checklist item
- **what:** Persistent DB logs beat ephemeral pod stdout: agent_error_log (every reported error, filterable by context site/domain), llm_call_log (every call incl. failures with error_message). Pod logs vanish on rotation/rollout; zap JSON must be grepped by message string not field=value; logger.Debug is invisible in-cluster (house rule: logger.Info); verify deploys against the artifact (curl/DB), not log presence.
- **sources:** 016#hunting for logs; 016b#Verifying a deploy
- **relations:** silent-completion; assumption checklist 3/15
- **verify-later:** —

<!-- SOURCE: U01_docs024_numbered_core.md -->
### SQL template-surgery method (needle-gate) and Postgres verification pitfalls
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v4 entry distilled from the scheme-to-components workstreams
- **what:** Safe in-DB template mutation: needle-gate read (LIKE booleans + occurrence counts, expectations counted mechanically not recalled), shell .bak of the column, guarded idempotent exact-string replace (or anchored regexp_replace), RETURNING checks, value-agnostic rollback. Pitfalls: regex quantifier bound ≤255; substring-with-parens returns first capture group; gradient-embedded hexes escape colon-anchored classification; % in needles breaks LIKE gates (use position()).
- **sources:** 016b#SQL verification pitfalls
- **relations:** marker-REPLACE anchoring entry (anchor attribute REPLACEs on the opening tag, not the bare attribute — the querySelector corruption bug)
- **verify-later:** —

<!-- SOURCE: U01_docs024_numbered_core.md -->
### sites.status is informational (never scope by status='active')
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v4 entry with the silently-dropped-site incident
- **what:** UpdateSiteStatusAction vocabulary draft/building/review/published/deployed/archived/error; 'active' is legacy hand-written; nothing filters on it — dispatch keys on site_work_items. Enumerate GROUP BY status before any blast-radius query. Reuse-gate corollary: check pg_proc/pg_trigger before adding helpers (shared set_updated_at exists).
- **sources:** 016b#sites.status
- **relations:** zero-rows-not-decisive
- **verify-later:** —

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Kafka trigger payload discipline (flat single-line JSON here-strings)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "kcat heredoc was silently mis-routing messages to a 'No-op scheduled task' handler … Documented as permanent ops pattern in 016_debugging_guide_v2.md §9" (2026-04-23)
- **what:** Multi-line heredocs mangle kcat JSON payloads silently (routing falls through to no-op handlers with input_data null). Use `<<<'{…flat json…}'` here-strings or jq -nc. Related manual-trigger pattern: psql jsonb_build_object → pipe to kcat with standard headers, used to trigger handlers directly when dispatch is blocked.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4f v2 smoke retest, #14; FOCUS_dispatch_diagnostic(4).md#Workarounds
- **relations:** dispatch workarounds; debugging guide §9
- **verify-later:** 016_debugging_guide §9 entry

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Discovery-checks list maintenance and the workflow-replace landmine
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "Closed — investigation found no overwriter" (2026-04-19); the jsonb `||` append pattern recommended; updateAgentWorkflow risk "Currently safe because nothing fires it"
- **what:** The suspected "checks keep falling off discovery agents" was manual SQL replacing the whole checks array (a stale in-code example being copy-pasted); the safe pattern is jsonb array append. Latent risk logged: updateAgentWorkflow does jsonb_set of the ENTIRE workflow subtree — when an automated improvement-proposal generator ships, partial proposals will silently erase workflows unless converted to deep-merge.
- **sources:** HANDOFF_2026-04-19_component_linking_news_template_discovery_checks.md#3, #4
- **relations:** improvement_proposals (empty table); ApproveImprovementAction
- **verify-later:** updateAgentWorkflow (context line ~61056); stale comment cleanup

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Debugging meta-lessons (evidence discipline)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** codified into 016_debugging_guide entries: §0 item 19 added 2026-05-26; dispatch doc "Lesson learned" 2026-05-15; naming FOCUS "Tests document behaviour, not intent" (2026-05-17)
- **what:** Recurring investigation disciplines earned across these sessions: grep the whole codebase for the verb (triage/promote/claim) before concluding a writer doesn't exist; a LIKE on prompt_rendered proves what the model was told, never what it did — read response_text; check the guide before generating fresh hypotheses; design tests to falsify; tests assert what a function does, not what was intended; grep chassis logs by the `caller` field (msg gets truncated); logger.Debug is invisible in production; spawned pods are app=dynamic-agent with 600s idle timeout so capture logs before they evaporate; work the smallest useful step; trust suspicion of implausible numbers.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Lesson; HANDOFF_2026-05-26…md#wrong-turns; HANDOFF_2026-04-18_enrichment…md#greps, #false-starts; FOCUS_naming_conventions…md#flags; FOCUS_finetuning…(13).md#14
- **relations:** debugging guide 016 (the canonical home)
- **verify-later:** 016_debugging_guide §0 items

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Deployed-binary-predates-disk failure class
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Notes (Tm): "Fork RESOLVED: extraction sound … → the on-disk code cannot produce the July-2 escalation → deployed predates disk; the skip_field fix exists and never shipped."
- **what:** A named diagnosis class: observed behaviour contradicts a correct code read because the running pod's image predates the working copy — the fix exists on disk and never shipped. Diagnostic: `git log -1 -- <file>` vs the running pod's image age; remedy: deploy the working copy. Sibling lessons from the same threads: verify the running image contains an edit before debugging it ("success path silent"), and prefer a forward test (one clean build + read both render and stored data) over forensic reconstruction of overlapping rebuild windows.
- **sources:** running_notes_scheme_to_components(55).md#Tl #Tm #Tt; RUNBOOK_scheme_to_components(50).md#W7-FINDINGS; w8_07_fresh_index_build.sql (the forward-probe pattern)
- **relations:** chassis deploy model; plan_sections deferral (the instance).
- **verify-later:** 016b guide entry for this class.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Postgres/SQL pitfall class (016b lessons)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Notes (St): "Guide updated → `016b_debugging_guide_7.md` (v4 log + three §9 entries …)"; each pitfall has an owned in-thread instance.
- **what:** The accumulated instrument-error catalogue this thread wrote into the debugging guide: Postgres ARE regex quantifier bounds max at 255 (`.{0,420}` is invalid — use substr+position); `substring(… from '(pattern)')` returns the FIRST CAPTURE GROUP, not the match; LIKE treats a needle's literal `%` as a wildcard (use position()); regexes like `background:\s*#` miss gradient-embedded hexes; a `0 rows` result is not decisive until the query and live state are checked (applies to one's own verification queries too); probes that grep for a key string are blind when objects are UUID-named; naive brace-counting false-fails on regex literals. Plus data-vocabulary lessons: sites.status vocabulary is draft/building/review/published/deployed/archived/error with legacy 'active'/'system' strays — never filter blast radius on status='active'.
- **sources:** running_notes_scheme_to_components(55).md#Sr #Ss #St #Sv #Sw #Tu #Ue; w2_02_verify_fixed.sql; RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS (3f)
- **relations:** SQL needle-gate surgery; debugging guide 016b (the home doc).
- **verify-later:** 016b_debugging_guide_7.md §9 entries.

<!-- SOURCE: U04_idea_uk.md -->
### Claimed-item timeout evidence gate (failed vs false-completed)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "the evidence-gate (migration_claimed_item_timeout_evidence_v2, live since 2026-06-04) refuses to complete a 0-component page… That is the gate working as intended."
- **what:** The dispatch gate that distinguishes the two failure signatures of the same coordinator bug: without it a stubbed page false-completes (gamesdesign); with it, a 0-component page's claim is reset and retried until attempts exhaust → an honest `failed`. Used here as diagnostic doctrine: don't conflate a silent stub with a genuine handler hang — read the parent's collected_data response to tell them apart.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (empty-index diagnosis); idea.uk/016_debugging_guide_v2_32(1).md (claimed-item sections)
- **relations:** coordinator result contract; work-item state machine.
- **verify-later:** the evidence-check migration in the chassis migrations.

<!-- SOURCE: U04_idea_uk.md -->
### LLM API shape disciplines (server-tool injection, per-model thinking shapes, long-call timeouts)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Items 24–27 added to the debugging guide from the 2026-06-04 validation run ("three API bugs found and fixed during validation").
- **what:** Standing disciplines from live API breakage: (24) a hosted server tool may auto-inject its documented dependency — declaring it yourself collides (web_search v2 injects code_execution; the 400 names the conflict); (25) the same capability has different wire formats across model generations — newer Opus-class models take adaptive thinking + output_config.effort while Sonnet 4.6 takes manual budget_tokens, so helpers must branch per model (and Opus also rejects non-default temperature/top_p); (26) long agentic calls (high effort + N searches) send no headers for minutes — size client timeouts for the worst-case step (180s→900s; streaming is the durable answer); (27) always confirm the current request shape from live docs before coding, especially after a model bump — remembered shapes are guesses and each failed round-trip costs real spend.
- **sources:** idea.uk/016_debugging_guide_v2_32(1).md (items 24–27); idea.uk/DEVELOPMENT_RUNBOOK(3).md#A1 (acceptance); idea.uk/running_notes(63).md (2026-06-04 checkpoints)
- **relations:** engine upgrade; llm-quality-testing; model-infrastructure.
- **verify-later:** engine.go usesAdaptiveThinking + client timeout.

<!-- SOURCE: U05_content_quality_linking.md -->
### Operator/assistant division-of-labour + DB-change safety conventions
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Standing-rules blocks repeated verbatim across HANDOFF_2026-06-09(2), HANDOFF_2026-06-15(2) §0, HANDOFF_page_pipeline(11) §0-1.
- **what:** The working covenant these threads ran under: the assistant reads code/writes deliverables with no cluster/DB access; the human runs all SQL/kubectl/builds. Safety conventions: snapshot before any DB change (snapshot_agent()/revert_agent() house helpers for agent rows; `CREATE TABLE <t>_bak_<tag> AS SELECT` for data, short names, in-txn, rollback documented); fresh `\d` before SQL; every template replace() verified by UPDATE 1 + a flag flip (a stuck flag = whitespace mismatch silently no-op'd); check_linking_sql_applied.sql as the idempotent "which SQLs are in" orient step; workflow changes are DB-only and immediate vs Go changes needing an image roll + image_tag bump; tags of co-deployed agents must move together (a lagging resolver tag = permanent silent fallback); don't roll the chassis image while a rebuild batch drains.
- **sources:** HANDOFF_2026-06-15(2).md#0; HANDOFF_page_pipeline(11).md#0-1; check_linking_sql_applied.sql; RUNBOOK_linking_phantom_fixes(7).md
- **relations:** debugging heuristics; documentation system (runbook discipline).
- **verify-later:** snapshot_agent/revert_agent function defs; agent_backups behaviour.

<!-- SOURCE: U05_content_quality_linking.md -->
### Debugging heuristics harvested into the 016 guide
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Guide bumped v2_31 → v2_45 → v2_49/50 → v2_56/57 across this unit's sessions, each with named new entries; 016 "CLOSED at v2_56", continued in 016b.
- **what:** The unit's investigations were systematically distilled into durable heuristics: trust rendered HTML/DB state over work-item status; a 0-rows/false result is not decisive until the query itself is verified; matching updated_at is not authorship — confirm the action writes the column; work-item completed_at is the orchestration END, not the write instant — trace by orchestration; an empty LEFT JOIN means wrong anchor id; never carry a site_id across a teardown; "git committed ≠ new content"; two rebuild routes; text-heuristic blind spots (prose guards miss markup/JS loss); psql vs shell variable syntax traps; Kafka consumer-group loss after topic wipes (restart-to-rejoin, park at latest, never replay).
- **sources:** running_notes_14(26).md#principles sections; running_notes_17(21).md; NOTES(44) passim; HANDOFF_page_pipeline(11).md#10
- **relations:** documentation-system (guide versioning); every defect thread here.
- **verify-later:** 016_debugging_guide_v2_56/016b content (owned by the debugging unit).

<!-- SOURCE: U06_finetuning.md -->
### Wrong-binary adapter image incident and the built-vs-running guard
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-14(8): "thunder-adapter:v1.0.1063 actually contains the analyser-adapter binary… Pattern (third deploy-regression in a row)… No guard between built and running. Logged in debug guide v2_47."
- **what:** An overwritten Dockerfile shipped the analyser-adapter binary under the thunder-adapter tag; the pod CrashLoopBackOff'd for ~31h and every provision parked runs at `pending`. Named as the third consecutive "the deploy didn't ship what I thought" regression (109 re-run revert; chassis/adapter tag confusion; Dockerfile overwrite). Prescribed guards: per-build `docker run --rm --entrypoint ls <image> -la /app` before push, never re-push a poisoned tag, and structurally a CI step failing the build if the expected binary is absent — the deploy-side sibling of the migration 2d state check.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-14-8; working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#2026-06-14-update
- **relations:** hand-applied migrations lesson; deployment essentials
- **verify-later:** whether the CI image-content guard was ever added

<!-- SOURCE: U06_finetuning.md -->
### Send-before-register await race and preRegisterAwaitedRequest
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-09: "Race fix works [verified-log]. Every presign_checkpoints_iter_N_presign_one logged ClaimAwaitedRequest: status_before=waiting … claimed:true"; recorded as the fourth cause of stuck-`waiting` awaits in debugging guide v2_36 §9.
- **what:** Local dispatch actions produced the adapter request and returned await_response:true BEFORE the coordinator inserted the awaited_requests row; a fast (~1s) reply beat the insert, ClaimAwaitedRequest found no `waiting` row, the reply was dropped, and the timeout handler re-dispatched forever with fresh request_ids (RetryVersion pinned at 0). spawn_agent/call_agent don't race because they call `preRegisterAwaitedRequest` (register-before-send, ON CONFLICT DO NOTHING). Fix: the dispatch pre-registers with the same request_id it uses everywhere — one row, one timeout owner; caveats: the helper hardcodes a 120s timeout that wins over step config, and the per-request timeout goroutine is skipped (background expiry sweep is the net). Moving stall point ⇒ race is the diagnostic heuristic.
- **sources:** working/phase5/HANDOFF_2026-06-06_checkpoint_upload_loop_await_race(2).md; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-08,#update-2026-06-09; working/docubundle/CONTEXT_PACK_thunder_checkpoint_race(1).md
- **relations:** O(K²) loop cost (found immediately after); reply-topic rules; awaited_requests machinery
- **verify-later:** preRegisterAwaitedRequest call in thunder_prepare_object_url_dispatch.go and the batch/resume dispatches

<!-- SOURCE: U06_finetuning.md -->
### O(K²) loop state-bloat and the batch-presign replacement
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-09(3): "Full launcher path completed in ~26s… ONE batch await… returned all 40 ckpt PUT URLs… Contrast the retired loop: Version 86 / still at iter_9 nine minutes in. The O(K²) class is gone."
- **what:** Every awaited loop substep re-persists the full orchestration state — the expanded ~80-substep workflow with verbose descriptions, growing collected_data, and ProcessingHistory — so a 40-iteration awaited loop costs O(K²) (iter_0-4 ~2-3s, iter_8 ~100s, then Kafka i/o timeouts) while a GPU bills throughout. Structural cure, not tuning: replace the per-item awaited loop with one batch adapter call (`prepare_object_urls`: keys[]→ordered urls[], reusing the single ObjectURL primitive per key), one await, one persist, no flatten step (migration 110). General platform lesson: awaited loops over cheap local operations are an anti-pattern; batch at the adapter.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-09—(3); working/phase5/110_training_launcher_batch_presign(2).sql (header); working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#2026-06-09-update
- **relations:** send-before-register race; loop_complete convention (every production loop ends on an explicit loop_complete substep — checked against all 11 production loops); durability manifest
- **verify-later:** orchestration state persistence cost in coordinator; whether other awaited loops exist at risk

<!-- SOURCE: U06_finetuning.md -->
### Hand-applied agent-def migrations: no ledger, re-run reverts, 2d state check
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-09(7): "re-running 109 silently REVERTED both [110 and 111]… A migration is idempotent only against its OWN prior application, never against LATER migrations that mutate the same object… There is NO migration runner."
- **what:** The flywheel-C def migrations (102–111) are hand-applied jsonb mutations to agent_definitions with no schema_migrations ledger — the def's live shape is the only "did it run" truth. Consequences codified: never re-run an earlier migration "to make sure" (it reverts later ones); run a per-migration state-check query (RUNBOOK 2d) after every deploy and before any launch; back up defs with the sanctioned `snapshot_agent()`/`revert_agent()` (hand-rolled CREATE TABLE backups collide with the existing agent_definitions_backup — discover DB helpers with `\df` first). Optional future hardening: a migration runner or applied_migrations log.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-7,(3); working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#2026-06-14-update
- **relations:** wrong-binary incident (same "shipped what I thought?" family); model swap/revert functions (snapshot_agent reuse)
- **verify-later:** RUNBOOK 2d query vs live launcher def

<!-- SOURCE: U06_finetuning.md -->
### agent_definitions source-of-truth: clients_db, not templates_db (for the rich schema)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-03 ~17:3x CORRECTION: "templates_db.agent_definitions has the OLD schema… holds only the 8 original website-builder agents… PIN (corrected): for the flywheel-C agent_definitions, always read AND patch clients_db."
- **what:** agent_definitions exists physically in BOTH clients_db and templates_db; the architecture doc's "source of truth is templates_db" refers only to the legacy website-builder catalog (old schema, no version column). The chassis loader (filters is_active/is_snapshot, ORDER BY version) can only run against clients_db's rich schema — so all flywheel-C and modern defs live there. This whipsawed twice in one day (103 first applied to the wrong DB, then the "always templates_db" pin issued and then reversed) — a live example of doc-claims diverging from code, and of why the clients_db copy of one def can silently diverge from the live one.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-03; working/phase5/103_call_data_preparer_optional_inputs.sql (header carries the superseded templates_db guidance); working/phase5/104_provisioner_output_fields_and_launcher_mapping.sql (header carries the correction)
- **relations:** hand-applied migrations; documentation-system (stale doc line in 002_system_architecture.md)
- **verify-later:** chassis definition-loader query; 002_system_architecture.md wording

<!-- SOURCE: U06_finetuning.md -->
### CLI/ops data-transfer pitfalls (kcat heredoc, COPY-vs-psql, kubectl exec/cp)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** PATCH_2026-05-06 (both bugs validated + corrected command); FOCUS(25) §2.4f v2 smoke retest (kcat heredoc mis-routing); HANDOFF 2026-04-23 lesson 7.
- **what:** A cluster of verified transfer traps: (1) Kafka trigger JSON must be flat single-line via here-string — multi-line kcat heredocs mangle payloads silently and route to a No-op handler; (2) `COPY … TO STDOUT` is not JSON-safe for jsonb (double escape layers) — use `psql -tAXc` with plain SELECT for JSONL; (3) `kubectl exec -i` without consumed stdin sporadically truncates stdout (1716/1958 rows, "next reader: unexpected EOF"); (4) `kubectl cp` truncates large files silently — use `exec cat > local`; (5) `tnr scp` of directories nests `{dest}/{source_basename}/` both ways.
- **sources:** working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted_PATCH_2026-05-06.md; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#14; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons(4)
- **relations:** dataset pull path; 016 debugging guide §9
- **verify-later:** 01_pull_dataset_from_postgres.sh uses the corrected form

<!-- SOURCE: U06_finetuning.md -->
### configOrInput numeric config coercion (expiry_minutes silently dropped)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-09(5): "expiry_minutes override — FIXED… configOrInput read config via Config[name].(string), so the JSON-number 3000 failed the assertion → fell through → adapter default". Debug guide v2_43.
- **what:** The shared configOrInput helper type-asserted config values to string, so JSON-number config (expiry_minutes:3000, timeout_seconds) silently fell through to defaults — presigned PUTs came back at 24h instead of 50h. Fixed with a `coerceConfigScalar` (string/float64/json.Number/int/bool). Class lesson: shared config readers must coerce scalars, and a numeric setting "applied" in a def is only proven by observing the effect (X-Amz-Expires on the URL).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-4,(5)
- **relations:** presigned data plane expiry caveat; launcher dispatch family
- **verify-later:** coerceConfigScalar in thunder_ssh_exec_dispatch.go

<!-- SOURCE: U06_finetuning.md -->
### Scheduler-fired chassis-resident observability gotcha (owner_agent_type='generic')
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** STATUS 05-12(1) architectural follow-ups: "filtering orchestration_states by owner_agent_type MISSES top-level chassis-resident workflows, which are owned by 'generic'".
- **what:** Scheduler-fired agents that run in the generic chassis (thunder-reaper, build-pipeline-trigger, etc.) have orchestration_states.owner_agent_type='generic'; the real agent type lives at `collected_data->'config'->>'agent_type'` and orchestration_name follows `sched-<task>-<ts>`. Filter on those instead. Related cosmetic anomaly, unresolved: a stale non-DB agent_config stub (old reaper-style no-op) persists in message envelopes across redeploys while the full WorkflowPlan executes — source of the cached representation never found.
- **sources:** working/flywheel_docs/STATUS_thunder_adapter_2026-05-12(1).md#6; working/phase5/NOTES_phase5_training_launcher_running(45).md#stub-source-narrowed
- **relations:** monitor testing; debugging guide
- **verify-later:** where the stale agent_config envelope field loads from

<!-- SOURCE: U06_finetuning.md -->
### Kafka topic-creation race self-heal (transient "Topic not yet on broker")
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-06: "Transient `Topic not yet on broker` for the launcher .responses topic self-healed on attempt 2 (topic-creation race) — normal."
- **what:** Per-spawn child topics (`job.<id>.requests`, per-agent responses topics) are created on demand; a first-publish race against broker propagation produces a transient failure that retries resolve. Recorded so it isn't chased as a real fault. Contrast: a *permanently* missing topic (Strimzi auto-create off) fails every attempt — the distinguishing signature is self-heal on retry.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-06; working/flywheel_docs/FOCUS_adapter_design(3).md#required-cluster-resources
- **relations:** adapter deployment essentials (KafkaTopic CRDs)
- **verify-later:** topic auto-creation settings for spawned-agent topics vs adapter topics

## Proposed NEW categories

None. Everything in this unit fits existing seed slugs — predominantly `finetuning-flywheel`, with `model-infrastructure` (Thunder/Ollama/endpoint/swap), `adapters`, `storage-architecture` (presign/credential boundary), `development-guide`/`debugging` (chassis contracts and failure signatures), `business-strategy` (finetuning.uk), `diagnosis-loop` (docubundle), and `documentation-system` (epistemic tagging).

## Cross-cutting flags for stage 2

- Hardcoded Thunder API bearer token committed in `working/flywheel_docs/ssh_probe.sh` — credential hygiene check.
- Persistent open items to verify: monitor schedule enabled?; first RUN_SH_DONE + final adapter in B2?; orphan-sweep built?; model-trainer call_agent fall-through fixed?; validator Tier-2 coverage extended?; iter_1 ever trained (fp16, 2-epoch, `<no value>` filter)?; any production model swap executed?

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Loose dispatch item-status semantics (complete ≠ done)
- **category:** debugging
- **status-signal:** aspirational
- **status-evidence:** "loose dispatch item-status semantics documented across the investigation (complete-at-dispatch, errors-in-complete, status-change without a timestamp bump, parent-topic-vanished noise) — worth a pass when convenient" (RUNBOOK(49) Part E Hygiene); seven dated sightings in NOTES.
- **what:** A documented defect class in the dispatch loop's work-item bookkeeping, observed seven times: items marked 'complete' at dispatch while the child orchestration runs or fails later; the child's full error text stored in the `error` column of a 'complete' item; status transitions that don't bump updated_at; batch claim stamps shared across differently-fated items; parent fire-and-forget topic lifecycle polluting child completions ("topic partition not found"). Operational rule derived: never trust item status as proof of work — verify the artefact (band stamp, render md5); agent_error_log (occurred_at) outranks status. Fix parked as hygiene.
- **sources:** NOTES(43).md §9i, §9l, §9m, §9aa, §9ac, §9ax, §9bd; RUNBOOK(49).md Step 9 reading guide + Part E
- **relations:** work-item dedup; F2 methodology (discriminator ordering); auto-escalation.
- **verify-later:** build-dispatch-loop status handling; whether items get failure statuses on child errors.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F2 tiered guard-verification methodology (unit → integration → live keep/reject fixtures)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "F2 COMPLETE: Tier 1 unit ✔, 3a preservation ✔ (×2 regens, md5-verified template change), 3b reject ✔ (live firing, three-level visibility, zero mutation)" (NOTES §9o).
- **what:** The verification pattern used to prove F1 without touching live shared components: Tier 1 deterministic unit tests of the guard logic (including the real incident's rename case); Tier 2 DB-backed reject-path test (folded into Tier 3 when no harness existed); Tier 3 end-to-end on throwaway zzz-* components — a KEEP fixture proving preservation-by-instruction (non-guessable check: template md5 changes while fields hold) and an intentionally INACTIVE REJECT fixture exploiting the store-vs-loader is_active divergence to force a rename and observe the guard fire live with zero mutation. Also codified the discriminator ordering (agent_error_log > pod logs > never item status) and prompt cleanup of leftover fixtures.
- **sources:** NOTES(43).md §9f, §9h, §9k–§9o; RUNBOOK(30) family (Step F2 tiers)
- **relations:** F1 guard; F4 (discovered by 3b run 1); loose status semantics.
- **verify-later:** store_generated_component_guard_test.go; zzz fixtures fully cleaned.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Needle-gate SQL template surgery pattern (and its catalogued pitfalls)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Applied on every template mutation W1–W3e with gates, RETURNING checks and verify regions; pitfalls promoted into 016b guide v4 ("count expectations mechanically from the dump… never from memory").
- **what:** The method for mutating shared templates/configs safely by SQL: dump + shell backup first; a gate query asserting exact needles (booleans) and mechanical occurrence counts derived by grep from the dump (mismatch = drift OR mis-derived expectation — stop); anchored exact-string or backreference replaces (multi-line needles to disambiguate repeated strings); guards for idempotency; RETURNING post-conditions; separate verify file; value-agnostic rollback file. Catalogued Postgres pitfalls: regex quantifier bound ≤255; substring() returns the first capture group; LIKE-wildcard `%` inside needles (use position()); `\set ON_ERROR_STOP on` when statements depend on earlier ones; run SQL as files, never pasted.
- **sources:** RUNBOOK_scheme_to_components(18).md W1–W3e blocks + RESULTS; running_notes(22).md Sr, Sv, St
- **relations:** prompt-migration convention (same family for jsonb); debugging guide 016b (where the lessons were codified).
- **verify-later:** 016b guide entries; w*_*.sql files referenced (outside unit).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### R6c artifact-forensics method: cache-busted, metric-consistent comparisons
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "md5sum: gd.html == gd2.html… ONE artifact all along. OWNED: my stale-cache story AND the earlier '4-of-8 mis-assembled' reading were metric artifacts" (NOTES §9al).
- **what:** Lessons from the gripper-detail "blank page" false trail: compare live artifacts only with identical metrics (a data-component inventory vs a class grep counted different things and manufactured a mis-assembly story); md5 the fetches before concluding stale-cache; distinguish 404/200-empty/200-styled-invisible with curl size + head; visually-blank ≠ missing content (fallback-vars insight: content present but dark-on-dark). The eventual truth (theming, not assembly or deploy) reshaped Part D.
- **sources:** NOTES(43).md §9af–§9al; RUNBOOK(49).md Part B
- **relations:** R6f (the real mechanism); assembly membership model; needle-gate pattern (same mechanical-counting ethos).
- **verify-later:** n/a (method; instances cited).

<!-- SOURCE: U08_travelling_docs.md -->
### max_tokens placement rule — dead config outside ai_service
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** FYI second addendum 2026-07-10: "verdict step max_tokens was DEAD CONFIG … Fixed on both agents (snapshot first)".
- **what:** `execute_llm_prompt` reads `max_tokens` from the agent's top-level config or from INSIDE the step's `ai_service` block — never from the step-config root, where several agents had it; the Anthropic client then silently defaults to 2048 output tokens. A truncated verdict JSON parses to UNVERIFIABLE. Standing grep: `config.max_tokens` outside `ai_service` is dead wherever execute_llm_prompt is the action.
- **sources:** FYI_from_fixloop_2026-07-10_verdict_prompt_symptom_check.md#second-addendum
- **relations:** silent-no-op config-path heuristic (016b durable invariants); execute_llm_prompt shared action.
- **verify-later:** ai_actions.go:252-256 max_tokens resolution; remaining workflows with root-level max_tokens.

<!-- SOURCE: U08_travelling_docs.md -->
### agent_error_log is the FIRST read
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v6 entry (2026-07-08), promoted after it settled the tool-generation blocker in one query.
- **what:** Step failures persist to `agent_error_log` (orchestration_id — TEXT not uuid — step_name, action, error_message, error_code, context) and outlive the pod. Read it first, filtered by orchestration_id; only then pod logs (may be reaped) or collected_data (may be enormous). `current_step` from polling is a sample, not an attribution (a 120s poll blamed the LLM step when save_tool failed); a terminal step's success_message can name the wrong phase.
- **sources:** 016b_debugging_guide_7_3_(7).md#agent-error-log-entry; HANDOFF_2026-07-08…md#§3,§5; RUNNING_NOTES_travelling_docs(39).md#rev29
- **relations:** schema drift incident; two failure envelopes; 0-rows rule.
- **verify-later:** agent_error_log schema (orchestration_id type).

<!-- SOURCE: U08_travelling_docs.md -->
### Code-ahead-of-DB schema drift (SQLSTATE 42703, latent until first caller)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Root-caused 2026-07-08 (create_tool_component vs missing content_components provenance columns, latent ~2 months since 2026-05-16); fix applied + proven 2026-07-09.
- **what:** A binary referencing new columns deployed before its migration; nothing fails until the rare code path is called. Detection: the failing INSERT's own comment names the missing migration; a last-successful-call latency probe distinguishes latent drift from fresh regression. Fix pattern: MIRROR column types dynamically from the table the code says it mirrors (format_type/pg_attribute + ADD COLUMN IF NOT EXISTS), additive/nullable/idempotent. The canonical migration file existed but was parked in a docs folder, never renumbered into the migrations path — the exact mechanism by which a deploy skips a migration (one motivation for the migrations system). Standing pre-deploy check: grep the diff for new column names and assert each exists in production.
- **sources:** 016b_debugging_guide_7_3_(7).md#schema-drift-entry; HANDOFF_2026-07-08…md#§3; RUNNING_NOTES_travelling_docs(39).md#rev29,#rev30
- **relations:** migrations system; content_components provenance columns (migration 133); "provenance stamps the chassis".
- **verify-later:** sql_for_agents/133_add_component_provenance.sql vs the docs019 design copy.

<!-- SOURCE: U08_travelling_docs.md -->
### Prompt-template vs config-path resolvers (TEMPLATE_FIELD_ERROR)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v7 entry; root cause of the missing first auto-PLAN (rev 32); latent same-class bug in Task-4 templates caught before it fired.
- **what:** `execute_llm_prompt` with `output_format: text` hands the prompt template the BARE string (`{{.X}}`); with `json` a map (live form `{{.X.result | toJSON}}`); action CONFIG field paths are a different resolver and keep `.result`. Never reach an unverified nested key from a template — dump whole objects with `| toJSON`. A render-time error fires before tokens are spent and, with error containment, the workflow "succeeds" while the step's product is missing (reading rule: normal terminal + missing downstream artefact = contained step failure).
- **sources:** 016b_debugging_guide_7_3_(7).md#template-entry; RUNNING_NOTES_travelling_docs(39).md#rev32; HANDOFF_2026-07-09_recreation_and_chassis_1_.md#§2
- **relations:** docs-never-fail containment (masking effect); seam rule.
- **verify-later:** template data shaping in ai_actions.go by output_format.

<!-- SOURCE: U08_travelling_docs.md -->
### EXECUTING_STEP forever = the worker died (OOMKill triage), superseding stall/leak readings
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v8 rewrite 2026-07-09; the earlier v5-era "error containment does not protect against a HANG" entry and the slow-leak hypothesis are explicitly superseded on the evidence trail kept in RUNBOOK.
- **what:** `orchestration_states` is written BY the worker: a dead pod (OOMKill exit 137, eviction, panic) writes nothing, so the row freezes at EXECUTING_STEP and `since_s` measures time since the crash. Triage order: RESTARTS column → describe pod Last State → `logs --previous` (capture crash logs IMMEDIATELY — a ReplicaSet replacement erases them). Probe suspected-stalled dependencies with a bound (`curl -m 5`) before assuming a hang. Related-but-distinct: genuine stalls from missing context deadlines deserve fixing as hygiene. The arc walked through three wrong hypotheses (stall → missing deadline → slow leak) before the real cause (chunkContent loop), each correction documented rather than discarded.
- **sources:** 016b_debugging_guide_7_3_(7).md#executing-step-entry; RUNBOOK_travelling_docs(38).md#superseded-incident-block; RUNNING_NOTES_travelling_docs(39).md#rev34,#rev35,#rev36
- **relations:** chunkContent bug (the answer); containment-limit corollary.
- **verify-later:** n/a (operational pattern).

<!-- SOURCE: U08_travelling_docs.md -->
### chunkContent() infinite loop — the OOM root cause, fixed with timeout regression tests
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "THE OOM ANSWER (closes the incident chain for good)" — confirmed 2026-07-10; fix deployed v1.0.1104; proof run 05d1fc97 with 0 pod restarts.
- **what:** `chunkContent()` in rag_actions.go never terminated on content > chunk_size: the final chunk ends at len(content), `start = end - overlap` steps BACKWARDS, the same tail appends forever → 2Gi in seconds; content ≤ 1000 chars returned early, hiding the bug for weeks (both OOMKills were PLAN-sized bodies through index_plan). Fixed with a final-chunk break + forward-progress guard and four regression tests with a 30s timeout that catches loop regressions. Durable class rule: content-below-threshold early returns can hide a non-terminating path; "a proof run is a probe — fire proofs early" (the 139 proof run found the real cause within the hour).
- **sources:** RUNBOOK_travelling_docs(38).md#task-6; RUNNING_NOTES_travelling_docs(39).md#v1.0.1103-proof-run,#fix-140-141; HANDOFF_2026-07-10…md#§1,§4
- **relations:** tool_docs indexing (unblocked); migrations 140/141; EXECUTING_STEP pattern.
- **verify-later:** rag_actions_chunk_test.go; chunkContent forward-progress guard.

<!-- SOURCE: U08_travelling_docs.md -->
### kcat -P is line-delimited — single-line trigger bodies
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Run 464102f4 post-mortem (rev 45); 086 and 087 scripts enforce single-line bodies.
- **what:** A pretty-printed JSON body piped to `kcat -P` becomes one message per line; the chassis can then marry your headers to a NEIGHBOURING message's body (observed: our correlation id completing "after 0 steps" holding a scheduler no-op's body — also flagged a chassis stale-buffer wrinkle worth a look). Trigger bodies must be compacted to a single line and scripts must refuse multi-line.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev45-run1; RUNBOOK_travelling_docs(38).md#new-durable-rules; 086/087 script headers
- **relations:** manual kcat trigger scripts; env-prefix trap (sibling).
- **verify-later:** the stale-buffer wrinkle in the chassis consumer (never followed up).

<!-- SOURCE: U08_travelling_docs.md -->
### Env-prefix trap — VAR=x on its own line (or with `;`) never reaches the child
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Cost two 3b.4 runs and one 085 run; 084/086 banners hardened to print the go/no-go tell ("Subject: NONE — will SKIP").
- **what:** Shell variables set on their own line (or terminated by `;` before the command) are not exported to child processes, so triggers silently run with defaults. Correct forms: same-line prefix or `export`. Scripts now print explicit banners of the effective values as the load-bearing tell.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev19,#rev33; RUNBOOK_travelling_docs(38).md#§8
- **relations:** trigger scripts; banner-tell convention.
- **verify-later:** n/a (operational pattern).

<!-- SOURCE: U08_travelling_docs.md -->
### Two failure envelopes — parent COMPLETED ≠ child success
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v5 entry; observed live in runs 1–2 of the 3a arc.
- **what:** A mid-run child failure returns header `status: complete` with the failure in the BODY (`body.status: failed`) — the parent forwards it and completes (a COMPLETED parent row with non-empty `error` = a forwarded child failure); a failed-to-START child sends `status: error_unrecoverable` / `CHILD_ORCHESTRATION_FAILED`. Consumers must check the body, never the header alone; which shape appears tells WHERE the child died.
- **sources:** 016b_debugging_guide_7_3_(7).md#failure-envelopes-entry; RUNNING_NOTES_travelling_docs(39).md#rev11,#rev12
- **relations:** agent_error_log first read; §0-REF reading rules.
- **verify-later:** sendWorkflowResponse / notifyParentOfFailure paths.

<!-- SOURCE: U08_travelling_docs.md -->
### Pod label `agent-type` (hyphen) + multi-pod log attribution
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v5 entry; proven by a working command vs a zero-match selector.
- **what:** The pod LABEL key is `agent-type` (hyphen) while log JSON fields say `agent_type` (underscore) — the underscore selector silently matches zero pods. A type-wide selector spans ALL live pods (idle reaper 3600s), so tails contain residue from earlier runs: attribute every line by orchestration id / pod / timestamp before reading it as current.
- **sources:** 016b_debugging_guide_7_3_(7).md#label-entry; RUNNING_NOTES_travelling_docs(39).md#rev11,#rev13
- **relations:** 0-rows rule; §0-REF.
- **verify-later:** n/a (operational pattern).

<!-- SOURCE: U08_travelling_docs.md -->
### 0-rows rule + gate-evidence capture window + state-dump substitute
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Run-3 closed via the state-dump substitute 2026-07-06; codified in 016b anchorless entry and RUNBOOK §7.
- **what:** 0 rows is decisive only after the query itself AND the run's completion are ruled in (a run that died upstream also produces 0 rows). When a step's NON-firing is the success condition, closure needs a COMPLETED child + the step's explicit skip log line + the 0-count. Skip log lines have a 3600s capture window (idle reaper); past it, a post-completion state dump (ProcessingHistory showing the step executed + terminal status + 0-count) is the accepted substitute. Placeholders are replaced INCLUDING the angle brackets.
- **sources:** 016b_debugging_guide_7_3_(7).md#anchorless-entry (verification discipline); RUNBOOK_travelling_docs(38).md#§7,#stage-3; RUNNING_NOTES_travelling_docs(39).md#rev16
- **relations:** persist_diagnosis_note gate proof; agent_error_log.
- **verify-later:** idle-reaper timeout value (3600s).

<!-- SOURCE: U08_travelling_docs.md -->
### Postgres guard-writing gotchas — RE_DUP_MAX 255, sticky aborted transactions, psql -f over paste
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b entry written from the supersede-migration attempts 1–3 (2026-07-09).
- **what:** Postgres ARE caps bounded regex repetition `{m,n}` at 255 — prefer `strpos`/`substr` in guards (plainer intent, no engine limit). After any in-transaction error the session is stuck (`clients_db=!#`) and ignores everything including BEGIN — `ROLLBACK;` is the only exit; migration files should open with a defensive ROLLBACK and be run with `psql -f`/`\i` (pasting mangles comments and dollar-quoted bodies). A guard that refuses a write can be RIGHT (it blocked an unverified selector) or WRONG (it refused a valid runtime-built selector) — guard design evolved to accept static OR dynamic evidence with a NOTICE saying which path verified.
- **sources:** 016b_debugging_guide_7_3_(7).md#postgres-regex-entry; RUNNING_NOTES_travelling_docs(39).md#rev37,#rev38,#rev39; 0NN_supersede_xp_curve_plan_selectors(2).sql
- **relations:** needle-gate template-surgery pattern; anchor rule (the design insight that came out of guard 1's refusal).
- **verify-later:** n/a (operational pattern).

<!-- SOURCE: U08_travelling_docs.md -->
### Untracked-file deploy trap — verify deploys by ancestry, not by tag or commit message
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Struck TWICE (Tier-2 checker missed two releases; check_tool_acceptance_due missed v1.0.1111); rules banked in HANDOFF T11 and the durable-rules list.
- **what:** `git commit -a` commits modified-tracked files only — an untracked (`??`) new file silently misses any number of release commits while its sibling changes ship. Guards: `git status` for `??` before every release (or commit new files as written); verify a deploy carries your files by ANCESTRY (`git merge-base --is-ancestor <commit> <release>`); this repo also reuses version tags, so pod-start-time vs commit-time settles what a tag actually contains, not the commit message. Safe-failure companion: unknown discovery-check names warn+skip (the 142 precedent), so wiring a check by migration before its binary deploys is safe.
- **sources:** HANDOFF_2026-07-10…md#T8,T11,#§4; RUNNING_NOTES_travelling_docs(39).md#stage-5-live,#v1.0.1111; README_summary_paragraph2_for_discussion.md
- **relations:** continuous sweep gate; migrations-before-binary safety.
- **verify-later:** n/a (operational pattern).

<!-- SOURCE: U08_travelling_docs.md -->
### Page build/rerender failure-shape thread family (Parts 1–5 + wrong-turns log)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016b open-threads status header: Part 1 DONE/verified; Part 2 partially verified; Part 3 code prepared not applied; Part 4 written un-deployed; Part 5 triaged.
- **what:** A connected series on "work that reports success but doesn't happen": result-contract drop replaced child output with a success stub (fixed, shipped 2026-06-18); no-LLM re-render pre-pass (partially verified); item_key canonicalization drift (needs_page vs needs_tool_recreation colliding on the dedup index — builder prepared); the interactive clobber (above); system-stats dropped because content_data and the component template share ZERO keys (a content↔template key-contract mismatch — the visible-content filter was correct). The companion "Wrong turns" log records false leads with the durable heuristic each violated — a deliberate documentation convention so the next pass doesn't re-walk them.
- **sources:** 016b_debugging_guide_7_3_(7).md#open-threads,#wrong-turns; (fix detail lives in the gamesdesign/scheme runbooks outside this unit)
- **relations:** silent-completion invariants below; travelling copies of 016b.
- **verify-later:** current state of Parts 2/3/5 fixes.

<!-- SOURCE: U08_travelling_docs.md -->
### Debugging durable invariants (trust the artefact; sampled steps; silent no-op config paths)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b "Durable invariants and heuristics" section, carried forward from 016 and extended through this arc.
- **what:** The distilled heuristics: trust the rendered artefact, not the status (work items/commits can report success on no-op work); completed_at is the orchestration end, not the write instant; a config key read on a different path than it's set is a silent no-op, not an error; only save_page_sections writes page_components; 0 rows is not decisive; a negative inference from an artefact's shape needs the mechanism checked in ALL cases; reuse before rebuild; check the schema before SQL. Plus 016b v4 additions: two page-assembly paths with different chrome sources (stale `site_components` renders fossilise; only a full page-build rebuild re-renders templates; provenance greps + legacy-variable tell); the needle-gate template-surgery pattern (LIKE booleans + occurrence counts + backup + guarded idempotent UPDATE + RETURNING + rollback); `sites.status` vocabulary (draft/building/review/published/deployed/archived/error — 'active' is legacy; nothing filters on it; never scope blast-radius by it).
- **sources:** 016b_debugging_guide_7_3_(7).md#durable-invariants,#light-site-dark-chrome,#sql-pitfalls,#sites-status
- **relations:** the whole debugging category; 016 back-catalogue (other unit).
- **verify-later:** n/a (heuristics; primary copies of 016/016b covered by their own unit).

<!-- SOURCE: U09_adoption.md -->
### Convergence inertness: []map[string]interface{} vs []interface{} type-assertion bug
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "A clean re-adoption proved `reconcilePlanWithRealised` has never run for any site… `query_database` returns []map[string]interface{} — a type that does NOT satisfy that assertion in Go… Fix… accepts both… plus an 'existing pages loaded for convergence' count log so an empty set can never fail silently again" (2026-06-05, verified fixed same day).
- **what:** The whole doc-029 Phase-1 convergence was dead since deploy because `ValidateSitePlanAction` asserted `ev.([]interface{})` on QueryDatabaseAction output, which is `[]map[string]interface{}` — the assertion always failed, existingPages stayed empty, and reconcile early-returned silently. A canonical instance of the "silent empty input" failure class; the fix pairs the type switch with a count log so emptiness is observable. Also documented: QueryDatabaseAction stringifies jsonb columns (sections arrive as JSON strings needing json.Unmarshal).
- **sources:** FOCUS_adoption_faithfulness_via_locks(5).md#2026-06-05-correction, running_notes_14(25)#part-14l
- **relations:** convergence; union-clobber fix (downstream of it); debugging-guide no-op pattern entry
- **verify-later:** v3_site_actions.go type switch; 016 debugging guide v2_31+ entry

<!-- SOURCE: U09_adoption.md -->
### Defect-catalogue discipline (families by root cause; read-pin-confirm-fix)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "Enumerate every observed defect as a separate item before fixing, so distinct causes are not conflated… Causes marked 'tentative' have NOT been pinned" (CATALOGUE header); practiced across Families A–J with per-item verification states.
- **what:** Post-run methodology: walk the deployed site, catalogue defects grouped by root cause (not symptom) into families (A deployment gaps, B silent-fallback links, C list-component content, D section-data gaps, E content quality, F guides duplication, G design fidelity, H hygiene, I unknowns, J dispatch throughput), then work each as its own thread — read the responsible action, pin the cause, confirm against data, only then fix. Paired with reading-discipline rules: site_plan_pages is the authoritative plan output; confirm run completion before diagnosing; teardown by site_id never domain; matching updated_at is not authorship; a hardcoded site_id is stale after any teardown (resolve via domain subquery); an empty LEFT JOIN means wrong anchor, not missing link.
- **sources:** CATALOGUE_gamesdesign_post_sync_fix_defects(9).md, HANDOFF_2026-05-25#reading-discipline, running_notes_14(25)#principles
- **relations:** 016 debugging guide checklist items 20–22; council/fix-loop methodology
- **verify-later:** n/a (methodology); debugging-guide entries

<!-- SOURCE: U09_adoption.md -->
### Kafka consumer-group recovery (restart-to-rejoin, never replay)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "Resolution: chassis restart re-established group membership; a fresh trigger produced orchestration… Kafka recovery: restart-to-rejoin + park at latest + one fresh trigger, rather than replay-from-earliest (which would spawn stale adoptions)" (running_notes_14 Part 4).
- **what:** After a topic wipe destroyed `__consumer_offsets`/membership, the chassis logged clean consumer setup but wasn't joined; triggers produced messages nobody consumed (site row created by trigger path, no orchestration row). Diagnostic: `kafka-consumer-groups --describe` empty = not consuming regardless of producer health. Recovery doctrine: restart to rejoin, park at latest; a `--reset-offsets --to-earliest` replay was a mistake that risked duplicate stale adoptions. Principle: a trigger printing IDs proves production, not consumption.
- **sources:** running_notes_14(25)#part-4
- **relations:** orchestration creation writes a state row at creation (absence = never consumed); scheduler tick races during DB cleanup are noise
- **verify-later:** n/a (operational doctrine)

<!-- SOURCE: U09_adoption.md -->
### Migration/prompt-edit gotcha conventions (replace() anchors, funcMap templates, backup snapshots)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "Gotchas worth keeping (hard-won during the build)" (FOCUS_directory_builder); the guide-page_type migration applies them (anchor pre-checks counting 1, quote-free/newline-free replaces, ::jsonb cast validation, in-txn snapshots).
- **what:** Conventions for safely editing prompts/configs stored as JSONB text: Postgres replace() is literal-byte and silently no-ops on missed anchors while reporting UPDATE 1 (verify anchors with COUNT, keep them short and unique; re-entrant replaces append on every run); CASE doesn't short-circuit sub-SELECTs (use DO blocks + RAISE); NAMEDATALEN 63 truncates backup-table names; prompt templates pass through Go text/template so literal {{…}} needs funcMap helpers (placeholder/rangeStart/rangeEnd); prompt self-check and the Go validator are two halves that must change together; every migration carries a restorable in-txn snapshot with documented rollback.
- **sources:** FOCUS_directory_builder_and_list_components.md#gotchas, migration_adoption_add_guide_page_type.sql, running_notes_14(25)#snapshot-standard
- **relations:** thin-slice constitution (snapshot rule); component-creator Tier-D prompt work
- **verify-later:** n/a (conventions)

<!-- SOURCE: U09_adoption.md -->
### Manual work-item insertion as an operational rebuild lever
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "Operational fact confirmed: a manually-inserted needs_page / needs_content_page work item IS claimed by build-dispatch-loop (status triaged → claimed → complete), so single-page (re)builds can be hand-triggered" (HANDOFF_2026-06-06).
- **what:** Canonical hand-trigger shapes: re-render existing components → `needs_page` (spec {reason,page_name}); generate content → `needs_content_page` (spec {mode:'recreate',source:'adoption',page_name,page_type}); both handler_agent='page-build-handler', status='triaged', ON CONFLICT DO NOTHING, ids resolved via domain subquery. Verified end-to-end unstick pattern for skinner-box, guides-index and the homepage.
- **sources:** HANDOFF_2026-06-06#key-references, GUIDE_deploy_from_context_packs.md#C, RUNBOOK(2)#4
- **relations:** dispatch pipeline; positive-evidence verification ("complete" alone is not proof)
- **verify-later:** n/a (operational recipe)

<!-- SOURCE: U10_imagery.md -->
### Kafka per-spawn response-topic partition race
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** "did not reproduce on second run… monitor" (2026-05-11); HANDOFF 2026-07-12 "Kafka per-job response-topic partition race — transient; now surfaces as failed items (mark_item_failed fix) instead of silent successes."
- **what:** An adapter (git-adapter observed; kafka-go LeastBytes balancer) occasionally writes its response to partition 1 of a single-partition per-spawn topic, losing the reply — work succeeds but the orchestration times out/fails. Root cause suspected stale partition metadata for just-created topics. Never structurally fixed; consequence downgraded from silent-success to visible failed items by the mark_item_failed pattern. The same race killed a content-writer reply and produced the "no-op complete" anomaly.
- **sources:** ANALYSIS_phase_2f_two_defects.md#Defect-2, RUNNING_NOTES_imagery_best_in_class.md#Turn-16, HANDOFF_imagery_best_in_class.md#Open-threads
- **relations:** mark_item_failed error-honesty fix; consumer-group race (separate doc, chassis replicas=1).
- **verify-later:** `platform/kafka/producer.go` balancer; adapter logs for "topic partition not found".

<!-- SOURCE: U10_imagery.md -->
### Work-item re-drive and zombie-claim operational semantics
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** Standing lessons recorded Turns 31–32 (2026-07-12/13); "the zombie-claim dispatch stall was the single biggest time cost of the 2026-07-09/10 verification" (B9, still open).
- **what:** Hard-won dispatch mechanics: a claimed item stuck >~10 min blocks its ENTIRE site via find_dispatchable_site's NOT-EXISTS clause (standing unstick UPDATE; real fix = reaper cadence + per-item-type circuit breaker, TODO 6/10/11); re-driving an item requires resetting `attempt_count=0` and claim metadata, not just status (capped items are silently excluded — dispatch looks dead but is correctly idle); a just-finished orchestration's tail can re-stamp a freshly-reset item complete (state-machine race); manually-inserted items are NOT auto-triaged (insert as triaged); dedup is a partial unique (site_id, item_key) over non-terminal statuses whose exact semantics made resets awkward. Historical: dispatch once didn't claim triaged imagery items behind page work; fairness/observability gaps (outer ORDER BY, trigger not writing orchestration_states) remain listed.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-32, HANDOFF_imagery_best_in_class.md#Mechanisms, TODO_imagery_followups.md#6/#8/#9/#10, RUNBOOK_imagery_best_in_class.md#B9
- **relations:** mark_item_failed; state-machine corruption on failed items (claim metadata not cleared); scheduler-and-tasks.
- **verify-later:** find_dispatchable_site SQL; reaper cadence; idx_swi_dedup definition.

<!-- SOURCE: U10_imagery.md -->
### Pipeline field as soft routing label
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "check_unfulfilled_imagery_plan.go hardcodes Pipeline: 'build' — the 2026-05-17 fix is in the code" (verified 2026-07-08); Part B dispatcher-filter loosening scoped alongside.
- **what:** Discovery checks running under design-discovery-agent inherited pipeline='design', which build-dispatch-loop's item_pipeline filter silently excluded — needs_imagery items required manual UPDATEs to dispatch. Two-part fix: checks write Pipeline:"build" at source (pipeline is the destination handler's side, not the origin's), and the dispatcher's filter was removed so any future mismatched emission still dispatches. The field survives as a soft routing label for possible future multi-pipeline dispatchers.
- **sources:** TODO_imagery_followups.md#7, RUNNING_NOTES_imagery_best_in_class.md#Turn-2 (verification)
- **relations:** work-item dispatch semantics; design-discovery-agent context.
- **verify-later:** build-dispatch-loop load_items config; Pipeline literal in imagery checks.

<!-- SOURCE: U11_traffic_probe.md -->
### Traffic-probe field lessons absorbed into the debug guide (#24–#28)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-12 "Debug guide updated … 016_debugging_guide_v2_46" and 2026-06-13(g) "Debug guide v2_48".
- **what:** Five checklist entries earned in this project's field work, each rule + dated instance: #24 a config/workflow file is only authoritative at its runtime read-path (the stale agentchassis/.git/workflows/deploy-to-b2.yml nearly produced a never-firing Action); #25 prove the harness delivered the intended input before debugging the system (dash not expanding $'…' made the field literally "$value"); #26 shell variables never reach child processes without export, die with the session, and error-text-vs-source mismatch means a stale deployed artifact — read state back from the artifact, not `echo $KEY`; #27 never invent an interface — compiling standalone ≠ satisfying the real DiscoveryCheck signature; #28 agent_definitions is UNIQUE(type,version) with two similar category columns. Plus operator-handover lessons: explicit file manifests + a loud go vet/build check, flat-shipped workflows (delivery channel rejects dot-dirs), git branch -M main before first push.
- **sources:** traffic_probe_running_notes(28).md#2026-06-12 (debug guide v2_46, operator execution) + #2026-06-13-g (v2_48), traffic_probe_runbook(13).md#3.5-3.6 (traps in place)
- **relations:** debugging (016 guide family), per-domain notes convention
- **verify-later:** 016_debugging_guide latest version contains #24–#28

<!-- SOURCE: U12_docs024_archives.md -->
### Debugging playbook (early runbook)
- **category:** debugging
- **status-signal:** superseded
- **status-evidence:** Both archive files are early drafts of the operational runbook; the current authoritative version is `016_debugging_guide_v2_58_consolidated.md`.
- **what:** A ten-section operational runbook: pod health check, work-item status queries, scheduled-task flight-status, orchestration-state staleness, agent error log, handler-agent-definition existence checks, timeout ordering chain, a failed-item cleanup transaction, named failure patterns, and a single "quick health dashboard" query. The second draft adds a systematic dispatch-loop `input_mapping` path-mismatch diagnosis, missing-handler-agent detection, and a log-hunting technique.
- **sources:** old/older1/016_debugging_guide.md; old/older1/016_debugging_guide_v2_april26.md
- **relations:** timeout chain ordering; dispatch-loop input_mapping mismatch; wont_fix/needs_section_data patterns
- **verify-later:** whether the consolidated live debugging guide still carries these same queries/patterns.

<!-- SOURCE: U12_docs024_archives.md -->
### Timeout chain ordering contract
- **category:** debugging
- **status-signal:** unknown
- **status-evidence:** Stated as a hard ordering requirement in both drafts (claim_timeout > call_handler timeout > workflow timeout), with the call_handler timeout bumped from 900s to 1200s between drafts; not verified against the current consolidated guide.
- **what:** Three timeouts must nest correctly or two failure modes occur: reset-claim double-handling, or dispatch marking an item failed while the handler is still working with nothing listening for its response.
- **sources:** old/older1/016_debugging_guide.md#"7. Timeout Chain"; old/older1/016_debugging_guide_v2_april26.md#"7. Timeout Chain"
- **relations:** debugging playbook
- **verify-later:** current values of `claimed-item-timeout`, `build-dispatch-loop` call_handler timeout, per-handler workflow timeouts.

<!-- SOURCE: U12_docs024_archives.md -->
### Early pipeline-failure triage priorities dropped by root-cause diagnosis
- **category:** debugging
- **status-signal:** abandoned
- **status-evidence:** The 2026-04-14 report's P3 (vonc.com raw CSS), P4 (stale-item process gap), P5 (timeout tuning) don't appear in the 2026-04-15 v3 report's P1-P10 list at all.
- **what:** First-pass triage of 57 stuck work items framed three priorities at the symptom level. Within a day, deeper diagnosis replaced these with concretely-fixed root causes not originally identified: rate-limit errors misclassified as non-transient (1,869 occurrences), `load_page_record` lacking a `page_id` fallback, and later audit-finding routing/classification bugs.
- **sources:** old/older1/105_dispatch-pipeline-failures-report.md#"Priority Fixes"; old/older1/105_dispatch-pipeline-failures-report_v3.md#"Priority Fixes"
- **relations:** plan_sections pre-check evolution; three-way audit-finding classification
- **verify-later:** current state of vonc.com's about page (raw-CSS-serving bug).

<!-- SOURCE: U12_docs024_archives.md -->
### CrashLoop `exec: "./X"` image/binary-content mismatch diagnosis
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Live v2_58 header: "Recovered two §9 entries that were present in the earliest file v2_47(1) but had been dropped from the v2_48-onward branch."
- **what:** A three-command image-inspection technique (`docker run --entrypoint ls`, `docker inspect .Config.Entrypoint`, `.RepoDigests`) for diagnosing `CrashLoopBackOff` with `exec: "./X": no such file or directory` — proves the running image lacks the named binary (wrong build context / tag-sharing), not a config problem.
- **sources:** archive_april_26/016_debugging_guide_v2_47(1).md#"§9 CrashLoop exec"; docs024_key_docs_latest/016_debugging_guide_v2_58_consolidated.md
- **relations:** temporarily abandoned in the v2_48→v2_57 main branch (this fork diverged at v2_45), recovered wholesale into live v2_58
- **verify-later:** whether a CI guard ("fail build if binary absent") was ever implemented for thunder-adapter/analyser-adapter Dockerfiles.

<!-- SOURCE: U12_docs024_archives.md -->
### Hand-applied agent/launcher-def migrations are not commutative
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** The other v2_47(1)-only recovered entry; incident resolved by re-applying 110 then 111, RUNBOOK "2d state check" added as a live procedural safeguard.
- **what:** Re-applying migration 109 (per a runbook's "safe to re-run" claim) silently reverted later migrations 110/111 because 109 rebuilt DB-object nodes that 110 had replaced. A migration is idempotent only against its own prior application, never against later migrations touching the same path.
- **sources:** archive_april_26/016_debugging_guide_v2_47(1).md#"§9 Re-running an idempotent migration"; docs024_key_docs_latest/016_debugging_guide_v2_58_consolidated.md
- **relations:** NEW:migration-governance proposal (below)
- **verify-later:** confirm the `training-launcher` agent_definitions row currently reflects migrations 109-111 in correct order.

<!-- SOURCE: U12_docs024_archives.md -->
### gamesdesign `index` silent-staleness investigation — superseded hypothesis chain
- **category:** debugging
- **status-signal:** superseded
- **status-evidence:** Three successive, explicitly-labelled hypotheses in the same changelog: "silent-completion from a pod dying mid-flight" → "NOT a timeout/deploy issue at all... content-regression guard errors masked as success" → "SUPERSEDED-PENDING-CONFIRMATION" opening a metadata-path-mismatch thread.
- **what:** A multi-week live diagnosis of why gamesdesign.co.uk's `index` page stayed stale despite repeatedly "completing" rebuilds. Each hypothesis explicitly superseded the previous as new evidence arrived. Eventually-confirmed root cause is a more general mechanism — "Child workflow result silently replaced by a stub" (`output_field` vs `output_fields`), shipped 2026-06-18.
- **sources:** archive_april_26/016_debugging_guide_v2_49.md, v2_49(1).md, v2_49(2).md#"§9"; docs024_key_docs_latest/016_debugging_guide_v2_58_consolidated.md
- **relations:** own recursive application of "don't trust a complete status" heuristic
- **verify-later:** confirm `platform/orchestration/result_spec.go` (`resolveResultSpec` fix) is present in the current codebase.

<!-- SOURCE: U12_docs024_archives.md -->
### Pod label key is `agent-type` (hyphen) vs log field `agent_type` (underscore)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Stated as a settled operational rule with a named failure mode already observed.
- **what:** Kubernetes pod labels use `agent-type` while structured log JSON fields use `agent_type`; using the underscore form in a `kubectl logs -l` selector silently matches zero pods. Separately, a correct selector spans ALL live pods of that type, so a tail can mix in a previous run's failure dump.
- **sources:** archive_april_26/016b_debugging_guide_7(4).md#"Pod label key is agent-type..."
- **relations:** older trigger scripts (082/083c) still carry the underscore form; absent from canonical live 016b
- **verify-later:** grep trigger scripts 082/083c for the underscore `agent_type=` selector.

<!-- SOURCE: U12_docs024_archives.md -->
### Two failure envelopes — a COMPLETED parent orchestration does not mean the child succeeded
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Presented as a settled mechanism with two named, confirmed code paths (`sendWorkflowResponse` vs `notifyParentOfFailure`).
- **what:** A mid-run step failure is reported via `sendWorkflowResponse` with header `status:"complete"` but the real failure in the body, which the parent forwards and then itself shows COMPLETED with a non-empty `error` column; a START-time failure instead uses `notifyParentOfFailure` with `status:"error_unrecoverable"`. Consumers must check the body, never the header status alone.
- **sources:** archive_april_26/016b_debugging_guide_7(4).md#"Two failure envelopes"
- **relations:** same "trust the artefact, not the status" family as the guide's core silent-completion heuristics; absent from canonical live 016b
- **verify-later:** read the current `sendWorkflowResponse`/`notifyParentOfFailure` implementations to confirm the two-envelope shape.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Assumed-status-values trap (debugging lesson)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Formalized as Section 9 addendum + Section 0 checklist candidate in 016_debugging_guide_addenda.md
- **what:** General lesson: never assume status-column values from naming conventions — always run `SELECT DISTINCT status FROM <table>` first. `pages.status` uses `'active'` exclusively platform-wide; other plausible values simply don't exist.
- **sources:** js_snippets_news_gaswholesalers/016_debugging_guide_addenda.md#Assumed-status-values-trap, js_snippets_news_gaswholesalers/old/design_actions_status_filter_fix.md
- **relations:** CSS component-list fallback bug
- **verify-later:** grep/inspect `SELECT DISTINCT status FROM <table>`; `pages.status`; `'active'`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### "Renders empty" diagnostic method (data-binding, not template, diagnosis)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Formalized into 016_debugging_guide_addenda.md Section 9 headline entry + Section 0 checklist item #16
- **what:** A general, reusable 5-step diagnostic method for "a component renders its structural shell but no repeated content": (1) check page_components for orphaning; (2) confirm input_schema expectations; (3) check whether structured data exists anywhere; (4) count rendered shells; (5) compare actual sections against site_plan for duplicate/stale slots. Core lesson: empty shells mean the template ran — the bug is in data binding, never trigger a rebuild before completing this walk.
- **sources:** js_snippets_news_gaswholesalers/016_debugging_guide_addenda.md#empty-shells, js_snippets_news_gaswholesalers/old/empty_faqs.md
- **relations:** FAQ duplicate content-surface bug; rendered_html snapshot-not-view pattern; isolated build test methodology
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### rendered_html as snapshot-not-view (stale render after content_components migration)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Verified via a diagnostic query comparing template_has_script_src vs rendered_has_script_src on live gaswholesalers data
- **what:** A migration to `content_components.html_template` does not retroactively affect already-built pages, because `page-rerender` uses `page_components.rendered_html` — frozen output from the last writer run — and never re-pulls from the live template. General principle: `rendered_html` is a snapshot, not a live view; migrations touching `content_components` must also update affected pages' snapshots or trigger a rebuild.
- **sources:** js_snippets_news_gaswholesalers/016_debugging_guide_addenda.md#Migration-updates, js_snippets_news_gaswholesalers/old/findings_and_plan_news_visual.md
- **relations:** files_field deploy bug; "Renders empty" diagnostic method
- **verify-later:** grep/inspect `content_components.html_template`; `page-rerender`; `page_components.rendered_html`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Isolated build test methodology (throwaway test-page pattern for pipeline diagnosis)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** used successfully to prove the content writer was not the bug
- **what:** A reusable diagnostic technique: create a throwaway page (kept out of nav) with a deliberately minimal/isolated sections list, drive it through the full production build path, then read out `page_components` to conclusively attribute a bug to a specific pipeline layer. Used to prove the FAQ writer works correctly in isolation.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md#The-page-content-creation-flow, js_snippets_news_gaswholesalers/old/page_content_creation_flow.md
- **relations:** FAQ duplicate content-surface bug; "Renders empty" diagnostic method; page content-creation build pipeline trace
- **verify-later:** grep/inspect `page_components`

<!-- SOURCE: U14_docs019_runbooks.md -->
### Orchestrator COMPLETED while child FAILED (body.status check)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B5 "orchestrator can show COMPLETED while the child FAILED — header status complete, body.status failed; consumers of child results must check body.status (behaviour, recorded)".
- **what:** Recorded platform behaviour: a parent orchestration's header status can read complete while the child's embedded body carries failed; any consumer of child results must check body.status, not the header. Adopted cross-thread from the tools chat's notes.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B5 (useful from their notes)
- **relations:** oversize delivery (child fails at complete); stage-by-stage verification
- **verify-later:** response-building code paths; parent/child rows of a failed diagnose run

<!-- SOURCE: U14_docs019_runbooks.md -->
### error_step-inside-config gotcha and pod-reap evidence substitute
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** diagnosis_fix_loop(9) "New gotcha ADOPTED (their 001 §16 finding): error_step belongs INSIDE a step's config — step-LEVEL error_step is silently ignored (dormant bug instances exist in tool agents) … idle pods reap at ~3600s — the post-completion STATE DUMP (ProcessingHistory) is the accepted evidence substitute."
- **what:** Two operational facts: workflow error routing only works when `error_step` sits inside the step's `config` object (top-level placement is silently ignored — dormant instances exist and should be corrected when touching a workflow, as its own noted change); and spawned agent pods are reaped ~3600s after idle, so post-mortem evidence comes from the orchestration state's ProcessingHistory dump, not pod logs.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#collision-surface
- **relations:** stage-by-stage verification; standing evidence rules
- **verify-later:** error_step placement across tool agent workflows; agent-job-cleanup timing

<!-- SOURCE: U14_docs019_runbooks.md -->
### Stage-by-stage rebuild verification and the false-complete rule
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** gamesdesign_index_rebuild §5 stages A–E with per-stage SQL; "status='complete' is only meaningful together with Stage C showing changed components; complete + unchanged components = the old false-complete".
- **what:** The verification method for a page rebuild: A writer delivered a flat result to the parent (sections_metadata path check) → B save attempted/blocked loudly (agent_error_log) → C components actually changed (content_hash/updated_at fingerprint vs baseline) → D work item completed on a REAL save (complete only meaningful with changed components) → E deploy. Baseline-first (capture fingerprints before triggering), re-open the existing work item rather than fabricating one, and the triage table maps each stopping stage to its likely cause.
- **sources:** docs019/RUNBOOK_gamesdesign_index_rebuild.md#2; docs019/RUNBOOK_gamesdesign_index_rebuild.md#5; docs019/RUNBOOK_gamesdesign_index_rebuild.md#7
- **relations:** oversize delivery (fix #3); content-regression guard; standing evidence rules
- **verify-later:** page_components fingerprint queries; site_work_items re-open pattern

<!-- SOURCE: U15_docs019_running_notes.md -->
### Gamesdesign silent-no-op-rebuild bug (content-regression + status-rollup)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** v2(36) STATE DIGEST: "gamesdesign silent-no-op bug — RESOLVED (the real fix is now a fixture)."
- **what:** A real production bug used repeatedly as the diagnosis loop's eval fixture, diagnosed across several sessions with TWO wrong hypotheses along the way (per-section `max_tokens:2000` cap; then recreate-mode discriminator) before the real cause was found: a January chassis regression made `SagaCoordinator.extractWorkflowResult` honour only the PLURAL `output_fields` key, while `page-content-writer` declares the SINGULAR `output_field`, so the compiled page collapsed into an oversized state-dump skip path that reported "completed" while the live page never updated. Fix: `resolveResultSpec` (new, `result_spec.go`) treats singular as FLATTEN, honouring the long-ignored mapping key. The reversals in this diagnosis are the canonical worked example baked into the diagnosis loop's verdict prompt.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-14/17 gamesdesign entries; NOTES_running_synthesis_principles(59) 2026-06-13/14 diagnosis narrative.
- **relations:** SagaCoordinator output_field contract; diagnosis loop; B4a embedding-quality finding (this bug's real fix is the "ceiling" ground-truth task).

<!-- SOURCE: U15_docs019_running_notes.md -->
### Code-retrieval corpus staleness (§7 route)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** v4(39) headers 2026-07-02: "corpus check result: the index is the blocker... the index is of a YEAR-OLD tree."
- **what:** After the diagnosis loop was measured to gain nothing from code retrieval (see B4a finding), a follow-up investigation (§7 route) found the underlying `code_symbols` index itself was built from a year-old stale checkout of the default branch (main stale since 2025-07-14) — a corpus problem, not a retrieval-quality problem — leading to a reindexing effort, ref-pinning strategy, and ultimately the decision to migrate the code-indexer's analysis step onto the already-proven `analyse_repo_local` path.
- **sources:** NOTES_running_synthesis_v4(39).md headers 2026-07-02/03; DECISIONS section.
- **relations:** B4a embedding-quality finding; code-context retrieval infrastructure.
- **verify-later:** Current freshness of the deployed `code_symbols` index; whether the analyse-step migration was applied.

<!-- SOURCE: U16_docs019_design_plans.md -->
### Spawn-consumed columns lesson (seeds copy image columns from a live donor)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NNN_fix_researcher_spawn_columns documents the incident and fix; HANDOFF_builder_thread carries it as a standing guard ("seeds must copy image columns from a live donor (the amended seed does)").
- **what:** getAgentDefinition SELECTs image_repository/image_tag/command/resources/health_config/env_vars/idle_timeout_seconds and gates on is_active=true; a seed populating only default_config leaves command NULL → the image's default entrypoint boots the GENERIC chassis service, which never reads the injected AGENT_TYPE env, so the dispatcher's call goes unheard and the item stays claimed. Fix and rule: copy the spawn-consumed infrastructure columns from a proven donor (deliberately NOT capabilities/topics/default_config). Related: image_tag DEFAULT 'latest' pointed at an ancient build; the makefile now pins IMAGE_TAG. Sibling gotchas carried with it: pod label key is agent-type (hyphen); check body.status not just the header; error_step belongs INSIDE step config (step-level silently ignored); idle pods reap ~3600s with ProcessingHistory dumps as post-reap evidence.
- **sources:** NNN_fix_researcher_spawn_columns.sql; HANDOFF_builder_thread.md#2,#5; HANDOFF_fixloop_thread(8).md#3
- **relations:** workflow-in-default_config lesson; index-orchestrator spawn wrapper
- **verify-later:** guidelines 001 New Agent checklist line (flagged residual)

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Debugging guide & assumption-checklist methodology
- **category:** debugging
- **status-signal:** superseded
- **status-evidence:** 016 v2_44 §0 "Most defects in recent sessions … came from acting on unverified assumptions"; archive copy, live successor in docs024
- **what:** The canonical symptom→cause→fix guide, fronted by a 23-item assumption checklist. Covers pod health, work-item/orchestration/scheduled-task/error-log queries, timeout chain, and ~50 specific failure patterns.
- **sources:** WM/016_debugging_guide_v2_44.md#0, WM/016_debugging_guide_v2_44.md#9, WM/016_debugging_guide_v2_44.md#7
- **relations:** superseded by docs024 live 016; architectural tensions; agent = row in agent_definitions
- **verify-later:** orchestration_states.error_preview; agent_error_log; llm_call_log

<!-- SOURCE: U18_sql_for_agents.md -->
### rag_index chunkContent OOM saga (bypass → reenable → rebypass → fix)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Four-migration arc 135/139/140/141 (2026-07-09/10) with root cause CONFIRMED in 140: "chunkContent() never terminated on content longer than chunk_size... ~2Gi of duplicate chunks in seconds. Both chassis OOMKills were this loop."
- **what:** A model incident record: tool creation hung/OOMed at index_plan. First hypothesis (no embedding deadline) produced 135's bypass + a hygiene deadline (139); reoccurrence disproved it; the real bug was a non-terminating chunk loop (start = end - overlap re-entering forever), fixed in Go with regression tests, then re-enabled by 141. Durable practices demonstrated: reversible SQL bypasses that keep truth in Postgres (write_plan) while sacrificing only derived indexing; explicit preconditions in re-enable migrations; superseding one's own root-cause statements on record.
- **sources:** 135_bypass_index_plan_until_embed_timeout.sql; 139_reenable_index_plan.sql; 140_rebypass_index_plan_chunk_loop.sql; 141_reenable_index_plan_after_chunk_fix.sql
- **relations:** rag knowledge base; travelling docs pipeline notes; 016b debugging lessons
- **verify-later:** rag_actions_chunk_test.go presence; deployed image ≥ fix commit

<!-- SOURCE: U18_sql_for_agents.md -->
### error_step-inside-config routing rule
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 128 "Durable rules this incident banked (016b): error_step lives INSIDE step.Config and must name an EXISTING step; derive convergence targets from the step's own next_step, never guess"; effect verified live 2026-07-10; 131/132 retro-move ten inert step-level error_steps into config.
- **what:** Chassis workflow convention discovered through failures: the coordinator reads step.Config["error_step"] only — step-LEVEL error_step keys are silently ignored; a routing target that names a non-existent step fails the whole workflow. Correct-while-touching policy migrates old inert keys whenever a workflow is edited.
- **sources:** 128_fix_load_runtime_error_step_target.sql; 127_diagnose_load_runtime_error_step.sql; 131_tool_generator_plan_writing.sql; 132_fix_agents_note_writing.sql
- **relations:** 016b debugging heuristics; template field-path rule (134)
- **verify-later:** coordinator error-routing code

<!-- SOURCE: U18_sql_for_agents.md -->
### Prompt-template field-path rule (text vs json output shapes)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 134 (2026-07-09) "THE RULE (proven by this run, not assumed)": text-format steps pass the bare string to downstream templates ({{.generated_html}}, not .result); json-format steps pass a map (use `| toJSON`); action-config field paths are a DIFFERENT resolver and keep .result.
- **what:** A durable rendering contract distinguishing three resolvers: Go template rendering of LLM text results (bare string), of JSON results (map, dump with toJSON rather than guessing keys), and action-config field paths (keep .result suffix). Applied as one blocker fix plus three pre-emptive corrections of the same bug class.
- **sources:** 134_fix_prompt_template_field_paths.sql
- **relations:** call metadata/response convention; error containment via config.error_step (docs steps can never fail tool creation, 131)
- **verify-later:** ExtractActionInputs / template renderer code

<!-- SOURCE: U19_sql_tables_components.md -->
### orchestration_state_audit investigation trigger
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Create-trigger + analysis queries (time_since_prev via LAG, pg_backend_pid, application_name) and explicit "Remove trigger when done investigating" teardown.
- **what:** A temporary, attachable audit table + AFTER UPDATE trigger capturing every version/status/current_step transition on orchestration_states — used to diagnose state races and stuck orchestrations, then removed. Distinct from permanent logs; also cleaned up by database-cleanup (keeps last 100k rows).
- **sources:** docs/agent_docs/sql_for_tables/010_orchestration_state_audit.sql; docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#database-cleanup
- **relations:** debugging guide; database cleanup.
- **verify-later:** whether trigger currently attached.

<!-- SOURCE: U19_sql_tables_components.md -->
### agent_error_log persistent error record
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "This replaces digging through kubectl logs to find error details"; captured from routeToErrorStep and notifyParentOfFailure; referenced later as the sink for Tier-D validator rejections.
- **what:** Queryable record of every agent error: what failed (site/domain/work_item), where (orchestration, agent_type/id, pod, step, action), the error (message, error_code, severity), a JSONB context snapshot, and resolution tracking (resolved/resolved_by). Indexed for dashboard recency, per-site, unresolved, and per-agent-type frequency views.
- **sources:** docs/agent_docs/sql_for_tables/022_agent_error_log.sql; docs/agent_docs/sql_for_tables/005_content_components.sql#migration-042
- **relations:** database cleanup retention; fix loops consuming structured errors.
- **verify-later:** writers in chassis error paths.

<!-- SOURCE: U19_sql_tables_components.md -->
### http_request_log outbound HTTP observability
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Migration "Follows the same pattern as llm_call_log"; stats view with calls_last_5min for rate-limit monitoring; cleanup 90d success / 180d errors.
- **what:** Centralised log of every outbound HTTP call from Go actions: caller identity (agent/step/orchestration/action_name), method/url/domain/path, response status/bytes/latency/success, metadata JSONB. Purposes: operational visibility and per-domain rate-limit tracking (e.g. Companies House).
- **sources:** docs/agent_docs/sql_for_tables/026_http_request_log.sql
- **relations:** llm_call_log (pattern sibling); companies-house rate limiting.
- **verify-later:** HTTP client wrapper writing rows.

<!-- SOURCE: U19_sql_tables_components.md -->
### Claimed-item timeout with evidence-based auto-completion
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** v1 then v2 ("SUPERSEDES... Apply THIS one") migrations with two confirmed production false positives dated 2026-05-12 and 2026-06-04 (gamesdesign homepage auto-completed with ZERO page_components — root cause of the missing root index.html).
- **what:** The stuck-claim recovery task distinguishes "work actually finished but the response was lost" from "handler died": items claimed >15 min are auto-completed only on artifact-specific evidence — needs_content_page requires page_components rows for that page updated after the claim (ground truth, not the untrustworthy build_status='deployed' flag), page_rerender requires page.deployed_at after claim, needs_design keeps a caveated site-level check; needs_rerender is deliberately excluded (site-level, retry is cheap). Everything else resets at >40 min with attempt accounting and fail-on-exhaustion.
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql#migration_claimed_item_timeout_evidence_v2; docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#claimed-item-timeout
- **relations:** work queue lifecycle; build_status CHECK (flag trust); UpdatePageStatusAction 0-component guard ("Option B").
- **verify-later:** live pre_query text of claimed-item-timeout; debugging guide section 9.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Message-flow logging / observability plan
- **category:** debugging
- **status-signal:** aspirational
- **status-evidence:** README.002 Week-2 objective: "MessageFlowLogger… Track every message through the system with database persistence"; docs002/0100 problem statement repeats the desire ("closely log and track the creation of agents, the messages…").
- **what:** Persist every send/receive event, agent creation, and topic routing decision to the DB for replay/debugging. Only zap logging plus orchestration_states processing_history is evidenced; a dedicated message-flow store never appears.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md; docs002_hitl_parallel/README.0100.state_of_play_for_creating_website.md
- **relations:** debugging category (docs 016 successors); processed_messages table (exists — see reset runbook).
- **verify-later:** processed_messages table purpose; any message audit table.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Orchestration environment reset runbook
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** The identical script repeated across ≥5 docs: scale agent-chassis to 0, TRUNCATE processed_messages/orchestration_states/pending_requests, delete spawned jobs, delete all job.* topics, delete bootstrap topics, reset all consumer-group offsets to earliest, scale back up.
- **what:** The standard clean-slate procedure for the early platform's test cycles — also documents the persistence surface of the era: processed_messages (dedupe), orchestration_states, pending_requests tables; job.* + system.agent.* topics; spawned-by=orchestrator job labels.
- **sources:** docs001_flow_general/README.095d.mycurrentinputmessagebeforechanging.md; docs001_flow_general/README.096d.robotics_startmessage.md; docs004_website_capture_project/initial_messages/initial_messages.txt
- **relations:** debugging (docs 016 successors); stateless-agents concept (what gets truncated).
- **verify-later:** pending_requests/processed_messages tables still present?

<!-- SOURCE: U20_legacy_docs_a.md -->
### Early message-routing failure modes (case-study catalogue)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Each bug has a trace + fix doc: nested-vs-flat input_data mismatch (flow2), verbose child responses breaking aggregation (flow2), silent root completion (flow2/3), duplicate second response to own topic = "poisoned pill" crash-loop (flow5), responses_topic dropped in header parsing → __initial_responses_topic__ empty (4.2), missing in_response_to_request_id (081.b), fire-and-forget spawn ignoring init responses (flow12).
- **what:** The canon of failure modes that shaped the architecture: every major convention (data normalisation, reply-to storage, perspective transformation, single completion path, await semantics) exists as the fix to one of these traced production bugs. Valuable as diagnostic priors for any council debugging agent.
- **sources:** docs001_flow_general/README.011.flow2.md; docs001_flow_general/README.016.flow5.md; docs001_flow_general/README.4.2.lifespanofresponsestopic.md; docs001_flow_general/README.023.flow12.await_response.md; docs001_flow_general/README.012.flow3.md
- **relations:** all system-architecture concepts above; debugging heuristics (docs 016b successor).
- **verify-later:** none — historical lessons.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Parent-timeout vs child-HITL race
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** docs014/001 log trace: "pageflow-builder times out (5 min)... content-reviewer: Cleaned up expired awaited requests count=1. The fix is to increase the parent's timeout."
- **what:** A failure class where a parent's call_agent timeout fires before the child's HITL request can be answered; the parent retries with null body and the child's awaited request is cleaned up as expired, losing the pause. Fix: parent timeouts must exceed child HITL timeout windows.
- **sources:** docs014_research_agent/001_human_in_the_loop_response_flow.md#Why-There-Were-No-Awaited-Requests
- **relations:** stale orchestration sweeper; HITL protocol; timeout heuristics in debugging docs.
- **verify-later:** current call_agent timeout_seconds vs HITL timeout defaults.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Orchestration debug log taxonomy
- **category:** debugging
- **status-signal:** superseded
- **status-evidence:** docs006/010 raw notes listing grep targets ("DEBUGaa: What have I done with CollectedData", "The Golden Search: grep -B 5 -A 30 generate_html") plus a real database lock incident (idle-in-transaction blocking INSERT INTO sites).
- **what:** The early debugging playbook: canonical log messages for action execution flow, LLM calls, data extraction and CollectedData tracking, with kubectl grep recipes; plus pg_stat_activity lock triage and pg_terminate_backend for idle-in-transaction blockers. Ancestor of the formal debugging guides.
- **sources:** docs006_workflow_builder/010_debugging.md
- **relations:** debugging category docs 016/016b; data-path problem.
- **verify-later:** whether DEBUGaa markers remain in code.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Legacy un-extracted Mode-B shells (js-not-extracted class)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** Root-caused 2026-06-29 ("stored through a path that did NOT run separateInlineJS — most likely they predate its addition"); cosmetic-script extraction for provocation-card/lobby-grid still on the backlog 2026-07-09.
- **what:** provocation-card, lobby-grid (and brief-explanation) were stored via a pre-separateInlineJS path: raw inline script still in html_template, empty js_content, empty schema, `<no value>` placeholders — so `/tools/assets/{fn}.js` was never produced and their built-in interactivity never deployed. provocation-card's stored script was additionally truncated at generation (no `</script>`), which once shipped and swallowed the page footer. One creation-era bug with several surface symptoms (`js-not-extracted`, `mode-b-template`, section drops). Fix direction: regenerate through the current store path.
- **sources:** docs/RUNBOOK_phase2_provocation_js(29).md#extraction-bug-findings; docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~19:30 + #2026-07-02-~19:35; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-07-gate-passed (side-evidence)
- **relations:** Mode A/Mode B taxonomy; store-path validation hardening; separateInlineJS
- **verify-later:** content_components js_content/html_template for provocation-card 6163ff14 and lobby-grid 9304f14d (still raw inline?)

<!-- SOURCE: U23_docs_root_vonc.md -->
### Mode A / Mode B broken-template taxonomy + repair/regeneration routing
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_vonc_session "Structural findings to carry forward"; code delivered 2026-06-22/23 (checkBrokenTemplateSlots, repair_template_slots); gauntlet-interface Mode-A repaired, archetype-result-card Mode-B regenerated to q100.
- **what:** Two distinct broken-template failure modes in the component library. Mode A: `<no value>FIELD</no>` — a render output stored as source with field names surviving as fallback text; repairable by string substitution (`repair_template_slots`). Mode B: bare `<no value>` — template rendered against an empty context and the cleaned output stored back; field names irretrievably lost; requires `needs_component_regeneration` → component-creator. `repair_template_slots` detects Mode B (no `</no>` tags) and returns needs_regeneration instead of attempting repair; `checkBrokenTemplateSlots` discovery check surfaces both.
- **sources:** docs/RUNBOOK_vonc_session(1).md#structural-findings; docs/RUNNING_NOTES_vonc(36).md#two-broken-template-failure-modes; docs/RUNBOOK_vonc_migrations(14).md#step-1
- **relations:** legacy un-extracted shells; store-path validation (rejects `<no value>` at the gate); component regeneration in place
- **verify-later:** check_component_standards.go; fix_component_template_action.go repairNoValueSlots

<!-- SOURCE: U23_docs_root_vonc.md -->
### Trust-the-artifact debugging doctrine (silent-success family + verification discipline)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b "Durable invariants" section; proven repeatedly (ten complete-with-nothing items; "after ten silent no-ops... NOT trusting 'complete' without artifacts").
- **what:** The unit's core debugging doctrine: a `complete` work item or green commit proves nothing — verify by artifact (DB row, curl, browser); completed_at is orchestration END, not the write instant (trace child orchestrations); a config key read on a different path than it is set is a silent no-op (compare producer output to consumer read by exact path); 0 rows is not decisive until the query is cleared (wrong column/id/schema/window); a negative inference from an artifact's shape needs the mechanism checked in all cases (the separateInlineJS attribute-skip example); pod logs are ephemeral across rollouts (grep zap by message + JSON field, never 'field=value'; agent_error_log outlives pods); copy full UUIDs, never hand-type; ±6-byte js_len paste drift is cosmetic — bundle and browser are ground truth; dated backup tables per change (never reuse an IF-NOT-EXISTS backup name); only save_page_sections writes page_components.
- **sources:** docs/016b_debugging_guide_merged(3).md#durable-invariants; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§8; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-03-section-drop-closed
- **relations:** complete_error family; zap/pod-log entry; SQL surgery pattern
- **verify-later:** n/a (doctrine); stage 2 can test individual heuristics against code

<!-- SOURCE: U23_docs_root_vonc.md -->
### system-stats key-contract mismatch (content_data ↔ template key sets)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016b Part 5 "TRIAGED, 2026-06-24"; remedy un-applied at that entry ("full content rebuild... then re-check"); the component itself later regenerated to q100 in the vonc arc.
- **what:** A populated-but-blank section is a content↔template KEY-CONTRACT problem, not a generation failure: system-stats' stored content_data keys (eyebrow/heading/stat_1_number...) shared ZERO keys with its template placeholders (eyebrow_label/section_headline/stat1_value...) after component-creator rewrote the component mid-flight, so every placeholder rendered empty and the (correct) visible-content filter dropped the band fleet-wide (usage_count 22). Durable heuristic: diff the two key sets directly; and a component schema change should trigger dependent rebuilds.
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 5) + #wrong-turns (#4)
- **relations:** shared library field guard (the same incident class the guard now blocks); visible-content filter
- **verify-later:** whether schema-change→dependent-rebuild triggering exists (markPagesPendingRebuild covers regen; mid-build rewrites?)

<!-- SOURCE: U23_docs_root_vonc.md -->
### SQL template-surgery pattern (needle-gate discipline)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b merged entry: "Template-surgery pattern that held up" with the 2026-07-02 false-alarm refinement; practised across the marker/ghost-row/truncation fixes.
- **what:** Safe in-DB template edits: (1) needle-gate read — every needle as a LIKE boolean PLUS occurrence counts so partial coverage is visible BEFORE mutating (counts must be counted from the dump, not recalled); (2) shell backup of the full column; (3) guarded idempotent UPDATE (exact-string nested replace or anchored regexp_replace with backreference, plus NOT LIKE pre-state guard); (4) RETURNING boolean checks; (5) rollback file. Postgres pitfalls: regex quantifier bounds cap at 255; substring-with-parens returns the capture group; gradient-embedded hexes escape naive background regexes; needles containing literal % can't be LIKE-gated (use position()). Anchor REPLACEs on the opening tag (see marker lesson); dump→edit-offline→full-text UPDATE for multi-line blocks.
- **sources:** docs/016b_debugging_guide_merged(3).md#sql-verification-pitfalls; docs/fix_archive_template_display(1).sql (header); docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§8
- **relations:** marker anchoring; hidden-vs-author-CSS fix; sanctioned edit paths (this is the fallback)
- **verify-later:** n/a (practice)

<!-- SOURCE: U23_docs_root_vonc.md -->
### Marker/attribute REPLACE anchoring lesson (fix_marker_selector)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Bug introduced twice (provocation-card, lobby-grid), fixed via fix_marker_selector.sql with RETURNING checks (still_broken=f ×4), corrected HTML redeployed 2026-07-04; guide entry added.
- **what:** Adding an attribute by replacing the bare string `data-component="X"` also hits the section's own inline `querySelector('[data-component="X"]')`, producing a malformed two-attribute selector → SyntaxError → the cosmetic IIFE dies (loaders unaffected). Rule: anchor marker REPLACEs on the OPENING TAG (the copy followed by more attributes), revert only the in-selector copy (the one followed by `]`); better still, emit markers at generation.
- **sources:** docs/fix_marker_selector.sql (header); docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-marker-replace-broke; docs/016b_debugging_guide_merged(3).md#data-runtime-fill-marker-anchoring
- **relations:** generation-time guards (the prevention); SQL surgery pattern
- **verify-later:** n/a (lesson; instance fixed)

<!-- SOURCE: U23_docs_root_vonc.md -->
### `hidden` attribute vs author CSS (clone-template ghost rows)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Ghost-row fix verified end-to-end 2026-07-08 (rendered_len 7455→7671; live grep 2); prevention added to guide + component-creator requirement.
- **what:** The `hidden` attribute maps to UA-stylesheet `display:none`, which loses to ANY author `display` rule on the same element — so a hidden clone-template item inside a `display:grid` item class renders as a ghost row. Fix: a more specific author rule `[data-…-template] { display:none; }` in template AND instance (the REPLACE correctly fired twice — base selector + its mobile media-query copy). Prevention: component-creator must emit the hiding rule alongside `hidden` for clone templates.
- **sources:** docs/fix_archive_template_display(1).sql (header); docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-08; docs/016b_debugging_guide_merged(3).md#hidden-attribute-loses
- **relations:** generation-time guards; clone-template list pattern
- **verify-later:** component-creator prompt includes the hiding-rule requirement?

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Consolidated debugging guide (016)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Live 016_debugging_guide_v2_58_consolidated.md has 53 §9 failure subsections; archived v2_48 has a strict subset (comm diff: nothing in v2_48 absent from v2_58); top-level structure identical across v2_34/v2_48/v2_58.
- **what:** The canonical operational runbook: assumption checklist, pod health, work-item status, orchestration states, timeout chain, failed-item cleanup, and a large accreting "Specific Failure Patterns" catalogue plus the `detected→triaged→claimed` work-item state machine. Archived v2_34–v2_48 are monotonically-growing earlier snapshots; v2_58 supersedes them with added patterns (Thunder GPU provisioning, presign O(n²) bloat, tool/game pages never deploying, adoption slug-mangling by WriteSitePlanAction). No concepts were dropped between archived and live.
- **sources:** 016_debugging_guide_v2_48(1).md#9-specific-failure-patterns, #work-item-lifecycle
- **relations:** replacement = live 016_debugging_guide_v2_58_consolidated.md; snapshot-shadowing defect; tool widget clobber
- **verify-later:** live 016_debugging_guide_v2_58 §9 vs archived deltas

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Snapshot-shadowing defect (version+1000 snapshots outrank active rows)
- **category:** debugging
- **status-signal:** superseded
- **status-evidence:** ANALYSIS_phase_2f_two_defects (2026-05-11): `snapshot_agent()` inserts snapshots as `version+1000, is_snapshot=true`; a loader `ORDER BY version DESC LIMIT 1` without `is_snapshot=false` reads version 1001 over active version 1, shipping pre-migration workflow despite correct DB state; PLAN_imagery_loop_closure 2F "loader-snapshot defect" patched `processor.go::loadAgentDefinition` and `spawn_actions.go::getAgentDefinition` with `is_active=true AND (is_snapshot IS NULL OR is_snapshot=false)`.
- **what:** The model-swap/rollback snapshot mechanism creates rows that sort ahead of the active definition in version-descending queries, so any naive "most recent" agent-definition loader silently reads stale config. Structural, latent since launch; surfaced when Phase 2F first depended on a value that differed between active and snapshot rows. Fixed by adding the snapshot filter to loaders.
- **sources:** imagery/old/ANALYSIS_phase_2f_two_defects(1).md#defect-1; imagery/old/PLAN_imagery_loop_closure(9).md#2f
- **relations:** replacement fix in processor.go/spawn_actions.go loaders; 021_model_swap_and_rollback.sql snapshot_agent()
- **verify-later:** grep "FROM agent_definitions" *.go for is_snapshot filter; 021_model_swap_and_rollback.sql

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Adapter per-spawn-topic partition defect (kafka LeastBytes balancer)
- **category:** debugging
- **status-signal:** unknown
- **status-evidence:** ANALYSIS_phase_2f defect 2 (2026-05-11): git adapter wrote a success response to partition 1 of a single-partition per-spawn topic (`job.…responses-0` only), Kafka rejected "topic partition not found", response lost; did NOT reproduce on the 13:27 rerun after pod restart — "may be transient kafka-go metadata caching".
- **what:** kafka-go `LeastBytes` balancer may pick an out-of-range partition on a freshly-created per-spawn topic before partition metadata refreshes, dropping an adapter's success response while the underlying git commit succeeded (orchestration reports failure on succeeded work). Suspected to affect any adapter writing to per-spawn topics (webscrape, image-generator). Parked for monitoring, not fixed.
- **sources:** imagery/old/ANALYSIS_phase_2f_two_defects(1).md#defect-2
- **relations:** ANALYSIS_chassis_response_consumer_group_race.md (sibling parked defect); platform/kafka/producer.go
- **verify-later:** platform/kafka/producer.go Balancer; topic_manager per-spawn partition count

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Secret hygiene: image-provider API key rotation
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** HANDOFF_robot_hands_rebuild "SECURITY (highest): scrub + rotate STABILITY_API_KEY and BANANA_API_KEY (plaintext in logs; Banana on paid tier)"; TODO_imagery_followups(15) "SECURITY — STILL OPEN, STILL HIGHEST PRIORITY (do not let slide)."
- **what:** Image-generation provider API keys (Stability, Banana — paid tier) were being logged in plaintext; the standing highest-priority remediation is to scrub logs and rotate both keys. Repeatedly carried forward across imagery sessions without closure.
- **sources:** imagery/old/HANDOFF_robot_hands_rebuild(2).md#carried-forward; imagery/old/TODO_imagery_followups(15).md#security
- **relations:** image generation pipeline; adapter deployment; storage secrets
- **verify-later:** adapter logging of STABILITY_API_KEY/BANANA_API_KEY; secret rotation in personae-default-secrets

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Adapter-vs-chassis deployment drift
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** loop_closure(9) known issue "Adapter deployment vs chassis deployment (2026-05-14)": image-generator adapter (`dynamic_adapter.go`) is a separate K8s resource from the chassis (`generate_image_actions.go`); after a chassis rebuild+rollout the adapter may run an older binary — the 2H `generateImage: stability request` log line wasn't found on adapter pods.
- **what:** The image-generator adapter and the chassis are deployed as distinct Kubernetes resources, so chassis rebuilds don't refresh the adapter binary — leaving action-layer changes (e.g. 2H per-kind cfg_scale/negative_prompt) potentially inactive at the adapter. Recommendation: document which deployment carries which binary and add the adapter to the rebuild/rollout sequence.
- **sources:** imagery/old/PLAN_imagery_loop_closure(9).md#known-issues; imagery/old/HANDOFF_robot_hands_rebuild(2).md#carried-forward
- **relations:** image request shape (2H); Stability timeout 30→120s side-fix; multi-cluster dispatch
- **verify-later:** image-generator-adapter deployment vs chassis image tag; rollout sequence in Makefile

## Additional carried operational deltas (not standalone concepts)
From imagery/old TODO/STATUS + loop_closure "Known issues": `llm_call_log.agent_type` populated empty (params.AgentType not threaded to LogLLMCall, also noted in doc 009); dispatch loop not claiming `triaged` image items behind page work; FAILED orchestrations accumulating in `orchestration_states` with no cleanup; variant chain missing `site_id` so variant heroes skip `imagery_direction`; legacy `image_prompts` age-out reframed to operational deregistration rather than a `check_legacy_image_prompts_aspect`.

## Note on many-images/per-component/product-imagery/audit-loop/adoption-image-mirror/vision-auditor/provider-router concepts
This sub-agent also surfaced several imagery-domain concepts (imagery generation pipeline, audit-and-fix loop, structured site_plan_imagery Phase 2G, asset locking 2A-2E, image-build-handler storage architecture, image request shape 2H, adoption image mirror, many-images-per-page, icon rendering via Lucide, product imagery via affiliate_products, vision-capable LLM auditor, image provider router, news feed pipeline + enrichment + price-aware filtering, rebuild-vs-rerender, files_field deploy dependency, tool generation pipeline, component schema-contract drift, cross-cluster Postgres/Kafka topology, multi-cluster dispatch MVP gaps) that substantially overlap U10 (imagery) and U09 (adoption) which were separately extracted with fuller code-scope access. Consolidation should de-duplicate against U09/U10 rather than re-litigate; where this unit's evidence adds a NEW dated fact (e.g. the snapshot-shadowing defect, the kafka partition defect, the RAG/GPU-infra superseded lineage), it is retained above as its own entry.

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Launcher reply-topic own-vs-parent derivation (Decision D4)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(39) "Update — 2026-06-02 16:12: D4 CONFIRMED live … the adapter's reply went to system.agent.generic.responses — the agent's own ExecutionContext.ResponsesTopic"
- **what:** An intermediate adapter reply must be routed to the agent's own `ExecutionContext.ResponsesTopic` (seeded from `__my_responses_topic__`), NOT `__parent_responses_topic__` (which is only for the child→parent final notification). The inherited handoff had this backwards; provision/decommission always used own-topic and worked. The same class of bug bit `dispatch_thunder_ssh_get_status` (cloned from ssh_exec) and was fixed to prefer `execCtx.ResponsesTopic`.
- **sources:** phase5/NOTES_phase5_training_launcher_running(39).md#3 (D4), #6, #10; docubundle/.../STATUS_thunder_adapter_2026-06_04.md#update-2026-06-04
- **relations:** corrects the superseded 2026-05-24 handoff claims; a latent same bug remains in ssh_exec dispatch if fired top-level
- **verify-later:** thunder_prepare_object_url_dispatch.go, thunder_ssh_exec_dispatch.go, thunder_ssh_get_status_dispatch.go; coordinator determineResponsesTopic

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### gpu-provisioner output shape flattening (output_fields plural)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(39) "Update — 2026-06-03 ~17:5x: 104 written" — "extractWorkflowResult … reads output_fields — PLURAL only. The gpu-provisioner complete uses output_field (SINGULAR) … falls to the fallback branch"
- **what:** `call_launcher` failed on `provisioning_result.provisioning_id not found` because gpu-provisioner's `complete` step used singular `output_field` (which `extractWorkflowResult` never reads), so its result came out step-name-keyed as `{dispatch_provision, input_data}`. Migration 104 fixed the provisioner's `complete` to plural `output_fields:["dispatch_provision"]` and re-pointed the launcher mapping to `provisioning_result.dispatch_provision.provisioning_id`; a proper chassis fix (honour singular output_field) was vetoed in favour of making the non-compliant agent conform.
- **sources:** phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-03-15:47, #17:4x, #17:5x
- **relations:** launcher input-mapping contract; same singular bug latent in thunder-reaper
- **verify-later:** extractWorkflowResult; agent_definitions gpu-provisioner (0bf9fa8a); migration 104

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Loop-await send-before-register race + preRegisterAwaitedRequest fix
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(39) "Update — 2026-06-09: pre-register fix CONFIRMED in prod … Every presign_checkpoints_iter_N_presign_one logged ClaimAwaitedRequest: status_before=waiting … claimed:true … The send-before-register race is closed"
- **what:** The central thunder-checkpoint-race bug: the local dispatch `dispatch_thunder_prepare_object_url` produced the adapter request and returned `await_response:true` BEFORE the coordinator inserted the `awaited_requests` row, so a fast (~1s) reply beat the insert, `ClaimAwaitedRequest` (WHERE status='waiting') found nothing, the reply was dropped, and the timeout re-dispatched every ~3 min forever. `spawn_agent`/`call_agent` avoid it via `preRegisterAwaitedRequest` (register-before-send, `ON CONFLICT (request_id) DO NOTHING`). Fix: call the same helper in the dispatch before `ProduceWithValidation` (guarded `if params.DB != nil`); note the helper's hardcoded 120s timeout_at then pins every presign await.
- **sources:** docubundle/.../HANDOFF_2026-06-06_checkpoint_upload_loop_await_race.md; docubundle/.../CONTEXT_PACK_thunder_checkpoint_race.md; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-06-3, #2026-06-08, #2026-06-08-2
- **relations:** fourth cause of the `awaited_requests`-stuck-waiting symptom (016 §9); the batch presign superseded the loop that exposed it
- **verify-later:** thunder_prepare_object_url_dispatch.go; spawn_actions.go preRegisterAwaitedRequest; awaited_requests table

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Probe debugging-guide entries #24–#28
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-12 "Debug guide updated … 016_debugging_guide_v2_46.md"; 2026-06-13(g) "Debug guide v2_48 … #27 invented-interface … #28 agent_definitions UNIQUE(type,version)".
- **what:** Reusable pitfalls harvested from probe execution: #24 a config/workflow file is only authoritative at its runtime read-path; #25 prove the test harness delivered the intended input before debugging; #26 shell vars need export/prefix and die with the session; #27 invented interface (compiles standalone ≠ satisfies interface — wire to registry early); #28 agent_definitions UNIQUE(type,version) + two look-alike category columns.
- **sources:** traffic_probe_running_notes(27).md#2026-06-12-debug-guide, traffic_probe_running_notes(27).md#2026-06-12-provisioning-ran, traffic_probe_running_notes(27).md#2026-06-13-g
- **relations:** #24 is the stale-artifact class; #27 fixed backend_unreachable; #28 fixed the agent INSERT
- **verify-later:** 016_debugging_guide_v2_48.md entries #24–#28

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### A4/homepage-missing-file — root cause hypothesis evolution to "auto-complete on lost response"
- **category:** debugging
- **status-signal:** superseded
- **status-evidence:** running_notes_14(20) Part 10 (2026-06-03): "Candidate cause (NOT concluded): ...silent non-fast-forward [git race]. Alternatives not ruled out: empty-assembly skip... or a path bug." CATALOGUE(3)→(4) diff (2026-06-04): "*Cause (updated... git race RULED OUT)*: ...empty-assembly case" superseded again by "*Cause PINNED*: ...work item is `complete` with error 'Auto-completed: work verified done despite lost response.'" Running_notes Part 11 confirms: "Root cause: the homepage's content build was dispatched, the handler's response was lost... and the recovery path optimistically auto-completed the work item without verifying the artifact."
- **what:** The homepage (`index`) was `build_status='deployed'`+`stamped` in the DB with zero rendered components and no committed file — three successive hypotheses (git-commit race, empty-assembly/planner-vs-composition gap, and finally the pinned cause) were tested and discarded in turn before landing on: a scheduled task's SQL `pre_query` (`claimed-item-timeout`) auto-completed a claimed work item using loose evidence ("any page on the site updated since claim") after the handler's response was lost to a pod death, without checking that *this* page actually produced components.
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Parts 10–12; CATALOGUE_gamesdesign_post_sync_fix_defects(3).md vs (4)
- **relations:** claimed-item-timeout evidence-check reliability mini-project (below); silent-completion family (page-build-handler, save_page_sections)
- **verify-later:** `migration_claimed_item_timeout_evidence_v2.sql` application state; `v3_site_actions_optionB.patch` deployment.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Sectionless-page silent completion (guide-skinner-box) + durability stack
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** running_notes_15(5) Part 1: "`complete_error` is a `complete_workflow` (SUCCESS) with message *'Content writer skipped — page has no sections defined.'* — the silent-success smell." Part 8: "Wrote `discovery_checks/check_sectionless_pages.go` (new check...)" with an explicit 4-layer "durability stack" logged, item 3 of which is marked "(next, optional for this gap; cleanliness)" i.e. not yet done at time of writing.
- **what:** A page reaching page-build with zero planned sections (`pages.sections=[]`) causes `check_has_ready_sections` to route to `complete_error`, which is a SUCCESS-labelled `complete_workflow` — so a genuinely broken page is marked `complete` and never retried. Root-caused (after correctly ruling out the convergence-union code, confirmed correct) to "the gap is reconciliation: nothing repairs a page in-plan with zero sections." Fix stack: (1) a read-time fallback in `load_page_sections_from_spec_action.go` that copies a same-role sibling's section layout ("skeleton only, not content") when both known sources are empty; (2) a new discovery check `check_sectionless_pages.go` that detects and retriggers stuck sectionless pages (chosen over patching the existing but **dormant** `checkEmptyPageSections`, see below); (3) a workflow-level fix so the genuinely-unrecoverable case routes to a flagged state instead of `complete` — logged as not yet shipped; (4) the broader positive-evidence-completion mini-project (shared with A4).
- **sources:** adoption/running_notes_15_skinner_box_and_adoption_sections(5).md Parts 1–8
- **relations:** dormant discovery-check machinery (below); A4 auto-complete-on-lost-response; FOCUS_page_build_handler_silent_completion.md
- **verify-later:** whether S2 (workflow-level flagged-state fix) ever shipped; `check_sectionless_pages.go` enablement state.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `save_page_sections` content-regression guard laundered into false success — theories falsified in sequence, course-corrected to a second mechanism
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** running_notes_17(16) "index deep-dive": four theories tried and explicitly falsified in order — (1) "Load — 21 rebuilds... starved the dispatcher. FALSIFIED"; (2) "Concurrent production deploy cycled the pods mid-flight. FALSIFIED for index"; (3) "index's build DURATION exceeds the... claim lease. FALSIFIED"; (4) caller-timeout theory, "partially real as a STATUS artifact... but this is noise, not the defect." Landed on: "**Content-regression guard... is the leading mechanism.**" Then a further course-correction: "**COURSE-CORRECTION: a second mechanism — page_components LOCKING**... NEW HYPOTHESIS (at least as plausible as the regression guard)."
- **what:** The homepage (`index`) repeatedly failed to rebuild despite the work item showing `complete` and git successfully committing a file — the committed file was stale (unchanged since 2026-06-06). Root cause hunt discarded four increasingly specific theories (load, concurrent deploy, claim-lease timeout, caller/callee timeout mismatch) before finding `save_page_sections_action.go`'s **content-regression guard** — a real safety check (refuses to overwrite existing deployed content with much-shorter new content) whose error return was silently laundered into `complete_error`, itself a SUCCESS-labelled `complete_workflow`. Before fully confirming this, the investigation surfaced a *second* candidate mechanism discovered via schema inspection — a `page_components` row-locking subsystem with an `auto_lock_on_deploy` trigger — and explicitly walked back single-mechanism confidence pending a discriminating query. Two distinct, independently-real bugs were named regardless of which mechanism fires: (1) the guard's legitimate refusal shouldn't route through `complete_error`; (2) deploy shouldn't proceed (re-render + git-commit) after a zero-row save.
- **sources:** content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(16).md, sections "index reproducibly stale" through "COURSE-CORRECTION"
- **relations:** A4 auto-complete-on-lost-response; sectionless-page silent completion; page-build-handler silent-completion family generally
- **verify-later:** which mechanism (regression guard vs component lock) actually fires on `index`; `page_build_handler_save_failure_visible.sql` application state; `auto_lock_on_deploy()` trigger function body.

<!-- SOURCE: U25_leopardess_social.md -->
### Silent no-op success class (`complete_error` and the ten empty builds)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** NOTES_provocations-index (2026-07-04→07): "SEVEN work items … ALL 'complete', no errors … A success status masked a no-op for two weeks"; HANDOFF §3 "ten builds 'completed' having built nothing"; preventions partially shipped (sectionless_pages enabled 2026-07-10).
- **what:** The defining failure shape of the platform: error paths implemented as successful completions. Canonical case: the planner emitted a page with no sections, and page-build-handler routes zero-sections to a step literally named complete_error — a complete_workflow reporting success ("Content writer skipped"). Diagnostic signature: a work-item result carrying only site_record (healthy runs emit sections_saved + deploy_result). Framework preventions specified: planner invariant (every planned page whose role page-build-handler builds must have ≥1 section, with an explicit role→pipeline map), fail-loudly on the zero-sections path, auditor rules for planned-but-linked pages, post-deploy URL presence checks. sectionless_pages (the exact detector) existed but was enabled nowhere until 2026-07-10.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-index(4).md; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#3, #9.1
- **relations:** discovery check wiring gaps; verify-by-artifact discipline; page section source precedence (the unblock)
- **verify-later:** page-build-handler workflow complete_error step; sectionless_pages check enablement

<!-- SOURCE: U25_leopardess_social.md -->
### Problem-category taxonomy for component/tool defects
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** In active use across every NOTES file in this unit ("Categories:" lines); seed set defined 2026-06-29, extended in practice (assembly-drop, planning-gap, silent-noop-success, cta-graph, css-specificity, method-correction).
- **what:** A shared greppable vocabulary tagging every incident so patterns roll up into the global debugging guide: css-variable-mismatch, empty-shell/mode-b-template, broken-template-slots, content-vs-runtime-mismatch, detool-on-rebuild, js-not-extracted, js-bundle-stale, schema-template-drift — plus organically-added tags. When a category recurs across tools it graduates to a global pattern with a systemic fix (exactly how the empty-shell and visible-content-filter issues surfaced).
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/TOOL_DOCS_convention(3).md#Problem-category-taxonomy; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocation-card(12).md (tags in use)
- **relations:** per-tool travelling docs; debugging guide 016b
- **verify-later:** 016b entries fed from these categories

<!-- SOURCE: U25_leopardess_social.md -->
### Editing-stored-HTML landmines (marker anchoring, hidden-vs-author-CSS, offline edits)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** HANDOFF §8 "all paid for in this thread" (fix_marker_selector.sql, fix_archive_template_display.sql both shipped); VERDICT length-delta verification used twice.
- **what:** Hard-won rules for touching stored templates/HTML: a marker/attribute REPLACE must anchor on the opening tag, never a bare attribute (the attribute string also appears inside the component's own querySelector and a plain replace corrupts it — happened twice); the `hidden` attribute is UA-stylesheet display:none and loses to any author display rule (clone templates render as ghost rows without an explicit [data-…-template]{display:none}); multi-line block removal is dump → edit offline → UPDATE full text → verify by length delta (never multi-line SQL REPLACE of nested markup); better still, emit markers at generation.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#8; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-archive-list.md#2026-07-08; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-index(4).md#2026-07-07
- **relations:** generation-time guards (the prevention); section-editor path
- **verify-later:** n/a (lessons; fixes live)

<!-- SOURCE: U25_leopardess_social.md -->
### Stuck-claim / zombie-handler dispatch noise
- **category:** debugging
- **status-signal:** unknown
- **status-evidence:** RUNBOOK_minilobby dispatch note (2026-07-10): "manual dispatch via 087 needed five passes because page-build handlers intermittently left items stuck at claimed without spawning (survived across the v1.0.1107 deploy)"; also "the known stuck-claim/zombie-handler noise recurred" 2026-07-12. Root cause not established.
- **what:** Recurring operational failure: dispatched items sit at claimed with no spawned handler, or late handler reports mark completed work failed. Recovery is documented (reset to triaged, NULL claim fields, re-run the dispatch pass; close by artifact when the work actually happened) but the underlying cause is unresolved — a live reliability question for the dispatch/spawn path.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#0 dispatch note; docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-12; docs/social001_vonc_tiktok_social/minilobby_task/087_dispatch_work_items_vonc.sh (header)
- **relations:** work-item dedup; leopardess O4 unstick procedure (same class)
- **verify-later:** claim/spawn/call sequence in build-dispatch-loop; agent_error_log around stuck claims

<!-- SOURCE: U26_misc_dirs.md -->
### Workflow monitoring REST endpoints
- **category:** debugging
- **status-signal:** unknown
- **status-evidence:** 004-agent-chassis-architecture.md lists GET /monitor/workflows, /monitor/workflow/{id}, /monitor/stuck?hours=n, /monitor/metrics as built ("Each agent exposes monitoring endpoints") but no later doc in this unit uses them — operational debugging instead goes through psql/db-inspector.
- **what:** Per-agent HTTP monitoring API over orchestration state: list active workflows per client, inspect a workflow's execution path/state, find stuck workflows not progressing for N hours, and aggregate metrics. Complemented by per-step execution_path timing records and execution_metadata counters in the state row.
- **sources:** docs/architecture/004-agent-chassis-architecture.md#monitoring-and-observability
- **relations:** database-backed workflow state; kcat/db-inspector runbook (the surviving practice); current debugging docs (016/016b spine)
- **verify-later:** /monitor routes in chassis HTTP server code

<!-- SOURCE: U26_misc_dirs.md -->
### kcat + db-inspector operational runbook
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** basic_usage/001 and 004 are working command logs (with real outputs pasted, e.g. correlation IDs returned, "0 rows" failure cases) for triggering and tracing workflows in the live cluster.
- **what:** The early ops playbook: scale deployments up/down; inject workflow-start messages via kcat from an in-cluster pod with full header sets; fetch the latest correlation_id from orchestrator_state; watch progress with the db-inspector tool (-watch); trace specific agents by finding spawned instance IDs then grepping shared chassis pod logs (agents don't get dedicated pods); check consumer-group lag, response topics, ServiceAccount job-creation rights, and events for spawned jobs.
- **sources:** docs/basic_usage/001basic_usage.txt; docs/basic_usage/004_debugging
- **relations:** agent spawning; website-builder group; current debugging spine (016)
- **verify-later:** tools/db-inspector, tools/kafka-producer existence; whether runbook matches current namespace/topics

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Debugging assumption checklist (28-item process discipline)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §0, each item tied to a dated real defect through 2026-06-13
- **what:** The distilled pre-change checks: per-action _field conventions; input_mapping required-by-default; error_preview before log grepping; partial DB rows; SQL-immediate-vs-Go-deployed; sibling functions as canonical pattern; token budgets vs structured output; set -u in trigger scripts; jq slurpfile nulls; manual triggers to isolate dispatch-vs-handler; parent/child orchestration rows; `?` placement; \d before SQL; refire-before-refactor; pod-rotation log loss; don't change evidence-proven values; deploy ≠ migration ran; interface widening breaks all importers; prompt_rendered proves input not output; updated_at is not authorship; re-resolve site_id after teardown (zero-row LEFT JOIN = wrong anchor); check design docs for deliberate deferral; output_fields plural; config authoritative only at its runtime read-path; prove the harness delivered input (dash vs bash); env vars vs shell locals + stale deployed copies; read the interface definition; agent_definitions UNIQUE(type,version).
- **sources:** 016 §0 items 1–28 + 016_additions
- **relations:** everything in §9; 016b durable invariants
- **verify-later:** —

<!-- SOURCE: U01_docs024_numbered_core.md -->
### An agent is a DB row; trust default_config over prose; two possible definition sources
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §6.0 with build-dispatch-loop description-vs-config contradiction example
- **what:** Agents live in agent_definitions.default_config.workflow — grepping Go finds actions, not agents. Descriptions can contradict configs (trust the config). agent_definitions may be read from templates_db or clients_db depending on pod — confirm which copy the running pod loads before patching.
- **sources:** 016 §6.0
- **relations:** orchestration state; snapshot conventions
- **verify-later:** which DB each deployment reads definitions from

<!-- SOURCE: U01_docs024_numbered_core.md -->
### LLM step config shadowing (ai_service resolution order; dead temperature paths)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016 §6.6: bug live as of 2026-05-18 (22 of ~60 agents shadowed); structural per-field fix "planned"
- **what:** ExecuteLLMPromptAction resolves ai_service top-level → step-level → StepConfig and stops at first match, so a top-level ai_service shadows every step override (incl. doc-023-style per-step model swaps); max_tokens falls to hardcoded 2048 (tell: output_tokens exactly 2048); step.config.max_tokens sibling is never read; temperature is read ONLY from default_config.temperature top-level (all other locations dead) and llm_call_log.temperature was universally NULL. Fix path: per-field fallback chain + raise floor to 8000 + log sent values.
- **sources:** 016 §6.6
- **relations:** model swap functions; __sent_* write-backs (001(5) suggests later capture landed)
- **verify-later:** whether per-field resolution shipped; llm_call_log temperature now populated

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Timeout chain ordering (claim > call_handler > handler workflow)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §7 with current values and both mis-ordering failure modes
- **what:** claim timeout (30m) must exceed dispatch call_handler (20m) must exceed handler workflow timeouts; otherwise duplicate handlers (claim reset mid-work) or orphaned completions (dispatch gave up early). Idle monitor 3600s fallback; K8s ActiveDeadline 24h ceiling.
- **sources:** 016 §7
- **relations:** claim-lease-too-short reproducible timeouts (v2_49 sub-case b)
- **verify-later:** current values across dispatch/handlers

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Reaper false-positive completions (claimed-item-timeout evidence checks too loose)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016 §9 with confirmed gaswholesalers instance (auto-complete 47 min before the real commit); fix described as small SQL, not marked applied
- **what:** The "verified done despite lost response" branch auto-completes on ANY page updated on the site (not the target page, and updated_at not deployed_at) — treat empty-result + 'Auto-completed' items as untrusted. Correct evidence: p.id = wi.page_id AND deployed_at > claimed_at; needs_rerender/needs_design shouldn't auto-complete this way. Sibling issue: orchestration engine doesn't enforce awaited_requests timeout_at (spawn-handler hangs until reapers paper over it).
- **sources:** 016 §9 claimed-item-timeout + spawn_handler-hang entries
- **relations:** silent-completion family; timeout chain
- **verify-later:** claimed-item-timeout pre_query current form

<!-- SOURCE: U01_docs024_numbered_core.md -->
### jsonb && operator class bug (silent CSS-snippet failure vs hard JS failure)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §9: css path "silently failing the entire time"; JS analog fixed in same change set (May 2026)
- **what:** Postgres has no jsonb&&jsonb overlap operator; `applies_to && $1::jsonb` errored forever, swallowed by a logger.Warn-return-"" handler, so css_snippets never reached any deployed styles.css. Fix: EXISTS + jsonb_array_elements_text. Wider lesson: silent-failure loaders + graceful consumers hide months-old breakage — prefer hard failure when the data is supposed to be there.
- **sources:** 016 §9 jsonb && entry
- **relations:** best-effort-needs-monitoring; audit grep pattern for other && uses
- **verify-later:** loadComponentCSSSnippets fixed in place

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Silent-completion family (trust the artefact, not the status)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016/016b recurring entries; 016b names it the first durable invariant
- **what:** Work items report complete while the work didn't happen, via: result-contract stub (fixed 06-18); content-regression guard error masked by error_step complete_error; pod dying mid-flight (complete with non-empty error); "git committed the file" re-committing stale stored components; zero-planned-sections completing as success. Verify against page_components timestamps + live HTML. Companion rules: completed_at is orchestration END not write instant (trace child orchestrations by page_id in collected_data — trap part 3); intermediate signals (work-item names, pod snapshots, mid-flight tables) lie (trap part 2).
- **sources:** 016 §9 several entries + traps 1–3; 016b invariants
- **relations:** workflow result contract; zero-planned-sections
- **verify-later:** —

<!-- SOURCE: U01_docs024_numbered_core.md -->
### save_page_sections is the sole page_components writer; its section-regex fallback and content-regression guard
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016 §9 tool-pages-never-deploy (fix shipped 2026-05-28, end-to-end verification honest-open); guard masking fix flagged not confirmed applied
- **what:** save_page_sections DELETE+INSERTs page_components (history row written; source_item_id NULL on overwrite path — gap). Its HTML fallback extracted only `<section>` blocks, so `<div class="tool-page">` tool HTML was silently discarded (all tool/game pages n_rendered=0, rerender skips, no file ever committed); fixed by whole-fragment-as-one-section fallback (guarded against full documents). The content-regression guard (new text < existing/4) protects prose but returned errors that complete_error converted to success. Deferred sections' instances are dropped on save (carry-forward pending, cousin of the interactive clobber).
- **sources:** 016 §9 tool-pages entry + guard entry; 016b Part 5/regenerated-section entries
- **relations:** de-tool hazard fix layers; deployed→needs_rebuild flip
- **verify-later:** patched save_page_sections deployed to all three callers

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Presign loop collapse: batch adapter calls over awaited-loop iterations (O(K²) state bloat)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §9 marked DONE + CONFIRMED IN PROD 2026-06-09 (migration 110; 26s vs never-finishing)
- **what:** Every step transition re-persists the whole orchestration state (expanded workflow + collected_data + history), so a K-iteration awaited loop is O(K²) and geometrically slows; the structural fix is one batch adapter call (prepare_object_urls returning all URLs in one reply) — deleting both the race class and the bloat class. Related fix: configOrInput now coerces numeric config scalars (expiry_minutes 3000 was silently dropped by a .(string) assertion).
- **sources:** 016 §9 presign entries
- **relations:** loop mechanisms; envelope race
- **verify-later:** training-launcher def shape (2d state check)

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Hand-applied agent-def migrations have no ledger; re-running an earlier one reverts later ones
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §9 (2026-06-09): re-running 109 silently reverted 110+111; runbook corrected
- **what:** Agent-definition jsonb migrations are hand-applied with no runner/ledger — the live def SHAPE is the only source of truth. A migration is idempotent only vs its own prior application, never vs later migrations on the same object; recover from doubt by checking state, never by re-running. Per-migration state checks (runbook 2d) after every deploy.
- **sources:** 016 §9 re-running-idempotent-migration entry
- **relations:** backup discipline; deploy≠migration checklist item
- **verify-later:** —

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Image doesn't contain the binary (CrashLoop exec not-found ⇒ build/packaging fault)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §9 (2026-06-14): thunder-adapter tag shipped the analyser binary (overwritten Dockerfile, shared digest); third deploy-regression in a row
- **what:** `exec ./X: no such file` means the image lacks X — inspect image contents (docker run ls /app; Image-ID vs Image digest tells tag collisions), restore the Dockerfile, push a FRESH tag (never re-push the poisoned one). Guard: pre-push ls /app or a CI binary-name assertion — "no guard between built and running" is the recurring gap.
- **sources:** 016 §9 CrashLoop entry
- **relations:** deploy≠migration; stale-artifact family (checklist 24/26)
- **verify-later:** CI assertion added?

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Open problem: nav-updater never spawns
- **category:** debugging
- **status-signal:** unknown
- **status-evidence:** 103 "Active Problem" — definition exists/active, topics exist, dispatch generic, yet no pod ever appeared; all nav_drift items claim-timeout
- **what:** nav_drift items route to nav-updater via the generic dynamic dispatch, but no nav-updater pod has ever started and items exhaust claim timeouts. Investigation was open at handoff (2026-04-12); distinct from the nav-link-fixer path.
- **sources:** 103#Active Problem
- **relations:** dispatch loop; missing-handler pattern (different: def exists)
- **verify-later:** whether resolved since; nav_drift item outcomes

<!-- SOURCE: U01_docs024_numbered_core.md -->
### 016b durable invariants + wrong-turns log as a debugging methodology
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v1–v5 changelog; wrong-turns section explicitly kept "so the next pass doesn't re-walk them"
- **what:** Vol. 2 distils the paying-off heuristics (trust artefact; completed_at ≠ write instant; config-key-path no-ops; who writes page_components; 0 rows not decisive; negative inference needs mechanism checked in ALL cases; reuse before rebuild) and logs false leads per arc with the heuristic each violates. Also fixes doc process: the guide had forked across chats; v5 is the explicit merge point.
- **sources:** 016b#Orientation, #Durable invariants, #Wrong turns
- **relations:** 016 §0; travelling docs
- **verify-later:** —

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Zero-planned-sections silent no-op success (planning gap + complete_error anti-pattern)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016b v5 + 2026-07-06 amendment: route confirmed from workflow dump; pages.sections UPDATE proven to change behaviour; guards/planner-invariant fixes listed as prevention, not applied
- **what:** A linked-everywhere page 404'd for two weeks while seven work items completed clean: planner emitted the page with no sections; page-build-handler's zero-ready branch is literally a complete_workflow step named complete_error ("an error path implemented as a successful completion" — diagnostic signature: result contains only site_record); rerender skips no-component pages quietly. Section sources in order: site_specs site_plan aspect → pages.sections; site_plan_sections table is NOT read by builds. Fixes: planner invariant (every page ≥1 section), fail-loud zero-planned guard, rerender warn, auditor rules (active+linked+planned; post-deploy URL HEAD), dynamic-list component vocabulary for archive pages.
- **sources:** 016b#Page build completes having built nothing + amendment
- **relations:** silent-noop-success/planning-gap tags; section-index vocabulary
- **verify-later:** complete_error branch fixed?; planner invariant added?

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Content↔template key-contract drift (system-stats class)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016b Part 5 TRIAGED 2026-06-24, remedy un-applied; fleet-wide (usage_count 22)
- **what:** component-creator rewrote a template AFTER pages were built; stored content_data keys share ZERO keys with the new template placeholders → renders text-empty → visible-content filter correctly drops the section. Remedy: full content rebuild (not page_rerender, which reuses mis-keyed content_data); structural need: component schema changes must trigger dependent rebuilds, or fix writer↔input_schema binding. Diagnostic: diff the two key sets directly (a populated-but-unrendered section is a key-contract check, not a generation failure).
- **sources:** 016b Part 5 + wrong-turn #4
- **relations:** schema-template-drift tag; component regeneration rerender items
- **verify-later:** schema-change-triggers-rebuild mechanism

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Log tables before pod stdout (agent_error_log, llm_call_log as forensic sources)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 "hunting for logs" section; pod-rotation checklist item
- **what:** Persistent DB logs beat ephemeral pod stdout: agent_error_log (every reported error, filterable by context site/domain), llm_call_log (every call incl. failures with error_message). Pod logs vanish on rotation/rollout; zap JSON must be grepped by message string not field=value; logger.Debug is invisible in-cluster (house rule: logger.Info); verify deploys against the artifact (curl/DB), not log presence.
- **sources:** 016#hunting for logs; 016b#Verifying a deploy
- **relations:** silent-completion; assumption checklist 3/15
- **verify-later:** —

<!-- SOURCE: U01_docs024_numbered_core.md -->
### SQL template-surgery method (needle-gate) and Postgres verification pitfalls
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v4 entry distilled from the scheme-to-components workstreams
- **what:** Safe in-DB template mutation: needle-gate read (LIKE booleans + occurrence counts, expectations counted mechanically not recalled), shell .bak of the column, guarded idempotent exact-string replace (or anchored regexp_replace), RETURNING checks, value-agnostic rollback. Pitfalls: regex quantifier bound ≤255; substring-with-parens returns first capture group; gradient-embedded hexes escape colon-anchored classification; % in needles breaks LIKE gates (use position()).
- **sources:** 016b#SQL verification pitfalls
- **relations:** marker-REPLACE anchoring entry (anchor attribute REPLACEs on the opening tag, not the bare attribute — the querySelector corruption bug)
- **verify-later:** —

<!-- SOURCE: U01_docs024_numbered_core.md -->
### sites.status is informational (never scope by status='active')
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v4 entry with the silently-dropped-site incident
- **what:** UpdateSiteStatusAction vocabulary draft/building/review/published/deployed/archived/error; 'active' is legacy hand-written; nothing filters on it — dispatch keys on site_work_items. Enumerate GROUP BY status before any blast-radius query. Reuse-gate corollary: check pg_proc/pg_trigger before adding helpers (shared set_updated_at exists).
- **sources:** 016b#sites.status
- **relations:** zero-rows-not-decisive
- **verify-later:** —

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Kafka trigger payload discipline (flat single-line JSON here-strings)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "kcat heredoc was silently mis-routing messages to a 'No-op scheduled task' handler … Documented as permanent ops pattern in 016_debugging_guide_v2.md §9" (2026-04-23)
- **what:** Multi-line heredocs mangle kcat JSON payloads silently (routing falls through to no-op handlers with input_data null). Use `<<<'{…flat json…}'` here-strings or jq -nc. Related manual-trigger pattern: psql jsonb_build_object → pipe to kcat with standard headers, used to trigger handlers directly when dispatch is blocked.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4f v2 smoke retest, #14; FOCUS_dispatch_diagnostic(4).md#Workarounds
- **relations:** dispatch workarounds; debugging guide §9
- **verify-later:** 016_debugging_guide §9 entry

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Discovery-checks list maintenance and the workflow-replace landmine
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "Closed — investigation found no overwriter" (2026-04-19); the jsonb `||` append pattern recommended; updateAgentWorkflow risk "Currently safe because nothing fires it"
- **what:** The suspected "checks keep falling off discovery agents" was manual SQL replacing the whole checks array (a stale in-code example being copy-pasted); the safe pattern is jsonb array append. Latent risk logged: updateAgentWorkflow does jsonb_set of the ENTIRE workflow subtree — when an automated improvement-proposal generator ships, partial proposals will silently erase workflows unless converted to deep-merge.
- **sources:** HANDOFF_2026-04-19_component_linking_news_template_discovery_checks.md#3, #4
- **relations:** improvement_proposals (empty table); ApproveImprovementAction
- **verify-later:** updateAgentWorkflow (context line ~61056); stale comment cleanup

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Debugging meta-lessons (evidence discipline)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** codified into 016_debugging_guide entries: §0 item 19 added 2026-05-26; dispatch doc "Lesson learned" 2026-05-15; naming FOCUS "Tests document behaviour, not intent" (2026-05-17)
- **what:** Recurring investigation disciplines earned across these sessions: grep the whole codebase for the verb (triage/promote/claim) before concluding a writer doesn't exist; a LIKE on prompt_rendered proves what the model was told, never what it did — read response_text; check the guide before generating fresh hypotheses; design tests to falsify; tests assert what a function does, not what was intended; grep chassis logs by the `caller` field (msg gets truncated); logger.Debug is invisible in production; spawned pods are app=dynamic-agent with 600s idle timeout so capture logs before they evaporate; work the smallest useful step; trust suspicion of implausible numbers.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Lesson; HANDOFF_2026-05-26…md#wrong-turns; HANDOFF_2026-04-18_enrichment…md#greps, #false-starts; FOCUS_naming_conventions…md#flags; FOCUS_finetuning…(13).md#14
- **relations:** debugging guide 016 (the canonical home)
- **verify-later:** 016_debugging_guide §0 items

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Deployed-binary-predates-disk failure class
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Notes (Tm): "Fork RESOLVED: extraction sound … → the on-disk code cannot produce the July-2 escalation → deployed predates disk; the skip_field fix exists and never shipped."
- **what:** A named diagnosis class: observed behaviour contradicts a correct code read because the running pod's image predates the working copy — the fix exists on disk and never shipped. Diagnostic: `git log -1 -- <file>` vs the running pod's image age; remedy: deploy the working copy. Sibling lessons from the same threads: verify the running image contains an edit before debugging it ("success path silent"), and prefer a forward test (one clean build + read both render and stored data) over forensic reconstruction of overlapping rebuild windows.
- **sources:** running_notes_scheme_to_components(55).md#Tl #Tm #Tt; RUNBOOK_scheme_to_components(50).md#W7-FINDINGS; w8_07_fresh_index_build.sql (the forward-probe pattern)
- **relations:** chassis deploy model; plan_sections deferral (the instance).
- **verify-later:** 016b guide entry for this class.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Postgres/SQL pitfall class (016b lessons)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Notes (St): "Guide updated → `016b_debugging_guide_7.md` (v4 log + three §9 entries …)"; each pitfall has an owned in-thread instance.
- **what:** The accumulated instrument-error catalogue this thread wrote into the debugging guide: Postgres ARE regex quantifier bounds max at 255 (`.{0,420}` is invalid — use substr+position); `substring(… from '(pattern)')` returns the FIRST CAPTURE GROUP, not the match; LIKE treats a needle's literal `%` as a wildcard (use position()); regexes like `background:\s*#` miss gradient-embedded hexes; a `0 rows` result is not decisive until the query and live state are checked (applies to one's own verification queries too); probes that grep for a key string are blind when objects are UUID-named; naive brace-counting false-fails on regex literals. Plus data-vocabulary lessons: sites.status vocabulary is draft/building/review/published/deployed/archived/error with legacy 'active'/'system' strays — never filter blast radius on status='active'.
- **sources:** running_notes_scheme_to_components(55).md#Sr #Ss #St #Sv #Sw #Tu #Ue; w2_02_verify_fixed.sql; RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS (3f)
- **relations:** SQL needle-gate surgery; debugging guide 016b (the home doc).
- **verify-later:** 016b_debugging_guide_7.md §9 entries.

<!-- SOURCE: U04_idea_uk.md -->
### Claimed-item timeout evidence gate (failed vs false-completed)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "the evidence-gate (migration_claimed_item_timeout_evidence_v2, live since 2026-06-04) refuses to complete a 0-component page… That is the gate working as intended."
- **what:** The dispatch gate that distinguishes the two failure signatures of the same coordinator bug: without it a stubbed page false-completes (gamesdesign); with it, a 0-component page's claim is reset and retried until attempts exhaust → an honest `failed`. Used here as diagnostic doctrine: don't conflate a silent stub with a genuine handler hang — read the parent's collected_data response to tell them apart.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (empty-index diagnosis); idea.uk/016_debugging_guide_v2_32(1).md (claimed-item sections)
- **relations:** coordinator result contract; work-item state machine.
- **verify-later:** the evidence-check migration in the chassis migrations.

<!-- SOURCE: U04_idea_uk.md -->
### LLM API shape disciplines (server-tool injection, per-model thinking shapes, long-call timeouts)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Items 24–27 added to the debugging guide from the 2026-06-04 validation run ("three API bugs found and fixed during validation").
- **what:** Standing disciplines from live API breakage: (24) a hosted server tool may auto-inject its documented dependency — declaring it yourself collides (web_search v2 injects code_execution; the 400 names the conflict); (25) the same capability has different wire formats across model generations — newer Opus-class models take adaptive thinking + output_config.effort while Sonnet 4.6 takes manual budget_tokens, so helpers must branch per model (and Opus also rejects non-default temperature/top_p); (26) long agentic calls (high effort + N searches) send no headers for minutes — size client timeouts for the worst-case step (180s→900s; streaming is the durable answer); (27) always confirm the current request shape from live docs before coding, especially after a model bump — remembered shapes are guesses and each failed round-trip costs real spend.
- **sources:** idea.uk/016_debugging_guide_v2_32(1).md (items 24–27); idea.uk/DEVELOPMENT_RUNBOOK(3).md#A1 (acceptance); idea.uk/running_notes(63).md (2026-06-04 checkpoints)
- **relations:** engine upgrade; llm-quality-testing; model-infrastructure.
- **verify-later:** engine.go usesAdaptiveThinking + client timeout.

<!-- SOURCE: U05_content_quality_linking.md -->
### Operator/assistant division-of-labour + DB-change safety conventions
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Standing-rules blocks repeated verbatim across HANDOFF_2026-06-09(2), HANDOFF_2026-06-15(2) §0, HANDOFF_page_pipeline(11) §0-1.
- **what:** The working covenant these threads ran under: the assistant reads code/writes deliverables with no cluster/DB access; the human runs all SQL/kubectl/builds. Safety conventions: snapshot before any DB change (snapshot_agent()/revert_agent() house helpers for agent rows; `CREATE TABLE <t>_bak_<tag> AS SELECT` for data, short names, in-txn, rollback documented); fresh `\d` before SQL; every template replace() verified by UPDATE 1 + a flag flip (a stuck flag = whitespace mismatch silently no-op'd); check_linking_sql_applied.sql as the idempotent "which SQLs are in" orient step; workflow changes are DB-only and immediate vs Go changes needing an image roll + image_tag bump; tags of co-deployed agents must move together (a lagging resolver tag = permanent silent fallback); don't roll the chassis image while a rebuild batch drains.
- **sources:** HANDOFF_2026-06-15(2).md#0; HANDOFF_page_pipeline(11).md#0-1; check_linking_sql_applied.sql; RUNBOOK_linking_phantom_fixes(7).md
- **relations:** debugging heuristics; documentation system (runbook discipline).
- **verify-later:** snapshot_agent/revert_agent function defs; agent_backups behaviour.

<!-- SOURCE: U05_content_quality_linking.md -->
### Debugging heuristics harvested into the 016 guide
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Guide bumped v2_31 → v2_45 → v2_49/50 → v2_56/57 across this unit's sessions, each with named new entries; 016 "CLOSED at v2_56", continued in 016b.
- **what:** The unit's investigations were systematically distilled into durable heuristics: trust rendered HTML/DB state over work-item status; a 0-rows/false result is not decisive until the query itself is verified; matching updated_at is not authorship — confirm the action writes the column; work-item completed_at is the orchestration END, not the write instant — trace by orchestration; an empty LEFT JOIN means wrong anchor id; never carry a site_id across a teardown; "git committed ≠ new content"; two rebuild routes; text-heuristic blind spots (prose guards miss markup/JS loss); psql vs shell variable syntax traps; Kafka consumer-group loss after topic wipes (restart-to-rejoin, park at latest, never replay).
- **sources:** running_notes_14(26).md#principles sections; running_notes_17(21).md; NOTES(44) passim; HANDOFF_page_pipeline(11).md#10
- **relations:** documentation-system (guide versioning); every defect thread here.
- **verify-later:** 016_debugging_guide_v2_56/016b content (owned by the debugging unit).

<!-- SOURCE: U06_finetuning.md -->
### Wrong-binary adapter image incident and the built-vs-running guard
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-14(8): "thunder-adapter:v1.0.1063 actually contains the analyser-adapter binary… Pattern (third deploy-regression in a row)… No guard between built and running. Logged in debug guide v2_47."
- **what:** An overwritten Dockerfile shipped the analyser-adapter binary under the thunder-adapter tag; the pod CrashLoopBackOff'd for ~31h and every provision parked runs at `pending`. Named as the third consecutive "the deploy didn't ship what I thought" regression (109 re-run revert; chassis/adapter tag confusion; Dockerfile overwrite). Prescribed guards: per-build `docker run --rm --entrypoint ls <image> -la /app` before push, never re-push a poisoned tag, and structurally a CI step failing the build if the expected binary is absent — the deploy-side sibling of the migration 2d state check.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-14-8; working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#2026-06-14-update
- **relations:** hand-applied migrations lesson; deployment essentials
- **verify-later:** whether the CI image-content guard was ever added

<!-- SOURCE: U06_finetuning.md -->
### Send-before-register await race and preRegisterAwaitedRequest
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-09: "Race fix works [verified-log]. Every presign_checkpoints_iter_N_presign_one logged ClaimAwaitedRequest: status_before=waiting … claimed:true"; recorded as the fourth cause of stuck-`waiting` awaits in debugging guide v2_36 §9.
- **what:** Local dispatch actions produced the adapter request and returned await_response:true BEFORE the coordinator inserted the awaited_requests row; a fast (~1s) reply beat the insert, ClaimAwaitedRequest found no `waiting` row, the reply was dropped, and the timeout handler re-dispatched forever with fresh request_ids (RetryVersion pinned at 0). spawn_agent/call_agent don't race because they call `preRegisterAwaitedRequest` (register-before-send, ON CONFLICT DO NOTHING). Fix: the dispatch pre-registers with the same request_id it uses everywhere — one row, one timeout owner; caveats: the helper hardcodes a 120s timeout that wins over step config, and the per-request timeout goroutine is skipped (background expiry sweep is the net). Moving stall point ⇒ race is the diagnostic heuristic.
- **sources:** working/phase5/HANDOFF_2026-06-06_checkpoint_upload_loop_await_race(2).md; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-08,#update-2026-06-09; working/docubundle/CONTEXT_PACK_thunder_checkpoint_race(1).md
- **relations:** O(K²) loop cost (found immediately after); reply-topic rules; awaited_requests machinery
- **verify-later:** preRegisterAwaitedRequest call in thunder_prepare_object_url_dispatch.go and the batch/resume dispatches

<!-- SOURCE: U06_finetuning.md -->
### O(K²) loop state-bloat and the batch-presign replacement
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-09(3): "Full launcher path completed in ~26s… ONE batch await… returned all 40 ckpt PUT URLs… Contrast the retired loop: Version 86 / still at iter_9 nine minutes in. The O(K²) class is gone."
- **what:** Every awaited loop substep re-persists the full orchestration state — the expanded ~80-substep workflow with verbose descriptions, growing collected_data, and ProcessingHistory — so a 40-iteration awaited loop costs O(K²) (iter_0-4 ~2-3s, iter_8 ~100s, then Kafka i/o timeouts) while a GPU bills throughout. Structural cure, not tuning: replace the per-item awaited loop with one batch adapter call (`prepare_object_urls`: keys[]→ordered urls[], reusing the single ObjectURL primitive per key), one await, one persist, no flatten step (migration 110). General platform lesson: awaited loops over cheap local operations are an anti-pattern; batch at the adapter.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-09—(3); working/phase5/110_training_launcher_batch_presign(2).sql (header); working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#2026-06-09-update
- **relations:** send-before-register race; loop_complete convention (every production loop ends on an explicit loop_complete substep — checked against all 11 production loops); durability manifest
- **verify-later:** orchestration state persistence cost in coordinator; whether other awaited loops exist at risk

<!-- SOURCE: U06_finetuning.md -->
### Hand-applied agent-def migrations: no ledger, re-run reverts, 2d state check
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-09(7): "re-running 109 silently REVERTED both [110 and 111]… A migration is idempotent only against its OWN prior application, never against LATER migrations that mutate the same object… There is NO migration runner."
- **what:** The flywheel-C def migrations (102–111) are hand-applied jsonb mutations to agent_definitions with no schema_migrations ledger — the def's live shape is the only "did it run" truth. Consequences codified: never re-run an earlier migration "to make sure" (it reverts later ones); run a per-migration state-check query (RUNBOOK 2d) after every deploy and before any launch; back up defs with the sanctioned `snapshot_agent()`/`revert_agent()` (hand-rolled CREATE TABLE backups collide with the existing agent_definitions_backup — discover DB helpers with `\df` first). Optional future hardening: a migration runner or applied_migrations log.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-7,(3); working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#2026-06-14-update
- **relations:** wrong-binary incident (same "shipped what I thought?" family); model swap/revert functions (snapshot_agent reuse)
- **verify-later:** RUNBOOK 2d query vs live launcher def

<!-- SOURCE: U06_finetuning.md -->
### agent_definitions source-of-truth: clients_db, not templates_db (for the rich schema)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-03 ~17:3x CORRECTION: "templates_db.agent_definitions has the OLD schema… holds only the 8 original website-builder agents… PIN (corrected): for the flywheel-C agent_definitions, always read AND patch clients_db."
- **what:** agent_definitions exists physically in BOTH clients_db and templates_db; the architecture doc's "source of truth is templates_db" refers only to the legacy website-builder catalog (old schema, no version column). The chassis loader (filters is_active/is_snapshot, ORDER BY version) can only run against clients_db's rich schema — so all flywheel-C and modern defs live there. This whipsawed twice in one day (103 first applied to the wrong DB, then the "always templates_db" pin issued and then reversed) — a live example of doc-claims diverging from code, and of why the clients_db copy of one def can silently diverge from the live one.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-03; working/phase5/103_call_data_preparer_optional_inputs.sql (header carries the superseded templates_db guidance); working/phase5/104_provisioner_output_fields_and_launcher_mapping.sql (header carries the correction)
- **relations:** hand-applied migrations; documentation-system (stale doc line in 002_system_architecture.md)
- **verify-later:** chassis definition-loader query; 002_system_architecture.md wording

<!-- SOURCE: U06_finetuning.md -->
### CLI/ops data-transfer pitfalls (kcat heredoc, COPY-vs-psql, kubectl exec/cp)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** PATCH_2026-05-06 (both bugs validated + corrected command); FOCUS(25) §2.4f v2 smoke retest (kcat heredoc mis-routing); HANDOFF 2026-04-23 lesson 7.
- **what:** A cluster of verified transfer traps: (1) Kafka trigger JSON must be flat single-line via here-string — multi-line kcat heredocs mangle payloads silently and route to a No-op handler; (2) `COPY … TO STDOUT` is not JSON-safe for jsonb (double escape layers) — use `psql -tAXc` with plain SELECT for JSONL; (3) `kubectl exec -i` without consumed stdin sporadically truncates stdout (1716/1958 rows, "next reader: unexpected EOF"); (4) `kubectl cp` truncates large files silently — use `exec cat > local`; (5) `tnr scp` of directories nests `{dest}/{source_basename}/` both ways.
- **sources:** working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted_PATCH_2026-05-06.md; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#14; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons(4)
- **relations:** dataset pull path; 016 debugging guide §9
- **verify-later:** 01_pull_dataset_from_postgres.sh uses the corrected form

<!-- SOURCE: U06_finetuning.md -->
### configOrInput numeric config coercion (expiry_minutes silently dropped)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-09(5): "expiry_minutes override — FIXED… configOrInput read config via Config[name].(string), so the JSON-number 3000 failed the assertion → fell through → adapter default". Debug guide v2_43.
- **what:** The shared configOrInput helper type-asserted config values to string, so JSON-number config (expiry_minutes:3000, timeout_seconds) silently fell through to defaults — presigned PUTs came back at 24h instead of 50h. Fixed with a `coerceConfigScalar` (string/float64/json.Number/int/bool). Class lesson: shared config readers must coerce scalars, and a numeric setting "applied" in a def is only proven by observing the effect (X-Amz-Expires on the URL).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-4,(5)
- **relations:** presigned data plane expiry caveat; launcher dispatch family
- **verify-later:** coerceConfigScalar in thunder_ssh_exec_dispatch.go

<!-- SOURCE: U06_finetuning.md -->
### Scheduler-fired chassis-resident observability gotcha (owner_agent_type='generic')
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** STATUS 05-12(1) architectural follow-ups: "filtering orchestration_states by owner_agent_type MISSES top-level chassis-resident workflows, which are owned by 'generic'".
- **what:** Scheduler-fired agents that run in the generic chassis (thunder-reaper, build-pipeline-trigger, etc.) have orchestration_states.owner_agent_type='generic'; the real agent type lives at `collected_data->'config'->>'agent_type'` and orchestration_name follows `sched-<task>-<ts>`. Filter on those instead. Related cosmetic anomaly, unresolved: a stale non-DB agent_config stub (old reaper-style no-op) persists in message envelopes across redeploys while the full WorkflowPlan executes — source of the cached representation never found.
- **sources:** working/flywheel_docs/STATUS_thunder_adapter_2026-05-12(1).md#6; working/phase5/NOTES_phase5_training_launcher_running(45).md#stub-source-narrowed
- **relations:** monitor testing; debugging guide
- **verify-later:** where the stale agent_config envelope field loads from

<!-- SOURCE: U06_finetuning.md -->
### Kafka topic-creation race self-heal (transient "Topic not yet on broker")
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-06: "Transient `Topic not yet on broker` for the launcher .responses topic self-healed on attempt 2 (topic-creation race) — normal."
- **what:** Per-spawn child topics (`job.<id>.requests`, per-agent responses topics) are created on demand; a first-publish race against broker propagation produces a transient failure that retries resolve. Recorded so it isn't chased as a real fault. Contrast: a *permanently* missing topic (Strimzi auto-create off) fails every attempt — the distinguishing signature is self-heal on retry.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-06; working/flywheel_docs/FOCUS_adapter_design(3).md#required-cluster-resources
- **relations:** adapter deployment essentials (KafkaTopic CRDs)
- **verify-later:** topic auto-creation settings for spawned-agent topics vs adapter topics

## Proposed NEW categories

None. Everything in this unit fits existing seed slugs — predominantly `finetuning-flywheel`, with `model-infrastructure` (Thunder/Ollama/endpoint/swap), `adapters`, `storage-architecture` (presign/credential boundary), `development-guide`/`debugging` (chassis contracts and failure signatures), `business-strategy` (finetuning.uk), `diagnosis-loop` (docubundle), and `documentation-system` (epistemic tagging).

## Cross-cutting flags for stage 2

- Hardcoded Thunder API bearer token committed in `working/flywheel_docs/ssh_probe.sh` — credential hygiene check.
- Persistent open items to verify: monitor schedule enabled?; first RUN_SH_DONE + final adapter in B2?; orphan-sweep built?; model-trainer call_agent fall-through fixed?; validator Tier-2 coverage extended?; iter_1 ever trained (fp16, 2-epoch, `<no value>` filter)?; any production model swap executed?

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Loose dispatch item-status semantics (complete ≠ done)
- **category:** debugging
- **status-signal:** aspirational
- **status-evidence:** "loose dispatch item-status semantics documented across the investigation (complete-at-dispatch, errors-in-complete, status-change without a timestamp bump, parent-topic-vanished noise) — worth a pass when convenient" (RUNBOOK(49) Part E Hygiene); seven dated sightings in NOTES.
- **what:** A documented defect class in the dispatch loop's work-item bookkeeping, observed seven times: items marked 'complete' at dispatch while the child orchestration runs or fails later; the child's full error text stored in the `error` column of a 'complete' item; status transitions that don't bump updated_at; batch claim stamps shared across differently-fated items; parent fire-and-forget topic lifecycle polluting child completions ("topic partition not found"). Operational rule derived: never trust item status as proof of work — verify the artefact (band stamp, render md5); agent_error_log (occurred_at) outranks status. Fix parked as hygiene.
- **sources:** NOTES(43).md §9i, §9l, §9m, §9aa, §9ac, §9ax, §9bd; RUNBOOK(49).md Step 9 reading guide + Part E
- **relations:** work-item dedup; F2 methodology (discriminator ordering); auto-escalation.
- **verify-later:** build-dispatch-loop status handling; whether items get failure statuses on child errors.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F2 tiered guard-verification methodology (unit → integration → live keep/reject fixtures)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "F2 COMPLETE: Tier 1 unit ✔, 3a preservation ✔ (×2 regens, md5-verified template change), 3b reject ✔ (live firing, three-level visibility, zero mutation)" (NOTES §9o).
- **what:** The verification pattern used to prove F1 without touching live shared components: Tier 1 deterministic unit tests of the guard logic (including the real incident's rename case); Tier 2 DB-backed reject-path test (folded into Tier 3 when no harness existed); Tier 3 end-to-end on throwaway zzz-* components — a KEEP fixture proving preservation-by-instruction (non-guessable check: template md5 changes while fields hold) and an intentionally INACTIVE REJECT fixture exploiting the store-vs-loader is_active divergence to force a rename and observe the guard fire live with zero mutation. Also codified the discriminator ordering (agent_error_log > pod logs > never item status) and prompt cleanup of leftover fixtures.
- **sources:** NOTES(43).md §9f, §9h, §9k–§9o; RUNBOOK(30) family (Step F2 tiers)
- **relations:** F1 guard; F4 (discovered by 3b run 1); loose status semantics.
- **verify-later:** store_generated_component_guard_test.go; zzz fixtures fully cleaned.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Needle-gate SQL template surgery pattern (and its catalogued pitfalls)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Applied on every template mutation W1–W3e with gates, RETURNING checks and verify regions; pitfalls promoted into 016b guide v4 ("count expectations mechanically from the dump… never from memory").
- **what:** The method for mutating shared templates/configs safely by SQL: dump + shell backup first; a gate query asserting exact needles (booleans) and mechanical occurrence counts derived by grep from the dump (mismatch = drift OR mis-derived expectation — stop); anchored exact-string or backreference replaces (multi-line needles to disambiguate repeated strings); guards for idempotency; RETURNING post-conditions; separate verify file; value-agnostic rollback file. Catalogued Postgres pitfalls: regex quantifier bound ≤255; substring() returns the first capture group; LIKE-wildcard `%` inside needles (use position()); `\set ON_ERROR_STOP on` when statements depend on earlier ones; run SQL as files, never pasted.
- **sources:** RUNBOOK_scheme_to_components(18).md W1–W3e blocks + RESULTS; running_notes(22).md Sr, Sv, St
- **relations:** prompt-migration convention (same family for jsonb); debugging guide 016b (where the lessons were codified).
- **verify-later:** 016b guide entries; w*_*.sql files referenced (outside unit).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### R6c artifact-forensics method: cache-busted, metric-consistent comparisons
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "md5sum: gd.html == gd2.html… ONE artifact all along. OWNED: my stale-cache story AND the earlier '4-of-8 mis-assembled' reading were metric artifacts" (NOTES §9al).
- **what:** Lessons from the gripper-detail "blank page" false trail: compare live artifacts only with identical metrics (a data-component inventory vs a class grep counted different things and manufactured a mis-assembly story); md5 the fetches before concluding stale-cache; distinguish 404/200-empty/200-styled-invisible with curl size + head; visually-blank ≠ missing content (fallback-vars insight: content present but dark-on-dark). The eventual truth (theming, not assembly or deploy) reshaped Part D.
- **sources:** NOTES(43).md §9af–§9al; RUNBOOK(49).md Part B
- **relations:** R6f (the real mechanism); assembly membership model; needle-gate pattern (same mechanical-counting ethos).
- **verify-later:** n/a (method; instances cited).

<!-- SOURCE: U08_travelling_docs.md -->
### max_tokens placement rule — dead config outside ai_service
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** FYI second addendum 2026-07-10: "verdict step max_tokens was DEAD CONFIG … Fixed on both agents (snapshot first)".
- **what:** `execute_llm_prompt` reads `max_tokens` from the agent's top-level config or from INSIDE the step's `ai_service` block — never from the step-config root, where several agents had it; the Anthropic client then silently defaults to 2048 output tokens. A truncated verdict JSON parses to UNVERIFIABLE. Standing grep: `config.max_tokens` outside `ai_service` is dead wherever execute_llm_prompt is the action.
- **sources:** FYI_from_fixloop_2026-07-10_verdict_prompt_symptom_check.md#second-addendum
- **relations:** silent-no-op config-path heuristic (016b durable invariants); execute_llm_prompt shared action.
- **verify-later:** ai_actions.go:252-256 max_tokens resolution; remaining workflows with root-level max_tokens.

<!-- SOURCE: U08_travelling_docs.md -->
### agent_error_log is the FIRST read
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v6 entry (2026-07-08), promoted after it settled the tool-generation blocker in one query.
- **what:** Step failures persist to `agent_error_log` (orchestration_id — TEXT not uuid — step_name, action, error_message, error_code, context) and outlive the pod. Read it first, filtered by orchestration_id; only then pod logs (may be reaped) or collected_data (may be enormous). `current_step` from polling is a sample, not an attribution (a 120s poll blamed the LLM step when save_tool failed); a terminal step's success_message can name the wrong phase.
- **sources:** 016b_debugging_guide_7_3_(7).md#agent-error-log-entry; HANDOFF_2026-07-08…md#§3,§5; RUNNING_NOTES_travelling_docs(39).md#rev29
- **relations:** schema drift incident; two failure envelopes; 0-rows rule.
- **verify-later:** agent_error_log schema (orchestration_id type).

<!-- SOURCE: U08_travelling_docs.md -->
### Code-ahead-of-DB schema drift (SQLSTATE 42703, latent until first caller)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Root-caused 2026-07-08 (create_tool_component vs missing content_components provenance columns, latent ~2 months since 2026-05-16); fix applied + proven 2026-07-09.
- **what:** A binary referencing new columns deployed before its migration; nothing fails until the rare code path is called. Detection: the failing INSERT's own comment names the missing migration; a last-successful-call latency probe distinguishes latent drift from fresh regression. Fix pattern: MIRROR column types dynamically from the table the code says it mirrors (format_type/pg_attribute + ADD COLUMN IF NOT EXISTS), additive/nullable/idempotent. The canonical migration file existed but was parked in a docs folder, never renumbered into the migrations path — the exact mechanism by which a deploy skips a migration (one motivation for the migrations system). Standing pre-deploy check: grep the diff for new column names and assert each exists in production.
- **sources:** 016b_debugging_guide_7_3_(7).md#schema-drift-entry; HANDOFF_2026-07-08…md#§3; RUNNING_NOTES_travelling_docs(39).md#rev29,#rev30
- **relations:** migrations system; content_components provenance columns (migration 133); "provenance stamps the chassis".
- **verify-later:** sql_for_agents/133_add_component_provenance.sql vs the docs019 design copy.

<!-- SOURCE: U08_travelling_docs.md -->
### Prompt-template vs config-path resolvers (TEMPLATE_FIELD_ERROR)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v7 entry; root cause of the missing first auto-PLAN (rev 32); latent same-class bug in Task-4 templates caught before it fired.
- **what:** `execute_llm_prompt` with `output_format: text` hands the prompt template the BARE string (`{{.X}}`); with `json` a map (live form `{{.X.result | toJSON}}`); action CONFIG field paths are a different resolver and keep `.result`. Never reach an unverified nested key from a template — dump whole objects with `| toJSON`. A render-time error fires before tokens are spent and, with error containment, the workflow "succeeds" while the step's product is missing (reading rule: normal terminal + missing downstream artefact = contained step failure).
- **sources:** 016b_debugging_guide_7_3_(7).md#template-entry; RUNNING_NOTES_travelling_docs(39).md#rev32; HANDOFF_2026-07-09_recreation_and_chassis_1_.md#§2
- **relations:** docs-never-fail containment (masking effect); seam rule.
- **verify-later:** template data shaping in ai_actions.go by output_format.

<!-- SOURCE: U08_travelling_docs.md -->
### EXECUTING_STEP forever = the worker died (OOMKill triage), superseding stall/leak readings
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v8 rewrite 2026-07-09; the earlier v5-era "error containment does not protect against a HANG" entry and the slow-leak hypothesis are explicitly superseded on the evidence trail kept in RUNBOOK.
- **what:** `orchestration_states` is written BY the worker: a dead pod (OOMKill exit 137, eviction, panic) writes nothing, so the row freezes at EXECUTING_STEP and `since_s` measures time since the crash. Triage order: RESTARTS column → describe pod Last State → `logs --previous` (capture crash logs IMMEDIATELY — a ReplicaSet replacement erases them). Probe suspected-stalled dependencies with a bound (`curl -m 5`) before assuming a hang. Related-but-distinct: genuine stalls from missing context deadlines deserve fixing as hygiene. The arc walked through three wrong hypotheses (stall → missing deadline → slow leak) before the real cause (chunkContent loop), each correction documented rather than discarded.
- **sources:** 016b_debugging_guide_7_3_(7).md#executing-step-entry; RUNBOOK_travelling_docs(38).md#superseded-incident-block; RUNNING_NOTES_travelling_docs(39).md#rev34,#rev35,#rev36
- **relations:** chunkContent bug (the answer); containment-limit corollary.
- **verify-later:** n/a (operational pattern).

<!-- SOURCE: U08_travelling_docs.md -->
### chunkContent() infinite loop — the OOM root cause, fixed with timeout regression tests
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "THE OOM ANSWER (closes the incident chain for good)" — confirmed 2026-07-10; fix deployed v1.0.1104; proof run 05d1fc97 with 0 pod restarts.
- **what:** `chunkContent()` in rag_actions.go never terminated on content > chunk_size: the final chunk ends at len(content), `start = end - overlap` steps BACKWARDS, the same tail appends forever → 2Gi in seconds; content ≤ 1000 chars returned early, hiding the bug for weeks (both OOMKills were PLAN-sized bodies through index_plan). Fixed with a final-chunk break + forward-progress guard and four regression tests with a 30s timeout that catches loop regressions. Durable class rule: content-below-threshold early returns can hide a non-terminating path; "a proof run is a probe — fire proofs early" (the 139 proof run found the real cause within the hour).
- **sources:** RUNBOOK_travelling_docs(38).md#task-6; RUNNING_NOTES_travelling_docs(39).md#v1.0.1103-proof-run,#fix-140-141; HANDOFF_2026-07-10…md#§1,§4
- **relations:** tool_docs indexing (unblocked); migrations 140/141; EXECUTING_STEP pattern.
- **verify-later:** rag_actions_chunk_test.go; chunkContent forward-progress guard.

<!-- SOURCE: U08_travelling_docs.md -->
### kcat -P is line-delimited — single-line trigger bodies
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Run 464102f4 post-mortem (rev 45); 086 and 087 scripts enforce single-line bodies.
- **what:** A pretty-printed JSON body piped to `kcat -P` becomes one message per line; the chassis can then marry your headers to a NEIGHBOURING message's body (observed: our correlation id completing "after 0 steps" holding a scheduler no-op's body — also flagged a chassis stale-buffer wrinkle worth a look). Trigger bodies must be compacted to a single line and scripts must refuse multi-line.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev45-run1; RUNBOOK_travelling_docs(38).md#new-durable-rules; 086/087 script headers
- **relations:** manual kcat trigger scripts; env-prefix trap (sibling).
- **verify-later:** the stale-buffer wrinkle in the chassis consumer (never followed up).

<!-- SOURCE: U08_travelling_docs.md -->
### Env-prefix trap — VAR=x on its own line (or with `;`) never reaches the child
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Cost two 3b.4 runs and one 085 run; 084/086 banners hardened to print the go/no-go tell ("Subject: NONE — will SKIP").
- **what:** Shell variables set on their own line (or terminated by `;` before the command) are not exported to child processes, so triggers silently run with defaults. Correct forms: same-line prefix or `export`. Scripts now print explicit banners of the effective values as the load-bearing tell.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev19,#rev33; RUNBOOK_travelling_docs(38).md#§8
- **relations:** trigger scripts; banner-tell convention.
- **verify-later:** n/a (operational pattern).

<!-- SOURCE: U08_travelling_docs.md -->
### Two failure envelopes — parent COMPLETED ≠ child success
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v5 entry; observed live in runs 1–2 of the 3a arc.
- **what:** A mid-run child failure returns header `status: complete` with the failure in the BODY (`body.status: failed`) — the parent forwards it and completes (a COMPLETED parent row with non-empty `error` = a forwarded child failure); a failed-to-START child sends `status: error_unrecoverable` / `CHILD_ORCHESTRATION_FAILED`. Consumers must check the body, never the header alone; which shape appears tells WHERE the child died.
- **sources:** 016b_debugging_guide_7_3_(7).md#failure-envelopes-entry; RUNNING_NOTES_travelling_docs(39).md#rev11,#rev12
- **relations:** agent_error_log first read; §0-REF reading rules.
- **verify-later:** sendWorkflowResponse / notifyParentOfFailure paths.

<!-- SOURCE: U08_travelling_docs.md -->
### Pod label `agent-type` (hyphen) + multi-pod log attribution
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v5 entry; proven by a working command vs a zero-match selector.
- **what:** The pod LABEL key is `agent-type` (hyphen) while log JSON fields say `agent_type` (underscore) — the underscore selector silently matches zero pods. A type-wide selector spans ALL live pods (idle reaper 3600s), so tails contain residue from earlier runs: attribute every line by orchestration id / pod / timestamp before reading it as current.
- **sources:** 016b_debugging_guide_7_3_(7).md#label-entry; RUNNING_NOTES_travelling_docs(39).md#rev11,#rev13
- **relations:** 0-rows rule; §0-REF.
- **verify-later:** n/a (operational pattern).

<!-- SOURCE: U08_travelling_docs.md -->
### 0-rows rule + gate-evidence capture window + state-dump substitute
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Run-3 closed via the state-dump substitute 2026-07-06; codified in 016b anchorless entry and RUNBOOK §7.
- **what:** 0 rows is decisive only after the query itself AND the run's completion are ruled in (a run that died upstream also produces 0 rows). When a step's NON-firing is the success condition, closure needs a COMPLETED child + the step's explicit skip log line + the 0-count. Skip log lines have a 3600s capture window (idle reaper); past it, a post-completion state dump (ProcessingHistory showing the step executed + terminal status + 0-count) is the accepted substitute. Placeholders are replaced INCLUDING the angle brackets.
- **sources:** 016b_debugging_guide_7_3_(7).md#anchorless-entry (verification discipline); RUNBOOK_travelling_docs(38).md#§7,#stage-3; RUNNING_NOTES_travelling_docs(39).md#rev16
- **relations:** persist_diagnosis_note gate proof; agent_error_log.
- **verify-later:** idle-reaper timeout value (3600s).

<!-- SOURCE: U08_travelling_docs.md -->
### Postgres guard-writing gotchas — RE_DUP_MAX 255, sticky aborted transactions, psql -f over paste
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b entry written from the supersede-migration attempts 1–3 (2026-07-09).
- **what:** Postgres ARE caps bounded regex repetition `{m,n}` at 255 — prefer `strpos`/`substr` in guards (plainer intent, no engine limit). After any in-transaction error the session is stuck (`clients_db=!#`) and ignores everything including BEGIN — `ROLLBACK;` is the only exit; migration files should open with a defensive ROLLBACK and be run with `psql -f`/`\i` (pasting mangles comments and dollar-quoted bodies). A guard that refuses a write can be RIGHT (it blocked an unverified selector) or WRONG (it refused a valid runtime-built selector) — guard design evolved to accept static OR dynamic evidence with a NOTICE saying which path verified.
- **sources:** 016b_debugging_guide_7_3_(7).md#postgres-regex-entry; RUNNING_NOTES_travelling_docs(39).md#rev37,#rev38,#rev39; 0NN_supersede_xp_curve_plan_selectors(2).sql
- **relations:** needle-gate template-surgery pattern; anchor rule (the design insight that came out of guard 1's refusal).
- **verify-later:** n/a (operational pattern).

<!-- SOURCE: U08_travelling_docs.md -->
### Untracked-file deploy trap — verify deploys by ancestry, not by tag or commit message
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Struck TWICE (Tier-2 checker missed two releases; check_tool_acceptance_due missed v1.0.1111); rules banked in HANDOFF T11 and the durable-rules list.
- **what:** `git commit -a` commits modified-tracked files only — an untracked (`??`) new file silently misses any number of release commits while its sibling changes ship. Guards: `git status` for `??` before every release (or commit new files as written); verify a deploy carries your files by ANCESTRY (`git merge-base --is-ancestor <commit> <release>`); this repo also reuses version tags, so pod-start-time vs commit-time settles what a tag actually contains, not the commit message. Safe-failure companion: unknown discovery-check names warn+skip (the 142 precedent), so wiring a check by migration before its binary deploys is safe.
- **sources:** HANDOFF_2026-07-10…md#T8,T11,#§4; RUNNING_NOTES_travelling_docs(39).md#stage-5-live,#v1.0.1111; README_summary_paragraph2_for_discussion.md
- **relations:** continuous sweep gate; migrations-before-binary safety.
- **verify-later:** n/a (operational pattern).

<!-- SOURCE: U08_travelling_docs.md -->
### Page build/rerender failure-shape thread family (Parts 1–5 + wrong-turns log)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016b open-threads status header: Part 1 DONE/verified; Part 2 partially verified; Part 3 code prepared not applied; Part 4 written un-deployed; Part 5 triaged.
- **what:** A connected series on "work that reports success but doesn't happen": result-contract drop replaced child output with a success stub (fixed, shipped 2026-06-18); no-LLM re-render pre-pass (partially verified); item_key canonicalization drift (needs_page vs needs_tool_recreation colliding on the dedup index — builder prepared); the interactive clobber (above); system-stats dropped because content_data and the component template share ZERO keys (a content↔template key-contract mismatch — the visible-content filter was correct). The companion "Wrong turns" log records false leads with the durable heuristic each violated — a deliberate documentation convention so the next pass doesn't re-walk them.
- **sources:** 016b_debugging_guide_7_3_(7).md#open-threads,#wrong-turns; (fix detail lives in the gamesdesign/scheme runbooks outside this unit)
- **relations:** silent-completion invariants below; travelling copies of 016b.
- **verify-later:** current state of Parts 2/3/5 fixes.

<!-- SOURCE: U08_travelling_docs.md -->
### Debugging durable invariants (trust the artefact; sampled steps; silent no-op config paths)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b "Durable invariants and heuristics" section, carried forward from 016 and extended through this arc.
- **what:** The distilled heuristics: trust the rendered artefact, not the status (work items/commits can report success on no-op work); completed_at is the orchestration end, not the write instant; a config key read on a different path than it's set is a silent no-op, not an error; only save_page_sections writes page_components; 0 rows is not decisive; a negative inference from an artefact's shape needs the mechanism checked in ALL cases; reuse before rebuild; check the schema before SQL. Plus 016b v4 additions: two page-assembly paths with different chrome sources (stale `site_components` renders fossilise; only a full page-build rebuild re-renders templates; provenance greps + legacy-variable tell); the needle-gate template-surgery pattern (LIKE booleans + occurrence counts + backup + guarded idempotent UPDATE + RETURNING + rollback); `sites.status` vocabulary (draft/building/review/published/deployed/archived/error — 'active' is legacy; nothing filters on it; never scope blast-radius by it).
- **sources:** 016b_debugging_guide_7_3_(7).md#durable-invariants,#light-site-dark-chrome,#sql-pitfalls,#sites-status
- **relations:** the whole debugging category; 016 back-catalogue (other unit).
- **verify-later:** n/a (heuristics; primary copies of 016/016b covered by their own unit).

<!-- SOURCE: U09_adoption.md -->
### Convergence inertness: []map[string]interface{} vs []interface{} type-assertion bug
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "A clean re-adoption proved `reconcilePlanWithRealised` has never run for any site… `query_database` returns []map[string]interface{} — a type that does NOT satisfy that assertion in Go… Fix… accepts both… plus an 'existing pages loaded for convergence' count log so an empty set can never fail silently again" (2026-06-05, verified fixed same day).
- **what:** The whole doc-029 Phase-1 convergence was dead since deploy because `ValidateSitePlanAction` asserted `ev.([]interface{})` on QueryDatabaseAction output, which is `[]map[string]interface{}` — the assertion always failed, existingPages stayed empty, and reconcile early-returned silently. A canonical instance of the "silent empty input" failure class; the fix pairs the type switch with a count log so emptiness is observable. Also documented: QueryDatabaseAction stringifies jsonb columns (sections arrive as JSON strings needing json.Unmarshal).
- **sources:** FOCUS_adoption_faithfulness_via_locks(5).md#2026-06-05-correction, running_notes_14(25)#part-14l
- **relations:** convergence; union-clobber fix (downstream of it); debugging-guide no-op pattern entry
- **verify-later:** v3_site_actions.go type switch; 016 debugging guide v2_31+ entry

<!-- SOURCE: U09_adoption.md -->
### Defect-catalogue discipline (families by root cause; read-pin-confirm-fix)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "Enumerate every observed defect as a separate item before fixing, so distinct causes are not conflated… Causes marked 'tentative' have NOT been pinned" (CATALOGUE header); practiced across Families A–J with per-item verification states.
- **what:** Post-run methodology: walk the deployed site, catalogue defects grouped by root cause (not symptom) into families (A deployment gaps, B silent-fallback links, C list-component content, D section-data gaps, E content quality, F guides duplication, G design fidelity, H hygiene, I unknowns, J dispatch throughput), then work each as its own thread — read the responsible action, pin the cause, confirm against data, only then fix. Paired with reading-discipline rules: site_plan_pages is the authoritative plan output; confirm run completion before diagnosing; teardown by site_id never domain; matching updated_at is not authorship; a hardcoded site_id is stale after any teardown (resolve via domain subquery); an empty LEFT JOIN means wrong anchor, not missing link.
- **sources:** CATALOGUE_gamesdesign_post_sync_fix_defects(9).md, HANDOFF_2026-05-25#reading-discipline, running_notes_14(25)#principles
- **relations:** 016 debugging guide checklist items 20–22; council/fix-loop methodology
- **verify-later:** n/a (methodology); debugging-guide entries

<!-- SOURCE: U09_adoption.md -->
### Kafka consumer-group recovery (restart-to-rejoin, never replay)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "Resolution: chassis restart re-established group membership; a fresh trigger produced orchestration… Kafka recovery: restart-to-rejoin + park at latest + one fresh trigger, rather than replay-from-earliest (which would spawn stale adoptions)" (running_notes_14 Part 4).
- **what:** After a topic wipe destroyed `__consumer_offsets`/membership, the chassis logged clean consumer setup but wasn't joined; triggers produced messages nobody consumed (site row created by trigger path, no orchestration row). Diagnostic: `kafka-consumer-groups --describe` empty = not consuming regardless of producer health. Recovery doctrine: restart to rejoin, park at latest; a `--reset-offsets --to-earliest` replay was a mistake that risked duplicate stale adoptions. Principle: a trigger printing IDs proves production, not consumption.
- **sources:** running_notes_14(25)#part-4
- **relations:** orchestration creation writes a state row at creation (absence = never consumed); scheduler tick races during DB cleanup are noise
- **verify-later:** n/a (operational doctrine)

<!-- SOURCE: U09_adoption.md -->
### Migration/prompt-edit gotcha conventions (replace() anchors, funcMap templates, backup snapshots)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "Gotchas worth keeping (hard-won during the build)" (FOCUS_directory_builder); the guide-page_type migration applies them (anchor pre-checks counting 1, quote-free/newline-free replaces, ::jsonb cast validation, in-txn snapshots).
- **what:** Conventions for safely editing prompts/configs stored as JSONB text: Postgres replace() is literal-byte and silently no-ops on missed anchors while reporting UPDATE 1 (verify anchors with COUNT, keep them short and unique; re-entrant replaces append on every run); CASE doesn't short-circuit sub-SELECTs (use DO blocks + RAISE); NAMEDATALEN 63 truncates backup-table names; prompt templates pass through Go text/template so literal {{…}} needs funcMap helpers (placeholder/rangeStart/rangeEnd); prompt self-check and the Go validator are two halves that must change together; every migration carries a restorable in-txn snapshot with documented rollback.
- **sources:** FOCUS_directory_builder_and_list_components.md#gotchas, migration_adoption_add_guide_page_type.sql, running_notes_14(25)#snapshot-standard
- **relations:** thin-slice constitution (snapshot rule); component-creator Tier-D prompt work
- **verify-later:** n/a (conventions)

<!-- SOURCE: U09_adoption.md -->
### Manual work-item insertion as an operational rebuild lever
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "Operational fact confirmed: a manually-inserted needs_page / needs_content_page work item IS claimed by build-dispatch-loop (status triaged → claimed → complete), so single-page (re)builds can be hand-triggered" (HANDOFF_2026-06-06).
- **what:** Canonical hand-trigger shapes: re-render existing components → `needs_page` (spec {reason,page_name}); generate content → `needs_content_page` (spec {mode:'recreate',source:'adoption',page_name,page_type}); both handler_agent='page-build-handler', status='triaged', ON CONFLICT DO NOTHING, ids resolved via domain subquery. Verified end-to-end unstick pattern for skinner-box, guides-index and the homepage.
- **sources:** HANDOFF_2026-06-06#key-references, GUIDE_deploy_from_context_packs.md#C, RUNBOOK(2)#4
- **relations:** dispatch pipeline; positive-evidence verification ("complete" alone is not proof)
- **verify-later:** n/a (operational recipe)

<!-- SOURCE: U10_imagery.md -->
### Kafka per-spawn response-topic partition race
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** "did not reproduce on second run… monitor" (2026-05-11); HANDOFF 2026-07-12 "Kafka per-job response-topic partition race — transient; now surfaces as failed items (mark_item_failed fix) instead of silent successes."
- **what:** An adapter (git-adapter observed; kafka-go LeastBytes balancer) occasionally writes its response to partition 1 of a single-partition per-spawn topic, losing the reply — work succeeds but the orchestration times out/fails. Root cause suspected stale partition metadata for just-created topics. Never structurally fixed; consequence downgraded from silent-success to visible failed items by the mark_item_failed pattern. The same race killed a content-writer reply and produced the "no-op complete" anomaly.
- **sources:** ANALYSIS_phase_2f_two_defects.md#Defect-2, RUNNING_NOTES_imagery_best_in_class.md#Turn-16, HANDOFF_imagery_best_in_class.md#Open-threads
- **relations:** mark_item_failed error-honesty fix; consumer-group race (separate doc, chassis replicas=1).
- **verify-later:** `platform/kafka/producer.go` balancer; adapter logs for "topic partition not found".

<!-- SOURCE: U10_imagery.md -->
### Work-item re-drive and zombie-claim operational semantics
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** Standing lessons recorded Turns 31–32 (2026-07-12/13); "the zombie-claim dispatch stall was the single biggest time cost of the 2026-07-09/10 verification" (B9, still open).
- **what:** Hard-won dispatch mechanics: a claimed item stuck >~10 min blocks its ENTIRE site via find_dispatchable_site's NOT-EXISTS clause (standing unstick UPDATE; real fix = reaper cadence + per-item-type circuit breaker, TODO 6/10/11); re-driving an item requires resetting `attempt_count=0` and claim metadata, not just status (capped items are silently excluded — dispatch looks dead but is correctly idle); a just-finished orchestration's tail can re-stamp a freshly-reset item complete (state-machine race); manually-inserted items are NOT auto-triaged (insert as triaged); dedup is a partial unique (site_id, item_key) over non-terminal statuses whose exact semantics made resets awkward. Historical: dispatch once didn't claim triaged imagery items behind page work; fairness/observability gaps (outer ORDER BY, trigger not writing orchestration_states) remain listed.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-32, HANDOFF_imagery_best_in_class.md#Mechanisms, TODO_imagery_followups.md#6/#8/#9/#10, RUNBOOK_imagery_best_in_class.md#B9
- **relations:** mark_item_failed; state-machine corruption on failed items (claim metadata not cleared); scheduler-and-tasks.
- **verify-later:** find_dispatchable_site SQL; reaper cadence; idx_swi_dedup definition.

<!-- SOURCE: U10_imagery.md -->
### Pipeline field as soft routing label
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "check_unfulfilled_imagery_plan.go hardcodes Pipeline: 'build' — the 2026-05-17 fix is in the code" (verified 2026-07-08); Part B dispatcher-filter loosening scoped alongside.
- **what:** Discovery checks running under design-discovery-agent inherited pipeline='design', which build-dispatch-loop's item_pipeline filter silently excluded — needs_imagery items required manual UPDATEs to dispatch. Two-part fix: checks write Pipeline:"build" at source (pipeline is the destination handler's side, not the origin's), and the dispatcher's filter was removed so any future mismatched emission still dispatches. The field survives as a soft routing label for possible future multi-pipeline dispatchers.
- **sources:** TODO_imagery_followups.md#7, RUNNING_NOTES_imagery_best_in_class.md#Turn-2 (verification)
- **relations:** work-item dispatch semantics; design-discovery-agent context.
- **verify-later:** build-dispatch-loop load_items config; Pipeline literal in imagery checks.

<!-- SOURCE: U11_traffic_probe.md -->
### Traffic-probe field lessons absorbed into the debug guide (#24–#28)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-12 "Debug guide updated … 016_debugging_guide_v2_46" and 2026-06-13(g) "Debug guide v2_48".
- **what:** Five checklist entries earned in this project's field work, each rule + dated instance: #24 a config/workflow file is only authoritative at its runtime read-path (the stale agentchassis/.git/workflows/deploy-to-b2.yml nearly produced a never-firing Action); #25 prove the harness delivered the intended input before debugging the system (dash not expanding $'…' made the field literally "$value"); #26 shell variables never reach child processes without export, die with the session, and error-text-vs-source mismatch means a stale deployed artifact — read state back from the artifact, not `echo $KEY`; #27 never invent an interface — compiling standalone ≠ satisfying the real DiscoveryCheck signature; #28 agent_definitions is UNIQUE(type,version) with two similar category columns. Plus operator-handover lessons: explicit file manifests + a loud go vet/build check, flat-shipped workflows (delivery channel rejects dot-dirs), git branch -M main before first push.
- **sources:** traffic_probe_running_notes(28).md#2026-06-12 (debug guide v2_46, operator execution) + #2026-06-13-g (v2_48), traffic_probe_runbook(13).md#3.5-3.6 (traps in place)
- **relations:** debugging (016 guide family), per-domain notes convention
- **verify-later:** 016_debugging_guide latest version contains #24–#28

<!-- SOURCE: U12_docs024_archives.md -->
### Debugging playbook (early runbook)
- **category:** debugging
- **status-signal:** superseded
- **status-evidence:** Both archive files are early drafts of the operational runbook; the current authoritative version is `016_debugging_guide_v2_58_consolidated.md`.
- **what:** A ten-section operational runbook: pod health check, work-item status queries, scheduled-task flight-status, orchestration-state staleness, agent error log, handler-agent-definition existence checks, timeout ordering chain, a failed-item cleanup transaction, named failure patterns, and a single "quick health dashboard" query. The second draft adds a systematic dispatch-loop `input_mapping` path-mismatch diagnosis, missing-handler-agent detection, and a log-hunting technique.
- **sources:** old/older1/016_debugging_guide.md; old/older1/016_debugging_guide_v2_april26.md
- **relations:** timeout chain ordering; dispatch-loop input_mapping mismatch; wont_fix/needs_section_data patterns
- **verify-later:** whether the consolidated live debugging guide still carries these same queries/patterns.

<!-- SOURCE: U12_docs024_archives.md -->
### Timeout chain ordering contract
- **category:** debugging
- **status-signal:** unknown
- **status-evidence:** Stated as a hard ordering requirement in both drafts (claim_timeout > call_handler timeout > workflow timeout), with the call_handler timeout bumped from 900s to 1200s between drafts; not verified against the current consolidated guide.
- **what:** Three timeouts must nest correctly or two failure modes occur: reset-claim double-handling, or dispatch marking an item failed while the handler is still working with nothing listening for its response.
- **sources:** old/older1/016_debugging_guide.md#"7. Timeout Chain"; old/older1/016_debugging_guide_v2_april26.md#"7. Timeout Chain"
- **relations:** debugging playbook
- **verify-later:** current values of `claimed-item-timeout`, `build-dispatch-loop` call_handler timeout, per-handler workflow timeouts.

<!-- SOURCE: U12_docs024_archives.md -->
### Early pipeline-failure triage priorities dropped by root-cause diagnosis
- **category:** debugging
- **status-signal:** abandoned
- **status-evidence:** The 2026-04-14 report's P3 (vonc.com raw CSS), P4 (stale-item process gap), P5 (timeout tuning) don't appear in the 2026-04-15 v3 report's P1-P10 list at all.
- **what:** First-pass triage of 57 stuck work items framed three priorities at the symptom level. Within a day, deeper diagnosis replaced these with concretely-fixed root causes not originally identified: rate-limit errors misclassified as non-transient (1,869 occurrences), `load_page_record` lacking a `page_id` fallback, and later audit-finding routing/classification bugs.
- **sources:** old/older1/105_dispatch-pipeline-failures-report.md#"Priority Fixes"; old/older1/105_dispatch-pipeline-failures-report_v3.md#"Priority Fixes"
- **relations:** plan_sections pre-check evolution; three-way audit-finding classification
- **verify-later:** current state of vonc.com's about page (raw-CSS-serving bug).

<!-- SOURCE: U12_docs024_archives.md -->
### CrashLoop `exec: "./X"` image/binary-content mismatch diagnosis
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Live v2_58 header: "Recovered two §9 entries that were present in the earliest file v2_47(1) but had been dropped from the v2_48-onward branch."
- **what:** A three-command image-inspection technique (`docker run --entrypoint ls`, `docker inspect .Config.Entrypoint`, `.RepoDigests`) for diagnosing `CrashLoopBackOff` with `exec: "./X": no such file or directory` — proves the running image lacks the named binary (wrong build context / tag-sharing), not a config problem.
- **sources:** archive_april_26/016_debugging_guide_v2_47(1).md#"§9 CrashLoop exec"; docs024_key_docs_latest/016_debugging_guide_v2_58_consolidated.md
- **relations:** temporarily abandoned in the v2_48→v2_57 main branch (this fork diverged at v2_45), recovered wholesale into live v2_58
- **verify-later:** whether a CI guard ("fail build if binary absent") was ever implemented for thunder-adapter/analyser-adapter Dockerfiles.

<!-- SOURCE: U12_docs024_archives.md -->
### Hand-applied agent/launcher-def migrations are not commutative
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** The other v2_47(1)-only recovered entry; incident resolved by re-applying 110 then 111, RUNBOOK "2d state check" added as a live procedural safeguard.
- **what:** Re-applying migration 109 (per a runbook's "safe to re-run" claim) silently reverted later migrations 110/111 because 109 rebuilt DB-object nodes that 110 had replaced. A migration is idempotent only against its own prior application, never against later migrations touching the same path.
- **sources:** archive_april_26/016_debugging_guide_v2_47(1).md#"§9 Re-running an idempotent migration"; docs024_key_docs_latest/016_debugging_guide_v2_58_consolidated.md
- **relations:** NEW:migration-governance proposal (below)
- **verify-later:** confirm the `training-launcher` agent_definitions row currently reflects migrations 109-111 in correct order.

<!-- SOURCE: U12_docs024_archives.md -->
### gamesdesign `index` silent-staleness investigation — superseded hypothesis chain
- **category:** debugging
- **status-signal:** superseded
- **status-evidence:** Three successive, explicitly-labelled hypotheses in the same changelog: "silent-completion from a pod dying mid-flight" → "NOT a timeout/deploy issue at all... content-regression guard errors masked as success" → "SUPERSEDED-PENDING-CONFIRMATION" opening a metadata-path-mismatch thread.
- **what:** A multi-week live diagnosis of why gamesdesign.co.uk's `index` page stayed stale despite repeatedly "completing" rebuilds. Each hypothesis explicitly superseded the previous as new evidence arrived. Eventually-confirmed root cause is a more general mechanism — "Child workflow result silently replaced by a stub" (`output_field` vs `output_fields`), shipped 2026-06-18.
- **sources:** archive_april_26/016_debugging_guide_v2_49.md, v2_49(1).md, v2_49(2).md#"§9"; docs024_key_docs_latest/016_debugging_guide_v2_58_consolidated.md
- **relations:** own recursive application of "don't trust a complete status" heuristic
- **verify-later:** confirm `platform/orchestration/result_spec.go` (`resolveResultSpec` fix) is present in the current codebase.

<!-- SOURCE: U12_docs024_archives.md -->
### Pod label key is `agent-type` (hyphen) vs log field `agent_type` (underscore)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Stated as a settled operational rule with a named failure mode already observed.
- **what:** Kubernetes pod labels use `agent-type` while structured log JSON fields use `agent_type`; using the underscore form in a `kubectl logs -l` selector silently matches zero pods. Separately, a correct selector spans ALL live pods of that type, so a tail can mix in a previous run's failure dump.
- **sources:** archive_april_26/016b_debugging_guide_7(4).md#"Pod label key is agent-type..."
- **relations:** older trigger scripts (082/083c) still carry the underscore form; absent from canonical live 016b
- **verify-later:** grep trigger scripts 082/083c for the underscore `agent_type=` selector.

<!-- SOURCE: U12_docs024_archives.md -->
### Two failure envelopes — a COMPLETED parent orchestration does not mean the child succeeded
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Presented as a settled mechanism with two named, confirmed code paths (`sendWorkflowResponse` vs `notifyParentOfFailure`).
- **what:** A mid-run step failure is reported via `sendWorkflowResponse` with header `status:"complete"` but the real failure in the body, which the parent forwards and then itself shows COMPLETED with a non-empty `error` column; a START-time failure instead uses `notifyParentOfFailure` with `status:"error_unrecoverable"`. Consumers must check the body, never the header status alone.
- **sources:** archive_april_26/016b_debugging_guide_7(4).md#"Two failure envelopes"
- **relations:** same "trust the artefact, not the status" family as the guide's core silent-completion heuristics; absent from canonical live 016b
- **verify-later:** read the current `sendWorkflowResponse`/`notifyParentOfFailure` implementations to confirm the two-envelope shape.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Assumed-status-values trap (debugging lesson)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Formalized as Section 9 addendum + Section 0 checklist candidate in 016_debugging_guide_addenda.md
- **what:** General lesson: never assume status-column values from naming conventions — always run `SELECT DISTINCT status FROM <table>` first. `pages.status` uses `'active'` exclusively platform-wide; other plausible values simply don't exist.
- **sources:** js_snippets_news_gaswholesalers/016_debugging_guide_addenda.md#Assumed-status-values-trap, js_snippets_news_gaswholesalers/old/design_actions_status_filter_fix.md
- **relations:** CSS component-list fallback bug
- **verify-later:** grep/inspect `SELECT DISTINCT status FROM <table>`; `pages.status`; `'active'`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### "Renders empty" diagnostic method (data-binding, not template, diagnosis)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Formalized into 016_debugging_guide_addenda.md Section 9 headline entry + Section 0 checklist item #16
- **what:** A general, reusable 5-step diagnostic method for "a component renders its structural shell but no repeated content": (1) check page_components for orphaning; (2) confirm input_schema expectations; (3) check whether structured data exists anywhere; (4) count rendered shells; (5) compare actual sections against site_plan for duplicate/stale slots. Core lesson: empty shells mean the template ran — the bug is in data binding, never trigger a rebuild before completing this walk.
- **sources:** js_snippets_news_gaswholesalers/016_debugging_guide_addenda.md#empty-shells, js_snippets_news_gaswholesalers/old/empty_faqs.md
- **relations:** FAQ duplicate content-surface bug; rendered_html snapshot-not-view pattern; isolated build test methodology
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### rendered_html as snapshot-not-view (stale render after content_components migration)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Verified via a diagnostic query comparing template_has_script_src vs rendered_has_script_src on live gaswholesalers data
- **what:** A migration to `content_components.html_template` does not retroactively affect already-built pages, because `page-rerender` uses `page_components.rendered_html` — frozen output from the last writer run — and never re-pulls from the live template. General principle: `rendered_html` is a snapshot, not a live view; migrations touching `content_components` must also update affected pages' snapshots or trigger a rebuild.
- **sources:** js_snippets_news_gaswholesalers/016_debugging_guide_addenda.md#Migration-updates, js_snippets_news_gaswholesalers/old/findings_and_plan_news_visual.md
- **relations:** files_field deploy bug; "Renders empty" diagnostic method
- **verify-later:** grep/inspect `content_components.html_template`; `page-rerender`; `page_components.rendered_html`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Isolated build test methodology (throwaway test-page pattern for pipeline diagnosis)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** used successfully to prove the content writer was not the bug
- **what:** A reusable diagnostic technique: create a throwaway page (kept out of nav) with a deliberately minimal/isolated sections list, drive it through the full production build path, then read out `page_components` to conclusively attribute a bug to a specific pipeline layer. Used to prove the FAQ writer works correctly in isolation.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md#The-page-content-creation-flow, js_snippets_news_gaswholesalers/old/page_content_creation_flow.md
- **relations:** FAQ duplicate content-surface bug; "Renders empty" diagnostic method; page content-creation build pipeline trace
- **verify-later:** grep/inspect `page_components`

<!-- SOURCE: U14_docs019_runbooks.md -->
### Orchestrator COMPLETED while child FAILED (body.status check)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B5 "orchestrator can show COMPLETED while the child FAILED — header status complete, body.status failed; consumers of child results must check body.status (behaviour, recorded)".
- **what:** Recorded platform behaviour: a parent orchestration's header status can read complete while the child's embedded body carries failed; any consumer of child results must check body.status, not the header. Adopted cross-thread from the tools chat's notes.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B5 (useful from their notes)
- **relations:** oversize delivery (child fails at complete); stage-by-stage verification
- **verify-later:** response-building code paths; parent/child rows of a failed diagnose run

<!-- SOURCE: U14_docs019_runbooks.md -->
### error_step-inside-config gotcha and pod-reap evidence substitute
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** diagnosis_fix_loop(9) "New gotcha ADOPTED (their 001 §16 finding): error_step belongs INSIDE a step's config — step-LEVEL error_step is silently ignored (dormant bug instances exist in tool agents) … idle pods reap at ~3600s — the post-completion STATE DUMP (ProcessingHistory) is the accepted evidence substitute."
- **what:** Two operational facts: workflow error routing only works when `error_step` sits inside the step's `config` object (top-level placement is silently ignored — dormant instances exist and should be corrected when touching a workflow, as its own noted change); and spawned agent pods are reaped ~3600s after idle, so post-mortem evidence comes from the orchestration state's ProcessingHistory dump, not pod logs.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#collision-surface
- **relations:** stage-by-stage verification; standing evidence rules
- **verify-later:** error_step placement across tool agent workflows; agent-job-cleanup timing

<!-- SOURCE: U14_docs019_runbooks.md -->
### Stage-by-stage rebuild verification and the false-complete rule
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** gamesdesign_index_rebuild §5 stages A–E with per-stage SQL; "status='complete' is only meaningful together with Stage C showing changed components; complete + unchanged components = the old false-complete".
- **what:** The verification method for a page rebuild: A writer delivered a flat result to the parent (sections_metadata path check) → B save attempted/blocked loudly (agent_error_log) → C components actually changed (content_hash/updated_at fingerprint vs baseline) → D work item completed on a REAL save (complete only meaningful with changed components) → E deploy. Baseline-first (capture fingerprints before triggering), re-open the existing work item rather than fabricating one, and the triage table maps each stopping stage to its likely cause.
- **sources:** docs019/RUNBOOK_gamesdesign_index_rebuild.md#2; docs019/RUNBOOK_gamesdesign_index_rebuild.md#5; docs019/RUNBOOK_gamesdesign_index_rebuild.md#7
- **relations:** oversize delivery (fix #3); content-regression guard; standing evidence rules
- **verify-later:** page_components fingerprint queries; site_work_items re-open pattern

<!-- SOURCE: U15_docs019_running_notes.md -->
### Gamesdesign silent-no-op-rebuild bug (content-regression + status-rollup)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** v2(36) STATE DIGEST: "gamesdesign silent-no-op bug — RESOLVED (the real fix is now a fixture)."
- **what:** A real production bug used repeatedly as the diagnosis loop's eval fixture, diagnosed across several sessions with TWO wrong hypotheses along the way (per-section `max_tokens:2000` cap; then recreate-mode discriminator) before the real cause was found: a January chassis regression made `SagaCoordinator.extractWorkflowResult` honour only the PLURAL `output_fields` key, while `page-content-writer` declares the SINGULAR `output_field`, so the compiled page collapsed into an oversized state-dump skip path that reported "completed" while the live page never updated. Fix: `resolveResultSpec` (new, `result_spec.go`) treats singular as FLATTEN, honouring the long-ignored mapping key. The reversals in this diagnosis are the canonical worked example baked into the diagnosis loop's verdict prompt.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-14/17 gamesdesign entries; NOTES_running_synthesis_principles(59) 2026-06-13/14 diagnosis narrative.
- **relations:** SagaCoordinator output_field contract; diagnosis loop; B4a embedding-quality finding (this bug's real fix is the "ceiling" ground-truth task).

<!-- SOURCE: U15_docs019_running_notes.md -->
### Code-retrieval corpus staleness (§7 route)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** v4(39) headers 2026-07-02: "corpus check result: the index is the blocker... the index is of a YEAR-OLD tree."
- **what:** After the diagnosis loop was measured to gain nothing from code retrieval (see B4a finding), a follow-up investigation (§7 route) found the underlying `code_symbols` index itself was built from a year-old stale checkout of the default branch (main stale since 2025-07-14) — a corpus problem, not a retrieval-quality problem — leading to a reindexing effort, ref-pinning strategy, and ultimately the decision to migrate the code-indexer's analysis step onto the already-proven `analyse_repo_local` path.
- **sources:** NOTES_running_synthesis_v4(39).md headers 2026-07-02/03; DECISIONS section.
- **relations:** B4a embedding-quality finding; code-context retrieval infrastructure.
- **verify-later:** Current freshness of the deployed `code_symbols` index; whether the analyse-step migration was applied.

<!-- SOURCE: U16_docs019_design_plans.md -->
### Spawn-consumed columns lesson (seeds copy image columns from a live donor)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NNN_fix_researcher_spawn_columns documents the incident and fix; HANDOFF_builder_thread carries it as a standing guard ("seeds must copy image columns from a live donor (the amended seed does)").
- **what:** getAgentDefinition SELECTs image_repository/image_tag/command/resources/health_config/env_vars/idle_timeout_seconds and gates on is_active=true; a seed populating only default_config leaves command NULL → the image's default entrypoint boots the GENERIC chassis service, which never reads the injected AGENT_TYPE env, so the dispatcher's call goes unheard and the item stays claimed. Fix and rule: copy the spawn-consumed infrastructure columns from a proven donor (deliberately NOT capabilities/topics/default_config). Related: image_tag DEFAULT 'latest' pointed at an ancient build; the makefile now pins IMAGE_TAG. Sibling gotchas carried with it: pod label key is agent-type (hyphen); check body.status not just the header; error_step belongs INSIDE step config (step-level silently ignored); idle pods reap ~3600s with ProcessingHistory dumps as post-reap evidence.
- **sources:** NNN_fix_researcher_spawn_columns.sql; HANDOFF_builder_thread.md#2,#5; HANDOFF_fixloop_thread(8).md#3
- **relations:** workflow-in-default_config lesson; index-orchestrator spawn wrapper
- **verify-later:** guidelines 001 New Agent checklist line (flagged residual)

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Debugging guide & assumption-checklist methodology
- **category:** debugging
- **status-signal:** superseded
- **status-evidence:** 016 v2_44 §0 "Most defects in recent sessions … came from acting on unverified assumptions"; archive copy, live successor in docs024
- **what:** The canonical symptom→cause→fix guide, fronted by a 23-item assumption checklist. Covers pod health, work-item/orchestration/scheduled-task/error-log queries, timeout chain, and ~50 specific failure patterns.
- **sources:** WM/016_debugging_guide_v2_44.md#0, WM/016_debugging_guide_v2_44.md#9, WM/016_debugging_guide_v2_44.md#7
- **relations:** superseded by docs024 live 016; architectural tensions; agent = row in agent_definitions
- **verify-later:** orchestration_states.error_preview; agent_error_log; llm_call_log

<!-- SOURCE: U18_sql_for_agents.md -->
### rag_index chunkContent OOM saga (bypass → reenable → rebypass → fix)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Four-migration arc 135/139/140/141 (2026-07-09/10) with root cause CONFIRMED in 140: "chunkContent() never terminated on content longer than chunk_size... ~2Gi of duplicate chunks in seconds. Both chassis OOMKills were this loop."
- **what:** A model incident record: tool creation hung/OOMed at index_plan. First hypothesis (no embedding deadline) produced 135's bypass + a hygiene deadline (139); reoccurrence disproved it; the real bug was a non-terminating chunk loop (start = end - overlap re-entering forever), fixed in Go with regression tests, then re-enabled by 141. Durable practices demonstrated: reversible SQL bypasses that keep truth in Postgres (write_plan) while sacrificing only derived indexing; explicit preconditions in re-enable migrations; superseding one's own root-cause statements on record.
- **sources:** 135_bypass_index_plan_until_embed_timeout.sql; 139_reenable_index_plan.sql; 140_rebypass_index_plan_chunk_loop.sql; 141_reenable_index_plan_after_chunk_fix.sql
- **relations:** rag knowledge base; travelling docs pipeline notes; 016b debugging lessons
- **verify-later:** rag_actions_chunk_test.go presence; deployed image ≥ fix commit

<!-- SOURCE: U18_sql_for_agents.md -->
### error_step-inside-config routing rule
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 128 "Durable rules this incident banked (016b): error_step lives INSIDE step.Config and must name an EXISTING step; derive convergence targets from the step's own next_step, never guess"; effect verified live 2026-07-10; 131/132 retro-move ten inert step-level error_steps into config.
- **what:** Chassis workflow convention discovered through failures: the coordinator reads step.Config["error_step"] only — step-LEVEL error_step keys are silently ignored; a routing target that names a non-existent step fails the whole workflow. Correct-while-touching policy migrates old inert keys whenever a workflow is edited.
- **sources:** 128_fix_load_runtime_error_step_target.sql; 127_diagnose_load_runtime_error_step.sql; 131_tool_generator_plan_writing.sql; 132_fix_agents_note_writing.sql
- **relations:** 016b debugging heuristics; template field-path rule (134)
- **verify-later:** coordinator error-routing code

<!-- SOURCE: U18_sql_for_agents.md -->
### Prompt-template field-path rule (text vs json output shapes)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 134 (2026-07-09) "THE RULE (proven by this run, not assumed)": text-format steps pass the bare string to downstream templates ({{.generated_html}}, not .result); json-format steps pass a map (use `| toJSON`); action-config field paths are a DIFFERENT resolver and keep .result.
- **what:** A durable rendering contract distinguishing three resolvers: Go template rendering of LLM text results (bare string), of JSON results (map, dump with toJSON rather than guessing keys), and action-config field paths (keep .result suffix). Applied as one blocker fix plus three pre-emptive corrections of the same bug class.
- **sources:** 134_fix_prompt_template_field_paths.sql
- **relations:** call metadata/response convention; error containment via config.error_step (docs steps can never fail tool creation, 131)
- **verify-later:** ExtractActionInputs / template renderer code

<!-- SOURCE: U19_sql_tables_components.md -->
### orchestration_state_audit investigation trigger
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Create-trigger + analysis queries (time_since_prev via LAG, pg_backend_pid, application_name) and explicit "Remove trigger when done investigating" teardown.
- **what:** A temporary, attachable audit table + AFTER UPDATE trigger capturing every version/status/current_step transition on orchestration_states — used to diagnose state races and stuck orchestrations, then removed. Distinct from permanent logs; also cleaned up by database-cleanup (keeps last 100k rows).
- **sources:** docs/agent_docs/sql_for_tables/010_orchestration_state_audit.sql; docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#database-cleanup
- **relations:** debugging guide; database cleanup.
- **verify-later:** whether trigger currently attached.

<!-- SOURCE: U19_sql_tables_components.md -->
### agent_error_log persistent error record
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "This replaces digging through kubectl logs to find error details"; captured from routeToErrorStep and notifyParentOfFailure; referenced later as the sink for Tier-D validator rejections.
- **what:** Queryable record of every agent error: what failed (site/domain/work_item), where (orchestration, agent_type/id, pod, step, action), the error (message, error_code, severity), a JSONB context snapshot, and resolution tracking (resolved/resolved_by). Indexed for dashboard recency, per-site, unresolved, and per-agent-type frequency views.
- **sources:** docs/agent_docs/sql_for_tables/022_agent_error_log.sql; docs/agent_docs/sql_for_tables/005_content_components.sql#migration-042
- **relations:** database cleanup retention; fix loops consuming structured errors.
- **verify-later:** writers in chassis error paths.

<!-- SOURCE: U19_sql_tables_components.md -->
### http_request_log outbound HTTP observability
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Migration "Follows the same pattern as llm_call_log"; stats view with calls_last_5min for rate-limit monitoring; cleanup 90d success / 180d errors.
- **what:** Centralised log of every outbound HTTP call from Go actions: caller identity (agent/step/orchestration/action_name), method/url/domain/path, response status/bytes/latency/success, metadata JSONB. Purposes: operational visibility and per-domain rate-limit tracking (e.g. Companies House).
- **sources:** docs/agent_docs/sql_for_tables/026_http_request_log.sql
- **relations:** llm_call_log (pattern sibling); companies-house rate limiting.
- **verify-later:** HTTP client wrapper writing rows.

<!-- SOURCE: U19_sql_tables_components.md -->
### Claimed-item timeout with evidence-based auto-completion
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** v1 then v2 ("SUPERSEDES... Apply THIS one") migrations with two confirmed production false positives dated 2026-05-12 and 2026-06-04 (gamesdesign homepage auto-completed with ZERO page_components — root cause of the missing root index.html).
- **what:** The stuck-claim recovery task distinguishes "work actually finished but the response was lost" from "handler died": items claimed >15 min are auto-completed only on artifact-specific evidence — needs_content_page requires page_components rows for that page updated after the claim (ground truth, not the untrustworthy build_status='deployed' flag), page_rerender requires page.deployed_at after claim, needs_design keeps a caveated site-level check; needs_rerender is deliberately excluded (site-level, retry is cheap). Everything else resets at >40 min with attempt accounting and fail-on-exhaustion.
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql#migration_claimed_item_timeout_evidence_v2; docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#claimed-item-timeout
- **relations:** work queue lifecycle; build_status CHECK (flag trust); UpdatePageStatusAction 0-component guard ("Option B").
- **verify-later:** live pre_query text of claimed-item-timeout; debugging guide section 9.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Message-flow logging / observability plan
- **category:** debugging
- **status-signal:** aspirational
- **status-evidence:** README.002 Week-2 objective: "MessageFlowLogger… Track every message through the system with database persistence"; docs002/0100 problem statement repeats the desire ("closely log and track the creation of agents, the messages…").
- **what:** Persist every send/receive event, agent creation, and topic routing decision to the DB for replay/debugging. Only zap logging plus orchestration_states processing_history is evidenced; a dedicated message-flow store never appears.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md; docs002_hitl_parallel/README.0100.state_of_play_for_creating_website.md
- **relations:** debugging category (docs 016 successors); processed_messages table (exists — see reset runbook).
- **verify-later:** processed_messages table purpose; any message audit table.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Orchestration environment reset runbook
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** The identical script repeated across ≥5 docs: scale agent-chassis to 0, TRUNCATE processed_messages/orchestration_states/pending_requests, delete spawned jobs, delete all job.* topics, delete bootstrap topics, reset all consumer-group offsets to earliest, scale back up.
- **what:** The standard clean-slate procedure for the early platform's test cycles — also documents the persistence surface of the era: processed_messages (dedupe), orchestration_states, pending_requests tables; job.* + system.agent.* topics; spawned-by=orchestrator job labels.
- **sources:** docs001_flow_general/README.095d.mycurrentinputmessagebeforechanging.md; docs001_flow_general/README.096d.robotics_startmessage.md; docs004_website_capture_project/initial_messages/initial_messages.txt
- **relations:** debugging (docs 016 successors); stateless-agents concept (what gets truncated).
- **verify-later:** pending_requests/processed_messages tables still present?

<!-- SOURCE: U20_legacy_docs_a.md -->
### Early message-routing failure modes (case-study catalogue)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Each bug has a trace + fix doc: nested-vs-flat input_data mismatch (flow2), verbose child responses breaking aggregation (flow2), silent root completion (flow2/3), duplicate second response to own topic = "poisoned pill" crash-loop (flow5), responses_topic dropped in header parsing → __initial_responses_topic__ empty (4.2), missing in_response_to_request_id (081.b), fire-and-forget spawn ignoring init responses (flow12).
- **what:** The canon of failure modes that shaped the architecture: every major convention (data normalisation, reply-to storage, perspective transformation, single completion path, await semantics) exists as the fix to one of these traced production bugs. Valuable as diagnostic priors for any council debugging agent.
- **sources:** docs001_flow_general/README.011.flow2.md; docs001_flow_general/README.016.flow5.md; docs001_flow_general/README.4.2.lifespanofresponsestopic.md; docs001_flow_general/README.023.flow12.await_response.md; docs001_flow_general/README.012.flow3.md
- **relations:** all system-architecture concepts above; debugging heuristics (docs 016b successor).
- **verify-later:** none — historical lessons.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Parent-timeout vs child-HITL race
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** docs014/001 log trace: "pageflow-builder times out (5 min)... content-reviewer: Cleaned up expired awaited requests count=1. The fix is to increase the parent's timeout."
- **what:** A failure class where a parent's call_agent timeout fires before the child's HITL request can be answered; the parent retries with null body and the child's awaited request is cleaned up as expired, losing the pause. Fix: parent timeouts must exceed child HITL timeout windows.
- **sources:** docs014_research_agent/001_human_in_the_loop_response_flow.md#Why-There-Were-No-Awaited-Requests
- **relations:** stale orchestration sweeper; HITL protocol; timeout heuristics in debugging docs.
- **verify-later:** current call_agent timeout_seconds vs HITL timeout defaults.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Orchestration debug log taxonomy
- **category:** debugging
- **status-signal:** superseded
- **status-evidence:** docs006/010 raw notes listing grep targets ("DEBUGaa: What have I done with CollectedData", "The Golden Search: grep -B 5 -A 30 generate_html") plus a real database lock incident (idle-in-transaction blocking INSERT INTO sites).
- **what:** The early debugging playbook: canonical log messages for action execution flow, LLM calls, data extraction and CollectedData tracking, with kubectl grep recipes; plus pg_stat_activity lock triage and pg_terminate_backend for idle-in-transaction blockers. Ancestor of the formal debugging guides.
- **sources:** docs006_workflow_builder/010_debugging.md
- **relations:** debugging category docs 016/016b; data-path problem.
- **verify-later:** whether DEBUGaa markers remain in code.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Legacy un-extracted Mode-B shells (js-not-extracted class)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** Root-caused 2026-06-29 ("stored through a path that did NOT run separateInlineJS — most likely they predate its addition"); cosmetic-script extraction for provocation-card/lobby-grid still on the backlog 2026-07-09.
- **what:** provocation-card, lobby-grid (and brief-explanation) were stored via a pre-separateInlineJS path: raw inline script still in html_template, empty js_content, empty schema, `<no value>` placeholders — so `/tools/assets/{fn}.js` was never produced and their built-in interactivity never deployed. provocation-card's stored script was additionally truncated at generation (no `</script>`), which once shipped and swallowed the page footer. One creation-era bug with several surface symptoms (`js-not-extracted`, `mode-b-template`, section drops). Fix direction: regenerate through the current store path.
- **sources:** docs/RUNBOOK_phase2_provocation_js(29).md#extraction-bug-findings; docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~19:30 + #2026-07-02-~19:35; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-07-gate-passed (side-evidence)
- **relations:** Mode A/Mode B taxonomy; store-path validation hardening; separateInlineJS
- **verify-later:** content_components js_content/html_template for provocation-card 6163ff14 and lobby-grid 9304f14d (still raw inline?)

<!-- SOURCE: U23_docs_root_vonc.md -->
### Mode A / Mode B broken-template taxonomy + repair/regeneration routing
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_vonc_session "Structural findings to carry forward"; code delivered 2026-06-22/23 (checkBrokenTemplateSlots, repair_template_slots); gauntlet-interface Mode-A repaired, archetype-result-card Mode-B regenerated to q100.
- **what:** Two distinct broken-template failure modes in the component library. Mode A: `<no value>FIELD</no>` — a render output stored as source with field names surviving as fallback text; repairable by string substitution (`repair_template_slots`). Mode B: bare `<no value>` — template rendered against an empty context and the cleaned output stored back; field names irretrievably lost; requires `needs_component_regeneration` → component-creator. `repair_template_slots` detects Mode B (no `</no>` tags) and returns needs_regeneration instead of attempting repair; `checkBrokenTemplateSlots` discovery check surfaces both.
- **sources:** docs/RUNBOOK_vonc_session(1).md#structural-findings; docs/RUNNING_NOTES_vonc(36).md#two-broken-template-failure-modes; docs/RUNBOOK_vonc_migrations(14).md#step-1
- **relations:** legacy un-extracted shells; store-path validation (rejects `<no value>` at the gate); component regeneration in place
- **verify-later:** check_component_standards.go; fix_component_template_action.go repairNoValueSlots

<!-- SOURCE: U23_docs_root_vonc.md -->
### Trust-the-artifact debugging doctrine (silent-success family + verification discipline)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b "Durable invariants" section; proven repeatedly (ten complete-with-nothing items; "after ten silent no-ops... NOT trusting 'complete' without artifacts").
- **what:** The unit's core debugging doctrine: a `complete` work item or green commit proves nothing — verify by artifact (DB row, curl, browser); completed_at is orchestration END, not the write instant (trace child orchestrations); a config key read on a different path than it is set is a silent no-op (compare producer output to consumer read by exact path); 0 rows is not decisive until the query is cleared (wrong column/id/schema/window); a negative inference from an artifact's shape needs the mechanism checked in all cases (the separateInlineJS attribute-skip example); pod logs are ephemeral across rollouts (grep zap by message + JSON field, never 'field=value'; agent_error_log outlives pods); copy full UUIDs, never hand-type; ±6-byte js_len paste drift is cosmetic — bundle and browser are ground truth; dated backup tables per change (never reuse an IF-NOT-EXISTS backup name); only save_page_sections writes page_components.
- **sources:** docs/016b_debugging_guide_merged(3).md#durable-invariants; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§8; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-03-section-drop-closed
- **relations:** complete_error family; zap/pod-log entry; SQL surgery pattern
- **verify-later:** n/a (doctrine); stage 2 can test individual heuristics against code

<!-- SOURCE: U23_docs_root_vonc.md -->
### system-stats key-contract mismatch (content_data ↔ template key sets)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016b Part 5 "TRIAGED, 2026-06-24"; remedy un-applied at that entry ("full content rebuild... then re-check"); the component itself later regenerated to q100 in the vonc arc.
- **what:** A populated-but-blank section is a content↔template KEY-CONTRACT problem, not a generation failure: system-stats' stored content_data keys (eyebrow/heading/stat_1_number...) shared ZERO keys with its template placeholders (eyebrow_label/section_headline/stat1_value...) after component-creator rewrote the component mid-flight, so every placeholder rendered empty and the (correct) visible-content filter dropped the band fleet-wide (usage_count 22). Durable heuristic: diff the two key sets directly; and a component schema change should trigger dependent rebuilds.
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 5) + #wrong-turns (#4)
- **relations:** shared library field guard (the same incident class the guard now blocks); visible-content filter
- **verify-later:** whether schema-change→dependent-rebuild triggering exists (markPagesPendingRebuild covers regen; mid-build rewrites?)

<!-- SOURCE: U23_docs_root_vonc.md -->
### SQL template-surgery pattern (needle-gate discipline)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b merged entry: "Template-surgery pattern that held up" with the 2026-07-02 false-alarm refinement; practised across the marker/ghost-row/truncation fixes.
- **what:** Safe in-DB template edits: (1) needle-gate read — every needle as a LIKE boolean PLUS occurrence counts so partial coverage is visible BEFORE mutating (counts must be counted from the dump, not recalled); (2) shell backup of the full column; (3) guarded idempotent UPDATE (exact-string nested replace or anchored regexp_replace with backreference, plus NOT LIKE pre-state guard); (4) RETURNING boolean checks; (5) rollback file. Postgres pitfalls: regex quantifier bounds cap at 255; substring-with-parens returns the capture group; gradient-embedded hexes escape naive background regexes; needles containing literal % can't be LIKE-gated (use position()). Anchor REPLACEs on the opening tag (see marker lesson); dump→edit-offline→full-text UPDATE for multi-line blocks.
- **sources:** docs/016b_debugging_guide_merged(3).md#sql-verification-pitfalls; docs/fix_archive_template_display(1).sql (header); docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§8
- **relations:** marker anchoring; hidden-vs-author-CSS fix; sanctioned edit paths (this is the fallback)
- **verify-later:** n/a (practice)

<!-- SOURCE: U23_docs_root_vonc.md -->
### Marker/attribute REPLACE anchoring lesson (fix_marker_selector)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Bug introduced twice (provocation-card, lobby-grid), fixed via fix_marker_selector.sql with RETURNING checks (still_broken=f ×4), corrected HTML redeployed 2026-07-04; guide entry added.
- **what:** Adding an attribute by replacing the bare string `data-component="X"` also hits the section's own inline `querySelector('[data-component="X"]')`, producing a malformed two-attribute selector → SyntaxError → the cosmetic IIFE dies (loaders unaffected). Rule: anchor marker REPLACEs on the OPENING TAG (the copy followed by more attributes), revert only the in-selector copy (the one followed by `]`); better still, emit markers at generation.
- **sources:** docs/fix_marker_selector.sql (header); docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-marker-replace-broke; docs/016b_debugging_guide_merged(3).md#data-runtime-fill-marker-anchoring
- **relations:** generation-time guards (the prevention); SQL surgery pattern
- **verify-later:** n/a (lesson; instance fixed)

<!-- SOURCE: U23_docs_root_vonc.md -->
### `hidden` attribute vs author CSS (clone-template ghost rows)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Ghost-row fix verified end-to-end 2026-07-08 (rendered_len 7455→7671; live grep 2); prevention added to guide + component-creator requirement.
- **what:** The `hidden` attribute maps to UA-stylesheet `display:none`, which loses to ANY author `display` rule on the same element — so a hidden clone-template item inside a `display:grid` item class renders as a ghost row. Fix: a more specific author rule `[data-…-template] { display:none; }` in template AND instance (the REPLACE correctly fired twice — base selector + its mobile media-query copy). Prevention: component-creator must emit the hiding rule alongside `hidden` for clone templates.
- **sources:** docs/fix_archive_template_display(1).sql (header); docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-08; docs/016b_debugging_guide_merged(3).md#hidden-attribute-loses
- **relations:** generation-time guards; clone-template list pattern
- **verify-later:** component-creator prompt includes the hiding-rule requirement?

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Consolidated debugging guide (016)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Live 016_debugging_guide_v2_58_consolidated.md has 53 §9 failure subsections; archived v2_48 has a strict subset (comm diff: nothing in v2_48 absent from v2_58); top-level structure identical across v2_34/v2_48/v2_58.
- **what:** The canonical operational runbook: assumption checklist, pod health, work-item status, orchestration states, timeout chain, failed-item cleanup, and a large accreting "Specific Failure Patterns" catalogue plus the `detected→triaged→claimed` work-item state machine. Archived v2_34–v2_48 are monotonically-growing earlier snapshots; v2_58 supersedes them with added patterns (Thunder GPU provisioning, presign O(n²) bloat, tool/game pages never deploying, adoption slug-mangling by WriteSitePlanAction). No concepts were dropped between archived and live.
- **sources:** 016_debugging_guide_v2_48(1).md#9-specific-failure-patterns, #work-item-lifecycle
- **relations:** replacement = live 016_debugging_guide_v2_58_consolidated.md; snapshot-shadowing defect; tool widget clobber
- **verify-later:** live 016_debugging_guide_v2_58 §9 vs archived deltas

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Snapshot-shadowing defect (version+1000 snapshots outrank active rows)
- **category:** debugging
- **status-signal:** superseded
- **status-evidence:** ANALYSIS_phase_2f_two_defects (2026-05-11): `snapshot_agent()` inserts snapshots as `version+1000, is_snapshot=true`; a loader `ORDER BY version DESC LIMIT 1` without `is_snapshot=false` reads version 1001 over active version 1, shipping pre-migration workflow despite correct DB state; PLAN_imagery_loop_closure 2F "loader-snapshot defect" patched `processor.go::loadAgentDefinition` and `spawn_actions.go::getAgentDefinition` with `is_active=true AND (is_snapshot IS NULL OR is_snapshot=false)`.
- **what:** The model-swap/rollback snapshot mechanism creates rows that sort ahead of the active definition in version-descending queries, so any naive "most recent" agent-definition loader silently reads stale config. Structural, latent since launch; surfaced when Phase 2F first depended on a value that differed between active and snapshot rows. Fixed by adding the snapshot filter to loaders.
- **sources:** imagery/old/ANALYSIS_phase_2f_two_defects(1).md#defect-1; imagery/old/PLAN_imagery_loop_closure(9).md#2f
- **relations:** replacement fix in processor.go/spawn_actions.go loaders; 021_model_swap_and_rollback.sql snapshot_agent()
- **verify-later:** grep "FROM agent_definitions" *.go for is_snapshot filter; 021_model_swap_and_rollback.sql

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Adapter per-spawn-topic partition defect (kafka LeastBytes balancer)
- **category:** debugging
- **status-signal:** unknown
- **status-evidence:** ANALYSIS_phase_2f defect 2 (2026-05-11): git adapter wrote a success response to partition 1 of a single-partition per-spawn topic (`job.…responses-0` only), Kafka rejected "topic partition not found", response lost; did NOT reproduce on the 13:27 rerun after pod restart — "may be transient kafka-go metadata caching".
- **what:** kafka-go `LeastBytes` balancer may pick an out-of-range partition on a freshly-created per-spawn topic before partition metadata refreshes, dropping an adapter's success response while the underlying git commit succeeded (orchestration reports failure on succeeded work). Suspected to affect any adapter writing to per-spawn topics (webscrape, image-generator). Parked for monitoring, not fixed.
- **sources:** imagery/old/ANALYSIS_phase_2f_two_defects(1).md#defect-2
- **relations:** ANALYSIS_chassis_response_consumer_group_race.md (sibling parked defect); platform/kafka/producer.go
- **verify-later:** platform/kafka/producer.go Balancer; topic_manager per-spawn partition count

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Secret hygiene: image-provider API key rotation
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** HANDOFF_robot_hands_rebuild "SECURITY (highest): scrub + rotate STABILITY_API_KEY and BANANA_API_KEY (plaintext in logs; Banana on paid tier)"; TODO_imagery_followups(15) "SECURITY — STILL OPEN, STILL HIGHEST PRIORITY (do not let slide)."
- **what:** Image-generation provider API keys (Stability, Banana — paid tier) were being logged in plaintext; the standing highest-priority remediation is to scrub logs and rotate both keys. Repeatedly carried forward across imagery sessions without closure.
- **sources:** imagery/old/HANDOFF_robot_hands_rebuild(2).md#carried-forward; imagery/old/TODO_imagery_followups(15).md#security
- **relations:** image generation pipeline; adapter deployment; storage secrets
- **verify-later:** adapter logging of STABILITY_API_KEY/BANANA_API_KEY; secret rotation in personae-default-secrets

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Adapter-vs-chassis deployment drift
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** loop_closure(9) known issue "Adapter deployment vs chassis deployment (2026-05-14)": image-generator adapter (`dynamic_adapter.go`) is a separate K8s resource from the chassis (`generate_image_actions.go`); after a chassis rebuild+rollout the adapter may run an older binary — the 2H `generateImage: stability request` log line wasn't found on adapter pods.
- **what:** The image-generator adapter and the chassis are deployed as distinct Kubernetes resources, so chassis rebuilds don't refresh the adapter binary — leaving action-layer changes (e.g. 2H per-kind cfg_scale/negative_prompt) potentially inactive at the adapter. Recommendation: document which deployment carries which binary and add the adapter to the rebuild/rollout sequence.
- **sources:** imagery/old/PLAN_imagery_loop_closure(9).md#known-issues; imagery/old/HANDOFF_robot_hands_rebuild(2).md#carried-forward
- **relations:** image request shape (2H); Stability timeout 30→120s side-fix; multi-cluster dispatch
- **verify-later:** image-generator-adapter deployment vs chassis image tag; rollout sequence in Makefile

## Additional carried operational deltas (not standalone concepts)
From imagery/old TODO/STATUS + loop_closure "Known issues": `llm_call_log.agent_type` populated empty (params.AgentType not threaded to LogLLMCall, also noted in doc 009); dispatch loop not claiming `triaged` image items behind page work; FAILED orchestrations accumulating in `orchestration_states` with no cleanup; variant chain missing `site_id` so variant heroes skip `imagery_direction`; legacy `image_prompts` age-out reframed to operational deregistration rather than a `check_legacy_image_prompts_aspect`.

## Note on many-images/per-component/product-imagery/audit-loop/adoption-image-mirror/vision-auditor/provider-router concepts
This sub-agent also surfaced several imagery-domain concepts (imagery generation pipeline, audit-and-fix loop, structured site_plan_imagery Phase 2G, asset locking 2A-2E, image-build-handler storage architecture, image request shape 2H, adoption image mirror, many-images-per-page, icon rendering via Lucide, product imagery via affiliate_products, vision-capable LLM auditor, image provider router, news feed pipeline + enrichment + price-aware filtering, rebuild-vs-rerender, files_field deploy dependency, tool generation pipeline, component schema-contract drift, cross-cluster Postgres/Kafka topology, multi-cluster dispatch MVP gaps) that substantially overlap U10 (imagery) and U09 (adoption) which were separately extracted with fuller code-scope access. Consolidation should de-duplicate against U09/U10 rather than re-litigate; where this unit's evidence adds a NEW dated fact (e.g. the snapshot-shadowing defect, the kafka partition defect, the RAG/GPU-infra superseded lineage), it is retained above as its own entry.

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Launcher reply-topic own-vs-parent derivation (Decision D4)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(39) "Update — 2026-06-02 16:12: D4 CONFIRMED live … the adapter's reply went to system.agent.generic.responses — the agent's own ExecutionContext.ResponsesTopic"
- **what:** An intermediate adapter reply must be routed to the agent's own `ExecutionContext.ResponsesTopic` (seeded from `__my_responses_topic__`), NOT `__parent_responses_topic__` (which is only for the child→parent final notification). The inherited handoff had this backwards; provision/decommission always used own-topic and worked. The same class of bug bit `dispatch_thunder_ssh_get_status` (cloned from ssh_exec) and was fixed to prefer `execCtx.ResponsesTopic`.
- **sources:** phase5/NOTES_phase5_training_launcher_running(39).md#3 (D4), #6, #10; docubundle/.../STATUS_thunder_adapter_2026-06_04.md#update-2026-06-04
- **relations:** corrects the superseded 2026-05-24 handoff claims; a latent same bug remains in ssh_exec dispatch if fired top-level
- **verify-later:** thunder_prepare_object_url_dispatch.go, thunder_ssh_exec_dispatch.go, thunder_ssh_get_status_dispatch.go; coordinator determineResponsesTopic

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### gpu-provisioner output shape flattening (output_fields plural)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(39) "Update — 2026-06-03 ~17:5x: 104 written" — "extractWorkflowResult … reads output_fields — PLURAL only. The gpu-provisioner complete uses output_field (SINGULAR) … falls to the fallback branch"
- **what:** `call_launcher` failed on `provisioning_result.provisioning_id not found` because gpu-provisioner's `complete` step used singular `output_field` (which `extractWorkflowResult` never reads), so its result came out step-name-keyed as `{dispatch_provision, input_data}`. Migration 104 fixed the provisioner's `complete` to plural `output_fields:["dispatch_provision"]` and re-pointed the launcher mapping to `provisioning_result.dispatch_provision.provisioning_id`; a proper chassis fix (honour singular output_field) was vetoed in favour of making the non-compliant agent conform.
- **sources:** phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-03-15:47, #17:4x, #17:5x
- **relations:** launcher input-mapping contract; same singular bug latent in thunder-reaper
- **verify-later:** extractWorkflowResult; agent_definitions gpu-provisioner (0bf9fa8a); migration 104

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Loop-await send-before-register race + preRegisterAwaitedRequest fix
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(39) "Update — 2026-06-09: pre-register fix CONFIRMED in prod … Every presign_checkpoints_iter_N_presign_one logged ClaimAwaitedRequest: status_before=waiting … claimed:true … The send-before-register race is closed"
- **what:** The central thunder-checkpoint-race bug: the local dispatch `dispatch_thunder_prepare_object_url` produced the adapter request and returned `await_response:true` BEFORE the coordinator inserted the `awaited_requests` row, so a fast (~1s) reply beat the insert, `ClaimAwaitedRequest` (WHERE status='waiting') found nothing, the reply was dropped, and the timeout re-dispatched every ~3 min forever. `spawn_agent`/`call_agent` avoid it via `preRegisterAwaitedRequest` (register-before-send, `ON CONFLICT (request_id) DO NOTHING`). Fix: call the same helper in the dispatch before `ProduceWithValidation` (guarded `if params.DB != nil`); note the helper's hardcoded 120s timeout_at then pins every presign await.
- **sources:** docubundle/.../HANDOFF_2026-06-06_checkpoint_upload_loop_await_race.md; docubundle/.../CONTEXT_PACK_thunder_checkpoint_race.md; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-06-3, #2026-06-08, #2026-06-08-2
- **relations:** fourth cause of the `awaited_requests`-stuck-waiting symptom (016 §9); the batch presign superseded the loop that exposed it
- **verify-later:** thunder_prepare_object_url_dispatch.go; spawn_actions.go preRegisterAwaitedRequest; awaited_requests table

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Probe debugging-guide entries #24–#28
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-12 "Debug guide updated … 016_debugging_guide_v2_46.md"; 2026-06-13(g) "Debug guide v2_48 … #27 invented-interface … #28 agent_definitions UNIQUE(type,version)".
- **what:** Reusable pitfalls harvested from probe execution: #24 a config/workflow file is only authoritative at its runtime read-path; #25 prove the test harness delivered the intended input before debugging; #26 shell vars need export/prefix and die with the session; #27 invented interface (compiles standalone ≠ satisfies interface — wire to registry early); #28 agent_definitions UNIQUE(type,version) + two look-alike category columns.
- **sources:** traffic_probe_running_notes(27).md#2026-06-12-debug-guide, traffic_probe_running_notes(27).md#2026-06-12-provisioning-ran, traffic_probe_running_notes(27).md#2026-06-13-g
- **relations:** #24 is the stale-artifact class; #27 fixed backend_unreachable; #28 fixed the agent INSERT
- **verify-later:** 016_debugging_guide_v2_48.md entries #24–#28

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### A4/homepage-missing-file — root cause hypothesis evolution to "auto-complete on lost response"
- **category:** debugging
- **status-signal:** superseded
- **status-evidence:** running_notes_14(20) Part 10 (2026-06-03): "Candidate cause (NOT concluded): ...silent non-fast-forward [git race]. Alternatives not ruled out: empty-assembly skip... or a path bug." CATALOGUE(3)→(4) diff (2026-06-04): "*Cause (updated... git race RULED OUT)*: ...empty-assembly case" superseded again by "*Cause PINNED*: ...work item is `complete` with error 'Auto-completed: work verified done despite lost response.'" Running_notes Part 11 confirms: "Root cause: the homepage's content build was dispatched, the handler's response was lost... and the recovery path optimistically auto-completed the work item without verifying the artifact."
- **what:** The homepage (`index`) was `build_status='deployed'`+`stamped` in the DB with zero rendered components and no committed file — three successive hypotheses (git-commit race, empty-assembly/planner-vs-composition gap, and finally the pinned cause) were tested and discarded in turn before landing on: a scheduled task's SQL `pre_query` (`claimed-item-timeout`) auto-completed a claimed work item using loose evidence ("any page on the site updated since claim") after the handler's response was lost to a pod death, without checking that *this* page actually produced components.
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Parts 10–12; CATALOGUE_gamesdesign_post_sync_fix_defects(3).md vs (4)
- **relations:** claimed-item-timeout evidence-check reliability mini-project (below); silent-completion family (page-build-handler, save_page_sections)
- **verify-later:** `migration_claimed_item_timeout_evidence_v2.sql` application state; `v3_site_actions_optionB.patch` deployment.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Sectionless-page silent completion (guide-skinner-box) + durability stack
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** running_notes_15(5) Part 1: "`complete_error` is a `complete_workflow` (SUCCESS) with message *'Content writer skipped — page has no sections defined.'* — the silent-success smell." Part 8: "Wrote `discovery_checks/check_sectionless_pages.go` (new check...)" with an explicit 4-layer "durability stack" logged, item 3 of which is marked "(next, optional for this gap; cleanliness)" i.e. not yet done at time of writing.
- **what:** A page reaching page-build with zero planned sections (`pages.sections=[]`) causes `check_has_ready_sections` to route to `complete_error`, which is a SUCCESS-labelled `complete_workflow` — so a genuinely broken page is marked `complete` and never retried. Root-caused (after correctly ruling out the convergence-union code, confirmed correct) to "the gap is reconciliation: nothing repairs a page in-plan with zero sections." Fix stack: (1) a read-time fallback in `load_page_sections_from_spec_action.go` that copies a same-role sibling's section layout ("skeleton only, not content") when both known sources are empty; (2) a new discovery check `check_sectionless_pages.go` that detects and retriggers stuck sectionless pages (chosen over patching the existing but **dormant** `checkEmptyPageSections`, see below); (3) a workflow-level fix so the genuinely-unrecoverable case routes to a flagged state instead of `complete` — logged as not yet shipped; (4) the broader positive-evidence-completion mini-project (shared with A4).
- **sources:** adoption/running_notes_15_skinner_box_and_adoption_sections(5).md Parts 1–8
- **relations:** dormant discovery-check machinery (below); A4 auto-complete-on-lost-response; FOCUS_page_build_handler_silent_completion.md
- **verify-later:** whether S2 (workflow-level flagged-state fix) ever shipped; `check_sectionless_pages.go` enablement state.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `save_page_sections` content-regression guard laundered into false success — theories falsified in sequence, course-corrected to a second mechanism
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** running_notes_17(16) "index deep-dive": four theories tried and explicitly falsified in order — (1) "Load — 21 rebuilds... starved the dispatcher. FALSIFIED"; (2) "Concurrent production deploy cycled the pods mid-flight. FALSIFIED for index"; (3) "index's build DURATION exceeds the... claim lease. FALSIFIED"; (4) caller-timeout theory, "partially real as a STATUS artifact... but this is noise, not the defect." Landed on: "**Content-regression guard... is the leading mechanism.**" Then a further course-correction: "**COURSE-CORRECTION: a second mechanism — page_components LOCKING**... NEW HYPOTHESIS (at least as plausible as the regression guard)."
- **what:** The homepage (`index`) repeatedly failed to rebuild despite the work item showing `complete` and git successfully committing a file — the committed file was stale (unchanged since 2026-06-06). Root cause hunt discarded four increasingly specific theories (load, concurrent deploy, claim-lease timeout, caller/callee timeout mismatch) before finding `save_page_sections_action.go`'s **content-regression guard** — a real safety check (refuses to overwrite existing deployed content with much-shorter new content) whose error return was silently laundered into `complete_error`, itself a SUCCESS-labelled `complete_workflow`. Before fully confirming this, the investigation surfaced a *second* candidate mechanism discovered via schema inspection — a `page_components` row-locking subsystem with an `auto_lock_on_deploy` trigger — and explicitly walked back single-mechanism confidence pending a discriminating query. Two distinct, independently-real bugs were named regardless of which mechanism fires: (1) the guard's legitimate refusal shouldn't route through `complete_error`; (2) deploy shouldn't proceed (re-render + git-commit) after a zero-row save.
- **sources:** content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(16).md, sections "index reproducibly stale" through "COURSE-CORRECTION"
- **relations:** A4 auto-complete-on-lost-response; sectionless-page silent completion; page-build-handler silent-completion family generally
- **verify-later:** which mechanism (regression guard vs component lock) actually fires on `index`; `page_build_handler_save_failure_visible.sql` application state; `auto_lock_on_deploy()` trigger function body.

<!-- SOURCE: U25_leopardess_social.md -->
### Silent no-op success class (`complete_error` and the ten empty builds)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** NOTES_provocations-index (2026-07-04→07): "SEVEN work items … ALL 'complete', no errors … A success status masked a no-op for two weeks"; HANDOFF §3 "ten builds 'completed' having built nothing"; preventions partially shipped (sectionless_pages enabled 2026-07-10).
- **what:** The defining failure shape of the platform: error paths implemented as successful completions. Canonical case: the planner emitted a page with no sections, and page-build-handler routes zero-sections to a step literally named complete_error — a complete_workflow reporting success ("Content writer skipped"). Diagnostic signature: a work-item result carrying only site_record (healthy runs emit sections_saved + deploy_result). Framework preventions specified: planner invariant (every planned page whose role page-build-handler builds must have ≥1 section, with an explicit role→pipeline map), fail-loudly on the zero-sections path, auditor rules for planned-but-linked pages, post-deploy URL presence checks. sectionless_pages (the exact detector) existed but was enabled nowhere until 2026-07-10.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-index(4).md; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#3, #9.1
- **relations:** discovery check wiring gaps; verify-by-artifact discipline; page section source precedence (the unblock)
- **verify-later:** page-build-handler workflow complete_error step; sectionless_pages check enablement

<!-- SOURCE: U25_leopardess_social.md -->
### Problem-category taxonomy for component/tool defects
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** In active use across every NOTES file in this unit ("Categories:" lines); seed set defined 2026-06-29, extended in practice (assembly-drop, planning-gap, silent-noop-success, cta-graph, css-specificity, method-correction).
- **what:** A shared greppable vocabulary tagging every incident so patterns roll up into the global debugging guide: css-variable-mismatch, empty-shell/mode-b-template, broken-template-slots, content-vs-runtime-mismatch, detool-on-rebuild, js-not-extracted, js-bundle-stale, schema-template-drift — plus organically-added tags. When a category recurs across tools it graduates to a global pattern with a systemic fix (exactly how the empty-shell and visible-content-filter issues surfaced).
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/TOOL_DOCS_convention(3).md#Problem-category-taxonomy; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocation-card(12).md (tags in use)
- **relations:** per-tool travelling docs; debugging guide 016b
- **verify-later:** 016b entries fed from these categories

<!-- SOURCE: U25_leopardess_social.md -->
### Editing-stored-HTML landmines (marker anchoring, hidden-vs-author-CSS, offline edits)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** HANDOFF §8 "all paid for in this thread" (fix_marker_selector.sql, fix_archive_template_display.sql both shipped); VERDICT length-delta verification used twice.
- **what:** Hard-won rules for touching stored templates/HTML: a marker/attribute REPLACE must anchor on the opening tag, never a bare attribute (the attribute string also appears inside the component's own querySelector and a plain replace corrupts it — happened twice); the `hidden` attribute is UA-stylesheet display:none and loses to any author display rule (clone templates render as ghost rows without an explicit [data-…-template]{display:none}); multi-line block removal is dump → edit offline → UPDATE full text → verify by length delta (never multi-line SQL REPLACE of nested markup); better still, emit markers at generation.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#8; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-archive-list.md#2026-07-08; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-index(4).md#2026-07-07
- **relations:** generation-time guards (the prevention); section-editor path
- **verify-later:** n/a (lessons; fixes live)

<!-- SOURCE: U25_leopardess_social.md -->
### Stuck-claim / zombie-handler dispatch noise
- **category:** debugging
- **status-signal:** unknown
- **status-evidence:** RUNBOOK_minilobby dispatch note (2026-07-10): "manual dispatch via 087 needed five passes because page-build handlers intermittently left items stuck at claimed without spawning (survived across the v1.0.1107 deploy)"; also "the known stuck-claim/zombie-handler noise recurred" 2026-07-12. Root cause not established.
- **what:** Recurring operational failure: dispatched items sit at claimed with no spawned handler, or late handler reports mark completed work failed. Recovery is documented (reset to triaged, NULL claim fields, re-run the dispatch pass; close by artifact when the work actually happened) but the underlying cause is unresolved — a live reliability question for the dispatch/spawn path.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#0 dispatch note; docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-12; docs/social001_vonc_tiktok_social/minilobby_task/087_dispatch_work_items_vonc.sh (header)
- **relations:** work-item dedup; leopardess O4 unstick procedure (same class)
- **verify-later:** claim/spawn/call sequence in build-dispatch-loop; agent_error_log around stuck claims

<!-- SOURCE: U26_misc_dirs.md -->
### Workflow monitoring REST endpoints
- **category:** debugging
- **status-signal:** unknown
- **status-evidence:** 004-agent-chassis-architecture.md lists GET /monitor/workflows, /monitor/workflow/{id}, /monitor/stuck?hours=n, /monitor/metrics as built ("Each agent exposes monitoring endpoints") but no later doc in this unit uses them — operational debugging instead goes through psql/db-inspector.
- **what:** Per-agent HTTP monitoring API over orchestration state: list active workflows per client, inspect a workflow's execution path/state, find stuck workflows not progressing for N hours, and aggregate metrics. Complemented by per-step execution_path timing records and execution_metadata counters in the state row.
- **sources:** docs/architecture/004-agent-chassis-architecture.md#monitoring-and-observability
- **relations:** database-backed workflow state; kcat/db-inspector runbook (the surviving practice); current debugging docs (016/016b spine)
- **verify-later:** /monitor routes in chassis HTTP server code

<!-- SOURCE: U26_misc_dirs.md -->
### kcat + db-inspector operational runbook
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** basic_usage/001 and 004 are working command logs (with real outputs pasted, e.g. correlation IDs returned, "0 rows" failure cases) for triggering and tracing workflows in the live cluster.
- **what:** The early ops playbook: scale deployments up/down; inject workflow-start messages via kcat from an in-cluster pod with full header sets; fetch the latest correlation_id from orchestrator_state; watch progress with the db-inspector tool (-watch); trace specific agents by finding spawned instance IDs then grepping shared chassis pod logs (agents don't get dedicated pods); check consumer-group lag, response topics, ServiceAccount job-creation rights, and events for spawned jobs.
- **sources:** docs/basic_usage/001basic_usage.txt; docs/basic_usage/004_debugging
- **relations:** agent spawning; website-builder group; current debugging spine (016)
- **verify-later:** tools/db-inspector, tools/kafka-producer existence; whether runbook matches current namespace/topics
