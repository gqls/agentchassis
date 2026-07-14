# Register — context-assembly
23 concepts, consolidated from 25 raw extractions across units U12, U14, U16, U17b,
U18, U19 (all tagged NEW:context-assembly, plus a handful of diagnosis-loop-tagged
raw blocks from U14/U18/U19 that describe the same production retrieval/assembly
infrastructure and were folded in here as the closer-fitting home). The
package_*.sh-vs-analyser trade-off writeup (U24d) was filed instead in
register/context-pack-tooling.md, matching its native category tag.

### CTXA-001 — "Documentation is code": the context-assembly tool and paid-service vision
- **status:** partial
- **status-evidence:** PLAN_context_assembly_tool_and_service(2) status: "A thin slice of Phase 1 is built and exercised on real code... first real run: a 30 KB bundle vs the script's 1.7 MB"; Phases 2–4 and the service itself unbuilt.
- **what:** The founding thesis: for any development task, assemble a task-scoped context bundle from documentation + codebase and feed generation against ground-truth verification, so results are "more likely to be correct" than pasting code into a chat. In an AI-driven workflow, documentation (standards, intent, trajectory) is itself an operational input — versioned, drift-detected, composed deterministically. Two audiences: dogfood on the chassis repo first, then a paid multi-tenant service behind a gateway. Design principles: engine/config split (tenant-agnostic engine + per-stack adapters — the decision that makes the service possible); seams for later machinery (cascade router, decision-point checkers, mediator) defined as interfaces from v1 but not built; dogfood first. Phased: 0 contracts+constitution, 1 bundle builder MVP, 2 verification loop, 3 service (sandboxed verification gates it), 4 cascade/checkers/mediator.
- **sources:** docs019/PLAN_context_assembly_tool_and_service(2).md; docs019/001_onboarding_discussion.txt; docs019/MAPPING_tool_to_actions_and_agents(2).md
- **relations:** bundle shape contract (CTXA-002); contextkit CLI toolchain (register/contextkit-toolchain.md); onboarding as the hard service problem
- **verify-later:** contextkit module state vs the plan's thin-slice claims; gateway project status

### CTXA-002 — Bundle shape contract (the task-scoped context package)
- **status:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) "Status: contract specification"; flagged by FOCUS_whole_plan_review §1.1 as "the next load-bearing contract".
- **what:** A bundle is assembled fresh per task (never stored/reused, so never stale) with fixed sections: metadata; task/target; the authored layer (constitution, why-chain, priority profile, direction-of-travel, matched standards); code context (in-scope code in full, neighbourhood signatures, reuse-search results, schema, definition data); database data in three kinds (CTXA-003); pointers to everything not inlined; and provenance (exactly what went in, logged as the decision log's `inputs_used`). Exists in both a canonical structured form and a rendered text view. Two integrity rules from the edge-case pass: assemble from a consistent snapshot (no torn reads), and log what the generator SAW (the rendered form), not what was assembled.
- **sources:** docs019/PLAN_bundle_shape_contract(2).md; docs019/FOCUS_whole_plan_review.md#1.1; docs019/FOCUS_pre_build_edge_cases(1).md#1.4,#4.3
- **relations:** three kinds of DB data (CTXA-003); altitude/step-type (CTXA-004); multipass fetch (CTXA-005); contextkit assembler (register/contextkit-toolchain.md — the harness prototype)
- **verify-later:** whether any structured bundle object exists in code vs the markdown-emitting harness

### CTXA-003 — Three kinds of database data (definition / operational / content)
- **status:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.1, verified against the live schema ("workflows are jsonb columns on agent_definitions... there is no workflows table and no tools table").
- **what:** In a data-defined system much of "the code" is DB rows, so the bundle distinguishes: definition data (the system's design as data — workflows in agent_definitions, tools as content_components rows, prompts as text columns; fetched routinely, covered by reuse-search); operational data (telemetry — work items, orchestration_states, error logs; multipass-capped); and content data (the output — sites/pages/tenant data; the gated set where privacy matters in the paid service).
- **sources:** docs019/PLAN_bundle_shape_contract(2).md#2.1
- **relations:** multipass fetch (CTXA-005); reuse search over definitions (CTXA-014); sensitivity gates
- **verify-later:** n/a (contract)

### CTXA-004 — Altitude: step type decides what the bundle emphasises
- **status:** deployed
- **status-evidence:** The harness assembler takes `-step framing|implementation|debug` (used in real bundle invocations, e.g. "-step debug", "the framing bundle that produced this plan used -step framing").
- **what:** The same task at different stages needs different context: framing/decision steps get full intent (why-chain, priority profile, direction-of-travel) with light code (in-scope shown as signatures only); implementation gets full code with a thin intent tether; debug leads with errors + runtime evidence + the expected-vs-actual diff, and adds a runtime-evidence section. "Right altitude at the right moment" made concrete as an explicit pipeline parameter in the bundle composer.
- **sources:** docs019/PLAN_bundle_shape_contract(2).md#4; docs019/PLAN_imagery_sprite_sheet.md (framing-bundle use); tasks/gameslink_missing_index_rerender/RUNBOOK...(2).md (-step debug); docs019/RUNBOOK_thin_slice(27).md#assembler-flags,#fuzzy-tasks
- **relations:** bundle shape contract (CTXA-002); salience-loss problem (register/context-engineering-principles.md)
- **verify-later:** assembler.go step handling

### CTXA-005 — Multipass fetch: probe → gate → include/reduce/point
- **status:** partial
- **status-evidence:** dbcontext "row data with multipass sizing" built in the harness (tool plan status); the full gate flow is contract-only.
- **what:** Query results have unknown size until run, so the builder probes with `LIMIT N+1` (not `count(*)` — counting a filtered query costs as much as running it), checks a size gate and a sensitivity gate, then includes rows in full, reduces (aggregate / representative sample / pointer), or gates behind confirmation. An oversized result becomes an aggregate or pointer, never an unbounded dump.
- **sources:** docs019/PLAN_bundle_shape_contract(2).md#3; docs019/FOCUS_pre_build_edge_cases(1).md#4.2; docs019/GUIDE_deploy_from_context_packs(1).md (dbcontext)
- **relations:** three kinds of DB data (CTXA-003); bounded bundle egress; dbcontext CLI (register/contextkit-toolchain.md)
- **verify-later:** dbcontext.go sizing logic

### CTXA-006 — Runtime evidence keyed by orchestration_id (the run narrative)
- **status:** partial
- **status-evidence:** dbcontext `-runtime-site` built and used throughout (bundles carried agent_error_log/site_work_items); the fuller run-trace composition is contract-level.
- **what:** For debugging, the most useful context is the narrative of a single run, reconstructable from one key: `orchestration_id` (+ `correlation_id`) spans `orchestration_states` (spawn tree, topics, status), `llm_call_log` (time-ordered step sequence), `agent_error_log` (error trail), pod logs (grep by run id) and the Kafka messages. Three cheap reads give a coherent single-run story instead of a scatter of lines. Log-correlation only works where the id is actually present in the log line — a coverage question the conventions agent audits.
- **sources:** docs019/PLAN_bundle_shape_contract(2).md#2.2; docs019/PLAN_onboarding_agent_specs(6).md#2.9,#1.9; docs019/README_02.md ("everything durable, one correlation id")
- **relations:** run signatures (CTXA-007); codebase-conditional capabilities (CTXA-009); diagnose_load_runtime (register/diagnosis-loop.md)
- **verify-later:** whether orchestration_id reliably appears in pod log lines (named open item)

### CTXA-007 — Run signatures: expected-vs-actual sequence diff
- **status:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.2 — designed; capture/storage named open in §9.
- **what:** Capture a healthy run's step sequence and spawn-tree shape once (from confirmed known-good runs), store it as an authored reference, and on a debug task diff the actual run against it to surface the divergence point — "matched the healthy path to step 7, then diverged here" instead of "read the logs." Verification-style thinking applied to runtime behaviour rather than static output.
- **sources:** docs019/PLAN_bundle_shape_contract(2).md#2.2,#9
- **relations:** runtime evidence by orchestration_id (CTXA-006); diagnostic playbooks (CTXA-008)
- **verify-later:** n/a (unbuilt)

### CTXA-008 — Diagnostic playbooks / failure fingerprints as authored knowledge
- **status:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.2: "seeded from those guides rather than authored fresh"; home (standards atoms vs sibling table) left open.
- **what:** Known failure fingerprints — a failure's signature plus the commands that confirm it plus the fix pattern — curated from existing debugging guides and failure writeups, surfaced into debug bundles the way standards are surfaced into build bundles, and grown as run-signature diffs reveal new ones.
- **sources:** docs019/PLAN_bundle_shape_contract(2).md#2.2,#9
- **relations:** run signatures (CTXA-007); debugging guide 016 (the seed corpus)
- **verify-later:** n/a (unbuilt)

### CTXA-009 — Codebase-conditional capabilities (degrade, don't break; partial config is normal)
- **status:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.3/§2.4 and agent-specs §2.9 — design rules, engine unbuilt.
- **what:** Capabilities like definition-data lookup, run-correlation, named logged steps, and log fetch all rest on structural facts that hold for THIS codebase but may not hold elsewhere. A stack-discovery step records which facts hold; each capability degrades to a weaker form or states "unavailable, because this codebase has no X" instead of breaking outright. Companion rule: distinguish not-yet-authored config (degrade gracefully, note what's pending) from malformed config (fail loud) — the no-fallbacks rule applies only to malformed data.
- **sources:** docs019/PLAN_bundle_shape_contract(2).md#2.3-2.4; docs019/PLAN_onboarding_agent_specs(6).md#2.9; docs019/FOCUS_pre_build_edge_cases(1).md#2.3; docs019/FOCUS_whole_plan_review.md#2.5
- **relations:** stack-discovery agent; convention coverage = capability reliability
- **verify-later:** n/a (design)

### CTXA-010 — Go analyser + call-graph neighbourhood (and the wiring-include gap)
- **status:** deployed
- **status-evidence:** Tool plan status: analyser/assembler "built and exercised on real code"; internal/analysis "already exists in the repo" per migration(19); bundles record "533 files analysed"; thin_slice(27) "known-limits: call-graph neighbourhood is name-matched, not type-resolved."
- **what:** A go/ast walk producing a structured index (signatures, types, per-function `calls` by name-based matching — full go/types resolution deliberately avoided as too heavy/fragile); the bundle's surrounding context is this call-graph neighbourhood (callees/callers/types) of the in-scope symbols, rendered as signatures, with `-neighbour package` as a fallback when name-matching misses (interface dispatch). Known structural blind spot: registration/init wiring (registry.go) is unreachable via `calls` since it fires through `init()`, not a call — the same gap as manually-named docs — so `-include` exists to force-include wiring files. Ubiquitous names (Run, String, New) are dropped when following the graph to avoid scope explosion. `ReadSymbolBody` (span-slice over start_line/end_line) was extracted once, tested byte-identical to `cmd/assembler`'s prior inline slicer, and shared with `diagnose_assemble_bundle` (see register/contextkit-toolchain.md for the tool-file detail).
- **sources:** docs019/PLAN_context_assembly_tool_and_service(2).md#5 status; docs019/GUIDE_deploy_from_context_packs(1).md; docs019/PLAN.md changelog (ReadSymbolBody); docs019/RUNBOOK_thin_slice(27).md#assembler-flags,#design-and-build-choices
- **relations:** call-graph re-scope mechanism (register/diagnosis-loop.md DIAG-007); analyser adapter (CTXA-013, wraps the same library); symbol-body slicer (register/contextkit-toolchain.md)
- **verify-later:** internal/analysis (chassis) vs contextkit copy drift (flagged in PLAN.md)

### CTXA-011 — code_symbols: the per-repo code index (pgvector sibling table)
- **status:** deployed
- **status-evidence:** Migration(19) changelog 2026-06-09: "code_symbols applied cleanly (table + 4 indexes)"; later populated (436 symbols, embeddings 436/436); code_retrieval_route(21) §7C "4,155 symbols; 499 distinct paths" after a full reindex.
- **what:** A sibling to `knowledge_base` reusing its proven shape (vector(768) nomic + trigram on content + idempotent dedup, same AIService embedder with caller-applied nomic prefixes) but keyed for code: unique(repo, path, symbol), SHA-versioned via `commit_sha`, identity upsert (`ON CONFLICT DO UPDATE WHERE content_hash IS DISTINCT` — a symbol persists across commits, re-embeds only on change), pruned hard on re-index (`commit_sha IS DISTINCT FROM $new`). Deliberate departures from chassis conventions (no version/previous_version_id, no deleted_at) because it is a rebuildable cache, not durable content. HNSW chosen over the knowledge_base's IVFFlat for incremental churn (pgvector 0.8.0 confirmed both available). One symbol = one row; no prose chunker (rag_index's character windows would fragment Go mid-function). Bodies are read from the repo at commit_sha, never stored.
- **sources:** NNN_create_code_symbols_index.sql; docs019/PLAN_workflows_and_actions_migration(19).md A5/B4/B4b; 048_NNN_create_code_symbols_index.sql (sql_for_tables unit — HNSW gate confirmation); code-indexer reuse mapping
- **relations:** hybrid retrieval (CTXA-012); embedding policy split (CTXA-016); knowledge_base reuse-not-copy; loop consumption (register/diagnosis-loop.md DIAG-036)
- **verify-later:** `\d code_symbols`; row counts per repo; embedding_model column use

### CTXA-012 — Hybrid code retrieval: index/lookup_code_symbols (vector + trigram, RRF in SQL)
- **status:** deployed
- **status-evidence:** Migration(19): actions built against real rag_actions.go, registered, deployed with the analyser (2026-06-12); lookup live in the diagnose workflow.
- **what:** `index_code_symbols` flattens the analyser's Output to symbol rows, skips unchanged content hashes, embeds non-fatally (trigram still works without an embedding), upserts and prunes; `lookup_code_symbols` mirrors `rag_lookup` (embed query → cosine vector search → trigram fallback → top-k + code_context). The hybrid RRF fusion moved into SQL (constant k=60), so the harness's in-Go `fuse` tool never graduated to production. Deliberately a sibling action, not a parameterised `rag_lookup` (knowledge-base columns are baked into `vectorSearchKB`); the three embedding helpers are shared package-level functions. Full read/write/dedup/prune SQL usage contract is documented in the migration's own comments.
- **sources:** docs019/PLAN_workflows_and_actions_migration(19).md (code-indexer reuse mapping); NNN_create_code_symbols_index.sql (query set); 048_NNN_create_code_symbols_index.sql
- **relations:** code_symbols table (CTXA-011); rag_lookup/rag_index (the mechanism source); repo-label symmetry (CTXA-017)
- **verify-later:** code_symbols_actions.go; registry entries (storage/code categories)

### CTXA-013 — Analyser adapter: in-cluster polyglot parsing service
- **status:** deployed
- **status-evidence:** Migration(19) changelog 2026-06-12: "ANALYSER DEPLOYED TO PRODUCTION (uk_001)... the code-indexer agent row is applied (INSERT 0 1)."; RUNBOOK(31) §6D "Adapter availability — RESOLVED 2026-06-24 (secret name/key mismatch)... the adapter came up 1/1 Running."
- **what:** A Kafka message-based adapter (`internal/adapters/analyser`, topic `system.adapter.analyser.requests`, mirroring git/thunder rather than HTTP) consuming `analyse` requests (owner, repo, ref, language), fetching source read-only via the GitHub tarball endpoint (one request, recovers the exact commit_sha, path-traversal-guarded), parsing through the `analysis.Analyse` library behind an `Analyser` seam so per-language parsers (JS next) drop in, and replying on the caller's responses topic. Security: its own least-privilege repo-scoped read-only PAT as a k8s Secret mounted only on this pod via secretKeyRef (never envFrom, which exposes every platform secret) — two narrow credentials (analyser read, git-adapter write), never one broad token. Built polyglot-ready from the start because the JS tools already existed (tech debt, not future planning). Deployment lessons: topic auto-create is off so the KafkaTopic CRD must exist; topic-addressed adapters legitimately show `target_agent_type='unknown'` in awaited_requests; idle consumer-poll timeouts log at ERROR cosmetically. Its per-iteration use inside the diagnosis loop was later removed in favour of `analyse_repo_local` (register/diagnosis-loop.md DIAG-023), but indexing and other consumers remain on it.
- **sources:** docs019/PLAN_workflows_and_actions_migration(19).md (analyser adapter + repo access & security sections); docs019/FOCUS_js_tools_documentation.md; docs019/RUNBOOK(31)_diagnosis_loop.md#6D; docs019/RUNBOOK_thin_slice(27).md#in-cluster-path
- **relations:** analyse_repo_local (CTXA-014, the in-process alternative for the diagnose/indexer paths); repo-cloning token gate (register/diagnosis-loop.md DIAG-022); adapter envelope contract
- **verify-later:** internal/adapters/analyser; deployments/kustomize/services/analyser-adapter; whether request_repo_analysis still has callers

### CTXA-014 — analyse_repo_local in-process analysis and the stale-index incident
- **status:** deployed
- **status-evidence:** NNN_swap_indexer_to_local_analysis header documents the incident and the applied swap; PATCH_lift_fetcher_and_register gives the registration; README_flows/handoffs treat the local path as current.
- **what:** The adapter round-trip left the diagnose-agent with no local checkout (`ReadSymbolBody` could not slice bodies), and once resolving `ref "HEAD"` silently indexed a year-old commit (69 files/436 symbols vs 572 current). `analyse_repo_local` fetches the repo tarball at an EXPLICIT ref to a local temp dir and analyses in-process (sharing the analyser's `internal/reposource` fetcher), with `pin_to_index_commit` defaulting true for the diagnose loop (bodies must match the index) and false for the indexer (which defines the commit). Corollary rules born from the incident: explicit git refs, never HEAD; the spawned pod needs `GITHUB_READ_TOKEN` via the `isRepoCloningAgent` spawn gate. See register/diagnosis-loop.md DIAG-023 for the diagnose-agent's own use of this mechanism; this entry is the historical incident + corpus-freshness angle.
- **sources:** NNN_swap_analyse_repo_to_local.sql; NNN_swap_indexer_to_local_analysis.sql; PATCH_lift_fetcher_and_register.md; TRIGGER_code_indexer_v2(1).sh
- **relations:** analyser adapter (CTXA-013, request_repo_analysis stays for the code-indexer); analyse_repo_local as used by the diagnose-agent (register/diagnosis-loop.md DIAG-023); stale-corpus doctrine (register/diagnosis-loop.md DIAG-037)
- **verify-later:** internal/reposource/github_source.go; analyse_repo_local_action.go; registry entry

### CTXA-015 — Code-indexer agent, index-orchestrator wrapper, and CI-triggered indexing
- **status:** partial
- **status-evidence:** code-indexer row applied (2026-06-12); index-orchestrator seeded (v2 after a v1 `name`-column failure, 2026-07-02); CI trigger still a queue item ("CI-triggered indexing (self-contained)" in HANDOFF_builder_thread); 118_code_indexer_for_analyser.sql marked DRAFT/experimental with a "[checked 2026-06-11]" annotation confirming composition landed in IndexCodeSymbolsAction.
- **what:** `code-indexer` is a thin orchestrator (`request_analysis` → `index_symbols` → `complete`): asks the analyser adapter to parse a repo@ref into symbols, then `index_code_symbols` upserts them into code_symbols (embedding changed symbols via ollama/nomic-embed-text, pruning symbols absent from the commit). Retrieval is a separate `lookup_code_symbols` action used by other agents. Run 6dfa37cd proved `orchestrate`+agent_type is adopted in-place on the shared chassis pod — which never holds `GITHUB_READ_TOKEN` — so `index-orchestrator` wraps it in the proven spawn pattern (register/diagnosis-loop.md DIAG-019/DIAG-022) so the SPAWNED pod receives the secret. Manual trigger `TRIGGER_code_indexer_v2.sh` sends an explicit branch ref; the planned durable form is a GitHub Actions step firing the envelope with `${GITHUB_SHA}` on push, retiring the index-staleness class for good. Non-git corpora may override the repo label (e.g. `domain:kruste.com`).
- **sources:** NNN_create_code_indexer_agent(2).sql; NNN_seed_index_orchestrator(1).sql; TRIGGER_code_indexer_v2(1).sh; HANDOFF_builder_thread.md#3; 118_code_indexer_for_analyser.sql; NNN_create_code_indexer_agent.sql (earlier draft, superseded by the applied version)
- **relations:** analyse_repo_local staleness incident (CTXA-014); repo-cloning token gate (register/diagnosis-loop.md DIAG-022); stale-corpus doctrine (register/diagnosis-loop.md DIAG-037)
- **verify-later:** index-orchestrator row; CI workflow file existence; agent_definitions CHECK constraints referenced in the draft migration

### CTXA-016 — Text-vs-code embeddings: share the mechanism, separate the policy (B4b)
- **status:** deployed
- **status-evidence:** Migration(19) B4b "decided 2026-06-09"; schema embodies it (sibling table + embedding_model column).
- **what:** Shared, not duplicated: the AIService embedder seam, provider implementations, pgvector, and the semantic+trigram hybrid pattern. Kept separate per domain so each upgrades independently: the model (prose vs code-specific), dimension, preprocessing (nomic `search_` prefixes are caller-side), retrieval tuning (HNSW vs IVFFlat, lexical-heavier for code), row definition, and evaluation. Turns the B4a question into "which model for code", independently measurable. Caution recorded: separation only pays off if the underlying mechanism stays shared.
- **sources:** docs019/PLAN_workflows_and_actions_migration(19).md B4a/B4b/A5 resolutions
- **relations:** code_symbols table (CTXA-011); B4a retrieval ceiling (register/context-engineering-principles.md); CPU-Ollama feasibility
- **verify-later:** embedding_model column values; whether a code-specific model was ever adopted

### CTXA-017 — code_symbols repo-label symmetry (shared owner/repo resolver)
- **status:** deployed
- **status-evidence:** RUNBOOK(31) "ROOT CAUSE CONFIRMED (2026-06-26) — repo-label asymmetry... the lookup queried WHERE repo='agentchassis' against rows under 'gqls/agentchassis' → 0 hits"; "Structural patch APPLIED"; PATCH_code_symbols_shared_repo_label written; NNN_fix_lookup_repo_label_workaround header: "apply that and REVERT this when the rebuilt image ships."
- **stage2-verified (2026-07-14):** partial → deployed — platform/orchestration/actions/code_symbols_actions.go:53 defines resolveCodeRepoLabel, called at code_symbols_actions.go:93,170 (index and lookup) AND diagnose_route_action.go:413 — one shared resolver used by both writer and reader in production code, confirming the structural fix landed, not just a config workaro...
- **what:** `code_symbols.repo` is always the composed `owner/repo` label (thin_slice(27) "Label convention — DECIDED 2026-06-11"). `index_code_symbols` composed it correctly but `lookup_code_symbols` did not, so the diagnose workflow's lookup queried the bare name against composed rows and returned nothing — iteration-1 seeding returned "no scope." Fixed twice: a config-only workaround (literal repo string hard-coded on the lookup step) then the structural fix, one shared `resolveCodeRepoLabel` used by BOTH index and lookup so they cannot diverge again. General lesson: writer and reader of a keyed store must share the key resolver. Also produced a standing diagnostic rule: confirm results by correlation_id, never by `LIMIT 1` (a COMPLETED LIMIT-1 row was a red herring twice during this investigation).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6F-run1; docs019/RUNBOOK_thin_slice(27).md#label-convention; PATCH_code_symbols_shared_repo_label.md; NNN_fix_lookup_repo_label_workaround.sql; NNN_create_code_indexer_agent(2).sql (label convention)
- **relations:** code_symbols table (CTXA-011); loop-back plumbing fault class (register/diagnosis-loop.md DIAG-026 — same dotted-path/shared-resolver config-contract class)
- **verify-later:** code_symbols_actions.go resolveCodeRepoLabel; whether the workaround REVERT ran

### CTXA-018 — Documentation indexing rides the prose rag path
- **status:** aspirational
- **status-evidence:** Migration(19) "Documentation indexing (related, lighter...)" — direction recorded, not reported built.
- **what:** Docs are prose, so they fit `rag_index`/`knowledge_base` as-is: index each guide under a collection (e.g. `standards`), retrieve with `rag_lookup`; flat files in git remain the editable source of truth, the DB copy a derived rebuildable index so the assembler can pull relevant sections instead of pasting 124KB guides whole. Precondition: docs must live in a versioned repo. A separate, smaller workstream from code_symbols.
- **sources:** docs019/PLAN_workflows_and_actions_migration(19).md (documentation indexing section); docs019/FOCUS_js_tools_documentation.md
- **relations:** JS tools documentation gap; standards/docs agent (the matched-guidelines slot)
- **verify-later:** knowledge_base collections for docs

### CTXA-019 — cmd/bundle robustness contract (validate early, fail loud, manifest input)
- **status:** partial
- **status-evidence:** 003_contextkit_bundle_issues (2026-06-24) — two real failed runs analysed; HANDOFF_fixloop(8) notes the skip-message/usage patches exist but "needs gofmt + build."
- **what:** Field-found tool defects and the rules they teach: validate the cheapest input (`-analysis` JSON) BEFORE the slow psql-shelling gather phases, with an actionable message naming file/size/regeneration; accept a manifest/config file instead of 20-line backslash shell commands (kills the unquoted-parentheses class — real filenames contain "(1)"); a missing `-doc`/`-scope` path must fail loudly naming the path, because a silently-omitted file means a downstream chat reasons from incomplete context without knowing it; single-quoted `-psql` argument, no TTY.
- **sources:** docs019/003_contextkit_bundle_issues.md; docs019/HANDOFF_fixloop_thread(8).md#2
- **relations:** bundle-first handoff practice (register/context-pack-tooling.md); fail-loud-vs-degrade rule; cmd/bundle (register/contextkit-toolchain.md)
- **verify-later:** cmd/bundle/main.go precondition ordering; -config support

### CTXA-020 — Reuse search before generation (code AND definition rows)
- **status:** aspirational
- **status-evidence:** Named in the tool plan §2 and bundle contract §2.1/§8; only the retrieval substrate (code_symbols, pgvector) exists — the pre-generation reuse step itself is unbuilt.
- **what:** Before generating, run "what already does something like this" over existing functions/structs and — critically, in a data-defined system — over definition rows (workflows/agents/tools), so reuse-before-recreate is mechanical rather than a remembered habit, and near-duplicate workflows are caught the way duplicate functions would be. Needs a searchable text projection for jsonb definitions (named open). The index is derived state that goes stale silently — re-index on change events and stamp freshness.
- **sources:** docs019/PLAN_context_assembly_tool_and_service(2).md#2; docs019/PLAN_bundle_shape_contract(2).md#2.1,#8; docs019/FOCUS_pre_build_edge_cases(1).md#2.4,#15
- **relations:** reuse_index_refresh trigger; code_symbols (CTXA-011); dev-guide reuse discipline; reuse-check retrieval pipeline design (register/context-engineering-principles.md — the general design lens)
- **verify-later:** any reuse-search action; definition-row indexing

### CTXA-021 — DB capabilities capture (\dx/\df into the bundle)
- **status:** partial
- **status-evidence:** Migration(19) B2a "decided, mechanism built" in the harness (dbcontext -capabilities, assembler -dbfacts); the durable indexing-workflow half is future.
- **what:** Generation that writes SQL should see installed extensions and helper functions (so it knows pgvector exists and reuses `snapshot_agent` instead of hand-rolling a backup — the migration-110 footgun). Captured as DB context (not the analyser's job), included for DB-touching tasks with a reuse nudge; the durable plan folds capture into the indexing workflow on a migration cadence so bundles always carry current DB facts without anyone remembering a flag.
- **sources:** docs019/PLAN_workflows_and_actions_migration(19).md B2a + 2026-06-09 changelog
- **relations:** multipass fetch (CTXA-005); schema-before-SQL discipline; dbcontext CLI (register/contextkit-toolchain.md)
- **verify-later:** dbcontext -capabilities flag; assembler dbfacts section

### CTXA-022 — contextkit module packaging and the graduation seam
- **status:** deployed
- **status-evidence:** Tool plan status: "The tools are now a single Go module, contextkit/... the two contracts defined once — internal/analysis and internal/candidates"; production-relevant parts graduated in-cluster per migration(19).
- **what:** All harness tools (analyser, embed, resolve_targets, fuse, eval_targets, assembler, dbcontext, bundle, dedup, diagnose) live in one module sharing two contracts; "graduation" moves the internal packages under the chassis module path and turns command mains into actions without changing the contracts. Production runs in-cluster; the harness remains the dev/measurement scaffold (eval_targets stays offline, the flywheel's measurement tool). The trial's output is throwaway; the rule it teaches (one shared contract, no per-tool copies) is durable.
- **sources:** docs019/PLAN_context_assembly_tool_and_service(2).md status; docs019/PLAN_workflows_and_actions_migration(19).md (analyser-adapter sections); docs019/MAPPING_tool_to_actions_and_agents(2).md
- **relations:** analyser adapter (CTXA-013); one-decision-core two realisations (register/diagnosis-loop.md DIAG-017); contextkit CLI toolchain (register/contextkit-toolchain.md)
- **verify-later:** go_files/contextkit module contents

### CTXA-023 — "Verified against ALL forks" claim did not actually reconcile (fork-drift finding)
- **status:** partial
- **status-evidence:** "This entry and its three siblings... are genuinely absent from the canonical live 016b_debugging_guide_8_consolidated.md/merged(1).md — they continue only in a parallel travelling_docs/016b_debugging_guide_7_3_(2..7).md fork the canonical consolidation's 'verified against ALL forks' claim did not actually reconcile."
- **what:** A documentation-integrity finding surfaced while tracing the `error_step` config-placement bug (register/diagnosis-loop.md DIAG-030): a consolidated debugging guide claimed to be "verified against ALL forks" of a travelling document family, but a parallel fork (`016b_debugging_guide_7_3_(2..7).md`) in fact continued independently and diverged, carrying content the canonical consolidation never merged back in. Filed here as a context-assembly/documentation-fidelity concern rather than a diagnosis-loop mechanism, since it is about the trustworthiness of an authored corpus a bundle might pull from.
- **sources:** archive_april_26/016b_debugging_guide_7(4).md#"error_step: config-level placement..."
- **relations:** category note — documentation-system fork-reconciliation concern; error_step / anchorless diagnosis (register/diagnosis-loop.md DIAG-030)
- **verify-later:** whether the 016b guide family has since been reconciled
