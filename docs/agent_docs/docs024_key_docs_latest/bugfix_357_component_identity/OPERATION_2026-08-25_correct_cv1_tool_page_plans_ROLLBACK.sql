-- Exact inverse of the plan edit in OPERATION_2026-08-25_correct_cv1_tool_page_plans.sql.
-- Values captured live 2026-08-25 12:1xZ before the edit.
--
-- It restores the PLANS only. It deliberately does NOT delete any page_components row the
-- re-queued recreation may have written by the time you run this: those rows are true and
-- regenerable (content_data.body reproduces rendered_html byte for byte), and deleting a
-- correct row to undo a plan edit would be the larger harm. If you also want the rows gone,
-- delete them explicitly and say so.
BEGIN;
UPDATE pages SET sections = '["hero", "features", "call-to-action", "contact-form"]'::jsonb, updated_at = now()
 WHERE id = 'aef1aa49-3778-4932-b7ea-1bc298230dcc';
UPDATE pages SET sections = '["generic-text-block", "features", "call-to-action"]'::jsonb, updated_at = now()
 WHERE id = 'f763ca0e-a5ad-4d25-9e6c-37d158c13493';
COMMIT;
