BEGIN;

INSERT INTO pages (site_id, name, url, title, page_type, status, build_status,
                   nav_label, nav_order, in_header, in_footer, content_direction)
SELECT s.id, 'playground', '/playground.html',
       'The playground: an hour with your own model',
       'content', 'active', 'planned',
       'Playground', 8, false, true,
       jsonb_build_object('required_links', jsonb_build_array('/contact.html', '/your-own-model.html'))
  FROM sites s WHERE s.domain='finetuning.uk'
   AND NOT EXISTS (SELECT 1 FROM pages p2 WHERE p2.site_id=s.id AND p2.name='playground');

INSERT INTO site_work_items (site_id, page_id, source, pipeline, item_type, severity, summary,
                             priority, handler_agent, status, created_by, spec, item_key, batch_id)
SELECT s.id, p.id, 'gap_plan', 'build', 'needs_content_page', 'medium',
       'Build the playground page — the included hour and how to book it (owner instruction 2026-09-02)',
       40, 'page-build-handler', 'triaged', 'finetuning_uk_service_lane',
       '{"check": "gap_plan_new_page", "page_name": "playground", "page_url": "/playground.html", "suggestion": "Build the playground page for finetuning.uk. It explains the hour that comes with the \u00a399 fine-tuned model, and how to book it.\n\nSIX SECTIONS, one subject each. Do not merge them and do not add more.\n\n1. hero \u2014 What the playground is: an hour with your own trained model, already loaded, so you can try it on your real work before you rely on it.\n2. generic-text-block \u2014 What you actually do in that hour: put your own tasks to the model and see how it answers. Bring the work you would normally do yourself.\n3. generic-text-block \u2014 When you can book it: you pick the time, 9am to 5pm UK time, Monday to Friday. Outside those hours can usually be arranged. The included hour is to be used within 30 days of handover; more hours can be bought.\n4. generic-text-block \u2014 What to have ready before the hour, so none of it is spent setting up: a few real examples of the work, and a clear idea of what a good answer looks like to you.\n5. faq \u2014 Questions a buyer actually asks about the hour. Ground every answer in a registered fact or leave the question out.\n6. call-to-action \u2014 How to book: get in touch and say which times suit you.\n\nFACTS. Use only what is registered in evidence_base: the \u00a399 price (ft-price-99), the included hour and its 30-day expiry (ft-playground-hour), the 9am-5pm UK weekday booking window with other times by arrangement (ft-booking-hours). Do not invent a price, a duration, a turnaround, a number of sessions, or any statistic.\n\nHONESTY, and this one matters more than the copy. There is NO self-serve calendar. Booking happens by getting in touch and agreeing a time. Do not write anything that implies instant booking, a live availability grid, or automatic confirmation. Do not promise a named person will attend. If you cannot say something from a registered fact, say less.\n\nREGISTER. Follow the site''s writing rules in content_direction. In particular: when a sentence sets up a comparison, write the first half and stop; leave out the \"not\", the \"instead of\", the \"rather than\". There is no hidden competition. State what a thing IS. No em-dash asides. British English. Plain words, short sentences, the fact before the flourish.\n\nLINKS. Link to /contact.html from the body of the call-to-action, and to /your-own-model.html where the \u00a399 offer is first mentioned.", "sections": ["hero", "generic-text-block", "generic-text-block", "generic-text-block", "faq", "call-to-action"], "reason": "owner instruction 2026-09-02: build the playground booking page (decision 4 of the 2026-08-26 set)"}'::jsonb,
       'gap_plan_new_playground_' || s.id::text, gen_random_uuid()
  FROM sites s JOIN pages p ON p.site_id=s.id AND p.name='playground'
 WHERE s.domain='finetuning.uk'
   AND NOT EXISTS (SELECT 1 FROM site_work_items w WHERE w.site_id=s.id
                    AND w.item_key='gap_plan_new_playground_' || s.id::text
                    AND w.status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled'))
ON CONFLICT DO NOTHING;

DO $$
DECLARE n_page int; n_item int; sects int;
BEGIN
  SELECT count(*) INTO n_page FROM pages p JOIN sites s ON s.id=p.site_id
   WHERE s.domain='finetuning.uk' AND p.name='playground' AND p.build_status='planned';
  IF n_page <> 1 THEN RAISE EXCEPTION 'playground page rows = %, want 1 planned', n_page; END IF;

  SELECT count(*), max(jsonb_array_length(w.spec->'sections')) INTO n_item, sects
    FROM site_work_items w JOIN sites s ON s.id=w.site_id
   WHERE s.domain='finetuning.uk' AND w.item_type='needs_content_page'
     AND w.spec->>'page_name'='playground' AND w.status='triaged';
  IF n_item <> 1 THEN RAISE EXCEPTION 'needs_content_page items = %, want 1 triaged', n_item; END IF;
  IF sects <> 6 THEN RAISE EXCEPTION 'sections = %, want 6 (the brief names one subject per section)', sects; END IF;

  RAISE NOTICE 'playground page PLANNED and dispatched: 1 page row, 1 triaged needs_content_page, 6 sections briefed';
END $$;

COMMIT;
