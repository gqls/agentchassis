-- W7b: the two plan-imagery rows + two shaped needs_imagery items. Shapes copied from
-- reality: spi columns per \d (scope_ref = page:ORDINAL — index:1 / tools:1 from the
-- corrected 0.5); item spec per the 0.6 full spec, with truthful provenance
-- (check/source manual). Keys follow the house style (hero_home → illustration_home);
-- W7c's ensureAssets extension maps kind → resolver path, as hero already does.
-- Dedup: NOT EXISTS mirrors spi's unique tuple and idx_swi_dedup respectively.

INSERT INTO site_plan_imagery (plan_id, scope, scope_ref, key, kind, prompt, style_hints, ordering, source)
SELECT sp.id, 'section', v.scope_ref, v.key, 'illustration', v.prompt,
       '{"mood": "warm", "aspect_ratio": "4:3"}'::jsonb, 0, 'manual'
FROM site_plans sp,
     (VALUES
       ('index:1', 'illustration_home',
        'A warm editorial illustration of an idea taking shape: a small paper boat folded from a lined notebook page resting on a linen surface beside a pencil, soft daylight, parchment tones with a single rust accent, minimal and quiet, no text, no people'),
       ('tools:1', 'illustration_tools',
        'A quiet editorial still-life of small hand tools arranged neatly on linen — a pencil, a ruler, a magnifying glass — parchment palette with a single rust accent, soft daylight, minimal, no text, no people')
     ) AS v(scope_ref, key, prompt)
WHERE sp.site_id = (SELECT id FROM sites WHERE domain = 'idea.uk')
  AND sp.is_current = true
  AND NOT EXISTS (
    SELECT 1 FROM site_plan_imagery x
    WHERE x.plan_id = sp.id AND x.scope = 'section'
      AND COALESCE(x.scope_ref,'') = v.scope_ref AND x.key = v.key
  )
RETURNING scope_ref, key;

INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary,
                             spec, priority, handler_agent, status, created_by, item_key)
SELECT s.id, 'manual', 'build', 'needs_imagery', 'medium',
       'Imagery section/' || v.key || ' (kind=illustration) requested but no asset for ' || v.key,
       jsonb_build_object(
         'key', v.key, 'kind', 'illustration', 'check', 'manual',
         'scope', 'section', 'scope_ref', v.scope_ref,
         'prompt', v.prompt,
         'purpose', 'illustration', 'asset_key', v.key,
         'style_hints', jsonb_build_object('mood','warm','aspect_ratio','4:3'),
         'brand_update', false),
       98, 'image-build-handler', 'triaged', 'w7b_manual_imagery',
       'needs_imagery:section:' || v.scope_ref || ':' || v.key
FROM sites s,
     (VALUES
       ('index:1', 'illustration_home',
        'A warm editorial illustration of an idea taking shape: a small paper boat folded from a lined notebook page resting on a linen surface beside a pencil, soft daylight, parchment tones with a single rust accent, minimal and quiet, no text, no people'),
       ('tools:1', 'illustration_tools',
        'A quiet editorial still-life of small hand tools arranged neatly on linen — a pencil, a ruler, a magnifying glass — parchment palette with a single rust accent, soft daylight, minimal, no text, no people')
     ) AS v(scope_ref, key, prompt)
WHERE s.domain = 'idea.uk'
  AND NOT EXISTS (
    SELECT 1 FROM site_work_items w
    WHERE w.site_id = s.id
      AND w.item_key = 'needs_imagery:section:' || v.scope_ref || ':' || v.key
      AND w.status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved')
  )
RETURNING item_key, status;
-- Expect: two RETURNING rows from each insert.
