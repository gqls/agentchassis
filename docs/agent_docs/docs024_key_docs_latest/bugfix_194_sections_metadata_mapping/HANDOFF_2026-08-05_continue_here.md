# HANDOFF — `bugfix_194` lane, 2026-08-05 10:05Z · **read this first, it is the whole state**

**The bug is CLOSED** (`bugs_closed/194_HANDOFF_2026-08-04_four_of_six_…`). Council
**APPROVED r1**, correlation `b6023fc1-ae70-4486-b752-d399e9b1afcc`. Everything is committed
and live. **What is left is two falsification checks, not work.** Do not re-open, re-plan or
re-fix — read §3, run the two checks, and record what they say.

---

## 1. What the bug was, in four lines

`save_page_sections` took the path to its own structured input from a per-caller config key.
Four of six callers never set it, so those saves fell through to regex HTML-parsing, wrote
`content_data = NULL` — the only thing `rerender_page_sections` can regenerate a section
from — and **reported success on a page that serves perfectly**. Live since 2026-02-18.

## 2. What shipped — all of it live, nothing pending

| half | what | state |
|---|---|---|
| config | seed **312**: `pageflow-builder` + `site-work-orchestrator` map `page_content.response.sections_metadata` | **LIVE** on apply 08-04 |
| config | seed **313**: `tool-recreation-handler` declares `expects_no_sections_metadata: true` | **LIVE** on apply 08-04 |
| Go | `save_sections_metadata_source.go` + 3 wiring points in `save_page_sections_action.go` (`47ee3ebce`) | **LIVE on v1.0.1252**, pod-verified both replicas 08-05 09:59Z |

**All six callers are now explicit** — five name the field, the sixth declares it has none.
Re-census any time with RUNBOOK **R1** (⚠ needs `jsonb_path_query('$.**.steps')`; the step is
inside a loop `sub_workflow` in four of six and a top-level `jsonb_each` finds only two).

The Go half adds: a **shared default** (`validate_page_content_stats.go`'s own
`defaultSectionsMetadataField`, *referenced not copied*), `expects_no_sections_metadata`,
`require_sections_metadata` (**opt-in refusal, seeded on NOBODY**, per RFC_010), and a
`CONTENT_DATA_REGRESSION` `agent_error_log` record. Registered as **PBP-031**.

## 3. THE TWO CHECKS THAT ARE OWED — this is the job

### 3a. The 24h no-regression read — **due 2026-08-06 ~09:10Z**, cheap, do this first

> ## UPDATE 2026-08-05 20:48Z (T+11h38m) — effectively PASSING, and now on real traffic
>
> | | 08-05 10:16Z (T+1h) | **08-05 20:48Z (T+11h38m)** |
> |---|---|---|
> | `CONTENT_DATA_REGRESSION` | 0 | **0** |
> | positive control, same run | 138 `PROCESSING_FAILED` | **476 `TIMEOUT`** + 4 others |
> | `page-rerender` runs | 3630 (**1** post-roll) | 3750 → **+120** |
> | `page-build-handler` runs | 355 (**0** post-roll) | 389 → **+34**, last 15:43Z |
>
> **This morning's zero was near-vacuous and was recorded as such** — one caller had run once
> and the other not at all. That objection is now answered: **~155 runs of the two callers on
> the fixed binary, still zero regressions**, against a positive control proving the table is
> being written. This is the substantive result; the 09:10Z tomorrow reading is a formality.
>
> ⚠ **A SECOND ROLL happened at 20:41Z — `v1.0.1254`, both replicas.** It does not reset this
> check: 194's fix shipped in `v1.0.1252` and is still present in `1254`, so the window spans
> two builds that both carry it. Note it when you write the result up rather than implying one
> continuous binary.
>
> **Still do the formal read tomorrow** — and re-check the run counts in the same breath, since
> the whole point of this section is that a zero without traffic means nothing.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -F'|' -c "
SELECT agent_type, count(*), max(occurred_at) FROM agent_error_log
WHERE error_code='CONTENT_DATA_REGRESSION' GROUP BY 1;
-- POSITIVE CONTROL, same run — without it a zero is not evidence:
SELECT error_code, count(*) FROM agent_error_log WHERE occurred_at > '2026-08-05 09:10:00+00' GROUP BY 1 ORDER BY 2 DESC LIMIT 5;"
```

- **`occurred_at`, NOT `created_at`.** The original version of this query named a column that
  does not exist and therefore ERRORED instead of returning zero — see `WRONG_CALLS.md`
  2026-08-05. A check that errors has no result to disagree with.
- **PASS:** zero rows for `page-build-handler` and `page-rerender`.
- **FAIL / STOP:** *any* `page-rerender` row. That means my report predicate is misconceived
  (~320 runs/day would flood it) — **do not proceed to any `require_sections_metadata`
  opt-in**, and re-read `shouldReportContentDataLoss` before anything else.
- **Baseline taken 08-05 09:59Z: 0 rows fleet-wide**, against a positive control of 102
  `PROCESSING_FAILED` in the same window. So a later non-zero is a real signal.
- ⚠ **A zero is only meaningful if the callers actually RAN.** In the first 50 minutes after
  the roll there were **zero** runs of any of the six. Confirm traffic before reading the
  zero as a pass:
  `SELECT agent_type, run_count, last_ran_at FROM agent_run_stats WHERE agent_type IN ('page-rerender','page-build-handler') ;`
  — compare `run_count` against the pre-roll figures below.

### 3b. The live acceptance on a seeded caller — **BLOCKED BEHIND `bugs_open/201`. Do not hunt for a target; there isn't one.**

> **STATUS 2026-08-05 (later) — this is no longer "needs a decision about which site".**
> Three independent constraints were measured today and each alone is fatal:
> 1. **No loadable item.** The loop needs `handler_agent='page-content-writer'` +
>    `status IN ('triaged','approved')`. Fleet-wide that is **one** item; the seven-site
>    candidate list below is 0 across the board (struck through, see the correction).
> 2. **That one item's site is LOCKED** — `mortgagecalculator.co.uk`, since 08-03, by the
>    adoption lane, "held pending owner decision on page rebuilds". A locked site returns
>    success with zero items (`skipped_reason: site_locked`), so it is a *silent* vacuous
>    pass. **Do not release it** — `aee11cb90` is the incident that lock exists because of.
> 3. **Even unlocked, that item is predicted to hard-fail upstream of the code under test.**
>    `bugs_open/201`: `page-content-writer` called without `section_plan` fails at
>    `fail_no_ready_sections` on an already-built page (11 of 11). `site-work-orchestrator`'s
>    `write_page_content` **omits `section_plan`** — verified live — so this route has 201's
>    defect too, and 201 §3 already names this exact item as one that will fail.
>
> **So: a fix to 201 unblocks 3b.** Contributed there (CONTRIB 2026-08-05) rather than forked.
> The pincer, with the producer table, is in NOTES 08-05. **`ai-agent-orchestration.com` is
> UNLOCKED and does need a rebuild — but it has 0 loadable items on two independent clauses,
> so that work is real and is NOT this check.**

This proves seed 312's mapping resolves on a caller that had never had it.

**⚠ The route this lane originally named was wrong TWICE. Do not follow the 08-04 text.**

1. `scripts/initial_messages/170_work_item_flow_build/075d_simple_maintain_trigger.sh`
   **cannot execute** — line 9 is a bare `-------------------` (aborts under its own
   `set -euo pipefail`), line 11 hardcodes `DOMAIN="finetuning.uk"` over the argument line 7
   demands. Committed that way in `5345ad7e2`. **Do not fix it in passing** — it is the
   finetuning lane's file and the hardcode may be theirs deliberately. Use its kcat envelope
   as a template and publish your own message.
2. `mode=maintenance` only reaches the save step **if the site has queued build work**:
   `check_mode` → *(maintenance)* `select_style_collection` → `set_default_components` → … →
   `load_work_items` → `check_has_items` → **`build_items_loop`** *(has items)* /
   `load_fix_items` *(none)*. Only the first branch contains `save_sections`. **An idle
   domain completes green having never run the code under test — a vacuous pass.**

**Preconditions for a real run:** a domain with open build-routed work items, pages at
`rebuild_policy != 'owned'` (087's run was blocked by exactly that guard), and **a lane that
is not already working that site**. ~~Candidates measured 08-05 (open build-routed items /
page policy): `leopardessconsulting.co.uk` 17 (40 generic, 5 owned), `robot-hands.com` 14,
`mortgagecalculator.co.uk` 13, `finetuning.uk` 13 (40 generic, 6 owned),
`ai-agent-orchestration.com` 6 (32 generic, 6 owned), `idea.uk` 7, `fundamentallyai.com` 6.~~

> **CORRECTED 2026-08-05 (later) — that candidate list measured the wrong quantity and every
> number in it is 0 against the predicate that actually gates the loop. Do not use it.**
> "Open build-routed items" is not what `load_work_items` loads; it needs
> `status IN ('triaged','approved')` **and `handler_agent='page-content-writer'`**
> (`load_work_item_actions.go:623-661` plus the step's own filters — see **RUNBOOK R7**,
> which now carries the exact query). Fleet-wide **exactly one item qualifies**:
> `mortgagecalculator.co.uk`, 1 × `literal_markdown` — the site this list ranked at 13.
> `ai-agent-orchestration.com` is **0**, failing two clauses independently. So firing at any
> site on that list produces the **vacuous pass this very section warns about**, and
> `page-content-writer` has held only **14 items fleet-wide, ever**. Logged in `WRONG_CALLS.md`.

> **NOT FIRED, deliberately.** Every candidate is a live customer site and at least four are
> actively owned by other lanes right now. A `site-work-orchestrator` maintenance run
> rebuilds pages with real LLM spend on somebody else's site — that is an owner's call, not
> a verification convenience. **Ask before dispatching**, and run `scripts/who-owns.py` plus
> the open-work-item check on the target first.

**PASS requires BOTH:** `content_data` non-NULL on every row at the new run's `updated_at`,
**and** the save step's result carrying `sections_source: 'metadata'`. The second half is not
decoration — `content_data` can also arrive via the Layer-2 interactive carry-forward, so the
bare column check is a **false pass**.
**DISCONFIRMING:** still NULL, or `sections_source: 'html_parse'` — the writer's reply is not
reaching the save on that path and the key name is not at fault.

## 4. What stays OPEN, and is a human's decision — surface it, do not quietly resolve it

The council's `bug_historian` seat, medium: **for every live caller today a loss is now
RECORDED, not PREVENTED.** `require_sections_metadata` is seeded on nobody by design
(RFC_010: new authority ships with the unsafe default OFF), and the record is what turns the
later per-caller opt-in into a measurement instead of a guess. **But a deferral is not a fix**
and whether it closes 194 is a judgement. The 3a reading is the input to that decision.

Also disclosed and unfixed: **161 of 1,201** `page_components` rows already NULL (repair is
re-running the build, **never** restoring `page_component_history` — its `component_id` is
NULLed by `ON DELETE SET NULL`, so pairing yesterday's content with today's HTML makes the
next rerender reinstate the old page); **partial** `content_data` loss, outside the report's
predicate on purpose; and `bugs_open/136`'s single-component writers.

## 5. Numbers a successor will otherwise re-derive, with their traps

- **Callers:** 6. **Not 3** — the concept register said three until 08-04 and that stale count
  is part of why this stayed invisible; PBP-011 now carries the correction.
- **Pre-roll run counts** (`agent_run_stats`, span 07-26 → 08-04): `page-rerender` 2878,
  `page-build-handler` 283, `page-rebuild` 4, `tool-recreation-handler` 3,
  **`pageflow-builder` and `site-work-orchestrator` ABSENT** (dormant — which is why the
  proof is offline: 7 tests, 4 mutations run against the shipped code).
- **Do NOT use `orchestration_states` for dormancy** — it retains ~1 day of `COMPLETED`.
  `agent_run_stats` has no reaper. (And it does track orchestrator-shaped agents, so an
  absence there means something.)
- **Cost of a NULL:** 44 `needs_page` escalations across 8 sites since 07-12 whose reason is
  "a section had no stored content_data", 13 FAILED on 08-03. **Exposure for the class, NOT
  damage attributed to these callers** — some of those pages predate `content_data` capture.
- **The new COUNT(\*) is bounded:** `idx_page_components_page` exists; max 8 rows per page,
  mean 2.3 across 520 pages.

## 6. Lane files

`PLAN_2026-08-04` (the five decisions **and the two shapes rejected**) ·
`RUNBOOK` (R1 census, R3 dormancy, R6 pod-grep, **R7 the two checks**) ·
`NOTES` (evidence + every misstep) · `README_where_we_are` (owner's plain prose) ·
`SUMMARY_2026-08-04`. Fleet-wide: `LANDMINES.md` (the absence of `sections_metadata_field` no
longer means the HTML path), `WRONG_CALLS.md` (**four** entries from this lane), `016b` §10
row 194, concept register **PBP-031**.
