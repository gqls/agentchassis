# RUNBOOK — bugs_open/133 (scrape truncation marker + reply delivery)

Every command here had to be got right once. The gotcha is attached to each.

---

## Ownership, before touching anything

```bash
./scripts/who-owns.py 133
```

**Gotcha: it says OWNED for 133 and that is a FALSE POSITIVE.** It counts
*mentions*, and all 21 are in `bugfix_100_101_scrape_provenance`, which filed the
bug and disclaimed the fix. Read the mentions before believing the verdict:

```bash
grep -n "133" docs/agent_docs/docs024_key_docs_latest/bugfix_100_101_scrape_provenance/*.md
# every hit says "filed, not fixed" / "neither this lane's to fix"
```

A lane that files a bug it will not fix is indistinguishable, to a mention count,
from a lane that is fixing it. Also check the tree, because a session mid-fix has
no commits yet:

```bash
git status --short internal/adapters/webscrape/ platform/kafka/
git log --oneline --since="5 days ago" -- internal/adapters/webscrape/
```

## Live exposure — which steps are actually at risk

```sql
SELECT ad.type AS agent, e.k AS step, v->>'action' AS action,
       COALESCE(v->'config'->>'upload_results','(unset)') AS upload_results
FROM agent_definitions ad,
     jsonb_each(ad.default_config->'workflow'->'steps') AS e(k,v)
WHERE ad.deleted_at IS NULL AND COALESCE(ad.is_snapshot,false)=false AND ad.is_active
  AND v->>'action' IN ('scrape_web','firecrawl_scrape','batch_webscrape')
ORDER BY upload_results, ad.type, e.k;
```

Re-run it; do not trust the table in the bug file. Confirmed unchanged 2026-07-30:
4 of 6 single-URL steps exposed (`site-scraper` explicitly false;
`domain-research-classifier`, `site-adoption-agent`, `vet-practice-verifier` unset).

**Gotcha: this walks only TOP-LEVEL steps.** A step inside a loop's sub-workflow
is invisible to it — that is `bugs_closed/144`. For scrape actions specifically
the count happens to be complete (checked with a `default_config::text LIKE`
cross-check), but do not reuse this query shape for a different action and assume
completeness.

## The census that decides how wide a fix should be

```bash
# who knows an oversized reply is DETERMINISTIC (was: exactly one file)
grep -rn "MessageSizeTooLarge\|Message Size Too Large" --include=*.go . | grep -v _test

# how many reply-producing sites there are (was: 9 across 5 services)
grep -rn "Failed to produce" --include=*.go internal/ platform/
```

**Gotcha: run BOTH.** The first alone says "one place handles this"; only the
ratio says the rule holds at 1 of 9 and that copying it is the wrong move. A
single-number measurement would have justified the copy.

## Tests

```bash
go test ./platform/kafka/ ./internal/adapters/webscrape/
```

**Gotcha: a passing suite is not evidence the mechanism is tested.** Prove the
tests can fail, by removing the thing they guard:

```bash
# the delivery policy: make every failure look transient
sed -i 's|if !IsMessageTooLarge(err) {|if true {|' platform/kafka/reply_delivery.go
go test ./platform/kafka/ -run TestDeliverReply     # MUST fail (4 tests)
git checkout platform/kafka/reply_delivery.go

# the marker: reintroduce the old claim as a real string literal
# (a comment containing it must NOT fail — that is the point of the ast check)
```

Verified 2026-07-30: 19 assertions in `platform/kafka`, and the literal-scan test
fails on a reintroduced literal while passing with the same sentence present in a
comment.

## Verify it is LIVE — and note WHICH image

**Gotcha: this fix ships in `web-scrape-adapter`, NOT the chassis.** The seam is
in `platform/kafka` so it is compiled into the chassis too, but it is inert
there; the behaviour change is in the adapter. Rolling the chassis proves nothing
about this bug.

```bash
# 1. Check the markers exist in the source you expect to be running FIRST —
#    one command, no cluster. bugs_closed/144: a marker must be a string the
#    binary EMITS, never a symbol, comment or doc phrase.
git grep -c "the remainder was DISCARDED and no copy was stored" -- '*.go'   # want >=1
git grep -c "full version in S3" -- '*.go'                                   # want 0 in LITERALS
#    ^ NOTE: this returns 1 because truncation.go QUOTES the old marker in a
#      comment to explain the defect. That is exactly why the repo test parses
#      string literals with go/ast instead of grepping. Do not "fix" the comment.

# 2. Then the pod.
POD=$(kubectl -n ai-persona-system get pods -l app=web-scrape-adapter -o name | head -1)
kubectl -n ai-persona-system exec $POD -- sh -c \
  'strings /app/web-scrape-adapter | grep -c "the remainder was DISCARDED and no copy was stored"'  # want >=1  ADDED
kubectl -n ai-persona-system exec $POD -- sh -c \
  'strings /app/web-scrape-adapter | grep -c "Content truncated for Kafka transport - full version in S3"'  # want 0  DELETED
kubectl -n ai-persona-system exec $POD -- sh -c \
  'strings /app/web-scrape-adapter | grep -c "Starting scrape"'  # positive control, want >=1
```

**Gotcha: the delete-marker is the load-bearing one, and it must not be a
substring of its replacement.** The new marker text shares the prefix
`"Content truncated for Kafka transport"` with the old one, so grepping that
prefix returns >=1 on a correct deploy and proves nothing. The discriminating
form is the full old sentence including `- full version in S3`, which the new
code cannot contain. (This is the mistake made in `bugs_closed/144`'s runbook —
published as a delete-marker, returned 1 on a correct deploy, and was adopted as
a positive control by `bugs_open/153` before it was caught.)

**Gotcha: `logs deploy/X` reads ONE pod of N, and web-scrape-adapter has 3
replicas on a 1-partition topic — so exactly one pod works and two are idle for
life.** Grep every replica, or `-l app=… --tail=-1` for logs:

```bash
for p in $(kubectl -n ai-persona-system get pods -l app=web-scrape-adapter -o name); do
  echo -n "$p: "; kubectl -n ai-persona-system exec $p -- sh -c \
    'strings /app/web-scrape-adapter | grep -c "the remainder was DISCARDED and no copy was stored"'
done
```

## Functional evidence after the roll

Better than a grep — the behaviour itself. Denominator FIRST; an error count over
zero attempts is not evidence (that is how `bugs_closed/062`'s watch read clean
for days):

```bash
kubectl -n ai-persona-system logs -l app=web-scrape-adapter --tail=-1 --since=1h | grep -c "Starting scrape"
kubectl -n ai-persona-system logs -l app=web-scrape-adapter --tail=-1 --since=1h | grep "Truncating large field"
# the new honest signal — a truncation with nothing stored now SAYS so:
kubectl -n ai-persona-system logs -l app=web-scrape-adapter --tail=-1 --since=1h | grep "Truncated content was DISCARDED"
# and the delivery outcomes:
kubectl -n ai-persona-system logs -l app=web-scrape-adapter --tail=-1 --since=1h | grep -E "Reply exceeded broker max message size|undeliverable"
```

**Gotcha: `stored_copy` is a structured field now, so prefer it to reading the
marker prose:** `... | grep "Truncating large field" | jq -r '.stored_copy'` — a
`false` there is the defect's fingerprint, and it used to be unobservable.

## Build

```bash
# commit FIRST — make build-* builds committed HEAD (and prints what it leaves out)
make build-web-scrape-adapter
make push-web-scrape-adapter    # check the actual target name in the makefile
```

**Gotcha: bump `IMAGE_TAG` (makefile ~line 16) or the node serves its stale
cached binary.** The makefile is shared and other sessions bump it too — re-read
it immediately before building rather than reusing a value from earlier in your
session.

## Firing a probe scrape (only if you need to reproduce end-to-end)

Two traps, both from the filing lane's `bugs_open/133` RUNBOOK section — worth
repeating because each costs a wasted cycle:

1. The adapter takes a `{"body":…,"headers":…}` envelope as the Kafka **value**
   and does not read Kafka message headers. A bare body is rejected at
   `adapter.go:199` and **committed**, so it vanishes with no retry.
2. **The reply topic must already exist.** If it does not, the produce fails and
   logs `Failed to produce response` — one of the strings you are testing for.
   You would manufacture the exact hit. Confirm with `kcat -L` first.
