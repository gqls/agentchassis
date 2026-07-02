-- W4b step 1: repoint idea.uk's header + footer site_components rows to the ACTIVE
-- fixed components. REPOINT ONLY — rendered_html is deliberately left in place (stale
-- but serviceable), so the rerender path has no chrome-less window; the forced
-- re-render (trigger chosen from w4b_02's read) overwrites it. Head row untouched.
-- Guards on the KNOWN old ids => idempotent and cannot touch any other site/slot.
UPDATE site_components
SET component_id = 'f420f3fa-43a2-4a2f-b2e1-39770d45b494',  -- active generated site-header
    updated_at   = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'idea.uk')
  AND slot_name = 'header'
  AND component_id = '9644c86f-18b0-4f75-b086-5b79a74a48d7'  -- old gradient header (inactive)
RETURNING slot_name, component_id;

UPDATE site_components
SET component_id = '4238e467-25a6-4174-bee0-6fce914398c8',  -- active fixed site-footer (W2a)
    updated_at   = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'idea.uk')
  AND slot_name = 'footer'
  AND component_id = '09034086-a581-4bba-a5b4-760d863bb2df'  -- old footer-4-column (inactive)
RETURNING slot_name, component_id;

-- Post-check: the three rows, now with active header/footer and the head unchanged.
SELECT sc.slot_name, sc.component_id, cc.function, cc.is_active,
       length(sc.rendered_html) AS rendered_len, sc.build_status
FROM site_components sc
LEFT JOIN content_components cc ON cc.id = sc.component_id
WHERE sc.site_id = (SELECT id FROM sites WHERE domain = 'idea.uk')
ORDER BY sc.slot_name;
-- Expect: UPDATE 1 + UPDATE 1; footer/header show is_active = t; head unchanged (f);
--         rendered_len unchanged everywhere (stale HTML still in place by design).
