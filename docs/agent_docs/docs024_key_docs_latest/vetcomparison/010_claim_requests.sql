-- ============================================================================
-- 010_claim_requests.sql
-- ============================================================================
-- Generic claim / opt-out / correction request log for business directories
-- (PLAN Phase 3). Not vertical-specific: keyed to business_intel.businesses,
-- so any comparison site on the chassis uses the same mechanism.
--
-- WHY THIS EXISTS
--   A claimed listing is the licence under which we publish a business's own
--   prices. That licence is only worth anything if we can show, per business:
--   who asked, how we verified they represent the business, what wording they
--   consented to, and when. This table is that record. businesses.claimed_by
--   (previously an unused, unconstrained uuid) now points at the verified
--   request that granted the claim.
--
--   Opt-out requests are logged here too, so a removal is equally traceable:
--   who asked, how we checked, when we actioned it.
--
-- CONSENT TEXT is snapshotted per row, not just versioned by reference. If we
-- later reword the consent, existing rows must still show exactly what that
-- business agreed to at the time.
--
-- IDEMPOTENT.
-- ============================================================================

BEGIN;

CREATE TABLE IF NOT EXISTS business_intel.claim_requests (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id          uuid NOT NULL REFERENCES business_intel.businesses(id),

    request_type         text NOT NULL DEFAULT 'claim'
                              CHECK (request_type IN ('claim', 'optout', 'correction')),

    requester_name       text NOT NULL,
    requester_email      text NOT NULL,
    requester_role       text,
    requester_message    text,

    -- how we satisfied ourselves the requester represents the business
    evidence_method      text CHECK (evidence_method IN
                              ('email_domain_match', 'callback_published_number',
                               'letterhead', 'companies_house', 'other')),
    evidence_note        text,

    status               text NOT NULL DEFAULT 'pending'
                              CHECK (status IN ('pending', 'verified', 'rejected',
                                                'withdrawn', 'actioned')),
    verified_by          text,
    verified_at          timestamptz,
    rejected_reason      text,

    -- exact wording consented to, snapshotted (see WHY above)
    consent_text_version text,
    consent_text         text,
    consent_given_at     timestamptz,

    notes                text,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE business_intel.claim_requests IS
    'Claim / opt-out / correction requests from businesses, with the evidence '
    'we verified and the consent wording they agreed to. A verified claim row '
    'is the licence under which we publish that business''s own prices '
    '(source=claimed_listing). See vetcomparison PLAN Phase 3.';

COMMENT ON COLUMN business_intel.claim_requests.consent_text IS
    'Snapshot of the exact consent wording agreed to, not a reference. Must '
    'survive later rewording of the consent statement.';

CREATE INDEX IF NOT EXISTS idx_bi_claim_requests_business
    ON business_intel.claim_requests (business_id);
CREATE INDEX IF NOT EXISTS idx_bi_claim_requests_status
    ON business_intel.claim_requests (status) WHERE status = 'pending';

-- Give the previously-unused businesses.claimed_by a real meaning: the
-- verified request that granted the claim. Safe to constrain — no rows are
-- claimed at migration time.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'businesses_claimed_by_fkey'
          AND conrelid = 'business_intel.businesses'::regclass
    ) THEN
        ALTER TABLE business_intel.businesses
            ADD CONSTRAINT businesses_claimed_by_fkey
            FOREIGN KEY (claimed_by) REFERENCES business_intel.claim_requests(id);
    END IF;
END $$;

COMMENT ON COLUMN business_intel.businesses.claimed_by IS
    'The verified claim_requests row that granted this claim (who asked, what '
    'evidence we checked, what consent they gave).';

SELECT COUNT(*) AS claim_requests_rows FROM business_intel.claim_requests;

COMMIT;
