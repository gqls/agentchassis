# HANDOFF 2026-07-28b (late evening) — continue here

Supersedes `HANDOFF_2026-07-28_continue_here.md` for **what to do next**. That file
is still the reference for how `100`/`101` were fixed and verified; this one is the
current state and the queue. Written for token load, not because anything stalled.

> **Read this first, then re-derive.** Everything below was true at ~21:45 BST on
> 2026-07-28 on a tree where several sessions commit concurrently. Figures go stale
> in hours. Every claim that can be re-run has its command attached — run it.

---

## 1. Where things stand, in one paragraph

`bugs_closed/101` is **CLOSED, live and council-APPROVED**. `bugs_open/100` is
**OPEN on one owed run** that only `vetcomparison` can trigger. The config-key
coverage ratchet moved **208 → 152 undeclared actions**, council-APPROVED
(`Council-Reviewed: 07cf67c6-…` on `ee8b9a9a3`), and is **live on chassis v1.0.1194
but not yet exercised by a single run**. Two new bugs were filed this session,
`bugs_open/133` and `bugs_open/134`, neither fixed and neither this lane's to fix.
An owner ruling was recorded (the crawl call). Nothing is owed to any council.

## 2. THE ONE THING TO CHECK FIRST — the ratchet is live and unexercised

`CheckConfig` opted **58 actions** into runtime unknown-config-key detection. It is
in the running binary (pod-grep, §17 of NOTES), but **0 of 13 orchestrations since
the roll touched any of them**, so the detector has never fired.

**Denominator first — this is the third time in one session a clean number turned
out to be an empty one.**

```bash
# 1. has anything actually exercised it yet?
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
WITH mine AS (SELECT jsonb_object_keys((SELECT jsonb_agg(x)->0 FROM (SELECT 1) s)) )
SELECT count(DISTINCT o.orchestration_id) AS runs_touching_opted_in
FROM orchestration_states o, jsonb_each(o.workflow_plan->'steps') AS e(k,v)
WHERE o.created_at > '2026-07-28 20:48:00+00'
  AND v->>'action' IN (
    -- paste from: go run ./cmd/config-key-audit --specs | jq -r 'to_entries[]|select(.value.opted_in)|.key'
    'scrape_web','create_rerender_items','load_page_record'  -- …58 total
  );"

# 2. only if (1) is non-zero is this meaningful:
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=-1 --since=6h \
  | grep "keys this action does not read"
```

**If a warning appears:** it names step, action and keys. It is **far more likely a
gap in that action's spec than a genuinely dead key** — the batch only included
actions whose spec already covered every live key, so a hit means a definition
changed or my classification was wrong for that one. Add the key to the spec if the
action reads it. It is **warn-only** and blocks nothing (`StrictConfig` is set by
nobody — grepped, not assumed).

**Do NOT "fix" a warning by adding the key to `ConfigKeys` without checking the
action reads it.** That is the `WRONG_CALLS.md` 2026-07-28 mistake: declaring a dead
key makes it *recognised*, silences the detector, and leaves the behaviour broken.

## 3. The queue, in priority order

### 3a. `bugs_open/100` — the only thing between it and closure
Blocked on `vetcomparison` restarting collection. **Their lane has been told** —
`vetcomparison/PLAN_2026-07-26_site_strength.md` §P1 and their `HANDOFF_2026-07-26`
now carry the notice, including the checked fact that all three relevant agents are
on `v1.0.1192+` (so their first crawl will not hard-fail against SQL 257's CHECK).

Re-check whether it is still blocked before doing anything else — one query:
```sql
SELECT source_url, source_type, raw_data ? 'source_url' AS llm_claimed, collected_at
FROM business_intel.data_observations ORDER BY collected_at DESC LIMIT 5;
```
Newest `collected_at` was still **2026-03-18** at 20:00Z. Closure needs `source_url`
non-empty **AND** `llm_claimed` still **false** — a populated column alone proves
nothing about where it came from. If provenance comes back empty, grep the chassis
log for `no fetch provenance available` (present in 1194, re-checked) rather than
guessing: the `data.url` shape is `[UNVERIFIED]`, traced through code, never observed.

### 3b. Finish the ratchet — 152 actions left, and they need READING not a flag
```bash
./scripts/audit-config-keys.sh                    # the gap
go run ./cmd/config-key-audit --specs             # who is one line away
```
The split, re-derivable by joining those two (was 56/34/85/30 before the batch):
- **spec already covers every live key AND is passed to `ExtractActionInputs`** →
  safe one-line `CheckConfig: true`, asserts nothing new. **All 56 of these are done.**
- **spec exists but misses live keys (34)** → read the action. Each missing key is
  either a spec gap or a genuinely dead key. **`bugs_open/134` came out of this pile.**
- **no spec at all (85)** → read the action, write the spec.
- **registers a spec but never passes it to the extractor (30)** → read the action;
  the "verified by construction" argument does **not** apply to these.

### 3c. Nothing else in this lane is owed
No council round outstanding. No trailer owed. The five standing docs are current
through NOTES §17.

## 4. What this session changed, so you do not redo it

| # | what | state |
|---|---|---|
| `f5888c912` | council round 3 APPROVED, trailer claimed | done (previous session) |
| `584cdd516` | **OWNER RULING**: the two agents stay on scrape, warning — recorded as a *decision*, not an omission | done |
| `50c8e50d9` | vetcomparison told their P1 blocker is lifted, in the file that branches on it | done |
| `2f96fa70c` | **`bugs_open/133`** filed + `016b` §9 + the watch command fixed in 4 docs | filed, **not fixed** |
| `ce9e28784` | `CheckConfig` + 56-action opt-in; 208→152 | live on 1194, **unexercised** |
| `d6ba8af91` | SCR-005 registered, SCR-002/3/4 added to the concept INDEX (they were missing), `WRONG_CALLS` entry | done |
| `c6f41c062` | **`bugs_open/134`** filed | filed, **not fixed** |
| `121031901` | `016b` §10 row for 134 + §9 *"~0% adoption measures the MECHANISM"* | done |
| `ee8b9a9a3` | folded the second binary into `config-key-audit --specs` after two council seats objected; **carries the trailer** | done |

## 5. The two bugs filed, neither this lane's to fix

**`bugs_open/133` — the single-scrape path truncates to an S3 copy it never wrote.**
`adapter.go:331-344` cuts content at 50KB and appends *"full version in S3"*; the
upload is guarded at `:313` by `if uploadResults && …`. **4 of 6 live single-URL
scrape steps** have it false/unset, including `vet-practice-verifier/scrape_website`.
MEASURED: `raw_html 53805 → 50000`, zero S3-upload lines, produced successfully.
Second defect same function: `sendSuccessResponse` logs a produce failure and
returns — no deliverable error, the gap `bugs_closed/062` left when it fixed only the
batch sibling.
**LANDMINE: it COMPLETES SUCCESSFULLY, so no error-based watch will ever see it.**
Watch `grep "Truncating large field"` instead.

**`bugs_open/134` — `"category?"` is doc notation, not a key name.** Two inert keys on
`product-spec-refresher`; origin is seed 156, whose own line-15 comment uses `?` as
"optional" notation. Latent: 0 runs ever. Fleet-swept — only instance.

## 6. Landmines this session paid for

- **`kubectl logs deploy/<x>` reads ONE POD OF N.** 3 replicas + 1 partition + 1
  consumer group ⇒ two pods idle for life, and `logs deploy/…` may pick one, giving a
  permanently clean log. It did. Use `-l app=<x> --tail=-1`. Fixed in 4 docs.
- **Count the denominator before believing any clean number.** Three instances in one
  session: the 062 watch (0 errors / 0 attempts), `UNKNOWN KEYS: none` that a
  declaration had silenced, and the ratchet's 0 warnings over 0 runs.
- **Firing a probe scrape:** the adapter takes a `{"body":…,"headers":…}` **envelope
  as the Kafka value** and ignores Kafka headers — a bare body is rejected at
  `adapter.go:199` and **committed**, so it vanishes. And **the reply topic must
  already exist**, or the produce failure logs `Failed to produce response`, which is
  one of the two strings the 062 watch greps — you would manufacture your own hit.
- **The shared tree does not always compile.** `go test ./platform/orchestration/actions/`
  was failing on another session's uncommitted `spawn_actions.go`. Test against
  `git archive HEAD` + only your files before believing you broke something.
- **An APPROVED verdict is where an advisory is easiest to ignore.** Two seats
  independently said the new binary should have been a flag. They were right.

## 7. Where everything is

| what | where |
|---|---|
| the five standing docs | this directory — NOTES §13–§17 are this session |
| what a non-specialist should read | `README_where_we_are.md`, last entry |
| the closed bug | `bugs_closed/101_…` (residual 2 carries the owner ruling) |
| the open bugs | `bugs_open/100_…`, `bugs_open/133_…`, `bugs_open/134_…` |
| transferable patterns | `016b` §9 ×2 this session; §10 rows for 133 and 134 |
| how we were wrong | `WRONG_CALLS.md` — "adoption is slow" was a diagnosis I never checked |
| callable mechanisms | concept register `adopting-and-scraping.md` SCR-002…005 (+ now in the INDEX) |
