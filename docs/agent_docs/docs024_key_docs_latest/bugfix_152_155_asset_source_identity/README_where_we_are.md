# Where we are — asset source identity (bugs 152 + 155)

Append only. Newest at the bottom. Plain prose.

## 2026-08-06 — what this is about, and what we did today

Two bugs had been sitting open, both about the same thing without either file
quite saying so: when the platform wants the ORIGINAL image file that an asset
row stands for, it was working out where that file lives instead of just reading
where it lives.

The bad one was 155. When you ask the platform to deploy a specific image — by
its id, which is the obvious and documented way — it did not go and look up that
image. It looked up a single note the site keeps saying "the last image of this
kind we made", and deployed that one instead. If a site only ever has one logo
and one hero, that note is always right and nothing is wrong. The moment a site
has several images of the same kind, it is wrong for all but one of them. That
used to be unusual and is now normal: one site has 23 heroes, another has 20
icons. It happened for real on dartsonline in July — six different icons were
deployed to six correctly-named files, every deploy reported success, and all six
files were byte-for-byte the same wrong picture. Nothing anywhere said so. You
could only find it by downloading the files and comparing them.

152 was the same idea from the other end, and it had quietly changed shape since
it was written. The bug file describes a problem that no longer happens, and I
nearly fixed that instead — worth saying plainly, because the file reads as
current. What is actually true now is that when we deploy an image we overwrite
the row's "where it came from" with "where it now lives", and we carefully save
the original location in another column that **nothing ever reads**. So we have
been throwing away the only copy of the answer while carefully filing a spare.
107 rows are in that state and recoverable; 49 have lost it altogether. Five
sites' logos were one step away from breaking their favicon and social-card
generation, and a fourth piece of the system — the one that keeps a site's images
looking like each other — had been silently giving up whenever an image it wanted
to copy the style of had been deployed.

The fix is one small shared function that answers "which stored file is this row's
original?", and everything that needs that answer now asks it instead of guessing.
The wrong-picture-with-a-green-light outcome is not possible any more: a row that
cannot say where its original lives now stops and says so, rather than fetching a
neighbour's file. We also record the original's location at the moment an image is
created, rather than hoping to reconstruct it later, and we deleted the
"last image of this kind" note entirely — checked three separate ways that nothing
else was reading it.

Two housekeeping things worth knowing. First, the database half is already done and
live: 205 rows have had their original location written down permanently, so nothing
else can be stranded the way those 49 were. That part needed no deployment. Second,
the code half is committed but **not running yet**. A fresh chassis was deployed this
morning, but it was built at 09:52 and this work was committed after that, so it does
not contain it — I checked the running binaries directly rather than assuming, and
they still contain the old, broken lookup. It will go live on the next build, which
on this shared branch is whenever any session builds.

I have put it through the review council (verdict pending — it does not block, and
the commit is tagged so it gets credited automatically when the verdict lands). The
one thing still owed after it ships is the proof: deploy two same-kind images on one
site by id alone and confirm the two files actually differ. That is the only check
that would have caught the original bug, so it is the only one worth calling closure.
