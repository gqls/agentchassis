#!/usr/bin/env python3
"""Build the valuation working table by joining every inbound source that
exists so far. Deterministic and re-runnable: run it again whenever a new
inbound file lands and the table refreshes. Writes WORKING_table.csv.

Sources (all optional except categories — a missing file just leaves its
columns blank, so the table is honest about what has landed):
  CATEGORIES_2026-09-02_full.csv                 category, subcategory (+ registrar)
  inbound/*_domains_*.csv                        expiry, nameservers
  inbound/dynadot_valuations_*.csv               dynappraisal (any registrar's domains)
  inbound/dynadot_listings_*.csv                 live asks (Dynadot marketplace)
  inbound/spaceship_sellerhub_listings_*.csv     live asks (onSale rows only, owned only)
  inbound/afternic_listings_*.csv                live asks + price_source (when it lands)
  inbound/porkbun_comps_*.csv                    per-category comps median (context only)
  sedo lane EXCLUDED_live_cloudflare_*.txt       keep_override=live-site
"""
import csv, glob, os, re, collections

HERE = os.path.dirname(os.path.abspath(__file__))
SEDO_FENCE = os.path.join(HERE, '..', 'sedo_domain_management', 'outbound')

TRADEMARK_FLAGS = {'mieleonline.com', 'rolex-submariners.com', 'webuyanycarandvan.com'}


def ns_class(ns: str) -> str:
    ns = (ns or '').lower()
    if 'afternic' in ns or 'aftermarket' in ns: return 'afternic'
    if 'atom.com' in ns: return 'atom'
    if 'dan.com' in ns: return 'dan'
    if 'sedoparking' in ns: return 'sedo'
    if 'cloudflare' in ns: return 'cloudflare'
    if not ns: return 'none'
    return 'other'


def latest(pattern):
    files = sorted(glob.glob(pattern))
    return files[-1] if files else None


def main():
    rows = {}
    with open(os.path.join(HERE, latest_name('CATEGORIES_*_estate.csv'))) as fh:
        for r in csv.DictReader(fh):
            d = r['domain'].lower()
            sld = d.split('.')[0]
            rows[d] = {
                'domain': d, 'registrar': r['registrar'], 'category': r['category'],
                'subcategory': r['subcategory'], 'tld': '.'.join(d.split('.')[1:]),
                'sld_length': len(sld), 'expiry': '', 'ns_class': '',
                'dynappraisal': '', 'dynappraisal_date': '',
                'live_ask_marketplace': '', 'live_ask_price': '', 'live_ask_date': '',
                'afternic_min_offer': '', 'afternic_price_source': '',
                'keep_override': '', 'trademark_flag': 'Y' if d in TRADEMARK_FLAGS else '',
                'category_comps_median': '',
            }

    for f in glob.glob(os.path.join(HERE, 'inbound', '*_domains_*.csv')):
        with open(f) as fh:
            for r in csv.DictReader(fh):
                d = r['domain'].strip().strip('"').lower()
                if d in rows:
                    # Nominet's walk yields expiry MONTH only (its list extension
                    # enumerates by month); retail registrars give a full date.
                    rows[d]['expiry'] = (r.get('expiry') or r.get('expiry_month', '')).strip('"')
                    if 'nameservers' in r:
                        rows[d]['ns_class'] = ns_class(r.get('nameservers', ''))

    f = latest(os.path.join(HERE, 'inbound', 'dynadot_valuations_*.csv'))
    if f:
        with open(f) as fh:
            for r in csv.DictReader(fh):
                d = r['domain'].lower()
                if d in rows:
                    rows[d]['dynappraisal'] = r['valuation']
                    rows[d]['dynappraisal_date'] = r['source'].rsplit('_', 1)[-1]

    f = latest(os.path.join(HERE, 'inbound', 'dynadot_listings_*.csv'))
    if f:
        date = re.search(r'(\d{4}-\d{2}-\d{2})', f).group(1)
        with open(f) as fh:
            for r in csv.DictReader(fh):
                d = r['domain'].lower()
                if d in rows and r.get('price'):
                    rows[d].update(live_ask_marketplace='dynadot',
                                   live_ask_price=r['price'], live_ask_date=date)

    f = latest(os.path.join(HERE, 'inbound', 'spaceship_sellerhub_listings_*.csv'))
    if f:
        date = re.search(r'(\d{4}-\d{2}-\d{2})', f).group(1)
        with open(f) as fh:
            for r in csv.DictReader(fh):
                d = r['domain'].strip('"').lower()
                if d in rows and r.get('status') == 'onSale':
                    price = r.get('bin_price') if r.get('bin_price_enabled') == 'true' else r.get('min_price')
                    if price:
                        rows[d].update(live_ask_marketplace='spaceship_sellerhub',
                                       live_ask_price=price.strip('"'), live_ask_date=date)

    # Afternic: price semantics per their lane (2026-09-03) — only buy_now is an
    # ASK; floor/min_offer are the lowest the owner would hear, so they feed the
    # afternic_min_offer column, never live_ask. Empty price = not set (never 0).
    f = latest(os.path.join(HERE, 'inbound', 'afternic_listings_*.csv'))
    unmatched = []
    if f:
        date = re.search(r'(\d{4}-\d{2}-\d{2})', f).group(1)
        with open(f) as fh:
            for r in csv.DictReader(fh):
                d = r['domain'].strip('"').lower()
                if d not in rows:
                    unmatched.append(d)
                    continue
                src = r.get('price_source', '')
                if r.get('price'):
                    if src == 'buy_now':
                        rows[d].update(live_ask_marketplace='afternic',
                                       live_ask_price=r['price'], live_ask_date=date)
                    else:
                        rows[d]['afternic_min_offer'] = r['price']
                rows[d]['afternic_price_source'] = src
        if unmatched:
            up = os.path.join(HERE, f'AFTERNIC_unmatched_{date}.txt')
            with open(up, 'w') as fh:
                fh.write('\n'.join(sorted(unmatched)) + '\n')
            print(f'afternic rows not in categorised estate: {len(unmatched)} -> {up}'
                  ' (mostly the Nominet .uk names — a preview of the missing list)')

    # per-category comps median from the .com keyword pull, -shire noise dropped
    f = latest(os.path.join(HERE, 'inbound', 'porkbun_comps_com_*.csv'))
    if f:
        stem_prices = collections.defaultdict(list)
        with open(f) as fh:
            for r in csv.DictReader(fh):
                sld = r['domain'].split('.')[0].lower()
                for stem in r['matched_stem'].split(';'):
                    if stem == 'hire' and sld.endswith('shire'):
                        continue
                    stem_prices[stem].append(float(r['price']))
        # a domain's comps context = median over comps sharing any of its stems
        for d, row in rows.items():
            sld = d.split('.')[0]
            ps = sorted(p for stem, prices in stem_prices.items() if stem in sld for p in prices)
            if ps:
                row['category_comps_median'] = f'{ps[len(ps)//2]:.0f}'

    fence = latest(os.path.join(SEDO_FENCE, 'EXCLUDED_live_cloudflare_*.txt'))
    if fence:
        for line in open(fence):
            d = line.strip().lower()
            if d in rows:
                rows[d]['keep_override'] = 'live-site'

    out = os.path.join(HERE, 'WORKING_table.csv')
    cols = ['domain', 'registrar', 'category', 'subcategory', 'tld', 'sld_length',
            'expiry', 'ns_class', 'dynappraisal', 'dynappraisal_date',
            'live_ask_marketplace', 'live_ask_price', 'live_ask_date',
            'afternic_min_offer', 'afternic_price_source',
            'category_comps_median', 'keep_override', 'trademark_flag']
    with open(out, 'w', newline='') as fh:
        w = csv.DictWriter(fh, fieldnames=cols)
        w.writeheader()
        for d in sorted(rows):
            w.writerow(rows[d])
    n = len(rows)
    filled = lambda c: sum(1 for r in rows.values() if r[c])
    print(f'{out}: {n} rows')
    for c in ['expiry', 'ns_class', 'dynappraisal', 'live_ask_price',
              'afternic_min_offer', 'afternic_price_source',
              'category_comps_median', 'keep_override', 'trademark_flag']:
        print(f'  {c:22s} filled {filled(c):5d}/{n}')


def latest_name(pattern):
    files = sorted(glob.glob(os.path.join(HERE, pattern)))
    if not files:
        raise SystemExit(f'no {pattern} — categorisation must exist first')
    return os.path.basename(files[-1])


if __name__ == '__main__':
    main()
