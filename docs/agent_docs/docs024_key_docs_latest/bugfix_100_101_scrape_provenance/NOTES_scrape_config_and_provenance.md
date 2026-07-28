# NOTES — `bugs_open/100` + `101` (scrape config + provenance)

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

---

## §1 — 2026-07-28 ~17:30 BST — coverage check and re-grounding, before any code

**Coverage.** `who-owns.py` over every plausibly-free bug in `/bugs_open/`. 100 → "no
owning workstream identified". 101 → names `vetcomparison`, but that thread is the
blocked party (its own line: "crawl restart BLOCKED on a Go change") and has written no
fix; last touch 2026-07-27.

Open work items touching this area — **zero rows**:

```sql
SELECT id, item_type, status, left(summary,90), created_at FROM site_work_items
WHERE status NOT IN ('complete','cancelled','rejected')
  AND (summary ILIKE '%scrape%' OR summary ILIKE '%provenance%' OR summary ILIKE '%config key%'
       OR summary ILIKE '%data_observation%' OR summary ILIKE '%only_main%');
-- (0 rows)
```

**101 re-grounded — unchanged.** Both definitions still carry the inert keys:

```
            type            |      step      | max_pages | follow_links | extract_mode | fallback_url
----------------------------+----------------+-----------+--------------+--------------+--------------
 domain-research-classifier | scrape_site    | t         | t            | t            | f
 vet-practice-verifier      | scrape_website | t         | t            | t            | t
```

**100 re-grounded — unchanged.** Still 2,970 rows, still zero provenance, still no key
in `raw_data`, collection still off since March:

```
 total | has_source_url | llm_claimed |           newest
-------+----------------+-------------+----------------------------
  2970 |              0 |           0 | 2026-03-18 22:09:03.579088
```

Note `llm_claimed = 0` is doing real work here: it is the *negative control* for the fix
(§How to verify in 100). It must still read 0 afterwards.

## §2 — the `[UNSETTLED]` box is settled, and the answer is a second bug

101 refuses to let candidate 2 proceed until this is answered: does production's
Firecrawl path strip nav/footers? If yes, more page fetches deliver nothing, because
company numbers live in footers.

Read the code rather than inferring from stored samples (the file's own evidence was
ambiguous — 75% of 2,452 samples retain footer nav text, but 0 contain
company-registration text, and those two facts do not separate the hypotheses).

`internal/adapters/webscrape/providers/firecrawl.go:77-111`:

```go
onlyMainContent := false
if mainContent, ok := config["only_main_content"].(bool); ok {
    onlyMainContent = mainContent
}
...
if onlyMainContent {                       // <-- only ever emits TRUE
    payload["onlyMainContent"] = onlyMainContent
}
```

So the key is emitted **only when true**. `false` is indistinguishable from unset, and
unset means Firecrawl's own default applies.

**The external premise was verified, not assumed** — Firecrawl `/scrape` API reference:
*"default: true"*, and it *"excludes headers, navs, footers"*. So a caller explicitly
asking for the full page gets main-content-only: the exact opposite.

The `/crawl` path in the **same file** (line 338) is correct:

```go
if onlyMain, ok := config["only_main_content"].(bool); ok {
    scrapeOptions["onlyMainContent"] = onlyMain
}
```

Two paths, one file, opposite semantics for the same key. This is 101's own class —
config that reads as live and is not — in a provider every scrape on the fleet goes
through.

**Consequence for 101's measured numbers:** the 22%→30% company-number figure was
measured by a **read-only probe fetching raw HTML**, not through production's `/scrape`.
Production has been receiving main-content-only pages, so the probe's 30% was never
reachable by adding page fetches alone. `[UNVERIFIED]` — how much of the gap this
accounts for is not measured here and should not be asserted; what is established is
that the probe and production were not fetching the same thing.

## §3 — the survey that sized the framework fix (and nearly oversized it)

101's candidate 1 is "reject unknown config keys for registered actions". Before
designing it, measured the surface it would have to be right about:

```sql
WITH steps AS (
  SELECT ad.type, e.k AS step, v->>'action' AS action, v->'config' AS cfg
  FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') AS e(k,v)
  WHERE ad.deleted_at IS NULL AND COALESCE(ad.is_snapshot,false)=false AND ad.is_active
    AND v->'config' IS NOT NULL AND jsonb_typeof(v->'config')='object'
)
SELECT count(*) AS total_steps_with_config, count(DISTINCT action) AS distinct_actions,
       (SELECT count(*) FROM (SELECT DISTINCT s.action, ck.key FROM steps s, jsonb_object_keys(s.cfg) AS ck(key)) x)
FROM steps;
--  1155 | 228 | 811
```

**811 distinct (action, key) pairs over 228 actions.** This killed the version of the
plan I was about to write, which would have hard-rejected unknown keys fleet-wide: a
declaration that wrong at that scale would reject live definitions, and an over-strict
validator is a considerably worse bug than the one being fixed (the same shape as 101's
own recorded trap — a fleet-wide `WHERE config ? 'max_pages'` cleanup would strip a live
page cap off `build-site-planner`). Opt-in per action it is.

**Reuse check before building.** `datahelpers.RegisterActionInputSpec` already exists and
134 files call it. `GetActionInputSpec` has **no callers at all** — grep returns only its
own definition. The registry is populated and read by nothing but `registry_parity_test.go`.
So the machinery to extend is already there and currently inert; building a second one
would have been the drift class this repo's council reviews for.

## §4 — 2026-07-28 ~18:00 BST — the audit found a FIFTH inert key on its first run

Not from reading the bug file. `scripts/audit-config-keys.sh`, run against the live
fleet immediately after declaring `scrape_web`'s contract, reported:

```
=== UNKNOWN KEYS (action declared its contract; these are not in it) ===
  scrape_web: add_protocol
```

`bugs_open/101` names **four** keys. This is a fifth, and it is a near-miss typo
rather than an aspiration — which is why no amount of re-reading the file would
have surfaced it:

```
$ grep -rn --include=*.go "add_protocol" . | grep -v '^./docs/'
platform/orchestration/actions/webscrape_actions.go:509:  if addProtocol, ok := config["add_protocol_if_missing"].(bool); ok && addProtocol {
```

The code reads `add_protocol_if_missing`; the definition writes `add_protocol`;
and line 509 belongs to a **different action** (the URL-validation action), not to
`scrape_web`. Live:

```
            type            |    step     | add_protocol
----------------------------+-------------+--------------
 domain-research-classifier | scrape_site | true
```

A bare domain reaching the adapter is a failed fetch, so this was not cosmetic.
Implemented (fires only when explicitly true AND the URL has no scheme).

**What this says about the fix, and it is the part worth keeping:** the detector
earned its place in one run, on a key class the human process structurally could
not catch. Four keys were found by a careful reader; the fifth needed a machine
that compares what the config says against what the binary declares. The value is
not the four keys — it is that the next typo surfaces without anyone going
looking.

## §5 — the shape I could not observe, and what I did instead

`ExtractFetchProvenance` needs to know what `collected_data["scraped_data"]`
actually looks like. Tried to sample it:

```sql
SELECT ... FROM orchestration_states WHERE collected_data ? 'scraped_data' ...;
-- (0 rows)
```

**Nothing survives.** Vet collection has been off since 2026-03-18 and
`orchestration_states` is on a retention clock (the recorded landmine: *every
history table is on a retention clock — record a RATE not a count*). So the shape
could not be measured, and guessing it would decide whether every future
observation is sourced or silently not.

Traced it through the code instead, which is weaker than an observation and
stronger than a guess:

```
adapter sendSuccessResponse    → body: {success, body:{data: <provider result>}, …}
types.ResponseBody.Body        → {data: <provider result>}
coordinator.parseResponseBody  → CleanDataMap of that
applyResponseToState           → collected_data[output_field] = {data:{url, captured_at, …}}
```

So the live path is **`data.url`**. Recorded as such in the code, with the
derivation, and marked `[UNVERIFIED]` for which OTHER shapes occur in practice —
the reader accepts six because the chain has several unwrap points
(`output_mapping` short-circuits it entirely) and being wrong here means silently
storing an unsourced row.

## §6 — what was committed, and what is still owed

- `2ebabf2ca` — the fix (14 files). **INERT until two images roll**: the
  web-scrape-adapter for `only_main_content`, the chassis for everything else.
- `70885daf0` — a declared gofmt sweep plus a doc-comment correction. The commit
  hook caught that `business_intel_actions.go` was not gofmt-clean; I had
  deliberately left that pre-existing drift alone to keep a reformat out of a
  bugfix diff, and the hook's note that **the build gate REJECTS un-gofmt'd code**
  made that the wrong call — it would have failed the build for everyone.
  Swept separately, said so in the message.
  > **MISSTEP, recorded:** "leave pre-existing drift alone" is right for review
  > hygiene and wrong when the drift is in a file you are committing and the gate
  > is a hard one. The check that would have caught it in seconds is
  > `gofmt -l <the files you are about to commit>` — not `gofmt -l <package>`,
  > which is what I ran, and which buried my file in ten others I had not touched.
- Council submitted, `SUBMISSION_CORR=f4cf0aab-5a08-4475-91ea-fa831cff323c`.

Still owed: the council verdict; both images; SQL 257 applied AFTER the chassis
image (order load-bearing — before it, the constraint refuses writes the running
binary cannot yet satisfy).
