# Where we are — the dead links that the platform spotted and published anyway

Plain prose, append-only, newest at the bottom.

---

**2026-07-26 — what the bug actually was**

Every time we build a page, the platform reads the links in it and checks each one against
the list of pages that really exist on that site. It gets this right. It finds the bad ones,
by name. Then it writes "warning" next to them and publishes the page anyway, because
warnings do not count towards the pass/fail decision.

The comment in the code explaining why that was acceptable said the improvement loop would
come along and fix them later. The improvement loop has been switched off since May. So the
excuse had not been true for months, and nobody noticed because the comment reads as though
someone had thought about it.

**What I measured before changing anything**

The bug report was explicit that we should not just make the check stricter without counting
the damage first, because "a page with one bad link" turning into "no page at all" is worse.
That was the right instinct and the numbers backed it up.

Over the last thirteen days, sixteen page builds went through the check. Three of them had
bad links — nineteen per cent. Seventeen bad links between them, fifteen distinct targets,
and every single one of those fifteen was a page that has never existed in any form; not one
was a typo we could fix. All three pages published.

The two pages affected were the **home pages** of oufe.com and webdesign.co.uk. So if I had
simply made bad links fatal, as the obvious fix, two home pages would have failed to publish
at all. That is worse than what we have now, and it is exactly what the bug report warned
about.

**What I did instead**

The page still publishes, but the bad link does not. If the target is real and the writer
just wrote the address slightly wrong — `/contact` where the page is actually
`/contact.html` — we correct the address. Otherwise we take the link off and leave the words
alone, so a sentence like "read our pricing guide for details" keeps its full meaning and
simply stops being clickable. Nothing gets deleted.

I also made it write down what it did, permanently. Before, the only record was a line in a
log that gets thrown away, so a day after a build nobody could tell you which links had been
wrong. Deliberately *not* a job on a to-do queue: I checked, and that queue has never once
been drained — those items have been created twenty-two times and actioned zero times — so
adding to it would have looked like progress and produced none.

**The thing I found that is arguably bigger than the bug**

We do tell the writer which pages exist. There is a step whose entire job is to hand it the
list and say "only link to these". I checked whether it was working. It is not: on the last
twenty page-writing runs it found zero pages every time, so the instruction is dropped from
the prompt entirely and the model is left to guess. That explains why the bad links are all
invented rather than mistyped.

That is a different fault in a different place, so I have written it up on its own rather
than quietly bundling it in. It is the more valuable of the two to fix next, but it changes
how the writer behaves everywhere, and that deserves its own measurement rather than riding
along with this one.

**Where this stands**

The code is written and tested. It is not live yet — this kind of change only takes effect
when a new image is built and rolled out, and I will confirm it against the running system
rather than against the code before I call it done.

---

**2026-07-26, later — a mistake worth recording**

I nearly shipped a test that could not fail.

The habit is to deliberately break your own code and check the tests notice. I did that, and
it reported four failures out of eight. I assumed four of my tests were weak. They were not:
they had never run at all. One test asked for "the first repair" without checking there was
one, so when there were none it crashed the whole test program, and every test written after
it was skipped in silence. The output of a run that did a third of the work looks exactly
like the output of a run that did all of it.

Fixed, and re-checked: all eight now fail when the code is broken, and the four tests whose
job is to confirm nothing changes still pass. Worth knowing generally — a crash in one test
quietly disables the rest of the file.

---

**2026-07-28, evening — the fix is written, and it turns out the problem was bigger than we thought**

Yesterday's fix was real but landed in the wrong place. The build gate finds the broken links
and repairs them, and then the very next step throws the repaired version away and saves the
original. So the platform knew, wrote down what it knew, and published the 404s anyway.

Today I moved the repair to the last possible moment — the point where sections are actually
written to the database. Nothing can route around that. Whatever a workflow's configuration
says, a section with a dead link cannot be saved any more.

Before submitting it for review I checked how many parts of the system this touches, because
the rule now is that you measure that yourself rather than asking the reviewer to. Six
different build routines save page sections. **Only two of them have a link check at all.**
That was the surprise: this was never just "the gate's repair gets discarded on one route" —
four of the six routes that write page content have never had any link repair, by any route,
ever. It also settles which of the two candidate fixes was right, and not by argument: fixing
it inside the gate could never have reached the four routes that have no gate.

A small thing worth admitting, because it is the kind of mistake that reads as competence.
My first version of that count returned three routes, not six, and it looked perfectly
credible — it just happened to ask about a place where three of the six do not keep their
configuration. I had a number that disagreed with the handoff and I was one keystroke from
submitting mine. Re-running it without the filter is what caught it. A query that quietly
describes a smaller world than you asked about is very hard to spot from its output alone.

Where it stands: written, tested, committed, and submitted to the reviewer council. It is
**not live**. This kind of change does nothing until a new image is built and rolled out, and
I will not call it fixed until I have watched a real page go through it and come out clean.
The bug file stays open until then, deliberately — the 404s are still reproducible today.

---

**2026-07-29 — it works, and the test I designed to prove it proved nothing**

The fix is live and the bug is closed. But the way it got proved is worth writing down, because
I nearly recorded a success I had not earned.

I had designed a careful test: take a page whose stored content still had two dead buttons on
it, push it through the rebuild, and watch the dead links get cleaned. I checked the
preconditions properly beforehand — several of them, any one of which would have invalidated
the test. It ran, it finished, and the dead links were gone. That is exactly the result I
predicted.

It was the wrong result for the right-looking reason. The clean-up is supposed to remove the
*link* and keep the *words* — a button that goes nowhere becomes plain text. But the words had
vanished too. That mismatch is what made me look again, and the explanation was that the
rebuild regenerates each section from stored data through the current template, and the
template simply declines to draw a button when it has no destination for it. The buttons were
never handed to my code at all. My test had destroyed its own subject.

What actually proved the fix was a run I had nothing to do with. Overnight, an ordinary rebuild
of a different site hit the same code and cleaned five genuinely broken links — I checked, all
five really do 404 — and this time the evidence was unambiguous: the record of the repair, the
saved page missing exactly those five links, and the live page missing them too. It is the
precise mirror image of the original bug, where the repair was recorded and then thrown away
400 milliseconds later.

The other thing that came out of this: the reviewer council caught a real hole in my reasoning.
I had measured how many build routines this touches by asking which ones are *configured* to
save page sections. The right question was which code actually *writes* the page content, and
that answer is ten places, not one. Three of them still have no protection — including,
awkwardly, the one our own documentation tells people to use for small edits. So following the
documented practice is the most reliable way to bypass this fix. That is now its own filed bug.
I have not measured how often it bites, and I have said so rather than guess.
