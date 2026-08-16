# SUMMARY — 2026-08-16b — bugfix 285 (the section-list assembler is lock-blind): approved, half live, not yet proven

*(Second summary of the day, and the state genuinely moved: the morning's file was written with
the review pending and nothing running. Since then the council approved it and a fleet build put
half of it into production — which changes what is true and, more importantly, changes what is
still owed.)*

**What we're trying to do.** Stop the page-rebuild pipeline proposing the removal of sections a
human has deliberately locked onto a page. The list a rebuild works from was assembled from the
site plan alone, so a locked section the plan never knew about could not get into it; every
rebuild proposed dropping it, a last-line guard saved the row, and the page's own cached list went
on saying the section did not exist. The owner's acceptance is specific: a rebuild of
webdesign.uk's contact page keeps the locked chat box **in its proposed list**, the cache tells
the truth, the locked row is untouched, an unlocked neighbour still rebuilds, and no
"tried to remove" item is filed.

**Where we've come from.** The filing lane diagnosed it through two diagnosis rounds — the first
blamed the wrong function and was refuted, the second confirmed the list assembler — and filed the
case on 2026-08-15. This lane picked it up, confirmed nobody was mid-fix, re-verified the defect
live, and found it was not one page but a class: thirteen pages fleet-wide, five fresh
"tried to remove" items filed in one afternoon on the loan-calculator site, and locked calculators
being pushed to the bottom of their own pages by each pass. Two claims in the original write-up
turned out to be wrong on reading the code, and both are corrected in the case file.

**What we've done.** The list assembler now also reads the page's locked rows and slots them into
the list at the position they actually occupy, using the very same "is this locked?" rule the
end-of-pipeline guard uses, so the two cannot drift apart. The page's cached list is written once
with that full list, and only when it really changed — the old "only if changed" test turned out
to be one that could never be true, so every build had been rewriting the row. The health check
that compares plan against cache was taught the same rule, without which its first day would have
produced thirteen false alarms. All of it is unit-tested, and the important tests are
mutation-proven: remove the merge and one test fails on its own; remove the alignment fix and a
different one does. The council reviewed it twice — the first round asked us to *show* the one
claim the design leans on rather than assert it, which was fair and cheap to answer — and the
second round **approved** it, fourteen reviewers, no blocking objections. Three advisory points
came back; the substantive one is now answered by measurement, and the architectural one has been
written up as its own proposal rather than absorbed into a comment.

**Where we are now.** The lunchtime fleet build carries the first half of the fix, and we have
checked that at the running program itself with controls, not at the version number. The second
half — an extra durable record for the case where the lock lookup itself fails — waits for the
next build. **The honest headline is that live is not yet proven working:** every page rebuilt
since the build went out happened to have no locked sections, so the new code has run twenty
times and correctly done nothing twenty times. And the obvious-looking good news — no new
"tried to remove" complaints since the build — is worth nothing as evidence, because none of the
affected pages have rebuilt. Nothing has asked the question yet. We have written it down that way
everywhere rather than letting a zero pass for a result.

**Where we're going.** One test settles it: rebuild a page that actually has a locked section. The
recipe is written and ready. We have not fired it, because it republishes a live shopfront page
and that acceptance run belongs to the lane that owns the chat box. The alternative is patience —
the next time one of the twelve loan-calculator pages rebuilds of its own accord, it will exercise
the merge with nobody doing anything, and it should show the locked calculator back in the list
and file no complaint. Until one or the other happens, the case stays open, the chat-box lock
stays on, and the register entry says "half live, unexercised" in those words. After that: move
the case file to the closed set, and let the owner rule on the separate architectural question of
whether "the page's section list" should have one owner in the code instead of eight places that
each work it out again.
