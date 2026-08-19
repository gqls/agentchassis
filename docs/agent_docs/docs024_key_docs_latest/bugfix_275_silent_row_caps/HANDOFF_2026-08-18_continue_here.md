# HANDOFF 2026-08-18 — continue here (bugs 257, 275, and the two tickets it spawned)

**Read this first; it is the cold start.** Two bugs were taken from `bugs_open/` and carried to
council-approved and live. Both are essentially done. What remains is **three owed verifications, all
gated on fleet traffic rather than on work**, plus two new tickets nobody has started.

> ## ⚠ STATE AS AT 2026-08-19 09:00Z — READ THIS FIRST, THEN THE 08-18 BLOCK BELOW
>
> **Build:** `v1.0.1314`, rolled 07:50–07:52Z, verified at the binary (stamp `d3590ca46`; three
> controls absent, **including the previous build's stamp**, so this shipped new code rather than a
> cached image). Detector still an ancestor. **Nothing in this lane is fixed by that roll** — 313 and
> 316 are config defects, and 313's live config still reads `array` / `candidate_pages.count > 0`.
>
> **Census (durable, `collected_data`, ~2-day retention):** `content-feed-trigger` **5 of 5 runs hit
> its cap**; `model-directory-trigger` **0 of 4** (the control). ⚠ Yesterday's "3 of 4" became "5 of 5"
> because the one under-cap run **aged out of the window** — the fleet did not change.
>
> **THREE DECISIONS ARE OWED BY THE OWNER, none urgent, all cheap** — see `README_where_we_are.md`,
> entry of 2026-08-19:
> 1. **Fix `bugs_open/313`?** The internal linker has produced no link in four months. Config-only,
>    live on apply, but BOTH halves (`output_format: object` **and** the `plan_links` template) or it
>    trades a dead branch for a broken prompt. Fixing it makes `bugs_open/298`'s cap live on 8 sites.
> 2. **Fix `bugs_open/316`?** Ordering (fairness, one config change) and capacity (spend-vs-freshness,
>    genuinely the owner's call) are separable — do not conflate them.
> 3. **Keep waiting on §3a?** `suggest_tools` has not run since 2026-08-15. The alternative is to
>    dispatch one deliberately, which creates real work at a site.
>
> **Witnessing the WARN: tried, failed, unattributed** (08-18 20:32Z — see the 08-18 block and NOTES).
> `collected_data` answers the same question retroactively; **do not spend more time on the log line.**

> ## ⚠ CORRECTED 2026-08-18 (evening) — READ THIS BEFORE §3b, §3c AND §4
>
> **The census instrument specified below cannot work, and its zero is not a negative result.**
> The chassis log retains **15–90 seconds** (measured, pods up 27 min with 0 restarts), not "time
> since the last pod restart". **Census `orchestration_states.collected_data` instead** — it holds
> every fact the WARN reports, survives rolls, and reaches back ~2 days. Doing so answers §3b
> (**4 cap hits**, with two working negative controls) and §3c (**answered, and it changes
> `bugs_open/298`**). §3a is unchanged and still owed. Full detail: NOTES, entry of 2026-08-18
> (evening); commands: RUNBOOK, "Census the caps from `collected_data`".
>
> **§3c produced a new bug: `bugs_open/313`** — `internal-linker`'s `check_candidates` can never be
> true, so `plan_links` has never run and the agent has never made a link. Diagnosis loop
> **CONFIRMED** (`c4aa3559`), first iteration. It dates to the agent's creation (2026-04-12), and 57
> `needs_internal_links` items read `complete` across it. **Fix 313 BEFORE 298's cap** — the cap fix
> alone changes nothing observable.
>
> **UPDATE 20:32Z — the cap FIRED and the WARN was STILL not witnessed.** `content-feed-trigger` returned
> **5 of cap 5** (durable, in `collected_data`); the streamed capture caught 0 WARN lines *and* missed
> that step's unconditional completion line, so the miss is at the observation layer and is
> **unattributed** — wrong-pod, disconnect, old-binary and general stream-loss were each ruled out (see
> NOTES). **Do not re-run this experiment expecting a different result without first fixing the
> selector** (`-l app=agent-chassis` covers 2 of 62 chassis-image pods; the rest are `app=dynamic-agent`
> ephemerals that are deleted within minutes). **Nothing depends on it:** `collected_data` answers the
> question retroactively, which is now the recommended method everywhere in this lane.
>
> ~~**THE ONE THING LEFT ON THE CLOCK:** the WARN has still never been observed firing.~~ Since the
> detector went live (15:45Z) only one capped step has run, and it was correctly silent (4 rows vs cap
> 12, 18:15Z). The next chance is `content-feed-trigger` at **~20:32Z**; a streaming capture is
> attached until 20:45Z at
> `~/.claude-scratch/…/scratchpad/lco009_capture.txt`. **If you are reading this after 20:45Z, do not
> hunt the logs — read `collected_data`** (RUNBOOK), which answers it whenever you look.

---

## 1. State, at a glance

| thing | state | live? |
|---|---|---|
| `bugs_open/257` — token budget at the provider client | **council APPROVED** (`366efae9`), fixed, verified | **LIVE** since v1.0.1305 |
| `bugs_open/275` — tool-suggester saw 30 of 74 tools | **council APPROVED** (`b684a399`) | migrations **LIVE**; Go detector **LIVE** at v1.0.1309 |
| register **MDL-041** (budget at the client) | registered, index row present | live |
| register **LCO-009** (silent row-cap detector) | registered, index row present | live, **but see §4** |
| `bugs_open/297` — tool-recreation, 10 of up to 107 | **FILED, not started** | n/a |
| `bugs_open/313` — **internal-linker has NEVER made a link** (branch after the capped step can never be true) | **FILED 08-18, diagnosis CONFIRMED (`c4aa3559`). NOT FIXED.** Live config still `array` / `.count > 0` as at 08-19 09:00Z | n/a |
| `bugs_open/316` — **the news-feed cap serves the alphabet**: ranks 1–5 0% late, ranks 6–9 always late; queue 2.1× oversubscribed | **FILED 08-19. NOT FIXED.** Narrows LCO-009's "eventual coverage is not a defect" gloss | n/a |
| `bugs_open/298` — internal-linker, 15 of up to 68 | **reachability ANSWERED 08-18** — the cap is moot; the agent never reaches `plan_links` at all (diagnosis `c4aa3559`) | n/a |

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

> **CORRECTED 2026-08-18 (evening) — ANSWERED, but not by this recipe.** The pod-log census below is
> superseded: see the banner. Two independent reasons its zero meant nothing — the log window is
> ~15–90 s, and `v1.0.1309` (the first release carrying the detector) shows an earliest evidenced
> roll of **15:45:31Z**, so most of that 24 h had no detector in the fleet at all.
>
> **The durable answer, from `collected_data`, over the ~2 days that table retains:**
>
> | agent | step | cap | runs measurable | **hit the cap** |
> |---|---|---|---|---|
> | `content-feed-trigger` | `find_news_sites` | 5 | 4 | **3** |
> | `internal-linker` | `load_candidate_pages` | 15 | 2 | **1** |
> | `model-directory-trigger` | `find_directory_sites` | 12 | 5 | **0** ← negative control, as predicted |
>
> `content-feed-trigger` also came back **under** its cap on one of its four runs, so the method
> demonstrably produces negatives. **Since the detector went live at 15:45Z, exactly one capped step
> has run** — `model-directory-trigger` at 18:15Z, 4 rows against a cap of 12, so no WARN was due.
> **The first genuine positive opportunity is `content-feed-trigger` at ~20:32Z** (eligible
> population 6 against a cap of 5 as at 18:35Z; the predicate moves with `next_fetch_at`).

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

> **ANSWERED 2026-08-18 (evening), and the answer inverts the ticket.** Two claims below are wrong.
> (1) "zero rows in all history" is true of `llm_call_log` **only** — the agent has **13
> `orchestration_states` rows** and 82 `site_work_items`, most recent run 08-18 07:33Z. That is the
> single-channel reading landmine #2 of §5 warns against, in this file's own §5. (2) The detector
> would NOT have answered it, for the reasons in the banner.
>
> **What is true:** two runs reached `load_candidate_pages`; one returned **exactly 15**, the cap.
> **Both then ended at `complete_no_candidates`.** `load_candidate_pages` declares
> `output_format: array`, so `QueryDatabaseAction` returns a bare slice with **no `count` key**
> (`database_actions.go:129`), while `check_candidates` tests `candidate_pages.count > 0`;
> `resolveFieldValue` fails all five strategies (strategy 5 needs a map, an array is not one),
> `ToFloat64(nil)` fails, and the numeric arm **returns false** — always the `else_step`. So
> `plan_links` cannot run, which is exactly why `llm_call_log` is empty. **The cap has never shaped a
> link decision because the linker has never made one.** Filed to the diagnosis loop before being
> asserted: `RUN_CORRELATION_ID=c4aa3559-86b1-4356-a28b-c71dfa661465`.

`internal-linker` has **zero `llm_call_log` rows in all history**, so whether its `LIMIT 15` has ever
shaped a link decision is **unmeasured** and 298 says so rather than guessing. **The LCO-009 detector
will answer this by itself** once `load_candidate_pages` runs — no separate work needed.

---

## 4. ⚠ THE LIMITATION THAT MAY DEFEAT 3b AND 3c ENTIRELY — read before trusting a quiet census

> **CORRECTED 2026-08-18 (evening): this section is right that the limitation defeats 3b and 3c, and
> wrong by ~3 orders of magnitude about its size — in the optimistic direction.** The window is not
> "time since the last pod restart" (hours to a day). The container log rotates **on size**, and the
> coordinator emits whole-state dumps (measured: mean 2.2 KB/line, worst single line **183 KB**), so
> it is wiped continuously while the pods keep running. **Measured 2026-08-18, both pods up since
> 18:00Z with 0 restarts: the oldest retrievable line was 3 s, 15 s, 21 s, 34 s and 91 s old across
> five samples.** There is no aggregator (`platform/logger/logger.go:37` — `OutputPaths: ["stdout"]`).
> **This was already in `LANDMINES.md`, filed 2026-08-08 at 0.4 s under load** — the lane built a
> log-only detector with that entry on file, and neither the design review nor the council round
> surfaced it. Its two stated remedies are the ones this session ended up rediscovering: attach
> `kubectl logs -f` *before* the event, or arrange a DB-visible observable — and one already exists
> (`collected_data`), which is why no code change is needed to census the caps.

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
