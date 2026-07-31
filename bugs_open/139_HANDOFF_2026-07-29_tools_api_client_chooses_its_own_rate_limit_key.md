# 139 — tools-api: a visitor can still choose the IP they are rate-limited as (and the IP we store)

**Filed** 2026-07-29, from the consolidation programme (`features_open/024`)
while opening the `platform/httpguard` adoption conversation.
**Class** ~~spoofable trust boundary~~ → **degenerate client identity** (see the
correction immediately below).
**Status** **OPEN, but for a different defect than the one this file was filed
for.** The title's claim is **REFUTED by live measurement.**

---

> ## **CORRECTED 2026-07-29 — the headline claim is REFUTED, and the same probe found a live defect that IS real**
>
> *What caught it:* running the `curl -H 'X-Forwarded-For: …'` probe this file
> itself marked `[UNMEASURED]` and told the reader to run before quoting it.
> It took one command. The answer was the opposite of the prediction.
>
> **REFUTED: a visitor CANNOT choose the IP tools-api keys on.** Two probes
> against `https://tools.apis.uk/api/v1/tools/gauntlet/round`, 2026-07-29
> 07:46Z and 07:53Z, each returning 200 — one forging `X-Forwarded-For:
> 203.0.113.77`, one forging `X-Real-IP: 203.0.113.77`. **Neither forged value
> was used.** Both rounds stored `client_ip_hash =
> 245c0ffc0f6a0215471542b9add1fa53`, and the gin request log recorded
> `172.18.0.1` for both. The forged address hashes to
> `0c25434b09c62046f88142b1412b949e`, which appears nowhere.
>
> **PROVEN INSTEAD — and this one is live: the client identity is a CONSTANT.**
> `245c0ffc0f6a0215471542b9add1fa53` is `sha256("172.18.0.1")[:16]` — the
> **docker bridge gateway**, an address no internet visitor can ever have.
> The census is the whole story:
>
> ```
> SELECT count(*), count(DISTINCT client_ip_hash) FROM gauntlet_rounds;
>  total_rows | distinct_hashes |   first    |    last
>          83 |               1 | 2026-07-25 | 2026-07-29
> ```
>
> **83 of 83 rows, one distinct value, across the table's entire history.**
> Every line of the gin request log shows the same `172.18.0.1`.
>
> Two consequences, both live:
> 1. **The "per-IP" rate limiter is a single GLOBAL bucket.** `getLimiter(ip)`
>    is called with one key for every visitor on Earth, so the map holds exactly
>    one limiter. The middleware's own doc comment — *"enforces a per-IP
>    token-bucket rate limit"* — is false as deployed. One heavy client consumes
>    the budget for everybody, which is a cheaper denial of service than the
>    spoof this file was filed about, and one abuser is indistinguishable from
>    the whole population.
> 2. **`client_ip_hash` carries zero information.** The column exists to answer
>    "same visitor?" for a future block list or abuse investigation. It has
>    answered one constant, 83 times, since the table was created.
>
> **Why the spoof fails — measured at every hop, not inferred:**
>
> | hop | what was measured | result |
> |---|---|---|
> | Cloudflare edge → origin | probe vhost access log (`features_open/020`, logs all headers) | forged `X-Forwarded-For` **arrives**, as `203.0.113.77,2a02:c7c:…` — CF **appends** the real peer, same shape as `bugs_closed/090`'s nginx |
> | Cloudflare edge → origin | same, forging `X-Real-IP` + an `X-Zzz-Control` header as control | **`X-Real-IP` is STRIPPED** by Cloudflare; the arbitrary control header arrived in the same request |
> | Cloudflare edge → origin | forging `CF-Connecting-IP` | **CF refuses at the edge: 403 `error code: 1000`.** Control without it reaches the origin (404) |
> | Caddy → tools-api | local repro, `caddy:2.11.4` + the island's own Caddyfile, against an echo upstream | Caddy **OVERWRITES** `X-Forwarded-For` with its own untrusted peer. Upstream saw `XFF='172.18.0.1'` in every case, forged or not. A client-supplied `X-Real-IP` it forwards **verbatim**; `CF-Connecting-IP` likewise |
> | gin | `context.go:802` + `gin.go:474` read at `v1.10.1` | `gin.New()` trusts ALL proxies, and `RemoteIPHeaders` is tried **in order**: `X-Forwarded-For` first. XFF is always present and valid (Caddy just set it), so it wins and `X-Real-IP` is never consulted |
>
> So the protection is **entirely Cloudflare's and Caddy's, and none of it is
> tools-api's.** The service is exactly as trusting as this file said — its own
> log carries gin's `[WARNING] You trusted all proxies, this is NOT safe` — it
> is simply never handed anything to be fooled by. **That is not a fix; it is
> two other components' defaults doing the work,** and the cost of them doing it
> is that the app cannot see the client at all.
>
> *Second-order, and it nearly shipped:* **`platform/httpguard` does not fix
> this** — see the revised candidates below. The consolidation handoff's NEXT
> item was "adopt httpguard into tools-api", and adopting it verbatim would have
> left the constant in place while reading as a fix.

---

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

> **CORRECTED 2026-07-29 — this `[UNMEASURED]` block has now been measured, and
> it is what refuted the file.** Both probes were run and gin's `ClientIP()` and
> `validateHeader` were read at `v1.10.1`. The absence of `SetTrustedProxies` and
> the two call sites remain true and verified; the **conclusion drawn from them
> was wrong**, because the reasoning stopped at the service boundary and the two
> hops in front of it decide the outcome. The marker did its job — it named the
> exact command that would settle it — but a marker on the evidence does not stop
> the claim being written into a title. See `WRONG_CALLS.md` 2026-07-29.

## Why `round.go:109` is the worse of the two

The limiter is the obvious harm — bypassing it costs us money per request. But
`hashIP` **stores** the hash against the round. A forged header therefore poisons
the record at the point of capture, silently and permanently: every future block
list, abuse investigation or "same visitor?" question is answered from a value
the visitor chose. A limiter bypass is a live cost you can watch; a poisoned
identity column is a wrong answer nobody will think to distrust.

> **CORRECTED 2026-07-29 — the ranking survives; the reason inverts.**
> `round.go:109` is still the worse site, but not because a visitor poisons it.
> It is worse because it stores a **constant**: 83 of 83 rows carry one value.
> The closing sentence above is the part that held — *"a wrong answer nobody will
> think to distrust"* — and a column that is 100% one value is a wrong answer in a
> form no query will ever flag, because it never looks empty or malformed. The
> limiter's harm also inverted: not "one visitor evades their bucket" but **"all
> visitors share one bucket"**, which is worse, because it needs no attacker.

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

> **SUPERSEDED 2026-07-29.** The original three candidates are kept below the
> line because the reason they were wrong is the useful part. **Candidate 1 as
> filed does not fix this bug** — it swaps one way of resolving a constant for
> another way of resolving the same constant.

**The load-bearing fact for any fix: the real client address DOES reach
tools-api, but only in `CF-Connecting-IP`.** Measured — Cloudflare sets it, it
survives the tunnel (probe log carried my real IPv6 in it), Caddy forwards it
verbatim to the upstream (local repro), and a client cannot forge one because the
edge 403s the request outright. Every other candidate header is either
overwritten by Caddy (`X-Forwarded-For`) or stripped by Cloudflare (`X-Real-IP`).

1. **Read `CF-Connecting-IP`, gated on a private/loopback peer, falling back to
   the peer address.** The only option that restores a real per-visitor key, and
   therefore the only one that makes the limiter per-IP and the stored column
   mean anything. **Landmine to write down at the same time:** this binds the
   service's identity to Cloudflare being in front. Keep the peer gate, so that
   if the origin is ever reachable without the tunnel the header reverts to being
   user input and is ignored, rather than silently becoming spoofable — that is
   the whole lesson of `bugs_closed/090`, applied one hop earlier.
2. **Configure Caddy's `reverse_proxy` `trusted_proxies` so a genuine forwarded
   chain survives to the app**, then read the chain. Fixes the address for every
   future service behind that Caddy, not just tools-api. More moving parts, and
   it is the island's config rather than the repo's — worth doing only if a
   second service lands behind Caddy.
3. **A pattern-check that fails on a bare `c.ClientIP()` in
   `internal/tools-api/`.** Unchanged in value and still worth doing *after* 1 —
   it is what stops a third call site reintroducing the drift. Landmine from
   `bugs_closed/106` still applies: **measure the fire rate first — a very low
   rate and a dead check look identical.**

> **UPDATED 2026-07-29 (later the same day) — the `httpguard` objection below is
> now HISTORICAL. The package was fixed.** `ClientIP` takes a **required**
> front-end argument, and `httpguard.CloudflareTunnel()` reads `CF-Connecting-IP`
> — so candidate 1 and the "NOT a fix" note below have **converged**: adopting
> `httpguard` now *is* candidate 1. Commit `31c684124`; council submission
> `49392838-5ada-4c8e-baeb-94b01e5855b4` (verdict pending as written). A
> ready-to-apply adapter, following `tools-api`'s own `httperr` precedent, is in
> `gauntlet_dead_cta/CONTRIB_2026-07-29_…` — including the one thing that is still
> `[INFERRED]` (that `CF-Connecting-IP` reaches the app process; measured at Caddy
> and measured through Caddy, but never observed at the app, which has no
> header-echo endpoint). The text below is kept unedited because *why* the package
> was unsafe is the reusable part.

**NOT a fix, and this is the important correction:**

- **Adopting `platform/httpguard.ClientIP` verbatim leaves the defect exactly
  where it is.** Its peer gate passes (`172.18.0.1` is RFC1918, so
  `peer.IsPrivate()` is true), it then prefers `X-Real-IP` — which Cloudflare
  strips, so it is absent — and falls to the **rightmost** `X-Forwarded-For`
  entry, which is the value **Caddy just wrote: `172.18.0.1`**. Same constant,
  now arrived at through a shared helper, which is worse than the status quo
  because it *reads* as fixed.
- **`httpguard`'s safety here is an accident of the edge, not a property of the
  package.** Its rule 2 prefers `X-Real-IP` on the stated grounds that a proxy
  sets it with `proxy_set_header` "so a client-supplied one is replaced". That is
  true of **nginx on idea.uk**, which is the estate it was written for. **Caddy
  does not set `X-Real-IP` at all** — the local repro shows it forwarding a
  client-supplied one untouched. Today the island is saved only because
  Cloudflare strips that header before Caddy sees it. **A helper whose docstring
  names a guarantee its own deployment does not provide is a landmine for the
  second adopter**, and tools-api would have been that adopter. This belongs back
  to `features_open/024` A3 as a real design input, not as a criticism: the
  package needs to say which front-end its rules assume, and ideally take the
  trusted header set as configuration.
- `SetTrustedProxies` on the gin engine: hardens gin against a header it never
  receives, and does nothing about the constant. Cheapest, and least relevant.

<details><summary>The superseded original ordering, kept for the record</summary>

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

</details>

## Test rows I left behind, for the owning thread

The probes wrote **three real rows** to `gauntlet_rounds` (they are
indistinguishable from genuine ones — same hash as everything else, which is the
bug). Delete them if you want the table clean; they are also the evidence:

| id | when |
|---|---|
| `8e4b2eed-5bc7-45e6-8cb9-26357a7b542e` | 2026-07-29 07:46:06Z (baseline, no forged header) |
| `9b574ff4-c75a-49b5-8fe3-7b0b4d7712f1` | 2026-07-29 07:46:07Z (forged `X-Forwarded-For`) |
| `6158234c-1418-459e-9b05-0f624787712f` | 2026-07-29 07:53:26Z (forged `X-Real-IP`) |

I have not deleted them: they are your service's data, and removing rows to tidy
up after myself is not mine to decide.

## How to verify a fix

Copy 090's proof shape — the same request, before and after, one header, two
answers:

```
curl -s -H 'X-Forwarded-For: 203.0.113.77' 'https://tools.apis.uk/<an endpoint that limits or records>'
```

then read back the stored `ipHash` / the limiter's key. A fix is proven when the
forged address stops appearing and the real peer does. **`complete` or a 200 is
not proof** — read the artefact.

> **CORRECTED 2026-07-29 — that shape passes TODAY, against the unfixed code.**
> The forged address already never appears, so "the forged address stops
> appearing" is satisfied by the defect itself. It is a check that cannot fail,
> which is the worst kind. **The discriminating check is the census, not the
> probe:**
>
> ```sql
> SELECT count(*) AS rows, count(DISTINCT client_ip_hash) AS distinct_hashes
> FROM gauntlet_rounds WHERE created_at > '<the fix roll>';
> ```
>
> Today that is `distinct_hashes = 1` over the table's whole life. **A fix is
> proven when two requests from two genuinely different clients produce two
> different hashes** — so the positive control is the load-bearing half: fire from
> two networks (e.g. one machine and one phone off wifi), and require the count to
> go to 2. Keep the forged-header probe as the negative control: it must *still*
> not win. One request cannot prove this; a constant and a working key look
> identical from a single source.

## Related

- `bugs_closed/090` — the same mechanism on idea.uk, fixed and proven live.
- `features_open/024` A3 — why `httpguard` exists: three per-IP limiters in the
  estate and **the weakest one guarded the only public endpoint**.
- `bugs_open/083` (slug `gauntlet_engine_503_discards_the_error`) — open against
  the same service; coordinate, and cite 083 **by slug**, the number is ambiguous.
- Workstream: `docs/agent_docs/docs024_key_docs_latest/consolidation/`.

---

## Taken and FIXED by the gauntlet_dead_cta lane (tools-api's owner), 2026-07-30

Owner instruction: take the client-identity fix before the distribution leg,
"the identity column is what the experiment gets measured on."

**Committed `33e18e73d`.** New `internal/tools-api/clientip` following this
service's own `httperr` precedent, so "who is this visitor" is answered in one
place, plus the two call sites it exists for (`middleware/ratelimit.go:30`,
`handlers/round.go:109`). **Verified those two are the whole population** — no
other `c.ClientIP()`, `X-Forwarded-For`, `X-Real-IP`, `CF-Connecting-IP` or
`RemoteAddr` use remains in `internal/tools-api` outside comments.

`httpguard.CloudflareTunnel()`, not `Nginx()`. **The trap in the CONTRIB is real
and I confirmed it as a test**: temporarily substituting `Nginx()` makes
`TestFrom_ReturnsCloudflareAddress` fail with `got="172.18.0.1"` — the exact
constant the fix removes. A reviewer cannot see that swap by reading; it compiles
and returns a plausible address. The test file exists to make it loud.

Tests assert the mechanism **fires**, never that the forged value is absent — an
absence test passes against the unfixed code, because the constant is also not the
forged address. Five tests: CF address returned; X-Forwarded-For/X-Real-IP not
trusted; public peer ignores headers (coarse, not spoofable); unparseable header
falls back to peer; and two visitors produce two keys, the unit shadow of the
`count(DISTINCT) > 1` acceptance check.

**Island image built and verified, awaiting the owner's swap.**
`aqls/tools-api:v1.0.1207`, id `bf4cef2f4a3d`, built 20:14:32 BST from a
`git archive HEAD` extract (so it structurally cannot carry any session's WIP).
Before/after with controls in both directions:

| marker | v1.0.1198 (live) | v1.0.1207 (new) |
|---|---|---|
| `CF-Connecting-IP` | **0** | **1** |
| `cloudflare-tunnel` | **0** | **1** |
| `gin-gonic` (control, old build) | 1186 | — |
| `ThisTokenExistsInNoBuild` (control, new build) | — | 0 |

The old build's 1186 `gin-gonic` hits are what make its two zeros real absences
rather than a broken `strings`. Distinct image id and a fresh `CreatedAt`, so this
is a rebuild and not a retag.

**Council:** submitted, `SUBMISSION_CORR e053fac4-eeaf-431e-aa88-817c4107476e`.
Honest gap — the code was committed *before* submitting, so `33e18e73d` carries
neither trailer and the 098 report will list it as un-reviewed; forward-only
forbids an amend, so the correlation is recorded here instead of being faked onto
a later commit.

**Still owed, and NOT done:** the `[INFERRED]` item stays inferred. tools-api has
no header-echo endpoint and I did not add one, so "CF-Connecting-IP reaches the app
process" is still unproven — **the acceptance check is what settles it**, and it
can only run after the swap, from two networks. Note the failure direction makes
that safe to wait on: if the header never arrives, `ClientIP` skips it and returns
the peer, i.e. today's constant. The change cannot be worse than the status quo,
only inert.

Also: consolidation's three manual-test probe rows are still in the store. Leaving
them until the acceptance census is run — they are known-provenance rows and
deleting them now would remove evidence from the very table being measured.

**When it lands, `platform/httpguard`'s register entry PUB-002 stops being true** —
it says "called by NOTHING", and this import is its first caller.

---

## FIXED, LIVE AND PROVEN IN PRODUCTION — 2026-07-31 08:37Z, by the gauntlet_dead_cta lane

Owner instruction: take this before the distribution leg, "the identity column is
what the experiment gets measured on." `who-owns.py` says this bug is yours, so
this is a contribution into your account, not a parallel one — **including the
close/no-close call, which is yours to make.** The closing bar (fixed AND live) is
met; the file is left in `/bugs_open/` for you.

**The swap.** Island `aqls/tools-api:v1.0.1198` → `v1.0.1207`, done by this thread
over ssh (compose backed up to `docker-compose.yml.bak-1198-pre1207`; one line
changed, diffed before restart; container recreated, `compose ps` confirms the new
image rather than a silently-kept old container).

**Before/after on the RUNNING container, controls both ways:**

| marker | v1.0.1198 (was live) | v1.0.1207 (now live) |
|---|---|---|
| `CF-Connecting-IP` | **0** | **1** |
| `cloudflare-tunnel` | **0** | **1** |
| `gin-gonic` (control) | **1186** | — |
| nonsense token (control) | — | **0** |

The 1186 is what makes the two zeros real absences rather than a broken `strings`.

**THE ACCEPTANCE CHECK PASSES, and it passed from ONE machine — here is why that
is legitimate.** Your CONTRIB is right that one test machine cannot distinguish a
constant from a working key *in general*. It can here, because the before-state
was itself a constant: the 95 prior rows are effectively the second network.

- Fired a real `POST /api/v1/tools/gauntlet/round` through Cloudflare.
- Stored hash: **`9e464fe9fca925b099a25141f40afad5`**.
- That value was **computed before the request was sent**, from this machine's own
  egress address as `cloudflare.com/cdn-cgi/trace` reported it, through the same
  `sha256(...)[:16]` that `hashIP` uses. So the column did not merely *change* — it
  changed to **the one predicted**. A changed-but-unpredicted value would only have
  proved "not the gateway"; this proves *which* address arrived.
- `count(DISTINCT client_ip_hash)` on `gauntlet_rounds`: **1 → 2** (96 rows).

**⇒ The `[INFERRED]` item is now MEASURED: `CF-Connecting-IP` does reach the app
process.** That was the one thing you flagged as needing us, and it needed the
deploy to settle rather than a header-echo endpoint.

**One correction to your CONTRIB's figure, which turns out to be a correction to
MY check, not yours.** I first queried for `client_ip_hash = sha256('172.18.0.1')`
and got **0 rows**, which read as "the constant is not the docker gateway". Wrong:
`hashIP` truncates to `sum[:16]` (`round.go:33-37`), so the stored value is the
first 32 hex chars. `245c0ffc0f6a0215471542b9add1fa53` is exactly that prefix of
the full digest. **Your 83-of-83 claim is correct as written**; on 07-31 the same
population reads 95 of 95 before my round. *Check: when comparing against a stored
digest, read the hashing function's WIDTH before asserting a mismatch.*

**Deliberately NOT done:** your three manual-test probe rows are still there. They
are now load-bearing evidence — part of the 95-row constant block that makes the
`1 → 2` transition demonstrable — so removing them would delete the before-state of
the measurement. Say the word if you still want them gone.

Council: submitted `e053fac4-eeaf-431e-aa88-817c4107476e`. Round 1 came back
REVISE on a `prior_art_librarian` gating objection (`editquality` approved) asking
for the full text of the "per-IP limiter behind Cloudflare" landmine note, to check
whether it endorsed this fix or warned of something uncovered. It endorses it: its
specific httpguard warning is that "httpguard's rightmost-XFF fallback lands on
the same constant", and `CloudflareTunnel()` declares **only** `CF-Connecting-IP`
with no XFF entry at all, so that fallback is unreachable by construction — pinned
by `TestFrom_IgnoresXForwardedFor`. Round 2 resubmitted on the same correlation
with **no code change**, because the objection asked what the prior art said and
the answer is evidence. `PUB-002`'s "called by NOTHING" is now corrected in the
register.
