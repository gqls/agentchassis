# NOTES — Sedo domain management (append-only, newest at the bottom)

## 2026-09-02 — lane opened (owner request: "set up sedo … so that you can manage my domains")

**Research** (web, sources in PLAN):
- API endpoint `https://api.sedo.com/api/v1/<Function>`, GET or POST,
  XML out; SOAP/WSDL also offered, not needed.
- Auth params on every call: `partnerid` (int), `signkey`, `username`
  (≤25 ch), `password` (≤16 ch — the API doc's own cap).
- Access: email api@sedo.com from the registered address; partner-programme
  registration required first; approval issues Partner ID + SignKey.
- Functions: DomainList/-Extended, DomainStatus, DomainInsert/Edit/Delete,
  DomainSearch, portfolio CRUD, parking setup/keyword/template, blacklist
  check. `results` ≤100/call, `startfrom` pages. currency 0=EUR 1=USD 2=GBP.
- Prior art: NO existing Sedo integration anywhere in the estate — repo
  grep hits were archive.org JS noise, the parking-200s landmine, and a
  stats-export aside in the traffic-probe brief.

**Measured today** (each could have come out otherwise):
- `busybox:1.36` ephemeral pod, in-cluster: pulled, resolved, TLSed to
  api.sedo.com, got `<SEDOFAULT> E7 "Partnerid doesn't exist"` on dummy
  creds. So: public image pull works, egress works, API answers, and
  **faults are in-band XML on a readable body** (the BusyBox
  4xx-body-drop limitation doesn't bite here).
- Same probe printed `wget: note: TLS certificate validation not
  implemented` → BusyBox wget REJECTED as transport for credentialed calls.
- `curlimages/curl:8.10.1` pod: pulled, POST with `--fail-with-body`, same
  E7. This is the pinned transport.
- `kubectl run -i` printed its own warning that commands+output land in
  container logs, and its attach-failure fallback double-printed the
  response → runner design changed to run → wait → logs → delete, creds
  via envFrom only.

**Missteps:**
- First cut of `PARAM_RE` used `\[\]` inside an ERE bracket expression —
  backslash does NOT escape there, so the first `]` closed the class and
  both positive self-tests failed. Fix: `[][A-Za-z0-9_.-]` (the
  `]`-first idiom). Caught by the script's own `--self-test` on first run —
  which is the argument for writing the self-test before the first use.

**Built:** `scripts/domains/sedo-api.sh` (self-test PASS 7/7; probe PASS; secret
missing → clear pointer to RUNBOOK §3). Credentialed path UNEXERCISED —
blocked on owner obtaining the account + SignKey (RUNBOOK §1–§3).
Registered as OPP-012.

## 2026-09-02 (later) — concurrent-lane discovery, script relocated

The first commit's scope report ran while HEAD had moved: 62 seconds
earlier the **domains_cloudflare_rollout** lane committed
`scripts/domains/porkbun.py` — that lane (open since 08-03, active now)
already holds registrar helpers (porkbun/dynadot/epp/spaceship plans) under
`scripts/domains/`. My prior-art grep for "sedo" could not have found it
(different keyword, different service class) — the commit stream found it,
which is the shared-tree working as designed. Moved `sedo-api.sh` →
`scripts/domains/` (git mv; both paths named on the commit per the
LANDMINES entry; verified single copy at HEAD), re-pointed every reference,
added the missing OPP-012 index row the pattern check caught, and wrote the
division of labour into PLAN "Cross-lane constraint".

## 2026-09-02 (evening) — owner supplied the bulk-listing format; DomainInsert mapped

Secret still absent (`--check-secret`, 17:31 BST) — P2 remains with the
owner; self-test still 7/7 at HEAD.

Owner provided Sedo's Domain Importer template
(`~/Downloads/Example_File_Domain_Importer.xlsx`). No xlsx reader installed
(openpyxl/pandas both absent) — unzipped it and parsed the sheet XML +
sharedStrings directly. Seven columns, eleven example rows, no embedded
`dataValidation` elements, authored 2021 in Excel. Decoded content and
column semantics now in RUNBOOK §6.

Fetched the DomainInsert function doc
(`apidocs/v1/Basic/functions/sedoapi_DomainInsert.html` — NB the guessed
URL `Basic/DomainInsert.html` 404s; the index page gives the real paths):
`domainentry` array ≤50/request; per-entry `domain`/`forsale`/`price`/
`minprice`/`fixedprice`/`currency` all required, plus `domainlanguage`
(NOT in the sheet — the web importer defaults it, API callers must pass
it); `category` optional ≤3. Two behaviours worth their lines: insert is
ASYNC (post-hoc checks, failures arrive by email), and EVERY insert
auto-enables parking regardless of `forsale` — recorded in RUNBOOK §6
because it interacts with the cross-lane parking constraint.

Wire shape for the array param is `[INFERRED]` from the doc's own
`http_build_query` example (the literal nested-key URL is never printed);
the planned first one-domain call doubles as its confirmation.

## 2026-09-02 (late evening) — owner: account EXISTS; portfolio sheet built (1,318 domains)

Owner message: has **partnership status** and an account under
**info@designconsultancy.co.uk**; wants most domains added; asked this lane
to (a) build the full-portfolio sheet in Sedo's importer format, (b) ask
the dynadot/porkbun/nominet/spaceship sessions for the domain lists, (c)
correspond with the new **domain valuation** session on Sedo+Afternic
pricing. All six are live peer sessions; the valuation lane had already set
the inbound contract, so this lane consumes their `inbound/` rather than
running a second collection.

**Cross-session outcomes (all same evening):**
- dynadot: 451 complete per API contract (one unpaginated response;
  panel-count cross-check pending); **Dynappraisal fetch running** →
  `dynadot_valuations_2026-09-02.csv`, rate-capped/day (50–300 by tier).
- porkbun: 683 delivered (`02ffa6f40`). **CORRECTION to my framing:** the
  global opt-in gates PER-DOMAIN endpoints (future NS ops), NOT the list —
  I had repeated the valuation lane's "delivery blocked on opt-in" line;
  wrong, and the correct fact is in their message. 13/683 NS cells blank
  (DoH SERVFAIL); 600/683 on afternic-family NS; 0 on Porkbun's own
  marketplace (43,203-listing intersection).
- spaceship: 203 confirmed complete (fetched == API `total`). ⚠ their
  seller-hub listings file is NOT an inventory (36/831 registered there).
- nominet: contract confirmed — one CSV (`domain,expiry_month`) will serve
  valuation + sedo; still owner-gated on the login/walk.
- domain valuation: contract agreed — canonical
  `OUTPUT_prices_<date>.csv` in THEIR lane (columns frozen by them, ~10
  fields incl. category/keep_or_sell/confidence); interim blank-price sheet
  approved with two owner-facing caveats (lowball risk without minimums;
  auto-parking disclosure). Cut confirmed: bottom ~500 BUY_NOW keen,
  categories stay together; £150 webdesign.co.uk transfer-away fee
  (2026-08-17) as de-facto keen floor.

**Built:** `scripts/domains/sedo-importer-xlsx.py` (self-test 9/9; emulates
the template's sharedStrings encoding; refuses non-ACE domains, unknown
selling options/currencies, malformed prices; maps prices by header name;
verifies its own artefact by re-reading the zip). Generated
`outbound/SEDO_IMPORT_2026-09-02_draft1.{xlsx,csv}` + provenance CSV +
`EXCLUDED_live_cloudflare_2026-09-02.txt`:
- **1,337 unique domains** in hand (451+683+203, 0 cross-registrar dupes —
  the earlier `uniq -d` "duplicate" was the two quoted header rows my
  `NR>1` failed to skip; `FNR>1` is the correct guard);
- **19 fenced out** (Cloudflare NS = live estate sites, incl.
  websitedesign.com, boxingonline.com, relojistas.com, vetcomparison.uk);
- **1,318 in the sheet**, all MAKE_OFFER / yes / no price;
- NS provenance: 1,193 afternic-family, 111 other, 13 none, **1 already on
  sedoparking.com** (`officestationery.net` — corroborates the owner's
  pre-existing Sedo relationship).
- Verified at the artefact: 1,319 `<row` occurrences (header + 1,318);
  sharedStrings uniqueCount 1,327 = 1,318 domains + 7 headers + 2 shared
  literals. (`grep -c '<row '` reads 1 — single-line XML; count with
  `grep -o | wc -l`.)

Outstanding: Nominet ~1,500 .uk (owner walk) → draft2; prices import once
the valuation lane freezes/ships `OUTPUT_prices`; owner still owes the §2
API-access email + §3 secret for the API route.

**Reconciliation closed (valuation lane, same evening):** their ingest
matches 1,337/0-dupes exactly; the 19-domain fence is adopted into their
model as `keep_or_sell=KEEP` overrides regardless of rank (relojistas.com
additionally carries the owner-confirmed $12k Afternic floor, 2026-07-28).
They message this session when OUTPUT columns freeze — nothing owed from
this lane until then.

**Dynadot completeness CLOSED + draft1 refreshed (same evening, later):**
owner panel cross-check + fresh re-pull confirm `list_domain` complete;
451 → **453** purely by addition (overhead-cranes.com, paper-cups.com,
added by the owner today; zero drops). Their inbound CSV refreshed in
place (`574a04290`); draft1 regenerated from it — **1,320** rows, CSV diff
exactly the two insertions, fence unchanged at 19. Standing practice per
their note: the owner adds domains on occasion, so **re-pull all registrar
CSVs immediately before any sheet the owner will actually upload**.

**Dynadot follow-up (same evening):** appraisals landed for **300 of 451**
(daily quota fired at exactly 300; remainder resumes tomorrow,
idempotent). And a correction to "0 listings": `isForSale` is unreliable
(their landmine) — a marketplace-dump control found **5 live Dynadot Buy
Nows** (see RUNBOOK §7's new priced-import gotcha for names/prices). All 5
verified present in draft1 as make-offer/no-price — no displayed conflict
today; the price-agreement constraint recorded for every priced draft and
relayed to the valuation lane. 394/451 Dynadot domains on Afternic NS —
existing Afternic asks must join the same sweep once the owner's export
lands.
