# VM estate — milestone summary, 25 July 2026

*The first summary of this workstream, written the day it opened. Plain
language, to be read aloud. Technical entry point:
`PLAN_2026-07-25_framework_controlled_vm_estate.md`.*

## What we're trying to do

The platform runs three rented servers — the relojistas watch portal's box, the
idea.uk box, and the tools-api "island" — and today every one of them is
configured by hand: a human runs a script, or copies files over. We want the
framework to own those machines the way it already owns web pages: their
configuration produced from what the database knows, applied automatically,
and checked for drift — and we want the three separate efforts merged into one,
rather than three lineages of scripts growing apart.

## Where we've come from

Each box was born inside a different project and inherited that project's
habits. The idea.uk box came first, with a good provisioning script. The
relojistas box copied that script and grew it — more features, and one new bug.
The island was built last and differently on purpose: it takes no incoming
traffic at all, reaches out to Cloudflare rather than being reached, and holds
no production credentials, because its job is to run experimental tools without
ever exposing the production cluster. The two script copies now agree on 61
lines and disagree on 614; the island's hand-copied files are a third lineage
starting now. The cost of this became concrete in July: a hand-edit to the
relojistas box's web config would have been silently destroyed by the next
script run, and had to be reconciled manually — exactly the kind of repair the
platform refuses to accept on web pages.

## What we've done

Read the relojistas provisioning script end to end and recorded an honest
walkthrough: it is genuinely good work — safe to re-run, guarded against
locking you out, clever about the certificate chicken-and-egg — with two flaws.
One was a real bug that would abort the script exactly when adding a new domain,
which is fixed; notably it exists only in the copy, not the original, which is
the case for merging in miniature. The other — every site on the box inherits
the watch forum's special addresses — we deliberately left, because the cure is
the new design, not another patch.

Confirmed that almost nothing needs inventing: the platform already provisions
machines, runs commands on them, and monitors them for GPU training; and the
script's own header has promised for months that the framework would take over
— a promise we checked and found nothing behind.

Designed the merge and put the one hard question to the owner, who has ruled:
**merge the generator, not the trust boundary.** One description of what each
box is, one renderer producing every box's configuration, one drift check — but
the public boxes are pushed to over SSH, while the island *pulls* its own
configuration outbound, the same direction it already dials, keeping the
isolation it was built for.

## Where we are now

Nothing is built, and that is deliberate — this is the design milestone. What
exists: the plan with the full walkthrough, the three boxes documented as they
actually are, one bug fixed in the current script, and the owner's decision on
the island recorded as a constraint. The three merges are sequenced: relojistas
first (best understood; prove the renderer by reproducing the live box byte for
byte before touching anything), idea.uk second (no new machinery — a second row
through the same renderer, which is where the fork dies), the island third (same
renderer, pull delivery — ideally before its engine lands, so the engine arrives
onto an already-managed box). The owner's pending relojistas server session is
unaffected by any of this and remains the only gating item on that project.

## Where we're going

First, describe the relojistas box as data and prove the renderer can reproduce
its live configuration exactly — a step that costs nothing and touches nothing.
Then apply through the framework on the public boxes; then the island's pull
path. Still open for the owner: whether to consolidate hosting providers,
whether ordering new machines is in scope or only configuring existing ones,
and where the rendered artefacts should live.

## The one-sentence version

Three hand-run servers from three projects are becoming one framework-managed
estate — configuration rendered from the database, drift detectable, the
island keeping its isolation by pulling rather than being pushed to — with the
design settled, one bug already dead, and the first step a proof that costs
nothing.
