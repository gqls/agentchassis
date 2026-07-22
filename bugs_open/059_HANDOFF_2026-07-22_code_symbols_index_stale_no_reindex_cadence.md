# bugs_open/059 — code_symbols index goes stale silently (no reindex cadence)

*Filed 2026-07-22 (chat "diagnosis fixloop 5"). Found while proving the diagnosis
code tier: the tier is live, but the index it reads was 3 weeks stale, so every
code-existence check silently answered against old code.*

## Symptom / impact

`code_symbols` (the indexed snapshot the diagnosis code tier and the council's
`prior_art_librarian` seat both query) was last indexed **2026-07-02 at commit
`e3176f8`** and had not been refreshed for ~20 days. It only refreshes when the
`code-indexer` workflow is run manually; nothing schedules it.

Two consumers read it, and BOTH silently degrade when it is stale:

1. **Diagnosis code tier** — `answerCodeCheck` / `lookup_code_symbols`
   (`platform/orchestration/actions/diagnose_code_lookup_action.go`) `SELECT ...
   FROM code_symbols`. A `code_request` about any symbol added since the last
   index returns **empty**, which the tier renders identically to "mechanism
   absent" — the documented worst-case (`diagnose_load_runtime` comment: "treat a
   stale or empty answer as 'unknown', NOT as 'absent'", but the verdicter has no
   way to tell empty-because-stale from empty-because-absent).
2. **`prior_art_librarian` council seat** — queries `code_symbols` for its
   existence checks. Its entire purpose is the ASSERTED-ABSENCE / DORMANT-MACHINERY
   class ("capability X does not exist / needs building"). Against a 3-week-old
   index it will **confirm a false absence** for anything built in the last three
   weeks — exactly the failure it exists to catch. (See bugs_closed/031, the
   adjacent stale-register face.)

## Evidence

```
-- before refresh (2026-07-22):
SELECT count(*), count(DISTINCT commit_sha), max(updated_at) FROM code_symbols;
--   3723 | 1 | 2026-07-02 19:01:56   (commit e3176f8)
-- recurring-cap symbols added after 2026-07-02: NOT present
SELECT symbol FROM code_symbols WHERE symbol IN
  ('withPriorCodeRequests','workflowRefsFromRuntime','validateRouteWiring');  -- 0 rows
-- only "index" orchestration in 30 days was a FAILED direct dispatch:
SELECT correlation_id, status, created_at FROM orchestration_states
  WHERE current_step IN ('index_symbols','request_analysis')
    AND created_at > now() - interval '30 days';
--   3fba7f65... | FAILED | 2026-07-22 15:05  (my own first attempt, see below)
```

## Root cause

**No reindex cadence.** `code_symbols` is refreshed only by the `code-indexer`
workflow (`request_analysis`=`analyse_repo_local` → `index_symbols`=
`index_code_symbols`), which must run in a SPAWNED repo-cloning pod (the token is
injected by `isRepoCloningAgent`, `spawn_actions.go`). Nothing triggers it on a
schedule or on image roll, so it drifts from the deployed code indefinitely.

**Secondary: the documented manual trigger is outdated.** `TRIGGER_code_indexer_v2.sh`
(docs019) dispatches `agent_type=code-indexer` DIRECTLY to `system.agent.generic.requests`.
That gets adopted in-place by a chassis pod, which has no `GITHUB_READ_TOKEN`, so
`analyse_repo_local` fails immediately ("must run in a SPAWNED repo-cloning agent
pod ... trigger via a spawning orchestrator (e.g. index-orchestrator) instead").
The correct entry is **`index-orchestrator`** (`spawn_indexer` → `call_indexer`,
1800s timeout), which spawns the code-indexer as its own pod with the token.
Verified 2026-07-22: direct dispatch FAILED; index-orchestrator dispatch COMPLETED
and brought the index current (4486 rows at `ca8dc7f`, `updated_at` today).

## Fix candidates

1. **Schedule the reindex** (immediate, structural). A `scheduled_task` firing
   `index-orchestrator` on a cadence — best tied to image roll (that is exactly
   when the deployed code changes), or daily as a floor. The code tier is only as
   good as the index's freshness; today it is a manual step nobody owns.
2. **Fix/replace the manual trigger** so a hand reindex works first try: point it
   at `index-orchestrator`, not `code-indexer` directly. (The docs019 v2 script is
   stale and wastes a run.)
3. **Freshness guard (the "don't read absence as an answer" principle applied to
   the index).** The code tier and the `prior_art` seat could compare the index's
   `commit_sha` / `max(updated_at)` against the diagnosis REF (or now()) and, when
   the index is materially behind, DEGRADE the answer to "unknown — index stale at
   <sha>, run index-orchestrator" instead of rendering empty-as-absent. This is the
   same shape as bugs_open/019's tolerate-truncation fix: a missing signal must not
   read as a negative one. The verdicter and the guardian seat both act on
   "absent"; that must be trustworthy.

## How to verify (after any fix)

```
SELECT count(*), (SELECT DISTINCT commit_sha FROM code_symbols LIMIT 1) AS sha,
       max(updated_at) FROM code_symbols;   -- sha recent, updated_at fresh
```
Refreshed manually 2026-07-22 to `ca8dc7f` as part of proving the code tier — but
that is a one-off; without a cadence it will drift again. The bug stays OPEN until
a cadence (or a freshness guard) ships, because the defect is the drift, not the
one stale snapshot.

## Transferable pattern (for 016b §9)

*An indexed snapshot that a correctness check reads is a silent freshness
dependency: a stale index answers "absent" identically to a genuine absence, so a
check that trusts absence (existence checks, drop-reporting, coverage) is only as
correct as the index is fresh — and nothing surfaces the staleness at read time.*
Same family as 019 (a truncated/missing signal read as a negative one).
