-- 182_legal_pages_aao_finetuning.sql — 2026-07-21, cta_link_integrity (bugs_open/049 candidate 2)
--
-- Creates the missing legal pages so ai-agent-orchestration.com and finetuning.uk
-- can be chrome-refreshed WITHOUT triggering bugs_open/053 (an empty `legal` nav
-- group falls back to listing every footer page in the legal slot). With real
-- legal pages + real legal nav items, GetNavItems('legal') returns rows and the
-- fallback never fires.
--
-- State before (verified 2026-07-21):
--   ai-agent-orchestration.com : NO privacy, NO terms, NO legal nav group.
--   finetuning.uk              : HAS /privacy-policy.html + a legal nav group with
--                                one item (Privacy Policy); NO terms page.
--
-- This migration adds:
--   aao  -> /privacy.html  + /terms.html  + a new `legal` nav group (Privacy, Terms)
--   ft   -> /terms.html                    + a Terms item in the existing legal group
--
-- Pages are content pages using the `generic-text-block` component (the same
-- mechanism as finetuning's existing, owner-approved /privacy-policy.html):
-- content lives in page_components.content_data {heading, content} and, because
-- rerender reads STORED rendered_html (rerenderLoadSections, no LLM), the
-- rendered_html is set here too so the re-render ships this exact text.
--
-- CONTENT PROVENANCE — verifiable facts only, no fabrication:
--   * The aao privacy policy MIRRORS finetuning's live, owner-approved privacy
--     policy (structure + honest "we'll name tools as confirmed" hedges), with
--     only the identity swapped (name/email/description from site_specs.identity
--     + sites.company_name).
--   * The terms pages are standard UK website terms of use for a consultancy,
--     asserting nothing site-specific beyond verified identity. They explicitly
--     do NOT invent: company registration number, VAT number, ICO registration,
--     registered/physical address (both sites' site_specs.contact.address IS NULL,
--     location "United Kingdom"), DPO, or named processors. Governing law is
--     stated as England and Wales (the conventional UK default) — the owner
--     should confirm if the entity is in Scotland or NI. See the HANDOFF for the
--     full list of owner-fill items.
--
-- Facts used (site_specs.identity + sites, 2026-07-21):
--   aao : "AI Agent Orchestration" · agents@contactforsales.com · +44 (0) 7934 524 911 · UK
--   ft  : "FineTuning"             · finetune@contactforsales.com · +44 (0) 7934 524 911 · UK
--
-- Protection: pages.rebuild_policy='owned' (save_page_sections refuses a generic
-- clobber on owned pages) + page_components lock_type='permanent'. finetuning's
-- existing privacy is only 'generic' and is therefore LESS protected — noted in
-- the handoff as a follow-up, not changed here.
--
-- Deploy is SEPARATE: after this applies, run rerender-pages with
-- refresh_site_components:true per site (049_TRIGGER_chrome_refresh.sh). Verify
-- live before closing.
--
-- ROLLBACK: bak_legal_pages_20260721 records the created ids; delete those
-- page_components, pages, site_nav_items and the aao nav group by id.

\set ON_ERROR_STOP on
BEGIN;

CREATE TABLE IF NOT EXISTS bak_legal_pages_20260721 (
  what text, id uuid, created_at timestamptz DEFAULT now()
);

DO $mig$
DECLARE
  gtb        uuid := '8d81e665-3ee0-443d-a873-690268c15fbb';  -- generic-text-block component
  aao        uuid := '2a8ebf9c-20a2-4c39-b191-840b012371da';
  ft         uuid := '1368e337-dd1d-4799-bbb3-8221a1b79bcc';
  ft_legal_g uuid := 'be717ff5-ebec-4e33-9206-2d509cda0649';  -- finetuning existing legal group
  aao_legal_g uuid;
  pg uuid;   -- page id, reused
  pc uuid;   -- page_component id, reused
  hd text;   -- heading
  ct text;   -- content
  rh text;   -- rendered_html
BEGIN
  RAISE NOTICE 'legal-pages migration starting';

  -- guards: the two sites must be as measured
  IF (SELECT count(*) FROM pages WHERE site_id=aao AND url IN ('/privacy.html','/terms.html') AND status='active') <> 0 THEN
    RAISE EXCEPTION 'aao already has an active /privacy.html or /terms.html — re-verify before applying';
  END IF;
  IF (SELECT count(*) FROM pages WHERE site_id=ft AND url='/terms.html' AND status='active') <> 0 THEN
    RAISE EXCEPTION 'finetuning already has an active /terms.html — re-verify before applying';
  END IF;
  IF (SELECT count(*) FROM site_nav_groups WHERE site_id=aao AND group_type='legal') <> 0 THEN
    RAISE EXCEPTION 'aao already has a legal nav group — re-verify before applying';
  END IF;

  -- ════════════════════════ aao legal nav group ════════════════════════
  aao_legal_g := gen_random_uuid();
  INSERT INTO site_nav_groups (id, site_id, group_key, group_label, group_type, position)
  VALUES (aao_legal_g, aao, 'legal', 'Legal', 'legal', 10);
  INSERT INTO bak_legal_pages_20260721(what,id) VALUES ('aao_nav_group', aao_legal_g);

  -- ════════════════════════ aao PRIVACY ════════════════════════
  hd := 'Privacy Policy';
  ct := $aao_priv$<p><strong>Last updated: July 2026</strong></p><p>This policy explains what personal data AI Agent Orchestration collects, why we collect it, how we store and use it, and what your rights are. We have tried to write it in plain English. If something is unclear, contact us and we will explain it.</p><h2>Who we are</h2><p>AI Agent Orchestration is a consultancy that designs, builds, and operates production-grade multi-agent AI systems on Kubernetes, Kafka, and Postgres for businesses. Our contact details are:</p><p>Email: <a href='mailto:agents@contactforsales.com'>agents@contactforsales.com</a><br>Phone: +44 (0) 7934 524 911</p><p>For the purposes of UK GDPR, AI Agent Orchestration is the data controller for personal data collected through this website and our business activities.</p><h2>What data we collect and why</h2><p><strong>Contact form submissions.</strong> When you fill in a contact or enquiry form on this website, we collect your name, email address, and any information you choose to include in your message. We use this to respond to your enquiry. The legal basis is legitimate interests — you have contacted us and expect a reply.</p><p><strong>Discovery calls.</strong> If you book or attend a discovery call with us, we may take notes about your business, the problems you are trying to solve, and the systems or processes you describe. These notes are used solely to prepare a relevant proposal or follow-up. We do not record calls without your explicit consent.</p><p><strong>Email correspondence.</strong> If you email us directly, we retain that correspondence as part of our business records. We do not add you to any mailing list without your explicit consent.</p><p><strong>Website analytics.</strong> We may collect anonymised usage data about how visitors interact with this website — pages visited, time on page, device type, and similar. Where we use analytics tools, we configure them to anonymise IP addresses and avoid storing personally identifiable information where possible. We will update this section with specific tool names as our analytics setup is confirmed.</p><h2>How we store your data</h2><p>Enquiry records and correspondence are stored in business tools with appropriate access controls. We take reasonable technical steps to protect personal data from unauthorised access, loss, or disclosure.</p><p>We do not sell your data to third parties. We do not share your data with marketing companies. We do not use your data to train AI models without a separate, explicit agreement with you.</p><h2>Third-party tools</h2><p>We use a small number of third-party services in the course of running our business — for example email and calendar services, website hosting and infrastructure, and video conferencing tools for calls. These are standard business tools operated by providers with their own privacy policies. Hosting providers may log IP addresses and request data as part of standard server operation. Where any third-party tool processes personal data on our behalf, we take reasonable steps to ensure appropriate data processing agreements are in place, and we will update this section with specific named tools as our tooling is confirmed and reviewed.</p><h2>How long we keep your data</h2><p>Enquiry records and correspondence are kept for as long as there is a legitimate business reason to retain them — typically for the duration of a business relationship and a reasonable period afterwards for reference and legal purposes. If you would like us to delete your data, contact us and we will do so unless we have a legal obligation to retain it.</p><h2>Your rights under UK GDPR</h2><p>If you are based in the UK or European Economic Area, you have the following rights regarding your personal data:</p><p><strong>Right of access.</strong> You can ask us what personal data we hold about you and receive a copy of it.</p><p><strong>Right to rectification.</strong> If data we hold about you is inaccurate or incomplete, you can ask us to correct it.</p><p><strong>Right to erasure.</strong> You can ask us to delete your personal data. We will do so unless we have a legal obligation to retain it.</p><p><strong>Right to restriction.</strong> You can ask us to restrict how we use your data while a dispute or request is being resolved.</p><p><strong>Right to object.</strong> You can object to our processing of your data where we rely on legitimate interests as the legal basis.</p><p><strong>Right to data portability.</strong> Where processing is based on consent or contract and carried out by automated means, you can ask for your data in a portable format.</p><p>To exercise any of these rights, contact us at <a href='mailto:agents@contactforsales.com'>agents@contactforsales.com</a>. We will respond within one month. We do not charge for reasonable requests.</p><p>If you are not satisfied with how we handle your data or respond to a request, you have the right to lodge a complaint with the Information Commissioner's Office (ICO) in the UK. Details are available at <a href='https://ico.org.uk' target='_blank' rel='noopener noreferrer'>ico.org.uk</a>.</p><h2>Cookies</h2><p>This website may use cookies for basic functionality and analytics. Where we use non-essential cookies, we will seek your consent before setting them. You can control cookies through your browser settings at any time. Disabling cookies may affect how some parts of the site function.</p><h2>Children's data</h2><p>Our services are directed at businesses and business professionals. We do not knowingly collect personal data from anyone under the age of 16. If you believe we have inadvertently collected such data, contact us and we will delete it promptly.</p><h2>Changes to this policy</h2><p>We may update this policy from time to time, particularly as our tooling or services change. The date at the top of this page reflects when it was last revised. If you have an active engagement with us and we make significant changes, we will notify you directly.</p><h2>Contact us about this policy</h2><p>If you have any questions about this privacy policy or how we handle your data, please get in touch:</p><p>Email: <a href='mailto:agents@contactforsales.com'>agents@contactforsales.com</a><br>Phone: +44 (0) 7934 524 911</p>$aao_priv$;
  rh := '<section id="8d81e665-3ee0-443d-a873-690268c15fbb" class="section section--generic">'||E'\n'||'  <div class="container">'||E'\n'||'    <h2 class="section__title">'||hd||'</h2>'||E'\n'||'    <div class="section__content">'||ct||'</div>'||E'\n'||'  </div>'||E'\n'||'</section>';
  pg := gen_random_uuid(); pc := gen_random_uuid();
  INSERT INTO pages (id, site_id, name, url, title, page_type, status, build_status,
                     meta_description, nav_label, nav_order, in_header, in_footer,
                     sections, rebuild_policy, deployed_at)
  VALUES (pg, aao, 'privacy', '/privacy.html', 'Privacy Policy', 'content', 'active', 'deployed',
          'How AI Agent Orchestration collects, uses, and protects your personal data, and your rights under UK GDPR.',
          'Privacy Policy', 90, false, true,
          '["generic-text-block"]'::jsonb, 'owned', now());
  INSERT INTO page_components (id, page_id, slot_name, component_id, content_data, rendered_html,
                              build_status, position, locked_at, locked_by, lock_type)
  VALUES (pc, pg, 'generic-text-block', gtb,
          jsonb_build_object('heading', hd, 'content', ct), rh,
          'deployed', 0, now(), '182_legal_pages', 'permanent');
  INSERT INTO site_nav_items (site_id, group_id, label, url, page_id, item_type, position, status)
  VALUES (aao, aao_legal_g, 'Privacy Policy', '/privacy.html', pg, 'page_link', 0, 'active');
  INSERT INTO bak_legal_pages_20260721(what,id) VALUES ('aao_privacy_page', pg), ('aao_privacy_pc', pc);

  -- ════════════════════════ aao TERMS ════════════════════════
  hd := 'Terms of Service';
  ct := $aao_terms$<p><strong>Last updated: July 2026</strong></p><p>These terms govern your use of the AI Agent Orchestration website. By using this website, you accept these terms. If you do not accept them, please do not use the site.</p><h2>Who we are</h2><p>AI Agent Orchestration is a consultancy that designs, builds, and operates production-grade multi-agent AI systems on Kubernetes, Kafka, and Postgres. We are based in the United Kingdom. You can contact us at <a href='mailto:agents@contactforsales.com'>agents@contactforsales.com</a> or on +44 (0) 7934 524 911.</p><h2>Using this website</h2><p>You may use this website for lawful purposes and for your own informational use. You agree not to misuse it — for example by attempting to gain unauthorised access, disrupting its operation, or using it to break the law.</p><h2>Information on this website</h2><p>We take care to keep the information on this website accurate and up to date, but we provide it for general information only. It does not constitute professional, legal, financial, or technical advice, and you should not rely on it as such. Before acting on anything you read here, seek advice appropriate to your circumstances.</p><h2>Interactive tools and calculators</h2><p>This website may provide interactive tools, calculators, and estimators. These are provided for general guidance and illustration only. Their results depend on the information you enter and on assumptions built into them, and they may not reflect your specific circumstances. They are not a substitute for professional advice, a formal quotation, or a binding assessment, and we are not liable for decisions made in reliance on their output.</p><h2>Our services</h2><p>Any professional services we provide are governed by a separate written agreement between us and the client. These website terms do not form part of, or vary, any such agreement. Nothing on this website is an offer capable of acceptance or a binding quotation.</p><h2>Intellectual property</h2><p>The content, design, branding, and materials on this website are owned by AI Agent Orchestration or our licensors and are protected by intellectual property laws. You may view and print pages for your own reference. You may not reproduce, republish, or commercially exploit our content without our permission.</p><h2>Links to other sites</h2><p>This website may link to third-party websites. We provide those links for convenience and do not control or endorse the content of external sites. We are not responsible for them.</p><h2>No warranties</h2><p>We provide this website on an "as is" and "as available" basis. We do not warrant that it will be uninterrupted, error-free, or free of harmful components, or that any result obtained from it will be accurate or reliable.</p><h2>Limitation of liability</h2><p>To the fullest extent permitted by law, we are not liable for any loss or damage arising from your use of, or inability to use, this website or anything on it, including any interactive tool. Nothing in these terms excludes or limits our liability where it would be unlawful to do so — for example for death or personal injury caused by negligence, or for fraud.</p><h2>Changes to these terms</h2><p>We may update these terms from time to time. The date at the top of this page shows when they were last revised. By continuing to use the website after a change, you accept the revised terms.</p><h2>Governing law</h2><p>These terms are governed by the laws of England and Wales, and any dispute relating to them or to this website is subject to the jurisdiction of the courts of England and Wales.</p><h2>Contact us</h2><p>If you have any questions about these terms, please get in touch:</p><p>Email: <a href='mailto:agents@contactforsales.com'>agents@contactforsales.com</a><br>Phone: +44 (0) 7934 524 911</p>$aao_terms$;
  rh := '<section id="8d81e665-3ee0-443d-a873-690268c15fbb" class="section section--generic">'||E'\n'||'  <div class="container">'||E'\n'||'    <h2 class="section__title">'||hd||'</h2>'||E'\n'||'    <div class="section__content">'||ct||'</div>'||E'\n'||'  </div>'||E'\n'||'</section>';
  pg := gen_random_uuid(); pc := gen_random_uuid();
  INSERT INTO pages (id, site_id, name, url, title, page_type, status, build_status,
                     meta_description, nav_label, nav_order, in_header, in_footer,
                     sections, rebuild_policy, deployed_at)
  VALUES (pg, aao, 'terms', '/terms.html', 'Terms of Service', 'content', 'active', 'deployed',
          'The terms that govern your use of the AI Agent Orchestration website, including its interactive tools.',
          'Terms of Service', 91, false, true,
          '["generic-text-block"]'::jsonb, 'owned', now());
  INSERT INTO page_components (id, page_id, slot_name, component_id, content_data, rendered_html,
                              build_status, position, locked_at, locked_by, lock_type)
  VALUES (pc, pg, 'generic-text-block', gtb,
          jsonb_build_object('heading', hd, 'content', ct), rh,
          'deployed', 0, now(), '182_legal_pages', 'permanent');
  INSERT INTO site_nav_items (site_id, group_id, label, url, page_id, item_type, position, status)
  VALUES (aao, aao_legal_g, 'Terms of Service', '/terms.html', pg, 'page_link', 1, 'active');
  INSERT INTO bak_legal_pages_20260721(what,id) VALUES ('aao_terms_page', pg), ('aao_terms_pc', pc);

  -- ════════════════════════ finetuning TERMS ════════════════════════
  hd := 'Terms of Service';
  ct := $ft_terms$<p><strong>Last updated: July 2026</strong></p><p>These terms govern your use of the FineTuning website. By using this website, you accept these terms. If you do not accept them, please do not use the site.</p><h2>Who we are</h2><p>FineTuning is an AI systems and automation consultancy. We build AI automation pipelines, agent systems, custom models, and data infrastructure for businesses. We are based in the United Kingdom. You can contact us at <a href='mailto:finetune@contactforsales.com'>finetune@contactforsales.com</a> or on +44 (0) 7934 524 911.</p><h2>Using this website</h2><p>You may use this website for lawful purposes and for your own informational use. You agree not to misuse it — for example by attempting to gain unauthorised access, disrupting its operation, or using it to break the law.</p><h2>Information on this website</h2><p>We take care to keep the information on this website accurate and up to date, but we provide it for general information only. It does not constitute professional, legal, financial, or technical advice, and you should not rely on it as such. Before acting on anything you read here, seek advice appropriate to your circumstances.</p><h2>Interactive tools and calculators</h2><p>This website may provide interactive tools, calculators, and estimators. These are provided for general guidance and illustration only. Their results depend on the information you enter and on assumptions built into them, and they may not reflect your specific circumstances. They are not a substitute for professional advice, a formal quotation, or a binding assessment, and we are not liable for decisions made in reliance on their output.</p><h2>Our services</h2><p>Any professional services we provide are governed by a separate written agreement between us and the client. These website terms do not form part of, or vary, any such agreement. Nothing on this website is an offer capable of acceptance or a binding quotation.</p><h2>Intellectual property</h2><p>The content, design, branding, and materials on this website are owned by FineTuning or our licensors and are protected by intellectual property laws. You may view and print pages for your own reference. You may not reproduce, republish, or commercially exploit our content without our permission.</p><h2>Links to other sites</h2><p>This website may link to third-party websites. We provide those links for convenience and do not control or endorse the content of external sites. We are not responsible for them.</p><h2>No warranties</h2><p>We provide this website on an "as is" and "as available" basis. We do not warrant that it will be uninterrupted, error-free, or free of harmful components, or that any result obtained from it will be accurate or reliable.</p><h2>Limitation of liability</h2><p>To the fullest extent permitted by law, we are not liable for any loss or damage arising from your use of, or inability to use, this website or anything on it, including any interactive tool. Nothing in these terms excludes or limits our liability where it would be unlawful to do so — for example for death or personal injury caused by negligence, or for fraud.</p><h2>Changes to these terms</h2><p>We may update these terms from time to time. The date at the top of this page shows when they were last revised. By continuing to use the website after a change, you accept the revised terms.</p><h2>Governing law</h2><p>These terms are governed by the laws of England and Wales, and any dispute relating to them or to this website is subject to the jurisdiction of the courts of England and Wales.</p><h2>Contact us</h2><p>If you have any questions about these terms, please get in touch:</p><p>Email: <a href='mailto:finetune@contactforsales.com'>finetune@contactforsales.com</a><br>Phone: +44 (0) 7934 524 911</p>$ft_terms$;
  rh := '<section id="8d81e665-3ee0-443d-a873-690268c15fbb" class="section section--generic">'||E'\n'||'  <div class="container">'||E'\n'||'    <h2 class="section__title">'||hd||'</h2>'||E'\n'||'    <div class="section__content">'||ct||'</div>'||E'\n'||'  </div>'||E'\n'||'</section>';
  pg := gen_random_uuid(); pc := gen_random_uuid();
  INSERT INTO pages (id, site_id, name, url, title, page_type, status, build_status,
                     meta_description, nav_label, nav_order, in_header, in_footer,
                     sections, rebuild_policy, deployed_at)
  VALUES (pg, ft, 'terms', '/terms.html', 'Terms of Service', 'content', 'active', 'deployed',
          'The terms that govern your use of the FineTuning website, including its interactive tools.',
          'Terms of Service', 91, false, true,
          '["generic-text-block"]'::jsonb, 'owned', now());
  INSERT INTO page_components (id, page_id, slot_name, component_id, content_data, rendered_html,
                              build_status, position, locked_at, locked_by, lock_type)
  VALUES (pc, pg, 'generic-text-block', gtb,
          jsonb_build_object('heading', hd, 'content', ct), rh,
          'deployed', 0, now(), '182_legal_pages', 'permanent');
  INSERT INTO site_nav_items (site_id, group_id, label, url, page_id, item_type, position, status)
  VALUES (ft, ft_legal_g, 'Terms of Service', '/terms.html', pg, 'page_link', 1, 'active');
  INSERT INTO bak_legal_pages_20260721(what,id) VALUES ('ft_terms_page', pg), ('ft_terms_pc', pc);

  -- ── post-conditions ──
  IF (SELECT count(*) FROM pages WHERE site_id=aao AND url IN ('/privacy.html','/terms.html') AND status='active' AND build_status='deployed' AND rebuild_policy='owned') <> 2 THEN
    RAISE EXCEPTION 'post: aao legal pages not created as expected';
  END IF;
  IF (SELECT count(*) FROM pages WHERE site_id=ft AND url='/terms.html' AND status='active' AND rebuild_policy='owned') <> 1 THEN
    RAISE EXCEPTION 'post: finetuning terms not created as expected';
  END IF;
  IF (SELECT count(*) FROM site_nav_items WHERE group_id=aao_legal_g AND status='active') <> 2
     OR (SELECT count(*) FROM site_nav_items WHERE group_id=ft_legal_g AND status='active') <> 2 THEN
    RAISE EXCEPTION 'post: legal nav items not wired as expected (aao=2, ft=2)';
  END IF;
  -- every new page_component has non-empty rendered_html carrying its heading
  IF EXISTS (SELECT 1 FROM page_components WHERE locked_by='182_legal_pages'
             AND (rendered_html IS NULL OR rendered_html NOT LIKE '%section__content%')) THEN
    RAISE EXCEPTION 'post: a legal page_component has no rendered content';
  END IF;

  RAISE NOTICE 'legal-pages migration OK: aao_legal_group=%', aao_legal_g;
END $mig$;

INSERT INTO schema_migrations (filename, notes)
VALUES ('182_legal_pages_aao_finetuning.sql',
        'bugs_open/049 candidate 2: create legal pages so aao + finetuning can be chrome-refreshed without triggering 053. aao: /privacy.html + /terms.html + new legal nav group; finetuning: /terms.html + terms nav item in existing legal group. generic-text-block content pages, rebuild_policy=owned + permanent lock, verifiable-facts-only content mirroring finetuning''s approved privacy policy. Deploy separately via rerender-pages refresh_site_components:true.');

COMMIT;
