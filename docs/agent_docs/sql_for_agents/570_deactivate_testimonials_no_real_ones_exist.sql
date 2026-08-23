-- 570 — deactivate the testimonial components; there are no real testimonials.
--
-- OWNER RULING 2026-08-23: "deactivate testimonials, we don't have any yet."
--
-- Closes two of the six live entries in bugs_open/362 by RETIRING the capability rather
-- than repairing its data source, which is the correct answer when the data does not
-- exist and inventing it would be worse than the defect.
--
-- ── WHY BOTH COMPONENTS, when the ruling named one ────────────────────────────
-- `social_proof` is the SAME capability under another name, measured not assumed: its
-- template renders `<div class="testimonials-grid">{{range .testimonials}}` into
-- `<blockquote>`, from the identical phantom source `site_specs.social_proof.testimonials`.
-- Retiring `testimonials` and leaving `social_proof` live would retire the name and keep
-- the behaviour — and `client-case-studies` carries BOTH, so that page would have kept
-- serving testimonial blockquotes from the surviving twin.
--
-- ── WHAT WAS ACTUALLY BEING SERVED, and why the ruling is right ───────────────
-- `[MEASURED 2026-08-23]` at the SERVED pages, not the stored HTML:
--   gaswholesalers.com/why-gas-wholesalers.html  — 3 <blockquote>
--   gaswholesalers.com/client-case-studies.html  — 6 <blockquote>
-- None is a customer testimonial. Every one is a first-person COMPANY statement
-- ("We built Gas Wholesalers on a simple principle…", "Pricing transparency matters…")
-- with `author` and `company` both EMPTY STRINGS, presented inside <blockquote> in a
-- section headed by a testimonials grid. That is the shape a reader parses as a customer
-- quote. The declared source has never resolved on any site, so nothing here was ever
-- fed by real data.
--
-- ⚠ `client-case-studies` serves 6 blockquotes while its `content_data->testimonials` is
-- EMPTY — the copy lives in the stored `rendered_html`, not in content_data. So
-- deactivating the component alone does NOT remove anything from the live site; it only
-- stops future selection. Removing what is served needs the tombstone below AND a
-- rerender. Anyone who checks this by reading content_data will conclude, wrongly, that
-- these pages show nothing.
--
-- ── WHAT THIS DOES, IN TWO PARTS ─────────────────────────────────────────────
-- (1) is_active=false on both components — they can no longer be selected or placed.
-- (2) build_status='removed' on the 3 live instances — the documented assembly-excluded
--     tombstone (`rerender_single_page_action.go:870` and the section editor's
--     `pageComponentNotRemovedSQL` both exclude it; 35 rows already carry it). The rows
--     are kept, not deleted, so the history and the copy remain recoverable.
--
-- Both pages stay coherent afterwards, checked before writing:
--   client-case-studies  keeps hero + case-studies-list + call-to-action (3 sections)
--   why-gas-wholesalers  keeps hero + differentiators + features + text block + CTA (5)
--
-- ⚠ A RERENDER IS REQUIRED AND IS NOT PART OF THIS FILE. Until the two pages are
-- rerendered they keep serving their deployed HTML, blockquotes included. Applying this
-- and stopping is a HALF STATE that looks done in the database and unchanged to a
-- visitor.
--
-- APPLY BY HAND (the runner takes EVERY pending file):
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
--     < docs/agent_docs/sql_for_agents/570_deactivate_testimonials_no_real_ones_exist.sql
--   then: scripts/migration/run-migrations.sh --record-only 570_deactivate_testimonials_no_real_ones_exist.sql

BEGIN;

CREATE TABLE IF NOT EXISTS bak_570_testimonials_20260823 AS
SELECT cc.id AS component_id, cc.name, cc.is_active,
       pc.id AS page_component_id, pc.build_status, pc.rendered_html, pc.content_data,
       now() AS backed_up_at
FROM content_components cc
LEFT JOIN page_components pc ON pc.component_id = cc.id
WHERE cc.name IN ('testimonials','social_proof') AND cc.is_active;

-- ── PRE-STATE GUARD ───────────────────────────────────────────────────────────
DO $guard$
DECLARE n_comp int; n_inst int;
BEGIN
  SELECT count(*) INTO n_comp FROM content_components
  WHERE is_active AND name IN ('testimonials','social_proof');
  IF n_comp <> 2 THEN
    RAISE EXCEPTION '570 ABORT: expected 2 active testimonial components, found %.', n_comp;
  END IF;

  SELECT count(*) INTO n_inst
  FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
  JOIN pages p ON p.id = pc.page_id
  WHERE cc.name IN ('testimonials','social_proof') AND cc.is_active
    AND p.status IN ('active','deployed')
    AND COALESCE(pc.build_status,'pending') <> 'removed';
  IF n_inst <> 3 THEN
    RAISE EXCEPTION '570 ABORT: expected 3 live un-tombstoned instances, found %. The placement has changed — re-measure before retiring.', n_inst;
  END IF;
END
$guard$;

-- ── (2) TOMBSTONE THE INSTANCES — before deactivating, so the join still resolves ──
UPDATE page_components pc
SET build_status = 'removed'
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND cc.name IN ('testimonials','social_proof') AND cc.is_active
  AND p.status IN ('active','deployed')
  AND COALESCE(pc.build_status,'pending') <> 'removed';

-- ── (1) RETIRE THE COMPONENTS ─────────────────────────────────────────────────
UPDATE content_components
SET is_active = false
WHERE is_active AND name IN ('testimonials','social_proof');

-- ── VERIFY (DO/RAISE — a block of SELECTs cannot stop the COMMIT) ─────────────
DO $verify$
DECLARE n_active int; n_live int; n_tomb int;
BEGIN
  SELECT count(*) INTO n_active FROM content_components
  WHERE is_active AND name IN ('testimonials','social_proof');
  IF n_active <> 0 THEN
    RAISE EXCEPTION '570 VERIFY FAILED: % testimonial component(s) still active.', n_active;
  END IF;

  SELECT count(*) INTO n_live
  FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
  JOIN pages p ON p.id = pc.page_id
  WHERE cc.name IN ('testimonials','social_proof') AND p.status IN ('active','deployed')
    AND COALESCE(pc.build_status,'pending') <> 'removed';
  IF n_live <> 0 THEN
    RAISE EXCEPTION '570 VERIFY FAILED: % live instance(s) not tombstoned — they would still assemble.', n_live;
  END IF;

  -- The rows must still EXIST. A delete would take the copy and the history with it,
  -- and this is a retirement, not a purge.
  SELECT count(*) INTO n_tomb
  FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
  WHERE cc.name IN ('testimonials','social_proof') AND pc.build_status = 'removed';
  IF n_tomb < 3 THEN
    RAISE EXCEPTION '570 VERIFY FAILED: only % tombstoned row(s) survive, expected at least 3.', n_tomb;
  END IF;

  RAISE NOTICE '570 OK: 2 components retired, % instance(s) tombstoned and retained. RERENDER THE TWO PAGES — until then they still serve their blockquotes.', n_tomb;
END
$verify$;

COMMIT;
