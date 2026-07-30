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
