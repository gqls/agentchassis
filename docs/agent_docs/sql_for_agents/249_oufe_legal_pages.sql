-- 249_oufe_legal_pages.sql
-- oufe.com: the legal pages the owner approved on 2026-07-26 and which were
-- never built (sections A, D, E, F of DRAFT_disclaimer_for_owner_approval.md),
-- plus a privacy notice.
--
-- WHY NOW. Checked 2026-07-28: /disclaimer.html, /terms.html, /privacy.html and
-- /legal.html all returned 404 and the footer linked to none of them. The site
-- publishes analysis of a named real company with a live contact form and had no
-- correction route a reader could reach and no privacy notice. Owner decision the
-- same day, having no solicitor: publish the approved text now, self-draft the
-- privacy notice, and record the proceed-without-review decision contemporaneously
-- (the vetcomparison precedent).
--
-- SECTION G (liability cap) IS DELIBERATELY NOT INCLUDED. It is the one item
-- genuinely flagged for a solicitor and it applies only to paid products. Nothing
-- is for sale, so it is parked until the first sale rather than guessed at.
--
-- EVERY FACTUAL CLAIM IN THE PRIVACY NOTICE WAS VERIFIED against the live site on
-- 2026-07-28 before being written, because a privacy notice is the worst possible
-- place to assert something unchecked:
--   third-party scripts   none (only first-party /assets/js/snippets.js)
--   Set-Cookie on GET /   none
--   analytics / beacons   none found (no GTM, GA, Cloudflare Insights, Plausible, Matomo)
--   contact form          action="mailto:" - it opens the visitor's own mail client
--                         and sends NOTHING to this website
-- If any of that changes, this page is wrong and must be updated in the same
-- change that introduces the tracking.
--
-- PROTECTION follows migration 182: pages.rebuild_policy='owned' (save_page_sections
-- refuses a generic clobber on an owned page) PLUS page_components lock_type
-- ='permanent', and rendered_html written here in the same statement. Without the
-- last part the row renders as nothing for ever, because save_page_sections
-- PRESERVES locked rows rather than rendering them (loadActiveLockedRows) - a trap
-- this workstream walked into on the Thames mechanism section earlier today.

BEGIN;

DO $$
DECLARE
  s        uuid;
  gtb      uuid;
  legal_g  uuid;
  pg       uuid;
  pc       uuid;
  hd       text;
  ct       text;
  rh       text;
BEGIN
  SELECT id INTO s FROM sites WHERE domain = 'oufe.com';
  IF s IS NULL THEN RAISE EXCEPTION 'no site row for oufe.com'; END IF;

  SELECT id INTO gtb FROM content_components
   WHERE function = 'generic-text-block' AND component_level = 'section'
     AND COALESCE(is_active, true) ORDER BY created_at LIMIT 1;
  IF gtb IS NULL THEN RAISE EXCEPTION 'no generic-text-block component'; END IF;

  -- Legal nav group (footer). Reuse if a previous run or another thread made one.
  SELECT id INTO legal_g FROM site_nav_groups
   WHERE site_id = s AND group_key = 'legal' LIMIT 1;
  IF legal_g IS NULL THEN
    legal_g := gen_random_uuid();
    -- NB: site_nav_groups uses group_label, NOT label (site_nav_items uses label).
    INSERT INTO site_nav_groups (id, site_id, group_key, group_label, group_type, position)
    VALUES (legal_g, s, 'legal', 'Legal', 'legal', 90);
  END IF;

  -- ════════════════════════════ DISCLAIMER (sections E + F) ════════════════════════════
  IF NOT EXISTS (SELECT 1 FROM pages WHERE site_id = s AND name = 'disclaimer') THEN
    hd := 'Disclaimer, and how to correct us';
    ct := $disc$<p><strong>Last updated: 28 July 2026</strong></p>

<p>OUFE publishes educational analysis of financial and legal mechanism: how restructuring, liability management and distressed debt actually work when a company is under strain. This page sets out what that is, what it is not, and what to do if we have got something wrong.</p>

<h2>What OUFE is</h2>
<p>OUFE is an independent research publication. It stands for Oxen Unity Financial Engineering.</p>
<p><strong>OUFE is not an incorporated company.</strong> There is no company behind this site. It is not a firm, and it is not authorised or regulated by the Financial Conduct Authority or by any other regulator. Nothing here is provided in the course of a regulated activity.</p>

<h2>This is not advice</h2>
<p>Nothing on this site is investment advice, a recommendation, a financial promotion, or an inducement to engage in investment activity. It does not tell you what to buy, sell or hold, and it does not imply that any outcome or return is likely.</p>
<p>Neither is it legal advice. We describe how legal mechanisms work in general. Your situation is specific, and the difference between the two is where the money is. If a decision matters, take advice from someone who is regulated to give it and who knows your facts.</p>
<p>Reading this site creates no relationship between us. We are not acting for you, we owe you no duty of care in respect of your decisions, and we do not know your circumstances.</p>

<h2>How we source things, and the limits of that</h2>
<p>Where we state a fact about a real company, a court decision or a statute, we record the source we read it from and the date we read it, and we keep the exact sentence we relied on. You can check us.</p>
<p><strong>A citation shows you where something came from. It does not prove we understood it.</strong> We can pick the wrong passage, miss a later judgment that changes the position, or read a document correctly and still draw the wrong conclusion from it. The source itself can be wrong, incomplete or superseded. Situations move, and a page can be out of date before you reach it.</p>
<p>Where we have not verified something, we say so rather than estimating. "We have not checked this" is always publishable here. A plausible guess is not.</p>

<h2>We use AI assistance, and what that means for you</h2>
<p>This site is assembled with substantial machine assistance. Language models invent things fluently, and the most dangerous output is the plausible one: a real-looking figure in a well-formed sentence next to a real citation.</p>
<p>We run automated checks against that, and we have designed the site so that a published figure has to come from our evidence register with a source attached. Those checks reduce the risk. <strong>They do not eliminate it, and we are telling you that plainly rather than claiming a reliability we cannot demonstrate.</strong> Check anything that matters against the primary source.</p>

<h2>The tools are simplifications</h2>
<p>The interactive tools on this site compute one arithmetic rule over the numbers you type. Real outcomes turn on security, guarantees, structural subordination, intercompany claims, contingent liabilities and contested valuation evidence, none of which the tools model.</p>
<p>Treat every result as a possibly inaccurate worked example, useful for understanding how a mechanism behaves, and not as a statement about any real company or any real decision. Any pre-filled figures are illustrative. Do not rely on them.</p>

<h2>Analysis of mechanism is not prediction of outcome</h2>
<p>Explaining what a document permits, who ranks ahead of whom, and what the law allows a majority to do to a minority is not the same as saying what will happen. Discretion, negotiation, evidence and timing decide real cases. We write about the machinery, not the result.</p>

<h2>If we have got something wrong, tell us</h2>
<p>This is the part we would most like you to use.</p>
<p><strong>If anything here is wrong, tell us. We will correct it and say that we have.</strong> We would rather be corrected than be wrong in public, and a correction costs us nothing but a little pride.</p>
<p><strong>If you are the subject of a page and believe it misstates your position, write to us and we will review it promptly.</strong> Tell us what is wrong and, if you can, what the correct position is and where we can verify it. We will look at it seriously and quickly. Where we are wrong we will correct it; where we cannot verify a correction we will say what we could and could not check, rather than quietly leaving it or quietly removing it.</p>
<p>Write to <a href="mailto:oufe@contactforsales.com">oufe@contactforsales.com</a>.</p>

<h2>Changes to this page</h2>
<p>We may update this page. The date at the top shows when it was last revised.</p>$disc$;

    pg := gen_random_uuid(); pc := gen_random_uuid();
    rh := '<section id="' || pc::text || '" class="section section--generic">' || E'\n' ||
          '  <div class="container">' || E'\n' ||
          '    <h2 class="section__title">' || hd || '</h2>' || E'\n' ||
          '    <div class="section__content">' || ct || '</div>' || E'\n' ||
          '  </div>' || E'\n' || '</section>';

    INSERT INTO pages (id, site_id, name, url, title, page_type, status, build_status,
                       meta_description, nav_label, nav_order, in_header, in_footer,
                       sections, rebuild_policy, deployed_at)
    VALUES (pg, s, 'disclaimer', '/disclaimer.html', 'Disclaimer', 'content', 'active', 'needs_rebuild',
            'What OUFE is and is not, how we source what we publish, the limits of that, and how to tell us we have got something wrong.',
            'Disclaimer', 90, false, true,
            '["generic-text-block"]'::jsonb, 'owned', NULL);

    INSERT INTO page_components (id, page_id, slot_name, component_id, content_data, rendered_html,
                                 build_status, position, locked_at, locked_by, lock_type)
    VALUES (pc, pg, 'generic-text-block', gtb,
            jsonb_build_object('heading', hd, 'content', ct), rh,
            'pending', 0, now(), '249_oufe_legal_pages', 'permanent');

    INSERT INTO site_nav_items (site_id, group_id, label, url, page_id, item_type, position, status)
    VALUES (s, legal_g, 'Disclaimer', '/disclaimer.html', pg, 'page_link', 0, 'active');
  END IF;

  -- ════════════════════════════════ PRIVACY ════════════════════════════════
  IF NOT EXISTS (SELECT 1 FROM pages WHERE site_id = s AND name = 'privacy') THEN
    hd := 'Privacy';
    ct := $priv$<p><strong>Last updated: 28 July 2026</strong></p>

<p>This is a short page because this is a simple site. We collect very little, and most of what a privacy policy usually has to explain does not apply here.</p>

<h2>Who we are</h2>
<p>OUFE (Oxen Unity Financial Engineering) is an independent research publication. It is not an incorporated company. You can reach us at <a href="mailto:oufe@contactforsales.com">oufe@contactforsales.com</a>, and that address is also where any question about this page should go.</p>

<h2>This site sets no cookies and runs no analytics</h2>
<p>We do not set cookies. We do not run analytics. There are no advertising trackers, no third-party scripts, no social media pixels and no beacons on this site. We are not counting you, and we do not know who you are.</p>
<p>Verified on 28 July 2026. If we ever add analytics, we will change this page in the same change that adds them, rather than afterwards.</p>

<h2>The contact form does not send anything to us</h2>
<p>The form on our contact page opens your own email program with a message ready to send. Nothing is transmitted to this website, and nothing about your enquiry is stored here. If you decide to send that email, it arrives with us as an ordinary email.</p>

<h2>What we hold, and why</h2>
<p><strong>Email you send us.</strong> If you email us, we hold your email address and whatever you chose to put in the message, so that we can reply and keep a record of the correspondence. Under UK GDPR our lawful basis is legitimate interests: you contacted us and you expect an answer.</p>
<p>We do not add you to a mailing list. We do not sell or share your data. We do not use it to train AI models.</p>

<h2>Our hosting provider</h2>
<p>This site is served through standard web hosting and content-delivery infrastructure. As with any website, those providers may log ordinary request information, including IP addresses, as part of operating and protecting the service. That logging is theirs rather than ours, we do not use it to identify or track visitors, and we do not combine it with anything else.</p>

<h2>How long we keep it</h2>
<p>We keep correspondence while there is a reason to, and delete it when there is not. If you want us to delete your emails, ask and we will, unless we have a legal reason to keep them.</p>

<h2>Your rights</h2>
<p>If you are in the UK or the European Economic Area you can ask us what we hold about you, ask us to correct it, ask us to delete it, ask us to restrict what we do with it, object to us holding it, or ask for a copy in a portable form. Write to <a href="mailto:oufe@contactforsales.com">oufe@contactforsales.com</a> and we will answer within one month. We do not charge for reasonable requests.</p>
<p>If you are unhappy with how we have handled your data, you can complain to the Information Commissioner's Office at <a href="https://ico.org.uk" target="_blank" rel="noopener noreferrer">ico.org.uk</a>. You do not need our permission and you do not have to raise it with us first, though we would rather you gave us the chance to fix it.</p>

<h2>Children</h2>
<p>This site is written for professionals. We do not seek or knowingly hold personal data about anyone under 16.</p>

<h2>Changes</h2>
<p>If this page changes, the date at the top changes with it.</p>$priv$;

    pg := gen_random_uuid(); pc := gen_random_uuid();
    rh := '<section id="' || pc::text || '" class="section section--generic">' || E'\n' ||
          '  <div class="container">' || E'\n' ||
          '    <h2 class="section__title">' || hd || '</h2>' || E'\n' ||
          '    <div class="section__content">' || ct || '</div>' || E'\n' ||
          '  </div>' || E'\n' || '</section>';

    INSERT INTO pages (id, site_id, name, url, title, page_type, status, build_status,
                       meta_description, nav_label, nav_order, in_header, in_footer,
                       sections, rebuild_policy, deployed_at)
    VALUES (pg, s, 'privacy', '/privacy.html', 'Privacy', 'content', 'active', 'needs_rebuild',
            'OUFE sets no cookies, runs no analytics and holds only the email you choose to send us.',
            'Privacy', 91, false, true,
            '["generic-text-block"]'::jsonb, 'owned', NULL);

    INSERT INTO page_components (id, page_id, slot_name, component_id, content_data, rendered_html,
                                 build_status, position, locked_at, locked_by, lock_type)
    VALUES (pc, pg, 'generic-text-block', gtb,
            jsonb_build_object('heading', hd, 'content', ct), rh,
            'pending', 0, now(), '249_oufe_legal_pages', 'permanent');

    INSERT INTO site_nav_items (site_id, group_id, label, url, page_id, item_type, position, status)
    VALUES (s, legal_g, 'Privacy', '/privacy.html', pg, 'page_link', 1, 'active');
  END IF;

END $$;

COMMIT;

-- VERIFY
SELECT p.name, p.url, p.rebuild_policy, p.build_status,
       pc.lock_type, length(pc.rendered_html) AS rendered_bytes
FROM pages p JOIN page_components pc ON pc.page_id = p.id
WHERE p.site_id = (SELECT id FROM sites WHERE domain = 'oufe.com')
  AND p.name IN ('disclaimer', 'privacy')
ORDER BY p.name;
