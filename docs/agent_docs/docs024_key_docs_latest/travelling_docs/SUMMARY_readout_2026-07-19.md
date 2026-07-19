# Self-verifying tools — a read-out

*Written 2026-07-19. Plain-language, meant to be read aloud. Covers: what this
subproject is, where it came from, what we built, where we are today, and what
comes next. Detail lives in `RUNBOOK_travelling_docs(39).md` §0 (position),
`HANDOFF_2026-07-10_stage5_live_and_next_fronts.md` (turn-by-turn), and
`016b_debugging_guide_8_consolidated.md` §9/§10 (patterns + bug index).*

---

## What we are doing

We are making the site-building agents accountable for their own work.

The platform builds and maintains websites with autonomous agents. The problem
we set out to solve is simple to state: **an agent could report success without
having done the job**, and nothing noticed. A work item said `complete`, a page
looked plausible, and a broken tool sat live on a customer site until a human
happened to look at it.

The answer we have built is a loop. Every tool the system creates writes its own
documentation at birth — what it is for, and crucially **its own definition of
"working"**, as machine-checkable acceptance criteria. A scheduled sweep later
picks that tool up, drives it in a real browser on desktop and mobile, and
decides whether it still meets its own criteria. If it does not, the system works
out **whose fault it is**, files the right repair ticket, lets the fixing agent
repair it, redeploys, and checks again. Every step writes a note into a permanent
record attached to that tool, so the next agent to touch it inherits the history
rather than re-deriving it.

The short version: **the tools document themselves, test themselves, and fix
themselves — and the record of what happened is written by the machine, not by
us.**

## Where we came from

The trigger was a specific embarrassment. A game on one of our sites had two
plain bugs — a slider that did nothing useful and a chart line stuck flat on the
axis. The tool had passed every check we had. That is the whole problem in one
sentence: **"passed the checks" is not "works."**

The early work established the foundations: documentation that travels with a
tool, written by the agents themselves; acceptance criteria captured at the
moment of creation; and a ladder of verification, from cheap static checks up to
driving the real page in a real browser.

Along the way we fixed the things that got in the way — a memory crash that
turned out to be an infinite loop rather than a leak, a migration system so
database changes stopped being ad-hoc, and a series of traps that are now written
down so nobody pays for them twice.

## What we have done

**The loop is complete and it has been proven on real bugs, not test cases.**

The proving run was a genuine mobile-layout defect on a live site. The system
found it, worked out that it belonged to the site's shared footer rather than the
tool, caught its own first repair attempt *lying* about success, traced the fault
to the durable template layer, fixed it, redeployed, and confirmed the failing
check now passed. Humans chose between well-framed options; they never touched
the mechanism.

Since then we have added:

- **Photographic evidence.** A failing check now photographs the live page, and
  the verdict carries a durable link to that image.
- **Precise blame.** This is the subtle one. The system used to report the
  *widest* broken element — which is usually an innocent container that merely
  inherited the problem. The fixing agent kept "repairing" that container, twice,
  and the fault never moved. It now drills in and names **the element that
  actually causes the problem, and why** — for example "this grid's items refuse
  to shrink; allow them to". The very next repair attempt got it right. That
  turned a loop that spun into a loop that converges.
- **Current-generation models** across the whole tool pipeline.

**And then the loop caught something serious.** During that last run, the
repairing agent hit its output limit mid-sentence and **saved the truncated
fragment straight over a working tool** — a 10,272-character component became
1,253 characters of stylesheet with no markup and no code — and reported success.
The live page survived only because it had not yet been re-rendered from the
damaged source. We restored the component, raised the limits, and — because three
other threads had hit the same underlying mechanism in different agents — the
platform now has a guard that **refuses to save a rewrite that has lost its
structure**, and detection of truncated model output at the source.

That is the deeper theme of this subproject: **a system that checks its own work
finds the bugs that a system reporting its own success never will.**

## Where we are now

Everything described above is live in production (chassis and browser adapter
`v1.0.1137`), and the protection chain around it was completed today:

- The verification ladder runs unattended: discovery sweep → real-browser run on
  desktop and mobile → verdict → routed repair → re-check.
- Blame is precise: tool defect versus site chrome, and within a defect, the
  element actually responsible.
- The machine has written **8 tool PLANs, 125 notes and 12 browser verdicts** so
  far — all without a human authoring any of them.
- A fix can no longer destroy the thing it is repairing: the write guard is live,
  and as of today a refusal is recorded as **"needs human review" with a written
  reason**, rather than a silent failure.

One tool — the loot-table balancer on the games site — is **deliberately left
with a minor mobile overflow**. It is the reproducible benchmark for the
convergence work; fixing it by hand would erase the evidence.

Honest caveats, because they matter more than the wins:

- The final green re-check on that benchmark tool has not been observed yet. The
  repair was correctly aimed but the run that produced it was the one that
  truncated. That is the next thing to watch.
- A non-converging loop still has no automatic stop. If the fixer cannot solve
  something, it will keep trying on a weekly cadence rather than escalating.
- Two of my own earlier claims about the guard were wrong and were corrected by
  the thread that wrote it. That correction is itself the system working: the
  shared bug files are how threads catch each other.

## Where we are going

In rough priority:

1. **Watch the benchmark tool go green** — the one outstanding proof.
2. **A convergence guard**: after N repair cycles that do not fix the criterion,
   stop and ask a human instead of retrying forever.
3. **Widen the write guard.** It protects the component table; the rendered-page
   tables share the same unguarded shape, and that is filed as the next incident
   waiting to happen.
4. **The Experience Loop** — a sister workstream, already building on this one.
   Where we verify that *a tool works*, it verifies that *an experience is
   coherent*: that a button's promise is kept, a journey has no dead end, and the
   numbers on a page are real. It reuses this machinery wholesale and extends the
   same travelling documents with a new subject type. The two fit together as
   layers, not rivals.

The mechanism is done and proven. What remains is polish, widening the same
guarantees to neighbouring surfaces, and building the layer above.
