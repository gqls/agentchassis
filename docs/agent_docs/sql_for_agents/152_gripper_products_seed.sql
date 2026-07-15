-- 152: real gripper product rows for robot-hands (Phase 4 data)
--
-- Every row below was fetched directly from a manufacturer or distributor
-- page on 2026-07-14 (via WebFetch, not LLM training-knowledge recall) —
-- four from a primary manufacturer source, one (Festo) from a distributor
-- listing after festo.com blocked automated fetch four times. source_url
-- and verified_date are stored in content_data and rendered on every card
-- by gripper-spec-sheet (151). No field was invented: manufacturer pages
-- disclose different spec subsets, and only fields actually stated on the
-- fetched page are populated — the component renders each field
-- conditionally, so gaps show as an absent row, never a fabricated one.
--
-- Requires 151 (gripper-spec-sheet component + query.products resolver)
-- applied first.
--
-- Verify after applying:
--   SELECT name, category, specifications->>'manufacturer' AS mfr,
--          content_data->>'source_url' AS source
--   FROM products WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92';
--   -- expect 5 rows

BEGIN;

INSERT INTO products (site_id, name, slug, category, subcategory, specifications, content_data, status)
VALUES
(
    '00ff3af5-dad8-4770-9f70-3edc267a3c92',
    'Schunk EGP 40-N-S-B',
    'schunk-egp-40-n-s-b',
    'gripper',
    'Electric parallel gripper (small components)',
    '{
        "manufacturer": "Schunk",
        "stroke": "6 mm per jaw",
        "gripping_force": "30 N",
        "payload": "0.15 kg (recommended workpiece weight)",
        "weight": "0.3 kg",
        "ip_rating": "IP30",
        "interface": "Digital I/O",
        "voltage": "24 V DC"
    }'::jsonb,
    '{
        "source_url": "https://schunk.com/de/en/gripping-systems/parallel-gripper/egp/egp-40-n-s-b/p/000000000000310942",
        "verified_date": "2026-07-14"
    }'::jsonb,
    'active'
),
(
    '00ff3af5-dad8-4770-9f70-3edc267a3c92',
    'OnRobot 2FG7',
    'onrobot-2fg7',
    'gripper',
    'Electric parallel gripper (cobot)',
    '{
        "manufacturer": "OnRobot",
        "stroke": "up to 73 mm external grip range",
        "gripping_force": "20–140 N",
        "payload": "11 kg",
        "ip_rating": "IP67"
    }'::jsonb,
    '{
        "source_url": "https://onrobot.com/en/products/2fg7",
        "verified_date": "2026-07-14"
    }'::jsonb,
    'active'
),
(
    '00ff3af5-dad8-4770-9f70-3edc267a3c92',
    'Robotiq 2F-85',
    'robotiq-2f-85',
    'gripper',
    'Adaptive parallel gripper (cobot)',
    '{
        "manufacturer": "Robotiq",
        "stroke": "85 mm",
        "gripping_force": "20–235 N",
        "payload": "5 kg",
        "weight": "925 g",
        "voltage": "24 V DC ±10%"
    }'::jsonb,
    '{
        "source_url": "https://assets.robotiq.com/website-assets/support_documents/document/online/2F-85_2F-140_TM_InstructionManual_HTML5_20190503.zip/2F-85_2F-140_TM_InstructionManual_HTML5/Content/6.%20Specifications.htm",
        "verified_date": "2026-07-14"
    }'::jsonb,
    'active'
),
(
    '00ff3af5-dad8-4770-9f70-3edc267a3c92',
    'Zimmer Group GEP5010IO-00-A',
    'zimmer-group-gep5010io-00-a',
    'gripper',
    'Electric parallel gripper (Series GEP5000)',
    '{
        "manufacturer": "Zimmer Group",
        "stroke": "10 mm per jaw",
        "gripping_force": "1520 N",
        "weight": "1.6 kg",
        "ip_rating": "IP64",
        "interface": "I/O (IO-Link option)",
        "voltage": "24 V"
    }'::jsonb,
    '{
        "source_url": "https://www.zimmer-group.com/en-us/technologies-components/components/handling-technology/grippers/electric/2-jaw-parallel-grippers/series-gep5000/produkte/gep5010io-00-a",
        "verified_date": "2026-07-14"
    }'::jsonb,
    'active'
),
(
    '00ff3af5-dad8-4770-9f70-3edc267a3c92',
    'Festo EHPS-20-A-LK',
    'festo-ehps-20-a-lk',
    'gripper',
    'Electric parallel gripper (sensor-ready, IO-Link)',
    '{
        "manufacturer": "Festo",
        "stroke": "13 mm per jaw",
        "gripping_force": "218 N"
    }'::jsonb,
    '{
        "source_url": "https://us.rs-online.com/product/festo/ehps-20-a-lk/71946902/",
        "verified_date": "2026-07-14"
    }'::jsonb,
    'active'
);

COMMIT;
