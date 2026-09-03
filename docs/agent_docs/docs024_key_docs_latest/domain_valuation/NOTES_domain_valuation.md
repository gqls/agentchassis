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

## 2026-09-03 — cross-registrar YES; Afternic export in; overpriced MEASURED

- **Dynappraisal accepts non-Dynadot domains — MEASURED**: aakn.com (Porkbun)
  → $4,554, HTTP 200, first call of the 09-03 window. Whole estate appraisable
  at 300/day. Day's walk running (inventory 153 then priority list): this
  session, background task, started ~09-03 morning.
- **Dynadot inventory 451→453** (owner panel read 452, fresh pull 453, diff =
  2 pure additions overhead-cranes.com/paper-cups.com, 0 drops — the estate
  grows between snapshots). Both categorised manually; walker appraises them
  in the inventory pass.
- **Afternic export DELIVERED** (owner's export dated 09-03; their commit
  `a48e37340`): 1,634 rows. price_source split: **buy_now 419 / floor 3 /
  min_offer 1,212** — so ~74% has NO ask, only a floor; "generally overpriced"
  is testable ONLY against the 419. Currency USD-assumed (export carries no
  marking). 0 never appears (empty = not set). Identity flag: owner quoted
  veterinarypractice.co.uk, export holds veterinarypractice.uk. Full snapshot
  (floors/landers/leads/fast-transfer) at
  afternic_domain_management/snapshots/portfolio_2026-09-03.json.
- **689 Afternic rows not in the retail estate = Nominet preview**: 340 .uk +
  337 .co.uk + 10 org.uk + 1 me.uk + 1 .us (`AFTERNIC_unmatched_2026-09-03.txt`)
  — ~half the ~1,500 Nominet estate, with floors, before the walk has run.
  Deterministic pass: 355/689 placed (financial 98, web-digital 85 — the UK
  estate is finance/web-heavy as expected); 334 to LLM second pass (running).
  Registrar = 'afternic-preview', reconcile against the real walk.
- **WORKING_table.csv builder live** (`build_working_table.py`, re-runnable):
  1,339 rows; expiry+NS 100%; live asks 422 (417 afternic buy_now + 5 dynadot);
  afternic floors 532; appraisals 326-and-climbing; comps medians 267;
  19 keep overrides; 3 trademark flags.
- **"Generally overpriced" MEASURED 09-03** (early sample, 80 domains with
  both buy_now ask and appraisal): ask/appraisal median **5.4×**, p25 1.78,
  p75 13.7; 63/80 above 1.5×; extremes ask $25k vs appr $193
  (free-credit-report-check.com). Even min_offer floors run median **5.8×**
  appraisal (n=181). The owner's judgement CONFIRMED and now has a scale:
  keen pricing near appraisal ≈ a ~5× cut from current asks. [Sample is the
  first 326 appraisals — re-run at walk completion.]
  > **CORRECTED same day — the ratio measured the WRONG VARIABLE.** The asks are
  > banded, not per-domain: **250 of 419 buy-now asks are the identical $4,999**
  > (21 distinct values total), **845 of 1,215 floors are exactly $10,000** (12
  > distinct). With one side near-constant the median ratio describes the spread
  > of the APPRAISALS. And the bands do not track quality — the $4,999 and
  > $25,000 ask bands hold names of median appraisal **$1,549 vs $1,646**, and
  > the $10,000 floor band spans **$25–$24,511**. The defect is the ABSENCE of
  > per-domain pricing, not a multiple. Caught by counting distinct values while
  > trying to USE the floors as a second signal (they carry none). Logged in
  > `WRONG_CALLS.md`; README corrected by APPEND after I first rewrote it in
  > place, which the append-only rule forbids and the pattern check caught.

## 2026-09-03 (evening) — estate complete, valuation provisional

- **Nominet CSV ingested**: 1,606 domains (their `f8ca8389d`), expiry by MONTH.
  Zero overlap with the 1,339 retail ⇒ **owned estate = 2,945**. Their walk had
  never actually run before; three EPP bugs fixed the same evening, incl. one
  that would have returned ZERO silently. Corroborated independently here: 683
  of the 692 listed-but-unaccounted names are in their CSV, and the RDAP sweep
  independently returns 683 registered to the owner's DESIGNCONSULT tag.
- **RDAP sweep COMPLETE, 692**: 687 registered (683 owner, 4 others),
  **3 NOT-REGISTERED** (chicklets.co.uk, demisexual.uk, protecty.co.uk),
  2 unsupported (.co has no RDAP service; whois DNS blocked from here).
  cheapbuild.co.uk + enables.co.uk are **LOST not misfiled** — NS verified
  (ben/fay Cloudflare = another account; GoDaddy domaincontrol), live sites we
  don't run. **qlp.us at NameSilo ⇒ possible unseen account.**
  ⇒ `UNACCOUNTED_2026-09-03_answer.md` (the owner's question, answered).
- **Categorisation COMPLETE for the whole estate** (`CATEGORIES_2026-09-03b_
  estate.csv`): 2,951 rows = 2,945 owned + 6 marked NOT-OWNED-*. Nominet second
  pass 578/578 (subcategories reused verbatim from earlier passes; no new
  category needed). Top: home-garden 357, financial 342, web-digital 318,
  consumer-products 256, ai-tech 194.
- **⚠ THE VALUATION IS PROVISIONAL AND MUST NOT DRIVE THE CUT YET.** Only
  **588/2,945 (20%)** have their own appraisal; **Nominet is 0/1,606**. The
  other 80% inherit category medians, so within-category RANKING — the exact
  thing the bottom-500 cut needs — is not yet real for them. Only 7 categories
  have ≥20 own appraisals; the eye-catching ones (education $3,601,
  agriculture $2,689) rest on 5 and 9 samples and mean nothing.
- **Queues built for the remaining 2,357** (~8 windows):
  `inbound/appraisal_queue_direct_2026-09-04.csv` (1,482 .com/.net/.uk) and
  `inbound/appraisal_queue_proxy_2026-09-04.csv` (875 .co.uk/.org.uk/.me.uk via
  the .com string equivalent, recorded as a PROXY). 12 domains on untested TLDs
  (org/cv/vin/biz/ai/io) need one probe each first.

## 2026-09-03 — OWNER RULING: NameSilo out of scope

*"namesilo exists but has nothing interesting in it, we can ignore them. Any
domains still listed there were probably lost."*

- **Do not enumerate NameSilo, do not re-raise it as a coverage gap.** The
  account exists; the owner has judged its contents not worth valuing.
- Anything traced to NameSilo is treated as **lost, not stock** — here `qlp.us`,
  already excluded as not-owned, so the 2,945 estate total is unchanged.
- Implication for listing hygiene, passed to the afternic lane: a live Afternic
  listing whose domain traces to NameSilo advertises something we cannot
  deliver, same class as the three expired names.
- **NOT covered by this ruling:** `pocketvaginas.com`, registered at *Dynadot*
  yet absent from Dynadot's own 453-domain inventory. Still unexplained. One
  low-value name — recorded, not chased.
