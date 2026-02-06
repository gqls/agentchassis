-- =============================================
-- Business Intelligence Platform Schema
-- Namespace: business_intel (within clients_db)
-- =============================================
-- Run against: clients_db as clients_user (or postgres superuser)
--
-- This creates all tables in the business_intel schema,
-- separate from the public schema where framework and
-- website builder tables live.

BEGIN;

-- Create the schema
CREATE SCHEMA IF NOT EXISTS business_intel;

-- Grant usage to clients_user (in case running as postgres)
GRANT USAGE ON SCHEMA business_intel TO clients_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA business_intel TO clients_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA business_intel GRANT ALL ON TABLES TO clients_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA business_intel GRANT ALL ON SEQUENCES TO clients_user;

-- =============================================
-- LAYER 1: Core business identity
-- =============================================

-- What vertical is this? Vets, seaweed farms, etc.
CREATE TABLE business_intel.business_verticals (
                                                   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                                   slug TEXT UNIQUE NOT NULL,               -- 'veterinary', 'seaweed-farming'
                                                   display_name TEXT NOT NULL,
                                                   description TEXT,
                                                   default_agent_type TEXT,                  -- 'vet-practice-verifier'
                                                   collection_config JSONB DEFAULT '{}',     -- vertical-specific search terms, etc.
                                                   created_at TIMESTAMP DEFAULT NOW()
);

-- Core business records - universal fields only
CREATE TABLE business_intel.businesses (
                                           id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                           vertical_id UUID REFERENCES business_intel.business_verticals(id),

    -- Identity
                                           name TEXT NOT NULL,
                                           slug TEXT,                                -- url-friendly, unique within vertical
                                           trading_name TEXT,                        -- if different from registered name

    -- Location
                                           address_line1 TEXT,
                                           address_line2 TEXT,
                                           town TEXT,
                                           county TEXT,
                                           postcode TEXT,
                                           country TEXT DEFAULT 'GB',
                                           latitude DECIMAL(9,6),
                                           longitude DECIMAL(9,6),

    -- Contact
                                           phone TEXT,
                                           email TEXT,
                                           website_url TEXT,

    -- Classification
                                           business_type TEXT,                       -- vertical-specific subtype
                                           group_name TEXT,                          -- parent company/chain if any
                                           is_independent BOOLEAN,

    -- Data quality tracking
                                           verification_status TEXT DEFAULT 'unverified',
                                           confidence_score DECIMAL(3,2),
                                           last_verified_at TIMESTAMP,

    -- Self-service / ownership (future)
                                           claimed_by UUID,
                                           claimed_at TIMESTAMP,
                                           is_claimed BOOLEAN DEFAULT FALSE,

    -- Lifecycle
                                           is_active BOOLEAN DEFAULT TRUE,
                                           created_at TIMESTAMP DEFAULT NOW(),
                                           updated_at TIMESTAMP DEFAULT NOW(),

                                           UNIQUE(vertical_id, slug)
);

CREATE INDEX idx_bi_businesses_vertical ON business_intel.businesses(vertical_id);
CREATE INDEX idx_bi_businesses_postcode ON business_intel.businesses(postcode);
CREATE INDEX idx_bi_businesses_verification ON business_intel.businesses(verification_status);
CREATE INDEX idx_bi_businesses_town ON business_intel.businesses(town);
CREATE INDEX idx_bi_businesses_group ON business_intel.businesses(group_name) WHERE group_name IS NOT NULL;
CREATE INDEX idx_bi_businesses_website ON business_intel.businesses(website_url) WHERE website_url IS NOT NULL;

-- =============================================
-- LAYER 2: Vertical-specific data (vet example)
-- =============================================

-- Vet-specific details (1:1 with businesses)
CREATE TABLE business_intel.vet_practice_details (
                                                     business_id UUID PRIMARY KEY REFERENCES business_intel.businesses(id),

    -- Vet-specific
                                                     species_treated TEXT[],                   -- '{dogs,cats,rabbits,exotic}'
                                                     emergency_service BOOLEAN,
                                                     out_of_hours_provider TEXT,

    -- Capacity/status
                                                     accepting_new_clients BOOLEAN,
                                                     accepting_new_clients_updated_at TIMESTAMP,

    -- Accreditations
                                                     accreditations TEXT[],                    -- '{RCVS_Practice_Standards, Cat_Friendly}'

    -- Staff summary
                                                     num_vets INTEGER,
                                                     num_nurses INTEGER,
                                                     head_vet_name TEXT,

    -- Facilities
                                                     has_own_lab BOOLEAN,
                                                     has_imaging BOOLEAN,
                                                     has_surgical_suite BOOLEAN,
                                                     parking_available BOOLEAN,
                                                     wheelchair_accessible BOOLEAN,

                                                     updated_at TIMESTAMP DEFAULT NOW()
);

-- Pricing (any vertical)
CREATE TABLE business_intel.business_prices (
                                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                                business_id UUID REFERENCES business_intel.businesses(id),

                                                service_category TEXT NOT NULL,            -- 'consultation', 'vaccination', 'surgery'
                                                service_name TEXT NOT NULL,                -- 'Initial consultation', 'Booster vaccination'

                                                price_gbp DECIMAL(10,2),
                                                price_qualifier TEXT,                     -- 'from', 'approximately', 'fixed'
                                                includes_vat BOOLEAN DEFAULT TRUE,

    -- Provenance
                                                source TEXT NOT NULL,                     -- 'website_scrape', 'business_submitted'
                                                source_url TEXT,
                                                observed_at TIMESTAMP DEFAULT NOW(),

    -- Historical: latest is max observed_at per service
                                                is_current BOOLEAN DEFAULT TRUE
);

CREATE INDEX idx_bi_prices_business ON business_intel.business_prices(business_id);
CREATE INDEX idx_bi_prices_current ON business_intel.business_prices(business_id) WHERE is_current = TRUE;

-- =============================================
-- LAYER 3: Source tracking / provenance
-- =============================================

-- Every scrape/search/submission is a "data observation"
CREATE TABLE business_intel.data_observations (
                                                  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                                  business_id UUID REFERENCES business_intel.businesses(id),
                                                  person_id UUID,                           -- FK added after people table created

    -- What source provided this
                                                  source_type TEXT NOT NULL,                 -- 'web_scrape', 'web_search', 'api',
    -- 'business_submitted', 'manual_entry', 'directory'
                                                  source_name TEXT,                          -- 'practice_website', 'yell', 'google_maps', 'rcvs'
                                                  source_url TEXT,

    -- The raw extracted data
                                                  raw_data JSONB NOT NULL,
                                                  extracted_fields JSONB,

    -- Quality
                                                  extraction_confidence DECIMAL(3,2),
                                                  extraction_notes TEXT,

    -- Processing link
                                                  orchestration_id UUID,                    -- links to the agent run that produced this
                                                  collected_at TIMESTAMP DEFAULT NOW(),
                                                  processed_at TIMESTAMP
);

CREATE INDEX idx_bi_observations_business ON business_intel.data_observations(business_id);
CREATE INDEX idx_bi_observations_source ON business_intel.data_observations(source_type, source_name);
CREATE INDEX idx_bi_observations_collected ON business_intel.data_observations(collected_at DESC);

-- Collection task queue
CREATE TABLE business_intel.collection_tasks (
                                                 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                                 business_id UUID REFERENCES business_intel.businesses(id),
                                                 vertical_id UUID REFERENCES business_intel.business_verticals(id),

                                                 task_type TEXT NOT NULL,                   -- 'initial_verification', 'price_refresh',
    -- 'status_check', 'discovery'
                                                 status TEXT DEFAULT 'pending',             -- 'pending', 'in_progress', 'completed',
    -- 'failed', 'needs_review'
                                                 priority INTEGER DEFAULT 5,               -- 1=highest

    -- Scheduling
                                                 scheduled_for TIMESTAMP,
                                                 started_at TIMESTAMP,
                                                 completed_at TIMESTAMP,

    -- Links to agent execution
                                                 orchestration_id UUID,

    -- Results
                                                 result_summary JSONB,
                                                 error_message TEXT,
                                                 retry_count INTEGER DEFAULT 0,

                                                 created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_bi_tasks_pending ON business_intel.collection_tasks(priority, scheduled_for)
    WHERE status = 'pending';
CREATE INDEX idx_bi_tasks_business ON business_intel.collection_tasks(business_id);
CREATE INDEX idx_bi_tasks_status ON business_intel.collection_tasks(status);

-- =============================================
-- LAYER 4: People / Contacts
-- =============================================

CREATE TABLE business_intel.people (
                                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Identity
                                       first_name TEXT,
                                       last_name TEXT,
                                       display_name TEXT,                        -- 'Dr Sarah Chen'
                                       slug TEXT UNIQUE,

    -- Contact (personal, not via a business)
                                       email TEXT,
                                       phone TEXT,
                                       linkedin_url TEXT,

    -- Professional
                                       qualifications TEXT[],                    -- '{BVSc, MRCVS, CertAVP}'
                                       specialisms TEXT[],                       -- '{feline, dermatology}'

    -- Data quality
                                       verification_status TEXT DEFAULT 'unverified',
                                       last_verified_at TIMESTAMP,

    -- Self-service (future)
                                       claimed_by UUID,
                                       is_claimed BOOLEAN DEFAULT FALSE,

                                       created_at TIMESTAMP DEFAULT NOW(),
                                       updated_at TIMESTAMP DEFAULT NOW()
);

-- People <-> Business relationships over time
CREATE TABLE business_intel.people_business_roles (
                                                      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                                      person_id UUID REFERENCES business_intel.people(id),
                                                      business_id UUID REFERENCES business_intel.businesses(id),

                                                      role_title TEXT,                           -- 'Head Vet', 'Practice Manager'
                                                      role_type TEXT,                            -- 'clinical', 'management', 'owner', 'contact'

    -- Time dimension
                                                      started_at DATE,                          -- when they joined (if known)
                                                      ended_at DATE,                            -- NULL = current
                                                      is_current BOOLEAN DEFAULT TRUE,

    -- Provenance
                                                      source TEXT,                              -- 'website_scrape', 'linkedin'
                                                      source_url TEXT,
                                                      observed_at TIMESTAMP DEFAULT NOW(),

    -- Contact-specific (for verticals where contact info matters)
                                                      is_primary_contact BOOLEAN DEFAULT FALSE,
                                                      contact_email TEXT,
                                                      contact_phone TEXT,

                                                      UNIQUE(person_id, business_id, role_title, started_at)
);

CREATE INDEX idx_bi_pbr_person ON business_intel.people_business_roles(person_id);
CREATE INDEX idx_bi_pbr_business ON business_intel.people_business_roles(business_id);
CREATE INDEX idx_bi_pbr_current ON business_intel.people_business_roles(business_id) WHERE is_current = TRUE;

-- Now add the FK from data_observations to people
ALTER TABLE business_intel.data_observations
    ADD CONSTRAINT fk_observations_person
        FOREIGN KEY (person_id) REFERENCES business_intel.people(id);

-- =============================================
-- Seed data: verticals
-- =============================================

INSERT INTO business_intel.business_verticals (slug, display_name, description, default_agent_type)
VALUES
    ('veterinary', 'Veterinary Practices', 'UK veterinary practices and clinics', 'vet-practice-verifier'),
    ('seaweed-farming', 'Seaweed Farms', 'UK seaweed farming and harvesting companies', NULL)
    ON CONFLICT (slug) DO NOTHING;

COMMIT;


-- now idempotent

-- =============================================
-- Business Intelligence Platform Schema
-- Namespace: business_intel (within clients_db)
-- =============================================
-- Run against: clients_db as clients_user (or postgres superuser)
--
-- This creates all tables in the business_intel schema,
-- separate from the public schema where framework and
-- website builder tables live.

BEGIN;

-- Create the schema
CREATE SCHEMA IF NOT EXISTS business_intel;

-- Grant usage to clients_user (in case running as postgres)
GRANT USAGE ON SCHEMA business_intel TO clients_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA business_intel TO clients_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA business_intel GRANT ALL ON TABLES TO clients_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA business_intel GRANT ALL ON SEQUENCES TO clients_user;

-- =============================================
-- LAYER 1: Core business identity
-- =============================================

-- What vertical is this? Vets, seaweed farms, etc.
CREATE TABLE IF NOT EXISTS business_intel.business_verticals (
                                                                 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT UNIQUE NOT NULL,               -- 'veterinary', 'seaweed-farming'
    display_name TEXT NOT NULL,
    description TEXT,
    default_agent_type TEXT,                  -- 'vet-practice-verifier'
    collection_config JSONB DEFAULT '{}',     -- vertical-specific search terms, etc.
    created_at TIMESTAMP DEFAULT NOW()
    );

-- Core business records - universal fields only
CREATE TABLE IF NOT EXISTS business_intel.businesses (
                                                         id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vertical_id UUID REFERENCES business_intel.business_verticals(id),

    -- Identity
    name TEXT NOT NULL,
    slug TEXT,                                -- url-friendly, unique within vertical
    trading_name TEXT,                        -- if different from registered name

-- Location
    address_line1 TEXT,
    address_line2 TEXT,
    town TEXT,
    county TEXT,
    postcode TEXT,
    country TEXT DEFAULT 'GB',
    latitude DECIMAL(9,6),
    longitude DECIMAL(9,6),

    -- Contact
    phone TEXT,
    email TEXT,
    website_url TEXT,

    -- Classification
    business_type TEXT,                       -- vertical-specific subtype
    group_name TEXT,                          -- parent company/chain if any
    is_independent BOOLEAN,

    -- Data quality tracking
    verification_status TEXT DEFAULT 'unverified',
    confidence_score DECIMAL(3,2),
    last_verified_at TIMESTAMP,

    -- Self-service / ownership (future)
    claimed_by UUID,
    claimed_at TIMESTAMP,
    is_claimed BOOLEAN DEFAULT FALSE,

    -- Lifecycle
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(vertical_id, slug)
    );

CREATE INDEX IF NOT EXISTS idx_bi_businesses_vertical ON business_intel.businesses(vertical_id);
CREATE INDEX IF NOT EXISTS idx_bi_businesses_postcode ON business_intel.businesses(postcode);
CREATE INDEX IF NOT EXISTS idx_bi_businesses_verification ON business_intel.businesses(verification_status);
CREATE INDEX IF NOT EXISTS idx_bi_businesses_town ON business_intel.businesses(town);
CREATE INDEX IF NOT EXISTS idx_bi_businesses_group ON business_intel.businesses(group_name) WHERE group_name IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_bi_businesses_website ON business_intel.businesses(website_url) WHERE website_url IS NOT NULL;

-- =============================================
-- LAYER 2: Vertical-specific data (vet example)
-- =============================================

-- Vet-specific details (1:1 with businesses)
CREATE TABLE IF NOT EXISTS business_intel.vet_practice_details (
                                                                   business_id UUID PRIMARY KEY REFERENCES business_intel.businesses(id),

    -- Vet-specific
    species_treated TEXT[],                   -- '{dogs,cats,rabbits,exotic}'
    emergency_service BOOLEAN,
    out_of_hours_provider TEXT,

    -- Capacity/status
    accepting_new_clients BOOLEAN,
    accepting_new_clients_updated_at TIMESTAMP,

    -- Accreditations
    accreditations TEXT[],                    -- '{RCVS_Practice_Standards, Cat_Friendly}'

-- Staff summary
    num_vets INTEGER,
    num_nurses INTEGER,
    head_vet_name TEXT,

    -- Facilities
    has_own_lab BOOLEAN,
    has_imaging BOOLEAN,
    has_surgical_suite BOOLEAN,
    parking_available BOOLEAN,
    wheelchair_accessible BOOLEAN,

    updated_at TIMESTAMP DEFAULT NOW()
    );

-- Pricing (any vertical)
CREATE TABLE IF NOT EXISTS business_intel.business_prices (
                                                              id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID REFERENCES business_intel.businesses(id),

    service_category TEXT NOT NULL,            -- 'consultation', 'vaccination', 'surgery'
    service_name TEXT NOT NULL,                -- 'Initial consultation', 'Booster vaccination'

    price_gbp DECIMAL(10,2),
    price_qualifier TEXT,                     -- 'from', 'approximately', 'fixed'
    includes_vat BOOLEAN DEFAULT TRUE,

    -- Provenance
    source TEXT NOT NULL,                     -- 'website_scrape', 'business_submitted'
    source_url TEXT,
    observed_at TIMESTAMP DEFAULT NOW(),

    -- Historical: latest is max observed_at per service
    is_current BOOLEAN DEFAULT TRUE
    );

CREATE INDEX IF NOT EXISTS idx_bi_prices_business ON business_intel.business_prices(business_id);
CREATE INDEX IF NOT EXISTS idx_bi_prices_current ON business_intel.business_prices(business_id) WHERE is_current = TRUE;

-- =============================================
-- LAYER 3: Source tracking / provenance
-- =============================================

-- Every scrape/search/submission is a "data observation"
CREATE TABLE IF NOT EXISTS business_intel.data_observations (
                                                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID REFERENCES business_intel.businesses(id),
    person_id UUID,                           -- FK added after people table created

-- What source provided this
    source_type TEXT NOT NULL,                 -- 'web_scrape', 'web_search', 'api',
-- 'business_submitted', 'manual_entry', 'directory'
    source_name TEXT,                          -- 'practice_website', 'yell', 'google_maps', 'rcvs'
    source_url TEXT,

    -- The raw extracted data
    raw_data JSONB NOT NULL,
    extracted_fields JSONB,

    -- Quality
    extraction_confidence DECIMAL(3,2),
    extraction_notes TEXT,

    -- Processing link
    orchestration_id UUID,                    -- links to the agent run that produced this
    collected_at TIMESTAMP DEFAULT NOW(),
    processed_at TIMESTAMP
    );

CREATE INDEX IF NOT EXISTS idx_bi_observations_business ON business_intel.data_observations(business_id);
CREATE INDEX IF NOT EXISTS idx_bi_observations_source ON business_intel.data_observations(source_type, source_name);
CREATE INDEX IF NOT EXISTS idx_bi_observations_collected ON business_intel.data_observations(collected_at DESC);

-- Collection task queue
CREATE TABLE IF NOT EXISTS business_intel.collection_tasks (
                                                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID REFERENCES business_intel.businesses(id),
    vertical_id UUID REFERENCES business_intel.business_verticals(id),

    task_type TEXT NOT NULL,                   -- 'initial_verification', 'price_refresh',
-- 'status_check', 'discovery'
    status TEXT DEFAULT 'pending',             -- 'pending', 'in_progress', 'completed',
-- 'failed', 'needs_review'
    priority INTEGER DEFAULT 5,               -- 1=highest

-- Scheduling
    scheduled_for TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,

    -- Links to agent execution
    orchestration_id UUID,

    -- Results
    result_summary JSONB,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,

    created_at TIMESTAMP DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_bi_tasks_pending ON business_intel.collection_tasks(priority, scheduled_for)
    WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_bi_tasks_business ON business_intel.collection_tasks(business_id);
CREATE INDEX IF NOT EXISTS idx_bi_tasks_status ON business_intel.collection_tasks(status);

-- =============================================
-- LAYER 4: Products / Medicines catalogue
-- =============================================

-- Products that can be compared across retailers and vets
CREATE TABLE IF NOT EXISTS business_intel.products (
                                                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vertical_id UUID REFERENCES business_intel.business_verticals(id),

    slug TEXT UNIQUE NOT NULL,                    -- 'm_cardisure_5_100' (stable ID for imports)
    name TEXT NOT NULL,                           -- 'Cardisure Tablets'
    dosage TEXT,                                  -- '5mg (100 tabs)'
    category TEXT,                                -- 'cardiac', 'anti-inflammatory', 'parasiticide'

-- Reference pricing (rough estimate, not from a specific source)
    typical_vet_price_gbp DECIMAL(10,2),

    -- Metadata
    requires_prescription BOOLEAN DEFAULT TRUE,
    species TEXT[],                                -- '{dog,cat}'
    active_ingredient TEXT,
    manufacturer TEXT,

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_bi_products_vertical ON business_intel.products(vertical_id);
CREATE INDEX IF NOT EXISTS idx_bi_products_name ON business_intel.products(name);
CREATE INDEX IF NOT EXISTS idx_bi_products_category ON business_intel.products(category) WHERE category IS NOT NULL;

-- Price observations for products from specific businesses (online retailers or vets)
CREATE TABLE IF NOT EXISTS business_intel.product_prices (
                                                             id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES business_intel.products(id),
    business_id UUID REFERENCES business_intel.businesses(id),

    price_gbp DECIMAL(10,2),
    price_qualifier TEXT,                         -- 'fixed', 'from', 'approximately'
    includes_vat BOOLEAN DEFAULT TRUE,
    in_stock BOOLEAN,
    product_url TEXT,                             -- direct link to this product on retailer site

-- Provenance
    source TEXT NOT NULL,                         -- 'website_scrape', 'manual_entry', 'seed_import'
    observed_at TIMESTAMP DEFAULT NOW(),

    -- Historical: latest per product+business is is_current=TRUE
    is_current BOOLEAN DEFAULT TRUE
    );

CREATE INDEX IF NOT EXISTS idx_bi_prodprices_product ON business_intel.product_prices(product_id);
CREATE INDEX IF NOT EXISTS idx_bi_prodprices_business ON business_intel.product_prices(business_id);
CREATE INDEX IF NOT EXISTS idx_bi_prodprices_current ON business_intel.product_prices(product_id, business_id) WHERE is_current = TRUE;

-- =============================================
-- LAYER 5: People / Contacts
-- =============================================

CREATE TABLE IF NOT EXISTS business_intel.people (
                                                     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Identity
    first_name TEXT,
    last_name TEXT,
    display_name TEXT,                        -- 'Dr Sarah Chen'
    slug TEXT UNIQUE,

    -- Contact (personal, not via a business)
    email TEXT,
    phone TEXT,
    linkedin_url TEXT,

    -- Professional
    qualifications TEXT[],                    -- '{BVSc, MRCVS, CertAVP}'
    specialisms TEXT[],                       -- '{feline, dermatology}'

-- Data quality
    verification_status TEXT DEFAULT 'unverified',
    last_verified_at TIMESTAMP,

    -- Self-service (future)
    claimed_by UUID,
    is_claimed BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
    );

-- People <-> Business relationships over time
CREATE TABLE IF NOT EXISTS business_intel.people_business_roles (
                                                                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    person_id UUID REFERENCES business_intel.people(id),
    business_id UUID REFERENCES business_intel.businesses(id),

    role_title TEXT,                           -- 'Head Vet', 'Practice Manager'
    role_type TEXT,                            -- 'clinical', 'management', 'owner', 'contact'

-- Time dimension
    started_at DATE,                          -- when they joined (if known)
    ended_at DATE,                            -- NULL = current
    is_current BOOLEAN DEFAULT TRUE,

    -- Provenance
    source TEXT,                              -- 'website_scrape', 'linkedin'
    source_url TEXT,
    observed_at TIMESTAMP DEFAULT NOW(),

    -- Contact-specific (for verticals where contact info matters)
    is_primary_contact BOOLEAN DEFAULT FALSE,
    contact_email TEXT,
    contact_phone TEXT,

    UNIQUE(person_id, business_id, role_title, started_at)
    );

CREATE INDEX IF NOT EXISTS idx_bi_pbr_person ON business_intel.people_business_roles(person_id);
CREATE INDEX IF NOT EXISTS idx_bi_pbr_business ON business_intel.people_business_roles(business_id);
CREATE INDEX IF NOT EXISTS idx_bi_pbr_current ON business_intel.people_business_roles(business_id) WHERE is_current = TRUE;

-- Now add the FK from data_observations to people
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_observations_person'
    ) THEN
ALTER TABLE business_intel.data_observations
    ADD CONSTRAINT fk_observations_person
        FOREIGN KEY (person_id) REFERENCES business_intel.people(id);
END IF;
END $$;

-- =============================================
-- Seed data: verticals
-- =============================================

INSERT INTO business_intel.business_verticals (slug, display_name, description, default_agent_type)
VALUES
    ('veterinary', 'Veterinary Practices', 'UK veterinary practices and clinics', 'vet-practice-verifier'),
    ('online-pharmacy', 'Online Vet Pharmacies', 'UK online veterinary medicine retailers', 'pharmacy-price-monitor'),
    ('seaweed-farming', 'Seaweed Farms', 'UK seaweed farming and harvesting companies', NULL)
    ON CONFLICT (slug) DO NOTHING;

COMMIT;