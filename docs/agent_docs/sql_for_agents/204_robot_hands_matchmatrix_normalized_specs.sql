-- 204_robot_hands_matchmatrix_normalized_specs.sql
--
-- Gripper dossier pilot (robot_hands_gripper_dossier workstream): seed the
-- normalised machine-readable spec block `content_data.matchmatrix` on the
-- 10 robot-hands.com gripper products, so the server-side score_grippers
-- action reads the same verified figures the live MatchMatrix v2 tool
-- carries in its client-side GRIPPERS array (content_components
-- fdfeaa7a-be17-46f9-9ecb-3ccba17c8ebc). Values transcribed 1:1 from that
-- array — same figures, same provenance (each row already carries
-- content_data.source_url + verified_date from the R7 datasheet work).
--
-- Contract: a missing figure is ABSENT (null) in the block — "not published
-- by the manufacturer", never inferred. Per-jaw strokes are already doubled
-- to total opening, matching the tool.
--
-- Merge discipline: content_data || jsonb_build_object(...) — adds the one
-- key, never clobbers existing keys (source_url, verified_date, etc.).
-- Idempotent: re-running overwrites only the matchmatrix key with identical
-- values. Names no Go actions — safe to apply BEFORE the image roll.
--
-- Verify after:
--   SELECT count(*) FROM products
--   WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92'
--     AND content_data ? 'matchmatrix';   -- must be 10

BEGIN;

UPDATE products SET content_data = content_data || jsonb_build_object('matchmatrix', '{
  "tech": "jaw", "tech_label": "Electric parallel-jaw", "maker": "Schunk",
  "force_n": 30, "force_text": "30 N",
  "stroke_mm_total": 12, "stroke_text": "6 mm per jaw (12 mm total)",
  "payload_kg": 0.15, "payload_text": "0.15 kg (recommended workpiece weight)",
  "ip": 30, "ip_text": "IP30",
  "extras": [["Weight","0.3 kg"],["Supply","24 V DC"],["Interface","Digital I/O"]]
}'::jsonb), updated_at = NOW()
WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND slug='schunk-egp-40-n-s-b';

UPDATE products SET content_data = content_data || jsonb_build_object('matchmatrix', '{
  "tech": "jaw", "tech_label": "Electric parallel-jaw", "maker": "OnRobot",
  "force_n": 140, "force_text": "20 N to 140 N",
  "stroke_mm_total": 73, "stroke_text": "up to 73 mm external grip range",
  "payload_kg": 11, "payload_text": "11 kg (24.3 lb)",
  "ip": 67, "ip_text": "IP67",
  "extras": []
}'::jsonb), updated_at = NOW()
WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND slug='onrobot-2fg7';

UPDATE products SET content_data = content_data || jsonb_build_object('matchmatrix', '{
  "tech": "jaw", "tech_label": "Electric parallel-jaw", "maker": "Robotiq",
  "force_n": 235, "force_text": "20 to 235 N",
  "stroke_mm_total": 85, "stroke_text": "85 mm",
  "payload_kg": 5, "payload_text": "5 kg",
  "extras": [["Weight","925 g"],["Supply","24 V DC ±10%"]]
}'::jsonb), updated_at = NOW()
WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND slug='robotiq-2f-85';

UPDATE products SET content_data = content_data || jsonb_build_object('matchmatrix', '{
  "tech": "jaw", "tech_label": "Electric parallel-jaw", "maker": "Zimmer Group",
  "force_n": 1520, "force_text": "1520 N",
  "stroke_mm_total": 20, "stroke_text": "10 mm per jaw (20 mm total)",
  "ip": 64, "ip_text": "IP64",
  "extras": [["Weight","1.6 kg"],["Supply","24 V"],["Interface","I/O (IO-Link option)"]]
}'::jsonb), updated_at = NOW()
WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND slug='zimmer-group-gep5010io-00-a';

UPDATE products SET content_data = content_data || jsonb_build_object('matchmatrix', '{
  "tech": "jaw", "tech_label": "Electric parallel-jaw", "maker": "Festo",
  "force_n": 218, "force_text": "218 N",
  "stroke_mm_total": 26, "stroke_text": "13 mm per jaw (26 mm total)",
  "extras": []
}'::jsonb), updated_at = NOW()
WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND slug='festo-ehps-20-a-lk';

UPDATE products SET content_data = content_data || jsonb_build_object('matchmatrix', '{
  "tech": "jaw", "tech_label": "Pneumatic parallel-jaw", "maker": "Festo",
  "force_n": 34.5, "force_text": "34.5 N per jaw closing (39 N opening) at 6 bar",
  "stroke_mm_total": 6, "stroke_text": "3 mm per jaw (6 mm total)",
  "extras": [["Weight","67 g"],["Operating pressure","2 to 8 bar"],["Repeat accuracy","0.02 mm"],["Supply","Compressed air"]]
}'::jsonb), updated_at = NOW()
WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND slug='festo-dhps-10-a';

UPDATE products SET content_data = content_data || jsonb_build_object('matchmatrix', '{
  "tech": "vacuum", "tech_label": "Electric vacuum", "maker": "OnRobot",
  "payload_kg": 15, "payload_text": "15 kg (35 lb)",
  "note": "Suction hold — needs a surface a vacuum cup can seal against; porous or heavily perforated surfaces reduce holding force. Built-in pump, no external air supply.",
  "extras": [["Zones","Dual, independently switchable"]]
}'::jsonb), updated_at = NOW()
WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND slug='onrobot-vg10';

UPDATE products SET content_data = content_data || jsonb_build_object('matchmatrix', '{
  "tech": "adhesive", "tech_label": "Adhesive (gecko)", "maker": "OnRobot",
  "payload_kg": 5, "payload_text": "5 kg",
  "note": "Van der Waals adhesion — requires clean, smooth, dry, flat surfaces; not suitable for greasy, wet or dusty parts. No air and no electricity required.",
  "extras": []
}'::jsonb), updated_at = NOW()
WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND slug='onrobot-gecko-sp5';

UPDATE products SET content_data = content_data || jsonb_build_object('matchmatrix', '{
  "tech": "soft", "tech_label": "Soft silicone", "maker": "OnRobot",
  "payload_kg": 2.2, "payload_text": "2.2 kg (depends on shape, softness and friction of the part)",
  "grip_min_mm": 11, "grip_max_mm": 118, "grip_text": "11 to 118 mm (cup-dependent)",
  "note": "Payload depends on part geometry — the rating is an upper bound, not a guarantee. Food-grade silicone cups; no external air supply.",
  "extras": [["Material","Food-grade silicone"]]
}'::jsonb), updated_at = NOW()
WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND slug='onrobot-soft-gripper-sg';

UPDATE products SET content_data = content_data || jsonb_build_object('matchmatrix', '{
  "tech": "magnetic", "tech_label": "Permanent magnetic", "maker": "Schmalz",
  "force_n": 385, "force_text": "560 N (without friction ring), 385 N (with friction ring)",
  "note": "Permanent-magnet surface hold on ferromagnetic material only. Assessed against the lower published figure (385 N, with friction ring). Workpiece temperatures up to 350 °C.",
  "extras": [["Diameter","50 mm"],["Max workpiece temp","350 °C"]]
}'::jsonb), updated_at = NOW()
WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND slug='schmalz-sgm-hp-50';

-- Sanity: exactly 10 blocks now present.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM products
  WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND content_data ? 'matchmatrix';
  IF n <> 10 THEN
    RAISE EXCEPTION 'expected 10 matchmatrix blocks, found %', n;
  END IF;
END $$;

COMMIT;
