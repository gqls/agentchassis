# RUNBOOK — §7 code-retrieval route (make the code channel earn its keep)

Continues from `RUNBOOK.md` (§6 diagnosis loop — complete). Detailed history in
`NOTES_running_synthesis_v4.md` (and v3, archived). Standing rules: user runs
all SQL/kubectl/builds; read outcomes by correlation_id only; snapshot_agent
before agent_definitions UPDATEs; schema before SQL; a 0-rows result is not
decisive until the query itself is checked.

> ## ▶ CURRENT POSITION — 2026-07-02 (after run 93ba14e6)
>
> **§7A–§7C DONE.** The corpus is rebuilt and current: 4,155 symbols / 499
> files at single commit `36710be` (branch 083_imagery), indexed by a SPAWNED
> code-indexer pod carrying the GitHub token. The morning's zero-rows query now
> returns build-domain symbols.
>
> **Open, in order:** §7C.1 (small: archived `docs/` code copies polluted the
> index — exclude + clean), §7D (the evidence-fed resolver — the change this
> route exists for), §7E (observe on a diagnosis run), §7F (deferred).

## Why this route exists (one paragraph)

The diagnosis loop's verdicts name where to look next (`next_scope`), but many
entries are natural-language descriptions, not `path:Symbol` handles. Today
those fuzzy entries pass through the engine verbatim, match nothing in the
call graph, and `assemble_bundle` cannot slice a body for them — the model
keeps asking for code it never receives, so diagnoses are carried entirely by
the runtime/DB channel. Fixing that needed a corpus that actually contains the
code (§7A–§7C, done) and then a resolver that turns the model's descriptions
into real symbols each iteration (§7D).

## §7A–§7C — COMPLETED (results, with pointers)

- **§7A (why the old index was useless):** it was a faithful index of a
  YEAR-OLD tree. Remote `HEAD` = default branch `main`, unmoved since
  2025-07-14 (`git ls-remote origin HEAD` = 4c2c172b…); the §6D trigger sent
  ref:"HEAD" and the adapter honestly fetched it. Decision: merge meaningful
  commits to main manually; envelopes ALWAYS carry an explicit branch/sha.
- **§7B (indexer rebuilt on the in-process path):** code-indexer's analysis
  step swapped to `analyse_repo_local` (migration applied; snapshot 971da9c9)
  with `pin_to_index_commit:false` (the indexer DEFINES the commit — the
  default true would re-fetch the old commit forever), `analysis_field:
  "repo_analysis"` (top-level Output shape), timeout 1800.
- **§7B.1 (token/spawn):** `orchestrate`+agent_type is adopted IN-PLACE on the
  chassis pod, which never holds GITHUB_READ_TOKEN — so `index-orchestrator`
  was seeded (spawn_indexer → call_indexer → complete, mirroring
  diagnose-orchestrator) and `isRepoCloningAgent` gained "code-indexer".
  Verified end to end by run 93ba14e6: spawned pod
  agent-code-indexer-65546fca fetched 083_imagery @ 36710be with the token.
- **§7C (reindex verified):** 4,155 symbols; 499 distinct paths (= 515
  analysed files − 16 files with no symbols); min=max=commit 36710be; prune
  cleared all 436 old rows. Measured indexing rate ≈ 5 symbols/sec through the
  single ollama-adapter (~14 min total).
- [x] **§7C residual CONFIRMED (2026-07-02):** all three targets present —
  `platform/orchestration/result_spec.go:resolveResultSpec` (+ ResultSpec/
  ResultMode/fallbackDumpInto), `plan_sections_action.go` (28 symbols incl.
  PlanSectionsAction + the sourceResolver methods that resolve site_specs
  references), `save_page_sections_action.go` (9 symbols incl.
  SavePageSectionsAction). The §6G cause symbol is retrievable. Query kept for
  re-verification after future reindexes:
  ```sql
  SELECT path, symbol, kind FROM code_symbols
  WHERE repo = 'gqls/agentchassis'
    AND (path ILIKE '%result_spec%' OR path ILIKE '%save_page_sections%'
         OR path ILIKE '%plan_sections%')
  ORDER BY path, symbol;
  ```

## §7C.1 — Index hygiene: archived code copies under docs/ — [ ]

The §7C output shows rows from `docs/agent_docs/…/idea.uk/golang_files/…` and
`…/imagery/old/phase1/…` — archived copies of code, now retrievable symbols.
Cause: `analyse_repo_local` calls `analysis.Analyse(dir)` (no excludes) even
though `AnalyseWithExclude` exists precisely for "trees that carry archived
copies of their own code" (analyse.go header). This pollutes §7D retrieval: a
fuzzy query like "landing page handler" can hit `(*App).opPage` from a dead
archive.

Size it first:
```sql
SELECT CASE WHEN path LIKE 'docs/%' THEN 'docs-archived' ELSE 'live' END AS bucket,
       count(*) AS symbols, count(DISTINCT path) AS files
FROM code_symbols WHERE repo = 'gqls/agentchassis' GROUP BY 1;
```

Fix, two layers:
1. STRUCTURAL (small Go change, rides the §7D build): give `analyse_repo_local`
   an `exclude_patterns` config key (Go default `["docs/"]`, overridable), and
   call `analysis.AnalyseWithExclude(dir, patterns)` instead of `Analyse(dir)`.
   Denylist style, consistent with the schema-section decision: new archive
   locations under docs/ never re-enter.
2. PRUNE SUBTLETY: re-indexing at the SAME commit will NOT remove the docs
   rows (prune is `commit_sha IS DISTINCT FROM $new`; they already carry
   36710be). No special handling needed in practice: pushing the §7C.1+§7D code
   creates a NEW commit — index that, and the prune clears everything at
   36710be including the pollution.
3. INTERIM (optional, so §7D testing isn't polluted meanwhile):
   ```sql
   DELETE FROM code_symbols WHERE repo = 'gqls/agentchassis' AND path LIKE 'docs/%';
   ```
   Safe; regrows only if a reindex runs without the patch.

- [x] census: docs-archived 430 symbols / 50 files vs live 3,725 / 449;
      interim DELETE run (430 rows removed — INDEX ROWS ONLY: entries for
      archived .go files; nothing on disk or in the repo touched, and actual
      documentation was never in this table). Live corpus now 3,725 / 449 @ 36710be.
- [ ] exclude_patterns patch written (with §7D, one build — prevents regrowth)

## §7D — Evidence-fed resolver in diagnose_route — [ ]

**What it is, plainly.** Each iteration the verdict model says where to look
next. Some `next_scope` entries are exact handles the engine can use
(`plan_sections_action.go:PlanSectionsAction`); many are descriptions in
English. The engine and the assembler can only act on exact handles — the
descriptions become inert labels. The resolver translates the descriptions
into exact handles, using the SAME vector search the seed lookup uses, against
the now-current index. The model describes code; retrieval hands it the actual
functions on the next iteration.

**Worked example (real entry from run 51f95cda, iteration 2):**

Verdict returns:
```
next_scope: [
  "plan_sections_action.go:PlanSectionsAction",                          ← exact
  "plan_sections_action or equivalent that resolves site_specs references at build time"   ← fuzzy
]
```
TODAY: entry 1 gets call-graph expansion + a body in the bundle; entry 2 passes
through verbatim, `cg.Neighbourhood` finds nothing, `assemble_bundle` logs
"could not read body" and skips it — the model never sees the code it asked for.

WITH §7D: before the engine runs, `diagnose_route` checks each entry against
the analysis Output. Entry 1 is a known `path:Symbol` → untouched. Entry 2 is
not → its TEXT is embedded (same nomic client, search_query prefix) and
vector-searched against `code_symbols` (top_k = config `resolver_top_k`,
default 3) → it is REPLACED by the hits, e.g.
`platform/orchestration/actions/plan_sections_action.go:resolveSpecReference`
(+2 more). The engine then expands the call graph from REAL symbols, and the
bundle carries their bodies. An entry that resolves to nothing stays as a
label — exactly today's behaviour, so the failure mode is "no worse".

**Where it runs, and why there.**
1. In `diagnose_route_action.go`, AFTER parsing the verdict and BEFORE calling
   `diagnose.Advance`. Before-Advance matters twice over: the call-graph
   expansion and the scope-narrowing guard then operate on real symbols, not
   prose; and the engine (pkg/diagnose) stays pure — resolution needs the DB,
   which only the action has.
2. "Is it exact?" uses the same identity `assemble_bundle` slices by: the
   entry, split at the last colon, names a file present in the analysis Output
   (and, if a symbol part exists, a symbol in that file).

**Reuse (no new retrieval machinery):** `resolveCodeRepoLabel` (repo label),
`createRAGEmbeddingClient` (embedding), `vectorSearchCodeSymbols` (search) —
all already in `code_symbols_actions.go` and proven by the seed lookup.

**Config:** two new optional keys on the route step, Go-defaulted (no
migration): `resolver_top_k` (default **2**, not 3 — substitution nets +K−1
entries per fuzzy item and the narrowing guard stops at prevSize+2, so K=2
keeps two fuzzy entries per verdict inside the guard; 0 disables) and
`resolver_min_similarity` (default **0.55**, a permissive floor just above the
measured stale-corpus garbage band; every similarity is logged so §7E can
calibrate it; trigram-fallback rows carry no similarity and are accepted as
lexical matches).

**FLAGGED deliberate change:** the resolver mutates `v.NextScope` before
`Advance`, so the evidence trail records the RESOLVED scope rather than the
model's prose. That is the more auditable record (you can see exactly which
symbols each iteration examined), but it is a change to the trail's contents —
noted here and in the code comment.

**Tests (diagnose_route side):**
- exact entry passes through untouched;
- fuzzy entry replaced by fake-DB hits;
- unresolvable entry survives as a label;
- empty `next_scope` unchanged;
- `resolver_top_k: 0` disables cleanly.

**Deploy:** one build together with the §7C.1 exclude patch → push (creates
the new commit) → deploy → retrigger the indexer at the branch (clears the
docs pollution via prune) → §7E.

- [x] code WRITTEN 2026-07-02 (gofmt-clean; chassis build/tests user-side):
      `diagnose_route_action.go` — resolver inserted as step 3.5 (after the
      call-graph build, before Advance), `analysisRaw` hoisted (same name,
      wider scope — no rename), header's "no DB" claim amended, resolver
      machinery + configFloatField appended; REUSES resolveCodeRepoLabel /
      createRAGEmbeddingClient / applyNomicPrefix / vectorSearchCodeSymbols /
      trigramSearchCodeSymbols / truncateForLog. Cited-Confirmed verdicts skip
      resolution (a citation-less Confirmed is coerced and continues, so it IS
      resolved). Exactness sets built action-side from repo_analysis
      (functions AND types) — considered and rejected extending the engine's
      AnalysisCallGraph (callsBySym covers functions only; avoids a second
      engine-file copy). Tests: diagnose_route_resolver_test.go (7 cases).
- [x] §7C.1 patch in the same change-set: analyse_repo_local gains
      exclude_patterns (Go default ["docs/"], configStringSlice REUSED),
      Analyse → AnalyseWithExclude, excludes logged.
- [ ] user: `go test ./platform/orchestration/actions/...` then
      `go build ./...`; push (the new commit), deploy (note tag)
- [ ] retrigger index-orchestrator at the branch — the new commit's prune
      clears 36710be rows (incl. any docs regrowth)
- [ ] then §7E diagnosis run

## §7E — Observe: does the code channel now contribute — [ ]

Re-run the §6F diagnosis envelope (NEW correlation_id; site_id non-empty; ref
handling unchanged — diagnose pins to the index commit automatically).
Success criteria:
- [ ] at least one iteration's scope contains a RESOLVER-substituted symbol
      whose body appears in the bundle (no could-not-read-body for it)
- [ ] at least one citation with tier=static from such a body
- [ ] seed lookup hits are no longer the flat generic band (corpus is current;
      re-measure the similarity spread)
- [ ] runtime/DB channel un-regressed (data_requests still run; outcome grounded)

## §7F — DEFERRED: seed-query improvement

Reorder `lookup_symbols` after `load_runtime`; build the seed query from the
symptom + salient error-log lines. Revisit only after §7E — the resolver may
make this unnecessary.

## Corpus enrichment (question raised 2026-07-02 — measure first, then in this order)

Should every function carry a human description for embedding-match? NO — the
embedded text is name + signature + FIRST doc line + path, so only one good
sentence counts; names/paths already carry signal; and stale docs make
retrieval confidently wrong (worst failure mode for a cite-or-abstain loop).
Order of investment, gated on the §7E measurement:
1. MECHANICAL, no authoring, no rot: extend composeSymbolContent with a
   function's string literals (capped) — diagnosis queries quote log lines and
   the literals ARE the log lines. Analyser tweak + reindex.
2. Go-convention docs on the EXPORTED surface + action entrypoints/handlers
   only (the few hundred symbols a diagnosis plausibly names) — one precise,
   domain-vocabulary first sentence. Size it first:
   ```sql
   SELECT count(*) FILTER (WHERE coalesce(doc,'') = '') AS exported_no_doc,
          count(*) AS exported_total
   FROM code_symbols
   WHERE repo = 'gqls/agentchassis'
     AND regexp_replace(symbol, '^.*\.', '') ~ '^[A-Z]';
   ```
3. NO separate tag system — the doc first line is the tag surface; parallel
   machinery violates reuse/simplicity.

## Parking lot (not this route)

- Trigger intermittently sends site_id="" (diagnosis envelope).
- Verdict quality: stale revised-hypothesis text carried into a CONFIRMED
  conclusion; data_requests emitted in a terminal verdict never run.
- Ref Structural A: CI-triggered indexing with ${GITHUB_SHA} post-deploy
  (runners are in-cluster) — removes manual triggers entirely.
- index-orchestrator row: category='diagnose' copied from the template row —
  cosmetic tidy.
- Analyser-adapter's own stale behaviour is moot for indexing now; check
  before any OTHER consumer relies on it.
- Stale agent_definitions columns (task_workflow / orchestrator_workflow /
  orchestration_workflow) — null out.
