# 341 — `claimed-item-timeout` runs a FIFTH copy of the work-item retry ladder, in SQL, outside the contract that now governs the other four

**Filed 2026-08-20** by the `bugfix_307_terminal_write_contract` lane, **at the council's
insistence and not on my own initiative** — which is the point of the file. Three independent
seats of council corr `4cdec68b-fa17-436d-8e25-8c422ee6c8c5` objected that naming this in prose
as "a residual" is not the same as tracking it:

> `reuse_agent` (medium): *"'Named as a residual' in rationale is not the same as tracked
> remediation."*
> `editquality` (missing): *"a fourth failure-ladder writer cited by the diagnosis with no
> covering edit; disclosed but not converged, so 'one contract' is only true for the two Go paths."*
> `architecture` (low): *"the four-writers-four-guarantees defect class this plan targets is only
> 4/5 converged, and the fifth is left as a residual with no ticket referenced."*
> `bug_historian` (low): *"the exact shape of 016b §9's 'one call site of a shared judgement gets
> the rigorous fix; the sibling stays heuristic' — the platform's own history says the untreated
> sibling is where the next incident originates, not a footnote."*

They are right, and the last one is the argument: this estate's own record says the untreated
sibling is where the next incident comes from.

## 1. What it is

`scheduled_tasks.claimed-item-timeout` — **enabled, every 120 s** — reclaims items whose claim
went stale. Its `reset` CTE (live pre_query, lineage
`docs/agent_docs/sql_for_agents/220_claimed_item_timeout_generic_evidence.sql`, last amended by
`482`) runs its **own** retry ladder in raw SQL:

```sql
status = CASE WHEN attempt_count + 1 >= max_attempts THEN 'failed' ELSE 'triaged' END,
attempt_count = attempt_count + 1
-- WHERE status = 'claimed' AND claimed_at < NOW() - INTERVAL '...'
```

It is a faithful copy of the ladder `bugs_open/307` has just replaced in Go — and therefore it
now **differs from the contract in every way 307 added**:

| | the Go contract (WII-024) | `claimed-item-timeout` |
|---|---|---|
| counts the attempt | yes | yes |
| **cooldown between attempts** (`retry_after`) | **yes** | **no — re-triages immediately** |
| **terminal-DECISION guard** | **yes** | **no — `WHERE status='claimed'` only** |
| **transient / burst release** | **yes** | **no** |
| clears `claimed_by`/`claimed_at` | yes | yes |
| writes `handled_by` | on the Go path only | **no — leaves it NULL** |

## 2. Why this is not merely untidy

1. **It re-creates 307's own defect (a) on its own path.** An item that times out during an
   infrastructure outage is re-triaged with **no wait**, re-claimed on the next dispatch tick,
   times out again against the same dead dependency, and reaches `failed` in three laps — which
   is exactly the mechanism that killed 88 of 100 items on 2026-08-17, arriving by a different
   door.
2. **It is not a rare path.** [MEASURED 2026-08-19] 23 `failed` rows in 14 days carry its
   `Claim timed out…` error; 18 are `tool-auditor` at `attempt_count=3`.
3. **It silently contaminates the census everyone uses.** `handled_by IS NULL` is the documented
   tell for "written by `update_work_item_status`" (307 §2.2 attributes its 12 attempt-1 deaths
   with it). This task also leaves `handled_by` NULL, so the tell over-counts. Any query using it
   needs `AND error NOT LIKE 'Claim timed out%'` — and nobody knows that until they read this.

## 3. Why 307 did not just fix it

Not laziness, and worth stating because the fix is less obvious than it looks:

- **A `scheduled_tasks.pre_query` is SQL executing in the database.** It cannot call
  `applyWorkItemFailureLadder`, which is Go. Converging them means either re-expressing the whole
  contract (guard list, `reaper_policies` lookup, burst detection over `agent_error_log`) as SQL —
  a **second implementation of one rule**, i.e. the drift class again — or moving the sweep out of
  `scheduled_tasks` and into a Go action, which is a change to how the sweep is scheduled and
  owned, not a bug fix.
- The pre_query is also **lockstep-pinned**: `claim_timeout_exclusion_lockstep_test.go` globs the
  lexicographically newest `*_claimed_item_timeout_generic_evidence.sql` and asserts a
  biconditional over its `item_type NOT IN (…)` list. Any edit ships as a **new** numbered file
  (`482` is current) and must not spell the clause in a comment above the declaration.

## 4. Fix candidates, ordered by what closes the door

1. **Move the sweep into Go and call the shared contract** (closes the door): a scheduled action
   that selects the timed-out claims and calls `applyWorkItemFailureLadder` per row. One
   implementation, one guarantee; the sweep's *selection* stays declarative. Cost: the task stops
   being a pure `pre_query`, and its exclusion list has to move with it — including the lockstep
   test's anchor.
2. **Teach the SQL ladder the two cheap halves** (partial, no Go): add the terminal-decision guard
   and a `retry_after` stamp to the `reset` CTE, reading `reaper_policies` in the same statement.
   Leaves burst detection Go-only, so an outage still costs these items their attempts — but it
   stops the immediate re-claim and stops decision statuses being overwritten. Ships as a new
   `NNN_claimed_item_timeout_generic_evidence.sql`.
3. **Make the divergence visible rather than fixing it** (weakest, and NOT sufficient alone): a
   lockstep test asserting the SQL ladder's CASE matches the Go one's shape. Records drift; does
   not stop it. Acceptable only alongside 1 or 2.

## 5. How to verify

Whichever ships: an item that times out its claim during a simulated burst must **not** lose an
attempt, must carry a `retry_after`, and must not overwrite a `needs_human_review`/`wont_fix` row.
The disconfirming case is the same one 307's suite uses: a lone item timing out with no burst must
still reach `failed` at `max_attempts`.

## 6. Relations

`bugs_open/307` §8 (the contract this diverges from) · register **WII-024** (the Go contract) and
**WII-002** (this sweep) · `SCH-015`/`SCH-022` (its two-phase reset, and the documented
"DO NOT BUILD THIS — the reset already exists" catch) · council corr
`4cdec68b-fa17-436d-8e25-8c422ee6c8c5` (four seats, quoted above) · `016b` §9 "one call site of a
shared judgement gets the rigorous fix; the sibling stays heuristic".
