#!/usr/bin/env python3
"""build_pages.py — the pages that are not ported calculators.

Home, the two section hubs, the guides hub, the 13 guides, legal, 404,
robots.txt and sitemap.xml. Chrome (head/nav/footer) is imported from
build_site.py so there is exactly one definition of it for the whole site.

Guide prose lives in guides_content.py — one entry per slug, body HTML only.

Run: python3 build_pages.py
"""
import html
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from build_site import (BASE, BRAND, DOMAIN, FOOTER, GUIDES, LOAN, MORTGAGE,
                        OUT, head, nav, write)
from guides_content import BODIES

TODAY = "2026-07-31"


def page(title, desc, canonical, active, body, extra="", tight=False):
    cls = "container container-tight" if tight else "container"
    return (head(title, desc, canonical, extra) + "<body>\n" + nav(active)
            + f'<div id="content" class="{cls}">\n' + body + "\n</div>\n" + FOOTER)


def tool_cards(tools, section):
    out = ['<div class="tool-grid">']
    for slug, title, desc, blurb in tools:
        out.append(
            f'<div class="card"><h3><a href="/{section}/{slug}.html">{html.escape(title)}</a></h3>'
            f'<p>{html.escape(blurb)}</p>'
            f'<a class="btn-primary btn-block" href="/{section}/{slug}.html">Open calculator</a></div>')
    out.append("</div>")
    return "\n".join(out)


# ─────────────────────────────── home ────────────────────────────────────────
HOME_BODY = f"""<div class="hero">
<p class="eyebrow">Loans and mortgages, in one place</p>
<h1>Your borrowing does not come in separate boxes. Neither do these calculators.</h1>
<p>24 free UK calculators for loans <em>and</em> mortgages &mdash; because a car
loan changes what a mortgage lender will offer you, and a remortgage changes what
your other debt really costs. No sign-up, no credit check, nothing sent anywhere.</p>
</div>

<div class="highlight-box">
<h3 class="mt-0">Start with the question most people get wrong</h3>
<p>Almost every borrowing decision gets made on the <strong>monthly payment</strong>,
because that is the number lenders advertise. It is the wrong number. A longer term
lowers the monthly payment and raises the total cost, every single time.</p>
<p><a href="/guides/how-loans-affect-mortgage-affordability.html">How your loans cut what you can borrow &rarr;</a>
&nbsp;&middot;&nbsp;
<a href="/guides/total-cost-of-borrowing.html">Why total cost beats monthly payment &rarr;</a></p>
</div>

<h2>Mortgage calculators</h2>
<p>Repayments, affordability, Stamp Duty, overpayments, buy-to-let yields and the
cost of the remortgage cliff.</p>
{tool_cards(MORTGAGE[:6], "mortgages")}
<p class="text-center mt-40"><a class="btn-primary" href="/mortgages/">All 12 mortgage calculators</a></p>

<h2>Loan and credit calculators</h2>
<p>Personal loans, consolidation, car finance, early settlement, credit health and
rate stress tests.</p>
{tool_cards(LOAN[:6], "loans")}
<p class="text-center mt-40"><a class="btn-primary" href="/loans/">All 12 loan calculators</a></p>

<h2>Guides that join the two up</h2>
<p>Every guide here is about the point where unsecured borrowing meets a mortgage
&mdash; the questions a single-subject site cannot answer.</p>
<div class="tool-grid">
""" + "\n".join(
    f'<div class="card"><h3><a href="/guides/{s}.html">{html.escape(t)}</a></h3>'
    f'<p>{html.escape(d)}</p></div>' for s, t, d in GUIDES[:6]
) + """
</div>
<p class="text-center mt-40"><a class="btn-primary" href="/guides/">All guides</a></p>
"""

HOME_LD = f"""<script type="application/ld+json">
{{"@context":"https://schema.org","@type":"WebSite",
"name":"{BRAND}","url":"{BASE}/",
"description":"Free UK loan and mortgage calculators, and guides on how the two affect each other."}}
</script>
"""

write("index.html", page(
    "UK Loan and Mortgage Calculators",
    "24 free UK calculators for loans and mortgages, plus guides on how your other "
    "borrowing changes what a mortgage lender will offer. No sign-up, no credit check.",
    f"{BASE}/", "", HOME_BODY, extra=HOME_LD))

# ───────────────────────────── section hubs ──────────────────────────────────
write("mortgages/index.html", page(
    "Mortgage Calculators",
    "12 UK mortgage calculators: repayments and amortisation, affordability, Stamp "
    "Duty, overpayments, rental yield, equity release, bridging and fee comparison.",
    f"{BASE}/mortgages/", "mortgages",
    f"""<p class="breadcrumb"><a href="/">Home</a><span>&rsaquo;</span>Mortgage tools</p>
<h1>Mortgage calculators</h1>
<p class="subtitle">Twelve calculators covering the whole life of a mortgage &mdash;
from what you could borrow, through what it costs, to what happens when the fixed
period ends.</p>
<div class="highlight-box">
<p class="mt-0"><strong>Before you start:</strong> if you have a personal loan, a
credit card balance or car finance, that reduces what a lender will offer you.
The affordability calculator accounts for it, and
<a href="/guides/how-loans-affect-mortgage-affordability.html">this guide explains
the arithmetic</a>.</p>
</div>
{tool_cards(MORTGAGE, "mortgages")}"""))

write("loans/index.html", page(
    "Loan and Credit Calculators",
    "12 UK loan calculators: repayments, comparing offers, debt consolidation, PCP "
    "vs HP car finance, early settlement, rate stress tests and credit health.",
    f"{BASE}/loans/", "loans",
    f"""<p class="breadcrumb"><a href="/">Home</a><span>&rsaquo;</span>Loan tools</p>
<h1>Loan and credit calculators</h1>
<p class="subtitle">Twelve calculators for unsecured borrowing &mdash; what it
costs, how offers compare, and what clearing it early is worth.</p>
<div class="fca-warning-box">
<p class="mt-0"><strong>Borrowing costs money.</strong> Late or missed repayments
can cause serious problems and may affect your credit file for years. For free,
impartial help go to <a href="https://www.moneyhelper.org.uk">MoneyHelper</a> or
<a href="https://www.stepchange.org">StepChange</a>.</p>
</div>
{tool_cards(LOAN, "loans")}"""))

# ────────────────────────────── guides hub ───────────────────────────────────
write("guides/index.html", page(
    "Guides: Where Loans and Mortgages Meet",
    "Guides on the point where unsecured borrowing meets a mortgage: borrowing "
    "power, consolidation, remortgaging with other debt, and total cost of credit.",
    f"{BASE}/guides/", "guides",
    """<p class="breadcrumb"><a href="/">Home</a><span>&rsaquo;</span>Guides</p>
<h1>Guides</h1>
<p class="subtitle">Most personal finance writing treats loans and mortgages as two
unrelated subjects. In a real household they are one budget, one credit file and one
set of trade-offs. These guides are about the crossing points.</p>
<div class="tool-grid">
""" + "\n".join(
        f'<div class="card"><h3><a href="/guides/{s}.html">{html.escape(t)}</a></h3>'
        f'<p>{html.escape(d)}</p>'
        f'<a href="/guides/{s}.html">Read the guide &rarr;</a></div>'
        for s, t, d in GUIDES) + """
</div>"""))

# ──────────────────────────────── guides ─────────────────────────────────────
missing = [s for s, _, _ in GUIDES if s not in BODIES]
if missing:
    raise SystemExit("guides_content.py is missing bodies for: " + ", ".join(missing))

for slug, title, desc in GUIDES:
    ld = f"""<script type="application/ld+json">
{{"@context":"https://schema.org","@type":"Article",
"headline":{html.escape(repr(title).replace("'", '"'))},
"description":"{html.escape(desc, quote=True)}",
"mainEntityOfPage":"{BASE}/guides/{slug}.html",
"publisher":{{"@type":"Organization","name":"{BRAND}"}},
"dateModified":"{TODAY}"}}
</script>
"""
    body = f"""<p class="breadcrumb"><a href="/">Home</a><span>&rsaquo;</span><a href="/guides/">Guides</a><span>&rsaquo;</span>{html.escape(title)}</p>
<div class="guide-header">
<h1>{html.escape(title)}</h1>
<p class="subtitle">{html.escape(desc)}</p>
</div>
<article class="guide-content">
{BODIES[slug].strip()}
</article>"""
    write(f"guides/{slug}.html", page(title, desc, f"{BASE}/guides/{slug}.html",
                                      "guides", body, extra=ld, tight=True))

# ───────────────────────────── legal and 404 ─────────────────────────────────
write("legal.html", page(
    "Legal, Privacy and Cookies",
    "Terms, privacy and cookie information for loanandmortgagecalculator.co.uk. "
    "The calculators run entirely in your browser and send nothing anywhere.",
    f"{BASE}/legal.html", "",
    f"""<h1>Legal, privacy and cookies</h1>
<article class="guide-content">
<h2>No advice, and no offer of credit</h2>
<p>{BRAND} publishes calculators and general information about UK borrowing. Nothing
on this site is financial, legal or tax advice, and no result produced by any
calculator is an offer of credit or a decision in principle. We are not a lender,
a broker or an intermediary, and we do not introduce you to any.</p>
<p>Every figure is an estimate produced from the numbers you type in. A lender's own
assessment uses information we do not have &mdash; your full credit file, your
verified income, its own policy rules and its own stress-test assumptions &mdash; and
will differ, sometimes substantially. Rates, tax bands and lending rules change.
Check the current position before you act, and take regulated advice for any
decision that matters.</p>

<h2>Your data</h2>
<p><strong>The calculators run entirely in your browser.</strong> The numbers you
enter are not transmitted to us, because there is nowhere for them to go &mdash;
these are static pages with no server-side processing and no forms that submit
anywhere.</p>
<p>Two tools &mdash; the property portfolio calculator and the application tracker
&mdash; save what you enter in your browser's own local storage so it is still there
when you come back. That data never leaves your device, and clearing your browser
data removes it.</p>
<p>We set no advertising or tracking cookies of our own. Our hosting provider may
process request logs, including IP addresses, for security and to serve the site;
that is a technical necessity of serving any web page.</p>

<h2>Accuracy, and telling us about a mistake</h2>
<p>The calculators implement standard published formulae and we test them, but
mistakes are possible. If a result looks wrong we would genuinely like to know, and
we would rather hear about it than not.</p>

<h2>If you are struggling</h2>
<p>If you are having difficulty with repayments, free and impartial help is
available and it is worth taking early. Try
<a href="https://www.moneyhelper.org.uk">MoneyHelper</a>,
<a href="https://www.citizensadvice.org.uk">Citizens Advice</a>,
<a href="https://www.stepchange.org">StepChange</a> or
<a href="https://nationaldebtline.org">National Debtline</a>. None of them charges
you, and several will deal with creditors on your behalf.</p>

<h2>Copyright</h2>
<p>The content and design of this site are &copy; {BRAND}. External sites we link
to are not under our control and we are not responsible for their content.</p>
</article>""", tight=True))

write("404.html", page(
    "Page not found",
    "That page does not exist on loanandmortgagecalculator.co.uk. The calculators "
    "and guides are all linked from here.",
    f"{BASE}/404.html", "",
    """<div class="hero">
<h1>That page isn't here</h1>
<p>The link may be old, or slightly mistyped. Everything on the site is one click
away below.</p>
</div>
<div class="tool-grid">
<div class="card"><h3><a href="/mortgages/">Mortgage calculators</a></h3>
<p>Repayments, affordability, Stamp Duty, overpayments, yields and more.</p></div>
<div class="card"><h3><a href="/loans/">Loan calculators</a></h3>
<p>Personal loans, consolidation, car finance, early settlement, stress tests.</p></div>
<div class="card"><h3><a href="/guides/">Guides</a></h3>
<p>How your loans and your mortgage affect each other.</p></div>
</div>"""))

# ─────────────────────────── robots and sitemap ──────────────────────────────
write("robots.txt", f"""# {DOMAIN}
User-agent: *
Allow: /

# Nothing here needs crawling and it is all machine noise
Disallow: /404.html

Sitemap: {BASE}/sitemap.xml
""")

urls = ([("/", "1.0"), ("/mortgages/", "0.9"), ("/loans/", "0.9"), ("/guides/", "0.9")]
        + [(f"/mortgages/{s}.html", "0.8") for s, *_ in MORTGAGE]
        + [(f"/loans/{s}.html", "0.8") for s, *_ in LOAN]
        + [(f"/guides/{s}.html", "0.7") for s, _, _ in GUIDES]
        + [("/legal.html", "0.2")])

write("sitemap.xml", '<?xml version="1.0" encoding="UTF-8"?>\n'
      '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n'
      + "".join(f"  <url><loc>{BASE}{u}</loc><lastmod>{TODAY}</lastmod>"
                f"<priority>{p}</priority></url>\n" for u, p in urls)
      + "</urlset>\n")

print(f"home, 2 section hubs, guides hub, {len(GUIDES)} guides, legal, 404, "
      f"robots.txt, sitemap.xml ({len(urls)} URLs)")
