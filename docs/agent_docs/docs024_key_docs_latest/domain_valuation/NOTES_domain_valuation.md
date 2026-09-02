# NOTES — domain valuation (append-only, newest at the bottom)

## 2026-09-02 — lane opened

- Live sessions found for all four registrar lanes plus afternic and sedo:
  dynadot [611984], porkbun [e9ba38], nominet [4ea0ea], spaceship [0d010f],
  afternic [17bc44], sedo [9822ef]. All idle at time of writing.
- `all_domains.txt` at repo root (18:48 today) is **0 bytes** — the Nominet
  walk's connection blip; the ~1,500 .uk list at Nominet is still unlistef
  anywhere. Owner must run `! python3 scripts/domains/nominet.py login` then
  `walk --months 120` (nominet lane README).
- Registrar tooling already exists: `scripts/domains/{dynadot.sh,porkbun.py,
  spaceship.py,nominet.py,afternic-csv.py,sedo-api.sh,classify_nameservers.py}`.
- Prior valuation discussion hunted by transcript grep: candidates cluster
  2026-08-10..08-14 (460a5226 84 hits, 839df212 39, db85f55f 33, 48fb60ee 25,
  7fe0cd84 24, a107ab07 14, e9ad9395 13). Mining delegated to a background
  agent → findings land in PRIOR_ART_2026-09-02_previous_valuation_discussions.md.
- Afternic lane (opened today) has **no export yet** — inbound/ empty; owner owed
  them a portfolio CSV. Their prices = comparison column only (owner: overpriced).
- portfolio_positioning lane holds domain classification prior art:
  `RUNBOOK_domain_inventory_and_classification.md`, `PORTFOLIO_domains.txt`
  (Jul 31 subset), `REGISTER_positioning.md`.

## 2026-09-02 evening — lane replies

- **Spaceship DELIVERED**: `inbound/spaceship_domains_2026-09-02.csv`, 203
  domains (almost all .com), parked at aftermarket.com/atom.com. Rode into the
  lane-skeleton commit `624990d55` (same-task passenger, kept).
- **Nominet accepted**: walk now emits `DOMAIN\t<name>\t<YYYY-MM>` (their
  `2a0bf9bb7`) — expiry MONTH is the walk's native granularity; ruled month
  precision sufficient here (no 1,500-call `domain:info` enrichment).
  ⚠ root `all_domains.txt` 0-byte = today's failed walk, do not ingest.
  Check walk TOTAL= against owner's ~1,500 (as of 2026-08-19) before trusting.
- **Registrar counts as of 2026-09-02** (their measurement, domains_cloudflare_rollout
  NOTES): Dynadot **451** (442 .com), Porkbun **683**, Spaceship **203**;
  + Nominet ~1,500 ⇒ estate ≈ **2,800+**. Dynadot list_domain carries per-domain
  expiry. ⚠ Porkbun per-domain reads gated on a global opt-in owner was asked to flip.
- **Afternic accepted** with two spec extensions (agreed): 5th column
  `price_source` (buy_now|floor|min_offer|none — floors are NOT askings, and the
  owner's "overpriced" applies to askings); `currency` USD-assumed until an
  export shows its own marking. Their data is only as fresh as the owner's
  latest export (no seller API) — every Afternic figure carries the FILE date.
  Repricing flow: desired prices from THIS lane, bulk-XLSX from theirs, upload
  click stays with the owner. Blocked until owner drops his first export in
  their inbound/.

## 2026-09-02 late evening — three of four lists in; prior-art search concluded

- **Dynadot DELIVERED**: 451 domains (inbound CSV), listings CSV header-only.
- **Porkbun DELIVERED** (their `02ffa6f40`): 683 domains. NO valuation endpoint
  (measured: grep over porkbun.com/llms-full.txt, all 3,261 lines). 0 of 683 in
  Porkbun marketplace (paged all 43,203 listings, intersected). NS classes:
  600/683 PARKED (overwhelmingly afternic NS), 56 REGISTRAR_DEFAULT,
  11 CLOUDFLARE, 2 CLOOK, 13 no-answer. ⚠ their NS column is PUBLIC DELEGATION,
  not registrar config (per-domain registrar reads behind the global opt-in).
  **Comps offer accepted**: their /marketplace/getAll (43k listings, price/tld/
  sld_length filters) can pull comparables; UK pull + keyword pull queued on my
  categorisation freezing the keyword list.
- **Spaceship COMPLETE** (their `51cffa2e2` for listings): 203 domains (API
  total independently 203). NO valuation endpoint (full docs.spaceship.dev
  section list checked). SellerHub 831 rows: 7 onSale (min-offer $10k each,
  robot-hands/mop-kits family; 5 with BIN 2,500 recorded but binPriceEnabled
  =false ⇒ effectively make-offer), 794 onSaleStopped, 30 verifying; sold
  report EMPTY all-time. ⚠ **SellerHub is NOT an inventory — only 36/831
  listed names are Spaceship-registered**; ownership keys on registrar CSVs
  only. NS: 144/203 aftermarket.com (Afternic), 58 atom.com, 1 Cloudflare ⇒
  **Atom.com dashboard export = NEW owner action** (no API keys for Atom).
- **Sedo lane contract agreed** (their proposal + my amendments): canonical
  OUTPUT_prices_<date>.csv in THIS lane, consumers map; their interim
  all-domains MAKE_OFFER blank-price sheet OK as a sheet (upload click is the
  owner's) with two caveats put in front of him (lowball exposure without
  minimums; Sedo auto-parking vs the ~85% of the estate parked at Afternic
  today). Cut reading confirmed: bottom ~500 BUY_NOW keen, kept majority
  MAKE_OFFER + minimum, chosen by CATEGORY BLOCKS not pure rank.
- **Prior-art search CONCLUDED**: no .co.uk/.uk valuation conversation in any
  of 646 retained transcripts (3 scan passes, owner-typed messages). Full
  findings + the five usable anchors → PRIOR_ART_2026-09-02_previous_
  valuation_discussions.md (relojistas $12k floor; £150 transfer-away fee;
  the 010 value ladder; Jul-31 152-domain subset; traffic≠value). Mining
  subagent died on a transient auth error mid-run; re-run inline.
- Estate in hand: **1,337 domains** (451+683+203); Nominet ~1,500 .uk owner-gated.

## 2026-09-02 night — categorisation complete; Dynadot trio incl. partial Dynappraisal

- **Categorisation DONE for all 1,337** (`CATEGORIES_2026-09-02_full.csv`):
  31 categories, 0 uncategorised. Top: home-garden 224, financial 125,
  consumer-products 122, sports-leisure 80, foreign-language 79, ai-tech 79,
  business-services 72, automotive 71. Second pass by subagent (843 rows,
  set-equality verified). ⚠ trademark-adjacent: mieleonline.com,
  rolex-submariners.com (UDRP risk — flag, never price up). Leasing/hire folded
  into subject verticals (recorded in the subagent's reply, 2026-09-02).
- **Porkbun UK comps**: 774 rows (196 .uk, 564 .co.uk, 12 .org.uk, 2 .me.uk),
  their `4eac3364e`. Asks, USD [INFERRED]. .co.uk ask median ~$875/p25 ~$399;
  .uk median sits on a $10k default-price wall. **.com keyword pull requested**
  (40 stems frozen, sent 2026-09-02 night; >2,000-row stems capped at the
  2,000 cheapest).
- **Dynadot DELIVERED all three** (their `a0ec13892`): 451 domains (no
  pagination in the API contract; owner control-panel total cross-check
  pending). ⚠ **`isForSale` is NOT listing state** (their landmine): 5 LIVE
  Buy Now listings found via the marketplace dump control — traderboltai.com
  $7,999 · currencyforecaster.com $3,999 · thailandstocks.com $2,988 ·
  riderlessbikes.com $2,888 · carsforchildren.com $2,508 (USD) = already-priced
  stock. 394/451 on Afternic NS.
- **Dynappraisal PARTIAL 300/451** (daily quota 429s at exactly 300, measured):
  sum $891,355, mean $2,971, max dialyzers.com $15,599. Algorithmic appraisals,
  not sales. Resume idempotent: `scripts/domains/dynadot-appraise-all.sh
  <domains csv> <valuations csv>` — either session, tomorrow. **Asked whether
  Dynappraisal accepts non-Dynadot-registered domains** — if yes, whole estate
  appraisable at 300/day (~5 days retail + ~5 nominet), priority list
  financial+home-garden first.
- **Sequencing rule adopted** (dynadot's point): no keen price ships anywhere
  until the owner's Afternic export is in — avoid undercutting/contradicting
  live asks we haven't seen.
- **Cross-registrar Dynappraisal test INCONCLUSIVE today** (429 fires before
  domain validation on an exhausted quota) — tests on tomorrow's reset; reset
  timezone unstated, first successful call dates the window. Plan recorded in
  rollout lane NOTES (`691f545e1`): test call → 151 resume → ~148 headroom
  starts the priority list. **Priority list committed**:
  `inbound/appraisal_priority_2026-09-03.csv` (886 non-Dynadot retail domains,
  financial → home-garden → value cats → brandables/foreign/misc last).
- **Porkbun .com comps DELIVERED** (their `2a8d3aed8`):
  `inbound/porkbun_comps_com_2026-09-02.csv`, 1,204 unique listings, 40 stems,
  matched_stem multi-valued (semicolon). Biggest stems agent 184 / health 115 /
  robot 109; cap never fired. ⚠ `hire` rows include -shire noise
  (worcestershire etc.) — filter before use. **aluminium AND aluminum = 0,
  packaging = 0 marketplace-wide** — thin-demand signal for those categories,
  usable in the keep/sell call. Porkbun lane's valuation involvement COMPLETE
  (API toggle = repricing writes only).
