# SUMMARY — 2026-07-30 — bug 133: the scrape reply that lied, and the one that vanished

Written to be read aloud. Current state only; the chronology is in
`README_where_we_are.md` and the evidence in `NOTES`.

---

## What we were trying to do

Make a scrape reply tell the truth about itself — about what it contains, and
about whether it arrived at all.

## Where we had come from

Another thread found two faults on 28 July while doing something else, wrote them
up carefully, and said plainly that they were not its bugs to fix. So they sat for
two days.

The first: when a scraped page is too big for our message bus, the adapter cuts it
and adds a note saying the full version is in S3. That note went out every time,
while the upload that would make it true is optional and mostly off. We were
destroying the end of a page and leaving a signpost to nothing.

The second: if the bus refused a reply for being too large, the adapter logged a
line and gave up. The caller is not reading that log; it waits out its retry
budget and fails for the wrong reason. We had already learned this lesson once,
written it down, and applied it to the batch version of the same code.

## What we have done

Both are fixed, live on the running adapter, verified on every replica, and
council-approved in nine minutes.

The note is now built *from* the storage address, so a stored copy can only be
claimed by naming one, and it resolves the address per field because the uploads
happen one field at a time and any one can fail quietly. There is no longer a way
to phrase the claim without the evidence in hand.

For the second fault we did not copy the working code across. We measured first,
and the measurement changed the job: exactly **one** place in the whole system knew
that "too big" is a permanent failure, while **nine** places send replies — eight
of them giving up quietly. Copying would have made it two of nine and left two
hand-written copies of one rule, which is the drift we closed a bug about last
week. So the policy now lives in one place with both scrape paths calling it, and
the batch version's private copy is gone.

We proved it by firing one real scrape of one of our own sites, reproducing the
exact case from the bug report, and reading what came back off the queue. The
reply says the text was discarded, says it in a machine-readable field as well as
in prose, and the old sentence appears nowhere.

## Where we are now

Closed. But the more useful outcome is three corrections to what we thought we
knew, and none of them came from the code being hard.

**The bug's own count of affected steps was too small — nine, not four.** Its
query named three actions and there are six. The missing one is the busiest
scraper we run, firing on live traffic continuously; another four are the
multi-page crawls, where a stored copy mostly cannot exist at all. We found this
only because we went to read a real message off the queue before sending a test
one. Re-running the query had confirmed the wrong number perfectly, because
re-running SQL reproduces its filter's blind spot exactly. **"Re-run it rather
than trusting the table" has to mean re-deriving the question, not re-executing
it** — and that is now written down in three places, because it looks exactly like
diligence.

**Twice we were a minute from writing a confident false finding into the bug file**,
both times from reading a long function through a window that stopped short of the
part that mattered. Caught both, ten minutes apart, with the lesson from the first
already written down.

**A test we wrote caught our own comment.** It scans for the old lying sentence to
stop anyone restoring it, and it fired on the line where we quoted that sentence
to explain what had been wrong. A substring scan cannot tell a real message from a
note about one — which is the same confusion that made another thread's evidence
unsound last week, when four of its five checks looked for sentences that only ever
existed in documentation. The test now parses the code and looks only at text that
can actually be sent.

The review was worth its cost twice over. It approved the change, and two seats
disagreed with each other about the one judgement we could not settle by measuring
— whether putting a shared mechanism into core messaging plumbing needs the heavier
architecture review. The resolution carries a caveat we have recorded word for
word: **the moment anyone adopts this in the other services, that is an
architecture review, and today's approval does not cover it.** Two other reviewers
asked questions we answered by measuring rather than arguing, and one of those
answers was more interesting than the question: the shared test double they
suggested we extend has no users at all and does not implement the interface it
claims to, while two other packages have quietly written their own.

## Where we are going

`bugs_open/158` holds everything we deliberately left out, in fix order and with
the measurement each one needs first. The main item — adopting the shared policy at
the remaining eight sites — is an RFC, not a bug patch, and should be sized before
it is written: it may turn out to be seven latent gaps and two live ones. Two
smaller items need an owner's decision rather than an engineer's, and one is a
two-minute tidy-up that removes a trap.

One question still needs you: whether these replies should be capped at 1MB or
5MB. We have both numbers in the system today and no stated intent, and guessing
would only add a third.
