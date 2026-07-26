# RUNBOOK — chrome lock gate (`bugs_open/069`) and the snapshot lock wipe (`bugs_open/088`)

Commands that were hard to get right, with the gotcha attached. Change them HERE, not in scrollback.

## Is anyone else on this bug?

```bash
python3 scripts/who-owns.py 069            # ~0.3s, no cluster calls
```
**Gotcha:** it reports "OWNED or recently active" for a bug whose only commit is the one that FILED
it. Read the commit subjects it prints before believing the verdict.

## Live lock state (never read this from the repo's .sql files — they have drifted)

```sql
SELECT count(*) total, count(*) FILTER (WHERE locked_at IS NOT NULL) locked FROM site_components;
SELECT count(*) total, count(*) FILTER (WHERE locked_at IS NOT NULL) locked FROM page_components;
SELECT sc.slot_name, sc.locked_by, sc.lock_type, sc.lock_expires_at
  FROM site_components sc JOIN sites s ON s.id = sc.site_id
 WHERE sc.locked_at IS NOT NULL;
```
**Gotcha:** `lock_type='admin'` violates `chk_site_components_lock_type` — the legal values are
`permanent | timed | review`, or NULL. A NULL type on a locked row is treated as hard.

## Which callers can actually trip the chrome gate

```sql
SELECT a.type, s.value->'config'->>'force_rerender'
  FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s
 WHERE a.is_active AND NOT COALESCE(a.is_snapshot,false) AND a.deleted_at IS NULL
   AND s.value->>'action' = 'render_site_components';
```
**Gotcha:** the gate sits below the `!force` idempotence exit, so only `force_rerender: true` callers
reach it (4 of 6 on 2026-07-26). A "verification" driven by an unforced call passes vacuously.

Does anything downstream read the action's result?

```sql
SELECT count(*) FROM agent_definitions
 WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
   AND default_config::text LIKE '%render_site_components.rendered%';   -- 0 on 2026-07-26
```

## Reading a DB function — the live body, not the repo copy

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -t -A -c "SELECT pg_get_functiondef(p.oid) FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace
            WHERE n.nspname='public' AND p.proname='take_site_snapshot';" > take_site_snapshot.sql
```
**Gotchas, both paid for on 2026-07-26:**
- `docs/agent_docs/sql_for_tables/031_site_snapshots.sql` disagrees with the live body. A subagent
  audit read the file and reported the capture set wrongly. Read `pg_get_functiondef`.
- `pg_get_functiondef` emits **no trailing semicolon**. Concatenate two of them into a migration and
  psql fails with `syntax error at or near "CREATE"` — append `;` after each `$function$`.

## Applying a migration when other threads have pending files

```bash
./scripts/migration/run-migrations.sh                 # dry run: lists ALL pending files
# probe just yours, rolled back:
sed 's/^COMMIT;$/ROLLBACK;/' docs/agent_docs/sql_for_agents/219_*.sql \
  | kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f -
# apply only yours:
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - \
  < docs/agent_docs/sql_for_agents/219_*.sql
./scripts/migration/run-migrations.sh --record-only 219_....sql --note "why it was applied by hand"
```
**Gotcha:** `--apply` runs **every** pending file in order, including other threads' (8 were pending
on 2026-07-26). Apply yours by hand, then `--record-only` so the ledger stays honest.

## Verifying a DB-side fix without leaving fixtures behind

Wrap the entire induced-fault test in `BEGIN; … ROLLBACK;`. It exercises the real deployed functions
against the real schema; only the commit is withheld, so the fleet never sees the fixtures and there
is nothing to clean up. Full script:
`/bugs_closed/088` "How it was verified" — assertions RAISE, so silence is failure.
**Include a control:** assert that an OLD snapshot row does NOT have the new key. Without it a test
that would have passed before the fix looks like a proof.

## Rollback artefact for migration 219 — the function definitions as they were

Captured 2026-07-26, immediately before 219 was applied. To roll back, run these two definitions.

```sql
CREATE OR REPLACE FUNCTION public.take_site_snapshot(p_site_id uuid, p_trigger text, p_git_sha text DEFAULT NULL::text, p_label text DEFAULT NULL::text, p_created_by text DEFAULT 'system'::text)
 RETURNS uuid
 LANGUAGE plpgsql
AS $function$
DECLARE
    v_snapshot_id   UUID;
    v_site_record   JSONB;
    v_spec_snapshot JSONB;
    v_spec_ids      UUID[];
    v_pages         JSONB;
    v_nav           JSONB;
    v_components    JSONB;
BEGIN
    SELECT jsonb_build_object(
        'id', s.id,
        'domain', s.domain,
        'status', s.status,
        'company_name', s.company_name,
        'tagline', s.tagline,
        'schema_mode', s.schema_mode,
        'style_collection_id', s.style_collection_id,
        'default_components', s.default_components,
        'content_data', s.content_data,
        'brand_assets', s.brand_assets,
        'deploy_config', s.deploy_config,
        'last_built_at', s.last_built_at,
        'last_deployed_at', s.last_deployed_at
    ) INTO v_site_record
    FROM sites s
    WHERE s.id = p_site_id;

    IF v_site_record IS NULL THEN
        RAISE EXCEPTION 'Site % not found', p_site_id;
    END IF;

    SELECT
        COALESCE(jsonb_agg(
            jsonb_build_object(
                'id', ss.id,
                'aspect', ss.aspect,
                'data', ss.data,
                'source', ss.source,
                'source_agent', ss.source_agent,
                'created_by', ss.created_by,
                'pinned', ss.pinned,
                'created_at', ss.created_at
            ) ORDER BY ss.aspect
        ), '[]'::jsonb),
        COALESCE(array_agg(ss.id), ARRAY[]::uuid[])
    INTO v_spec_snapshot, v_spec_ids
    FROM site_specs ss
    WHERE ss.site_id = p_site_id AND ss.is_current = true;

    SELECT COALESCE(jsonb_agg(page_row ORDER BY page_row->>'nav_order'), '[]'::jsonb)
    INTO v_pages
    FROM (
        SELECT jsonb_build_object(
            'id', p.id,
            'name', p.name,
            'url', p.url,
            'title', p.title,
            'page_type', p.page_type,
            'status', p.status,
            'meta_description', p.meta_description,
            'topics', to_jsonb(p.topics),
            'nav_label', p.nav_label,
            'nav_order', p.nav_order,
            'in_header', p.in_header,
            'in_footer', p.in_footer,
            'build_status', p.build_status,
            'version', p.version,
            'sections', p.sections,
            'rendered_header', p.rendered_header,
            'rendered_footer', p.rendered_footer,
            'rendered_head', p.rendered_head,
            'page_spec', p.page_spec,
            'content_direction', p.content_direction,
            'site_area_id', p.site_area_id,
            'components', COALESCE(
                (SELECT jsonb_agg(
                    jsonb_build_object(
                        'id', pc.id,
                        'component_id', pc.component_id,
                        'position', pc.position,
                        'slot_name', pc.slot_name,
                        'rendered_html', pc.rendered_html,
                        'content_data', pc.content_data,
                        'build_status', pc.build_status
                    ) ORDER BY pc.position
                )
                FROM page_components pc
                WHERE pc.page_id = p.id
                ), '[]'::jsonb
            )
        ) AS page_row
        FROM pages p
        WHERE p.site_id = p_site_id
    ) sub;

    SELECT jsonb_build_object(
        'groups', COALESCE(
            (SELECT jsonb_agg(
                jsonb_build_object(
                    'id', g.id,
                    'group_key', g.group_key,
                    'group_label', g.group_label,
                    'group_type', g.group_type,
                    'parent_group_id', g.parent_group_id,
                    'position', g.position
                ) ORDER BY g.position
            )
            FROM site_nav_groups g
            WHERE g.site_id = p_site_id
            ), '[]'::jsonb
        ),
        'items', COALESCE(
            (SELECT jsonb_agg(
                jsonb_build_object(
                    'id', ni.id,
                    'group_id', ni.group_id,
                    'parent_item_id', ni.parent_item_id,
                    'label', ni.label,
                    'url', ni.url,
                    'page_id', ni.page_id,
                    'item_type', ni.item_type,
                    'position', ni.position,
                    'status', ni.status,
                    'metadata', ni.metadata
                ) ORDER BY ni.position
            )
            FROM site_nav_items ni
            WHERE ni.site_id = p_site_id
            ), '[]'::jsonb
        )
    ) INTO v_nav;

    SELECT COALESCE(jsonb_agg(
        jsonb_build_object(
            'id', sc.id,
            'slot_name', sc.slot_name,
            'component_id', sc.component_id,
            'rendered_html', sc.rendered_html,
            'content_data', sc.content_data,
            'build_status', sc.build_status,
            'locked_at', sc.locked_at,
            'locked_by', sc.locked_by
        )
    ), '[]'::jsonb)
    INTO v_components
    FROM site_components sc
    WHERE sc.site_id = p_site_id;

    INSERT INTO site_snapshots (
        site_id, trigger, git_commit_sha, label,
        site_record, spec_snapshot, pages_snapshot, nav_snapshot,
        components_snapshot, spec_ids, created_by
    ) VALUES (
        p_site_id, p_trigger, p_git_sha, p_label,
        v_site_record, v_spec_snapshot, v_pages, v_nav,
        v_components, v_spec_ids, p_created_by
    ) RETURNING id INTO v_snapshot_id;

    RAISE NOTICE 'Snapshot % created for site % (trigger: %, specs: %, pages: %)',
        v_snapshot_id, p_site_id, p_trigger,
        jsonb_array_length(v_spec_snapshot),
        jsonb_array_length(v_pages);

    RETURN v_snapshot_id;
END;
$function$

;

CREATE OR REPLACE FUNCTION public.revert_site_to_snapshot(p_snapshot_id uuid, p_reverted_by text DEFAULT 'admin'::text)
 RETURNS jsonb
 LANGUAGE plpgsql
AS $function$
DECLARE
    v_snap          RECORD;
    v_safety_id     UUID;
    v_spec          JSONB;
    v_page          JSONB;
    v_page_id       UUID;
    v_comp          JSONB;
    v_group         JSONB;
    v_item          JSONB;
    v_sc            JSONB;
    v_pages_restored INT := 0;
    v_specs_restored INT := 0;
    v_comps_restored INT := 0;
BEGIN
    SELECT * INTO v_snap FROM site_snapshots WHERE id = p_snapshot_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Snapshot % not found', p_snapshot_id;
    END IF;

    v_safety_id := take_site_snapshot(
        v_snap.site_id, 'pre_revert', NULL,
        'Auto-snapshot before revert to ' || p_snapshot_id::text,
        p_reverted_by
    );

    -- ── 1. Restore site_specs ──────────────────────────────────────────

    UPDATE site_specs
    SET is_current = false, superseded_at = NOW()
    WHERE site_id = v_snap.site_id AND is_current = true;

    FOR v_spec IN SELECT * FROM jsonb_array_elements(v_snap.spec_snapshot)
    LOOP
        INSERT INTO site_specs (
            site_id, aspect, data, source, source_agent,
            created_by, pinned, is_current
        ) VALUES (
            v_snap.site_id,
            v_spec->>'aspect',
            v_spec->'data',
            'snapshot_revert',
            v_spec->>'source_agent',
            p_reverted_by,
            COALESCE((v_spec->>'pinned')::boolean, false),
            true
        );
        v_specs_restored := v_specs_restored + 1;
    END LOOP;

    -- ── 2. Restore pages and page_components ───────────────────────────

    DELETE FROM page_components
    WHERE page_id IN (SELECT id FROM pages WHERE site_id = v_snap.site_id);

    DELETE FROM pages WHERE site_id = v_snap.site_id;

    FOR v_page IN SELECT * FROM jsonb_array_elements(v_snap.pages_snapshot)
    LOOP
        INSERT INTO pages (
            id, site_id, name, url, title, page_type, status,
            meta_description, topics, nav_label, nav_order,
            in_header, in_footer, build_status, version,
            sections, rendered_header, rendered_footer, rendered_head,
            page_spec, content_direction, site_area_id
        ) VALUES (
            (v_page->>'id')::uuid,
            v_snap.site_id,
            v_page->>'name',
            v_page->>'url',
            v_page->>'title',
            v_page->>'page_type',
            v_page->>'status',
            v_page->>'meta_description',
            CASE WHEN v_page->'topics' IS NOT NULL AND v_page->'topics' != 'null'::jsonb
                THEN ARRAY(SELECT jsonb_array_elements_text(v_page->'topics'))
                ELSE NULL
            END,
            v_page->>'nav_label',
            COALESCE((v_page->>'nav_order')::int, 100),
            COALESCE((v_page->>'in_header')::boolean, true),
            COALESCE((v_page->>'in_footer')::boolean, true),
            COALESCE(v_page->>'build_status', 'deployed'),
            COALESCE((v_page->>'version')::int, 1),
            COALESCE(v_page->'sections', '[]'::jsonb),
            v_page->>'rendered_header',
            v_page->>'rendered_footer',
            v_page->>'rendered_head',
            v_page->'page_spec',
            v_page->'content_direction',
            CASE WHEN v_page->>'site_area_id' IS NOT NULL
                THEN (v_page->>'site_area_id')::uuid
                ELSE NULL
            END
        );

        v_page_id := (v_page->>'id')::uuid;

        FOR v_comp IN SELECT * FROM jsonb_array_elements(v_page->'components')
        LOOP
            INSERT INTO page_components (
                page_id, component_id, position, slot_name,
                rendered_html, content_data, build_status
            ) VALUES (
                v_page_id,
                CASE WHEN v_comp->>'component_id' IS NOT NULL
                    THEN (v_comp->>'component_id')::uuid
                    ELSE NULL
                END,
                COALESCE((v_comp->>'position')::int, 0),
                v_comp->>'slot_name',
                v_comp->>'rendered_html',
                COALESCE(v_comp->'content_data', '{}'::jsonb),
                COALESCE(v_comp->>'build_status', 'deployed')
            );
            v_comps_restored := v_comps_restored + 1;
        END LOOP;

        v_pages_restored := v_pages_restored + 1;
    END LOOP;

    -- ── 3. Restore navigation ──────────────────────────────────────────

    DELETE FROM site_nav_items WHERE site_id = v_snap.site_id;
    DELETE FROM site_nav_groups WHERE site_id = v_snap.site_id;

    FOR v_group IN SELECT * FROM jsonb_array_elements(v_snap.nav_snapshot->'groups')
    LOOP
        INSERT INTO site_nav_groups (
            id, site_id, group_key, group_label, group_type,
            parent_group_id, position
        ) VALUES (
            (v_group->>'id')::uuid,
            v_snap.site_id,
            v_group->>'group_key',
            COALESCE(v_group->>'group_label', ''),
            v_group->>'group_type',
            CASE WHEN v_group->>'parent_group_id' IS NOT NULL
                THEN (v_group->>'parent_group_id')::uuid
                ELSE NULL
            END,
            COALESCE((v_group->>'position')::int, 0)
        );
    END LOOP;

    FOR v_item IN SELECT * FROM jsonb_array_elements(v_snap.nav_snapshot->'items')
    LOOP
        INSERT INTO site_nav_items (
            id, site_id, group_id, parent_item_id,
            label, url, page_id, item_type,
            position, status, metadata
        ) VALUES (
            (v_item->>'id')::uuid,
            v_snap.site_id,
            (v_item->>'group_id')::uuid,
            CASE WHEN v_item->>'parent_item_id' IS NOT NULL
                THEN (v_item->>'parent_item_id')::uuid
                ELSE NULL
            END,
            v_item->>'label',
            v_item->>'url',
            CASE WHEN v_item->>'page_id' IS NOT NULL
                THEN (v_item->>'page_id')::uuid
                ELSE NULL
            END,
            COALESCE(v_item->>'item_type', 'page_link'),
            COALESCE((v_item->>'position')::int, 0),
            COALESCE(v_item->>'status', 'active'),
            COALESCE(v_item->'metadata', '{}'::jsonb)
        );
    END LOOP;

    -- ── 4. Restore site_components ─────────────────────────────────────

    DELETE FROM site_components WHERE site_id = v_snap.site_id;

    FOR v_sc IN SELECT * FROM jsonb_array_elements(v_snap.components_snapshot)
    LOOP
        INSERT INTO site_components (
            id, site_id, slot_name, component_id,
            rendered_html, content_data, build_status
        ) VALUES (
            (v_sc->>'id')::uuid,
            v_snap.site_id,
            v_sc->>'slot_name',
            CASE WHEN v_sc->>'component_id' IS NOT NULL
                THEN (v_sc->>'component_id')::uuid
                ELSE NULL
            END,
            v_sc->>'rendered_html',
            COALESCE(v_sc->'content_data', '{}'::jsonb),
            COALESCE(v_sc->>'build_status', 'pending')
        );
    END LOOP;

    -- ── 5. Restore site record fields ──────────────────────────────────

    UPDATE sites SET
        status = COALESCE(v_snap.site_record->>'status', status),
        schema_mode = COALESCE(v_snap.site_record->>'schema_mode', schema_mode),
        default_components = COALESCE(v_snap.site_record->'default_components', default_components),
        updated_at = NOW()
    WHERE id = v_snap.site_id;

    RETURN jsonb_build_object(
        'reverted', true,
        'snapshot_id', p_snapshot_id,
        'safety_snapshot_id', v_safety_id,
        'site_id', v_snap.site_id,
        'specs_restored', v_specs_restored,
        'pages_restored', v_pages_restored,
        'components_restored', v_comps_restored,
        'snapshot_trigger', v_snap.trigger,
        'snapshot_created_at', v_snap.created_at
    );
END;
$function$

;
```
