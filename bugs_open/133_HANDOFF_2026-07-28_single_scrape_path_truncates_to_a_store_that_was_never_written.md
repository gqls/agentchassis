# 133 — the single-scrape path truncates content to an S3 copy it never wrote, and still drops an oversized reply in silence

Filed 2026-07-28 by session "bugsearch 3", while giving the `bugs_closed/062`
payload watch a non-zero denominator (that watch had been reading **0 errors over
0 scrape attempts** since the 07-28 roll, so it was unfalsifiable rather than
clean — see `bugfix_100_101_scrape_provenance`).

**Found by firing one scrape and reading what the adapter actually did**, not by
reading code first. The code reading came after, to explain the log line.

---

## Two defects, same function, both in the SINGLE-URL scrape path

`internal/adapters/webscrape/adapter.go`. The **batch** path
(`batch_handler.go`) is correct on both counts — it was fixed by `bugs_closed/062`.
Its single-URL sibling was explicitly left "untouched" by that fix, and the
reason given was narrow (it "never carried the bool-trap fields"). That is true
and it is not the same as safe.

### A. The truncation marker names a store that was never written — NEW

`adapter.go:331-344` truncates three content fields at 50,000 chars and appends:

```
[Content truncated for Kafka transport - full version in S3]
```

justified by the comment directly above it:

```go
// Truncate large content fields before sending through Kafka.
// Full content is already in S3 (via storage URIs).
```

**The premise is conditional; the truncation is not.** The S3 upload is guarded
40 lines earlier at `adapter.go:313`:

```go
if uploadResults && a.storageClient != nil {
```

`uploadResults` comes from the step's own `upload_results` config. When it is
false or unset, **nothing is uploaded and the content is destroyed**, while the
appended marker tells the reader — human or model — that a full copy exists in
S3. Nothing is written, so nothing can be recovered.

**MEASURED, not inferred.** One scrape of `https://vetcomparison.uk` with
`upload_results: false`, correlation `1e97bd22-6823-486b-a5e8-86679b3e32e0`,
2026-07-28 19:35:39Z, adapter pod `web-scrape-adapter-c576d96b-qr72m`
(`v1.0.1192`):

```json
{"msg":"Truncating large field for Kafka","field":"raw_html",
 "original_len":53805,"truncated_to":50000}
```

and in the same correlation, **zero** `result of upload to S3` lines. 3,805
characters of the page were discarded and replaced with a pointer to a file that
does not exist. The response was then produced successfully — so this failure is
invisible from the 062 watch, which only greps for produce errors.

**Live exposure — 4 of the 6 single-URL scrape steps in the fleet:**

| agent | step | action | `upload_results` |
|---|---|---|---|
| `site-scraper` | `scrape_site` | `firecrawl_scrape` | **false** |
| `domain-research-classifier` | `scrape_site` | `scrape_web` | **(unset)** |
| `site-adoption-agent` | `fetch_primary_css` | `firecrawl_scrape` | **(unset)** |
| `vet-practice-verifier` | `scrape_website` | `scrape_web` | **(unset)** |
| `website-capture-firecrawl` | `scrape_main_page` | `firecrawl_scrape` | true — safe |
| `website-extract-structured` | `scrape_page` | `firecrawl_scrape` | true — safe |

Query in the RUNBOOK section below; re-run it rather than trusting this table.

> **`[INFERRED]`, and it is the reason this is filed rather than noted:** the
> field that gets cut is the **tail** of the document, and a UK company
> registration number is conventionally in the page **footer**. `vet-practice-
> verifier/scrape_website` exists to extract exactly that, and is the pipeline
> `bugs_open/100` and vetcomparison's P1 are waiting on. **This is a mechanism,
> not a measurement** — no vet practice page has been scraped since 2026-03-18,
> so it has never been observed. It is cheap to settle and the P1 pilot will
> settle it: if company-number extraction has a poor hit rate on pages over 50KB
> and a good one under, this is why.

### B. An oversized reply is still dropped in silence — the gap 062 left

`sendSuccessResponse` (`adapter.go:437-447`) ends:

```go
); err != nil {
    a.logger.Error("Failed to produce response", zap.Error(err), ...)
}
```

Log, and return. No degrade, no resend, **no error response** — although
`sendErrorResponse` is defined two lines below at `adapter.go:450` and is already
called from four other places in the same function's file.

Compare the batch sibling, `batch_handler.go:368-378`, whose comment states the
rule this one breaks:

> *"a reply the broker refuses as too large is degraded … and resent ONCE; if
> even the stub reply is refused, an error response goes out instead. **A
> response that cannot be delivered must become a deliverable error, never
> silence** — the caller is listening on the reply topic, not reading this pod's
> logs, and the silent drop starved callers through 4 × 180s of retries."*

That is `016b §9`'s own entry, derived from 062. It holds on one path in the
package and not the other.

**Why A makes B rarer but not gone.** The 50KB truncation covers
`markdown_content` / `html_content` / `raw_html` and strips `screenshot_base64`.
It does **not** bound `links`, `metadata`, or any other field, and the `pages`
array is truncated per-field with no cap on the number of pages. The effective
limit is the broker default — `system.agent.generic.responses` carries **no
topic-level override** (`kafka-configs.sh --describe` returns an empty dynamic
config) and the cluster CR sets no `message.max.bytes`, so ~1MB applies.

> Noted in passing, not a defect claim: `platform/kafka/topic_manager.go:151,227`
> sets `max.message.bytes` to 5,242,880 on topics it creates, which the
> auto-created reply topics are not. Whoever fixes B should decide which number
> is intended rather than inheriting both.

---

## Why this was not found earlier

**The documented 062 watch reads one pod of three.**

```bash
kubectl -n ai-persona-system logs deploy/web-scrape-adapter --since=3h | grep -i "Message Size Too Large"
```

`deploy/web-scrape-adapter` has **3 replicas**, all in consumer group
`webscrape-adapter-group`, on a topic with **1 partition** — so exactly one pod
consumes and the other two are idle for their whole life. `kubectl logs
deploy/...` picks one pod arbitrarily; on 2026-07-28 it picked `d8h2w`, an idle
one. **The watch can report a clean log while the only pod doing work is
erroring.** Use `-l app=web-scrape-adapter --tail=-1`, which reads all three.

This is the same shape as the zero-denominator problem it sat next to: the
command answered a narrower question than the one it appeared to answer, and
returned a reassuring number either way.

---

## How to verify / re-measure (RUNBOOK)

**Fire one scrape.** The two things the obvious version gets wrong:

1. **The adapter takes a `{"body":…,"headers":…}` envelope as the Kafka VALUE**
   and does not read Kafka message headers at all. A bare body is rejected at
   `adapter.go:199` (*"Invalid message format - missing headers or body"*) and
   **committed**, so it vanishes with no retry and no redelivery.
2. **The reply topic must already exist.** If it does not, the produce fails and
   logs `Failed to produce response` — one of the two strings the 062 watch
   greps. You would manufacture the exact hit you are testing for. Create and
   confirm the probe topic with `kcat -L` *before* firing.

Working script (kept out of the repo deliberately — it publishes to prod):
see NOTES in `bugfix_100_101_scrape_provenance` §14 for the full text.

```bash
# denominator FIRST — an error count over zero attempts is not evidence
kubectl -n ai-persona-system logs -l app=web-scrape-adapter --tail=-1 --since=1h | grep -c "Starting scrape"
kubectl -n ai-persona-system logs -l app=web-scrape-adapter --tail=-1 --since=1h | grep -ci "Message Size Too Large\|Failed to produce"
# and the defect this bug is about, which produces NO error at all:
kubectl -n ai-persona-system logs -l app=web-scrape-adapter --tail=-1 --since=1h | grep "Truncating large field"
```

**Live exposure table:**

```sql
SELECT ad.type AS agent, e.k AS step, v->>'action' AS action,
       COALESCE(v->'config'->>'upload_results','(unset)') AS upload_results
FROM agent_definitions ad,
     jsonb_each(ad.default_config->'workflow'->'steps') AS e(k,v)
WHERE ad.deleted_at IS NULL AND COALESCE(ad.is_snapshot,false)=false AND ad.is_active
  AND v->>'action' IN ('scrape_web','firecrawl_scrape','batch_webscrape')
ORDER BY upload_results, ad.type, e.k;
```

---

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Make the marker tell the truth, and make truncation conditional on the copy
   existing.** If `uploadResults` is false, either (a) do not truncate and let B's
   fix handle an oversized reply, or (b) truncate with a marker that says the
   remainder was **discarded**, not stored. **(b) is a two-line change and removes
   the false claim immediately**; (a) is the better behaviour but needs B first.
   Best of all: pass the storage URI into the marker so it is *impossible* to
   claim a copy without naming one.
2. **Apply 062's fix to `sendSuccessResponse`** — degrade, resend once, then
   `sendErrorResponse`. The function it needs is already in the file. This closes
   B and makes 1(a) safe.
3. **Default `upload_results` to true for the single-URL path**, so the comment's
   premise is actually met. Behaviour + cost change across 4 other owners' agents
   — an owner call, not a bug fix. Listed for completeness, not recommended.
4. **Fix the watch command in `bugs_closed/062` and everywhere it is quoted**
   (`-l app=…` not `deploy/…`). Costs nothing and stops the next investigation
   starting from a false clean.

**Do not do 3 as a side effect of 1 or 2.**

## Ownership

`internal/adapters/webscrape/` — `bugs_closed/062` was the
`model_directory_pipeline` workstream. This file was raised by the
`bugfix_100_101_scrape_provenance` lane, which does **not** own the adapter and
has not taken the fix. `scripts/who-owns.py 062` before routing work at it.

Affected agents span at least four other lanes (`site-scraper`,
`site-adoption-agent`, `domain-research-classifier` — which has **no owner at
all** — and `vet-practice-verifier`).

## Related

- `bugs_closed/062` — the batch half, fixed. Its "single-scrape path is
  untouched" note is where this was hiding in plain sight.
- `bugs_closed/101` — same class: config/text describing behaviour that does not
  happen. The marker here is the message-level version of it.
- `bugs_open/100` — depends on `vet-practice-verifier/scrape_website`, which is
  one of the four exposed steps.
- `016b §9` — *"A response that cannot be delivered must become a deliverable
  error"* (from 062), and the new entry this case adds.
