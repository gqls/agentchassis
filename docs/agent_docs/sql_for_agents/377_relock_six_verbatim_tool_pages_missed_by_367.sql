-- 377_relock_six_verbatim_tool_pages_missed_by_367.sql
--
-- WHAT THIS FIXES. Migration 367 (applied 2026-08-10 12:22Z) unlocked the pages
-- on loanandmortgagecalculator.co.uk + loancash.co.uk that "carry NO calculator",
-- and deliberately refused the ones that do, because a verbatim tool page flipped
-- to 'generic' gets its calculator replaced by LLM prose and shipped (TL-001,
-- migration 164, `owned_page_guard.go`).
--
-- SIX OF THE PAGES IT UNLOCKED CARRY CALCULATORS. It classified with
--     bool_or(rendered_html ~ 'onclick=|addEventListener')
-- and these six bind their handlers by OTHER spellings — `oninput=`, `onsubmit=`,
-- `onchange=` — and two of them load the shared `/assets/js/calculators.js`,
-- where the `addEventListener` calls live in the EXTERNAL file, not in the stored
-- HTML the detector read. All six are named as calculators by this lane's own
-- hand-authored `CALCULATOR_URLS` set in `decompose_lmc.py`, written days earlier
-- and independent of any of this SQL:
--
--     /loans/compare-loans.html            (also loads calculators.js)
--     /loans/interest-rate-stress-test.html (also loads calculators.js)
--     /loans/loan-vs-savings.html
--     /loans/settlement-calculator.html
--     /loans/damage-checker.html
--     /mortgages/fact-finder.html
--
-- WHY 367's NEGATIVE CONTROL DID NOT CATCH IT: the control asserted "17 LMC + 3
-- loancash tool pages are still owned" using THE SAME `onclick=|addEventListener`
-- expression as the filter. A control that shares its subject's blind spot agrees
-- with it by construction — it was induced and it did fire, and it still could not
-- have seen these six. The generalisable rule is in
-- docs/agent_docs/docs024_key_docs_latest/LANDMINES.md.
--
-- WHY THIS IS URGENT AND NOT COSMETIC, measured 2026-08-10 ~19:30Z:
--   * all six are still ONE verbatim `page_components` row, `sections =
--     ["ported-page"]`, calculator `<script>` inline — the exact shape 367 refused;
--   * all six sit at `build_status='needs_rebuild'`, which is what
--     `get_pages_to_build` selects on, and its only ownership filter is
--     `COALESCE(rebuild_policy,'generic') <> 'owned'`
--     (`get_pages_to_build_actions.go`, `ownedPageExclusionSQL`);
--   * each already carries an open `page_rerender` work item;
--   * the generic full-rebuild path HAS ALREADY RUN against these pages. On
--     2026-08-09, items `needs_page:loans-compare-loans` and 19 siblings reached
--     step `save_sections` and died there with
--       "page loans-compare-loans is rebuild_policy=owned (tool/widget-owned): a
--        generic section save would clobber it ... Refusing to overwrite."
--     That refusal is the only reason those calculators still exist, and 367
--     removed it for six of them. (The sites repo took no clobbering commit in
--     that window — `git log` 2026-08-08 20:00 → 2026-08-09 03:00 shows only the
--     224 fix and a consolidation rerender — so on this path the DB guard fired
--     before anything reached the repo. Do not rely on that ordering: it is the
--     guard that saved the page, and this migration restores it.)
--
-- WHAT THIS DOES: puts those six back to 'owned'. Nothing else changes. It is a
-- restoration of the state they were in this morning, not a new policy.
--
-- WHAT THIS IS NOT: it is not a decision to leave them un-upgradable. The route
-- to a fully framework-controlled tool page is DECOMPOSITION (prose components +
-- one `lock_type='permanent'` tool component, the shape `loans/consolidation.html`
-- is already in), and THEN the flip. That work is Track B in
-- docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/HANDOFF_2026-08-10c_continue_here.md
-- Re-locking is what makes it safe to do that work one page at a time instead of
-- racing the build loop.
--
-- DETECTOR USED HERE — three independent spellings, OR'd, so no single blind spot
-- decides: (a) any inline event-handler attribute or addEventListener or a
-- calculators.js tag; (b) a form control (`<input `/`<select `/`<textarea `/
-- `<form `); (c) DOM addressing (`getElementById`/`querySelector`). Measured over
-- all 38 generic verbatim pages on the two sites: the six match on ALL THREE, the
-- other 32 match on NONE. There is no borderline case to argue about.
--
-- Idempotent: the UPDATE is predicated on the current value; a replay is a no-op.
-- ROLLBACK: 377_..._ROLLBACK.sql re-unlocks exactly the rows this stamped.

BEGIN;

CREATE TABLE IF NOT EXISTS _mig377_relocked_tool_pages (
    page_id     uuid PRIMARY KEY,
    domain      text NOT NULL,
    url         text NOT NULL,
    prev_policy text NOT NULL,
    stamped_at  timestamptz NOT NULL DEFAULT now()
);

WITH tool_flag AS (
    SELECT pc.page_id,
           bool_or(pc.rendered_html ~ 'onclick=|addEventListener|oninput=|onsubmit=|onchange=|onkeyup=|onblur=|onkeydown=|calculators\.js'
                   OR pc.rendered_html ~ '<input |<select |<textarea |<form '
                   OR pc.rendered_html ~ 'getElementById|querySelector') AS has_tool
    FROM page_components pc
    GROUP BY pc.page_id
),
target AS (
    SELECT p.id AS page_id, s.domain, p.url,
           COALESCE(p.rebuild_policy, 'generic') AS prev_policy
    FROM sites s
    JOIN pages p ON p.site_id = s.id
    JOIN tool_flag tf ON tf.page_id = p.id
    WHERE s.domain IN ('loanandmortgagecalculator.co.uk', 'loancash.co.uk')
      AND COALESCE(p.rebuild_policy, 'generic') = 'generic'
      AND p.sections::text = '["ported-page"]'   -- single verbatim row: the danger class
      AND tf.has_tool IS TRUE
)
INSERT INTO _mig377_relocked_tool_pages (page_id, domain, url, prev_policy)
SELECT page_id, domain, url, prev_policy FROM target
ON CONFLICT (page_id) DO NOTHING;

UPDATE pages p
   SET rebuild_policy = 'owned',
       updated_at = now()
 WHERE p.id IN (SELECT page_id FROM _mig377_relocked_tool_pages)
   AND COALESCE(p.rebuild_policy, 'generic') = 'generic';

-- ---------------------------------------------------------------------------
-- VERIFY — DO/RAISE, because a verify block made of SELECTs cannot stop a COMMIT
-- (ON_ERROR_STOP ignores a non-empty result set).
--
-- Assertion 1 is checked against an INDEPENDENT source: the six URLs below are
-- transcribed from `decompose_lmc.py`'s hand-authored CALCULATOR_URLS, which was
-- written on 2026-08-05 and has never read this SQL or 367's detector. If the
-- detector in this file disagrees with that list in either direction, this aborts.
-- Assertion 3 is the over-locking control: 32 generic verbatim pages must remain.
-- Assertion 4 is stated for replay safety and is TAUTOLOGICAL with respect to the
-- UPDATE above — it is not evidence, and is marked as such rather than counted.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    -- 'domain|url', not url alone: `pages.url` is NOT unique across sites — both
    -- of these sites have a /guides/jargon-buster.html and a /legal.html, so a
    -- url-only comparison silently conflates them. Found by inducing this very
    -- block: a deliberate over-lock of "one" page stamped TWO rows and the
    -- assertion reported neither as unexpected.
    expected text[] := ARRAY[
        'loanandmortgagecalculator.co.uk|/loans/compare-loans.html',
        'loanandmortgagecalculator.co.uk|/loans/damage-checker.html',
        'loanandmortgagecalculator.co.uk|/loans/interest-rate-stress-test.html',
        'loanandmortgagecalculator.co.uk|/loans/loan-vs-savings.html',
        'loanandmortgagecalculator.co.uk|/loans/settlement-calculator.html',
        'loanandmortgagecalculator.co.uk|/mortgages/fact-finder.html'
    ];
    n_stamped   int;
    missing     text;
    extra       text;
    n_notowned  int;
    n_gen_lmc   int;
    n_gen_lc    int;
    n_left      int;
BEGIN
    SELECT count(*) INTO n_stamped FROM _mig377_relocked_tool_pages;

    SELECT string_agg(u, ', ') INTO missing
      FROM unnest(expected) AS u
     WHERE u NOT IN (SELECT domain || '|' || url FROM _mig377_relocked_tool_pages);
    SELECT string_agg(domain || '|' || url, ', ') INTO extra
      FROM _mig377_relocked_tool_pages
     WHERE (domain || '|' || url) <> ALL (expected);

    IF n_stamped <> 6 OR missing IS NOT NULL OR extra IS NOT NULL THEN
        RAISE EXCEPTION 'mig377: stamped set disagrees with decompose_lmc.py''s '
                        'CALCULATOR_URLS. stamped=%, missing from stamp=[%], '
                        'unexpected in stamp=[%]. Re-measure before forcing.',
                        n_stamped, COALESCE(missing,'-'), COALESCE(extra,'-');
    END IF;

    SELECT count(*) INTO n_notowned
      FROM pages p JOIN _mig377_relocked_tool_pages m ON m.page_id = p.id
     WHERE COALESCE(p.rebuild_policy, 'generic') <> 'owned';
    IF n_notowned <> 0 THEN
        RAISE EXCEPTION 'mig377: % stamped pages are still not owned', n_notowned;
    END IF;

    SELECT count(*) INTO n_gen_lmc
      FROM sites s JOIN pages p ON p.site_id = s.id
     WHERE s.domain = 'loanandmortgagecalculator.co.uk'
       AND COALESCE(p.rebuild_policy,'generic') = 'generic'
       AND p.sections::text = '["ported-page"]';
    SELECT count(*) INTO n_gen_lc
      FROM sites s JOIN pages p ON p.site_id = s.id
     WHERE s.domain = 'loancash.co.uk'
       AND COALESCE(p.rebuild_policy,'generic') = 'generic'
       AND p.sections::text = '["ported-page"]';

    IF n_gen_lmc <> 17 OR n_gen_lc <> 15 THEN
        RAISE EXCEPTION 'mig377: OVER-LOCKING CONTROL FAILED — generic verbatim '
                        'pages should be 17 lmc / 15 loancash after this, got %/%. '
                        'A prose page has been re-locked and 367''s work partly '
                        'undone.', n_gen_lmc, n_gen_lc;
    END IF;

    SELECT count(*) INTO n_left
      FROM sites s JOIN pages p ON p.site_id = s.id
      JOIN page_components pc ON pc.page_id = p.id
     WHERE s.domain IN ('loanandmortgagecalculator.co.uk','loancash.co.uk')
       AND COALESCE(p.rebuild_policy,'generic') = 'generic'
       AND p.sections::text = '["ported-page"]'
       AND (pc.rendered_html ~ 'onclick=|addEventListener|oninput=|onsubmit=|onchange=|onkeyup=|onblur=|onkeydown=|calculators\.js'
            OR pc.rendered_html ~ '<input |<select |<textarea |<form '
            OR pc.rendered_html ~ 'getElementById|querySelector');
    IF n_left <> 0 THEN
        RAISE EXCEPTION 'mig377: % verbatim page(s) carrying tool machinery are '
                        'still generic', n_left;
    END IF;

    RAISE NOTICE 'mig377 OK: 6 verbatim calculator pages re-locked to OWNED '
                 '(367 had unlocked them on a narrow detector); 32 genuine prose '
                 'pages left generic (17 lmc / 15 loancash).';
END $$;

COMMIT;
