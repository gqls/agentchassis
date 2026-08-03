-- 295 — queue a SECTION-REGENERATING rerender for finetuning.uk /index.html
--
-- Sibling of 294, which did the same for /about.html. Same reasoning, and 294's
-- header carries the full argument; this file records only what differs.
--
-- WHAT DIFFERS. /index.html was NOT in the same danger as /about.html. It already
-- carries `misdirected_cta:index:…` with `spec.reason='cta_links_stale'`, which
-- IS one of the three reasons that take the section-regenerating branch — so the
-- template fix from 293 would have landed there eventually with no help from
-- anyone. This file is about WHEN, not WHETHER.
--
-- WHY IT IS WORTH A FILE ANYWAY. Measured 2026-08-03 10:35, after the
-- improvement-loop promoted 128 items:
--
--     triaged 121 · complete 58, draining at roughly one item per minute
--     misdirected_cta:index:… → status triaged, priority 35, claimed_at NULL
--
-- At that rate the homepage waits about two hours. /index.html is the page the
-- owner is actually looking at and carries 8 of the 19 broken images, so it is
-- the one page where the queue position is worth overriding. Priority 10 puts it
-- ahead of everything currently queued (the front of the queue was priority 30).
--
-- WHY NOT JUST RAISE THE EXISTING ITEM'S PRIORITY: that row belongs to
-- check_misdirected_cta, which set 35 deliberately for its own repair. Mutating
-- another creator's item to borrow its slot makes this session's intent invisible
-- and would leave a reader wondering why a CTA finding jumped the queue. A
-- separate item with its own key says who queued it and why. The CTA item stays
-- exactly as it was and does its own job in its own time; both regenerate
-- sections, and running twice is idempotent.
--
-- ITEM_KEY names the cause, as 294's does, so a reader meeting two rerenders for
-- one page can tell them apart: 'page_rerender:index:deptgrid-template-293'.
-- idx_swi_dedup is UNIQUE on (site_id, item_key) outside the terminal statuses,
-- so a distinct key is required and this one is distinct from both existing rows.
--
-- APPLY:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < 295_….sql

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
    'Regenerate index sections — departments-grid template fixed by 293, 8 broken images on the homepage',
    jsonb_build_object(
        'domain',    'finetuning.uk',
        'page_id',   'a716cacc-eec2-4aa6-a08b-7e6732506f41',
        'page_name', 'index',
        'filename',  'index.html',
        -- THE LOAD-BEARING FIELD (see 294 and the page_rerender landmine).
        'reason',    'section_data_resolved'
    ),
    'a716cacc-eec2-4aa6-a08b-7e6732506f41',
    10,                     -- ahead of the whole queue; front was 30
    'page-rerender',
    'triaged',
    'finetuning-repair-293',
    'page_rerender:index:deptgrid-template-293'
);

DO $$
DECLARE
    r record;
BEGIN
    SELECT spec->>'reason' AS reason, status, handler_agent, page_id, priority INTO r
    FROM site_work_items
    WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
      AND item_key = 'page_rerender:index:deptgrid-template-293';

    IF r IS NULL THEN
        RAISE EXCEPTION '295: the item was not inserted';
    END IF;
    IF r.reason IS DISTINCT FROM 'section_data_resolved' THEN
        RAISE EXCEPTION '295: spec.reason is %, so this would RE-STAPLE stale html', r.reason;
    END IF;
    IF r.status <> 'triaged' THEN
        RAISE EXCEPTION '295: status is %, and the dispatcher claims only triaged/approved', r.status;
    END IF;
    IF r.page_id IS NULL THEN
        RAISE EXCEPTION '295: page_id is NULL — the handler resolves the page from it';
    END IF;
    -- The whole point of this file is queue position; assert it rather than hope.
    IF r.priority >= (SELECT MIN(priority) FROM site_work_items
                      WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
                        AND status = 'triaged'
                        AND item_key <> 'page_rerender:index:deptgrid-template-293') THEN
        RAISE EXCEPTION '295: priority % does not lead the queue — the wait was the reason for this file', r.priority;
    END IF;

    RAISE NOTICE '295 OK: index queued at the FRONT for section regeneration (priority %)', r.priority;
END $$;

COMMIT;
