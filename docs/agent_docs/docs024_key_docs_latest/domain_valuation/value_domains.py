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
# OWNER RULING 2026-09-03: "keep USD, round it up." Prices round UP only, so
# rounding can move a price away from the floor but never through it.
# £150 x 1.27 = $190.50, rounded up to a clean $200.
FLOOR_USD = 200


def round_up_clean(x):
    """Round UP to the next tidy marketable figure. The step widens with size
    so a $215 name is not rounded to $500, and a $12,000 one does not carry a
    spurious $12,050."""
    import math
    for limit, step in ((500, 25), (2000, 50), (10000, 250), (50000, 1000)):
        if x <= limit:
            return int(math.ceil(x / step) * step)
    return int(math.ceil(x / 5000) * 5000)

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
# ⚠ 'generic-word' was in this set until 2026-09-03 and that was BACKWARDS.
# A single English dictionary word is the most valuable class in a domain
# portfolio, not a thin-demand one: it was penalising cartoon.co.uk -25% (the
# owner paid £5,000+ for it) and 52 others. Removed on evidence, not taste.
THIN_EVIDENCE = {'foreign-language', 'names-places', 'misc', 'packaging-print'}

# Categories the owner keeps WHOLE for an advertising-network play, overriding
# the sub-category cut unit. See NETWORK_KEEP_categories.txt for the reasoning.
try:
    NETWORK_KEEP = {l.split('#')[0].strip() for l in
                    open(os.path.join(HERE, 'NETWORK_KEEP_categories.txt'))
                    if l.split('#')[0].strip()}
except OSError:
    NETWORK_KEEP = set()

# Domains the owner has given a real figure for — a realised sale, what he paid,
# or a floor he set. These are the ONLY per-domain numbers in the estate that
# did not come from an algorithm, and every one measured so far is far above
# what the model produces. Recorded in prose was not enough: holidaytime.com
# sold for $12,000 and still sat in the sell cut at $450, because no scarcity
# rule happened to catch a two-word compound. So the file is now DATA.
def _owner_figures():
    import csv as _csv
    out = {}
    try:
        with open(os.path.join(HERE, 'OWNER_FIGURES.csv')) as fh:
            for row in _csv.DictReader(fh):
                d = row['domain'].strip().lower()
                if d:
                    out[d] = row
    except OSError:
        pass
    return out


OWNER_FIGURES = _owner_figures()

# A single dictionary word with no hyphen or digit. These are the estate's
# premium end and the model cannot price them: only 4 of 144 have an appraisal,
# and the two we have owner figures for (cartoon.co.uk £5,000+ paid,
# free.co.uk "sold for a lot of money") are both far above anything the model
# would produce. So they are HELD OUT of the keen tail for the owner's eye
# rather than auto-priced — the same treatment as a name that looks bought.
try:
    _DICT = {w.strip().lower() for w in open('/usr/share/dict/british-english')
             if "'" not in w and len(w.strip()) > 2}
except OSError:
    _DICT = set()


def is_single_word(domain):
    sld = domain.split('.')[0].lower()
    return ('-' not in sld and not any(c.isdigit() for c in sld)
            and sld in _DICT)


# TLD: .com is the liquid one; UK TLDs sell to a UK buyer only, and .uk trails
# .co.uk in the UK market despite being shorter.
# ⚠ .uk was 0.70 until 2026-09-03 — a GUESS, and wrong by 3-4x.
# `[MEASURED]` COMPARABLES_2026-09-03_realised_sales.md §1.3(c), from realised
# sales (domainlore.uk / domainsaleshistory.uk): a one-word commercial generic
# realises £2,000-£5,000 in standalone .uk against £6,000-£35,000 for the .co.uk
# equivalent — roughly a THIRD TO A FIFTH, tight across 2024 and 2025, on a
# market ~20x thinner by volume. So .uk = 0.85 (co.uk) x 0.25 (midpoint) = 0.21.
#
# `[MEASURED 2026-09-04]` CORROBORATED by a second, independent method —
# PROBE_tld_results_2026-09-04.csv, 15 calls appraising the same SLD in both
# TLDs. Across 11 ORDINARY names the appraiser's own .uk/.com ratio lands in
# 0.115-0.185, median 0.165, against the 0.21 these comps produced. Two methods
# that share no inputs agree, so 0.21 is sound and mildly generous.
# The ratio COLLAPSES on premium and short names (ant 0.003, design 0.008,
# healthcare 0.035, refueller 0.094) — exactly the class the PREMIUM-REVIEW
# guards below hold out of automatic pricing, so the multiplier's failure mode
# is confined to names this model already refuses to price.
TLD_MULT = {'com': 1.0, 'co.uk': 0.85, 'uk': 0.21, 'org.uk': 0.20,
            'me.uk': 0.12, 'net': 0.55, 'org': 0.55, 'biz': 0.30,
            'info': 0.30, 'io': 1.0, 'ai': 1.0, 'shop': 0.35,
            'club': 0.30, 'cv': 0.25, 'vin': 0.25, 'us': 0.40}


def tld_multiplier(tld):
    return TLD_MULT.get(tld, 0.40)


def com_basis(value, row):
    """An appraisal re-expressed on a .com basis, so appraisals of different
    TLDs can be pooled into one median. A PROXY appraisal already IS a .com
    string's value; a DIRECT one is a value in its own TLD, so divide the TLD
    factor back out. Validated on the same probe: dividing the .uk appraisal by
    0.21 predicts the real .com appraisal within ~10-35% on ordinary names
    (financial-adviser 3,014 vs 4,475; bacteriology 3,376 vs 4,327;
    calculation 31,852 vs 39,135; catered 15,757 vs 17,839)."""
    if row.get('appraisal_kind') == 'proxy':
        return value
    return value / tld_multiplier(row['tld'])


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

    if row['category'] in THIN_EVIDENCE:
        m *= 0.75; why.append('thin-demand category -25%')

    if row['trademark_flag']:
        m *= 0.50; why.append('trademark risk -50%')

    return m, '; '.join(why)


def hold_reason(r):
    """The reason this domain may never carry an automatic price, or None.

    Hoisted out of the pricing loop on 2026-09-04 so that ONE definition serves
    both the per-row decision and the median POOL below. Nothing here depends on
    a value, so it can be evaluated before anchoring.
    """
    d = r['domain']
    if r['registrar'].startswith('NOT-OWNED'):
        return r['registrar']
    if d in OWNER_FIGURES:
        return 'OWNER-FIGURE:' + OWNER_FIGURES[d]['kind']
    # ⚠ NOT conditional on lacking an appraisal -- that condition was on the
    # single-word and short-name rules until 2026-09-04 and it made the guards
    # DISARM THEMSELVES. The premium queues exist to appraise exactly these
    # names, so the act of covering them stripped the hold and dropped the
    # estate's best names into the algorithmic tail: 137 flipped inside one
    # window, and the pre-fix code run against that day's data puts 129 of them
    # into the sale list carrying $9,447,275 of automatic asking prices.
    # The premise (an appraisal makes a premium name safe to price) is refuted
    # by that same window. WITH a direct appraisal, free.uk prices at ~$8,600
    # against the ~£160,000 its sibling free.co.uk realised -- ~4% of a known
    # real figure -- while at the other extreme the model would ask $4,190,000
    # for ant.co.uk. An appraisal moves this class from unpriceable to
    # CONFIDENTLY wrong, in either direction. Held regardless.
    if is_single_word(d):
        return 'PREMIUM-REVIEW:single-word'
    # OWNER RULE 2026-09-03: "4 letters with vowels or vaguely pronounceable
    # with e.g. y's are worth good money." A scarcity class the model cannot
    # see -- it priced ipry.com at $1,650 keen while the owner rejected a €150
    # offer on it as far too low. This rule was ALREADY unconditional, and the
    # other two were brought into line with it.
    if (r['tld'] == 'com' and int(r['sld_length']) == 4
            and d.split('.')[0].isalpha()
            and set('aeiouy') & set(d.split('.')[0])):
        return 'PREMIUM-REVIEW:4-letter-com'
    # Short names, same scarcity argument: the model priced 2w.uk / 4l.uk /
    # 5s.uk at the $200 floor against realised .uk sales of tp.uk £5,200,
    # fpp.uk £3,500, va.uk £3,300, egg.uk £2,000 (COMPARABLES §1.3c).
    if int(r['sld_length']) <= 3:
        return 'PREMIUM-REVIEW:short-name'
    if r['category'] in NETWORK_KEEP:
        return 'KEEP:network-' + r['category']
    # Owner-ruled quote-together group: a standalone price for one of these is
    # the exact mistake the ruling exists to prevent, so the model refuses to
    # produce one rather than producing one with a caveat a consumer may skip.
    if r.get('quote_with'):
        return 'QUOTE-AS-PAIR:' + r['quote_with']
    if r.get('keep_override'):
        return 'KEEP:' + r['keep_override']
    return None


def main():
    rows = list(csv.DictReader(open(os.path.join(HERE, 'WORKING_table.csv'))))
    appr = {r["domain"]: float(r["dynappraisal"]) for r in rows
            if r["dynappraisal"] and r["dynappraisal"].replace(".", "").isdigit()}

    def med(vals):
        return statistics.median(vals) if vals else None

    # Medians are pooled on a .com BASIS. Before 2026-09-04 they pooled raw
    # appraisals, which mixed .com values with .uk values ~6x smaller and then
    # applied a TLD discount to whatever inherited the median — so a block's
    # anchor moved with the TLD MIX of the few names that happened to be
    # appraised, not with the block.
    sub_med, cat_med = {}, {}
    by_sub, by_cat = collections.defaultdict(list), collections.defaultdict(list)
    # ⚠ POOL = SELLABLE STOCK ONLY (added 2026-09-04). A median exists to price
    # the ordinary names in a block, so it must be built from ordinary names.
    # Held-out names are held precisely because this appraiser cannot price
    # their class, and their appraisals are outliers in BOTH directions — so
    # letting them set a block's anchor propagates the error sideways onto the
    # neighbours instead of containing it.
    # The worked case: pets-vet/general's .com-basis median came out at $11,978,
    # set by proxy appraisals of felines.co.uk ($169,614), bunnies.co.uk
    # ($40,714) and veterinary.co.uk ($48,517) — three single dictionary words
    # the model refuses to price at all. That median priced vetzy.co.uk, an
    # invented brandable, at a $6,250 keen price. It is the mirror image of the
    # failure in HANDOFF §4: median anchoring cannot tell premium from ordinary,
    # so it drags premium names DOWN and, once premium names are in the pool,
    # ordinary ones UP.
    for r in rows:
        if r['domain'] in appr and hold_reason(r) is None:
            v = com_basis(appr[r['domain']], r)
            by_sub[(r['category'], r['subcategory'])].append(v)
            by_cat[r['category']].append(v)
    for k, v in by_sub.items():
        if len(v) >= 3: sub_med[k] = med(v)
    for k, v in by_cat.items():
        if len(v) >= 3: cat_med[k] = med(v)
    global_med = med(list(appr.values())) or 1000.0

    out = []
    for r in rows:
        d = r['domain']
        # `apply_tld` says whether the anchor still needs converting into this
        # domain's TLD. Every anchor is on a .com basis EXCEPT a direct
        # appraisal of this very domain, which the appraiser already priced in
        # its own TLD.
        if d in appr and r.get('appraisal_kind') == 'proxy':
            # The .com equivalent's value, applied to a UK name. The TLD
            # multiplier discounts it to the UK market, but the keyword itself
            # may be worth more or less here than in .com, so this never counts
            # as a direct measurement.
            anchor, src, conf = appr[d], f'proxy-via-{r.get("appraisal_proxy_domain","?")}', 'medium'
            apply_tld = True
            if is_single_word(d):
                # ⚠ NOT ESTIMABLE (2026-09-04). The proxy route imports a .com
                # valuation and the .co.uk multiplier (0.85) assumes near
                # parity. That holds well enough for an ordinary compound, and
                # is CATASTROPHIC for a single dictionary word, where the .com
                # is a category-defining asset and the .co.uk is not.
                # TWO owner figures, both far below what this produced:
                #   scales.co.uk   model $393,917 vs £3,500 PAID   (~89x high)
                #   cartoon.co.uk  model $739,424 vs £5,000+ PAID  (~187x high)
                # and today's TLD probe measured the same collapse from the
                # other side: for single words the appraiser's own .uk/.com
                # ratio is 0.003-0.035, against the 0.21 that ordinary names
                # justify. We have NO evidence base for a premium .co.uk
                # multiplier, so this model does not invent one -- the row
                # keeps the raw proxy figure for reference and is excluded
                # from the portfolio total. Refusal, not a better multiplier.
                conf = 'not-estimable'
        elif d in appr:
            # ⚠ NO TLD MULTIPLIER HERE, and that is the fix of 2026-09-04.
            # Dynappraisal is TLD-AWARE: it prices the actual domain in its
            # actual TLD, so a .uk appraisal is ALREADY a .uk-market number and
            # discounting it again by 0.21 was a ~5x double discount.
            # `[MEASURED]` PROBE_tld_results_2026-09-04.csv — the same SLD in
            # both TLDs: ant.uk 23,144 vs ant.com 8,208,882; design.uk 23,558 vs
            # design.com 3,121,760; healthcare.uk 18,193 vs healthcare.com
            # 516,065. If it were keyword-driven those pairs would be equal.
            # What this cost: effectiveness.uk, appraised 3,576, carried a $350
            # keen price — 3,576 x 0.21 x 0.45.
            anchor, src, conf = appr[d], 'own-appraisal', 'high'
            apply_tld = False
        elif (r['category'], r['subcategory']) in sub_med:
            anchor, src, conf = sub_med[(r['category'], r['subcategory'])], 'subcategory-median', 'medium'
            apply_tld = True
        elif r['category'] in cat_med:
            anchor, src, conf = cat_med[r['category']], 'category-median', 'low'
            apply_tld = True
        else:
            anchor, src, conf = global_med, 'portfolio-median', 'very-low'
            apply_tld = True

        mult, why = quality(r)
        if apply_tld:
            tm = tld_multiplier(r['tld'])
            if tm != 1.0:
                mult *= tm
                why = (why + '; ' if why else '') + f'{r["tld"]} x{tm}'
        else:
            why = (why + '; ' if why else '') + 'direct appraisal in own TLD (no TLD factor)'
        value = anchor * mult
        # A withdrawn or live-site name still gets a value (the owner may want
        # to know what he is holding) but can never be priced for sale.
        # ONE definition, in hold_reason() above, shared with the median pool.
        held = hold_reason(r)
        keen_out, sell = ('', held) if held else (None, 'tbd')
        for tier, floor, keen_frac in TIERS:
            if value >= floor:
                break
        keen = keen_out if keen_out == '' else max(round_up_clean(value * keen_frac), FLOOR_USD)

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
    # ⚠ The portfolio total EXCLUDES not-estimable rows. Before 2026-09-04 it
    # included them and 74% of the headline figure ($17.6m of $23.7m) came from
    # 72 single-word .co.uk names valued off their .com twin -- a number two
    # owner figures show is ~90-190x too high.
    est = [r for r in out if r['confidence'] != 'not-estimable']
    nonest = [r for r in out if r['confidence'] == 'not-estimable']
    tot_val = sum(float(r['value_usd']) for r in est)
    sellable = [r for r in out if r['keen_price_usd']]
    tot_keen = sum(float(r['keen_price_usd']) for r in sellable)
    held = len(out) - len(sellable)
    print(f'portfolio value (USD): {tot_val:,.0f}  [{len(est)} rows]')
    if nonest:
        print(f'  + {len(nonest)} NOT-ESTIMABLE rows excluded (single-word .co.uk '
              f'valued only via their .com twin); their raw proxy figures sum to '
              f'{sum(float(r["value_usd"]) for r in nonest):,.0f} and are not a valuation')
    print(f'{len(sellable)} sellable; if all sold keen: {tot_keen:,.0f} '
          f'({held} held back: live-site or owner-withdrawn)')
    print('\nby category (value desc):')
    bycat = collections.defaultdict(list)
    for r in out: bycat[r['category']].append(float(r['value_usd']))
    for c, vs in sorted(bycat.items(), key=lambda x: -statistics.median(x[1])):
        print(f'  {c:22s} n={len(vs):4d} median={statistics.median(vs):7.0f} total={sum(vs):9.0f}')


if __name__ == '__main__':
    main()
