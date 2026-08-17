-- SQL_p23_nominet_domains_page.sql — webdesign.co.uk
--
-- Deliver the Nominet registrar disclosures page (/domains/index.html) THROUGH
-- THE FRAMEWORK, per owner ruling D16 (PLAN §D16, 2026-08-17). This supersedes
-- the hand-spliced file the same session built first (untracked in
-- ~/projects/sites, now the content reference only, never committed).
--
-- WHAT THIS DOES, and why each piece:
--   1. Creates the site's evidence_base spec (it had none) carrying every fact
--      the page must state, all attested by the owner on 2026-08-17 in this
--      session, with a hand-written writer_block (writer_block_managed absent
--      = unmanaged, the default). The claims layer keys on this row existing
--      WITH facts — a writer_block-only row is the documented silent no-op
--      (validate_page_content_stats.go header; fabricated-stats landmine).
--   2. Registers the page in the AUTHORITATIVE plan store (site_plan_pages +
--      site_plan_sections on the current plan 07ac4459…, which holds only
--      'index'), with assigned_fact_ids per section — the RFC_016/151 path:
--      load_page_sections_from_spec tier 1 emits section_facts, and
--      page-build-handler wires it (checked live 2026-08-17).
--   3. Creates the pages row (build_status='planned', rebuild_policy='generic'
--      — the FRAMEWORK owns this page) with page_spec.purpose, which
--      save_page_sections reads into every section's content_brief.
--   4. Files the needs_page item for page-build-handler (SQL_p18 pattern —
--      nothing watches pages for planned rows; item_key follows the
--      reconciler's convention so it co-dedups).
--
-- Component choice: hero + 4x generic-text-block, NOTHING that {{range}}s over
-- a writer-supplied array (bugs_open/260: a mistyped nested array parks the
-- build at needs_human_review) and NO call-to-action (the news build shipped
-- empty CTA buttons, SQL_p19; contact details close the last text block
-- instead).
--
-- Facts with small numeric values (1, 2, 5, 10, 30) all carry context_terms so
-- they support only their own claim windows, never every small number on the
-- site. Two facts share value 10 (price, resolution target); the terms keep
-- them apart.

\set ON_ERROR_STOP on

BEGIN;

-- 1 ── evidence base ────────────────────────────────────────────────────────
INSERT INTO site_specs (site_id, aspect, data, source, is_current, created_by, notes)
SELECT s.id, 'evidence_base', $eb$
{
  "audit_doc": "docs/agent_docs/docs024_key_docs_latest/webdesign_couk/PLAN_2026-07-25_webdesign_couk.md §D16 (owner ruling 2026-08-17). Requirements source: Nominet .UK Registry-Registrar Agreement 20.03.2024, Schedule D.1.1-D.1.7, Key Terms definition, B.1.8/B.1.9/B.1.10/B.1.13.",
  "governing_rule": "This register backs the Nominet registrar page. Every commitment on that page is one the business must keep, and Nominet reviews the page as part of a tag application. State each fact exactly as its claim reads. Do not invent a price, a timescale, a guarantee or a policy. If a fact is not in this register, the page does not say it.",
  "facts": [
    {
      "id": "trading_name",
      "kind": "attestation",
      "claim": "The domain registration service on this site is provided under the trading name webdesign.co.uk. There is no separate company name.",
      "source": {"attested_by": "owner, 2026-08-17, webdesign_couk PLAN §D16"},
      "verified_at": "2026-08-17",
      "writer_line": "webdesign.co.uk"
    },
    {
      "id": "postal_address",
      "kind": "attestation",
      "claim": "Our postal address is 37 Fleetside, West Molesey, East Surrey KT8 2NF, United Kingdom.",
      "source": {"attested_by": "owner, 2026-08-17, webdesign_couk PLAN §D16 (house number added by owner the same day)"},
      "verified_at": "2026-08-17",
      "writer_line": "37 Fleetside, West Molesey, East Surrey KT8 2NF"
    },
    {
      "id": "phone",
      "kind": "attestation",
      "claim": "Our telephone number is +44 (0) 7934 524 911.",
      "source": {"attested_by": "owner, 2026-08-17, webdesign_couk PLAN §D16"},
      "verified_at": "2026-08-17",
      "writer_line": "+44 (0) 7934 524 911"
    },
    {
      "id": "contact_email",
      "kind": "attestation",
      "claim": "Our email address for all domain matters, including complaints and abuse reports, is info@designconsultancy.co.uk. Complaints should carry Complaint in the subject line and abuse reports should carry Abuse report. Anyone may report abuse, customer or not.",
      "source": {"attested_by": "owner, 2026-08-17, webdesign_couk PLAN §D16"},
      "verified_at": "2026-08-17",
      "writer_line": "info@designconsultancy.co.uk"
    },
    {
      "id": "nominet_tag",
      "kind": "attestation",
      "claim": "Our Nominet registrar tag is DESIGNCONSULT, with Channel Partner classification. We provide .uk domain registration under Nominet's Registry-Registrar Agreement.",
      "source": {"attested_by": "owner, 2026-08-17, webdesign_couk PLAN §D16"},
      "verified_at": "2026-08-17",
      "writer_line": "DESIGNCONSULT"
    },
    {
      "id": "price_monthly",
      "kind": "metric",
      "claim": "Registration of a .uk family domain name (.uk, .co.uk, .org.uk, .me.uk) costs ten pounds per month, and renewal costs the same ten pounds per month. Renewal is charged at the same rate as registration.",
      "value": 10,
      "tolerance": "exact",
      "context_terms": ["month", "pound", "price", "cost", "registration", "renewal", "charge"],
      "source": {"attested_by": "owner, 2026-08-17, webdesign_couk PLAN §D16 (price and renewal price stated separately, same figure)"},
      "verified_at": "2026-08-17",
      "writer_line": "£10 per month"
    },
    {
      "id": "transfer_free",
      "kind": "attestation",
      "claim": "Transferring a domain name to another registrar is free of charge. We do not charge anything when a customer ends their contract with us, and we do not obstruct the move.",
      "source": {"attested_by": "owner, 2026-08-17, webdesign_couk PLAN §D16 (owner chose free for this offering; the £150 buyout belongs to the webdesign.uk build service, not this page)"},
      "verified_at": "2026-08-17",
      "writer_line": "free of charge"
    },
    {
      "id": "registration_1wd",
      "kind": "metric",
      "claim": "New registrations are normally completed within one working day of instruction and payment.",
      "value": 1,
      "tolerance": "exact",
      "context_terms": ["working day", "registration", "registrations", "register"],
      "source": {"attested_by": "owner-approved commitment set, 2026-08-17, webdesign_couk PLAN §D16"},
      "verified_at": "2026-08-17",
      "writer_line": "one working day"
    },
    {
      "id": "changes_2wd",
      "kind": "metric",
      "claim": "Changes to an existing domain, such as nameservers or contact details, and transfers to another registrar, are actioned within two working days of the request.",
      "value": 2,
      "tolerance": "exact",
      "context_terms": ["working days", "changes", "transfer", "nameserver"],
      "source": {"attested_by": "owner-approved commitment set, 2026-08-17, webdesign_couk PLAN §D16"},
      "verified_at": "2026-08-17",
      "writer_line": "two working days"
    },
    {
      "id": "customer_name",
      "kind": "attestation",
      "claim": "Domains registered for a customer are registered in the customer's name, not ours, unless the customer explicitly asks otherwise. The customer is the registrant, and if they leave the domain goes with them.",
      "source": {"attested_by": "owner, 2026-08-17, webdesign_couk PLAN §D16; RRA B.1.8 obligation"},
      "verified_at": "2026-08-17",
      "writer_line": "registered in your name"
    },
    {
      "id": "expiry_notice_30d",
      "kind": "metric",
      "claim": "We send an expiry notice no more than 30 days before a domain's renewal date. A customer can renew at any point up to the moment Nominet would otherwise cancel and delete the domain. If they choose not to renew, the registration expires and Nominet suspends and then deletes it under its published expiry process, and we never let a registration lapse without having sent notice first.",
      "value": 30,
      "tolerance": "exact",
      "context_terms": ["days", "expiry", "notice", "renewal date"],
      "source": {"attested_by": "owner-approved commitment set, 2026-08-17, webdesign_couk PLAN §D16; RRA B.1.13 bounds"},
      "verified_at": "2026-08-17",
      "writer_line": "30 days"
    },
    {
      "id": "ack_5wd",
      "kind": "metric",
      "claim": "We acknowledge every message, including complaints and abuse reports, within five working days of receipt, and usually within one.",
      "value": 5,
      "tolerance": "exact",
      "context_terms": ["working days", "acknowledge", "complaints", "abuse"],
      "source": {"attested_by": "owner-approved commitment set, 2026-08-17, webdesign_couk PLAN §D16; RRA D.1.3/D.1.7 maxima"},
      "verified_at": "2026-08-17",
      "writer_line": "five working days"
    },
    {
      "id": "resolve_10wd",
      "kind": "metric",
      "claim": "We aim to resolve the issues customers raise within ten working days, and where something will take longer we say so and keep the customer informed.",
      "value": 10,
      "tolerance": "exact",
      "context_terms": ["working days", "resolve", "issues"],
      "source": {"attested_by": "owner-approved commitment set, 2026-08-17, webdesign_couk PLAN §D16; RRA D.1.2"},
      "verified_at": "2026-08-17",
      "writer_line": "ten working days"
    },
    {
      "id": "nominet_terms",
      "kind": "attestation",
      "claim": "Every .uk domain registration is also a contract between the registrant and Nominet under Nominet's Terms and Conditions of Domain Name Registration, published in Nominet's .UK policy library at nominet.uk/uk-registry/uk-policy/. Customers are made aware of those terms before registration and before each renewal.",
      "source": {"attested_by": "owner-approved, 2026-08-17; RRA B.1.9; URL verified by search 2026-08-17 (nominet.uk 403s curl, do not re-verify that way)"},
      "verified_at": "2026-08-17",
      "writer_line": "nominet.uk/uk-registry/uk-policy/"
    },
    {
      "id": "nominet_escalation",
      "kind": "attestation",
      "claim": "A customer who is not satisfied with the outcome of a complaint about a .uk domain can escalate it to Nominet, the .uk registry, at nominet.uk/complaints/, by telephone on +44 (0) 330 236 9470, or by email to domainsupport@nominet.uk.",
      "source": {"attested_by": "owner-approved, 2026-08-17; RRA D.1.4; contact details from Nominet's published complaints page, found by search 2026-08-17"},
      "verified_at": "2026-08-17",
      "writer_line": "nominet.uk/complaints/"
    }
  ],
  "banned_claims": [
    {
      "pattern": "accredited channel partner",
      "reason": "The tag classification is Channel Partner, not Accredited. Claiming Accredited status would breach RRA B.1.11 (misleading claims about tag classification)."
    }
  ],
  "writer_block": "HOW TO WRITE ABOUT THE DOMAIN REGISTRATION SERVICE, AND WHAT MUST NOT BE SAID.\n\nNever use an em dash anywhere. Use a full stop, a comma, a colon or brackets instead. Plain British English throughout. Do not use the word honest or honestly. Do not oversell: no superlatives, no client counts, no years of experience.\n\nThis page exists to satisfy Nominet's registrar requirements, so every commitment on it is one the business must keep. State the facts in the register exactly. Do not round a number, soften a timescale, or add a guarantee the register does not carry.\n\nThe only prices you may state: registration of a .uk family domain (.uk, .co.uk, .org.uk, .me.uk) costs £10 per month, and renewal costs £10 per month, the same rate. Transferring a domain to another registrar is free of charge, and ending the contract costs nothing. Phrase these with 'free of charge' or 'we do not charge'; never leave a bare 'no' as the only negation in the sentence.\n\nThe contact details, exactly as registered: trading name webdesign.co.uk, postal address 37 Fleetside, West Molesey, East Surrey KT8 2NF, United Kingdom, telephone +44 (0) 7934 524 911, email info@designconsultancy.co.uk. The same email takes complaints and abuse reports; say to put Complaint or Abuse report in the subject line, and that anyone may report abuse, customer or not.\n\nThe service commitments, exactly as registered: acknowledge every message including complaints and abuse reports within five working days and usually within one; aim to resolve issues within ten working days and say so when something will take longer; send an expiry notice no more than 30 days before the renewal date; allow renewal up to the moment Nominet would cancel and delete the domain; complete new registrations normally within one working day; action changes and transfers within two working days; register domains in the customer's name unless they explicitly ask otherwise.\n\nNominet references: the registrar tag is DESIGNCONSULT, Channel Partner classification. Nominet's Terms and Conditions of Domain Name Registration are in its .UK policy library at nominet.uk/uk-registry/uk-policy/. A complaint we cannot resolve can go to Nominet at nominet.uk/complaints/, telephone +44 (0) 330 236 9470, email domainsupport@nominet.uk.\n\nNever claim Accredited Channel Partner status. Never name another registrar. Never publish a comparative ranking of named agencies (standing site rail)."
}
$eb$::jsonb, 'operator', true, 'webdesign_couk_nominet_thread',
'Nominet registrar page register (owner ruling D16, 2026-08-17). First evidence_base on this site; created for the /domains/ page and any future build that states these facts.'
FROM sites s
WHERE s.domain = 'webdesign.co.uk'
  AND NOT EXISTS (SELECT 1 FROM site_specs x WHERE x.site_id = s.id AND x.aspect = 'evidence_base' AND x.is_current);

-- 2 ── authoritative plan registration ──────────────────────────────────────
INSERT INTO site_plan_pages (plan_id, name, role, slug, url, in_header, in_footer, nav_order, title, meta_description, nav_label)
SELECT sp.id, 'domains', 'content', 'domains', '/domains/index.html', false, true, 90,
       'Domain registration | webdesign.co.uk',
       'How we register and manage .uk domain names: prices, renewal and expiry policy, service commitments, complaints procedure and abuse contact.',
       'Domain registration'
FROM site_plans sp JOIN sites s ON s.id = sp.site_id
WHERE s.domain = 'webdesign.co.uk' AND sp.is_current
  AND NOT EXISTS (SELECT 1 FROM site_plan_pages x WHERE x.plan_id = sp.id AND x.name = 'domains');

INSERT INTO site_plan_sections (plan_id, page_name, ordering, component_name, assigned_fact_ids)
SELECT sp.id, 'domains', v.ord, v.comp, v.facts::jsonb
FROM site_plans sp JOIN sites s ON s.id = sp.site_id,
     (VALUES
       (0, 'hero',               '["nominet_tag","trading_name"]'),
       (1, 'generic-text-block', '["trading_name","postal_address","phone","contact_email"]'),
       (2, 'generic-text-block', '["price_monthly","transfer_free","registration_1wd","changes_2wd","customer_name"]'),
       (3, 'generic-text-block', '["expiry_notice_30d","ack_5wd","resolve_10wd"]'),
       (4, 'generic-text-block', '["nominet_escalation","contact_email","nominet_terms","nominet_tag"]')
     ) AS v(ord, comp, facts)
WHERE s.domain = 'webdesign.co.uk' AND sp.is_current
  AND NOT EXISTS (SELECT 1 FROM site_plan_sections x WHERE x.plan_id = sp.id AND x.page_name = 'domains');

-- 3 ── the planned pages row (framework-owned: rebuild_policy generic) ──────
INSERT INTO pages (site_id, name, url, title, page_type, status, meta_description,
                   sections, rebuild_policy, build_status, in_header, in_footer,
                   nav_label, nav_order, page_spec)
SELECT s.id, 'domains', '/domains/index.html', 'Domain registration | webdesign.co.uk',
       'content', 'active',
       'How we register and manage .uk domain names: prices, renewal and expiry policy, service commitments, complaints procedure and abuse contact.',
       '["hero","generic-text-block","generic-text-block","generic-text-block","generic-text-block"]'::jsonb,
       'generic', 'planned', false, true, 'Domain registration', 90,
       jsonb_build_object('purpose',
         'This page satisfies Nominet''s registrar website requirements for the DESIGNCONSULT tag (Channel Partner classification). It must state, in plain British English, using the site evidence register exactly: who provides the service and the full contact details (trading name, postal address, telephone, email); what registration and renewal of .uk family domains cost and that transferring away and ending the contract are free of charge; how long registrations, changes and transfers take; that domains are registered in the customer''s name; the renewal and expiry policy including the 30 day expiry notice; the customer service commitments (five working day acknowledgement, ten working day resolution target); how to complain and how to escalate an unresolved complaint to Nominet; how to report abuse; and that registrations are also subject to Nominet''s own Terms and Conditions of Domain Name Registration. Nothing beyond the register; no invented commitments; never use an em dash.')
FROM sites s
WHERE s.domain = 'webdesign.co.uk'
  AND NOT EXISTS (SELECT 1 FROM pages x WHERE x.site_id = s.id AND x.name = 'domains');

-- 4 ── the build item (SQL_p18 pattern) ─────────────────────────────────────
INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary,
    spec, page_id, priority, handler_agent, status, created_by, item_key
)
SELECT s.id, 'operator', 'build', 'needs_page', 'medium',
       'Build Nominet registrar page /domains/index.html (owner ruling D16; evidence register + plan rows seeded by SQL_p23)',
       jsonb_build_object('reason', 'not_built', 'page_name', 'domains'),
       p.id, 40, 'page-build-handler', 'triaged', 'webdesign_couk_nominet_thread',
       'needs_page:domains'
  FROM sites s
  JOIN pages p ON p.site_id = s.id AND p.name = 'domains'
 WHERE s.domain = 'webdesign.co.uk'
ON CONFLICT DO NOTHING;

-- 5 ── verify ───────────────────────────────────────────────────────────────
DO $verify$
DECLARE v_site uuid; v_plan uuid; v_facts int; v_secs int; v_pages int; v_items int; v_planpages int;
BEGIN
    SELECT id INTO v_site FROM sites WHERE domain = 'webdesign.co.uk';
    SELECT id INTO v_plan FROM site_plans WHERE site_id = v_site AND is_current;

    SELECT jsonb_array_length(data->'facts') INTO v_facts FROM site_specs
     WHERE site_id = v_site AND aspect = 'evidence_base' AND is_current;
    IF v_facts IS DISTINCT FROM 15 THEN RAISE EXCEPTION 'expected 15 facts, found %', v_facts; END IF;

    SELECT count(*) INTO v_planpages FROM site_plan_pages WHERE plan_id = v_plan AND name = 'domains';
    IF v_planpages <> 1 THEN RAISE EXCEPTION 'expected 1 plan page, found %', v_planpages; END IF;

    SELECT count(*) INTO v_secs FROM site_plan_sections
     WHERE plan_id = v_plan AND page_name = 'domains' AND assigned_fact_ids IS NOT NULL;
    IF v_secs <> 5 THEN RAISE EXCEPTION 'expected 5 plan sections with facts, found %', v_secs; END IF;

    SELECT count(*) INTO v_pages FROM pages
     WHERE site_id = v_site AND name = 'domains' AND build_status = 'planned'
       AND rebuild_policy = 'generic' AND in_footer AND NOT in_header
       AND page_spec->>'purpose' IS NOT NULL;
    IF v_pages <> 1 THEN RAISE EXCEPTION 'expected 1 planned domains page, found %', v_pages; END IF;

    SELECT count(*) INTO v_items FROM site_work_items
     WHERE site_id = v_site AND item_key = 'needs_page:domains'
       AND status NOT IN ('complete','cancelled','rejected');
    IF v_items <> 1 THEN RAISE EXCEPTION 'expected exactly 1 open needs_page:domains item, found %', v_items; END IF;

    RAISE NOTICE 'SQL_p23 verified: 15 facts, plan page + 5 fact-assigned sections, planned pages row (footer only), needs_page:domains triaged';
END $verify$;

COMMIT;
