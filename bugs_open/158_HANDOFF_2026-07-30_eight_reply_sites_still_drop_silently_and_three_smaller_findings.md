# 158 — eight reply-producing sites still drop an undeliverable reply in silence, plus three smaller findings from 133's fix

Filed 2026-07-30 by session "bugsearch 8", while closing `bugs_closed/133`. Each
item here was **deliberately excluded** from that fix rather than overlooked —
133's own guidance is not to ride a behaviour change in on a bug patch, and the
council's architecture seat ruled explicitly on item 1's venue.

Grepped `bugs_open/` and `bugs_closed/` for the mechanism before filing: the only
prior member is `bugs_closed/062` (the batch scrape half, fixed), and there is no
existing bug about the other adapters' reply drops.

---

## 1. THE MAIN ITEM — the rule holds at 2 of 9 sites, and widening it is an RFC

`016b §9`'s rule, from `bugs_closed/062`:

> *A response that cannot be delivered must become a deliverable error, never
> silence — the caller is listening on the reply topic, not reading this pod's
> logs.*

**MEASURED 2026-07-30** (both halves; run both, because the first alone reads as
"one place handles this, copy it"):

```bash
grep -rn "MessageSizeTooLarge\|Message Size Too Large" --include=*.go . | grep -v _test
#  -> was 1 file (webscrape/batch_handler.go). NOW: platform/kafka/reply_delivery.go
#     + its two webscrape callers.

grep -rn "Failed to produce" --include=*.go internal/ platform/
```

| site | file | state |
|---|---|---|
| webscrape single | `internal/adapters/webscrape/adapter.go` | **FIXED** (133) |
| webscrape batch | `internal/adapters/webscrape/batch_handler.go` | **FIXED** (062, repointed by 133) |
| reasoning | `internal/agents/reasoning/agent.go` | log-and-return |
| contentcreator ×2 | `internal/agents/contentcreator/agent.go` | log-and-return |
| websearch ×2 | `internal/adapters/websearch/adapter.go` | log-and-return |
| thunder ×2 | `internal/adapters/thunder/adapter.go:398,596` | log-and-return |

So the rule now holds at **2 of 9** and the mechanism to adopt is already built,
tested and live: `platform/kafka.DeliverReply` (**ADP-017**). Adoption is ~10
lines per site — a degrade callback and answering `FailedUndeliverable` with that
service's own error response.

**⚠ THIS IS AN RFC, NOT A BUG PATCH — ruled by the council's architecture seat on
`7478233b`:**

> *"when adoption widens past webscrape, that IS the RFC moment (it changes 4
> services' caller-observable failure behaviour) — the author should not treat
> ADP-017 registration here as pre-clearing that later expansion."*

The behaviour change is real and it is other people's: today a caller of
reasoning / contentcreator / websearch / thunder that hits this **times out**;
afterwards it gets an **error response**. That is strictly better and it is still
four other services' observable failure semantics. The guardian seat's related
concern in the same round is the standing reason to be careful: *"'inert today'
does not mean inert once other 8 sites see a ready-made shared function sitting in
platform/kafka"* — i.e. `platform/kafka`'s surface area growing one opt-in caller
at a time with no further gate per caller is exactly the drift to avoid.

**How to size it before writing the RFC** (do this first — it may be small):

```sql
-- do any of these services' replies actually approach the limit?
-- ~1MB applies: system.agent.generic.responses has NO topic-level override
-- (kafka-configs.sh --describe returns empty) and the cluster CR sets no
-- message.max.bytes, while platform/kafka/topic_manager.go:151,227 sets 5MB on
-- topics IT creates — which the auto-created reply topics are not.
```

```bash
# have any of them ever actually been refused? (retention-clocked — record a RATE)
kubectl -n ai-persona-system logs -l app=reasoning-agent --tail=-1 --since=24h | grep -c "Message Size Too Large"
# repeat for content-creator-agent, web-search-adapter, thunder-adapter
```

**`[UNMEASURED]`, and it is the whole severity question:** whether any of the
eight has ever been refused a reply in production. `bugs_closed/062` proves the
failure mode is real *somewhere*; it does not establish these eight are exposed.
A site that never produces a >1MB reply is a latent gap, not an active bug — and
"latent gap in 7 sites" is a different RFC from "active starve in 3".

**LANDMINE for whoever takes this.** `FailedUndeliverable` is a **return value a
caller can ignore.** An adapter that calls `DeliverReply` and drops the outcome
compiles, passes every `platform/kafka` test, and reintroduces the silent starve
unchanged. That is why 133's adoption tests live in the *adapter* package and
assert an error response actually goes out. Copy that shape; do not rely on
`platform/kafka`'s own tests to tell you a site is fixed.

## 1b. THE CONTENT IS STILL DISCARDED — 133 made the claim honest, not the payload

Stated plainly because it is the easiest thing to lose: **`bugs_closed/133` did not
stop content being destroyed.** It stopped us lying about it.

133's fix candidate 1 offered two variants:

> *(a) do not truncate and let B's fix handle an oversized reply, or (b) truncate
> with a marker that says the remainder was **discarded**, not stored. (b) is a
> two-line change and removes the false claim immediately; **(a) is the better
> behaviour but needs B first**.*

133 shipped (b) plus the "best of all" URI-naming variant, **and it also shipped B**
— so (a) is now available for the first time and was deliberately not taken. Why
not: (a) means replies carry full content up to the broker limit and rely on the
degrade path when they do not fit, which changes reply sizes on every scrape in the
fleet and is entangled with the unresolved cap question below. That is a throughput
and cost decision, not a bug fix.

**What still bites, unchanged since 2026-07-28:** the cut is the document **tail**;
a UK company registration number conventionally sits in the page **footer**;
`vet-practice-verifier/scrape_website` exists to extract exactly that and still has
`upload_results` unset. `bugs_open/100` waits on that pipeline. This is
`[INFERRED]` — a mechanism, not a measurement — and 133 flagged it as cheap to
settle: **if company-number extraction has a poor hit rate on pages over 50KB and a
good one under, this is why.** No vet page has been scraped since 2026-03-18, so it
has still never been observed.

**The cheapest thing that would help without deciding (a):** turn `upload_results`
on for the handful of steps whose *purpose* is extraction from a long page. That is
per-step, reversible, and is not the fleet-wide default change that 133's fix
candidate 3 rules an owner call. It also makes the marker *truthful and useful*
rather than truthful and bleak, because there will be a URI to name.

**And the cap question, unresolved, which (a) depends on:**
`system.agent.generic.responses` has no topic-level `max.message.bytes` and the
cluster CR sets none, so ~1MB applies — while
`platform/kafka/topic_manager.go:151,227` sets **5,242,880** on topics it creates,
which the auto-created reply topics are not. Two numbers, no stated intent. 133
flagged it and deliberately did not guess; whoever decides (a) has to decide this
first.

## 2. `storage.pages` is index-misaligned with `result.pages`

`internal/adapters/webscrape/adapter.go:812-856`. The per-page upload record is
**compacted**:

```go
if len(pageInfo) > 0 {
    pageURIs = append(pageURIs, pageInfo)
}
...
uploadInfo["pages"] = pageURIs
```

So `storage.pages[i]` does **not** correspond to `result.pages[i]` whenever any
page uploaded nothing (no `html` and no `markdown`, or both uploads failed). Any
consumer joining the two lists by position gets another page's URIs.

**Not fixed in 133** because the obvious fix — always append, even an empty entry
— changes the `storage` output shape that downstream consumers see, and that is a
behaviour change riding in on a bug patch. 133 worked around it instead, by
resolving a page's URI from the index **embedded in the URI string** (the uploader
builds it from `"%s/pages/page_%d.md"`), which is position-independent.

**Consequence worth stating:** because 133 took the workaround, this defect is now
*harmless to the truncation marker* but still live for **any other consumer** that
reads `storage.pages`. Nobody has been surveyed.

**`[UNMEASURED]`:** who reads `storage.pages`. `grep -rn '"pages"'` across the
consumers of a scrape reply is the check.

**LANDMINE:** the workaround in `pageStorageURIFor` couples the marker to the
uploader's format string. If you fix this misalignment properly, **do not delete
the URI-embedded-index lookup without replacing it** — it is what stops the marker
naming a wrong URI, and it fails to the honest "discarded" form rather than to a
false claim.

## 3. Five of six truncatable per-page fields have no stored copy at any setting

The uploader stores `pageMap["html"]` and `pageMap["markdown"]`. The truncator
cuts `content`, `markdown_content`, `html_content`, `markdown`, `rawHtml`,
`raw_html`. **Only `markdown` overlaps.**

133 handled this honestly rather than fixing it: those five fields now say the
remainder was discarded, because it was, and that is true. But it means a
`firecrawl_crawl` result — **4 live steps**, `site-adoption-agent` and
`vertical-exemplar-researcher` — loses page content over 50KB with no recoverable
copy, per page, silently, on a successful response.

The fix is either to upload the fields that are truncated, or to truncate only the
fields that are uploaded. That is a design decision about what a crawl result is
*for*, which is why it is not a bug patch.

## 4. `platform/kafka.MockProducer` is dead, incomplete, and looks like shared infrastructure

Raised as an objection by the reuse seat on `7478233b` ("why a new test double
rather than extending the shared mock?"). Measured, and the answer is worse than
the question assumed:

* `platform/kafka/mock_producer.go` has **zero callers** anywhere in the tree.
* It implements **2 of `Producer`'s 3 methods** (`Produce`, `Close`) — no
  `ProduceWithValidation` — so **it does not satisfy the interface it claims to
  mock** and never has.
* Meanwhile `platform/orchestration/orchestration_test.go` and
  `test/unit/.../helpers` have each written their own `MockProducer`.

So the reuse seat's founding concern — *"two code paths independently solving
overlapping problems because nobody unified them"* — is **already realised** as
three doubles, and extending the dead one would have unified nothing.

**Why it is worth a line rather than nothing:** it is a trap, not just clutter. It
sits in the platform package, is named like the canonical double, and the next
person to reach for it will find it does not compile against `Producer` — after
writing tests around it. Either delete it or complete it; both are two-minute
changes, and the choice is whether a shared double is wanted at all.

## Fix order

0. **Item 1b's cheap half** — turn `upload_results` on for the few steps whose
   purpose is extraction from a long page (`vet-practice-verifier` first, since
   `bugs_open/100` waits on it). Per-step, reversible, not the fleet-wide default
   change 133 rules an owner call — and it unblocks the company-number hit-rate
   measurement that would settle a two-day-old `[INFERRED]`.
1. **Item 4** — trivial, no behaviour risk, removes a trap. Do it first.
2. **Item 1** — the real work, and it needs an RFC. Size it with the measurement
   above before writing one; the answer may be "7 latent, 2 exposed".
3. **Item 3** — needs an owner decision about crawl results.
4. **Item 2** — lowest value until someone is shown to read `storage.pages` by
   position; 133's workaround removed the only known consumer.

## Related

- `bugs_closed/133` — where all four came from; its workstream is
  `docs024_key_docs_latest/bugfix_133_scrape_truncation_and_delivery/`.
- `bugs_closed/062` — the rule's origin, and the only site that had it before 133.
- `bugs_closed/144` — why the policy was extracted rather than copied.
- `ADP-017` in `docs026_concept_register/register/adapters.md` — the mechanism,
  its landmine, and the architecture seat's RFC caveat.
- `016b §9` — "A false claim in a message is a STRING LITERAL reachable without
  its evidence" and the literal-vs-comment marker pattern, both from 133.

---

# 2026-08-03 — re-validated, sized, and the denominator is WRONG in this file's own table

Worked by a bugs-sweep session. Lane docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_158_reply_drop_sizing/`.
**Still OPEN** — items 1 and 3 remain gated on an RFC and an owner decision
respectively, and this session did not force either.

## Item 4 is ALREADY FIXED — this file is stale on it

`platform/kafka/mock_producer.go` was completed on 2026-07-31 (`76bd35dc1`, as a
side effect of "three suites that stopped compiling when the APIs under them
moved"). It now implements `ProduceWithValidation` and carries a compile-time
assertion:

```go
var _ Producer = (*MockProducer)(nil)
```

with a comment naming exactly the drift this file complained of. **Fix order item 1
("do it first") is discharged.** The three-doubles observation stands; nothing
unified them.

## Item 1 — SIZED, as this file asks. The answer is "latent", and the table undercounts

**The limit is real: ~1MB, confirmed first-hand.**

```
broker message.max.bytes = 1048588   (DEFAULT_CONFIG, no override anywhere)
topic  system.agent.generic.responses — no dynamic config at all
```

> ⚠ **This file's suggested command cannot produce that answer, and the way it
> fails is silent.** There is no Kafka broker in `ai-persona-system` — the cluster
> is `personae-kafka-cluster-combined-pool-prod-{0,1,2}` in namespace **`kafka`**.
> A `kubectl exec` at the wrong pod with stderr discarded returns *empty*, which
> reads exactly like "no topic-level override is set". I made that mistake before
> catching it. Run it against the right pod, and never with `2>/dev/null`.

**Exposure, measured over `llm_call_log` (47,577 calls):**

| | bytes |
|---|---|
| largest response_text ever recorded, fleet-wide | **48,327** |
| calls above 512KB | **0** |
| calls above the 1,048,588 broker limit | **0** |

The largest payload the fleet has ever produced is **4.6%** of the limit. The
services in the table carry LLM output or structured API results, both bounded;
the one site that carries whole documents is **webscrape**, and webscrape is
exactly the one already fixed. So this file's own guess is right and is now
evidence: **"7 latent, 2 exposed"** — the RFC is a consistency argument, not an
incident report. Size it that way.

## ⚠ The "2 of 9" is really 2 of 13 — the census keyed on ONE log-message spelling

This file's grep was `grep -rn "Failed to produce"`. Four more reply-produce sites
swallow their error under **different wording**, so a string-keyed census could not
see them — and they are in the chassis's own messaging spine, not in adapters:

| site | how it swallows | why the grep missed it |
|---|---|---|
| `platform/agentbase/agent.go:887` | error **never captured** | no log line at all to match |
| `platform/messaging/processor.go:2000` | logged, no answer | `"KAFKA_SEND_ERROR: Failed to send message"` |
| `platform/orchestration/coordinator.go:3663` | logged, no answer | `"Failed to notify parent of success"` |
| `platform/orchestration/helpers.go:427` | logged, no answer | `"Failed to send timeout response"` |

All four verified by reading them. This **widens the RFC**: it is no longer four
adapters' failure semantics, it is also `agentbase` and the orchestration
coordinator, which every agent inherits. Whoever writes it must scope to 11, not 7.

## SHIPPED — the class is closed prospectively, with no behaviour change

`scripts/pattern-check.py` gains `check_silent_reply_drop` (`2091903ab`). It cannot
fix the eleven — that needs the RFC — but it stops a **twelfth** being written, and
it makes the population visible without touching any service's semantics, so it
implicates no ruling.

Catches three shapes, because a version that knew only the first missed 3 of the 4
sites named above: `if err := …; err != nil { log }` falling through; the same with
a bare `return`; and the error never captured. It also fires when `DeliverReply`'s
outcome is assigned to `_` — **this file's own landmine**.

**Controls, both run and both recorded:** the 4 known-bad files fire at the exact
lines this file names; both adopted webscrape files stay quiet; fleet sweep is
**8 findings over 777 non-test files** (8 not 11 — one finding per file, and the
`_ = DeliverReply` arm has no live instance). The sweep also caught three FALSE
positives the known-bad control never could: `return producer.Produce(...)`
*propagates* the error, and two produces to **request** topics matched only because
"response" appeared a few lines away. The reply gate now reads the call's own
arguments.

## Item 0 (1b's cheap half) — premise confirmed, NOT actioned, and here is the survey

22 live scrape/crawl steps fleet-wide: **3** set `upload_results: true`, **1** sets
it false, **18 unset** — including `vet-practice-verifier/scrape_website`, exactly
as this file says. Left alone deliberately: it is another lane's pipeline
(`bugs_open/100`), the change is inert until a vet page is next scraped (none since
2026-03-18), and turning it on for someone else's pipeline unannounced is what the
ownership rule exists to prevent. **The survey is the handover** — the 18 are listed
in the lane's RUNBOOK.

## Item 2 — its `[UNMEASURED]` is now measured: no consumer found

`grep` for readers of a scrape reply's `storage` across Go, SQL and live
`agent_definitions` configs finds only `truncation.go`'s own position-independent
workaround. `storage_actions.go:171,529` reads an *agent config* `storage` block —
a different thing entirely. On that evidence this file's ranking (lowest value) is
right and it should stay unfixed until a consumer exists.

## What is still owed

1. **The RFC for item 1**, scoped to **11 sites, not 7**, and argued as consistency
   rather than incident — the numbers above are the sizing it asked for.
2. **Item 3** — an owner decision about what a crawl result is for.
3. **Item 2** — nothing, until a consumer of `storage.pages` exists.

---

# 2026-08-03 (later) — OWNER RULING on item 1, and a correction to my own census

## OWNER RULING, 2026-08-03: option (b) — the PLUMBING sites only

Presented with the sizing above and four options, the owner ruled: **adopt
`DeliverReply` at the sites every agent inherits, and leave the four adapter/agent
sites alone.** That discharges the council architecture seat's RFC condition on
`7478233b` for this scope — the decision was taken by a human on measured evidence,
which is what the RFC was for.

**SHIPPED** in `2d976c026`. Council `f13212d4-2787-448f-bf49-b57506ded74e`
(submitted; verdict pending at the time of writing).

| site | what it now does when the reply cannot be delivered |
|---|---|
| `coordinator.go` `notifyParentOfSuccess` | tells the parent the workflow **FAILED**, via the existing `notifyParentOfFailure` — the same answer the adjacent size-precheck path already chose |
| `agentbase/agent.go` `processMessage` | degrades to headers + a marked `ERROR_FORWARD_TOO_LARGE` stub so the error **chain** survives when the payload cannot; logs plainly if even that fails |
| `helpers.go` `sendTimeoutResponse` | stops logging an **undelivered** notice at the same level as a delivered one (this reply IS the error response — there is no fallback to choose) |

Detector sweep: **7 → 4**. The four remaining are the deferred adapter/agent sites.

## ⚠ CORRECTION — one of the four sites I added yesterday was NOT an instance

`platform/messaging/processor.go:2000` **does not swallow its error.** It
`return err`s two lines below the block, and its enclosing function
`sendErrorResponseOLD` has **zero callers** — it is dead code. My detector read only
the *body* of the `if err != nil` block, so it missed the return; I then copied its
output into this ticket and into `016b §9` without opening the file far enough.

- **The census is 2 of 12, not 2 of 13.** The other three additions
  (`agentbase:887`, `coordinator` (now `:3721`), `helpers:427`) were re-verified
  line by line and are genuine.
- The detector is fixed (`2d976c026`): an error **returned after** the block is
  propagated, not swallowed. Sweep went 8 → 7 before the adoptions.
- Logged in `WRONG_CALLS.md`.

**Also worth knowing:** `coordinator.go`'s site moved `:3663 → :3721` between my
sweep and my edit, because another session committed to that file in between. Line
numbers in this ticket are a snapshot; find these by symbol, not by line.

## STATUS

**Still OPEN**, and now for exactly two reasons:

1. **Item 3** — an owner decision about what a crawl result is *for*. Untouched.
2. **The four adapter/agent sites** — deliberately deferred by the owner's ruling,
   not forgotten. They are held by `check_silent_reply_drop`, so no fifth can
   appear quietly. If they are ever to be adopted, that is a fresh decision.

Items 4 (already fixed), 2 (no consumer exists) and 0 (surveyed, handed to
`bugs_open/100`'s owner) need nothing further from this lane.

---

# 2026-08-03 (evening) — decisions 1, 2 and 5 actioned

## Decision 1 — item 3, "upload everything that gets truncated": SHIPPED, council pending

`a92a32fba`. Two layers: `uploadScrapingResults` now uploads all 6 per-page fields
(was 2 — and fixes a real key-name mismatch found while extending it: the old code
checked `pageMap["html"]`, which never matched `batch_handler.go`'s own key
`html_content` for the same concept, so per-page HTML for that path was silently
never archived even with `upload_results:true`). `truncateResultForTransport` and
`stripResultForRetry` gain a `FieldUploader` fallback that uploads the FULL
pre-truncation content the moment a field is about to be cut and has no static URI
yet — closing the gap for the 18 of 22 steps that never set `upload_results` at all.
`pageFieldsNeverUploaded` is retired.

8 new tests, all mutation-proved (4 independent mutations, each caught). All ~13
pre-existing tests pass unchanged. Full build + suite green.

**Council `ee9f6210-3bda-4efa-a25c-92ce4a7666a1` submitted, verdict PENDING** at the
time of writing — `review_debug_historian` was still running after 13+ minutes, an
outlier against this lane's other rounds (all landed in 2-6 minutes). Check it before
trusting this as approved:
```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = 'ee9f6210-3bda-4efa-a25c-92ce4a7666a1';
```
Commit carries `Council-Submitted:` — correct for a pre-verdict commit, `098`
resolves it automatically once approved. **Not yet live**: needs a chassis roll
after approval, then pod-verify per this file's own established discipline.

## Decision 2 — align the under-provisioned reply topics: DONE, live immediately

Kafka topic config, no chassis roll needed. **74 `*.responses` topics** were on the
1MB broker default (re-measured fresh, not the earlier 91 — topics churn
continuously); all now carry the 5MB override. One (`system.thunder.smoke.responses`)
appeared between the snapshot and the batch alter and needed a second pass. **Final
verification: 0 of 3,828 responses topics missing the override.**

## Decision 5 — unblock `bugs_open/100`: DONE

`upload_results: true` set live on `vet-practice-verifier/scrape_website`
(`agent_definitions`, DB config, no roll needed). Note added to `100` itself
(`e4de1a7d0`) rather than only here, since the owner starts that bug in a separate
thread.

## Decisions 3 and 4 — explained to the owner, no action taken pending his read
