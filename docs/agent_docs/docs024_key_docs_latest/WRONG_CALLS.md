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
| read the code before asserting a mechanism | 9 |
| **read the CALLEE's logs before diagnosing the CALLER's timeout** | **1** |
| **attach the query to a load-bearing absence claim — "checked" without the check text is a claim about diligence** | **1** |
| **read the CONTRACT a thing plugs into, not just its logic** — *the council-submission case AUTOMATED 2026-07-28: `097` now type-checks `operation`/`grounded_in`/`risks` client-side* | **3** |
| **name the LAYERS a claim spans, and touch each one** | **3** |
| wait / query again before calling an absence a failure | 9 |
| **after fixing a class of defect, grep your OWN diff for the same shape before committing — fixing the instances reviewers point at is not fixing the class** | **1** |
| **test against the ARTEFACT, never against a fixture you wrote to match your assumption about it — a fixture named after real data is not real data** | **1** |
| **read the ITEM's own row, with the columns that can carry bad news, before inferring its state from an aggregate over the queue containing it — "absence is not failure" does not license "absence is progress"** | **1** |
| **grep for the capability before asserting it does not exist** | **6** |
| **prove the artefact is current before reasoning from it** | **4** |
| measure a property before describing it | **2** |
| **record the CLOCK beside a reading, never infer it afterwards** | **2** |
| **run a census against a known-positive control before reporting the count — and for a binary classifier, sample its BOUNDARY, because every implementation agrees at the extremes** | **3** |
| **look at the real values before designing for the assumed ones** | **5** |
| **follow the value across EVERY hop before sizing a fix — a struct→map→struct conversion is a hop, and a defect that fully explains the symptom can still be one of three** | **1** |
| **read the SCHEMA before naming a column — a Go map key is not a column, and a CHECK constraint's allowed set is not guessable from the column name** | **3** |
| **enumerate the SIBLING instances before quantifying — "generic"/"fleet-wide"/"the listings all X" needs a count, in EITHER direction: a defect that generalises, or a safeguard that does** | **2** |
| **verify a control by what the USER perceives, not that the handler fired — an invisible-in-context effect is a dead control** | **1** |
| **verify the runtime that will EXECUTE the code — a deployment pod-grep is a false green for spawn-class agents, AND for the right pod running the wrong code path** | **2** |
| **bound the BEHAVIOUR, not the function — "X has exactly one caller" is true and does not answer "what else does this job without calling X"; a verified scoping claim closes the question hardest** | **1** |
| **check an example you write against the artifact it constrains** | **2** |
| **re-derive an inherited residual's prescription; a previous session's fix note is a hypothesis, not a spec** | **1** |
| **grep the index before filing — and hardest when a human is waiting, because their report went to more than one session** | **2** |
| **grade a probe on the ACTION's own output (`collected_data-><step>`), never on the run's terminal status — a harness that never delivered its payload completes GREEN** | **1** |
| **check whether an existing bug has an owning workstream before routing work to it** | **1** |
| **read before write — never `cat >` a file you did not create** | **1** |
| **re-resolve a file:line you carried across sessions — above all one you edited yourself** | **1** |
| **verify an embedded/quoted artifact is COMPLETE before asserting it — a fixed `[:N]` slice is an unmarked truncation; an author's own ellipsis in evidence is the same defect by hand** | **3** |
| **re-read the row AFTER a render, not after your own write** | **2** |
| **check the column actually means what you are measuring** | **3** |
| read the rule before inferring its purpose | 1 |
| **re-ground a figure before repeating it — one copied from a sibling doc inherits ITS measurement date, one copied out of a since-corrected tool keeps the old tool's answer, and one handed to you by a sub-agent sweep carries no measurement date at all; never let any of them land in a commit message, council submission or code comment unmeasured** | **5** |
| **a duplication audit identifies SHAPE, never INTERCHANGEABILITY — before calling two things duplicates, open BOTH and query live USAGE. A header states intent, not adoption; three of one sweep's "clear duplicates" failed this check (8 "byte-identical" health servers were 8 distinct bodies; two "duplicate" exporters shared a purpose and 0 of 16 functions; the "generic" Firecrawl action had no callers at all while the "bespoke" one was live)** | **1** |
| **prove a transform against the ENGINE that will run it, not the one you reasoned in** | **2** |
| **resolve BOTH operands to the same ground before comparing — same run, same namespace** | **4** |
| **confirm the record you are reading is the one that produced the artefact** | **5** |
| **compare a STRUCTURAL property of the output against the source — counts and a zero exit code cannot see silent loss** | **1** |
| **derive a count from the artefact; never type one** | **2** |
| **review the plan's CONTENTS, not just its top-level shape** | **1** |
| **commit with an explicit PATHSPEC — a bare directory is `git add -A` wearing different clothes in a shared tree** | **1** |
| **read the target file's own stated contract before appending to it** | **1** |
| **pair a negative assertion with a positive control over the same fetch — "the bad string is gone" also passes on a 404, a typo and an empty file; and run any pod-grep marker against the CURRENT binary first — if it passes before the change ships, it is not a test** | **2** |
| **give an "absence means wait" rule an exit condition — check whether anything NEWER has drained past you before concluding you are merely queued** | **2** |
| **never suppress stderr on a fetch whose ABSENCE becomes a finding — a transient failure then renders identically to a real empty result, and prints as a fact about the estate** | **1** |
| **ask which line of your test fails if the feature is removed — assert on the mechanism the code actually uses to signal failure, not the one you assumed** | **1** |
| **GROUP BY the variable you were about to filter on, so the excluded slice appears as a row instead of vanishing — and re-derive an inherited filter against YOUR question, because a filter copied from another document is still yours** | **1** |
| **prove the CHANNEL carries signal before reading a zero from it — an unscraped metric, an unwired counter and a fixed bug are byte-identical** | **1** |
| **write out the actual resolution/lookup order for the SPECIFIC input before tuning the knob that governs it — the obvious direction can be strictly worse** | **1** |
| **when you write a counter-argument to your own change in its risks section, that is the change failing review, not a disclosure — split it out before shipping** | **1** |
| **never read a `| head`-truncated grep as a complete enumeration — the cap is silent and looks identical to "that's all of them"** | **1** |
| **enumerate the SIBLING instances before quantifying** — *(existing row above; incremented again 2026-07-26: fixed a phantom broker at the site I tripped over and never grepped for the second one, which the council found)*
| **read the ROW's own annotation before theorising about what moved it — a mechanism that labels its own work has already answered you; and never truncate the field that might carry the answer** | **1** |
| **verify with an INDEPENDENT witness — a check that shares the fix's regex, query or assumption can only echo it, never falsify it** | **2** |
| **establish the healthy BASELINE before calling a reading abnormal — and treat a famous failure mode that fits your symptom as a hypothesis, not a diagnosis** | **2** |
| **prove it before writing it into a SHARED doc — a runbook or landmine entry asserts at higher confidence than a note, and propagates to every later reader** | **1** |
| **read the step's CONFIG, not its name — `select_sections` selects nothing, and a name-shaped inference can get the right fix candidate rejected** | **1** |
| **re-check the deployed binary AFTER committing platform code — "inert until a roll" is someone else's decision to make, and in a shared tree it expires without telling you** | **1** |
| **treat a live artefact changing under you mid-investigation as an OWNERSHIP signal — `who-owns.py` reads commits and is blind to an uncommitted session working the same ticket right now** | **1** |
| **measure a SYMPTOM's exposure across every cause, not only the one you are fixing — a bug file's exposure figure is read as "is this biting?", not "is my code path biting?"; and FETCH the live artefact you have already named in your own prose** | **1** |
| **RE-RUN the prior-art search when the design outlives it — an absence is true only at the moment you looked, and a peer built the next day is invisible to a search you already did. A search is a reading, not a property.** — *AUTOMATED 2026-07-28: `check_new_capability_surface` in `scripts/pattern-check.py` fires when a staged `.md` proposes a `cmd/`, dockerfile or package that does not exist, and prints the existing peers marked `(new)`. Advisory. Measured 1.33% / 1,500 commits. It re-runs on every later commit touching the doc, which is the half a one-shot review cannot do.* | **1** |
| **read `decided_by` before writing a `Council-Reviewed:` trailer — and again if the submission went to another round, because a later APPROVAL can attach to a materially DIFFERENT plan and the coverage report cannot tell** | **1** |
| **read `features_open/` before treating a capability question as open — the backlog holds DESIGNED-NOT-BUILT answers, and a feature file is invisible to a code grep, a component query and a bug-dir grep alike** | **1** |
| **measure a RENDERED property in the thing that renders it — a stylesheet cannot say what colour is painted (it cannot resolve the cascade, ancestors, alpha or gradients), and an audit that over-reports on live sites is worse than none, because its findings get "fixed" into real regressions** | **1** |
| **treat any check whose failure triggers an automated REWRITE as production configuration — a wrong test does not merely fail, it becomes the specification the repair loop faithfully implements, and it will damage correct code to satisfy it** | **1** |

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

**Update 2026-07-25 (bugs_open/021 closure).** Four increments from ONE session,
no new rows — and the lack of a new row is the finding. *Wait/query before calling
an absence a failure* (6→7), *grep for the capability before asserting it does not
exist* (3→4), *prove the artefact is current* (3→4), and *verify a quoted artifact
is complete* (1→2). Every one is the same underlying act: **a confident inference
from evidence that was present but incomplete.** Read in sequence they show it
escalating in blast radius — first I misread an absence (cost: 20 minutes), then I
recommended building something already built (cost: a wrong runbook entry), then I
handed an elided SQL quote to twelve reviewers and one of them raised a MEDIUM
objection against correct code (cost: a manufactured defect on the permanent
record). Same error, three audiences, widening consequences.

Two things worth noting for the distribution. First, *wait/query before calling an
absence a failure* is now the second-largest row at 7 and it took **two entries on
the same day from two threads who could not warn each other** — the strongest
signal yet that this one should stop being a discipline and start being tooling;
`scripts/dispatch-queue-depth.sh` (built that day by the 030 thread) is exactly
that, and the remaining gap is that `091`/`092` do not call it. Second, the
quoted-artifact row generalises in an uncomfortable direction: it was filed about
*code* truncating a quote, and its second instance is an *author* abbreviating one
by hand. The check does not care which; a reviewer cannot tell the difference.

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

---

**2026-07-21 — assumed the council fix_plan `operation` vocabulary instead of reading it (bugfix-026).**
Submitted a council-gate plan with `operation: "create"` for two new files (a git-verb
guess). The run completed at `complete_invalid` with a persist-time reject — *"operation
\"create\" not in the allowlist"* — before any reviewer ran. The allowlist is
`modify | add | remove | config_change` (`diagnose_persist_fix_plan_action.go:80`); a new
file is `add`. **The cheap check:** the vocabulary is a 1-line constant in the persist
action — read it (or copy an existing valid submission) before submitting, rather than
guessing from git. **Cost:** one wasted council dispatch (no credits — rejected pre-review;
just the ~queue latency and a resubmit). Low, but it is the same "assumed the contract
instead of reading it" shape this whole bug (026) is about — a small irony worth the row.
The 097 trigger validates plan SHAPE client-side but not the operation enum; validating the
enum there too would move this from runtime-invalid to a client-side error (candidate for the
council-gate owner).

---

**2026-07-21 — a council submission that said an `add` edit was "confirmed present in
production" read as the false-already-deployed anti-pattern (bugfix-053).** Round-1 of the
053 council submission described the new `siteHasAnyNavItems` helper as `operation: "add"`
in the plan, while the risks section noted "confirmed present in production binary v1.0.1146
(pod-grep)". Both were true — this thread commits per task and the fleet builds from HEAD, so
the new function had been committed and shipped *before* the (advisory, post-ship) review — but
`prior_art_librarian` correctly flagged the pair as **self-contradictory at HIGH severity**: a
brand-new `add` cannot already be live, which is exactly the `bugs_closed/031` false-already-
deployed shape it exists to catch. **The cheap check:** before submitting, read the risks/evidence
section against the edit `operation`s for internal consistency — an "add" carrying an "already in
production" claim is a contradiction on its face; if you commit-then-review (which the repo's rules
encourage), say so explicitly up front ("this edit is already committed and shipped in <tag>; the
review is advisory and post-ship") rather than pairing "add" with a bare liveness claim. **Cost:**
one of the objections in a REVISE round (round 1 also found a *real* gap, so the round was not
wasted); resolved by explanation in round 2, which the seat accepted. Low, but it is the reviewer
reading the submission literally — the same "the sketch is all they see" discipline that cost the
round-2 and round-3 nitpicks (`containsGroupType` not shown as pre-existing; blast radius asserted
not independently confirmed).

---

**2026-07-22 — "neutralise the predicate in place to prove the tests discriminate" on a
hotly-contended file, and the edit silently vanished (bugfix-037).** To prove the 037 tests
were discriminating I edited `realisedPageCompositionIsPreserved` in `v3_site_actions.go`
in place (drop the `needs_rebuild` case), ran the tests — and they *passed*, which was
impossible. Cause: the whole `actions` package was a large uncommitted multi-session WIP, and
an owner `git add -A` build sweep (`fe2ba5e52`, v1.0.1146) committed/rewrote the file
**underneath the running experiment**, restoring the non-neutralised version before the test
binary built. I briefly misread the green as "my tests aren't discriminating" rather than "my
edit didn't take". **The cheap check:** never run a throwaway in-place edit (neutralise,
bisect, spike) on a file under active concurrent commit — do it in an isolated `git worktree`
(`git worktree add --detach <dir> HEAD`, overlay your uncommitted files, run there, remove).
That is what finally proved discrimination cleanly. Corollary confirmed the same day: the file
could not be committed independently anyway (it referenced helpers from another session's
uncommitted `component_validation.go`), so the fix reached prod only *because* the sweep took
it — a same-file passenger in the other direction. **Cost:** two confused verification cycles;
no wrong claim shipped (the worktree caught it before I recorded anything).

---

**2026-07-21 — bare `git stash pop` in the shared tree popped another session's stash (bugfix-026).**
To check whether a RED `discovery_checks` test was pre-existing, I ran `git stash push -- pathA
pathB` (pathB untracked → errored, stashed nothing useful), then a bare `git stash pop`. In this
many-session working tree `stash@{0}` was **another branch's** WIP (`066_hitl_questionnaire`,
carrying `coordinator.go` + awaited_requests changes), not mine. `pop` with no id pops
`stash@{0}` regardless of owner — it would have merged that session's half-finished code into my
tree. It failed harmlessly only because `coordinator.go` had conflicting local edits. **The cheap
checks skipped:** (1) `git stash list` before any `pop` — the top entry is shared state, not
yours; (2) don't use `git stash` at all to test "is this RED mine?" — instead reason whether your
change *can* reach the failing assertion (here it can't: the failing item types are in files I
never touched), or overlay your file onto `git show HEAD:<path>`. Cost: none (the pop failed),
but a hair from importing another thread's uncommitted work under my branch. `git stash` is
process-global; the multi-session CLAUDE.md rules about pathspec commits apply to it doubly.

---

**2026-07-22 — "round 4 is wording, not new evidence" (idea.uk chrome council, corr 7152c7cf).**
The RESUME handoff, written between round 3's submission and its verdict, asserted that if round 3
came back REVISE "the objections are wording now, not evidence — the submission JSON has every
measurement already attached, so a 4th round is wording." Round 3 DID come back REVISE, and the
prediction was wrong on all three live objections: (1) the missing-field detector used a
control-flow-blind **regex** that false-flags `{{.x}}` nested in `{{range}}`/`{{if}}` — up to ~30
active components would log false Errors; (2) `bug_historian` asked whether a **sibling** silent-drop
path existed, and one did (`RenderTemplateWithMap`), un-audited; (3) `RenderTemplate`'s caller set was
asserted small, but is 8 sites across 5 pipelines. None was wording; two needed a real code redesign
(regex → `text/template/parse` walk) and the third a genuine grep. **What caught it:** running the
objections' own read-only checks against the code BEFORE assuming they were wording —
`git grep RenderTemplate` (caller set), `git grep "no value"/missingkey` (sibling audit), and reading
`missingBareFields` to see the regex really was blind. **The cheap check that would have:** the same
three greps, ~2 minutes, done before writing "it's just wording" into a handoff. This is the exact
failure the debugging-guide's diagnosis section names — *confidence is not a signal; the wrong claim
felt obvious.* A handoff prediction stated in the same confident voice as a finding IS the error.

### 2026-07-22 — robot-hands — "gripper-detail's re-render never delivered — the deployed file is from May" (a delivery gap that wasn't)
**Asserted:** curled `https://robot-hands.com/gripper-detail.html`, saw an empty
stat skeleton whose gqls/sites file dated to 2026-05-02, and began writing it up as
a bug-024 delivery gap — a re-render marked complete/deployed that never shipped.
**Actually:** that root file is a stale unrelated artefact. The page's real path is
`/entities/gripper-detail.html` (`pages.url`), where R7's values (10/6/4/39) were
live all along. Nothing failed to deliver.
**Caught by:** reading `pages.url` — after the DB already told me: the component's
`rendered_html` was populated and `build_status=deployed`, which contradicts "never
delivered" and should have said "you're looking at the wrong file".
**The cheap check that would have caught it:** `SELECT url FROM pages WHERE name=…`
before curling. The page NAME is not the deployed PATH.
**Cost:** ~5 min and a near-miss bug write-up; no wrong claim reached a handoff.

### 2026-07-22 — bugfix 020 — "field paths verified" (they were; the extractor re-resolved them)
**Asserted:** to the review council (rounds 3–4) that `check_tool_fabrication`'s
`collected_data` field paths were "verified against the real workflow shape" —
`completeness_check.clean_html` is exactly what `validate_tool` reads, etc. That
much was true. The unstated, unchecked assumption was that the ACTION would *use*
those config values as literal paths.
**Actually:** the wrapper read them via `datahelpers.ExtractActionInputs`, whose
**Strategy 0** resolves any dotted config VALUE against `collected_data` before the
handler runs. So `html_field="completeness_check.clean_html"` was turned into the
1412-char HTML *content*, and the handler then extracted *again* with that content
as a path → `""` → the detector saw empty input → returned `uninspectable`. On the
fail-OPEN detector (v1.0.1146) this is a silent no-op — the gate inspects nothing
and deploys everything, including real fabrications. On the fail-SAFE detector
(v1.0.1149) it over-HOLDS every recreation. I had wired the gate live before finding
this.
**Caught by:** the induced-fault probe (a scratch agent running the live gate on a
stubbed fabrication). It reached the HELD terminal but via `tier:"uninspectable"`,
not the expected `tier:"declaration"` — so the fabrication content never reached the
detector. Static verification (graph correct, paths correct, detection unit-proven)
said "green"; only inducing the fault exposed it. **The fail-safe change (council R4)
is what made the bug loud enough to catch** — the old fail-open would have looked
identical to a clean pass.
**The cheap check that would have caught it:** read how the action *consumes* the
config field, not just whether the path resolves — `check_tool_completeness` reads
`config["html_field"]` directly for exactly this reason; I used `ExtractActionInputs`
and never traced what it does to a dotted value. A one-off `go test` of the WRAPPER
(not just the pure detector) with a dotted config path would have failed instantly.
**Cost:** wired a broken gate live (~1h, zero real recreations hit it, unwired on
discovery); one extra image roll now needed before 020 can close. The lesson is the
one the debugging guide keeps making: static "verified" ≠ correctness — induce the
fault for anything whose job is to detect one, and test the layer that actually runs.

---

## 2026-07-22 — routed a work item to a handler without checking the handler handles it (bugs_open/054, bugfix o54 session)

**The claim:** filing the new `chrome_dead_control` finding at `status='detected'` with
`handler_agent='nav-link-fixer'` (the `phantom_internal_links` site_component routing convention)
would DRAIN it — "a re-render fixes the data-lag case, a genuinely-unresolvable field exhausts
max_attempts → failed and surfaces to the human queue." I grounded the routing in
`check_phantom_internal_links.go`'s `routeBySurface` table and asserted the drain worked.

**Why it was false:** `routeBySurface` is only the DISPATCH table — it says which agent gets the
item, not that the agent's workflow handles this item's shape. `nav-link-fixer`'s live workflow
(`ensure_site_record → fix_nav_templates → rerender_site_components → complete_workflow`) marks the
item **complete unconditionally** — it never verifies the dead field resolved. So a genuinely-
unresolvable chrome control would be re-dropped on the re-render and marked *done*, on every render,
and **never reach a human** — the exact silent-loss shape the whole bug exists to close.

**Caught by:** the council gate (guardian + bug_historian, MEDIUM), on the first submission. Its
recommendation ("confirm the handler discriminates on the new item_type / actually fixes it") was
correct; owner then confirmed the reroute to `needs_human_review`, no handler.

**The cheap check that would have caught it:** before routing an item to a handler, READ THE
HANDLER'S WORKFLOW (`agent_definitions.default_config.workflow`), not just the dispatch routing
table. One query would have shown `complete_workflow` has no verify step. The routing table answers
"who gets it", never "does it get fixed". Same shape as the 020 wrong call above (trace how the
consumer USES the thing, not just that the wiring resolves).

**Cost:** one council round (REVISE→revise→APPROVED) + one owner question. Cheap — the gate is exactly
where being wrong is supposed to cost the least. A REFUTED-by-council routing is a success, not a waste.

**2026-07-22 — treated a transient `orchestration_states` retention as a stable policy
and reasoned from it.** Building the 044 dormant-agents detector on 07-21, I measured the
run-history window as `now - min(orchestration_states.created_at)` = ~55 days (106k rows back
to 05-28) and wrote that into NOTES as "the retention window", reasoning that legacy agents
show as never-run because they predate it. **The 55 days was a transient, not a policy.** By
07-22 the table was 1,737 rows, oldest 9 days, 94% in the last 36h — the hourly
`database-cleanup` task deletes COMPLETED/FAILED orchestrations after **24h**, and it simply
had not been pruning on 07-20. On the real 24h window "never observed" over-flags badly:
`fix-proposer` (runs constantly) is flagged because its runs were pruned. The detector shipped
`dry_run=true` so nothing false was emitted, and I added a window guard + filed the substrate
gap (`bugs_open/060`). **Cheap check I skipped:** never treat `min(created_at)` as "the
window" — read the retention policy itself: `SELECT pre_query FROM scheduled_tasks WHERE
name='database-cleanup'` (or grep for the DELETE). A table's current age-span is an
observation; the retention rule is the fact. Family: ground a figure against the mechanism
that produces it, not against a snapshot of its output (cf. "ground every figure against the
live system" — here the *figure was live and still wrong*, because it was volatile).

### 2026-07-22 — robot-hands — "clearing the stat suffix in content_data fixes the '10%' render" (R7b)
**Asserted:** R7b set `stat*_suffix=""` in `page_components.content_data` and queued a
re-render, reporting the placeholder-suffix bug fixed.
**Actually:** the suffix reverted to `%/ms/+/x` within one render. The field is
`source=static` with a junk `fallback` in the shared component's `input_schema`; an
empty static field resolves to its fallback and persists it. My *value* and
*description* edits (source=llm, no fallback) held — only the suffix, which needed a
schema-level fix, did not.
**Caught by:** verifying the live page after the re-render, then reading the
`input_schema` field `source`/`fallback` — not by anything up front.
**The cheap check that would have caught it:** before editing a field in
`content_data`, check whether it is derived on render — `render_mode` and the field's
`source` in `input_schema`. A `source=static`+`fallback` or an agent-rendered field
will not hold a content_data edit. This is the "DERIVED on render" landmine already in
memory; I did not apply it.
**Cost:** one wasted re-render cycle (R7b) before R7c fixed the root; no wrong claim
reached a handoff (the correction is in NOTES Turn 12 before this closed).

### 2026-07-22 — vonc gauntlet — "the tools ship dead `href="#"` CTAs" framed as a fleet-wide pattern
**Asserted (heading toward a durable claim):** on finding the gauntlet's two hero
CTAs were `href="#"`, and primed by the request to make the fix "generic to any new
site", I framed this as a blanket "every tool ships dead CTAs" pattern before checking
any other tool.
**Actually:** the two sibling tools on the SAME site — `arena` and
`archetype-taster-quiz` — have **zero** `href="#"`. The gauntlet was the anomaly. The
*generic* defect was real but lived elsewhere (the detector's `build_status` filter),
not in a shared "tools emit dead CTAs" behaviour.
**Caught by:** a two-line sibling curl (`for t in arena archetype-taster-quiz; do curl
… | grep -c 'href="#"'`) — run BEFORE the claim reached a handoff. Cost ≈ nil; the
correction is in NOTES and the framing never shipped.
**The cheap check that would have caught it:** before calling a single-page defect
"generic" or "fleet-wide", curl 2–3 sibling instances and count. "Make it generic" is a
request about the FIX's blast radius, not evidence that the SYMPTOM is already
widespread — those are different questions and I conflated them. Family: measure the
population before describing it.

### 2026-07-22 — vonc gauntlet — four SQL queries written against columns that don't exist
**Asserted (implicitly, by querying):** `pages.page_name`; `content_components.component_type`;
`schema_migrations.version`/`id`; `orchestration_states.name`/`agent_type`.
**Actually:** the columns are `pages.name`, no `component_type` at all,
`schema_migrations.filename`, `orchestration_states.orchestration_name`/`owner_agent_type`.
Each threw and cost a retry (four in one session).
**Caught by:** Postgres, immediately, every time — so no wrong claim propagated, only
wasted round-trips.
**The cheap check that would have caught it:** `\d <table>` before writing SQL — the
standing CLAUDE.md rule ("Schema first"). I guessed column names from what felt natural
(`page_name`, `version`) instead of reading the schema. This is the same skipped check
as the existing "read the SCHEMA before naming a column" row; four instances in one
session is exactly the kind of repetition the tally exists to surface.

### 2026-07-22 — vonc gauntlet — "made it genuinely work / genuinely functional and honest, live"
**Asserted:** after removing the dead `href="#"` CTAs and wiring the buttons via JS, I
wrote in NOTES + SUMMARY + to the owner that the gauntlet was "genuinely functional and
honest, live" and "make it genuinely work" was done.
**Actually:** the tool is a hollow shell. "Enter the Gauntlet" fires `startTimer()` +
`scrollIntoView` on a panel already on screen + `focus()` on a checkbox; "Preview Rules"
scrolls to a rules card already visible. The objective checkboxes tick a progress bar
wired to nothing. Every handler FIRES — but produces nothing a user can perceive, and
nothing that constitutes a working product. The owner tried it: "I can check off the
challenges but I don't know why … nothing happens when I click enter the gauntlet or
preview rules."
**Caught by:** the owner using it. My own verification checked that the JS bound the
button (`data-gi-enter-btn` present in the served asset) and that the HTML was correct —
i.e. I verified the MECHANISM fired, never the EXPERIENCE. A curl/grep can't perceive
"nothing visibly happens."
**The cheap check that would have caught it:** verify an interactive control by what the
USER perceives, not that the handler ran. For a button, ask "what changes on screen that
a person would notice, and is that change meaningful?" — if the effect targets a region
already in view, or updates state nothing consumes, it is a dead control wearing a live
handler. This is the same "trust the rendered artefact / lived behaviour, not the status"
invariant, pushed one level past DELIVERY into EXPERIENCE — which is precisely the gap
the experience-loop workstream exists to close, and I should have run its lens (does the
journey deliver what the button promises?) before calling it done.
**Cost:** shipped a cosmetic fix as "working", the owner had to catch it, and I had
declared victory in a SUMMARY. Corrected in NOTES + README; the real fix (backend + AI
competitor via the experience loop) is now scoped.

---

## 2026-07-23 — "no `%med%` table exists live" (a schema-filtered query read as fleet truth)

**Asserted:** in the `bugs_closed/054` close-out (2026-07-22, commits `cde8b4da0`,
`5deea39ea` message): "No med price data source exists live — zero tables matching
`%med%` in clients_db; the exporter's loadMedPricesForExport has nothing to read." Also
a derived claim in session memory: "the Phase-1 product_prices/kind unified schema is
NOT in clients_db."
**Actually:** both false. The query filtered `table_schema='public'`. The med arm lives
in `business_intel.*` — med_retailers (3 active), med_retailer_listings (304 with
per-listing retailer_url), med_price_snapshots (2,587), med_scrape_evidence (2,157),
matview med_price_current — alongside the unified `business_intel.products`/
`product_prices`. The exporter had plenty to read; it was merely stale (>14-day window).
**Caught by:** the owner saying "look at the latest docs — things have changed" while
scoping the med revival; the vetcomparison RUNBOOK's `business_intel.claim_requests`
reference prompted `\dn`, which showed 8 schemas.
**The cheap check that would have caught it:** `\dn` FIRST, or query
`information_schema.tables` with **no schema filter**, before asserting any "table does
not exist" claim. A negative existence claim inherits the blind spots of its search — the
same failure shape as grepping one directory and declaring a symbol absent.
**Cost:** one false durable claim in a closed bug file + a poisoned memory entry; both
corrected 2026-07-23 (bugs_closed/054 CORRECTED block). Caught before any thread built on
it — but only because the owner pushed back.

---

## 2026-07-24 — "the discovery run is stuck — this is the bug-003 spawn-loss shape" (model_directory_pipeline, run 1)

**Asserted:** in the workstream NOTES and the session read-out (2026-07-22): the first
`model-directory-discovery` orchestration was "stuck, not merely slow" at `search_web`,
and "the shape matches bug-003 spawn-loss more than it points at new code."
**Actually:** neither. The run was mid-retry when observed, and resolved itself to a clean
`FAILED — Request timed out after 3 retries` ~12 minutes in: the coordinator's own
await-timeout machinery working exactly as designed. The real cause (found two runs
later) was in the adapter's REPLY path — a different mechanism, a different component.
**Caught by:** re-reading the orchestration row the next day instead of building on the
observation.
**The cheap check that would have caught it:** read the row's `status`/`error` to a
TERMINAL state before classifying it — "still running when I looked" is an observation
about my look, not about the run. And naming a known bug family from symptom shape alone
is asserting a mechanism without reading anything.
**Cost:** one wrong durable line in NOTES (corrected in place), plus it pre-framed the
next day's debugging toward spawn-loss when the defect was in the adapter.

---

## 2026-07-24 — "the batch is too slow for the await — raise the timeout, shrink the batch" (runs 2–3)

**Asserted:** (NOTES + a live config change, 2026-07-24 morning): the scrape batch
plausibly exceeds the 120s await every time — fix by raising `scrape_pages.timeout_seconds`
120→300 and `max_scrapes` 4→3.
**Actually:** the scrape took **4.69 seconds**. The reply was refused by the Kafka broker
(`Message Size Too Large`) and dropped silently by the adapter — timing was never the
mechanism. Worse, the "raise" edited a STEP-LEVEL `timeout_seconds`, a field
`models.Step` does not have: it is silently dropped at unmarshal, and only
`config.timeout_seconds` is read (`ConvertStepTimeout`). Every await had been running at
the 180s default all along, including under the seed's original 120.
**Caught by:** finally reading the ADAPTER's logs inside the failure window (the two
prior windows were destroyed by fleet pod rolls before I looked); the inert field by the
run-3 row's `workflow_plan` carrying my `max_scrapes` change but no timeout at all.
**The cheap check that would have caught it:** (1) read the CALLEE's logs before
diagnosing the CALLER's timeout — two theories were built and one config change shipped
entirely from caller-side evidence; (2) before "fixing" a config value, grep for its
READER — a field nothing consumes accepts any value you like.
**Cost:** one wasted run, one inert config change presented as a fix, and the same
inert field sits in the evidence-researcher seed it was copied from (flagged in 062).

---

## 2026-07-24 — "no orchestration appeared — consistent with dispatch-queue latency" (run 3 watcher)

**Asserted:** (session read-out): the run-3 watcher timed out with "no orchestration
appeared", which I attributed to the known ~30-min dispatch-queue latency.
**Actually:** the orchestration was created at 10:01:46 — dispatch was instant. My
watcher's filter (`initial_request_data->>'research_query'`) didn't match the row's
actual shape, so the query returned nothing and I read my filter's blind spot as the
fleet's queue.
**Caught by:** a broader query minutes later (all orchestrations since 10:00).
**The cheap check that would have caught it:** when a filtered query returns nothing,
WIDEN THE FILTER before asserting the event didn't happen — a negative claim inherits
the blind spots of its search (same shape as the 2026-07-23 `%med%` entry, one day
apart).
**Cost:** minutes only, but only because the wider query was cheap; the same habit on a
slower loop is how a healthy system gets diagnosed as backed-up.

---

## 2026-07-24 — "run 6 fired against the genuinely-new chassis binary" (the watcher/rollout race)

**Asserted:** (session read-out): run 6 was fired "after the 300s post-restart quiet
window", i.e. against the freshly-rolled v1.0.1154 chassis carrying the pipe-folding
fix — so when it rejected every claim again, the natural reading was "the fix doesn't
work".
**Actually:** the v1.0.1154 ReplicaSet was created at 11:23:52; run 6 was created at
11:16:16, fired by my background watcher — launched before the rollout had actually
landed, sleeping a fixed 360s — against the OLD pod. The fix was never exercised.
**Caught by:** testing the fix locally FIRST (the exact rejected quote through the real
Go matcher against the live page: matches) and only then comparing the ReplicaSet
creation time against the run's `created_at`.
**The cheap check that would have caught it:** pin WHICH binary processed a run (pod
start time vs run created_at) before crediting or blaming a code change — a fixed-delay
watcher launched "after" a rollout command races the rollout itself; gate on the new
pod's start time, not wall-clock patience.
**Cost:** one wasted run and twenty minutes of doubting a correct fix. The local-test
habit is what kept it from becoming a wrong "the fix failed" entry in NOTES.

---

## 2026-07-24 — "checked: no live workflow reads raw_html from a batch output" (an absence claim without its lookup)

**Asserted:** (council submission for the 062 fix): "checked: no live workflow config
references raw_html/html_content from a batch_webscrape output", and "directory-researcher
is the ONLY user of batch_webscrape" — both load-bearing for shipping a lean-by-default
reply as a safe change, both stated as prose with no query attached.
**Actually:** both turned out TRUE — but two council seats (guardian, prior_art_librarian)
correctly refused to inherit them, naming the exact shape: ASSERTED ABSENCE standing in
for an existence check. The exhaustive `agent_definitions` scan took ~2 minutes when
actually run, found 3 batch_webscrape consumers and 6 false-positive `raw_html` hits
(all self-referential), and went into the case file.
**Caught by:** the council round — i.e. by process, after submission, rather than by me
before it.
**The cheap check that would have caught it:** attach the QUERY to any load-bearing
absence claim at the moment of writing it — "checked" without the check text is a claim
about my diligence, not about the system. If it's cheap enough to run when challenged,
it was cheap enough to attach unprompted.
**Cost:** none this time (the claims held) — which is exactly why it's worth logging:
the identical habit with a false absence ships a breaking change on my say-so.

### 2026-07-24 — gauntlet/B4 — "the implementer fire was dropped" called at 120 seconds, twice
**Asserted (operationally):** my fire-with-retry script declared an implementer
dispatch "NOT ingested — refiring" after a 120s window, twice, and an earlier fire
"never ingested" after ~4 minutes of polling.
**Actually:** all three fires ingested — LATE (one ~9 minutes). The refires piled
three concurrent implementers onto one correlation; they raced branch creation and
all mutually died on the E4 pre-existing-branch guard. Cost: three wasted runs and a
cleanup round.
**Caught by:** the post-mortem query listing three implementer orchestrations created
within 3 minutes of each other, all `complete_refused` on E4.
**The cheap check that would have caught it:** my OWN memory already says it —
"no orchestration row means QUEUED (~16+ min under backlog), not dropped … do not
retry on that evidence (it costs a duplicate round)." I wrote an automatic retrier
whose timeout contradicted a recorded lesson. Ingest-latency under load is MINUTES;
any auto-refire must wait longer than the worst observed latency (≥10 min here) or
not exist. Family: wait / query again before calling an absence a failure.

### 2026-07-24 — robot-hands/043 — superseded a site_specs row I had not read (caught same minute)
**Asserted:** (by action, not in prose) an unconditional "supersede any current
evidence_base row" UPDATE across four sites, assuming none existed — three had
none, but vonc carried migration 166's structured evidence_base whose
banned_claims regexes feed the experience-loop claims checkers. My
writer_block-only replacement silently removed them from "current".
**Actually:** another thread's live machinery was keyed on the row I clobbered.
**Caught by:** the apply output — `UPDATE 1` where I expected `UPDATE 0`. Read
the superseded row, found the structure, fixed with a MERGE row carrying both.
**The cheap check that would have caught it:** SELECT the current rows BEFORE
writing a supersede — the same "read before write on any file you did not
create" rule, applied to DB config rows. One query.
**Cost:** ~10 minutes of checker-pattern absence on one site; nothing reached a
handoff; the merge row is strictly better than either original (166 had no
writer_block, which is why the WRITER kept inventing on vonc despite checkers).

### 2026-07-24 — gauntlet/B4 — "the formatter fix is live — pod-verified" (it was, on the wrong pod)
**Asserted:** after rolling chassis v1.0.1155 I reported the fix "POD-VERIFIED live"
(discriminating symbol present, positive control present) and re-fired the implementer.
**Actually:** the implementer runs as a SPAWNED dedicated pod whose image comes from
`agent_definitions.image_tag` — pinned `v1.0.1151`. Round 6 failed with the exact error
the fix removes, on a binary four tags old. The deployment pod-grep proved the image
exists, not that the agent would run it. Fleet census after: 168 active chassis-image
agent rows pinned to v1.0.1151 → filed `bugs_open/066`.
**Caught by:** round 6's refusal reproducing the "impossible" error.
**The cheap check that would have caught it:** verify the runtime that will EXECUTE the
code path — for spawn-class agents, `SELECT image_tag FROM agent_definitions WHERE
type='<agent>'` (and after firing, the spawned pod's `.spec.containers[0].image`),
never only the deployment pod. "Pod-verify" must name WHICH pod and why that pod is the
one that runs the code.
**Cost:** one wasted implementer round + a false "live" report in NOTES (corrected).

### 2026-07-24 — gauntlet/B4 — my own corrective rule's example seeded the next failure
**Asserted (as an instruction):** seed 199's module-path rule gave the example import
`github.com/gqls/agentchassis/internal/tools-api/config` — implicitly asserting that a
config PACKAGE DIRECTORY was the right layout.
**Actually:** the approved plan's file is `internal/tools-api/config.go` (no config
package). The model followed my example, relocated the file to
`internal/tools-api/config/config.go`, and the deterministic allowlist rightly refused
the whole stage (round 4).
**Caught by:** the allowlist violation naming both paths side by side.
**The cheap check that would have caught it:** check an example you WRITE against the
artifact it constrains — my rule was authored next to the approved plan and I never
cross-read the two; one glance at the plan's file list would have caught the
contradiction. An instruction's example is a claim about the target artifact and needs
the same verification as any claim.
**Cost:** one implementer round + migration 200 to say precisely what 199 should have.

### 2026-07-24 — per_site_ai — a capabilities inventory carried an unverified [LIVE] tag for two strategy rounds
**Asserted:** `CAPABILITIES_framework_inventory_2026-07-21.md` §3: "Hard rule:
data charts are code-rendered from real series (go-echarts)" — presented under
a [LIVE] section heading, implying a working chart renderer.
**Actually:** go-echarts appears nowhere in go.mod/go.sum/*.go; no chart action
exists; the concept register's own entry (data-charts.md) says "aspirational —
not started". The doctrine was real; the implementation was not.
**Caught by:** a plan-mode Explore agent grepping go.mod while scoping the
gripper-dossier pilot — three days and two external-LLM strategy rounds after
the claim was written (Gemini echoed it back as a platform strength both times).
**The cheap check that would have caught it:** for any capability tagged
[LIVE] in a compiled inventory, one grep of go.mod / the action registry for
the named artifact. A compiled summary inherits none of its sources'
verification — every [LIVE] tag is a fresh claim needing its own check.
**Cost:** none realised (caught before the pilot's chart step was designed
around it), but the pilot design now has to solve chart rendering it believed
was free.

### 2026-07-24 — bugfix 063 — council sketches authored as proposals for a change already committed
**Asserted:** (implicitly, by form) the R1 council submission's edit sketches
presented the 063 email fail-closed fix as work to be authored, while the
rationale said "Fix committed as fb3d5f5ea" — two claims about the same diff
in mutual contradiction.
**Actually:** the fix was committed before submission (deliberately, per the
strengthened-advisory norm). The sketches should have been the landed diff,
labelled as such; reviewers cannot resolve past-tense-rationale vs
proposal-tense-sketch without reading the repo, which is not their seat.
**Caught by:** prior_art_librarian's high-severity DORMANT-MACHINERY objection
— the gating objection of an otherwise 6-approve round (corr `7080124b` R1).
**The cheap check that would have caught it:** the relojistas thread had
ALREADY logged this exact lesson the same week ("sketches must be FINAL-state",
after 3×REVISE on the emitter fix). Reading the sibling thread's council
lessons before authoring a submission — or one pass over the submission asking
"does every edit describe the repo's current state?" — costs a minute.
**Cost:** one council round (~35 min queue + run) + the R2 authoring.

**2026-07-24 — called a section drop "UNKNOWN" when the record had simply been
pruned.** `bugs_open/040-partial-build` carried, in my own CORRECTED block of
2026-07-20, the claim that dartsonline's dropped `testimonials` was **not** a
`skip_section` deferral because *"it is not in `sections_deferred` or
`sections_skipped`"*. That check was worthless and I did not notice: those lists
live **only** in the orchestration's `collected_data`, which `database-cleanup`
prunes at ~24h, so by the time I looked the record of a *correctly working*
mechanism had evaporated. I read absence-of-record as absence-of-mechanism and
wrote UNKNOWN into a handoff, where it sat for four days as an open mystery —
and `bugs_closed/041` had already stated the true answer in its "What is NOT this
bug" section, which I had read. The mechanism was confirmed on 2026-07-24 by
grounding it in state that does NOT expire: the component's `input_schema`
(`required:true, min_items:1, on_missing:skip_section`) against the site's
current `site_specs` aspects (no `social_proof`). **Cheap check: before treating
"not in the log/record" as evidence, ask what that record's RETENTION is and
compare it to your investigation lag.** Same 24h-pruning trap that made
`bugs_open/044`'s dormant-agent sweep over-flag `fix-proposer` — the recurrence
is the point. Family: absence-is-not-evidence, retention variant.

**2026-07-25 — built and refuted two theories about a parked work item while the
row's own `summary` column named the mechanism in plain text.** Re-queued page
rebuilds on fundamentallyai.com kept flipping to `unresolved`. I read
`reconcile_superseded_reviews_action.go` (predicate is `needs_human_review`, mine
were `triaged` — refuted), then the two-strike rule in
`load_work_item_actions.go:1041` (counts terminal rows per `item_key`; these had
none — refuted), then wrote up "single-flight per site" as the surviving working
theory and told the owner so. The actual answer was
`[stale: triaged 48h+] [stale: triaged 48h+] Build index page (not_built)` —
**stored in the row I had been SELECTing from all along**, naming both the
mechanism and its rule, twice. One `grep -rn "stale: triaged"` reached the writer
in seconds: a `scheduled_tasks.pre_query` that keys on `created_at` instead of
time-in-state (`bugs_open/070`), which is why no amount of Go-reading would ever
have found it. **Cheap check: read every column of the row whose fate you are
explaining — `summary`, `status`, `attempt_count`, `claimed_at` — before reasoning
about which code moved it. A mechanism that annotates its own work has already
answered the question.** Two corollaries I got wrong for free: my `SELECT`s had
been truncating `summary` with `left(...,50)` for brevity, cutting off the very
prefix that mattered; and I never asked whether the mover was code at all —
platform logic lives in DB config too, so "not in the Go" is not "not in the
system". Cost: ~2 hours of the owner's rebuild sequence run by hand, one
incorrect mechanism stated to the owner (corrected same session), and a NOTES
entry that had to be rewritten. Family: annotation-names-the-mechanism,
truncated-your-own-evidence, config-not-code.

**2026-07-25 — reported "all three surfaces live" from a PARENT item's status
while the child that builds one of them had failed three times.** The
model-directory arc's read-out to the owner named three delivered surfaces on
ai-agent-orchestration.com: the listing page, the published JSON, and a
**homepage snippet**. The first two were real. The third never existed: the
gap-planner's child item (`content_rewrite`, "Add content to index",
`add_sections:["model-directory"]`) had gone `failed` at 16:28:31Z the previous
afternoon — `attempt_count` 3/3, `error = "Claim timed out — handler pod likely
died"` — killed by the very chassis roll I was shipping. What I actually
observed was the **parent** `missing_model_directory_section` item going
`complete`, which it does when the gap PLAN is applied, not when the section is
built. **Cheap check: `curl -s <url> | grep -c <component-slug>` — one second,
and it returns 0.** I had already fetched the page's sibling surfaces by curl
that same session; I just never fetched the one whose status looked green.
Family: status-is-not-artefact (`016b` says this in as many words), plus a new
variant worth naming — **parent-complete ≠ child-shipped**: in a planner→handler
chain the parent's terminal status reports only that work was *emitted*. Second
call wrong the same morning, same root habit: I wrote that the open-weight
research run "produced no distinct entities" and reasoned about why, from a
registry snapshot taken while its claims were still landing — the live count was
27 entities across seven owners including all four open-weight vendors. Cost:
one wrong sentence in a delivery report to the owner, one speculative
explanation of a non-existent blind spot committed to NOTES, and a day in which
the flagship section the owner asked to be *prominent* was absent from the
homepage. Family: status-is-not-artefact, measured-mid-flight.

**2026-07-25 — blamed the nearest restart for a mid-run death that a cron on a
10-minute grid caused.** The first feature-implementer run (2b1a154e) died at an
s4 `stage_commit` await; I recorded the cause as "minutes after the 16:29:38Z
chassis restart; git-adapter replicas restarted; response unrecoverable" (bug 003
sighting 5, commit 442c4b48d) and re-fired on that theory. The re-fire died the
same way at 20:03 — nowhere near any restart — and THAT forensics showed
git-adapter producing the response 4s after the request while nobody consumed
it: the `agent-job-cleanup` cron was deleting every live `job.*` topic each tick,
because its pod-label guard has never matched anything (`bugs_open/071`). The
restart theory failed its own timeline (the "killed" consumer consumed
successfully at 16:32:25, three minutes AFTER the restart) — I never re-checked
the story against the sequence I already had. **Cheap check, two parts: (1) when
an event's cause is unknown, put its timestamp against every `*/N` schedule
running in the cluster (`kubectl get cronjobs`) before accepting the nearest
dramatic event; (2) a causal story must survive its own timeline — anything the
"cause" should have broken that demonstrably worked afterwards refutes it.**
Cost: one wrong durable claim in 003's sighting list (corrected in place), one
re-fire burned on an unfixed cause, ~4h of implementer progress lost a second
time. Family: correlation-not-cause, absence-family (the deleted topic left no
error anywhere — the produce SUCCEEDED into the recreated topic).

## 2026-07-25 — "we know which boards people subscribed to" (relojistas legacy feeds)

**The claim.** The rebuild manifest recorded `Subscribed variants seen:
forumids=2,4,13,44,58,78,145,288`, and the handoff promoted it into a planned feature:
map those boards to topic feeds "so a Rolex-board subscriber gets Rolex news". It sat in
the plan of record for two weeks and was the last item on the build list.

**What was actually true.** The ids were an unchecked sample (real set: 123 boards), and
the traffic is ~88% self-identified crawlers — meta-webindexer, Googlebot, DotBot,
SERanking, plus a scraper spoofing Chrome/30 — with **zero conditional GETs**. The
genuine subscriber signal sits on the *bare* feed URL (one client, 42 × `304`). There is
no board subscriber to serve. The named example was wrong too: the Rolex board is id 43
and is not in the recorded set.

**What caught it.** Going to gather the forumid→board mapping in order to *build* the
feature, and reading the user-agent column while there.

**The cheap check that would have caught it two weeks earlier.** One extra column in the
grep that produced the claim: `awk '{print $9, $NF}'` — status and user-agent alongside
the query string. Counting query strings measures *requests*; the word "subscribed" is a
claim about *clients*, and nothing in that grep looked at a client.

**Transferable rule.** When a log-derived count is about to be described with a noun that
implies a person — *subscriber*, *visitor*, *user* — the grep must include whatever column
distinguishes people from machines, or the noun must change to "requests".

---

## 2026-07-25 — "ratified D4: F2+F3 ship together" (bugfix-003 council submission)

**The claim.** The bug-003 council submission stated as an owner ruling: "owner-ratified
D4: F2+F3 ship together — either alone is a no-op or worse", and the workstream memory
carried the same line for five days ("F2/F3 ship TOGETHER").

**What was actually true.** D4 as ratified binds **F3's two halves** to each other (the
offset-commit fix must ship with the completion-time dedupe): *"F3 ships only together
with the completion-time dedupe … Either alone is a no-op or worse — 003 §4.4a-bis."*
It says nothing about F2. Shipping F2+F3 in one roll was the owner's *scope choice* in
the 2026-07-23 session — a preference, not an interlock. The distinction mattered the
day the council's guardian proposed splitting the roll: the rebuttal I nearly wrote
("D4 forbids the split") would have been false, and checking the pre-edit code showed
F2-without-F3 is in fact mechanically viable (exact-version dedupe match + a swallowed
unique-violation error fail-open into processing).

**What caught it.** The guardian veto forced a re-read of the PLAN's decision text
verbatim before drafting the rebuttal.

**The cheap check that would have caught it.** Quote the ruling, don't paraphrase it:
when citing a numbered decision, paste its actual sentence next to the claim. The
paraphrase had drifted through three documents (bug file sequencing note → memory →
submission), each copying the previous, none re-reading D4 itself.

**Transferable rule.** An owner ruling cited from memory is a claim, not a citation.
Before wielding "ratified DN says X" in an argument — especially against a reviewer —
open the doc and paste the sentence. If the sentence doesn't contain X, the argument
changes shape.

---

## 2026-07-25 — "the spawn was dropped" (idea.uk guides-hub rerender). Wrong: it was queued.

**The claim.** Direct-fired a `page-rerender` for idea.uk's `/guides/index.html` via kcat.
Four minutes later no `orchestration_states` row existed for the correlation, so I said
*"the hub rerender spawn was dropped (no orchestration row — the bugs_open/003 shape)"*,
re-fired it, and wrote it into the workstream RUNBOOK as a standing trap ("the spawn can be
silently dropped"). Two identical messages were now in flight.

**What was actually true.** The message was on the topic the whole time (confirmed by reading
`system.agent.generic.requests` at offset 103566). Nothing was dropped — the generic-requests
consumer was **stalled**: `SELECT ... FROM orchestration_states WHERE created_at > now() -
interval '10 minutes'` returned **zero rows fleet-wide**, i.e. every thread's dispatch was
waiting, not just mine (`bugs_open/029/030`).

**What caught it.** Running that fleet-wide query — which I only ran *after* re-firing, while
hunting for the spawner that was supposedly at fault.

**The cheap check that would have caught it.** One query, before re-firing: *are ANYONE's
orchestrations starting right now?* Absence of my row is not evidence about my message until
I know the consumer is otherwise alive. Corollary noticed while chasing this: "no
`agent-page-rerender` pod is running" proves nothing either — those are one-shot Jobs that
idle-shut-down after ~3 minutes (`agentbase/agent.go:1541`, observed `idle_duration 184s`).

**Transferable rule.** This is exactly `memory/council-queue-latency-trap`, already written
down after the same mistake against the council gate — *"no orchestration rows means QUEUED,
not dropped; check when OTHER orchestrations started before resubmitting"*. Having the note
did not stop me; the note fires on the word "council", and this was a page rerender. **Absence
of your row is a claim about the queue, and a claim about the queue needs a query about the
queue — whatever subsystem you are firing at.** Cost here was one duplicate (idempotent)
rerender; the same reflex against the council costs a whole duplicate review round.

**2026-07-25 — wrote a cause into the RUNBOOK from a single observation, and had
to correct it 90 minutes later.** A direct kcat dispatch produced no orchestration
row. It had three headers; the working script in another workstream uses ten. I
wrote the RUNBOOK entry "the FULL header set is load-bearing — a dispatch with only
correlation_id/client_id/request_id produces NO orchestration row and NO error",
stated as mechanism, and committed it. Then I sent four more dispatches WITH all ten
headers: **all four also produced nothing — 5 of 5, header count irrelevant.** The
real situation is that direct dispatch was not landing at all that morning while the
chassis was demonstrably healthy and consuming the same topic from other producers,
and none of the five correlation ids reached its log — i.e. the messages were never
received, which the header hypothesis would not explain either. **Cheap check: I had
the disconfirming test available before I wrote the claim — send one with the full
set and see.** Instead I ran the experiment after publishing the conclusion, which
is the wrong order and produced a RUNBOOK that would have sent the next thread
hunting a header bug that does not exist. What makes this worth logging rather than
shrugging off: the entry LOOKED like exactly the kind of hard-won gotcha a RUNBOOK
is for — specific, mechanical, actionable — and that plausibility is precisely why
an unverified cause in an operational doc is more dangerous than one in NOTES.
Corrected in place with the five-row evidence table and an [UNVERIFIED] marker on
what remains unknown. Family: cause-invented-to-fit-one-observation,
published-before-tested.

**2026-07-25 — reported a re-verification sweep as "live and proven" when it had
never processed a single row, and said so to the owner.** The model directory's
daily citation re-check was described in NOTES as live, and in the owner's log as
"each one due to be automatically re-checked every day so a stale price can't sit
there looking authoritative". Measured today: **0 orchestrations had ever carried
`refresh_directory_claims`, and 0 of 108 claims had ever been re-verified.** The
task put its workflow inline in `scheduled_tasks.input_data.config.workflow` and
targeted `generic`, whose own workflow is a single no-op step — so every fire
created a run, completed instantly, and stamped `last_triggered_at` AND
`last_completed_at`. What I checked at the time was that the task fired and
completed. It did. That was never the question. **Cheap check: for any job whose
output is a row change, count the row change, not the job — one
`count(*) FILTER (WHERE verified_at > created_at + interval '1 minute')` would
have shown 0 on day one.** What actually caught it was inducing a fault
(corrupting a stored quote and waiting for a flip that never came), which I did
for an unrelated reason — a 17/17 verification score I found implausible. Two
lessons, and the second is the one I keep re-learning: (1) a config field that is
ACCEPTED but never READ is indistinguishable from one that works, and the same
shape is live in another workstream's `evidence-freshness` (filed as
bugs_open/074 rather than fixed for them); (2) I had the workstream's own memory
note "verify the failing branch — a green happy path proves DEPLOYMENT not
CORRECTNESS; induce the fault for anything whose job is to detect one" and did
not apply it to my own sweep, which is *precisely* a thing whose job is to detect
a fault. Family: green-status-no-work, checked-the-job-not-the-output,
own-rule-not-applied-to-own-work.

---

## 2026-07-25 — "the lock protects the copy, not the derived list". It freezes the whole row.

**The claim.** Locking idea.uk's guides-hub listing section (`p4_03`), I wrote in the SQL and
repeated it in the commit message: *"The hub's listing copy (the items themselves stay
query-resolved and must NOT be frozen — the lock protects the surrounding copy, not the derived
list)."* Stated as fact, in the artefact that performs the action, with no check.

**What was actually true.** A lock is applied to the `page_components` **row**, and cannot
distinguish authored copy from derived items because both live in that row.
`SavePageSectionsAction` (`save_page_sections_action.go:487-534`) preloads actively-locked rows,
holds them out of the rebuild DELETE and re-attaches them verbatim — the comment in the code reads
*"Human-locked rows must survive the rebuild with copy AND row identity"*, and it logs *"preserving
human-locked section over rebuilt copy"*. So the lock freezes the derivation outright.

**What it would have cost.** The guides hub derives its list from
`query.pages_where_type:guide` — that self-populating listing is the entire reusable contribution
of the day's work. Locked, it would have kept rendering exactly one card forever: every future
guide written, deployed, and silently never listed, with each render reporting success. A frozen
derivation is indistinguishable from a working one until the data it tracks has moved on, which is
why nothing would have flagged it.

**What caught it.** Luck, in the end-to-end sweep. The copyright guide shipped 25 minutes after the
hub's last render, so the hub was legitimately one card behind at the moment I looked — and that
made me ask why. Had I locked before adding a second guide, or looked five minutes earlier, the
sweep would have shown one card and one guide and looked perfect.

**The cheap check that would have caught it.** Read the code path that honours the lock before
asserting what the lock does — one grep for `locked` in `save_page_sections_action.go`, which is
exactly what I did *after* the sweep and which answered it in seconds. The claim was about a
mechanism, and I had the mechanism on disk.

**Transferable rule.** **Never lock a section whose component schema has any `query.*` source.**
Locks are for authored content only; protect a deriving section by making its authored fields
content_data-driven (so nothing wants to regenerate them) rather than by freezing the row. Recorded
as a `doc_note` under `component_locks` and corrected by `p4_08`. More generally: this is the
second entry today where I described platform behaviour confidently from intent rather than from
the code — see the queued-vs-dropped-spawn entry above. Both were one grep or one query away.

**2026-07-25 — a pathspec commit of a `git mv` took the new files and left the
staged deletions, and committed HEAD did not compile for ~50 minutes.** Phase E
renamed three files (`git mv model_directory_items.go directory_items.go`, etc.).
A mv stages a delete(old)+add(new) PAIR; my commit named only the new paths, so
the four old files stayed in HEAD alongside their renamed copies —
`ModelDirectoryClaim redeclared in this block`, `QueryModelDirectoryEntries
redeclared` — and any concurrent session running the committed-HEAD image build
(`make build-<service>`, the fleet default) in that window would have failed at
compile. Verified by actually building `git archive HEAD` — which is also the
**cheap check I skipped**: the shared-tree rule I already carry
([[shared-tree-wont-compile]]) says test against `git archive HEAD` + your files,
and I built the *working tree* instead, where both halves of the mv are present
and everything passes. Caught by an end-of-task `git status` sweep noticing four
staged `D` entries that were "mine" but uncommitted. The trap generalises: **the
commit-per-task pathspec discipline and `git mv` interact badly** — the pathspec
must name BOTH sides of every rename (old path and new), exactly as CLAUDE.md's
"name it twice" rule for `add`+`commit` of new files, one file earlier in the
same lifecycle. Loud failure, so cheap as these go: a build break announces
itself, unlike the silent classes this file mostly records. Family:
rename-is-two-paths, verified-the-wrong-tree.

**2026-07-25 — I read "queued" off a polling loop for 60 minutes; the run had
failed validation in 7 seconds.** Submitted `040` candidate 2 to the council gate
and wrote my own watcher instead of using the two queries the 097 trigger prints.
Mine said `SELECT ... FROM orchestration_states WHERE id='<uuid>'` — the column is
`orchestration_id`; there is no `id`. Every iteration errored, and because I had
written `2>/dev/null` the error went nowhere and the empty result printed as my
own hand-written default string, `<no row yet: queued>`. So the loop confidently
reported a queue wait, 60 times, for a run that had reached `complete_invalid`
before the first tick: `plan does not match the fix-plan schema: json: cannot
unmarshal array into Go struct field fixPlan.risks of type string` (`risks` is a
`string`, not a list; and my second edit's `operation: "create"` is not in the
allowlist either — `add` is). The real state was in `collected_data.__step_error`
the whole time.

**What caught it.** The loop hitting its own 60-minute timeout. Nothing else would
have — the output looked healthy, and CLAUDE.md's own guidance ("a missing
orchestration row is almost always latency, not a dropped dispatch — do not
retry") is exactly the story my broken output was telling, so the more carefully I
followed the runbook the longer I would have waited.

**The cheap check that would have caught it.** Run the query **once, in the
foreground, with stderr visible**, before wrapping it in a loop — the SQL error
names the bad column immediately. Or simply paste the trigger's own printed
watch-queries, which are correct.

**Transferable rule.** **`2>/dev/null` in a polling loop converts every error into
your default branch, and a default branch you wrote yourself will always sound
plausible.** Never suppress stderr in a watcher; distinguish "no row" from "query
failed" explicitly, and give a poll a fail-fast case for the terminal states
(`complete_invalid`, `FAILED`) as well as the success one, so a dead run announces
itself instead of aging into a timeout. The near-miss beneath it is worse than the
lost hour: the same silence would have hidden a *rejected* verdict just as well as
an invalid one. Family: absence-of-evidence (with [[the 24h-pruned record]] entry
above — the record was there, my instrument wasn't), verified-the-wrong-thing.

**Footnote worth the irony.** This happened while shipping a fix whose entire
subject is *a failure that was durably recorded and simply not shown on the
surface anyone reads*. The council run recorded its reason in `__step_error`, in
the same field the fix propagates. I built the fix and then walked into the bug.

---

## 2026-07-25 — I declared two of three regulatory claims "unverified" from a landing page that only lists filenames (vetcomparison.uk session)

**The claim I made.** vetcomparison.uk's homepage news section carries generated
framing copy asserting the CMA's draft Order contains three remedies: a standard
price list requirement, a written prescription fee cap, and ownership disclosure
obligations. Checking this matters more here than anywhere — this site was
remediated for publishing fabricated content, and unsourced regulatory assertions
are its exact failure mode. I fetched the gov.uk consultation page, found it
supported the price list, said nothing about a fee cap amount, and listed no
ownership document at all, and reported: **"Two of the three claims don't check
out against the source."**

**What was actually true.** All three are in the Order, plainly:
Article 7 (Price List), Article 18 — *"imposes maximum fees (also referred to as
price caps)"*, establishing a Primary and an Additional Prescription Fee Cap — and
Article 5 (Ownership Information), whose own explanatory text runs to eight
paragraphs. The generated copy was accurate. My "finding" was an artefact of the
document I read.

**Why the landing page said otherwise.** It is an index. It renders *document
titles* — "Draft schedule 1 - price list schedule", "Draft schedule 2 - standard
written prescription notice" — and no substance. Ownership has no schedule of its
own because it is an Article, not a schedule, so it is invisible at that altitude.
I treated an absence in a table of contents as an absence in the law.

**What caught it.** Continuing to the primary source anyway — pulling the eleven
linked PDFs and running `pdftotext -layout`. Roughly four minutes. Note that the
first two WebFetch attempts on those same PDFs *also* failed ("heavily
compressed/encoded... text layers are not clearly readable") — a second plausible
"I checked and could not confirm" that was still an instrument failure, not a fact.
Two different tools returned nothing and neither absence was evidence.

**The cheap check that would have caught it.** Ask *what kind of document am I
reading* before drawing a conclusion from it. A consultation landing page, a
search-results page, a file listing and an index all answer "does X exist?" with
"there is no row for X" regardless of the truth. For any claim about the contents
of a document, **the artefact that settles it is the document.**

**Transferable rule.** **A negative from an index is not a negative about the
thing indexed** — and the failure escalates rather than resolves when the tool
also fails, because "fetched, read, not found" and "fetched, unreadable, not
found" produce the same sentence in your notes. State the artefact you actually
read next to the claim ("the landing page does not mention X" — not "X is not in
the Order"). Family: absence-of-evidence, verified-the-wrong-thing — the same
shape as the 24h-pruned-record and suppressed-stderr entries above, reached
through a third instrument.

**The sting.** The claim I nearly filed was that a *fabrication-remediated site
was publishing unsupported regulatory claims* — a serious, credible-sounding
allegation about the one failure mode this site is watched for. It would have
been believed. Confidence tracked how *bad* the finding sounded, not how well I
had checked it, and a bug report is exactly where an unchecked negative does its
damage. **On this site the unverified-claim discipline applies to my own findings
about it, not only to its content.**

---

## 2026-07-25 — I wrote a `Council-Reviewed:` trailer for a verdict I had not read

**The claim.** Commit `8e8b55818` (gripper dossier prose gate) carries the
trailer `Council-Reviewed: 7ed137d1`. It does not have council approval. The
verdict on 7ed137d1 was **REVISE**, gated by a high-severity compliance
objection, and had been sitting COMPLETED in the DB since 20:30Z the previous
evening — before I wrote the commit.

**What I actually did.** I knew the submission existed and that it was mid-run
when I last looked. When I came to commit, I reached for the correlation id
from my own NOTES and appended it, treating "I submitted this lane for review"
as though it were "this lane was reviewed". I never ran the one query that
distinguishes them.

**What caught it.** Writing the commit message made me notice the trailer
discipline note in my own memory index — *the trailer is earned by an APPROVED
verdict ONLY* — so I checked the verdict immediately after committing. Thirty
seconds later. Forward-only means the false trailer is now permanent in the
history.

**The cheap check that would have caught it.** The query is one line and takes
two seconds:
`SELECT current_step, status FROM orchestration_states WHERE
 collected_data->'input_data'->>'fix_correlation_id' = '<CORR>';`
Run it **before** typing the trailer, not after. A pending submission and an
approved one look identical from your notes; they differ only in the DB.

**Transferable rule.** **A correlation id proves a submission, not a verdict.**
The two are different facts and only one of them is a claim of review. The
trailer is not an audit trail of "I engaged with the council" — 098 joins on it
to assert *this change was approved*, so a trailer on a REVISE does not merely
overstate, it feeds a false row into the coverage report that exists to find
unreviewed platform code. Family: the same shape as every other entry here —
an inference (submitted ⇒ reviewed) written in the same voice as a finding,
with no marker and no check.

**The sting.** The council had done its job well: the REVISE was gated on a
*real* gap — a fabricated plain-word vendor name passing my honesty gate — in a
feature whose entire value proposition is that names and numbers trace or the
run fails. I had stamped "reviewed" on the very batch whose review was telling
me the honesty gate had a hole in it. The objections were worth acting on, and
acting on them took an hour; claiming they had already been satisfied took
eight characters.

---

## 2026-07-25 — twice we called it a "borrowed" component; nothing was ever borrowed

**The claim.** `bugs_open/028` was filed (2026-07-19) asserting a mechanism:
*"Something is falling back to site-level or sibling components when a page has
none of its own, and that fallback is invisible in the output."* The case title
itself says **"deploys borrowed components"**. The 2026-07-21 re-scope corrected
the detail but kept the shape: *"it is a site-level/default hero **applied**
where no page-specific hero was authored."*

Both are wrong in the same way. **No component is copied, borrowed, or applied
from anywhere.** Each hero is generated fresh by `page-content-writer`. The
writer's only page-subject signal is `{{.current_page.title}}`, and on a glossary
entity-page that title is a single bare term ("Tourbillon"). With no
`content_direction`, no `meta_description` and an empty `content_brief`, the
model has essentially no subject, so it writes a site-level hero paraphrased from
the site's `value_proposition`. 6 of 8 pages fell that way; 2 did not. There is
no fallback path to find, because the variance is in the model, not in a branch.

**What caught it.** Reading the three `content_data` blobs side by side — a
broken page, a correct page, and the homepage. If the hero had been *copied* the
rows would match. They do not: only `headline` matches the homepage; `subheadline`,
`cta_text`, `hero_url` and `background_image` all differ. A copy cannot produce
that. Everything after followed from noticing the mismatch.

**The cheap check that would have caught it.** One query, run at filing time:

```sql
SELECT p.name, jsonb_pretty(pc.content_data)
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE pc.slot_name = 'hero' AND p.name IN ('<broken>', '<the page you think it came from>');
```

**Transferable rule.** **"Borrowed" is a conclusion about provenance, and
provenance is not visible in a rendered page — only in a field-by-field diff of
the source rows.** We saw one string appear on two pages and named a mechanism
(a fallback) to explain it. Two strings matching is equally consistent with
*generated from shared context*, which is what happened, and that hypothesis
costs one query to separate from the other. The tell we both walked past: the
generic hero's headline was not even the *current* homepage headline, which
already made "copied from the homepage" awkward — the 07-21 note recorded that
oddity and still reached for a fallback to explain it, rather than letting it
falsify the family.

**The sting.** The 07-21 re-scope was otherwise a careful, well-grounded piece of
work — it correctly closed defect 1 against live data, correctly delegated
defect 2 to `040`, and correctly bounded the blast radius to one site. It
inherited exactly one thing from the original filing without re-testing it: the
word "borrowed". A re-scope that re-derives everything else from the live DB is
the *cheapest* possible place to catch an inherited premise, and it is the place
we didn't look, because the premise was in the title.

---

## 2026-07-25 — "confirmed live: four terminal improve_tool cycles" (bugs_open/010 b)

**The claim.** `bugs_open/010` recorded, as the live-evidence base for the
convergence guard: *"Confirmed live 2026-07-20 by the count the guard now runs:
**four** terminal `improve_tool` cycles against the same criterion (`mobile-fit`)
on this one tool between 07-17 and 07-18."* It was written in the voice of a
measurement, next to the query that supposedly produced it.

**What was actually true.** Four is the number of `acceptance-fail` **doc_notes**.
The guard counts **`site_work_items` rows** under
`item_key='acceptance_fail:<fn>:<site>'` with `status IN ('complete','failed')`.
Exactly **one** such row has ever existed for that tool. At `max_fix_cycles=2`
the guard would have counted **1** on the benchmark and **would not have
escalated** — on the very case it was built to catch. The mechanism the bug
describes (the loop repeats unboundedly) was real; the evidence that the fix
addressed it was not.

**What caught it.** Re-running the guard's own `convergenceAttempts` SQL against
prod while trying to close the case, rather than re-reading the sentence that
asserted it. Two queries, four days late.

**The cheap check that would have caught it.** Run the shipped query itself,
with a positive control, at the moment you claim it confirms something:

```sql
-- the count as the guard computes it (0 = correctly reset by a later pass)
SELECT count(*) FROM site_work_items w
WHERE w.item_key='acceptance_fail:<fn>:<site>' AND w.status IN ('complete','failed')
  AND w.created_at > COALESCE((SELECT max(created_at) FROM doc_notes
        WHERE subject_type='tool' AND subject_key='<fn>' AND source='tool-acceptance'
          AND categories @> '["acceptance-run"]'::jsonb), '-infinity'::timestamptz);
-- positive control: same query, reset-bound removed. If this is 0 too, the
-- query matches nothing ever and the first 0 told you nothing.
```

**Transferable rule.** **A number is only evidence for a guard if it is the
number the guard actually computes.** Counting notes and counting work items feel
interchangeable when both are "how many times did this fail" — they are not, and
here they differ by 4×, across the threshold. When the claim is "the fix would
have fired", the only admissible evidence is the fix's own predicate, executed,
**with a positive control** — a bare `0` is equally consistent with "correctly
reset" and "query is broken", and only the control separates them.

**The sting.** The file was, by then, unusually careful: it had already corrected
one half of its own diagnosis, disclosed that the escalation branch was
live-unexercised, and banked a lesson about mistaking a historical count for
current state. It carried a *number* forward unchecked while re-checking
everything around it — and that number was the one thing standing between
"logic is unit-tested" and "confirmed live". Same shape as the entry above:
the premise nobody re-derives is the one already written down.

---

## 2026-07-25 — "the kcat stdin race ate my probe message" (bugs_open/021 INSTANCE 2). Wrong: queued. **Second time today, same trap.**

**The claim.** Fired a scratch `complete_work_item` probe at
`system.agent.generic.requests`. Nine minutes later there was no
`orchestration_states` row, no chassis log line mentioning the agent type, and no
spawned pod. I concluded the message had never been produced — blaming the
documented `kubectl run -i` stdin race — added `-c 1` to the kcat invocation and
re-fired. I then wrote the `-c 1` requirement into the workstream RUNBOOK as a
standing trap, with a cost figure attached.

**What was actually true.** Both messages were on the topic the entire time — I
confirmed it by reading the topic tail *after* re-firing, and saw my payload
twice. Ten minutes later **both** ran, 24 seconds apart, and both behaved
identically. `-c 1` changed nothing. The generic-requests consumer was wedged:
`CURRENT-OFFSET` frozen at 104102 while `LAG` climbed 20 → 28, because a
16-seat council run belonging to another session was executing its LLM steps
inside the single consumer loop (`bugs_open/030`).

**What caught it.** Counting orchestration rows for all three of my fires and
finding two where I expected one. The duplicate run is the evidence: had the
first message really been lost, there would have been nothing to run twice.

**The cheap check that would have caught it.** Read the topic BEFORE re-firing —
sixty seconds, no credits, and it answers the only question that matters
(produced or not):

```
kubectl -n kafka run -i --rm kcat-read-$(date +%s) --image=edenhill/kcat:1.7.1 \
  --restart=Never -- kcat -C -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests -o -6 -e -q
```
Then, if it IS there, `kafka-consumer-groups.sh --describe --group
generic-requests-group`: a frozen `CURRENT-OFFSET` is the whole answer.

**Why this entry earns its place despite being a duplicate.** It IS the duplicate
— *"the spawn was dropped" (idea.uk guides-hub rerender). Wrong: it was queued*,
filed earlier the SAME DAY by another session, which also re-fired and also wrote
its wrong inference into a runbook as a standing trap. Two threads, one trap, one
day, neither able to warn the other. Both had a live-fleet symptom (a stalled
consumer) that is invisible from inside one thread's evidence, and both reached
for a mechanism they had read about instead of the observation that separates the
two. **The tally is the argument:** "no row yet" is being read as "dropped"
often enough that the diagnostic belongs in the trigger scripts — print the
consumer lag on fire, so the queued case announces itself and nobody has to
suspect it.

> **CORRECTED, within the hour, by checking instead of recommending.** That
> paragraph originally ended *"noted here rather than built, because the trigger
> scripts are the fixloop thread's"* — a tidy suggestion I had not verified. **It
> already exists.** `scripts/dispatch-queue-depth.sh` (commit `a5a494459`, filed
> **the same day** by the `bugs_open/030` thread) prints exactly this — consumer
> position, `QUEUE DEPTH (LAG)`, and in its own words *"your message is QUEUED. It
> is not lost. Do NOT re-fire."* — and `097` and `090` already call it on every
> fire. I did not see it because I copied my probe envelope from **`091`**, which
> is not wired to it (nor is `092`). So the honest finding is smaller and more
> useful than the recommendation: **the fix is written, and the trap survives in
> the un-wired triggers and in every hand-rolled kcat fire copied from them.**
> Wiring `091`/`092` is one line each; anyone hand-rolling a probe should just run
> the script. Recommending a build for something already built is the same error
> as this entry's headline — asserting the state of the world from inside one
> thread's view.

**The second-order sting.** My wrong inference was already written into
`RUNBOOK_durable_write_guard.md` — with a fabricated-looking cost figure — before
I disproved it, and it read entirely plausibly. What made it removable was
finishing the verification instead of stopping at the fix that appeared to work:
re-firing "worked", and if I had stopped there the RUNBOOK would now teach a
`-c 1` cargo cult to every thread that follows. **A workaround that coincides
with the real cause resolving is indistinguishable from a fix** unless you go
back and check what the original attempt did.

---

## 2026-07-25 — "wire `reconcile_section_data_action` and 48 items drain" (`bugs_open/033`). Wrong: it drains **zero**.

**The claim.** `bugs_open/033`'s grounding pass (2026-07-20) named the automated
consumer the queue was missing: *"The one genuine automated consumer,
`reconcile_section_data_action.go:114-116` (re-opens `needs_section_data` when
query-sourced data later resolves — **48 items of the queue**), is registered as
an action but wired to **0 live agents**."* That became fix candidate A, it was
carried into the memory index and into `OPEN_THREADS_RESTART_LIST.md`, and it sat
there for five days as the thing to do next.

**What was actually true.** The action only re-triggers an item when **every**
one of its missing fields is `query.`-sourced (`strings.HasPrefix(m.Source,
"query.")`, else `allQuery = false` and the item is skipped). Live, 2026-07-25:

```sql
WITH s AS (SELECT (SELECT bool_and(m->>'source' LIKE 'query.%')
                   FROM jsonb_array_elements(spec->'missing') m) AS all_query
           FROM site_work_items
           WHERE status='needs_human_review' AND item_type='needs_section_data'
             AND jsonb_typeof(spec->'missing')='array')
SELECT all_query, count(*) FROM s GROUP BY 1;
--  f | 30      <- and no 't' row at all
```

**0 of 45.** Thirty carry `site_specs.*` / `site_assets.*` sources (image, email,
pricing tiers, case-study urls — human data by construction) and fifteen carry
`missing: null`. Wiring it would have re-triggered nothing, and the wiring would
have looked correct while doing so: the action returns success with
`retriggered_pages: []`.

**What caught it.** Sizing the fix before building it — one `GROUP BY` over the
population the fix was aimed at, run because the plan said "48 items" and I
wanted the per-site split.

**The cheap check that would have caught it.** The count that the number itself
came from. "48 items of the queue" was a count of `needs_section_data` **rows**,
not a count of rows the action would ACT on; the two were never the same number
and no query was run against the action's actual predicate. The check is: take
the condition out of the code, put it in the `WHERE`, and count that.

**Why it matters more than one stale figure.** The claim was in the *fix
candidates* section, which is the part of a bug file the next thread executes
rather than re-derives — and it had already propagated into two indexes. The
failure mode is specific and worth naming: **counting the population a fix is
ABOUT instead of the population it would ACT on.** A fix sized that way looks
proportionate right up to the moment it ships and moves nothing, and the thread
that ships it is not obviously wrong — the code does what it says.

**Same day, my own version of it, so this is not a finger-point.** Re-validating
those items, I keyed the first queries on `spec->>'component_id'` and got "30 of
30 components gone" — which I nearly wrote up as "the queue points at deleted
components". `page_components.id` is not stable across re-renders; the sections
were all still there under new row ids, and my own memory notes already said so.
What caught it was the number being *too* clean: 30/30 is a join bug, not a
defect signature. The cheap check was to print the page's actual slot list beside
the wanted slot.

---

## 2026-07-25 — "68 ungated anchors / 37 components" (bugs_open/023 HANDOFF §5): a heuristic reading, quoted as a live re-measurement

**The claim.** The CTA/link-integrity HANDOFF's still-open worklist said, under a heading
stating all figures were *"re-measured live 2026-07-21"*: **"P2.1 — gate the ungated anchors.
68 ungated / 37 components (was 75/38 on 07-19)."**

**Why it was false.** That is an **R9** reading — the 60-character-lookback regex whose own
RUNBOOK entry, corrected the day before, already said it undercounts by **2.4x**, because
`regexp_matches(…,'g')` returns non-overlapping matches and each match's greedy prefix eats the
*previous* anchor, so in runs of adjacent anchors (nav lists, footer columns — exactly where CTAs
cluster) roughly every other anchor is never counted. The true figure on 2026-07-21 was ~171/41.
On 2026-07-25 the real parse gave **173 ungated / 43 components**, of which **156 / 31** are the
CTA worklist and 17 are range-scoped item links (a different class).

**What caught it.** Running the parse (R9b / `parse_gates.py`) before scoping the migration, and
noticing the answer was 2.5x the handoff's. Not a re-read — a re-measurement.

**The cheap check that would have.** Run the corrected tool. The correction and the stale figure
lived **in the same directory, one file apart**: R9's warning note and R9b were written on
2026-07-20; the 68/37 line was written on 2026-07-21.

**Why it matters more than one stale figure.** It was in the *still open* section — the part a
resuming thread executes rather than re-derives — and it under-sized the job by 2.5x. A worklist
sized that way looks finishable in an afternoon and then does not close the bug. **Fixing a
measurement tool does not retroactively fix the numbers already copied out of it, and a figure
inside a "re-measured live" block wears a confidence it has not earned.** The general form is the
one this file keeps recording: *a number travels; its caveat does not.*

**My own version of the same shape, same session, caught before it reached anything.** I wrote
the migration's SQL as a direct translation of a Python transform I had already audited
line-by-line — and believed the two were equivalent because they are character-for-character the
same regex. They are not: in POSIX ARE the **first** quantifier fixes the greediness of the whole
expression, so a leading greedy `[^>]*` makes a following `.*?` greedy too, and the anchor match
ran to the **last** `</a>` in the template. It would have swallowed the tail of 21 of 35 live
component templates. What caught it was hashing the expected result offline and asking Postgres
to hash its own — 21 of 31 rows came back `false`. The cheap check is the entry in the tally:
**prove a transform against the engine that will run it.** Reading it again would not have done
it; I had read it several times, and it looks obviously right.

**2026-07-25 — declared a link fix verified using a check that shared the fix's
own blind spot.** I censused fundamentallyai.com's internal links with
`regexp_matches(html, 'href="(/[^"#?]*)"')`, found 21 broken of 22, repaired them
with quoted-exact string replacement, re-ran the census, got **zero remaining**,
and reported the site's links fixed. The census regex requires the closing quote
immediately after the path, so it silently **excluded every anchored href** —
`/capabilities#approach` and 20 siblings. My replacement used the same
quoted-exact form, so it skipped precisely the class the census could not see,
and the post-check confirmed success because it reused the blind pattern. A live
crawl of the served pages found the 21 survivors minutes later. **Cheap check:
when the fix and its verification share a regex, a query, or an assumption, the
verification cannot falsify the fix — it can only echo it. Verify against an
independent witness: the served artefact, not the source you just edited.** This
is the same day's second instance of the same shape (the other: my `SELECT`s
truncated `summary` to 50 chars, hiding the `[stale: triaged 48h+]` prefix that
named the mechanism I then spent two theories guessing at). Correct form here:
capture `href="(/[^"]*)"`, then `split_part(href,'#',1)` before resolving. Cost:
one false "fixed" claim to the owner, corrected in the same session, plus a
second repair pass and republish cycle. Family: shared-blind-spot,
verification-echoes-the-fix, truncated-your-own-evidence.

**2026-07-25 — wrote "the dispatch failed silently" into a runbook as a landmine,
from two checks made ~5 minutes too early.** Four `049b_deploy_single_page.sh`
republish dispatches for fundamentallyai.com showed no `orchestration_states` row
when I queried at roughly +2 and +5 minutes. There is a documented stdin-race
failure mode for that envelope (016b §9), so I matched my observation to it,
declared four silent failures, switched to a different script, and **recorded the
claim in a new RUNBOOK and in a new script's header as established fact.** All
four had landed: their rows appeared at 17:12–17:13 for dispatches fired ~17:05,
each carrying the right `page_id`. The replacement script took ~9 minutes for its
own first row. **Neither route is unreliable; dispatch latency here is minutes,
and I had no measurement of what normal looked like before calling it broken.**
Compounded twice over: I then armed a monitor filtering
`created_at > '2026-07-25 18:00'` while the clock read **17:54** — a window that
cannot match anything and reports identically to a dead dispatch, which produced
four more "no orchestration row" ticks that felt like confirmation; and because I
never checked `start_step`, I credited the new script for a homepage republish the
queued work item may equally have done. **Cheap checks: (1) before concluding
"nothing happened", establish the normal latency — one `SELECT max(created_at)`
against comparable dispatches would have shown minutes, not seconds; (2) an
absent row is only evidence if your query window can contain it — print the
window and the clock together; (3) a known failure mode that MATCHES your symptom
is a hypothesis, not a diagnosis, and it is at its most dangerous when the docs
have already made it famous.** Cost: one wrong landmine published in two places
(both corrected same day, with the correction left visible), one unnecessary
route switch, ~15 minutes of false-negative polling. Family:
too-early-is-not-absent, window-cannot-contain-the-evidence,
famous-failure-mode-as-default-diagnosis.

**Follow-up, same day (the sting sharpened).** Correlation `7ed137d1` was
resubmitted twice more and eventually came back **APPROVED** — so the trailer on
`8e8b55818` now resolves to a genuine approval. It is still false, and now
falsely in a way that is harder to see: the round that approved is precisely the
round that change was **removed from**, at the edit-quality seat's request for
minimality. The trailer reads as reviewed, resolves to an APPROVED verdict, and
the verdict is for other code.

**The rule is therefore sharper than "read the verdict before writing the
trailer".** A correlation id identifies a *submission*, and a submission's
contents change between rounds. **A trailer is only true if the approved round
still contained your change.** Re-check the approved plan's edit list, not just
its decision. (Fixed forward: the change was resubmitted alone as
`37a32e02-19a7-409a-a74f-9363556bb39e`, with the situation stated in its
rationale, so its real status resolves by that correlation.)

**Closed out the same day.** `37a32e02` came back **approved** (18:33Z, 12 seats,
5 advisory objections, none high-severity). So the change is genuinely reviewed
— under a correlation its own commit does not name. **The permanent artefact of
this mistake is a commit whose trailer resolves to an approval for other code**,
and `098` cannot detect it, because both correlations return `approved`. Only
the NOTES annotation distinguishes them.

Worth recording the shape of the fix as well as the error: the split the
council forced (one change, one submission) produced a *better* record than the
bundle would have — every change now resolves to the verdict that actually
judged it. The bundling was mine, and I resisted the split as bureaucratic
before doing it.

---

## 2026-07-26 — "idea-uk@leopardess.uk is stale/dead" (idea.uk tool). Wrong: it is the tool's correct address.

**The claim.** Across §X.14–X.17, the paid-tool AUDIT doc and two owner walkthroughs, I called
`idea-uk@leopardess.uk` "the stale/dead address", flagged the engine's fallback to it as a bug,
and issued a fix whose both variants would have removed it from the box's env file.

**What was actually true.** Owner: it is the CORRECT contact address for the tool. The site and
the tool deliberately use different addresses. And because the last EnvironmentFile line wins,
the deployed tool's effective address was correct the entire time — the only defect was an inert
duplicate line, and the right fix deletes the OTHER line.

**What caught it.** The owner, before running the command. Had they pasted my step 3 as written,
a correct production setting would have been silently replaced with a wrong one, under a
walkthrough that said "fix".

**The cheap check that would have caught it.** The claim wasn't derivable from the repo at all —
whether a mailbox is real is owner knowledge. The available check was to mark it: the memory
note that spawned this says "stale … in sites.content_data" — a SITE claim. I widened it to the
tool and stated it as fact. `[ASSUMED — owner to confirm]` at first mention would have surfaced
it a day earlier and kept it out of the fix commands.

**Transferable rule.** A fact about the world outside the system (is this mailbox alive? is this
person reachable? does this company still trade?) can never graduate from assumption to fact by
repetition inside the docs. Scope-widening is the specific trap: a claim true of X (the site)
quietly restated about Y (the tool). Fourth entry in three days where the fix was one marker or
one question at first-write time.

---

## 2026-07-25 — I re-fired yesterday's evidence base without re-checking a line of it (bugs_open/021 INSTANCE 2)

**The claim.** Resubmitting a council submission that had died in validation, I
treated the payload as "the same plan, one field corrected" and re-fired it with
its `rationale` and all six `grounded_in` entries carried over verbatim from the
previous day. Implicitly: *this evidence was true yesterday, so it is true now and
says what it says.*

**What was actually true — two separate defects in that inherited evidence.**

1. **A stale figure.** One entry read *"Live DB 2026-07-24: hardcoded_section_colors
   items — complete:21, unresolved:7, failed:5, detected:2"*. Live at resubmission
   the whole item type was **13 rows** (4/8/0/1). I did not discover this until
   after the submission was in flight, when I finally re-ran the count for a
   different purpose. I still cannot reconcile the two, so the old figure is
   marked `[UNVERIFIED]` rather than corrected — but it went to twelve reviewers
   as a live measurement.
2. **An abbreviated quote that changed the claim.** Another entry quoted the
   detector's SQL as `rendered_html ~ '…' AND rendered_html LIKE '%<style%'`,
   eliding the `AND pc.locked_at IS NULL` line. The `bug_historian` seat read
   that, compared it to the verifier's query (which *does* filter `locked_at`),
   and raised a **MEDIUM** objection about a false-positive hole where a locked
   in-remit component would be silently excluded. The two queries are
   **byte-identical** (`check_hardcoded_section_colors.go:100` and `:214`). There
   is no hole. **We manufactured a medium objection against our own correct code.**

**What caught it.** The objection itself, on a verdict I only read because the
run happened to finish while I was still working. Had it come back an hour later
I would have closed the bug with an unexamined medium objection on the record —
and the stale figure would never have been noticed at all.

**The cheap check that would have caught it.** Re-run every `grounded_in` entry
against source at resubmission time. Both defects were one command each:

```sql
-- (1) the figure, re-measured rather than remembered
SELECT status, count(*) FROM site_work_items
WHERE item_type='hardcoded_section_colors' GROUP BY 1;
```
```
# (2) the quote, diffed against the file it claims to quote
sed -n '95,103p;209,217p' platform/orchestration/actions/discovery_checks/check_hardcoded_section_colors.go
```

**Transferable rule — two, and the second is the new one.**

- *A resubmission is judged standalone, so its evidence must be re-verified
  standalone.* The runbook already says to carry the full evidence base forward;
  it does not say the carried evidence is exempt from being true today. Both
  halves of that are needed.
- **An abbreviated quote is not a shorter quote, it is a DIFFERENT claim.**
  Reviewers cannot open files. An ellipsis inside evidence is an implicit
  assertion that nothing load-bearing was elided — and here the elided line was
  precisely the one the objection turned on. Paste whole predicates, whole `WHERE`
  clauses, whole guard conditions. The plan cap is 64KB and submissions rarely
  reach a tenth of it, so there is no economy being served.

**The sting.** This is the third time in one session I reasoned confidently from
a partial view of the evidence — the `-c 1` entry above, the "someone should build
the queue-depth diagnostic" recommendation inside it, and now this. Twice I was
the one holding the partial view; the third time I handed the partial view to a
reviewer and it failed in exactly the way I had just failed twice. The
distribution is not about carelessness with process — every one of these was a
confident inference from evidence that was *present but incomplete*.

---

## 2026-07-26 — `bugs_open/034`: the drop site everyone reasoned toward was not the one that fires

**The claim.** The 034 handoff (2026-07-20) located a whole failure class at one
site, `platform/agentbase/agent.go:828`, and traced its worked example — a
trigger omitting `client_id` — through that site: *"The error text contains `is
required`, so `agent.go:828` swallows it."* A later thread fixed exactly that
site. I then reasoned from the same starting point, found a **second** classifier
one layer down, and wrote in the bug file that the reachable route to the
agentbase site was `processWithoutContext` — marked `[INFERRED from code read,
not exercised]`.

**What was actually true.** A request with no `client_id` never reaches either
classifier. It is rejected by `ValidateIncomingMessage` at `agent.go:825`, a gate
ahead of both, which publishes an error envelope and returns **without even
incrementing the counter** the later sites increment. A sweep then found a fourth
site (`missing_orchestration_id`) in the same shape. Four sites; the handoff named
one; the fix had landed on one; both of the sites a malformed trigger actually
hits were unfixed and unmentioned.

**What caught it.** Publishing the envelope. The induced fault was expected to
*confirm* the fix and instead landed somewhere nobody had looked — it produced
zero rows anywhere in the database, which is the bug's own signature.

Also, independently and *before* the fault was run: the council gate's
`reuse_agent` seat objected that *"the plan's own search stopped at the two sites
named in the bug file"* and asked for a sweep for further duplicates. It was
right, and it was right from reading the plan alone.

**The cheap check that would have caught it.** One grep, before theorising about
which branch swallows what:

```
grep -n "MessagesDropped\|return$" platform/agentbase/agent.go | head -40
# every early return between message receipt and ProcessMessage is a drop site
```

**Transferable rule.** *When a bug is "there is no durable record", enumerate the
drop sites BEFORE explaining any one of them.* The failure mode of this bug class
is that it hides its own instances — so the site you can see is selected for by
being the one that left a trace, not by being the one that fires. Reasoning
inward from a symptom finds the classifier; reasoning outward from message
receipt finds the gates in front of it.

**Second rule, about markers.** The `[INFERRED]` marker did its job. The claim it
guarded was wrong, a later reader was warned, and the correction cost one
paragraph instead of a handoff. This is the argument for the marker discipline
stated positively — the tally in this file is mostly unmarked claims that went
on to be believed.

---

## 2026-07-25 — "a wrong auto-close costs one re-raise, it cannot lose a finding" (`bugs_open/033`). Caught by the council gate before it shipped.

**The claim.** The whole safety case for letting a sweep auto-close human-review
items: *"Every terminal status is excluded from the `idx_swi_dedup` predicate, so
closing an item RELEASES ITS DEDUP KEY. If a 'resolved' verdict is wrong, the next
run of the check that produced the item raises it again, fresh and correctly
dated. A wrong close costs one re-raise; it cannot lose a finding."* Written into
the action's file header, the commit message, the bug file and the council
submission — in the same voice as the measured claims around it.

**What was actually true.** Half of it. The producers do insert with
`ON CONFLICT DO NOTHING` on a deterministic `item_key`, so a terminal row does
not block a re-raise — that part holds. But *"the next run of the check"*
smuggles in an assumption never checked: **there is no next run on a schedule.**
All three producers fire on a page build or a discovery pass over that site.
A page that is never rebuilt again never re-raises, and a wrong close on such a
page is a permanent silent loss. And the path is almost entirely untested:

```
unresolved_cta          : 70 rows, ALL still needs_human_review — not one has EVER gone terminal
required_fields_missing : 45 parked + 1 complete
needs_section_data      : 45 parked + 7 complete
```

**8 items in the platform's history.** I had asserted a recovery mechanism that
has essentially never run.

**What caught it.** The council gate — `bug_historian`, medium severity, in an
otherwise APPROVED verdict: *"the entire safety case for auto-closing rests on an
unverified assumption … The plan's own risks section admits this … but ships
without confirming it."* It also named the class: a mechanism that silently
discards something while trusting an external invariant nobody verified.

**The cheap check that would have caught it.** Two queries, both under a minute.
*Has anything of this type ever gone terminal?* and *when a key did go terminal,
did it come back?* I ran neither. Worse, my first attempt at the second one —
"do these keys ever recur?" → 0 rows — is **uninterpretable**: nearly every row
of these types is still open, so the dedup index would have suppressed a second
row regardless. Absence of duplicates was never going to be evidence either way,
and I nearly read it as reassurance.

**Why this one is worth the entry even though it was caught.** It was caught by a
reviewer, not by me, and it was the single load-bearing claim in the change — the
one sentence that made auto-closing acceptable at all. Everything around it was
measured and cited: 321 of 370 stale, the leopardess ghost proof, 0 of 45 for the
existing action. **This one had no query and nothing in the prose said so.** That
is exactly the asymmetry the CLAUDE.md marker rule (`[INFERRED]`/`[UNMEASURED]`)
exists to break, and I applied the markers nowhere — least of all to my strongest
claim, which is where they are worth most. A durable claim with no evidence and
no marker, sitting in a paragraph of well-cited ones, inherits their credibility
for free.

**Also worth recording: the fix was fine.** The audit trail
(`result.revalidation` + `resolution_path='auto:revalidated'`) already made every
close individually identifiable and reversible, which holds unconditionally. The
defect was in the *argument*, not the code — and an over-strong argument for a
sound mechanism is still the thing that gets a future thread to skip a check.

---

## 2026-07-26 — bugs_open/029: a residual deferred to another bug on a reason that stopped being true the moment the fix landed

**The claim.** The 07-21 PLAN for the phantom-tool-links fix listed under
*Residuals (deliberately out of scope)*: *"Tool page created (`planned`) but its
content build never deploys → link still 404s. That is `049` mechanism 2
(planned-but-unbuilt page linked), a broader class, not this emitter's defect."*
Written as a scoping decision, in the same confident register as the rest of the
plan, and carried forward into the bug file.

**Why it was wrong.** It was true of the code *as it stood* — an emitter running
at suggestion time has no relationship to any build, so a page that never
deploys is somebody else's class. The whole point of the fix is to move the
emitter **into** the build path. After that move the same residual is this
code's own remaining failure mode, and it reproduces precisely the damage the
bug is about: a live page carrying a reference to a tool page that never goes
live. That is the leopardess 404, arriving by a second route.

**What caught it.** Re-reading the plan while implementing, against a query I
happened to run for a different reason: `needs_content_page` statuses fleet-wide
are **19 needs_human_review / 13 complete / 1 unresolved**. A "residual" that
fires on the majority of tool pages is not a residual. The fix now gates each
emitted item behind the tool page actually going live (`depends_on` the open
build item; no open item → emit nothing).

**The cheap check that would have caught it at planning time.** One question,
asked of every deferral: *after this change lands, does the deferred residual
still belong to the bug I am deferring it to?* A scoping decision is made
against the code as it is and then survives into a world where the code is
different — nothing re-tests it, because a deferral reads like a boundary rather
than a claim. And one query: *how often does the precondition of the residual
actually hold?* Both under a minute.

**The class.** Not a false measurement — a **judgement whose premise the fix
itself invalidates**. Deferrals and "out of scope" notes are where this lives,
because they are written once, in the plan, and inherited by every later reader
as settled.

---

## 2026-07-26 — bugs_open/015: a live probe that "passed" without ever running the code under test

**The claim I was one step from writing.** Closing 015 required inducing the
`retype_existing` refusal branch — proving that a plan naming a page *outside*
the authorising candidate set is refused and mutates nothing. I fired the scratch
one-step probe, polled until `orchestration_states.status` was terminal, got
`current_step=complete, status=COMPLETED`, and queried the fixtures: both pages
still `section-index`, both `sections` still `[]`, `updated_at` unchanged. Every
signal I had said *the fail-closed guard held*. The next line in the bug file
would have been "refusal branch verified live".

**Why it was wrong.** The action never ran. `kcat -P` publishes **one message per
line of stdin**, and I had built `input_data` with a multi-line heredoc, so the
payload was fragmented; `-c 1` then sent only the first fragment — invalid JSON.
The chassis did not fail visibly. It created an `orchestration_states` row under
my `orchestration_id` with `input_data: null`, a fallback no-op `agent_config`
("No-op — scheduled task pre_query already did the work"), **its own**
`orchestration_name` in place of mine, and marked it `COMPLETED`. The fixtures
were untouched because nothing had touched them. A refusal that works and an
action that never executes are **identical** in every field I had looked at.

**What caught it.** Wanting the action's return value verbatim for the bug file,
so I opened `collected_data->'do_retype'` — and it was empty. The step key was
absent entirely, which is impossible if the step had run. Then
`orchestration_name` read `generic-orchestrate-0726-1342` rather than the
`scratch015-refusal-…` I set, which confirmed the message had never parsed.

**The cheap check that would have caught it.** **Grade a probe on what the ACTION
returned — `collected_data-><step name>` — never on the run's terminal status.**
The step key's absence is the discriminator, and it costs one more column in the
poll I was already running. Two supporting checks, equally cheap: assert the
payload is exactly one line before firing (`wc -l` = 0), and assert the
`orchestration_name` that came back is the one you set.

**The class.** A **false green from a harness, not from the system under test** —
the most expensive shape, because the whole point of the probe was to be the
evidence. Note it defeats the standing rule it looks like it should obey: "induce
the fault, don't trust the happy path" was followed exactly — I *did* induce the
refusal branch — and the induction silently didn't happen. So "I tested the
failing branch" is not sufficient; **"I have the failing branch's own output in
front of me" is.** Second-order damage worth knowing: a cleanup keyed on
`orchestration_name LIKE 'scratch…%'` misses the row the fallback left behind,
so clean up by `orchestration_id`. Defences contributed back into the harness
runbook (`durable_write_guard/RUNBOOK_durable_write_guard.md`), which is where
the next session will copy the probe from.

---

## 2026-07-26 — bugs_open/029: I shipped a rollback recipe that would have restored nothing, and it read as reassurance

**The claim.** Migration `211_tool_crosslink_emit_at_build.sql` opened with three
`snapshot_agent(...)` calls and a header that said, in as many words:

```sql
-- ROLLBACK: restore the three rows from the snapshots taken below, e.g.
--   UPDATE agent_definitions a SET default_config = s.default_config
--   FROM agent_definitions s
--   WHERE a.type = s.type AND a.is_active AND s.is_snapshot
--     AND s.snapshot_reason LIKE '211_tool_crosslink_emit_at_build%';
```

Committed, applied, recorded. Two independent defects in four lines:

1. **`snapshot_agent(type, reason)` does not write to `agent_definitions`.** It writes to
   `agent_definitions_backup`. The subquery matches **zero rows**, so the UPDATE is a silent
   no-op — the exact failure a rollback must not have.
2. **`agent_definitions` has no `snapshot_reason` column at all**, so that particular statement
   would have errored rather than no-op'd. (The error is the *lucky* case. A neighbouring query
   I wrote against `agent_definitions WHERE is_snapshot` returned 0 rows cleanly and read as
   *"the safety net was never taken"* — I briefly believed the migration had run without a
   backup.)

**And a third, found while fixing those two:** I applied 211 **twice** (re-ran it to read output
that had scrolled past). So there are two snapshot sets, and the second is the state *after* the
first apply. A rollback taking the newest — which my corrected draft did — restores the
**migrated** config and reports success. It has to be `min(snapshot_taken_at)`, plus a guard that
the chosen set actually contains the thing being rolled back.

**What caught it.** The council gate's `debug_historian` seat, objecting that 211 had
"backup + guard but not the pre-mutation counted-needle assertion or separate verify/rollback
artifacts the lore requires". Writing the separate `_ROLLBACK.sql` it asked for is what made me
*run* the snapshot query. Had the objection not landed, the recipe would have sat in an applied
migration looking like a safety net until someone needed one.

**The cheap check that would have caught it.** Run the rollback's own SELECT — not the UPDATE, the
SELECT inside it — at the moment you write it, and assert the row count. Ten seconds. A rollback
recipe is code, and untested code that only runs in an emergency is the worst kind.

**The class.** A **safety net asserted rather than exercised**. Same shape as the entry above it
(a recovery mechanism nobody had ever observed working), and the tell is identical: prose in the
confident register — "restore the three rows from the snapshots taken below" — with no evidence
the restore had ever been tried. The snapshots were real; everything I said about *reaching* them
was wrong.

**2026-07-26 — shipped a three-kind publish chain that published the same kind
three times, and every step reported success.** The Phase E publisher was
extended to render model → company → protocol, each step configured
`{"kind": "company"}` / `{"kind": "protocol"}`. First live run: both new steps
returned `data/model-directory.json` with `entity_count 39`, and git-adapter
committed MODEL files under the messages "Update adoption tracker" and "Update
protocol tracker". Six steps, six successes, one register published three
times. **Cause was mine and the platform was right:** `ExtractActionInputs`
treats every string config value as a REFERENCE, never a literal — deliberately,
and the reason is spelled out in that file's own Strategy 5 comment citing
`bugs_open/042` (a `max_age_hours: 72` config that agreed with a Go default of
72, so config and behaviour matched and nothing looked wrong until the value was
changed to 720 and the render kept behaving like 72). My literal resolved to
nothing, the field was dropped, and my Go default won in silence. **Cheap check:
I had read that exact comment block earlier the same day** — it is thirty lines
above the code I was calling — and still wrote a string literal into step
config. Reading a warning is not applying it. **The check that actually caught
it: inspecting the committed FILES (`files` keys + `entity_count`) rather than
the step statuses** — the artefact, not the report, for the third time in three
days. Two-line variant worth keeping: when an action takes a selector, assert
the OUTPUT differs per selector; identical output across two "different" runs is
the whole signal. Family: read-the-warning-didn't-apply-it,
status-is-not-artefact, silent-default-wins.

**2026-07-26 — wrote a live-verification command that could only ever return the
passing answer, and it sat in a RUNBOOK for five days as the case's definitive
test.** Closing `bugs_closed/045` (a tool hero hard-wired to a Bayesian ranker),
the runbook's final proof was
`curl -s https://finetuning.uk/tools/ai-agent-roi-estimator/ | grep -ciE '…Bayesian'  # expect 0`.
The URL 404s — the fleet serves `/tools/<name>.html`, not `/tools/<name>/` — and
the 404 body is a 304-byte B2 error JSON containing, naturally, no Bayesian
strings. So the command prints `0` and the check passes **against a page that
does not exist**. Nothing about it was conditional on the fix; it would have
printed `0` before the fix, after it, and against a typo'd hostname.
**What caught it:** curling with `-w '%{http_code}'` out of ordinary caution
while gathering closure evidence, and noticing a 404 next to a "clean" result.
Pure luck of habit — the runbook does not tell you to. **The cheap check:** never
pipe `curl` straight into `grep` for a negative assertion. Fetch once to a file,
then assert **both** the bad string's absence and a positive marker the page must
contain (`data-component="hero-tool"`, the expected `<h1>`); the positive half
fails loudly on every failure mode of the check itself. Guarding the status, or
just recording the expected byte count, kills the same class for free.
**The class — and why it earns a row rather than a shrug:** a negative assertion
is *satisfied by nothing existing*, so every way the check can be broken produces
"pass". This is the same shape as `ON CONFLICT DO NOTHING` returning `err == nil`
while inserting nothing, and "zero rows is not a green light" — three entries now
where the success signal is indistinguishable from having done nothing at all.
The general rule those three want: **when an assertion's pass state is also its
do-nothing state, it needs a positive control or it is not a test.** Family:
vacuous-verification, absence-is-not-evidence, false-green. Pattern in 016b §9.

**2026-07-26 — asserted "the listings demonstrably respect archived" from the two
listing paths I had looked at, and a third did not.** `bugs_open/052`'s containment
section (written 2026-07-20) recommends `pages.status='archived'` as the house route
for a dead page, "which the listings demonstrably respect", citing real evidence: R6
archived six dead article rows and the hub then listed exactly the three real guides.
That evidence is sound and the conclusion still overreached. **"The listings"** was a
claim about every derivation of the page set; what had been demonstrated was two of
them. `rebuild_blog_listing_action.go` selects blog posts with **no `status` filter at
all**, so archiving a post does not delist it there — the containment the file
recommends fails on that path, and fails *silently*, since a listing that keeps showing
an archived page looks exactly like one that was never archived.
**What caught it:** running the file's own `[UNMEASURED]` survey — it had explicitly
flagged that only the tool list was traced — and grepping every query against `pages`
in listing code rather than only the ones the fix touched. Six days after filing.
**The cheap check:** when a containment route is recommended in a durable doc, grep for
**every** consumer of the column it relies on (`grep -rn "page_type = 'blog-post'\|FROM
pages p" --include=*.go`) and confirm each one honours it. Two confirmations do not
make a fleet-wide property; the word "demonstrably" should have been the prompt to
count how many paths were actually demonstrated.
**The class:** this is the quantifier failing quietly. "The listings respect X" and "the
listings I checked respect X" are written identically and read identically, and only the
first is falsifiable by a path you never opened. Same family as the 016b §9 entry on a
sibling call site keeping the defect — but note the asymmetry that makes *this* direction
worse: an over-broad **capability** claim gets caught the first time someone relies on it,
whereas an over-broad claim that a **safeguard works** is only caught when the safeguard
was needed and silently did nothing. Family: unchecked-quantifier, two-instances-is-not-
a-property, containment-assumed-not-verified.

**2026-07-26 — swept a fleet for a rendered element I had only seen on one site, and
read the resulting silence as a clean bill of health.** Closing `bugs_open/053` (an
empty `legal` nav group filling the footer's legal slot with every footer page) meant
checking what each live site actually serves in that slot. I keyed the sweep on
`<div class="footer-legal">`, taking the element from the markup quoted at the top of
the case file — which was robot-hands' markup, one site's component. **idea.uk emits
`<nav class="footer-legal" aria-label="Legal">`.** So idea.uk reported **0 legal links
while holding 1 active legal nav row**, and I carried that for several steps as a live
anomaly that might be a regression *caused by the very fix I was closing* — on a site
belonging to another workstream. Re-measured on the class alone, idea.uk serves
`/privacy.html`, exactly its one row: a **passing** case, and in fact the third
regression guard in the closure's evidence.
**What caught it:** reading the footer component's `html_template` to find out *why*
the block had not rendered, and seeing `{{if .legal_links}}<nav class="footer-legal"` —
i.e. going after the mechanism instead of filing the anomaly. It never reached a
durable doc, but it was three edits away from being written into the case file, 049 and
016b as a fabricated regression.
**The cheap check:** when sweeping rendered HTML across a fleet, **match the class,
never the element** — `grep -A4 'class="footer-legal"'`, not
`grep '<div class="footer-legal">'`. Component libraries share the class contract and
vary the wrapper freely; the element is a property of one template, not of the contract.
**The class:** a filter that is *stricter than the contract* fails silently, and here it
failed in the **reassuring** direction — it under-counted links when "too many links"
was the entire symptom, so the artefact of my own pattern looked exactly like the fix
working. That is the dangerous half. An over-strict filter that under-reports a
**problem** reads as success and stops the investigation; the same error over-reporting
would have been caught immediately by the next check. Sibling of the
`vacuous-verification` entry above — both are tests that pass because they cannot see.
Family: over-strict-pattern, measurement-artefact, silent-under-report,
false-clean-in-the-flattering-direction.

---

**2026-07-26 — concluded a guard had become unnecessary because the failure it guards
against had stopped appearing, and was about to close the case on that basis.** Closing
`bugs_open/061` (the med scraper's LLM fallback inventing price tables) I gathered every
green signal there is: full-table fidelity sweep 0 PRICE_ABSENT across 2,577 rows, the
published data file rebuilt clean, the fix's own counter (`fidelity_skipped`) coming back
in production run results, and — the one that did the damage — **16 consecutive post-fix
fallback calls all returning `[]`**. From that I wrote into the plan that the prompt
hardening had removed the fabrication *at source*, that the guard's drop branch was
therefore belt-and-braces, and that it **"cannot" be induced live**. I put that in writing,
in the plan the owner approved, as a stated limitation of the verification.
**It was wrong on the first attempt.** The induced re-scrape of the original page had the
model invent three complete variants (£19.25 / £34.99 / £68.75, plus a pack size not on the
page); the guard dropped all three and stored nothing. The 16 empty responses were *other
pages* — a sample that never contained the failing case at all.
**What caught it:** running the induced test anyway, purely because the standing rule says
a green happy path proves deployment and not correctness. Nothing else would have; every
other signal I had said "fixed".
**The cheap check:** before concluding that a guard is redundant, **check whether your
sample ever contained the case it guards** — 16 calls on pages that were never the failing
page is not 16 chances for the guard to fire. Restate the claim with the sample named
("no fabrication on *these* pages") and it stops sounding like a finding.
**The class:** absence-of-symptom read as removal-of-cause, where the reassuring reading
also happens to be the one that ends the work. Note the compounding factor — a *layered*
fix makes this worse, because the upstream layers (window gate, prompt hardening) genuinely
do suppress the symptom most of the time, so the downstream guard looks idle precisely
because the layers above it are working. 016b §9 had already written "prompt rules alone
are hope, not enforcement" for this exact bug, and I was one edit from contradicting a
pattern this workstream itself had filed.
Family: absence-as-evidence, unrepresentative-sample, quiet-branch-mistaken-for-dead-branch,
false-clean-in-the-flattering-direction (sibling of the footer-legal entry above).

**2026-07-26 — repeated a sibling doc's six-day-old count into a permanent commit message
and a live council submission, having read the rule against doing exactly that.** Fixing
`bugs_open/052` I quoted its own addendum — *"four fleet pages are `needs_rebuild` AND never
deployed"* — in a Go comment, a commit message and a council `rationale`. The predicate that
figure justifies is correct and unaffected; the figure was not. Re-measured the same
afternoon: **10, not 4**, with `planned`-never-deployed up 18 → 27 over the same six days.
The population had **more than doubled**, which inverts the impression the file's *"Scale —
measured, and smaller than it looks"* section gives, and I had propagated the shrinking
version into the most permanent artefact available.
**What caught it:** deciding, after committing, to re-ground the figure because
`CLAUDE.md` says to — *"Ground every figure against the live system before repeating it from
another doc"*. So the rule worked, but only on a second pass, and only because I went back;
nothing in the first pass flagged it, and the commit message is now unfixable (forward-only).
**The cheap check:** one `GROUP BY build_status, (deployed_at IS NOT NULL)` — about fifteen
seconds — **before** the figure enters a commit message, a council submission, or a code
comment, since those three are exactly the artefacts you cannot re-ground later. A code
comment is the worst host for a live count: it has no measurement date, no way to re-derive,
and it will be read as current for years. State the mechanism there and put the number, with
its date, in the doc.
**The class:** a figure copied between documents silently inherits the **older** measurement
date while acquiring the **newer** document's apparent freshness. Both docs then read as
current and only one was ever measured. Note this is not the usual staleness failure of a
number drifting slightly — the direction of travel had reversed, so the stale figure did not
merely understate, it argued the opposite case. Family: figure-inherits-its-source-date,
stale-premise-carried-forward, unre-groundable-host.

**2026-07-26 — was one edit from "fixing" a DNS problem in the direction that makes it
worse, and one commit from shipping a metric into a void and reading its silence as
success** (bugfix_040_kafka_dial).

Two near-misses on one bug, both caught by a check rather than by noticing.

**(a) The knob that goes the wrong way.** `bugs_open/040-kafka-dial`'s brokers advertise
a **three-dot** name (`...kafka.svc`, no `.cluster.local`). Pods run `ndots:5`, so every
Kafka connection walks three NXDOMAIN rounds before the one that works — measured at
**73% of all cluster DNS (384,392 of 525,152 responses in 24h)**. The obvious fix is
"ndots:5 is too high, lower it", and I had it in a draft plan. It is **strictly worse**:
the name genuinely needs `cluster.local` appended, so at `ndots:2` the resolver tries it
absolute (NXDOMAIN) **and then still** walks all three search domains — **four rounds
instead of three**.
**What caught it:** writing out the resolver's actual attempt order for that specific
name, rather than reasoning from what `ndots` means in general.
**The cheap check:** enumerate the literal sequence of lookups for the exact input, before
and after, and count them. Fifteen seconds, and it inverts the answer.
**The class:** a tuning knob whose semantics are symmetric ("try absolute vs try search
first") but whose *cost* is asymmetric, because one branch can succeed and the other
cannot. Reasoning about the knob's meaning is not reasoning about the outcome. Family:
plausible-inversion, mechanism-understood-outcome-not-computed.

**(b) The metric that would have gone nowhere.** The whole point of the change was to make
040 measurable: add `ai_persona_kafka_dial_total`, then read it. Before committing I
checked the counter would actually be *collected*. It would not: nothing in the fleet had
ever served `/metrics` — `observability.NewMetricsServer` had **zero callers**,
`cmd/agent-chassis/main.go` built a mux with only `/health` and `/ready`, and the live
Prometheus held **zero `ai_persona_*` series** despite every spawned pod being annotated
`prometheus.io/port: "9090"` + `path: /metrics`. Dozens of actively-maintained counters,
dead since written.
Had I not checked, the sequence was: ship the metric → roll → query it → read zero →
report the flake rate as nil. **Every step of that is what "verified with a metric" looks
like**, and the conclusion would have been fabricated.
**What caught it:** my own plan's "positive control before trusting a zero" step, written
because of the `verify-the-failing-branch` rule. It earned its place on first use.
**The cheap check:** ask the monitoring server for the whole metric namespace (not for your
new series) before believing any reading from it; and put a test in the suite that induces
the failure and asserts the counter moves, reading it back through
`prometheus.DefaultGatherer.Gather()` — the registry the HTTP handler serves — not off the
collector, which passes even when nothing is reachable.
**The class:** sibling of *"pair a negative assertion with a positive control"* one row up,
but the failure is a whole level higher — there the fetch was real and the string absent;
here the **channel itself** was never connected, so no assertion made through it could ever
have been false. An unscraped metric, an unwired counter, and a genuinely fixed bug are
byte-identical, and the metric is the *most* convincing of the three. Family:
absence-as-evidence, unproven-channel, unfalsifiable-green.

**Note the shape both share with the dominant class above:** neither was a process lapse.
Both were mechanisms I believed I understood well enough not to compute. The tally rows are
"compute the specific case" and "prove the channel", not "be more careful".

---

## 2026-07-26 — I re-fired a queued dispatch twice, having the rule in memory (bugs_open/049 session)

**The claim I acted on:** "no `orchestration_states` row after several minutes = the message
was dropped, fire it again." I fired the same gaswholesalers chrome refresh **three times**
(14:58, ~15:14, ~15:20).

**What was actually true:** all three were **queued and none was lost.**
`system.agent.generic.requests` has **one partition**, the chassis drains it one job at a
time, and three long council runs were in flight ahead of me — one of them my own council
submission, fired 13 minutes after my first chrome refresh. `scripts/dispatch-queue-depth.sh`
said so in as many words: `QUEUE DEPTH (LAG): 5 … It is QUEUED, NOT LOST: an absent
orchestration_states row means 'not started yet'. DO NOT re-fire — a duplicate spends the
same LLM credits and lands even further back in this same lane.`

**What caught it:** running that script — which exists *because of this exact failure* and
prints the instruction I had just disobeyed, twice.

**The cheap check:** run `scripts/dispatch-queue-depth.sh` **before** the second fire. It
takes seconds, needs no reasoning, and answers the only question at issue.

**Why this row is worth writing even though the rule was already written down.** It was not
merely written down somewhere — it was in my own auto-memory
(`council-queue-latency-trap`: *"no orchestration rows = QUEUED (~16-30 min), not dropped —
don't resubmit"*), which loads every thread, and it is in CLAUDE.md's council section. I had
read both. What defeated them was a **plausible competing explanation**: postgres was in a
liveness-probe restart loop and the chassis had rolled a new image 17 minutes earlier, and
CLAUDE.md separately says *"No orchestration dispatch within ~300s of a chassis pod (re)start
— the spawn is silently dropped."* So I had a specific, documented, genuinely-occurring
mechanism that predicted exactly the symptom I saw. It was not the cause.

**The class, and it is the interesting part:** the trap is not "I forgot the rule". It is
that **two documented mechanisms produce an identical observation — an absent orchestration
row — and only one is distinguishable, cheaply, by a tool that already exists.** Real
corroborating evidence for the wrong theory (the pod really had just rolled; postgres really
was flapping) makes the wrong theory *more* attractive, not less. Confidence rose with each
piece of true-but-irrelevant evidence.
The rule that generalises is not "remember the landmine" but **"when two known mechanisms
predict the same observation, run the discriminator before acting — never pick by
plausibility."** Family: identical-observation-two-causes, corroboration-of-the-wrong-theory,
tool-exists-and-was-not-run.

**The cost was real, not notional:** three chrome refreshes on a live customer site instead
of one, each fanning out to ~31 `page_rerender` items with a git commit, B2 sync and
Cloudflare purge apiece — and the trigger script's own header warns "Three at once floods the
queue." I did to one site exactly what it tells you not to do across three.

## 2026-07-26 — gauntlet_dead_cta (P4): "the handoff says the CTA is broken, so it's broken"

**The claim:** the handoff and the PLAN both had Step 1 as "fix the homepage
provocation-card CTAs", with a whole risk paragraph about the homepage being
`rebuild_policy='generic'` and the danger of a concurrent rebuild clobbering the
edit. I was one command away from editing the homepage on that basis.

**What was actually true:** `today.primary_cta.url` was already
`/tools/gauntlet/index.html`, and `provocation-card-loader` already sets both
hrefs from the feed at runtime. There was nothing to fix and no reason to touch
the homepage at all. The open `cta_names_unknown_destination` work item that
started it is a scan of the *static shell* of a component that is filled at
runtime — it never saw the real hrefs.

**What caught it:** re-reading the live feed and the loader's source before
building, rather than trusting the handoff's summary of them. Two more of the
same session's premises fell the same way: the "Enter today's Arena" CTA that a
council advisory asked us to re-point already pointed at the right page, and
`tool-arena-interface` turned out to have a 38 KB template with `js_content IS
NULL` rather than the unknown-source risk the plan had gated on.

**The cheap check that would have:** one `curl` of the feed and one `SELECT` of
the loader — under a minute, against a document that had been written 90 minutes
earlier by a thread with full context.

**The transferable bit:** a handoff's *identity* section (ids, rows, paths) is
usually still true; its *diagnosis* section decays fastest, because it was
written before anyone re-checked the thing it describes. A work item is evidence
that something was flagged, not evidence that it is still broken — and a
detector that reads static markup cannot see a runtime-filled component. Verify
the defect before you plan around it, especially when planning around it is what
makes the task risky.

**2026-07-26 — wrote the counter-argument to my own change, put it in the risks section,
and shipped it anyway; the council caught what I had already spotted**
(bugfix_040_kafka_dial, council corr `7abe1a57`, guardian hard veto, 2× HIGH).

Instrumenting `bugs_open/040-kafka-dial` I also unified the fleet's four divergent Kafka
dial configurations, cut the default dial timeout **10s → 5s fleet-wide**, and raised the
producer's `IdleTimeout` 30s → 5m and `MetadataTTL` 6s → 30s. The guardian vetoed it:
behaviour changes to shared messaging plumbing and to failover reactivity across every
pipeline, bundled into what I had framed as a metrics change.

**The damning part is that I had already made the argument against it, in writing, in the
submission itself.** My own `risks` field said the timeout cut *"makes it fail sooner
rather than hang… if the residual flake turns out to be a genuine multi-second-but-
recoverable path"*. I wrote that, called it a trade-off, and submitted. My justifications
were a **remark in a bug file** (§4.6 "10s is pathological") and **"the Java client's
default is 300s"** — an opinion and an appeal to another product's defaults. Neither is a
measurement, and the whole point of the change was that **nothing about this path had ever
been measured**. I was tuning a constant blind while building the instrument that would
have told me what to set it to.

**What caught it:** the council gate, on a submission I nearly did not send because the
change "was only instrumentation".
**The cheap check:** read your own `risks` section back before submitting and ask which
entries describe *this* change rather than the world. A risk that says "this change could
make X worse" is not a disclosure — it is the reviewer's objection, pre-written. Split
that part out. A second, near-free check: for every constant you are changing, name the
measurement that chose the new value. "A doc said so" and "another library does" are both
absent measurements wearing a citation.
**The class:** scope creep laundered through a legitimate change. The instrumentation was
sound and needed, which made the behaviour changes riding along feel like part of the same
tidy-up. Note the ordering error too — the change that would *produce* the evidence and the
change that *consumes* it went in together, so the tuning could not possibly be evidence-
based. **Sequencing was available and free**: ship the counters, read the histogram, then
choose the timeout. The veto did not cost a round, it improved the change.
Family: behaviour-change-bundled-into-instrumentation, constant-tuned-without-measurement,
self-refuting-risk-section, evidence-and-consumer-shipped-together.

**Two smaller ones from the same review, both about the submission rather than the code:**
`editquality` flagged `consumer.go` as un-rewired — it *was* rewired in the commit, but I
hit the 8-edit cap and dropped it from the plan while still quoting its pre-change state in
`grounded_in`. And `reuse_agent` objected that I bypassed `observability.NewMetricsServer`
having cited its zero callers as the defect, without saying why. Both correct **from what
the reviewers could see**. The standing lesson (already a row above, about trimmed quotes)
generalises: **the submission is the artefact under review, not the commit.** An edit you
leave out reads as an edit you did not make, and a reason you did not write down does not
exist. The cap is not an excuse — if the plan will not fit in 8 edits, that is information
about the change's scope, which in this case it was.

**2026-07-25 — wrote an em dash into the rule forbidding em dashes, and shipped it
live before catching it.** Refining the `page-content-writer` voice prompt, I added
a worked example teaching the model never to use `—`, and ended my own inserted
paragraph with one: *"Hunt for this appositive shape specifically — it slips
through…"*. It was written to the live agent config before I read it back. Trivial
to fix, and worth logging because of what it revealed one query later: **the prompt
already contained 17 em dashes**, 14 of them in its own instructional prose
including the `## Voice & Style (how the copy must READ — follow strictly)`
heading. So two prior refinements had been telling the model to avoid a character
while modelling it fourteen times in the most authoritative text in its context —
which is the actual reason both failed, and I had spent two rounds blaming the
model. **Cheap check: after editing an instruction, grep the whole instruction for
the thing it prohibits.** One `count(regexp_matches(prompt,'—','g'))` located a
cause that two rounds of reasoning had missed. Family:
instruction-violates-its-own-rule, check-the-example-against-its-constraint.

**2026-07-25 — filtered a "does this page exist" lookup on `build_status`, and
mislabelled a live page as invented.** My internal-link census resolved each href
against `pages` with `AND p2.build_status='deployed'`. `/contact` therefore came
back as an **invented target** — in the same report as nine genuinely invented ones
— while `/contact.html` was serving **200**, because that page's row read
`needs_rebuild` while its artefact was live. I had already recorded the general form
of this trap in `016b` §9 (*"a liveness filter keyed on the PAGE-level build_status
misses live-serving pages whose flag has drifted"*, 2026-07-22) and reintroduced it
three days later. **Cheap check: when the question is "does this URL exist", the
answer is a row's existence, not its build state — and the live probe settles it in
one `curl`.** Cost: nearly repointing a working link, and one wrong row in a report
handed to the owner. Family: build_status-is-not-liveness,
column-does-not-mean-what-you-are-measuring, re-read-your-own-§9.

**2026-07-26 — fixed the phantom broker at the site I tripped over, never grepped for the
second one; and "reverting a behaviour change" nearly became its own behaviour change**
(bugfix_040_kafka_dial, council round 2, REVISE).

**(a) The sibling I did not enumerate.** `topic_manager.go`'s fallback broker list
contained `kafka-0.kafka-headless.kafka:9092` — a host that cannot resolve, so it could
only ever burn a full dial timeout before failing over. I removed it and moved on. The
identical entry was also in `spawn_actions.go:1019`, the fallback list **every spawned
agent inherits**, and I never looked. The council's editquality seat found it by pushing
on a different objection entirely.
**What caught it:** the gate, indirectly — nothing in my own process did.
**The cheap check:** `grep -rn "<the exact bad literal>"` before writing the fix comment.
One command. The literal string was right there in the line I was deleting.
**The class:** this file already has a tally row for it — *"enumerate the SIBLING instances
before quantifying"*, previously logged twice. This is the third. The row is not landing
because it reads as being about *counting* ("before quantifying"); the failure here was not
a miscount but a **fix that silently declared itself complete**. A defect found by reading
one file is a hypothesis about a pattern, and the grep that tests it costs nothing.
Family: fix-the-instance-not-the-class, unenumerated-siblings.

**(b) The mirrored error, caught one step before committing.** Round 2 correctly flagged
that I had threaded the caller's `ctx` into `topic_manager`'s eight dial sites while
claiming the change altered no dial behaviour — six of them previously used bare
`kafka.Dial` (i.e. `context.Background()`), so a caller holding a sub-10s deadline would
silently have got a shorter dial. Fair. My fix was a **blanket** replace of `ctx` →
`context.Background()` across all eight. But two of them (`WaitForTopic`,
`WaitForTopicOld`) had **always** used `ctx` — so the blanket revert would have introduced
the same class of change in the opposite direction, inside the commit whose entire purpose
was removing it.
**What caught it:** diffing each site against the pre-change blob instead of trusting the
replace — `git show <base>:<file> | grep -n "kafka.Dial"` showed six `Dial(` and two
`DialContext(ctx`.
**The cheap check:** when reverting to "how it was", the baseline is **per-site**, not
per-file. A global find-and-replace assumes uniformity that the original did not have.
**The class:** a correction applied at coarser granularity than the thing it corrects.
Note the shape — the mistake and its fix were the *same* mistake, and being mid-way through
writing a careful revert is exactly when it feels safe to sweep. Family:
blanket-revert, granularity-mismatch, correction-reintroduces-the-defect.

**2026-07-26 — the enumeration that found the missing sibling was ITSELF truncated, and I
read the truncation as the answer** (bugfix_040_kafka_dial; immediately after logging the
entry above about not enumerating siblings).

Having been caught by the council for fixing a phantom broker at one site and not grepping
for the others, I ran the enumeration — and piped it to `head`:

```bash
grep -rn "kafka-headless" --include=*.go --include=*.yaml . | head
```

Ten lines came back. Eight were `deployments/` YAML, two were the Go sites I already knew
about. I wrote "two sites" into a commit message, a bug file and a council submission.
There was a **third**: `internal/core-manager/admin/agent_handlers.go:766`, cut off by
`head`. And it was the worst of the three — a fallback of *three* nonexistent brokers with
no valid entry at all, so with `KAFKA_BROKERS` unset core-manager had no route to Kafka
whatsoever, only three consecutive dial timeouts. It also read the wrong env var, making it
*more* likely than the other two to reach that dead list.

**What caught it:** re-running the same grep without `| head` during a final verification
pass, for no reason other than wanting the full output on screen.
**The cheap check:** when a grep is establishing *completeness* — "these are all the
sites", "nothing else references this" — never pipe it to `head`. If the output is too long
to read, that is the finding. Use `| wc -l` first, or `sort -u`, but do not cap it. A
truncated enumeration and a complete one are visually identical, and `head` gives no signal
that it dropped anything.
**The class:** the same defect as the entry above it, one meta-level up — a **check that
silently declares its own scope complete**. I was actively fixing "a fix that didn't
enumerate its siblings" and the tool I reached for to do the enumerating had a silent cap
in it. Note this is the *inverse* of the usual truncation trap in this file (an artefact
truncated on write): here the artefact was fine and the **measurement** was truncated, so
nothing downstream could have caught it — no length check, no structural check, only
re-running it. Family: silent-cap, truncated-enumeration-read-as-complete,
fix-the-instance-not-the-class.

**2026-07-26 — wrote a post-roll verification into a bug file's close-out checklist using a
string my change does not create. Second vacuous-verification incident logged today, by a
different thread, from a different angle — that pairing is the entry's whole point.**
Closing out `bugs_open/052` I specified `strings /app/agent-chassis | grep -c "p.status IN
('active', 'deployed')"` as the proof the fix had shipped. That fragment already occurs **4
times** in `v1.0.1167`, an image built *before* the fix was committed — other queries use it.
The check passes unconditionally, forever.
**What caught it:** running it against the current pod while checking something else, and
getting `4` where `0` was expected. Not review — accident, again.
**The cheap check:** run the pod-grep against the CURRENT (pre-fix) binary before writing it
down. Expected-fail now, expected-pass later; if it passes today it is not a test. A
second-order trap specific to SQL: `strings` splits on newlines, so a multi-line Go SQL const
is *not* one searchable blob — you can only match a single line, which is why the durable
marker here had to be the **disappearance** of the old predicate line plus a positive control
that the query still exists at all.
**Why it earns a row next to the 045 entry above rather than folding into it:** that one was a
*negative* assertion passing on a 404; this one is a *positive* assertion passing on a
pre-existing string. Opposite polarity, same root — the check's pass state was never
distinguishable from its do-nothing state, and in both cases habit caught it, not process.
Two independent instances in one day is the argument for automating it: a pre-commit that
extracts pod-grep markers from changed docs and asserts they currently return zero is
mechanical, and the tally row below is now at 2. Family: vacuous-verification,
false-green, marker-not-created-by-the-change.

**2026-07-26 — treated a vanished council dispatch as latency for 62 minutes because the
standing rule says to.** The house rule (`[[council-queue-latency-trap]]`, and `CLAUDE.md`)
is that a missing `orchestration_states` row means queued, ~16–30 min, and that resubmitting
on that evidence wastes a round. It is a good default and it was wrong here: the submission
left **no trace anywhere**, two *newer* submissions overtook and completed it, the consumer
was demonstrably live, and the chassis had **rolled 8 minutes before the publish**. The rule
protects against resubmitting too eagerly; it offers nothing for deciding when waiting has
become the mistake.
**What caught it:** the monitor timing out at an hour and forcing a second look — otherwise I
would still be waiting.
**The cheap check:** before concluding "queued", ask whether anything *newer* has finished.
`SELECT min(created_at), count(*) FROM orchestration_states WHERE workflow_plan::text ILIKE
'%council_decide%' GROUP BY fix_correlation_id` answers it in one query: a later submission
that completed while yours has no row at all is not a queue, and the pod's `startTime`
compared against your publish time names the likely cause.
**The class:** a rule of the form "absence means X, don't act" has no expiry, so it silently
converts into "wait forever". Any such rule needs a stated bound and a distinguishing test for
the other branch — the useful version is *"absence means queued **unless** something newer has
drained past you"*. Family: default-with-no-exit-condition, absence-is-ambiguous,
rule-outlived-its-window.

### 2026-07-26 — webdesign.co.uk — "98 page_rerender items will never dispatch; page-rerender is broken"
**Asserted:** that the `page-rerender` agent was not receiving spawns — `bugs_open/003`
class — because 98 items sat `triaged` with `attempt_count=0` while `needs_imagery` items on
the same site kept completing. Written into a commit message and a NOTES section as a
platform blocker.
**Actually:** the queue was working normally. First claim came **20m40s** after the items
were created (the documented publish→run-start latency); all 98 completed over 3h28m at
~2.1 min each, which is simply what single-flight-per-site costs. By morning every page was
live. The "concurrent" imagery was nothing of the kind: every imagery item I watched finish
had been **claimed at 17:04–17:16, before the page items existed at 17:12**, and imagery then
stopped dead at 17:18 and did not resume until **20:40:38 — 18 seconds after the last page
completed**. My own priority change (`page_rerender`→5, imagery→90) had pre-empted imagery
perfectly for three and a half hours. I read the starvation I had just caused as a fault.
**Caught by:** walking away. The queue drained overnight and the site was fully live — luck
of the calendar, not a check. I gave up **8 minutes** before the first claim.
**The cheap check that would have caught it:** `SELECT item_type, min(created_at),
min(claimed_at), max(completed_at) FROM site_work_items WHERE site_id=… GROUP BY 1` — items
claimed *before yours existed* say nothing about yours. I was reading counts change, never
timestamps. And `attempt_count=0` means *not yet tried*, indistinguishable from *never will
be* until the documented window has elapsed.
**Cost:** one duplicate direct dispatch (harmless — the importer keys on
`(page_id, slot_name)`), a false blocker in a commit message and a handoff, and a day of the
owner believing the site was stuck. Note this sits near the entry above from the *opposite*
direction — waiting 62 minutes because "absence means queued". Same ambiguity, opposite
conclusions, same unrun timestamp comparison.

**2026-07-26 — wrote a lockstep test that could not fail, and nearly shipped it as the
guard's proof.** Fixing `bugs_open/076` I added `truncationAwareActions`, a registry naming
each action whose code reads the `__truncated` marker, plus
`TestTruncationAwareActionsReadTheMarker` to hold registry and code in lockstep: it scans the
package for a non-test file that both names a registered action and reads `"__truncated"`.
The test passed. It would have passed for **any** name, forever — `truncation_guard.go`, the
file holding the registry, names `__truncated` in its own doc comment, so every entry
satisfied the check simply by being declared. The test asserted that I had written the
registry, not that the guard existed.
**What caught it:** running a falsification probe instead of trusting green — inserting a
deliberately bogus `"render_page_html"` entry and requiring the test to FAIL. It passed, which
is the whole finding. Excluding the registry's own file from the scan made it fail with the
right message; restoring the real registry made it pass again.
**The cheap check:** after writing any check, break the thing it checks and watch it fail.
Ten seconds. Without it a green test is evidence of nothing — and this one would have been
cited in a bug file as proof the guard was held in place.
**The class:** the checker and the checked sharing ground. Already recorded twice in the
brochure-component lane as "a check sharing the fix's regex can't falsify it"; this is the
third instance and the first where the shared ground was a **doc comment** rather than a
regex, which is why it did not look like the known pattern. The general form: a check that
reads the same file it validates will validate itself. Family: vacuous-assertion,
checker-shares-ground-with-checked, green-without-a-probe.

**2026-07-26 — re-read a bug file's contents from context and never checked it was still
where I left it.** `bugs_open/064` (the doc-subject split contract) was in my context from a
session summary, and I wrote a full update into it: pod-grep evidence from today's v1.0.1167
binary, plus "still not closed — the failing branch needs one live run", plus a plan to fold
that run into a later change-set. All of it framed around a file that had been **closed and
moved to `bugs_closed/` the previous day** (commit `eb81de7b5`), after the closing session
proved both the accepting and the rejecting branch with live orchestration runs on v1.0.1156.
My "still open" claim was a day stale and would have appeared in a commit and in a plan
update.
**What caught it:** the Edit tool refused the path — file does not exist. Nothing I reasoned
about caught it.
**The cheap check:** `ls bugs_open/ bugs_closed/ | grep <number>` before writing to a bug
file — the standing "grep BOTH directories" rule, which I had followed when *filing* 064 and
skipped when *updating* it. Half a second.
**The class:** context is a snapshot of contents, not of location or status. A file's body can
be word-perfect in context while its *state* — which directory it lives in, whether it is
open, who owns it — has moved on; and state is exactly what a bug file's directory encodes
here. Adjacent to "a handoff's identity survives, its diagnosis decays" (2026-07-26, gauntlet)
and to the who-owns rule for existing bugs. Family: stale-context, status-not-in-the-body,
snapshot-mistaken-for-live.

**2026-07-26 — a mechanism stated as fact, inferred from a definition and an outcome, that
one query refuted.** `bugs_open/068` (filed 07-24, not by me) said an extraction-time contract
violation "bypasses the step's `error_step`" and "never reaches that routing", and built a
whole fix-candidate C around routing extraction failures through `error_step`. Both halves are
false: the coordinator routes extraction failures like any other step failure
(`coordinator.go:869`), and the step's `error_step` was declared correctly. What was missing
was the field — `convertToWorkflowPlan` builds `models.Step` field by field and never copied
`error_step`, so **no persisted plan has ever carried one**. Candidate C would have "fixed" a
path that already worked, on a fleet-wide class that would have stayed broken.
**What caught it:** reading `orchestration_states.workflow_plan` instead of
`agent_definitions` — one `jsonb_each` over three days of plans: 0 of 14,209 steps carry a
step-level `error_step`, 1,828 carry the `config.error_step` twin. The 07-24 claim was
consistent with the definition and with the observed fatality; it was never checked against
the artefact the runtime consumes.
**The cheap check:** when a config field appears not to take effect, diff the **materialised**
artefact against the definition before theorising about the code path that reads it. One query.
**A second one in the same file, same habit:** "this generation of the writer selects its own
sections" — inferred from a step named `select_sections`, which is an `extract_fields` over two
sources and selects nothing. That inference is what got fix candidate A rejected; A was right
all along (`bugs_open/087`).
**The class:** a claim about runtime behaviour derived from config plus outcome, written in the
same voice as a measurement, with a step's *name* doing the work of its config. Family:
inference-in-the-voice-of-a-finding, read-the-artefact-not-the-source, name-is-not-a-spec.

**2026-07-26 — "this ticket is unowned", written into an approved plan while another session
was fixing it.** I picked up `bugs_open/078` and ran `scripts/who-owns.py 078` first, as the
rule says. It returned **OVERALL: OWNED or recently active** — and I reasoned past it: the only
commits were the filing session's, and that session had closed, so I recorded "no owning
workstream, the advisory is stale" and planned a full fix. Ninety minutes later, at the moment
I started implementing, `git log` showed `912ddc1db fix(bugs_closed/078)` — a different session
had diagnosed, fixed, applied migration `217`, verified and closed it while I was planning.
Nothing was lost (the re-check caught it before a line was written, and I converted the work
into an independent verification plus a residual for `bugs_open/033`), but a full plan cycle
was spent on a ticket that was already being fixed.
**What caught it:** re-running `git log` at implementation start — CLAUDE.md's "your
session-start `git status` is a snapshot; re-run it before acting on it", applied to the log.
**The cheap check I actually skipped, and it was in front of me:** at 17:42 I measured the
offending NULL-handler row live; at 17:48 I re-queried and it had been **repaired**. I noticed,
looked up what happened to it, wrote "someone repaired it" — and did not draw the one inference
that mattered: *someone is working this ticket, right now*. A live artefact changing underneath
you mid-investigation is an ownership signal, and **the thing that changed names the ticket
they are on**. `who-owns.py` cannot supply that: it reads commits, and an in-flight session has
not committed yet. So the advisory was not wrong when it said OWNED — it was merely *early*,
and I overrode it on the one piece of evidence it could not have.
**The class:** treating a stale-looking advisory as a false positive, when the live system was
concurrently emitting the fresh evidence for it. Adjacent to "check whether an existing bug has
an owning workstream" (existing row) but distinct: that one is a skipped check, this one is a
check that ran, said the right thing, and was argued away. Family: advisory-overridden-by-
reasoning, concurrency-invisible-to-git, the-artefact-that-moved-was-the-signal.
**Worth noting for the next thread:** the same ticket was worked twice by unrelated sessions in
24 hours (07-25 filing, 07-26 fix) and its *symptom* was repaired by hand three times by three
different sessions. A hot bug attracts concurrent threads precisely because it is biting
everyone at once — which is exactly when `who-owns` is least able to see them.

---

## 2026-07-26 — I wrote two confident causes for a vanished dispatch, and both were refuted within twenty minutes (bugs_open/076)

**The claim(s), in order.** Two kcat dispatches for the 076 induced-fault probe never became
orchestrations. I wrote into the bug file, as fact, that the cause was a UUID `client_id` where
the working trigger uses `demo_client`. Then, when that failed, I wrote that it was the ~300s
post-restart drop rule.

**What caught them.** The first: the successful re-fire used **the same UUID `client_id`** — it
was in my own shell history, in the command I had just run. The second: the third probe was
fired at **T+100s** on a freshly rolled pod, comfortably inside the window I had blamed, and it
landed in seconds.

**The cheap check that would have caught both:** *does my explanation also explain the case that
WORKED?* Both stories were built by staring at the failure and ignoring the success sitting
beside it. A cause that does not discriminate between the failing run and the passing one is not
a cause — it is a coincidence with a narrative. This is the same discipline as the positive
control I had *just* insisted on for the guard itself (an unguarded probe that fails proves
nothing without a guarded probe that passes) — I applied it rigorously to the code under test
and not at all to my own diagnosis in the same hour.

**The class:** explaining a failure without testing the explanation against the neighbouring
success. Also: writing a mechanism into a durable doc while it was still a hypothesis, with no
`[UNVERIFIED]` marker — which is precisely what CLAUDE.md's "mark the UNVERIFIED ones too" rule
exists to stop, and typing the marker would have been enough to make me go and check.

**What it cost, and what it did not.** Nothing but the correction, because the claims were about
my own test harness rather than about the platform. Had either reached a handoff as "how
dispatch works", it would have cost every thread that then believed it. The file now records
both refutations and asserts only the **exit test**, which does not depend on knowing the cause:
*has anything newer drained past me?* If newer orchestrations are completing while yours is
absent, it was dropped — re-fire rather than wait.

**A THIRD wrong call on the same question, twenty minutes later.** Having been refuted twice
on *why* a dispatch vanished, I then concluded — via `bugs_open/052`'s own exit test — that a
council resubmission had been **dropped**, and re-fired it twice. It had not been dropped. It
landed **5m37s** after publication. Two ways I misapplied a sound test: the comparator was on
the **cron lane** (`bugs_closed/030` gave those their own topic, so they drain regardless), and
then it was a run **published before mine** — on a FIFO queue, "an older job is progressing" is
entirely compatible with "mine is still waiting". The comparator must be on your lane AND
published after you.

**Tally note.** This is the third entry in the "wrote a mechanism before checking it"
family, and the second in two days where the missing check was *test the story against the
case that worked*. Within this single entry the same error recurred three times in under
half an hour, each time with a *different* tidy explanation — which is the actual finding.
The failure mode is not any one bad theory; it is **theorising at all about a system whose
normal latency I had not measured first**. Each guess cost a duplicate dispatch; waiting
~6 minutes would have cost nothing and answered it.

**2026-07-26 — wrote "inert until an image roll" into a commit message, a bug file and a council
submission, about code that shipped twenty minutes later.** Fixing `bugs_open/086` I committed a
converter change at 18:02Z and described it, three times, as inert until the owner chose to roll.
Another thread built and rolled the chassis at ~18:22Z; `make build` takes committed HEAD, so my
change rode their build into v1.0.1169. The whole risk conversation with the owner — a
council-vetoed change activating 55 dormant error handlers, ten of which turn a failure into a
green run — was conducted on the premise that there was a roll's worth of time in hand to audit
them. There wasn't. The audit happened after they were already armed.
**What caught it:** a sibling commit in `git log` mentioning "live in v1.0.1169" while I was
finishing the docs, then a pod-grep for a string my change creates (present, with a positive
control). Nothing in my own process would have caught it — I had checked the pod at the START of
the session and never again.
**The cheap check:** pod-grep for your own new string **after** committing platform code, not only
before; and treat "inert" as a claim with an expiry, not a property. `kubectl exec <chassis pod> --
strings /app/agent-chassis | grep -c "<a string only my change creates>"`.
**The class:** this is a KNOWN landmine in this repo — CLAUDE.md says builds take committed HEAD,
and a memory entry from the relojistas lane literally reads "committed code rides ANYONE's next
build". I repeated it anyway, because the *timing* was in another session's hands and my mental
model had a single owner-controlled roll in it. The general form: a safety property that depends on
someone else's future action, asserted as if it were mine to hold. Family: shared-tree-timing,
inert-until-someone-else-decides, known-landmine-repeated.

### 2026-07-26 — webdesign.co.uk — "no point in the pipeline asserts that a page's JavaScript works"
**Asserted:** filed as `bugs_open/084`, claiming the platform has only presence checks and
no integrity checks for JavaScript, and recommending a live script-integrity sweep be built.
**Actually:** there is a four-tier verification ladder whose Tier 4 drives the deployed page
in **real headless Chromium** — `internal/adapters/browserrunner/run_checks_action.go`,
playwright-go, real `fill`/`click`/`select`, post-interaction DOM assertions, `console.error`
and uncaught page errors via `OnConsole`/`OnPageError`, desktop + mobile profiles, failure
screenshots. **Live in production at v1.0.1167**, made continuous by `tool_acceptance_due`,
and documented under a heading reading *"Does it actually work in a browser?"*. Plus live
`dead_controls`, `truncated_component` and `tool_health` checks.
**Caught by:** the owner asking "consult the docs for where we have built mechanisms to check
whether the JS functions — how does this compare to your diagnosis?" One question. No check
of mine would have surfaced it.
**The cheap check that would have caught it:** the one CLAUDE.md already mandates — *grep
before you file*. `grep -ri "dead_control\|browser-runner\|acceptance" docs/ platform/`
returns the ladder in seconds. I had even read a memory line naming the dead-controls
detector as live, and filed anyway.
**Cost:** would have sent someone to build a headless-browser tier that has been running for
weeks. Rewritten and renamed the same day. **The transferable mechanism:** I generalised from
*my* population (owned, ported, `component_level='section'` pages, which genuinely get no
browser run because Tier 4 gates on `'tool'`) to *the platform*. A coverage boundary and an
absent capability read identically in a bug file and demand completely different work.

---

## 2026-07-26 — I read a probe's output as "4 of my tests are vacuous" when 4 of them had never run

**The claim:** while fixing `bugs_open/079` I ran the induced-fault probe on my own new
tests — stub `RepairPageLinks` to a no-op, expect every repair assertion to fail. It reported
**4 failures out of 8** repair tests. I wrote that down as "4 tests are vacuous" and started
hunting for what was weak about them.

**Why it was wrong:** those four tests never executed. An unguarded `repairs[0]` on an empty
slice **panicked the test binary**, and every test declared after it in the file was skipped
in silence. `go test` without `-v` prints nothing for a passing test and nothing for a test
that never ran — the two are indistinguishable in the default output, so a run that did a
third of the work looked exactly like a run that did all of it.

**What caught it:** distrusting my own reading enough to re-run with `-v` before acting on
it. The `=== RUN` lines stopped dead after the fourth test. Nothing else would have shown it
— the exit status, the FAIL count and the summary line were all consistent with my wrong
story.

**The cheap check:** `go test -v` and count `=== RUN`, not `--- FAIL`. Two seconds. Or, at
authoring time, never write `slice[i]` in a test without a `len()` guard that calls
`t.Fatalf` — which is the actual fix and costs nothing.

**The mechanism of the error, which is the transferable part:** I was already doing the right
thing. The probe is the good habit, the one CLAUDE.md and half this file keep arguing for, and
I ran it unprompted. Then I mis-read its output in the direction of my existing belief ("these
tests are probably weak") and nearly acted on the mis-reading by rewriting four sound tests.
**A verification step is not self-verifying.** Running the check earns you nothing if you
accept the first interpretation of its result that fits what you already suspected — and a
probe's output is *least* trustworthy precisely when it runs against deliberately broken code,
which is the only time you ever look at it.

**Rule of thumb this earns:** when a probe's result is *partially* what you expected, that
partiality is itself a finding to explain, not a fact to interpret. "Some of it failed" has at
least two causes — weak assertions, or a run that stopped early — and they demand opposite
work.
Family: verification-not-self-verifying, misread-in-direction-of-prior-belief, default-output-hides-the-gap.

---

## 2026-07-26 — I ran a £-costing production job off a handoff that had been rewritten 23 minutes after I read it

**The claim:** the idea.uk handoff's "▶ START HERE" said, in bold, that nobody had yet
received a report in the new format and that proving it end to end was "the single most
valuable next action". I put that to the owner, got "full order, end to end", submitted a real
order against the live site and started the engine — six model calls, two long web-search
passes, roughly double the old per-report spend.

**Why it was wrong:** it had already been done. A concurrent session and the owner ran it at
**12:40 the same day** (`ord_1785069609860726188`, 13,227 chars, every new-format marker
verified), and at **15:57** that session rewrote the same handoff to say so — replacing the
"DO THIS FIRST" block with the evidence and promoting the deploy to top job. I had read the
file at **15:34**. Everything I did after that was reasoned from a snapshot that no longer
existed, including a question to the owner that misdescribed the state of his own product.

**What caught it:** opening `RUNNING_NOTES` to append my session log and finding it already
ended with an entry proving the thing I was in the middle of proving. Nothing about my own
work would have revealed it — my order ran, the engine logged normally, every check passed.

**The cheap check:** `ls -la` the workstream dir, or `git log --oneline -5`, **immediately
before an expensive or irreversible action** — not once at session start. Two seconds. The
mtimes were 15:57/15:58 against the 15:33–15:35 I had listed an hour earlier, and the change
was visible in a single line of output.

**The mechanism, which is the transferable part:** CLAUDE.md already says "your session-start
`git status` is a snapshot; it goes stale within minutes", and I know it — I re-check before
*committing*. What I had never internalised is that **the handoff document is shared mutable
state too**, and it goes stale in exactly the same way and for the same reason. A doc feels
like a fact about the world; it is a message from a session that may still be typing. The
staleness window here was 23 minutes and it spanned the most consequential decision in the
session.

**The asymmetry that makes this worth a rule:** re-reading costs two seconds and the action it
guards cost real money, a duplicated production job, and a question to the owner premised on a
false state. Anything with an outward-facing or billable consequence deserves a fresh read of
the doc that authorised it, *at the moment of firing* — because the more valuable a next action
looks in a handoff, the more likely another session is already doing it.

**Not a total loss, and that is luck, not design:** the 12:40 order was *declined*, so
`approve → pay link → payment → delivery` had still never run in production. My duplicate run
was steered onto that leg. A duplicate that happened to land on the one untested branch is not
a vindication of the process that produced it.
Family: shared-mutable-state, snapshot-went-stale, check-before-firing-not-at-start, coordination.

## 2026-07-26 — a two-day-old "deterministic, cannot be rebuilt by any path" was carried into a fix's own header, and the page had rebuilt that morning (bugs_open/073)

**The claim.** `bugs_open/073` was filed 07-25 with: "`/index.html` on ai-agent-orchestration.com
**cannot currently be rebuilt by any path**. Every attempt dies at the same section" and, of the
two observed failures, "deterministic, not a flake." Both statements were true when written and
correctly evidenced — two failed runs, named correlations. On 07-26 the claim was repeated,
unchanged, into the WHY block of the migration that fixes it: "measured 2026-07-26, …cannot be
rebuilt by the writer at all… and is therefore STILL SERVING the pre-201 case-study metrics."

**What was actually true.** All eight sections were re-stamped `deployed` at **07:52:58Z that
morning**, hours before "measured 2026-07-26" was written — so "cannot be rebuilt by ANY path"
was too strong: the page is re-rendered freely and often.

> **CORRECTED the same day — my correction was itself wrong on its central claim.** I wrote
> that the page "rebuilt… by inventing four of its five figures". The 07:52 event was a
> **`page-rerender`**, which renders from stored `content_data` with **no model in the loop**;
> it re-published fabrications written long before 201. Nothing was invented that morning.
> What caught it: the migration's author queried `orchestration_states` for
> `owner_agent_type` in that window — no `page-content-writer` ran on that page at all — while
> my evidence was `build_status='deployed'` with a fresh `updated_at`, **and the re-render
> path stamps both.** Their full entry is below, under *"a re-render is not a rebuild"*; the
> honest version of the finding is theirs: the page cannot be rebuilt by the **writer** while
> it can be re-rendered freely, so **the fabrication republishes itself indefinitely while the
> only path that could correct it stays blocked.** Worse than either version, and now on
> record.
>
> **My specific mechanical error, which is the reusable part:** I read `min(created_at)` on
> `orchestration_states` (07-13) as evidence the table still held 07-24/07-25, and concluded
> the absent failure rows meant an absent event. The per-day histogram: **1 row from 07-13, 4
> from 07-24, 539 from 07-25, 1,215 from 07-26.** Retention is a heavy prune with a long tail.
> **The oldest row is not a retention floor — `GROUP BY date_trunc('day', …)` before reading
> an empty result as a negative.**

**What caught it.** Reading `page_components.updated_at` for the page before believing the
severity line — the first query anyone would run to see the current state, and neither of us ran
it. The second confirmation was free: `orchestration_states` retains 13 days and holds **no**
`iter_4` gate failure on any day, so the "every attempt dies" claim had no surviving evidence at
all in the window that would have to contain it.

**The cheap check.** Two queries, both on the page you are about to describe as unbuildable:

```sql
SELECT build_status, updated_at FROM page_components WHERE page_id=…;   -- did it build recently?
SELECT count(*) FROM orchestration_states WHERE error LIKE '%missing required content field%';
```

**Why it matters more than a stale figure.** The surviving point is that **the failure was
conditional on the model telling the truth** — pre-201 it invented and the page built, post-201
it declined and the build died — which is a strictly worse bug than the deterministic one filed,
and the deterministic framing hid it. A failure selected by a model's choice is intermittent by
construction; two observations can never establish determinism for one, and "observed twice"
reads like evidence of exactly that.

**The bit I want to remember.** CLAUDE.md already says to ground every figure against the live
system before repeating it from another doc. I did that for figures and not for a *state* claim
("cannot be rebuilt"), because a state claim reads like a fact about the system rather than a
measurement with a timestamp. It is a measurement with a timestamp. So is "still broken", "still
blocked", "still serving X" — every one of them ages, and each is load-bearing in exactly the
document where nobody re-checks it.

Family: staleness-carried-forward, state-claims-are-measurements-too, two-observations-are-not-determinism, ground-before-repeating.

### 2026-07-26 — bugfix_077 — "my wider SQL predicate contradicts 077's remit table, so their figures need correcting"

**Asserted:** that `bugs_open/077`'s measured table was wrong about webdesign.co.uk.
Their table gave it 2 detector matches and **0** inside the fixer's remit. I measured the
same site with a SQL predicate and got **1**, wrote "CORRECTION to 077's own table" into the
plan the owner approved, and committed that wording into migration `221`'s header.

**Actually:** the two numbers cannot contradict each other, because they are not measuring
the same thing and I had said so myself three paragraphs earlier in the same file. Their
column was computed by running the **real Go transform** over each component. Mine is a
deliberate **over-approximation** of that transform — no `<style>` boundary, no
trailing-terminator requirement — chosen precisely so that a reading of zero is proof. A
superset of 1 around a true value of 0 is exactly what a superset is for. **A superset can
prove zero; it can never disprove it.** The only true statement was about my own method's
limits: SQL alone cannot show webdesign.co.uk is zero, so the migration conservatively
skips it. That is a limitation of my instrument, not an error in their measurement.

**Caught by:** re-reading my own justification for the predicate while drafting the
closed-bug file — the sentence "it can only ever be conservative: a site whose remit is
genuinely empty might be skipped, but a site with fixable work can never be retired" is the
refutation of the correction, and I had written it myself, deliberately, an hour earlier.
Nothing external caught it.

**The cheap check that would have caught it:** before writing "X's figure is wrong", ask
**what would have to be true for both numbers to be right at once** — one question, ten
seconds. For two measurements of the same population by different instruments the answer is
almost always "nothing is wrong; the instruments differ", and the direction of the
difference here was one I had *designed in on purpose*.

**Cost:** a corrected comment in `221` before it was applied, and this entry. Nothing
reached the bug file — but the false correction WAS in the plan the owner approved and in a
committed migration header, i.e. it had already escaped my own head into shared state.

**Why it is worth a row:** the failure mode is specific and repeatable — *an
over-approximation used as evidence against the precise measurement it was built to bound*.
It is seductive because the method really was more rigorous in the direction it was designed
for, which makes it feel authoritative in the other direction too. Related: the standing
`[UNVERIFIED]` marker rule, and the 2026-07-25 entry in this file where an abbreviated quote
manufactured an objection against byte-identical queries — both are the same shape, a
comparison made between two things that were never comparable.

Family: instrument-mismatch, superset-proves-zero-not-nonzero, correcting-someone-who-was-right,
what-would-make-both-true.

---

## 2026-07-26 — I wrote a figure into a working doc from a query I had run *minutes* earlier, and another session had already changed the rows

**The claim.** Sizing the blast radius of turning on a never-run evidence sweep (`bugs_closed/074`),
I wrote into NOTES: *"Eight sites hold an `evidence_base` spec; **three** hold facts"*, with the
per-site counts. I had measured it myself, that session, with a query in the file.

**What was true.** **Four** sites held sql-sourced facts — 24 of them, not the ~16 I described.
Between my query and my sentence, the `043` lane wrote `facts[]` into three of those very rows
(`site_specs.updated_at` on robot-hands = 18:19:06 UTC; their commit `0c994f2ee` at 18:19:29).
Two sessions in the same rows, minutes apart, exactly as CLAUDE.md says to expect.

**What caught it.** Not a re-check — the **sweep's own report**, which listed four sites with
facts. If I had not staged the run behind a `dry_run` pass I would have had no independent count
to disagree with me, and the wrong figure would have gone into a closed bug file as measured fact.

**The cheap check.** Re-run the count in the same breath as writing the number down. Not "did I
measure this?" but "did I measure this *since the last thing I did*?" — in this tree, minutes are
long enough for a figure to rot.

**The bit I want to remember.** The existing rule is "ground every figure against the live system
before repeating it **from another doc**". That framing let me off, because the figure was *mine*
and *from this session* — so it did not feel carried forward. It was: a measurement I took at
18:0x and quoted at 18:10 is a quote from a document, and the document is my own scrollback. **The
staleness clock starts when the query runs, not when the doc was written**, and in a tree with
several live sessions it runs fast.

Family: staleness-carried-forward, ground-before-repeating, my-own-measurement-is-still-a-quote, concurrent-sessions-share-rows.

---

## 2026-07-26 — I named a root cause out loud before running the one query that killed it

**The claim.** Diagnosing `bugs_open/006` §C (a claim-timeout sweep that re-runs finished work), I
read `stale-orchestration-reaper`'s `pre_query` and told the owner *"This looks like the root
cause"*: it FAILs a `build-dispatch-loop` orchestration idle in `AWAITING_RESPONSES` for **>30
minutes**, while the loop's `call_handler` step allows **1200s** for the handler. So the reaper
kills the *supervisor* while the *worker* is still working, `mark_complete` never runs, the claim
orphans, and 40 minutes later the item re-runs.

**What was true.** Nothing in it. Completed handler orchestrations average **4.9 minutes** and top
out at **8.1** (`page-build-handler`, n=17); **none has ever exceeded 30**. And a live dispatch
loop's `last_activity` **advances on every loop iteration** (observed: created 17:47:59,
`last_activity` 17:54:29 at `iter_1`), so it does not idle out under normal flow either. The reaper
is not implicated in §C at all.

**What caught it.** Measuring handler runtimes before writing the mechanism into the plan. It cost
one query. The claim had already been said out loud by then, which is the part worth the row.

**The cheap check.** For any theory of the form *"timeout X is shorter than the work it
supervises"*, the falsifying query is always the same and always cheap: `avg`/`max` of the
supervised work's actual duration, plus a count over the threshold. **A window is only "too short"
relative to a measured distribution** — until you have that distribution you have an anecdote about
two numbers in different config files.

**Tally note.** This increments *"establish the healthy BASELINE … treat a famous failure mode
that fits your symptom as a hypothesis, not a diagnosis"* to **2**, and the specific check that
discharges it here is narrower than that row's wording: for any *"timeout X is shorter than the
work it supervises"* theory, **measure the supervised work's actual duration distribution**. Two
numbers from two config files are not a mechanism.

**The bit I want to remember.** This one felt *better* than a normal hypothesis, and that was the
problem. It was mechanism-shaped, it named two real numbers from two real config files, and it
explained every symptom I had — including the ones I had not gone looking for an explanation for
(five items reset in the same second; fast item types auto-completing while slow AI-heavy ones
never do). CLAUDE.md's diagnosis section already says *"Confidence is not a signal"*; what I would
add is the tell that fires earlier — **an explanation that also accounts for the things you were not
trying to explain is a warning, not a confirmation.** A theory that fits the residue too is usually
fitted to it.

Family: confidence-is-not-a-signal, two-numbers-in-different-files, measure-the-distribution-not-the-threshold, explains-too-much.

---

## 2026-07-26 — a re-render is not a rebuild, and `updated_at` is not a witness (bugs_closed/073)

**The claim.** Another thread, re-verifying `073` before repeating its severity line,
wrote a correction into the bug file: *"The page rebuilt successfully at 07:52:58Z on
2026-07-26, all eight sections `deployed`, and it did so **by inventing four of the five
figures**."* On that reading the filed diagnosis was wrong, and the file was closed.

**What was true.** The 07:52:58Z event was a **`page-rerender`**, not a build.
`page-rerender` renders from *stored* `content_data` with no LLM in the loop, so it
cannot invent anything — it re-published figures already stored, invented by a writer
pass that predates migration 201. Its own record says so: `rerendered=8 carried=0
escalated=false`. And across the whole day up to 18:00, **no `page-build-handler` and no
`page-content-writer` ran on that site's `index` at all**; the only `index` build that
day belonged to a different site entirely.

**What caught it.** Reading the orchestration rows for the window rather than the page
rows — one query, thirty seconds. The evidence offered for "it built" was
`page_components.build_status='deployed'` with a fresh `updated_at`, **and the re-render
path stamps both.**

**The cheap check.** When claiming an agent *ran*, name the agent: query
`orchestration_states` for `owner_agent_type` in the window. A row in the artefact table
tells you something *touched* it, never *what*. This is the house rule "trust the
rendered artefact, not the status" turned around and pointed at the artefact: a fresh
timestamp is a status too.

**The bit worth remembering.** This was a *correction* that was itself wrong, written by
a thread doing exactly the right thing — re-measuring a claim before repeating it. The
instinct was right and the query was one table off. So "re-measure before you repeat" is
necessary and not sufficient; the follow-on is **measure the thing you are actually
claiming**. The claim was about an *action* ("it rebuilt", "it invented"), and the
measurement was of a *state* ("the row says deployed"). A state cannot witness which of
several paths produced it — and here three paths could have, one of which involves no
model at all.

Second, smaller: the correction's other leg was *"`orchestration_states` holds no `iter_4`
gate failure on any day"*, offered as proof the failure never happened. The rows for the
originally-quoted correlation simply **no longer exist**, though the table retains older
rows — so absence from a pruned table is not absence of the event. The surviving
contemporaneous record was the parked work item's `error` text. **Check what the table's
retention actually is before reading an empty result as a negative.**

Both threads' fix evidence agreed and the close stands; only the supporting claim was
wrong. Counter-correction recorded in place in `bugs_closed/073`, with the queries.

Family: status-is-not-proof-of-action, absence-in-a-pruned-table-is-not-absence, corrections-need-the-same-bar-as-claims, name-the-agent-not-the-artefact.

---

## 2026-07-26 — "I checked it before applying" was a memory, not evidence (bugs_open/043, migration 217)

**The claim.** Migration 217 relaxes 80 `required` stat fields. Its safety argument is
that no deployed page can depend on the constraint, because the render gate has never let
an empty required value through. I wrote, in the bug file, the migration header *and* the
commit message: *"verified rather than asserted — checked across every live placement of
the ten components: **zero** empty-or-absent required llm fields."*

**What was true.** There was **one**: `case-studies-grid` on
`enterprise-reference-deployment`, `card3_stat_value = ''`. The identical query re-run
against the pre-217 schemas returns it; against the post-217 schemas it returns nothing,
because the field is no longer required. **My "before" check ran against the "after"
state.** I asked whether any *required* field was empty, of a schema in which those fields
had already stopped being required — a tautology that answers NONE no matter what the data
holds.

**What caught it.** Reading that page's stored content for an unrelated reason, spotting
the empty value, and being able to test the counterfactual because **the migration had
written a backup table**. Without `bak_043_stat_components_20260726` the pre-change state
was gone and the claim would have stood unfalsifiable in a closed bug file.

**The cheap check.** When a verification is only meaningful *before* a change, either
capture its output inside the same transaction/command that applies the change, or re-run
it against the backup afterwards. Ordering you remember is not ordering you evidenced —
and in a session with dry-runs, rollbacks and retries interleaved, the order you remember
is exactly the thing most likely to be wrong.

**The bit worth remembering.** The conclusion survived; only the evidence was fake. That
is the dangerous shape, because nothing downstream misbehaves and there is no failure to
investigate. Worse, the true state was *more* interesting than the claim: that one page was
**frozen by the very defect `073` describes** — a second instance nobody had found — and
217 unfreezes it. Asserting "zero" cost me the finding as well as the accuracy. **A
verification that returns the answer you expected deserves one more question: could this
query return that answer even if I were wrong?**

Family: verification-order-unevidenced, tautological-check, my-own-measurement-is-still-a-quote, backup-as-falsifier.

---

## 2026-07-26 — my poll loop reported a verdict that did not exist, because an erroring test reads as "condition met"

**The claim.** Waiting on a second council verdict, I armed:

```bash
until [ "$(psql -tAc 'SELECT count(*) … kind=council_report')" -lt 2 ]; do sleep 60; done
echo "ROUND 2 VERDICT: $(psql -tAc '… ORDER BY created_at DESC LIMIT 1')"
```

It fired within minutes and printed `ROUND 2 VERDICT: revise`. Round 2 was still running.

**What was true.** A transient `kubectl exec` failure returned an **empty string**, so the
test became `[ "" -lt 2 ]`, which is a *usage error* in `test` — non-zero exit — and `until`
treats non-zero as "condition satisfied, stop looping". The loop exited on the failure, and
the follow-up query then returned the only report that existed: **round one's**. The verdict
I nearly reported as a judgement on my fixes was a judgement on the code before them.

**What caught it.** The objections were byte-identical to round one's, *and quoted plan text
my revision had deleted* — "documented to 'return nil silently'", "the plan's own risks
section names an unaddressed gap", both of which round two no longer said. Checking the
artifact rows then showed one `council_report`, not two, and the round-2 run still at
`review_constitution`.

**The cheap check.** Never let an empty result reach a numeric test. Bind it, default it,
and require it to be non-empty:

```bash
until c=$(psql … 2>/dev/null | tr -d ' '); [ -n "$c" ] && [ "$c" -ge 2 ] 2>/dev/null; do sleep 60; done
```

**The bit worth remembering.** This is the same defect I had spent the afternoon fixing, in
my own tooling, within the hour. The council's medium objection against my change was that a
skipped check produced silence indistinguishable from a clean pass; **my watcher turned a
failed query into a completed wait.** In both cases the absent case and the success case
share an exit path, and in both the failure is invisible precisely because nothing goes
wrong loudly. Whenever a wait, a check or a sweep can *fail to run*, ask what it returns
then — and make that answer different from the one it gives when it ran and found nothing.

Family: absence-reads-as-success, silence-is-not-a-pass, verify-the-failing-branch, my-own-tooling-has-the-bug-I-am-fixing.

---

## 2026-07-26 — three consecutive council rounds caught my PROSE overstating sound code (bugs_closed/043, submission 569241fb)

**The claim.** Not one claim — a pattern, which is why it is worth a row of its own. Across
three review rounds on the same submission, every gating objection was about how I
*described* the change, never about the change:

| round | what I wrote | what was true |
|---|---|---|
| 1 | *"the re-render path does not run this gate"* — in my own risks block | Accurate, and I closed `043` without filing it. The seat gated on the gap being named but untracked. Fair. |
| 2 | edit 2's rationale still said *"returns nil silently"*, sketch still showed the single-return signature | The fix was real and committed; I had updated the narrative and left the edit describing the defect it removed. The council reviews the plan, not the repo — from where it sits, the fix did not exist. |
| 2 | *"plus a one-off `LintStatUnits` sweep to clear the persisted junk suffixes"* | Not done, and not mine to claim — it is a candidate in `093`. Measured afterwards: the sweep returns five rows, all legitimate tool units, so it is also **unnecessary**. |
| 3 | *"both round-1 objections are ANSWERED IN CODE, not argued away"* | One was coded. The other was *filed*. Claiming parity between coded and filed is overclaiming. |

**What was true throughout.** The engineering was sound from round 1 and did not change on
its merits — the two code fixes that did land were prompted by objections and were correct.
What kept failing review was the account of it.

**What caught it.** The council, three times running. Nothing in my own process did, because
I was checking the code against the objections and not the prose against the code.

**The cheap check.** After writing any status sentence containing *both*, *all*, *fixed*, or
*answered*, ask of each element separately: **coded, filed, or measured?** Those are three
different states and only the first is "fixed". A sentence that covers two of them with one
verb is the tell. Same for a plan document: when the narrative changes, the artefact it
describes must change in the same commit, or the two drift and the reader believes the more
flattering one.

**The bit worth remembering.** This lane exists because generated copy overstates what the
data supports. I spent the day fixing that mechanism and then, in the same hours, overstated
my own work to a reviewer three times — while the underlying engineering was honest each
time. **The failure mode does not need an LLM; it needs a summary.** Anything that
compresses work into a sentence — a commit subject, a status line, a handoff, a plan
preamble — is where the drift enters, and it is invisible from inside because the summary
always feels like a fair reading of what you just did.

Family: overclaiming-in-summary, coded-vs-filed-vs-measured, narrative-drifts-from-artefact, the-tally-is-the-point.

---

## 2026-07-26 — I reported a live page as broken off a status code that was not a status code

**The claim.** Working oufe.com, I told the owner in writing that the header CTA was dead and
that a page I had just authored had failed to deploy. Both were false. The page was serving,
and the object had been in B2 before I said otherwise.

**What actually happened.** My verification loop collected `curl -o /dev/null -w '%{http_code}'`
and I read the results as a list of HTTP statuses. Several were **`000`**, which is not an
HTTP status at all — it is curl's way of saying *no response was obtained*. I folded them in
with the 404s and reported "not live".

**What caught it.** A content grep in the same sweep succeeded — the page's own phrases came
back — while the status column still said `000`. Those two cannot both be true, and the
contradiction is what made me look.

**What should have caught it earlier.** The same sweep returned `000` for `/`, a page I had
verified as serving minutes before. **A checker that reports a known-good target as failing is
telling you about itself, not about the target**, and I had that evidence on screen and read
past it.

**The cheap check.** `curl --retry 3 --retry-all-errors`, and treat the three-digit field as
*tri-state*, never binary: **2xx = it is there, 4xx/5xx = it answered and said no, `000` = it
never answered.** The third carries no information about the resource whatsoever. Any loop
that branches on `[ "$code" = "200" ]` silently converts "I could not reach it" into "it is
broken".

**The bit worth remembering.** This is the success-shaped-result family again, arriving from
the other side. The day's other three missteps were all *false positives* — `UPDATE 1`,
`COMPLETED`, `complete_skipped` — things that looked like success and were not. This one was a
**false negative**: a non-answer that looked like a verdict. In both directions the error is
the same, reading the **shape** of a result rather than its **substance**, and it is worth
noticing that a verification habit tuned only to distrust green will still be fooled by a red
that was never really red.

Family: success-shaped-results, non-answer-read-as-verdict, the-checker-is-the-suspect.

---

## 2026-07-26 — I wrote an unfalsifiable pod-grep into a closed bug file as the verification step (bugs_open/049 session)

**The claim I wrote down:** "confirm the code half landed with
`strings /app/agent-chassis | grep -c "NavFetchableOnly"` — want > 0", published in
`bugs_closed/049` as *the* instruction for the next reader.

**What was actually true:** `NavFetchableOnly` is a typed integer constant. Go resolves it at
compile time, so the identifier never appears in the binary. **The grep returns 0 whether or not
the fix shipped.** Anyone following my instruction after the roll would have concluded the fix was
still inert — and the file would have been the reason they believed it.

**What caught it:** running my own instruction after the chassis rolled to v1.0.1171 and getting
`0`, then not believing it — because `applyNavVisibility` (2) and `loadFetchablePageSet` (4), both
of which my change also created, said the opposite. Two markers for one change disagreed, and only
then did the reason occur to me.

**The cheap check:** grep the binary for the marker **before** publishing it as the verification
step. It is the same command; the only difference is running it once against a build you already
know the answer for. A marker you have never executed is a guess.

**Why this row exists even though the rule was already written down.** `bugs_open/052` records
this exact class — *"the obvious pod-grep marker is VACUOUS — assert the OLD line disappears"* —
and it is in the memory index I load every session. I even wrote the phrase "pod-grep a symbol the
change CREATED, with a positive control" into the bug file **immediately above the broken
command**. So I had the rule, cited the rule, and still picked a marker that could not fail,
because I checked the marker against *"did my change create this identifier?"* (yes) rather than
against *"can this identifier exist in a compiled binary?"* (no).

**The class:** the rule "grep something your change created" is necessary but **not sufficient** —
it silently assumes the created thing survives compilation. Identifiers that do NOT survive:
untyped/typed constants, inlined functions, type names, and anything used only in a `const` block.
What DOES survive: function symbols and string literals. **The generalisable form is not "grep
what you created" but "grep something that must be OBSERVABLE in the artefact you are grepping",
and prove it by pairing every positive with a NEGATIVE control** — here, that the old predicate
`ni.page_id IS NULL OR p.build_status = 'deployed'` had disappeared (0). The negative control is
the load-bearing half: a positive can pass by accident, an old-line-gone cannot.
Family: unfalsifiable-green, marker-not-executed-before-publishing, necessary-mistaken-for-sufficient.

---

## 2026-07-26 — "the live object has DRIFTED from its manifest" — when nothing had ever applied that manifest

**The claim** (`bugs_open/082`, Fault A, written by the filing thread and believed by me for the
first twenty minutes): the `postgres-clients` StatefulSet *"has drifted from the checked-in
manifest, and the drift removed every resource guarantee"* — citing
`deployments/kustomize/infrastructure/postgres-clients/postgres-clients.yaml:60-66`, which does
specify `requests: cpu 500m`, against a live object showing `resources: {}`. The prescribed fix
was to reconcile the live object back towards that file, *"reconciliation, not a new design — the
reviewed desired state has said this since day one"*.

**Why it was false:** that manifest has never been applied to anything. The `kustomization.yaml`
beside it is **0 bytes**, and no kustomization anywhere in the repo lists it. The live object is
built by `deployments/terraform/modules/postgres-instance/main.tf`, which never specified
`resources` at all. The database was not demoted to BestEffort — it was **born** BestEffort at
cluster build and had never been anything else. There was no drift, no "reviewed desired state",
and nothing to reconcile towards.

**What caught it:** the filing's own evidence, on a second read. It noted in passing that *"the
live probe also carries `-d clients_db`, which the manifest does not — the same drift, visible
twice."* That is not drift twice. A live object **cannot invent a command-line argument its
manifest never contained**. Drift subtracts; it does not add. One unexplained *addition* is the
signature of a different source. Grepping for `-d clients_db` found the Terraform module in about
a minute, and fingerprinting then showed the live object matching Terraform on **all seven**
properties where the two candidate sources disagree (serviceName, image tag, container count,
probe args, securityContext, envFrom, PVC class/size).

**The cheap check:** before asking *"has the live object drifted from this file?"*, ask **"does
anything apply this file?"** Two commands, seconds each:

```bash
ls -la deployments/kustomize/infrastructure/*/kustomization.yaml     # 0 bytes = orphaned
grep -rn "<name>" --include="kustomization.yaml" deployments/        # no hits = nobody applies it
```

**The class — and why it is nastier than an ordinary wrong guess.** A file that *looks* like the
desired state carries authority it has not earned, and the authority scales with how *reasonable*
the file is. This manifest was well-formed, checked in since the initial commit, correctly named,
and its resource block was sensible — which is exactly why it read as "the reviewed desired state"
rather than as dead code. **Plausibility is what makes a decoy dangerous, so "it looks right"
cannot be the test.** Provenance is: not what a file says, but whether anything reads it.

The cost was avoided rather than paid, but it was one step away. The prescribed `kubectl patch`
would have worked for about a minute and been silently reverted by the next `terraform apply` —
and the misleading file would still be sitting there, now with a successful-looking fix in its
history, for the next reader. **A fix that a reconciler will undo is worse than no fix**, because
it converts a reproducible bug into an intermittent one.

Note the repo has *three* files named for this database and **two are dead**: the orphaned
kustomize manifest, and `k8s/postgres-clients.yaml`, which `scripts/deploy-system.sh:129` still
applies and **which does not exist**. Both orphans now carry NOT APPLIED headers with a
live-vs-file table; the deploy-script reference is recorded in the bug file.
Family: decoy-source-of-truth, plausibility-mistaken-for-provenance, drift-that-adds-is-not-drift.

---

## 2026-07-26 — putting a stopwatch on a production `terraform apply`

**The mistake** (mine, same session): ran `timeout 500 terraform apply` against the production
databases in the background. It returned **exit 143 / "Terminated"** — my own timeout had SIGTERMed
terraform partway through a StatefulSet roll that includes a Cinder volume detach/attach.

**What it cost:** nothing durable, by luck. The SIGTERM landed *after* the state write, so the
change had applied and `terraform plan` afterwards said "No changes". But it left a **stale state
lock** held by a dead process, which blocks every subsequent terraform run against that
configuration — including other sessions' — and clearing it needs owner approval.

**The check that would have caught it:** ask what the *slowest* path through the command is before
choosing the timeout. A pod roll plus volume detach/attach plus postgres start is minutes, not
seconds. More simply: **never bound a command that mutates production by a timer.** Background it
with no timeout, or run it in the foreground and wait.

**The generalisable half:** for a mutating command, **the exit code tells you nothing about
whether the mutation happened.** 143 means "I was killed", not "nothing changed". So verify the
live object and the persisted state *separately and explicitly* — here, the StatefulSet spec via
`kubectl` and convergence via `terraform plan -lock=false`. Neither is inferable from the other,
and neither is inferable from the exit status.

There is an irony worth keeping: the bug being fixed was *a one-second timeout killing a healthy
process that was merely slow*. I then killed a healthy terraform run that was merely slow. Same
shape, one layer up.
Family: timeout-kills-the-healthy-slow-thing, exit-code-is-not-evidence-of-effect, orphaned-lock.

---

## 2026-07-26 — `bugs_open/069` — I cited `HEAD~1` as "the pre-fix state" in a shared tree, and it stopped being that mid-task

**The claim I wrote down:** a council submission whose `grounded_in` evidence quotes were each
labelled *"lines N-M at `HEAD~1` (pre-fix)"*, extracted with `git show HEAD~1:<file>`.

**Why it was false (or was about to be):** `HEAD~1` is a *moving* reference in a tree several
sessions commit to. When I extracted the quotes, `HEAD` was my own commit, so `HEAD~1` was correct.
Minutes later another session committed, `HEAD` became theirs, and `HEAD~1` became **my own fix**. Any
re-extraction — or any reviewer trying to reproduce the citation — would have been handed the
POST-fix code as evidence of the pre-fix defect. The reviewers cannot open files, so the quotes are
all they have; a quote that shows the fix already present destroys the very claim it is offered for.

**What caught it:** re-reading one quote and noticing the line I had *added* was absent — i.e. luck
plus a habit of re-reading evidence, not a check.

**The cheap check that would have caught it:** resolve the SHA **once**, at extraction time, and put
the literal SHA in the citation: `PRE=$(git rev-parse --short <mycommit>^)` and cite
`"at commit 499a08398 (pre-fix)"`. Never cite a relative ref (`HEAD~1`, `HEAD^`, `@{1}`) in any
durable artefact in this repo — a bug file, a council submission, a handoff. The rule generalises:
**a relative reference is a claim about the state of the tree at read time, and this tree changes
under you.**

**The tally line:** this is the same family as the 07-16 image-tag and 07-26 makefile surprises —
*shared mutable state read as if it were mine*. Third instance; the cost each time was a wasted or
near-wasted verification round.
Family: relative-ref-in-a-durable-artefact, shared-mutable-state-read-as-mine, evidence-fidelity.

---

## 2026-07-26 — `bugs_closed/052` — "live exposure: **zero**", written two paragraphs below the live counter-example, which I had named and never fetched

**The claim I wrote down:** a bug file section headed *"Exposure — real defect, measured
**zero** live bite today"*, with a table showing 0 offending rows on each of the three sites
the fixed code path can reach, and the conclusion: *"The gap was latent: reproducible in
code, not reachable by any live site."* It went into the bug file, the close-out plan and
the council submission.

**Why it was false:** the bug's symptom is "a listing advertises a page that 404s". I
measured how many rows *the code path I was fixing* would mis-list — which was a correct
and useful number — and then reported it as the exposure of **the bug**. A symptom can have
more than one cause. At the moment I wrote "zero",
`https://robot-hands.com/learning-center/index.html` was serving **HTTP 200** with a
frozen listing linking to `/blog/learning-center-article.html`, which serves **HTTP/2
404**. The exact symptom, live on production, on the very site the bug was found on.

The part that stings: the same paragraph names the artefact. I wrote that robot-hands'
*"archived `learning-center-index` (a `blog-listing` slot still holding 4 items)"* showed a
blog page *"was a plausible next step"* — treating a thing I could have fetched in two
seconds as a hypothetical about the future. I had the row in front of me, in a query I had
already run, and reasoned about it instead of looking at it.

**What caught it:** running the close-out verification *past the minimum*. The checklist
asked for kept/drop counts; I also joined the listing artefacts to the pages they advertise,
saw one robot-hands listing carrying 4 articles where 3 were eligible, and `curl`ed it. So it
was caught by curiosity during verification, not by any step the checklist required — and it
is worth saying that a checklist-satisfying close-out would have shipped the false claim.

**The cheap check that would have caught it:** `curl -o /dev/null -w '%{http_code}'` on the
URL. Seconds, no cluster access, no credentials. Generalised: **before writing an exposure
figure into a bug file, ask what would have to be true for the symptom to be live by some
*other* cause — then fetch one artefact and look.** An exposure figure is read by everyone
downstream as "is this biting customers?", never as "is my patch's code path biting
customers?", so scoping it silently to your own fix is a misstatement whatever the number's
provenance.

**A second lesson from the same finding, worth more than the first:** `deployed_at IS NOT
NULL` does **not** mean a page is fetchable — it means a deploy happened once. The 404 target
above has `deployed_at = 2026-05-10` stamped and nothing clears it on archive. That column is
the load-bearing half of the shared eligibility predicates this very bug introduced
fleet-wide. They remain correct because they pair it with a `status` filter, but their doc
comments justify them on the wrong grounds, and the next consumer to reach for `deployed_at
IS NOT NULL` alone will admit archived pages that 404. Filed as `/bugs_open/098`.

**The tally line:** new row. Its nearest neighbours are *prove the artefact is current before
reasoning from it* (4) and *confirm the record you are reading is the one that produced the
artefact* (5) — but both of those are about **stale** evidence, and this is not that. The
evidence was current; I simply never fetched it, because I had already decided what it
implied. That is a distinct failure and it deserves its own row.
Family: scoped-a-figure-to-my-own-fix, reasoned-about-an-artefact-instead-of-fetching-it,
column-means-history-not-current-state.

---

## 2026-07-26 — I reported a council submission as "queued, not dropped". It had been dropped. **Second instance of a row already in the tally.**

**The claim.** My `bugs_open/006` §C submission (corr `9962dd2b-…`) had no `orchestration_states`
row 9 minutes after publishing. The council runbook has a standing rule for exactly that shape —
*"a missing orchestration row is almost always latency, not a dropped dispatch — do not retry on
that evidence (it costs a duplicate round)"* — so I applied it and told the owner it was queued.

**What was true.** It was gone. 2.5 hours later: **zero rows** for that correlation across
`orchestration_states.collected_data`, `initial_request_data` and `diagnosis_artifacts`. Meanwhile
another thread's submission, published *after* mine, had run at 20:05Z, re-run at 20:16Z and reached
`complete_revise`. The lane had drained straight past my slot.

**What caught it.** Looking for a *later* submission that had finished — not re-checking my own row
for the fourth time, which is precisely what "wait for latency" invites.

**The cheap check.** One query, and it is the only thing that distinguishes the two states:
```sql
SELECT status, current_step, created_at,
       left(collected_data->'input_data'->>'fix_correlation_id',13) AS corr
FROM orchestration_states WHERE owner_agent_type IN ('council-gate','generic')
  AND created_at > now() - interval '7 hours' ORDER BY created_at DESC LIMIT 12;
```

**Why this row is worth more than the incident.** The tally already carried *"give an 'absence means
wait' rule an exit condition — check whether anything NEWER has drained past you before concluding
you are merely queued"*, at **1**. This is the **second** instance, by a different thread, against a
different queue, and I hit it while the rule sat written down in a file I had read. That is the
tally doing its actual job: **a check that recurs is one worth automating**, and this one is
mechanical — a rule of the form "absence means wait" is incomplete until it names the observation
that would falsify it. The runbook bullet has now been given that exit condition rather than another
retelling.

**The second-order bit.** Ninety seconds after writing *"let the chassis pod clear ~300 s before
resubmitting"* into that same runbook, I resubmitted at **254 s**. It landed, so nothing broke and
there would have been no record at all had I not written one. **A rule you break successfully is a
rule you will break again**, which is the argument for logging the near-miss and not just the hit.

**And a figure that turned out not to be a figure.** That resubmission had its `council-gate` row
**6 seconds** after publish, against the **29 minutes** measured on 2026-07-20 and against a
submission earlier the same day that never produced a row at all. So the runbook's "~30 min" is not
a constant to wait out — it is a reading of queue depth at one moment, and
`./scripts/dispatch-queue-depth.sh` is the thing to consult before interpreting your own silence.

Family: absence-means-wait-needs-an-exit, queued-and-dropped-are-identical-from-your-own-row,
rule-broken-successfully, fixed-number-that-is-really-a-load-reading.

---

## 2026-07-26 — I ran a pod-grep against a path with no binary in it, and 0 looked like a pass (bug 003)

**The claim I was about to make.** That `git-adapter` v1.0.1167 carried bug 003's F3 change,
evidenced by `strings /app/git-adapter | grep -c 'Consume() called'` returning **0** — the
removed literal being gone is the discriminating marker this repo insists on.

**Why it was worthless.** There is no binary at `/app/` in that image. `/app` holds only
`configs/`. The process runs `./git-adapter` from cwd `/root`, which is `drwx------ root` and
unreadable as `appuser`. So the grep read nothing, and `grep -c` on nothing is 0 — the same
answer as success.

**What caught it** — and it was luck-adjacent, so the lesson is to make it not be. I had put
positive controls in the same loop: `FetchMessage`, `CommitMessages`, `HandleGitCommit`.
**All three also returned 0**, and they must be in that binary. Three impossible zeros is what
exposed the path, not any suspicion about the fix.

**The cheap check.** *In the same command*, grep a string that MUST be present. A
removed-literal grep on its own is **structurally unfalsifiable**: 0 means "the fix shipped"
and "I pointed at nothing" and "the container has no shell utilities", and you cannot tell
which. Add `readlink -f /proc/1/exe` or `ls -la` on the path before believing any count.

**Standing tally — this is the THIRD vacuous pod-grep logged, and they are not the same
mistake twice.** `052`: grepped a marker the change merely *used*, so any image passed.
`049`: published an unfalsifiable pod-grep as a bug file verification step. This one: correct
marker, correct discriminating direction, **wrong path**. The class survives each specific
fix because the failure is always "the command could not have returned anything else", and
each time it wears a different costume. **The general form of the guard is a positive control
in the same breath as the assertion** — not a better choice of literal, which is what the
last two entries concluded and which was evidently not enough.

**Second, smaller call the same session.** `scripts/who-owns.py 075` said this workstream
owned 075, so I planned to build its fix. Another session had already written it, uncommitted
and invisible to a tool that reads commits. What stopped me was the **Edit tool refusing a
stale write** — a `cat >>` or `sed -i` would have clobbered a live fix silently. That is the
argument for CLAUDE.md's Write/Edit preference stated as a near-miss rather than a rule, and
it is the second time the who-owns blind spot has cost a plan cycle (078 was the first).

Family: unfalsifiable-check-reads-as-a-pass, positive-control-in-the-same-command,
zero-means-three-different-things, who-owns-reads-commits-not-working-trees,
the-tool-that-refuses-a-stale-write-is-the-guard.

---

## 2026-07-26 — I closed a bug on the word "extinct", from a day that had six hours left (bug 003)

**The claim.** In the CLOSED box of `bugs_closed/003`, in the 016b index row, and in the
commit message: *"0 on 07-26"* and **"the symptom is extinct"**.

**What was actually true.** I ran the count at **17:53**. The next
`AWAITING_RESPONSES → FAILED` landed at **17:56** — three minutes later — and three more
followed by 20:55. The real post-fix figure is **0.46/h against a pre-fix 2.24/h: a ~80%
reduction, not extinction.** I found it myself, an hour after publishing, only because I ran
one last sanity query before writing the summary.

**Two distinct errors, and the second is the more dangerous.**

1. **A same-day count is a partial count.** "0 today" at 17:53 is a statement about 74% of a
   day, written in the present perfect. The fix is trivial: say *"0 as at 17:53Z"*, or use a
   closed window that has actually closed.
2. **`orchestration_state_audit` is pruned too — about 66 hours.** Between my two
   measurements, four hours apart, the pre-roll window shrank from **38.9 h / 91 events** to
   **31.2 h / 70**. I had already noted that `orchestration_states` keeps ~13 days and
   congratulated myself on finding a table that reached further back. It reaches back less.
   **The before-picture was evaporating while I wrote about it**, and neither figure can ever
   be reproduced.

**The checks.** Record a **rate**, not a count — a rate survives the window shrinking, a count
silently becomes a different measurement. **Take the baseline the moment you realise you will
need one**, because on this platform every history table is on a retention clock and the
pre-fix world is the part that expires first. And treat *extinct*, *never*, *always* as
claims requiring a closed window, not an impression from a quiet afternoon.

**Why it did not change the verdict, stated so the correction is not read as bigger than it
is.** 003 closes on four root causes being fixed and live, each verified independently —
not on this statistic. An 80% rate reduction with a known, named, still-inert residual
(`075`) supports the same conclusion the wrong number did. **That is exactly why the error
was easy to make: it agreed with everything else.** A figure that confirms what you already
believe gets checked least, which is the argument for checking it most.

Family: partial-day-counted-as-a-whole-day, the-baseline-table-is-also-on-a-retention-clock,
record-a-rate-not-a-count, extinct-needs-a-closed-window,
the-figure-that-agrees-with-you-gets-checked-least.

### 2026-07-26 — bugfix_077 — "the council submission died in the ~300s post-restart window"

**Asserted:** that a council submission which vanished without an orchestration row was
eaten by CLAUDE.md's documented "no dispatch within ~300s of a chassis pod (re)start" rule.
The pod's `startTime` was 18:35:07Z and I had published 2–4 minutes later, so the timing fit
a mechanism that is genuinely real and genuinely documented. I wrote it into the closed bug
file, into a fresh instance on `bugs_open/029`, into a memory topic file, and into the
session summary — four places, stated as the cause.

**Actually:** later the same evening I fired five `kubectl -n kafka run -i --rm … kcat -P`
publishes at a chassis that had been up for 10–20 minutes, nowhere near a restart. **Four of
the five silently produced nothing** — no orchestration row, no chassis log line matching the
correlation, `exit 0`, and the wrapper printing a correlation id and "pod deleted" exactly as
it does on success. The `097` council trigger publishes the same way
(`097_TRIGGER_council_review_v1.sh:121` — `printf '%s\n' "$PAYLOAD" | kubectl -n kafka run -i
--rm … kcat -P`). So "the publish never happened" explains the vanished round at least as
well as the restart window does, and it is the one I have direct evidence for. I have not
established either; what I got wrong was writing one of them down as settled.

**Caught by:** chasing a different bug entirely. Three dispatches to leopardess vanished
while one to finetuning had worked, so I started building a site-specific theory — then ran
the A/B (fire both, same minute) and **both** failed, which killed the site theory and moved
the fault to the publisher. The A/B was luck as much as method: I ran it to test the wrong
hypothesis.

**The cheap check that would have caught it:** **make the publisher confirm itself.** Put the
payload in the container COMMAND rather than on stdin and append `&& echo PUBLISH_OK`. One
line. `kubectl run -i` attaches stdin asynchronously; if the container reaches `kcat -P -c 1`
first it sees EOF, produces nothing, exits 0, and `--rm` tidies away the evidence. With the
marker, the very next publish landed first time — and so did every one after it.

**Cost:** four lost dispatches (~25 minutes), a wrong cause propagated into four documents,
and a near-miss on a site-specific theory I would have written up if the A/B had gone the
other way.

**The transferable part, and why it is worse than an ordinary wrong call:** the timing
coincided with a real, documented failure mode, so the evidence felt confirmatory rather than
merely consistent. A documented mechanism you already believe in is the easiest thing in the
world to over-fit a single observation to — and "no rows" is consistent with *every*
hypothesis, which is exactly the trap already recorded in `016b` §9 ("A queued orchestration
is indistinguishable from a dropped one"). The fix is not more inference. It is instrumenting
the step that has no receipt.

Family: named-the-cause-before-isolating-it, silent-failure-with-exit-0, no-receipt-no-claim,
documented-mechanism-overfit.

---

## 2026-07-26 — "I ran assemble-only" — I ran a full section re-render, on two live sites

**The claim:** that the two page-rerenders I fired at ai-agent-orchestration.com and
relojistas.com were **assemble-only** — the cheap branch that stitches stored
`rendered_html` and cannot pick up a template change. I said so to myself while choosing
the command, and it was the whole reason I judged the dispatch safe.

**What was actually true:** both ran `section_data_resolved`, re-rendering every section
from `content_data` through the current templates.
`docs024/oufe/TRIGGER_rerender_page.sh` reads `REASON="${3:-section_data_resolved}"`. Its
header documents "No reason -> assemble-only", so I passed `""`. But `${3:-word}`
substitutes the default when the parameter is unset **or null** — an empty string is null.
Passing a placeholder to mean "none" selected the heaviest branch on offer.

**What caught it:** the evidence contradicted the story. `page_components.updated_at` moved
to 21:16 on both pages, and an assemble-only run does not touch those rows. I had expected
the injected `data-component-css` marker on the page and found the component's own style
block instead.

**The cheap check that would have:** one line, before dispatching —
`bash -c 'f(){ echo "${3:-DEFAULT}"; }; f a b ""'` — or simply reading the line of the
script I was about to run. I read its 35-line header and not its parameter defaults.

**Cost:** none, by luck rather than judgement. The script's own guard refuses when any
section has NULL `content_data` (which would escalate to the content writer and regenerate
copy); both pages happened to be clean, so nothing was rewritten. Had either carried a NULL
I would have silently regenerated authored copy on a live customer site. The heavier branch
also happened to be the one that made the fix visible, which is the most dangerous kind of
accident: it rewarded the mistake.

**The transferable part:** `${VAR:-default}` and `${VAR-default}` are different operators,
and every "pass empty to mean none" idiom silently picks the default under the first one. A
script whose *documentation* says "no reason" and whose *code* says `:-` cannot express
"no reason" as an argument at all — the only way to get it is to omit the parameter. When a
flag's absence and its emptiness must differ, the script has to use `${3-...}` or an
explicit `[ $# -ge 3 ]`, and the caller cannot tell which from the header comment.

Family: read-the-header-not-the-code, empty-is-not-unset, rewarded-by-luck.

---

## 2026-07-26 — I published a row count I never counted (bugs_open/075 council submission)

**The claim.** In the round-2 council rationale, and in two docs, I wrote that
`processing_node` "across all **1,547** rows" resolves only to single-pod
services. The census was real and the conclusion was right. The number was not:
I had run a `GROUP BY` census, read the per-service counts, and then wrote a
total I had never asked for. The true total, queried afterwards, is **1,644**
(1,282 agent-chassis, 5 business-intel, the rest spawned single-pod Job agents,
378 distinct pod names).

**What caught it.** Adding up the per-service counts in my own earlier output —
after the submission had gone out. Nothing in the system would have: a wrong
total inside a correct conclusion never fails a query, and the council cannot
open the database to check it.

**The cheap check.** `SELECT count(*)` is one line, and a GROUP BY census does
not contain its own total. If a number is worth quoting to a reviewer, it is
worth a query of its own; if it is not worth a query, quote the shape ("every
row resolves to a single-pod service") and leave the number out.

**Why this one matters more than it looks.** The figure was load-bearing
rhetoric in a safety argument — "we checked ALL 1,547 rows" is precisely the
kind of specificity that makes a reviewer stop checking. Inventing precision to
sound thorough is worse than saying "the census returned only single-pod
services", because it converts a verified claim into an unverifiable one while
making it read as more verified.

**Standing tally, uncounted-figure family:** this is the same shape as the 040
"240 dials / 7 days" call (a 20-minute window quoted as a week) and the 052
population table carried forward unchecked — three now. The recurring check is
identical: **the aggregate you are about to quote must be the aggregate you
actually ran.** Family: fabricated-precision, aggregate-not-run,
right-conclusion-wrong-number.

**RESOLVED 2026-07-26, by the shipped code rather than by argument.** Discovery ran on
webdesign.co.uk with the partitioning check live: it filed a `capability_gap` with
`population=2, residue=2` — i.e. a true remit of **0**, precisely what `bugs_open/077`'s
table said and what my over-approximating SQL (which returned 1) could never have
established either way. Three independent measurements now agree: their Go transform over a
`row_to_json` dump, the shipped detector's own partition, and — by not contradicting either —
my superset. The entry above stands as written; this is the confirmation that the
"correction" I nearly published would have been the only wrong number in the set.

---

## 2026-07-26 — "no other lane has touched it", stated while another session was four minutes into the same work

**The claim.** Picking up `bugs_closed/076` R1, I ran `scripts/who-owns.py 076`
as the rule says, got the `truncation_contract_076` workstream ACTIVE with one
commit (the handoff itself) and no competing lane, cross-checked `git log`, and
told the user **"Ownership is clear — that directory is this workstream's, and no
other lane has touched it."** I then wrote a PLAN, a NOTES and a RUNBOOK on that
basis.

**What was true.** Another session had picked up the same residual off the same
handoff and had written its PLAN, NOTES and `README_where_we_are` into that
directory **four minutes before I started writing mine**, and had `scripts/
truncation_registry.py` on disk before I finished reading the guard. Two
independent PLANs for R1 existed simultaneously, with different filenames, in one
directory.

**What caught it.** Not a check — the `Write` tool refusing
`README_where_we_are.md` as unread. I had listed that directory twenty minutes
earlier, seen one file, and carried "the directory holds one file" forward as if
it were still true.

**The cheap check.** `ls -lat` on the workstream directory immediately before
creating anything in it. Three files with timestamps inside the last five minutes,
one second of work, before three documents were written instead of after.

**Why `git log` could not have saved it, and this is the new part.** The existing
practice note says re-run `git log` at implementation start. Here that would have
returned exactly what it returned at session start, because **the other session
had committed nothing at all.** Both `who-owns` and the log read commits; a lane
that is two hours in and pre-commit is invisible to both *by construction*. The
only signals available were filesystem mtimes and the Write refusal.

**Second, smaller call in the same session.** In that deleted PLAN I wrote that
`177_council_tolerate_truncation.sql` is *"the only seed that has ever set
`tolerate_truncation`"*. There are **three** (`177`, `PATCH_fix_proposer_021`,
`PATCH_feature_designer_022`). My first grep was repo-wide and showed all three;
I then re-ran it narrowed to `sql_for_agents/` and quoted the narrowed result as
a repo-wide claim. The design conclusion survived — all three are `jsonb_set`
patches, so the point about pre-commit checks holds — but the count was wrong and
would have been quoted onward. **Cheap check: the grep whose result you quote must
have the scope your sentence claims.** Same family as the aggregate-not-run
entries above: right conclusion, wrong number, stated at full confidence.

**Standing tally, who-owns-is-blind family:** `bugs_open/078`, `bugs_open/073`,
and now `076` R1 — three in one day. The first two were caught by a changed row
and a one-minute-old mtime; this one by a tool refusal. **The recurring check is
now unambiguous and is not `git log`: it is mtime on the artefacts you are about
to touch.** That is mechanical, it costs one command, and it is the third time it
would have paid. Worth automating into the `who-owns.py` output itself — it
already knows the workstream directory; it could stat it.

---

## 2026-07-26 — three wrong calls while BUILDING a checker (076 R1)

All three were in the checker itself, which is the point: a check that is wrong
is worse than no check, because its silence reads as evidence.

**1. I stated a mirror of `GetBoolField` without opening it.** `103_LINT`'s first
draft said its `is_true` "mirrors datahelpers.GetBoolField: a JSON true, or the
string `"true"`" — and I wrote a *fixture asserting the string form*, which is
how a guess gets promoted to a test. `data_helpers.go:1570` type-asserts
`m[key].(bool)` and returns the default on anything else. So a step configured
`"tolerate_truncation": "true"` gets **no tolerance at all**; flagging it as an
offender would have been a false positive on config that is already failing
closed. **The cheap check: open the function you claim to mirror.** Twenty
seconds. Caught only because writing the fixture sent me to read it.

**2. My scanner silently covered 166 of 171 workflows.** `load_rosters()` keyed a
dict by agent `type`. Five live types have TWO active rows (`chief-strategist`,
`content-creator`, `content-creator-contact`, `multipage-website-builder`,
`site-component-architect`), so one row of each pair was overwritten and never
scanned — and a duplicate row is exactly where stale or hand-edited config hides.
Caught because the tool printed its own denominator and it disagreed with a
`SELECT count(*)` I had run minutes earlier. **The cheap check: make a tool print
what it scanned, then reconcile that number against the source once.** A tool
that reports findings but not coverage cannot be checked at all. Same family as
`bugs_open/098` — every guard excludes a population, and nobody owns the
intersection — except this exclusion was not even deliberate.

**3. My parser dropped a registry entry with no error** (found by another session,
reproduced by me before accepting). The entry regex required key, value and comma
on one line; a `mechanism` string long enough for gofmt to wrap is
`"part one " +\n\t\t"part two"` and matched nothing. Three entries present, two
parsed, exit 0. Direction matters: a *missing* reader makes correctly-guarded
workflows look like offenders, so the lint cries wolf on a clean fleet — the
failure mode `pattern-check.py`'s own header says is fatal. **The cheap check: a
parser must compare what it found against what is structurally there** (count the
keys, compare with the parsed entries) — "it returned something" is not a parse.

**The tally these three add to:** two of the three are the same shape as entries
already above — a claim stated at full confidence with the one-command check
skipped (#1), and a coverage denominator nobody reconciled (#2). #2 is the second
this week where the bug was in *what a tool did not look at* rather than in what
it concluded. Worth noting that all three were caught by mechanical means —
a fixture, a printed denominator, another reader — and none by re-reading my own
code, which I did several times.

---

## 2026-07-26 — I recommended a config change without checking the config key is read (vetcomparison P1)

**The claim.** Having measured that widening the vet verifier's scrape to legal/terms pages would
lift the company-number hit rate from 4/25 to 7/25, I wrote — in `NOTES_vetcomparison.md`, in
`README_where_we_are.md`, in a `PLAN` correction block and in a **commit message** (`096276f90`)
— that this was *"a config-only win"*, *"a DB config change: live immediately, no build, no
roll"*. I named the exact key: `scrape_website.config.follow_links`.

**What was true.** `follow_links` is not read by any Go code in this repository. Nor are three
other keys on the same step: `max_pages`, `extract_mode`, `fallback_url_field`. `WebscrapeAction`
(`webscrape_actions.go:27-147`) honours `url_field`, `url`, `action`, `upload_results` and
`scrape_config`, resolves **one** URL and dispatches it. The step is written to read like a
six-page crawl with a search-result fallback and is a single homepage GET.

**What caught it.** Not review — I had already committed. I went looking for *how* `max_pages: 3`
would interact with a longer `follow_links` list (would extra paths displace the useful ones?),
which meant finding the implementation. The grep that answered that question returned no hits at
all, which was not the answer I was looking for. **I was saved by asking a follow-up question
about the mechanism, not by checking the claim.** Had `max_pages` not made me curious, the false
claim would have stood in four documents and a commit.

**The cheap check.** One command, before calling any config change a win:
`grep -rn "<the_key>" --include=*.go .` — if the key is not read, the change is not a change.
Cost: seconds. It is the same check as "read the whole seed before applying it" (the seed-037
landmine, which this workstream already had written down) pointed at config keys instead of SQL.

**Why this shape recurs.** Config that is *read* and config that is merely *present* are
indistinguishable by inspection — unknown keys are silently ignored, so a stale or aspirational
key looks exactly like a live one, and it reads as documentation of behaviour while being
evidence of nothing. This is the same family as the inert-layer entries already in this file: the
artefact exists, so the capability is assumed. **Direction of error matters here** — it made me
*understate* the cost of a fix (a "free" tweak that is actually council + build + roll), which is
the direction that gets a fix scheduled as trivial and then found to be a deploy.

**Tally note.** This is another instance of the file's most common row: *a durable claim written
at full confidence with the one-command check skipped*. The variant worth flagging is that the
claim was about **config being live immediately** — the repo's own CLAUDE.md advertises "DB config
is live immediately; prefer it" as the fast path, and that advice quietly assumes the key is
wired. The fast path is only fast if something reads it.

## 2026-07-26 — I read a stalled queue as a dropped dispatch, and the handoff I wrote already disproved it (bugs_closed/040 candidate 2)

**The claim.** In `bugfix_040_partial_build/HANDOFF_2026-07-26_continue_here.md` §4, under a
heading in capitals — **"THE UNSOLVED PROBLEM — read before re-firing"** — I wrote that five
publishes of a scratch probe had produced **zero** `orchestration_states` rows, tabulated six
hypotheses ruled out "each by a check", and left the next session three more to work through
starting with a column-by-column diff of the agent row against a known-working one. The framing
was that something about the agent definition was silently rejecting the dispatch.

**What was true.** Nothing was rejecting anything. The messages were sitting in the topic behind
a **stalled consumer**, and they ran normally the moment it cleared — at **21:16:09 and
21:16:16 UTC, one minute after I wrote the handoff saying they never would**. Both assertions
passed on that run. The agent definition was correct from the first attempt.

**What caught it.** The next session opening the handoff and, before working the hypothesis list,
running the one query that asks what actually happened rather than why it didn't:
`SELECT status, error FROM site_work_items WHERE item_type='scratch_cand2_probe'` — which came
back `failed` with the routed error and `complete` with a blank, i.e. the experiment had already
succeeded.

**The cheap check.** *Re-read the outcome table before theorising about the cause.* One SELECT
against the thing the experiment was supposed to change. Cost: one query, three seconds. It would
have replaced five re-fires, six ruled-out hypotheses and a section of hand-off debt with a line
saying "passed".

**Why this shape recurs, and the part that stings.** §5 of the *same document* diagnoses the
stall, names the frozen committed offset (**105196**), and prints the backlog — in which my own
probe messages are sitting at offsets **105197, 105202 and 105204**, immediately behind it. I had
the evidence, formatted, two sections below the problem it solved. The failure was not missing
information; it was **not joining two findings inside one document** because I had already
committed to a frame ("the dispatch is being dropped") and §5 was filed under a different story
("found on the way — an unrelated lane stall"). A frozen queue and an eaten message are
indistinguishable from the publisher's side, so the frame was never tested — every subsequent
check was aimed inside it.

**Tally note.** Distinct from the file's most common row (a claim written with the check skipped):
here **the check had been run and written down**, in the same file, by me. The variant to watch is
**a section headed "found on the way"** — incidental findings get filed as digressions, and a
digression is never re-read as evidence for the main problem. Worth a habit: when a document
contains both an unexplained failure and an unrelated infrastructure fault in the same window,
those are the same paragraph until proven otherwise.

---

## 2026-07-26 — I corrected an over-claim and over-claimed in the opposite direction in the same breath (vetcomparison P1, second entry this session)

**The claim.** Having just corrected "config-only win" (entry above), the correction itself
asserted: *"the 16% is exact, not approximate — homepage-only **is** production."* Written into
NOTES, README, a PLAN block and commit `9d35f719a`.

**What was true.** Two things, and the second is the real one.
1. It was 16% on n=25 and **22% on n=100**. A point estimate from 25 samples was never "exact".
2. **I had not read the last leg of the pipeline.** My probe fetched raw HTML and stripped tags.
   Production goes through the webscrape adapter to **Firecrawl**, whose `onlyMainContent` is set
   `false` in code but **only added to the payload when true** (`firecrawl.go:77-111`) — so a
   caller passing no `scrape_config` (the vet verifier passes none) has the key omitted and
   Firecrawl applies its own default. If that default strips footers, production sees **less**
   than my probe, and company numbers live in footers. I asserted equivalence between my probe and
   production having read only the two components nearest me.

**What caught it.** Checking whether widening `follow_links` would even help — which meant
following the request one hop further than I had. The same shape as the previous entry: the error
surfaced from a *follow-up question about the mechanism*, not from re-reading the claim. Both times
the claim read fine on re-reading; it was only wrong against something I had not opened.

**Resolution: I left it UNSETTLED rather than guessing.** 2,452 stored Firecrawl samples split
ambiguously — 75% retain footer nav (suggesting footers survive), 0 contain registration text
(equally explained by page type). It is recorded as `[UNSETTLED]` with the one-run check that
answers it, at the top of `bugs_open/101` where it gates the fix. **Not-knowing, written down
where the decision is made, is a legitimate output** — the failure mode was never uncertainty, it
was uncertainty dressed as a finding.

**Why this one is worth the tally.** The two entries are minutes apart and opposite in direction:
first I made a fix sound cheaper than it was, then I made a measurement sound firmer than it was.
**Having just been corrected is not evidence about the next claim** — if anything it primes a
confident-sounding correction, because a correction reads as the careful, chastened version. The
cheap check is unchanged and mechanical: *for any claim of the form "X is what production does",
name every component between you and production, and say which ones you have opened.* I had opened
two of three.

## 2026-07-26 — I told the owner the council had no pre-build review point, while three rosters ran and two fired before any code existed

**The claim.** Asked whether we had ever discussed an architecture council member, I researched
the council gate, read the guardian's charter and the RFC track, and then argued in chat that a
forward-looking architecture seat *could not* live at the council at all: *"the council reviews a
plan that already exists — by then the shape is decided, so an architecture seat there can only
ever say no. The forward half has to be asked before a plan exists."* I built a design proposal on
top of that, whose central move was that the seam for a pre-plan review did not exist and the RFC
track had to supply it.

**What was true.** The owner corrected it mid-turn in one line: *"the council is also used in
advance of the build too - and the diagnosis loop."* Three council rosters exist, at three
lifecycle points, and two of them fire before any code is written:
`experience-planner` (seats: journeys, contracts, honesty, **feasibility**, **mvp**) at plan
composition; `feature-designer` (guardian, editquality, bug_historian, guidelines, reuse_agent)
at design time from an owner-approved capability spec; `fix-proposer`/`council-gate` (16 seats) at
edit-plan time. The seam I said was missing is where the guardian *already sits*. Worse for the
claim: `feasibility` and `mvp` are forward-looking seats that already do a version of the job I
was arguing had nowhere to live.

**What caught it.** The owner, from memory of his own platform. Not a check I ran.

**The cheap check.** One query, widened by one clause from the one I actually ran:
`SELECT type, key FROM agent_definitions d, jsonb_object_keys(d.default_config->'workflow'->'steps') key WHERE key LIKE 'review_%';`
— i.e. **list the seats across ALL agent types, not just the one named `council-gate`**. Cost:
one query, five seconds. I ran the `type='council-gate'` version four times while researching.

**Why this shape recurs.** The search term was the answer's name. "Council" resolved to the agent
literally called `council-gate`, and once that returned a rich, coherent 16-seat structure it read
as *the* council rather than *a* council — a complete-looking answer stops the search. The
platform's own vocabulary made this easy: CLAUDE.md documents "the council gate" in the singular,
so the singular was never questioned. Note the direction of error: it made me propose **building**
a review point that already existed, which is the reuse failure the `reuse_agent` seat exists to
catch — I made it while reading that seat's own charter.

**Tally note.** A new variant for this file. The common row is *a durable claim with the
one-command check skipped*; the recent one is *the check was run and written down but not joined
up*. This is a third: **the check WAS run, correctly, and its scope silently defined the
conclusion.** `WHERE type='council-gate'` is not a wrong query, it is a narrow one, and a narrow
query returns a confident answer about a small world with nothing in the result to indicate the
world was small. Worth a habit: when a query filters by a name you took from the question, run it
once without that filter before building on the result.

---

## 2026-07-26 — I wrote "no scanner can catch this" into a live council seat, and it was false

**The claim.** Having found a site publishing a promise of its own infallibility, I
concluded the class was structurally invisible to the platform and wrote that into
migration 223 — into the *contract text* of the compliance council seat and again
into its judge clause:

> "it is the one class every scanner misses … invisible to all of them"
> "remember no scanner will catch this, so this seat is the only control"

I repeated it in a bug-adjacent workstream note, a milestone summary, a decisions
register, a memory file, and two commit messages, and I built a design
recommendation on top of it.

**What is actually true.** `ScanBannedClaims`
(`platform/orchestration/datahelpers/claims.go:284-325`) is a bare
case-insensitive regex over prose blocks. It contains no number extraction, no
`businessClaimContextRe`, no `isExcludedNumber` — those gate only
`ScanUnregisteredNumbers` (`claims.go:365,369`). It catches whatever patterns a
site has been given, about anyone, numeric or not. Live registers already carried
purely qualitative patterns — `leaderboard`, `live now`, `price target`, `years of
experience`. **The capability had been there the whole time. No pattern for this
class had ever been written on any site, and there is still no way to write one
once for the fleet.** Coverage gap, not capability gap. Arming it took one UPDATE
and no image roll, and bought a build-time *blocker* plus a high-severity
post-deploy finding.

**What caught it.** The owner. He read the proposal, said *"we have existing
functionality that double checks claims, and we have the council — look hard at
our existing documentation and solutions"*, and would not take the new-machinery
answer. Research then took ten minutes and the crux was one function body.

**What should have caught it earlier — the check I skipped.** I had *already read
the limitation that misled me*. Earlier the same day I correctly established that
`ScanUnregisteredNumbers` is inert on finance prose, and wrote it up accurately.
Then I let "the number scanner cannot see this" become "the scanner cannot see
this", and never opened the sibling function twelve lines away. **A limitation
established about one component does not transfer to the component next to it, and
the cheap check is to read the function you are about to write off** — thirty
seconds, and I spent an afternoon building on the assumption instead.

There was also a documented answer I never looked for.
`SPEC_claims_verification.md:250-252` records this exact open question — *"Should
`banned_claims` be fleet-shareable (some patterns are universal) … Proposal:
per-site only until two sites have evidence bases"* — deferred when n=1, and its
precondition lapsed long ago at n=8. **Grep the spec for your problem before
concluding the platform has no answer to it.**

**Why this one is worse than an ordinary wrong call.** It is not a status line or
a handoff sentence — it went into the **standing instructions of a live reviewing
agent**, where it would have shaped every future compliance verdict by telling the
seat it was the only control and inviting it to substitute for a mechanism instead
of asking for one. And the content of the error is precisely the failure the seat
exists to catch: **a confident, unverified claim about what our system guarantees,
written by the person building the overclaim detector.** Corrected in migration 227
and mirrored to both rosters.

**The bit worth remembering.** "Nothing in the estate does X" is a claim about the
whole estate, and it is enormous — far larger than the evidence I had, which was
one component's documented limit and my own failure to find a pattern. **A
universal negative about a large system needs a search, not an inference**, and the
tell that I had not done one is that my sentence named four mechanisms and I had
read the source of exactly one of them. When you catch yourself writing *nothing*,
*never*, *no*, or *the only*, that is the sentence to go and disprove.

Family: universal-negative-from-local-evidence, limitation-bled-across-components,
the-spec-already-answered-it, overclaiming-while-building-the-overclaim-detector.

---

## 2026-07-26 — a third, while cleaning up the second: I declared my own entry "clobbered" on a grep for a string that was only ever in the commit message

**The claim.** That the WRONG_CALLS entry above had been **destroyed by a concurrent session**
between my `cat >>` and my `git commit`. I appended a duplicate copy carrying a confident
operational moral — *"append-then-commit-later is not safe in this tree"* — and committed it.

**What was true.** Nothing was lost. My append had already been **committed by another session**:
`cbd020a55` (*"docs(arch-review): decisions open for owner…"*) swept the file in, which is why my
own `git commit` correctly reported *"no changes added to commit"*. That is the `git add -A`
passenger case CLAUDE.md documents — **a real event, and the opposite of destruction.**

**How I got there — the same vacuous-grep failure twice in one night.** I checked for survival with
`grep -c 'opposite direction, minutes after the first'`. That string was in my **commit message**,
never in the file, whose heading reads *"…in the opposite direction in the same breath"*. The grep
could only ever return 0. I had **already logged this exact failure mode twice tonight** (the
`kubectl run -q` consumer, §above) and had written the fix down — *put a positive control in the
same command* — and then did not apply it, because this time I was not "verifying" anything, I was
just "checking whether my file was there". **The habit only fires when I think of the command as a
measurement.** That framing gap is the actual defect.

**Cost.** A duplicated entry in an append-only file, a false operational moral committed to a
fleet-wide doc, and this cleanup. Cheap in tokens, expensive in credibility: the false version
read as a hard-won lesson, which is exactly the kind of line a later session would adopt.

**The check, stated so it fires on "is it there?" and not only on "is it correct?":** *grep for a
string you can see on screen in the artefact itself, and in the same command grep a control you
know is present.* If both return 0, your grep is broken, not the world. **Never grep a string you
composed in a different buffer** — commit messages, plans and prose about the file are not the file.

Family: vacuous-grep-no-positive-control (3rd instance tonight), string-from-the-wrong-buffer,
sweep-is-not-destruction, moral-drawn-from-an-unverified-mechanism.

---

## 2026-07-27 — a pod-grep control asserted a COUNT, and the count is not ours to predict

`bugs_closed/076`'s re-verification recipe (its handoff §6) told the next thread
to run `strings /app/agent-chassis | grep -c "tolerate_truncation"` and **expect
3**. On chassis v1.0.1172 it returns **4**, with the 076 source byte-identical
since `511670fc8` (empty `git diff` across all three files).

`strings` does not emit one Go string literal per line. The linker packs string
data into blobs and `strings` prints each blob, so `grep -c` on a common substring
counts **blobs that happen to contain it** — and the split points move between
builds for reasons that have nothing to do with your change. Piping the same grep
through `cut` shows it plainly: one "line" carries `stop_reason=refusal`, a heap
metric, an SQL `DO NOTHING` and an OCI runtime description.

Direction matters, and it is why this is worth an entry rather than a shrug. The
false claim here does not hide a defect — it **invents** one. A control that reads
FAILED on a correct binary sends the next thread hunting a regression that does
not exist, on the exact day they are trying to establish whether a roll broke
something. That is expensive precisely when time is short.

**The cheap check: assert PRESENCE (`>= 1`) plus a negative control at 0, never
equality on a count.** If you want a number to mean something, count something you
control — rows you wrote, a symbol you named — not a substring in a linker's
output.

**Tally, pod-grep family.** This is the third entry: *a removed-literal grep is
unfalsifiable without a positive control in the same command* (twice), and now
*a positive control must not be an equality assertion*. The shape underneath all
three is the same — **the pod-grep is a test, and a test with no known-good and
known-bad answer is not a test.** Both halves have now failed in the wild, one
by being absent and one by being over-specified.

### 2026-07-27 — webdesign.co.uk — "97 pages transformed, 0 warnings" while 60 tools shipped dead
**Asserted:** that the port had converted every page correctly. The transform printed
`transformed 97 pages, 27 assets, 0 warnings` on every run, each count was right, the
fragments were well-formed HTML, and the manifest was complete.
**Actually:** **every tool's JavaScript had been silently discarded.** Both source sites put
their `<script>` tags after `</main>`, at body level; the transform chose its content root
(`<main>`) *before* extracting scripts, so it harvested only what was inside it — nothing.
`tool-bayesian-rank` had zero scripts and `bayes.js` was not even in the manifest. Around 60
of 63 interactive tools would have published as static markup: correct-looking pages that do
nothing when you click them.
**Caught by:** luck. Grepping one fragment for `<script` while chasing an unrelated question
about sibling assets. No gate, no warning, no failing count — and a browser would have been
needed to notice it in production.
**The cheap check that would have caught it:** compare a **structural property of the output
against the source**, not just count the outputs. Scripts in ⇒ scripts out. Now enforced as
`checkScriptParity` (`cmd/webdesignport/transform.go`), which **fails** rather than warns —
and was proved by re-introducing the original bug in a scratch build, producing 60 failures,
one per dead tool. A gate only ever seen passing has not been tested; it has been observed
not complaining.
**Cost:** nothing, by luck alone. Had it shipped, the failure is invisible to every
DB-side check (`build_status='deployed'`, artefact present, HTML valid) and would have been
found by a human clicking. Filed the platform-wide version as `bugs_open/084`.

### 2026-07-27 — webdesign.co.uk — committed with `git add <directory>` in a shared tree
**Asserted:** implicitly, that `git add bugs_open` staged my work. The commit message
described a JavaScript-verification correction.
**Actually:** it also swept in two files belonging to another live session — their move of
`bugs_open/006` to `bugs_closed/` (580 lines) and their new `bugs_open/088` (171 lines).
751 lines of someone else's work now sit under a commit message about JavaScript, where
`git log --follow` and `git bisect` will not expect them. Both files are intact; nothing was
lost.
**Caught by:** reading the `create mode` / `delete mode` lines in git's own commit output —
a habit, not a mechanism. The yellow commit-scope block had printed exactly this and I had
not read it.
**The cheap check that would have caught it:** `git status --short <dir>/` before adding, or
simply reading the commit-scope block that the pre-commit hook prints for this purpose.
CLAUDE.md forbids `git add -A` / `.` / `*`; **a bare directory is the same mistake wearing
different clothes**, because in a shared tree a directory is not a unit of work, it is a
shared namespace.
**Cost:** unclean history for another thread; forward-only, so it stands. Recorded rather
than repaired.

### 2026-07-27 — webdesign.co.uk — invented a statistic in the act of removing invented statistics
**Asserted:** replacing the about page's two unmeasurable claims ("100 Lighthouse score",
"0.1s First Content Paint" — figures nobody has measured for this domain) with facts, I
hand-typed "**64** Tools".
**Actually:** there are **63**. Worse, the same 64 had originated in the mission brief I
wrote before the catalogue existed, and had already propagated into **eight live specs** —
`identity.about_us`, `strategy.value_proposition`, and the briefing the page planner reads.
The home page would have opened by advertising a tool that does not exist.
**Caught by:** running `jq '.tools|length'` over the generated catalogue for an unrelated
reason. Then a *second* instance of the same class: after replacing the literal with a
`{{TOOL_COUNT}}` placeholder, the substitution ran **before** the rewrite that introduces the
placeholder, so a literal `{{TOOL_COUNT}}` reached the page. Exit code 0 both times.
**The cheap check that would have caught it:** **derive the count from the artefact; never
type one.** A number that cannot be typed cannot drift. Corrected in the DB by
`SQL_p4_fix_tool_count.sql` before the planner ran.
**Cost:** none live, caught in time — but only because the airlock held the planner back long
enough to look. Had the cascade run unattended, a fabricated figure would have been the first
sentence on the home page.

### 2026-07-27 — webdesign.co.uk — reviewed the planner's page LIST and called the plan checked
**Asserted:** that the planner had obeyed the mission brief, because it produced exactly one
page (`/index.html`) as instructed rather than inventing a site.
**Actually:** it had also chosen the *sections* for that page, and picked a full-bleed hero
painting `linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.6))` over a background image with
`--hero-ink:#fff`. A dark hero — contradicting both the brief ("two-column hero, copy left,
image right") and `design_intent.avoid` ("Dark backgrounds of any kind"). It shipped live and
the owner spotted it.
**Caught by:** the owner looking at his own site. Not a check.
**The cheap check that would have caught it:** `SELECT sections FROM pages WHERE name='index'`
— one query, at the moment I was already reviewing the plan. **Reviewing a plan's top-level
shape is not reviewing the plan.** Note also what this says about pins: the palette pin held
flawlessly (every colour in the committed `styles.css` is a pinned one), because a pin governs
colour *values* and cannot govern component *selection* — the darkness was a literal `rgba()`
inside the chosen component's own template, drawn from no palette, and `avoid` is prose the
planner's component choice never reads.
**Cost:** a wrong hero live for a day, and a manual fix that turned out to need a
template-render because **no rerender path handles a section whose component changed** —
assemble-only republishes stored HTML, and the data-refresh path correctly does nothing when
no data changed.

### 2026-07-27 — webdesign.co.uk — appended to this file without reading its own row shape
**Asserted:** implicitly, that adding two bold paragraphs at the end of `WRONG_CALLS.md` was
filing a wrong call.
**Actually:** this file specifies a **row shape** (`### YYYY-MM-DD — <thread> — <claim>`, then
`Asserted` / `Actually` / `Caught by` / `The cheap check` / `Cost`) and a **standing tally**
with the instruction *"update it when you add a row"*. I matched neither, so my entries were
unparseable by the file's own conventions and, worse, invisible to the aggregate — which the
file states plainly is the entire point: *"One entry is an anecdote. The value is the tally."*
**Caught by:** the owner asking me to write the missteps into the missteps file, which made
me open it properly for the first time instead of appending to its tail.
**The cheap check that would have caught it:** read the target file's header before writing
to it. A file that documents its own contract in its first 150 lines is telling you it has one.
**Cost:** two malformed entries for a day, reformatted here, and one tally left un-incremented
— which is the part that actually mattered, because a tally row that never reaches 2 never
gets automated.

### 2026-07-27 — brochure_component_library — "only the index rebuild broke links; capabilities is clean"
**Asserted:** that my page rebuilds introduced six broken links, all on the index,
and that the capabilities page came through clean. Written into `bugs_open/071`, a
summary, the owner's log and a commit message.
**Actually:** sixteen, across both pages. The capabilities rebuild produced ten
more — `href="/capabilities#review-council"` and five siblings, extension-less
*with a fragment*, 4 in `hero-card-carousel` and 6 in `info-card-grid`. All ten
404 as served.
**Caught by:** the live crawl in my own verifier, which captures
`href="(/[^"]*)"` and strips the fragment afterwards. The DB check I had trusted
used `href="(/[^"#?]*)"`, which **excludes every href containing `#`** — the exact
anchor-blind pattern this workstream recorded as landmine L2 the previous day,
after it hid 21 broken links behind a census, a repair and a post-check that all
agreed with one another. I read that landmine, wrote it into a handoff section I
authored, and then used the pattern in the next query I typed.
**The cheap check that would have caught it:** never let a character class decide
what counts as a link. Capture the whole `href` and normalise afterwards — the
fragment is data, not a delimiter. More generally: when a landmine names a
specific pattern as dangerous, grep your own new queries for that pattern before
trusting their output.
**Cost:** ten broken links live on the capabilities page for a day, a false
"clean" recorded in four places including a bug file other threads read, and a
correction pass across all of them. The gate had detected all sixteen and shipped
them as warnings, so nothing else was going to catch it.

### Addendum, 2026-07-27 — why the existing answer was invisible, which is the transferable bit

Four things existed that I did not find. Naming *why* each was unfindable is more
use than the entry above, because the pattern will recur.

**Code search finds things by their effects.** Grep a function and you get its
callers; grep a work-item type and you get its producers and consumers; open a
table and you find the queries. Anything that *runs* leaves a trail of references,
and that trail is what "look hard at the existing code" actually follows.

**Dormant capability leaves no trail at all**, and every one of the four was dormant
in a different way:

- `EvidenceFact.Kind` (`claims.go:73`) governs no behaviour, so it has zero call
  sites, zero tests, zero log lines, zero work items. It exists in a struct
  definition and a spec paragraph. Grepping for "where does the platform handle
  capability claims" returns nothing — **not because there is no slot, but because
  the slot is empty.**
- The deferred decision lives in `SPEC_claims_verification.md:250-252`, under a
  heading reading *"Open questions for the owner"*. It is not a bug, not a feature
  file, not a TODO in code. I ran the documented prior-art check — grep
  `bugs_open/` and `features_open/` — and it was in neither.
- The precedent (`globalTellPhrases`, `voicetells.go:121-137`, unioned at `:109`)
  sits in the sibling file of the engine I was reading, solving the same
  per-site-versus-fleet problem in the opposite direction. **Nothing in
  `claims.go` references it.** The two are conceptual siblings with no code
  linkage; only a person who knows both files knows they are related.
- The live sweep exists and is correct, but its scheduled task is disabled, so it
  produces no findings, no items, no logs — indistinguishable at runtime from a
  check that was never built.

**So the absence of references is exactly what "does not exist" and "exists but
dormant" both look like from inside the code.** That is why *"nothing in the estate
does X"* is so dangerous a sentence: the evidence that would refute it is, by
construction, the evidence code search cannot surface.

**The correction, and it is a change of instrument, not of effort.** Dormant
capability is found in *design artefacts*, not code: specs and their open-questions
sections, struct comments listing a vocabulary, "deliberately not built" lists,
and sibling subsystems that solved the same shape. Had I read the claims spec
end-to-end — twenty minutes — I would have found the question, the proposal, and
the reasoning, already written and waiting for an answer.

**A second-order finding worth its own line: a deferral with a numeric trigger and
no watcher becomes a permanent decision.** *"Per-site only until two sites have
evidence bases"* was right at n=1. It is now n=8, nothing re-read it, and the
deferral has been silently operating as policy for months. Nothing in this estate
re-examines its own deferred decisions when their preconditions lapse — every
other trigger we rely on (cooldowns, staleness, sweeps) has a watcher, and this
class has none.

Family: dormant-capability-is-invisible-to-grep, absence-of-references-is-ambiguous,
read-the-spec-not-just-the-code, deferral-without-a-watcher-is-a-decision.

---

## 2026-07-26/27 — I said the council submission was "queued, not dropped". It was dropped.

**The claim:** my council submission (`97904892`) had no orchestration row an hour after
submitting. I told the owner this was "the documented queue latency, **not** a dropped dispatch",
and declined to resubmit, citing the standing rule that a missing orchestration row is almost
always latency and retrying costs a duplicate round.

**Why it was wrong:** thirteen hours later there was still no row and no artifact — while **676
orchestrations were created fleet-wide** in that same window. The lane was draining the whole
time. It was a drop.

**What caught it:** coming back the next day and, before repeating the claim, asking what *else*
had run. Nothing else would have caught it — the submission's own row stayed absent in exactly
the way a queued message's row stays absent.

**The cheap check:** one query, `SELECT count(*) FROM orchestration_states WHERE created_at >
'<dispatch time>'`. If that number is large, your message is not queued behind anything. Seconds,
and I could have run it at the one-hour mark instead of the thirteen-hour mark.

**The mechanism of the error, which is the transferable part:** the standing rule is *correct* —
a missing row usually IS latency, and resubmitting on that evidence usually IS a wasted round. I
applied a true heuristic without ever asking what would distinguish it from its opposite. "Absent
because queued" and "absent because dropped" produce **identical evidence at the site you are
looking at**; they differ only in a place I never looked, which is whether the queue is moving.
A heuristic that cannot fail is not a heuristic, it is a habit — and this one had a one-query
falsifier available the entire time.

Compounding it: the same session had *already* logged the near-identical `bugfix_006` finding
("the council submission was DROPPED, not queued"). I had the counterexample in my own memory
index and still reached for the general rule.

**Rule of thumb this earns:** before explaining an absence with latency, measure the throughput
of the thing you claim is slow. An absence is evidence about a *system*, never about your own
item alone.
Family: heuristic-applied-without-a-falsifier, absence-has-two-causes, ignored-my-own-counterexample.

---

## 2026-07-27 — I shipped "1170+ agents" to production while fixing a claim that understated the fleet

**The claim I wrote down:** migration `231`'s post-condition asserted that ai-agent-orchestration.com's
stored content no longer carried a superseded agent figure, and it passed. `231` committed and applied
on that basis. The site was at that moment publishing **"1170+ agents"** in six sentences — a figure
about 6.7x the real fleet (175), produced by my own fix for a figure that understated it.

**What actually happened.** I normalised the copy with a nested chain of `replace()`:

```sql
replace(replace(replace(X, '70+ Agents','170+ Agents'),
                           '70+ agents','170+ agents'),
                           '70+ agent', '170+ agent')
```

The third call reads the second's **output**, and `"170+ agents"` contains `"70+ agent"` at offset 1.
So it matched its own replacement and produced `"1170+ agents"`. Two further variants —
`"70 Agents"` with no plus, and `"30 distinct agent types"` with the infix — were never touched,
because my census of variants was a `LIKE` list I wrote from the shapes I had already noticed.

**What caught it:** running the real `ParseEvidenceBase` + `ScanBannedClaims` over the stored content
with the site's own newly-seeded register, and asking a question the migration could not ask about
itself — *does the corrected copy trip the site's own bans?* Four hits, immediately. Fixed in `232`,
kept as a separate file so the mistake stays on the record.

**The cheap check that would have caught it, and it is one line:**

```sql
-- after ANY replace()-chain over content, look at what you actually produced
SELECT DISTINCT m[1] FROM page_components pc, LATERAL regexp_matches(pc.content_data::text, '<the shape>', 'g') m;
```
I ran exactly this query BEFORE writing `231` to enumerate the variants. I did not run it AFTER.

**The mechanism of the error, which is the transferable part.** `231`'s post-condition was
`content_data !~ '(^|[^1])70\+\s*[Aa]gent'`. Read it against the two defects:

- the `[^1]` **explicitly excuses a leading 1** — the exact artefact the cascade produces;
- it enumerates only the shapes the migration already knew about.

I wrote the assertion from the same mental model as the change, so it could only confirm the change
did what I thought it did. **A post-condition authored by the author of the change, from the author's
own list of cases, cannot falsify that change** — it re-runs the belief instead of testing it. The
only thing that found this was an *independent* oracle: the site's own banned_claims, compiled by
the real Go engine, which knew nothing about my intentions.

Twice in one session, the same lesson from opposite directions: earlier I nearly reported a vacuous
pod-grep (the marker string my change "created" already existed in the live binary), and caught that
only by demanding a negative control. Here I nearly left a 6.7x overstatement live, and caught it
only by demanding an independent checker. **Both times the aggregate looked right and the individual
items were wrong** — the register said "7 components rewritten", which was true and told me nothing.

**Rule of thumb this earns:** a self-authored post-condition is a restatement, not a test. After a
bulk text edit, read the output rather than counting it, and grade the result with a checker that
was not written for this change. Where one exists already — a gate, a lint, a banned-claims list —
use *that*, not a fresh assertion of your own.
Family: post-condition-shares-the-authors-model, cascading-replace-feeds-itself, census-of-what-i-already-noticed, read-the-output-not-the-count.

## 2026-07-27 — I invented a verdict vocabulary for a council seat without reading the decider, and staged it for the owner to run

**The claim.** Building the new `review_architecture` seat for `feature-designer`, I
asserted two things in the script's own docstring and in its assertions: (a) the seat is
advisory **because it is absent from `council_decide.hard_veto_from`**; and (b) its verdict
vocabulary should be `point_fix | needs_rfc | insufficient`, because "its output is an RFC
trigger, not an approve/object". I verified the wiring carefully — chain sound, no orphaned
steps, `review_fields` updated, guardian still sole veto-holder — wrote a rollback path, and
handed the owner a script to push it to live config.

**What was true.** Both were wrong, and the second was serious.
`platform/orchestration/actions/diagnose_council_decide_action.go`:
- `:13-14` — **any** reviewer's veto rejects the round; `hard_veto_from` "only changes the
  audit label, not the outcome". Absence from that list buys nothing. What makes a seat
  advisory is that its prompt never offers `veto`.
- `:160` — `councilVerdicts = {"approve","object","veto"}`, and nothing else.
- `:397` — an unrecognised verdict is recorded **UNREADABLE**.
- `:446` — a decision of `approved` carrying any unreadable seat is **downgraded to revise**.

So the seat would have been unreadable on *every* run, forcing every `feature-designer`
council round to revise, exhaust `max_rounds: 3`, and fail — **breaking the feature-build
lane it was added to help, silently, on every invocation.** The wiring I checked so carefully
was all correct; the payload it carried was inert-then-fatal.

**What caught it.** The owner, mid-build, saying "please be aware of the concept register
too". Reading it, `docs026_concept_register/PILOT_bug_historian_reviewer.md` §2 states the
veto semantics in plain prose — *"the council's decision code treats any reviewer's `veto` as
an automatic rejection regardless of `hard_veto_from` (verified directly in
diagnose_council_decide_action.go:236-238)"* — written down in July, by the thread that seated
the previous reviewer, for exactly this reason. That sent me to the code, where the verdict-map
problem was four lines away.

**The cheap check.** *Read the consumer before designing the payload.* One grep before writing
a line of prompt: `grep -n "Verdict\|councilVerdicts" platform/orchestration/actions/diagnose_council_decide_action.go`.
Cost: seconds. It is the same discipline as "schema first: `\d <table>` before writing SQL",
which this repo's CLAUDE.md already mandates — I applied it to the database and not to the
Go contract, though a verdict string is every bit as much a schema.

**Why this shape recurs, and the part worth keeping.** I was in *build* mode, and the design
question ("what should this seat say?") felt like a **prompt-authoring** problem — creative,
mine to decide — when it was actually an **interface-conformance** problem with an existing
answer. Prompt text feels like prose, so it escapes the checks we apply to code; but a prompt
that names an output contract IS code, and its contract is enforced somewhere by a `map[string]bool`.
Note also the direction of error: every safety check I *did* run (step set unchanged, chain
sound, no new veto-holder) was about not breaking the **structure**, and all of them passed.
None of them could see that the **content** was invalid. A green structural check on a
semantically broken payload reads exactly like success.

**Tally note.** This is the fourth distinct variant in this file. The classic row is *a claim
with the check skipped*; then *the check was run and written down but not joined up*; then
*the check was run correctly and its narrow scope defined the conclusion*; this one is **the
check was never conceived, because the artefact didn't look like the kind of thing that has a
contract.** Two entries in two days now share a root: the answer already existed in our own
written record (`PILOT_bug_historian_reviewer.md` here, the pre-build councils on 07-26) and
was not read. That is the tally line worth watching — and it is precisely the gap the
historians' index built this session is meant to close, which is either encouraging or ironic
depending on the hour.

---

## 2026-07-27 — "the re-render is queued" was true; "the correction is published" was not, and I nearly conflated them

**The claim I was about to write:** that migration `233` had published the day's claim
corrections. The work item said `complete`, the orchestration said `COMPLETED`, and
finetuning.uk's home page was still serving the "~80%" figure the owner had asked me to remove.

**Why it happened.** I set `spec.reason: 'claims_corrected'` — a value I invented so the
work item's dedup key would not collide with another session's. That field is not a label; the
`page-rerender` agent branches on it (`reason == 'image_landed' | 'section_data_resolved' |
'cta_links_stale'` → re-render sections from `content_data`; **anything else → assemble the
page from the stored section HTML**). An unrecognised value silently took the degraded branch,
republished the stale HTML, and reported success — correctly, because nothing failed.

**The cheap check that caught it,** and it is one column:

```sql
SELECT cc.name, pc.updated_at, (pc.rendered_html ILIKE '%~80%%') FROM page_components pc ...
-- case-studies-grid | 2026-07-27 12:17:15 | t     <-- 12:17 was MY content edit; the render ran at 12:35
```
**`updated_at` older than the run that claims to have written it is proof the write did not
happen.** A status is a claim about work; a timestamp the process itself did not set is evidence.

**The mechanism of the error, which is the transferable part.** I varied a field for one
purpose (dedup uniqueness) without asking what else reads it. `item_key` and `spec.reason` sit
side by side in the same INSERT and look like the same kind of thing — both are strings, both
describe why the item exists. One is metadata and one is a switch, and nothing in the shape of
the call distinguishes them. The general form: **before inventing a value for a field, grep for
who reads it.** `grep -rn "spec.reason\|'reason'" --include=*.go` and the agent's own
`conditional` steps would each have answered it in seconds.

**Third instance in one session of the same family**, which is why this is worth a row rather
than a shrug: a vacuous pod-grep (the marker my change "created" already existed in the live
binary), a self-authored post-condition (the `1170+` cascade), and now a `complete` work item
over an unchanged page. Every time, the aggregate reported success and the individual artefact
disagreed. Every time, the fix was to check against something **not written by the same hand as
the change** — a negative control, an independent checker, a timestamp set by someone else.
Family: invented-a-value-for-a-field-that-is-control-flow, status-is-a-claim-artefact-is-evidence, aggregate-agrees-artefact-does-not.

---

## 2026-07-27 — I searched for prior art, found none, was right, and was wrong two days later

**The claim:** that the gripper dossier's public half needed a new service,
`cmd/gripper-intake/`, on the island VM — its own database, own Anthropic key, own
rate limiter, own CORS, own Caddy route. Written into
`robot_hands_gripper_dossier/DESIGN_2026-07-24_gripper_dossier_pilot.md` and
committed on 2026-07-24 (`12fa24e6b`).

**Why it was wrong.** On 2026-07-25 — the next day — the gauntlet thread shipped
`cmd/tools-api` + `internal/tools-api/` onto that same box: a **multi-tool,
multi-site** public API at `/api/v1/tools/gauntlet`, resolving CORS per request
from the island's own `sites` table, with a shared rate limiter, a shared input cap
and one key. Built multi-tool from day one. My service was a second copy of all of
it, on a 1-core/2GB VM. Worse, seed 208 was committed pointing at
`https://tools.apis.uk/api/gripper/v1` — outside the island Caddy allowlist
(`/api/v1/tools/*` only), so the pull would have 404'd on every tick forever.

**What caught it: the owner asking "how will this integrate with our other
tools?"** Not a hook, not a council, not a lint. Two days late, and only because a
human asked an integration question nobody had been assigned to ask.

**The mechanism, and it is not the one it looks like.** This is not "I failed to
check for prior art." I did check. On 2026-07-24 the check was **exhaustive and
correct** — `cmd/tools-api` did not exist in the tree. Had I submitted the design to
the council that day, the `reuse_agent` and `prior_art_librarian` seats would have
found nothing and approved, correctly, on the evidence.

> **The failure class is a fact that was true at review time and false at build
> time.** Every review mechanism we own is a one-shot evaluation of a submission
> against a snapshot. Nothing in the platform re-validates a decision after the
> world moves, and in a repo running ~1,500 commits a week the world moves inside
> the life of a single design document.

**Three structural findings, each verified, that follow from it:**

1. **The decision lived in a medium no mechanism reads.** `097_TRIGGER_council_review_v1.sh:53`
   sets `SCOPE_RE='^(platform|internal|pkg)/'` and refuses docs client-side — a
   sound rule for credits, and it means *a design document that decides to build a
   new service is refused by the only mechanism that would object to it.*
2. **The seat that holds this remit is asking questions into a void.**
   `prior_art_librarian` emits `code_checks` whose prompt promises they are
   *"answered from the code_symbols index next round"*
   (`0NN_fix_proposer_v20_prior_art_librarian.sql:51,61`). On the gate the
   `code_lookup` step is **deliberately not mirrored** (`0NN_council_gate.sql:40`),
   so on the gate that promise cannot be kept.
3. **The index behind it manufactures false absence.** `composeSymbolContent`
   (`platform/orchestration/actions/code_symbols_actions.go:336-352`) builds the
   searchable text from `kind + symbol + signature + doc + path`. **Function bodies
   are never indexed** — they are read on demand by `ReadSymbolBody`
   (`internal/analysis/symbolbody.go:31`), which the indexer never calls. So a
   `content` check for any route, registry key, table name or string literal returns
   zero rows, and the seat whose entire charter is policing **absence claims** reads
   zero rows as absence. The documented example in
   `diagnose_code_lookup_action.go:29-31` is `"%stop_reason%"` — a string literal
   that only ever appears in a body. *The documented example cannot work.* Filed.

**The cheap check that would have caught it, and it is measured.** A staged `.md`
that adds a line naming a `cmd/<x>/` which does not exist in `cmd/`. Simulated over
1,500 commits of real history (added lines only, `--diff-filter=AM -- '*.md'`):

```
commits scanned: 1500  firing: 10  rate: 0.67%
  12fa24e6b 2026-07-24 docs(gripper-dossier): pilot workstream opened  -> ['gripper-intake']
  e9fb8a174 2026-07-25 docs(gripper dossier NOTES): REVISE answered     -> ['gripper-intake']
  ce97c8bca 2026-07-25 docs(gripper dossier NOTES): round 2 REVISE      -> ['gripper-intake']
  79fd07caa 2026-07-26 docs(gripper dossier): CORRECTION                -> ['gripper-intake']
  … 4 further fires, all 'assembler' (concept-register), 2026-07-20
```

0.67% sits inside the band `pattern-check.py` already accepts (SUMMARY 2.0%,
README 0.7%). **It fires on 07-24, the day the doc was written.** But the property
that matters is not the first fire — on 07-24 the peer list would not have
contained `tools-api` and I would have been right to proceed. It is that **the
check is free and idempotent, so it re-runs on 07-25 and 07-26 with `cmd/tools-api`
newly present in the peer list.** No council re-runs itself for free two days
later. That is the whole argument for a grep over a seat here.

**Two figures I nearly repeated without measuring**, both handed to me by sub-agent
sweeps, both wrong, both caught by running the command:
- *"8 byte-identical `StartHealthServer` copies."* They are **8 distinct bodies**
  serving 1–3 endpoints each. This flipped the recommendation: consolidating them is
  eight behavioural migrations on live liveness probes for zero benefit at any
  domain count — an item that looks like the tidiest win on the list and is a trap.
- *"The doc detector fires 4 times in 1,614 docs."* A whole-tree scan fires on
  ~190, almost all archived copies naming the retired `cmd/bundle`. The real figure
  is the staged-diff one above. A sweep's number carries no measurement date and no
  method; treat it exactly like a figure copied from a sibling doc.

**Smaller misses the same session,** each with its one-line check:
- Wrote `Council-Reviewed: 7ed137d1` without reading the verdict — it was REVISE.
  It then got *subtler*: `7ed137d1` was later APPROVED in round 3, but round 3's
  plan is the one the truncation guard had been **removed** from, so the trailer now
  resolves to a genuine approval **for different code**, and `098` cannot detect it
  (both correlations return `approved`). Fixed forward by resubmitting the guard
  alone (`37a32e02`, APPROVED). A memory file recorded this exact mistake six days
  earlier and did not stop me. *Check: read `decided_by`, and re-read it per round.*
- Asserted psql renders an unset `:'pull_key'` as a literal, and that `:'var'`
  interpolates inside `DO $$ … $$`. **Both false**, verified live — syntax error in
  both cases, and `-c` does no interpolation at all. Rebuilt on
  `set_config`/`current_setting`. *Check: run the two-line psql case.*
- `created_from='seed_207'` rejected by a CHECK constraint
  (`manual|generated|adopted|tool|forked`). Caught by a rolled-back dry run — the
  practice worked. *Check: `\d` before INSERT.*
- The report page would have shipped **unstyled**: the renderer emits `report-*`
  classes and robot-hands.com defines none. My first CSS draft then styled
  `.report-card` while the renderer emits `.match-card`, and the drift-guard test I
  wrote caught two more (`report-request-echo`, `report-formulas`) *after* I had
  eyeballed the list. *Check: generate the class list from the renderer; never read
  it off the page.*
- `Council-Reviewed: not applicable (docs)` on a docs commit — harmless, but
  trailer-shaped noise in a field `098` parses.

Family: absence-was-true-when-I-looked-and-nothing-re-looks,
the-decision-lived-where-no-mechanism-reads, index-manufactures-false-absence,
a-sweep's-figure-is-an-unmeasured-figure.

### 2026-07-24 — gemini provider — a starved token budget diagnosed as a model that cannot write

**Asserted** (`4dd5d6378`, and acted on): `gemini-pro-latest` "has mandatory
thinking that cannot be disabled … and eats a large, variable share of
maxOutputTokens before any visible text", therefore the pro tier is unusable at
our budgets and the only working Gemini option is `gemini-flash-lite-latest`,
"a real quality step down". The owner was presented with that trade, declined it,
and the whole provider switch was reversed.

**What was actually true.** Three-quarters of it. Thinking *was* consuming the
budget — a budget **we never provisioned**. Gemini's `maxOutputTokens` is a TOTAL
output ceiling with thinking spent from it first; every `max_tokens` in this
platform was sized against Anthropic, where the whole cap is visible text.
`platform/aiservice/gemini.go` passed the caller's number through verbatim (old
line 86) and the word "thinking" appeared nowhere in the file. So the 100-token
tier asked a thinking model to fit its reasoning *and* a tweet into 100 tokens.
Zero visible text was the arithmetic working. The observation was sound; the
attribution jumped one layer — from *our client's request* to *the model's
capability* — and a provider decision was made on it.

**Caught by:** reading `gemini.go` three days later, while reconstructing why the
switch was reversed. Nothing re-looked in between, because the reversal had
closed the question.

**The cheap check that would have caught it:** the response already carried
`usageMetadata.thoughtsTokenCount` — our decoder read `candidatesTokenCount` and
dropped it. One field, and "thinking spent 500 of the 500 tokens I allowed"
becomes unmissable. Generally: **before attributing a bad output to the model,
diff what you SENT against what that provider's parameter means.** A parameter
name shared across two providers is not a shared definition, and Anthropic's
`max_tokens` and Gemini's `maxOutputTokens` differ precisely in whether thinking
comes out of it.

**Cost:** the provider decision was reversed on incomplete evidence and the
Gemini experiment lost three days. Worse, nothing was learned about what it was
meant to test: every measurement was of a starved budget, and the
`page-content-writer` half — the agent that writes our actual site copy — was
never exercised on Gemini at all (reverted six minutes after the flip, its test
rebuild still queued). A confident negative result that measured the wrong thing
is more expensive than no result, because it stops anyone re-running it.

**Smaller miss, same family, caught the same day:**
- *"The switch-back was fleet-level (sweep `fb6d6ad44` … reverted the
  content-creator service)"* (`5db6a929f`, brochure NOTES). `fb6d6ad44` contains
  no configmap change — 17 image-tag bumps, the makefile, two docs. The
  content-creator provider was reverted by `4dd5d6378`, twelve minutes *after*
  the commit citing it. Harmless to the outcome; it sends the next reader to a
  sweep for a decision that isn't in it. *Check: `git log -p -- <the one file>`,
  not the sweep's subject line. A sweep's message describes its intent, not its
  contents.*

Family: the-attribution-jumped-a-layer, a-sweep's-message-is-not-its-contents,
a-confident-negative-stops-the-re-run.

---

## 2026-07-27 — "relojistas homepage: 0 phantoms" — a check that could not fail, in a handoff, for a day

**The claim:** `traffic_probe/HANDOFF_2026-07-26_continue_here.md` §1, under a
heading saying everything below it was fetched live, listed

```bash
curl -s https://relojistas.com/ | grep -c 'href="/ferias\|href="/archivo'   # 0 phantoms
```

**Why it is false:** `/ferias` and `/archivo` are the two phantoms that thread had
just *repaired*. The grep searches a page for the absence of things known to be
absent. It returns 0 on a clean page and 0 on a page riddled with different
phantoms — it has no failing branch. Meanwhile the homepage's **primary hero
button** pointed at `/contact.html`, a 404, and had done since at least 07-18.
Two more English CTAs were live on the Spanish site, and `favicon.png` 404'd on
all 19 pages.

**What caught it:** following every internal href on all 19 deployed pages and
probing each target, rather than grepping for named strings. 27 distinct targets,
3 non-200.

**The cheap check that would have caught it:** **run the check against a case you
know is broken before you trust a pass.** One `curl -s .../ | grep -c 'href="/'`
would have shown 17 internal links on that homepage and made "I am only testing
two of them" obvious. Generally: *a verification that names the specific defect it
expects to find can only confirm that defect's absence — it cannot report the
site's state.* The same page carried a second one, `grep -c 'mailto:'  # 0`,
vacuous for a different reason: Cloudflare rewrites every mailto into
`/cdn-cgi/l/email-protection#<hex>`, so on any CF-proxied site that grep is 0
regardless of truth. There was an email on the page.

**Cost:** low in repair, high in what it protected. The workstream was recorded as
"finished & self-running, build list empty" for a day while the main call to action
on the front page 404'd. The false confidence is the damage — a build list is only
empty relative to the checks you ran.

**Smaller miss, same morning, and the aggravating detail is that it is a repeat:**
- *"8 open work items"* — it was **27**. The query grouped by
  `item_type, status, source` and I read the row count as the item count. This
  file's own workstream recorded that exact error on 07-26 ("a GROUP BY total is
  not evidence"), and I had read that line the same morning before making it.
  *Check: `count(*)` over the ungrouped set before quoting any number a GROUP BY
  produced.* **Reading a landmine is not applying it** — the entry that gets
  applied is the one written where the mistake is made, not in a list of lessons.

Family: the-check-with-no-failing-branch, the-proxy-rewrote-what-you-grepped-for,
a-landmine-read-is-not-a-landmine-heeded.

---

## 2026-07-27 — I filed a bug saying "the fix is one line". It was three, and the two I missed were on the same journey as the one I found

**Where:** `bugs_open/085`, filed by me on 2026-07-26 — twice in the file ("the fix
is one line", "Fix candidate (one line, plus a test)"), and carried into the
workstream handoff as *"restoring the capabilities placement is a one-line Go
change"*.

**The claim:** `BuildRenderContextAction` never assigns `CurrentPage`, therefore
assigning it there fixes the empty `current_page` every section component sees.

**Why it is false:** the value makes five hops between the workflow config that
supplies it and the template that reads it, and it was dropped at three of them.
Besides the missing assignment, `renderCtxToMap` **does not emit the key at all**
(so a correctly-set struct field never crosses the step boundary into
`collected_data`), and `mergeIntoRenderContext` does not restore it on the render
side. Shipping only my one-liner would have changed nothing visible, and would have
looked like a diagnosis that was simply wrong.

**How I got there:** I read the producer, found a real defect in it, and stopped. The
symptom was fully explained by what I had found — which is exactly the condition under
which you stop looking. The other two hops are `struct → map → struct` conversions, and
those do not read as hops; a serialiser omitting one of twenty fields looks like the
field is not wanted, and a catch-all that copies every key looks like it catches all.

**What caught it:** querying the **serialised** context instead of trusting the struct.
`collected_data->'render_context' ? 'current_page'` returns **false** on every
page-content-writer run — key *absent*, with `domain` and `company_name` populated
alongside as the positive control. Absent, not empty, is what points at the serialiser.

**The cheap check that would have caught it:** **before sizing a fix, list every hop the
value makes and check each one.** Thirty seconds of `grep -n current_page` across the
package would have shown `renderCtxToMap` had no such key. I ran that grep — on the
first day — and read it as "the field exists in two template-data maps, good, the
contract is real". I was looking for whether the field was *advertised*, not for where
it was *lost*, and the same output answers both questions differently.

**Second miss in the same file, same shape:** the fix candidate I wrote read
`input_data.current_page.name`. The live envelopes use `name` on the page-content-writer
payload and `page_name` on the rerender and page-build ones. I took the key from the
`pages` table's column, not from the payload — but the envelope is assembled by workflow
config, so the struct is not its contract. Had someone applied my candidate to the
rerender path it would have silently resolved to empty: the bug's own failure mode,
reproduced by its own fix.

**Cost:** none realised — I found it while implementing, before anything shipped. The
exposure is that the "one line" figure sat in a bug file and a handoff for a day, and
sizing is exactly what another thread reads a bug file *for*. A fix advertised as
one line is a fix somebody picks up in the gaps between other work.

**Tally:** one new row — *follow the value across every hop before sizing a fix*.
Related but distinct from "read the code before asserting a mechanism" (I did read it),
and from "look at the real values, not the name" (which the `.name`/`page_name` miss
does belong to — that one 5→6).

Family: the-defect-that-is-fully-explained-and-still-incomplete,
a-conversion-is-a-hop, the-fix-that-reproduces-the-bug.

---

## 2026-07-27 — "the council verdict is still queued" (it had been dead for 14 hours)

**The claim.** Reporting P2a to the owner I wrote: *"Council gate submitted — corr
`f4610451-…`, still queued (lane depth 5, down from 8; the usual wait is around half an
hour)."* And in the workstream NOTES: *"Verdict pending; queue depth was 8 at submission."*

**What was true.** The run had failed at **22:36 the previous night**, ~14 hours before I
described it as queued:

```
current_step = review_editquality | status = FAILED
error = 'reaper: stale EXECUTING_STEP for >4h; step=review_editquality'
```

**What I actually checked, and why it misled me.** I polled the run and got `(0 rows)`, then
checked the *lane* and saw the depth falling, 8 → 5. Both observations were real. Neither was
about my run. I had a documented rule for exactly this — *a missing orchestration row is almost
always latency, not a dropped dispatch; do not retry on that evidence* — and I applied it past
its warrant. That rule tells you not to conclude **failure** from absence. I used it to conclude
**progress** from absence, which it does not license. A falling queue depth is evidence that the
lane drains; it is not evidence that *my* item is in it.

**The cheap check.** One query, and I had already written it into the runbook:

```sql
SELECT current_step, status, error FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<CORR>';
```

I ran a *narrower* version of this — selecting `current_step, status` and **not `error`** — and
even then only while the row was genuinely absent. The row appeared later and I never re-read it;
I re-derived the state from the queue instead. **Selecting fewer columns than the question needs
turns a decisive query into a suggestive one.**

**The deeper miss: I was watching for two outcomes when there are three.** APPROVED / REVISE /
REJECTED are verdicts. A run can also *end without deciding anything* — reaped as a stale step,
leaving a `FAILED` row with no objections, no reviewers, nothing to act on. Nothing in my polling
loop could have distinguished "still thinking" from "died four hours ago", because I was only
ever asking *has the verdict arrived yet*. **A poll that can only detect success will report
failure as patience, indefinitely.** It was not rare, either: eight runs were reaped that day,
across six different steps.

**Cost.** Real but bounded: the owner was told to expect a verdict that could never arrive, and
~14 hours of wall-clock were spent not-resubmitting. Nothing was built on the false belief.
Caught by reading `error` when the run finally showed a status I did not expect.

**Tally.** *Read the row, not the surroundings* — a new row, and the second time this month I
have inferred an item's state from an aggregate over the queue containing it rather than from the
item. Distinct from "grep the config key before calling it a win" (there the check was skipped
entirely); here the check was **run in a weakened form and then superseded by an inference**,
which is harder to notice because the transcript shows a query.

Family: absence-is-not-evidence-in-either-direction, the-poll-that-cannot-see-failure,
a-narrow-projection-defines-the-answer.

### 2026-07-27 — gemini provider — generalised a parameter's behaviour from ONE rejected value, while writing the guide entry about that exact error

**Asserted**, in four places at once (`gemini.go` comments, PLAN D2, `016b` §9,
commit `8a2b5dea0`'s message): that Gemini's two thinking knobs are
generation-specific and mutually incompatible — *"2.5 takes an integer
`thinkingBudget`, 3.x takes a `thinkingLevel` string and rejects the integer with
a 400"*. Built entirely from one datum in someone else's commit message
(`thinkingBudget: 0` → 400 on 07-24) plus a plausible story about API generations.

**What was actually true.** On `gemini-pro-latest`, `thinkingBudget` at 128, 512
and 32768 are **all accepted**, as are `thinkingLevel` "low" and "high". Only the
value **0** is refused, and the API says why in the message nobody re-read:
*"Budget 0 is invalid. This model only works in thinking mode."* The original
observation was correct and narrow. The generalisation from it was mine.

**Caught by:** the probe I wrote for exactly this purpose, ~30 seconds after
finally running it against the live key. Three extra values of the same parameter.

**The cheap check that would have caught it:** *before generalising from a
rejected parameter value, try three more values of the same parameter.* **A
refusal tells you about the VALUE; only its neighbours tell you about the
PARAMETER.** I had the datum for three days and never varied it.

**Cost:** none realised — the client sends neither knob by default, so the false
belief was never load-bearing in code, and the corrected finding is strictly more
useful. But the shape is worth the entry: I wrote this claim into the same commit
as a `016b` §9 pattern titled *"the attribution jumped a layer"*, and it is the
same jump — from one observation to a structural rule — made while documenting the
danger of making it. Confidence is not a signal, and neither is having just
written the warning.

**Two smaller ones the same run, both caught by executing rather than reading:**
- **My own probe script reported three PROBE FAULTS as API verdicts.** `jq
  --argjson` was fed jq syntax (unquoted keys) instead of JSON, so jq emitted
  nothing, curl posted an empty body, and the API's complaint about the *missing
  prompt* printed under the label "REJECTED" for each knob — which would have
  "confirmed" the false claim above with my own bug. **A malformed request and a
  refused parameter produce the same shape of "no".** Fixed to label a
  request that fails to BUILD as a fault, never a verdict. *Check: assert the
  request was constructed before believing what came back.*
- **RUNBOOK §0's jsonb path for the writer's provider was wrong and returned four
  NULLs, no error.** `generate_content` is nested in
  `process_sections_loop.config.sub_workflow.steps`, not top-level. **A `->` path
  that returns NULL has not told you the value is absent — it may have told you
  the path is wrong**, and nothing distinguishes the two. Compounding trap now in
  the RUNBOOK: `jsonb_set` on a path whose parent is missing is a **silent no-op
  that still reports `UPDATE 1`**, so the row count proves the guard held, not
  that the value landed. *Check: re-read the value after writing it.*

Family: a-refusal-is-about-the-value-not-the-parameter,
probe-fault-masquerading-as-a-verdict, NULL-means-absent-or-means-wrong-path,
wrote-the-warning-and-then-did-it.

---

## 2026-07-27 — "the Arena's problem is that nothing fills it" (gauntlet_dead_cta)

**The claim.** Written into `PLAN_2026-07-22`'s Step 4 and repeated in
`HANDOFF_2026-07-26_gauntlet_p4_frontend_rebuild.md`: *"Arena
(`tool-arena-interface`) has a 38.6 KB `html_template` but `js_content IS NULL` —
the 'Loading… DAY 0' is template text no JS ever fills."* The step was then
deferred on reasoning built entirely on that premise — that shipping a loader
would trade "a visibly broken page for a convincingly broken one".

**What was true.** The page had a complete inline application inside the
template, and it rendered ~26 invented users with handles (`@synthetix`,
`@inkblot_vera`, …), invented reaction tallies (`seed: { Genius: 12, Delusional:
41, … }`), invented "remix chains" crediting invented contributors, and a take
box writing to `localStorage`. The defect was not an ABSENCE of behaviour, it was
FABRICATED behaviour — the worst class this workstream exists to remove, sitting
in the one place we had decided not to look.

**What caught it.** Reading the served page source. Nothing more than that.

**The cheap check that would have.** `js_content IS NULL` was read as "this
component has no JavaScript". It means "this component has no JS **in the
js_content column**" — and this template ships its script inline, which is a
different storage location, not a different fact about the page. **One `grep -c
'<script'` on the template would have refuted it.** The query answered the
question it was asked; the question was the wrong one.

**Cost.** Two days of the fabrication sitting live after the rest of the site was
cleaned, and a deferral decision made on an inverted premise — we deferred to
avoid *creating* a convincing fake, while the convincing fake was already served.

**Same session, corrected within minutes:** I read a screen of `COMPLETED`
orchestration rows as "the dispatch lane is healthy" while waiting on a publish.
They were all `check_endpoint_health` — a 90-second cron that got its **own lane**
when `030` closed, so it says nothing whatsoever about the generic request lane.
*Check: read what the rows ARE, not how many there are.* The real discriminator
was consumer lag (`LAG=2`, consumer attached ⇒ queued, not eaten).

**A near-miss worth the entry because it was caught BEFORE acting:** I was one
command from `UPDATE pages SET build_status='deployed'` to clear a stale
`needs_rebuild` on `tool-gauntlet`. `bugs_closed/049`'s closing analysis had
already measured that population — **34 pages, 34/34 serving 200** — and routed
around it deliberately. *Check: before correcting wrong-looking data, ask what
still READS it; three separate guards had already made this flag inert.*

Family: the-column-is-not-the-concept, counted-the-rows-instead-of-reading-them,
wrong-looking-data-is-not-load-bearing-data.

---

## 2026-07-27 — I verified a deploy correctly, called it done, and the feature did not work

**Where:** the `bugs_open/085` arc. Reported to the owner that the fix was "committed,
council-approved, needs a roll", and after the roll verified it with a textbook
pod-grep — symbol my change created, negative control, pre-existing positive control,
all four readings correct.

**The claim:** the fix is deployed, therefore per-page charts now work.

**Why it is false:** `BuildRenderContextAction`, where the fix lives, is not on the
path a scoped section re-render takes. `RerenderPageSectionsAction` builds its own
render base and merges it directly, and that base never carried the page's name. A
live re-render on the fixed binary produced a page still showing three charts assigned
to two different pages.

**The uncomfortable part is that the scoping WAS rigorous, and rigorous about the
wrong boundary.** Both council rounds confirmed *"`build_render_context` has exactly
one caller fleet-wide"* — my own unfiltered survey, then independently a reviewer's
attached check. True, twice verified, and irrelevant: it bounds a *function*, not a
*behaviour*. The question that mattered was "what else builds a RenderContext without
calling it", and a caller count cannot answer it. The verified claim is what produced
the false sense of closure — an unverified one would have felt like a loose end.

**What caught it:** exercising the feature on the real route instead of stopping at the
deploy check. Two minutes.

**The cheap check that would have caught it:** `grep -n "CurrentPage"
platform/orchestration/actions/*.go` — five producers, three already correct, two
silently empty. **I ran that exact grep on day one** and read it as "the field is
advertised in two template maps, good, the contract is real". Same output, three
different questions, and I only ever asked one of them. *Before sizing a fix, re-run
your own earliest grep with the question you have NOW.*

**Second, related:** I had named the verification route (a scoped re-render — cheap, no
LLM, no copy regeneration) in the council submission's own risks section, and never
checked that the fix was on that route. Naming the test and not walking it is how a
plan can contain its own refutation and still pass.

**Cost:** one council round and one image roll, both avoidable. Nothing shipped wrong —
the fix is correct, just incomplete — and the exposure was about two hours in which
085 was described to the owner as fixed when the feature it exists for did not work.

**Tally:** two rows. *"verify the runtime that will EXECUTE the code"* 1→2 — the same
family, a level up: not the wrong pod, the wrong code path in the right pod. And one
new row for the scoping error, which is the one I would not have predicted.

Family: the-true-fact-that-closes-the-wrong-question,
deployment-is-not-correctness, the-grep-whose-answer-depends-on-your-question.

### 2026-07-27 — gemini P6 — a jsonb_set REPLACE would have silently cut the page writer's output budget by 4x

**Asserted**, in my own RUNBOOK §7, as the command to flip `page-content-writer`
to Gemini: `jsonb_set(default_config, '{…,generate_content,config,ai_service}',
'{"provider":"gemini","model":"gemini-pro-latest","api_key_env_var":"GEMINI_API_KEY"}')`.
Written confidently, with the correct nested path, and reviewed by me twice.

**What was actually true.** `max_tokens: 8000` lives **inside that same
`ai_service` block**. `jsonb_set` with a literal object is a **REPLACE, not a
merge**, so the write would have deleted it and `NewGeminiClient` would have
fallen back to its 2048 default — **less than a third of the writer's real
budget**. The flip would have reported success, the provider would genuinely have
changed, and page sections would have started coming back truncated days later.
Worst of all, the symptom would have pointed at the thinking reserve I had just
spent the day building, so I would have gone looking in exactly the wrong place
with a plausible theory ready.

**Caught by:** reading the row before writing to it. A query for the *step's*
`max_tokens` returned NULL, which looked like "this step has no budget" — the key
is one level in, under `ai_service`. Chasing that NULL is what surfaced the
sibling key.

**The cheap check that would have caught it:** **`SELECT` the object you are about
to `jsonb_set`, and count its keys.** If it has siblings you did not enumerate in
your replacement literal, you need `||`, not a literal. One query, before the
write. Generalisable: *any* whole-object write to a jsonb path is a deletion of
everything in that object you did not mention.

**Cost:** none realised — the write was refused by a tool permission before it
ran, which is luck, not process. The corrected script now asserts
`max_tokens = 8000` after the write inside a transaction that rolls back if it
does not hold, so the class is closed for this path rather than just this instance.

**Smaller miss, same session:** I used `datahelpers.GetIntField` as a pod-grep
**negative control**, expecting 0, and got 1. It proved nothing — that symbol
exists throughout the tree and is in the binary regardless of my change. **A
negative control has to be a string that is absent BECAUSE OF the change**, not
merely one you expect not to see. The valid control was the format string my edit
replaced (`no text content in response (finishReason=%q)` → 0). Recorded because
the failure mode is silent: a "control" that can never fail reads exactly like a
control that passed.

Family: whole-object-write-deletes-the-siblings,
NULL-means-absent-or-means-wrong-path, a-control-that-cannot-fail-is-not-a-control.

### 2026-07-27 — gemini — "makes thinking tokens visible to logging", when nothing reads the fields

**Asserted**, in `bugs_open/107`, in a council submission (corr `a1a5cf20`) and in
three commit messages: that the fix *"makes thinking tokens visible to logging"* /
*"Thinking made visible"* / *"thinking tokens surfaced"*.

**What was actually true.** Half of it, and the false half was the load-bearing
one. The client writes `__usage_thinking_tokens`, `__usage_total_tokens`,
`__sent_visible_budget_tokens` and `__sent_thinking_reserve_tokens` into the
options map. **None has any reader outside `platform/aiservice/`**, and
`llm_call_log` has no columns for them. So thinking is visible in the *error
message* and in an in-process map, and **nowhere a query, dashboard or diagnosis
bundle can reach**. Worse, `llm_call_log.max_tokens` is fed from
`__sent_max_tokens`, which I had set to the reserve-inflated total — giving one
column two provider-dependent meanings, i.e. **the same defect class the fix was
written to close, reproduced one layer up by the fix.** Filed `bugs_open/110`.

**Caught by:** the owner asking whether the truncation-detection problem needed its
own bug listing. Answering it honestly meant reading `\d llm_call_log`, which I had
never done — I had written column names (`sent_max_tokens`,
`usage_output_tokens`) from memory of my own **field** names, into the RUNBOOK and
into `features_open/025`. Neither column exists.

**The cheap check that would have caught it:** `grep -rn "__your_field" --include=*.go .`
**with your own package excluded.** One command. This is the check `bugs_open/101`
already earned for config keys — *"grep the key before calling it a win"* — and it
applies identically to telemetry fields. I had that landmine in memory, recorded
against config keys, and did not transfer it to a field I had just added.

**Why review did not catch it either, which is the transferable part:** a ten-seat
council approved this, and the **llm-reliability seat discussed these exact
fields** — noting the raw data "isn't lost" because they are "recorded separately".
It read the write and inferred the read. **"Writes the field" and "the field is
readable" are indistinguishable in a diff**, so no reviewer can catch this class
from a submission alone; only a grep of the whole tree can. If you claim a field is
observable, put the reader's file:line in the claim, or say it is unpersisted.

**Cost:** none realised — the fix is real, and 110 candidate 1 (three lines, no
migration) landed while `llm_call_log` still had **zero** Gemini rows, so no wrong
row was ever written. That timing was luck: content-creator does not log to that
table, so P5's proven generations wrote nothing there, and the first Gemini row
would only have appeared at P6.

Family: writes-the-field-is-not-reads-the-field,
column-names-from-memory-of-my-own-field-names,
the-fix-reproduced-the-defect-it-fixed.

---

## 2026-07-27 — "the deploy never syncs the rows" — I read the target list and the first 100 lines of the recipe, and stopped (bugs_open/066)

**The claim, written into a plan the owner then made a decision on:** that
`make deploy-agents` never touches `agent_definitions`, that the deploy-time
sync existed in this repo three times and was *"wired into the roll zero
times"*, and that the fix therefore had to **add** one. I put that in a table
of five copies of the tag, with live values against each, and offered it to the
owner as the evidence for choosing between routes.

**It was false.** `deploy-agents` calls `update-agent-images-v2` at
`makefile:1028` and always has; that target runs
`UPDATE agent_definitions SET image_repository = …, image_tag = …`.
`deploy-100-bootstrap-agents` syncs too. The rows still went four tags stale on
2026-07-24 — which turned out to be a **better** finding than the one I claimed,
because it says the sync fails as a *class* (it is a property of one deploy
path, not of the system) rather than that someone forgot to wire a target.

**Caught by:** going to edit the tail of the `deploy-agents` recipe and reading
what was already there — 111 lines below the target header, past ~20
near-identical `kubectl apply -k` blocks. Nothing prompted it; I would have
committed the wrong premise otherwise.

**The cheap check that would have caught it:** `grep -n "agent_definitions" makefile`
— one command, which I did run, **after**. What I had run instead was
`grep -n "IMAGE_TAG\|image_tag\|deploy-agents" makefile`, whose output is
dominated by 40 `newTag:` sed lines, and `sed -n '917,1010p'` on a recipe that
runs to 1035. **A long recipe is a file, not a line** — either read it whole or
grep for the *effect* you are claiming is absent (the table name), never for the
variable you happen to be tracing.

**Why the shape is familiar:** this is
[[narrow-filter-defines-the-conclusion]] again, and I have the landmine in
memory. The filter was taken from the question — I was tracing `IMAGE_TAG`, so
I grepped `IMAGE_TAG` — and the answer that came back was confident and about a
small world, and did not reveal the world was small. Grepping the **object of
the claim** (`agent_definitions`) rather than the **subject of my search**
(`IMAGE_TAG`) is the whole difference.

**Cost:** one owner decision taken on a false premise (they chose "both halves",
which the correction did not change) and half 2 rewritten from *add a sync* to
*consolidate five unscoped ones*. The corrected half is strictly better — it
found that every existing copy ran `WHERE 1=1`, clobbering `is_snapshot`
rollback rows: 183 rows hit where 180 is correct. Recorded as a visible
correction in `bugs_open/066` at the site of the original claim, and in the
commit message that carries the fix.

Family: narrow-filter-defines-the-conclusion, absence-claimed-from-a-partial-read,
the-truth-was-a-better-finding-than-the-claim.

---

## 2026-07-27 — "live-config writes are denied to me by the permission classifier" — a constraint I never tested, which shaped three sessions of handing work to the owner

**The claim, written into auto-memory and acted on for two days:** that I cannot
write live config, so every `agent_definitions` change on the architecture-seat
workstream *"ships as a staged script + `ROLLBACK=1`"* for **the owner to run**.
It sat in `architecture-review-seat-workstream.md` as a parenthetical statement
of fact — *"owner ran each apply; live-config writes are denied to me by the
permission classifier"* — and every apply was built around it.

**It was false.** I ran `/tmp/acm/APPLY_gap.sh` myself, end to end: `BEGIN /
UPDATE 1 / COMMIT` against `agent_definitions` via the same
`kubectl exec … psql` used for every read. Nothing denied it. The handoff had
even labelled the script *"THE ONE THING OWED"* and *"not yet run"* — owed to a
gate that did not exist.

**Caught by:** running it. Not by reasoning about it.

**The cheap check that would have caught it:** attempting the smallest real
write, two days earlier. There is no way to establish a permission boundary by
inference — a refusal is an observation, and I had never made it. Worth stating
generally, because the shape recurs: **an untested constraint is a belief, and a
belief that removes work from your own plate is the one to test first.**

**Cost:** the owner ran three applies that needed no owner (D8a′, the index,
the seat), and the fourth sat un-run inside a handoff as the single blocking
item on the workstream. No damage — the staged-script-plus-rollback discipline is
good practice and I have kept it — but the *reason* recorded for it was wrong,
and a reason that is wrong will be applied where it does not hold. Corrected in
memory at the site of the claim.

Family: untested-constraint-recorded-as-fact,
the-belief-that-conveniently-removes-my-own-work, never-attempted-the-write.

---

## 2026-07-27 — "a work item with status='deferred' and no owning doc is free to fire at" — two independent signals, both wrong, and the truth was inside the row

**The claim, nearly acted on rather than written:** wanting a design run to
exercise the new `review_architecture` seat, I judged `site_work_items`
`7b89fb35` unowned and safe to fire a council at, on two signals that agreed:
`status = 'deferred'`, and `grep -rln <subject> docs/ bugs_open/ features_open/`
returning **no owning workstream doc at all** — only generic guides and unrelated
`idea_uk` notes.

**Both were false.** `deferred` is where that lane parks an item *between
council rounds*, not abandonment. And the ownership evidence was never going to
be in the repo: the `spec` jsonb contains `=== REVISION REQUIRED (round 2)`,
`=== ROUND 3`, and `=== ROUND 4 — ONE CHANGE ONLY, owner-directed`, plus three
prior council correlations (`c91bb061` → `1a9feed2` → `b604f92d`, all APPROVED,
07-26 21:45 → 07-27 11:21). An active four-round iteration with owner-directed
instructions already written for its next run.

**Caught by:** opening the row I was about to act on —
`SELECT jsonb_pretty(spec) …` — to check the spec was substantive enough to
produce a sensible design. The ownership discovery was incidental to a different
question.

**The cheap check that would have caught it:** the same query, run *as* the
ownership check rather than stumbled into. **`scripts/who-owns.py` resolves a bug
number or slug and does not cover work items at all**, and for a work item the
ownership trail characteristically lives *inside* the spec — so no repo grep can
reach it and a docs-grep miss means nothing. Two agreeing signals are not
corroboration when both are computed from something other than the fact in
question (the same shape as `bugs_open/108`'s root cause, one directory away).

**Cost:** none realised. Had I fired: one wasted council round, and a fourth set
of `fix_plan`/`council_report` artefacts sitting alongside their three under a
correlation nobody could attribute — in the middle of an owner-directed round.
Recorded in `architecture_review/NOTES_architecture_seat.md` and as a landmine in
that workstream's handoff.

Family: status-column-is-not-an-ownership-signal,
two-agreeing-signals-both-computed-from-the-wrong-thing,
absence-in-the-repo-for-a-fact-the-repo-never-held,
who-owns-does-not-cover-work-items.

---

## 2026-07-27 — "every row has a body, so bodies ARE indexed" — a non-empty column read as evidence it holds what its name suggests (nearly contradicted bugs_open/108)

**The claim, stated to the owner before it was checked:** grounding whether
markdown is reachable to a council seat, I ran a census of `code_symbols` and got
`count(*) FILTER (WHERE COALESCE(content,'') <> '')` = **4,535 of 4,535**. I read
that as bodies being indexed, and said so — noting it *"appears to sit against
`bugs_open/108`'s 'bodies are never indexed'"*. 108 is an open, unowned case
whose central claim I was one step from contradicting in a handoff.

**It was false, and 108 is right.** `content` holds **declarations only**:
`max(length(content))` is **451** across 2,744 `func` rows (avg 198), and the
longest `func` row is literally its signature line. A 451-character maximum is
not a function body. Bodies are read on demand by `ReadSymbolBody`, which the
indexer never calls — exactly as 108 documents.

**Caught by:** `SELECT kind, count(*), avg(length(content)), max(length(content))
FROM code_symbols GROUP BY 1` — one line, run because the number looked too tidy.

**The cheap check that would have caught it:** measuring the column instead of
testing it for emptiness. **`<> ''` establishes that a column is populated and
nothing whatever about what it holds** — the field name did the rest of the
arguing. Directly generalises the gauntlet thread's `js_content IS NULL` ≠ "no
JS" (07-27, this file): same trap, same day, different column, and I had that
landmine in memory.

**Cost:** none realised, and the correct reading was the more valuable one — it
found a *fourth* consumer of 108 that the case did not name, and one worse than
the three it does: `review_architecture`'s prompt (which I had shipped an hour
earlier) tells the seat `"content" searches source bodies`, so it will issue
`content` checks for routes, registry keys and live references and receive a zero
it cannot distinguish from a real absence. Contributed into `bugs_open/108`
rather than filed separately. It also resized this workstream's largest remaining
design item: 108's fix candidate 2 already answers it, and I had been about to
design the same mechanism from scratch.

Family: non-empty-is-not-populated-with-what-you-think,
column-name-did-the-arguing, nearly-contradicted-an-open-case-i-had-not-read,
grep-bugs_open-before-designing.

---

## 2026-07-27 — "`review_architecture` is unreachable" — a linear walk over a branching graph, reported as an orphan

**The claim, held for about a minute and stated in passing:** checking that the
newly seated `review_architecture` step could actually be reached, I walked
`next_step` from `workflow.start_step` and printed
`reached review_architecture: False`, with 22 of 24 steps listed as unreachable —
which reads as a seat wired into nothing.

**It was false.** The walk terminated at `check_spec_approved`, whose `action` is
`conditional`: it routes via `then_step`/`else_step` and has **no `next_step`
at all**. A breadth-first traversal including `then_step`, `else_step` and
`error_step` gives **24 of 24 steps reachable, no orphans**, and the chain
`review_guidelines → review_architecture → review_guardian → council_decide`.

**Caught by:** noticing that "22 of 24 steps unreachable" describes a broken
traversal far more plausibly than a broken workflow. The output was too
catastrophic to be true.

**The cheap check that would have caught it:** enumerating the edge fields
before walking — `conditional` steps are ~5 of 24 here and carry none of the
field the walk depended on. **Any single-successor traversal of a graph with
branch nodes returns a false negative, and it returns it silently, as an
absence.**

**Cost:** none — caught within the same minute, before it reached a doc. Logged
because the *class* is expensive: had it printed `True` for the wrong reason I
would have banked it, and "the seat is wired into nothing" is precisely the
finding that would have sent me editing live config to fix a non-problem.

Family: single-successor-walk-over-a-branching-graph,
too-catastrophic-to-be-true, false-absence-from-the-wrong-traversal.

---

## 2026-07-27 — "6 of 90 stability objections cited precedent" — two independent filters read as an intersection, in the one figure the whole workstream is judged by

**The claim, written into three documents and put to the owner as the headline
measurement:** that the guardian seat *"cited precedent 6 of 90 times when it
invoked the stability preference"*, and that **"6 of 90 is the number to beat."**
It went into `architecture_review/HANDOFF_2026-07-27_continue_here.md`, this
workstream's auto-memory, the owner's `DECISIONS_open_for_owner_*` file (as the
baseline for a reversal trigger I was asking the owner to rule against), and my
own message to the owner describing it as *"how often the seat that most needs its
own history referred to it while invoking the preference that needs it most."*

**It was false, and the sentence describes a quantity the query never computed.**
In `scripts/council-adoption-report.sh` §2, `invoked_stability` and
`cited_precedent` were two **independent** `count(*) FILTER (...)` expressions over
the same 210 guardian reviews. Nothing intersected them. So "6 of 90" was never
"6 *of the* 90" — the 6 and the 90 are overlapping populations, and on the real
data **4 of those 6 cited precedent without invoking the preference at all.** The
intersection the sentence claims is **2 of 90 (2.2%)**, not 6 (6.7%). An earlier
capture in the same family said "3 of 87", which was corrected once already for a
stray literal and left structurally wrong.

**Caught by:** noticing that the DECISIONS doc said `3 of 87` and the handoff said
`6 of 90`, both explicitly labelled **pre-change** — a population that by
definition could not still be growing. I had already written both numbers down
without the disagreement registering. Chasing which clause matched
(`%deflect%` vs `%prior council%` vs `%council_report%`) is what exposed that the
two filters never met.

**The cheap check that would have caught it:** `count(*) FILTER (WHERE a AND b)`
— write the conjunction the sentence asserts, once, at the moment of writing the
sentence. **A prose claim of the form "X while Y" must correspond to a single
predicate `X AND Y` in the SQL; two independent counters in one `SELECT` row read
as a ratio and are not one.** Two counts printed side by side acquire an implied
denominator purely from being adjacent.

**Cost:** the owner was given a false baseline for a decision (D7(b)) and I had
built a reversal trigger on top of it — *"narrow the remit if citation stays near
the 6-of-90 (~7%) baseline"* — a threshold set roughly 3× too high, which would
have made the trigger fire on behaviour that is actually a large improvement.
Corrected in all four places before the ruling. **The correction strengthens the
case it was offered as evidence for**: the seat referred to its own history *less*
often than claimed (2.2%, not 6.7%), so the instrument was needed more, not less.
The script now prints `both_invoked_and_cited` and `pct_of_invoked` as the
headline and keeps `cited_but_did_not_invoke` visible so the gap cannot re-hide.

**Second-order lesson, worth more than the arithmetic.** The same review this
metric scored `cited_precedent = 0` is the one I quoted to the owner as the best
evidence the change worked — the guardian reasoning its way out of a deflection.
So the metric **under**-counts good behaviour in one direction and **over**-counts
it in the other (`%deflect%` matches the word alone, and the new prompt *itself*
says "deflected upward", so a seat echoing its own instructions scores a
citation). I had already told the owner about the undercount and still quoted the
overcounted headline in the same message. **A metric you have just discovered is
wrong in one direction deserves a check in the other before you quote it again.**

Family: two-independent-filters-read-as-an-intersection,
adjacent-counts-acquire-an-implied-denominator,
the-figure-the-workstream-is-judged-by-was-the-unchecked-one,
disagreeing-snapshots-of-a-fixed-population,
found-one-direction-of-error-and-quoted-the-other.

---

## 2026-07-27 — four bug files, four figures, three wrong: a DB query stood in for the rendered artefact

**Who:** the render/page-build triage sweep, re-grounding `bugs_open/080`, `095`,
`098`, `111` against the live system before reporting them as open.

**The claims, and what they actually were:**

| case | filed claim | measured 2026-07-27 |
|---|---|---|
| `111` | "**8 of 14** live sites show a heading over nothing" | **2** — `gamesdesign.co.uk`, `relojistas.com`. Five of the eight named render a *populated* Contact block |
| `080` | "**No duplicate exists today.** [VERIFIED]" | `robot-hands.com` has carried the duplicate since **2026-07-08**, i.e. *before the case was filed*; both URLs serve 200 |
| `098` | detector query offered as ready to run fleet-wide | returns **4 rows, all false positives**; 0 true positives |
| `095` | mechanism: wrong `slot_name` → assembly finds nothing | `getPageSections` **never reads `pages.sections`**; the empty assembly came from `rendered_html` being NULL |

**What caught it:** curling the deployed page and reading the function, instead of
trusting the query the file shipped with. For `111`, `curl https://<domain>/` and
stripping tags out of the `.footer-contact` block — thirty seconds per site.

**The single mistake underneath all four.** Each file measured its blast radius
with a query against *the table the author reasoned from*, never against **the
thing the user sees**. `111`'s query reads `site_specs.identity.contact`; the
footer renders from the **`sites.email` / `sites.phone` columns** — a different
store entirely, populated on 13 of 14 sites. The query was not wrong SQL; it was
SQL about the wrong object, and it returned a confident, plausible, precise
number. `080`'s survey filtered by `page_type`, and the duplicate row it was
looking for is typed `section-index` — so the filter excluded the very thing it
was written to find, and reported the absence as VERIFIED.

**The cheap check that would have caught it:** **for any claim about what a page
displays, the evidence is the fetched page.** A DB query is evidence about the
database. Before writing "N sites are affected", fetch one site the query says is
affected *and* one it says is not — a two-request check that discriminates. If the
claim is about rendering, `curl | grep` outranks any `SELECT`, and CLAUDE.md
already says so: *"Trust the rendered artefact, not the status."*

**Second lesson: [VERIFIED] on an absence is the shape to distrust.** `080` wrote
"No duplicate exists today. [VERIFIED 2026-07-26]" — the marker made an unchecked
*population* look like a checked *fact*. The query ran and returned nothing;
what went unchecked was whether it could have returned the thing. **An absence is
only as strong as the filter that produced it, so an absence claim must quote its
filter** — and when the filter names a column (`page_type`) the bug is *about*
disagreeing on, it cannot be the instrument.

**Cost:** none reached production — all four were corrected in the case files
before anyone acted (commit `2a1e86544`). The near-miss is `098`: a detector built
to its filed query would have opened 4 work items against healthy nav on day one
and found nothing real, which is precisely the discredited-detector failure
`bugs_open/033` and `/071` already describe.

Family: narrow-filter-defines-the-conclusion,
queried-the-database-about-a-question-only-the-page-can-answer,
verified-marker-on-an-unchecked-absence,
an-absence-claim-must-quote-its-filter,
the-detector-was-never-run-before-being-recommended.

---

## 2026-07-27 — a verification recipe that names a parameter the system does not have (`bugs_open/108`, post-roll triage sweep)

**The claim.** `bugs_open/108`'s *"How to verify a fix"* section instructs:
*"Defect A, induced: **point the indexer at a deliberately old ref**, run the
refresh, and assert the banner says STALE despite `updated_at = now()`."* Written
with full confidence, in a file that is otherwise unusually well-grounded (every
other claim in it carries its query or its file:line).

**What caught it.** Reading the scheduled task's `input_data` while checking
something else:

```sql
SELECT name, last_completed_at, input_data::text FROM scheduled_tasks WHERE name='code-index-refresh';
--  code-index-refresh | 2026-07-27 13:35:30+00 | {"repo": "agentchassis", "owner": "gqls", "language": "go"}
```

There is **no ref, branch or commit parameter to point.** `commit_sha` arrives from
an upstream repo-analysis step (`code_symbols_actions.go:174`,
`commit_field: "repo_analysis.commit_sha"`), so the indexed ref is whatever that
step happened to fetch. The recipe is not merely awkward — it is **not executable**,
and the fix candidate it belongs to (candidate 1) silently acquires extra scope:
you must first *add* the ability to name a ref before you can test that naming an
old one is caught.

**The cheap check that would have caught it:** before writing "point X at Y",
read X's actual parameters — one `SELECT input_data` on the scheduled task, or one
grep for the config key. This is the same check as the 2026-07-26 entry
(*"I recommended a config change without checking the config key is read"*) applied
to a **verification** step rather than a fix step, which is why it slipped: a
verification recipe reads as a check rather than as a claim, so it does not attract
the scepticism a claim would.

**Cost:** none yet — caught before anyone attempted the induced test. The cost it
*would* have had is the expensive shape: a session budgets an induced-fault
verification, discovers mid-run that the knob does not exist, and has to re-scope
the fix it had already gated on that verification.

**Generalisable form.** *A "how to verify" section is load-bearing prose and is
graded by nobody.* Fix candidates get council review; evidence gets quoted and
re-checked; verification recipes are read only by the one session that eventually
executes them, by which time the author is gone. **A verification step that names a
knob, a flag or a parameter deserves the same "does this exist?" check as a claim
about behaviour** — and it is cheaper, because it is always one grep or one
`SELECT`. Mark unchecked ones `[UNVERIFIED RECIPE]` rather than letting them read
as instructions.

Family: the-artefact-exists-so-the-capability-is-assumed,
grep-the-config-key-before-calling-it-a-win,
a-verification-recipe-is-a-claim-in-a-checklist-costume.

---

## 2026-07-27 — I linked two bugs by their matching counts, without checking they measured the same thing

**The claim.** Triaging the backlog I found `bugs_open/072` (contact-info cannot
render) reporting **"8 of 13 live sites"** and `bugs_open/111` (footer Contact
heading over nothing) reporting **"8 of 14 live sites"**. I queried `site_specs`,
found exactly 8 sites with nested-only `identity` keys, and wrote into the 072
case file that 111 was **downstream of 072, the same eight sites**, that fixing
072 would stop the heading stranding on **seven** of them, and that the two must
be sequenced 072 → 111. Committed as `ca53bc19c` and messaged to the session
triaging 111.

**What was actually true.** The footer does not read `site_specs.identity` at
all — it renders from the **`sites.email` / `sites.phone` columns**, a different
store, populated on **13 of 14** sites. Measured from the rendered footers of all
14 live sites, 111 affects **2 sites, not 8** (gamesdesign.co.uk and
relojistas.com), and only **one** of those is reachable by any contact-data fix —
relojistas is an owner ruling of *no contact route*, so it needs the gate no
matter what 072 does. So 072 buys at most one site here. The two "8"s were two
different measurements of two different populations that happened to coincide.

**What caught it.** The render-cluster session, because it measured the
**rendered artefact** (14 live footers) instead of reasoning from the store I had
guessed at. It also surfaced that relojistas renders an empty anchor *despite* a
populated `sites.email`, i.e. a third path feeds that render — which my story
could not have accommodated.

**The cheap check that would have caught it:** open the template and see which
field it interpolates. One grep for `footer-contact` in `content_components`
shows `{{if .email}}` and, one step up, where `.email` is bound from. Ten
seconds. **I never established the reader** — I matched two numbers and inferred
a mechanism to join them.

**Cost:** one wrong sequencing recommendation committed to a case file and sent
to another session, plus a re-argued priority (111 is a 2-site cosmetic fix, not
a fleet-wide one riding on 072). Caught same-hour, before anyone acted.

**Generalisable form.** *Two bugs reporting the same count are not evidence of
the same cause — a coincident denominator is the weakest possible join.* The
count is the least specific thing a bug file publishes; fleets have many
populations of a similar size, and "8 of 14" will collide constantly across 39
open cases. Worse, a matching count feels like *corroboration* — it arrives with
the emotional weight of independent confirmation while carrying almost no
information. **Join two bugs by a shared reader or a shared write path, named to
file:line — never by their arithmetic.** This is the standing
`writes-the-field-is-not-reads-the-field` rule, which I had in memory and did not
apply: I asserted where the data came *from* without once opening the thing that
consumes it.

Family: writes-the-field-is-not-reads-the-field,
narrow-filter-defines-the-conclusion,
a-verification-recipe-is-a-claim-in-a-checklist-costume.

---

## 2026-07-27 — "bugs_open/029 has halted every build since 19 July" (gemini_content_provider)

**The claim.** That the Gemini workstream's last acceptance test (P7: build one
page and read the copy) was blocked by `bugs_open/029`, which had "halted every
build fleet-wide since 19 July". Written into `bugs_open/107`, into the
workstream handoff, and into two commit messages (`bfbbb7cfa`, `5bd32602a`) — all
of which then told the next thread to go and wait on somebody else's bug.

**Why it was false.** `build-dispatch-loop` had **62 COMPLETED orchestrations on
07-26 and 30 on 07-27**; its only CANCELLED rows were two on 07-24. Page builds
were completing throughout — `ai-agent-orchestration.com/model-directory`
COMPLETED at 02:27, 08:27 and 14:27 on 07-27. The claim was already false on the
day it was written, not overtaken by events.

**What caught it.** A triage sweep that listed pods before querying anything.
`agent-build-dispatch-loop-*` pods were visibly Running and Completed minutes
old, which is impossible if the loop has been dead for eight days.

**The cheap check that would have caught it:** one unfiltered group-by.
`SELECT date_trunc('day',created_at)::date, status, count(*) FROM
orchestration_states WHERE owner_agent_type='build-dispatch-loop' GROUP BY 1,2
ORDER BY 1 DESC;` — three rows, and it contradicts the claim outright. I had
looked only for *my own* work item's row, found it unclaimed, and reached for the
first open bug whose title matched the shape of what I was seeing.

**The second, worse omission: I cited a bug without reading its corrections.**
`029`'s own file carries a CORRECTED DIAGNOSIS from `23e58e1bf` (07-21, six days
earlier) stating that the trigger is **an image roll** — a transient window after
each deploy, not a standing outage. Everything I needed to not make this claim
was inside the file I was citing.

**Cost.** A handoff, a bug file and two commits pointing the next thread at an
unrelated, actively-owned bug; and — the expensive part — it stopped me looking.
Because I "knew" why P7 could not run, I never asked whether the writer could
reach Gemini at all. It cannot: spawned pods are never given `GEMINI_API_KEY`
(`bugs_open/112`, filed from this sweep). The wrong blocker concealed the real
one for the rest of the session.

**Generalisable form.** *An open bug whose title matches your symptom is a
hypothesis, not a diagnosis — and its own file will often already refute it.*
Before adopting another case as your blocker: (a) read its corrections, not just
its headline, and (b) verify its stated effect is happening **now**, with a query
that could come back negative. A bug file is a claim written on a date; the
effect it describes may be transient, fixed, or have been mis-scoped from the
start. Adopting one as your blocker is also the cheapest possible way to stop
investigating, which is what makes it dangerous rather than merely untidy.

Family: narrow-filter-defines-the-conclusion, prior-art-search-goes-stale,
verify-the-failing-branch.

---

## 2026-07-27 — "dropping `cta_text` makes the gate omit the anchor" — it refilled it, in English, pointing at a 404

**The claim:** carrying out an owner ruling that relojistas.com has no contact route, I
removed two self-referential hero CTAs by deleting `cta_text` from `content_data`,
reasoning from the component's own guard, which I had read:

```
{{if and .cta_text .cta_url}}<a href="{{.cta_url}}" …>{{.cta_text}}</a>{{end}}
```

**What actually shipped**, on both live pages:

```html
<a href="/contact.html" class="btn btn-primary">Get Started</a>
```

English "Get Started" pointing at a 404, on a wholly Spanish site — **worse than the
Spanish-text-on-a-dead-link I was fixing.** `component_library.go` defaults both fields
when absent, and the refilled pair then *satisfied* the guard.

**What caught it:** a live sweep of every page after the re-render, not the DB. The stored
`content_data` looked exactly as intended — the defect exists only in the rendered output.

**The cheap check that would have caught it:** `grep 'cta_text\|cta_url' component_library.go`
— **fifteen seconds, and I had already read the very line that morning** and quoted it into
the `bugs_open/071` sighting (`defaultString(ctx.CTAUrl, "/contact.html")`). *Before deleting
a field to suppress output, grep that field name in the defaults layer: if it is defaulted,
deletion is the one action that guarantees the default fires.* Generally: **knowing a default
exists and reasoning about what happens when you remove its input are two different acts**,
and the first feels like the second.

**Cost:** ~25 minutes on a live site, and two pages briefly worse than before. Recovered by
giving both heroes real destinations. The finding was worth more than the cost — it is the
sharpest statement of 071's mechanism (the default fires *upstream* of the guard, so it
manufactures the condition the guard tests for; "no CTA here" is therefore unrepresentable in
`content_data`), and it rules out every data-side fix candidate in that case, since they all
work by writing `content_data`.

Family: the-default-fires-upstream-of-the-guard, i-had-already-read-the-line,
the-DB-looked-right-and-the-page-did-not.

---

## 2026-07-27 — I investigated the owner's design report and filed a duplicate of two bugs my own workstream had already filed

**Where:** `bugs_open/110` → renumbered `112` → renumbered `115`, filed within about
forty minutes of the owner's report, before grepping `/bugs_open/`.

**The claim:** that the palette divergence and the thin imagery were undiagnosed, and
needed a new case file.

**Why it is false:** a sibling thread **in this same workstream**, working from the same
owner report, had already filed `bugs_open/113` (the layout's light literals fill the
slots the generated palette never supplies) and `bugs_open/114` (21 generated images
live and serving, 3 referenced), plus `features_open/026` and a working
`scripts/render_audit.py`. Their measurement is strictly better than mine: headless
Chromium, every visible text node, background composited through transparent ancestors —
**101 WCAG-AA failures across five pages**. I had computed six ratios by hand from the
declared variables and would have missed anything involving transparency or inheritance.

**What caught it:** a filename collision. I went to commit and found another `110`
already there, then `111`, then `112` — and only while resolving that did I read the
neighbouring files and discover 113 and 114 were mine-in-all-but-authorship.

**The cheap check that would have caught it:** `ls bugs_open/ | tail -20` — three
seconds, and it is written in CLAUDE.md as "grep before you file", in this workstream's
own memory, and in the standing docs I had read that morning. I skipped it because the
owner had asked a direct question and filing felt like progress. **Urgency from a human
is exactly when the dedup check gets skipped, and exactly when several threads are most
likely to be looking at the same thing** — the owner's report went to more than one
session.

**Second, smaller, and its own lesson:** my free-number check was
`ls bugs_open/${n}_* bugs_closed/${n}_*`, which exits non-zero when **either** glob
misses — so it reported 112 free while `bugs_open/112_*` existed, and I renamed onto an
occupied number. *A multi-argument `ls` is not a per-argument existence test.* Build the
used-set once and test membership.

**Cost:** low and self-inflicted — two renames and a rewritten file, no wrong
information published. The duplicate never reached the owner because I re-measured
before answering, which is the only thing that went right here.

**Tally:** "grep the index before filing" 1→2.

Family: urgency-suppresses-the-dedup-check, the-shell-test-that-tests-something-else.

## 2026-07-27 — I planned a fix to code owned by two active workstreams, and only checked afterwards

**The claim.** In an approved plan for the webdesign.co.uk dead-link fix, I wrote
Step 2 as: change `NormalizePagePath` (`platform/orchestration/datahelpers/links.go`)
so the platform stops treating `/tools` as equivalent to `/tools/index.html`,
"through the council gate, then build and roll".

**Why it was wrong.** Not the diagnosis — that part holds, and is now recorded in
`bugs_open/071`. What was wrong was assuming the change was mine to make.
`bugs_open/071` **already contains a section reasoning about that exact function**,
concluding the repair belongs at the writer and not in the normaliser. 071 is owned
by `brochure_component_library` (68 commits/14d) and `gauntlet_dead_cta`
(60 commits/14d); the adjacent `092` is owned by `bugfix_079_phantom_link_gate`,
which committed to it the same day. I would have shipped a competing fix into three
active workstreams' code.

**What caught it.** Reading 071 before editing — which I only did because I paused
to check whether my finding was already filed. Luck, in other words, dressed as
diligence.

**The cheap check that would have.** `scripts/who-owns.py 071` — advisory, ~0.3s,
no cluster calls. I ran it, but *after* writing the plan rather than before. CLAUDE.md
already says to run it "before routing work AT an existing bug"; what this adds is
that **writing a plan that touches a bug's code IS routing work at it**, and the
check belongs before the plan, not before the edit. By the time a plan is approved
the direction feels settled and is much more expensive to reverse.

**Tally note.** This is the second entry in this file about acting on an
insufficiently-checked ownership/prior-art assumption (cf. the 2026-07-26
"capability that had been in production for weeks" class). The recurring shape is
not a missing fact but a missing *look*, and the look costs under a second.

---

## 2026-07-27 — I reported 41 broken images. Six were broken.

**Thread.** brochure_component_library, fundamentallyai.com contrast/imagery pass.

**The claim.** My new render-audit tool reported **41 broken images across 7 pages**.
I very nearly wrote that number into the owner's log and into `bugs_open/114` as the
scale of the imagery failure — it fitted the story I already had (that the site's
generated assets were never wired in), which is exactly why it went down so easily.

**Why it was wrong.** A headless render fires every image request on a page at once.
Our own origin throttles that burst and resets connections, so the browser reports
`naturalWidth === 0` for images that are perfectly present. Re-checking each one
serially over HTTP: **35 of the 41 returned 200.** Only 6 were genuine — all of them
one writer-invented path (`/images/illustrations/*.svg`). The first file in the list
was `/assets/images/logo.jpg`, which is on every page of the site and obviously fine;
that is what made me look.

**What caught it.** Noticing that the site's *logo* was in the broken list. A wrong
number containing one absurd entry is luckier than a wrong number that is uniformly
plausible — if the throttling had spared the logo I would have shipped all 41.

**The cheap check that would have.** Re-request each failure over HTTP before
reporting it, serially, with a retry. It is now built into `scripts/render_audit.py`,
which reports only statuses that survive that pass and prints the discarded count as
"N slow-loading image(s) re-checked OK".

**Tally note.** *"Retry before condemning"* was **already in my own memory** as a
landmine from 2026-07-25 ("cache-busted probing in a tight loop throttles the origin
and reads as broken links"), written by this same workstream. I had read it this
session. Knowing a landmine is not the same as having the check wired in: the memory
told me to be careful, and being careful is not a mechanism. The entry is only worth
the tally if the next tool that measures a live site inherits the retry by default —
which is the difference between a note and a fix.

---

## 2026-07-27 — I called `design_intent.palette.reference_values` a pin. The prompt calls it a suggestion.

**Thread.** Same session, deciding how to apply a corrected palette.

**The claim.** That updating `design_intent.palette.reference_values` would hold the
core colours steady across a `webdesign-agent` run, so running the agent was a safe
way to regenerate a stylesheet. My own memory index says so in as many words:
*"generic_theme misfires fleet-wide → webdesign re-rolls core colours each run; **pin
via design_intent.palette.reference_values**"*.

**Why it was wrong.** The prompt hands those values to the model like this:

> *"A design direction has been set for this site. You have creative freedom to
> interpret this intent. **The reference values are starting points — you may adjust
> them to better express the described character.**"*

It is an invitation, not a constraint. And the merge rule gives the model's output
authority over the palette row for all eight core slots, so a run can and does move
them. Proven, not inferred: re-rendering the layout template locally against the
palette row and diffing against the served stylesheet showed every structural rule
matching byte-for-byte while **all five core colours differed by a shade** and
line-height differed — the live file was never generated from its own palette row.

**What caught it.** Reading the prompt text before dispatching, because the run was
going to cost minutes and I wanted to know what it would do. Had I dispatched first,
the agent would have re-rolled the exact colours I was correcting and the result
would have looked like my fix had failed.

**The cheap check that would have.** Read the step's `prompt_template` before
trusting any claim that a spec value is authoritative. A "pin" that lives in a
prompt is a request; only code that rejects a contradicting value is a pin — as
`enforceLayoutScheme` is for `background`, and as nothing is for the rest.

**Tally note.** Second entry in this file about **trusting a remembered
characterisation over the artefact it describes**. The memory was not fabricated; it
was a fair summary of an outcome ("colours churn, this field helps") that hardened
into a mechanism claim ("this field pins them"). Summaries of behaviour drift into
claims about mechanism, and the drift is invisible because both sound like knowledge.

---

## 2026-07-27 — "8 of 14 live sites render an empty footer contact block" — I measured the table whose NAME matched the concept

**The claim:** filed as the justification for `bugs_open/111`, with a query and a list of
eight named domains, from:

```sql
SELECT s.domain, ss.data->'contact'->>'email', ss.data->'contact'->>'phone'
FROM sites s LEFT JOIN site_specs ss ON ss.site_id=s.id AND ss.aspect='identity' …
```

**Why it is false:** the renderer does not read `site_specs.identity.contact`. It reads
**`sites.email`**, which is populated on **13 sites** with a deliberate house convention,
`<name>@contactforsales.com` — including all eight I listed as having nothing. Once
relojistas' chrome was actually regenerated, its footer rendered
`relojistas@contactforsales.com` correctly, and the gate I had added correctly omitted the
null phone line.

The empty block that started it all was **stale chrome frozen on 2026-07-16**
(`bugs_open/117`), from before `sites.email` was set — a four-day-old artefact read as a
live rendering defect.

**What caught it:** rendering it. The block came back populated, which no reading of my own
measurement could have predicted.

**The cheap check that would have caught it:** **find what populates the field before
measuring a proxy for it.** `grep -rn 'ctx.Email' component_library.go` leads to
`sites.email` in seconds. I picked the table whose *name* matched the concept ("identity →
contact → email") over the column the code actually reads, and then measured it precisely
enough — 14 rows, named domains, a reproducible query — that it looked like evidence.
**Precision in the wrong place reads exactly like rigour.**

**Cost:** a filed bug whose headline severity was wrong for about three hours, corrected in
place before anyone acted on it. Cheap here; a fleet-wide "8 sites are broken" figure quoted
into a handoff would not have been.

**Same investigation, same shape, twice more** — all three are one error repeated:
- *"the footer nav comes from the `pages` query"* — it comes from `site_nav_items`.
- *"`InjectFooter`'s skip-guard freezes the footer"* — **the single-page path never calls
  `InjectFooter`**; `grep -n 'InjectFooter' rerender_single_page_action.go` returns nothing.
  I filed a bug on that mechanism and had to rewrite it.
  *Check: before asserting a function is the cause, grep the calling path for its name.*

Family: measured-a-proxy-for-the-real-source, first-plausible-function-is-not-the-path,
precision-in-the-wrong-place-reads-as-rigour.

---

## 2026-07-27b — I invented the shape of the thing this workstream exists to stop people inventing

**The claim.** A validator, a migration and a set of passing tests, all asserting that an
experience-register entry's `contract` is an object carrying a `triggers` array, each trigger
keyed `when`/`then`. Committed, council-submitted, council-**approved**, and the migration
applied to the live database.

**What was true.** All nine harvested entries — the only entries that exist — use `contract` as
an **array** of clauses keyed `control_role` / `primitive` / `outcome`. **The live write path
would have refused 9 of 9 real entries.** Two more of the same shape came out of the same look:
`CC-006` has a deliberately EMPTY contract (nothing a visitor does drives a count-up stat band;
its behaviour lives in `automatic_triggers`), so a "contract must be non-empty" rule refused a
legitimate entry; and `honesty_clauses` and `latency_envelope` are fields the harvest produces
for which the table had no columns, so they would have been **silently dropped on write**.

**Why every check I ran passed.** The fixture was an inline literal I wrote myself:

```go
func validHarvestedEntry() map[string]interface{} {
    return map[string]interface{}{ ... "contract": map[string]interface{}{"triggers": ...
```

The function was *called* `validHarvestedEntry`. Nothing about it was harvested. Seven behaviour
tests, a lockstep test against the migration, induced-fault probes in both directions — all green,
all measuring the code against a copy of its own assumptions. **A fixture invented to satisfy the
code under test proves only that the code is self-consistent.** The name made it worse, not
better: it asserted provenance the value did not have, so every later reading of that test
inherited a false premise.

**The council did not catch it either**, and could not have: reviewers see the submission, not
the repo. Approval is not evidence about facts nobody checked.

**The cheap check.** Load the real files. It is three lines, it was available from the first hour,
and it is what found the bug in the end:

```bash
python3 -c "import json,glob; [print(f, type(json.load(open(f))['contract']).__name__) for f in glob.glob('harvest/entries/*.json')]"
```

**The aggravating factor, and the reason this entry is worth its length.** This is the
experience-register workstream. Its founding finding — written by me, the day before, in its own
PLAN and SUMMARY — is that harvesting bottom-up from live implementations catches shapes invented
top-down, and that doing it in that order *"is what caught the ten errors"*. I then built the
validator from the harvest **notes** rather than the harvest **entries**, and reproduced the exact
failure mode the workstream was created to prevent. Writing the lesson down does not apply it.

**Cost.** Contained but not zero: one council round spent approving a shape that was wrong, a
migration applied to the live database that then needed 239 to complete it, and a day of work
that would have shipped a write path refusing every input it will ever be given. Nothing reached
production — the Go had not rolled.

**The fix that generalises.** Not the corrected shape: the **fixtures are now the real files.**
`validHarvestedEntry()` reads `CC-001` off disk, and two tests load all nine — one asserts the
validator accepts every one, the other that every field they carry has a column. Proved by induced
fault (restoring the non-empty rule fails naming `CC-006`). The drift cannot recur silently
because there is no longer a second copy of the truth to drift from.

**Tally.** *Test against the artefact, never against a copy of your assumption about it* — a new
row, and the sharpest instance yet of the family that already includes "look at the real values,
not the name" and "prove the artefact is current before reasoning from it". Both of those are
about reading data; this one is about **manufacturing** it and then reading that.

Family: the-fixture-that-agrees-with-the-code, a-name-is-not-provenance,
the-lesson-written-down-and-not-applied.

---

## 2026-07-27 — a handoff's landmine outlived the bug it warned about (webdesign.co.uk)

**The claim.** `HANDOFF_2026-07-27b_continue_here.md` §1, written that afternoon as the
cold-start document for the next thread, on the order of building the news page:

> *"only after it builds, re-render chrome to publish the News nav link. **That order is not
> optional** — the nav row already exists in the DB, so re-rendering chrome early puts a 404 in
> the header of all 98 pages (`bugs_open/049`'s exact shape)."*

**Why it was false.** The platform already drops the item. Every chrome path passes
`NavFetchableOnly` to `GetNavItems`, and `applyNavVisibility` (`nav_tables.go:152`) filters any
target matching `NeverDeployedPagePredicate` — which `/news/index.html` matches exactly
(`deployed_at IS NULL`, `build_status='planned'`). It is in the running binary, not just the tree:
pod-grep of the drop log on `v1.0.1175` returns 1. The guard shipped with the **049 fix itself**
(`a9083d51b`, `759cb2b77`) — the very bug the warning cites as the hazard.

**What caught it.** Reading the function. I was already inside `getNavItemsFromTables` looking at
the nav query when a comment at `nav_tables.go:368` said the deployment predicate had **moved to
`applyNavVisibility`**. One jump. The code volunteered its own correction.

**The cheap check that would have.** The same one, before writing "not optional" into a cold-start
doc: open the function that renders the nav and follow where it filters. Under a minute. The
handoff instead reasoned from the DB row (`site_nav_items` has an active News row pointing at an
unbuilt page — true) straight to the rendered outcome (a 404 ships — false), with the renderer
never opened. **This is the exact shape CLAUDE.md's 2026-07-19 correction was written about:
"the failure mode is not missing information — it is not looking."**

**The new part, and the reason this is worth a row of its own.** The claim was *true when the
mechanism was first learned* and quietly stopped being true when another workstream fixed it.
**A landmine written into a handoff has no expiry date and no owner.** `bugs_open/049` was closed
by the CTA/link-integrity thread; nothing in that closure could reach into a site workstream's
handoff and retire the warning it had spawned. So the warning propagated forward under its own
momentum, into the one document a fresh thread is told to trust first — where it is *more*
load-bearing than an ordinary note, because a cold-start reader has no context to doubt it with.

The asymmetry is what makes it dangerous rather than merely stale: a **stale hazard warning costs
silently**. It never fires, never contradicts anything, and just makes people take a longer route
— here, it would have made the news sequence look fragile and, worse, ruled out the chrome
re-render that Route B of the Cloudflare beacon needs. Nothing would have looked wrong.

**Tally.** Two families. *Read the function, don't infer the behaviour from the data* — a
recurrence, now at least the third instance, and the one CLAUDE.md already names. And a genuinely
new one: **a warning inherits the confidence of the document it lands in, not the freshness of the
fact behind it.** Practical form, cheap enough to adopt: when a handoff states a hazard as a
constraint on someone else's future work, **cite the code that makes it true, not just the symptom**
— `NormalizePagePath`-style file:line, so the next reader can re-check in seconds instead of
believing. A hazard with no citation should be re-derived before it is obeyed.

Family: read-the-function-dont-infer-from-the-data, the-fixed-bug-whose-warning-outlived-it,
a-cold-start-doc-is-believed-harder-than-it-is-checked.

---

## 2026-07-27 — "zero live instances fleet-wide" was true when I measured it and false twenty minutes later, and I had already put it in a council submission

**Thread:** bugs thread (bugs_open sweep). **Claim:** in the council submission for the
`bugs_open/095` fix, the central risk argument was *"the defect shape is ZERO rows
fleet-wide, so failing it breaks nothing today"*. I measured it, quoted the query, dated
it, and treated the risk calculation as settled.

**What it actually was.** True at ~18:05 UTC. **False by ~18:35**, while the council was
still deliberating on the submission that rested on it: `oufe.com/tool-recovery-waterfall`
entered the defect shape at 18:16:53 — one component row, zero usable, one planned
section. The submission argued for a new hard-error path on the grounds that nothing in
the estate could hit it; by the time the verdict came back, something could.

**What caught it.** A council seat (`debug_historian`) objecting to something *else* — that
my blast-radius query was scoped `WHERE p.status='active'` while bucketing by
`build_status`, mixing two columns. That objection turned out to be wrong on its own terms
(archived pages are also zero), but re-running the query to answer it is what surfaced the
new row. **The objection was not correct and was still worth exactly what it cost.**

**The cheap check that would have.** There isn't one, and that is the point — this is not
a check I skipped, it is a figure with a shorter half-life than I assumed. The practical
form is: **re-run the measurement when you act on it, not when you write it.** The gap
between submitting and committing was thirty minutes and the number moved inside it.

**Why this is a different family from the usual staleness row.** The standing rule is
"ground every figure against the live system before repeating it from another doc", and
its stated horizon is *days* ("volumes, counts and statuses go stale within days"). That
horizon is wrong for anything the fleet actively produces. A count of *pages in a broken
state* is not a slow-moving inventory figure; it is a queue depth, and queue depths move
in minutes. **The half-life of a figure is set by what produces it, not by what kind of
number it looks like.**

The asymmetry that makes it worth a row: a zero is the most dangerous figure to cache,
because it is the one that *removes* a safeguard from the argument. "17 rows" ages into
"some rows" and the reasoning survives. "0 rows" ages into "1 row" and the conclusion
built on it — *this change cannot break anything today* — is simply false, with nothing
in the sentence to signal it.

**Second, smaller call in the same sitting.** The `bugs_open/080` submission said the
gap-planner was "the fourth" page-creation surface and that three others canonicalised.
There are **four** canonicalising creation call sites plus two read-only lookups; it was
the fifth. I had run the grep, read seven results, and then wrote the number from the bug
file's framing instead of from my own output. The `guardian` seat asked for the
enumeration to be confirmed rather than asserted. Cheap check: **when you have already run
the command that answers a count, take the count from the output, not from the prose you
are paraphrasing.**

**Tally.** *Zero is the figure most worth re-measuring at the moment of action* — new.
*Take the number from your own output* — a recurrence of narrow-filter/figure-carried-
forward, third-ish instance. And an argument for the council that is not about correctness:
**a wrong objection still bought a real correction**, because answering it forced a re-run.

Family: a-zero-ages-worse-than-any-other-figure, figure-carried-forward-from-prose,
the-half-life-of-a-count-is-set-by-what-produces-it.

---

## 2026-07-27 — idea.uk: I nearly reported a pass from a test that could not have failed

**The call.** After one full engine run I checked the deployed report for the three copy
defects fixed on 07-26. All three absent, every format marker present. I was one sentence
away from writing "the copy fixes are proven in a rendered artefact."

**Why it was wrong.** Two of the three **could not have appeared in that run whatever the
code did**:

- the doubled full stop fires only when the *submitted text itself* ends in a full stop and
  the template appends a second — **I wrote the submission, and it didn't end in one**;
- the malformed score line lives in a block that never rendered, because the run hit
  `NO FURTHER IDEA CLEARED THE BAR` and nothing was scored. The fix marker
  `(each out of 5)` was absent for the same reason — which I first read as a *second*
  finding rather than as the same absence twice.

Checking for a defect in a section that does not exist is not a check. The fixes remain
**unproven in a rendered artefact**; settling it needs a submission ending in a full stop
that is strong enough for at least one idea to clear the bar.

**What caught it.** Reading the artefact's structure instead of stopping at the grep
result. The `(each out of 5)` marker coming back `False` was the thread to pull: a *fix*
marker missing alongside all *defect* markers missing is incoherent unless the whole
section is gone. Had every marker agreed, I would have shipped the false claim.

**The cheap check, and why this one is different.** The standing rule is *verify the
failing branch* — third instance on this workstream. But the previous two were code paths I
controlled. **Here the failing branch is a property of the INPUT**, and I chose the input,
so I had authored a test that could not fail without noticing. The check is one question,
asked before reading the result, not after: **"could this input have produced the bug?"**
If the answer isn't a confident yes, the run is not a test of the fix, whatever it prints.

Corollary worth keeping: **when a fix marker and its defect marker are both absent, suspect
the section, not the fix.** Two markers disagreeing about whether a feature exists is
structural evidence, and it is cheaper to read than a re-run.

**Second, smaller call in the same sitting — a landmine I had already written down.** My
deploy script ran `set -e` with a `grep -ac …` **negative control** in the verification tail.
The control did its job — zero matches in the rollback binary — `grep` exited 1, and `set -e`
killed the script mid-verification. The deploy itself had already succeeded, so for a moment
it read as a failed deploy. This is the same family as the recorded `printf | grep -q` under
`pipefail` trap: **a grep asserting absence exits non-zero on success**, so it must never sit
inside `set -e`/`pipefail` unguarded. Cheap check: **capture greps into `$(…)` and print
them, or append `|| true`** — never let an intentional zero-match terminate the run that is
checking for it.

**Tally.** *Could this input have produced the bug?* — new, and the sharper form of
verify-the-failing-branch for input-driven faults. *A grep that asserts absence fails under
`set -e`* — recurrence, second instance, and the first was already in memory when I did it.

Family: verify-the-failing-branch, a-green-result-from-an-input-that-cannot-fail,
absence-assertions-exit-non-zero.

---

## 2026-07-27 — every misstep this week had a designed interceptor, and I used none of them

The owner's observation, and it reframes the whole week: *"these are what the
diagnose loop and council and architect council are all designed to intercept."*

He is right, and it is worse than a list of individual errors. The platform
already contains a machine for each of the mistakes below. Every one was
available, none was reached for, and in each case the reason was that the mistake
did not feel like the kind of thing that needed one.

### The four, and the interceptor each one walked past

**1. "Nothing in the estate looks for this" — the diagnosis loop.**
I asserted a structural property of the platform (a whole class of claim is
invisible to every scanner) from one function's documented limit, and wrote it
into a live council seat. CLAUDE.md's rule is not ambiguous: file a diagnosis
**before** committing to a root cause whenever the claim is durable, and it names
this exact case — *"a mechanism, a structural property of the platform, a cause
that lives outside the symptom."* Debug directly only when the fix is local and
self-evidencing.

Mine was neither. It was a claim about four subsystems, and it reached other
sessions' standing instructions. **A REFUTED verdict costs one run; this cost a
migration, a mirror, five documents and an afternoon of building on it.** The
loop would have refuted it by reading the function I skipped — which is precisely
its documented behaviour, the 9.5-minute refutation of the two-rerender-paths
claim.

**2. "Build a promise register" — the reuse seat on the council.**
I proposed new machinery for something the estate already models three ways
(`evidence_base` with its unused `kind: capability`, the EXPERIENCE_PLAN's promise
ledger, CTA integrity's label-implies-a-destination). The council has a
`review_reuse_agent` seat whose **founding incident is a session reinventing a
trigger+triage SQL pair that already existed.** That seat exists for this exact
error, and the proposal never went in front of it.

The owner intercepted it manually instead, with "we have existing functionality —
look hard at what exists." That is a human doing a seat's job.

**3. Migration 223 — the council gate.**
It changed a live reviewing agent's standing instructions and carried a false
premise in its contract text. It was DB config rather than `platform/`, so the
gate's stated scope did not compel a submission. But the *content* was a durable
claim about what the platform guarantees, which is exactly what a compliance seat
reads for — and I was editing that seat while making the error it catches.

**4. `slot_name='main'` and the AI voice — no loop, and that is the finding.**
Bug 095 was filed and my own prepared SQL still carried the defect. Rule 3 was in
the writer prompt and the copy still broke it. Neither is a reasoning failure a
council would catch, because neither involved a decision. They are artefacts that
already existed when the rule arrived.

### The pattern the owner named, stated plainly

**Three interceptors exist. All three are opt-in, and all three are opted into by
the person about to make the mistake.** The diagnosis loop is filed by the thread
that thinks it already knows the cause. The council is submitted to by the thread
that thinks its plan is sound. Nothing routes work to them; confidence is the
gate, and confidence is exactly what is broken in the moment they are needed.

CLAUDE.md already says this, in the correction dated 2026-07-19: *"Confidence is
not a signal. The wrong claim felt obvious; that is exactly why 'obvious' cannot
be the gate. Full context is no protection, because the failure mode is not
missing information — it is not looking."* I read that file at the start of the
session and then did the thing it warns about, twice.

### The cheap checks, in the order they would have fired

- **Before writing "nothing/never/no X in the estate":** that is a universal
  negative about a large system. File the diagnosis, or read the source of every
  mechanism you are about to name. My sentence named four and I had read one.
- **Before proposing anything new:** grep the spec's open-questions section, not
  just `bugs_open/` and `features_open/`. The answer to this week's design
  question had been written down, with a proposal and a trigger, and was sitting
  in `SPEC_claims_verification.md:250-252` the whole time.
- **Before applying a prepared file:** re-read it against bugs filed *since* it
  was written. Mine predated the bug that describes its defect by one day.
- **After shipping copy:** count the tells in the rendered output. The rule being
  in the prompt is not evidence the output followed it, and measuring took one
  command.

### The structural point, which outlives these four

Filing a rule changes nothing that already exists. The estate now has: a writer
rule the existing copy broke, a bug report the existing SQL broke, a register that
went stale with no watcher, and a deferral that became policy because its trigger
had no watcher. **Every one of those is a rule without a sweep.**

The interceptors have the same shape one level up. A loop nobody is routed to is a
rule without a sweep. Making them non-optional — even just for claims of a named
class, like "no X exists in the platform" — is the difference between machinery
that exists and machinery that fires.

Family: interceptor-existed-and-was-not-used, confidence-is-the-gate-and-should-not-be,
rule-without-a-sweep, universal-negative-from-local-evidence.

---

## 2026-07-27b — the check I wrote to catch a bug reported CLEAN because it was scanning nothing

**The claim.** Having added `check_logged_model_output` to `pattern-check.py`, I ran
it and reported the result as meaningful. It found nothing, and nothing was wrong.

**What was true.** It was **vacuous**. The check gated on the *file* containing
`GenerateText` — but in the code it was written to police, the LLM call lives in
`handlers/defend.go` and the log sink in its sibling `handlers/ailog.go`. The file
holding the defect never mentions `GenerateText`, so it was skipped entirely. The
check scanned zero files and printed a clean result, which is exactly what a
passing check looks like.

**What caught it.** A **positive control**: auditing `a37a2037c` — the commit that
had *introduced* the raw-excerpt logging — and requiring a finding. It produced
none, which is the only reason I looked at the gate.

**The cheap check that would have.** The one I then used. **A new check must be
run against a commit known to contain the defect, and must fire.** Otherwise
"0 findings" is indistinguishable from "0 files examined" — and the second is
worse than no check at all, because it is now a documented reassurance.

**Cost.** None realised; caught before it was committed as working. But the shape
is the point: this is the same failure as the *vacuous pod-grep* already in this
ledger twice (a comment is not in the binary; a typed const is inlined away), now
in a third costume. **Every "we now detect X" claim needs the negative case
demonstrated, not just the positive one asserted.**

**Same session, smaller:** I `git stash`ed a shared working tree to isolate a
negative control, when the command I was running (`pattern-check --ref <sha>`)
diffs committed state and never saw my uncommitted work. The stash was pure risk —
three other sessions' stashes are in that list — and bought nothing. *Check: does
the tool I am running actually read the working tree before protecting it from
itself.*

**And a THIRD instance of the same family, hours later, in the very section
warning about it.** The runbook block I wrote to verify the island rebuild —
titled *"verify against the RUNNING CONTAINER"* — shipped with two defects:
it grepped **`/app/tools-api`** when the dockerfile puts the binary at
**`/tools-api`** (so every check would have returned 0 and read as a failed
deploy), and its negative control grepped **`JSONError(c, 502`**, which is Go
*source* and is not in a compiled binary at all — it returns 0 before and after
and proves nothing. Caught by running the greps against the built image before
shipping. *Check: a verification command is code too; run it against a known-good
AND a known-bad input before writing it down as the procedure.*

**The tally is the point.** Three in one day, each in a different costume —
a file-gated linter, a source-string grep of a binary, a wrong path — and one of
them written *inside the warning about the other two*. The generalisation that
covers all three: **an assertion that cannot fail is not an assertion.** Every
"we now verify X" needs its failing case demonstrated once, at the moment it is
written.

Family: a-clean-result-and-an-unrun-check-are-identical, vacuous-detector,
protected-against-a-risk-the-tool-does-not-have,
the-verification-command-is-code-too.

## 2026-07-27 — three zeros in one session, none of them findings: a null result means nothing until the instrument is shown able to return non-zero

Bug-sweep thread, the session that shipped `bugs_open/112`. Not one wrong claim —
a *shape*, which turned up three times in about two hours and was caught three
times only because a control happened to be present. It very nearly cost a
wasted image build.

**Zero #1 — the anchored pod-grep.** Verifying that the Gemini fix was in the
newly built image, I ran:

```
strings /app/agent-chassis | grep -c "^GEMINI_API_KEY$"    -> 0
strings /app/agent-chassis | grep -c "^ANTHROPIC_API_KEY$" -> 0   <-- the tell
```

The natural reading of the first line alone is "the fix is not in the image" —
and the next action is to rebuild and re-roll, which costs a tag, a push and a
fleet restart. It is wrong. `ANTHROPIC_API_KEY` has been in that binary for
months, so a 0 for it proves the **grep** is broken, not the image: Go string
constants are laid down inside a larger string table, so `strings` emits them as
substrings of longer lines and `^...$` anchors can never match. Unanchored, the
real answer was `GEMINI_API_KEY -> 2` on the new image against `1` on the old.
**What saved it was including a positive control in the same loop** — not skill,
habit.

**Zero #2 — the payload that "wasn't there".** The drain's own trigger script
prints the query to read its result:

```sql
SELECT jsonb_pretty(collected_data->'complete'->'result'->'response') ...   -- returns EMPTY
```

Run after a `COMPLETED` sweep, that empty result reads exactly like "the sweep
ran and did nothing" — which is *the very failure mode `bugs_open/033` is about*,
so it was a plausible finding rather than an obvious error. It is a wrong path:
the sweep step's `output_field` is `revalidation_result`, and the payload was
sitting at `collected_data->'revalidation_result'` all along, with 50 items in
it. Caught by dumping `jsonb_object_keys(collected_data)` instead of trusting the
documented query.

**Zero #3 — the filtered grep, inherited.** `bugs_open/101` evidences "four
config keys are inert" with:

```
$ grep -rn "max_pages" --include=*.go . | grep -i webscrape      # no hits
```

True, and true of the claim it supports. But the conclusion a reader carries away
is "these four keys are dead", and the obvious next action — a fleet-wide
`UPDATE ... WHERE config ? 'max_pages'` — would strip a **live** page cap from
`build-site-planner` (80) and `site-planner` (20), where the code logs *"preserved
pages exceed max_pages; keeping all preserved, dropping all net-new"*. Silently
uncapping two site planners is worse than the bug being fixed. Caught by re-running
it unfiltered.

### The through-line

**A zero, an empty and a no-hit are all the same object: the absence of a signal.
Absence has two causes — the thing is not there, or the instrument cannot see it
— and nothing in the output distinguishes them.** Every one of these three read
as a finding, and in each case the finding was about my own instrument.

The asymmetry is what makes it dangerous: a *non-zero* result is self-validating
(the instrument demonstrably works), so we check positives casually and negatives
not at all — while negatives are exactly what we build conclusions on
("nothing shipped", "it did no work", "the key is dead").

### The cheap checks, in the order they would have fired

- **Never run a `grep -c` for a verification without a positive control in the
  same command** — a string you KNOW is present. If the control returns 0, the
  grep is wrong. This is already the rule for pod-greps in CLAUDE.md; the new
  part is that it applies to the *syntax* of the grep, not just the choice of
  marker. An anchor, a `-w`, a locale, a non-UTF-8 file will each silently
  return 0.
- **Before believing an empty payload, dump the container's keys.**
  `jsonb_object_keys(collected_data)` costs one query and answers "wrong path" vs
  "no work" outright. A runbook's printed query is a claim like any other and
  goes stale when a step's `output_field` changes.
- **Before carrying someone else's negative into an action, re-run it
  unfiltered.** The filter that made their claim precise is the filter that makes
  your generalisation false. (This is the recorded
  `narrow-filter-defines-the-conclusion` landmine, arrived at from a new
  direction: here the filter was in *evidence I inherited*, not in a query I wrote.)

Family: absence-is-not-evidence, positive-control-missing,
narrow-filter-defines-the-conclusion, instrument-failure-read-as-finding.

## 2026-07-27 — "2 resolved out of 50" was never a rate: a `LIMIT` is a filter you did not write

Same session, `bugs_open/033`. The review-queue drain's first dry run came back
`scanned 50, resolved 2, still_holds 2, unknown 46`. The seed's own header
predicts *"roughly 51 resolved"*. Two against a predicted fifty-one is a
four-percent hit rate against an expected fourteen, and I got as far as drafting
"the drain does not work on the real population" before checking.

It would have been wrong, and expensively so — it condemns a mechanism that had
just been seeded after being missing for two days.

**The sweep's query is `... WHERE status='needs_human_review' ORDER BY created_at
ASC LIMIT $n`.** Oldest first, deliberately (the code says so: *"the oldest items
are the ones most likely to be describing a page state that no longer exists"*).
So the 50 it scanned were the 50 **oldest**, and the composition was
`needs_section_data` 20, `content_rewrite` 16, `needs_content_page` 11,
`empty_section` 3 — with **zero `unresolved_cta` and zero
`required_fields_missing`**, which are two of the three types the revalidator
covers and together are 115 items, 30% of the queue. The run never met the
population it was written for. The estimate was neither confirmed nor refuted.

**Why it was convincing:** `scanned: 50` reads as a sample, and 50 of 381 sounds
like a reasonable one. Nothing in the payload says "this sample was selected by
age, and age correlates with type". The `LIMIT` did not feel like a filter
because I did not write it — it arrived as a config default (`max_items: 50`).

### The cheap check

**Before quoting any figure from a capped query, read its `ORDER BY` and ask what
the ordering correlates with.** `LIMIT` without `ORDER BY RANDOM()` is not a
sample, it is a selection — and the selecting column is almost always correlated
with something you care about (age with type, id with creation order, name with
tenant). One query settles it: compare the batch's distribution against the
population's.

Corollary worth keeping: **a default in a config file is still a filter.** The
recorded landmine is about filters taken from the question; this one was taken
from a default nobody chose, which is harder to notice precisely because no one
authored it.

Family: narrow-filter-defines-the-conclusion, sample-is-not-the-population,
a-default-is-a-decision.

---

## 2026-07-27 — "This plan fixes routing and content" — a submission's rationale describing work that was not in its own edits array (council HIGH objection)

**The claim, written into a council submission and judged against:** the rationale
of `18fe4035` opened by naming three defects — unrouted lanes, `review_prior_art`
missing from `code_check_fields`, and declarations-only `content` — and then stated
**"This plan fixes routing and content."**

**It was false, and the council caught it as its only HIGH objection.**
`bug_historian`, verbatim: *"nothing in the edits array touches
agent_definitions.task_workflow … or code_check_fields … Only the content-side
defect is actually edited. Defect (1) … is left fully exploitable after this plan
ships."* It went further and named the consequence I had missed: *"The council's ONLY
forward-fitness voice (review_architecture, per the D9 ruling this plan itself
cites) runs on feature-designer — the exact lane left unrouted."* My own plan cited
the ruling that made its omission matter.

**Caught by:** the council gate, not by me — on a submission I had pre-flighted for
quote fidelity, JSON schema and scope, and still shipped with a rationale describing
work that was not in it.

**The cheap check that would have caught it:** read the rationale's verbs against the
edits array as a checklist — *fixes routing* → which edit? *fixes content* → which
edit? One pass, mechanical. **A rationale is not a summary of the problem, it is a
claim about the edits, and the council judges the plan AGAINST it** — so any verb in
it that no edit implements is a defect I authored, not context I supplied.

**Cost:** one council round. Answered by *shipping* the routing half rather than
arguing it away, so the round produced a real change; but the round was spent on my
own overclaim, and a reviewer's attention on it is attention not spent on the design.

Family: rationale-claimed-work-not-in-the-edits, my-own-citation-indicted-me,
the-verbs-are-the-checklist.

---

## 2026-07-27 — "CREATE INDEX CONCURRENTLY, to be safe" — caution copied from big-table habit onto a 4,535-row table, where it was strictly worse

**The claim, written into a council submission as a design choice:** the migration
sketch used `CREATE INDEX CONCURRENTLY IF NOT EXISTS` for the new trigram index,
with the risk list noting it "must not run inside a transaction — the migration
runner should be checked for that". I wrote the caveat and **did not do the check**,
then shipped the plan with the unresolved question in it.

**It was wrong twice over, and both seats that objected were right.**
(a) The migration runner's **dry-run probe deliberately wraps the file in a poisoned
transaction** (`run-migrations.sh:129-139`) to prove the file reaches its own COMMIT —
so `CONCURRENTLY` would have **failed the probe** even though it would work on apply,
i.e. the safe-looking choice breaks the safety mechanism. (b) `code_symbols` holds
**4,535 rows**: a plain index build is milliseconds and its brief lock is irrelevant.
`CONCURRENTLY` bought nothing and cost the file its conventional, guarded,
probe-clean shape — 86 of 92 live migrations carry their own BEGIN/COMMIT.

**Caught by:** `guardian` and `debug_historian`, independently, in the same round.

**The cheap check that would have caught it:** `SELECT count(*)` on the table I was
indexing — a number I had **already queried twice that session** for other reasons
and did not connect. **`CONCURRENTLY` is a size-dependent choice, and I never asked
the size.** Corollary worth keeping: when a risk list says "X should be checked",
that is a note to *me*, not to the reviewer — shipping it unresolved converts my
homework into their objection.

**Cost:** none realised. Fixed in round 2; the migration is now conventional.

Family: caution-without-measuring, the-safe-looking-option-broke-the-safety-check,
i-had-already-run-the-query-that-answered-it,
an-unresolved-risk-note-is-unfinished-work.

---

## 2026-07-27 — "reuse doesn't fit here" — rejecting a shared function, then hand-writing a copy explicitly built to match its convention

**The claim, written into a council submission:** that
`internal/analysis.ReadSymbolBody` could not be reused by the indexer because "that
function needs an analyser Output plus a checked-out root … which is the bundle
assembler's problem, not the indexer's". I then specified a new `sliceLines(fileLines,
start, end)` in the indexer, with the sketch comment *"matching
analysis.ReadSymbolBody's convention"*.

**The rejection was false, and my own sketch comment was the proof.**
`reuse_agent`, verbatim: *"the plan itself concedes the two functions must stay
behaviorally identical, which is exactly the condition under which a shared helper …
is the correct fix, not two hand-synced implementations."* The interface mismatch was
real but irrelevant — it argued against reusing *that signature*, not against
extracting the primitive underneath it. Worse, `symbolbody.go`'s own header **already
asked for this**: *"this is intended as the ONE slicer … collapse it onto this
function."* I quoted that file's line numbers in the submission and did not read its
header.

**Caught by:** the reuse seat, on the strength of a comment I wrote myself.

**The cheap check that would have caught it:** if the justification for a new
function contains the phrase *"matching <existing function>'s convention"*, the
correct edit is to extract the shared part — the sentence is the diagnosis. More
generally: **an interface mismatch is an argument against a call, never against a
primitive.** Ask what the smaller shared thing is before writing the second copy.

**Cost:** none realised — round 2 extracts `analysis.SliceLines` and both callers use
it, which is a better change than the one I proposed.

Family: rejected-reuse-then-cloned-the-behaviour,
my-own-comment-was-the-refutation, read-the-header-of-the-file-you-cite,
interface-mismatch-argues-against-the-call-not-the-primitive.

---

## 2026-07-27 — "the bug file says mirror it onto the gate, so mirror it" — a fix candidate read as licence, with the deliberate reason one file away (NEAR MISS)

**The claim, nearly acted on:** implementing the routing fix, `bugs_open/108`
candidate 5 reads *"Either mirror `code_lookup` onto the gate, or stop `prior_art`
promising an answer there."* I read that as licence to add the step to all three
lanes, which is the obvious reading and would have been a live config change to the
busiest council in the fleet (36 runs/day).

**It would have been wrong.** `099_SYNC_gate_roster.py:24-29` states the exclusion
and its reason: `code_lookup`/`repropose` *"serve the blind reproposer, which the
gate has no equivalent of (its authors read the objections themselves)."* The gate
has no reproposer for the results to feed. The same reason is what **includes**
`feature-designer` — that lane *does* have a blind reproposer — so one principle
produced both answers, and only reading it revealed that.

**Caught by:** opening the file the bug cited, before acting on the bug's
recommendation. Nothing prompted it beyond the number `:24-29` sitting in the case
file unread.

**The cheap check that would have caught it:** a bug file's fix candidates are
**hypotheses written by someone who may not have read the thing they propose
changing** — including when that someone is me, earlier the same day. Before acting
on a candidate that removes an asymmetry, find out whether the asymmetry was
*chosen*: `grep -n "not mirror\|deliberately\|no equivalent" <the mirroring script>`.
An asymmetry with a comment explaining it is a decision; one without is a bug.

**Cost:** none — caught before acting. Logged because the failure mode is a "fix"
that deletes a deliberate design decision, which is expensive precisely because the
diff looks like tidying and the bug file appears to authorise it.

Family: fix-candidate-read-as-licence,
an-explained-asymmetry-is-a-decision-not-a-defect,
the-reason-was-one-file-away, my-own-earlier-writeup-was-the-thing-misleading-me.

> **Pattern across the four above, worth more than any one of them.** Three were
> caught by the **council gate** and one by reading a file before acting. None was
> caught by thinking harder about the plan, and all four were in a submission I had
> already pre-flighted for quote fidelity, schema and scope. That is the
> architecture-seat workstream's own thesis arriving as evidence against its author:
> a reviewer with the written record in front of it caught what its author, who wrote
> that record the same day, did not.

---

## 2026-07-27 — I wrote a check with no failing branch, inside the commit that adds a guard against checks with no failing branch

**Where:** verifying the `platform/colour` extraction (`features_open/026` Phase 2a).

**The claim:** that I had confirmed `isDarkHex` and `isDarkColor` were not wired to
each other's implementation — the one mistake in that refactor that compiles, passes
every existing test, and silently inverts the renderer's light/dark classification. I
had named it myself as the residual risk in the council submission.

**Why it is false:** my probe printed both functions for `#000000`, `#666666`,
`#888888`, `#ffffff`. **The two agree on all four.** Had they been swapped, the output
would have been byte-identical. The check could not fail, so it confirmed nothing — and
I had just written, in the same submission, that a checker which cannot discriminate is
worse than none.

**What caught it:** reading my own output and noticing every row agreed. That is luck,
not method; a row of four `true/true, false/false` pairs looks like a clean pass.

**The cheap check that would have caught it:** **before trusting a probe, run it against
a case you know is wrong.** Ten seconds: swap the two delegations, re-run, see whether
the output changes. It does not.

The discriminating window turned out to be six greys wide — `#767676`–`#7b7b7b`, where
`isDarkHex` is false and `isDarkColor` is true — because their crossovers sit at ≈
`#777777` (does white beat black) and ≈ `#7c7c7c` (luminance < 0.2). **A binary
classifier can only be tested near its boundary**, and I sampled the extremes, which is
where every implementation agrees. Now a permanent guard
(`color_util_delegation_test.go`), verified by inducing the swap.

**Cost:** none realised. The wiring was correct. But I would have reported "checked" on
evidence that could not have contradicted me, which is how the entries above this one
in this file get written.

**Tally:** "run a census against a known-positive control before reporting the count"
1→2 — same shape, different subject: *induce the fault before believing the probe*.

Family: the-check-with-no-failing-branch (again, and this time in the guard against it),
sample-the-boundary-not-the-extremes.

---

## 2026-07-27 — "extract a shared slicer" — proposing to create a function that already existed, in a plan whose own subject was unchecked absence claims

**The claim, written into a council submission (round 2) as the fix for a reuse
objection:** that `internal/analysis` needed a shared line-slicing primitive
*extracted* from `ReadSymbolBody`, so the indexer and the on-demand reader would stop
being two hand-synced implementations. I specified
`func SliceLines(lines []string, startLine, endLine int) string` as new work, and
cited `symbolbody.go`'s header asking for "the ONE slicer" as support.

**It already existed.** `internal/analysis/symbolbody.go:133`:
`func sliceLines(src []byte, start, end int) (string, error)` — unexported, and
`ReadSymbolBody` already calls it. The correct edit was two lines (export it, keep an
alias), not an extraction. I had cited that file **twice**, by line number, in two
consecutive rounds.

**Caught by:** `prior_art_librarian`, whose entire charter is "does it propose
BUILDING something that already exists" — and which was answering `code_checks` for
the first time that evening **because I had added it to `code_check_fields` an hour
earlier**. The seat I had just wired up caught me proposing to build a thing we had.

**The cheap check that would have caught it:** `grep -n "func.*[Ss]lice" internal/analysis/*.go`
— one command against the very file I was citing. **Citing a file's line numbers is
not reading it**; I quoted `:31` and `:133` was forty lines further down. Note also
that the code index could not have helped here even if consulted: it holds
declarations only, so a `content` search would have returned nothing — the same
defect this plan exists to fix, shielding the error from the tool meant to catch it.

**Cost:** none realised — caught in review, and the resulting edit is smaller and
strictly better. The irony is the finding: a plan about unchecked absence claims,
containing an unchecked absence claim, in a codebase whose absence-checking index is
broken in the way the plan was written to repair.

Family: proposed-building-what-already-existed,
cited-the-file-without-reading-it, the-seat-i-had-just-wired-caught-me,
the-broken-tool-would-not-have-saved-me.

---

## 2026-07-27 — "86 of 92 live migrations carry their own BEGIN/COMMIT" — a precise figure lifted from a code comment and presented as measurement

**The claim, written into a council submission and used to justify a design choice:**
that the migration should carry its own `BEGIN/COMMIT` because "86 of 92 live
migrations" do. It reads as a census. It sounds checked. It has a denominator.

**I never counted anything.** The figure is quoted from a **comment inside
`run-migrations.sh`** (`:132`), written by someone else at some earlier date, and I
repeated it in my own voice. Measured properly: **263** migration files (excluding
UPPERCASE sidecars), **173** containing `BEGIN` — **66%, not 93%**, against a
denominator nearly three times larger. The conclusion it supported happens to survive
(carrying `BEGIN/COMMIT` is still the majority convention and what the dry-run probe
expects), which is luck rather than method.

**Caught by:** `prior_art_librarian`, flagging it as "precise-sounding but
self-reported, not council-verified", and by this council's standing rule that
rationale-reported numbers get an attached check rather than trust.

**The cheap check that would have caught it:** two `ls | wc -l` commands, which took
about fifteen seconds when I finally ran them. Generally: **a number carried out of a
comment inherits that comment's date and none of its provenance.** A comment is a
claim someone made once about a corpus that has since grown — here, by 171 files. If
a figure is load-bearing enough to put in a rationale, it is cheap enough to re-count;
if it isn't worth re-counting, it isn't worth citing.

**Cost:** none realised. But this is the *third* entry today in the same family — a
stale-or-borrowed figure repeated in my own voice — after the "6 of 90" metric and the
"3 of 87" baseline. The tally is the point: this is now a check worth making reflexive
rather than a lesson worth learning again.

Family: number-lifted-from-a-comment-and-voiced-as-measurement,
a-comments-figure-inherits-its-date-not-its-provenance,
precise-sounding-is-not-checked, third-in-this-family-today.

### 2026-07-27 — voice v4 — an assertion that checks PRESENCE and not POSITION passes a misplacement

**Asserted:** that the three new Voice & Style rules had been "appended to the
block" on `page-content-writer`. My apply script's guard verified each new rule's
text was in the `prompt_template`, printed
`OK: rule 1 rewritten, 3 rules appended, em-dash rule intact (16149 chars)`, and I
reported it as done in a commit message and to the owner.

**What was actually true.** The rules were appended to the end of the **template**,
not the end of the **block**. The block sits at char 272 of 16,150; the rules landed
at char 15,455 — **~11,500 characters away, after the JSON output instructions** —
and one of them still read *"that breaks the word-weight rule above"*, a reference
that no longer resolved anywhere near it. The guard passed because every string it
looked for was present. It never asked *where*.

**Caught by:** a different migration refusing itself. `241` tried to swap the block
for a placeholder on the assumption it was the final section, its own guard fired
(`template is only 289 chars - too much was cut. ROLLING BACK.`), and chasing *that*
error is what revealed the block's real position — and therefore the earlier
misplacement. **The second guard caught the first guard's blind spot.** Neither
would have surfaced it alone.

**The cheap check that would have caught it:** assert the **position**, not just the
presence. `position('{{.voice_style}}' in t) <= 500` is one clause. Generally:
a guard over a large text blob that only tests `LIKE '%needle%'` cannot distinguish
"in the right place" from "anywhere at all", and appends are exactly the operation
where that distinction is the whole point.

**Cost:** none realised — the misplaced rules were deleted by `241` v2, which
replaced the whole block with a placeholder anyway. But they were live in the
production prompt for roughly two hours, and every page built in that window (none,
as it happens, because `029` has builds stopped) would have carried a style rule
citing a rule 11,500 characters away.

**Three more from the same session, each caught by the owner rather than by me:**
- **I created a duplicate of the house voice inside the commit message warning about
  duplication.** Saw it, described it accurately, shipped it behind a
  `[KNOWN DUPLICATION]` comment. *A comment naming a defect is not a mitigation* —
  and it felt responsible, which is precisely why it was the wrong call. Refiled as
  `bugs_open/121` at the owner's direction; I had filed it as a *feature*, which
  puts a defect in the queue nobody treats as urgent.
- **I invented an override the owner never asked for** — a `voice_style_block`
  config switch with a present-but-empty opt-out, plus a paragraph defending the
  empty-vs-absent distinction. The owner meant *"a request has its own prompt in the
  request"*. *Check: when a directive contains a word like "override", ask what it
  means before building the mechanism.*
- **I put prose in Go.** Corrected by the owner: *"one place for the prompt, and
  probably not in go by choice."* Prompt text a non-engineer may want to edit does
  not belong where changing it costs a compile and a fleet roll.

**And one thing the tooling caught that I did not:** committing the fix, the
`pre-commit` hook printed *"migration + platform code in one commit — needs a staged
rollout order"*. It was right, and it is the observation the architecture seat
exists to make — a seat that has **never fired** (`bugs_open/121`). A twenty-line
shell hook beat a sixteen-seat council to the architectural point.

Family: presence-is-not-position, a-comment-naming-a-defect-is-not-a-mitigation,
built-the-mechanism-before-asking-what-the-word-meant,
the-second-guard-caught-the-first-guard's-blind-spot.

---

## 2026-07-27 — I called a settled convention "a decision for the owner", twice, in a bug file and a commit message

**Thread:** bugs thread (bugs_open sweep). **Claim:** in `bugs_open/080` and in the
commit that closed its code half, I wrote that robot-hands.com's two live news pages
were *"a live-site data decision for the owner, not a code fix"* and that *"which row
survives is yours"*.

**What it actually was.** Decided, shipped, and council-approved weeks earlier. The
section-index family convention in `page_canonical.go` (doc 029 Phase 0, extended by
`bugs_closed/015`) fixes the shape as `(name=<section>-index, url=/<section>/index.html,
page_type=<flavour>)`. `bugs_closed/015` names it "the family's stated design" and
`relojistas.com` is the live worked example. The relojistas running notes even record the
same repair being chosen deliberately: *"re-type noticias-index → news-index and drive
the build onto it (keep the Spanish /noticias URL + nav), rather than let gap-planner
mint an English /news.html."*

**What caught it.** The owner, saying it had been discussed extensively and telling me to
go and find it. Nothing in my own process would have.

**The cheap check that would have.** `grep -rn "index.html" docs/…/029_site_plan_and_reconciler*`
— thirty seconds. Or simply reading the header of the file my own fix called into:
`page_canonical.go` states the convention in its first twenty lines, and I had read that
file, quoted it in a council submission, and still did not connect it to the question of
which row wins.

**Why this is its own family and not just "grep first".** The rule I broke is not
"research before you assert". I DID research: I read the helper, quoted its header
verbatim as evidence, and measured the live rows. What I failed to do was notice that
**the convention I was citing as the FIX already answered the question I was calling
open.** The evidence for the decision was inside the evidence I had already gathered.

Deferring to the owner *feels* like the safe, humble move, which is exactly what makes it
dangerous: an unnecessary escalation looks like diligence and costs nothing visible. But
it pushes work back to the one person who cannot delegate it, it stalls a repair that was
already authorised, and — worst — it puts "undecided" into the durable record of a
question that was decided, so the next thread inherits a live decision as an open one.
**A false "this needs a ruling" is a documentation defect with the same shape as a false
finding.**

**Tally.** *The answer is often inside the evidence you already collected* — new.
*Escalation is not free and is not neutral* — new, and worth watching for: I should be
able to say WHY a question is the owner's (it trades off money, taste, or risk he owns)
rather than defaulting to escalation whenever a change touches a live site.

Family: escalated-a-decision-that-was-already-made, the-answer-was-in-the-evidence-I-had.

---

## 2026-07-27 — I called the spawn-wrapper pattern "proven". It succeeds 2 times in 4.

**The claim.** Designing the council-gate wrapper (`bugs_open/096` candidate 4), I told
the owner the spawn-then-call pattern was "proven in `0NN_fix_implementer_orchestrator.sql`
and diagnose-orchestrator" and that the precedent was "strong". I built on that and
recommended it over the alternative.

**What caught it.** Running the thing. The very first test submission stuck at
`spawn_council` / `AWAITING_RESPONSES` and never reached `call_council`.

**The cheap check I skipped**, one query, under a minute, available from the start:

```sql
SELECT owner_agent_type, status, count(*) FROM orchestration_states
WHERE owner_agent_type IN ('diagnose-orchestrator','fix-implementer-orchestrator')
  AND created_at > now() - interval '14 days' GROUP BY 1,2;
--  diagnose-orchestrator | COMPLETED | 2
--  diagnose-orchestrator | FAILED    | 2     <- both "timed out after 3 retries"
```

Two of four, over the table's entire retained history. I had *already run* a query against
these very rows — to measure the wrapper's ~8s overhead — and I filtered it to
`status='COMPLETED'`. **The filter I took from my question ("how long does a successful
wrapper run take?") deleted the evidence that answered the question I should have been
asking ("does it work?").** That is [[narrow-filter-defines-the-conclusion]] with the
failure rows sitting one predicate away.

**Why "proven" was the wrong word specifically.** My evidence for it was that the pattern
is *in daily use* and that feature-builder once produced a merged PR through it. Both true.
Neither is a reliability measurement. **"It is used a lot" and "it works" are different
claims, and a busy system supplies endless evidence for the first while saying nothing
about the second** — especially here, where the failure mode is a silent hang that looks
like slowness.

**Tally.** *`status='COMPLETED'` in an exploratory query is a conclusion, not a filter* —
recurring, and the second time this week a narrow filter answered confidently about a
small world. *"In daily use" is not a reliability measurement* — new; when about to call
something proven, count the failures, not the successes, and say the denominator out loud.

Family: narrow-filter-defines-the-conclusion, in-use-is-not-proven, counted-successes-not-attempts.

---

## 2026-07-27 — "241 is the next free migration number" — true when I looked, false by the time the plan was approved, and a reviewer had said so

**The claim, written into a council submission and into the filename of its first
edit:** that `docs/agent_docs/sql_for_agents/241_code_symbols_body_column.sql` was the
next free number, "so a concurrent session re-running it is harmless".

**It was true when measured and false 40 minutes later.** Another session committed
`241_page_writer_uses_voice_placeholder.sql` between my submitting the plan and the
council approving it. The approved plan therefore names a filename that is already
taken; implementation must use 242 and re-check again at write time.

**Caught by:** `guardian`, which objected (low) that *"241 being 'the next free number'
… can't be verified from the DB schema — this is a fact a directory listing would
settle, not SQL … flagging since it's asserted rather than checked this round."* I then
re-ran the listing out of diligence rather than expectation, and it had moved.
**Vindicated inside the hour.**

**The cheap check that would have caught it:** re-running `ls … | sort -n | tail -1`
**immediately before writing the file**, not when drafting the plan. The deeper error is
category, not timing: **I treated a shared mutable sequence as a fact I could measure
once.** `CLAUDE.md` already says this about `git status` — "your session-start snapshot
goes stale within minutes" — and a migration number is the same kind of object. It was
the *second* numbering collision I hit the same day; the first was two threads claiming
`D9` in one decision register within an hour.

**Cost:** none realised — caught before any file was written. Recorded in the handoff
rather than edited into the approved JSON, so the approved artifact keeps matching the
verdict it earned.

Family: a-shared-sequence-is-not-a-fact-you-measure-once,
true-when-measured-false-when-used, the-low-severity-objection-was-the-right-one,
second-numbering-collision-in-one-day.

---

## 2026-07-27 — a VERIFY script written to catch the plan's most dangerous failure could not itself have run (`content::bytea`)

**The claim, written into a council submission as the answer to the most serious risk
in it:** that `content_hash` integrity would be guarded *as code, not as a request that
a human look* — a companion sidecar asserting
`content_hash IS DISTINCT FROM encode(sha256(content::bytea),'hex')` returns zero rows
before and after. I offered this specifically to answer two seats objecting that the
invariant was asserted in prose.

**The expression is invalid and errors on the first row.** `content::bytea` does not
hash the text — it tries to **parse** the text as a bytea literal, and Postgres returns
*"invalid input syntax for type bytea"*. The working cast is
`convert_to(content,'UTF8')`. So the guard I presented as the fix for "guarded only by
someone remembering to look" was a guard that could not execute. The underlying
mechanism is right, and is now verified live: **4,535 of 4,535 rows match**
`encode(sha256(convert_to(content,'UTF8')),'hex')`, 0 mismatches.

**Caught by:** `editquality`, objecting (medium) that the formula *"is asserted, not
verified against the live schema/trigger definition — if the actual hash function
differs … the VERIFY script will produce false positives/negatives"*. It suspected the
wrong failure (a different algorithm) and was right about the real one (my SQL was
broken), because both come from the same omission: I never ran it.

**The cheap check that would have caught it:** running the query. Once. Against the live
table, which I had queried perhaps a dozen times that session for other things.
**A verification script is code, and unrun code is not a verification** — writing an
assertion and shipping it unexecuted reproduces, one level up, precisely the defect it
was written to prevent.

**Cost:** none realised — caught in review, corrected in the handoff before
implementation. But note what it nearly did: had it shipped, the mass-re-embed guard
would have errored during a migration window and been read as a broken migration rather
than a broken check.

Family: the-guard-i-wrote-could-not-run,
unrun-code-is-not-a-verification, right-objection-wrong-mechanism,
i-had-the-table-open-a-dozen-times.

---

## 2026-07-27 — I built a mechanism out of three timestamps and never opened the function that answers it. Refuted in five minutes.

**The claim.** The council wrapper's spawn stuck at `spawn_council` /
`AWAITING_RESPONSES`. I had ms-precision evidence that the child replied 1.92 s
*before* the parent transitioned (19:50:24.189 vs 19:50:26.109), and
`SpawnAgentAction` visibly sleeps 5 s + 5 s before returning `await_response:true`.
So I concluded the reply landed in a window the parent was not yet in and was
discarded. I marked it `[INFERRED]` and filed it — which was right — but I had
already written it into a bug file, a seed header **and a memory file** that
auto-loads into every future session.

**What caught it.** The diagnosis loop, verdict **REFUTED**, correlation
`eb8df254`, in about five minutes. `persistAwaitingStateWithRetry`
(`coordinator.go:1863-1879`) re-loads state on every attempt and returns early on
`"Response already arrived during state persist - continuing"`;
`processResponseClaimWithRetry` retries the claim for literally
*"response may arrive before awaited_request is inserted"*. Both paths already
cover the race I invented. I verified both citations in source rather than
trusting the verdict, and they are real.

**The cheap check I skipped:** open the function. One function. I had grepped
`coordinator.go` repeatedly that hour and read `continueExecution`,
`ProcessResponse` and `buildCallResult` — but never
`persistAwaitingStateWithRetry`, the one whose name is literally the thing I was
theorising about.

**This is the failure CLAUDE.md's "Diagnosis before debugging" section was
rewritten to describe, repeated inside the same subsystem it was written about.**
That section records a thread filing a structural claim from grep hits whose
functions it had never opened, refuted by the loop in 9.5 minutes. Mine took 5.
Reading the warning is not the same as being protected by it: I filed *because*
of that section, and still shipped the claim into three files first.

**The compounding error, which is the worse one.** Before filing, I cancelled the
stuck orchestration and deleted its Job — tidy-up, and defensible on its own terms
(a row parked in `AWAITING_RESPONSES` feeds the `029` saturation class). But the
loop then recorded that it could find **no orchestration row parked at a spawn
step** to examine, so it refuted me on static evidence with the runtime half
missing. **The failing row WAS the evidence.** Cleanup and evidence preservation
are in direct tension and I did not notice the trade at the time.

**Tally.** *Timestamps are a symptom, not a mechanism* — new; three precise
numbers feel like proof and constrain nothing about which code path ran.
*Open the function whose name matches your theory, before writing the theory
down* — recurring, now twice in this file. *Never destroy the failing artefact
until the diagnosis has run* — new, and the one I'd most want automated: the 090
trigger could refuse, or warn, when the orchestration named in the symptom is
already `CANCELLED`.

Family: built-a-mechanism-from-timestamps, never-opened-the-named-function,
cancelled-the-evidence-before-filing, wrote-it-down-before-i-checked-it.

---

## 2026-07-27 — "the indexer is already walking the file, so this needs no new file I/O pass" — it walks a JSON blob with no source text in it, and the whole plan depended on the opposite

**The claim.** In the council submission for D11 layer 1 (corr `18fe4035`,
APPROVED round 3), the rationale for the edit that populates `code_symbols.body`
read: *"symbolRow carries lineStart/lineEnd (set from td.StartLine/td.EndLine),
and the indexer is already walking the file, so this needs no new file I/O pass
and no re-parse."* The sketch followed from it:
`body, err := analysis.SliceLines(fileSrc, td.StartLine, td.EndLine)` — with
`fileSrc` appearing from nowhere.

**What is actually true.** `flattenSymbols(out analysis.Output)` walks a
**JSON-decoded `analysis.Output`**: paths, line spans, signatures, doc strings.
`analysis.FileInfo` has no source-text field. There is no file being walked and
no `fileSrc` in scope. The plan's central edit had **no stated source for the
bytes it was slicing**, and twelve council seats approved it.

**Why it works anyway, which is the uncomfortable part.** The LIVE `code-indexer`
workflow's first step is `analyse_repo_local`, which fetches the tarball into the
indexer pod's own temp dir and deliberately does not clean it up — so `out.Root`
is a real local path and the bodies can be read from it. **That is live config, not
code, and the repo's own seed for that agent
(`118_code_indexer_for_analyser.sql`) still shows the OLD wiring**
(`request_repo_analysis`), under which the analyser parses in a different pod and
`out.Root` names a directory that does not exist in the indexer's. Under the seed,
every read fails, every body is NULL, and the change ships **looking done and
being inert**.

So the claim was false, the design survived on a fact the claim did not mention,
and the repo's own SQL would have led a reader to the opposite conclusion.

**Caught by:** opening `flattenSymbols` to write the edit, then querying
`agent_definitions` for the live workflow because the signature made no sense.
Not by any reviewer — 12 seats, including `prior_art_librarian`, which had caught
a *different* absence claim in the same plan the round before.

**The cheap check that would have caught it:** *the sketch names a variable that
is not in scope at the site you are editing.* `fileSrc` was never defined
anywhere in the submission. A plan sketch is not compiled, and that is exactly why
an undefined identifier in one is a signal and not a typo — it marks the place
where the plan assumed a fact it never checked. **Read the enclosing function's
signature before writing the sketch that lives inside it.**

**Second cheap check, the one with legs:** *when a claim about runtime behaviour
can be settled from either the repo's seed SQL or the live `agent_definitions`
row, the live row is the fact and the seed is a historical document.* This tree
has seeds that were superseded by config edits months ago and never updated. The
same trap is already in this file under other names; this is the first time it
nearly shipped an inert feature rather than a wrong belief.

**Cost:** none realised — caught before the build. Had it not been, the cost would
have been a green deploy, an all-NULL column, and a `content` check that kept
returning zero rows while everyone believed the fix had landed. That is worse than
the original bug, because the original bug at least had `bugs_open/108` pointing
at it.

**Tally.** *Wrote it down before I checked it* — recurring, and now the dominant
family in this file by some distance. *The seed is not the system* — new.
*A council approving a plan is not evidence the plan is implementable* — new, and
worth saying plainly: the gate reviews reasoning against evidence, and an
undefined variable in a sketch is neither.

Family: wrote-it-down-before-i-checked-it, the-seed-is-not-the-system,
approval-is-not-implementability.

---

## 2026-07-27 — ran a provisioning script from the section that omitted a required parameter, and broke the site's deploy pipeline

**What I did:** ran the relojistas box convergence exactly as the runbook's **§P5.2** gives it:

```bash
DOMAINS="relojistas.com" LETSENCRYPT_EMAIL=<real> MODE=full bash /root/setup.sh
```

**What the same runbook says at line 43**, in its header block:

```bash
LETSENCRYPT_EMAIL=you@your-real-domain.tld DEPLOY_USER=deploy bash /tmp/setup.sh
```

`DEPLOY_USER` is load-bearing (`setup.sh:65-70`):

```bash
WEBROOT_OWNER="${WEBROOT_OWNER:-${DEPLOY_USER:+$DEPLOY_USER:www-data}}"
WEBROOT_OWNER="${WEBROOT_OWNER:-www-data:www-data}"     # ← the fallthrough I took
```

Without it the webroot is chowned to `www-data:www-data`, and **every vm-sites GitHub Action
deploy fails**:

```
rsync: mkstemp "/var/www/vm-sites/relojistas.com/.feed.xml.VAsIfL" failed: Permission denied (13)
rsync error: some files/attrs were not transferred (code 23)
```

The news pipeline could not publish. Repaired with `chown -R deploy:www-data /var/www/vm-sites`,
verified by actually writing as `deploy` and reading as `www-data` rather than by re-reading the
`ls` output.

**What caught it:** the owner pasting the Action's failure. **Nothing I ran would have.** I
verified the site (200s), the endpoints, the feed and the real-ip change — every check I chose
was a *read* of the public site, and the thing I broke was a *write* by a third party. The site
looked perfect precisely because the last successful deploy was still sitting there.

**The cheap check that would have caught it:** after any change to a directory a CI job writes
into, **write to it as that user** — `sudo -u deploy touch <dir>/.probe`. Ten seconds. More
generally: *when you change ownership or permissions, the verification is an attempted write by
the identity that lost access, never a read by yours.*

**And the reason it happened at all:** I read the runbook's §P5.2 and ran it. The header block
40 lines earlier had the fuller command. **A runbook with two invocations of the same script is
a runbook with one wrong one** — I took the nearer one because it was under the heading that
matched my task. Fixed in place: §P5.2 now carries `DEPLOY_USER=deploy` and a comment naming
this failure, so the two agree.

**Bonus cost, still unpaid:** setting `DEPLOY_USER` also installs
`/usr/local/sbin/site-engine-deploy` and its sudoers entry (`setup.sh:438-451`) — which is the
*intended* no-root route for swapping the engine binary, and exactly the operation that was
subsequently blocked. Dropping the parameter removed the sanctioned path to the next step.

Family: two-invocations-one-is-wrong, i-verified-reads-and-broke-a-write,
the-site-looked-fine-because-the-last-good-deploy-was-still-there.

---

## 2026-07-27 — I rolled the chassis on top of my own in-flight council run and killed it, then spent an hour reading "running" as slow

**Thread:** bugs thread (bugs_open sweep). **Claim:** across four status checks over
roughly seventy minutes I reported `bugs_open/105`'s council review as *"running"*,
*"still at review_guardian"*, and *"still deliberating"* — treating a step that had not
moved as a step that was working.

**What it actually was.** Dead since the moment I rolled the chassis.
`orchestration_states` for that run last advanced at **19:22:29**; the chassis pod I
replaced went down and its successor started at **19:22:02**. The step was executing on
the pod I killed. It never resumed, and it never would have: sixty-six minutes later it
sat on the same step with the same `updated_at`.

**What caught it.** Nothing in my own loop. I only looked at `updated_at` because the
user mentioned deploying a new chassis, which made me check what my roll had touched.
Every check before that read `current_step` and `status` — `review_guardian |
EXECUTING_STEP` — both of which say "running" forever on a dead run. **A step name and a
status cannot distinguish working from wedged. Only the clock can.**

**The cheap check that would have.** Two of them, and I skipped both:

- BEFORE rolling: `SELECT current_step, status, updated_at FROM orchestration_states
  WHERE status NOT IN ('COMPLETED','FAILED','CANCELLED')` — one query, and it would have
  shown my own council run mid-flight. I checked pod health, image tags, drift and
  registry state before that roll. I did not check whether I was about to interrupt
  anything, including my own work.
- AFTER, in every status poll: select `updated_at` alongside `status`, and read the
  gap. `EXECUTING_STEP` with a 66-minute-old `updated_at` is not a slow step.

**The blast radius was not just mine.** The dispatch thread's own residual says a wedged
head orchestration freezes the interactive lane until a pod roll, and nothing notices.
Two other runs stalled behind mine (`call_council` 19:58, `call_diagnoser` 20:08 — the
latter almost certainly the `bugs_open/097` diagnosis I had filed myself). The lane only
recovered when ANOTHER session rolled the chassis at 20:26. So a roll I performed to make
five fixes live cost an hour of a shared lane, and I reported the symptom of that outage
four times without recognising it.

**The asymmetry worth keeping.** CLAUDE.md warns about dispatching *within ~300s after* a
pod restart. The inverse — do not restart the pod on top of work already in flight — is
not written down anywhere, and it is the one I broke. A roll is not a read-only act just
because the deployment reports healthy afterwards: **every orchestration mid-step at that
instant is collateral, and the pod-grep verification that follows says nothing about it.**

**Tally.** *A status field cannot tell you a run is alive; only `updated_at` can* — new,
and it generalises past orchestrations to anything with a state column.
*Check what is in flight before you roll* — new. And a recurrence of a familiar one:
**I verified the thing I changed (the binary) thoroughly and never looked at what I
disturbed.**

Family: status-says-running-forever-on-a-dead-run, rolled-over-my-own-in-flight-work,
verified-what-i-changed-not-what-i-disturbed.

> **CORRECTED 2026-07-27, ~20 minutes after writing the entry above — I misidentified one of
> the two "collateral" runs, in the file about misidentifying things.**
>
> The entry claims *"two other runs stalled behind mine (`call_council` 19:58,
> `call_diagnoser` 20:08 — the latter almost certainly the `bugs_open/097` diagnosis I had
> filed myself)"*. The `call_diagnoser` identification is **wrong on both halves**. That
> orchestration (`7803075d`) advanced to `load_runtime` at 20:30:59, so it is alive, not
> stalled; and it carries no `fix_correlation_id`, so nothing tied it to my 097 run except
> that the step name matched the concept I was looking for.
>
> **That is the exact error pattern already logged on 2026-07-27 ("I measured the table
> whose NAME matched the concept").** I did it again, one file down, within the hour, while
> writing about being careless.
>
> **What survives.** My own council run (`0d6bb5f8`) is genuinely dead at my roll —
> unchanged at 69 minutes, and that is measured, not inferred. And the 097 diagnosis IS
> stalled: its last artifact is `bundle` iteration 3 at **19:25:10**, three minutes after
> the roll, with nothing in the 66 minutes since. So the *conclusion* about 097 stands on
> its own evidence (its artifact clock), which is what I should have cited in the first
> place instead of a same-named orchestration step.
>
> **What does NOT survive.** "Two other runs stalled behind mine" is unproven. One was
> alive. The other (`900327ca`, `call_council` at 19:58) belongs to another session and I
> have not established whether my wedge caused it. And the lane is plainly healthy now —
> 21 completions in the last 15 minutes — so "an hour of a shared lane" is an overclaim I
> cannot support.
>
> **The cheap check:** cite the artifact clock of the thing you actually care about, not an
> orchestration row whose step name sounds right. `diagnosis_artifacts.created_at` for MY
> correlation answered this in one query and needed no guessing about ownership.

---

## 2026-07-27 (evening) — I called a live page dead, and the cheap check was a page I had not touched

**The claim.** Verifying a Gemini page build on dartsonline, I probed
`https://dartsonline.com/sale` and got **404**. The DB said `build_status='deployed'`,
`deployed_at` set. I had two open bugs ready to explain exactly that shape —
`bugs_open/098` ("`deployed_at IS NOT NULL` means a deploy happened once, not that the
page is fetchable") and `bugs_open/120` (a merge commit skips the site deploy) — and I
was one sentence from writing "deployed but not fetchable, consistent with 098".

**What caught it.** Before believing it, I probed **`new-arrivals`, a page I had not
touched**, deployed cleanly on 07-26 → **also 404**. A page my build never went near
cannot have been broken by my build, so the defect had to be in my probe, not in the
system. It was: **this site serves `.html` extensions.** `/sale.html` → 200,
`/about.html` → 200, `/new-arrivals.html` → 200; every extensionless path → 404. The
page had been live and correct since 20:10:21, ten minutes before I looked.

**Why this one was well-disguised, and the general shape.** Having a *plausible open bug
that predicts your symptom* is what makes this dangerous. 098 predicted the exact
observation. The wrong belief was one grep away from looking well-researched — I would
have cited a real bug, with a real mechanism, against a real 404. Confirmation had a
ready-made home. **A hypothesis that explains your evidence is not thereby the cause of
it**, and the more apt the bug you have in hand, the less you interrogate the probe.

**The cheap check (~30 seconds):** before attributing a failure to the system, run the
same probe against something in the same batch that your change did not touch. If the
peer fails too, the fault is in your instrument. This is already memory
(`check-an-untouched-peer-in-the-same-batch`) — this row is the tally, and the point is
that I applied it *because it was memory*, not because anything about the 404 looked
suspicious. It looked entirely convincing.

**Distinct from the 07-27 "I measured the table whose NAME matched the concept" rows**
above: those were the wrong *query*. This was the right query pointed at the wrong
*URL shape* — and no amount of re-reading my own reasoning would have surfaced it,
because the reasoning was sound and only the input was wrong.

---

## 2026-07-28 — I wrote an acceptance test that could not pass, and it was one dispatch away from deleting a legal disclaimer

**The claim.** That the oufe.com recovery-waterfall tool worked. I had inspected the
rendered markup, seen every element present and correct, and described the tool as
working in a handoff. The Tier-4 browser acceptance was recorded as "owed" rather than
blocking.

**What happened when I finally ran it.** It failed 2 of 11 checks. Both interaction
checks in my criteria fence clicked the tool's consent gate `#rw-accept` first, and my
own note under the fence *argued that they must* — "the tool body is hidden behind the
condition-of-use gate and a hidden input cannot be filled". That reasoning is correct
and it is irrelevant, because it silently assumes each check gets a fresh page.

It does not. `internal/adapters/browserrunner/run_checks_action.go` opens ONE page per
profile (`:569`), navigates once (`:584`), then runs every check against that same page
(`:398`). The gate sets `display:'none'` on click. So check one opened it and check two
clicked an element that still resolved in the DOM and was no longer visible.

**The check I skipped** is the existing ENGINE row, incremented above: I proved the
fence against the runner I imagined rather than the one that would execute it. One
`grep` for `NewPage` in the runner — about fifteen seconds — settles it. I had written
a multi-step behavioural test against a harness whose execution model I had never read.

**What nearly made this expensive, and why it earns its own row.** The failure
auto-raised an `improve_tool` work item carrying my fence as `acceptance_test`, for
dispatch to `tool-improver`. Its task would have been: *make this tool pass a test that
requires the consent gate to be clickable twice on one page load.* The only ways to
satisfy that are to stop the gate hiding itself, make it reappear, or remove it — **all
three weaken or delete the disclaimer**, which is section B of the owner-approved
wording and the proximate placement the whole negligent-misstatement position rests on.
It is the one element on that page that most needs to be immovable, and nothing in the
loop knows that.

A green run afterwards would have looked like success. I cancelled the item `wont_fix`
by hand before it dispatched — that was attention, not a guard, and attention does not
scale to the next gated tool. Filed as `bugs_open/126`.

**The generalisation.** An automated repair loop inherits the authority of whatever
test it is given. A wrong test does not merely fail; it becomes a specification, and
the loop implements it faithfully against correct code. Every previous row in this file
is about a wrong *belief* costing a wrong *conclusion*. This is the first where a wrong
belief of mine was queued to be executed by another agent, on production content, with
a plausible green result at the end.

**What I got right, and want to make deliberate rather than lucky.** I did not trust
the failure's headline. "Locator resolved to `<button…>` … element is not visible"
reads exactly like a broken button, and the tool was the obvious suspect. What pointed
the other way was that `high-ev-covers-all` **passed** — and it clicks the same button,
so the button demonstrably works on a fresh page. **A check that passes is evidence
about the checks that fail**, and a suite is a set of controls for itself if you read
it that way.

The confirming discipline: I changed only the fence, never the tool source, and re-ran.
13 of 13. That is what establishes the fence rather than the tool as the defect —
had I "fixed" both at once, I would have learned nothing and published a false cause.

**A quieter finding underneath it.** `zero-ev-breaks-at-top` had **never executed**. It
died at its gate click on the only run that ever happened, so the tool's behaviour at
zero enterprise value was unverified for the entire period I was describing the tool as
working. A test that has only ever failed at step 1 has told you nothing about steps 2
onward, and a suite reporting "2 failed" invites you to read the other 9 as coverage.

---

## 2026-07-28 — I filed a table of eleven measured contrast ratios, and the measurement could not have worked

**The claim.** `bugs_open/122`: "generated stylesheets fail WCAG AA on four live sites",
with a table of eleven sites and two ratios each, presented as measurement. It named a
cause — the webdesign agent ignoring its own `a { color: var(--color-accent) }`
instruction — and ranked fix candidates against it.

**What was wrong.** I measured `--color-primary` against each site's background, because
that is what oufe's CSS used for links. **It is not the link colour anywhere else.**
dartsonline's `--color-primary` is `#111520` and its background is `#0D1019`, so the
"1.11 FAIL" was the background compared against itself. No rendered text was involved in
any row of that table.

The deeper error is that the question is not answerable from a stylesheet at all. Which
`a` rule wins needs cascade resolution; the contrast also depends on ancestors, alpha
and gradients. I wrote a regex for a question that requires a renderer, and then wrote
the output up as a table, which is the format that reads most like fact.

**The check I skipped:** confirm the thing you are measuring is the thing that gets
rendered — on *each* subject, not on the one that gave you the idea. One `grep` of any
other site's `a {}` rule would have shown a different variable and killed the
generalisation in seconds. (Tally row "check the column actually means what you are
measuring", incremented; and a new row for measuring a rendered property outside the
renderer.)

**What re-measuring properly found — the bug was real, and worse, and elsewhere.** In a
browser, with computed style and an alpha-composited backdrop:

- The shared header chrome hardcodes `color: white` on a button whose background is the
  site's **accent**, so passing is luck. **Five of six measured sites fail.** No palette
  change can fix it, because the text colour is not derived from the palette. It is also
  not in `styles.css` at all — it lives in stored chrome, so my entire class of
  investigation was looking in the wrong file.
- **robot-hands.com's primary call-to-action is white text on a white button.** Not low
  contrast: invisible. A live commercial site has been shipping a blank rectangle where
  its main CTA should be, and the audit I filed could not see it because it only looked
  at link colour.

**Three false positives on the way, each caught by screenshotting the element.** Comparing
against the page background rather than the element's; treating a gradient header as
transparent and falling through to the body (a blue header with white nav "measured" 1.05
against near-white); and treating `rgba(255,255,255,0.1)` as opaque white (a readable
white-on-purple button "measured" 1.00). **Every one looked like a serious live defect and
every one was disproved by looking at the page.** They are now guards in
`cmd/contrastscan`, each commented with the case that motivated it.

**The thing worth carrying.** For live public sites, an audit that over-reports is worse
than no audit: its findings get "fixed" into real regressions, and people stop reading it.
So the discipline is not "measure more", it is **screenshot anything you are about to
report as broken on a page a user can visit**. Nobody had complained about any of the
three false positives, and that silence was information I nearly overrode with a number.

**Distinct from the same day's acceptance-fence entry** above: that was a wrong test
becoming a specification for an agent. This one is a wrong *instrument* becoming a table
in a bug file — which propagates to every thread that reads it, with no agent involved at
all. The fence failure was loud. This one was silent, and would have stayed silent, because
a table of ratios is exactly the shape people do not re-derive.

---

## 2026-07-28 — "the Afternic minimum offer is 0, so the anti-lowball floor is absent" — I read a number out of a misaligned paste

**The claim:** that relojistas' domain listing had **no minimum-offer floor**, and that the
locked design's anti-lowball protection was therefore missing. I wrote it into **five
documents and a DB row** — including another workstream's NOTES, the fleet restart list, and
the Spanish copy spec — each time framed as a decision the owner needed to make.

**It is false. The floor is $12,000.** The owner corrected it.

**Where the 0 came from:** he had pasted his Afternic dashboard as plain text. The header row
listed ~11 columns (`… Minimum Offer · Sale Lander · Views · Leads · 30-day Searches …`) and
the value row carried **two**:

```
relojistas.com
Listed
0
0
```

I mapped the first `0` onto `Minimum Offer` because it was the first numeric column in the
header. **There was no column mapping to read** — those zeros were almost certainly Views and
Leads on a listing nobody has visited yet.

**What caught it:** the owner, four exchanges later. Nothing in my own process would have —
I never queried it again, because I had written it down as settled.

**The cheap check that would have caught it:** **quote the label back.** *"Reading your paste
as Minimum Offer = 0 — is that right?"* One clause. Or simply `[UNVERIFIED]` it.

**And the part that stings, because I nearly got it right:** in the *same sentence* I marked
the listing's STATUS as *"the owner's dashboard reading, NOT independently verified"* — and
then treated a **number from that same paste** as fact. I had the doubt and applied it to one
field. *A source is trusted or it isn't; you cannot mark one value from a paste unverified and
promote its neighbour.*

**Cost:** low to repair, but it propagated furthest of anything I got wrong this week —
because a *false problem* invites work. It sat in another workstream's NOTES as an open
design question ("is a floor part of what `for_sale_requested` asserts?"), and someone could
have spent a session on protection that was in place the whole time. **A wrong fact wastes one
correction; a wrong problem wastes a thread.**

Family: no-column-mapping-in-a-pasted-table, i-doubted-one-field-and-promoted-its-neighbour,
a-false-problem-propagates-further-than-a-false-fact.

---

## 2026-07-28 — the hand-maintained list, three times in one week, by me, in one file

**The claim.** Two separate lists in `write_experience_pattern_action.go`, both written as Go
literals, both believed complete:

```go
// which documents may contain {{binding.x}} placeholders
ValidateExperienceCriteria(tmpl, schema,
    entry["contract"], entry["states"], entry["data_contract"],
    entry["degraded_states"], entry["entry_points"])
```

**What was true.** The harvested entries put placeholders in `section_types` and
`destination_roles` too. Both were missing from the list, so bindings used in them were reported
**declared but never used** — and two real entries were refused for a defect they did not have.
The refusal message was confident, specific, and wrong.

**Why this one is worth a row rather than a shrug.** It is the *third* instance of the same
mistake in the same week, and I had already fixed the other two **by name**:

1. `experiencePatternJSONFields` — a jsonb column absent from the list is silently never written.
   A council seat raised it; I closed it with a test that re-derives the columns from the
   migrations.
2. `experiencePatternContractFields` — a clause-bearing field absent from the list means a changed
   clause keeps its approval. Three seats raised it; I closed it by making every column require a
   classification, so an unclassified one fails the build. I wrote, in that commit message, that a
   hand-maintained list *"drifts within one change"*.
3. **Then I left a third hand-maintained list untouched in the same file**, eighty lines above the
   fix for the second one.

The first two were closed because reviewers pointed at them. The third had no reviewer, and I did
not go looking — **I fixed the instances I was shown rather than the class I had just named.**
That is the actual failure: not the missing entry, but treating "list drifts" as a fact about two
specific lists instead of a fact about lists.

**The cheap check.** After fixing a class of defect, grep your own change for the same shape
before committing. In this file that was one command:

```bash
grep -nE '= \[\]string\{|entry\["[a-z_]+"\], entry\[' write_experience_pattern_action.go
```

**The fix.** No list. Pass every field except the two that cannot contain placeholders (the schema
and the template itself). There is no reason to enumerate documents-that-might-contain-placeholders
when "everything else" is both correct and shorter.

**Cost.** Bounded and self-revealing: it surfaced within minutes because the entries were being
loaded for real. The exposure is what it would have cost later — a refusal message that names the
wrong cause sends the next person to fix an entry that was already correct.

**Tally.** *After fixing a class, grep your own diff for the same shape* — a new row. Distinct
from "test against the artefact" (07-27, same file, same day's work): that one is about
manufacturing evidence, this one is about **fixing an instance and filing the class as done**.

Family: the-list-that-drifts, fixed-the-instance-not-the-class,
the-lesson-applied-to-everything-except-the-file-it-was-written-in.

## 2026-07-28 — I wrote "a null result is not a finding" last night, then made the mirror-image error three times before breakfast

Bug-sweep thread, `bugs_open/087`'s acceptance test. Twelve hours earlier I filed
the entry above about three zeros that were instrument failures, and added the
rule to 016b §9 and to memory. This morning I made three fresh errors in one
forty-minute exercise. **None of them was a zero.** They were the opposite, and
my own entry had named the reason and I did not act on it.

**Error 1 — the one that misinformed the owner.** Choosing a rebuild target, I
tested whether the page was live by guessing its URL from its name:

```
https://finetuning.uk/ai-agent-roi-estimator.html   -> 404
https://finetuning.uk/ai-agent-roi-estimator/       -> 404
```

and reported *"the page 404s on both URL shapes — nothing live is at risk"*. The
owner then chose that target **partly on that assurance**. The page was live the
whole time at `/tools/ai-agent-roi-estimator.html`, serving 35,129 bytes — and
`pages.url` held that exact string, in a row I had already SELECTed and printed
to my own terminal minutes earlier. I had the right answer on screen and tested a
guess instead.

**Error 2 — a poller that reported a live run as finished.** To find the parent
orchestration I wrote `WHERE correlation_id = … ORDER BY created_at DESC LIMIT 1`.
That returns the **newest** row under the correlation, which in a spawning
workflow is a grandchild. It came back `COMPLETED @ complete`, I printed
`=== PARENT TERMINAL ===`, and reported the run over. The parent was still
running and went on to fail three minutes later at a different step. The correct
predicate — `parent_orchestration_id IS NULL` — was one column away.

**Error 3 — a target that could never pass.** I checked blast radius, deployment
status, section shape and workstream ownership before choosing the page, and did
not check `rebuild_policy`. It was `owned`, so `save_page_sections` was always
going to refuse, and the acceptance test could not have completed on that page
whatever the fix did.

### Why the fresh rule did not catch any of them

Last night's entry ends: *"Non-zero results are self-validating (the instrument
demonstrably works), so we check positives casually and negatives not at all."*

I fixed the second half and left the first. Every error this morning is a
**positive result read as the thing I wanted**: a 404 is a real HTTP response, a
row came back from the database, the target satisfied four genuine checks. Each
returned *something*, so each felt verified — and none of them answered the
question I actually had.

**The sharper statement, which the earlier entry only half-made:** the danger is
not the shape of the result. It is that **a check answers the question you
encoded, and you then read it as answering the question you meant.** `curl
<guessed-url>` answers "is there a page at this string I made up". `ORDER BY
created_at DESC LIMIT 1` answers "what is the newest row". Both answered
correctly. Neither was asked what I thought I was asking.

### The cheap checks

- **Never synthesise an identifier the database already stores.** URL, path,
  filename, topic, key: if a column holds it, SELECT it. I had already printed
  `url` and typed a guess anyway — so the rule is not "look it up", it is *"do
  not type a value you could have pasted"*.
- **Naming a row is a query about relationships, not recency.** Parent, head,
  canonical, latest-of-its-kind — express the relationship (`parent_orchestration_id
  IS NULL`), never a proxy for it (`ORDER BY created_at`). A proxy is right until
  the shape of the data changes, and a spawning workflow changes it every run.
- **Before an acceptance test, list what would make it fail for reasons unrelated
  to the fix**, and check each. I checked four such things and missed the fifth;
  ten seconds on `SELECT rebuild_policy` would have found it. A test that cannot
  pass is worse than no test — it burns the setup cost and returns an ambiguous
  answer.
- **On writing a rule, ask which half you have actually adopted.** I generalised
  correctly and then applied it to one shape only. The entry was in the file, in
  016b, and in memory; recall was not the problem.

Family: interceptor-existed-and-was-not-used, the-check-answers-the-question-you-
encoded, proxy-for-a-relationship, rule-adopted-in-half.

---

## 2026-07-28 — "there is no chart renderer", said three times, while two of them were live

**The claim.** That the platform has no way to draw a chart, so the owner's request for
graphs could not be met without building one. I put it in the oufe handoff as a warning
("**Before starting 5, know this: there is no chart renderer**"), repeated it in two
summaries to the owner, and used it to justify calling graphs a blocked item needing a
decision.

**It was wrong, and the owner corrected it** ("please check that there is no chart
renderer as there was some work done on that").

**What actually exists:**

- **`evidence-chart`** — a live, active section component. Horizontal bars drawn in **CSS**
  custom properties, no SVG and no dependency. Every plotted point resolves through a
  `fact_id` into the evidence register; the denominator can itself be a `max_fact_id`; each
  row renders `verified <date>` beside the value and the figure carries a `source_note`.
  **A chart point structurally cannot carry its own number.** That is the exact doctrine I
  was describing as unbuilt, already built and running on a live site.
- **`renderBarChartSVG` / `renderHeadroomChart`** in
  `platform/orchestration/actions/report_charts.go` — dependency-free inline SVG.
- **`features_open/023`** — the designed rule that infographic prompts must be assembled
  from `evidence_base`, plus the R4 scoping boundary: *generated images explain,
  code-rendered SVG states*, for values that must be exact, selectable or translatable.

**Why I got it wrong, and it is not the obvious reason.** I did not invent the claim: I
inherited it from a handoff I wrote earlier in this workstream, and the handoff was
honest about its own basis (`go-echarts` is not in `go.mod`, and `report_charts.go` is
unexported and bound to one report page). Both of those facts are still true. **The error
was the conclusion drawn from them** — "no charting library and one narrow helper" became
"no chart renderer", and the missing step was looking anywhere other than Go source.
`evidence-chart` is a row in `content_components`; no amount of grepping `go.mod` or
`platform/` will ever surface it. The capability lived in the database.

**The check I skipped, twice over.** First: `grep for the capability before asserting it
does not exist` — which for this platform means the **component library and the seeds**,
not just the code, because most capability here is data. Second, and the one I want as its
own row: **I never read `features_open/` at all.** Two of the questions I was treating as
open and unanswered — how infographic figures stay sourced (023), and how a topic gets
decomposed into a packaged feature (001, raised by the owner himself on 2026-07-19) — were
sitting there designed, with the tradeoffs already worked through. A feature file is
invisible to a code grep, to a `content_components` query, and to the "grep both bug dirs"
rule, so nothing I habitually do would have found it.

**What is genuinely missing**, stated properly this time: a **time-series** renderer. Both
existing renderers compare magnitudes and neither has a temporal axis. And underneath that
is a substrate gap I would have missed if I had stopped at "we have charts": an
`evidence_base` fact holds **one value plus provenance dates** (`accessed`, `published`,
`verified_at`, `staleness_days`). None of those is *the date the value applies to*. A
historical graph needs an observation series with an `as_of` per point, so the gap is a
schema question before it is a rendering question.

**The shape to carry.** The dangerous inherited claim is not the false one — it is the one
whose *evidence* is true and whose *scope* is wrong. "go-echarts is absent" was correct and
checkable, and it made the surrounding sentence feel checked. **A warning I wrote myself,
in a handoff, in bold, was the thing I never re-derived** — precisely because it looked
like the output of an investigation rather than an input to one.

---

## 2026-07-28 — `node --check` printed SYNTAX OK on a machine with no node

**The claim.** Having patched a live component's JavaScript, I ran
`node --check file.js 2>&1 | head -3 && echo "  SYNTAX OK"` and reported the file
as syntactically valid before delivering it to production.

**What was true.** `node` is not installed here. The shell printed
`node: command not found`, `head` exited 0 because *it* succeeded, the `&&` fired,
and `SYNTAX OK` was printed for a file that was never parsed by anything.

**What caught it.** Reading my own output. The "command not found" line was right
there above the reassurance.

**The cheap check that would have.** `command -v node || echo "NO CHECKER"`.
**If a verification tool might be absent, assert it exists before trusting its
silence.** A missing tool and a passing check produce the same exit status.

**Cost.** None realised — the file was afterwards driven in a real browser, which
is a stronger check and found it fine. But it was luck that the stronger check
existed: had I relied on the `node` line, an unparsed file would have gone to a
page a visitor was actively using.

**Same session, the one that nearly shipped:** `(el.objectives || []).map(...)`.
`querySelectorAll` returns a **NodeList**, which has `forEach` but **no `map`** —
it would have thrown on the first save. Caught by re-reading the patch, not by any
tool. *Check: NodeList is not an Array; slice it before using Array methods.*

**And a third, caught by measurement:** I wrote a CSS override at specificity
(0,1,1) against a component rule at (0,2,0) and would have called it done. Class
count is compared before element count, so mine lost silently — the width simply
did not change. *Check: after writing an override, MEASURE the property, never
assume the rule applied.*

**The tally now stands at four "checks that could not fail" in two days** —
a file-gated linter, a source-string grep of a binary, a wrong binary path, and
now a checker that was not installed. Each produced output identical to success.
**An assertion that cannot fail is not an assertion**, and the failure mode is
always the same shape: something returned zero, or nothing, or exit 0, and I read
that as a result rather than asking whether the instrument was connected.

Family: a-clean-result-and-an-unrun-check-are-identical, vacuous-detector,
the-verification-command-is-code-too, missing-tool-reads-as-passing-check.

---

## 2026-07-28 — I shipped a TEMPLATE to production and called it a page. The owner found it.

**The claim.** Having changed the gauntlet's type scale, I verified brace balance,
line count, selector presence, computed font sizes on both viewports, and no
horizontal overflow — all green — and reported the change live and correct.

**What was true.** The page was serving **32 raw `{{.placeholder}}` strings**. The
owner saw `{{.hero_title_plain}} {{.hero_title_accent}}` where the title should be
and sent it back with "(oops)".

**The mechanism.** `RUNBOOK §10` — *which I wrote three days earlier* — says to
write the same string to both `content_components.html_template` and
`page_components.rendered_html`. That rule is correct **only for a component with
no template variables**, which is what I derived it from (the Arena: 0 vars, empty
`content_data`). The gauntlet has **27 placeholders and a populated
`content_data`**; `rendered_html` is the template *with those substituted*.
Copying the template over it replaced the page with its own unrendered source.

**What caught it.** The owner, on the live site. Nothing I ran would ever have.

**The cheap check that would have.** One command:
`curl -s "<url>?cb=$(date +%s)" | grep -c '{{\.'` → expect 0. It is now RUNBOOK
§11, with the discriminating query (count `{{.vars}}` before copying) and the
substitution recipe.

**The real lesson, which is bigger than the rule.** **Every check I ran was on the
change I had MADE.** Font sizes — the thing I changed. Braces and line count — the
integrity of the thing I changed. Overflow — the consequence of the thing I
changed. Not one of them asked whether the page still did its own job. **A diff
proves what you changed; it cannot tell you what you destroyed.** After any
delivery, assert something the change was NEVER ABOUT: that the page still renders
its own content, still serves 200, still completes its journey.

**Aggravating, and worth stating plainly:** I had spent the previous two days
logging four separate "checks that could not fail" and writing that *an assertion
that cannot fail is not an assertion*. This is the inverse and it is worse — a
suite of assertions that all COULD fail, all passed, and all pointed the same
wrong way. Comprehensiveness within the blast radius of your own edit is not
coverage.

**Also:** I generalised a runbook rule from a single instance (the Arena) and did
not mark it as such. The rule was true of the case it came from and false of the
next one. *Check: when writing a rule from one example, state the property that
made the example work — here, "the component has no template variables" — so the
next reader can tell whether they are in scope.*

Family: verified-only-what-i-touched, a-rule-generalised-from-one-instance,
the-owner-as-the-last-line-of-defence.

## 2026-07-28 — I flipped an env flag for code that was not in the binary, five minutes after writing the rule against exactly that

The claim: setting `CHASSIS_RESPONSES_START_AT=latest` on the chassis would
stop the response-topic replay. The reality: the running image was
v1.0.1184, built BEFORE the flag's code existed (commit `f4d24252f` came
later); an unknown env key is silently ignored, so the "fix" was a no-op that
cost one more pod restart — which itself restarted the 2–3 hour replay from
zero. Caught within minutes because the verification I ran next
(`RESPONSES_START_AT_LATEST` log line + fresh group LAG) contradicted the
flip; the pod-grep of the binary then showed 0 hits for the symbol.

What made this one embarrassing rather than merely wrong: the pre-flip
pod-grep gate was already written down TWICE by me that same hour — in the
CS-2 audit artifact ("before CHASSIS_INTAKE_MODE or CHASSIS_DB_MAX_OPEN_CONNS
is set non-default anywhere, grep the RUNNING pod's binary…") and in the
CS-3a council submission itself — and a council seat (debug_historian) had
flagged precisely this failure shape for the OTHER knob in the same change.
I applied the gate to the flags I'd been warned about and skipped it for the
one I'd just invented. A checklist you exempt your newest change from is not
a checklist; the newest change is the one that has never been through it.

The cheap check that would have caught it, and now has a tally mark:
`kubectl exec <pod> -- strings /app/agent-chassis | grep -c <symbol my
change created>` BEFORE any `set env` naming that change. Also the standing
one-liner: an env flip is a DEPLOY of config against a binary — the binary's
contents are part of the precondition, not an assumption.

Family: grep-the-config-key-before-calling-it-a-win,
verified-only-what-i-touched, the-rule-exempted-its-own-author.

---

## 2026-07-28 (later) — my contrast checker reported the page clean, and the owner sent back a screenshot of unreadable text

**The claim.** That the Thames Water page passed contrast. `cmd/contrastscan` — the tool
I had built that same day, and written up as the fix for `bugs_open/122` — reported
"3 measurable pair(s), all pass (worst 4.78)". I quoted that in a summary as evidence
the page was fine.

**What the owner saw.** A screenshot with the section eyebrow invisible against the
page, and the chart's title and caption invisible against a white card. Re-measured
with a wide selector: **1.23**, **1.29**, **2.85**.

**The defect was the tool's DEFAULT SCOPE.** It measured
`a, button, .btn, [role=button]` — interactive elements only. Links and buttons are a
small fraction of the text on a page, and nothing about a contrast defect confines
itself to them. Headings, labels, captions and chart text were never looked at, and
the tool reported that as *passing* rather than as *not checked*.

**This is the mirror of the lesson I wrote in this file earlier the same day.** That
entry said an over-reporting audit is worse than none, because its findings get
"fixed" into real regressions. The inverse is worse still and I walked straight into
it: **an under-scoped audit produces false green, and false green is trusted.** Noise
gets ignored; a clean bill of health gets quoted. I quoted it.

**Both defects were ones I had already written down.** `bugs_open/122` records that
oufe's `--color-primary` is identical to its surface colour, and describes
`--color-card-bg: #ffffff` as *"latent rather than live — light body text on a white
card is the same invisibility bug waiting for a news page to be added."* I then
shipped a component that renders a card, on that palette, and did not connect it.
**A hazard I documented did not become a hazard I checked for.**

**The check that would have caught it, and it is embarrassingly cheap:** look at the
page. One screenshot of the section I had just added. I had a working screenshot
harness open — I had used it three times that day to disprove false positives — and
did not point it at my own work, because the numbers were green.

**And fixing it broke something the numbers still could not see.** Making the card
dark fixed the text and made two of the three bars invisible, because the default bar
fill is `var(--color-primary)`, which was now exactly the card colour. `contrastscan`
cannot see that either: bars carry no text, and the probe skips elements with no text
content. This is concept VIZ-011 in the register — *chart furniture is a graphical
object needing the 3.0 non-text threshold* — written by me the same day, and not
implemented in the tool that was supposed to enforce it. Caught only by screenshotting
the result of my own fix.

**The compounding shape worth carrying:** a measuring instrument you built yourself is
the hardest thing to distrust, because its scope was chosen by the same reasoning that
chose what to build. When it agrees with you, it may only be agreeing with your blind
spot. **Verify a fix by looking at the artefact, not by re-running the check that
passed before you made it.**

---

## 2026-07-28 — "PROVEN END TO END" on a route that could not reach the failing step (079 closure)

**The claim:** `bugs_open/079` closed as FIXED, "PROVEN END TO END on the deployed
binary" — the link repair rewrites/unlinks dead hrefs "in `clean_html` — the string
`save_sections` persists."

**What was true:** the repair action works, logs durably, and is deployed. **What was
false:** `save_sections` never reads `clean_html` on the primary build plan — the
structured `sections_metadata` path wins whenever metadata exists, which
`require_sections_metadata: true` guarantees. The repair output is discarded on every
natural build. Caught 2026-07-28 by a routine link crawl on fundamentallyai finding all
9 "repaired" targets live and 404ing, 400ms after the repair log row.

**Why the proof missed it:** the induction route (`content-reviewer`, `html_field`
repointed) had no `save_sections` step — chosen because no natural induction was
available at the time. The proof verified the ACTION's return map on a synthetic route,
and the claim quietly widened from "the action repairs" to "the page ships repaired."

**The cheap check that would have caught it:** one SELECT of the saved
`page_components.rendered_html` after the first natural build that logged a repair —
the repaired href must be absent from the PERSISTED row. Same family as "writes the
field ≠ reads the field" and 016b §"pod-grep proves the binary, not the path" — this
adds the third rung: on the path, ran, and the next step discarded its output.

---

## 2026-07-28 — idea.uk: I promoted "not one of our IPs" into "our first customer"

**The call.** An order appeared that I had not made. I checked the evidence I had —
an IP that was not the owner's, an Android user-agent we had not seen before, a name
and Gmail address we did not recognise — and wrote into `RUNNING_NOTES §X.27`, in
bold: *"The first genuine external prospect this product has ever had."* Then said the
same thing to the owner as the headline finding of the session.

**It was a test.** The owner corrected it in one line: he will not pay, he was a test.
idea.uk has still never had a genuine external buyer.

**Why it was wrong, precisely.** Every piece of evidence I had was consistent with the
claim and none of it *entailed* it. **A device and an address we do not recognise do
not make a stranger** — they make an origin we cannot attribute. I had an inference
about identity and I wrote a conclusion about the market, which is a much bigger claim
resting on the same thin fact. The give-away was in my own sentence: "an IP that is
not the owner's" is a statement about what something *is not*, and I used it to
establish what it *is*.

**What caught it.** The owner, immediately, because he happened to read the claim. If
he had not, it would have sat in the permanent record as the moment this product
proved market demand — and every later decision about where to spend effort would have
been taken against a fact that was never true.

**The cheap check:** `[INFERRED]`. One marker, at the moment of writing, on a claim I
could not verify from the data in front of me: *"appears not to originate from one of
our own IPs"*. That phrasing invites the correction; the bold assertion does not. The
rule this workstream already had — *mark the unverified ones too* — exists exactly for
this, and I applied it carefully to the cost figure in the same session and not at all
to the sentence I led with.

**The asymmetry worth naming.** I was scrupulous about the number ($1.23, with a
`[FLOOR]` caveat the day before) and careless about the identity, because the number
*felt* like the claim under scrutiny. **The claim that gets the marker should be the
one that would change decisions if wrong, not the one that looks quantitative.** "We
have a customer" redirects a whole workstream; "$1.23" does not.

**What survived, and why it matters that it did:** the cost measurement and the
copy-fix proof both rest on the *artefact* — a real six-call run and a real rendered
report — so who submitted it is irrelevant to either. Findings anchored to an artefact
survive a wrong attribution; findings anchored to a person do not. That is a reason to
prefer the former when both are available.

**Tally.** *An unrecognised origin is not an identity* — new. *Mark the claim that
would change decisions, not the one that looks like a statistic* — new, and the sharper
form of the existing mark-the-unverified rule. Third session running in which the
correction came from asking whether the evidence could have come out any other way.

Family: figure-carried-forward-from-prose, a-green-result-from-an-input-that-cannot-fail,
an-unrecognised-origin-is-not-an-identity.

---

## 2026-07-28 — "seed 220's snapshots do not exist", from a diff whose baseline was empty

**Session:** the 086 per-handler audit (picked up after an owner machine crash).

**The claim, said out loud before it was checked.** Auditing the ten `error_step`
handlers seed 220 disabled, I wanted to know whether any had drifted since. I diffed the
live definitions against the seed's own pre-update snapshots and got **0 rows — nothing
lost**. Then, checking the snapshots themselves, I found **one row in the whole table,
for none of the seven agents**, and reported to the owner: *"Positive control fails —
exactly one snapshot row exists in the whole table."* The implication I was one step from
committing was that seed 220's safety net had never been created, i.e. the ten renames
were unrevertable.

**It was false.** `snapshot_agent` is **overloaded**. The one-arg form writes into
`agent_definitions` with `is_snapshot = true`; the two-arg form — `snapshot_agent(type,
reason)`, the one seed 220 actually calls — writes into a **different table**,
`agent_definitions_backup`. All seven snapshots are there, `snapshot_reason LIKE '220_%'`,
taken `2026-07-26 18:32:26.229Z`. The safety net was intact the whole time and revert was
never in doubt.

**What caught it:** reading `pg_get_functiondef` for the function before believing the
query about its output. Cost: two wasted queries and one wrong sentence to the owner.

**The cheaper check, and the one with wider reach.** The *first* query — the diff that
said "nothing lost" — was already worthless, and worthless in the more dangerous
direction: **its snapshot CTE was empty, so the EXCEPT returned nothing and read as
reassurance.** An empty baseline answers every question with "all clear". Had I not
gone on to look at the snapshots for an unrelated reason, "no handlers have drifted"
would have gone into the audit as a finding.

> **The rule: an `EXCEPT` / `NOT EXISTS` / anti-join diff must COUNT its baseline before
> it is allowed to report a null result.** `SELECT count(*) FROM <baseline>` on its own
> line, in the same query. A diff with no baseline is not a passing check, it is an
> absent one — and it is indistinguishable from a passing one in the output.

**Family resemblance.** This is [[check-answers-the-question-you-encoded]] wearing new
clothes: there was no filter to notice being wrong, and the query was well-formed and
returned a clean, plausible, non-empty-looking answer. It is also the third distinct
instance on this fleet of *a green result from an input that cannot fail* — the same
shape as a dead control, a `LIKE` guard on a blob, and an acceptance test whose
distinguishing input cannot occur. The generalisation those three want is: **any check
that can only emit "fine" should be made to emit "fine" for a case you know is broken,
once, before you trust it.**

**Tally.** *Count the baseline before believing an anti-join* — new. *Two functions of
the same name can write to different tables — read the definition, not the name* — new.
*A green result from an input that cannot fail* — fourth occurrence, and the one that
keeps earning its place.

Family: a-green-result-from-an-input-that-cannot-fail, check-answers-the-question-you-encoded,
verify-the-failing-branch.

## 2026-07-28 — the handoff said the 50 were invisible to the dashboard; they were on its default screen all along

**The claim.** `review_queue_drain/HANDOFF_2026-07-28_…` §4.2: the 50 items needing a
human answer *"are `status='detected'` with no handler, so the dashboard cannot see
them at all"* — so the next session's task was to build a way to surface them, and the
owner was told decision A ("promote the 50") had an execution step.

**Where it came from.** The classification (50 / 186 / 5) was computed in
`bugs_open/083_…_detected_findings_never_reach_a_handler.md` over "non-terminal items
with no handler" — a population spanning several statuses. The bug file is about the
`detected` queue, so when the handoff paraphrased its own classification, the
population inherited the bug's status. Nobody had run `GROUP BY status` over the three
item types; the 083 file itself never makes the claim. The error is purely in the
paraphrase — and it survived because it was written in the same confident voice as the
measured numbers around it, in a handoff whose own §5 warns *"before proposing a
route, measure the destination"*.

**What it nearly cost.** A session building a filter/promotion mechanism for items
that are already at `needs_human_review`, in the build pipeline, on the dashboard's
default screen, with a working per-item form. The dashboard session's first
measurement killed it: 42 + 6 + 2, all `needs_human_review`, one query.

**The cheap check.** Before asserting *where* a population sits — a status, a queue, a
directory — group the population by that column and paste the result next to the
claim. If the sentence names a status and no query in the document groups by status,
the sentence is a guess. This is [[check-answers-the-question-you-encoded]]'s sibling:
the numbers in the document were all real, they were just answers to a different
question than the sentence they decorated.

**Tally.** *A property asserted about a population that no query grouped by* — new,
but it is the narrow-filter family's third face. *Confident voice indistinguishable
from measured voice* — recurring; the `[INFERRED]` marker existed for exactly this
sentence and was not applied, again.

Family: narrow-filter-defines-the-conclusion, check-answers-the-question-you-encoded,
writes-the-field-is-not-reads-the-field.

---

## 2026-07-28 — "Four of nine patterns do not resolve" — a join keyed on the wrong column (brochure §4d)

**The claim** (brochure HANDOFF_2026-07-28 §4d, stated as measured with a printed
truth table): *"Four of nine [experience_patterns section_types] do not resolve, and
they are exactly four of the five components this workstream built"* — presented as
evidence that the register was written with approximate names and needed fixing.

**What was actually true.** All four resolve. `content_components` names a component
TWO ways — `function` and `section_type` — and the register's entries name the
`section_type` (which is the selector key: `idx_cc_selector`). The morning's join
tested `function` only. The truth table was real; it answered "which entries match
`function`", not "which entries resolve".

**What caught it.** `sql_for_agents/256` (the join check §4d itself asked for), written
dual-column after a `\d content_components` showed the second name column: zero rows
missing. The correction is in the brochure NOTES and the 07-28b handoff, and step 2 of
§4d ("fix the four names") died with it — a write to another workstream's actively
maintained rows that would have "fixed" four correct values.

**The cheap check.** `\d` the join TARGET before declaring a foreign key broken. If a
reconciliation claim names one column and the target table has a second name-shaped
column, run the join both ways before printing a truth table. Schema-first is already
the SQL rule; this extends it to join keys: the table you are joining TO is part of
the question you are encoding.

**Tally.** *check-answers-the-question-you-encoded* — recurring, and this instance
nearly caused writes: the "fix" for the four names would have corrupted a live
register. Second face this week of: a printed measurement lends false weight to a
mis-keyed question.

Family: check-answers-the-question-you-encoded, narrow-filter-defines-the-conclusion.

---

## 2026-07-28 — a read-only probe measured a path production does not use, and the delta was carried as achievable

**The claim** (`bugs_open/101` §Consequences 1, stated as measured, n=100,
"deterministic sample"): company-number extraction from the home page alone yields
**22/100**; reading legal/terms pages as the config implies yields **30/100** — the
8-hit delta presented as what implementing `follow_links`/`max_pages` would deliver.

**What was actually true.** The probe **fetched raw HTML**. Production fetches through
`FirecrawlScrapingProvider.Scrape`, which could not express `only_main_content: false`
and so received Firecrawl's default — `onlyMainContent: true`, which strips headers,
navs and **footers**. Company registration numbers live in footers. So the probe and
production were not fetching the same bytes, and the 30% was never reachable by adding
page fetches: the pages would have arrived with the very region the numbers live in
removed. `[UNVERIFIED]` how much of the 22→30 gap this accounts for — not measured, and
should not be asserted either way.

**What caught it.** The bug file's own `[UNSETTLED]` box, which refused to let candidate
2 proceed until someone answered "does production strip footers?". Nobody had. Reading
the provider settled it in one pass — and the answer was a third bug, in shared adapter
code every scrape on the fleet goes through (three live steps ask for `false` and have
been getting the opposite).

**To its credit:** the box is why this cost one read instead of an implementation. The
author flagged the uncertainty, marked it `[UNSETTLED]`, and wrote "READ THIS BEFORE
IMPLEMENTING candidate 2 — it may not be sufficient". That is the marker discipline
working exactly as intended.

**The cheap check.** **Measure through the path production uses, or state the path you
measured as part of the number.** A probe that constructs its own HTTP call is measuring
a *sibling* of the production path, not the production path — and the difference is
invisible in the result, because both return plausible HTML. One
`grep -n "only_main_content" <the provider>` before quoting the delta would have shown
the two paths disagreeing inside a single file.

**Tally.** *a-green-result-from-the-wrong-path* — same family as
`deployment-is-not-spawned` (016b §9) and `check-answers-the-question-you-encoded`: the
measurement was real and answered a question about a system we do not run. Second face
this week of: an honest number, obtained through a route the fleet does not take.

Family: check-answers-the-question-you-encoded, verify-the-failing-branch,
omission-is-an-instruction.

---

## 2026-07-28 — a `[VERIFIED]` marker on a claim read out of a *print statement*

*(Recorded by the `bugfix_124_double_dispatch` thread. The claim is not mine; the
tally is the fleet's, and this one is worth the row because the marker discipline
did not merely fail to help — it actively vouched for the error.)*

**Asserted**, in `bugs_open/124` §Mechanism, item 1, tagged **`[VERIFIED]`**: *"Nothing
marks a `needs_diagnosis` item complete on success — the 090 trigger still prints
'closing it by hand until a diagnose dispatch loop exists'. The loop now exists and
still does not close them, so a diagnosed item stays claimable."* Fix candidate 1
was then ranked first on the strength of it: *"Make a diagnosed item terminal…
This makes the duplicate unrepresentable."* The same belief is asserted in
`bugs_open/029` §6.

**What was actually true.** The live `diagnose-dispatch-loop` row has had a
`mark_complete` step (`complete_work_item`) all along, and it works — every
090-filed item sits at `complete` or `failed` with
`claimed_by='diagnose-dispatch-loop'`. Candidate 1 required no work whatsoever. The
real cause was elsewhere: the 090 script publishes its own dispatch **and** leaves
the row claimable, so the loop runs it a second time.

**What caught it.** Reading the live `agent_definitions` row before building
anything on the claim. One query:

```sql
SELECT jsonb_object_keys(default_config->'workflow'->'steps') FROM agent_definitions
WHERE type='diagnose-dispatch-loop' AND is_active AND COALESCE(is_snapshot,false)=false;
```

**The cheap check.** **A print statement is not a config row.** The evidence quoted
for the `[VERIFIED]` tag was *text the script emits to a terminal* — a sentence
written by a human in 2026-07-09 describing the world at that time. It is the same
class as a stale comment, and we already have a standing rule for its cousin ("the
seed is history, the live row is fact"). The rule needs saying about **prose inside
executable files** too: headers, banners, `echo` output and usage text are all
documentation, and documentation does not change when the system does.

Concretely: before tagging `[VERIFIED]`, ask *what did I read?* If the answer is a
string that a human wrote, the tag is not earned no matter how authoritative the
file. Read the thing the string describes.

**Tally.** *documentation-in-executable-clothing* — third face of the same family
this month, after `seed-sql-is-history-live-row-is-fact` and `grep-the-config-key-
before-calling-it-a-win`. Also a second, milder instance in the same file, caught
the same way: an orchestration cited as running "43 minutes after the diagnosis
finished" was created 91 seconds after intake — `SELECT created_at` on the id the
note already quoted. **Before repeating a claim about *when* something ran, select
its timestamp.**

**Worth saying plainly:** the filer marked the *other* mechanism `[UNVERIFIED]` and
was scrupulous about it, which is why that one cost nothing. The discipline works.
This row is about the one case it can't help with — where the check felt done.

Family: seed-sql-is-history-live-row-is-fact, grep-the-config-key-before-calling-it-a-win,
writes-the-field-is-not-reads-the-field, check-answers-the-question-you-encoded.

---

## 2026-07-28 — I reported "zero unknown config keys" as a result, when my own declaration is what made it zero

**The claim** (commit `2ebabf2ca` message, and concept-register entry SCR-004,
both stated as a measured outcome of the new audit): *"It now reports zero unknown
keys and 208 undeclared actions."*

**What was actually true.** The audit reports zero because **I declared the
offending keys as recognised**. `max_pages` and `follow_links` went into
`scrape_web`'s `ActionInputSpec.ConfigKeys`, which is what the detector consults —
so the two live steps that still describe a three-page crawl nothing performs
(`vet-practice-verifier/scrape_website`, `domain-research-classifier/scrape_site`)
now pass silently. The defect is unchanged; only its visibility to my own tool is.
It survives as a runtime warning at `buildScrapeConfig` time, which is strictly
weaker than the audit line it replaced.

**What caught it.** The council gate's `editquality` seat, round 1, correlation
`f4cf0aab` — an advisory objection on an edit it otherwise approved:
*"max_pages/follow_links are declared as valid config keys (so they now pass the
new validator silently) … it means the 22%->30% company-number gap … is not
actually closed by this plan, only explained."*

**Why this one is worth the entry.** I had *already* modelled this failure mode and
built a guard against it: `UnknownConfigKeys` returns a `checked` bool precisely so
"declared and clean" can never read as "never examined", and I wrote a test
asserting that distinction. Then I produced a **third** state I had not modelled —
*declared, but honoured only under a condition that does not hold here* — and it
reports as clean through the guard I designed. Building the discrimination did not
make me apply it to my own output.

**The cheap check.** **After adding any suppression, allow-list or declaration, re-run
the detector against the ORIGINAL failing case and confirm it still fires.** If
declaring something makes the detector quiet about it, you have not fixed the case —
you have exempted it, and the exemption needs its own state rather than sharing
"clean". One `./scripts/audit-config-keys.sh` run read with the question *"would
this still have found the bug I started from?"* answers it in seconds.

**Tally.** *a-quiet-result-reads-as-a-pass* — third face this week, and the first
where the person who built the anti-quiet mechanism is the one it caught. Related:
`narrow-filter-defines-the-conclusion`, `check-answers-the-question-you-encoded` —
same family, except here the filter was authored by the fix itself.

Family: check-answers-the-question-you-encoded, a-quiet-result-reads-as-a-pass,
verify-the-failing-branch.

---

## 2026-07-28 — a measurement, a population, and a total that matched. The mechanism was imaginary.

**The claim, one command from being filed fleet-wide.** Two `git commit`s timed
out at 120s. Theory: CLAUDE.md *mandates* pathspec commits → `git commit <path>`
takes files from the working tree and **ignores the index** → `git diff --cached`
is therefore empty → `check-secrets.sh` falls back to `git ls-files` → it scans
**every tracked file**. I measured the repo at **7,948 files** and one pattern
pass at **8.2 seconds**. Multiply by the pattern array and it lands almost exactly
on the 120s I had seen. Twice. I was drafting the bug.

**What was true.** `scripts/check-secrets.sh:29-34` scans `git ls-files` **only**
when passed `--all`. Otherwise it takes the staged set, and
`[[ -z "$files" ]] && exit 0` returns immediately. **There is no full-repo scan on
a normal commit.** The real cause was `index.lock` contention — five commits from
concurrent sessions inside five minutes — and the 120s was my own tool timeout,
not git's.

**What caught it.** Reading the script. Nothing else would have: every number in
the theory was real and independently correct.

**The check that would have.** I had already run `check-secrets.sh` standalone and
watched it finish in **0.1 seconds**. That result refuted the theory outright, and
I explained it away as "it took the cheap path" rather than letting it stop me.
*Check: when an early observation contradicts the theory you are building, that is
the refutation, not an anomaly. And read the code path before believing
arithmetic — **a quantitative fit is not a mechanism.***

**Why this one is worth the entry even though nothing shipped.** Every previous
entry in this ledger is a check that COULD NOT FAIL. This is the opposite failure
and the more seductive one: three independent, correct measurements assembled into
a causal story that did not exist. A wrong bug filed fleet-wide costs every thread
that then believes it, and this one would have carried real numbers — the most
persuasive possible form of being wrong.

**Same session, the smaller ones:**
- **A pod-grep through `strings` on a container that has no `strings`.** Returned
  0 for the marker AND 0 for `no_horizontal_overflow`, the check's own type name,
  which is impossible. Caught by the POSITIVE control; the negative control alone
  could never have distinguished "absent" from "instrument disconnected".
  CLAUDE.md's verify-against-the-pod recipe assumes `strings` — present on the
  chassis, **absent on browser-runner-adapter**.
- **I asserted what a check could not do without reading it.** The claim happened
  to be true when written, then I nearly RETRACTED it on reading a source comment
  that postdated my own bug and described the fix it had caused. *Date the source
  before retracting.*
- **I did arithmetic on a count my own grep had failed to produce** (`0` patterns,
  printed, then fed into a projection).

Family: quantitative-fit-is-not-a-mechanism, explained-away-the-refutation,
missing-tool-reads-as-passing-check, computed-with-a-broken-count.

---

## 2026-07-28 (later still) — I built a tool that already existed, because my prior-art grep filtered to the wrong language

**The claim.** That the platform had no browser-based contrast audit, so I built
`cmd/contrastscan` — and then wrote it into `bugs_open/122` as that bug's own fix
candidate 2, into the concept register as VIZ-010, and into a summary for the owner.

**`scripts/render_audit.py` had existed since the previous day** and does the same
job better. It renders every element (`body *`, not a selector list), walks up
through transparent ancestors compositing alpha, applies 4.5:1 and 3.0:1, flags a
backdrop it cannot resolve as `overImage` rather than guessing, and additionally
reports images that failed to load. Run against the page I had just fixed, it
agrees: `contrast=0 broken-img=0`.

**How I missed it.** I did search for prior art. The command was:

```
grep -rln "wcagContrastRatio\|contrastRatio" --include=*.go cmd/ scripts/
```

I even searched the right directory. **`--include=*.go` excluded it, because the
prior art is Python.** The filter came from my own assumption that platform tooling
is Go — reinforced by CLAUDE.md's "Go, not Python" convention, which is a rule about
what to *write*, not a description of what *exists*. A filter taken from an
assumption returns a confident, well-formed, empty answer, and an empty answer from
a search you designed looks like evidence of absence.

This is the third distinct form of the same error in one day, and the three are
worth seeing together, because each defeated the previous fix:
1. "there is no chart renderer" — searched Go source; the renderer was a **database
   row**;
2. "the contrast bug is the link colour" — searched the **stylesheet**; the answer
   needed a browser;
3. this one — searched the right place, with the right terms, filtered to the wrong
   **language**.

**The compounding cost.** `render_audit.py`'s own header names, in its opening
paragraphs, the two defect families I shipped to a live site that afternoon:
*"a palette slot the layout supplies a literal for (a #ffffff card_bg on a dark
site)"* and *"a token used in two roles (--color-primary as both a fill and a
foreground)"*. Those are exactly the white card and the invisible eyebrow the owner
screenshotted back to me. **Running the existing tool once would have found both
before he saw them.** I did not run it because I did not know it existed, and I did
not know because of a four-character flag.

**The cheap check:** when grepping for prior art, **do not filter by language**.
Filter by directory if you must, never by extension. The variant that works is
`grep -rln "<concept>" scripts/ cmd/ platform/` with no `--include` at all; the noise
is a few dozen lines and the cost of missing is a duplicate tool plus the defects it
would have caught.

**Resolution:** `cmd/contrastscan` deleted. Register VIZ-010 and bug 122 redirected
to `render_audit.py`. Nothing distinctive was lost — its one different behaviour
(refusing to score an unknown backdrop rather than flagging it) is a stricter
variant of `overImage`, and not worth a second tool.

**The thing worth carrying:** a search is only evidence within the world its filters
describe, and you wrote the filters. **When a prior-art search comes back empty,
re-run it with one fewer constraint before believing it** — the constraint you would
defend most confidently is the one most likely to be carrying your assumption.

---

## 2026-07-28 (evening) — a census with a filter is not a census, and the count fell because the *recommended fix* was applied

**The claim.** `bugs_closed/086`'s per-handler audit, that morning, could not explain
why the fleet's step-level `error_step` handler count had gone 45 (07-26) → 44 (07-28).
It reasoned that the seven snapshotted agents had gone **+1** — `tool-improver.update_component`
having gained a handler — so the expected total was 46, and therefore
`[UNRESOLVED] **two handlers are unaccounted for**`.

**Both halves were wrong, and the second was wrong in an interesting way.**

The `+1` never happened: seed 220's own snapshot, taken at the exact moment the 45 was
counted, already shows `update_component` carrying `error_step: refuse_mangled_write`.
Nothing was gained after it. Expected total 45, observed 44 — **one** unaccounted, not two.

The deeper error is what the number measures. The census filter is

```sql
value->>'error_step' IS NOT NULL AND value->'config'->>'error_step' IS NULL
```

— *step-level **ONLY***. A step that gains a `config.error_step` twin silently **leaves
the population** without anything being removed. And mirroring a step-level handler into
`config` is precisely what seed 219 did, and precisely what 086 recommended as the safe,
contained remedy. **The metric goes down when the fix is applied.** Eight steps sit in
that state today (`page-build-handler` ×7, `page-content-writer.resolve_links`,
`tool-improver.update_component`), invisible to a count that reads as "how many handlers
exist".

**What caught it.** Diffing seed 220's snapshots against live, per step, on the exact
fields the filter tests — rather than re-deriving the total. On all seven snapshotted
agents the only differences are the ten disabled handlers plus the two re-pointed that
evening. Nothing lost, nothing gained.

**A first attempt at the same question produced six false positives**, worth recording
because both causes are generic:
- **loop-expanded steps** (`process_sites_iter_0_call_orchestrator`) are synthesised at
  runtime by the loop handler and never exist in any definition — a plan step does not
  imply a definition step;
- steps carrying **both** forms were excluded by the live-side filter but present on the
  plan side, so the diff invented losses out of a filter mismatch it had written itself.

**The cheap check:** when a filtered count moves, **diff the population, not the total**
— and re-read the filter before trusting the delta, because the filter defines what
"missing" means. A count whose predicate excludes the remediated state will always report
the remediation as loss. Related: `[[a-check-answers-the-question-you-encoded]]`,
`[[narrow-filter-defines-the-conclusion]]`.

**Residual, stated honestly:** the remaining −1 is still unattributed. It can only lie
among agents that had no baseline at all — six of the sixteen handler owners had never
been snapshotted (`css-patch-agent`, `improvement-loop`, `tool-auditor`,
`design-audit-agent`, `site-review-agent`, `internal-linker`; 17 handlers between them).
Seed `260` baselines all sixteen at one timestamp, so the next time this moves it is a
two-table diff. That does not recover the missing one — **a baseline only answers
questions asked after it is taken**, which is the whole argument for taking it before you
need it.

## 2026-07-28 — "adoption is slow" was a diagnosis I never checked, and it was wrong

**The claim.** `bugs_closed/101`'s coverage ratchet sat at 1 adopted action of 152.
Every doc describing it — the bug file, the concept register, the handoff I was
handed, and my own first plan for the evening — framed the remaining 208 undeclared
actions as an adoption problem: keep declaring, drive the number down. §3c of the
handoff is titled *"Drive the coverage number down"*.

**What was actually true.** Opt-in is gated on `len(spec.ConfigKeys) == 0`, and
`ConfigKeys` has a specific declared meaning — *"keys this action reads that are NOT
data-input fields — settings rather than references"*. For the large class of
actions whose every config key **is** a data-input field, there was no honest way to
opt in at all: you had to duplicate keys into a list they do not belong in, or call
a reference a setting. 151 authors did the cheapest correct thing, which was nothing.

**What caught it.** Trying to do the work. The first action I opened to declare —
`append_doc_note` — already listed all eight of its live keys, in `Optional`. That
made no sense against "nobody has declared their keys", and reading the gate
explained why.

**The cheap check that would have caught it months earlier.** *When a voluntary
mechanism has ~0% adoption, read the mechanism before exhorting the population.*
One in 152 is not a measurement of 152 authors' diligence; it is a measurement of
the mechanism's cost. I had the number in front of me in three documents and read it
as a backlog every time.

**Note the recurrence, which is the part worth the entry.** This is the same lane's
second self-inflicted blindness in the same tool in one day: earlier, declaring two
keys to fix a bug silenced the very report built to catch that bug (entry above).
Both have the shape *the fix authored the thing that hid its own case* — first the
filter, then the gate. A mechanism's author is the worst-placed person to notice
that its cost is prohibitive, because for them it was not.

Fixed in `ce9e28784` (`CheckConfig`), 208 → 152 undeclared actions.

---

## 2026-07-28 — "complete coverage", from a query about a neighbouring question (bugfix 129, retry replay)

**The claim.** In a council submission I wrote that `call_agent` and `spawn_agent`
are "the only awaited senders seeded fleet-wide, so coverage is complete" — and,
worse, I wrote it in the confident register reserved for measured things, *because
I had measured something*.

**What I actually ran.** A census of `agent_definitions` for every distinct
`$.workflow.steps.*.action`, filtered to names matching spawn/call/orchestrate. It
came back `spawn_agent` (93) and `call_agent` (80), and that is true.

**Why it does not support the claim.** That query answers *"which spawn/call
actions are seeded"*. Coverage turns on *"which actions await a response"*. Those
are different questions, and the first cannot falsify the second — there is no
filter to notice being wrong, because the query is simply about something else.
`scrape_web` and `web_search` are seeded, do await responses, and are nowhere near
a name containing "spawn" or "call".

**The measurement that settles it** — 6 of 428 retried requests in 14 days were
produced by neither wired action:

```sql
SELECT CASE WHEN COALESCE(target_agent_id,'') <> '' THEN 'call_agent/spawn_agent (wired)'
            WHEN requests_topic LIKE 'system.adapter%' THEN 'adapter (re-executes)'
            ELSE 'OTHER awaited sender' END, count(*)
FROM awaited_requests WHERE sent_at > now() - interval '14 days' AND retry_version > 0
GROUP BY 1;
```

**The cheap check that would have caught it.** Split the population you are
claiming to cover **by a property only the covered producers have** — here,
`call_agent` and `spawn_agent` both always set `target_agent_id`, so its absence
names every sender they did not produce. One query, over the actual retried rows,
instead of over the config that describes them.

**What caught it.** Three council seats independently pressing on the coverage
claim — none of them could re-run my SQL (`awaited_requests` is not in the schema
they are given), so all three flagged it as *asserted*. The verdict itself was a
veto on an unrelated axis; the correction came from the objections underneath it.

**The recurrence worth noting.** This is the twin of the entry pattern
"[a narrow filter defines the conclusion]" and it is the *harder* twin: there, a
filter quietly describes a small world; here there is **no filter to notice**. The
question was substituted, not narrowed. The tell is that the query's subject
(actions in config) and the claim's subject (requests on the wire) were different
nouns — and I never wrote them next to each other.

---

## 2026-07-28 (evening) — "the grep returns one hit": the conclusion was right, the evidence for it was undercounted

**The claim.** In council rounds 6 and 7 of the markdown-index plan I supported
the blast-radius claim *"no reader switches on `code_symbols.kind`"* with a stated
measurement: **"re-checked round 6: one hit, resolved as the analyser's
`typeDef.Kind`, not the table."** One hit, named and dismissed. It reads like a
census.

**What is actually there.** Four branch sites on a variable named `kind` across
the three live files that touch the table:

```
diagnose_code_lookup_action.go:246   switch kind { case "content" … }
diagnose_code_lookup_action.go:415   func validCodeCheckKind(kind string) switch kind { case "symbol","content","ls" }
diagnose_code_lookup_action.go:550   return fmt.Errorf("unrecognised kind %q", c.Kind)
code_symbols_actions.go:430          kind := td.Kind      // the analyser's typeDef, write path
```

**The conclusion still holds** — none is the column. The first three are the
*code_check* kind, a separate namespace enumerated at `:414-420` (`symbol|content|ls`);
the fourth is the one I had named. **That is what makes this worth logging.** A
wrong conclusion gets caught. A right conclusion resting on evidence that was
never actually gathered survives, gets quoted forward, and is indistinguishable in
the register from one that was.

**Why it happened.** I was greping *for the column* — scanning hits for
`code_symbols.kind` and mentally discarding the rest as obviously-not-it. That is
a filter applied in the head, after the tool ran, and it leaves no trace in the
number you then report. "One hit" was the count of hits I had **kept**, reported
as the count the grep **returned**.

**The cheap check that would have caught it.** Write the alternation out and read
the raw count before interpreting anything:

```bash
grep -rnE 'switch|case |== *"|!= *"|\["kind"\]|\.Kind' --include=*.go \
  platform/orchestration/actions/{diagnose_code_lookup,code_symbols,analyse_repo_local}_action*.go
```

Then dismiss the hits **in the write-up, one by one, with their line numbers** —
which is exactly what makes the dismissal auditable by someone who cannot run your
grep. Round 8 does this and it takes four lines.

**What caught it.** The `guardian` seat, which did not challenge the conclusion at
all. It said: *"the blast-radius claim rests on a human grep, which is fine, but it
is exactly the kind of claim my own tier cannot re-verify — `code_checks` only see
declarations, never switch statements inside function bodies."* It asked instead
for the claim to be corroborated by **which pipelines actually touch the table**.
Going to answer *that* question is what made me re-run the grep properly.

**The transferable bit.** A reviewer saying *"I cannot verify this"* is not a
weaker objection than *"this is wrong"* — it is the objection that finds the
claims nobody will ever check. And when a seat names the epistemic limit of its
own tier, **concede it in the resubmission rather than answering as if it were a
challenge to the conclusion**; the honest form ("this is a human grep, here are
the four hits and why none is the column") is both shorter and more persuasive
than a defence.

**Recurrence.** Same family as *"a narrow filter defines the conclusion"* and
*"a check answers the question you ENCODED"*, with the filter moved one step
later: not in the query, not in the question, but in the **reading of the output**.
The tell is a suspiciously round, suspiciously small count — *one* hit, *zero*
collisions — reported without the hits themselves being listed.

---

## 2026-07-28 — "no council seat sets `max_tokens`" (council parallelism thread)

**The claim, written into a committed handoff:** *"No seat sets `max_tokens`. All
16 set `tolerate_truncation: true` and leave `max_tokens` NULL (live
`agent_definitions` row)."* Every seat sets it, uniformly **8000**.

**How it happened.** I queried `value->'config'->>'max_tokens'` across the 16
`review_*` steps, got NULL for all 16, and reported "not set". The real path is
one level deeper — `config.ai_service.max_tokens`. I had already made the same
mistake minutes earlier: an "always-on seats" query on
`value->'config'->>'relevance_footprint'` returned NULL for all 16, which I
briefly read as *"all seats always run"* when it equally meant *"wrong path"*.
**A uniform result across every row is the tell** — real config drifts; sixteen
identical NULLs usually means the query missed, not that the fleet agrees.

**What caught it.** Two sources disagreeing about one field: `llm_call_log` had
`max_tokens = 8000` on every row while my config query said NULL. Neither reading
was checked against the other until the numbers collided.

**The cheap check that would have.** Print the object **once**, instead of
guessing a path three times:

```sql
SELECT jsonb_pretty((default_config->'workflow'->'steps'->'review_editquality')
                    - 'prompt' - 'prompt_template')
FROM agent_definitions WHERE type='council-gate' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

**The transferable bit.** `->>` returns NULL for *"key absent"* and for *"path
wrong"* identically, so **a NULL is never evidence that a setting is unset until
the path itself is proven**. When a config query returns NULL, the next query
should widen (print the object), not narrow (try another guessed path). Cost here
was low only because the correction landed in the same session, before anyone
acted on it — the claim was already committed, and the fix it implied
("set a `max_tokens` nobody sets") would have been aimed at the wrong thing.

**Recurrence.** Same family as *"a check answers the question you ENCODED"*, in
the schema layer: the query was well-formed, ran clean, and answered a question
about a path that does not exist.

---

## 2026-07-28 — a verification recipe that greps the wrong pod, and a coverage query that spans a fork

**Thread:** "bugsearch 5", closing `bugs_closed/129`. Two claims, both written into
a durable handoff by the previous thread, both false, both caught within minutes of
being acted on — which is the only reason they cost nothing.

### Claim 1 — "grep `-l app=agent-chassis` for the replay"

The handoff's §2 gave this as *the* command that decides whether the fix works. It
returns **zero lines even when a replay has just succeeded**. The retry executes in
whichever service hosts the awaiting orchestration — here `business-intel` — and
every service runs the same image, so the code is under every label and only the
execution is localised.

**What caught it.** The database and the log disagreed: `awaited_requests` showed a
row at `retry_version 1` with a fresh `sent_at`, while the log grep showed nothing
at all. A retry that happened and a retry that did not cannot both be true, so one
of the two probes was wrong — and the DB row is the harder evidence.

**The cheap check that would have.** Resolve the executor before grepping for its
output: `processing_pod` is on the row. Or loop the labels, which costs one line.

**Why it is worth a row here.** Zero lines from the expected pod is
*indistinguishable* from "hasn't happened yet", and the handoff had primed exactly
that reading ("no timeout has occurred yet"). A false negative that agrees with the
story you already believe will not be questioned. The generalisation — **grep the
orchestration's host, not the subsystem's name** — is in 016b §9.

### Claim 2 — "any NULL payload other than scrape/search is a genuine gap"

§5 of the same handoff. The live census returned 4 `deploy_page` rows; by that rule
they were gaps needing wiring. They are not. `coordinator.go:2809` forks adapter
actions into step **re-execution** before anything reads the payload column, so
those NULLs are correct by construction.

**What caught it.** Reading the branch instead of trusting the rule — the rows all
carried `target_agent_type='unknown'` and an empty `target_agent_id`, which is not
what a mis-wired agent call looks like.

**The cheap check that would have.** Follow the column to its only consumer before
writing a coverage query for it. One `grep` for `DecodeRetryPayload` shows the fork
sitting above it.

**The transferable bit.** A coverage query for a column consumed by **one branch of
a fork** silently measures both branches. Write the exclusion at the same time you
write the fork, or you manufacture a backlog aimed at a non-defect. Tell: the
"gaps" cluster perfectly on one step name or topic prefix. Same family as *"a
uniform result across every row is the tell"* — a uniform shape is usually a
branch, not a bug.

### The tally point

Both are the **same underlying skipped check**: *the probe was inherited, not
re-derived.* Each was handed over as a finished command with a stated expected
answer, and both were run without asking what they actually measure. Related to
*"a verify-later that states an expectation rather than a question"* (016b §9) —
here the expectation travelled in a handoff instead of a comment, which is worse,
because a handoff is read exactly once by someone with no context to doubt it.

**Third occurrence of the "check answers the question you ENCODED" family this
week.** Worth automating the cheapest guard: any coverage/verification query
committed to a RUNBOOK should carry its denominator, so an empty result cannot be
read as success. Done for 129's runbook; not yet general.

---

## 2026-07-28 — "relojistas is finished and self-running" — while its logo was unreadable on every page

**Where I wrote it.** `traffic_probe/HANDOFF_2026-07-28_continue_here.md` §0: *"relojistas.com
is **finished and self-running**"*, with a state table whose every row said healthy. Repeated in
the SUMMARY series and carried into a second session as settled background.

**The truth.** The site's header logo is a **two-up brand specification sheet** — the wordmark
printed twice, on a light swatch and a dark swatch — served as `logo.jpg` at 1408×768 and styled
`.logo-img { max-height: 44px; width: auto; }` with **no crop anywhere in the only stylesheet**.
So every page renders the whole board at roughly 81×44px and both wordmarks are illegible. It
had been live for weeks.

**What caught it.** Running the og-card derivation (`bugs_open/131`) and then *looking at the
PNG it produced*. The card was a picture of the same spec sheet — which is only visible to an
eye, since the job reported `complete`, the URL returned 200, `file` said a valid 1200×630
`image/png`, and the provenance rows were written. Every mechanical signal was green.

**The cheap check that would have.** **Look at the site.** Once. The entire verification behind
"finished" was a battery of `curl`s: status codes, `og:` tags, robots.txt, sitemap.xml,
llms.txt, feed items, 301/302/410 behaviour. Every one of those asks *what does a crawler
receive*. Not one of them asks *what does a person see*. I had a rich, genuinely rigorous audit
pointed at exactly one of the two audiences, and no line in it that could ever have failed on a
visual defect.

**The transferable bit.** *A site is "verified" only against the audience your checks encode.*
Crawler-facing verification and human-facing verification share the word "live" and overlap
nowhere. This is the `check answers the question you ENCODED` family again — but the missing
question here was not a filter or a denominator, it was **an entire audience**. Tell: an audit
whose every command is `curl … | grep` has no opinion about rendering, and the more thorough it
looks, the more confidently it will be cited as proof the site is fine.

**Corollary for image artefacts, worth its own line.** For a PNG, dimensions and MIME type are
*not* "checking the artefact" — they are the equivalent of checking a status code. **What the
picture shows is the artefact, and the only way to know is to look at it.** `bugs_open/012` said
check structure after a rewrite rather than trusting status; in an image medium that rule needs
an eye, because every automated signal available passed here.

### The tally point

That makes **two** entries this week where the probe measured a real thing correctly and the
real thing was not the question (the 129 inherited-probe pair above, and this). The difference:
those inherited a wrong question, this one never had the question at all. **No check I could
have added to my `curl` battery would have caught it** — the fix is not a better command, it is
noticing that a whole class of user was absent from the plan. Cheapest general guard: any doc
claiming a *site* is finished should have to name what a human saw, and when.

## 2026-07-28 — a handoff named the wrong council objection as the gating one, and it read as fact

**The claim.** `consolidation/HANDOFF_2026-07-28_continue_here.md` §4, written as the cold-start
brief for the next session, said of council corr `721ac4f7`: *"Corr … → `revise`, gating
objection from `bug_historian`… **The objection, and it is fair:** the change introduces
`prose_sections` and `no_match_sentence` as cross-step contract fields, and the council's own
read-only check found that `report-builder` has `input_contract` NULL and `output_contract`
NULL."* It then gave a two-step remediation built entirely on that reading.

**What was true.** The seat is right; the objection is not. `council_decide.decided_by` does say
*"gating objection from bug_historian"* — but bug_historian's only **high** is on edit 3 and is
about **enforcement**: *"the plan never establishes what the CALLER does with those violations…
if logged/recorded but not used to block report delivery, this is exactly the documented shape in
bugs_open/079 and bugs_open/083."* The DECLARED CONTRACTS point is real but **medium**, and came
from three *other* seats (`guidelines` ×2, `prior_art`, `guardian`). The handoff had fused the
gating seat's *name* to a different seat's *content*.

**What it would have cost.** The remediation as written answers four mediums and leaves the high
untouched — a near-certain second `revise`, at ~30 minutes of queue and a full council round.
Worse, it had already been carried forward once: the same §4 also instructed *"fix edit 1's
`symbol` field"*, which on inspection was **correct as submitted** (`assessPayloadRated` is a
two-hop delegation from `scoreGrippers`, and the guard genuinely lives there). Following the
brief would have introduced an error into a plan the seat had merely asked to have *confirmed*.

**What caught it.** Reading the artefact instead of the summary — one query printing every seat's
verdict and objection severities side by side:

```sql
SELECT k, collected_data->k->'result'->>'reviewer', collected_data->k->'result'->>'verdict',
       (SELECT string_agg((o->>'severity')||':edit'||(o->>'edit'), ', ')
          FROM jsonb_array_elements(collected_data->k->'result'->'objections') o)
FROM orchestration_states, jsonb_object_keys(collected_data) k
WHERE orchestration_id='<run>' AND k LIKE 'review\_%' ORDER BY 1;
```

Ten rows, one per seat. The high stands out immediately; so does the fact that it is on a
different subject from the one the handoff named.

**The transferable bit.** *`decided_by` names the seat, not the objection.* A REVISE verdict has
one gating reason and often several advisory ones, and a prose summary written at the end of a
long session is exactly where those get merged — the most-repeated objection displaces the most
*severe* one, because three seats saying "contracts" is louder than one seat saying "enforcement".
**Never inherit a council verdict through prose.** This is the `writes the field ≠ reads the
field` family: the handoff faithfully recorded that a gating objection existed and silently
substituted its content.

**Second, sharper point — the brief was confidently wrong in the direction of the easier task.**
Declaring two contract fields is a small, mechanical, satisfying job. Proving that a detector's
output actually blocks delivery is an open-ended read through the engine. Both readings were
available at write time; the one that survived into the handoff was the tractable one. Worth
distrusting on that shape alone.

**Cheapest general guard.** A handoff that routes the next session at a verdict should paste the
**severity table**, not a paragraph — and should quote the gating objection *verbatim*, since it
is the one thing the next session must answer.

**Footnote, a landmine walked into on the way.** Verifying the enforcement claim, the first query
asked `s.value->>'error_step'` and got NULL for every step in report-builder — which reads exactly
like "no error routing anywhere" and would have *confirmed* bug_historian. `error_step` persists
at `s.value->'config'->>'error_step'`. **This was already written down** in `016b` (§ ~663, and
the census at ~4890: 0 of 14,209 persisted plan steps carry the step-level twin vs 1,828 carrying
the `config` one). Knowing a trap exists does not help if you query the field name you expected;
the tell is a result that is uniformly NULL across *every* row, which is almost never a fact about
the world.

---

## 2026-07-28 (later) — a reversal trigger that counts the opposite of what it means, and a snapshot I declared missing

**Thread:** "bugsearch 5", seating `review_architecture` on the fix lane after the
owner reversed D9. Two more, both caught inside a minute, both worth the row
because each was about to be handed to the owner as fact.

### Claim 1 — "the trigger has fired 68 times across 50 submissions"

D9's reversal condition was *≥3 architecture-level escalations routed to a human*,
with the query written into the decision doc: `body ILIKE '%rchitecture-level
concern%'` over `council_report`. It returns **68 / 50**. I was one sentence from
reporting that as a 16×-threshold mandate.

**It measures the negation.** The phrase is seeded in **four prompts**, so seats are
*taught* to use it — and 26 of the matches are seats writing *"**No**
architecture-level concern in this area"*. Split properly: **13 affirmative reports
across 10 submissions**. Still over threshold, so the conclusion survived; the
number I would have quoted was wrong by 5×, in the direction that flatters the
decision.

**What caught it.** Asking the cheap question before quoting a number that was
about to justify reversing an owner ruling: *is this string in any prompt?*
```sql
SELECT type FROM agent_definitions WHERE deleted_at IS NULL
  AND default_config::text ILIKE '%<the phrase>%';   -- 4 rows ⇒ the count is polluted
```
Then read three matches verbatim. Both checks together took under a minute.

**The transferable bit.** This is
[[declaring-a-key-silences-your-own-detector]] inverted: there, declaring a key
silenced a detector; here, **teaching seats a phrase made the detector for that
phrase fire on everyone who considered the question and said no.** Any
free-text detector whose phrase also appears in the instructions is measuring
compliance with the instructions, not the signal. **Before believing a
`body ILIKE` count, grep the prompts for the same string.**

Aggravating factor worth naming: the trigger was authored in the same document that
recorded it as `[UNMEASURED]`. **An unrun query acquires authority by sitting in a
decision register** — it looks like a measurement because of where it lives.

### Claim 2 — "099 printed 'snapshot taken' and no snapshot exists"

I checked `agent_definitions` for `is_snapshot=true`, found nothing, and wrote that
the safety mechanism was a false print — citing
[[a-print-statement-is-not-a-config-row]] by name.

**Wrong: snapshots go to `agent_definitions_backup`.** Reading `snapshot_agent`'s
body showed the INSERT target immediately, and the row was there — 21:56:37Z, at 16
seats, exactly as advertised. The mechanism works.

**The transferable bit.** *Having the right pattern loaded is not the same as
checking it applies.* I reached for a known failure shape and let it substitute for
reading the function — the same substitution the pattern itself warns against. When
a safety mechanism appears to have failed, **read its implementation before
reporting it broken**: a missing row is far more often the wrong table than a lying
print. Cheap check: `SELECT prosrc FROM pg_proc WHERE proname='<fn>'`.

### The tally point

Both are the family already dominating this file — *the check answers the question
you encoded*. The new wrinkle in Claim 1 is that the polluting text was **our own
prompt**, which no amount of care about filters would have caught: the query was
well-formed, the table right, the population right, and the string meant something
different in the corpus than in the author's head. **Add "grep the prompts" to the
standing checks for any council-corpus metric** — that now covers both this and the
`%deflect%` false-citation trap logged 07-27.

---

## 2026-07-28 — I recommended a tool that had been deleted the same day, from my own memory index

**The claim.** In the opening plan for the webdesign.uk build service, the free-teaser section
listed the machinery we could point at a prospect's existing site: browser-runner checks, the
44px rule, and *"the contrast scanner (`cmd/contrastscan`)"*.

**`cmd/contrastscan` does not exist.** It was built earlier the same day by the oufe thread,
found to duplicate `scripts/render_audit.py`, and **deleted on discovery**. The register entry
(VIZ-010) already recorded the whole episode, including its cause: *"the prior-art grep was
`--include=*.go` and the prior art is Python."*

**What caught it.** `scripts/pattern-check.py`'s `new-capability-surface`, on the commit hook,
about ninety seconds after the file was written. Worth noting *how* it fired: not on code
creating a path, but on **a document proposing one**. It printed the real `cmd/` listing
underneath, so the correction took one `ls`.

**The cheap check that would have.** `ls cmd/`. One command, no cluster, no grep syntax to get
wrong. I had run a dozen live database queries to ground everything else in that plan and did not
run this one, because the tool name did not feel like a claim — it felt like a fact I already
knew.

**The transferable bit, and it is not "check your facts".** *The name came from my own auto-memory
index*, which said `cmd/contrastscan` is the browser witness. That line was written before the
deletion and was stale by hours. **A memory index is a claim with no evidence attached and no
expiry**, and it is read in the one posture where nothing gets verified — cold start, where its
whole purpose is to save you the lookup. So: **a path, a symbol or a flag recalled from memory is
an [UNVERIFIED] claim wearing the clothes of a settled one.** The system prompt already says to
verify a memory that names a file before recommending it; the interesting part is that the memory
was *right when written*, which is the failure mode no amount of care at write time prevents.

**Follow-through.** The stale memory was corrected in place rather than only the plan — otherwise
the next thread reads the same line and makes the same claim. It now says the tool does not exist
and names what caught it. Related: [[prior-art-search-goes-stale]] (an absence is only true when
you looked) — this is its mirror image, **a presence is only true when you looked**, and the
half-life is the same.

---

## 2026-07-28 — "suites green" was true of my working tree and false of the commit, and HEAD did not compile

**The claim.** Commit `93003e6e0` (bugs_open/104, fleet-wide banned claims) says in its own
message: *"8 unit tests; full datahelpers/actions/discovery_checks suites green."* I had run them,
they were green, and I wrote it down. **The commit itself did not compile.**

**What was actually wrong.** The pathspec named `validate_page_content.go` and
`check_unverified_claims.go` because I had edited them. Both files *also* carried another
session's uncommitted `ClaimSurface` plumbing (bugs_open/102) — the same-file passenger CLAUDE.md
says no hook can catch. So my commit shipped the two **consumers** of a type whose **definition**,
in `claims.go`, was still sitting uncommitted in the shared tree:

```
check_unverified_claims.go:295: undefined: datahelpers.ClaimSurface
check_unverified_claims.go:414: too many arguments in call to eb.ScanUnregisteredNumbers
```

That matters more than an untidy commit: **`make build-<service>` builds from committed HEAD**, so
for the four minutes until the 102 session committed its half (`3ddb4ed2d`), any session that
started an image build would have got a compile failure with my name on the commit.

**What caught it.** Not the tests, not the commit-scope hook — both were happy. It was reading the
commit's own diff afterwards because the insertion/deletion counts looked too large for the edits
I remembered making, and noticing a hunk header that read
`func scanComponentClaims(html, slotName string, eb *datahelpers.EvidenceBase)` — **without** the
`surface` parameter that was in my working tree. A hunk whose context is code you did not write is
the tell.

**The cheap check that would have.** One command, before or straight after committing:

```bash
git archive HEAD | tar -x -C /tmp/headcheck && (cd /tmp/headcheck && go build ./platform/...)
```

**The transferable bit.** Every test I ran, and every build, was against **my working tree** — which
contained the other session's definition. The tree is the union of everybody's WIP; HEAD is what
ships. **A green test in a shared tree proves the union compiles, never that your commit does**, and
the gap is invisible precisely when the missing half is someone else's uncommitted dependency. This
is the mirror of [[shared-tree-wont-compile]], which is *their* WIP breaking *my* tests; this is
*their* WIP being carried by *my* commit and breaking everyone.

**Also worth saying plainly:** the pathspec discipline did its job on every file it could — it kept
five other sessions' modified files out. It cannot help with a file two sessions are editing at
once, and this is what that costs. The fix was forward-only: I prepared the missing half as a
labelled `sweep:` commit, and by the time it ran the owning session had committed it themselves.

---

## 2026-07-28 — a total that was before+after summed, and two "verbatim" fixtures that were paraphrased

Session "bugsearch 7", fixing `bugs_open/102` (the claims layer reading page type). Two wrong
calls, both mine, both caught inside the same hour — and the second one only because the test
that caught it was written to be able to fail.

**1. "124 findings before, 187 before." Both were printed by me, ten minutes apart.**

The fleet survey wrote per-site scanner output to `survey/<domain>.out`, and the re-run with the
fixed binary to `survey/<domain>.fixed.out`. Then:

```bash
cat survey/*.out | grep -cE '^(BANNED|NUMBER)'     # "before" = 187
```

`survey/*.out` matches `survey/x.fixed.out` too. **187 = 124 + 63 — the before and after totals
added together** — and it did not look wrong, because 187 is an entirely plausible number of
findings for nine sites. Had I quoted it in the bug file, the fix would have read as suppressing
124 findings rather than 61, i.e. as roughly twice as blunt an instrument as it is.

*What caught it:* the per-page-type breakdown summed to 124 while the total said 187. Two views
of one population disagreeing.

*The cheap check that would have:* **never let a glob define a population you are going to
compare.** Iterate an explicit list of the things you exported, and print the row count per item
alongside the total, so the two must agree in front of you. A suffix that is also a suffix of
another suffix (`.out` ⊂ `.fixed.out`) is the specific trap; `-name '*.out' -not -name '*.fixed.out'`
or, better, different directories.

**2. I wrote "every fixture below is a VERBATIM live false positive" over two I had typed from
memory of the scanner's output.**

The file header of `claims_surface_test.go` claimed all eight test fixtures were live blocks. Two
(the `tool` and `game` cases) were shortened paraphrases of the ±60-character snippet the scanner
prints, not the block it actually scanned. Shortened past the point where they contained the word
that made them flag at all — so they flagged **nothing on any surface** and could not have
discriminated the fix from a scanner that was simply switched off.

*What caught it:* the negative control in the same commit — the test asserting the same blocks
still flag on `content`/`landing`/`report`/unknown surfaces. It failed on exactly those two rows.
**Had I written only the "editorial pages raise nothing" direction, both fixtures would have
passed and pinned nothing**, and the file would have carried a claim of live provenance over
invented text.

*The cheap check that would have:* if you label a fixture with its provenance, **paste it from the
source** — and assert the fixture's precondition (that it flags at all) before asserting what the
change does to it. A fixture that cannot fail the old code proves nothing about the new.

*A third, same session, of the same family:* my first regression test looped over seven page types
calling `ScanBannedClaims(block)` — a function that takes no page type. Seven iterations, one
assertion, loop variable unused. It read as a per-page-type matrix and was a single check wearing
a costume. **`go vet` does not flag an unused loop variable that is used in the failure message.**
Rewritten as the discriminating pair (on a guide: banned fires, the number scan does not).

**Related, and already logged from the other side:** the entangled-commit incident above
(`## 2026-07-28 — "suites green" was true of my working tree…`) is the same afternoon and the same
four files — that entry is by the session whose commit took my half; `3ddb4ed2d` is mine restoring
it. Read the two together: neither session did anything careless, and HEAD still stopped compiling
for four minutes.

---

### 2026-07-28 — experience register — "attribute assertion will give the openable/inert split real Tier 2 coverage"

**The claim.** Written in `NOTES_experience_register.md` 2026-07-28e, mid-session: implementing
attribute assertion would turn `template_row_not_a_control` executable and so give CC-001's central
clause real coverage, because the hidden template IS in the served HTML while the cloned rows are not.

**What caught it.** Running it. The check does become executable, and it **FAILS**: the served
template carries `href="#"`, and the platform's dead-control sweep already exempts
`data-runtime-fill` shells for precisely that reason. The clause is a claim about the
post-hydration DOM being asserted against pre-hydration markup.

**The cheap check that would have.** `curl` the page and look at the element **before** writing the
prediction — one command, and I had already fetched the page for another purpose. I reasoned from
"the template is in the HTML" (true) to "so the check will pass" (never verified) in the same
paragraph, with nothing marking the second half as unchecked.

**The second-order one, which is the reason this entry is worth its space.** The resolution I
reached — re-tier the clause to 4 — is *also the reading that makes my own red result disappear*.
It may well be right. But I noticed I had arrived at it quickly and comfortably, and that is the
shape to distrust. Recorded as such in the entry, in NOTES, and filed as `bugs_open/137` for
someone else to settle rather than closed by me.

### 2026-07-28 — experience register — 7 entries "fail validation", reported by a harness I under-fed

**The claim, nearly written up as a finding.** Running the nine stored register entries through
`ValidateExperienceCriteria` reported 7 of 9 failing on unused bindings — alarming, since every one
had been written *through* that validator and must therefore have passed it.

**What caught it.** The contradiction itself. Entries written through a validating path cannot fail
that path; something about the measurement had to be wrong. It was: the validator takes `extra`
documents (`contract`, `states`, `data_contract`) whose placeholders also close, and my harness
passed only `criteria_template` + `binding_schema`. With the full document set: 0 errors.

**The cheap check that would have.** Read the signature of the function before calling it —
`extra ...interface{}` is the whole story, and it is documented in the comment directly above.

**The pattern.** This is `check-answers-the-question-you-encoded` again: no filter to notice, no
error raised, just a question quietly narrower than the one I meant to ask. The tell was not in the
output — it was that the output disagreed with something I already knew to be true.

---

## 2026-07-29 — I removed a time filter and briefly believed my own fix had collapsed

**Thread:** "bugsearch 5", re-verifying `bugs_closed/129` after a fresh roll.

Chasing 6 unrecorded payloads, I re-ran the capture census **without a `sent_at`
bound**. It returned ~7,000 NULL-payload rows across ~100 step names —
`spawn_dispatch` 720, `call_dispatch` 553 — i.e. the fix appeared to have failed
fleet-wide on exactly the steps it was written for.

**It is pre-fix history.** `request_payload` has only been written since the
07-28 20:48 roll, so every row before that is necessarily NULL. Correctly scoped:
**145 replay-path requests, 139 recorded (95.9%)**, one step unrecorded.

**What caught it.** The shape of the result. A fix that had genuinely collapsed
would not produce a *uniform* failure across every step name including ones that
demonstrably worked an hour earlier — and I had a verified 18/18 from the same
query with a bound on it. Two runs of "the same" query disagreeing means one of
them is a different question.

**The cheap check that would have.** **When a column is new, bound every census by
the migration or roll that introduced it.** Its "missing" count is otherwise
dominated by the era before it existed. One clause: `AND sent_at > '<roll>'`.

**The transferable bit, and it is the mirror image of the usual one.** The standing
lesson is [[narrow-filter-defines-the-conclusion]] — a filter taken from the
question describes a small world and flatters you. This is the same error running
backwards: **removing** a filter silently widened the population to one that
predates the thing being measured. Both are the same discipline —
*state the population in words first, then check the query describes it* — and a
count over a column that has not always existed is the most common instance.

**Related, same session, same file:** §5 of the 129 handoff listed *step names*
where the real defect is per-*action*, so `search_news` presented as a new gap when
it is `WebSearchAction` under another name. **A "known exceptions" list should ship
the predicate that generates it, not the instances** — instances go stale the moment
someone adds a caller, and each staleness event costs another investigation.

---

## 2026-07-29 — six wrong calls in one webdesign.co.uk session, and the two families they fall into

Logged together because the **pattern** is worth more than any single row. Six
errors over two days, and every one belongs to one of two families.

### FAMILY A — I reported a mechanism I had not watched run

**A1. "Re-rendering chrome is safe because `applyNavVisibility` drops the News item."**
Written into a handoff correction *and* the memory index as the reason a documented
hazard was obsolete. The **conclusion was right** — no 404 shipped — and the
**mechanism was wrong**. What actually fired was `refresh_nav_tables`, which DELETES
and repopulates `site_nav_items` from deployed pages, so the row was simply never
recreated. Two independent code paths would each have produced the observed outcome
and I asserted the one that did not run.
*Caught by:* reading the real orchestration's `collected_data` afterwards.
*Cheap check:* the log says which path ran. **I had marked the claim `[UNVERIFIED —
no production execution trace]`, and that marker is the only reason this was cheap
to correct.** The discipline worked exactly as intended; the lesson is to keep using
it, not that it failed.

**A2. "The feed is failing fleet-wide."** Told the owner at 19:56 on a zero count
across all sites over two hours, reading it as the same stall class as the previous
tick. It was **slow, not stalled** — sites process one at a time, ~6 min per worker,
and ours was 5th of 5. Items landed 15 minutes later.
*Caught by:* waiting.
*Cheap check:* **a slow job in progress and a dead one look identical.** The signal
I used — parent `updated_at` frozen — is what a parent legitimately does while
awaiting a child. Before declaring a stall, find something that would have moved.

**A3. "No child orchestrations, so the spawn never took."** Wrong shape entirely:
children are spawned **pods**, not rows carrying `parent_orchestration_id`. The
worker was running the whole time.
*Cheap check:* `kubectl get pods` before concluding a spawn was lost.

**A4. Nearly filed the exact inverse of a real bug.** About to report that news feeds
default to a web search *because the sources do not set `search_type`* — true of the
config in front of me, and wrong about the system, because the action **forces** it
one hop later. The real defect was further on again (the provider interface cannot
carry it). A reviewer could not have caught this without opening the same file.
*Cheap check:* **grep the key in the CONSUMER**, and follow it to the call that uses
it. Presence in a payload, a log line or a doc comment proves nothing.

### FAMILY B — a filter or a count that quietly answered a different question

**B1. Two consecutive wrong link-topology counts, five minutes apart.** First: "63
tool pages link to a guide" — an artefact of counting `/learn/index.html`, which is
in the **footer of every page**. Second, after stripping chrome: "0 links
everywhere" — my exclusion of landing pages dropped every URL ending `index.html`,
which is **the entire tool namespace**. The true answer was the opposite of my first:
tools linked nowhere, guides linked to tools well.
*Caught by:* a debug sample printing a row that contradicted the aggregate.
*Cheap check:* **print sample ROWS beside every aggregate.** An aggregate cannot
contradict itself. Already recorded as `narrow-filter-defines-the-conclusion`; this
is the third instance.

**B2. Category counts summed to 62 against 63 tools on disk.** A regex requiring
`<h3>` immediately after the anchor missed one card. The number was about to be
printed on the live home page, on a site that has shipped invented figures twice.
*Caught by:* the totals not reconciling.
*Cheap check:* **reconcile any count against an independent source before
publishing it.** The mismatch was the entire signal.

**B3. A published prediction, half wrong.** I wrote before the tick that a retuned
query would return **fewer than 9** items with better relevance. It returned **9
again** — relevance right (9/9 on topic, up from ~1/9), count wrong.
Kept as a row deliberately: **writing the prediction down first is what made the
miss visible instead of retrofittable**, and the relevance half is the one that
mattered. A half-wrong falsifiable claim beats a vague right one.

### Two near-misses that cost nothing because the check ran first

- **"Widen the starved feeds' window to a year."** Would have achieved nothing —
  `feed_actions.go:878` hard-codes a 30-day ceiling, so a wider window fetches a
  year only to discard past day 30. Checked before recommending; never left my head.
- **My own verify blocks failed twice** — `array_agg(… ORDER BY 1)` sorts by the
  *constant* 1, and `jsonb_array_elements_text(…) k` aliases the table, not the
  value. Both times `ON_ERROR_STOP` rolled the transaction back and I confirmed
  nothing had been applied. **A verify block that can fail closed is worth more than
  one that always passes** — these cost a re-run each and would have caught a real
  drift identically.

### The row that is really about work, not words

**"The 404 page now carries analytics."** True, deployed, verified present at
`/404.html` — and **inert**, because a missing path never reaches that file. I had
checked that the file was deployed and contained the change, which was true and
beside the point.
*Cheap check:* **fetch the artefact by the route a user takes.** That one probe
turned a cosmetic fix into `bugs_open/132`, a fleet-wide finding.

### Tally

**Family A is the expensive one** — four of six, all of them "I explained a
mechanism I had not watched run", and all four reached either the owner or a
document before being corrected. Family B is well-covered by existing rows and keeps
recurring anyway. **The single check that would have prevented the most damage here
is the cheapest: before asserting *why* something happened, find the line that says
it happened.** Everything else in this session was already written down somewhere.

Family: read-the-function-dont-infer-from-the-data, narrow-filter-defines-the-conclusion,
a-slow-job-and-a-dead-one-look-identical, verify-on-the-path-a-user-takes.

## 2026-07-29 — a repro that the render destroys, and a census that answered a different question

**"The gamesdesign rerender will drive the stored `href=""` through `save_page_sections` and
prove the 079 fix."** Pre-registered in a council submission, written into a verify script's
own `EXPECT` line, and stated to the owner. **It ran, COMPLETED in 40s, and the `href=""`
count went 2 → 0 — the predicted result, for the wrong reason.** `rerender_sections`
re-renders each section from `content_data` through the CURRENT template; that page's
`content_data` carries `cta_primary_label`/`cta_secondary_label` and **no url fields at all**,
so the template's skip gate omitted the buttons entirely. The repair was never handed an
anchor. No `action='save_page_sections'` repair row was written for the run at all.

I had checked the preconditions I knew about — `content_data` non-NULL on every section (else
the page escalates to the writer and `save_sections` is skipped), not runtime-fill exempt, not
interactive, `rebuild_policy='generic'`, and the step graph confirming the `reason` routes
through `save_sections`. Every one of those was a real trap and none of them was this one.

*What caught it:* **the success was the wrong SHAPE.** Unlinking is defined to drop the `<a>`
and keep the inner text; the anchor text had vanished too. A result that matches your headline
metric while violating the mechanism's own invariant is the tell.

*Cheap check that would have:* **before using a regenerate-from-source rerender as a repro,
confirm the defect exists in the SOURCE, not only in the rendered artefact.** One query —
`SELECT content_data FROM page_components …` — showed no url fields, i.e. nothing that could
render an `href=""` the repair would then see. The defect lived in `rendered_html`; the render
does not read `rendered_html`.

*Second-order:* I nearly attributed the 2 unlinks that WERE logged for that page to my fix.
They came from the outbound rerender seam (`action='rerender_page'`, LNK-023), a different
call site acting on the assembled page's chrome. **When several call sites share one error
code, the discriminating field is the only thing standing between you and a false positive** —
here `action`, exactly as `linkRepairOrigin` was designed to allow.

**"Six agent types call `save_page_sections`, so this closes the door."** Withdrawn under a
council objection (`bug_historian`). The census enumerated `agent_definitions` carrying a
`save_page_sections` STEP NAME; the load-bearing question was *who writes
`page_components.rendered_html`*. Ten Go call sites do, and three persist LLM prose with no
repair at all — including the one our own code comment tells operators to use for targeted
edits. Same family as the earlier `narrow-filter-defines-the-conclusion` rows, one level up:
not a filter that was too narrow, but **a question that was adjacent to the one the claim
rested on**. Filed as `bugs_open/136_…section_editor_and_three_siblings…`.

*Cheap check:* when a claim is about a COLUMN's contents, grep for writers of the COLUMN
(`grep -rn "INTO <table>\|UPDATE <table>"`), never for the name of the step you happen to
know about.

Family: narrow-filter-defines-the-conclusion (question-adjacent variant),
verify-the-failing-branch, writes-the-field-is-not-reads-the-field.

## 2026-07-29 — consolidation programme (tools-api client identity)

**"tools-api: a visitor can still choose the IP they are rate-limited as (and the IP we
store)."** Written as the TITLE of `bugs_open/139` and filed, on the strength of three
verified code facts: two `c.ClientIP()` call sites, `gin.New()` with no
`SetTrustedProxies`, and `bugs_closed/090` having proven that exact mechanism against
production on a sibling service. **Refuted the next morning by two probes.** A forged
`X-Forwarded-For` and a forged `X-Real-IP` each returned 200 and each stored the same
hash as every other row: `sha256("172.18.0.1")` — the docker bridge gateway. Caddy
overwrites `X-Forwarded-For` with its own peer before the app sees it, and Cloudflare
strips `X-Real-IP` at the edge. The visitor chooses nothing.

Every code fact in the filing was true. **The conclusion was drawn at the service
boundary, and the two hops in front of the service decide the outcome** — one of which
(Caddy) is in a config file on the island, not in this repo, and the other (Cloudflare)
is not in any file at all. `090`'s mechanism transferred; `090`'s *proxy* did not, and
the proxy was the load-bearing half.

*What caught it:* **the check the file itself named.** The filing carried an honest
`[UNMEASURED]` marker saying the probe had not been run and telling the reader to run it
before quoting the bug as proven. That marker is the only reason this was caught in a
day instead of surviving into the fix. **It is also not enough:** a marker on the
evidence paragraph did nothing to stop the unhedged claim being written into the file's
TITLE, into its `Class` line, and into the workstream handoff's NEXT action. **A hedge
that does not reach the headline is not a hedge** — the title is what the next thread
greps, and titles have no room for markers, which is precisely why an unmeasured claim
should not be one.

*Cheap check that would have:* the single `curl` already written into the file, ~30
seconds. There was no obstacle — the probe was named, the command was drafted, the
endpoint was known. The gap was purely that filing felt like the end of the work.
**If you can write the verification command, you can run it; if you are not going to
run it, the claim is not ready to be a title.**

*What the same probe found, which is the actual bug:* the identity is a **constant** —
83 of 83 rows, one distinct value, since the table was created. So the "per-IP" limiter
is one global bucket shared by every visitor, and the persisted `client_ip_hash` column
has never distinguished anybody. **Worse than the filed bug and it needs no attacker.**
Also: the recommended fix (`platform/httpguard`) would NOT have fixed it, and its
docstring's guarantee — "`X-Real-IP`, set with `proxy_set_header`, so a client-supplied
one is replaced" — is a property of **nginx on idea.uk**, not of Caddy on the island,
which forwards a client-supplied one verbatim. A second adopter inherits the docstring
and not the nginx.

*Second-order:* my first hypothesis on seeing the constant was that httpguard would
therefore *create* the spoof by preferring `X-Real-IP`. That was wrong too, and I nearly
wrote it into the bug file — Cloudflare strips the header, which I only learned by
firing it at the `020` probe vhost **with an arbitrary control header alongside** to
prove the request itself had carried it. Two wrong mechanism-stories in one morning, on
the same defect, both plausible, both refuted by one cheap measurement each.

Family: verify-the-failing-branch, a-scope-objection-is-not-answered-by-more-evidence
(inverse: here more measurement WAS the answer), writes-the-field-is-not-reads-the-field.

## 2026-07-29 — oufe render-audit agent seed

> Heading added 2026-07-29 by the consolidation thread. This entry was appended
> headingless while I was appending mine directly above it, so it briefly read as
> part of the consolidation entry. **The words below are not mine and are
> unedited** — only this heading was added, so the entry is attributed to the lane
> that actually made the call.

**"The agent row is correct" / "What is NOT the problem: … the agent row" (oufe handoff,
2026-07-29 morning).** The row WAS the problem. The seed wrote `initial_step` where the
chassis reads `start_step`, so every dispatch of `render-audit-agent` was rejected as
`WORKFLOW_INVALID` — and the rejection went to `system.agent.generic.responses`, which no
human reads, so it presented as "consumed, no row, no error". The row had been "verified"
active / not-snapshot / not-deleted / 4 steps — every property EXCEPT the one the code
branches on. Worse, the seed's own VERIFY block read back `initial_step`, so the
verification asserted the wrong key existed, which it did (the recorded
`check-answers-the-question-you-encoded` shape, in SQL). Caught by: reading the response
payload in `chassis_intake_events` for the correlation — the error names the missing key
verbatim.

*Cheap check that would have:* compare the seeded workflow's top-level keys against ANY
working definition — `SELECT type, default_config->'workflow'->>'start_step' FROM
agent_definitions WHERE is_active LIMIT 8;` shows every live agent carrying the key and
the new one NULL, in one screen. The handoff's own prescribed step 1 (untouched-peer
diff) was this check; it was written down and not run before "the agent row is correct"
was.

Family: check-answers-the-question-you-encoded, grep-the-config-key-before-calling-it-a-win,
a-print-statement-is-not-a-config-row (the VERIFY block variant).

---

**"The nav row will be recreated automatically once the page deploys" — written in
`webdesign_couk/HANDOFF_2026-07-29` §1 and §5, marked as the CORRECTION of an
earlier wrong warning, wrong itself (caught 2026-07-29 morning, session 4).** Two
independent errors under one confident sentence. First: nothing runs
`refresh_nav_tables` on a deploy — it runs inside a nav-updater orchestration,
which needs a `nav_drift` work item, and discovery-filed ones can sit at
`detected` indefinitely (robot-hands has one). Second and worse: when a run WAS
forced by hand, the row still could not appear — `isSectionIndexType` omits
`news-index`, so `classifyPagesForNav` classifies the canonical
`/news/index.html` as its own child and skips it, permanently, logging only an
info line. The claim had survived one correction round already, which made it
look tested; the first half had merely never been exercised (the page was never
deployed until now), and the second half was invisible without a deployed page.
Caught by: waiting for the promised row and, when it did not come after a
COMPLETE nav_drift run, reading the nav-updater pod log for the page by name.
Filed + fixed as `bugs_open/141`.

*Cheap check that would have:* before writing "X happens automatically", find the
code path that does X and run its selection query against the row you care
about. Here: `classifyPagesForNav` is one grep from `refresh_nav_tables`, and
the child-prefix list with the three-type exemption is on one screen —
news-index's absence is visible by eye. The general form: an "automatically"
claim is a claim about a MECHANISM, and it needs the mechanism's file:line, not
an inference from a table's name.

Family: writes-the-field-is-not-reads-the-field (nav flags are written, the
classifier never reads them for this type), a-check-answers-the-question-you-encoded
(the earlier session verified the DELETE happened, not that the re-ADD could).

### 2026-07-29 — experience register — I logged the doubt, then acted against it anyway

**The claim.** On 07-28 I re-tiered `template_row_not_a_control` from Tier 2 to Tier 4, reasoning
that its clause is true only post-hydration. **In the same session I wrote, in this file and in the
entry itself, that this was "the reading which makes my own red result disappear" and "the shape of
reasoning I should distrust in myself."** Then I shipped the re-tier.

**What caught it.** The approval council's `deferral_honesty` seat, gating on [high]: *"Moving a
failing check to a tier the platform doesn't execute, rather than reporting the failure, reads as
evasion even with the honest note attached."* Reversed: the check is back at Tier 2, fails
honestly, and `bugs_open/137` now blocks something real instead of sitting as a curiosity.

**The cheap check that would have.** There isn't one, and that is the entry's whole value.
**Recording a suspicion about your own reasoning is not a control on it.** I did everything the
working-docs rules ask — marked it unresolved, flagged it as self-serving, filed the bug for
someone else — and still took the action the suspicion was about, because each of those steps
*felt* like discharging the doubt. The rule that would have worked is procedural, not
introspective: **when you notice that a resolution makes your own failing result disappear, do not
ship that resolution in the same session you noticed it.** Leave the red result standing and let
someone else break the tie. The council was that someone; it should not have had to be.

---

## 2026-07-29 — I proposed building a new mechanism beside a subsystem I had never read, from inside its own directory

Session "bugsearch 7", immediately after closing `bugs_closed/102`. The owner asked whether the
page-type misclassification check should live in the check-and-fix system fired from the
improvement loop. **I had already recommended something else, and the recommendation was wrong
on both of its premises.**

**What I said.** That the natural home — a discovery check in
`platform/orchestration/actions/discovery_checks/` — was unusable because "it would never run:
the improvement-sweep that reaches those checks has been disabled since 2026-05-02". I
recommended a bespoke scheduled SQL sweep instead, raising its own work item type.

**Premise 1, wrong: "the discovery-check lane never runs."** One fact was true — the
`improvement-sweep` scheduled task is `enabled=f` since 2026-05-02 — and I generalised it from
*one check's route* to *the whole subsystem*. The measurement, which I only ran when pushed:

```
completeness-discovery-agent   144 work items   latest 2026-07-25
design-discovery-agent         108 work items   latest 2026-07-25
quality-discovery-agent          7 work items   latest 2026-07-17
```

Two of the three agents had produced 252 items and had run **three days earlier**, fired by
other routes — including a demonstrated one-shot pattern sitting in `scheduled_tasks`
(`oneshot-discovery-aao-20260726`, a task aimed straight at a discovery agent). **A disabled
scheduler is not a dead subsystem.** What is genuinely dead is narrower and I could have said it
precisely: `claims_unverified` items number **zero, ever**.

I even had the sentence that misled me in front of me — CLM-004 says the sweep "effectively
never runs", scoped to that one check's route. I widened it to the package.

**Premise 2, wrong: that a new mechanism was cheap.** The package carries **build-enforced**
invariants I did not know existed: `handler_coverage_test.go` (a check may not route at a
handler agent that does not exist — two live violations found 07-26, noticed by nothing),
`verifier_coverage_test.go` (every `item_type` must be verified or classified with a reason,
enforced by a sensor that reads the package source), and `remit.go` (detector-wider-than-handler
residue becomes one `capability_gap` item). A bespoke sweep with its own item type is invisible
to all three — a second detection mechanism that must be kept in step by memory, which is the
drift class this repo keeps paying to fix.

And the platform's written policy for exactly my situation already existed: **`IMP-016`** — a
discovery check is enabled once its handler exists, observe-only ahead of that. I proposed
routing around a policy I had not read. My check needs no handler at all: it is HITL-terminal,
the same `HandlerAgent: ""` / `needs_human_review` shape as its two neighbours.

**What caught it.** The owner: *"should it be part of our check and fix system that gets fired
from the improvement loop (which is currently switched off) see the files in
discovery_checks/ look in the docs under improvement loop"* — three pointers, all of which I
could have followed unprompted.

**The cheap check that would have.** `ls` the directory. **I had been editing a file inside
`discovery_checks/` all session** — `check_unverified_claims.go`, the post-deploy half of the
very layer I was fixing — and never listed its 69 siblings or opened `registry.go`, `remit.go`
or either coverage guard. Then: `ls docs/agent_docs/docs026_concept_register/register/ | grep
improve` returns `improvement-loop.md`, 47 entries, one of which is the policy. CLAUDE.md names
the register as the thing to consult **before concluding a capability does not exist**, and I
consulted it for the thing I was building (CLM-016) but not for the thing I was proposing to
build *beside*.

**The transferable bit.** *Editing one file in a package is not knowing the package.* I had read
`check_unverified_claims.go` closely enough to modify its query and its call sites, which felt
like familiarity with the subsystem and was not — the invariants live in files I had no reason
to open for my own change. **Before proposing a NEW mechanism, read the OWNING subsystem's
registry, its guards, and its register category** — not the one file you happen to be standing
in. And when the reason for a new mechanism is "the existing one is dead", **measure the
existing one's output** before saying so: a scheduler flag is one hop, and the thing you
actually care about is whether rows are still being produced.

---

## 2026-07-29 — the rest of that session's wrong calls: four more, and who caught each

Session "bugsearch 7", `bugs_closed/102`. Two entries above already carry the summed-glob total,
the paraphrased "verbatim" fixtures and the loop test that could not fail; this is the
remainder, kept together because **the pattern across them is who did the catching.** Two of my
own checks caught two, the council caught two, and none of the four was caught by me reading my
own work.

**1. A verification marker that would have produced a confident false "it is live".**
My RUNBOOK told the next reader to confirm the roll by pod-grepping `section-index`, on the
reasoning that it is a member of the new editorial page-type map. Both halves were wrong:
`section-index` is a string literal in at least four other Go files (`page_growth_budget.go`,
`v3_site_actions.go`, `apply_gap_plan_action.go`, `populate_nav_tables_action.go`), so a match
proves nothing about my change; and I then **removed** `section-index` from the map entirely, so
the documented grep would have returned `1` on a binary that did not contain the fix at all.
*Caught by:* removing the page type, and re-reading the RUNBOOK because of it. *The cheap check:*
before writing a marker down, `grep -rn '"<marker>"' --include='*.go'` — and prefer a **new
function name**, which only your change can put in the binary. `resolvePageType` was 0 on the
live pod before the roll and 2 after, with `scanComponentClaims` (2, unchanged) as the control.

**2. A headline figure assembled from the part of the evidence I had read.**
My PLAN said "47 of 47 editorial findings are false positives". 47 was `blog-post` + `guide` —
the two page types whose snippets I had read when I wrote the sentence. I reviewed `tool`,
`game`, `section-index` and `news-index` further down the same document and never carried them
back into the headline. The class is 61 (and, after the council narrowed it, 59 suppressed).
*Caught by:* the `comm` diff of the before/after finding lists, which is arithmetic rather than
memory. *The cheap check:* derive a headline number from the artefact, never from the reading —
if a figure and a breakdown appear in one document, make the document compute the figure.

**3. I put a page type in the shipped set on analogy, in a change whose entire argument was
"measure it".** `blog-index` went into `editorialPageTypes` because it looked like `blog-post`.
It has three pages fleet-wide and raises zero findings even scanned against an empty register —
no evidence in either direction. My own code comment two lines above says "do not widen this
from intuition". *Caught by:* the council's edit-quality seat, which quoted my own method back
at me. *The cheap check:* for a membership list, require the evidence cell to be non-empty
before the row exists — write the table with the measurement column first and let a blank row
be visibly unearned.

**4. I named a risk and did not check it, in the same document that named it.**
Risk #2 in my submission read: "a site filing marketing copy under an editorial `page_type`
would go prose-unscanned. **Check whether any site does.**" I shipped it as a risk rather than
as a query. *Caught by:* the guardian and compliance seats, both asking for the check rather
than the caveat. The answer took one query and was not "no": **two of the twenty `section-index`
pages are `about-index` and `contact-index`** — marketing bodies under an index name — which
removed a whole page type from the fix. *The cheap check, and the one worth generalising:*
**a risk you can answer with a query is not a risk, it is an unrun query.** Grep your own
submissions for "check whether" and "verify that" before sending; every hit is either work you
owe or a sentence you should delete.

**Tally note.** Of the seven wrong calls in this session, the ones my own harness caught were the
ones I had built a control for — the negative control caught the dead fixtures, the arithmetic
caught the summed glob and the bad headline. **Every misstep with no control attached was caught
by somebody else**, and two of those went to a council that costs credits and thirty minutes.
That is the argument for writing the control first, not the summary first.

---

**"Publish the chrome with a nav_drift item and nav-updater does the rest — the
News row lands with it" — written into `webdesign_couk/SQL_p19`'s own header as
the justification for firing it, wrong (caught 2026-07-29 ~08:00, same session,
by the run itself).** The same session that logged the previous entry — the
handoff's false "nav row reappears by itself" — then ACTED on that claim's
mechanism without reading it, in the very hour it was writing the correction.
The nav_drift completed green, the nav rebuilt WITHOUT News (`bugs_closed/141`:
`isSectionIndexType` omitted `news-index`), and the run queued ~99 assemble
re-renders of which ~98 republished byte-identical pages — then a second full
fan-out was needed once the fix rolled. Cost: hours of queue occupancy, doubled.
Caught by: the completed run's artefact disagreeing with its status (nav table
unchanged), then the nav-updater pod log naming the skip.

*Cheap check that would have:* the one this file's previous entry already
prescribes — before writing (or obeying) "X happens automatically", find the
code path that does X and run its selection against your row.
`classifyPagesForNav` is one grep from `refresh_nav_tables`; the three-type
exemption list is on one screen. **A correction you are writing about someone
else's unverified mechanism claim is not immunity against acting on the same
claim yourself — the entry above and this one are the same skipped check, one
session apart, and the second one had the first open in its editor.**

Family: writes-the-field-is-not-reads-the-field, the entry directly above.

---

**"The build queue has stalled post-roll" — stated as the working diagnosis
2026-07-29 ~08:35 (session transcript, not a durable doc — logged here because
the shape recurs), wrong within minutes.** Evidence read as a stall: 0 claimed,
82 queued, priority-20 item unpicked "for ~10 minutes", trigger "not re-firing".
Actual state: a normal long-running dispatch orchestration (`AWAITING_RESPONSES`
means working, not stuck), the item claimed moments later, and the queue moving
at exactly its measured ~2 min/item baseline — 18 items in 40 minutes. The
"~10 minutes" itself was inflated by wall-clock drift: the session believed
~15 more minutes had passed than had, having printed `SELECT now()` once at
07:40 and never again. Nothing was cancelled (the standing "never cancel the
failing row pre-diagnosis" rule held), so the cost was only wasted diagnosis
down the spawn-drop path.

*Cheap check that would have:* two, both one-liners. Print `SELECT now()`
beside every latency judgement — client clock estimates drift within an hour of
multitasking. And before saying "stalled", divide observed progress by the
measured baseline: 18/40min ÷ 1/2min ≈ on pace, which is the opposite of a
stall, from numbers already on screen.

Family: council-queue-latency-trap ("no rows = QUEUED, not dropped"),
check-an-untouched-peer-in-the-same-batch (the cadence baseline IS the peer).

---

## 2026-07-29 — six wrong calls from the `bugs_open/104` session, and the family three of them share

Session "bugsearch 6", making `banned_claims` fleet-wide. All six are mine. The
"suites green" entry above (2026-07-28) is from the same session and is a seventh.
**Every check named here is now written where the error was made** — the workstream
`RUNBOOK_fleetwide_claim_patterns.md`, gotchas 1–5 and §§8–12 — because a check in a
ledger nobody greps is a paragraph, not a check.

**Three of the six are one family: a measurement that answered a different question
than the one I asked.** That is the tally line worth reading, not the individual rows.

### 1. "Zero findings on all fifteen sites" — I grepped for a string the tool never prints

The fleet dry run reported **0** banned-claim findings on every one of 15 sites. I
counted `grep -c "banned_claim"`. `cmd/claimscan` prints the line prefixes `BANNED`
and `NUMBER`; `banned_claim` is the JSON `check` value and appears **nowhere** in its
CLI output. The real count was **7**, four of them false positives — the entire finding
of the session, invisible.

**What caught it:** the positive control, not review. I fed the scanner nine synthetic
sentences it *must* block and saw `BANNED` lines with my own eyes. Had I skipped the
control, I would have reported a clean estate and shipped the false-positive class.
**The cheap check:** run a known-positive through the tool and read its real output
format before counting anything.

### 2. `2>/dev/null` on a fetch turned a `kubectl` flake into a fact about the estate

My per-site loop suppressed stderr. One `kubectl exec` failed transiently, the output
file was empty, and my table printed **"vonc.com: no register"**. vonc has a current
`evidence_base` row with 9 patterns and 4,651 characters — and vonc was the *one site
whose register mattered most* to the finding.

**What caught it:** it contradicted the § Measurement query I had run ten minutes
earlier. Nothing else would have. **The cheap check:** never suppress stderr on a fetch
whose *absence* becomes a finding; retry, and print `FETCH_FAIL` as a distinct state
from `no-row`. An empty result and a failed request must not render identically.

### 3. I wrote "fires on nothing fleet-wide" into a bug file before running it

Recording the narrowed 9-pattern set in `bugs_open/104`, I wrote that it "measurably
fires on nothing fleet-wide". Then ran it: it fires **once**, on a true positive. The
difference mattered — it meant the option was not free, it landed a blocker on a live
leopardess page. Corrected in place with a visible correction block.

**What caught it:** running the command I had already described the output of. **The
cheap check:** the word "measurably" is a promise; if the measurement has not run,
write `[UNMEASURED]` instead.

### 4. Two regression fixtures were sentences no site had ever published

I built the negated-disclosure fixtures by retyping from `claimscan`'s output, whose
snippets are elided with `…`. Two came out as plausible inventions: vonc's real sentence
begins *"Competitor characterisations reflect general platform mechanics…"*, not
*"These reflect platform mechanics…"*. The test comment said "real copy taken from a
live site". **The tests passed either way** — they assert a negated sentence is not
flagged, and a paraphrase negates too — which is exactly why it was worth fixing rather
than shrugging at.

**What caught it:** checking `grounded_in` quote fidelity for the council submission,
then decoding the component base64 to get the sentence. **This is the second instance of
this row in one day** — session "bugsearch 7" logged the same defect (entry
2026-07-28, "two 'verbatim' fixtures that were paraphrased") while working the sibling
bug. Two independent sessions, same tool, same class, same afternoon.

### 5. My first gate test could not fail in the direction that mattered

Testing that the build gate scans an unarmed site, I asserted on the returned `issues`
list and treated a returned error as a test failure. On any blocker the action returns
`(nil, error)` — **the error IS how the build fails.** So the test reported FAIL on the
one outcome the bug wants, and had I "fixed" it by loosening the assertion, it would
have passed on a working *and* a broken gate, because the issues map is nil on that path.

**What caught it:** reading the failure message — `content validation failed: 1
blockers, 0 errors` — which was the gate working correctly. **The cheap check:** before
trusting a test, ask which line fails if the feature is removed. Same shape as the
`psql -t -A` guard that could not fail, from this session's predecessor.

### 6. THE ONE I DID NOT CATCH: I scoped a measurement by a column the code never reads — while correcting someone else's version of the same error

The load-bearing claim of the whole change was "0 findings across 908 live components".
Round 1 of the council ran its own check and returned `count 0` for that population,
because it filtered `sites.status='live'` — a value that does not exist in this estate.
I caught that, corrected it to `status NOT IN ('pool','archived')`, got 908, and felt
pleased.

Round 2's gating objection (`debug_historian`, HIGH) was that I had **"replaced it with
a different status-based exclusion rather than dropping status as a scoping variable
entirely"** — and that nothing in the build-gate path filters on `sites.status` at all.
There were **17 pool-status sites against my 15 measured**, so the slice I had silently
excluded was *larger than the slice I measured*.

Re-run with status dropped and **grouped** instead of filtered: `deployed | 908 | 14`
and no other row — pool and system sites hold zero components with stored
`rendered_html`, so the exclusion happened to cost nothing. **The answer was favourable
and the objection was still correct.** I had inherited that filter from `104`'s own
§ Measurement query, where it is right for *"which live sites are armed"* and wrong for
*"what will the gate fire on"*.

**What caught it:** a reviewer, at HIGH severity, two rounds in. Not me, and not any
check I would have run. **The cheap check:** when measuring an enforcement surface,
`GROUP BY` the variable you were about to filter on, so the excluded slice appears as a
row instead of vanishing. And: **a filter inherited from another document is still your
filter** — re-derive it against *your* question before quoting the total.

**Tally**, per the house convention:

- "run a census against a known-positive control before reporting the count" **2→3** (row 1)
- "measure a property before describing it" **1→2** (row 3)
- "verify an embedded/quoted artifact is COMPLETE before asserting it" **2→3** (row 4) —
  **its second instance in one day**, by a different session on the sibling bug, with the
  same tool. Two independent hits on one afternoon is the signature this file exists to
  surface: `claimscan`'s elided snippets are a quoting trap, and the fix belongs in the
  tool (print a copy-safe full sentence, or refuse to print a truncated one) rather than
  in two sessions' good intentions.
- three NEW rows added: stderr suppression on a finding-bearing fetch; a test that cannot
  fail in the direction that matters; `GROUP BY` the variable you were about to filter on.

### The honest read-out

Five of six I caught myself, and four of those only because something forced a second
look: a positive control, a contradicting query, a failure message, a fidelity check for
someone else's benefit. The sixth — the biggest, sitting under the headline number of
the entire change — I did not catch, and neither would any check I had. It took an
adversarial reader who asked *what is this filter for* rather than *is this filter
correct*.

---

## 2026-07-29 — TWICE IN ONE DAY: what gets re-measured is whatever a reviewer challenged

**The claims.** Two, both mine, both in the same 12 hours, both the same mechanism.

1. Round 8 of a council submission cited *"0 markdown, **4992** total"* for the code
   index. It was **5,017** — and it was already 5,017 when I submitted, a reindex
   having landed 19 minutes earlier. I carried the figure forward from round 7
   without re-running it.
2. The handoff document that the next session cold-starts from stated
   **"`review_architecture` — still 0 reviews"**, in its state table *and* its
   open-items list. That was true when I wrote it at ~22:00 and **false by 22:10**.
   It is the oldest open question in that workstream and its literal headline.

**What is common to both, and it is not laziness.** In the same submission as (1) I
*corrected a different stale figure at source* and wrote that leaving one in the
evidence is exactly the drift this workstream reviews for. I was re-measuring
carefully. **I re-measured every claim a REVIEWER had challenged, and nothing else.**
The council's objections silently defined the scope of my re-verification. Nobody
had queried 4,992, so it was never re-run. Nobody challenges a handoff at all, so
nothing in it was re-run.

**A figure nobody objects to is precisely the one that rots quietly**, because the
only trigger for re-checking it — someone pushing back — never fires.

**How each was actually caught.** Neither by checking. (1) fell out of re-grepping a
pod after an unrelated deploy. (2) fell out of reading an unrelated line in the
memory index during a compaction task. **Both were accidents**, which is the whole
problem: there was no path by which either would have been caught on purpose.

**The cheap check.** *A state table is a set of claims, and each claim has a query.*
Before publishing a state table or a `grounded_in` block, re-run the query for every
row — not the rows under dispute. For the handoff it was four seconds:

```sql
WITH r AS (SELECT da.body::jsonb->>'decision' AS d, rev FROM diagnosis_artifacts da,
           jsonb_array_elements(da.body::jsonb->'reviews') rev WHERE da.kind='council_report')
SELECT count(*) FROM r WHERE rev->>'reviewer'='architecture';
```

**The recurrence is the finding, not either instance.** One is an anecdote; two in a
day, with the second landing in the document whose entire job is to be true at
cold-start, says the trigger for re-verification is misplaced. It is currently
*external* (someone objected). It has to be *structural* (this is a state claim, so
it gets its query re-run before publication).

Distinct from *"a count you KEPT is not a census"* (same day): there the tool was run
and its output misread. **Here the tool was never run at all, because nothing asked
me to.**

## 2026-07-29 — "14 orchestrations reference it" counted the check's NAME, not its runs

**The claim** (`bugs_open/131`, gauntlet slug, §"no manual dispatch is needed"): the
`no_horizontal_overflow` clause "is on an actively exercised path, so this will
happen on its own — 14 orchestrations reference it, most recently 15:17:51Z". Written
to justify waiting for a natural witness instead of firing one.

**What was false.** Classified by agent type, the referencing orchestrations are
council-gate reviews and experience-register writes whose *text contains the check
spec* — the 15:17:51Z row is an `experience-register-writer` doc write. The actual
execution lane (`acceptance_run` work items → `tool-acceptance-agent` →
`request_browser_run`) has fired **four times ever**, zero since the fixed adapter
rolled. "Wait for a natural run" meant days, not the hours the sentence implies —
and a session that believed it would have parked the witness indefinitely.

**What caught it.** The next session (vonc 6) ran the reference query, then asked
what *produced* each row before counting them — one `GROUP BY agent_type` collapsed
14 references to 0 executions.

**The cheap check that would have.** This is [prompt-text-poisons-its-own-detector]
wearing acceptance-lane clothes: a `LIKE '%<name>%'` over orchestration state counts
every document that *talks about* the thing. Before reading a reference count as an
activity count, group it by what wrote the rows — or query the lane's own artefacts
(here: `site_work_items WHERE item_type='acceptance_run'`, four rows, seconds).

**Resolution.** Corrected in place in `bugs_open/131` (dated, visible); witness fired
manually as work item `4e06c4ab` on the 010b manual precedent.

**"Sourcing each observation from its own contemporaneous Ofwat report is the strongest
possible provenance" (oufe timeseries planning, 2026-07-29).** The premise was refuted by
the first cross-check: Ofwat's 2022-23 report records Thames leakage at −10.7 where the
company's 2024-25 APR chart shows 11.1 for the same year — Thames RESTATED its history
after 2023-24 methodology improvements, so the "strongest" per-point sourcing would have
mixed two measurement methodologies inside one series and produced a chart that looked
consistent and quietly wasn't. Caught before anything shipped, by cross-checking one
overlapping year across the candidate sources.

*Cheap check that would have (and did):* **before choosing per-point sources for any
series, cross-check ONE overlapping period across all candidate sources.** Provenance
quality is not additive across documents — five impeccable citations can still assemble a
series no single source would publish. The failure class is invisible to every
per-observation check, because each point individually verifies.

Family: survey-the-premise-before-building; a new sub-family worth naming —
**per-point verification cannot see a cross-point inconsistency.**

**Mig 265 was written, applied and COMMITTED with no replay guard (2026-07-29).** The
implicit claim — a supersede-then-insert migration file is a safe replayable record — was
false, and in a sharper way than the recorded pattern (`bugs_open/007` Class C) names: a
replay would not die on 23505, it would `||`-append the four facts a SECOND time,
silently, leaving 40 facts where 36 belong and every downstream count wrong. Caught by
the commit hook's `unguarded-migration-insert` advisory, post-commit. The 254 file this
was modelled on has the same defect; copying an applied migration's shape copies its
replay behaviour.

*Cheap check that would have:* before applying any supersede-then-insert, ask "what does
the SECOND run do?" — and if the answer involves `||` on a jsonb array, the failure is
silent duplication, not a loud constraint error. Guard both statements on the new id
being absent, and PROVE the no-op by actually replaying (UPDATE 0 / INSERT 0 / count
unchanged — 30 seconds against the live DB).

Family: verify-the-failing-branch (the replay IS the failing branch of a migration);
a-doc-comment-is-not-an-enforcement-mechanism (the hook caught what the 007 write-up
alone did not prevent).

**Two same-day instances of recorded families, logged for the tally (2026-07-29):**
(1) `pages_failed: 1` in the render-audit summary was read as "one page failed to
load" and a `failed_pages` key was queried that does not exist — the struct comment
defines it as "pages with a *non-approximate* contrast failure" and unreachable pages
land in `unreachable`. A new instrument's field names are not their plain-English
meanings; read the struct before interpreting the first result
(family: check-answers-the-question-you-encoded, instrument-field variant).
(2) A consumer-group listing against a guessed Kafka pod name returned nothing and was
briefly read as "no matching groups" — the NotFound error had gone to `2>/dev/null`.
Zero results with a suppressed stderr is the recorded grep-goes-silent shape wearing
kubectl clothes; drop the suppression before believing an empty result
(family: grep-silent-on-non-utf8 / a-check-answers-what-you-encoded).

## 2026-07-29 — I deployed a site asset to the wrong deploy repo, then invented a lagging origin to explain it (bugfix_131_og_card, session relojistas-5)

**The claim:** relojistas.com's live header kept serving the old logo after a green
"Deploy to B2" run and a successful Cloudflare purge, so I concluded the serving chain
is "CF → an intermediate origin that pulls from B2 on its own cadence", inferred from
the nginx-style etag on the response.

**What was actually true:** there are TWO deploy repos and the site picks one —
`sites.github_repo`. relojistas.com is `vm-sites` (nginx on a box, `gqls/vm-sites` +
`deploy-to-vm.yml`); I had published to `gqls/sites`, the B2 route, which for that
domain is a dead duplicate folder that nothing serves. There is no lagging origin. The
correct write went live and was verified by eye in about two minutes.

**What caught it:** wondering *which* origin, and running
`SELECT domain, github_repo FROM sites` — a query I could have run before choosing a
route, since the column exists precisely because there are two.

**The cheap check that would have prevented it:** ask the DB which repo serves the site
BEFORE publishing. One query, no cluster access needed.

**Why it is worth a row:** the failure is silent by construction — the wrong repo
contains a same-named `<domain>/` folder, so the write succeeds, the workflow runs
green, the purge reports success, and nothing changes. Every mechanical signal was
green and the artefact was untouched, which is this lane's own recurring lesson
arriving through a completely different door.

**The near-miss that made it cheap:** I had marked the origin claim `[INFERRED]`. That
marker is the only reason it did not go into the RUNBOOK as fact — the discipline is
working, and it is worth noting that the marker paid off on the very first durable
claim I made after writing it.

Family: check-answers-the-question-you-encoded (the check asked "did the write land in
the repo I chose", never "is that the repo that serves this site");
a-print-statement-is-not-a-config-row (a green workflow run vouches for itself, not for
the site); narrow-filter-defines-the-conclusion (one route assumed, two exist).

---

## 2026-07-29 — "charset-guarded identifiers, quote-doubled free text" — offered as a safety feature

**Where I wrote it.** The council submission for the asset-amend path, describing
`scripts/amend-asset.sh`: the loader built its SQL by interpolating shell variables into the
statement text, and I listed the escaping around that as one of the edit's merits.

**The truth.** The ALWAYS-ON parameterisation rule prohibits building SQL by interpolation
**regardless of how well it is escaped**. The `constitution` seat called it flatly, and was
right: my guards were sound as far as they went (a charset class on identifiers, doubled
quotes on free text) and entirely beside the point. Fixed in `048dbd96b` — every operator value
now travels as a psql variable, including the server-side guard's domain via
`set_config`/`current_setting`.

**What caught it.** The council gate, round 1. Nothing in my own testing could have: the
dry-run output looked correct, the escaping worked, and no input I would ever have typed
breaks it.

**The cheap check that would have.** Read the rule before writing the file, not after the
objection. The rule is ALWAYS-ON — it is in the standing set precisely so it does not need
re-deriving per case.

**The transferable bit, and it is the uncomfortable part.** *I did not merely break the rule —
I advertised the workaround as a feature.* Escaping discipline presented as a merit is the
signature of having chosen the wrong construction and then defended it, and it reads as
diligence, which is what makes it survive review by anyone reading quickly. **Tell: if an edit's
rationale is proud of its escaping, the escaping is load-bearing, and it should not have been.**
The same shape shows up wherever a submission's rationale explains how carefully something
unsafe is handled, rather than why the unsafe thing is absent.

*Second, smaller, same round:* I invented `source='operator'` for a work item where
`018_site_work_items.sql:18` sanctions `'manual'|'improvement'|'side_effect'`. The live table
carries 8 rows at `'operator'` from other lanes, so I could have cited precedent — but
**precedent in the data is not permission in the rules**, and the drift is exactly how a
vocabulary stops meaning anything. Conformed rather than ratified.

### The tally point

Both entries this session are the same underlying skipped check: **the standing rule was not
re-read before the work that engaged it.** The 07-28 entry above missed an entire audience
because no check encoded it; this one missed a rule that was already written down. Those need
different fixes — the first needs someone to notice the gap, the second needs nothing but
looking. Cheapest general guard for the second class: when an edit touches SQL construction,
credentials, or a vocabulary column, open the standing rule for that thing *first* and quote it
in the rationale. A rationale that quotes the rule cannot also be proud of working around it.

## 2026-07-29 — gauntlet_dead_cta (vonc6): "the inner-page census" that measured four homepages

**The claim I nearly wrote:** "og:url = site root is vonc-only — other fleet
sites set it per page." I ran a census across four sites to check, and it came
back clean-looking.

**What was actually wrong:** my script found each site's "inner page" by
grepping the homepage for the first `href="/…​.html"`. On all four sites that
first match was **`/index.html` — the homepage itself**. So the census compared
each site's homepage against its own homepage, where `og:url` = root is
*correct*, and would have licensed exactly the wrong conclusion.

**What caught it:** reading the output rather than the verdict — the same path
`/index.html` appearing in all four rows is not what four different sites'
inner pages look like.

**The cheap check that would have:** print the thing you selected, not just the
result you computed. One glance at the four identical paths ends it. Corollary
to [[narrow-filter-defines-the-conclusion]]: the filter here didn't narrow the
world, it **silently selected the control group as the treatment group**.

**Re-run properly:** pulled real inner pages from `pages` (`url NOT IN
('/','/index.html')`, `deployed_at IS NOT NULL`) and re-measured — 7 pages, 4
sites, og:url = root everywhere, canonical absent everywhere. The finding held;
the first instrument simply could not have found it either way.

## 2026-07-29 — I inferred a live path from the existence of built machinery (gauntlet_dead_cta, vonc 6)

**The claim, written into a handoff:** having watched my 131-B witness produce a
real `improve_tool` item, I wrote *"if it cycles without fixing, that is
`bugs_closed/010`'s guard — expect escalation at `fix_cycles_spent=2`."*

**What was false:** it never cycles. `judge_acceptance_results` hardcodes the
item's status to `'detected'`; the dispatch loop reads only `triaged`/`approved`;
the sole promoter (`triage_detected_items`) lives inside the `improvement-sweep`
scheduled task, **disabled, last triggered 2026-05-02**. Every `improve_tool`
item filed since 2026-07-17 — 7 of 7 — is parked. So the fixer, the convergence
guard and the escalation are all real, tested, deployed **and unreachable**.

**Why I believed it.** I had *read the machinery*: 010's guard is well documented,
the escalation path is unit-tested, `handler_agent` on the row literally says
`tool-improver`. Every artefact I looked at was genuine. **None of them was a
reader running on a schedule.** This is
`[[writes-the-field-is-not-reads-the-field]]` one level up: there I would have
needed the reader's file:line; here I needed the reader's *trigger*.

**What caught it.** Curiosity, not a check — I went to watch the fixer work and
found the row hadn't moved in an hour. Had I not looked, the wrong sentence would
have shipped in a cold-start handoff, where the next thread would have waited for
an escalation that cannot arrive.

**The cheap check.** After writing *"the system will now handle X"*, name the
component that reads X **and confirm something calls it on a cadence** — one
query: `SELECT enabled, last_triggered_at FROM scheduled_tasks WHERE
target_agent_type = '<reader>';` A disabled row and a busy one look identical
from the producer's side, and "the handler exists" is not "the handler runs".

**Not a new bug:** already filed and correctly diagnosed as `bugs_open/083`
**by slug** (`…detected_findings_never_reach_a_handler`). Grepping before filing
turned a would-be duplicate into a contribution — the acceptance ladder was a
consumer that file had not listed, and the pile has grown 157 → 250 since 07-27.

**"Tool pages are EXCLUDED from the orphan check — the exact case here" (oufe, 2026-07-29).**
Said out loud, from reading `check_orphan_pages.go:204` — `NOT IN ('blog-index','tool')
OR in_header OR in_footer` — and stopping at the exclusion without reading the `OR`, and
without querying the flags. **Every one of the 11 unreachable tool pages fleet-wide carries
a nav flag, oufe's included (`in_footer=true`), so the exclusion applies to none of them**
and the check would have flagged all 11. Retracted one tool call later, by the census I was
running anyway. Had the census not been the next step, a wrong cause ("the check excludes
tools") would have gone into a bug file — and it is a plausible-sounding cause that would
have sent the next thread to change the SQL instead of scheduling the agent.

*Cheap check that would have:* when a predicate has an `OR` escape, **run it against the
real rows before describing what it does** — the flags were one query away and they invert
the reading. A clause read is not a clause evaluated.

Family: check-answers-the-question-you-encoded; a-count-you-kept-is-not-a-census (this is
its inverse — a *condition* kept in the reading and dropped from the evaluation).

## 2026-07-29 — "template edits are fleet-shared" (brochure_component_library)

**The claim.** Presenting an em-dash cleanup choice to the owner, I wrote that
editing the three component templates would change other sites too, and made that
a stated trade-off of the option he was choosing between.

**What was true.** Those components have a **separate row per site**:
`tool-llm-cost-calculator` alone has four ids across fundamentallyai,
ai-agent-orchestration, finetuning.uk and leopardess. Blast radius was one site.
I put a scary and false consideration in front of a decision.

**What caught it.** Measuring before editing — a `count(DISTINCT s.domain)` grouped
by `content_components.id` on the way to writing the UPDATE.

**The cheap check that would have.** The same query, before writing the question:
`SELECT cc.id, count(DISTINCT s.domain) FROM content_components cc JOIN
page_components pc … GROUP BY cc.id`. **A component is shared or copied as a matter
of fact, not of category** — "it's a shared library" is a description of the table,
not of the row.

**Second error in the same breath: "21 template em-dashes" was three populations.**
4 were inside CSS comments (invisible), 2 were table cell placeholders meaning "not
applicable" (correct typography), and only 15 were visible prose. An earlier
correction in this lane had already split writer prose from template text; the
character count needed splitting again. *A character count is not a style
measurement — grep the CONTEXT of each hit before quoting the total.*

## 2026-07-29 — "do not write a check; schedule the agent" (oufe / bugs_open/146)

**The claim.** Having found 11 unreachable tool pages and traced the detector
(`check_orphan_pages`) to a discovery agent that has raised nothing automatically
since 2026-07-17, I wrote the finding up as a pure cadence problem and put
**"Do not write a check; schedule the agent"** at the top of the fix candidates.
I filed it in `/bugs_open/` in that form.

**What was true.** Scheduling the agent would have detected all 11 and fixed none.
The check classifies a nav-flagged page as `nav_drift` and routes it to
`nav-updater` → `populate_nav_tables`, and that action **skips every `/tools/` URL
by design** (`populate_nav_tables_action.go:294,339` — "the parent /tools.html …
represents them in navigation"). The handler completes and changes nothing. The
proof was already in the database when I wrote the claim: a `nav_drift` item raised
for a tool page on 2026-07-24 is `complete`, and that page has had **0 nav items
and 0 chrome links** ever since. Fleet-wide, **2 nav items point at a tool page out
of 95 deployed tool pages** — a number that alone should have stopped me.

**What caught it.** The owner asking a question I had not asked: *why were these
unreachable when they were created?* I had explained why nobody NOTICED, and
mistaken it for why it HAPPENED.

**The cheap check that would have.** Read the handler before recommending it. I had
already read the check that RAISES the item and never opened the action that
CONSUMES it — one `sed` on `populate_nav_tables_action.go` was the whole distance.
**A work item's existence says a defect was detected; it says nothing about whether
its handler can act on it. Before proposing "schedule the detector", follow one
completed item of that type to the artefact and confirm something changed.**

**Related error in the same file, same day.** I wrote that "every one of the 11
carries a nav flag", presenting it as evidence the pages had *declared* they belong
in nav. They had not: `pages.in_header` and `pages.in_footer` **default to `true`**
in the schema, and one of the two creators omits both columns from its INSERT. The
flag was a column default, not a decision — and it is what routes these pages into
the branch that cannot fix them. *A boolean that is true on every row is a default
until you have checked the DDL; `information_schema.columns.column_default` is the
one-line check.*

Family: grep-the-config-key-before-calling-it-a-win (a mechanism that looks live
and is inert); zero-adoption-means-read-the-mechanism (I invoked this pattern by
name and then applied it only to the detector, not to the handler).

## 2026-07-29 — "all nine built pages clean" — a phrase-list sweep reported as a judgement about the pages (dartsonline_traffic)

**The claim.** Session 1 of the dartsonline traffic workstream ended with a sweep
over `page_components.rendered_html` for the nine built pages, printed as nine
lines each reading `clean`, and I reported it as: the shop language is gone from
every built page. I committed that as the session's closing state
(`d57686ab9`) and repeated it in the SUMMARY.

**What was true.** `/sale.html` was serving, and still serves as I write this:

> "We cut prices across our sale range."
> "We move high-density tungsten barrels, shafts, and flights into clearance
> regularly."
> "Testing a new grip profile or barrel weight costs less when you shop the sale
> section."

The sweep was not broken and it did not lie. It tested a fixed list of banned
phrases — `stock`, `Add to Bag`, `filter our ranges`, `Portland`, `darts.com`,
the ones from the three fabrication sites I had just fixed. `clearance`,
`cut prices`, `sale range` and `shop the sale section` were not on it, so the
row printed `clean`. **The sweep answered the question I had encoded. I reported
it as the answer to the question I had asked**, and those were not the same
question — the first is "does this page contain the phrases I already knew
about", the second is "does this page still read like a shop".

**What caught it.** Reading the served page, an hour later, for an unrelated
reason. Not a check — an accident.

**The cheap check that would have.** `curl` the page and read it. Ten seconds per
page, nine pages. The stored-HTML sweep is the right tool for *regression* on
phrases you have already found, and it is the wrong tool for *discovery*, which
is what I was actually claiming to have done. A grep can only ever confirm the
absence of what you thought of.

**The stronger form, because this is the second time in one session.** The same
list-shaped blindness had already bitten me twice that day in the other
direction: `data::text ILIKE '%stock%'` matched my own new `honesty_rails` text
("Never claim to stock…"), and `%Add to Bag%` matched `cta_style.never_use`. I
noticed both, fixed the query, and did not draw the general lesson — that I was
reasoning about a *string list* while writing sentences about *pages*. **When a
check is a list of literals, the finding is "none of these literals appear",
and that is the only sentence it licenses.** Write it that way in the doc, and
the gap between it and what you wanted to say becomes visible instead of
invisible.

**A fourth home found by the same reading.** `site_plan_pages` — the rows a
reconcile rebuilds pages FROM — still carried `"Darts Online | Specialist Darts
Equipment & Accessories"` and `"About Darts Online | Specialist Darts Retailer"`.
I had corrected `identity`, `briefing`/`classification`, `content_direction` and
per-page `page_spec.purpose`, called it three homes, and the plan would have
restored the lie on the next reconcile. *Fixing every reader of a false premise
is not the same as fixing its source; ask which table REGENERATES the ones you
fixed.*

Family: check-answers-the-question-you-encoded; narrow-filter-defines-the-conclusion;
curl-audit-has-no-opinion-about-rendering (inverted — here the DB had no opinion
about what was served); a-pass-from-a-blind-check-outlives-the-blindness (the
`clean` sweep was already quoted in a commit message and a SUMMARY before it was
caught).

## 2026-07-29 — "the live header serves stale chrome, so the 404 links remain" (dartsonline_traffic)

**The claim.** Written into the same closing commit as a caution for the next
thread: *"the nav fix is DATA-ONLY … the live header still serves stale chrome
(bugs_open/117), so the three 404 links remain on the served pages until a chrome
rebuild runs. Do not report the nav as fixed until curl says so."*

**What was true.** Curl says the dead links are on **four** pages, not on all of
them, and the four are exactly the pages that have not been rebuilt since the nav
data was corrected — `/sale.html`, `/new-arrivals.html`, `/guides/index.html`,
`/contact.html`. Every page rebuilt that day came out clean **without anybody
touching chrome**, because the header is regenerated per page at build time and
`GetNavItems` already prunes never-deployed targets. `bugs_open/117` (chrome is a
stored artefact that no page re-render rebuilds) is a real bug and it is not this
one.

**What caught it.** Doing the thing my own caution told the next thread to do —
`curl` — instead of reasoning from the stored `site_components.header`, which
does contain all three dead links and is not what the pages are serving.

**The cheap check that would have.** The one in my own sentence. I wrote "do not
report the nav as fixed until curl says so" and then reported the *diagnosis*
without asking curl either. **A caveat is not a measurement.** Attaching a
correct warning to an unmeasured claim makes the claim read as careful rather
than as unchecked, which is worse than stating it flatly.

**What it cost, and what it would have cost.** Nothing, because it was caught
before acting: the fix I had queued up in my head was a chrome rebuild, which
would have rebuilt `site_nav_items` from *stale* `pages` rows and produced a
header still missing Guides. The real order is nav-table rebuild first, then the
four page rebuilds. One curl separated a wrong fix from the right one.

Family: verify-the-failing-branch; measure-rendered-property-in-the-renderer;
a-doc-comment-is-not-an-enforcement-mechanism (a caveat is not a control either).

---

## 2026-07-29 — NEAR MISS (caught before it was written): "no council seat has ever had its token cap raised" — off my own wrong-depth JSON path

**Thread** "bugsearch 5", working `bugs_open/138`.

**The claim I was about to write.** That candidate 3 (right-size `max_tokens` per
seat) was barely started, because a query returned `(unset→default)` for **all 17**
council seats — a clean, uniform, decisive-looking answer. I was one paragraph from
putting it in the bug file, where it would have justified a whole line of work.

**Why it was false.** I queried `s.value->'config'->>'max_tokens'`. The cap lives at
`config.ai_service.max_tokens`. **A JSON path at the wrong depth does not error —
it returns NULL for every row**, which my `COALESCE` then rendered as a tidy
"(unset→default)". Four seats were already at 16000, including `editquality`, the
single worst offender in my own 14-day table. The true state was the opposite of my
claim: candidate 3 is nearly done for every seat that has actually truncated.

**What caught it.** Not method — **luck**. The answer contradicted something I
happened to know first-hand (I had raised `architecture` myself that morning and
seen `max_tokens=16000` in `llm_call_log`). With any seat I had not personally
touched, the wrong answer was unfalsifiable from the inside.

**The cheap check that would have.** `jsonb_object_keys` on ONE object before
querying a path into it — three seconds. Or: treat a uniform result across a
heterogeneous population as a smell. 17 independently-configured seats agreeing
exactly is either a fleet-wide default or a broken query, and both are worth one
confirming query before they become a finding.

**Why this is logged even though nothing was published.** The tally is the point,
and this is the same silent-plausible-answer family as
[grep-the-config-key-before-calling-it-a-win] and [a-count-you-kept-is-not-a-census]:
a check that answers a different question than the one asked, cleanly, with no
error. It is now a landmine on the register entry (FIX-055) and in
`bugfix_138_degraded_gates/RUNBOOK_degraded_gates.md` §4.

**Second, smaller, same session.** I read `go vet` failing on
`datahelpers/claims.go: undefined: negatedClaimMatch` as possibly my own breakage.
It was another session's uncommitted WIP — `git archive HEAD` built clean. The
session-start `git status` I was working from did not list the file, because it is a
snapshot and stale within minutes. **In a shared tree, a red build is not evidence
about your change until you have separated HEAD from the working tree.**

> **CORRECTED 2026-07-29, same day, by the thread that wrote the entry above.** I
> filed that as a novel near-miss. **It is a REPEAT of a trap already written down
> in `016b` on 2026-07-20** — same file, same field, same wrong conclusion, nine
> days earlier:
>
> > "(b) A NULL from a JSON path query is not evidence of absence: the per-seat cap
> > lives at `steps.<seat>.config.ai_service.max_tokens`, and a query missing
> > `->'config'` read all 13 seats as 'unset'."
>
> (`016b_debugging_guide_8_consolidated.md:1385`. It read 13 seats then; 17 now.)
>
> **What actually went wrong is therefore not the query — it is that I did not grep
> the guide before trusting a uniform result.** "Grep before you file" is written for
> BUGS; I had not applied it to a *measurement* I was about to build a conclusion on.
> The cheap check was `grep -n 'ai_service.max_tokens' docs/.../016b*.md` — two
> seconds, and it hands you the answer with the reason attached.
>
> **This is the tally doing its job.** One occurrence is an anecdote; the same trap
> twice, on the same field, in the same subsystem, by two different threads, is the
> signal that it should stop depending on memory. Deliberately NOT adding a fourth
> §9 entry for it — a third restatement of a documented trap is not what is missing.
> **What is missing is that nothing makes the wrong query fail.** Candidate, recorded
> here rather than built unasked: a `sql_for_agents` helper (or a pattern-check rule)
> that reads a seat's cap by name, so no thread hand-writes the path again — the same
> reasoning that earned `check_append_only_docs` its place in `scripts/pattern-check.py`.

---

## 2026-07-29 — bugsearch 6 — three in one session, all on the same change, all caught by measurement rather than by me

Arming the negation guard for the fleet-wide claim set (`bugs_closed/104` follow-up,
commit `116fdffd8`). The change is sound; these are the three claims I made along the
way that were not.

**1. I narrowed a pattern to dodge a false positive I had invented, and made it
inert.** I shipped the restored external-verification pattern subject-anchored
(content noun + `is/are` nearby) to avoid flagging "our accounts are independently
audited". Dry run over 919 live components: the anchored form matched **nothing at
all**, while the bare form found **2 real overclaims** and the guard suppressed the 4
false positives. *The cheap check that would have:* run the pattern you are replacing
next to its replacement, in the same run — one extra invocation. *The tell:* my
justification was a sentence I made up, not one I found in the corpus. Now 016b §9.

**2. I mis-added a column of numbers to "confirm" a prior figure, and recorded the
corpus as unchanged when it had grown.** The per-site component counts summed to
**919**; I read them as matching yesterday's 908 and wrote "matches the runbook, so
the corpus is unchanged". Yesterday's 908 was correct yesterday. *The cheap check that
would have:* one `count(*)` query instead of mental arithmetic over a 14-row table —
and the general form, **never confirm a prior figure by re-deriving it the hard way;
re-run the query that produced it.** Same family as the reflex that keeps stale
premises alive, except here I actively manufactured the agreement.

**3. A "verbatim" fixture list contained a sentence that does not belong to it.** The
07-28 dry run reported **4 findings**; I read that as 4 sentences and recruited a
nearby sentence containing the same phrase to make up the number. There were **3**
distinct sentences — one site's was counted twice, on two components. The recruited
one contains no negation at all and is in fact an assertion, so it now sits in the
must-block list. *What caught it:* not review — a new test that asserts the pattern
still matches the fixture. It could not have been caught while the pattern was
excluded from the set, which is the point of the 016b §9 entry it produced. *The cheap
check that would have:* **findings ≠ sentences.** A per-pattern scanner reports once
per pattern per block, so a count of findings is a count of (pattern × component)
pairs. Read the distinct snippets, not the total.

**What the tally says.** All three are the same failure at different altitudes: I
believed a number or a sentence I had produced by reasoning, when the corpus was
sitting right there and answering in one command. The two that reached a file were
caught by a *test* and a *dry run* — not by re-reading my own work, which I did
several times in between.

---

## 2026-07-29 — webdesign.co.uk tools repair: I confirmed a finding against a page I invented

**The claim.** A new static check (orphan element references) flagged 10 pages
fleet-wide. I "independently confirmed" one of them, `css-filter-playground`, in
a real browser: six sliders, all `MISSING`, `rangeCount: 0`. On that basis I
wrote — into a commit message, a concept-register entry, a Go file header and a
**live council submission** — that the check finds what a browser probe misses,
with that page as the worked example.

**It was false, and so was the confirmation.** I fetched
`/tools/css-filter-playground.html`. The URL the database gives is
`/tools/css-filter-playground/index.html`. The first returns **404**, so I ran my
browser evidence against Chrome's error page and read "no sliders here" as a
finding about the tool. On the real URL: `rangeCount: 9`, all six ids present.
The tool has never been broken. The check had a genuine false-positive class —
the page builds its sliders with `id="${f.name}"` from a data array, so the ids
exist in the browser and nowhere in the source.

*What caught it:* not review, and not re-reading — I had re-read that evidence
twice while quoting it into three documents. It fell out of a **different**
question: two other flagged pages appeared to 404, which was implausible enough
that I finally read `pages.url` out of the database instead of assuming its
shape. The same mistake had been silently corrupting the evidence all along.

*The cheap checks that would have caught it, in order of cheapness:*

1. **`%{http_code}` on every fetch you draw a conclusion from.** The 404 was in
   my terminal the whole time, as a `LOG:` line under a `VALUE:` line I was
   reading. One `curl -o /dev/null -w '%{http_code}'` makes it unmissable.
2. **Never construct a URL you can read.** `SELECT url FROM pages` is one query.
   This platform serves at least three URL shapes and the bug file for that
   (`bugs_open/029`) is about a tool cross-linker that *constructed*
   `/tools/{function}.html` and matched no page on any of them. I re-made a
   documented mistake.
3. **A negative browser result deserves a positive control.** "The sliders are
   absent" and "the page is absent" produce the same reading. Asking for
   `document.title` in the same expression would have returned Chrome's error
   page title and ended it instantly.

**The generalisation worth keeping.** My own NOTES already said *"a negative
verdict is a hypothesis until the page has been asked directly"* — and I did ask
the page directly. The rule was not wrong, it was **incomplete**: asking the
page directly is worthless if you cannot show you asked the right page.
Make the identity of the thing you measured part of the measurement.

**Second, smaller call the same hour.** I nearly filed "animated-favicon's code
generator never runs" after driving two frames and seeing the placeholder text
unchanged — there is a "Generate Script" button ten lines above in the same
file, and I had not pressed it. That is the fifth harness fault from this
workstream's own NOTES, repeated by the person who wrote it down. *What caught
it:* grepping for the handler name before writing the finding, which took four
seconds. **A tool that does nothing until you press the button is the norm on
this site, not the exception.**

---

## 2026-07-29 — I measured against the superset flag and wrote the conclusion before checking which flag it was (bugfix 144)

**The claim:** "0 of the 66 nested `(action, key)` pairs trips the strict-config rule",
written into the PLAN as a safety argument for shipping hard errors.

**What I actually measured:** `opted_in` from `config-key-audit --specs`, which is
`CheckConfig || len(ConfigKeys) > 0` — every action that checks config at all. The rule
that produces a hard error gates on `StrictConfig`, a strictly narrower set
(`IsStrictConfigAction`). I had measured a superset and reported it as the thing.

**What caught it:** a test I wrote failing — I registered a fixture action with
`CheckConfig: true`, expected a hard error, and got none. Not the measurement, not
review: a test, for an unrelated reason.

**Why it survives:** the direction is conservative. Zero under the superset is zero
under the subset, so the conclusion holds. **That is luck, not method.** Had the answer
been "3 pairs", I would have shipped a safety argument built on the wrong population,
and it would have read exactly as authoritative.

**The cheap check:** before quoting a number as evidence, say out loud which predicate
produced it and whether that predicate is the one the code gates on. One grep of
`IsStrictConfigAction` would have done it. Related: the fleet-wide habit of naming the
filter in the sentence — "0 of 66 pairs against the 63 config-CHECKING actions" is a
claim you can audit; "0 pairs trip the strict rule" is not.

## 2026-07-29 — I named a consumer that is not a consumer, in the direction that flattered the submission (bugfix 144)

**The claim,** in a council submission, listing who is affected by a change to what
`ValidateWorkflow` guarantees: *"platform/messaging/processor.go:276 (every agent
message), platform/agentbase/agent.go, and scripts/audit-config-keys.sh."*

**The fact:** `agentbase` never calls `ValidateWorkflow`. It constructs a
`validation.Validator` and uses it for `ValidateIncomingMessage`. `ValidateWorkflow` has
exactly **one** production call site. I had grepped for the package name, seen
`platform/agentbase/agent.go` in the import list, and written it down as a consumer.

**What caught it:** the guardian seat asking for a full call-site enumeration —
`grep -rn "ValidateWorkflow(" --include=*.go` , which takes two seconds and which I had
not run before asserting.

**Note the direction of the error.** Overstating the blast radius made the submission
look more thorough, not less: "I have thought about who this affects". An error that
flatters the work is the one least likely to be re-checked by its author. Both of this
session's wrong calls point the same way — the reassuring answer went unverified while
the alarming ones got queries.

**The cheap check:** an importer is not a caller. Grep the SYMBOL, not the package.

## 2026-07-29 — my sketch summarised in prose what the code already did, and drew two objections for it (bugfix 144)

Two council seats objected that nested cycle detection was "asserted in a trailing
comment with no corresponding code shown" and untested. Both were already implemented
and already tested before I submitted — `validateSubWorkflow` calls `checkForCycles`,
and `TestNestedCycleDetected` pins A→B→A inside a sub-workflow.

The gap was in the SUBMISSION: I showed three checks as code and listed six in a
comment. **A reviewer can only review what is in front of it**, and "hard errors also
cover: …" is a promise, not evidence. The cost was two medium objections on a change
that already satisfied them, and an approval qualified by "approve once cycle detection
is made concrete" — a condition that was already met and could not be seen to be.

**The cheap check:** for each claim in a submission's prose, ask whether the sketch
shows it. If the sketch is a summary of the diff rather than a sample of it, the
reviewer is being asked to take the interesting parts on trust.

> **CORRECTED 2026-07-29 (same day, while building `bugs_open/149`).** The entry
> above says I traced the detector "to a discovery agent that has raised nothing
> automatically since 2026-07-17". **That date is attached to the wrong agent.**
> `orphan_pages` lives in `completeness-discovery-agent`, which **is** running — 144
> work items, most recent **2026-07-25**. The 07-17 date belongs to
> `quality-discovery-agent` (7 items in its whole history, dead since 07-17), which
> carries `unverified_claims` and `voice_tells`. The accurate statement about the
> orphan branch is narrower and stranger: the agent runs, and **no `nav_drift` item
> has ever been raised by a discovery agent** — all 16 came from named sessions or
> `created_by='generic'`. Cause still `[UNMEASURED]` (dispatch coverage / swallowed
> check error / dedup suppression) and filed as `149` B2.
> *Two errors in one line, and the same shape both times: I attached a measurement
> to the nearest plausible subject instead of the one it was `GROUP BY`-ed on.* The
> cheap check is to put the grouping column in the sentence — "**this agent** last
> raised **this item type** on this date" — because a bare date cannot be wrong out
> loud. Family: `a-count-you-kept-is-not-a-census`.

---

## 2026-07-29 — "the pipeline was idle and my run woke it up"

**The claim.** Watching a hand-fired `improvement-sweep` promote 67 items, I saw
`build-pipeline-trigger` go from `complete_idle` at 17:02 and 17:04 to
`call_dispatch` at 17:07 and 17:09, and wrote in chat that the downstream build
pipeline *"was idle for want of promotion, and woke up on it"* — and started to
frame it as the fleet-level cost of `bugs_open/083`.

**What caught it.** Nothing external. I went to quantify the claim for the bug
file — "how long had it been idling?" — and the query answered a different
question than the one I had already concluded:

```
 hour   | idle | dispatched
 16:00  |   11 |          6
 15:00  |    5 |          6
 14:00  |    7 |         17
 13:00  |    0 |         14
```

It had dispatched 6–17 times an hour, all day, for other sites. The two
`complete_idle` rows were a lull of about four minutes, and I had read a
two-sample window as a state of the world.

**Why it matters.** It would have made 083 look like it stalls the fleet's build
pipeline. It does not — parked items are *invisible* to that pipeline, not
blocking it. That distinction is the whole difference between "urgent, everything
is stuck" and "one class of finding is missing from a queue that looks healthy",
and the second is the true and much harder-to-notice version.

**The cheap check:** before describing a system as idle/stuck/dead from
consecutive observations, `GROUP BY` an hour or a day and look at the column you
did NOT sample. Two adjacent rows are a moment, not a rate. Family:
`log-measurement-discipline` and `a-count-you-kept-is-not-a-census` — but the
distinctive part here is that I had *already published the conclusion* and was
only measuring to decorate it. **A number fetched to illustrate a claim is still
allowed to refute it, and it is worth fetching for that reason alone.**

## 2026-07-29 — "these handlers have never repaired anything" (bugs_open/149)

**The claim.** Filing the checker-layer queue, I wrote that `check_orphan_pages`
**"has never repaired a page by any of its three branches"**, evidencing it with
`needs_internal_links` at 33 items / 0 complete since 2026-04-23 and
`orphan_blog_posts` at 3 / 0. I put it at the top of Group A as a measured finding.

**What was true.** The owner's reply: *"the lack of evidence of these tools working
is not evidence that they don't work — they may not have run often."* Correct.
`claimed_by` and `claimed_at` are **NULL on all 37 rows**. Those handlers have never
been offered one of these items. `internal-linker` is not implicated by anything I
measured. The rows are unreachable by the dispatcher, which claims only
`status IN ('triaged','approved')` (`claim_work_item_action.go:102`) — and
**`unresolved` is a TERMINAL status** (`work_items_common.go:29-35`). I had read 27
closed rows as a three-month backlog.

**What caught it.** The owner, on reading the file. Nothing in my own process would
have — I had the completion counts and never asked what else must be true for a
completion to be possible.

**The cheap check that would have.** One column: `claimed_by IS NOT NULL`. And the
general form: **before writing "X does not work", ask what would have had to happen
for X to leave a trace, and verify THAT first.** A zero from a path that was never
exercised looks identical to a zero from a path that fails.

**The corrected finding is sharper and points elsewhere**, which is the usual reward
for this: **20 of the 24 `unresolved` rows were BORN `unresolved`** (`updated_at`
within 5s of `created_at`), across **16 distinct `item_key`s for 24 rows** — repeat
detections branded terminal at birth. That is the recurrence failure already pinned
in `work_item_recurrence_test.go:20,103` ("*born 'unresolved' and never dispatched …
which is how the fix loop silently died*"). So the item is a dispatch/recurrence
defect, not a handler defect, and it is testable.

**The structural lesson, now a banner on 149.** Label every finding by the evidence
it rests on: **MECHANISM** (the code path cannot do the thing, artefact confirms —
survives "it hasn't run much") vs **NEVER RAN** (a zero that means unexercised —
useful for prioritising cadence, useless for judging code). Four items in that file
had to be relabelled, and two of my summary sentences to the owner were wrong in the
same way. Note the asymmetry that makes this insidious: the surviving branch
(`nav_drift`) IS a genuine defect, proven by the code path plus the artefact — so the
paragraph was half right, which is exactly why it read as measured.

Family: zero-adoption-means-read-the-mechanism; two-blind-checks-agree-with-each-other
("a rule measuring ZERO impact may just not be firing" — I have this written down and
applied it to detectors while missing it for handlers).

## 2026-07-29 — I wrote a "delete-marker" that the new code still contains, and it escaped into another session's evidence (bugfix 144)

**The claim,** written into `bugs_open/144`, the workstream RUNBOOK and the concept
register as the way to verify the fix was live: *"`strings /app/agent-chassis | grep -c
"Checking disconnected step"` → 0 … it is a string this change DELETED, and a
delete-marker cannot be satisfied by a stale binary."*

**The fact:** it returns **1** on a correctly deployed pod. I replaced
`fmt.Printf("Checking disconnected step: %s\n", …)` with
`logger.Debug("Checking disconnected step for cycles", …)` — the new string **contains
the old phrase as a prefix**. The check I had labelled load-bearing could not
discriminate in either direction.

**What caught it:** running it. The fix was live and my own marker said it was not.

**Why it is worse than a private slip:** within hours another session had adopted it as
a **positive control** in `bugs_open/153` and drawn the inference *"⇒ 144's pre-fix code
IS what is running"*. A wrong verification method propagates faster than a wrong fix,
because it looks like rigour and gets quoted. Chasing it back showed **all five** of
that bug's markers were unfindable in any binary — four are phrases from the 138 and 104
workstreams' own README/RUNBOOK prose that were never `.go` strings, and the fifth is
inside a Go comment. Three hand-picked markers, two sessions, one day.

**The cheap check, which costs one command and no cluster access:**
`git grep -c "<marker>" -- '*.go'` at the commit you expect to be running. Mine returns
1 for a phrase I claimed to have deleted; theirs returns 0 for four phrases that were
never code. **A marker must be a string the binary EMITS** — not a symbol name, not a
comment, not a sentence from your own docs — and a delete-marker must be one the new
code cannot contain **as a substring**.

Note this is the SECOND time I have picked a bad pod-grep marker in two sessions (the
first: `section-index`, a string in four other files that my change then removed). The
first was caught before it left my desk; this one was not. The recurrence is the
argument for `bugs_open/153`'s fix candidate 1 — stamp the commit sha into the binary
and retire per-fix marker hunting entirely.

---

## 2026-07-29 — I read the documents that stated the position, and not the one that recorded the corrections

Two false claims in the same plan, from the same cause, caught the same way. Logging them
together because separately they look like carelessness and together they look like a
method failure, which is what they are.

**The claims.** In the opening plan for the webdesign.uk build service:

1. That idea.uk *"has taken real money"* and its payment code had *"survived a real sale"*,
   citing the 27 July first sale as evidence the funnel works.
2. That the unit cost of a production report was **$0.641**, cited from
   `idea_uk_vm_site/EVIDENCE_2026-07-27_ai_unit_economics.md`.

**Both wrong, and note that both cited a real source accurately.**

1. **The buyer was the owner.** Genuine external buyers: still zero. The order that later
   looked external was a test, and that lane had *already corrected its own inference*
   (`HANDOFF_RESUME_idea_uk_vm_site.md:17`, `RUNNING_NOTES…:2764`). What I wrote was true
   of the *transaction* and false of the *demand* — and demand is the entire question the
   product I was planning has to answer.
2. **$0.641 was a floor, not a total** — two of five calls — and the EVIDENCE file says so,
   in bold, at the top. I repeated the caveat faithfully and still got it wrong, because by
   then the complete measurement existed and gives a **range**: `~$1.20–$1.45 depending on
   length`, since output tokens are ~92% of spend and cost tracks artefact length.

**What caught it.** Adding this lane's line to `MEMORY_workstreams.md` — the idea.uk entry
there carries both corrections. Cost: one grep.

**The cheap check that would have.** Open the lane's `HANDOFF_RESUME` / `HANDOFF_*_continue_here`
before quoting anything from its `PLAN` or its evidence files. **A workstream's plan holds
its intentions; its resume doc holds its corrections.** I listed `idea_uk_vm_site/` as prior
art, read three files in it, and skipped the one whose entire job is to say what has changed.

**The transferable bit, and it is not "read more docs".** *Being faithful to a superseded
source is still being wrong, and it does not feel like guessing* — which is what makes this
class survive every instinct that catches ordinary invention. Both claims had a citation, a
file path, and a line number. The failure mode of a well-documented project is not the
unsourced claim; it is the **correctly-sourced stale one**. A document named `EVIDENCE_…`
is the most dangerous shape here, because the name asserts that the matter is settled.

**Two smaller ones from the same two sessions, recorded for the tally rather than the story:**

- `SELECT type, name FROM agent_definitions` → the column is `display_name`. The "schema
  first, `\d` before SQL" rule, skipped because the query felt too small to be worth a
  check. Now a landmine entry.
- After the owner ruled on the "thousand sites" figure, §2 of the same plan still said
  *"see §12 — this matters"*, pointing at what was by then a closed decision. **A ruling
  landing makes other parts of the same document lie**, and nothing checks cross-references
  inside a file. Caught on a manual consistency pass; worth one grep of your own section
  numbers whenever a decision closes.

**Where the distilled checks went.** New file, `LANDMINES.md`, created this session at the
owner's direction — the middle rung this file's own header has always implied
(incident → landmine → automation) and that `PROPOSAL_D9_landmines_as_a_footprinted_corpus.md`
(open as **D10**) measured the cost of not having.

---

## 2026-07-29 (later) — "that rung has never had a home", written into two shared docs while seven rows of it sat in Postgres

**The claim.** Creating `LANDMINES.md` at the owner's direction, I wrote that the middle rung
of the ladder (incident → landmine → automation) *"has never had a home"*, and put the same
framing into `CLAUDE.md`. It is the sentence the whole file is justified by.

**False.** `doc_notes` has carried a **`landmine` category since 2026-07-27** — 7 rows,
written by two other threads (*architecture council 2*, *bugsearch-thread*), footprinted by
action name (`subject_type='action'`, `subject_key` = `index_code_symbols`, `scrape_web`,
`store_business_verification`, …). The rung had a home; what it lacked was a *decision*.

**What caught it.** Being asked to present the D10 options, which meant measuring `doc_notes`
properly — `SELECT jsonb_array_elements_text(categories), count(*) … GROUP BY 1`. The
category was fifteenth in the list.

**The cheap check that would have.** One query, before creating a store:
`SELECT count(*) FROM doc_notes WHERE categories ? 'landmine';` — or more honestly, *any*
query at all. **My prior-art search was `find . -iname "*landmine*"` plus a grep of
`docs024_key_docs_latest/`. It never left the filesystem.**

**The transferable bit, and it is the second time in two days.** The `cmd/contrastscan` entry
above has the identical shape: the prior art existed, the search was real, and **the filter
decided the answer** — there it was `--include=*.go` against Python, here it was the
filesystem against Postgres. **On this platform, "does this already exist?" is not a
filesystem question.** Config lives in `agent_definitions`, notes live in `doc_notes`, work
lives in `site_work_items`, and none of it is greppable from a checkout. The proposal I was
implementing *said so on its face* — D10's own §3 is titled "Reuse before building: the store
already exists and is already used this way" — and I read that section, cited it, built
alongside it anyway, because I checked for a *file*.

**The irony worth keeping.** The new file's stated purpose is to catch exactly this: a trap
that fires when you touch something, where the wrong result looks like the right one. Its
own founding sentence was one. Corrected in place in both documents rather than edited away,
and the file now opens by warning that **three** stores exist and that nobody should
consolidate them unilaterally before D10 is ruled.

## 2026-07-29 — I fixed a defect with a derivation that could not reach it (dartsonline_traffic)

**The claim.** Diagnosing why 17 generated icons were unusable, I found the tile
behind an icon rendering from a light literal — `var(--color-icon-chip-bg,
#EEF2F8)` — on a site whose background is `#111520`. I added `icon_chip_bg` to
`darkSchemeDerivations` in `palette_specialised_slots.go`, wrote two tests, and it
compiled and passed. I was ready to submit it as the second half of the fix.

**What was true.** It would have changed nothing at all. The palette reaches a
stylesheet **only** through `{{palette "X" "literal"}}` calls in a LAYOUT template,
so a slot no layout names is never emitted and the component's own fallback ships
regardless of what the derivation list says. Measured across all 18 layouts:
`card_bg` declared by **18**, `surface_alt` by **3**, `icon_chip_bg` by **0**.

And the literal was far narrower than it looked: `icon-chip-bg` appears in exactly
**one** active component fleet-wide (`info-card-grid`, image variant only), while
`image-hover-card-grid` — the component image cards actually use — already reads
`var(--color-surface-alt, var(--color-surface))`, which *is* derived. So the path
that mattered was already correct.

**What caught it.** Going to check who consumed the slot, before submitting, in
order to write the blast radius honestly. Two queries. Nothing prompted them
except the habit of having to state a number.

**The cheap check that would have.** `SELECT count(*) FROM layouts WHERE
css_template LIKE '%palette "<slot>"%'` — one line, before writing any code.
**A derivation is only live if a consumer asks for it.** Passing tests are not
evidence of reach: mine asserted the derivation produced a value, which it did,
and could not see that nothing read it.

**The general form, and the reason this is the entry I most want to keep.** This
is *dead config* — the shape CLAUDE.md warns about — and I was the one about to
create it, while in the middle of fixing a different defect of the same family.
A fix that cannot reach the artefact is indistinguishable from one that does, from
everywhere except the artefact. Removed, and the negative result left as a comment
where the entry would have gone, so the next author does not re-add it.

Family: grep-the-config-key-before-calling-it-a-win; writes-the-field-is-not-reads-the-field;
a-doc-comment-is-not-an-enforcement-mechanism.

## 2026-07-29 — I typed plausible baseline numbers into a verification block, then ran the query (dartsonline_traffic)

**The claim.** Writing a SQL file that queued a head-only rebuild of two pages, I
recorded a "before" baseline in the verification comment so a later reader could
tell whether the body copy had been rewritten:

> `about: hero-about 2723 · about-content 4544 · differentiators 3903 · call-to-action 2179`
> `index: hero 2621 · info-card-grid 4423 · category-listing 2151 · call-to-action 2118`

**What was true.** `about: 2320 · 3358 · 2562 · 2227` and
`index: 3007 · 7144 · 315 · 2336`. **Every one of the eight was wrong.** One was
out by more than 2×, and `category-listing` was 315 bytes — an empty section that
had been broken for a fortnight, which the invented figure of 2151 actively
concealed.

**What caught it.** Running the query, a minute later, because I wanted to paste
the real output into the file. Not a check — an accident of ordering.

**The cheap check that would have.** Running the query first. There is no
sophistication to add here.

**Why it is worth an entry rather than a shrug.** The numbers were not a guess I
labelled as one; they were written in the position where measurements go, in a
file whose whole purpose was to let someone verify a claim. A future reader
comparing against them would have concluded the body had been rewritten when it
had not — and the one genuinely diagnostic figure in the set, the 315-byte empty
section, would have been read as normal. **An invented figure in a verification
block is worse than no figure: it converts an unchecked claim into a checked-looking
one, and it will be trusted by exactly the person who is trying to be careful.**

Family: a-print-statement-is-not-a-config-row; check-answers-the-question-you-encoded.

## 2026-07-29 — my council submission asserted an import cycle and a palette shape, and neither was evidence (dartsonline_traffic)

**Two claims in one submission, both correctly objected to, recorded together
because they are the same error at different scales.**

**Claim 1: "discovery_checks is imported BY actions, so reusing its helper would be
a cycle."** True, as it happens — 6 files one way, 0 the other — but in round 1 I
asserted it without attaching the import graph. `prior_art_librarian` called it
"exactly the ASSERTED-ABSENCE shape: a claim that a reuse path does not exist,
used to license a duplicate implementation, with no import-graph or symbol
evidence attached." Correct. And when I did produce the evidence in round 2, it
showed the *real* obstacle was smaller than I had said: the helper was merely
**unexported and took a `DiscoveryCheckContext`**. Round 2 was APPROVED — and
**seven of thirteen seats** still objected that I had duplicated the query rather
than exporting the function. Exporting it took twenty lines. **My reason for not
reusing something was itself the whole fix, and I had not noticed because I had
described the obstacle instead of measuring it.**

**Claim 2: "no live palette has a dark background with a light surface."** I gated
a new code path on `background` being dark while building its output from
`surface`, and defended the gap in the submission's own risk section by saying I
had manually checked the ten live palettes. The `guardian` seat: that is "an
assumption baked into new code with a silent-failure mode". Correct — and *"I
looked at the current rows"* expires the moment an eleventh palette is authored.
Asking the same question of the same value costs nothing and cannot go stale; both
now call one function.

**What caught them.** The council, both times. Not me.

**The cheap check that would have.** For claim 1: `grep -l` in both directions,
plus `grep "func loadComposedPalette"` to see the signature — under a minute, and
it would have turned "I cannot reuse this" into "I need to export this". For
claim 2: gate on the value you use, not on a neighbouring one.

**The pattern across both.** I wrote a *reason* where a *measurement* belonged, and
in both cases the measurement was cheaper than the sentence. CLAUDE.md already says
"'no collision is possible' is a query, not an argument" — I had that line quoted
in my own round-2 rationale about a *different* defect while committing the same
error twice more in the same document.

Family: answer-review-objections-with-evidence; a-scope-objection-is-not-answered-by-more-evidence;
survey-the-premise-before-building.

## 2026-07-30 — the "safe default" in a script I wrote would have committed the change (dartsonline_traffic)

**The claim.** I wrote a config-change SQL file in the house dry-run style —
`-- BEGIN;` commented out at the head, `ROLLBACK;` live at the foot, with a comment
reading *"safe default: flip to COMMIT only when step 3 and 4 read right"* — and
handed it forward as safe to run unchanged.

**What was true.** Run as written, there is no transaction. Every statement
autocommits and the trailing `ROLLBACK` is a no-op that emits one line —
`WARNING: there is no transaction in progress` — at the very end, after all the
verification output you are actually reading. **The safe default would have
committed the change while appearing not to.**

**What caught it.** Wanting a genuine dry run, and therefore uncommenting `BEGIN`
before the first execution. I got the dry run I intended by accident of intent, not
because the file provided it.

**The cheap check that would have.** Read the first non-comment statement. If it is
not `BEGIN;`, there is no transaction to roll back.

**Why it matters beyond one file.** The convention is sound and widely used here;
the failure is that its two halves are commented out *independently*, so a file can
sit in a state where the guard is present and inert. That is the same shape as the
dead-config entry above: a mechanism that reads as protection and is not.
`(echo 'BEGIN;'; cat file.sql; echo 'ROLLBACK;') | psql` cannot get this wrong.

Family: logging-a-doubt-is-not-a-control-on-it; a-doc-comment-is-not-an-enforcement-mechanism.

## 2026-07-30 — I was one step from filing "the snapshot silently failed" (dartsonline_traffic)

**The claim, not made.** After a config change I checked that its
`snapshot_agent('build-site-planner', '<reason>')` call had produced a restorable
backup. `SELECT count(*) FROM agent_definitions WHERE type='build-site-planner'
AND is_snapshot` returned **0**, and fleet-wide there was exactly **1** snapshot
row, for an unrelated agent. The function had printed
`NOTICE: Snapshot captured: type=build-site-planner, source_id=…` and returned a
uuid. That is a mechanism reporting success while doing nothing — a bug file, and
a good one.

**What was true.** `snapshot_agent` has **two overloads with two destinations.**
The one-argument form inserts an `is_snapshot = true` row into
`agent_definitions`; the two-argument form I had called writes to
**`agent_definitions_backup`**, with `snapshot_taken_at` and `snapshot_reason`
columns. The snapshot was there, timestamped, and — the part that actually matters
— carrying the pre-change text.

**What caught it.** Reading the function body before writing the claim up.
`pg_proc` showed two rows for `proname='snapshot_agent'` with different signatures,
and I had read the wrong one's source first.

**The cheap check that would have.**
`SELECT pg_get_function_identity_arguments(oid) FROM pg_proc WHERE proname='…'`
before concluding anything about what a function does. More usefully: I was asking
the wrong question. **"Does a snapshot row exist" is not the check — "does the
backup hold the PRE-change config" is**, because a snapshot written after the
UPDATE would exist and restore nothing. `backup_has_old_hex = t` is the assertion,
and it happens to answer the existence question for free.

**The near-miss is the point.** Everything about the evidence supported the wrong
conclusion: a success NOTICE, a returned uuid, zero rows in the table I checked, and
a fleet-wide count of one that made "snapshots barely work here" feel plausible. It
would have been a confident, cited, false bug report, and the cost would have been
borne by whoever tried to reproduce it. Two overloads is a boring explanation and it
was the true one.

Family: check-answers-the-question-you-encoded; grep-the-config-key-before-calling-it-a-win;
a-count-you-kept-is-not-a-census.

---

## 2026-07-30 — I told the owner a live page carried four fabricated claims. One did, and it had already been fixed.

**The claim.** An LLM copy rewrite (from the `improvement-sweep` run of 07-29)
replaced gamesdesign.co.uk's hero. I reported four false statements in it — two
wrong counts, a fabricated mechanism, and an invented human credential — called
it "fabricated claims live on a commercial-looking site", **escalated it to the
owner as urgent, and obtained a decision to revert.** Then, re-grounding the
figures before writing, three of the four collapsed:

| what I said | what is true |
|---|---|
| "11 tools" is wrong, there are 14 | **11 is exact** — 11 pages under `/tools/` |
| "10 guides" is wrong, there are 5 | **10 is exact** — 5 `guide-*` + 5 `tool-*-guide` |
| "10,000 Monte Carlo trials" is fabricated | **overstated, not fabricated** — the simulator is real, defaults to `value="50"`, and 10,000 is `Math.min(val, 10000)`, a browser-freeze cap |
| "built by a shipped live-service designer" is invented | **correct, and the only one** — and a later rewrite at 18:10 had already removed it |

**What caught it.** Listing the rows instead of counting them. `SELECT name, url`
on the tool and guide pages showed the site names its guides **two different
ways** — `guide-economy-basics` AND `tool-loot-table-balancer-guide`, both served
from `/guides/`. My `name LIKE 'guide-%'` filter missed five of the ten; my
`name LIKE 'tool-%'` filter swept those same five IN. **One naming convention I
had not looked at produced both errors, in opposite directions, and they looked
like independent confirmation of each other.**

The Monte Carlo error was different and worse in kind: I tested
`rendered_html ~* 'monte carlo'` and read `false` as "the tool does not do this".
A Monte Carlo simulation is repeated random sampling; **nothing obliges the code
to contain the phrase.** I searched for the marketing wording and reported on the
behaviour.

**Why it matters more than the earlier entry on the same day.** The idle-pipeline
call was wrong in a chat message I corrected myself. This one **reached the owner
as an urgent finding with a recommended action, and he decided on it.** An
approval obtained on a false premise is not an approval; had I acted on it
immediately I would have reverted accurate copy, destroyed the one genuine
improvement the loop made, and reported it as a fix. The pause between his
decision and my acting on it was the only thing that caught this, and it was
luck — the token expiring — not method.

**The cheap checks, both of which I own as memory entries already:**
1. **Before asserting a count is wrong, LIST the rows.** A prefix filter encodes a
   naming convention you are assuming; `GROUP BY` and `LIKE` will both confirm it
   back to you. Family: `narrow-filter-defines-the-conclusion`.
2. **Never test a capability by grepping for its NAME.** Test for the behaviour —
   the loop, the constant, the call. A technique that has a marketing name and an
   unremarkable implementation will read as absent every time.
3. **When escalating something as urgent, re-ground every figure in the escalation
   BEFORE sending it, not before acting on the reply.** Urgency is exactly the
   condition under which the checks get skipped, and it is also the condition under
   which being wrong costs the most, because someone else now acts on it.

---

## 2026-07-29 — "your 07-29 ruling makes commit-before-verdict the default" — a policy mis-attribution, caught by the owner

**Thread** "bugsearch 5", reporting a finding from `bugs_open/138`'s council round.

**The claim.** That the owner's 2026-07-29 ordering-exemption ruling had made
committing before a council verdict the normal case, and that this was what caused
`098`'s UNREVIEWED bucket to conflate "never submitted" with "approved, but
committed first".

**Why it was false, on three independent checks — none of which I ran before
writing it.**
1. **Wrong axis.** The ruling answered whether a seam must ship behind a
   **default-OFF switch**. "Review here is after the fact" means after the change is
   **live** — because HEAD is shared and any session's build ships your commit. It
   constrains *submission* timing only ("submit to the gate before or alongside
   committing") and is silent on the verdict.
2. **Wrong source.** The commit-early norm is owner feedback of **2026-07-20**, nine
   days earlier, from a different incident — a thread held the `bugs_open/011` fix
   across four council rounds and the owner's own sweep took it to production on a
   REVISE verdict. That feedback also already anticipated this state ("let the
   trailer *or its deliberate absence* record the review status" + state the verdict
   status in the message body), which I had not done.
3. **"Default" was never measured.** 3 of 3 sampled trailered commits that day were
   committed **after** approval — 2, 51 and 56 minutes after. Threads routinely
   wait. I generalised from my own single case.

**What caught it.** The owner asking me to revisit the claim — not new evidence.
Everything needed to falsify it was already on disk and in my own memory files
before I wrote it.

**The cheap check that would have.** Two: **read the ruling's own text for what
question it answers** before citing it as the cause of anything (the CLAUDE.md
section names its trigger, RFC 002, and that RFC costs the options explicitly); and
**never write "the default" from n=1** — `git log` + one DB query on the trailered
commits was the whole distance, and it took two minutes when I finally did it.

**The pattern, which is the transferable part.** A *mechanism* defect was real and
verified (`3a59b5012` does sit in UNREVIEWED for ever). I then reached for the most
recent salient *policy* as its cause, because I had been working inside that ruling
all day. **A verified symptom lends its credibility to whatever explanation is
adjacent in your context.** The correction did not weaken the finding — it sharpened
it into two live practices that cannot both be satisfied, which is what made the fix
obvious (FIX-056) and what a policy blame would never have produced.

Family: [[narrow-filter-defines-the-conclusion]] (the conclusion was set by what I
had in mind, not by what I measured); [[a-print-statement-is-not-a-config-row]]
(citing a document's existence rather than its content).

## 2026-07-30 — I ran `git stash` on a shared, concurrently-edited tree to answer a question a path-scoped build would have answered with zero risk (bugsearch 8)

**The action.** Fixing `bugs_open/148`, I ran `go build ./...` to sanity-check my
change and hit a pre-existing failure: `found packages main (accessdigest.go) and
working_dir (env.go) in .../traffic_probe/deploy_setup/working_dir` — a package-name
collision in a docs sandbox directory nothing to do with my change. To confirm it
predated me rather than being something I'd caused, I ran `git stash`, re-ran the
build, saw the identical failure, and `git stash pop`'d to restore the tree.

**What was true.** `git stash` captures *every* uncommitted change in the tree, not
just mine. At that moment several other sessions had real WIP sitting there —
`platform/orchestration/actions/save_page_sections_action.go`,
`discovery_checks.go`, and others, mid-edit on `bugs_open/149`. The pop restored
everything correctly and no other session committed in the ~20 seconds the stash
was live, so nothing was actually lost. But CLAUDE.md's entire premise for this
repo is that the tree is shared, mutable state across sessions running right now —
a stash is safe only as long as nobody else touches the tree or commits while it's
off to the side, and I had no way to guarantee that held. I got away with it; that
is not the same as it having been a safe thing to do.

**What caught it.** Nothing did, at the time. I noticed the risk only while writing
this entry up afterwards — the near-miss was invisible in the moment because
nothing observable went wrong.

**The cheap check that would have.** The failing path shares no directory with
anything I touched (`cmd/config-key-audit/`, `scripts/`); reading the path, or
`git log -1 -- <that path>`, would have shown it long predates my change without
touching a single other file. More directly: `go build ./cmd/config-key-audit/...`
— the only build that actually mattered for my change, and the one I ran anyway
right after — made the whole-repo build irrelevant to check at all.

**Why it earns an entry despite nothing breaking.** CLAUDE.md's own example of the
damage class this file warns about (a `git add -A` sweeping another thread's WIP
into an unrelated commit, `69d6f3ecc`) was "an unremarkable 16 files" — ordinary
until it wasn't. A stash/pop round-trip on a tree with concurrent uncommitted work
is the same shape of risk with a shorter fuse and no commit message to reveal it
after the fact. The fix is procedural, not diagnostic: a whole-tree git operation
is never the cheapest way to answer a question a path-scoped build or a targeted
`git log` can answer instead — check that FIRST, not after reaching for the
tree-wide tool.

Family: committing-is-shipping-on-shared-head; shared-tree-wont-compile.

---

## 2026-07-30 — I filed a "zero rows ever" two and a half minutes after it stopped being true

**The claim.** `bugs_open/149` item C3, committed `b15b1456f` at 17:08:30 UTC:

> *"Fleet-wide, `claims_unverified` has **zero rows ever**."*
>
> *"**NEVER RAN, not broken.** The zero rows are what you would expect from a check
> sitting in an agent that has raised 7 items in its life."*

**Wrong at the moment of writing.** Re-measured 2026-07-30: two rows, both
`created_by = 'quality-discovery-agent'` — the detector itself, not a session firing
by hand — created at **17:06:02 UTC on 2026-07-29**. That is **2m28s before the
commit that said they did not exist.** The whole item is built on the zero: it
argues the detector is unexercised, recommends seating it elsewhere, and pairs it
with C1 as "neither a write-time gate *nor* a working backstop". The backstop was
working. It had just found two real things.

**What caught it.** Re-running the count before quoting it, because the handoff that
sent me here said its own figures were a day old and needed re-running. One query.

**The cheap check that would have.** Re-run the count **immediately before writing
it down**, not once at the start of the session. The measurement was almost
certainly correct when taken; it went stale during the session that took it.

**Why this one is worth a row rather than a shrug.** The failure is not carelessness
and it is not a stale *source* — the source was live data, read first-hand, by me.
It is that **a measurement has a timestamp and a claim does not**, and the gap
between them is invisible in the finished sentence. `[UNMEASURED]` markers do
nothing here: the figure *was* measured. The estate already has the rule —
*"ground every figure against the live system before repeating it"* — and I read it
as being about quoting **other documents**. It is not. **Your own measurement from
this morning is another document by the afternoon.**

**And the correction is sharper than the claim was**, which is the tell that this
class is worth catching: what survives is a cadence fact (5 checks / 9 items vs 30
and 22 / 172 and 152), not a code judgement — and it explicitly does *not* justify
the rewrite or the reseating C3 recommended. The same file's own banner warns
*"a lack of evidence that a check works is NOT evidence that it doesn't"*. I did the
mirror image: **evidence of absence that had already expired.**

### Three smaller ones from the same session, recorded for the tally

- **Two of the three traps I hit were already written down, in the concept register
  entry for the exact tool I was using** (`CLM-014`, `cmd/claimscan`): that
  `kubectl exec -i` eats a `while read` loop's stdin (it did — I scanned one site of
  fourteen and the script exited looking successful), and that plain `grep -c`
  returns empty on non-UTF-8 output (it did). I had grepped `LANDMINES.md` by
  footprint, as the rules say, and these were not there — they were in the register.
  **Grepping one of five destinations is not grepping.** Both are now in
  `LANDMINES.md` too. This is `D10`'s "solves authoring, not delivery" gap, observed
  live rather than argued.
- **`CLM-014`'s stated fix for the grep trap is wrong in this shell.** It says use
  `LC_ALL=C grep -ac`. `LC_ALL=C` changes nothing, because `grep` here is a **shell
  function** wrapping `ugrep -I` (skip binary) — the file is skipped entirely and
  `grep -c` prints **nothing at all**, not `0`. `command grep -a` is the fix. I
  applied the documented remedy, got the same empty output, and only then read the
  function. **A remedy that fails silently in the same way as the disease will be
  read as "still clean".** Register updated.
- **I called a scan clean while looking at an empty file.** `grep` over the scan
  output returned nothing and I wrote "zero findings *and* zero output lines — that's
  the shape of a scanner that never ran". That instinct was right and saved it, but I
  had already typed the clean verdict once. The discipline that worked:
  **`grep -c` printing nothing is not a count of zero** — real grep always prints a
  number, so blank output is a tooling failure, never a result.

## 2026-07-30 — bugfix_099 candidate 2 (recoverable plan refusal)

- **I wrote a prompt template field that does not exist, and it would never have
  errored.** The new `repair_plan` step's context section referenced
  `{{.spec_row.body}}`. There is no `body` field on `spec_row` — the design step in
  the same agent uses `{{.spec_row.work_item_id}}`, `{{.spec_row.summary}}` and
  `{{.spec_row.spec_text}}`. **A wrong template path renders an empty string and
  reports nothing**, so the repair step would have run, produced a plan with no spec
  context, and looked entirely healthy. Caught only because I happened to check the
  field name before applying the migration rather than after. The cheap check that
  would have caught it, and now lives in the RUNBOOK:
  `SELECT unnest(regexp_matches(prompt_template, '\{\{\.spec_row[^}]*\}\}', 'g'))`
  over the agent row — read the paths a working step already uses instead of guessing
  from the field's name. Same family as the `[VERIFIED]`-off-an-echo entries: the
  failure mode is **an absence that renders as success**, not an error.
- **I nearly implemented the bug file's own fix candidate verbatim, and it was
  wrong.** `bugs_open/099` candidate 2 says to route the validation failure into the
  existing `repropose` step, "which exists". It does exist — and its prompt is written
  for a *council* revision ("A council reviewed your previous plan and asked for
  revision"), rendering `{{.council_reviews.body}}` and two other review fields.
  `persist_plan` runs **before any council**, so on the very path the fix serves those
  render against nothing. I had already started wiring to it. **What caught it:**
  reading the step's live `prompt_template` instead of trusting its name and the bug
  file's say-so. The lesson is not "bug files can be wrong" — it is that a fix
  candidate naming an existing mechanism is an **untested claim about that
  mechanism**, and the check is one query against the live row.
- **My first design gave a new artefact its own `kind`, which compiles and fails at
  runtime.** `diagnosis_artifacts.kind` is CHECK-constrained to five values. `go
  build` cannot see it, and neither can any test with a mocked DB — it would have
  failed on the **failing branch**, which is the branch nobody exercises before
  shipping. The check is `\d <table>` before choosing a column value, which this
  repo's own conventions already say and I skipped on the way in.
- **I wrote the trap down in my own query, then walked into it 30 seconds later —
  because the wrong answer was a better story.** Measuring council seat headroom
  (`bugs_open/138` candidate 2), I annotated my working query with the observation
  that `max(max_tokens)` is the *highest* cap in the window and that caps had been
  raised mid-window. The next result said `review_editquality` was at **p95 95.1% of
  a 16000 cap**, and I began drafting it as the headline finding: *"the cap raise
  created no headroom — the seat just wrote longer, which proves a raise only moves
  the cliff."* It fitted the bug file's existing argument perfectly. It was an
  artefact of the exact trap I had just typed: the p95 spanned rows at BOTH caps, the
  95% belonged to the retired 8000-cap rows, and the 16000-cap rows peak at **62.9%**.
  The raise worked. **What caught it:** the number was inconsistent with a `max(output_tokens)`
  of 10,071 I had printed two queries earlier — 63% of 16000, not 95%. The cheap check
  is to filter to the seat's CURRENT cap (`join on c.cap = p.cap`) before computing any
  ratio. Two lessons, and the second is the transferable one: **naming a trap is not
  avoiding it** — the note went in the query, not into a filter — and **a result that
  confirms the argument you are already making is the one to re-derive first**, because
  nothing in you will object to it. Same family as the mis-attribution logged above:
  there, a real symptom lent credibility to an adjacent explanation; here, a real
  mechanism lent credibility to a measurement artefact that illustrated it.
- **The same wrong-depth JSON path, one field over, one day later, by the thread that
  documented it.** On 2026-07-29 I logged that `agent_definitions` step caps live at
  `config.ai_service.max_tokens` and that querying `config.max_tokens` returns a
  confident uniform `(unset→default)` for every seat. On 2026-07-30 I read prompts
  from `config.ai_service.prompt_template` and got NULL for **all 51 live review
  seats** — a uniform answer that reads as "these seats have no prompts". The prompt
  is `config.prompt_template`, a SIBLING of `ai_service`. **What caught it:** I had
  successfully read one prompt minutes earlier via `jsonb_pretty(s.value->'config')`
  and could see the shape, so the 51 NULLs contradicted something already on screen.
  The cheap check is `jsonb_pretty` on one row before writing a query against 51.
  **Why this is worth a second entry rather than a footnote on the first:** the
  general lesson ("watch the depth") was already recorded and did **not** transfer,
  because it does not say WHICH keys are nested. A rule that requires you to already
  know the answer is not a check. Fixed by naming both paths in one table —
  `bugfix_138_degraded_gates/RUNBOOK` §4 and LANDMINES (footprint
  `agent_definitions.default_config`).
- **REPEAT of an already-logged trap: I read a council verdict with the wrong JSON
  field and got a clean, uniform, false answer.** Rendering each objection as
  `o->>'detail'` printed `(no objections)` for **all five** objecting seats. The field
  is `problem`. This is the identical shape logged on 2026-07-29 (`ad791d6db`, "a
  wrong-depth JSON path returns a clean uniform answer for all 17 seats rather than
  erroring") — and, per that entry's own correction, itself a repeat of a trap
  documented in `016b` on 07-20. **Third recurrence of one trap in ten days.** What
  saved it: five independent seats unanimously having nothing to say is not what
  disagreement looks like, so the *shape* of the answer was wrong even though the
  query succeeded. The cheap check is `jsonb_pretty(r)` on ONE element before writing
  an aggregation over all of them — read the shape, then query it. Logging the repeat
  rather than the lesson: a check that has now cost three sessions is one to automate,
  not one to remember harder.
- **I costed a platform change by reading ONE of its two enforcement points, wrote
  "the smallest possible platform change" into five documents, and shipped a plan that
  would have reproduced a bug filed for exactly that mistake — twice already.**
  2026-07-30, `staged_component_build`. I read `doc_plans`/`doc_notes`, found
  `subject_type` restricted by a CHECK constraint, and concluded the fix was one
  additive migration with a four-times precedent. That claim went into the lane's PLAN
  (as decision D5), RUNBOOK §3, the PROPOSAL's ordered build list, NOTES, the
  `features_open/027` anchor and a memory topic file. **It was wrong: the contract has
  a second enforcement point in Go** — `validDocSubjectTypes`
  (`platform/orchestration/actions/doc_subjects_common.go`), gating `write_doc_plan`,
  `append_doc_note`, `load_doc_context` and `persist_diagnosis_note`. DDL alone = the
  DB accepts what every doc action refuses, which **is `bugs_open/064`**, filed because
  migration 184 did precisely this and left its own seeded docs unreachable, after
  migration 163 had already missed a different gate. A third instance of one mistake.
  **What caught it:** not review and not care — a **code comment** in the file I was
  about to edit, pointing at
  `experience_register/design/subject_type_addition.md`, a checklist that already
  enumerates all four enforcement points and exists because *"every addition so far has
  missed at least one"*. I would not have found it had the edit lived elsewhere.
  **The cheap check: grep the VALUE you are adding, not the table you are changing.**
  `git grep -n "experience-pattern"` returns the Go list, the migration and the
  checklist in one command — three seconds, and it names every point that must move.
  Reading `\d <table>` tells you what the DATABASE enforces and is silent about every
  gate in front of it; "schema first" is necessary and was not sufficient.
  **A second, smaller wrong call in the same hour, same root:** because I thought the
  DDL was the whole change, I decided (D5) not to number the migration into
  `sql_for_agents/`, to stop another session's `--apply` sweeping it in. That was
  unbuildable — `TestValidDocSubjectTypes_LockstepWithMigrationCheck` parses the newest
  *numbered* migration and fails on drift, so withholding the number reddens HEAD for
  every session the moment the Go edit lands. **A precaution derived from a wrong model
  of the change was itself wrong**, which is the more general shape worth remembering:
  when the diagnosis is incomplete, the mitigations inherit the error.
  Corrected in place in all six places; shipped as one commit carrying both halves
  (`c659e312b`), mutation-proven (Go half alone → the lockstep test fails naming 184's
  failure mode). Landmine filed under footprint `doc_plans`/`doc_notes`/`subject_type`.

- **2026-07-30 — I shipped a claim about our OWN review gate that was false, into
  customer-facing copy, and the thing that caught it was a number disagreeing with
  itself.** Building `tool-review-council-simulator` (fundamentallyai.com), I labelled
  the middle position of its severity slider *"Medium and high block — **what we run**"*
  and made it the default. That is not what we run. **What caught it:** not a review and
  not the copy — the model's own output. With that setting it predicted ~5% of sound
  changes would pass, against our measured 51%. A 10x gap is too big to be the
  independence assumption I had already documented, so the disagreement was the signal.
  **The cheap check, one query, decisive:** count approvals by whether they contain an
  objection of each severity —
  `count(*) FILTER (WHERE EXISTS (… o->>'severity'='medium'))` grouped by
  `decision`. Answer: of **110 approvals, 99 carried a medium objection and passed**, 1
  carried a high; of 15 rejections, **all 15** carried a high. High blocks, medium is
  advisory, and the setting we actually run is the **loosest** of the three, not the
  middle. **Why I got it wrong:** I read `decided_by` on one sample row — *"approved with
  2 advisory objection(s) — none high-severity"* — which states the rule correctly, and
  then encoded the opposite anyway, because my per-seat table was keyed on
  `pct_med_plus` and the column I had built became the rule I assumed. **The
  measurement I chose to compute quietly became the claim I made.**
  **A second, smaller wrong call in the same hour, and it went the other way:** having
  correctly found that `hero-tool` renders no CTA row unless `cta_primary_url` is set, I
  then *talked myself out of my own finding* with `grep -c 'htl-cta-row'`, which returned
  1 and looked like proof the row rendered. It was matching the class **definition**
  inside the component's own inline `<style>` block. **A grep for a CSS class on a page
  that inlines its stylesheet always returns at least one hit**; extracting the element
  showed 0 anchors on the sibling page and 2 on mine. I published the retraction before
  re-checking, so the sequence was right-answer → wrong retraction → right answer, and
  the middle step would have stood if I had not looked twice. Both filed as landmines.

- **2026-07-30 (same session, hours later) — I published a "correction" to CLAUDE.md that
  was itself the error CLAUDE.md's own recorded measurement warns about, and the fleet had
  already got this right two days earlier.** CLAUDE.md's council-gate section says approval
  "ran ~80%". I measured 51% post-fix, concluded the doc was quoting a two-day peak, and
  wrote that into NOTES, a handoff and a commit message before checking anything else.
  **It is not a peak — it is a different denominator.** Per ROUND 50.7%; per SUBMISSION
  77.2% (105 of 136 correlations eventually approved). Both true; the doc's figure is the
  per-submission one and is sound on that basis. **What caught it:** opening
  `council-review-practice-index.md` to add a line, where **line 24 already read
  "per-ROUND 51% / per-SUBMISSION 76% — a REVISE is the median"**, measured 2026-07-28 by
  another thread, with the query and the same conclusion. I reproduced their number
  exactly and read the agreement as *novelty* instead of as corroboration.
  **The cheap check, and it is embarrassingly cheap: read the existing memory on a
  mechanism BEFORE publishing a correction to the doc about it.** `grep -ril "approval"
  ~/.claude/projects/*/memory/` or simply opening the topic file the index already points
  at. I did open it — but only *after* writing the claim, and only because I wanted to add
  a line to it. **Had I been adding nothing, I would never have looked, and the wrong
  explanation would have shipped in a handoff as fact.**
  **The general shape, which is the reusable part:** a figure that *disagrees* with a doc
  prompts a check; a figure that *agrees with your own new measurement* does not, so
  "my number matches the recorded number" felt like confirmation of my reading rather than
  evidence that the recorded reading was already correct. Two measurements agreeing tells
  you nothing about whether your *interpretation* of them is the one already on file.
  Tally: "read the fleet's existing record before correcting a shared doc" → this is the
  first entry; adjacent to [[prior-art-search-goes-stale]] but the inverse failure — the
  prior art was fresh, findable, and indexed, and I simply did not look.
  Ended up net-positive rather than only a retraction: the tool now prints **both**
  denominators, since conflating them is exactly how a normal REVISE reads as a failing
  plan. Corrected in place in NOTES and the handoff; commit `32653bd85` carries the wrong
  explanation and cannot be amended (forward-only).

---

## 2026-07-30 — I called a scheduled task's pre-query defective, from a `left(...,120)` truncation of it

**The claim I made (in chat, to the owner):** `report-dispatch`'s `pre_query` "counts every
`report_request` row **with no status filter**. It will read '3 queued' forever, including
when the answer is zero" — offered as the platform making the same measurement mistake I had
logged against myself the same afternoon, and as "a one-line fix if you want it".

**It is false, and the task is well built.** The full query is:

```sql
SELECT count(*)::text AS queued_reports FROM site_work_items
WHERE pipeline = 'reports' AND item_type = 'report_request'
  AND ((status = 'awaiting_report' AND attempt_count < max_attempts)
    OR (status = 'reporting' AND claimed_at < NOW() - INTERVAL '30 minutes'))
HAVING count(*) > 0
```

It has the status filter, a `max_attempts` guard, a stuck-claim reaper clause **and** a
`HAVING count(*) > 0` whose entire purpose is to return **zero rows** when there is nothing
to do — which is what makes the scheduler skip publishing rather than wake an agent for
nothing. The scheduler's own log said so plainly and I only read it afterwards:
`"Pre-query found no rows — task ran with nothing to do","task":"report-dispatch"`.

**What caught it:** the scheduler log contradicting my prediction. I had predicted the task
would "fire every 90 seconds and claim nothing"; it did not fire a message at all, and
chasing *why not* is what made me fetch the untruncated query.

**The cheap check that would have:** do not display a query with `left(col,120)` and then
reason about the query. I wrote the truncation myself, for table width, and then read the
result as if it were the whole thing — the `HAVING` clause that refutes the entire claim was
at character ~250. **If you truncate a field for display, you may quote what you SAW but not
what it MEANS.** Same family as [[council-submission-quote-fidelity]] ("an abbreviated quote
is a DIFFERENT claim"), which I had cited in my own memory index hours earlier — the trap is
not ignorance of the rule, it is that a truncation you applied yourself does not *look* like
someone else's abbreviation.

Tally: "reasoned from a deliberately truncated read" → 1. Adjacent and worth counting
together with the mutation-testing pair filed the same day in `LANDMINES.md`: all three are
the same error at different altitudes — **the evidence was available and I read a reduced
form of it.**

No harm done beyond the wrong statement: nothing was changed on the basis of it, because I
had explicitly not touched the task. Corrected to the owner in the next message.

---

## 2026-07-30 — I nearly filed two confident false findings into a bug file, both because my `sed` window stopped short (bugs_closed/133)

**The claims:** while fixing `bugs_open/133` I was ready to write, as new findings extending
the bug, (1) *"`raw_html` is never uploaded, so even the `upload_results: true` rows are
unsafe"* and, minutes later, (2) *"per-page content is never uploaded, so a page marker can
never be true."*

**Both false.** `uploadScrapingResults` is ~340 lines. I read it with
`sed -n '525,700p'`. `raw_html` is uploaded at **line 760**; the per-page uploads are at
**812–856**. Both were past my window, in a function I had "read".

**What caught it:** re-running the grep over the *whole* function rather than my window —
`sed -n '525,830p' … | grep -n 'uploadInfo\['` — because I wanted the exact URI key names to
build the field→URI map. Nothing about the first read felt partial; it had a plausible,
complete-looking answer in it (three uploads, three truncated fields, one missing).

**The cheap check that would have:** `grep -c` the symbol across the whole **function or
file** before concluding an absence, never across the window you happened to read. An absence
claim is only as wide as your search, and a window chosen to answer the question you are
already asking will tend to confirm it. One command:
`awk '/^func \(a \*Adapter\) uploadScrapingResults/,/^}/' file.go | grep -n needle`.

**Direction of the error matters and it was the bad one:** both false findings would have made
the bug look *worse* and would have been quoted forward as measured facts in a file whose
whole value is that its measurements are trustworthy.

Tally: "concluded an absence from a partial read" → 2 (same session, ten minutes apart, with
the lesson from the first already written down). Same family as the truncated-query entry
above: **the evidence was available and I read a reduced form of it** — now 5 instances.

## 2026-07-30 — and then the same error in the bug's OWN measurement, which is why "re-run it" means re-derive the WHERE clause (bugs_closed/133)

**The claim (not mine — the bug file's, carried forward by me for most of a session):**
*"Live exposure — 4 of the 6 single-URL scrape steps in the fleet."* I re-ran its query as
instructed, got the identical result, and recorded "confirmed unchanged 2026-07-30" in my PLAN
and in my council submission.

**It is 9 of 14.** The query filters
`v->>'action' IN ('scrape_web','firecrawl_scrape','batch_webscrape')`. **Six** actions reach
that adapter; the list omits `fetch_scrape`, `firecrawl_crawl` and `firecrawl_extract`. The
omission includes `feed-ingester`, which is the **highest-volume live scraper on the topic**
(real messages 07-29 19:54Z and 07-30 07:57Z, both `upload_results:false`), and four
`firecrawl_crawl` steps, which are exactly the multi-page case where five of six per-page
fields can never have a stored copy.

**What caught it:** not the re-run — the re-run *confirmed* the wrong number, because I
re-executed the SQL instead of re-deriving it. It was caught by reading a real message off the
request topic (to copy a shape the adapter accepts before firing a probe) and noticing a
`requesting_agent_type` that appeared nowhere in the table. The prior-art-librarian council
seat had independently asked for exactly this claim to be re-checked against
`agent_definitions`; it was right, and acting on it more than doubled the measured exposure.

**The cheap check that would have:** derive the action list from the **code that dispatches**,
not from the bug file's `IN (...)`:
`grep -rn "webscrapeAdapterTopic" platform/orchestration/actions/` → then enumerate every
action registered against those files. One command, and it is the difference between 4 and 9.

**The transferable rule, which is stronger than "re-run the query":** *"re-run it rather than
trusting this table"* is satisfied by re-executing the SQL, and re-executing the SQL
reproduces the filter's blindness perfectly. **A re-measurement inherits every assumption
encoded in the WHERE clause.** If the figure is load-bearing, re-derive the population, not
just the count. Same shape as [[narrow-filter-defines-the-conclusion]] — but arriving in
someone else's careful, well-evidenced measurement, which is what made it credible.

Tally: "re-ran a query and called it re-measured" → 1. This is the one I would most expect to
repeat, because it looks exactly like diligence.

---

## 2026-07-31 — "the first real consumer of the schema-driven fill", about a mechanism that had existed for two months

**The claim.** Shipping GTM on idea.uk, I designed a per-site chrome-config seam
(`site_specs` key + map-valued `input_schema` field with `source: config.*` + an `{{if}}`
gate) and wrote it up as concept **STY-050**, describing it in the commit message and the
register as *"the first real consumer of `bugs_open/018`'s schema-driven fill"*.

**It was false.** `head-seo-standard` — a head component covering 4 live domains — has
carried the identical pattern since **2026-05-13**: a gated `{{if .analytics_id}}` block
feeding gtag.js, declared `{"type":"text","source":"config.analytics_id","required":false,
"on_missing":"skip_field"}`. Not a near-miss: the same source prefix, the same gate, for
the same purpose (analytics), two months earlier.

**What caught it.** Luck, one day late. A Phase C query checking `<meta charset>` anchors
across the three head components happened to also select `has_gtm`, and a component I had
never touched came back `true`. Nothing about the original work would ever have surfaced it.

**The cheap check that would have:** grep the **live component corpus** for the thing being
built, before designing it —
`SELECT name FROM content_components WHERE html_template ILIKE '%googletagmanager%' OR
 html_template ILIKE '%gtag%';`
One query, ~1 second. I *did* follow CLAUDE.md's "grep before you file" — `/bugs_open/`,
`/bugs_closed/`, the workstream dirs, the concept register. **Every one of those searches
was over prose.** The mechanism lived in a `content_components` row, and no amount of
grepping documentation finds a rendering seam that only exists as data. The register is
explicitly known to be blind post-2026-07-13 (`bugs_open/106`), so for anything
config-shaped the *repo* is the wrong corpus by construction.

**Second error, same investigation, nearly a second false entry.** Having found it, I
queried `input_schema->'analytics_id'`, got NULL, and wrote "not declared — a dead seam".
The shape is **wrapped**: `input_schema->'fields'->'analytics_id'` is the live path, and
`render_site_components_action.go:604-607` handles both shapes *precisely because both
exist in the fleet*. A correctly-built seam looked like a broken one because I guessed the
JSON path. Caught only by re-checking before asserting — the same class as the
`pages.rendered_html` / `site_components.function` guesses the day before, except **jsonb
returns NULL instead of erroring**, so this one would have shipped.

**What it cost.** Nothing live — the mechanism is sound and the correction is inline in
STY-050. What it nearly cost is worse than a duplicate: two seams for one job across
overlapping domains, where a site carrying both would load GA4 directly **and** through
GTM and double-count every pageview. That is now a stated Phase C decision instead of a
discovery someone makes from a bad traffic graph.

**The transferable rule:** *"has this been built already?"* and *"is it written down?"* are
different questions with different corpora. On this platform a mechanism can be entirely
**config** — a component template, an `input_schema`, an agent definition row — and be
invisible to every documentation search you are told to run. **Query the live tables for
the mechanism, not just the docs for the idea.** Related: [[prior-art-search-goes-stale]]
(an absence is only true when you looked), [[seed-sql-is-history-live-row-is-fact]] (the
live row is the fact), [[grep-the-config-key-before-calling-it-a-win]].

Tally: "searched the docs and called it prior-art-checked" → 1. "Guessed a jsonb path and
read NULL as absence" → 1 (and this one fails silently, unlike a bad column name).

---

## 2026-07-31 — a security comment I wrote said "the exact bypass this closes", and my own test proved it false before it shipped

**The claim.** Writing `platform/fetchguard`'s dial-time IP check, I called `ip.Unmap()`
before classifying each resolved address, and wrote: *"`::ffff:169.254.169.254` must be
judged as the v4 address it wraps"* — with the surrounding prose calling it *"the exact
bypass unmapping exists to close."* Confident, specific, and read like settled fact.

**False.** `net/netip`'s `IsPrivate()`, `IsLinkLocalUnicast()`, `IsLoopback()` and the rest
already resolve an IPv4-in-IPv6-mapped address correctly with **no** unmap step. Checked
directly: `IsPubliclyRoutable(mapped) == IsPubliclyRoutable(mapped.Unmap())` for the
metadata address and for a private one. `Unmap()` still does something real — it makes
`ip.String()` print `"169.254.169.254"` instead of `"::ffff:169.254.169.254"` in an error
message — just not the thing the comment claimed.

**What caught it.** My own test, in the same sitting, before the code was committed —
`TestIsPubliclyRoutable_IPv4MappedRequiresUnmap` was written to *prove* the claim and
instead disproved it: both forms classified as blocked with or without unmapping, so the
"unmap made no difference" assertion inside the test itself fired.

**The cheap check that would have.** Exactly the one that caught it: write the comparison
as a test before writing the comment as fact. The failure mode here isn't "didn't check" —
I did write a test — it's that **the comment was drafted with the same confidence as the
code**, and nothing distinguished "I verified this" from "this sounds right and matches how
these bugs usually work" until the test ran.

**The transferable bit.** A security-rationale comment is a *claim about behaviour*, the
same family as this file's other entries (`a-doc-comment-is-not-an-enforcement-mechanism`,
`a-print-statement-is-not-a-config-row`) — and it is dangerous in a **new** way here,
because nobody else's finding could have caught it: there was no prior art to contradict,
no existing bug to grep for, just an assertion about a standard-library function's exact
semantics. The only check that works on a claim with no external referee is the one this
package happened to already be writing anyway: **a test that tries to prove the claim,
not just tests the code that assumes it.** Corrected in place in both the code comment and
the test itself (renamed, since the old name asserted the disproved claim) rather than
deleted — the corrected version explains what `Unmap()` is *actually* for, which is more
useful than silence would have been.

---

## 2026-07-31 — a working page read as broken, because the number was right and the inference from it was not

**The claim.** `gauntlet_dead_cta/HANDOFF_2026-07-30_B` told the next thread that
vonc's `/provocations/index.html` "paints neither today's provocation nor,
apparently, much else (1,293 chars of visible text)", and to "check what that page
is actually showing before designing the archive; it may already be broken
independently." I picked that handoff up and set out to confirm it.

**False.** The page works. All 8 archive entries paint with date, title and
teaser; 7 are openable and the 8th is deliberately non-openable because no case
was written for it (the builder's documented behaviour). The empty state is
correctly hidden. 1,293 characters is simply what 8 short entries plus chrome
measures.

**What caught it.** Rendering the page and printing 600 characters of DOM context
around each match, rather than reading a count. Note the sequence, because the
first two things I found were also wrong: `grep -c` said the empty-state string
`"Nothing filed yet"` was present (it is — inside `hidden=""`), and that
`"Nobody actually"` was present, which I read as today's headline leaking onto the
archive page. It was the **29 Jun entry**, *"Nobody actually reads terms of
service"*; today's headline is *"Nobody actually wants a personalised internet"*.
A 15-character substring matched two different provocations. I then suspected a
visible blank row, which turned out to be a correctly hidden
`data-archive-template`. **Three false positives in one file before the real
answer, all from the same habit of trusting a match without reading it.**

**The cheap check that would have.** Print the context around a match before
drawing any conclusion from its presence or its count. Two lines of Python.

**The transferable bit, which is the reason this is worth a row.** The original
measurement was **not wrong** — 1,293 chars is accurate, and today's provocation
genuinely is not on that page. The probe was built to answer *"does today's
headline leak here?"*, and its correct answer was "no". The error was reusing that
run's incidental by-product — a character count — as evidence for a different
question, *"is this page working?"*, which it was never designed to answer and for
which a low number is not diagnostic. This is the family already in memory as
[[check-answers-the-question-you-encoded]] and
[[narrow-filter-defines-the-conclusion]], with a twist worth naming: here there
was **no faulty filter to notice**. The instrument was sound and the number was
true, so nothing looked suspicious. **A measurement's validity does not travel
with the number to the next question you ask of it** — and a hedge ("apparently")
does not stop a downstream thread acting on it, because the next reader inherits
the claim and not the doubt.

**Second, smaller, mine alone.** I had drafted a decision question for the owner
whose framing rested on the page being broken. When the correction landed I
updated the facts in my notes and **did not re-read the question built on them** —
the superseded premise went out in the question. Re-read outward-facing artefacts
after a correction, not just the notes where you recorded it.
- **I named a column for what I wanted it to mean, and my detector then reported health it
  had never measured — on the exact row I had just "fixed".** 2026-07-31,
  `staged_component_build/CHECK_naming_contract.sh`. The check tested
  `EXISTS (SELECT 1 FROM doc_plans WHERE subject_type='tool' AND subject_key=…)` and I named
  the result **`has_fence`**. It proves a PLAN **row** exists and says nothing about whether
  that PLAN contains a ```criteria fence — which is the thing an acceptance run actually
  needs. `tool-review-council-simulator` has a PLAN and no fence. So when I renamed its page
  to make it resolvable, the check promoted it from BROKEN to **"testable now"**, and it is
  not testable: the run starts and then **SKIPS with `needs_criteria`**, which is the precise
  silent class the check was written to find. **A detector that reports health it has not
  measured is worse than no detector**, and this failed in the worst direction, on the one
  row I had acted on, in the output I was about to quote as progress.
  **What caught it:** not the check. I went to fire an acceptance run to prove the rename had
  achieved something, read the PLAN to see what its fence asserted, and **there was no
  fence**. The intent to verify the *outcome* rather than the *checker* is the only reason.
  **The cheap check: grep your own predicate against the thing it claims to measure.** One
  query — `count(*) FILTER (WHERE plan_row AND fence)` versus `FILTER (WHERE plan_row AND NOT
  fence)` — returns 10 versus 1 and settles it in seconds. More generally: **a column name is
  a claim, and an `EXISTS` on a parent row is not evidence about a child property.**
  **Two further faults in the same file within the hour, same root — writing checks faster
  than I was testing them.** (1) `kubectl exec -i` inside a `while read` loop ate the loop's
  stdin, so the script listed one of two findings while the summary correctly said two;
  under-reporting, caught only because the count is computed separately from the list.
  (2) ```` '%```criteria%' ```` inline inside a double-quoted bash string is **command
  substitution**, so the script would not parse — a landmine already recorded fleet-wide
  (*"backticks in `-m` execute"*), hit **one edit after** writing a comment about a different
  silent-failure trap in the same file. **Tally: eight instances of this one class in two
  days, and the last three were all inside the detector built to catch the class.** The
  transferable rule is not "be more careful" — it is **watch every branch of a new check
  fail before quoting anything it says**, and that applies hardest to the checks you write to
  police everyone else's.

---

**2026-07-31 — gauntlet_dead_cta — "599 characters fits inside a 737-character budget."**
It did not, and the card I built to prove it **overlapped its own ruling line**.

I measured the share card's capacity properly — canvas `measureText`, real font metrics,
per type size — and produced a table saying a 1200×630 card holds ~737 characters at 32px.
Then I built the mock with 599 characters of real round text and shipped it into a
comparison for the owner. The defence block ran straight through the verdict line
underneath it.

**What was wrong:** the budget measured *prose against the frame*. The actual card also
carries two section labels, a rule, a ruling line and a footer, and those take vertical
space the prose then cannot have — about 25% of it. The number was not miscalculated; it
was **answering a different question** than the one I used it for, which is the
`narrow-filter-defines-the-conclusion` family: my measurement encoded "how much text fits
in a rectangle", and I read it as "how much text fits on this card".

**What caught it:** looking at the rendered PNG. Nothing else would have — every figure
in the table was correct, and a check that recomputed the budget would have agreed with
itself. This is the *two-blind-checks-agree* shape: the verifier and the thing verified
shared the same blind spot.

**The cheap check I skipped:** auto-fit against the **drawn layout** rather than a
character budget — binary-search the type size with the labels and chrome included, which
is four lines of code and cannot disagree with the output because it *is* the output. Both
the shipped renderer and every later mock now do this, and the shipped one records why in
a comment so the budget table cannot be reintroduced as a shortcut.

**Tally note:** this is the second time in two days a *correct* measurement was used to
answer a question it did not encode. The recurring fix is not more precision — it is
**assert on the artefact, not on the model of the artefact.**

---

## 2026-07-31 — "12 calculators": a figure five documents repeated and nobody ever measured

**Lane:** loancalculator_couk. **Claim:** *"a hand-built 27-page UK loan site (12
inline-JS calculators, 13 guides)"* — written in the PLAN, the NOTES, the SUMMARY, the
README and finally the HANDOFF, which is where I read it and believed it.

**What was wrong:** it is **11**. `tools/credit-roadmap.html` is a static prose page
that happens to live in the tools folder: zero `<input>`/`<button>`/`<select>`/
`onclick`/`addEventListener`, and its only `<script>` is the shared `nav.js`.
12 = 27 − 13 guides − index − legal. **The number was derived by subtraction from the
directory layout and then written in the voice of an inventory.**

**Why it mattered more than a miscount.** The lane's acceptance bar is the owner's own
sentence — *"starts similarly enough with working tools"* — operationalised as "every
calculator still computes in a real browser". Measured over 12 that gate is
**unpassable**: one of the twelve cannot compute and never could, so a correct build
reports `11/12` for ever. An always-failing gate gets explained away, and then
ignored, and then the one real regression it existed to catch arrives looking exactly
like the noise. **The wrong denominator does not weaken the gate, it destroys it** —
and it fails in the direction that blames the site.

**What caught it:** taking the baseline *before* changing anything, which is the only
reason I ran anything over that page at all. Static grep and a real-browser audit
disagreed with the docs and agreed with each other — `NO-CONTROL — nothing a visitor
can touch` against `RESPONDS` for the other 11.

**The cheap check I skipped, and every prior session skipped:**
`grep -c '<input\|<button\|<select\|onclick\|addEventListener' tools/*.html` — one
command, over files already on disk, at any point in the previous two days. Nobody ran
it because the figure never looked like a claim; it looked like a description of a
directory.

**The transferable shape.** This is the `[UNMEASURED]`-marker rule failing in its
blindest spot: **an inference that is stated once, then quoted forward, launders itself
into a finding.** By the fifth document it carried the authority of five sources and
had exactly one origin, which was arithmetic on a folder listing. A marker on the
first writing would have stopped it; so would asking "which page is the twelfth?"

Two smaller instances the same morning, both the same shape and both caught by reading
the mechanism rather than the symptom:

- **"the blocker is nested `<html>`"** — true and *incomplete*. `site_components` is
  **empty (0 rows)**, and `assemblePage` reads chrome from that table, so the flip
  this lane was warning about would also have shipped every page with **no head, no
  nav and no footer**. Recorded as the whole blocker for a day. Caught by reading
  `getSiteComponents` before writing the extractor.
- **"27/27 byte-exact"** — this one **held**, but my first re-verification appeared to
  refute it (`length()` = 5,730 vs a 5,734-byte file) and I nearly filed a fidelity
  regression. The gap was four `£` signs; `length()` counts characters. Now a landmine,
  because the in-pipeline gate the council asked for was about to be built on it.

**Tally note:** every entry above was caught by *measuring the thing itself before
acting on a description of it*. The recurring cheap check is not a new tool — it is
running the one-line query against the live system **before** repeating a figure, and
marking it `[UNMEASURED]` when you don't.
