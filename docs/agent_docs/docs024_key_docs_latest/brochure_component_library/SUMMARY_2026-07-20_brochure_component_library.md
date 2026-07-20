# SUMMARY — brochure component library (2026-07-20)

**What we're trying to do.** Build a set of genuinely good-looking, interactive
brochure-site components — the kind Bain, BCG and McKinsey use: auto-advancing
hero carousels, images that gently zoom instead of playing a video, swipeable
mobile carousels, varied page designs — and make them reusable through the
existing site-generation framework rather than one-off HTML. The proving ground
is a new brand, fundamentallyai.com, which the owner will also use to market
this platform's own real capabilities as consultancy-style service lines and
case studies, in the same register as those reference sites.

**Where we've come from.** The platform already runs one consultancy brochure
site, leopardessconsulting.co.uk, which itself was found last week to have
shipped fabricated content — an invented founder, an invented "70+ agents in 8
departments" structure, invented client case studies, a fake uptime figure. It
has since been cleaned up and is now verified true. That history is directly
relevant here: fundamentallyai.com is being built under the same
claims-verification discipline from day one, so nothing it publishes needs a
rescue later.

**What we've done.** Two parallel research passes. Externally: a design-pattern
study of Bain, BCG, McKinsey and peers (Accenture, KPMG, EY), producing concrete,
implementable recipes for the specific effects the owner described — the
hover-zoom image (a plain CSS scale transform, no JavaScript, with the
accessibility handling it needs), and the swipeable carousel (CSS scroll-snap
alone, again no JavaScript for the core interaction) — plus the real
accessibility rules an auto-advancing hero carousel is expected to follow, which
are not optional extras but standard practice at this level. Internally: a full
map of how this platform actually turns a site brief into a live page, which
confirmed there is no such carousel or hover-zoom component anywhere in the
system today (a genuine new build, not something merely unused), identified
exactly where a new component type has to be registered for the planner to
actually choose it, and found one already-known trap (a component's JavaScript
silently never reaching the live site) that turns out not to apply to the
components we're about to build. Separately, a thorough pass through the
platform's own documentation and code turned up a grounded, evidence-checked
inventory of what this platform genuinely does well: a real fine-tuning cycle
with an honestly-reported result, a live 13-seat AI review council with a real
decision record including catching a genuine bug another team missed, a real
Stripe payment integration, and a clean case study of reviving a dead domain's
abandoned subscriber feed. It also found that the owner's idea of letting a
partner safely search their own data via embeddings is built on real, working
technology — but the safety part, keeping one client's data invisible to
another, doesn't exist yet on our own shared system, so that offer needs to be
pitched as something we'd build properly for a client, not something already
proven safe today.

**Where we are now.** The owner has already resolved two of the open questions:
fundamentallyai.com is confirmed owned (hosting to follow shortly), and people
in the new site's imagery will be drawn as line illustrations rather than
photographed, which neatly avoids ever appearing to show a real person who
doesn't exist. What remains is a small number of genuinely owner-level calls —
how to word the private-search offer honestly, whether to tell the story of the
platform catching its own past mistake as a selling point, and whether to name
the sibling leopardess site directly — plus the ordinary engineering work of
building the first new component and registering it properly.

**Where we're going.** Once the owner has weighed in on those calls, the plan is
to build one component end to end through the real pipeline first — most likely
the hero card carousel, since it exercises the carousel behaviour, the
hover-zoom styling and the imagery question all at once — prove it on
fundamentallyai.com once its hosting is live, then extend to the remaining
component types. The actual page copy will always be written by the platform's
own content-writer from a proper brief, never typed up by hand in a chat, so the
next real step after the owner's steer is turning this research into that
brief and firing the framework for real.
