# RUNBOOK — bugs_open/158 (reply-drop sizing and the silent-drop detector)

## 1. The Kafka limit — and the command in the ticket CANNOT answer it

**There is no Kafka broker in `ai-persona-system`.** The cluster is in namespace
`kafka`: `personae-kafka-cluster-combined-pool-prod-{0,1,2}` (CR
`personae-kafka-cluster`). A `kubectl exec` at a pod that does not exist, with
stderr discarded, prints **nothing** — which reads exactly like "no override set".

```bash
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
  /opt/kafka/bin/kafka-configs.sh --bootstrap-server localhost:9092 \
  --entity-type topics --entity-name system.agent.generic.responses --describe
# -> "Dynamic configs for topic ... are:" with nothing under it = a REAL empty

kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
  /opt/kafka/bin/kafka-configs.sh --bootstrap-server localhost:9092 \
  --entity-type brokers --entity-name 0 --describe --all | grep '^\s*message\.max\.bytes'
# -> message.max.bytes=1048588 synonyms={DEFAULT_CONFIG:message.max.bytes=1048588}
```

**Gotcha:** never `2>/dev/null` a probe whose *empty* result is the finding. The
header line ("Dynamic configs for topic … are:") is the proof the command ran; its
absence is the proof it did not.

## 2. Exposure — is any reply near 1MB?

```sql
SELECT max(length(response_text)) AS fleet_max, count(*) AS calls,
       count(*) FILTER (WHERE length(response_text) > 524288)  AS over_half_mb,
       count(*) FILTER (WHERE length(response_text) > 1048588) AS over_limit
FROM llm_call_log;
-- 2026-08-03: 48,327 | 47,577 | 0 | 0
```

**Gotcha:** do not filter by `agent_type ILIKE '%reasoning%'` etc. first — those
labels **do not exist** in this table (`agent_type` was relabelled 2026-07-26, and
these four services do not log here at all). Filtering by guessed names returns
0 rows, which reads as "no data" rather than "wrong question". **Enumerate the
labels before filtering on them:** `SELECT agent_type, count(*) FROM llm_call_log
GROUP BY 1 ORDER BY 2 DESC;`

## 3. The scrape-step / `upload_results` survey (item 0's handover)

**Gotcha:** the action names are not the obvious ones. `scrape_website`,
`firecrawl_crawl` and `web_scrape` as a guessed set returns 4 rows; the real
vocabulary is `scrape_web`, `batch_webscrape`, `firecrawl_scrape`,
`firecrawl_extract`, `fetch_scrape`, `med_scrape_prices`, `format_crawl_for_analysis`
(13 distinct). **Match on the pattern, do not enumerate from memory:**

```sql
SELECT ad.type AS agent, e.k AS step, e.v->>'action' AS action,
       COALESCE(e.v->'config'->>'upload_results','(unset)') AS upload_results
FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') AS e(k,v)
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND (e.v->>'action' ILIKE '%scrape%' OR e.v->>'action' ILIKE '%crawl%')
ORDER BY 4 DESC, 1, 2;
```

2026-08-03: **22 steps — 3 true, 1 false, 18 unset**, including
`vet-practice-verifier/scrape_website` (the one whose *purpose* is extraction from
a long page). The 18 are the handover for item 0; none was changed by this session.

## 4. The detector, and how to re-prove it

```bash
python3 - <<'PY'
import importlib.util, subprocess
spec = importlib.util.spec_from_file_location("pc", "scripts/pattern-check.py")
pc = importlib.util.module_from_spec(spec); spec.loader.exec_module(pc)
allgo=[f for f in subprocess.run(["git","ls-files","*.go"],capture_output=True,text=True).stdout.split()
       if not f.endswith("_test.go")]
out=[]; pc.check_silent_reply_drop(allgo, None, out)
print(len(allgo), "files ->", len(out), "findings")
for k,w,what,_ in out: print("  ", w)
PY
```

Expected 2026-08-03: **8 findings over 777 files** — thunder:391, websearch:614,
contentcreator:763, reasoning:290, agentbase:887, processor:2000, coordinator:3663,
helpers:427.

**Gotcha, and it is the one that matters:** the known-bad control (does it fire on
the 4 sites the ticket names?) **cannot** find false positives, because it only ever
looks at files that should fire. Only the fleet sweep can, and it found three:
`return producer.Produce(...)` *propagates* the error rather than swallowing it, and
two produces to **request** topics matched because the word "response" sat within
the ±12-line window. Run **both** controls, always, and gate the reply test on the
call's own arguments rather than on surrounding lines.
