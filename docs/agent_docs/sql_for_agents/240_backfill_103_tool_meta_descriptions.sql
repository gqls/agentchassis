-- 240_backfill_103_tool_meta_descriptions.sql
--
-- bugs_open/103 candidate 3 — repair the tool pages that publish their internal
-- BUILD BRIEF as their public meta description.
--
-- WHY THIS IS A SEPARATE FILE FROM THE CODE FIX. The code fix (commit for 103)
-- stops the NEXT occurrence and repairs nothing already live, because the tool
-- page's INSERT has:
--
--     ON CONFLICT (site_id, name) DO UPDATE SET
--         url = EXCLUDED.url, title = EXCLUDED.title,
--         sections = EXCLUDED.sections, updated_at = NOW()
--
-- meta_description is NOT in that list, so no redeploy overwrites a bad row.
-- A fix that ships without this backfill looks correct in review and changes
-- nothing on the web.
--
-- NOT RUN BY THE AUTHORING SESSION. This rewrites public-facing copy on six
-- client sites, so it is the owner's call, not a side effect of a bug fix.
--
-- ============================================================================
-- STEP 1 — READ-ONLY. What would change, and what the new text would be.
-- Run this first. Read the composed column. If any row's existing description
-- is real copy that merely happens to be long, take it OUT of the UPDATE below
-- by name rather than loosening the predicate.
-- ============================================================================

SELECT s.domain,
       p.name,
       char_length(p.meta_description)                       AS old_len,
       left(p.meta_description, 120)                         AS old_starts,
       'Use our free interactive '
         || lower(regexp_replace(COALESCE(NULLIF(p.title, ''), p.name), '^UK ', ''))
         || ' in your browser, with a companion guide explaining how it works.'
                                                             AS composed
FROM pages p
JOIN sites s ON s.id = p.site_id
WHERE p.page_type = 'tool'
  AND p.meta_description IS NOT NULL
  AND p.meta_description <> ''
  AND (char_length(p.meta_description) > 320
       OR p.meta_description ~* 'no fetch calls|elements, in order|embed [0-9]+ sample|fully self-contained client-side|no backend\)|:\s*\(1\)')
ORDER BY 1, 2;

-- Expected 2026-07-27: 17 rows across 6 sites (dry-run verified). The predicate mirrors the Go
-- guard (datahelpers.MetaDescriptionLooksInternal) — over 320 runes OR a brief
-- marker. NOTE it is DELIBERATELY stricter than the census in bugs_open/103,
-- which uses 400 and therefore reports 15. The two extra rows were read
-- individually and are genuine briefs:
--   gaswholesalers.com/tool-fuel-cost-estimator      (352) "Allows fleet managers ... to input weekly/monthly volume ..."
--   gamesdesign.co.uk/tool-damage-formula-designer   (390) "Lets designers define a damage formula ..."
--
-- NEGATIVE CONTROL — bugs_open/103 §5.3 asks for this, and it is the difference
-- between "zero rows because it is fixed" and "zero rows because the query is
-- broken". This must return a NON-ZERO count, i.e. the predicate still matches
-- a known-bad string:
--
--   SELECT count(*) FROM (VALUES
--     ('The Arena is Spark''s competitive mode, v1 as a fully self-contained client-side experience (no fetch calls, no backend).')
--   ) AS t(d)
--   WHERE char_length(t.d) > 320
--      OR t.d ~* 'no fetch calls|elements, in order|embed [0-9]+ sample|fully self-contained client-side|no backend\)|:\s*\(1\)';
--   -- must be 1

-- ============================================================================
-- STEP 2 — THE WRITE. Only after reading step 1's output.
--
-- Composes from the page's own title, matching the register the companion guide
-- page has always used. Does NOT truncate the existing brief: bugs_open/103
-- candidate 3 is explicit that truncating leaves a clipped spec rather than a
-- description, which is worse than either extreme.
--
-- Scoped to page_type='tool' so it cannot touch a content page whose long
-- description is legitimate editorial copy.
-- ============================================================================

-- BEGIN;
--
-- UPDATE pages p
--    SET meta_description = 'Use our free interactive '
--          || lower(regexp_replace(COALESCE(NULLIF(p.title, ''), p.name), '^UK ', ''))
--          || ' in your browser, with a companion guide explaining how it works.',
--        updated_at = NOW()
--  WHERE p.page_type = 'tool'
--    AND p.meta_description IS NOT NULL
--    AND p.meta_description <> ''
--    AND (char_length(p.meta_description) > 320
--         OR p.meta_description ~* 'no fetch calls|elements, in order|embed [0-9]+ sample|fully self-contained client-side|no backend\)|:\s*\(1\)');
--
-- -- Expect UPDATE 17. If the count differs from step 1's row count, STOP and
-- -- re-read — the set moved under you (this database changes by the minute).
-- COMMIT;

-- ============================================================================
-- STEP 3 — VERIFY ON THE SERVED PAGE, not on the row.
--
-- A repaired row is not a repaired page: pages.meta_description feeds the next
-- render, so each affected page needs a re-render and deploy before the public
-- HTML changes. bugs_open/103 §5.2 makes this the acceptance bar.
--
--   curl -s https://gaswholesalers.com/tools/<page>.html \
--     | grep -o '<meta name="description"[^>]*>'
--
-- Re-run STEP 1 afterwards: it must return 0 rows, AND the negative control
-- above must still return 1.
-- ============================================================================
