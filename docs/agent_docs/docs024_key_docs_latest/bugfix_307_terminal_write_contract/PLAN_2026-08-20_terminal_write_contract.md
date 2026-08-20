# PLAN — the work-item terminal-write contract (bugs_open/307)

**Started 2026-08-20.** Fixes `bugs_open/307` per its own §7 spec skeleton and the owner
ruling of 2026-08-18: *"a transient blip should return the item to queued."*

## What we are building, in plain terms

A work item is a row in `site_work_items`. When a handler fails on one, something has to
decide: try again, or give up. Today **four different pieces of code make that decision
differently**, and two of them get it wrong:

- `fail_work_item` counts the attempt and retries — but with **no waiting period**, so
  three attempts can all land inside the same five minutes of a dependency being down.
- `update_work_item_status` with `status: failed` **does not count at all** — it just
  writes `failed`. One failure is the end, in fair weather as much as in an outage.
- Neither of them refuses to overwrite a status a handler **deliberately** chose
  (`needs_human_review`, `wont_fix`), where the two sibling writers on the *success* path
  both do refuse.

We replace the failure half of all of that with **one shared helper** that every failure
path calls, so there is one answer to "what happens when a work item fails" instead of four.

## The design decisions and why

| decision | why |
|---|---|
| A `retry_after timestamptz` column, honoured by the claim + the dispatch reads | The alternatives were all measured and rejected: `blocked` is drained every 600s by `feasibility-recheck` (which also clears `error`), `deferred` holds the dedup slot, and a brand-new status is invisible to the dedup index, the promoter floor and the stale reaper. A NULLable timestamp on a row that stays `triaged` perturbs nothing. |
| Backoff numbers come from `reaper_policies`, not from literals in Go | RFC_018 (SCH-024) exists and explicitly invites `site_work_items` as its second consumer: *"adopt reaper_policies for its numbers first, executor second."* Hand-writing a third set of backoff literals is the exact drift the architecture seat already objected to once. |
| Transient classification is LAYERED, not swapped | `isAIUnavailable` and `RetryDisposition`'s needle lists disagree in **both** directions (EOF/401/credit/api-key on one side; temporary/service-unavailable/bad-gateway on the other). Replacing either loses real coverage. So: burst OR `isAIUnavailable` OR `RetryDisposition`-transient. |
| A transient release never consumes an attempt, but ALWAYS stamps a cooldown | The existing count-free branch retries **forever**, which is tolerable for LLM credit and dangerous for a 404 (a genuinely deleted repo looks identical to a GitHub blip). Cooldown + a release cap of 10 bounds it. |
| The burst detector keys on the normalised `error_message` and on `domain` | `agent_error_log.site_id` is **89.8% NULL** (measured 2026-08-19) — a "≥2 distinct sites" rule would not have fired during the very outage it exists for. And `error_code` is unusable: `ClassifyError` labels the git 404 `LLM_API_ERROR`. |
| Ships ARMED with env kill-switches, not opt-in-default-OFF | Two live precedents pull opposite ways: WII-019 (opt-in, unsafe default OFF) and WDS-018 (armed + env disarm, *"the owner has ruled against default-OFF switches that rot unexercised"*). This change implements an explicit owner ruling rather than granting caller-licensed authority, so WDS-018 is the fit. Both precedents go to the council in the rationale. |
| `status_override` is left exactly as it is | It is `bugs_open/033`'s D2 territory, with its own owner ruling. Changing it from inside this patch would quietly overrule that deferral. |
| Transient release is refused when `max_attempts = 1` | 67 live `needs_diagnosis` items use `max_attempts=1` as a deliberate one-shot lane. A retry mechanism that "helpfully" retries them would break that. |

## Corrections to the originating brief (`bugs_open/307`), recorded not edited away

1. **§2.2's tell is contaminated.** The bug reads `handled_by IS NULL` as "the write came
   from `update_work_item_status`". A fifth writer — the `claimed-item-timeout` scheduled
   task, every 120s, in raw SQL — runs its own correct ladder and *also* leaves
   `handled_by` NULL. 23 rows in 14 days carry its `Claim timed out…` error.
2. **`handled_by` has two more writers** than the bug lists: `plan_sections_action.go`
   and `apply_gap_plan_action.go`, both on `complete`-only paths. The §2.2 inference still
   holds for `failed` rows, but the census predicate should say so.
3. **The bug's headline population is not the biggest half.** The outage cost 100 items
   once. The unladdered `update_work_item_status` path costs **141 items in 14 days**
   (52% of all failures; 71% over the last 48h), 139 of them at `attempt_count=1` with
   `max_attempts=3`, in ordinary weather. Candidate 2 as written would have repaired 88%
   of the incident and none of the daily bleed.
4. **The status vocabulary is widening under us, right now.** `owned_page_refusal_status:
   wont_fix` is live in `page-build-handler.mark_item_failed` config, and 9 `wont_fix`
   rows were written on 2026-08-19. The guard half is therefore not prophylactic any more.

## Phasing

1. Standing five + this plan (done first, deliberately).
2. Go: the shared helper, both failure writers routed through it, the two Go read sites.
3. Migrations: `502` (column + policy seed), `503` (the three SQL read sites).
4. Tests, including the first tests that have ever driven `FailWorkItemAction`.
5. Council submission (migrations are in scope since 2026-08-19), then commit.
6. Apply migrations; verify at the artefact after the next roll.
