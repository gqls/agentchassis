# SUMMARY 2026-08-23 — the editor ships copy, and the brief defect is closed

Sixth in the series (08-12 *why the copy reads wrong* → 08-14 *the house voice ships* → 08-15 *plan
complete, four decisions* → 08-17 *the editorial pass exists and showed restraint* → 08-19 *the fault
was never in the writer* → this). Written because everything the 08-19 entry listed under "where
we're going" has now either landed or been handed to someone else, which is the test for whether a
new one is due.

## What we're trying to do

Make the copy on our sites read like a person wrote it. The machinery that builds a page is graded on
facts, coverage, structure, links and styling — never on whether the result is worth reading. Design
faults had a fixer; copy faults had none.

## Where we've come from

A house voice that lived in seven drifting copies, then one row every writer reads. Then a second,
editorial pass that reads a whole page and proposes changes — built without new code, because
everything it needed already existed. Then, on the 19th, the finding that reframed the lane: the
fault the owner complained about was **in the brief**, not the writer. And underneath that, found
while building the tool to measure it, a defect that had been quietly running for four months.

## What we've done

**The brief defect is fixed, live, and closed.** A site's page brief is a document of about twenty
parts, and the writer reads only a single derived summary of it. That summary was being rebuilt from
*only the part being edited*, so any narrow correction silently deleted the other nineteen from the
writer's view while the document still read complete. It is fixed at the source, it shipped, and — the
part that took longest — it is now **proven on real writes**, not just present in the binary: two
sites were written by two different producers this week and every part of their briefs reached the
writer. The bug is closed.

Getting there cost two review rounds, and the reviewers were right twice. One caught that the fix
rested on an assumption I had asserted rather than tested; that assumption now has a test that fails
when you break it. Another caught that my instructions for checking whether the fix had shipped
pointed at two machines out of seventy-five.

**The editorial pass now ships copy, not just proposals.** Its three approved edits are live on the
AI-orchestration homepage — the call-to-action is down from 733 characters to 496, with the version
that ran the same statistic past the reader twice gone. Verified on the served page, not just in our
database.

**And we answered the question we had left open about it:** the three-edit limit is not a ceiling on
what the editor can see. Run it again and it finds three *different* things — no overlap at all with
the previous run. It also declined to pretend three edits were enough, reporting that the real problem
on that page is two sections that duplicate each other and needs a structural merge rather than
editing.

## Where we are now

**The lane's two halves both work end to end**, and the third — the input side — has been taken up by
another team, who are building a writer-side gate using our diagnosis and our tooling as its
specification. We reviewed their design and sent back a measurement worth having: the writer's own
prompt demonstrates the mannerism sixteen times per call, so the pattern they were least sure about
may be an artefact of the instrument rather than a fault in the writer.

**What is deliberately not done:** three sites still serve a shrunken brief. The fix stops the damage
spreading; it repairs nothing by itself, and those are other teams' sites. The owner ruled that it
stays with them, and the trap that fires when they next touch those briefs — including that one
site's restored examples teach the very mannerism he objected to — is written where a session meets
it before acting rather than after.

**Three separate teams have now asked for the same missing piece:** a small agent that takes one named
component and one named defect and makes exactly that change. We wrote the specification so nobody
derives it a fourth time, and did not build it.

## Where we're going

**The editor still runs only when a human fires it.** Nothing routes work to it, by choice. Wiring the
quality auditor's findings into it is the obvious next step, and it is the step that turns a tool we
trust into a tool that works while nobody watches — which is also the step that most needs a human
canary, because volume is where a proposal-only, approve-everything posture stops being affordable.

**And the honest limit on everything above:** we can prove a phrase handed to the writer comes back
out — we have traced one word for word. We still cannot say whether the *form* of a brief shapes the
writing independently of the phrases in it. The team building the writer-side gate will produce
exactly the evidence that question needs, and it should be answered with its refutation condition
written down first, because this lane has already published one answer to a question of that shape and
had to withdraw it.
