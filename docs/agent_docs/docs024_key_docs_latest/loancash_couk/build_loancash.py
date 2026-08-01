#!/usr/bin/env python3
"""build_loancash.py — loancash.co.uk: the borrower's guide to the FCA rulebook.

REGISTER ENTRY L10 (portfolio_positioning/REGISTER_positioning.md). The direction, in one
line: the name "loan cash" attracts the most vulnerable borrowing query there is, and the
owner's ruling (P6, 2026-08-01) is to make that landing a PROTECTION — the visitor gets
their rights under the FCA's consumer-credit rules, not a lender.

HARD CONSTRAINTS, from the register row:
  * NOT a lender, NOT a broker. No applications, no lead-gen, no "apply now" anywhere.
  * INDEPENDENT of the FCA, and the chrome says so on every page — a site championing the
    regulator's rules must never read as the regulator.
  * Regulatory constants (0.8%/day, £15, 100%, 2 rollovers, 2 CPA attempts, 8 weeks,
    6 months) are the ONE sanctioned exception to the house no-quoted-numbers rule: they
    are rules, not market rates, and each is quoted WITH the rule it comes from and a
    check-the-source pointer. Market rates are still never quoted.
  * If the question is "which loan should I get" -> link out (whichloan territory, L2).
    If it is "what are they allowed to do to me" -> it lives here.

BUILD SAFETY (inherited from the loanandmortgagecalculator builders, where each bit once):
  every file funnels through write(), which refuses (3) any reference naming a directory
  — an object store cannot resolve a directory index, so "/tools/" is a live 404 — and
  (4) any ld+json block that does not parse. Counts in copy are DERIVED, never typed.

Run:  python3 build_loancash.py            # build
      python3 build_loancash.py --check    # assert only, write nothing
"""
import html
import json
import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from content_loancash import GUIDES, BODIES  # noqa: E402

OUT = os.path.expanduser("~/projects/sites/loancash.co.uk")
DOMAIN = "loancash.co.uk"
BASE = f"https://{DOMAIN}"
BRAND = "LoanCash.co.uk"
TODAY = "2026-08-01"
CHECK_ONLY = "--check" in sys.argv


def hub(section):
    # An object store cannot resolve a directory index (bugs_open/116) — every internal
    # reference names the file. Same rule, same reason, same near-miss as the sibling.
    return f"/{section}/index.html"


TOOLS = [
    ("price-cap-checker", "Price Cap Checker",
     "Enter what a short-term lender charged you and check it against the FCA's price cap: "
     "0.8% per day, a £15 default fee limit, and a 100% total cost cap.",
     "Were you charged more than the law allows? Check in ten seconds."),
    ("true-cost-calculator", "True Cost of a Short-Term Loan",
     "See what a payday or short-term loan really costs at the legal maximum rate, what the "
     "total cost cap means for you, and what a credit union could charge instead.",
     "The full cost at the legal maximum — and the cheaper route most people never check."),
    ("complaint-deadline-calculator", "Complaint Deadline Calculator",
     "Complained to a lender? Work out the exact date their 8 weeks runs out and your "
     "6-month window to go to the Financial Ombudsman for free.",
     "Two dates decide your complaint. Know both before the lender does."),
]

N_TOOLS, N_GUIDES = len(TOOLS), len(GUIDES)


# ── shared chrome ─────────────────────────────────────────────────────────────
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
<meta name="theme-color" content="#0B3D2E">
<link rel="stylesheet" href="/assets/css/style.css">
<link rel="icon" href="/favicon.svg" type="image/svg+xml">
{extra}</head>
"""


def nav(active=""):
    def link(href, label, key):
        cur = ' aria-current="page"' if key == active else ""
        return f'<a href="{href}"{cur}>{label}</a>'
    return f"""<a class="skip-link" href="#content">Skip to content</a>
<header>
<nav class="site-nav">
<div class="brand"><a href="/">Loan<span class="brand-accent">Cash</span>.co.uk</a>
<span class="brand-tag">Know the rules before you borrow</span></div>
<button id="mobile-menu-btn" aria-expanded="false" aria-controls="nav-links-menu" aria-label="Open menu">&#9776;</button>
<div class="nav-links" id="nav-links-menu">
{link(hub("tools"), "Check your loan", "tools")}
{link(hub("guides"), "Your rights", "guides")}
{link("/guides/if-you-cant-pay.html", "Free help now", "help")}
</div>
</nav>
</header>
"""


FOOTER = f"""<footer>
<div class="footer-inner">
<div>
<h4>Check your loan</h4>
<ul>
<li><a href="/tools/price-cap-checker.html">Price cap checker</a></li>
<li><a href="/tools/true-cost-calculator.html">True cost calculator</a></li>
<li><a href="/tools/complaint-deadline-calculator.html">Complaint deadlines</a></li>
</ul>
</div>
<div>
<h4>Your rights</h4>
<ul>
<li><a href="/guides/the-payday-loan-price-cap.html">The price cap</a></li>
<li><a href="/guides/how-to-complain-and-win.html">How to complain</a></li>
<li><a href="/guides/stopping-payments-the-cpa-rules.html">Stopping payments</a></li>
<li><a href="{hub('guides')}">All guides</a></li>
</ul>
</div>
<div>
<h4>Free help, right now</h4>
<ul>
<li><a href="https://www.moneyhelper.org.uk" rel="noopener">MoneyHelper</a></li>
<li><a href="https://www.stepchange.org" rel="noopener">StepChange</a></li>
<li><a href="https://www.citizensadvice.org.uk" rel="noopener">Citizens Advice</a></li>
<li><a href="https://nationaldebtline.org" rel="noopener">National Debtline</a></li>
</ul>
</div>
<div>
<h4>About</h4>
<ul>
<li><a href="/legal.html">Who we are &amp; legal</a></li>
</ul>
</div>
</div>
<div class="footer-legal">
<p><strong>{BRAND} does not lend money, broker loans, or take applications — and never
will.</strong> We publish plain-English information about the rules that protect UK
borrowers. Nothing here is financial or legal advice.</p>
<p><strong>We are independent. We are not the Financial Conduct Authority and are not
affiliated with it.</strong> We think its consumer-credit rules deserve to be better
known — that is the whole point of this site. The rules themselves live at
<a href="https://www.fca.org.uk" rel="noopener">fca.org.uk</a>; always check the source
for the current position.</p>
</div>
</footer>
<script src="/assets/js/site.js"></script>
</body>
</html>
"""


def page(title, desc, canonical, active, body, extra="", tight=False):
    cls = "container container-tight" if tight else "container"
    return (head(title, desc, canonical, extra) + "<body>\n" + nav(active)
            + f'<div id="content" class="{cls}">\n' + body + "\n</div>\n" + FOOTER)


# ── write(), with the two assertions that have already caught real defects ───
DIR_REF = re.compile(r'(?:href|src)="(/[^"]*/|[a-z0-9][^":]*/)"')
LD_JSON = re.compile(r'<script type="application/ld\+json">(.*?)</script>', re.S)


def write(rel, content):
    for m in DIR_REF.finditer(content):
        raise SystemExit(
            f'ABORT {rel}: reference "{m.group(1)}" names a directory, not a file.\n'
            f"  An object store cannot resolve a directory index — this is a live 404.")
    for m in LD_JSON.finditer(content):
        try:
            json.loads(m.group(1))
        except ValueError as e:
            raise SystemExit(f"ABORT {rel}: ld+json does not parse ({e}).")
    path = os.path.join(OUT, rel)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    if not CHECK_ONLY:
        with open(path, "w", encoding="utf-8") as f:
            f.write(content)
    return path


# ── home ──────────────────────────────────────────────────────────────────────
HOME_BODY = f"""<div class="hero">
<p class="eyebrow">The borrower's guide to the FCA rulebook</p>
<h1>Quick cash has rules. The lender knows them. Now you do too.</h1>
<p>If you are looking at a payday loan, doorstep loan, rent-to-own deal or any other
high-cost credit: the law strictly limits what you can be charged, what a lender must
check, and what they may do if you struggle. This site exists to put those rules in your
hands — {N_TOOLS} checkers and {N_GUIDES} plain-English guides. <strong>We do not lend
money and never will.</strong></p>
</div>

<div class="rule-box">
<h3 class="mt-0">The three numbers every short-term borrower should know</h3>
<p><strong>0.8% per day</strong> — the most a payday-style lender may charge in interest
and fees while you repay on time. <strong>&pound;15</strong> — the most they may charge
in default fees if you miss a payment. <strong>100%</strong> — the total cost cap: you can
never be made to pay back more in interest and fees than the amount you borrowed. All
three are FCA rules (CONC 5A), in force since January 2015.</p>
<p><a href="/tools/price-cap-checker.html">Check what you were charged against the cap
&rarr;</a></p>
</div>

<h2>Check your loan</h2>
<div class="tool-grid">
""" + "\n".join(
    f'<div class="card"><h3><a href="/tools/{s}.html">{html.escape(t)}</a></h3>'
    f'<p>{html.escape(b)}</p>'
    f'<a class="btn-primary btn-block" href="/tools/{s}.html">Open the checker</a></div>'
    for s, t, d, b in TOOLS) + f"""
</div>

<h2>Know your rights</h2>
<p>Each guide covers one protection, what it means in practice, and exactly what to do
when a lender falls short of it.</p>
<div class="tool-grid">
""" + "\n".join(
    f'<div class="card"><h3><a href="/guides/{s}.html">{html.escape(t)}</a></h3>'
    f'<p>{html.escape(d)}</p></div>' for s, t, d in GUIDES[:6]) + f"""
</div>
<p class="text-center mt-40"><a class="btn-primary" href="{hub('guides')}">All {N_GUIDES} guides</a></p>

<div class="help-box">
<h3 class="mt-0">Struggling right now?</h3>
<p>Free, confidential debt help exists and it works: <a href="https://www.moneyhelper.org.uk"
rel="noopener">MoneyHelper</a>, <a href="https://www.stepchange.org" rel="noopener">StepChange</a>,
<a href="https://www.citizensadvice.org.uk" rel="noopener">Citizens Advice</a> and
<a href="https://nationaldebtline.org" rel="noopener">National Debtline</a> charge you
nothing, ever. <a href="/guides/if-you-cant-pay.html">What the rules say when you can't
pay &rarr;</a></p>
</div>
"""

HOME_LD = ('<script type="application/ld+json">\n'
           + json.dumps({
               "@context": "https://schema.org", "@type": "WebSite",
               "name": BRAND, "url": f"{BASE}/",
               "description": "Plain-English guides and checkers for the FCA rules that "
                              "protect UK borrowers from unfair high-cost credit."},
               ensure_ascii=False)
           + "\n</script>\n")

write("index.html", page(
    "The Rules That Protect UK Borrowers",
    f"{N_TOOLS} free checkers and {N_GUIDES} plain-English guides to the FCA rules on "
    "payday and high-cost loans: the price cap, your complaint rights, and how to stop "
    "unfair charges. Not a lender.",
    f"{BASE}/", "", HOME_BODY, extra=HOME_LD))

# ── tools hub ─────────────────────────────────────────────────────────────────
write("tools/index.html", page(
    "Check Your Loan Against the Rules",
    f"{N_TOOLS} free checkers: was your loan within the FCA price cap, what a short-term "
    "loan really costs, and the exact deadlines on your complaint.",
    f"{BASE}{hub('tools')}", "tools",
    f"""<p class="breadcrumb"><a href="/">Home</a><span>&rsaquo;</span>Check your loan</p>
<h1>Check your loan against the rules</h1>
<p class="subtitle">Everything runs in your browser. Nothing you type is sent anywhere,
stored anywhere, or seen by anyone — including us.</p>
<div class="tool-grid">
""" + "\n".join(
        f'<div class="card"><h3><a href="/tools/{s}.html">{html.escape(t)}</a></h3>'
        f'<p>{html.escape(d)}</p>'
        f'<a class="btn-primary btn-block" href="/tools/{s}.html">Open the checker</a></div>'
        for s, t, d, b in TOOLS) + """
</div>"""))

# ── the three tools ───────────────────────────────────────────────────────────
CAP_TOOL = """<p class="breadcrumb"><a href="/">Home</a><span>&rsaquo;</span><a href="/tools/index.html">Check your loan</a><span>&rsaquo;</span>Price cap checker</p>
<h1>Price cap checker</h1>
<p class="subtitle">For payday-style loans (high-cost short-term credit) taken since
2&nbsp;January&nbsp;2015. Enter what happened; get the legal maximums and a verdict.</p>
<div class="tool-panel">
<label for="pc-amount">Amount you borrowed (&pound;)</label>
<input type="number" id="pc-amount" min="1" step="1" placeholder="300">
<label for="pc-days">Days from taking the loan to your final payment</label>
<input type="number" id="pc-days" min="1" step="1" placeholder="30">
<label for="pc-charged">Total interest and fees you paid or were asked to pay (&pound;)</label>
<input type="number" id="pc-charged" min="0" step="0.01" placeholder="75">
<label for="pc-default">Of that, default fees for missing a payment (&pound;, 0 if none)</label>
<input type="number" id="pc-default" min="0" step="0.01" value="0">
<button id="pc-go" class="btn-primary">Check against the cap</button>
<div id="pc-result" class="result" aria-live="polite"></div>
</div>
<article class="guide-content">
<h2>What this checks</h2>
<p>Three FCA limits (rule CONC 5A) apply to high-cost short-term credit: interest and fees
while you repay on time may not exceed <strong>0.8% of the amount borrowed per day</strong>;
default fees are capped at <strong>&pound;15 in total</strong> (default interest stays
inside the 0.8%/day limit); and the <strong>total cost cap</strong> means interest and fees
can never exceed <strong>100% of what you borrowed</strong>, no matter what.</p>
<p>If your result says the cap was breached, that money is claimable.
<a href="/guides/how-to-complain-and-win.html">Complain to the lender first, then the
Ombudsman — the guide walks you through it</a>. If the lender was not FCA-authorised at
all, stop — <a href="/guides/loan-sharks-and-illegal-lending.html">different, stronger
rules apply</a>.</p>
<p><em>This checker covers high-cost short-term credit — typically loans due within 12
months at very high rates. Other products have their own rules:
<a href="/guides/types-of-high-cost-credit-and-their-rules.html">see which rules cover
your loan</a>.</em></p>
</article>
<script>
(function () {
  var gbp = new Intl.NumberFormat('en-GB', { style: 'currency', currency: 'GBP' });
  document.getElementById('pc-go').addEventListener('click', function () {
    var amount = parseFloat(document.getElementById('pc-amount').value);
    var days = parseInt(document.getElementById('pc-days').value, 10);
    var charged = parseFloat(document.getElementById('pc-charged').value);
    var dflt = parseFloat(document.getElementById('pc-default').value) || 0;
    var out = document.getElementById('pc-result');
    if (!(amount > 0) || !(days > 0) || !(charged >= 0)) {
      out.innerHTML = '<p class="warn">Fill in the amount, the days and the total charged.</p>';
      return;
    }
    var maxInitial = amount * 0.008 * days;
    var maxTotal = amount;              // the 100% total cost cap
    var lines = [];
    lines.push('<p>Maximum interest + fees at 0.8%/day for ' + days + ' days: <strong>'
      + gbp.format(Math.min(maxInitial, maxTotal)) + '</strong></p>');
    lines.push('<p>Absolute ceiling (100% total cost cap): <strong>' + gbp.format(maxTotal)
      + '</strong></p>');
    var breaches = [];
    if (dflt > 15) breaches.push('default fees of ' + gbp.format(dflt)
      + ' exceed the \\u00a315 cap');
    if (charged > maxTotal + 0.005) breaches.push('total charges of ' + gbp.format(charged)
      + ' exceed the 100% total cost cap of ' + gbp.format(maxTotal));
    else if (charged > maxInitial + Math.min(dflt, 15) + 0.005)
      breaches.push('total charges of ' + gbp.format(charged)
        + ' exceed the 0.8%/day limit for your term ('
        + gbp.format(maxInitial + Math.min(dflt, 15)) + ' incl. capped default fees)');
    if (breaches.length) {
      lines.push('<p class="verdict-bad"><strong>The cap looks breached:</strong> '
        + breaches.join('; ') + '. That money is claimable \\u2014 '
        + '<a href="/guides/how-to-complain-and-win.html">here is exactly how</a>.</p>');
    } else {
      lines.push('<p class="verdict-ok"><strong>Within the cap</strong> on these numbers. '
        + 'If the loan still felt unaffordable, that is a separate right \\u2014 '
        + '<a href="/guides/affordability-checks-what-lenders-must-do.html">the lender had '
        + 'to check you could repay</a>.</p>');
    }
    out.innerHTML = lines.join('');
  });
})();
</script>"""

COST_TOOL = """<p class="breadcrumb"><a href="/">Home</a><span>&rsaquo;</span><a href="/tools/index.html">Check your loan</a><span>&rsaquo;</span>True cost calculator</p>
<h1>True cost of a short-term loan</h1>
<p class="subtitle">What a loan costs at the legal maximum rate &mdash; and what the same
money could cost from a credit union instead.</p>
<div class="tool-panel">
<label for="tc-amount">Amount (&pound;)</label>
<input type="number" id="tc-amount" min="1" step="1" placeholder="300">
<label for="tc-days">Days until fully repaid</label>
<input type="number" id="tc-days" min="1" step="1" placeholder="30">
<label for="tc-rate">Daily rate the lender quotes (%; the legal maximum is 0.8)</label>
<input type="number" id="tc-rate" min="0" max="0.8" step="0.01" value="0.8">
<button id="tc-go" class="btn-primary">Show the true cost</button>
<div id="tc-result" class="result" aria-live="polite"></div>
</div>
<article class="guide-content">
<h2>Why the daily rate hides the real price</h2>
<p>0.8% a day sounds small. Compounded over a year it is an APR in the thousands of
percent &mdash; which is why the law forces short-term lenders to show that APR in their
adverts, and why the total cost cap exists. The comparison to a credit union below is the
one most borrowers never see: <strong>credit unions may charge at most 3% a month</strong>
on a reducing balance, and lend small amounts to people banks refuse.
<a href="/guides/cheaper-ways-to-borrow-small-amounts.html">Where to find one, and the
other cheap routes &rarr;</a></p>
</article>
<script>
(function () {
  var gbp = new Intl.NumberFormat('en-GB', { style: 'currency', currency: 'GBP' });
  document.getElementById('tc-go').addEventListener('click', function () {
    var amount = parseFloat(document.getElementById('tc-amount').value);
    var days = parseInt(document.getElementById('tc-days').value, 10);
    var rate = parseFloat(document.getElementById('tc-rate').value);
    var out = document.getElementById('tc-result');
    if (!(amount > 0) || !(days > 0) || !(rate >= 0)) {
      out.innerHTML = '<p class="warn">Fill in all three boxes.</p>';
      return;
    }
    var capped = false;
    if (rate > 0.8) { rate = 0.8; capped = true; }
    var charge = amount * (rate / 100) * days;
    if (charge > amount) { charge = amount; capped = true; }
    var eqApr = (Math.pow(1 + rate / 100, 365) - 1) * 100;
    var cuCharge = amount * 0.03 * (days / 30);  // ceiling: 3%/month flat approximation
    var lines = [];
    lines.push('<p>Interest at ' + rate + '%/day for ' + days + ' days: <strong>'
      + gbp.format(charge) + '</strong> \\u2014 total to repay <strong>'
      + gbp.format(amount + charge) + '</strong>'
      + (capped ? ' <em>(limited by the FCA cap)</em>' : '') + '</p>');
    lines.push('<p>That daily rate compounds to an equivalent APR of roughly <strong>'
      + Math.round(eqApr).toLocaleString('en-GB') + '%</strong>.</p>');
    lines.push('<p>A credit union charging its legal maximum (3%/month) could charge at '
      + 'most about <strong>' + gbp.format(cuCharge) + '</strong> for the same loan \\u2014 '
      + 'and usually less, because interest is charged on the reducing balance.</p>');
    out.innerHTML = lines.join('');
  });
})();
</script>"""

DEADLINE_TOOL = """<p class="breadcrumb"><a href="/">Home</a><span>&rsaquo;</span><a href="/tools/index.html">Check your loan</a><span>&rsaquo;</span>Complaint deadlines</p>
<h1>Complaint deadline calculator</h1>
<p class="subtitle">Two dates decide a lender complaint: the day their eight weeks runs
out, and the last day you can take it to the Financial Ombudsman for free.</p>
<div class="tool-panel">
<label for="cd-complained">Date you complained to the lender</label>
<input type="date" id="cd-complained">
<label for="cd-final">Date of the lender's final response (leave blank if none yet)</label>
<input type="date" id="cd-final">
<button id="cd-go" class="btn-primary">Work out my deadlines</button>
<div id="cd-result" class="result" aria-live="polite"></div>
</div>
<article class="guide-content">
<h2>How the clock works</h2>
<p>Once you complain, the lender has <strong>8 weeks</strong> to send a final response.
If the 8 weeks pass with no final response — or you get one and disagree with it — you can
take the complaint to the <a href="https://www.financial-ombudsman.org.uk"
rel="noopener">Financial Ombudsman Service</a>, which is free and independent. You
normally have <strong>6 months from the final response</strong> to refer it. The
<a href="/guides/how-to-complain-and-win.html">complaints guide</a> covers what to write
and what a good outcome looks like.</p>
</article>
<script>
(function () {
  function fmt(d) {
    return d.toLocaleDateString('en-GB', { day: 'numeric', month: 'long', year: 'numeric' });
  }
  document.getElementById('cd-go').addEventListener('click', function () {
    var c = document.getElementById('cd-complained').value;
    var f = document.getElementById('cd-final').value;
    var out = document.getElementById('cd-result');
    var lines = [];
    if (!c && !f) {
      out.innerHTML = '<p class="warn">Enter at least one date.</p>';
      return;
    }
    if (c) {
      var d8 = new Date(c); d8.setDate(d8.getDate() + 56);
      lines.push('<p>The lender\\u2019s 8 weeks run out on <strong>' + fmt(d8)
        + '</strong>. No final response by then \\u2192 you can go straight to the '
        + 'Ombudsman.</p>');
    }
    if (f) {
      var d6 = new Date(f); d6.setMonth(d6.getMonth() + 6);
      lines.push('<p>Your 6-month window to refer the final response to the Ombudsman '
        + 'closes on <strong>' + fmt(d6) + '</strong>. Do not sit on it.</p>');
    }
    out.innerHTML = lines.join('');
  });
})();
</script>"""

for (slug, title, desc, _), body in zip(TOOLS, (CAP_TOOL, COST_TOOL, DEADLINE_TOOL)):
    write(f"tools/{slug}.html", page(title, desc, f"{BASE}/tools/{slug}.html",
                                     "tools", body, tight=True))

# ── guides hub + guides ──────────────────────────────────────────────────────
write("guides/index.html", page(
    "Your Rights When Borrowing: the Guides",
    f"{N_GUIDES} plain-English guides to the FCA rules that protect UK borrowers: the "
    "price cap, affordability duties, stopping payments, complaining, and free help.",
    f"{BASE}{hub('guides')}", "guides",
    """<p class="breadcrumb"><a href="/">Home</a><span>&rsaquo;</span>Your rights</p>
<h1>Your rights, one rule at a time</h1>
<p class="subtitle">Every guide covers one protection: what the rule says, what it looks
like when a lender breaks it, and exactly what to do next. No jargon survives contact
with this site.</p>
<div class="tool-grid">
""" + "\n".join(
        f'<div class="card"><h3><a href="/guides/{s}.html">{html.escape(t)}</a></h3>'
        f'<p>{html.escape(d)}</p>'
        f'<a href="/guides/{s}.html">Read the guide &rarr;</a></div>'
        for s, t, d in GUIDES) + """
</div>"""))

missing = [s for s, _, _ in GUIDES if s not in BODIES]
if missing:
    raise SystemExit("content_loancash.py is missing bodies for: " + ", ".join(missing))

for slug, title, desc in GUIDES:
    ld = ('<script type="application/ld+json">\n'
          + json.dumps({
              "@context": "https://schema.org", "@type": "Article",
              "headline": title, "description": desc,
              "mainEntityOfPage": f"{BASE}/guides/{slug}.html",
              "publisher": {"@type": "Organization", "name": BRAND},
              "dateModified": TODAY}, ensure_ascii=False)
          + "\n</script>\n")
    body = f"""<p class="breadcrumb"><a href="/">Home</a><span>&rsaquo;</span><a href="{hub('guides')}">Your rights</a><span>&rsaquo;</span>{html.escape(title)}</p>
<div class="guide-header">
<h1>{html.escape(title)}</h1>
<p class="subtitle">{html.escape(desc)}</p>
</div>
<article class="guide-content">
{BODIES[slug].strip()}
</article>"""
    write(f"guides/{slug}.html", page(title, desc, f"{BASE}/guides/{slug}.html",
                                      "guides", body, extra=ld, tight=True))

# ── legal + 404 ───────────────────────────────────────────────────────────────
write("legal.html", page(
    "Who We Are, and the Small Print",
    "LoanCash.co.uk is an independent information site about UK borrower protections. "
    "We do not lend, broker, or take applications, and we are not the FCA.",
    f"{BASE}/legal.html", "",
    f"""<h1>Who we are, and the small print</h1>
<article class="guide-content">
<h2>What this site is</h2>
<p>{BRAND} is an independent information site about the rules that protect people who
borrow money in the UK &mdash; especially the rules on high-cost credit. We publish
plain-English guides and browser-based checkers. That is all we do.</p>
<h2>What this site is not</h2>
<p><strong>We are not a lender and we are not a broker.</strong> You cannot apply for
anything here, we take no applications, we sell no leads, and no lender pays us to appear.
If any page ever appears to offer you a loan, something is wrong &mdash; leave and check
the address.</p>
<p><strong>We are not the Financial Conduct Authority</strong> and we are not affiliated
with, endorsed by, or connected to it in any way. We are members of the public who think
its consumer-credit rules work better when borrowers actually know them. The rules
themselves are published at <a href="https://www.fca.org.uk" rel="noopener">fca.org.uk</a>
and the authoritative text always wins over our summary of it.</p>
<h2>Not advice, and rules change</h2>
<p>Nothing here is financial or legal advice. The regulatory limits we quote &mdash; the
price cap, fee caps, time limits &mdash; are stated with the rule they come from and were
correct when written, but rules change: check the source before acting on anything that
matters, and take advice for anything serious. Where we describe rules for England and
Wales (such as Breathing Space), Scotland and Northern Ireland can differ.</p>
<h2>Your data</h2>
<p>The checkers run entirely in your browser. Nothing you type is transmitted, stored, or
seen by anyone, including us &mdash; these are static pages with no server-side
processing. We set no advertising or tracking cookies. Our host may process request logs
(including IP addresses) as a technical necessity of serving any website.</p>
<h2>If you are in difficulty</h2>
<p>Free, impartial debt help: <a href="https://www.moneyhelper.org.uk" rel="noopener">MoneyHelper</a>,
<a href="https://www.stepchange.org" rel="noopener">StepChange</a>,
<a href="https://www.citizensadvice.org.uk" rel="noopener">Citizens Advice</a>,
<a href="https://nationaldebtline.org" rel="noopener">National Debtline</a>. None of them
charges you, and several will deal with your creditors for you.</p>
<h2>Copyright</h2>
<p>Content and design &copy; {BRAND}. External sites are not under our control.</p>
</article>""", tight=True))

write("404.html", page(
    "Page not found",
    "That page does not exist on loancash.co.uk. The checkers and guides are all one "
    "click from here.",
    f"{BASE}/404.html", "",
    f"""<div class="hero">
<h1>That page isn't here</h1>
<p>The link may be old or mistyped. Everything on the site is one click away.</p>
</div>
<div class="tool-grid">
<div class="card"><h3><a href="{hub('tools')}">Check your loan</a></h3>
<p>The price cap, the true cost, and your complaint deadlines.</p></div>
<div class="card"><h3><a href="{hub('guides')}">Your rights</a></h3>
<p>The rules that protect you, one guide per rule.</p></div>
<div class="card"><h3><a href="/guides/if-you-cant-pay.html">Free help now</a></h3>
<p>If you are struggling, start here — it costs nothing.</p></div>
</div>"""))

# ── robots + sitemap ──────────────────────────────────────────────────────────
write("robots.txt", f"""# {DOMAIN}
User-agent: *
Allow: /

Disallow: /404.html

Sitemap: {BASE}/sitemap.xml
""")

urls = ([("/", "1.0"), (hub("tools"), "0.9"), (hub("guides"), "0.9")]
        + [(f"/tools/{s}.html", "0.8") for s, *_ in TOOLS]
        + [(f"/guides/{s}.html", "0.7") for s, _, _ in GUIDES]
        + [("/legal.html", "0.2")])

write("sitemap.xml", '<?xml version="1.0" encoding="UTF-8"?>\n'
      '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n'
      + "".join(f"  <url><loc>{BASE}{u}</loc><lastmod>{TODAY}</lastmod>"
                f"<priority>{p}</priority></url>\n" for u, p in urls)
      + "</urlset>\n")

print(f"{'CHECKED' if CHECK_ONLY else 'BUILT'}: home, tools hub, {N_TOOLS} tools, "
      f"guides hub, {N_GUIDES} guides, legal, 404, robots.txt, sitemap.xml "
      f"({len(urls)} URLs)")
