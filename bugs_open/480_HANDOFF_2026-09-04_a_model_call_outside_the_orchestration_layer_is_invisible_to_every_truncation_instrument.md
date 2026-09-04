# 480 — a model call outside the orchestration layer is invisible to every truncation instrument the estate has

**Filed** 2026-09-04, by the `bugfix_257_token_budget_at_the_client` lane, at the owner's direction.
**Owner decision 3, 2026-09-04:** *"own piece of work, let me know the bug number."* This is that number.

**Lane:** `docs/agent_docs/docs024_key_docs_latest/direct_caller_llm_observability/`
**Parent:** `bugs_open/257` §DECISION 3. 257 keeps the budget-*resolution* half; this file has the
*observability* half, and 257 no longer carries it.

---

## 1. What is wrong, in plain terms

When the platform asks a language model for something, one function writes a row into a table called
`llm_call_log`: which agent, which step, how big a reply it asked for, how big a reply it got, and
whether the answer was cut off.

Every question the estate can ask about truncation is a query over that table. "Which steps are running
out of room?" is a query over that table. The two scheduled checks that watch for token pressure are
queries over that table.

**A model call that does not go through the orchestration layer writes no row.** It is not recorded as
healthy and not recorded as failing — it is simply absent, and an absence in that table is
indistinguishable from a step that never ran.

## 2. The census

[MEASURED 2026-09-04] `grep -rln "llm_call_log" --include=*.go platform/ internal/ pkg/` returns files
in **`platform/orchestration/actions` and nowhere else**. Every writer of the table lives in one package.

Model callers outside that package, all of which are therefore unlogged:

| file | how it sizes the call | logged? |
|---|---|---|
| `internal/tools-api/handlers/defend.go:101` | literal `8192` in the options map | stdout only |
| `internal/tools-api/handlers/position.go:90` | literal `8192` in the options map | stdout only |
| `internal/tools-api/handlers/gripper.go:160` | `gripper.MaxTokensPerReply`, a Go constant | stdout only |
| `internal/agents/contentcreator/agent.go:720,755,757,762` | literals `3000` / `1200` / `6000` / `100`, with one config read at `:732` | no |
| `internal/agents/reasoning/agent.go:127` | passes `nil` — inherits `ai_service.max_tokens` via 257 Path A | no |
| `internal/tools-api/handlers/playground.go:156` | `cfg.MaxTokens` — genuinely config-driven | stdout only |

`internal/tools-api/handlers/ailog.go` is the partial exception and is worth reading before designing
anything: it already reports truncation **distinctly** from other failures, for the tools-api handlers,
on the stated reasoning that truncation "is NOT an upstream fault but our own configured cap, and it
needs a different fix". It writes to stdout via the standard library logger, so
`docker compose logs tools-api` sees it and no fleet query ever will.

## 3. Why this is its own lane and not part of 257

Three weeks inside 257 and it never started, which is the practical argument. The structural ones:

- **A different blast radius.** 257's ladder is one package. This touches every direct caller in the
  estate, including two services (`tools-api`) that do not share the orchestration layer's database
  handle at all.
- **A different kind of change.** 257 removed literals and corrected a precedence rule. This ADDS
  writes — a new caller of a logging seam, on paths that currently have none — which is a different
  risk profile and a different review.
- **A stated trap of its own, from 257 round 2.** More reporting is not automatically more truth. A
  logged number that a hardcoded default could equally have produced tells you nothing: `offer-analyser`
  declared `ai_service.max_tokens: 2000` on the two steps whose Go literal was ALSO `2000`, so
  `llm_call_log.max_tokens` read 2000 whether the configuration was honoured or dropped on the floor,
  and no query over the fleet's own instrument could separate the two hypotheses. Whatever this lane
  builds must be able to distinguish those, or it will manufacture confident wrong answers at scale.

## 4. What a first session should do

1. **Re-run the census, do not quote §2.** It is dated for a reason: a census goes stale BY ADDITION and
   reads as current for ever (owner ruling 2026-08-22). `git log --since=2026-09-04 --diff-filter=A --
   internal/ platform/` lists what has been added since; a non-empty result means re-count first. Census
   the CONCEPT, not the interface — a caller reaching a provider over raw HTTP has no `GenerateText` in
   it, and four successive censuses of 257 missed `feed_actions.go` exactly that way:
   `grep -rnE '"(max_tokens|max_output_tokens|maxOutputTokens|num_predict|max_completion_tokens)"' --include=*.go . | grep -v _test.go`
2. **Read `platform/orchestration/actions/llm_call_logger.go` before designing a seam.** The logger
   already exists; the question is whether it can be reached from `internal/` and `tools-api` without an
   import cycle, and that question has a known shape — `platform/orchestration/actions` imports
   `platform/aiservice`, so anything pushed DOWN into `aiservice` cycles.
3. **Decide whether the budget guard extends too, or state that it will not.**
   `platform/orchestration/actions/llm_budget_call_sites_test.go` binds "no hardcoded budget" at the
   PACKAGE, and a Go test is package-scoped by nature. Covering `internal/agents/*` and `tools-api`
   needs either one test per package or a check in `scripts/pattern-check.py` — and note that
   `pattern-check.py` is in council scope as of 2026-08-23.
4. **Tell the two live consumers what changes.** `fleet-step-token-pressure` and
   `council-seat-token-pressure` (both enabled, both every 21600s) are queries over `llm_call_log`;
   adding rows changes what they report. Owner ruling 2026-07-29 §3: a shared mechanism's other
   consumers must be told, not merely measured.

## 5. What is NOT in scope

The budget-*resolution* rule. That is 257, it is fixed, and the ladder is live in code
(`platform/orchestration/actions/llm_options.go`, commit `d88afbf84`) with the configuration migrated
(`769`, applied 2026-09-04). This lane is about whether a call that has already chosen its budget is
VISIBLE, not about how it chose.

## 6. Related

- `bugs_open/257` — the budget contract; §DECISION 3 is where this was spun out from.
- `bugs_open/205` — 8 of 126 active LLM steps ran with no configured budget and 64 truncations happened
  before anything said so. That count came from `llm_call_log`, so it was a count of the callers that
  DO log; the true fleet figure has never been measurable.
- `bugs_open/138` — the council-seat truncation lineage, and the reason a truncated call is counted from
  `error_message ILIKE '%stop_reason=max_tokens%'` and **never** from `output_tokens >= max_tokens`: a
  truncated call has `output_tokens` NULL, and the obvious form undercounted 94 truncations to 4.
- `bugs_closed/076`, `012`, `046` — the silent-truncation lineage.

---

## CONTRIB 2026-09-04, `copy_quality_two_stage` lane — the table has a SECOND blindness: a row that exists and a column that cannot record what it names

Invited by the 257 lane, who declined to fold it in themselves so it would be tracked where it
belongs. **This is not the same defect as §1 and does not change its census** — §1 is about a call
that writes NO ROW. This is about a call that writes a row, in the normal way, through the
orchestration layer, where one column is structurally incapable of holding the thing it is named
after. Both make the same table answer confidently and wrongly, which is why it belongs here.

**`llm_call_log.thinking_tokens` cannot see extended thinking on `claude-sonnet-5`.** The column is
populated from parsed thinking-block TEXT, and Sonnet 5 returns thinking blocks **empty** by default
(`display: "omitted"`, a silent change from 4.6 where summaries came back). The model reasons; the
blocks arrive empty; the column records nothing. `[MEASURED 2026-09-04]`:

```sql
SELECT count(*) FILTER (WHERE thinking_tokens > 0) AS with_thinking, count(*) AS total
FROM llm_call_log WHERE agent_type='page-content-writer' AND created_at > now() - interval '7 days';
--  0 | 6724

SELECT count(*) FILTER (WHERE thinking_tokens > 0), count(*) FROM llm_call_log
WHERE created_at > now() - interval '30 days';
--  1 | 40716      -- the ONE non-null row in 30 days fleet-wide is a Gemini call
```

`page-content-writer` runs adaptive thinking on **every one** of those 6,724 calls: the client omits
`thinking` (`platform/aiservice/anthropic.go:278-300` sends it only when `budget_tokens` is set), and
omitting it on Sonnet 5 means adaptive is ON. So the column reads 0 for an agent that reasons
constantly. **Fleet-wide it is not measuring which agents reason — it is measuring which PROVIDER
returns thinking summaries**, and today that is Gemini, once.

**Why it is worth a line in this bug rather than a note somewhere.** A zero in a column named
`thinking_tokens` does not look like missing instrumentation. It looks like a finding, and it is the
finding a reader wants: it reads as "this agent does not use extended thinking", which is exactly the
sentence I wrote into a lane document before the API reference contradicted me. §1's absence at least
*looks* like an absence; this one looks like data.

**The honest instruments, both available today.** `output_tokens` against the visible text — the
benchmark row spent 2,713 output tokens to return ~1,100 tokens of JSON, and that gap IS the
reasoning. And a replay with `"thinking":{"type":"adaptive","display":"summarized"}`, which returns
the summary text the column wants; two such replays this morning carried 2,334 and 761 characters of
summary while the production column stayed null.

**Sizing note for whoever takes §4.** If a fix populates `thinking_tokens` from the API's own
`usage`, it is worth checking whether the same call site can record `stop_reason` verbatim. A related
trap from the same session: at `output_config.effort: "max"` a writer prompt spent its ENTIRE budget
on thinking and returned zero characters of text, four times, at `max_tokens` 16,000 and again at
40,000, `stop_reason: max_tokens` each time. A row like that is indistinguishable from a successful
call in every column this table currently exposes, and any prose measure over its output scores it as
a perfect zero (`LANDMINES.md`, *"A model can spend its ENTIRE token budget on thinking and return NO
text"*). The `fleet-step-token-pressure` check is the live instrument closest to it.

**Not ours to fix and we have not started on it** — filed here so it is tracked with the rest of the
table's blind spots. Evidence trail:
`docs/agent_docs/docs024_key_docs_latest/copy_quality_two_stage/NOTES_two_stage_copy.md` 2026-09-04,
`WRONG_CALLS.md` 2026-09-04 (the claim it produced, and the check that would have caught it).
