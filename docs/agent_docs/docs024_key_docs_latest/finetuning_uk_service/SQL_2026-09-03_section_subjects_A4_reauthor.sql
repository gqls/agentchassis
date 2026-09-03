-- SQL_2026-09-03_section_subjects_A4_reauthor.sql — finetuning.uk: re-author the three
-- hand-authored `pages.section_subjects` arrays to the OPTION A4 register (owner decision
-- 2026-09-03, relayed by the prompts lane; spec: CONTRIB_2026-09-03b_…_subject_phrasing_spec.md,
-- original framework_prompts_positive_voice/SPEC_2026-09-03_section_subject_phrasing.md).
--
-- Under A4 a subject is THE LINE THE SECTION OPENS ON, written to the reader in the site's
-- voice; it is printed verbatim as its section's first line and listed in every sibling's
-- prompt. So: one sentence each; distinct within a page; no em dashes; nothing you would not
-- publish; no numbers (facts come from the Verified Facts block, a subject is not a route
-- around it). Section 4 of the playground is the owner's own exemplar, verbatim.
--
-- Order is index-aligned with the page's sections (hero, gtb, gtb, gtb, faq, cta on all three),
-- exactly as the 2026-09-03 Stage A backfill was. Data only: live on apply, read at the next
-- page-build-handler rebuild (Stage B), which is held on 641.
--
-- PREVIOUS VALUES (for rollback by hand):
--   /playground.html        ["what the playground is","what you actually do in the hour","when you can book it","what to have ready before the hour","what people ask about the hour","how to book"]
--   /your-own-model.html    ["what the offer is","how it works","what you get, exactly","how £99 can be enough","what the words mean","how to book a discovery call"]
--   /technical-details.html ["exactly what the £99 fine-tune contains","which model it is and what its licence allows","what file you receive and where it runs","how the training works and what we handle","what a technical reader asks","where to go next"]
--
-- Rehearse with COMMIT -> ROLLBACK first. Run:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < this_file
\set ON_ERROR_STOP on
BEGIN;
DO $$
DECLARE
  v_site uuid := '1368e337-dd1d-4799-bbb3-8221a1b79bcc';
  r record; n int; s text; arr jsonb;
  new_pg  jsonb := '["The playground is an hour with your own trained model, already loaded and ready for your real work.",
                     "In the hour itself, you put your everyday tasks to the model and see how it answers.",
                     "You choose when your hour happens, and we agree the time with you when you get in touch.",
                     "If you''d like to prepare in advance of your hour, you might want to get these things ready.",
                     "These are the questions people ask about the hour before they book.",
                     "To book your hour, get in touch and tell us which times suit you."]'::jsonb;
  new_yom jsonb := '["Your company''s voice, in a model you own.",
                     "Here is how it works, from the examples you send to the hour with your model.",
                     "This is exactly what you get, and what each part means.",
                     "This is what makes the price possible.",
                     "A short glossary of the words on this page, in plain English.",
                     "If you''d like to talk it through first, book a discovery call."]'::jsonb;
  new_td  jsonb := '["This page sets out exactly what the fine-tune contains, for the person who checks these things.",
                     "Before training starts, you''ll know which model we are using and what its licence allows.",
                     "You receive one file, and it runs on ordinary hardware, wherever you choose to keep it.",
                     "Here is how the training works, step by step, and which parts we handle for you.",
                     "The questions a technical reader usually asks before signing off.",
                     "From here you can read the offer, look at the playground, or book a call."]'::jsonb;
BEGIN
  -- Pre-flight: the three pages carry the arrays this file was written against (guards a concurrent edit).
  SELECT count(*) INTO n FROM pages WHERE site_id=v_site AND url='/playground.html'
    AND section_subjects->>3 = 'what to have ready before the hour' AND jsonb_array_length(section_subjects)=6;
  IF n<>1 THEN RAISE EXCEPTION 'pre-flight: /playground.html subjects are not the 2026-09-03 Stage A backfill'; END IF;
  SELECT count(*) INTO n FROM pages WHERE site_id=v_site AND url='/your-own-model.html'
    AND section_subjects->>3 = 'how £99 can be enough' AND jsonb_array_length(section_subjects)=6;
  IF n<>1 THEN RAISE EXCEPTION 'pre-flight: /your-own-model.html subjects are not the 2026-09-03 Stage A backfill'; END IF;
  SELECT count(*) INTO n FROM pages WHERE site_id=v_site AND url='/technical-details.html'
    AND section_subjects->>1 = 'which model it is and what its licence allows' AND jsonb_array_length(section_subjects)=6;
  IF n<>1 THEN RAISE EXCEPTION 'pre-flight: /technical-details.html subjects are not the 2026-09-03 Stage A backfill'; END IF;

  -- Spec checks on the NEW arrays: 6 each, distinct within page, no em dash, no digit, <=120 chars, all end in . or a phrase.
  FOREACH arr IN ARRAY ARRAY[new_pg, new_yom, new_td] LOOP
    IF jsonb_array_length(arr)<>6 THEN RAISE EXCEPTION 'spec: array length %', jsonb_array_length(arr); END IF;
    SELECT count(DISTINCT x) INTO n FROM jsonb_array_elements_text(arr) x;
    IF n<>6 THEN RAISE EXCEPTION 'spec rule 3: duplicate subject within a page'; END IF;
    FOR s IN SELECT x FROM jsonb_array_elements_text(arr) x LOOP
      IF s LIKE '%—%' OR s LIKE '%–%' THEN RAISE EXCEPTION 'spec rule 4: dash in "%"', s; END IF;
      IF s ~ '[0-9£$%]' THEN RAISE EXCEPTION 'no numbers in a subject: "%"', s; END IF;
      IF length(s) > 120 THEN RAISE EXCEPTION 'spec rule 2: too long (% chars): "%"', length(s), s; END IF;
      IF s !~ '^[A-Z]' THEN RAISE EXCEPTION 'spec rule 6: sentence case: "%"', s; END IF;
    END LOOP;
  END LOOP;

  UPDATE pages SET section_subjects=new_pg,  updated_at=NOW() WHERE site_id=v_site AND url='/playground.html';
  UPDATE pages SET section_subjects=new_yom, updated_at=NOW() WHERE site_id=v_site AND url='/your-own-model.html';
  UPDATE pages SET section_subjects=new_td,  updated_at=NOW() WHERE site_id=v_site AND url='/technical-details.html';

  SELECT count(*) INTO n FROM pages WHERE site_id=v_site AND url IN ('/playground.html','/your-own-model.html','/technical-details.html')
    AND jsonb_array_length(section_subjects)=6 AND section_subjects->>0 ~ '^[A-Z]';
  IF n<>3 THEN RAISE EXCEPTION 'post: expected 3 pages re-authored, found %', n; END IF;
  RAISE NOTICE 'section_subjects re-authored to A4 on 3 pages';
END $$;
COMMIT;
