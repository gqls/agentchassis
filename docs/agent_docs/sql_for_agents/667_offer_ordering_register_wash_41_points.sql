-- 667_offer_ordering_register_wash_41_points.sql
--
-- OWNER DECISION C (2026-08-31, in-session: "Please go ahead and wire in the benefit
-- priorities"), step 1 of the amended-C order: the REGISTER WASH over offer_ordering's
-- lead_with points, ACKed by the owning offer-analysis lane 41-of-51 (their CONTRIB
-- 2026-08-31c, commit 906a669ff; the 10 exclusions are theirs to re-judge and are NOT
-- touched here — their >=40%-reduction-on-differentiated rule, plus noted.co.uk r4's
-- factual data-handling commitment, a different act from a comparison truncation).
-- Every replacement: repaired under ruling 7's truncation instruction (fable-5 batch,
-- no demonstrations in the prompt), battery-verified by BOTH lanes independently,
-- RANK PINNED, text-only. Identity is the point's exact text, never an array index.
-- ROLLBACK: 667_..._ROLLBACK.sql (restores each touched site's offer_ordering row
-- from migration_backups).

BEGIN;

INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '667_offer_ordering_register_wash_41_points', 'site_specs', sp.id::text,
       jsonb_build_object('data', sp.data), 'pre-667 offer_ordering for ' || s.domain
FROM site_specs sp JOIN sites s ON s.id = sp.site_id
WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain IN (
'adversecreditmortgage.co.uk', 'agritec.uk', 'ai-agent-orchestration.com', 'cookly.uk', 'cv1.co.uk', 'dartsonline.com', 'farmerinsurance.uk', 'finetuning.uk', 'fundamentallyai.com', 'gamesdesign.co.uk', 'garden-tools.uk', 'gaswholesalers.com', 'leopardessconsulting.co.uk', 'loanandmortgagecalculator.co.uk', 'loancalculator.co.uk', 'mortgagecalculator.co.uk', 'oufe.com', 'robot-hands.com', 'vonc.com', 'webdesign.co.uk');

DO $mig$
DECLARE
  lw jsonb;
  n int;
BEGIN

  -- ── adversecreditmortgage.co.uk ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='adversecreditmortgage.co.uk';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for adversecreditmortgage.co.uk'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'The decision-tree tool routes you through your specific combination — type, age, satisfaction status — and returns a structured explanation of what that typically means for a lender, rather than a generic article.';
  IF n <> 1 THEN RAISE EXCEPTION '667: adversecreditmortgage.co.uk rank 3: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'The decision-tree tool routes you through your specific combination — type, age, satisfaction status — and returns a structured explanation of what that typically means for a lender, rather than a generic article.' THEN jsonb_set(e, '{point}', to_jsonb('The decision-tree tool routes you through your specific combination of type, age, and satisfaction status and returns a structured explanation of what that typically means for a lender.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'When your circumstances change — a CCJ ages past a lender threshold, a DMP ends, six months of clean conduct passes — the guidance here updates to reflect where you stand at that point, not where you stood when you first arrived.';
  IF n <> 1 THEN RAISE EXCEPTION '667: adversecreditmortgage.co.uk rank 5: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'When your circumstances change — a CCJ ages past a lender threshold, a DMP ends, six months of clean conduct passes — the guidance here updates to reflect where you stand at that point, not where you stood when you first arrived.' THEN jsonb_set(e, '{point}', to_jsonb('When your circumstances change, whether a CCJ ages past a lender threshold, a DMP ends, or six months of clean conduct passes, the guidance here updates to reflect where you stand at that point.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Where your situation is smaller than you fear and where it genuinely limits your options, this resource says so plainly — because this audience has been given enough false hope already.';
  IF n <> 1 THEN RAISE EXCEPTION '667: adversecreditmortgage.co.uk rank 4: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Where your situation is smaller than you fear and where it genuinely limits your options, this resource says so plainly — because this audience has been given enough false hope already.' THEN jsonb_set(e, '{point}', to_jsonb('Where your situation is smaller than you fear and where it genuinely limits your options, this resource says so.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'If your situation is active financial difficulty you cannot service right now, this is not the right starting point — and this resource will tell you where to go instead.';
  IF n <> 1 THEN RAISE EXCEPTION '667: adversecreditmortgage.co.uk rank 6: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'If your situation is active financial difficulty you cannot service right now, this is not the right starting point — and this resource will tell you where to go instead.' THEN jsonb_set(e, '{point}', to_jsonb('If your situation is active financial difficulty you cannot service right now, this is not the right starting point, and this resource will tell you where to go instead.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='adversecreditmortgage.co.uk';

  -- ── agritec.uk ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='agritec.uk';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for agritec.uk'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Every figure you encounter here carries an inline link to the primary document that states it, so you can check the rate, the constant, or the payment threshold in one click rather than trusting us to have read it correctly.';
  IF n <> 1 THEN RAISE EXCEPTION '667: agritec.uk rank 1: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Every figure you encounter here carries an inline link to the primary document that states it, so you can check the rate, the constant, or the payment threshold in one click rather than trusting us to have read it correctly.' THEN jsonb_set(e, '{point}', to_jsonb('Every figure you encounter here carries an inline link to the primary document that states it, so you can check the rate, the constant, or the payment threshold in one click.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'The technical explainers teach the governing mechanism — the physics, chemistry, biology, or policy — so you can evaluate whether a result from any source is telling you something real, not just read off a number from this one.';
  IF n <> 1 THEN RAISE EXCEPTION '667: agritec.uk rank 3: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'The technical explainers teach the governing mechanism — the physics, chemistry, biology, or policy — so you can evaluate whether a result from any source is telling you something real, not just read off a number from this one.' THEN jsonb_set(e, '{point}', to_jsonb('The technical explainers teach the governing mechanism, the physics, chemistry, biology, or policy, so you can evaluate whether a result from any source is telling you something real.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='agritec.uk';

  -- ── ai-agent-orchestration.com ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='ai-agent-orchestration.com';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for ai-agent-orchestration.com'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'The next step is a Technical Discovery Call — an engineering discussion about what broke in your pipeline, not a sales process.';
  IF n <> 1 THEN RAISE EXCEPTION '667: ai-agent-orchestration.com rank 6: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'The next step is a Technical Discovery Call — an engineering discussion about what broke in your pipeline, not a sales process.' THEN jsonb_set(e, '{point}', to_jsonb('The next step is a Technical Discovery Call: an engineering discussion about what broke in your pipeline.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='ai-agent-orchestration.com';

  -- ── cookly.uk ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='cookly.uk';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for cookly.uk'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Know what tonight''s dinner will cost before you leave the house — cost per portion is shown upfront, not buried after you''ve already committed to cooking.';
  IF n <> 1 THEN RAISE EXCEPTION '667: cookly.uk rank 2: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Know what tonight''s dinner will cost before you leave the house — cost per portion is shown upfront, not buried after you''ve already committed to cooking.' THEN jsonb_set(e, '{point}', to_jsonb('Know what tonight''s dinner will cost before you leave the house: cost per portion is shown upfront.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Ingredient swap suggestions are flagged upfront so the plan adapts to what''s actually in stock or affordable this week, not what a recipe assumed you''d have.';
  IF n <> 1 THEN RAISE EXCEPTION '667: cookly.uk rank 5: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Ingredient swap suggestions are flagged upfront so the plan adapts to what''s actually in stock or affordable this week, not what a recipe assumed you''d have.' THEN jsonb_set(e, '{point}', to_jsonb('Ingredient swap suggestions are flagged upfront so the plan adapts to what''s actually in stock or affordable this week.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='cookly.uk';

  -- ── cv1.co.uk ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='cv1.co.uk';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for cv1.co.uk'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'You can work through your job search in a structured, organised way — using a private interactive checklist built from your actual CV and target roles, not a generic template.';
  IF n <> 1 THEN RAISE EXCEPTION '667: cv1.co.uk rank 1: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'You can work through your job search in a structured, organised way — using a private interactive checklist built from your actual CV and target roles, not a generic template.' THEN jsonb_set(e, '{point}', to_jsonb('You can work through your job search in a structured, organised way, using a private interactive checklist built from your actual CV and target roles.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='cv1.co.uk';

  -- ── dartsonline.com ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='dartsonline.com';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for dartsonline.com'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'When you change your barrel weight or switch shaft length, the Setup Builder is here to revisit — so your setup evolves with your throw rather than staying fixed at first purchase.';
  IF n <> 1 THEN RAISE EXCEPTION '667: dartsonline.com rank 5: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'When you change your barrel weight or switch shaft length, the Setup Builder is here to revisit — so your setup evolves with your throw rather than staying fixed at first purchase.' THEN jsonb_set(e, '{point}', to_jsonb('When you change your barrel weight or switch shaft length, the Setup Builder is here to revisit, so your setup evolves with your throw.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Every guide states tungsten percentages, barrel weights, and shaft lengths plainly, explains what each one changes about your throw, and helps you leave knowing exactly what you''re throwing and why.';
  IF n <> 1 THEN RAISE EXCEPTION '667: dartsonline.com rank 2: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Every guide states tungsten percentages, barrel weights, and shaft lengths plainly, explains what each one changes about your throw, and helps you leave knowing exactly what you''re throwing and why.' THEN jsonb_set(e, '{point}', to_jsonb('Every guide states tungsten percentages, barrel weights, and shaft lengths, explains what each one changes about your throw, and helps you leave knowing exactly what you''re throwing and why.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='dartsonline.com';

  -- ── farmerinsurance.uk ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='farmerinsurance.uk';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for farmerinsurance.uk'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Farm insurance decisions involve significant financial exposure — a misunderstood exclusion can cost tens of thousands of pounds — so every guide here names the actual policy terms and FCA rules, not a paraphrase of them.';
  IF n <> 1 THEN RAISE EXCEPTION '667: farmerinsurance.uk rank 3: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Farm insurance decisions involve significant financial exposure — a misunderstood exclusion can cost tens of thousands of pounds — so every guide here names the actual policy terms and FCA rules, not a paraphrase of them.' THEN jsonb_set(e, '{point}', to_jsonb('Farm insurance decisions involve significant financial exposure, and a misunderstood exclusion can cost tens of thousands of pounds, so every guide here names the actual policy terms and FCA rules.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Unlike insurer or broker sites, this resource has no policy to sell and no panel to route you to — so every guide can tell you what to watch for in a policy document, not just what to buy.';
  IF n <> 1 THEN RAISE EXCEPTION '667: farmerinsurance.uk rank 2: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Unlike insurer or broker sites, this resource has no policy to sell and no panel to route you to — so every guide can tell you what to watch for in a policy document, not just what to buy.' THEN jsonb_set(e, '{point}', to_jsonb('This resource has no policy to sell and no panel to route you to, so every guide can tell you what to watch for in a policy document.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'The illustrative tools — rebuild cost estimator, machinery declared value calculator, livestock value estimator — give you a working figure to bring to a broker conversation, not a quote and not a commitment.';
  IF n <> 1 THEN RAISE EXCEPTION '667: farmerinsurance.uk rank 5: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'The illustrative tools — rebuild cost estimator, machinery declared value calculator, livestock value estimator — give you a working figure to bring to a broker conversation, not a quote and not a commitment.' THEN jsonb_set(e, '{point}', to_jsonb('The illustrative tools (rebuild cost estimator, machinery declared value calculator, livestock value estimator) give you a working figure to bring to a broker conversation, not a quote and not a commitment.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='farmerinsurance.uk';

  -- ── finetuning.uk ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='finetuning.uk';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for finetuning.uk'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'If your team spends hours each week on work that follows the same pattern every time, AI can likely remove most of it — and we will tell you plainly if it cannot.';
  IF n <> 1 THEN RAISE EXCEPTION '667: finetuning.uk rank 1: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'If your team spends hours each week on work that follows the same pattern every time, AI can likely remove most of it — and we will tell you plainly if it cannot.' THEN jsonb_set(e, '{point}', to_jsonb('If your team spends hours each week on work that follows the same pattern every time, AI can likely remove most of it, and we will tell you if it cannot.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Once we have built the system, practical articles and updated case studies give you an honest, ongoing picture of where AI is actually working for businesses like yours.';
  IF n <> 1 THEN RAISE EXCEPTION '667: finetuning.uk rank 5: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Once we have built the system, practical articles and updated case studies give you an honest, ongoing picture of where AI is actually working for businesses like yours.' THEN jsonb_set(e, '{point}', to_jsonb('Once we have built the system, practical articles and updated case studies give you an ongoing picture of where AI is actually working for businesses like yours.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'The next step is a single conversation — no commitment, no sales pitch — just an honest assessment of whether your problem is one AI can solve.';
  IF n <> 1 THEN RAISE EXCEPTION '667: finetuning.uk rank 6: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'The next step is a single conversation — no commitment, no sales pitch — just an honest assessment of whether your problem is one AI can solve.' THEN jsonb_set(e, '{point}', to_jsonb('The next step is a single conversation, an assessment of whether your problem is one AI can solve.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'We pick the best tool for your problem, not our preferred vendor — and if AI is not the right answer, we will say so before you spend anything.';
  IF n <> 1 THEN RAISE EXCEPTION '667: finetuning.uk rank 2: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'We pick the best tool for your problem, not our preferred vendor — and if AI is not the right answer, we will say so before you spend anything.' THEN jsonb_set(e, '{point}', to_jsonb('We pick the best tool for your problem, and if AI is not the right answer, we will say so before you spend anything.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='finetuning.uk';

  -- ── fundamentallyai.com ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='fundamentallyai.com';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for fundamentallyai.com'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'If your organisation needs private semantic search over internal data, the technical groundwork already exists and works — and the site tells you that directly rather than implying a finished product.';
  IF n <> 1 THEN RAISE EXCEPTION '667: fundamentallyai.com rank 4: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'If your organisation needs private semantic search over internal data, the technical groundwork already exists and works — and the site tells you that directly rather than implying a finished product.' THEN jsonb_set(e, '{point}', to_jsonb('If your organisation needs private semantic search over internal data, the technical groundwork already exists and works.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Every time the platform makes a substantial change, a new dated decision record entry is added — so returning here gives you fresh evidence that the governance is live, not a static claim made once and left.';
  IF n <> 1 THEN RAISE EXCEPTION '667: fundamentallyai.com rank 5: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Every time the platform makes a substantial change, a new dated decision record entry is added — so returning here gives you fresh evidence that the governance is live, not a static claim made once and left.' THEN jsonb_set(e, '{point}', to_jsonb('Every time the platform makes a substantial change, a new dated decision record entry is added, so returning here gives you fresh evidence that the governance is live.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'You can evaluate what this platform has demonstrated on its own infrastructure separately from what it has delivered for a paying client, because the site states the distinction plainly rather than obscuring it.';
  IF n <> 1 THEN RAISE EXCEPTION '667: fundamentallyai.com rank 3: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'You can evaluate what this platform has demonstrated on its own infrastructure separately from what it has delivered for a paying client, because the site states the distinction plainly rather than obscuring it.' THEN jsonb_set(e, '{point}', to_jsonb('You can evaluate what this platform has demonstrated on its own infrastructure separately from what it has delivered for a paying client, because the site states the distinction.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='fundamentallyai.com';

  -- ── gamesdesign.co.uk ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='gamesdesign.co.uk';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for gamesdesign.co.uk'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'The tool inventory expands as new calculators cover additional game systems, so the platform grows alongside your career rather than becoming a resource you read once.';
  IF n <> 1 THEN RAISE EXCEPTION '667: gamesdesign.co.uk rank 6: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'The tool inventory expands as new calculators cover additional game systems, so the platform grows alongside your career rather than becoming a resource you read once.' THEN jsonb_set(e, '{point}', to_jsonb('The tool inventory expands as new calculators cover additional game systems, so the platform grows alongside your career.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='gamesdesign.co.uk';

  -- ── garden-tools.uk ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='garden-tools.uk';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for garden-tools.uk'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Find the right tool for your specific garden — your soil, your budget, your physical ability — not just a ranked list of products.';
  IF n <> 1 THEN RAISE EXCEPTION '667: garden-tools.uk rank 1: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Find the right tool for your specific garden — your soil, your budget, your physical ability — not just a ranked list of products.' THEN jsonb_set(e, '{point}', to_jsonb('Find the right tool for your specific garden: your soil, your budget, your physical ability.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='garden-tools.uk';

  -- ── gaswholesalers.com ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='gaswholesalers.com';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for gaswholesalers.com'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'You deal directly with a wholesale fuel distribution business, not a broker or intermediary, so negotiations, volume commitments, and delivery logistics are handled in one relationship.';
  IF n <> 1 THEN RAISE EXCEPTION '667: gaswholesalers.com rank 3: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'You deal directly with a wholesale fuel distribution business, not a broker or intermediary, so negotiations, volume commitments, and delivery logistics are handled in one relationship.' THEN jsonb_set(e, '{point}', to_jsonb('You deal directly with a wholesale fuel distribution business, so negotiations, volume commitments, and delivery logistics are handled in one relationship.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='gaswholesalers.com';

  -- ── leopardessconsulting.co.uk ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='leopardessconsulting.co.uk';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for leopardessconsulting.co.uk'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Working systems run on Kubernetes, Kafka, and Postgres in production now — not proposals, not prototypes, and the infrastructure choices are explained so you can evaluate them.';
  IF n <> 1 THEN RAISE EXCEPTION '667: leopardessconsulting.co.uk rank 1: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Working systems run on Kubernetes, Kafka, and Postgres in production now — not proposals, not prototypes, and the infrastructure choices are explained so you can evaluate them.' THEN jsonb_set(e, '{point}', to_jsonb('Working systems run on Kubernetes, Kafka, and Postgres in production now, and the infrastructure choices are explained so you can evaluate them.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Leopardess designs and deploys hierarchical multi-agent AI systems on Kubernetes, Kafka, and Postgres for UK B2B SaaS engineering teams — delivering working orchestration infrastructure in days, not months.';
  IF n <> 1 THEN RAISE EXCEPTION '667: leopardessconsulting.co.uk rank 2: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Leopardess designs and deploys hierarchical multi-agent AI systems on Kubernetes, Kafka, and Postgres for UK B2B SaaS engineering teams — delivering working orchestration infrastructure in days, not months.' THEN jsonb_set(e, '{point}', to_jsonb('Leopardess designs and deploys hierarchical multi-agent AI systems on Kubernetes, Kafka, and Postgres for UK B2B SaaS engineering teams, delivering working orchestration infrastructure in days.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Reliability, observability, and security are not traded away for speed — the stack is designed so that when something fails it can be picked up where it stopped rather than started again.';
  IF n <> 1 THEN RAISE EXCEPTION '667: leopardessconsulting.co.uk rank 6: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Reliability, observability, and security are not traded away for speed — the stack is designed so that when something fails it can be picked up where it stopped rather than started again.' THEN jsonb_set(e, '{point}', to_jsonb('Reliability, observability, and security are not traded away for speed; the stack is designed so that when something fails it can be picked up where it stopped.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='leopardessconsulting.co.uk';

  -- ── loanandmortgagecalculator.co.uk ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='loanandmortgagecalculator.co.uk';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for loanandmortgagecalculator.co.uk'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'The guidance here explains mechanisms — the exact figures behind how lenders calculate affordability — rather than quoting current rates, so it does not silently go stale.';
  IF n <> 1 THEN RAISE EXCEPTION '667: loanandmortgagecalculator.co.uk rank 6: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'The guidance here explains mechanisms — the exact figures behind how lenders calculate affordability — rather than quoting current rates, so it does not silently go stale.' THEN jsonb_set(e, '{point}', to_jsonb('The guidance here explains mechanisms, the exact figures behind how lenders calculate affordability, so it does not silently go stale.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Your borrowing situation changes each time a deal ends, a loan is paid off, or rates move — these tools stay useful at every stage of that journey, not just for a single lookup.';
  IF n <> 1 THEN RAISE EXCEPTION '667: loanandmortgagecalculator.co.uk rank 5: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Your borrowing situation changes each time a deal ends, a loan is paid off, or rates move — these tools stay useful at every stage of that journey, not just for a single lookup.' THEN jsonb_set(e, '{point}', to_jsonb('Your borrowing situation changes each time a deal ends, a loan is paid off, or rates move; these tools stay useful at every stage of that journey.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='loanandmortgagecalculator.co.uk';

  -- ── loancalculator.co.uk ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='loancalculator.co.uk';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for loancalculator.co.uk'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'The calculator output is the start of understanding your situation, not the end of a comparison funnel — the numbers are explained in plain English so you know what they mean and what you can do about them.';
  IF n <> 1 THEN RAISE EXCEPTION '667: loancalculator.co.uk rank 2: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'The calculator output is the start of understanding your situation, not the end of a comparison funnel — the numbers are explained in plain English so you know what they mean and what you can do about them.' THEN jsonb_set(e, '{point}', to_jsonb('The calculator output is the start of understanding your situation: the numbers are explained so you know what they mean and what you can do about them.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='loancalculator.co.uk';

  -- ── mortgagecalculator.co.uk ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='mortgagecalculator.co.uk';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for mortgagecalculator.co.uk'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'You will leave with a specific number and know what it means for your next step — not a range, not a rough guide.';
  IF n <> 1 THEN RAISE EXCEPTION '667: mortgagecalculator.co.uk rank 5: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'You will leave with a specific number and know what it means for your next step — not a range, not a rough guide.' THEN jsonb_set(e, '{point}', to_jsonb('You will leave with a specific number and know what it means for your next step.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='mortgagecalculator.co.uk';

  -- ── oufe.com ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='oufe.com';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for oufe.com'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'You can move the assumptions yourself and observe directly how creditor recovery changes — the interactive tools are built on verified figures from named documents, not illustrative placeholders.';
  IF n <> 1 THEN RAISE EXCEPTION '667: oufe.com rank 2: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'You can move the assumptions yourself and observe directly how creditor recovery changes — the interactive tools are built on verified figures from named documents, not illustrative placeholders.' THEN jsonb_set(e, '{point}', to_jsonb('You can move the assumptions yourself and observe directly how creditor recovery changes. The interactive tools are built on verified figures from named documents.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'This publication is written for practitioners who already work in or around restructurings — it assumes the vocabulary and leads with mechanism, not with first-principles explanation.';
  IF n <> 1 THEN RAISE EXCEPTION '667: oufe.com rank 5: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'This publication is written for practitioners who already work in or around restructurings — it assumes the vocabulary and leads with mechanism, not with first-principles explanation.' THEN jsonb_set(e, '{point}', to_jsonb('This publication is written for practitioners who already work in or around restructurings; it assumes the vocabulary and leads with mechanism.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'The analysis is focused specifically on the UK statutory restructuring toolkit — Part 26A restructuring plans, cross-class cramdown, schemes of arrangement, and the special administration regime — rather than generic global insolvency coverage.';
  IF n <> 1 THEN RAISE EXCEPTION '667: oufe.com rank 6: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'The analysis is focused specifically on the UK statutory restructuring toolkit — Part 26A restructuring plans, cross-class cramdown, schemes of arrangement, and the special administration regime — rather than generic global insolvency coverage.' THEN jsonb_set(e, '{point}', to_jsonb('The analysis is focused specifically on the UK statutory restructuring toolkit: Part 26A restructuring plans, cross-class cramdown, schemes of arrangement, and the special administration regime.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Every figure you encounter here traces to a named, dated document you can check yourself — where we cannot verify something, we say so plainly.';
  IF n <> 1 THEN RAISE EXCEPTION '667: oufe.com rank 1: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Every figure you encounter here traces to a named, dated document you can check yourself — where we cannot verify something, we say so plainly.' THEN jsonb_set(e, '{point}', to_jsonb('Every figure you encounter here traces to a named, dated document you can check yourself; where we cannot verify something, we say so.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='oufe.com';

  -- ── robot-hands.com ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='robot-hands.com';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for robot-hands.com'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'You leave with a ranked shortlist of gripper candidates demonstrably matched to your specific payload, environmental rating, cycle time, and interface requirements — produced against auditable scoring criteria, not a vendor recommendation.';
  IF n <> 1 THEN RAISE EXCEPTION '667: robot-hands.com rank 1: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'You leave with a ranked shortlist of gripper candidates demonstrably matched to your specific payload, environmental rating, cycle time, and interface requirements — produced against auditable scoring criteria, not a vendor recommendation.' THEN jsonb_set(e, '{point}', to_jsonb('You leave with a ranked shortlist of gripper candidates matched to your specific payload, environmental rating, cycle time, and interface requirements, produced against auditable scoring criteria.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'No existing platform combines a cross-technology gripper catalog benchmarked on consistent criteria, a structured and auditable selection methodology, and interactive calculation tools in one workflow — so you consult this before going to a manufacturer, not instead of one.';
  IF n <> 1 THEN RAISE EXCEPTION '667: robot-hands.com rank 2: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'No existing platform combines a cross-technology gripper catalog benchmarked on consistent criteria, a structured and auditable selection methodology, and interactive calculation tools in one workflow — so you consult this before going to a manufacturer, not instead of one.' THEN jsonb_set(e, '{point}', to_jsonb('This platform combines a cross-technology gripper catalog benchmarked on consistent criteria, a structured and auditable selection methodology, and interactive calculation tools in one workflow, so you consult it before going to a manufacturer.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Before you act on any MatchMatrix output, the scoring methodology is documented and auditable, the catalog data cites real sources with verification dates, and trade-offs are acknowledged rather than hidden.';
  IF n <> 1 THEN RAISE EXCEPTION '667: robot-hands.com rank 4: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Before you act on any MatchMatrix output, the scoring methodology is documented and auditable, the catalog data cites real sources with verification dates, and trade-offs are acknowledged rather than hidden.' THEN jsonb_set(e, '{point}', to_jsonb('Before you act on any MatchMatrix output, the scoring methodology is documented and auditable, the catalog data cites real sources with verification dates, and trade-offs are acknowledged.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='robot-hands.com';

  -- ── vonc.com ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='vonc.com';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for vonc.com'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Your Archetype — emerging from how you actually argue, not from how you describe yourself — is a shareable identity you earn inside the arena.';
  IF n <> 1 THEN RAISE EXCEPTION '667: vonc.com rank 5: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Your Archetype — emerging from how you actually argue, not from how you describe yourself — is a shareable identity you earn inside the arena.' THEN jsonb_set(e, '{point}', to_jsonb('Your Archetype, emerging from how you actually argue, is a shareable identity you earn inside the arena.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='vonc.com';

  -- ── webdesign.co.uk ──
  SELECT sp.data->'lead_with' INTO lw FROM site_specs sp JOIN sites s ON s.id=sp.site_id
    WHERE sp.aspect='offer_ordering' AND sp.is_current AND s.domain='webdesign.co.uk';
  IF lw IS NULL THEN RAISE EXCEPTION '667: no offer_ordering lead_with for webdesign.co.uk'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(lw) e WHERE e->>'point' = 'Every tool is paired with a plain-English article explaining the concept behind it, so you leave knowing why the output is what it is, not just what to copy.';
  IF n <> 1 THEN RAISE EXCEPTION '667: webdesign.co.uk rank 4: FROM text matches % elements, want exactly 1 - rows have drifted since the ACK, do not apply blind', n; END IF;
  SELECT jsonb_agg(CASE WHEN e->>'point' = 'Every tool is paired with a plain-English article explaining the concept behind it, so you leave knowing why the output is what it is, not just what to copy.' THEN jsonb_set(e, '{point}', to_jsonb('Every tool is paired with an article explaining the concept behind it, so you leave knowing why the output is what it is.'::text)) ELSE e END ORDER BY idx)
    INTO lw FROM jsonb_array_elements(lw) WITH ORDINALITY t(e, idx);
  UPDATE site_specs sp SET data = jsonb_set(sp.data, '{lead_with}', lw)
    FROM sites s WHERE s.id=sp.site_id AND sp.aspect='offer_ordering' AND sp.is_current AND s.domain='webdesign.co.uk';

  RAISE NOTICE '667 OK: 41 points washed across 20 sites; ranks untouched; the 10 exclusions untouched.';
END $mig$;

COMMIT;
