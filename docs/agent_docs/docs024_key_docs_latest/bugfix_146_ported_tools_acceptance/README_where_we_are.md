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

## 2026-08-18 — the fix is written, tested, and waiting on two things

The code change went in today: when one of our automatic quality runs fails a ported tool,
the verdict now lands in the human review queue with everything needed to act on it —
before, it was written to a log-style note that nothing and nobody read. It is committed
and under council review, but it only takes effect when the next fleet build ships (today's
fresh build predates it).

Worth telling: the first version of my safety test had exactly the blind spot our own
handbook warns about — it "proved" the guard by asking the test database, which cannot see
a write it wasn't told to expect. I broke the guard deliberately to check, watched the test
wrongly pass, and rebuilt the proof on a channel that does see it (the code's own logging).
The broken-guard version now fails the test, which is the point.

Two things wait on you — set out under "Decisions" in my summary to you, and in the plan
file: whether ported tools should get an automatic baseline mobile-fit check (it costs
browser runs), and nothing else — the rest is machinery that rides the next release.

## 2026-08-18 (evening) — your ruling recorded, the fix is live, and the proof run is in flight

The new build carries the fix — I checked the actual running program on both server
copies, with a deliberately fake name as the control, and it's there.

Your ruling — rewrite every ported tool properly rather than bolt checks onto the old
copies — is recorded as the decision on the open question I'd left you. The measured
backlog is 72 pages across 6 sites; webdesign's rebuild programme already covers 56 of
them and is running. On loancash you were half right: the two newest tools were indeed
built by the framework (editable, mobile-checked, all as intended) — but three older
calculators and the tools index page came in as verbatim copies during the August 1st
adoption, before those two were built. They join the rewrite list.

And rather than wait days for the schedule, I fired the proof run now: a real acceptance
check against the vibe-equalizer page (one of the broken seven). If the fix works, the
failure lands in the review queue as an actionable item — the first ever from this route.
Result in the next update.
