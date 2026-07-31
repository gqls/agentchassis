#!/usr/bin/env python3
"""build_site.py — port 24 working calculators onto loanandmortgagecalculator.co.uk.

WHY A BUILDER AND NOT 24 HAND EDITS. The calculators do correct arithmetic and
have been live for months. The one thing that must not happen is a subtle change
to their logic, and the way that happens is a human editing 24 files and getting
23 of them right. So the transformation is mechanical and its safety property is
ASSERTED, not inspected:

    every inline <script> block in the output is byte-identical to the source.

If that assertion fails the build stops. Only <head>, the nav, the footer and
internal link targets are rewritten — never the body's controls or its logic.

Sources (both verified live, byte-for-byte, before porting):
    mortgage  ~/projects/domains/mortgagecalculator.co.uk/gemini/02/   (NOT the
              top-level dir — gemini/02 is what the live site actually serves)
    loan      ~/projects/sites/loancalculator.co.uk/tools/             (another
              lane owns this site; these files are READ, never written)

Run:  python3 build_site.py            # build
      python3 build_site.py --check    # assert only, write nothing
"""
import html
import os
import re
import sys

HOME = os.path.expanduser("~")
MORT_SRC = f"{HOME}/projects/domains/mortgagecalculator.co.uk/gemini/02"
LOAN_SRC = f"{HOME}/projects/sites/loancalculator.co.uk/tools"
OUT = f"{HOME}/projects/sites/loanandmortgagecalculator.co.uk"
DOMAIN = "loanandmortgagecalculator.co.uk"
BASE = f"https://{DOMAIN}"
BRAND = "LoanAndMortgageCalculator.co.uk"

CHECK_ONLY = "--check" in sys.argv

# ── the tool table ────────────────────────────────────────────────────────────
# title/desc are REWRITTEN for this site's audience (someone whose unsecured
# borrowing and their mortgage interact) — they are not the sources' metadata.
# `blurb` is what the section hub shows.
MORTGAGE = [
    ("simple", "Quick Mortgage Payment Check",
     "A 10-second monthly payment estimate. Enter the amount, rate and term — no sign-up, no credit check, no email.",
     "The fastest way to sanity-check a monthly payment before you look at anything else."),
    ("repayment", "Full Repayment and Amortisation Calculator",
     "See the monthly payment, the total interest over the full term, and a year-by-year breakdown of how much of your balance actually goes down.",
     "The whole term, year by year — including how little of year one touches the capital."),
    ("affordability", "How Much Could You Borrow?",
     "An affordability estimate based on income and outgoings. Existing loan and card payments reduce what a lender will offer, and this shows by how much.",
     "Income in, borrowing estimate out — and the effect of the debts you already have."),
    ("stamp-duty", "Stamp Duty (SDLT) Calculator",
     "Work out the Stamp Duty Land Tax due on a purchase, including the additional-property surcharge for a second home or buy-to-let.",
     "The tax bill on completion day, including the second-property surcharge."),
    ("overpayment", "Mortgage Overpayment Calculator",
     "What a regular monthly overpayment does to your term and your total interest. Compare it against clearing an unsecured loan first.",
     "How many years and how much interest a monthly overpayment actually removes."),
    ("investor", "Rental Yield and LTV Calculator",
     "Gross and net yield, loan-to-value and monthly cashflow for a single rental property.",
     "Yield, LTV and cashflow for one property, on one screen."),
    ("portfolio", "Property Portfolio Calculator",
     "Aggregate LTV, yield and monthly cashflow across several properties at once. Saves in your browser, so nothing is sent anywhere.",
     "Several properties at once, with the totals that lenders look at."),
    ("equity-release", "Equity Release Cost Calculator",
     "What a lifetime mortgage costs over time once interest is rolled up rather than paid monthly — the figure that surprises people.",
     "Roll-up interest is compound interest running the other way. See the size of it."),
    ("bridging-loan", "Bridging Loan Cost Calculator",
     "The real cost of short-term bridging finance once monthly interest, arrangement fees and exit fees are all counted.",
     "Short-term borrowing priced monthly — what speed actually costs."),
    ("fee-analyser", "Mortgage Fee Analyser",
     "Compare two mortgage deals on total cost over the fixed period rather than on the headline rate. A lower rate with a big product fee often loses.",
     "The cheapest rate and the cheapest deal are frequently not the same product."),
    ("rate-forecaster", "Payment Change Forecaster",
     "What your payment becomes at higher rates when your fixed period ends, so the end of the fix is not a surprise.",
     "Model the remortgage cliff before you reach it."),
    ("fact-finder", "Mortgage Approval Fact Finder",
     "A structured self-assessment of the things a lender actually checks, and where an application is most likely to come unstuck.",
     "What an underwriter looks at, in the order they look at it."),
]

LOAN = [
    ("standard-calc", "Loan Repayment Calculator",
     "Monthly repayment and total cost of credit for a personal loan. The total interest figure is the one worth comparing between offers.",
     "Monthly payment and the total cost of credit — the number the advert leaves out."),
    ("compare-loans", "Compare Two Loans Side by Side",
     "Put two loan offers next to each other and compare total repayable, not the APR headline. Term length usually matters more than rate.",
     "Two offers, side by side, compared on what you actually repay."),
    ("consolidation", "Debt Consolidation Calculator",
     "Add up what you owe now and see what consolidating would cost — including the case where a lower monthly payment costs you more overall.",
     "A lower monthly payment and a lower total cost are different things."),
    ("car-finance-calculator", "PCP vs HP Car Finance Calculator",
     "Compare Personal Contract Purchase against Hire Purchase on total cost, and see the balloon payment PCP defers to the end.",
     "PCP against HP on total cost, with the final balloon payment in plain sight."),
    ("credit-health-check", "Credit Health Check",
     "A structured look at the factors that move a credit file, and which of them a lender weighs most heavily.",
     "What is actually on your file, and which parts a lender cares about."),
    # NOT PORTED — credit-roadmap.html. 1,816 bytes, zero controls, zero script:
    # it is a static read filed as a tool, and the browser audit called it
    # NO-CONTROL on the live source for that reason, correctly. Its subject is
    # covered properly (and in more depth) by the six-month plan in
    # /guides/credit-file-before-a-mortgage.html, so shipping it here would add
    # a page that fails its own section's promise. Recorded rather than dropped
    # silently, per the no-silent-caps rule.
    ("damage-checker", "Car Finance Return Damage Checker",
     "Check likely end-of-contract damage charges against the BVRLA fair wear and tear standard before you hand a car back.",
     "What counts as fair wear and tear, and what gets charged for."),
    ("interest-rate-stress-test", "Interest Rate Stress Test",
     "What a rate rise does to your repayments. Run it on every variable-rate debt you hold, not just the largest one.",
     "Rates go up. Find out what that does to your monthly total first."),
    ("loan-vs-savings", "Clear the Loan or Keep the Savings?",
     "Compare the interest a loan costs you against the interest your savings earn, and see which way round leaves you better off.",
     "Borrowing costs more than saving earns. This shows by how much."),
    ("overpayment-calculator", "Loan Overpayment Calculator",
     "What overpaying an unsecured loan saves in interest and time — usually a much higher return than overpaying a mortgage.",
     "Unsecured debt is the expensive debt. Overpay it first, and this shows why."),
    ("settlement-calculator", "Early Settlement Calculator",
     "Estimate an early settlement figure including the 58-day interest rule lenders are allowed to apply under the Consumer Credit Act.",
     "Settling early costs less than the remaining payments — but not by as much as you would think."),
    ("application-tracker", "Application Tracker",
     "Keep track of borrowing applications and the searches they leave on your credit file. Saved in your browser only.",
     "Several applications in a short window is itself a signal to lenders. Track them."),
]

# ── guide set (all new; every one about where the two subjects cross) ─────────
GUIDES = [
    ("how-loans-affect-mortgage-affordability",
     "How Your Loans Cut What You Can Borrow",
     "Every £100 a month of loan or card payments reduces your mortgage offer by roughly £5,000-£7,000. Here is the arithmetic lenders actually use, and what to do about it."),
    ("consolidating-debt-into-your-mortgage",
     "Consolidating Debt Into Your Mortgage",
     "Moving unsecured debt onto your mortgage cuts the monthly payment and usually raises the lifetime cost. It also turns debt you could not lose your home over into debt you can."),
    ("total-cost-of-borrowing",
     "Total Cost, Not Monthly Payment",
     "Why comparing borrowing on the monthly figure is the single most expensive habit in personal finance, and what to compare instead."),
    ("deposit-or-clear-the-debt",
     "Deposit, or Clear the Debt First?",
     "You have a lump sum and both a loan and a house deposit to think about. Which pound works harder depends on three things, and one of them is not the interest rate."),
    ("credit-file-before-a-mortgage",
     "Your Credit File Before a Mortgage Application",
     "What a mortgage underwriter reads that a loan underwriter does not, why timing your applications matters, and the searches that quietly count against you."),
    ("secured-vs-unsecured-what-changes",
     "Secured or Unsecured: What Actually Changes",
     "The rate difference is the obvious part. The part that matters more is what the lender can do if you stop paying."),
    ("car-finance-and-your-mortgage",
     "Car Finance and Your Mortgage Application",
     "PCP and HP are credit commitments, and a mortgage lender counts them in full. What that means if you are buying a car and a house in the same year."),
    ("remortgaging-with-other-debt",
     "Remortgaging When You Have Other Debt",
     "Remortgaging is where your unsecured borrowing and your mortgage meet, and where the biggest avoidable mistakes get made."),
    ("fixed-vs-variable-on-both",
     "Fixed or Variable, on Both Sides",
     "Most people fix the mortgage and leave everything else on a variable rate. That is often exactly the wrong way round."),
    ("stress-testing-the-whole-budget",
     "Stress-Testing the Whole Budget",
     "Lenders stress-test your mortgage against a higher rate. Almost nobody stress-tests the rest of their borrowing at the same time."),
    ("the-fees-nobody-quotes",
     "The Fees Nobody Quotes You",
     "Product fees, valuation fees, arrangement fees, early repayment charges, settlement interest. Where they hide and which ones are negotiable."),
    ("when-repayments-are-a-struggle",
     "When Repayments Become a Struggle",
     "What to do first, in what order, and the free UK services that will act on your behalf. Priority debts come before everything else."),
    ("jargon-buster",
     "Loan and Mortgage Jargon, Translated",
     "APR, APRC, LTV, LTI, ERC, SVC, roll-up, balloon, 58-day rule. Both vocabularies in one place, in plain English."),
]

# old guide slug → new guide slug (the sources link at guides I am not copying)
GUIDE_MAP = {
    # loancalculator's guides
    "how-loans-are-calculated": "total-cost-of-borrowing",
    "can-i-overpay": "total-cost-of-borrowing",
    "debt-consolidation-explained": "consolidating-debt-into-your-mortgage",
    "loan-eligibility-uk": "credit-file-before-a-mortgage",
    "secured-vs-unsecured": "secured-vs-unsecured-what-changes",
    "fixed-vs-variable-loans": "fixed-vs-variable-on-both",
    "hidden-loan-fees": "the-fees-nobody-quotes",
    "car-finance-explained": "car-finance-and-your-mortgage",
    "finance-damage-and-insurance": "car-finance-and-your-mortgage",
    "uk-lending-landscape": "total-cost-of-borrowing",
    "jargon-buster": "jargon-buster",
    "document-checklist": "credit-file-before-a-mortgage",
    "debt-help-uk": "when-repayments-are-a-struggle",
    # mortgagecalculator's guides
    "first-time-buyer": "deposit-or-clear-the-debt",
    "how-banks-decide": "how-loans-affect-mortgage-affordability",
    "lender-restrictions": "how-loans-affect-mortgage-affordability",
    "missed-payments": "when-repayments-are-a-struggle",
    "remortgaging": "remortgaging-with-other-debt",
    "your-mortgage-scorecard": "credit-file-before-a-mortgage",
    "mortgage-scorecard": "credit-file-before-a-mortgage",  # the dead link, S2
    "market-structure": "total-cost-of-borrowing",
    "negative-equity": "secured-vs-unsecured-what-changes",
    "buy-to-let": "total-cost-of-borrowing",
}

MORT_SLUGS = {s for s, *_ in MORTGAGE}
LOAN_SLUGS = {s for s, *_ in LOAN}


# ── shared chrome ────────────────────────────────────────────────────────────
def head(title, desc, canonical, extra=""):
    return f"""<!DOCTYPE html>
<html lang="en-GB">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{html.escape(title)} | {BRAND}</title>
<meta name="description" content="{html.escape(desc, quote=True)}">
<link rel="canonical" href="{canonical}">
<meta property="og:title" content="{html.escape(title)}">
<meta property="og:description" content="{html.escape(desc, quote=True)}">
<meta property="og:url" content="{canonical}">
<meta property="og:type" content="website">
<meta property="og:site_name" content="{BRAND}">
<meta name="twitter:card" content="summary">
<meta name="theme-color" content="#0A1222">
<link rel="stylesheet" href="/assets/css/style.css">
<link rel="icon" href="/favicon.ico" sizes="any">
<link rel="icon" href="/favicon.svg" type="image/svg+xml">
<link rel="apple-touch-icon" href="/apple-touch-icon.png">
{extra}</head>
"""


def nav(active=""):
    def link(href, label, key):
        cur = ' aria-current="page"' if key == active else ""
        return f'<a href="{href}"{cur}>{label}</a>'
    return f"""<a class="skip-link" href="#content">Skip to content</a>
<header>
<nav class="site-nav">
<div class="brand"><a href="/"><img src="/assets/img/logo.svg" alt="" width="34" height="34" aria-hidden="true">LoanAndMortgage<span class="brand-suffix">Calculator.co.uk</span></a></div>
<button id="mobile-menu-btn" aria-expanded="false" aria-controls="nav-links-menu" aria-label="Open menu">&#9776;</button>
<div class="nav-links" id="nav-links-menu">
{link("/mortgages/", "Mortgage tools", "mortgages")}
{link("/loans/", "Loan tools", "loans")}
{link("/guides/", "Guides", "guides")}
</div>
</nav>
</header>
"""


FOOTER = f"""<footer>
<div class="footer-inner">
<div>
<h4>Mortgage tools</h4>
<ul>
<li><a href="/mortgages/repayment.html">Repayment &amp; amortisation</a></li>
<li><a href="/mortgages/affordability.html">Affordability</a></li>
<li><a href="/mortgages/stamp-duty.html">Stamp Duty</a></li>
<li><a href="/mortgages/">All mortgage tools</a></li>
</ul>
</div>
<div>
<h4>Loan tools</h4>
<ul>
<li><a href="/loans/standard-calc.html">Loan repayments</a></li>
<li><a href="/loans/consolidation.html">Debt consolidation</a></li>
<li><a href="/loans/car-finance-calculator.html">PCP vs HP</a></li>
<li><a href="/loans/">All loan tools</a></li>
</ul>
</div>
<div>
<h4>Guides</h4>
<ul>
<li><a href="/guides/how-loans-affect-mortgage-affordability.html">Loans vs borrowing power</a></li>
<li><a href="/guides/consolidating-debt-into-your-mortgage.html">Consolidating into a mortgage</a></li>
<li><a href="/guides/when-repayments-are-a-struggle.html">If repayments are a struggle</a></li>
<li><a href="/guides/">All guides</a></li>
</ul>
</div>
<div>
<h4>About</h4>
<ul>
<li><a href="/legal.html">Legal &amp; privacy</a></li>
</ul>
</div>
</div>
<div class="footer-legal">
<p><strong>{BRAND}</strong> provides calculators and general information only. Nothing here is financial advice, and no result is an offer of credit. Figures are estimates: a lender's own assessment will differ. Always check current rates and rules before making a decision, and take regulated advice for anything that matters.</p>
<p>Free, impartial UK debt help: <a href="https://www.moneyhelper.org.uk">MoneyHelper</a>, <a href="https://www.citizensadvice.org.uk">Citizens Advice</a>, <a href="https://www.stepchange.org">StepChange</a>.</p>
</div>
</footer>
<script src="/assets/js/site.js"></script>
</body>
</html>
"""


# ── link rewriting ───────────────────────────────────────────────────────────
def rewrite_links(body, section):
    """Point every internal link at its new home. Never touches script bodies —
    callers pass only the markup between </header> and the first <script>."""
    def guide(m):
        slug = m.group(1)
        return f'href="/guides/{GUIDE_MAP.get(slug, "index")}.html"'
    # /guides/x.html and guides/x.html
    body = re.sub(r'href="/?guides/([a-z0-9-]+)\.html"', guide, body)
    # loan tools: /tools/x.html -> /loans/x.html
    body = re.sub(r'href="/tools/([a-z0-9-]+)\.html"',
                  lambda m: f'href="/loans/{m.group(1)}.html"'
                  if m.group(1) in LOAN_SLUGS else 'href="/loans/"', body)
    # home
    body = re.sub(r'href="/?index\.html"', 'href="/"', body)
    # bare siblings, e.g. href="affordability.html" on a mortgage page
    def sibling(m):
        slug = m.group(1)
        if slug in MORT_SLUGS:
            return f'href="/mortgages/{slug}.html"'
        if slug in LOAN_SLUGS:
            return f'href="/loans/{slug}.html"'
        return f'href="/{section}/"'
    body = re.sub(r'href="([a-z0-9-]+)\.html"', sibling, body)
    # assets that were relative on the mortgage side
    body = body.replace('src="images/', 'src="/assets/img/')
    body = body.replace('href="css/style.css"', 'href="/assets/css/style.css"')
    return body


INLINE_SCRIPT = re.compile(r"<script(?![^>]*\bsrc=)[^>]*>(.*?)</script>", re.S)


def inline_scripts(s):
    return INLINE_SCRIPT.findall(s)


def port(src_path, section, slug, title, desc):
    raw = open(src_path, encoding="utf-8").read()

    # body = everything after the site chrome, up to the first <script>
    if section == "mortgages":
        after = re.split(r"</header>", raw, maxsplit=1)
        body = after[1] if len(after) > 1 else raw
    else:
        body = re.split(r'<div id="nav-placeholder"></div>', raw, maxsplit=1)[-1]

    cut = re.search(r"<script", body)
    tail = body[cut.start():] if cut else ""
    body = body[:cut.start()] if cut else body
    body = rewrite_links(body, section)

    # ── external dependencies ───────────────────────────────────────────────
    # Scanned from the WHOLE source, not just the body. bridging-loan,
    # equity-release and fee-analyser put <script src="js/calculators.js"> in
    # the <head>, and this builder rebuilds the head from scratch — so the first
    # version of it silently dropped their only dependency. bridging-loan threw
    # `formatGBP is not defined`; the other two carried on LOOKING fine, which
    # is why this is asserted below rather than eyeballed.
    #
    # Emitted immediately before the inline scripts, which preserves the only
    # ordering that matters: the helpers load before the code that calls them.
    deps, drops = [], ("cloudflareinsights", "nav.js")
    for m in re.finditer(r'<script[^>]*\bsrc="([^"]+)"[^>]*>\s*</script>', raw):
        src = m.group(1)
        if any(d in src for d in drops) or any(d in m.group(0) for d in drops):
            continue            # S7 analytics token; B1 nav is static now
        if src in ("js/calculators.js", "/assets/js/calculators.js"):
            src = "/assets/js/calculators.js"
        if src not in deps:
            deps.append(src)

    keep = [f'<script src="{d}"></script>' for d in deps]
    for m in INLINE_SCRIPT.finditer(tail):
        keep.append(m.group(0))

    canonical = f"{BASE}/{section}/{slug}.html"
    out = (head(title, desc, canonical) + "<body>\n" + nav(section)
           + '<div id="content">\n' + body.strip() + "\n</div>\n"
           + "\n".join(keep) + "\n" + FOOTER)

    # ── safety property 1: the logic is untouched ───────────────────────────
    want, got = inline_scripts(raw), inline_scripts(out)
    want = [w for w in want if "cloudflareinsights" not in w]
    if want != got:
        raise SystemExit(
            f"ABORT {section}/{slug}: inline script blocks changed during the port.\n"
            f"  source blocks: {len(want)}  output blocks: {len(got)}\n"
            f"  This is the one thing the build must never do.")

    # ── safety property 2: nothing the logic DEPENDS ON went missing ────────
    # Property 1 passed on all three broken pages, because their inline blocks
    # were byte-identical — what vanished was the file those blocks call into.
    # A byte-identical script with a missing dependency is still a broken page.
    for d in deps:
        if f'src="{d}"' not in out:
            raise SystemExit(f"ABORT {section}/{slug}: dependency {d} lost in the port.")
    return out


def write(rel, content):
    path = os.path.join(OUT, rel)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    if not CHECK_ONLY:
        with open(path, "w", encoding="utf-8") as f:
            f.write(content)
    return path


def main():
    built = []
    for slug, title, desc, _ in MORTGAGE:
        out = port(f"{MORT_SRC}/{slug}.html", "mortgages", slug, title, desc)
        write(f"mortgages/{slug}.html", out)
        built.append(("mortgages", slug, len(out)))
    for slug, title, desc, _ in LOAN:
        out = port(f"{LOAN_SRC}/{slug}.html", "loans", slug, title, desc)
        write(f"loans/{slug}.html", out)
        built.append(("loans", slug, len(out)))

    for section, slug, n in built:
        print(f"  {'OK ' if not CHECK_ONLY else 'CHK'} {section}/{slug}.html  {n:>6,} bytes")
    print(f"\n{len(built)} calculators ported; "
          f"every inline script block byte-identical to source.")


if __name__ == "__main__":
    main()
