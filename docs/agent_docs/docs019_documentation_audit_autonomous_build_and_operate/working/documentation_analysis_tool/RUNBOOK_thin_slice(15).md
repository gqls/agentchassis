# Runbook — thin-slice bundle pipeline

How to run the thin slice today. The tools are a small Go module, `contextkit/`, with two shared contracts and a command per tool:

```
contextkit/
  go.mod
  internal/analysis/     the analyser's output shape — defined once, consumed by analyser/assembler/embed/resolve_targets
  internal/candidates/   the ranked-candidate shape — emitted by resolve_targets/embed/fuse, read by fuse/eval_targets
  cmd/analyser/  cmd/assembler/  cmd/embed/  cmd/dbcontext/  cmd/resolve_targets/  cmd/fuse/  cmd/eval_targets/
```

Build everything with `go build ./...`; run a tool with `go run ./cmd/<name>`. Each contract is defined once in its `internal/` package and referenced by qualified name (`analysis.FuncDef`, `candidates.Candidate`) — no per-tool copies or aliases. `thin_slice_constitution.md` and `groundtruth_targets.json` sit alongside (pass them by path). A Go toolchain is needed (and, for `dbcontext`, `psql` access to the DB; for `embed`'s real recall, reach to your Ollama).

---

## The pipeline

```
# 1. Analyse the repo once (re-run when code changes meaningfully).
go run ./cmd/analyser /path/to/repo > analysis.json

# 1b. (Optional) Propose what to scope for a task — lexical first cut; confirm before use.
go run ./cmd/resolve_targets -analysis analysis.json \
  -task "why does plan_sections see no ready sections" -n 12

# 1c. (Optional) Semantic recall layer — build the index once, then query per task.
#     Real recall needs your Ollama embedding model; -local is an offline stand-in (NOT semantic).
go run ./cmd/embed build -analysis analysis.json -out embeddings.json \
  -ollama http://ollama-adapter.ai-persona-system.svc.cluster.local:11434 -model nomic-embed-text
go run ./cmd/embed query -embeddings embeddings.json -n 12 \
  -ollama http://ollama-adapter...:11434 -model nomic-embed-text \
  -task "what source does plan_sections read for readiness and how is ready_count computed"
#     Union the resolve_targets (lexical) and embed (semantic) candidate sets for the -scope.

# 1d. (Optional) Merge the two into one ranking (RRF), and measure against ground truth.
go run ./cmd/resolve_targets -analysis analysis.json -task "T" -n 25 -json > lex.json
go run ./cmd/embed query -embeddings embeddings.json -task "T" -n 25 -json -ollama ... -model ... > sem.json
go run ./cmd/fuse -in lex.json -in sem.json -n 12 -json > fused.json   # combined -scope proposal
go run ./cmd/fuse -in lex.json -in sem.json -n 12                       # same, human-readable
go run ./cmd/eval_targets -truth groundtruth_targets.json -candidates fused.json -n 12  # recall@N, MRR
#     Run eval on lex.json / sem.json / fused.json for the same task to compare the three.
#     Caution: plain RRF rewards agreement, so fusing a strong list with a weak one degrades it —
#     only trust the fused/semantic numbers once sem.json comes from the real Ollama model, not -local.

# 2. (Optional) Pull live DB context — schema, rows, and/or runtime evidence.
go run ./cmd/dbcontext -psql 'psql "postgresql://user:pass@host/clients_db"' \
  -schema site_work_items,pages,site_plan_sections > schema.md
go run ./cmd/dbcontext -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -rows "SELECT name, build_status FROM pages WHERE site_id='<id>'" > dbrows.md
# runtime evidence (recent errors + work-item lifecycle) for a debug task:
go run ./cmd/dbcontext -psql '…same…' -runtime-site gamesdesign.co.uk -runtime-page guide-skinner-box > runtime.md

# 3. Assemble a bundle for one task.
go run ./cmd/assembler \
  -analysis analysis.json \
  -root /path/to/repo \
  -constitution thin_slice_constitution.md \
  -task "one-sentence task" \
  -step implementation \
  -scope internal/foo/handler.go:HandleBuild \
  -scope internal/foo/store.go \
  -include internal/foo/registry.go \
  -doc docs/016_debugging_guide.md \
  -schema schema.md \
  -runtime runtime.md \
  > bundle.md
```

Then paste `bundle.md` into a chat (or feed it to the API — see "large bundles" below).

### assembler flags

- `-analysis` — the analyser JSON. Required.
- `-root` — repo root, so the assembler can read full bodies. Required.
- `-constitution` — the flat constitution file. Required.
- `-task` — one sentence. Required.
- `-scope` — repeatable. A whole file (`path.go`) or one symbol (`path.go:Name`). At least one required.
- `-step` — `framing` | `implementation` | `debug`.
  - `implementation` / `debug`: in-scope code shown in full.
  - `framing`: in-scope shown as signatures (intent over detail) — use when the brief still needs expanding.
  - `debug`: adds a runtime-evidence placeholder (no run trace in the thin slice yet).
- `-neighbour` — `callgraph` (default) | `package`. Call-graph shows only the callees/callers/types of the in-scope symbols (tight); package shows the whole in-scope package as signatures (the fallback when the call graph misses something, e.g. interface dispatch).
- `-doc` — repeatable. Any authored doc to paste in verbatim: a debug guide for a `debug` task, a 003 contract section, or `dbcontext` row output. The manual stand-in for the matched-standards layer the real tool will pull in automatically.
- `-include` — repeatable. A wiring/shared file to force-include as signatures regardless of the call graph (e.g. `registry.go`, which registers actions via init and so isn't reached by following calls). Closes the blind spot the first adoption run found.
- `-schema` — a text/markdown file of schema (from `dbcontext -schema`, or hand-pasted `\d`).
- `-runtime` — a runtime-evidence file (from `dbcontext -runtime-site`) for the Runtime evidence section; without it a `debug` bundle shows a "generate it with dbcontext" placeholder.
- `-max-neighbour` — cap on neighbourhood signatures per group (default 60).

### dbcontext flags

- `-psql` — the psql invocation: a direct connection (`psql "postgresql://…"`) or your kubectl pattern (`kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db`). The query is appended as `-c` args, not via a shell.
- `-schema` — comma-separated tables; runs `\d` for each (complete, bounded).
- `-rows` — a SELECT; fetched with multipass sizing (probe `LIMIT N+1`, then all rows if within the cap, else a sample + the query as a pointer). Never an unbounded dump.
- `-runtime-site` — a domain; pulls **runtime evidence** for it: recent `agent_error_log` rows and the `site_work_items` lifecycle, as a "Runtime evidence" block. `-runtime-page` narrows the work-items to one page. Feed the output to the assembler via `-runtime`.
- `-max-rows` — the row cap (default 50).

---

## Large bundles — how to feed them to Claude

A bundle's size matters, and bigger is not better. Rough conversion: ~4 characters per token, so ~1.5 MB of text ≈ ~375K tokens and ~4 MB ≈ ~1M tokens.

**Context-window facts (verified against the Claude docs, June 2026):**
- Every current Claude model has a **200K-token** context window.
- **Opus 4.x and Sonnet 4.x support a 1M-token window in beta**, gated behind the `context-1m-2025-08-07` beta header and restricted to higher usage tiers (≈ tier 4 / custom limits).
- Requests **over 200K tokens are billed at premium rates** (≈ 2× input, 1.5× output).
- **Context rot is real:** models do not use a full 1M window evenly — accuracy and recall degrade as the window fills. "It fit" is not "it was used well."

**What this means for us:**
- A ~1.5 MB bundle (~375K tokens) already exceeds the standard 200K window — it needs the 1M beta and is premium-priced. A 4 MB blob (~1M tokens) sits at the ceiling and is squarely in the context-rot zone.
- So **a large bundle is a smell, not a goal.** The fix is to include less, not to buy a bigger window. This is the main reason to build the call graph next (below): include the functions actually involved, not whole packages/directories. It improves *quality* (less rot), not just cost.

**Three ways to feed a bundle:**
1. **Paste into a chat** — fine for small bundles (well under the window). Simplest.
2. **A Project (claude.ai)** — attach the bundle (or the larger compiled context) as project knowledge. This is the home for the bigger "project files" scale (the 4 MB compiled context goes here, not in a single chat message).
3. **The API** — send the bundle as message content (or document blocks). For a large, stable bundle reused across a session, use **prompt caching**: cache writes cost ~25% over base input, cache reads ~10% — so a big stable prefix becomes nearly free to reuse turn to turn. Cache the stable parts (constitution, reference docs, code context); leave the task/question uncached. For >200K-token bundles, add the 1M beta header (and you need the tier for it).

Rule of thumb: aim to keep a working bundle **under ~200K tokens (~800 KB of text)** so it fits the standard window and stays in the zone where the model actually uses it. If you can't, that's the signal the selection is too broad.

---

## Fuzzy tasks need the brief expanded before targets can be picked

Some tasks (e.g. "improve site imagery") are under-specified — you can't pick in-scope files until the brief is consolidated into something concrete. For those, the first pass is a **framing bundle**: `-step framing`, scope the relevant area at signature level, `-doc` the relevant existing plans/assessments, and use it to expand the brief into a spec. Once the spec names the real targets, assemble an `implementation` bundle. (This is the framing-vs-implementation altitude split, done by hand for now.)

---

## First real run — what to observe

The thin slice has been reasoned about but not yet run on a real task and checked. The first run is the actual test, and four observations matter more than the bundle looking tidy. When you produce a bundle for a task you'd otherwise hand-curate, note:

1. **Neighbourhood adequacy.** Did the call-graph slice pull the surrounding code you actually needed, or did it miss something (interface dispatch is the usual gap)? If it missed, does `-neighbour package` recover it? This tells you the call graph's real-world limit.
2. **Scope delta.** Compared to what you'd reach for by hand, did you add or drop any `-scope` files? That delta is the target-resolution signal — it's what an automatic "find the in-scope set" step would have to reproduce.
3. **Schema vs live rows.** Was the schema enough, or did you want live table content too? That tells you whether `dbcontext -rows` belongs in this class of bundle by default.
4. **Size.** Is the bundle a sane size, or already bloated? Bloat means the scope is too broad — the signal to narrow, not to widen the window.

These four say where to invest next far better than more design does.

## Known limits of the thin slice (so the test is judged fairly)

- **Scope is manual.** You name the files/symbols; there's no automatic "given this task, find the target" yet.
- **Call-graph neighbourhood is name-matched**, not type-resolved: a name shared across packages can show extra candidates, and interface dispatch isn't followed. Use `-neighbour package` when it misses something, and `-include` for wiring/registration files (like `registry.go`) that it can't reach by following calls.
- **DB context is on-demand, not automatic.** `dbcontext` pulls schema and rows, but you choose the tables/queries by hand and run it as a separate step; the assembler doesn't yet decide which tables a task touches.
- **No runtime evidence.** `debug` shows a placeholder; there's no `orchestration_id`-correlated run trace, error trail, or pod logs yet.
- **Analyses the on-disk tree.** If the authoritative code lives in the running system and the checkout/snapshot is stale, the bundle ships stale code. Analyse the current source; the analysis records no commit/SHA yet, so staleness isn't visible — watch for it.
- **No bundle record.** Nothing logs what went in — fine while you're pasting it by hand and can see it; it returns as provenance once assembly is automated.

---

## Next improvements (in likely order)

1. **Populate the target-resolution comparison with the real model, and grow the ground-truth set.** The merge (`fuse.go`, RRF) and the scorer (`eval_targets.go`, recall@N / MRR against `groundtruth_targets.json`) are built and run end-to-end; what's missing is the *real* semantic list — the offline stand-in is non-semantic noise, and the eval confirms that fusing it with the lexical list degrades the result (RRF rewards agreement). So: run `embed query -ollama` for the seeded tasks, feed the resulting `sem.json` to `fuse` and `eval_targets`, and read off recall@N for lexical vs semantic vs fused. Then add a task entry per known case so the numbers mean something. If the real numbers show one list dominating, weight the fusion rather than using plain RRF.
2. **Matched-guidelines retrieval** — find the relevant docs/standards for a task automatically instead of naming them by hand with `-doc`. Same retrieval problem as target resolution, for documentation.
3. **Build-time review contributors (with revise/HITL)** — liability and **morality** review for output that could create legal exposure or be ethically wrong. The morality review applies a *configured, layered* standard (a chosen base framework + operator judgement + jurisdiction/current-focus, prioritised), and routes contested calls to a human rather than auto-resolving. (These are contributors to the build, distinct from improvement-loop *checkers*, which monitor deployed sites against plan/spec — a separate concept, to investigate later.)
4. **Code-freshness guard** — record the commit/SHA the analysis was taken from, so a stale checkout is visible rather than silent.

*(Done: call-graph neighbourhood; live schema + row data via `dbcontext` with multipass sizing; `-include` force-include for wiring/shared files; runtime evidence via `dbcontext -runtime-site` + assembler `-runtime`; a lexical target-resolution baseline (`resolve_targets.go`); a semantic vector index (`embed.go`, Ollama-backed, with an offline stand-in for testing); rank-fusion of candidate lists (`fuse.go`, RRF) and a recall@N/MRR scorer against a ground-truth set (`eval_targets.go` + `groundtruth_targets.json`).)*

## B4a — measuring code-domain embedding quality (the one open risk)

The question: does `nomic-embed-text` retrieve *code* well enough to beat the lexical
(trigram) baseline — and if not, which model does? The harness embedder now matches
the chassis exactly (single-call `/api/embeddings`, same model + `search_document:` /
`search_query:` prefixes as `rag_index`/`rag_lookup`), so these numbers predict
production. You run this; it needs the chassis repo and the live adapter.

Adapter URL (from `ollama.go`): `http://ollama-adapter.ai-persona-system.svc.cluster.local:11434`
— reachable in-cluster. From outside: `kubectl -n ai-persona-system port-forward svc/ollama-adapter 11434:11434` and use `http://localhost:11434`.

```
# 0. Analyse the chassis repo once.
go run ./cmd/analyser /path/to/agent-chassis > chassis.json

# 1. Lexical baseline for each ground-truth task (no model needed).
go run ./cmd/resolve_targets -analysis chassis.json -task "<task>" -json > lex.json
go run ./cmd/eval_targets -candidates lex.json -truth groundtruth_targets.json

# 2. Semantic, against the REAL adapter. Index once (time this — it is the realistic
#    single-call indexing latency, the SPEED half of B4a), then query per task.
time go run ./cmd/embed build -analysis chassis.json -out sem_index.json \
     -ollama http://localhost:11434 -model nomic-embed-text
go run ./cmd/embed query -embeddings sem_index.json -task "<task>" -n 12 \
     -ollama http://localhost:11434 -model nomic-embed-text -json > sem.json
go run ./cmd/eval_targets -candidates sem.json -truth groundtruth_targets.json

# 3. Hybrid, then eval the same task across all three.
go run ./cmd/fuse -in lex.json -in sem.json > fused.json
go run ./cmd/eval_targets -candidates fused.json -truth groundtruth_targets.json
```

Read the recall@N / MRR across lex / sem / fused. Decision rule:
- semantic or fused beats lexical by a clear margin on the code ground truth → embeddings
  earn their place in the code path as designed;
- it doesn't → lexical (trigram + `resolve_targets`) carries the spine and embeddings are
  the tie-breaker. Either way the `code_symbols` table already supports both.

"Which model for code" (the B4b consequence): pull a code-specific embedding model into
the adapter, then repeat step 2 with `-model <code-model>`. The index and query MUST use
the same model (same vector space), so rebuild `sem_index.json` per model. Compare recall
to nomic on the same ground truth. This only touches the code path — the knowledge base
stays on nomic throughout.

Caveats: the seed `groundtruth_targets.json` is a first signal — grow the task→expected-symbol
pairs before trusting the numbers. And per the fusion note above, trust `fused`/`sem` only
once `sem.json` is from the real adapter, never the `-local` stand-in.

## In-cluster path — deploy the analyser adapter and run the first index (2026-06-11)

The CLI pipeline above stays the offline harness; this section is the in-cluster
equivalent: analyser adapter → code-indexer agent → code_symbols → lookup. All
artifacts are drafted (destinations: contextkit/README.md map). User runs every
step.

```bash
# ── 0. Copy the drafts into the repo (run from the repo root; <bundle> = where
#      you saved chassis-drafts/analyser-adapter) ──
mkdir -p internal/adapters/analyser cmd/analyser-adapter
cp <bundle>/adapter.go <bundle>/analyse_action.go <bundle>/github_source.go internal/adapters/analyser/
cp <bundle>/cmd/analyser-adapter/main.go cmd/analyser-adapter/
cp <bundle>/code_symbols_actions.go <bundle>/analyser_request_action.go platform/orchestration/actions/
cp <bundle>/configs/analyser-adapter.yaml configs/
cp -r <bundle>/deployments/kustomize/services/analyser-adapter deployments/kustomize/services/
cp <bundle>/deployments/kafka/analyser-requests-topic.yaml deployments/kafka/  # or where thunder's topic CRDs live
# Makefile: replace with the updated workspace copy AFTER a diff (4 insertions only):
#   diff makefile <downloaded-updated-makefile>
# internal/analysis already exists in the repo — confirm it matches contextkit's copy:
diff -r internal/analysis <contextkit>/internal/analysis   # expect no output

# ── 1. Pre-deploy gates (order matters; each is a known failure mode) ──
# 1a. Read-only GitHub secret (fine-grained PAT, contents:read, chassis repo only):
kubectl -n ai-persona-system create secret generic analyser-github-read \
  --from-literal=token=REDACTED   # or via sealed/external-secrets
kubectl -n ai-persona-system get secret analyser-github-read   # verify it exists
# 1b. Requests topic — auto-create is OFF; without this the consumer hangs
#     ("context deadline exceeded" on every fetch, no clear error):
kubectl apply -f deployments/kafka/analyser-requests-topic.yaml
kubectl -n kafka get kafkatopic | grep analyser   # expect: ...analyser.requests  Ready=True

# 1c. REGISTRY PASTE — platform/orchestration/actions/registry.go
#   i.   Open registry.go; find the "training_data_export" entry; place the
#        cursor after its closing "}," (the FEED banner comment follows).
#   ii.  Paste the whole block from registry_insertions.md (the CODE SYMBOLS
#        banner + three entries) there.
#   iii. CATEGORY DECISION: request_repo_analysis ships Category "code" — a
#        NEW category. To reuse the existing taxonomy instead, change that one
#        word to "analysis". Either works; pick once.
#   iv.  Compile check (catches a bad paste before the release does):
go build ./platform/... ./internal/adapters/analyser/... ./cmd/analyser-adapter/

# 1c'. DOCKERFILE COPY-SWAP (the one undrafted file):
cp build/docker/backend/thunder-adapter.dockerfile build/docker/backend/analyser-adapter.dockerfile
sed -i 's/thunder-adapter/analyser-adapter/g' build/docker/backend/analyser-adapter.dockerfile
grep -in thunder build/docker/backend/analyser-adapter.dockerfile   # expect NOTHING
# Read the file once end-to-end: the go build line should target
# ./cmd/analyser-adapter and the CMD should run ./analyser-adapter. If any
# thunder-ONLY lines remain (ssh assets etc.), stop and flag before building.
make -n build-analyser-adapter   # dry-run: prints docker build with the new -f path

# 1d. Agent INSERT — DONE 2026-06-12 (INSERT 0 1; domain_tags jsonb cast fixed
#     after the first attempt failed on an ARRAY[] literal). Re-run guard:
#     UNIQUE (type, version) — revise via snapshot_agent + UPDATE, never re-INSERT.
SELECT type, version, status FROM agent_definitions WHERE type = 'code-indexer';

# 1e. KUSTOMIZE DRY-BUILD (30 seconds, catches wiring before the cluster does):
kubectl kustomize deployments/kustomize/services/analyser-adapter/overlays/production/uk_001
# EXPECT three objects, all namespace ai-persona-system:
#   ConfigMap  analyser-adapter-config-<hash>   (generated from the overlay's configs/)
#   Service    analyser-adapter
#   Deployment analyser-adapter — image docker.io/aqls/analyser-adapter:v1.0.1060
#              (deploy-agents seds the tag at release) and the config volume's
#              configMap name carrying the SAME <hash> as the ConfigMap.
# Failure modes seen:
#   - "evalsymlink failure on .../overlays/production/uk_001/configs: no such
#     file or directory" (loading KV pairs) → the overlay's configs/ SUBDIR did
#     not survive the copy into the repo. Fix:
#         mkdir -p deployments/kustomize/services/analyser-adapter/overlays/production/uk_001/configs
#         cp <bundle>/.../overlays/production/uk_001/configs/analyser-adapter.yaml \
#            deployments/kustomize/services/analyser-adapter/overlays/production/uk_001/configs/
#     A selective cp drops nested subdirs — cp -r the whole analyser-adapter
#     kustomize tree to avoid this.
#   - "Warning: 'patchesStrategicMerge' is deprecated" → non-fatal; the overlay
#     uses the modern `patches: [{path: deployment-env.yaml}]` form.
#   - "accumulating resources" path error → the kustomize tree was copied at the
#     wrong depth (must sit at
#     deployments/kustomize/services/analyser-adapter/{base,overlays/...}).

# ── 2. Release ──
make release-backend IMAGE_TAG=vX.Y.Z   # build-adapters now includes analyser

# ── 3. Post-deploy verification ──
kubectl -n ai-persona-system get pods -l app=analyser-adapter        # 1/1 Running
kubectl -n ai-persona-system logs deploy/analyser-adapter --tail=30  # "Analyser adapter initialized" then "starting message processing"
kubectl -n ai-persona-system port-forward deploy/analyser-adapter 8081:8080 &
curl -s localhost:8081/health; curl -s localhost:8081/ready
# kcat smoke (envelope: action in the BODY — 035 §1.2):
#   body: {"action":"analyse","reply_to_topic":"system.<x>.smoke.responses",
#          "data":{"owner":"gqls","repo":"agentchassis","ref":"HEAD"}}
# Expect a reply with in_response_to_request_id=<request_id>, status=complete,
# real-bool is_complete in the body headers.

# ── 4. First indexing run ──
# Trigger the code-indexer via the generic entry point with
# input_data {owner, repo, ref}. Watch BOTH logs:
#   adapter:   "analysed repo" (file_count) → "analyse response sent"
#   indexer:   request_repo_analysis → await → index_code_symbols counts
# RUNTIME CONFIRMATION (the standing caveat): if index_code_symbols errors with
# "no analysis output at repo_analysis.output", inspect the awaited reply's
# nesting in collected_data and adjust the step's analysis_field/commit_field/
# repo_field configs (ExtractNestedField tolerates a .response wrapper; the
# payload key is "data" in git/thunder vs "body" in the canonical type).

# ── 5. Verify code_symbols (schema-check discipline applies to any new SQL) ──
# SELECT count(*), count(embedding) AS embedded, count(DISTINCT path) AS files
#   FROM code_symbols WHERE repo = 'gqls/agentchassis';
# SELECT kind, count(*) FROM code_symbols WHERE repo='gqls/agentchassis' GROUP BY kind;
# A 0-rows result: check the WHERE repo value against what the indexer logged
# BEFORE concluding the index failed (the repo-naming convention below).

# ── 6. B4a, now against the LIVE path ──
# sem.json from lookup_code_symbols (or embed -query against the adapter) —
# the real nomic model, never -local. Then eval lex/sem/fused per §B4a above.
```

**Label convention — DECIDED 2026-06-11:** `code_symbols.repo` is the
`owner/repo` form (`gqls/agentchassis`), COMPOSED by `index_code_symbols` from
the analyser reply's owner+repo (so the label always matches what was fetched;
no trigger-supplied label to mismatch). The agent SQL sets no repo config on
the index step. Every lookup filters the same form. Non-git corpora later
override via the step's `config.repo` (e.g. `domain:kruste.com`); tenancy, when
it comes, is an explicit `client_id` column, not a label prefix.

## Tool-doc header rollout (2026-06-11) — apply order is load-bearing

All artifacts in chassis-drafts/tool-docs/ (destinations in each file header).
Three stages; do not reorder — the gate without the prompt fails every
generation, and the stamps without the columns fail every insert.

```bash
# ── 1. Provenance columns (clients_db) ──
#    Pre-checks are in the file header (source_% absent; system_events trail).
psql ... -f NNN_add_component_provenance.sql

# ── 2. Prompts (clients_db) ──
#    \df snapshot_agent first. Anchored + idempotent; ABORTS if the prompts
#    drifted since the 2026-06-11 dump — re-pull the rows and re-anchor.
psql ... -f NNN_update_tool_prompts_doc_header.sql
#    Expect NOTICEs "applied" ×2 and the VERIFY query showing t / t.

# ── 3. One binary release with the six Go files ──
#    tool_doc_header.go (NEW, platform/content/), create_tool_component_action.go,
#    deploy_tool_action.go, rerender_single_page_action.go,
#    rerender_pages_actions.go, check_tool_health.go (+ the two 019 splices
#    into the canonical doc).
make release-backend IMAGE_TAG=vX.Y.Z

# ── 4. End-to-end verification: generate ONE tool ──
#    (trigger an add_tool item with tool_component_id null, or run the
#    tool-generator entry point directly)
# a. header present in the DB template:
#    SELECT position('=== tool-doc ===' IN html_template) > 0
#    FROM content_components WHERE component_level='tool'
#    ORDER BY created_at DESC LIMIT 1;
# b. provenance stamped:
#    SELECT source_agent_type, source_orchestration_id FROM content_components
#    WHERE component_level='tool' ORDER BY created_at DESC LIMIT 1;
# c. header ABSENT from what shipped: open the committed page (and the
#    /tools/assets/{function}.js if the component has js_content) in the git
#    repo / on the CDN — grep for 'tool-doc' must find nothing.
# d. tool_health quiet on it: next sweep creates no no_doc_header item for it.
# A failed gate ("generated tool HTML lacks the tool-doc header") right after
# stage 3 means stage 2 didn't apply — check the NOTICEs before blaming the LLM.
```

Old tools converge via the no_doc_header WARNING on the normal sweep — no
retrofit campaign, no bulk regeneration.
