# Summary — an architecture seat for the council (2026-07-27)

First summary in this workstream. Current state only; the chronology is in the
decisions file and in `WRONG_CALLS.md`.

---

## What we're trying to do

We want to be sure the foundations of the platform are sound — that they are
robust, that they don't keep shifting under everyone's feet, and that they are
good enough for what we're planning to build next, not just for what we've
already built. Those three goals pull against each other. Making the system the
best it could be for each new decision means rewriting; keeping what we know
works means not rewriting. Both instincts are right, and something has to decide
between them, case by case, without simply defaulting to whichever voice is
loudest.

The specific question was whether that something should be a new member of the
review council we already run, or a process of some other shape.

## Where we've come from

We already had half an answer and didn't fully realise it.

One of the sixteen reviewers on the council — the one we call the guardian — is
the only seat with a hard veto, and its written brief already tells it to protect
long-stable, battle-tested infrastructure and to prefer fixing things at a higher,
less foundational layer wherever possible. That is, almost word for word, the
"don't let it change under us" half of the question.

Two days ago that seat vetoed a large delivery-guarantee rewrite twice, on the
grounds that a change of that size was an architecture change dressed up as a bug
fix, and that we had no review process sized for one. You agreed, and created an
architecture-review track: a short written procedure with a template covering
blast radius, staged rollout and rollback. It has one entry so far. But as a bug
file written yesterday put it bluntly: there is no architecture-review *agent* —
"architecture review" means you.

So the gap was never the conservative half. It was that nothing in the system
argues the other side: that a design won't carry us where we're going, and that
the cost of not changing is real too.

## What we've done

We researched what already existed, measured two things that had never been
measured, and got corrected twice along the way — both corrections mattering more
than the original claims.

The first correction was yours. I told you the council only ever reviews plans
that already exist, so a forward-looking seat would have nowhere useful to sit.
That was wrong: there are three separate councils at three points in the
lifecycle, and two of them run before any code is written. The seam I said was
missing is exactly where the guardian already sits.

The second correction came out of the measurement work. I told you the council
couldn't read its own past verdicts. In fact it can — every verdict it has ever
reached is stored in a table the reviewers are already allowed to query, two
hundred and fifty-nine of them, with every objection preserved in full. The
problem is simply that not one of the thirty-two reviewer briefs mentions that
the table is there. The memory has been on the shelf the whole time and nobody
told the seats.

Then the measurements. The first was the "does it change too often beneath us"
worry, and it refutes itself. The orchestrator looks like it churns wildly — over
three hundred and sixty commits in two months — but almost all of that is new
entries in a plug-in registry, which is exactly what a registry is for. The
actual core moved fifty-five times, and its central file nine times, against
two thousand one hundred and twenty-three commits across the repo. The
foundations are not moving. The architecture's own shape is what keeps them
still.

The second measurement was the opposite worry, and this one is real. Across the
council's history the guardian has issued four hundred and thirty-seven
objections, twenty-nine of them invoking that "fix it at a higher layer"
preference — and they land on the same few places over and over. One file in the
orchestrator was sent back upstairs by six separate, independent submissions
inside seven days. The agent-spawning path, four. Four bugs are still open in
that same core right now. One objection in the set even records the guardian's
own suggested alternative being refuted by evidence: the safer higher-level fix
it named did not actually exist.

Put beside the churn numbers, the two measurements agree. Pressure to change the
core is high and rising; actual change to the core is near zero; and the
difference is being absorbed as workarounds in the layer above.

## Where we are now

The question that gated everything — is any of this worth staffing? — has been
answered, with evidence, and the answer is yes. But the risk we've confirmed is
the opposite of the one we set out to check. It isn't churn. It's ossification: a
core that has become expensive to touch, with the pressure diverted upward
instead of resolved.

You've ruled that the guardian's veto survives as a hard block, and I think that's
right. The seat isn't the problem. The problem is that it has been asked to judge
whether a change is worth its risk while having no way to see that it has already
sent this exact file upstairs five times before, and no counterpart whose job is
to argue what not changing costs.

One concrete fix is written, verified and waiting: a paragraph added to five of
the reviewer briefs telling them their own minutes exist, plus a specific
instruction to the guardian to count how often a site has already been deflected
before deflecting it again. It touches prompt text only — no new steps, no schema
change, nothing that needs a rebuild — and it can be undone with one command. It
is blocked purely because writing to the live fleet configuration needs your
permission, not mine.

We also now know why the historians on the council have never learned anything
from our written record: it's all markdown, and markdown is invisible to them.
The debugging guides, the bug files and the wrong-calls log come to about three
and a third megabytes across a hundred and twenty-four files, which is far beyond
what any reviewer could be shown. Your instinct that this needs a different agent
rather than a bigger prompt is correct. The opening is that one of those guides is
already written as a list of one-line, dated lessons — an index of just the
headings would be small enough to hand over directly, and would also replace the
seven hand-typed examples currently frozen into the bug historian's brief.

## Where we're going

Five things need your decision. The first is immediate and the rest are sequenced
behind it.

**One — release the staged change.** It gives the council its own memory back and
gives the guardian the deflection count. Everything is verified; it needs you to
run one command, then the mirror that copies it across to the second council.
This is the cheapest useful thing on the list by a wide margin.

**Two — should the guardian still weigh benefit at all?** You left this open, and
holding it was right. My inclination is that it should not: it has no instrument
for measuring benefit and its judgement has been overturned every time it was
escalated. But the honest argument against is that risk and benefit aren't
separable — "is this worth it" is the question a veto answers, and a seat judging
blast radius alone would have to block every wide change, which is more
conservative, not less. This one can't be settled until a forward-looking seat
exists to supply the other half, so it stays open deliberately.

**Three — do we add that forward seat, and where?** The measurement now says yes.
My proposal is a seat that argues sufficiency and the cost of not changing,
advisory with no veto, whose verdict is "this needs an RFC" or "this is fine as a
point fix" rather than an objection. It should sit at the design-stage council,
because that is both the earliest point platform code takes shape and the place
where the guardian currently sits with the fewest counterweights — four
colleagues, none of them looking forward.

**Four — how do the historians get the written record?** Three options, from a
small heading index, through a dedicated retrieval agent that searches the full
corpus and hands back the three most relevant past cases, to indexing all our
markdown properly. I'd do the heading index first: it's small, it makes the
historian's stale list self-maintaining, and building the retrieval agent first
would mean retrieval over a pile nobody has indexed.

**Five — make the RFC trigger automatic.** The architecture-review track has a
clear four-condition test for when a change needs an RFC, but nothing fires it —
which is why its own first entry was written after the code was already running
in production. Most of that test can be computed from the staged changes at commit
time, reusing the scope report that already runs on every commit and already reads
exactly that information.

There is one dependency underneath all of this worth naming. A seat asked whether
the architecture is sufficient for our anticipated plans needs those plans written
somewhere it can actually read. At the moment they're spread across some forty
workstream directories, and — given what we've just learned about markdown being
invisible — a roadmap document would not have solved it. It has to live where the
reviewers can query it.
