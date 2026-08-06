# SUMMARY 2026-08-06 — asset source identity: the fix is live, the proof is half done

First read-out for this lane. Written to be read aloud.

## What we're trying to do

Make the platform, when it needs the original file behind an image, **read where
that file is** instead of working it out. Two open bugs said the same thing from
opposite ends and neither said it in those words, which is part of why both had sat
open for a week.

## Where we've come from

Bug 155 was the sharp one. When you asked the platform to deploy a particular image
— by its id, the obvious and documented way — it did not go and fetch that image. It
read one note the site keeps, meaning "the last image of this kind we made", and
fetched whatever that pointed at. On a site with a single logo and a single hero
that note is always right. The moment a site has several images of one kind it is
wrong for all but one of them, and that is now the normal case: one site has
twenty-three heroes, another twenty icons. It happened for real on dartsonline in
July — six different icons, six correctly named destination files, six green ticks,
and all six files byte-for-byte the same wrong picture. Nothing anywhere said so.
You could only find it by downloading the files and comparing them.

Bug 152 was the same fault seen from the other side, and it had quietly changed
shape since it was filed. The file describes an expired-link failure that can no
longer happen, and we nearly fixed that instead — worth saying plainly, because the
file still reads as current. What was actually true is that when we deploy an image
we overwrite the row's "where it came from" with "where it lives now", carefully
saving the original location in a second column that **nothing ever read**. We had
been discarding the only copy of the answer while diligently filing a spare. A
hundred and seven rows were in that state and recoverable; forty-nine had lost it
altogether. Five sites' logos were one run away from breaking their own favicon and
social-card generation. And a fourth piece of the system that nobody had listed —
the part that keeps a site's images looking like each other — had been silently
giving up whenever the image it wanted to match had been deployed.

## What we've done

One small shared function that answers "which stored file is this row's original?",
and every part that needs that answer now asks it rather than guessing. The
wrong-picture-with-a-green-light outcome is no longer expressible: a row that cannot
say where its original lives now stops and says so, instead of quietly fetching a
neighbour's file. We also write the original's location down at the moment an image
is created, and we deleted the "last image of this kind" note entirely, having
checked three separate ways that nothing else read it.

The review council approved it first time round, with eight advisory objections and
none serious. Two of them were right about real things, which is worth recording
because our instinct was that they were generic caution. One said our
"nothing else reads this" check couldn't see callers built at runtime or living
outside the codebase — and widening it found a live operator crib sheet that hands a
person one address per *kind of image*, which is this very bug in manual form. The
other said a migration number isn't yours until you've looked; a concurrent session
had claimed the number we took, within the hour.

We also caught ourselves. A test we had written to protect the fix passed against
code with the bug reinstated, because both halves of its key example pointed at the
same file — a test named after a property it could not observe. Only deliberately
breaking the code found it. Both that and the schema misstep are in the
fleet-wide wrong-calls log.

## Where we are now

**The fix is live.** It shipped on chassis v1.0.1259 and we proved it at the
binary on both replicas, not at the tag: the new code is present, and — the part
that actually matters — the deleted branch's own log line is now absent, with a
nonsense control confirming the check can tell the difference. The database half was
already done and needed no deployment: two hundred and five rows have had their
original location written down permanently, so nothing can be stranded the way those
forty-nine were, and all five at-risk logos can now find their source.

On the founding case, the arithmetic is stark. Across dartsonline's twenty icons the
old route produced **one** distinct source; the new route produces **twenty**. That
is the bug and its fix in a single line.

**Both bugs nonetheless remain open, deliberately.** What we have proved is that the
right code is running and that it is being handed the right inputs. What we have not
done is run a real deployment of two same-kind images and confirmed the two files
differ. That is the only check that would have caught the original fault — success
messages and correct filenames were both already true while it was shipping identical
bytes — so it is the only one we will treat as closure.

## Where we're going

Run that deployment, compare the files, and look at one of them. Then re-derive a
favicon on one of the five recovered logo sites, which is the equivalent proof for
the other bug. Both bugs close on those two results, and not before. Neither needs
another build; both need about ten minutes and a willingness to open the image.
