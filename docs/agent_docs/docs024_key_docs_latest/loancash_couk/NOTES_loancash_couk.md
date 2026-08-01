# NOTES — loancash.co.uk

Append-only, newest at the bottom.

---

## 2026-08-01 — session 1: built and deployed to B2 in one sitting

Owner ruling P6 inverted this lane's own recommendation: the register had said don't
build loancash.co.uk (payday-adjacent name, vulnerable audience); the owner's direction
is to build it AS the protection — the borrower's guide to the FCA rulebook. Register
entry L10 written FIRST, per the portfolio rule; `check_register.py` green before a line
of site code.

**What was built:** 3 rule-checking tools (price-cap checker, true-cost calculator,
complaint-deadline calculator) + 11 guides (one FCA protection each) + hubs, legal, 404,
robots, sitemap — 24 files. Fresh compact CSS, deliberately divergent palette (deep
green/amber vs the siblings' navy). Chrome carries the two identity statements on every
page: not a lender, and independent of the FCA.

**Verified:** 0 dead refs of 20 under the strict worker-model resolver; sitemap exact
both ways; canonicals self-name; 12/12 ld+json parse; both `write()` assertions
mutation-tested red; 3/3 tools RESPONDS in real chromium; and the arithmetic
hand-verified on known cases in both directions — a constructed breach (£300/30d/£400
charged → max £72.00, ceiling £300.00, BREACH) and a within-cap case, plus the deadline
tool (2026-06-01 + 56 days = 27 July 2026). Deployed: sites repo `803bf68c3`, run
`30691645835`, **24/24 uploads**, domain named.

Missteps, because they are the point:

- **The first browser probe failed on a missing favicon**, not the tool — the browser's
  automatic `/favicon.ico` request 404'd and `evalpage.py` surfaced it as the only
  console error, drowning the VALUE line. Same lesson as the mortgagecalculator baseline
  (its one "real" console error was the same thing). Fixed properly (SVG favicon +
  `<link rel="icon">`) rather than excused.
- **Killed my own shell twice** stopping the local server: `kill $(pgrep -f "http.server
  8766")` matches the invoking shell's own command line, SIGTERMs it, exit 144, output
  lost — and the first loss swallowed a `git commit` that then had to be redone. The
  bracket trick (`pkill -f "http[.]server 8766"`) is the fix; now in the RUNBOOK.
- The first commit attempt also carried backticks in `-m` — caught BEFORE running it
  this time (the WRONG_CALLS row from yesterday did its job); message went via `-F`.

**[UNVERIFIED], marked as such:** the regulatory facts are stated from knowledge, each
tied to its rule name (CONC 5A, CONC 5.2A, CONC 7, DISP 8-weeks/6-months, the 2015 cap,
2019 RTO cap, 2020 overdraft reforms, Breathing Space 60 days, credit-union 3%/month,
Stop Loan Sharks numbers). They are stable, but no external verification pass has been
run against fca.org.uk / gov.uk current text. Before adoption hands these guides to the
framework, a citation-check pass is the right gate — and the site's own legal page
already tells readers the source always wins over our summary.

**Not done:** DNS/route (owner), live verification, adoption, site_specs. RUNBOOK §4 has
the order.
