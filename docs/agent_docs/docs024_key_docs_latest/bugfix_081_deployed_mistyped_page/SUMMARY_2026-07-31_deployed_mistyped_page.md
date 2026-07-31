# SUMMARY — 2026-07-31 — bug 081, the page nothing could repair

## What we're trying to do

Stop the platform from quietly damaging a live page while trying to fix a
different one — and stop it retrying that same non-repair for ever.

## Where we've come from

When a check notices a site is missing a page, a model plans one and a piece of
code creates it. That code was written to create. But its database instruction
said "…and if a page of that name already exists, update it instead" — and the
update only covered *half* the row. It replaced the page's title and content and
left the page's **type** alone.

The type is what tells the rest of the platform what a page is for. So after this
ran you had a live public page carrying content written for a completely
different job, still labelled as the old job. And because the label never
changed, the check that noticed the missing page noticed it again on the next
sweep. On `ai-agent-orchestration.com` that has been going round since **1 May** —
three months, one work item exhausted to `unresolved`, another still waiting.

The obvious repair — "also update the type" — is how you break a working website.
The session that filed this bug in July had already found out why, and stopped:
to fix a mistyped page automatically you must know which page is *supposed* to
hold the role, and on `robot-hands.com` the real news page and the gripper-catalog
page are indistinguishable in the database. Both carry exactly the same single
section. Guessing means a coin-flip chance of relabelling a working page.

## What we've done

Re-checked the bug before touching it, in both the code and the live database —
it was five days old on a tree about thirty sessions are editing. Everything was
still exactly as filed, and the loop was still running.

Then re-ran the discriminator query the previous session used, and it came back
**worse**: five pages of that identical shape now, not four, and the new one is
archived. So two of five are false positives. That settled the design — the
platform must not guess.

**The fix removes authority rather than adding it.** The page-creating code now
only ever creates. If the name is taken by a page that is **live** and has a
**different** type, it changes nothing at all: it files a decision for a human,
carrying the exact one-line SQL that would resolve it, and marks the original job
**blocked** rather than complete — because calling it complete when nothing was
repaired is the false green that let this run for three months unnoticed.

The key move is that we never have to answer the hard question. The planner has
already chosen the page name. We don't need to work out which page *should* hold
the role; we only need to notice the name is taken by a page doing a different
job, and ask. That is why this was buildable today when the previously-proposed
fix is still blocked.

We also cut half of our own fix. A draft quietly re-typed pages that had never
been published, where relabelling is harmless. Then we counted: **every** mistyped
page in the fleet — all five — is published; there are none in the unpublished
state. That half would have repaired nothing while granting an automated
component broad authority to relabel pages. Deleted, with a test left behind that
fails if anyone adds it back without re-running the count.

Four tests, and only one of them is the new behaviour: a guard proved only on the
case where it fires is satisfied by deleting the guard and refusing everything, so
three of the four are controls that must still take the old path.

## Where we are now

Committed, reviewed-pending, and **not live**. Go code does nothing until the
chassis image is rebuilt and rolled, so production is unchanged; the commit means
the next build carries it. The bug file has moved to `bugs_closed/` with that gap
stated in a banner rather than glossed, following the precedent set by
`bugs_closed/167` a few hours earlier.

The council gate is mid-review (`ccd4384c`), and the commit carries
`Council-Submitted:` rather than `Council-Reviewed:` — a trailer that asserts
nothing, and gets credited automatically when the verdict lands.

Two live pages are still mislabelled and were deliberately not touched. Fixing
them changes what they serve to the public immediately, which is an owner's call,
not ours.

## Where we're going

Three things, in order.

**Read the verdict and act on it.** The code is already on the shared branch, so
a REVISE is a follow-up commit, not a hold.

**After the next roll, prove it in production.** Both branches, induced —
scripted in the runbook. The one trap worth naming: snapshot the page's title and
content *before* inducing, because the whole claim is that they do not change,
and without a before-value "unchanged" and "changed back" look identical.

**Then the owner's decision on the two live pages.** For
`ai-agent-orchestration.com` relabelling is almost certainly right — it closes a
three-month loop. For `idea.uk` it is not sufficient on its own: that page's
content is separately stale, so a relabel alone would leave it wrong in a new way.
