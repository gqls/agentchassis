# 139 — tools-api: a visitor can still choose the IP they are rate-limited as (and the IP we store)

**Filed** 2026-07-29, from the consolidation programme (`features_open/024`)
while opening the `platform/httpguard` adoption conversation.
**Class** spoofable trust boundary / a control that costs money to bypass.
**Status** **OPEN.** The defect is live. **The fix is already written, tested and
council-APPROVED — it is simply not called.**

> **This is not a reopening of `bugs_closed/090`.** 090 is the same *mechanism*
> and is correctly closed: it was fixed, deployed and proven live **on idea.uk**,
> a different service on a different box. `tools-api` was never touched by it and
> is a second, still-live instance. Filed separately so 090's closure stays
> honest.

**Owner note — I am NOT taking this on.** `tools-api` belongs to the
**gauntlet_dead_cta** thread (`scripts/who-owns.py 083` → that lane; it also owns
the island). This file is the evidence half of a conversation, not a claim on the
work. Contribute here rather than forking a second account.

---

## The defect

`tools-api` keys its per-IP controls on gin's `c.ClientIP()` at **two** sites:

| site | what the key is used for |
|---|---|
| `internal/tools-api/middleware/ratelimit.go:30` | `getLimiter(c.ClientIP())` — the rate-limit bucket |
| `internal/tools-api/handlers/round.go:109` | `ipHash := hashIP(c.ClientIP())` — **persisted** against the round |

And the server is constructed with **`gin.New()` and no `SetTrustedProxies` call**
(`internal/tools-api/api/server.go:14`), so gin's default trust-everything
posture applies and the resolved address comes from the forwarding headers the
caller supplied.

**The mechanism, proven against production for the sibling case** (090, verbatim
from that file): nginx forwards with `$proxy_add_x_forwarded_for`, which
**appends** the real peer to whatever arrived. A request carrying
`X-Forwarded-For: 203.0.113.77` therefore arrives as `203.0.113.77, <real ip>`,
and anything reading the **first** entry reads the value the client wrote for
itself. On idea.uk the refusal log recorded `203.0.113.77` verbatim; after the
fix the identical request logged the real IPv6 peer.

**[UNMEASURED — the one thing I did not verify]** I did not run the equivalent
`curl -H 'X-Forwarded-For: …'` probe against tools-api, and I did not read gin's
`ClientIP()` source to confirm its exact left-vs-right walk under the default
trusted-proxy set. What IS verified is the absence of `SetTrustedProxies` and the
two call sites. **Do the probe before quoting this as proven** — 090 shows the
probe is one command and settles it outright.

## Why `round.go:109` is the worse of the two

The limiter is the obvious harm — bypassing it costs us money per request. But
`hashIP` **stores** the hash against the round. A forged header therefore poisons
the record at the point of capture, silently and permanently: every future block
list, abuse investigation or "same visitor?" question is answered from a value
the visitor chose. A limiter bypass is a live cost you can watch; a poisoned
identity column is a wrong answer nobody will think to distrust.

## The fix already exists and has zero callers

`platform/httpguard` (`clientip.go`, `limiter.go`, `intake.go`) and
`platform/mailer` are **built, unit-tested and council-APPROVED (`6db59c8b`)**,
and `grep -rl 'agentchassis/platform/httpguard'` over every `.go` file in the
repo returns **nothing**. Re-measured 2026-07-29, not inherited from the handoff.

`httpguard.ClientIP` is deliberately `net/http`-only (no gin), because tools-api
is gin and idea.uk is net/http; the caller writes a three-line adapter. Its rule,
from the package doc:

1. Forwarding headers are believed **only** from a peer that is plausibly our own
   proxy (loopback or RFC1918). A direct caller's headers are user input.
2. Within those, take what **our** proxy wrote: `X-Real-IP`, else the
   **RIGHTMOST** `X-Forwarded-For` entry — never the first.

The package's own header already states the boundary this file exists to respect:
*"This package does NOT wire itself into any existing service. Adopting it in
tools-api is a change to another workstream's code (gauntlet thread,
`bugs_open/083` open against it) and belongs in a separate, coordinated commit."*

## Fix candidates, ordered by what closes the door

1. **Adopt `httpguard.ClientIP` at both sites** via a small gin adapter
   (`func clientIP(c *gin.Context) string { return httpguard.ClientIP(c.Request) }`).
   Closes both sites with one shared implementation and deletes the divergence.
   Makes the bad state harder to reach, but a *third* site could still call
   `c.ClientIP()` tomorrow.
2. **(1) plus a lint/pattern-check that fails on a bare `c.ClientIP()` in
   `internal/tools-api/`.** This is what makes the bad state unrepresentable
   rather than merely fixed — the estate already has `scripts/pattern-check.py`
   and this is a one-rule addition. Note the landmine from
   `bugs_closed/106`: **measure the fire rate first — a very low rate and a dead
   check look identical.**
3. `SetTrustedProxies` on the gin engine. Cheapest, and it does harden the
   limiter — but it leaves two implementations of "who is the client" in the
   estate, which is the drift `httpguard` was approved to end. Weakest by the
   ordering rule: it relies on operators remembering to keep the list right.

## How to verify a fix

Copy 090's proof shape — the same request, before and after, one header, two
answers:

```
curl -s -H 'X-Forwarded-For: 203.0.113.77' 'https://tools.apis.uk/<an endpoint that limits or records>'
```

then read back the stored `ipHash` / the limiter's key. A fix is proven when the
forged address stops appearing and the real peer does. **`complete` or a 200 is
not proof** — read the artefact.

## Related

- `bugs_closed/090` — the same mechanism on idea.uk, fixed and proven live.
- `features_open/024` A3 — why `httpguard` exists: three per-IP limiters in the
  estate and **the weakest one guarded the only public endpoint**.
- `bugs_open/083` (slug `gauntlet_engine_503_discards_the_error`) — open against
  the same service; coordinate, and cite 083 **by slug**, the number is ambiguous.
- Workstream: `docs/agent_docs/docs024_key_docs_latest/consolidation/`.
