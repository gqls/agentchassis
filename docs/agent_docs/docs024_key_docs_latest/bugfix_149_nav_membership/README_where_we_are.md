# Where we are — nav membership (bug 149, Group A)

Plain-prose log, append-only, newest at the bottom.

## 2026-07-31 (morning)

You asked me to take bug 149 — the list of things wrong with the checkers — and fix
it properly, framework-first, checking nobody else was already on it.

**Somebody was, on part of it.** Another session was working in the tree an hour
before me on the same bug's *scheduling and dispatch* half: the finding that we have
about sixty working checkers and almost nothing that runs them on a timetable, and
263 findings sitting in a queue that no handler is allowed to pick up. I left that
alone and took the other half — *routing*: what happens when a checker does find
something and hands it to an agent to fix.

**What I found is the cleanest broken loop I have seen on this platform.** One of
the checkers looks for a page that has been marked as belonging in the site's
navigation but is missing from it, and hands that to the agent whose whole job is
rebuilding navigation. That agent rebuilds the nav, re-renders the header and
footer, and pushes them out to every page. It does everything needed. And then one
line inside it throws the page away before it looks at whether the page was marked
for navigation at all — because the page's address begins with `/tools/`. So the job
finishes, reports success, changes nothing, and the checker finds the same page
again next time.

**It is on the record, not just in the code.** On 29 July a checker filed a job for
gamesdesign.co.uk naming four tool pages. The job completed that afternoon. Two days
later all four are still missing from the navigation. Same story on
ai-agent-orchestration.com back on the 25th.

**And I could prove it was that one line, because the same job worked elsewhere.**
On 28 July the identical checker filed the identical kind of job for robot-hands.com,
naming two pages — and both of those *did* get fixed. The only difference between
the pages that got fixed and the pages that did not is that the broken ones live
under `/tools/`. That is about as close to a controlled experiment as we get here.

**The fix is smaller than the bug.** The code already knew what to do with a page
that shouldn't be in the main menu but has been marked for navigation: put it in the
footer. It says so in its own comments — "tool pages are never top-level". There
were simply two rules doing overlapping jobs, and the cruder one ran first and
answered a different question. I merged them into one rule: *a page's address can
decide **where** it appears in the navigation; it can never decide **whether** it
appears.* Whether it appears is what the page's own settings are for.

**The second thing I found was worse and I nearly made it worse still.** There were
two different bits of code writing to the navigation table: the proper one, which
rebuilds the whole thing from the pages, and an older one that scribbled a single
extra row in when a tool was created. The proper one wipes the table and rebuilds
it, and it could not reproduce what the scribbler wrote — so it deleted it. Seven
links across three sites are in that state right now: present, and destroyed the
next time anyone rebuilds navigation on that site. Six of the seven survive after
this change; the seventh is a link on leopardessconsulting that was added by hand to
a page never marked for navigation, so nothing could ever reproduce it. That one
needs a one-line data fix, which belongs to whoever owns that site — I have written
it down rather than reached into their site.

So I deleted the scribbler. There is now one thing that writes navigation. Where the
old code wrote a row, the new code asks for a rebuild instead.

**The bit I want to flag, because the bug report told me to do the opposite.** The
report said: fix it at creation time, write the navigation row when the tool is
made, because making a bad state impossible beats detecting it. That is usually
right and here it is wrong, and the reason is subtle. A navigation row is not a
link. The header and footer of every page are *saved files* — writing a row into a
table changes nothing that a visitor sees until something re-renders them. But the
checker that spots unreachable pages treats the presence of a row as proof the page
is reachable. So writing the row at creation time would have left the page exactly
as unreachable as before **and switched off the only thing that would have
noticed**. I would have been congratulating myself on a fix whose entire effect was
to hide the problem. So the creators now ask for the rebuild that actually makes the
link real.

**What changes on the live sites: nothing, today.** That sounds like a let-down and
it is the honest answer. Navigation rows only become links when a site's header and
footer are re-rendered, and nothing re-renders them on its own. What this change
does is make the next rebuild correct: 26 footer links across 9 sites that should
have been there and were being silently dropped, at most 5 per site, into footers
that already carry up to 14. I checked the two sites whose tool links exist in the
database today — neither link is in the served header, because nothing has
re-rendered since they were written. So there is no visible before-and-after to show
you yet, and I would rather say that than dress it up.

**Process, briefly.** I put the change through the reviewer council before
committing (it is advisory, and the verdict is still coming). And 26 minutes before
I was ready to commit, you had ruled that a structural claim like this one must go
through the diagnosis loop rather than my own verification alone — with an escape
hatch for stating why you did it differently. I had a good argument for the hatch
and used the loop instead, because it cost nothing while the council was queued.
Worth noting that I only saw the new rule because I checked whether the repo had
moved under me; I had read the rules at the start of the session and would otherwise
have acted on a 90-minute-old copy of them.

**Still open in bug 149, so nobody thinks this closed it.** A tool page marked as
*not* wanted in navigation is still invisible — correctly kept out of the nav, and
absent from the tools listing page that nothing keeps up to date. The two navigation
flags on a page still default to "yes" at the database level, which means a page
often ends up in navigation by inheritance rather than by anyone deciding; changing
that is a schema change and wants its own review. And the footer's "Our Services"
column is still built by a completely separate query with its own opinions. Plus the
scheduling and dispatch half, which is the other session's.

## 2026-07-31 (late afternoon)

It works. Here is the honest version of what "works" means, because there are three
different things it could mean and only two of them are true.

**The fix is live in the running system.** I checked it the only way that counts — by
reading the actual binary inside both running copies of the service, looking for three
pieces of text my change added, and for one piece it deleted. Three present, one gone,
plus a control from an earlier fix to prove I wasn't fooling myself with a grep that
matches anything.

**And it does the thing.** On gamesdesign.co.uk, the site the evidence came from: before,
five tool pages were marked as belonging in the navigation and were missing from it,
including the exact four that a repair job had reported fixing on 29 July. I ran the
repair again. All of them are now in the footer navigation, correctly labelled, and the
main menu is untouched at five items — which matters, because the cheapest way to break a
navigation fix is to break the main menu while fixing the footer. The odd little
"Tools" group that the old code used to scribble into the database has also cleaned itself
up, with no migration, exactly as I predicted it would.

**What is NOT true yet: visitors cannot see it.** The header and footer of every page are
saved files, so the change has to be pushed out to each page — and that push-out queue
**stopped working across the whole platform at 13:21 today**, before I touched anything.
Thirty-four page updates for this site alone are sitting in it. So: the navigation data is
right, the newly-built footer is right and saved, and the pages people actually load still
show yesterday's footer. I have handed that queue problem, measured, to the session that
owns it, and I have not tried to fix it myself.

**Two things went wrong that I want to tell you about rather than bury.**

The first is that my fix broke something else, and I only found it by looking at the
result. Six of the new footer links came out reading "Tools/Damage Formula Designer/Index"
— the page's whole web address, title-cased. The function that shortens navigation labels
had always been handed simple page names; my change was the first thing ever to hand it a
full path, and it had no idea what to do with one. **Nothing in my change touches that
function**, so no amount of re-reading my own work would have found this. Only looking at
the finished labels did. It is fixed, and I have also written the six proper labels
straight into the database so the live site is correct now rather than after the next
deployment. Then, while fixing it, my first attempt deleted the word "Home" from every
homepage link — caught by a test I had written specifically to check I hadn't broken the
ordinary cases. That test is the reason this is a story about a near miss rather than a
second bug.

The second is that I gave the reviewers a wrong answer with a straight face. One of them
objected that my fix depends on a queue that might not be running. I answered with real
numbers — 1,580 jobs picked up out of 1,664 over the past week, every day covered, the
scheduler enabled and firing minutes ago. All true, and all beside the point: the queue
had been dead for two hours at the moment I said it. A week-long average is exactly what a
lane that died at lunchtime looks like. Worse, the specific trap I fell into — reading
"the scheduler fired" as "the work started" — is written down in the notes of the very
session I had read that morning to avoid treading on their work. The reviewer was more
right than my rebuttal. I have corrected it in the three places I had already written it
down, rather than quietly deleting it.

**Where that leaves bug 149.** Four of its twelve items are now closed by this work (two
fixed, one half-fixed, one answered as "not a defect, and here is the reason on the
record"), on top of the three another session closed yesterday. Seven are done, five
remain, and **I am not closing the ticket** — three of the five are the other session's
live work, one needs a database schema change that wants its own review, and one is a new
piece of build work. Closing it would delete the record of five real defects. I would
rather leave it open and honest than tick it off.

## 2026-07-31 (early evening)

Picked this up cold from the handoff. The job on the list was to prove yesterday's label
fix on a second website — one of the reviewers had said, fairly, that checking it on a
single site wasn't enough to justify running it across the fleet.

The first thing I did was re-check the claim at the top of the handoff, that the queue
which publishes pages was dead. The obvious query said it had recovered: something had
been picked up a minute ago. That was misleading, and in a way worth writing down. The
only thing picking work up was **another team's investigation into why the queue is
dead**. The real build queue hadn't touched anything for two and a half hours. So the
measurement was right and the conclusion would have been wrong — the same trap this lane
fell into yesterday, one level down. I've fixed the recipe so the next person has to
break the number down by *who* picked the work up.

Then, choosing the site. The handoff suggested six candidates, and two of them were no
good — not because anything had changed, but because the list had been built on the wrong
property. What matters isn't how many tool pages a site has; it's whether anybody typed
the menu labels in by hand. If they did, the code I'm trying to test never runs. Two of
the six had every label hand-typed, so they'd have produced a clean, meaningless pass. I
picked gaswholesalers.com, where all four tool pages have no hand-typed label at all.

It worked, and there's a small detail that makes it properly convincing rather than just
encouraging. The page is titled "Wholesale Break-**E**ven Volume Calculator", with a
hyphen. The menu label came out as "Break**e**ven Volume Calculator", without one. The
only place that spelling exists is the page's web address. So the label demonstrably came
from the code I was testing, and not from the title by coincidence — which is the kind of
thing a bare "it looks right" check can't tell you.

**And then the useful part, which I wasn't looking for.** While working out which sites
had hand-typed labels, I read the labels themselves — and found that several are written
as "Tools / AI Readiness Quiz", with a slash in them. Yesterday's fix exists specifically
because a slash in a menu label is always wrong. But it only fixed the labels the system
*works out for itself*. The ones a person *typed in* were handed straight through
untouched, as long as they were reasonably short. So the rule was being enforced on one
side of a door and not the other, and one of these has already reached a live website:
ai-agent-orchestration.com is right now showing a footer link labelled "Tools / AI
Readiness Quiz".

I've fixed that. It's a small change — when a label has been typed as "Category / Page
Name", just use the page name, since the link is already sitting inside a menu group and
doesn't need telling twice. I kept the typed words rather than regenerating them from the
web address, because the address gives you "Ai Readiness Quiz" and a person wrote "AI".

Two things I checked before writing it, because both could have made this worse. First,
whether the same mistake lives anywhere else — it does, in an older copy of the same
function that still has the original bug in it. That one turned out to be dead code,
reachable only through a line somebody commented out, so I've left it alone and written
down clearly that it *looks* like a live bug and isn't. Second, whether stripping the
category could let an over-long label sneak through the length limit — it could, so the
strip happens before the length check, not after, and there's a test that fails if anyone
reverses those two steps.

Measured rather than guessed: this changes at most seven menu labels across every site we
run, and fixes one that's live and wrong today. Before committing I checked the guard
test actually catches the bug by putting the bug back for a few seconds and watching it
fail — a test that has never been seen to fail isn't evidence of anything.

The change is committed and has gone to the review council. It won't be visible to anyone
until the next time the software is rebuilt and rolled out, and the four new menu links on
gaswholesalers still aren't on the live site either, for the same reason as everything
else today: the queue that republishes pages is still down, and that remains somebody
else's problem to fix.
