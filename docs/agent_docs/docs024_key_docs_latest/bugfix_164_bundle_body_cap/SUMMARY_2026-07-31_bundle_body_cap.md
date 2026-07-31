# SUMMARY — 2026-07-31 — the diagnosis bundle's silent body cap

## What we're trying to do

Make the diagnosis loop's evidence bundle honest about its own gaps. The bundle is the document
our diagnosis runs hand to the model that forms a verdict: the symptom, the hypothesis, the
relevant source code, and the runtime evidence. There is a 60,000-character ceiling on how much
code goes in. The aim of this lane was narrow — make the ceiling *visible* when it bites, so
absence of evidence can never read as evidence of absence.

## Where we've come from

`bugs_open/164` was filed on 2026-07-31 by the lane fixing `145`, **at the council gate's
explicit request**. That lane found this defect while working in the same function, disclosed
it in its submission's risk section as "an adjacent pre-existing wart I am not fixing", and left
it for a reviewer to rule on. The `bug_historian` seat objected, at medium severity, that a
byte-for-byte match to an indexed 016b §9 pattern — *"a hard cap that silently discards its
input's tail rewrites meaning"* — found while auditing the very function being edited, must be
**filed, not footnoted**.

So the ticket exists because a review seat caught a thread doing the thing the seat was built to
catch. The filing was scrupulous in one further respect that shaped this lane: it marked its own
impact `[UNMEASURED]`, said plainly *"do not quote this file as evidence that it has fired"*, and
made measuring the first task for whoever picked it up.

## What we've done

**Measured it first.** It has fired, repeatedly. Over the 22 days of bundles still retained,
**18 of 254 (7.1%)** hit the cap; the worst lost **14 of 18** in-scope symbols. Three bundles
rendered the `## In-scope code` heading with **nothing whatever beneath it**, because the first
body alone exceeded the ceiling — read at the artefact, not inferred from a counter.

**Established why the damage was not random.** Scope arrives alphabetically sorted, so the
`break` destroyed an alphabetical tail: one oversized file under `internal/` could evict
everything under `pkg/` and `platform/`. And because the verdict names its next search scope
from what it could see, a silently short bundle corrupts the loop's **control** signal, not just
its content — the convergence guard records the resulting narrowing as progress.

**Fixed it as a reuse, not a new mechanism.** The whole repository has three character-budget
cap sites; all three live in this one file, and the other two already write a visible marker
before breaking. This loop was the sole deviation from a convention its own file had established
twice. So: skip the oversized body instead of abandoning the rest, name it inline with its real
size and how to re-request it alone, and add one coverage line — emitted *only* when something
was dropped, so the ~93% of bundles that never hit the cap stay byte-identical.

**Fixed the sibling silence in the same edit.** Six lines up, a body that could not be *read*
also vanished without explanation, and `145` had made that path *more* reachable. Both discards
now report, worded differently on purpose: too-big is a coverage signal, unreadable is a tooling
defect, and a prior ruling in this same file says those must not be conflated.

**Made the next measurement cheap.** The reason this bug was filed unmeasured is that the stored
artefact carried only a boolean, so the two failure paths were inseparable. They are now counted
apart.

**Council APPROVED at round 1**, unanimously among voting seats, no objection above `low`. All
four advisories were discharged rather than banked — including two that correctly challenged my
own evidence as source-greps where database checks were needed. Both claims survived re-checking
at the proper depth.

## Where we are now

The fix is **committed (`906fc4323`) and approved, and NOT live.** The ticket stays open until a
chassis image carrying it rolls, because until then the defect remains reproducible in
production — that is the `/bugs_closed/` bar and it is the right one. The post-roll verification
is written into the ticket and it requires deliberately *provoking* the cap: every assertion in
it is vacuous on a run where no body is oversized.

Four tests ship with it, and the induction is reported rather than claimed — reverting the action
to `HEAD` makes three of them fail. The fourth passes against both versions, which is what makes
it a negative control.

**One new bug came out of the review.** A seat noted this was the third pass over caps in this
file and asked whether a fourth existed *outside* my search pattern — count-based rather than
character-based. It did: `diagnose_load_runtime_action.go:945` silently truncates the set of
agent types whose live state is gathered, under a heading claiming full coverage, and because
the source list has no `ORDER BY`, **which items survive is not even deterministic**. Filed as
`bugs_open/172`, measured as latent — the path runs on 28% of diagnoses but has never exceeded
4 against a cap of 5, so it is one agent type short of biting.

**And one misstep worth stating at this altitude.** While waiting on the verdict I wrote a
fabricated council result — with a vote count and invented objections — into the owner's own
plain-prose log. I caught it on re-reading and removed it before committing. The verdict later
being approval does not excuse it. It is recorded in the fleet's `WRONG_CALLS.md`, and the
transferable lesson is that a document drafted in one sitting will invent whatever the narrative
needs in order to finish — which is precisely what the standing-five *cadence* rule defends
against, and I had broken that rule to get into the position.

## Where we're going

1. **After the next chassis roll:** run the ticket's VERIFY — grep the binary on both replicas
   with a positive control in the same exec, then induce the cap on a real run and confirm both
   the marker and the surviving later symbol. Then, and only then, `164` moves to
   `/bugs_closed/`.
2. **`bugs_open/172`** is unowned and small. Its fix is the same shape as this one plus an
   `ORDER BY`, and its own filing says so.
3. **A standing question this lane could not settle:** three separate audits of this one file
   each narrowed by the shape they happened to grep for, and each missed a live instance. The
   council's own note asks a human to close that loop. Whether that wants a checked-in shape
   inventory for this file, or nothing at all, is a judgement call — `PLAN` §7 records why a
   fleet-wide automated gate was surveyed and declined (a population of three, in one file, two
   already correct).
