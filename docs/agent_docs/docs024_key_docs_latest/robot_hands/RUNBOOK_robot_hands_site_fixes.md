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
