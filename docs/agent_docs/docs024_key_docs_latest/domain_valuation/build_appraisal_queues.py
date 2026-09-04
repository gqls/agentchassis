#!/usr/bin/env python3
"""Rebuild the Dynappraisal queues from the CURRENT estate. Deterministic and
re-runnable — run it after every window, and after any inbound re-pull.

WHY THIS IS A SCRIPT AND NOT A HAND-BUILT CSV. The handoff has said "rebuild
the queues after each window" since 2026-09-03; with no script, that meant
reconstructing them by hand, so it did not happen. Measured 2026-09-04, the
queues then in the directory:
  * led with 95 financial rows, a category the owner ruled a whole-category
    network KEEP at ~19:00 on 09-03 -- nine hours AFTER the queue was built.
    Financial is never sold, so a third of a 300-call day bought nothing.
  * contained all 23 owner-withdrawn domains (the Appleby family names, the
    wyke/pastured-egg group, wpx.uk, rolex-submariners.com).
  * predated the estate growing by 78 domains, so the new names were absent.
A rule in prose is not a control -- this lane has now been bitten by that three
times (quote-as-a-pair, the owner figures, and this).

ORDERING PRINCIPLE: spend each call where it changes a DECISION.
value_domains.py anchors a domain on its sub-category median once that block
has >=3 appraisals, and falls back to the category median (low confidence) or
the portfolio median (very low) otherwise. So the first three appraisals in a
block are worth far more than the fourth: they re-anchor every domain in it.
The bulk queue is therefore ordered by BLOCK LEVERAGE -- under-covered blocks
first, biggest block first -- not by category rank. Measured 2026-09-04: 182
calls bring every sellable block to that threshold, re-anchoring 694 domains.

Outputs (all with a `domain` column, which the walker finds BY NAME):
  appraisal_queue_PREMIUM_direct_<date>.csv   scarcity classes, appraisable TLD
  appraisal_queue_PREMIUM_proxy_<date>.csv    scarcity classes, .com proxy
  appraisal_queue_direct_<date>.csv           sellable, block-leverage order
  appraisal_queue_proxy_<date>.csv            sellable, .com proxy
  appraisal_queue_LOW_held_<date>.csv         held stock (network-keep, live
                                              sites): valued for the estate
                                              picture, never for a sale price,
                                              so it queues LAST -- listed in
                                              its own file so the deprioritise
                                              is visible, not buried in a sort
  appraisal_probe_untested_tlds_<date>.csv    one of each unproven TLD
"""
import csv, os, collections, datetime, sys

HERE = os.path.dirname(os.path.abspath(__file__))
TODAY = datetime.date.today().isoformat()

# TLDs Dynappraisal is PROVEN to cover, from the appraisals already on file
# (2026-09-04: com 577, uk 4, net 2, org/biz/club/info/shop 1 each -- every one
# returned a real number). .co.uk/.org.uk/.me.uk are proven NOT covered: they
# answer HTTP 200 with "$--", which is a real outcome, not an error.
DIRECT_TLDS = {'com', 'net', 'uk', 'org', 'biz', 'club', 'info', 'shop'}
PROXY_TLDS = {'co.uk', 'org.uk', 'me.uk'}
# Everything else is untested -- one probe call each before it may be queued.


def load():
    rows = list(csv.DictReader(open(os.path.join(HERE, 'WORKING_table.csv'))))
    vpath = os.path.join(HERE, 'VALUATION_2026-09-03_draft.csv')
    if not os.path.exists(vpath):
        sys.exit('run value_domains.py first -- the queues read its sale_status, '
                 'so the premium/keep rules are never re-implemented here')
    status = {r['domain']: r['sale_status'] for r in csv.DictReader(open(vpath))}
    return rows, status


def has_appraisal(r):
    return bool(r['dynappraisal'] and r['dynappraisal'].replace('.', '').isdigit())


def write(path, rows, proxy=False):
    cols = ['priority', 'domain'] + (['proxy_domain'] if proxy else []) + \
           ['category', 'subcategory', 'reason']
    with open(path, 'w', newline='') as fh:
        w = csv.DictWriter(fh, fieldnames=cols)
        w.writeheader()
        for i, r in enumerate(rows, 1):
            r['priority'] = i
            w.writerow({c: r.get(c, '') for c in cols})
    print(f'  {os.path.basename(path):52s} {len(rows):5d}')
    return len(rows)


def main():
    rows, status = load()
    todo = [r for r in rows if not has_appraisal(r)]
    skipped_done = len(rows) - len(todo)

    prem_d, prem_p, bulk_d, bulk_p, low, probe = [], [], [], [], [], []
    seen_untested = set()

    # Block coverage, counted over SELLABLE rows only -- a block's median is
    # what prices its sellable members, so held stock does not earn a call.
    cov = collections.defaultdict(lambda: [0, 0])
    for r in rows:
        if status.get(r['domain']) != 'tbd':
            continue
        k = (r['category'], r['subcategory'])
        cov[k][0] += 1
        if has_appraisal(r):
            cov[k][1] += 1

    for r in todo:
        st = status.get(r['domain'], 'tbd')
        tld = r['tld']
        rec = dict(domain=r['domain'], category=r['category'],
                   subcategory=r['subcategory'])
        if tld in PROXY_TLDS:
            rec['proxy_domain'] = r['domain'].split('.')[0] + '.com'
            proxyish = True
        elif tld in DIRECT_TLDS:
            proxyish = False
        else:
            if tld not in seen_untested:      # one probe per unproven TLD
                seen_untested.add(tld)
                rec['reason'] = f'untested TLD .{tld} -- probe before queueing'
                probe.append(rec)
            continue

        if st.startswith('NOT-OWNED') or st.startswith('OWNER-FIGURE'):
            continue                          # not ours / already has a real number
        if st.startswith('KEEP'):
            rec['reason'] = st                # network-keep or live site
            low.append(rec)
            continue
        if st.startswith('PREMIUM-REVIEW'):
            rec['reason'] = st
            (prem_p if proxyish else prem_d).append(rec)
            continue

        n, a = cov[(r['category'], r['subcategory'])]
        deficit = max(0, min(3, n) - a)       # calls still needed to re-anchor
        rec['reason'] = (f'block {r["category"]}/{r["subcategory"]} n={n} '
                         f'appraised={a} deficit={deficit}')
        rec['_sort'] = (-deficit, -n, r['domain'])
        (bulk_p if proxyish else bulk_d).append(rec)

    for q in (bulk_d, bulk_p):
        q.sort(key=lambda x: x.pop('_sort'))

    print(f'estate {len(rows)}; already appraised {skipped_done}; to do {len(todo)}')
    print('written:')
    p = os.path.join(HERE, 'inbound')
    write(f'{p}/appraisal_queue_PREMIUM_direct_{TODAY}.csv', prem_d)
    write(f'{p}/appraisal_queue_PREMIUM_proxy_{TODAY}.csv', prem_p, proxy=True)
    write(f'{p}/appraisal_queue_direct_{TODAY}.csv', bulk_d)
    write(f'{p}/appraisal_queue_proxy_{TODAY}.csv', bulk_p, proxy=True)
    write(f'{p}/appraisal_queue_LOW_held_{TODAY}.csv', low)
    if probe:
        write(f'{p}/appraisal_probe_untested_tlds_{TODAY}.csv', probe)

    under = [(k, n, a) for k, (n, a) in cov.items() if a < 3]
    need = sum(min(3, n) - a for k, n, a in under)
    print(f'\nsellable blocks {len(cov)}; under-covered {len(under)}; '
          f'{need} calls re-anchor {sum(n for k, n, a in under)} domains')


if __name__ == '__main__':
    main()
