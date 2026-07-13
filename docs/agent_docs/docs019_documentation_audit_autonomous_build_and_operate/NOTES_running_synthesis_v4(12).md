# NOTES — running synthesis v4 (started 2026-07-02)

Continues from `NOTES_running_synthesis_v3.md` (archived — do not append there).
Older transcripts are catalogued in `journal.txt` alongside the transcripts.

## STATE OF THE WORLD (as of 2026-07-02, chassis replicaset 5786cffbd4)

- **Diagnosis loop (§6): DONE.** §6A–§6G all passed; §6G accepted on run `51f95cda`
  (refute-and-confirm-a-grounded-cause bar). Engine (pkg/diagnose) + diagnose-agent
  workflow live. Deployed + verified: data_request-as-progress (SeenRequests),
  verdict_wire.go DataRequests sync, ## Schema bundle section (denylist +
  schema_full toggle) + EXPLAIN size-guard, sqlguard stripQuoted (literal
  false-positive fix), composing repo label (resolveCodeRepoLabel in lookup+index;
  workaround REVERTED; composition verified live: lookup logged
  repo="gqls/agentchassis").
- **Diagnoses are carried entirely by the runtime/DB channel.** The code-retrieval
  channel contributes nothing (measured: flat similarity band 0.547–0.574 across
  all 12 seed hits; zero code citations in four full runs). Current route (§7, new
  runbook `RUNBOOK_code_retrieval_route.md`) is to make it earn its keep.
- Open, parked: trigger sends site_id intermittently empty; two verdict-quality
  wrinkles (stale text carried into a CONFIRMED conclusion that its own citation
  contradicts; data_requests emitted in a terminal verdict never run); non-manual
  triggering deferred until the above.

## 2026-07-02 — corpus check result: the index is the blocker (route §7 opened)

- User ran the corpus verification (columns taken from code_symbols_actions.go's
  own INSERT/SELECT). Result: **0 rows** for any path/symbol ILIKE %page% /
  %section% / %result_spec% / %resolveResultSpec% in `code_symbols` for
  gqls/agentchassis; 436 total rows, 67 with empty doc.
- 0-rows discipline applied BEFORE concluding: ILIKE is case-insensitive (receiver
  forms like `(*SagaCoordinator).resolveResultSpec` still substring-match), the OR
  block is parenthesised, and the 436-count query proves the repo literal matches.
  The query is sound; the absence is real.
- **Analyser cleared as the cause** (read /mnt/project/analyse.go in full):
  AnalyseWithExclude has NO size caps and NO directory skips beyond
  vendor/testdata/hidden-dirs/*_test.go/dup-suffix `(N).go`, plus caller-supplied
  substring excludes. A walk of the chassis tree with this code includes
  result_spec.go and the page/section action files. One sharp edge noted: a
  mid-walk error returns a PARTIAL Output alongside the error (analyse.go:103), so
  a caller that ignores the error keeps a truncated tree.
- **Inference chain (premises stated):**
  1. The tree at 4c2c172 contains the missing files — the deployed binary logs
     from `orchestration/result_spec.go:185`, and production executed the
     `save_page_sections` action (error-log rows), so those sources exist in the
     repo the image was built from.
  2. The analyser has no rule that would drop them (read this turn).
  3. The index has zero of them (user's query).
  4. The §6D writer — the `code-indexer` agent, workflow
     `request_repo_analysis → await analyser → index_code_symbols` (old RUNBOOK
     §6D) — is the ONE path between analyser and index that differs from the
     verified-good local path: it crosses Kafka via the analyser ADAPTER (its own
     fetch, possible excludes, message-size exposure, or an ignored partial-walk
     error).
  ⇒ The fault lives in the §6D adapter path. WHICH mechanism is not yet known —
  the §7A census discriminates (whole directories absent ⇒ adapter excludes or a
  subset fetch; scattered/alphabetical cut ⇒ truncation or ignored partial walk;
  commit anomalies ⇒ wrong ref/prune interplay). Do not pre-commit to a mechanism.
- **Consequence for the retrieval design:** the evidence-fed resolver
  (resolve-before-Advance in diagnose_route) stays the right architecture move but
  is pointless until the corpus contains the code the loop needs — no query
  improvement can retrieve absent rows. So the route is corpus first, resolver
  second. Full plan in `RUNBOOK_code_retrieval_route.md` (§7A–§7F).

## 2026-07-02 pm — §7A answered: the index is of a YEAR-OLD tree; §7B migration written

- Census (user-run): 29 dirs, 69 files, 436 symbols, ALL at commit 4c2c172;
  orchestration_all=20 (2 files), orchestration_actions=0. Ground truth:
  `git log -1 4c2c172` ⇒ **2025-07-14** "additional test fuctions"; 572 non-test
  .go files on disk today; result_spec.go present (mtime Jun 21).
- Reading: the census sums to exactly 69 files — a COHERENT early repo snapshot
  (auth-service 24 files dominant; actions/ absent because it barely existed),
  not a cut of the current tree. No alphabetical truncation pattern; single
  commit ⇒ prune interplay excluded. The §6D trigger envelope sent ref:"HEAD"
  (old RUNBOOK block) ⇒ the analyser ADAPTER resolved HEAD to a stale checkout
  on its side (baked-in/never-pulled clone). Adapter named as the fault; trigger
  and analyser cleared.
- RETROACTIVE CONSEQUENCE: diagnose's analyse_repo_local pins to the dominant
  index commit (pin_to_index_commit default TRUE), so every diagnose run fetched
  and body-sliced the July-2025 tree. The whole code channel — index AND slicing
  — has been reading a repo that predates the build-domain code. CORRECTION of
  emphasis: I earlier attributed the dead-weight channel primarily to embedding
  semantics (symptom vs code vocabulary); the dominant cause is the stale corpus
  — the flat 0.547–0.574 band is what a query looks like against a corpus that
  simply does not contain its subject matter. The semantics point may still
  apply post-reindex; re-measure at §7E before acting on it.
- §7B migration WRITTEN (`NNN_swap_indexer_to_local_analysis.sql`), against the
  user-dumped code-indexer default_config. Two corrections forced by READING the
  deployed action (both would have silently broken a from-memory migration):
  (1) pin_to_index_commit DEFAULTS TO TRUE in analyse_repo_local — the indexer
  must set FALSE or it re-fetches the old dominant commit forever; (2) the
  return map has the Output fields TOP-LEVEL (outputToMap + commit_sha/owner/
  repo/ref alongside; NO .output key) ⇒ index_symbols.analysis_field
  "repo_analysis.output" → "repo_analysis" [DELIBERATE CHANGE]. Also timeout
  300→1800 (thousands of ollama embeds) and complete.output_fields → ["index_result"]
  (drop the multi-MB analysis JSON from the Kafka completion) [DELIBERATE].
  Step name request_analysis retained (action swapped, name not).
- Token: analyse_repo_local reads GITHUB_READ_TOKEN (const, line 69); comment
  says injected to diagnose-agent pods via spawn_actions. The pod executing
  code-indexer must carry it — check at trigger time; failure is loud (401/403).
- PARKED: the analyser adapter's own staleness (its HEAD ≠ repo HEAD). The §7B
  swap removes it from the indexer path; if the adapter serves anything else,
  its checkout/image needs fixing separately.

## 2026-07-02 eve — pin-down: remote HEAD (stale default branch); token design fact; §7B applied

- §7B migration APPLIED (snapshot 971da9c9); verify returned exactly the intended
  config (analyse_repo_local + pin_to_index_commit:false; analysis_field
  "repo_analysis"; complete=["index_result"]; timeout 1800).
- ADAPTER PIN-DOWN (user asked "why did it use an old commit — it should always
  use the latest"): read adapter.go + analyse_action.go + github_source.go +
  analyser-adapter.yaml (uploads). The adapter's code is EXONERATED: ref passes
  through verbatim to `GET {api}/repos/{o}/{r}/tarball/{ref}`; "HEAD" only as the
  empty-ref default; NO cache (fresh temp dir per call); empty GITHUB_API_BASE =
  https://api.github.com. Therefore GitHub itself resolved tarball/HEAD → 4c2c172
  on 2026-06-26. Remote HEAD is a symbolic ref to the DEFAULT BRANCH ⇒ the
  repo's default branch has not moved since 2025-07-14 while development happens
  on another branch. The semantic gap: locally HEAD = your checkout; remotely
  HEAD = the default branch. Nothing errored because nothing was wrong from any
  component's view. Nobody noticed because GitHub Actions triggers on push to
  the WORKING branch — the default branch is invisible to the deploy path; the
  analyser was the first consumer to resolve remote HEAD.
  CONFIRMATION (one command, pending): `git ls-remote origin HEAD` — expect
  4c2c172b…; supporting: `git remote show origin | grep 'HEAD branch'`.
- CORRECTION (of the 07-02 pm note "adapter named as the fault"): the adapter
  path was the messenger; the cause is repo configuration. The §7B swap remains
  right (one fetcher, no Kafka round-trip, shared code with diagnose) but was
  aimed at the wrong culprit — and is INSUFFICIENT alone: analyse_repo_local
  builds the identical tarball/HEAD URL, so an unpinned reindex at ref:"HEAD"
  would fetch 4c2c172 again. §7C now gates on: default branch corrected AND/OR
  explicit branch ref in the envelope (both, ideally).
- TOKEN DESIGN FACT (spawn_actions.go ~2503-2520 + isRepoCloningAgent ~3025,
  user-pasted + read): GITHUB_READ_TOKEN is injected as a secretKeyRef from
  Secret personae-platform-secrets into SPAWNED pods only, gated by
  isRepoCloningAgent = ["diagnose-agent"]. Comment: "the spawning chassis pod
  never holds it". ⇒ my earlier remedy "add the env to the standing/chassis
  deployment" is WITHDRAWN for the chassis. Decision tree (probe first:
  `kubectl -n ai-persona-system get deploy,pods | grep -i index`):
  (a) standing code-indexer deployment → add the same secretKeyRef to ITS
  manifest; (b) in-place on chassis → add 'code-indexer' to isRepoCloningAgent
  and run the indexer as a spawned agent (reuse the orchestrator spawn→call
  pattern). No spawn_actions patch shipped yet — conditional on the probe.
- NEXT: user runs `git ls-remote origin HEAD` + the deploy/pods probe; fix
  default branch; then §7C trigger with REF='<branch>' explicit.

## 2026-07-02 late — pin CONFIRMED (main stale since 2025-07-14); gate line agreed; ref strategy set

- CONFIRMED: `git ls-remote origin HEAD` = 4c2c172bb3faee… ; default branch =
  `main`; working branch = `083_imagery`. Refinement: not an abandoned master —
  main is the default but NOTHING has merged into it since 2025-07-14; dev runs
  on serial numbered feature branches and deploys from their pushes. Remote
  HEAD (= main) therefore truthfully reports a branch nobody uses.
- TOKEN RESOLVED: code-indexer is dynamically spawned (user-stated; pod list
  shows agent-<type>-<id> jobs, no standing indexer deployment). Remedy = the
  user's exact one-line change: repoCloningAgents gains "code-indexer". Causal
  chain worth recording: §6D never needed the token in the indexer pod because
  the ADAPTER held it and fetched; §7B moved the fetch into the indexer pod,
  creating the need. Requires rebuilding the spawning image(s) (core-manager /
  chassis; remote-job-spawner too if it builds pod specs from this repo);
  verify on the spawned pod (describe | grep GITHUB_READ_TOKEN).
- REF STRATEGY (answer to "make ref selection dynamic, not hardcoded"):
  IMMEDIATE — derive in the trigger script: REF="$(git -C $CHASSIS rev-parse
  --abbrev-ref HEAD)" (tarball/<branch> = remote tip; push first).
  STRUCTURAL A (recommended) — CI-triggered indexing: post-deploy Actions step
  sends the §6D envelope with REF="${GITHUB_SHA}" via kcat from the in-cluster
  runners ⇒ index commit == deployed commit by construction; ref choice ceases
  to exist. Indexing is a safe writer — distinct from the human-gated
  auto-diagnosis question.
  STRUCTURAL B (composable) — CI fast-forwards main to the deployed sha
  post-deploy ⇒ main = deployed-pointer ⇒ remote HEAD truthful for every
  consumer. Changes main's meaning; user's call.
  REJECTED — in-action "most recently pushed branch" GitHub-API resolver:
  N+1 walk and latest-pushed ≠ deployed (experimental branches would poison
  the index); the deploy event already knows the sha.
- Diagnose side unaffected (pin override follows the new index commit).

## 2026-07-02 night — ref decision recorded; trigger reviewed; ordering constraint

- USER DECISION (HEAD hygiene): merge meaningful commits to main manually; main
  stays the default branch; envelopes still always carry an explicit branch.
  (Structural B, manual variant; CI fast-forward remains a later option.)
- Trigger script reviewed against the migrated workflow: envelope shape correct
  (orchestrate + agent_type=code-indexer; input_data owner/repo/ref/language;
  ref_field reads input_data.ref). Stale classifier-template cosmetics only
  (pod/orchestration names, watch greps naming search_domain/scrape_site steps
  that do not exist here) — corrected copy written: TRIGGER_code_indexer_v2.sh
  (bash -n clean) with accurate step markers (request_analysis|index_symbols,
  analyse_repo_local|index_code_symbols), token-injection check, §7C pointers.
- HARD ORDERING (the one blocker): deploy the spawn-gate line BEFORE
  triggering — the spawned indexer pod's GITHUB_READ_TOKEN secretKeyRef is
  written by the NEW spawning image; trigger-first fails loudly at the token
  check (recoverable, wasted run). Sequence: gate line -> build -> deploy ->
  push branch (same push feeds tarball tip) -> trigger -> watch -> §7C verify.
- Expected scale: first run embeds every symbol (all new at the new commit) via
  ollama-adapter — minutes; timeout now 1800s. Subsequent runs embed only
  changed content hashes.

## 2026-07-02 night+1 — run 6dfa37cd: token error was CORRECT; "dynamically spawned" was wrong; §7B.1 opened

- Run 6dfa37cd (trigger REF=083_imagery, new image 7c65fbdf64 WITH the gate
  line): failed at request_analysis with the GITHUB_READ_TOKEN error — ON THE
  GENERIC CHASSIS POD (agent_type="generic", agent-chassis-7c65fbdf64-69wr9).
  So: orchestrate+agent_type=code-indexer is ADOPTED IN-PLACE; nothing spawned;
  the gate (spawn-time only) never consulted; the chassis deliberately never
  holds the token ⇒ the error is correct and correctly placed. CORRECTION of
  yesterday's "code-indexer is dynamically spawned" (typed pods exist only for
  workflows containing spawn_agent). §6D worked in-place because the ADAPTER
  fetched. Two run-blockers stacked: token, and (user-noted) 083_imagery may
  not exist on GitHub yet — a push is required either way.
- The user's "that pod does have it" describe output was a command artefact:
  grep matched no code-indexer pod ⇒ empty $() ⇒ bare `kubectl describe pod`
  described EVERY pod ⇒ the -A3 grep landed on the ANALYSER-ADAPTER's env
  (its POD_NAME visible in the paste). Trigger's watch command now guards the
  empty case.
- The user's hunch "maybe it's the error message placement" was half right:
  placement correct, WORDING stale ("the diagnose-agent pod must carry…").
  Patched analyse_repo_local_action.go — ONE string changed, no names — to say
  it must run in a spawned repo-cloning pod and to point at the orchestrator.
- FIX (§7B.1): index-orchestrator mirroring diagnose-orchestrator verbatim
  (spawn_agent code-indexer role=indexer → call_agent forwarding
  owner/repo/ref/language timeout 1800 → complete result_from
  code-indexer_result). Seed: NNN_seed_index_orchestrator.sql (schema-check
  preamble; copies version/category/status from the diagnose-orchestrator
  row; REVERT=DELETE). Trigger retargeted to index-orchestrator. REJECTED:
  token-on-chassis (design forbids); §7B revert to the adapter path (explicit
  ref WOULD fetch correctly now, but the 572-file Output reply over Kafka is
  multi-MB vs ~1MB default max.message — §6D's 69 files masked this); bare k8s
  Job (outside the agent architecture).
- NEXT: push 083_imagery; \d agent_definitions + apply seed; retrigger via
  index-orchestrator; expect agent-code-indexer-<id> pod WITH the secretKeyRef;
  then §7C verification.

## 2026-07-02 night+2 — seed v1 failed on schema; v2 written from the \d paste

- v1 INSERT failed: no `name` column (display_name only) and `category`
  (varchar, NOT NULL, no default — distinct from the nullable CHECK'd
  agent_category) was omitted. The migration's own preamble mandated \d first
  and the column list was still written from memory — rule violated, owned.
- v2 keyed to the pasted schema: columns (type, display_name, description,
  category, agent_category, status, version, default_config); category/
  agent_category/status COPIED from the live diagnose-orchestrator row with
  is_snapshot/deleted_at filters (snapshot_agent writes into this table;
  (type,version) is unique); version=1 explicit; workflow JSON unchanged.
- Schema paste also re-confirms the parked stale columns
  (task_workflow/orchestrator_workflow/orchestration_workflow) for later
  nulling.

## 2026-07-02 night+3 — run 93ba14e6: spawn chain WORKS; index in flight at 36710be

- Seed v2 applied (INSERT 0 1). Copied values: category='diagnose' (harmless —
  no CHECK on `category`; cosmetic tidy parked), agent_category=coordinator,
  status=experimental, version=1.
- RUN 93ba14e6 / parent orch b0365057 / child f6909b33: spawn_indexer →
  agent-code-indexer-65546fca-cssc8 (a REAL spawned pod this time) →
  call_indexer forwarded owner/repo/ref/language, reply_to the parent's
  system.agent.generic.responses. analyse_repo_local SUCCEEDED on the spawned
  pod: NO token error ⇒ isRepoCloningAgent("code-indexer") + secretKeyRef
  verified end to end; fetched 083_imagery @ commit 36710be (current, not
  4c2c172); 515 files analysed (~9s); cmd/analyser-adapter + test/ visible in
  the Output ⇒ full tree.
- index_symbols IN FLIGHT: since_s grew to 516 with status EXECUTING_STEP —
  EXPECTED, not a hang signal: one long action produces no state transitions;
  first run at a new commit embeds EVERY symbol (est. 2.5–4k) through the
  single ollama-adapter ⇒ 10–25 min; child ceiling 1800s. RESUME property: a
  timeout re-trigger skips already-upserted unchanged content hashes.
- Progress probe (recorded in runbook §7C): count by commit_sha — growing
  36710be beside the static 436×4c2c172 (prune runs at the end) ⇒ working.
- 515 vs local find 572: working tree vs pushed tip + analyser testdata/ and
  (N).go dup skips — expected; §7C expects count(DISTINCT path) ≈ 515.
- NEXT: on COMPLETED run §7C verification (rows for page/section/result_spec;
  single commit 36710be; ~515 paths), then §7D (evidence-fed resolver) opens.

## 2026-07-02 night+4 — probe: index healthy at ~5 symbols/sec; in-place commit flip observed

- Probe series (36710be / 4c2c172): 2941/121 @17:44:29 -> 3054/121 -> 3368/121
  @17:45:54 -> 3953/73 -> 4121/73. Rate ≈ 5 rows/sec sustained (427 rows / 85s
  between timestamped polls). Child still EXECUTING_STEP index_symbols
  (since_s 1060 — expected: one long action, no state transitions). Projection:
  total ~4.2–5k symbols at 515 files ⇒ completion minutes away, inside the
  1800s ceiling (deadline ≈ 18:00:06).
- MECHANICS observed from the data (hedged — consistent with, not code-read
  this turn): the old-commit count decays DURING the run (436 -> 121 -> 73)
  because the upsert flips overlapping identities (repo/path/symbol) in place —
  commit_sha moves 4c2c172 -> 36710be as the loop reaches each. The residue at
  4c2c172 when the loop ends = symbols deleted/renamed since 2025-07-14; the
  end-of-run prune (commit IS DISTINCT FROM '36710be') removes them ⇒ final
  state ONE commit.
- NEXT (on COMPLETED): paste index_result counts from the completion response +
  the §7C verification block (expect resolveResultSpec + page/section symbols
  present; ~515 distinct paths; min=max=36710be). Then §7D opens.

## 2026-07-02 night+5 — §7C CLOSED (run 93ba14e6); docs-archive pollution found (§7C.1); runbook rewritten

- COMPLETED both orchestration rows. Final: 4,155 symbols / 499 distinct paths,
  min=max=36710be; prune removed all 4c2c172 rows (single commit). 499 vs 515
  analysed files = 16 symbol-less files (functions:null,types:null — docs.go /
  swagger stubs). Zero-rows query from the morning NOW RETURNS ROWS
  (page_admin_handlers, aggregate_webpage, apply_gap_plan, …). Residual check:
  LIMIT 40 cut alphabetically BEFORE result_spec/save_page_sections/
  plan_sections — targeted confirm query added to runbook (§7C residual box).
- NEW FINDING (§7C.1): the index now contains ARCHIVED code copies —
  docs/agent_docs/…/idea.uk/golang_files/service.go, …/imagery/old/phase1/… .
  Cause: analyse_repo_local calls analysis.Analyse (no excludes) although
  AnalyseWithExclude exists for exactly this (analyse.go header). Fix layered:
  structural = exclude_patterns config (Go default ["docs/"], denylist style,
  consistent with the schema-section decision) → AnalyseWithExclude; PRUNE
  SUBTLETY: re-indexing at the SAME commit cannot remove the polluted rows
  (they already carry 36710be) — but pushing the §7C.1+§7D code creates a new
  commit, and indexing THAT prunes everything at 36710be incl. the pollution;
  interim optional DELETE path LIKE 'docs/%'.
- Runbook REWRITTEN for clarity (user request): completed phases compressed to
  results+pointers; §7D restructured around a worked example (real fuzzy entry
  from run 51f95cda iter-2), plain-language purpose, mechanics, reuse list,
  resolver_top_k config, flagged v.NextScope mutation, tests, combined-build
  deploy plan with §7C.1.
- NEXT: user runs the §7C residual confirm + §7C.1 census (+ optional interim
  delete); then I write §7D resolver + §7C.1 exclude patch as ONE change-set.

## 2026-07-02 night+6 — §7C residual + §7C.1 interim closed; corpus-enrichment policy set

- User worry "did we just delete all the docs?" — NO: the DELETE removed 430
  code_symbols ROWS (index entries + embeddings for the 50 archived .go files
  under docs/). Nothing on disk / in the repo touched; documentation was never
  in this table (analyser indexes .go only); reversible by reindex. Intended
  §7C.1 interim; the exclude patch (with §7D) prevents regrowth.
- §7C residual CONFIRMED: resolveResultSpec (+ResultSpec/ResultMode/
  fallbackDumpInto) at platform/orchestration/result_spec.go; plan_sections_
  action.go 28 symbols (incl. sourceResolver.resolveSpecPath etc. — the exact
  mechanism run 51f95cda's fuzzy entry described); save_page_sections_action.go
  9 symbols. The §6G cause symbol is now retrievable. Live corpus: 3,725
  symbols / 449 files @ 36710be.
- DECISION — corpus-enrichment policy (user asked: comment every function?):
  NO mass annotation. Rationale: only the FIRST doc line embeds; CamelCase
  names + paths already carry signal; the measured failure causes (stale
  corpus, query timing) are fixed/in-flight so RE-MEASURE AT §7E first; stale
  docs = confidently-wrong retrieval, the worst mode for cite-or-abstain.
  Order: (1) mechanical string-literal enrichment of composeSymbolContent
  (log-quote queries match literals near-lexically; no authoring, no rot;
  analyser tweak + reindex, gated on §7E); (2) Go-convention one-sentence
  domain-vocabulary docs on the EXPORTED surface + action entrypoints only
  (census query in runbook sizes it); (3) NO separate tag system — the doc
  first line IS the tag surface.
- NEXT: write §7D resolver + §7C.1 exclude_patterns as ONE change-set.

## 2026-07-03 — §7D resolver + §7C.1 exclude WRITTEN (one change-set)

- Census recorded: 476 / 1,786 exported symbols lack docs (~73% covered) —
  targeted docs job moderate; stays gated on §7E per the enrichment policy.
- Contracts pinned by reading before writing: createRAGEmbeddingClient(ctx,
  config) + embClient.GenerateEmbedding(ctx, text) + applyNomicPrefix(config,
  text, "search_query") (defined elsewhere in package actions — call shapes
  taken from lookup_code_symbols' own call site); vectorSearchCodeSymbols rows
  carry path/symbol/similarity(float64); trigramSearchCodeSymbols exists →
  resolver mirrors the lookup's vector→trigram per-entry fallback; scope
  identity is path+":"+Name (scopeFromCodeResults); AnalysisCallGraph keys
  callsBySym by "path:Name" but covers FUNCTIONS only.
- §7D implementation decisions (all in code comments too):
  (1) resolver_top_k DEFAULT 2 not 3 — substitution nets +K−1 per fuzzy entry
      and guardAfter stops at prevSize+2; K=2 tolerates two fuzzy entries.
  (2) resolver_min_similarity default 0.55 — permissive floor just above the
      measured stale-corpus garbage band (0.547–0.574); §7E calibrates; ALL
      similarities logged; trigram rows (no similarity key) accepted as
      lexical matches.
  (3) Cited-Confirmed skips resolution (loop stops in DecideStep); a
      citation-less Confirmed is item-24-coerced to Unverifiable and
      continues, so it IS resolved.
  (4) Fail-open everywhere: no DB handle → disabled; client/embedding error →
      trigram; search error / all-below-floor → prose label kept (previous
      behaviour, "no worse").
  (5) Exactness sets built ACTION-SIDE from repo_analysis (files + path:Name
      for functions AND types). Considered extending the engine's
      AnalysisCallGraph with Has(): rejected — callsBySym covers functions
      only (types define no calls) and it would force a second engine-file
      copy/deploy for a read-only classification concern.
  (6) DELIBERATE mutations flagged: verdict.NextScope mutated pre-Advance
      (trail records the RESOLVED scope); analysisRaw hoisted (same name,
      wider scope); route header's "no DB" claim amended. NO renames.
- §7C.1: exclude_patterns on analyse_repo_local (Go default ["docs/"];
  configStringSlice reused from diagnose_load_runtime; Analyse →
  AnalyseWithExclude(dir, excludes); excludes logged). Default lives in the Go
  var only (single source of truth), not the Defaults map.
- Tests: diagnose_route_resolver_test.go — exact passthrough (no search
  calls), fuzzy substitution + floor + dedupe, unresolvable stays label,
  search error fail-open, trigram similarity-less acceptance, blank entries,
  knownScopeIdentities parsing (functions+types+null-functions file).
- DEPLOY SEQUENCE: go test ./platform/orchestration/actions/... ; go build
  ./... ; push (creates the new commit) ; deploy ; retrigger index-orchestrator
  at the branch (prune clears 36710be incl. docs regrowth) ; §7E envelope.

## 2026-07-03 — reindex f284b749 fired; order-of-operations gate before §7E

- User triggered index-orchestrator again (19:00:09, correlation f284b749,
  orch 20b33469, REF=083_imagery) and asked whether §6F is next. Answer: §6F
  IS §7E's envelope, but two gates first — and this reindex's own outcome
  depends on whether build/push/deploy preceded the trigger (unverified from
  the trigger output alone).
- Why order matters: the SPAWNED indexer pod runs the DEPLOYED image (§7C.1
  excludes only apply on the new image); the fetch takes the REMOTE TIP (a new
  commit only exists if the §7D/§7C.1 change-set was pushed); the prune keys
  off that commit. Three scenarios; two queries discriminate (commit census +
  docs_rows count) — decision tree recorded in the runbook §7D checklist.
  Notably 36710be + docs_rows>0 means the interim DELETE was undone by an
  old-image same-commit reindex (the documented regrowth caveat) — harmless,
  redo after deploy.
- §7E gates: resolver image deployed; reindex verified; site_id non-empty
  (one-line guard added to the trigger script recommendation). Watch list for
  the run: resolver log lines with similarities (0.55 floor calibration);
  bundle bodies present for resolved entries; static-tier citation; seed
  lookup band re-measure (decides §7F).

## DECISIONS (with rationale)

- **Corpus before resolver (2026-07-02).** The 0-rows result shows the
  cause-relevant symbols are not indexed at all; retrieval quality work on top of
  a missing corpus optimises the wrong layer. Verified the query itself before
  accepting the 0 rows (standing rule).
- **Fix direction: migrate code-indexer's analysis step to analyse_repo_local
  (pending §7A census confirmation).** Reuse over recreate: analyse_repo_local is
  already built, already proven in the diagnose workflow at this very commit
  (fresh tarball + in-process analyse.go, no transport), and index_code_symbols
  was aligned to the same repo-label convention via resolveCodeRepoLabel. Swapping
  the step removes the suspect component (adapter round-trip) in every candidate
  mechanism, rather than patching whichever one the census names. Migration will
  be written against the REAL dumped code-indexer default_config, not from memory.
