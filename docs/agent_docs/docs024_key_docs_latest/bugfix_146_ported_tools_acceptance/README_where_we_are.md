# Where we are — ported tools and the safety checks that never look at them

Plain-prose log for the owner. Append-only, newest at the bottom.

## 2026-08-17 — picked up, and the first finding is that the bug is half-fixed by other work

I was asked to take on bugs 080 and 081, but both were finished and verified live weeks ago
(1 and 4 August). So I went down the open-bug list for the oldest one nobody is working on,
and that is bug 146: the tool pages we *ported* onto webdesign.co.uk (rather than built with
our own framework) are invisible to the automatic quality checks that our own generated tools
get. Seven of them were shipping with layouts that break on a phone screen, and nothing could
ever have noticed.

What's changed since the bug was filed in July: a new "tool health" checker now does look at
ported tools — but it only reads the code for known bad patterns (things like "no mobile
stylesheet section"). It caught 2 of the 7 broken pages. The other 5 are broken in a way you
can only see by actually rendering the page at phone size and measuring it — which is exactly
the check our generated tools get, and ported tools still never do. So the gap is real,
just narrower and better-defined than the bug file says.

Also relevant: you directed (15 August) that webdesign.co.uk's 63 ported tools be rebuilt
natively — once that finishes, those pages stop being ported at all. But other sites also carry
ported or adopted tools (gamesdesign, mortgagecalculator, and more), so the framework-level
question — "does a ported tool page join the same measured checks as a generated one?" — still
needs an answer that isn't "someone remembers to rebuild them".

Next: I'm mapping who runs which check and on what population, then I'll write a fix plan
biased toward the framework answer, and put it through the council before committing.
