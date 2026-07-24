# 062 — batch_scrape response exceeds Kafka max message size; adapter drops it silently; caller starves through timeout retries

Filed 2026-07-24 by the model_directory_pipeline workstream (session
"ai-agent-orchestration"), which hit it on the first real exerciser of the
batch-scrape path since bug 047 unblocked it.

## Symptom

Any orchestration whose workflow includes `batch_webscrape` on content-heavy
pages fails at that step with
`Request <id> timed out after 3 retries` after ~12 minutes (4 × the 180s
default await), having made no progress. The scrape itself is NOT slow and
NOT failing — the caller just never hears the answer.

Observed on three consecutive `directory-researcher` runs, 2026-07-22 →
2026-07-24 (orchestrations `78da45b7`, `fe011a93`, `5073941d`). The first
was misdiagnosed twice before the logs were caught in time (see the
corrections trail in
`docs/agent_docs/docs024_key_docs_latest/model_directory_pipeline/NOTES_model_directory_pipeline.md`).

## Root cause — three defects compounding, all evidenced

**Evidence (adapter pod `web-scrape-adapter-7c495b684d-t5gvw`, 2026-07-24
10:01:59–10:02:03 UTC, orchestration `5073941d-08ca-49c5-b922-7dac8a90feb2`):**
the batch of 3 URLs scraped successfully in **4.69 s**
(`"msg":"Batch scrape completed","total":3,"success":3,"errors":0,"duration":4.694009773`),
then:
```
"msg":"Failed to produce Kafka message","send_to_topic":"system.agent.generic.responses",
"error":"[10] Message Size Too Large: the server has a configurable maximum message size ..."
"msg":"Failed to produce batch response" (batch_handler.go:327)
```
Nothing else happens. The orchestration's awaited request then times out
through 3 retry resends — each retry re-scrapes successfully and re-fails
identically — and the workflow FAILs ~12 minutes after it could have failed
in 5 seconds.

1. **The firecrawl provider fetches every page in four formats,
   unconditionally.** `internal/adapters/webscrape/providers/firecrawl.go:62`
   hard-codes `formats := ["markdown", "html", "rawHtml", "links"]` for the
   `/scrape` endpoint, with NO config override — unlike the `/crawl` path
   (line 314–327), which honours `config["formats"]`. Every scraped page
   comes back ~3× its useful size before the handler touches it.

2. **The batch handler forwards all of it, plus a duplicate.**
   `internal/adapters/webscrape/batch_handler.go:193-211` passes through
   `markdown`, `raw_html`, `html_content`, AND a backward-compatible
   `content` field that is a byte-for-byte copy of `markdown`. A modern
   pricing/docs page renders at 100–500 KB per format; 3 pages × ~4 copies
   trivially exceeds the broker's ~1 MB default `max.message.bytes`.

3. **On the produce failure, the adapter logs and gives up** —
   `batch_handler.go:320-330`. No truncation retry, no degraded response, no
   error response. From the caller's side this is indistinguishable from the
   adapter being down, so the coordinator burns its full retry budget on a
   failure that is deterministic. **A response that cannot be delivered
   must become a deliverable error, never silence.**

## Why nobody hit this before

`batch_scrape` was dead on arrival until 2026-07-21 (bug 047: "Empty URL"
rejection before the handler — CLOSED & LIVE v1.0.1145). Since then, the
directory-researcher runs above are the ONLY orchestrations to touch
`batch_webscrape` (checked: 3-day sweep of `orchestration_states.workflow_plan`
found no other user). The path has never once round-tripped a real
content-heavy batch in production. `evidence-researcher` (same step shape)
is seeded but has not run live either — when it does, it hits this
immediately.

Related but distinct: 005/008/009/012's truncation-and-config family is
about LLM output being CUT (max_tokens); this is transport-layer loss of a
COMPLETE result. The failure smell differs: work provably done (adapter
logs success), caller times out anyway.

## Also noteworthy (found during diagnosis, not part of this defect)

Step-level `"timeout_seconds"` in a workflow seed is **silently dropped** —
`models.Step` has no such field; only `config.timeout_seconds` (inside the
step's `config` object) is read (`datahelpers.ConvertStepTimeout`,
`timeout_helpers.go:23`). The `evidence-researcher` seed
(claims_verification/SEED_evidence_researcher.sql) and the
`directory-researcher` seed copied from it both carry step-level values
(120s) that have never had any effect; every await in both workflows runs at
the 180s `DefaultRequestTimeout`. Not the cause of this bug (the response is
lost, not late), but a config-that-lies trap worth its own sweep — how many
other seeds set step-level timeout_seconds that nothing reads?

## Fix candidates

**A (the structural fix, adapter-side, needs image roll):**
1. `batch_handler.go`: stop duplicating — emit `content` (markdown preferred)
   and `metadata` only; include `raw_html`/`html_content` ONLY when the
   request's scrape_config explicitly asks (`include_raw_html: true`). No
   current caller asks.
2. Cap per-result `content` at a configurable max (suggest 150 KB) with an
   explicit `"truncated": true` marker on the result — a visible cut, per
   the 012-family lesson that an invisible cut is the damage.
3. On `Message Size Too Large` from the producer: strip to
   metadata+truncated-content and retry ONCE; if still too large, send
   `sendBatchErrorResponse` (recoverable) so the caller fails in seconds
   with a real error instead of starving for 12 minutes.

**B (complementary, provider-side):** honour `config["formats"]` in
`FirecrawlScrapingProvider.Scrape` the same way the crawl path does, so
text-only callers (evidence-researcher, directory-researcher) can request
markdown-only and cut the fetch itself by ~3×.

**C (out of scope here, noted for completeness):** raising the broker's
`max.message.bytes` treats the symptom fleet-wide and invites the next
bigger page to re-break it; rejected as primary fix.

## How to verify

Failing branch first (the bug IS the silent branch): batch-scrape a set of
pages whose combined 4-format payload exceeds 1 MB (llm-stats.com/models
alone did it on run 2) and confirm the caller receives EITHER a truncated
success (with `truncated: true` markers) or a fast error response —
anything but a timeout. Then the happy path: a small batch round-trips with
content intact. Then re-fire `model-directory-discovery`
(`UPDATE scheduled_tasks SET last_triggered_at = NULL WHERE name='model-directory-discovery'`)
and confirm `directory_claims` gains rows.
