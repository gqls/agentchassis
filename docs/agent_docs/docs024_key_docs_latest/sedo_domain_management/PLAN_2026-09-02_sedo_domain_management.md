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
- **P2 — owner obtains access: BLOCKED on the owner + Sedo approval**
  (external; community reports suggest days, `[UNVERIFIED]`). Owner steps:
  RUNBOOK §1–§3.
- **P3 — first credentialed calls.** `--check-secret`, then
  `DomainList 'results=100'`: inventory what (if anything) the account
  holds, and reconcile against the estate's domains. First writing call
  (DomainInsert) only after reading its function doc and with one domain,
  not a batch.
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
