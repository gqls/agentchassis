# Where we are — the image-404 check

Plain prose, append-only, newest at the bottom.

---

## 2026-07-31, evening

**The problem, in one sentence.** We have a check whose job is to notice when a page
points at an image that isn't there, and it has been unable to see most of them.

Here is how it went wrong. Every image we generate has a *purpose* — "hero", "logo",
"icon" — which is a role, like a job title. Every image also has a *path*, which is
where the file actually sits, like `/assets/images/hero.jpg`. The check was comparing
one against the other: if a page pointed at something called `hero.jpg`, and the site
owned any image at all whose job title was "hero", the check said "fine, nothing to see
here" and moved on. It never looked at whether the file was there.

The effect is worse than it sounds, because job titles are exactly what the common
images have. A site's hero picture is called "hero"; its logo is called "logo". So the
images most likely to be on every page were precisely the ones the check was blindest
to. I measured it today rather than trusting the note in the file: of the 127 image
references across our 13 live sites, the check was reporting 21 pictures that were
working perfectly well, and staying completely silent about six that were broken. Six
sites — dartsonline, gamesdesign, idea.uk, oufe, relojistas and vonc — are painting a
broken image right now, and no sweep we run could ever have told us.

**The fix turned out to be something we already owned.** I did not need to invent a
way of asking "where does this file live?", because the code that *writes* the pages
already knows — there is a small shared function, used by five different parts of the
system, that works out the exact path an image gets published to. Pointing the check at
that same function makes it the mirror image of the thing that writes the pages. If the
writers ever change where they put images, the check moves with them automatically,
which is the sort of fix that stays fixed. New numbers: one false alarm instead of
twenty-one, and nothing missed instead of six.

The one remaining false alarm is worth explaining because it is honest rather than
sloppy. On webdesign.co.uk there is a 455KB image file sitting in the site that no part
of our system has any record of — somebody or something put it there long ago. The
check will flag it, because as far as the database is concerned it does not exist. I
think that is arguably the right answer: it is a file nothing maintains, and the next
tidy-up would delete it without warning anyone.

**Two things came along for the ride.** The first is that the check had never looked at
the site's *chrome* — the header and the bits that appear on every single page. That is
how idea.uk has been serving a broken favicon and a broken social-sharing card on every
page of the site without a single complaint from our tooling. It looks at those now.

The second is smaller but I liked finding it. Some pages carry an image tag with
nothing in it at all — no path, just an empty slot, which browsers draw as a little
broken-image icon. Every path-based check in the world is blind to that, because there
is no path to check. Worse, if we ever built the "just fetch the URL and see" version
everyone keeps asking for, it would score those as *working*, because an empty image
address quietly resolves to the page itself. So the check now spots them structurally.
Three are live today.

**One thing I nearly got badly wrong, and want on the record.** Two of our images —
the favicon and the social card — are published under slightly different filenames than
the rule would suggest (a hyphen where you'd expect an underscore). I assumed the shared
function handled that. It doesn't. Had I not re-read my own comment against the actual
code, the check would have announced a broken favicon and a broken social card on
*every site we run* — a fleet-wide false alarm inside a fix whose entire purpose was
getting rid of false alarms. What makes it worse is that another thread had written a
warning about that exact trap a few hours earlier, and I never looked. That is now
logged as a wrong call, and there is a one-line command in the runbook that would have
saved me.

**Where it stands.** The code is committed and has been put in front of the reviewer
council. It does nothing at all until the next chassis image is built and rolled out —
that is normal here, and the bug stays open until then, because until it ships the
six broken images are still broken. When it does ship, there is a short list of
specific pictures on specific sites to check, in the bug file.

---

## 2026-07-31, later the same evening — it shipped, and it works

A fresh chassis went out (v1.0.1219) carrying the fix, so I was able to finish this
properly rather than leaving it committed-but-dormant.

I checked the running pods rather than trusting the version number, and checked it the
careful way: not just "is the new code in there" but "is the old code *gone*". The old
wording the check used to print returns zero hits on both replicas, which is the answer
that actually settles it — new code can sit alongside old code quite happily.

Then I ran the real thing on two live sites. vonc.com's broken hero image — the one its
own images were hiding, the exact case this bug is about — was reported. idea.uk's broken
favicon and broken social-sharing card were reported for the first time ever, marked as
high priority because they are on every page of that site. And the part I care about just
as much: of the nineteen image references those two sites carry, seventeen are fine and
the check said nothing about any of them. The old version would have shouted about
several.

The bug is closed and filed away.

Two honest loose ends. The reviewer council sent the change back once with a fair
question, I answered it, and then the council could not sit again because it hit its
spending cap for the day — so on paper the last verdict is still "revise". I have marked
the commits as *submitted for review* rather than *reviewed*, because that distinction is
the only thing keeping the review record worth anything, and the resubmission is written
down as owed after midnight. And there are nine stale entries in a queue left over from
the old check, four of which are its own false alarms; I have left them alone because
another thread is working in that queue right now and yanking rows out from under someone
is how you create a confusing hour for both of you.
