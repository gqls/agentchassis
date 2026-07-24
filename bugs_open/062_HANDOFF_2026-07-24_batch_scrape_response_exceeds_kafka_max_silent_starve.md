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

## Defect 4, found by the fixed path (2026-07-24, run 4)

With defects 1–3 fixed and pod-verified live (v1.0.1152), run 4
(`orchestration 5db0890c`, request `1c11033d`) STILL timed out — and the
logs now prove a second, independent kill-switch stacked on the same path:
the reply was small (79,397 bytes, markdown-only honoured), **successfully
produced** to `system.agent.generic.responses` at 10:51:00, **consumed by
the chassis** at 10:51:01.720 with perfect headers — and then dropped at
`processor.go:1469`: `Failed to unmarshal response message (ResponseMessage)`,
followed by `Cannot propagate error to parent - missing replyToRequestID`.

Cause: the batch handler's JSON-body headers carried `"is_complete": "true"`
and `"is_error": "true"` — STRINGS — while the chassis unmarshals the
envelope into `types.ResponseHeaders`, whose `IsComplete`/`IsError` are
`bool`. `json.Unmarshal` hard-fails on the type mismatch and the whole
response is discarded. This is the **known 035 §1.5 bool trap**: the
browserrunner, analyser and thunder adapters all carry explicit corrections
("is_complete/is_error are real JSON bools") — the batch handler, added
2026-07-21 for bug 047, copied the older broken pattern. The Kafka HEADER
map (strings by contract) was never the problem; only the typed JSON body.

Consequence worth stating plainly: **no batch_scrape response had EVER been
consumed successfully** — the size refusal (defects 1–3) masked this parse
failure, which was waiting underneath. Each fix peeled back the next layer.
Fixed: both fields are real bools; the envelope construction is extracted to
a pure `buildBatchSuccessEnvelope` and a test round-trips it through the
real `types.ResponseMessage` (plus a regression guard demonstrating the
string form fails), so the contract cannot silently regress.

## Layer 3, found by the fully-working pipeline (2026-07-24, run 5)

With defects 1–4 fixed and live (adapter v1.0.1153, pod-verified), run 5
**COMPLETED end-to-end for the first time** — search → scrape → reply
delivered AND parsed → LLM extraction → live verification — and the
fail-safe engaged exactly as designed: every candidate was rejected
`citation_lost` and terminated at a `directory_citation_unverified` human-
review item rather than being silently registered or dropped.

But the rejections themselves exposed the next (and, for this case, final)
representational gap: **researchers quote from the scrape's MARKDOWN
rendering; the verifier re-fetches HTML and flattens it to visible text.**
The live evidence: every quote was a pricing-table ROW in firecrawl's
markdown table syntax — verbatim from run 5's collected_data:
`"gpt-5.6-sol | $5.00 | $0.50 | $6.25 | $30.00 | ..."` — while the
verifier's fetch of the same page (checked directly: 200, 17,460 chars of
visible text, prices present) flattens that row to space-joined cells with
NO pipes. The quote can never match, deterministically, for ANY claim
quoted from a table — which is where pricing facts live.

Fix (shared `datahelpers.NormalizeForQuoteMatch`, chassis-side): fold `|`
to a space, exactly as the normaliser already folds nbsp/curly quotes/
thousands separators — the pipe is presentation (markdown table syntax),
not source content, and the file's own rule is "forgiving about
presentation, strict about content". Strictness is preserved and
test-pinned: an altered price in the same row shape still fails, and cells
from different rows cannot be stitched into one matching quote. Test case
uses the verbatim failing quote from run 5. Benefits `evidence-researcher`
identically — its acquisition lane would have hit the same wall on its
first tabular source.

NOTE this fix ships in the **agent-chassis image** (the verify action runs
chassis-side), unlike defects 1–4 (web-scrape-adapter image). Committed to
HEAD; rides the next chassis build/roll.

## Council verdict + objection follow-ups (2026-07-24)

**APPROVED round 1** (corr `fe468218-d2c3-477e-a1ff-3f0f6cd1e57d`, "3
advisory objections, none high-severity"). The three checkable objections
were each closed with evidence rather than argued:

1. **Asserted-absence on the blast radius** (guardian + prior_art_librarian,
   both medium): the claim "no live workflow reads raw_html/html_content
   from a batch_webscrape output" now has its lookup attached. Exhaustive
   `agent_definitions` scan (active, non-snapshot, not deleted):
   `batch_webscrape` is invoked by exactly THREE workflows —
   `evidence-researcher`, `directory-researcher`, `research-agent`, each at
   a `scrape_pages` step feeding an LLM-extraction step. Six OTHER agents
   match `raw_html|html_content` in config text, all self-referential or
   unrelated: `html-developer-chunked` reads its own generation steps'
   `raw_html`; `tool-generator` maps its own `generated_html.result`;
   `tool-recreation-handler` reads `existing_content.raw_html` populated by
   `load_existing_content` (a DB/storage read, not the webscrape adapter);
   `council-gate`/`feature-designer`/`fix-proposer` hits are prompt/config
   text. The single-scrape path (adapter.go `sendSuccessResponse`) is
   untouched by this fix.
2. **Substring-only error classification** (editquality, low): the
   kafka-go client (segmentio/kafka-go v0.4.47) DOES expose typed errors —
   `kafka.MessageSizeTooLarge` (broker code 10) and
   `kafka.MessageTooLargeError` (writer-side pre-send detection) — and the
   producer wraps with `%w`, so `isKafkaMessageTooLarge` now checks
   `errors.Is`/`errors.As` first, keeping the substring as a fallback for
   composite shapes (`kafka.WriteErrors`) the unwrap chain can miss.
   Test covers all three routes.
3. **Deploy verification by commit hash is the documented trap**
   (debug_historian, medium): the post-roll check is a pod-grep of a symbol
   this change CREATED, with a positive control, against the
   web-scrape-adapter pod (not agent-chassis — the adapter is its own
   image/service):
   ```
   kubectl -n ai-persona-system exec <web-scrape-adapter-pod> -- \
     sh -c 'strings /app/web-scrape-adapter | grep -c stripBatchResultsForRetry'
   ```
   (expect >0; positive control: grep `batch_scrape`, which predates the fix).

## How to verify

Failing branch first (the bug IS the silent branch): batch-scrape a set of
pages whose combined 4-format payload exceeds 1 MB (llm-stats.com/models
alone did it on run 2) and confirm the caller receives EITHER a truncated
success (with `truncated: true` markers) or a fast error response —
anything but a timeout. Then the happy path: a small batch round-trips with
content intact. Then re-fire `model-directory-discovery`
(`UPDATE scheduled_tasks SET last_triggered_at = NULL WHERE name='model-directory-discovery'`)
and confirm `directory_claims` gains rows.
