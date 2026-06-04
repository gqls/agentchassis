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

Confirmed claim/recovery mechanics (read of `load_work_item_actions.go` and the chassis `ClaimWorkItemAction`):
- **Claim is an atomic conditional update:** `UPDATE … SET status='claimed', claimed_by=$2, claimed_at=NOW() WHERE id=$1 AND status IN ('triaged','approved') AND attempt_count < max_attempts RETURNING id`; `ErrNoRows` → `claimed:false`. Double-dispatch safety is here, not in the per-site `NOT EXISTS`. `claimed_by` is the agent **type** (default `"dispatch-loop"`), **not** a handler orchestration id.
- **`LoadWorkItemsAction` never reclaims stale claims** — it selects only `status IN ('triaged','approved')`. So nothing on the dispatch path resets a `claimed` item.
- **The only fast reset path is the dispatch loop's own error handling:** `call_handler` has a 1200s timeout → `error_step: mark_failed` → `FailWorkItemAction` (`attempt_count+1`, `status = CASE WHEN attempt_count+1 >= max_attempts THEN 'failed' ELSE 'triaged'`). This fires **only while the dispatch-loop pod is alive**.
- **Orphaned claims** (the dispatch-loop pod itself died) are reset by nothing fast: the chassis `StuckOrchestrationTimeout=5min` is an in-process goroutine that "die[s] when pods restart, leaving orchestrations stuck … forever" (dev guide: "#1 cause of pipeline stalls"), and work-item recovery "only kicks in after 30+ [min]." (The exact "Claim timed out — handler pod likely died" string lives in orchestration-engine code not in the captured dump; behaviour above is what matters.)
- **Deadlock:** a stale `claimed` item excludes its site via `NOT EXISTS`, so no dispatch loop runs for that site, so nothing reclaims it — it clears only via the slow secondary path. The observed 47–67 min gaps are this.

A **scheduled-task SQL sweep** survives pod restarts and runs independently of dispatch, breaking the deadlock and unfreezing the site as soon as the claim is reset (the `NOT EXISTS` exclusion clears immediately). It **complements** the dispatch-loop's 1200s `mark_failed` (it catches the orphaned case that path can't), so its threshold sits **above** 1200s, not below.

### Shape (reuse existing patterns)

Model on `work-item-archiver` (a `scheduled_tasks`-driven `query_database` agent calling a SQL function) and `thunder-reaper` (a periodic single-purpose sweep). Concretely:

1. A SQL function, e.g. `reset_stale_claims(threshold_minutes int, max_rows int)`, mirroring `archive_completed_work_items(...)` and matching `FailWorkItemAction`'s default branch exactly:
   ```sql
   UPDATE site_work_items
      SET attempt_count = attempt_count + 1,
          status = CASE WHEN attempt_count + 1 >= max_attempts THEN 'failed' ELSE 'triaged' END,
          claimed_by = NULL,
          claimed_at = NULL,
          error = 'Claim reset by watchdog (stale claim, handler likely died)',
          updated_at = NOW()
    WHERE status = 'claimed'
      AND claimed_at < now() - make_interval(mins => threshold_minutes);
   ```
   Note: **no** `attempt_count < max_attempts` filter — that would strand an already-exhausted stale claim as `claimed` forever. The `CASE` routes an exhausted item to `'failed'` (terminal), exactly as `FailWorkItemAction` does, so a perpetually-failing item stops rather than looping. Idempotent (only touches stale `claimed` rows), and the `attempt_count++` matches the existing retry accounting so the watchdog and `mark_failed` agree.
2. A new `scheduled_tasks` row, e.g. `name='claim-watchdog'`, its own concurrency group, calling a tiny agent that runs the function and updates `last_completed_at` (same pattern as `work-item-archiver`). Interval can be short (~60s) — the threshold, not the interval, governs when a claim is eligible.

### Threshold — the one real tension

Too short resets a legitimately-running handler (it later completes and writes to a re-claimed item → duplicate/conflicting work); too long keeps the slow unfreeze. A handler can run until the `call_handler` timeout of **1200s (20 min)**, so the threshold must sit **above** that — **~25 min** — to guarantee no live handler is reset. That's the safe floor, and it's still far better than the 47–67 min observed, deterministic, and sits above the dispatch-loop's own 1200s `mark_failed` so the two don't fight (the loop resets the in-loop case at 20 min; the watchdog catches only the orphaned case it can't).

A faster unfreeze would need an **orchestration-liveness fast path** (reset immediately when the owning handler orchestration is terminal/`FAILED`, ignoring age). **Confirmed not feasible in v1:** `claimed_by` is the agent *type*, not an orchestration id, and no other column on `site_work_items` links the claim to its handler orchestration. Adding that link (store the spawned handler's `orchestration_id` on the row at claim/spawn time) is a small schema + claim-path change and is the prerequisite for the fast path — a follow-up, not part of v1.

---

## Open items / to verify

1. **Source of "Claim timed out — handler pod likely died" — RESOLVED enough.** Not in the captured chassis dump (orchestration-engine code). The reset paths that matter are confirmed: dispatch-loop `call_handler` 1200s → `mark_failed`/`FailWorkItemAction` (loop alive); orphaned otherwise. Watchdog threshold sits **above** 1200s so it complements rather than races these. (Reading the orchestration-engine/reaper source would let us retire any redundant slow path, but isn't needed to ship the watchdog.)
2. **`load_work_items` stale-claim behaviour — RESOLVED.** It selects only `triaged`/`approved`; it does **not** reclaim stale claims. No overlap with the watchdog.
3. **`claimed_by` linkage — RESOLVED (negative).** `claimed_by` is the agent type, not an orchestration id; no work-item→orchestration link exists. So the fast path needs a schema addition (item 3a) and v1 is age-based only.
   - 3a. *(follow-up)* Store the spawned handler's `orchestration_id` on `site_work_items` at claim/spawn time to enable the orchestration-liveness fast path.
4. **git-adapter cross-site race (guardrail for A):** add retry-on-non-fast-forward in `updateRef`. **(2026-06-04: the missing-homepage lead was NOT this race** — diagnosis showed `n_rendered=0`, i.e. empty assembly, and git-adapter logs showed successful commits with no 422. Open item 4 remains a *latent* risk with no confirmed instance yet.)
5. **Handler pod resource requests/limits (guardrail for B):** set accurate memory requests so overload → Pending, not Evicted/OOM. Verify `tool-recreation-handler`.
6. **Scheduler timeout alignment (guardrail for A/B):** reconcile `scheduled_tasks.timeout_seconds=300` with `call_handler=1200` / dispatch-loop `4200`.
7. **Fairness `ORDER BY`** in `find_dispatchable_site` (fixes starvation + the LIMIT-1 convergence waste) — pairs with A or B.
8. **Schema confirm before writing the function:** verify the `site_work_items` columns the function touches (`status`, `claimed_at`, `claimed_by`, `attempt_count`, `max_attempts`, `error`, `updated_at`) and copy the exact `scheduled_tasks` INSERT shape from an existing task (e.g. `work-item-archiver`).

---

## Sequencing

**C (watchdog) → guardrails (handler requests + timeout alignment + git-adapter retry) → B (bounded per-site concurrency + fairness `ORDER BY`).** A is optional and lower priority given the 30s/8 cadence; its valuable parts (fairness ordering, timeout fix) overlap the guardrails.
