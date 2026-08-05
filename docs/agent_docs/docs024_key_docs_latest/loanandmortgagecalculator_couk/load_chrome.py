#!/usr/bin/env python3
"""load_chrome.py — install this site's chrome, refusing to install broken chrome.

Adapted from loancalculator_couk/decompose/load_chrome.py (read its docstring
for why each refusal exists — every one of them caught a real fault there).
Differences here, each deliberate:

  * INSERT, not UPDATE: this site has ZERO site_components rows (the CONTRIB
    file measured it, and it is why the first decomposed page would otherwise
    assemble against buildDefaultHead's plural `styles.css` — a 404 here —
    with no header or footer at all). If a row exists at apply time, another
    session wrote it since this check ran: STOP, do not overwrite.
  * Rows are born LOCKED (permanent, like the sibling's chrome ended up):
    authored chrome, no automation may rewrite it.
  * The brand link is `href="/"`, which is not a `pages.url`. The Cloudflare
    worker rewrites exactly `/` → `/index.html` and nothing else (measured in
    this lane's 07-31 session 2, defect 1 — the same worker 404s `/loans/`),
    so the pages-membership check applies the same single rewrite and no
    other, mirroring verify_site.py.
  * Both header AND footer link sets are validated against `pages` — this
    site's nav is 4 links and its footer carries 13, so the footer is where a
    gutted link list would actually bite.

SAFE BY ORDERING, same as the sibling: while every page holds one verbatim
row, site_components is read by nothing that ships. The moment the first page
is decomposed this becomes that page's live chrome.

Usage:  python3 load_chrome.py --check
        python3 load_chrome.py --apply
"""
import os
import re
import subprocess
import sys
import urllib.request

HERE = os.path.dirname(os.path.abspath(__file__))
CHROME = os.path.join(HERE, "chrome")
SITE_ID = "ed633ada-f8af-424b-b4d4-8af79160dbcd"
DOMAIN = "loanandmortgagecalculator.co.uk"
LOCKED_BY = "loanandmortgage_authored_chrome_20260805"
MIN_HEADER_LINKS = 4   # brand + the three section links
MIN_FOOTER_LINKS = 12  # the hand-built footer carries 13 internal links

PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
        "--", "psql", "-U", "clients_user", "-d", "clients_db"]
BALANCED = [("<script", "</script>"), ("<style", "</style>"), ("<section", "</section>"),
            ("<div", "</div>"), ("<fieldset", "</fieldset>")]


def psql(sql, stdin=None):
    args = PSQL + (["-v", "ON_ERROR_STOP=1"] if stdin else
                   ["-tA", "-v", "ON_ERROR_STOP=1", "-c", sql])
    r = subprocess.run(args, input=stdin, capture_output=True, text=True)
    if r.returncode != 0:
        raise RuntimeError((r.stderr or r.stdout).strip()[:600])
    return r.stdout.strip()


def http_status(url):
    # The User-Agent is load-bearing: Cloudflare answers urllib's default UA
    # with 403 (sibling lane, measured three ways). A checker without it marks
    # every asset on a healthy site unreachable.
    req = urllib.request.Request(url, method="HEAD", headers={
        "User-Agent": "Mozilla/5.0 (compatible; loanandmortgage-chrome-check/1.0)"})
    try:
        with urllib.request.urlopen(req, timeout=20) as resp:
            return resp.status
    except urllib.error.HTTPError as e:
        return e.code
    except Exception:  # noqa: BLE001
        return 0


def dollar_tag(*bodies):
    for i in range(1000):
        tag = "$ch%d$" % i
        if all(tag not in b for b in bodies):
            return tag
    raise RuntimeError("no free dollar-quote tag")


def main():
    apply = "--apply" in sys.argv
    if not apply and "--check" not in sys.argv:
        print(__doc__)
        return 2

    slots = {}
    for slot in ("head", "header", "footer"):
        slots[slot] = open(os.path.join(CHROME, slot + ".html"),
                           encoding="utf-8").read().rstrip("\n")

    problems = []

    # structural balance (a literal tag inside a comment breaks this — it is
    # how the sibling's first head.html draft failed)
    for slot, html in slots.items():
        low = html.lower()
        for op, cl in BALANCED:
            if low.count(op) != low.count(cl):
                problems.append("%s: unbalanced %s (%d open, %d close)"
                                % (slot, op, low.count(op), low.count(cl)))

    # assembly's injection points
    if "<title></title>" not in slots["head"]:
        problems.append("head: no literal <title></title>")
    if 'content=""' not in slots["head"]:
        problems.append('head: no literal content=""')
    else:
        first = slots["head"].index('content=""')
        before = slots["head"][:first]
        if 'name="description"' not in before.split("<meta")[-1]:
            problems.append('head: the FIRST content="" is not on the description '
                            "meta — assembly writes the description into whichever "
                            "tag holds it")

    # every referenced asset resolves, with a positive AND a negative control:
    # a checker that cannot tell style.css (200) from styles.css (404) from the
    # same code path has told you nothing (sibling landmine).
    refs = set()
    for html in slots.values():
        refs |= {m for m in re.findall(r'(?:href|src)="(/[^"]+)"', html)
                 if m.startswith("/assets/") or m.endswith((".css", ".js", ".png",
                                                            ".ico", ".svg"))}
    neg = http_status("https://%s/assets/css/styles.css" % DOMAIN)
    if neg == 200:
        problems.append("negative control /assets/css/styles.css returned 200 — "
                        "the checker cannot distinguish good from bad")
    print("== assets referenced by the chrome (negative control: styles.css -> %s) =="
          % neg)
    for u in sorted(refs):
        code = http_status("https://%s%s" % (DOMAIN, u))
        print("   %-36s %s" % (u, code))
        if code != 200:
            problems.append("%s returns %s" % (u, code))

    # header and footer links resolve against pages, modelling the worker's
    # single rewrite
    valid = set(x for x in psql(
        "SELECT url FROM pages WHERE site_id='%s' AND status='active';" % SITE_ID
    ).splitlines() if x)

    for slot, minimum in (("header", MIN_HEADER_LINKS), ("footer", MIN_FOOTER_LINKS)):
        links = [h for h in re.findall(r'href="(/[^"#?]*)"', slots[slot])]
        internal = [h for h in links if not h.startswith("//")]
        print("== %s: %d internal link(s), %d distinct ==" %
              (slot, len(internal), len(set(internal))))
        if len(internal) < minimum:
            problems.append("%s carries %d internal links, expected >= %d"
                            % (slot, len(internal), minimum))
        for h in sorted(set(internal)):
            resolved = "/index.html" if h == "/" else h
            if resolved not in valid:
                problems.append("%s links %s which is not an active page" % (slot, h))

    if problems:
        print("\nREFUSING TO WRITE (%d problem(s)):" % len(problems))
        for p in problems:
            print("  " + p)
        return 1
    print("\nall checks pass: assets resolve (controls proven), links resolve, "
          "injection points intact, tags balanced")

    existing = psql("SELECT count(*) FROM site_components WHERE site_id='%s';" % SITE_ID)
    if existing != "0":
        print("REFUSING: %s site_components row(s) exist for this site now and "
              "there were 0 at design time — another session wrote them. Read "
              "them before deciding anything." % existing)
        return 1

    if not apply:
        print("--check: would INSERT %d locked site_components row(s)" % len(slots))
        return 0

    stmts = ["BEGIN;"]
    for slot, html in slots.items():
        t = dollar_tag(html)
        stmts.append(
            "INSERT INTO site_components (site_id, slot_name, rendered_html, "
            "build_status, locked_at, locked_by, lock_type) "
            "VALUES ('{s}', '{n}', {t}{h}{t}, 'rendered', now(), '{lb}', "
            "'permanent');".format(s=SITE_ID, n=slot, t=t, h=html, lb=LOCKED_BY))
    stmts.append("COMMIT;")
    psql(None, stdin="\n".join(stmts))

    print("\n== read back ==")
    print(psql("SELECT slot_name, octet_length(rendered_html), build_status, "
               "lock_type, locked_by FROM site_components WHERE site_id='%s' "
               "ORDER BY slot_name;" % SITE_ID))
    print("\nINERT until a page is decomposed: assemblePage never runs for a "
          "page that still holds one verbatim row.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
