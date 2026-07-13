# RUNBOOK — §7 code-retrieval route (make the code channel earn its keep)

Continues from `RUNBOOK.md` (§6 diagnosis loop — complete). Notes live in
`NOTES_running_synthesis_v4.md`. Standing rules apply throughout: user runs all
SQL/kubectl/builds; read outcomes by correlation_id, never ORDER BY created_at;
snapshot_agent before any agent_definitions UPDATE; schema before SQL; 0 rows is
not decisive until the query is checked.

> ## ▶ CURRENT POSITION — 2026-07-02 (read this first)
>
> **The corpus is the blocker, not the query.** `code_symbols` (436 rows,
> gqls/agentchassis @ 4c2c172) contains NO page/section/result_spec symbols —
> verified by a sound query — while the tree at that commit demonstrably contains
> those files (the deployed binary logs from `orchestration/result_spec.go:185`).
> The analyser (`internal/analysis`) has no caps or skips that would drop them
> (read in full). The §6D writer — `code-indexer`, workflow
> `request_repo_analysis → await analyser → index_code_symbols` — crossed Kafka
> via the analyser ADAPTER; that path is the remaining suspect. §7A discriminates
> the mechanism; §7B removes the suspect component by reusing `analyse_repo_local`
> (in-process, already proven in the diagnose workflow). Only then does the
> resolver work (§7D) make sense: no query improvement retrieves absent rows.
>
> **Success criterion for the whole route:** a diagnosis run in which at least one
> citation is tier=static from a body that RETRIEVAL put in scope — i.e. the code
> channel demonstrably contributed — with the runtime/DB channel un-regressed.

## The problem, in flow terms (why this route exists)

The diagnose workflow seeds iteration 1's code scope from
`lookup_code_symbols(query = input_data.symptom)`. Three compounding faults:

1. **Timing** — the lookup runs BEFORE `load_runtime`, at the point of maximum
   ignorance; the only text it has is the 8-word product-vocabulary symptom.
2. **Idle resolver** — on iterations 2+, the verdict's free-text `next_scope`
   entries ("LLM call site that produces section content…") pass through the
   engine's `nextScope` verbatim, match nothing in the call graph, and
   `assemble_bundle` warns could-not-read-body and skips them. The one NL→symbol
   resolver we have (the lookup) is never consulted after iteration 1 — exactly
   when the loop starts producing good, evidence-derived descriptions of the code
   it wants.
3. **Missing corpus** — discovered 2026-07-02: the symbols the loop needs are not
   in the index at all (0 rows for page/section/result_spec patterns).

Fault 3 gates faults 1–2. Hence the phase order below.

## §7A — Corpus diagnosis: which mechanism emptied the index — [ ]

Run and paste back (columns verified against code_symbols_actions.go's own
INSERT/SELECT statements):

```sql
-- A1. Directory census: which parts of the tree made it in?
SELECT split_part(path,'/',1) || '/' || coalesce(nullif(split_part(path,'/',2),''),'-') AS dir,
       count(*) AS symbols, count(DISTINCT path) AS files
FROM code_symbols
WHERE repo = 'gqls/agentchassis'
GROUP BY 1 ORDER BY symbols DESC;

-- A2. Is platform/orchestration present at all (actions vs the rest)?
SELECT count(*) FILTER (WHERE path ILIKE 'platform/orchestration/%')             AS orchestration_all,
       count(*) FILTER (WHERE path ILIKE 'platform/orchestration/actions/%')     AS orchestration_actions
FROM code_symbols WHERE repo = 'gqls/agentchassis';

-- A3. Commit hygiene: one commit or a mix (prune interplay)?
SELECT commit_sha, count(*) FROM code_symbols
WHERE repo = 'gqls/agentchassis' GROUP BY 1 ORDER BY 2 DESC;

-- A4. Files indexed vs files on disk (ground truth below)
SELECT count(DISTINCT path) AS files_indexed FROM code_symbols WHERE repo = 'gqls/agentchassis';
```

Repo ground truth (run in $CHASSIS):

```bash
cd "$CHASSIS"
git log -1 --format='%h %ad %s' 4c2c172
find . -name '*.go' ! -name '*_test.go' -not -path './vendor/*' -not -path './.git/*' | wc -l
ls platform/orchestration/actions/ | head -20
ls -la platform/orchestration/result_spec.go
```

Reading the answers (do not pre-commit before seeing them):
- Whole directories absent (e.g. `platform/orchestration/-` missing from A1
  entirely) ⇒ adapter-side excludes or a subset fetch.
- Scattered / alphabetical cut within directories ⇒ transport truncation or an
  ignored partial-walk error (analyse.go:103 returns a PARTIAL Output alongside
  the error; a caller that ignores err keeps the truncation).
- A3 shows multiple commit_shas ⇒ a later partial run + the prune
  (`DELETE … WHERE commit_sha IS DISTINCT FROM $2`) interacted.
- files_indexed ≈ on-disk count but symbols still missing ⇒ different story
  (per-file parse errors — check `SELECT path FROM code_symbols WHERE repo='gqls/agentchassis' AND path ILIKE '%result_spec%'` again per-file) — unlikely given A2, but check before concluding.

- [ ] A1–A4 run and pasted back; mechanism named in NOTES.

## §7B — Corpus fix: swap code-indexer's analysis step to analyse_repo_local — [ ]

Direction (pending §7A confirmation): replace the `request_repo_analysis` (adapter,
Kafka round-trip) step in the code-indexer workflow with `analyse_repo_local`
(in-process, fresh tarball, the SAME analyse.go, no transport) — the component we
already built and proved in the diagnose workflow at this same commit. This
removes the suspect path under every candidate mechanism instead of patching one.

Steps:
1. Dump the REAL workflow first (never write the migration from memory):
   ```sql
   SELECT default_config FROM agent_definitions WHERE type = 'code-indexer';
   ```
   Paste it back; the migration is written against it.
2. Compatibility notes for the migration (from the code as deployed):
   - `analyse_repo_local` output_field should remain the field
     `index_code_symbols` reads (`analysis_field` default
     `repo_analysis.output`); analyse_repo_local returns the decoded Output map
     under `.output` plus `commit_sha`/`owner`/`repo` alongside — same contract
     the diagnose workflow consumes.
   - `index_code_symbols` composes the repo label via `resolveCodeRepoLabel`
     (owner+repo from repo_analysis) — already aligned since the composing-lookup
     image; no config needed.
   - `pin_to_index_commit` must be FALSE/absent for the indexer (it defines the
     index commit; pinning is for readers).
   - snapshot_agent('code-indexer', …) before the UPDATE (standing rule).
3. Migration `NNN_swap_indexer_to_local_analysis.sql` — WRITTEN NEXT TURN against
   the dumped config, with a REVERT block.
4. If §7A instead shows the adapter path is fine and the cause is elsewhere
   (e.g. prune interplay only), fix THAT instead — this section then becomes
   optional hardening.

- [ ] real code-indexer default_config dumped and pasted
- [ ] migration written + applied (snapshot first)
- [ ] REVERT block present

## §7C — Reindex + verify coverage — [ ]

Re-trigger the code-indexer (the §6D kcat trigger block in the old RUNBOOK still
applies — NEW correlation_id; read outcome by correlation_id only). Then verify:

```sql
-- must now return rows (the 2026-07-02 zero-rows query)
SELECT path, symbol, kind, left(coalesce(doc,''),70) AS doc
FROM code_symbols
WHERE repo = 'gqls/agentchassis'
  AND (path ILIKE '%page%' OR path ILIKE '%section%' OR path ILIKE '%result_spec%'
       OR symbol ILIKE '%page%' OR symbol ILIKE '%section%' OR symbol ILIKE '%resolveResultSpec%')
ORDER BY path, symbol LIMIT 40;

-- count materially above 436; single, recent commit_sha
SELECT count(*), count(DISTINCT path), min(commit_sha), max(commit_sha)
FROM code_symbols WHERE repo = 'gqls/agentchassis';
```

- [ ] resolveResultSpec / SavePageSectionsAction / PlanSectionsAction present
- [ ] total count consistent with the on-disk .go file count from §7A
- [ ] one commit_sha, current

## §7D — Evidence-fed resolver in diagnose_route — [ ]

The self-contained lookup change (design agreed 2026-07-02, NOTES v3/v4):

- In `diagnose_route_action.go`, BEFORE calling `diagnose.Advance`: for each
  verdict `next_scope` entry that is NOT an exact `path:Symbol` present in the
  analysis Output (same identity `assemble_bundle` slices by), run the existing
  vector lookup with THAT ENTRY as the query (`top_k` ~3, config
  `resolver_top_k`), and substitute the resolved `path:Symbol`s; exact entries
  pass through untouched; unresolvable entries remain as labels (benign — today's
  behaviour). REUSES `resolveCodeRepoLabel`, `createRAGEmbeddingClient`,
  `vectorSearchCodeSymbols` — no new retrieval machinery.
- Resolve-before-Advance so the call-graph expansion and the narrowing guard
  operate on real symbols; the ENGINE stays pure (resolution lives in the action,
  which has the DB).
- FLAGGED deliberate change: this mutates `v.NextScope` before Advance, so the
  trail records the RESOLVED scope (more auditable, but a change to note).
- Tests: exact-entry passthrough; fuzzy-entry resolution (fake DB rows);
  unresolvable stays label; empty next_scope unchanged.

- [ ] code written (gofmt-clean; sandbox cannot build chassis deps)
- [ ] `go test ./...` on the touched packages (user)
- [ ] deployed (note image tag)

## §7E — Observe: does the code channel now contribute — [ ]

Re-run the §6F envelope (NEW correlation_id; ensure site_id non-empty). Read the
full trail by correlation_id and the lookup log lines. Success criteria:

- [ ] lookup log shows repo="gqls/agentchassis" and hits from the build/coordinator
      domain (not the flat 0.55 generic band)
- [ ] at least one iteration's scope contains a retrieval-resolved symbol whose
      BODY appears in the bundle (assemble does not warn could-not-read-body for it)
- [ ] at least one citation tier=static from such a body
- [ ] runtime/DB channel un-regressed (data_requests still run; outcome grounded)

## §7F — DEFERRED: seed-query improvement

Reorder `lookup_symbols` after `load_runtime` and build the seed query from the
symptom + salient error-log lines (action names in errors near-lexically match
symbol names). Workflow-order migration + Go query construction. Only after §7E
is observed — the resolver may make this unnecessary.

## Parking lot (not this route)

- Trigger intermittently sends site_id="" — fix so it is always present.
- Verdict quality: (a) stale revised-hypothesis text carried verbatim into a
  CONFIRMED conclusion that its own citation contradicts; (b) data_requests
  emitted in a terminal verdict are never run — either strip them at emit or note
  them as "unexecuted follow-ups" in the conclusion.
- Non-manual triggering (was gated on §6G; now gated on the above hygiene).
