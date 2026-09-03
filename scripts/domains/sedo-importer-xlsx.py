#!/usr/bin/env python3
"""Generate a Sedo Domain Importer spreadsheet (.xlsx) from registrar CSVs.

The output mimics Sedo's own template (Example_File_Domain_Importer.xlsx,
decoded 2026-09-02 — sedo_domain_management RUNBOOK §6): one sheet named
Sheet1, seven columns, text via sharedStrings exactly as their example
encodes it, prices as native numeric cells.

    Domain Name | Selling Option | For Sale | Price | Minimum Price | Currency | Action Type

Inputs are the domain_valuation lane's inbound CSVs (first column = domain,
header row skipped) plus, later, the valuation lane's OUTPUT_prices file.
Without prices every row defaults to MAKE_OFFER / yes / blanks — the agreed
interim shape; prices arrive as a second import.

Usage:
  sedo-importer-xlsx.py --self-test
  sedo-importer-xlsx.py build --out SHEET.xlsx [--csv-out SHEET.csv]
      [--provenance-out PROV.csv]
      --domains a.csv [--domains b.csv ...]
      [--prices OUTPUT_prices.csv]
      [--exclude-file live_domains.txt [--exclude-file other_reasons.txt ...]]
      [--owner-authorized-buy-now-prices]

BUY-NOW PRICES ARE HARD-BLOCKED BY DEFAULT (owner ruling, 2026-09-03).
Any row that would carry `selling_option=BUY_NOW` or a non-empty `price`
(from --prices) stops the build with every blocked domain+price listed,
UNLESS --owner-authorized-buy-now-prices is passed. That flag exists for
the owner to use himself, deliberately, on a run he has actually
authorized — a session must never pass it on its own initiative, on
inference from a conversation, or as a default. `min_price` (a floor
under MAKE_OFFER) is NOT gated — the whole point of a minimum is
protecting against a lowball, the opposite of what this guard exists to
prevent.

Multiple --exclude-file are UNIONED — keep separate files per REASON (e.g.
one for live-site protection, a distinct one for an owner-requested
withdrawal) rather than appending unrelated domains into a fence file
whose name states a specific reason; a future regeneration of one file
must not silently lose the other's exclusions.

The prices CSV is mapped by HEADER NAME (afternic-csv.py's lesson: never
positional): domain, selling_option, price, min_price, currency, and either
forsale or keep_or_sell. Unknown headers are reported, never interpreted.
A domain that fails ACE-form validation stops the build — refuse, don't guess.
"""

import argparse
import csv
import io
import os
import re
import sys
import tempfile
import zipfile
from xml.sax.saxutils import escape

HEADERS = ["Domain Name", "Selling Option", "For Sale", "Price",
           "Minimum Price", "Currency", "Action Type"]
DOMAIN_RE = re.compile(r"^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$")
SELLING_OPTIONS = {"MAKE_OFFER", "BUY_NOW", ""}
CURRENCIES = {"EUR", "USD", "GBP", ""}

NS_CLASSES = [  # provenance only; the fence itself is the --exclude-file
    ("afternic_aftermarket", re.compile(r"afternic|aftermarket|dan\.com|atom\.com")),
    ("cloudflare", re.compile(r"cloudflare")),
    ("sedo", re.compile(r"sedopark")),
]


def ns_class(ns):
    for name, rx in NS_CLASSES:
        if rx.search(ns):
            return name
    return "other" if ns else "none"


def read_domains(paths):
    """-> ordered {domain: (source_file, ns_class)}; first occurrence wins."""
    out, bad = {}, []
    for path in paths:
        with open(path, newline="", encoding="utf-8") as fh:
            rows = csv.reader(fh)
            header = next(rows, None)
            if header is None:
                raise SystemExit(f"sedo-importer: empty file: {path}")
            if header and header[0].strip().strip('"').lower() != "domain":
                raise SystemExit(
                    f"sedo-importer: {path} first column is {header[0]!r}, expected 'domain'")
            for row in rows:
                if not row or not row[0].strip():
                    continue
                d = row[0].strip().lower()
                try:
                    d.encode("ascii")
                    ok = bool(DOMAIN_RE.match(d))
                except UnicodeEncodeError:
                    ok = False  # not ACE form — Sedo wants punycode, not raw IDN
                if not ok:
                    bad.append((path, d))
                    continue
                ns = row[3].strip().lower() if len(row) > 3 else ""
                out.setdefault(d, (os.path.basename(path), ns_class(ns)))
    if bad:
        for path, d in bad:
            print(f"sedo-importer: REJECTED domain {d!r} from {path}", file=sys.stderr)
        raise SystemExit(f"sedo-importer: {len(bad)} invalid domain(s); fix the input")
    return out


def read_exclude_files(paths):
    """Union of one-domain-per-line files -> lowercased set."""
    out = set()
    for p in paths:
        with open(p, encoding="utf-8") as fh:
            out |= {l.strip().lower() for l in fh if l.strip()}
    return out


def read_prices(path):
    """-> {domain: row-dict}, mapped by header name only."""
    with open(path, newline="", encoding="utf-8") as fh:
        rows = csv.DictReader(fh)
        headers = set(rows.fieldnames or [])
        known = {"domain", "selling_option", "price", "min_price", "currency",
                 "forsale", "keep_or_sell", "category", "valuation",
                 "valuation_method", "confidence"}
        unmapped = headers - known
        if unmapped:
            print(f"sedo-importer: prices file has unmapped headers, ignored: "
                  f"{sorted(unmapped)}", file=sys.stderr)
        if "domain" not in headers:
            raise SystemExit("sedo-importer: prices file has no 'domain' column")
        out = {}
        for r in rows:
            d = (r.get("domain") or "").strip().lower()
            if d:
                out[d] = r
    return out


def sheet_rows(domains, prices, exclude):
    """-> (rows in HEADERS order, excluded list, provenance list)."""
    rows, excluded, prov = [], [], []
    for d in sorted(domains):
        src, nsc = domains[d]
        if d in exclude:
            excluded.append(d)
            continue
        opt, forsale, price, minprice, cur = "MAKE_OFFER", "yes", "", "", ""
        p = prices.get(d)
        if p:
            opt = (p.get("selling_option") or opt).strip().upper()
            if opt not in SELLING_OPTIONS:
                raise SystemExit(f"sedo-importer: bad selling_option {opt!r} for {d}")
            cur = (p.get("currency") or "").strip().upper()
            if cur not in CURRENCIES:
                raise SystemExit(f"sedo-importer: bad currency {cur!r} for {d}")
            price = (p.get("price") or "").strip()
            minprice = (p.get("min_price") or "").strip()
            for label, v in (("price", price), ("min_price", minprice)):
                if v and not re.match(r"^\d+(\.\d+)?$", v):
                    raise SystemExit(f"sedo-importer: bad {label} {v!r} for {d}")
            if "forsale" in p and p["forsale"] not in (None, ""):
                forsale = "yes" if p["forsale"].strip().lower() in ("yes", "y", "1", "true") else "no"
            elif (p.get("keep_or_sell") or "").strip().lower() == "keep":
                forsale = "no"
        rows.append([d, opt, forsale, price, minprice, cur, ""])
        prov.append([d, src, nsc])
    return rows, excluded, prov


def find_buy_now(rows):
    """-> [(domain, selling_option, price), ...] for every row carrying
    BUY_NOW or a non-empty price. `min_price` alone does NOT trigger this
    — a floor under MAKE_OFFER is protective, not a listed asking price."""
    return [(r[0], r[1], r[3]) for r in rows if r[1] == "BUY_NOW" or r[3]]


def enforce_buy_now_gate(rows, authorized):
    """Hard-block BUY_NOW/priced rows unless explicitly authorized. Refuses
    (SystemExit) rather than silently dropping the price or the domain —
    a silent drop would let a session believe pricing shipped when it
    didn't, which is its own kind of unsafe default."""
    hits = find_buy_now(rows)
    if hits and not authorized:
        lines = "\n".join(f"  {d}: {opt or '(no selling_option)'} "
                           f"price={p or '(none)'}" for d, opt, p in hits)
        raise SystemExit(
            f"sedo-importer: REFUSED — {len(hits)} row(s) carry a BUY_NOW "
            f"selling option or a non-empty price, and "
            f"--owner-authorized-buy-now-prices was not passed:\n{lines}\n"
            f"This flag is for the OWNER to use himself, deliberately, on a "
            f"run he has actually authorized — never pass it on inference "
            f"or as a default. min_price (a floor) is unaffected by this "
            f"gate; only BUY_NOW/price are blocked.")
    if hits and authorized:
        print(f"sedo-importer: OWNER-AUTHORIZED BUY-NOW PRICES — "
              f"{len(hits)} row(s) shipping with a set price:", file=sys.stderr)
        for d, opt, p in hits:
            print(f"  {d}: {opt} price={p}", file=sys.stderr)
    return hits


# ---- xlsx writing: sharedStrings for text (as Sedo's template), numbers native

def col_letter(i):
    s = ""
    i += 1
    while i:
        i, r = divmod(i - 1, 26)
        s = chr(65 + r) + s
    return s


def write_xlsx(path, rows):
    strings, index = [], {}

    def sref(s):
        if s not in index:
            index[s] = len(strings)
            strings.append(s)
        return index[s]

    body = []
    for rn, row in enumerate([HEADERS] + rows, start=1):
        cells = []
        for cn, val in enumerate(row):
            if val == "":
                continue
            ref = f"{col_letter(cn)}{rn}"
            if rn > 1 and re.match(r"^\d+(\.\d+)?$", val):
                cells.append(f'<c r="{ref}"><v>{val}</v></c>')
            else:
                cells.append(f'<c r="{ref}" t="s"><v>{sref(val)}</v></c>')
        body.append(f'<row r="{rn}">' + "".join(cells) + "</row>")

    ns = 'xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"'
    sheet = (f'<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
             f'<worksheet {ns}><sheetData>' + "".join(body) + "</sheetData></worksheet>")
    sst = (f'<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
           f'<sst {ns} count="{len(strings)}" uniqueCount="{len(strings)}">'
           + "".join(f"<si><t>{escape(s)}</t></si>" for s in strings) + "</sst>")
    workbook = (f'<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
                f'<workbook {ns} xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">'
                f'<sheets><sheet name="Sheet1" sheetId="1" r:id="rId1"/></sheets></workbook>')
    wb_rels = ('<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
               '<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">'
               '<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>'
               '<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings" Target="sharedStrings.xml"/>'
               '<Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>'
               '</Relationships>')
    root_rels = ('<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
                 '<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">'
                 '<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>'
                 '</Relationships>')
    styles = (f'<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
              f'<styleSheet {ns}><fonts count="1"><font><sz val="12"/><name val="Calibri"/></font></fonts>'
              f'<fills count="1"><fill><patternFill patternType="none"/></fill></fills>'
              f'<borders count="1"><border/></borders>'
              f'<cellStyleXfs count="1"><xf/></cellStyleXfs>'
              f'<cellXfs count="1"><xf xfId="0"/></cellXfs></styleSheet>')
    ctypes = ('<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
              '<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">'
              '<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>'
              '<Default Extension="xml" ContentType="application/xml"/>'
              '<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>'
              '<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>'
              '<Override PartName="/xl/sharedStrings.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sharedStrings+xml"/>'
              '<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>'
              '</Types>')

    with zipfile.ZipFile(path, "w", zipfile.ZIP_DEFLATED) as z:
        z.writestr("[Content_Types].xml", ctypes)
        z.writestr("_rels/.rels", root_rels)
        z.writestr("xl/workbook.xml", workbook)
        z.writestr("xl/_rels/workbook.xml.rels", wb_rels)
        z.writestr("xl/styles.xml", styles)
        z.writestr("xl/sharedStrings.xml", sst)
        z.writestr("xl/worksheets/sheet1.xml", sheet)


def read_back(path):
    """Independent re-read of a written xlsx -> list of rows (self-test's check)."""
    import xml.etree.ElementTree as ET
    m = "{http://schemas.openxmlformats.org/spreadsheetml/2006/main}"
    with zipfile.ZipFile(path) as z:
        sst_root = ET.fromstring(z.read("xl/sharedStrings.xml"))
        strings = ["".join(t.text or "" for t in si.iter(f"{m}t"))
                   for si in sst_root.findall(f"{m}si")]
        sheet = ET.fromstring(z.read("xl/worksheets/sheet1.xml"))
    rows = []
    for row in sheet.iter(f"{m}row"):
        cells = {}
        for c in row.iter(f"{m}c"):
            ref, t = c.get("r"), c.get("t")
            v = c.find(f"{m}v")
            val = v.text if v is not None else ""
            if t == "s" and val != "":
                val = strings[int(val)]
            cells[re.match(r"[A-Z]+", ref).group()] = val
        rows.append([cells.get(col_letter(i), "") for i in range(len(HEADERS))])
    return rows


def self_test():
    fails = 0

    def check(name, cond):
        nonlocal fails
        print(f"  {'PASS' if cond else 'FAIL'} {name}")
        fails += 0 if cond else 1

    with tempfile.TemporaryDirectory() as td:
        src = os.path.join(td, "in.csv")
        with open(src, "w", encoding="utf-8") as fh:
            fh.write("domain,expiry,auto_renew,nameservers\n"
                     "bbb.example.com,2027-01-01,yes,ns1.afternic.com\n"
                     "aaa.example.com,2027-01-01,yes,ns1.cloudflare.com\n"
                     "bbb.example.com,2027-01-01,yes,ns1.afternic.com\n")
        prices = os.path.join(td, "p.csv")
        with open(prices, "w", encoding="utf-8") as fh:
            fh.write("domain,selling_option,price,min_price,currency,keep_or_sell\n"
                     "bbb.example.com,BUY_NOW,150,100,GBP,sell\n")
        doms = read_domains([src])
        check("dedupe: 2 unique from 3 rows", len(doms) == 2)
        rows, excluded, prov = sheet_rows(doms, read_prices(prices),
                                          {"aaa.example.com"})
        check("fence excludes aaa", excluded == ["aaa.example.com"])
        check("one sheet row remains", len(rows) == 1)
        check("priced row carries BUY_NOW/150/100/GBP",
              rows[0][1:6] == ["BUY_NOW", "yes", "150", "100", "GBP"])
        check("provenance has ns class", prov[0][2] == "afternic_aftermarket")

        out = os.path.join(td, "out.xlsx")
        write_xlsx(out, rows)
        back = read_back(out)
        check("round-trip header exact", back[0] == HEADERS)
        check("round-trip row exact", back[1] == rows[0] + [""] * (len(HEADERS) - len(rows[0])) or back[1][:7] == rows[0])

        defaults, _, _ = sheet_rows({"zzz.example.com": ("x", "none")}, {}, set())
        check("default row is MAKE_OFFER/yes/blanks",
              defaults[0] == ["zzz.example.com", "MAKE_OFFER", "yes", "", "", "", ""])

        ex1 = os.path.join(td, "ex1.txt")
        ex2 = os.path.join(td, "ex2.txt")
        with open(ex1, "w", encoding="utf-8") as fh:
            fh.write("aaa.example.com\n")
        with open(ex2, "w", encoding="utf-8") as fh:
            fh.write("bbb.example.com\n")
        check("two exclude files union (via the actual reader function)",
              read_exclude_files([ex1, ex2]) == {"aaa.example.com", "bbb.example.com"})

        bad = os.path.join(td, "bad.csv")
        with open(bad, "w", encoding="utf-8") as fh:
            fh.write("domain,expiry\nnot a domain!!,2027-01-01\n")
        try:
            read_domains([bad])
            check("invalid domain rejected", False)
        except SystemExit:
            check("invalid domain rejected", True)

        # BUY-NOW gate: prove it actually blocks (not just that the happy
        # path works) and prove authorization actually lifts it — two
        # separate assertions, since a gate that never fires and a gate
        # that always fires both look like "the flag exists".
        buy_now_rows = [["ccc.example.com", "BUY_NOW", "yes", "9999", "", "GBP", ""]]
        try:
            enforce_buy_now_gate(buy_now_rows, authorized=False)
            check("BUY_NOW blocked without authorization", False)
        except SystemExit:
            check("BUY_NOW blocked without authorization", True)
        try:
            enforce_buy_now_gate(buy_now_rows, authorized=True)
            check("BUY_NOW allowed WITH authorization", True)
        except SystemExit:
            check("BUY_NOW allowed WITH authorization", False)

        priced_no_buy_now = [["ddd.example.com", "MAKE_OFFER", "yes", "500", "", "GBP", ""]]
        try:
            enforce_buy_now_gate(priced_no_buy_now, authorized=False)
            check("MAKE_OFFER with a set price also blocked", False)
        except SystemExit:
            check("MAKE_OFFER with a set price also blocked", True)

        min_price_only = [["eee.example.com", "MAKE_OFFER", "yes", "", "100", "GBP", ""]]
        try:
            enforce_buy_now_gate(min_price_only, authorized=False)
            check("min_price ALONE is not gated (it's a floor, not an ask)", True)
        except SystemExit:
            check("min_price ALONE is not gated (it's a floor, not an ask)", False)

    print(("self-test PASS" if fails == 0 else f"self-test FAIL ({fails})"))
    return 1 if fails else 0


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--self-test", action="store_true")
    ap.add_argument("mode", nargs="?", choices=["build"])
    ap.add_argument("--out")
    ap.add_argument("--csv-out")
    ap.add_argument("--provenance-out")
    ap.add_argument("--domains", action="append", default=[])
    ap.add_argument("--prices")
    ap.add_argument("--exclude-file", action="append", default=[])
    ap.add_argument("--owner-authorized-buy-now-prices", action="store_true")
    args = ap.parse_args()

    if args.self_test:
        sys.exit(self_test())
    if args.mode != "build" or not args.out or not args.domains:
        ap.error("need: build --out X.xlsx --domains a.csv [...]")

    exclude = read_exclude_files(args.exclude_file)
    prices = read_prices(args.prices) if args.prices else {}
    domains = read_domains(args.domains)
    rows, excluded, prov = sheet_rows(domains, prices, exclude)
    enforce_buy_now_gate(rows, args.owner_authorized_buy_now_prices)

    write_xlsx(args.out, rows)
    if args.csv_out:
        with open(args.csv_out, "w", newline="", encoding="utf-8") as fh:
            w = csv.writer(fh)
            w.writerow(HEADERS)
            w.writerows(rows)
    if args.provenance_out:
        with open(args.provenance_out, "w", newline="", encoding="utf-8") as fh:
            w = csv.writer(fh)
            w.writerow(["domain", "source_file", "ns_class"])
            w.writerows(prov)

    back = read_back(args.out)  # verify the artefact, not the intent
    assert back[0] == HEADERS and len(back) == len(rows) + 1, "round-trip mismatch"
    # defense in depth: re-check the GATE against the WRITTEN artefact, not
    # just the in-memory rows the gate already saw — catches any future
    # code path that reaches write_xlsx() without going through sheet_rows()
    written_hits = find_buy_now(back[1:])
    if written_hits and not args.owner_authorized_buy_now_prices:
        raise SystemExit(
            f"sedo-importer: INTERNAL GUARD FAILURE — the written artefact "
            f"{args.out} carries {len(written_hits)} BUY_NOW/priced row(s) "
            f"despite no authorization. The file has already been written; "
            f"delete it and treat this as a bug in the gate, not the input.")
    priced = sum(1 for r in rows if r[3])
    print(f"sedo-importer: wrote {args.out}: {len(rows)} domains "
          f"({priced} priced, {len(rows) - priced} default MAKE_OFFER/no-price); "
          f"excluded {len(excluded)}: {', '.join(excluded) if excluded else '-'}")


if __name__ == "__main__":
    main()
