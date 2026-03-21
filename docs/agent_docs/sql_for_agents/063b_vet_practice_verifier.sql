-- Fall back to town or just name + UK when postcode missing
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,search_practice,config,query_template}',
        '"{{.business_record.business.name}} {{if .business_record.business.postcode}}{{.business_record.business.postcode}}{{else if .business_record.business.town}}{{.business_record.business.town}}{{end}} veterinary practice UK"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'vet-practice-verifier';

---

-- include company number
-- First check the current prompt section for the business extraction
SELECT substring(
               default_config->'workflow'->'steps'->'extract_and_reconcile'->'config'->>'prompt_template',
    1, 500
) FROM agent_definitions WHERE type = 'vet-practice-verifier';

-- Update the extraction prompt to include registration_number
-- This applies to ALL future verifications across all verticals
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,extract_and_reconcile,config,prompt_template}',
        to_jsonb(
                replace(
                        default_config->'workflow'->'steps'->'extract_and_reconcile'->'config'->>'prompt_template',
                        '   - group_name, business_type',
                        '   - group_name, business_type
               - registration_number (Companies House registration number if shown on the website, typically in the footer or terms page. 7-8 digit number, sometimes prefixed SC for Scottish or NI for Northern Irish companies. null if not found)'
                )
        ),
        true
                     ),
    updated_at = NOW()
WHERE type = 'vet-practice-verifier';

-- ============================================================================
-- Company Number Extraction — Column + Prompt Update
-- ============================================================================
-- Adds company_number_scraped column to businesses table and updates the
-- vet-practice-verifier prompt to extract registration numbers.
-- The regex fallback runs in business_intel_actions.go StoreBusinessVerificationAction.
-- ============================================================================

-- Column for storing scraped/extracted company registration numbers.
-- NULL = not yet checked, '' = checked but not found, '12345678' = found.
ALTER TABLE business_intel.businesses
    ADD COLUMN IF NOT EXISTS company_number_scraped VARCHAR(10);

-- Index for businesses that haven't been checked yet
CREATE INDEX IF NOT EXISTS idx_bi_businesses_company_number_scraped
    ON business_intel.businesses (company_number_scraped)
    WHERE company_number_scraped IS NULL AND verification_status = 'verified';

-- ============================================================================
-- Update verifier prompt to extract registration_number.
-- Applies to all future verifications across all verticals.
-- ============================================================================

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,extract_and_reconcile,config,prompt_template}',
        to_jsonb(
                replace(
                        default_config->'workflow'->'steps'->'extract_and_reconcile'->'config'->>'prompt_template',
                        '   - group_name, business_type',
                        '   - group_name, business_type
               - registration_number (Companies House registration number if shown on the website, typically in the footer or terms page. 7-8 digit number, sometimes prefixed SC for Scottish or NI for Northern Irish companies. null if not found)'
                )
        ),
        true
                     ),
    updated_at = NOW()
WHERE type = 'vet-practice-verifier';