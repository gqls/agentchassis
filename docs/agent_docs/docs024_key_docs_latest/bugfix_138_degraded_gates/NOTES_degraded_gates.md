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
