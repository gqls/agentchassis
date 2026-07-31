# Summary, 31 July 2026 — we cut the ladder from eight stages to three, and why

*Written to be read aloud. This exists because the design changed materially: the owner
asked whether the ladder was actually worth it, and the honest answer was "yes, but much
less of it than I proposed". Previous: `SUMMARY_2026-07-30_the_lane_adopts_the_ladder.md`.*

---

## What we're trying to do

Stop building things in one leap. Give each small part of a build a claim written down
before the work starts, and a check that can genuinely fail, so that a more complicated
component or tool is more small steps rather than a bigger gamble.

## Where we've come from

Yesterday this lane adopted an eight-stage ladder — S0 through S7, one question and one
gate per stage — derived from a carousel that took five rounds and whose fifth round found
a bug present since the first. It was a careful piece of reasoning and it was mostly
scaffolding.

Two things then happened on the same day that we could not have arranged. Another team
picked up the idea from a document and **ran it forwards on a real tool**, which nobody
including me had ever done. And the piece of platform work that would let components join
the scheme went through the review council twice and came back approved.

## What we've done

**We put the ladder to the only test that matters: someone else used it.** They built a
vendor-trust checklist tool for leopardess, wrote the claim before the code, built a
harness with twelve checks and twelve deliberate breakages to prove the checks could fail,
then drove it in a real browser. It worked. More usefully, **it corrected me twice.** Two
of my eight gates were simply wrong — one of them described the wrong mechanism entirely,
and had they followed it, it would have *blocked* their build rather than helped it.

**Then you asked whether the whole thing was worth it, and the honest answer was: less of
it than I wrote.** Looking at what the evidence actually supported rather than what I'd
argued:

- Three parts earned their place. Writing the claim before building genuinely changes what
  gets built. Verifying by driving the thing the way a visitor would catches a class of
  fault nothing else does — two teams were fooled by the same illusion on the same day.
  And any check must be proven able to fail; that failed five times in one day, four of
  them mine, twice while writing the document warning about it.
- The rest did not. The mutation suite cost the other team about forty minutes and found
  nothing in their actual product. The first stage was "a five-minute look that prevented
  nothing". The last stage cannot even be finished, because a check it depends on is broken.
- **And the gates themselves were wrong twice out of eight.** That is the part that decided
  it. Eight gates is eight chances to be confidently wrong at someone else's expense, and
  we already have a measured example elsewhere of checkers multiplying until twenty-two
  agents are configured and only two actually run anything.

**So we cut it.** It stays as a written checklist that a builder reads and ignores where it
doesn't fit — which is exactly what the other team did, correctly, twice — and we build
machinery for only the two or three things that pay.

**And we found the thing that actually blocks progress, which is not the ladder at all.**
The other team measured that a chunk of our tools cannot be automatically tested *at all*,
because three names have to match exactly and often don't — and when they don't, the test
run doesn't fail, it quietly does nothing and reports a clean result. I re-measured it on
our own definitions rather than taking the number: **of 28 tool components, 10 have no page
the test system can find. Two of those have a written test specification that has therefore
never once run.** One of them is on fundamentallyai.com — `review-council-simulator` — where
the page and the tool are named slightly differently, so it has never been testable and
never will be until it's renamed. That rename is safe: the published web address comes from
a different field.

## Where we are now

The scaled-back version is the plan of record. The one-line database change that lets
components carry their own documentation was **approved by the review council** (second
round, eleven approve, three advisory objections, none serious) — and it is deliberately
**not live**: it needs the next routine deployment, and the safety check I wrote correctly
refuses to let it be applied before then. I did not force a deployment, because that would
have killed other teams' work in progress.

The three surviving objections all made the same point and all three were right: I had left
the definitive test as a note for a human to remember rather than as code. That is now code,
and it runs as a matched pair — it proves it can refuse before we trust it to approve.

## Where we're going

1. **The name-matching check first, ahead of everything** — including my own database work.
   It is one query, it needs nothing that isn't already live, and until it passes, every
   other check's green result is untrustworthy.
2. **Rename `review-council-simulator`** so our own site's tool becomes testable.
3. **Real-browser verification for components**, which is wiring rather than construction:
   the machinery exists and has been proven; it has simply never been pointed at components.
4. Then, and only then, revisit whether any of the discarded stages deserve to come back —
   with evidence, not argument.

## The honest caveat, kept from yesterday

This was my lane and my proposal, and I made the same mistake four times while building the
thing designed to prevent it. That is a reason to discount my enthusiasm rather than
evidence for the design. The strongest argument in favour is not anything I did: it is that
a different team read a document, executed it correctly first time, and sent back two
corrections and a rule I hadn't thought of. **That part transfers. My eight stages mostly
didn't, and cutting them is the finding — not a retreat from it.**
