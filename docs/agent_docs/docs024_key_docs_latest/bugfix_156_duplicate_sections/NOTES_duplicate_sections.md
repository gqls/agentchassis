# NOTES — bugfix 156, the prevention half

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-04 — picking the bug, and how many were already taken

The fleet is busy: 35+ session transcripts touched in the last three hours. Choosing an
unowned bug was most of the opening cost, and `who-owns.py` alone was not enough, because it
reads **commits** and a session mid-fix is invisible to it.

What I actually did: greped every `.jsonl` transcript modified in the last 180 minutes for
`bugs_open/NNN` and counted. That is the check the memory line "EVERY ownership check is
LAGGING — grep live `.jsonl` transcripts" describes, and it changed the answer three times:

- **093** (stat audit, one guarded call site) looked ideal — unstarted, structural, council
  verdict already written. Rejected: its own post-roll triage says it is **blocked on
  `bugs_open/083`**, the check shipped and has never run because `improvement-sweep` is off,
  and 083 has an active thread (11 mentions in a live transcript). Fixing 093 today would
  produce another correct mechanism behind something that never runs.
- **071** (the gate detects every broken link then discards it) is the highest-value bug in
  the directory and `who-owns.py` returns **OWNED or recently active** on it, naming two live
  workstreams. Not mine to start a competing fix on.
- **156** — zero mentions across every live transcript, last commit 2026-07-31, and the file
  itself says of the remaining half: *"What remains, and is unowned."*

## The premise moved between filing and today, and in the direction that matters

The bug file's census: 17 duplicate `(page_id, slot_name)` groups, 6 of them content-identical
(vonc). Re-run today: **12 groups, 0 content-identical.** The vonc rows were fixed on 07-30.

So the honest framing of this lane changed before I wrote a line: **there is no damage to
repair.** The guard is judged entirely on whether it could ever destroy a live section, and
its value is preventing a recurrence of one measured incident. I would rather state that than
let a reader infer the guard is cleaning something up.

## The bug file's own fix candidate is wrong, and its own footnote says why

`156` candidate 1 proposes collapsing on `(slot_name, md5(content_data))`. Its own census
footnote records that finetuning.uk/our-position-on-ai has two rows whose `content_data` is
**NULL on both**. Under the proposed key those two are "identical" — so shipping the file's
literal recommendation would have deleted a live section on a shape the file itself flags,
one paragraph apart. The two halves were written for different purposes (the footnote warns a
future *measurer*, the candidate instructs a future *fixer*) and nothing joined them up.

Corrected rule, and the reason it cannot over-collapse: collapse only when **every column the
INSERT binds is equal, `position` excluded**. Then the collapsed row would have been
indistinguishable from its survivor in the database. `rendered_html` in the key is what fixes
the NULL-content case; `component_id` in the key is free narrowing.

## What the planning pass found that I had not

Ran the plan through `fable`. Three things came back that I had missed on my own read, and
one correction to my own brief:

1. **My brief said `save_page_sections_action.go` was dirty in the shared tree** (the 194
   lane's WIP) and I designed around a same-file collision. By the time the plan came back
   that session had committed and the file was clean. Stale-by-minutes, exactly as CLAUDE.md
   warns about session-start `git status`. The sibling-file shape is still right — it is the
   house pattern — but the collision argument I built it on had evaporated. Re-checked before
   editing rather than trusting either reading.
2. **The locked-slot path manufactures a duplicate of locked copy.** With a doubled list the
   first copy of a locked slot consumes the lock and is discarded (`lr.consumed = true`); the
   **second copy falls through and is INSERTed beside the locked row.** Nobody had recorded
   that; it is an independent argument for collapsing before the insert loop rather than
   relying on a post-hoc detector.
3. **A doubled list halves the content-regression guard and doubles the completeness floor's
   numerator.** Both guards are currently measuring a lie whenever this defect fires: a page
   truly cut to 13% of its text reads as 26% and passes the 25% floor. So the placement is
   not a style choice — it is what makes four other guards measure the truth.
4. **Plan-guard parity.** The repair half refuses to delete a repetition the effective plan
   specifies (council-forced, owner decision 07-31). A save-time collapse that ignores the
   plan would make the two halves disagree about the same question. Same helper, same per-slot
   accounting — but the **opposite failure direction**, because the repair's conservative
   move is "refuse to delete" and a collapse guard's is "refuse to collapse".

---

## 2026-08-04 (later) — implementation

Written as `save_sections_dedup.go` (pure seam + DB seam + record) with the call site at
`save_page_sections_action.go` immediately after the arrival diagnostic. Test file carries a
named mutation per test in its header, because a test that could not have come out otherwise
is not evidence.

---

## 2026-08-04 — the seven mutations, run rather than imagined

Each produced exactly the predicted failure, and each restored file diffed byte-identical to
the pristine copy afterwards (the file is untracked at that point, so `git checkout` cannot
restore it — back it up to the scratchpad first, and `diff -q` the restore).

| mutation | tests that went red |
|---|---|
| identity reduced to `slot_name` (**the rejected unique-index rule**) | LegitimateRepeatedSlots, NullContentDataWithDifferentHTML, DiffersOnlyByComponentID, IdentityJoinIsUnambiguous |
| `rendered_html` dropped from the key (**the bug file's own candidate 1**) | NullContentDataWithDifferentHTML — **and only that one** |
| guard returns its input unchanged | 6 tests |
| `planned` map ignored | both plan tests |
| `content_data` marshalled unconditionally (no nil/empty → NULL) | NilAndEmptyContentData |
| adjacency signature hardcoded to `adjacent` | NonAdjacentDuplicates |
| marshal error falls through to the NULL sentinel | AbstainsWhenUnmarshallable |

The second row is the one worth keeping. It fails **one** test and no others, which is what
tells me the suite discriminates between the shipped rule and the documented wrong answer
rather than merely between working and broken code.

## The measurement I nearly wrote down as evidence, and why it would have been worthless

I was about to run the shipped guard over the whole live corpus and record "0 false positives
fleet-wide". **That number could not have come out otherwise and would have proved nothing.**
The census had already established that no live page has two rows with matching
`md5(content_data::text)`; my rule requires that match **and more**; so the result is 0 by
arithmetic, not by measurement. This is exactly the trap the standing rule names — a
`[MEASURED]` figure is only evidence if the disconfirming result was reachable.

**The version that IS disconfirmable, and why:** the census used the DB's own
`md5(content_data::text)`, and my key marshals a jsonb-read **Go map**. Those are different
rulers. Numeric representation in particular can round-trip differently (jsonb `1.0`, Go
`float64(1)`, `json.Marshal` → `1`), so two rows the census called *different* could still
collide under mine — a false positive on live data that no unit test in the package would see.
That question has a reachable "yes".

Ran it with the shipped `collapseDuplicateSections` (a throwaway in-package test reading a
JSON dump; the code under test is the shipped function, not a SQL approximation of it) over
**every** row that could possibly match — slot equality is necessary, so the duplicate
`(page_id, slot_name)` groups are the COMPLETE population, not a sample:

```
live corpus: 28 rows over 12 pages, 0 collapsed
positive control: a real live row (7102a697…/generic-text-block) duplicated verbatim IS collapsed
```

**The positive control is not decoration.** Without it, "0 collapsed" is indistinguishable
from a harness that never called the guard — which is the failure shape this lane's own
WRONG_CALLS neighbours are full of. The harness was deleted after the run; the corpus would go
stale within days and a committed copy would be a fact nobody re-measures.

## Two things I got wrong, both logged where they belong

1. **I shipped the code one commit before its concept-register entry** — condition (2) of the
   platform-seams ruling, the one thing it still requires. The `bugs_open/190` lane made the
   identical mistake hours earlier and wrote it up, and **I read their write-up while deciding
   whether my guard needed an entry at all** — took the fact I wanted from it and walked past
   its lesson. In `WRONG_CALLS.md`, with the argument that two lanes on one rule on one day is
   a missing control rather than two mistakes.
2. **I nearly made a pathspec commit that would have broken HEAD for the whole fleet.** The
   190 lane's call site was sitting in `save_page_sections_action.go` while its defining file
   was staged-but-uncommitted; a pathspec commit takes their hunk and not their file, so HEAD
   would have carried a call to an undefined symbol while my working tree built green. Caught
   by `git ls-files` on the callee — not by anything that failed. Waited seven minutes for
   their commit instead. Now a LANDMINES entry, because it fires with no symptom and the local
   build is green throughout.

---

## 2026-08-04 — council APPROVED r1, and the pre-roll baseline measured with controls

Verdict `1a3f4f27`: *"approved with 3 advisory objection(s) — none high-severity"*. Two
answered with evidence (`ComponentName` IS bound to `slot_name` at the INSERT's `$4`;
`SectionIdentityKey` + `PlanSpecifiedSectionCounts` are the only two such functions fleet-wide,
by grep — deliberately not by `code_checks`, whose index is frozen at `d98010e8b` and reads a
symbol added today as absent). The third, `bug_historian` at MEDIUM, is left OPEN and recorded:
one guarded call site of seven, no DB-level invariant, and my own scope boundary is *a fact
about current callers, not an enforced mechanism*. It is right, and the answer is a DB-side
generated column, which is an owner call and its own lane.

**Pre-roll baseline, both replicas, with a positive AND a negative control in the same exec:**

```
agent-chassis-5455ddcdcc-crnb6 / -gpr92
  CONTENT_DUPLICATE_SECTIONS_COLLAPSED : 0    <- new; 0 before the roll, ≥1 after
  DUPLICATE SECTIONS COLLAPSED         : 0    <- new
  CONTENT_DUPLICATE_SECTIONS_REFUSED   : 0    <- NEGATIVE control (string this change never adds)
  CONTENT_CLAIMS_FLOOR_DETAIL          : 1    <- POSITIVE control (already live)
```

The positive control is what makes the two zeros mean anything: without it, "0" is equally
consistent with the grep being wrong, the pod being the wrong one, or `strings` failing. **So
the bug is measurably NOT live and stays OPEN** — the bar is fixed AND live, and a roll here is
whole-fleet and the owner's.

## Not done, and named rather than left implicit

- **The behavioural induction.** Cannot be run until the code is in a pod. The recipe and its
  negative control are in RUNBOOK R5; the reason it is not optional is that a guard which
  collapses nothing is indistinguishable from one that never ran.
- **Finding the producer** (`156` candidate 4). Needs a fresh reproduction. The record this
  guard writes is what makes it possible; nothing about it is possible today.
- **The DB-level invariant.** The council's open objection. Owner call.
