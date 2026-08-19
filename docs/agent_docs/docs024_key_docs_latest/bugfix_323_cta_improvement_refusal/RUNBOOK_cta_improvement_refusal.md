# RUNBOOK — `bugs_open/323`, cta_improvement refusal completes green

The commands. Every query that had to be got right, with its gotcha attached. Update HERE when one changes.

```bash
# DB shell helper used throughout (ON_ERROR_STOP so a typo is loud)
Q() { kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -X "$@"; }
```

## The census (archive-inclusive — the live table is a ~7-day window)

Gotcha: `SELECT * FROM site_work_items UNION ALL SELECT * FROM site_work_items_archive` FAILS — the two
tables do not have the same column count. Name the columns.

```sql
WITH a AS (SELECT id,item_type,status,result,handler_agent,site_id,created_at FROM site_work_items
           UNION ALL SELECT id,item_type,status,result,handler_agent,site_id,created_at FROM site_work_items_archive)
SELECT item_type, status, count(*) total,
  count(*) FILTER (WHERE result #>> '{response,fix_result,fixed}'='true') fixed_true,
  count(*) FILTER (WHERE result #>> '{response,fix_result,fixed}'='false' AND result #>> '{response,fix_result,action}' IS NULL) noop_no_action,
  count(*) FILTER (WHERE result #>> '{response,fix_result,action}'='needs_review') needs_review,
  count(*) FILTER (WHERE result #>> '{response,fix_result,fixed}' IS NULL) unreadable,
  max(created_at)::date last
FROM a WHERE handler_agent='component-template-fixer' GROUP BY 1,2 ORDER BY 3 DESC;
```
The disconfirming result would be a non-zero `needs_review` count on a type whose no-op is legitimate, or
an idempotent-reason row (`already has flex CSS`) carrying `action=needs_review`. Neither occurs (2026-08-19).

## What the stored `result` actually is (287 / foreign-blob check)

```sql
SELECT (SELECT string_agg(k, ',' ORDER BY k) FROM jsonb_object_keys(COALESCE(result,'{}'::jsonb)) k) AS top_keys, count(*)
FROM site_work_items WHERE item_type='cta_improvement' AND status='complete' GROUP BY 1 ORDER BY 2 DESC;
```
`agent_id,agent_type,role,topics` = a SPAWN RECORD (bugs_closed/287, fixed 08-17 but parents reaped rows
stay wrong); `color_scheme,design_notes,spacing,typography` = webdesign-agent's design-token blob written
onto a fixer item (the March–May 441); `approach,reasoning,update_spec…` = a triage decision.

## Producers of `cta` findings (live, not the seed)

```sql
SELECT created_by, count(*), max(created_at)::date FROM site_work_items WHERE item_type='cta_improvement' GROUP BY 1;
```

## Did another route do the work? — grade at the artefact, never by re-running a detector

```sql
-- content_data history of the CTA-bearing component on the page the item names
SELECT h.created_at::timestamp(0), h.source, h.op,
  (SELECT jsonb_object_agg(k, v) FROM jsonb_each_text(COALESCE(h.content_data,'{}')) AS e(k,v) WHERE k ILIKE '%cta%') cta_fields,
  (SELECT string_agg(m[1]||' -> '||m[2], ' | ') FROM regexp_matches(h.rendered_html,
     '<a[^>]*href="([^"]*)"[^>]*class="[^"]*btn[^"]*"[^>]*>([^<]*)<', 'g') m) rendered_btns
FROM page_component_history h JOIN pages p ON p.id=h.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='<domain>' AND p.name='<page>' AND h.slot_name IN ('hero','call-to-action','call_to_action')
ORDER BY h.created_at DESC LIMIT 10;
```
Gotcha: `page_components` has NO `resolved_data` column and `page_component_history` has no
`change_source`/`changed_by` — the columns are `source`, `application_name`, `op`, `source_item_id`.

## The handler's live workflow (the seed is history)

```sql
SELECT jsonb_pretty(default_config->'workflow') FROM agent_definitions
WHERE type='component-template-fixer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## Who reads `action: needs_review`? (answer: nobody — check it, do not trust the comment at fix_component_template_action.go:58)

```bash
grep -rn '\["action"\]' --include=*.go platform/orchestration | grep -v _test | grep -v 'StepConfig\|config\['
```
```sql
SELECT type FROM agent_definitions, jsonb_object_keys(default_config->'workflow'->'steps') k
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->k->'config'->>'condition' LIKE '%needs_review%';
```

## Diagnosis loop

```bash
# fired 2026-08-19 ~15:55Z; intake corr 31375905-263c-4a86-8bce-bd600f788651, RUN corr b218f39d-48b7-400f-88c4-31254ce468b0
Q -c "SELECT current_step, status FROM orchestration_states WHERE correlation_id='b218f39d-48b7-400f-88c4-31254ce468b0' ORDER BY updated_at DESC LIMIT 3;"
Q -c "SELECT status, left(error,300) FROM site_work_items WHERE item_key='needs_diagnosis:cta-improvement-refusal-completes';"
Q -At -c "SELECT body FROM doc_notes WHERE body LIKE '%b218f39d%' OR body LIKE '%31375905%' ORDER BY created_at DESC LIMIT 1;"
```
Gotcha: origin/087_towards_multiple_domains was 325 commits behind local HEAD at dispatch; the loop reads
ORIGIN. The mechanism (punt arm from 2026-03-14, routing older) is on origin, so the diagnosis is valid for it;
anything about `complete_work_item_no_change.go`'s 08-18/19 state is NOT visible to it.

## Apply / verify migration 495 (done 2026-08-19 20:02Z)

```bash
# probe first in a doomed transaction (COMMIT→ROLLBACK copy), then apply, then record
sed 's/^COMMIT;$/ROLLBACK;/' docs/agent_docs/sql_for_agents/495_fixer_parks_refusals_as_needs_human_review.sql > /tmp/495_probe.sql
Q -f - < /tmp/495_probe.sql            # expect: NOTICE 495 verified … then ROLLBACK
Q -f - < docs/agent_docs/sql_for_agents/495_fixer_parks_refusals_as_needs_human_review.sql
./scripts/migration/run-migrations.sh --record-only docs/agent_docs/sql_for_agents/495_fixer_parks_refusals_as_needs_human_review.sql --note "…"
```
Gotcha: `run-migrations.sh --apply` takes EVERY pending file; apply by hand and `--record-only`.

## Prove the parking branch with a REAL dispatch (no side effects on any real site)

Never point `build-dispatch-loop` at a real site for this — it dispatches every open item on that site.
Dispatch the fixer itself at a probe item on `system.internal`:

```sql
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, priority, handler_agent, status, created_by, claimed_by, claimed_at, max_attempts, spec, item_key)
VALUES ('eac60db8-b032-432b-b36d-76f37632045d','bugfix_323_probe','build','cta_improvement','low','PROOF bugs_open/323 (TEMPORARY)',999,'component-template-fixer','claimed','bugfix_323_probe','bugfix_323_probe',now(),1,
 '{"category":"cta","fix_type":"cta_improvement","page_name":"index","audit_source":"bugfix_323_probe"}'::jsonb,'bugfix_323_probe:cta_improvement_refusal') RETURNING id;
```
```bash
# confirmed-publish pattern (payload in the container COMMAND, echo a marker) — kcat -P on stdin silently drops
PAYLOAD='{"action":"orchestrate","config":{"agent_type":"component-template-fixer"},"input_data":{"site_id":"eac60db8-b032-432b-b36d-76f37632045d","domain":"system.internal","work_item_id":"<ID>","item_type":"cta_improvement","source":"bugfix_323_probe","spec":{"category":"cta","fix_type":"cta_improvement","page_name":"index","audit_source":"bugfix_323_probe"}}}'
kubectl -n kafka run "kcat-323probe-$(date +%s)-$RANDOM" --rm --restart=Never --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "printf '%s' '$PAYLOAD' | kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 -t system.agent.generic.requests \
  -H correlation_id=$CORR -H request_id=$(uuidgen) -H message_id=$(uuidgen) -H orchestration_id=$(uuidgen) -H orchestration_name=bugfix-323-probe \
  -H step_name=start -H client_id=demo_client -H message_type=request -H action=orchestrate -H from_agent_type=user -H from_agent_id=cli \
  -H responses_topic=system.agent.generic.responses && echo PUBLISH_OK"
```
Expect within ~2 min: item `claimed → needs_human_review`, `.error` = the 495 literal; `orchestration_states.collected_data`
keys include `check_refused`,`park_refused`; a `doc_notes` row on system.internal titled `## refused: cta_improvement`.
**Teardown:** `DELETE FROM doc_notes WHERE id=<note>`; `DELETE FROM site_work_items WHERE id=<ID> AND created_by='bugfix_323_probe'`.
(Done 2026-08-19, corr `64f89f97`, both deleted.)

## Committing ONE hunk of a file another session has dirty (temporary index)

```bash
OLD=$(git rev-parse HEAD); export GIT_INDEX_FILE=/tmp/idx; rm -f /tmp/idx
git read-tree $OLD && git apply --cached mine.patch && TREE=$(git write-tree); unset GIT_INDEX_FILE
COMMIT=$(git commit-tree $TREE -p $OLD -F msg.txt) && git update-ref HEAD $COMMIT $OLD   # CAS: fails if HEAD moved
git reset -q -- <path>      # ⚠ REQUIRED: otherwise the shared index shows a STAGED REVERT of your hunk (status MM)
git diff --cached --name-only   # must not list your path
```

## Watch for the first REAL exercise of 495 (the probe is not a production item)

```sql
SELECT id, site_id, item_type, status, left(error,80), created_at FROM site_work_items
WHERE handler_agent='component-template-fixer' AND status='needs_human_review' AND error LIKE 'component-template-fixer REFUSED%'
ORDER BY created_at DESC LIMIT 10;
```

## Council + roll

```bash
# r1 corr 92829711-aecb-4e1a-8457-d011b4a635af
Q -c "SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts WHERE correlation_id='92829711-aecb-4e1a-8457-d011b4a635af' AND kind='council_report';"
Q -At -c "SELECT body FROM doc_notes WHERE categories ? 'council-gate' AND body LIKE '%92829711%' ORDER BY created_at DESC LIMIT 1;"
# did the Go half ship? (per SERVICE, not fleet)
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'   # then
git merge-base --is-ancestor 0e4622bab <stamp> && echo SHIPPED
```
