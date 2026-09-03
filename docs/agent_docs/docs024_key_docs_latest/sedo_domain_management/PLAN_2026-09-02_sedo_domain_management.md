# PLAN — Sedo domain management (opened 2026-09-02, owner-directed)

**Goal.** The owner wants their domains manageable on Sedo (marketplace +
parking) by Claude sessions, via Sedo's API.

## What was established 2026-09-02

- Sedo has a usable API: `https://api.sedo.com/api/v1/`, plain GET/POST with
  XML responses; auth = `partnerid` + `signkey` + `username` + `password` on
  every call. Access is granted by emailing api@sedo.com from the registered
  address after joining their partner programme; approval issues the
  Partner ID + SignKey. Sources: api.sedo.com/apidocs/v1/Basic/ (function
  pages incl. DomainList), api.sedo.com landing, community summaries.
- The cluster can reach it: public image pull + egress + a structured
  `SEDOFAULT E7` on dummy creds, measured 2026-09-02 (see NOTES).

## Decisions, with reasons

1. **Bootstrap client is a script (`scripts/domains/sedo-api.sh`), not a Go
   adapter.** Nothing consumes Sedo data yet and the owner has no account
   yet; a platform adapter before the first credentialed call would be
   speculation. Revisit at Phase 4: if Sedo becomes load-bearing (scheduled
   price sync, parking stats feeding traffic-analytics), that is a proper
   adapter service under `platform/` — council-gate scope, register entry,
   the lot. The script is the measuring instrument that earns that design.
2. **Credentials follow the 2026-08-23 owner ruling** (never read a key's
   value into a session): K8s secret `sedo-api-credentials`, injected
   `envFrom: secretRef` into an ephemeral pod, expanded in-container.
   The owner creates the secret from an env-file in their own terminal, so
   values never enter any transcript (RUNBOOK §3).
3. **Transport is `curlimages/curl:8.10.1`, not the fleet's BusyBox wget** —
   BusyBox wget does no TLS cert validation (measured 2026-09-02), and these
   calls carry the account password.
4. **Runner is run → wait → logs → delete, not `kubectl run -i`** — the
   interactive form records its command line in container logs (kubectl's
   own warning), and attach-failure fallback double-prints output.

## Phases

- **P1 — research + scaffold: DONE 2026-09-02.** Script live with
  `--self-test` (offline), `--probe` (cluster, credential-free E7 control),
  `--check-secret`; both probes PASS. The credentialed path is **built but
  unexercised** — no credentials exist anywhere yet.
- **P2 — owner obtains access: PARTIALLY CLEARED 2026-09-02 (evening).**
  The owner already has a Sedo account (registered address
  info@designconsultancy.co.uk) **with partnership status** — so RUNBOOK §1
  is done and pre-dated this lane (corroborated: `officestationery.net`
  already delegates to ns1/ns2.sedoparking.com in the 2026-09-02 Porkbun
  export). Remaining owner steps: §2 (API-access email from that address)
  and §3 (install the secret). API approval wait still applies.
- **P2b — bulk-import sheet (web route, added 2026-09-02 evening, owner
  request).** Sedo's web importer takes the same data with NO API
  credentials — the owner uploads the xlsx in the dashboard. Built:
  `scripts/domains/sedo-importer-xlsx.py` (self-test 9/9). **Superseded by
  draft2, 2026-09-03: full portfolio, 2,895 domains** — Dynadot 453 +
  Porkbun 683 + Spaceship 203 + Nominet 1,606, 0 cross-source dupes, minus
  a **50-domain live-site fence** (widened from 19 — union of Nominet's
  Cloudflare zone list, the live `sites` table, and the original NS-based
  check; RUNBOOK §7 has the method and why one source alone is not
  enough). `outbound/SEDO_IMPORT_2026-09-03_draft2.{xlsx,csv}`. Interim
  shape agreed with the domain_valuation lane: MAKE_OFFER / for-sale yes /
  no price; prices arrive as a second import from their canonical
  `OUTPUT_prices_<date>.csv` (their lane, their column freeze).
- **P3 — first credentialed calls.** `--check-secret`, then
  `DomainList 'results=100'`: inventory what (if anything) the account
  holds, and reconcile against the estate's domains. First writing call
  (DomainInsert) with one domain, not a batch. *Function doc read and
  column mapping recorded 2026-09-02 (RUNBOOK §6 — the owner's importer
  sheet maps 1:1 onto `domainentry`); the one-domain call also confirms
  the `[INFERRED]` nested-key wire shape, and note every insert
  auto-enables Sedo parking regardless of `forsale`.*
- **P4 — decide what to automate (owner decision).** Candidates: listing
  chosen domains for sale, price management, pulling parking stats into
  traffic-analytics. Also decide the Go-adapter question (decision 1).

## Cross-lane constraint

Some estate domains are already parked at **Dan.com/Afternic** (idea.uk —
whose Google snippet is the Dan "for sale" page, a live problem in the
idea.uk lane; boxingonline + adversecreditmortgage sit on marketplace
nameservers, a decided item in the improvement-loop lane, D2). **Moving any
domain's parking or nameservers to Sedo routes through the lane that owns
that domain** — this lane provides the API plumbing, it does not re-point
domains.

**Added 2026-09-02, same day:** the **domains_cloudflare_rollout lane**
(active — its porkbun.py landed 62 seconds before this lane's first commit)
owns the registrar estate itself: Nominet EPP + Dynadot/Porkbun/Spaceship
helpers under `scripts/domains/`, rolling the portfolio onto Cloudflare.
`sedo-api.sh` was moved into that directory to sit with its family.
Division of labour: **registrars + DNS/NS = their lane; the Sedo
marketplace/parking account = this one.** A future "park domain X at Sedo"
touches both (Sedo-side setup here, NS repoint there) — coordinate via the
lane docs, do not reach into their scripts.
