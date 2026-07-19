# relojistas.com — Rebuild Runbook (exact commands)

Operator commands for the rebuild. Box: **Hetzner CPX22, 167.233.33.159** (nbg1,
CF-proxied), hostname `ubuntu-4gb-nbg1-1-relojistas`. Companion to
`relojistas_rebuild_plan.md` / `_manifest.md` / `_running_notes.md`.
Section 0 is proven (run 2026-07-15). Sections 1–8 are forward-looking; treat the
chassis/DB steps as templates to fill against live schema at build time.

---

## 0. Reproduce the traffic read (proven)

```bash
BOX=167.233.33.159
# Log inventory + volume + status mix
ssh root@$BOX 'ls -la /var/log/nginx/ | grep -i access; wc -l < /var/log/nginx/access.log;
  awk "{print \$9}" /var/log/nginx/access.log | sort | uniq -c | sort -rn | head'
# Engine store: captured intent events + counters
ssh root@$BOX 'cat /var/lib/site-engine/events-*.jsonl; cat /var/lib/site-engine/counters.json'
# RSS feed: hit count, per-day, who pulls it, status
ssh root@$BOX 'grep -c "type=RSS2" /var/log/nginx/access.log;
  grep "type=RSS2" /var/log/nginx/access.log | awk "{print \$4}" | cut -d: -f1 | tr -d "[" | sort | uniq -c;
  grep "type=RSS2" /var/log/nginx/access.log | awk -F\" "{print \$6}" | sort | uniq -c | sort -rn | head;
  grep "type=RSS2" /var/log/nginx/access.log | awk "{print \$9}" | sort | uniq -c'
# Which legacy feeds are subscribed (normalise session id out)
ssh root@$BOX 'grep -oE "/external.php\?[^ ]*type=RSS2[^ ]*" /var/log/nginx/access.log |
  sed -E "s/s=[a-f0-9]{16,}&?//; s/&?\$//" | sort | uniq -c | sort -rn | head -25'
# Top organic terms (once volume grows)
ssh root@$BOX 'cat /var/lib/site-engine/events-*.jsonl | jq -r "select(.host==\"relojistas.com\").value" |
  tr "A-ZÁÉÍÓÚÜÑ" "a-záéíóúüñ" | sort | uniq -c | sort -rn | head -25'
```
Reference result (2026-07-15): 405,701 reqs / 5.5d, 83.5% 404, 0.34% 200; RSS 749 pulls
(~136/day, all 404/301); 9 events (~8 organic ES/CL/MX).

## P0. CF real-ip prerequisite (accurate logs/country under proxy)

```bash
# On the box, re-run the idempotent provisioner with Cloudflare real-ip enabled.
# (Installs cloudflare-realip.conf so nginx logs the true client IP and CF-IPCountry
#  reaches the engine.) Confirm exact flag name against the deployed setup.sh first:
ssh root@$BOX 'grep -n "CLOUDFLARE\|real_ip\|CF-Connecting" /tmp/setup.sh /etc/nginx/ -r'
CLOUDFLARE=true DOMAINS="relojistas.com" WWW_ALIAS=true \
  LETSENCRYPT_EMAIL=you@your-real-domain.tld DEPLOY_USER=deploy bash /tmp/setup.sh
# Verify:
ssh root@$BOX 'tail -5 /var/log/nginx/access.log'   # real IPs, not 172.x CF ranges
```

## P1. Site record + classification — artifact: `relojistas_rebuild_seed.sql`

Concrete, schema-verified SQL is in **`relojistas_rebuild_seed.sql`** (P1a onboarding
UPDATE + P1b `relojistas_set_news_feed()`). Apply order:
```bash
# 0. Confirm the site row exists (create via the normal path first if not):
psql -c "SELECT id, domain, status, github_repo FROM sites WHERE domain='relojistas.com';"
# 1. Load the functions + run the onboarding UPDATE (edit <INTERNAL_API_KEY> first —
#    read it from the box: /etc/site-engine/site-engine.env, never echo $KEY):
psql -f relojistas_rebuild_seed.sql
# 2. Set the news_feed recommendation on the classification spec:
psql -c "SELECT relojistas_set_news_feed('<site-uuid>');"
```
`content_features.news_feed.recommended=true` is what the 6-hourly `content-feed-trigger`
keys on; `source_types=[rss,api_news,news_search]` drives the auto-seeder (rss skipped by
it — P2 inserts those explicitly).

## P2. Seed news sources + first ingest

```sql
-- content_sources rows (types: api_news / rss / news_search). api_news = fabrication-safe.
--
-- (1) api_news PRIMARY — Grok-4-1-fast Responses API (web_search + x_search), prompt:
--   "Noticias de relojería en español, últimas 24-72h: novedades de marcas, ferias,
--    subastas, lanzamientos, reparación. Devuelve título, URL de la fuente, resumen breve."
--   Add a SECOND api_news row for Gemini as an alternate provider later (operator request).
--
-- (2) rss SUPPLEMENTS — VERIFIED live + on-vertical Spanish (checked 2026-07-15; re-check
--     liveness at seed time). Diversity: 5 independent magazines + Grok = ideal spread.
--       https://www.debajodelreloj.com/feed/   (Debajo del Reloj)   -- updated daily
--       https://tiempoderelojes.com/feed/       (Tiempo de Relojes)
--       https://trmagazine.es/feed/             (TR Magazine)
--       https://www.maquinasdeltiempo.com/feed/ (Máquinas del Tiempo)
--       https://relojesyestilo.es/feed/         (Relojes y Estilo)
--   Optional intl (English; needs translation or an "internacional" category, v2):
--       https://monochrome-watches.com/feed/    (Monochrome Watches)
--   REJECTED (do NOT seed): elcronometro.com/feed (stale 2025), relojesmania.com/feed
--     (stale 2024-25, smartwatch/off-vertical), horalatina.com/feed (404 — no feed there).
--
-- Concrete seed = seed_relojistas_sources() in relojistas_rebuild_seed.sql (P2):
--   SELECT seed_relojistas_sources('<site-uuid>');   -- 5 verified RSS + Spanish Grok
```
```bash
# Kick one pipeline pass, then confirm items landed + were triaged:
# (trigger content-feed-orchestrator for this site, or wait for the 6h heartbeat)
# Verify:
#   SELECT status, count(*) FROM content_feed_items WHERE site_id=... GROUP BY status;
#   expect ingested>0 then relevant>0 after the next triage pass (two-pass by design).
```

## P3. Build + deploy the static site (chassis → vm-sites)

```bash
# Run the normal build pipeline (planner → content → design → assemble → deploy) for the
# site. git_deployer resolves 'vm-sites' via resolveGitRepoName from sites.github_repo.
# Pages: home (hero + latest-news + search + category tiles), /noticias, evergreen ES pages.
# Deploy lands in the vm-sites repo → the vm-sites Action rsyncs to
#   /var/www/vm-sites/relojistas.com/ on the box.
# Verify:
curl -sS https://relojistas.com/ | grep -i "noticias\|buscar"      # Spanish page live
curl -sS https://relojistas.com/data/latest-news.json | jq '.items | length'
```

## P4. Outbound RSS (`/feed.xml`) — net-new render step

```text
Build render_rss_feed (thin variant of render_news_section_action.go):
  - read top ~30 curated content_feed_items for the site
  - emit RSS 2.0 XML (channel es-ES; <link>/atom:self = legacy URL; item <link> = source URL)
  - git_commit to /feed.xml (files_field, not content_field — see the JS-assets deploy bug)
Add nginx location for the static file:
  location = /feed.xml { add_header Content-Type application/rss+xml; }
```
```bash
curl -sS https://relojistas.com/feed.xml | head -20         # valid <rss version="2.0">
xmllint --noout <(curl -sS https://relojistas.com/feed.xml) && echo RSS-WELLFORMED
```

## P5. Legacy-URL handler + clever engine (D2)

```text
Engine (site-engine, stdlib Go) — add a handler for GET /external.php:
  - type=RSS2, no board id            -> serve master feed (same bytes as /feed.xml)
  - type=RSS2 & forumids=N / cat=N    -> serve mapped category feed (v1: master-to-all)
  - Content-Type application/rss+xml; strong caching; ETag/Last-Modified
Option B (search-that-answers): on /intent submit, keep recording the event AND return
  results matched against /data/latest-news.json + a small watch reference set (link out).
Deploy engine via the site-engine Action (build amd64 -> ship -> sudo-hook swap).
```
```bash
# nginx: route the legacy path to the engine (beside /stats, /events, /intent).
#   location = /external.php { proxy_pass http://127.0.0.1:<engine_port>; ... }
# Re-run idempotent setup.sh to install the location, then:
curl -sS 'https://relojistas.com/external.php?type=RSS2'            # 200 valid RSS
curl -sS 'https://relojistas.com/external.php?type=RSS2&forumids=44'# 200 (mapped or master)
ssh root@$BOX 'journalctl -u site-engine -n 20 --no-pager'
```

## P6. Verify reactivation

```bash
# Legacy feed now returns 200 to subscribers (was 100% 404):
ssh root@$BOX 'grep "type=RSS2" /var/log/nginx/access.log | awk "{print \$9}" | sort | uniq -c'
# Real feed reader: subscribe to https://relojistas.com/external.php?type=RSS2 and confirm render.
# Intent still captured after the search upgrade:
ssh root@$BOX 'tail -5 /var/lib/site-engine/events-*.jsonl'
# Stats:
curl -sS -H "X-Internal-Key: <KEY>" https://relojistas.com/stats
```

## Guardrails
- Never seed fabricated RSS URLs (triage rejects them; also our credibility). Prefer api_news.
- Link-out + summary only; no full-text republication.
- git_commit component/data files via **files_field** (the HTML-only content_field bug drops JS).
- setup.sh is idempotent + has box-takeover semantics — only ever run it against this box.
- Confirm exact env/flag names against the **deployed** setup.sh, not the repo copy (stale-copy trap, debug guide #26).

---

## P7 — content pages (Guías + Glosario). Commands that had to be right.

### Dispatch any agent at a site (the ONE pattern; kcat)

Source: `FOCUS_dispatch_diagnostic(4).md:130-159`. Works for
`completeness-discovery-agent`, `build-dispatch-loop`, `design-audit-agent`, …

```bash
set -u
AGENT_TYPE="build-dispatch-loop"           # or completeness-discovery-agent, …
SITE_ID="ecf15e75-a966-4900-bcb0-1c85f689dbfd"; DOMAIN="relojistas.com"
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
kubectl -n kafka run -i --rm "kcat-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID -H orchestration_id=$ORCHESTRATION_ID \
  -H request_id=$REQUEST_ID -H message_id=$MESSAGE_ID \
  -H message_type=request -H client_id=demo_client \
  -H action=orchestrate -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"${AGENT_TYPE}"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON
```
**Gotchas.** Print `$ORCHESTRATION_ID` before firing — a kcat dispatch can vanish with no
orchestration row; if nothing moves, refire. Don't dispatch within ~300s of a chassis pod
restart (`kubectl -n ai-persona-system get pods | grep agent-chassis` — check AGE).
`orchestration_states` keys on **`orchestration_id`**, not `id`.

### Do NOT use build-site-planner to add pages

`reconcile_site_plan` lives ONLY inside `build-site-planner`, and running that agent is the
`bugs_open/001` clobber. To add pages without it, write the work item reconcile *would* have
written and let `build-dispatch-loop` pull it. Shape copied from the proven `/noticias` row:

```sql
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, priority,
   handler_agent, status, created_by, item_key, pipeline)
VALUES ('<site>', 'reconcile_site_plan', 'needs_page', 'medium',
  'Build <page> page (not_built)',
  jsonb_build_object('reason','not_built','plan_id','<plan>',
                     'page_name','<page>','page_role','<role>'),
  50, 'page-build-handler', 'triaged', '<you>', 'needs_page:<page>', 'build');
```
- `status='triaged'` is what makes it dispatchable — the loop selects
  `status IN ('triaged','approved')` ordered by `priority ASC, created_at ASC`.
- **`item_key` collides on a partial unique index** `idx_swi_dedup (site_id,item_key)
  WHERE status NOT IN (complete,verified,rejected,wont_fix,failed,unresolved,cancelled)`.
  A stale non-terminal item with the same key blocks the insert — UPDATE it instead.
- **Order children BEFORE their index page.** `query.pages_where_type` filters on
  `pages.status IN ('active','deployed')`, **not** `build_status` — so an index built while
  its children are still `planned` renders a list of links to pages that do not exist.
  Give children a lower priority number and hold the index back until they deploy.

### The two queries that tell you where a build actually is

```sql
-- queue
SELECT status, count(*) FROM site_work_items
 WHERE site_id='<site>' AND item_type='needs_page' AND priority=50 GROUP BY status;
-- pages
SELECT build_status, count(*) FROM pages
 WHERE site_id='<site>' AND page_type IN ('guide','entity-page') GROUP BY build_status;
```
Wait on them with an `until` loop, not a fixed sleep — a 12-page batch is claimed a few at a
time and takes minutes.

### Claims fence (evidence_base) — apply, verify, and what it does NOT cover

```sql
SELECT jsonb_array_length(data->'facts')         AS facts,
       jsonb_array_length(data->'banned_claims') AS bans,
       length(data->>'writer_block')             AS block_chars
  FROM site_specs
 WHERE site_id='<site>' AND aspect='evidence_base' AND is_current;   -- 13 | 9 | 3453
```
- The column is **`data`**, not `specs` (the prompt path `site_specs.specs.evidence_base…`
  misleads — `specs` is the prompt's own keying by aspect).
- Confirm the fence is live in the **prod prompt**, never the repo copy:
  ```sql
  SELECT default_config::text ~ 'writer_block' FROM agent_definitions
   WHERE type='page-content-writer' AND deleted_at IS NULL;   -- expect t
  ```
- **Test banned_claims patterns by compiling them in Go RE2 before applying.** They are
  BLOCKER severity, and an invalid regex silently degrades to a literal substring rather
  than erroring. A throwaway `regexp.MustCompile` + match table over ~12 fabrication
  sentences and ~8 legitimate ones caught three real defects here that reading did not.
- **`ScanUnregisteredNumbers` is inert on any non-English or product-spec site** — it gates
  on `businessClaimContextRe`, an English business word list (clients|customers|records|…).
  A clean claims report on such a site means "no banned pattern matched", NOT "no invented
  numbers".

### Is a `claimed` work item stuck, or just slow? (I got this wrong once)

A `needs_page` item sits at `status='claimed'`, `claimed_by='generic'`, and its
`updated_at` **does not move for the whole build** — because the timestamp reflects the last
row write, not build progress. After ~13 minutes I read that as a stuck claim and was about
to reset it. It was healthy: each page runs an LLM call per section, so minutes per page is
normal.

**Progress lives in `orchestration_states`, not in the work item:**
```sql
SELECT orchestration_id, status, current_step, updated_at
  FROM orchestration_states
 WHERE updated_at > now() - interval '25 minutes'
 ORDER BY updated_at DESC LIMIT 12;
```
A live build shows `EXECUTING_STEP` at `call_content_writer` or
`process_sections_loop_iter_N_render_section`; the dispatch loop itself shows
`AWAITING_RESPONSES` at `process_item_iter_0_call_handler`. If you see those, wait — do not
re-queue, or you will duplicate work and lose a half-finished page.

Also: **the dispatch loop takes roughly one item per invocation.** Firing it three times in
a row does not parallelise a 12-page batch — the extra dispatches complete as no-ops while
one item is in flight. Fire once, wait on the page count with an `until` loop, fire again.
