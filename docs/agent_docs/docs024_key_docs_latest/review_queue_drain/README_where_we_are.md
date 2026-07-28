# Where we are — the human review queue

Plain-prose running log, append-only, newest at the bottom.

---

## 2026-07-25 — opening the drain

You have a queue of things the platform decided a person should look at. It has
370 items in it. Nobody has ever actioned a single one.

That is not because nobody could get to it. Five days ago another session fixed
the two reasons it was invisible — the dashboard was only ever loading the newest
50 items so the queue reported itself as empty, and there was no way in from
outside the cluster — and set up a VPN so you can reach it. Since that fix the
queue has grown by 67 and shrunk by nothing.

So I went looking for why, and the answer is not "nobody has got round to it".

**Most of what is in there is no longer true.** 321 of the 370 items describe a
page that has been rebuilt since the item was filed, and nothing re-checks them.
Here is a concrete one. On leopardessconsulting.co.uk, the "how we work" page has
two items from 10 July saying its two buttons have nowhere to point. The page was
rebuilt on 18 July, and both buttons now point at real pages. The items are still
sitting there saying otherwise. If you sat down to work this queue tonight, that
is the kind of thing you would be reading — findings about a version of the site
that stopped existing weeks ago — and you would have no way of telling those from
the ones that are still real.

That is the actual defect. Not the size of the pile, the fact that you cannot
trust any item in it.

**What I have built.** A sweep that goes through the parked items and re-checks
each one against the site as it stands today. Three possible answers per item:

- *this is no longer true* — close it, and record exactly what it checked and
  what it found, so the close is auditable rather than a bulk delete;
- *this is still true* — leave it in the queue, but stamp it with today's date,
  so when you open the queue you can see it was confirmed rather than merely old;
- *I can't tell* — leave it alone and say why.

The third one matters as much as the first. Some components on our sites render
from templates rather than from stored content, and for those the sweep genuinely
cannot answer the question. It says so instead of guessing. On today's numbers it
closes 51 items, keeps 35 with a fresh confirmation, and openly declines 72.

**Why it is safe to let a machine close these.** If it gets one wrong, the check
that raised the item in the first place will simply raise it again next time it
runs — closing an item releases the lock that was stopping that. So a mistake
costs one duplicate, not a lost finding. That is what let me build it as an
automatic sweep rather than something that needs you to approve each one.

**One thing in the original bug write-up turned out to be wrong**, and I want to
flag it because it was the fix everyone assumed we would do. The bug file says
there is an existing piece of machinery for exactly this — a "section data
reconciler" — that just needs wiring up, and that it would clear 48 items. I
checked: it would clear zero. It only handles a kind of missing data that no item
in the queue actually has. It is not broken, it just does not fit this pile. I
have written the correction into the bug file and the wrong-calls log.

**Where it goes next.** The code is written and tested and has gone to the review
council. It ships switched off — it will report what it *would* do without
touching anything, so we can read that first and only then let it write. Two
questions in the original bug are still yours and I have not touched them: what
to do with the ~78 items that are actually machine failures parked in the human
queue by mistake, and whether we ever record *who* made a decision (the handlers
have no login attached to them, so that one is an authentication decision, not a
one-line fix).

---

## 2026-07-25, later — the council caught something, and you changed the shape of the job

Two things happened after I wrote the above.

**The review council approved the drain, but one reviewer caught a real problem
in my reasoning and was right.** I had said the sweep is safe to let run
unattended because if it closes something wrongly, the original check will just
raise it again — closing an item releases the lock that was stopping that. I
wrote that as a fact. It was not; I had worked it out from how the database index
is defined rather than checking it.

Checked properly: half of it holds and half of it does not. The checks genuinely
can raise a finding again — nothing blocks them. But they only run when a page is
rebuilt or when a site gets a discovery pass. They do not run on a timer. So for
a page that never gets rebuilt again, a wrong close would be a quiet loss, not a
one-cycle inconvenience. Worth knowing too: in the whole history of the platform
only 8 items of these three kinds have ever been closed at all, so that
re-raise path has barely been exercised.

I have corrected it everywhere I claimed it. The safety net that does hold no
matter what is the paper trail — every close records exactly what it checked and
what it found, so a bad one can be found and undone. That was always there; it
just is not what I had been leaning on.

The reason I am writing this at length is that the wrong claim read exactly like
the right ones. The 321-of-370, the leopardess example, the "that existing fix
would clear zero items" — those all came with a query attached. This one did not,
and nothing in how I wrote it up told you which was which.

**And your answer on the 78 stuck items changed the job.** I offered you four
ways to tidy them up and you rejected all four: the framework should be able to
answer every one of them. If the email is a placeholder, the placeholder should
not be on the site. If the content failed a quality check, rewrite the content.
If data is missing, go and find it.

That is a bigger and better answer than the question I asked, and it means this
bug is not really "the queue needs a drain" — it is "the queue should not be
filling". I have found the exact places where it fills: three configuration lines
in the page builder that send failures to the human queue instead of dealing with
them, and one in the tool improver. All four are database configuration, which
means they take effect immediately with no rebuild.

I have not changed them. Two reasons, and I would rather say them than quietly
act. First, those lines are in the main page-building pipeline for every site, so
changing them takes effect everywhere the moment I press return — that is a
bigger step than the image roll you just told me to hold off on. Second, the
obvious change is wrong: simply sending failures back to be retried would have
the same page rebuild and fail on the same problem over and over, spending money
each time, because the retry counter is not being incremented either. Doing this
properly means giving the builder a way to rewrite content in response to what
the checker complained about, and that deserves its own design pass.

So: the drain is built, reviewed and committed, waiting on the next build. The
bigger job you have just described is written up with the four exact places it
needs to happen, ready to start.

---

## 2026-07-28, later — you can sit down and work the 50 today; one earlier claim corrected

You said you would take all the decisions yourself, working through them in the
admin dashboard, and asked for the surface to be checked. It is checked, and the
news is better than this morning's note said.

**The correction first.** This morning's handoff said the 50 items that need an
answer from you were invisible to the dashboard — parked under a status the
screen does not show, needing to be "promoted" before you could see them. That
was wrong. I measured it before building anything: all 50 are already in the
queue the dashboard shows by default, and they have been all along. So decision
A — "promote the 50" — needs no doing. There is nothing to promote. You can
simply open the dashboard and they are there.

**How to open it, today:**

1. In the repo, run `make dashboard-port-forward`, then open
   http://localhost:8080 in your browser and log in.
2. Set the status filter to "Needs Review". Then use the type dropdown to pick
   the class you want: `needs_section_data` (42 — the questions, some waiting
   since March), `owned_page_review` (6), `incomplete_page_group` (2).
3. For the section-data ones, the item opens with a form built from exactly the
   fields the platform is missing — the pricing tier names and so on. Fill in
   what you know, press **Save & Rebuild**, and your answer is written into the
   site's spec and a rebuild of that page is queued automatically, at high
   priority. The item closes itself. This path is real code, read end to end
   today — your answer does not go into a note somewhere, it goes into the site.

I verified the whole journey from a browser's point of view — the page loads,
the login screen talks to the auth service, the API answers behind it, and the
running dashboard genuinely contains the paging and filter fixes from the 20th
(checked against what the pod actually serves, not the version label). The only
thing I could not do is log in as you, because the credentials are yours. So the
first real test is you opening one item — five minutes, and if the login itself
misbehaves that is the one part I could not exercise.

On how you reach it longer-term: the port-forward above works today and costs
nothing. The VPN route you chose in July also still exists. Building a permanent
web address for it is new work — worth it only if working this queue becomes a
weekly habit, and I have not started it.

Decisions B (turn the 186 advisory items into a report) and D (refuse to write
findings that nobody reads) are untouched and still yours; nothing in today's
work pre-empts them.
