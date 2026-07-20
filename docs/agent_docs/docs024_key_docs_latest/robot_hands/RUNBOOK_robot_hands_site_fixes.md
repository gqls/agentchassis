# RUNBOOK — robot-hands.com site fixes

Commands that were hard to get right, with the gotcha attached. Site
`00ff3af5-dad8-4770-9f70-3edc267a3c92`, domain `robot-hands.com`.

## Orientation

```bash
# DB
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

**Four places colour lives — a layout swap must move all of them.** Checking
one and declaring victory is how R1 stayed broken for a week:

```sql
SELECT ct.name AS theme, l.name AS layout, l.scheme, p.name AS palette
  FROM css_themes ct
  LEFT JOIN layouts l  ON l.id = ct.layout_id
  LEFT JOIN palettes p ON p.id = ct.palette_id
 WHERE ct.id = 'b1b60faf-ca68-43f5-a1e6-da3a769e4a25';

-- the three palette COPIES (component rendering reads style_collections):
SELECT colours->>'background'      FROM palettes          WHERE id='617e93c7-b1f1-4c5b-b7c4-482f3c0e9736';
SELECT color_palette->>'background' FROM style_collections WHERE id='cb95d40f-9bd2-4480-ba99-98b263aea44b';
SELECT color_palette->>'background' FROM css_themes        WHERE id='b1b60faf-ca68-43f5-a1e6-da3a769e4a25';
```

**Which components the slots actually use** — note `renderAndStoreSiteComponent`
ignores `is_active`, so a deactivated component renders happily:

```sql
SELECT sc.slot_name, cc.name, cc.is_active, sc.updated_at,
       (sc.rendered_html LIKE '%var(--color%') AS var_based,
       (sc.rendered_html LIKE '%#3b82f6%')     AS has_blue
  FROM site_components sc
  LEFT JOIN content_components cc ON cc.id = sc.component_id
 WHERE sc.site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92';
```

## Triggering agents (kcat)

**Gotcha: the JSON payload must be ONE line.** `kcat -P` splits on newlines and
you get a stream of malformed messages.

Re-render site components + queue every page:

```bash
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid); ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid); MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
kubectl -n kafka run -i --rm kcat-rr --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID -H orchestration_id=$ORCHESTRATION_ID \
  -H request_id=$REQUEST_ID -H message_id=$MESSAGE_ID -H message_type=request \
  -H client_id=demo_client -H action=orchestrate -H sender_agent_type=cli \
  -H sender_agent_id=cli-user -H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"demo_client","action":"orchestrate","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"agent_type":"rerender-pages"},"input_data":{"site_id":"00ff3af5-dad8-4770-9f70-3edc267a3c92","domain":"robot-hands.com","refresh_site_components":true}}
JSON
```

Swap `"agent_type":"rerender-pages"` for `"webdesign-agent"` to re-render CSS
(and drop `refresh_site_components`). **Order matters:** CSS first, then pages —
pages bake the components in at assembly.

**Before dispatching:** no orchestration within ~300s of a chassis pod restart
(the spawn is silently dropped), and check nothing is mid-flight:

```sql
SELECT id, item_type, claimed_by, claimed_at FROM site_work_items
 WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND status='claimed';
```

## Making a batch actually run

**Dispatch is one item per site at a time** — `build-pipeline-trigger` (every
30s) skips any site that has something `claimed`. A 50-item backlog takes hours,
and a fresh batch lands *behind* whatever churn is already queued. Promote it:

```sql
UPDATE site_work_items
   SET priority = 20, triaged_at = COALESCE(triaged_at, now()), updated_at = now()
 WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
   AND item_type = 'page_rerender' AND status = 'triaged' AND priority = 80;
```

Priority is **ASC** — lower runs sooner. `create_rerender_items` leaves
`triaged_at` NULL, which is harmless to dispatch but confusing to read.

**Watching a batch drain — use an ABSOLUTE cutoff, never `now() - interval`.**
A relative window silently empties as items age and the watcher reports "done"
when nothing finished:

```sql
SELECT status, count(*) FROM site_work_items
 WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92'
   AND item_type='page_rerender' AND created_at > '2026-07-18 15:00'
 GROUP BY status;
```

## Verifying — the live artefact, never the status

```bash
# the deployed page (gqls/sites is the repo; no local checkout exists)
curl -s https://robot-hands.com/index.html | grep -c '3b82f6'          # expect 0
curl -s https://robot-hands.com/index.html | grep -o '<header class="[^"]*"' | head -1
curl -s https://robot-hands.com/assets/css/styles.css | grep -E '^  --color-(background|card-bg):'

# page history when you need "what did it look like when it was good"
gh api "repos/gqls/sites/commits?path=robot-hands.com/index.html&per_page=20" \
  --jq '.[] | .sha[0:10] + "  " + .commit.author.date'
gh api "repos/gqls/sites/contents/robot-hands.com/index.html?ref=<sha>" --jq '.content' | base64 -d
```

`component_versions` is NOT a fallback — it holds no rows for header/footer/head.
Deployed git history is the only record of a previous good render.

**A Go change is inert until an image is built AND rolled.** Verify against the
running pod, never the tag or the commit:

```bash
POD=$(kubectl -n ai-persona-system get pods -o name | grep agent-chassis | head -1 | cut -d/ -f2)
kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c '<your unique symbol>'"
kubectl -n ai-persona-system get pod $POD -o custom-columns='IMAGE:.spec.containers[0].image,START:.status.startTime' --no-headers
```

Pick a symbol unique to *your* commit (a new message string), not one shared
with the previous version — that is how to tell which of two candidate commits
is actually live.

## Catching the "complete but failed" lie

`complete` is not proof the work happened. Sweep for the class:

```sql
SELECT id, site_id, item_type, completed_at
  FROM site_work_items
 WHERE status = 'complete' AND result->'response'->>'status' = 'failed';
```

## Council gate

```bash
# submit (JSON schema is in the 097 script header). Use ABSOLUTE paths — the
# shell's cwd resets between calls.
RESUBMIT_CORR=<corr> /home/ant/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh sub.json
```

```sql
-- verdict + per-seat objections (body is JSON: reviews[].objections[])
SELECT metadata->>'decision', created_at FROM diagnosis_artifacts
 WHERE correlation_id='<corr>' AND kind='council_report' ORDER BY created_at;
```

A round takes **~25 minutes** end to end (submission → `fix_plan` artifact →
`council_report`), not the ~2 minutes the docs suggest — size any waiter
accordingly. A round that produces no `fix_plan` artifact at all did not run;
see `bugs_open/019` (a truncated reviewer can void a whole round).

## Building a tool page by hand (R4, 2026-07-20)

**When:** the tool is data-backed. `tool-generator` has no fake-data rule
(`has_no_fake_data_rule=f`) and forbids fetch, so it invents the dataset —
`/bugs_open/020`. Check before routing anything through it:

```sql
SELECT type,
       (default_config::text ILIKE '%no fake data%')      AS has_no_fake_data_rule,
       (default_config::text ILIKE '%no fetch%')          AS forbids_fetch
FROM agent_definitions WHERE type IN ('tool-generator','tool-recreation-handler');
```
**Gotcha:** `agent_definitions` keys on `type`, not `name`.

**The shell.** Do not hand-write page chrome. Take a working tool page and splice:

```bash
curl -s https://robot-hands.com/tools/gripper-payload-calculator/index.html > pc.html
grep -n -oE '<(main|/main)' pc.html        # -> 423 <main   1024 </main
sed -n '1,423p'    pc.html > head.part     # head + header chrome + <main>
sed -n '1024,1083p' pc.html > tail.part    # </main> + footer
cat head.part MY_TOOL.html tail.part > index.html
sed -i 's#<title>.*</title>#<title>NEW TITLE</title>#' head.part   # BEFORE cat
```
**Gotcha:** `sed 's|…|…|'` breaks here — the replacement contains `|`. Use `#`.

**House style is `var(--color-*)` throughout** (R1's whole saga was hardcoded
hex). Check before deploying: `grep -c '#3b82f6' index.html` → 0.

**Test the JS.** No node locally; use a container:
```bash
docker run --rm -v "$PWD":/w -w /w node:20-alpine node --check mm.js
docker run --rm -v "$PWD":/w -w /w node:20-alpine node test_mm.js
```
A DOM stub of ~25 lines (`document.getElementById` returning fake elements with
`classList`/`addEventListener`) is enough to exercise the real submit handler —
see `NOTES` Turn 9. **Test the arithmetic, not just that it renders**: a tool that
renders and computes wrongly is worse than no tool.

**Deploy** (no local checkout of gqls/sites):
```bash
gh api -X PUT "repos/gqls/sites/contents/robot-hands.com/tools/<slug>/index.html" \
  -f message="..." -f content="$(base64 -w0 index.html)" --jq '.commit.sha'
```
Then **verify the artefact, not the commit** — takes ~1 min to publish:
```bash
until [ "$(curl -s -o /dev/null -w '%{http_code}' https://robot-hands.com/tools/<slug>/index.html)" = "200" ]; do sleep 15; done
```

## Auditing CTA label↔URL pairs (R4)

**Repoint by LABEL, never by the old URL.** A URL-keyed find-and-replace cements
every mismatch — 14 of 20 primaries here named a destination other than the one
they pointed at.

```sql
-- primaries: mind the THREE different field spellings
SELECT p.name, cc.name,
       COALESCE(pc.content_data->>'cta_text',  pc.content_data->>'primary_cta',
                pc.content_data->>'cta_label') AS label,
       COALESCE(pc.content_data->>'cta_url',   pc.content_data->>'primary_cta_url') AS url
FROM page_components pc JOIN pages p ON p.id=pc.page_id
LEFT JOIN content_components cc ON cc.id=pc.component_id
WHERE p.site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92'
ORDER BY p.name;

-- secondaries: these are the invisible ones — they 200, they are just wrong
SELECT p.name, pc.content_data->>'secondary_cta', pc.content_data->>'secondary_cta_url'
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND pc.content_data ? 'secondary_cta_url';
```

## Auditing statistics for fabrication (`/bugs_open/043`)

```sql
-- note BOTH spellings: stat_1_value AND stat1_value
SELECT p.name, e.k, e.v
FROM page_components pc JOIN pages p ON p.id=pc.page_id,
LATERAL jsonb_each_text(pc.content_data) AS e(k,v)
WHERE p.site_id='...' AND e.k ~ 'stat[_]?[0-9]+_?(value|label)'
ORDER BY p.name, e.k;
```
Every value must trace to a query. Ground truth for this site:
```sql
SELECT count(*)                                        FROM products WHERE site_id='...' AND category='gripper'; -- 5
SELECT count(DISTINCT specifications->>'manufacturer') FROM products WHERE site_id='...' AND category='gripper'; -- 5
SELECT count(*) FROM products p, LATERAL jsonb_each_text(p.specifications) s(k,v)
 WHERE p.site_id='...' AND p.category='gripper' AND s.k<>'manufacturer';                                          -- 24
```
**Also grep the rendered page** — the placeholder-suffix tell (`2,400+%`,
`140+ms`) is invisible in `content_data`:
```bash
curl -s <page> | grep -oE '[0-9,]+\+(%|ms|x)'
```

## Queueing a re-render that actually re-renders

Reason matters. Per `/bugs_open/024` an item with no usable reason falls to
stale-HTML assembly and the page never changes while every status reports green.
`cta_links_stale` reaches the real `rerender_sections` branch.

```sql
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, status, created_by, pipeline,
   priority, triaged_at, spec)
SELECT p.site_id, '<your-source-tag>', 'page_rerender', 'medium',
       'Rerender '||p.name||' — <why>', 'triaged', '<your-thread>', 'build',
       20, now(),
       jsonb_build_object('domain','robot-hands.com','reason','cta_links_stale',
                          'page_id',p.id,'page_name',p.name,'filename',ltrim(p.url,'/'))
FROM pages p WHERE p.site_id='...' AND p.name IN (...);
```
Required NOT NULL columns: `site_id, source, item_type, severity, summary,
status, created_by, pipeline`. Priority is **ASC** — 20 jumps the inherited
backlog. Watch it with an **absolute** cutoff, never `now() - interval`.
