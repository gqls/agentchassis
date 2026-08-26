-- 646 — de-demonstrate finetuning.uk's content_direction: the brief teaches the
--       construction it bans, and that is the floor this lane measured three times
--
-- WHY. copy_quality_two_stage shipped 627/628/629/630, which cut the fleet writer
-- prompt's negation demonstrations 63-65 -> 36. But content_direction is SITE-owned
-- and those migrations cannot reach it: finetuning.uk's brief still scans **7
-- 'rather than'** [MEASURED 2026-08-26, independently, not taken on report] — the
-- exact demonstrations this lane measured as its immovable floor across three builds
-- (tells floored 9->9->6 while brief demonstrations went to zero everywhere else).
-- Their mechanism, proven in their lane: classes track their demonstration counts.
-- So the demonstrations have to go from HERE or a rebuild reproduces them.
--
-- FORM ONLY. Every substitution below preserves the instruction and removes the
-- demonstrated construction. No rule loses force; the ISN'T rule keeps its ban and
-- loses only the two exemplars that performed the banned shape inside it.
--
-- ⚠ DELIBERATELY NOT TOUCHED, and the distinction is the point:
--   * quoted BRAND POSITIONING — "we pick the best tool, not our favourite vendor" —
--     changing it changes what the site claims, not how the brief teaches;
--   * quoted EXAMPLE UTTERANCES the voice may emit — "'this won't work for everyone'
--     or 'AI isn't the right answer here'" — those are sentences we would happily
--     publish, so they are content, not a form lesson;
--   * rules that NAME a banned phrase ('cutting-edge', 'act now'). A ban must be
--     allowed to name what it bans.
--
-- ⚠ 'formatted' IS A DERIVED FIELD (FormatContentDirection: sorted keys, skips
-- 'formatted' itself, called by write_site_spec). Editing it alone would be erased by
-- the next spec write — the surface-the-renderer-overwrites trap. So the SAME literal
-- substitutions are applied to the whole document, source keys AND formatted, which
-- is byte-equivalent to what a regeneration would produce because each needle sits
-- inside one string value and never spans the formatter's scaffolding.
--
-- Rollback: 646_..._ROLLBACK.sql

BEGIN;

UPDATE site_specs ss SET is_current = false, superseded_at = now()
  FROM sites s WHERE s.id = ss.site_id AND s.domain = 'finetuning.uk'
   AND ss.aspect = 'content_direction' AND ss.is_current;

INSERT INTO site_specs (site_id, aspect, data, source, created_by, is_current, notes)
SELECT ss.site_id, ss.aspect,
       (replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(ss.data::text, 'invite a conversation, not a commitment. Frame as a starting point rather than a closing action.', 'invite a conversation. Frame the ask as a starting point.'), 'let warmth come from taking time to explain rather than from hype.', 'let warmth come from taking time to explain.'), 'real data sources rather than generic claims', 'real data sources'), 'address it head-on rather than ignoring it', 'address it head-on'), '''analysing 10,000 supplier invoices overnight'' rather than ''processing large volumes of data''', '''analysing 10,000 supplier invoices overnight'''), '''a veterinary industry data project'' rather than ''a leading healthcare innovator''', '''a veterinary industry data project'''), 'real outcomes rather than generic promises.', 'real outcomes.'), '''we build automation pipelines that handle invoice processing'' not ''we automate your business''.', '''we build automation pipelines that handle invoice processing''.'), 'what problem it solves, not just what it does technically.', 'the problem it solves as well as the mechanism.'), 'Never open a sentence or section with what something ISN''T, or with a negative frame that sets up a reveal (''X isn''t Y. It''s Z'', ''Nothing here is X''). This applies in every grammatical form.', 'Open every sentence and section on what a thing IS. A negative frame that sets up a reveal is the shape to avoid, in every grammatical form.'))::jsonb,
       'owner_ruling', 'claude-finetuning-uk-lane', true,
       'De-demonstrated 2026-08-26: 7 rather-than + 3 further construction demonstrations removed form-only, per copy_quality_two_stage measurement that this site-owned brief is the layer 627-630 cannot reach. Meaning preserved; brand positioning and example utterances untouched.'
  FROM site_specs ss JOIN sites s ON s.id = ss.site_id
 WHERE s.domain = 'finetuning.uk' AND ss.aspect = 'content_direction'
   AND ss.superseded_at IS NOT NULL
 ORDER BY ss.superseded_at DESC LIMIT 1;

DO $$
DECLARE n_rt int; n_fmt int; n_keys int; len_before int; len_after int;
BEGIN
  SELECT (length(ss.data::text) - length(replace(ss.data::text,'rather than',''))) / length('rather than'),
         (SELECT count(*) FROM jsonb_object_keys(ss.data)),
         length(ss.data->>'formatted')
    INTO n_rt, n_keys, len_after
    FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='finetuning.uk' AND ss.aspect='content_direction' AND ss.is_current;

  IF n_rt <> 0 THEN RAISE EXCEPTION '646: % rather-than demonstration(s) remain in the brief', n_rt; END IF;
  IF n_keys < 8 THEN RAISE EXCEPTION '646: only % top-level keys survived — a substitution ate structure', n_keys; END IF;
  IF len_after IS NULL OR len_after < 10000 THEN RAISE EXCEPTION '646: formatted is % chars, want ~12k — it was truncated or lost', len_after; END IF;

  -- formatted must have moved in step with the source keys, or the two have drifted.
  SELECT (length(ss.data->>'formatted') - length(replace(ss.data->>'formatted','rather than',''))) / length('rather than')
    INTO n_fmt FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='finetuning.uk' AND ss.aspect='content_direction' AND ss.is_current;
  IF n_fmt <> 0 THEN RAISE EXCEPTION '646: formatted still carries % rather-than — it drifted from the source keys', n_fmt; END IF;

  RAISE NOTICE '646 OK: 0 rather-than in brief and in formatted, % keys intact, formatted % chars', n_keys, len_after;
END $$;

COMMIT;
