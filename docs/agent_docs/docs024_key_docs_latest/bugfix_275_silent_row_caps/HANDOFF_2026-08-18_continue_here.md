# HANDOFF 2026-08-18 — continue here (bugs 257, 275, and the two tickets it spawned)

**Read this first; it is the cold start.** Two bugs were taken from `bugs_open/` and carried to
council-approved and live. Both are essentially done. What remains is **three owed verifications, all
gated on fleet traffic rather than on work**, plus two new tickets nobody has started.

---

## 1. State, at a glance

| thing | state | live? |
|---|---|---|
| `bugs_open/257` — token budget at the provider client | **council APPROVED** (`366efae9`), fixed, verified | **LIVE** since v1.0.1305 |
| `bugs_open/275` — tool-suggester saw 30 of 74 tools | **council APPROVED** (`b684a399`) | migrations **LIVE**; Go detector **LIVE** at v1.0.1309 |
| register **MDL-041** (budget at the client) | registered, index row present | live |
| register **LCO-009** (silent row-cap detector) | registered, index row present | live, **but see §4** |
| `bugs_open/297` — tool-recreation, 10 of up to 107 | **FILED, not started** | n/a |
| `bugs_open/298` — internal-linker, 15 of up to 68 | **FILED, not started** | n/a |

**Nothing is half-committed.** Working tree is clean for every file either lane touched; the clean
`git archive HEAD` tree builds and `platform/aiservice` + `platform/orchestration/actions` are green.

---

## 2. What was actually done

**257** — `ai_service.max_tokens` was read only by `ExecuteAIStepAction`, so any caller building its own
client and calling `GenerateText(ctx, prompt, nil)` silently got a hardcoded 2048. All three provider
clients now resolve the budget from the `ai_service` config they were constructed with. Precedence:
per-call options (wins, unchanged) → construction config (new) → provider fallback (unchanged).
Measured: **zero fleet requests changed** — the canonical path was provably unaffected because
`createAIClient` and the options build read one merged map.

**275** — `load_library_tools` ended `ORDER BY display_name LIMIT 30` against a **74**-tool library.
Fixed by bounding `description` (80% of the payload) instead of coverage: migration **445** removed the
cap, migration **446** marks the truncation (`[…truncated]`). Both **applied and live**. Plus the class
fix: `QueryDatabaseAction` now WARNs when a result reaches its query's own trailing literal `LIMIT`
(register LCO-009), making all such caps visible at the one point they share.

---

## 3. ⚠ THE THREE OWED VERIFICATIONS — this is the actual work left

### 3a. 275's "after" half (the bug's own disconfirming pair)

The **before** half is **proven at the artefact**: the last pre-fix `suggest_tools` prompt, ranked
against the library *as it stood then* (71 masters), contained **29 tools, 0 past rank 30, highest rank
exactly 30**. The model saw the first thirty alphabetically and nothing beyond.

**Owed:** the same query against a POST-fix prompt, asserting a rank > 30 tool IS present.
**Blocked on:** `suggest_tools` has run **0 times** since migration 445 went live (2026-08-17 11:22Z).

```sql
-- has it run yet?
SELECT count(*) FROM llm_call_log WHERE agent_type='tool-suggester' AND step_name='suggest_tools'
  AND created_at > timestamptz '2026-08-17 11:22:23+00';
-- when it has, run this (⚠ rank against the library AS IT STOOD AT THE PROMPT'S TIMESTAMP)
WITH lib AS (SELECT id::text tid, row_number() OVER (ORDER BY display_name) rn
             FROM content_components
             WHERE component_level='tool' AND forked_from IS NULL AND is_active AND html_template != ''
               AND created_at <= (SELECT created_at FROM llm_call_log
                                  WHERE agent_type='tool-suggester' AND step_name='suggest_tools'
                                  ORDER BY created_at DESC LIMIT 1)),
     p AS (SELECT prompt_rendered pr FROM llm_call_log
           WHERE agent_type='tool-suggester' AND step_name='suggest_tools'
           ORDER BY created_at DESC LIMIT 1)
SELECT count(*) FILTER (WHERE p.pr LIKE '%'||lib.tid||'%') AS in_prompt,
       count(*) AS eligible,
       count(*) FILTER (WHERE p.pr LIKE '%'||lib.tid||'%' AND rn > 30) AS past_rank_30_MUST_BE_NONZERO
FROM lib CROSS JOIN p;
```

⚠ **Ranking today's library against an older prompt gives a FALSE answer** — it reported "1 tool past
rank 30" purely because tools were added after that prompt rendered. Constrain by the timestamp.

### 3b. The live cap census

**Run 2026-08-18, and the answer was "nothing it watches has executed".** 41 pods, 24h: **0 WARNs, 21
`query_database` completions, not one of them a capped step** (`find_dispatchable_site` ×9,
`notify_scheduler` ×6, `load_entry` ×3, `notify_scheduler_idle` ×3).

⚠ **Never report the zero without that attribution** — the log line carries `step_name`, so there is no
excuse for an unattributed zero, and one would be a blind pass.

```bash
kubectl -n ai-persona-system get pods -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.spec.containers[0].image}{"\n"}{end}' \
  | grep 'agent-chassis:v1.0.13' | awk '{print $1}' > /tmp/pods.txt
cat /tmp/pods.txt | xargs -P 12 -I{} sh -c '
  L=$(kubectl -n ai-persona-system logs {} --since=24h 2>/dev/null)
  echo "{} $(printf "%s" "$L" | grep -c "EQUALS the query.s LIMIT") $(printf "%s" "$L" | grep -c "QueryDatabaseAction: Complete")"
' | awk "{w+=\$2; q+=\$3} END {print \"WARNs:\", w, \" completions:\", q}"
```
⚠ **Serial `kubectl logs` over 41 pods TIMES OUT at 2 minutes.** Use `xargs -P`.

**First expected positive:** `content-feed-refresh` (`find_news_sites`, cap 5, population 9), 6-hourly,
last fired 14:32Z. `model-directory-publish` is also 6-hourly but its population is 3 against a cap of
12, so it should stay silent — **that is a useful live negative control.**

### 3c. `bugs_open/298`'s reachability question

`internal-linker` has **zero `llm_call_log` rows in all history**, so whether its `LIMIT 15` has ever
shaped a link decision is **unmeasured** and 298 says so rather than guessing. **The LCO-009 detector
will answer this by itself** once `load_candidate_pages` runs — no separate work needed.

---

## 4. ⚠ THE LIMITATION THAT MAY DEFEAT 3b AND 3c ENTIRELY — read before trusting a quiet census

**The WARN is a log line, so its history dies with the pod.** The observable window is *time since the
last pod restart*.

Measured 2026-08-18: pods restarted **15:45Z**, while `content-feed-refresh` last fired **14:32Z** and
`model-directory-publish` **12:15Z** — **both before the restart, so both are invisible**.

Rolls land roughly **daily** on this tree; the capped steps are on **6-hourly** schedules. **A cap that
fires shortly before a roll is invisible for ever.** The detector is correct and live (unit tests
mutation-proven; binary probe with controls) — but whether it *catches* the caps in practice is a race
nobody has characterised.

**This is the strongest argument for the recorded follow-up**: a `LIMIT n+1` probe writing a
**durable** result (a `doc_notes` row) rather than a log line. Deliberately NOT done — it changes what
a shared action RETURNS, which is a separate change with its own review, and the council was explicit
that observational scope was right for this one.

---

## 5. Traps that will mislead you (all measured, all in LANDMINES)

1. **A "fresh build" can ship NO new code.** A same-tag rebuild serves the node's cached image: new
   pods, new start times, real deploy event, zero new code. Happened on 2026-08-17 14:42Z — 238
   commits (26 touching Go) were unshipped while everything said "deployed". **Check the image
   digest, or probe the binary with a POSITIVE and a PLAUSIBLE-FAKE control in the same breath.**
2. **`orchestration_states.owner_agent_type` is NOT "which agent ran".** It returned 0 for
   `tool-recreation-handler`, which has 290 `llm_call_log` rows. It fails toward *dormant* — the
   answer that makes a bug go away. Use `llm_call_log.agent_type` AND `site_work_items.handler_agent`;
   each has a blind spot, so agreeing zeros are evidence and one zero is not.
3. **Two live agents one word apart.** `internal-link-resolver` (busy, no capped query) vs
   `internal-linker` (the one carrying the defect). **The step list is the identity.**
4. **A census answers the question you ENCODED.** Twice in this work: counting root `ai_service` blocks
   and reporting it as "agents that configure a budget" (13 vs a true **68**); and counting query texts
   containing a multi-row LIMIT and reporting it as "silent caps" (seven vs a true **three** — two were
   subquery limits, two were work queues).
5. **Rank a stored artefact against the state AT ITS TIMESTAMP**, not today's (§3a).
6. **A migration sketch must show the SAFETY-CRITICAL lines** (`snapshot_agent`, pre-state gates), not
   the interesting ones. Two council rounds were spent on objections to sketches whose files were fine.

---

## 6. Where everything lives

- **This lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_275_silent_row_caps/` — PLAN, NOTES
  (the full misstep record), RUNBOOK (every command with its gotcha), README (owner-facing prose),
  SUMMARY, the council submission JSON.
- **257's lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_257_token_budget_at_the_client/`
- **Bugs:** `bugs_open/257`, `bugs_open/275` (both carry full §LIVE / §APPROVED sections),
  `bugs_open/297`, `bugs_open/298`
- **Register:** MDL-041 (`model-infrastructure.md`), LCO-009 (`llm-call-observability.md`), both with
  index rows in `000_concept_index.md`
- **Migrations:** `sql_for_agents/445` + `446` and their `_ROLLBACK` sidecars, both recorded in
  `schema_migrations`
- **Code:** `platform/aiservice/max_tokens.go` (+ test), `platform/orchestration/actions/query_row_cap.go`
  (+ test), `database_actions.go`

## 7. If you want new work instead

`bugs_open/297` is the strongest available ticket: **live** (290 LLM calls), bites at **19 of 24
sites**, and at the median site the model sees 10 of 26 — **worse than the bug this lane fixed**. Its
fix shape is already worked out in the file: measure which column dominates the payload, bound that,
drop the row cap, **and mark the truncation** (migration 446 is the worked remedy).
