# EXTRACTION U24f — docs/_archive gap-fill: imagery/old top levels, small archived
project leftovers, sql_for_agents/sql_for_tables, and the nested docs019 archive-of-archive

Extracted 2026-07-13. Files in scope: 46. Concepts found: 21.

## Scope note on the two zero-file directories

`find docs/_archive/agent_docs/docs024_key_docs_latest/imagery -maxdepth 1 -type f` and
the equivalent for `.../old` both return **empty**. Both directories contain only a
single excluded subdirectory (`imagery/old/`, `old/older1/`) and no files of their
own at the top level. This is not an extraction gap — there is nothing at that depth
to read. Recorded here so the "every file accounted for" audit trail is explicit
about why these two paths contribute zero rows to the coverage table below.

## Coverage

| file | treatment |
|---|---|
| docs/_archive/agent_docs/docs004_website_capture_project/playwright/002_claude_thought_process.md | skipped-generated (0 bytes, empty file) |
| docs/_archive/agent_docs/docs007_brochure_builder/002_brochure_test_guide.md | skipped-generated (0 bytes, empty file) |
| docs/_archive/agent_docs/docs020_llm_training_rag/007_rag_deployment_README.md | full |
| docs/_archive/agent_docs/docs025_ai_chatbot_idea_uk/excellent_discussions/PLAN_isolated_chat_environment(3).md | full |
| docs/_archive/agent_docs/sql_for_agents/007_content_creator.sql | skipped-generated (header banner only, no SQL statements follow) |
| docs/_archive/agent_docs/sql_for_agents/028_page_builder_agent.sql | skipped-generated (0 bytes, empty file) |
| docs/_archive/agent_docs/sql_for_agents/097_api_news_grok.sql | skipped-generated (0 bytes, empty file) |
| docs/_archive/agent_docs/sql_for_agents/sql_for_agents_v1/018_briefing_questionnaire.sql | skipped-generated (0 bytes, empty file) |
| docs/_archive/agent_docs/sql_for_agents/sql_for_agents_v1/021_multipage_planning.sql | skipped-generated (0 bytes, empty file) |
| docs/_archive/agent_docs/sql_for_agents/sql_for_agents_v2/001_validator_sql.sql | full |
| docs/_archive/agent_docs/sql_for_tables/002_links_clients_networks_etc_tables.sql | full |
| docs/_archive/agent_docs/sql_for_tables/040b_migration_cleanup_bare_guide_duplicates(1).sql | full |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/archive_april_26/022b_companies_house_matching_cascade_plan_v2.md | full |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/go_files_old/analyser(2).go | header-scan |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/go_files_old/assembler(2).go | header-scan |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/go_files_old/dbcontext(1).go | header-scan |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/go_files_old/go_files/embed(1).go | header-scan |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/go_files_old/go_files/embed.go | header-scan |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/go_files_old/go_files/eval_targets.go | header-scan |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/go_files_old/go_files/fuse.go | header-scan |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/go_files_old/go_files/groundtruth_targets.json | header-scan |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/go_files_old/go_files/resolve_targets(1).go | header-scan |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/go_files_old/go_files/resolve_targets.go | header-scan |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/old/012d_tool_lifecycle_guide_v4.md | family-delta (byte-identical via md5 to docs/agent_docs/docs024_key_docs_latest/archive_april_26/012d_tool_lifecycle_guide_v4.md; zero delta) |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/thin_slice_run/analyser(2).go | family-delta (byte-identical via md5 to go_files_old copy above; zero delta) |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/thin_slice_run/assembler(2).go | family-delta (byte-identical via md5 to go_files_old copy above; zero delta) |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/thin_slice_run/dbcontext.go | header-scan (earlier variant, lacks -runtime-site/-runtime-page flags present in go_files_old/dbcontext(1).go) |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/thin_slice_run/thin_slice_constitution.md | full |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/GUIDE_deploy_from_context_packs.md | full |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md | full (family-latest) |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(7).md | family-delta (headers identical subset of v20; two whole sections — "Building discipline", "Reuse-checking" — added later, nothing dropped) |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/PLAN_change_layer_integration_contract(1).md | family-delta (diffed against (3); wording/detail refinements only, no dropped concepts) |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/PLAN_change_layer_integration_contract(3).md | full (family-latest) |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/PLAN_trust_ledger_contract(1).md | full |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/RUNBOOK_thin_slice(10).md | full |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/adapter(4).go | header-scan |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/analyser_request_action(1).go | header-scan |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/main/001_development_guide(3).md | family-delta (header structure — 186 headers — identical to live docs024_key_docs_latest/001_development_guide(5).md; superseded snapshot, no delta) |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/main/016_debugging_guide_addendum_adopted_tools_no_widget(3).md | full |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/main/016_debugging_guide_v2_21_.md | family-delta (headers confirmed pure subset of live 016_debugging_guide_v2_58_consolidated.md; no delta) |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/main/016_debugging_guide_v2_36b.md | family-delta (headers confirmed pure subset of live v2_58; no delta) |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/main/026_component_regeneration_flow.md | family-delta (headers confirmed pure subset of live 026_component_regeneration_flow(2).md; no delta) |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/main/030_phase1_plan_and_reconciler(2).md | family-delta (headers identical to live 030_phase1_plan_and_reconciler(5).md; byte-identical to sibling (3).md below) |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/main/030_phase1_plan_and_reconciler(3).md | family-delta (byte-identical via md5 to (2).md above; zero delta) |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/main/FOCUS_imagery_assessment.md | family-delta (headers confirmed pure subset — sections 1-8 only — of live imagery/FOCUS_imagery_assessment.md, which continues to section 13; no delta) |
| docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/go_files/contextkit/README(1).md | full |

## Concepts

### contextkit target-resolution & bundle-assembly toolchain
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** RUNBOOK_thin_slice(10).md: "The thin slice has been reasoned about but not yet run on a real task and checked." README(1).md's directory map shows a standalone `contextkit` Go module (7 commands) plus a `chassis-drafts/analyser-adapter` staging tree destined for the real repo — and that adapter is confirmed **built and deployed**: `internal/adapters/analyser/`, `cmd/analyser-adapter/`, and a live `deployments/kustomize/services/analyser-adapter/` overlay all exist in the working tree.
- **what:** A seven-command Go pipeline for assembling task-scoped LLM context bundles from a codebase: `analyser` (AST walk → JSON structural summary of package/imports/functions/types with line ranges), `resolve_targets` (deterministic lexical-overlap baseline that proposes scope candidates), `embed` (semantic vector index over symbols, Ollama-backed with a non-semantic offline stand-in for pipeline-proving), `fuse` (reciprocal-rank fusion of lexical + semantic candidate lists, k=60), `eval_targets` (recall@N / MRR scorer against a hand-authored `groundtruth_targets.json`), `assembler` (renders the final paste-ready bundle: constitution + task + in-scope code in full + neighbourhood signatures + schema + a "what was left out" pointer note), and `dbcontext` (shells out to psql for schema/rows/runtime-evidence with multipass row sizing — never an unbounded dump). Two contracts (`internal/analysis`, `internal/candidates`) are defined once and shared across commands.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/go_files/contextkit/README(1).md, docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/RUNBOOK_thin_slice(10).md, docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/go_files_old/{analyser(2).go,assembler(2).go,dbcontext(1).go,go_files/embed.go,go_files/resolve_targets.go,go_files/fuse.go,go_files/eval_targets.go}
- **relations:** flat-file constitution (below); reuse-check retrieval pipeline design (below); adapter response-envelope contract (below, the chassis-integration half); fix-loop council (docs024 fixloop_eg_dartsonline)
- **verify-later:** internal/adapters/analyser/, cmd/analyser-adapter/, deployments/kustomize/services/analyser-adapter/ — confirm whether the standalone contextkit CLI tools themselves (analyser/assembler/embed/fuse/eval_targets/dbcontext binaries) were ever run on a real task per the runbook's "first real run" checklist, or whether only the adapter integration shipped.

### adapter response-envelope contract (request/response wiring conventions)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** adapter(4).go's header states the envelope is "grounded in the orchestrator (coordinator.go) and the three working adapters, not the docs — which disagree (003 vs FOCUS) and were resolved empirically." internal/adapters/analyser/ exists live in the repo, confirming the draft shipped.
- **what:** A reverse-engineered (from working adapters, not from docs) contract for how an adapter must shape its Kafka request/response envelope so the orchestrator actually routes the reply instead of timing out: action comes from `body.action` not headers; `in_response_to_request_id` (echoing the incoming `request_id`) is the load-bearing claim field in `coordinator.go`'s `ProcessResponse`, with `request_id` as fallback; the reply body must use canonical `types.ResponseHeaders` via `ToResponseHeaders` so `is_complete`/`is_error` marshal as real JSON bools (the "bool trap" — a `map[string]string` sending the string `"true"` fails the receiver's struct-bool unmarshal); sends must go via `ProduceWithValidation`, never plain `Produce`. websearch-adapter is flagged as the one adapter still on the deprecated string-bool/plain-Produce path.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/adapter(4).go, docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/analyser_request_action(1).go
- **relations:** contextkit toolchain (above); analyser-adapter service
- **verify-later:** platform/orchestration/coordinator.go ProcessResponse, internal/adapters/analyser/adapter.go, whether websearch-adapter has since been migrated off the string-bool map

### deploy-mechanics taxonomy (six ways a change ships)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** GUIDE_deploy_from_context_packs.md is written as an operational runbook with concrete `make`/`kubectl`/`kcat` commands and per-project worked examples (gamesdesign adoption, thunder checkpoint race, idea.uk go-live) dated against real image tags (`v1.0.1057`).
- **what:** A named taxonomy of the distinct ways a change actually reaches production in this system, used to scope any task before touching it: (A) chassis platform image — Go code changes need a rebuilt/pushed/retagged image and a k8s rollout; (B) database — SQL/migrations via psql, snapshot-first, re-query to verify; (C) work-items — insert a `site_work_items` row for the dispatch loop to claim; (D) orchestration trigger — a kcat `orchestrate` message to `system.agent.generic.requests`; (E) generated static sites — downstream/automatic via git → GitHub Actions → Backblaze once `build_status='deployed'`; (F) the idea.uk binary — a separate non-k8s, file-based Go binary with its own build/scp/restart cycle. Cross-cutting cautions: bump image tags or a rollout won't pull the change; "complete" is not "succeeded" — verify positive evidence, not just terminal status.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/GUIDE_deploy_from_context_packs.md
- **relations:** contextkit toolchain (above); deployment-github
- **verify-later:** Makefile targets referenced (build-*, deploy-agents, update-kustomization-images)

### flat-file constitution (precursor to the `standards` table, scope=constitution)
- **category:** contracts-and-standards
- **status-signal:** aspirational
- **status-evidence:** "This is the flat-file version for the thin slice. Later it becomes the `standards` rows with `scope = constitution`; the content is the same."
- **what:** A single always-on rules document meant to be included in full in every LLM context bundle, distinct from the task-specific 003 contracts which are pulled in only when a task touches them. Covers reuse-before-recreate, structural-over-symptomatic fixes, one-orchestrator-per-agent, no-subworkflows-in-SQL, the snake_case/kebab-case string-naming test, chassis storage conventions (text+CHECK not native enums, version+previous_version_id, deleted_at not status=archived), logging discipline (no `logger.Debug`, always log the orchestration_id), and deployment/namespace facts. Explicitly designed to later graduate from a hand-pasted flat file into database rows.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/thin_slice_run/thin_slice_constitution.md
- **relations:** contextkit toolchain (above); trust ledger contract (below, references a `standards` lifecycle)
- **verify-later:** whether a `standards` table with `scope='constitution'` was ever actually created

### trust ledger (bidirectional trust ratchet, per-tenant per-capability)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** "Status: contract specification... Now given concrete shape." — a design document, no implementation claimed. A later/fuller version of this exact contract exists live at docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/PLAN_trust_ledger_contract(4).md (outside this unit's scope).
- **what:** A `trust_ledger` table (one row per tenant×capability) holding a mutable `trust_level` (`confirm_every`/`confirm_exceptions`/`notify`/`autonomous`) that cascade routers read to floor the production tier and gate-policy engines read to decide autonomy — derived from, but distinct from, the immutable append-only `decision_log`. The capability's **ceiling** (max reachable trust, set by verifiability × containment) lives on a separate capability catalog, not the ledger row, so it's a property of the capability, not the tenant. Mutation is asymmetric: graduation (trust up) is always confirm-not-initiate via a `config_work_items` proposal; de-graduation (trust down) may auto-apply with notification on severe evidence — "losing trust is reversible; falsely gaining trust is what allows mistakes to apply unsupervised."
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/PLAN_trust_ledger_contract(1).md
- **relations:** change-layer integration contract (below, feeds the ledger's evidence_summary); governance/HITL principles (below); the fuller live PLAN_trust_ledger_contract(4).md (not in this unit's scope)
- **verify-later:** any `trust_ledger` or `capabilities` table in the live schema

### change-layer integration contract (change_events, trigger filter, in-band emission)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** "Status: contract specification... Closes the final contract gap before implementation." A fuller live version exists at docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/PLAN_change_layer_integration_contract(4).md (outside this unit's scope).
- **what:** Defines how code/doc diffs reach a maintenance agent: a first-class `change_events` table (source ∈ git_webhook/polling/in_band/periodic_sweep/manual, at-least-once with commit_id dedup) feeding a **trigger filter computed from the mechanical config** (not stored — so it self-updates when doc/code paths move) that fans out into typed triggers (`conventions_reextraction`, `schema_check`, `code_audit_refresh`, `reuse_index_refresh`, `intent_revalidation`, `freshness_check`). The `in_band` source is the mechanism that "closes the loop on self-modification" — when the tool's own bundle-builder or a layer agent applies a confirmed change, it emits its own change event so the drift detector doesn't go blind to its own effects; a scoped guard prevents a just-confirmed entry from re-triggering on itself while still letting genuine downstream effects (e.g. auditing existing code against a newly confirmed convention) fire.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/PLAN_change_layer_integration_contract(3).md (family-latest in this unit), PLAN_change_layer_integration_contract(1).md (delta-checked, no drops)
- **relations:** trust ledger (above); reuse-check retrieval pipeline (below, reuse_index_refresh trigger)
- **verify-later:** any `change_events` table in the live schema

### context substrate model (authored vs derived, salience over presence)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Framed explicitly as "a list of synthesis points... captured as reusable lenses for further design work, not as a final theory."
- **what:** A framing for how LLM context should be built: documentation/standards/intent are operational inputs to generation, not passive reference, and split into two epistemic categories — **authored** (has an owner and lifecycle, can be wrong, needs maintenance) vs **derived** (no-owner true-right-now readout, can only be current or superseded; source code sits on this line). The **change layer** (diffs) is derived-but-narrative — the natural audit/learning surface. Authored layers should hold **references, not copies** of derived material so they don't drift when reality moves. Two staleness modes need two different fixes: authored drift is fixed by keeping authored content thin and pointer-rich; derived snapshot-staleness is fixed by fetching at reasoning time, not paste-time. LLMs lose the big picture from **salience, not window size** — local detail crowds out context mid-reasoning even when the text is still "in the window."
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §Context substrate
- **relations:** contextkit toolchain (above); flat-file constitution (above)
- **verify-later:** none (a design framing, not a built artifact)

### mediator model for competing design concerns ("right" as requirement-relative balance)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Same running-synthesis document, framed as an unbuilt lens.
- **what:** A model for resolving conflicting design dimensions (fast/secure/generic/simple/functional) as a requirement-relative balance rather than a pick-one-winner or naive-merge: authored solutions are treated as extremes that bound the solution space, and a mediator finds the point inside it that the requirement's priority profile dictates (ordered priority, not numeric weights, since real-world priority shifts arrive as "X now outranks Y"). A satisfied concern demotes from active author to passive checker (re-promoting if a later change breaks it) — unifying "checker" and "multi-author" as two modes of one process. Non-convergence among concerns is treated as the genuine escalation signal, isolating the one real tradeoff that needs human judgement from everything else that settles on its own. Multi-author surfaces tradeoffs vividly but cannot resolve value-laden conflicts — it's an option-generation engine, not a decision engine.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §"Right" as balance, not a single answer
- **relations:** governance/HITL principles (below)
- **verify-later:** none (a design framing, not a built artifact)

### governance/HITL principles (confirm-not-initiate, decision publishing, sealed inheritance)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Same running-synthesis document; principles, not shipped mechanisms.
- **what:** A cluster of governance rules for an autonomous build-and-operate system: **confirm-not-initiate** (agent-led reasoning, human confirms via a decision package, never authors from scratch); **every decision publishes its reasoning**, since drift detection is only possible because premises are logged and can be compared to the current premise; **two precedence directions in inheritance** — normal entries are child-wins (local refinement) but sealed constraints are ancestor-wins (legal floors, mission non-negotiables), so a leaf can't defeat a new law by prior relaxation; **three resolutions to a doc/code disagreement** (code drifted / doc drifted / legitimate exception) with a configurable default presumption that the human can always override; **one path to a privileged state transition** (e.g. `proposed → active`) routed through a single central confirmer rather than reimplemented per producer, so confirm-not-initiate is airtight in one place; **newer supersedes pending** — a fresh proposal for an already-pending target expires the older one rather than blocking on staleness.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §Governance and HITL
- **relations:** trust ledger (above); change-layer integration contract (above); mediator model (above)
- **verify-later:** none (a design framing, not a built artifact)

### reuse-check retrieval pipeline design (catalog → lexical/structural → embeddings → rerank)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** Directly implemented by the contextkit toolchain's `resolve_targets`/`embed`/`fuse` commands (see contextkit concept above, status partial/unrun); the design principles themselves are framed as reusable lenses.
- **what:** A layered design for "has this already been solved" reuse-checking that treats it as a retrieval problem with a judgement tail, not a generation problem: a maintained capability catalog is the cheapest check (lookup, not search); "identical" (token/AST fingerprinting, algorithmic, high-precision) and "similar" (semantic, embeddings) are split into different mechanisms because lexical/structural matching misses genuine near-duplicates with different names; every narrowing layer is tuned for **recall over precision** since a false-negative reuse check manufactures confident duplication that's worse than no check at all; a cheap model narrows the candidate set, a strong model decides on the shortlist — never the reverse; near-duplicate detection runs post-generation against a concrete draft (a real artifact to fingerprint), while fuzzy "what's there to build on" retrieval runs pre-generation. Signature+docstring embeddings are framed as a general retrieval substrate (also serving target resolution and capability-catalog curation), not a narrow dedup optimisation, and — at the scale of a few thousand symbols — need no vector database, just in-memory cosine.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §Reuse-checking; contextkit toolchain go files (embed.go, fuse.go)
- **relations:** contextkit toolchain (above); change-layer integration contract (above, reuse_index_refresh trigger)
- **verify-later:** whether any capability catalog or reuse index has been built beyond the contextkit prototype

### autonomous-system building-block hardening checklist
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Framed as "edge cases caught before building" in the running-synthesis notes — a pre-implementation checklist, not a verified implementation.
- **what:** A catalog of structural safety patterns for building the autonomous-operate machinery itself, distilled from design review: self-referential structures (trees, version chains) need a cycle guard on write plus a detect-and-fail walk; a multi-step apply (writes + an emitted event) must be all-or-nothing via one transaction with an outbox, or a mid-crash leaves a live row with no log/event; assembling from several tables needs one consistent point-in-time snapshot; "at most one live X per target" must be enforced at every layer down to the underlying row, not just the queue; bulk operations need bulk confirmation (per-item confirm doesn't scale to an onboarding flood); transient/infrastructure failures must be filtered out before they're allowed to lower a capability's trust; derived indexes/caches go stale silently and need a freshness stamp; recovery must not depend on the thing being recovered (the rollback path can't route through the agents it's rolling back); blast radius caps the trust ceiling regardless of verifiability, because self-modification is a residual risk that's managed (conservative early trust, human-in-the-loop, external rollback), not solved.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §Building discipline
- **relations:** trust ledger (above); change-layer integration contract (above)
- **verify-later:** none (a design checklist, not a built artifact)

### Companies House matching cascade (revised 7-tier signal architecture)
- **category:** companies-house-enrichment
- **status-signal:** partial
- **status-evidence:** "Current matching achieves 676/2,767 (24.4%)... Target: 70-80% automated match rate, with HITL bringing it to 90%+." Presented as a revision (v2) of an earlier plan (docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/old/022_companies_houise_matching_cascade_plan.md, a v1 outside this unit's scope), and the enrichment domain itself has a live anchor doc (docs/agent_docs/docs024_key_docs_latest/017_companies_house_enrichment.md).
- **what:** A priority-ordered cascade replacing a flat two-pass matcher, where each business flows down tiers (matched / pass-to-next / queue-for-HITL) until resolved: Tier 0 scrapes the practice's own website for a company registration number (definitive, no CH API cost); Tier 1 exact-name+geography; Tier 2 exact-name unique-in-CH regardless of geography; Tier 3 postcode+moderate-name (raised threshold from 0.40→0.50 with a mandatory name-overlap component to cut false positives); Tier 4 LLM review with the top-3 trigram candidates and full business context (not just the single best match); Tier 5 corporate-group-parent mapping for chains sharing one CH registration (Medivet, CVS, Vets4Pets, IVC Evidensia, VetPartners — addressing ~800 corporate-branch businesses that a per-business match can never resolve 1:1); Tier 6 a human-review HITL queue for the remainder. New tables proposed: `businesses.company_number_scraped`, `ch_match_candidates`, `ch_corporate_groups`.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/archive_april_26/022b_companies_house_matching_cascade_plan_v2.md
- **relations:** companies-house-enrichment (017 anchor doc); the v1 predecessor plan (022_companies_houise_matching_cascade_plan.md)
- **verify-later:** business_intel.ch_match_candidates, business_intel.ch_corporate_groups, business_intel.businesses.company_number_scraped — confirm which tiers actually shipped and current match rate

### tool-page canonicalisation misroute (adoption Features key desync)
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** "RESOLVED 2026-05-26 → b1" (root cause confirmed via query G) but the numbered "Potential solutions" section proposes the actual code fix as still to be applied — diagnosis complete, fix landing unconfirmed from the doc alone.
- **what:** A specific, resolved-diagnosis bug postmortem: an adopted `page_type='tool'` page deploys with prose describing the tool but no interactive widget. Two distinct causes were disambiguated (M1 — a widget existed and a later `SavePageSectionsAction` rebuild deleted it via a text-only regression guard blind to script-heavy content; M2 — no widget was ever generated because adoption captures text but has no JS-parse stage). Root cause for the gamesdesign.co.uk case was M2, but *not* because generation is unowned — `tool-recreation-handler` exists and should have run. The actual fault: `apply_adoption_plan` routes by `len(page.Features)`, but `buildPageFeatureMap` keys its map by the **raw** page name the adoption LLM wrote, while the routing lookup uses the **canonicalised** name (`CanonicalisePage` prepends `tool-` for tools) — so every tool page's feature lookup misses even when the LLM correctly detected interactivity, silently misrouting it to the static `page-build-handler` path instead of `tool-recreation-handler`. Games pages don't hit this because their canonical prefix (`game-`) already matches the raw key.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/main/016_debugging_guide_addendum_adopted_tools_no_widget(3).md
- **relations:** doc 029 (CanonicalisePage, Phase-0 helper); component-regeneration-flow (SavePageSectionsAction clobber path); spurious-duplicate-pages pattern (below, same "adoption vs. a second surface" family)
- **verify-later:** buildPageFeatureMap in the adoption/orchestration action code — confirm whether the canonical-key fix was applied

### spurious duplicate pages from "planner ignores adopted state"
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** The migration is a real, executed cleanup with before/after verification queries and an explicit rollback path via `_bak_del_bare` snapshot tables, dated against a specific incident (gamesdesign.co.uk, created 2026-06-03 20:25:30, cleaned up in this migration).
- **what:** A confirmed, named failure pattern: a post-adoption planner pass (`build-site-planner`/`blog-content-planner`) invents new `page_type='blog-post'` pages (`sections=[]`, `build_status='planned'`, never rendered) that duplicate content already faithfully recreated by adoption as `page_type='guide'` pages at a different URL — "a second surface invents parallel pages after adoption" because the planner doesn't check adopted state before generating its own content plan. The cleanup migration is durable — it removes the bare pages from the pages table, the *current* `site_plan_pages`/`site_plan_sections` (so the reconciler won't recreate them), and terminalises the dangling `site_work_items` rows (which have no FK to pages and would otherwise linger holding a dedup slot) — but explicitly does not fix the upstream planner logic that would reintroduce the same duplicates on a future `plan_site` run.
- **sources:** docs/_archive/agent_docs/sql_for_tables/040b_migration_cleanup_bare_guide_duplicates(1).sql
- **relations:** tool-page canonicalisation misroute (above, same "adoption vs. second surface" bug family); FOCUS_planner_ignores_adopted_state.md; doc 029; site-plan-and-reconciler
- **verify-later:** FOCUS_planner_ignores_adopted_state.md, whether the upstream planner prompt/logic has since been tightened to check adopted state first

### RAG pipeline deployment (ollama-adapter, rag_lookup/rag_index, knowledge_base)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** Live repo confirms `platform/orchestration/actions/rag_actions.go` and other RAG-related action files exist and are registered, matching the deployment bundle's file-placement manifest.
- **what:** A deployment bundle for adding retrieval-augmented generation to the chassis: a new `ollama-adapter` k8s service (own kustomize base+overlay, `ollama/ollama` image, an idempotent init container that pulls an embedding model onto a PVC — `nomic-embed-text` ~300MB, sized for a 8Gi memory limit that also leaves room for future 7B models), two new SQL migrations (`llm_call_log` with a stats view, and `knowledge_base` for RAG storage), two new registered actions (`rag_index` — chunks and embeds content into `knowledge_base`; `rag_lookup` — embeds a query and returns top-k matches), plus a nullable-helpers Go package and patches to `ai_actions.go`/`registry.go`/`anthropic.go` (adds LLM call timing/token logging and an `ollama` case to `createAIClient`).
- **sources:** docs/_archive/agent_docs/docs020_llm_training_rag/007_rag_deployment_README.md
- **relations:** model-infrastructure (009 anchor); contextkit's embed.go (same Ollama-embeddings-endpoint pattern reused independently)
- **verify-later:** platform/orchestration/actions/rag_actions.go, deployments/kustomize/services/ollama-adapter/, llm_call_log / knowledge_base tables

### isolated chat satellite architecture (three blast vectors: load/hack/bug)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Current lean (not committed — kept open)." Explicitly a plan document with an open central decision (Option X vs Y).
- **what:** A plan to run the site-chatbot's server-side pieces (turn storage, drain/analytics, and any chat workflow code) on infrastructure **separate** from the core build cluster, so that live chat traffic, a compromise of the internet-facing edge worker, or a chat-code bug cannot degrade or reach the webdesign/build system. Deliberately **not** built on the existing multi-cluster dispatch (Phase 4a, `remote-job-spawner`), which shares cluster A's Kafka/Postgres by design — the chat satellite instead reuses only the chassis *binaries* and action code, deployed against its own Kafka/DB/storage, with a one-directional async boundary (core publishes install triggers and content; nothing on the chat side has synchronous or write access back into core). Two options are weighed: **Option X (minimal, recommended MVP)** — pack-building stays on core, the satellite is just a turn store + puller + analytics, possibly needing no Kafka/chassis at all; **Option Y (full satellite chassis)** — the whole chat pipeline including install/pack-building moves to a cut-down copy of the chassis on the satellite. A worked "building-and-hosting-as-a-service via chat" example (a customer types a domain into another site's chatbot and gets a fully built, hosted site with its own chatbot) reframes the satellite as a second, customer-facing instance of the whole platform and pushes the design toward Option Y for that specific use case.
- **sources:** docs/_archive/agent_docs/docs025_ai_chatbot_idea_uk/excellent_discussions/PLAN_isolated_chat_environment(3).md
- **relations:** multicluster (Phase 4a, the pattern explicitly rejected as a template); SaaS commercial model (below, same document, §13)
- **verify-later:** whether any satellite cluster / separate chat Postgres exists; site_chat_turns table; remote-job-spawner (the Phase 4a mechanism used as contrast)

### SaaS commercial model — operator-primary, vendor-optional, with entitlement seams
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "Resolved direction" language for the model itself, but the concrete implementation ("Design the seam now... build the billing depth later") is explicitly deferred — "None of these constrain the build as long as separability and the seams above are honoured."
- **what:** A commercial model for operating the platform at scale while keeping individual sites/backends sellable: **operator-primary, vendor-optional** — operate thousands of domains directly, with the option to sell a domain plus its backend (the common case) or, rarely, the whole framework/instance. The key structural insight is that the unit of blast-radius isolation (the satellite/cluster) and the unit of sale-separability (the domain) are different granularities — operating thousands of domains does not require thousands of clusters; it requires clean per-domain partitioning (keyed on `site_id`/`domain`) plus the ability to extract one domain's artifact, data, and credentials at sale time. Five seams are flagged as cheap now / expensive to retrofit: `owner_id` on site rows; an entitlement check at both build-submission and maintenance-run (never calling Stripe directly — always through a pluggable billing-adapter interface); credential parameterization everywhere (no hardcoded keys); and a build-tier/cost-profile flag (`saas_cheap` vs `portfolio`) driving cheaper model/batching choices so low-price builds retain margin.
- **sources:** docs/_archive/agent_docs/docs025_ai_chatbot_idea_uk/excellent_discussions/PLAN_isolated_chat_environment(3).md §13
- **relations:** isolated chat satellite architecture (above, same document)
- **verify-later:** whether an owner_id/entitlement layer or billing adapter exists anywhere in the schema/codebase today

### core client→network→site→page content hierarchy (early MVP schema)
- **category:** database-and-infrastructure
- **status-signal:** superseded
- **status-evidence:** Filed as an "MVP Migration" ("Designed for patch-only updates from the start") establishing `clients`/`networks`/`sites`/`site_flows`/`pages`/`flow_pages`/`page_components` with a minimal column set (e.g. `pages` here has no `sections`, `build_status` as later became standard, no `site_plan_*` linkage) — an early snapshot of a hierarchy that has since grown substantially richer elsewhere in the live schema.
- **what:** The foundational multi-tenant hierarchy this platform is built on: `clients` (external_id linking to auth-service) → `networks` (affiliate/network-wide settings) → `sites` (domain, brand_dna, github_repo/branch) → `site_flows` (multi-track audience journeys with a narrative_arc) → `pages` (page_type, nav ordering, content_hash for change detection) → `page_components` (template instances with rendered_html, content_data, and a semantic `data_path`/`data_uuid` addressing scheme intended for future granular editing).
- **sources:** docs/_archive/agent_docs/sql_for_tables/002_links_clients_networks_etc_tables.sql
- **relations:** link registry + navigation cache (below, same migration file); system-architecture; site-plan-and-reconciler (the later, richer plan/reconciler layer this hierarchy predates)
- **verify-later:** the current `clients`/`networks`/`sites`/`pages` schema shape vs. this early version — confirm which columns/tables here are still live as originally designed vs. superseded by site_plans/site_work_items

### link registry, cached navigation structures, and redirects (link-management foundation)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** Live Go code references `link_registry` and `navigation_structures` (e.g. platform/orchestration/actions/html_actions.go, site_db_actions.go, discovery_checks/check_phantom_internal_links.go, platform/orchestration/datahelpers/links.go), and a live doc `024_link_management_v2.md` exists — confirming this MVP schema's core concept shipped and was later versioned.
- **what:** The original link-management schema: a `link_registry` table indexing every link extracted from rendered components (source component/page/site, resolved target page/site, a `scope` of internal/page/site/network/external, a `link_type` of navigation/content/semantic/affiliate/reference, plus validation status for broken-link detection); `navigation_structures` as a **cached, versioned** JSONB nav tree per site+type (header/footer/mobile/sidebar), invalidated by a trigger on any `pages` INSERT/UPDATE/DELETE and rebuilt lazily via `get_current_navigation`/`build_navigation_for_site`; and a `redirects` table (301/302/307/410, hit_count, expiry). Deliberately reuses the existing generic `relationships` table for semantic content relationships (pillar/cluster, related-content, cross-site-reference) rather than inventing a parallel structure.
- **sources:** docs/_archive/agent_docs/sql_for_tables/002_links_clients_networks_etc_tables.sql
- **relations:** core client→network→site→page hierarchy (above, same migration file); link-management (024 anchor, 024_link_management_v2.md)
- **verify-later:** 024_link_management_v2.md — confirm what changed between this v1 schema and "v2"

### workflow field-path audit query (jsonb_path_query over agent_definitions)
- **category:** development-guide
- **status-signal:** unknown
- **status-evidence:** A single ad-hoc diagnostic query with no surrounding narrative or dated claim of use.
- **what:** A small but reusable diagnostic technique: a recursive `jsonb_path_query('$.**.<key>')` sweep over every `agent_definitions.default_config->'workflow'->'steps'` row to extract every field-path value referenced by a fixed set of workflow keys (`agent_type_field`, `default_from`, `content_field`, `iterate_over`, and any `*_from`/`*_field` wildcard key) across the whole workflow corpus at once — a way to audit real field-path usage in stored workflow JSON without opening each agent definition individually.
- **sources:** docs/_archive/agent_docs/sql_for_agents/sql_for_agents_v2/001_validator_sql.sql
- **relations:** development-guide (001 anchor, "grep before using" / field-path resolution canonical-functions guidance)
- **verify-later:** none — a query technique, not a stored artifact

### docs019 working/main snapshot bundle (duplicate early-draft staging copy)
- **category:** documentation-system
- **status-signal:** superseded
- **status-evidence:** Header-by-header comparison against each doc's live counterpart shows this whole directory is a pure prefix/subset with zero unique content: `001_development_guide(3).md` (186 headers, identical structure to live `001_development_guide(5).md`); `016_debugging_guide_v2_21_.md`/`v2_36b.md` (headers are an exact ordered subset of the live `016_debugging_guide_v2_58_consolidated.md`, which continues with ~30 more sections); `026_component_regeneration_flow.md` (identical up to "Dispatch fails to pick up the rerender item," live version adds a confirmed-2026-06-22 section plus an entire second appended doc); `030_phase1_plan_and_reconciler(2).md`/`(3).md` (byte-identical to each other, headers identical to live `(5).md`); `FOCUS_imagery_assessment.md` (identical through section 8, live version continues to section 13); and `old/012d_tool_lifecycle_guide_v4.md` (byte-identical via md5 to docs/agent_docs/docs024_key_docs_latest/archive_april_26/012d_tool_lifecycle_guide_v4.md).
- **what:** This nested archive-of-archive preserves a working-copy staging snapshot of six of the platform's core numbered guides (development guide, debugging guide ×2 vintages, component-regeneration-flow, phase-1 plan/reconciler ×2 copies, imagery assessment) plus a duplicate tool-lifecycle-guide vintage, all captured mid-iteration before being superseded by later-numbered/consolidated versions that already live (and are presumably already registered) under docs024_key_docs_latest and docs014_documentation_collection. No content unique to this snapshot survives comparison against the live versions — its value is purely as a dated waypoint in each guide's version history, not as a source of new concepts.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/main/{001_development_guide(3).md,016_debugging_guide_v2_21_.md,016_debugging_guide_v2_36b.md,026_component_regeneration_flow.md,030_phase1_plan_and_reconciler(2).md,030_phase1_plan_and_reconciler(3).md,FOCUS_imagery_assessment.md}, docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/old/012d_tool_lifecycle_guide_v4.md
- **relations:** development-guide (001 anchor); debugging (016/016b anchor); imagery; site-plan-and-reconciler; tool-lifecycle (020 anchor)
- **verify-later:** none — superseded in full by already-covered live docs
