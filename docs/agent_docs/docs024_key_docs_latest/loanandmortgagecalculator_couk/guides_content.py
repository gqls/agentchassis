#!/usr/bin/env python3
"""guides_content.py — the guide prose for loanandmortgagecalculator.co.uk.

WRITTEN NEW, not adapted. The brief was that these must be "different and better
than the mortgagecalculator.co.uk guides", and that the two sites should evolve
toward different target markets. So none of these is a rewrite of a guide on
either source site: every one is about the point where unsecured borrowing and a
mortgage MEET, which is the one subject neither single-topic site can cover.

EDITORIAL RULE, applied throughout: explain the MECHANISM, never quote a current
rate or tax band. The source sites hard-code "3.75% base rate" and a March 2026
date; copy like that is wrong within weeks and there is nothing to tell the
reader it has gone stale. Mechanisms do not go stale. Where a figure is
structural (the 4.5x loan-to-income flow limit, the 58-day settlement rule,
Section 75's £100 floor) it is stated as structural and attributed. Where a
number moves, the reader is sent to the calculator or to gov.uk.
"""

BODIES = {}

# ─────────────────────────────────────────────────────────────────────────────
BODIES["how-loans-affect-mortgage-affordability"] = """
<p>This is the single most useful thing to understand if you are borrowing money
and also want a mortgage, and it is almost never explained properly: <strong>a
mortgage lender does not look at what you owe. It looks at what you pay each
month.</strong></p>

<p>That distinction decides how much house you can buy.</p>

<h2>The arithmetic lenders actually use</h2>

<p>A mortgage lender starts from your income, applies an income multiple, and then
subtracts your committed monthly outgoings from the money available to service a
mortgage. Committed outgoings include personal loan payments, car finance, credit
card minimums, and any other regular credit commitment.</p>

<p>The rough conversion, and it is rough on purpose because every lender's model
differs: <strong>each £100 a month of existing credit payments reduces your
maximum mortgage by somewhere around £5,000 to £7,000.</strong></p>

<p>So a £320 a month car finance agreement is not a £320 problem. It is a
£16,000&ndash;£22,000 hole in your borrowing power, for as long as the agreement
runs.</p>

<div class="highlight-box">
<p class="mt-0"><strong>Try it both ways.</strong> Put your income into the
<a href="/mortgages/affordability.html">affordability calculator</a> with your
current commitments, then again with them at zero. The gap between the two answers
is what your existing borrowing is costing you in house.</p>
</div>

<h2>Why a £2,000 balance can hurt more than a £15,000 one</h2>

<p>Here is the counter-intuitive part. Because lenders count the monthly payment
rather than the balance, <em>how fast you are repaying something matters more than
how much of it is left.</em></p>

<ul>
<li>A £15,000 loan with three years left, at £440 a month, might cost you £25,000
of borrowing power.</li>
<li>A £2,000 loan you are aggressively clearing at £400 a month costs you almost
exactly the same.</li>
</ul>

<p>The second debt is a seventh of the size and does nearly identical damage to
your application. And it is far cheaper to remove: two thousand pounds clears it.</p>

<p>This is why "pay off your smallest expensive debts before you apply" is better
advice than "pay off your biggest debt", if the goal is the mortgage. It is not
the same as the advice that minimises your interest bill &mdash; the two goals
genuinely pull in different directions, and it is worth being clear with yourself
about which one you are optimising for.</p>

<h2>The income multiple, and the cap you cannot argue with</h2>

<p>Most UK lenders work to roughly four to four-and-a-half times income, and there
is a structural reason the number clusters there. Under rules set by the Financial
Policy Committee, lenders are limited in the <em>proportion</em> of their new
mortgage lending that can go above <strong>4.5 times loan-to-income</strong>. It is
a limit on the lender's book, not a hard ban on your individual application &mdash;
which is why above-4.5x lending exists but is rationed, and tends to go to
applicants who are strong in every other respect.</p>

<p>Practical consequence: if you are near the multiple, your commitments are the
lever you can actually move. You cannot argue your way past a cap. You can clear a
loan.</p>

<h2>Stress testing, and why the rate you are offered is not the rate you are
assessed at</h2>

<p>Lenders do not check that you can afford the payment at the rate you are being
offered. They check that you could still afford it if rates rose. The specific
mechanism has changed over the years &mdash; the Bank of England withdrew its
prescriptive affordability stress test in 2022 &mdash; but the underlying FCA
requirement to assess affordability against future rate rises did not go away, and
lenders still apply their own stress rate.</p>

<p>What this means for you is simple and slightly brutal: <strong>your other debt
is subtracted from your income, and then what is left is tested against a rate
higher than the one you will pay.</strong> Both effects compound. Use the
<a href="/mortgages/rate-forecaster.html">payment change forecaster</a> to see what
a higher rate does to the payment, and the
<a href="/loans/interest-rate-stress-test.html">rate stress test</a> to see what it
does to your variable-rate borrowing at the same time.</p>

<h2>What to do, in order</h2>

<ol>
<li><strong>List every monthly credit payment.</strong> Not balances &mdash;
payments. Loans, cards, car finance, buy-now-pay-later, overdraft. This total is
the number that costs you borrowing power.</li>
<li><strong>Find the worst ratio.</strong> For each debt, divide the monthly
payment by the balance. The highest number is the debt doing the most damage per
pound it would take to clear.</li>
<li><strong>Clear what you can, and close it properly.</strong> An account settled
but left open can still be counted. Get it closed and confirmed.</li>
<li><strong>Do not open anything new.</strong> Not in the six months before you
apply, and especially not a car.</li>
<li><strong>Leave the credit card, but empty it.</strong> Long-standing accounts
help your credit file; balances hurt your affordability. Zero balance, account
open, is the best of both.</li>
<li><strong>Then re-run the affordability calculator.</strong> The improvement is
usually larger than people expect, and seeing it is what makes the discipline
stick.</li>
</ol>

<h2>The one trap worth naming</h2>

<p>Consolidating your debts into one loan with a longer term lowers your monthly
payment, which improves your mortgage affordability on paper. It also increases what
you repay overall, and it puts a fresh credit agreement on your file immediately
before you apply. Whether that trade is worth it is genuinely case-by-case &mdash;
<a href="/guides/consolidating-debt-into-your-mortgage.html">we have written it up
properly here</a> &mdash; but do not treat it as a free improvement. It is not.</p>
"""

# ─────────────────────────────────────────────────────────────────────────────
BODIES["consolidating-debt-into-your-mortgage"] = """
<p>Consolidating unsecured debt into your mortgage is the most consequential
decision in this whole subject, and the one most often made for the wrong reason.
The monthly saving is immediate, visible and pleasant. The two costs are delayed
and invisible, and one of them is not financial at all.</p>

<h2>What actually happens</h2>

<p>You add, say, £25,000 of loans and card balances to your mortgage when you
remortgage. Your unsecured payments &mdash; perhaps £600 a month &mdash; disappear.
Your mortgage payment rises by perhaps £130. You are £470 a month better off, and
the relief is real.</p>

<p>Then two things follow.</p>

<h3>Cost one: you have stretched expensive debt over a very long time</h3>

<p>A personal loan is priced high and repaid fast &mdash; typically three to five
years. A mortgage is priced low and repaid slowly &mdash; typically twenty to
thirty. Moving a debt from the first to the second lowers the rate and multiplies
the number of years you pay it.</p>

<p><strong>A lower rate over a much longer term routinely costs more in total than a
higher rate over a short one.</strong> That is the whole trap, and it is arithmetic,
not opinion. Run your own figures: the
<a href="/loans/consolidation.html">consolidation calculator</a> gives you the
total, and the <a href="/mortgages/repayment.html">repayment calculator</a> shows
what adding to the mortgage does across the full term. Compare the two totals, not
the two monthly payments.</p>

<h3>Cost two: you have changed what happens if it goes wrong</h3>

<p>This is the part that gets left out, and it matters more than the money.</p>

<p>An unsecured loan is not attached to your house. If you genuinely cannot pay it,
the consequences are serious &mdash; default, a damaged credit file, possibly a
county court judgment &mdash; but you do not lose your home over it.</p>

<p><strong>Once that debt is part of your mortgage, non-payment is a step on the
road to possession.</strong> You have converted a debt that could cost you your
credit rating into a debt that could cost you where you live. If your income is
stable and your job is secure, that risk may be acceptable. If either is uncertain,
you have just moved your safety margin in the wrong direction, at exactly the moment
you were feeling relieved.</p>

<div class="highlight-box">
<p class="mt-0"><strong>The honest test.</strong> Ask what happens in the version of
next year where your household income drops by a third. Under the current
arrangement, which creditor do you fall behind with, and what do they do? Under the
consolidated arrangement, the answer to "which creditor" is always your mortgage
lender.</p>
</div>

<h2>When it is genuinely the right call</h2>

<p>It is not always wrong. It is a reasonable decision when most of the following
hold:</p>

<ul>
<li><strong>You keep the term short.</strong> If your lender will let you add the
debt over five or ten years rather than the remaining mortgage term, most of the
lifetime-cost objection disappears. Some will; ask explicitly, because the default
is the full term.</li>
<li><strong>You have enough equity that the loan-to-value band does not worsen.</strong>
Crossing from one LTV band to a higher one can raise the rate on the <em>entire</em>
mortgage, which can wipe out the saving several times over. Check the band, not just
the amount.</li>
<li><strong>The unsecured debt is genuinely expensive</strong> &mdash; store cards,
high-rate cards, subprime loans &mdash; rather than a cheap loan you took at a good
rate.</li>
<li><strong>The overspending has actually stopped.</strong> Consolidating and then
refilling the cards is the classic failure mode, and it leaves you with the old debt
plus a larger mortgage. If you are not confident about this, that is the thing to
fix first.</li>
<li><strong>Your income is stable and you are not near the edge.</strong> See the
honest test above.</li>
</ul>

<h2>The costs people forget to count</h2>

<ul>
<li><strong>Early repayment charges</strong> on your existing mortgage if you are
inside a fixed period. Often one to five per cent of the balance &mdash; on
£200,000 that is thousands, and it can make the whole exercise pointless until the
fix ends.</li>
<li><strong>Product, valuation and legal fees</strong> on the new mortgage.</li>
<li><strong>Early settlement interest</strong> on the loans you are clearing. You do
have a statutory right to settle early, but the lender may add up to 58 days'
interest &mdash; see <a href="/loans/settlement-calculator.html">the settlement
calculator</a>. It is rarely a deal-breaker, but it is real money and it is not in
the balance shown on your statement.</li>
<li><strong>A higher rate on the whole mortgage</strong> if the extra borrowing
pushes you into a worse LTV band. This is the big one, and the easiest to miss.</li>
</ul>

<h2>The alternatives worth pricing first</h2>

<p>Consolidating into the mortgage is one option among several, and it is the least
reversible. Before committing, price these:</p>

<ol>
<li><strong>A 0% balance transfer</strong> for card debt, if your credit file
supports it. Pay a small transfer fee, pay no interest, keep the debt unsecured.</li>
<li><strong>An unsecured consolidation loan</strong> over a <em>short</em> term. One
payment, no security over your home, and the term forces the debt to actually
end.</li>
<li><strong>Overpaying the most expensive debt</strong> and leaving the rest alone.
Unglamorous, usually the cheapest, and it removes the highest-rate debt first &mdash;
see <a href="/loans/overpayment-calculator.html">what overpaying saves</a>.</li>
<li><strong>Doing nothing until the fix ends</strong>, then reassessing without an
early repayment charge in the way.</li>
</ol>

<p>If you are consolidating because the payments have become genuinely unmanageable
rather than merely annoying, stop and read
<a href="/guides/when-repayments-are-a-struggle.html">this instead</a>. Free
services will negotiate with your creditors on your behalf, and they will do it
without putting your house into the arrangement.</p>
"""

# ─────────────────────────────────────────────────────────────────────────────
BODIES["total-cost-of-borrowing"] = """
<p>Every credit advert in Britain leads with the monthly payment. There is a reason
for that, and it is not your convenience.</p>

<p>The monthly payment is the one number that can be made to look better without
anything actually improving. Stretch the term and it falls. Nothing about the debt
has got cheaper &mdash; you have simply agreed to pay for longer, and therefore to
pay more.</p>

<h2>The demonstration</h2>

<p>Take any loan and put it into the <a href="/loans/standard-calc.html">repayment
calculator</a> twice: once over three years, once over seven. Then compare the two
figures the calculator gives you.</p>

<ul>
<li>The <strong>monthly payment</strong> falls substantially on the longer term.
This is what the advert shows you.</li>
<li>The <strong>total cost of credit</strong> rises substantially. This is what you
actually pay.</li>
</ul>

<p>Same amount borrowed. Same interest rate. Two very different prices. The only
thing that changed was time, and time is what interest is charged for.</p>

<h2>The number to compare, and the one to ignore</h2>

<p><strong>Compare total repayable.</strong> Amount borrowed, plus every pound of
interest and fee, across the whole agreement. It is the only figure that cannot be
flattered by restructuring.</p>

<p>Use APR to compare two loans <em>of the same term and amount</em>, which is what
it was designed for. It is a rate, so it cannot see term length &mdash; two loans
with identical APRs can cost wildly different amounts if one runs twice as long.
APR is a good tool being used, most of the time, for the wrong job.</p>

<p>For mortgages the equivalent figure is <strong>APRC</strong>, which folds in fees
and assumes you stay for the full term. Since almost nobody stays on the same
mortgage product for twenty-five years, APRC is close to useless for choosing
between fixed deals. What you want there is the total cost over the fixed
period &mdash; that is exactly what the
<a href="/mortgages/fee-analyser.html">fee analyser</a> computes, and it is why a
market-leading rate with a large product fee frequently loses to a slightly worse
rate with no fee.</p>

<h2>Representative APR: what "representative" means</h2>

<p>An advertised "representative APR" must be the rate offered to at least
<strong>51%</strong> of people accepted. Read that carefully: up to 49% of accepted
applicants can be paying more, and the advertised figure is still lawful.</p>

<p>Which means the rate in the advert is not an offer, and your quote may be worse
for reasons you never see. Get the actual figure before you compare anything.</p>

<h2>Where the cost hides that no rate can show you</h2>

<ul>
<li><strong>Product and arrangement fees.</strong> A £999 mortgage product fee on a
two-year fix is roughly £42 a month of real cost that the rate does not
mention.</li>
<li><strong>Interest charged on fees.</strong> Add a fee to the loan instead of
paying it, and you pay interest on it for the full term.</li>
<li><strong>Balloon payments.</strong> PCP car finance keeps the monthly figure low
by deferring a large lump to the end. The monthly payment is not the price of the
car &mdash; see <a href="/loans/car-finance-calculator.html">PCP against HP</a>.</li>
<li><strong>Early repayment charges.</strong> The cost of changing your mind, which
you only discover you needed after you needed it.</li>
<li><strong>Settlement interest.</strong> Clearing a loan early can attract up to 58
days' additional interest under the early settlement rules.</li>
<li><strong>Insurance sold alongside.</strong> Priced monthly, presented as small,
occasionally optional in ways that are not made obvious.</li>
</ul>

<h2>Paying monthly for things that are not loans</h2>

<p>Two everyday cases where you are borrowing without being told you are:</p>

<p><strong>Car insurance paid monthly.</strong> This is credit. The insurer runs a
check and charges interest for spreading an annual premium. Paying annually is
routinely ten to twenty per cent cheaper for exactly this reason, and the saving is
one of the largest available for a single decision.</p>

<p><strong>Buy now, pay later.</strong> Whether interest-bearing or not, these are
credit agreements. Several are reported to credit reference agencies, and a mortgage
lender that sees a cluster of them reads it as a signal about how you manage
money &mdash; regardless of whether you paid on time.</p>

<h2>The habit worth building</h2>

<p>Before agreeing to any credit, get one figure: <strong>the total you will have
handed over by the end.</strong> Not the monthly payment. Not the rate.</p>

<p>Then decide whether the thing is worth that. Very often it is. Sometimes the
number is startling enough to change your mind, and those are the occasions that pay
for the habit many times over.</p>
"""

# ─────────────────────────────────────────────────────────────────────────────
BODIES["deposit-or-clear-the-debt"] = """
<p>You have come into some money &mdash; a bonus, an inheritance, savings that
finally added up to something. You have a loan, and you want to buy a house. Which
does the money go to?</p>

<p>The internet's answer is "compare the interest rates". That answer is incomplete
enough to be misleading, because for this particular decision the rates are usually
the <em>least</em> important of the three factors that matter.</p>

<h2>Factor one: the loan-to-value cliff</h2>

<p>Mortgage pricing is banded by loan-to-value, and the bands are steps rather than
a slope. Crossing one changes the rate on the <strong>whole</strong> mortgage, not
on the marginal pound.</p>

<p>So if £4,000 more deposit takes you from just above a band boundary to just
below it, that £4,000 is not earning you the mortgage interest rate. It is earning
you a rate reduction across your entire balance, for the length of the deal. On a
large mortgage that can be worth several times the sum involved, and it dwarfs
anything the loan is costing you.</p>

<p><strong>Conversely, if you are nowhere near a boundary, extra deposit is
comfortably the weakest use of the money.</strong> The pound that moves you across a
threshold and the pound that does not are worth wildly different amounts, and it is
the same pound. Find out where the boundaries sit for the mortgages you are actually
looking at before you decide anything.</p>

<h2>Factor two: affordability, which is about monthly payments</h2>

<p>As covered in <a href="/guides/how-loans-affect-mortgage-affordability.html">the
borrowing power guide</a>, lenders subtract your monthly credit commitments from the
income available to service a mortgage. Clearing a loan removes its monthly payment
and typically returns £5,000&ndash;£7,000 of borrowing power per £100 a month
eliminated.</p>

<p>That gives clearing the debt a second, larger effect: it does not just save you
interest, it <em>raises the ceiling on what you can buy.</em> If the ceiling is what
is stopping you, this is decisive and no amount of extra deposit substitutes for
it.</p>

<div class="highlight-box">
<p class="mt-0"><strong>Which constraint are you actually up against?</strong> If
you cannot borrow enough, clear the debt. If you can borrow enough but the rate is
poor because your LTV is marginal, add to the deposit. Establish which of the two
problems you have before optimising, because the answers point in opposite
directions.</p>
</div>

<h2>Factor three: what you can undo</h2>

<p>Money in a savings account can be used for anything. Money paid into a mortgage
is extremely difficult to get back out &mdash; you would need to remortgage or
borrow again, at whatever rates and in whatever circumstances happen to obtain at
the time. Money used to clear a loan is gone too, but it takes an obligation with
it.</p>

<p>Which means the order should almost always be:</p>

<ol>
<li><strong>An emergency fund first</strong>, three to six months of essential
outgoings, and do not raid it for either purpose. Every debt problem that starts as
a crisis starts with not having this.</li>
<li><strong>Then expensive unsecured debt.</strong> Cards and high-rate loans, worst
rate first &mdash; and note that clearing these helps your affordability too, so it
is doing two jobs at once.</li>
<li><strong>Then the deposit</strong>, aiming specifically at the nearest LTV
boundary rather than at "as much as possible".</li>
<li><strong>Then, once you have the mortgage,</strong> overpayments &mdash; which are
worth far less per pound than any of the above, because the rate is the lowest.
See <a href="/mortgages/overpayment.html">what overpaying achieves</a>.</li>
</ol>

<h2>The timing detail that catches people out</h2>

<p>If you clear a debt in order to improve a mortgage application, do it <strong>at
least a couple of months before applying</strong>, and check your credit file
afterwards to confirm the account shows as settled and closed.</p>

<p>Credit files update on the lenders' own reporting cycles, not on the day you pay.
An account you cleared last week may still show a balance and a monthly payment when
the underwriter looks, and the underwriter will assess what is in front of them. The
work gets done and none of the benefit arrives.</p>

<h2>A worked way to think about it</h2>

<p>Run the numbers rather than reasoning about them:</p>

<ol>
<li><a href="/mortgages/affordability.html">Affordability calculator</a>, with the
debt. Note the maximum.</li>
<li>Again, with the debt cleared. Note the new maximum. The difference is what
clearing it buys you in purchasing power.</li>
<li><a href="/loans/loan-vs-savings.html">Loan versus savings</a> for the pure
interest comparison &mdash; the answer the internet gives, which is now one input
among three rather than the whole decision.</li>
<li>Find the LTV boundary nearest your position and work out what deposit reaches
it. If it is within range of your lump sum, price the rate reduction across the
whole mortgage.</li>
</ol>

<p>Do those four things and the answer is usually obvious, and frequently not the one
you expected walking in.</p>
"""

# ─────────────────────────────────────────────────────────────────────────────
BODIES["credit-file-before-a-mortgage"] = """
<p>A mortgage underwriter and a personal loan underwriter look at the same credit
file and read different things from it. Understanding that difference is worth more
than any generic "improve your credit score" advice, because it tells you what
actually to do and when.</p>

<h2>There is no such thing as your credit score</h2>

<p>The number a credit reference agency shows you is that agency's own marketing
product. No lender uses it. Lenders build their own scorecards from the underlying
<em>data</em>, weight the factors themselves, and reach their own conclusions. Two
lenders can look at one file on the same afternoon and disagree completely.</p>

<p>So the useful work is not raising a number. It is making sure the underlying data
says what you want it to say.</p>

<h2>What a mortgage lender weighs more heavily than a loan lender</h2>

<ul>
<li><strong>Stability over time.</strong> A loan decision is largely about the
present. A mortgage decision is a bet on twenty-five years, so long-standing
accounts, a settled address history and continuous employment count for more.</li>
<li><strong>Every committed monthly payment.</strong> Not just whether you pay, but
what you have promised to pay. This is the affordability calculation, and it is
where existing loans do their damage &mdash;
<a href="/guides/how-loans-affect-mortgage-affordability.html">the arithmetic is
here</a>.</li>
<li><strong>Recent applications.</strong> A cluster of hard searches in the months
before a mortgage application reads as someone under financial pressure, whether or
not that is true, and whether or not you were accepted.</li>
<li><strong>Where the deposit came from.</strong> Anti-money-laundering rules mean
large recent credits into your account will be questioned. A gift needs a letter; a
loan used as a deposit is usually fatal to the application, and lenders are good at
spotting one.</li>
<li><strong>The current account itself.</strong> Bank statements typically get read
by a human. Gambling transactions, regular unarranged overdraft use and returned
direct debits all register, and none of them appear on a credit file at all.</li>
</ul>

<h2>Hard searches, soft searches, and the difference</h2>

<p>A <strong>soft search</strong> &mdash; an eligibility check, or looking at your
own file &mdash; is visible only to you and affects nothing. A <strong>hard
search</strong> is recorded when you formally apply for credit, is visible to other
lenders, and stays visible for around a year.</p>

<p>One hard search is unremarkable. Five in three months is a pattern, and it
suggests either that you keep needing credit or that you keep being declined. Both
readings hurt.</p>

<p>Practical rules:</p>

<ul>
<li>Use eligibility checkers, which are soft, before applying for anything.</li>
<li>Do not "shop around" by making real applications.</li>
<li>Leave a clear run &mdash; ideally six months, minimum three &mdash; between your
last credit application and your mortgage application.</li>
<li>Track them if you are applying for several things:
<a href="/loans/application-tracker.html">the application tracker</a> exists for
precisely this, and it stores everything in your own browser.</li>
</ul>

<h2>The six-month plan</h2>

<p>If you have half a year before you apply, this is the order that produces the most
improvement. There is a
<a href="/loans/credit-roadmap.html">month-by-month version as a tool</a>.</p>

<ol>
<li><strong>Month one: get all three files.</strong> Experian, Equifax and TransUnion
hold different data, and lenders do not all use the same one. You have a statutory
right of access. Check every account, every address, and every date.</li>
<li><strong>Month one: dispute anything wrong.</strong> Errors are common &mdash; a
settled debt still showing a balance, an account that was never yours, an address you
never lived at. Agencies must investigate. This is the single highest-return action
available and it costs nothing.</li>
<li><strong>Month two: get on the electoral roll.</strong> Lenders use it to confirm
you are who and where you say. Not being registered causes a surprising amount of
friction.</li>
<li><strong>Months two to five: bring balances down, especially cards.</strong>
Utilisation &mdash; balance against limit &mdash; matters and updates monthly.
Getting well below your limits is visible within a couple of cycles.</li>
<li><strong>Months two to five: clear the small, fast-repaying debts.</strong> Not
the biggest ones. The ones with the highest monthly payment relative to the balance,
because those are the ones costing you borrowing power per pound to clear.</li>
<li><strong>Throughout: apply for nothing.</strong> No cards, no car, no buy now pay
later, no phone contract you could pay for outright.</li>
<li><strong>Throughout: never miss a payment.</strong> One missed payment can undo
several months of everything above. Direct debits for at least the minimum, on
everything.</li>
<li><strong>Month six: clean statements.</strong> Assume a human will read three
months of your current account, because one will.</li>
</ol>

<div class="highlight-box">
<p class="mt-0"><strong>Do not close old accounts to tidy up.</strong> A credit card
you have held for eleven years and do not use is an asset on your file &mdash; long
history, available credit, no balance. Closing it shortens your history and raises
your utilisation ratio at a stroke. Empty it; keep it.</p>
</div>

<h2>If something on your file is genuinely bad</h2>

<p>Defaults, county court judgments and arrangements to pay stay on your file for
<strong>six years</strong> from the date recorded. You cannot remove accurate
adverse data, and any firm promising to is selling you something worthless.</p>

<p>What you can do is understand how it ages. Most lenders care far more about a
default from eight months ago than one from four years ago, and several will lend
normally once adverse data is three or more years old and everything since has been
clean. There are also lenders who specialise in exactly this, at a price.</p>

<p>So the answer to "I have a default, can I get a mortgage" is usually yes, and the
useful questions are <em>when</em> and <em>at what rate</em> &mdash; which is
territory where a broker earns their fee, because knowing which lenders take which
view of which adverse data is the whole job.</p>
"""

# ─────────────────────────────────────────────────────────────────────────────
BODIES["secured-vs-unsecured-what-changes"] = """
<p>The difference everyone knows is the interest rate: secured borrowing is cheaper.
The difference that actually matters is what the lender can do when you stop
paying, and it is the reason the rate is lower in the first place.</p>

<h2>The mechanism, in one paragraph</h2>

<p>A secured debt is attached to an asset &mdash; almost always your home. If you
fail to pay, the lender has a route to that asset. That route makes the loan far
less risky for them, which is why they charge less for it. <strong>The lower rate is
not a favour. It is the price of the security, and you are the one who
provided it.</strong></p>

<h2>What happens when it goes wrong</h2>

<table>
<thead><tr><th>&nbsp;</th><th>Unsecured</th><th>Secured on your home</th></tr></thead>
<tbody>
<tr><td>Typical rate</td><td>Higher</td><td>Lower</td></tr>
<tr><td>Typical term</td><td>1&ndash;7 years</td><td>Up to the rest of your mortgage term</td></tr>
<tr><td>First consequence of non-payment</td><td>Arrears, then a default on your file</td><td>Arrears, then contact from the lender's collections team</td></tr>
<tr><td>Escalation</td><td>Debt collection, then possibly a county court judgment</td><td>Possession proceedings</td></tr>
<tr><td>Worst realistic outcome</td><td>CCJ, six years of damaged credit, enforcement against income or goods</td><td><strong>You lose your home</strong></td></tr>
<tr><td>Can you negotiate?</td><td>Often substantially &mdash; unsecured creditors settle</td><td>Less room; the lender has a stronger position and knows it</td></tr>
</tbody>
</table>

<h2>The bit almost nobody is told: unsecured debt can become secured</h2>

<p>People treat the categories as permanent. They are not.</p>

<p>If an unsecured creditor obtains a county court judgment and you do not pay it,
they can apply for a <strong>charging order</strong> over your property. That
converts an unsecured debt into one secured against your home, without you ever
having agreed to give security. In some circumstances an order for sale can follow.</p>

<p>This is not the common outcome and it takes time and several steps, all of which
you get notice of. But it changes how you should read the table above: the wall
between the two columns has a door in it, and ignoring court paperwork is how you go
through it. <strong>If you are ever served with a claim form, respond to it.</strong>
Responding is free and it preserves every option you have.</p>

<h2>Second charge loans</h2>

<p>A "homeowner loan", "second charge" or "secured loan" sits behind your mortgage on
the same property. Your mortgage lender is the first charge and gets paid first; the
second charge lender gets what is left.</p>

<p>Two things worth knowing. First, because they are second in the queue, they price
higher than a mortgage &mdash; so this is not cheap borrowing, merely cheaper than
unsecured. Second, they are regulated under the same mortgage rules as a first
charge, which means proper affordability assessment and proper complaint rights.
That is a genuine protection and it is worth confirming any lender you are talking to
is authorised.</p>

<p>The honest use case is real: you need a substantial sum, your existing mortgage
has an early repayment charge that makes remortgaging expensive, and a second charge
avoids disturbing a good rate on the main loan. Price it against the
<a href="/guides/consolidating-debt-into-your-mortgage.html">alternatives</a>,
including simply waiting for the fix to end.</p>

<h2>How to decide</h2>

<p>The rate is the last question, not the first. Work through these in order:</p>

<ol>
<li><strong>Is this borrowing for something that lasts?</strong> Securing a
twenty-year debt against your home to fund a structural improvement is defensible.
Securing it to fund a holiday or clear cards you will refill is not. Match the term
of the debt to the life of the thing.</li>
<li><strong>What is my income risk over the term?</strong> Not today &mdash; over
the whole term. Self-employed, on commission, in a volatile sector, one income
supporting a household: all argue for keeping debt unsecured even at a higher
rate.</li>
<li><strong>Could I service this if rates rose?</strong> Run
<a href="/loans/interest-rate-stress-test.html">the stress test</a>. If the answer
is uncomfortable, the extra risk of security is not a good trade.</li>
<li><strong>What is the total cost, both ways?</strong> Total repayable, not monthly
payment &mdash; see <a href="/guides/total-cost-of-borrowing.html">why</a>. Short
unsecured frequently beats long secured on total cost.</li>
<li><strong>Only then, the rate.</strong></li>
</ol>

<div class="highlight-box">
<p class="mt-0"><strong>The sentence to hold on to.</strong> Unsecured debt can cost
you your credit rating. Secured debt can cost you your home. Any time you are
converting the first into the second, that is the trade you are making &mdash;
whatever the monthly payment says.</p>
</div>
"""

# ─────────────────────────────────────────────────────────────────────────────
BODIES["car-finance-and-your-mortgage"] = """
<p>If you are buying a house in the next couple of years, the car decision is a
mortgage decision. Most people do not find this out until an underwriter tells
them.</p>

<h2>Why a car agreement costs you so much borrowing power</h2>

<p>Car finance is a committed monthly payment, and mortgage affordability works by
subtracting committed monthly payments from the income available to service a
mortgage. At roughly £5,000&ndash;£7,000 of lost borrowing power per £100 a month,
a typical car agreement of £300&ndash;£400 a month costs you somewhere in the
region of <strong>£15,000 to £28,000 of mortgage</strong>.</p>

<p>Not £300. Twenty thousand pounds of house, give or take, for as long as the
agreement runs. Put your own numbers through the
<a href="/mortgages/affordability.html">affordability calculator</a> with and
without the car payment and the gap is usually sobering.</p>

<h2>PCP and HP are not the same shape</h2>

<p><strong>Hire purchase</strong> spreads the whole cost of the car over the term. At
the end you have paid for it and you own it. The monthly payment is higher because
you are actually buying a car.</p>

<p><strong>Personal contract purchase</strong> spreads only the depreciation, then
leaves a large optional final payment &mdash; the balloon, based on the car's
guaranteed future value. The monthly payment is lower because you are, in effect,
paying to use the car rather than to own it. At the end you hand it back, pay the
balloon, or roll into another agreement.</p>

<p>Two consequences that matter here:</p>

<ul>
<li><strong>PCP's lower monthly payment does less damage to your mortgage
affordability</strong> than HP's higher one, for the same car. If you must have
finance and you are also applying for a mortgage, this is a real argument for
PCP &mdash; and it is the opposite of the usual advice, which is about total
cost.</li>
<li><strong>PCP usually costs more overall and leaves you owning nothing.</strong>
Compare the totals properly with the
<a href="/loans/car-finance-calculator.html">PCP versus HP calculator</a>, which
shows the balloon rather than hiding it.</li>
</ul>

<p>So the two goals genuinely conflict, and you should pick deliberately rather than
letting a dealer pick for you.</p>

<h2>The order of operations</h2>

<p>If both purchases are happening, the sequence matters more than almost anything
else:</p>

<ol>
<li><strong>Mortgage first, car second.</strong> Always, if you have the choice.
Once the mortgage has completed, a car agreement affects nothing about it. Do it the
other way round and the car reduces the house.</li>
<li><strong>If the car cannot wait,</strong> keep the monthly payment as low as you
can live with and understand you are trading house for car.</li>
<li><strong>Never take car finance between mortgage application and
completion.</strong> This is the expensive mistake. Lenders re-check, and a new
credit agreement appearing after a decision in principle can collapse the offer at
the point where you have already paid for a survey and legal work.</li>
</ol>

<div class="highlight-box">
<p class="mt-0"><strong>A decision in principle is not a mortgage.</strong> The
lender can and does look again before completion. Between application and keys,
change nothing: no new credit, no job change if avoidable, no large unexplained
movements of money.</p>
</div>

<h2>Existing agreement, mortgage coming: the options</h2>

<ul>
<li><strong>Settle it.</strong> Ask for a settlement figure &mdash; you have a
statutory right to one. Clearing the agreement removes the monthly payment and
restores the borrowing power. Best option if you can afford it, and note that the
figure will include some early settlement interest;
<a href="/loans/settlement-calculator.html">estimate it here</a>.</li>
<li><strong>Voluntary termination.</strong> Once you have paid <strong>50%</strong>
of the total amount payable under a regulated HP or PCP agreement, you have a
statutory right under the Consumer Credit Act to end it and hand the car back,
owing nothing further &mdash; provided the car is in reasonable condition. This is a
genuine and under-used right. It is not the same as voluntary surrender, which is a
negotiated hand-back with the debt still owing, and the two get confused
constantly.</li>
<li><strong>Wait for it to end</strong>, and time the mortgage application for
after.</li>
<li><strong>Do nothing and accept the reduced figure</strong>, having checked with
the affordability calculator that it still buys what you need.</li>
</ul>

<h2>Handing a car back: the charges that appear at the end</h2>

<p>Return charges are the most common unpleasant surprise in car finance, and they
are largely predictable. Finance companies assess condition against the
<strong>BVRLA fair wear and tear standard</strong>, which is published and specific
&mdash; it says how large a scratch may be, what kerbing on an alloy is acceptable,
what counts as damage rather than use.</p>

<p>Two practical moves. Inspect the car against that standard a couple of months
before the return date, not on the day, which leaves time to get small repairs done
independently for a fraction of what the finance company charges. And go over the
excess mileage terms: the per-mile rate is small and the totals are not.</p>

<p>The <a href="/loans/damage-checker.html">return damage checker</a> walks the
standard item by item so you can see where you stand before an inspector does.</p>
"""

# ─────────────────────────────────────────────────────────────────────────────
BODIES["remortgaging-with-other-debt"] = """
<p>Remortgaging is the moment your unsecured borrowing and your mortgage meet. It is
also, for most households, the largest single financial decision they will make
without advice. Both facts argue for slowing down.</p>

<h2>Start here: what are you actually remortgaging for?</h2>

<p>Three different jobs get called "remortgaging" and they have different right
answers.</p>

<ol>
<li><strong>Rate only.</strong> Your fix is ending; you want another one. Same
balance, same term. Usually straightforward and usually worth doing.</li>
<li><strong>Rate plus a bit more borrowing</strong> for something specific &mdash; an
extension, a new roof. Reasonable, and priced as additional lending.</li>
<li><strong>Rate plus clearing unsecured debt.</strong> This is the consequential one.
It has <a href="/guides/consolidating-debt-into-your-mortgage.html">its own
guide</a>, because the monthly saving is real and so are two costs that are not
obvious.</li>
</ol>

<p>Be honest about which you are doing. Job three dressed as job one is how people
end up with a thirty-year commitment they did not deliberately choose.</p>

<h2>The timing that saves the most money</h2>

<p>Two dates govern everything: when your early repayment charge ends, and when your
fixed rate ends. They are usually the same date but not always, and the gap between
them is worth knowing about.</p>

<ul>
<li><strong>Leaving during the fix</strong> normally triggan early repayment charge
&mdash; commonly one to five per cent of the balance. On a substantial mortgage that
is thousands of pounds and it will usually swamp any saving.</li>
<li><strong>Falling off the fix onto the standard variable rate</strong> is the
expensive default. SVRs are typically well above anything you could arrange, and
lenders are not obliged to remind you.</li>
<li><strong>The window</strong> is the three to six months before your fix ends. Most
offers are valid for that long, so you can secure a rate early and complete the day
the fix expires &mdash; no early repayment charge, no month on the SVR.</li>
</ul>

<p>Put the date your fix ends in a calendar now, with a reminder six months before
it. That single act is worth more than most rate-shopping.</p>

<h2>What has changed since you last did this</h2>

<p>A remortgage is a full application, not a formality, and things that were fine
five years ago may not be:</p>

<ul>
<li><strong>Your commitments.</strong> Car finance or loans taken since last time
reduce what you can borrow &mdash;
<a href="/guides/how-loans-affect-mortgage-affordability.html">the arithmetic</a>
applies to a remortgage exactly as to a purchase.</li>
<li><strong>Your income shape.</strong> Moving from employment to self-employment is
the big one; most lenders want two or three years of accounts.</li>
<li><strong>Your property's value,</strong> and therefore your LTV band. If prices
have risen, your LTV has improved and you may reach a cheaper band without doing
anything. If they have fallen, the reverse &mdash; and in the worst case there is
no equity to remortgage against at all.</li>
<li><strong>Your age against the term.</strong> Lenders cap the age at which the
mortgage must end, which can force a shorter term and a higher payment.</li>
</ul>

<div class="highlight-box">
<p class="mt-0"><strong>Check your LTV band before you shop.</strong> Take a current
valuation estimate, divide the outstanding balance by it, and see which side of a
band boundary you are on. If you are just the wrong side, a modest lump sum off the
balance before you apply can move you into a cheaper band &mdash; and that reduction
applies to the whole mortgage. It is the highest-return small payment available at
this moment.</p>
</div>

<h2>A product transfer versus a full remortgage</h2>

<p>Your existing lender will offer you a new deal &mdash; a product transfer.
Usually no valuation, minimal paperwork, often no affordability reassessment, and
quick.</p>

<p>It is genuinely useful in two situations: when your circumstances have worsened
and you might not pass a fresh assessment elsewhere, and when you simply want it
done. The cost is that you are choosing from one lender's range without knowing what
the rest of the market would offer, and the difference over a fixed period is
frequently substantial.</p>

<p>The sensible approach is to get the product transfer offer, then compare it
properly against the open market on <em>total cost over the fixed period</em> rather
than on rate &mdash; the <a href="/mortgages/fee-analyser.html">fee analyser</a>
does exactly this comparison, because a fee-free transfer can beat a lower rate that
comes with a large product fee.</p>

<h2>The sequence</h2>

<ol>
<li><strong>Six months out:</strong> find the date your fix ends and your ERC
position. Get your credit files and fix anything wrong on them.</li>
<li><strong>Five months out:</strong> stop applying for credit. Bring card balances
down.</li>
<li><strong>Four months out:</strong> work out your LTV and whether a lump sum
crosses a band. Decide whether debt consolidation is part of this, deliberately and
on paper.</li>
<li><strong>Three months out:</strong> get your lender's product transfer offer, and
compare the market against it on total cost. Consider a broker &mdash; on a
remortgage with any complication, they usually earn their fee.</li>
<li><strong>Secure the offer</strong>, timed to complete as the fix ends.</li>
<li><strong>Change nothing</strong> between offer and completion.</li>
</ol>

<p>If the reason you are remortgaging is that the payments have become genuinely
unmanageable, that is a different problem and a remortgage may not be the answer.
<a href="/guides/when-repayments-are-a-struggle.html">Start here instead</a> &mdash;
and speak to your lender before you miss anything, because the options available
before arrears are much better than the ones after.</p>
"""

# ─────────────────────────────────────────────────────────────────────────────
BODIES["fixed-vs-variable-on-both"] = """
<p>Most households fix the mortgage and leave everything else floating. It is worth
asking whether that is a decision or just what happened, because on the numbers it
is often the wrong way round.</p>

<h2>What fixing actually buys</h2>

<p>Fixing does not buy you a lower rate. On average, over time, it costs slightly
more &mdash; the lender is pricing in the risk they are taking off your hands.</p>

<p>What it buys is <strong>certainty</strong>, and certainty is worth most where a
rate change would hurt most. So the useful question is not "will rates rise?"
&mdash; nobody knows &mdash; but "which of my debts would a rate rise hurt most, and
have I protected that one?"</p>

<h2>Ranking your debts by rate sensitivity</h2>

<p>Two things determine how much a rate rise costs you on a given debt: the
<strong>balance</strong> and the <strong>remaining term</strong>. Large and long is
maximum exposure; small and short is almost none.</p>

<p>Which usually produces this order:</p>

<ol>
<li><strong>The mortgage.</strong> Largest balance, longest term. A one-point rise
here costs more in pounds than a one-point rise on everything else combined. This is
why fixing the mortgage is conventional wisdom, and the conventional wisdom is
right.</li>
<li><strong>Any second charge or homeowner loan.</strong> Same logic, smaller
balance.</li>
<li><strong>Credit cards and overdrafts.</strong> Variable by nature and repriceable
at will. Usually smaller balances, but the rates are high to start with, and card
rates move for reasons that have nothing to do with the Bank of England.</li>
<li><strong>Personal loans and car finance.</strong> Almost always fixed at the
outset in the UK, so there is frequently nothing to decide. Worth checking rather
than assuming.</li>
</ol>

<div class="highlight-box">
<p class="mt-0"><strong>The check most people skip.</strong> Rather than reasoning
about it, model it. Put every variable-rate balance through the
<a href="/loans/interest-rate-stress-test.html">rate stress test</a> and the
mortgage through the <a href="/mortgages/rate-forecaster.html">payment change
forecaster</a>, at the same rate rise. Add up the extra monthly cost. That total is
your genuine exposure, and it is normally larger than expected because it was never
totalled before.</p>
</div>

<h2>Choosing a fixed period</h2>

<p>Longer fixes cost a little more and protect for longer. The trade is not really
about rates, though, it is about your life:</p>

<ul>
<li><strong>Short fix</strong> if you might move, if your income is about to change
materially, or if you expect to make large overpayments that a longer deal's terms
would restrict.</li>
<li><strong>Long fix</strong> if you are staying put, your budget is tight enough
that a payment rise would genuinely hurt, and you value not thinking about it.</li>
<li><strong>Watch the early repayment charge either way.</strong> A five-year fix you
exit in year two can cost thousands. The ERC schedule is the part of the deal to read
properly &mdash; it is where a long fix stops being free optionality and becomes a
commitment.</li>
</ul>

<p>One under-appreciated point: <strong>check whether the deal is portable.</strong>
A portable mortgage can move with you to a new property, which preserves a good rate
through a move and avoids the ERC. It is not automatic and it is not guaranteed at
the time you move &mdash; you still have to qualify on the new property &mdash; but
its absence is worth knowing about before you sign a five-year deal.</p>

<h2>Trackers, discounts and the standard variable rate</h2>

<p>Three different things get called "variable" and they behave differently:</p>

<ul>
<li>A <strong>tracker</strong> follows the Bank of England base rate by a set margin.
It moves when base rate moves, transparently and by a known amount.</li>
<li>A <strong>discount</strong> is a reduction from the lender's own standard
variable rate. Since the lender sets the SVR, they can move your rate without base
rate moving at all.</li>
<li>The <strong>standard variable rate</strong> is the lender's default, where you
land when a deal ends. It is typically well above what you could arrange, and
drifting onto it is the single most common avoidable mortgage cost.</li>
</ul>

<p>If you want variable, a tracker is the honest version: you can see what governs
it. A discount off an SVR gives the lender a discretion you cannot control.</p>

<h2>The answer for most people</h2>

<p>Fix the mortgage, for a period matched to how settled your life is. Then, instead
of leaving the rest floating by default, look at your variable-rate balances and
<strong>clear them rather than trying to protect them</strong> &mdash; because on
unsecured debt the interest rate is high whether it is fixed or variable, and
removing the balance is worth far more than fixing the rate on it. The
<a href="/loans/overpayment-calculator.html">overpayment calculator</a> will show
you what that is worth, and it is usually a better return than any mortgage decision
available to you.</p>
"""

# ─────────────────────────────────────────────────────────────────────────────
BODIES["stress-testing-the-whole-budget"] = """
<p>Your mortgage lender stress-tests your mortgage. Nobody stress-tests the rest of
your borrowing, and nobody adds the two together. That gap is where household
financial trouble actually starts.</p>

<h2>What the lender does, and what it leaves out</h2>

<p>A mortgage lender checks you could still afford the mortgage payment if rates were
higher than the rate you are being offered. It is a real test and it protects you as
well as them.</p>

<p>But it is a test of <em>one</em> debt against <em>one</em> shock. It assumes your
income holds, your other commitments stay as declared, and nothing else happens.
Actual financial difficulty rarely arrives that tidily. It usually arrives as two
things at once: the rate rose <em>and</em> the hours were cut; the fix ended
<em>and</em> the car needed replacing.</p>

<h2>Running the test properly</h2>

<p>Do this once, on paper, and keep it. It takes twenty minutes.</p>

<h3>Step one: the current picture</h3>

<p>List every monthly commitment: mortgage or rent, every loan, every card minimum,
car finance, insurance, council tax, energy, water, phone, essential subscriptions,
food, transport. Total it. Compare against take-home pay. The difference is your
actual monthly margin &mdash; the number that determines whether a shock is an
inconvenience or a crisis.</p>

<h3>Step two: the rate shock</h3>

<p>Take every variable or soon-to-expire debt and re-price it two or three
percentage points higher.</p>

<ul>
<li>Mortgage: <a href="/mortgages/rate-forecaster.html">payment change
forecaster</a>. If your fix ends within two years, use the higher rate, not
today's.</li>
<li>Everything variable: <a href="/loans/interest-rate-stress-test.html">rate stress
test</a>, one debt at a time.</li>
</ul>

<p>Add up the new payments. Subtract from income. <strong>Is the margin still
positive?</strong></p>

<h3>Step three: the income shock</h3>

<p>Now, separately, hold rates where they are and cut household income by a third
&mdash; one earner losing hours, a bonus not arriving, a period of illness. Same
question.</p>

<h3>Step four: both at once</h3>

<p>This is the one that matters, and it is the one nobody runs. Higher rates
<em>and</em> lower income. If the margin is still positive here, you are genuinely
robust. If it is not &mdash; and for a lot of households it is not &mdash; you have
found something useful, early, while there are still cheap options.</p>

<div class="highlight-box">
<p class="mt-0"><strong>Why do this before you borrow more, not after.</strong>
Before, every option is open: borrow less, keep the term short, clear a debt first,
wait. After, your options are the expensive ones. The test costs twenty minutes and
it is the highest-value twenty minutes in personal finance.</p>
</div>

<h2>What to do when the answer is uncomfortable</h2>

<p>In rough order of effect:</p>

<ol>
<li><strong>Build the emergency fund first.</strong> Three to six months of essential
outgoings, in an instant-access account. This is what converts a shock into an
inconvenience, and no clever debt structuring substitutes for it.</li>
<li><strong>Kill the highest-payment-to-balance debts.</strong> Those are the ones
consuming the most monthly margin per pound owed. Clearing them buys back margin
fastest &mdash; and, as it happens, buys back mortgage borrowing power too.</li>
<li><strong>Fix what you cannot afford to have move.</strong> If the mortgage rising
would break the budget, certainty is worth paying for.</li>
<li><strong>Borrow less than the maximum.</strong> The lender's maximum is a
constraint, not a recommendation. Nobody has ever regretted a mortgage that was
comfortably affordable.</li>
<li><strong>Keep the term short where you can.</strong> A longer term is a lower
payment and a higher total cost. If a long term is the only way the sums work, that
itself is information about whether the purchase is the right size.</li>
</ol>

<h2>Redo it when something changes</h2>

<p>The test goes out of date. Re-run it when your fix is within a year of ending,
when you take on any new commitment, when your income changes shape, and once a year
regardless. It is much quicker the second time.</p>

<p>And if the honest answer today is that the margin is already negative or close to
it, do not wait for the stress test to become real.
<a href="/guides/when-repayments-are-a-struggle.html">There are free services who
will help</a>, and every one of them will tell you the same thing: the options are
much better before arrears than after.</p>
"""

# ─────────────────────────────────────────────────────────────────────────────
BODIES["the-fees-nobody-quotes"] = """
<p>Rates are advertised. Fees are disclosed. Those are different verbs, and the gap
between them is where a meaningful amount of money lives.</p>

<p>None of what follows is hidden in the sense of unlawful &mdash; it is all in the
paperwork, and firms are required to tell you. It is simply not in the headline, and
the headline is what people compare.</p>

<h2>On a mortgage</h2>

<ul>
<li><strong>Product or arrangement fee.</strong> Frequently several hundred to around
a thousand pounds. You can usually add it to the loan, which means paying interest on
it for the whole term &mdash; so a £999 fee added to a mortgage is considerably more
than £999. This single fee is why the market-leading rate is often not the cheapest
deal, and the <a href="/mortgages/fee-analyser.html">fee analyser</a> exists to
settle that comparison on total cost.</li>
<li><strong>Valuation fee.</strong> Sometimes free, sometimes not, and a full
structural survey is a separate and much larger cost that is nothing to do with the
lender.</li>
<li><strong>Legal fees.</strong> Sometimes included on a remortgage, rarely on a
purchase.</li>
<li><strong>Early repayment charge.</strong> The big one. Typically one to five per
cent of the balance if you leave during the fixed period, tapering by year. On
£250,000 at three per cent that is £7,500 &mdash; and it is the reason a five-year
fix is a commitment rather than merely a longer deal. Read the schedule.</li>
<li><strong>Exit or deeds release fee</strong> at the end, a modest but real
administrative charge.</li>
<li><strong>Higher lending charge</strong> at very high loan-to-value, less common
now but not extinct.</li>
</ul>

<h2>On a loan or credit agreement</h2>

<ul>
<li><strong>Arrangement fees.</strong> Uncommon on mainstream personal loans,
frequent at the subprime end.</li>
<li><strong>Early settlement interest.</strong> You have a statutory right to settle
early, but under the early settlement rules the lender may add up to
<strong>28 days'</strong> interest after the settlement date &mdash; and for
agreements longer than twelve months, up to <strong>30 days</strong> beyond that.
That is the "58-day rule". So settling never costs the full remaining interest, and
never costs merely the outstanding balance either. <a
href="/loans/settlement-calculator.html">Estimate it here</a> before you assume
paying off early is free.</li>
<li><strong>Default and late payment fees,</strong> plus the interest that continues
to accrue on top.</li>
<li><strong>Optional insurance,</strong> priced monthly and presented as small. Check
whether you already have the cover through something else.</li>
</ul>

<h2>On car finance specifically</h2>

<ul>
<li><strong>The balloon payment.</strong> Not a fee, but the largest number in a PCP
agreement and the one the monthly figure is designed to distract from. See <a
href="/loans/car-finance-calculator.html">PCP versus HP</a>.</li>
<li><strong>Excess mileage.</strong> A small-sounding per-mile rate against a limit
that is easy to exceed. The totals are not small.</li>
<li><strong>Return condition charges,</strong> assessed against the BVRLA fair wear
and tear standard. Largely predictable, and largely avoidable if you inspect early
&mdash; the <a href="/loans/damage-checker.html">damage checker</a> walks the
standard.</li>
<li><strong>Option to purchase fee</strong> on HP, a small charge to actually take
ownership at the end.</li>
</ul>

<h2>Which of these are negotiable</h2>

<p>More than you would think, and it costs nothing to ask.</p>

<ul>
<li><strong>Often negotiable:</strong> mortgage product fees (frequently there is a
fee-free version of the same deal at a slightly higher rate &mdash; ask for both and
compare on total cost); valuation fees; legal fees on a remortgage; car dealer
admin fees.</li>
<li><strong>Sometimes waived on request:</strong> excess mileage, if you are taking
another agreement with the same finance company.</li>
<li><strong>Rarely negotiable:</strong> early repayment charges, statutory settlement
interest, Stamp Duty.</li>
</ul>

<div class="highlight-box">
<p class="mt-0"><strong>The one question that surfaces all of it.</strong> Ask, in
writing: <em>"What is the total amount I will have paid by the end of this
agreement, including every fee, assuming I do nothing unusual?"</em> A regulated firm
can answer it, and the answer is directly comparable between offers in a way that a
rate is not.</p>
</div>

<h2>Where to find them before you sign</h2>

<p>For a mortgage, the illustration or European Standardised Information Sheet lists
fees and the APRC. For a regulated credit agreement, the pre-contract information
sets out the total amount payable, the APR and any charges. Both are required, both
are dull, and both are the only place the real price is written down.</p>

<p>Read the total amount payable. Compare that between offers, and nothing else. See
<a href="/guides/total-cost-of-borrowing.html">why that is the only number worth
comparing</a>.</p>
"""

# ─────────────────────────────────────────────────────────────────────────────
BODIES["when-repayments-are-a-struggle"] = """
<p>If you are reading this because the payments have stopped adding up, two things
are worth saying before anything else.</p>

<p><strong>Free help exists, it is good, and it is not the same as the companies that
advertise.</strong> And <strong>acting early gives you materially better options
than acting late.</strong> Almost everything below works better this month than next.</p>

<h2>First: separate priority debts from the rest</h2>

<p>This is the step that changes outcomes, and it is not intuitive, because the
creditors who chase hardest are usually not the ones who matter most.</p>

<p><strong>Priority debts</strong> &mdash; pay these first, always. The consequences
of not paying are losing your home, losing your energy supply, or court action:</p>

<ul>
<li>Mortgage or rent</li>
<li>Council tax</li>
<li>Gas and electricity</li>
<li>Court fines</li>
<li>Child maintenance</li>
<li>Income tax, VAT and National Insurance</li>
<li>TV licence</li>
</ul>

<p><strong>Non-priority debts</strong> &mdash; serious, but the consequences are
financial rather than existential: credit cards, personal loans, overdrafts,
catalogues, buy now pay later, most other unsecured borrowing.</p>

<p>Credit card companies phone daily. Councils and mortgage lenders often do not. It
is easy to end up paying the loudest creditor while arrears build with the one who
could take your home. <strong>Pay priority debts first even while non-priority
creditors are shouting.</strong> A free adviser will tell you exactly this, and will
tell your other creditors so on your behalf.</p>

<h2>Second: talk to your mortgage lender before you miss a payment</h2>

<p>Lenders are required to treat customers in financial difficulty fairly, and they
have a range of options they can apply &mdash; but nearly all of them are easier to
arrange before arrears than after:</p>

<ul>
<li>A temporary reduction in payments</li>
<li>A switch to interest-only for a period</li>
<li>Extending the term to lower the monthly payment</li>
<li>A payment holiday, in some circumstances</li>
<li>Adding arrears to the balance</li>
</ul>

<p>Possession is a last resort and lenders have to demonstrate they explored the
alternatives. But the conversation is much better when you start it. If you have
already fallen behind, it is still worth having &mdash; just have it today.</p>

<h2>Third: get free advice, from one of these</h2>

<p>All of these are free, impartial, and will deal with creditors for you. None will
charge you a fee, and none of them needs to.</p>

<ul>
<li><strong><a href="https://www.stepchange.org">StepChange</a></strong> &mdash; the
largest UK debt charity. Full debt advice and managed plans.</li>
<li><strong><a href="https://nationaldebtline.org">National Debtline</a></strong>
&mdash; free advice by phone and webchat, with excellent written guides and template
letters.</li>
<li><strong><a href="https://www.citizensadvice.org.uk">Citizens Advice</a></strong>
&mdash; face-to-face help locally, and the strongest option if your situation
involves benefits, housing or employment as well as debt.</li>
<li><strong><a href="https://www.moneyhelper.org.uk">MoneyHelper</a></strong> &mdash;
government-backed guidance and a directory of regulated advisers.</li>
<li><strong><a href="https://www.mind.org.uk">Mind</a></strong> &mdash; because debt
and mental health run together far more often than either gets discussed, and either
one makes the other harder.</li>
</ul>

<div class="highlight-box">
<p class="mt-0"><strong>Be careful who you call.</strong> Search results for debt
help are full of commercial firms charging for what the charities above do free, and
some steer people toward solutions that pay the firm best rather than suit the
person. If a service asks for a fee, stop and call one of the five above instead.</p>
</div>

<h2>Breathing Space</h2>

<p>In England and Wales the <strong>Debt Respite Scheme</strong> &mdash; usually
called Breathing Space &mdash; gives you up to <strong>60 days</strong> during which
most interest, fees and enforcement action stop while you get advice and make a plan.
There is a longer version for people receiving mental health crisis treatment.</p>

<p>You apply through a debt adviser rather than directly, which is another reason the
first call is to one of the organisations above. Scotland has its own statutory
moratorium under the Debt Arrangement Scheme.</p>

<h2>What not to do</h2>

<ul>
<li><strong>Do not borrow to pay debt</strong> without proper advice. It is the
instinct, and it usually makes things worse &mdash; particularly if it means securing
previously unsecured debt against your home. See
<a href="/guides/secured-vs-unsecured-what-changes.html">what that actually
changes</a>.</li>
<li><strong>Do not ignore court paperwork.</strong> A claim form is not the end of
anything &mdash; responding is free and preserves every option. Ignoring it produces
a judgment by default, which is how an unsecured debt can eventually become a charge
on your house.</li>
<li><strong>Do not pay a fee for debt advice.</strong> There is no version of this
worth paying for that the charities do not do better.</li>
<li><strong>Do not prioritise the loudest creditor.</strong> Priority debts first.
See above.</li>
<li><strong>Do not cancel your home or buildings insurance</strong> to save money if
you have a mortgage. It is usually a condition of the loan, and an uninsured loss
would be catastrophic.</li>
</ul>

<h2>The order, in one list</h2>

<ol>
<li>Work out your actual income and essential outgoings. Honestly, on paper.</li>
<li>Sort your debts into priority and non-priority.</li>
<li>Contact your mortgage lender or landlord and your council. Today, not after the
next missed payment.</li>
<li>Call one of the five free services. Let them contact the non-priority
creditors.</li>
<li>Ask your adviser about Breathing Space.</li>
<li>Do not sign anything, and do not borrow anything, until you have had that
conversation.</li>
</ol>

<p>This is a much more common situation than it feels like from inside it, the
services above deal with it every day, and it is fixable. The hardest part is the
first phone call.</p>
"""

# ─────────────────────────────────────────────────────────────────────────────
BODIES["jargon-buster"] = """
<p>Loans and mortgages use two overlapping vocabularies, and the overlap is where
confusion lives &mdash; APR and APRC are not the same measure, and a "variable rate"
means something different depending on which product you are holding. Both sets, in
one place.</p>

<h2>Rates and how cost is measured</h2>

<dl>
<p><strong>APR</strong> &mdash; Annual Percentage Rate. Interest plus compulsory
fees, expressed as a yearly rate, for consumer credit. Designed for comparing two
agreements <em>of the same amount and term</em>. It cannot see term length, so it
will not tell you that a seven-year loan costs far more than a three-year one at the
same APR.</p>

<p><strong>Representative APR</strong> &mdash; the rate offered to at least
<strong>51%</strong> of accepted applicants. Up to 49% can lawfully be paying more.
Not a quote.</p>

<p><strong>APRC</strong> &mdash; Annual Percentage Rate of Charge, the mortgage
equivalent. Includes fees and assumes you keep the mortgage for its full term, which
almost nobody does &mdash; so it is a poor way to choose between fixed deals. Compare
total cost over the fixed period instead.</p>

<p><strong>Total amount payable / total cost of credit</strong> &mdash; everything you
hand over by the end. The only figure that cannot be flattered by stretching the
term, and therefore the one to compare. See
<a href="/guides/total-cost-of-borrowing.html">why</a>.</p>

<p><strong>Base rate</strong> &mdash; the Bank of England's rate. Trackers follow it;
standard variable rates are influenced by it but set by the lender.</p>

<p><strong>SVR</strong> &mdash; Standard Variable Rate. Your lender's default rate,
where you land when a deal ends. Typically well above what you could arrange.
Drifting onto it is the most common avoidable mortgage cost.</p>

<p><strong>Tracker</strong> &mdash; follows base rate by a fixed margin. Moves
transparently.</p>

<p><strong>Discount rate</strong> &mdash; a reduction from the lender's own SVR.
Because the lender sets the SVR, they can move your rate without base rate
moving.</p>

<p><strong>Fixed rate</strong> &mdash; the rate cannot change for a set period.
Certainty, not cheapness. See
<a href="/guides/fixed-vs-variable-on-both.html">choosing between them</a>.</p>
</dl>

<h2>Mortgage-specific</h2>

<dl>
<p><strong>LTV</strong> &mdash; Loan to Value. Borrowing as a percentage of property
value. Priced in <em>bands</em>, and crossing a band boundary changes the rate on the
whole mortgage &mdash; which is why a small extra deposit is sometimes worth a great
deal and sometimes worth almost nothing.</p>

<p><strong>LTI</strong> &mdash; Loan to Income. Borrowing as a multiple of income.
Lenders are limited in the proportion of new lending they may write above
<strong>4.5x</strong>, which is why offers cluster below it.</p>

<p><strong>ERC</strong> &mdash; Early Repayment Charge. What leaving a deal early
costs, commonly 1&ndash;5% of the balance, tapering by year. The part of a long fix
that makes it a commitment.</p>

<p><strong>Product fee</strong> &mdash; arrangement fee for a specific mortgage deal.
Add it to the loan and you pay interest on it for the whole term.</p>

<p><strong>Porting</strong> &mdash; moving your existing mortgage deal to a new
property, preserving the rate and avoiding the ERC. Not automatic; you must qualify
again.</p>

<p><strong>Product transfer</strong> &mdash; a new deal from your existing lender
without a full remortgage. Quick, less paperwork, but you are choosing from one
lender's range.</p>

<p><strong>Repayment vs interest-only</strong> &mdash; repayment clears capital and
interest, so the debt ends. Interest-only pays interest alone; the full balance is
still owed at the end and you need a credible way to repay it.</p>

<p><strong>Amortisation</strong> &mdash; how a repayment balance actually reduces.
Early payments are mostly interest; capital repayment accelerates later. Visible in
the <a href="/mortgages/repayment.html">repayment calculator</a>.</p>

<p><strong>Negative equity</strong> &mdash; the mortgage exceeds the property's
value. Remortgaging and moving both become very difficult.</p>

<p><strong>Decision in principle / AIP</strong> &mdash; an indication, not a mortgage.
The lender can and does look again before completion.</p>

<p><strong>Second charge</strong> &mdash; a loan secured on your property behind the
main mortgage. Regulated as a mortgage. Priced higher than a first charge, lower than
unsecured.</p>

<p><strong>SDLT</strong> &mdash; Stamp Duty Land Tax, payable on completion, with a
surcharge on additional properties. Rates change; use the
<a href="/mortgages/stamp-duty.html">calculator</a> and check gov.uk.</p>

<p><strong>Equity release / lifetime mortgage</strong> &mdash; borrowing against your
home in later life with interest usually rolled up rather than paid monthly.
Compounding runs against you; see the
<a href="/mortgages/equity-release.html">cost calculator</a>.</p>

<p><strong>Bridging loan</strong> &mdash; short-term borrowing priced monthly, to
cover a gap. Fast and expensive.</p>
</dl>

<h2>Loan and credit-specific</h2>

<dl>
<p><strong>Secured vs unsecured</strong> &mdash; whether the debt is attached to an
asset. Determines the rate, and determines what happens if you cannot pay. See
<a href="/guides/secured-vs-unsecured-what-changes.html">what changes</a>.</p>

<p><strong>Early settlement / the 58-day rule</strong> &mdash; you have a statutory
right to settle a regulated agreement early. The lender may add up to 28 days'
interest, and up to 30 more on agreements over twelve months. Hence "58 days".</p>

<p><strong>PCP</strong> &mdash; Personal Contract Purchase. You pay the depreciation
monthly and face an optional final balloon payment. Lower monthly, higher total, you
own nothing unless you pay the balloon.</p>

<p><strong>HP</strong> &mdash; Hire Purchase. You pay the whole cost over the term and
own the car at the end. Higher monthly, lower total.</p>

<p><strong>GFV</strong> &mdash; Guaranteed Future Value. The balloon figure on a PCP,
set at the outset.</p>

<p><strong>Voluntary termination</strong> &mdash; once you have paid <strong>50%</strong>
of the total amount payable on a regulated HP or PCP, you may hand the car back and
owe nothing further. A statutory right, and not the same as voluntary
<em>surrender</em>, where the debt remains.</p>

<p><strong>Fair wear and tear</strong> &mdash; the BVRLA standard used to assess a
returned vehicle. Published and specific; check against it before an inspector
does.</p>

<p><strong>Section 75</strong> &mdash; Consumer Credit Act protection making the card
provider jointly liable with the retailer for purchases between
<strong>£100 and £30,000</strong>. Credit cards, not debit cards.</p>

<p><strong>Right to withdraw</strong> &mdash; 14 days to withdraw from most regulated
credit agreements after signing. You repay what you borrowed plus any interest
accrued.</p>

<p><strong>Hard vs soft search</strong> &mdash; a hard search is recorded on your file
and visible to other lenders; a soft search is not. Eligibility checkers are
soft.</p>

<p><strong>Utilisation</strong> &mdash; how much of your available credit you are
using. Lower is better, and it updates monthly.</p>

<p><strong>Default</strong> &mdash; a formal record that an agreement broke down.
Stays on your file six years from the date recorded.</p>

<p><strong>CCJ</strong> &mdash; County Court Judgment. Six years on your file. If
unpaid, can lead to a <strong>charging order</strong>, which secures a previously
unsecured debt against your home.</p>

<p><strong>Breathing Space</strong> &mdash; the Debt Respite Scheme. Up to 60 days
with interest, fees and enforcement paused while you get advice. Applied for through
a debt adviser.</p>

<p><strong>DMP</strong> &mdash; Debt Management Plan. An informal arrangement to pay
non-priority creditors at a reduced rate, usually via a free charity.</p>

<p><strong>IVA</strong> &mdash; Individual Voluntary Arrangement. A formal, binding
insolvency arrangement. Serious consequences and not for everyone; take free advice
before considering one.</p>
</dl>

<p>Anything here you would explain differently, or a term we have missed? These pages
get revised, and the vocabulary is the part most worth getting right.</p>
"""
