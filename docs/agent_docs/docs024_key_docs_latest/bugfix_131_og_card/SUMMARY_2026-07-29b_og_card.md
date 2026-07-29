# SUMMARY — 29 July 2026 (second) — the blocked decision was answered, and relojistas is fixed

Milestone read-out for `bugs_open/131` (og-card slug), session "relojistas 5". Current state
only; chronology is in NOTES and `README_where_we_are.md`. The morning's
`SUMMARY_2026-07-29_og_card.md` ended with "three sites blocked on one owner decision" — that
decision has been answered and acted on, which is why this is a new entry rather than an edit.

## What we're trying to do

Make the estate look like itself when a link to it is shared, and — as of this session's
findings — when a person simply loads the page. Every site advertises a social preview image;
on 11 of 14 that image did not exist. Fixing it turned out to be less about the preview and
more about what we store as each site's *logo*.

## Where we've come from

The generator was already built and had never been run; running it fixed 12 of 13 cards in an
afternoon and revealed the real defect — three sites whose stored "logo" is not a logo, and one
(relojistas) whose spec-sheet logo was also its illegible page header. That work ended blocked:
the corrected logos had to be written to object storage, and reading the storage credentials
was refused by the permission classifier.

## What we've done

**The owner answered all three questions, and every answer has been executed or handed on.**
Storage writes go **through the chassis** — an in-cluster job holding the credentials inside
the cluster did the upload, so no session ever saw them. relojistas' corrected crop was applied
**everywhere**: uploaded to object storage, its database row repointed and **locked** as
owner-approved, and its header image republished. **The header is live and verified by eye** —
the site shows a legible "Relojistas" wordmark instead of a two-up specification sheet crushed
to thumbnail size, on every page, for the first time.

**gaswholesalers and idea.uk were reassigned mid-session to the "relojistas 4" thread.** Rather
than compete, the two sessions agreed a written split in the lane directory. This session had
already generated candidate logos for both (via the image pipeline, nothing installed, both
looked at by eye); they were handed over with seven landmines, **the owner has since approved
both**, and that thread is installing them.

**The platform fix was approved by the council on its second round, and the first round's
objection was right.** The change makes favicon derivation aspect-preserving (wide wordmarks
were being stretched into illegible smears) and makes an "approved, never overwrite" lock
actually protect the *file*, not just its database row — previously the artefact was committed
before the lock was ever consulted. The council caught that the guard trusted an `active`
status filter on a free-text column with no constraint; the filter is gone, so an approval now
fails closed whatever status the record carries. Reviewers also asked whether any sibling code
shares the flaw. It does — the content-card generator — measured, filed as its own bug, not
quietly folded into this fix.

**Two bugs filed, both found by looking rather than by failing:** the watchdog for this whole
class is structurally blind (its starting population is the table of assets that *exist*, so a
missing asset can never appear, while a working one fires forever because its evidence lives in
page markup it never reads); and the content-card lock gap above.

**And one mistake of my own, logged.** I published the header to the wrong deploy repository —
we have two routes, chosen per site by a database column — and when the live page did not
change, I inferred a lagging server that does not exist. One query settled it. It is written up
where the next thread will hit it, including for the other session, whose idea.uk is on the
same route.

## Where we are now

**relojistas: header fixed and live; card and tab icon pending one deploy.** Its logo is now a
genuine, owner-approved, locked brand asset in storage.

**leopardess: protected deliberately at last.** Its hand-made approved card and favicon now
have locked records. Until the new build is running, the thing actually protecting it remains
an accident of a malformed row — so nobody should tidy that row yet.

**The fix is built, pushed and approved but not yet running.** The deploy is being held
deliberately while other sessions' review rounds are in flight, because a fleet restart kills
them — which is what happened to my own first round this morning.

**The repair list is five sites, not nine.** I had said nine; measuring the actual deployed
logos says the stretch only damages non-square sources, and only five are non-square:
relojistas.com, fundamentallyai.com, oufe.com, robot-hands.com and vetcomparison.uk. The other
five are square marks that survive the old code unharmed.

## Where we're going

1. **Roll the approved build** (already pushed), then confirm it on the running pod — not on
   the tag.
2. **Re-derive relojistas' card and favicon**, and look at both.
3. **Repair the four other squashed favicons** (fundamentallyai, oufe, robot-hands,
   vetcomparison), each result looked at.
4. Then the still-open items: the tag gate for sites with no card, the letterbox rectangles,
   and the bare-domain `og:title` on about eight sites.
