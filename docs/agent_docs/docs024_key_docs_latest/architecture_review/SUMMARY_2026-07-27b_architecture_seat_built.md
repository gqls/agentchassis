# Summary — the architecture seat, built (2026-07-27b)

Second summary today, and a new file rather than an edit of this morning's. The
first one ended with five decisions open; four have now been taken and three of
them built, so "where we are now" genuinely differs. The morning's summary stands
as the record of what we believed before the work was done.

---

## What we're trying to do

We want confidence that the foundations of the platform are sound: robust, not
shifting under everyone's feet, and good enough for what we intend to build next
rather than only for what already exists. Those goals conflict. Making the system
the best it could be for each new decision means rewriting; keeping what we know
works means not rewriting. Something has to arbitrate, case by case, rather than
defaulting to whichever instinct shouts loudest.

## Where we've come from

We already had one side of that argument and not the other.

The council's guardian seat — the only one with a hard veto — is told to protect
long-stable infrastructure and to prefer fixing things at a higher, less
foundational layer. That is the "don't let it change under us" half, and it has
been doing its job. Two days ago it vetoed a large delivery rewrite twice, which
is what caused you to create the architecture-review track in the first place.

What nothing did was argue the other side: that a design will not carry us where
we're going, and that not changing costs something too. As a bug file put it
bluntly this week, there is no architecture-review agent — architecture review
means you.

We also, it turned out, had two beliefs about all this that were simply wrong, and
both were corrected before anything got built on them. I told you the council only
ever reviews plans that already exist, so a forward-looking seat would have nowhere
useful to sit; in fact there are three councils at three points in the lifecycle,
two of which run before any code is written. And I told you the council couldn't
read its own past verdicts; in fact it always could — two hundred and fifty-nine of
them, in a table the reviewers were already allowed to query. Nobody had ever
mentioned the table in any of the thirty-two reviewer briefs.

## What we've done

First we measured, because the whole exercise rested on a premise nobody had
tested.

The "does it change too often beneath us" worry turned out to refute itself. The
orchestrator looks like it churns wildly, but almost all of that is new entries in
a plug-in registry, which is what a registry is for. The core moved fifty-five
times in two months and its central file nine, against two thousand one hundred
and twenty-three commits across the repo. The foundations are not moving.

The opposite worry is the real one. Across the council's history the guardian has
made four hundred and thirty-seven objections, twenty-nine of them invoking the
"fix it higher up" preference — and they land on the same few places repeatedly.
One orchestrator file was sent back upstairs by six separate, independent
submissions inside seven days. The agent-spawning path, four. Four bugs are open
in that same core right now, and one objection on record shows the guardian's own
suggested safer alternative being refuted: the higher-level fix it named did not
exist. So pressure to change the core is high, actual change is near zero, and
the difference is being absorbed as workarounds in the layer above. That is
ossification, and it is measured rather than suspected.

On that basis you took four decisions, and three things are now built.

**The council has its own memory back.** Five reviewer briefs — the guardian, both
historians, the prior-art librarian and the reuse seat — now say that the table of
past verdicts exists and how to read it. The guardian additionally gets the
measurement above as an instrument: before it sends a change to a higher layer, it
is told to count how often that site has already been sent upward, and that a site
which keeps returning is evidence the deflections are not holding. That went live
this morning across both councils that share the roster.

**There is now a seat that argues the other side.** It sits on the design-stage
council, which is the earliest point platform code takes shape and where the
guardian previously sat with only four colleagues, none of them looking forward.
Its job is forward fitness and the cost of not changing: whether the design carries
the work we can already see coming, and what the running cost is of taking the
contained route again. It is deliberately advisory with no veto — two veto-holders
would make the gate unpassable — and its verdict is a routing decision rather than
an objection: this is a point fix, or this needs an RFC, or the plan is fine but
the architecture underneath it is not and that should be on the record.

**The historians can now recognise a shape outside their curated seven.** I had
this one wrong and you corrected it by pointing at the concept register. I'd
described the bug historian's seven examples as hand-typed and frozen. They are
nothing of the kind: they were selected in July by counting which concepts in the
register are most often independently rediscovered across the project's history —
the strongest signal available — and deliberately kept narrow, with the pilot
document saying in as many words that broadening the corpus was future work once
the seat proved useful. So the seven are a curated depth sample with a written
rationale, and what I built is that future work rather than a correction of
anyone's laziness. It is additive: the seven stay exactly where they are.

What was genuinely missing is breadth — a shape outside those seven could not be
recognised at all, however well documented elsewhere. The historians now also
receive a generated index: a hundred and seven failure patterns, a hundred and
twenty-three bug case files, seventy-seven back-catalogue patterns and sixty-seven
of our own logged wrong calls. The full corpus is three and a third megabytes and
cannot be shown to anything, so the seats get titles and are told plainly that
titles are all they have: cite by name, say which file it lives in, and never treat
absence from the list as proof a shape is new. It regenerates from the files. One
honest weakness recorded in the script: it ranks by nothing, whereas the register's
rediscovery-frequency signal is the principled way to choose. That better version
is deferred, not overlooked.

**And the RFC trigger now fires by itself.** The architecture-review track has a
four-condition test for when a change needs an RFC, and nothing fired it, which is
why its own first entry was written after the code was already running in
production. The commit hook that already runs on every commit now checks the staged
changes against that test and prints a note when they meet it — three or more
platform packages at once, an exported symbol removed, a migration shipping
alongside platform code, or a change touching one of the sites we just measured as
repeatedly deflected. It is advisory and never blocks. Backtested against the last
three hundred commits it fires on ten of them: rare enough to mean something, often
enough to matter.

## The near-miss, which is the most useful thing that happened today

Pointing me at the concept register did more than correct a characterisation. Its
pilot document for the bug historian states, in passing, how the council's decision
code actually treats a veto — and reading that sent me to the code, where I found
that the seat I had just written and staged for you to run would have broken the
feature-build lane on every single run.

Two mistakes, stacked. I had assumed a seat is advisory if it is left out of the
council's veto list; in fact any reviewer's veto rejects a round outright, and that
list only changes the audit label. What actually makes a seat advisory is that its
brief never offers a veto in the first place — which is precisely the reasoning
already written down for the bug historian back in July, in a document I had not
read. Worse, I had invented a verdict vocabulary for the new seat — words like
"needs an RFC" — without checking what the decider accepts. It accepts three words
and no others. Anything else is recorded as unreadable, and an unreadable seat
downgrades an otherwise-approved round to "revise". The seat would have forced
every design review to revise, exhaust its three rounds, and fail. Silently, and
only in the lane it was added to help.

It is fixed: the seat now speaks the vocabulary the decider knows, carries its
argument through objection severity, and routes to the RFC track through a separate
field that does not touch the decision. But the general lesson is worth more than
the fix. This is the third time in two days that a confident piece of my own
reasoning was corrected by going and reading something rather than by argument —
first the claim that no pre-build council existed, then the claim the council could
not read its own verdicts, now this. All three were caught by you pointing at
something, or by a measurement I ran for another purpose entirely. None were caught
by thinking harder.

That is also the argument for the whole exercise in miniature. A seat that had our
written history in front of it would have known the veto behaviour, because we had
already written it down.

## Where we are now

The gating question is settled and the answer was yes, so the work went ahead. The
argument that was one-sided now has two sides, with the veto intact on the
conservative one exactly as you ruled.

Three of the four builds are complete. The RFC trigger is committed and live in the
commit hook. The forward seat and the historians' index are written, verified and
staged, but need you to run them, because writing to live fleet configuration is
not something I can do — the same permission boundary as this morning. Both are
config only: no schema change, no image, no rollout, live the moment they commit,
and reversible with one command.

One thing is deliberately still open, and I think rightly. You were undecided about
whether the guardian should weigh benefit at all, and that could not be settled
before now because there was nothing else to supply that half of the judgement.
There is now. My view remains that it should stop weighing benefit — it has no
instrument for it and has been overturned every time it was escalated — but the
honest counter-argument is that risk and benefit are not really separable, and a
seat judging blast radius alone would have to block every wide change, which is
more conservative rather than less. That is a real question and it is now answerable
by evidence rather than argument: we can watch what the new seat actually says.

## Where we're going

The immediate step is yours: run the two staged changes and mirror the historians'
index across to the second council. That takes about a minute.

After that, the honest test of all of this is not that the text is in place but
whether it changes behaviour, and I have recorded that as explicitly unmeasured.
Three things worth reading after the next handful of council rounds. Do the seats
actually query their own minutes, or does the paragraph sit unused? Does the
guardian start citing prior deflections when it invokes the stability preference?
And does the new seat say anything useful, or does it produce confident noise —
because a seat that argues for change without evidence is worse than no seat, and
it should be pulled if that is what it does.

Then the guardian question can be decided on what we observe rather than on which
of us argues it better.

There is one dependency still unaddressed underneath all of this, and the concept
register turns out to be most of the answer to it.

A seat asked whether the architecture is sufficient for our anticipated plans needs
those plans written somewhere it can read. I had said they were scattered across
forty workstream directories with no single home, and that a roadmap document would
not solve it because markdown is invisible to every reviewer. The first half was
too pessimistic. The register already holds the platform's concepts organised by
subject, including — valuably — the ones that were abandoned, which is exactly the
history an architecture seat needs. And the direction ledger alongside it already
names what is fixed rather than negotiable: the constitution and the mission are
blessed documents, hash-checked, with a commit hook that refuses changes to them
without your explicit sign-off.

So the question is narrower and more tractable than I made it sound. It is not
"where do we write the roadmap" — much of it exists. It is "how does a reviewer
query it", since the register is markdown too and therefore invisible for the same
reason. Answering that would serve the architecture seat, both historians and the
reuse and prior-art seats at once, which makes it the highest-leverage piece of
design left on this list. It should be done deliberately, and it should reuse the
register's own rediscovery-frequency signal rather than inventing a new way to rank
what matters.
