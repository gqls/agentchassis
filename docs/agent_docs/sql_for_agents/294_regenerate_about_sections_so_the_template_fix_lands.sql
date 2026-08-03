-- 294 — queue a SECTION-REGENERATING rerender for finetuning.uk /about.html
--
-- WHAT: one `page_rerender` work item carrying `spec.reason='section_data_resolved'`
-- for page `about` (c0c68034-469f-420c-90bd-d3c0fc0e13d2), status 'triaged' so
-- build-pipeline-trigger claims it on its next 120s tick.
--
-- WHY IT IS NEEDED AT ALL, which is the whole point of this file. Migration 293
-- fixed the `departments-grid` template at source and it was live immediately
-- (DB config). The site already had 68 `page_rerender` items queued. It would be
-- natural — and wrong — to assume the fix propagates when that queue drains.
--
-- The live `page-rerender` agent routes on `spec.reason` ALONE
-- (`check_rerender_mode`):
--
--     reason IN ('image_landed','section_data_resolved','cta_links_stale')
--         -> rerender_page_sections   REGENERATES each section from html_template
--     otherwise
--         -> rerender_single_page     re-staples sections from html ALREADY STORED
--
-- `rerender_single_page_action.go` says so in its own header, line 4:
-- "Simple concatenation - no template re-rendering". A reason-less item completes,
-- reports success, and faithfully preserves whatever was wrong.
--
-- MEASURED ON THIS SITE, 2026-08-03, before writing this file:
--
--     reason                            count
--     (NULL — re-staples STORED html)      42
--     cta_links_stale                      26
--
-- and for the two pages that actually mount `departments-grid`:
--
--     /index.html   cta_links_stale   <- regenerates; the fix WILL land here
--     /index.html   (NULL)
--     /about.html   (NULL)            <- re-staples; the fix would NEVER land
--
-- So /index.html is already covered by an item someone else's check queued, and
-- /about.html — eleven of the nineteen broken images — is not. This file covers it.
--
-- THE PRECEDENT THIS AVOIDS REPEATING. bugs_open/140, 2026-08-02: a shared
-- component template was fixed at source and live the whole time; the
-- `page_rerender` backlog then drained 294 -> 0 and SIX affected pages ran a
-- COMPLETED rerender with not one section updated (one still stamped 2026-05-02).
-- Seven items carrying an explicit reason repaired all seven in twenty minutes.
-- A completed work item is not a repaired artefact.
--
-- WHY `section_data_resolved` AND NOT `cta_links_stale`: both take the
-- regenerating branch, and this one is the narrowest honest description of what
-- happened — a stored render is stale against its source. Blast radius is one
-- page's sections. `component_id` is deliberately NOT set: the agent requires only
-- the reason, and the `scoped := reason != "" && componentIDStr != ""` line in
-- create_rerender_items_action.go is a PRODUCER-side gate for that creator, not
-- the consumer's rule. Leaving it unset re-renders every section on the page,
-- which is what we want — the template fix is not the only thing stale here.
--
-- ITEM_KEY and the dedup index. idx_swi_dedup is UNIQUE on (site_id, item_key)
-- for any status outside the terminal set, so a distinct key is required or this
-- collides with the existing reason-less item for the same page. Key is
-- 'page_rerender:about:deptgrid-template-293' — it names the CAUSE, so a reader
-- who finds it later can tell why a second rerender was queued for one page.
--
-- APPLY:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < 294_….sql

BEGIN;

INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary, spec,
    page_id, priority, handler_agent, status, created_by, item_key
) VALUES (
    '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
    'manual',
    'build',
    'page_rerender',
    'high',
    'Regenerate about sections — departments-grid template fixed by 293, stored render is stale',
    jsonb_build_object(
        'domain',    'finetuning.uk',
        'page_id',   'c0c68034-469f-420c-90bd-d3c0fc0e13d2',
        'page_name', 'about',
        'filename',  'about.html',
        -- THE LOAD-BEARING FIELD. Without it this item takes rerender_single_page
        -- and re-staples the stale HTML it is supposed to be replacing.
        'reason',    'section_data_resolved'
    ),
    'c0c68034-469f-420c-90bd-d3c0fc0e13d2',
    20,                     -- ahead of the reason-less backlog (default 100)
    'page-rerender',
    'triaged',              -- directly claimable; triage already ran this cycle
    'finetuning-repair-293',
    'page_rerender:about:deptgrid-template-293'
);

-- VERIFY as DO/RAISE: a block of SELECTs cannot stop a COMMIT, because
-- ON_ERROR_STOP does not fire on a non-empty result set.
DO $$
DECLARE
    r record;
BEGIN
    SELECT spec->>'reason' AS reason, status, handler_agent, page_id INTO r
    FROM site_work_items
    WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
      AND item_key = 'page_rerender:about:deptgrid-template-293';

    IF r IS NULL THEN
        RAISE EXCEPTION '294: the item was not inserted';
    END IF;
    IF r.reason IS DISTINCT FROM 'section_data_resolved' THEN
        RAISE EXCEPTION '294: spec.reason is %, so this item would RE-STAPLE stale html', r.reason;
    END IF;
    IF r.status <> 'triaged' THEN
        RAISE EXCEPTION '294: status is %, and the dispatcher only claims triaged/approved', r.status;
    END IF;
    IF r.handler_agent <> 'page-rerender' THEN
        RAISE EXCEPTION '294: handler_agent is %', r.handler_agent;
    END IF;
    IF r.page_id IS NULL THEN
        RAISE EXCEPTION '294: page_id is NULL — the handler resolves the page from it';
    END IF;

    RAISE NOTICE '294 OK: about queued for SECTION regeneration (reason=section_data_resolved)';
END $$;

COMMIT;
