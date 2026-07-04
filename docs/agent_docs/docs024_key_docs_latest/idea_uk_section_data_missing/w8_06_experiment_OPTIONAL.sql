-- W8 step 6 — OPTIONAL behavioural experiment. Run ONLY if pasting the applied Edit-B
-- block is awkward; the code paste is the preferred witness.
-- Mechanism: prefix the three icon rows' scope_ref ('index:4' → 'xindex:4' — still
-- contains ':' so the CHECK holds) so Edit B's 'index:%' match returns ONLY the
-- illustration row. Rebuild index. has_image flips t ⇒ first-row-only/early-exit
-- variant CONFIRMED behaviourally. Then RESTORE. Nothing else consumes these rows
-- (info-card-grid's icons are LLM svg — 4.3), so the toggle is inert beyond the test.

-- 6.1 TOGGLE (guarded, reversible):
UPDATE site_plan_imagery spi
SET scope_ref = 'x' || scope_ref
FROM site_plans sp
WHERE sp.id = spi.plan_id AND sp.is_current = true
  AND sp.site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND spi.scope = 'section' AND spi.scope_ref = 'index:4' AND spi.kind = 'icon'
RETURNING spi.key, spi.scope_ref;   -- expect three rows, all 'xindex:4'

-- 6.2 REBUILD index (same shape as w8; dedup passes — prior keys complete):
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary,
                             spec, priority, handler_agent, status, created_by, item_key, page_id)
SELECT p.site_id, 'manual', 'build', 'needs_page', 'medium',
       'Edit-B first-row experiment: index', 
       jsonb_build_object('reason','editb_experiment','page_name',p.name),
       99, 'page-build-handler', 'triaged', 'w8_experiment',
       'page_rerender:' || p.name, p.id
FROM pages p
WHERE p.site_id = (SELECT id FROM sites WHERE domain='idea.uk') AND p.name = 'index'
  AND NOT EXISTS (SELECT 1 FROM site_work_items w
    WHERE w.site_id = p.site_id AND w.item_key = 'page_rerender:' || p.name
      AND w.status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved'))
RETURNING item_key;

-- 6.3 After it completes: has_image on index (t ⇒ variant confirmed):
SELECT p.name, pc.updated_at, (pc.rendered_html LIKE '%illustration_%') AS has_image
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND p.name = 'index' AND pc.slot_name = 'brief-explanation';

-- 6.4 RESTORE (run regardless of outcome):
UPDATE site_plan_imagery spi
SET scope_ref = substr(scope_ref, 2)
FROM site_plans sp
WHERE sp.id = spi.plan_id AND sp.is_current = true
  AND sp.site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND spi.scope = 'section' AND spi.scope_ref = 'xindex:4' AND spi.kind = 'icon'
RETURNING spi.key, spi.scope_ref;   -- expect three rows, all back to 'index:4'
