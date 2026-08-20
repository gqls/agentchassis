# 334 — `complete_work_item`'s Optional `commit_sha` draws a resolver conflict on every dispatch-loop completion, and it blocks RFC_029's flip to refusal

**Filed:** 2026-08-20, by the staged_component_build lane, from the RFC_029 §10.13 step-4
done-condition read (the candidate-set query it exists to run). **090 diagnosis dispatched**
the same morning — intake correlation `23296951-c032-4842-9e90-aae9b2430870`; find the run by
payload (`spec.dispatch_correlation_id`), not by that printed id. Cross-cutting per the
2026-07-31 owner ruling (it spans the git-adapter reply change, `complete_work_item`'s
registered input spec, and RFC_029's resolver), so the root cause below is **[FILED, loop
verdict pending]** — treat it as the filing session's first-hand read, not settled.

**Severity:** Low today / structural blocker. No wrong value is known to be persisted — the
winner appears to be the right one (see §3) — but the class writes a conflict row on **every**
build-dispatch-loop completion, and RFC_029 §10.13 step 5 (conflicts → refusal) cannot ship
over a population that fires on every build. Under refusal, `complete_work_item` would resolve
NO `commit_sha` and work items would silently lose the sha their `result` now carries.

## 1. Symptom, measured (all 2026-08-20 morning)

```sql
SELECT agent_type, context->>'field', context->'candidate_paths', context->>'winner_path', count(*)
FROM agent_error_log
WHERE error_code='RESOLVER_CONFLICTING_CANDIDATES' AND context->>'field'='commit_sha'
GROUP BY 1,2,3,4 ORDER BY 5 DESC;
```

- `agent_type = build-dispatch-loop`, field `commit_sha`, ~179 rows total; the class **begins
  2026-08-19 20:40:07Z** — before the 22:26Z v1.0.1317 fleet roll. `[UNVERIFIED]` the exact
  git-adapter tag live at 20:40Z; the current adapter pods are from the 22:26Z roll. The start
  coincides with git_commit replies first carrying `commit_sha` (DGH-013's adapter half).
- Candidate paths are the loop's accumulated copies — `handler_result.…`, `handler_result_N.…`,
  `process_item_iter_N_call_handler.…`, all ending `deploy_result(.response).data.commit_sha`,
  with the 7.7% deeper `…deploy_result.response.deploy_result.response…` variant present.
- Winner is always the **unsuffixed** `handler_result.…` path.
- The values land: `SELECT count(*) FILTER (WHERE result ? 'commit_sha'), count(*) FROM
  site_work_items WHERE status IN ('complete','verified') AND updated_at >= '2026-08-19
  20:40:00Z'` → **250 | 323**.

## 2. Mechanism (read at code; the 090 run re-derives independently)

1. `CompleteWorkItemInputSpec` (`platform/orchestration/actions/load_work_item_actions.go:52–62`)
   declares `commit_sha` **Optional** — and its `Deprecated` map shows an explicit
   `commit_sha_field` mapping once existed and was retired in favour of generic resolution.
   An Optional field with no explicit mapping is recovered by the whole-tree search
   (`findFieldRecursive` / `collectFieldCandidates`, `unified_extractor.go`).
2. `build-dispatch-loop` calls `complete_work_item` once per loop iteration; the loop keeps
   every prior iteration's handler result in `collected_data` (unsuffixed `handler_result` =
   the current iteration; `handler_result_N` and `process_item_iter_N_call_handler` = the
   accumulated per-iteration copies).
3. Until 2026-08-19 ~20:40Z no `commit_sha` key existed anywhere in a bdl tree, so the search
   found at most one candidate and the class was invisible. The moment git_commit replies
   carried it (DGH-013), every accumulated copy became a candidate, each with a **different**
   sha → `reflect.DeepEqual` disagrees → a Phase-1 conflict row per completion.
4. **This is NOT the RFC_038 stamp reader.** `resolveDeployEvidence`
   (`deploy_evidence.go`) deliberately does not use the shared search — it has its own
   refuse-on-conflict collector, and its file header explains why. The colliding requester is
   the work-item bookkeeping: `complete_work_item` writes `resultData["commit_sha"] = sha`
   (`load_work_item_actions.go:937–939`). Three mechanisms share the reply key; only this one
   guesses.

## 3. Why the winner is (probably) right, and why that is not good enough

Equal depth, all sibling-recursion rank → the tie-break falls to the collector's sorted-key
DFS, and `handler_result` sorts before `handler_result_0` and `process_item_iter_*`. The
unsuffixed key is overwritten each iteration, so at the moment iteration N's
`complete_work_item` runs it holds iteration N's own result — the correct sha for that item.
`[INFERRED from the sort + bugs_closed/306's declared-rank mechanism; NOT yet proven against a
specific item — the check is to join one item's `result->>'commit_sha'` to the git_commit
result inside its own handler saga in `orchestration_states.collected_data`.]`
Even if right every time, "right by accident of sort order" is bugs_closed/306's exact shape,
and the row-per-completion population blocks §10.13 step 5 for ever if left standing.

## 4. Fix candidates, ordered by what closes the door

1. **Explicit mapping in bdl's `complete_work_item` step config** (config-only, live
   immediately): `"commit_sha?": "handler_result.response.deploy_result.response.data.commit_sha"`.
   Strategy-0 resolution + the LIVE step-1 prune (v1.0.1310) means the whole-tree search never
   runs for a Strategy-0-resolved field → rows stop. The `?` keeps it optional for the 7.7%
   deeper shape (those items would lose the sha unless a second alias row handles them — count
   that population before choosing). Check first what bdl's step config actually looks like
   (`jsonb` of the live definition, not the seed) and whether the deprecated `commit_sha_field`
   still parses.
2. **Un-declare it**: drop `commit_sha` from `CompleteWorkItemInputSpec.Optional` and have the
   dispatch loop pass it explicitly where wanted. Cleanest door-close (the spec is why the
   search runs at all), but a Go change (build+roll) and needs the consumer census: who reads
   `site_work_items.result->>'commit_sha'`? If nothing does (DGH-013's register entry says
   nothing consumes `content_hash` yet; the result sha may be equally unconsumed), the right
   fix may be deletion, not plumbing.
3. **Safe-by-inspection note only** (no code): document that the winner is pinned by the
   declared tie-break to the current iteration and accept the rows until step 5's design
   handles per-loop accumulation generally. Weakest — it normalises a row-per-completion
   instrument population, which is exactly what §10.13 step 4 exists to drain.

Do NOT fix this by adding `commit_sha` to `renderContextStepContractRenames` — that map is the
render-context step contract, size-pinned by test, and this collision is same-type
(string vs string), not the two-types-one-key shape. The producers here are the loop's own
bookkeeping copies, not two namespaces.

## 5. Relations

- **Blocks:** RFC_029 §10.13 step 5 (staged_component_build lane) — its gate is "window reads
  zero, or every surviving pair carries a written safe-by-inspection note".
- **Trigger:** DGH-013 (git-adapter reply carries `commit_sha`/`files_sha256`; register entry
  records the caller traps), `bugs_open/315` / RFC_038 lane (owner of the stamp mechanism —
  their reader is NOT implicated; notified so they know a third consumer of their key exists).
- **Mechanism:** RFC_029 §9/§10 (Phase-1 resolve-and-warn), CTS-060, `bugs_closed/306`
  (the tie-break that makes today's winner deterministic).
- **Sibling residual, same gate:** `tool-generator` / `reason` + `related_pages` conflict rows
  (since 2026-08-16, 23+19 rows) — generic field names matching spec-array keys; untraced;
  whoever picks this file up should trace those to their requesting step the same way.

## 6. How to verify a fix

The RUNBOOK candidate-set query (staged_component_build, "step 4's done-condition") reads zero
rows for `build-dispatch-loop`/`commit_sha` against live demand (completions in the window:
`site_work_items` rows newly `complete`), AND completed items that deployed still carry
`result.commit_sha` (the fix must not silently drop the value — 250/323 is the baseline).
