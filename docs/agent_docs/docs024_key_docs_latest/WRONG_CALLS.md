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
| read the code before asserting a mechanism | 6 |
| wait / query again before calling an absence a failure | 3 |
| **grep for the capability before asserting it does not exist** | **1** |
| **prove the artefact is current before reasoning from it** | **1** |
| measure a property before describing it | 1 |
| **look at the real values before designing for the assumed ones** | **1** |
| grep the index before filing | 1 |
| read the rule before inferring its purpose | 1 |

**What that distribution says right now:** the dominant failure is not sloppiness
about process — it is **reasoning about a mechanism from its data instead of its
code**. Seven of nine were absences or statistics interpreted as mechanisms. That
class is exactly what the diagnosis loop exists to catch, and the
reasoning-dataset thread used it **zero times** while making nine durable claims.
That is the single biggest structural miss recorded here.

**Update 2026-07-20 (bugs_open/002 D).** The two new rows sharpen that reading
rather than change it. Both are the same shape as the dominant class — asserting
instead of running one command — but they name the two *directions* it goes:
claiming a thing does not exist (`asserted-absence`), and claiming a thing does
exist because a stale artefact showed it. Note the second is the first row here
whose check is not about *code* at all: it is about proving your **evidence** is
current, which the "trust the rendered artefact" rule silently assumed. Both came
from one thread on one bug entry within an hour, and **neither was caught by any
process — both were caught by the owner asking for a double-check.** That is the
uncomfortable line in this file: the tally measures checks we skipped, but the
*catch* column is still mostly luck and human review.

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

### 2026-07-20 — work_item_completion_integrity — "page_rerender is the obvious first verifier: 40% of completions and it already carries page_id"
**Asserted:** that `page_rerender` was the highest-leverage and *easiest* item type
to write a completion verifier for, and I put it to the owner as the recommended
scope on that basis. The reasoning was volume plus identity: 1,849 of ~4,644
completions, `page_id` on 1,914 of 1,929 rows, per-target items, predicate already
in `check_misdirected_cta`. I also used it to argue the council-reviewed
`submission_B` had picked badly by targeting two 184-item types instead.
**Actually:** the volume and identity figures were right, and the conclusion was
still wrong. `check_misdirected_cta`'s own file header records that the
page-rerender handler only rewrites CTA url fields of components in the actions
package's `ctaFieldNames` set, and that a misdirect anywhere else — prose, say —
is *deliberately* left for the next discovery pass to re-detect and escalate to
human review via the two-strike rule. So re-running the whole-page predicate at
completion is **stricter than the handler's remit**: a correctly-handled rerender
with an out-of-remit misdirect would verify as unresolved, burn its attempts and
strand in `failed`, destroying the designed escalation across 1,849 items. The
verifier I had written and tested was a regression wearing a fix's clothing.
**Caught by:** writing the verifier's own scope-guard doc-comment. Explaining *why*
an unrecognised `spec.reason` must refuse forced me to state what the handler is
actually responsible for, which sent me to the check's header. Half luck — the
comment was being written for a different reason (bugs_open/032's lesson about
never guessing on input you cannot check) and the remit note happened to be in the
same file. Nothing would have caught it at commit time; the tests I wrote all
passed, because they tested the predicate I had chosen rather than the one the
handler implements.
**The cheap check that would have caught it:** before calling a defect
"verifiable", read the HANDLER's remit, not just the detector's predicate. A
verifier asserts the handler did its job — so the question is what the handler is
responsible for, and that was written down, in English, at the top of the same
file I took the predicate from.
**Cost:** nothing shipped — the verifier was held before commit on the owner's
call, and the finding is recorded in the coverage guard's own gap entry for
`page_rerender` so the next thread starts from the remit problem. It did cost most
of a session's build, and it briefly made a council-reviewed plan (`submission_B`)
look worse than mine when its two "harder, lower-volume" targets may simply have
been the ones whose handlers had matching remits.

---

### 2026-07-20 — bugfix thread (bugs_open/002 D) — "no targeted single-section repair path exists, so D is an architectural gap"

**Asserted:** in a sign-off of `bugs_open/002`, that repairing one empty section
on a content-rich page had no mechanism and needed one designed — repeating the
2026-07-15 handoff's own fix-candidate wording ("fix needs … a TARGETED
single-section repair") and upgrading it to "an architectural gap (guard vs
repair)". It went into the handoff, the status table, a commit message and a
spoken summary to the owner.

**Actually:** `apply_section_edit` / the `section-editor` agent has existed since
**2026-02-19** — five months before the handoff that declared it missing. It
edits one `page_components` row's `content_data`, re-renders that component and
reassembles the page, and it does **not** pass through `save_page_sections`, so
the content-regression guard was never in its path. Seeded, active, present in
the deployed pod binary, trigger script committed, **3 COMPLETED production
runs** (latest 2026-07-10). The real gap was far narrower: nothing *generates*
the field content, and nothing *routes* `empty_section` items to it.

**Caught by:** the owner asking "please double check your claims". **Not a
check** — nothing in the process would have surfaced it. It had already survived
the original handoff, a deliberate re-grounding pass, and a sign-off, because a
"fix candidates" section reads as design guidance and inherits none of the
evidence standard applied to the symptom report it sits under.

**The cheap check that would have caught it:** one grep of the action registry —
`grep -nE '"[a-z_]*(section|component|repair)[a-z_]*":' registry.go` puts
`apply_section_edit` on screen. Generally: "X does not exist" is exactly as
checkable as "X is broken", and cheaper. The four-lookup existence check is
registry → `agent_definitions` (active? workflow?) → pod binary (`strings`) →
`orchestration_states` (ever ran?). Two minutes.

**Cost:** no wasted build — the work was never started. But it put a false
"architectural gap" into a handoff, a commit message and an owner-facing summary,
and it mis-sized the remaining work as an investigation when it is an afternoon's
wiring. Filed as a §9 pattern (`asserted-absence` / `dormant-machinery`) and
proposed to the council as a new seat, because two threads made it independently.

### 2026-07-20 — bugfix thread (bugs_open/002 D) — "the empty section is live on the site, visibly broken"

**Asserted:** that `gripper-cycle-time-estimator.html` was serving a broken
`tool-guide-intro` section — a "How-To Guide" eyebrow badge, an empty
`<h1 class="tgi-headline"></h1>`, and "Read time:"/"Level:" labels with blank
values — as an orphan with no `page_components` row. I had reconstructed a whole
mechanism around it (removed from the DB on 2026-07-10 but surviving in the
deployed artifact) and was one step from creating a `page_rerender` work item to
force a production re-assemble and re-deploy of a live commercial page.

**Actually:** the section was already gone — from the DB *and* from the live
site. The page had been rebuilt correctly. My evidence was a **stale
CDN/edge-cache response**: the origin file had been updated at 12:46 that day,
`cache-control: max-age=3600` was still serving the pre-update copy, and my
first `curl` caught it. Stale copy 52,624 bytes with the orphan; real page
44,880 bytes without it. Everything downstream of that first fetch was
reasoning about a document no visitor was being served.

**Caught by:** re-fetching with a cache-buster **before** creating the work
item — but only because the DB kept contradicting the page. Four separate
checks (no `page_components` row, not in any of the three section sources, the
component_id dangling in `content_components`, and two *successful* assemble-only
re-deploys of the page HTML on 07-17 and 07-18) all said the section did not
exist, and I kept inventing mechanisms to explain why the page disagreed instead
of doubting the page. **Half luck** — I busted the cache to identify the
*source* of the orphan, not because I suspected the fetch. Had the DB been
ambiguous, or had I checked the DB after the page instead of before, I would
have shipped a needless production deploy.

**The cheap check that would have caught it:** when the live artefact
contradicts the database, **suspect the fetch before inventing a mechanism** —
`curl -H 'Cache-Control: no-cache' "<url>?cb=$(date +%s)"` and compare
`content-length`, or read `last-modified` against `pages.deployed_at` first. Here
`last-modified` was *two hours before my fetch and two days after the recorded
deploy*, which was the tell, sitting in a header I had already printed and
skimmed past. Corollary to the standing rule "trust the rendered artefact, not
the status": **only if you have proven the artefact is the current one.** An
unbusted GET is not the rendered artefact, it is a cache's opinion of it.

**Cost:** nothing shipped, and no credits — the item was never created. It cost
maybe twenty minutes of investigation and produced a genuinely wrong intermediate
claim ("a visible defect on the live site") that I stated to the owner in the
same confident voice as the verified findings around it, with no `[INFERRED]`
marker. The counterfactual cost was a pointless production deploy of a live
client page justified by a screenshot of a cache. Also worth noting: this is the
**second** correction in one investigation of the same entry — the first was
asserting a repair mechanism did not exist (`016b` §9 `asserted-absence`) — and
both came from asserting rather than checking a thing that was one command away.

---

### 2026-07-20 — bugfix-036 — "the reviewer is emitting `"3"`, so a numeric-string-tolerant type fixes it"

**Asserted:** carried forward from `bugs_open/036` §5, which named `json.Number` as
a candidate on the reasoning that a reviewer answering "which edit?" would emit
`"3"` where an `int` was wanted. I started implementing against that shape.
**Actually:** none of the three live payloads was a numeric string. All three were
plan-level *descriptions* — `"plan-level (deploy verification)"`, `"risks note on
the 54 mis-stamped rows"`, `"risks/summary (item 5)"`. `json.Number` would have
parsed **zero** of them, and the fix built on that premise would have shipped,
passed its tests, and left the bug fully live. The deeper miss: the reviewers were
not malformed at all — they were saying "this objection is plan-wide", which the
contract already spells `0`. The strict type was discarding a *meaning*, not noise.
**Caught by:** a check, not luck — querying the actual `collected_data` payloads
out of the three voided orchestrations before writing the tolerant type, because
the fix's own test cases had to come from somewhere real.
**The cheap check that would have caught it:** one `jsonb_path_query` for the
offending field's real values across every occurrence, before choosing the
tolerant type. Roughly 30 seconds.
**Cost:** nothing shipped — caught during implementation. It would have cost a
full build/roll cycle and a bug that reads as fixed, which is the expensive kind.
The bug file's own suggestion was the trap, which is the transferable part: a
handoff's fix *candidates* are hypotheses, and inherit no evidence from the
diagnosis above them.

### 2026-07-20 — bugfix-036 — "a mistyped field deep in a JSON struct loses the whole object"

**Asserted:** a test expecting that an objection carrying `"severity": 3` (int
where string was wanted) would be dropped entirely by the per-field salvage.
**Actually:** `encoding/json` continues past a **type** error (unlike a syntax
error) and keeps everything that did decode, so the objection survives with only
the offending field zeroed. The salvage retains materially more than assumed —
and, notably, the *original* failing code had the fully-populated struct in hand
and discarded it because `err != nil`.
**Caught by:** the test failing — `objections: want 0 got 1`. A check, and the
cheapest possible place to be wrong.
**The cheap check that would have caught it:** asserting the behaviour of the
standard library rather than assuming it — which is what the test did, one step
later than ideal.
**Cost:** nothing. Recorded because the assumption is load-bearing for the fix's
design and would have made a reader think the salvage was weaker than it is.

### 2026-07-20 — bugfix-036 — "both my cluster dispatches were dropped"

**Asserted:** forming, not yet written down — after ~28 minutes, neither the
diagnosis run nor the council submission had an `orchestration_states` row and
neither correlation appeared anywhere in the chassis logs, while the chassis was
visibly consuming other work (34 orchestrations in 10 minutes). The pull was
toward resubmitting both.
**Actually:** both were queued. `kafka-consumer-groups.sh --describe --group
generic-requests-group` showed **LAG 85** on `system.agent.generic.requests`
behind a **single** consumer — the known `bugs_open/030` backlog, aggravated by
three orchestrations hung for 1–3.4 hours (`bugs_open/029`) holding the group.
**Caught by:** a check prompted by a prior recorded incident — the memory of the
2026-07-18 council-latency trap (which cost three redundant runs) said "no rows
means queued, not dropped; verify before resubmitting", so I measured the lag
instead of resubmitting.
**The cheap check that would have caught it:** the consumer-group lag query. It is
the only reading that distinguishes "dropped" from "queued", and neither the
absence of DB rows nor the absence of log lines can.
**Cost:** nothing — this is the same wrong call as 2026-07-19's "the council spawn
was dropped (×3)" above, and this time the recorded row did its job. Logged as a
*repeat prevented*, which is the tally this file exists to produce.

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
