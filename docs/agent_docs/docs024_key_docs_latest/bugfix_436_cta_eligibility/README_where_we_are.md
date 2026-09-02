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
