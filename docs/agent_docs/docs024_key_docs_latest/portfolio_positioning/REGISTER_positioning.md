# REGISTER — portfolio positioning

One entry per **proposition**; every domain in the portfolio appears exactly once. This is
the source of truth a lane reads BEFORE building or adopting a site, and the source
`set_divergence_specs.py` (generalised) writes into `identity` + `content_direction`.

Field meanings — see `PLAN_2026-07-31_differentiation_axes.md` for the axis catalogue.
`primary:` names the domain to build; `twins:` are spelling variants with **no axis left**
(owner call per pair: 301 or accepted duplicate — marked `⚑OWNER`). `neighbours:` name the
propositions this one must not drift into, with the separating rule.

**Collision invariant: no two entries may share (family × audience × mode).** Check before
adding or changing an entry.

Status legend: LIVE (serving) · BUILT (repo, not adopted) · ADOPTED · — (greenfield).

---

## Family: MORTGAGE

### M1 — crossing-point (loans × mortgage) — LIVE+ADOPTED
- **primary:** loanandmortgagecalculator.co.uk
- **audience:** borrower whose unsecured debt and mortgage interact · **age:** 25–55 ·
  **stage:** preparing/shopping · **mode:** calculator+guide · **stance:** consumer champion
- **owns:** affordability-with-debts, consolidate-into-mortgage *overview*, deposit-vs-debt,
  credit-file-before-application
- **neighbours:** M2 (single-subject mortgage → theirs), L1 (single-subject loans → theirs),
  M8 (consolidation *depth* → theirs). Rule: crossing-point framing always wins here.

### M2 — mortgage general tool (narrow authority) — handed off for adoption
- **primary:** mortgagecalculator.co.uk · **also:** mortgageinterestcalculator.co.uk
  (angle: the *interest mechanics* tool — how much of a payment is interest, daily vs
  monthly calculation; distinct mode, same family)
- **audience:** anyone with a mortgage question · **stage:** shopping · **mode:**
  calculator-first · **stance:** neutral educator
- **neighbours:** M1 (anything touching other debt → theirs), M3 (live rate data → theirs)

### M3 — mortgage rates (the market, today)
- **primary:** mortgage-rates.co.uk · **also:** bankmortgagerates.co.uk (angle: big-bank
  vs broker pricing), fixedmortgagerates.co.uk + fixedinterestmortgagerates.co.uk (⚑OWNER
  near-twins; angle if kept: fixed-only, when to fix, fix length trade-offs)
- **audience:** deadline-driven shopper · **urgency:** high · **mode:** data-first, dated ·
  **stance:** data journalist
- **neighbours:** B3 (savings rates), M2 (tools). Rule: this family shows the market; it
  never computes a personal number — link to M2 for that.

### M4 — remortgage (stage: fix ending)
- **primary:** remortgagecalculator.uk · **also:** mortgage-refinance.co.uk (angle: the
  US-term searcher, often expat/new-to-UK — explain UK remortgaging to someone who says
  "refinance") · **twins:** remortgagequotation.co.uk/.uk, remortgagequote.uk ⚑OWNER
- **audience:** existing owner, fix ending in ≤6 months · **age:** 30–60 · **stage:**
  owning→shopping · **mode:** calculator + deadline checklists · **register:** urgency
- **neighbours:** M8 (remortgaging *to consolidate* → M8 depth), M3 (rate tables)

### M5 — adverse credit mortgage
- **primary:** adversecreditmortgage.co.uk · **also:** adversemortgage.uk (angle if kept:
  shorter, triage-first "can I get a mortgage at all" checker) · poor-credit-mortgages.co.uk
  + poorcreditmortgages.co.uk (⚑OWNER twins of each other; angle vs "adverse": plain-English
  register for people who would never say "adverse" — same facts, novice sophistication) ·
  **twins:** adversecreditmortgage.uk ⚑OWNER
- **audience:** declined / CCJ / default / DMP / post-bankruptcy · **experience:** adverse ·
  **register:** reassurance, zero judgement · **mode:** eligibility decision-trees + guides
- **compliance:** no "guaranteed acceptance" language, ever.
- **neighbours:** M1 (credit-file prep guide → overview only), L3 (debt consolidation)

### M6 — equity release (age-fixed: 55+)
- **primary:** equityreleasecalculator.co.uk · **twins:** equity-release-calculator.co.uk
  ⚑OWNER · **also:** equitycalculator.co.uk (angle: NOT age-fixed — plain equity: how much
  do I own, LTV, negative equity; feeds M4 and M7 audiences too)
- **audience:** 55+, asset-rich · **stage:** later life · **mode:** calculator + long-form
  duty-of-care guides (compound roll-up shown honestly) · **register:** duty, family framing
- **compliance:** ER is advised-sale territory — content must say "this requires regulated
  advice" plainly.
- **neighbours:** M7 (extending term as the alternative), B-family (funding retirement from
  savings instead)

### M7 — term extension / payment pressure
- **primary:** mortgageextension.co.uk · **twins:** mortgageextensions.co.uk ⚑OWNER
- **audience:** owner struggling with payments or wanting lower outgoings · **stage:**
  owning/struggling · **register:** reassurance, practical · **mode:** should-I decision
  support (extend vs overpay-stop vs consolidate vs sell)
- **neighbours:** M6 (55+ route), M8, L3. Rule: this is the *mortgage-term* lever only.

### M8 — consolidate into the mortgage (depth)
- **primary:** consolidatemortgage.co.uk
- **audience:** homeowner with unsecured debt · **stage:** owning · **mode:** should-I in
  depth — eligibility, worked examples, broker questions, secured-vs-unsecured trade-off
- **neighbours:** M1 holds the *overview* guide and links here; L3 holds non-mortgage
  consolidation. This entry resolves the collision flagged in the PLAN.

### M9 — buy-to-let / landlord borrowing
- **primary:** buytoletcalculator.uk
- **audience:** landlord/investor · **purchase type:** BTL · **mode:** yield/ICR
  calculator-first · **sophistication:** competent-to-expert
- **neighbours:** I5 (landlord *insurance*), M12 (company borrowing incl. SPV mortgages)

### M10 — offset mortgage
- **primary:** offset-mortgage.co.uk
- **audience:** saver with a mortgage — higher sophistication, often self-employed with tax
  reserves · **mode:** what-is + calculator (offset vs overpay vs save)
- **neighbours:** B-family (savings rates side of the same decision)

### M11 — bespoke / large / complex
- **primary:** bespokemortgages.uk
- **audience:** HNW, complex income, £1m+ · **size:** large · **mode:** guide-first, broker
  orientation · **register:** discreet, expert
- **neighbours:** X3 (private-finance.uk — HNW *whole* balance sheet; this is property only)

### M12 — commercial / SMB mortgages
- **primary:** smbmortgages.co.uk
- **audience:** business owner buying premises · **borrower:** company · **mode:** guide +
  eligibility
- **neighbours:** L5 (business lending generally — this is property-secured only)

## Family: LOAN / DEBT

### L1 — loan calculator (narrow authority) — LIVE+ADOPTED (other lane)
- **primary:** loancalculator.co.uk · **also:** loanrepayment.uk (angle: repayment
  schedules/overpayment mechanics), loanmanagement.co.uk (angle: **owning stage** — manage
  existing loans: statements, settlement figures, PPI-era rights, overpayment rules)
- **audience:** anyone pricing a personal loan · **mode:** calculator-first
- **neighbours:** M1 (anything touching a mortgage), L2 (choosing between products)

### L2 — which loan / guidance
- **primary:** whichloan.co.uk · **also:** whatloan.co.uk (⚑OWNER near-twin; angle if
  kept: definitional "what is a…" glossary site vs whichloan's decision-trees),
  borrowing.co.uk (angle: the umbrella authority — every way to borrow, ranked by cost,
  the "start here" site), loanscentre.co.uk (⚑OWNER generic; hold), loancash.co.uk (⚑OWNER
  — payday-adjacent name; **recommend not building**: the audience it attracts is the one
  FCA rules protect hardest)
- **audience:** undecided borrower · **mode:** which/compare decision-trees · **stance:**
  consumer champion
- **neighbours:** L1 (the numbers), L4/L6/L7 (product-specific depth)

### L3 — debt consolidation (unsecured)
- **primary:** consolidateloans.co.uk · **twins:** consolidateloans.uk ⚑OWNER · **also:**
  debtconsolidating.co.uk (⚑OWNER near-twin; angle if kept: **in-difficulty** register —
  DMPs, Breathing Space, StepChange signposting — vs the primary's cost-optimiser register)
- **audience:** multiple debts, wants one payment · **stage:** struggling or optimising ·
  **register:** reassurance
- **neighbours:** M8 (into the mortgage), M5 (adverse credit)

### L4 — long-term loans
- **primary:** longtermloan.co.uk · **twins:** longtermloan.uk ⚑OWNER
- **audience:** larger unsecured amounts, 5–10yr terms · **mode:** total-cost honesty (the
  term-vs-total-cost lesson is this site's whole reason)
- **neighbours:** L1, L6 (secured alternative)

### L5 — business finance
- **primary:** financeforcompanies.co.uk (fullest name) · **also:**
  business-lenders.co.uk (angle: the lender directory — who lends to whom, data-first),
  companyloan.uk (angle: the product explainer + calculator), businessfinances.uk (⚑OWNER
  generic; angle if kept: managing company money, not borrowing — cashflow, accounts)
- **audience:** SME director · **borrower:** company · **sophistication:** competent ·
  **compliance:** unregulated lending — different disclosure world, say so.
- **neighbours:** M12 (premises), L8 (asset finance), I2 (key person cover)

### L6 — secured / specific-structure personal credit
- **primary:** unsecuredpersonalloans.co.uk
- **audience:** borrower choosing security structure · **mode:** what-is + which (secured
  vs unsecured is the site's core question, from the *unsecured* side)
- **neighbours:** M8/L3 (consolidation), L4

### L7 — investment borrowing
- **primary:** investmentloan.uk · **twins:** investmentloans.uk ⚑OWNER
- **audience:** borrowing to invest (portfolio landlords margin aside — mostly property
  deposits, business stakes) · **sophistication:** expert · **register:** risk-forward —
  leverage cuts both ways is the editorial spine
- **neighbours:** M9 (BTL specifically), B7 (rateforecast — cost-of-carry)

### L8 — asset finance (high-value, niche)
- **primary:** fleetfinancing.co.uk (B2B vehicles) · **also:** yachtfinancing.co.uk
  (HNW marine — genuinely separate market), mobilehomeloans.co.uk (park homes — NB often
  NOT mortgageable land, a genuinely underserved niche)
- **audience:** per domain — fleet manager / HNW buyer / park-home buyer · each is its own
  micro-authority; they share chassis, not content
- **neighbours:** L5, M11

### L9 — loan brandables (direction unassigned, deliberately)
- **domains:** loansy.uk, loanzy.co.uk, loanzy.uk, finance.org.uk, banker.co.uk, banker.uk
- **status:** HOLD. Brandables carry no intent; assign them to whichever proposition still
  lacks a good name after the exact-match domains are placed, or to products (an actual
  comparison tool brand). finance.org.uk is the strongest hold-back — .org.uk carries
  trust weight worth spending deliberately.

## Family: BANKING / SAVINGS / RATES

### B1 — savings rates: the live table
- **primary:** savingsrates.co.uk · **twins:** savings-rates.co.uk, savingsrates.uk,
  savingrates.co.uk, saving-rate.co.uk† ⚑OWNER (†see B6 — different sense available)
- **audience:** saver checking the market · **frequency:** recurring · **mode:** data-first,
  visibly dated · **freshness is the product** — do not build without a data pipeline plan
- **neighbours:** B2, B3, B4, B7 — this family is the most collision-prone in the
  portfolio; every entry's out-of-scope line matters

### B2 — the rate chaser
- **primary:** highestinterest.co.uk · **also:** highestrate.co.uk (angle: cross-product —
  savings vs premium bonds vs gilts vs money-market funds), highestsavingsrate.co.uk
  (⚑OWNER near-twin of B2 primary)
- **audience:** optimiser, moves money · **risk posture:** optimiser · **register:**
  vigilance — loyalty penalties, when to switch · **mode:** best/ranked with methodology
- **neighbours:** B1 (the table), B7 (whether to fix)

### B3 — savings by institution type
- **primary:** buildingsocietyrates.co.uk (mutual sector; older, branch-loyal audience) ·
  **also:** banksavingsrates.co.uk (big banks vs challengers), banksrates.co.uk +
  banks-rates.co.uk (⚑OWNER twins; broader bank-pricing angle: savings AND borrowing at
  the big banks)
- **audience:** chooses the institution first · **counterparty axis is the identity**
- **neighbours:** B1, B2

### B4 — savings by account type
- **primary:** savingsaccountrates.co.uk
- **audience:** tax-aware saver · **mode:** which — ISA vs fixed bond vs easy-access vs
  notice; the wrapper logic, not just the headline number
- **neighbours:** B1 (numbers), pensions/investment ground NOT owned here (out of scope)

### B5 — interest rates (macro)
- **primary:** interestrates.co.uk · **also:** moneyrates.co.uk (⚑OWNER generic; angle if
  kept: all money prices in one place — savings, mortgage, loan, cards),
  competitiverates.co.uk + ratecomparison.co.uk (⚑OWNER generics; comparison-mode angle)
- **audience:** wants to understand the base rate and what it does to them · **mode:**
  what-is + explainer, MPC-date aware · **stance:** data journalist
- **neighbours:** B7 (prediction — B5 explains, B7 forecasts), M3 (mortgage pricing)

### B6 — your saving rate (different sense)
- **primary:** saving-rate.co.uk — IF the owner keeps it out of B1's twin set
- **audience:** habit-builder — what % of income to save, emergency fund sizing ·
  **stage:** planning · **mode:** calculator + workbook
- **neighbours:** B1 (product rates), pensions (out of scope)

### B7 — rate forecast (predictive — unique)
- **primary:** rateforecast.co.uk
- **audience:** anyone timing a fix (mortgage or savings) · **mode:** predictive — market
  expectations, swap curves in plain English, scenario tools · **stance:** data journalist
- **compliance:** forecasts framed as market expectations, never advice; every forecast
  dated and revisited — the site's honesty IS the archive of its own past calls.
- **neighbours:** B5 (explains today), M3/M4 (act on it)

### B8 — savings app (product)
- **primary:** savingsapp.co.uk · **twins:** savingsapp.uk ⚑OWNER
- **status:** product, not content — HOLD until there is an actual product; a content site
  here would burn the name.

### B9 — banking equipment (B2B — different market entirely)
- **primary:** bankingequipment.co.uk · **twins:** bankingequipment.uk ⚑OWNER
- **audience:** procurement — branch/office kit · **NOT consumer finance**; excluded from
  every collision check in this register; build as trade content or HOLD.
- **neighbours:** none in this portfolio — different market. If built, it competes with
  trade press, not with any sibling here.

## Family: X — cross-cutting

### X1 — mortgage payment protection
- **primary:** mortgagerepaymentsinsurance.co.uk · **twins:** mortgagerepaymentsinsurance.uk
  ⚑OWNER — appears in both the mortgage and insurance lists; ONE proposition, filed with
  insurance, cross-linked from M-family
- **audience:** new mortgagor, duty register · **compliance:** MPPI's mis-selling history
  is part of the honest story — say it
- **neighbours:** I4 (income protection — the usually-better answer; this site must say so)

### X2 — HNW private finance
- **primary:** private-finance.uk
- **audience:** HNW whole-balance-sheet — lombard lending, private banks, structured credit
- **neighbours:** M11 (property only), B-family (retail savings — out of scope here)

## Family: INSURANCE (72 domains → 13 propositions)

### I1 — private medical insurance, individual
- **primary:** privatehealthinsurancequote.co.uk · **the 19-domain cluster:**
  besthealthinsurancerate(s).co.uk/.uk, healthinsurancerate(s).co.uk/.uk,
  healthinsurancedeal.co.uk/.uk, healthinsurancequote.uk, healthinsurancequotation.co.uk/.uk,
  medicalinsurancequotation.co.uk/.uk, privatehealthinsurancequotation.uk,
  privatemedicalinsurancequotation.uk, personalhealthinsurancequote.uk,
  bestprivatehealthcare.uk†, comparehealthcare.uk†
  — mostly ⚑OWNER twins by grammar (rate/deal/quote/best variants). †The two healthcare
  names are a REAL distinct angle: choosing private *treatment/hospitals* (self-pay vs
  insured), not choosing a policy — recommend building that as its own site.
- **audience:** individual leaving NHS waiting lists · **register:** duty + plain English ·
  **mode per grammar:** quote-names = estimator-first; rate-names = pricing data;
  best-names = ranked with methodology
- **compliance:** FCA ICOBS territory — exclusions prominent, pre-existing-condition rules
  stated, no advice implied. This line applies to ALL I-family rows.
- **neighbours:** I2 (family), I3 (employer-paid), the healthcare† angle

### I2 — PMI, family
- **primary:** comparefamilyhealthinsurance.co.uk · **twins/near:**
  comparefamilyhealthinsurance.uk, familyhealthinsurancequote.co.uk/.uk ⚑OWNER
- **audience:** parents costing cover for children · **mode:** compare
- **neighbours:** I1 (individual), I3

### I3 — PMI, employer-paid
- **primary:** corporatehealthinsurance.uk · **also:** staffhealthinsurance.co.uk/.uk,
  workerhealthinsurance.co.uk/.uk (⚑OWNER twins; angle if kept: SME <50 heads vs
  corporate schemes), keyworkerinsurance.co.uk/.uk (**different buyer** — the public-sector
  individual: NHS/teachers/police discounts and occupational quirks; keep separate)
- **audience:** HR/founder (corporate/staff) · the keyworker pair: the worker themselves
- **neighbours:** I1, L5 (business costs), I8 (key person — the OTHER "key" name)

### I4 — income protection
- **primary:** incomeprotectioncover.co.uk · **twins:** incomeprotectioncover.uk ⚑OWNER
- **audience:** self-employed + sole earners · **register:** duty · the anti-PPI story —
  what MPPI should have been
- **neighbours:** X1 (mortgage-payment cover — narrower product this one usually beats),
  I6 (life)

### I5 — landlord insurance
- **primary:** landlordinsurancerates.co.uk · **cluster:**
  bestlandlordinsurancerate(s).co.uk/.uk, landlordinsurancedeal(s).co.uk/.uk,
  landlordinsurancerate.co.uk (⚑OWNER grammar twins) · rentalsinsurance.co.uk/.uk (angle:
  tenant-side / rent-guarantee — arguably a different buyer; hold for owner call)
- **audience:** landlord — same person M9 serves; **cross-link M9↔I5 is the portfolio's
  best internal seam**
- **neighbours:** M9, I7 (home — owner-occupier, different policy type)

### I6 — life insurance
- **primary:** lifeinsurancequotation.co.uk · **twins:** lifeinsurancequotation.uk,
  lifeinsurancerate.co.uk/.uk, lifeinsurancedeal.uk ⚑OWNER
- **audience:** new parents / new mortgagors · **register:** duty · **mode:** quote-first
- **neighbours:** I4 (living-benefit cover), X1

### I7 — home insurance
- **primary:** homeinsurancerates.co.uk · **twins/near:** homeinsurancerate.uk,
  homeinsurancedeal.co.uk/.uk, cheaphouseinsurance.uk (price-first entry — angle: the
  cheap-seeker, excess/cover trade-offs stated honestly) ⚑OWNER
- **audience:** owner-occupier at renewal · **frequency:** annual · **mode:** rates/compare
- **neighbours:** I5 (landlord), I1 (grammar cousins)

### I8 — key person cover (business)
- **primary:** keypersoncover.co.uk · **twins:** keypersoncover.uk ⚑OWNER
- **audience:** SME director insuring the business against losing a person · **buyer:**
  company — NOT the keyworker pair (individual)
- **neighbours:** I3/keyworker (the collision-in-waiting; both rows carry the distinction)

### I9 — professional indemnity
- **primary:** indemnitycover.co.uk · **also:** journalistinsurance.co.uk/.uk (worked
  niche example: PI for journalists — libel/defamation; either the PI site's flagship
  vertical or its own micro-site)
- **audience:** freelancer/consultant required to hold PI
- **neighbours:** I8, L5

### I10 — insurance generic / brandables
- **domains:** bestinsurancedeal(s).co.uk/.uk, bestinsurancerate(s).co.uk/.uk,
  cheapinsurancequote.uk, insuranceapp.co.uk/.uk (product — HOLD like B8),
  reviewsinsurance.co.uk/.uk (angle: the evaluation layer — reviews of insurERS not
  prices; genuinely distinct mode), sportsreviewinsurance.co.uk/.uk (unclear; hold)
- **status:** assign after the exact-match rows are built, same rule as L9.

---

## Twin-pair decisions owed by the owner (⚑OWNER rollup)

Every ⚑OWNER above, ~40 domains. Default recommendation for each: **301 to the built
sibling** (preserves the asset, consolidates authority, costs nothing to reverse later).
The alternative — a deliberately accepted near-duplicate — should be chosen knowingly per
pair, never arrived at by two lanes building independently. **No lane builds a ⚑OWNER
domain without this register saying which way the pair went.**

---

## Claims table (machine-readable — `check_register.py` parses THIS, not the prose)

Every domain, claimed by exactly one entry. The prose above is for humans; this block is
the contract. A new domain enters the portfolio by being added here AND to
`PORTFOLIO_domains.txt`, and the check must pass before any lane builds it.

```claims
M1: loanandmortgagecalculator.co.uk
M2: mortgagecalculator.co.uk mortgageinterestcalculator.co.uk
M3: mortgage-rates.co.uk bankmortgagerates.co.uk fixedmortgagerates.co.uk fixedinterestmortgagerates.co.uk
M4: remortgagecalculator.uk mortgage-refinance.co.uk remortgagequotation.co.uk remortgagequotation.uk remortgagequote.uk
M5: adversecreditmortgage.co.uk adversecreditmortgage.uk adversemortgage.uk poor-credit-mortgages.co.uk poorcreditmortgages.co.uk
M6: equityreleasecalculator.co.uk equity-release-calculator.co.uk equitycalculator.co.uk
M7: mortgageextension.co.uk mortgageextensions.co.uk
M8: consolidatemortgage.co.uk
M9: buytoletcalculator.uk
M10: offset-mortgage.co.uk
M11: bespokemortgages.uk
M12: smbmortgages.co.uk
L1: loancalculator.co.uk loanrepayment.uk loanmanagement.co.uk
L2: whichloan.co.uk whatloan.co.uk borrowing.co.uk loanscentre.co.uk loancash.co.uk
L3: consolidateloans.co.uk consolidateloans.uk debtconsolidating.co.uk
L4: longtermloan.co.uk longtermloan.uk
L5: financeforcompanies.co.uk business-lenders.co.uk companyloan.uk businessfinances.uk
L6: unsecuredpersonalloans.co.uk
L7: investmentloan.uk investmentloans.uk
L8: fleetfinancing.co.uk yachtfinancing.co.uk mobilehomeloans.co.uk
L9: loansy.uk loanzy.co.uk loanzy.uk finance.org.uk banker.co.uk banker.uk
B1: savingsrates.co.uk savings-rates.co.uk savingsrates.uk savingrates.co.uk savingsrate.co.uk
B2: highestinterest.co.uk highestrate.co.uk highestsavingsrate.co.uk
B3: buildingsocietyrates.co.uk banksavingsrates.co.uk banksrates.co.uk banks-rates.co.uk
B4: savingsaccountrates.co.uk
B5: interestrates.co.uk moneyrates.co.uk competitiverates.co.uk ratecomparison.co.uk
B6: saving-rate.co.uk
B7: rateforecast.co.uk
B8: savingsapp.co.uk savingsapp.uk
B9: bankingequipment.co.uk bankingequipment.uk
X1: mortgagerepaymentsinsurance.co.uk mortgagerepaymentsinsurance.uk
X2: private-finance.uk
I1: besthealthinsurancerate.co.uk besthealthinsurancerate.uk besthealthinsurancerates.co.uk healthinsurancerate.co.uk healthinsurancerate.uk healthinsurancerates.co.uk healthinsurancedeal.co.uk healthinsurancedeal.uk healthinsurancequote.uk healthinsurancequotation.co.uk healthinsurancequotation.uk medicalinsurancequotation.co.uk medicalinsurancequotation.uk privatehealthinsurancequote.co.uk privatehealthinsurancequotation.uk privatemedicalinsurancequotation.uk personalhealthinsurancequote.uk bestprivatehealthcare.uk comparehealthcare.uk
I2: comparefamilyhealthinsurance.co.uk comparefamilyhealthinsurance.uk familyhealthinsurancequote.co.uk familyhealthinsurancequote.uk
I3: corporatehealthinsurance.uk staffhealthinsurance.co.uk staffhealthinsurance.uk workerhealthinsurance.co.uk workerhealthinsurance.uk keyworkerinsurance.co.uk keyworkerinsurance.uk
I4: incomeprotectioncover.co.uk incomeprotectioncover.uk
I5: bestlandlordinsurancerate.co.uk bestlandlordinsurancerate.uk bestlandlordinsurancerates.co.uk landlordinsurancedeal.co.uk landlordinsurancedeal.uk landlordinsurancedeals.co.uk landlordinsurancerate.co.uk landlordinsurancerates.co.uk rentalsinsurance.co.uk rentalsinsurance.uk
I6: lifeinsurancedeal.uk lifeinsurancequotation.co.uk lifeinsurancequotation.uk lifeinsurancerate.co.uk lifeinsurancerate.uk
I7: cheaphouseinsurance.uk homeinsurancedeal.co.uk homeinsurancedeal.uk homeinsurancerate.uk homeinsurancerates.co.uk
I8: keypersoncover.co.uk keypersoncover.uk
I9: indemnitycover.co.uk journalistinsurance.co.uk journalistinsurance.uk
I10: bestinsurancedeal.co.uk bestinsurancedeal.uk bestinsurancedeals.co.uk bestinsurancerate.co.uk bestinsurancerate.uk bestinsurancerates.co.uk cheapinsurancequote.uk insuranceapp.co.uk insuranceapp.uk reviewsinsurance.co.uk reviewsinsurance.uk sportsreviewinsurance.co.uk sportsreviewinsurance.uk
```
