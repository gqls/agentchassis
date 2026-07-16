-- ============================================================================
-- 008_publication_optout.sql
-- ============================================================================
-- Generic publication opt-out for businesses (PLAN Phase 3 mechanism,
-- required by the Phase 2 exporter's attributed-prices query).
--
-- Opt-out removes a business from per-business price display only: it stays
-- in the directory (facts) and in unnamed aggregates. A later claim reverses
-- it (claimed listings supersede scraped data entirely).
--
-- IDEMPOTENT.
-- ============================================================================

BEGIN;

ALTER TABLE business_intel.businesses
    ADD COLUMN IF NOT EXISTS publication_optout boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS optout_at timestamptz,
    ADD COLUMN IF NOT EXISTS optout_note text;

COMMENT ON COLUMN business_intel.businesses.publication_optout IS
    'True = do not display per-business prices for this business (it remains '
    'in the directory and in k-anonymous aggregates). Set after verifying the '
    'requester represents the business; reversed by a claim. See vetcomparison '
    'PLAN Phase 3.';

SELECT COUNT(*) FILTER (WHERE publication_optout) AS opted_out,
       COUNT(*) AS total
FROM business_intel.businesses;

COMMIT;
