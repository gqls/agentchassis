# SUMMARY — per-site AI section / operation (2026-07-21)

*First milestone read-out for this workstream. Plain prose, written to be read
aloud. Five parts, in order. Framework capabilities are not repeated here — see
`CAPABILITIES_framework_inventory_2026-07-21.md` in this directory.*

## What we're trying to do
Give most of the sites we build a genuinely useful AI offering, reached from a nav
link, that helps that vertical's audience get ahead in a fast-moving AI world —
and that quietly advertises our own framework as something you can buy, rent or
subscribe to. It started as "an AI-insights section plus a vertical chatbot plus
honest articles about AI and jobs," and over the conversation it turned into
something sharper: **each site gets one signature AI *operation* — an interactive
tool that produces a real, original, site-specific artifact** — with the honest
editorial demoted to the supporting layer around it.

## Where we've come from
The original brief bundled three things together: a state-of-the-art vertical
chatbot, a section of brutally honest job-loss-and-opportunity articles, and a
paywall after a couple of free tries — all wired to the best specialist models,
ideally on a separate cluster. The first thing we did was pull those apart and
challenge the load-bearing assumption, which was that we could out-answer a
mainstream LLM. We agreed we can't, and shouldn't try to. The owner then steered
the conversation to the better question — not "how do we talk better than
ChatGPT" but "what can a multi-agent, multi-modal system actually *produce* that a
single model call can't" — and gave the example that anchored everything after:
type a domain name into a chatbot and we build and deploy you a real site on a
sub-URL. From there the centre of gravity moved from content to a produced
deliverable.

## What we've done
We settled the honest competitive story: we don't sell intelligence, we sell a
finished, verified, deployed, multi-modal artifact — something that needs
orchestration, tool-use with real side-effects, self-verification and deployment,
which is exactly what a single chat can't do. We reframed the chatbot as the cheap
*front door* to an operation, not an open question-answer bot. We took the owner's
idea of trialling free client-side tools and placed it correctly: they're a
brilliant demand-probe and top-of-funnel, but they can't be the paid product, and
we should rank them by intent depth, not raw traffic. We looked at the real fleet
— twelve live sites and seventeen pool verticals across a thousand-plus-domain
portfolio — and connected to the target-market ("news feed pooling") thread, whose
core rule we adopted: design once per vertical, specialise per site, because the
per-site difference *is* the product. Crucially, we found that the seventeen
verticals collapse into about five reusable operation-patterns (produce-and-deploy
a listing; a verified decision report; a cited comparison; a novelty shareable
piece; and business content/asset production) — so "a hundred domains" is really
about five things, specialised. We designed two passes of examples on real
domains, deliberately showing the weak first pass being sharpened into the
pattern-based second. And we compiled a full, honest inventory of what the
framework can actually do today.

## Where we are now
The strategy is clear and written down; nothing has been built and that was
deliberate — the owner asked to deliberate, not code. We have a workstream
directory with the standing docs (plan, notes, running log, this summary, and the
capabilities inventory). We have one important correction on the record: despite
talking about narrated video and voiceovers, **no voice, video or animation
modality is actually wired in** — only text, image and embeddings — so any
operation that needs those needs them added first. My standing recommendation is
to pilot one pattern on one vertical end-to-end, and specifically to prove it on
**Property** (produce-and-deploy a listing) rather than the domain-to-site version,
because it exercises the whole moat, has obvious willingness to pay, and isn't
sitting on our most bug-fought code path. The conversation is pausing here because
we're low on credits and the owner is taking the capabilities list to another
chat.

## Where we're going
Two open choices decide the next move. First, pick the pilot — my recommendation
is Property, but the owner may want the domain-to-site headline act, or to map all
seventeen pools to patterns first so he can see where his hundred domains land.
Second, decide how the free client-side widget layer and the paid operation fit
together as one funnel, and where the paywall actually sits (my position: free
cheap chat and free widgets as lead-gen; pay for the produced deliverable). Once a
pilot is chosen we move from deliberation into a real plan: confirm what
platform machinery it reuses, decide whether it needs the missing modalities, and
build one operation properly before generalising the pattern across the fleet.
