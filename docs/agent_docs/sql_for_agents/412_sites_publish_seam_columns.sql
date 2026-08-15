-- 412: sites gains the publish-seam opt-in columns
--
-- site_delivery_and_editor Phase 2 (owner-approved PLAN 2026-08-14): a
-- provider-agnostic publish seam mirrors the built artefact tree in
-- b2://portfolio-sites/<domain> to a hosted copy (Cloudflare Pages Direct
-- Upload primary; B2->ugg2 worker as the in-estate backend). Per the
-- 2026-08-02 opt-in ruling, the seam ships default-OFF: publish_target is
-- NULL on every row and the publish action is a recorded no-op until a row
-- names a backend. Deliberately NOT overloading sites.github_repo (single
-- string, one consumer — the git deploy path; a second meaning would be the
-- two-writers drift this estate documents).
--
-- Blast radius measured before writing (2026-08-15): zero bare positional
-- INSERT INTO sites across go/sql/sh; zero SELECT * FROM sites in Go; zero
-- existing mentions of publish_target/published_hash anywhere in the tree.
--
-- Rollback: 412_sites_publish_seam_columns_ROLLBACK.sql (sidecar, hand-run).

BEGIN;

ALTER TABLE sites
  ADD COLUMN IF NOT EXISTS publish_target  text,
  ADD COLUMN IF NOT EXISTS publish_project text,
  ADD COLUMN IF NOT EXISTS published_hash  text,
  ADD COLUMN IF NOT EXISTS published_at    timestamptz;

COMMENT ON COLUMN sites.publish_target IS
  'Publish-seam backend key (platform/publish). NULL = seam OFF for this site (the default, 2026-08-02 opt-in ruling). Known values registered in platform/publish/publisher.go.';
COMMENT ON COLUMN sites.publish_project IS
  'Provider-side project identifier (e.g. the CF Pages project name). Meaning is backend-specific; NULL until first publish provisions it.';
COMMENT ON COLUMN sites.published_hash IS
  'Tree hash of the built artefact set at the last SUCCESSFUL publish. The reconciler publishes on drift between this and the current B2 tree hash; it is written only after served-hash acceptance, never on API 200.';
COMMENT ON COLUMN sites.published_at IS
  'Timestamp of the last successful publish (same write as published_hash).';

DO $$
DECLARE
  n int;
BEGIN
  SELECT count(*) INTO n FROM information_schema.columns
   WHERE table_schema = 'public' AND table_name = 'sites'
     AND column_name IN ('publish_target', 'publish_project', 'published_hash', 'published_at');
  IF n <> 4 THEN
    RAISE EXCEPTION '412: expected 4 publish-seam columns on sites, found %', n;
  END IF;
  -- Opt-in default-OFF is the load-bearing property: a non-NULL default here
  -- would turn the seam ON fleet-wide. Assert it.
  SELECT count(*) INTO n FROM information_schema.columns
   WHERE table_schema = 'public' AND table_name = 'sites'
     AND column_name = 'publish_target' AND column_default IS NOT NULL;
  IF n <> 0 THEN
    RAISE EXCEPTION '412: publish_target must have no default (opt-in OFF)';
  END IF;
END $$;

COMMIT;
