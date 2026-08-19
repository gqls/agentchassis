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

## Council round-1 refinements (2026-08-18, corr 060bcc0a)

- **Before promoting past the canary, probe the BINARY for the strip gates, not just
  provenance** (debug_historian): on each replica,
  `kubectl exec <pod> -- grep -aq "strip_literal_markdown" /proc/1/exe && echo PRESENT`
  plus a nonsense-string negative control. Four Go files changed; the pod is the truth.
- **Scope note** (guardian): literal_markdown items only ever name BODY sections — the
  check's population is `page_components` rows. A markdown defect in site chrome/head
  (stored in site chrome artefacts, not page_components) is structurally outside this
  item_type and this repair path; if one is ever seen, it is a NEW check, not a widening.
- **True lifetime counts, live + archive** (prior_art_librarian — site_work_items alone
  is a ~7-day window): `page_rerender → page-rerender` = 13,993 complete / 142 failed
  (99.0%); `literal_markdown → page-build-handler` = 3 complete / 36 failed (7.7%);
  the retired first routing `literal_markdown → page-content-writer` = 2 complete /
  9 failed, all archived. Measured 2026-08-18 across both tables — the archive makes the
  case STRONGER in both directions.
- **A completed rerender with carried (bailed-out) sections cannot false-green**
  (render_guardian's concern): VerifyLiteralMarkdownResolved is registered for the
  item_type in init() and CompleteWorkItemAction runs it on EVERY completion
  (verifiers.go — RFC_017 fail-closed), not just canaries; a carried dirty section keeps
  the page's scan non-empty, so completion is refused and the item rides the attempt
  machinery to human review. Honest failure, not silent success.

## Verifying the strip flags (gotcha from the 8d lane, 2026-08-19)

A shallow query (`jsonb_each(default_config->'workflow'->'steps')`) finds only TWO of the
three flags and reads exactly like a half-applied 474 — page-content-writer's flag lives
inside `process_sections_loop`'s sub_workflow. Use the deep query:
```sql
SELECT type, jsonb_path_query_array(default_config, '$.**.strip_literal_markdown')
  FROM agent_definitions
 WHERE type IN ('page-rerender','page-content-writer','section-editor')
   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## Re-arming after the NEXT roll (the resolver-layer fix, f3939f27d)

1. Binary-probe both replicas for the new layer: `grep -aq "stripped literal markdown from fresh resolved_data" /proc/1/exe`
   (plus nonsense control).
2. Re-arm the two burned generic canaries + remaining generic items (same UPDATE as §Rollout
   step 4; they are `failed` at attempt 3, so reset `attempt_count=0`).
3. Expect news-page items to now converge; expect owned/ported-slot items to keep failing
   honestly (301's remit / tool-rebuild programme). Verify at the served page; the
   dartsonline items[18] summary will still show TABLE pipes (outside the pattern set —
   feed-quality follow-up).

## After the roll that carries round 6 (`f6d632291`): the kill switch, and probing EVERY pod

1. The r6 code is INERT until a roll. Prove it shipped at the binary, on EVERY
   agent-chassis pod — not two replicas (debug_historian, council r5). The
   kill-switch env NAME is the literal to probe, with both controls in the same breath:
   ```bash
   for p in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name); do
     printf '%s ' "$p"
     kubectl -n ai-persona-system exec ${p#pod/} -- sh -c \
       'grep -aq DISABLE_NEWS_MARKDOWN_STRIP /proc/1/exe && printf PRESENT || printf absent; printf " | ctrl+:"; grep -aq DISABLE_UNREGISTERED_HANDLER_DEMOTION /proc/1/exe && printf ok || printf FAIL; printf " ctrl-:"; grep -aq ZZ_NOT_A_REAL_SYMBOL_QQ /proc/1/exe && printf FAIL || echo ok'
   done
   ```
   Gotcha: the positive control (`DISABLE_UNREGISTERED_HANDLER_DEMOTION`) is a literal
   that has been in the binary since before v1.0.1300, so a pod where it is absent is a
   pod whose `/proc/1/exe` you are not reading, not a pod missing the feature.
2. **Disarming the news strip without a roll** (only if a feed shape ever needs raw
   markdown displayed — none known): set `DISABLE_NEWS_MARKDOWN_STRIP=1` in the
   agent-chassis deployment env (overlay patch + `apply -k`, or `kubectl set env` —
   remember the next `apply -k` reverts an imperative `set env`, the same way it reverts
   `kubectl scale`). Unset = strip ON. It is process-wide, not per-site. Re-arming is
   removing the variable. The only thing that exercises it otherwise is
   `TestProjectNewsItemsKillSwitchRestoresRawText`.
3. **Where the strip is visible**: `kubectl logs -l app=agent-chassis --tail=2000 | grep
   'stripped literal markdown from news items'` — `items_stripped`/`items` per
   projection. Zero lines over a day with news pages rendering means either the feed
   rows are clean (check `content_feed_items.source_summary` with the §Scope regexes) or
   the switch is set (check the env) — do not read silence as "working".

## Council rounds: find the run by PAYLOAD, and never re-run the trigger to re-read it
```sql
SELECT orchestration_id, created_at, current_step, status,
       md5(collected_data->'input_data'->>'rationale') AS payload
  FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' LIKE '060bcc0a%'
 ORDER BY created_at;
```
Gotcha (cost me a duplicate round, 2026-08-19): the 097 trigger PUBLISHES within its
first seconds; its printed output is in your scrollback or the task's output file. Two
rows with the same payload md5 seconds apart are one submission sent twice — both run,
both verdicts are valid, the second is pure credit cost.

## The exhaustiveness queries behind round 6 (prior_art asked for them verbatim, not prose)

Every live step at ANY depth running `render_component` / `rerender_page_sections`, with the
merge_with + strip flag (the shallow `jsonb_each(...->'steps')` misses sub_workflows — the
same trap as the strip-flag query above):
```sql
SELECT ad.type, s->>'action' AS action, s->'config'->>'merge_with' AS merge_with,
       s->'config'->>'strip_literal_markdown' AS strip, s->'config'->>'content_from' AS content_from
  FROM agent_definitions ad, jsonb_path_query(ad.default_config, '$.**') s
 WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
   AND jsonb_typeof(s)='object' AND s->>'action' IN ('render_component','rerender_page_sections')
 ORDER BY 1,2;
-- 2026-08-19 20:50Z: page-content-writer|render_component|current_section.resolved_data|true|generated_content.result  (render_section)
--                    page-content-writer|render_component|current_section.resolved_data|<null>|render_context          (render_from_template)
--                    page-rerender|rerender_page_sections||true|
```
Step NAMES (the deep query above loses the key): name page-content-writer's loop steps with
`jsonb_each(default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps')`.

Writers of `reason='literal_markdown'` (a grep over bodies, which the code index cannot do):
```bash
grep -rn '"literal_markdown"' --include=*.go platform/ | grep -v _test | grep -i reason
# 2026-08-19: check_literal_markdown.go:376 only (+ the operator UPDATE in §Rollout step 4)
```
Callers of `projectNewsItems` (the signature changed in r6):
```bash
grep -rn 'projectNewsItems(' --include=*.go platform/ | grep -v _test   # two, both news_items.go
```
Precedent for the kill switch, by name: `grep -n DISABLE_UNREGISTERED_HANDLER_DEMOTION platform/orchestration/actions/load_work_item_actions.go`.
