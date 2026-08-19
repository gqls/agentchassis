#!/usr/bin/env python3
"""Pick domains to RESERVE as register-free test domains.

THE ASK (owner, 2026-08-19): *"leave me a bunch of domains (50?) unregistered as
test domains that I can run without the register."* — i.e. domains deliberately
given NO positioning-register entry, so builds can be run on them without the
register influencing (or being polluted by) the result.

WHY THIS IS NOT "pick 50 at random". Once the registry covers the whole estate
(owner ruling 2026-08-19), a domain with no entry normally means *nobody has got
to it yet*. A reserved test domain must therefore be EXPLICITLY reserved, or the
next person filling gaps will helpfully write an entry and silently destroy the
control. Absent-because-reserved and absent-because-pending must not look alike.

THE SAFETY RULE, and it is the whole reason this is a script and not a shuffle:
the register exists to stop our own sites competing with each other in search. An
unregistered test site is exactly the thing that could collide with a positioned
one. So a candidate must be:

  1. OWNED and resolving      — NXDOMAIN may mean the registration lapsed.
  2. PARKED, not live         — never build over a serving site.
  3. NOT named in the register — not as a primary, not as a twin, not anywhere.
  4. NOT a near-variant of a registered domain — "savings-rates" vs "savingsrates"
     is the same search result; a test build there competes with the real one.

Rule 4 is the one a human picking by eye gets wrong, and it is why the tool
compares normalised labels (punctuation and TLD stripped) rather than strings.

Inputs:
  --inventory  file of all owned domains, one per line (from Nominet/registrars)
  --register   REGISTER_positioning.md (default: the lane's copy)
  --classified TSV from classify_nameservers.py (optional; enables rules 1-2)
  -n           how many to pick (default 50)

Output: TSV — domain, verdict, reason. Only ELIGIBLE rows are candidates; the
rest are printed with the reason they were excluded, because "why not this one?"
is the question that gets asked next.
"""
import argparse
import os
import re
import sys

DEFAULT_REGISTER = os.path.join(
    "docs", "agent_docs", "docs024_key_docs_latest", "portfolio_positioning",
    "REGISTER_positioning.md")

SAFE_NS_CLASSES = {"PARKED", "REGISTRAR_DEFAULT"}
UNOWNED_CLASSES = {"NXDOMAIN"}


def norm(domain):
    """Normalise to a comparable label: drop the TLD, then punctuation.

    savings-rates.co.uk and savingsrates.uk both become 'savingsrates', which is
    the point — they are the same phrase to a search engine and to a reader."""
    label = domain.strip().lower().split(".")[0]
    return re.sub(r"[^a-z0-9]", "", label)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--inventory", required=True)
    ap.add_argument("--register", default=DEFAULT_REGISTER)
    ap.add_argument("--classified")
    ap.add_argument("-n", type=int, default=50)
    ap.add_argument("--show-excluded", action="store_true")
    args = ap.parse_args()

    reg_text = open(args.register, encoding="utf-8").read().lower()
    # Every domain-shaped token in the register, however it is mentioned.
    reg_domains = set(re.findall(r"\b[a-z0-9][a-z0-9-]*\.(?:co\.uk|org\.uk|me\.uk|uk|com|net|org|io|ai)\b",
                                 reg_text))
    reg_norms = {norm(d) for d in reg_domains}

    ns_class = {}
    if args.classified:
        with open(args.classified, encoding="utf-8") as f:
            next(f, None)
            for ln in f:
                p = ln.rstrip("\n").split("\t")
                if len(p) >= 2:
                    ns_class[p[0].strip().lower()] = p[1].strip()

    domains = []
    with open(args.inventory, encoding="utf-8") as f:
        for ln in f:
            ln = ln.strip().lower()
            if ln and not ln.startswith("#"):
                domains.append(ln)

    print("domain\tverdict\treason")
    eligible, counts = [], {}

    def emit(d, v, why):
        counts[v] = counts.get(v, 0) + 1
        if v == "ELIGIBLE" or args.show_excluded:
            print(f"{d}\t{v}\t{why}")

    for d in domains:
        n = norm(d)
        if d in reg_domains:
            emit(d, "EXCLUDED", "named in the register")
            continue
        if n in reg_norms:
            clash = sorted(x for x in reg_domains if norm(x) == n)[:2]
            emit(d, "EXCLUDED", f"near-variant of registered {', '.join(clash)}")
            continue
        cls = ns_class.get(d)
        if cls is None and args.classified:
            emit(d, "EXCLUDED", "not in the classification file")
            continue
        if cls in UNOWNED_CLASSES:
            emit(d, "EXCLUDED", f"{cls} — may not be registered")
            continue
        if cls and cls not in SAFE_NS_CLASSES:
            emit(d, "EXCLUDED", f"nameservers are {cls} — not parked, may be serving")
            continue
        eligible.append(d)
        emit(d, "ELIGIBLE", f"unregistered, {cls or 'unclassified'}")

    # Spread the pick across first letters so the reserved set is not 50 domains
    # about one subject — a single-vertical test set only ever tests one shape.
    eligible.sort()
    picked, seen_initial = [], {}
    for d in eligible:
        k = norm(d)[:3]
        if seen_initial.get(k, 0) < 2:
            picked.append(d)
            seen_initial[k] = seen_initial.get(k, 0) + 1
        if len(picked) >= args.n:
            break
    for d in eligible:
        if len(picked) >= args.n:
            break
        if d not in picked:
            picked.append(d)

    print(f"\n# eligible: {len(eligible)}   picked: {len(picked)} of {args.n} asked for",
          file=sys.stderr)
    print("# " + "  ".join(f"{k}={v}" for k, v in sorted(counts.items())), file=sys.stderr)
    if len(picked) < args.n:
        print(f"# ⚠ SHORT BY {args.n - len(picked)} — the inventory does not contain enough "
              f"unregistered, parked domains. Do NOT make up the difference by releasing "
              f"registered ones without an explicit decision per domain.", file=sys.stderr)
    if picked:
        print("\n# --- the reserved set ---", file=sys.stderr)
        for d in picked:
            print(d, file=sys.stderr)


if __name__ == "__main__":
    main()
