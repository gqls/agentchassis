# RUNBOOK — bugfix_380_claims_fail_open

Every command here was hard to get right at least once. Gotcha attached to each.

## 1. Dispatch one claims audit by hand (the proof instrument)

```bash
./docs/agent_docs/docs024_key_docs_latest/bugfix_380_claims_fail_open/TRIGGER_claims_audit.sh <domain>
```
- Refuses a domain with no `sites` row — `ensure_site_record` upserts by domain and would
  otherwise MINT a site. Never bypass that.
- Publishes via `scripts/kafka-publish-lib.sh` (receipt asserted). The `kubectl run -i` kcat
  form in older triggers drops ~4 publishes in 5 at exit 0.
- No dispatch within ~300s of a chassis pod restart (the spawn is silently dropped).
- Measured 2026-08-24: quiet queue → COMPLETED in ~25s; under load budget ~30 min.

## 2. Did the model actually LOOK? (the demand control — a `[]` is not a verdict)

```sql
SELECT created_at, input_tokens, output_tokens,
       position('no verified facts are registered for this business' in prompt_rendered) > 0 AS cold_arm,
       position('<a sentence you KNOW is on the page>' in prompt_rendered) > 0 AS sentence_reached_model,
       left(response_text, 300)
  FROM llm_call_log WHERE agent_type='claims-auditor' ORDER BY created_at DESC LIMIT 3;
```
- `cold_arm = t` on a site with no facts; `f` on a site with a roster (the control).
- `sentence_reached_model = f` with the sentence present in `page_components.rendered_html`
  is the 601 defect class (extraction), not a clean page. Before 601 this was the case on
  EVERY multi-component page (Postgres greedy-first regex; see LANDMINES).
- `llm_call_log` is the training corpus (owner 2026-08-22) — never delete rows from it.

## 3. Coverage: which sites were not audited in N days, and what did each run do?

The rotation stamp records SELECTION; the doc_notes receipt records what the run DID.
Join both; never read the stamp alone.

```sql
WITH last_run AS (
  SELECT DISTINCT ON (site_id) site_id, created_at, categories
    FROM doc_notes
   WHERE subject_type='pipeline' AND subject_key='claims-audit' AND site_id IS NOT NULL
   ORDER BY site_id, created_at DESC)
SELECT s.domain, s.status,
       r.last_selected_at AS last_selected,
       lr.created_at      AS last_receipt,
       CASE WHEN lr.categories ? 'audit-findings' THEN 'ran: findings filed'
            WHEN lr.categories ? 'audit-ran'      THEN 'ran: clean'
            ELSE 'NEVER RECEIPTED' END AS last_outcome,
       (r.last_selected_at IS NOT NULL AND (lr.created_at IS NULL OR lr.created_at < r.last_selected_at)) AS selected_but_no_receipt,
       EXISTS (SELECT 1 FROM site_specs ss WHERE ss.site_id=s.id AND ss.aspect='evidence_base' AND ss.is_current
                 AND jsonb_array_length(COALESCE(ss.data->'facts','[]'::jsonb)) > 0) AS has_facts
  FROM sites s
  LEFT JOIN site_discovery_rotation r ON r.site_id=s.id AND r.agent_type='claims-auditor'
  LEFT JOIN last_run lr ON lr.site_id=s.id
 WHERE s.status IN ('active','deployed')
   AND (lr.created_at IS NULL OR lr.created_at < now() - interval '7 days')
 ORDER BY lr.created_at NULLS FIRST, s.domain;
```
- `selected_but_no_receipt = t` means the rotation dispatched and the run never wrote its
  receipt — a FAILED run (RFC_017 fail-closed; look in `agent_error_log` / the immune
  sweep), not a clean one.
- A site absent from the rotation entirely: check it has a page with
  `build_status IN ('deployed','active')` and `locked_at IS NULL` — 600's pre_query.

## 4. Findings

```sql
SELECT s.domain, w.item_key, w.status, w.created_at, left(w.summary, 80)
  FROM site_work_items w JOIN sites s ON s.id=w.site_id
 WHERE w.item_type='claims_unverified' ORDER BY w.created_at DESC;
```
- `claims_llm_<domain>` = the LLM auditor (this lane); `claims:<page_id>` = the Go discovery
  check. Both are `claims_unverified`; split by prefix or you conflate two producers.
- The LLM item carries no `spec.page_id` and parks under the revalidator's `spec_no_page_id`
  arm — pre-existing shape (RFC_012-era), recorded as a limit, not fixed here. While it is
  open, later passes dedup onto it (`idx_swi_dedup`).

## 5. Live-template anchor counts before ANY prompt splice

```sql
SELECT (length(t)-length(replace(t,$a$<ANCHOR>$a$,'')))/length($a$<ANCHOR>$a$) AS n
  FROM (SELECT default_config #>> '{workflow,steps,<STEP>,config,prompt_template}' AS t
          FROM agent_definitions WHERE type='<agent>' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) x;
```
- page-content-writer's template is NOT at `steps.generate_content` — it is at
  `steps.process_sections_loop.config.sub_workflow.steps.generate_content.config.prompt_template`
  (a loop sub-workflow). The claims-auditor's LLM step uses key `prompt`, not `prompt_template`.
- The committed migration text is NOT the live text (writer: 330's base64 is 1,718 chars
  behind live). Count on live, generate any plaintext for a human read from live.

## 6. Migrations: apply and record

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/NNN_name.sql
./scripts/migration/run-migrations.sh --record-only NNN_name.sql --note "<what you verified>"
```
- Every file here opens with `snapshot_agent(...)` and ends with a DO/RAISE verify — a verify
  block of bare SELECTs verifies nothing (ON_ERROR_STOP ignores a non-empty result).
- `_HOLD` files are excluded by the runner; 599 stays `_HOLD` until the owner approves
  `brochure_component_library/sql/page_content_writer_prompt_v5_2026-08-24.txt`.

## 7. DB hygiene learned the hard way (2026-08-24)

- Shell `timeout N` around a kubectl-exec psql kills the CLIENT; the server backend keeps
  running (mine queued for an hour behind a lock). Use
  `PGOPTIONS='-c statement_timeout=60000'` on the psql side instead, or cancel your own pid.
- A `COPY … TO STDOUT` whose client died sits in `wait_event = ClientWrite` for ever, holding
  its locks; `pg_cancel_backend` cannot reach it (the backend is blocked in a write) —
  `pg_terminate_backend` is the remedy. Diagnose first:
  `SELECT pid, pg_blocking_pids(pid), wait_event_type, wait_event, now()-query_start FROM pg_stat_activity WHERE wait_event_type IN ('Lock','Client')`.
- When a trivial query on a table times out, time `SELECT count(*) FROM <table>` before
  believing your own query is slow (the 378 lane's lesson from the same incident).
