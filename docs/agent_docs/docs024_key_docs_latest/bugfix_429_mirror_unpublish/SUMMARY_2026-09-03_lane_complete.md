# SUMMARY — bugfix 429 lane, 2026-09-03: complete

**What we were trying to do.** Close the last gap in the platform's unpublish
chain: the hosted mirror (the copy every customer site actually serves from in
its first month) could copy pages but never remove one, so a retracted page
kept serving for ever. Found on the first paid site, minutes after the
platform's first successful page retraction.

**Where we'd come from.** The delivery and boxingonline sessions filed the bug
on 2026-09-02 with the root cause already nailed (the mirror code has no
deletion path at all, and the orphan's frozen header proved it wasn't a delay).
Every other link in the chain — archive, navigation, git, origin bucket —
already did its deletion half. The bug was explicitly unowned.

**What we did.** Taught the mirror to converge: after copying, it now deletes
hosted files whose source is gone and verifies they're gone, with guards so it
can never be tricked into deleting a whole site (an empty source refuses; a
sweep that would remove most of the site needs an explicit human flag). The
"did it work" check now proves both directions at the real website: the removed
page must return not-found, a kept page must still serve. To clean up orphans
created before the fix, we used the drift detector's own designed lever (a
version prefix bump) so every mirrored site republished itself exactly once —
no manual intervention anywhere. Reviewed by the peer lane, an adversarial
reviewer, and the council (approved first round; all four advisory points
answered with measurements).

**Where we are now.** Live and proven. The retracted contact page returns 404;
every kept page serves; both mirrored sites converged on their normal hourly
schedule with nobody forcing anything; the sweep's own record shows it deleted
exactly the one right file. Bug 429 is closed, the owner's page-deletion
ruling is satisfied on both halves, and the peer lanes have struck their
blocked items.

**Where we're going.** Nothing further owed on this lane. Two known residuals,
both recorded where the next person will trip on them: retracting the LAST
page of a site still can't unpublish the whole site (bug 304's decision — the
new flag is the ready hook), and anything hand-placed directly on a mirror now
gets swept on the next publish (a landmine entry warns about it).
