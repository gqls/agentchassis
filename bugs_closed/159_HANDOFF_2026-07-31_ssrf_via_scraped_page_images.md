# 159 — image downloads triggered by scraped page content had no destination guard (SSRF)

Filed 2026-07-31 by the `webdesign_uk_build_service` thread, found while checking
whether an existing package (`platform/httpguard`) already covered the SSRF risk
a new domain-intake feature would carry. It does not — see its own package doc,
which scopes itself explicitly to *"the platform's ONE set of INBOUND-abuse
primitives."*

Grepped `bugs_open/` and `bugs_closed/` for `ssrf` and for any private-IP/metadata
guard before filing: no hits. This is a new finding, not a duplicate.

**Status: fixed and live in this commit.** Filed anyway, per CLAUDE.md's rule that
the bar for closed is *fixed AND live* — this is fixed in the tree but not yet
rolled to a running pod, so it stays open until a `web-scrape-adapter` deploy is
verified against a real pod.

---

## 1. The mechanism

`internal/adapters/webscrape/adapter.go`'s `downloadImage` fetches an image URL
with a bare `*http.Client` and no checks at all:

```go
// BEFORE
req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
resp, err := a.httpClient.Do(req)
```

`imageURL` is not a URL this platform chose. It comes from the `images` array in
a scraped page's own content (`adapter.go:642-679` — a string, or an object with
a `url` field, taken directly from whatever a scraping provider parsed out of the
target page's HTML). **The target page belongs to a domain a customer submitted**
— so the URL is attacker-influenced by construction, the same way any user input
is, and nothing here treats it that way.

**Consequence.** A page containing `<img src="http://169.254.169.254/latest/
meta-data/iam/security-credentials/...">` (the cloud metadata endpoint pattern),
or one pointing at an internal service reachable from the cluster network, causes
`web-scrape-adapter` itself — the pod, with its own network position — to fetch
it. Two further defects compound it: no response-size cap (`buf.ReadFrom(resp.
Body)` reads unbounded), and the default `http.Client` follows up to 10 redirects
with no re-validation of the target, so a URL that *looks* external can 302 to a
private one and the naive per-scheme check a less careful fix might add would
still miss it.

**This is not hypothetical for one site — it is the image-ingestion path every
site build on the fleet already uses.**

## 2. The fix

New package `platform/fetchguard` (register `DBI-025`), a sibling to
`httpguard` covering the mirror direction — what WE are allowed to fetch, not
who is allowed to call us. `NewClient(cfg)` returns an `*http.Client` whose
`Transport.DialContext` resolves the target itself and refuses to dial any
candidate address that is not publicly routable (private, loopback, link-local
— which is where cloud metadata endpoints live — multicast, unspecified, and
`0.0.0.0/8`), checking the **specific address about to be dialed**, not a
pre-resolved hostname, which closes the DNS-rebinding TOCTOU gap a "check then
connect" design would leave open. Because every redirect hop dials through the
same transport, a redirect to a private target is refused by the identical
check, with no separate redirect-target inspection to fall out of sync.
`LimitedRead` caps the response and reports truncation explicitly rather than
silently handing back a partial-but-complete-looking image.

`downloadImage` now uses a dedicated `imageHTTPClient` (fetchguard-wrapped),
kept separate from `httpClient` (used only for the scraping provider's own
fixed, trusted API host) so the guard applies exactly to the fetch that touches
attacker-influenced input and nowhere else.

**Verification:** `platform/fetchguard`'s own test suite proves the properties
directly — a real request to a loopback `httptest.Server` is refused
(`TestGuardedClient_RefusesPrivateTarget`), a request that redirects from a
throwaway server to a loopback one is refused at the redirect hop
(`TestGuardedClient_RedirectToPrivateTargetIsRefused`), a literal metadata-shaped
IP in the URL is refused (`TestGuardedClient_RefusesLiteralPrivateIP`), and the
IPv4-in-IPv6-mapped form of the metadata address is refused identically to the
bare form (`TestIsPubliclyRoutable_IPv4Mapped`) — plus the existing
`internal/adapters/webscrape` test suite still passes unchanged.

## 3. What is deliberately NOT fixed here

`internal/adapters/browserrunner/run_checks_action.go` drives Playwright
(`page.Goto(url, ...)`) — a headless **browser** navigating to a URL, not a Go
`net/http` request. `fetchguard`'s transport-level guard cannot see a browser's
own DNS resolution and connections at all. That needs network-layer
interception inside the browser (Playwright's `page.Route`, or an egress
firewall in front of the browser-runner pod) and is a different piece of work.
**Flagged explicitly so it does not silently read as covered** — anyone
reaching for "does something already guard our outbound fetches" should find
this note, not conclude fetchguard already applies there.

## 4. How to verify once rolled

```bash
kubectl -n ai-persona-system exec <web-scrape-adapter pod> -- \
  strings /app/web-scrape-adapter | grep -c "fetchguard: destination resolves"
```

A positive count confirms the guarded client's error string is compiled into the
running binary — the standard "grep a string your change ADDED" pod-verify, not
a same-tag-roll assumption.

## Sources / relations

- `webdesign_uk_build_service/PLAN_2026-07-28_webdesign_uk_build_service.md` §8 —
  where the gap was first named as a P0 risk for the new domain-intake product.
- `platform/httpguard` — the sibling inbound-abuse package; its own doc comment
  is what disproved the "already covered" assumption.
- Register `DBI-025` — the new `fetchguard` mechanism.
- `LANDMINES.md` — *"platform/httpguard is INBOUND-abuse only"*, the entry
  recording the trap this bug walked into on the way to being found.

---

## CLOSED 2026-07-31 — fixed AND live, verified on all three pods

`web-scrape-adapter` rolled to **v1.0.1215** (pods 15m old at time of check),
carrying the fix. Verified the way §4 of this file specified — **grep for a
string the change ADDED, with a positive control in the same exec**, because a
roll is not by itself evidence a fix shipped (the image can predate the commit):

```
fetchguard marker (ADDED by the fix): 1     <- on all 3 pods
positive control (pre-existing):      1     <- proves the grep works
```

Both markers present on **all three replicas**, not just one — `logs deploy/X`
and single-pod checks read one pod of N, and a partial roll is exactly the state
that looks green from one pod.

Meets this repo's bar for closed: the defect is no longer reproducible in
production, not merely fixed in the tree.

**Still open, and deliberately not closed by this:** `browser-runner-adapter`'s
Playwright `page.Goto` navigation is a different fetch surface — a Go
`http.Transport` guard cannot see a browser's own DNS and connections. Named in
§3 above and in register `DBI-025`'s verify-later so it is not mistaken for
covered by this closure.
