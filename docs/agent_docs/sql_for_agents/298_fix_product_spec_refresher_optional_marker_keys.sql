-- 298 — product-spec-refresher: two config keys carry a documentation marker,
-- so neither has ever resolved
--
-- WHAT: renames the `refresh_specs` step-config keys "category?" -> "category"
-- and "limit?" -> "limit" on the live product-spec-refresher definition. Values
-- are untouched ("input_data.category", "input_data.limit"); only the key names
-- change.
--
-- WHY (bugs_open/134): `category?` is documentation notation — the ordinary
-- convention for "this field is optional" — and seed 156 uses it correctly in a
-- comment on its line 15 ("{site_id, category?}"). Forty-five lines later the
-- same notation appears inside the real JSON, where it is just a key name. The
-- action reads "category" and "limit" (refresh_product_specs_action.go:177,
-- 211, 215), no suffix, and no Go code anywhere reads a `?`-suffixed key. An
-- unrecognised step-config key is silently ignored at execution, so the defect
-- has no symptom: `limit` quietly takes its hard-coded default of 20 and
-- `category` arrives empty, whatever the caller passes in input_data.
--
-- NO BEHAVIOUR-REGRESSION SURFACE. The agent has never run — 0 rows in
-- orchestration_states for 'product-spec-refresher', and no scheduled_tasks row
-- matches it (measured 2026-07-28, bugs_open/134 §Severity). Correcting the keys
-- IS a behaviour change (limit would start honouring the caller instead of
-- defaulting), but nothing depends on the current behaviour and the seed's own
-- comment says what was intended.
--
-- FIXED AT SOURCE IN THE SAME COMMIT. Seed 156 lines 60-61 are corrected
-- alongside this file — a replay of 156 against a fresh database would otherwise
-- restore the defect, and the seed is what a rebuild reads. Line 15's comment is
-- deliberately left as it is: there `category?` is prose, and correct.
--
-- ALSO IN THE SAME COMMIT, the door-closing half: `CheckConfig: true` on
-- RefreshProductSpecsInputSpec opts the action into unknown-config-key
-- detection, and `config-key-audit --suspicious-keys` makes doc-notation
-- punctuation in a live key name a reported finding fleet-wide. Renaming two
-- keys fixes today's instance; those two make the next one visible.
--
-- Idempotent: the UPDATE is fenced on the bad key still being present, so a
-- replay matches no row. DB-only; snapshot-prefixed.

BEGIN;

SELECT snapshot_agent('product-spec-refresher',
    '298_fix_product_spec_refresher_optional_marker_keys.sql: pre-update');

-- One write, not two. The object is rebuilt by subtracting both marker keys and
-- merging the corrected pair back in, so a partially-applied state (one key
-- renamed, one not) is not representable.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,refresh_specs,config}',
      ((default_config #> '{workflow,steps,refresh_specs,config}')
         - 'category?' - 'limit?')
        || jsonb_build_object(
             'category', 'input_data.category',
             'limit',    'input_data.limit'),
      false)
WHERE type = 'product-spec-refresher'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config #> '{workflow,steps,refresh_specs,config}' ? 'category?';

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline', 'build',
$note$## 298: a doc convention for "optional" leaked into two real config key names
Observed: product-spec-refresher's refresh_specs step carried {"site_id": "site_record.site_id", "category?": "input_data.category", "limit?": "input_data.limit"}. The refresh_product_specs action reads "category" and "limit", so neither key ever resolved — limit silently took its hard-coded default of 20 and category arrived empty, whatever the caller passed. No symptom, because an unrecognised step-config key is ignored rather than reported, and the agent has 0 runs ever.
Root cause: seed 156 describes the agent's input contract in a comment as "{site_id, category?}", where the trailing ? is documentation notation for "optional". The same notation was then written into the actual JSON forty-five lines below, where it is part of the key name.
Fix: the two live keys renamed to "category" and "limit" (values unchanged) by this migration; seed 156 corrected at source in the same commit so a replay cannot restore the defect; CheckConfig: true added to RefreshProductSpecsInputSpec so an unrecognised key on this action is warned at runtime and reported by scripts/audit-config-keys.sh; and a new fleet-wide offline check, config-key-audit --suspicious-keys, flags doc-notation punctuation (? * : and space) in any live step-config key name.
Verified: the DO blocks below assert the corrected pair is present with the right values, that neither marker key survives, that the neighbouring site_id mapping is intact, and that the sibling steps ensure_site_record and complete still exist — a verify block that only reads its own key cannot tell a surgical rename from a write that flattened the workflow.
Categories: migration$note$,
'["migration"]'::jsonb, 'human', 'run-migrations.sh');

DO $$
DECLARE cfg jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,refresh_specs,config}' INTO cfg
    FROM agent_definitions
    WHERE type = 'product-spec-refresher'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF cfg IS NULL THEN
        RAISE EXCEPTION '298: no live product-spec-refresher config at {workflow,steps,refresh_specs,config}';
    END IF;
    IF cfg ->> 'category' IS DISTINCT FROM 'input_data.category' THEN
        RAISE EXCEPTION '298: category does not map to input_data.category (%)', cfg;
    END IF;
    IF cfg ->> 'limit' IS DISTINCT FROM 'input_data.limit' THEN
        RAISE EXCEPTION '298: limit does not map to input_data.limit (%)', cfg;
    END IF;
    -- The rename is only a fix if the OLD keys are gone. Leaving them would keep
    -- the inert pair sitting beside the live one, which is the state
    -- bugs_open/134 describes and reads, by inspection, exactly like a fix.
    IF cfg ? 'category?' OR cfg ? 'limit?' THEN
        RAISE EXCEPTION '298: a marker key survived the rename (%)', cfg;
    END IF;
END $$;

-- A verify block that only checks its own keys cannot tell a surgical rename
-- from a write that replaced the object, the step or the workflow. Assert
-- neighbours at both levels: the sibling key in the same config, and the
-- sibling steps one level up.
DO $$
DECLARE cfg jsonb; steps jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,refresh_specs,config}',
           default_config #> '{workflow,steps}'
      INTO cfg, steps
    FROM agent_definitions
    WHERE type = 'product-spec-refresher'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF cfg ->> 'site_id' IS DISTINCT FROM 'site_record.site_id' THEN
        RAISE EXCEPTION '298: the neighbouring site_id mapping did not survive (%)', cfg;
    END IF;
    IF steps -> 'ensure_site_record' IS NULL OR steps -> 'complete' IS NULL THEN
        RAISE EXCEPTION '298: a sibling step vanished — the write flattened the workflow (%)', steps;
    END IF;
END $$;

COMMIT;

-- Verify:
--   SELECT default_config #> '{workflow,steps,refresh_specs,config}'
--   FROM agent_definitions WHERE type='product-spec-refresher'
--     AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--   -- expect exactly: {"limit": "input_data.limit", "site_id": "site_record.site_id",
--   --                  "category": "input_data.category"}
-- Fleet-wide re-check that no other definition carries the same class of key
-- (bugs_open/134 found these two and nothing else):
--   SELECT ad.type AS agent, e.k AS step, ck.key AS suspicious_key
--   FROM agent_definitions ad,
--        jsonb_each(ad.default_config->'workflow'->'steps') AS e(k,v),
--        jsonb_object_keys(v->'config') AS ck(key)
--   WHERE ad.deleted_at IS NULL AND COALESCE(ad.is_snapshot,false)=false AND ad.is_active
--     AND jsonb_typeof(v->'config')='object'
--     AND (ck.key LIKE '%?' OR ck.key LIKE '%*' OR ck.key LIKE '% %' OR ck.key LIKE '%:%')
--   ORDER BY 1,2;
--   -- the Go equivalent, which also descends into loop sub-workflows:
--   --   scripts/audit-config-keys.sh   (SUSPICIOUS KEYS section)
-- Rollback: restore the snapshot taken above. Reverting by hand would mean
-- writing the marker keys back, which the audit will then report — restoring the
-- snapshot is the only clean undo.
--
-- WHERE THE SNAPSHOT LIVES (council round 1, editquality's gating check —
-- there are TWO snapshot_agent overloads writing to two different tables):
-- pg_get_functiondef shows snapshot_agent(text) inserts an is_snapshot=true row
-- into agent_definitions, while the TWO-ARG form used above inserts into
-- agent_definitions_backup. Verified at the artefact after the 2026-08-03
-- apply: the backup row exists (snapshot_taken_at 11:06:17Z, reason naming this
-- file) and its stored config still carries 'category?' — i.e. the restore
-- target genuinely holds the pre-update state. agent_definitions_backup keeps
-- the SOURCE row's id and created_at, so order by snapshot_taken_at, never
-- created_at, when picking the row to restore.
