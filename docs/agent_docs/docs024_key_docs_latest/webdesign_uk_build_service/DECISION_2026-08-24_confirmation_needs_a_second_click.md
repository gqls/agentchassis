# DECISION 2026-08-24 — confirmation requires a second click (OWNER RULING)

**Owner, 2026-08-24, in-session:** *"We can't have email scanners clicking the accept
button so we'll need a separate page."*

This CLOSES the question that `DECISION_2026-08-21b` §4 and the
`box/links.webdesign.uk.nginx` header both carried as "the owner's open call": the
mail-scanner residual is **not** accepted risk. The remedy is the second click.

## What the ruling means mechanically

Today `GET /c/<token>` **mutates** — the click IS the confirmation (owner ruling
2026-08-19, "recording the click is the state"). The prefetch guard (`0e9cb31ee`, live in
`v1.0.1332`, council-approved `6b1726ab`) refuses HEAD and *announced* speculative
fetchers — but a corporate mail scanner that issues a plain GET is indistinguishable from
a person, and it would confirm a transfer nobody confirmed. That is exactly the failure
the guard's own rationale names, on the vector it could not close.

So the endpoint splits:

- **`GET /c/<token>` becomes non-mutating.** It renders a small confirmation page —
  who/what is being confirmed, one button. Valid token → page; used/expired token → the
  existing "no longer active" page (that flow already exists in the handler).
- **The mutation moves behind the button** — a `POST` (form submit) to the confirm
  route. Scanners follow GETs; they do not submit forms. A human pressing the button is
  the second click.
- The 2026-08-19 "the click is the state" ruling is **amended, not repealed**: the
  *button* click is the state; the *link* click is now just navigation.

## What does NOT change

- The token semantics: unguessable, hashed at rest, expiring, scoped to one site and one
  purpose. The secret stays in the link.
- The prefetch guard stays as the outer layer (412 + no-store on HEAD/announced
  prefetch) — defence in depth, not superseded.
- `box/links.webdesign.uk.nginx`: no structural change — `location /c/` proxies the
  prefix, so GET and POST both pass. (Header comment updated to record this ruling.)
- The delivery-email builder still mints the same-shaped links on `links.webdesign.uk`.

## Sequencing — this BLOCKS the first delivery email

The residual "becomes live at the first delivery email". The owner has now ruled it must
not exist, so: **no delivery email goes out before the second-click page is live in
core-manager.** `customer_access_tokens` = 0 rows (2026-08-24), so nothing is at risk
while this is pending.

## The build (owed, not started)

- `internal/core-manager/handlers/delivery.go` — `HandleConfirmTransfer` splits: GET
  renders the page (NO state change on any arm), new POST handler performs the confirm.
  Page is handler-served transactional HTML, same class as the existing "no longer
  active" page (this is not a framework site; the every-site-through-the-framework
  ruling governs sites, not a service's own transactional page).
- `internal/core-manager/api/server.go` — add the POST route beside the GET.
- Tests: GET with a valid token mutates NOTHING (assert at the DB, not the status);
  POST confirms once; second POST → "no longer active"; guard still 412s HEAD.
- Council round (internal/ — normal gate), then it rides a core-manager roll.

## Falsifiers

- `customer_access_tokens` count — 0 as of 2026-08-24; any non-zero row before the page
  ships raises the stakes of this being unbuilt.
- Whether `HandleConfirmTransfer` has grown other callers/semantics by build time — read
  the handler fresh, don't trust this sketch.
