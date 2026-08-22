# Where we are — bugs 235 / 155 / 071 close-out (plain prose, append-only)

## 2026-08-22 — starting position and the plan

We looked at three old bug files together because each one turned out to be nearly finished —
the code fixes had already shipped, mostly weeks ago, but each file was still open for a
different kind of leftover.

**Bug 235** (logos saved as heroes) is completely fixed and proven on the live sites. The only
thing keeping it open was a question reserved for you: whether to delete the old wrongly-made
logo.jpg files still sitting in the site repositories. Nothing links to them any more — we
re-checked today. You decided: delete them, using the platform's own new deletion tool, which
refuses to delete anything that is still referenced or still owned by a live record, and does a
rehearsal run first. That is scheduled as the last phase.

**Bug 155** (the wrong image getting deployed because the system looked things up by "purpose"
instead of by the specific image's id) — the dangerous code is long gone. Three loose ends
remained: the real end-to-end test the file demanded was never actually run; the database
change that made deployments carry the image id was applied to the live system but never put
into git (which also makes the migration runner stop at it every time anyone runs it); and two
or three lines of code still WRITE the old "purpose" lookup table even though nothing reads it
any more. You chose the full clean-up: run the test, fix the git bookkeeping, and delete the
dead writes so the whole class of bug becomes impossible rather than just dormant.

**Bug 071** (the build checker finds broken links, then throws its findings away) — the
headline problem was fixed a while ago: links are now repaired before a page is saved. Two real
leftovers: first, when the checker finds problems it can NOT repair, on an otherwise-good page,
the findings still vanish within a day — we're adding a permanent record. Second, while
checking, we found one more place in the section-editing code that invents a "Get Started"
button pointing at /contact.html (a page many sites don't have) — the same disease bug 203
cleaned up elsewhere, so we're removing this last instance. Two bigger leftovers (fragment
anchors that jump nowhere, and a subtle problem with how directory-style links are checked)
you chose to leave for another day; they stay written down in the bug file.

The first test dispatch is already running. The code changes will be committed today and will
ride the next fleet release.
