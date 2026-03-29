-- 085_site_snapshots.sql
-- Site-level snapshots for full site revert capability
--
-- Captures the complete state of a site at a point in time:
--   - site record fields
--   - all current site_specs aspects
--   - all pages with their page_components (content_data + rendered_html)
--   - navigation structure (nav_groups + nav_items)
--   - site_components
--
-- Snapshots are self-contained JSONB blobs so they survive row deletions.

-- ============================================================================
-- TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS site_snapshots (
                                              id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id         UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    trigger         TEXT NOT NULL,              -- 'deploy', 'manual', 'pre_edit', 'scheduled'
    label           TEXT,                       -- optional: "pre-redesign", "v2 launch"
    git_commit_sha  TEXT,                       -- links to deployed file state in git

-- Full state capture as JSONB (self-contained)
    site_record     JSONB NOT NULL,             -- key fields from sites row
    spec_snapshot   JSONB NOT NULL,             -- all current site_specs aspects
    pages_snapshot  JSONB NOT NULL,             -- pages + page_components
    nav_snapshot    JSONB NOT NULL,             -- site_nav_groups + site_nav_items
    components_snapshot JSONB NOT NULL DEFAULT '[]'::jsonb,  -- site_components

-- Reference IDs for cross-referencing
    spec_ids        UUID[],                     -- site_specs.id values captured

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      TEXT NOT NULL               -- 'deploy-agent', 'admin', 'scheduler'
    );

CREATE INDEX IF NOT EXISTS idx_site_snapshots_site
    ON site_snapshots(site_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_site_snapshots_git
    ON site_snapshots(git_commit_sha) WHERE git_commit_sha IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_site_snapshots_trigger
    ON site_snapshots(trigger, created_at DESC);

COMMENT ON TABLE site_snapshots IS 'Full site state snapshots for point-in-time revert';

-- ============================================================================
-- FUNCTION: take_site_snapshot
-- ============================================================================
-- Captures the current state of a site into a single snapshot row.
-- Returns the snapshot ID.
--
-- Usage:
--   SELECT take_site_snapshot(
--       '961fdd76-4daf-435b-bdef-0ebc0252e911',
--       'deploy',
--       'abc123def',
--       'Post homepage redesign',
--       'deploy-agent'
--   );

CREATE OR REPLACE FUNCTION take_site_snapshot(
    p_site_id       UUID,
    p_trigger       TEXT,
    p_git_sha       TEXT DEFAULT NULL,
    p_label         TEXT DEFAULT NULL,
    p_created_by    TEXT DEFAULT 'system'
) RETURNS UUID AS $$
DECLARE
v_snapshot_id   UUID;
    v_site_record   JSONB;
    v_spec_snapshot JSONB;
    v_spec_ids      UUID[];
    v_pages         JSONB;
    v_nav           JSONB;
    v_components    JSONB;
BEGIN
    -- 1. Capture site record (key fields, not the entire row)
SELECT jsonb_build_object(
               'id', s.id,
               'domain', s.domain,
               'status', s.status,
               'company_name', s.company_name,
               'industry', s.industry,
               'schema_mode', s.schema_mode,
               'style_collection_id', s.style_collection_id,
               'default_components', s.default_components,
               'content_data', s.content_data,
               'last_built_at', s.last_built_at,
               'last_deployed_at', s.last_deployed_at
       ) INTO v_site_record
FROM sites s
WHERE s.id = p_site_id;

IF v_site_record IS NULL THEN
        RAISE EXCEPTION 'Site % not found', p_site_id;
END IF;

    -- 2. Capture all current site_specs aspects
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

-- 3. Capture pages with their page_components
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

-- 4. Capture navigation structure
SELECT jsonb_build_object(
               'groups', COALESCE(
                (SELECT jsonb_agg(
                                jsonb_build_object(
                                        'id', g.id,
                                        'name', g.name,
                                        'location', g.location,
                                        'sort_order', g.sort_order
                                ) ORDER BY g.sort_order
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
                                               'page_id', ni.page_id,
                                               'label', ni.label,
                                               'url', ni.url,
                                               'sort_order', ni.sort_order,
                                               'is_active', ni.is_active
                                       ) ORDER BY ni.sort_order
                               )
                        FROM site_nav_items ni
                        WHERE ni.site_id = p_site_id
                       ), '[]'::jsonb
                        )
       ) INTO v_nav;

-- 5. Capture site_components
SELECT COALESCE(jsonb_agg(
                        jsonb_build_object(
                                'id', sc.id,
                                'component_id', sc.component_id,
                                'role', sc.role,
                                'config', sc.config,
                                'is_active', sc.is_active
                        )
                ), '[]'::jsonb)
INTO v_components
FROM site_components sc
WHERE sc.site_id = p_site_id;

-- 6. Insert snapshot
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
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION take_site_snapshot IS 'Capture the complete current state of a site into a snapshot row';

-- ============================================================================
-- FUNCTION: revert_site_to_snapshot
-- ============================================================================
-- Restores a site to the state captured in a snapshot.
-- Takes a pre-revert snapshot first (trigger='pre_revert') for safety.
--
-- What it restores:
--   - site_specs: supersedes all current, inserts from snapshot
--   - pages: deletes current pages, re-inserts from snapshot
--   - page_components: re-inserted with pages
--   - navigation: replaces nav_groups and nav_items
--   - site_components: replaces
--   - site record fields: status, default_components, schema_mode
--
-- What it does NOT do:
--   - Git revert (caller must handle separately if deployed files need reverting)
--   - Restore content_components templates (those are global, not per-site)

CREATE OR REPLACE FUNCTION revert_site_to_snapshot(
    p_snapshot_id   UUID,
    p_reverted_by   TEXT DEFAULT 'admin'
) RETURNS JSONB AS $$
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
    -- Load snapshot
SELECT * INTO v_snap FROM site_snapshots WHERE id = p_snapshot_id;
IF NOT FOUND THEN
        RAISE EXCEPTION 'Snapshot % not found', p_snapshot_id;
END IF;

    -- Safety: take a pre-revert snapshot of current state
    v_safety_id := take_site_snapshot(
        v_snap.site_id, 'pre_revert', NULL,
        'Auto-snapshot before revert to ' || p_snapshot_id::text,
        p_reverted_by
    );

    RAISE NOTICE 'Safety snapshot % taken before revert', v_safety_id;

    -- ── 1. Restore site_specs ──────────────────────────────────────────

    -- Supersede all current specs
UPDATE site_specs
SET is_current = false, superseded_at = NOW()
WHERE site_id = v_snap.site_id AND is_current = true;

-- Insert specs from snapshot
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

    -- Delete current page_components (cascade from pages won't work for
    -- components with ON DELETE SET NULL, so delete explicitly)
DELETE FROM page_components
WHERE page_id IN (SELECT id FROM pages WHERE site_id = v_snap.site_id);

-- Delete current pages
DELETE FROM pages WHERE site_id = v_snap.site_id;

-- Re-insert pages from snapshot
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

        -- Re-insert page_components for this page
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
    INSERT INTO site_nav_groups (id, site_id, name, location, sort_order)
               VALUES (
                   (v_group->>'id')::uuid,
                   v_snap.site_id,
                   v_group->>'name',
                   v_group->>'location',
                   COALESCE((v_group->>'sort_order')::int, 0)
                   );
END LOOP;

FOR v_item IN SELECT * FROM jsonb_array_elements(v_snap.nav_snapshot->'items')
                                LOOP
    INSERT INTO site_nav_items (
    id, site_id, group_id, page_id, label, url, sort_order, is_active
) VALUES (
                  (v_item->>'id')::uuid,
                  v_snap.site_id,
                  CASE WHEN v_item->>'group_id' IS NOT NULL
                  THEN (v_item->>'group_id')::uuid ELSE NULL END,
                  CASE WHEN v_item->>'page_id' IS NOT NULL
                  THEN (v_item->>'page_id')::uuid ELSE NULL END,
                  v_item->>'label',
                  v_item->>'url',
                  COALESCE((v_item->>'sort_order')::int, 0),
                  COALESCE((v_item->>'is_active')::boolean, true)
                  );
END LOOP;

    -- ── 4. Restore site_components ─────────────────────────────────────

DELETE FROM site_components WHERE site_id = v_snap.site_id;

FOR v_sc IN SELECT * FROM jsonb_array_elements(v_snap.components_snapshot)
                              LOOP
    INSERT INTO site_components (id, site_id, component_id, role, config, is_active)
            VALUES (
                (v_sc->>'id')::uuid,
                v_snap.site_id,
                CASE WHEN v_sc->>'component_id' IS NOT NULL
                THEN (v_sc->>'component_id')::uuid ELSE NULL END,
                v_sc->>'role',
                COALESCE(v_sc->'config', '{}'::jsonb),
                COALESCE((v_sc->>'is_active')::boolean, true)
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
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION revert_site_to_snapshot IS 'Restore a site to a previous snapshot state. Takes a safety snapshot first.';

-- ============================================================================
-- VIEW: recent site snapshots
-- ============================================================================

CREATE OR REPLACE VIEW v_site_snapshots AS
SELECT
    ss.id,
    ss.site_id,
    s.domain,
    ss.trigger,
    ss.label,
    ss.git_commit_sha,
    jsonb_array_length(ss.spec_snapshot) AS spec_count,
    jsonb_array_length(ss.pages_snapshot) AS page_count,
    ss.created_at,
    ss.created_by
FROM site_snapshots ss
         JOIN sites s ON ss.site_id = s.id
ORDER BY ss.created_at DESC;

COMMENT ON VIEW v_site_snapshots IS 'Summary of site snapshots with domain and counts';

-- ============================================================================
-- COMPLETION
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Migration 085: site_snapshots table, take/revert functions created';
END $$;
---

-- 085_site_snapshots.sql
-- Site-level snapshots for full site revert capability
--
-- Captures the complete state of a site at a point in time:
--   - site record fields
--   - all current site_specs aspects
--   - all pages with their page_components (content_data + rendered_html)
--   - navigation structure (nav_groups + nav_items)
--   - site_components
--
-- Snapshots are self-contained JSONB blobs so they survive row deletions.

-- ============================================================================
-- TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS site_snapshots (
                                              id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id         UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    trigger         TEXT NOT NULL,              -- 'deploy', 'manual', 'pre_edit', 'scheduled'
    label           TEXT,                       -- optional: "pre-redesign", "v2 launch"
    git_commit_sha  TEXT,                       -- links to deployed file state in git

-- Full state capture as JSONB (self-contained)
    site_record     JSONB NOT NULL,             -- key fields from sites row
    spec_snapshot   JSONB NOT NULL,             -- all current site_specs aspects
    pages_snapshot  JSONB NOT NULL,             -- pages + page_components
    nav_snapshot    JSONB NOT NULL,             -- site_nav_groups + site_nav_items
    components_snapshot JSONB NOT NULL DEFAULT '[]'::jsonb,  -- site_components

-- Reference IDs for cross-referencing
    spec_ids        UUID[],                     -- site_specs.id values captured

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      TEXT NOT NULL               -- 'deploy-agent', 'admin', 'scheduler'
    );

CREATE INDEX IF NOT EXISTS idx_site_snapshots_site
    ON site_snapshots(site_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_site_snapshots_git
    ON site_snapshots(git_commit_sha) WHERE git_commit_sha IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_site_snapshots_trigger
    ON site_snapshots(trigger, created_at DESC);

COMMENT ON TABLE site_snapshots IS 'Full site state snapshots for point-in-time revert';

-- ============================================================================
-- FUNCTION: take_site_snapshot
-- ============================================================================
-- Captures the current state of a site into a single snapshot row.
-- Returns the snapshot ID.
--
-- Usage:
--   SELECT take_site_snapshot(
--       '961fdd76-4daf-435b-bdef-0ebc0252e911',
--       'deploy',
--       'abc123def',
--       'Post homepage redesign',
--       'deploy-agent'
--   );

CREATE OR REPLACE FUNCTION take_site_snapshot(
    p_site_id       UUID,
    p_trigger       TEXT,
    p_git_sha       TEXT DEFAULT NULL,
    p_label         TEXT DEFAULT NULL,
    p_created_by    TEXT DEFAULT 'system'
) RETURNS UUID AS $$
DECLARE
v_snapshot_id   UUID;
    v_site_record   JSONB;
    v_spec_snapshot JSONB;
    v_spec_ids      UUID[];
    v_pages         JSONB;
    v_nav           JSONB;
    v_components    JSONB;
BEGIN
    -- 1. Capture site record (key fields, not the entire row)
SELECT jsonb_build_object(
               'id', s.id,
               'domain', s.domain,
               'status', s.status,
               'company_name', s.company_name,
               'industry', s.industry,
               'schema_mode', s.schema_mode,
               'style_collection_id', s.style_collection_id,
               'default_components', s.default_components,
               'content_data', s.content_data,
               'last_built_at', s.last_built_at,
               'last_deployed_at', s.last_deployed_at
       ) INTO v_site_record
FROM sites s
WHERE s.id = p_site_id;

IF v_site_record IS NULL THEN
        RAISE EXCEPTION 'Site % not found', p_site_id;
END IF;

    -- 2. Capture all current site_specs aspects
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

-- 3. Capture pages with their page_components
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

-- 4. Capture navigation structure
SELECT jsonb_build_object(
               'groups', COALESCE(
                (SELECT jsonb_agg(
                                jsonb_build_object(
                                        'id', g.id,
                                        'name', g.name,
                                        'location', g.location,
                                        'sort_order', g.sort_order
                                ) ORDER BY g.sort_order
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
                                               'page_id', ni.page_id,
                                               'label', ni.label,
                                               'url', ni.url,
                                               'sort_order', ni.sort_order,
                                               'is_active', ni.is_active
                                       ) ORDER BY ni.sort_order
                               )
                        FROM site_nav_items ni
                        WHERE ni.site_id = p_site_id
                       ), '[]'::jsonb
                        )
       ) INTO v_nav;

-- 5. Capture site_components
SELECT COALESCE(jsonb_agg(
                        jsonb_build_object(
                                'id', sc.id,
                                'component_id', sc.component_id,
                                'role', sc.role,
                                'config', sc.config,
                                'is_active', sc.is_active
                        )
                ), '[]'::jsonb)
INTO v_components
FROM site_components sc
WHERE sc.site_id = p_site_id;

-- 6. Insert snapshot
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
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION take_site_snapshot IS 'Capture the complete current state of a site into a snapshot row';

-- ============================================================================
-- FUNCTION: revert_site_to_snapshot
-- ============================================================================
-- Restores a site to the state captured in a snapshot.
-- Takes a pre-revert snapshot first (trigger='pre_revert') for safety.
--
-- What it restores:
--   - site_specs: supersedes all current, inserts from snapshot
--   - pages: deletes current pages, re-inserts from snapshot
--   - page_components: re-inserted with pages
--   - navigation: replaces nav_groups and nav_items
--   - site_components: replaces
--   - site record fields: status, default_components, schema_mode
--
-- What it does NOT do:
--   - Git revert (caller must handle separately if deployed files need reverting)
--   - Restore content_components templates (those are global, not per-site)

CREATE OR REPLACE FUNCTION revert_site_to_snapshot(
    p_snapshot_id   UUID,
    p_reverted_by   TEXT DEFAULT 'admin'
) RETURNS JSONB AS $$
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
    -- Load snapshot
SELECT * INTO v_snap FROM site_snapshots WHERE id = p_snapshot_id;
IF NOT FOUND THEN
        RAISE EXCEPTION 'Snapshot % not found', p_snapshot_id;
END IF;

    -- Safety: take a pre-revert snapshot of current state
    v_safety_id := take_site_snapshot(
        v_snap.site_id, 'pre_revert', NULL,
        'Auto-snapshot before revert to ' || p_snapshot_id::text,
        p_reverted_by
    );

    RAISE NOTICE 'Safety snapshot % taken before revert', v_safety_id;

    -- ── 1. Restore site_specs ──────────────────────────────────────────

    -- Supersede all current specs
UPDATE site_specs
SET is_current = false, superseded_at = NOW()
WHERE site_id = v_snap.site_id AND is_current = true;

-- Insert specs from snapshot
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

    -- Delete current page_components (cascade from pages won't work for
    -- components with ON DELETE SET NULL, so delete explicitly)
DELETE FROM page_components
WHERE page_id IN (SELECT id FROM pages WHERE site_id = v_snap.site_id);

-- Delete current pages
DELETE FROM pages WHERE site_id = v_snap.site_id;

-- Re-insert pages from snapshot
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

        -- Re-insert page_components for this page
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
    INSERT INTO site_nav_groups (id, site_id, name, location, sort_order)
               VALUES (
                   (v_group->>'id')::uuid,
                   v_snap.site_id,
                   v_group->>'name',
                   v_group->>'location',
                   COALESCE((v_group->>'sort_order')::int, 0)
                   );
END LOOP;

FOR v_item IN SELECT * FROM jsonb_array_elements(v_snap.nav_snapshot->'items')
                                LOOP
    INSERT INTO site_nav_items (
    id, site_id, group_id, page_id, label, url, sort_order, is_active
) VALUES (
                  (v_item->>'id')::uuid,
                  v_snap.site_id,
                  CASE WHEN v_item->>'group_id' IS NOT NULL
                  THEN (v_item->>'group_id')::uuid ELSE NULL END,
                  CASE WHEN v_item->>'page_id' IS NOT NULL
                  THEN (v_item->>'page_id')::uuid ELSE NULL END,
                  v_item->>'label',
                  v_item->>'url',
                  COALESCE((v_item->>'sort_order')::int, 0),
                  COALESCE((v_item->>'is_active')::boolean, true)
                  );
END LOOP;

    -- ── 4. Restore site_components ─────────────────────────────────────

DELETE FROM site_components WHERE site_id = v_snap.site_id;

FOR v_sc IN SELECT * FROM jsonb_array_elements(v_snap.components_snapshot)
                              LOOP
    INSERT INTO site_components (id, site_id, component_id, role, config, is_active)
            VALUES (
                (v_sc->>'id')::uuid,
                v_snap.site_id,
                CASE WHEN v_sc->>'component_id' IS NOT NULL
                THEN (v_sc->>'component_id')::uuid ELSE NULL END,
                v_sc->>'role',
                COALESCE(v_sc->'config', '{}'::jsonb),
                COALESCE((v_sc->>'is_active')::boolean, true)
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
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION revert_site_to_snapshot IS 'Restore a site to a previous snapshot state. Takes a safety snapshot first.';

-- ============================================================================
-- VIEW: recent site snapshots
-- ============================================================================

CREATE OR REPLACE VIEW v_site_snapshots AS
SELECT
    ss.id,
    ss.site_id,
    s.domain,
    ss.trigger,
    ss.label,
    ss.git_commit_sha,
    jsonb_array_length(ss.spec_snapshot) AS spec_count,
    jsonb_array_length(ss.pages_snapshot) AS page_count,
    ss.created_at,
    ss.created_by
FROM site_snapshots ss
         JOIN sites s ON ss.site_id = s.id
ORDER BY ss.created_at DESC;

COMMENT ON VIEW v_site_snapshots IS 'Summary of site snapshots with domain and counts';

-- ============================================================================
-- COMPLETION
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Migration 085: site_snapshots table, take/revert functions created';
END $$;