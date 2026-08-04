# Where we are — the page saver that quietly throws away the content it just wrote

Plain prose, append-only, newest at the bottom. Owner's document.

---

## 2026-08-04 evening — what this is about

When the platform builds a page, it writes each section to the database twice over, in two
different forms. One is the finished HTML — what a visitor sees. The other is the structured
content behind it: the headline, the paragraphs, the button labels, as data rather than
markup. The second one is the only thing the platform can *re-render* from later. If you want
to refresh a page cheaply — new prices, a new image, a corrected link — the platform reads the
structured half, re-renders it, and puts fresh HTML back. That is the difference between a
few seconds of work and a full, expensive rewrite by the language model.

The bug is that the code doing the saving does not know where to find the structured half by
itself. It is told, per caller, in configuration. Six different parts of the platform call it.
Four of them were never told. So those four save the HTML perfectly, throw the structured half
away, and report success. Nothing complains. The page looks right, serves fine, and has
quietly lost its ability to be updated cheaply.

That was found yesterday morning by another thread's own test run — it rebuilt a page
correctly and stripped the structured content off all three of its sections in the process.
That thread fixed one of the four callers and filed the rest as bug 194. I have picked it up.

## What I have established today, before changing anything

**The bug is still real.** I re-read the live configuration rather than trusting the bug file:
the fixed caller carries the setting, and three still do not.

**One of the three is not actually broken.** The bug file flagged it as unmeasured, and it was
right to. `tool-recreation-handler` rebuilds a whole interactive tool as one lump of HTML —
it has no sections and no structured content to keep. A NULL there is the correct answer, not
a loss. Copying the setting onto it, as the other two want, would have been a guess dressed up
as a fix.

**The other two are dormant.** Neither has run in the nine days our durable run-counter
covers. That is good news for risk and bad news for proof: I cannot demonstrate a fix on live
traffic through a path that never runs, so whatever I write has to be provable offline, by
tests that can genuinely fail.

**The damage is not hypothetical, and it has a price tag.** When a page with missing
structured content comes up for a cheap refresh, the platform correctly refuses to render it
(rendering from nothing would blank the page) and escalates it to a full rebuild instead.
There are **44 such escalations across 8 sites** since 12 July — 21 completed, **13 failed
outright on 3 August**, 4 sent for human review. Each one is a cheap job turned into an
expensive one. I am not claiming all 44 were caused by these four callers; some pages simply
predate the structured-content era. What I am claiming is that this is what a NULL costs,
whoever wrote it.

## The choice in front of me

The obvious fix is to add the missing line of configuration to the two callers that need it.
It is one key each, it goes live immediately with no software release, and it closes the bug
as filed.

The better fix, I think, is to stop the saver depending on being told at all — let it look for
the structured content in the places it is always kept, so that no caller, now or in future,
can get this wrong. That is exactly the shape the thread next door chose yesterday for the
sister bug, and this same file already carries a comment arguing the same thing for a
different defect: persistence is the one place every page-writing path flows through, so a
guarantee made there holds whatever the configuration says.

I have put the decision to a planning pass and will record what it comes back with, and what I
decide, below.

## 2026-08-04, later that evening — what I decided, and what is now true

I did both halves, in this order.

**First, the two callers that were simply missing the line got it** (a database config
change, live the moment it applied, no software release needed). That closes the bug as it
was written. Before applying it I deliberately broke the seed's own safety check to watch
it fail — twice, once for each caller — because a check you have never seen fail is not a
check, it is a hope. The second one mattered: the two checks run in sequence, so the first
failure would have stopped the script before the second was ever tested. It would have
looked proven and been half-proven.

**Then I fixed the thing underneath**, which is the part I actually care about. The saver
now looks for the structured content itself, in the one place the rest of the platform
already agrees it lives — and it borrows that address from the code that validates pages,
rather than keeping its own copy, because two copies of one address is precisely the
disease here. If a caller genuinely has no structured content, it can now say so in its own
configuration instead of leaving a silence that looks identical to a mistake. And a caller
that wants a missing one to be a hard failure rather than a quiet loss can ask for that,
though I have switched that on for nobody: adding a new way for the busiest path in the
fleet to fail, on a hunch, is how safety checks get deleted in a fortnight. Finally, when a
page that had structured content gets saved without any, the platform now writes that down.
The whole reason this ran for six months is that it wrote nothing down.

**What is live right now:** the configuration fix, for both callers. **What is not:** the
software half, which sits in the code but does nothing until the next release build — as is
normal here.

**One honest caveat about proof.** The two callers I fixed are dormant; neither has run in
the nine days our records cover. So I cannot show you a live run proving it. What I can
show is that I broke the new code four different ways on purpose and each break was caught
by the test written for it, and that the whole area still passes against a clean copy of
the shared code. When the next release goes out, one of the two can be triggered by hand,
and I have written down in advance what would count as failure — including one trap: the
structured content can also survive by a different route, so simply seeing it present after
the run would be a false pass. The run has to say *which* route it took, and it now does.

I have put the change to the reviewer council and am waiting on its verdict before closing
the ticket.
