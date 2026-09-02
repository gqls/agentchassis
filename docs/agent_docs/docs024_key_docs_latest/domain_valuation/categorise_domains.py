#!/usr/bin/env python3
"""Deterministic first-pass categoriser for the domain valuation lane.

Reads every inbound/*_domains_*.csv, assigns (category, subcategory, method)
per domain, writes CATEGORIES_<date>_draft.csv beside itself. Word-aware:
keywords match as whole tokens after splitting on hyphens/digits, or as
compound substrings only when listed in COMPOUNDS (the bare substring trap:
'ai' matches 'drain' — measured 85 false vs 13 real on 2026-09-02).

Anything unmatched lands in 'uncategorised' for the second (human/LLM) pass —
this script never guesses. Re-run freely; output is deterministic.
"""
import csv, glob, os, re, sys, datetime

HERE = os.path.dirname(os.path.abspath(__file__))

# token → (category, subcategory). First hit wins; order in RULES is priority.
RULES = [
    # financial — the owner's worked example of a category that stays together
    (('mortgage','mortgages','remortgage','remortgages'), ('financial','mortgages')),
    (('loan','loans','lending','lender','lenders','borrow','borrowing'), ('financial','loans')),
    (('insurance','insurances','insure','insurer','insurers'), ('financial','insurance')),
    (('credit','debt','debts'), ('financial','credit-debt')),
    (('bank','banks','banker','banking','savings'), ('financial','banking-savings')),
    (('invest','investment','investments','invested','investing','investeren','equity','pension','pensions'), ('financial','investing')),
    (('finance','finances','financial','financing'), ('financial','general')),
    # ai & tech
    (('ai','agent','agents','agentic','automation','automate','llm','gpt'), ('ai-tech','ai')),
    (('robot','robots','robotic','roboten'), ('ai-tech','robotics')),
    (('data','api','apis','software','app','apps','saas','cloud'), ('ai-tech','software')),
    # web & digital services
    (('website','websites','web','site','sites','webdesign','seo','hosting','host'), ('web-digital','web')),
    (('design','designs','designer','designers','designed'), ('web-digital','design')),
    (('mailing','marketing','advertise','advertising','promotion'), ('web-digital','marketing')),
    # audits & business services
    (('audit','audits','auditing','auditor','auditors'), ('business-services','audits')),
    (('account','accounts','accounting','accountancy','accountant','accountants'), ('business-services','accounting')),
    (('business','businesses','enterprise','company','companies','team','teams','sme','smb','b2b','corporate'), ('business-services','general')),
    (('jobs','job','recruitment','hire','hiring'), ('business-services','jobs-hire')),
    # home, garden, trades
    (('kitchen','kitchens','diner','diners','extension','extensions','renovation'), ('home-garden','kitchens-extensions')),
    (('window','windows','door','doors','glass','frosted','glazing','pergola','pergolas','aluminium','aluminum','upvc'), ('home-garden','windows-doors')),
    (('boiler','boilers','heating','heater','heaters','heat','radiator','radiators','plumbing','plumber','plumbers','drain','drains','drainage'), ('home-garden','heating-plumbing')),
    (('solar','energy','grid'), ('home-garden','energy')),
    (('garden','gardens','gardening','landscape','landscaped','landscaping','grass','lawn','planters'), ('home-garden','garden')),
    (('roof','roofing','tile','tiles','tiling','floor','flooring','paving','scaffolding','builder','builders','buildings','prefab'), ('home-garden','building-trades')),
    (('furniture','sofa','sofas','mattress','mats','mat','blinds','curtains','lights','lighting'), ('home-garden','furnishings')),
    (('clean','cleaner','cleaners','cleaning','maintenance','repairs','repair','handyman'), ('home-garden','services')),
    (('home','homes','house','houses','residential','property','properties','estate'), ('property','general')),
    # consumer & retail
    (('frier','friers','fryer','fryers','appliance','appliances','vacuum','mop','mops','kit','kits','gadget','gadgets'), ('consumer-products','appliances')),
    (('gift','gifts','xmas','christmas','toys','toy'), ('consumer-products','gifts')),
    (('shop','shopping','store','buy','discount','deals'), ('consumer-products','retail')),
    # niches
    (('travel','holiday','holidays','villa','villas','airport','flights','hotel','hotels'), ('travel','general')),
    (('vet','vets','veterinary','pet','pets','dog','dogs','cat','cats','animal','animals'), ('pets-vet','general')),
    (('coffee','beans','tea','food','foods','egg','eggs','recipe','recipes','restaurant'), ('food-drink','general')),
    (('health','healthcare','surgery','medical','dental','dentist','clinic','therapy','seaweed','monitors'), ('health-medical','general')),
    (('pickleball','golf','boxing','sports','sport','fitness','gym','cycling','fishing'), ('sports-leisure','general')),
    (('car','cars','auto','autos','van','vans','suv','suvs','motor','motoring','vehicle','vehicles'), ('automotive','general')),
    (('calculator','calculators','calc','comparison','compare','tables','checker','check'), ('tools-comparison','general')),
    (('law','legal','solicitor','solicitors','claims'), ('legal','general')),
    (('school','schools','course','courses','cursus','learning','training','tuition'), ('education','general')),
]

# Substring stems, derived from RULES: every keyword >=5 chars matches anywhere
# in the SLD (unhyphenated compounds are the norm here — 'capitalrepaymentmortgage').
# Longest stem wins so 'automation' can never fall through to 'auto'.
# Short keywords (<5) stay token-only: 'ai' in 'drain', 'law' in 'lawn',
# 'heat' in 'wheat', 'bank' in 'sandbanks' are the measured traps.
EXTRA_STEMS = [
    ('webdesign', ('web-digital', 'web')), ('websitebuild', ('web-digital', 'web')),
    ('automat', ('ai-tech', 'ai')), ('insur', ('financial', 'insurance')),
    ('loan', ('financial', 'loans')),   # curated 4-char: no English word contains 'loan'
    ('debt', ('financial', 'credit-debt')),
]

def _stems():
    stems = []
    for kws, cat in RULES:
        for k in kws:
            if len(k) >= 5:
                stems.append((k, cat))
    stems.extend(EXTRA_STEMS)
    stems.sort(key=lambda s: -len(s[0]))
    return stems

STEMS = _stems()


def tokens(sld: str):
    return [t for t in re.split(r'[-0-9]+', sld) if t]


def categorise(domain: str):
    sld = domain.split('.')[0].lower()
    toks = set(tokens(sld))
    for kws, cat in RULES:
        if toks & set(kws):
            return cat + ('token',)
    for stem, cat in STEMS:
        if stem in sld:
            return cat + ('stem',)
    if len(sld) <= 4:
        return ('brandable-short', f'{len(sld)}-char', 'length')
    return ('uncategorised', '', 'none')


def main():
    date = datetime.date.today().isoformat()
    rows = {}
    for f in sorted(glob.glob(os.path.join(HERE, 'inbound', '*_domains_*.csv'))):
        registrar = os.path.basename(f).split('_')[0]
        with open(f) as fh:
            for row in csv.DictReader(fh):
                d = row['domain'].strip().lower().strip('"')
                rows.setdefault(d, registrar)
    out = os.path.join(HERE, f'CATEGORIES_{date}_draft.csv')
    counts = {}
    with open(out, 'w', newline='') as fh:
        w = csv.writer(fh)
        w.writerow(['domain', 'registrar', 'category', 'subcategory', 'method'])
        for d in sorted(rows):
            cat, sub, method = categorise(d)
            counts[cat] = counts.get(cat, 0) + 1
            w.writerow([d, rows[d], cat, sub, method])
    print(f'{out}: {len(rows)} domains')
    for c, n in sorted(counts.items(), key=lambda x: -x[1]):
        print(f'  {n:5d} {c}')


if __name__ == '__main__':
    main()
