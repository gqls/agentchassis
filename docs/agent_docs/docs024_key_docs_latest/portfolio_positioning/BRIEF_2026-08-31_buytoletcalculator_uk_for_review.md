# Brief for review — buytoletcalculator.uk

> Rendered 2026-08-31 from the live row for reading only. The database row is authoritative
> (`site_specs` aspect `mission_brief`, `is_current`); this file is a snapshot and is
> not read by anything. Nothing builds until the review item is released.

## Proposition

buytoletcalculator.uk is a calculator-first tool site for UK landlords and property investors who need to run the two numbers that govern every BTL mortgage decision: rental yield and the interest coverage ratio (ICR). The site exists because these two calculations — though mechanically simple — are poorly explained and inconsistently applied across lenders, and the landlord who understands them negotiates better, borrows more efficiently, and avoids the deals that quietly fail stress tests. The site serves the competent-to-expert landlord: someone already in, or seriously entering, the BTL market, who wants to stress-test a deal, understand how much a given property will support in borrowing, and know what the ICR looks like under different rate and coverage scenarios — all before speaking to a broker.

## Audience

**Primary:** UK landlords and property investors — from first-property buyers to portfolio holders — at the moment of evaluating a specific deal: they have a property in mind, a rental estimate, and they need to know whether the numbers stack up for a BTL mortgage before committing further time or money
**Secondary:** Mortgage brokers and intermediaries who want a quick, transparent tool to sense-check a client's deal or explain ICR mechanics to a landlord client

## Reader intent

- Calculate the maximum BTL mortgage loan amount a given property and rent will support, under realistic ICR coverage ratios
- Understand what rental yield a property produces and whether it meets typical lender thresholds
- See how the ICR calculation changes between individual borrower (typically 145%) and limited company / SPV (typically 125%) scenarios
- Find out what monthly rent is needed to support a desired loan amount at a given interest rate
- Understand the difference between gross and net rental yield and what each means for mortgage eligibility
- Learn what interest coverage ratio means, how lenders apply it, and why it differs by borrower type and fix length
- Work out how a rate rise or stress rate changes the maximum borrowing on a property they already hold
- Understand the basics of BTL mortgage mechanics — how rent-based lending works, what lenders look at, and what affects the deal

## Stance

The site is for landlords who want to understand their numbers, not just be told a figure — it is pro-transparency, anti-black-box, and treats the ICR calculation as something every landlord should be able to do themselves; it is against the industry habit of hiding the mechanics behind broker mystique.

## Differentiation

The main competitors surfaced by research are: (1) lender/intermediary tools like Skipton Intermediaries' BTL calculator — accurate but tied to one lender's criteria and designed for brokers, not landlords; (2) investment-focused tools like London Central Portfolio's calculator — oriented toward capital growth and ROI modelling, not mortgage affordability; (3) specialist lender tools like MFS — complex-case focused and lender-branded. This site's differentiation is neutrality (no lender affiliation, no broker lead-generation), specificity to the ICR/yield mechanics that govern BTL mortgage borrowing, and the depth of its explanatory content for the competent-to-expert landlord who wants to understand the calculation not just receive a number. The site deliberately does not serve the general homebuyer (M2 territory), does not compare rates (M3 territory), and does not go deep on company/SPV structures as a mortgage product (M12 territory).

## Content plan

-
  **Kind:** tool
  **Name:** BTL Mortgage Borrowing Calculator
  **What:** Core calculator: user inputs property value, monthly rent, interest rate, ICR coverage ratio (with lender-typical presets), and borrower type (individual / limited company). Output: maximum loan amount, LTV, and minimum rent needed to support a target loan. This is the primary reason anyone visits.
  **Priority:** core
-
  **Kind:** tool
  **Name:** Rental Yield Calculator
  **What:** Inputs: purchase price (or value), monthly rent, estimated annual costs (management fees, voids, insurance, maintenance). Outputs: gross yield, net yield, and a plain-English verdict on whether that yield is likely to satisfy mainstream BTL lenders. Bridges the gap between 'investment return' thinking and 'mortgage eligibility' thinking.
  **Priority:** core
-
  **Kind:** tool
  **Name:** ICR Stress Test Calculator
  **What:** Shows how the same property's borrowing capacity changes as the stress rate moves — e.g. at 5%, 6%, 7%, 8%. Particularly useful for landlords on tracker rates or approaching a fix end who want to know their refinance headroom. Outputs a table of max loan vs stress rate.
  **Priority:** core
-
  **Kind:** tool
  **Name:** Minimum Rent Needed Calculator
  **What:** Reverse of the borrowing calculator: user inputs target loan amount, interest rate, and ICR ratio; tool outputs the minimum monthly rent required to support that borrowing. Useful when assessing whether a property asking price is viable given its rental market.
  **Priority:** core
-
  **Kind:** guide
  **Name:** What Is the Interest Coverage Ratio — and Why Does It Control Your BTL Mortgage?
  **What:** The definitive plain-English explanation of ICR: what it is, why lenders use it (PRA regulation post-2017), why individual borrowers face a higher ratio than limited companies, why five-year fixes often get a more lenient stress rate, and what a landlord can do to improve their ICR. This is the most-searched explainer topic in this space.
  **Priority:** core
-
  **Kind:** guide
  **Name:** Gross vs Net Rental Yield: What Lenders Care About and What You Should
  **What:** Explains the two yield figures, what each includes, which one lenders use for affordability assessment, and what a 'good' yield looks like in different UK markets. Includes worked examples with realistic cost assumptions.
  **Priority:** core
-
  **Kind:** guide
  **Name:** Individual vs Limited Company BTL: The ICR and Tax Trade-off Explained
  **What:** Explains why ICR coverage ratios differ (125% vs 145% typical), how Section 24 mortgage interest tax relief changes interact with structure choice, and what factors push a landlord toward a company structure from a borrowing-capacity perspective. Deliberately stops short of advising — this is the mechanics, not the decision.
  **Priority:** valuable
-
  **Kind:** guide
  **Name:** HMO and Multi-Unit BTL: How Lenders Calculate Borrowing Differently
  **What:** HMO and MUFBs often have different ICR requirements and specialist lender panels. This guide explains how room-by-room rent is treated, why some lenders cap HMO lending at 65% LTV, and what the borrowing calculation looks like in practice.
  **Priority:** valuable
-
  **Kind:** data
  **Name:** BTL ICR Reference Table: Typical Coverage Ratios by Lender Type and Borrower Profile
  **What:** A reference table showing the range of ICR ratios used across the BTL market — by borrower type (individual, limited company), fix term (2-year, 5-year), and property type (standard, HMO, MUB). Gives the landlord context for the calculator's default assumptions and shows the spread of market practice.
  **Priority:** valuable
-
  **Kind:** guide
  **Name:** How BTL Lenders Stress Test Your Mortgage Application
  **What:** Explains the PRA stress-testing rules introduced in 2017: why lenders use a notional stress rate rather than the product rate for individual borrowers on sub-5-year fixes, how the calculation works, and what changed for portfolio landlords (four or more mortgaged properties).
  **Priority:** valuable
-
  **Kind:** guide
  **Name:** Portfolio Landlord Rules: What Changes When You Have Four or More BTL Mortgages
  **What:** The PRA's portfolio landlord rules require lenders to assess the whole portfolio, not just the new property. This guide explains what that means in practice, what information lenders will want, and how it affects borrowing capacity at a portfolio level.
  **Priority:** valuable
-
  **Kind:** editorial
  **Name:** Why Most BTL Yield Figures Quoted in Property Listings Are Misleading
  **What:** A pointed piece explaining why advertised gross yields routinely overstate actual returns, what costs are typically omitted, and what a landlord should actually calculate before treating a yield figure as meaningful. Establishes the site's voice as a sceptical, investor-side tool.
  **Priority:** valuable
-
  **Kind:** data
  **Name:** BTL Stamp Duty Reference: Current Rates Including the Additional Dwelling Supplement
  **What:** A clean reference table of current SDLT rates for BTL purchases in England and Northern Ireland, with the additional dwelling surcharge applied. Includes worked examples for common purchase prices. Supports deal-level cost calculations without building a full stamp duty calculator (that would be scope creep).
  **Priority:** valuable
-
  **Kind:** guide
  **Name:** First-Time Landlord Guide: What You Need to Know Before Taking Out a BTL Mortgage
  **What:** Entry-level orientation for someone crossing from residential to BTL: what is different about BTL mortgage qualification, what lenders typically require (minimum income, existing home ownership rules, age limits), and how the yield/ICR calculation governs what you can borrow. Funnel entry point for new landlords who find the site via broad searches.
  **Priority:** valuable
-
  **Kind:** research
  **Name:** BTL Mortgage Affordability: What £X in Monthly Rent Can Actually Support in Borrowing
  **What:** A worked-examples reference piece showing — at different rent levels from £600/month to £3,000/month — what maximum loan the ICR calculation supports under typical individual and limited company scenarios at a range of interest rates. Gives landlords a quick orientation table before they run the calculator.
  **Priority:** aspirational
-
  **Kind:** tool
  **Name:** Portfolio ICR Checker
  **What:** For portfolio landlords subject to PRA portfolio rules: inputs multiple properties with their respective rents, loan amounts, and rates; computes the portfolio-level ICR to assess whether a new purchase is likely to pass the portfolio stress test. More complex than the single-property tools but serves the expert landlord segment.
  **Priority:** aspirational

## Tool opportunities

-
  **Why:** This is the single most useful calculation a BTL landlord needs and no neutral, non-broker site does it cleanly
  **Name:** BTL Mortgage Borrowing Calculator
  **Input:** Monthly rent, interest rate, ICR coverage ratio (preset or custom), borrower type (individual / limited company), fix term
  **Output:** Maximum loan amount, implied LTV against entered property value, whether the deal passes or fails typical lender thresholds
-
  **Why:** Every BTL investor needs to know their yield; the net figure requires realistic cost assumptions most tools skip
  **Name:** Rental Yield Calculator (Gross and Net)
  **Input:** Purchase price, monthly rent, annual costs (management %, void allowance months, insurance, maintenance)
  **Output:** Gross yield %, net yield %, verdict on typical lender yield thresholds
-
  **Why:** Landlords approaching remortgage or holding trackers need to model rate sensitivity; no clean tool for this exists outside broker portals
  **Name:** ICR Stress Test Calculator
  **Input:** Loan amount, multiple stress rate scenarios
  **Output:** Table of minimum rent needed at each stress rate; pass/fail verdict
-
  **Why:** The reverse calculation — essential when evaluating whether a purchase price is viable given local rental market
  **Name:** Minimum Rent Needed Calculator
  **Input:** Target loan amount, mortgage rate, ICR ratio
  **Output:** Minimum monthly rent required to support that borrowing
-
  **Why:** Portfolio landlords face whole-portfolio assessment under PRA rules — no consumer-facing tool exists for this
  **Name:** Portfolio ICR Checker
  **Input:** Up to 10 properties: rent, loan balance, current rate per property
  **Output:** Portfolio-level ICR, weakest properties, aggregate borrowing headroom

## Directory opportunity

A directory of BTL lender types would be marginal value here — the site is not a comparison or broker site. However, a reference section listing named specialist BTL lender categories (high-street, specialist, bridging, portfolio lenders) with their typical ICR ratios, LTV limits, and minimum loan sizes would serve as reference data rather than a lead-generation directory and would be distinct from a broker comparison. This is worth considering as a data table rather than a full directory.

## Regulated subject

True

## Must nots

- Must not present as a regulated mortgage adviser, broker, or lender — the site has no FCA authorisation and must not imply one
- Must not recommend specific mortgage products, specific lenders, or specific interest rates as suitable for a user's circumstances
- Must not generate or facilitate mortgage applications or enquiries routed to a lender or broker panel without explicit FCA-compliant infrastructure and disclosed authorisation
- Must not quote real-time or live mortgage rates as current market rates — any rate data must be clearly labelled as illustrative or for calculation purposes only
- Must not advise on whether a user should hold property personally or through a limited company — the tax and legal implications require regulated advice
- Must not advise on individual tax positions, Section 24 implications for a specific user, or capital gains tax liability
- Must not imply that calculator outputs constitute a mortgage offer, decision in principle, or lender commitment of any kind
- Must not collect personally identifiable financial information and process it in a way that constitutes regulated credit broking
- Must not drift into SPV/company mortgage product comparison (M12 — smbmortgages.co.uk territory) or landlord insurance content (I5 territory)

## Open questions

- Should the site carry any affiliate or broker referral mechanism — if so, the must-nots around regulated activity need to be revisited with legal input and the site's FCA position clarified
- What is the monetisation model — pure traffic/ad-supported, broker referral, or something else? This affects whether the ICR calculator should include a broker CTA and how the regulated-activity boundary is managed
- The portfolio landlord tool (four or more mortgaged properties) is aspirational but touches on complex PRA rules — confirm whether the owner wants to serve that more expert segment or focus on single-property landlords
- Should the site acknowledge the limited company / SPV route in its tools and guides, or defer fully to M12 (smbmortgages.co.uk)? The ICR difference between individual and limited company is core to the BTL calculation, so some coverage seems necessary — but the boundary needs deciding
- Is there appetite for a regularly-updated data section on ICR ratios across the market? This would require a maintenance commitment and a source methodology
- The domain is .uk — confirm whether content should be England/Wales-specific or cover Scottish and Northern Irish LBTT/LTT stamp duty variants

## Research quality

adequate — the research returned strong evidence of the subject matter (ICR calculations, BTL yield mechanics, lender tools) and confirmed the calculator-first orientation is correct, but returned no direct competitor audit of the specific UK .uk domain space, and the research sources are mostly third-party tools and YouTube explainers rather than the SERP for the exact target keywords

## Confidence

0.82

