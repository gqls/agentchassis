# SUMMARY — brochure component library (2026-07-22): the story now flows

**What we're trying to do.** Stand up fundamentallyai.com — a new brand that
markets this platform's own real, verified capabilities as consultancy case
studies — and, separately, build genuinely best-in-class interactive brochure
components (hero carousels, hover-zoom cards, swipeable mobile carousels,
code-rendered stat bands) reusable through the framework. This summary is about
the first half: getting the built site unblocked so it can tell its story.

**Where we've come from.** The site was built overnight by the pipeline, on-brief
(navy/amber, line-illustration, pages named for our chosen pillars) — but not
live. Two blockers stood in the way: a content-validation gate holding five
pages, and no serving path yet. The gate's reason was thought lost to log
rotation; it wasn't — the checker persists its blocker detail to the database on
purpose, so the diagnosis was free: the cross-site contamination guard (built to
stop one site's copy leaking into another's) was firing on this site's
*deliberate, owner-approved* naming of leopardessconsulting.co.uk — the very
worked example of the self-correction story it exists to tell.

**What we've done.** Resolved that blocker, and verified it by reading the pages,
not by trusting a status flag. The correct fix — a per-site allowlist so a
portfolio brand may legitimately name our own sites — turned out to be **already
implemented and live in production** (the owner caught this; I had wastefully
re-derived and council-reviewed it, a miss now recorded in full in NOTES and
WRONG_CALLS). The one thing genuinely missing was the *data*: telling this
specific site it's allowed to name those domains. A guarded, backed-up seed did
that. The four core content pages were then re-queued and rebuilt: About,
Capabilities and Multi-Agent Review Council are **deployed with the
self-correction narrative present in their rendered content**, and the homepage
carries it too. Zero new blocks from the guard. The contamination bug
(`bugs_open/055`) is fixed, live, and closed.

**Where we are now.** The site's pages can finally tell the story they were built
to tell. The central blocker is gone. Three separate last-mile items remain, none
of them the thing we just fixed: the dedicated self-correction page (and two
others) came out of the build with **zero sections** — a planning-stage gap, not
the gate; two sections are correctly asking for real data (our own site list for
the homepage showcase; a business email); and nothing serves yet (the
Cloudflare→storage cut-over for the new domain, an owner/infra step). A deeper,
generic fault surfaced along the way and is filed as `bugs_open/056`: when a page
trips any validation blocker, regeneration can silently ship a version with the
flagged content simply dropped, recording nothing — that's what had erased the
story from the pages that "passed" before the seed. And the owner's observation
about the review council became `features_open/010`: an evaluator that weighs
whether an objection is substantive or a form nit, so a human needn't adjudicate
every "revise".

**Where we're going.** Populate the empty-section pages (the self-correction page
first — it's the story's dedicated home), feed the two real-data sections, and —
the owner's step — wire the serving path so the site reaches a browser; then read
the live pages to confirm. After that, the original second half of the brief still
stands: the from-scratch interactive components, which today's site does not yet
use.
