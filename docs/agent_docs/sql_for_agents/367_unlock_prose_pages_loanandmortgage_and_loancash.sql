-- 367_unlock_prose_pages_loanandmortgage_and_loancash.sql
--
-- OWNER REQUEST 2026-08-10: "unlock them both and make their components and
-- tools fully editable and upgradable ... do it all through the framework."
--
-- WHAT THIS DOES: flips `pages.rebuild_policy` from 'owned' to 'generic' for the
-- 39 pages on these two sites that carry NO calculator, so the generic pipeline
-- may rebuild and upgrade them.
--
-- WHAT THIS DELIBERATELY DOES NOT DO, AND WHY IT WOULD BE DESTRUCTIVE:
-- the 20 pages that DO carry a calculator (17 on loanandmortgagecalculator,
-- 3 on loancash) are single-component VERBATIM pages — ONE `page_components`
-- row holding the whole page, calculator `<script>` included. `rebuild_policy`
-- is the exact guard that keeps the generic composer off them
-- (`owned_page_guard.go:53`, migration 164's CHECK allows only
-- 'generic'|'owned'), and the three composition loops run
-- `assemble_page -> deploy_page(git_commit) -> save_sections`, i.e. they commit
-- freshly LLM-written HTML to the sites repo — which the site deploys from —
-- one step BEFORE any database-level refusal. Flipping a verbatim tool page to
-- 'generic' therefore replaces a working calculator with prose and ships it.
-- That is the vonc arena clobber (TL-001) which migration 164 exists to prevent.
--
-- Two of those calculators are the ones `bugs_open/224` and `bugs_open/225`
-- were just fixed and oracle-verified on.
--
-- The route to making a TOOL page fully editable is decomposition, not
-- unlocking: split it into prose components (editable) plus a tool component
-- with `lock_type='permanent'`, re-slotted to a `slot_name` the plan will emit
-- (`matchLockedRow` matches on slot_name; a positional `tool-1` does not match,
-- and `save_page_sections_action.go:855` then moves the unmatched locked row to
-- `len(sections)+1` — the calculator lands at the BOTTOM of the page, measured
-- 2026-08-06). `loans/consolidation.html` is already in that shape and is the
-- worked example. Recipe + per-page list:
-- docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/HANDOFF_2026-08-10_unlock_and_upgrade.md
--
-- ALSO NOTE, because unlocking alone is not sufficiency: NEITHER SITE HAS A SITE
-- PLAN (`site_plans` = 0 rows for both, measured 2026-08-10). `rebuild_policy`
-- governs whether the generic pipeline MAY touch a page; `site_plans` /
-- `site_plan_pages` / `site_plan_sections` are what it builds FROM. This
-- migration removes the refusal; it does not create the demand. Seeding a plan
-- is the next step and is in the handoff.
--
-- Editing existing content was never gated by this flag either way: re-assembly
-- of existing `page_components` (page-rerender, section-editor / apply_section_edit)
-- is deliberately NOT guarded — migration 164 says in terms that it "is how owned
-- pages deploy". So these 39 pages were already editable in that sense; what this
-- changes is that the generic pipeline may now REBUILD them wholesale.
--
-- Idempotent: the UPDATE is predicated on the current value, so a replay is a
-- no-op. Verified with DO/RAISE rather than a bare SELECT, because
-- `ON_ERROR_STOP` ignores a non-empty result set and a verify block made of
-- SELECTs cannot stop the COMMIT.
--
-- ROLLBACK: 367_..._ROLLBACK.sql re-locks exactly the rows this stamped.

BEGIN;

-- Record what we are about to change, so the rollback is exact rather than
-- "everything on these sites" — another thread may legitimately flip a page
-- between this migration and any future revert.
CREATE TABLE IF NOT EXISTS _mig367_unlocked_prose_pages (
    page_id     uuid PRIMARY KEY,
    domain      text NOT NULL,
    url         text NOT NULL,
    prev_policy text NOT NULL,
    stamped_at  timestamptz NOT NULL DEFAULT now()
);

WITH tool_flag AS (
    SELECT pc.page_id,
           bool_or(pc.rendered_html ~ 'onclick=|addEventListener') AS has_tool
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
      AND COALESCE(p.rebuild_policy, 'generic') = 'owned'
      AND tf.has_tool IS FALSE          -- NO calculator on the page
)
INSERT INTO _mig367_unlocked_prose_pages (page_id, domain, url, prev_policy)
SELECT page_id, domain, url, prev_policy FROM target
ON CONFLICT (page_id) DO NOTHING;

UPDATE pages p
   SET rebuild_policy = 'generic',
       updated_at = now()
 WHERE p.id IN (SELECT page_id FROM _mig367_unlocked_prose_pages)
   AND COALESCE(p.rebuild_policy, 'generic') = 'owned';

-- ---------------------------------------------------------------------------
-- VERIFY — DO/RAISE, so a wrong result actually aborts the transaction.
-- Three assertions, each of which could come out false:
--   1. the stamped set is the 39 prose pages we measured, split 24/15;
--   2. every stamped page is now 'generic';
--   3. NOT ONE tool page was touched — still 20 'owned', split 17/3.
-- Assertion 3 is the one that matters: it is the negative control, and it is
-- what distinguishes "unlocked the safe pages" from "unlocked the site".
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    n_stamped   int;
    n_lmc       int;
    n_lc        int;
    n_notgen    int;
    n_tool_lmc  int;
    n_tool_lc   int;
BEGIN
    SELECT count(*) INTO n_stamped FROM _mig367_unlocked_prose_pages;
    SELECT count(*) INTO n_lmc FROM _mig367_unlocked_prose_pages
      WHERE domain = 'loanandmortgagecalculator.co.uk';
    SELECT count(*) INTO n_lc FROM _mig367_unlocked_prose_pages
      WHERE domain = 'loancash.co.uk';

    IF n_stamped <> 39 OR n_lmc <> 24 OR n_lc <> 15 THEN
        RAISE EXCEPTION 'mig367: expected 39 prose pages (24 lmc / 15 loancash), '
                        'got % (% / %). The population moved — re-measure before '
                        'forcing this.', n_stamped, n_lmc, n_lc;
    END IF;

    SELECT count(*) INTO n_notgen
      FROM pages p JOIN _mig367_unlocked_prose_pages m ON m.page_id = p.id
     WHERE COALESCE(p.rebuild_policy, 'generic') <> 'generic';
    IF n_notgen <> 0 THEN
        RAISE EXCEPTION 'mig367: % stamped pages are still not generic', n_notgen;
    END IF;

    SELECT count(*) INTO n_tool_lmc
      FROM sites s JOIN pages p ON p.site_id = s.id
      JOIN (SELECT pc.page_id, bool_or(pc.rendered_html ~ 'onclick=|addEventListener') ht
            FROM page_components pc GROUP BY 1) t ON t.page_id = p.id
     WHERE s.domain = 'loanandmortgagecalculator.co.uk'
       AND t.ht IS TRUE AND COALESCE(p.rebuild_policy,'generic') = 'owned';
    SELECT count(*) INTO n_tool_lc
      FROM sites s JOIN pages p ON p.site_id = s.id
      JOIN (SELECT pc.page_id, bool_or(pc.rendered_html ~ 'onclick=|addEventListener') ht
            FROM page_components pc GROUP BY 1) t ON t.page_id = p.id
     WHERE s.domain = 'loancash.co.uk'
       AND t.ht IS TRUE AND COALESCE(p.rebuild_policy,'generic') = 'owned';

    IF n_tool_lmc <> 17 OR n_tool_lc <> 3 THEN
        RAISE EXCEPTION 'mig367: NEGATIVE CONTROL FAILED — tool pages still owned '
                        'should be 17/3, got %/%. A calculator page has been '
                        'unlocked and would be clobbered by the next generic '
                        'build.', n_tool_lmc, n_tool_lc;
    END IF;

    RAISE NOTICE 'mig367 OK: 39 prose pages unlocked (24 lmc / 15 loancash); '
                 '20 tool pages left OWNED (17/3) by design.';
END $$;

COMMIT;
