-- SQL_2026-08-18: give build_duration a NUMERIC value so the stat gate can
-- support the writer's "1 day" turnaround stat.
--
-- Written by the site_delivery_and_editor session (2026-08-18, at the owner's
-- direction to get the new terms live on preview.webdesign.uk), contributing
-- into this lane per its handoff — the blocker hunt this lane recorded is
-- RESOLVED: the failing step's issues ARE now persisted
-- (agent_error_log error_code CONTENT_VALIDATION_BLOCKER_DETAIL, live on
-- v1.0.1308), and the 09:57Z run's row names today's sole failure:
--   unregistered_stat, value "1 day", location brief-explanation.stat_2_value
--   ("a figure published in a stat field matches no evidence_base fact value")
-- The 08-17 "1 blockers" and the 01:00Z "{{end}}" template blockers are GONE
-- from today's run — the writer's copy is clean except this one figure.
--
-- WHY value:1 IS ATTESTABLE, not new: the owner attested "usually next day"
-- (2026-08-14). "1 day" is that same fact in figure form, and the rendered
-- stat KEEPS the hedge (label "Usual turnaround"). context_terms scope the
-- number to turnaround windows (StatClaim.Window = label + value + detail;
-- numberSupported requires a term match when context_terms are present,
-- claims.go:979-991), so this fact can never license a bare "1" in any other
-- stat (e.g. "1 million customers" would still be refused).
--
-- The claim text, writer_line and attestation are UNTOUCHED — the bot's
-- answers (which re-read the register in ~5 min) do not change.

BEGIN;

DO $fix$
DECLARE
  spec_id uuid;
  idx int;
  fact jsonb;
  n_before int;
  n_after int;
BEGIN
  SELECT id, jsonb_array_length(data->'facts') INTO spec_id, n_before
    FROM site_specs
   WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
     AND aspect='evidence_base' AND is_current;
  IF spec_id IS NULL THEN
    RAISE EXCEPTION 'no current evidence_base row for webdesign.uk';
  END IF;

  SELECT (o.ord - 1)::int, o.f INTO idx, fact
    FROM site_specs s,
         jsonb_array_elements(s.data->'facts') WITH ORDINALITY o(f, ord)
   WHERE s.id = spec_id AND o.f->>'id' = 'build_duration';
  IF idx IS NULL THEN
    RAISE EXCEPTION 'build_duration fact not found';
  END IF;
  IF fact ? 'value' THEN
    RAISE EXCEPTION 'build_duration already carries value % — previously applied or concurrently edited; read the row before re-running', fact->'value';
  END IF;
  IF fact->>'writer_line' IS DISTINCT FROM 'usually ready the next day' THEN
    RAISE EXCEPTION 'build_duration writer_line is %, not the attested 2026-08-14 text — pre-state mismatch, aborting', fact->>'writer_line';
  END IF;

  UPDATE site_specs
     SET data = jsonb_set(jsonb_set(data,
           ARRAY['facts', idx::text, 'value'], '1'::jsonb),
           ARRAY['facts', idx::text, 'context_terms'],
           '["turnaround", "day", "ready"]'::jsonb),
         updated_at = now()
   WHERE id = spec_id;

  -- Read back: value set, terms set, claim text untouched, count unchanged.
  SELECT jsonb_array_length(data->'facts'),
         (SELECT o.f FROM jsonb_array_elements(data->'facts') WITH ORDINALITY o(f, ord)
           WHERE o.f->>'id'='build_duration')
    INTO n_after, fact
    FROM site_specs WHERE id = spec_id;
  IF n_after IS DISTINCT FROM n_before THEN
    RAISE EXCEPTION 'fact count changed % -> %', n_before, n_after;
  END IF;
  IF (fact->>'value')::numeric IS DISTINCT FROM 1
     OR NOT (fact->'context_terms' @> '["turnaround"]'::jsonb)
     OR fact->>'writer_line' IS DISTINCT FROM 'usually ready the next day' THEN
    RAISE EXCEPTION 'post-state wrong: %', jsonb_pretty(fact);
  END IF;
  RAISE NOTICE 'build_duration now carries value=1, context_terms=[turnaround,day,ready]; % facts unchanged in count', n_after;
END $fix$;

COMMIT;
