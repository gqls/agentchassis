#!/usr/bin/env python3
"""Test a site's evidence_base ban list the way it will actually be used.

WHY THIS EXISTS. Inspecting a ban list tells you what you MEANT. It does not tell you
what the regexes DO. Every defect this script has found on apis.uk passed a `jq -e .`
and looked correct on the page:

  * `\\\\.` decoding to "literal backslash" rather than "decimal point" — a VALID regex,
    so the evidence_base schema note's own safety net ("an invalid regex degrades to a
    literal substring") never fired. It would have reported clean for ever.
  * "2 million flowers" — the single most repeated bee statistic — escaping every
    digit-adjacent pattern, because a magnitude WORD sat between number and noun.
  * "colonies fell by 40%" escaping the decline pattern, which required the decline word
    to come AFTER the figure.
  * "the population has halved since 1990" escaping both, because `halved` carries its
    own magnitude and offers no number to anchor on.
  * "sign up to our newsletter" slipping between `sign up` and `our`.

Five real gaps, none of them visible by reading the list. All five were found by asserting
on SENTENCES.

USAGE
  python3 check_evidence_base.py <domain> [--url https://<domain>/]

It asserts three things, and the third is the one people forget:
  1. every FORBIDDEN sentence is caught  (no gaps)
  2. every PERMITTED sentence is clean   (no false positives)
  3. the site's OWN LIVE PROSE is clean  (the list does not block the copy we want)

Exit 0 only if all three hold; exit 2 if check 3 could not run at all (--skip-live to
allow that deliberately, e.g. before first deploy). A check that silently does not run
is worse than no check. Edit FORBIDDEN/PERMITTED for a site that is not about bees;
the two lists are the specification, and they are cheap to extend the moment you think of
a sentence the page must never contain.
"""
import json, re, sys, subprocess, urllib.request, html

FORBIDDEN = [
    "a colony holds sixty thousand bees", "she flies 55,000 miles", "two million flowers",
    "2 million flowers", "bees pollinate 75% of crops", "a worker lives six weeks",
    "colonies fell by 40%", "numbers have declined by a third",
    "the population has halved since 1990", "hives dropped by 12%",
    "the species was named in 1758", "numbers have doubled",
    "Einstein said four years left to live", "studies show bees are struggling",
    "one in three bites", "270 species of bee", "buy our honey today",
    "see our API documentation", "an unrelated technical service on a different hostname",
    "I have kept bees for years", "sign up to our newsletter", "held at 35 degrees",
    "follow us for more bee facts",
]
PERMITTED = [
    "The cells of comb have six sides.", "a sheet of six-sided cells", "A bee has six legs.",
    "a colony has one queen", "the hexagon uses the least wall for the most space",
    "Half the bees leave with the old queen to find a new home.",
    "A swarm is a colony reproducing.", "Most bees live alone.",
    "A colony in high summer is crowded.",
    "Pollination happens along the way. The bee came for the nectar.",
    "A returning forager climbs onto the vertical comb and dances the direction she flew.",
    "Wax comes out of the bee herself, in pale flakes along the underside of her abdomen.",
]

def load_bans(domain):
    q = ("SELECT data::text FROM site_specs ss JOIN sites s ON s.id=ss.site_id "
         f"WHERE s.domain='{domain}' AND ss.aspect='evidence_base' AND ss.is_current;")
    r = subprocess.run(["kubectl","-n","ai-persona-system","exec","-i","postgres-clients-0","--",
                        "psql","-U","clients_user","-d","clients_db","-t","-A","-c",q],
                       capture_output=True, text=True)
    if r.returncode or not r.stdout.strip():
        sys.exit(f"no current evidence_base for {domain}: {r.stderr[:300]}")
    return json.loads(r.stdout.strip())

def prose(url):
    """Fetch the served page. A FAILED FETCH IS A FAILURE, NOT A SKIP.

    The first cut of this returned "" on any exception and printed a skip line, so
    check 3 silently did not run and the script still exited 0 saying ALL CONSISTENT.
    Cloudflare 403s urllib's default User-Agent, so that fired immediately and on the
    site this was written for. A check that quietly does not run is worse than no
    check: it produces a PASS that outlives the blindness and gets quoted later.
    """
    req = urllib.request.Request(url, headers={
        "User-Agent": "Mozilla/5.0 (evidence-base-check; +apis_uk_bees_homepage lane)",
        "Accept": "text/html,*/*",
    })
    raw = urllib.request.urlopen(req, timeout=25).read().decode("utf-8", "replace")
    body = re.sub(r"(?is)<(script|style|head)[^>]*>.*?</\1>", " ", raw)
    txt = re.sub(r"\s+", " ", html.unescape(re.sub(r"(?s)<[^>]+>", " ", body)))
    if len(txt) < 500:
        raise RuntimeError(f"served page is only {len(txt)} chars of text — refusing to "
                           "call the ban list clean against nothing")
    return txt

def main():
    domain = sys.argv[1] if len(sys.argv) > 1 else sys.exit(__doc__)
    url = next((a.split("=",1)[1] for a in sys.argv if a.startswith("--url=")), f"https://{domain}/")
    eb = load_bans(domain); bans = eb["banned_claims"]
    bad = 0

    for b in bans:                       # a pattern that will not compile is a silent hole
        try: re.compile(b["pattern"])
        except re.error as e:
            bad += 1; print(f"  INVALID REGEX {b['pattern']!r}: {e}")

    for t in FORBIDDEN:
        if not [b for b in bans if re.search(b["pattern"], t, re.I)]:
            bad += 1; print(f"  GAP (nothing catches this): {t}")

    for t in PERMITTED:
        f = [b["pattern"] for b in bans if re.search(b["pattern"], t, re.I)]
        if f:
            bad += 1; print(f"  FALSE POSITIVE: {t}\n     <- {f}")

    try:
        txt = prose(url)
    except Exception as e:
        # deliberately fatal: see prose.__doc__
        print(f"  LIVE PROSE CHECK COULD NOT RUN: {e}")
        print(f"     (pass --skip-live ONLY if the page is not deployed yet)")
        if "--skip-live" not in sys.argv:
            print("\nRESULT: INCOMPLETE — 2 of 3 checks ran. Not a pass.")
            sys.exit(2)
        txt = ""
    for b in bans:
        m = next(iter(re.finditer(b["pattern"], txt, re.I)), None) if txt else None
        if m:
            bad += 1
            print(f"  FIRES ON LIVE PROSE: {b['pattern']}\n     ...{txt[max(0,m.start()-60):m.end()+60]}...")

    print(f"\nbans={len(bans)}  forbidden={len(FORBIDDEN)}  permitted={len(PERMITTED)}  "
          f"live prose={len(txt)} chars  facts={len(eb.get('facts',[]))}")
    print("RESULT:", "ALL CONSISTENT" if bad == 0 else f"{bad} PROBLEM(S)")
    sys.exit(0 if bad == 0 else 1)

if __name__ == "__main__":
    main()
