#!/usr/bin/env python3
"""A1 — re-shape the SDLT register to one fact per threshold/rate.

Every `quote` below is LIFTED PROGRAMMATICALLY from the visible text the real Go
extractor produces (quotecheck/main.go, which imports
datahelpers.VisibleTextFromHTML) — nothing here is retyped, because this
channel rewrites some characters on emission and a retyped quote is how a
day-one `citation_lost` happens.

The script REFUSES to emit SQL if any quote is not found by
datahelpers.QuoteFoundInText against the live page.
"""
import json
import re
import subprocess
import sys

RATES = "https://www.gov.uk/stamp-duty-land-tax/residential-property-rates"
HIGHER = "https://www.gov.uk/guidance/stamp-duty-land-tax-buying-an-additional-residential-property"

QUOTECHECK = "/home/ant/.claude-scratch/claude-1000/-home-ant-projects-agentchassis/0370b809-4dc6-42a1-969f-a5c75688642e/scratchpad/quotecheck"


def visible_text(url):
    out = subprocess.run(["go", "run", ".", url, "dump"], cwd=QUOTECHECK,
                         capture_output=True, text=True, check=True)
    return out.stdout


def lift(text, pattern, label):
    """Return the single match of `pattern` in `text`.

    The patterns below use `.` where the source has a currency symbol or a
    curly apostrophe, so that NOTHING non-ASCII is typed into this file — the
    characters come from the fetched page. Exactly one match is required, so a
    page edit that duplicates or drops a phrase fails loudly rather than
    silently lifting the wrong copy.
    """
    hits = re.findall(pattern, text)
    if len(hits) != 1:
        sys.exit(f"quote {label!r}: {len(hits)} matches for {pattern!r}, need exactly 1")
    return hits[0]


rates_text = visible_text(RATES)
higher_text = visible_text(HIGHER)

# --- lift each quote out of the fetched page --------------------------------
q_nil = lift(rates_text, r"Up to .125,000 Zero", "nil band")
q_2pc = lift(rates_text, r"The next .125,000 \(the portion from .125,001 to .250,000\) 2%", "2% band")
q_5pc = lift(rates_text, r"The next .675,000 \(the portion from .250,001 to .925,000\) 5%", "5% band")
q_10pc = lift(rates_text, r"The next .575,000 \(the portion from .925,001 to .1\.5 million\) 10%", "10% band")
q_12pc = lift(rates_text, r"The remaining amount \(the portion above .1\.5 million\) 12%", "12% band")
q_ftb_nil = lift(rates_text, r"no SDLT up to .300,000", "ftb nil")
q_ftb_5pc = lift(rates_text, r"5% SDLT on the portion from .300,001 to .500,000", "ftb 5%")
q_ftb_cap = lift(rates_text, r"If the price is over .500,000, you cannot claim the relief\.", "ftb cap")
q_surch = lift(rates_text, r"You.ll usually have to pay 5% on top of SDLT rates if buying a new residential property means you.ll own more than one\.", "surcharge")
q_floor = lift(higher_text, r"You must pay the higher Stamp Duty Land Tax \(SDLT\) rates when you buy a residential property \(or a part of one\) for .40,000 or more", "surcharge floor")

P = chr(163)  # pound sign, constructed rather than typed (emission-rewrite trap)
TODAY = "2026-08-10"


def fact(fid, unit, claim, value, quote, url, title, writer_line=None):
    f = {
        "id": fid,
        "kind": "metric",
        "unit": unit,
        "claim": claim,
        "value": value,
        "source": {"citation": {
            "url": url,
            "quote": quote,
            "title": title,
            "accessed": TODAY,
            "published": "",
            "publisher": "GOV.UK",
        }},
        "verified_at": TODAY,
        "staleness_days": 365,
    }
    if writer_line:
        f["writer_line"] = writer_line
    return f


T_RATES = "Stamp Duty Land Tax: Residential property rates"
T_HIGHER = "Stamp Duty Land Tax: buying an additional residential property"

facts = [
    # ---- standard residential bands: one fact per band EDGE and per RATE ----
    fact("sdlt-standard-nil-band-upper", "GBP",
         f"No SDLT is due on the portion of the price up to {P}125,000 (standard residential rates)",
         125000, q_nil, RATES, T_RATES,
         f"Standard stamp duty is nothing on the first {P}{{value}} of the price"),
    fact("sdlt-standard-rate-125k-250k", "percent",
         f"2% applies to the portion of the price from {P}125,001 to {P}250,000",
         2, q_2pc, RATES, T_RATES,
         f"{{value}}% is charged on the portion from {P}125,001 to {P}250,000"),
    fact("sdlt-standard-band-250k-upper", "GBP",
         f"The 2% band ends at {P}250,000 — above it the 5% band begins",
         250000, q_2pc, RATES, T_RATES),
    fact("sdlt-standard-rate-250k-925k", "percent",
         f"5% applies to the portion of the price from {P}250,001 to {P}925,000",
         5, q_5pc, RATES, T_RATES,
         f"{{value}}% is charged on the portion from {P}250,001 to {P}925,000"),
    fact("sdlt-standard-band-925k-upper", "GBP",
         f"The 5% band ends at {P}925,000 — above it the 10% band begins",
         925000, q_5pc, RATES, T_RATES),
    fact("sdlt-standard-rate-925k-1500k", "percent",
         f"10% applies to the portion of the price from {P}925,001 to {P}1.5 million",
         10, q_10pc, RATES, T_RATES,
         f"{{value}}% is charged on the portion from {P}925,001 to {P}1.5 million"),
    fact("sdlt-standard-band-1500k-upper", "GBP",
         f"The 10% band ends at {P}1.5 million — above it the top rate applies",
         1500000, q_10pc, RATES, T_RATES),
    fact("sdlt-standard-top-rate", "percent",
         f"12% applies to the portion of the price above {P}1.5 million",
         12, q_12pc, RATES, T_RATES,
         f"{{value}}% is charged on the portion above {P}1.5 million"),
    # ---- first-time buyer relief ------------------------------------------
    fact("sdlt-ftb-nil-band-upper", "GBP",
         f"First-time buyers pay no SDLT on the portion of the price up to {P}300,000",
         300000, q_ftb_nil, RATES, T_RATES,
         f"First-time buyers pay no stamp duty up to {P}{{value}}"),
    fact("sdlt-ftb-rate-300k-500k", "percent",
         f"First-time buyers pay 5% on the portion of the price from {P}300,001 to {P}500,000",
         5, q_ftb_5pc, RATES, T_RATES,
         f"First-time buyers pay {{value}}% on the portion from {P}300,001 to {P}500,000"),
    fact("sdlt-ftb-relief-cap", "GBP",
         f"First-time buyer relief cannot be claimed at all if the price is over {P}500,000 — standard rates apply to the whole purchase",
         500000, q_ftb_cap, RATES, T_RATES,
         f"Above {P}{{value}} first-time buyer relief disappears entirely — the standard bands apply to the whole price"),
    # ---- additional property ----------------------------------------------
    fact("sdlt-additional-surcharge", "percent",
         "Buying an additional residential property usually adds 5 percentage points on top of each SDLT band",
         5, q_surch, RATES, T_RATES,
         "An additional residential property usually costs {value} percentage points on top of each stamp duty band"),
    fact("sdlt-additional-surcharge-floor", "GBP",
         f"The higher (additional property) rates apply to purchases of {P}40,000 or more",
         40000, q_floor, HIGHER, T_HIGHER,
         f"The additional-property surcharge applies to purchases of {P}{{value}} or more"),
]

# --- refuse to emit unless every quote verifies against the live page -------
by_url = {}
for f in facts:
    by_url.setdefault(f["source"]["citation"]["url"], []).append(f["source"]["citation"]["quote"])

failed = False
for url, quotes in by_url.items():
    args = ["go", "run", ".", url, "check"] + quotes
    r = subprocess.run(args, cwd=QUOTECHECK, capture_output=True, text=True)
    sys.stderr.write(f"--- {url}\n{r.stdout}")
    if r.returncode != 0:
        failed = True
if failed:
    sys.exit("REFUSING to emit: at least one quote is not found by the real Go verifier")

print(json.dumps(facts, ensure_ascii=False, indent=2))
