# RUNBOOK — bugs_open/029 tool-suggester phantom links

DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -f - < file.sql`
(pipe a `.sql` file with `-f -`; do NOT paste a heredoc line that starts with `\` — psql runs it
as a meta-command).

## R1 — fleet blast radius: every tool-suggester content_rewrite vs a real page

The load-bearing query. `matched_page_url` NULL = phantom. (2026-07-21: 0 of 24 matched.)

```sql
SELECT s.domain, swi.status,
       swi.spec->>'tool_function' AS tool_function,
       '/tools/'||(swi.spec->>'tool_function')||'.html' AS constructed_url,
       p.url AS matched_page_url, p.build_status
FROM site_work_items swi
JOIN sites s ON s.id = swi.site_id
LEFT JOIN pages p ON p.site_id = swi.site_id
     AND p.url = '/tools/'||(swi.spec->>'tool_function')||'.html'
WHERE swi.source = 'tool-suggester' AND swi.item_type = 'content_rewrite'
ORDER BY s.domain, tool_function;
```

## R2 — the item spec (see the fabricated URL in `suggestion` + `acceptance_test`)

```sql
SELECT jsonb_pretty(spec) FROM site_work_items
WHERE source='tool-suggester' AND item_type='content_rewrite'
  AND spec->>'tool_function' = 'tool-monitoring-coverage-gap-finder' LIMIT 1;
```

## R3 — actual deployed tool URL shapes (proves the emitter cannot reconstruct them)

```sql
SELECT s.domain, p.name, p.url, p.build_status
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.url ILIKE '/tools/%' ORDER BY s.domain, p.name;
```

## R4 — add_tool spec carries related_pages (the fix relies on this)

```sql
SELECT s.domain, swi.status, swi.spec->>'name' AS tool, swi.spec->>'function' AS function,
       swi.spec->'related_pages' AS related_pages, swi.handler_agent
FROM site_work_items swi JOIN sites s ON s.id=swi.site_id
WHERE swi.item_type='add_tool' ORDER BY swi.created_at DESC LIMIT 12;
```

## R5 — live page-build-handler threads spec.suggestion into the writer

(Backup SQL in `k8s/` is stale — check the live row, not the backup.)

```sql
SELECT default_config->'workflow'->'steps'->'call_content_writer'->'config'->'input_mapping'->>'rewrite_guidance?',
       image_tag
FROM agent_definitions
WHERE type='page-build-handler' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- expect: input_data.spec.suggestion
```

## Code anchors

- `platform/orchestration/actions/create_tool_cross_link_items.go:142` — the fabrication.
- `platform/orchestration/actions/deploy_tool_action.go:265-388` — library fork: creates page
  (url `:277`, strips `tool-` `:269`), emits `needs_content_page` `:373-385`.
- `platform/orchestration/actions/create_tool_component_action.go:211-293` — novel:
  `CanonicalisePage` `:211-218`, emits follow-on `:281-293`.
- `platform/orchestration/actions/validate_page_content.go:540-582` — in-body href check,
  warning-only (`:571`), non-blocking (`:257`).
- `platform/orchestration/actions/resolve_internal_links_action.go:98-105,319-350` — CTA-field
  resolver; does NOT touch in-body prose links.
- Suggester seed: `docs/agent_docs/sql_for_agents/062_tool_suggester_and_improver.sql`
  (`create_items_loop`, `spec_data: current_suggestion` at :1086/:1104) +
  `098_tool_suggester_cross_linking.sql` (`create_cross_links` step to remove).

## Build + verify (P3, once implemented)

```
# commit the task first (committed-ref build), then:
make build-agent-chassis            # bump IMAGE_TAG in makefile ~line 16 first
# push + deploy per house makefile targets, then verify against the POD:
kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "<a literal from the changed line>"'
```
Then trigger a fresh tool suggestion on a test site and assert the emitted `content_rewrite`
carries a URL equal to the tool page's real `pages.url` (R1 shows a non-NULL match).
