-- SQL_2026-08-23_seed_rfc032_componentid_conversions.sql
--
-- RFC_032 §8 (owner ruling 2026-08-22): converge on {{.InstanceID}} and retire
-- {{.ComponentID}}. This seeds the conversion of the SECTION templates that
-- still spell the retired placeholder, through the same fixer and the same
-- acceptance gate the 124-row programme used. Lane dir, NOT sql_for_agents/, so
-- no migration runner ever sweeps it up (the 2026-08-17 precedent's rule).
--
-- WHY THIS COULD NOT BE RUN BEFORE 2026-08-23. The converter harvests declared
-- ids with id="([^"{}]+)", a class that excludes braces, so a template whose
-- only id is id="{{.ComponentID}}" produced an EMPTY harvest and refused with
-- "template declares no literal element ids" — the opposite of the truth. These
-- five items would have completed as five polite no-ops: a queue that drains
-- green with nothing changed and no reason for anyone to look. Pass 0 (commit
-- 67d34e6c1, council cd6a5ef6 APPROVED) fixes that, and it is LIVE: chassis
-- v1.0.1328, pods started 2026-08-23T11:51Z, verified by probing the running
-- binary for the string that commit introduced (grep -ac on /proc/1/exe -> 1)
-- with a present-string control (1) and an absent-string control (0).
--
-- ⚠ DO NOT APPLY AGAINST A CHASSIS OLDER THAN v1.0.1328. Re-run that probe
-- rather than trusting this comment: the tag is not the evidence, the binary is.
--
-- WHAT THE CONVERSION DOES, per row: one id attribute whose whole value was
-- {{.ComponentID}} becomes {{.InstanceID}} (the BARE token — the wrapper id IS
-- the instance identity, so the {{.InstanceID}}- prefix form would render a
-- trailing hyphen with nothing after it). Measured 2026-08-23, all five rows are
-- the same shape: exactly one ComponentID occurrence, it IS a well-formed id
-- attribute, no <script>, no getElementById, exactly one id attribute in total.
-- So templated_id_swaps=1 is the ONLY non-zero counter the report will carry —
-- which is why that key was added to the fixer's result map in the same round
-- (without it a real conversion is indistinguishable from a no-op at the item).
--
-- POPULATION, derived here rather than pasted, because a census goes stale by
-- ADDITION (owner ruling 2026-08-22). Counts as of 2026-08-23:
--   generic-text-block   179 placements / 152 pages / 21 sites   (row 8d81e665)
--   faq                   82 placements /  82 pages / 15 sites   (row d91e7be1)
--   mechanism-flow         6 placements /   6 pages /  3 sites   (row fa5e4524)
--   evidence-timeseries    3 placements /   3 pages /  3 sites   (row fb870e82)
-- and 18 pages carry a repeated one (27 redundant placements) — all
-- generic-text-block, apis.uk/index.html carrying six.
--
-- ⚠ ONE ROW IS DELIBERATELY NOT SEEDED HERE: `pricing` (row 6175e049), active,
-- carrying the same placeholder, with ZERO placements. site_work_items.site_id
-- is NOT NULL and the site is only reachable through a placement, so there is no
-- honest site to file it against — inventing one would put a false site
-- relationship in the queue to satisfy a constraint. It is inert today (nothing
-- renders it), but it is a PRECONDITION OF DELETING THE BINDINGS: retire
-- {{.ComponentID}} from the render paths while this row still spells it and its
-- first placement renders id="". Recorded in the lane NOTES as the named
-- residual of this step.
--
-- DRY RUN (expect 4; re-run at apply time and expect it to have MOVED):
--   SELECT count(DISTINCT c.id) FROM content_components c
--     JOIN page_components pc ON pc.component_id=c.id
--   WHERE c.is_active AND c.html_template LIKE '%ComponentID%'
--     AND c.html_template NOT LIKE '%{{.InstanceID}}%';
--
-- VERIFY after applying:
--   SELECT item_key, status FROM site_work_items
--    WHERE item_key LIKE 'instance-scope:%' AND created_by='283-rfc032-seed';
--   -- then watch the corpus, which is the thing that actually matters:
--   SELECT function, html_template LIKE '%{{.InstanceID}}%' AS converted
--     FROM content_components WHERE is_active AND html_template LIKE '%ComponentID%';
--   -- and the result the fixer wrote — a completed item with no swap is a no-op:
--   SELECT item_key, result->>'templated_id_swaps' FROM site_work_items
--    WHERE created_by='283-rfc032-seed';

INSERT INTO site_work_items (
  site_id, source, pipeline, item_type, severity, summary,
  priority, handler_agent, status, created_by, spec, item_key, batch_id
)
SELECT DISTINCT ON (c.id)
  s.id,
  'manual', 'build', 'instance_scope_conversion', 'low',
  'RFC_032 step 2: convert ' || c.function || ' (row ' || left(c.id::text, 8)
    || ') from the retired {{.ComponentID}} to {{.InstanceID}} — one templated id, no script',
  35,
  'component-template-fixer', 'triaged', '283-rfc032-seed',
  jsonb_build_object(
    'fix_type', 'scope_component_instance',
    'component_id', c.id::text,
    'category', 'seam',
    'note', 'RFC_032 section 8 owner ruling; pass 0 live on chassis v1.0.1328 (binary-probed). Expect templated_id_swaps=1 and every other counter 0.'
  ),
  'instance-scope:' || left(c.id::text, 8),
  gen_random_uuid()
FROM content_components c
JOIN page_components pc ON pc.component_id = c.id
JOIN pages p  ON p.id = pc.page_id
JOIN sites s  ON s.id = p.site_id
WHERE c.is_active
  AND c.html_template LIKE '%ComponentID%'
  AND c.html_template NOT LIKE '%{{.InstanceID}}%'   -- idempotency
ORDER BY c.id, s.id
ON CONFLICT DO NOTHING;
