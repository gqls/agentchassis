# RUNBOOK — bugfix 203 cleanup

Commands that were hard to get right, with their gotchas.

## Council verdict + full report for the source fix

```sql
-- verdict summary (doc_notes):
SELECT body FROM doc_notes WHERE body LIKE '%42eda9a5%' ORDER BY created_at DESC LIMIT 1;
-- full seat-by-seat report:
SELECT body FROM diagnosis_artifacts
WHERE kind='council_report' AND correlation_id='42eda9a5-6188-4e89-a11a-adb1dcbb135f'
ORDER BY created_at DESC LIMIT 1;
```
Gotcha: `diagnosis_artifacts` has no `content` column — the text is `body`. Schema-first.
Gotcha: `orchestration_states` lookup by
`collected_data->'input_data'->>'fix_correlation_id'` returned nothing for this corr —
the run had long completed; the durable artefacts are doc_notes + diagnosis_artifacts.

## Liveness of a Go fix without a pod-grep (ancestry route)

```bash
git merge-base --is-ancestor 880a405a6 1e349d046 && echo carried
# 1e349d046 = fix(197), pod-proven live on v1.0.1259 by that lane, real traffic.
kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o jsonpath='{range .items[*]}{.metadata.name} {.spec.containers[0].image}{"\n"}{end}'
```
Gotcha: only valid because builds are `git archive` from committed HEAD on a
forward-only tree AND the anchor commit was proven at the pod by someone. An image tag
alone proves nothing (bugs_open/153).

## Census of shipped phantom-CTA instances (stored artefact)

```sql
SELECT s.domain, p.url, pc.content_data->>'cta_text' AS cta_text
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=pc.site_id
WHERE pc.content_data ? 'cta_text' AND NOT (pc.content_data ? 'cta_url')
  AND pc.rendered_html ~ 'href="[^"]*"[^>]*>[^<]*</a>';
```
Gotcha: reads STORED html. Served pages can differ in both directions (LANDMINE:
"phantom check reads STORED html, not served"). Verify per-row at the live URL before
and after cleanup.

## Repairing a misdirected CTA through the framework (route 2, owner-authorised 2026-08-07)

**The shape**: a `content_rewrite` work item with `spec.mode='edit_live'`, then fire
`build-dispatch-loop` at the site. Worked example, with the before-state pinned:
`SQL_2026-08-07_canary_cta_repair_finetuning_risk_checker.sql`.

Two things that decide whether the item is ever picked up:
- **`status='triaged'`** (or `approved`) — `load_work_item_actions.go:653` selects
  `wi.status IN ('triaged','approved')`. Any other status and it sits for ever.
- **`handler_agent='page-build-handler'`** and **`page_id` in BOTH the column and
  `spec.page_id`**, plus `spec.page_name` (the dispatch loop maps `spec.page_name` →
  `input_data.page_name`).

Copy the spec keys from the one live emitter, `create_tool_cross_link_items.go:242-279` —
including its precedent for naming an exact URL in `suggestion` ("use that URL exactly as
written, do not alter or invent one"). That is what keeps a targeted repair
framework-legitimate instead of hand-authored: the URL comes from a real `pages` row and the
framework still writes the copy.

**Fire the dispatcher** (it is scheduler-driven with a fixed input that defaults to
`system.internal`, so it never fires for a real site on its own — see
`scripts/initial_messages/180_adoption/081b_trigger_dispatch_gamesdesign.sh`):

```bash
kubectl -n kafka run -i --rm kcat-dispatch-$(date +%s) --image=edenhill/kcat:1.7.1 \
  --restart=Never -- kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORR -H request_id=$REQ -H message_id=$MSG \
  -H orchestration_id=$ORCH -H orchestration_name=... -H step_name=start \
  -H client_id=demo_client -H message_type=request -H action=orchestrate \
  -H from_agent_type=user -H from_agent_id=cli \
  -H responses_topic=system.agent.generic.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"build-dispatch-loop"},"input_data":{"site_id":"$SITE_ID","domain":"$DOMAIN"}}
JSON
```

⚠ **`kcat -P` exits 0 having sent nothing** — its exit code is not evidence. Verify at the DB:
```sql
SELECT orchestration_id, current_step, status FROM orchestration_states
 WHERE orchestration_id='<ORCH>' OR correlation_id='<CORR>';
SELECT status, claimed_by, claimed_at FROM site_work_items WHERE item_key='<key>';
```
A claim shows up within ~15s when it works (measured: published 08:09:46Z, `claimed` by
`build-dispatch-loop` at 08:09:46.18Z, child orchestration at `spawn_content_writer` by
08:10:07Z). ⚠ Also: **no dispatch within ~300s of a chassis pod restart** — it is silently
dropped.

**Verify the repair at three layers, in this order** — a green work item is not a repaired page:
```sql
-- 1. content_data actually carries the resolved target
SELECT slot_name, content_data->>'cta_url', length(rendered_html), md5(rendered_html)
  FROM page_components WHERE page_id='<page_id>' ORDER BY position;
-- 2. nothing else moved: compare md5/length against the before-state pinned in the SQL file
-- 3. the claims guard did not refuse the save
SELECT context FROM agent_error_log WHERE error_code='CONTENT_CLAIMS_FLOOR_DETAIL'
  AND created_at > '<dispatch time>';
```
Then fetch the **served** URL — stored ≠ served, and the phantom check reads stored html.
