# 090 — idea.uk: a visitor can choose the IP they are rate-limited as

**Filed** 2026-07-26, from the idea.uk VM-site workstream. Found by following
`bugs_open/089` — the same session, one endpoint apart.
**Class** spoofable trust boundary / a control that costs money to bypass.
**Status** **CLOSED 2026-07-26 — fixed, deployed and proven live.** Binary
deployed 18:44 UTC (rollback kept at `/opt/idea/idea.prev-2026-07-26-089only`).

**The live proof is the same request, before and after — one header, two answers:**

```
# 18:32 UTC, binary WITHOUT the fix
$ curl -s -H 'X-Forwarded-For: 203.0.113.77' 'https://idea.uk/order/success?o=ord_xff_probe&fake=1'
Jul 26 18:32:22 … (order "ord_xff_probe", ip 203.0.113.77)        ← the forged address

# 18:45 UTC, binary WITH the fix — identical request, identical header
$ curl -s -H 'X-Forwarded-For: 203.0.113.77' 'https://idea.uk/order/success?o=ord_xff_probe3&fake=1'
Jul 26 18:45:44 … (order "ord_xff_probe3", ip 2a02:c7c:f61f:ac00:…) ← my real IPv6 peer
```

**A near-miss worth recording.** The 18:29 deploy was built before this defect
existed as a fix — I found 090 at 18:32, *while verifying 089 on the box* — so
the first deployed binary carried 089 only. Re-running the forged-header probe
after that deploy still returned `203.0.113.77`, which is the only reason it was
caught before being written up as shipped. **The deploy that closes a bug is not
the deploy that happened to precede it.** Discriminating marker for this one:
`grep -ac "X-Real-IP" /opt/idea/idea` → 1 (0 in both earlier binaries), the
string literal the fix introduces.
**Scope** the idea.uk tool only
(`docs/agent_docs/docs024_key_docs_latest/idea.uk/golang_files/`). Not chassis code.

## Symptom

Every per-IP control in the tool is keyed on an address the caller supplies.
Send `X-Forwarded-For: <anything>` and that is the identity you are metered,
recorded and (in future) blocked as.

**Proven against production**, not reasoned about:

```
$ curl -s -H 'X-Forwarded-For: 203.0.113.77' 'https://idea.uk/order/success?o=ord_xff_probe&fake=1'
# box journal:
Jul 26 18:32:22 idea1 idea[106548]: orderSuccess: refused fake=1 payment shortcut
  under *main.StripeProvider (order "ord_xff_probe", ip 203.0.113.77)
```

`203.0.113.77` is TEST-NET-3 — it can never be a real client. The service
recorded it verbatim.

## Root cause

`audience_check.go:clientIP` took the **first** `X-Forwarded-For` entry:

```go
if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
	// XFF is comma-separated; first entry is the original client.
	if i := strings.IndexByte(xff, ','); i > 0 {
		return strings.TrimSpace(xff[:i])
	}
	return strings.TrimSpace(xff)
}
```

The comment is true of the *internet's* convention and false of *our* deployment.
nginx forwards with `$proxy_add_x_forwarded_for`
(`/etc/nginx/snippets/proxy_tool.conf:13`), which **appends** the real peer to
whatever arrived. So a forged header becomes `203.0.113.77, <real ip>` and the
code reads the forged half. The trustworthy value is the **last** entry — or
`X-Real-IP`, which the same snippet sets with `proxy_set_header` and therefore
replaces rather than merges.

## What it costs

| Control | Keyed on | Effect of the bypass |
|---|---|---|
| `/audience-check` taster limiter — 3/hour, 20/day | `clientIP` | **Unbounded LLM spend.** The taster is the only endpoint open to the world with no auth and no payment, and it makes a real model call per run (~£0.02 each, `audience_check.go:6`). Rotate the header and the cap is gone. |
| `/request` intake limiter — 5/hour, 15/day | `clientIP` | Unbounded spam orders. 60 spam rows from a June flood already sit in `orders.json`. |
| `Order.IP` captured at intake | `clientIP` | Captured expressly "so a future IP block list can be seeded from real orders" (`store.go`). Every value is attacker-chosen, so such a list would be worthless — and could be poisoned into blocking innocent addresses. |

**Bounded by** nginx's own `limit_req_zone $binary_remote_addr zone=idea_rl:10m
rate=10r/s` + `burst=20` (`sites-enabled/idea.conf:26`), which keys on the real
peer and **is not spoofable**. So the box cannot be flooded off the network — but
10 requests/second is a flood limit, not a spend limit. The control protecting
*money* is the one that fails.

No evidence it has been exploited: the June spam flood is all `requested` rows
with no taster involvement.

## The fix (committed, awaiting deploy)

Two rules, in `clientIP`:

1. **Believe forwarding headers only from a peer that is actually our proxy**
   (loopback or private). In production every request arrives from nginx at
   `127.0.0.1` — `proxy_pass http://127.0.0.1:8080`, and the tool's port is
   firewalled (ufw allows only OpenSSH/80/443, and `curl http://116.203.204.115:8080/health`
   from outside times out). A direct caller's headers are user input.
2. **Within them, take the value our proxy wrote**: `X-Real-IP` first, else the
   **rightmost** `X-Forwarded-For` entry. Never the first.

Also replaces `strings.LastIndexByte(addr, ':')` with `net.SplitHostPort`, which
was returning IPv6 peers wrapped in brackets — the order placed during this
session came from an IPv6 client, so that is the live shape, not a hypothetical.

## The test that had to be corrected — and why it matters

`request_hardening_test.go:51` asserted the defect in so many words:

```go
if o.IP != "203.0.113.7" {
	t.Errorf("want IP 203.0.113.7 (first XFF entry), got %q", o.IP)
}
```

It was written with the hardening work in July and passed every run since. The
spoofable behaviour was not an oversight that slipped past the tests — it was
**pinned in place by one**. Corrected in place with a dated note rather than
deleted, so the record shows what was believed.

`client_ip_test.go` adds seven cases: the exact live spoof shape, multi-hop,
`X-Real-IP` winning over a forged XFF, the clean single-entry case, an untrusted
direct peer, an IPv6 peer, and a limiter test that rotates forged headers and
asserts the 4th call is still refused.

**Induced against pre-fix source** (`git show HEAD:…/audience_check.go` over a
scratch copy of the module) — all five fail, e.g.:

```
client_ip_test.go:63: clientIP = "203.0.113.77", want "198.51.100.9"
  — a caller can choose its own rate-limit key
client_ip_test.go:87: clientIP = "[2a02:c7c:f61f:ac00::1]", want 2a02:c7c:f61f:ac00::1
```

Post-fix: all pass, full suite green, `go vet` clean.

## How to verify once deployed

```bash
curl -s -o /dev/null -H 'X-Forwarded-For: 203.0.113.77' \
  'https://idea.uk/order/success?o=ord_xff_probe2&fake=1'
ssh root@116.203.204.115 'journalctl -u idea --since "-2 min" | grep ord_xff_probe2'
# expect the refusal line to carry YOUR real address, not 203.0.113.77
```

## The transferable pattern (added to 016b §9)

**`X-Forwarded-For`'s "first entry is the client" convention is a statement about
the whole internet, not about your deployment.** With exactly one trusted proxy
that appends, the first entry is the only part an attacker controls and the last
is the only part you can trust. Any per-IP control — rate limit, block list,
audit trail, geo rule — inherits its worth from that choice. Check which end your
proxy writes (`$proxy_add_x_forwarded_for` appends; `X-Real-IP` replaces) and
gate on the peer address before believing any of it.

## Related

- `bugs_open/089` — the payment bypass, same service, found first; this was found
  while proving 089's fix on the live box.
- Workstream: `docs/agent_docs/docs024_key_docs_latest/idea_uk_vm_site/`
  (`RUNNING_NOTES` §X.19).
- Closes out, by refutation, the handoff's open item *"Real-client-IP in nginx —
  idea.uk is behind Cloudflare"*: it is not. Nameservers are Hetzner's, `idea.uk`
  resolves straight to `116.203.204.115`, and responses carry no `cf-ray`. No
  `set_real_ip_from` is needed; the real defect was in the Go, not the nginx.
