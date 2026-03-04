# Vertical Cluster Architecture

## Specialised Knowledge Verticals for Domain Content Authority

---

## 1. The Concept

The pipeline currently treats every domain the same way: classify it, plan pages, write content, build site. The content comes from the LLM's general training knowledge, supplemented by whatever the research agent finds through web scraping. This produces generic content that looks like every other AI-generated site.

The vertical cluster architecture changes this. Instead of a flat pipeline, domains are routed to specialised vertical orchestrators that maintain their own deep knowledge bases, research strategies, content patterns, and monetisation approaches. Each vertical accumulates expertise over time — the tenth domain in a vertical benefits from everything the first nine taught it.

The verticals are logical separations running on shared infrastructure. They use existing mechanisms: `knowledge_base` table collections for knowledge storage, `rag_lookup` and `rag_index` for knowledge access, agent definitions with vertical-specific workflows, and the standard spawn/call orchestration patterns. No new Go code is required for the core concept — it's a workflow and data organisation change.

As a vertical matures and its workload grows, it can graduate to physical cluster separation by changing `spawn_agent` to `dispatch_agent` with a target cluster. The vertical orchestrator's workflow doesn't change. This is the hybrid path: logical now, physical later.

---

## 2. What a Vertical Contains

A vertical is four things:

### 2.1 A Knowledge Base Collection

The `knowledge_base` table already supports collection-scoped storage via its `collection` field. The `rag_lookup` action already filters by collection. Each vertical owns one or more collections:

- `collection: "veterinary"` — breed health profiles, treatment protocols, cost structures, practice quality indicators
- `collection: "energy_wholesale"` — market structure, contract analysis, regulatory framework, supplier data, benchmarking
- `collection: "finance_mortgage"` — interest rate dynamics, lender criteria, product structures, affordability models, specialist situations

Knowledge accumulates across all domains processed in that vertical. The first domain triggers foundational research. The tenth domain benefits from everything already indexed.

### 2.2 A Research Strategy

Different verticals need different kinds of research from different authoritative sources. The research strategy defines:

- **Source list**: What authoritative sources to consult (BSAVA manuals for veterinary, Ofgem data for energy, PRA rules for mortgage)
- **Research priorities**: What knowledge gaps to fill first
- **Refresh schedule**: How often to update pricing data, regulatory changes, market conditions
- **Quality thresholds**: What source authority level is required (regulatory > industry body > trade publication > blog)

The research strategy lives in the vertical orchestrator's configuration — either in the agent definition's `default_config` or in a dedicated `vertical_registry` table.

### 2.3 Content Patterns and Page Templates

Each vertical has its own set of page types that work in its niche:

- Veterinary: breed health profiles, procedure guides, practice directories, insurance comparison pages, cost transparency guides
- Energy: market analysis pages, contract comparison guides, supplier directories, benchmarking tools, regulatory compliance guides
- Mortgage: interactive calculators, affordability deep-dives, specialist situation guides, rate analysis, product comparison pages

These are not hard-coded page templates — they're prompt pattern libraries that the content writer agents receive as context alongside the RAG knowledge. The content writer agent code is the same across verticals; what changes is the knowledge and pattern context it receives.

### 2.4 Monetisation Configuration

Each vertical converts differently:

- Veterinary: pet insurance affiliate (£15-35/signup), practice listing fees (£30-100/month), lead generation for specialist referrals
- Energy: qualified lead generation (£30-60/qualified lead, £10-25/raw lead), B2B display advertising (£10-30 CPM), supplier directory fees
- Mortgage: broker lead generation (£50-150/lead), comparison affiliate, financial services display (£5-15 CPM), calculator engagement
- Seasonal/Gift: product affiliate (3-17% commission), display advertising (seasonal RPMs £15-30 in Nov/Dec)

Monetisation configuration tells the site planner what conversion elements to include (lead forms, affiliate link patterns, directory structures) and informs the content architecture (what pages serve as conversion endpoints vs traffic attractors).

---

## 3. Domain Strategies by Vertical

### 3.1 Veterinary Vertical — vetcomparison.uk

**Classification**: veterinary
**Primary visitor**: Pet owners choosing or switching vets
**Monetisation model**: Insurance affiliate + practice listings + specialist lead gen

**Knowledge clusters to build:**

**Cluster: Canine and Feline Biology (partially exists)**
Already built through the multi-cluster canine biology project. Covers anatomy, physiology, genetics, nutrition, behaviour. Needs indexing into the `knowledge_base` table with `collection: "veterinary"`.

**Cluster: Treatment and Procedure Knowledge**
Sources: BSAVA manuals, veterinary clinical guidelines, RCVS practice standards, published surgical outcome studies.
Produces: What common procedures involve, what good practice looks like, what questions to ask, what recovery looks like, what drives cost differences.

Content this enables:
- "What happens during dog neutering — the procedure, recovery, and what to look for in your vet"
- "Understanding dental cleanings for dogs — why prices vary from £200 to £800 and what you are paying for"
- "ACL surgery in dogs — TPLO vs lateral suture, success rates, and finding a surgeon"
- "Brachycephalic surgery: soft palate resection explained, anaesthesia risks, and choosing a vet"

**Cluster: Breed-Specific Health Profiles**
Sources: Kennel Club breed health surveys, breed club data, published genetic studies, insurance claims data (aggregated).
Produces: Comprehensive health timelines by breed — not just "Cavaliers get heart problems" but year-by-year health expectations, screening schedules, early warning signs, and what to look for in a vet for this breed.

Content this enables:
- "Labrador health guide: year-by-year, what to expect and what to screen for"
- "German Shepherd health: why hip scores matter and how to read them"
- "Dachshund back problems: IVDD explained, prevention, treatment options, and finding a specialist"
- "French Bulldog health: brachycephalic issues, spinal problems, and choosing the right vet"
- "Cavalier King Charles Spaniel: the first 10 years — health timeline and screening schedule"

**Cluster: Veterinary Cost Structures**
Sources: Aggregated fee survey data, insurance claims data, RCVS practice costs analysis, comparative cost studies.
Produces: Typical procedure costs with meaningful ranges, what drives price differences between practices (location, equipment, specialist qualifications, staffing ratios).

This is the single most requested piece of information that nobody provides. Vet pricing in the UK is opaque — the RCVS does not regulate fees and practices rarely publish them. Even approximate ranges with honest explanations of why prices vary would be more useful than anything currently available.

Content this enables:
- "Vet costs UK 2025: what to expect for common procedures"
- "Why does my vet charge so much? Understanding veterinary pricing"
- "Pet health budget planner: what to expect in your dog's first year"
- "Emergency vet costs: what out-of-hours care costs and why"

**Cluster: Practice Quality Indicators**
Sources: RCVS Practice Standards Scheme documentation, veterinary accreditation frameworks, clinical audit standards.
Produces: Objective quality indicators — RCVS accreditation levels (General Practice vs Veterinary Hospital vs Emergency Service), equipment standards, staffing ratios, registered veterinary nurses vs animal care assistants.

Content this enables:
- "RCVS Practice Standards: what the different levels mean for your pet's care"
- "Questions to ask when choosing a vet — beyond the Google reviews"
- "What to look for in an emergency vet — the equipment and staffing that matters at 2am"

**The synthesis advantage**: A page titled "Choosing a vet for your elderly Labrador" draws from breed health profiles (what conditions to expect), treatment/procedures (what screenings and treatments are involved), cost structures (what to budget), and practice quality (what capabilities to look for). No competitor can produce this because no competitor has the underlying knowledge infrastructure.

**Revenue projection:**

| Timeframe | Monthly Visits | Insurance Rev | Practice Listings | Lead Gen | Display | Total/Month |
|---|---|---|---|---|---|---|
| Months 1-6 | 500-2,000 | £50-350 | £0 | £0 | £5-40 | £55-390 |
| Months 6-12 | 2,000-8,000 | £250-1,400 | £200-500 | £100-300 | £40-160 | £590-2,360 |
| Months 12-18 | 8,000-25,000 | £1,000-4,375 | £500-2,000 | £300-1,000 | £160-500 | £1,960-7,875 |
| Months 18-24 | 25,000-60,000 | £3,125-10,500 | £2,000-5,000 | £1,000-3,000 | £500-1,200 | £6,625-19,700 |

Assumptions: 0.5% insurance conversion at £25 average per signup. Practice listings £50-100/month. Specialist referral leads £10-25.

### 3.2 Energy Vertical — gaswholesalers.com

**Classification**: energy_wholesale
**Primary visitor**: B2B buyers — facility managers, procurement officers, business owners
**Monetisation model**: Qualified lead generation + B2B display/sponsorship

**Knowledge clusters to build:**

**Cluster: Gas Market Structure and Pricing Dynamics**
Sources: Ofgem market reports, National Grid gas transmission data, ICE (Intercontinental Exchange) futures explanations, Cornwall Insight analysis, BEIS energy statistics, historical NBP pricing data.

Produces: How wholesale gas pricing actually works — National Balancing Point (NBP) pricing, forward curves, seasonal contracts, day-ahead vs month-ahead vs quarterly vs annual pricing. Why contract timing matters (signing a 3-year fixed at the wrong point in the price cycle costs 20-40% more). What contango and backwardation mean for procurement decisions. The relationship between wholesale and retail pricing.

Content this enables:
- "How UK gas pricing works: NBP, forward curves, and what they mean for your contract"
- "When to renew your business gas contract: reading the market signals"
- "Fixed vs flexible gas contracts: understanding what you are actually buying"
- "Wholesale gas price tracker: monthly analysis of NBP movements and what they mean for business buyers" (recurring-value hook — procurement managers return monthly)

**Cluster: Contract Structure Analysis**
Sources: Standard industry contract terms, Ofgem supplier licensing conditions, published contract comparison frameworks, Energy Ombudsman complaint data.

Produces: Analysis of the contractual elements that actually cause problems — consumption tolerance bands (out-of-tolerance charges can add thousands to annual costs), credit terms and payment structures, metering arrangements and estimation disputes, auto-rollover clauses, deemed rates (what you pay if your contract expires without renewal — often 50-100% above market rate), green gas certificate structures.

Content this enables:
- "Business gas contract terms explained: the 8 clauses that actually matter"
- "Out-of-tolerance charges: the hidden cost that can double your gas bill"
- "Deemed rates explained: why letting your gas contract expire costs a fortune"
- "Green gas for business: what the certificates mean and what you are actually paying for"

**Cluster: Supplier Differentiation and Analysis**
Sources: Ofgem supplier complaint data, Energy Ombudsman statistics, aggregated review analysis, company financial filings, published credit ratings.

Produces: Meaningful comparison on dimensions that affect the customer experience — billing accuracy and systems quality, customer service responsiveness, financial stability (supplier insolvency risk, relevant after 2021-2022 supplier failures), specialist capabilities for different business types, multi-site portfolio management.

Content this enables:
- "UK business gas suppliers compared: what the complaint data tells you"
- "Choosing a gas supplier for multi-site businesses: what to look for beyond price"
- "Supplier financial stability: how to check your gas provider will not go bust"
- "Energy broker vs direct: when a broker adds value and when they cost you money" (addresses broker conflict of interest — brokers earn 0.05p-2p per kWh commission)

**Cluster: Regulatory and Compliance Framework**
Sources: Ofgem regulatory guidance, BEIS publications, Environment Agency requirements.

Produces: Practical guidance on Climate Change Levy (CCL), Energy Savings Opportunity Scheme (ESOS), Streamlined Energy and Carbon Reporting (SECR), upcoming net zero obligations.

Content this enables:
- "Climate Change Levy on gas: what your business pays and how to check your invoice"
- "ESOS compliance guide: what your business needs to do and when"
- "Net zero and business gas: what upcoming regulation means for your energy strategy"

**Cluster: Consumption Benchmarking**
Sources: CIBSE Guide F energy benchmarks, Display Energy Certificate data (publicly available), published sector-specific energy use surveys, degree-day analysis methodology.

Produces: Benchmarking by building type, size, usage pattern, and climate zone. Powers a unique tool: "Enter your building type and floor area, and we'll tell you how your gas consumption compares to similar buildings."

Content this enables:
- "Business gas benchmarks: how much should your building use?"
- "Gas consumption calculator: compare your usage to similar businesses"
- "Reducing business gas costs: where the biggest savings are for each building type"

**The competitive positioning**: Current comparison sites treat gas as a commodity and compete on price alone. This site educates the buyer so they know what to compare, what to negotiate on, and when to buy. The transparency about broker commissions and contract traps positions it as the buyer's advocate, not another intermediary. Google's quality raters respond positively to this kind of structural honesty.

**Revenue projection:**

| Timeframe | Monthly Visits | Raw Leads | Qualified Leads | Lead Rev | Display/Sponsor | Total/Month |
|---|---|---|---|---|---|---|
| Months 1-6 | 200-1,000 | 5-15 | 2-5 | £100-425 | £0 | £100-425 |
| Months 6-12 | 1,000-4,000 | 15-50 | 5-20 | £325-1,450 | £100-300 | £425-1,750 |
| Months 12-18 | 4,000-12,000 | 50-150 | 20-60 | £950-4,350 | £300-1,000 | £1,250-5,350 |
| Months 18-24 | 12,000-30,000 | 150-400 | 60-150 | £2,400-10,500 | £1,000-3,000 | £3,400-13,500 |

Assumptions: 1.5% raw lead form submission, 0.5% qualified (more data requested). Raw leads £15, qualified £45. B2B display at premium RPMs after traffic established.

### 3.3 Finance/Mortgage Vertical — mortgagecalculator.co.uk

**Classification**: finance_mortgage
**Primary visitor**: People considering a mortgage or remortgage — first-time buyers, movers, remortgagers
**Monetisation model**: Broker lead generation + comparison affiliate + financial display ads

**Knowledge clusters to build:**

**Cluster: Interest Rate Dynamics and Swap Rates**
Sources: Bank of England base rate history and forward guidance, ICE sterling swap rate data, published analysis of swap rate/mortgage rate relationship, historical Moneyfacts/ICAEW data.

Produces: The single most important piece of mortgage knowledge no consumer site explains — how fixed mortgage rates are actually determined. Fixed rates track swap rates, not the base rate. When the base rate drops but swap rates stay high (as happened in late 2024), fixed rates don't fall as expected. Understanding this prevents the common mistake of waiting for "rates to come down" when the market has already priced in expected drops.

Content this enables:
- "How mortgage rates are actually set: the swap rate explained"
- "Monthly mortgage rate outlook: what swap rates tell us about where fixed rates are heading" (recurring-value hook — visitors return monthly)
- "2-year fix vs 5-year fix: what the yield curve says about which is better value right now"
- "Why mortgage rates did not fall when the base rate dropped — and when they might"

**Cluster: Lender Affordability Models and Criteria**
Sources: PRA stress testing requirements, FCA MCOB rules, published lender criteria documents, Building Societies Association data.

Produces: Why two people with the same income get offered different amounts. How stress testing works (lenders test at rates 3% above reversion). How different lenders treat self-employed income, bonus income, overtime, commission, rental income. What expenditure categories lenders assess.

Content this enables:
- "Why the bank says you can borrow £250,000 but your broker says £350,000"
- "Getting a mortgage when self-employed: which lenders are friendliest and what they need"
- "Mortgage affordability with bonus income: how 30 different lenders treat your bonus"
- "How much can I borrow? Beyond the basic calculators — what actually determines your maximum"

**Cluster: Mortgage Product Structures**
Sources: Moneyfacts product data analysis, FCA product governance rules, published analysis of fee structures, ERC data, lender product transfer policies.

Produces: Comparison beyond headline rates — total cost including fees, early repayment charge structures (flat vs declining), overpayment allowances and calculation basis, porting policies, product transfer options at deal end.

Content this enables:
- "The true cost of a mortgage: why comparing rates alone costs you money"
- "Early repayment charges explained: flat vs declining and what to watch for"
- "Best mortgages for overpaying: which lenders give you the most flexibility"
- "Product transfers explained: how to get a new deal without remortgaging"

**Cluster: Specialist Mortgage Situations**
Sources: Specialist lender criteria, building survey requirements, leasehold mortgage requirements, published guidance on non-standard properties.

Produces: Guidance for situations mainstream sites handle poorly — non-standard construction (steel frame, concrete prefab, thatched), short leases (under 80 years), new build restrictions, ex-local authority properties, credit history issues (missed payments, defaults, CCJs, IVAs, bankruptcy with specific timeframes).

Content this enables:
- "Getting a mortgage on a flat with a short lease: what lenders require"
- "Mortgages on non-standard construction: which lenders accept steel frame, concrete, timber"
- "Mortgage after a CCJ or default: which lenders will consider you and when"
- "New build mortgage restrictions: what developers and construction types cause problems"

**The long-tail strategy**: The big sites rank for "mortgage calculator" and won't be displaced. But thousands of specific, high-intent queries go underserved:

| Query | Monthly Searches (est.) | Knowledge Base Advantage |
|---|---|---|
| "Can I get a mortgage on 30k salary" | 3,000-5,000 | Specific affordability analysis by lender |
| "Mortgage on new build flat" | 1,000-2,000 | Specific lender restrictions and requirements |
| "Joint mortgage only one name on deeds" | 500-1,000 | Legal and lending implications explained |
| "Mortgage after debt management plan" | 500-1,000 | Specific lender criteria and timeframes |
| "How do mortgage rates work base rate" | 2,000-4,000 | Swap rate explanation nobody else gives |
| "Mortgage on 50k salary how much" | 2,000-3,000 | Stress-tested affordability by lender type |
| "Shared ownership mortgage calculator" | 1,000-2,000 | Specific calculator for staircasing scenarios |
| "Buy to let mortgage calculator interest only" | 1,000-2,000 | ICR calculation with actual lender stress rates |
| "Remortgage to release equity calculator" | 1,000-2,000 | LTV impact analysis with product suggestions |

Collectively these represent tens of thousands of monthly searches with very high commercial intent.

**Unique calculators the knowledge base enables:**

- **Enhanced repayment calculator**: Shows payments at deal rate PLUS at lender's SVR (what you revert to when the deal ends). Banks don't show this because it makes their products look worse.
- **True cost comparison**: Factors in product fees, cashback, free valuation/legal fees. A £999 fee on a £150,000 mortgage is 0.67% — roughly +0.3% to the effective rate over 2 years. On a £500,000 mortgage the same fee is only 0.2%, roughly +0.1%.
- **Overpayment impact modeller**: £100/month overpayment on £200,000 at 4.5% over 25 years saves approximately £20,000 in interest and cuts 4 years off the term.
- **Rate scenario modeller**: "If you take a 2-year fix at 4.2% and rates at renewal are 5.0%, here is what happens. If you take a 5-year fix at 4.6%, here is your total cost over 5 years."
- **Specialist affordability calculator**: "Enter your last 2 years of self-employed income and we will show you how different lender types would assess your affordability."

**Revenue projection:**

| Timeframe | Monthly Visits | Calculator Users (70%) | Form Fills (3%) | Revenue @ £50/lead | Display | Total/Month |
|---|---|---|---|---|---|---|
| Months 1-6 | 1,000-4,000 | 700-2,800 | 21-84 | £1,050-4,200 | £50-200 | £1,100-4,400 |
| Months 6-12 | 4,000-15,000 | 2,800-10,500 | 84-315 | £4,200-15,750 | £200-750 | £4,400-16,500 |
| Months 12-18 | 15,000-40,000 | 10,500-28,000 | 315-840 | £15,750-42,000 | £750-2,000 | £16,500-44,000 |
| Months 18-24 | 40,000-100,000 | 28,000-70,000 | 840-2,100 | £42,000-105,000 | £2,000-5,000 | £44,000-110,000 |

Mortgage is one of the highest-value consumer verticals. The exact-match .co.uk domain provides significant ranking advantage. These projections assume the calculators are genuinely useful tools that people share and return to.

### 3.4 Seasonal/Gift Vertical — xmaspresents.com

**Classification**: seasonal_gifts
**Primary visitor**: Gift shoppers (consumer, seasonal peak Oct-Dec)
**Monetisation model**: Product affiliate (Amazon 3-4%, John Lewis 5-7%, experience companies 10-17%) + seasonal display advertising

**Knowledge clusters to build:**

**Cluster: Gift Category Intelligence**
Sources: Retail trend reports, Google Shopping trends data, Amazon bestseller analysis, John Lewis/Not On The High Street category analysis.

Produces: What gift categories perform in which demographics, price point analysis by recipient type, trending products by season, gift satisfaction research.

Content this enables:
- Recipient-segmented guides: "Christmas presents for him/her/mum/dad/kids/teenagers/someone who has everything"
- Price-segmented guides: "Christmas presents under £10/£25/£50, luxury Christmas presents"
- Interest-segmented guides: "Christmas presents for gamers/gardeners/foodies/bookworms"
- Each updated annually with fresh product selections

**Cluster: Affiliate Programme Optimisation**
Sources: Affiliate network data, commission rate comparisons, conversion rate benchmarks by retailer.

Produces: Knowledge of which retailers to prioritise linking to. Amazon pays 3-4% on most categories but converts well. John Lewis pays 5-7%. Not On The High Street pays 6-10%. Experience companies (The Gift Experience) pay 10-17%. Higher-commission retailers should be preferred in guides where the products are comparable.

**The seasonal challenge and extension strategy**: "Christmas presents" searches peak at ~458,000 in December and drop to near zero January-September. Two approaches:

**Option A — Stay seasonal**: Build for Oct-Dec, accept 8-10 weeks of revenue. Update guides annually. Low maintenance for 9 months. Aim for £3,000-8,000 annual revenue, site value £5,000-15,000.

**Option B — Extend the brand** (recommended): Add year-round gift content — birthday presents, anniversary gifts, Valentine's Day, Mother's Day, Father's Day. Each has its own seasonal peak. Together they create a more even traffic profile. The .com domain works internationally, especially for the US market. This could double or triple total annual revenue.

**Revenue projection (Option B, extended brand):**

| Period | Monthly Visits | Clicks to Retailers (30%) | Purchases (5%) | Affiliate Rev | Display | Total/Month |
|---|---|---|---|---|---|---|
| Jan-Sep (avg) | 2,000-8,000 | 600-2,400 | 30-120 | £50-200 | £20-80 | £70-280 |
| Oct | 5,000-15,000 | 1,500-4,500 | 75-225 | £120-360 | £50-150 | £170-510 |
| Nov | 15,000-50,000 | 4,500-15,000 | 225-750 | £360-1,200 | £150-750 | £510-1,950 |
| Dec wk 1-3 | 20,000-60,000 | 6,000-18,000 | 300-900 | £480-1,440 | £200-900 | £680-2,340 |
| Annual total | — | — | — | £4,500-15,000 | £1,500-5,000 | £6,000-20,000 |
| Monthly average | — | — | — | £375-1,250 | £125-415 | £500-1,665 |

At 24-32x multiples on the monthly average: site value £12,000-53,000. The wide range reflects the difference between modest affiliate returns and building a genuinely popular gift guide site with established seasonal traffic.

### 3.5 Creative Services Vertical — design.co.uk

**Classification**: premium_domain (recommended: sell rather than develop)

`design.co.uk` is most valuable as a domain name sale rather than a content development project. The ambiguity that makes it hard to build content for (web design? graphic design? interior design? product design?) is exactly what makes it valuable as a brandable domain — it works for any design business.

A single-word generic .co.uk in a major commercial category is worth £20,000-100,000+ to the right buyer — a design agency wanting a prestigious address, a design platform entering the UK market, or a media company building a design publication.

Building a content site on it risks pigeon-holing the domain into one interpretation, reducing its appeal to buyers in other categories. A web design directory makes the domain less attractive to an interior design firm, and vice versa.

**If developing anyway** (to build value through traffic before selling): the web design agency directory interpretation scores highest — clear monetisation (web design leads at £50-200 per qualified lead, UK web design market ~£640 million), manageable scope, filterable by type/location/budget with lead generation forms. But the name-value sale likely exceeds 2-3 years of content site revenue.

**Pipeline recommendation**: Route `design.co.uk` to a "premium domain" pathway that produces a professional holding page with sale inquiry form, and lists it with domain brokers (Sedo, Dan.com, or UK specialists). The pipeline should include a "don't develop, sell the domain" output as a valid recommendation.

---

## 4. The Routing Architecture

### 4.1 Domain Intake and Classification

The main pipeline orchestrator receives a domain and must determine which vertical to route it to. This extends the existing site-classifier agent:

```
Domain arrives (intake orchestrator / briefing / API)
  → site-classifier produces vertical_slug alongside existing classifications
  → Main orchestrator looks up vertical orchestrator:
      SELECT type FROM agent_definitions
      WHERE tags @> '["vertical-orch"]'
      AND default_config->>'vertical_slug' = $1
  → Spawns vertical orchestrator, passes domain + briefing data
  → Vertical orchestrator takes full ownership from here
```

The site classifier needs a "vertical classification" output — not just "what kind of site is this" but "what knowledge vertical does this domain belong to." For domains that don't match any existing vertical, they route to a `generic` vertical that uses standard research and content patterns.

For ambiguous domains (like `design.co.uk`), the classifier can also output a `disposition` recommendation: `develop`, `sell_as_domain`, or `hold_for_review`.

### 4.2 The Vertical Registry

A lightweight registry maps vertical slugs to their configuration. This could be a database table or config in agent definitions. The registry stores:

```
vertical_slug:        "veterinary"
orchestrator_type:    "vertical-orch-veterinary"
knowledge_collection: "veterinary"
research_sources:     ["bsava", "rcvs", "kennel_club", "ofgem_complaints"]
content_patterns:     ["breed_profile", "procedure_guide", "practice_directory", "insurance_comparison", "cost_guide"]
monetisation_config:  {"primary": "insurance_affiliate", "secondary": ["practice_listings", "lead_gen"]}
maturity_stage:       "seeding" | "early" | "steady" | "specialised"
```

### 4.3 The Vertical Orchestrator Workflow

Each vertical orchestrator follows the same workflow skeleton, with vertical-specific configuration at each step:

```
receive_domain
  → load_vertical_knowledge        (rag_lookup from this vertical's collection)
  → assess_domain_fit              (does this domain match this vertical well? score 1-10)
  → commission_research            (spawn research agents targeting vertical-specific sources)
  → build_content_architecture     (using vertical-specific page patterns and conversion paths)
  → generate_content               (content writers receive vertical knowledge as RAG context)
  → build_tools                    (if vertical needs interactive tools — calculators, directories)
  → configure_monetisation         (set up affiliate links, lead forms, directory structures)
  → deploy
```

The key difference from the current site-work-orchestrator is in steps 2-4. Currently the site planner does all of this with generic prompts. With vertical orchestrators, each step is informed by accumulated vertical knowledge.

---

## 5. Separating Research Clusters from Build Clusters

This is where things get architecturally interesting. Research and build have fundamentally different characteristics:

### 5.1 Why Separate Them

**Research is messy.** It involves web scraping that may fail, rate-limited API calls, PDFs that need parsing, data that needs cleaning, and LLM calls that need to extract structured knowledge from unstructured sources. Research agents might need to retry failed scrapes, handle CAPTCHAs or blocks, process documents in multiple formats, and deal with sources that change their structure. Research produces variable-quality output that needs validation before it enters the knowledge base.

**Build is structured.** Once the knowledge base is populated and the content architecture is planned, the build pipeline is predictable: generate content from knowledge + prompts, assemble pages, apply design, deploy to git. The same patterns repeat reliably. Build agents have clear inputs and outputs, predictable resource usage, and straightforward error handling.

**Research is slow and exploratory. Build is fast and deterministic.** A research agent investigating BSAVA guidelines might spend 30 minutes scraping, parsing, chunking, and indexing. A content writer agent produces a page in 30 seconds. Mixing these in the same workflow creates scheduling problems — the build pipeline waits for research that may take hours, while research agents sit idle during the build phase.

**Research is shared. Build is per-site.** Research findings go into the vertical's knowledge base and benefit all future sites. Build output goes to a specific site. Research should run independently of any particular site build — it's a vertical-level concern, not a site-level concern.

**Research needs different infrastructure.** Web scraping needs external network access, potentially proxy rotation, and the webscrape adapter. PDF processing needs document parsing libraries. LLM extraction needs larger context windows and more compute. Build agents mostly need the git adapter, the image generator, and standard LLM calls for content. Separating them allows different resource profiles.

### 5.2 The Two-Cluster Model

```
┌─────────────────────────────────────────────────────┐
│                 RESEARCH CLUSTER                      │
│                                                       │
│  ┌─────────────┐   ┌─────────────┐   ┌────────────┐ │
│  │  Research    │   │  Research    │   │  Research   │ │
│  │  Orch:      │   │  Orch:      │   │  Orch:     │ │
│  │  Veterinary │   │  Energy     │   │  Mortgage  │ │
│  └──────┬──────┘   └──────┬──────┘   └─────┬──────┘ │
│         │                  │                 │        │
│  ┌──────▼──────────────────▼─────────────────▼─────┐ │
│  │              Research Agent Pool                  │ │
│  │  - Web scraper agents                            │ │
│  │  - PDF/document parser agents                    │ │
│  │  - LLM extraction agents (large context)         │ │
│  │  - Knowledge indexer agents (rag_index)           │ │
│  │  - Source validator agents                        │ │
│  └──────────────────────┬──────────────────────────┘ │
│                         │                             │
│  ┌──────────────────────▼──────────────────────────┐ │
│  │           Knowledge Base (Postgres)              │ │
│  │  collection: "veterinary"                        │ │
│  │  collection: "energy_wholesale"                  │ │
│  │  collection: "finance_mortgage"                  │ │
│  │  collection: "seasonal_gifts"                    │ │
│  └──────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────┘
                          │
                    Kafka (shared)
                          │
┌─────────────────────────────────────────────────────┐
│                  BUILD CLUSTER                        │
│                                                       │
│  ┌──────────────┐                                    │
│  │ Main Pipeline │──── Domain intake + classification │
│  │ Orchestrator  │                                    │
│  └───────┬──────┘                                    │
│          │                                            │
│  ┌───────▼───────────────────────────────────────┐   │
│  │          Vertical Build Orchestrators           │   │
│  │  - vertical-build-veterinary                    │   │
│  │  - vertical-build-energy                        │   │
│  │  - vertical-build-mortgage                      │   │
│  │  - vertical-build-generic                       │   │
│  └───────┬───────────────────────────────────────┘   │
│          │                                            │
│  ┌───────▼───────────────────────────────────────┐   │
│  │              Build Agent Pool                   │   │
│  │  - Site planner (uses rag_lookup for knowledge) │   │
│  │  - Content writers (receive RAG context)         │   │
│  │  - HTML developers                               │   │
│  │  - Design agents (CSS, layout)                   │   │
│  │  - Tool builders (calculators, directories)      │   │
│  │  - Asset deployers (images, files)               │   │
│  │  - Site publishers (git commit/deploy)           │   │
│  └───────────────────────────────────────────────┘   │
│                                                       │
│  ┌───────────────────────────────────────────────┐   │
│  │          Maintenance Agent Pool                 │   │
│  │  - Discovery agents (content, links, SEO)       │   │
│  │  - Fix agents (section rewriter, nav updater)   │   │
│  │  - Knowledge refresh triggers                    │   │
│  └───────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

### 5.3 How the Two Clusters Interact

The research cluster and build cluster communicate through two mechanisms:

**1. Shared knowledge base (Postgres).** Both clusters connect to the same database. Research agents write to `knowledge_base` via `rag_index`. Build agents read from `knowledge_base` via `rag_lookup`. No Kafka messages needed for knowledge transfer — it's just database reads and writes.

**2. Kafka for orchestration coordination.** When the build cluster needs research that doesn't exist yet, the build orchestrator dispatches a research request via Kafka to the research cluster. The research orchestrator runs the research, indexes the results, and responds. The build orchestrator then proceeds with `rag_lookup` to access the newly indexed knowledge.

The workflow interaction:

```
BUILD CLUSTER                              RESEARCH CLUSTER
─────────────                              ────────────────
Domain arrives
  → classify to vertical
  → rag_lookup: check knowledge exists
  │
  ├─ Knowledge sufficient?
  │   YES → proceed to content planning
  │   NO  → dispatch_agent to research cluster
  │           ├─ research orch receives request
  │           ├─ spawns scraper/parser agents
  │           ├─ validates and indexes findings
  │           └─ responds "research complete"
  │         ← response received
  │         → rag_lookup (knowledge now available)
  │         → proceed to content planning
  │
  → plan content architecture (with RAG context)
  → generate content (with RAG context)
  → build tools
  → deploy
```

### 5.4 The Research Orchestrator

Each vertical gets a research orchestrator that lives in the research cluster. It manages:

**Seeding research**: Initial knowledge base population when a vertical is first created. Processes the foundational authoritative sources. This runs once (or periodically for major refreshes).

**On-demand research**: Triggered by the build cluster when it encounters a knowledge gap. For example, the build cluster is processing a domain about dachshund health and finds no IVDD knowledge in the veterinary collection. It dispatches a research request, the research orchestrator processes IVDD-specific sources, indexes the findings, and the build can proceed.

**Scheduled maintenance research**: Periodic refresh of time-sensitive knowledge. Gas pricing data needs monthly updates. Swap rate analysis needs weekly updates. Breed health surveys update annually. Each vertical's research orchestrator has a refresh schedule appropriate to its knowledge domains.

**Quality validation**: Research output goes through a validation step before entering the knowledge base. A validator agent (or LLM step) checks that extracted knowledge is coherent, doesn't contradict existing knowledge, and meets a minimum quality threshold. Low-confidence extractions get flagged for review rather than indexed directly.

The research orchestrator workflow:

```
receive_research_request
  → identify_knowledge_gaps      (what does the vertical need that it doesn't have?)
  → prioritise_sources           (which authoritative sources to consult first?)
  → dispatch_scrapers            (spawn web scraper agents for each source)
  → collect_raw_content          (wait for scraper responses, handle failures)
  → extract_knowledge            (LLM extraction: raw content → structured knowledge chunks)
  → validate_quality             (check coherence, accuracy, source authority)
  → index_knowledge              (rag_index to knowledge_base with collection tag)
  → update_research_metadata     (record what was researched, when, from what sources)
  → respond_to_requester         (tell build cluster research is complete)
```

### 5.5 Research Agent Types

The research cluster runs specialised agents that the build cluster doesn't need:

**Source scraper agents**: Extended versions of the existing webscrape adapter. Can handle different source types — HTML pages, PDFs, data tables, API responses. Rate-limited per source domain. Retry logic for transient failures. Source-specific parsing rules (Ofgem data tables have different structure from BSAVA guidelines).

**Document parser agents**: Process PDFs, Word documents, and structured data files that research discovers. Extract text, tables, and structured data. Handle OCR for scanned documents if needed. Output is clean text chunks ready for knowledge extraction.

**Knowledge extraction agents**: LLM-powered agents that take raw scraped/parsed content and extract structured knowledge. These need larger context windows (to process full documents) and more careful prompting (to extract facts rather than opinions, to identify source authority, to flag uncertainty). They produce knowledge chunks tagged with source, confidence, and topic.

**Source validator agents**: Check extracted knowledge against existing knowledge base for contradictions. Flag low-confidence extractions. Verify that source URLs are still accessible. Score source authority (regulatory > industry body > trade publication > blog > forum).

**Knowledge indexer agents**: The existing `rag_index` action, but orchestrated in batch — processing multiple chunks with dedup checking, embedding generation, and collection tagging. Could batch embed operations for efficiency.

### 5.6 Why This Separation Matters

**Independent scaling.** Research is bursty — seeding a new vertical might process hundreds of documents in a day, then nothing for weeks. Build is more steady — sites are built at a consistent rate. Separating them means the research cluster can scale up during seeding and scale down during quiet periods without affecting build throughput.

**Independent failure.** If a web scraper fails because a source site is down, that shouldn't block site builds. The build cluster uses whatever knowledge already exists. Research failures are retried independently.

**Clean logging and debugging.** Research agent logs are about source quality, extraction accuracy, and knowledge gaps. Build agent logs are about content quality, deployment success, and site structure. Mixing them makes both harder to debug. Separate clusters means separate log streams.

**Different security profiles.** Research agents need external network access (web scraping). Build agents mostly need internal access (database, git adapter, image generator). Separating them allows tighter network policies on the build cluster.

**Cost allocation.** Research is an investment in the vertical — its cost should be amortised across all domains in that vertical. Build cost is per-site. Separate clusters make cost attribution straightforward.

### 5.7 Implementation Path

**Phase 1: Logical separation on shared infrastructure.**
Research orchestrators and build orchestrators are separate agent definitions running on the same K8s cluster. They share Kafka and Postgres. The knowledge base is the shared state. Research and build run as separate orchestrations that happen to be on the same machine.

This requires:
- New agent definitions for vertical research orchestrators
- New agent definitions for vertical build orchestrators
- Extending the site classifier to output `vertical_slug`
- A routing step in the main pipeline orchestrator
- Seeding the first vertical knowledge bases

No new Go code — just SQL agent definitions with workflows that use existing actions (`rag_lookup`, `rag_index`, `spawn_agent`, `call_agent`, `webscrape`, `execute_llm_prompt`).

**Phase 2: Physical separation.**
When a vertical's research workload justifies its own cluster (likely when processing hundreds of sources or when web scraping volume triggers rate limiting concerns), the research orchestrators migrate to a dedicated cluster using `dispatch_agent`. The build orchestrators stay on the main cluster or get their own.

This requires:
- Deploying a second K8s cluster (or using the existing multi-cluster dispatch infrastructure)
- Changing `spawn_agent` to `dispatch_agent` for research requests
- Ensuring the remote cluster has webscrape adapter and sufficient compute

**Phase 3: Vertical-specific optimisation.**
Fine-tuned local models per vertical (e.g., a veterinary knowledge extraction model, an energy contract analysis model). Vertical-specific scraping infrastructure with source-aware parsers. Dedicated embedding models optimised for each vertical's terminology.

---

## 6. Knowledge Accumulation Loop

The compounding advantage of the vertical architecture comes from the knowledge accumulation loop:

```
New domain arrives in vertical
  → rag_lookup: what knowledge exists?
  → Identify gaps specific to this domain
  → Commission targeted research for gaps only
  → New knowledge indexed → benefits this domain AND all future domains
  → Build site using full accumulated knowledge
  → Maintenance agents later detect knowledge staleness
  → Knowledge refresh research commissioned
  → Updated knowledge benefits ALL sites in vertical
```

The first domain in a vertical is expensive — it triggers foundational research. The second domain is cheaper — most knowledge already exists, only gap-filling needed. By the tenth domain, the vertical has comprehensive knowledge and new domains are built almost entirely from existing knowledge with minimal incremental research.

This creates a defensible moat. A competitor copying one site's content doesn't get the knowledge base. They'd need to replicate the entire research investment to produce content of comparable depth across multiple sites.

**Metrics to track per vertical:**

- Knowledge base size (chunks indexed per collection)
- Knowledge coverage (what percentage of the vertical's topic map is covered)
- Knowledge freshness (average age of chunks, staleness rate)
- Research ROI (chunks produced per research hour, domains benefiting per chunk)
- Content quality correlation (do sites with more RAG context rank higher?)

---

## 7. The Full Hierarchy

```
Main Pipeline Orchestrator (build cluster)
  ├── Domain Intake + Classification
  │     └── Vertical classifier: "What vertical? What disposition?"
  │
  ├── Vertical: Veterinary
  │     ├── Research Orch (research cluster)
  │     │     ├── Sources: BSAVA, RCVS, Kennel Club, breed clubs, insurance data
  │     │     ├── Refresh: breed surveys annually, procedure costs quarterly, regulatory as-needed
  │     │     └── Agents: source scrapers, document parsers, knowledge extractors
  │     ├── Knowledge Base: collection "veterinary"
  │     │     ├── Canine/feline biology (existing)
  │     │     ├── Treatment and procedure knowledge
  │     │     ├── Breed-specific health profiles
  │     │     ├── Veterinary cost structures
  │     │     └── Practice quality indicators
  │     ├── Build Orch (build cluster)
  │     │     ├── Page patterns: breed profiles, procedure guides, directories, insurance comparison
  │     │     ├── Tools: vet finder, cost estimator, breed health checker
  │     │     └── Monetisation: insurance affiliate, practice listings, lead gen
  │     └── Domains served: vetcomparison.uk, [future pet/vet domains]
  │
  ├── Vertical: Energy/Utilities
  │     ├── Research Orch (research cluster)
  │     │     ├── Sources: Ofgem, National Grid, ICE, Cornwall Insight, BEIS, CIBSE
  │     │     ├── Refresh: pricing monthly, regulatory quarterly, benchmarks annually
  │     │     └── Agents: market data scrapers, regulatory document parsers, contract analysers
  │     ├── Knowledge Base: collection "energy_wholesale"
  │     │     ├── Gas market structure and pricing dynamics
  │     │     ├── Contract structure analysis
  │     │     ├── Supplier differentiation data
  │     │     ├── Regulatory and compliance framework
  │     │     └── Consumption benchmarking data
  │     ├── Build Orch (build cluster)
  │     │     ├── Page patterns: market analysis, contract guides, supplier directories, benchmarking
  │     │     ├── Tools: consumption benchmarker, contract cost calculator, price tracker
  │     │     └── Monetisation: qualified leads, B2B display, supplier directory fees
  │     └── Domains served: gaswholesalers.com, [future energy/utility domains]
  │
  ├── Vertical: Finance/Mortgage
  │     ├── Research Orch (research cluster)
  │     │     ├── Sources: PRA, FCA, BoE, Moneyfacts, ICE swap rates, lender criteria docs
  │     │     ├── Refresh: swap rates weekly, lender criteria monthly, regulatory as-needed
  │     │     └── Agents: financial data scrapers, regulatory document parsers, lender criteria extractors
  │     ├── Knowledge Base: collection "finance_mortgage"
  │     │     ├── Interest rate dynamics and swap rates
  │     │     ├── Lender affordability models and criteria
  │     │     ├── Mortgage product structures
  │     │     └── Specialist mortgage situations
  │     ├── Build Orch (build cluster)
  │     │     ├── Page patterns: calculators, affordability guides, specialist situations, rate analysis
  │     │     ├── Tools: repayment calc, true cost calc, overpayment modeller, scenario modeller, specialist affordability
  │     │     └── Monetisation: broker leads (£50-150), comparison affiliate, financial display
  │     └── Domains served: mortgagecalculator.co.uk, [future finance domains]
  │
  ├── Vertical: Seasonal/Gifts
  │     ├── Research Orch (research cluster)
  │     │     ├── Sources: retail trend reports, Google Shopping data, affiliate programme data
  │     │     ├── Refresh: product trends quarterly, affiliate rates annually, seasonal prep in September
  │     │     └── Agents: product scrapers, trend analysers, affiliate programme monitors
  │     ├── Knowledge Base: collection "seasonal_gifts"
  │     │     ├── Gift category intelligence
  │     │     └── Affiliate programme optimisation data
  │     ├── Build Orch (build cluster)
  │     │     ├── Page patterns: gift guides by recipient/price/interest, seasonal landing pages
  │     │     ├── Tools: gift finder quiz, budget calculator
  │     │     └── Monetisation: product affiliate (3-17%), seasonal display (£15-30 RPM peak)
  │     └── Domains served: xmaspresents.com, [future gift/seasonal domains]
  │
  ├── Vertical: Generic
  │     ├── Research Orch: competitor analysis, keyword research only
  │     ├── Knowledge Base: collection "generic" (thin, per-domain research)
  │     ├── Build Orch: standard business site templates
  │     └── Domains served: unclassified domains, low-value niches
  │
  └── Premium Domain Pathway
        └── Domains like design.co.uk → professional holding page + broker listing
```

---

## 8. What Needs Building

### Already exists (reuse):

- `knowledge_base` table with collection support and pgvector
- `rag_lookup` and `rag_index` actions
- Multi-cluster dispatch via `dispatch_agent` and remote job spawner
- Agent group definitions and spawn/call orchestration
- `vertical_slug` concept in multiple places (vet collection pipeline, agent definitions)
- Site classifier agent
- Web scrape adapter
- Content writer agents
- Site work orchestrator with dispatch loop
- Git adapter for deployment

### Needs extending:

- **Site classifier**: Add `vertical_slug` and `disposition` to output. Extend prompt to classify domains into verticals and recommend develop/sell/hold.
- **`rag_index` action**: Add `source_authority` field (1-5 scale) so knowledge from BSAVA ranks higher than knowledge from a blog. Add `source_url` and `source_date` for provenance tracking.
- **`rag_lookup` action**: Add optional `min_authority` filter so build agents can request only high-confidence knowledge.
- **Research agents**: Extend existing webscrape action with retry logic, rate limiting per domain, and PDF handling.

### Needs creating (agent definitions only, no new Go code):

- **Vertical research orchestrator definitions** — one per vertical. Each has a workflow that knows what sources to consult, how to process them, and where to index results.
- **Vertical build orchestrator definitions** — one per vertical. Each has a workflow that loads vertical knowledge, plans content architecture, and orchestrates build agents with appropriate RAG context and page patterns.
- **Vertical registry** — database table mapping vertical_slug to orchestrator types, knowledge collections, research sources, and configuration. Or simpler: a tag convention in agent_definitions plus `default_config` fields.
- **Knowledge seeding workflows** — initial research programmes for each vertical. Run once to populate foundational knowledge before processing any domains.
- **Knowledge refresh scheduler** — a maintenance agent that triggers periodic research refreshes per vertical based on the vertical's refresh schedule.

### Schema additions:

```sql
-- Extend knowledge_base with source provenance (if not already present)
ALTER TABLE knowledge_base ADD COLUMN IF NOT EXISTS source_authority integer DEFAULT 3;
ALTER TABLE knowledge_base ADD COLUMN IF NOT EXISTS source_url text;
ALTER TABLE knowledge_base ADD COLUMN IF NOT EXISTS source_date timestamp with time zone;
ALTER TABLE knowledge_base ADD COLUMN IF NOT EXISTS vertical_slug text;
ALTER TABLE knowledge_base ADD COLUMN IF NOT EXISTS knowledge_type text; -- 'factual', 'procedural', 'pricing', 'regulatory'

CREATE INDEX IF NOT EXISTS idx_kb_vertical ON knowledge_base(vertical_slug);
CREATE INDEX IF NOT EXISTS idx_kb_authority ON knowledge_base(source_authority DESC);

-- Vertical registry
CREATE TABLE IF NOT EXISTS vertical_registry (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    vertical_slug text UNIQUE NOT NULL,
    display_name text NOT NULL,
    description text,
    research_orch_type text NOT NULL,  -- agent_definitions.type for research orchestrator
    build_orch_type text NOT NULL,     -- agent_definitions.type for build orchestrator
    knowledge_collection text NOT NULL, -- knowledge_base.collection value
    research_sources jsonb DEFAULT '[]',  -- list of source configurations
    content_patterns jsonb DEFAULT '[]',  -- list of page type patterns
    monetisation_config jsonb DEFAULT '{}',
    refresh_schedule jsonb DEFAULT '{}', -- per-knowledge-type refresh intervals
    maturity_stage text DEFAULT 'seeding', -- seeding, early, steady, specialised
    domain_count integer DEFAULT 0,
    knowledge_chunk_count integer DEFAULT 0,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);
```

---

## 9. Implementation Order

1. **Add vertical_slug to site classifier output.** Small prompt change. Test with known domains.

2. **Create vertical_registry table and seed initial verticals.** Veterinary first (existing knowledge base gives head start), then energy and mortgage.

3. **Extend knowledge_base schema** with source_authority, source_url, vertical_slug columns.

4. **Index existing canine biology knowledge** into knowledge_base with `collection: "veterinary"`. This gives the veterinary vertical immediate depth.

5. **Create first vertical research orchestrator** (veterinary). Workflow: receive research request → identify sources → dispatch scrapers → extract knowledge → validate → index. Test with a research request for "breed health profiles".

6. **Create first vertical build orchestrator** (veterinary). Workflow: receive domain → rag_lookup → plan content architecture → generate content with RAG context → build → deploy. Test with vetcomparison.uk.

7. **Add routing to main pipeline orchestrator.** Classify domain → look up vertical → spawn vertical build orchestrator. Domains without a matching vertical go to generic.

8. **Create energy and mortgage vertical orchestrators.** Follow same pattern as veterinary. Commission seeding research for each.

9. **Separate research cluster** when research workload justifies physical separation. Change spawn_agent to dispatch_agent for research requests.

10. **Add knowledge refresh scheduling** per vertical. Maintenance agents trigger periodic research based on vertical's refresh schedule.

Each step is independently testable and doesn't break existing functionality. The generic/existing pipeline continues to work for all domains until they're explicitly routed to a vertical.