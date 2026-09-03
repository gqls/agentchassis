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
  `scripts/domains/sedo-importer-xlsx.py` (self-test 10/10, `--exclude-file`
  repeatable/unioned since 2026-09-03). **Current: draft3, 2026-09-03 —
  2,888 domains** — the 2,895-domain draft2 full portfolio (Dynadot 453 +
  Porkbun 683 + Spaceship 203 + Nominet 1,606, 0 cross-source dupes, minus
  a 50-domain live-site fence — union of Nominet's Cloudflare zone list,
  the live `sites` table, and the NS-based check; RUNBOOK §7) minus a
  **separate 7-domain owner-requested withdrawal** (the Appleby family;
  kept in its own fence file, never merged into the live-site one — RUNBOOK
  §7). `outbound/SEDO_IMPORT_2026-09-03_draft3.{xlsx,csv}`. Interim shape
  agreed with the domain_valuation lane and confirmed against the artefact
  2026-09-03: MAKE_OFFER / for-sale yes / **no price and no minimum on any
  row**; prices arrive as a second import from the valuation lane's
  canonical `OUTPUT_prices_<date>.csv` (their lane, their column freeze).
  **Superseded same day by draft5: 2,879 domains** — two more
  owner-requested withdrawals landed after draft3 (williama.co.uk, joined
  to the Appleby reason-file; 8 wyke-farm/pastured-egg names, a new
  reason-file). One open item from the valuation lane's own question,
  unresolved: 4 other person-name domains (ianstirling.com, kapoor.uk,
  keeler.uk, anne-marie.co.uk) — owner did not answer, left IN the sheet.
  `outbound/SEDO_IMPORT_2026-09-03_draft5.{xlsx,csv}`. **Superseded same
  day by draft6: 2,878 domains** — copyonline.co.uk withdrawn on a direct,
  time-sensitive owner statement (keeper, not stock) relayed via
  copy_quality_two_stage; its `sites` row postdated the live-site fence's
  one-time build by 30 minutes and was never re-caught because the fence
  file was reused across drafts 4–5 instead of re-queried (RUNBOOK §7 now
  states: re-query fresh before EVERY draft). `outbound/SEDO_IMPORT_2026-09-03_draft6.{xlsx,csv}`.
  **Superseded same day by draft7: 2,839 domains** — domain_valuation's
  nameserver sweep found a structural blind spot: the fence is derived
  ONLY from the framework's own Cloudflare/`sites` sources, and is
  therefore blind to the estate's OTHER hosting stack (Clook,
  `dns*.uk-noc.com`, the pre-Cloudflare-rollout sites). 33 confirmed-live
  + 6 ambiguous domains fenced, incl. the owner's own email domain
  (wpx.uk), his own company domain (designconsultancy.co.uk), and a
  client relationship (leopardess.co.uk/.uk, adjacent to the already-
  fenced leopardessconsulting.co.uk) — held pending owner confirmation,
  not assumed. `outbound/SEDO_IMPORT_2026-09-03_draft7.{xlsx,csv}`.
  Sheet-generation work on the ORIGINAL (non-live) portfolio is
  effectively done pending only the priced re-import.
- **P4 widened same day, OWNER RULING**: not merely "decide what to
  automate" — the owner has ruled **live sites should ALSO be listed on
  Sedo, priced high** (his example: webdesign.uk potentially worth
  £1M+ within a year), reversing the fence's original "protect live sites
  from listing" premise. Now a SEPARATE track from the portfolio sheet:
  live sites need real per-domain valuations from the domain_valuation
  lane before anything ships (never a blank-price bulk add), and the
  handful of domains that are the owner's own operating infrastructure or
  an active client relationship stay held out pending his explicit
  confirmation that "all live sites" was meant to include them too.
  **Progress same day**: owner reconciled the 39-domain Clook batch by
  name — 17 stay excluded, 21 released back to ORDINARY stock (not the
  high-value tier). One (`2v.uk`) unaddressed, still held. `draft8 =
  2,860 domains`, `outbound/SEDO_IMPORT_2026-09-03_draft8.{xlsx,csv}`.
  `leopardessconsulting.co.uk` (not part of the 39) confirmed
  PERMANENTLY excluded on a prior, named owner precedent (D4) — needs its
  own by-name reconfirmation, not covered by any blanket answer. A real
  cost basis surfaced unprompted (cartoon.co.uk, owner paid £5,000+) —
  relayed to the valuation lane as a floor. **P4 CLOSED same day**: owner
  ruled the "real prices before listing" gate off entirely ("we'll bear
  with the low balls"), then refined to "Sedo floors allowed, site
  display never floored, unlink the two, never derive a floor from
  `tier`/appraisal." **Draft9 = 2,943 domains** folds the live-sites
  track into the main sheet (blank Minimum Offer, same shape as ordinary
  stock) — the SEPARATE-track design above is superseded; RUNBOOK §9 has
  the standing floor policy. Only owner-named holds remain excluded (36
  total): 18 Clook, 7 appleby, 8 wykefarm/pasturedegg, 1 copyonline, 1
  leopardessconsulting (own dedicated file now, reason reclassified from
  "pending pricing" to "owner-ruled client-protection standard,
  permanent").
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
