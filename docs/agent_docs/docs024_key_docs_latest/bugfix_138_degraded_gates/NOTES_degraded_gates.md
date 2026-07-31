# NOTES — bugs_open/138 degraded gates

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-07-29 — filing, and the accident that found it

138 was not looked for. It was found while verifying the `review_architecture`
seat this same thread had shipped hours earlier: the seat's verdicts kept coming
back `revise` with `decided_by: gating objection from architecture` while every
objection it had raised was `"severity": "medium"` — and its own prompt tells it
medium is advisory. The seat appeared to be contradicting itself.

It was not. It was being cut off at `max_tokens`, and `Degraded` gates
unconditionally.

**The near-miss worth recording: the seat looked like a bad seat.** Object rate
2-of-3 in its first three reviews. The documented kill-switch for a noisy seat is
exactly a high object rate, and I was the person who would have pulled it — I had
seated it that morning and was watching for evidence it was not earning its place.
Every observable pointed at "retire it". The truncation was invisible because the
verdict named the seat.

## 2026-07-29 ~12:30Z — the seat fix proved, and the confounder demonstrated

Raising that seat to 16000 + putting `notes` first + a length budget: truncations
2 → 0 over 12 subsequent reviews, and the object rate went **2-of-3 → 2-of-12**.

That number is the whole argument. The seat was never noisy. **Acting on the raw
object rate would have retired a working seat**, and no amount of care would have
caught it, because the signal that would have shown the truncation was itself in
the part being truncated (`ARCHITECTURE_SIGNAL` lived in `notes`, emitted last).

One residual `degraded` after the cutover, and it is explained rather than
excused: orchestration `815b38c3` spawned 07:16:59, **before** the 07:19:36 config
change. **An orchestration carries the workflow definition it loaded at SPAWN** —
so "DB config is live immediately" is true of the row and false of any running
round. A verification that looks at in-flight rounds reads as a failed change.

## 2026-07-29 ~13:00Z — writing candidate 1, and three things I got wrong

**Misstep 1 — I first wrote the split as `objectionGatesOnMerits`, and it was
wrong on the zero-objections case.** My initial version said "gates on merits =
`len(Objections)==0` OR any gating severity", mirroring the original rule. That
labels a review *cut off before it wrote any objection* as a merits gate — which
is precisely backwards, because emptiness in a Degraded review is the clearest
possible evidence OF truncation. Caught while writing the test table, not by the
compiler: both versions compile, both pass the old tests, and the difference only
shows up when you have to write down what each case *means*. Rewrote as
`hasGatingObjection` + `gatesOnlyBecauseTruncated`, which splits on `Degraded`
explicitly.

> The general lesson: mechanically extracting a predicate preserves BEHAVIOUR but
> can silently invert MEANING. The extraction was faithful; the naming made a
> claim the code did not support.

**Misstep 2 — I nearly recorded `editquality` as still at 8000, off my own bad
query.** I queried `s.value->'config'->>'max_tokens'` and got `(unset→default)`
for all 17 seats, which looked like a decisive finding ("nobody has right-sized
these"). It was a wrong-depth JSON path: the real location is
`config.ai_service.max_tokens`. **The wrong path does not error — it returns a
clean, plausible, uniform answer.** editquality was in fact already at 16000, and
so were the other two worst offenders. Had I written it up, the bug file would
have carried a confident false claim about the live roster, and the "candidate 3
is barely started" conclusion that follows from it.

Caught only because the answer disagreed with something I already knew (I had
raised `architecture` myself and seen 16000 in `llm_call_log`). **That is luck, not
method** — the check that would have caught it without luck is dumping one object's
keys before querying a path into it.

**Misstep 3 — the package would not compile, and it was not mine.**
`go vet` failed on `platform/orchestration/datahelpers/claims.go:494: undefined:
negatedClaimMatch`. First instinct was that I had broken something. `git status`
showed the file MODIFIED in the working tree by another session mid-edit; building
`git archive HEAD` in a scratch dir compiled clean. **A red build in a shared tree
is not evidence about your own change until you separate the two** — and the
session-start `git status` I was carrying did not list that file, because it is a
snapshot and goes stale in minutes.

## 2026-07-29 ~13:10Z — measured before submitting, per the ordering-exemption ruling

Replayed the new labelling rule over 14 days of stored `reviews[]` rather than
asserting a blast radius (the RUNBOOK §2 query). 63 gated revise rounds → **10
would now read TRUNCATED**, 3 mixed, 50 unchanged, and **exactly 1 round changes
which seat it names**.

Reconciling that against this bug's own headline "17" took a moment and is worth
recording, because it looked at first like a contradiction: the 17 counts **seats**
(degraded objections that gate), the 10 counts **rounds where nothing else gated**.
Same window, different units. The full chain: 18 degraded gating seats (17 became
18 as the rolling window moved) → 15 gate solely on truncation → in 15 distinct
reports → 13 of which were decided by a gating objection → 10 with no merits gate
at all. Every step of that is a filter, and stating only the endpoints would have
made the two numbers look irreconcilable.

**Also worth noting what the measurement CHANGED, not just confirmed:** it showed
every seat that has actually produced a truncation gate is now at 16000 except
`guardian` (1 occurrence in 14 days). I had expected to find candidate 3 barely
started and to argue for it; the data says it is nearly done for the seats that
matter. So the argument for candidate 1 shifted while writing it — from "the caps
are wrong" to "a cap raise moves the door rather than closing it, and
`architecture` proved that within hours by being a longer prompt against the same
cap". The second argument is the true one and the first was never needed.

## 2026-07-29 ~13:25Z — submitted to the council gate

`SUBMISSION_CORR = 919a05bf-c51a-440b-865e-bd07e69e1c36`, via
`TARGET_AGENT_TYPE=council-gate-orchestrator` (proven end to end 07-28; releases
the request lane in ~8s so the round does not head-of-line block the fleet, per
`bugs_open/096`). Budget ~30 minutes, not ~2 — the council itself takes 2–5 min
and the rest is queue.

Submitted **after** committing (`3a59b5012`), which the 2026-07-29 owner ruling
says is the honest order on this tree rather than a lapse: HEAD is shared, and any
other session's build ships my commit, so there is no version of this where I hold
it back pending a verdict. Condition (1) of the old ordering exemption — "state
your ordering constraint" — was retired precisely because no thread can supply one.

Find the run by PAYLOAD, not by the printed id:

```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '919a05bf-c51a-440b-865e-bd07e69e1c36';
```

**A missing row is almost always latency, not a dropped dispatch.** Do not retry
on that evidence — it costs a duplicate round.

### One thing I did NOT do, and the reason

Condition (2) of the seam rule — register in the SAME commit that ships it — I
missed. FIX-055 went in as a follow-up (`6dd5cbb3d`) instead. Forward-only, so
there is no amend; the commit message says plainly that it should have been in the
previous one rather than presenting the split as the plan. Recording it here
because the whole value of that condition is that it stops a seam becoming
folklore, and a seam registered ten minutes late by the same thread is the *near*
miss, not the safe case — the thread that forgets is usually the one that has
moved on.

---

## 2026-07-30 — candidates 2 and 4. Two measurements went wrong before they went right.

### Candidate 1 is now LIVE and verified at the pod, which was owed from yesterday

Both replicas of `agent-chassis-785d5499c6` (yesterday's pre-roll pod was
`agent-chassis-6fd7d88c4d-f6pgj`, a different replicaset, so this was a real roll):

| marker | pre-roll 07-29 | now |
|---|---|---|
| `gated_by_truncation` | 0 | **1** |
| `gating TRUNCATED objection from` | 0 | **1** |
| `all reviewers approve` | 1 | 1 — control |
| `gating objection from` | 1 | 1 — control |

Both controls non-zero **in the same exec**, which is what makes the two zeros→ones
readable at all (this change deletes no string, so there is no delete-marker).

And it has RUN, not merely shipped: two reports now carry the field, both `false`,
one of them `gating objection from prior_art_librarian` at 14:50 today — a genuine
merits gate, so the *preference* rule (name the merits seat, not the truncated one)
is exercised live and not only in the unit tests. **What is still unproven live is
the `true` branch**: no post-roll round has had a truncated gating seat. Unit-tested
(7 cases), and the persistence path is proven by the two `false` rows, so what
remains unproven is the wording. Catch it when it happens rather than inducing it:

```sql
SELECT created_at, body::jsonb->>'decided_by' FROM diagnosis_artifacts
WHERE kind='council_report' AND metadata->>'gated_by_truncation'='true'
ORDER BY created_at DESC LIMIT 5;
```

### MISSTEP 1 — I flagged a trap in a comment and then walked into it 30 seconds later

Measuring headroom as `output_tokens/max_tokens`, I noted in my own working query
that `max(max_tokens)` is the *highest* cap in the window and that caps had been
raised mid-window. Then I read the very next result as
"`council-gate/review_editquality` is at p95 **95.1% of a 16000 cap** — the raise
created NO headroom, the seat just wrote longer!" and started drafting that as a
finding, because it was a dramatic result that fitted the story.

It was an artefact of exactly the trap I had just written down. The p95 was computed
over rows at BOTH caps; the 8000-cap rows sat at ~95% of *8000*, and the displayed
denominator was 16000. The 16000-cap rows peak at **62.9%**. The raise worked.

The fix is to filter to the seat's CURRENT live cap, which is what the shipped
report does. The lesson is not about SQL: **naming a trap is not avoiding it**, and
a result that flatters the narrative you are already holding is the one to re-derive
first. Logged in `WRONG_CALLS.md`.

### MISSTEP 2 — the wrong-depth JSON path, again, one field over

Yesterday's runbook §4 warns that the cap is at `config.ai_service.max_tokens`, not
`config.max_tokens`. So today I read prompts from `config.ai_service.prompt_template`
— and got `prompt_template IS NULL` for **all 51 live seats**, a beautifully uniform
answer that reads as "these seats have no prompts". `prompt_template` is a SIBLING of
`ai_service`, at `config.prompt_template`. The cap is nested; the prompt is not.

Same silent-zero family as yesterday, and the near-repeat is the point: knowing the
*rule* ("watch the depth") did not help, because the rule does not say WHICH keys are
nested. Runbook §4 now names both paths together, which is the only form of that
warning that would have stopped me.

### Candidate 4: the premise is right, the fleet-wide prescription was already met

Filed as "emit the load-bearing field FIRST in every seat's output schema". I
surveyed all 51 live templates and measured what truncation actually destroys, and
three parts of the expected answer are **refuted**:

| claim | measurement | verdict |
|---|---|---|
| the head is wrong | `reviewer`,`verdict` are first in **51 of 51** | already right |
| `severity` last inside an objection loses grades | **0 of 2,713** stored objections ungraded | REFUTED |
| `notes` should move to the head fleet-wide | notes survives 2/30 truncations (6.7%) vs 3,067/3,076 (99.7%) — but objections survive **80%** of truncations and carry both the severities the gate reads and the content the proposer revises against | would make it WORSE |
| guardian's veto needs its contained alternative saved | **15 vetoes all-history, 0 degraded, 0 empty notes** | no observed instances |

The severity one is the one I would have got wrong by reasoning: `severity` really is
last inside each objection, a cut inside the long `problem` text really would lose
it, and an ungraded objection really does gate. It just never happens — the repair
keeps a whole objection or drops it. **A mechanism that is real at every step can
still have a zero rate, and only the count tells you.**

So there is no ordering that saves everything. The current order already sacrifices
the cheapest field, and `review_architecture` is the exception *because its own remit*
puts the mandated signal in notes — a seat-by-seat judgement, not a rule.

What generalises is the OTHER half of the architecture fix: the **length budget**,
the half the evidence credits (outputs got shorter — peak 4,443 tokens, 28% of the
new cap — rather than merely having a higher ceiling). Deliberately NOT generalising
its "at most 3 objections" clause: budgeting coverage across every council would
lose real objections invisibly. The shipped block budgets prose and says explicitly
to cut words, never findings.

### Coordination note — three of today's appends landed under another thread's message

`LANDMINES.md`, `WRONG_CALLS.md` and the concept index's two new rows were committed
by another session's `84a10e6aa` ("fix(pattern-check): the stdin-eater rule…", 14:08)
between my writing them and my own commit. **Nothing is lost** — all three appends
are in HEAD, verified by grepping HEAD's copy of each file — but they are attributed
to a commit about a different subject, and `git log --oneline -- LANDMINES.md` will
send a future reader to the wrong thread.

Recorded because CLAUDE.md asks for it explicitly, and because it is another instance
of the pattern it warns about: committing narrowly stops *me* sweeping *others'* work;
it cannot stop a session running `git add -A` from sweeping mine. The three files
involved are all **append-only shared registers**, which is the best possible case for
this to happen to — concurrent appends merge, so the cost is attribution only. The
same sweep across a source file would have taken a half-finished edit.

The residual asymmetry worth knowing: my `fix-loop.md` register ENTRIES are in my
commit `8387a9c44` while the matching `000_concept_index.md` ROWS are in theirs, so
the register's two halves for FIX-058/059 are split across two commits by two threads.
Both are present and consistent; only `git log` on either file misleads.

### 2026-07-31 — the alert paid for itself in 18 hours, and it paid on the threshold I nearly didn't build

Four ticks after being switched on, the task had written **2 notes, 2 distinct flag
sets** — so both halves of the design are now proven in production rather than argued:
the identical set at 21:21 and 03:21 wrote **nothing** (dedup by digest holds across
ticks), and the changed set at 09:21 spoke (an event, not a heartbeat).

**What changed is the interesting part.** A new seat crossed:

```
N  review_debug_historian@8000 — n=283 (holder 128), p95 62.2%, peak 99.8%, truncated 0
```

**Peak 99.8% of cap at a p95 of 62.2%.** A review came within ~16 tokens of being cut
off, on a seat whose typical output uses under two-thirds of its budget — 128 of those
calls attributable, so this is measurement and not inference. A p95-only rule would
have rated it 62.2% and said nothing, for ever.

That is the threshold split earning itself on live data less than a day after I
hesitated over whether two named thresholds were over-engineering. The reasoning I
wrote at the time — *truncation is a tail event, so the maximum is the primary signal
and the body of the distribution is a different question* — turns out to be the whole
value of the instrument. **Worth noting against my own habits: this is the opposite of
the day's two missteps.** Those were places where reasoning ran ahead of measurement;
this was a place where reasoning about the SHAPE of a risk (tail, not body) correctly
told me which measurement to take. The distinction is not "trust reasoning less" — it
is that reasoning is for choosing the question and measurement is for answering it, and
both of yesterday's errors were reasoning answering.

Added `review_debug_historian` to the length-budget script's targets on fix-proposer +
council-gate, which exercises the claim that adding a target is one line — the first
target this lane did not choose by hand.

### 2026-07-31 15:12 — the length budget is APPLIED, and one scare was my own check

7 of 7 targets written, one row each (the row count is the only check available —
`jsonb_set(create_if_missing := false)` makes a wrong path a silent no-op). All 7 now
read `APPLIED`; a second `--apply` reports "nothing to do", so idempotency is proven
rather than claimed. 099 roster dry-run: **drift (none)** — the mirrored pair stayed
identical, which is the thing that would have broken if I had targeted one council of
the pair. 102 parity lint reports 3 findings, all pre-existing `max_tokens` value drift
from the deliberate 16000 raises; nothing about prompts, nothing new.

Placeholders survived: 3 per fix-lane template, 4 on feature-designer, unchanged; the
block sits immediately before `## Output` with the JSON schema line intact after it.
Guardian went 6,174 → 7,205 chars (+1,031, the block).

**The scare.** I checked the snapshots with
`SELECT count(*) FROM agent_definitions WHERE is_snapshot AND created_at > now() - interval '20 minutes'`
and got **0**, having just watched the script print three successful snapshots. My
query was wrong, not the script: `snapshot_agent()` copies into
`agent_definitions_backup` and stamps `snapshot_taken_at`, and it copies `created_at`
verbatim from the source row so a recency filter on that column finds nothing either.
The real check shows all three, and — the part that actually matters — `has_block =
false` on each, proving they are **pre**-update copies and therefore a real rollback. A
snapshot taken after the write would look identical in every column except that one.

Worth its own landmine (added, footprint `snapshot_agent`) because every live-config
query in this repo filters `COALESCE(is_snapshot,false)=false`, which strongly implies
snapshot rows sit in `agent_definitions`. Two conventions coexist, and the wrong check
returns 0 rather than an error — a thread that stopped there would have concluded it had
no rollback.

**What is NOT yet verified, and cannot be today:** whether the block changes behaviour.
DB config is live immediately, but only rounds **spawned after 15:12:5x** carry it, and
none has run yet. Cutover times taken from `agent_definitions.updated_at`, not from
scrollback — RUNBOOK §10 has them and the post-cutover measurement query, with the
pre-cutover peaks to beat (guardian 99.2%, debug_historian 99.8%,
improvement_guardian 96.6%). The architecture precedent says outputs should get
*shorter*; if the peaks do not move, the block is being ignored and that is the finding,
not a null result.

### 2026-07-31 15:30 — owner raised feature-designer's caps, and the raise immediately exposed the argument it was made under

Owner call: `review_architecture` / `review_editquality` / `review_guidelines`
8000 → 16000 on **feature-designer**, matching the two councils the 07-29 ruling
reached. Applied as `sql_for_agents/277_…sql` — snapshot first, guarded on the current
value being 8000 (so idempotent and it cannot stomp a value someone else has set),
`create_if_missing := false` so a wrong path is a silent no-op, and the RETURNING
count as the check. 1 row, then all nine (3 seats × 3 councils) verified at 16000
inside the same transaction.

**Snapshot verified as a PRE-update copy**, per the landmine I wrote two hours ago:
`agent_definitions_backup` shows `arch_cap_in_backup = 8000` while live reads 16000. A
snapshot taken *after* a write differs in no other column, so this assertion is the
only thing separating a rollback from a souvenir.

Predicted and confirmed side effect: **102 now reports value drift on
feature-designer** (3 seats at 16000, 3 at 8000) exactly as it already did for
fix-proposer and council-gate. That is the deliberate kind, and saying so before
running it is the difference between a prediction and an excuse. 099 unchanged —
feature-designer is not in its mirror, which is the whole reason this gap existed.

#### And then the re-run found the thing that matters more

`review_editquality@16000` — the fix lane's, at the cap the 07-29 ruling gave it —
**flags at peak 98.3% with all 52 calls attributable.** Checked directly rather than
trusting the aggregate, because today has earned that caution:

| when | output tokens | % of 16000 cap |
|---|---|---|
| 07-30 19:36 | 13,115 | 82.0 |
| 07-30 19:48 | 13,592 | 85.0 |
| **07-31 14:52** | **15,721** | **98.3** |
| **07-31 15:30** | **15,525** | **97.0** |

Not an outlier — a **rising trend**, with the last two inside an hour. Earlier today
the same pair read peak 62.9% on 28 calls; 24 calls later it is at 98.3%.

**This is the sharpest measurement this lane has produced, and it is of its own core
claim.** The bug file has argued since 07-29 that "a cap raise MOVES the cliff, it does
not close it", evidenced by `review_architecture` reintroducing truncation against a
new cap within hours. That was one seat and could be dismissed as a longer prompt
landing at the same time. This is the *other* raised seat, with no prompt change,
growing into its doubled ceiling over three days on fully attributable evidence. **The
07-28 raise bought review_editquality roughly three days.**

Direct consequence for the change just applied: raising feature-designer to 16000 buys
time, not immunity, and nobody should read it as closing anything. Added
`review_editquality` on all three councils to the length-budget targets (10 now) —
cutover `15:39:26`–`15:39:30` from `agent_definitions.updated_at`. 099 drift none,
re-run idempotent.

**What this does NOT establish.** Whether the length budget will hold editquality
below its cap is unmeasured — only rounds spawned after 15:39:3x carry it, and the
seat's trajectory is steeper than any other. If its next peaks stay near 98%, the
budget is being ignored by the seat that most needs it, and the honest next step is a
per-seat instruction rather than a third cap raise. Watch it with RUNBOOK §10's query,
restricted to `step_name='review_editquality' AND max_tokens=16000`.

### 2026-07-31 ~18:00 — THE LENGTH BUDGET WORKS, measured on the seat that most needed it

A fresh chassis build deployed (replicaset `6fd67d6649`, pods started 17:45). Two things
became checkable.

**1. FIX-055 survived the rebuild.** Both new pods: `gating TRUNCATED objection from` 1,
`gated_by_truncation` 1, controls `all reviewers approve` 1 and `gating objection from` 1
— all in the same exec. This is the `bugs_open/153` check ("a roll is not evidence your
fix shipped; the image may predate your commit") run in the *other* direction: not "did
it arrive" but "did it survive". A new build from an older ref would have silently
removed a fix that was live an hour earlier, and nothing else would have told us.

**2. The behaviour change is verified — and my first attempt to measure it repeated the
exact mistake this RUNBOOK warns about.** I filtered `llm_call_log.created_at >
15:13:00`, which is the wrong clock: **an orchestration keeps the workflow definition it
loaded at SPAWN**, so a call made after the cutover can belong to a round spawned before
it and carrying the old prompt. It also silently mixed `review_editquality`, whose own
cutover was 15:39:30, into a filter built for the 15:12 batch. RUNBOOK §10 says both of
these in writing, and `council-adoption-report.sh` was once 45 minutes wrong the same
way. **Third time today that a check answered a different question from the one I
meant.**

Redone against `orchestration_states.created_at` — the round's spawn time — per seat,
against that seat's own cutover:

| `review_editquality` @16000 | calls | peak | peak % of cap | mean |
|---|---|---|---|---|
| rounds spawned **BEFORE** the block | 10 | 15,721 | **98.3%** | 9,848 |
| rounds spawned **AFTER** the block | 8 | 8,793 | **55.0%** | 6,569 |

**Peak 98.3% → 55.0%. Mean down 33%.** Both arms are the same afternoon at the same cap
on the same seat, so this is not a time confound and not a cap effect — the only
difference is the length budget. And the direction survives the obvious objection about
small samples: a maximum grows with n, and the arm with MORE calls (10, before) is the
one with the higher peak, so fewer samples cannot explain the drop.

This is the architecture-seat result reproduced on a second seat, and this time cleanly:
there, the cap raise and the length budget shipped together and could not be separated.
Here the cap did not move — editquality was already at 16000 — so **the budget alone
took it from the edge of truncation to just over half its ceiling.**

**What this does NOT show.** The other three seats (`guardian`,
`improvement_guardian`, `debug_historian`) have **zero** pre-cutover calls in this
window, so they have no control arm and I am not claiming an improvement for them.
Their post-cutover peaks (77.6%, 37.5%, 70.8%) sit well below their 14-day historical
peaks (99.2%, 96.6%, 99.8%) — but **a max over 14 calls is not comparable to a max over
278**, because a maximum grows with the sample. That comparison is exactly the shape of
error I have made twice today, so it is written here as a non-finding rather than a
result. They need their own before/after arms, which will accumulate.

Also unexplained and left alone: 3 `review_editquality` calls at cap 8000 after 15:29
with NULL `output_tokens` (failed calls), belonging to rounds spawned before
feature-designer's raise. [UNMEASURED] — noted rather than chased.

### 2026-07-31 ~18:20 — both owner decisions applied. Candidate 4 is finished.

**Decision 1 — `feature-designer/review_architecture` notes-first.** Applied via a new
`scripts/reorder-seat-notes-first.py`. The reorder is done by **parsing the output line
as JSON and rebuilding it with the key order I want**, not by regex surgery: the line is
syntactically valid JSON even though it is a template, and its `notes` value contains
`\n\n`, `|` and angle brackets that a regex would eventually mangle. All three
`review_architecture` seats now report an identical key order —
`reviewer, verdict, notes, objections, …` — and all three carry a length block (the
other two hand-written, correctly refused by the applier). Idempotent: a second run says
`ALREADY`.

**Decision 2 — the budget reaches every eligible seat.** 48 seats now carry it:

| | count | |
|---|---|---|
| eligible and applied | **48** | |
| HAND-WRITTEN, refused | 2 | `review_architecture` on fix-proposer + council-gate — they already have a hand-authored budget, and overwriting deliberate prose is not this script's job |
| excluded, with reason | 1 | `domain-research-classifier/review_mission_alignment` |
| **total live review seats** | **51** | every one accounted for |

099 drift: **none**. Snapshots: one per council touched (5), each verified to LACK the
block on a seat that run actually wrote — a snapshot taken *after* a write differs in no
other column, so that assertion is the whole difference between a rollback and a
souvenir.

**Two design points worth keeping.**

1. **The target list is now DISCOVERED, not written down.** A hand-maintained roster of
   48 pairs would be stale the first time a seat is seated — which is precisely why
   `102_LINT` exists (a 16th seat added, one key forgotten) and why `099` exists (two
   rosters that must be kept identical by hand). The script asks the database every run.
2. **Eligibility is a claim about the mechanism, not a convenience.** The block asserts
   that a degraded `object` gates the round regardless of severity. That is true only
   where `diagnose_council_decide` is the decider, so eligibility is defined as "the
   council has that step" — measured. `domain-research-classifier` has **zero** decide
   steps and a different output schema (`objection_found`/`concerns`/`note`), so the
   block would be a **false claim** in its prompt. Putting text a reviewer will act on
   into a prompt where it does not hold is worse than leaving the seat uncovered, so it
   is excluded and the reason prints on every run.

**Told, not merely measured.** Ten of the 48 seats belong to the experience-loop lane
(`experience-planner`, `experience-approval-council`). Notice written into their
`RUNNING_NOTES_experience_loop.md` per the 07-29 ruling's third limb — saying what
changed about their *guarantee* (shorter reviews, same number of findings; the block
says "cut words, never findings" explicitly), the evidence, the datum that concerns them
most (`review_deferral_honesty` truncated 3 of 5 calls at cap **12000**, the worst rate
anywhere and already above the default — direct evidence a bigger cap alone is not the
fix), and how to reverse it.

**What is now unverified at a larger scale.** The effect is measured on ONE seat. 47
more just changed behaviour, and the honest expectation is that most were never near
their cap and will show no difference — 24 of 31 (seat, cap) pairs sat below 75%. **The
thing to watch for is not whether peaks fall; it is whether OBJECTION COUNTS fall**,
which would mean the budget is trading coverage for brevity despite being told not to.
That is the failure this rollout could cause and the earlier narrow one could not:

```sql
SELECT date_trunc('day', d.created_at)::date AS day, count(*) AS reviews,
       round(avg(COALESCE(jsonb_array_length(CASE WHEN jsonb_typeof(r->'objections')='array'
              THEN r->'objections' ELSE '[]'::jsonb END),0))::numeric,2) AS mean_objections
FROM diagnosis_artifacts d, LATERAL jsonb_array_elements(d.body::jsonb->'reviews') r
WHERE d.kind='council_report' AND d.created_at > now() - interval '10 days'
GROUP BY 1 ORDER BY 1;
```
Cutover for the fleet-wide batch: **2026-07-31 18:16:46–18:16:53**
(`agent_definitions.updated_at`). Compare rounds spawned either side, by SPAWN time.
