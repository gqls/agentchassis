-- 668_offer_ordering_expansion_10_points.sql
--
-- OWNER DECISION C, wash step 2: the 10 points the offer lane excluded from 667 as
-- substantive (their >=40%-on-differentiated rule), repaired by EXPANSION — rulings 7+13
-- as a pair: remove the shape, keep the distinction, longer where needed. Verdict trail:
-- their CONTRIB 2026-08-31d + the in-session final verdict — 4 unreserved, 1 with a
-- wording fix applied ('the full record' -> 'the record': invented completeness), 2
-- judgement calls accepted as SAFER than their originals (each original carried an
-- unevidenced third-party claim), 3 RE-DERIVED BY THEM from strategy/evidence provenance
-- after the expansion invented substance (the method note: it invented a UI mechanism and
-- an overclaim, not grounded facts). Independent battery by this lane on the final 10:
-- v1-clean ('plain words' in mortgagecalculator r6 flagged as a register-labelling-by-read
-- candidate for BANNED_REGISTER v2, present in the original, both lanes accepted).
-- Identity by exact text with drift RAISEs (the corpus MOVES: the producer re-minted
-- mid-wash on 08-31); ranks pinned. ROLLBACK: 668_..._ROLLBACK.sql.

BEGIN;

INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '668_offer_ordering_expansion_10_points', 'site_specs', sp.id::text,
       jsonb_build_object('data', sp.data), 'pre-668 offer_ordering for ' || s.domain
FROM site_specs sp JOIN sites s ON s.id = sp.site_id
WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain IN (
'agritec.uk', 'finetuning.uk', 'fundamentallyai.com', 'garden-tools.uk', 'homegarden.uk', 'lampenkap.com', 'mortgagecalculator.co.uk', 'noted.co.uk');

DO $mig$
DECLARE
  lw jsonb;
  n int;
BEGIN

  -- ── agritec.uk ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='agritec.uk';
  IF lw IS NULL THEN RAISE EXCEPTION '668: no offer_ordering lead_with for agritec.uk'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Where a figure is not yet verified, this site says so plainly rather than substituting an estimate — because a plausible guess presented as a sourced figure is precisely the failure the citation infrastructure exists to prevent.';
  IF n <> 1 THEN RAISE EXCEPTION '668: agritec.uk rank 6: FROM matches % elements, want 1 - drifted, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Where a figure is not yet verified, this site says so plainly rather than substituting an estimate — because a plausible guess presented as a sourced figure is precisely the failure the citation infrastructure exists to prevent.' THEN jsonb_set(e, '{point}', to_jsonb('Where a figure has not been verified, this site publishes that it has not been verified and leaves the gap visible, so a reader always knows which numbers carry a checked source.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='agritec.uk';

  -- ── finetuning.uk ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='finetuning.uk';
  IF lw IS NULL THEN RAISE EXCEPTION '668: no offer_ordering lead_with for finetuning.uk'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Your client and operational data stays private — we deploy on infrastructure you control, not third-party cloud tools that train on what you paste into them.';
  IF n <> 1 THEN RAISE EXCEPTION '668: finetuning.uk rank 3: FROM matches % elements, want 1 - drifted, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Your client and operational data stays private — we deploy on infrastructure you control, not third-party cloud tools that train on what you paste into them.' THEN jsonb_set(e, '{point}', to_jsonb('Your client and operational data stays private because we deploy on infrastructure you control. Everything you put into the system remains within your own environment and stays yours alone.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'You get a working system, not a prototype that stalls after the first demo.';
  IF n <> 1 THEN RAISE EXCEPTION '668: finetuning.uk rank 4: FROM matches % elements, want 1 - drifted, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'You get a working system, not a prototype that stalls after the first demo.' THEN jsonb_set(e, '{point}', to_jsonb('You get a working system that runs in real, daily use and keeps delivering long after the first demo.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='finetuning.uk';

  -- ── fundamentallyai.com ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='fundamentallyai.com';
  IF lw IS NULL THEN RAISE EXCEPTION '668: no offer_ordering lead_with for fundamentallyai.com'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'This platform has publicly caught and corrected its own AI-generated errors — naming the specific site where the error appeared and the correction that followed — because a buyer who has been burned by overpromised AI projects deserves an honest account, not a polished one.';
  IF n <> 1 THEN RAISE EXCEPTION '668: fundamentallyai.com rank 2: FROM matches % elements, want 1 - drifted, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'This platform has publicly caught and corrected its own AI-generated errors — naming the specific site where the error appeared and the correction that followed — because a buyer who has been burned by overpromised AI projects deserves an honest account, not a polished one.' THEN jsonb_set(e, '{point}', to_jsonb('This platform publishes public corrections of errors in its own AI-generated content. Each published correction names the specific site where the error appeared and describes the fix that was made, so a buyer who has been burned by overpromised AI projects can read the record of what went wrong and how it was repaired.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='fundamentallyai.com';

  -- ── garden-tools.uk ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='garden-tools.uk';
  IF lw IS NULL THEN RAISE EXCEPTION '668: no offer_ordering lead_with for garden-tools.uk'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Honest coverage at every price point, including the budget end that premium-brand sites and broad editorial sites rarely treat seriously.';
  IF n <> 1 THEN RAISE EXCEPTION '668: garden-tools.uk rank 3: FROM matches % elements, want 1 - drifted, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Honest coverage at every price point, including the budget end that premium-brand sites and broad editorial sites rarely treat seriously.' THEN jsonb_set(e, '{point}', to_jsonb('This site covers tools at every price point, from allotment-practical to long-term investment, and records where a tool falls short as readily as where it performs.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='garden-tools.uk';

  -- ── homegarden.uk ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='homegarden.uk';
  IF lw IS NULL THEN RAISE EXCEPTION '668: no offer_ordering lead_with for homegarden.uk'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'The answer here will say plainly whether your timing varies by region — rather than giving one confident number that is wrong for half the country.';
  IF n <> 1 THEN RAISE EXCEPTION '668: homegarden.uk rank 2: FROM matches % elements, want 1 - drifted, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'The answer here will say plainly whether your timing varies by region — rather than giving one confident number that is wrong for half the country.' THEN jsonb_set(e, '{point}', to_jsonb('When a task''s timing depends on where you live, the answer says how it varies by region, so you can tell whether a national date applies to your own area.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='homegarden.uk';

  -- ── lampenkap.com ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='lampenkap.com';
  IF lw IS NULL THEN RAISE EXCEPTION '668: no offer_ordering lead_with for lampenkap.com'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'The photometric reasoning here is grounded in specific quantities — Kelvin values, lumen figures, reflectivity percentages — so that the analysis reads as a calculated conclusion rather than a commercial suggestion.';
  IF n <> 1 THEN RAISE EXCEPTION '668: lampenkap.com rank 3: FROM matches % elements, want 1 - drifted, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'The photometric reasoning here is grounded in specific quantities — Kelvin values, lumen figures, reflectivity percentages — so that the analysis reads as a calculated conclusion rather than a commercial suggestion.' THEN jsonb_set(e, '{point}', to_jsonb('The photometric reasoning here is grounded in specific quantities: Kelvin values, lumen figures, reflectivity percentages. That grounding means the analysis reads as a calculated conclusion you can check against the numbers.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='lampenkap.com';

  -- ── mortgagecalculator.co.uk ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='mortgagecalculator.co.uk';
  IF lw IS NULL THEN RAISE EXCEPTION '668: no offer_ordering lead_with for mortgagecalculator.co.uk'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'The guides here are written in plain words and tell you what lenders look for, not what lenders say they look for — there is a difference, and it matters when you apply.';
  IF n <> 1 THEN RAISE EXCEPTION '668: mortgagecalculator.co.uk rank 6: FROM matches % elements, want 1 - drifted, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'The guides here are written in plain words and tell you what lenders look for, not what lenders say they look for — there is a difference, and it matters when you apply.' THEN jsonb_set(e, '{point}', to_jsonb('The guides here are written in plain words. They tell you the criteria lenders genuinely apply when they assess an application, and knowing those real criteria matters when you apply.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'If you want to understand what the bank actually sees when it looks at your application — not just what you hope to borrow — this site is the only UK resource built to show you that.';
  IF n <> 1 THEN RAISE EXCEPTION '668: mortgagecalculator.co.uk rank 3: FROM matches % elements, want 1 - drifted, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'If you want to understand what the bank actually sees when it looks at your application — not just what you hope to borrow — this site is the only UK resource built to show you that.' THEN jsonb_set(e, '{point}', to_jsonb('This site is the only UK resource built to show you what the bank actually sees when it looks at your application. It covers the amount you hope to borrow and everything else the lender weighs when making its decision.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='mortgagecalculator.co.uk';

  -- ── noted.co.uk ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='noted.co.uk';
  IF lw IS NULL THEN RAISE EXCEPTION '668: no offer_ordering lead_with for noted.co.uk'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Voice recordings are kept as recordings — not converted to text, not discarded.';
  IF n <> 1 THEN RAISE EXCEPTION '668: noted.co.uk rank 4: FROM matches % elements, want 1 - drifted, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Voice recordings are kept as recordings — not converted to text, not discarded.' THEN jsonb_set(e, '{point}', to_jsonb('Voice recordings are preserved in their original audio form, and every recording stays saved for you exactly as you made it.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='noted.co.uk';

  RAISE NOTICE '668 OK: the 10 expansion points applied across 8 sites; ranks untouched; the wash of the ACKed corpus is COMPLETE.';
END $mig$;

COMMIT;
