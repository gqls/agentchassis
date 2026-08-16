# SUMMARY — 2026-08-16 — bugfix 285 (the section-list assembler is lock-blind)

**What we're trying to do.** Stop the page-rebuild pipeline proposing the removal of sections a
human has locked onto a page. The list of sections a rebuild works from was assembled from the
site plan alone; a locked section the plan didn't know about could never enter it, so every
rebuild tried to remove it, a last-line guard kept the row, and the page's own cached section
list said the section wasn't there. The owner's ruling: check and fix it. The acceptance: a
rebuild of webdesign.uk's contact page keeps the locked chat box IN its proposed list, the cache
tells the truth, the locked row is untouched, an unlocked sibling still rebuilds, and no
"tried to remove" item is filed.

**Where we've come from.** The filing lane (webdesign_uk_build_service) diagnosed it through
two 090 rounds — the first blamed the wrong function and was refuted; the second confirmed the
list assembler — and filed the case on 2026-08-15 with a recommended fix. This lane picked it up
the same evening, confirmed it was unowned-in-flight, re-verified it live, and measured that it
was not one page but a fleet class: thirteen pages, five fresh "tried to remove" items filed
that afternoon on loancalculator.co.uk, and locked calculators being pushed to the bottom of
their pages by each pass.

**What we've done.** Read the whole path end-to-end (loader → planner → link resolver → content
writer → compile → save guard), found two errors in the bug file's cross-references and recorded
them, designed the fix as a shared piece (one predicate, one merge, usable by both the loader
and the drift check), wrote it, tested it (twelve merge cases, the loader's first-ever tests,
mutation-proven, and the drift check), verified it against committed HEAD, submitted it to the
council, committed it (`7d9b7334a`), registered it (LOCK-008), and wrote the landmines and the
wrong-call it surfaced. Along the way found that the loader's "only write the cache if changed"
guard could never be true (a text-vs-jsonb comparison) and fixed that in the same statement.

**Where we are now.** Fix committed and council-submitted (verdict pending at time of writing,
`79f70435`); NOT live — it rides the next chassis roll. The bug stays open, the chat-box lock
stays on. The lane docs, RUNBOOK (incl. the exact post-roll acceptance recipe), and the bug file
carry everything a fresh session needs.

**Where we're going.** (1) Read the council verdict; if REVISE, revise and resubmit — the code is
already on the shared branch, so a REJECTED means reverting shipped behaviour and must be acted
on. (2) After the next roll: stamp check, one page-build-handler pass over contact via the
work-item recipe, the five criteria; then watch that the loancalculator "remove" items stop and
the drift check files nothing for the thirteen pages. (3) Then move the bug file to
`bugs_closed/`. (4) Open questions left for the owner: whether locked sections should be carried
through the writer untouched (silence, no render cost) — LOCK-008's review question; and whether
the calculators pushed to the tail should be moved back by hand or by a re-plan naming them.
