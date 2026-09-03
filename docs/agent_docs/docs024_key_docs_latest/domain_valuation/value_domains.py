#!/usr/bin/env python3
"""Value every domain in WORKING_table.csv and assign a tier.

Every number it produces is decomposed into named parts on the row, so any
figure can be argued with. Re-runnable: as more appraisals land, re-run and
the anchors improve.

WORKING CURRENCY IS USD — both Afternic and Sedo default to it and every
appraisal is USD. The owner's £150 transfer-away floor converts at a rate he
should confirm; FLOOR_USD below is that conversion, marked as an assumption.

Method, in order:
 1. ANCHOR. Dynappraisal where we have one (source=own). Otherwise the median
    appraisal of its category × subcategory, else of its category (source=
    category-median, lower confidence). Nothing else is trusted as a level:
    Afternic asks are measured at ~5.4x appraisal (2026-09-03) so they are a
    consistency input, never an anchor.
 2. QUALITY multipliers, transparent and small. Each is a stated view about
    what a buyer pays for, not a fitted parameter — we have NO realised sales
    to fit against (Spaceship's sold report is empty all-time), so pretending
    to precision would be false.
 3. TIER from the adjusted value, then a KEEN price for anything sold: priced
    to move, never below the floor.
"""
import csv, os, re, statistics, collections

HERE = os.path.dirname(os.path.abspath(__file__))
FLOOR_GBP = 150            # owner's transfer-away fee, 2026-08-17
GBP_USD = 1.27             # [ASSUMED] owner to confirm
FLOOR_USD = round(FLOOR_GBP * GBP_USD)

# Tier bands on adjusted value (USD). Keen price = the fraction of the anchor
# we would actually ask to move it, by tier: better names hold value, tail
# names sell on price.
TIERS = [
    ('A', 10000, 0.60),
    ('B',  3500, 0.55),
    ('C',  1200, 0.50),
    ('D',   500, 0.45),
    ('E',     0, 0.40),
]

# Categories whose commercial demand is thin on the evidence we have.
# Marked, not auto-sold — the cut is by whole blocks and is the owner's.
THIN_EVIDENCE = {'foreign-language', 'names-places', 'generic-word', 'misc',
                 'packaging-print'}


def quality(row):
    """Return (multiplier, [reasons]). Multiplicative, each stated."""
    d = row['domain']
    sld = d.split('.')[0]
    tld = row['tld']
    m, why = 1.0, []

    hyphens = sld.count('-')
    if hyphens == 1:
        m *= 0.75; why.append('hyphen -25%')
    elif hyphens >= 2:
        m *= 0.55; why.append(f'{hyphens} hyphens -45%')

    words = len([w for w in re.split(r'[-]', sld) if w])
    if len(sld) >= 22:
        m *= 0.80; why.append('very long -20%')
    elif len(sld) <= 6 and hyphens == 0 and tld == 'com':
        m *= 1.20; why.append('short .com +20%')

    if re.search(r'\d', sld):
        m *= 0.80; why.append('digits -20%')

    # TLD: .com is the liquid one; UK TLDs sell to a UK buyer only, and .uk
    # trails .co.uk in the UK market despite being shorter.
    tld_mult = {'com': 1.0, 'co.uk': 0.85, 'uk': 0.70, 'org.uk': 0.45,
                'me.uk': 0.35, 'net': 0.55, 'org': 0.55, 'biz': 0.30,
                'info': 0.30, 'io': 1.0, 'ai': 1.0, 'shop': 0.35,
                'club': 0.30, 'cv': 0.25, 'vin': 0.25, 'us': 0.40}.get(tld, 0.40)
    if tld_mult != 1.0:
        m *= tld_mult; why.append(f'{tld} x{tld_mult}')

    if row['category'] in THIN_EVIDENCE:
        m *= 0.75; why.append('thin-demand category -25%')

    if row['trademark_flag']:
        m *= 0.50; why.append('trademark risk -50%')

    return m, '; '.join(why)


def main():
    rows = list(csv.DictReader(open(os.path.join(HERE, 'WORKING_table.csv'))))
    appr = {r["domain"]: float(r["dynappraisal"]) for r in rows
            if r["dynappraisal"] and r["dynappraisal"].replace(".", "").isdigit()}

    def med(vals):
        return statistics.median(vals) if vals else None

    sub_med, cat_med = {}, {}
    by_sub, by_cat = collections.defaultdict(list), collections.defaultdict(list)
    for r in rows:
        if r['domain'] in appr:
            by_sub[(r['category'], r['subcategory'])].append(appr[r['domain']])
            by_cat[r['category']].append(appr[r['domain']])
    for k, v in by_sub.items():
        if len(v) >= 3: sub_med[k] = med(v)
    for k, v in by_cat.items():
        if len(v) >= 3: cat_med[k] = med(v)
    global_med = med(list(appr.values())) or 1000.0

    out = []
    for r in rows:
        d = r['domain']
        if d in appr and r.get('appraisal_kind') == 'proxy':
            # The .com equivalent's value, applied to a UK name. The TLD
            # multiplier below already discounts it to the UK market, but the
            # keyword itself may be worth more or less here than in .com, so
            # this never counts as a direct measurement.
            anchor, src, conf = appr[d], f'proxy-via-{r.get("appraisal_proxy_domain","?")}', 'medium'
        elif d in appr:
            anchor, src, conf = appr[d], 'own-appraisal', 'high'
        elif (r['category'], r['subcategory']) in sub_med:
            anchor, src, conf = sub_med[(r['category'], r['subcategory'])], 'subcategory-median', 'medium'
        elif r['category'] in cat_med:
            anchor, src, conf = cat_med[r['category']], 'category-median', 'low'
        else:
            anchor, src, conf = global_med, 'portfolio-median', 'very-low'

        mult, why = quality(r)
        value = anchor * mult
        # A withdrawn or live-site name still gets a value (the owner may want
        # to know what he is holding) but can never be priced for sale.
        # A domain we do not own can never be priced, listed or counted as
        # stock — checked FIRST, because it outranks every other reason.
        if r['registrar'].startswith('NOT-OWNED'):
            keen_out, sell = '', r['registrar']
        elif r.get('keep_override'):
            keen_out, sell = '', 'KEEP:' + r['keep_override']
        else:
            keen_out, sell = None, 'tbd'
        for tier, floor, keen_frac in TIERS:
            if value >= floor:
                break
        keen = keen_out if keen_out == '' else max(round(value * keen_frac), FLOOR_USD)

        ask = float(r['live_ask_price']) if r['live_ask_price'] else None
        floor_now = float(r['afternic_min_offer']) if r['afternic_min_offer'] else None
        out.append(dict(r,
                        anchor=f'{anchor:.0f}', anchor_source=src, confidence=conf,
                        quality_multiplier=f'{mult:.2f}', quality_reasons=why,
                        value_usd=f'{value:.0f}', tier=tier, keen_price_usd=str(keen),
                        sale_status=sell,
                        ask_vs_value=f'{ask/value:.1f}' if ask and value else '',
                        floor_vs_keen=f'{floor_now/keen:.1f}' if floor_now and keen else ''))

    cols = list(out[0].keys())
    path = os.path.join(HERE, 'VALUATION_2026-09-03_draft.csv')
    with open(path, 'w', newline='') as fh:
        w = csv.DictWriter(fh, fieldnames=cols)
        w.writeheader(); w.writerows(out)

    print(f'{path}: {len(out)} rows')
    tc = collections.Counter(r['tier'] for r in out)
    cc = collections.Counter(r['confidence'] for r in out)
    print('tiers:', dict(sorted(tc.items())))
    print('confidence:', dict(cc))
    tot_val = sum(float(r['value_usd']) for r in out)
    sellable = [r for r in out if r['keen_price_usd']]
    tot_keen = sum(float(r['keen_price_usd']) for r in sellable)
    held = len(out) - len(sellable)
    print(f'portfolio value (USD): {tot_val:,.0f}')
    print(f'{len(sellable)} sellable; if all sold keen: {tot_keen:,.0f} '
          f'({held} held back: live-site or owner-withdrawn)')
    print('\nby category (value desc):')
    bycat = collections.defaultdict(list)
    for r in out: bycat[r['category']].append(float(r['value_usd']))
    for c, vs in sorted(bycat.items(), key=lambda x: -statistics.median(x[1])):
        print(f'  {c:22s} n={len(vs):4d} median={statistics.median(vs):7.0f} total={sum(vs):9.0f}')


if __name__ == '__main__':
    main()
