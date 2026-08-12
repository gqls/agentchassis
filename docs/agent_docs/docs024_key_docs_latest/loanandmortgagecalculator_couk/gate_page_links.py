#!/usr/bin/env python3
"""gate_page_links.py — assert a rewritten page still links everything it is required to.

WHY THIS EXISTS. Nothing in the platform counts links. The section-shrink floor
(`bugs_open/178`) measures TEXT VOLUME, so a rewrite that keeps the word count and
silently drops five navigation links sails through it; `bugs_open/253` is about markup
CLASSES, not hrefs; the claims checker is opt-in and this site has never opted in. So
the only thing standing between a homepage rewrite and a quietly orphaned guide is
somebody diffing by hand — which is exactly how both of these were caught:

  round 4 (2026-08-11): swapped 2 calculator cards         (allowed, but unnoticed)
  round 6 (2026-08-12): dropped 5 of 13 GUIDE links        (not allowed, not noticed)

Both times the brief said, in prose, "keep every internal link that exists on the page
today". A prose instruction to preserve a SET is not reliably followed. The fix in the
brief is to enumerate the set as data (`content_direction.required_links`); the fix
here is to make the assertion mechanical, so the next rewrite cannot be graded by
whether somebody remembered to look.

WHAT IT CHECKS. For each page that carries `content_direction.required_links`, every
URL named there must appear as an `href` in that page's own `page_components` rows.
The required set is the page's OWN declaration, so the gate needs no configuration of
its own and cannot drift from the brief the writer was given — if you change the brief,
you change the gate in the same edit.

WHAT IT DELIBERATELY DOES NOT CHECK. It does not require the link TEXT to match, and it
does not forbid extra links. A rewrite is allowed to add navigation and to choose its
own twelve calculator cards; what it may not do is drop something the brief named. A
tighter gate would fail on legitimate editorial choices and get switched off.

⚠ Read the hrefs from `page_components`, NOT from the served page. The served page
carries header and footer nav from `site_components`, which links most of the site
regardless — so a served-page check passes while the page's own body has been gutted.
Round 6 measured 29 unique links on the served page and 23 in the component.

Run:  python3 gate_page_links.py                    # report; exit 1 if any page fails
      python3 gate_page_links.py --domain loancash.co.uk
      python3 gate_page_links.py --self-test        # MUST fail, or the gate is inert
"""
import argparse
import re
import subprocess
import sys

DOMAIN = "loanandmortgagecalculator.co.uk"
PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db"]


def psql(sql):
    r = subprocess.run(PSQL + ["-t", "-A", "-F", "\t", "-c", sql],
                       capture_output=True, text=True, timeout=180)
    if r.returncode != 0:
        sys.exit(f"psql failed:\n{r.stderr.strip()}")
    return r.stdout


def check(domain, induce=False):
    """The assertion runs IN POSTGRES, one row per (page, required link), and only the
    verdict crosses the wire.

    ⚠ THE FIRST VERSION OF THIS FUNCTION SHIPPED THE HTML BACK AND SPLIT PSQL'S
    TAB-DELIMITED OUTPUT IN PYTHON. Component HTML contains tabs and newlines, so one
    logical row arrived as dozens of lines, a `len(parts) < 3: continue` guard silently
    dropped the continuation, and the gate reported 10 links missing from a page that
    demonstrably had 37 hrefs. It failed the page it was written to protect, for the
    wrong reason, and a green run would have been just as untrustworthy. Digest and
    compare in the database; never parse HTML out of psql's text format.

    The membership test is `href="<url>"` including both quotes, so /loans/index.html
    cannot be satisfied by /loans/index.html?x or by a longer path that contains it.
    """
    induced_sql = ""
    if induce:
        # A required URL that no page can possibly link. The gate MUST report it.
        induced_sql = "UNION ALL SELECT p.id, p.name, '/guides/this-guide-does-not-exist.html'"
        induced_sql += f""" FROM pages p JOIN sites s ON s.id=p.site_id
            WHERE s.domain='{domain}' AND p.content_direction ? 'required_links'"""

    out = psql(f"""
        WITH pages_with_reqs AS (
            SELECT p.id, p.name, p.content_direction->'required_links' AS reqs
            FROM pages p JOIN sites s ON s.id = p.site_id
            WHERE s.domain = '{domain}' AND p.content_direction ? 'required_links'
        ),
        required AS (
            SELECT id, name,
                   split_part(btrim(x), ' ', 1) AS url
            FROM pages_with_reqs, jsonb_array_elements_text(reqs) AS t(x)
            {induced_sql}
        ),
        html AS (
            SELECT page_id, string_agg(rendered_html, ' ') AS h
            FROM page_components GROUP BY page_id
        )
        SELECT r.name,
               count(*) AS required,
               count(*) FILTER (WHERE position('href="' || r.url || '"' in coalesce(h.h,'')) = 0) AS missing,
               coalesce(string_agg(r.url, ' ') FILTER (WHERE position('href="' || r.url || '"' in coalesce(h.h,'')) = 0), '') AS missing_urls
        FROM required r LEFT JOIN html h ON h.page_id = r.id
        GROUP BY r.name ORDER BY r.name;""")

    failures, checked = 0, 0
    for line in out.splitlines():
        if not line.strip():
            continue
        name, required, missing, missing_urls = (line.split("\t") + ["", "", ""])[:4]
        checked += 1
        if int(missing) > 0:
            failures += 1
            print(f"FAIL {name}: {missing} of {required} required link(s) absent from its own components")
            for m in missing_urls.split():
                print(f"       missing: {m}")
        else:
            print(f"ok   {name}: all {required} required links present")

    if not checked:
        print(f"NOTHING CHECKED on {domain} — no page declares content_direction.required_links.")
        print("That is not a pass. Enumerate the set in the brief first; see the module docstring.")
        return 1
    print(f"\n{checked} page(s) checked, {failures} failing.")
    return 1 if failures else 0


if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("--domain", default=DOMAIN)
    ap.add_argument("--self-test", action="store_true",
                    help="add an unlinkable required URL; the gate MUST fail")
    a = ap.parse_args()
    if a.self_test:
        rc = check(a.domain, induce=True)
        if rc == 0:
            sys.exit("CONTROL FAILED: the gate passed with an impossible required link. "
                     "It is not reading the components. Do not trust a green run.")
        print("\nCONTROL OK: the gate fails when a required link is absent.")
        sys.exit(0)
    sys.exit(check(a.domain))
