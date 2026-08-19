#!/usr/bin/env python3
"""Classify domains by who runs their DNS, so "parked" can be separated from "live".

WHY THIS EXISTS. The owner's question — "which of my domains aren't parked?" — is
answerable without any registrar API, because nameserver delegation is public. The
registrar (Nominet EPP) is needed only to ENUMERATE the domains; once you have a
list, this classifies it from public DNS alone.

WHAT IT DOES NOT DO. It reports the delegation, not whether a site serves. A domain
can be on Cloudflare and still serve nothing (measured: apis.uk and ugg2.com are on
our own Cloudflare account with a worker route and have no site behind them at all).
So `--check-http` is offered separately, and the two answers are kept in separate
columns rather than blended into one "parked" verdict — a blended verdict is the one
that would be wrong in both directions.

Reads one domain per line on stdin (blank lines and #-comments ignored), or a file
given as an argument. Queries Cloudflare's DNS-over-HTTPS resolver, so it does not
depend on the local resolver, which is a real hazard: a negative answer cached
locally is indistinguishable from "no such record" (LANDMINES, 2026-08-18).

Output: TSV — domain, class, nameservers, [http status, bytes] with --check-http.
"""
import argparse
import concurrent.futures as cf
import json
import sys
import urllib.request

DOH = "https://1.1.1.1/dns-query?name={}&type={}"
UA = "Mozilla/5.0 (compatible; portfolio-domain-audit/1.0)"

# Patterns are matched against the whole nameserver hostname, lowercased.
# PARKING is deliberately broad: a false "parked" is cheap to check by hand,
# a false "live" hides a domain the owner wanted to see.
# MEASURED on this estate 2026-08-19 (152 portfolio domains) rather than guessed:
# 124 sat on ns1/ns2.dan.com, 3 on aftermarket.com, 1 on namepros-dns, 1 on
# domainlore.co.uk. Marketplace/for-sale DNS is the dominant parking shape here,
# not the classic parking-ad services — so those are listed first and the rest are
# kept as a wider net for domains outside the sample.
PARKING = [
    # marketplace / for-sale DNS (all four seen live on this estate)
    "dan.com", "aftermarket.com", "namepros-dns", "domainlore",
    "sedoparking", "sedo.com", "undeveloped.com", "afternic", "namedrive",
    "sav.com", "efty.com", "squadhelp", "brandbucket",
    # parking-ad networks
    "parkingcrew", "bodis.com", "above.com", "parklogic", "voodoo.com",
    "fabulous.com", "parkingpage", "cashparking",
    # registrar defaults — a domain left on these is unconfigured, which for the
    # owner's question ("which aren't parked?") usually means the same thing.
    # They get their OWN class (REGISTRAR_DEFAULT, checked before PARKED) rather
    # than being folded in, because it is not the same claim: a registrar's
    # default nameservers can also front a perfectly real site.
    "domaincontrol.com", "registrar-servers.com", "dynadot.com",
    "porkbun.com", "spaceship", "namesilo", "hostinger",
]
REGISTRAR_DEFAULT = [
    "domaincontrol.com", "registrar-servers.com", "dynadot.com",
    "porkbun.com", "spaceship", "namesilo", "hostinger",
]
CLOOK = ["clook"]
CLOUDFLARE = ["ns.cloudflare.com"]


def doh(name, rtype):
    req = urllib.request.Request(
        DOH.format(name, rtype),
        headers={"accept": "application/dns-json", "user-agent": UA},
    )
    with urllib.request.urlopen(req, timeout=15) as r:
        return json.load(r)


def classify(domain):
    try:
        d = doh(domain, "NS")
    except Exception as e:  # network/DoH failure is NOT "no nameservers"
        return domain, "ERROR", f"lookup failed: {type(e).__name__}"
    status = d.get("Status")
    if status == 3:
        return domain, "NXDOMAIN", ""
    ns = sorted(
        a.get("data", "").rstrip(".").lower()
        for a in d.get("Answer", [])
        if a.get("type") == 2
    )
    if not ns:
        # No NS in the answer section can mean the name exists but the resolver
        # returned the authority section instead. Report it as UNKNOWN rather
        # than inventing "unparked".
        return domain, "NO_NS" if status == 0 else f"STATUS_{status}", ""
    joined = ",".join(ns)
    low = joined.lower()
    if any(p in low for p in CLOUDFLARE):
        cls = "CLOUDFLARE"
    elif any(p in low for p in CLOOK):
        cls = "CLOOK"
    elif any(p in low for p in REGISTRAR_DEFAULT):
        cls = "REGISTRAR_DEFAULT"
    elif any(p in low for p in PARKING):
        cls = "PARKED"
    else:
        cls = "OTHER"
    return domain, cls, joined


def http_probe(domain):
    """Read the BODY size, never just the status: a parked domain answers 200 on
    every path (LANDMINES). Size + status together are what discriminate."""
    for url in (f"https://{domain}/", f"http://{domain}/"):
        try:
            req = urllib.request.Request(url, headers={"user-agent": UA})
            with urllib.request.urlopen(req, timeout=20) as r:
                body = r.read(200_000)
                return r.status, len(body)
        except Exception:
            continue
    return 0, 0


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("file", nargs="?", help="file of domains; default stdin")
    ap.add_argument("--check-http", action="store_true",
                    help="also fetch each domain and report status + body bytes")
    ap.add_argument("--workers", type=int, default=12)
    args = ap.parse_args()

    src = open(args.file) if args.file else sys.stdin
    domains = [
        ln.strip().lower() for ln in src
        if ln.strip() and not ln.strip().startswith("#")
    ]
    seen, ordered = set(), []
    for d in domains:
        if d not in seen:
            seen.add(d)
            ordered.append(d)

    hdr = ["domain", "class", "nameservers"]
    if args.check_http:
        hdr += ["http", "bytes"]
    print("\t".join(hdr))

    counts = {}
    with cf.ThreadPoolExecutor(max_workers=args.workers) as ex:
        results = list(ex.map(classify, ordered))
        http = dict(zip(ordered, ex.map(http_probe, ordered))) if args.check_http else {}

    for domain, cls, ns in results:
        counts[cls] = counts.get(cls, 0) + 1
        row = [domain, cls, ns]
        if args.check_http:
            st, by = http.get(domain, (0, 0))
            row += [str(st), str(by)]
        print("\t".join(row))

    print("\n# totals: " + "  ".join(
        f"{k}={v}" for k, v in sorted(counts.items(), key=lambda kv: -kv[1])
    ), file=sys.stderr)
    print(f"# domains in: {len(domains)}  unique: {len(ordered)}", file=sys.stderr)


if __name__ == "__main__":
    main()
