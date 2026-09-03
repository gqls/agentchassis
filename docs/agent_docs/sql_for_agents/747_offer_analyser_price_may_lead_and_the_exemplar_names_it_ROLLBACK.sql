-- 747_..._ROLLBACK.sql — restores both sentences verbatim.
-- ⚠ Restoring the absolute makes price unwritable as a lead_with point again, so any
-- money_flow question will go back to unanswered on the NEXT analysis of each site. It does
-- not rewrite specs already produced under 747; those keep whatever they were given.
BEGIN;
DO $mig$
DECLARE
    new_a text := 'a benefit to the reader. A point that describes us or our inventory — what we stock, how many, what it costs — is allowed only where the reader genuinely needs it: rank it LAST BY DEFAULT — unless it is genuinely this site''s strongest reader benefit, which is rare — keep it to a single clause, and prefer the reader-benefit form wherever one exists ("find the one you need in 30 seconds", not "we have 500 templates"). Say less or leave it out.';
    old_a text := 'a benefit to the reader, never a description of us or of our inventory.';
    new_b text := 'For most sites that is some form of what will this actually get me, how much work is it to get it, and what does it cost me;';
    old_b text := 'For most sites that is some form of what will this actually get me and how much work is it to get it;';
    p text; n int; agent_id uuid; step_key text;
BEGIN
    SELECT ad.id, st.key, st.value->'config'->>'prompt' INTO agent_id, step_key, p
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') st
     WHERE ad.type='offer-analyser' AND ad.is_active
       AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
       AND st.value->'config'->>'prompt' LIKE '%' || new_a || '%';
    IF p IS NULL THEN RAISE EXCEPTION 'ABORT: 747 is not applied (or the prompt has moved)'; END IF;

    n := (length(p) - length(replace(p, new_a, ''))) / length(new_a);
    IF n <> 1 THEN RAISE EXCEPTION 'ABORT: replacement A occurs % times, expected 1', n; END IF;
    n := (length(p) - length(replace(p, new_b, ''))) / length(new_b);
    IF n <> 1 THEN RAISE EXCEPTION 'ABORT: replacement B occurs % times, expected 1', n; END IF;

    p := replace(p, new_a, old_a);
    p := replace(p, new_b, old_b);
    IF position(new_a in p) > 0 OR position(new_b in p) > 0 THEN
        RAISE EXCEPTION 'ABORT: a replacement survived the restore';
    END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
             ARRAY['workflow','steps',step_key,'config','prompt'], to_jsonb(p), false),
           updated_at = now()
     WHERE id = agent_id;
    RAISE NOTICE '747 ROLLBACK: both sentences restored verbatim';
END $mig$;
COMMIT;
