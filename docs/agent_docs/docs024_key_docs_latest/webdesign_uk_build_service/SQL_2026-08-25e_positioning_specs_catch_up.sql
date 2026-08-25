-- SQL_2026-08-25e — the POSITIONING SPECS catch up with the register (owner brief 2026-08-25).
--
-- WHY. The 2nd index rebuild regenerated an unlocked hero that STILL read "A complete
-- website for your business": the writer takes what to SAY from the site's spec aspects
-- (content_direction, identity, mission_brief, briefing, strategy, submission) and the
-- page record's title/meta, and takes only HOW to say it from writer_block. Those aspects
-- still described a non-technical business owner ("They are not technical"; "hosting and
-- DNS are the studio's problem"), listed a RETIRED service ("Post-acceptance changes"),
-- carried the RETIRED + BANNED figure "usually ready the next day" (identity USPs,
-- briefing about_us + services, strategy value_proposition), told CTAs to say "Call us"
-- against no_presales_service, and said "transfer freely" (scrubbed from writer_block
-- 2026-08-21, still here). One writer_block line said "written for a business owner who
-- is not technical". None of that is reachable from evidence_base edits.
--
-- METHOD. content_direction is edited as TEXT (replace chain over data::text, then
-- ::jsonb) so its structured fields and its rendered `formatted` duplicate move together
-- (LANDMINES: the `formatted` copy stays stale and authoritative when only the structured
-- field is edited). Every needle is verbatim from the live row and guarded: present N
-- times before, absent after, replacement present. No em dashes are introduced; the word
-- "honest" (banned in copy; prompt-example trap) is removed from strategy.content_strategy.
-- Page titles/metas are page-record fields a rerender does not regenerate: set here.
-- roadmap_brief (a phase log) is deliberately untouched.

BEGIN;

-- ── 1. content_direction: text-level replace so structured + formatted move together ──
WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='content_direction' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(
      c.data::text,
      $n$Talks to a busy business owner the way a competent tradesperson would$n$,
      $n$Talks to an experienced web designer, or anyone comfortable running a static site themselves, the way one tradesperson talks to another$n$),
      $n$Get in touch. Send us an email. Call us. Ask a question. These are the register.$n$,
      $n$Start your website brief. Ask the chat box. Read the full terms. These are the register: the pages and the chat box do the answering, and there is no pre-sales service to email or call.$n$),
      $n$The CTA is an invitation to a conversation, not a close.$n$,
      $n$The CTA is an invitation to start a brief or talk to the chat box, not a close.$n$),
      $n$They do not know the difference between hosting providers, DNS records, or CMS platforms, and they do not need to — those are the studio's problem, and the content should say so.$n$,
      $n$They can take a folder of HTML and CSS files, put it on free hosting and edit it: that is who this is for, and the copy says so before the buy button. Hosting, the domain and editing are the reader's own job after delivery; the content names the free hosting options, says the ZIP carries set-up instructions, and never says the studio handles them.$n$),
      $n$Technical terms that the customer does not need to understand should be handled by the studio — the copy says so and moves on. Where a term must be used (domain, hosting, CMS), define it once in the plainest possible way.$n$,
      $n$Technical terms the reader will meet after delivery (ZIP, static site, hosting, domain, DNS) are used plainly and defined once, because the reader handles those things themselves.$n$),
      $n$Judge it accordingly."]$n$,
      $n$Judge it accordingly.", "This is for experienced web designers: you get the files, and the site is yours to host, yours to edit and yours to maintain.", "We are not a hosting company. We can help you set it up on free hosting like Netlify."]$n$),
      $n$Judge it accordingly.\nWould never say:$n$,
      $n$Judge it accordingly.\n- This is for experienced web designers: you get the files, and the site is yours to host, yours to edit and yours to maintain.\n- We are not a hosting company. We can help you set it up on free hosting like Netlify.\nWould never say:$n$),
      $n$will block the page."]$n$,
      $n$will block the page.", "Where a page says what can be asked for, name the categories (business and service sites, tools and calculators, guides, reviews and enthusiast sites, reference and comparison directories, portfolios, personal, community and project sites) and say that we do not build online shops that take payment. Name categories, never example sites. Where a page discusses what happens after delivery, answer how the site is edited: the files are plain HTML and CSS, any text editor opens them, Visual Studio Code is free, and any web developer can take them on."]$n$),
      $n$will block the page.\n\nSentence style:$n$,
      $n$will block the page.\n- Where a page says what can be asked for, name the categories (business and service sites, tools and calculators, guides, reviews and enthusiast sites, reference and comparison directories, portfolios, personal, community and project sites) and say that we do not build online shops that take payment. Name categories, never example sites. Where a page discusses what happens after delivery, answer how the site is edited: the files are plain HTML and CSS, any text editor opens them, Visual Studio Code is free, and any web developer can take them on.\n\nSentence style:$n$),
      $n$State the limits plainly and early, in the same breath as the price: payment first, no sight of the site before paying, no approval stage, no changes, no refunds.$n$,
      $n$State the limits plainly and early, in the same breath as the price: who it is for (people who can host and edit a folder of files themselves), payment first, no sight of the site before paying, no approval stage, no changes, no refunds, and that we are not a hosting company.$n$)
    ::jsonb AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'content_direction', r.newdata, 'owner-ruling',
  'SQL_2026-08-25e: audience/hosting/editing/categories repositioning (owner brief 2026-08-25); structured fields and the formatted duplicate edited together as text.',
  true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

-- ── 2. identity ──
WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='identity' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(jsonb_set(jsonb_set(jsonb_set(jsonb_set(c.data,
      '{tagline}', to_jsonb($t$A starter website built once, for £149: the files are yours to host, edit and keep.$t$::text)),
      '{services}', $j$[
        {"name": "Starter website build", "description": "A starter site built by AI in one pass: the pages it needs, initial copy written for it, the design, and 30 days live at an address we provide. One fixed price of £149, paid once before the build, no VAT, no ongoing fee. Delivered as a ZIP of the files that is the customer's to host, edit and maintain."},
        {"name": "Domain rental or purchase", "description": "A .co.uk or .uk domain rented at £10 a month, or bought outright for a one-off £200 and then moved to the customer's own registrar; we do not stay their registrar. Hosting is not included: we are not a hosting company, and the ZIP carries instructions for free hosting."}
      ]$j$::jsonb),
      '{target_audience}', to_jsonb($t$Experienced web designers, and anyone comfortable taking a folder of HTML and CSS, putting it on free hosting and editing it themselves. UK-based, sceptical of agency pricing and page-builder subscriptions, and willing to trade revisions and hand-holding for a low fixed price and a working starting point. Not the non-technical owner who needs someone to run the site for them: the copy says so before the buy button.$t$::text)),
      '{unique_selling_points}', $j$[
        "One fixed price of £149, paid once, with no subscription, no monthly fee and no lock-in",
        "One-shot: no approval stage and no revision rounds, ready in three or four days, usually sooner, which is how the price stays down",
        "A starter site, not a finished product: initial copy included, and the files are the customer's to host, edit and maintain",
        "Any sort of site within the content limits: business and service sites, tools and calculators, guides and reviews, directories, portfolios, community and project sites; we do not build online shops that take payment",
        "Plain language and plain terms, stated before the buy button, including who this is and is not for",
        "Free hosting recommended by name, with set-up instructions in the ZIP; we are not a hosting company"
      ]$j$::jsonb),
      '{about_summary}', to_jsonb($t$webdesign.uk builds starter websites by AI for a single fixed price of £149 with no VAT and no ongoing fees, for experienced web designers and anyone comfortable running a static site themselves. Any sort of site within the content limits, not businesses only; we do not build online shops that take payment. The customer pays first and does not see the site before paying; there is no approval stage and the site is built once. It is delivered as a ZIP of the files to keep, which is theirs to host, edit and maintain, plus the site live at an address we provide for 30 days. We are not a hosting company: free hosting is recommended by name and the ZIP carries set-up instructions. No changes are included and we do not offer refunds.$t$::text))
    AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'identity', r.newdata, 'owner-ruling',
  'SQL_2026-08-25e: services (retired Post-acceptance changes REMOVED), target_audience, USPs (retired next-day figure REMOVED), about_summary, tagline repositioned per owner brief 2026-08-25.',
  true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

-- ── 3. mission_brief (aspect) and 6. submission.mission_brief (embedded copy): same replaces ──
WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='mission_brief' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(c.data, '{text}', to_jsonb(
      replace(replace(replace(replace(replace(c.data->>'text',
        $n$webdesign.uk sells complete starter websites for one fixed price.$n$,
        $n$webdesign.uk sells starter websites, built by AI, for one fixed price.$n$),
        $n$we build the whole site: the pages it needs, the words on them, the design, and getting it live.$n$,
        $n$we build the site once: the pages it needs, the words on them, the design, and 30 days live at an address we provide.$n$),
        $n$plus the site already live at a link we host for about a month.$n$,
        $n$plus the site already live at a link we host for 30 days. We are not a hosting company: after that the customer hosts it themselves, and we point at free hosting by name and put set-up instructions in the ZIP.$n$),
        $n$The people buying this need a decent website and have neither the time to make one nor the appetite for a drawn-out agency process. They are not technical.$n$,
        $n$The people buying this are experienced web designers, or anyone comfortable taking a folder of HTML and CSS, putting it on free hosting and editing it themselves: the site says so plainly before the buy button, because from delivery the site is theirs to host, edit and maintain. They want a working starting point rather than a blank page, because editing a site that already works is far less work than starting from scratch.$n$),
        $n$because the site is built in one pass from what they give us.$n$,
        $n$because the site is built in one pass from what they give us. It should say what kinds of site can be asked for: business and service sites, tools and calculators, guides, reviews and enthusiast sites, reference and comparison directories, portfolios, personal, community and project sites; and that we do not build online shops that take payment. And it should answer how the site is edited after delivery, because that is the trickier half of the offer: the files are plain HTML and CSS, any text editor opens them, Visual Studio Code is free, and any web developer can take them on.$n$)
    )) AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'mission_brief', r.newdata, 'owner-ruling',
  'SQL_2026-08-25e: buyer profile, 30 days, not-a-hosting-company, categories, how-to-edit (owner brief 2026-08-25).',
  true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='submission' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(c.data, '{mission_brief,text}', to_jsonb(
      replace(replace(replace(replace(replace(c.data->'mission_brief'->>'text',
        $n$webdesign.uk sells complete starter websites for one fixed price.$n$,
        $n$webdesign.uk sells starter websites, built by AI, for one fixed price.$n$),
        $n$we build the whole site: the pages it needs, the words on them, the design, and getting it live.$n$,
        $n$we build the site once: the pages it needs, the words on them, the design, and 30 days live at an address we provide.$n$),
        $n$plus the site already live at a link we host for about a month.$n$,
        $n$plus the site already live at a link we host for 30 days. We are not a hosting company: after that the customer hosts it themselves, and we point at free hosting by name and put set-up instructions in the ZIP.$n$),
        $n$The people buying this need a decent website and have neither the time to make one nor the appetite for a drawn-out agency process. They are not technical.$n$,
        $n$The people buying this are experienced web designers, or anyone comfortable taking a folder of HTML and CSS, putting it on free hosting and editing it themselves: the site says so plainly before the buy button, because from delivery the site is theirs to host, edit and maintain. They want a working starting point rather than a blank page, because editing a site that already works is far less work than starting from scratch.$n$),
        $n$because the site is built in one pass from what they give us.$n$,
        $n$because the site is built in one pass from what they give us. It should say what kinds of site can be asked for: business and service sites, tools and calculators, guides, reviews and enthusiast sites, reference and comparison directories, portfolios, personal, community and project sites; and that we do not build online shops that take payment. And it should answer how the site is edited after delivery, because that is the trickier half of the offer: the files are plain HTML and CSS, any text editor opens them, Visual Studio Code is free, and any web developer can take them on.$n$)
    )) AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'submission', r.newdata, 'owner-ruling',
  'SQL_2026-08-25e: the embedded mission_brief copy updated in step with the mission_brief aspect (the two-copies trap). roadmap_brief copy untouched.',
  true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

-- ── 4. briefing ──
WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='briefing' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(jsonb_set(jsonb_set(c.data,
      '{tagline}', to_jsonb($t$A starter website built once, for £149: the files are yours to host, edit and keep.$t$::text)),
      '{about_us}', to_jsonb($t$webdesign.uk builds starter websites by AI for a single fixed price of £149 with no VAT, no monthly fees to us, and no lock-in, for experienced web designers and anyone comfortable running a static site themselves. It builds any sort of site within the content limits, not just business sites, and does not build online shops that take payment. You tell us what you want and we build it once. You pay the £149 in full before any work starts, and you do not see the site before paying. It is ready in three or four days, usually sooner, from having what we need from you. What you get is a starter site with its initial copy written: a site you start with and edit, not a finished final product. You receive a ZIP of the finished site that is yours to host, edit and maintain, with instructions for putting it on free hosting, and your site already live at a link we host for 30 days. We are not a hosting company. No changes are included and we do not offer refunds. That is what the low fixed price pays for.$t$::text)),
      '{services}', $j$[
        {"name": "Starter website build", "description": "A starter site built by AI in one pass. One fixed price of £149, no VAT, no monthly fee to us, and no lock-in. Payment is taken in full before the build starts and there is no approval stage: the site is built once, and that is how the price stays down. Ready in three or four days, usually sooner, from having what we need from you. Delivered as a ZIP of the files that is yours to host, edit and maintain, plus the site live at a link we host for 30 days. No changes are included and we do not offer refunds."},
        {"name": "Domain rental or purchase", "description": "The site is delivered hosted by us for 30 days under a domain the customer can rent at £10 a month, with the subscription link in the delivery email, or buy outright for a one-off £200 and then move to their own registrar, because we do not stay their registrar. Keeping the site online beyond the included 30 days means hosting it themselves: we are not a hosting company, free options are recommended by name, and the ZIP carries the set-up instructions."}
      ]$j$::jsonb)
    AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'briefing', r.newdata, 'owner-ruling',
  'SQL_2026-08-25e: about_us, services, tagline repositioned; retired next-day figure and "transfer freely" REMOVED.',
  true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

-- ── 5. strategy ──
WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='strategy' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(jsonb_set(jsonb_set(c.data,
      '{value_proposition}', to_jsonb($t$A starter website, any sort of site within the content limits, built by AI in one pass and ready in three or four days, usually sooner, for £149 with no VAT and no ongoing fees. You pay first, there is no approval stage and no changes are included. You get a ZIP of the files that is yours to host, edit and maintain, plus the site live for 30 days at an address we provide. We are not a hosting company.$t$::text)),
      '{satisfaction_condition}', to_jsonb(replace(c.data->>'satisfaction_condition',
        $n$A visitor leaves knowing exactly what they get for £149, what is not included,$n$,
        $n$A visitor leaves knowing exactly what they get for £149, who it is for (people who can host and edit a folder of files themselves), what is not included,$n$))),
      '{content_strategy}', to_jsonb(replace(c.data->>'content_strategy',
        $n$The process is numbered and honest about where the work sits.$n$,
        $n$The process is numbered and plain about where the work sits.$n$)))
    AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'strategy', r.newdata, 'owner-ruling',
  'SQL_2026-08-25e: value_proposition rewritten (retired next-day figure REMOVED), satisfaction_condition gains who-it-is-for, content_strategy loses the banned word honest.',
  true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

-- ── 7. evidence_base.writer_block: the register line and the "complete website" phrase ──
WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(c.data, '{writer_block}', to_jsonb(replace(replace(c.data->>'writer_block',
      $n$Register: plain, direct British English, written for a business owner who is not technical and does not enjoy being sold to.$n$,
      $n$Register: plain, direct British English, written for an experienced web designer, or anyone comfortable running a static site themselves, who does not enjoy being sold to.$n$),
      $n$what the customer gets for £149 is a real, complete website built by software$n$,
      $n$what the customer gets for £149 is a real, working starter site built by software$n$))) AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'evidence_base', r.newdata, 'owner-ruling',
  'SQL_2026-08-25e: writer_block register line now names the experienced-web-designer reader; "complete website" -> "working starter site". Facts untouched.',
  true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

-- ── 8. page records: title/meta a rerender does not regenerate ──
UPDATE pages p SET
  title = CASE p.name WHEN 'index' THEN 'webdesign.uk: A starter website built once, for £149' ELSE p.title END,
  meta_description = CASE p.name
    WHEN 'index' THEN 'A starter site built by AI for £149, delivered as a ZIP that is yours to host, edit and maintain, and live at our address for 30 days.'
    WHEN 'contact' THEN 'Ask the chat box or start your website brief. We do not offer a pre-sales service: the pages and the chat box do the answering.'
    WHEN 'what-you-get' THEN 'A starter site as a ZIP you keep and host yourself, plus 30 days live at our address. What is included, and what is not.'
    WHEN 'tool-website-brief-starter' THEN 'Answer quick questions about the site you want, to create a brief for us to build from.'
    ELSE p.meta_description END
FROM sites s WHERE s.id=p.site_id AND s.domain='webdesign.uk' AND p.name IN ('index','contact','what-you-get','tool-website-brief-starter');

-- ── GUARDS ──
DO $chk$
DECLARE
  cd jsonb; pcd jsonb; t text; pt text;
  idn jsonb; mb text; pmb text; sub text; br jsonb; st jsonb; wb text; pwb text; n int;
BEGIN
  -- content_direction: both copies moved, needles gone, JSON valid
  SELECT ss.data INTO cd  FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='content_direction' AND ss.is_current;
  SELECT ss.data INTO pcd FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='content_direction' AND NOT ss.is_current ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1;
  t := cd::text; pt := pcd::text;
  IF (length(pt)-length(replace(pt,'Talks to a busy business owner','')))/length('Talks to a busy business owner') <> 2 THEN RAISE EXCEPTION 'control: register needle was not present twice before'; END IF;
  IF strpos(t,'Talks to a busy business owner')>0 THEN RAISE EXCEPTION 'content_direction: register needle survives'; END IF;
  IF strpos(t,'Send us an email. Call us.')>0 THEN RAISE EXCEPTION 'content_direction: Call us survives'; END IF;
  IF strpos(t,'those are the studio')>0 THEN RAISE EXCEPTION 'content_direction: studio-handles-it survives'; END IF;
  IF strpos(t,'should be handled by the studio')>0 THEN RAISE EXCEPTION 'content_direction: terminology needle survives'; END IF;
  IF strpos(cd->'voice'->>'register','experienced web designer')=0 THEN RAISE EXCEPTION 'content_direction structured register not updated'; END IF;
  IF strpos(cd->>'formatted','experienced web designer')=0 THEN RAISE EXCEPTION 'content_direction FORMATTED copy not updated'; END IF;
  IF strpos(cd->>'formatted','we do not build online shops that take payment')=0 OR jsonb_array_length(cd->'writing_rules') <> jsonb_array_length(pcd->'writing_rules')+1
    THEN RAISE EXCEPTION 'content_direction: categories rule missing from one copy'; END IF;
  IF jsonb_array_length(cd->'example_phrases'->'characteristic') <> jsonb_array_length(pcd->'example_phrases'->'characteristic')+2
    THEN RAISE EXCEPTION 'content_direction: example phrases not +2'; END IF;
  IF strpos(cd->'persuasion_approach'->>'trust_building','who it is for')=0 OR strpos(cd->>'formatted','who it is for')=0 THEN RAISE EXCEPTION 'trust_building not updated in both copies'; END IF;
  IF (length(t)-length(replace(t,'—',''))) > (length(pt)-length(replace(pt,'—',''))) THEN RAISE EXCEPTION 'content_direction gained an em dash'; END IF;

  -- identity
  SELECT ss.data INTO idn FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='identity' AND ss.is_current;
  IF idn::text ILIKE '%post-acceptance%' OR idn::text ILIKE '%next day%' OR idn::text ILIKE '%on acceptance%' THEN RAISE EXCEPTION 'identity: a retired phrase survives'; END IF;
  IF jsonb_array_length(idn->'services') <> 2 OR idn->'services'->0->>'name' <> 'Starter website build' THEN RAISE EXCEPTION 'identity services wrong'; END IF;
  IF strpos(idn->>'target_audience','Experienced web designers')=0 OR strpos(idn->>'about_summary','not a hosting company')=0 THEN RAISE EXCEPTION 'identity audience/about not updated'; END IF;
  IF idn->>'sub_industry' <> (SELECT ss.data->>'sub_industry' FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='identity' AND NOT ss.is_current ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1)
    THEN RAISE EXCEPTION 'identity: an untouched field moved'; END IF;

  -- mission_brief + submission copy
  SELECT ss.data->>'text' INTO mb FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='mission_brief' AND ss.is_current;
  SELECT ss.data->>'text' INTO pmb FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='mission_brief' AND NOT ss.is_current ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1;
  SELECT ss.data->'mission_brief'->>'text' INTO sub FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='submission' AND ss.is_current;
  IF strpos(pmb,'They are not technical.')=0 THEN RAISE EXCEPTION 'control: mission needle absent before'; END IF;
  IF strpos(mb,'They are not technical.')>0 OR strpos(mb,'about a month')>0 OR strpos(mb,'complete starter websites')>0 THEN RAISE EXCEPTION 'mission_brief: a needle survives'; END IF;
  IF strpos(mb,'online shops')=0 OR strpos(mb,'Visual Studio Code')=0 THEN RAISE EXCEPTION 'mission_brief: additions missing'; END IF;
  IF sub <> mb THEN RAISE EXCEPTION 'submission.mission_brief.text differs from mission_brief.text (the two-copies trap)'; END IF;

  -- briefing
  SELECT ss.data INTO br FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='briefing' AND ss.is_current;
  IF br::text ILIKE '%next day%' OR br::text ILIKE '%transfer freely%' OR br::text ILIKE '%about a month%' THEN RAISE EXCEPTION 'briefing: a retired phrase survives'; END IF;
  IF strpos(br->>'about_us','not a hosting company')=0 OR br->'services'->0->>'name' <> 'Starter website build' THEN RAISE EXCEPTION 'briefing not updated'; END IF;
  IF br->>'contact_email' <> 'webdesign@contactforsales.com' THEN RAISE EXCEPTION 'briefing: untouched field moved'; END IF;

  -- strategy
  SELECT ss.data INTO st FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='strategy' AND ss.is_current;
  IF st->>'value_proposition' ILIKE '%next day%' OR strpos(st->>'content_strategy','honest')>0 THEN RAISE EXCEPTION 'strategy: retired figure or banned word survives'; END IF;
  IF strpos(st->>'satisfaction_condition','who it is for')=0 OR strpos(st->>'value_proposition','not a hosting company')=0 THEN RAISE EXCEPTION 'strategy not updated'; END IF;
  IF st->>'money_flow' IS NULL OR st->>'growth_path' IS NULL THEN RAISE EXCEPTION 'strategy: an untouched field vanished'; END IF;

  -- writer_block
  SELECT ss.data->>'writer_block' INTO wb FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  SELECT ss.data->>'writer_block' INTO pwb FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1;
  IF strpos(pwb,'written for a business owner who is not technical')=0 THEN RAISE EXCEPTION 'control: writer_block register line absent before'; END IF;
  IF strpos(wb,'written for a business owner who is not technical')>0 OR strpos(wb,'real, complete website')>0 THEN RAISE EXCEPTION 'writer_block needle survives'; END IF;
  IF strpos(wb,'written for an experienced web designer')=0 THEN RAISE EXCEPTION 'writer_block register line not updated'; END IF;
  IF (SELECT ss.data->'facts' FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current)
     <> (SELECT ss.data->'facts' FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1)
    THEN RAISE EXCEPTION 'evidence_base facts moved; this write must not touch them'; END IF;

  -- pages
  SELECT count(*) INTO n FROM pages p JOIN sites s ON s.id=p.site_id WHERE s.domain='webdesign.uk'
    AND ((p.name='index' AND p.title LIKE '%starter website built once%' AND p.meta_description LIKE '%30 days%')
      OR (p.name='contact' AND p.meta_description LIKE '%We do not offer a pre-sales service%')
      OR (p.name='what-you-get' AND p.meta_description LIKE '%30 days%')
      OR (p.name='tool-website-brief-starter' AND p.meta_description LIKE '%the site you want%'));
  IF n <> 4 THEN RAISE EXCEPTION 'pages: expected 4 updated title/meta rows, found %', n; END IF;
  IF EXISTS (SELECT 1 FROM pages p JOIN sites s ON s.id=p.site_id WHERE s.domain='webdesign.uk' AND (p.title LIKE '%—%' OR p.meta_description LIKE '%—%' OR p.meta_description ILIKE '%two or three%' OR p.meta_description ILIKE '%your business%'))
    THEN RAISE EXCEPTION 'pages: an em dash or a retired phrase survives in a title/meta'; END IF;

  RAISE NOTICE 'ALL GUARDS PASSED: content_direction (both copies), identity, mission_brief + submission copy, briefing, strategy, writer_block, 4 page records';
END $chk$;

COMMIT;
