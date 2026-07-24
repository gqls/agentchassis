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

**2026-07-24 — this idea already existed, and a parallel conversation sharpened
the middle of the funnel**

Two things happened this session. First, before touching the new material I
finally did the grep I should have done on day one, and found that this whole
line of thinking isn't new — four days before this workstream opened, the owner
had already committed to almost exactly this idea for one real site.
`features_open/006` is gaswholesalers.com being repositioned away from a fake
"we supply gas" story (it never did) toward an honest analysis-and-tools site
for oil and gas executives, with a flagship "AI influence" page covering the
threats and advantages of AI in that industry — truthfully, and remediated at
three separate levels: what the company should do, what an employee should do,
and what a person should do for their own career. `features_open/007` is the
advisory chatbot that was meant to sit on that page, and — importantly — the
owner was explicit that it should run on a genuinely high-quality model with
real research behind it, not a cheap one. That directly contradicts something I
had decided on day one of this workstream (that the chatbot should always be a
cheap front door), so I've corrected that: a cheap sticky utility and an
expensive, high-quality advisory conversation are two different products, and
which one a given site gets depends on how valuable its audience is, not a
blanket rule. I filed a new feature entry, 013, that ties all of this together
so it isn't scattered across three places.

Second, the owner took our capabilities list to a separate conversation with
Gemini and brought the results back. Most of it confirmed what we'd already
worked out, but one genuinely new and useful idea came out of it: a middle
tier, between the free calculators and the expensive paid deliverable — a
cheap, narrow, AI-powered tool that a professional in that field would
bookmark and come back to every week, not to make money directly but to build
the habit and the traffic. That was missing from our plan; the free tools and
the expensive report were covered, but nothing sat in between. I've kept that,
along with a handful of reusable shapes for what such a tool can look like, and
a way of thinking about each site along two dimensions at once — who they are
and what kind of risk they carry, separately from what we actually produce for
them.

I did not keep everything. The Gemini conversation also worked through fifteen
example domains from the owner's portfolio in a lot of depth, but when I
checked, none of those fifteen are actually set up on the platform yet — so
every guess about what each one is for was made purely from the domain name,
which is exactly the mistake we've been warned about before on this project.
I've kept those examples as raw brainstorming material only, clearly marked as
unverified, not as a real plan for those sites.

Nothing has been built. The next real decision is which single site or pool to
prove this whole approach on first — and now there are two strong candidates
rather than one: a plain business site to prove the "produce and deploy a real
thing" pattern cleanly, or gaswholesalers.com itself, which already has an
owner-approved brief waiting and would prove the more ambitious,
conversation-based version of the idea at the same time.

**2026-07-24 — a second batch, this time on real sites, changed the pilot pick**

The owner brought back a second set of domains from the same outside
conversation, and this time asked for them explicitly because they're already
live on our platform, not hypothetical. That mattered, because checking them
properly against what's actually built found something better than anything
we'd designed ourselves.

Robot-hands.com turns out to already be a real, working gripper-comparison
site for robotics engineers, with three calculators and a matching tool
already live. The outside suggestions for it mostly duplicated what already
exists — which is actually useful information, because it means the free and
mid-tier layers of our plan are already proven there in practice, and the only
real gap is the expensive, verified deliverable at the top. That's a much
smaller, much safer first build than starting from nothing, and it happens to
also fix a real data-honesty bug we already knew about on that same site. That
is now my top recommendation for where to prove this first.

Two of the outside guesses turned out to be flatly wrong once checked against
the real site content, and one of those was instructive. Idea.uk was guessed
to be about patents and inventions, purely because of the word "idea" in the
domain name — but the real site is about validating a business idea, nothing
to do with patents at all. While writing this up, a separate team working on
that exact site reported, on the same day, that they'd just verified live
that idea.uk already has a real, working, paid version of exactly the kind of
tool we've been designing in the abstract. So idea.uk isn't a site for us to
build on — it's the best evidence we have that the whole idea already works in
practice, and it belongs to someone else's ongoing work, so we're reading their
notes rather than touching it.

Leopardessconsulting.co.uk was also guessed wrong — assumed to be a general
management consultancy, when it's actually an AI-engineering consultancy with
real, audited case studies of our own platform's work, closely related to
another of our own AI-services sites, ai-agent-orchestration.com, which the
outside suggestions matched very well. Both of those already have real tools
live and are a strong second candidate, as a pair.

Nothing has changed about gaswholesalers.com — it's still waiting on its own
content rewrite before any tool gets built on top of it, and building
something shiny on a site that still has known false claims on it would make
that problem worse, not better.

So the shortlist for "prove this first" is now, in order: robot-hands.com,
then the two AI-services sites together, with gaswholesalers.com queued behind
its own rewrite and idea.uk kept as a reference to learn from rather than a
site to build on.

**2026-07-24 — pausing, deliberately, until robot-hands.com's own thread is
done**

The owner's call: don't start building the robot-hands.com pilot yet — wait
until the team already working that site finishes what they're doing. That's
the right call, and checking their status confirms why: they're still
mid-stream. Six rounds of fixes are done and live, but the specific
fabricated-stats bug that made robot-hands attractive as a pilot in the first
place is **not actually closed yet** — only patched on that one site, with the
proper fix and a check across the rest of the fleet still to come. Building a
verified, evidence-based tool on top of a site whose own data-honesty bug is
still open would be building on ground that's still moving.

So this workstream is paused here. Nothing else is blocked — the strategy,
the docs, and the ranked shortlist all stand — we're just not starting
implementation on robot-hands.com until that other team's work reaches a real
stopping point.
