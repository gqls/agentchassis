ALTER TABLE sites ADD COLUMN company_name VARCHAR(255);
ALTER TABLE sites ADD COLUMN tagline TEXT;
ALTER TABLE sites ADD COLUMN email VARCHAR(255);
ALTER TABLE sites ADD COLUMN phone VARCHAR(100);
ALTER TABLE sites ADD COLUMN logo_text VARCHAR(255);
ALTER TABLE sites ADD COLUMN contact_address TEXT;

-- Migrate existing data from content_data
UPDATE sites SET
                 company_name = content_data->>'company_name',
    tagline = content_data->>'tagline',
    email = content_data->>'email',
    phone = content_data->>'phone'
WHERE content_data IS NOT NULL;

-- ============================================================
-- 5. Add granular contact fields to sites table
-- ============================================================
ALTER TABLE sites ADD COLUMN IF NOT EXISTS company_name VARCHAR(255);
ALTER TABLE sites ADD COLUMN IF NOT EXISTS tagline TEXT;
ALTER TABLE sites ADD COLUMN IF NOT EXISTS email VARCHAR(255);
ALTER TABLE sites ADD COLUMN IF NOT EXISTS phone VARCHAR(100);
ALTER TABLE sites ADD COLUMN IF NOT EXISTS logo_url VARCHAR(500);
ALTER TABLE sites ADD COLUMN IF NOT EXISTS logo_text VARCHAR(255);

COMMENT ON COLUMN sites.company_name IS 'Display name for the company';
COMMENT ON COLUMN sites.tagline IS 'Company tagline/slogan';
COMMENT ON COLUMN sites.email IS 'Primary contact email';
COMMENT ON COLUMN sites.phone IS 'Primary contact phone';
COMMENT ON COLUMN sites.logo_url IS 'URL to logo image';
COMMENT ON COLUMN sites.logo_text IS 'Text to display if no logo image';


        --

        -- ============================================================
-- 7. Backfill site contact fields from content_data
-- ============================================================
UPDATE sites
SET
    company_name = COALESCE(
            company_name,
            content_data->>'company_name',
        content_data->>'business_name',
        name
    ),
    tagline = COALESCE(
            tagline,
            content_data->>'tagline',
        content_data->>'slogan'
    ),
    email = COALESCE(
            email,
            content_data->>'email',
        content_data->>'contact_email'
    ),
    phone = COALESCE(
            phone,
            content_data->>'phone',
        content_data->>'contact_phone'
    ),
    logo_text = COALESCE(
            logo_text,
            content_data->>'logo_text',
        company_name,
        name
    )
WHERE company_name IS NULL
   OR tagline IS NULL
   OR email IS NULL;


==

-- site area is a site section eg design, graphic design, illustration, architecture etc

-- ============================================================
-- 8. Create default area for existing sites
-- ============================================================
INSERT INTO site_areas (site_id, name, display_name, url_prefix, is_default)
SELECT id, 'main', 'Main', '/', true
FROM sites
WHERE NOT EXISTS (
    SELECT 1 FROM site_areas WHERE site_areas.site_id = sites.id AND is_default = true
)
    ON CONFLICT (site_id, name) DO NOTHING;