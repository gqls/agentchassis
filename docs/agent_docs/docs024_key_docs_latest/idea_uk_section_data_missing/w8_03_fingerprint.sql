-- W8 step 3 (read-only): fingerprint the index/tools split + the 17:25 anomaly.

-- 3.1 Do index's info-card-grid renders carry the three icon URLs? Interpretation:
--     all three t  → Edit B's loop ran fully on index; the illustration miss is elsewhere;
--     only k1 t    → first-row-only / early-exit variant in the APPLIED loop;
--     all f        → the loop errored or never ran for index (pod logs decide).
SELECT p.name, pc.slot_name,
       (pc.rendered_html LIKE '%icon_private_tools%') AS k1,
       (pc.rendered_html LIKE '%icon_ai_tools%')      AS k2,
       (pc.rendered_html LIKE '%icon_editorial%')     AS k3
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND p.name = 'index' AND pc.slot_name = 'info-card-grid';

-- 3.2 The 17:25 anomaly: which rows share that exact timestamp? Only the two
--     brief-explanation rows → a component-targeted single UPDATE (fixer fingerprint);
--     every slot on both pages → a site-wide single-statement pass.
SELECT p.name, pc.slot_name, pc.updated_at
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND p.name IN ('index','tools')
ORDER BY pc.updated_at DESC, p.name, pc.slot_name;
