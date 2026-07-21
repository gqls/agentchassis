# 047 — every batch_scrape is rejected as "Empty URL" before reaching its own handler

> **CLOSED 2026-07-21 — FIXED AND LIVE on v1.0.1145.** The reordered guard
> (commit `8d9d9051a`) built into `web-scrape-adapter:v1.0.1145` and rolled to all
> three adapter pods (`84bc6fd69d-*`, 0 restarts). **Verified behaviourally**, not
> by pod-grep (the fix is a pure reordering — no new symbol): a `batch_scrape`
> smoke request (`client_id=verify047`, corr `4619736a-…`) at **2026-07-21
> 10:49:20Z** on pod `…-z2gnc` was logged `Processing webscrape request`
> (action=batch_scrape, **url=""**) → **`Processing batch scrape request`**
> (`batch_handler.go:85`, i.e. it reached `handleBatchScrape`) → `Scraping URL` →
> `Batch scrape completed` total=1 success=1 errors=0 → success response.
> **No "Empty URL in request".** The exact condition that used to reject at the
> door (empty top-level url) now flows to the batch handler. See "Verified live"
> at the foot of this file. Moved to `/bugs_closed/`.

**Found:** 2026-07-20 by the claims-verification thread, on the evidence-researcher's
first smoke run. **Fix committed same day** (guard reordered in
`internal/adapters/webscrape/adapter.go`) — was **OPEN until a web-scrape-adapter
image was rebuilt and rolled**; that happened with **v1.0.1145** (2026-07-21).

## The defect

`internal/adapters/webscrape/adapter.go` validated the **top-level** `url` field
before switching on the request's `action`:

```go
url, _ := data["url"].(string)     // batch requests carry data["urls"] instead
...
if url == "" { sendErrorResponse("URL cannot be empty"); return }
...
switch action { ... case "batch_scrape": handleBatchScrape(...) ... }
```

A `batch_scrape` request carries its URLs in `data["urls"]` (an array) and by
construction has **no** top-level `url` — verified against the sender,
`batch_webscrape_action.go`, which builds `body.data = {urls, config}`. So every
batch request ever sent was rejected at the door with "Empty URL in request" and
never reached `handleBatchScrape`.

## Blast radius

Everything that uses the `batch_webscrape` action against this adapter:

- **`research-agent`'s `scrape_pages` step** — i.e. the page-content-writer's
  whole research lane. Research could search but never scrape.
- **`evidence-researcher`** (V5 acquisition) — how this was found.

The failure is expensive to see from the caller's side: the adapter's error
response is not claimed as the awaited response, so the calling step's
`awaited_requests` row retries (observed `retry=3`) and then **expires**; the
workflow fails at `scrape_pages` with **no `__step_error`** recorded. From the
orchestration you see only an expired await; the one-line truth is in the
adapter pod's log.

## Evidence (smoke run, correlation f930dc2f-9080-4996-94f8-22390350bcf4)

- evidence-researcher: `FAILED @ scrape_pages`, `__step_error` null;
  `prepared_urls.urls_to_scrape` held 4 valid primary-source URLs (IGU, EIA,
  IEEFA ×2).
- `awaited_requests`: `scrape_pages | expired | retry=3` (the retry ladder ran;
  every retry met the same rejection).
- adapter pod `web-scrape-adapter-...-br2z7`: `Processing webscrape request` →
  `Empty URL in request` → `Sending error response`, once per attempt.

## The fix (committed)

Branch `batch_scrape` out **before** the single-url validation; the guard now
covers only the actions that genuinely require a top-level url
(scrape/crawl/map/extract). Three lines moved, one comment. No behaviour change
for single-url requests.

## How to verify after the adapter image rolls

1. Pod-grep is useless here (no new symbol) — verify by behaviour: re-run any
   `batch_webscrape` step, e.g. the evidence-researcher smoke run in
   `claims_verification/RUNBOOK` §7 shape with input
   `{site_id, domain, research_query}`.
2. Adapter log should show the batch reaching `handleBatchScrape` (and NO
   "Empty URL in request" for it).
3. The step's awaited request completes instead of expiring at `retry=3`.

## Verified live (2026-07-21, closing the case)

Deployment state at close:

- All three `web-scrape-adapter-84bc6fd69d-*` pods Running `docker.io/aqls/web-scrape-adapter:v1.0.1145`, 0 restarts.
- Fix commit `8d9d9051a` is an ancestor of the HEAD that v1.0.1145 was built from
  (`make build-*` archives committed HEAD), so the reordering is in the binary.
- Kustomization catch-up is lagging behind the live deploy at close time (the
  committed overlay still read v1.0.1143, a concurrent thread's uncommitted diff
  bumps it to v1.0.1144) — the running image is v1.0.1145, deployed directly. The
  verification is against the **running pod**, not the overlay, exactly because
  the overlay is not the source of truth for what is live.

Behavioural proof (pod `…-z2gnc` log, 2026-07-21T10:49:20Z, `client_id=verify047`):

```
Processing webscrape request      action=batch_scrape url="" corr=4619736a-…
Processing batch scrape request   batch_handler.go:85  url_count=1     ← reached handleBatchScrape
Scraping URL                      index=0 url=https://example.com
Batch scrape completed            total=1 success=1 errors=0
Sending batch scrape success response  result_count=1
```

The rejection string `Empty URL in request` does **not** appear for this request.
An empty top-level `url` — the precise state that used to be rejected before the
action switch — now branches straight into the batch handler and completes. This
is the verification the handoff's "How to verify" section asked for (batch reaches
`handleBatchScrape`; no "Empty URL"; await completes rather than expiring).

## Related

- The invisible-from-the-caller shape (error response sent but never claimed;
  await expires) is the `bugs_open/003` / 035-envelope family — this bug hid
  behind it.
- §9 pattern added to 016b: a shared guard placed before an action switch
  breaks exactly the action that legitimately lacks the guarded field, and it
  fails at 100% while looking like a transient timeout to every caller.
