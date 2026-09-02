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
