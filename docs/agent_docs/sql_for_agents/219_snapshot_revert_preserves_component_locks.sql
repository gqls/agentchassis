-- 219_snapshot_revert_preserves_component_locks.sql
--
-- bugs_open/088 — a snapshot revert silently destroyed every component lock on
-- the site, on BOTH page_components and site_components.
--
-- WHAT WAS WRONG (read from the live function bodies via pg_get_functiondef on
-- 2026-07-26; the repo's sql_for_tables copies have drifted and disagree):
--
--   * take_site_snapshot captured site_components.locked_at/locked_by, but not
--     lock_type / lock_expires_at, and captured NO lock columns at all for
--     page_components.
--   * revert_site_to_snapshot DELETEs both tables for the site and re-INSERTs
--     from the snapshot with no lock columns in either INSERT — so every
--     locked_at / locked_by / lock_type / lock_expires_at became NULL,
--     including the two that WERE captured. The pre_revert safety snapshot
--     could not help: it has the same hole.
--
-- Those four columns are the only thing standing between a human-corrected
-- artefact and automation (bugs_closed/058 enforces them for page sections,
-- bugs_open/069 for chrome). A revert quietly returned every protected row to
-- agent-writable, and nothing recorded that it had happened. Exposure when
-- filed: 39 locked page_components rows live, 0 locked site_components rows,
-- 11 snapshots, most recent 2026-06-24.
--
-- WHAT THIS CHANGES — one rule: a revert restores CONTENT; it never locks or
-- unlocks anything.
--
--   1. take_site_snapshot captures all four lock columns on both tables, so a
--      snapshot is a true record and the pre_revert copy is worth having.
--   2. revert_site_to_snapshot reads the site's CURRENT lock state before the
--      deletes and re-applies it after the inserts. Deliberately NOT restoring
--      the snapshot's lock state: replaying as-captured would silently release
--      a lock added after the snapshot was taken, which is the defect class
--      being fixed. Content still comes from the snapshot — that is what the
--      human asked for, and a revert is a human-initiated surface, exempt in
--      the same way as the admin endpoints.
--   3. The result JSON reports page_locks_preserved / chrome_locks_preserved,
--      so a revert says what it carried across instead of being silent.
--
-- Keys for carrying a lock across: slot_name for chrome (unique per site), and
-- (page_id, slot_name) for page components — pages keep their ids across a
-- revert, so that key holds. A page with duplicate slot names re-locks every
-- match: conservative, locks more rather than fewer.
--
-- Both statements are CREATE OR REPLACE and the file is re-runnable.
--
-- ROLLBACK: re-apply the previous definitions, which are recorded verbatim in
--   the migration ledger's predecessor state — recover them with
--     SELECT pg_get_functiondef('take_site_snapshot(uuid,text,text,text,text)'::regprocedure);
--     SELECT pg_get_functiondef('revert_site_to_snapshot(uuid,text)'::regprocedure);
--   taken BEFORE this file runs (a copy of both, as at 2026-07-26, is in
--   docs/agent_docs/docs024_key_docs_latest/bugfix_069_chrome_locks/
--   RUNBOOK_chrome_lock_gate.md). Reverting reinstates the lock-wiping
--   behaviour; there is no data migration to undo.

BEGIN;

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
                        'build_status', pc.build_status,
                        -- bugs_open/088: the lock columns were never captured,
                        -- so a snapshot could not record what was protected.
                        'locked_at', pc.locked_at,
                        'locked_by', pc.locked_by,
                        'lock_type', pc.lock_type,
                        'lock_expires_at', pc.lock_expires_at
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
            'locked_by', sc.locked_by,
            -- bugs_open/088: locked_at/locked_by were captured but the
            -- classifier (lock_type) and its expiry were not, so the record
            -- could not say whether a captured lock was hard or timed.
            'lock_type', sc.lock_type,
            'lock_expires_at', sc.lock_expires_at
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
$function$;

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
    -- bugs_open/088: a revert restores CONTENT; it must never lock or unlock
    -- anything. These hold the site's lock state as it is NOW, read before the
    -- deletes and re-applied after the inserts.
    v_pc_locks      JSONB := '[]'::jsonb;
    v_sc_locks      JSONB := '[]'::jsonb;
    v_lock          JSONB;
    v_pc_locks_kept INT := 0;
    v_sc_locks_kept INT := 0;
    v_touched       INT := 0;
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

    -- bugs_open/088: capture the CURRENT locks before the delete. Restoring the
    -- snapshot's lock state instead would silently RELEASE any lock a human
    -- added after the snapshot was taken — the defect class this fixes. Only a
    -- human unlock releases a lock (031_LOCKS, and bugs_closed/058).
    SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'page_id',         pc.page_id,
        'slot_name',       COALESCE(pc.slot_name, ''),
        'locked_at',       pc.locked_at,
        'locked_by',       pc.locked_by,
        'lock_type',       pc.lock_type,
        'lock_expires_at', pc.lock_expires_at
    )), '[]'::jsonb)
    INTO v_pc_locks
    FROM page_components pc
    JOIN pages p ON p.id = pc.page_id
    WHERE p.site_id = v_snap.site_id
      AND pc.locked_at IS NOT NULL;

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

    -- bugs_open/088: re-apply the locks captured above. Keyed on
    -- (page_id, slot_name): pages keep their ids across the revert, so the key
    -- holds, but page_components are re-inserted with NEW ids and a page may
    -- carry duplicate slot names — so every matching row is re-locked.
    -- Conservative by design: it locks more, never fewer. A locked slot the
    -- snapshot does not contain matches nothing and is gone with its row; the
    -- pre_revert safety snapshot is where it survives.
    FOR v_lock IN SELECT * FROM jsonb_array_elements(v_pc_locks)
    LOOP
        UPDATE page_components pc SET
            locked_at       = (v_lock->>'locked_at')::timestamptz,
            locked_by       = v_lock->>'locked_by',
            lock_type       = v_lock->>'lock_type',
            lock_expires_at = CASE WHEN v_lock->>'lock_expires_at' IS NOT NULL
                                   THEN (v_lock->>'lock_expires_at')::timestamptz
                                   ELSE NULL END
        WHERE pc.page_id = (v_lock->>'page_id')::uuid
          AND COALESCE(pc.slot_name, '') = v_lock->>'slot_name';
        GET DIAGNOSTICS v_touched = ROW_COUNT;
        v_pc_locks_kept := v_pc_locks_kept + v_touched;
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

    -- bugs_open/088 / bugs_open/069: same rule for site chrome. Keyed on
    -- slot_name, which is unique per site.
    SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'slot_name',       sc.slot_name,
        'locked_at',       sc.locked_at,
        'locked_by',       sc.locked_by,
        'lock_type',       sc.lock_type,
        'lock_expires_at', sc.lock_expires_at
    )), '[]'::jsonb)
    INTO v_sc_locks
    FROM site_components sc
    WHERE sc.site_id = v_snap.site_id
      AND sc.locked_at IS NOT NULL;

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

    FOR v_lock IN SELECT * FROM jsonb_array_elements(v_sc_locks)
    LOOP
        UPDATE site_components sc SET
            locked_at       = (v_lock->>'locked_at')::timestamptz,
            locked_by       = v_lock->>'locked_by',
            lock_type       = v_lock->>'lock_type',
            lock_expires_at = CASE WHEN v_lock->>'lock_expires_at' IS NOT NULL
                                   THEN (v_lock->>'lock_expires_at')::timestamptz
                                   ELSE NULL END
        WHERE sc.site_id = v_snap.site_id
          AND sc.slot_name = v_lock->>'slot_name';
        GET DIAGNOSTICS v_touched = ROW_COUNT;
        v_sc_locks_kept := v_sc_locks_kept + v_touched;
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
        'page_locks_preserved', v_pc_locks_kept,
        'chrome_locks_preserved', v_sc_locks_kept,
        'snapshot_trigger', v_snap.trigger,
        'snapshot_created_at', v_snap.created_at
    );
END;
$function$;

-- ── post-conditions (inside the transaction: a failure rolls the file back) ──

DO $guard$
DECLARE
    v_take   TEXT := pg_get_functiondef('take_site_snapshot(uuid,text,text,text,text)'::regprocedure);
    v_revert TEXT := pg_get_functiondef('revert_site_to_snapshot(uuid,text)'::regprocedure);
BEGIN
    IF position('''lock_type'', pc.lock_type' IN v_take) = 0 THEN
        RAISE EXCEPTION '219: take_site_snapshot still does not capture page_components.lock_type';
    END IF;
    IF position('''lock_expires_at'', pc.lock_expires_at' IN v_take) = 0 THEN
        RAISE EXCEPTION '219: take_site_snapshot still does not capture page_components.lock_expires_at';
    END IF;
    IF position('''lock_type'', sc.lock_type' IN v_take) = 0 THEN
        RAISE EXCEPTION '219: take_site_snapshot still does not capture site_components.lock_type';
    END IF;
    IF position('''lock_expires_at'', sc.lock_expires_at' IN v_take) = 0 THEN
        RAISE EXCEPTION '219: take_site_snapshot still does not capture site_components.lock_expires_at';
    END IF;

    IF position('v_pc_locks' IN v_revert) = 0 OR position('v_sc_locks' IN v_revert) = 0 THEN
        RAISE EXCEPTION '219: revert_site_to_snapshot does not carry the current lock state';
    END IF;
    IF position('page_locks_preserved' IN v_revert) = 0
       OR position('chrome_locks_preserved' IN v_revert) = 0 THEN
        RAISE EXCEPTION '219: revert_site_to_snapshot does not report what it preserved';
    END IF;

    RAISE NOTICE '219 OK: snapshots capture all four lock columns on both tables; revert carries the current lock state across';
END
$guard$;

COMMIT;
