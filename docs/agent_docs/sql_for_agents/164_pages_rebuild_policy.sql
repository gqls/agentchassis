-- 164_pages_rebuild_policy.sql — page-ownership marker (Experience Loop guard
-- rail 1, RUNBOOK T2.1; mechanises TL-001).
--
-- A page marked rebuild_policy='owned' belongs to a tool/widget or is a
-- runtime-fill shell: the generic rebuild paths (reconcile -> needs_page ->
-- page-build-handler -> save_page_sections DELETE+reinsert) clobber it — the
-- vonc arena clobber is the proof. The marker makes the refusal mechanical:
--   * ReconcileSitePlanAction emits owned_page_review (needs_human_review, no
--     handler) instead of needs_page for owned/tool-role pages — retires the
--     manual "park needs_page:provocation after every reconcile" step;
--   * SavePageSectionsAction hard-refuses a generic section save on an owned
--     page (apply_section_edit and the tool pipeline remain the edit paths);
--   * page_rerender / assemble (re-assembly of EXISTING page_components) is
--     deliberately NOT gated — it is how owned pages deploy.
--
-- DB is live immediately; the Go refusals activate on the image carrying them
-- (safe on the current image: old code never reads the column).
--
-- Applied out of band (psql -f + ledger row same sitting per
-- bugs_open/aaa_fails_to_mend/007).

BEGIN;

ALTER TABLE pages ADD COLUMN rebuild_policy text NOT NULL DEFAULT 'generic'
  CONSTRAINT pages_rebuild_policy_check CHECK (rebuild_policy IN ('generic', 'owned'));

-- Every tool page fleet-wide is tool-owned by definition.
UPDATE pages SET rebuild_policy = 'owned' WHERE page_type = 'tool';

-- vonc runtime-fill shell (archive index hydrated client-side) and the parked
-- per-provocation page the experience spec will define (T4.3): both must never
-- receive a generic build.
UPDATE pages SET rebuild_policy = 'owned'
WHERE site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
  AND name IN ('provocations-index', 'provocation');

INSERT INTO doc_notes (id, subject_type, subject_key, body, categories, source, created_by)
VALUES (
  gen_random_uuid(),
  'pipeline', 'build',
  '## pages.rebuild_policy — page-ownership marker (guard rail 1)
Observed: nothing structural stopped a generic rebuild clobbering a tool-owned page (TL-001 was a habit-rule; the vonc arena clobber and the standing manual park of needs_page:provocation are the proof).
Fix: pages.rebuild_policy generic|owned (migration 164); owned seeded for all page_type=tool pages fleet-wide + vonc provocations-index/provocation. Go refusals (reconcile emits owned_page_review instead of needs_page; save_page_sections hard-refuses) ride the next chassis image.
Note: page_rerender/assembly is NOT gated — re-assembly of existing page_components is how owned pages deploy.
Categories: fix, guard-rail',
  '["fix","guard-rail"]'::jsonb,
  'migration', '164_pages_rebuild_policy'
);

COMMIT;
