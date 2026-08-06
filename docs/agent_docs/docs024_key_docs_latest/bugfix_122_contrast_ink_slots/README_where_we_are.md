# Where we are — the contrast bug (122)

Plain-prose log for the owner. Append below; never rewrite or reorder.

---

## 2026-08-06 — picked up the contrast bug, and most of it turned out to be already fixed

I went looking for the next open bug nobody else is working on. The contrast one
(number 122) came out coldest by a wide margin — it was filed on 27 July after you
reported a link on oufe.com as "dark blue on the black background and not easily
readable", and it has essentially been sitting since. The one team that had touched it
recently, the darts site lane, had written in the file itself that the main fix "still
belongs to whoever takes it", so it was genuinely free.

**Before planning anything I re-measured the whole fleet in a real browser**, because
the file's numbers are nine days old and this codebase takes something like fifteen
hundred commits a week. That was worth doing, because the file is substantially out of
date in our favour.

Two of its three named problems are **fixed**. The big one was that every site's
header button had white text hardcoded into it, so whether it was readable was pure
luck depending on how dark that site's accent colour happened to be — five of six
sites failed. That is now done properly: the button takes its text colour from the
site's own palette, and I checked all nineteen stored headers on the fleet — not one
still has white hardcoded. Separately, robot-hands.com's main "Run a MatchMatrix
Query" button used to be white text on a white button, i.e. genuinely invisible. That
has gone too.

The one problem from the original file that is still live is the vonc.com Gauntlet
buttons — purple text on a pink button. That surface belongs to the Gauntlet team and
the file already says to coordinate rather than reach into it, so I have left it
alone.

**What I did find is a different problem, and it is on more sites than the original.**
Twelve of the fifteen homepages I measured still fail, 109 failures in total. They
fall into three groups:

The first is a colour being asked to do two jobs at once. Each site's palette has a
"primary" colour. On the dark sites that primary is a near-black — which is perfect
when it is used as a *button background* with light text on top, and invisible when it
is used as *text* on the page. Seventeen of our eighteen layouts use it as text
somewhere. So the same value is simultaneously right and wrong, and repointing the
palette just moves the failure from one place to the other — the darts lane worked
that arithmetic out in July and correctly decided to change nothing. The real gap is
that we have never given components a "primary, but guaranteed readable" colour to
reach for. We already compute exactly that answer in two places in the code; we just
never offer it in that direction.

The second is the worst single thing I found and **it is in no bug file at all**:
ai-agent-orchestration.com is serving six headings in exactly the same colour as the
background behind them. Not low contrast — identical. That page has been publishing
six invisible headings and nothing noticed. I have deliberately **not** guessed at why,
because on paper it should be impossible, and a confident guess about shared code is
the specific mistake our own guidelines were rewritten to prevent. It is going to the
automated diagnosis loop.

The third is components with white text painted onto a mid-tone coloured button —
finetuning.uk, gaswholesalers.com and gamesdesign.co.uk between them. The frustrating
part: we *built* the fix for this on 27 July. The platform now works out a readable
text colour for exactly this situation and publishes it. **Nothing uses it.** Not one
template, layout, or page anywhere on the fleet references it. So we did the hard part
and never connected it.

There is a decision worth flagging to you. The fix I want to make adds one new
"guaranteed readable" colour variable to every site's stylesheet, computed by the
renderer. Nothing changes until a component chooses to use it, so it cannot break
anything on its own — but it is a shared mechanism, so it goes through the review
council before it ships, and gets registered so the next person can find it.

**Also, good news from the build you just deployed.** The tool that checks live pages
in a browser and files the failures as work items — the thing this bug asked for back
in July — is now fully built and running in the new image. I confirmed it is in the
binary and that the agent is correctly configured end to end. The only thing missing
is that **nothing ever tells it to run**. It has been run by hand exactly once, on
4 August, on one site. So what was described in July as "build the checker" is now a
single scheduling row. That is the cheapest useful thing in this whole file, and it is
the same pattern we keep hitting: something built correctly, then parked behind a
switch nobody turned on.
