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
| read the code before asserting a mechanism | 7 |
| **read the CONTRACT a thing plugs into, not just its logic** | **2** |
| **name the LAYERS a claim spans, and touch each one** | **3** |
| wait / query again before calling an absence a failure | 4 |
| **grep for the capability before asserting it does not exist** | **3** |
| **prove the artefact is current before reasoning from it** | **3** |
| measure a property before describing it | 1 |
| **record the CLOCK beside a reading, never infer it afterwards** | **1** |
| **run a census against a known-positive control before reporting the count** | **1** |
| **look at the real values before designing for the assumed ones** | **4** |
| **read the SCHEMA before naming a column — a Go map key is not a column** | **1** |
| **re-derive an inherited residual's prescription; a previous session's fix note is a hypothesis, not a spec** | **1** |
| grep the index before filing | 1 |
| **check whether an existing bug has an owning workstream before routing work to it** | **1** |
| **read before write — never `cat >` a file you did not create** | **1** |
| **re-resolve a file:line you carried across sessions — above all one you edited yourself** | **1** |
| **verify an embedded/quoted artifact is COMPLETE before asserting it — a fixed `[:N]` slice is an unmarked truncation** | **1** |
| **re-read the row AFTER a render, not after your own write** | **1** |
| **check the column actually means what you are measuring** | **1** |
| read the rule before inferring its purpose | 1 |
| **resolve BOTH operands to the same ground before comparing — same run, same namespace** | **4** |
| **confirm the record you are reading is the one that produced the artefact** | **2** |

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

**Update 2026-07-20 (chassis v1.0.1140).** A tenth row lands, and it is the most
expensive so far because it escaped the most checks: a council submission whose
headline proposal was *unimplementable*, through two council rounds and into
another thread's queue. Its new tally row — *read the CONTRACT a thing plugs
into, not just its logic* — is a sibling of the dominant class, but note what is
different about it. The dominant class is asserting a mechanism without reading
the code; this is reading the code and stopping at the part that was interesting.
I quoted `ItemVerifier`'s file twice in that same submission for other purposes
and never once looked at its signature. **Neither the council nor any check
caught it** — the implementing thread hit the wall. That is now twice in this
file where a REVIEW passed something only implementation could refute, which is
the honest limit of plan-stage review and an argument for handing a plan over
sooner rather than polishing it through another round.

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
`page_rerender` so the next thread starts from the remit problem.
**FOLLOW-UP 2026-07-20, and it is its own small wrong call:** the entry above rests on
the DETECTOR's header comment. I never opened the handler — so my correction to a
comment-reading error was itself sourced from a comment, and I propagated it into four
documents in that state. Now read: `rerender_page_sections_action.go:283-296` does gate
the recompute on `ctaFieldNames[fn]`, so the claim HOLDS; and `applyCTARecompute` has two
further early-returns (keeps an already-valid authored link; declines when no valid
target), making the remit NARROWER than I claimed. Right conclusion, unearned method,
twice in one session. The cheap check is unchanged and I skipped it twice: **read the
handler.** It did cost most
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

### 2026-07-20 — bugfix thread (bugs_open/044) — "110 seeded, active agents have never run"

**Asserted:** stated to the owner, in the same voice as the session's verified
findings, that a one-line probe showed **110** active agents with no row in
`orchestration_states` — offered as evidence that "dormant machinery is common
rather than exceptional". I did hedge that spawn-only children might inflate it,
which turned out to be the *wrong* caveat and gave the number false credibility:
it looked like a figure someone had already thought about.

**Actually:** the query was invalid. It counted via `owner_agent_type`, and
**95,797 of ~101k orchestrations carry `owner_agent_type='generic'`** because
that is the dispatch path, not the agent. Counting that way reports
`fix-proposer` and `council-gate` as never-run — both demonstrably run (councils
appear in 92 `workflow_plan`s; the new `review_prior_art` seat in 20). The
spawn-only caveat was also wrong in the opposite direction: spawned children
**do** appear as owners (`page-content-writer`, 638 rows). Re-measured by
step-fingerprint the real figure is **57 of 122 measurable agents**, with a
34-agent blind spot — and the 57 is mostly *retired* agents still flagged
`is_active`, not dormant capability. The headline claim survived; the number
that made it sound authoritative was about double, and its composition was
nothing like what I implied.

**Caught by:** recognising `fix-proposer` in the output list. **Not a check —
luck.** Had the list happened to contain only agent names I did not know, I
would have filed the 110. Worse, the invalidating fact was already written in my
own memory (*"council_report source_agent='generic' fleet-wide — partition by
another key"*), so I had been told and did not connect it.

**The cheap check that would have caught it:** before reporting a census, run it
against a **known-positive control** — pick two things you are certain are true
(an agent you have watched run) and confirm the query says so. A count with no
control is an untested query, and an untested query stated as a finding is the
same error as an unchecked absence, wearing a number instead of a claim.

**Cost:** nothing shipped — I flagged it as needing proper measurement before
acting, which is the only reason it did not become a bug's headline figure. It
did put a wrong number in front of the owner, and it is the **fourth**
assertion-shaped error in one session (repair path, live defect, council seat,
now this). Three of the four were caught by the owner asking; this one by luck.
That distribution is the finding.

### 2026-07-20 — bugfix-036 — "the dispatch queue drains at 0.21 msg/min, so my probe waits ~6.5 hours"

**Asserted:** a throughput figure written into `bugs_open/030` as **measured**,
with a timestamped table, and repeated to the owner as the reason a live
verification could not be completed. I offered to abandon the probe on the
strength of it.
**Actually:** ~2.4 msg/min — the figure was **~12× too slow** and the queue
estimate wrong by the same factor. My "two readings 14 minutes apart" were **69
seconds apart**. I never recorded when the sampling job started; I *inferred*
19:16 from when I thought I had launched it, and wrote the inference into a table
as a measurement. The probe I said would take 6.5 hours was picked up at 19:30:32
and had written its report by **19:34:12** — roughly 27 minutes end to end, and it
landed four minutes after I told the owner it was hopeless.
**Caught by:** the bugfix-030 thread, which had a continuous 30-second sampler
running over the same window and matched my two rows to its own samples at
19:29:48 and 19:30:57, to the digit. **Not by me** — and it should have been:
**my own data contained the disproof.** My third sample read `current=96016` and
the "14 minutes later" reading also read `current=96016`. Identical offsets mean
no time passed. I had the contradiction in front of me and did not look at it.
**The cheap check that would have caught it:** print the clock in the same command
that prints the reading (`date -u; kafka-consumer-groups.sh …`) so a rate is
computed from two recorded times, never two remembered ones. Then: this queue
drains as a **sawtooth** (pins for minutes on one long message, then bursts), so
two point samples cannot measure it at all — sample continuously for ≥20 min and
take the slope.
**Cost:** a wrong "measured" figure sat in an open bug file for about an hour and
was quoted to the owner as grounds to give up on a verification that in fact
succeeded. Nothing shipped, and the probe was already queued so no work was lost —
but this is the one entry here where the error went *to the owner as a decision
input*, which is worse than a wrong line in a doc.

### 2026-07-20 — work_item_completion_integrity — "the diagnosis dispatch produced no orchestration; something didn't dispatch"
**Asserted:** that two `090_TRIGGER_needs_diagnosis` runs had failed to dispatch. I had
built what felt like a careful case: zero `orchestration_states` rows queried **by
correlation id** (the documented method), zero artifacts after 30 minutes, zero
occurrences of the correlation in the chassis log, and other council runs visibly
completing at 19:35, 19:58 and 20:08 — so the pipe was demonstrably moving. I explicitly
told the owner "the evidence is now much stronger than my earlier mistake", and I had
already re-fired the trigger once and was drafting a new bug file.
**Actually:** both runs were **queued**. `kafka-consumer-groups.sh --describe --group
generic-requests-group` → **LAG 181**, PartitionCount 1, one consumer. Everything I
listed as evidence of a drop is listed in `bugs_open/030` — filed the day before — as the
signature of the backlog, under the heading *"Read this first if you are about to
conclude 'my dispatch was dropped'. It probably was not."* Its measured latency is
**25–36 minutes**; I called it at 31 and re-fired at ~30.
**Caught by:** `grep -ril` over `/bugs_open/` and `/bugs_closed/` for the mechanism —
the de-duplication check CLAUDE.md requires *before filing*. It fired one step before I
wrote the duplicate. Not luck, but later than it should have been: the same grep was
available before I concluded anything, and I ran it only when about to file.
**The cheap check that would have caught it:** grep the bug directories for the
mechanism BEFORE forming the conclusion, not before filing the write-up. And for this
specific class, run the one command `030` puts in its own summary line — consumer-group
lag — which answers "queued or dropped" in seconds and which I had not run in either of
my two attempts at this question today.
**Cost:** one redundant diagnosis dispatch (now also queued), a near-miss duplicate bug,
and a confidently wrong status report to the owner. Second time today I read a queue as a
failure; the first cost three council runs. The tally row below is the point — this class
now has a filed bug, a §9 entry, a memory note and two ledger rows, which is the argument
for automating the lag check rather than writing it down again.

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

### 2026-07-20 — robot-hands — "the OnRobot 2FG7 is rated 11 kg, so it clears an 8 kg part"
**Asserted:** written into a test as an expected PASS while validating the new MatchMatrix tool —
the assertion being that a gripper whose published payload is 11 kg obviously satisfies an 8 kg
requirement, so a tool that failed it must be miscalculating.
**Actually:** the tool was right and the test was wrong. An 8 kg part on dry steel (μ 0.15, S=2)
needs **523.2 N** of clamping force; the 2FG7 publishes **140 N** maximum. Its 11 kg headline
implies μ ≈ 0.77 — rubber or form-fit fingers, not a bare machined surface. The two figures are not
in conflict, they answer different questions, and the headline one is the wrong question.
**Caught by:** the test itself, on first run — I had written the expectation and the implementation
from different premises, so they disagreed immediately. Luck, in the sense that I only wrote a test
at all because the tool was going to production; had I hand-checked the output I would have "fixed"
the correct code to match my wrong expectation.
**The cheap check that would have caught it:** do the arithmetic before asserting the expectation.
One line of Python. The published payload rating is a *derived* figure resting on the
manufacturer's own friction assumption — the primary figure is the force.
**Cost:** nothing, and it turned into the tool's most useful feature: MatchMatrix now explains this
exact discrepancy inline whenever a gripper passes on rated payload but fails on computed force,
naming the implied μ. The trap that caught me is the one the tool's users were most likely to hit.

### 2026-07-20 — bugfix 003 — "the new prober bridges into platform/health's existing Checkers machinery"
**Asserted:** in a council-gate submission (rounds 2 and 3, trail `3a18a1a4`), as the
answer to a reuse objection. I moved a chassis-local Kafka prober into
`platform/health/` and gave it a `Checker() CheckFunc` method, then described this as
reuse of "the existing `Checkers` machinery" in `health.Server` — implying the new type
plugs into something the fleet already runs.
**Actually:** `health.NewServer` has **zero callers** anywhere in the tree
(`grep -rn "health.NewServer" --include='*.go' .` → nothing). The `Checkers`/`CheckFunc`
types are real and well-shaped, but nothing constructs the server that consumes them, so
my "bridge" connects to machinery that is itself dead. The *file placement* was still the
right call — a shared package is where a shared prober belongs, and the other seven
binaries with hardcoded health endpoints are now plausible adopters — but the reuse
claim as written overstated what exists today.
**Caught by:** the `prior_art_librarian` seat, round 3: it noticed that every OTHER
absence claim in the submission carried an attached grep, and this one — the only
*presence* claim — did not. That asymmetry was the tell, and it was correct.
**The cheap check that would have caught it:** the same grep I ran for every other claim
in the same submission. `grep -rn "health.NewServer"` — one command, and I ran its
equivalent three times that hour for the claims I doubted, but not for the one I liked.
**Cost:** none functionally (the method is three lines and harmless), but it is the exact
dormant-machinery pattern this repo keeps rediscovering: a plausible-sounding "we already
have this" that nobody re-greps. Recorded because the failure mode is *asymmetric
scepticism* — I checked the claims that would have made my plan look worse, and not the
one that made it look better.

### 2026-07-20 — reasoning-dataset — proposed a verifier that could not be written
**Asserted:** council submission B named `hardcoded_section_colors` as the first
item type to gain a completion verifier — *"Predicate is purely deterministic"*,
*"the check already knows how to decide"* — and shipped a sketch calling
`hardcodedColorVerdict(html)` after a `component_id` lookup. Two council rounds
argued about it. Nobody, me included, checked whether the verifier signature
could reach the data that predicate needs.
**Actually:** it could not. `ItemVerifier` took `(ctx, db, spec, logger)`, and
`hardcoded_section_colors` files **one item per site** — its predicate needs
`site_id`, which measured over all 5,514 live items appears in just **9** specs
(against 2,370 with `page_id`). The verifier was **unwritable**, however willing
the author. My submission's central proposal was unimplementable as specced.
**Caught by:** the `work_item_completion_integrity` thread, when it went to
implement the handoff and hit the wall. Not by me, not by two council rounds, and
not by any check.
**The cheap check that would have caught it:** read the *contract* the thing
plugs into, not just the logic it would run. I described what the predicate
computes without once opening `ItemVerifier`'s signature — the same file I had
quoted twice for other purposes.
**Cost:** real. It sent a handoff into another thread's queue with its headline
proposal broken, and cost that thread the detour of widening the contract
(`VerifyTarget{ItemID,SiteID,PageID,ItemType,Spec}`, commit `08b35ccc4`) before
it could do the work at all. Their correction also improves on my diagnosis: I
blamed the coverage gap on the mechanism being opt-in — *"stays at one unless an
author remembers"* — which aims the fix at discipline. The real reason was that
for a whole class of item types a verifier was impossible to write.

### 2026-07-20 — robot-hands — "bugs_open/023 is unowned, so it is this site's next action"
**Asserted:** wrote `023` into a fresh handoff as "now the highest-value fix here,
and it is a code fix", with implementation direction, as though nobody was on it.
**Actually:** the `cta_link_integrity` workstream owns it, was **six council rounds
in**, and its observe-only stage had **already shipped live in v1.0.1140**
(`f6b4aea5a`) — in the same image roll I was told about. Its PLAN already carried
the defect classes, including the `ctaFieldNames` coverage gap I was proposing to
go and find.
**Caught by:** the owner mentioning a fresh chassis build. Checking what shipped in
it surfaced the CTA commit, which led to the workstream. **Luck** — nothing about
my own process would have found it; I had grepped `/bugs_open/` before filing `043`
but never checked whether an *existing* bug I was routing work to had an owner.
**The cheap check that would have caught it:** `ls
docs/agent_docs/docs024_key_docs_latest/` for a workstream matching the bug, before
promoting that bug to a next action. CLAUDE.md's "grep before you file" covers new
bug files; **routing work to an existing bug needs the same check and I did not
give it one.**
**Cost:** nothing, caught before the handoff was acted on. Had it not been, the next
chat would have started a competing fix against a live staged rollout, on a shared
branch, in the exact area a council trail was mid-flight.

### 2026-07-20 — robot-hands — overwrote another thread's memory file with `cat >`
**Asserted:** implicitly, that `memory/cta-link-integrity-workstream.md` was mine to
rewrite because my index line for it was stale.
**Actually:** the file belonged to the owning thread and held their state. `cat >`
destroyed it. The memory directory is not under git, so it was unrecoverable — I
reconstructed it from the surviving `MEMORY.md` index line and their repo docs, and
marked the file as a reconstruction so they can restore what I lost.
**Caught by:** an `assert` in the *next* step failing — my `MEMORY.md` edit asserted
on the old index text, which had already been updated by them with much better
detail than my replacement. The guard that saved the index is the one I had not put
on the file itself.
**The cheap check that would have caught it:** read before write. The Write tool
enforces this and refused me later in the same minute; `cat >` in Bash does not, and
that is the whole difference. **Prefer Read-then-Write over `cat >` for any file you
did not create** — CLAUDE.md already says a stale-looking file is not permission to
replace it, and this is the mechanical version of that rule.
**Cost:** one destroyed memory file, partially reconstructed; the owning thread may
have lost detail it will have to rewrite.

### 2026-07-20 — travelling docs / bugs_open/024 — "12 of 122 components" and "no tool has an input_schema"
**Asserted:** in a council submission's `risks` block, that the tool-section
exemption matches **12 of 122** active components and that the marker is safe
because **no tool has an `input_schema`**. Both were stated flatly, as
established facts, in the section a reviewer reads to judge blast radius.
**Actually:** live at the time of writing it was **13 of 123**, and **14 of the
27 active `component_level='tool'` components DO have an `input_schema`** — a
majority-adjacent slice of the exact population I was making a safety claim
about. The figures were carried forward unchecked from the round-4 submission
written the previous day.
**Caught by:** the council. `guardian` and `bug_historian` both refused to take
the population on prose and demanded it be settled by query; one query settled
both. Not caught by me at any point across two submissions.
**The cheap check that would have caught it:** one `GROUP BY component_level`
against `content_components` — the same query I eventually ran, which took
seconds. CLAUDE.md's "ground every figure against the live system before
repeating it from another doc" is exactly this rule, and the figure was
**one day old**, which is precisely the age at which it still feels current.
**Cost:** none to the code — the predicate tests the schema, so it was right
while my justification for it was wrong. That is the uncomfortable part: a
correct implementation defended with a false premise reads as verified, and had
the population actually been what I claimed, nothing in the submission would
have caught it.

### 2026-07-20 — travelling docs / bugs_open/044 — a stale file:line and "silent"
**Asserted:** that the sibling deferral heuristic lives at
`plan_sections_action.go:1090-1108` and that it carries a section **silently**.
Written into a council submission's `risks` block, then into the first draft of
the bug file.
**Actually:** it is at **1141-1160**, and it is **not silent at the decision** —
it logs `plan_sections: content component has empty schema, deferring` at
**Warn** with the function and section, and sets an explicit `item.Reason`. What
is invisible is the downstream *consequence* (a deferred section is carried, so
a template fix is discarded), which is a narrower and more useful claim than the
one I made. The line numbers were inherited from the round-4 submission and then
shifted **by my own edit to that same file** earlier in the session.
**Caught by:** opening the function to check the citation before committing the
bug file — i.e. by the filing discipline, one step before it became durable.
**The cheap check that would have caught it:** read the code at the line you are
citing, especially when you have edited that file yourself in the same session.
A file:line carried across submissions is a moving target the moment anyone
touches the file, and I was the one who touched it.
**Cost:** none — corrected in place in `bugs_open/044` before first commit, with
the correction kept visible. But it went into a council submission first, so
reviewers reasoned about a citation that did not resolve.

---

## 2026-07-20 — appended a case update to `bugs_open/008` after it had moved to `bugs_closed/`

**The claim:** implicit but load-bearing — "008 still lives in `/bugs_open/`"
(and, in the update I wrote into it, "closure awaits behavioural proof"). Both
false when written: another thread had closed the case ~30 minutes earlier,
with its own pod verification, and moved the file. My `cat >>` recreated the
path as an untracked 1.6KB orphan — a fork of a closed case, the exact
"second account that drifts" CLAUDE.md warns about — and my follow-up
`git commit <path>` failed on it, which is what surfaced the mistake.
**Caught by:** the commit's pathspec error ("did not match any file(s) known to
git") + `git ls-files | grep 008` showing the tracked copy in `bugs_closed/`.
**The cheap check that would have caught it:** `ls bugs_open/NNN* bugs_closed/NNN*`
(or `git ls-files | grep NNN`) BEFORE writing to a case file — CLAUDE.md already
says "grep BOTH directories before filing"; the same check applies before
*updating*, because on a busy day a case can close between your premise and
your write. Same freshness class as "your session-start git status is a
snapshot": case-file locations are live state, not context.
**Cost:** near-zero — orphan deleted before any commit; the closing thread's
record stands untouched. But the update I wrote argued a closure position the
owning thread had already decided differently with better evidence, so if the
commit had succeeded it would have shipped a contradictory fork of a closed
case.

### 2026-07-20 — bugfix 003 — "the health fix is verified live: the endpoint returns real JSON"
**Asserted:** to the owner, and written into `bugs_open/003`, after v1.0.1140 shipped: the new
`/health` was "verified against running pods, not git" because the chassis pod and a spawned
Job pod both returned `{"kafka_last_ok_seconds_ago":N,"status":"ok"}`, the discriminating
literal greped 1 in the pod binary and the old `READY` literal greped 0.
**Actually:** every one of those checks was true, and the endpoint was still **broken**. All I
had proven was that the new code was *running* — not that it *reported correctly*. A
deliberately wedged test pod (Kafka pointed at an unroutable address; `wget` timed out, every
real client logged `i/o timeout`, topic creation failed) reported `{"status":"ok"}`
continuously for six minutes and never restarted. Two causes: (1) the prober read
`cfg.Infrastructure.KafkaBrokers` — viper, i.e. the config *file* — while every real client
resolves brokers from the *environment* via `kafka.GetBrokers()`, so it probed a different
Kafka than the process was using; (2) a bare `net.DialTimeout` to `10.255.255.1:9092`
**succeeded**, so TCP connect is not evidence of a reachable broker in this pod network.
**Caught by:** the restart test the owner asked for — and only because it was run against a
genuinely wedged pod rather than a healthy one. A "does the endpoint respond" test would have
passed forever.
**The cheap check that would have caught it:** ask what the endpoint says when the answer
should be NO. I verified the healthy path four ways and never once exercised the unhealthy
path — the only path the change exists to serve.
**Cost:** a fleet-wide health check that could not detect the failure it was built for, live
in production for ~4 hours, described to the owner as verified. Fixed in `976618dbb`
(GetBrokers + a Kafka metadata round-trip via `Conn.Brokers()`), inert until the next roll.
**The transferable rule:** *"verified live" must name which BRANCH was exercised.* A green
positive path plus a discriminating grep proves deployment, not correctness — for anything
whose job is to detect a fault, the fault must be induced.

### 2026-07-20 — robot-hands — "the CTA label/URL pairing is repaired"
**Asserted:** to the owner in a summary, and in a commit message (`d497fbe24`,
"CTA pairing repaired"), after applying label-keyed UPDATEs to `content_data` and
running a verify query that came back clean.
**Actually:** `content_data` is not the source of truth for a CTA URL.
`resolve_internal_links_action.go` **owns** those fields (`ctaFieldNames`, `:99-105`)
and recomputes them on every render via `chooseCTATargets` (`:319`), which never reads
the label — it sorts candidates by `NavOrder`/`Name` and takes `[0]` and `[1]`. The
next render put back both the URL and its `_target_title`, with a later `updated_at`
than my write. The edits that still look correct do so only because the resolver would
choose that URL anyway; *building the tool* is what fixed those, not the SQL.
**Caught by:** watching `services.html` after it re-rendered, for an unrelated reason
(I was checking whether the statistics had landed). **Luck.** Nothing in my process
would have surfaced it — my verify ran inside the writing transaction, where it could
not, in principle, detect a field the render recomputes.
**The cheap check that would have caught it:** re-read the row **after a render**, not
after the write. CLAUDE.md's "trust the rendered artefact, not the status" already says
this about pages; it applies to the DB row just as much, and I applied it to one and
not the other.
**Cost:** a false "done" reached the owner and a commit message. Corrected in NOTES, in
the handoff, and as an addendum to `/bugs_open/023` — where it turned out to sharpen
that bug's stated cause, so the error paid for itself. But it was reported as finished
work first, which is the part that matters.

### 2026-07-20 — robot-hands — "zero work items completed fleet-wide since the roll"
**Asserted:** to the owner, as the headline evidence for a fleet-wide build halt, from
`SELECT count(*) ... WHERE status='complete' AND updated_at > '<roll>'` → 0.
**Actually:** `site_work_items.updated_at` is **not maintained** (`/bugs_open/035`), so
that query returns 0 whether or not anything completed. The halt was real — re-measured
on `completed_at`, there were 0 completions between the roll and the recovery and 1
immediately after — so **the claim was true and the evidence for it was worthless.**
That is the more dangerous shape: I would have said the same thing had the fleet been
perfectly healthy.
**Caught by:** my own recovered work item completing and still not appearing in the
query. The contradiction was only visible because I re-ran the check after the fix.
**The cheap check that would have caught it:** `\d site_work_items` and a moment's
thought about which column means what — or simply leading with the evidence that does
not depend on a timestamp column at all: **N hung `build-pipeline-trigger` rows in
`AWAITING_RESPONSES` against `max_concurrent`**, which is the mechanism itself and was
sitting right there.
**Cost:** nothing, because the conclusion held. Recorded because a right answer from a
broken instrument is not a verification, and the next person reading that query in my
runbook would have inherited the instrument. Fixed in `/bugs_open/029`'s addendum and
in the robot-hands RUNBOOK.

### 2026-07-20 — bugfix 003 — "the test pod inherited KAFKA_TOPIC from the shared ConfigMap"
**Asserted:** in the bugfix_003 runbook, the bug file, and to the owner — explaining why a
throwaway test pod ended up consuming `system.agent.generic.*` despite a custom `AGENT_TYPE`.
**Actually:** `personae-prod-config` contains **no topic keys whatsoever** (`KAFKA_TOPIC` is
set in the agent-chassis *Deployment's* env block, which the test pod never copied), and
`main.go`'s `cfg.Custom["topic"]` is not what opens the consumers in the first place. The real
mechanism is a hardcoded fallback in `setupConsumers()` (`agent.go:332`, `:362`): unset
`REQUESTS_TOPIC`/`RESPONSES_TOPIC` → listen on the generic topics, comment and all
(*"Only the main orchestrator listens on the generic topic"*). Spawned dynamic agents never
reach it — the spawner injects their `job.*` topics.
**Caught by:** the owner, flatly — *"we are not using system.agent.generic.\*, the dynamic pods
create their topics dynamically."* It did not match my story, which is what made me go and read
`setupConsumers()` and find the real mechanism.

> **FOLLOW-UP 2026-07-20:** the owner then retracted the premise, and the docs settle it —
> `001_development_guide(5).md:476-486` documents BOTH: `job.<stable-identity>.*` per-spawn
> topics for the dynamic fleet (*"Always use this when you can"*), AND
> `system.agent.generic.requests` as *"the generic entry point"* for trigger scripts and the
> scheduler, live and consuming. Both statements were true of different halves of the system.
> **The retraction does not retire this entry**: the challenge was right that my explanation
> was wrong, and my ConfigMap claim was independently false whoever raised it. A correction
> prompted by a premise that later turns out to be mistaken is still a correction — the lesson
> (read the thing before asserting its contents) is untouched.
**The cheap check that would have caught it:** `kubectl get cm personae-prod-config -o json`
and one grep for the topic key — I asserted the ConfigMap's contents without ever reading it,
in the same breath as reporting a possible production-traffic incident. Reading the function
that opens the consumer would have done it too.
**Cost:** a wrong mechanism written into a runbook whose whole purpose is to stop the next
person repeating the mistake — the most expensive place to be wrong. Corrected in place.
**Pattern:** this is the third entry today from the same root — asserting a plausible cause
without opening the thing that would confirm it. The other two were `health.NewServer` having
no callers, and "verified live" that only exercised the healthy branch.

### 2026-07-20 — reasoning-dataset — "the human-review queue has no working surface"
**Asserted:** filed `bugs_open/033` under that title, having established that the
three Go admin routes had never run, that a fourth handler was written and never
registered, and that two columns were dead. Concluded there was nowhere for a
human to action a review item, and put the prior question to the owner as
*"queue or bin?"*.
**Actually:** a **complete** review surface exists and always did — in the admin
dashboard frontend. `frontends/admin-dashboard/src/App.tsx` has the work-item
list, a `needs_human_review` filter with a count badge, and wired
Approve / Retry / Resolve / Reject / Skip buttons calling the very endpoints I
had declared unused. The real defect is narrower and better: the list loads the
**newest 50** non-complete items, so it shows **0 of the 208** build-pipeline
review items and reports the queue as empty. Not absent — **blind**.
**Caught by:** a bugfix thread re-grounding the file the same day, and reading
the frontend.
**The cheap check that would have caught it:** `grep -rn "needs_human_review"
frontends/` — one command, in the layer where a *human* surface would obviously
live. I searched `platform/` and `internal/` exhaustively and never once left Go.
**Cost:** a bug filed under a false title, and a "queue or bin?" question put to
the owner that did not need asking — most of the fix needs no ruling at all.

> **This is the third time in one session the same shape has failed**, and the
> repetition is the point. `work_item_id` "is being dropped" (checked the log
> table, not the dispatch path). The resolution reason "is never captured"
> (checked one status, not the table). The review surface "does not exist"
> (checked the backend, not the frontend). Each time I searched **one layer
> exhaustively** and reported a property of **the whole system**. Thoroughness
> inside a boundary reads exactly like thoroughness overall, which is what makes
> it convincing and wrong. The check is not "look harder" — it is **name the
> layers the claim spans, and touch each one**, before writing "never", "no", or
> "does not exist".

### 2026-07-20 — bugs_open/033 — "the human-review queue has no working surface"
**Asserted:** as the title and premise of a filed bug — 292 items, four months old,
"read by nobody", with three admin routes that "have never run" and a fourth handler
needing registration. The whole file was structured around an owner decision
(queue or bin?) that had to be answered *before any code*.
**Actually:** a complete review surface exists and always has — list, filters,
per-status counts, Approve & Continue, Save & Rebuild, Retry, Resolve, editable
review forms, auto-built forms for `needs_section_data`
(`frontends/admin-dashboard/src/App.tsx:397,958,1160-1189`). It shows **0 of the 208**
build-pipeline review items, because `HandleListWorkItems` hardcodes `limit := 50`
`ORDER BY created_at DESC` (`site_admin_handlers.go:483,519`) and the frontend requests
no `status` filter, capping and then filtering client-side. The "Needs Review (N)"
count is computed over that same 50-row window, so it reads **0**. Nobody ignored the
queue — the dashboard reported it as empty. The handlers had never run for the same
class of reason: "Approve & Continue" renders only on `spec.checkpoint === true`, and
**zero of 5,622 items in the platform's history have ever carried that key.**
**Caught by:** asking "is the surface reachable?" as a separate question from "does the
surface exist?", then running the UI's own query against live data. The 0-of-208 number
is what turned a judgement call into a bug.
**The cheap check that would have caught it:** open the dashboard source and grep for
the endpoint the bug says is never called — `grep -n "work-items/.*\(approve\|resolve\|retry\)"
frontends/admin-dashboard/src/App.tsx` returns the call sites in one command. "No code
calls this" was asserted from a **backend-only** grep; the caller was in the frontend,
which was never searched. A second cheap one: run the handler's own SQL and count what
it returns.
**Cost:** contained — caught while working the bug, before anything was built. But the
file had already framed the work as blocked on an owner ruling, and the three real
defects (50-cap, hardcoded `pipeline=build` hiding 94 more items, no Ingress) need no
ruling at all. Had the ruling been "bin", the honest reading — a surface that works and
was never once looked at through — would have been thrown away with it.
**Pattern:** "nothing uses X" is a claim about the *whole* system, and a grep scoped to
one language or one directory cannot support it. Same shape as the `health.NewServer`
entry above: the absence was in the search, not in the code.

### 2026-07-20 — council gate — "the dispatch was dropped, not slow"

**The claim.** A council-gate submission produced no orchestration row 13 minutes after
publishing, so I wrote: *"Round 2 never started… the dispatch was dropped, not slow"*,
retried it, watched another 5 minutes of nothing, and started diagnosing the publish
path — probing Kafka, reading the topic tail, hunting for a payload-size threshold.

**It was slow.** The run (`0b8bcc1b`, the orchestration id my *first* submission printed)
was created at **19:20:36 — about 29 minutes after the 18:51:57Z publish** — and went on
to review normally. Nothing was dropped and nothing was wrong with the publish path.
Two other correlations dispatched at 19:01 and 19:12 showed the same "0 runs" at the
moment I looked, and were presumably also just queued.

**Caught by:** re-running the same count against a *later* clock while investigating
something else — the correlation that had shown zero orchestrations at 19:05 showed one
at 19:25. The topic read (`kcat -C -o -400`) also disproved the drop theory on the way
past: my messages were on the topic all along.

**The cheap check that would have caught it:** wait, then ask again — the runbook I
maintain says exactly this, four bullets into its own trap list: *"Runs can be slow to
start. A submission may sit before its orchestration row appears. Absence a minute later
is not evidence of a dropped dispatch."* I had **quoted that trap to the owner earlier in
the same session** and still called it a drop 13 minutes in. Writing a rule down is not
the same as having it available at the moment it applies.

**Cost:** a duplicate submission (the retry will run a second, identical 16-seat round on
the same correlation — real credits, and a confusing extra round on the trail), ~25
minutes of publish-path forensics that found nothing because nothing was broken, and a
separate 10-minute stall from a `case` statement matching lowercase `completed` against a
status that is `COMPLETED`. Contained: no wrong state was written, and the misdiagnosis
never left this session.

**Pattern:** on an asynchronous queue, *"has not arrived yet"* and *"will never arrive"*
are the same observation — a null result — and only elapsed time or reading the queue
tells them apart. My threshold for "long enough" (13 minutes) was set by impatience, not
by any measured dispatch latency; had I measured one first, the answer was ~29 minutes.
An absence needs a stated waiting period *before* you look, or it is not evidence.

### 2026-07-20 — bugs_open/001 + 039 + 040 (bugfix-001 thread) — five wrong calls in one session, four of them the same shape

Logged together because the **repetition** is the finding. Individually each is
small and each was caught; the pattern underneath them is not.

**(1) "The section went missing because the request timed out."** Written into
`bugs_open/040` as the explanation for dartsonline `index` rendering 5 of its 6
planned sections.
**Actually:** the section is dropped by builds that *succeed* — a later
`image-build-handler` rebuild reported `status='complete'` and produced the same
5-of-6. And the timeout's failing step was **`deploy_page`**, which runs *after*
component writing, so it could not have consumed a component the earlier steps
were responsible for producing.
**Caught by:** re-checking the bug's own evidence against live state after a
chassis deploy, and noticing the component timestamps had moved.
**The cheap check that would have caught it:** read the failing step's NAME before
attributing what the failure destroyed. `deploy_page` was in the error row I had
already pasted into the bug file. I quoted the record and did not read the field.

**(2) "The error was never written where a human or a sweep would look."** Same
bug, offered as a fix candidate ("propagate the orchestration error onto the work
item").
**Actually:** `agent_error_log` had it — message, step, and `work_item_id` — and
has 32,398 rows going back to 2026-04-02. The true, much narrower claim is that
`site_work_items.error` is empty. The fix candidate is a join, not a missing record.
**Caught by:** reading a new file in the v1.0.1140 diff (`agent_error_log.go`) and
going to look at whether the table it writes to already had my failure in it.
**The cheap check that would have caught it:** `\dt *error*` — one command, before
writing "never recorded".

**(3) "`testimonials` appears in neither `sections_deferred` nor `sections_skipped`."**
Written into the same bug while ruling out explanations.
**Actually:** it is in `sections_skipped`, with an explicit
`"reason": "on_missing=skip_section triggered"` — and the section is being dropped
**correctly**, because the component requires
`site_specs.social_proof.testimonials` with `min_items: 1` and the site has none.
The platform was refusing to invent customer testimonials. I had queried the
wrong orchestration: one that recorded `"no sections to plan"` and never built
anything, picked because it was the nearest run in the time window.
**Caught by:** looking for the run that actually wrote the five components, rather
than the run that was closest to the timestamp.
**The cheap check that would have caught it:** confirm the record you are reading
is the one that produced the artefact — match it to the artefact (here,
`save_sections.sections_saved = 5`), not to the clock.

**(4) "30 section entries resolve to no component."** A fleet census for
`bugs_open/039`.
**Actually 11.** The other 19 were `snake_case` and resolve fine —
`NormalizeComponentFunction` converts `call_to_action` → `call-to-action`, and I
had compared raw strings.
**Caught by:** one live page (`gaswholesalers.com`) rendering a `call-to-action`
from a `call_to_action` section entry, which my number said was impossible.
**The cheap check that would have caught it:** apply the platform's own normaliser
before declaring a mismatch — the function was in the tree, named exactly what I
was doing by hand.

**(5) "Pass B2 snapped `about` back and prevented a regression."** Written as the
proof that the `bugs_open/001` fix worked on its first live run.
**Actually:** the LLM proposed `differentiators-section` and the plan kept
`differentiators` — which are the **`name`** and the **`function`** of the *same
component row*. The snap proves the code path executed; it prevented nothing.
**Caught by:** resolving both strings against `content_components` before writing
the claim up — barely in time, and only because (4) had just made me suspicious of
string comparisons.
**Cost:** none, corrected before the claim left the session. A second re-plan later
produced two genuine rescues, so the fix is proven — but had I stopped at run 1,
`bugs_open/001` would carry a proof that was really a naming variant.

**(6) "leopardess has no clients and no case studies."** Repeated from
`bugs_open/001`'s FRESH EVIDENCE section into `bugs_open/040`, and used to
recommend **deleting** a blank case-study page.
**Actually:** the owner corrected it — the case studies describe real systems
built and running; they are simply not *client* case studies. `/case-studies.html`
is live, in the nav, and good. "No clients" had been read as "nothing to write
about".
**Caught by:** the owner, in conversation.
**The cheap check that would have caught it:** fetch the page before recommending
its deletion. I had the URL and never opened it. Also: a claim about the owner's
own business, inherited from another thread's prose, is exactly the kind to ask
about rather than act on.

> **Four of these six — (1), (3), (4), (5) — are one shape: I compared two things
> without first resolving them to the same ground.** Two records that were not the
> same run; two strings that were not in the same namespace, twice. Each time the
> comparison was *executed carefully* and the inputs were never checked for
> identity. That is why it does not feel like carelessness from the inside: the
> reasoning is sound and the operands are wrong.
>
> It maps onto the existing dominant class — reasoning from data instead of code —
> but it is a distinguishable sub-shape and worth naming separately, because the
> remedy is different. The remedy is not "read the code"; it is **one extra
> resolution step before the comparison**: which run produced this artefact, and
> what namespace is this string in. Both are single queries.
>
> This platform makes the string half of it especially easy to hit: `pages.sections`
> stores a component's `function`, `page_components` references its `name`,
> section names arrive in `snake_case` and components are `kebab-case`, and there
> is a normaliser that some paths call and others do not (`bugs_open/041`). Three
> namespaces for one concept. A census or a diff that does not normalise first will
> be wrong, and will look thorough while being wrong.

---

*(Entries below added 2026-07-20 by the bugfix-030 thread, after the six-item
synthesis above. Both are the same shape as each other and new to this file: a
rate or a state inferred from too few samples of a bursty signal.)*

**(7) "The dispatch queue drains at 0.21 msg/min; a submission waits ~6.5 hours."**
Written into `bugs_open/030` as a measured figure, from two
`kafka-consumer-groups.sh` readings labelled 14 minutes apart.
**Actually:** the two readings were **69 seconds** apart. They match, to the digit,
two samples from a continuous 30-second run another thread had going at the same
time (19:29:48 and 19:30:57). The first could not have been at 19:16, because the
committed offset was pinned at 96010 across 19:21:16–19:28:40 — it would have had
to run backwards. True rate over 22 minutes of continuous sampling: **~2.4
msg/min**, so a queue 86 deep clears in ~36 min, matching the 25–36 min the same
bug file had already measured the day before. The figure was ~12× too slow.
**Caught by:** a second thread that happened to be sampling the same consumer group
continuously, and recognised its own sample values in the other thread's write-up.
That is luck, not process.
**The cheap check that would have caught it:** *subtract the timestamps you are
about to divide by.* The interval was the load-bearing number in the calculation
and was never verified against the wall clock. Second-cheapest: sample three times,
not twice — the third reading falsifies a straddled stall immediately.

**(8) NEAR-MISS, same session, opposite direction: "consumption has stopped; ~19
hours of backlog."** Not written down — caught before it left the session, so it is
logged here as a near-miss because the tally is the point.
**Actually:** I sampled for ~2 minutes, saw the offset frozen and `LAG` climbing,
found zero fetch lines in a 30-minute log window, and was drafting a live-outage
claim. Ten minutes later the offset jumped **+47 messages**. Nothing was wrong.
**Caught by:** the frozen-offset landmine already in `bugs_open/030` — written by
the bugfix-028 thread the same week, after making this exact error. **The landmine
worked**, which is worth recording: it is the clearest evidence in this file that
writing these up changes a later session's behaviour.
**The cheap check that would have caught it:** keep sampling. The signal is a
sawtooth — it pins for ~8 minutes while one long message is processed, then bursts.
Any window shorter than one full tooth is uninformative in *both* directions.

> **(7) and (8) are one class, and it is not the "resolve your operands" class
> above — it is closer to "reasoning from data instead of code", with a twist.**
> Both read a *bursty* signal over a window shorter than its period and reported the
> instantaneous value as the steady state. (8) got the direction wrong and would have
> declared a healthy cluster dead; (7) got the magnitude wrong by 12× and pushed a
> "multi-hour, no failure signal" warning into a bug file that other threads plan
> around. Note they were made by different threads, hours apart, on the *same
> signal* — so this is a property of the signal, not of either session.
>
> **Remedy, and it is mechanical: a rate or a liveness claim about the dispatch
> queue needs ≥20 minutes of continuous sampling, and the slope, not two points.**
> The command is in
> `docs024_key_docs_latest/dispatch_queue_serialisation/RUNBOOK_dispatch_queue_serialisation.md`
> R2. The deeper tell is that **both threads had the code available and used the
> clock instead** — the reason the queue behaves this way (each message is processed
> synchronously, and a workflow's consecutive steps run inline on the consumer
> goroutine) is legible in `agent.go` and `coordinator.go` in about ten minutes, and
> it predicts the sawtooth outright.

---

**(9) "Generated contact forms POST to a dead `/contact` endpoint, fleet-wide" —
cited `k8s/bk_page_components.sql:140` as the emitter.** Written into
`bugs_open/006` §B on 2026-07-16, carried unchallenged for four days, with fix
options costed against it.
**Actually:** on 2026-07-20 the live fleet had **zero** components with
`action="/contact"`. The template is `action="{{.form_action}}"` fed from each
component's own `content_data`; live values are `#contact` (8 sites), `""` (3)
and one hand-fixed `mailto:`. The *defect* was real and arguably worse than filed
(10 of 11 live sites cannot deliver a contact message) — but the cause, and
therefore both proposed fixes, were aimed at a string that is not in the system.
**Caught by:** one `GROUP BY` over `page_components` before starting the fix.
**The cheap check that would have caught it:** query the live artifact for the
literal you are about to fix, *before* costing the fix. Ten seconds.

> **The transferable trap is the citation, not the staleness.** `bk_*.sql` files
> are **backup dumps of a table**, not source. Citing one as the place a value is
> emitted reads exactly like citing code — same path shape, same `file:line` — and
> it silently converts "this row existed in March" into "this is what the platform
> generates". The real emitter here (`bk_content_components.sql:134`) is *also* a
> dump; the actual authority is the live `content_data`, and **no Go code sets
> `form_action` at all**, which is the finding that matters and which no amount of
> grepping the dumps would have produced.
> **Rule: a `bk_` path is evidence about a past table state and nothing else.**
> Before citing one as a cause, grep the Go tree for the field name — if nothing
> sets it, the value is content, not code, and the fix is a default plus a
> validation check rather than an edit to a template.

**(9) "The queue drains at ~2.4 msg/min; 86 deep clears in ~36 min; nothing had
degraded."** Mine (bugfix-030 thread), written into `bugs_open/030` **as the
correction to entry (7) above** — i.e. I made this error in the very act of
documenting it, two paragraphs after writing "any two samples inside a burst give
an arbitrarily high rate".
**Actually:** I keyed on a window containing the +47-message burst and ending just
before the stalls that followed. Straddling a burst, where (7) straddled a stall.
The completed 40-sample run: **0.62 msg/min** over the full 22.8-min sampler window,
**1.50 msg/min** over all 40.7 min I watched, with `LAG` **growing 82 → 130** and a
single message pinning the offset for **≥15.4 minutes**. The queue was *diverging*.
"Clears in ~36 min" was false — it never cleared and got 48 deeper.
**Worse than the arithmetic:** the thread I corrected had concluded "slow,
multi-hour possible, variance is large, no stable answer". That was **right**, and I
overturned it with a wrong number while correctly faulting their derivation. Being
right about their method licensed me to be wrong about their conclusion.
**Caught by:** letting my own sampler finish. The first 21 samples said one thing;
all 40 said the opposite. I published at 21.
**The cheap check that would have caught it:** *wait for the measurement you already
started.* I had a 20-minute run in flight and wrote the correction from a partial
read of it — for no reason except that I had the answer I expected.

> **(7), (8) and (9) are now three threads, three "measured" rates from the same
> queue on one afternoon — 0.21, 2.4, 0.62 — each arithmetically defensible and each
> useless as a forecast. At that point the fault is not in anyone's window choice.
> There is no single rate to measure**: throughput is `1 / (duration of the
> orchestration segment currently running inline on the consumer goroutine)`, which
> ranges from milliseconds to ≥15 minutes depending on what is at the head. An
> average over that describes no moment and predicts nothing.
>
> **So the remedy in (7)/(8) — "sample ≥20 min and take the slope" — is itself
> wrong, and I am retracting it.** A longer window makes the number stabler, not
> truer; it still answers a question the system does not have an answer to. Report
> `LAG` (depth) and what kind of work is at the head, publish raw samples if you
> report anything, and when asked "how long will it take" say **there is no reliable
> answer** rather than producing a defensible one.
>
> The generalisable shape, and it is new to this file: **a metric can be
> well-defined, correctly computed, and still not exist as a property of the
> system.** All three of us assumed the rate was a fact awaiting better measurement.
> Two of us then "improved" the measurement and got further from the truth. The tell
> was available to all three: the mechanism is legible in `agent.go` and
> `coordinator.go` in ten minutes and predicts non-stationarity outright — we each
> reached for the stopwatch with the source code open.

---

### 2026-07-20 — bugfix 023 — "70 ungated CTA anchors across 37 components"

**The claim.** That the remaining gating sweep was **70 ungated anchors / 37 components**. It sat
in `bugs_open/023`'s verify-criteria table, in the PLAN's P2.1 sizing, and in three separate status
updates written across two days (75/38 at filing, 70/37 after migration 179). I inherited it and was
about to build the edit worklist on it.

**What was true.** A real template parse gives **171 ungated anchors across 41 components** — a
**2.4× undercount**. The mechanism: `regexp_matches(col, pattern, 'g')` returns *non-overlapping*
matches, and R9's `.{0,60}` lookback prefix is part of the match, so each match consumes the text
before it. In a run of adjacent anchors — nav lists, footer link columns, i.e. **precisely where CTA
anchors cluster** — every second anchor is swallowed and never counted. The error was therefore
largest exactly where the problem was densest.

**Caught by:** doing what the runbook entry told me to do. R9 carried the line *"Good enough to size
the problem and to prove the direction; **re-derive the exact list with a real template parse before
mass-editing**"* — attached to the figure from the day it was written.

**The cheap check that would have caught it:** the one already written down, next to the number,
in the same file. Nobody had to invent a method; the correction cost about fifteen minutes of
scripting.

**Why this one is worth a row.** Not because a heuristic was imprecise — it was *labelled*
imprecise and honestly so. Because **a figure with a caveat attached travels without its caveat.**
It was copied into a bug file, a plan and three status updates, and by the third copy it read as a
measurement. The tell was never hidden; it was one line below, every time, and each thread
(including me, for the first hour) read the number and skipped the sentence under it.

**Generalises to:** any figure carried between documents. If you repeat a number you did not
measure, carry its caveat or re-derive it — and if the caveat says *"do X before acting on this"*,
then reaching the point of acting **is** the trigger to do X. A warning attached to a measurement
is not a disclaimer, it is a pending task.

Related: `016b` §9 *"A `regexp_matches(…,'g')` census with a lookback prefix silently drops every
other match"* (the mechanism, transferably), `bugs_open/023` (correction + resizing), RUNBOOK R9/R9b.

---

## "The configs will self-heal once the code ships" — asserted from a count, never queried (2026-07-20)

**The claim.** `bugs_open/009` §4, written 2026-07-17, sized the fleet sweep like this: *"17
agent defs have a root `ai_service` with NO max_tokens → every call at 2048. **10 of them ALSO
declare max_tokens elsewhere (dead)**"*, and concluded *"If the overlay fix ships FIRST, the 10
step-level declarations start working on their own (configs self-heal — this is the argument for
code-first)."* That reads as a measurement and it shaped the sequencing recommendation.

**What was true.** Running §4's own audit query against live `agent_definitions` before writing
the fix: **16** such agents, not 17 (`content-creator-contact` has two rows and was
double-counted), and **zero** of them declare a step-level `ai_service.max_tokens`. Not ten —
none. The "self-healing" population is empty. Every one of those agents still runs at the
hardcoded 2048 after the code fix ships, and still needs a deliberate per-agent decision.

**Caught by:** running the query in the section I was reading, instead of carrying its prose
forward into my own commit message. It took about ninety seconds.

**The cheap check that would have caught it:** the SQL was *printed in the same section as the
claim* — §4 even flags its own snippet as pseudo-code needing a real query ("practical version:
… eyeball each def"). The claim was built from an eyeball pass and then written in the register
of a count.

**Why this one is worth a row.** The damage was not the number. It was that "configs self-heal"
made the sweep sound like it would take care of itself, so a thread that shipped the code fix and
stopped would have left 16 agents capped at 2048 believing the job was done — and the bug file
would have agreed with them. **A wrong figure gets corrected; a wrong *implication* gets acted
on.** Note also that the surrounding mechanism was completely correct and had been proven by
direct experiment; being right about the hard part is not evidence about the arithmetic.

**Generalises to:** any "N of them also do X" in a doc you are about to sequence work from. A
population count is a query, not a recollection — and if the doc hands you the query, the fact
that you are reading the section **is** the trigger to run it. Same shape as the R9 anchor-count
row above: a figure travelling between documents, arriving as fact.

Related: `bugs_open/009` §7 (the live audit + the correction, in place), `016b` §9 *"First-found-
wins config lookup makes the MORE SPECIFIC block dead"*.

### 2026-07-20 — bugs_closed/001 close-out (bugfix-001 thread, second session) — two wrong calls, one of them nearly shipped

**(1) "The residual is real; the fix is a small Go change + test."** Reported to the
owner as one of four ready workstreams, quoting `/bugs_open/001`'s own prescription
for its surviving residual: *"take the LLM's sections when the realised ones are
empty"*. I had read the code, confirmed the code path, and measured the reachable
set at 18 pages. All of that was correct and the conclusion was still wrong.
**Actually:** those 18 pages are what the fix would have damaged. All of them are
`tool` (14), `blog-index` (2) or tool-ish `content` (2), and **15 of the 18 render
content while carrying no sections at all** — they are composed by a different
subsystem. For a **deployed** page, `sections=[]` is a positive statement ("not
section-composed"), not an absence awaiting a fill. Implementing the prescription
would have let a re-plan attach a generic `hero`/`features` layout to every one of
them: content injection onto built pages, which is the exact failure class
`/bugs_open/001` exists to prevent. Filed properly as `/bugs_open/050`.
**Caught by:** reading the doc comment on `normaliseRealisedToPlanPage` — which
explains that the union carries `sections` *because* the page sync's
`<col> = EXCLUDED.<col>` would otherwise clobber them — and stopping to ask what an
empty value in that column actually asserts, rather than what the bug file assumed
it asserted.
**The cheap check that would have caught it:** `SELECT` the rows the change would
affect and **look at them** before writing the change, not after. One query
described the whole population as tool pages and settled it. I had already counted
those 18 rows for the report — I used the COUNT and never looked at the ROWS.
**Second-order lesson, and the more useful one:** a residual documented by a
previous session arrives with its diagnosis pre-attached and reads as a spec. It is
a *hypothesis by someone who had stopped working*, and it deserves the same
re-derivation as any other inherited claim. This one had been written, reviewed by
its author, and carried into a handoff, and it was still wrong in the direction that
would have caused damage.

**(2) Guessed a column name instead of reading the schema.** Queried
`p.adoption_locked` on `pages` to measure the preserved set.
**Actually:** there is no such column — `adoption_locked` is *derived per query* by
a `CASE` in the planner's `load_existing_pages` step, which is the finding that
became `/bugs_open/051`. The error was free (Postgres refused it) and it led
somewhere useful, but that was luck, not method.
**The cheap check:** `\d pages`, which CLAUDE.md already mandates as "schema first"
and which I skipped because I "knew" what the column was called from reading Go code
that reads a **map key**, not a column.

**Both are the same underlying move as the four logged from this workstream's first
session:** trusting a name or a count as a stand-in for the thing itself. A map key
is not a column; a row count is not the rows.

**One tally row incremented ("look at the real values", 3→4); two new rows.**

> **CORRECTED immediately after writing:** this line first read *"Two tally rows
> incremented; one new row"*, which is the wrong way round. Noting it rather than
> silently fixing it, because miscounting my own tally in an entry about not
> trusting counts is the joke writing itself — and because these two edits were
> swept into another session's commit (`d90a2a5b0`, a bug-009 commit) between the
> append and the commit, which is the CLAUDE.md hazard doing exactly what it says
> on the tin. Nothing was lost; forward-only holds; the content is simply recorded
> under someone else's message.

---

## 2026-07-20 — bugfix-016 thread: "the bug file's own closing line is wrong" (it wasn't)

**The claim I nearly wrote.** `bugs_open/016` ends with *"016 stays OPEN on finding
1 — no fix-proposer repropose has STARTED post-fix"*. An earlier section of the
same file shouts that the fix is *"PROVEN IN THE WILD"* via run `a8b66dee`
(started 15:27:33Z, well after the 13:15:11Z fix, `<no value>`: false). I
reproduced that query, got exactly the documented numbers, and concluded the file
contradicted itself and that finding 1 was provable-closed. I was about to report
"the closing line is stale — 016 finding 1 is proven".

**Actually:** `a8b66dee` is the **feature-designer**, not `fix-proposer`. Its
steps are `load_spec`/`design`/`check_spec_approved` + 5 seats; fix-proposer's
are `load_diagnosis`/`propose`/`code_lookup`/`select_panel` + 13 seats + 12
`gate_*`. Fix-proposer's last repropose belongs to `48cf0339`, started 13:11:13Z
— **four minutes before** its own fix. The closing line is right; the
PROVEN-IN-THE-WILD section is the misattributed one. Had I "corrected" the file,
I would have closed a finding on evidence from a different agent, and the next
thread would have inherited it stated with confidence.

**What caught it:** reading the *rendered prompt text* on a whim, not the
timestamps. It opens "REVISE the staged build plan … stages are commits …
capabilities listed missing" — designer vocabulary. Fix-proposer revises an edit
plan against a diagnosis. Nothing in the numbers would ever have shown this.

**The cheap check:** the file *already told me*. It warns two sections earlier
that `persist_plan` is not fix-proposer-only and that filtering on
`agent_type='fix-proposer'` reads as "never ran" because the chassis stamps rows
`generic`. `repropose` is shared by the same three agents for the same reason. I
read that warning, applied it to `persist_plan`, and did not generalise it to the
step I was actually querying. **One `jsonb_object_keys` on `workflow_plan->'steps'`
— which is what finally settled it — costs one query and identifies the agent
unambiguously.**

**The move, again:** trusting a *step name* as a stand-in for the agent that ran
it — the same "a name is not the thing" failure as the `adoption_locked` map-key
row above. New wrinkle worth its own line: **a document warning you about a trap
does not inoculate you against that trap in its neighbouring paragraph.** The
warning was scoped to one step name and I let that scoping stand.

**Tally:** "look at the real values, not the name" 4→5. One new row.

> **Postscript, same session, ~1 hour later — I did it again with the corrected number.**
> Having replaced 70/37 with **171/41**, I wrote 171 into the bug file, this file, the
> debugging guide and a commit message. **15 of those 171 are not CTA anchors.** They sit
> inside a `{{range}}`, so the field belongs to the ranged item, not the component:
> `{{range .items}}<a href="{{.url}}">` is an item link from a query-provided list — a
> different class with a different fix. The CTA worklist was **156/32**, not 171/41.
>
> **Caught by:** writing a migration whose post-condition was a blanket *"no ungated
> `{{.x_url}}` anchor remains"*. `tool-cta`'s range-scoped `.url` would have tripped it and
> rolled back a correct change. I read the third anchor before applying, and the distinction
> fell out. So it was caught by **the strictness of a gate I wrote for a different purpose**,
> not by re-examining the number.
>
> **The lesson is not "be more careful with numbers."** It is that a corrected figure inherits
> the authority of the correction and stops being questioned — I had just finished documenting
> that a figure travels without its caveat, and then produced a fresh figure with a fresh
> uncaveated meaning. The durable fix is not a better number, it is **making the tool emit the
> distinction**: `parse_gates.py` now prints the range/CTA split on every run, so the next
> reader cannot collapse them by accident. Where a caveat has to be remembered, it will be
> forgotten; where the tool prints it, it survives.

## 2026-07-20 — "it's gone through the review council" (bugs_open/010 b)

**The claim.** In `travelling_docs/README_where_we_are.md`, written and committed
within minutes of firing the submission: the convergence guard "has gone through
the review council".

**Why it was false.** I had *submitted* it. No verdict existed — no
`council_report` artifact, no `orchestration_state_audit` row, the run had not
started. I described a submission as a review.

**What caught it.** Re-reading my own paragraph before moving on, prompted by
the standing rule that a review claim is earned by a verdict and nothing else.
Nobody else had read the file yet.

**The cheap check that would have.** The one already printed by the 097 trigger:
`SELECT metadata->>'decision' FROM diagnosis_artifacts WHERE correlation_id=...
AND kind='council_report'`. Empty result = no verdict = no review claim.

**Tally row: stating an in-flight process as a completed one.** Same shape as the
`Council-Reviewed:` trailer failure (2026-07-19, a trailer put on a REVISE) and
as the queue-latency trap (a missing run row read as a drop rather than a
queue). All three are the same error: **treating the absence of a result as a
result.** The fix in each case is to name the state you actually observed —
"submitted", "queued", "no artifact yet" — rather than the state you expect.

---

## 2026-07-20 — `git stash pop` with no stash of my own popped another branch's WIP

**The call.** Mid-task, to verify whether HEAD's discovery_checks package compiled without my
uncommitted test changes, I ran `git stash push -- <two test files>` then `git stash pop`. The
push found "No local changes to save" (my changes had already been swept into a commit), so it
created **no** stash — and the subsequent `pop` therefore popped `stash@{0}`, an **unrelated WIP
stash from another branch** (`066_hitl_questionnaire`). It half-applied and left a merge conflict
in `coordinator.go`, staging a new file (`awaited_requests_repo.go`) that belongs to that other
work.

**What caught it.** The `git status` immediately after the pop showed files I never touched
(`coordinator.go` UU, `awaited_requests_repo.go` A) and the pop's own output naming an unrelated
SQL file. `git stash list` then showed `stash@{0}` was another branch's.

**Why it was recoverable with nothing lost.** A `pop` that hits a conflict does **not** drop the
stash — `stash@{0}` stayed intact, so discarding the working-tree application lost nothing. I
restored `coordinator.go` to HEAD and removed the pop's new file (verified byte-identical to the
stash first), touching none of the *other* concurrent sessions' live WIP.

**The cheap check that would have.** Two, either sufficient:
1. `git stash list` **before** any `pop`. On a shared tree the stash stack is not yours; a bare
   `pop` acts on whatever is at `{0}`.
2. Never use `stash push`/`pop` to reverse *your own* change. If the goal is "test HEAD without
   file X's local edits", use `git checkout HEAD -- X` on a named path (or a throwaway worktree),
   which cannot touch anyone else's stash.

**Tally row: acting on `{0}` as if it were mine.** Same family as the queue-latency and
Council-Reviewed traps already logged — assuming a shared, mutable slot reflects my own state.
On this tree, *nothing* at rest is private: not the index (commit-per-task exists for that), not
HEAD (it moved four times mid-session), and not the stash stack.

### 2026-07-20 — travelling docs / bugs_open/024 — submitted a TRUNCATED SQL sketch and called it complete
**Asserted:** in the round-6 council submission, that migration 180's edit "now
opens with a defensive ROLLBACK, has a pre-flight that counts target rows and
RAISEs… a post-condition block that re-counts and RAISEs… writes a
('pipeline','build') doc_note" — i.e. the rationale described a complete,
hardened migration.
**Actually:** the `sketch` field the reviewers actually saw was cut off
mid-statement, inside the `jsonb_build_object` call, with no closing UPDATE, no
RETURNING, no post-condition block, no doc_note, no COMMIT. I had built the
submission JSON with `open(...).read()[:5200]` — a hard character slice — and the
migration is longer than 5,200 chars. The FILE on disk is complete and was
applied cleanly; the REVIEW COPY was truncated.
**Caught by:** the council's editquality seat, at MEDIUM, on the single
load-bearing edit. Not by me, and the irony is total: this entire workstream
exists because agents persist truncated artifacts and report them as complete
(`output_tokens==max_tokens`, bugs_open/012), and here I truncated an artifact
and asserted its completeness in the same breath.
**The cheap check that would have caught it:** after building a submission that
embeds a file, read back the embedded field and check it ends where the file
ends — `submission.plan.edits[i].sketch.endswith('COMMIT;')`. Or do not slice at
all: embed the whole file, or a deliberately-marked excerpt that says it is one.
A fixed `[:5200]` is a truncation with no marker, which is exactly the shape 016b
§9 warns about — I built the machine's failure mode by hand.
**Cost:** one wasted council round on that axis — editquality could not judge the
edit's completeness and objected instead of approving. No production impact; the
applied migration was whole. But it is the second row in this file where I put a
truncated thing in front of a reviewer, and the first where I did it in a
submission whose whole subject is truncation.

### 2026-07-20 — vetcomparison — declared the content-feed sweep "broken fleet-wide" from ten minutes of absent rows
**Asserted:** that the 6-hourly `content-feed-refresh` sweep was completing without
doing any per-site work across the whole fleet — and filed it to the diagnosis
loop (`12ff5852`) on that basis.
**Actually:** the sweep was LATE, not broken. It fetched ~90 minutes later and
every site went through cleanly. The likely trigger was a chassis pod roll (a fresh
build had just deployed) silently dropping the spawn — a mechanism CLAUDE.md already
documents.
**Caught by:** waiting and re-querying: `max(last_fetched_at)` across all
`content_sources` advanced on its own. I then withdrew the diagnosis item with the
reason attached.
**The cheap check that would have caught it:** before calling a scheduled job
broken, confirm the scheduler is even beating (`last_triggered_at` on a 30s task)
AND wait at least one more interval. Absence of a row on a system that queues is not
evidence of failure — it is the default state between cycles.
**Cost:** one wasted diagnosis-loop filing, later withdrawn.

### 2026-07-20 — vetcomparison — called a council submission "dropped" and resubmitted it; it was invalid, twice
**Asserted:** that my council submission (`712be028`) had been silently dropped,
because 13 minutes after submitting it had zero `orchestration_state_audit` rows and
showed "not-started". I resubmitted (`563462b8`), and wrote into `bugs_open/043`
that "the council gate is not starting either… the same machinery stalling."
**Actually:** BOTH runs started and completed — about an hour after submission, not
dropped at all. CLAUDE.md (updated 2026-07-20, which my session-start copy predated)
says publish→run-start was measured at ~29 minutes and that a missing orchestration
row is almost always latency; retrying costs a duplicate round. I did the exact
thing it warns against. Worse: the submission was structurally INVALID both times —
it failed at `persist_submission` with `edit 2: operation "create" not in the
allowlist`, because I included a file-*create* edit (the new test file). So neither
run could ever have produced a verdict, and the resubmission repeated the same
invalid plan.
**Caught by:** re-reading CLAUDE.md the next day, then finding the runs by PAYLOAD
(`collected_data->'input_data'->>'fix_correlation_id'`) rather than the printed id —
both COMPLETED at `complete_invalid`, and `__step_error` named the allowlist
rejection.
**The cheap check that would have caught it:** (1) budget ~30 min for a council run
and find it by payload before concluding anything; zero audit rows means QUEUED, not
dropped. (2) The council fix-plan allowlist is `modify | remove | config_change` —
NOT `create`. A submission that adds a file will be rejected at intake; describe the
new file's content some other way, or expect the create edit to invalidate the whole
plan. Both were client-side knowable before spending a single credit.
**Cost:** two wasted council submissions, and a wrong causal claim written into
`bugs_open/043` (now corrected) linking the council to the diagnosis-loop route-hang.

### 2026-07-20 — vetcomparison — reported a diagnosis "filed" when it had never registered
**Asserted:** to the owner that the directory-exporter diagnosis was filed as
correlation `2c5bb9e2`.
**Actually:** no intake row was ever created — the 090 trigger had refused or errored,
and I never saw it because I piped its output through
`grep -iE "Correlation|SAVE|item_key"`, which hid the refusal line. I reported a
filing that did not exist.
**Caught by:** checking `site_work_items` for the item_key afterwards and finding
nothing; refiled properly as `55dc0fa4` and verified the row this time.
**The cheap check that would have caught it:** read the trigger's FULL output (it
prints "intake recorded and dispatched" on success), or assert the intake row exists
before reporting a correlation id as filed. A narrow grep over a tool's output can
hide the one line that says it failed.

### 2026-07-21 — bugs_closed/042 grouped its sibling as a "literal-string" bug without reading the failing action
**Asserted:** (in 042's §Related and fix-candidate 2) that the directory-exporter
sibling — `DirectoryExportAction` aborting on an empty domain, correlation `55dc0fa4` —
was the same `ExtractActionInputs` literal-string family as 042 itself: "a literal domain
string does not reach the action".
**Actually:** `DirectoryExportJSONAction` does not use `ExtractActionInputs` at all; it
reads `config["domain"].(string)` directly. The domain is an ordinary string. The real
cause is that the scheduled task's `input_data` was authored as a full message envelope,
which the scheduler's `fireTrigger` wraps a SECOND time, so the domain lands at
`input_data.input_data.domain` — one nesting level below where the action reads. A data
bug, not a code bug; fixed live in the DB + seed with no image roll (`bugs_open/054`).
**Caught by:** reading the failing action's source and running
`jsonb_pretty(collected_data->'input_data')` on the failed run (`6271b72d`), which showed
the double-nested envelope verbatim.
**The cheap check that would have caught it:** open the action named in the symptom
before assigning the bug to a family. The family label was inferred from the shared
symptom ("required field absent → abort"); one read of `directory_export_action.go`
would have shown it shares neither the code path nor the mechanism. Symptom-family is not
cause-family.
**Cost:** none downstream — caught before the literal-string `ExtractActionInputs` change
was attempted on a bug it could not have fixed. The literal-string half of 042 remains a
real, separate gap.

---

## 2026-07-21 — "the validation blocker reason is unrecoverable from the DB" (brochure_component_library / fundamentallyai.com)

**Claim (in HANDOFF_2026-07-21_start_here.md + the 07-21 SUMMARY/README):** the
content-validation blocker that held 5 fundamentallyai.com pages was NOT
recoverable from the database — `site_work_items.result` is empty and the
overnight pod logs rotated on the v1.0.1144 restart — so the next thread must
re-fire a page live and tail the chassis log during `validate_page_content` to
capture it (a ~30-min single-consumer queue wait per `bugs_open/030`, plus a
live log-tail).
**Actually:** `ValidatePageContentAction` *deliberately persists* the full
structured issue list to `agent_error_log` on failure
(`writeValidationFailureLog`, validate_page_content.go:344-420,
`error_code='CONTENT_VALIDATION_BLOCKER_DETAIL'`) — the code comment says in as
many words that it exists so "post-mortem debugging doesn't require pod-log
access" and survives log rotation. 9 such rows were already in the DB, each
naming the exact blocker (`cross_site_domain` / `leopardessconsulting.co.uk`).
The premise checked only ONE table (`site_work_items.result`) and generalised
"empty there" to "unrecoverable anywhere".
**Caught by:** reading `validate_page_content.go` end-to-end *before* acting on
the handoff's re-fire prescription — the persistence path is right there in the
same function that emits the error.
**The cheap check that would have caught it:** `SELECT DISTINCT error_code FROM
agent_error_log WHERE site_id = <site>` — one query, ~0.2s, before concluding a
failure reason is "not in the DB". A gate that returns a deliberately-vague
error almost always logs the detail somewhere on purpose; grep the action for
where it writes on the failure path.
**Cost:** none — the wasted re-fire + queue-wait + log-tail was avoided; the
diagnosis was free and immediate. Recording it because "checked one table,
declared it unrecoverable" is a repeatable shortcut worth naming.

---

## 2026-07-21 — "no existing loop-controller action" — an ABSENCE claim shipped without the search (fixloop feature builder 2)

**The claim:** in the delta-2 council-gate resubmission (corr `5a65ec4c`, round 2)
I wrote, to justify the new `feature_stage_route` action against reuse_agent's
round-1 objection: *"Registry search for an existing generic 'stage advance' /
'loop controller' action: none — the diagnosis loop's iteration is inline control
flow, not a reusable action, so feature_stage_route is genuinely the only new
control machinery."*

**Actually:** the registry has a generic **`loop`** action ("Iterate over a
collection, executing sub-steps for each item", registry.go:47) AND
`loop_complete` (:53) AND `conditional_route` (:73) — core-category iteration
primitives. Whether `loop` could actually host the stage loop is a real open
question (the bespoke parts are emitting each stage AS a single-plan shape and
threading branch/ref between iterations — a generic collection-loop may not do
that), but I never opened `LoopAction` to find out. I asserted "none" having
grepped the registry only for my OWN new action's name.

**Caught by:** `prior_art_librarian` in round 2 — verdict object, HIGH: *"a
load-bearing absence claim with no attached search output ... the defect class
this council seat exists to catch is exactly this shape."* Correct, and the seat
named the class precisely.

**The cheap check that would have caught it:** `grep -nE '"[a-z_]+":'
registry.go | grep -iE 'loop|iterate|stage|route|sequence'` — one command, and it
surfaces `loop`/`loop_complete`/`conditional_route` immediately. Attach the
output to the submission instead of asserting the absence.

**Why this one stings:** the workstream's own memory
([[reread-claudemd-and-standing-docs]]) already records this exact failure mode
from 2026-07-19 ("asserted only one call site could carry a defect, having read
four of eight and inferred three from their filenames"), and CLAUDE.md's
"Diagnosis before debugging" section is built around "confidence is not a signal
… the failure mode is not missing information — it is not looking." I did it
again, on a claim I had time to check, in a submission TO the very council that
exists to catch it. The gate worked; I didn't.

**Cost:** none beyond the credits of the round it was raised in — the code is not
wrong (the loop-vs-bespoke-action question is a design judgement, unresolved), but
the *claim* was, and a round-3 close now owes a real `LoopAction` read, not a
better-worded assertion.

## 2026-07-21 — a monitor re-read round 2's verdict as round 3's (bugs_open/010 b)

**The claim (nearly).** A background monitor watching for the council round-3
verdict emitted "COUNCIL ROUND 3 verdict: revise". I was one query away from
recording round 3 as REVISE and stopping the council process on it.

**Why it was false.** The monitor filtered council_reports by
`correlation_id AND created_at > '10:58:00'` — a cutoff I set to "just after
round 2". Round 2's report was actually stamped **10:58:10**, ten seconds later,
so it slipped past the cutoff and the monitor returned round 2's decision as if
it were round 3's. Round 3 (orch 1e288285) had produced **zero** artifacts — it
was still queued behind the ~30-min dispatch latency.

**What caught it.** Dumping the report body to parse the objections — it came
back **0 bytes**, because I filtered that query by round 3's `orchestration_id`
and there was no such row. The emptiness, not the decision, exposed it.

**The cheap check that would have.** My OWN recorded guidance, in this
workstream's NOTES: *"Poll diagnosis_artifacts filtered by orchestration_id,
never by correlation alone — every round shares the trail correlation."* The
monitor used correlation + a hand-set timestamp instead of the round's
orchestration_id. A `created_at >` cutoff is a footgun: it must be strictly
greater than the previous round's EXACT timestamp, and "round it to the minute"
loses by seconds.

**Tally row: reading a shared-correlation council report by a loose filter.**
This is the THIRD of this family — trailer-on-a-REVISE (07-19),
submission-read-as-verdict (07-20), now stale-round-read-as-new (07-21). All
three treat a stale/absent result as the current one. The durable fix is
mechanical: **always key a council poll on the round's orchestration_id**, not
on correlation + time. I wrote that rule and then didn't follow it in a monitor.

### 2026-07-21 — diagnose-dispatch resilience (bugfix-001 thread) — "raise max_attempts, it's a safe one-liner that makes the loop resilient"
**Asserted:** to the owner, recommending a fix for the diagnosis queue burning items on
transient spawn failures — *"Raise max_attempts to 2–3 on the diagnosis items so a single
spawn flake doesn't burn a request. One-line change; makes the loop resilient."* The owner
approved it on that basis.
**Actually:** `max_attempts=1` on `needs_diagnosis` items is **load-bearing**, and the 090
trigger's header note 5 says so. Raising it re-exposes a laundering hazard, every link of
which I confirmed live only *after* recommending the change:
- the failure/verification path re-queues a not-yet-exhausted item to **`status='triaged'`**,
  not `awaiting_diagnosis` (`complete_work_item_verification.go:245`, `ELSE 'triaged'`);
- the **build** dispatcher claims `status IN ('triaged','approved')` with **no pipeline
  filter** in its live config (`load_work_item_actions.go`, `build-dispatch-loop.load_items`
  sets only `site_id`/`max_items`);
- its trigger fires for `system.internal`, which is an **unlocked** `sites` row that **does**
  carry `pipeline='build'` items (41). So a `triaged` diagnose item on the system site is
  reachable by a build handler — a diagnosis spec handed to the wrong pipeline.
`max_attempts=1` forces the failed diagnose item to terminal `failed` instead, by construction.
So the value I called an oversight is a guard, and raising it trades a contained
"burns-on-flake" for an uncontained "wrong-pipeline-claims-it".
**Caught by:** reading the *contract* before executing the approved change — the retry path's
target status, and the build claim's pipeline filter — instead of assuming "more attempts =
more resilience". The requeue goal was then met safely a different way: set the 4 transient
failures to `awaiting_diagnosis` + `attempt_count=0` (a status only the diagnose loop claims),
leaving `max_attempts` at 1. No laundering surface touched.
**The cheap check that would have caught it:** before calling a retry-count change "safe",
answer two one-line questions — *what status does the failure path move the item to, and who
else claims that status?* Both are a single grep / config read. I answered neither before
recommending.
**Cost:** none realised — caught before the change, and the owner's approval was acted on in
the safe form, not the unsafe one I had described. But the recommendation itself was the error:
an "approved" action is only as sound as the basis I gave for it, and I gave a false one.
**Pattern:** same family as the "resolve both operands" cluster above, one level up — I reasoned
about a *change* from its stated intent ("more retries") instead of from the *code path it
plugs into*. "Read the CONTRACT a thing plugs into" (existing tally row) applied to a config
value, not just a code call.

## 2026-07-21 — tried to prove a guard on a tool I assumed was still broken (bugs_open/010 b)

**The claim (acted on).** I queued a live acceptance run on `tool-loot-table-balancer`
to watch the new convergence guard escalate it, on the basis that it was "past
threshold — 4 complete cycles at mobile-fit, verified live". I set up a monitor
to catch the escalation.

**Why it was wrong.** The count of 4 is about the PAST — four failed repair
cycles between 07-17 and 07-18. It says nothing about whether the tool fails
NOW. Between 07-18 and 07-21 the tool was fixed (its grid became a wrapping flex,
024's delivery chain rendered it), so the run PASSED and the guard's escalation
branch never executed. I proved nothing about the escalation, and briefly framed
an unrun branch as a proof-in-flight.

**What caught it.** Checking the actual run outcome instead of assuming — the
acceptance-run note said PASSED, and the live `.ltb-row-grid` rule was
`display:flex;flex-wrap:wrap`, not the overflowing grid.

**The cheap check that would have.** 024's OWN lesson, which I even cite in this
bug: verify a tool's state by its specific rendered rule, not its history. One
`substring(html_template from '\.ltb-row-grid…')` before firing would have shown
the flex-wrap and told me it would pass.

**Tally row: treating a historical failure count as current state.** A count of
past failures is not a measurement of present failure. Before inducing a failure
path on a specific target, RE-MEASURE the target now — the world moved since the
count was taken (here: another workstream fixed the tool). Same family as the
stale-figure entries: a number carried forward unchecked.

---

**2026-07-21 — shipped a fix under a headline framing a concurrent thread was, at
that moment, refuting.** Commit `cdd858402` fixes `bugs_open/024` defect 6 (the
`page_rerender` item_key mode-collision) and its message + my council submission
(corr `746c7d60`) call it *"the delivery blocker that kept every tool-improver fix
off the live page."* **That framing was already being overturned in the same
working tree.** Another travelling-docs session had found that the whole generic
section-render path defect 6 sits on is **deliberately forbidden for tool pages**
by the experience-loop's ownership guard (`save_page_sections_action.go:138`,
migration 164 — `rebuild_policy='owned'` hard-refuses the write), so defect 6 is
**moot for tools**; the sanctioned path (`section-editor`/`apply_section_edit`)
had already delivered the benchmark fix LIVE (rendered_html 9,901→10,705, grid→
flex). The fix is CORRECT and keeps value for **non-tool generic pages** (the
idea.uk reproduction), but the "tool delivery blocker" claim is false. **The
signal I walked past:** I hit *"File has been modified since read"* on
`bugs_open/024` mid-edit — an actively-owned bug file moving under me — and pressed
on with a fix built on its *earlier* framing instead of re-reading the top first.
Cheap check: when a bug is actively owned, re-read the case file before investing
in and shipping a fix premised on it. Family: stale-premise, but sourced from a
concurrent session rather than my own earlier claim.

**2026-07-21 — wrote "fixed-but-inert until the image rolls" for a fix that was
already LIVE.** Applying `bugs_open/040-partial-build` I edited `v3_site_actions.go`
and, before committing, wrote in the bug file that the Go change was "fixed-but-inert
until an image roll" — the reflexive framing for a Go fix. **It was already deployed.**
A concurrent add-all sweep had committed my file into `fe2ba5e52` ("v1.0.1146 -
sweep"), which built the image from HEAD and rolled it; the running chassis pod was
on v1.0.1146 with my guard's literals in its binary. I only found out because I
checked the pod tag AFTER writing "inert". Cheap check that would have caught it
before the claim: `kubectl get pod -l app=agent-chassis -o …image` + a discriminating
`strings /app/agent-chassis | grep "<a literal my change created>"` BEFORE writing any
inert/live status. This is not rare — the SAME sweep took `bugs_open/038` and `041`
live too, and both owning sessions filed the same correction the same hour (commits
`412c88edb`, `0ff96a972`). Family: "committed code rides ANYONE's next HEAD build" —
already logged for `bugs_open/047`; the recurrence is the point. **A Go fix is inert
until an image rolls; whether one has ALREADY rolled is a pod check, not an assumption.**

**2026-07-21 — nearly wrote "nothing releases orphaned `claimed` items → sites wedge
forever" as `bugs_open/029`'s residual. False.** Diagnosing 029 (build halt after a roll) I
built the chain: hung `build-dispatch-loop` → items stuck in `status='claimed'` →
`find_dispatchable_site`'s `NOT EXISTS(claimed)` per-site mutex → site undispatchable. Correct
so far. Then I checked the reaper, `load_work_items`, `fail_work_item`, and the DB triggers,
found none reset `claimed`→`triaged` on orchestration death, and was about to assert a
permanent wedge. **The whole self-heal lives in one scheduled task I had not opened —
`claimed-item-timeout` — which resets any claim >40 min back to `triaged`.** The wedge is
bounded, not permanent; my residual would have been the opposite of the truth. **The signal I
walked past:** I had *grepped* for claim-reset paths and got hits in the files I expected
(`claim_work_item_action.go`, `load_work_item_actions.go`) — a scheduled-task `pre_query` is a
DB row, invisible to a Go grep, so "I searched and found the release paths" was false comfort.
What caught it: the `090` trigger script's own header comment names the 40-minute reset in
prose. Cheap check that would have caught it first: `SELECT name, pre_query FROM
scheduled_tasks WHERE pre_query ILIKE '%claimed%'` — enumerate the DB-resident logic, not just
the Go. Family: not-looking at one component (here a *scheduled task*, the class Go-only greps
always miss — cf. the admin-dashboard `.tsx` memory), confidence no protection.

**2026-07-21 — nearly emitted `feature-designer` as a false positive from the new
dormant-agents detector, trusting `bugs_open/044`'s own validation list.** Building the 044
detector, my live reproduction flagged `feature-designer` as never-run — but 044 explicitly
lists it as a positive control that "correctly detects as run" (its designer half was "PROVEN
2026-07-18, run 8e837814"). A false positive is the detector's whole failure mode, so I
stopped to reconcile instead of shipping around the contradiction. **The handoff was
imprecise, not my code:** feature-designer's 3 unique steps (`check_spec_approved`,
`load_council_report`, `load_spec`) appear NOWHERE in `orchestration_states` — not as
top-level keys, not even by full-text `LIKE` over `workflow_plan::text`. Its own workflow has
never fired; the "PROVEN" run was the review council approving its plan through other
machinery (councils log `agent_type='generic'`), which is a different execution path. So the
flag is correct and 044's positive-control claim was loose. What caught it: not trusting a
prose validation list against the live data — reran the actual query. Cheap check that
settles it: `SELECT count(*) FROM orchestration_states WHERE workflow_plan::text LIKE
'%<a_unique_step>%'` — if an agent's unique steps appear nowhere even by text, its own
workflow genuinely never ran (whatever a handoff remembers as "it worked"). Family: a
confident claim in a handoff is a claim, not a measurement; "observed running" must name the
exact signal (here: a unique TOP-LEVEL workflow step in retained history — which misses
council/subtree paths, so the detector reports for triage, never asserts "N never ran").

---

## 2026-07-22 — "the contamination-allowlist fix needs writing + a fleet roll" (brochure_component_library / fundamentallyai.com) — it was ALREADY LIVE

**Claim (acted on for most of a session):** fundamentallyai's pages are blocked by
the contamination check with no per-site allowlist mechanism; therefore I must
WRITE the allowlist fix (`loadAllowedReferenceDomains` etc.), council-review it,
and roll a new chassis image to make it live.
**Actually:** the entire allowlist implementation was **already committed and
live**. `git log -S 'allowed_reference_domains' -- validate_page_content.go`
shows it was introduced by `fe2ba5e52` ("v1.0.1146 sweep"), built into the
v1.0.1146 image, and running on production since 2026-07-21 18:50 UTC. A pod-grep
of the live pod found `loadAllowedReferenceDomains` (2), `checkDomainContamination`
(2) and the `allowed_reference_domains` literal (3) — the fix was already there,
byte-identical to what I "wrote". My commit `f2780e1bd` changed only a **comment**
in that file (+ added a test + the seed script). The one thing actually missing
was the **data seed** (`content_data->'allowed_reference_domains'` for the site),
which no code roll could ever provide — and which the council's "you never flip
the switch" objection was pointing straight at.
**Caught by:** the OWNER — "double check your findings, v1.0.1146 is on
production" — which finally made me pod-grep the running binary instead of
INFERRING from build-time-vs-commit-time that the fix couldn't be live.
**The cheap checks that would have caught it, at the very start:**
`grep -rn 'allowed_reference_domains\|allowlist' platform/orchestration/actions/`
(would have shown the existing mechanism) and a pod-grep of the running chassis
for the symbol BEFORE writing a line. I checked `/bugs_open` + `/bugs_closed` for
the mechanism (found none) but never checked the CURRENT CODE or the LIVE POD for
an existing implementation — the exact "is this already built/live?" liveness
check CLAUDE.md's own DORMANT-MACHINERY / prior-art discipline exists for.
**Cost:** a diagnosis that was correct but a fix-and-council effort that was
largely redundant — ~2 council rounds + credits re-reviewing already-live code,
and I nearly cut an unnecessary fleet image roll (stopped only by the owner).
Salvage: the diagnosis, the seed script, the test, `bugs_open/056` (a real new
finding) and `features_open/010` (the council-decider improvement) are all still
valid work. But the headline lesson is blunt: **before writing OR rolling a fix,
grep the current code and pod-grep the live binary for the mechanism — "is it
already done?" comes before "how do I do it?".**
