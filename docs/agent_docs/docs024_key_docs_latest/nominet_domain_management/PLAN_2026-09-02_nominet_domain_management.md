# PLAN — Nominet domain management (opened 2026-09-02, owner-directed)

**Mandate.** Owner, 2026-09-02, verbatim: *"if it doesn't exist please create it
and take responsibility for everything nominet here."* This lane owns everything
Nominet: the TAG(s), EPP access and its allowlist, the EPP client family, the
tag's domain inventory, registrations, transfers in/out, NS changes for
Nominet-held domains, and chasing the second-TAG application.

## What this lane inherited (state as of 2026-09-02, all previously scattered)

- **TAG `DESIGNCONSULT`** ("for now", owner 2026-08-11). EPP password held.
  `~/.config/nominet/credentials` (`TAG=` + `EPP_PASSWORD=` lines, 0600 —
  existence re-verified 2026-09-02, contents never read by a session).
- **EPP**: `epp.nominet.org.uk:700`, TLS, RFC 5734 framing, **IPv4 pin** (the
  families get different treatment). LOGIN proven 2026-08-11 from cluster node
  `.26`; greeting (2,531 B) re-measured 2026-09-02 from node `.37` — but the
  greeting is served to ANY IP, so only a login proves the allowlist (LANDMINES).
- **Allowlist**: the five cluster node IPs (`134.213.168.26/.37/.44/.54/.56`),
  owner-confirmed added 2026-08-11. Stale office IPs likely still present
  (151.226.83.138 et al., from the 08-11 reading) — removal is low-priority
  hygiene, owner-side.
- **Client family** (paths are load-bearing — cited fleet-wide, do not move):
  - `scripts/domains/epp.pl` — login + list-by-expiry-month, read-only,
    credentials on stdin. Staged at `/tmp/epp.pl` in `postgres-clients-0`
    (re-copied 2026-09-02; pod restarts wipe it).
  - `idea_uk_vm_site/box/nominet-epp-ns-change.py` (VMB-015) — `domain:update`
    NS repoint; proven live.
  - `idea_uk_vm_site/box/nominet-epp-domain-check.py` (VMB-016) — `domain:check`;
    proven from a cluster pod, both result classes.
  - `idea_uk_vm_site/box/nominet-epp-domain-register.py` (VMB-017) —
    `domain:create`; BUILT, dry-run default, **never `--apply`d** (costs money,
    creates a registry object). TLD-fenced .uk/.co.uk.
  - `site_delivery_and_editor/find_customer_domain.sh` (VMB-018) — domain
    finding; proven (28 names live-checked).
- **Rulings in force**:
  - 2026-08-21: selling a domain = TWO Nominet operations — Registrant Transfer
    (registry-only, £10+VAT normally, NOT doable over EPP) + the free TAG change.
    Manual per-domain is fine for now. (`458affaf7`.)
  - D1: the domain stays in the owner's name until a sale is agreed.
  - 2026-08-26: customer domains register under `DESIGNCONSULT` interim; they
    move to the second TAG when Nominet grants it.
- **Second TAG application**: SUBMITTED 2026-08-11 (Channel Partner shape —
  Nominet allows one Self-Managed tag per registrar). **No status heard since.**
- **Estate size**: **1,606 `.uk` domains, ENUMERATED 2026-09-03** (`walk
  --months 120`, exit 0, 120/120 list calls 1000, zero parser-mismatch
  warnings) — ahead of the owner's ~1,500 estimate (08-19), plausible growth.
  Delivered to the domain_valuation and sedo lanes as
  `domain_valuation/inbound/nominet_domains_2026-09-03.csv`
  (`domain,expiry_month`). The walk itself needed three bug fixes on its
  first real run (nested XML, undeclared std-list extension, wrong parser
  element) — see NOTES 2026-09-03 and the two LANDMINES entries it produced.

## Constraint that shapes everything: sessions cannot run credentialed EPP or CF writes

A session's permission classifier refuses to pipe a credentials file into
another process (2026-08-19, re-confirmed 2026-09-02 on the Cloudflare
zone-create POST) — correctly, since that is indistinguishable from
exfiltration. So this lane's operating model is: **stage exact commands, owner
runs them** (`! <command>` in a session prompt lands the output in the
conversation). Read-only paths (registry `dig`, CF GETs, greeting probes) run in
session; anything credentialed-mutating is owner-run.

## Phases

- **P0 — lane creation + the dangling-delegation incident: DONE 2026-09-02.**
  See NOTES: the owner's Nominet NS batch (advertise.co.uk, designblog.co.uk,
  seotools.co.uk, websitepromotion.co.uk → alexis/leah) ran with no Cloudflare
  zones behind it; all four are dark/going dark. Recovery staged:
  `scripts/domains/cf-zone-bootstrap.sh` (read paths proven; owner runs the
  mutating form).
- **P1 — the tag inventory: DONE 2026-09-03** (`walk --months 120`, 1,606
  domains, delivered — see above). Still owed: classify with
  `classify_nameservers.py` and reconcile into the estate picture alongside the
  three registrar inventories.
- **P2 — NS rollout for Nominet-held domains → Cloudflare.** The
  domains_cloudflare_rollout lane owns the Cloudflare half (zones, records,
  routes, token); this lane owns the Nominet half (EPP `domain:update`, or
  staged Online Services batches). **Ordering rule from P0: zone FIRST, NS
  second** — never repoint a domain whose zone does not exist and answer.
- **P3 — second TAG.** Chase status with the owner; when granted, customer
  registrations (VMB-017) move off `DESIGNCONSULT`, and the site-delivery lane's
  domain programme un-gates.
- **P4 — standing operations.** Renewals/expiries once the inventory exists;
  transfer-out execution when sales are agreed (the two-operation process);
  keeping the registrar page (webdesign.co.uk `/domains/`) honest against what
  the tag actually does.

## Cross-lane division (recorded in their NOTES too)

- **domains_cloudflare_rollout** — keeps Cloudflare + the three retail
  registrars (Spaceship/Dynadot/Porkbun). Everything Nominet moves HERE (owner
  directive 2026-09-02). Joint work (an NS cutover) splits: CF zone = theirs,
  Nominet NS = ours, and either lane may stage the other's half for the owner
  when the work is one incident (as P0 was).
- **site_delivery_and_editor** — consumes VMB-016/017/018 and the second TAG;
  the clients and tag policy live here.
- **idea_uk_vm_site** — the `box/` scripts stay at their cited paths; their
  Nominet function is this lane's responsibility now.
- **portfolio_positioning** — the remake programme's cutovers land here as asks
  (advertise.co.uk was the first); they own the sites, we own the delegation.
- **webdesign_couk** — owns the `/domains/` registrar page artefact; this lane
  owns its factual accuracy about the tag.
