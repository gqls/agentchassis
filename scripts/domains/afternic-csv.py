#!/usr/bin/env python3
"""Ingest an Afternic portfolio export (CSV) into a dated snapshot + report.

WHY THIS EXISTS. Afternic has no self-serve seller API (verified 2026-09-02:
their documented APIs are registrar-partner-only; the seller-facing automation
is an in-dashboard agent). The owner chose a no-credential CSV loop: he exports
his portfolio from the Afternic dashboard, drops the file in the lane's
inbound/ directory, and this parses it — so sessions reason from a parsed,
validated snapshot instead of a pasted dashboard.

THE RULE THIS MECHANISES (WRONG_CALLS 2026-07-28). A dashboard paste with 11
header columns and 2 value cells was read positionally and produced a false
"Minimum Offer = 0" that reached five documents. So: columns are mapped by
HEADER NAME, never by position; a row whose cell count disagrees with the
header is refused loudly, never padded; unmapped headers are reported, not
guessed; and --control lets a known value (e.g. relojistas floor) validate the
whole mapping.

Usage:
  afternic-csv.py ingest <export.csv> [--out-dir DIR] [--known FILE]
                  [--control DOMAIN:FIELD:VALUE ...] [--baseline FILE|auto]
  afternic-csv.py valuation-csv <snapshot.json> [--out FILE] [--currency CUR]
  afternic-csv.py --self-test

  valuation-csv writes the normalised feed the domain_valuation lane asked
  for (2026-09-02): domain,price,currency,status,price_source — price is
  buy_now, else floor, else min_offer; price_source says which, because a
  floor is not an asking price and a valuation should know the difference.

  --known    file of domains (one per line) to cross-check presence against,
             e.g. the estate's sites table or a registrar enumeration.
  --control  assert a value you know from the dashboard; a mismatch means the
             HEADER MAPPING is wrong, so it fails the ingest (repeatable).
  --baseline a previous snapshot JSON to diff against; 'auto' picks the newest
             snapshot already in --out-dir.

Output: report to stdout; snapshot JSON to <out-dir>/portfolio_<date>.json.
Exit nonzero on refusal (no domain column, malformed rows, control mismatch).
"""
import argparse
import csv
import datetime as dt
import json
import pathlib
import re
import sys

LANE_DIR = pathlib.Path(__file__).resolve().parents[2] / (
    "docs/agent_docs/docs024_key_docs_latest/afternic_domain_management")

# Canonical field <- header aliases. Aliases are compared after lowercasing and
# stripping every non-alphanumeric character, so "Buy It Now ($)", "buy_it_now"
# and "BuyItNow" all land together. Built from the headers seen in the owner's
# 2026-07-28 dashboard paste plus Afternic's own bulk/portfolio vocabulary;
# the first REAL export locks this list — record any additions in NOTES.
ALIASES = {
    "domain":       {"domain", "domainname", "domains"},
    "status":       {"status", "listingstatus", "auctionstatus"},
    "buy_now":      {"buynow", "buyitnow", "buynowprice", "buyitnowprice",
                     "bin", "asking", "askingprice", "price"},
    "floor":        {"floor", "floorprice", "reserve", "reserveprice"},
    "min_offer":    {"minimumoffer", "minoffer", "minimumofferprice"},
    "lander":       {"salelander", "lander", "landertype"},
    "fast_transfer": {"fasttransfer", "fasttransferstatus", "ftstatus"},
    "lease_to_own": {"leasetoown", "lto", "leasetoownstatus"},
    "views":        {"views", "pageviews"},
    "leads":        {"leads", "offers", "inquiries", "enquiries"},
    "searches_30d": {"30daysearches", "searches30day", "searcheslast30days",
                     "30dsearches"},
    "verified":     {"verified", "ownershipverified", "verification",
                     "verificationstatus"},
    "date_listed":  {"datelisted", "listeddate", "dateadded", "added"},
}
PRICE_FIELDS = {"buy_now", "floor", "min_offer"}
COUNT_FIELDS = {"views", "leads", "searches_30d"}


def canon(header):
    return re.sub(r"[^a-z0-9]", "", header.lower())


def map_headers(headers):
    """Return (index -> canonical field or None, unmapped header list)."""
    mapping, unmapped = {}, []
    for i, h in enumerate(headers):
        c = canon(h)
        for field, names in ALIASES.items():
            if c in names:
                # first occurrence wins; a duplicate header is reported
                if field in mapping.values():
                    unmapped.append(f"{h!r} (duplicate of mapped field {field})")
                else:
                    mapping[i] = field
                break
        else:
            if c:  # ignore genuinely empty header cells
                unmapped.append(repr(h))
    return mapping, unmapped


def parse_price(raw):
    """'$12,000' -> 12000.0; ''/None/'-' -> None; junk -> ValueError."""
    if raw is None:
        return None
    s = raw.strip().replace("$", "").replace(",", "").replace("£", "")
    if s in ("", "-", "—", "N/A", "n/a"):
        return None
    return float(s)


def parse_export(path):
    """Parse the CSV. Returns (rows, unmapped, malformed).

    rows: list of dicts with canonical fields + 'extras' (unmapped columns,
    kept verbatim under their original headers). malformed: (line_no, reason)
    for every row REFUSED — cell count differing from the header count is a
    refusal, not a best-effort parse: that is the 2026-07-28 failure shape.
    """
    with open(path, newline="", encoding="utf-8-sig") as f:
        reader = csv.reader(f)
        try:
            headers = next(reader)
        except StopIteration:
            raise SystemExit(f"REFUSED: {path} is empty")
        mapping, unmapped = map_headers(headers)
        if "domain" not in mapping.values():
            raise SystemExit(
                "REFUSED: no recognisable domain column in header "
                f"{headers!r} — if the export renamed it, extend ALIASES and "
                "record the new header in the lane NOTES")
        rows, malformed = [], []
        for n, cells in enumerate(reader, start=2):
            if not any(c.strip() for c in cells):
                continue
            if len(cells) != len(headers):
                malformed.append(
                    (n, f"{len(cells)} cells against {len(headers)} headers"))
                continue
            row, extras = {}, {}
            for i, cell in enumerate(cells):
                field = mapping.get(i)
                if field is None:
                    if headers[i].strip():
                        extras[headers[i]] = cell
                    continue
                if field in PRICE_FIELDS:
                    try:
                        row[field] = parse_price(cell)
                    except ValueError:
                        malformed.append((n, f"unparseable {field}: {cell!r}"))
                        row[field] = None
                elif field in COUNT_FIELDS:
                    try:
                        row[field] = int(cell.replace(",", "")) if cell.strip() else None
                    except ValueError:
                        row[field] = None
                else:
                    row[field] = cell.strip()
            row["domain"] = row.get("domain", "").lower().rstrip(".")
            if not row["domain"]:
                malformed.append((n, "empty domain cell"))
                continue
            row["extras"] = extras
            rows.append(row)
    return rows, unmapped, malformed


def check_controls(rows, controls):
    """controls: list of 'domain:field:value'. Returns list of failures."""
    by_domain = {r["domain"]: r for r in rows}
    failures = []
    for spec in controls:
        try:
            domain, field, want = spec.split(":", 2)
        except ValueError:
            failures.append(f"bad --control spec {spec!r} (want DOMAIN:FIELD:VALUE)")
            continue
        row = by_domain.get(domain.lower())
        if row is None:
            failures.append(f"control domain {domain} not in export")
            continue
        got = row.get(field)
        if field in PRICE_FIELDS:
            ok = got is not None and abs(got - float(want)) < 0.005
        else:
            ok = str(got) == want
        if not ok:
            failures.append(
                f"control MISMATCH {domain}.{field}: export says {got!r}, "
                f"you said {want!r} — the header mapping is suspect, fix it "
                "before trusting any figure in this report")
    return failures


def diff_snapshots(old_rows, new_rows):
    old = {r["domain"]: r for r in old_rows}
    new = {r["domain"]: r for r in new_rows}
    changes = {"added": sorted(new.keys() - old.keys()),
               "removed": sorted(old.keys() - new.keys()),
               "changed": []}
    watched = ["status", "buy_now", "floor", "min_offer", "lander", "verified"]
    for d in sorted(new.keys() & old.keys()):
        delta = {f: (old[d].get(f), new[d].get(f))
                 for f in watched if old[d].get(f) != new[d].get(f)}
        if delta:
            changes["changed"].append({"domain": d, **{
                f: f"{a!r} -> {b!r}" for f, (a, b) in delta.items()}})
    return changes


def report(rows, unmapped, malformed, known, controls_failed, changes):
    out = []
    out.append(f"rows parsed: {len(rows)}")
    if malformed:
        out.append(f"rows REFUSED (never guessed at): {len(malformed)}")
        for n, why in malformed[:20]:
            out.append(f"  line {n}: {why}")
    if unmapped:
        out.append(f"unmapped headers (kept in extras, not interpreted): "
                   f"{', '.join(unmapped)}")
    by_status = {}
    for r in rows:
        by_status[r.get("status") or "(no status column)"] = \
            by_status.get(r.get("status") or "(no status column)", 0) + 1
    out.append("by status: " + ", ".join(
        f"{k}={v}" for k, v in sorted(by_status.items(), key=lambda kv: -kv[1])))
    for f in sorted(PRICE_FIELDS):
        have = sum(1 for r in rows if r.get(f) is not None)
        out.append(f"{f} set on {have}/{len(rows)}")
    leads = [(r["domain"], r["leads"]) for r in rows if r.get("leads")]
    if leads:
        top = sorted(leads, key=lambda t: -t[1])[:10]
        out.append("domains with leads: " + str(len(leads)) + " — top: "
                   + ", ".join(f"{d}({n})" for d, n in top))
    if known is not None:
        in_export = {r["domain"] for r in rows}
        missing = sorted(known - in_export)
        out.append(f"known-list cross-check: {len(known & in_export)}/{len(known)} "
                   f"present in export; missing: "
                   + (", ".join(missing[:30]) if missing else "none")
                   + (" …" if len(missing) > 30 else ""))
    for f in controls_failed:
        out.append("CONTROL FAIL: " + f)
    if changes is not None:
        out.append(f"vs baseline: +{len(changes['added'])} added, "
                   f"-{len(changes['removed'])} removed, "
                   f"{len(changes['changed'])} changed")
        for c in changes["changed"][:20]:
            out.append("  " + c["domain"] + ": " + ", ".join(
                f"{k}={v}" for k, v in c.items() if k != "domain"))
        for label in ("added", "removed"):
            if changes[label]:
                shown = ", ".join(changes[label][:15])
                out.append(f"  {label}: {shown}"
                           + (" …" if len(changes[label]) > 15 else ""))
    return "\n".join(out)


def cmd_ingest(args):
    rows, unmapped, malformed = parse_export(args.export)
    known = None
    if args.known:
        known = {ln.strip().lower() for ln in open(args.known)
                 if ln.strip() and not ln.startswith("#")}
    controls_failed = check_controls(rows, args.control or [])
    out_dir = pathlib.Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    changes = None
    baseline = args.baseline
    if baseline == "auto":
        prev = sorted(out_dir.glob("portfolio_*.json"))
        baseline = str(prev[-1]) if prev else None
    if baseline:
        changes = diff_snapshots(json.load(open(baseline)), rows)
    print(report(rows, unmapped, malformed, known, controls_failed, changes))
    snap = out_dir / f"portfolio_{dt.date.today().isoformat()}.json"
    if snap.exists():
        snap = out_dir / (snap.stem + dt.datetime.now().strftime("_%H%M") + ".json")
    snap.write_text(json.dumps(rows, indent=1, sort_keys=True))
    print(f"snapshot: {snap}")
    if controls_failed or malformed:
        sys.exit(1)


def fmt_price(p):
    return "" if p is None else (str(int(p)) if p == int(p) else str(p))


def write_valuation_csv(rows, out_path, currency):
    """The domain_valuation lane's feed (requested 2026-09-02): one price per
    domain — buy_now, else floor, else min_offer — with price_source naming
    which, because a floor is not an asking price and the valuation should
    never mistake one for the other."""
    out_path = pathlib.Path(out_path)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    with open(out_path, "w", newline="") as f:
        w = csv.writer(f)
        w.writerow(["domain", "price", "currency", "status", "price_source"])
        for r in sorted(rows, key=lambda r: r["domain"]):
            for src in ("buy_now", "floor", "min_offer"):
                if r.get(src) is not None:
                    price = r[src]
                    break
            else:
                price, src = None, "none"
            w.writerow([r["domain"], fmt_price(price),
                        currency if price is not None else "",
                        r.get("status") or "", src])
    return out_path


def cmd_valuation(args):
    rows = json.load(open(args.snapshot))
    out = args.out or (
        pathlib.Path(__file__).resolve().parents[2]
        / "docs/agent_docs/docs024_key_docs_latest/domain_valuation/inbound"
        / f"afternic_listings_{dt.date.today().isoformat()}.csv")
    print(f"wrote {write_valuation_csv(rows, out, args.currency)}")


def self_test():
    """Offline; proves the MECHANICS (mapping, refusal, control, diff).
    It cannot lock the real export's headers — only the first real file does
    that — so a pass here is not evidence the live format parses."""
    import tempfile
    fails = 0

    def t(name, ok):
        nonlocal fails
        print(f"  {'PASS' if ok else 'FAIL'} {name}")
        fails += (not ok)

    with tempfile.TemporaryDirectory() as td:
        td = pathlib.Path(td)
        good = td / "good.csv"
        good.write_text(
            "Domain Name,Status,Buy It Now,Floor Price,Minimum Offer,"
            "Sale Lander,Views,Leads,30-day Searches,Mystery Column\n"
            "Relojistas.com,Listed,\"$25,000\",\"$12,000\",\"$5,000\","
            "Custom,7,0,12,huh\n"
            "example.co.uk,In Verification,,,,Default,0,0,,x\n")
        rows, unmapped, malformed = parse_export(good)
        t("2 rows parse", len(rows) == 2)
        t("domain lowercased", rows[0]["domain"] == "relojistas.com")
        t("floor $12,000 -> 12000.0", rows[0]["floor"] == 12000.0)
        t("empty price -> None", rows[1]["buy_now"] is None)
        t("unknown header reported not guessed",
          any("Mystery" in u for u in unmapped))
        t("extras keep unmapped cell", rows[0]["extras"].get("Mystery Column") == "huh")

        # THE 2026-07-28 SHAPE: 10 headers, 3 value cells. Must be refused.
        short = td / "short.csv"
        short.write_text(good.read_text().split("\n")[0] + "\n"
                         "relojistas.com,Listed,0\n")
        rows2, _, malformed2 = parse_export(short)
        t("short row REFUSED, not padded", len(rows2) == 0 and len(malformed2) == 1)
        t("refusal names the count mismatch", "3 cells against 10" in malformed2[0][1])

        ok = check_controls(rows, ["relojistas.com:floor:12000"])
        t("control passes on true value", ok == [])
        bad = check_controls(rows, ["relojistas.com:floor:0"])
        t("control fails on false value", len(bad) == 1 and "MISMATCH" in bad[0])

        changed = [dict(r) for r in rows]
        changed[0]["floor"] = 15000.0
        d = diff_snapshots(rows, changed[:1])
        t("diff sees price change", d["changed"] and "floor" in d["changed"][0])
        t("diff sees removal", d["removed"] == ["example.co.uk"])

        # header-alias variants: bulk-template-style names land on same fields
        alt = td / "alt.csv"
        alt.write_text("domain,buy_now_price,minimum_offer\nx.com,100,50\n")
        rows3, _, _ = parse_export(alt)
        t("alias variants map", rows3[0]["buy_now"] == 100.0
          and rows3[0]["min_offer"] == 50.0)

        # valuation feed: BIN wins, floor is the fallback, source is named
        val = write_valuation_csv(rows + [{"domain": "floor-only.com",
                                           "floor": 500.0, "status": "Listed"}],
                                  td / "val.csv", "USD")
        lines = val.read_text().splitlines()
        t("valuation header", lines[0] == "domain,price,currency,status,price_source")
        t("BIN preferred + integral price",
          "relojistas.com,25000,USD,Listed,buy_now" in lines)
        t("floor fallback named as floor",
          "floor-only.com,500,USD,Listed,floor" in lines)
        t("no price -> empty cells + source none",
          "example.co.uk,,,In Verification,none" in lines)

    print("self-test:", "PASS" if fails == 0 else f"{fails} FAILURES")
    sys.exit(1 if fails else 0)


def main():
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--self-test", action="store_true")
    sub = ap.add_subparsers(dest="cmd")
    ing = sub.add_parser("ingest")
    ing.add_argument("export")
    ing.add_argument("--out-dir", default=str(LANE_DIR / "snapshots"))
    ing.add_argument("--known")
    ing.add_argument("--control", action="append")
    ing.add_argument("--baseline", default="auto")
    val = sub.add_parser("valuation-csv")
    val.add_argument("snapshot")
    val.add_argument("--out")
    val.add_argument("--currency", default="USD")
    args = ap.parse_args()
    if args.self_test:
        self_test()
    elif args.cmd == "ingest":
        cmd_ingest(args)
    elif args.cmd == "valuation-csv":
        cmd_valuation(args)
    else:
        ap.print_help()
        sys.exit(1)


if __name__ == "__main__":
    main()
