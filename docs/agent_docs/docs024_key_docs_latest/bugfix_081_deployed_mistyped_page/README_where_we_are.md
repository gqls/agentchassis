# Where we are — bug 081, the deployed page nothing could fix

## 2026-07-31, evening

Picked up bug 081 off the open pile. Finding one that nobody else was already
working turned out to be most of the evening's work, and it is worth saying why,
because it will happen to whoever reads this next.

There are around thirty sessions working this repository at the same time. The
tool we have for asking "is somebody on this bug?" reads the commit history — so
it can only tell you about work that has already been written down. A session
that started an hour ago and has changed nothing yet is completely invisible to
it. I picked three bugs in a row that looked free, and all three had somebody
elbow-deep in them: one had a session with 57 references to the exact function I
was about to rewrite. I only found that by reading the other sessions' live logs
and searching for the *function names* I intended to touch, rather than the bug
number. That is the check that works; the bug number is not enough, because a
session fixing your file may never type it.

081 came back completely cold — the only open bug with no mention anywhere in the
fleet in four hours. So: 081.

## What the bug is, in plain terms

When the platform notices a site is missing a page — say, a news page — it asks a
model to plan one, and then a piece of code creates it. That code was written to
CREATE a page. But if a page with that name already exists, the database
instruction it used quietly switched to UPDATE instead, and only updated *half*
the page: it replaced the title and the content, and left the page's *type* alone.

That is bad in two separate ways, and both were live.

The page's type is what tells the rest of the platform what the page is FOR. So
after this ran, you had a live public page whose content had been replaced by a
plan written for a completely different job, still labelled as the old job. And
because the label never changed, the check that noticed the missing page noticed
it again on the next sweep, and again, and again. On one site
(ai-agent-orchestration.com) that loop has been running since the first of May —
three months. One of its work items has run out of retries; another is still
sitting there waiting.

## The trap in fixing it, and why the obvious fix is wrong

The obvious fix is "also update the page's type". That makes the loop stop. It is
also how you break a working website.

The previous session on this bug had already found out why, and I confirmed it
and made it worse. To fix a mistyped page automatically you have to be able to
tell which page is *supposed* to hold the role. On robot-hands.com the real news
page and the gripper-catalog page are, as far as the database is concerned,
identical — same structure, same contents list. There is nothing to tell them
apart. If the platform guessed, it would have a roughly even chance of relabelling
the catalog page as the news page and breaking a live, working page. When I ran
the query myself I found a *third* page with the same shape that the bug file had
not noticed, which is archived. So the guessing gets worse the more you look, not
better.

## What I did instead

I stopped the code from guessing at all, and stopped it doing damage while it
guessed.

The page-creating code now only creates. If the name is already taken by a page
that is **live** and has a **different** type, it changes nothing whatsoever. It
files a note for a human that says "this site wants a news page; the page called
'news' is live and labelled 'content'; here is the one line of SQL that fixes it
if that is what you want" — and it marks the original job as *blocked* rather
than *complete*, because calling it complete when nothing was repaired is exactly
the kind of false green that let this run for three months unnoticed.

The key insight is that we never have to solve the hard question. The planner has
already picked the page name. We don't need to work out which page *should* be
the news page — we only need to notice that the name is taken by a page doing a
different job, and ask.

## One decision I made and then reversed

I had also written a second half: quietly fix the page type for pages that have
never been published, where relabelling is harmless. It sounded free.

Then I actually counted. Every mistyped page in the entire fleet — all five — is
published. There are none in the unpublished state. So that half would have fixed
nothing that exists, while giving an automated component broad authority to
relabel pages. I deleted it, and left a test behind that will fail if somebody
adds it back without re-running the count first.

## What still needs you

**Two live pages are still mislabelled and I have not touched them**, because
relabelling them changes what they serve to the public immediately, and that is
your call, not mine:

- `ai-agent-orchestration.com` — the page `news` is labelled `content`. Fixing
  this almost certainly is what you want: it closes a three-month-old loop.
- `idea.uk` — the page `news-index` is labelled `section-index`. Note that this
  one is not just a label problem: its content is stale and empty for a separate
  reason recorded elsewhere, so relabelling alone will not make it right.

The one-line SQL for each is in the bug file, and from now on the platform will
also put it in the work item it files.

**The fix is not live yet.** Go code does nothing until the chassis image is
rebuilt and rolled out, which happens when somebody builds it. I have committed
it so the next build picks it up. After that roll, there is a scripted check that
re-runs the failing case against production and confirms the page is genuinely
left untouched — that is written down and owed, and until it has been run I would
not call this proven, only committed.

## 2026-08-01, after the review

The review council came back and said **revise**. Two things came out of it, and
the second one is mine and is worse than theirs.

### What the council caught

Four of the reviewers, independently, pointed at the same thing. When my code
files that "a human needs to decide this" note, it was writing the note into the
database by hand instead of going through the platform's shared function for
creating work items. That shared function does two things my hand-written version
did not: it stops the same note being filed twice while an earlier one is still
open, and it notices when something has been filed, resolved, and come back
repeatedly. So the very thing I claimed in my write-up — "a repeat won't file a
duplicate" — wasn't actually guaranteed. It is now.

Fixing that had a useful side effect. The shared function insists on being run
inside a database transaction, and a different reviewer had separately pointed
out that my code read a page's status and then acted on it without one — meaning
a page in the middle of being published could be read as "not published yet" and
overwritten a moment later as it went live. That is the exact damage this whole
fix exists to prevent, arriving through the fix itself. Both are closed by the
same change.

The reviewers also asked me to prove several things I had merely asserted. All of
them checked out, and a couple came out stronger than I had claimed — for
instance, exactly one thing in the whole platform uses the code I changed, and
nothing at all reads the field whose value I altered.

### What I got wrong, and how it surfaced

**My test was decoration.** I wrote a test whose stated job was to prove that the
refusal path does not touch the live page, and I put that claim in five places
including the commit message and the review submission. The test could not fail.
The function I was relying on to catch a stray database write only checks that
the things you *told* it to expect actually happened — it never notices an *extra*
one. And my code was throwing away the error that would have revealed it.

I did not reason my way to this. I proved it by deliberately adding the exact
forbidden write to the code and running the test: it passed. After the rewrite,
the same edit fails, which is what a working guard looks like.

The uncomfortable part is how I found out. It was not the council and it was not
me reviewing my own work. A safety check on one of my commits complained that I
had removed lines from a file that is only supposed to be added to. Those lines
turned out to be another session's tidying, which had ridden along in my commit
because we share one working copy. While reading that to make sure I had not
destroyed someone's work, I saw an entry they had written a few hours earlier
warning about exactly this testing trap — they had hit it the same day, in four
places, in a different part of the system.

So: an unrelated warning about an unrelated file is the only reason a false claim
of mine got caught. I have written it up in the shared log of wrong calls, because
the lesson generalises and it is not a comfortable one: I had written three
paragraphs *in that same commit* about how to stop tests being vacuous, and the
technique I described does not detect this. The only thing that does is breaking
the code on purpose and watching the test fail. It takes a minute.

### One more thing found, and deliberately not fixed

A reviewer asked whether the original mistake appears anywhere else rather than
assuming it was a one-off. It does: **four more places** in the codebase create
pages the same way and drop the page's type in the same manner, and a fifth does
the opposite. I have written that up as its own item (172) with the survey done,
but I have not fixed it here. Changing six places at once inside a bug fix is
exactly the sort of sprawl the review process exists to stop, and the right answer
genuinely differs between them — for one of them, the current behaviour may be
correct.

### Where it stands

The revision is committed and back with the council. Everything from my earlier
note still holds: it is **not live** until the chassis is rebuilt, the two
mislabelled pages are still untouched and still your call, and the production
check is written down and owed.
