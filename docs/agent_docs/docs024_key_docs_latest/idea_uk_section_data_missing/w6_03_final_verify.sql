-- W6 step 3 (read-only): final verification, DB side.
-- 3.1 The retained error text from the three retried items (know the transient failure
--     class before calling the retries healthy):
SELECT item_key, attempt_count, left(error, 240) AS error_head
FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND created_by = 'w6_scheme_rebuild'
  AND error IS NOT NULL
ORDER BY item_key;

-- 3.2 The rebuilt stored sections on index — the fixed templates should now be IN
--     page_components (the fossil inverted: legacy var absent, ink/pair present):
SELECT p.name AS page, pc.slot_name,
       (pc.rendered_html LIKE '%--hero-ink%')             AS hero_ink,
       (pc.rendered_html LIKE '%var(--color-cta-bg%')     AS cta_pair,
       (pc.rendered_html LIKE '%color-mix%')              AS has_mix,
       (pc.rendered_html LIKE '%var(--accent-color%')     AS legacy_fossil,   -- expect f
       (pc.rendered_html LIKE '%rgba(255,255,255,0.9)%')  AS old_white        -- expect f
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND p.name IN ('index','about')
ORDER BY p.name, pc.slot_name;
-- Expect: index/hero hero_ink=t + legacy_fossil=f; index/call-to-action cta_pair=t;
--         about/hero-about hero_ink=t; brief-explanation has_mix from the pass-through
--         is NOT expected (it uses plain core-var references) — its success check is
--         old_white=f; every row old_white=f and legacy_fossil=f.
