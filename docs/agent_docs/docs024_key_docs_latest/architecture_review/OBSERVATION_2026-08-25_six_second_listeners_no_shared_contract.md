# OBSERVATION 2026-08-25 — six independent "second listener" implementations, no shared contract

**Status: OPEN, unowned, not scheduled — but no longer trigger-less.** Filed at the
architecture seat's request during council `25cd3044` (round 2, `architecture`, medium;
`reuse_agent` raised the same at low). This is a tracked observation, not a proposal to
refactor anything now.

> **TRIGGER (added 2026-08-26, from the same council's round 3):** the architecture seat
> objected that an indefinitely unowned residual is how an observation becomes furniture,
> and asked for "an owner and a trigger threshold (e.g. next occurrence forces an RFC)".
> The owner is the owner's call, not a session's. The threshold is adopted here as stated:
> **the NEXT session that adds another second-listener implementation — an eighth — must
> file an RFC for the shared contract first, citing this file, instead of hand-rolling.**
> The cheap detector is the census command below (`grep -rn "http.Server{" internal/
> platform/ cmd/`): if your change grows its count, this trigger is about you.

## What was observed

Adding the delivery-only listener (register **SYS-095**) made core-manager the **sixth**
place in this estate that stands up a second, constrained HTTP listener beside a service's
main one:

| # | where | what it serves |
|---|---|---|
| 1 | `platform/health/server.go` | `/health`, `/ready` (+ a separate metrics listener) |
| 2–6 | `internal/adapters/{analyser,git,browserrunner,webscrape,thunder}/adapter.go` | each hand-rolls its own `healthServer` rather than using (1) |
| 7 | `internal/core-manager/api/server.go` | the delivery routes (this change) |

`[MEASURED 2026-08-25]` by `grep -rn "http.Server{" internal/ platform/ cmd/`.

So a shared helper **exists** and five services already do not use it, and the seventh
implementation has now been added with reasons that are individually sound.

## Why the delivery listener did not use `platform/health`

Recorded so that the next person does not have to re-derive it, and so the reasons can be
challenged rather than assumed:

- it is **gorilla/mux**, while the delivery handler's single route-table definition takes a
  `gin.IRouter`;
- it **always mounts `/health` and `/ready`** — which breaks the delivery listener's whole
  value, that it serves the delivery routes and nothing else. That property is asserted by a
  test, and a mutation adding one extra route to the engine fails it;
- it retains **no `*http.Server`**, so it cannot be gracefully drained — a customer mid-click
  would simply be dropped;
- `Start()` also launches a **metrics listener**, a second port nobody asked for.

## The actual objection, which is not "you should have reused it"

The architecture seat's point is sharper than reuse:

> *"none of the safety properties this plan builds (opt-in guard, route-table assertion,
> wildcard capture check, drain ordering) are captured anywhere reusable, so the next
> service needing this will either re-derive it or copy this one without the tests that
> prove it."*

That is correct and it is the part worth tracking. The delivery listener now carries four
properties that any constrained public listener wants, and all four live in
`internal/core-manager/api/server.go` where nothing else can reach them:

1. **opt-in with the unsafe default OFF** (an unset port serves the routes nowhere);
2. **a route-table assertion** that the protected routes are not on the main router,
   including wildcard capture by prefix dispatch;
3. **a shape assertion** that every route is servable through the edge that fronts it
   (here: the box's anchored regex — a suffix route 404s at the box with no trace in the
   cluster);
4. **drain ordering** — close the public door first, and do not let its error mask the
   main listener's.

Property 3 is arguably the most transferable and the least obvious: it is only meaningful
because something *outside* the binary decides what can reach it, which is true of every
door in SYS-094's two-door pattern.

## What would close this

A `platform/` helper carrying the four properties, adopted by core-manager first and offered
to the adapters, **or** a decision that the adapters' health listeners are a different enough
shape that unifying them is not worth it — with that decision written down.

**Not done now, deliberately:** it would mean editing five live adapters inside a change
scoped to one service's exposure boundary, which is the bundling the guardian seat vetoed in
`bugs_closed/124`. The estate's own rule is that a seam gets its own round.

## Relations

SYS-095 (the sixth listener and the four properties) · SYS-094 (the two-door pattern that
makes property 3 general) · council `25cd3044` (where this was raised) ·
`architecture_review/RFC_054` §6 (the other open residual from the same change).
