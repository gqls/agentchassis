# Business Plan — finetuning.uk

Pragmatic, solo-operator scale. Not a corporate deck — a decision doc.
Updated alongside the FOCUS doc as we learn.

Last touched: 2026-04-21

---

## 1. One-page summary

**What:** A RAG platform for SMEs. Upload your documents, the platform
cleans and organises them automatically, you get a chat interface and API
that answer questions grounded in your own content.

**Who for:** Technical-adjacent SME operators — 10-50 person businesses,
agencies, consultancies, knowledge-intensive services (law, accounting,
specialist consulting). People with messy knowledge in Drive / Notion /
Slack / email who've failed once already at "just plug ChatGPT into our
docs".

**Why us:** The framework we've already built does automatic data
curation — deduplication, quality scoring, classification, PII scan,
inconsistency flagging — as part of ingestion. Competitors treat bad
data as the customer's problem. We treat it as the product.

**Operator:** Solo, UK. Runway for 12+ months. Target £9-12k/month gross
within 12 months.

**Acquisition:** Cold only. Content-led inbound through finetuning.uk,
which is itself built and maintained by the framework. Interim contract
gigs (capped at 50% of time) pad early cash flow.

**Flagship first, bespoke later:** RAG platform is the product. Custom
AI assistants, text LoRA, image LoRA, multi-agent workflows are later
tiers added as product matures and customer demand surfaces.

---

## 2. Problem and opportunity

### The problem SMEs actually have

Not "I need to fine-tune a model" — they don't know what that means and
mostly don't need it. The real problem:

> **"Our team wastes hours every week looking up information we already
> have written down somewhere. We tried ChatGPT on our docs but it didn't
> know our stuff properly. Our data is a mess and we don't have time to
> clean it."**

The gap between "we have docs" and "we have an AI that knows our docs"
is 90% data curation and 10% retrieval. Every vendor sells the retrieval
half. Nobody sells the curation half integrated.

### Why RAG specifically

RAG is the tool that fits this problem:
- No training dataset needed — just the docs they already have
- Answers cite sources — trust is buildable
- Updates immediately when docs change
- Runs on commodity infrastructure (we already have it)

Fine-tuning only makes sense when tone, format, or offline deployment
matters. For "make AI know our stuff", RAG is the honest answer.

### Why now

Three things converge:
- LLM costs dropping makes hosted chat affordable per query
- Open-weight models (Llama 3, Qwen) are good enough at grounded answering
- pgvector and Ollama mean the whole stack self-hosts cleanly
- SMEs have tried DIY ChatGPT integrations and many have hit the data wall

---

## 3. Product

### Flagship: finetuning.uk RAG Platform

Core features at launch (month 2-3):

| Feature | What it does | Status |
|---|---|---|
| Upload | PDF, DOCX, HTML, CSV, plain text, URLs | Built (rag_index action) |
| Data curation pipeline | Parse → classify → dedup → quality-score → PII scan → inconsistency flag | Partial (dedup exists, rest to build) |
| Curation report | User reviews what was kept, dropped, flagged, before commit | To build |
| Index | Chunk + embed + store | Built (rag_index) |
| Chat UI | Grounded answers with citations | To build (thin wrapper) |
| API | Query endpoint for integration into customer tools | Built (rag_lookup), UI to build |
| Source management | List, preview, remove, re-index | To build |
| Analytics | Queries/day, answer confidence, failing queries | To build |

Later features (month 4-12):

- Google Drive / Notion / Slack integrations
- Text LoRA training for brand voice (wraps existing Unsloth flow)
- Image LoRA training for brand visuals
- Multi-agent workflows (upsell to bespoke tier)
- SSO / audit logs (enterprise tier)

### What we're NOT building

- General chatbot not grounded in user data — plenty of alternatives
- Our own foundation model — use Llama/Qwen/Mistral or Claude API
- Vector DB as a product — use pgvector, it's enough
- Marketplace for third-party LoRAs — scope creep
- Free tier (beyond generous trial) — kills margin, attracts wrong users

---

## 4. Target customer (detailed)

### Primary ICP

| Attribute | Value |
|---|---|
| Size | 10-50 people |
| Sector | Knowledge-intensive services — law, accountancy, consultancy, niche SaaS, specialist engineering |
| Geography | UK and EU first (data-residency story) |
| Role we sell to | Operations lead, Head of Knowledge, Partner, CTO, or founder who does ops |
| Budget | £500-2,000/month recurring feels normal, £3-10k setup tolerable |
| Tech maturity | Uses Google Workspace / M365, maybe Notion, knows what an API is but doesn't build one |
| Trigger | Tried ChatGPT Enterprise or Copilot, found it insufficient, doesn't know why |

### Why this ICP

- Big enough to have a data-volume problem
- Small enough that buying is one person's decision
- Knowledge-intensive → the product actually saves time
- Technical-adjacent → they can evaluate the product without hand-holding
- UK/EU → data residency is a real concern they'll pay to solve

### Not the right customer

- Consumer — no willingness to pay
- Pre-product startups — no stable content to index
- 200+ person enterprises — procurement cycle too long for solo sales
- Anyone asking "can you integrate with SAP / Oracle / Salesforce enterprise" — too much integration work

---

## 5. Pricing

### Platform subscriptions

| Tier | Price | Limits | Target customer |
|---|---|---|---|
| **Trial** | Free, 14 days | 100MB, 100 queries | Evaluation |
| **Starter** | £199/month | 1GB, 5k queries, 1 source type | Solo operator / micro-business |
| **Growth** | £499/month | 10GB, 25k queries, 3 source types | SME primary tier |
| **Pro** | £1,499/month | 100GB, unlimited queries, all integrations, priority | Larger SME / agency with multiple clients |
| **Enterprise** | Quote | Custom — SSO, private deploy, DPA | Case-by-case |

### Setup / concierge fees (the services layer)

| Service | Price | Who needs it |
|---|---|---|
| Data audit + curation plan | £750 | Anyone with >5GB messy data |
| Guided onboarding | £2,000 | Most customers in months 1-3 while self-serve matures |
| Custom integration (source connector, API work) | £2-5k | Customers with unusual data sources |
| Custom LoRA (voice or image) | £5-10k | Customers who outgrow RAG defaults |
| Bespoke multi-agent workflow | £15-30k | Occasional large projects |

### Why this shape

- **£199 starter** kills the objection "this is too expensive to try"
- **£499 is the expected revenue tier** — enough margin to sustain a few support conversations, low enough for a department card without approval
- **£1,499 pro** exists for customers with real scale; higher-margin
- **Setup fees front-load cash flow** — a customer at £499/month with a £2k setup pays back in months one, not six
- **Bespoke services let us take on work that doesn't fit the platform**
  without saying no

---

## 6. Unit economics

### Cost per customer per month (Growth tier, £499)

| Cost | Estimate | Notes |
|---|---|---|
| Storage (S3 + PG) | £1-5 | Scales with data volume |
| Embeddings (Ollama CPU) | £2-10 | Re-embed on re-index only |
| Inference (chat) | £15-50 | Claude Haiku/Sonnet or Llama on GPU — varies with usage |
| Shared GPU amortised | £20-80 | If GPU always-on, split across customers |
| Payment processing | £15 | Stripe ~3% |
| Bandwidth, monitoring, misc | £5 | |
| **Total variable** | **£58-165** | |
| Support (30 min avg/month at £100/hr notional) | £50 | Not real cash but real time |
| **Gross margin per customer** | **£284-391 (57-78%)** | Healthy |

### Fixed costs (monthly)

| Cost | Estimate | Notes |
|---|---|---|
| Cluster (k8s nodes) | £200-500 | Depending on size |
| GPU (shared, always-on) | £200-700 | CPU ollama is free; GPU is for larger models |
| Domain + DNS + email | £20 | |
| Auth (Clerk/Supabase) | £25-100 | Scales with MAU |
| Stripe, tax handling | £0 + %age | |
| Dev/staging envs | £50 | |
| Claude API (dev + fallback inference) | £100-500 | Variable |
| Tools (Linear, Notion, etc.) | £50-100 | |
| **Total fixed** | **£645-1,970** | |

### Break-even

| Customers at Growth tier | MRR | Variable costs | Gross profit | Coverage of fixed |
|---|---|---|---|---|
| 2 | £998 | £200 | £798 | Partial |
| 5 | £2,495 | £500 | £1,995 | Full |
| 10 | £4,990 | £1,000 | £3,990 | Full + £2k surplus |
| 20 | £9,980 | £2,000 | £7,980 | Target reached |

Realistic first-year count: 5-10 customers mixed across tiers plus one
or two Pro. That gets us to £5-9k MRR before setup fees and bespoke work.

---

## 7. Revenue projection (12-month view)

Numbers are **planning estimates**, not forecasts. Adjust as real data
arrives.

| Month | Customers | Avg tier | Recurring | Setup fees | Bespoke | Total | Cumulative |
|---|---|---|---|---|---|---|---|
| 1 | 0 | — | £0 | £0 | £0 | £0 | £0 |
| 2 | 1 concierge | — | £499 | £2,000 | £0 | £2,499 | £2,499 |
| 3 | 2 | Growth | £998 | £2,000 | £2,000 interim gig | £4,998 | £7,497 |
| 4 | 3 | Growth | £1,497 | £2,000 | £3,000 interim gig | £6,497 | £13,994 |
| 5 | 4 | Growth | £1,996 | £2,000 | £0 | £3,996 | £17,990 |
| 6 | 5 | Growth + 1 Pro | £3,495 | £4,000 | £5,000 (first LoRA) | £12,495 | £30,485 |
| 7 | 6 | | £3,994 | £2,000 | £0 | £5,994 | £36,479 |
| 8 | 7 | | £4,493 | £2,000 | £0 | £6,493 | £42,972 |
| 9 | 8 | | £4,992 | £4,000 | £10,000 bespoke | £18,992 | £61,964 |
| 10 | 9 | | £5,491 | £2,000 | £0 | £7,491 | £69,455 |
| 11 | 10 | | £5,990 | £2,000 | £5,000 | £12,990 | £82,445 |
| 12 | 10 + 2 Pro | | £8,988 | £4,000 | £5,000 | £17,988 | £100,433 |

Year-end run rate: ~£9k/month recurring + ~£3-6k/month setup and bespoke
averaged = **£12-15k/month** in month 12 terms, **£100k** annual revenue
for year 1.

Hits the £9-12k target if the middle of the range holds.

### Sensitivity

- **Drop customer count by 40%** (real churn or slow acquisition):
  month 12 recurring ~£5,400. Supplemented by bespoke to reach £8-10k.
  Still viable but tight.
- **Raise average tier** (more Pro customers, fewer Starter): month 12
  recurring could be £12-15k from same customer count.
- **No bespoke projects at all**: strip £30k out of year 1 revenue.
  Forces reliance on recurring, which means customer count 2x what's shown.

---

## 8. Costs detail (first 12 months)

| Category | Month 1-3 | Month 4-6 | Month 7-9 | Month 10-12 | Year total |
|---|---|---|---|---|---|
| Infrastructure (k8s, GPU, storage) | £1,500 | £1,800 | £2,400 | £3,000 | £8,700 |
| Claude API + LLM inference | £600 | £900 | £1,500 | £2,100 | £5,100 |
| Third-party services (auth, Stripe, tools) | £450 | £450 | £600 | £600 | £2,100 |
| Domain, comms, misc | £150 | £150 | £150 | £150 | £600 |
| Legal (terms, DPA, privacy review) | £1,500 | £0 | £500 | £0 | £2,000 |
| Accounting | £300 | £300 | £300 | £300 | £1,200 |
| Marketing (ads, if any) | £0 | £500 | £1,000 | £1,500 | £3,000 |
| **Total** | **£4,500** | **£4,100** | **£6,450** | **£7,650** | **£22,700** |

Net margin on the £100k revenue projection ~£77k. For a solo operator
that's the full salary plus reinvestment room.

---

## 9. Go-to-market

### Primary channel: content-led inbound via finetuning.uk

The site does its own marketing:
- Decision guides ("Should you fine-tune?", "RAG vs context stuffing",
  "How to prep messy docs for AI")
- Vendor comparisons (honest, includes us and competitors)
- Free tools (ROI calculator, readiness checker, cost estimator)
- Case studies once we have them
- Newsletter for retention

Our framework produces this content faster than any manual competitor
can match. That's the structural moat.

### Secondary: targeted outbound

Not mass cold email. Targeted:
- Specific SMEs in primary ICP sectors, researched and reached with a
  warm observation (not "checking in" spam)
- AI-adjacent communities (Indie Hackers, r/LocalLLaMA subset, specific
  Slacks) where contribution buys standing
- Podcast guest appearances as positioning builds

### Tertiary: partner directory

Listed in our directory + we're listed in theirs. Mutual referrals.
Not scaleable but high-quality.

### What we won't do

- Mass cold email / LinkedIn spam
- Paid ads before product-market fit
- Conferences (too expensive for solo, low signal)
- Reseller deals (lock-in, low margin, premature)

---

## 10. Operations

### Solo-operator time allocation

Weekly rough plan, 40-50 hours:

| Activity | Hours |
|---|---|
| Product build (platform features) | 15-20 |
| Content + SEO (via framework, with human polish) | 5-8 |
| Customer delivery (onboarding, support, concierge work) | 8-12 |
| Sales (calls, follow-ups, proposals) | 3-5 |
| Interim gigs (capped at 50%) | 0-20 |
| Admin (accounting, invoicing, emails) | 2-3 |

Interim gig hours come out of product-build hours when taken. Cap:
interim can take at most 50% in any week.

### Outsourcing / contractors

Not hiring full-time in year 1. Potentially outsource:
- Ad-hoc copyediting for published content
- Legal review for DPA / terms
- Specific technical jobs that aren't core (design polish, logo work)

Partner directory = overflow channel for work we can't take.

### Support SLA (internal commitment, not published)

- Starter: 48h response, best-effort
- Growth: 24h response, 1 business-day resolution on P1 issues
- Pro: 8h response, same-day on P1
- Enterprise: per DPA

---

## 11. Risks and how to think about them

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Acquisition slower than projected | **High** | High | Interim gigs, content compounding, lower burn rate |
| Customer data leak | Low | Extinction | Tenant isolation from day one, encryption at rest, regular security review |
| Key competitor releases similar data-curation angle | Medium | Medium | Build the reputation first, keep iterating on the pipeline, move to verticals if generic space gets crowded |
| Claude / OpenAI get cheap enough that hosted chat becomes free | Medium | Medium | We don't compete on chat — we compete on data curation + integration. Cheaper inference helps us. |
| Operator burnout (solo, no co-founder) | Medium | High | Strict time-caps, interim gigs for cash flow relief, partner directory to hand off work |
| Infrastructure cost spike (unexpected usage) | Medium | Medium | Usage-based pricing aligns cost to revenue; hard rate limits on free tier |
| One big customer dependency | Medium (later) | High | Cap any single customer at <30% of revenue once we have enough to choose |
| Compliance / GDPR violation | Low | High | UK/EU data residency, DPA template reviewed by lawyer, PII handling from day one |
| LoRA / fine-tuning feature fails on first customer | Medium | Low | That feature is month 4+; RAG is proven first; LoRA is upsell, not foundation |

### The biggest risk

Is that we build a product nobody wants. Mitigated by: using the product
ourselves first, onboarding concierge customers during build so they
shape it, and not committing to features with no customer asking.

The second biggest is acquisition time. Mitigated by interim gig capacity
and the fact that the content engine compounds the longer it runs.

---

## 12. Milestones and decision points

### Milestone 1 — End of month 1
- Platform MVP deployed at finetuning.uk/app, behind auth
- `tenant_id` isolation enforced and tested
- We can upload, index, and chat on our own test data
- Site repositioned publicly

**Decision:** if the MVP isn't working well enough for us to use, don't
start selling. Fix it first.

### Milestone 2 — End of month 3
- 2 paying customers (concierge-onboarded)
- Data curation pipeline visible in UI
- Stripe billing live
- First case study published

**Decision:** if no customers yet despite product being usable, spend
month 4 exclusively on acquisition (ads, outbound, partnerships).

### Milestone 3 — End of month 6
- 5 paying customers, mixed tiers
- Self-serve signup live (with trial)
- First LoRA customer onboarded (as upsell from RAG)
- MRR £2-4k

**Decision:** at this point, is the product working? Is pricing right?
Iterate based on real data, not plan.

### Milestone 4 — End of month 12
- 10+ paying customers
- £9-12k/month total (recurring + setup + bespoke)
- Content engine producing consistent inbound
- Clear picture of which offers drive the book

**Decision:** raise prices, hire a contractor, or stay solo? Data will tell.

---

## 13. Assumptions worth testing early

These are the assumptions the plan rests on. Each should have a cheap
test in the first 60 days.

1. **SMEs in primary ICP will pay £499/month for RAG over their docs.**
   Test: talk to 10 such SMEs, pitch the concept, gauge willingness to
   pre-pay. Cost: 10-15 hours.

2. **Data curation is a real, valued differentiator.**
   Test: in those conversations, check whether "we automatically clean
   and organise your data" lands harder than "we do RAG". If it doesn't,
   the pitch angle is wrong.

3. **Content-led inbound works in this space for a solo operator.**
   Test: publish 10 articles in first 60 days, measure traffic and
   contact-form submissions. If zero inbound, channel mix needs to change.

4. **The framework can automate curation well enough that we don't drown
   in manual review.**
   Test: run 3-5 concierge jobs manually, measure time spent fixing
   agent output vs doing it from scratch. If >50%, automation isn't ready.

5. **Setup fees are collectible without friction.**
   Test: on first customer, charge the setup fee before doing the work.
   If they baulk, the model needs adjustment (subscription-only?).

---

## 14. Open questions

1. **Legal entity and VAT.** Sole trader for simplicity, or Ltd for
   liability shield? Accountant question.
2. **Liability cap in terms of service.** Maximum refund = last month
   paid is standard; verify with lawyer.
3. **Base model licence compliance.** If we offer LoRA on Llama, we
   need to comply with Meta's commercial terms. Check thresholds before
   launching LoRA tier.
4. **Pricing currency.** GBP primary, but Stripe can charge USD/EUR —
   set at launch or later?
5. **UK AI regulation trajectory.** Not yet settled; monitor. Being
   UK-based may become either advantage or burden depending on where
   the rules land.

---

## 15. Changelog

- 2026-04-21 — Initial draft. Spun out from FOCUS doc after UI-first
  decision. Revenue projections, unit economics, 12-month roadmap,
  milestones, assumptions. Numbers are planning estimates, refine as
  data arrives.
