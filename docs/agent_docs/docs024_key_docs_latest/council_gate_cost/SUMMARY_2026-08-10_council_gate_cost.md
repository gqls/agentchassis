# SUMMARY 2026-08-10 — council-gate prompt caching: built, reviewed, live, proven at 74%

## What we're trying to do

Keep the council gate — the panel of AI reviewers that checks every platform
code change before it ships — affordable enough that the owner can let it run
freely, without it eating the credits meant for the product itself.

## Where we've come from

The owner hit an Anthropic spending limit and asked to switch off the
improvement loop, believing that was the drain. It wasn't: that system was
already idle, its main scheduler off since May, its reviewing agents silent
for over a day. The real spender, once actually measured, was the council
gate — 85% of all AI spend across the whole platform, in a single day.

The cause turned out to be structural rather than wasteful in any obvious
way. The council has 17 reviewers, and each one genuinely needs the same
large body of evidence — the plan, the reasoning behind it, the database
schema. The waste was that this identical material was being sent, and paid
for, 17 separate times. A caching feature exists for exactly this shape of
problem, but nobody had ever built it, and the way the reviewers' prompts
were written would have stopped it working even if someone had — each
reviewer's own job description sat in front of the shared material, so no
two reviewers' prompts began the same way, and caching only works on a
shared beginning.

## What we've done

Built the caching mechanism itself, deliberately off by default so every
other part of the platform that shares the same code is unaffected unless
it explicitly opts in. Rewrote all 17 reviewers' prompts so the shared
material comes first. Started recording the figures needed to prove any of
this actually works, since the existing cost figures would otherwise have
silently understated real usage by about 95% once caching began.

Put the change through the council gate itself before shipping it, which
caught two real problems — one of which would have taken the review system
offline the first time it ran for real, not merely cost more than hoped.
Both were fixed the same day, and the tests were made to catch the specific
bugs the reviewers found, not just to pass again.

Once the owner's next platform build went out, verified the change directly
on the running system rather than trusting that the deploy had gone
smoothly — checked the actual running code on every copy of the service,
then watched one real review happen and confirmed, seat by seat, that the
first reviewer paid to write the shared material and every reviewer after
it paid a tenth of the price to read it back.

Also answered a direct question along the way: whether moving some
reviewers to a cheaper AI model would help further, and whether it would
make the reviews noticeably worse. It turned out the honest answer to the
first half reversed an earlier suggestion — mixing models actually costs
more than caching alone, because the saving from caching only applies
within one model, so using two models means paying to set up the cache
twice. The owner has kept everything on the higher-quality model.

## Where we are now

The saving is real, live, and measured at the artefact rather than
estimated: **74% off the council's own AI costs**, which is expected to
save somewhere in the region of $17 a day now, rising to around $24 a day
once a temporary launch discount on the AI model ends later this month.
Nothing about how the reviewers judge a change has altered — only how many
times the same evidence gets paid for.

A small follow-up piece — a standing check to make sure this saving can't
silently break again later, the way it broke the first time — was itself
sent through the council for review and came back asking for changes. The
reviewers found a real weakness in how that particular check was written,
even though what has already shipped is unaffected by it; that follow-up
piece is not yet finished.

## Where we're going

Finish the standing check, addressing exactly what the reviewers asked for,
and let this settle as a normal, unremarkable part of how the council runs
from now on. No further owner input is needed for that. Two smaller,
optional ideas were surfaced along the way and are recorded but not
started: an active alert if the caching ever silently stops working, and
extending the same caching approach to any other part of the platform that
develops a similar shape of repeated, shared context.
