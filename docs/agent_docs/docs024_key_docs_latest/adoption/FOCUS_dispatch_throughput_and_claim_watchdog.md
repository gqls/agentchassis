# FOCUS: Dispatch throughput, the claim-timeout watchdog (Lever C), and OOM safety

**Status:** Investigation complete; **Lever C (claim watchdog) selected to build first.** Levers A/B and guardrails scoped below as follow-ups.

**Builds on:** `FOCUS_dispatch_diagnostic_4` (the dispatcher architecture and the `detected→triaged→claimed` state machine). That doc is the architecture reference; this doc is the throughput-improvement plan and corrects two facts from it with live data. Do not duplicate `_4_`; read it for the selection SQL and state machine.

---

## Problem

Sites with many tool/game pages drain extremely slowly (hours) and can look stalled. Surfaced verifying bug A1 on the `e9609749` adoption of gamesdesign.co.uk (11 `needs_tool_recreation` items): tool-recreation claims were 20–67 min apart (14:41, 15:04, 16:11, 16:58), with long idle stretches, and two items logged "Claim timed out — handler pod likely died." This is **not** an A1 correctness problem (the per-tool build/deploy works once it runs) — it is dispatch throughput.

---

## Corrected facts (live data overrides the docs)

The live `scheduled_tasks` row for `build-pipeline-trigger`:

| field | doc (`010` table) | **live** |
|---|---|---|
| `interval_seconds` | 120 | **30** |
| `max_concurrent` (group `dispatch`) | 2 | **8** |
| `timeout_seconds` | — | **300** |

So the trigger cadence is **not** the bottleneck — 30s ticks, 8-wide. (`last_completed_at` was 0.6s after `last_triggered_at` on inspection = an idle tick; the fast path.)

The actual throughput limiters for a single multi-tool site, all confirmed from the agent definitions and `_4_`:

1. **Per-site `NOT EXISTS` is an absolute exclusion.** `find_dispatchable_site` excludes a site entirely while *any* of its items is `status='claimed'`. So tools dispatch **serially** within a site — each held ~5 min for its Opus build. 11 tools serial ≈ an hour minimum even with infinite trigger capacity.
2. **Claim-timeout freezes.** A dead handler leaves an item `claimed` with stale `claimed_at`; per (1) that excludes the whole site until the claim is reset. This produced the 47–67 min gaps. See the deadlock note under Lever C.
3. **Deterministic-by-`site_id` ordering + `LIMIT 1`.** `DISTINCT ON (site_id) … ORDER BY site_id … LIMIT 1` has no fairness ordering: the lowest-UUID eligible site wins consistently, so a steadily-busy lower-UUID site can **starve** a higher-UUID one, and 8 concurrent triggers tend to converge on the *same* one site (the others find its items already claimed and do nothing) — the 8-wide concurrency is partly wasted and partly unfair.
4. **Scheduler timeout mismatch (also an OOM risk — see below).** Scheduled-task `timeout_seconds=300` but the trigger's `call_dispatch` waits 900s and `build-dispatch-loop` runs to 4200s with `call_handler` at 1200s. The scheduler stops counting a trigger as "in-flight" at 300s while its work runs for minutes more, so it fires fresh triggers and concurrent dispatch loops/handlers accumulate past the nominal cap of 8.

Stale-description note (the kind that misleads): `build-dispatch-loop`'s definition is described as "Processes one work item per invocation, then spawns itself if more remain. No loops or sub_workflows." The actual workflow **uses a `process_item` loop**, loads `max_items=5`, iterates `max_iterations=5` sequentially (`continue_on_error=false`), and has **no self-spawn** — re-invocation is entirely the trigger's job. Decisions based on the description would be wrong.

---

## The git deploy path and its concurrency (gates Lever B)

`git_commit` (`GitCommitAction`) does **no local git** — it publishes one message to `system.adapter.git.requests` (Kafka key = `correlation_id`) with `AwaitResponse:true` and returns. The commit happens in the **git-adapter**.

The git-adapter (`github_client.go` `CommitToRepo`) uses the GitHub Git Data API: `createOrGetRepo` → `getLatestCommitSHA` → `createBlob`×N → `createTree(base_tree = latestSHA)` → `createCommit(parent = latestSHA)` → `updateRef(force:false)`. It is a read-modify-write with **`force:false` and no retry**. The repo is a single shared `"sites"` repo (default `repo_name`), with files prefixed `{domain}/{path}`.

Consumer model (`adapter.go`): a **single sequential consume loop per replica**; **two replicas** in group `git.adapter.group`, so partitions split across them.

Consequences:
- **Same adoption → same `correlation_id` → same partition → one replica → commits applied sequentially.** So **same-site parallel builds (Lever B) are git-safe**: commits queue and each reads the updated HEAD. Commit latency is a few GitHub API calls (~1–2s) each, negligible against ~5-min Opus builds, so serial commits within a site are not a throughput problem.
- **Different sites → different `correlation_id` → can hit different partitions/replicas concurrently → race on the shared `sites` repo.** With `force:false` and **no retry**, the loser gets a non-fast-forward rejection and the deploy **fails silently** (the `git_commit` step errors). This is a **latent bug today** for concurrent multi-site builds, and it gets worse if we add multi-site parallelism (Lever A). Fix: retry-on-non-fast-forward in `updateRef` (re-read HEAD, rebuild tree on the new base, re-commit, re-`updateRef`) — standard optimistic-concurrency retry. This is an adapter change, tracked as a guardrail, not part of C.

---

## OOM / memory

Live `kubectl top nodes`: 5 nodes at 5–38% memory, 1–4% CPU — substantial headroom now. But `reasoning-agent` shows **5/8 pods `Evicted`** — node-pressure eviction has already occurred, so the concern is real under load, not theoretical. Two git-adapter replicas; one dispatch loop running at inspection.

The memory driver is the **handler pods** (LLM generations — 64k-token Opus for tool-recreation), not the small trigger/dispatch-loop pods (256–512Mi). Concurrent handler count is what must fit the node pool. The timeout mismatch (item 4 above) is a **latent over-spawn**: handlers can accumulate past the nominal cap of 8, so OOM/eviction is reachable today and any parallelism increase multiplies it.

Guardrails (do before raising any concurrency; not part of C):
- **Accurate handler pod memory `requests`** (not just `limits`) so the scheduler won't place more handler pods than the nodes hold — excess goes **Pending** (backpressure, slower) instead of **Evicted/OOMKilled** (crash, lost work). Pending is the safe failure mode. Verify `tool-recreation-handler` requests/limits.
- **Align the scheduler timeout** with real dispatch+handler runtime (or make the dispatch loop return fast/async) so the concurrency cap is real.
- **Size the cap to node memory:** `max concurrent handlers × peak handler memory ≤ node-pool memory − (chassis + kafka + postgres + system)`. For Lever B: `Σ over active sites (per-site cap) × peak handler memory`.

---

## The three levers

- **Lever A — multi-site throughput.** Decouple the trigger from the dispatch loop (spawn fire-and-forget; notify scheduler immediately), and/or fix the timeout mismatch, and add a fairness `ORDER BY`. Lets several *different* sites dispatch concurrently (different repos). Does **not** speed up a single multi-tool site. Cadence is already 30s/8, so this is lower priority than first thought; the valuable parts are the fairness ordering and the timeout fix. **Requires the git-adapter retry fix first** (multi-site concurrency exercises the cross-site commit race).
- **Lever B — same-site throughput (the actual single-site speedup).** Relax `NOT EXISTS` from "0 claimed" to "fewer than K claimed" per site (K=2–3). Safe at the claim level (claim is atomic, loop skips already-claimed) and git-safe (same-adoption commits serialize via `correlation_id`). Gated on the OOM guardrails (more concurrent handlers).
- **Lever C — claim-timeout watchdog (SELECTED, build first).** Below.

---

## DECISION: Lever C — claim-timeout watchdog

Cheapest, no extra memory, no new concurrency, helps under A *or* B, and addresses the largest chunk of the observed gap time. Build it first.

### Why a DB-driven watchdog (not the existing mechanism)

Existing recovery is inadequate for this:
- The chassis has `StuckOrchestrationTimeout = 5 * time.Minute`, but timeout handling "uses in-process goroutines [that] die when pods restart, leaving orchestrations stuck … forever" (dev guide: "#1 cause of pipeline stalls"). Restart-fragile.
- Work-item-level recovery appears slow ("recovery via the reaper only kicks in after 30+ [min]"), which matches the observed gaps.
- **Deadlock:** a stale `claimed` item excludes its site via `NOT EXISTS`, so no dispatch loop runs for that site, so `load_work_items` never runs for it, so nothing on the dispatch path reclaims its own stale claim. It clears only via the slow secondary path.

A **scheduled-task SQL sweep** survives pod restarts and runs independently of dispatch, breaking the deadlock and unfreezing the site as soon as the claim is reset (the `NOT EXISTS` exclusion clears immediately).

### Shape (reuse existing patterns)

Model on `work-item-archiver` (a `scheduled_tasks`-driven `query_database` agent calling a SQL function) and `thunder-reaper` (a periodic single-purpose sweep). Concretely:

1. A SQL function, e.g. `reset_stale_claims(threshold_minutes int, max_rows int)`, mirroring `archive_completed_work_items(...)`, that runs:
   ```sql
   UPDATE site_work_items
      SET status = 'triaged',
          claimed_at = NULL,
          claimed_by = NULL,
          attempt_count = attempt_count + 1,
          error = 'Claim reset by watchdog (stale claim)'
    WHERE status = 'claimed'
      AND claimed_at < now() - make_interval(mins => threshold_minutes)
      AND attempt_count < max_attempts;     -- exhausted items fall through to a terminal state, not infinite retry
   ```
   Idempotent (only touches stale claims), respects `max_attempts` (a perpetually-failing item exhausts attempts and stops), and the `attempt_count++` matches the existing retry accounting.
2. A new `scheduled_tasks` row, e.g. `name='claim-watchdog'`, short interval (start ~60s), its own concurrency group, calling a tiny agent that runs the function and updates `last_completed_at` (same pattern as `work-item-archiver`).

### Threshold — the one real tension

Too short resets a legitimately-running handler (it later completes and writes to a re-claimed item → duplicate/conflicting work); too long keeps the slow unfreeze. `call_handler` timeout is 1200s (20 min), so a claim older than ~20–25 min is definitely dead — a safe, conservative v1 that is still far better than the 47–67 min observed and is deterministic. To unfreeze faster without risking live handlers, add a **fast path keyed on orchestration liveness**: reset immediately (ignoring age) when the claim's owning child orchestration is in a terminal/`FAILED` state. That requires linking the work-item claim to its handler orchestration — **confirm whether `claimed_by` (or another column) carries that id** before relying on the fast path; if not, ship the age-based v1 first and add the fast path once the linkage exists.

---

## Open items / to verify

1. **Source of "Claim timed out — handler pod likely died."** Find where this is set (chassis call-timeout path vs an existing reset) and ensure the watchdog and it do not both reset / double-increment `attempt_count`. Set the watchdog threshold below the existing mechanism so the watchdog wins.
2. **`load_work_items` stale-claim behaviour.** Confirm it does not already reclaim stale claims on load (would interact with / partly duplicate the watchdog).
3. **`claimed_by` linkage** to the handler orchestration (for the watchdog fast path).
4. **git-adapter cross-site race (guardrail for A):** add retry-on-non-fast-forward in `updateRef`.
5. **Handler pod resource requests/limits (guardrail for B):** set accurate memory requests so overload → Pending, not Evicted/OOM. Verify `tool-recreation-handler`.
6. **Scheduler timeout alignment (guardrail for A/B):** reconcile `scheduled_tasks.timeout_seconds=300` with `call_handler=1200` / dispatch-loop `4200`.
7. **Fairness `ORDER BY`** in `find_dispatchable_site` (fixes starvation + the LIMIT-1 convergence waste) — pairs with A or B.

---

## Sequencing

**C (watchdog) → guardrails (handler requests + timeout alignment + git-adapter retry) → B (bounded per-site concurrency + fairness `ORDER BY`).** A is optional and lower priority given the 30s/8 cadence; its valuable parts (fairness ordering, timeout fix) overlap the guardrails.
