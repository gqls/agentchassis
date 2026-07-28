# 132 — every B2-hosted site serves a raw JSON error blob instead of a 404 page, and leaks the bucket object key

**Filed 2026-07-28** from `webdesign.co.uk`. **Unowned.** Affects **every
B2-hosted site in the fleet**. Measured today on webdesign.co.uk, robot-hands.com
and oufe.com; `relojistas.com` is the control that passes, and passes for a reason
that matters (below).

---

## Symptom

Request any path that does not exist and the visitor gets this, with
`HTTP 404` and `content-type: application/json`:

```json
{
  "error": "B2 returned error",
  "objectKey": "webdesign.co.uk/definitely-not-real-123",
  "status": 404,
  ...
}
```

Not a page. A raw internal error object, in the browser, to a member of the public.

Measured 2026-07-28 14:10:

| site | missing path returns | has a `404.html`? |
|---|---|---|
| webdesign.co.uk | **raw B2 JSON, leaks `objectKey`** | **yes** |
| robot-hands.com | **raw B2 JSON, leaks `objectKey`** | no |
| oufe.com | **raw B2 JSON, leaks `objectKey`** | no |
| relojistas.com | `<title>404 Not Found</title>` — a real page | no |

## Two things this is

1. **A user-facing quality defect.** A visitor who follows a stale link, a bad
   redirect or a mistyped URL gets a JSON error object. On `webdesign.co.uk`
   specifically this is acutely bad: it is a **web design** site, and the page a
   prospect is most likely to hit by accident is the one that looks broken and
   unfinished. It is also the exact failure mode this site had last week
   (`bugs_open/092`/`116`: 10 of 13 home page links were 404) — so the page that
   should absorb that damage gracefully instead advertises it.
2. **A minor information disclosure.** `objectKey` exposes the bucket path
   convention (`<domain>/<path>`), confirming the storage layout and that the
   origin is an object store. Low severity on its own, free to fix alongside 1.

## Why the control passes, and why that is the useful part

`relojistas.com` returns a proper 404 **and has no `404.html` in the repo at all.**
It is not better configured — it is **on a different stack**. That site was moved
to a VM behind nginx (`relojistas_rebuild` / `traffic_probe` workstream), and nginx
serves an error page by default. Every site still on **B2 behind Cloudflare** fails.

So the split is not per-site configuration drift; it is **B2 has no error-document
concept**. S3 *website* endpoints support one; the B2 S3-compatible API does not,
and we front it directly. This is the same family as the standing fleet-wide fact
that **an object store cannot resolve directory indexes** (`/tools/` 404s while
`/tools/index.html` serves) — recorded in `SQL_p10` and `bugs_open/116`. Same root
property, second consequence, and the first one has already cost this fleet a live
incident.

## The trap this exposes — a 404 page that exists and is never served

`webdesign.co.uk/404.html` **exists, is deployed, and returns 200 when fetched
directly.** It is simply never used, because nothing routes a missing path to it.

I walked into this myself and it is worth recording as the shape to distrust: a
live sweep of all 99 deployed pages found 98 carried the new analytics beacon and
`404.html` did not, so I added the beacon to `404.html` (`gqls/sites da02ee09f`,
deployed, verified present at `/404.html`). **That fix is inert.** The beacon only
fires if a visitor navigates to `/404.html` by name, which nobody does. The
verification that mattered was requesting a *genuinely missing path* and reading
the response body — not fetching the file and confirming the tag was in it.

> **"The file is deployed and contains the change" is not "the change is on the
> path users take."** Fetch the artefact by the route a user would reach it by.

## Fix candidates — ordered by what closes the door

1. **A Cloudflare Worker (or Snippet) on the zone that intercepts an origin 404
   and returns `<domain>/404.html` with status 404.** Closes it for every site at
   once, at the layer that already fronts every site, with no per-site work and no
   bucket change. Also the natural place to stop the JSON body reaching the client
   at all. One implementation, fleet-wide.
2. **A Cloudflare Custom Error Response / error-page rule** on each zone. No code,
   but per-zone configuration that will drift as sites are added, and availability
   depends on plan — check before promising it.
3. **Move remaining sites to the VM/nginx stack**, where this is free (as
   relojistas demonstrates). Correct long-term direction and already under way, but
   far too large a change to be *this* bug's fix.
4. **Generate a `404.html` for every site.** On its own this fixes **nothing** —
   webdesign.co.uk already has one and still serves JSON. Only worth doing as the
   companion to 1 or 2, which need a target document to serve.

Candidate 1 is the one to take. Note 4 is listed explicitly because it is the
obvious-looking fix and it is a trap: this bug was found on the one site that
already had the file.

## How to verify a fix

Request a path that certainly does not exist and **read the response body**:

```
curl -s https://<domain>/no-such-page-$RANDOM | head -5
```

- **Before:** a JSON object containing `"error": "B2 returned error"` and `objectKey`.
- **After:** the site's own 404 markup, still with HTTP status `404`
  (a fix that returns 200 is worse than the bug — it makes every broken link
  invisible to crawlers and to any link checker).

Do **not** verify by fetching `/404.html` directly, and do **not** verify by
checking that the file deployed. Both pass today, on a site that is broken.

## Related

- `bugs_open/116` — the three link checks have never run on any site. The two
  compound: nothing detects broken links, **and** the page that catches the
  fallout is not wired up. Analytics on a working 404 page would be the cheapest
  broken-inbound-link detector we could have, which is why I tried to add it.
- `SQL_p10` (webdesign_couk) — the object-store directory-index finding, the same
  root property.
- `bugs_open/125` — page rebuild deploys to a path derived from `name`, ignoring
  `pages.url`. Another way a URL that should exist does not, i.e. another
  generator of 404s that land on this.
