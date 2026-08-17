# SUMMARY — 2026-08-17 — bugfix 285 (the section-list assembler is lock-blind): closed, proven twice, and the rule made mechanical

*(Third and final summary. The 08-16 pair were written with the review pending and then with the
fix live but never exercised; this one records the two things that could not be said before — it
has been proven on the pages that matter, and the convention it rests on is now a build failure
rather than a comment.)*

**What we were trying to do.** Stop the rebuild pipeline proposing the removal of sections a
human had locked onto a page. The list a rebuild works from was assembled from the site plan
alone, so a locked section the plan had never heard of could not get into it: every rebuild
proposed dropping it, a last-line guard saved the row, and the page's own cached list went on
saying the section did not exist. The owner's acceptance was specific — a rebuild of webdesign.uk's
contact page had to keep the chat box **in its proposed list**, not merely fail to delete it.

**Where we came from.** Filed by the webdesign lane on 15 August after two diagnosis rounds (the
first blamed the wrong function and was refuted). This lane picked it up, confirmed nobody was
mid-fix, and found it was not one page but a class: thirteen pages, five "tried to remove"
complaints in a single afternoon on the loan-calculator site, locked calculators being pushed to
the bottom of their own pages. Two claims in the original write-up turned out to be wrong on
reading the code, and both were corrected in the case file rather than quietly dropped.

**What we did.** The assembler now merges the page's locked live rows into the list it builds,
using the very same "is this locked?" rule the end-of-pipeline guard uses — one predicate, not two
that can drift — and the health check that compares plan against cache was taught the same rule,
without which its first day would have raised thirteen false alarms. The cache write was fixed in
passing: its "only if changed" test could never be true, so every build had been rewriting the
row. Two council rounds approved it. Then, on the architecture seat's signal and the owner's
ruling, the convention was made mechanical: a test that fails the build when a new reader of the
plan's section list neither honours locks nor says in writing why it must not.

**Where we are now.** Closed. The fix is live and has been proven twice, both times on rebuilds
that happened by themselves rather than runs we fired: the loan-calculator front page on 16
August, and the owner's own contact page on 17 August — proposed list carrying the chat box, cache
telling the truth, the locked row untouched, its neighbours rebuilt normally, no complaint filed,
and the chat box visible on the live page. The enforcement test is committed and mutation-proven,
and it earned itself before it shipped by failing on a reader our own hand census had missed. Two
questions the case raised are settled: the architecture question is ruled and built, and the
"should we skip regenerating a locked section" question is ruled *don't* — measured, because those
pages rebuild four times a month between them.

**Where we're going.** Nothing is owed on this case. Three things are left as pointers rather than
work: a dormant rebuild path that plans from the cache without going through the fixed loader,
which would reopen the class if it ever woke; forty-one stale "tried to overwrite" notes sitting
in a human queue, which are hygiene rather than a fault; and the standing trigger to revisit the
skip-regeneration decision if a locked, AI-written section ever lands on a frequently rebuilt
page. The lane's more transferable output may be the three near-misses it logged, which are one
mistake wearing three coats: a check that answers the question you *encoded* rather than the one
you *asked* — a grep for an identifier the page never contained, a fetch of a domain that
redirected somewhere else entirely, and a census that gained a member while its total was left
where it was. Each was caught by a control, and each is now a recipe in the runbook.
