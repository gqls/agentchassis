-- 437 — content_components REFUSES the legacy JSON-Schema input_schema dialect
--       (bugs_open/265; lane bugfix_265_legacy_dialect_unrepresentable)
--
-- WHY THIS FILE EXISTS
-- The house dialect for a component's content fields is `{"fields": {...}}`. An older
-- dialect — JSON-Schema, `{"type":"object","properties":{...},"required":[...]}` — was
-- declared EXTINCT in a code comment on 2026-07-21 (0 of 173). It has been reintroduced
-- four times since: seeds 207 (2026-07-25, four days after the census), 247 and 250
-- (2026-07-28), and a hand seed on 2026-08-10 (loans-consolidation, since converted by
-- its own lane). Every one was hand-authored SQL: `created_from='manual'`,
-- `source_agent_type` NULL. The component-creator has produced 0 legacy rows in 69.
-- So no Go-side gate would have caught any of them — components are routinely written
-- by migrations and by hand with no commit at all (RFC_009's own words), and the only
-- seam that sees every producer is this table. Hence a CHECK constraint.
--
-- WHAT IT DOES, IN ORDER
--   1. Guards: not already applied; the legacy population is EXACTLY the three rows
--      this file was written against (drift → refuse, re-census, rewrite).
--   2. Backs the three rows up (id, function, input_schema, updated_at).
--   3. Converts them to v2, mirroring datahelpers.SchemaContentFields' projection
--      EXACTLY so behaviour is preserved: the six copied keys (source, on_missing,
--      fallback, missing_reason, items, min_items); minItems→min_items; type
--      "string"→"text"; llm_guidance, else description → llm_guidance; source defaults
--      to "llm" when absent; top-level required[] folded into per-field required:true.
--      Nested JSON-Schema inside `items` is copied verbatim, as the helper does.
--   4. Verifies (DO/RAISE — a SELECT cannot stop a COMMIT): 0 legacy rows remain; every
--      converted row's field-name set equals its old properties set; every field has a
--      source; required flags equal the old required[].
--   5. Adds `chk_input_schema_no_legacy_dialect`: a top-level `properties` key is
--      refused. NULL and non-object schemas pass. Nested `properties` under
--      fields.<x>.items is the shape of an ITEM and is legitimate — mechanism-flow and
--      evidence-timeseries both carry one — so only the TOP level is constrained.
--
-- WHAT IT DELIBERATELY DOES NOT DO
--   - It does not correct report-dossier's `source`. Its seed says the body is
--     "never authored by an LLM", yet the projection (and so this conversion) marks it
--     source:"llm" — that is the over-report bug 265's addendum 1 predicted, and it is
--     the owning lane's content decision, surfaced in the bug file, not made here.
--   - It does not touch component_versions (history) or any bak_* table. A restore
--     from one that carries the old dialect now FAILS on the constraint, loudly. That
--     is the intended behaviour: the alternative was silent projection.
--
-- ROLLBACK: 437_..._ROLLBACK.sql drops the constraint and restores the three rows
-- from the backup table by id.
--
-- The Go half (birth-path refusal in store_generated_component, the corrected doc
-- comment, the tripwire message) is inert until a chassis roll; THIS FILE IS LIVE
-- THE MOMENT IT IS APPLIED and needs nothing from the roll. Image-first ordering does
-- not apply: nothing here names a Go affordance.

BEGIN;

-- ── 1. Guards ───────────────────────────────────────────────────────────────
DO $$
DECLARE
  n_legacy   int;
  n_expected int;
  n_nonobj   int;
BEGIN
  IF EXISTS (SELECT 1 FROM pg_constraint
              WHERE conrelid = 'public.content_components'::regclass
                AND conname  = 'chk_input_schema_no_legacy_dialect') THEN
    RAISE EXCEPTION 'chk_input_schema_no_legacy_dialect already exists — migration 437 is already applied; record it with --record-only rather than re-running';
  END IF;

  SELECT count(*) INTO n_legacy FROM public.content_components WHERE input_schema ? 'properties';
  SELECT count(*) INTO n_expected FROM public.content_components
   WHERE input_schema ? 'properties'
     AND id IN ('fb870e82-2f01-46e4-9552-e764515e18d8',   -- evidence-timeseries
                'fa5e4524-e9e9-41f1-826a-8474c03c48e2',   -- mechanism-flow
                '5fe24e56-2526-469b-9554-1de39aa038b2');  -- report-dossier
  IF n_legacy <> 3 OR n_expected <> 3 THEN
    RAISE EXCEPTION 'legacy-dialect population is % rows (% of the 3 expected ids) — not the shape 437 was written against (3 rows: evidence-timeseries, mechanism-flow, report-dossier, measured 2026-08-16). Re-run the census, extend the conversion to the new rows, then apply', n_legacy, n_expected;
  END IF;

  -- Every property definition must be an object, or the projection below would error
  -- mid-statement rather than at this guard.
  SELECT count(*) INTO n_nonobj
    FROM public.content_components c, jsonb_each(c.input_schema->'properties') p
   WHERE c.input_schema ? 'properties' AND jsonb_typeof(p.value) <> 'object';
  IF n_nonobj <> 0 THEN
    RAISE EXCEPTION '% legacy property definition(s) are not JSON objects — extend the conversion before applying', n_nonobj;
  END IF;
END $$;

-- ── 2. Backup ───────────────────────────────────────────────────────────────
CREATE TABLE public.content_components_bak_20260816_265_legacy_dialect AS
SELECT id, function, input_schema, updated_at
  FROM public.content_components
 WHERE input_schema ? 'properties';

-- ── 3. Convert: the SchemaContentFields projection, in SQL ──────────────────
UPDATE public.content_components c
   SET input_schema = jsonb_build_object('fields', (
         SELECT jsonb_object_agg(p.key,
                  -- the six keys the helper copies when present
                  (SELECT coalesce(jsonb_object_agg(k.key, k.value), '{}'::jsonb)
                     FROM jsonb_each(p.value) k
                    WHERE k.key IN ('source','on_missing','fallback','missing_reason','items','min_items'))
                  -- minItems → min_items, only when min_items itself is absent
               || CASE WHEN (p.value ? 'min_items') OR NOT (p.value ? 'minItems') THEN '{}'::jsonb
                       ELSE jsonb_build_object('min_items', p.value->'minItems') END
                  -- type: string-typed only; "string" becomes "text"
               || CASE WHEN jsonb_typeof(p.value->'type') = 'string'
                       THEN jsonb_build_object('type', CASE WHEN p.value->>'type' = 'string' THEN 'text' ELSE p.value->>'type' END)
                       ELSE '{}'::jsonb END
                  -- llm_guidance from llm_guidance, else from description (non-empty strings only)
               || CASE WHEN jsonb_typeof(p.value->'llm_guidance') = 'string' AND p.value->>'llm_guidance' <> ''
                       THEN jsonb_build_object('llm_guidance', p.value->>'llm_guidance')
                       WHEN jsonb_typeof(p.value->'description') = 'string' AND p.value->>'description' <> ''
                       THEN jsonb_build_object('llm_guidance', p.value->>'description')
                       ELSE '{}'::jsonb END
                  -- source defaults to "llm" when the property declares none
               || CASE WHEN p.value ? 'source' THEN '{}'::jsonb ELSE '{"source": "llm"}'::jsonb END
                  -- top-level required[] folded in
               || CASE WHEN c.input_schema->'required' ? p.key THEN '{"required": true}'::jsonb ELSE '{}'::jsonb END
                )
           FROM jsonb_each(c.input_schema->'properties') p
       )),
       updated_at = now()
 WHERE c.input_schema ? 'properties';

-- ── 4. Verify — RAISE, because a bare SELECT cannot stop the COMMIT ─────────
DO $$
DECLARE
  n_legacy    int;
  n_bad_names int;
  n_no_source int;
  n_bad_req   int;
BEGIN
  SELECT count(*) INTO n_legacy FROM public.content_components WHERE input_schema ? 'properties';
  IF n_legacy <> 0 THEN
    RAISE EXCEPTION 'conversion left % legacy row(s) — aborting', n_legacy;
  END IF;

  -- field-name set must equal the old properties set, row by row
  SELECT count(*) INTO n_bad_names
    FROM public.content_components_bak_20260816_265_legacy_dialect b
    JOIN public.content_components c ON c.id = b.id
   WHERE (SELECT coalesce(array_agg(k ORDER BY k), '{}') FROM jsonb_object_keys(b.input_schema->'properties') k)
      <> (SELECT coalesce(array_agg(k ORDER BY k), '{}') FROM jsonb_object_keys(c.input_schema->'fields') k);
  IF n_bad_names <> 0 THEN
    RAISE EXCEPTION '% converted row(s) have a field-name set that differs from their old properties set — aborting', n_bad_names;
  END IF;

  -- every field carries a source (the helper's llm default)
  SELECT count(*) INTO n_no_source
    FROM public.content_components c, jsonb_each(c.input_schema->'fields') f
   WHERE c.id IN (SELECT id FROM public.content_components_bak_20260816_265_legacy_dialect)
     AND NOT (f.value ? 'source');
  IF n_no_source <> 0 THEN
    RAISE EXCEPTION '% converted field(s) have no source — aborting', n_no_source;
  END IF;

  -- required flags equal the old required[] exactly (both directions)
  SELECT count(*) INTO n_bad_req
    FROM public.content_components_bak_20260816_265_legacy_dialect b
    JOIN public.content_components c ON c.id = b.id
    CROSS JOIN LATERAL jsonb_each(c.input_schema->'fields') f
   WHERE (coalesce((f.value->>'required')::boolean, false)) <> (b.input_schema->'required' ? f.key);
  IF n_bad_req <> 0 THEN
    RAISE EXCEPTION '% converted field(s) have a required flag that disagrees with the old required[] — aborting', n_bad_req;
  END IF;

  RAISE NOTICE '437: 3 legacy rows converted to v2; names, sources and required flags verified';
END $$;

-- ── 5. The guarantee ────────────────────────────────────────────────────────
ALTER TABLE public.content_components
  ADD CONSTRAINT chk_input_schema_no_legacy_dialect
  CHECK (input_schema IS NULL OR NOT (input_schema ? 'properties'));

COMMENT ON CONSTRAINT chk_input_schema_no_legacy_dialect ON public.content_components IS
  'bugs_open/265, migration 437 (2026-08-16): input_schema must be the v2 {"fields":{...}} dialect. A top-level "properties" key is the retired JSON-Schema dialect and is refused for every producer (seeds, scripts, Go, admin). Nested properties under fields.<x>.items are fine. Reader: datahelpers.SchemaContentFields.';

COMMIT;
