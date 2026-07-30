# PLAN — 2026-07-30 — bugs_open/133: a truncation marker that lies, and a reply that vanishes

Owner thread: session "bugsearch 8" (took 133 on 2026-07-30).
Bug file: `bugs_open/133_HANDOFF_2026-07-28_single_scrape_path_truncates_to_a_store_that_was_never_written.md`
Filed by the `bugfix_100_101_scrape_provenance` lane, which explicitly did **not**
take the fix ("neither this lane's to fix", `HANDOFF_2026-07-28b:20`).

---

## 1. What the bug is, in one line each

* **A.** `internal/adapters/webscrape/adapter.go` truncates content at 50,000 chars
  and appends `[Content truncated for Kafka transport - full version in S3]`
  **unconditionally**, while the S3 upload that would make that sentence true is
  guarded by `if uploadResults && a.storageClient != nil`. When it is false,
  content is destroyed and the message claims a copy exists.
* **B.** `sendSuccessResponse` ends `log.Error(...)` and returns. A reply the
  broker refuses is never delivered, never degraded, and never turned into an
  error — the caller waits on the reply topic and starves.

## 2. Ownership check (done first, per CLAUDE.md)

`scripts/who-owns.py 133` returns **OWNED or recently active** — 21 mentions in
`bugfix_100_101_scrape_provenance`. **That verdict is a false positive here** and
the reason matters: `who-owns` counts *mentions*, and every one of those 21 is a
disclaimer ("filed, **not fixed**", "neither this lane's to fix"). A lane that
files a bug it will not fix looks identical, to a mention count, to a lane that
is fixing it. Read the mentions, not the verdict.

Also checked, all clear:
* `git log --since="5 days ago" -- internal/adapters/webscrape/` → one commit, `2ebabf2ca` (100/101, provenance — a different concern in the same package).
* `git status internal/adapters/webscrape/ platform/kafka/` → clean, so no session is mid-edit.
* `site_work_items` open rows matching scrape/truncat → two unrelated content items.

## 3. The measurement that sets the scope

The bug's exposure table, **re-run 2026-07-30, unchanged**: of the 6 single-URL
scrape steps live in the fleet, **4 are exposed** (`site-scraper` explicitly
`false`; `domain-research-classifier`, `site-adoption-agent`,
`vet-practice-verifier` unset) and 2 are safe (`upload_results: true`).

Then the census that decides how wide the fix should be:

```
$ grep -rn "MessageSizeTooLarge\|Message Size Too Large" --include=*.go . | grep -v _test
internal/adapters/webscrape/batch_handler.go   (only)
```

**Exactly one function in the whole codebase knows that an oversized reply is a
deterministic failure**, and it is unexported with one call site. Meanwhile
`grep -rn "Failed to produce"` finds **9 reply-producing sites across 5
services** (reasoning, contentcreator, websearch, webscrape ×2, thunder ×2), and
8 of them are `log and return` — the silent starve. So `016b §9`'s own rule,
*"a response that cannot be delivered must become a deliverable error, never
silence"*, currently holds at **1 of 9 sites**. Defect B is one instance of a
framework-level gap.

## 4. Design — and the trap I am deliberately avoiding

The obvious fix for B is to copy `batch_handler.go`'s degrade-resend-else-error
block into `sendSuccessResponse`. **That would recreate, inside 133's fix, the
exact defect `bugs_closed/144` was about**: two hand-written implementations of
one rule, which then drift. The reuse seat objected to precisely this shape in
144's own council round, and it was right.

So instead:

**(i) One shared policy, in `platform/kafka/reply_delivery.go`.**

* `IsMessageTooLarge(err) bool` — exported, the typed-checks-then-substring
  predicate lifted verbatim from `batch_handler.go` (including the comment
  explaining why the substring fallback stays).
* `DeliverReply(...)` — produce; on success return; on a *non*-size error return
  it unchanged (transient, the coordinator's retry is the right answer); on a
  size refusal call the caller's `degrade` callback and resend **once**; if that
  also fails, return an outcome the caller must answer with an error response.

  It returns a typed `DeliveryOutcome`, so the *envelope* stays with the caller
  (each adapter's error response has its own shape) while the *policy* lives in
  one place. Callers cannot get the policy subtly wrong; they can only choose
  how to degrade and how to report.

**This is a platform seam** (CLAUDE.md, owner rulings 2026-07-28/29), so:
* It is **additive and inert** — nothing calls it until a call site opts in, and
  no existing consumer's guarantee changes. Per the 2026-07-29 ruling §1 that
  makes it normal council-gate scope, not an RFC: it adds an opt-in capability
  rather than changing what a shared mechanism *guarantees*.
* It is **registered in the concept register in the same commit that ships it**
  (condition (2), which still stands).
* Its other consumers are **named and told**, not merely measured (§3 ruling):
  the 8 remaining produce sites, filed as a follow-up bug with the census.
* I claim **no ordering constraint** — condition (1) was retired 2026-07-29 and
  on this shared tree I could not hold the change back anyway.

**(ii) Make the marker unable to lie, in `internal/adapters/webscrape/truncation.go`.**

The root cause of A is that `"full version in S3"` is a **string literal
reachable without a URI in hand**. Fix candidate 1's best variant in the bug file
says it: *"pass the storage URI into the marker so it is impossible to claim a
copy without naming one."* So the marker is derived from the URI:

```go
func transportTruncationMarker(uri string) string   // "" => says DISCARDED
```

There is deliberately **no way to spell the claim without a URI**. A future
edit cannot reintroduce the lie without deleting a parameter.

The truncator then resolves a URI **per field**, because the upload is per-field
and best-effort — `html_content`→`html_uri`, `markdown_content`→`markdown_uri`,
`raw_html`→`raw_html_uri`, each set only `else` of a `logger.Warn` on failure. So
even with `upload_results: true`, one failed upload leaves the old marker lying;
per-field resolution is not gold-plating, it is the difference between honest and
usually-honest.

**(iii) Both webscrape paths call the one policy.** `sendSuccessResponse` adopts
`DeliverReply` (closing B), and `sendBatchSuccessResponse` is repointed at it too,
deleting its private copy. Two callers, one implementation — the 144 rule.

## 5. Field→URI map, and the drift guard

| truncated field | uploaded as | can the marker name a copy? |
|---|---|---|
| `markdown_content` | `storage.markdown_uri` | yes |
| `html_content` | `storage.html_uri` | yes |
| `raw_html` | `storage.raw_html_uri` | yes |
| per-page `markdown` | `storage.pages[].markdown_uri` | yes, by URI-embedded index (see §6) |
| per-page `content`, `markdown_content`, `html_content`, `rawHtml`, `raw_html` | never uploaded | **no** — always the discarded form |

Two lists that must agree is the drift class this repo keeps getting bitten by,
so a test asserts every field in the truncate lists is either in the URI map or
explicitly named as never-uploaded. Adding a field to one list and not the other
fails the test. (Same shape as `SubWorkflowStepFields` in 144's fix.)

## 6. Two corrections I made to my own reading before writing any code

> **CORRECTED (before commit): "raw_html is never uploaded" was WRONG, and it
> was my grep window that was wrong, not the code.** I read
> `sed -n '525,700p'` of `uploadScrapingResults`, saw uploads for
> `html_content`/`markdown_content`/screenshot and no `raw_html`, and was ready
> to file "even the `upload_results: true` rows are unsafe for raw_html" as a new
> defect extending the bug. The function is 340 lines; `raw_html` is uploaded at
> **line 760**, past my window. Re-grepping the *whole* function found it. A
> filter chosen from the question I was asking would have produced a confident
> false finding in a bug file — the `narrow-filter-defines-the-conclusion`
> pattern, caught by widening rather than by luck.

> **CORRECTED (same way, minutes later): "per-page content is never uploaded"
> was also wrong** — pages *are* uploaded (`pageMap["html"]`, `pageMap["markdown"]`
> → `storage.pages[i]`), again just past my window. What survives is narrower and
> is a real finding: the upload covers `html`/`markdown` while the *truncation*
> covers `content`/`markdown_content`/`html_content`/`markdown`/`rawHtml`/`raw_html`,
> so **only `markdown` overlaps**; the other five per-page fields have no stored
> copy and the marker on them can never be true.

**A third finding, unfiled and NOT fixed here:** `uploadInfo["pages"]` is
**compacted** — `if len(pageInfo) > 0 { pageURIs = append(...) }` — so
`storage.pages[i]` does **not** correspond to `result.pages[i]` when any page
uploads nothing. Naming a URI by list index would therefore attach *another
page's* URI to this page's marker: a false claim of exactly the kind this bug is
about, arrived at by fixing it carelessly. Mitigation used instead: resolve the
page URI by searching for the `page_<i>.` substring the upload's own format
string (`"%s/pages/page_%d.md"`) puts *inside* the URI, which is
index-independent and degrades to the honest discarded form if the naming
convention ever changes. The misalignment itself is a separate defect and is
filed separately — fixing it here would be an output-shape change riding in on a
bug patch, which is what the bug file's own "do not do 3 as a side effect of 1
or 2" warns against.

## 7. Explicitly out of scope

* **Fix candidate 3** (default `upload_results` to true) — a behaviour and cost
  change across four other lanes' agents. The bug file already rules it an owner
  call and says not to do it as a side effect. Not done.
* **Adopting `DeliverReply` in the other 8 produce sites** — changes what callers
  of reasoning/contentcreator/websearch/thunder observe on failure (an error
  response instead of a timeout). Strictly better, but it is four other services'
  behaviour and belongs in its own reviewed change. Filed with the census so the
  next thread starts from a measurement.
* **Fix candidate 4** (the `deploy/…` vs `-l app=…` watch command) — already done
  by the filing lane in `2f96fa70c`; verified, not redone.
* The `max.message.bytes` 5MB-vs-1MB question the bug notes in passing: recorded
  as an open owner decision, not silently resolved.

## 8. How this gets verified

Go code is inert until an image is built and rolled, so: unit tests offline, then
a pod-grep on a string this change **adds** plus one it **deletes**, plus a
positive control — and per `bugs_closed/144`'s lesson, the delete-marker is
checked with `git grep -c` first and must not be a substring of its replacement.
Commands in the RUNBOOK.
