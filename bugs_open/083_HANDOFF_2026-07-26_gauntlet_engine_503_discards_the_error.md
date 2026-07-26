# 083 — the Gauntlet engine returns a bare 503 and throws the reason away, so an intermittent failure cannot be diagnosed

**Filed:** 2026-07-26 · **By:** gauntlet_dead_cta (P4) · **Severity:** MEDIUM —
the flagship journey visibly fails for a visitor, but recovers on retry and the
page degrades honestly · **Status:** OPEN, not fixed

## 1. Symptom

`POST /api/v1/tools/gauntlet/position` and `.../defend` intermittently return
`503 {"error":"gauntlet opponent unavailable"}` /
`{"error":"gauntlet judge unavailable"}`. The same request retried usually
succeeds. Failures are **bursty rather than steady** — clustered in time, with
long clean runs either side.

Measured today, all against the live island through Cloudflare:

| window | calls | failures |
|---|---|---|
| first probe | 1 `/defend` | 1 failed at 24.8 s, then again at 25.0 s on the same round, then **succeeded** on the third attempt at 14.3 s |
| 5-round curl sample | 5 `/position`, 5 `/defend` | 0 |
| live browser verification, ~15:2x | 2 `/position` (one per profile) | **2 failed** — both recovered on the automatic retry |
| 8-round curl sample, immediately after | 12 `/position`, 11 `/defend` | 0 |

So: **not reproducible on demand, and not rare enough to ignore** — it hit the
first attempt of both profiles during one verification window and nothing either
side of it. [UNMEASURED] no honest overall rate can be quoted from this; the
clustering is the finding, not a percentage.

Latency when it succeeds (23 samples): `/position` 8.2–12.9 s, `/defend`
9.9–23.4 s. Failures land at ~25 s, i.e. **longer than a typical success**.

## 2. Why it cannot be diagnosed from outside — this is the actionable part

`internal/tools-api/handlers/defend.go:89-93` (and the equivalent in
`position.go:83`):

```go
text, err := client.GenerateText(ctx, prompt, map[string]interface{}{})
if err != nil {
    httperr.JSONError(c, http.StatusServiceUnavailable, "gauntlet judge unavailable")
    return
}
```

`err` is **discarded** — not logged, not wrapped, not counted. There is no
request logging on the island either (`docker compose logs tools-api` shows only
gin's startup banner), so nothing anywhere records *why* the call failed. Every
distinct upstream condition arrives at the client as the same opaque 503:

- an HTTP error from `api.anthropic.com` (429, 529 overloaded, 500) —
  `aiservice/anthropic.go:150` returns `API request failed with status %d`
- a network/timeout failure — note `anthropic.go:63` builds
  `&http.Client{}` with **no timeout at all**, so nothing here bounds the call
- **a truncated completion** — `anthropic.go:209` returns a non-nil
  `TruncatedError` when `stop_reason == "max_tokens"`, and the handler cannot
  tell that apart from a real failure. Default `max_tokens` is 2048
  (`anthropic.go:72`) and both handlers pass empty options, so nothing raises it.

> **A note on a theory that does NOT fit, recorded so nobody re-walks it.**
> Truncation is the tempting explanation because the failures are *slower* than
> the successes and CLAUDE.md warns about exactly this class. But the successful
> responses measure only ~373 output tokens — nowhere near 2048 — so a 5×
> overrun would be needed, and the ~25 s failure time is too fast for 2048
> tokens at the observed generation rate. **NOT asserted, and not to be treated
> as the cause without evidence.** The point of this bug is that the evidence
> does not currently exist.

## 3. Blast radius

`vonc.com/tools/gauntlet/index.html` — the Position and Defence steps. The
front-end handles it correctly (honest "the AI opponent is offline" message, no
clock started, no objective marked, the visitor's text preserved), so this
degrades rather than breaks. But it is the flagship journey, and a visitor who
gets it on their first attempt may not try again.

It also produced the only failing check in the P4 live acceptance run —
`no_console_errors`, because the browser logs each 503. That check is behaving
correctly; the upstream is what is wrong.

## 4. Fix candidates, in order

1. **Log the error before returning the 503**, in both handlers, including the
   upstream status and whether it was a `TruncatedError` (`aiservice.IsTruncated`
   already exists for exactly this). One line each. **Nothing else on this list
   can be evaluated until this lands** — everything below is currently a guess.
2. **Give the HTTP client an explicit timeout.** `&http.Client{}` with no
   timeout means a hung upstream is bounded only by whatever the platform
   happens to impose.
3. **Retry once on a transient upstream failure** (429/529/timeout), with a
   short backoff. The evidence that a retry usually works is strong: every
   observed failure succeeded on a subsequent attempt.
4. **Raise `max_tokens` for the judge** only if (1) shows truncation actually
   occurring — see the note above; do not do this speculatively.

Any of these is a change to `internal/`, so it goes through the council gate per
CLAUDE.md, and the island must be rebuilt and `compose up -d`'d for it to take
effect (`RUNBOOK_gauntlet_dead_cta.md` §5) — a fix committed in-repo is inert.

## 5. How to verify a fix

After (1), reproduce by sampling until a burst appears, then read the island log
for the recorded reason:

```
ssh root@toolsapisuk.vs.mythic-beasts.com 'cd /opt/island && docker compose logs --since 30m tools-api | grep -i judge'
```

A fix for (3) is verified when a sampling loop of ~30 rounds returns no 503 to
the caller **while the underlying transient is still occurring** — i.e. the log
from (1) must still show retried failures. A quiet period proves nothing here,
because the fault is bursty: 23 consecutive clean calls were recorded today
within minutes of two live failures.
