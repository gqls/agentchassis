# RUNBOOK — Nominet domain management

Every command here that logs in is **owner-run** (`! <command>` in a session
prompt): a session's permission classifier refuses to pipe credentials, and that
refusal is correct. Sessions run the read-only halves.

## §1 Credentials + transport (inherited, verified where dated)

| what | where | notes |
|---|---|---|
| TAG + EPP password | `~/.config/nominet/credentials` | `TAG=DESIGNCONSULT` + `EPP_PASSWORD=` lines, 0600. Exists 2026-09-02 |
| legacy password file | `~/.config/nominet/epp-password` | single line; superseded by the credentials file |
| endpoint | `epp.nominet.org.uk:700` | TLS; RFC 5734 framing (4-byte BE length incl. itself, then XML) |
| egress | cluster only, **IPv4 pin** | five allowlisted node IPs `134.213.168.26/.37/.44/.54/.56`; office line rotates BOTH families — never allowlist it |
| transport pod | `postgres-clients-0` | has openssl 3.0.20 + perl; client staged at `/tmp/epp.pl` (wiped by pod restart — re-copy, §2 step 0) |

**⚠ THE GREETING PROVES NOTHING** — Nominet serves its 2,531-byte greeting to
any IP. Only `LOGIN_CODE=1000` proves the allowlist. `LOGIN_CODE=2200` = egress
IP not allowlisted, NOT a wrong password. (LANDMINES; proven 2026-08-04.)

Cheap session-safe reachability probe (no credentials):

```sh
kubectl -n ai-persona-system exec postgres-clients-0 -- sh -c \
  "printf '' | timeout 15 openssl s_client -connect epp.nominet.org.uk:700 -4 -quiet 2>/dev/null | head -c 4000 | wc -c"
# expect ~2531
```

## §2 The tag inventory (owner-run; P1)

```sh
# 0. re-stage the client if the pod restarted:
kubectl -n ai-persona-system cp scripts/domains/epp.pl postgres-clients-0:/tmp/epp.pl

# 1. allowlist test — the only thing that proves egress:
( set -a; . ~/.config/nominet/credentials; set +a; printf '%s\n%s\n' "$TAG" "$EPP_PASSWORD" ) \
  | kubectl -n ai-persona-system exec -i postgres-clients-0 -- perl /tmp/epp.pl login

# 2. twelve-month expiry walk:
for m in 2026-09 2026-10 2026-11 2026-12 2027-01 2027-02 2027-03 2027-04 2027-05 2027-06 2027-07 2027-08; do
  ( set -a; . ~/.config/nominet/credentials; set +a; printf '%s\n%s\n' "$TAG" "$EPP_PASSWORD" ) \
    | kubectl -n ai-persona-system exec -i postgres-clients-0 -- perl /tmp/epp.pl list "$m"
done | grep '^DOMAIN' | cut -f2 | sort -u > all_domains.txt; wc -l all_domains.txt
```

**Sanity-check against ~1,500** (owner estimate 2026-08-19). A domain expiring
outside the window is silently absent — a short count means widen the walk, not
"the estate shrank". **The CSV export from Online Services is the better source**
(checkable total, no EPP); prefer it if the owner is at the console anyway.

Classification (session-safe, no credentials —
`portfolio_positioning/RUNBOOK_domain_inventory_and_classification.md` has the
full detail incl. the person-name extractor):

```sh
python3 scripts/domains/classify_nameservers.py all_domains.txt > classified.tsv
```

## §3 Registry reads (session-safe)

```sh
# delegation as the REGISTRY has it (bypasses every cache):
dig +norec NS <domain> @dns1.nic.uk
# a SERVFAIL from 1.1.1.1 with Cloudflare NS at the registry = dangling delegation (§5)
```

## §4 Per-domain EPP operations (owner-run clients, proven family)

- **check**: `idea_uk_vm_site/box/nominet-epp-domain-check.py` (VMB-016, read-only).
- **NS change**: `idea_uk_vm_site/box/nominet-epp-ns-change.py` (VMB-015) —
  dry-run default, `--apply` to execute.
- **register**: `idea_uk_vm_site/box/nominet-epp-domain-register.py` (VMB-017) —
  dry-run default; **`--apply` COSTS MONEY** (~£4/yr) and creates a registry
  object in the owner's name. Registrant = owner until sale (D1);
  `--registrant-from idea.uk` reuses a proven contact. Never run `--apply`
  without an explicit owner instruction naming the domain.
- **transfer-out** (sale): TWO operations — Registrant Transfer at the registry
  (£10+VAT, NOT over EPP, owner does it in Online Services) + the TAG change
  (free). Do not quote the one-operation version; it was corrected 2026-08-21
  (`458affaf7`).

## §5 NS cutover to Cloudflare — ordering, and the recovery when it is violated

**Zone FIRST, NS second.** A domain delegated to alexis/leah with no zone in the
account gets REFUSED at the edge and goes dark as resolver caches (2-day
delegation TTL) expire — while `dig` at the registry looks perfectly configured
(LANDMINES, "dangling delegation"; it bit four domains on 2026-09-02).

```sh
# session-safe: is the zone there and correctly shaped?
scripts/domains/cf-zone-bootstrap.sh --check <domain> [...]

# owner-run: create + wire + activate (idempotent top-up; safe on an existing zone)
scripts/domains/cf-zone-bootstrap.sh <domain> [...]
```

Then verify SERVING with a body property, never a status code — parked domains
200 on every path, and your own resolver holds the stale answer:

```sh
IP=$(dig +short @1.1.1.1 <domain> A | head -1)
curl -s --resolve "<domain>:443:$IP" "https://<domain>/" | head -c 300
```

A TLS `handshake failure` right after activation = Universal SSL still
provisioning (minutes) — test plain HTTP to separate it from a broken route.
`PUT /activation_check` returning 200 is acceptance, not activation — poll the
zone status.

## §6 Cloudflare token state (the other half's credential, for --check)

- `~/.config/cloudflare/portfoliotoken` — the read-write token; works
  (measured 2026-09-02).
- `~/.config/cloudflare/token` — **DEAD as of 2026-09-02** (`9109 Invalid
  access token` on every call). Was the read-only token. Anything still using
  it fails; use portfoliotoken or fix the token.
