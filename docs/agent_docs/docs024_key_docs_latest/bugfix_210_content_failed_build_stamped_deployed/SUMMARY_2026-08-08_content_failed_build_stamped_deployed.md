# SUMMARY — bugfix 210, 2026-08-08 (milestone: fix built and committed)

**What we're trying to do.** Stop a silent lie in the page-build pipeline: when a page rebuild's
content generation fails, the system was marking the page "deployed" anyway, because its checks
looked at the previous build's leftovers. The rebuild request vanished — no retry, no error, a
stale page presented as current.

**Where we've come from.** The bug was found by the thread that fixed the related owned-pages
bug (208); they fixed their narrow case and deliberately filed this general one separately,
because fixing it naively would create the opposite problem — a page that fails deterministically
being rebuilt (and paid for, in LLM calls) forever.

**What we've done.** Built the fix with the retry bound the bug file demanded: the false stamp is
refused for any skipped assembly; the page goes honestly back to "needs rebuilding"; every
refusal is written to the error log (so, for the first time, we can count how often this
happens); after three failures the page is parked behind a visible "needs a human" ticket that
also blocks all automatic rebuild attempts; a successful rebuild clears the ticket by itself.
Proven by tests that were each shown to fail when their guard was sabotaged. Reviewed structure:
the change was submitted to the review council before implementation, and the commit carries the
submission trailer. Registered in the concept register (PBP-038), with the two traps a future
session could step on written into LANDMINES, and the two affected teams' notes updated.

**Where we are now.** Committed on the shared tree, so it ships with the next chassis build. The
code is inert until that roll. Council verdict pending at the time of writing.

**Where we're going.** After the roll: verify on the pods (the new error-code string must appear
on every replica), then watch the error log — the first refusals it records are also the first
real measurement of how often this bug ever fired, which nobody has been able to answer.
