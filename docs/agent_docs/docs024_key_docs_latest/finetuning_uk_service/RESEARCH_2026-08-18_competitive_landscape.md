# RESEARCH — competitive landscape for finetuning.uk (2026-08-18)

Web sweep, four angles, run at the owner's request the day the £99 starting
price was set. Sources at the bottom; figures are the ranges the sources quote,
not our measurements.

## The four market tiers, and where £99 sits

| tier | what the buyer gets | price | our relation to it |
|---|---|---|---|
| **Done-for-you fine-tuning consultancy** | engineers, data prep, eval, deploy | **$5k–$180k** (data prep alone 30–50% of cost) | not our market; useful as the "what it normally costs" anchor |
| **Productised fixed-price fine-tune** | bounded package, still consultancy-shaped | **from ~$4,800** | the CHEAPEST comparable done-for-you offer is ~40× our £99 |
| **UK "AI chatbot trained on your data" agencies** | mostly RAG chatbots, hosted, branded | **£300–£500 setup + £200–£400/month**, custom builds £4k–£25k | the price anchor OUR buyer has actually seen; recurring-heavy |
| **DIY fine-tuning APIs** (OpenAI, Together, Mistral) | you do the work; OpenAI's adapter is locked to their API | compute pennies-to-dollars (Together LoRA from ~$0.48/M tokens; OpenAI $25/M + $0.50/hr; Mistral $4 min/job) | cheap but TECHNICAL, and mostly no model ownership — the segment we de-tech-ify |

## What this says about the position

- **"A genuine fine-tune, you own the file, ~£99, no technical skill needed"
  appears to be an unoccupied cell.** Everything done-for-you starts ~40× dearer;
  everything cheap requires the customer to be the engineer; and most of the UK
  agency tier is RAG-not-fine-tuning with a monthly meter running.
- **£99 is below the setup fee alone of a UK chatbot agency** — so we are not
  undercutting anyone's like-for-like offer; we are a different product at a
  price that invites a try.
- **Two risks the copy must handle:** (1) *suspiciously cheap* — next to $4.8k
  the buyer wonders what's missing; the answer is honest and should be said
  plainly (small model, automated pipeline, bounded scope, real person checks
  every run). (2) *category confusion* — at these prices UK buyers expect "a
  chatbot on my website"; we deliver a model file + a booked playground hour.
  Say what it is and is not, early. (Hosted chat on their site is the natural
  LATER upsell — that is exactly where the agency tier charges £200–£400/month.)
- **Margin context** (our measured numbers, RESULTS §7): ~£1 of GPU per customer
  journey ⇒ £99 carries ~99% gross margin on compute; the real costs are LLM
  spend and operator attention, which the concierge phase will measure.

## Owner's copy direction (recorded same day)

"Your company's voice, in a model you own" is the agreed message — but the copy
must be **friendly and expansive, not dense**; possibly a **glossary** for the
handful of unavoidable terms (fine-tune, model, open-source, GGUF). Positioning:
techie thing, non-techie voice, genuinely helpful in tone, enough visual
authority to be credible, not enough to intimidate. £99 start, low hundreds
credible for business buyers, reduce-later-not-raise-later.

## Sources

- [TechAhead — Custom LLM Development Cost 2026](https://www.techaheadcorp.com/blog/custom-llm-development-cost/)
- [AI Superior — Cost of Fine-Tuning LLM 2026](https://aisuperior.com/cost-of-fine-tuning-llm/)
- [Price Per Token — LLM Fine-Tuning Pricing 2026](https://pricepertoken.com/fine-tuning)
- [Stratagem Systems — LLM Fine-Tuning Business Guide / LoRA cost analysis](https://www.stratagem-systems.com/blog/llm-fine-tuning-business-guide)
- [APIpulse — AI API Fine-Tuning Costs 2026](https://www.getapipulse.com/blog-fine-tuning-costs-2026.html)
- [AI Pricing Guru — Together.ai pricing](https://www.aipricing.guru/together-pricing/) · [Mistral pricing](https://www.aipricing.guru/mistral-ai-pricing/)
- [Delegait — AI Chatbot Cost UK](https://delegait.co.uk/blog-ai-chatbot-cost-uk) · [AskMind](https://askmind.co.uk/blog/ai-chatbot-cost-uk-2026) · [AI Optimised](https://ai-optimised.co.uk/blog/ai-chatbot-cost-uk-small-business) · [Launchwork](https://launchworkdigital.co.uk/blog/ai-chatbot-business-uk)
- [Rapid Innovation — fine-tuning services](https://www.rapidinnovation.io/service-development/fine-tuning-language-model) · [Layer3 Labs — open-weights buyer's guide](https://www.layer3labs.io/open-weights/fine-tuning-open-weights-models)

## Positioning principles — OWNER-RATIFIED 2026-08-18 (restated by the owner for the record)

The owner restated the following analysis verbatim as the agreed basis; the
consulting threads (`copy_quality_two_stage`, `vigilant_designer_offer_analysis`)
should treat these as settled direction, not proposals:

1. **"Helpful though automated" has a concrete design implication: show the
   journey, not the machinery.** "Send us your examples → we train overnight →
   you chat with it the next day, booked hour, no setup" is genuinely reassuring
   and true — every step of it is now measured. **The concierge phase is an
   asset, not a stopgap:** "a real person checks every run" is differentiation
   the big platforms structurally cannot offer at £99.
2. **The authority-without-intimidation balance is earned through SPECIFICITY,
   not aesthetics:** real numbers ("trained in under an hour", "your data never
   trains anyone else's model"), a plain-English guarantee, one screenshot of
   the playground chatting. Authoritative to a business buyer in a way
   terminal-green styling never is — and it can't scare anyone.
3. **"Doesn't sound techie" must not slide into VAGUE.** The buyer who signs off
   £99–£300 at a small business is often its one semi-technical person; they
   will want one page that says exactly what they get (model, licence, file
   they can download, where it runs). **Shape: friendly front door, one honest
   "for your technical person" page behind it.**

> **OWNER CORRECTION, 2026-08-18, same day:** "though I am not sure a real person
> checks every run to be fair." — The "a real person checks every run" line in
> principle 1 is **[UNVERIFIED AS A STANDING PROMISE]** and MUST NOT reach site
> copy unless the concierge process genuinely commits to it. What is true today:
> the concierge phase means orders are run and reviewed by hand while volumes are
> tiny. Whether "checked by a person" survives as a durable, advertised guarantee
> is an OWNER OPERATIONAL COMMITMENT, not a fact a session may assert. The
> evidence_base claims gate would rightly refuse it as things stand. Safe forms
> until decided: "run by people, not left to a queue" (true of the concierge
> phase) or drop the line entirely.
