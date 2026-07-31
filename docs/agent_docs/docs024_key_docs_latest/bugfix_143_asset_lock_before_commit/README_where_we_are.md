# Where we are — bug 143, the asset lock that was checked too late

Plain prose, append-only, newest at the bottom.

---

**2026-07-31**

I picked up bug 143 off the open pile. Short version of what it was: when you
approve one of a site's images, we "lock" it, which is meant to mean nothing
automated may ever replace it. One of our jobs builds a small card image for an
article by cropping the article's main photo. That job wrote the new file into the
site's repository first, and only afterwards asked whether the image was locked.

So if you had approved a card image, the job would quietly throw your approved
picture away, put its own crop in its place, and then discover the lock — at which
point the only thing the lock could still protect was the database row describing
the file that no longer existed. You would be left with a record saying "approved"
pointing at something you never approved.

There was a second, quieter part that the original bug report did not mention, and
it is the reason nobody had noticed. The database write that got blocked by the
lock fails *silently* — no error, no complaint, nothing. The code threw away the
one value that would have told it the write had been refused. So the job reported
success. It genuinely believed it had worked.

Nobody has been bitten by this yet: I checked, and none of the thirteen card
images on any site is currently locked. So this was closing a door before anyone
walked through it, not cleaning up after a disaster.

**What I did.** The same job's sibling — the one that makes favicons and social
sharing images — had exactly this bug and had it fixed properly a couple of days
ago, on a different thread. So the fix here was not to invent something, it was to
take that thread's answer, lift it out into one shared piece of code, and use it in
both places. There is now a single place in the system that answers "is automation
allowed to replace this image?", the card job asks it *before* doing any work at
all, and the silently-refused database write is now noticed and reported instead
of swallowed.

I also added a test that will fail if somebody adds a *new* image-building job and
forgets to ask. That was the actual request from the reviewer who spotted this in
the first place — they were less worried about this one job than about the next
one somebody writes.

**One thing I deliberately did not do**, because it looked like the same bug and
is not. There is a third job that deploys images to a site and has no lock check
at all. I nearly wrote it down as another instance. But that job publishes the
image you asked it to publish — if that image is locked and approved, publishing
it is exactly right, and "protecting" it there would leave the approved image
pointing at a web address that expires after a week. So I left it alone and wrote
down why, in the test itself, so the next person doesn't have to re-derive it.

**Something worth your knowing about how the day went.** Three separate sessions
were working in the same folder at once, and we got in each other's way twice.
Another session was fixing the *same* bug at the same moment — I only found out
when the compiler complained about a duplicate. Their version was better than mine
in several respects, so rather than argue, I kept the good parts of theirs and said
so in the file. Their draft has since disappeared from the working folder, so
nothing was lost.

The other collision was my fault and worth being honest about. Halfway through my
change, the code in the shared folder was temporarily broken — not committed
anywhere, just sitting there mid-edit. A third session hit that breakage, reasonably
concluded it had been broken for two days by somebody's old commit, and started
"fixing" my unfinished work. They caught themselves before committing anything, so
no harm done, but the cause was me leaving a half-finished change lying around
while I worked through it. Both of us have written it up. The lesson on my side is
specific and I have recorded it: when you change the shape of a function, change
everything that calls it in one go and commit immediately, because a half-done
change of that kind blames the wrong file and looks exactly like an old bug.

**A small thing I fixed on the way past.** Our commit process prints a reminder if
you commit platform code without a review stamp. It was printing that reminder at
people who had done the right thing — there are two stamps, one for "reviewed" and
one for "submitted, verdict still pending", and the reminder only knew about the
first. So the thread that followed the rules got told off, which is the fastest way
to teach everybody to ignore a reminder. It now recognises both, and tells you the
one thing you do still owe: go and read the verdict when it lands.

**Where it stands.** The fix is committed and the wider test suite passes, both in
my working copy and against the committed code on its own. It has gone to the
review council and I am waiting on the verdict. It will not actually be running in
production until somebody rebuilds and rolls the chassis image — Go changes sit
inert until then — so I am holding the bug open until I can prove it from the
running pod rather than from the code. I am deliberately not doing that rebuild
while my own review is still in flight, because a deployment kills an in-flight
review and we would pay for it twice.

**2026-07-31, later — done and live.**

The fix is running in production. I checked it the only way that actually proves
anything: by reading the compiled binary inside both of the running containers and
looking for a phrase this change introduced, alongside a second phrase that was
already there to prove the check itself works. Both came back positive on both
machines. I deliberately did not trust the version number — the version that got
deployed already existed before I wrote any of this, so the number would have told
me nothing.

The review council approved it first time round, across twelve reviewers, with no
serious objection. Three smaller points came back and I answered all three in code
rather than in a reply: one about proving the lock check looks up exactly the same
image the job writes (now pinned by a test, and I deliberately broke it first to
confirm the test can fail), and two that were answered by going and measuring
rather than arguing.

**One thing I got wrong and want on the record.** I built a shared piece of code
for the write side of this guard, wrote a comment calling it "the enforcement", and
then never actually used it — all three places that do the writing still had their
own hand-typed copy of the rule. It compiled, the tests passed, and it looked
exactly like a job well done. The session that had been fixing the same bug and
stood down had spotted it and written it in the bug file as something to check.
That is the second time today their withdrawn work improved mine, and it is worth
saying plainly: a session that gives up on a task and writes down what it found is
more useful than one that races to finish. It is now fixed, with a test that fails
if anyone hand-types that rule again.

The bug is closed and filed away. Nothing is outstanding on it. The one open
question — whether an approved image's lock should ever expire on its own — is
recorded against the older project that owns that decision, along with the
measurement showing our current data genuinely cannot answer it either way.
