# The ratchet from judgement to mechanical — a read-aloud account, 2026-08-03

**Not a series SUMMARY** (those are current-state-only). This is doctrine, prompted by a
direct question after the review-machinery talk: *is there a way to stop these mistakes
being made, not just catch them afterwards?* Written down because the answer turned out to
be a general rule, not a one-off fix.

---

## Start from the measurement, not the intuition

Before answering, I checked whether my five mistakes were personal carelessness or a class
the estate already knows about. `WRONG_CALLS.md` — the fleet-wide, append-only log of
exactly this — has **300 entries**. Two shapes recur through it: a number or claim stated
without running the thing that would produce it, and a check reporting a clean result that
its own conditions made the only possible answer.

That is the finding, before any proposal: **this is not carelessness, and writing it down
has demonstrably not prevented it.** I walked into one of those two shapes while the
warning for it sat in my own auto-loaded memory index. Knowing a pattern does not fire it.
Something at the moment of the mistake has to.

## The marker discipline is necessary and not sufficient

CLAUDE.md already asks for `[MEASURED]` / `[UNMEASURED]` tags on every durable claim, and
the logic is sound: typing the marker is meant to be the check, because most of the time
you'll go and do the query instead of typing `[ASSUMED]` next to your own name.

It didn't save me once. Every wrong figure I wrote this session had a marker, a date, and a
command behind it — because I genuinely believed each one was measured. The marker
distinguishes "I checked" from "I didn't". It says nothing about whether the check I ran
could ever have told me I was wrong.

## The sharper question: could this have come out otherwise?

That's the actual gate. A measurement is only evidence if the conditions of the measurement
permit a different answer.

- I ran a new detector over a tree that already contained my own fix, got zero findings,
  and nearly recorded that as "no false positives". Zero was the *only* number that run
  could produce — the defect was gone before the check looked.
- I published a blast-radius figure from a query that filtered on a status column I'd never
  enumerated. The filter had quietly removed exactly the rows that would have changed the
  conclusion.
- I told the owner a fix wasn't live yet, twice. True the moment I checked; false by the
  time I said it, because someone else had rolled the cluster in between and the claim
  itself has no shelf life once conditions can move under it.

All three are dated, marked, and wrong. The tag that would actually catch them is one I
have to ask myself, not tick: *what result would have proven me wrong, and was it still
possible when I checked?*

## Ask why the wrong thing was easier, before writing the landmine down

The most instructive failure this session wasn't a mistake in a sentence — it was a
mistake that had already been made, correctly documented, and kept happening anyway. The
platform had one tested rule for "has this page shipped". It didn't stop getting re-typed
by hand, wrong, in file after file, because the shared version **couldn't be used** by any
query that aliased its table — which is nearly all of them. The landmine recorded the
symptom faithfully for months. It never removed the reason the symptom kept recurring.

The fix wasn't a sharper warning. It was giving the rule a parameter for the alias, so the
easy path and the correct path became the same path. The same shape appeared in the
adoption authority the owner ruled on the same day: it was safe only because of a sentence
in a comment saying "all callers happen to share this property". Turning that into a field
with the unsafe default off means nobody has to remember the sentence to stay safe.

So: **before writing a landmine, ask why the wrong version was the path of least
resistance.** If there's an answer, fix that, and the landmine becomes unnecessary rather
than merely accurate.

## Some mistakes aren't judgement at all

One error this session had nothing to do with confidence or measurement: I wrote a
placeholder into a commit trailer meaning to fill it in later, except there is no later —
the forward-only rule here forbids amending, so the placeholder is permanent. That's not a
belief that turned out false. It's an ordering mistake, and the fix is mechanical:
require the step that produces the value before the step that records it, and the class of
error stops being available to make.

Not every item on that list of 300 is a judgement failure wearing a confident sentence.
Some are just steps done in the wrong order, and those are the cheapest to close for good.

## What can't be reached from the inside

Underneath all of it is one fact that no amount of care changes: confidence is invisible to
the person having it. Every wrong sentence I wrote today read as true to me at the moment I
wrote it, in the same voice as the sentences that were true. Re-reading it wouldn't have
helped, because I would have re-read the same false sentence with the same confidence.
That is the entire argument for an external check existing at all — not that the person is
careless, but that carefulness cannot see its own blind spot from inside.

What can move is the ratio. Every time a judgement-dependent check repeats — the same
shape catching a different person on a different day — that's the signal to convert it into
something mechanical, so the expensive, slow, genuinely adversarial review is spent on the
claim that's actually novel, not on re-discovering the same five traps. That's the founding
idea behind this platform's own commit-time pattern checks: spend the reviewing council on
judgement, not on what a string comparison can already settle. It is also, now, the plan
for closing the exact three holes this session found — a check for the disconfirmability
question, a mechanical guard on the trailer's shape, and the doctrine line written where
the marker rule already lives, so the next reader gets the sharper version the first time.

---

## Addendum, same day — the plan to mechanise this produced a demonstration of its own thesis

Building the check this essay argues for, I found it doesn't work, and the reason it
doesn't is the essay's own point, arriving in the most direct way possible.

I wrote a rule to flag a count stated with no date, no marker, and no evidence nearby, and
tested it against real history before shipping — which is this file's own standing
practice, and the reason the test mattered rather than the intention. On markdown bug
files, over 53% of ordinary historical commits fired it. On Go comments fleet-wide, 32%.
Reading the false positives, they were not sloppy writing. They were this codebase's
normal, good style: a section states its measurement once, then several paragraphs argue
from it without repeating the tag on every sentence. My own true figures this session were
written exactly that way. The check could not tell my correct claims from my wrong ones,
because at the level of *text* there is no difference — only the epistemic status differs,
and epistemic status doesn't show up as a lexical pattern.

So the rule was not shipped. What ships instead is the question this essay already asked,
now written where the marker discipline itself lives: not "is there a tag", but "could this
have come out otherwise" — which only the author can answer, at the moment of writing. That
the attempt to mechanise it failed on measurement, in the same session that argued
mechanisation only works for genuinely lexical properties, is not an embarrassment. It's the
cleanest confirmation available that the claim was right.
