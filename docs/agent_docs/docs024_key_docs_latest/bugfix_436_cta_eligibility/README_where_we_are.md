# Where we are — CTA eligibility lever (bug 436)

Append-only, newest at the bottom. Plain prose for the owner.

## 2026-09-02 — lane opened

This lane builds the thing you approved on 25 August as "decision 3", now that bug 391 is closed:
a switch on each page that says "never make this page a call-to-action button", plus an alarm that
notices when a site's main button has been won by a page with a suspiciously ancient menu number
(the shape that put the password tool on three consultancies' front pages).

First I checked nobody else had already picked this up — no other session, no lane folder, and
the 391 close-out explicitly left it to a new lane. Then I re-checked the bug is still real
against today's code: it is, unchanged, and the platform's own review queue is currently
complaining about the same shape on another site (a 63-tool site whose front-page buttons are
just the first two tools alphabetically).

The plan, in one breath: a new true/false field on pages (defaulting to "eligible", so nothing
changes until someone flips it); the button-choosing code refuses ineligible pages both in its
ranking and in its label matching (the label half is what stops the lock-in loop); and a new
automatic check that files a review item when a site's top button looks fossil-ranked. It goes
through the council before committing, since it changes what the ranking promises.

## 2026-09-02, evening — built, committed, under review; the switch exists in the database

The whole thing is now written, tested and committed (main commit `215c7eead`, plus three small
follow-ups the repo's own automatic advisors asked for — locking in an improvement they noticed,
and marking a code branch the way their scanner expects). The council is reviewing it now
(reference 9faa2a23); the code itself does nothing until the next fleet software roll, by design.

The new database column is already live — I applied it and every page defaults to "eligible", so
nothing anywhere behaves differently yet. Two things wait for the roll: the code that reads the
switch, and the new automatic check (its enable file is deliberately held back until the software
that knows the check's name is running, because turning it on early would break the nightly
discovery run outright).

Two bumps worth knowing about, both handled: another session took migration number 710 while I was
working, so mine became 714/715; and two of my document edits got swept into other sessions'
commits before I committed — nothing lost, the repo's rules cover exactly this, and my commit
message says where they went. Also told the session working bug 114 that one of their files is
currently failing a shared test for everyone.

Nothing needs a decision from you yet. The one that's coming: once this rolls, do you want any
live page opted out today (the demoted password tool trio are already harmless at their new menu
number, so there may be nothing to do).

## 2026-09-02, late evening — the council approved it

The review came back in two rounds. Round one sent it back with one hard question — "if someone
switches off every tool on a small site, does the button-chooser crash or silently produce
nothing?" — and the honest answer was that the code already handles it gracefully (the build files
a review item, repairs keep what's there, the header just shows no button), but nothing proved it.
So now tests prove it, plus two smaller things the reviewers were right about: a backup is taken
before the config change that turns the new check on, and that change has a proper undo file.

Round two: approved, fifteen seats, including the architecture seat. All committed with the
approval reference. What's left is tied to the next software roll: turn the check on, and run the
two-way live test (switch a page off, watch the buttons move — including the header one, which can
only be checked on the live site; switch it back, watch them return). Then your call on whether
any current page should be switched off.

## 2026-09-03, morning — the new software is running everywhere, and the alarm is switched on

The fresh build rolled overnight. I proved it the careful way (asked each running service what it
contains, with a known-present and a known-absent name as controls — 412 machines carry the new
check, exactly matching the control) and then turned the new automatic check on, taking a backup
of the configuration first and confirming the backup really holds the old version.

So as of this morning: the switch exists on every page (all set to "eligible" — nothing has
changed anywhere), the button-chooser respects it fleet-wide, and the fossil-alarm is armed. What
remains is proving it end-to-end on a real site: flip one page off, watch every button move away
from it (including the one in the site header, which can only be checked on the live site), flip
it back, watch it return. Plus watching the alarm's first sweep of the fleet, and your one
decision — whether any current page should be switched off (probably none needed).

Everything a fresh session needs is in HANDOFF_2026-09-03_continue_here.md in this folder.

## 2026-09-03, mid-morning — the live test passed, and it found four real sites

Two things happened this morning. The first was a check on my own instrument, and it changed how I
read everything after it.

The alarm had filed nothing since being switched on, and I nearly wrote that down as good news. It
isn't readable on its own: a fleet with no problems and an alarm that never ran look identical from
the outside. So before touching anything I did two separate things. I found where the system records,
run by run, which checks it actually executed — so "it ran" is now a fact I can query rather than a
log line that scrolls away. And I wrote out, by hand, the same sums the alarm does, and ran them
across every site we have, to predict what it *ought* to find.

That prediction says **four sites** currently have the fault this was built for — a primary button
won by a page whose ordering number is a fossil nobody has looked at in months:

- **cv1.co.uk** — the button points at a page called **"example"**. On the live site, right now.
- **boxingonline.com** — a fight calendar.
- **vetcomparison.uk** — a compliance deadline calculator.
- **gamesdesign.co.uk** — a TTK calculator, ahead of 23 others.

The check that this prediction is honest, rather than a sum rigged to agree with itself: idea.uk has
a low ordering number on its top page too, and my sums say it is **fine** — the gap to the next page
is small, which means someone arranged that order deliberately. The alarm agrees, and stayed quiet
there. It could have come out the other way and didn't.

Then I asked the alarm to look at two of the four. It found exactly what the sums predicted, naming
the same pages and the same numbers. And on cv1.co.uk I confirmed it independently by fetching the
live page: the button on the wire really does point at `/tools/example/index.html`. So this is not a
theory about a database — it is what a visitor gets.

The second thing was the live test of the switch itself, on cv1.co.uk. Switch the "example" page off:
the chooser's shortlist drops from three pages to two, and the alarm withdraws its own finding,
saying so in those words. Switch it back on: the shortlist returns to three and the finding comes
back. That number moving 3 → 2 → 3 is the point — it is the running software, on the live database,
obeying the new switch. Both directions, as designed.

**One piece is not done, and I could not do it.** Checking the button *in the site header* means
re-rendering and redeploying that site, and my session's safety layer refused to send that command.
Nothing was sent; nothing is half-done. Everything else about that button is established — the
chooser it uses is the one I just tested both ways, and its output is visible on the live page — but
the last step, watching the header button physically move, needs someone to authorise that render.

**Your decision, and it is a real one now.** I put cv1.co.uk back exactly as it was, because switching
pages off is your call, not mine. But the finding stands: four live sites are pointing their main
button at a page chosen by an accident of ordering, and on one of them that page is called "example".
Each has two possible remedies — switch the page out of button-candidacy (changes nothing else), or
move its ordering number (which also moves the visible menu). Tell me which sites, if any, and which
remedy, and I will do it.
