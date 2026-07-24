# GROUNDED domain profiles — second Gemini batch, checked against live data (2026-07-24)

*Reference document. Five of the eight domains in this batch are real, deployed
platform sites — unlike the first batch (0/15 grounded), this one is worth
building on. Each profile below states the real facts (pulled from `sites` and
`pages`, not the domain name), Gemini's guess, a corrected verdict, what
already exists live, and — where clear — a sharper Tier-3 candidate than
Gemini's. Decisions drawn from this file are in `PLAN` (D15/D16); the
extraction log is in `NOTES`.*

**Meta-lesson (see NOTES for full statement):** Gemini's accuracy correlated
directly with how literally the domain name described the real business — spot
-on for `ai-agent-orchestration.com`, badly wrong for `idea.uk` and
`leopardessconsulting.co.uk`. A domain that exists is not the same as a domain
whose real content was checked. Pulling real page titles + tagline was what
actually caught the misses; `audience.v1` itself was empty on every site
checked so far, so it isn't yet the mechanism to lean on.

---

## robot-hands.com — STRONG pilot candidate

**Real:** *"Robot-Hands.com | The Gripper Intelligence Platform for Serious
Robotics Engineers"* — a vendor-neutral robotic-gripper (end-effector)
selection and comparison platform. Real live pages include a gripper catalog,
**MatchMatrix** (a gripper-to-application matching tool, with a published
methodology page), a selection guide, a learning centre, technology-comparison
articles (pneumatic vs electric), and a services page ("integration support &
custom selection"). Three calculators are **already live**: Gripper Payload
Calculator, Grip Force & Friction Calculator, Gripper Cycle Time Estimator.

**Gemini's guess:** Technical & B2B; personas = automation engineers,
warehouse robotics integrators, *prosthetic & assistive tech researchers*.

**Corrected verdict — partially aligned.** The technical/engineering instinct
was right (a fairly literal domain name), but the real site is narrower and
more mature than guessed: it's a comparison/selection authority, not a general
robotics-hardware content site. "Prosthetics researchers" has no basis in the
real content.

**What this means for design:** Tier 1 is **already built** (3 calculators,
matching/exceeding Gemini's own Tier-1 idea). MatchMatrix already functions as
a Tier-2-ish sticky utility. **The real gap is Tier 3.** Known constraint:
`bugs_open/043` found live fabricated stats on this exact site (per memory:
"tool-generator has no fake-data rule") — any Tier-3 build here must route
through real claims-verification, not add polish on top of an unverified base.

**Sharper Tier-3 candidate than Gemini's** (Gemini proposed an "Automated
Dexterous Task Feasibility Audit" needing CAD-file/video ingestion — high
build cost, no real precedent on this platform): extend MatchMatrix into a
verified, deployed **Gripper Procurement & Cell-Integration Spec** —
Firecrawl-scrape real current vendor spec sheets, verify every quoted spec
against `evidence_base` before it can appear, render `go-echarts`
payload/cycle-time comparison charts, deploy a locked, shareable report. This
is P1/P2-shaped (D7), reuses infrastructure the site already has, and **its
verification requirement doubles as the structural fix for bugs_open/043** —
going forward, comparisons are sourced, not invented.

**Why this is the strongest pilot found so far:** real audience, real existing
Tier 1/2, a single well-scoped Tier-3 gap, and the build inherently remediates
a known live bug rather than being pure net-new risk.

---

## gaswholesalers.com — strong concept, sequence-blocked

**Real:** checked 2026-07-24, the site **still shows the stale, false content**
`006` exists to replace — tagline "Wholesale Gas Supply Solutions", company
name "Gas Wholesalers". `006`'s evidence (2026-07-20): 174 first-person
operational claims across 19 pages asserting a supply business that does not
exist. The rewrite has not shipped.

**Gemini's guess:** Technical & B2B; personas = Commercial Energy Procurement
Directors, Industrial Plant Operators, Utility Commodity Traders; tools:
therms↔kWh conversion, NBP/TTF spot-rate arbitrage inspector, a "Corporate Gas
Procurement & Hedging Strategy Dossier".

**Corrected verdict — surprisingly well-aligned with `006`'s *corrected*
positioning** (traders and oil/gas executives, analysis-and-tools-and-training),
even though Gemini was almost certainly reading the stale content or just the
domain name, not `006`. Good convergent validation of `006`'s direction from an
independent source.

**One framing risk to carry forward:** Gemini's Tier-3 dossier must be built as
**analysis/advisory output for the reader's own decision** — never as the site
itself claiming first-person supply or hedging capability. That exact
first-person-claim pattern is what triggered `006` in the first place; a Tier-3
tool that slips back into "we hedge your gas for you" language reintroduces the
defect it's supposed to help fix.

**Sequencing:** do not build tools on top of a site still carrying 174 false
claims — it compounds the problem rather than fixing it. This is a strong
pilot for `007`'s conversational Tier-3 variant (a high-quality-model advisory
chat for senior readers) but **only after `006`'s content rewrite ships**, and
that rewrite is itself gated on claims-verification's V5 (a `citation` source
kind, designed not built — CAPABILITIES §4).

---

## idea.uk — sharpest miss in the batch, needs its own follow-up

**Real:** *"idea.uk — Where You Take an Idea Seriously"*. Real pages: `tools`,
**`report.html` titled "Verified Idea Report"**, `guides`, `news`, `about`,
`contact`, and a live **"Free Audience Check"** tool.

**Gemini's guess:** Historical/Cultural/Creative + "Technical IP" archetype;
personas = product designers, **patent attorneys**, R&D directors, startup
founders; tools: an IPC/CPC patent-classification matrix, a UKIPO/EPO
prior-art fast-checker, an "IP Strategy & Prototype Blueprint".

**Corrected verdict — wrong.** Gemini free-associated "idea" → "invention →
patent". The real site validates a **business/product idea** (market and
audience fit, under a "verified" framing) for founders — nothing about patent
law, IP classification codes, or prior-art search anywhere in the real content.
This is the sharpest miss in the batch precisely because it sounds plausible on
a first read.

**Upgraded finding (2026-07-24, same day):** this is not just a page that
*looks* Tier-3-shaped — `idea-uk-vm-site-workstream` (a separate, active
session) verified **live today** that the home+header CTA funnels into the
paid `/report.html` tool, and confirmed the tool page carries a
`report-request-form` section (checked directly against the live `pages` row,
2026-07-24). **idea.uk already has a working, real, paid Tier-3-shaped funnel
in production** — the closest thing on the whole platform to a proof this
workstream's concept works. A follow-on (`bugs_open`/escalate `054`) is
**owned by that other session** — per CLAUDE.md's ownership-check rule, this
workstream should read their docs (`HANDOFF_RESUME_idea_uk_vm_site.md`) before
proposing anything on this domain, not start parallel work. **Recommendation
upgraded:** before designing any new Tier-3 pattern from scratch, study
idea.uk's live report-request flow as the working reference implementation.

---

## leopardessconsulting.co.uk — wrong archetype, strong real pilot underneath

**Real tagline:** *"AI systems that do a defined job, run without supervision,
and keep a record of every decision they make."* Real pages: an **"AI
Production Readiness Assessment"** (quiz) and genuine case studies — a
multi-stage data pipeline with Companies House verification, an LLM-driven
tool-generation and deployment pipeline, a real-time multi-source news
pipeline with credibility scoring, a hierarchical agent architecture for
autonomous website operations. Per memory (`leopardess-rebuild-workstream`):
"marketing fabrications audited out, no claim ships without an AUDIT row" —
these case studies are deliberately true, audited descriptions of this
platform's own real work.

**Gemini's guess:** High-Trust & Regulated; personas = C-suite executives, PE
operating partners, corporate restructuring boards; tools: a span-of-control/
overhead-ratio calculator, an ESG/governance gap checker, a "Board-Level
Strategy & Restructuring Dossier".

**Corrected verdict — wrong archetype and wrong tool set.** This is not a
generalist management/ESG consultancy. It's an **AI-engineering delivery
consultancy** whose credibility asset is genuine, audited case studies of
building production multi-agent systems — a sibling of
`ai-agent-orchestration.com`, both flagship framework-demonstration sites (PLAN
D8). The governance/ESG framing has no basis in the real content.

**What already exists:** an AI Production Readiness Assessment quiz — the same
shape as `ai-agent-orchestration.com`'s own readiness quiz — is a real,
already-live Tier-1/2-ish tool.

**Sharper Tier-3 direction than Gemini's:** not a board-deck/restructuring
dossier, but something like an **Automation Feasibility & Architecture Audit
for the visitor's own process** — verified, cited against the real case
studies, potentially culminating in a live demo build. Directly aligned with
D8 (the signature operation *is* the sales demo for the framework).

---

## ai-agent-orchestration.com — best-matched domain in the batch, strong pilot

**Real:** *"Production-Grade Multi-Agent Systems. Built Right."* Rich real
content: four real interactive tools already live — **AI Agent ROI Estimator,
LLM Provider Cost Comparison Calculator, AI Readiness Assessment quiz, Agent
Architecture Complexity Estimator** — plus a genuine technical engineering blog
(10+ real posts: orchestration failure modes, state management, Kafka/Postgres
architecture decisions) and case studies including "Enterprise Reference
Deployment: 70+ Agents in Production" and "Case Study: Fixing Kafka Consumer
Group Misconfiguration in a 40-Agent Financial Services Pipeline".

**Gemini's guess:** Technical B2B / Direct Flagship Showcase; personas =
enterprise CTOs, AI engineers, ops directors, system architects; tools: a
multi-agent token/infra cost simulator (close match to the real LLM cost
calculator), an agent-workflow-schema/circuit-breaker generator (plausible,
not built), Tier 3 = a "Live Autonomous Agent Fleet Deployment Demo" — enter a
brief, trigger a live saga, deploy a live multi-agent sub-domain fleet node in
under 3 minutes.

**Corrected verdict — strongly aligned, the best match in the whole batch.**
Likely because the domain name literally describes the real business.
Confirms PLAN D8 directly (AI-services sites run a pattern live as the sales
demo) from an independent source.

**Already exists:** effectively Tier 1+2 already live (4 real tools) plus a
strong technical credibility layer that could directly seed a Tier-3 op's
evidence base.

**Sequencing note on Gemini's Tier-3 idea:** the "live 3-minute fleet
deployment demo" is genuinely on-strategy for D8, but it depends on the
domain→live-site delivery path, which CAPABILITIES already flags as this
platform's most bug-fought surface. **Lower-risk first step:** turn the real
case-study material into an interactive, personalised "Architecture Fit
Report" (input the visitor's stack/scale → a verified analysis citing the real
case studies, no live deploy required) before attempting the live-deploy
version.

---

## Not onboarded (checked, do not use as real plans)

`agritec.uk`, `mortgagecalculator.co.uk`, `websitedesign.com` — **none exist in
the `sites` table** (checked 2026-07-24), same pattern as the first batch's 15.
Every archetype/tool guess for these three is inferred from the domain string
alone. Kept as ungrounded brainstorm seed only, appended to
`IDEAS_gemini_domain_brainstorm_2026-07-24.md` §3b — not used in any
recommendation here.

## Revised pilot recommendation

Superseding the earlier "Property" pick (PLAN, 2026-07-21), which was never
grounded in a real domain — Property was a pool, not a specific site:

1. **robot-hands.com — first choice.** Real audience, Tier 1/2 already live,
   one well-scoped Tier-3 gap, and the build remediates a known bug
   (`bugs_open/043`) rather than adding pure new risk.
2. **ai-agent-orchestration.com + leopardessconsulting.co.uk — close second,
   as a pair.** Both validate D8 specifically and are further along toward the
   3-tier funnel than any other domain checked; the "Architecture Fit Report"
   variant is lower-risk than Gemini's live-deploy-demo idea and should come
   first.
3. **gaswholesalers.com — strong concept, sequence it after `006` ships.**
4. **idea.uk — not a pilot target, it's the reference implementation.** Its
   home→CTA→paid-report funnel is already live and owned by another
   workstream. Read `HANDOFF_RESUME_idea_uk_vm_site.md` before designing any
   new Tier-3 pattern — the working example may already answer questions this
   workstream would otherwise re-derive from scratch.
