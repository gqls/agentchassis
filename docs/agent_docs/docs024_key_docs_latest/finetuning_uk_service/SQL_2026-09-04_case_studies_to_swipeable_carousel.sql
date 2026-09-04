-- SQL_2026-09-04_case_studies_to_swipeable_carousel.sql — finetuning.uk /index.html: the CARD CANARY.
-- Owner, 2026-09-03 23:45 BST: "homepage cards, case studies can be swipeable. For any number greater than 3."
-- Critic (design_critique_run 204f1ff7): the four-or-five case-study cards resolve to 3+1/3+2, an orphan.
--
-- WHAT THIS DOES, in one transaction:
--   1. page_components row c00de077-163c-445d-be84-7ed2e18219d4 (slot case-studies-grid, component case-studies-grid
--      3f946437-1dc7-4164-987d-620933589076, rendered md5 47033dd1781d43c2066c8778d2d59cbb, 17670 B)
--      -> component swipeable-insight-carousel (cbd81d06, function swipeable-insight-carousel,
--      render_mode agent, template-rendered from content_data by rerender_page_sections — no LLM),
--      slot_name = the function, content_data = the five cards mapped VERBATIM by a script
--      (label=category_label, headline=title, body=excerpt, link_url, link_label=read_case_study_label,
--      attribution=client_name; section_title=section_headline; section_eyebrow=eyebrow_label).
--   2. pages.sections: the string "case-studies-grid" -> "swipeable-insight-carousel" (the assembler
--      joins slots by name). site_plan_sections: NONE for this site (0 site_plans rows, measured
--      2026-09-04) — so two placement rows, not three.
--   3. a page_rerender item with spec.reason='section_data_resolved' so page-rerender takes the
--      rerender_sections branch (re-renders every section of the page from STORED content_data through
--      each component's template; all six other components are render_mode template) and deploys.
--
-- WHAT IS LOST BY CONSTRUCTION (the carousel's contract has no field for them; recorded, told the owner):
--   the section intro sentence, the section's own CTA (label/subtext/link — the page keeps its
--   call-to-action section), and the five card images (they stay on /case-studies.html).
--
-- ACCEPTANCE at the SERVED page: five cards; every headline/body/label/attribution byte-identical to
--   the mapped JSON; the orphan gone (a scroll-snap track, not a 3-column grid); the five other
--   sections' rendered_html md5 unchanged from the 2026-09-04 snapshot; CDP: the track scrolls.
-- ROLLBACK: the archived page_components row (archive trigger) or the snapshot file
--   case_studies_slot_before.json (content_data + component_id + slot_name), restore both placement
--   rows, then the same rerender.
\set ON_ERROR_STOP on
BEGIN;
DO $$
DECLARE
  v_site uuid := '1368e337-dd1d-4799-bbb3-8221a1b79bcc';
  v_page uuid := 'a716cacc-eec2-4aa6-a08b-7e6732506f41';
  v_row  uuid := 'c00de077-163c-445d-be84-7ed2e18219d4';
  v_old_comp uuid := '3f946437-1dc7-4164-987d-620933589076';
  v_new_comp uuid := 'cbd81d06-429e-4fce-80e2-0cd2c5fc5d63';
  v_data jsonb := $json${"section_eyebrow": "Case Studies", "section_title": "Real problems, systems that keep running", "cards": [{"label": "Automation", "headline": "Cutting the quote turnaround bottleneck", "body": "A logistics operator was losing bids because quotes took days to prepare. We built an automation pipeline that reads incoming requests and drafts a quote for review. Turnaround times dropped substantially, and the sales team stopped losing deals to slower competitors.", "link_url": "/case-studies.html", "link_label": "Read case study", "attribution": "A logistics operator losing bids to slow quoting"}, {"label": "Knowledge Base", "headline": "Answering the same client question, automatically", "body": "A professional services firm had staff answering the same procedural questions by email, day after day. We built a retrieval system (RAG) over their internal documents, so a request now finds an answer immediately. The team spends less time chasing information and more on client work.", "link_url": "/case-studies.html", "link_label": "Read case study", "attribution": "A professional services firm managing dispersed documentation"}, {"label": "Model Training", "headline": "Training a model that speaks the client's language", "body": "A financial services team needed a model that understood their own filings. We trained a specialised model on their documents using LoRA, a technique that adapts an existing model efficiently. The model answers the way their own analysts would, and it belongs to them.", "link_url": "/case-studies.html", "link_label": "Read case study", "attribution": "A financial services team reviewing regulatory filings"}, {"label": "Data Pipelines", "headline": "Turning company registry filings into structured data", "body": "A market research firm needed structured data pulled from thousands of company filings, updated regularly. We built a data pipeline that collects and organises this information automatically from public registries. What used to take a research team days now runs unattended overnight.", "link_url": "/case-studies.html", "link_label": "Read case study", "attribution": "A market research firm tracking company filings"}, {"label": "Agent Systems", "headline": "Coordinating maintenance requests across multiple sites", "body": "A facilities management contractor was juggling maintenance requests across many sites with no shared view. We built a network of AI agents that triage, route, and track each request, with human review built into key steps. Little falls through the cracks between sites now.", "link_url": "/case-studies.html", "link_label": "Read case study", "attribution": "A facilities contractor covering multiple sites"}]}$json$::jsonb;
  n int;
BEGIN
  -- pre-flight: the row is what this file was written against
  SELECT count(*) INTO n FROM page_components WHERE id=v_row AND page_id=v_page AND component_id=v_old_comp AND slot_name='case-studies-grid'
    AND md5(rendered_html)='47033dd1781d43c2066c8778d2d59cbb' AND lock_type IS NULL;
  IF n<>1 THEN RAISE EXCEPTION 'pre-flight: the case-studies-grid row is not as snapshotted (%)', n; END IF;
  SELECT count(*) INTO n FROM content_components WHERE id=v_new_comp AND name='swipeable-insight-carousel' AND is_active AND function='swipeable-insight-carousel';
  IF n<>1 THEN RAISE EXCEPTION 'pre-flight: carousel component missing/inactive'; END IF;
  SELECT count(*) INTO n FROM pages WHERE id=v_page AND site_id=v_site AND url='/index.html' AND status='active' AND sections @> '["case-studies-grid"]'::jsonb;
  IF n<>1 THEN RAISE EXCEPTION 'pre-flight: pages.sections does not carry case-studies-grid'; END IF;
  SELECT count(*) INTO n FROM site_work_items WHERE page_id=v_page AND status IN ('triaged','claimed','in_progress') AND item_type IN ('page_rerender','needs_content_page','content_rewrite');
  IF n<>0 THEN RAISE EXCEPTION 'pre-flight: % open build/rerender item(s) on the homepage — wait', n; END IF;
  IF jsonb_array_length(v_data->'cards') <> 5 THEN RAISE EXCEPTION 'mapping: expected 5 cards'; END IF;
  IF v_data::text LIKE '%—%' THEN RAISE EXCEPTION 'mapping: em dash'; END IF;

  UPDATE page_components SET component_id=v_new_comp, slot_name='swipeable-insight-carousel', content_data=v_data, updated_at=NOW() WHERE id=v_row;
  UPDATE pages SET sections = (SELECT jsonb_agg(CASE WHEN e = '"case-studies-grid"'::jsonb THEN '"swipeable-insight-carousel"'::jsonb ELSE e END) FROM jsonb_array_elements(sections) e), updated_at=NOW() WHERE id=v_page;

  INSERT INTO site_work_items (site_id, page_id, source, pipeline, item_type, severity, summary, priority, handler_agent, status, created_by, spec, item_key, batch_id)
  SELECT v_site, v_page, w.source, w.pipeline, 'page_rerender', w.severity,
         'index: case-studies-grid -> swipeable-insight-carousel (owner 2026-09-03: swipeable for >3 cards); section rerender from stored content_data',
         w.priority, 'page-rerender', 'triaged', 'finetuning_uk_service_lane',
         jsonb_build_object('reason','section_data_resolved','domain','finetuning.uk','page_id',v_page::text,'filename','index.html','page_name','index'),
         'page_rerender_index_' || v_site::text || '_carousel_canary_' || to_char(now(),'YYYYMMDDHH24MI'), gen_random_uuid()
    FROM site_work_items w WHERE w.id='50c2a394-3cee-48af-8f0f-46d1f0a8d834';
  IF NOT FOUND THEN RAISE EXCEPTION 'template rerender item 50c2a394 not found'; END IF;

  SELECT count(*) INTO n FROM page_components WHERE id=v_row AND component_id=v_new_comp AND slot_name='swipeable-insight-carousel' AND jsonb_array_length(content_data->'cards')=5;
  IF n<>1 THEN RAISE EXCEPTION 'post: slot not repointed'; END IF;
  SELECT count(*) INTO n FROM pages WHERE id=v_page AND sections @> '["swipeable-insight-carousel"]'::jsonb AND NOT sections @> '["case-studies-grid"]'::jsonb AND jsonb_array_length(sections)=6;
  IF n<>1 THEN RAISE EXCEPTION 'post: pages.sections not updated (expect 6 entries, carousel in, grid out)'; END IF;
  SELECT count(*) INTO n FROM site_work_items WHERE page_id=v_page AND item_type='page_rerender' AND status='triaged' AND spec->>'reason'='section_data_resolved';
  IF n<>1 THEN RAISE EXCEPTION 'post: rerender item not queued (%)', n; END IF;
  RAISE NOTICE 'canary: slot repointed to the carousel with 5 verbatim cards; pages.sections updated; section rerender queued';
END $$;
COMMIT;
