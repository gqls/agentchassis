# RUNBOOK — cta_dials_phone (bugs_open/299 slug + bugs_open/312)

Commands that were hard to get right, with their gotchas attached.

## Verify the bug on the served page (the false-pass trap)

```bash
curl -s https://preview.webdesign.uk/index.html | grep -o '<a[^>]*>[^<]*Brief Starter[^<]*</a>'
```
**Gotcha:** nav and footer link the tool CORRECTLY, so a page-wide grep for the correct URL
passes while the button stays broken. Assert on the anchor whose TEXT names a destination,
never on the URL's presence anywhere in the page. After the 08-18 rewrite the broken anchor's
text is "See how it works" — grep the cta-section block instead:
```bash
curl -s https://preview.webdesign.uk/index.html | grep -A3 'cta-btn-secondary'
```

## The stored pair (label vs url), per site

```sql
SELECT p.name, pc.slot_name, pc.updated_at,
       pc.content_data->>'secondary_cta' AS s_txt, pc.content_data->>'secondary_cta_url' AS s_url
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='webdesign.uk' AND pc.slot_name IN ('call-to-action','hero')
ORDER BY p.name, pc.position;
```

## Fleet census of CTA url scopes (sizes the class)

```sql
SELECT CASE WHEN v.url LIKE 'tel:%' THEN 'tel' WHEN v.url LIKE 'mailto:%' THEN 'mailto'
            WHEN v.url LIKE 'http%' OR v.url LIKE '//%' THEN 'external'
            WHEN v.url LIKE '#%' THEN 'anchor' WHEN v.url = '' THEN 'empty' ELSE 'page' END AS scope,
       count(*), count(DISTINCT s.domain)
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id,
LATERAL (VALUES (pc.content_data->>'cta_url'),(pc.content_data->>'primary_cta_url'),
                (pc.content_data->>'secondary_cta_url')) AS v(url)
WHERE v.url IS NOT NULL AND p.status='active' GROUP BY 1 ORDER BY 2 DESC;
-- 2026-08-18: page 1006 | empty 27 | tel 5 | external 2 | anchor 1 | mailto 1
```

## The 312 wiring proof (resolver output vs what rendered)

Find a page-content-writer orchestration and compare the two sides IN ITS OWN collected_data:

```sql
SELECT jsonb_path_query_array(collected_data->'resolved_links'->'response',
         '$.sections_ready[*] ? (@.name == "call-to-action").resolved_data') AS resolver_wrote,
       jsonb_path_query_array(collected_data->'sections_for_render',
         '$.sections_ready[*] ? (@.name == "call-to-action").resolved_data') AS render_used
FROM orchestration_states WHERE orchestration_id='05e3839d-8e18-4935-9c7e-3c6d741665d6';
```
**Gotchas:** (1) `left(jsonb,n)` does not exist — cast `(x)::text` first. (2) The parent holds
the child's result under BOTH `resolve_links` and `resolved_links`, each as `{response: …}`;
the config path reads `resolved_links`. (3) Retention: `resolved_links` rows go back only to
08-17 — do not claim "never in history", claim the 0/150 window.

The negative/positive control pair (path 1 has never matched; the real shape is normal):

```sql
SELECT count(*) AS runs,
       count(*) FILTER (WHERE collected_data->'resolved_links'->'response' ? 'link_resolution') AS path1_would_hit,
       count(*) FILTER (WHERE collected_data->'resolved_links'->'response' ? 'sections_ready') AS real_shape
FROM orchestration_states WHERE collected_data ? 'resolved_links';
-- 2026-08-18: 150 | 0 | 149
```

## The live select_sections config (the defective path)

```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'select_sections')
FROM agent_definitions WHERE type='page-content-writer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## Does a value reach the writer prompt? (the measurement trap)

```sql
-- WRONG (what I did first): counts the guidance SENTENCE, which contains the field name
SELECT count(*) FROM llm_call_log WHERE prompt_rendered LIKE '%_target_title%';
-- RIGHT: separate the phrase from a value-shaped occurrence
SELECT count(*) FILTER (WHERE prompt_rendered LIKE '%e.g. cta_target_title for cta_url%') AS guidance_text,
       count(*) FILTER (WHERE prompt_rendered ~ '(primary|secondary|cta)_target_title"?\s*[:=]') AS value_shaped
FROM llm_call_log WHERE created_at > now() - interval '36 hours' AND prompt_rendered LIKE '%_target_title%';
-- 2026-08-18: 179 guidance / 0 value-shaped (of 182)
```

## Ownership / collision checks before touching CTA machinery

```bash
./scripts/who-owns.py 248     # the page-scheme keep half — bugfix_248_authored_cta_destinations, ACTIVE
./scripts/who-owns.py 299     # this bug (by slug; number is ambiguous)
grep -n "cta_links_stale" docs/agent_docs/docs024_key_docs_latest/LANDMINES.md
```
**Gotcha:** `who-owns` reads COMMITS — a session mid-fix is invisible. `ListAgents` + live
`.jsonl` transcripts are the uncommitted check. The `bugfix 248` peer session exists.

## Is the fix actually in the running binary? (the check that ISN'T available)

**Do not start with `git merge-base --is-ancestor <commit> <stamp>`.** Twice out of twice on
2026-08-19 the stamp was unobtainable and the obvious fallback lied. Full trap in
`LANDMINES.md` ("Probing the live binary for YOUR commit returns ABSENT…"). In short: the
`build provenance` line is a STARTUP line (gone from a full `kubectl logs` three hours after
the roll), and `buildinfo.GitCommit` is ONE string, so grepping the binary for your own
commit says *absent* for a binary that contains it.

**Ask for the CAPABILITY instead, on every pod, always with a control that must fail:**

```bash
for POD in $(kubectl -n ai-persona-system get pods -l app=agent-chassis \
               -o jsonpath='{range .items[*]}{.metadata.name} {end}'); do
  echo "== $POD"
  for s in NormalizeTelHref IsAuthoredNonPageCTADestination DescribeCTADestination \
           storedCTADestinationIsAuthored cta_nonpage_destination NormalizeTelHrefXX; do
    printf '   %-34s ' "$s"
    kubectl -n ai-persona-system exec $POD -- grep -aq "$s" /proc/1/exe && echo PRESENT || echo absent
  done
done
# 2026-08-19, v1.0.1316: all five PRESENT on both pods; NormalizeTelHrefXX absent (control).
```
**Gotchas:** (1) unexported Go helpers DO probe (pclntab keeps names for stack traces) — 2 hits
each is normal. (2) `grep -c` inside `exec` exits 1 on zero matches and kills a `&&`-chain, so
use `-q` with an explicit `|| echo`. (3) One pod is not the fleet — a partial roll makes
replicas disagree and the failing run is the one that lands on the old one. (4) Six symbols ×
two pods through `kubectl exec` takes >2 minutes; budget for it or split the loop.

## Applying ONE migration when 25 other lanes' files are pending

`--apply` takes EVERY pending file. On 2026-08-19 the pending set was 25+ files from other
lanes, several with drifted pre-gates. The narrow, sanctioned path:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/NNN_x.sql
./scripts/migration/run-migrations.sh --record-only NNN_x.sql --note "<what you verified>"
```
**Gotchas:** (1) stdin-to-psql has `-f` semantics — that satisfies "run files, never paste".
(2) `--record-only` takes the filename as its OWN argument (`--record-only F`), and is mutually
exclusive with `--apply`. (3) The dry run probes every pending file through `kubectl exec` and
took ~10 minutes with no output until the end — it is not hung; do not kill it. (4) The README
rule "every migration touching `agent_definitions` opens with `snapshot_agent('<type>', …)`"
is easy to miss when you have written a bespoke `_backup_NNN` table instead — 475 and 476 both
had, and both needed the snapshot added before applying.

## The discard control, measured correctly (the trap is in the cast)

```sql
WITH r AS (
  SELECT jsonb_path_query_array(collected_data->'resolved_links'->'response',
           '$.sections_ready[*].resolved_data')::text AS res,
         jsonb_path_query_array(collected_data->'sections_for_render',
           '$.sections_ready[*].resolved_data')::text AS ren
  FROM orchestration_states
  WHERE collected_data ? 'resolved_links' AND collected_data ? 'sections_for_render'
)
SELECT count(*) AS runs,
       count(*) FILTER (WHERE res LIKE '%_target_title%')                          AS resolver_minted,
       count(*) FILTER (WHERE res LIKE '%_target_title%' AND ren LIKE '%_target_title%')     AS survived,
       count(*) FILTER (WHERE res LIKE '%_target_title%' AND ren NOT LIKE '%_target_title%') AS discarded,
       count(*) FILTER (WHERE res = ren) AS identical, count(*) FILTER (WHERE res <> ren) AS differ
FROM r;
-- 2026-08-19: 48 | 26 | 0 | 26 | 18 | 30
```
**Gotcha, and it inverts the answer:** casting the CONTAINER
(`(collected_data->'sections_for_render')::text LIKE '%_target_title%'`) matches the string
elsewhere in the structure and returns **31 / 31 / 0** — "everything survives", the comfortable
and wrong result. Anchor the cast to `resolved_data` on BOTH sides. Logged in `WRONG_CALLS.md`.

## Pushing a `content_data` fix to the served page (three tries, all reporting success)

```bash
# BOTH keys are required. reason -> re-render the sections; page_name -> SAVE them.
kubectl -n kafka run -i --rm kcat-rr-$(date +%s) --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$(uuidgen) -H orchestration_id=$(uuidgen) -H request_id=$(uuidgen) \
  -H message_id=$(uuidgen) -H message_type=request -H client_id=demo_client -H action=orchestrate \
  -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses -H timestamp=$(date -u +%FT%TZ) <<'JSON'
{"action":"orchestrate","config":{"agent_type":"page-rerender"},"input_data":{
  "site_id":"<site uuid>","domain":"<domain>","page_id":"<page uuid>",
  "spec":{"reason":"cta_links_stale","page_name":"<page name>"}}}
JSON
```
**Gotchas — and all three failure modes report success** (full trap in `LANDMINES.md`,
*"A `page-rerender` dispatched without `spec.page_name`…"*):
1. **No `spec.reason`** → `check_rerender_mode` routes to `render_page`, which assembles the
   components' STORED `rendered_html`. Your `content_data` fix never reaches the page and a real
   git commit ships the old bytes.
2. **No `spec.page_name`** → sections ARE re-rendered, then `save_page_sections` returns
   `{"success": true, "sections_saved": 0, "skipped": true, "reason": "no page name"}` and the
   workflow deploys the stale assembly anyway.
3. `cta_links_stale` is only safe on a chassis carrying bug 299's KEEP #3 (**≥ v1.0.1317**); on
   an older binary that reason IS the clobber trigger (register LNK-034).

**Assert on the save and then on the artefact, never on the orchestration status:**
```sql
SELECT collected_data->'sections_saved'->>'sections_saved' AS saved,
       collected_data->'sections_saved'->>'reason'         AS skip_reason
  FROM orchestration_states WHERE orchestration_id='<yours>';   -- saved must be NON-ZERO
```
then confirm `page_components.updated_at` MOVED and `rendered_html` contains the new value. A
row still carrying the timestamp of *your own SQL fix* is the tell it was never touched.
Finally read the deployed commit, which is authoritative before the CDN catches up:
```bash
gh api repos/gqls/vm-sites/contents/<domain>/index.html?ref=<commit_sha> --jq '.content' \
  | base64 -d | grep -o '<a[^>]*cta-btn-secondary[^>]*>[^<]*</a>'
```

## Pruning `service_binary_capabilities` (BLD-023) until the next roll

The retention prune ships in the Go half and is **inert until a chassis roll carries it**. Until
then the table grows ~20k rows/hour, because the binary runs as ephemeral per-job pods:
```sql
DELETE FROM service_binary_capabilities WHERE last_seen_at < now() - interval '2 hours';
VACUUM (ANALYZE) service_binary_capabilities;
```
**Gotcha:** plain `VACUUM` marks space reusable, it does not return it to the OS — so
`pg_total_relation_size` still reads the old figure after a big delete. That is expected; judge
by `count(*)`, not by size.
