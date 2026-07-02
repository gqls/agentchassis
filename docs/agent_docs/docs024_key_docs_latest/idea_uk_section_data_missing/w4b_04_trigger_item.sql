-- W4b step 4: the trigger item — crafted from the real rows (pipeline build, severity
-- medium, priority 99, handler rerender-pages, status triaged) with truthful deviations
-- noted: source 'manual' and created_by 'w4b_chrome_refresh' (the real rows say
-- component-creator / store_generated_component because they ARE component regens; ours
-- is not, and lying in provenance columns costs later debugging). spec carries ONLY the
-- gate flag + a reason — function/component_id are consumed nowhere in the v6 workflow.
-- Schema first, then check-first dedup (CreateNeedsNewComponentItem pattern), then insert.
\d site_work_items

SELECT EXISTS(
  SELECT 1 FROM site_work_items w
  WHERE w.site_id = (SELECT id FROM sites WHERE domain = 'idea.uk')
    AND w.item_key = 'chrome_refresh_rerender:' ||
        (SELECT id::text FROM sites WHERE domain = 'idea.uk')
    AND w.status NOT IN ('complete','verified','rejected','wont_fix','failed')
) AS already_queued;   -- expect f; if t, the item is already pending — do not insert again

INSERT INTO site_work_items (
  site_id, source, pipeline, item_type, severity, summary,
  spec, priority, handler_agent, status, created_by, item_key
)
SELECT s.id, 'manual', 'build', 'needs_rerender', 'medium',
       'Refresh site chrome (header/footer/head) after component repoint; rerender pages',
       '{"reason": "chrome_repoint_refresh", "refresh_site_components": true}'::jsonb,
       99, 'rerender-pages', 'triaged', 'w4b_chrome_refresh',
       'chrome_refresh_rerender:' || s.id::text
FROM sites s
WHERE s.domain = 'idea.uk'
  AND NOT EXISTS (
    SELECT 1 FROM site_work_items w
    WHERE w.site_id = s.id
      AND w.item_key = 'chrome_refresh_rerender:' || s.id::text
      AND w.status NOT IN ('complete','verified','rejected','wont_fix','failed')
  )
RETURNING id, item_key, status, handler_agent, priority;
-- Expect: INSERT 0 1 with the returned row; the build dispatch loop claims triaged items.
