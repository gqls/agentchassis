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
