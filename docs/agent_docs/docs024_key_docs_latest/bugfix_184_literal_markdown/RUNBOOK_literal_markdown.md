# RUNBOOK — bugfix 184 (literal markdown)

DB access:
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## State of the item queue (the bug's live footprint)
```sql
SELECT status, count(*), max(created_at)::date FROM site_work_items
 WHERE item_type='literal_markdown' GROUP BY status ORDER BY 2 DESC;
```
Gotcha: `detected` rows are NOT waiting on triage — they are parked by migration 444's
promoter success floor (`literal_markdown → page-build-handler` is 1/28 lifetime, held until
the ratio recovers past 25%; ~9 hand-promoted successes unholds it). See bugs_open/184's
2026-08-17 CONSUMER NOTICE.

## What forms of markdown are actually live (findings breakdown)
```sql
SELECT f->>'pattern', f->>'source', count(*)
  FROM site_work_items swi, jsonb_array_elements(swi.spec->'findings') f
 WHERE swi.item_type='literal_markdown' AND swi.status NOT IN ('complete','cancelled','rejected')
 GROUP BY 1,2 ORDER BY count DESC;
```
Gotcha: this reads the *check's* findings, so it can only ever show the three patterns the
check knows (bold/code_span/heading). To see forms the check misses, scan content_data raw:
```sql
SELECT count(*) FILTER (WHERE pc.content_data::text ~ '\[[^\]]{2,60}\]\(https?://[^)]{4,120}\)') AS md_link
  FROM page_components pc;
```
(content_data::text is the JSON text — a newline inside a value appears as the two
characters `\n`, so line-anchored regexes must match `\\n` not `^`.)

## Prompt rule / check enablement (are 303/304 still live?)
```sql
SELECT (default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}')
       LIKE '%Plain string also means NO markdown syntax%' AS rule9_extended
  FROM agent_definitions WHERE type='page-content-writer' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
SELECT default_config #> '{workflow,steps,run_checks,config,checks}' ? 'literal_markdown'
  FROM agent_definitions WHERE type='quality-discovery-agent' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## Is another session working this bug?
```bash
python3 scripts/who-owns.py 184   # AMBIGUOUS number — the open case is the llm_markdown slug
grep -c literal_markdown ~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl | grep -v ':0'
```
Gotcha: who-owns reads COMMITS only; the transcript grep is what sees an uncommitted session.
A hit can be loaded context (MEMORY/LANDMINES), so read the hit lines before concluding.

## Rollout (order is load-bearing)

1. Part 2 at HEAD (check re-route + rerender hook — after the 299 lane's commit).
2. Build + push + deploy agent-chassis at a FRESH tag (same-tag rebuild ships the cached
   image — the 08-17 landmine). Verify by provenance stamp, not the roll:
   ```bash
   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
   git merge-base --is-ancestor <part2-commit> <the stamp && echo SHIPPED
   ```
   If the startup line has scrolled: binary probe with BOTH controls (expected sha
   present, absent sha absent).
3. Apply 473 then 474 (`kubectl exec -i postgres-clients-0 -- psql ... -f` shape; both
   print `OK` NOTICEs; a RAISE means an anchor moved — read the live row, re-anchor).
4. Canary TWO pages (different sites — they must be able to disagree):
   ```sql
   -- pick two open items on different sites, then per item:
   UPDATE site_work_items
      SET status='triaged', handler_agent='page-rerender', attempt_count=0,
          spec = spec || '{"reason":"literal_markdown"}'::jsonb
    WHERE id = '<item-id>';
   ```
   Dispatch rides build-dispatch-loop's normal cadence (or fire it for the site).
5. Verify at the ARTEFACT, not the item: curl the page, scan visible text for all four
   patterns (the bug file's §Scope query for the DB side). `result` may be the spawn
   record (bugs_open/287). Check the strip actually logged:
   `kubectl logs -l app=agent-chassis --tail=2000 | grep 'stripped literal markdown'`
   — a repair with ZERO strip lines and a clean page means the page was already clean
   (retraction should have caught it), so ask why it was open.
6. If both canaries clean: batch-promote the remaining open population (same UPDATE,
   widened WHERE — statuses detected/failed/unresolved; leave needs_human_review for a
   deliberate decision). The new pair earns its ratio; the 444 floor never engages
   (fresh pair). Next discovery pass retracts leftovers via Resolved.
7. Gotcha for step 6: items whose defect lives in rendered_html ONLY with clean
   content_data heal too (rerender regenerates rendered_html) UNLESS the markdown is in
   the component's html_template (1 template fleet-wide matches) — those will verify
   FAILED honestly; file the template case separately, do not weaken the verifier.
