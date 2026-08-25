-- SQL_2026-08-25 — OWNER COPY BRIEF 2026-08-25: bolder home-page truth, wider scope.
-- Brief + readings (incl. ".csv" read as ZIP): BRIEF_2026-08-25_bold_home_copy_and_wider_scope.md
--
-- WHAT THIS ATTESTS (new): audience_experienced_webdesigners (the buyer gate: files
-- are yours to host, edit, maintain); not_a_hosting_company (in those words, with the
-- Netlify help alongside). WHAT IT AMENDS: the included month is fixed at 30 DAYS
-- (delivery_live_link_and_zip, keep_it_online, writer_block x5); any_site_type gains
-- the category vocabulary (categories on the page, example names still OFF per the
-- owner's 2026-08-18 attestation); third_party_options gains an Editing category
-- (Visual Studio Code) for the new how-to-edit answer; allowed_entities gains VS Code.
--
-- DEFECT FIXED IN PASSING: writer_block's "HOW THE SITE LEADS" told the writer to
-- name example sites "which fact any_site_type_examples attests" — that fact does not
-- exist and the instruction contradicts the no-examples rule two paragraphs up. A
-- ghost from before the 08-18 examples-deferred ruling; rewritten to match.
--
-- DELIBERATELY NOT DONE: the owner's "until we do start hosting" hedge (a future
-- tease invites pre-sales questions nobody may answer, per the 2026-08-21 "for now"
-- precedent); any "no online shops" category exclusion (a commercial term no fact
-- attests — flagged to the owner instead).
--
-- Run: trial with COMMIT->ROLLBACK sed first; guards RAISE on any drift. Two lanes
-- write this row: guards compare against the row THIS transaction supersedes.

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned
  FROM site_specs ss JOIN sites s ON s.id = ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(
      jsonb_set(
        jsonb_set(
          c.data,
          '{facts}',
          (
            SELECT jsonb_agg(
              CASE f->>'id'
                WHEN 'delivery_live_link_and_zip' THEN
                  jsonb_set(jsonb_set(jsonb_set(f,
                    '{claim}', to_jsonb(replace(f->>'claim',
                      'it stays live for about a month','it stays live for 30 days'))),
                    '{writer_line}', to_jsonb(replace(f->>'writer_line',
                      'It stays live for about a month','It stays live for 30 days'))),
                    '{source,attested_by}', to_jsonb((f->'source'->>'attested_by') ||
                      $s$; owner, 2026-08-25, copy brief ("You can see what we've built for 30 days at a link we'll give you"): the month is fixed at 30 days$s$))
                WHEN 'keep_it_online' THEN
                  jsonb_set(jsonb_set(jsonb_set(f,
                    '{claim}', to_jsonb(replace(f->>'claim',
                      'beyond the included month','beyond the included 30 days'))),
                    '{writer_line}', to_jsonb(replace(f->>'writer_line',
                      'after that month','after those 30 days'))),
                    '{source,attested_by}', to_jsonb((f->'source'->>'attested_by') ||
                      '; owner, 2026-08-25, copy brief: the included month is fixed at 30 days'))
                WHEN 'any_site_type' THEN
                  jsonb_set(jsonb_set(jsonb_set(f,
                    '{claim}', to_jsonb(replace(f->>'claim',
                      'personal, community and project sites are all welcome.',
                      'personal, community and project sites are all welcome, and so are tools and calculators, guides, reviews and enthusiast sites, and reference and comparison directories.'))),
                    '{writer_line}', to_jsonb(
                      'Not just business sites: tools and calculators, guides and reviews, comparison directories, personal, community and project sites all fit, within the content limits.'::text)),
                    '{source,attested_by}', to_jsonb((f->'source'->>'attested_by') ||
                      $s$; owner, 2026-08-25, copy brief: scope read as too limited, more categories wanted (motivating examples dartsonline.com, enthusiast buying guides and news, and robot-hands.com, technical reference with interactive tools; names drawn from the estate classifier vocabulary: interactive / hub / editorial / brochure). Example-site names stay off the page per the 2026-08-18 attestation in this claim$s$))
                WHEN 'third_party_options' THEN
                  jsonb_set(jsonb_set(f,
                    '{claim}', to_jsonb(replace(f->>'claim',
                      'email the result.',
                      'email the result. Editing the files: Visual Studio Code, a free editor that opens and edits the site files, which are plain HTML and CSS that any text editor can open.'))),
                    '{source,attested_by}', to_jsonb((f->'source'->>'attested_by') ||
                      '; editing category (Visual Studio Code) added by owner direction 2026-08-25, the how-to-edit answer; VS Code verified free and currently maintained, 2026-08-25'))
                ELSE f
              END ORDER BY ord)
            FROM jsonb_array_elements(c.data->'facts') WITH ORDINALITY AS t(f, ord)
          ) || jsonb_build_array(
            jsonb_build_object(
              'id','audience_experienced_webdesigners',
              'kind','attestation',
              'claim', $a$The offer is aimed at experienced web designers, and at anyone comfortable editing and hosting a static site: what is delivered is a ZIP of files, and from delivery the site is the customer's to host, to edit and to maintain. This bounds the buyer, not the site's subject: any_site_type still governs what the sites may be about.$a$,
              'source', jsonb_build_object('attested_by',
                $a$owner, 2026-08-25, copy brief (verbatim: "This is for experienced webdesigners because we're just giving you a .csv file with your static site after which it's yours to host, yours to edit and yours to maintain."), clarified by the owner the same day: "It is really just being more honest about who could handle what I give them" - so the claim is a capability statement, not a credentials gate. The ".csv" is read as the ZIP of delivery_live_link_and_zip: the attested and built deliverable is a ZIP and a static site cannot be a .csv; the reading was flagged back to the owner the same day.$a$),
              'verified_at','2026-08-25',
              'writer_line','This is for experienced web designers: you get the files, and the site is yours to host, yours to edit and yours to maintain.',
              'context_terms', jsonb_build_array('experienced','web designer','web designers','yours to host','yours to edit','maintain')
            ),
            jsonb_build_object(
              'id','not_a_hosting_company',
              'kind','attestation',
              'claim', $a$We are not a hosting company, and hosting is not part of what is sold. The link we provide exists so the customer can see the finished site for its 30 days; keeping the site online is the customer's own job, and the help we give is the ZIP's written instructions and the named free-hosting recommendations, per keep_it_online and third_party_options.$a$,
              'source', jsonb_build_object('attested_by',
                $a$owner, 2026-08-25, copy brief ("We are not a hosting company so, instead, we can help you set it up on free hosting like netlify."). The draft's "until we do start hosting" hedge is deliberately not attested: a future tease invites pre-sales questions nobody may answer (no_presales_service), per the 2026-08-21 "for now" precedent on TLD scope.$a$),
              'verified_at','2026-08-25',
              'writer_line','We are not a hosting company. We can help you set it up on free hosting like Netlify.',
              'context_terms', jsonb_build_array('hosting company','host it','netlify','free hosting')
            )
          )
        ),
        '{writer_block}',
        to_jsonb(
          replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(
            c.data->>'writer_block',
            'The register already carries six named third-party services by category',
            'The register already carries seven named third-party services by category'),
            'it stays live for about a month and the ZIP is permanent.',
            'it stays live for 30 days and the ZIP is permanent.'),
            'staying live for about a month.',
            'staying live for 30 days.'),
            'After payment the site is simply live, at an address we provide, for about a month.',
            'After payment the site is simply live, at an address we provide, for 30 days.'),
            'keeping it online beyond that month means hosting it yourself.',
            'keeping it online beyond those 30 days means hosting it yourself.'),
            'that keeping the site online beyond the included month means hosting it yourself',
            'that keeping the site online beyond the included 30 days means hosting it yourself'),
            'Name the real sites built by this system, which fact any_site_type_examples attests, and say plainly that this site was built the same way, which fact ai_built attests.',
            'Say plainly that this site was built by the same system that builds the sites it sells, which fact ai_built attests, and name the kinds of site on offer, per any_site_type. Do not name example sites: none built by this one-shot route exist yet.'),
            'DO name the six services',
            'DO name the seven services'),
            'The six and their groups, exactly and no others:',
            'The seven and their groups, exactly and no others:'),
            'For making a contact form work: Formspree and Basin. Never rank them',
            'For making a contact form work: Formspree and Basin. For editing the files: Visual Studio Code. Never rank them'),
            'the six named third-party services (Cloudflare Pages, Netlify, Fathom Analytics, Plausible, Formspree, Basin)',
            'the seven named third-party services (Cloudflare Pages, Netlify, Fathom Analytics, Plausible, Formspree, Basin, Visual Studio Code)'),
            'that no pre-sales service is included,',
            'that the offer is aimed at experienced web designers and the site is theirs to host, edit and maintain, that we are not a hosting company, that no pre-sales service is included,')
          || $wb$

WHO THIS IS FOR, SAID BEFORE THE BUY BUTTON (owner ruling, 2026-08-25). This is for experienced web designers. Say it early and plainly, in the same flat voice as the price: the product is a ZIP of files, and from delivery the site is the customer's to host, to edit and to maintain. Anyone comfortable editing and hosting a static site passes the test; it is a fit test the reader applies to themselves, never gatekeeping and never a warning. It bounds the buyer, not the site: any sort of site can still be asked for, per any_site_type.

WE ARE NOT A HOSTING COMPANY, in those words (owner ruling, 2026-08-25). Say it wherever hosting comes up, and follow it at once with the help, per the useful-not-promotional rule: we can help you set it up on free hosting like Netlify. The help is the ZIP's instructions and the named recommendations, never our time and never a done-for-you set-up.

WHY A STARTER SITE, the argument to keep (owner draft, 2026-08-25, keep its substance and its plainness): editing an existing site to suit what you want is a whole lot easier than starting from scratch. These starter sites are exactly that: a more or less complete starter template for the customer to carry on with.

HOW TO EDIT THE SITE gets its own answer (owner ruling, 2026-08-25). It is the trickier half of the offer, and the page says so plainly instead of hiding it. Then give the ways, as help: the files are plain HTML and CSS, so any text editor opens them; Visual Studio Code is free and made for exactly this, per third_party_options; and anyone the customer hires can take the files on, per taking_it_further. Never offer our time with it, and never promise what a tool or a hire will do.

THE CATEGORIES, wherever a page says what can be asked for (owner ruling, 2026-08-25): business and service sites, tools and calculators, guides, reviews and enthusiast sites, reference and comparison directories, and portfolios, personal, community and project sites. Name categories, never example sites.$wb$
        )
      ),
      '{allowed_entities}',
      (c.data->'allowed_entities') || '["Visual Studio Code"]'::jsonb
    ) AS newdata
  FROM cur c
),
retire AS (
  UPDATE site_specs ss SET is_current=false, superseded_at=now()
  WHERE ss.id=(SELECT id FROM cur) RETURNING 1
)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'evidence_base', r.newdata, 'owner-ruling',
  'SQL_2026-08-25: owner copy brief. NEW facts audience_experienced_webdesigners + not_a_hosting_company; month fixed at 30 days (2 facts + writer_block x5); any_site_type gains category vocabulary (examples still off-page per 2026-08-18); third_party_options + allowed_entities gain Visual Studio Code; writer_block gains 5 steering paragraphs and loses the any_site_type_examples ghost reference. Pages NOT yet rebuilt: served copy trails this row until the index rebuild.',
  true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

DO $chk$
DECLARE
  d jsonb; prev jsonb; n int;
  wb text; pwb text; f jsonb; pf jsonb;
BEGIN
  SELECT ss.data INTO d FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  SELECT ss.data INTO prev FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current
   ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1;
  IF prev IS NULL THEN RAISE EXCEPTION 'no superseded row to compare against'; END IF;

  -- facts: exactly two ADDED, none removed (no fixed total: two lanes write this row)
  SELECT count(*) INTO n FROM (
    SELECT x->>'id' FROM jsonb_array_elements(prev->'facts') AS t(x)
    EXCEPT SELECT x->>'id' FROM jsonb_array_elements(d->'facts') AS t(x)) q;
  IF n <> 0 THEN RAISE EXCEPTION '% fact id(s) VANISHED', n; END IF;
  SELECT count(*) INTO n FROM (
    SELECT x->>'id' FROM jsonb_array_elements(d->'facts') AS t(x)
    EXCEPT SELECT x->>'id' FROM jsonb_array_elements(prev->'facts') AS t(x)) q;
  IF n <> 2 THEN RAISE EXCEPTION 'expected exactly 2 new fact ids, got %', n; END IF;

  PERFORM 1 FROM jsonb_array_elements(d->'facts') AS t(x)
   WHERE x->>'id'='audience_experienced_webdesigners'
     AND x->>'claim' LIKE '%experienced web designers%'
     AND x->'source'->>'attested_by' LIKE '%2026-08-25%'
     AND COALESCE(x->>'writer_line','') <> '';
  IF NOT FOUND THEN RAISE EXCEPTION 'audience fact absent or malformed'; END IF;
  PERFORM 1 FROM jsonb_array_elements(d->'facts') AS t(x)
   WHERE x->>'id'='not_a_hosting_company'
     AND x->>'claim' LIKE '%not a hosting company%'
     AND x->>'writer_line' LIKE '%Netlify%';
  IF NOT FOUND THEN RAISE EXCEPTION 'hosting fact absent or malformed'; END IF;

  -- reconstruction: the 30-days amendments derive EXACTLY from the superseded text
  SELECT x INTO f  FROM jsonb_array_elements(d->'facts')    AS t(x) WHERE x->>'id'='delivery_live_link_and_zip';
  SELECT x INTO pf FROM jsonb_array_elements(prev->'facts') AS t(x) WHERE x->>'id'='delivery_live_link_and_zip';
  IF strpos(pf->>'claim','it stays live for about a month')=0 THEN RAISE EXCEPTION 'control: delivery needle was not in the superseded claim'; END IF;
  IF f->>'claim' <> replace(pf->>'claim','it stays live for about a month','it stays live for 30 days')
    THEN RAISE EXCEPTION 'delivery claim reconstruction failed'; END IF;
  IF f->>'writer_line' <> replace(pf->>'writer_line','It stays live for about a month','It stays live for 30 days')
    THEN RAISE EXCEPTION 'delivery writer_line reconstruction failed'; END IF;

  SELECT x INTO f  FROM jsonb_array_elements(d->'facts')    AS t(x) WHERE x->>'id'='keep_it_online';
  SELECT x INTO pf FROM jsonb_array_elements(prev->'facts') AS t(x) WHERE x->>'id'='keep_it_online';
  IF strpos(pf->>'claim','beyond the included month')=0 THEN RAISE EXCEPTION 'control: keep_it_online needle was not in the superseded claim'; END IF;
  IF f->>'claim' <> replace(pf->>'claim','beyond the included month','beyond the included 30 days')
    THEN RAISE EXCEPTION 'keep_it_online claim reconstruction failed'; END IF;
  IF f->>'writer_line' <> replace(pf->>'writer_line','after that month','after those 30 days')
    THEN RAISE EXCEPTION 'keep_it_online writer_line reconstruction failed'; END IF;

  SELECT x INTO f FROM jsonb_array_elements(d->'facts') AS t(x) WHERE x->>'id'='any_site_type';
  IF strpos(f->>'claim','tools and calculators')=0 OR strpos(f->>'claim','No example sites are named yet')=0
    THEN RAISE EXCEPTION 'any_site_type: categories missing or the no-examples rule was lost'; END IF;
  PERFORM 1 FROM jsonb_array_elements(d->'facts') AS t(x)
   WHERE x->>'id'='third_party_options' AND x->>'claim' LIKE '%Visual Studio Code%';
  IF NOT FOUND THEN RAISE EXCEPTION 'third_party_options: Visual Studio Code missing'; END IF;

  -- untouched keys stay byte-identical; allowed_entities gains exactly VS Code
  IF d->'banned_claims'  <> prev->'banned_claims'  THEN RAISE EXCEPTION 'banned_claims changed'; END IF;
  IF d->'governing_rule' <> prev->'governing_rule' THEN RAISE EXCEPTION 'governing_rule changed'; END IF;
  IF d->'allowed_entities' <> (prev->'allowed_entities') || '["Visual Studio Code"]'::jsonb
    THEN RAISE EXCEPTION 'allowed_entities is not exactly prev + Visual Studio Code'; END IF;

  wb := d->>'writer_block'; pwb := prev->>'writer_block';
  -- controls first: the superseded block really carried what this write removes
  IF strpos(pwb,'any_site_type_examples')=0 THEN RAISE EXCEPTION 'control: ghost ref already absent before this write'; END IF;
  IF strpos(pwb,'about a month')=0 THEN RAISE EXCEPTION 'control: about-a-month already absent before this write'; END IF;
  IF strpos(pwb,'the six named third-party services')=0 THEN RAISE EXCEPTION 'control: six-services form already absent'; END IF;
  -- removals
  IF strpos(wb,'any_site_type_examples')>0 THEN RAISE EXCEPTION 'ghost fact reference survives'; END IF;
  IF strpos(wb,'about a month')>0 OR strpos(wb,'beyond that month')>0 OR strpos(wb,'beyond the included month')>0
    THEN RAISE EXCEPTION 'a month-form survives in writer_block'; END IF;
  IF strpos(wb,'six named third-party')>0 OR strpos(wb,'The six and their groups')>0 OR strpos(wb,'DO name the six services')>0
    THEN RAISE EXCEPTION 'a six-services form survives in writer_block'; END IF;
  -- additions, each exactly once
  IF (length(wb)-length(replace(wb,'WHO THIS IS FOR, SAID BEFORE THE BUY BUTTON','')))/length('WHO THIS IS FOR, SAID BEFORE THE BUY BUTTON') <> 1
    THEN RAISE EXCEPTION 'WHO THIS IS FOR paragraph not exactly once'; END IF;
  IF (length(wb)-length(replace(wb,'WE ARE NOT A HOSTING COMPANY, in those words','')))/length('WE ARE NOT A HOSTING COMPANY, in those words') <> 1
    THEN RAISE EXCEPTION 'NOT A HOSTING COMPANY paragraph not exactly once'; END IF;
  IF (length(wb)-length(replace(wb,'HOW TO EDIT THE SITE gets its own answer','')))/length('HOW TO EDIT THE SITE gets its own answer') <> 1
    THEN RAISE EXCEPTION 'HOW TO EDIT paragraph not exactly once'; END IF;
  IF (length(wb)-length(replace(wb,'THE CATEGORIES, wherever a page says','')))/length('THE CATEGORIES, wherever a page says') <> 1
    THEN RAISE EXCEPTION 'CATEGORIES paragraph not exactly once'; END IF;
  IF strpos(wb,'For editing the files: Visual Studio Code.')=0
    THEN RAISE EXCEPTION 'FAQ service list did not gain the editing entry'; END IF;
  IF strpos(wb,'that we are not a hosting company, that no pre-sales service is included,')=0
    THEN RAISE EXCEPTION 'facts enumeration not extended'; END IF;
  -- em dashes: the additions must introduce none (the block bans them; 6 legacy ones exist)
  IF (length(wb)-length(replace(wb,'—',''))) > (length(pwb)-length(replace(pwb,'—','')))
    THEN RAISE EXCEPTION 'writer_block GAINED an em dash'; END IF;

  RAISE NOTICE 'ALL GUARDS PASSED: 2 facts added, 4 amended, writer_block 12 edits + 5 paragraphs, entities +1';
END $chk$;

COMMIT;
