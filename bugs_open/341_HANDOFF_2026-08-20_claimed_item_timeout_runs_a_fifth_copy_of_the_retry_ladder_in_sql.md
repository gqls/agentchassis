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

---

## §2b — CORRECTION 2026-08-21: this file overstates its own case, and I wrote it

**The "no terminal-DECISION guard" row in §1's table is WRONG.** I recorded the sweep's predicate
as *"`WHERE status='claimed'` only"* and scored that as an absent guard. It is the opposite:
`status = 'claimed'` is **strictly stronger** than the decision guard for this purpose. A row a
handler has deliberately parked — `needs_human_review`, `wont_fix`, `rejected` — is by definition
no longer `claimed`, so this UPDATE cannot reach it, and the race closes because the WHERE is
evaluated at execution time rather than against a list read earlier.

The Go path needs an explicit `status NOT IN (…)` because it is handed a `work_item_id` and writes
whatever it finds. This sweep **selects its own population** and has already excluded every
decision status by construction. Different shape, so the same audit line does not apply to both —
which is precisely the error: I compared the two writers on a checklist instead of on what each
one can actually reach.

**Consequence for the fix:** candidate 2 narrows to the **cooldown stamp alone**. Adding a guard
would be seven statuses of dead SQL implying a risk that does not exist here, which is its own
kind of misinformation.

**What remains genuinely divergent, unchanged:** no cooldown (fixed by the file below), no
transient release (cannot be fixed in SQL — burst detection reads `agent_error_log` and layers two
Go classifiers; re-expressing that here is the second-implementation drift `307` exists to end),
and `handled_by` left NULL (deliberate — writing it would change what `handled_by IS NULL` means
mid-stream for queries already written against it; `error NOT LIKE 'Claim timed out%'` stays the
discriminator).

## §3b — candidate 2 is BUILT and HELD: `524_claimed_item_timeout_honours_the_cooldown_HOLD.sql`

Written, parse-checked against the live schema in a rolled-back transaction, committed, and
**deliberately not applied**.

**How it edits:** a verbatim-anchored `replace()` on the `pre_query`, not a rewrite. It asserts its
anchor occurs exactly once *before* writing, touches only that fragment, and never retypes the
other ~6 kB — so a concurrent edit elsewhere in that column survives instead of being silently
reverted. This answers by construction the objection the council's `debug_historian` seat raised
against `506` (corr `4cdec68b`, high): a full-value overwrite with no pre-state check, in a table
with drift history.

**It carries a real parse check, and it exists because the first version was broken.** The initial
edit inserted the new assignment without a comma after the error CASE's `END` — invalid SQL — and
**every string-matching assertion in the file passed on it**. That is the landmine exactly: a
`pre_query` is DATA to a migration's probe and parses only when the task next RUNS, 120 s later, in
a job nobody watches. A sweep that fails to parse raises nothing a human sees; it simply stops
reclaiming timed-out items, which looks identical to "nothing is timing out". The file now runs
`EXECUTE 'EXPLAIN ' || v_pq` — which parses and plans without executing, even for data-modifying
CTEs — and aborts with the SQLSTATE if it fails.

**⚠ WHY IT IS HELD, and it is an ordering fault rather than a defect in the file.** `bugs_open/344`:
the dispatch loop's `mark_complete` overwrites a ladder-re-triaged item to `complete` ~2 s after the
failure write, because `triaged` is not in the completion guard. Harmless before `307`; load-bearing
now that the ladder makes `triaged` the post-failure state. Fingerprint: `retry_after > completed_at`.
Stamping `retry_after` from this sweep **creates exactly the row shape 344 destroys**, so applying it
first would widen a Go-ladder defect into one that also reaches every claim timeout.

**Release condition:** `344` resolved, or its candidate 1 (completion refuses while `retry_after` is
in the future) live. Then re-run the file's pre-state gate — it aborts if another lane has edited the
`pre_query` since 2026-08-21 — rename off `_HOLD`, apply, record.

**The `_HOLD` suffix is load-bearing, not a label:** `run-migrations.sh`'s `SIDECAR_RE` excludes it
from `--apply` while still listing it. A banner in a comment cannot stop another lane's sweep.

**Not yet council-submitted** — migrations are in scope since 2026-08-19, and this one goes up
when it is released rather than while it is held, so the round reviews what will actually run.

## §4b — a fact worth knowing before anyone reorders this work

Until `bugs_open/344`'s sibling defect is fixed — the Go ladder's terminal transition erroring with
SQLSTATE **42P18**, so `failed` at `max_attempts` is unreachable through it — **this sweep's SQL
ladder is the only writer in the estate that can actually reach terminal `failed`.** Which is an
argument for touching it carefully, not quickly.

## §5b — the sweep's auto-COMPLETE arm is 344's defect one status over, and a fix on the Go side alone will not reach it

Contributed by the `307/317/313/301` session, 2026-08-21, and it sharpens §2b rather than softening it.

§2b is right that this sweep cannot overwrite a **decision** status: `WHERE status='claimed'` excludes
every parked row by construction. **But that argument covers the RESET arm only.** The sweep also has
two auto-**COMPLETE** CTEs (`completed_by_orchestration`, `completed_by_evidence`), and those stand in
exactly the same relationship to a mid-cooldown row as `mark_complete` does in `bugs_open/344`:
they write `complete` over a row the ladder has deliberately put back in the queue.

**Why this matters for how 344 is fixed, not just for this bug.** If 344's candidate 1 lands as a
predicate on `CompleteWorkItemAction` — "refuse while `retry_after` is in the future" — then this
sweep's **direct SQL UPDATE bypasses it entirely**, the same way it already bypasses both completion
gates (`bugs_open/317`'s mechanism, one status over). A Go-side-only fix would close the door in the
room and leave the window open, and the fingerprint (`retry_after > completed_at`) would keep
appearing with no Go writer to blame — which is the kind of residue that gets diagnosed as a fresh
bug six weeks later.

**So the sweep needs the same predicate**, in SQL, whenever 344 is built. That is a second surgical
edit to this same `pre_query`, and it should ship in the SAME file as this bug's cooldown stamp
(`524_..._HOLD.sql`) rather than as a third visit — the release condition for that HOLD is already
"344 resolved or its candidate 1 live", so the two land together by construction.

⚠ **Note the asymmetry, because it is not obvious:** the RESET arm needs no guard (§2b) and the
COMPLETE arms need one. Same `pre_query`, opposite conclusions, and the reason is which population
each arm selects — `status='claimed'` for the reset, evidence-of-completion for the others. **A
checklist applied to "the sweep" as a unit gets both wrong**; that is the same error §2b was written
to correct, met a second time in the same file.

## §5c — CORRECTION 2026-08-21: §5b is WRONG, and it is the same error §2b exists to correct

§5b says the sweep's two auto-COMPLETE arms need the retry-cooldown predicate because they stand in
the same relationship to a mid-cooldown row that `mark_complete` does. **They do not.** All three
arms — the reset and both auto-completes — carry `WHERE wi.status = 'claimed'`, and a
ladder-re-triaged row is `triaged` with its claim columns **cleared by the ladder itself**. Nor can a
row be claimed *and* mid-cooldown: the claim path refuses to re-claim before the stamp expires.

[MEASURED 2026-08-21] **0** rows at `status='claimed'` carry any `retry_after`; **0** fingerprint rows
(`retry_after > completed_at`) are attributable to this sweep (`error LIKE 'Auto-completed%'`).

So **no SQL half is needed for `bugs_open/344`**, and adding the predicate here would be dead SQL —
precisely what §2b concluded about the reset arm and refused to write.

**This is the third time in two days that a checklist applied to "the sweep" as a unit gave the wrong
answer** (§2b was the first, §5b the second, this is the correction). The question that works every
time is *which population does this arm SELECT* — and for all three arms it is `status='claimed'`,
which is why all three are safe. I recorded §5b from a peer contribution without applying my own §2b
test to it, which is the more useful lesson: **a correction you wrote does not automatically apply
itself to the next claim you accept.**

## §6b — the HOLD's release condition, SHARPENED

`524_..._HOLD.sql` was held for "`344` resolved, or its candidate 1 live". `344`'s Go fix is now
**committed** (`0f80f5ea1`) — and that is **not sufficient**, because the owner has deliberately
deferred the roll. Committed code guards nothing.

**Release condition, corrected: `344`'s Go fix must be LIVE — i.e. rolled and verified at the
artefact.** Until then, stamping `retry_after` from this sweep creates mid-cooldown rows while
`mark_complete` is still unguarded in production, which is exactly the widening the hold exists to
prevent. The distinction matters because "the fix is committed" is the reading that would have
released it a day early.

## §7 — 2026-08-22: the proof event may have become RARE because 307's fix removed its supply

`524` has been live for **~25 h as of 2026-08-22 18:00Z** and the sweep has reset **0** items in
that time — so candidate 2 is applied and structurally verified but still has **no artefact proof**.
Before reading that zero as anything, the disconfirming check, run today so nobody repeats it:

- **the sweep is healthy** — `enabled`, firing on its 120 s interval (last fire 136 s before the
  check), `last_completed_at` equal to `last_triggered_at`;
- **its `pre_query` still parses** — `EXECUTE 'EXPLAIN ' || pre_query` returns `PARSES OK` against
  the live column. This is the check this bug's own migration header demands, because a sweep that
  fails to parse raises nothing a human sees and looks *exactly* like "nothing is timing out".

So the zero is a real absence of claim timeouts, not a broken sweep.

### ~~Why the absence is probably CAUSAL, not quiet~~ — **REFUTED the same day; see §7b**

[MEASURED 2026-08-22, boundary = v1.0.1322 rolled 2026-08-21 16:53Z]

| | before the 42P18 fix (7 d) | since it rolled (~25 h) |
|---|---|---|
| claim timeouts / day | **4.71** | **0** |
| claims / day | **859.71** | **480.31** |

⚠ The "since" figure for timeouts is **0 counted separately** — the rate query returned *no row*,
because an aggregate over an empty filtered set produces no row rather than a zero. Do not read a
missing row as a measured zero; that is why it is stated twice here.

Claim volume roughly halved, which explains part of it. Volume-adjusted, the expectation over the
window is ≈ **2.6** timeouts against **0** observed — about a **7 %** outcome by timing alone, so
suggestive rather than demonstrated.

**The hypothesis, and it is mechanically plausible rather than merely convenient:** `bugs_closed/307`
§9b's 42P18 defect made the ladder's TERMINAL write error. A failure write that errors fails the
saga — and an item whose saga dies mid-flight stays `claimed`, which is precisely the population
this sweep exists to reap. **So the defect we fixed was itself manufacturing claim timeouts**, and
fixing it removed part of this bug's own supply of proof events.

### What that means for anyone waiting on this bug

- **Its historical rate (39 resets in the 14 d to 2026-08-21, ~2.8/day, of which only ~half took the
  stamping arm) OVERSTATES the wait's end.** The honest expectation now is "rarer than that, by an
  unmeasured amount".
- **A continued zero is NOT evidence of breakage** — the two checks above are what would show that,
  and both are clean. Re-run them rather than re-deriving the reasoning.
- The event still required is unchanged: one item whose claim goes stale past the 40-minute
  threshold **with attempts remaining**, so the sweep takes its non-terminal arm and stamps a
  cooldown. The terminal arm writes `retry_after = NULL` by design and proves nothing.
- If a week passes with no qualifying event, the useful move is not to wait longer but to ask
  whether this path is now rare enough that candidate 1 (moving the sweep into Go, where it would
  share the contract rather than mirror its numbers) is the cheaper way to retire the divergence
  than proving the mirror works.


## §7b — CORRECTION 2026-08-22: §7's "probably causal" is NOT SUPPORTED, and the flaw was my model, not my query

The `bugs_open/358` lane refuted one of their own headline claims today and sent me the
transferable form: **a discriminator has to be run at a finer resolution than the effect it claims
to discriminate.** Theirs was computed on daily buckets. So was mine. Applied to §7, it refutes it.

§7 said 0 observed against ~2.6 expected was *"about a 7% outcome by timing alone, so suggestive"*.
**That figure assumes claim timeouts arrive at an even rate. They do not — they are bursty.**
[MEASURED 2026-08-22, the 7 days before the fix] 33 timeouts spread over 168 hours, but only **21
hours contained any** (12.5%), with up to **8 in a single hour**.

Once events cluster, the comparison is not a rate but the **gap distribution**
[MEASURED, 14 days before the fix]:

| | |
|---|---|
| median gap between timeouts | **0.8 h** |
| **longest natural quiet gap** | **2 days 3 h 24 m (≈51 h)** |
| natural gaps ≥ 25 h | **1** |
| the current silence | **24.2 h** |

**A 51-hour silence occurred naturally, before anything was fixed.** The current 24.2 h is shorter
than that, and shorter than a gap that had already happened once in a fortnight. So the silence is
**well within normal variation and is evidence of nothing** — neither of the fix removing supply,
nor of breakage.

**What survives from §7 unchanged:** the sweep-health check (enabled, firing on its 120 s interval,
`pre_query` parses) — that was sound and it is what rules out the alarming reading. And the
*mechanism* by which 42P18 could manufacture claim timeouts (an erroring terminal write fails the
saga; the item stays `claimed`) remains plausible. **What does not survive is the claim that the
data supports it.**

**What WOULD be evidence**, so nobody re-derives this: a silence exceeding the **51 h** natural
maximum, or a rate drop measured over a window containing several bursts — not a quiet day. That
means the absence says nothing before **2026-08-23 ~20:00Z** at the earliest.

**Recorded rather than edited away** because the error is instructive: I wrote *"suggestive rather
than demonstrated"* and thought the hedge made the figure safe. It did not — **a cautious sentence
around an unsound statistic is still an unsound claim**, and the hedge does double damage: it makes
the number feel handled, and it reads as conservative rather than unfounded. `WRONG_CALLS.md`,
2026-08-22.
