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

**The fix is live**, on v1.0.1199 and carried forward into another session's v1.0.1201 rebuild,
verified both times by grepping the running pods for a string that exists only in the approved
revision. The deploy was held until other sessions' review rounds had finished, because a fleet
restart kills them — which is what happened to my own first round this morning.

**relojistas is finished.** Its card was re-derived after the fix and checked by eye: the
wordmark once, legible, on the brand cream, **and no letterbox rectangle** — which confirms
that defect is a property of the logo file, not the code, and that knocking the background out
resolves it.

**The tab icon needs an honest qualification rather than a tick.** The distortion is fixed —
wide logos are no longer squashed to fill a square. But rendered at the sixteen pixels a
browser tab actually uses, relojistas' icon is a grey smudge: the wordmark puts 19 pixels of
ink in a 64-pixel canvas. That is not a failure of the fix; it is that a long thin wordmark
cannot be a legible tab icon at any resize quality. **The real fix is a square source** — the
gear glyph inside the logo would do — and the deriving code cannot express that, because it
always reads the site's `logo` asset.

**Because of that I dropped the four-site sweep I had planned, rather than doing it quietly.**
Re-deriving fundamentallyai, oufe, robot-hands and vetcomparison would give each an undistorted
but equally illegible icon, at the cost of four slots in a queue currently a hundred deep with
another lane's work. Undistorted-illegible is not materially better than distorted-illegible.

**And the repair list was five sites, not nine.** I had said nine; measuring the deployed logos
shows the stretch only damages non-square sources, and only five are non-square.

## Where we're going

1. **A square favicon source.** This is now the top item and it is the one that actually helps:
   five of fourteen sites have a wide logo and therefore an illegible tab icon, distorted or
   not. Needs a way to say "use this square mark for the favicon" — the deriving code has no
   such concept today.
2. **The tag gate** for sites with no card — unchanged, with its original constraint: do not
   key it on an asset row, or it regresses the one site that always worked.
3. **The bare-domain `og:title` on about eight sites**, and the missing description.
4. **webdesign.co.uk emits no preview tag at all** — never investigated.
5. The two bugs opened along the way: the blind watchdog and the sibling lock gap.
