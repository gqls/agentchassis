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
