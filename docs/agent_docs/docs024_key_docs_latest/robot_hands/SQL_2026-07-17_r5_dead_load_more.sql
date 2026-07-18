-- SQL_2026-07-17_r5_dead_load_more.sql
--
-- R5 of HANDOFF_2026-07-17_robot_hands_site_fixes.md: the content-listing
-- component renders a "Load More" button with NO behaviour wired anywhere in
-- the fleet ({{if .show_load_more}}<button>…</button>{{end}}, schema fallback
-- true). With 3 real articles on robot-hands, hiding it is the honest v1
-- (handoff's sanctioned option; pagination JS is future work).
--
-- Two changes:
--   1. robot-hands learning-center-hub instance: content_data
--      show_load_more true → false (it opted in explicitly).
--   2. Component schema fallback true → false — a dead control must be
--      explicit opt-in, not default-on. Affects only instances WITHOUT an
--      explicit value, at their next re-render. Explicit trues elsewhere
--      (dartsonline index, idea.uk guides-index) are other sessions' sites —
--      left alone, flagged in the handoff doc.

\set ON_ERROR_STOP on

CREATE TABLE IF NOT EXISTS content_components_backup_20260717_r5 AS
SELECT * FROM content_components WHERE id = 'aa3e4b68-bcea-49ca-890a-c111acefa551';
SELECT count(*) AS bak_component FROM content_components_backup_20260717_r5;

BEGIN;

UPDATE page_components
SET content_data = jsonb_set(content_data, '{show_load_more}', 'false'::jsonb)
WHERE id = '0253c122-d686-4353-9c01-cc65a588962b';

UPDATE content_components
SET input_schema = jsonb_set(input_schema, '{fields,show_load_more,fallback}', 'false'::jsonb),
    updated_at = now()
WHERE id = 'aa3e4b68-bcea-49ca-890a-c111acefa551';

DO $verify$
DECLARE v_slm text; v_fb text;
BEGIN
    SELECT content_data->>'show_load_more' INTO v_slm
    FROM page_components WHERE id = '0253c122-d686-4353-9c01-cc65a588962b';
    IF v_slm <> 'false' THEN RAISE EXCEPTION 'hub instance still %', v_slm; END IF;

    SELECT input_schema->'fields'->'show_load_more'->>'fallback' INTO v_fb
    FROM content_components WHERE id = 'aa3e4b68-bcea-49ca-890a-c111acefa551';
    IF v_fb <> 'false' THEN RAISE EXCEPTION 'component fallback still %', v_fb; END IF;

    RAISE NOTICE 'R5 applied: hub opt-out + component fallback now false';
END
$verify$;

COMMIT;
