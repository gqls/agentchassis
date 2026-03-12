-- ==========================================================================
-- Companies House Enrichment Setup
-- ==========================================================================
-- Separate process from sweep/verify. Runs against already-verified
-- businesses, adding financial data, officer details, and ownership info.
--
-- Rate limiting: ~1 business every 30 seconds (2/min).
-- Companies House allows 600 req/5min = 120/min.
-- We use ~3-4 calls per business, so 2 businesses/min = 8 calls/min.
-- That's 7% of their limit. Very polite.
-- ==========================================================================


-- =========================================================================
-- 1. SCHEMA
-- =========================================================================

CREATE TABLE IF NOT EXISTS business_intel.companies_house_data (
                                                                   business_id UUID NOT NULL REFERENCES business_intel.businesses(id),

    -- Company identification
    company_number VARCHAR(10),
    company_name TEXT,
    company_status TEXT,            -- active, dissolved, liquidation, etc.
    company_type TEXT,              -- ltd, llp, plc, etc.
    incorporation_date DATE,
    cessation_date DATE,            -- if dissolved
    sic_codes TEXT[],
    registered_address JSONB,       -- {address_line_1, locality, postal_code, country}

-- Financial data (from latest accounts)
    accounts_date DATE,
    accounts_type TEXT,             -- micro-entity, small, medium, full
    total_assets_gbp NUMERIC(12,2),
    net_worth_gbp NUMERIC(12,2),
    turnover_gbp NUMERIC(12,2),    -- only if full/medium accounts
    profit_loss_gbp NUMERIC(12,2), -- only if full/medium accounts
    employee_count INTEGER,         -- exact if in accounts
    employee_count_band TEXT,       -- 1-50, 51-250, etc.

-- Officers (directors, secretaries)
    officers JSONB,                 -- [{name, role, appointed_on, resigned_on, dob_month, dob_year, nationality, occupation}]

-- Persons of Significant Control (ownership)
    psc JSONB,                      -- [{name, kind, notified_on, natures_of_control, name_elements}]
    parent_company_name TEXT,       -- corporate PSC name if applicable
    parent_company_number VARCHAR(10),

    -- Owner age signals
    owner_name TEXT,                -- primary director/PSC individual
    owner_dob_year INTEGER,         -- from Companies House (month/year disclosed)
    owner_dob_month INTEGER,
    owner_estimated_age INTEGER,    -- calculated: current_year - dob_year
    owner_appointment_date DATE,    -- earliest appointment as director
    owner_tenure_years INTEGER,     -- years since first appointment
    is_sole_director BOOLEAN,       -- single director = higher succession risk
    is_corporate_owned BOOLEAN,     -- already acquired by a group

-- Succession risk (derived)
    succession_risk TEXT,           -- high, medium, low, acquired, unknown

-- Matching metadata
    match_confidence NUMERIC(3,2),  -- 0.0-1.0
    match_method TEXT,              -- exact_name_postcode, fuzzy_name, sic_filtered, llm_matched
    search_query TEXT,              -- what we searched for
    search_results_count INTEGER,   -- how many results came back

-- Housekeeping
    enriched_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    enrichment_source TEXT DEFAULT 'companies-house-api',
    raw_response JSONB,             -- full API responses for debugging

    PRIMARY KEY (business_id)
    );

-- Index for finding un-enriched businesses
CREATE INDEX IF NOT EXISTS idx_ch_enrichment_pending
    ON business_intel.businesses (verification_status, id)
    WHERE verification_status = 'verified';

-- Index for succession risk queries
CREATE INDEX IF NOT EXISTS idx_ch_succession_risk
    ON business_intel.companies_house_data (succession_risk)
    WHERE succession_risk IS NOT NULL;

COMMENT ON TABLE business_intel.companies_house_data IS
    'Companies House enrichment data. Populated by ch-enricher agent, separate from verification pipeline.';





-- =========================================================================
-- 3. SCHEDULED TASK
-- =========================================================================
-- Runs every 20 minutes. Pre-query checks if there are verified businesses
-- without CH data. Processes 20 per run at ~30s each = ~10 min per batch.
-- Targets the vet-intel pod.

INSERT INTO scheduled_tasks (
    id, name, description, interval_seconds,
    target_agent_type, target_topic, input_data,
    concurrency_group, max_concurrent, timeout_seconds,
    pre_query, enabled
) VALUES (
             gen_random_uuid(),
             'ch-enrichment',
             'Enriches verified businesses with Companies House data (financials, officers, ownership). Very gentle rate limiting.',
             1200,  -- every 20 minutes
             'ch-enricher',
             'system.agent.vet-intel.requests',
             '{"batch_size": 20, "vertical_slug": "veterinary"}',
             'ch-enrichment',
             1,      -- only 1 enricher at a time
             900,    -- 15 min timeout window
             '
         SELECT COUNT(*)::text as unenriched
         FROM business_intel.businesses b
         JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
         LEFT JOIN business_intel.companies_house_data ch ON ch.business_id = b.id
         WHERE bv.slug = ''veterinary''
           AND b.verification_status = ''verified''
           AND ch.business_id IS NULL
         HAVING COUNT(*) > 0
         ',
             false  -- disabled until Go actions are built
         );


-- =========================================================================
-- 4. VERIFY
-- =========================================================================

SELECT name, target_agent_type, target_topic, enabled,
       interval_seconds, timeout_seconds
FROM scheduled_tasks
WHERE name = 'ch-enrichment';

SELECT COUNT(*) as table_exists
FROM information_schema.tables
WHERE table_schema = 'business_intel'
  AND table_name = 'companies_house_data';