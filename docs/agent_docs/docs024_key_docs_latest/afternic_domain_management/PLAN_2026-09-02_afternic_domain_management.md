# PLAN — Afternic domain management (opened 2026-09-02, owner-directed)

**Goal.** The owner wants his domains manageable at Afternic (marketplace;
absorbed Dan.com, so it holds the bulk of the parked portfolio — ~998 of
1,567 domains sat on dan.com nameservers in the 2026-07-30 registry export)
by Claude sessions. Scope, chosen by the owner 2026-09-02: **listings +
pricing, sales + leads, verification + NS state.**

## What was established 2026-09-02

- **Afternic has NO self-serve seller API.** Its documented APIs
  (partnerportaldocs.afternic.com) are for registrar partners in the Fast
  Transfer/DLS programme. The seller-facing automation shipped 2025/26 is
  "Portfolio Agent" — a conversational agent inside their own dashboard, not
  an API with keys. What automating sellers actually use is the dashboard's
  internal JSON endpoints with a logged-in session cookie (unofficial,
  brittle). Sources: blog.afternic.com (2024/2025 reviews, portfolio-agent,
  bulk-upload-walkthrough), partnerportaldocs.afternic.com, domainnamewire
  2025-03-19 (self-brokering), NamePros threads.
- **Bulk changes go in via an XLSX template** (`bulk_upload_sample_v3.xlsx`,
  downloaded from the dashboard's bulk upload page; blank cell = no change).
  The portfolio export's exact format is `[UNVERIFIED]` until the first real
  file arrives — the parser is built to refuse rather than guess.
- GoDaddy's public Aftermarket API does not help: allowlisted resellers
  only, and the estate's registrars are Dynadot/Porkbun/Spaceship/Nominet,
  not GoDaddy.

## Decisions, with reasons

1. **Route = no-credential CSV loop (OWNER'S CHOICE, 2026-09-02, in chat).**
   Options put to him: (a) session-cookie integration against the dashboard
   JSON endpoints, (b) requesting official API access, (c) CSV loop. He
   chose (c): he exports the portfolio CSV and uploads the bulk XLSX; we
   parse, reconcile, report, and prepare. Zero credential and ToS risk; the
   cost is that the owner is the transport for every change. If the loop
   proves too slow in practice, (a) remains open — it would follow the
   sedo-api.sh secret-in-pod pattern, session cookie in a K8s secret.
2. **Ingest before generate.** The parser (`scripts/domains/afternic-csv.py`)
   maps columns by HEADER NAME only, refuses rows whose cell count disagrees
   with the header, reports unmapped headers rather than interpreting them,
   and takes `--control DOMAIN:FIELD:VALUE` assertions. This mechanises the
   WRONG_CALLS 2026-07-28 lesson (a positional read of a dashboard paste
   invented "Minimum Offer = 0" and reached five documents). The generate
   half is NOT built until a real export has locked the vocabulary and the
   v3 template file is in hand — a writer built against guessed columns
   would ship guessed prices.
3. **A script beside the registrar-helper family, not a Go adapter** — same
   reasoning as the Sedo lane's decision 1: nothing consumes this data yet;
   the script is the measuring instrument that earns any platform design.
4. **Verification + NS state reuses `scripts/domains/classify_nameservers.py`**
   — it already classifies afternic/dan/marketplace delegation (measured on
   this estate) and separates delegation from serving (`--check-http`). No
   new build; the RUNBOOK documents the invocation.

## Phases

- **P1 — research + scaffold: DONE 2026-09-02.** Route decided by owner;
  `afternic-csv.py` live, 13-case `--self-test` PASS and the cell-count
  guard proven by mutation (disabling it fails the test).
- **P2 — owner supplies the two files: BLOCKED on owner** (RUNBOOK §1–§2):
  portfolio export → `inbound/`, bulk template v3 → lane dir. Optional
  third: whatever the dashboard offers as a sales/leads export.
- **P3 — first real ingest.** Lock `ALIASES` against the real headers
  (record any additions in NOTES), set `--control` from a value the owner
  quotes off his dashboard, cross-check `--known` against the estate
  (`sites` table; later the registrar enumerations — Dynadot listed 451
  domains 2026-09-02). Snapshot becomes the baseline; every later ingest
  diffs against it.
- **P3a — valuation hand-off (added 2026-09-02, same day).** After every
  successful ingest, `valuation-csv` writes the domain_valuation lane's
  feed (`domain_valuation/inbound/afternic_listings_<date>.csv`,
  `domain,price,currency,status,price_source`) — committed by pathspec,
  their session notified. RUNBOOK §3 has the commands.
- **P4 — generate half.** Writer that fills `bulk_upload_sample_v3.xlsx`
  (or its CSV equivalent) from a desired-state source. Source is an OWNER
  decision: for the ~40 estate sites, `site_specs` aspect `commercial`
  (price-by-tier, about_page_commercial lane — only 2 rows exist today, so
  most tier calls are still owed); for the wider ~1,500, a prices file he
  edits. Every generated file lists its own diff before he uploads it.
  **First named customer (2026-09-02): the domain_valuation lane's
  bottom-~500 repricing** — the owner considers current Afternic prices
  generally overpriced; the valuation lane produces the new prices, this
  lane turns them into the bulk upload.
- **P5 — cadence.** Owner exports on a rhythm he chooses; each ingest
  reports the diff (sales show up as removed/changed rows; leads/views as
  count movement). Revisit whether any of this earns automation.

## Cross-lane constraints (mirrors the Sedo lane)

- **Registrars + DNS/NS = domains_cloudflare_rollout lane** (their helpers:
  dynadot.sh, porkbun.py, spaceship.py, epp.pl). **The Afternic account =
  this lane.** Re-pointing any domain's nameservers routes through the lane
  that owns the domain (idea.uk lane; improvement-loop D2 for
  boxingonline/adversecreditmortgage).
- relojistas.com's listing (floor $12,000, owner 2026-07-28;
  `marketplace_url` = forsale.godaddy.com lander) is the about_page_commercial
  lane's worked case — commercial seam changes go through that lane.
- The Sedo lane (sedo_domain_management, OPP-012) is the sibling: same owner
  goal, credentialed API route. Neither lane re-points domains.
