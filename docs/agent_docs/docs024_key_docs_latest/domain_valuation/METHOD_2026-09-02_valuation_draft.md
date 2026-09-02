# METHOD (draft) — how each domain gets a value and the bottom-500 gets chosen

Status 2026-09-02: DRAFT — columns freeze only when Dynappraisal + the Afternic
export are in. Consumers (sedo, afternic lanes) build against the frozen
RUNBOOK copy, not this draft.

## Inputs, per domain

| input | source | state |
|---|---|---|
| category, subcategory | `categorise_domains.py` + LLM second pass | first pass done (494/1337), second running |
| registrar, expiry | inbound registrar CSVs | dynadot/porkbun/spaceship in; nominet owner-gated |
| Dynappraisal value | dynadot lane | fetch running (their session) |
| Afternic ask/floor/Views/Leads/searches | owner export → afternic lane | owner-gated |
| Atom.com asks | owner export (58 spaceship domains park there) | owner-gated, NEW |
| marketplace comps | porkbun lane (43k listings; UK pull done, 774 rows) | keyword .com pull awaits frozen keyword list |
| live-site fence | sedo lane EXCLUDED_live_cloudflare_2026-09-02.txt (19 domains) | adopted: KEEP regardless of rank |
| prior anchors | PRIOR_ART doc | £150 transfer fee floor; relojistas $12k floor; 010 value ladder |

Comps caution: marketplace prices are ASKINGS, not sales (.co.uk ask median
~$875, p25 ~$399; the .uk median sits on a $10k default-price wall — 2026-09-02
snapshot). Keen pricing prices to SELL, i.e. well under ask medians.
"High views ≠ valuable views" (dormant-domain research) applies to Afternic
Views/Leads.

## Scoring → tier

Transparent per-domain score, recorded with its parts (no black box):
keyword commercial value in its vertical · exact-match quality (singular,
unhyphenated, .com or the natural UK TLD for UK-intent terms) · length/
brandability · TLD (.com > .co.uk ≳ .uk > minor) · negative marks (hyphens,
plurals-of-plurals, misspellings, foreign-language without a market, trademark
adjacency — trademark-adjacent names are flagged, not priced up) · demand
signals when they land (Dynappraisal, Views/Leads, comps density).

Tiers: **A** premium (five figures) · **B** strong (£2k–7.5k) · **C** solid
(£750–2k) · **D** weak (£250–750) · **E** dross (keen quick-sale £150–250 —
never below the £150 the owner already charges as a transfer-away fee).
Bands are DRAFT and owner-adjustable; keen pricing for the sell-500 means
pricing at or below the band floor of the tier.

## The cut (owner rules, 2026-09-02)

1. Rank within category, not globally.
2. **Categories stay together**: the sell-~500 is assembled from whole weak
   categories/subcategory blocks first; a kept category (e.g. financial)
   retains its weak members rather than having them isolated into the sale.
3. The 19 live-site domains are KEEP outrights; so is anything carrying a
   built framework site by the time of the cut (the value ladder says build
   beats sell for names the programme will actually reach).
4. Output: `OUTPUT_prices_<date>.csv` — draft columns (NOT frozen):
   domain, registrar, category, subcategory, tier, keep_or_sell,
   selling_option (BUY_NOW|MAKE_OFFER), price, min_price, currency,
   valuation, valuation_source(s), confidence, afternic_current_ask,
   afternic_ask_date, notes.
   Consumers: sedo lane (Sedo importer), afternic lane (bulk XLSX).
