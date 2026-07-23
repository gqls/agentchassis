# README — where we are: per-site AI section / operation

*Owner's running log. Plain prose, append-only, newest at the bottom. Add dated
corrections below rather than editing earlier entries.*

---

**2026-07-21 — the brief, and the first pass at challenging it**

The idea: on most sites we're about to build, put a link in the nav to an "AI
insights" (or similar) area. Under it, two things. First, a state-of-the-art AI
chatbot specialised to that site's vertical and that site in particular. Second,
a page/section of honest articles about the risks and benefits of AI in that
field — deliberately straight about AI's threat to specific jobs (not general
copy, but named job titles), the honest capabilities of AI, how fewer people
will be needed to do the same work — paired with a strong constructive side on
how to be the person who *doesn't* get laid off, and how businesses can use AI to
grow rather than to cut. All of it grounded in real, current news and research,
and made site-specific so we don't trip duplicate-content problems across similar
sites.

The chatbot should ideally run on a separate cluster wired to the best,
most-specialised models (text, image, voice, code, reasoning, embedding), be
prompted and gated to stay on-vertical, and be paywalled after a couple of free
tries a day. The whole thing should also quietly advertise that our framework
itself is for sale/rent/subscription, aimed at people who want to be ahead in AI.
There may be affiliate/commission money later (newsletters, courses) but that's
secondary. The real point: genuinely help our users get ahead of a fast-moving,
possibly-rough AI transition.

My first response pushed back on three things. (1) This is really *two* products
with opposite risk and cost — cheap, safe, controllable editorial vs. an
expensive, open-ended live chatbot — so we should ship the editorial first and
treat the chatbot as its own phase rather than bundling them. (2) We should not
claim to be "smarter than ChatGPT" — we'd lose that fight; we compete on
grounding, specificity, structure and freshness instead. (3) The paywall probably
shouldn't sit on the chat itself (that just adds friction against free frontier
chat) but on a *deliverable*, with a cheap free chat as the lead-in.

**2026-07-21 — the reframe that matters (owner steered it here)**

The owner's follow-up moved the centre of gravity, and it's a better centre. The
question isn't "how do we make a chatbot that beats a foundation model at
talking." It's "what can a *multi-agent, multi-modal* system actually *produce*
that a single model call can't." The honest answer isn't intelligence — it's
that we can chain specialist models (a cheap one for chat, a reasoning one for
the hard step, a good image model, a voice model, an embedding model for
retrieval), verify the result with our own review loops, and then *do something
real with it* — deploy it, publish it, hand back a finished artifact. A foundation
chat gives you pieces inside a conversation; we can give you a produced, checked,
delivered thing.

So the shape is shifting away from "AI insights + honest job articles" as the
headline, toward: **each site gets one signature AI *operation*** — an
interactive, chatbot-fronted tool that produces a real, original, site-specific
artifact or service, mostly business-oriented (because most sites are) but not
always. The owner's own example is the clearest: a web-design site where you type
a domain name and we generate and deploy a real site for you on a sub-URL of one
of our short domains. That's the platform selling itself by *doing the thing*
live. The honest job articles don't go away — they become the supporting,
trust-building, SEO layer around the operation, not the product itself.

The cheap-chat idea fits neatly: the chatbot is the *front door* to the operation
— a cheap model with a tight vertical prompt that answers light questions and,
more importantly, gathers what the operation needs and then triggers the
expensive multi-agent pipeline for the paid deliverable.

Still deliberating, deliberately not building yet. The open question I most want
to work through next: what is the *right* signature operation per site — the one
artifact each site's audience would genuinely want produced — and how much of it
is business utility vs. shareable novelty.
