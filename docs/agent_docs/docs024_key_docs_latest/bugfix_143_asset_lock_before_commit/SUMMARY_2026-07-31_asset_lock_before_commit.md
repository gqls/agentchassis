# Summary — bug 143: the lock that was checked after the damage

2026-07-31. First and closing summary for this workstream.

## What we were trying to do

When someone approves one of a site's images, we lock it, and that lock is meant to
mean one thing: no automated process may ever replace this. Bug 143 said one of our
image jobs was checking that lock too late. The task was to fix it in a way that
holds for the *next* image job somebody writes, not just this one — which is what
the reviewer who first spotted the pattern actually asked for.

## Where we came from

Two days earlier, a different thread fixed exactly this defect in the job that
builds favicons and social-sharing images. During its review, one of the council
seats asked a question worth more than the fix itself: *does any other action share
this shape?* It did. That thread filed 143, wrote in its own notes "that is 143's
job", and moved on. So we inherited a bug that was already diagnosed, already had a
worked precedent, and had an explicit unclaimed handoff attached to it.

## What we've done

The job wrote the new image into the site's repository first and asked about the
lock afterwards, so a locked image lost its file and kept the database row that
said it was approved. Underneath that was a quieter fault the original report did
not mention, and it explains why nobody had noticed in two days: the database write
that the lock blocks fails **silently** — no error, no exception — and the code
discarded the one value that would have revealed it. The job reported success. It
believed it had worked.

Both halves are fixed, and the fix is one shared piece of code rather than a patch:
there is now a single place in the platform that answers "may automation replace
this image?", used by both image jobs, on the read side and the write side. Two
tests stand guard on it — one fails if somebody adds a new image-building job and
forgets to ask, one fails if somebody hand-writes the lock condition again instead
of using the shared one. Both are built so they cannot quietly stop working, and
both were proved to fail before being trusted to pass.

Three things we deliberately did **not** do, each written down with its reason:

- **A third job that deploys images has no lock check, and we left it that way.** It
  looks like the same bug and is not — it publishes the image you asked it to
  publish, so "protecting" it would leave an approved image pointing at a web
  address that expires in a week.
- **We did not make asset locks expire**, though the neighbouring mechanism for page
  sections does. Unifying them would have quietly weakened three existing
  protections. We measured it, found the live data cannot distinguish the two rules
  at all, and concluded that settles nothing — a question like that has to be
  decided on what we are promising, not on what the current rows happen to say. It
  is logged against the project that owns it.
- **We did not fix the review-reminder bug we found by accident**, we fixed it
  properly and separately: the commit reminder was telling off the threads that had
  followed the rules.

## Where we are now

Closed. The fix is live on chassis v1.0.1218 and confirmed by reading the running
binary on both replicas — not the version tag, which proves nothing here, and not
the source. The council approved it at the first round across twelve reviewers with
no serious objection, and the three smaller points they did raise were answered in
code rather than in a reply. Nobody had been bitten yet: none of the thirteen card
images on any site was locked, so this shut the door before anyone walked through
it.

The day's other lesson was about how we work rather than what we built. Three
sessions were editing the same folder at once. One was fixing this same bug in the
same minute and stood down, leaving a contribution that caught a real gap in our
work — we had built the shared write-side guard and then not actually used it
anywhere, which looks identical to having centralised it. Another session tripped
over our own half-finished edit and nearly "fixed" it. Both incidents are written
up from both sides, and the practical rule that came out of it is specific: change
a function and everything that calls it in one pass, then commit, because a
half-applied change blames the wrong file and reads exactly like an old bug.

## Where we're going

Nothing is owed on this bug. Two threads lead out of it, both recorded where the
next person will meet them:

- **LOCK-004**, the long-standing project to give the platform one lock rule instead
  of two, now has exactly one line to change here instead of four. The open question
  — should an approved image's lock expire? — is stated with the measurement that
  shows the data cannot answer it.
- **The "next producer" guard is the reusable part.** The pattern that mattered was
  never really about images: a condition on a database row is not a guard on the
  thing the row describes, and a blocked write that returns success is worse than a
  missing guard. That is written into the debugging guide with the three traps we
  met while fixing it, so the next person meets it before they have a symptom rather
  than after.
