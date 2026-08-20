# NOTES — the work-item terminal-write contract (bugs_open/307)

Append-only, newest at the bottom. Missteps are the point, not an appendix.

## 2026-08-20 (session start) — research, and what it changed about the plan

Picked up `bugs_open/307` on the owner's instruction. `scripts/who-owns.py 307` says
OWNED/active (`staged_component_build`), but that lane's own §7 ends *"whoever builds it
should treat this section as the spec's skeleton"* and its handoff lists 307 as open work —
so this is contribution into a shared account, not a competing fix. Re-checked at start of
implementation: **0** council runs in the last 12h mention the seam.

**Bug re-validated before building** (it is two days old and this tree moves):
- Code unchanged: `FailWorkItemAction`'s three branches all still `WHERE id = $1`;
  `UpdateWorkItemStatusAction` still increments `attempt_count` on both arms with no ladder
  and no guard. `git log` on both files shows no fix landed.
- Still bleeding: `failed` rows at `attempt_count=1` with `handled_by` NULL continue daily
  (72 on 08-18, 2 on 08-19).

### The 090 diagnosis run — filed, and it FAILED, which is worth recording precisely

Fired `090` (intake corr `87f3b06f-91d6-4220-9533-796bce5cb196`, run corr
`103ad179-6674-4f0f-9d7d-79a5372cfdc9`). It assembled **four** evidence bundles
(63k/37k/92k/62k chars) whose contents agree with our reading, and then its **`verdict`
step died**:

```
step verdict failed: … response truncated: stop_reason=max_tokens
(output_tokens=32000 reached the configured cap, 0 chars recovered)
```

The intake item went terminally `failed` at `attempt_count=1` — correctly, because the
diagnose lane runs `max_attempts=1` by design. **So there is no gradable verdict.**
[MEASURED — the item row and the four `diagnosis_artifacts` rows]

Two things follow, and I am recording both rather than the convenient one:
1. **My symptom statement packed THREE mechanisms into one filing** (no-delay ladder,
   unladdered writer, LLM-only classifier) when the trigger's own guidance is *"one
   coherent bug per run"*. That is the most likely contributor to a 32k-token verdict.
   The next filing on this seam should carry one mechanism.
2. The bug was filed under the 2026-07-31 ruling's **named** escape hatch (first-hand
   verification, declared in the bug's own header). We proceed on that, plus this session's
   independent re-verification in code, in the live DB, and across three sweeps — not on a
   verdict we do not have. Owner asked; owner answered: proceed, record the failed run.

### What the research actually changed (i.e. what I would have got wrong)

- **I would have used `blocked` as the cooldown state.** The bug file suggests it. It is
  unusable: `feasibility-recheck` (enabled, every 600s) releases *every* `blocked` row whose
  handler exists — no timestamp condition — **and clears `error`**, destroying the reason.
  That is why the table holds **0** `blocked` rows and has held 0 for its whole history:
  not unused, continuously drained. [MEASURED]
- **I would have written backoff literals into Go.** `reaper_policies` already exists
  (migration 335, register SCH-024) and RFC_018 explicitly names `site_work_items` as the
  second consumer it is waiting for. Literals would have been the third hand-rolled copy of
  a mechanism the architecture seat already objected to once.
- **I would have keyed the burst detector on `site_id`.** In `agent_error_log`, `site_id` is
  **89.8% NULL** over 7 days (and 2025 of 2084 rows NULL *inside the outage window itself*),
  while `domain` is **0%** NULL. A "≥2 distinct sites" rule would have failed to fire during
  the exact outage it was written for. [MEASURED]
- **I would have swapped `isAIUnavailable` for the shared `RetryDisposition`.** Their needle
  lists disagree in both directions; swapping silently drops EOF/401/402/credit/api-key
  coverage. Layering keeps both.
- **I assumed the outage was the main damage.** It is not. The unladdered path costs
  **141 items in 14 days** (52% of all failures, 71% in the last 48h), 139 at
  `attempt_count=1` of 3 — in fair weather. The incident was the alarm, not the fire.

### One prediction I am writing down BEFORE it can be checked

The guard half is currently **unprovable from data**: both writers write `failed`, so no
query can separate "the handler said failed and it stood" from "the loop overwrote it".
9 `wont_fix` rows exist as of 2026-08-19. **Prediction: once the guard ships, the count of
`wont_fix`/`needs_human_review` rows that later flip to `failed` should be 0, and before it
ships that count is not measurable at all.** If someone later reports the guard "did
nothing", that is the expected reading of a prophylactic — check the skip log line, not the
row count.

## 2026-08-20 (implementation) — appended as it happens

### Build, and the four things that changed shape while building

**1. `build-dispatch-watchdog` does not exist.** The plan named three SQL read
sites, on the strength of `docs/agent_docs/sql_for_agents/214_build_dispatch_watchdog.sql`
and a research sweep that quoted its pre_query verbatim. It is not live:

```
$ git status --short docs/agent_docs/sql_for_agents/214_build_dispatch_watchdog.sql
?? docs/agent_docs/sql_for_agents/214_build_dispatch_watchdog.sql     -- UNTRACKED
$ git log --oneline -1 -- .../214_build_dispatch_watchdog.sql          -- (no commits)
SELECT name FROM scheduled_tasks WHERE name ILIKE '%watchdog%';        -- 0 rows
```
Some session wrote it, never committed it, never applied it. **A file in
`sql_for_agents/` is not a live task — the live inventory is the table**, and a
research sweep that reads the repo will keep reporting this one as real. So 506
patches TWO read sites, not three, and the "false BUILD_DISPATCH_STALLED" risk
the plan listed does not exist. Recorded in 506's own header too, where the next
reader of that file will be standing. [MEASURED 2026-08-20]

**2. The live inventory of tasks that UPDATE `site_work_items.status` is five**,
and it matches what the research said: `claimed-item-timeout` (120s, and it runs
its OWN copy of the ladder), `detected-item-promoter` (900s), `feasibility-recheck`
(600s), `held-pair-canary-escalation` (86400s), `stale-work-item-reaper` (3600s).
None reads `retry_after`. The claimed-item-timeout duplicate ladder is left alone
in this change and named in the submission as the remaining divergence.

**3. `handled_by` nearly got broken by my own fix.** The helper takes an
`agentType` and writes it to `handled_by`; `update_work_item_status` has no agent
identity to pass, so it would have written `''`. That column being **NULL** is the
documented tell that separates the two writers — bug 307 §2.2 attributes its 12
attempt-1 deaths with it, and other lanes' censuses use it. Writing `''` would have
populated the column without identifying anyone and silently invalidated all of
them mid-stream. Caught before commit; the SQL now reads
`handled_by = COALESCE(NULLIF($3, ''), handled_by)`, which also keeps every bind
parameter referenced (the Go-side branch I tried first left `$3` unused and would
have failed at runtime with "bind message supplies 5 parameters").

**4. Two of my own test fixtures were wrong, and the code was right.** Both
failures were the burst probe not firing: the detector refuses to query on an
empty signature, because an empty one would collapse every blank-message failure
in the fleet into one group that matches itself. My fixtures assumed a probe on
cases with no readable `__step_error.message`. Fixed in the tests, and the reason
is now stated per case rather than inferred.

**Also:** a Go raw-string literal cannot contain a backtick, and I put
`` `handled_by IS NULL` `` inside a SQL comment inside one. It compiled as a
syntax error 20 lines later. And `nullableJSON` already existed in this package
returning `[]byte` (which lib/pq sends as bytea and will not cast to jsonb), so
the text form is a separate `jsonbTextOrNil` rather than a change to a helper
this work has no business touching.

### Mutation proof — the five behaviours that must not silently rot

Each mutation applied to a clean copy at HEAD, tests run, code restored:

| mutation | caught by |
|---|---|
| drop `wont_fix` from the guard list | `GuardListIsNotTheCompletionPathList`, `TheGuardIsActuallyInTheStatement` |
| never build the guard clause | `TheGuardIsActuallyInTheStatement` |
| never stamp `retry_after` | `RetryAfterIsActuallyStamped` |
| drop the agent-type leg of the burst conjunction | `BurstThresholdsRequireTheConjunction/volume_alone,_one_agent_type` |
| let `max_attempts=1` items be released | `AOneShotLaneIsNeverReleased` |

The first two matter most: a mock decides how many rows come back, so a test that
only asserts "skipped" passes against a statement with no guard at all. The SQL
text has to be read.

### Migration validation

Both files, then both rollbacks, executed against the live schema inside one
transaction and rolled back: `ALTER TABLE / COMMENT / INSERT 0 1 / DO / UPDATE 1 /
UPDATE 1 / DO / … / ROLLBACK`. Each of 506's two UPDATEs matches exactly one row,
and both DO-blocks pass — including 506's assertion that the selector still
carries every pre-existing dispatchability clause, so a hand-typed near-copy
cannot silently drop the `depends_on` or claim guards.
