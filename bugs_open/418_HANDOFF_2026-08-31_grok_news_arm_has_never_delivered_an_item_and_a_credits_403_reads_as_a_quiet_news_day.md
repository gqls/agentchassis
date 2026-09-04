# 418 — the Grok news arm has never delivered an item, and a credits 403 reads as a quiet news day

**Filed:** 2026-08-31, `copy_quality_two_stage` lane (found while running the owner's model
trials, not by any news-lane symptom — the owner said *"We use Grok daily for the news"* and the
evidence said otherwise, in both directions at once).

**Diagnosis-loop note (owner ruling 2026-07-31):** filed on first-hand verification in place of a
`090` run, stated substitution: the failing API's own error body is captured verbatim in
`orchestration_states` rows (quoted below), the swallowing code path is quoted at file:line, and
the census is one GROUP BY — nothing here is inferred. Neighbouring lanes checked:
`bugfix_410_feed_phase_lock` owns CADENCE (when fetches fire — different mechanism, their census
keys on `next_fetch_at` and is unaffected by empty results); `bugfix_316_news_feed_ordering` owns
ordering. Neither owns delivery/error-handling. `LIKE` greps of `bugs_open`/`bugs_closed` for
grok/xai/credit: no prior filing.

## The two facts

**1. The xai provider went live 2026-08-30 14:55Z and has NEVER written an item.**
`[MEASURED 2026-08-31]`:

```sql
SELECT date_trunc('day', created_at)::date, count(*),
  count(*) FILTER (WHERE collected_data->'fetched_items'->>'error' IS NOT NULL) AS errored,
  sum(COALESCE((collected_data->'write_items'->>'written')::int,0)) AS items
FROM orchestration_states
WHERE collected_data->'input_data'->'source_config'->>'provider'='xai'
GROUP BY 1 ORDER BY 1;
--  2026-08-30 | 28 | 14 | 0
--  2026-08-31 | 28 | 14 | 0
-- min(created_at)=2026-08-30 14:55:30Z; 56 runs; ALL COMPLETED; items EVER: 0
```

**2. Every fetch is refused by x.ai, and the refusal reason is in our own rows** — the error body
BusyBox wget cannot show, the Go client captured:

```
xai API returned HTTP 403: {"code":"permission-denied","error":"Your team
d443dd72-09cf-4ba7-8209-1395f0edb4f0 has either used all available credits or
reached its monthly spending limit..."}
```

So the owner's *"we use Grok daily"* is TRUE of intent and schedule (28 dispatches/day) and FALSE
of delivery — and nothing anywhere said so. News pages keep filling from the RSS arm
(provider-less runs: 121 items written over the same window), which masks the dead arm
completely.

## Root cause — two, stacked

- **Operational:** the xAI team (`d443dd72-09cf-4ba7-8209-1395f0edb4f0`) is out of credits or at
  its monthly spending cap. Whether it exhausted on day one or arrived empty is not decidable
  from our side; only the xAI console (owner) can say. Probes from the pod confirm both
  `XAI_API_KEY` and `GROK_API_KEY` draw the same 403 (garbage key → 400, no-auth → 422, so the
  keys are recognised and refused — the account, not the wiring).
- **Structural (the durable half):** `fetchViaResponsesAPI`
  (`platform/orchestration/actions/feed_actions.go:431`) converts EVERY API failure into
  `emptyResult(errMsg)` — `{items: [], item_count: 0, error: "..."}` — and **returns it as a
  SUCCESSFUL action result**. The step completes, `write_items` writes 0, the orchestration
  COMPLETES. The error string rides inside `collected_data` where nothing reads it: no work item,
  no receipt, no failure row for the immune-system sweep (which only sees *recorded* failures —
  CLAUDE.md), no alert. **A total provider outage is indistinguishable from a day with no news.**
  This is the estate's standing pattern (a `complete` status is not the work having happened;
  a receipt nobody asserts on is a log line) expressed in the feed pipeline.

## Fix candidates, ordered by what closes the door

1. **Make the refusal a recorded failure** — on non-2xx, return an error (failing the step) OR
   write a `site_work_items`/`agent_error_log` row before returning empty, so the sweep and the
   dashboards can see it. The distinction that matters: "the API refused" must land in a table
   something READS. (An `error` key inside a COMPLETED run's `collected_data` is a receipt nobody
   asserts on — measured here: 56 of them, zero readers.)
2. **A zero-items streak check at the source level** — N consecutive empty fetches on a source
   that used to deliver (or a brand-new source that has NEVER delivered in its first M runs) files
   one work item. Catches every silent-refusal variant, including future auth/model/shape breaks,
   without parsing provider error formats.
3. **Owner tops up / raises the xAI limit** — necessary regardless, sufficient only until the
   next quota event; on its own it re-arms the silent failure.

## How to verify (post-fix and post-top-up)

- Top-up first: one xai run with `write_items.written > 0` and `fetched_items.error` NULL —
  that will be the FIRST successful Grok call in the platform's history.
- Fix first: induce the 403 (it currently self-induces on every run) and assert the failure
  lands in the table the fix names, not only in `collected_data`.

## Cross-references

- `bugs_open/410` (phase-lock slug — cadence; ambiguous number, resolve by slug) — unaffected by
  this, and this is invisible to it: their instrument reads `next_fetch_at`, ours reads delivery.
- The model-trials record of the same discovery from the trials side:
  `docs/agent_docs/docs024_key_docs_latest/copy_quality_two_stage/AUDIT_prompts/EXPERIMENT_2026-08-31_model_trials.md`.
- `WRONG_CALLS.md` 2026-08-31: the trials doc first claimed "the xAI path has never run live"
  off `llm_call_log` — the wrong instrument (the news path calls HTTP directly and never logs
  there). The conclusion happened to survive re-measurement at the right instrument; the census
  would have missed a WORKING arm entirely.

---

> **UPDATE 2026-09-04 (copy_quality_two_stage) — the OPERATIONAL half is resolved; the STRUCTURAL
> half is untouched and the bug stays open on it.** `[MEASURED 2026-09-04 ~11:30Z]` the xAI arm's
> first run that wrote items is `2026-09-03 15:06:22Z`; since then 10 runs, 45 items written, and
> the `fetched_items.error` key is NULL on every row in the window (`orchestration_states` is a
> rolling window — the 08-30/31 rows quoted above have aged out, so the 403 evidence now lives only
> in this file). So the owner funded team `d443dd72-…` on the afternoon of 09-03. Nothing in
> `fetchViaResponsesAPI` changed: a refusal would still complete as an empty result nobody reads.
> The Grok WRITER trial that this bug was found while running has now run — results in
> `docs/agent_docs/docs024_key_docs_latest/copy_quality_two_stage/AUDIT_prompts/EXPERIMENT_2026-08-31_model_trials.md`
> (Grok arms section). Query used:
> ```sql
> SELECT min(created_at), count(*), sum(COALESCE((collected_data->'write_items'->>'written')::int,0))
> FROM orchestration_states
> WHERE collected_data->'input_data'->'source_config'->>'provider'='xai'
>   AND COALESCE((collected_data->'write_items'->>'written')::int,0) > 0;
> ```
