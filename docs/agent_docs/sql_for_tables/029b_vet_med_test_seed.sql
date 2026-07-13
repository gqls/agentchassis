-- med_pricing_test_seed.sql
-- Insert a handful of test products and Pet Drugs Online listings
-- for manual extraction testing.
-- Run after med_pricing_schema_migration.sql and med_pricing_seed_retailers.sql.

-- ============================================================================
-- Test products — 5 common vet medicines
-- ============================================================================
INSERT INTO business_intel.med_products (id, name, generic_name, brand, manufacturer, species, category, form, strength, prescription_required)
VALUES
    ('m_metacam_dog_oral', 'Metacam Oral Suspension (Dog)', 'Meloxicam', 'Metacam', 'Boehringer Ingelheim', '{dog}', 'nsaid', 'oral_suspension', '1.5mg/ml', true),
    ('m_apoquel_dog_16mg', 'Apoquel Film-Coated Tablets 16mg (Dog)', 'Oclacitinib', 'Apoquel', 'Zoetis', '{dog}', 'dermatology', 'tablet', '16mg', true),
    ('m_metacam_cat_oral', 'Metacam Oral Suspension (Cat)', 'Meloxicam', 'Metacam', 'Boehringer Ingelheim', '{cat}', 'nsaid', 'oral_suspension', '0.5mg/ml', true),
    ('m_nexgard_spectra_dog', 'NexGard Spectra (Dog)', 'Afoxolaner/Milbemycin', 'NexGard Spectra', 'Boehringer Ingelheim', '{dog}', 'antiparasitic', 'chewable_tablet', NULL, true),
    ('m_synulox_dog_cat', 'Synulox Palatable Tablets', 'Amoxicillin/Clavulanic Acid', 'Synulox', 'Zoetis', '{dog,cat}', 'antibiotic', 'tablet', '250mg', true)
    ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- Test listings — Pet Drugs Online URLs for the 5 products above
-- ============================================================================
INSERT INTO business_intel.med_retailer_listings (retailer_id, product_id, retailer_url, retailer_product_name, match_confidence, match_method)
VALUES
    ('pet_drugs_online', 'm_metacam_dog_oral', 'https://www.petdrugsonline.co.uk/metacam-oral-suspension-for-dogs', 'Metacam Oral Suspension for Dogs 1.5mg/ml', 1.0, 'manual'),
    ('pet_drugs_online', 'm_apoquel_dog_16mg', 'https://www.petdrugsonline.co.uk/apoquel-16mg', 'Apoquel Film-Coated Tablets for Dogs 16mg', 1.0, 'manual'),
    ('pet_drugs_online', 'm_metacam_cat_oral', 'https://www.petdrugsonline.co.uk/metacam-oral-suspension-for-cats', 'Metacam Oral Suspension for Cats 0.5mg/ml', 1.0, 'manual'),
    ('pet_drugs_online', 'm_nexgard_spectra_dog', 'https://www.petdrugsonline.co.uk/nexgard-spectra-for-dogs', 'NexGard Spectra for Dogs', 1.0, 'manual'),
    ('pet_drugs_online', 'm_synulox_dog_cat', 'https://www.petdrugsonline.co.uk/synulox-palatable-tablets', 'Synulox Palatable Tablets', 1.0, 'manual')
    ON CONFLICT (retailer_id, retailer_url) DO NOTHING;
