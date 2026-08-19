# SUMMARY 2026-08-19 — meta descriptions

Written to be read aloud. First summary in this lane.

---

## What we're trying to do

Every page on every site should carry a meta description: the one sentence a search
engine prints underneath the page title in its results. It is also what our own blog and
article index pages show as each card's blurb, so it is read by far more people than the
page it describes.

The aim is that every page has one, that the system writes them rather than a person, and
that once written they stay written.

## Where we've come from

This started as a small job handed over from another piece of work. The Platform Log
index on fundamentallyai.com lists six articles as cards and none of them is clickable.
The repair for that was designed, approved and applied the day before; it would not go
live because the system refused to save the new version of the page. The reason given was
that five of the eight articles have no meta description, so the new card list came out
at 42% of the length of the old one, and a safety guard refuses to replace a page section
with less than half its previous size. That guard is correct and we have not touched it.

So the job was: get five sentences written, by the framework rather than by hand, and the
page unblocks.

## What we've done

The five turned out not to be five.

**First, the explanation we were given was wrong.** The note said the descriptions were
missing because of a known queue problem — the system detects the fault but has nobody
assigned to fix it. That check turns out to look at three things: the page title, the
"skip to content" link, and the footer. It has never looked at meta descriptions. Fixing
that queue would not have filled a single one of the five, and we would have spent a day
on it.

**Second, it is 407 pages, not five.** Across the estate, 407 of 731 live pages — 56% —
have no meta description, on 26 of our 27 sites. Three sites have none at all on any
page.

**Third, we found why, and there were two separate causes.** When the planner designs a
site it is handed a template of what to return for each page: name, title, navigation
label, ordering, sections. That template never had a field for a description, while the
code that saves the page asked for one anyway and quietly accepted nothing. So every page
was born without one. Separately, one line in the same save routine would overwrite an
existing description with a blank every time a site was replanned — and we proved that
had really happened, not merely that it could: four pages on one site held real
descriptions of 97 to 329 characters in an April snapshot and hold nothing today.

**Fourth, nothing could repair one.** No code anywhere updated that column on a page that
already existed. None of the 58 automated site checks looked at it. And the one route
that looked promising was worse than nothing: jobs are already being filed about missing
meta descriptions, they complete successfully, and they do not write anything — we tested
two and both target pages are still empty today, one of them demonstrably touched by the
handler on the way past.

**Then the owner chose the full fix, and we built it.** The overwriting line is guarded.
The planner is now asked for a description, with a short instruction about writing it for
a visitor rather than describing the build — that part is configuration, so it went live
immediately. And there is now a piece that can write a description onto a page that
already exists, which is the thing that genuinely did not exist before. We kept it
small: the framework already knows how to find pages and how to write a sentence, so all
we added was the ability to save one.

**The review process then caught us out, and it was right.** We put the code through the
platform's own review council. It came back asking for changes, on the grounds that
there are several page-saving routines and we had only fixed one. We checked instead of
arguing. There were three more with exactly the same fault, and the most likely of them
to have actually caused damage was the one that adopts pages from an existing website: it
returns an empty description whenever the source page has none, so re-adopting such a
page wiped whatever description we had. All four are now fixed.

## Where we are now

The causes are closed and the tool exists. **No page has been filled in yet**, and that
distinction matters more than it sounds — "the fix shipped" is not "the problem is gone".

The configuration half is live. The code half is committed and waiting for the next
release; there are around 145 commits queued for it, ours among them. The workflow that
actually runs the backfill is written but deliberately held back, because the code it
calls is not in the running system yet, and a job that names a function the server does
not have simply fails. We named the file so the migration tool cannot apply it by
accident.

We have also not put it on a schedule. It writes copy that appears on the owner's sites
under his name, so the first runs should be one site at a time with somebody reading the
result. The platform has been bitten before by a generator that ran unattended and
published internal build instructions where the sales pitch should have been.

## Where we're going

Three things, in order.

The next fleet release carries the code. After that, the roughly 295 pages the planner
manages will pick up descriptions the next time their sites are replanned, and the
remaining 112 need the held workflow to be switched on, one site at a time, with the
output read before it goes further.

The five pages blocking the Platform Log index are in the first group, so that job — the
one this all started as — finishes then, and the check is at the served page rather than
in the database.

And the review round is not finished: we have resubmitted and owe it a reading of the
verdict. The first round found a real defect we had shipped as closed, so the second is
worth waiting for rather than assuming.
