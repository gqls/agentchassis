-- technical_details_stage_b_dispatch.sql — finetuning.uk /technical-details.html:
-- the 443 STAGE B rebuild, with the brief REWRITTEN to the owner's 2026-09-03 verdict.
--
-- WHY A NEW BRIEF, not the RUNBOOK's copy-the-original recipe: the original
-- gap_plan item (896bb245, 2026-08-24 brief) asked in its §2 for "the three
-- families we work with and their licences" listed one by one — and the served
-- page is exactly that (h3s "Llama models / Mistral 7B / Phi models / Three
-- families, three licences"). The owner, 2026-09-03 ~10:00Z, verbatim:
-- "an unhelpful page listing on 3 types of model". The writer did what it was
-- told; the brief is what changes. Everything else in the original brief
-- (register, numbers discipline, licence facts only as registered, six
-- sections one subject each) is kept word for word.
--
-- WHAT §2 SAYS NOW: the model is a small open-weight one, chosen for the job
-- and NAMED — with its licence — in writing before training starts; we only
-- use models whose licence lets a business this size use the result
-- commercially at no charge; exact terms are stated at handover, version-pinned.
-- No family listing, no per-family terms on the page. The three licence facts
-- stay registered (evidence_base ft-licence-*) and the writer_block still
-- forbids stating any licence term not recorded there.
--
-- Also folded in (owner DIRECTION 2026-09-03: "the site is the tool"):
-- /playground.html joins required_links; the FAQ's playground answer points at it.
--
-- HELD — DO NOT RUN until BOTH gates pass; the DO block below ASSERTS them:
--   G1. migration 641 (option A) is applied to the live page-content-writer —
--       the template names .current_section.subject. Without it Stage B
--       rebuilds the page with the same repeated h2s (443 Stage A proved the
--       subject reaches the writer's DATA only). As of 2026-09-03 17:00Z: NOT applied.
--   G2. no open needs_content_page on the page (a second one is not harmless —
--       RUNBOOK "Rebuild an ALREADY-DEPLOYED page").
-- Then: dispatch >=300 s after any chassis pod start; watch the item; assert
-- distinct h2s on the served page (the Stage B acceptance) with a tier-1 page
-- as control; check the </strom> of bugs_open/456 (malformed_closing_tag slug)
-- is gone from the served bytes.
--
-- Run:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--         psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < this_file
-- Rehearse first with COMMIT swapped for ROLLBACK.
-- REHEARSAL RECORD 2026-09-03 ~17:05Z: run under ROLLBACK against the live DB it
-- stopped at G1 as intended (641 absent). The path BEYOND G1 is UNREHEARSED as a
-- transaction (the session's permission classifier refused the stubbed-G1 run);
-- what WAS proven read-only: the jsonb spec expression evaluates (6 sections, no
-- mode, no not_dispatchable, names /playground.html, 3,221 chars) and every
-- INSERT column exists. The first cut of the family-listing post-condition used
-- '%Llama%', which matches 'Ollama'/'llama.cpp' and would have REFUSED a correct
-- brief — found by that read-only check, fixed below. BEFORE the real run:
-- stub G1's LIKE to '%process_sections_loop%', swap COMMIT->ROLLBACK, run, expect
-- the NOTICE, then confirm 0 triaged items on the page and build_status unchanged.

\set ON_ERROR_STOP on
BEGIN;

DO $$
DECLARE
  v_site uuid := '1368e337-dd1d-4799-bbb3-8221a1b79bcc';   -- finetuning.uk
  v_page uuid := 'a32b8822-db49-4e45-88f8-bda06d73de62';   -- /technical-details.html
  v_orig uuid := '896bb245-3233-4563-b84b-b052ab19d461';   -- the 2026-09-03 Stage A item (row fields copied, spec NOT)
  n int;
BEGIN
  -- G1: 641 live?
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='page-content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config::text LIKE '%current_section.subject%';
  IF n <> 1 THEN
    RAISE EXCEPTION 'G1 FAILED: migration 641 is not on the live page-content-writer (rows naming current_section.subject = %). Stage B before 641 reproduces the repeated h2s. Do not dispatch.', n;
  END IF;

  -- G2: page state + no open needs_content_page
  SELECT count(*) INTO n FROM pages WHERE id=v_page AND site_id=v_site AND build_status='deployed' AND status='active';
  IF n <> 1 THEN RAISE EXCEPTION 'G2 FAILED: technical-details is not a deployed active page (%).', n; END IF;
  SELECT count(*) INTO n FROM site_work_items
   WHERE page_id=v_page AND item_type='needs_content_page'
     AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled','deferred');
  IF n <> 0 THEN RAISE EXCEPTION 'G2 FAILED: % open needs_content_page item(s) on the page already.', n; END IF;

  -- required_links: add the playground (idempotent)
  UPDATE pages
     SET content_direction = jsonb_set(COALESCE(content_direction,'{}'::jsonb), '{required_links}',
           (SELECT jsonb_agg(DISTINCT x) FROM jsonb_array_elements_text(
              COALESCE(content_direction->'required_links','[]'::jsonb) || '["/playground.html"]'::jsonb) x)),
         build_status='planned', updated_at=NOW()
   WHERE id=v_page;

  INSERT INTO site_work_items (site_id, page_id, source, pipeline, item_type, severity, summary,
                               priority, handler_agent, status, created_by, spec, item_key, batch_id)
  SELECT w.site_id, w.page_id, w.source, w.pipeline, w.item_type, w.severity,
         'technical-details: 443 Stage B rebuild (641 live) with the brief rewritten to the owner''s 2026-09-03 verdict — no three-model listing',
         40, 'page-build-handler', 'triaged', 'finetuning_uk_service_lane',
         jsonb_build_object(
           'title', 'The Technical Details | FineTuning',
           'page_name', 'technical-details',
           'page_type', 'content',
           'source', 'owner-request',
           'reason', 'bugs_open/443 Stage B, 2026-09-03: first rebuild with 641 (section subject in the PROMPT); brief rewritten after the owner called the three-model listing unhelpful',
           'purpose', 'The honest technical page behind the £99 offer: which model and what its licence allows, the file, where it runs, what we handle.',
           'sections', jsonb_build_array('hero','generic-text-block','generic-text-block','generic-text-block','faq','call-to-action'),
           'suggestion',
             'This is finetuning.uk''s honest technical page, sitting behind the £99 offer page (/your-own-model.html): written for the one semi-technical person at a small business who signs the purchase off. They want one page that says exactly what they get. Register: plain, specific, calm; every technical term explained in the same breath; state every fact in its positive form (define-by-negation is banned in all its grammatical forms); no em-dash asides; no hype. Authority comes from specificity here. Six sections, one subject each, and no two sections may open on the same claim:' || E'\n' ||
             '(1) hero: one plain statement of what this page is: the exact contents of the £99 fine-tune, written down for the person who checks these things, with an invitation to read on.' || E'\n' ||
             '(2) generic-text-block, subject WHICH MODEL IT IS AND WHAT ITS LICENCE ALLOWS: the base model is a small open-weight model (weights you can download and run yourself), chosen to suit the job; before training starts we tell you in writing which model it is and which licence it comes under; we only work with models whose licence lets a business of your size use the trained result commercially at no charge; the exact terms of your model''s licence are stated at handover, for the exact version used. Do NOT list model families or vendors on this page and do NOT state any licence''s terms here; the reader learns the name and terms for their own model, in writing, before we start.' || E'\n' ||
             '(3) generic-text-block, subject THE FILE AND WHERE IT RUNS: you receive the trained model as a GGUF file (a standard format for AI models); it runs on ordinary hardware with widely available free software such as Ollama or llama.cpp; a laptop handles the small models we train; it runs offline once downloaded, and where it lives after that is entirely the customer''s choice.' || E'\n' ||
             '(4) generic-text-block, subject HOW THE TRAINING WORKS AND WHAT WE HANDLE: fine-tuning done efficiently with LoRA (adjusting a small set of added weights so the model picks up your voice while keeping its general ability); you send examples of how your business writes; we prepare the data, run the training overnight on a rented GPU, convert the result to GGUF, and hand the file over; your documents are used to train your model and for nothing beyond it.' || E'\n' ||
             '(5) faq, subject THE QUESTIONS A TECHNICAL READER ASKS: does running it need a GPU (answer: training uses one, running the small models does not; a modern laptop is fine); which operating systems (anywhere Ollama or llama.cpp runs: Windows, macOS, Linux); can it go on our website later (yes, as a follow-on service); what is in the playground hour (a booked session chatting with your trained model with us on hand; the playground page at /playground.html says how it works).' || E'\n' ||
             '(6) call-to-action: low-pressure: read the offer at /your-own-model.html, see the playground at /playground.html, or book a discovery call at /contact.html.' || E'\n' ||
             'Numbers discipline: state only the registered facts: the £99 price if needed. Licence terms appear on this page only as the rule in (2), never as a specific licence''s terms. No performance figures, no percentages, no time-saved claims. Do not make any promise about data retention periods; say only what is registered.'
         ),
         w.item_key, gen_random_uuid()
    FROM site_work_items w WHERE w.id=v_orig;
  IF NOT FOUND THEN RAISE EXCEPTION 'original item % not found', v_orig; END IF;

  -- Post-conditions the RUNBOOK names (landmine: an audit VERDICT copied as a brief; a spec with mode)
  SELECT count(*) INTO n FROM site_work_items w
   WHERE w.page_id=v_page AND w.status='triaged' AND w.item_type='needs_content_page'
     AND NOT (w.spec ? 'not_dispatchable') AND NOT (w.spec ? 'mode') AND (w.spec ? 'sections')
     AND jsonb_array_length(w.spec->'sections') = 6
     -- the family-listing shape the owner rejected (NB: '%Llama%' would match 'Ollama' and
     -- 'llama.cpp', which the brief legitimately names — caught in the 2026-09-03 read-only rehearsal)
     AND w.spec->>'suggestion' NOT ILIKE '%Mistral%' AND w.spec->>'suggestion' NOT ILIKE '%Llama Community%'
     AND w.spec->>'suggestion' NOT ILIKE '%Llama models%' AND w.spec->>'suggestion' NOT ILIKE '%Phi models%'
     AND w.spec->>'suggestion' NOT ILIKE '%Apache 2.0%' AND w.spec->>'suggestion' NOT ILIKE '%under MIT%';
  IF n <> 1 THEN RAISE EXCEPTION 'post-condition FAILED: expected exactly 1 clean triaged brief, found %', n; END IF;
  SELECT count(*) INTO n FROM pages WHERE id=v_page AND build_status='planned'
     AND jsonb_array_length(section_subjects) = 6
     AND content_direction->'required_links' ? '/playground.html';
  IF n <> 1 THEN RAISE EXCEPTION 'post-condition FAILED: page not planned / subjects<>6 / playground link missing'; END IF;
  RAISE NOTICE 'Stage B item inserted for technical-details; page planned; required_links now carry /playground.html';
END $$;

COMMIT;
