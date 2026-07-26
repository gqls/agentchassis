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

---

## R6 — the sweep AFTER the fix (2026-07-26): key on item_key, not source

R1 keys on `source='tool-suggester'` and reconstructs the URL itself. Both stop being right once
the fix ships: the emitter is now the build agent (`source` = `tool-deployer`/`tool-generator`),
and the URL is in the spec rather than derivable. Key on the item_key namespace, which is stable
across both eras, and join on the spec's own URL:

```sql
SELECT s.domain, swi.status, swi.created_at::date, swi.source,
       swi.spec->>'tool_function' AS tool_function,
       swi.spec->>'tool_page_url' AS spec_url,
       p.url AS matched_page_url, p.build_status,
       swi.depends_on
FROM site_work_items swi
JOIN sites s ON s.id = swi.site_id
LEFT JOIN pages p ON p.site_id = swi.site_id AND p.url = swi.spec->>'tool_page_url'
WHERE swi.item_key LIKE 'tool_crosslink:%'
ORDER BY swi.created_at DESC;
```

- Rows with `spec_url` NULL are pre-fix rows — the existing damage (27 as of 2026-07-25).
- Rows created after the image roll must have `matched_page_url` non-NULL. One that doesn't is a
  regression, not a residual.
- `depends_on` non-NULL is expected on a tool whose page was still `planned` at emit time.

## R7 — applying a migration when other threads have files pending

`run-migrations.sh --apply` applies **every** pending file in the directory, other threads'
included (9 were pending on 2026-07-26). To apply only your own:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/NNN_x.sql
./scripts/migration/run-migrations.sh --record-only NNN_x.sql --note "why it went out of band"
```

Read the apply output with `| head -25` (the NOTICEs and the COMMIT are at the top; the
verification SELECTs at the bottom will fill your screen and hide them). Do NOT re-run the file
to see output you missed — an idempotent migration re-runs cleanly but duplicates its `doc_notes`
row and takes a second set of snapshots.

## R8 — is the fix live? (post-roll)

```bash
# 1. the binary carries the new emitter (a string the CHANGE created, plus a control)
kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "emitToolCrossLinkItems: refusing to emit without a real tool page URL"'   # expect >=1
kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "CreateToolCrossLinkItemsAction: Starting"'                                # control, expect >=1

# 2. config half (live since 2026-07-26, independent of the image)
#    create_cross_links absent, create_items_loop -> complete, related_pages wired on both builds
```
Then trigger a tool build on a test site and check R6: the new row's `spec_url` must equal the
tool page's `pages.url`.
