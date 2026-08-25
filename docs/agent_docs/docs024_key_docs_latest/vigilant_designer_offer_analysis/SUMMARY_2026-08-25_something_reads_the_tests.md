# Where we are — something reads the tests now (2026-08-25)

*Tenth in the series. The ninth, `SUMMARY_2026-08-24_the_tests_become_checkable.md`, ended with a
mechanism that wrote machine-checkable tests and nothing that read them. That is what changed.*

---

## What we're trying to do

We build websites automatically, and we have agents that look at those sites and say what is wrong
with them — this page's description leads with a number when it should lead with the promise, that
page claims something the site cannot support. Each complaint is written down as a piece of work,
handed to another agent to fix, and marked done when that agent reports back.

The thing we are trying to fix is the word **done**. It has never meant what a reader assumes. It
means "the agent we sent came back and said it finished" — not "the thing we asked for happened".
For most work those are the same. When they are not, nobody finds out, because the only record is a
column that says `complete`.

## Where we've come from

For weeks we could only catch this by hand: read the complaint, read the page, and judge. Slow, and
it does not scale past the person doing it.

Yesterday we changed the producing side. When one of our analysers writes a complaint, it can now
also write a small **test** — a plain, mechanical condition over one field of one page, of the kind a
machine can check without judgement. We keep such a test only if it **fails right now**, at the moment
it is written, because a test that already passes tells us nothing about the problem being reported.

That went live and the analyser used it, unprompted, on its first run — writing three tests across
webdesign.co.uk. And one of them immediately caught the thing we had been describing in prose for
weeks: a page had been rebuilt, deployed, and its item marked done, while the test that item had set
itself was still failing. For the first time it was a machine saying so, not a person.

We wrote it up as a bug and stopped, because we had built the half that writes tests and nothing
that reads them.

## What we've done

Built and reviewed the reading half.

There is now a check that runs in the moment between an agent reporting success and the system
recording the work as done. If the item set itself one of these tests, the check re-runs it against
the page as it stands at that moment. It is switched on for one kind of work item, off everywhere
else, and everything it touches is unchanged for anything that has not opted in.

**It records its verdict. It does not refuse anything yet.** That is deliberate and it is the most
important thing in this summary, so it has its own section below.

It went through the review council: fourteen reviewers, approved on the first round, five advisory
notes and nothing serious. One reviewer asked a question I had not thought to ask — whether there is
a second route by which work gets marked done that would go round the new check. I went and measured:
the obvious candidate turns out to use the same route, and 1,600 of the 1,638 items of this kind in
our entire history came through it, including the one this bug is about. Around 2% did not, and that
is written down as something to settle before the check is allowed to block anything.

Two reviewers separately observed that this is the fourth such check bolted onto the same piece of
machinery. Neither objected to it shipping; one asked that the pattern be named rather than quietly
absorbed. It is now a written proposal for the owner.

## Where we are now

**The loop is closed on paper and not yet in the world.**

The code is committed and will start running at the next fleet update. When it does, it will watch
and write down what it sees.

The reason it only watches is worth stating plainly, because "why not just block the bad ones" is the
obvious question. **Every one of these tests we have ever had is currently failing.** All three. We
have never once seen this check look at a real fixed page and say so. A check that has only ever
returned "no" cannot be distinguished from a check that always returns "no" — and giving that
authority to block work would be trusting something we have no evidence about.

So it watches, and it records a verdict on **every** item it examines, including the ones it passes.
That last detail is the instrument: if it only recorded failures, a check that had silently stopped
working would look exactly like a check that was passing everything.

And it will not be allowed to watch for ever by default. Switching it from watching to blocking now
**breaks the build** until the person doing it has also closed a separate, older side door — a
timeout mechanism that can mark work done while going round every check we have. That is enforced by
the compiler, not by a note someone has to read.

Two things went wrong in the building, both recorded in full. I nearly built the check for a reason
borrowed from a different check answering a different question — the conclusion was right and the
justification was not, and I had already written that wrong justification into our main debugging
guide the day before, so it is struck out there now. And I found, by reading the live records rather
than by anything failing, that the tests as they are **stored** cannot be handed straight back to the
thing that checks them: the system stamps a small provenance record onto each one, and the checker
does not recognise it. Left alone, the whole check would have run, reported no problems, and been
completely blind — with an error message pointing at the wrong culprit. No test we had would have
caught it, because the one test that uses real data types the tests out by hand, without the stamp.

## Where we're going

One thing, and everything else waits behind it: **we need to see this check pass a real page.**

After the next fleet update, re-run the analyser and watch for the first item that completes with the
check saying "this criterion is met". That single row is what turns a mechanism we believe in into
one we have evidence about, and it is the precondition for letting it refuse anything.

After that, in order: measure how widespread the problem actually is, now that it can be counted
rather than read; work out why the fixing agent produces content that fails the test it was given,
which this check detects and does not address; and answer the owner-facing proposal about the four
hand-wired checks.

There is also unfinished work from before this: the rest of the analyser's second-version batch,
three config-only changes that have been waiting for a reason to open the file, and an approved but
undesigned feature for auditing the claims a site makes about itself.
