#!/usr/bin/env python3
"""load_chrome.py — install the site chrome, and refuse to install broken chrome.

SAFE BY ORDERING, like load_components.py before it. `assemblePage` reads
site_components only when it assembles, and it never assembles a page that is
still verbatim — `loadVerbatimPageHTML` returns first. So while all 27 pages hold
one verbatim row, writing this table changes nothing that ships, and the result
can be checked at leisure. The moment the first page is decomposed, it becomes
live for that page.

WHY THE VALIDATION IS THE POINT OF THIS SCRIPT. The rows it replaces were written
on 2026-08-01 and were wrong in three ways that no query against this table could
see: a stylesheet href that 404s, a nav with no links in it, and two 404 image
links. A rebuild of all 27 pages ran against them and reported success, because
none of those pages assembles. **"Are there rows?" is satisfiable by chrome that
would take the site down.** So this refuses to write unless:

  1. every asset the chrome references returns 200 (fetched, not assumed);
  2. the header carries at least MIN_NAV_LINKS internal links, and every one
     resolves to a real row in `pages`;
  3. the head still contains assembly's two literal injection points —
     `<title></title>` and `content=""` — because assembly rewrites them by
     exact string match and a reordered head silently drops the page title and
     puts the meta description in whatever tag holds the first empty content=";
  4. all five structural tag pairs balance, the same predicate the birth-write
     guard uses. A literal tag inside a comment breaks this, which is how the
     first draft of head.html failed.

Usage:  python3 load_chrome.py --check
        python3 load_chrome.py --apply
"""
import os
import re
import subprocess
import sys
import urllib.request

HERE = os.path.dirname(os.path.abspath(__file__))
LANE = os.path.dirname(HERE)
CHROME = os.path.join(LANE, "chrome")
SITE_ID = "0162cde4-633e-45e9-8ca6-87a6b2fe1d26"
DOMAIN = "loancalculator.co.uk"
MIN_NAV_LINKS = 20  # the hand-built nav has 25; a nav that lost most of its
                    # links is the exact failure being guarded against

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
    # ⚠ THE USER-AGENT IS LOAD-BEARING. Cloudflare fronts these zones and answers
    # 403 to urllib's default `Python-urllib/3.x` — measured against
    # /assets/css/style.css: curl 200 (GET and HEAD, with or without a UA),
    # urllib 403 (GET and HEAD), urllib with a browser UA 200. So it is the agent
    # string, not the method.
    #
    # Left unset, this checker reports every asset on a perfectly healthy site as
    # missing, and its output is indistinguishable from the real 404s it was
    # written to catch — which is worse than no checker, because the honest
    # response to "everything is broken" is to stop believing the check.
    req = urllib.request.Request(url, method="HEAD", headers={
        "User-Agent": "Mozilla/5.0 (compatible; loancalculator-chrome-check/1.0)"})
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
        slots[slot] = open(os.path.join(CHROME, slot + ".html"), encoding="utf-8").read().rstrip("\n")

    problems = []

    # 4. structural balance
    for slot, html in slots.items():
        low = html.lower()
        for op, cl in BALANCED:
            if low.count(op) != low.count(cl):
                problems.append("%s: unbalanced %s (%d open, %d close)"
                                % (slot, op, low.count(op), low.count(cl)))

    # 3. assembly's injection points
    if "<title></title>" not in slots["head"]:
        problems.append("head: no literal <title></title> — assembly's title "
                        "replacement matches <title>[^<]*</title> and would have "
                        "nothing to write into")
    if 'content=""' not in slots["head"]:
        problems.append('head: no literal content="" — assembly writes the meta '
                        "description into the FIRST one it finds")
    else:
        first = slots["head"].index('content=""')
        before = slots["head"][:first]
        if 'name="description"' not in before.split("<meta")[-1]:
            problems.append('head: the first content="" is not on the description '
                            "meta — assembly would write the description into "
                            "whichever tag holds it")

    # 1. every referenced asset resolves
    refs = set()
    for html in slots.values():
        refs |= {m for m in re.findall(r'(?:href|src)="(/[^"]+)"', html)
                 if m.startswith("/assets/") or m.endswith((".css", ".js", ".png",
                                                            ".ico", ".svg"))}
    print("== assets referenced by the chrome ==")
    for u in sorted(refs):
        code = http_status("https://%s%s" % (DOMAIN, u))
        print("   %-36s %s" % (u, code))
        if code != 200:
            problems.append("%s returns %s — the 2026-08-01 chrome shipped exactly "
                            "this and nothing noticed" % (u, code))

    # 2. the nav resolves against `pages`
    valid = set(x for x in psql(
        "SELECT url FROM pages WHERE site_id='%s' AND status='active';" % SITE_ID
    ).splitlines() if x)
    navlinks = [h for h in re.findall(r'href="(/[^"#?]*)"', slots["header"])]
    print("\n== nav ==\n   %d internal link(s), %d distinct"
          % (len(navlinks), len(set(navlinks))))
    if len(navlinks) < MIN_NAV_LINKS:
        problems.append("header carries %d internal links, expected >= %d — the "
                        "row this replaces had ZERO and read as present"
                        % (len(navlinks), MIN_NAV_LINKS))
    for h in sorted(set(navlinks)):
        if h not in valid:
            problems.append("header links %s which is not an active page" % h)

    if problems:
        print("\nREFUSING TO WRITE (%d problem(s)):" % len(problems))
        for p in problems:
            print("  " + p)
        return 1
    print("\nall checks pass: assets resolve, nav resolves, injection points intact, "
          "tags balanced")

    if not apply:
        print("--check: would replace %d site_components row(s)" % len(slots))
        return 0

    # The previous rows are kept, not dropped: they are another lane's write and
    # the only copy of what was there. Restore is a single UPDATE from the backup.
    stmts = ["BEGIN;",
             "CREATE TABLE IF NOT EXISTS site_components_bak_20260802_decomp "
             "(LIKE site_components INCLUDING ALL);",
             "INSERT INTO site_components_bak_20260802_decomp "
             "SELECT * FROM site_components WHERE site_id='%s' "
             "AND slot_name NOT IN (SELECT slot_name FROM "
             "site_components_bak_20260802_decomp WHERE site_id='%s');"
             % (SITE_ID, SITE_ID)]
    for slot, html in slots.items():
        t = dollar_tag(html)
        stmts.append(
            "UPDATE site_components SET rendered_html={t}{h}{t}, updated_at=now() "
            "WHERE site_id='{s}' AND slot_name='{n}';".format(
                t=t, h=html, s=SITE_ID, n=slot))
    stmts.append("COMMIT;")
    psql(None, stdin="\n".join(stmts))

    print("\n== read back ==")
    print(psql("SELECT slot_name, octet_length(rendered_html), updated_at "
               "FROM site_components WHERE site_id='%s' ORDER BY slot_name;" % SITE_ID))
    print("\nINERT until a page is decomposed: assemblePage never runs for a page "
          "that still holds one verbatim row.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
