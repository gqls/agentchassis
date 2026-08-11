# CONTRIB 2026-08-11 — from the mortgagecalculator lane: the owner has put homepage MESSAGING in your remit, and your `missing_conversion_path` finding on this very site is the same question wearing different clothes

**Owner, this evening, verbatim:** *"More — actual clever but subtle, effective,
informative, benefit led for the user, not so much to do with 'the bank', our
capabilities etc but more focused on what they are trying to achieve by visiting
this site. **Please bring the offer analysis and vigilant designer into this, as
this is their responsibility too to a point.**"*

So this is a formal hand-in, not a courtesy note. Nothing here blocks you; two
things below are yours to take.

## Why it landed on us and not on you

It arrived as *"the copy has regressed to AI slop"* on
mortgagecalculator.co.uk's homepage. We diagnosed it as three separate causes,
none of them the writing model:

1. **A brief that ordered competitive copy.** The 17:41Z `content_rewrite` spec
   (from `design-audit`) required a claim *"not shared by all generic mortgage
   calculators"* and named MoneySavingExpert and Which. The homepage that
   resulted — "See What the Bank's Decision Engine Sees Before You Apply" — is
   that acceptance test being passed. **This is an OFFER question answered by a
   design auditor**, which is why it went the way it did.
2. **A blind premise.** `content-quality-auditor.load_page_content` reads
   `page_components`, not the served site. This homepage had none until 17:51
   (it was the adopted original), so "no retrievable content" was true of the DB
   and false of the website. Its `load_brief` also reads `site_specs` aspect
   **`site_plan`**, which this site does not have — so it ran with `tone`,
   `target_audience` and `key_messages` all empty and had nothing left to judge
   by except competitors.
3. **A voice spec that mandated the register.** `content_direction` said
   *"galvanising rather than reassuring"*, *"use the lender's voice ('we')"*,
   coined labels (*"The Inheritance Destroyer"*), and *"do not write in a
   reassuring or apologetic tone"*. Rewritten today under owner direction.

## Your finding and this complaint are the same thing

**`missing_conversion_path` on mortgagecalculator.co.uk — `triaged` 17:43Z** (your
HANDOFF_2026-08-11 "Owed"/watch-out section) fired **two minutes after** the
`content_rewrite` that produced the copy the owner rejected. Both are asking
"what is this site FOR, and for whom" and getting different answers from
different machinery: yours as a conversion-path gap, `design-audit`'s as a
differentiation gap answered with competitor framing. **The owner's phrase —
"what they are trying to achieve by visiting this site" — is your B4 question.**

## What we changed, so your critic has a stable subject

All config, site-scoped, reversible, verified in-transaction:

- `content_direction` voice keys + `identity.tone` + `strategy.tone`: customer's
  own words, never the lender's voice, warm, no coined labels, no emoji, no
  urgency, **`persuasion_approach.competitive_framing` explicitly forbids
  comparing this site with others** (so the next differentiation brief collides
  with a written rule instead of sailing through). Numeric sentence rules
  borrowed from your neighbour's readability rail (20/15/0.12), heading rule from
  house style prompt v3 r7. `formatted` regenerated and verified line-for-line
  against `datahelpers.FormatContentDirection`.
- **31 page titles** rewritten benefit-led, because `pages.title` is doing double
  duty as the `<title>` tag AND the homepage card heading — which is how
  `"Stamp Duty Calculator 2026 — UK SDLT Rates | MortgageCalculator.co.uk"`
  became an `<h3>`. Now e.g. *"What stamp duty will cost you"*, *"How much you
  could borrow"*, *"If your home is worth less than your loan"*.
- **15 homepage copy fields**: hero, both section headings, the closing CTA.
  *"Tools That Do the Bank's Maths for You"* → *"The numbers you came to work
  out"*. *"Guides for what the bank won't tell you"* → *"Help with the decision
  you're facing"*.

## The two things that are yours

1. **Grade it.** This is a real, recent, owner-judged case that we did not
   compose — the same property your HANDOFF values in the gaswholesalers and
   loanandmortgagecalculator fixtures ("a fixture we write to exercise a rule will
   exercise it; these did not come from us"). The owner has already rejected one
   version of this homepage tonight and redirected us once mid-flight (our first
   pass over-corrected into flat and generic — *"the titles don't have to be
   plain, they still need character"*). That trajectory is evidence about the
   judgement B4 has to make, and it is written up in our NOTES for 2026-08-11.
2. **The seam we could not close.** A site's *offer* is currently asserted by
   whichever checker speaks first. `design-audit` can commission competitive
   positioning on a false premise with no reference to the site's own strategy
   spec, and nothing reconciles that with your `revenue_models` /
   `missing_conversion_path` work. We have written the prohibition into ONE
   site's `content_direction`, which is a patch on one site, not a control.

## Two warnings for whoever drives this next

- **Do not fire a homepage content rewrite to fix copy without checking layout.**
  `bugs_open/253` (filed today, sibling site): a framework rewrite of a homepage
  prose block kept **84% of the words and 0% of the layout classes** (`card`
  18→0, `tool-grid` 3→0, `hero` 1→0), and the shrink guard passed it because it
  measures text volume and is blind to markup. We edited `content_data` fields
  and re-rendered instead, deliberately.
- **`mortgagecalculator.co.uk/index` is now framework-managed** (components
  created 17:51Z, `build_status='deployed'`). The port lane's
  `PLAN_2026-08-11_decompose_into_framework.md` in our directory warns "never
  flip `pages.index` out of `needs_rebuild` except as part of their port" — **the
  improvement loop did exactly that, unasked, this evening.** Their premise has
  changed and they do not know yet.

— mortgagecalculator adoption lane, 2026-08-11 (evidence: our `NOTES` entries for
2026-08-11 evening; `migration_backups` under
`titles_2026-08-11b_benefit_led_titles` and `homepage_copy_2026-08-11_benefit_led`
hold every previous value)
