# SUMMARY — 2026-08-12 — a spelling mistake between two halves of our own tooling

*Written to be read aloud. Current state only — the chronology is in `README_where_we_are.md`, the
technical log in `NOTES_…`.*

---

## What we're trying to do

Two of our sites were shipping pages with no hero image and no logo, and saying nothing about it —
the code that looked for the picture found nothing, shrugged, and carried on. The job was first to
make that failure *audible*, so we would know when it happened, and then to find out why the picture
goes missing in the first place.

## Where we've come from

The first half is done and live. Three places in the code now record a durable, queryable row
whenever a deployed image arrives without a usable address, instead of failing in silence. That went
through review, was approved, and has been running in production since yesterday evening. It has
recorded nothing yet — but we can say precisely why, which is the point: nothing has tried to deploy
a hero or a logo since it shipped, so there has been no opportunity. We know that because we
measured the demand as well as the failures.

The second half — *why* the address goes missing — has been stuck. Twice we pointed our automated
diagnosis system at it, and twice it came back saying it could not reach a conclusion. That is
supposed to be a rare answer.

## What we've done

We found out why it kept failing, and it had nothing to do with the original bug.

When our diagnosis system investigates a question, it assembles a pack of evidence: the relevant
functions, the database tables, the live records. It could not read the functions. Not some of them
— **any function that belongs to a type, which is about a quarter of our codebase.**

The cause is a mismatch between two halves of our own tooling. When we catalogue the code, a
function belonging to a type is written down in the standard style, with the type in brackets in
front: `(*SagaCoordinator).applyResponseToState`. That is the name in the catalogue, and it is the
name the evidence pack displays. But the piece of code that goes and fetches the text of a function
expected the name written *without* the brackets. It looked, did not recognise the form, and
reported "symbol not found" — which reads exactly like "that function does not exist". Our diagnosis
system is deliberately built to abstain rather than guess when it cannot cite evidence. So it
abstained, correctly, on evidence that was sitting in the catalogue the whole time.

Counting every diagnosis we have ever run: **335 function bodies lost this way, across 47 separate
investigations.** There was a smaller second version of the same fault — package-level values, some
1,238 of them, which the catalogue has recorded since July and the fetcher was never taught to look
for.

The uncomfortable detail is that our own written guidance caused it. We keep a file of traps for the
team, and yesterday someone added an entry saying: when you ask about one of these functions, use
the bracketed name, or you will be told it does not exist. That is correct for *searching the
catalogue* and exactly wrong for *fetching the text*, because those are two different pieces of code
that never shared a rulebook. Following our own instructions was what walked into the fault.

It is fixed, tested, committed and approved — the review took six minutes and twelve reviewers, and
one of them earned its place: it pointed out we had only checked two of the three places that
produce these names. Checking the third made the problem *worse* than we had described. One step in
the system, whose whole job is to tidy up a vague request into a precise one, was taking a name that
would have worked and rewriting it into the form that could not be read — and logging that it had
succeeded.

## Where we are now

The fix is written and approved but **not yet live**. It is Go code, so it does nothing until the
next time we build and roll the images. The fault is still costing us right now: fourteen more
function bodies were lost in the forty minutes it took to write this up.

We also have two of our own mistakes on the record, deliberately. One is that a claim went into the
review submission before it was checked — it said all twenty odd cases were of a certain type, and
when the check was finally run, nineteen were. The honest figure is 334 of 335, not all of them, and
the fix does not cover the twentieth. The galling part is that the sentence existed *precisely* to
be the one that could prove us wrong, and it was published before it was allowed to. The other is
smaller: a bug number was checked as free, and half an hour later another session had taken it — so
the fix's commit message points at a number that now belongs to someone else's problem. Ours is 261.

## Where we're going

Nothing here is blocked on us, and there is a choice to make.

The immediate step, once the images roll, is to point the diagnosis system back at the original
hero-and-logo question. It has failed twice for want of exactly the function bodies this fix
restores, and it is the last unanswered part of that bug. We will know the fix worked by seeing a
function body appear in the evidence pack — not by watching the error count fall, which would also
fall if nobody asked.

Beyond that, the outstanding decision is unchanged and better informed than yesterday: which of the
remaining pieces of the commission to take next. One is large and mostly investigation, with a
design decision reserved to you. The other is medium-sized, spans three layers, and needs a call on
whether it goes through the normal review or gets written up as an architecture proposal first.
