# bugs_closed/059 — code_symbols index goes stale silently (no reindex cadence)

> **CLOSED 2026-07-24 — fixed AND live AND verified on v1.0.1155.**
> Fix #1 (cadence): `scheduled_tasks` `code-index-refresh`, live 2026-07-24,
> proven on its first fire (derived ref, `ca8dc7f`→`adb00fd`). Fix #3 (read-time
> freshness guard): council-APPROVED (corr `8ed67200`), live in v1.0.1155, and
> verified by a LIVE probe — a real gather on the deployed pod rendered
> `(index freshness: refreshed 3h ago at commit e19aa5d)` above a code answer
> (probe corr `24365c79`, scratch agent `diagnose-wiring-probe-ok`, deactivated
> and kept as evidence). The stale/empty/error branches are falsification-tested
> in `TestFreshnessBanner`; a pattern-check declared pair
> (`FROM code_symbols` ↔ `codeIndexFreshness`) guards future consumers.
> Transferable pattern filed in 016b §9 ("An indexed snapshot read by a
> correctness check is a silent freshness dependency").
> **Residual, deliberately left open as a docs item, not a defect:** fix #2 —
> the docs019 `TRIGGER_code_indexer_v2.sh` still documents the wrong (direct)
> dispatch; a hand reindex should use `index-orchestrator`. The cadence makes a
> hand reindex rarely necessary.

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

## UPDATE 2026-07-24 — fix #1 (reindex CADENCE) shipped & live

`scheduled_tasks` row **`code-index-refresh`** applied live (no image roll):
`SEED_code_index_refresh_cadence.sql` in `docs024/…/fixloop_eg_dartsonline/`. Fires
`index-orchestrator` every 24h. The load-bearing choice was **not to hardcode a
ref** (that would BE this bug's own class — a snapshot that drifts): the ref is
DERIVED at fire time by the task's `pre_query` — the most-recent human/diagnosis-
driven feature-branch ref (`NNN_*`), excluding the reindex's own runs
(`owner_agent_type NOT IN ('index-orchestrator','code-indexer')`) so it can't pin
itself to a dead branch. The scheduler shallow-merges `{ref}` into `input_data`
(`cmd/scheduler/main.go`); no rows → tick skipped (safe). Verified: row enabled,
pre_query resolves to the live branch. First-fire behaviour verification: see NOTES
turn 48b / the session monitor.

RESIDUALS (bug stays OPEN):
- ~~**fix #3 — the freshness guard is NOT built.**~~ **BUILT & COUNCIL-APPROVED
  2026-07-24** (commit `f21e54687` + follow-up `4920cd629`; council corr `8ed67200`
  APPROVED, 2 advisory objections both applied). Every rendered answer from
  code_symbols (diagnose gather + council code_checks incl. the prior_art seat) now
  carries a freshness banner: quiet age+commit when fresh, loud STALE/EMPTY warning
  (>48h = one missed 24h-cadence fire) naming age/commit/remedy, fail-open note on
  a query error. Plus a pattern-check declared pair (`FROM code_symbols` ↔
  `codeIndexFreshness`) so a future third consumer cannot silently skip the guard.
  **INERT until the next chassis image roll** — the bug stays open until the guard
  is verified IN A POD (fixed-and-live bar), then it can close with fix #2 noted.
- **fix #2 — the docs019 `TRIGGER_code_indexer_v2.sh` is still the wrong (direct)
  dispatch.** A hand reindex should go via `index-orchestrator`; the stale script
  wastes a run. Left for a docs sweep.
- The cadence is 24h (matches other freshness tasks). Tighten the one column if the
  index must stay within a single image roll (images roll several times a day).
