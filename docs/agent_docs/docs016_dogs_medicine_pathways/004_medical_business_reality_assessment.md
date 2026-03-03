# Business Reality Assessment — Agent Framework Commercialisation
## Date: 2026-03-03
## Status: Working notes for strategic direction

## Context

This document captures a frank assessment of the commercial viability of the agent orchestration framework, the canine biology project, and various business directions considered. It is intended as a reference for decision-making, not as a polished pitch.

## The Canine Biology Project as a Revenue Path

The project is intellectually interesting but it's a long road to revenue.

**Pharmaceutical company interest** is possible but the sales cycle would be 6-12 months minimum. Finding the right person in pharma is hard. Getting past procurement is harder. Their data governance concerns around an external AI system touching proprietary research data would be significant. And the framework would compete against established informatics tools these companies already use and have budget lines for — DiscoverX, Ingenuity Pathway Analysis, Elsevier's Entrez systems. These companies spend millions on tools like this from established vendors with track records. They don't buy from one-person operations with a demo, regardless of how good the demo is.

Not impossible. But not a quick path to income.

**A university veterinary research group** might be a stepping stone — they're more accessible, often have small budgets for tools, and would be genuinely interested. A collaboration gives credibility when approaching pharma later. But the income from an academic collaboration is minimal.

**The honest assessment**: the canine biology project is best treated as marketing spend, not as a product. Build one branch as a polished showcase (a few hundred dollars in LLM costs, a few days of work). Use it to demonstrate what the framework can do. The full 1M-agent run should wait until there's a reason — a customer, a conference talk, a funding pitch.

## What the Framework Actually Is

The framework does something specific well: it takes a high-level objective, decomposes it into a hierarchy of tasks, distributes those tasks across agents that communicate through Kafka, collects results, and assembles a final output.

The differentiators versus LangGraph, CrewAI, AutoGen and similar:
- Runs on real infrastructure (Kubernetes, Kafka, Postgres), not a Python library
- Hierarchical agent spawning with parent-child communication
- Workflow engine driven by data (SQL definitions), not code
- Multi-cluster dispatch (proven working)
- Agent-chassis pattern — every agent uses the same runtime
- Production-grade message passing, not in-process function calls

The weakness versus those competitors: they have documentation, tutorials, getting-started experiences, large communities, and VC backing. The framework has none of these yet.

## Revenue Paths Assessed

### 1. Website Building Pipeline (Most Mature)

The framework already takes a domain name or business description and produces a multi-page website deployed to GitHub/Cloudflare. This is a real service businesses pay for — agencies charge £500-5,000 for a basic business website.

**Strengths**: Pipeline exists end-to-end. Market is proven. Shortest path to actual revenue.

**Weaknesses**: The AI website builder space is crowded. Wix ADI, Durable, 10Web, Hostinger AI, B12 all exist. Differentiator would need to be output quality, niche targeting, or the multi-page content depth that simpler tools don't achieve. Current output quality may not be at sellable standard without refinement.

### 2. SEO/Content Generation at Scale

Agencies and marketing teams need large volumes of structured content — landing pages, product descriptions, localised variants. The framework could produce 1,000 structured pages for a keyword set.

**Strengths**: High volume, B2B buyer (agencies, in-house marketing teams), repeatable.

**Weaknesses**: Competitive space. Content quality needs to be consistently good. Buyers in this space are price-sensitive and comparison-shop aggressively.

### 3. Document Processing and Analysis

Take a large document corpus (contracts, regulatory filings, research papers, internal policies) and decompose, analyse, synthesise into structured knowledge. The agent framework's decompose-research-synthesise pattern maps directly to this.

**Strengths**: High price points (law firms pay £10,000-100,000+ for due diligence). Framework is genuinely well-suited to the task pattern.

**Weaknesses**: Requires domain expertise or a domain partner. Nobody buys document analysis tooling from someone who doesn't understand their documents. The legal/compliance space specifically was not of interest — finance or small business preferred.

### 4. Framework Sales (Infrastructure)

Sell the orchestration layer to companies building AI agent systems.

**Strengths**: High per-transaction value. Genuine technical differentiator (production infrastructure vs Python libraries). Growing market as more companies build agent systems.

**Weaknesses**: Developer adoption of frameworks is slow and fiercely competitive. Needs documentation, tutorials, getting-started experience. Probably needs open-sourcing the core to build adoption. Revenue would come from hosted services, enterprise support, or managed deployments — all of which take time to build. Longest path to income.

### 5. Domain Name Sales with Websites Attached

Register domains, generate websites for them using the framework, sell the domains with ready-made sites.

**Strengths**: Demonstrates the framework in action. Passive income potential.

**Weaknesses**: Most domains sell for £10-50 unless genuinely premium. Even with a nice website attached, competing with millions of parked domains. Low margin per unit, high volume needed to be meaningful. Likely not worth the time and attention versus other paths.

## The Three-Tier Model

The model that emerged from discussion: sell the output, sell the service, sell the tooling.

### How It Works

**Tier 1 — Sell the output**: Run the framework for a specific niche, produce websites (or reports, or content), sell them directly. This generates income and creates a live portfolio.

**Tier 2 — Sell the service**: Offer the production process as a service. "We'll build your website / generate your content / process your documents." Higher price point, some per-customer support needed.

**Tier 3 — Sell the setup**: Once the process is proven with 20-50 live outputs, sell the whole setup — framework, agent definitions, deployment pipeline, prompt library, training — as a business-in-a-box. Price point £5,000-25,000. The buyer runs it themselves for their market.

### Why This Works Better Than Selling the Framework Cold

Each tier validates the tier above it. A framework buyer can see the live output. A service buyer who outgrows the service might buy the setup. The framework is always running, always producing, always demonstrable. No vapourware.

### Where It Has Problems

Running three tiers simultaneously means three different customer types, three different sales processes, three different support needs. One person can't do all three well at the same time. Need to be sequential or very selective about which tiers to operate.

The domain selling tier specifically was assessed as not worth the time — low margin, high competition, distracting from higher-value work. Recommendation: drop this tier.

### Recommended Shape

Two tiers, not three:

**Tier 1 — The service**: Pick a niche. Build websites (or another product) in that niche. Charge enough per unit that each sale matters. Accumulate a portfolio of live output.

**Tier 2 — The setup**: Once 20-50 live outputs exist and the process is repeatable, sell the whole setup to agencies or entrepreneurs who want to offer the same service in their market. They're buying a business-in-a-box, not just software.

Then repeat for a different product. Each new product follows the same pattern: build the agents, run the service, sell the setup. The framework improves with each iteration.

### Niche Selection Matters

"We build websites" is too broad. The service needs a specific niche with:
- Definable content needs (so the agents can be tuned)
- Reachable buyers (so you can find them without a large marketing budget)
- Enough willingness to pay (so each sale covers the cost of production plus profit)
- Not well-served by existing AI tools (so there's a gap)

Examples considered: veterinary practices, independent financial advisors, tradespeople, local hospitality businesses. The right choice depends on which market is most accessible and which produces the best output quality from the existing pipeline.

### Second and Third Products

The framework is not limited to websites. The same two-tier model applies to:
- Report generation for a specific industry (property reports, market analyses, due diligence packs)
- Content production at scale (social media calendars, email sequences, product descriptions)
- Document processing (policy review, contract analysis, compliance checking)

Each new product strengthens the pitch for setup buyers: "we've deployed this framework for website generation, content production, and document processing — here's the live output from each."

## Practical Next Step

The website pipeline already works. Rather than building more infrastructure or planning the 1M agent demo, the highest-value thing to do right now might be:

1. Pick a niche
2. Produce 10 websites in that niche using the existing pipeline
3. Put them live
4. Attempt to sell them as a service
5. Validate the model with real money before investing more time in the framework

If that works, the running business funds the framework development, and the framework development makes the business better. That's a sustainable loop rather than a speculative bet on a large demo.

If it doesn't work — if the output quality isn't sellable, or the niche doesn't have buyers, or the economics don't work — that's useful information too, and it's cheaper to discover by building 10 sites than by building a 1M-agent infrastructure run.

## The Canine Biology Project's Role Going Forward

The project remains valuable as:
- A portfolio piece and demo of the framework's capabilities
- A technically impressive showcase for conference talks, blog posts, LinkedIn content
- A proof point when pitching framework sales: "here's what 1M agents produced"
- A genuine research tool if academic partnerships materialise

It should not be:
- The primary path to revenue
- The first thing built (the revenue-generating service should come first)
- Framed as a finished product until it's had expert review

Build one branch (cardiovascular system) as a polished showcase when resources allow. Defer the full 1M run until there's a commercial reason for it.

## Open Decisions

1. **Which niche for the website service?** Needs to be a market that's accessible and where the pipeline produces good output.

2. **Is the website pipeline at sellable quality?** Needs honest assessment — if the output isn't good enough, time needs to be spent on quality before attempting sales.

3. **Website service or different first product?** Websites are the most mature pipeline but also the most competitive market. Is there a different product type that would be less competitive and more profitable?

4. **Solo or with a partner?** The two-tier model is a lot for one person. A technical co-founder, a sales partner, or even a freelance designer who improves the website output could change the economics.

5. **Timeline and financial runway.** How urgently is income needed? This affects whether to optimise for fast but small revenue (website service) or slower but larger revenue (framework setup sales).

## Related Documents

- 014: Canine biology project baseline (the technical design for the demo project)
- 013: Scaling analysis (infrastructure scaling from 10K to 1M agents)
- 006c: Multi-cluster dispatch handoff (current state of the framework)
