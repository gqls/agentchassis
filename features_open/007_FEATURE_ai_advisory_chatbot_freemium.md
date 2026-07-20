# 007 FEATURE — AI-advisory chatbot on the AI-influence page, freemium/paid

**Raised:** 2026-07-20, by the owner, alongside `006`. Explicitly deferred by the
owner at the time of raising: *"that is another feature for another time"*.
**Status:** idea captured, deliberately NOT designed. Do not start without the
owner reopening it.
**Depends on:** `features_open/006_FEATURE_gaswholesalers_repositioning_and_ai_influence_page.md`
— the page must exist and be credible first; the bot is an extension of it.

## The idea

Owner, verbatim:

> "Later I want to have an AI chat bot on that page that hits a high quality model
> to further explore corporate options for these AI challenges that can be a
> freemium/paid model (perhaps 2 free goes then pay per hour or per 50 submissions
> or something) The prompts will be very well created for their purpose - better
> than theirs and will involve deep research and caching for similar queries"

So: a conversational adviser sitting on the AI-influence page, letting a senior
reader explore *their own* corporate situation beyond what a static page can cover.

## The four load-bearing pieces of that quote

1. **"hits a high quality model"** — this is not the cheap tier. The product is the
   quality of the answer to an expert reader; a weak model is worse than no bot,
   for the same reason a generic page is (see `006` §4).
2. **"better than theirs"** — the differentiator is *prompt engineering plus
   research*, not model access. The buyer could open a chat window themselves; what
   they are paying for is a prompt built for this domain and this decision, primed
   with material they do not have to hand.
3. **"deep research"** — answers should be grounded in retrieved material, not
   model recall. Overlaps the platform's existing research-agent and RAG machinery
   (`rag_lookup` / `knowledge_base`) — check before building anything new.
4. **"caching for similar queries"** — the margin mechanism. Executives in one
   industry ask overlapping questions; a semantic cache turns the expensive first
   answer into a cheap repeat one. This is what makes per-submission pricing work.

## Pricing shape (owner's sketch, not settled)

Two free interactions, then paid — "per hour or per 50 submissions or something".
Not designed. Note it interacts with `003_FEATURE_paid_tier_beyond_news`: this is a
second paid surface, and the platform should not end up with two unrelated payment
mechanisms.

## What will bite when this is built (capture now, cheap later)

- **Claims exposure.** A bot advising executives is the platform asserting things,
  at scale, without a human between generation and reader. Everything the
  claims-verification workstream exists for applies here **with no post-deploy
  review lane** — nothing scans a live conversation. If this ships, the
  verification question becomes *pre-response*, not post-publish. That is a
  genuinely harder problem than the page.
- **Advice liability.** Corporate advice to named industries, sold for money,
  is a different risk class from published content — cf. the vetcomparison
  fabricated-prices episode, which had legal exposure at far lower stakes.
- **Cache poisoning.** A cached answer that was wrong gets served repeatedly and
  cheaply. Cache invalidation needs to be part of the design, not an afterthought.
- **Freemium abuse.** "Two free goes" needs an identity boundary or it is two free
  goes per browser session, i.e. unlimited.

## Open questions

1. Is the bot the product, or the lead-generation front end for a consulting
   conversation? Those are different builds.
2. Does an answer cite its sources to the reader? (For this audience: probably
   yes, and that also constrains the model to grounded output.)
3. Does the claims layer gate bot responses, or is that out of scope by design?
   **This is the question that most affects the claims-verification roadmap** and
   should be answered before that thread plans beyond V4.
