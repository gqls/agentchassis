# SUMMARY — webdesign.co.uk: live, corrected, and ready to hand over

**2026-07-27.** Second summary in the series (the first,
`SUMMARY_2026-07-26_site_live.md`, marked the site going live). Written because
the read-out genuinely changed: the hero is fixed, and two of the things the
previous summary asserted turned out to be wrong in ways that alter what happens
next.

## What we're trying to do

Merge the owner's two hand-built static sites — `website-design.com` (55 tools,
23 articles, austere Swiss styling) and `websitedesign.com` (10 tools, 10 guides,
a warm sage-and-terracotta homepage that none of its own sub-pages ever received)
— into one chassis-managed home at `webdesign.co.uk`, in the warm design,
carrying every feature except the client-side LLM builder the owner asked to skip.

## Where we've come from

The previous summary recorded the site going live: 98 pages, the design pinned
and holding, three source defects fixed on the way through, and two of our own
mistakes caught late. Since then the work has been almost entirely **correction**
— which is why this entry exists rather than being folded into the last one.

## What we've done

**The hero is fixed.** The planner had chosen a full-bleed banner painting a
dark overlay across a photograph — the opposite of the brief, and forbidden by
the site's own design rules. It is now the two-column layout the brief asked for:
copy left, image right, both buttons live, stacking on a phone. Verified on the
live page, with zero dark overlay inside the hero.

Getting it there exposed something worth knowing about the platform: **no
re-render path handles a section whose *component* has changed.** Assemble-only
republishes the stored HTML (it cheerfully re-published the dark hero), and the
data-refresh path correctly did nothing because no data had changed. The section
had to be re-rendered from its template directly.

**JavaScript loss is now impossible in this pipeline.** The converter refuses to
build if any page loses a script, and that gate was proved by deliberately
reintroducing the original bug — 60 failures, one per tool that would have
shipped dead.

**Two of our own claims were withdrawn.**

The first: we reported the final publishing step as blocked by a platform bug.
It was not. The queue takes about twenty minutes to pick up the first job and
runs at roughly two minutes a page — three and a half hours for ninety-eight —
and we gave up eight minutes before the first one started. The supporting
evidence, that other work was completing concurrently, was an artefact: that work
had been claimed *before* the page jobs existed and stopped dead the moment they
began, because we had just told the queue to prioritise pages. We mistook the
queue obeying our own instruction for the queue being broken.

The second, and the more expensive kind: we filed a bug asserting that **nothing
in the platform checks whether a page's JavaScript works.** That is false. There
is a four-tier verification ladder whose top tier drives the deployed page in
real headless Chromium — genuine clicks, post-interaction assertions, console
error capture, desktop and mobile — live in production, and documented under a
heading reading *"Does it actually work in a browser?"*. The owner found it with
one question. The bug is rewritten and renamed: the real finding is a **coverage
boundary**, not an absent capability — that tier only ever visits pages marked as
`tool`-level components carrying declared acceptance criteria, so none of this
site's 97 pages is ever browser-tested.

That distinction matters more than the embarrassment: a coverage gap and a
missing capability read identically in a bug file and demand completely different
work. The first version would have sent someone to build a headless browser tier
that has been running for weeks.

**Seven missteps are now filed** in the fleet-wide `WRONG_CALLS.md`, in that
file's own row shape, with its standing tally updated — including two of ours
from the previous day that had been appended in the wrong format and were
therefore invisible to the aggregate, which the file states plainly is the whole
point of keeping it.

## Where we are now

Live and correct. Every page returns 200, the hero matches the brief, chrome and
search work, the tools have their JavaScript, and every colour on the site is one
of the owner's — verified in the published stylesheet, not the configuration
meant to produce it.

The platform's own immune system has begun sweeping the site, which is the
expected end state, and has raised its first items. One is a stale re-render job
of ours that failed three times on a malformed spec and can be cancelled. Twelve
are `undeployed_asset` warnings claiming the generated images were never
published — but all twelve images return 200, so that is either a detector false
positive or a database flag that was never set. Nobody has established which yet,
and it should not be guessed at.

## Where we're going

The remaining work is small, and one item has become considerably more
interesting than it was.

About sixteen tools — the ones using canvas, clipboard, file upload or browser
storage — still need clicking through by a human. **But the discovery above
changes the shape of that job.** Rather than sixteen manual browser sessions, the
better move is to widen the existing browser-verification tier by one predicate
so it covers owned pages. The machinery is built, live and proven; it simply
never looks at pages like ours. That would convert a one-off chore into a
recurring automated check, and cover the rest of the fleet's owned pages at the
same time.

Beyond that: establish what the twelve asset warnings actually mean; implement
the port tool's `verify` command, whose checks currently exist only as recipes in
the runbook; and confirm the deploy robot's permission to clear the CDN cache,
whose only failure symptom is changes appearing slowly.

One question remains open for the owner. The two source sites are untouched and
still live, so the same content now sits on three domains. That needs a decision
— redirect, canonical tags, or leave them — and it is his call, not ours.
