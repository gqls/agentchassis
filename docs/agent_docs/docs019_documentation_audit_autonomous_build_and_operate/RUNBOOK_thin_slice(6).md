# Runbook — thin-slice bundle pipeline

How to run the thin slice today. Files: `analyser.go`, `assembler.go`, `dbcontext.go`, `resolve_targets.go`, `embed.go`, `thin_slice_constitution.md`. Go toolchain on the machine with the repo (and, for `dbcontext`, `psql` access to the DB; for `embed`'s real recall, reach to your Ollama).

---

## The pipeline

```
# 1. Analyse the repo once (re-run when code changes meaningfully).
go run analyser.go /path/to/repo > analysis.json

# 1b. (Optional) Propose what to scope for a task — lexical first cut; confirm before use.
go run resolve_targets.go -analysis analysis.json \
  -task "why does plan_sections see no ready sections" -n 12

# 1c. (Optional) Semantic recall layer — build the index once, then query per task.
#     Real recall needs your Ollama embedding model; -local is an offline stand-in (NOT semantic).
go run embed.go build -analysis analysis.json -out embeddings.json \
  -ollama http://ollama-adapter.ai-persona-system.svc.cluster.local:11434 -model nomic-embed-text
go run embed.go query -embeddings embeddings.json -n 12 \
  -ollama http://ollama-adapter...:11434 -model nomic-embed-text \
  -task "what source does plan_sections read for readiness and how is ready_count computed"
#     Union the resolve_targets (lexical) and embed (semantic) candidate sets for the -scope.

# 2. (Optional) Pull live DB context — schema, rows, and/or runtime evidence.
go run dbcontext.go -psql 'psql "postgresql://user:pass@host/clients_db"' \
  -schema site_work_items,pages,site_plan_sections > schema.md
go run dbcontext.go -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -rows "SELECT name, build_status FROM pages WHERE site_id='<id>'" > dbrows.md
# runtime evidence (recent errors + work-item lifecycle) for a debug task:
go run dbcontext.go -psql '…same…' -runtime-site gamesdesign.co.uk -runtime-page guide-skinner-box > runtime.md

# 3. Assemble a bundle for one task.
go run assembler.go \
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

1. **Target resolution — merge lexical + semantic, then measure.** The lexical baseline (`resolve_targets.go`) and the semantic index (`embed.go`, Ollama-backed) both exist; the remaining work is (a) merging the two candidate sets into one ranked `-scope` proposal and (b) assembling a small ground-truth task→files set so the semantic layer's recall gain over lexical is *measured*, not assumed (the `LoadPageSectionsFromSpecAction` miss is the first case to check). Real recall needs a run against the Ollama model; the offline stand-in only proves the pipeline.
2. **Matched-guidelines retrieval** — find the relevant docs/standards for a task automatically instead of naming them by hand with `-doc`. Same retrieval problem as target resolution, for documentation.
3. **Build-time review contributors (with revise/HITL)** — liability and **morality** review for output that could create legal exposure or be ethically wrong. The morality review applies a *configured, layered* standard (a chosen base framework + operator judgement + jurisdiction/current-focus, prioritised), and routes contested calls to a human rather than auto-resolving. (These are contributors to the build, distinct from improvement-loop *checkers*, which monitor deployed sites against plan/spec — a separate concept, to investigate later.)
4. **Code-freshness guard** — record the commit/SHA the analysis was taken from, so a stale checkout is visible rather than silent.

*(Done: call-graph neighbourhood; live schema + row data via `dbcontext` with multipass sizing; `-include` force-include for wiring/shared files; runtime evidence via `dbcontext -runtime-site` + assembler `-runtime`; a lexical target-resolution baseline (`resolve_targets.go`); a semantic vector index (`embed.go`, Ollama-backed, with an offline stand-in for testing).)*
