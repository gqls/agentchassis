# RUNBOOK — `bugs_open/100` + `101`

Commands that were hard to get right, with the gotcha attached.

## Re-ground 101 (are the keys still inert in live definitions?)

```sql
SELECT ad.type, e.k AS step,
       (v->'config' ? 'max_pages') AS max_pages, (v->'config' ? 'follow_links') AS follow_links,
       (v->'config' ? 'extract_mode') AS extract_mode, (v->'config' ? 'fallback_url_field') AS fallback_url
FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') AS e(k,v)
WHERE ad.deleted_at IS NULL AND COALESCE(ad.is_snapshot,false)=false AND ad.is_active
  AND v->>'action' = 'scrape_web'
  AND (v->'config' ?| array['max_pages','follow_links','extract_mode','fallback_url_field']);
```

**Gotcha:** the three `deleted_at / is_snapshot / is_active` predicates are all load-bearing.
Without them the same query returns snapshot rows and reads as a much larger blast radius.

## Re-ground 100 — and the column that actually discriminates

```sql
SELECT count(*) AS total,
       count(*) FILTER (WHERE COALESCE(source_url,'')<>'') AS has_source_url,
       count(*) FILTER (WHERE raw_data ? 'source_url')     AS llm_claimed,
       max(collected_at) AS newest
FROM business_intel.data_observations;
```

**Gotcha:** `has_source_url` going non-zero is **not** the pass condition. `llm_claimed`
must stay **0**. If both rise, the fix is the rejected candidate 4 (ask the model for its
own provenance) and must be reverted — a populated column proves the column was written,
never *by what*.

## Is a config key actually read by any Go code?

```bash
grep -rn --include=*.go "<key>" . | grep -v '^./docs/'
```

**Gotcha — this is the recorded landmine, do not narrow it.** 101's own diagnostic was
`grep -rn "max_pages" ... | grep -i webscrape`, which is true about `webscrape` and false
about the fleet: `max_pages` is live and load-bearing for `select_representative_content`
and `validate_site_plan`. Run it unfiltered, then decide per (action, key) — never per key.
`docs/` is excluded because it holds vendored copies.

## Size the config surface before promising any fleet-wide validation

```sql
WITH steps AS (
  SELECT e.k AS step, v->>'action' AS action, v->'config' AS cfg
  FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') AS e(k,v)
  WHERE ad.deleted_at IS NULL AND COALESCE(ad.is_snapshot,false)=false AND ad.is_active
    AND v->'config' IS NOT NULL AND jsonb_typeof(v->'config')='object'
)
SELECT count(DISTINCT action) AS actions, count(*) AS steps,
       (SELECT count(*) FROM (SELECT DISTINCT s.action, ck.key FROM steps s, jsonb_object_keys(s.cfg) AS ck(key)) x) AS pairs
FROM steps;
```

**Gotcha:** `jsonb_typeof(v->'config')='object'` is required — some steps carry a *string*
config (a reference, not a literal; see the model-directory landmine), and
`jsonb_object_keys` errors on those rather than skipping them.

## Audit which live config keys no action declares

```bash
./scripts/audit-config-keys.sh          # undeclared (action,key) pairs, fleet-wide
```

Reads the live DB and compares against the declared allow-lists compiled into the binary
(`--json` for machine output). An (action, key) pair listed here is either a real inert key
or an action that has not opted in yet — the report does not distinguish, by design.

## Verify P1 without calling Firecrawl

```bash
go test ./internal/adapters/webscrape/providers/ -run TestScrapePayload -v
```

The assertion is on the **payload**, not on a live response: `only_main_content:false` must
produce a payload where `onlyMainContent` is **present and false**. Asserting on scraped
content instead would need Firecrawl and would still not distinguish "we sent false" from
"Firecrawl happened to keep the footer".

## Post-roll: confirm the chassis actually carries the fix (do this BEFORE SQL 257)

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$POD" -- sh -c \
  'echo -n "marker: ";           strings /app/agent-chassis | grep -c "unrecognised_keys";
   echo -n "positive control: "; strings /app/agent-chassis | grep -c "scrape_web"'
```

Pass = marker ≥1 **and** control ≥1. **Measured 2026-07-28 pre-deploy: marker 0,
control 1** — which is what makes `unrecognised_keys` a discriminating marker rather
than a vacuous one, and gives the check a known before-state to flip.

**Gotcha — never confirm the roll by the image tag.** A same-tag rebuild ships the
node's stale cached binary, and there was a live same-tag collision at v1.0.1186. At
the time of writing `IMAGE_TAG` in the makefile equals the tag already deployed, so a
build without bumping it produces exactly that. Applying SQL 257 off a tag signal
against a stale binary turns a silent data-quality defect into a **hard outage of vet
verification**, because the constraint refuses writes the running code cannot satisfy.

## Post-roll: the bugs_closed/062 payload-size watch (web-scrape-adapter)

The `only_main_content` fix makes three steps receive the **full page** they always
asked for, so their scrape responses grow. `bugs_closed/062` was a Kafka
*Message Size Too Large* failure whose root cause #1 is in this same provider file.

> **CORRECTED 2026-07-28 ~20:30 — `deploy/web-scrape-adapter` reads ONE POD OF THREE,
> and every command in this section had it wrong.** The deployment runs **3 replicas**
> in one consumer group on a **1-partition** topic, so exactly one pod consumes and
> the other two are idle for their entire life. `kubectl logs deploy/…` picks a pod
> arbitrarily — on 2026-07-28 it picked `d8h2w`, an idle one, whose log is
> **permanently clean no matter what the working pod is doing**. Use
> `-l app=web-scrape-adapter --tail=-1`, which reads all three. Filed as part of
> `bugs_open/133`. Caught by firing a probe scrape and finding the evidence in a pod
> the documented command does not read.

```bash
kubectl -n ai-persona-system logs -l app=web-scrape-adapter --tail=-1 --since=2h \
  | grep -i "Message Size Too Large\|Failed to produce"
```

Exposure, worst first (`formats` as configured live):

| step | formats | after the fix |
|---|---|---|
| `site-scraper/scrape_site` | *(none — 4-format default + screenshot)* | **grows most** |
| `website-capture-firecrawl/scrape_main_page` | `["markdown","html","screenshot"]` | grows |
| `site-adoption-agent/fetch_primary_css` | `["rawHtml"]` | unchanged — `rawHtml` is the raw page, which `onlyMainContent` does not filter |

**FIRST, ask whether ANY scrape has run — otherwise a clean log is not evidence:**

```bash
kubectl -n ai-persona-system logs -l app=web-scrape-adapter --tail=-1 --since=6h | grep -c "Starting scrape"
```

Measured 2026-07-28 ~19:00, ~40 min after the roll: **0**. Zero scrapes attempted,
therefore zero errors — the clean 062 result above was **uninformative, not
reassuring**. This one extra count is the difference between "the payload change is
fine" and "nothing has exercised it yet", and the two are indistinguishable in the
error grep alone. Re-run the watch only once this returns non-zero.

**EXERCISED 2026-07-28 ~19:35 — the denominator is no longer zero.** One probe scrape
of `https://vetcomparison.uk` (corr `1e97bd22`) through the adapter: **1 attempt, 0
`Message Size Too Large`, 0 `Failed to produce`, produced successfully.** So the 062
risk did not materialise on this page. **It also found a defect that no error-based
watch can see** — the response was only deliverable because the adapter had silently
truncated `raw_html` from 53,805 to 50,000 chars and appended *"full version in S3"*
when `upload_results` was false and **nothing had been uploaded**. `bugs_open/133`.

**Firing a probe scrape** — two traps, both of which cost this session a round:

1. The adapter takes a `{"body":…,"headers":…}` **envelope as the Kafka value** and
   ignores Kafka message headers. A bare body is rejected at `adapter.go:199`
   (*"Invalid message format - missing headers or body"*) and **committed** — it
   vanishes with no retry, and the rejection is only visible in the consuming pod.
2. **The reply topic must already exist.** If it does not, the produce fails and logs
   `Failed to produce response` — one of the two strings this watch greps. You would
   manufacture the exact hit you are testing for. Seed the probe topic and confirm it
   with `kcat -L` **before** firing.

```bash
# the guard firing IS the signal — it succeeds, so no error grep will ever show it
kubectl -n ai-persona-system logs -l app=web-scrape-adapter --tail=-1 --since=1h \
  | grep "Truncating large field"
```

**Gotcha:** the 062 failure is SILENT to the caller — the adapter logs the produce
error and the orchestration then starves through ~12 minutes of timeout retries, each
re-scraping successfully and re-failing identically. So the absence of a workflow error
is not evidence this is fine; grep the ADAPTER log. Mitigation if it fires is
config-only and needs no roll: set `scrape_config.formats` on the offending step (the
override exists precisely because of 062).

## Post-roll: is `CheckConfig` live, and has it been exercised? (TWO questions)

Added 2026-07-28 after `v1.0.1194`. **They are separate and the second is the one
that gets skipped.**

**1. Is it in the running binary?** Never the tag — a retag is not a rebuild.

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$POD" -- sh -c '
  echo -n "checksConfig (created by the change): "; strings /app/agent-chassis | grep -c "checksConfig"
  echo -n "POSITIVE CONTROL UnknownConfigKeys  : "; strings /app/agent-chassis | grep -c "UnknownConfigKeys"
  echo -n "NEGATIVE CONTROL bogus_symbol_xyz   : "; strings /app/agent-chassis | grep -c "bogus_symbol_xyz"'
```
On 1194: 2 / 5 / 0. `checksConfig` is discriminating because the change created it.

**2. Has it actually run? COUNT THIS BEFORE READING ANY WARNING COUNT.**

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT count(DISTINCT o.orchestration_id) AS runs_touching_opted_in
FROM orchestration_states o, jsonb_each(o.workflow_plan->'steps') AS e(k,v)
WHERE o.created_at > '<roll time>' AND v->>'action' IN (<the 58>);"
# list them with: go run ./cmd/config-key-audit --specs | jq -r 'to_entries[]|select(.value.opted_in)|.key'
```
Measured 21:45Z on 1194: **13 orchestrations since the roll, 0 touching an opted-in
action.** So the warning count below was 0 over a denominator of 0 — unfalsifiable,
not reassuring.

**3. Only then, the warning itself — across ALL pods:**

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=-1 --since=6h \
  | grep "keys this action does not read"
```
A hit names step, action and keys. It is warn-only (`StrictConfig` is set by nobody —
grepped, not assumed) and is **more likely a spec gap than a dead key**, because the
batch only included actions whose spec already covered every live key. Add the key to
the spec if the action reads it; **do not add it to `ConfigKeys` without checking**,
which silences the detector and leaves the behaviour broken.
