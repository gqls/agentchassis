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
