# SUMMARY — 2026-08-24b — closed: the fix is live, and the proof that couldn't be had

*Second summary of the day, and a real inflection: the morning's read-out ended "one half is
committed but does nothing until the image service is rebuilt". It has been rebuilt. That changes
the state, the close decision, and one thing we believed about the bug itself.*

## What we're trying to do

Stop the platform quietly using its weaker image model. Everything we generate should go to
Gemini, which renders legible text and can be anchored to a site's own brand images. The older
model, SDXL, can do neither, and a site that genuinely wants it can say so in its own settings.
None ever has.

## Where we've come from

July's fix made the model chooser loud when it meets an image type it doesn't recognise, but left
one hole open on purpose: when the type is *missing entirely*, it stayed silent, on the assumption
that a missing type meant an old caller who had chosen the old model deliberately. Nobody checked
that assumption. Fifteen hero images went to SDXL after that fix, on five sites, none of which had
asked for it — found only because the owner looked at a face.

This morning we found the cause: the step that makes per-page heroes sent no type at all, and had a
setting beside it that looked exactly like the safety net and was connected to nothing. An earlier
fix had found that same class two weeks ago, fixed two of the three places, and written in its own
notes that everything else was fine. One thing was not.

## What we've done

Both halves shipped and both are now live.

The configuration half went live this afternoon and gave that step the missing type — and the
site's identity, which it had also never been sent, meaning per-page heroes were being made with no
knowledge of the site's palette while the main hero on the same site got all of it.

The code half rode the rebuild. We checked it three ways rather than trusting a version number:
both copies of the image service run the same image, so this isn't the case where a rebuild quietly
serves the old one; the service itself states which commit built it, and ours is in it; and the new
code's own text is inside the running program, checked against one control that had to be present
and one that had to be absent, so the check could have failed.

Along the way we corrected our own debugging guide, which was actively instructing future sessions
to keep the silence, and wrote up two traps: a configuration key on this kind of step reaches
nothing, and a work-item table we routinely query is a rolling window that answers historical
questions about the last day only.

## Where we are now

Closed. The defect as filed cannot happen any more: there is no path on which a request with no
image type is silently served by the weak model.

Two things about *how* it closed are worth more than the close itself.

First, a condition we set this morning could not be met, and finding out why corrected a belief. We
wanted to watch a real per-page hero get made after the fix. There will not be one for a while:
across the whole fleet there is currently not a single per-page hero that has been asked for and
not made. The queue is empty. That also re-reads the evidence — "no new bad images since 11 August"
was not the problem going quiet, it was that producer finishing its list. Same numbers, different
meaning, and only the second meaning tells you the fix cannot be proved by waiting. Rather than
manufacture an unwanted image on a customer site to satisfy our own checklist, we retracted the
condition in writing, said why, and left three queries that anyone can run — the third being the
one that asks whether anything generated at all.

Second, the review found things we hadn't. One reviewer refused to accept a promise that a
companion fix "ships alongside", on the grounds that this exact file has a history of exactly that
promise being false — which is why the bug existed. Another approved and then objected past its own
approval to say this is the third bug in six weeks on one underlying seam, none found by any check,
all three found by a person looking at an image. And a second lane, sent a courtesy heads-up,
checked its own site instead of taking our word and found that the naming pattern we had used to
attribute nine of the fourteen affected images proves nothing about which code path made them.
Re-run properly the attribution held — which is the uncomfortable outcome, not the reassuring one.

## Where we're going

Nothing on this lane. Three loose ends are named in the closed bug file, each handed somewhere a
person will meet it: two workflows still carry image steps naming no type, which cannot be fixed by
configuration and are now harmless and loud, left with the imagery lane; a rarer producer with the
same gap where we deliberately did not apply the obvious fix, because it would have made six of
eleven cases worse, with the reasoning written down so nobody undoes it; and the seam underneath,
written up as a proposal for a human decision rather than changed on our own judgement.
