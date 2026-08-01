# SUMMARY — 2026-08-01 — the improvement loop no longer reports a busy site as clean

## What we're trying to do

Stop the improvement loop — the thing that sweeps a site, queues the fixes it finds, and
finishes with a re-render so those fixes reach the live pages — from ending with the message
**"No issues found — site is clean"** on sites where it had just queued dozens of problems.
The message was the visible part. The costly part was what came with it: the loop skipped its
own closing re-render and its own dispatch, and anything reading the run's result was told a
site with open findings was fine.

## Where we've come from

The case was filed on 2026-07-29 from a hand-fired run that queued 67 findings and then
declared the site clean. It sat unowned for two days, and it recorded its own central claim —
that this happens on *every* run — honestly, as an assumption: the database had already
deleted the run history that would have proved it.

The cause is not a broken step. It is three agents sharing one greedy step. The step that
promotes findings takes **everything** on the site, and the improvement loop runs its own copy
**last**, after two child agents have already run theirs. So its copy correctly reports "I
promoted nothing" — and the loop read that one sentence as "there is nothing to do". It
survived for months because a separate scheduled job picks the queued fixes up anyway every
two minutes, so the fixes happened and only the closing pass went missing.

## What we've done

**Reproduced it first, on a second site, before changing anything** — which turned the file's
assumption into an observation and, in passing, corrected it: the one escape route the file
thought might sometimes save the loop does fire, and does not help.

**Fixed it at the shared step rather than at the one caller.** The step now also reports the
*site's* state — is there work waiting here, whoever queued it — beside the old "did I
personally queue anything" signal. The old signal was deliberately left alone: three other
places in the platform read it correctly about their own work, and redefining it would have
fixed one branch by making a shared word mean two things.

**Put it through the review council**, which approved it with eight advisory comments. Four of
those were questions answerable by a database query rather than an argument, so we ran them;
all four came back in the change's favour, and one of them — *do all three agents file their
work under the same label?* — would have been fatal if it hadn't. Two more asked for things
we had only written down in prose to be given proper tickets, which they now have.

**Shipped it in the right order**, which was the part with the sharpest edge: the new field
comes from the program, the switch that reads it lives in the database and takes effect
instantly. Applied early, it would have sent *every* run down the "clean" path — worse than
the bug. So the switch was physically held back until the program was confirmed running, by
inspecting the live containers rather than trusting a version number.

## Where we are now

**Closed, live and proven.** The same site, the same command, one day apart, one thing
changed:

| | before | after |
|---|---|---|
| the loop's own copy queued | 0 | 0 |
| what it concluded | *"site is clean"* | *there are 42 items waiting* |
| closing re-render created | none | one, at the instant it decided |

The input was identical in both runs. That is what makes it a proof rather than a
demonstration.

Two things were deliberately left open and neither is unfinished business from this fix. There
is a **second, different route** to the same false "clean" message — a site that has been
audited three times is skipped entirely and told it is clean — which is filed as its own case;
no site is currently in that state, so it is a trap rather than a live fault. And there is a
**written proposal for you to decide**: should a step that sweeps up everything on a site have
exactly one owner, instead of three agents racing for it? This fix makes the race harmless; it
does not make it impossible.

## Where we're going

The proposal is the only thing needing a decision, and it is not urgent. Beyond that, the
useful next move belongs to a different case: the loop that now behaves correctly is still
**disabled**, so it only runs when someone fires it by hand. Everything above is worth more
the day that changes — and the second false-clean route becomes live on exactly the same day,
which is why it was filed rather than mentioned.
