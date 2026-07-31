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
