-- med_pricing_seed_retailers.sql
-- Seed the 4 target retailers.
-- Note: Pet Drugs Online URL structure changed — no longer /prescriptions/m/, now flat /{slug}
-- Category URLs updated to reflect current site structure.

INSERT INTO business_intel.med_retailers (id, name, domain, group_name, base_url, category_urls, scrape_config)
VALUES
    (
        'pet_drugs_online',
        'Pet Drugs Online',
        'petdrugsonline.co.uk',
        'IVC Evidensia',
        'https://www.petdrugsonline.co.uk',
        ARRAY[
            'https://www.petdrugsonline.co.uk/dog-prescriptions',
        'https://www.petdrugsonline.co.uk/cat-prescriptions'
            ],
        '{
            "price_selector": "Size Options list items",
            "has_tvp": true,
            "notes": "URL structure changed from /prescriptions/m/{slug} to /{slug}. Products have size variants with TVP."
        }'::jsonb
    ),
    (
        'animed_direct',
        'Animed Direct',
        'animed.co.uk',
        'CVS Group',
        'https://www.animed.co.uk',
        ARRAY[
            'https://www.animed.co.uk/prescriptions'
            ],
        '{
            "notes": "Domain is animed.co.uk not animeddirect.co.uk — updated from plan."
        }'::jsonb
    ),
    (
        'viovet',
        'VioVet',
        'viovet.co.uk',
        'Covetrus',
        'https://www.viovet.co.uk',
        ARRAY[
            'https://www.viovet.co.uk/Dogs/Pharmacy-Dog/c6/',
        'https://www.viovet.co.uk/Cats/Pharmacy-Cat/c158/'
            ],
        '{}'::jsonb
    ),
    (
        'hyperdrug',
        'Hyperdrug',
        'hyperdrug.co.uk',
        'Independent',
        'https://www.hyperdrug.co.uk',
        ARRAY[
            'https://www.hyperdrug.co.uk/pet-prescriptions/'
            ],
        '{}'::jsonb
    )
    ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
                            domain = EXCLUDED.domain,
                            group_name = EXCLUDED.group_name,
                            base_url = EXCLUDED.base_url,
                            category_urls = EXCLUDED.category_urls,
                            scrape_config = EXCLUDED.scrape_config,
                            updated_at = NOW();

