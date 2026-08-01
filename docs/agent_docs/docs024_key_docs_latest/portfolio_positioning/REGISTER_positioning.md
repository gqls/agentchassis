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
- **audience:** anyone pricing a personal loan · **age:** broad, 20–60 · **stage:**
  shopping (the primary); `loanmanagement` claims the empty **owning** stage · **size:**
  £1k–£25k unsecured · **experience:** novice-to-competent — the tool answers before any
  jargon is needed · **mode:** calculator-first
- **neighbours:** M1 (anything touching a mortgage), L2 (choosing between products)

### L2 — which loan / guidance
- **primary:** whichloan.co.uk · **also:** whatloan.co.uk (⚑OWNER near-twin; angle if
  kept: definitional "what is a…" glossary site vs whichloan's decision-trees),
  borrowing.co.uk (angle: the umbrella authority — every way to borrow, ranked by cost,
  the "start here" site), loanscentre.co.uk (**seat: intermediary** per the P7 map — the
  loan-broker layer)
- **audience:** the undecided borrower — knows they need money, not which shape · **age:**
  20–45, skewing to first-substantial-borrowing years · **stage:** dreaming→shopping ·
  **size:** £500–£15k · **experience:** novice; this is many visitors' FIRST credit
  decision, so the duty of care is highest here · **mode:** which/compare decision-trees ·
  **stance:** consumer champion
- **neighbours:** L1 (the numbers), L4/L6/L7 (product-specific depth)

### L3 — debt consolidation (unsecured)
- **primary:** consolidateloans.co.uk · **twins:** consolidateloans.uk ⚑OWNER · **also:**
  debtconsolidating.co.uk (⚑OWNER near-twin; angle if kept: **in-difficulty** register —
  DMPs, Breathing Space, StepChange signposting — vs the primary's cost-optimiser register)
- **audience:** multiple debts, wants one payment · **age:** 25–50 · **stage:** two
  distinct visitors — the optimiser tidying cards and the borrower already struggling; the
  site serves the first and hands the second to free debt help without a sales pitch ·
  **size:** £3k–£25k combined balances · **experience:** heavy credit use, often
  thin savings · **register:** reassurance, zero judgement
- **neighbours:** M8 (into the mortgage), M5 (adverse credit)

### L4 — long-term loans
- **primary:** longtermloan.co.uk · **twins:** longtermloan.uk ⚑OWNER
- **audience:** larger unsecured amounts, 5–10yr terms · **age:** 25–55 · **stage:**
  shopping, often for a single big purchase (car, wedding, home improvement) · **size:**
  £7.5k–£25k+ · **experience:** competent enough to be dangerous — comfortable with
  monthly-payment thinking, which is exactly the trap · **mode:** total-cost honesty (the
  term-vs-total-cost lesson is this site's whole reason)
- **neighbours:** L1, L6 (secured alternative)

### L5 — business finance
- **primary:** financeforcompanies.co.uk (fullest name) · **also:**
  business-lenders.co.uk (angle: the lender directory — who lends to whom, data-first),
  companyloan.uk (angle: the product explainer + calculator), businessfinances.uk (⚑OWNER
  generic; angle if kept: managing company money, not borrowing — cashflow, accounts)
- **audience:** SME director/founder, 1–50 staff · **age:** of the BUSINESS, not the
  person — startup (<2yrs, hardest to place) vs trading (asset/cashflow lending opens up) ·
  **stage:** growth purchase or cashflow gap — different products, the site must triage
  which early · **size:** £10k–£500k · **experience:** expert in their trade, novice in
  credit structures — competent register, no baby talk, define every instrument once ·
  **compliance:** mostly unregulated lending — different disclosure world, say so, and say
  what protections the director gives up vs personal credit (esp. personal guarantees).
- **neighbours:** M12 (premises), L8 (asset finance), I2 (key person cover)

### L6 — secured / specific-structure personal credit
- **primary:** unsecuredpersonalloans.co.uk
- **audience:** borrower choosing security structure — usually a homeowner being offered
  a cheaper *secured* rate and wondering what the catch is (the catch is the house) ·
  **age:** 30–55 · **stage:** shopping at the £10k–£50k boundary where both products bid ·
  **size:** £10k–£50k · **experience:** competent · **mode:** what-is + which (secured vs
  unsecured is the site's core question, argued from the *unsecured* side)
- **neighbours:** M8/L3 (consolidation), L4

### L7 — investment borrowing
- **primary:** investmentloan.uk · **twins:** investmentloans.uk ⚑OWNER
- **audience:** borrowing to invest — property deposits, business stakes, occasionally
  markets · **age:** 30–60 · **stage:** planning a deliberate leverage decision, not an
  urgent need — the ONE loan family where the visitor has time · **size:** £25k+ ·
  **experience:** expert, or believes so — the site's duty is to the second group ·
  **register:** risk-forward — leverage cuts both ways is the editorial spine
- **neighbours:** M9 (BTL specifically), B7 (rateforecast — cost-of-carry)

### L8 — asset finance (high-value, niche)
- **primary:** fleetfinancing.co.uk (B2B vehicles) · **also:** yachtfinancing.co.uk
  (HNW marine — genuinely separate market), mobilehomeloans.co.uk (park homes — NB often
  NOT mortgageable land, a genuinely underserved niche)
- **audience:** per domain, three genuinely different markets — `fleetfinancing`: a B2B
  operations/fleet manager, 5–100 vehicles, £50k–£1m+, competent-to-expert, cost-per-mile
  thinking; `yachtfinancing`: HNW buyer, 45+, £100k–£5m, aspiration register, marine
  survey/VAT/flagging complexities are the content; `mobilehomeloans`: park-home buyer,
  50+, often downsizing retirees, £50k–£200k — and the load-bearing fact that park homes
  usually sit on land you do NOT own, so they are often not mortgageable at all; a
  genuinely underserved audience · each is its own micro-authority; they share chassis,
  not content
- **neighbours:** L5, M11

### L10 — the borrower's FCA rulebook (OWNER DIRECTION 2026-08-01) — BUILDING
- **primary:** loancash.co.uk
- **history:** this register originally recommended NOT building it (payday-adjacent name
  attracts the audience FCA rules protect hardest). **The owner's ruling inverts the risk
  into the point: build it AS the protection** — the person searching "loan cash" lands on
  their rights, not on a lender.
- **audience:** the high-cost, small-sum, urgent borrower — payday/HCSTC, doorstep,
  rent-to-own, logbook, guarantor territory · **experience:** often already borrowing ·
  **stage:** about to borrow, or already in trouble · **register:** protective, zero
  judgement, urgent-friendly (short answers first)
- **mode:** rights-first guides + rule-checking tools (price-cap checker, true-cost
  calculator, complaint-deadline calculator) — a civilian champion of the FCA rulebook:
  the price cap, affordability duties, CPA limits, rollover limits, complaint rights,
  authorised-lender checks, loan-shark reporting, and the free-help routes
- **stance:** consumer champion, explicitly **pro-regulator** — the demand-side enforcer:
  readers who know the caps and complain are how the rules bite
- **compliance:** NOT a lender, NOT a broker, no applications, no lead-gen, and —
  load-bearing — **independent of and not affiliated with the FCA**, stated plainly on
  every-page chrome; regulatory constants (0.8%/day, £15, 100%) are quoted WITH the rule
  they come from and a check-the-source pointer, as the one sanctioned exception to the
  no-quoted-numbers house rule (they are rules, not market rates)
- **neighbours:** L2 (choosing between mainstream loans → theirs; this site is rights,
  not choice), L3 (consolidation depth → theirs; this site signposts), M5 (adverse-credit
  *mortgages* → theirs). Rule: if the question is "which product", link out; if the
  question is "what are they allowed to do to me", it lives here.

### L9 — loan brandables (direction unassigned, deliberately)
- **domains:** loansy.uk, loanzy.co.uk, loanzy.uk, finance.org.uk, banker.co.uk, banker.uk
- **status:** loansy/loanzy remain HOLD (product-brand candidates). **P7 assigned the
  other two** (see the seat map): `banker.co.uk` = the setter-seat trade title — the
  estate's industry-facing voice and link engine (`banker.uk` per P5); `finance.org.uk` =
  the referee/education hub — the .org trust weight spent as the estate's citation anchor.

## Family: BANKING / SAVINGS / RATES

### B1 — savings rates: the live table
- **primary:** savingsrates.co.uk · **twins:** savings-rates.co.uk, savingsrates.uk,
  savingrates.co.uk, saving-rate.co.uk† ⚑OWNER (†see B6 — different sense available)
- **audience:** saver checking the market · **age:** broad 25–75, skewing 45+ (that is
  where UK cash savings actually sit) · **stage:** recurring management, monthly-ish ·
  **size:** £1k–£85k — and the £85k FSCS protection boundary is standing content, because
  it is the one number that changes a saver's behaviour · **experience:** mixed; tables
  must work for a novice without insulting an expert · **mode:** data-first, visibly
  dated · **freshness is the product** — do not build without a data pipeline plan
- **neighbours:** B2, B3, B4, B7 — this family is the most collision-prone in the
  portfolio; every entry's out-of-scope line matters

### B2 — the rate chaser
- **primary:** highestinterest.co.uk · **also:** highestrate.co.uk (angle: cross-product —
  savings vs premium bonds vs gilts vs money-market funds), highestsavingsrate.co.uk
  (⚑OWNER near-twin of B2 primary)
- **audience:** the optimiser who actually moves money — a minority of savers and they
  know it · **age:** 30–60 · **stage:** recurring, trigger-driven (maturity dates, rate
  cuts) · **size:** £10k–£85k+, often multi-account · **experience:**
  competent-to-expert; wants the method shown and the loyalty-penalty maths made explicit ·
  **risk posture:** optimiser · **register:** vigilance — loyalty penalties, when to
  switch · **mode:** best/ranked with methodology
- **neighbours:** B1 (the table), B7 (whether to fix)

### B3 — savings by institution type
- **primary:** buildingsocietyrates.co.uk (mutual sector; older, branch-loyal audience) ·
  **also:** banksavingsrates.co.uk (big banks vs challengers), banksrates.co.uk +
  banks-rates.co.uk (⚑OWNER twins; broader bank-pricing angle: savings AND borrowing at
  the big banks)
- **audience:** chooses the institution first · **counterparty axis is the identity** ·
  two distinct markets inside it — `buildingsocietyrates`: 55+, branch-loyal, values the
  mutual model and being able to walk in; often holds larger, longer-standing balances;
  register respectful of that loyalty while showing its price · `banksavingsrates`: 25–45,
  challenger-curious, wants to know why the high street pays less and whether the app
  bank is safe (FSCS answer front and centre) · **stage:** recurring · **experience:**
  novice-to-competent
- **neighbours:** B1, B2

### B4 — savings by account type
- **primary:** savingsaccountrates.co.uk
- **audience:** the tax-aware saver choosing a STRUCTURE, not a rate · **age:** 30–65 ·
  **stage:** annual (ISA season) plus life events — redundancy money, inheritance, house
  sale proceeds parked · **size:** near or above the personal savings allowance and ISA
  limits, where the wrapper starts to out-earn the rate · **experience:** competent ·
  **mode:** which — ISA vs fixed bond vs easy-access vs notice; the wrapper logic, not
  just the headline number
- **neighbours:** B1 (numbers), pensions/investment ground NOT owned here (out of scope)

### B5 — interest rates (macro)
- **primary:** interestrates.co.uk · **also (P7 seats — see the map):**
  moneyrates.co.uk = observer, the cross-product price-of-money tracker;
  competitiverates.co.uk = setter, how providers benchmark; ratecomparison.co.uk =
  referee-facing-buyer, how to compare honestly (APR vs AER, the 51% representative-rate
  rule)
- **audience:** wants to understand the base rate and what it does to them — arrives from
  a news headline, not a product search · **age:** broad · **stage:** understanding before
  acting; this family's job is to hand informed visitors to the acting sites (M3, M4, B1) ·
  **experience:** novice-to-competent; the explainer register, not the data table ·
  **mode:** what-is + explainer, MPC-date aware · **stance:** data journalist
- **neighbours:** B7 (prediction — B5 explains, B7 forecasts), M3 (mortgage pricing)

### B6 — your saving rate (different sense)
- **primary:** saving-rate.co.uk — IF the owner keeps it out of B1's twin set
- **audience:** the habit-builder — what % of income to save, emergency fund sizing ·
  **age:** 20–40, first proper income years · **stage:** planning; the only B-family site
  aimed BEFORE the money exists · **size:** £50–£500/month flows, not balances ·
  **experience:** novice, and the register is coach-like rather than market-facing ·
  **mode:** calculator + workbook
- **neighbours:** B1 (product rates), pensions (out of scope)

### B7 — rate forecast (predictive — unique)
- **primary:** rateforecast.co.uk
- **audience:** anyone timing a fix (mortgage or savings) — the one visitor both M and B
  families share · **age:** 30–60 · **stage:** a dated decision ahead of them (fix ending,
  bond maturing) · **size:** decisions worth timing — mortgage-scale or £20k+ savings ·
  **experience:** competent-to-expert; wants the reasoning shown, tolerates uncertainty
  stated as uncertainty · **mode:** predictive — market expectations, swap curves in
  plain English, scenario tools · **stance:** data journalist
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
- **audience:** the new mortgagor being offered payment cover at completion · **age:**
  25–45 · **stage:** the fortnight around completion, when every add-on gets said yes to ·
  **size:** small premium, mortgage-scale consequence · **experience:** novice in
  protection products, and time-pressured — the two conditions mis-selling feeds on ·
  **register:** duty, unhurried · **compliance:** MPPI's mis-selling history is part of
  the honest story — say it
- **neighbours:** I4 (income protection — the usually-better answer; this site must say so)

### X2 — HNW private finance
- **primary:** private-finance.uk
- **audience:** HNW whole-balance-sheet — lombard lending, private banks, structured
  credit · **age:** 40–70 · **stage:** ongoing advisory relationships, not transactions ·
  **size:** £1m+ balance sheets · **experience:** expert, or professionally advised —
  the site talks to the adviser conversation, not around it · **register:** discreet
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
- **audience:** the individual priced off NHS waiting lists — most are FIRST-TIME private
  medical buyers · **age:** 30–60, skewing 45+ where premiums and worries both rise ·
  **stage:** triggered shopping — a diagnosis, a waiting-list letter, a birthday premium
  jump · **size:** meaningful monthly premium, rising with age — the age-pricing curve is
  content, not small print · **experience:** novice; underwriting jargon (moratorium vs
  full underwriting, pre-existing conditions) is where buyers get hurt, so it is where the
  site earns its keep · **register:** duty + plain English · **mode per grammar:**
  quote-names = estimator-first; rate-names = pricing data; best-names = ranked with
  methodology
- **compliance:** FCA ICOBS territory — exclusions prominent, pre-existing-condition rules
  stated, no advice implied. This line applies to ALL I-family rows.
- **neighbours:** I2 (family), I3 (employer-paid), the healthcare† angle

### I2 — PMI, family
- **primary:** comparefamilyhealthinsurance.co.uk · **twins/near:**
  comparefamilyhealthinsurance.uk, familyhealthinsurancequote.co.uk/.uk ⚑OWNER
- **audience:** parents costing cover for the household, children included · **age:**
  30–50 · **stage:** usually triggered by a child's health scare or a second income
  arriving · **size:** family premiums — the child-pricing and who-needs-covering
  trade-offs ARE the content · **experience:** novice · **mode:** compare, family-shaped
  worked examples
- **neighbours:** I1 (individual), I3

### I3 — PMI, employer-paid
- **primary:** corporatehealthinsurance.uk · **also:** staffhealthinsurance.co.uk/.uk,
  workerhealthinsurance.co.uk/.uk (⚑OWNER twins; angle if kept: SME <50 heads vs
  corporate schemes), keyworkerinsurance.co.uk/.uk (**different buyer** — the public-sector
  individual: NHS/teachers/police discounts and occupational quirks; keep separate)
- **audience:** HR/founder buying for staff · **company size is the real axis** —
  `staffhealthinsurance`/`workerhealthinsurance`: the 5–50-head SME where the founder IS
  HR and price sensitivity is total; `corporatehealthinsurance`: 50+ heads, scheme
  structures, P11D/benefit-in-kind tax treatment as first-class content · **stage:**
  benefits round or a retention scare · **experience:** competent professionally, novice
  in group-scheme mechanics · the keyworker pair: the WORKER themselves (NHS, teachers,
  police), 25–55, buying individually around occupational cover — a different buyer
  entirely, kept deliberately distinct from I8's business "key person"
- **neighbours:** I1, L5 (business costs), I8 (key person — the OTHER "key" name)

### I4 — income protection
- **primary:** incomeprotectioncover.co.uk · **twins:** incomeprotectioncover.uk ⚑OWNER
- **audience:** the self-employed and sole household earners — the people statutory sick
  pay does not reach · **age:** 25–55; the earlier bought, the cheaper, and saying so
  plainly is the site's best service · **stage:** often triggered by a peer's illness or
  going self-employed · **size:** cover sized to committed outgoings, not income vanity ·
  **experience:** novice; deferral periods and own-occupation definitions are the two
  places buyers get quietly shortchanged · **register:** duty · the anti-PPI story — what
  MPPI should have been
- **neighbours:** X1 (mortgage-payment cover — narrower product this one usually beats),
  I6 (life)

### I5 — landlord insurance
- **primary:** landlordinsurancerates.co.uk · **cluster:**
  bestlandlordinsurancerate(s).co.uk/.uk, landlordinsurancedeal(s).co.uk/.uk,
  landlordinsurancerate.co.uk (⚑OWNER grammar twins) · rentalsinsurance.co.uk/.uk (angle:
  tenant-side / rent-guarantee — arguably a different buyer; hold for owner call)
- **audience:** landlord — same person M9 serves; **cross-link M9↔I5 is the portfolio's
  best internal seam** · **age:** 35–65 · **stage:** annual renewals ×N properties, plus
  every new purchase · **size:** 1–10 properties; the accidental landlord (inherited or
  couldn't-sell) is a distinct novice segment the *rates* grammar underserves ·
  **experience:** competent on property, novice on policy detail — loss-of-rent cover,
  tenant-type exclusions and unoccupancy clauses are where claims die, so they are the
  content spine
- **neighbours:** M9, I7 (home — owner-occupier, different policy type)

### I6 — life insurance
- **primary:** lifeinsurancequotation.co.uk · **twins:** lifeinsurancequotation.uk,
  lifeinsurancerate.co.uk/.uk, lifeinsurancedeal.uk ⚑OWNER
- **audience:** new parents and new mortgagors — the two moments life cover actually gets
  bought · **age:** 25–45 · **stage:** a life event weeks old · **size:** cover sized to
  the mortgage and the years to independence, term-shaped; level vs decreasing term is
  the site's first fork · **experience:** novice; trust-and-nomination basics and the
  in-trust question are the highest-value content because nobody else explains them
  plainly · **register:** duty · **mode:** quote-first
- **neighbours:** I4 (living-benefit cover), X1

### I7 — home insurance
- **primary:** homeinsurancerates.co.uk · **twins/near:** homeinsurancerate.uk,
  homeinsurancedeal.co.uk/.uk, cheaphouseinsurance.uk (price-first entry — angle: the
  cheap-seeker, excess/cover trade-offs stated honestly) ⚑OWNER
- **audience:** owner-occupier at renewal · **age:** 30–70 · **stage:** the renewal
  letter fortnight — urgency is real but artificial, and saying so is the site's angle ·
  **size:** rebuild-cost sums, and under-insurance via wrong rebuild figures is the
  quiet catastrophe worth owning · **experience:** novice, price-led — `cheaphouseinsurance`
  serves the price-first entry honestly by pricing the excess/cover trade-offs it hides ·
  **frequency:** annual · **mode:** rates/compare
- **neighbours:** I5 (landlord), I1 (grammar cousins)

### I8 — key person cover (business)
- **primary:** keypersoncover.co.uk · **twins:** keypersoncover.uk ⚑OWNER
- **audience:** SME director insuring the business against losing a person · **buyer:**
  company — NOT the keyworker pair (individual) · **age:** company 2+ years old, past
  survival mode · **stage:** often lender-triggered (a bank requiring cover on a founder
  as a loan condition — the L5 cross-link) · **size:** £100k–£1m sums assured ·
  **experience:** competent commercially, novice in the tax treatment of premiums and
  payouts, which is the content that converts
- **neighbours:** I3/keyworker (the collision-in-waiting; both rows carry the distinction)

### I9 — professional indemnity
- **primary:** indemnitycover.co.uk · **also:** journalistinsurance.co.uk/.uk (worked
  niche example: PI for journalists — libel/defamation; either the PI site's flagship
  vertical or its own micro-site)
- **audience:** freelancer/consultant required to hold PI — usually by a client contract,
  so the trigger is a contract clause, not a fear · **age:** 25–60 · **stage:** urgent
  (a contract waiting on a certificate) or renewal · **size:** £1m/£2m/£5m indemnity
  bands; matching the band to the contract's demand is the ten-second answer the site
  leads with · **experience:** expert in their profession, novice in cover — run-off
  cover after closing down is the trap nobody warns them about · the `journalistinsurance`
  pair: libel/defamation-weighted PI for a niche whose risk is the WORK ITSELF, either
  this site's flagship vertical or its own micro-site
- **neighbours:** I8, L5

### I10 — insurance generic / brandables
- **domains:** bestinsurancedeal(s).co.uk/.uk, bestinsurancerate(s).co.uk/.uk,
  cheapinsurancequote.uk, insuranceapp.co.uk/.uk (product — HOLD like B8),
  reviewsinsurance.co.uk/.uk (angle: the evaluation layer — reviews of insurERS not
  prices; genuinely distinct mode), sportsreviewinsurance.co.uk/.uk (unclear; hold)
- **status:** assign after the exact-match rows are built, same rule as L9.

---

## Twin-pair rulings (⚑OWNER rollup)

**OWNER RULING 2026-08-01 — "split the co.uk and .uk": cross-TLD pairs are NOT 301s.**
This supersedes the register's original default (301 to the built sibling) for every
`.co.uk`/`.uk` pair of the same phrase. Both get built, and since no *audience* axis
exists between two spellings of one phrase, the split runs on the **depth-contract and
mode axes**, applied as one portfolio-wide convention rather than 40 improvisations:

- **`.co.uk` = the authority.** The thick site: guides, methodology, the 20-minute answer.
- **`.uk` = the instrument.** Tool-first, mobile-first, the 10-second answer — one
  calculator or checker doing the phrase's job immediately, linking to its `.co.uk`
  sibling for the depth. Modest by design, but real: not a doorway page, a working tool.

The pair cross-link as **siblings, not mirrors** (each canonical points at itself; the
`.uk` links "the full guide" out, the `.co.uk` links "the quick tool" out). When a `.uk`
twin is built it gains its own register entry (suffix `u`, e.g. `M7u`) claiming the
domain out of the parent's claims line — the collision check holds because the mode
coordinate differs by convention.

**OWNER RULING 2026-08-01 (P7) — same-TLD twins are split by SEAT, and the last ⚑OWNER
defaults are retired.** The register's original claim — "within one TLD the two spellings
have no axis at all" — held one thing constant without noticing: the reader's **seat at
the market's table**. Every earlier axis segmented *consumers*. The owner's insight is
that the same phrase means different things to different SEATS: a shopping list to a
saver, a pricing decision to the bank setting the rate, a commission structure to a
broker, a data series to an analyst, a body of rules to a compliance officer. Those are
different audiences with different search behaviour, so the collision invariant holds.

**The seats:** buyer (consumer — the default everywhere else in this register) · setter
(provider-side: how the product is priced) · intermediary (brokers, platforms, advisers —
including commission transparency, which is near-unoccupied ground) · observer (analysts,
journalists, treasurers — rates as data) · referee (compliance, case law, the rulebook).
**The archive is a SECTION, not a seat** (owner: "I am still after the money") — history
pages earn citations cheaply and citations feed the buyer sites' rankings, but a
standalone archive site has no transaction anywhere near it.

**The money, because every seat must have some** (owner correction applied — commercial
intent ranks the seats, it does not exclude them): buyer = affiliate/lead-gen, the
estate's direct revenue; intermediary = B2B affiliate + adviser lead-gen; referee =
professional readership + adviser/solicitor lead-gen; setter/observer = press citations
and authority. **And the authority ROUTE is explicit policy:** professional-seat sites
link contextually down to the buyer domains; buyer sites cite up. The professional sites
are simultaneously their own B2B money and the estate's link engine.

**One correction to "mainly buyer unless it's hard to differentiate":** for same-PHRASE
twins it is ALWAYS hard — two buyer sites on one phrase chase the same search results and
split each other's authority, which is the exact self-competition this register exists to
prevent. So the operating rule is: **exactly one buyer per phrase; every additional
spelling takes a different seat or waits.** Different-phrase domains keep buyer as the
default, unchanged.

**Allocation rule (owner-confirmed):** the cleanest spelling takes the buyer seat, because
buyers type domains; professional/referee positions are reached by question-search and
citation links, where memorability barely matters — so hyphens and awkward variants take
the seats, turning the defect into the asset.

### The seat map (authoritative — supersedes any per-entry ⚑OWNER twin label above)

| domain | seat | the site |
|---|---|---|
| `savingsrates.co.uk` | buyer | B1 unchanged — the live consumer table |
| `savingsrate.co.uk` | setter/observer | "THE savings rate": what banks depend on to set it — base-rate passthrough, swap rates, funding pressure, front-book/back-book gap |
| `savings-rates.co.uk` | referee | the savings rulebook: FCA cash savings market review + remedies, Consumer Duty fair value on cash, rate-change notification rules, variation-clause case law, ombudsman decisions (per-vertical casebook, owner-chosen shape) |
| `savingrates.co.uk` | intermediary | acting on SETS of rates: laddering, cash platforms and how they earn, FSCS-spreading for £100k+, the SME/charity/school treasurer's handbook |
| `highestsavingsrate.co.uk` | buyer-adjacent w/ archive section | the record as a citable series feeding B2; not a standalone archive |
| `banksrates.co.uk` | buyer | B3 unchanged — big-bank consumer comparison |
| `banks-rates.co.uk` | observer (institution grain) | passthrough accountability, bank by NAMED bank — distinct from `savingsrate.co.uk`'s market aggregate |
| `poorcreditmortgages.co.uk` | buyer | M5 unchanged — plain-English adverse-credit |
| `poor-credit-mortgages.co.uk` | intermediary | the adverse-credit BROKER's criteria handbook — which lender tolerates what, case-placement intelligence |
| `mortgageextension.co.uk` | buyer | M7 unchanged — extend MY term, the decision |
| `mortgageextensions.co.uk` | observer/policy | term extensions as a phenomenon — Mortgage Charter, arrears data, maturity creep |
| `equityreleasecalculator.co.uk` | buyer | M6 unchanged — the consumer calculator |
| `equity-release-calculator.co.uk` | intermediary/referee | the adviser's compliance toolkit — advised-sale rules, ERC standards, suitability documentation |
| `fixedmortgagerates.co.uk` | buyer | when to fix, fix-length trade-offs |
| `fixedinterestmortgagerates.co.uk` | setter | how lenders PRICE a fix off the swap curve — why the 2y/5y move before base |
| `investmentloan.uk` | buyer | my leverage decision (P5 instrument register) |
| `investmentloans.uk` | observer | the market overview of investment-lending products |
| `healthinsurancerates.co.uk` | buyer | market tables (plural = the market) |
| `healthinsurancerate.co.uk` | setter-facing-buyer | pricing factors: what moves YOUR premium — age curve, postcode, underwriting (singular = yours) |
| `besthealthinsurancerates.co.uk` | buyer (verdicts) | ranked with methodology |
| `besthealthinsurancerate.co.uk` | — residual | true twin of the two above even under seats; hold/301 ⚑OWNER |
| `landlordinsurancerates.co.uk` | buyer | I5 unchanged — the tables |
| `landlordinsurancerate.co.uk` | setter-facing-buyer | landlord pricing factors — tenant type, unoccupancy, claims history |
| `landlordinsurancedeal.co.uk` | buyer (switching) | the switching playbook, singular = your move |
| `landlordinsurancedeals.co.uk` | intermediary | commission transparency: how landlord policies are sold and who earns what |
| `bestlandlordinsurancerates.co.uk` | buyer (verdicts) | ranked with methodology |
| `bestlandlordinsurancerate.co.uk` | — residual | hold/301 ⚑OWNER |
| `bestinsurancerates.co.uk` | buyer (verdicts) | cross-line ranked tables |
| `bestinsurancerate.co.uk` | setter-facing-buyer | cross-line pricing factors — what drives premiums in every line |
| `bestinsurancedeal.co.uk` | buyer (switching) | the cross-line switching playbook |
| `bestinsurancedeals.co.uk` | buyer (offers) | time-boxed offers listing |
| `loanscentre.co.uk` | intermediary | the loan-broker layer: how brokers earn, whether to use one, commission disclosure — pro-consumer AND trade |
| `moneyrates.co.uk` | observer (cross-product) | every money price tracked as data — savings, mortgage, loan, cards; the estate's price-of-money tracker |
| `competitiverates.co.uk` | setter | what makes a rate "competitive" — how providers benchmark and chase the tables |
| `ratecomparison.co.uk` | referee-facing-buyer | how to compare rates HONESTLY — APR vs AER, the representative-rate rule (only 51% must get the advertised rate), teaser mechanics |
| `banker.co.uk` | setter (trade title) | the industry-facing brand — the estate's trade voice and link engine; `banker.uk` per P5 |
| `finance.org.uk` | referee/education hub | the .org trust weight spent deliberately: neutral consumer-finance education, the estate's citation anchor |

**Honest residuals, still ⚑OWNER (hold or 301):** `besthealthinsurancerate.co.uk`,
`bestlandlordinsurancerate.co.uk` — even five seats cannot make seven members of one
phrase family distinct; two remain true twins. Plus the product holds (`savingsapp.*`,
`insuranceapp.*`, `loansy`/`loanzy`, `sportsreviewinsurance.*`) and `bankingequipment.*`
(different market). **No lane builds a residual without this register changing first.**

**Per-vertical casebooks (owner-chosen):** the referee seat is one substantial casebook
per vertical — savings law at `savings-rates.co.uk`, the mortgage-side law living inside
M-family sites until a domain warrants it, insurance claims/ombudsman case law as the
evaluation layer's depth (`reviewsinsurance.co.uk` already holds the evaluation seat).
Referee sites are information ABOUT law, never advice, sources cited, solicitor
signposting — 3–4 thick casebooks beat 15 thin ones, and thin legal content is the
riskiest thin content there is.

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
L2: whichloan.co.uk whatloan.co.uk borrowing.co.uk loanscentre.co.uk
L10: loancash.co.uk
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
