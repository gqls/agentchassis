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

## §2 The family client: `scripts/domains/nominet.py` (added 2026-09-02)

One consolidated client in the porkbun.py/dynadot.sh family style — credentials
from `~/.config/nominet/credentials`, never printed, never argv; transport
tunnels through `kubectl exec … openssl s_client` from the allowlisted cluster
egress by default (`--direct` for a plain socket from an already-allowlisted
box). Verbs: `probe` (credential-free) · `login` · `list YYYY-MM` ·
`walk [--months N]` · `check` · `info` · `set-ns … [--apply]` (dry-run
default). `register` is deliberately REFUSED there — VMB-017 keeps that verb
(it costs money and carries the registrant rulings).

**Proof state `[MEASURED 2026-09-02]`:** `--self-test` 15/15 (offline);
`probe` 2,527 B greeting through the pod tunnel. **Every credentialed verb is
UNEXERCISED** — the session classifier refuses them (correctly); the owner's
first `login` is their proof.

### The tag inventory (owner-run; P1)

```sh
# 1. allowlist test — the only thing that proves egress:
python3 scripts/domains/nominet.py login

# 2. the expiry walk — 120 months, because .uk registers up to 10 YEARS and a
#    12-month walk is structurally short (any multi-year registration whose
#    expiry falls outside the window is silently absent):
python3 scripts/domains/nominet.py walk --months 120 > all_domains.txt
grep -c '^DOMAIN' all_domains.txt
```

**Sanity-check against ~1,500** (owner estimate 2026-08-19). **The CSV export
from Online Services is still the better source** (checkable total, no EPP);
prefer it if the owner is at the console anyway.

Fallback (proven end-to-end with credentials 2026-08-19, unlike the new
client): the `epp.pl` stdin recipe —
`kubectl cp scripts/domains/epp.pl postgres-clients-0:/tmp/epp.pl`, then
`( set -a; . ~/.config/nominet/credentials; set +a; printf '%s\n%s\n' "$TAG" "$EPP_PASSWORD" ) | kubectl -n ai-persona-system exec -i postgres-clients-0 -- perl /tmp/epp.pl login` (or `list YYYY-MM`).

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

## §4 Per-domain EPP operations (owner-run)

Go-forward: the §2 client — `nominet.py check <names…>` / `info <domain>` /
`set-ns <domain> --ns alexis.ns.cloudflare.com --ns leah.ns.cloudflare.com
[--apply]` (dry-run default; host:create retry on 2303; verifies by re-reading).
The proven originals remain as fallbacks until the owner's first nominet.py runs:

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
