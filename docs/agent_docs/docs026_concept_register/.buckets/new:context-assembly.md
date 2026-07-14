
<!-- SOURCE: U16_docs019_design_plans.md -->
### code_symbols repo-label symmetry (shared owner/repo resolver)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** PATCH_code_symbols_shared_repo_label written; NNN_fix_lookup_repo_label_workaround header: "TEST-NOW ENABLER … apply that and REVERT this when the rebuilt image ships".
- **what:** index_code_symbols composes the code_symbols.repo label as owner/repo ("gqls/agentchassis") but lookup_code_symbols did not compose, so the diagnose workflow's lookup queried the bare name → 0 hits → empty code_results → "assemble_bundle: no scope". Structural fix: one shared resolveCodeRepoLabel used by both writer and reader so they can never drift; temporary config-only workaround hard-codes the literal until the image ships. General lesson: writer and reader of a keyed store must share the key resolver.
- **sources:** PATCH_code_symbols_shared_repo_label.md; NNN_fix_lookup_repo_label_workaround.sql; NNN_create_code_indexer_agent(2).sql (label convention)
- **relations:** code_symbols index; loop-back plumbing fault class
- **verify-later:** code_symbols_actions.go resolveCodeRepoLabel; whether the workaround REVERT ran

<!-- SOURCE: U16_docs019_design_plans.md -->
### analyse_repo_local in-process analysis + the stale-index incident
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** NNN_swap_indexer_to_local_analysis header documents the incident and the applied swap; PATCH_lift_fetcher_and_register gives the registration; README_flows/handoffs treat the local path as current.
- **what:** The adapter round-trip left the diagnose-agent with no local checkout (ReadSymbolBody could not slice bodies) and once resolved ref "HEAD" to a year-old commit, silently indexing a July-2025 tree (69 files/436 symbols vs 572 today). analyse_repo_local fetches the repo tarball at an explicit ref to a local temp dir and analyses in-process (shared internal/reposource fetcher), with pin_to_index_commit defaulting true for the diagnose loop (bodies match the index) and set false for the indexer (the indexer defines the commit). Corollary rules: explicit git refs never HEAD; the spawned pod needs GITHUB_READ_TOKEN via the isRepoCloningAgent spawn gate.
- **sources:** NNN_swap_analyse_repo_to_local.sql; NNN_swap_indexer_to_local_analysis.sql; PATCH_lift_fetcher_and_register.md; TRIGGER_code_indexer_v2(1).sh
- **relations:** analyser adapter (request_repo_analysis stays for the code-indexer); index freshness / CI-triggered indexing
- **verify-later:** internal/reposource/github_source.go; analyse_repo_local_action.go; registry entry

<!-- SOURCE: U16_docs019_design_plans.md -->
### "Documentation is code" — the context-assembly tool and paid service
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** PLAN_context_assembly_tool_and_service(2) status: "A thin slice of Phase 1 is built and exercised on real code … first real run: a 30 KB bundle vs the script's 1.7 MB"; Phases 2–4 and the service unbuilt.
- **what:** For any development task, assemble a task-scoped context bundle from documentation + codebase and feed generation against ground-truth verification, so results are "more likely to be correct" than pasting code into a chat. The thesis: in an AI-driven workflow documentation (standards, intent, trajectory) is an operational input — version it, drift-detect it, compose it deterministically. Two audiences: dogfood on the chassis repo first, then a paid multi-tenant service behind the gateway. Design principles: engine/config split (tenant-agnostic engine + per-stack adapters, the decision that makes the service possible); seams for the optimal machinery (cascade router, decision-point checkers, mediator) defined as interfaces from v1; dogfood first. Phases: 0 contracts+constitution, 1 bundle builder MVP, 2 verification loop, 3 service (sandboxed verification is gating), 4 cascade/checkers/mediator.
- **sources:** PLAN_context_assembly_tool_and_service(2).md; 001_onboarding_discussion.txt; MAPPING_tool_to_actions_and_agents(2).md
- **relations:** bundle shape contract; six governance contracts (Phase 0); onboarding as the hard service problem; thin-slice-first principle
- **verify-later:** contextkit module state vs the plan's thin-slice claims; gateway project status

<!-- SOURCE: U16_docs019_design_plans.md -->
### Bundle shape contract (the task-scoped context package)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) "Status: contract specification"; flagged by FOCUS_whole_plan_review §1.1 as "the next load-bearing contract".
- **what:** A bundle is assembled fresh per task (never stored/reused, so never stale) with fixed sections: metadata; task/target; the authored layer (constitution, why-chain, priority profile, direction-of-travel, matched standards); code context (in-scope code in full, neighbourhood signatures, reuse-search results, schema, definition data); database data in three kinds; pointers to everything not inlined; and provenance (exactly what went in, logged as the decision log's inputs_used). Exists in a canonical structured form and a rendered text view. Two integrity rules from the edge-case pass: assemble from a consistent snapshot (no torn reads), and log what the generator SAW (rendered form), not what was assembled.
- **sources:** PLAN_bundle_shape_contract(2).md; FOCUS_whole_plan_review.md#1.1; FOCUS_pre_build_edge_cases(1).md#1.4,#4.3
- **relations:** decision log inputs_used; altitude/step-type; multipass fetch; contextkit assembler (the harness prototype)
- **verify-later:** whether any structured bundle object exists in code vs the markdown-emitting harness

<!-- SOURCE: U16_docs019_design_plans.md -->
### Three kinds of database data (definition / operational / content)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.1, verified against the live schema ("workflows are jsonb columns on agent_definitions … there is no workflows table and no tools table").
- **what:** In a data-defined system much of "the code" is DB rows, so the bundle distinguishes: definition data (the system's design as data — workflows in agent_definitions, tools as content_components rows, prompts as text columns; fetched routinely, covered by reuse-search); operational data (telemetry — work items, orchestration_states, error logs; multipass-capped); and content data (the output — sites/pages/tenant data; the gated set where privacy matters in the service).
- **sources:** PLAN_bundle_shape_contract(2).md#2.1
- **relations:** multipass fetch; reuse search over definitions; sensitivity gates
- **verify-later:** n/a (contract)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Multipass fetch: probe → gate → include/reduce/point
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** dbcontext "row data with multipass sizing" built in the harness (tool plan status); the full gate flow is contract-only.
- **what:** Query results have unknown size until run, so the builder probes with LIMIT N+1 (not count(*) — counting a filtered query costs as much as running it), checks a size gate and a sensitivity gate, then includes rows in full, reduces (aggregate / representative sample / pointer), or gates behind confirmation. An oversized result becomes an aggregate or pointer, never an unbounded dump.
- **sources:** PLAN_bundle_shape_contract(2).md#3; FOCUS_pre_build_edge_cases(1).md#4.2; GUIDE_deploy_from_context_packs(1).md (dbcontext)
- **relations:** three kinds of DB data; bounded bundle egress
- **verify-later:** dbcontext.go sizing logic

<!-- SOURCE: U16_docs019_design_plans.md -->
### Runtime evidence keyed by orchestration_id (the run narrative)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** dbcontext -runtime-site built and used throughout (bundles carried agent_error_log/site_work_items); the fuller run-trace composition is contract-level.
- **what:** For debugging, the most useful context is the narrative of a run, reconstructable from one key: orchestration_id (+ correlation_id) spans orchestration_states (spawn tree, topics, status), llm_call_log (time-ordered step sequence), agent_error_log (error trail), pod logs (grep by run id) and the Kafka messages. Three cheap reads give a coherent single-run story instead of a scatter of lines. Log-correlation only works where the id is actually in the log line — a convention whose coverage the conventions agent audits.
- **sources:** PLAN_bundle_shape_contract(2).md#2.2; PLAN_onboarding_agent_specs(6).md#2.9,#1.9; README_02 ("everything durable, one correlation id")
- **relations:** run signatures; codebase-conditional capabilities; diagnose_load_runtime
- **verify-later:** whether orchestration_id reliably appears in pod log lines (named open item)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Run signatures: expected-vs-actual sequence diff
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.2 — designed; capture/storage named open in §9.
- **what:** Capture a healthy run's step sequence and spawn-tree shape once (from known-good runs, confirmed), store as authored reference, and on a debug task diff the actual run against it to surface the divergence point — "matched the healthy path to step 7, then diverged here" instead of "read the logs". Verification applied to runtime.
- **sources:** PLAN_bundle_shape_contract(2).md#2.2,#9
- **relations:** runtime evidence by orchestration_id; diagnostic playbooks
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Diagnostic playbooks / failure fingerprints as authored knowledge
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.2: "seeded from those guides rather than authored fresh"; home (standards atoms vs sibling table) open.
- **what:** Known failure fingerprints — a failure's signature + the commands that confirm it + the fix pattern — curated from the existing debugging guides and failure writeups, surfaced into debug bundles the way standards are surfaced into build bundles, and grown as run-signature diffs reveal new ones.
- **sources:** PLAN_bundle_shape_contract(2).md#2.2,#9
- **relations:** run signatures; debugging guide 016 (the seed corpus)
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Codebase-conditional capabilities (degrade, don't break; partial config is normal)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.3/§2.4 and agent-specs §2.9 — design rules, engine unbuilt.
- **what:** The definition-data / run-trace / run-signature / log-correlation capabilities rest on structural facts (behaviour stored as data, a run-correlation key, named logged steps, a known log fetch) that hold on our codebase but may not elsewhere. Stack-discovery records which facts hold; each capability degrades to a weaker form or states "unavailable, because this codebase has no X" instead of breaking. Companion rule: distinguish not-yet-authored config (degrade gracefully, note what's pending) from malformed config (fail loud) — the no-fallbacks rule applies to malformed data only.
- **sources:** PLAN_bundle_shape_contract(2).md#2.3-2.4; PLAN_onboarding_agent_specs(6).md#2.9; FOCUS_pre_build_edge_cases(1).md#2.3; FOCUS_whole_plan_review.md#2.5
- **relations:** stack-discovery agent; convention coverage = capability reliability
- **verify-later:** n/a (design)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Altitude: step type decides what the bundle emphasises
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** The harness assembler takes `-step framing|debug` (used in real bundle invocations, e.g. "-step debug", "the framing bundle that produced this plan used -step framing").
- **what:** The same task at different stages needs different context: framing/decision steps get full intent (why-chain, priority profile, direction-of-travel) with light code; implementation gets full code with a thin intent tether; debug leads with errors + runtime evidence + the expected-vs-actual diff. "Right altitude at the right moment" made concrete in the bundle composer.
- **sources:** PLAN_bundle_shape_contract(2).md#4; PLAN_imagery_sprite_sheet.md (framing-bundle use); tasks/gameslink_missing_index_rerender/RUNBOOK…(2).md (-step debug)
- **relations:** bundle shape contract; salience-loss problem
- **verify-later:** assembler.go step handling

<!-- SOURCE: U16_docs019_design_plans.md -->
### Go analyser + call-graph neighbourhood (and the wiring-include gap)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Tool plan status: analyser/assembler "built and exercised on real code"; internal/analysis "already exists in the repo" per migration(19); bundles record "533 files analysed".
- **what:** A go/ast walk producing a structured index (signatures, types, per-function `calls` by name-based matching — full go/types resolution deliberately avoided as too heavy/fragile); the assembler slices in-scope symbols in full plus a signature-level caller/callee neighbourhood. Known structural blind spot: registration/init wiring (registry.go) is unreachable via calls, so `-include` exists for wiring files — the same gap as manually-named docs. ReadSymbolBody (span-slice over start_line/end_line) was extracted, tested byte-identical to cmd/assembler, and shared with diagnose_assemble_bundle.
- **sources:** 001_claude_reasoning; PLAN_context_assembly_tool_and_service(2).md#5 status; GUIDE_deploy_from_context_packs(1).md; PLAN.md changelog (ReadSymbolBody)
- **relations:** analyser adapter (wraps the same library); evidence-follows re-scoping; broad-script-vs-lean-assembler tradeoff
- **verify-later:** internal/analysis (chassis) vs contextkit copy drift (flagged in PLAN.md)

<!-- SOURCE: U16_docs019_design_plans.md -->
### contextkit module packaging and the graduation seam
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Tool plan status: "The tools are now a single Go module, contextkit/ … the two contracts defined once — internal/analysis and internal/candidates"; production-relevant parts graduated in-cluster per migration(19).
- **what:** All harness tools (analyser, embed, resolve_targets, fuse, eval_targets, assembler, dbcontext, bundle, dedup, diagnose) live in one module with two shared contracts; graduation moves the internal packages under the chassis module path and turns command mains into actions without changing the contracts. Production runs in-cluster; the harness remains the dev/measurement scaffold (eval_targets stays offline, the flywheel's measurement tool). The trial's output is throwaway; the rule it teaches is durable.
- **sources:** PLAN_context_assembly_tool_and_service(2).md status; PLAN_workflows_and_actions_migration(19).md (analyser-adapter sections); MAPPING_tool_to_actions_and_agents(2).md
- **relations:** analyser adapter; one-decision-core two realisations (same seam idea)
- **verify-later:** go_files/contextkit module contents (unit U17 territory)

<!-- SOURCE: U16_docs019_design_plans.md -->
### code_symbols: the per-repo code index (pgvector sibling table)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19) changelog 2026-06-09: "code_symbols applied cleanly (table + 4 indexes)"; later populated (436 symbols, embeddings 436/436 per the repo-label workaround header).
- **what:** A sibling to knowledge_base reusing its proven shape (vector(768) nomic + trigram on content + idempotent dedup, same AIService embedder with caller-applied nomic prefixes) but keyed for code: unique(repo,path,symbol), SHA-versioned via commit_sha, identity upsert (ON CONFLICT DO UPDATE WHERE content_hash IS DISTINCT — a symbol persists across commits, re-embeds only on change), pruned hard on re-index. Deliberate departures from chassis conventions (no version/previous_version_id, no deleted_at) because it is a rebuildable cache. HNSW chosen over the KB's IVFFlat for incremental churn (pgvector 0.8.0 confirmed both). One symbol = one row; no prose chunker (rag_index's character windows fragment Go mid-function).
- **sources:** NNN_create_code_symbols_index.sql; PLAN_workflows_and_actions_migration(19).md A5/B4/B4b + code-indexer reuse mapping
- **relations:** lookup/index_code_symbols; embedding policy split; knowledge_base reuse-not-copy
- **verify-later:** \d code_symbols; row counts per repo; embedding_model column use

<!-- SOURCE: U16_docs019_design_plans.md -->
### Hybrid code retrieval: index/lookup_code_symbols (vector + trigram, RRF in SQL)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19): actions built against real rag_actions.go, registered, deployed with the analyser (2026-06-12); lookup live in the diagnose workflow.
- **what:** index_code_symbols flattens analyser Output to symbol rows, skips unchanged content hashes, embeds non-fatally (trigram still works without an embedding), upserts and prunes; lookup_code_symbols mirrors rag_lookup (embed query → cosine vector search → trigram fallback → top-k + code_context). The hybrid RRF fusion moved into SQL, so the in-Go fuse tool never graduated. Deliberately a sibling action, not a parameterised rag_lookup (KB columns are baked into vectorSearchKB); the three embedding helpers are shared package-level functions.
- **sources:** PLAN_workflows_and_actions_migration(19).md (code-indexer reuse mapping + consumer-side-built entries); NNN_create_code_symbols_index.sql (query set)
- **relations:** code_symbols table; rag_lookup/rag_index (the mechanism source); repo-label symmetry
- **verify-later:** code_symbols_actions.go; registry entries (storage/code categories)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Analyser adapter: in-cluster polyglot parsing service
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19) changelog 2026-06-12: "ANALYSER DEPLOYED TO PRODUCTION (uk_001) … the code-indexer agent row is applied (INSERT 0 1)."
- **what:** A Kafka message-based adapter (mirroring git/thunder, not HTTP) consuming `analyse` requests (owner, repo, ref, language), fetching source read-only via the GitHub tarball endpoint (one request, recovers the exact commit_sha, path-traversal-guarded), parsing through the analysis.Analyse library behind an `Analyser` seam so per-language parsers (JS next) drop in, and replying on the caller's responses topic. Security: its own least-privilege repo-scoped read-only token as a k8s Secret mounted only on this pod — two narrow credentials (analyser read, git-adapter write), never one broad token. Built polyglot-ready NOW because the JS tools already exist (tech debt, not future planning).
- **sources:** PLAN_workflows_and_actions_migration(19).md (analyser adapter sections + repo access & security); FOCUS_js_tools_documentation.md
- **relations:** analyse_repo_local (the in-process alternative that later took the diagnose+indexer paths); adapter envelope contract
- **verify-later:** internal/adapters/analyser; analyser-adapter deployment in uk_001; whether request_repo_analysis still has callers

<!-- SOURCE: U16_docs019_design_plans.md -->
### Text-vs-code embeddings: share the mechanism, separate the policy (B4b)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19) B4b "decided 2026-06-09"; schema embodies it (sibling table + embedding_model column).
- **what:** Shared and not duplicated: the AIService embedder seam, provider implementations, pgvector, and the semantic+trigram hybrid pattern. Separate per domain so each upgrades independently: the model (prose vs code-specific), dimension, preprocessing (nomic search_ prefixes are caller-side), retrieval tuning (HNSW vs IVFFlat, lexical-heavier for code), row definition, and evaluation. Turns B4a into "which model for code", measurable independently. Caution recorded: separation pays only if the mechanism stays shared.
- **sources:** PLAN_workflows_and_actions_migration(19).md B4a/B4b/A5 resolutions
- **relations:** code_symbols; B4a ceiling; CPU-Ollama feasibility (bulk-index speed, code-domain recall)
- **verify-later:** embedding_model column values; whether a code-specific model was ever adopted

<!-- SOURCE: U16_docs019_design_plans.md -->
### Code-indexer agent, index-orchestrator wrapper, and CI-triggered indexing
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** code-indexer row applied (2026-06-12); index-orchestrator seeded (v2 after the v1 `name`-column failure, 2026-07-02); CI trigger still a queue item ("CI-triggered indexing (self-contained)" in HANDOFF_builder_thread).
- **what:** The code-indexer is a thin orchestrator (analyse → index_code_symbols → complete). Run 6dfa37cd proved orchestrate+agent_type is adopted IN-PLACE on the shared chassis pod — which never holds GITHUB_READ_TOKEN — so the index-orchestrator wraps it in the proven spawn pattern so the spawned pod receives the secret (isRepoCloningAgent gate). Manual trigger TRIGGER_code_indexer_v2.sh sends the explicit branch ref; the planned durable form is a GitHub Actions step firing the envelope with GITHUB_SHA on push, retiring the index-staleness class for the diagnosis corpus.
- **sources:** NNN_create_code_indexer_agent(2).sql; NNN_seed_index_orchestrator(1).sql; TRIGGER_code_indexer_v2(1).sh; HANDOFF_builder_thread.md#3
- **relations:** analyse_repo_local staleness incident; spawn-consumed columns lesson; reuse-index freshness (governance)
- **verify-later:** index-orchestrator row; CI workflow file existence

<!-- SOURCE: U16_docs019_design_plans.md -->
### Documentation indexing rides the prose rag path
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** Migration(19) "Documentation indexing (related, lighter…)" — direction recorded, not reported built.
- **what:** Docs are prose, so they fit rag_index/knowledge_base as-is: index each guide under a collection (e.g. `standards`), retrieve with rag_lookup; flat files in git remain the editable source of truth, the DB copy a derived rebuildable index so the assembler pulls relevant sections instead of pasting 124KB guides. Precondition: docs must live in a versioned repo. Separate, smaller workstream from code_symbols.
- **sources:** PLAN_workflows_and_actions_migration(19).md (documentation indexing section); FOCUS_js_tools_documentation.md
- **relations:** JS tools documentation gap; standards/docs agent (the matched-guidelines slot)
- **verify-later:** knowledge_base collections for docs

<!-- SOURCE: U16_docs019_design_plans.md -->
### cmd/bundle robustness contract (validate early, fail loud, manifest input)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** 003_contextkit_bundle_issues (2026-06-24) — two real failed runs analysed; HANDOFF_fixloop(8) notes the skip-message/usage patches exist but "needs gofmt + build".
- **what:** Field-found tool defects and the rules they teach: validate the cheapest input (-analysis JSON) BEFORE the slow psql-shelling gather phases, with an actionable message naming file/size/regeneration; accept a manifest/config file instead of 20-line backslash shell commands (kills the unquoted-parentheses class — real filenames contain "(1)"); a missing -doc/-scope path must fail loudly naming the path, because a silently-omitted file means a downstream chat reasons from incomplete context without knowing; single quoted -psql argument, no TTY.
- **sources:** 003_contextkit_bundle_issues.md; HANDOFF_fixloop_thread(8).md#2
- **relations:** bundle-first handoff practice; fail-loud-vs-degrade rule
- **verify-later:** cmd/bundle/main.go precondition ordering; -config support

<!-- SOURCE: U16_docs019_design_plans.md -->
### Reuse search before generation (code AND definition rows)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** Named in the tool plan §2 and bundle contract §2.1/§8; only the retrieval substrate (code_symbols, pgvector) exists — the pre-generation reuse step is unbuilt.
- **what:** Before generating, run "what already does something like this" over existing functions/structs and — critically in a data-defined system — over definition rows (workflows/agents/tools), so reuse-before-recreate is mechanical rather than a remembered habit and near-copies of existing workflows are caught like duplicate functions. Needs a searchable text projection for jsonb definitions (named open). The index is derived state that goes stale silently — re-index on change events and stamp freshness.
- **sources:** PLAN_context_assembly_tool_and_service(2).md#2; PLAN_bundle_shape_contract(2).md#2.1,#8; FOCUS_pre_build_edge_cases(1).md#2.4,#15
- **relations:** reuse_index_refresh trigger; code_symbols; dev-guide reuse discipline
- **verify-later:** any reuse-search action; definition-row indexing

<!-- SOURCE: U16_docs019_design_plans.md -->
### DB capabilities capture (\dx/\df into the bundle)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** Migration(19) B2a "decided, mechanism built" in the harness (dbcontext -capabilities, assembler -dbfacts); the durable indexing-workflow half is future.
- **what:** Generation that writes SQL should see installed extensions and helper functions (so it knows pgvector exists and reuses snapshot_agent instead of hand-rolling a backup — the migration-110 footgun). Captured as DB context (not the analyser's job), included for DB-touching tasks with a reuse nudge; the durable plan folds capture into the indexing workflow on a migration cadence so bundles always carry current DB facts without anyone remembering a flag.
- **sources:** PLAN_workflows_and_actions_migration(19).md B2a + 2026-06-09 changelog
- **relations:** multipass fetch; schema-before-SQL discipline
- **verify-later:** dbcontext -capabilities flag; assembler dbfacts section

<!-- SOURCE: U16_docs019_design_plans.md -->
### code_symbols repo-label symmetry (shared owner/repo resolver)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** PATCH_code_symbols_shared_repo_label written; NNN_fix_lookup_repo_label_workaround header: "TEST-NOW ENABLER … apply that and REVERT this when the rebuilt image ships".
- **what:** index_code_symbols composes the code_symbols.repo label as owner/repo ("gqls/agentchassis") but lookup_code_symbols did not compose, so the diagnose workflow's lookup queried the bare name → 0 hits → empty code_results → "assemble_bundle: no scope". Structural fix: one shared resolveCodeRepoLabel used by both writer and reader so they can never drift; temporary config-only workaround hard-codes the literal until the image ships. General lesson: writer and reader of a keyed store must share the key resolver.
- **sources:** PATCH_code_symbols_shared_repo_label.md; NNN_fix_lookup_repo_label_workaround.sql; NNN_create_code_indexer_agent(2).sql (label convention)
- **relations:** code_symbols index; loop-back plumbing fault class
- **verify-later:** code_symbols_actions.go resolveCodeRepoLabel; whether the workaround REVERT ran

<!-- SOURCE: U16_docs019_design_plans.md -->
### analyse_repo_local in-process analysis + the stale-index incident
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** NNN_swap_indexer_to_local_analysis header documents the incident and the applied swap; PATCH_lift_fetcher_and_register gives the registration; README_flows/handoffs treat the local path as current.
- **what:** The adapter round-trip left the diagnose-agent with no local checkout (ReadSymbolBody could not slice bodies) and once resolved ref "HEAD" to a year-old commit, silently indexing a July-2025 tree (69 files/436 symbols vs 572 today). analyse_repo_local fetches the repo tarball at an explicit ref to a local temp dir and analyses in-process (shared internal/reposource fetcher), with pin_to_index_commit defaulting true for the diagnose loop (bodies match the index) and set false for the indexer (the indexer defines the commit). Corollary rules: explicit git refs never HEAD; the spawned pod needs GITHUB_READ_TOKEN via the isRepoCloningAgent spawn gate.
- **sources:** NNN_swap_analyse_repo_to_local.sql; NNN_swap_indexer_to_local_analysis.sql; PATCH_lift_fetcher_and_register.md; TRIGGER_code_indexer_v2(1).sh
- **relations:** analyser adapter (request_repo_analysis stays for the code-indexer); index freshness / CI-triggered indexing
- **verify-later:** internal/reposource/github_source.go; analyse_repo_local_action.go; registry entry

<!-- SOURCE: U16_docs019_design_plans.md -->
### "Documentation is code" — the context-assembly tool and paid service
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** PLAN_context_assembly_tool_and_service(2) status: "A thin slice of Phase 1 is built and exercised on real code … first real run: a 30 KB bundle vs the script's 1.7 MB"; Phases 2–4 and the service unbuilt.
- **what:** For any development task, assemble a task-scoped context bundle from documentation + codebase and feed generation against ground-truth verification, so results are "more likely to be correct" than pasting code into a chat. The thesis: in an AI-driven workflow documentation (standards, intent, trajectory) is an operational input — version it, drift-detect it, compose it deterministically. Two audiences: dogfood on the chassis repo first, then a paid multi-tenant service behind the gateway. Design principles: engine/config split (tenant-agnostic engine + per-stack adapters, the decision that makes the service possible); seams for the optimal machinery (cascade router, decision-point checkers, mediator) defined as interfaces from v1; dogfood first. Phases: 0 contracts+constitution, 1 bundle builder MVP, 2 verification loop, 3 service (sandboxed verification is gating), 4 cascade/checkers/mediator.
- **sources:** PLAN_context_assembly_tool_and_service(2).md; 001_onboarding_discussion.txt; MAPPING_tool_to_actions_and_agents(2).md
- **relations:** bundle shape contract; six governance contracts (Phase 0); onboarding as the hard service problem; thin-slice-first principle
- **verify-later:** contextkit module state vs the plan's thin-slice claims; gateway project status

<!-- SOURCE: U16_docs019_design_plans.md -->
### Bundle shape contract (the task-scoped context package)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) "Status: contract specification"; flagged by FOCUS_whole_plan_review §1.1 as "the next load-bearing contract".
- **what:** A bundle is assembled fresh per task (never stored/reused, so never stale) with fixed sections: metadata; task/target; the authored layer (constitution, why-chain, priority profile, direction-of-travel, matched standards); code context (in-scope code in full, neighbourhood signatures, reuse-search results, schema, definition data); database data in three kinds; pointers to everything not inlined; and provenance (exactly what went in, logged as the decision log's inputs_used). Exists in a canonical structured form and a rendered text view. Two integrity rules from the edge-case pass: assemble from a consistent snapshot (no torn reads), and log what the generator SAW (rendered form), not what was assembled.
- **sources:** PLAN_bundle_shape_contract(2).md; FOCUS_whole_plan_review.md#1.1; FOCUS_pre_build_edge_cases(1).md#1.4,#4.3
- **relations:** decision log inputs_used; altitude/step-type; multipass fetch; contextkit assembler (the harness prototype)
- **verify-later:** whether any structured bundle object exists in code vs the markdown-emitting harness

<!-- SOURCE: U16_docs019_design_plans.md -->
### Three kinds of database data (definition / operational / content)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.1, verified against the live schema ("workflows are jsonb columns on agent_definitions … there is no workflows table and no tools table").
- **what:** In a data-defined system much of "the code" is DB rows, so the bundle distinguishes: definition data (the system's design as data — workflows in agent_definitions, tools as content_components rows, prompts as text columns; fetched routinely, covered by reuse-search); operational data (telemetry — work items, orchestration_states, error logs; multipass-capped); and content data (the output — sites/pages/tenant data; the gated set where privacy matters in the service).
- **sources:** PLAN_bundle_shape_contract(2).md#2.1
- **relations:** multipass fetch; reuse search over definitions; sensitivity gates
- **verify-later:** n/a (contract)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Multipass fetch: probe → gate → include/reduce/point
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** dbcontext "row data with multipass sizing" built in the harness (tool plan status); the full gate flow is contract-only.
- **what:** Query results have unknown size until run, so the builder probes with LIMIT N+1 (not count(*) — counting a filtered query costs as much as running it), checks a size gate and a sensitivity gate, then includes rows in full, reduces (aggregate / representative sample / pointer), or gates behind confirmation. An oversized result becomes an aggregate or pointer, never an unbounded dump.
- **sources:** PLAN_bundle_shape_contract(2).md#3; FOCUS_pre_build_edge_cases(1).md#4.2; GUIDE_deploy_from_context_packs(1).md (dbcontext)
- **relations:** three kinds of DB data; bounded bundle egress
- **verify-later:** dbcontext.go sizing logic

<!-- SOURCE: U16_docs019_design_plans.md -->
### Runtime evidence keyed by orchestration_id (the run narrative)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** dbcontext -runtime-site built and used throughout (bundles carried agent_error_log/site_work_items); the fuller run-trace composition is contract-level.
- **what:** For debugging, the most useful context is the narrative of a run, reconstructable from one key: orchestration_id (+ correlation_id) spans orchestration_states (spawn tree, topics, status), llm_call_log (time-ordered step sequence), agent_error_log (error trail), pod logs (grep by run id) and the Kafka messages. Three cheap reads give a coherent single-run story instead of a scatter of lines. Log-correlation only works where the id is actually in the log line — a convention whose coverage the conventions agent audits.
- **sources:** PLAN_bundle_shape_contract(2).md#2.2; PLAN_onboarding_agent_specs(6).md#2.9,#1.9; README_02 ("everything durable, one correlation id")
- **relations:** run signatures; codebase-conditional capabilities; diagnose_load_runtime
- **verify-later:** whether orchestration_id reliably appears in pod log lines (named open item)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Run signatures: expected-vs-actual sequence diff
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.2 — designed; capture/storage named open in §9.
- **what:** Capture a healthy run's step sequence and spawn-tree shape once (from known-good runs, confirmed), store as authored reference, and on a debug task diff the actual run against it to surface the divergence point — "matched the healthy path to step 7, then diverged here" instead of "read the logs". Verification applied to runtime.
- **sources:** PLAN_bundle_shape_contract(2).md#2.2,#9
- **relations:** runtime evidence by orchestration_id; diagnostic playbooks
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Diagnostic playbooks / failure fingerprints as authored knowledge
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.2: "seeded from those guides rather than authored fresh"; home (standards atoms vs sibling table) open.
- **what:** Known failure fingerprints — a failure's signature + the commands that confirm it + the fix pattern — curated from the existing debugging guides and failure writeups, surfaced into debug bundles the way standards are surfaced into build bundles, and grown as run-signature diffs reveal new ones.
- **sources:** PLAN_bundle_shape_contract(2).md#2.2,#9
- **relations:** run signatures; debugging guide 016 (the seed corpus)
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Codebase-conditional capabilities (degrade, don't break; partial config is normal)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.3/§2.4 and agent-specs §2.9 — design rules, engine unbuilt.
- **what:** The definition-data / run-trace / run-signature / log-correlation capabilities rest on structural facts (behaviour stored as data, a run-correlation key, named logged steps, a known log fetch) that hold on our codebase but may not elsewhere. Stack-discovery records which facts hold; each capability degrades to a weaker form or states "unavailable, because this codebase has no X" instead of breaking. Companion rule: distinguish not-yet-authored config (degrade gracefully, note what's pending) from malformed config (fail loud) — the no-fallbacks rule applies to malformed data only.
- **sources:** PLAN_bundle_shape_contract(2).md#2.3-2.4; PLAN_onboarding_agent_specs(6).md#2.9; FOCUS_pre_build_edge_cases(1).md#2.3; FOCUS_whole_plan_review.md#2.5
- **relations:** stack-discovery agent; convention coverage = capability reliability
- **verify-later:** n/a (design)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Altitude: step type decides what the bundle emphasises
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** The harness assembler takes `-step framing|debug` (used in real bundle invocations, e.g. "-step debug", "the framing bundle that produced this plan used -step framing").
- **what:** The same task at different stages needs different context: framing/decision steps get full intent (why-chain, priority profile, direction-of-travel) with light code; implementation gets full code with a thin intent tether; debug leads with errors + runtime evidence + the expected-vs-actual diff. "Right altitude at the right moment" made concrete in the bundle composer.
- **sources:** PLAN_bundle_shape_contract(2).md#4; PLAN_imagery_sprite_sheet.md (framing-bundle use); tasks/gameslink_missing_index_rerender/RUNBOOK…(2).md (-step debug)
- **relations:** bundle shape contract; salience-loss problem
- **verify-later:** assembler.go step handling

<!-- SOURCE: U16_docs019_design_plans.md -->
### Go analyser + call-graph neighbourhood (and the wiring-include gap)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Tool plan status: analyser/assembler "built and exercised on real code"; internal/analysis "already exists in the repo" per migration(19); bundles record "533 files analysed".
- **what:** A go/ast walk producing a structured index (signatures, types, per-function `calls` by name-based matching — full go/types resolution deliberately avoided as too heavy/fragile); the assembler slices in-scope symbols in full plus a signature-level caller/callee neighbourhood. Known structural blind spot: registration/init wiring (registry.go) is unreachable via calls, so `-include` exists for wiring files — the same gap as manually-named docs. ReadSymbolBody (span-slice over start_line/end_line) was extracted, tested byte-identical to cmd/assembler, and shared with diagnose_assemble_bundle.
- **sources:** 001_claude_reasoning; PLAN_context_assembly_tool_and_service(2).md#5 status; GUIDE_deploy_from_context_packs(1).md; PLAN.md changelog (ReadSymbolBody)
- **relations:** analyser adapter (wraps the same library); evidence-follows re-scoping; broad-script-vs-lean-assembler tradeoff
- **verify-later:** internal/analysis (chassis) vs contextkit copy drift (flagged in PLAN.md)

<!-- SOURCE: U16_docs019_design_plans.md -->
### contextkit module packaging and the graduation seam
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Tool plan status: "The tools are now a single Go module, contextkit/ … the two contracts defined once — internal/analysis and internal/candidates"; production-relevant parts graduated in-cluster per migration(19).
- **what:** All harness tools (analyser, embed, resolve_targets, fuse, eval_targets, assembler, dbcontext, bundle, dedup, diagnose) live in one module with two shared contracts; graduation moves the internal packages under the chassis module path and turns command mains into actions without changing the contracts. Production runs in-cluster; the harness remains the dev/measurement scaffold (eval_targets stays offline, the flywheel's measurement tool). The trial's output is throwaway; the rule it teaches is durable.
- **sources:** PLAN_context_assembly_tool_and_service(2).md status; PLAN_workflows_and_actions_migration(19).md (analyser-adapter sections); MAPPING_tool_to_actions_and_agents(2).md
- **relations:** analyser adapter; one-decision-core two realisations (same seam idea)
- **verify-later:** go_files/contextkit module contents (unit U17 territory)

<!-- SOURCE: U16_docs019_design_plans.md -->
### code_symbols: the per-repo code index (pgvector sibling table)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19) changelog 2026-06-09: "code_symbols applied cleanly (table + 4 indexes)"; later populated (436 symbols, embeddings 436/436 per the repo-label workaround header).
- **what:** A sibling to knowledge_base reusing its proven shape (vector(768) nomic + trigram on content + idempotent dedup, same AIService embedder with caller-applied nomic prefixes) but keyed for code: unique(repo,path,symbol), SHA-versioned via commit_sha, identity upsert (ON CONFLICT DO UPDATE WHERE content_hash IS DISTINCT — a symbol persists across commits, re-embeds only on change), pruned hard on re-index. Deliberate departures from chassis conventions (no version/previous_version_id, no deleted_at) because it is a rebuildable cache. HNSW chosen over the KB's IVFFlat for incremental churn (pgvector 0.8.0 confirmed both). One symbol = one row; no prose chunker (rag_index's character windows fragment Go mid-function).
- **sources:** NNN_create_code_symbols_index.sql; PLAN_workflows_and_actions_migration(19).md A5/B4/B4b + code-indexer reuse mapping
- **relations:** lookup/index_code_symbols; embedding policy split; knowledge_base reuse-not-copy
- **verify-later:** \d code_symbols; row counts per repo; embedding_model column use

<!-- SOURCE: U16_docs019_design_plans.md -->
### Hybrid code retrieval: index/lookup_code_symbols (vector + trigram, RRF in SQL)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19): actions built against real rag_actions.go, registered, deployed with the analyser (2026-06-12); lookup live in the diagnose workflow.
- **what:** index_code_symbols flattens analyser Output to symbol rows, skips unchanged content hashes, embeds non-fatally (trigram still works without an embedding), upserts and prunes; lookup_code_symbols mirrors rag_lookup (embed query → cosine vector search → trigram fallback → top-k + code_context). The hybrid RRF fusion moved into SQL, so the in-Go fuse tool never graduated. Deliberately a sibling action, not a parameterised rag_lookup (KB columns are baked into vectorSearchKB); the three embedding helpers are shared package-level functions.
- **sources:** PLAN_workflows_and_actions_migration(19).md (code-indexer reuse mapping + consumer-side-built entries); NNN_create_code_symbols_index.sql (query set)
- **relations:** code_symbols table; rag_lookup/rag_index (the mechanism source); repo-label symmetry
- **verify-later:** code_symbols_actions.go; registry entries (storage/code categories)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Analyser adapter: in-cluster polyglot parsing service
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19) changelog 2026-06-12: "ANALYSER DEPLOYED TO PRODUCTION (uk_001) … the code-indexer agent row is applied (INSERT 0 1)."
- **what:** A Kafka message-based adapter (mirroring git/thunder, not HTTP) consuming `analyse` requests (owner, repo, ref, language), fetching source read-only via the GitHub tarball endpoint (one request, recovers the exact commit_sha, path-traversal-guarded), parsing through the analysis.Analyse library behind an `Analyser` seam so per-language parsers (JS next) drop in, and replying on the caller's responses topic. Security: its own least-privilege repo-scoped read-only token as a k8s Secret mounted only on this pod — two narrow credentials (analyser read, git-adapter write), never one broad token. Built polyglot-ready NOW because the JS tools already exist (tech debt, not future planning).
- **sources:** PLAN_workflows_and_actions_migration(19).md (analyser adapter sections + repo access & security); FOCUS_js_tools_documentation.md
- **relations:** analyse_repo_local (the in-process alternative that later took the diagnose+indexer paths); adapter envelope contract
- **verify-later:** internal/adapters/analyser; analyser-adapter deployment in uk_001; whether request_repo_analysis still has callers

<!-- SOURCE: U16_docs019_design_plans.md -->
### Text-vs-code embeddings: share the mechanism, separate the policy (B4b)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19) B4b "decided 2026-06-09"; schema embodies it (sibling table + embedding_model column).
- **what:** Shared and not duplicated: the AIService embedder seam, provider implementations, pgvector, and the semantic+trigram hybrid pattern. Separate per domain so each upgrades independently: the model (prose vs code-specific), dimension, preprocessing (nomic search_ prefixes are caller-side), retrieval tuning (HNSW vs IVFFlat, lexical-heavier for code), row definition, and evaluation. Turns B4a into "which model for code", measurable independently. Caution recorded: separation pays only if the mechanism stays shared.
- **sources:** PLAN_workflows_and_actions_migration(19).md B4a/B4b/A5 resolutions
- **relations:** code_symbols; B4a ceiling; CPU-Ollama feasibility (bulk-index speed, code-domain recall)
- **verify-later:** embedding_model column values; whether a code-specific model was ever adopted

<!-- SOURCE: U16_docs019_design_plans.md -->
### Code-indexer agent, index-orchestrator wrapper, and CI-triggered indexing
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** code-indexer row applied (2026-06-12); index-orchestrator seeded (v2 after the v1 `name`-column failure, 2026-07-02); CI trigger still a queue item ("CI-triggered indexing (self-contained)" in HANDOFF_builder_thread).
- **what:** The code-indexer is a thin orchestrator (analyse → index_code_symbols → complete). Run 6dfa37cd proved orchestrate+agent_type is adopted IN-PLACE on the shared chassis pod — which never holds GITHUB_READ_TOKEN — so the index-orchestrator wraps it in the proven spawn pattern so the spawned pod receives the secret (isRepoCloningAgent gate). Manual trigger TRIGGER_code_indexer_v2.sh sends the explicit branch ref; the planned durable form is a GitHub Actions step firing the envelope with GITHUB_SHA on push, retiring the index-staleness class for the diagnosis corpus.
- **sources:** NNN_create_code_indexer_agent(2).sql; NNN_seed_index_orchestrator(1).sql; TRIGGER_code_indexer_v2(1).sh; HANDOFF_builder_thread.md#3
- **relations:** analyse_repo_local staleness incident; spawn-consumed columns lesson; reuse-index freshness (governance)
- **verify-later:** index-orchestrator row; CI workflow file existence

<!-- SOURCE: U16_docs019_design_plans.md -->
### Documentation indexing rides the prose rag path
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** Migration(19) "Documentation indexing (related, lighter…)" — direction recorded, not reported built.
- **what:** Docs are prose, so they fit rag_index/knowledge_base as-is: index each guide under a collection (e.g. `standards`), retrieve with rag_lookup; flat files in git remain the editable source of truth, the DB copy a derived rebuildable index so the assembler pulls relevant sections instead of pasting 124KB guides. Precondition: docs must live in a versioned repo. Separate, smaller workstream from code_symbols.
- **sources:** PLAN_workflows_and_actions_migration(19).md (documentation indexing section); FOCUS_js_tools_documentation.md
- **relations:** JS tools documentation gap; standards/docs agent (the matched-guidelines slot)
- **verify-later:** knowledge_base collections for docs

<!-- SOURCE: U16_docs019_design_plans.md -->
### cmd/bundle robustness contract (validate early, fail loud, manifest input)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** 003_contextkit_bundle_issues (2026-06-24) — two real failed runs analysed; HANDOFF_fixloop(8) notes the skip-message/usage patches exist but "needs gofmt + build".
- **what:** Field-found tool defects and the rules they teach: validate the cheapest input (-analysis JSON) BEFORE the slow psql-shelling gather phases, with an actionable message naming file/size/regeneration; accept a manifest/config file instead of 20-line backslash shell commands (kills the unquoted-parentheses class — real filenames contain "(1)"); a missing -doc/-scope path must fail loudly naming the path, because a silently-omitted file means a downstream chat reasons from incomplete context without knowing; single quoted -psql argument, no TTY.
- **sources:** 003_contextkit_bundle_issues.md; HANDOFF_fixloop_thread(8).md#2
- **relations:** bundle-first handoff practice; fail-loud-vs-degrade rule
- **verify-later:** cmd/bundle/main.go precondition ordering; -config support

<!-- SOURCE: U16_docs019_design_plans.md -->
### Reuse search before generation (code AND definition rows)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** Named in the tool plan §2 and bundle contract §2.1/§8; only the retrieval substrate (code_symbols, pgvector) exists — the pre-generation reuse step is unbuilt.
- **what:** Before generating, run "what already does something like this" over existing functions/structs and — critically in a data-defined system — over definition rows (workflows/agents/tools), so reuse-before-recreate is mechanical rather than a remembered habit and near-copies of existing workflows are caught like duplicate functions. Needs a searchable text projection for jsonb definitions (named open). The index is derived state that goes stale silently — re-index on change events and stamp freshness.
- **sources:** PLAN_context_assembly_tool_and_service(2).md#2; PLAN_bundle_shape_contract(2).md#2.1,#8; FOCUS_pre_build_edge_cases(1).md#2.4,#15
- **relations:** reuse_index_refresh trigger; code_symbols; dev-guide reuse discipline
- **verify-later:** any reuse-search action; definition-row indexing

<!-- SOURCE: U16_docs019_design_plans.md -->
### DB capabilities capture (\dx/\df into the bundle)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** Migration(19) B2a "decided, mechanism built" in the harness (dbcontext -capabilities, assembler -dbfacts); the durable indexing-workflow half is future.
- **what:** Generation that writes SQL should see installed extensions and helper functions (so it knows pgvector exists and reuses snapshot_agent instead of hand-rolling a backup — the migration-110 footgun). Captured as DB context (not the analyser's job), included for DB-touching tasks with a reuse nudge; the durable plan folds capture into the indexing workflow on a migration cadence so bundles always carry current DB facts without anyone remembering a flag.
- **sources:** PLAN_workflows_and_actions_migration(19).md B2a + 2026-06-09 changelog
- **relations:** multipass fetch; schema-before-SQL discipline
- **verify-later:** dbcontext -capabilities flag; assembler dbfacts section
