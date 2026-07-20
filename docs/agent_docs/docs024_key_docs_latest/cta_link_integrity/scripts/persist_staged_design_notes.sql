-- persist_staged_design_notes.sql — cta_link_integrity, council trail 2525f980
--
-- PROTOCOL (tooling_provenance): load before write. Run the SELECT first; the
-- 2026-07-20 run found two CORRECTION rows from the bugfix-023 session
-- (commit b6e374fc2) which reshaped the plan — these design rows APPEND
-- alongside them (append-only convention), they do not replace them.
--
--   SELECT subject_key, created_by, created_at, left(body,80)
--   FROM doc_notes
--   WHERE subject_key IN ('resolve_internal_links','plan_sections','rerender_page_sections')
--   ORDER BY created_at;

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by) VALUES
('pipeline','resolve_internal_links',
 '2026-07-20 staged rollout v6 (trail 2525f980): schema-derived CTA pairing (datahelpers/ctafields.go — 3 sibling forms: _label/_text/bare stem, plus a site_assets source guard, per the bugfix-023 corrections in the sibling rows here) lands OBSERVE-ONLY. ctaFieldNames still decides every write; the action logs the derivation delta and Warns per uncovered field. Precedence inverts only after a real build''s delta log is reviewed, via a further council round. Flip constraints accumulated on the trail: needle-gated separate apply/verify/rollback for any jsonb surgery; the Warn-only uncovered guard replaced by a consumed detection path (named handler, NOT needs_human_review); guidelines re-reviews the on_missing/required skip branch; loadSingleComponentSchema converges onto ParseInputSchemaValue.',
 '["cta-link-integrity","staged-rollout"]'::jsonb,'council-gate','leopardess3'),
('pipeline','rerender_page_sections',
 '2026-07-20 staged rollout v6 (trail 2525f980): the cta ownership-conflict observe log lives in THIS action''s merge loop — the true loss site, where stored content_data (the resolver''s last write) merges first and fresh plan.ResolvedData merges last and wins. It logs per derived CTA field when fresh would replace a differing stored value, carrying the rerender reason (cta_links_stale = deliberate recompute). An earlier sketch placed the log inside planSection, where resolvedData is a fresh local map and the condition could never fire — see the bugfix-023 correction row under plan_sections. Merge behaviour is UNCHANGED this stage.',
 '["cta-link-integrity","staged-rollout"]'::jsonb,'council-gate','leopardess3');
