# Summary — the poisoned page sources (bugs_open/190), 2026-08-04

First summary in this lane's series. Written to be read aloud.

---

## What we are trying to do

Stop the platform from storing, and then endlessly re-storing, the wrong kind of thing in the
place a page is rebuilt from.

Every page is held twice: the HTML a visitor sees, and the structured data underneath it —
headline, body, features, each in its own named field. The second is the one that matters,
because every rebuild regenerates the page *from* it. Two live pages have the delivery
envelope of an AI reply sitting there instead of those named fields: a wrapper saying "this is
text", with the entire real reply crammed into one long string. The filing cabinet has been
replaced by a sealed envelope pushed through the letterbox.

Both pages look perfectly fine today. That is the difficulty, not a mitigation: the damage is
scheduled rather than visible, and it fires the next time either page is rebuilt.

## Where we have come from

The *cause* was fixed in July. When an AI reply could not be read as structured data, the
platform used to wrap it in that envelope and carry on; three rounds of repair work during
July taught it to recover properly and to fail loudly when it cannot. No new envelope has been
created since the middle of July, and we can show that.

What was never fixed is the other half. Nothing ever checked the *shape* of what was being
saved. So the handful of envelopes that got through in the spring became permanent: every
rebuild reads one, writes it straight back, and reports success. The bug was filed on 3 August
by another lane as leftover residue — two rows, cause already fixed, tidy-up required.

## What we have done

Re-checked it, and found the filed account wrong in two places. One of the two rows it names
no longer exists: that page was rebuilt an hour *after* the report was written, and the
rebuild produced a fresh row with the same rubbish in it. And the report's own instruction for
how to detect these rows — look for "exactly two fields" — would have missed the other one,
which has three. Following the case file faithfully would have shipped a guard blind to half
the known population, and every test would still have passed.

We also nearly made a worse error ourselves and did not: a history table holds 65 records of
these rows being overwritten, which reads exactly like "the system has written 65 of them,
most recently yesterday". It has not — that table stores what was *replaced*. Only reading the
code settled it, and no further database query could have. Doing it properly showed the
problem has touched **25 pages across 6 sites** over its life, far more than the two visible
today.

Then we built the fix: a guard at the point where page content is written, which opens the
envelope when that can be done without losing anything, and **refuses the save outright** when
it cannot. It is wired into both places the platform writes page content, and it is registered
so the next person can find it. Fourteen tests, and — because this guard's normal behaviour is
to do nothing at all, which is also what a completely broken one does — we sabotaged our own
code four times to confirm the tests could actually detect a broken guard. All four failed the
right test and recovered.

## Where we are now

The code is committed and correct, and the review council is partway through assessing it.

**It is not live.** It is Go code, so nothing changes until the next chassis image is built and
rolled, and that is a whole-fleet operation the owner runs. Until then both pages are still
reproducible, so the ticket stays open — that is this project's standing rule, and it is the
right one here.

One decision deserves to be stated out loud rather than buried. **One of our pages will now
start failing its rebuilds loudly, and permanently, until a person fixes it by hand.** The
gaswholesalers pricing page *can* be opened mechanically, but what comes out is a
131-character fragment while the real copy sits outside the wrapper in a form no machine can
attribute to the right fields. We chose the page that fails noisily over the page that quietly
guts itself. If that noise becomes irritating, the fix is the page, not the guard — loosening
the guard by one line is precisely what would destroy the content.

Worth recording about how the work went, because it is a fair picture of this shared
workspace: two other sessions were editing the same file during the session, the build broke
under us once for reasons that were nothing to do with us, and in the end another session's
broad commit swept up our final change and our landmine notes into commits about entirely
different bugs. Nothing was lost, and the project's rules anticipate exactly this — but it
does mean part of this fix now lives under a commit message about a different bug, which we
have written down where someone will find it.

## Where we are going

Four things, in order.

1. **The next chassis roll**, whenever it happens for other reasons — then confirm the guard is
   genuinely in the running binary rather than trusting the tag.
2. **Repair the finetuning page through the framework**, not by hand. Its envelope contains a
   good article, so an ordinary rebuild should now come out clean — the machinery repairing its
   own page, which is the outcome worth having.
3. **Hand the gaswholesalers page to a person.** It already has a ticket. No automation should
   touch it.
4. **Read the council's verdict** and act on it if it asks for changes.

After that the count goes from two to one, and to zero only when the human finishes. **A count
of one is the expected result, not a failure** — which is exactly the kind of thing worth
saying in advance, before someone reads it as the fix not having worked.
