# idea.uk — v2 method test run (four domains)

Second test of the method, with the richer generation step (multi-lens) and v2
patches (audience-fit challenge; seller-bundles-support check; richer capability
menu). Same four scoring factors plus Durability. Four domains, deliberately
pushing generation harder before drawing conclusions.

For each domain, I flag which candidates are **[new in v2]** (surfaced by the
new lenses) versus **[would have appeared in v0 too]** — that's the real test of
whether the v2 generation actually changes the output.

---

## Run 5 — websitedesign.com (v2)

### Audience (with audience-fit challenge)

The v0 run defaulted to "founders prototyping with AI builders." The audience-fit
challenge surfaces three more:

- **A.** Founders/non-technical prototypers (the v0 audience; Bolt/Lovable users).
- **B.** Existing-site owners who want their current site improved (different
  problem: not "build me one" but "fix what I have"). Likely larger market.
- **C.** Agencies/freelance designers serving clients (B2B, recurring spend,
  professional users).
- **D.** Small businesses needing compliance (accessibility, GDPR) — most
  AI-built sites are non-compliant out of the box.

Most interesting from the willingness-to-pay angle: **C** (agencies have real
budgets and recurring needs) and **D** (compliance is paid-for today, audits run
£hundreds–thousands).

### Generate — multi-lens

**Demand lens:**
- Generic AI-aesthetic — every Lovable/Bolt site looks similar. Audience wants
  on-brand sites without paying for a full design agency.
- Conversion optimisation — most don't know how to make a site actually convert.
- Compliance — WCAG 2.2, GDPR cookie consent, accessibility statements.
  Audience doesn't know the rules.
- Maintenance/improvement — sites get built and never improved. The audience
  wants ongoing value, not a one-shot build.

**Generalist-failure lens:**
- ChatGPT can't actually visit your site and audit it.
- Confident-wrong on accessibility specifics (WCAG version differences,
  jurisdictional rules).
- Stateless — no memory of your brand, prior decisions, or what's been tried.
- No actual computation (can't run Lighthouse, can't measure performance).
- Generic copy that doesn't sound like a brand.

**Frontier lens:**
- Agentic browsing → an agent that *uses* the site, identifies friction.
- Vision precision → screenshot critique, accessibility audit from rendered output.
- Code execution → real Lighthouse / axe-core / WAVE audits.
- Long-context → load whole brand guidelines + competitor sites at once.
- Multi-modal → analyse video of user testing.

**Outcome lens:**
- "My site is built, on-brand, converts well, compliant, and continually improves."

**Asset × capability sweep (v0 step):**
- Site spec/plan × action → orchestration layer (the Run 1 candidate).

### Candidates (after dedup)

1. **Compliance-aware site builder + audit agent** [new in v2] — WCAG 2.2, GDPR,
   cookie consent, accessibility statement, automated audit via code execution
   (axe-core/Lighthouse). Generalist failure: confident-wrong on compliance
   specifics, can't actually run the audit.
2. **CRO audit agent** [new in v2] — agentic browsing actually uses the user's
   live site, identifies friction with reasoning, proposes specific changes.
   Generalist failure: can't visit, only describes.
3. **Performance optimisation agent** [new in v2] — runs real Lighthouse via
   code execution, generates fixes, validates. Generalist failure: approximates,
   doesn't actually measure.
4. **Brand-consistent imagery suite** [new in v2] — load brand reference, produce
   all site imagery in that consistent style; image-editing precision over generic
   gen. Generalist failure: every site ends up looking the same.
5. **Continuous improvement subscription** [new in v2] — ongoing audits, weekly
   suggestions, persistent memory of brand and prior decisions. Generalist
   failure: stateless.
6. **Orchestration layer over AI builders** [v0 carryover, 1b] — spec → run →
   critique → iterate.

### Cut

- **#4 imagery**: image editing is improving fast in generalist models; brand
  consistency claim is real but durability is low. Defensibility 3 if we have a
  strong brand-extraction process, otherwise 2. **Keep, mark low-Durability.**
- **#6 orchestration**: previous run failed it on Willingness (paying for a prompt
  is hard) and Durability (builders' own planning improves). Same issue.
  **Drop unless reframed; covered in Run 1.**
- Seller-bundles-free check: Bolt/Lovable themselves are improving their own
  planning steps; that's a free substitute for some of #6 but *not* for #1–#3,
  #5 — those are things the builders don't do.

### Verify

- **Compliance market is real:** WCAG audits commonly £500–£3,000 one-off;
  ongoing compliance monitoring services exist (Siteimprove, AccessiBe, axe DevTools)
  at SaaS prices £50–500/month. So the *value* of compliance tooling is
  established. The competition exists but is enterprise-priced — there's room
  for a more accessible AI-driven version.
- **Continuous improvement (#5)** competes with site monitoring tools (Hotjar,
  FullStory) for analytics + ContentKing for SEO monitoring — these are paid
  SaaS, established.

### Score

```
1. Compliance-aware builder + audit agent      [new in v2]
   Defensibility 4 (regulatory specificity + verified computation hard to fake
        for a generalist), Willingness 4 (paid market exists), Buildability 3
        (axe-core/Lighthouse via code exec is moderate), Reuse 4 (every site
        needs this), Durability 4 (regulations stable, frontier tooling not yet
        in chat-builders).
   Sum 19. ADVANCES. [test now via fake-door: "free compliance audit, pay to fix"]

2. CRO audit agent                              [new in v2]
   Defensibility 4 (agentic-browsing actually-using-the-site is hard for
        generalists right now), Willingness 4 (CRO consultants charge £k+/audit),
        Buildability 2 (needs agentic browsing + reasoning + brand context),
        Reuse 4, Durability 3 (agentic capabilities will commoditise).
   Sum 17. ADVANCES. [consider — expensive build, fake-door demand test first]

3. Performance optimisation agent               [new in v2]
   Defensibility 3 (Lighthouse-via-code-exec is doable but not exclusive),
        Willingness 3 (paid tools exist but not all owners care), Buildability 4
        (real Lighthouse is straightforward), Reuse 4, Durability 2 (will
        commoditise quickly as chat-tools add code exec).
   Sum 16. ADVANCES with Durability flag (short-lived).

5. Continuous improvement subscription          [new in v2]
   Defensibility 3 (persistent memory is becoming standard; the asset is the
        process and accumulated profile), Willingness 4 (recurring is what
        agencies and SaaS already pay for), Buildability 2 (needs stack
        integration + memory + scheduled runs), Reuse 4, Durability 3.
   Sum 16. ADVANCES.
```

### Read

**v2 produced four advancing candidates from one domain, where v0 produced
zero advancing on this audience (and one expensive consider).** All four lean on
specialism the v0 generation didn't reach for: regulatory specificity (#1),
agentic action (#2), verified computation (#3), persistent memory (#5).
**#1 is the strongest** — clear regulatory moat, established paid market,
buildable now. Worth a fake-door demand test next.

---

## Run 6 — gaswholesalers.com (v2)

### Audience (with audience-fit challenge)

v0 ran high-paid traders → failed (over-served by Bloomberg). Audience-fit
surfaces:

- **A.** Mid-market energy procurement managers (manufacturers, hospitality,
  transport, agriculture) — buy gas/oil/electricity but lack Bloomberg.
- **B.** Energy consultants/brokers (B2B; might use a tool).
- **C.** Industrial users with energy-intensive ops (e.g. greenhouses, cold
  storage).
- **D.** Small/regional gas wholesalers themselves (the literal "wholesalers"
  audience the domain name implies).

**A** is the underserved middle. **D** is interesting because they're the
domain's literal audience and have real margin in their business.

### Generate — multi-lens

**Demand lens (audience A — mid-market procurement):**
- Knowing when to buy vs wait.
- Comparing supplier quotes (different units, terms, hedging structures).
- Hedging strategy without trader expertise.
- Budget forecasting under price uncertainty.
- Validating broker recommendations (brokers earn commission — misaligned).

**Generalist-failure lens:**
- Stale on actual prices; no live data.
- Confident-wrong on hedging maths.
- Generic on contract terms (no industry-specific knowledge).
- Can't access supplier quotes.
- No persistent context of this buyer's profile.

**Frontier lens:**
- Reasoning models → real hedging/risk math.
- Long-context → load supplier contracts in full.
- Agentic → fill RFQs, compare quotes structurally.
- Code execution → run scenario simulations.
- Real-time data → cheap APIs ($9–$129/mo).
- Persistent memory → track this buyer's history, preferences, risk profile.

**Outcome lens:**
- "Our energy buying is as good as a big trader's, without hiring one."

### Candidates (after dedup)

1. **Aligned-incentive procurement advisor** [new in v2] — flat-fee subscription
   alternative to commission-based brokers. Personalised recommendations
   grounded in this farm/factory's consumption profile + live prices. Generalist
   failure: no live data, no persistent profile, no aligned-incentive framing.
2. **RFQ agent** [new in v2] — agent fills supplier RFQs, normalises quotes
   (different units, payment terms, hedge structures), compares apples-to-apples,
   ranks. Generalist failure: can't act.
3. **Hedge strategy assistant** [new in v2] — reasoning-model-driven hedging
   recommendations matched to the buyer's risk profile, with actual computation
   of payoff scenarios. Generalist failure: confident-wrong on financial maths.
4. **Contract intelligence** [new in v2] — load supplier contract, flag
   unfavourable terms vs benchmarks, propose negotiation points. Long-context +
   domain knowledge. Generalist failure: stale on contract patterns.
5. **Energy risk monitor** [new in v2] — ongoing alerts on price moves, supplier
   credit risk, geopolitical events affecting supply. Real-time + monitoring.
   Generalist failure: not real-time, not stateful.
6. **Budget forecaster** [new in v2] — load consumption history, market data,
   produce uncertainty-banded forecasts for procurement planning. Verified
   computation. Generalist failure: can't run the stats.

(Candidate "buy a proprietary feed" — v0's strong-looking option — still drops on
verification (cost, licensing). Not regenerated.)

### Cut

- All six are aimed at the mid-market procurement audience that *doesn't* have
  Bloomberg. Free substitute: brokers (commission-based, conflicted), DIY
  spreadsheets, or growing into a paid Bloomberg/Refinitiv seat later.
- Seller-bundles-free check: brokers give advice "free" but their incentives are
  misaligned (they earn on the deal). That's an *opening* for an aligned-incentive
  alternative, not a barrier. Important: this is a case where the seller-free
  pattern actually *creates* an opportunity rather than blocking one.

### Verify

- Mid-market procurement spend is substantial: a mid-sized factory might spend
  £100k–£1m/year on energy. A 2–5% optimisation is £2k–£50k. So willingness to
  pay for a tool that demonstrably saves that is real.
- Aligned-incentive framing: subscription-fee energy advisors do exist in the
  UK (firms like Inenco, Cornwall Insight, etc. for larger clients). The
  mid-market gap is genuine.
- API access ($9–$129/mo) confirmed in Run 3 verification.

### Score

```
1. Aligned-incentive procurement advisor       [new in v2]
   Defensibility 4 (curated procurement-context reasoning + aligned-incentive
        positioning is a real differentiator vs free conflicted brokers),
        Willingness 4 (real spend, real savings to fund a £20–100/mo sub),
        Buildability 2 (needs domain knowledge + live data + persistent profile),
        Reuse 3 (transfers to other commodity procurement domains), Durability 3.
   Sum 16. ADVANCES.

2. RFQ agent                                   [new in v2]
   Defensibility 4 (agentic action + structured comparison), Willingness 4,
        Buildability 2 (agentic browsing on supplier portals is non-trivial),
        Reuse 3, Durability 3.
   Sum 16. ADVANCES.

3. Hedge strategy assistant                    [new in v2]
   Defensibility 3 (hedging maths via verified computation; competitive with
        free models that improve at math), Willingness 3 (smaller buyers don't
        hedge), Buildability 3, Reuse 2 (commodity-specific), Durability 2
        (reasoning models will close this).
   Sum 13. ADVANCES with Durability flag.

5. Energy risk monitor                         [new in v2]
   Defensibility 3, Willingness 3 (alerts are useful but rarely high-pay),
        Buildability 3, Reuse 3, Durability 3.
   Sum 15. Advances.

(#4 contract intelligence and #6 budget forecaster — both score ~13–14; mark as
considers, not the headline candidates.)
```

### Read

**v2 surfaces three to four advancing candidates for the same domain v0 said
"no candidate advances" on.** The audience-fit challenge (shifting from high-paid
traders to mid-market procurement managers) was the unlock; the lenses then
generated multiple specialism-grounded candidates. **#1 (aligned-incentive
advisor) is the most distinctive** — the broker-conflict-of-interest framing is
a genuine positioning advantage and reusable to other commodity procurement
domains.

**Caveat:** these candidates assume a B2B/SaaS payment model (£20–100/mo
subscription), not a £2 day-pass. The simple-paid-multidomain-chat plan's
day-pass economics won't work for this audience. So gaswholesalers may
genuinely be a different *kind* of product (B2B SaaS, not consumer chat).
Flagged as an open question.

---

## Run 7 — robot-hands.com (v2)

### Audience (with audience-fit challenge)

v0 ran three framings (prosthetic users, industrial buyers, hobbyists) — all
failed. Audience-fit challenge for v2: are there *other* audiences? Useful
additions:

- **E.** Occupational therapists / prosthetists (professionals fitting
  prosthetics; B2B).
- **F.** Parents of children with limb difference (specific support need;
  emotional + practical decisions).
- **G.** Researchers in robotic manipulation (academic + industry R&D).
- **H.** Insurance/funding administrators (the people who approve the £8k–60k
  device purchase — a niche but potentially valuable audience).

### Generate — multi-lens (compressed)

**Most promising new candidates from the lenses:**

1. **Prosthetist productivity tool** [new in v2; audience E] — generate
   patient-specific exercise plans, progress notes, communication templates,
   training video scripts. B2B, recurring. Generalist failure: generic, not
   tuned to prosthetic-specific rehab.
2. **Insurance/funding case-builder** [new in v2; audience F or H] — assemble
   the funding case for a specific patient: medical justification, cost
   comparison, prior-art for insurance appeal. Long-context across patient
   records + insurer rules. Generalist failure: stale on specific insurer
   policies, no case-building structure.
3. **Patient progress tracker** [new in v2; audience E or family] — analyse
   video of prosthetic use over weeks, score improvement, suggest exercise
   adjustments. Vision + persistent memory. Generalist failure: no continuity,
   no quantified assessment.
4. **Robotic-manipulation literature digest** [new in v2; audience G] — load
   arxiv on manipulation, summarise, identify benchmarks. Already-served by
   Elicit/Connected Papers/SciSpace.
5. **Family decision support** [new in v2; audience F] — help parents weigh
   device options, ages, funding pathways. Sensitive context.

### Cut

- **#1 prosthetist tool**: real B2B audience, but tiny market (UK
  occupational therapists working with upper-limb prosthetics — a few hundred).
  Niche. Reuse low.
- **#2 funding case-builder**: most promising. Open Bionics' free Customer
  Success Officer already helps with this — seller-bundles-free check fires.
  But: the CSO is one person at one manufacturer; a tool that serves users of
  *any* prosthetic, across insurers, would not be displaced by one
  manufacturer's free service. **Keep.**
- **#3 progress tracker**: defensible (vision + memory), but small audience
  again; Open Bionics' Sidekick App already covers some of this for their users.
- **#4 research digest**: free substitutes already excellent. **Drop.**
- **#5 family decision support**: sensitive context; thin defensibility; free
  AI tutors increasingly handle this empathetically. **Drop or treat with care.**

### Verify

- Funding case complexity: insurance appeals for prosthetics are real and
  involved. Specialist consultants exist (rare, expensive). The case-building
  pattern is reusable across other expensive medical devices (wheelchairs,
  CGMs, hearing aids).
- Audience size for prosthetics alone: ~270,000 arm-prosthesis users globally;
  US ~50,000+ upper-limb amputees. Small market. **But cross-device reuse
  (#2 generalised) reframes the audience to "expensive-medical-device funding
  navigation" — much larger.**

### Score

```
1. Prosthetist productivity tool
   Defensibility 3, Willingness 3, Buildability 3, Reuse 2, Durability 3.
   Sum 14. Advances barely; small audience caps the upside.

2. Funding case-builder (prosthetic + broader medical devices)
   Defensibility 3 (the asset is curated insurer-rule knowledge + case-building
        process; not exclusive but laborious to replicate), Willingness 4
        (high-value devices, real appeals), Buildability 2 (needs insurer-rule
        curation, hard to bootstrap), Reuse 4 (cross-device generalisation),
        Durability 3.
   Sum 16. ADVANCES.

3. Patient progress tracker (vision-based)
   Defensibility 3, Willingness 3 (B2B OT use), Buildability 2, Reuse 2,
        Durability 3.
   Sum 13. Borderline.
```

### Read

**v2 surfaces a much better candidate (#2 funding case-builder generalised
across expensive medical devices) where v0 found nothing.** The audience-fit
challenge pushed past the obvious patient/hobbyist/B2B framings; the
generalist-failure lens spotted that insurance appeals are exactly where
specialism (curated rules + case-building structure) wins.

**However:** the domain name "robot-hands.com" doesn't naturally market a
funding-case-builder. The candidate is good; the brand fit is poor. **Flag as
an asset/positioning mismatch** — the candidate may belong on a different
domain. Worth carrying forward as a candidate looking for a domain, not the
other way round.

---

## Run 8 — agritec.uk (v2)

### Audience (with audience-fit challenge)

UK agritech. Plausible audiences:

- **A.** Small UK farmers (3–50ha — exactly the SFI26 Window 1 eligibility
  criterion).
- **B.** Mid-size commercial farms.
- **C.** Agronomists/agricultural consultants.
- **D.** Rural diversification operators (farm shops, glamping, agritourism).

**A** is the most concrete: SFI26 Window 1 opens **June 2026** — *next month*.
Eligibility specifically targets farms 3–50ha or those without an existing ELM
agreement; the scheme cap is £100k/year per business. There's anxiety because
SFI 2024 was abruptly closed when funds ran out. **Strong audience-fit + real
time-sensitive event.**

### Generate — multi-lens

**Demand lens (audience A):**
- Understand SFI26 eligibility for *my specific* farm.
- Pick the best actions from the 71 available.
- Maximise payment under the £100k cap.
- Apply during Window 1 (small farms) before funds run out (as in 2024).
- Stay compliant during the agreement period.

**Generalist-failure lens:**
- Stale on SFI26 specifics (rules just published; many models cut off before).
- Confident-wrong on eligibility nuance.
- Doesn't read the farmer's RPA digital map.
- Can't actually file the application.
- No persistent memory of the farm's actions and decisions.

**Frontier lens:**
- Long-context → load all DEFRA SFI26 guidance + NFU/CLA/AHDB analysis at once.
- Vision → read the farmer's RPA-issued map; estimate eligible parcels.
- Agentic → navigate the Rural Payments portal (notoriously unfriendly).
- Reasoning → optimise selection of 71 actions under the £100k cap and the
  rotational/area-based limits.
- Real-time → alerts when application windows open / when funding gets close
  to exhausted.

**Outcome lens:**
- "We're enrolled in SFI26 with the right actions, paid the maximum we're
  eligible for, before funding runs out."

### Candidates (after dedup)

1. **SFI26 eligibility & action selector** [new in v2] — load the farm's
   profile, identify eligibility (Window 1 vs 2, small-farm/no-ELM), recommend
   action mix to maximise payment under the £100k cap, prep application data.
2. **RPA filing agent** [new in v2] — agentic browsing fills the actual Rural
   Payments portal forms.
3. **Funding-window alert service** [new in v2] — monitor window status,
   alert when funding looks like running out (the 2024 abrupt-close was the
   pain point).
4. **Map/parcel analyser** [new in v2] — read RPA digital maps via vision,
   identify eligible parcels per action, flag boundary issues.
5. **Scheme compliance tracker** [new in v2] — after enrollment, track action
   compliance over the agreement period.
6. **Cross-scheme optimiser** [new in v2] — beyond SFI: CSHT, HLS, capital
   grants, private nature markets. The £100k cap + interaction with other
   schemes is genuinely complex; reasoning + long-context.
7. **Disease/pest identifier from photo** — vision-tuned for UK crops/animals.
   Different product; defensibility narrows because general vision is improving.
8. **Conference / discount matcher** (user's original suggestion) — partnership-
   based offers matched to farmer profile.

### Cut

- Free substitutes: DEFRA's own guidance is free; NFU (subscription), CLA,
  AHDB all provide free analysis; Carver Knowles / advisor firms charge for
  paid advice. **Seller-bundles-free check:** DEFRA is the "seller" of the
  scheme and provides free guidance — but the *navigation* of that guidance is
  the pain point, not the *availability*. The free substitute is voluminous and
  hard to use; this is exactly the "free substitute exists but is bad" pattern.
- **#7 disease identifier**: existing free apps (Plantix etc.) cover this; general
  vision models are getting good. **Drop unless we have a UK-specific edge.**
- **#8 discount matcher**: weak defensibility (the assets are the partnerships,
  which we don't yet have). **Park until partnerships exist.**

### Verify

- SFI26 details confirmed: Window 1 = June 2026, small farms 3–50ha or
  no-existing-ELM-agreement, 71 actions, £100k/year cap. SFI 2024 was abruptly
  closed when budget exhausted (March 2025) — creates real urgency and trust
  issues that the 2026 round may behave similarly.
- Window 2 = September 2026, open to all farms, no fixed end date.
- Existing advisors charge £hundreds to £k+ for SFI applications.
- Audience size: tens of thousands of UK small farms eligible.

### Score

```
1. SFI26 eligibility & action selector             [new in v2]
   Defensibility 4 (UK-specific scheme curation + optimisation logic;
        generalists are stale and confident-wrong on specifics; rules change
        annually so the curation is an ongoing asset), Willingness 4 (real
        money — a small farm could capture £k+ per year), Buildability 3 (load
        DEFRA docs + write the optimiser; achievable), Reuse 3 (UK-only but
        applies to many farms; transfers to other UK scheme contexts), Durability
        3 (rules change annually but curation is ongoing).
   Sum 17. ADVANCES. [test now via fake-door: "free SFI26 eligibility check"]

2. RPA filing agent                                [new in v2]
   Defensibility 4 (agentic action on a specific notoriously-bad government
        portal — high specialism win), Willingness 4 (filing is the main pain
        point), Buildability 2 (agentic browsing on RPA portal; HMG portals are
        complex/changeable), Reuse 3 (generalises to other HMG/insurance/utility
        portals), Durability 3.
   Sum 16. ADVANCES.

3. Funding-window alert service                    [new in v2]
   Defensibility 3 (the asset is monitoring + the trust framing), Willingness 4
        (the 2024 close burned people), Buildability 4 (light), Reuse 3,
        Durability 3.
   Sum 17. ADVANCES. [cheap to build; test now]

6. Cross-scheme optimiser                          [new in v2]
   Defensibility 4 (multi-scheme reasoning + £100k-cap interactions is
        genuinely hard for generalists), Willingness 4, Buildability 2, Reuse 3,
        Durability 3.
   Sum 16. ADVANCES.

4. Map/parcel analyser                             [new in v2]
   Defensibility 3 (vision on RPA maps is feasible; not exclusive but
        domain-tuned), Willingness 3, Buildability 3, Reuse 3, Durability 2
        (vision will commoditise).
   Sum 14. Advances barely.

5. Scheme compliance tracker
   Defensibility 3, Willingness 3, Buildability 3, Reuse 3, Durability 3.
   Sum 15. Advances.
```

### Read

**Agritec produces the strongest candidate set of the four runs — five
advancing candidates, three high-scoring.** The real-world hook (SFI26 Window 1
in June 2026 with the 2024-burn-out as urgency context) gives the audience
acute, time-bounded pain that AI specialism can demonstrably help with. The
"navigation of free guidance is bad" pattern is the unlock — DEFRA gives away
the information, but the *experience* is awful; that's the wedge.

**Strongest:** #1 (eligibility + action selector) and #3 (funding-window
alerts — cheap, urgent, immediate value). Both are test-now via fake-door.

---

## Cross-domain summary (v2)

| Domain          | v0 result                       | v2 result                           | Best v2 candidate |
|-----------------|---------------------------------|-------------------------------------|---|
| websitedesign   | 1 expensive consider            | **4 advancing**                     | Compliance-aware builder + audit |
| gaswholesalers  | 0 advancing                     | **3–4 advancing** (B2B SaaS, not £2 chat) | Aligned-incentive procurement advisor |
| robot-hands     | 0 advancing                     | **1 advancing** (but brand mismatch)| Funding case-builder (cross-device) |
| agritec.uk      | not previously run              | **5 advancing**                     | SFI26 eligibility + window alerts |

### What v2 changed

The v2 method produced **13 advancing candidates across four domains**, versus
**one expensive-consider** that v0 produced across three of these domains.
That's not a small change. The reasons it changed:

1. **Audience-fit challenge** moved gaswholesalers from "high-paid traders
   (overserved)" to "mid-market procurement (underserved)" — and that single
   move turned 0 advancing into 3. Same pattern, weaker, for robot-hands.
2. **Generalist-failure lens** consistently surfaced regulatory specificity
   (websitedesign compliance, agritec SFI26), agentic action (RPA portal, RFQ
   filing, CRO browsing), and verified computation (Lighthouse, hedging math)
   as wins. v0's "asset × capability" sweep never reached for these because they
   weren't in our asset list — but they're capabilities the audience needs.
3. **Frontier lens** brought in agentic browsing and code execution explicitly,
   producing several candidates that wouldn't have come from the v0 capability
   menu.
4. **Seller-bundles-free check** sharpened the cut: it killed weak candidates
   *and* identified one case (gaswholesalers brokers) where the seller-free
   pattern *creates* opportunity rather than blocks it — that's a real
   distinction the v1 cut would have missed.

### Open observations

- **Payment model fragmentation.** websitedesign's compliance audit, gas
  procurement advisor, and agritec SFI work are all probably **B2B SaaS**
  (£20–500/month subscriptions or per-job pricing), not £2 day-pass. The
  simple-paid-multidomain-chat plan's day-pass economics fit a narrower set of
  use cases than the original plan assumed. Worth revisiting that plan: the
  "chat-is-the-product" framing may need to expand to "chat + tool, B2B
  subscription" for the strongest candidates here.
- **Real-world timing matters.** SFI26 Window 1 in June 2026 makes agritec
  uniquely urgent right now; the same opportunity didn't exist 6 months ago and
  may not exist in the same form 6 months from now. The watchlist concept
  should track **scheme/event windows** for each domain, not just AI
  capabilities.
- **Robot-hands brand-fit mismatch.** v2 found a good candidate (cross-device
  funding case-builder) but the URL doesn't market it. Some domains may be
  better as URLs in search of a product than as products in search of a URL.
- **Compliance is a recurring wedge.** Two of the four domains (websitedesign
  WCAG/GDPR, agritec SFI/RPA) had their strongest candidates rooted in
  regulatory specificity. This is a *reusable pattern*: where regulations
  exist and the audience is non-expert, specialism wins.

### Verdict on the v2 method

The richer generation step works. v0 was filtering well but starving the funnel;
v2 produced 13 advancing candidates where v0 produced effectively 1. The
audience-fit challenge and the generalist-failure lens did most of the work;
the frontier lens added a few; the outcome lens didn't add as much as I'd hoped
on these particular domains (worth watching whether it earns its place on later
runs).

**Next:** test the **agritec SFI26 eligibility checker** as a fake-door demand
test (cheap to set up, time-sensitive — Window 1 opens June). If demand is
there, the build is reasonable and the timing is unique. websitedesign
compliance audit is a strong second.
