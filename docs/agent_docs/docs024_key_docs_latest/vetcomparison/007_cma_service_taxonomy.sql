-- ============================================================================
-- 007_cma_service_taxonomy.sql
-- ============================================================================
-- Seeds the CMA's mandated 36-item standard price list (final report Part B,
-- 24 Mar 2026, Table 3.1 ¶3.74) as canonical products rows. This taxonomy is
-- the mapping target for every scraped, claimed or (later) RCVS-feed price.
--
-- Source of truth: CMA final decision report Part B. The remedies Order was
-- NOT yet made when this was written (due by 23 Sep 2026) — re-verify item
-- wording and the pet bands against the final Order before public launch.
-- Known ambiguity: Part B ¶3.88 lists six pet categories (cat, small dog
-- <10kg, medium 10-25kg, large 25-40kg, XL 40-60kg, giant >60kg) while the
-- gov.uk guidance table merges cat with small dog into one <10kg band. We
-- model the six categories; bands may be combined per ¶3.88.
--
-- WHAT THIS DOES
--   (A) Adds regulatory-metadata columns to business_intel.products:
--       cma_item, cma_category, checkbox_disclosures.
--   (B) Adds pet_band to business_intel.product_prices (default 'any').
--   (C) Upserts the 36 CMA items as products rows (kind='service',
--       vertical=veterinary, slug='cma-<category-n>-<item>').
--
-- IDEMPOTENT: safe to re-run; upserts refresh names/categories in place.
-- ============================================================================

BEGIN;

-- (A) regulatory metadata on products -----------------------------------------
ALTER TABLE business_intel.products
    ADD COLUMN IF NOT EXISTS cma_item boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS cma_category smallint,
    ADD COLUMN IF NOT EXISTS checkbox_disclosures text[];

COMMENT ON COLUMN business_intel.products.cma_item IS
    'True for the 36 services on the CMA mandated standard price list '
    '(final report Part B Table 3.1). These rows are the canonical mapping '
    'target for scraped/claimed/feed prices.';
COMMENT ON COLUMN business_intel.products.cma_category IS
    '1=Consultation and preventative care, 2=Prescription, dispensing and '
    'administration, 3=Surgeries and treatments, 4=Diagnostics and '
    'laboratory tests, 5=End-of-life care.';
COMMENT ON COLUMN business_intel.products.checkbox_disclosures IS
    'For the four starred surgical items: the standard inclusion checkboxes '
    'the CMA requires (¶3.92) instead of free text.';

CREATE INDEX IF NOT EXISTS idx_bi_products_cma
    ON business_intel.products (cma_category) WHERE cma_item;

-- (B) pet band dimension on price observations --------------------------------
ALTER TABLE business_intel.product_prices
    ADD COLUMN IF NOT EXISTS pet_band text NOT NULL DEFAULT 'any';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'product_prices_pet_band_check'
          AND conrelid = 'business_intel.product_prices'::regclass
    ) THEN
        ALTER TABLE business_intel.product_prices
            ADD CONSTRAINT product_prices_pet_band_check
            CHECK (pet_band IN ('any','cat','dog_small','dog_medium',
                                'dog_large','dog_xlarge','dog_giant'));
    END IF;
END $$;

COMMENT ON COLUMN business_intel.product_prices.pet_band IS
    'CMA standardised pet category the price applies to (Part B ¶3.88): '
    'cat, dog_small <10kg, dog_medium 10-25kg, dog_large 25-40kg, '
    'dog_xlarge 40-60kg, dog_giant >60kg; ''any'' where the price does not '
    'vary by pet (bands may be combined per ¶3.88).';

-- (C) the 36 items -------------------------------------------------------------
WITH vet AS (
    SELECT id FROM business_intel.business_verticals WHERE slug = 'veterinary'
),
items(slug, name, cat_n, cat_name, checkboxes) AS (VALUES
    -- Category 1 — Consultation and preventative care (12)
    ('cma-1-first-consultation',            'First consultation (incl. appointment length)',              1, 'Consultation and preventative care', NULL::text[]),
    ('cma-1-repeat-consultation',           'Repeat consultation',                                        1, 'Consultation and preventative care', NULL),
    ('cma-1-out-of-hours-consultation',     'Out-of-hours consultation (incl. provider if outsourced)',   1, 'Consultation and preventative care', NULL),
    ('cma-1-nurse-consultation',            'Nurse consultation (incl. appointment length)',              1, 'Consultation and preventative care', NULL),
    ('cma-1-nail-clipping',                 'Nail clipping',                                              1, 'Consultation and preventative care', NULL),
    ('cma-1-anal-gland-expression',         'Anal gland expression',                                      1, 'Consultation and preventative care', NULL),
    ('cma-1-microchipping',                 'Microchipping',                                              1, 'Consultation and preventative care', NULL),
    ('cma-1-animal-health-certificate',     'Animal health certificate (incl. additional-animal charges)',1, 'Consultation and preventative care', NULL),
    ('cma-1-vaccinations-primary-course',   'Vaccinations — primary course (incl. consultation)',         1, 'Consultation and preventative care', NULL),
    ('cma-1-vaccinations-booster',          'Vaccinations — booster (incl. consultation)',                1, 'Consultation and preventative care', NULL),
    ('cma-1-kennel-cough-vaccination',      'Kennel cough vaccination (incl. consultation)',              1, 'Consultation and preventative care', NULL),
    ('cma-1-pet-care-plan',                 'Pet care plan (monthly cost, incl. inclusions link)',        1, 'Consultation and preventative care', NULL),
    -- Category 2 — Prescription, dispensing and administration (6)
    ('cma-2-prescription-fee',              'Prescription fee',                                           2, 'Prescription, dispensing and administration', NULL),
    ('cma-2-prescription-additional-items', 'Additional items prescribed in same consultation',           2, 'Prescription, dispensing and administration', NULL),
    ('cma-2-prescription-fee-repeat',       'Repeat prescription fee',                                    2, 'Prescription, dispensing and administration', NULL),
    ('cma-2-dispensing-fee',                'Dispensing fee',                                             2, 'Prescription, dispensing and administration', NULL),
    ('cma-2-injection-administration',      'Medicine administration — injection (skin or muscle)',       2, 'Prescription, dispensing and administration', NULL),
    ('cma-2-insurance-administration',      'Insurance administration fee',                               2, 'Prescription, dispensing and administration', NULL),
    -- Category 3 — Surgeries and treatments (6; * = checkbox disclosures)
    ('cma-3-dental-assessment',             'Dental assessment (exam, scale and polish)',                 3, 'Surgeries and treatments', ARRAY['general_anaesthesia_or_sedation','pain_relief','hospitalisation','pre_op_bloods','post_op_checks']),
    ('cma-3-castration',                    'Castration',                                                 3, 'Surgeries and treatments', ARRAY['general_anaesthesia_or_sedation','pain_relief','hospitalisation','pre_op_bloods','post_op_checks']),
    ('cma-3-spay-traditional',              'Spay (traditional/open)',                                    3, 'Surgeries and treatments', ARRAY['general_anaesthesia_or_sedation','pain_relief','hospitalisation','pre_op_bloods','post_op_checks']),
    ('cma-3-spay-laparoscopic',             'Spay (laparoscopic/keyhole)',                                3, 'Surgeries and treatments', ARRAY['general_anaesthesia_or_sedation','pain_relief','hospitalisation','pre_op_bloods','post_op_checks']),
    ('cma-3-physiotherapy-session',         'Physiotherapy session',                                      3, 'Surgeries and treatments', NULL),
    ('cma-3-laser-therapy-session',         'Laser therapy session',                                      3, 'Surgeries and treatments', NULL),
    -- Category 4 — Diagnostics and laboratory tests (9; all incl. interpretation)
    ('cma-4-x-ray',                         'X-ray (incl. sedation/GA and three images)',                 4, 'Diagnostics and laboratory tests', NULL),
    ('cma-4-ultrasound-abdominal',          'Ultrasound — abdominal (single organ)',                      4, 'Diagnostics and laboratory tests', NULL),
    ('cma-4-ultrasound-echocardiogram',     'Ultrasound — echocardiogram',                                4, 'Diagnostics and laboratory tests', NULL),
    ('cma-4-cytology-ear-swab',             'Cytology — ear swab',                                        4, 'Diagnostics and laboratory tests', NULL),
    ('cma-4-cytology-fna',                  'Cytology — fine needle aspiration',                          4, 'Diagnostics and laboratory tests', NULL),
    ('cma-4-urine-screen',                  'Basic urine screen (dipstick + specific gravity)',           4, 'Diagnostics and laboratory tests', NULL),
    ('cma-4-ct-scan',                       'CT scan per body part (incl. GA and images)',                4, 'Diagnostics and laboratory tests', NULL),
    ('cma-4-mri-scan',                      'MRI scan per body part (incl. GA and images)',               4, 'Diagnostics and laboratory tests', NULL),
    ('cma-4-pre-anaesthetic-blood-test',    'Pre-surgical/pre-anaesthetic blood test',                    4, 'Diagnostics and laboratory tests', NULL),
    -- Category 5 — End-of-life care (3)
    ('cma-5-euthanasia',                    'Euthanasia',                                                 5, 'End-of-life care', NULL),
    ('cma-5-cremation-communal',            'Cremation — communal',                                       5, 'End-of-life care', NULL),
    ('cma-5-cremation-individual',          'Cremation — individual',                                     5, 'End-of-life care', NULL)
)
INSERT INTO business_intel.products
    (vertical_id, slug, name, category, kind, requires_prescription, is_active,
     cma_item, cma_category, checkbox_disclosures)
SELECT vet.id, i.slug, i.name, i.cat_name, 'service', false, true,
       true, i.cat_n, i.checkboxes
FROM items i CROSS JOIN vet
ON CONFLICT (slug) DO UPDATE SET
    name                 = EXCLUDED.name,
    category             = EXCLUDED.category,
    cma_item             = true,
    cma_category         = EXCLUDED.cma_category,
    checkbox_disclosures = EXCLUDED.checkbox_disclosures,
    updated_at           = NOW();

-- Verification — expect 36 total: 12/6/6/9/3 across categories 1-5 ------------
SELECT cma_category, COUNT(*) AS items,
       COUNT(*) FILTER (WHERE checkbox_disclosures IS NOT NULL) AS with_checkboxes
FROM business_intel.products
WHERE cma_item
GROUP BY cma_category ORDER BY cma_category;

SELECT COUNT(*) AS total_cma_items FROM business_intel.products WHERE cma_item;

COMMIT;
