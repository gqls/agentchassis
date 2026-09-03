# README — where we are (bugfix_361 render-check ratchet)

*Append-only, newest at the bottom. Plain prose for the owner.*

## 2026-09-03 — what this is, and what I changed

There is a job that runs every morning and asks one question of every component in the
library: *if this field were missing, would the page render a visible hole?* It has been
**failing every day since 9 August — twenty-five days**. Nobody had looked, and the job's own
design note predicted exactly that: *"a permanently-red job is a job everybody learns to
ignore."*

**The job was not broken.** It was telling the truth and being ignored, which is worse. The
fault was in how it decided what counts as *bad news*. It held a list of problems it already
knew about, taken once on 4 August, and flagged anything not on that list as new. Since then
the library has roughly doubled — 282 components to 497. Every component born after 4 August
is, by definition, not on the list, so everything it finds looks new. This morning it reported
**478 "new" problems**. Almost none of them were new; they were just *recent*.

**What I changed.** The job now asks a sharper question: *did this component get worse?* To
answer that it has to know which components it had actually looked at when it took the list —
and it never recorded that. It recorded the problems it *found*, not the components it
*checked*. So it now records both.

That distinction is not a detail, and it is where I disagreed with the bug report. The report
suggested a simpler fix: treat any component with no known problems as "too new to judge". But
the old list itself shows why that fails — it says it examined **139** components while naming
problems in only **115**. So **24 components were examined and found clean**. Under the simpler
fix, those 24 would be treated as unknown, and if one of them broke tomorrow the job would say
nothing. A component that was clean and then breaks is the single case this job exists to
catch. None of the 24 has broken yet, so no harm has been done — but it is the shape of hole
that stays quiet until the day it matters.

**The result.** The job now reports **18 real problems in 5 components**, plus **460 known-new
items in 62 components** listed separately so the debt stays visible rather than vanishing.

**It is still red, and I want to be straight about that.** It is red for eighteen things
somebody can actually look at, instead of 478 it invented. Making it *green* means either
fixing those five components or deliberately declaring the current state the new normal — and
that second one is a judgement about debt, not a code change, so I have not taken it. It is
yours. The five are `blog-listing_pre_037`, `social_proof`, `tool-ab-test-calculator_pre_037`,
`tool-equity-release_pre_037` and `tool-gas-unit-converter-gaswholesalers-com` — and at least
the first is a template that was rewritten rather than something that decayed.

**One thing I got wrong on the way.** I wrote a test to prove the fix, and the test could not
have failed. I only found out because I deliberately broke the code to check the test would
notice — and it did not. The reason is worth keeping: the test checked a value by writing it
and reading it straight back, and the reading step quietly cleaned up the exact mistake the
writing step was supposed to be caught making. It has been rewritten to inspect what was
actually written to disk, and it now fails when it should. Written up in the shared
wrong-calls log, because it is a mistake anyone can make and it looks like success.
