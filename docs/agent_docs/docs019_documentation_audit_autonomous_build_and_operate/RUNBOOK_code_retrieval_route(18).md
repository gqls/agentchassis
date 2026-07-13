# RUNBOOK — §7 code-retrieval route (make the code channel earn its keep)

Continues from `RUNBOOK.md` (§6 diagnosis loop — complete). Detailed history in
`NOTES_running_synthesis_v4.md` (and v3, archived). Standing rules: user runs
all SQL/kubectl/builds; read outcomes by correlation_id only; snapshot_agent
before agent_definitions UPDATEs; schema before SQL; a 0-rows result is not
decisive until the query itself is checked.

> ## ▶ CURRENT POSITION — 2026-07-03 (read this first — supersedes all earlier position blocks)
>
> **Where we are:** §7A–§7D complete. §7E attempt 1 (run `17933a83`) executed
> the FULL loop but surfaced two independent defects — **both now fixed**; the
> fixed image is **deploying**. §7E attempt 2 is the next action.
>
> **State of each piece:**
> - **Corpus (§7A–§7C + §7C.1):** current and clean — 3,723 symbols / 449 files
>   at single commit `e3176f8`; docs/ archives excluded at source
>   (exclude_patterns, default ["docs/"]) and pruned; `resolveResultSpec` and
>   the plan/save action symbols confirmed present.
> - **Resolver (§7D):** live — fuzzy next_scope entries resolve to real
>   path:Symbol handles (resolver_top_k 2; min_similarity 0.55 pending §7E
>   calibration; trigram fallback; fail-open).
> - **Delivery fix (applied 2026-07-03, no rebuild needed):** `result_from`
>   was DEAD CONFIG — CompleteWorkflowAction never reads it, so the fallback
>   shipped the full 1.27MB collected_data past the kafka cap. Child complete →
>   `output_fields ["diagnosis"]`, parent → `output_fields
>   ["diagnose-agent_result"]` (snapshots 34f4afc8 / e8e96d24; verify green).
> - **Guard fix (in the build deploying now — note the tag when it lands):**
>   the narrowing guard measures the MODEL-NAMED scope (namedScope); call-graph
>   expansion runs only after the guard passes and is capped (default 18,
>   named entries always kept; <0 unlimited); Advance threads PrevScopeSize as
>   the named size. Four regression tests in diagnose/loop_scopeguard_test.go.
> - **Evidence already banked (17933a83's salvaged trail):** seed relevance MET
>   — twelve build-domain seed symbols, first time ever; correct abstention;
>   three well-formed site-scoped data_requests (never ran — the guard, now
>   fixed); no static-tier citation yet; resolver substitution unexercised
>   (all six next_scope entries arrived exact).
>
> **§7E ATTEMPT 2 TRIGGERED — 2026-07-03:** correlation
> `73ed55c6-6978-4951-b40d-0cec771b5208`, orch `0062d43e-aa70-44e3-932f-8b62658d9c0c`.
> **SCORED 2026-07-03 from compounded_logs.txt (5 of 6 ✔; formal close pending
> the sql_result.txt re-upload — it arrived 0 bytes):**
> ✔ (1) delivery: full completion chain, response produced + parent notified,
>       ZERO Message Size errors — the output_fields migration proven live.
> ✔ (2) guard: ZERO scope-not-narrowing; resolver ran on two route passes
>       (16:23:36 in 3→out 6; 16:30:26 in 2→out 4) with completion at 16:38 —
>       a ~three-iteration run whose final verdict skipped the resolver (the
>       cited-Confirmed skip), itself hinting the terminal outcome.
> ✔ (3) resolver calibration: TEN resolutions, similarities 0.66–0.87
>       (identity-ish 0.74–0.87; the descriptive site_specs/cta entry →
>       applyFieldUpdate/loadCurrentSpecData at 0.66–0.67); ZERO floor filters
>       — 0.55 safe, headroom to ~0.60 later. BONUS behaviour worth keeping:
>       the resolver CANONICALISES basename entries (plan_sections_action.go:
>       planSection) to full paths at 0.81–0.83 — six entries that would
>       otherwise have been dead labels.
> ✔ (4) bundle bodies: proven by (5).
> ✔ (5) STATIC-TIER CITATION — THE ROUTE'S SUCCESS CRITERION:
>       Tier 0, Where plan_sections_action.go:planSection, quoting the body:
>       `case "skip_field": // Just omit it` — the code branch that silently
>       omits a section field whose spec is missing (the cta_url mechanism) —
>       followed by a NEW data_request measuring length(rendered_html) per
>       slot. Code read → mechanism found → targeted follow-up read.
> ? (6) seed band: not captured (tail attached after lookup); rides any later run.
> PENDING for formal close: terminal status/conclusion/full trail from the
> re-uploaded sql_result.txt; delivery-verification row (has_child_result).
> NEW PARKING-LOT ITEM: result_spec.go:174 WARN at completion — "deprecated
> complete-step key in use; migrate to preferred name" (resolved + delivered
> anyway); paste sed -n '160,190p' platform/orchestration/result_spec.go to
> name the preferred key, then a one-line config migration.
> Original green-list kept for reference:
> (1) the CHILD row COMPLETED — not FAILED-at-complete (delivery fix);
> (2) >1 iteration in the trail and the three data_requests RAN at the next
>     gather (guard fix);
> (3) resolver similarities in the pod logs (floor calibration);
> (4) bundle bodies present for resolved entries;
> (5) any static-tier citation — the route's success criterion;
> (6) the seed similarity band (decides whether §7F stays retired).
>
> ```sql
> -- state (expect parent COMPLETED, child COMPLETED)
> SELECT status, current_step, EXTRACT(EPOCH FROM (NOW()-last_activity))::int AS since_s,
>        substring(COALESCE(error,''),1,400) AS err
> FROM orchestration_states
> WHERE correlation_id = '73ed55c6-6978-4951-b40d-0cec771b5208'::uuid ORDER BY created_at;
>
> -- delivery verification: parent must now HOLD the child result, small
> SELECT status, pg_column_size(collected_data) AS cd_bytes,
>        (collected_data ? 'diagnose-agent_result') AS has_child_result
> FROM orchestration_states
> WHERE correlation_id = '73ed55c6-6978-4951-b40d-0cec771b5208'::uuid ORDER BY created_at;
>
> -- the diagnosis + full trail (the child row, selected by key presence)
> SELECT jsonb_pretty(collected_data->'diagnosis')
> FROM orchestration_states
> WHERE correlation_id = '73ed55c6-6978-4951-b40d-0cec771b5208'::uuid
>   AND collected_data ? 'diagnosis';
> ```
>
> ```bash
> # resolver + guard signals from the spawned pod (before agent-job-cleanup reaps it)
> P=$(kubectl -n ai-persona-system get pods -o name | grep agent-diagnose-agent | head -1)
> [ -n "$P" ] && kubectl -n ai-persona-system logs "$P" --tail=8000 | \
>   grep -E 'scope resolver applied|resolved fuzzy scope entry|below similarity floor|generated query embedding|scope-not-narrowing|data_request'
> ```
>
> **Then:** score §7E; if green the route closes, and the parking lot queues
> behind it (site_id intermittency; verdict-quality wrinkles; a size guard on
> the complete-fallback; CI-triggered indexing; RuntimeSite hygiene).

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
- [x] reindex f284b749 VERIFIED (2026-07-03): commit **e3176f8**, 3,723
      symbols, docs_rows=0 → the full build→push→deploy→retrigger sequence ran.
      §7C.1 verified end-to-end (excludes applied on the new image; prune
      cleared everything at 36710be, pollution included) and the §7D resolver
      image is LIVE. Symbol delta 3,725→3,723 across the commit = ordinary
      branch churn; the gates are single-commit + zero docs rows, both green.
- [!] §7E ATTEMPT 1 (correlation 17933a83, orch fc60af15) FAILED — PRODUCER
      NAMED (orchestration_states): the CHILD's complete_workflow response to
      the parent's responses topic — Kafka broker error [10] Message Size Too
      Large. THE LOOP ITSELF RAN TO COMPLETION (child FAILED only at
      complete); the parent recorded CHILD_ORCHESTRATION_FAILED via the small
      failure notification. agent_error_log 0-rows = wrong key (parent orch
      id; site-build table) — not chased, the state row sufficed.
      ARITHMETIC ANOMALY: response = result_from "diagnosis" (status/
      conclusion/stopped_by/trail); trail Steps carry Iteration/Hypothesis/
      Scope/BundlePath/Verdict/GuardStop, verdicts ≤ ~32KB (max_tokens 8000),
      ≤5 iterations ⇒ ~200KB expected ≪ 1MB default cap. So EITHER emit's Go
      defaults add a heavy field beyond the four the migration names (emit
      code not held locally — request upload), OR the cluster max.message.bytes
      is set BELOW default. DECISIVE PROBE (one number):
      pg_column_size(collected_data->'diagnosis') on the child's FAILED
      orchestration_states row (\d orchestration_states first; adjust column).
      MB-scale → heavy field in emit → slim it. ~100–300KB → low cluster cap →
      kubectl -n kafka get kafka personae-kafka-cluster -o yaml | grep -i message.max
      FIX (both branches, structural): emit gains a RESPONSE BUDGET — the
      Kafka-borne diagnosis = status/conclusion/stopped_by + trail truncated to
      a byte budget (per-quote caps); the FULL trail stays persisted in
      collected_data, retrievable by correlation_id. Responses are summaries;
      heavy artifacts live in the DB. Broker-cap raise = inversion, last resort.
      MEASURED (2026-07-03): diagnosis_bytes=6,198 / trail_bytes=4,786 — emit
      EXONERATED (its upload no longer needed); cd_bytes=1,270,781 — the child
      collected_data is 1.27MB, just over the ~1.05MB Kafka default ⇒ THE
      COMPLETION RESPONSE CARRIED THE FULL collected_data alongside the 6KB
      result (repo_analysis dominates). Open fork vs f284b749 (indexer, same
      515-file analysis, delivered fine): (a) output_fields slims but
      result_from ships context — fix = one-line migration to output_fields
      ["diagnosis"] (§7B precedent); or (c) both ship full cd and the indexer
      squeaked under (~0.9–1.0MB cd, no loop artifacts) — fix = Go, response
      builder sends the selected result only. DISCRIMINATE: pg_column_size on
      f284b749's child row; CONFIRM MECHANISM: upload the complete_workflow
      action + the "failed to send response" builder
      (grep -rn "failed to send response" platform/ internal/).
      FOLLOW-UP HYGIENE (non-blocking either way): 1.27MB cd is also written to
      orchestration_states per transition — Output-to-pod-file slimming stays
      listed; touches route/resolver/assemble, so not the first move.
      NOTE trail_bytes 4.8KB ⇒ likely 1–2 iterations only — stopped_by in the
      diagnosis will say why.
      §7E SCORED on 17933a83's salvaged trail (2026-07-03):
      ✔ SEED RELEVANCE MET — all twelve seed symbols build-domain
        (resolveSectionIndexForType, isSectionIndexURL, findPagesWithNoContent,
        crawlPageContent, sourceResolver, plan/save territory) — first time
        ever; §7F (seed reorder) substantially retired.
      ✘ static-tier citation — not met (two tier-2 runtime citations; loop cut
        at iteration 1).
      ? resolver substitution — unknown: all six next_scope entries arrived
        EXACT (model may have named them after reading relevant seed bodies —
        itself a seed win); resolver pod logs settle it if not reaped.
      ⊘ runtime channel — blocked, not regressed: three well-formed
        site-scoped data_requests were never run.
      BLOCKER 2 FOUND (engine, mechanism CONFIRMED in code): guardAfter
      measures the POST-EXPANSION scope; nextScope expands every named symbol
      by its full Neighbourhood, unbounded. The stale corpus masked this for
      all of §6 (69-file graph ⇒ Neighbourhood empty ⇒ expansion no-op). With
      the real 515-file graph, six named action symbols expanded past the
      prev+2 allowance (15) ⇒ scope-not-narrowing at iteration 1 — BEFORE the
      data_requests ran; the scope guard sits before the new-request escape so
      the #1 rule cannot rescue it.
      FIX (specified, pending go-ahead; ships with the delivery fix in one
      build): guard on the MODEL-NAMED scope — DecideStep builds `named`
      (carried fields + deduped v.NextScope, no expansion), guardAfter
      compares named.size() vs prevNamedSize+2, Advance threads PrevScopeSize
      as the NAMED size; the expanded scope is used only for the gather.
      Companion: Config.MaxExpandedScope (Go default ~18, named-first) so
      expansion cannot drown bundle signal. REJECTED: a data_request escape on
      the scope guard (would render it near-inert). FLAGGED: StepDecision/
      threading gains the named size; Config gains one field; step.go needs
      the import rewrite on the chassis copy.
      PARKED: verdict wrote RuntimeSite="page-build-handler" (an agent, not a
      domain) — nextScope carries it and would poison the next runtime gather;
      verdict-prompt/wire hygiene item.
      BOTH FIXES DELIVERED 2026-07-03:
      • Blocker 1 (delivery) — mechanism READ FROM CODE (workflow_actions.go
        extractFinalResult): `result_from` is a key CompleteWorkflowAction
        NEVER reads; the fallback ships the ENTIRE non-system collected_data.
        Empirically sealed: indexer (output_fields) delivered at cd 1,213,045;
        diagnose (result_from) failed at 1,270,781. FIX =
        NNN_fix_diagnose_complete_output_fields.sql — child complete →
        output_fields ["diagnosis"], parent → output_fields
        ["diagnose-agent_result"]; snapshots; REVERT included. APPLIED 2026-07-03
        (snapshots 34f4afc8 / e8e96d24; verify shows output_fields on both). PARKED (Go hardening, future build): a
        size guard + warn on the ship-everything fallback.
      • Blocker 2 (guard-vs-expansion) — engine fix WRITTEN:
        diagnose/{loop,step,advance}.go + route action + 4 regression tests
        (loop_scopeguard_test.go). Guard now measures the MODEL-NAMED scope
        (namedScope: deduped v.NextScope, no expansion); expansion runs only
        after the guard passes and is CAPPED (Config.MaxExpandedScope, engine
        default 18, named entries always kept, <0 = unlimited); Advance
        threads PrevScopeSize = NamedScopeSize. FLAGGED changes: StepInput +
        StepDecision + LoopState + Config each gain ONE field; guard-stop
        decisions now record no expanded scope (expansion skipped on stop);
        route step gains optional max_expanded_scope (Go default 18). All
        engine files here are the CHASSIS variants (step.go retains the
        agentchassis import).
      DEPLOY: apply the migration (immediate); go test ./diagnose/...
      ./platform/orchestration/actions/... ; go build ./... ; push; deploy
      (note tag). NO reindex needed. Then §7E ATTEMPT 2 (same trigger script,
      fresh ids).
      §7E EVIDENCE SALVAGE (this run counts): the trail in collected_data holds
      every iteration's Scope + Verdict (resolver substitutions + citation
      tiers) — jsonb_pretty(collected_data->'diagnosis'->'evidence_trail') by
      correlation; resolver pod-log grep if agent-job-cleanup has not reaped.

## §7E — Observe: does the code channel now contribute — [ ]

TRIGGER (user script, reviewed 2026-07-03 — matches the proven §6F envelope;
REF is pin-overridden to the index commit by analyse_repo_local, which is the
intended behaviour; SITE_ID guard added as insurance against the historical
intermittent-empty pattern):

```bash
set -euo pipefail
TARGET_AGENT_TYPE='diagnose-orchestrator'
OWNER='gqls'; REPO='agentchassis'; REF='083_imagery'
SYMPTOM='index page completed but content is a stub'
RUNTIME_SITE='gamesdesign.co.uk'
SITE_ID='e33263f4-74f8-494f-b191-546845dbbddf'
[ -n "$SITE_ID" ] || { echo "site_id empty — aborting"; exit 1; }
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID='demo_client'
echo "SAVE: CORRELATION_ID=$CORRELATION_ID  ORCHESTRATION_ID=$ORCHESTRATION_ID"
kubectl -n kafka run -i --rm "kcat-diagnose-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=manual-diagnose-$(date +%Y%m%d-%H%M%S)" \
  -H "step_name=start" \
  -H "client_id=$CLIENT_ID" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "responses_topic=system.agent.generic.responses" <<JSON
{"action":"orchestrate","config":{"agent_type":"$TARGET_AGENT_TYPE"},"input_data":{"owner":"$OWNER","repo":"$REPO","ref":"$REF","symptom":"$SYMPTOM","runtime_site":"$RUNTIME_SITE","site_id":"$SITE_ID"}}
JSON
echo "Tail by correlation:"
echo "  kubectl -n ai-persona-system logs -f -l agent_type=diagnose-agent --tail=500 | grep '$CORRELATION_ID'"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis    --tail=500 | grep '$CORRELATION_ID'"
echo "Resolver signals (the §7E point of the run):"
echo "  kubectl -n ai-persona-system logs -l agent_type=diagnose-agent --tail=2000 | grep -E 'scope resolver applied|resolved fuzzy scope entry|below similarity floor|generated query embedding'"
```

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
