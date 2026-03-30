INSERT INTO business_intel.med_retailer_listings (retailer_id, product_id, retailer_url, retailer_product_name, match_confidence, match_method)
VALUES
    ('hyperdrug', 'm_metacam_dog_oral', 'https://www.hyperdrug.co.uk/metacam-1-5mg-ml-oral-suspension-for-dogs/', 'Metacam 1.5mg/ml Oral Suspension for Dogs', 1.0, 'manual'),
    ('animed_direct', 'm_metacam_dog_oral', 'https://www.animed.co.uk/metacam-1-5mg-ml-oral-suspension-for-dogs', 'Metacam 1.5mg/ml Oral Suspension for Dogs', 1.0, 'manual'),
    ('viovet', 'm_metacam_dog_oral', 'https://www.viovet.co.uk/Metacam-Oral-Suspension/c5/', 'Metacam Oral Suspension', 1.0, 'manual')
    ON CONFLICT (retailer_id, retailer_url) DO NOTHING;