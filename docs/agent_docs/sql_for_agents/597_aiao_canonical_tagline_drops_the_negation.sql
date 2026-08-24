-- 597 — D2, OWNER RULING 2026-08-24: "correct that instruction narrowly".
--
-- WHAT THIS CHANGES, in one line: the site's mandated tagline stops being written
-- as a define-by-negation. "in days, not months" -> "in days". Nothing else.
--
--   before: Multi-agent systems deployed to production in days, not months
--           — on Kubernetes, Kafka, and Postgres
--   after:  Multi-agent systems deployed to production in days
--           — on Kubernetes, Kafka, and Postgres
--
-- The stack-specificity is the genuinely strong half and is kept verbatim. Only
-- the negation clause goes. This is `bugs_closed/305`'s D2, the one instance the
-- copy gate can never fix by itself.
--
-- WHY THE GATE COULD NOT FIX IT, so nobody re-files this as a gate defect: the
-- gate EXEMPTS a phrase the brief supplied, by design (four council rounds). The
-- writer was ORDERED to use this sentence, so repairing it would mean the system
-- silently disobeying its own brief. It can only move by a human editing the
-- instruction — which is exactly what D2 was.
--
-- ⚠ FIVE COPIES, NOT ONE, AND THAT IS THE WHOLE DIFFICULTY. `[MEASURED 2026-08-24]`
-- The sentence is duplicated across three aspects:
--     content_direction :: emphasis                (string)
--     content_direction :: formatted               (string)
--     identity          :: core_value_proposition  (string)
--     site_plan         :: strategy_notes          (string)
--     site_plan         :: content_direction       (OBJECT, 3598 chars)
-- Correcting only `identity` would achieve NOTHING VISIBLE: the exemption is
-- computed over the FLATTENED brief corpus (all aspects, 679 strings on this
-- site), so while ANY copy still carries the old phrase the gate keeps exempting
-- it and no page changes. Half a fix here is indistinguishable from no fix.
--
-- ⚠ AND `site_plan.content_direction` IS AN OBJECT. The obvious patch —
-- `jsonb_set(data, '{k}', to_jsonb(replace(data->>'k', ...)))` — FLATTENS it to a
-- JSON string and destroys the spec. This migration therefore replaces at the
-- DOCUMENT level (`data::text` -> replace -> `::jsonb`), which is structure-
-- preserving for both shapes. That is safe here and only here because neither the
-- needle nor the replacement contains a quote, backslash or brace, so the
-- serialised JSON stays valid and only string CONTENTS change.
--
-- ⚠ THE TAGLINE IS NOT IN THE COLUMNS CALLED `tagline`. `sites.tagline` and
-- `site_specs.identity->>'tagline'` both hold "Production-Grade Multi-Agent
-- Systems. Built Right." — a DIFFERENT sentence, untouched by this migration.
-- Anyone told to "fix the tagline" edits those and changes the wrong line.
--
-- WHAT HAPPENS NEXT, and it is the elegant half: once this applies, the old
-- sentence is no longer brief-supplied, so the gate STOPS exempting it. Existing
-- page copy carrying "in days, not months" becomes a normal repairable hit on the
-- next render — 7 components across 5 pages as of today. No page editing is
-- needed and none should be attempted by hand.
--
-- Versioned, not overwritten: supersede + re-insert, the `585` pattern, so the
-- brief's history survives. Config-only; live on apply, no image, no roll.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
    FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain = 'ai-agent-orchestration.com' AND ss.is_current
     AND ss.data::text LIKE '%in days, not months%';
  IF n <> 3 THEN
    RAISE EXCEPTION '597 ABORT: expected 3 current spec ROWS carrying the phrase (identity, content_direction, site_plan), found % — the brief has changed since this was written; re-census before applying', n;
  END IF;
END $$;

WITH superseded AS (
  UPDATE site_specs ss
     SET is_current = false, superseded_at = now()
    FROM sites s
   WHERE s.id = ss.site_id
     AND s.domain = 'ai-agent-orchestration.com'
     AND ss.is_current
     AND ss.data::text LIKE '%in days, not months%'
  RETURNING ss.site_id, ss.aspect, ss.data, ss.source, ss.source_agent,
            ss.source_item_id, ss.pinned
)
-- ⚠ `created_by` is NOT NULL. Omitting it aborts the whole migration on the
-- INSERT (caught here on first apply, 2026-08-24). It names the WRITER of this
-- row, so it is the migration rather than the brief's original author — the
-- original attribution survives on the superseded row, which is the point of
-- versioning rather than overwriting.
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, source_item_id,
                        pinned, notes, is_current, created_at, created_by)
SELECT site_id, aspect,
       replace(data::text, 'in days, not months', 'in days')::jsonb,
       source, source_agent, source_item_id, pinned,
       'D2, owner ruling 2026-08-24 (bugs_closed/305): narrow correction — the canonical tagline stops being written as a define-by-negation. Only "in days, not months" -> "in days"; every other byte carried forward. Migration 597.',
       true, now(), 'migration-597-d2-tagline'
  FROM superseded;

DO $$
DECLARE old_left int; new_rows int; obj_ok int;
BEGIN
  SELECT count(*) INTO old_left
    FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain = 'ai-agent-orchestration.com' AND ss.is_current
     AND ss.data::text LIKE '%in days, not months%';
  IF old_left <> 0 THEN
    RAISE EXCEPTION '597 FAILED: % current spec row(s) still carry the old phrase — a partial correction leaves the gate exempting it and changes nothing', old_left;
  END IF;

  SELECT count(*) INTO new_rows
    FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain = 'ai-agent-orchestration.com' AND ss.is_current
     AND ss.data::text LIKE '%production in days —%';
  IF new_rows <> 3 THEN
    RAISE EXCEPTION '597 FAILED: expected 3 current rows carrying the corrected sentence, got % — the replacement did not land where the precheck said it would', new_rows;
  END IF;

  -- The object-valued key must still be an OBJECT. This is the assertion that
  -- catches the to_jsonb flattening trap if anyone ever "simplifies" the UPDATE.
  SELECT count(*) INTO obj_ok
    FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain = 'ai-agent-orchestration.com' AND ss.is_current
     AND ss.aspect = 'site_plan'
     AND jsonb_typeof(ss.data->'content_direction') = 'object';
  IF obj_ok <> 1 THEN
    RAISE EXCEPTION '597 FAILED: site_plan.content_direction is no longer an object — the replacement flattened a nested spec';
  END IF;

  RAISE NOTICE '597 OK: 3 spec rows re-versioned, old phrase gone, nested spec intact. The gate will now treat the old sentence as repairable on the next render.';
END $$;

COMMIT;
