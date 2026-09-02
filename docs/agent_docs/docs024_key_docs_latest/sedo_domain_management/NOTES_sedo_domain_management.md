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
