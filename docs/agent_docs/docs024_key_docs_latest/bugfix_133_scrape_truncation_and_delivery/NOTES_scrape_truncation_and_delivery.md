# NOTES — bugs_open/133 (scrape truncation marker + reply delivery)

Append-only, newest at the bottom. Evidence, commands, and every misstep.

---

## §1 — 2026-07-30 ~15:05 — taking the bug, and a false OWNED verdict

`scripts/who-owns.py 133` said **OWNED or recently active**: 21 mentions in
`bugfix_100_101_scrape_provenance`. My memory line said "OPEN, unowned". One of
them was wrong.

Reading the mentions settled it in one command:

```bash
grep -n "133" docs/agent_docs/docs024_key_docs_latest/bugfix_100_101_scrape_provenance/*.md
```

Every hit is a **disclaimer** — "filed, **not fixed**", and at
`HANDOFF_2026-07-28b:20` the explicit *"neither fixed and neither this lane's to
fix."* The memory line was right; `who-owns` was reporting mention density.

> **The transferable bit:** a lane that files a bug it will not fix is
> indistinguishable, to a mention count, from a lane that is fixing it — and the
> filing lane will always be the top hit, because filing a bug means writing
> about it more than anyone else. `who-owns` is advisory and says so; the
> mistake would have been to accept the verdict and go elsewhere, leaving a bug
> nobody owns permanently unowned *because* it was well documented.

Also checked, all clear: `git log --since="5 days ago" -- internal/adapters/webscrape/`
(one commit, `2ebabf2ca`, a different concern), `git status` on the two packages
(clean, so nobody mid-edit), and the work-item queue (two unrelated content rows).

## §2 — the census that changed the shape of the fix

The bug's fix candidate 2 says "apply 062's fix to `sendSuccessResponse`". Before
copying anything I asked how widespread the *absence* was:

```bash
grep -rn "MessageSizeTooLarge\|Message Size Too Large" --include=*.go . | grep -v _test
#   -> internal/adapters/webscrape/batch_handler.go   ONLY

grep -rn "Failed to produce" --include=*.go internal/ platform/
#   -> 9 reply-producing sites across 5 services; 8 are log-and-return
```

So `016b §9`'s rule held at **1 of 9 sites**, and the one implementation was
unexported with a single caller. Copying it would have made it 2 of 9 and created
a second hand-written copy of one rule — **which is precisely the defect
`bugs_closed/144` was about, and the reuse seat objected to that exact shape in
144's own council round.** Hence the shared seam.

> **Run BOTH greps, not one.** The first alone reads as "one place handles this,
> copy it". Only the *ratio* says the rule is systematically missing and that
> copying is the wrong move. A single-number measurement would have justified the
> copy I was trying to avoid.

## §3 — MISSTEP: "raw_html is never uploaded" — and my grep window was the bug

Reading `uploadScrapingResults` with `sed -n '525,700p'`, I found uploads for
`html_content`, `markdown_content` and the screenshot, and **no `raw_html`**. The
truncated fields are `markdown_content`/`html_content`/`raw_html`. I was ready to
write a new finding into the bug: *even the two `upload_results: true` rows are
unsafe for raw_html.*

The function is ~340 lines. `raw_html` is uploaded at **line 760**, past my
window. Re-grepping the whole function found it immediately:

```bash
sed -n '525,830p' internal/adapters/webscrape/adapter.go | grep -n 'uploadInfo\['
```

**Caught by widening the window, not by luck** — but the window had been chosen
to answer the question I was already asking, which is the
`narrow-filter-defines-the-conclusion` shape. A confident false finding was about
to go into a bug file where the next thread would have believed it.

## §4 — MISSTEP, same shape, minutes later: "per-page content is never uploaded"

Same error, same cause. Pages **are** uploaded (`pageMap["html"]`,
`pageMap["markdown"]` → `storage.pages[i]`), again just past the window I looked
at. Twice in ten minutes, with the lesson already written down from §3.

What survived is narrower and is real: the uploader stores `html`/`markdown`
while the truncator cuts `content`/`markdown_content`/`html_content`/`markdown`/
`rawHtml`/`raw_html`, so **only `markdown` overlaps** — the other five per-page
fields have no stored copy at any setting and their marker can never be true.

## §5 — the compaction trap, found by reading rather than by failing

`uploadInfo["pages"]` is built with `if len(pageInfo) > 0 { pageURIs = append(...) }`
— **compacted**. So `storage.pages[i]` is *not* `result.pages[i]` whenever a page
uploaded nothing. Resolving a page's URI by list position would have attached
**another page's URI** to this page's marker: a false claim of exactly the kind
this bug is about, arrived at by fixing it carelessly.

Fix used instead: the URI carries its own index, because the uploader builds it
from `"%s/pages/page_%d.md"`. Searching for `/page_<i>.` is position-independent
and, if the convention ever changes, returns `""` so the marker degrades to the
honest DISCARDED form. **It cannot degrade into a wrong URI** — the failure
direction matters more than the failure rate.

(The trailing dot in the needle is load-bearing: `page_1` must not match
`page_10`. That is a test, `TestPageURIIndexMatchIsNotAPrefixMatch`, not a
comment.)

The misalignment itself is a separate defect. Left unfixed deliberately: fixing
it changes the `storage` output shape, and the bug file's own guidance is not to
ride a behaviour change in on a bug patch.

## §6 — MISSTEP: my own test caught my own comment, and the fix was better than the test

`TestTheOldUnconditionalClaimIsUnreachable` scans the package source for the
sentence `full version in S3`. It **failed** on first run — on `truncation.go`
line 16, where I *quote the old marker in a doc comment* to explain the defect.

The test was right to fire and the scan was wrong: a bare substring scan cannot
tell an emittable string from a comment describing it. Rewrote it to parse with
`go/parser` and inspect only `*ast.BasicLit` string literals.

> **This is the same distinction that made `bugs_open/153`'s evidence table
> unsound** (four of its five pod-grep markers were doc prose, the fifth was
> inside a Go comment). Getting it wrong in a *test* is cheap; getting it wrong
> in a verification recipe cost another session a false conclusion about which
> binary was running. A claim only ships if it is in a literal.

Verified both directions: reintroducing the sentence as a real literal fails the
test; the same sentence in a comment passes.

## §7 — the deploy check, made a real test in both directions BEFORE shipping

Pre-fix baseline on all three live replicas, so the post-roll check has a
known-bad answer as well as a known-good one:

```
OLD lying marker  1 · NEW honest marker  0 · control  1     (x3 replicas, v1.0.1208)
```

Then, before pushing, on the built image locally — which answers "did my commit
make the build" with no deploy at all:

```bash
docker run --rm --entrypoint sh docker.io/aqls/web-scrape-adapter:v1.0.1209 -c \
  'strings /app/web-scrape-adapter | grep -c "<marker>"'
```

```
NEW 1 · OLD 0 · degrade-log 1 · undeliverable->error 2 · control 1 · never-existed 0
```

After the roll, on all three replicas of v1.0.1209: identical, counts moved in
opposite directions from the baseline. Digest
`sha256:e2d376bdc92c6b09c38268230b640ae3470f41210ef19ba3fa9afa4d06d5d90b`.

> **`git grep` returned 0 for my own new marker and it was not a problem with the
> marker.** `git grep` searches *tracked* files; `truncation.go` was still
> untracked, so it was skipped — silently, exit 0, no error. Another member of
> the 0-matches-and-no-error family. Use plain `grep` on the working tree, or
> `git add` first.

## §8 — another session rolled webscrape under me, and I re-verified rather than reasoned

Partway through, the overlay changed beneath me to **v1.0.1211** (another
session's build). My fix is in HEAD, so a later build "must" contain it — which
is exactly the assumption `bugs_open/153` exists to warn about. Re-ran the
markers on all three new pods: NEW 1, OLD 0, control 1. **Verified, not
inferred.** A tag that is numerically later than mine is not evidence it was
built from a commit later than mine.

## §9 — functional proof: the measured incident, reproduced on the fixed binary

Fired one probe scrape of `https://vetcomparison.uk` with `upload_results: false`
— the filing lane's own target and settings. Reply topic
`system.probe.webscrape.watch062` already existed (seeded by that lane), so I
could not manufacture the `Failed to produce` signal I was testing near.

Probe correlation `35c24f46-9f58-4015-b473-0b4d891f2dfa`:

```
22:30:48  Processing webscrape request
22:30:54  Truncating large field for Kafka  field=raw_html original_len=53536 truncated_to=50000 stored_copy=false storage_uri=""
22:30:54  Truncated content was DISCARDED, not stored — the reply carries less than was scraped  field=raw_html discarded_chars=3536 ref=bugs_open/133
22:30:55  Successfully produced message
22:30:55  Request processed successfully
```

Compare the bug's original measurement (2026-07-28, v1.0.1192), which was the
same page and the same field:

```
Truncating large field for Kafka  field=raw_html original_len=53805 truncated_to=50000
```

— and nothing else. **The defect's fingerprint (`stored_copy=false`) is now
observable, and it was previously unobservable by construction:** the old code
logged the truncation without ever recording whether the copy it claimed existed.

And the delivered message — read off the reply topic, which is what a caller
actually receives:

```
status        : complete
truncated     : True
truncated_fields : ['raw_html']
has storage   : False
raw_html len  : 49983
marker tail   : "...The Competition and M\n\n[Content truncated for Kafka transport at 50000 chars
                 - the remainder was DISCARDED and no copy was stored]"

old lying sentence anywhere in the reply?  False
honest DISCARDED marker present?          True
```

Checked rather than assumed: the 50,000-**byte** cut yields 49,874 characters and
**zero** U+FFFD replacement characters, so the byte/char difference is encoding
arithmetic, not corruption at the cut point. I nearly wrote up "the cut can split
a UTF-8 character" as a minor finding; it does not, here.

## §10 — THE BUG'S OWN EXPOSURE TABLE WAS TOO NARROW, and this is the biggest correction

The bug says *"4 of the 6 single-URL scrape steps in the fleet"* and its query
filters `v->>'action' IN ('scrape_web','firecrawl_scrape','batch_webscrape')`.

Reading a real message off the request topic (to copy a shape the adapter
accepts) showed production traffic from **`feed-ingester`** with
`upload_results: false` — an agent that appears nowhere in that table. Its step
uses action **`fetch_scrape`**, which the query never looked for.

There are at least **six** actions that reach this adapter:

```bash
grep -rn "\"fetch_scrape\"\|\"scrape_web\"\|\"firecrawl_scrape\"\|\"firecrawl_crawl\"\|\"firecrawl_extract\"" platform/orchestration/actions/*.go
# registry.go: scrape_web, firecrawl_scrape, firecrawl_crawl, firecrawl_extract, fetch_scrape (+ batch_webscrape)
```

Re-measured across all six (`is_active`, not snapshot, not deleted):

| action | upload_results | steps | agents |
|---|---|---|---|
| `firecrawl_scrape` | **false** | 1 | site-scraper |
| `fetch_scrape` | **(unset)** | 1 | feed-ingester |
| `firecrawl_crawl` | **(unset)** | 4 | site-adoption-agent, vertical-exemplar-researcher |
| `firecrawl_scrape` | **(unset)** | 1 | site-adoption-agent |
| `scrape_web` | **(unset)** | 2 | domain-research-classifier, vet-practice-verifier |
| `firecrawl_scrape` | true | 2 | website-capture-firecrawl, website-extract-structured |
| `firecrawl_extract` | true | 1 | website-extract-structured |
| `batch_webscrape` | (unset) | 5 | 5 researcher agents — batch path, different handler |

`(unset)` **is** exposed: `webscrape_actions.go:209` reads
`uploadResults := false` then only overwrites it if the key is present, and the
adapter side (`adapter.go:218`) does `data["upload_results"].(bool)`, which
yields `false` on a missing key.

**So the single-URL exposure is 9 of 14 steps, not 4 of 6** — and it includes
`feed-ingester`, which is the highest-volume real scraper on the topic (messages
07-29 19:54Z and 07-30 07:57Z), and four `firecrawl_crawl` steps, which are
exactly the multi-page case where five of six per-page fields can never have a
stored copy.

> **The prior-art-librarian seat asked for this claim to be re-checked against
> `agent_definitions` before the scope argument was treated as closed. It was
> right to, and acting on it more than doubled the measured exposure.** The bug's
> figure was not wrong about what it counted; the *filter* described a small
> world. Same shape as §3 and §4, in someone else's measurement rather than mine
> — which is why "re-run it rather than trusting the table" is in the bug file
> and should be obeyed literally, including re-deriving the WHERE clause.

## §11 — council round: approved, and the one thing I could not settle myself

Submitted `7478233b-3986-4505-a747-c059dc87e9e7` at 15:23Z, **APPROVED 15:32Z**
(9 minutes; 12 seats reviewed, 6 abstained, 2 objecting verdicts, none
high-severity in the aggregate).

I deliberately flagged my own seam in the rationale — *"if the architecture seat
disagrees that additive-and-inert is the right reading for a file in
platform/kafka specifically, I want to hear that as an objection — it is the one
judgement in here I cannot settle by measurement."* Two seats engaged:

* **guardian** (medium): held the stability-preference concern anyway, because
  *"'inert today' does not mean 'inert once other 8 sites see a ready-made shared
  function sitting in platform/kafka'"*. Contained alternative: keep the helper
  package-local and defer promotion until a second **service** needs it.
* **architecture** (approve, `ARCHITECTURE_SIGNAL: point_fix`): the multi-package
  trigger has not fired — one consuming top-level package, two call sites — and
  the 2026-07-29 ruling carves out exactly this shape. But with a caveat I have
  copied into ADP-017 verbatim: **"when adoption widens past webscrape, that IS
  the RFC moment ... the author should not treat ADP-017 registration here as
  pre-clearing that later expansion."**

Two seats reaching opposite defensible conclusions is the shape that produced
`RFC_002`; this time the round resolved it, and the resolution is recorded where
the mechanism lives rather than only in the verdict.

Objections acted on with measurement rather than argument:

* **reuse_agent** asked whether a generic transport-truncation helper already
  existed in a sibling adapter. **No** — every `truncat*` outside webscrape
  (`imagegenerator/banana`, `imagegenerator/stability`, `thunder/api`) is
  *log-preview* truncation, `truncate(s, 400)` for a log field. Nothing to extend.
* **reuse_agent** also asked whether `platform/kafka.MockProducer` should have
  been extended instead of a new test double written. Measured, and the answer is
  worse than the objection assumed: **that mock has ZERO callers and implements
  only 2 of `Producer`'s 3 methods, so it does not satisfy the interface it
  claims to mock** — while `platform/orchestration/orchestration_test.go` and
  `test/unit/.../helpers` have each written their own. The seat's founding concern
  is already realised as three doubles; extending a dead one would unify nothing.
  Recorded, not fixed — out of scope, and filed.
* **debug_historian** (medium) noted the submission described git-based isolation
  checks but no pod-grep. Fair against the submission: the pod work in §7/§8 was
  done but was not in the plan text. It is in the RUNBOOK now.
* **editquality** (medium/low) objected that my mutation-testing claims ("verified
  to fail when the guard is neutered") are narrative a reviewer cannot rely on.
  Correct as stated. The response is not to argue: the exact `sed` commands are
  now in the RUNBOOK so the mutation is reproducible rather than asserted.
* **guidelines** (approve) added a forward note worth keeping: the new response
  keys `truncated`/`truncated_fields`/`degraded_for_transport` are the shape that
  "could quietly become an undeclared dependency the day some workflow starts
  branching on them", since DECLARED CONTRACTS only triggers on what a workflow
  *reads*. Nothing reads them today.

## §12 — trailer mistake I cannot amend

The seam commit `6d71efa69` carries `Council-Submitted: pending` — written before
I had submitted, so it names no correlation and the `098` report can never resolve
it. Forward-only forbids an amend. The real correlation is on the second commit
(`92e50d038`) and the verdict is approved, so the *change* is credited; the seam
commit will read as un-reviewed for ever.

> **Do not write a placeholder in a trailer.** `Council-Submitted:` exists
> precisely so you can commit before a verdict — but its whole value is the
> correlation, so a trailer without one is strictly worse than no trailer
> (it looks like coverage and resolves to nothing). Submit first, then commit.
