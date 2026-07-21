# SUMMARY — bug 037, needs_rebuild composition guard (2026-07-21)

**What we're trying to do.** Decide, on evidence, whether a `needs_rebuild` page should keep its
built layout when a site is re-planned — and if so, protect it. Bug 037 was filed as a deliberate
open decision left at the edge of bug 001's re-plan guard.

**Where we've come from.** Bug 001 stopped a re-plan silently redesigning or dropping a *deployed*
page, but left `needs_rebuild` pages outside the protected set. The handoff argued it *might* be the
wanted behaviour, reading `needs_rebuild` as an explicit "recompose me" request.

**What we've done.** Checked the code: every one of the four writers of `needs_rebuild` keeps the
page's layout and means "re-render as planned", never "recompose from scratch" — two of them flag the
rebuild precisely so a component the layout already names gets rendered. That refutes the "recompose
me" reading, so 037 is a real defect. Implemented the fix as a separate membership predicate
(`realisedPageCompositionIsPreserved` = `deployed OR needs_rebuild`) that composes cleanly with a
concurrent session's in-flight bug-050 work in the same file, and wrote discriminating tests
(including one that proves the shortcut fix would break a genuinely-uncomposed page).

**Where we are now.** The fix is **live fleet-wide on v1.0.1146** (it rode an owner sweep into the
build; verified in the running binary, not just the tag). Tests committed. 19 previously-exposed live
pages are now protected. Bug left open pending two owner decisions.

**Where we're going.** Two open items, neither blocking: (1) whether to build an explicit
"redesign this page" signal (bug 001's deferred fix step 4) or accept the "clear the layout, then
re-plan" route; (2) whether to run a live re-plan on a real site as gold-standard verification. Once
those are settled, 037 moves to `/bugs_closed/`. Closely related bug 050 (what an *empty* layout means
on a deployed page) remains with its owning session.
