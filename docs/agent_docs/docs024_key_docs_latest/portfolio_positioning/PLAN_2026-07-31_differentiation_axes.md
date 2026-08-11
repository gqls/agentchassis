# PLAN — portfolio positioning: how ~154 finance domains become thick sites that cannot converge

**Opened 2026-07-31** by the loanandmortgagecalculator lane, at the owner's direction.
This is the direction-setting document for every future site lane in the finance/insurance
portfolio. The register seed lives beside it (`REGISTER_positioning.md`); the standing five
for this workstream start here.

## The brief, verbatim intent

> "I'd like to make them quite thick sites but different directions so it's important that
> we make them differentiate from each other from the start."
> Axes the owner named: target markets, age, stage in their process, size of purchase,
> type of purchase, experience.

Decisions taken 2026-07-31 (owner, via question round + correction):

| # | decision |
|---|---|
| P1 | **Thick sites, not thin exact-match pages** (owner corrected an earlier answer that had said thin — the correction supersedes) |
| P2 | **Roughly one site per proposition** — scale is high double digits, not 3–5 authorities |
| P3 | **Insurance vertical included now**, same framework |
| P4 | Differentiation must be **recorded as framework config from the start**, not maintained by memory — the same requirement that produced D3 on the live pair |

## Measured ground (2026-07-31, live DB + repo)

- **154 domains** across finance (82) and insurance (72) → **42 propositions** when
  clustered by what the name promises.
- Only **2** domains have `sites` rows (`loancalculator.co.uk`,
  `loanandmortgagecalculator.co.uk`); only **3** are in the deploy repo (those two +
  `mortgagecalculator.co.uk`). **~150 are greenfield** — the direction is being set at the
  cheapest possible moment.
- **The platform has NO cross-site overlap detection.** `check_content_duplication` is
  single-site (`WHERE p.site_id = $1`); `cross_site_contamination` detects company-name
  bleed, not topical overlap. Whatever discipline exists must live in the specs and in a
  register-level check we run ourselves.
- The two seams a site's positioning actually reaches a generation prompt through are
  `identity.target_audience` / `identity.key_differentiators` and
  **`content_direction.formatted`** (the writer reads no other field of
  `content_direction`). The aspect named `audience` is populated on 29/33 sites and read by
  **nothing**. (All measured 2026-07-31; see
  `loanandmortgagecalculator_couk/RUNBOOK…md` §11 and `LANDMINES.md`.)

## The three tiers — the name decides how much freedom exists

**Tier A — the name fixes the direction (~78 domains).** `equityreleasecalculator` has
chosen 55+; `adversecreditmortgage` has chosen "already declined"; `smbmortgages` has
chosen businesses. The register row *obeys* the name. Building anything else wastes the
exact-match value and confuses the visitor the name attracted.

**Tier B — the name is pure topic (~51 domains).** `savingsrates`, `whichloan`,
`interestrates`. The direction is free and MUST be assigned deliberately, because visitors
arrive unselected. These are where the axis catalogue does real work.

**Tier C — spelling twins (~25 domains).** `savingsrates.co.uk` vs `savings-rates.co.uk`
vs `savingsrates.uk`; every `.co.uk`/`.uk` pair of the same phrase. **No differentiation
axis exists between a hyphen and its absence.** Two thick sites for the same phrase are
the same site, and nothing in this framework can make them otherwise. Options per pair:
canonical/301 to the built sibling, or a deliberately accepted near-duplicate. **Each twin
pair is flagged in the register for an owner call; this plan does not decide them.**

## The axis catalogue

A site's **direction** is its coordinate on these axes. Most sites fix 3–4 axes hard and
leave the rest neutral. The name fixes whichever coordinates it encodes (Tier A); the
register assigns the rest.

### WHO — audience identity
1. **Age / life stage** — first-time buyer (20s–30s) · mover · later-life 55+ · retired
2. **Credit experience** — clean · thin-file (young, new to UK) · adverse (declined, CCJ,
   default, DMP, discharged bankruptcy) · rebuilding
3. **Financial sophistication** — novice (jargon-free, reassurance) · competent (tables,
   trade-offs) · expert (assumptions exposed, downloadable numbers)
4. **Borrower type** — individual · couple/family · landlord · self-employed/contractor ·
   limited company · SME
5. **Employment shape** — PAYE · contractor · key worker (NHS/teachers/police) · retired.
   NB `keyworkerinsurance` (an individual public-sector worker) and `keypersoncover` (a
   business insuring an employee) LOOK like twins and are **different buyers**.
6. **Wealth tier** — mass market · mass affluent · HNW (`private-finance.uk`,
   `bespokemortgages.uk`, `yachtfinancing.co.uk`)

### WHAT — the purchase
7. **Size** — £1k–25k loan · mortgage-scale · £1m+ bespoke · fleet/commercial
8. **Purchase type** — home · BTL · commercial premises · vehicle/fleet · boat · park home
9. **Product structure** — fixed vs variable · offset · interest-only · secured vs
   unsecured · term length
10. **Risk posture** — safety-first (protection, fixing, insuring) vs optimiser (highest
    rate, switching, chasing)

### WHEN — the moment
11. **Journey stage** — dreaming (years out) → preparing (6 months: credit file, deposit)
    → shopping → applying → **owning/managing** → struggling → exiting. Most sites fight
    over "shopping"; **preparing** and **owning** are near-empty ground and two of the
    portfolio's best claims (`mortgageextension`, `loanmanagement` are owning-stage names).
12. **Urgency** — deadline-driven (fix ending, completion date) vs unhurried
13. **Frequency** — once-a-decade decision (mortgage) vs monthly management (savings) —
    this axis dictates *freshness architecture*, not just copy: a rates site that is not
    visibly dated is dead on arrival

### HOW — the site's job (rescues the Tier B clusters; hides in the name's grammar)
14. **Content mode, read off the name's grammar:**

| name says | intent attracted | the site is |
|---|---|---|
| *calculator* | "give me my number" | tool-first |
| *rates* | "show me the market" | live data tables, dated |
| *quote / quotation* | "price my situation" | structured estimator, form-first |
| *deal(s)* | "what offers exist" | switching/offers, time-boxed |
| *best / highest* | "rank them and defend it" | verdicts with visible methodology |
| *compare / which / what* | "help me choose" | decision-trees, side-by-side |
| *forecast* | "what happens next" | predictive (`rateforecast.co.uk` — unique in the portfolio) |
| *review(s)* | "is it any good" | criteria-led evaluation |
| *app* | "give me the product" | product, not content |

15. **Editorial stance** — neutral educator · consumer champion (mildly adversarial to the
    industry — the register the two live sites already carry) · data journalist · insider
16. **Depth contract** — the 10-second answer vs the 20-minute deep dive
17. **Reasoning direction** — *what is* · *which* · *should I* · *how do I* · *what will
    happen*
18. **Emotional register** — reassurance (adverse, struggling) · aspiration (yacht,
    investment) · vigilance (chasers) · duty (protection, family cover)

### WHERE — market position
19. **Counterparty focus** — banks · building societies (mutual sector; older,
    branch-loyal — `buildingsocietyrates` is genuinely distinct) · challengers · brokers
20. **B2B vs B2C** — `bankingequipment.*` is B2B equipment, not consumer finance;
    `corporatehealthinsurance`/`staffhealthinsurance`/`workerhealthinsurance` sell to an
    employer
21. **Vertical integration** — pure information → lead-gen → product (`savingsapp`,
    `insuranceapp`)
22. **Price vs quality entry** — *cheap* names attract price-first; *best* attract
    quality-first; *compare* attract the methodical

### Compliance axis (insurance + regulated credit)
23. Insurance quotes/comparisons and regulated credit carry FCA presentation duties that
    calculators do not (fair-clear-not-misleading, prominence of exclusions, no implied
    advice). **Every insurance register row carries an explicit compliance line.** The
    house rule from the live pair — *explain mechanisms, never quote a current rate or tax
    band* — already keeps most of this out of trouble and stays the default.

## Worked proof — the hardest cluster carved by the axes

The 12 savings-rate domains looked undifferentiable. Grammar + counterparty + audience
produce **eight** genuinely distinct thick sites; the remainder are twins:

| domain | direction |
|---|---|
| `savingsrates.co.uk` | the live table — data-first, dated, the market right now |
| `highestinterest.co.uk` | the chaser — vigilance: loyalty penalties, when to move |
| `buildingsocietyrates.co.uk` | the mutual sector — building societies only |
| `banksavingsrates.co.uk` | big banks vs challengers — why the high street pays less |
| `savingsaccountrates.co.uk` | by account type — ISA vs bond vs easy-access vs notice |
| `saving-rate.co.uk` | a different sense entirely — *your* saving rate: % of income |
| `highestrate.co.uk` | cross-product — savings vs premium bonds vs gilts vs money market |
| `rateforecast.co.uk` | predictive — where rates go, and what that means for fixing |
| `savingsapp.co.uk/.uk` | product |
| `savings-rates.co.uk`, `savingsrates.uk`, `savingrates.co.uk` | **twins — no axis left**; owner call |

## Enforcement from day one (all machinery already proven on the live pair)

1. **The register** (`REGISTER_positioning.md`) — one row per domain: tier, proposition,
   fixed coordinates, assigned coordinates, out-of-scope statement, **named nearest
   neighbours + the sentence that separates them**. The register is the source of truth
   for every future lane.
2. **Coordinates → the seams the writer reads.** `identity.target_audience`,
   `identity.key_differentiators`, and a positioning block inside `content_direction`
   (with `formatted` regenerated via the gated port —
   `loanandmortgagecalculator_couk/set_divergence_specs.py` generalises to read from the
   register; parameterising its hardcoded domain is part of that job). **Never the
   `audience` aspect.**
3. **The collision check we lack platform-side:** no two rows may share the same
   (proposition × primary-audience × content-mode) coordinate. A small script over the
   register, run before any new site is built. This is the guard that makes convergence
   structurally impossible instead of unmonitored.
4. **Sequence rule: register row BEFORE adoption/build.** Adoption auto-writes a
   *generic* identity (measured: the crossing-point site's came back nearly identical to
   the loan-only site's). The row exists first; the adoption's guess is superseded
   immediately after.
5. **Cross-site linking policy:** hub-and-spoke by proposition family, never a reciprocal
   mesh; a site links out where its out-of-scope statement says the answer lives. The
   `divergence_rule` written for the live pair ("when a topic could be single-subject or
   crossing-point, choose the crossing-point framing") is the template — every row gets an
   equivalent rule naming its neighbours.
6. **Design divergence is also config:** pin each site's palette via
   `design_intent.palette.reference_values` (the webdesign colour-churn landmine) so the
   estate doesn't converge visually either.

## Known collisions to resolve in the register (flagged now, cheap now)

- `consolidatemortgage.co.uk` vs the crossing-point site's live guide
  *consolidating-debt-into-your-mortgage*: the domain owns the **decision in depth**
  (eligibility, broker questions, worked examples); the guide keeps the overview and links
  out. Same pattern for `equity-release-calculator` vs the mortgage site's existing
  equity-release tool, and `debtconsolidating` vs the loan sites' consolidation tools.
- `mortgagerepaymentsinsurance.*` sits in BOTH lists (mortgage + insurance) — one
  proposition, filed under insurance, cross-linked from the mortgage family.
- `banker.co.uk/.uk` and `loansy/loanzy` are brandables with no inherent direction —
  candidates for whatever proposition still lacks a good name after assignment, rather
  than forcing a direction now.

## What this plan does NOT decide

- Which twin of each Tier C pair gets built (owner call, per pair, in the register).
- Build order across the 42 propositions (traffic/commercial data the owner holds).
- Anything about serving infrastructure (every new site inherits the
  mortgagecalculator-handoff rule: **into the deploy repo before any adoption**).

## AMENDMENTS

**P5 (owner ruling, 2026-08-01) — cross-TLD twins are split, not 301'd.** Every
`.co.uk`/`.uk` pair of the same phrase gets both sides built, differentiated on the
depth-contract/mode axes by one portfolio-wide convention: **`.co.uk` = the authority
(thick, guide-led), `.uk` = the instrument (tool-first, the 10-second answer)**, cross-
linked as siblings. This supersedes the Tier C default for cross-TLD pairs; same-TLD
hyphen/plural twins remain ⚑OWNER with the 301 default. Full text: register rollup.

**P6 (owner ruling, 2026-08-01) — `loancash.co.uk` IS built, as the borrower's FCA
rulebook.** The register had recommended against building it because the name attracts
the audience FCA rules protect hardest. The owner inverted the risk into the point: the
site is a civilian champion of the FCA's consumer-credit rules — price cap, affordability
duties, CPA and rollover limits, complaint rights, authorised-lender checks, loan-shark
reporting — so the vulnerable query lands on protection instead of a lender. Register
entry **L10**; explicitly independent of the FCA on every page.

**P7 (owner ruling, 2026-08-01) — same-TLD twins split by SEAT, and the seat lens applied
portfolio-wide.** The Tier C claim ("no axis exists within one TLD") held the reader's
ROLE constant without noticing. The seats — buyer, setter, intermediary, observer,
referee — are different audiences with different search behaviour, so identical phrases
can hold genuinely different thick sites. Owner corrections encoded: the archive is a
SECTION not a seat (no transaction near it); exactly ONE buyer per phrase (two buyer
sites on one phrase is self-competition, the register's own failure mode); each seat's
money is named (buyer=affiliate/leads, intermediary=B2B affiliate + adviser lead-gen,
referee=professional readership, setter/observer=citations); and the authority route is
policy — professional seats link down to buyer domains, which is where direct revenue
concentrates. Cleanest spelling = buyer; awkward/hyphenated spellings take the seats.
Referee = per-vertical casebooks (3–4 thick, never 15 thin). Full map + honest residuals:
the register's seat-map table. Two domains remain true twins even under seats; they stay
⚑OWNER.

**P8 (owner ruling, 2026-08-11) — the two seat-framework residuals are closed, by axes the
seat framework doesn't reach.** Both were the superlative ("best…") tier, where the
singular/plural grammar that carried the rest of the framework stops distinguishing
anything. `besthealthinsurancerate.co.uk` splits from its plural sibling by **regional
scope** (England vs whole-UK — NHS is devolved, so this is real content, not a relabelling,
provided the England site's spine is region-specific data). `bestlandlordinsurancerate.co.uk`
splits by **portfolio size** (1–10 properties, I5's existing audience, vs 3+/HMO portfolio
landlords — a genuinely different insurance product and sophistication tier). Both axes were
checked against the rest of the register for collisions before use (regional: unclaimed
anywhere; portfolio size: I5's family only). Not a general licence — most residuals in this
register were never twins-with-an-axis, they're holds for other reasons (product-app
domains, a different market). Full text: register rollup.
