-- 714: pages.eligible_as_cta_target — the owner-approved CTA opt-out
-- (bugs_open/436; 391 decision 3, ruled 2026-08-25; bugfix_436_cta_eligibility
-- lane, 2026-09-02).
--
-- WHAT IT SAYS. "May the framework CHOOSE this page as a CTA destination?"
-- Today that question is not sayable at all: chooseCTATargets ranks tool/game
-- pages by COALESCE(nav_order,100) then name and takes [0], so a fossil
-- nav_order set at page creation makes an off-topic toy the primary button on
-- every page of a site (bugs_closed/391 — three sites, months, label-locked
-- on 20 of 80 fields), and hiding the page from the nav changes nothing
-- because the ranking does not read in_header.
--
-- DEFAULT TRUE = today's behaviour on every existing row, byte for byte. The
-- new authority (excluding a page) defaults OFF, per the 2026-08-02 §2 ruling
-- that new authority on a shared seam ships as an opt-in field. NOT NULL so
-- the Go readers never meet a three-valued boolean.
--
-- WHO READS IT (the enumeration RFC_022 requires, as of 2026-09-02):
--   * datahelpers.CTAPositionalInteractiveSQL / CTAPositionalHubsSQL — the
--     positional supply for chooseCTATargets' three callers (build resolve,
--     rerender recompute, site header fallback). The RANKING drops opted-out
--     pages (RankCTAPositionalCandidates); the SQL only carries the flag.
--   * datahelpers.CTALabelUniverseSQL — the label-match supply for both CTA
--     writers, check_misdirected_cta and the cta_label_audit. The match
--     REFUSES an opted-out best match (it stays in the pool so a weak-token
--     runner-up cannot win); the judge reports names_ineligible_page.
--   * discovery check cta_rank_anomaly (enabled by 715_HOLD, post-roll) — the
--     alarm half of the pairing: names this column as one of its two remedies.
--
-- WHAT IT DOES NOT TOUCH: nav placement, listings, linkability, and the keep
-- branches (an authored or stored destination pointing AT an opted-out page
-- survives exactly as before — the flag governs the framework's own CHOICES,
-- not links people wrote).
--
-- ORDERING: safe to apply before or after the carrying image rolls. Go code
-- older than this column never mentions it; Go code newer than it (the
-- readers above) requires it — so this file must apply BEFORE that image
-- serves traffic, which the ordinary migration cadence already guarantees.
-- No data change: no page is opted out here; those are owner decisions.

BEGIN;

ALTER TABLE pages
    ADD COLUMN IF NOT EXISTS eligible_as_cta_target boolean NOT NULL DEFAULT true;

COMMENT ON COLUMN pages.eligible_as_cta_target IS
    'May the framework choose this page as a CTA destination? false = never '
    'offered by the positional ranking OR the label match (bugs_open/436). '
    'Governs framework choices only: nav, listings and authored links are '
    'unaffected. Default true = pre-436 behaviour.';

-- Verify inside the transaction, DO/RAISE so a failure aborts the COMMIT
-- (a SELECT-only verify block cannot stop it — ON_ERROR_STOP ignores a
-- non-empty result).
DO $$
DECLARE
    bad integer;
BEGIN
    SELECT count(*) INTO bad FROM pages WHERE eligible_as_cta_target IS DISTINCT FROM true;
    IF bad <> 0 THEN
        RAISE EXCEPTION '714: % pages are not eligible_as_cta_target=true after ADD COLUMN — the default did not take', bad;
    END IF;
END $$;

COMMIT;
