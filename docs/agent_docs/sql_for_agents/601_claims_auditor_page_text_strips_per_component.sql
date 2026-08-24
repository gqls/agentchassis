-- 601_claims_auditor_page_text_strips_per_component.sql
--
-- bugs_open/380 follow-up, found by the first live cold audit (garden-tools.uk, corr
-- 86ef3e17, 2026-08-24 16:52Z): the auditor's page-text extraction silently DROPS most of a
-- page. how-we-assess: 14,762 chars of component html -> 3,732 chars of "text", and the
-- owner's own quoted sentence ("we buy the tool at the same price a reader would pay",
-- present in the faq component's rendered_html) reached the model NOWHERE — position('we
-- buy') = 0 across the whole rendered prompt. index stripped to ONE character.
--
-- MECHANISM (PostgreSQL ARE, not a tag problem — every component's <style>/</style> pairs
-- are balanced): in Postgres a regex takes the greediness of its FIRST quantified atom, so
--     '<style[^>]*>.*?</style>'
-- is GREEDY as a whole — the '[^>]*' is greedy, and the '.*?' does not get to be lazy. Run
-- over a string_agg of four components it matches from the FIRST '<style' to the LAST
-- '</style>', eating every component's body text in between. string_agg also carries no
-- ORDER BY, so WHICH text survived depended on aggregation order — the same query returned
-- the sentence in one session's hand-run and not in the auditor's run minutes apart.
-- (Go's RE2 has no such rule; the two rerender actions using this shape in Go are fine.)
--
-- FIX: strip each component ON ITS OWN before aggregating (one style block per component,
-- so greediness cannot cross a boundary), make the first quantifier lazy too as belt and
-- braces ('<style[^>]*?>'), and aggregate in a DETERMINISTIC order (pc.position, slot_name).
-- Measured per-component on garden-tools how-we-assess: 8,269 chars with the sentence
-- intact, vs 3,732 without it. Cap per page (12000, from 597) is unchanged.
--
-- Fleet census 2026-08-24: the SQL shape exists on exactly ONE live agent (claims-auditor).
-- LANDMINES entry filed for the regex rule.

BEGIN;

DO $$
DECLARE q text; n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN RAISE EXCEPTION '601 REFUSED: % active claims-auditor rows', n; END IF;

  SELECT default_config #>> '{workflow,steps,load_page_text,config,query}' INTO q
    FROM agent_definitions WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF q IS NULL THEN RAISE EXCEPTION '601: load_page_text.config.query not found'; END IF;
  IF position('string_agg(pc.rendered_html, '' '')' in q) = 0 THEN
    RAISE EXCEPTION '601: already applied or drifted — aggregate-then-strip shape not found';
  END IF;
  IF position(', 12000) AS page_text' in q) = 0 THEN
    RAISE EXCEPTION '601: 597 not applied (cap is not 12000) — apply 597 first';
  END IF;

  PERFORM snapshot_agent('claims-auditor',
                         '601_claims_auditor_page_text_strips_per_component.sql: pre-update');
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,load_page_text,config,query}',
         to_jsonb($Q$SELECT p.name, LEFT(regexp_replace(string_agg(regexp_replace(regexp_replace(regexp_replace(pc.rendered_html, '<style[^>]*?>.*?</style>', ' ', 'gi'), '<script[^>]*?>.*?</script>', ' ', 'gi'), '<[^>]*>', ' ', 'g'), ' ' ORDER BY pc.position, pc.slot_name), '\s+', ' ', 'g'), 12000) AS page_text FROM pages p JOIN page_components pc ON pc.page_id = p.id WHERE p.site_id = $1 AND p.build_status IN ('deployed','active') AND pc.rendered_html IS NOT NULL AND pc.rendered_html <> '' AND pc.locked_at IS NULL GROUP BY p.id, p.name ORDER BY p.name$Q$::text),
         false),
       updated_at = now()
 WHERE type='claims-auditor' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Verify: the new query must (a) be stored, (b) recover the motivating sentence on the
-- motivating site when executed — the demand control, not a syntax check.
DO $$
DECLARE q text; n int; hwa_len int; has_sentence boolean;
BEGIN
  SELECT default_config #>> '{workflow,steps,load_page_text,config,query}' INTO q
    FROM agent_definitions WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF position('ORDER BY pc.position, pc.slot_name' in q) = 0
     OR position('string_agg(regexp_replace(' in q) = 0 THEN
    RAISE EXCEPTION '601 VERIFY: per-component strip not stored';
  END IF;

  -- Demand control on the motivating site/page: the SAME expression the stored query uses,
  -- run explicitly (null on a fresh DB where garden-tools.uk does not exist — then only
  -- the storage check above applies).
  SELECT length(t), position('we buy the tool' in t) > 0 INTO hwa_len, has_sentence
    FROM (
      SELECT LEFT(regexp_replace(string_agg(regexp_replace(regexp_replace(regexp_replace(pc.rendered_html, '<style[^>]*?>.*?</style>', ' ', 'gi'), '<script[^>]*?>.*?</script>', ' ', 'gi'), '<[^>]*>', ' ', 'g'), ' ' ORDER BY pc.position, pc.slot_name), '\s+', ' ', 'g'), 12000) AS t
        FROM pages p JOIN page_components pc ON pc.page_id = p.id JOIN sites s ON s.id = p.site_id
       WHERE s.domain='garden-tools.uk' AND p.name='how-we-assess'
         AND pc.rendered_html IS NOT NULL AND pc.rendered_html <> '' AND pc.locked_at IS NULL
       GROUP BY p.id) x;
  IF hwa_len IS NOT NULL AND NOT has_sentence THEN
    RAISE EXCEPTION '601 VERIFY: the motivating sentence still does not survive extraction (how-we-assess text % chars)', hwa_len;
  END IF;
  RAISE NOTICE '601 OK: per-component strip stored; how-we-assess extracts % chars with the sentence present=%', hwa_len, has_sentence;
END $$;

COMMIT;
