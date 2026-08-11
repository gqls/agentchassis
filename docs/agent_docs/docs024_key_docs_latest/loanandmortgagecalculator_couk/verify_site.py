#!/usr/bin/env python3
"""verify_site.py — assert the site is correct, on disk and then AGAINST THE LIVE ORIGIN.

WHY THIS EXISTS, AND WHY IT DEFAULTS TO --live.

The first version of this site passed every pre-launch check and still shipped three
dead URLs on all 42 pages, in the sitemap, and in three self-referencing canonicals.
Two instrument faults caused that, and both are the reason this file is shaped the way
it is:

  1. Verification ran against `python3 -m http.server`, which RESOLVES DIRECTORY
     INDEXES. Production is a Cloudflare worker over a B2 bucket, which cannot — it
     maps {hostname}{path} to an object key. The local server was a MORE FORGIVING
     server than production, so "/loans/" passed locally and 404'd live.
  2. When the first link checker flagged "/loans/" as dead, it was "fixed" to resolve
     "/loans/" -> "loans/index.html". That taught the instrument the same forgiveness
     and turned a TRUE POSITIVE into silence.

So: --live is the real check and it is the default. --disk is a fast pre-flight only,
and it deliberately does NOT resolve directory indexes either, so the two modes agree
about what a valid reference is.

The other lesson encoded here: every check asserts VALIDITY, not PRESENCE. The
pre-launch run confirmed a canonical existed on 43 of 43 pages and that ld+json was
present — never that the canonical resolved or that the JSON parsed. Both were broken.

Run:  python3 verify_site.py            # live (default)
      python3 verify_site.py --disk     # offline pre-flight
"""
import glob
import hashlib
import json
import os
import re
import sys
import urllib.error
import urllib.request
from urllib.parse import urldefrag, urljoin

SITE = os.path.expanduser("~/projects/sites/loanandmortgagecalculator.co.uk")
BASE = "https://loanandmortgagecalculator.co.uk"
LIVE = "--disk" not in sys.argv

# Cloudflare's zone-level "Managed robots.txt" prepends content-signal directives to
# whatever we upload, so the served bytes legitimately differ from the repo file. Every
# other zone in the fleet does the same; our own rules and Sitemap: line survive at the
# tail. Verified 2026-07-31. This is the ONLY sanctioned byte difference.
BYTE_EXEMPT = {"robots.txt"}

# ── decomposed pages and the per-page og:* tags ────────────────────────────────
# An ASSEMBLED page is built from page_components rows through the shared chrome,
# not from a hand-built document. The shared <head> cannot carry per-page
# og:title/og:description/og:url, so those three are lost the moment a page is
# decomposed. That is a STATED, ACCEPTED loss, not a regression discovered here:
# PLAN_2026-08-05 §6 lists it beside nav aria-current and lang="en-GB" -> "en",
# "each is visible in the assemble-mirror diff and none is silent". Assembly
# injects JSON-LD and rel=canonical per page instead.
#
# Without this exemption the og:url check fails on every decomposed page for
# ever — 2 pages on 2026-08-06, 19 after Track A, all 59 eventually — and a
# checker that is permanently red is a checker nobody reads. So: still a hard
# FAIL on a hand-built page (where the tag is real and its absence is a defect),
# and a counted NOTE on an assembled one.
#
# The marker is the skip-link target the shared header carries
# (`<span id="content" tabindex="-1">`), not anything content-specific: a
# hand-built page wraps its content in `<div id="content" class="container…">`.
# It therefore identifies a decomposed TOOL page (Track B) just as well as a
# prose one, which a `ported-prose` marker would not.
ASSEMBLED_MARKER = '<span id="content" tabindex="-1">'
OG_PER_PAGE = ('property="og:url"',)

EXTERNAL = ("http://", "https://", "mailto:", "tel:", "#", "data:", "javascript:")
fails, checks = [], []
n_assembled_og = 0


def fail(check, detail):
    fails.append((check, detail))


def note(check, detail):
    checks.append((check, detail))


def fetch(url):
    """Return (status, bytes). Never raises for an HTTP error status."""
    req = urllib.request.Request(url, headers={"User-Agent": "verify_site/1.0"})
    try:
        with urllib.request.urlopen(req, timeout=30) as r:
            return r.status, r.read()
    except urllib.error.HTTPError as e:
        return e.code, e.read() if e.fp else b""
    except Exception as e:                                    # noqa: BLE001
        return 0, str(e).encode()


def resolves(ref):
    """Is this internal reference fetchable?

    --disk models EXACTLY ONE rewrite, because the worker performs exactly one:
    `path === '/' || path === ''` -> '/index.html' (worker.js:9-11). Nothing else.
    In particular "/loans/" is NOT resolved to "loans/index.html", because the worker
    does not do that either — inventing that resolution here is precisely the
    instrument fault that let three dead URLs ship.
    """
    if LIVE:
        return fetch(BASE + ref)[0] == 200
    if ref in ("/", ""):
        ref = "/index.html"
    return os.path.isfile(os.path.join(SITE, ref.lstrip("/")))


os.chdir(SITE)
pages = sorted(glob.glob("**/*.html", recursive=True))
note("pages found", len(pages))

# ── 1. every internal reference resolves ─────────────────────────────────────────
refs = {}
for f in pages:
    body = open(f, encoding="utf-8").read()
    for m in re.finditer(r'(?:href|src)="([^"]+)"', body):
        u = m.group(1)
        if u.startswith(EXTERNAL):
            continue
        refs.setdefault(urldefrag(urljoin("/" + f, u))[0], set()).add(f)

dead = {u: v for u, v in refs.items() if not resolves(u)}
note("distinct internal references", len(refs))
for u, srcs in sorted(dead.items()):
    fail("dead internal reference", f"{u}  (referenced by {len(srcs)} page(s))")

# ── 2. every sitemap URL resolves, and the sitemap matches the page set ──────────
sm = open("sitemap.xml", encoding="utf-8").read()
locs = re.findall(r"<loc>([^<]+)</loc>", sm)
note("sitemap URLs", len(locs))
for loc in locs:
    if not loc.startswith(BASE):
        fail("sitemap URL is off-domain", loc)
    elif not resolves(loc[len(BASE):] or "/"):
        fail("sitemap URL does not resolve", loc)

# a page that exists but is absent from the sitemap is invisible to search
listed = {loc[len(BASE):] for loc in locs}
for f in pages:
    if f == "404.html":
        continue
    p = "/" + f
    if p not in listed and not (f == "index.html" and "/" in listed):
        fail("page missing from sitemap", p)

# ── 3. every canonical resolves AND names the page it is on ─────────────────────
for f in pages:
    body = open(f, encoding="utf-8").read()
    m = re.search(r'<link rel="canonical" href="([^"]+)"', body)
    if not m:
        fail("no canonical", f)
        continue
    can = m.group(1)
    if not can.startswith(BASE):
        fail("canonical is off-domain", f"{f} -> {can}")
        continue
    path = can[len(BASE):] or "/"
    if not resolves(path):
        fail("canonical does not resolve", f"{f} -> {can}")
    expect = "/" if f == "index.html" else "/" + f
    if path != expect:
        fail("canonical names another page", f"{f} -> {can} (expected {expect})")

# ── 4. every ld+json block PARSES (presence is not validity) ────────────────────
n_ld = 0
for f in pages:
    for m in re.finditer(r'<script type="application/ld\+json">(.*?)</script>',
                         open(f, encoding="utf-8").read(), re.S):
        n_ld += 1
        try:
            json.loads(m.group(1))
        except ValueError as e:
            fail("ld+json does not parse", f"{f}: {e}")
note("ld+json blocks", n_ld)

# ── 5. head essentials, and no leakage of either source site ───────────────────
for f in pages:
    body = open(f, encoding="utf-8").read()
    assembled = ASSEMBLED_MARKER in body
    for needle, what in (("<title>", "title"), ('property="og:url"', "og:url"),
                         ('class="skip-link"', "skip link"), ("<footer>", "footer")):
        if needle not in body:
            if assembled and needle in OG_PER_PAGE:
                n_assembled_og += 1      # accepted loss — see OG_PER_PAGE above
            else:
                fail(f"missing {what}", f)
    # "LoanAndMortgageCalculator.co.uk" CONTAINS "MortgageCalculator.co.uk", so the
    # exclusion has to be case-insensitive or it reports the old brand on every page.
    for m in re.finditer(r"(?i)(?<![a-z])(mortgagecalculator\.co\.uk|loancalculator\.co\.uk)",
                         body):
        if "loanandmortgagecalculator" not in body[max(0, m.start() - 20):m.end()].lower():
            fail("old-domain reference", f"{f}: …{body[max(0, m.start()-30):m.end()+10]}…")
    for bad in ("cloudflareinsights", "nav-placeholder", "nav.js"):
        if bad in body:
            fail("leaked artefact from a source site", f"{f}: {bad}")

# ── 6. live only: every file byte-identical to the repo ────────────────────────
if LIVE:
    same = diff = 0
    for f in sorted(glob.glob("**/*", recursive=True)):
        if not os.path.isfile(f) or f.startswith(".git"):
            continue
        status, got = fetch(f"{BASE}/{f}")
        if status != 200:
            fail("file does not serve", f"{f} -> HTTP {status}")
            continue
        want = open(f, "rb").read()
        if hashlib.sha256(want).digest() == hashlib.sha256(got).digest():
            same += 1
        elif f in BYTE_EXEMPT:
            note(f"{f} differs live (sanctioned)",
                 f"repo {len(want)}B vs live {len(got)}B — Cloudflare Managed robots.txt")
        else:
            diff += 1
            fail("live bytes differ from repo", f"{f}: repo {len(want)}B vs live {len(got)}B")
    note("files byte-identical live", f"{same} (+{len(BYTE_EXEMPT)} sanctioned exempt)")

# Counted, never silent: the number is what makes the accepted loss visible as it
# spreads. If this is 0 while decomposed pages exist, the marker has moved and the
# exemption is inert — check ASSEMBLED_MARKER against a live assembled page.
if n_assembled_og:
    note("per-page og:* dropped (assembled)",
         f"{n_assembled_og} page(s) — accepted loss, PLAN_2026-08-05 §6")

# ── report ─────────────────────────────────────────────────────────────────────
print(f"\n{'LIVE' if LIVE else 'DISK'} verification of {BASE}\n")
for k, v in checks:
    print(f"  {k:<38} {v}")
print()
if fails:
    for k, v in fails:
        print(f"  FAIL  {k}: {v}")
    print(f"\n{len(fails)} FAILURE(S)")
    sys.exit(1)
print("  all checks pass")
