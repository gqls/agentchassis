# SUMMARY 2026-09-04 — nominet_domain_management

## What we're trying to do

Own everything Nominet for this estate — the registrar tag, the EPP
connection, the domain inventory, registrations, transfers, and nameserver
changes for domains still at Nominet — as one lane, on the owner's explicit
directive ("take responsibility for everything nominet here"), matching the
shape already built for the other registrars (Dynadot, Porkbun, Spaceship).

## Where we've come from

Before this lane existed (2026-09-02), everything Nominet was scattered
across four other lanes: EPP access and the allowlist sat in
domains_cloudflare_rollout, the client scripts sat under idea_uk_vm_site's
`box/`, the inventory recipe sat in portfolio_positioning, and the
second-TAG question sat with site_delivery_and_editor. Nobody owned the
whole. The inventory itself had never actually been produced — only the EPP
login had ever been proven (2026-08-11), and every document since had cited
that as if the whole recipe worked.

## What we've done

- **Consolidated the client.** `scripts/domains/nominet.py` — probe, login,
  list, walk, check, info, set-ns, all behind one command, credentials never
  printed or on argv, transport tunnelled through the cluster so egress is
  always allowlisted. `register` deliberately refuses (that stays a separate,
  careful tool — it costs money).
- **Fixed a live incident on day one.** The owner's own Nominet nameserver
  batch had pointed four domains at Cloudflare with no zone created behind
  them — a real trap (a domain can look perfectly delegated at the registry
  while the edge refuses to answer). Built the recovery tool
  (`cf-zone-bootstrap.sh`), found a second wrinkle along the way (new
  Cloudflare zones on this account get a different nameserver pair than the
  older ones), fixed the Nominet side to match, and watched all four domains
  come back — including the first public launch of a rebuilt site
  (advertise.co.uk).
- **Built and proved the tag inventory, the hard way.** The first real
  attempt to list domains failed, and peeling that back found three separate
  bugs stacked on top of each other in the EPP client — two that failed
  loudly, and a third that would have failed *silently*, returning an empty
  list with no error at all. Fixed all three, added a permanent check against
  the registry's own count so that class of failure can't recur unnoticed,
  and ran the real walk: **1,606 domains**.
- **Delivered and corroborated.** Two other sessions working the same
  portfolio (domain valuation, Sedo listing) needed exactly this list;
  delivered it to both, and the valuation session's independent cross-check
  against a completely separate export agreed almost exactly — real
  confirmation, not just a clean-looking run.
- **Answered a standing owner question along the way**: which domains have
  genuinely expired versus merely sit under an unfamiliar registrar. Five
  confirmed gone; two undetermined for a structural reason (no public
  records for that domain ending); the rest accounted for.
- **One near-miss, caught before any harm**: the owner asked to "release" 50
  domains to the tag; the registry showed they were already there. Checking
  first — rather than building and firing a write based on the ask alone —
  surfaced that he actually meant a different Nominet operation (transfer of
  legal ownership, not the registrar record), which he then did himself
  through the correct channel.

## Where we are now

Every verb the client offers has been proven against the live registry:
reading, writing (the four-domain nameserver repoint), and the inventory
walk. The full estate is known and delivered. Nothing is currently broken or
waiting on a fix. The lane's own documents (PLAN, RUNBOOK, NOTES) are
up to date, two LANDMINES entries are filed for the bug classes found, and
the register entry for the client reflects everything proven.

## Where we're going

Two items sit open, both external (nothing to build): whether Nominet has
replied to the second-tag application (silent since 11 August), and — low
priority — reconciling the domain list against where each one is actually
parked, the way the retail-registrar inventories already have been. Standing
operational work (renewals, transfer-outs, keeping the registrar page
honest) continues as it comes up; nothing structural is owed right now.
