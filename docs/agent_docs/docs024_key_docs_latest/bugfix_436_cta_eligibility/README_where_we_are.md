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
