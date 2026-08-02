# Where we are — the sibling link-repair bug (`bugs_open/136`, section-editor)

Plain prose, append-only, newest at the bottom. No jargon where I can avoid it.

---

## 2026-08-02, morning — what this is and why I picked it

I went looking for an open bug nobody else was working on. That check is worth describing,
because it changed my answer twice.

The obvious pick was a bug about images being deployed wrongly (155): high severity, cause
already proven, fix already sketched. Then I searched the live transcripts of the other
sessions running on this machine and found one of them had mentioned that bug 119 times in
the last few hours. It is being fixed as I write. The repo's own ownership script did not
know — it reads commits, and a session halfway through a fix has not committed anything yet.
A second candidate (093) turned out to be waiting on a scheduler that has been switched off
since May, so there was no code to write.

What I settled on is this. When we generate a page, a writer sometimes invents a link to a
page that does not exist — "see our pricing" pointing at `/pricing` on a site that has no
pricing page. Visitors click it and get a 404. A while ago we built a repair for that: if
the target really exists under a slightly different address we fix the address, and if it
does not exist at all we keep the words and drop the link. The repair was wired into the
place where a whole page's sections are saved.

The complaint that became this bug came from our own review council: nobody had checked
whether that was the *only* place a page's HTML gets saved. It was not. There are several
other pieces of code that write the same column, and none of them repaired anything. So the
more carefully you followed our own documentation — which tells you to use the targeted
"edit one section" path — the more reliably you skipped the protection.

## What I actually changed

Two of those writers now repair their links: the targeted section-editor, and the report
generator. They both call one new shared function, which is a thin wrapper around the repair
that already existed — deliberately, so there is one set of rules and one place to get them
wrong, rather than two copies that drift apart.

While doing it I found the section editor saved its work in two different places depending
on which kind of edit you asked for. That is the same trap one level down: put the guard
before one of them and the other still slips past. So the two saves are now one save, with
the repair in front of it. Nothing can reach the database from that action without going
through it.

I also added a check to the script that runs when anyone commits. If a future change writes
that column and does not repair links, the author gets told at the moment they do it. This
matters more than the two fixes: between the day this bug was filed and today, somebody added
a *new* writer of that column, and nobody noticed. A list in a bug file cannot keep up; a
check at the moment of the edit can.

## What I deliberately did not do, and why

- **The blog listing.** It is on the bug's list of unprotected writers, and I nearly "fixed"
  it for consistency. Then I looked: every link it emits comes from our own page table — the
  same table the repair checks against — so the repair could never find anything to do. It
  would have been a database query per rebuild for a guaranteed nothing. Left alone, with the
  measurement written down so the next person does not re-argue it.
- **The two tool-page writers.** These genuinely do have the problem, and they hold the
  largest share of the live damage. I left them out because several other sessions are
  editing those exact files right now, and our commit rules cannot stop two sessions'
  changes to one file getting mixed together. They are named as the next job, and the new
  commit-time check will keep pointing at them until somebody does it.
- **The big structural version** — making every one of the nine writers physically unable to
  save without repairing. That is a change to a shared mechanism, which by our own rules is
  an architecture-review decision rather than something to slip inside a bug fix.

## The honest size of it

I measured what is actually broken out there today: **35 links in stored page HTML that
point at pages which do not exist**, spread over 13 components on 6 sites. Seventeen of them
would simply be removed as dead; eighteen are near-misses we can repair automatically.

Two caveats I want to be straight about. First, that stock cannot be blamed on the writers I
fixed — a stored link does not record who wrote it, and the older fix has been live for a
while, so some of this predates everything. Second, neither of the two paths I guarded has
actually run in the twenty days of history we keep. So this is prevention on paths that are
live, documented and reachable, not a bleed I have stopped. I would rather say that plainly
than let the fix sound bigger than it is.

## Something I got wrong, in public

The numbers in the paragraph above are the second set I wrote. The first set went into the
commit message — "30 links across 7 sites" — and was wrong in every figure. I had read them
off the bottom of a results table on screen instead of asking the database to count. Writing
the runbook is what caught it, because making the query re-usable meant adding a proper
count, and the proper count disagreed with what I had already committed. Commits cannot be
amended here, so the correction lives in the bug file and the working notes, and in the
fleet-wide log of wrong calls.

## Where it stands

The code is committed and it is in the review council now (submitted before committing, which
is what our rules ask for when the verdict will land later). It is **not live** — Go changes
do nothing until a new chassis image is built and rolled out, and I have recorded the exact
before-and-after checks to prove it when that happens: one string my change adds, one string
my change removes, and one from the older fix that must stay put. Until then the bug stays
open, because the defect is still reproducible on the running system.

## 2026-08-02, afternoon — it is live, it is closed, and it found something else on the way out

Someone else's build went out as v1.0.1229 and carried this fix. I proved it three ways on
both running copies of the service: a phrase my change adds is now present, a phrase my
change deletes is now gone, and an older phrase that had to survive is still there. The
middle one is the one that matters — a new phrase appearing only tells you *something*
shipped, whereas a deleted phrase disappearing tells you it is *this* code.

I also checked that the live path I had to touch on the way through is still behaving: it
has written 21 correctly-formed records since the roll. One caveat I have written into the
ticket rather than glossed: the specific piece of shared plumbing I extracted only runs when
a database read fails, which has not happened, so it is proven by tests and by the fact that
three callers compile against it — not by real traffic.

The ticket is closed and moved.

**On the way out it turned up a different bug, and this one is worth explaining.** Re-running
my count of broken links, one new entry had appeared on a veterinary site. It was not a
broken link. It was a piece of JavaScript that *builds* a link while the visitor uses the
tool — the code contains the text `href="' + q.link + '"`, which to a human is obviously a
program and to our repair's pattern-matching looks like a link with an empty address. So the
repair "fixes" it by deleting the link from the program. The result is still valid
JavaScript, the page still reads correctly, and the button simply stops being clickable.

I proved it by running our real repair code over the real stored bytes and watching the link
vanish, then filed it as its own ticket (180). Exposure today is one component on one site.

The consequence I care about most: I had written that the obvious next step was to switch on
this same repair for tool pages. **That is now the wrong order.** Tool pages are exactly
where JavaScript builds links, so switching it on there today would quietly delete working
buttons. The new ticket has to be fixed first. I would rather find that here than in a
client's inbox.

---

**2026-08-03 — the JavaScript problem is fixed, and fixed once for everything rather than
once for this case.**

Picking this up from the handoff, the first thing I did was re-run our repair code over six
different bits of real-looking markup instead of just the one from the ticket. Five of the six
came back damaged, and only one of them was the case we had written down. A link built with
JavaScript string-joining gets deleted one way; a link built with the newer backtick syntax
gets deleted a *different* way; a link written inside a stylesheet comment, inside a text box,
or inside a commented-out block all get rewritten too. Same single cause, five different faces.

That mattered for the decision. The cheap option was to teach the repair to recognise the
exact JavaScript spelling from the ticket and skip it. That would have fixed one of the five
and left the other four, and it would have gone stale the moment someone wrote a link a sixth
way. So instead I gave the platform one shared answer to the question "which parts of this
page are not really markup?" — scripts, stylesheets, text boxes, and comments — and pointed
the two bits of code that *rewrite* pages at it. Both are fixed by one change, and so is any
future spelling.

**What I checked before believing it.** I ran both the old and the new code over every one of
our 509 pages, on all 19 sites. Today exactly one page is being damaged (the vet-comparison
CMA tool, as the ticket said), eleven places across the fleet are at risk, and — the number I
actually cared about — **zero** genuine repairs are lost by the change. That last one is the
thing that could have gone wrong: a guard that is too cautious stops fixing real broken links,
and it would not be obvious. It doesn't.

I also wired the same protection into a second, older function that deletes dead buttons from
site headers and footers. It has the identical blind spot and its mistake would be worse — it
deletes the whole button rather than just the link. Nothing on the fleet triggers it today, so
that half is prevention: measured to change nothing now, and closed before something writes
the markup that opens it.

**One thing I got wrong, and it is slightly humbling.** I made a deliberate technical choice
(what to blank the ignored regions *with*), wrote a test to protect that choice, and wrote a
confident comment saying "change this and the test fails". Then I actually changed it — and
the test passed. A second safety net further down had quietly caught it, so my test had been
proving nothing. The previous session in this very lane had written down that exact trap, in
its handoff, which I had read that morning. Reading it did not stop me doing it. I have logged
it in the shared mistakes file, because the useful lesson is not "be careful" — it is: never
write "change this and it breaks" until you have changed it and watched it break. It costs
about thirty seconds.

**Where this leaves us.** The blocker is gone: switching the link repair on for tool pages —
the step I said last time we must *not* take — is now safe to take, and that is the next piece
of work. The change is committed and sent to the review council; it is not live until the next
chassis image ships, and I will not call the ticket closed until I have proved it on the
running system rather than in git.
