# SUMMARY — 2026-08-22, component identity (`bugs_open/357`)

## What we're trying to do

Stop the platform from recording a page block as one thing while it stores something entirely
different. On 22 live pages a block declares itself the shared **hero** banner — a title, a sentence
and a button — and actually holds a complete interactive calculator or tool, ten to twenty-two
thousand characters of it. Everything downstream that reasons about that page reasons about a banner
that is not there.

## Where we've come from

The session began on `bugs_open/277`, which turned out to have been closed hours earlier on the
owner's ruling. That close was verified against the live system rather than taken on trust, and it
holds. Its own closing note named the successor and the lane that filed it had finished, so `357`
was unowned and was taken up here.

## What we've done

Established the cause, in committed code, with the evidence cited. A tool page arrives as one
`<div class="tool-page">` lump with no internal markers, so the extractor gives up and labels it with
a placeholder meaning *identity unknown*. The next step fills that placeholder in **by position** —
first block, plan says first block is a banner, therefore banner — and never looks at the bytes. Hero
is planned first on all 22 pages, so every tool became a hero.

Corrected the bug file on three counts: the population is **22, not the 9 filed** (357's own query
says so, and the newest was born the same day the bug was filed, on a site homepage); the **writer is
settled** by row fingerprint, which is the gap the earlier diagnosis run returned UNVERIFIABLE on;
and the class is **self-renewing**, because the mechanism that protects these tools from being
blanked is the same one that re-attaches them to the wrong identity on every rebuild.

Measured a discriminator for a guard rather than guessing one: comparing a component's own
`data-component` self-declaration against the stored HTML gives 1,550 agreements, zero
disagreements, and 24 genuine defects — against 158 false-positive-laden hits from the obvious
alternative.

Took the plan to the review council twice. Both rounds came back REVISE and both were right.

## Where we are now

**Nothing has been changed** — no code, no data, no live site. Two candidate repairs are blocked, and
the reason is a mechanism nobody had written down: the rebuild path matches stored blocks to new ones
**by slot name alone**, so correcting a block's label makes the next rebuild miss the match, re-append
the tool and keep the freshly generated banner beside it. A repair described as byte-preserving would
have changed what four live sites serve.

The council's second round rejected the cut-down "record only" version on the ground that it stops
nothing — the estate's own documented "detected but never blocked" pattern. Accepted.

Also settled a question previously marked unmeasurable, in the reassuring direction: **no tool has
been destroyed by this**. Searched the page history archive with a control — 182 slots still
interactive, 17 changed, and fifteen of those grew. None is in this population.

## Where we're going

The council's architecture seat named the real answer and it is now **`RFC_046`**: a component row's
identity is **inferred five different ways and stamped none**, and a sixth inference is not the fix —
the proof being that the safe plan and the effective plan could not be the same plan. The RFC asks
for a decision between stamping identity at the point of production, consolidating the five
inferences into one, or knowingly accepting the class.

Until that decision, the 22 pages stay as they are. They serve their tools correctly today; only the
label is wrong, and every route that would correct the label is currently more dangerous than the
defect.
