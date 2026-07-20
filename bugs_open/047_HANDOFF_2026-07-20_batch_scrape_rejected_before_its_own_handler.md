# 047 — every batch_scrape is rejected as "Empty URL" before reaching its own handler

**Found:** 2026-07-20 by the claims-verification thread, on the evidence-researcher's
first smoke run. **Fix committed same day** (guard reordered in
`internal/adapters/webscrape/adapter.go`) — **OPEN until a web-scrape-adapter image
is rebuilt and rolled**; the defect is reproducible in prod until then
(bugs_closed bar: fixed AND live).

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

## Related

- The invisible-from-the-caller shape (error response sent but never claimed;
  await expires) is the `bugs_open/003` / 035-envelope family — this bug hid
  behind it.
- §9 pattern added to 016b: a shared guard placed before an action switch
  breaks exactly the action that legitimately lacks the guarded field, and it
  fails at 100% while looking like a transient timeout to every caller.
