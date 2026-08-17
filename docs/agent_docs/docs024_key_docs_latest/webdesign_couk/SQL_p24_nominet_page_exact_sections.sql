-- SQL_p24_nominet_page_exact_sections.sql — webdesign.co.uk
--
-- REPLACE the /domains/ page's section copy with the exact Nominet disclosure
-- set, via the framework's own section-editor (item_type section_edit,
-- edit_type content_edit, literal field_updates — the loanandmortgagecalculator
-- index item is the completed fleet precedent).
--
-- WHY NOT ANOTHER WRITER ROUND. Round 3 (SQL_p23's needs_page) passed every
-- gate and produced copy that (a) stated "webdesign.co.uk doesn't register
-- domains itself" — the opposite of the page's purpose, (b) transmuted the
-- registered service-commitment VALUES into invented DNS/transfer timelines
-- (5wd ack became "5 working days to become fully active"; 10wd resolution
-- became "transfers can take up to ten working days", contradicting the
-- correct 2wd two sections earlier), and (c) omitted the postal address, the
-- complaints procedure and the abuse contact — three things the RRA REQUIRES
-- on the page (D.1.1, D.1.4, D.1.7). The numeric audit passed because the
-- NUMBERS matched the register; the MEANINGS did not. On a page whose every
-- sentence is a commitment to a regulator, the content path must be exact:
-- the operator supplies the copy, every sentence traceable to the D16 fact
-- register, and the framework validates/renders/deploys it.
--
-- Five items, one per component, distinct item_keys. They may run in
-- parallel; each updates its OWN component row, so the terminal state is
-- correct regardless of order. The deployed FILE is verified afterwards and a
-- single page_rerender re-assembles if an interleaved deploy read a stale
-- sibling row.

\set ON_ERROR_STOP on

BEGIN;

INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary,
                             spec, page_id, priority, handler_agent, status, created_by, item_key)
SELECT s.id, 'operator', 'build', 'section_edit', 'high',
       'Nominet page (D16): exact disclosure copy for section ' || v.pos || ' — ' || v.label,
       v.spec::jsonb, p.id, 30, 'section-editor', 'triaged', 'webdesign_couk_nominet_thread',
       'nominet-exact:domains:' || v.pos
FROM sites s
JOIN pages p ON p.site_id = s.id AND p.name = 'domains',
(VALUES
  (1, 'hero', $j$
  {
    "domain": "webdesign.co.uk",
    "edit_type": "content_edit",
    "page_name": "domains",
    "page_component_id": "acc9c5c9-55a1-4e84-a96f-ea393726cc89",
    "field_updates": {
      "headline": "Domain registration",
      "subheadline": "We register and look after .uk domain names for the people whose websites we build. This page sets out what it costs, how it works, and the commitments we make. Our Nominet registrar tag is DESIGNCONSULT."
    }
  }$j$),
  (2, 'who you are dealing with', $j$
  {
    "domain": "webdesign.co.uk",
    "edit_type": "content_edit",
    "page_name": "domains",
    "page_component_id": "9343491e-d627-45a0-aaf0-f1e2eb66fe2c",
    "field_updates": {
      "heading": "Who you are dealing with",
      "content": "<p>webdesign.co.uk is the trading name under which we provide domain registration. There is no separate company name.</p><ul><li>Postal address: 37 Fleetside, West Molesey, East Surrey KT8 2NF, United Kingdom</li><li>Telephone: +44 (0) 7934 524 911</li><li>Email: info@designconsultancy.co.uk</li></ul><p>You can use any of these to reach us about a domain name, whether or not you are already a customer. The details we collect to register a domain go to Nominet's register of .uk domains as the registrant record. They must be accurate, we will ask you to confirm them whenever anything changes, and we use them for nothing else.</p>"
    }
  }$j$),
  (3, 'costs and timescales', $j$
  {
    "domain": "webdesign.co.uk",
    "edit_type": "content_edit",
    "page_name": "domains",
    "page_component_id": "b6ccc8e1-8a6e-4152-a01c-5e8774460681",
    "field_updates": {
      "heading": "What it costs and how long it takes",
      "content": "<p>Registration of a .uk family domain (.uk, .co.uk, .org.uk or .me.uk) costs £10 per month. Renewal costs £10 per month, the same rate as registration. If our charges change, we tell you before the change takes effect and before any renewal.</p><p>Transferring your domain to another registrar is free of charge, and we do not charge anything when you end your contract with us. We do not obstruct a move.</p><ul><li>New registrations are normally completed within one working day of your instruction and payment.</li><li>Changes to an existing domain, such as nameservers or contact details, and transfers to another registrar, are actioned within two working days of your request.</li></ul><p>Domains we register for you are registered in your name, not ours, unless you explicitly ask otherwise. You are the registrant, and if you leave us the domain goes with you.</p>"
    }
  }$j$),
  (4, 'renewal, expiry, service commitments', $j$
  {
    "domain": "webdesign.co.uk",
    "edit_type": "content_edit",
    "page_name": "domains",
    "page_component_id": "5184e4f7-66b7-412d-aa08-f91f7b77fa3e",
    "field_updates": {
      "heading": "Renewal, expiry and our service commitments",
      "content": "<p>We send you an expiry notice no more than 30 days before your domain's renewal date. You can renew at any point up to the moment Nominet would otherwise cancel and delete the domain. If you choose not to renew, the registration expires and Nominet suspends and then deletes it under its published expiry process. We never let a registration lapse without having sent you notice first.</p><p>We acknowledge every message, including complaints and abuse reports, within five working days of receiving it, and usually within one. We aim to resolve the issues you raise within ten working days. Where something will take longer, we say so and keep you informed.</p>"
    }
  }$j$),
  (5, 'complaints, abuse, Nominet terms', $j$
  {
    "domain": "webdesign.co.uk",
    "edit_type": "content_edit",
    "page_name": "domains",
    "page_component_id": "c2025840-1b13-4c74-8768-35aa8012f16f",
    "field_updates": {
      "heading": "Complaints, abuse reports and Nominet's terms",
      "content": "<p>To complain about the service you have received from us, email info@designconsultancy.co.uk with Complaint in the subject line, or write to the postal address above. We acknowledge complaints within five working days, and our reply will tell you what we found and how to escalate it. If you remain unhappy, ask for the complaint to be reviewed again and we will look at it afresh. If it concerns a .uk domain and you are still not satisfied, you can take it to Nominet, the .uk registry, under its own complaints procedure at <a href=\"https://nominet.uk/complaints/\">nominet.uk/complaints/</a>, by telephone on +44 (0) 330 236 9470, or by email to domainsupport@nominet.uk.</p><p>If you believe a domain or website we manage is being used abusively, for phishing, malware or spam, email info@designconsultancy.co.uk with Abuse report in the subject line. Anyone may report abuse, customer or not. We acknowledge abuse reports within five working days.</p><p>Every .uk registration is also a contract between you, the registrant, and Nominet, under Nominet's Terms and Conditions of Domain Name Registration, published in its .UK policy library at <a href=\"https://nominet.uk/uk-registry/uk-policy/\">nominet.uk/uk-registry/uk-policy/</a>. We make sure you have seen those terms before we register a domain for you and before each renewal.</p>"
    }
  }$j$)
) AS v(pos, label, spec)
WHERE s.domain = 'webdesign.co.uk'
ON CONFLICT DO NOTHING;

DO $verify$
DECLARE v_site uuid; v_items int;
BEGIN
    SELECT id INTO v_site FROM sites WHERE domain = 'webdesign.co.uk';
    SELECT count(*) INTO v_items FROM site_work_items
     WHERE site_id = v_site AND item_key LIKE 'nominet-exact:domains:%'
       AND status NOT IN ('complete','cancelled','rejected');
    IF v_items <> 5 THEN RAISE EXCEPTION 'expected 5 open section_edit items, found %', v_items; END IF;
    RAISE NOTICE 'SQL_p24 verified: 5 section_edit items triaged for section-editor';
END $verify$;

COMMIT;
