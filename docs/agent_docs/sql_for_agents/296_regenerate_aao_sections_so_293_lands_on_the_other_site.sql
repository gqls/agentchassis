-- 296 — apply 293 to the OTHER site that mounts departments-grid
--
-- WHAT: two `page_rerender` items, `spec.reason='section_data_resolved'`, status
-- 'triaged', for ai-agent-orchestration.com /index.html and /about.html.
--
-- WHY THIS SITE IS HERE AT ALL. 293 fixed `departments-grid` at source, and that
-- component is mounted by exactly two sites. The census that motivated 293:
--
--     ai-agent-orchestration.com   16 bare-token <img src>
--     finetuning.uk                15
--
-- finetuning.uk is now 0 in the database and 0 on both live pages. This site is
-- still 16, because a template fix changes nothing until the pages that mount it
-- are re-rendered — and nothing has re-rendered them.
--
-- WHY IT WILL NOT HAPPEN BY ITSELF, which is the reason this file exists rather
-- than a note saying "it will drain". Both pages ALREADY carry a `page_rerender`
-- with `reason='cta_links_stale'`, which is one of the three reasons that take the
-- section-regenerating branch. So the routing is right and the items are correct.
-- They are simply unreachable:
--
--     SELECT status, count(*) FROM site_work_items WHERE site_id='2a8ebf9c-…';
--     complete 129 · needs_human_review 58 · DETECTED 37 · unresolved 36
--     ...  triaged: 0
--
-- The dispatcher claims `status IN ('triaged','approved')`. Every open item on
-- this site is `detected`. The only detected→triaged promoter is `triage_findings`
-- inside improvement-loop, whose schedule has been off since 2026-05-02. This is
-- the same fleet-wide parking that left finetuning.uk untouched for months.
--
-- WHY NOT JUST FIRE THE IMPROVEMENT-LOOP HERE TOO. That would promote and dispatch
-- this site's whole 37-item backlog — content rewrites, design reviews, tool work —
-- against a site nobody asked about, as a side effect of a template repair. The
-- fleet-wide question (what to do about 204 parked items across 10 sites) is
-- explicitly the owner's, and is recorded in
-- `docs024_key_docs_latest/finetuning_uk_repair/PLAN_2026-08-03_…md`. So this file
-- does the NARROWEST thing that finishes the repair whose cause was changed: two
-- items, both rerenders, nothing else touched. Everything else on this site stays
-- exactly as parked as it was.
--
-- IDEMPOTENT AND NON-COMPETING: these regenerate the same sections the existing
-- `cta_links_stale` items would, so if that backlog is ever promoted, running both
-- is harmless. Distinct item_key, so idx_swi_dedup (UNIQUE on (site_id, item_key)
-- outside the terminal statuses) does not collide with the existing rows.
--
-- EXPECTED RESULT: the fleet census returns zero rows on both sites. Verify at the
-- served page, not the work item status:
--   curl -sS -L https://ai-agent-orchestration.com/ | grep -oE '<img[^>]*src="[^"/.]+"' | wc -l
--
-- KNOWN AND DELIBERATELY NOT FIXED HERE: eight of this site's icon values are not
-- lucide names at all — strategy, research, content, design, development, quality,
-- operations, data (all on /index.html). lucide leaves an unknown name untouched,
-- so those eight render as an empty badge rather than a broken image. That is an
-- improvement and not a repair, and the defect is in content_data, not the
-- template. Recorded in 293's header and in the lane's README.
--
-- APPLY:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < 296_….sql

BEGIN;

INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary, spec,
    page_id, priority, handler_agent, status, created_by, item_key
)
SELECT
    '2a8ebf9c-20a2-4c39-b191-840b012371da',
    'manual', 'build', 'page_rerender', 'high',
    'Regenerate ' || v.page_name || ' sections — departments-grid template fixed by 293, stored render is stale',
    jsonb_build_object(
        'domain',    'ai-agent-orchestration.com',
        'page_id',   v.page_id,
        'page_name', v.page_name,
        'filename',  v.page_name || '.html',
        'reason',    'section_data_resolved'   -- load-bearing; see 294 and the landmine
    ),
    v.page_id::uuid,
    10,
    'page-rerender',
    'triaged',
    'finetuning-repair-293',
    'page_rerender:' || v.page_name || ':deptgrid-template-293'
FROM (VALUES
    ('9baea9f9-4d4c-449d-949c-e5753d4faf67', 'index'),
    ('06f7b3bc-1269-4cf9-aef1-3481b855aeee', 'about')
) AS v(page_id, page_name);

DO $$
DECLARE
    n_ok int;
BEGIN
    SELECT count(*) INTO n_ok
    FROM site_work_items
    WHERE site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da'
      AND item_key IN ('page_rerender:index:deptgrid-template-293',
                       'page_rerender:about:deptgrid-template-293')
      AND spec->>'reason' = 'section_data_resolved'
      AND status = 'triaged'
      AND handler_agent = 'page-rerender'
      AND page_id IS NOT NULL;

    IF n_ok <> 2 THEN
        RAISE EXCEPTION '296: expected 2 correctly-shaped triaged items, found %', n_ok;
    END IF;

    RAISE NOTICE '296 OK: 2 section-regenerating rerenders queued for ai-agent-orchestration.com';
END $$;

COMMIT;
