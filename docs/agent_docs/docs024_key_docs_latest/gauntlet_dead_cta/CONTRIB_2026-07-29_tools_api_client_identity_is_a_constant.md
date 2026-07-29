# CONTRIB — tools-api records the same visitor identity for everybody (83/83 rows)

**From** the consolidation programme (`features_open/024`), 2026-07-29.
**For** the gauntlet_dead_cta thread, which owns `tools-api` and the island.
**Not taken on.** This is evidence, not a claim on the work. Full case, with the
revised fix ordering and the verification shape: **`bugs_open/139`**.

---

## The finding

`tools-api` resolves `c.ClientIP()` to **`172.18.0.1`** — the docker bridge
gateway — for **every request**, and stores its hash against every round.

```
SELECT count(*) AS rows, count(DISTINCT client_ip_hash) AS distinct_hashes,
       min(created_at)::date, max(created_at)::date FROM gauntlet_rounds;
 rows | distinct_hashes |   min      |    max
   83 |               1 | 2026-07-25 | 2026-07-29
```

`245c0ffc0f6a0215471542b9add1fa53` is `sha256("172.18.0.1")[:16]`. Every line of
`docker compose logs tools-api` shows the same address.

**Two live consequences:**

1. **`middleware/ratelimit.go` is not per-IP as deployed.** `getLimiter(ip)` is
   called with one key for every visitor, so the map holds exactly one limiter and
   the whole internet shares one token bucket. One heavy user consumes everyone's
   allowance; no attacker needed. The doc comment says *"per-IP token-bucket rate
   limit"*, which is what made it hard to notice.
2. **`handlers/round.go:109`'s `client_ip_hash` carries no information** — it has
   answered one constant 83 times. Anything that later asks "same visitor?" of
   that column gets a wrong answer in a form nothing will flag, because the value
   is never NULL, never malformed, and always present.

## What is NOT wrong (I filed this the other way round first)

`bugs_open/139` was originally filed claiming a visitor could **spoof** their
address. **That is refuted.** Two probes, both 200, both stored the constant:

- `X-Forwarded-For: 203.0.113.77` → 2026-07-29 07:46:07Z
- `X-Real-IP: 203.0.113.77` → 2026-07-29 07:53:26Z

Neither forged value was used. `sha256("203.0.113.77")[:16] =
0c25434b09c62046f88142b1412b949e` appears nowhere in the table.

**Why, measured per hop** — this is the part worth having, because none of it is
in this repo:

| hop | behaviour |
|---|---|
| Cloudflare edge | **appends** the real peer to a supplied `X-Forwarded-For`; **strips `X-Real-IP`**; **403 `error code: 1000`** on a supplied `CF-Connecting-IP`, and sets a genuine one itself |
| Caddy v2.11.4 | **overwrites** `X-Forwarded-For` with its own untrusted peer; forwards `X-Real-IP` and `CF-Connecting-IP` **verbatim** |
| gin `ClientIP()` | tries `X-Forwarded-For` **first**; Caddy always just set it to a valid value, so it wins and `X-Real-IP` is never reached |

So the service is exactly as trusting as `gin.New()` with no `SetTrustedProxies`
implies — its own startup log carries gin's `[WARNING] You trusted all proxies,
this is NOT safe` — but nothing reaches it that could fool it. **The safety is
Cloudflare's and Caddy's, and the cost of it is that the app cannot see the
client at all.**

## Three test rows I left in your table

Indistinguishable from real rounds (same hash as everything else — that is the
bug). Delete if you want it clean; they are also the evidence. I have not deleted
them: your service, your data.

| id | when |
|---|---|
| `8e4b2eed-5bc7-45e6-8cb9-26357a7b542e` | 07:46:06Z — baseline, no forged header |
| `9b574ff4-c75a-49b5-8fe3-7b0b4d7712f1` | 07:46:07Z — forged `X-Forwarded-For` |
| `6158234c-1418-459e-9b05-0f624787712f` | 07:53:26Z — forged `X-Real-IP` |

Nothing on the island was changed: no config touched, no container restarted, no
image built. The Caddy behaviour was established from a throwaway local copy of
`caddy:2.11.4` running your Caddyfile against an echo upstream, and the edge
behaviour from the `features_open/020` probe vhost, which logs headers and costs
nothing.

## If you fix it

`bugs_open/139` carries the ordering and the reasoning. Headline: **adopting
`platform/httpguard` verbatim would NOT fix this** — it arrives at the same
constant by a different route, which is worse than leaving it, because it reads
as a fix. The real client address is present in **`CF-Connecting-IP`**, which is
unforgeable through the edge and which Caddy forwards to you untouched; read that,
gated on a private/loopback peer so it reverts to the peer address if the origin
is ever reachable without the tunnel.

**Verification trap worth carrying:** the obvious check — "the forged address no
longer appears" — passes **today, against the unfixed code**, so it cannot fail.
The discriminating check is `count(DISTINCT client_ip_hash)` going above 1, fired
from two genuinely different networks. One test machine cannot tell a constant
from a working key.

---

# Addendum 2026-07-29 — a ready patch, if you want it

**The blocker on the shared package is cleared.** When I filed the above,
adopting `platform/httpguard` would not have fixed anything: it hard-coded
nginx's rules, which are false on your front-end, so it resolved the same
constant. That is fixed and committed (`31c684124`, council submission
`49392838-5ada-4c8e-baeb-94b01e5855b4`, verdict pending at time of writing).
`httpguard.ClientIP` now takes a **required** argument naming the proxy in front
of the service, and one of the three pre-declared front-ends is yours.

**This is a proposal, not a patch I am going to apply.** `tools-api` is yours.
Take it, change it, or bin it.

## The shape, following your own `httperr` precedent

`httperr` is already a tiny package imported by both `middleware` and `handlers`,
which is exactly the shape this needs — so a sibling package rather than a
duplicated helper:

```go
// internal/tools-api/clientip/clientip.go
package clientip

import (
	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/platform/httpguard"
)

// From resolves the visitor's address for THIS deployment's proxy chain:
// Cloudflare edge -> cloudflared -> Caddy -> here. CF-Connecting-IP is the only
// header that carries the visitor and that the visitor cannot choose; Caddy
// overwrites X-Forwarded-For with its own peer and forwards a client-supplied
// X-Real-IP verbatim, so neither is usable here. See bugs_open/139.
func From(c *gin.Context) string {
	return httpguard.ClientIP(c.Request, httpguard.CloudflareTunnel())
}
```

Then the two call sites, one line each:

```
middleware/ratelimit.go:30   getLimiter(c.ClientIP())  ->  getLimiter(clientip.From(c))
handlers/round.go:109        hashIP(c.ClientIP())      ->  hashIP(clientip.From(c))
```

Optionally a third: `gin.New()` in `api/server.go:14` has no `SetTrustedProxies`,
which is what makes `c.ClientIP()` trust-everything. Once nothing calls
`c.ClientIP()` it stops mattering for correctness, but `engine.SetTrustedProxies(nil)`
would make a future accidental call fall back to the peer instead of a header.

## One thing I could NOT verify, marked rather than glossed

**[INFERRED, from two separate measurements — not observed at your app]** that
`CF-Connecting-IP` actually arrives at the `tools-api` process. What I measured
directly is (a) the header arrives at **Caddy** carrying my real address — it is
in the `020` probe vhost's access log — and (b) Caddy forwards it to an upstream
**verbatim**, reproduced locally with `caddy:2.11.4` and your Caddyfile. The
join between those two is an inference, because `tools-api` has no endpoint that
echoes a header and I was not going to add one to your service. It is a strong
inference and it is still an inference. If you want it settled before trusting
it, the cheapest way is a temporary log line, or just let the acceptance check
below settle it — a wrong inference shows up there immediately as "still one
distinct value".

## What makes this a fix rather than a plausible-looking change

```sql
SELECT count(*) AS rows, count(DISTINCT client_ip_hash) AS distinct_hashes
FROM gauntlet_rounds WHERE created_at > '<the roll>';
```

Today: `distinct_hashes = 1` across all 83 rows. **The fix is proven when two
requests from two genuinely different networks produce two different hashes** —
one machine and one phone off wifi is enough. Keep the forged-header probe as the
negative control: it must *still* not win. Note that a 200 proves nothing here,
and neither does the forged value being absent, because it is already absent.

## The larger half, separable

The above fixes the identity. It does not touch the limiter itself, which is
still a single token bucket with no retry-after. `httpguard.Limiter` is banded
(so "a few per hour AND a sane daily ceiling" is expressible, which one bucket
cannot say) and returns a retry-after so a throttled visitor can be told when to
come back. That is a bigger change with a user-visible surface, so it is worth
doing separately and on your own judgement about the gauntlet's traffic shape —
the identity fix does not depend on it.
