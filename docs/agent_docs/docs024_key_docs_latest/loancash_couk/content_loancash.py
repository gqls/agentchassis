# content_loancash.py — the guide bodies for loancash.co.uk (register entry L10).
#
# Editorial rules, from the register row:
#   * One protection per guide: the rule, what a breach looks like, what to do next.
#   * Regulatory constants are quoted WITH the rule they come from (the sanctioned
#     exception); market rates are never quoted.
#   * Protective register, zero judgement. The reader may be in trouble RIGHT NOW, so
#     every guide front-loads the actionable answer and ends at the free-help routes.
#   * "Which loan should I get" questions link OUT (L2's ground). "What are they allowed
#     to do to me" lives here.
#   * England & Wales specifics (e.g. Breathing Space) say so; Scotland/NI flagged.

GUIDES = [
    ("the-payday-loan-price-cap", "The Payday Loan Price Cap",
     "0.8% a day, £15 in default fees, and never more than you borrowed. The three limits "
     "on high-cost short-term credit, and how to claim when they are broken."),
    ("check-your-lender-is-authorised", "Check Your Lender Is Authorised",
     "Two minutes on the FS Register tells you whether a lender is allowed to lend at "
     "all. If they are not, the law is dramatically on your side."),
    ("affordability-checks-what-lenders-must-do", "Affordability Checks: What Lenders Must Do",
     "A lender must check you can repay without hardship before lending. If they didn't "
     "and you suffered, you can claim the interest back — years later."),
    ("stopping-payments-the-cpa-rules", "Stopping Payments: the CPA Rules",
     "That 'continuous payment authority' is not unstoppable. Two failed attempts is the "
     "limit, part-grabs are banned, and your bank must cancel it when you say so."),
    ("rollovers-the-two-strike-rule", "Rollovers: the Two-Strike Rule",
     "A payday loan can only be rolled over twice, and the lender must point you to free "
     "debt advice before either one. The trap the rule closed, explained."),
    ("how-to-complain-and-win", "How to Complain — and Win",
     "The 8-week clock, the free Ombudsman, the 6-month window, and what a winning "
     "unaffordable-lending complaint actually says."),
    ("if-you-cant-pay", "If You Can't Pay",
     "What a lender must do when you're struggling, what they may not do, Breathing "
     "Space, and the free help that changes outcomes."),
    ("loan-sharks-and-illegal-lending", "Loan Sharks and Illegal Lending",
     "An unauthorised lender generally cannot enforce the debt, and reporting them is "
     "safe, confidential and free. The numbers to call, by nation."),
    ("cheaper-ways-to-borrow-small-amounts", "Cheaper Ways to Borrow Small Amounts",
     "Credit unions, benefit advances, salary advances and arranged overdrafts — the "
     "small-sum routes that cost a fraction of a payday loan."),
    ("types-of-high-cost-credit-and-their-rules", "Types of High-Cost Credit and Their Rules",
     "Payday, doorstep, rent-to-own, logbook, guarantor, buy-now-pay-later: each has its "
     "own rulebook, and knowing which one covers you is half the battle."),
    ("jargon-buster", "High-Cost Credit Jargon, Translated",
     "APR, CPA, default notice, final response, forbearance, HCSTC. Every term a "
     "struggling borrower meets, in plain English."),
]

BODIES = {}

BODIES["the-payday-loan-price-cap"] = """
<p><strong>The short version:</strong> for payday-style loans, the law caps what you pay
at <strong>0.8% of the borrowed amount per day</strong>, caps default fees at
<strong>&pound;15 total</strong>, and — the backstop — says you can <strong>never be made
to pay more in interest and fees than the amount you borrowed</strong>. These are the
FCA's rules for high-cost short-term credit (rule book CONC, chapter 5A), in force since
2 January 2015. If your loan broke any of them, the excess is claimable.</p>

<h2>The three limits, in practice</h2>
<p><strong>The 0.8% daily cap.</strong> Borrow &pound;300 for 30 days and the most you can
be charged while repaying on time is &pound;300 &times; 0.8% &times; 30 = &pound;72. Every
fee the lender charges on the way in counts against this cap, not just "interest".</p>
<p><strong>The &pound;15 default cap.</strong> Miss the repayment date and the lender may
charge at most &pound;15 in default fees, once. Interest can continue, but only at the
same 0.8% daily rate, and only inside the third limit.</p>
<p><strong>The 100% total cost cap.</strong> Whatever happens — defaults, extensions,
collection — the total of all interest and all fees can never exceed the amount you
borrowed. Borrow &pound;300 and the absolute ceiling on charges is &pound;300; you can
never owe more than &pound;600 in total. There is no exception to this one.</p>

<h2>What a breach looks like</h2>
<p>Charges that keep growing after they equal the amount borrowed. More than &pound;15 in
"missed payment" or "arrears" fees. A "faster payout" or "same day" fee that pushes the
total over the daily cap. A rolled-over loan whose combined charges sail past 100%.
<a href="/tools/price-cap-checker.html">Put your numbers into the checker</a> — it does
the arithmetic in ten seconds.</p>

<h2>What to do about it</h2>
<p>The excess is your money. <a href="/guides/how-to-complain-and-win.html">Complain to
the lender first, then to the Financial Ombudsman if they don't put it right</a> — both
steps are free. If the lender has since gone bust, a claim may still be possible through
the administrators or, in some cases, the compensation scheme; the free debt charities can
tell you which applies.</p>

<p><em>Scope: the cap covers high-cost short-term credit — broadly, loans due within a
year at very high rates. Doorstep loans, rent-to-own and others have different rules:
<a href="/guides/types-of-high-cost-credit-and-their-rules.html">find your product
here</a>. And the authoritative text is the FCA's, not ours — check
fca.org.uk if a dispute turns on the wording.</em></p>
"""

BODIES["check-your-lender-is-authorised"] = """
<p><strong>The short version:</strong> anyone lending consumer credit in the UK must be
authorised by the Financial Conduct Authority. Checking takes two minutes on the
<a href="https://register.fca.org.uk" rel="noopener">FS Register</a> — the FCA's public
list — and it is the single highest-value check a borrower can make, because an
unauthorised lender generally <strong>cannot enforce the debt against you at all</strong>.</p>

<h2>How to check, exactly</h2>
<p>Go to <strong>register.fca.org.uk</strong>. Search the firm's name from your paperwork
— then check the details match: trading names, website address, and the permission to
carry on consumer-credit lending. Two things matter:</p>
<p><strong>The name matching is the point.</strong> Fraudsters "clone" real firms — using
a genuine firm's name and register number with their own phone number and bank details.
If the contact details you were given differ from the ones on the Register, treat it as a
clone. Use the Register's contact details, not theirs, to verify.</p>
<p><strong>"Authorised" for the right thing.</strong> A firm can be on the Register for
insurance but not lending. The entry lists permissions; consumer credit lending should be
among them.</p>

<h2>The fee-upfront scam, because it starts the same way</h2>
<p>If a "lender" asks for a fee, a transfer, vouchers or your card details <em>before</em>
paying out a loan — that is the classic loan-fee fraud, and no authorised lender operates
that way. Stop contact, keep the messages, and report it to Action Fraud (England, Wales
and NI) or Police Scotland.</p>

<h2>What happens if the lender wasn't authorised</h2>
<p>Lending without authorisation is a criminal offence, and the law's response is to make
the credit agreement generally <strong>unenforceable against you</strong> — the lender
needs a court's or regulator's permission to recover anything, which is rarely given. You
should still get advice before simply stopping payment (the free debt charities will help),
but the balance of power is not what the "lender" pretends it is.
<a href="/guides/loan-sharks-and-illegal-lending.html">If there are threats involved, this
is loan-shark territory — different page, stronger help</a>.</p>
"""

BODIES["affordability-checks-what-lenders-must-do"] = """
<p><strong>The short version:</strong> before lending, a firm must assess whether you can
repay <strong>without borrowing more, without missing priority bills, and without
hardship</strong> — the FCA's creditworthiness rules (CONC 5.2A). A loan that was only
repayable by re-borrowing was mis-sold, and complaints about it succeed, often years
after the event.</p>

<h2>What the lender was supposed to do</h2>
<p>The check must be proportionate to the loan — but for high-cost credit to someone
already stretched, "proportionate" is a real check: income, committed spending, existing
credit, and the pattern of your borrowing. The question the rules pose is not "will this
person probably pay us back?" — desperate people pay back at terrible cost — but
<strong>"can they pay us back without harm?"</strong> That distinction is the whole
rule.</p>

<h2>The tell-tale signs a check failed</h2>
<p>You took a new loan to repay the last one, more than once. You held several high-cost
loans at the same time. The repayments took most of what was left after rent and bills.
Your borrowing with the same lender climbed month after month. A lender who could see
that pattern — and they could — and kept lending was not checking affordability; they
were farming it.</p>

<h2>What an unaffordable-lending claim gets you</h2>
<p>The standard award when the Ombudsman upholds one: the lender refunds the
<strong>interest and charges</strong> on the unaffordable loans (you repay only what you
actually borrowed), adds interest on what you were out of pocket, and removes the loans
from your credit file. For someone who cycled through many loans, that can be
substantial.</p>

<h2>How to claim</h2>
<p>No solicitor and no claims company needed — the process is designed for you to do
free. <a href="/guides/how-to-complain-and-win.html">The complaints guide has the
step-by-step and what to write</a>; the phrase that frames it is "irresponsible and
unaffordable lending". List the loans, say why the repayments were unaffordable at the
time, and let the lender's own records do the rest. Then the 8-week clock runs and the
<a href="/tools/complaint-deadline-calculator.html">deadline calculator</a> keeps it
honest.</p>
"""

BODIES["stopping-payments-the-cpa-rules"] = """
<p><strong>The short version:</strong> most short-term lenders collect by "continuous
payment authority" (CPA) — permission to pull money from your card. It is not the blank
cheque it feels like. For payday-style loans a lender gets <strong>two failed collection
attempts</strong> and then must stop; they may <strong>not take a partial amount</strong>;
and you can cancel a CPA <strong>through your bank</strong>, which is legally required to
obey.</p>

<h2>The limits on the lender</h2>
<p>The FCA's rules for high-cost short-term credit say: if a CPA collection attempt fails
twice, the lender may not try again — they must contact you instead. And they may not use
the CPA to sweep whatever happens to be in the account: for these loans it is the full
instalment or nothing. A lender nibbling &pound;20 here and &pound;40 there off your card
is breaking the rule, and that is complaint material.</p>

<h2>Cancelling a CPA — the part everyone is told wrongly</h2>
<p>You do <strong>not</strong> need the lender's permission. Under the Payment Services
Regulations you can cancel a CPA by telling <strong>your bank or card issuer</strong>, and
they must stop future collections. Do it by phone, note the date, time and the name of the
person you spoke to, and follow up in the banking app's message system so there is a
written trace. If money is taken after you cancelled, <strong>the bank must refund it</strong>
— that is their obligation, not a favour.</p>
<p>Tell the lender too — not for permission, but so the missed collection is handled as an
arrangement rather than a surprise. <strong>Cancelling the collection method does not
cancel the debt</strong>: pair this with a repayment offer you can actually afford, and if
that is hard, <a href="/guides/if-you-cant-pay.html">the can't-pay guide is the next
stop</a>.</p>

<h2>Why you would do this</h2>
<p>Because a CPA takes its money the moment your wages land — before rent, before food,
before the electric. The rules on priority are yours to set, not the lender's. Rent,
council tax and energy come first; the law's own debt-collection rules recognise exactly
that, and so do the <a href="https://www.stepchange.org" rel="noopener">free debt
advisers</a> who will make the budget with you.</p>
"""

BODIES["rollovers-the-two-strike-rule"] = """
<p><strong>The short version:</strong> a payday-style loan can be rolled over (extended
for a fee) <strong>at most twice</strong>. Before any rollover the lender must give you an
information sheet pointing to free debt advice. The rule exists because rollovers were the
engine of the payday debt trap — and if a lender rolled you more than twice, or pushed
rollovers without the warnings, that is a complaint with teeth.</p>

<h2>Why the rule exists</h2>
<p>The old model was simple: lend &pound;200, collect a fee to extend it on payday,
repeat. The loan was never designed to be repaid — the fees were the product. Two
rollovers is the regulator's answer: enough for a genuinely short-term hiccup, too few to
build a business on trapping you.</p>

<h2>What counts as a rollover</h2>
<p>Any arrangement where the loan you could not repay is extended, refinanced or replaced
with another loan from the same lender covering the same money — whatever it is called on
the paperwork. "Refinance", "extension", "new agreement": if the practical effect is that
last month's unpaid loan became this month's loan plus charges, it counts. The 100% total
cost cap keeps running across the whole chain — charges on the rolled loan still can
never exceed the amount originally borrowed.
<a href="/guides/the-payday-loan-price-cap.html">The price cap guide covers that
maths</a>.</p>

<h2>The pattern to check in your own history</h2>
<p>Look at your statements with one question: how many consecutive months did the same
lender's name appear? Three or more extensions or same-day re-borrowings is a pattern the
lender's own records made obvious. That is both a rollover-rule issue and an
<a href="/guides/affordability-checks-what-lenders-must-do.html">affordability failure</a>
— and the affordability route is usually the stronger claim, because it covers the whole
chain of loans, not just the extension fees.</p>
"""

BODIES["how-to-complain-and-win"] = """
<p><strong>The short version:</strong> complain to the lender in writing; they get
<strong>8 weeks</strong> to give a final response; then the
<a href="https://www.financial-ombudsman.org.uk" rel="noopener">Financial Ombudsman
Service</a> decides it for free — and you have <strong>6 months from the final
response</strong> to go to them. No claims company, no solicitor, no fee. The
<a href="/tools/complaint-deadline-calculator.html">deadline calculator</a> turns your
dates into the two that matter.</p>

<h2>Step one: the complaint itself</h2>
<p>Email or letter, to the lender, with the word "complaint" in it. Three parts:</p>
<p><strong>What happened</strong> — the loans, roughly when, and the numbers you have.
You do not need perfect records; the lender must have them and the Ombudsman can make
them produce everything.</p>
<p><strong>What went wrong</strong> — pick the rule:
"<em>you charged more than the price cap allows</em>"
(<a href="/tools/price-cap-checker.html">check first</a>);
"<em>the lending was irresponsible and unaffordable — proper checks would have shown I
could not repay without hardship</em>"
(<a href="/guides/affordability-checks-what-lenders-must-do.html">the strongest and most
common claim</a>); "<em>you continued CPA attempts / took partial payments against the
rules</em>"; "<em>you rolled the loan over more than twice</em>".</p>
<p><strong>What you want</strong> — refund of interest and charges, interest on top for
the time you were without the money, and the loans removed from your credit file.</p>

<h2>Step two: the 8 weeks</h2>
<p>The lender must send a <strong>final response</strong> within 8 weeks. Expect a
settlement offer somewhere in between — compare it against the full remedy above before
accepting, because accepting usually closes the complaint. If 8 weeks pass with no final
response, you go to the Ombudsman without waiting for one.</p>

<h2>Step three: the Ombudsman</h2>
<p>Free, independent, designed to be used without help, and used successfully against
short-term lenders tens of thousands of times. You fill in their form, attach what you
have, and they take it from there. The one hard rule is the <strong>6-month window</strong>
from the final response — miss it and they usually cannot help. Put the date in your
calendar the day the final response arrives.</p>

<p><em>If the lender has gone out of business, complain instead to the administrators —
the free debt charities can tell you whether a compensation-scheme claim also applies.
And if a claims company offers to do all this "for a percentage": everything above is
free, and the Ombudsman treats your own words just as seriously.</em></p>
"""

BODIES["if-you-cant-pay"] = """
<p><strong>The short version:</strong> a lender must treat customers in difficulty with
<strong>forbearance and due consideration</strong> — that is a rule (CONC 7), not a
kindness. In practice: they must consider freezing interest, accepting reduced payments,
and giving you breathing room; they may not harass you, mislead you, or pretend powers
they don't have. And before anything else: <strong>rent, council tax and energy come
before any loan</strong>, whatever the lender's tone suggests.</p>

<h2>What the rules make them do</h2>
<p>When you tell a lender you are struggling, the rulebook expects a real response:
consider suspending or reducing interest and charges, accept a repayment plan that leaves
you enough to live on, point you to free debt advice, and give you time to get it. A
lender who answers "our system can't do that" is describing their software, not the
rules.</p>

<h2>What they may not do</h2>
<p>Contact you at unreasonable hours or relentlessly; discuss your debt with your
employer, family or neighbours; claim court action is underway when it is not; send
letters dressed up to look like court documents; refuse to deal with a debt adviser
acting for you; or ignore a notified mental-health crisis. Harassment of debtors is not
just against FCA rules — it is a criminal offence under the Administration of Justice
Act 1970. Note the incidents; they strengthen
<a href="/guides/how-to-complain-and-win.html">the complaint</a>.</p>

<h2>Breathing Space: the legal pause</h2>
<p>In England and Wales, the <strong>Breathing Space</strong> scheme gives you, via any
free debt adviser, <strong>60 days</strong> during which enforcement is paused and most
interest and charges on the debts in the scheme are frozen while you get a plan together.
There is a separate version, without the 60-day limit, for people receiving mental-health
crisis treatment. Scotland has its own moratorium scheme with different periods; Northern
Ireland differs too — the advisers below know the local position.</p>

<h2>Where the actual help is</h2>
<p>Free, confidential, and they negotiate with lenders daily:
<a href="https://www.stepchange.org" rel="noopener">StepChange</a>,
<a href="https://www.nationaldebtline.org" rel="noopener">National Debtline</a>,
<a href="https://www.citizensadvice.org.uk" rel="noopener">Citizens Advice</a>, with
<a href="https://www.moneyhelper.org.uk" rel="noopener">MoneyHelper</a> as the
government-backed directory of all of it. Debt advice from these sources costs nothing —
anyone charging you for it is taking money a charity would not.</p>

<p><em>One more thing: if any part of the trouble is a payment authority draining your
account on payday, <a href="/guides/stopping-payments-the-cpa-rules.html">you can stop
that today</a>.</em></p>
"""

BODIES["loan-sharks-and-illegal-lending"] = """
<p><strong>The short version:</strong> lending money without FCA authorisation is a
crime. A loan from an unlicensed lender is generally <strong>unenforceable</strong> —
in the eyes of the law you usually do not owe the interest, and often not the principal
either — and the specialist teams that deal with loan sharks protect the borrower, not
punish them. Reporting is confidential and free, and borrowing from a shark is
<strong>not</strong> an offence: theirs is the crime, not yours.</p>

<h2>Who to call, by nation</h2>
<p><strong>England:</strong> Stop Loan Sharks (the Illegal Money Lending Team) —
<strong>0300 555 2222</strong>, 24 hours, or stoploansharks.co.uk.<br>
<strong>Wales:</strong> the Wales Illegal Money Lending Unit — <strong>0300 123 3311</strong>.<br>
<strong>Scotland:</strong> the Scottish Illegal Money Lending Unit via Trading Standards —
<strong>0800 074 0878</strong>.<br>
<strong>Northern Ireland:</strong> the Consumer Council or the police on 101 — or 999 if
you are being threatened right now, anywhere in the UK.</p>

<h2>What a loan shark looks like now</h2>
<p>Less baseball bat, more WhatsApp. A "friend of a friend" who lent cash with no
paperwork; a lender in a Facebook group; someone who took your bank card or benefits card
as "security"; a debt that grows on their say-so with no statement you can check. No
paperwork, no APR, no register entry — no authorisation.
<a href="/guides/check-your-lender-is-authorised.html">Two minutes on the FS Register
settles it</a>.</p>

<h2>What the teams actually do</h2>
<p>They prosecute the lender, not you. They can have illegal debts written off, help you
recover money and documents, and support you if there have been threats — and they have
done it for tens of thousands of borrowers. Keeping paying quietly does not make the
problem smaller; it funds it.</p>

<p><em>Different problem, same first move: if the "lender" wanted a fee before paying
out a loan that never came, that is loan-fee fraud — report it to Action Fraud (or
Police Scotland) and tell your bank immediately.</em></p>
"""

BODIES["cheaper-ways-to-borrow-small-amounts"] = """
<p><strong>The short version:</strong> the routes below cost a fraction of high-cost
credit, and most people who end up with a payday loan qualified for at least one of them.
This page exists because the cheapest loan is the one you compare against — and high-cost
lenders depend on you not knowing the comparison exists.</p>

<h2>Credit unions</h2>
<p>Not-for-profit lenders owned by their members, and the law caps their loan interest at
<strong>3% a month on the reducing balance</strong> (about 42.6% APR) in Great Britain —
against the 0.8% <em>per day</em> a payday lender may charge. They lend small sums, they
consider people banks refuse, and many run save-as-you-borrow accounts that leave you
with savings at the end instead of a renewal offer. Find yours through
<a href="https://www.findyourcreditunion.co.uk" rel="noopener">findyourcreditunion.co.uk</a>.
The <a href="/tools/true-cost-calculator.html">true cost calculator</a> shows the
difference on your own numbers.</p>

<h2>If you receive benefits</h2>
<p>A <strong>Budgeting Advance</strong> (Universal Credit) or Budgeting Loan (legacy
benefits) from the DWP is repaid from future payments at <strong>zero interest</strong> —
for essentials like a cooker, rent in advance or work expenses. Zero interest beats every
lender on this page; check eligibility at gov.uk before borrowing commercially. Many
councils also run local welfare assistance schemes for emergencies — grants, not loans, in
some cases.</p>

<h2>An arranged overdraft</h2>
<p>Since the FCA's 2020 overdraft reforms, banks must charge a single interest rate — no
daily "usage fees" — and an <em>arranged</em> overdraft for a short gap is often far
cheaper than a short-term loan. The trap is letting it become permanent: an overdraft is
the most expensive place to live and one of the cheaper places to visit.</p>

<h2>Salary advances and employer schemes</h2>
<p>Some employers offer wage advances or hardship loans repaid through payroll. Terms
vary and this corner is lightly regulated — read what happens if you leave the job — but
a genuine employer scheme at low or no cost beats commercial high-cost credit.</p>

<h2>The ones that only look cheap</h2>
<p><strong>Buy-now-pay-later</strong> is interest-free until it isn't — late fees and
collections turn "free" expensive, and the sector is only now coming fully inside the
FCA's rules. <strong>Unarranged</strong> overdrafts, card cash withdrawals and
"just this once" catalogue credit all price worse than they present.
For choosing between mainstream products on cost, that is a different site's job — this
one's is making sure you know the protective floor under all of them.</p>
"""

BODIES["types-of-high-cost-credit-and-their-rules"] = """
<p><strong>The short version:</strong> "high-cost credit" is not one product, and the
protections differ by type. Find yours below; each entry says which rules apply and links
the deeper guide.</p>

<h2>Payday / high-cost short-term loans (HCSTC)</h2>
<p>Loans up to about a year at very high rates. The full protective set applies:
<a href="/guides/the-payday-loan-price-cap.html">the price cap</a> (0.8%/day, &pound;15
default cap, 100% total),
<a href="/guides/rollovers-the-two-strike-rule.html">the two-rollover limit</a>, the
<a href="/guides/stopping-payments-the-cpa-rules.html">CPA restrictions</a>, and adverts
must carry a risk warning pointing to MoneyHelper.</p>

<h2>Doorstep loans (home-collected credit)</h2>
<p>An agent calls weekly for cash repayments. No price cap — but strict rules instead:
the lender needs your written request to visit and discuss a new loan, refinancing an
existing doorstep loan needs clear cost comparisons, and the affordability duties apply
in full. If the same agent has refinanced you repeatedly for years,
<a href="/guides/affordability-checks-what-lenders-must-do.html">that is exactly the
pattern the affordability rules exist for</a>.</p>

<h2>Rent-to-own (hire purchase on household goods)</h2>
<p>Weekly-payment stores for appliances and furniture have their own FCA price cap
(since 2019): the <strong>credit cost may not exceed 100% of the product's price</strong>,
and the store must benchmark its product prices against ordinary retailers. The goods are
not yours until the final payment — but once you have paid a third, the lender needs a
court order to repossess them.</p>

<h2>Logbook loans (bills of sale on your car)</h2>
<p>You borrow against the car and the lender technically owns it until repayment. Old,
harsh law — but FCA affordability and arrears rules fully apply to the lender, and if
you bought a car that unknowingly carried someone else's logbook loan, you have defences.
Get advice before surrendering a car; the free debt charities know this ground.</p>

<h2>Guarantor loans</h2>
<p>A friend or relative promises to pay if you don't. The affordability rules protect
<strong>both of you</strong> — the lender had to check the guarantor could afford the
worst case too, and guarantors have won complaints in exactly those terms. If a guarantor
is being pursued, they can complain in their own right:
<a href="/guides/how-to-complain-and-win.html">same route, same free Ombudsman</a>.</p>

<h2>Buy-now-pay-later</h2>
<p>Historically outside regulation entirely; now being brought inside the FCA's
perimeter, with affordability checks and Ombudsman access following. The rules here are
the newest and most in motion of anything on this page — check the current position at
fca.org.uk before relying on it.</p>

<h2>Catalogue credit and store cards</h2>
<p>Regulated credit with the ordinary protections — the risk is the pattern, not the
paperwork: minimum payments that never touch the balance. The persistent-debt rules make
card firms intervene when payments mostly service interest; if that letter arrives, take
it seriously — it is one of the few letters from a lender written in your interest.</p>
"""

BODIES["jargon-buster"] = """
<p><strong>How to use this page:</strong> every term links back to the guide where it
does real work. If a lender's letter contains a word that is not here and not plain
English, that is their failure, not yours — the rules require communications to be clear.</p>

<h2>The products</h2>
<p><strong>HCSTC</strong> — high-cost short-term credit: the regulator's name for
payday-style loans; the trigger for
<a href="/guides/the-payday-loan-price-cap.html">the price cap</a>.<br>
<strong>Home-collected credit</strong> — doorstep loans with a calling agent.<br>
<strong>Rent-to-own</strong> — hire purchase on household goods, with its own price cap.<br>
<strong>Logbook loan</strong> — borrowing secured on your car via a bill of sale.<br>
<strong>BNPL</strong> — buy-now-pay-later instalments at the checkout.</p>

<h2>The charges</h2>
<p><strong>APR</strong> — the yearly cost of credit including fees, for comparing like
with like. A small daily rate compounds to a four-figure APR, which is precisely why
adverts must show it.<br>
<strong>Daily rate</strong> — what short-term lenders quote because it sounds small;
capped at 0.8% for HCSTC.<br>
<strong>Default fee</strong> — a charge for missing a payment; capped at &pound;15 total
for HCSTC.<br>
<strong>Total cost cap</strong> — the 100% ceiling: charges can never exceed the amount
borrowed.</p>

<h2>The collection machinery</h2>
<p><strong>CPA (continuous payment authority)</strong> — permission to pull money from
your card; <a href="/guides/stopping-payments-the-cpa-rules.html">limited to two failed
attempts, no partial grabs, cancellable through your bank</a>.<br>
<strong>Rollover</strong> — extending a loan for a fee;
<a href="/guides/rollovers-the-two-strike-rule.html">two is the legal maximum</a>.<br>
<strong>Default notice</strong> — the formal Consumer Credit Act warning a lender must
serve before most enforcement; it has a statutory format and gives you at least 14 days.<br>
<strong>Forbearance</strong> — the duty to give struggling borrowers real accommodation:
<a href="/guides/if-you-cant-pay.html">frozen interest, reduced payments, time</a>.</p>

<h2>The complaint machinery</h2>
<p><strong>Final response</strong> — the lender's formal last word on your complaint;
starts your <strong>6-month</strong> window to go to the Ombudsman.<br>
<strong>FOS</strong> — the Financial Ombudsman Service: free, independent,
<a href="/guides/how-to-complain-and-win.html">and the reason you never need a claims
company</a>.<br>
<strong>The 8 weeks</strong> — the lender's deadline to produce a final response.
<a href="/tools/complaint-deadline-calculator.html">Both clocks, calculated</a>.<br>
<strong>FS Register</strong> — the FCA's public list of authorised firms;
<a href="/guides/check-your-lender-is-authorised.html">two minutes that can change
everything</a>.<br>
<strong>Breathing Space</strong> — the England-and-Wales scheme freezing enforcement and
most charges for 60 days while you get debt advice.</p>

<h2>The people</h2>
<p><strong>FCA</strong> — the Financial Conduct Authority, the regulator whose rulebook
this site exists to put in borrowers' hands. (We are not them, and they did not write
this.)<br>
<strong>IMLT / Stop Loan Sharks</strong> — the teams that prosecute
<a href="/guides/loan-sharks-and-illegal-lending.html">illegal lenders</a> and protect
their borrowers.<br>
<strong>Debt adviser</strong> — free at StepChange, National Debtline and Citizens
Advice. Anyone charging for what they do free is part of the problem this site is
about.</p>
"""
