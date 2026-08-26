# SUMMARY 2026-08-26 — the buttons that describe one page and link to another

## What we are trying to do

Stop the framework shipping buttons whose words promise one thing and whose link goes somewhere
else. The owner reported four of them on dartsonline.com and asked for the general fix rather than
four individual repairs. This is that general fix — or as much of it as the evidence supports.

## Where we have come from

The framework has always known the answer. When it builds a page it decides where each button should
go and it writes down two things side by side: the link, and **the title of the page that link leads
to**. It writes the title down on purpose — the comment in the code says it exists "so the content
writer can write CTA copy for the actual destination instead of guessing one".

And then nothing ever compared them. The words and the destination sit in adjacent slots of the same
record, on 665 components across the estate, and no piece of code had ever read them together.

Instead the system found these the hard way: render the page, deploy it, sweep it days later, parse
the HTML back out, and try to reconstruct from the finished article a fact that had been free at the
moment of writing. That sweep runs on one of five discovery agents, and it does not converge — 182
of its findings have been filed more than once, because pages that re-author themselves several
times a day re-mint the defect faster than a periodic sweep can clear it.

There is an older bug here too, and finding it changed how we describe this one. `bugs_closed/023`
is titled "A button's label and its destination are never checked against each other" — almost
exactly this bug's claim — and it closed on 25 July **without ever building the comparison**. It
solved a different problem: it made *missing* destinations impossible to render. A *wrong*
destination was never in its scope. So this is that question reopened, not a duplicate, and we have
said so in the file so nobody closes it as one.

## What we have done

First, the measurement that decided everything else. The obvious theory was that the writer had
never been shown the destination — in which case the fix is plumbing, not policing. So we checked.
The writer is shown it **twice**: once in the button field's own instructions, and once at the moment
of writing, in a sentence reading *"the destination is fixed — write this text for that destination;
never promise a different one."* Three-quarters of a thousand prompts carried that sentence in three
days.

And it is still wrong about **one time in seven**. Of the buttons written since that instruction
started reaching the writer, 155 out of 1,060 describe somewhere they do not go. That result could
have come back near zero, and it did not — which is what makes it evidence rather than reassurance.
Telling the writer more firmly is not going to fix this.

So we built the check. The important decision was to make it ask a question the system *already
asks* rather than invent a new one: the sweep's own test — "which page does this wording name, and
is that where the link goes?" — lifted into one shared function, with the sweep rewritten to call it.
The proof that lifting it changed nothing is that the sweep's existing tests pass untouched. Then we
ask that same question once more, before the page is saved.

Two things we deliberately did **not** build, both of which the bug file asked for.

We did not rewrite the wording to match the link. That sounds like the obvious repair and it makes
things worse: a neighbouring investigation found the system often picks the *destination* badly, and
once wording has been written to describe a bad pick, the cheap fix for the pick can no longer reach
it. We would have been converting easy problems into hard ones at about 150 a week.

And we did not move the link to match the wording, because it almost never can. Of the 186
disagreements live today, the wording clearly names exactly one other page in **13** of them. In 78
it names two pages equally well, and in 95 it names nothing that exists. An automatic correction
would reach 7% while risking the rest — and there is a dated case from 24 August of an automatic
button repair turning a *correct* contact link into a wrong one.

That left recording it, which sounds like a retreat and is not: 173 of those 186 buttons need a human
or a writing pass, and nothing anywhere told anyone they existed. Two other pieces of work are
stalled for exactly that reason, and both have now been handed the list.

## Where we are now

Built, tested, committed and submitted for review. Six deliberate sabotage tests were run against the
new code and all six were caught by a failing test, so the guard is real rather than decorative. It
is inert until the next fleet image rolls and a small configuration change applies — deliberately, so
it can be watched rather than discovered.

Three near-misses are worth knowing about, because each was caught by a different thing.

We nearly put the check where the words are first written. That turns out to miss the *repair* path
entirely — the half of the system that actually churns. Both halves meet one step later, at the
point the page is saved, and that is where it went.

We nearly broke a working check while tidying. Our design had the shared logic reject anything that
is not an ordinary page link — sensible, since a phone number is not a page. But there is a live
check for phone and email buttons that feeds that exact logic phone numbers on purpose. Tidying it
would have switched that check off silently, with no error and no failing test. Caught by reading the
callers before writing the code.

And we nearly armed the new recording on two places when there are six. Four of them are buried
inside loops where the obvious lookup cannot see them. For a *measurement* that is not a partial
rollout — it is a number that reads as fleet-wide and is quietly wrong.

## Where we are going

Three things, in order.

Read the review when it lands and act on it, including if it disagrees with the call not to build a
repair arm — that is the decision furthest from what the bug originally asked for and the one most
worth a challenge.

After the next roll, confirm the recording actually fires from **both** halves of the system. Seeing
it from only one would mean the coverage is failing quietly, which is the precise failure this design
exists to avoid. Then re-measure the rate at the new checkpoint; it will differ from the one-in-seven
figure because it asks a sharper question, and the new number is what any later decision should rest
on.

And the honest weak point: **a record nobody reads is not a fix.** What matters is the rate, not the
individual rows — nobody should read 155 records, but somebody should notice if one-in-seven becomes
one-in-three, or drops to one-in-fifty after a change to how the writer is instructed. There is a
query and a standing monthly obligation written down in two places. That is a promise on paper, and
paper promises are exactly the failure another session filed a bug about on the same day. If that
work produces a general way to surface this kind of finding, this record is one query away from using
it.
