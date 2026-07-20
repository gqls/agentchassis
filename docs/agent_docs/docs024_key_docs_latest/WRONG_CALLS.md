# WRONG CALLS — a fleet-wide ledger of things we asserted that were not true

**Append-only. Newest at the bottom. Any thread, any time.**

---

## What this is, and why it is not 016b §9

`016b` §9 records how **the system** fails — truncation persisted as success,
`<no value>` injections, stdin-eating loops. This file records how **we** fail:
a claim written down at a confidence the evidence did not support.

They are different, and mixing them buries both. A §9 entry teaches you about
the platform. An entry here teaches you about your own reading of it.

## The point is the last column, aggregated

One entry is an anecdote. The value is the **tally of "the cheap check that
would have caught it"** — because a check that keeps appearing is a check worth
automating, and that is a decision you can only make from the distribution.

That tally is what put `check_append_only_docs` into
`scripts/pattern-check.py` (2026-07-20): two documented incidents, measured at
a 2.0% fire rate over 300 commits, wired in as advisory.

**Standing tally** (update it when you add a row):

| the check that was skipped | times |
|---|---|
| read the code before asserting a mechanism | 4 |
| wait / query again before calling an absence a failure | 3 |
| measure a property before describing it | 1 |
| grep the index before filing | 1 |
| read the rule before inferring its purpose | 1 |

**What that distribution says right now:** the dominant failure is not sloppiness
about process — it is **reasoning about a mechanism from its data instead of its
code**. Seven of nine were absences or statistics interpreted as mechanisms. That
class is exactly what the diagnosis loop exists to catch, and the
reasoning-dataset thread used it **zero times** while making nine durable claims.
That is the single biggest structural miss recorded here.

## Row shape

```
### YYYY-MM-DD — <thread> — <the claim, in one line>
**Asserted:** what was written down, and where it was heading.
**Actually:** what was true.
**Caught by:** what surfaced it — and whether that was luck or a check.
**The cheap check that would have caught it:** the one action, before asserting.
**Cost:** what the error actually cost (nothing / a wasted run / a stale premise
in someone else's handoff).
```

Be specific about **Caught by**. "I noticed" is not a mechanism and cannot be
strengthened; "the impossible output tripped an invariant I had written" is.

---

## Entries

### 2026-07-18 — reasoning-dataset — "this `<no value>` defect is live and unfiled"
**Asserted:** drafted a new `/bugs_open/` case file and a plan section calling it
an open bug for another thread to fix.
**Actually:** filed the same day as `bugs_open/016` by the experience-loop thread,
and already fixed in the live row at 13:15:11Z by the council-gate thread — about
45 minutes before I looked.
**Caught by:** grepping `/bugs_open/` before filing, which CLAUDE.md mandates.
**The cheap check:** that grep. One command.
**Cost:** nothing — a duplicate case file and a false alarm to two threads, both
avoided.

### 2026-07-18 — reasoning-dataset — "the 016 fix didn't take"
**Asserted:** two `repropose` calls post-dated the 13:15:11Z fix and still showed
the defect, so the fix had failed.
**Actually:** both belonged to orchestration `48cf0339`, **started 13:11:13Z** —
pre-fix config. The log timestamp is the *step's*, not the *run's*.
**Caught by:** joining to `orchestration_states.created_at` before writing it up.
**The cheap check:** grade a config fix on the RUN start, never the step
timestamp. A long run straddles the boundary and reads either way.
**Cost:** nothing. Now a documented trap in the RUNBOOK.

### 2026-07-18 — reasoning-dataset — "`work_item_id` is being dropped by the auditors"
**Asserted:** four judgement agents log 0% `work_item_id` while another logs
100%, so it is a propagation bug — "plumbing, not physics". Went into a committed
plan as the single highest-ROI change.
**Actually:** those agents are item **producers**. They raise work items and are
never dispatched to work on one, so no id exists at their spawn time. Nothing was
being dropped. One of the four (`feed-triage`) does not touch `site_work_items`
at all.
**Caught by:** reading the code when I sat down to write the change plan.
**The cheap check:** read the code before asserting a mechanism. A 0% column is
evidence of **absence**, not of **dropping**.
**Cost:** a wrong recommendation sat in a committed doc for a day.

### 2026-07-18 — reasoning-dataset — "the human-resolution reason is never captured"
**Asserted:** `approved_by` and `resolution_path` are empty, so the reason is
thrown away; proposed adding writes to the admin handlers.
**Actually:** worse *and* different. The handlers have never been called at all,
so adding writes to them would change nothing. And the reasons **are** captured —
8 of them, written by working threads via direct SQL, which the first query
missed because it was scoped to `status='complete'`.
**Caught by:** re-running the query unscoped, on a hunch, while filing the bug.
**The cheap check:** before concluding "never", drop the WHERE clause and look
again.
**Cost:** nothing shipped; the proposed fix would have been dead work.

### 2026-07-18 — reasoning-dataset — "the guard block is the most distinctive signal in the corpus"
**Asserted:** twice, in a committed plan, as a reason to invest in the design.
**Actually:** 7 trips, all carrying the identical diagnostic. One failure mode,
seven times. An illustration, not a signal.
**Caught by:** measuring it, once the extractor existed — two days after writing
the claim.
**The cheap check:** do not describe a property you have not measured. If you
must, mark it as an expectation.
**Cost:** nothing material, but it was load-bearing in a design argument.

### 2026-07-19 — reasoning-dataset — "the council spawn was dropped" (×3)
**Asserted:** no `orchestration_states` row appeared within ~7 minutes, so the
submission was silently lost. Cited CLAUDE.md's ~300s post-restart rule, checked
the pod, declared it lost anyway — and **re-fired one, spending a council run**.
**Actually:** every run landed. The row is created when the coordinator picks the
message up, which lags the publish by minutes under load — and the platform was
busy.
**Caught by:** eventually waiting longer. Not by a check.
**The cheap check:** poll 10–15 minutes before calling an absence a loss, and
look at whether *other* orchestrations are being created meanwhile.
**Cost:** **real money** — one wasted council run. The only entry here that cost
budget rather than credibility.

### 2026-07-19 — reasoning-dataset — "SQL `LIKE 'review_%'` is inflating the count"
**Asserted:** a plausible mechanism for an 18-row discrepancy — `_` is a
single-char wildcard in LIKE, so it must be matching more than the regex.
**Actually:** zero rows differ. The 18 were council runs — *including my own* —
that arrived after the extract. The corpus is live.
**Caught by:** testing the theory before acting on it
(`WHERE ... LIKE 'review_%' AND ... !~ '^review_'` → 0 rows).
**The cheap check:** a plausible mechanism is not a diagnosis. Test it; it takes
one query.
**Cost:** nothing. This is what the good version looks like.

### 2026-07-19 — reasoning-dataset — assigned a handoff to the wrong owning thread
**Asserted:** picked the owner from the target file's git history.
**Actually:** a different workstream owned that class, and **its own PLAN already
named the exact gap**, down to the same item type.
**Caught by:** listing the workstream directories before writing the handoff.
**The cheap check:** owners live in the workstream docs, not the commit log. `ls`
the directories and grep them first.
**Cost:** nothing — caught in about two minutes.

### 2026-07-20 — reasoning-dataset — overwrote a SUMMARY snapshot
**Asserted:** SUMMARY is documented as "current state only", so rewriting the
morning file in place to reflect the evening's state seemed correct.
**Actually:** CLAUDE.md:177 already said *"Every summary is a NEW FILE — never an
edit of the last one"*, and :191 gave the b-suffix convention for a same-day
second. Each file is current-state-only **and the series is the record** — I
honoured the first property by destroying the second.
**Caught by:** the owner. Not by me, and not by any check.
**The cheap check:** read the rule rather than infer its purpose from one of its
clauses.
**Cost:** a destroyed snapshot, restored verbatim from `4b8d2bca0`. This entry is
why `check_append_only_docs` now exists.

---

## When an entry should become a check

Not every row can be automated, and a speculative check is worse than none — it
fires on ordinary work and teaches people to ignore the whole script
(`scripts/pattern-check.py` says this in its own header, and holds itself to it).

The bar, following the checks already wired in:

1. **A documented incident**, not a guess. At least one real row here.
2. **Mechanically decidable** from a staged diff or a command's output.
3. **Measured** over ≥150 recent commits, firing on **≤2%**. Narrow the predicate
   until it does — "SUMMARY was modified" fired at 4.3% and would have caught
   legitimate appends; "SUMMARY lost ≥20 lines" fires at 2.0% and catches only
   rewrites.
4. **Advisory, never blocking.** It prints; a human decides.

Rows 1, 3, 4, 5 and 8 above are **not** mechanically checkable — they need
judgement at assertion time, which is what the diagnosis loop is for.
