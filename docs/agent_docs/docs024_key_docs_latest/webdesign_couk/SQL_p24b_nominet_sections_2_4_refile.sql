-- SQL_p24b_nominet_sections_2_4_refile.sql — webdesign.co.uk
--
-- p24's sections 2 and 4 were refused by the SLOT FLOOR (visible-text shrink
-- guard, floor 50% of the existing slot's stripped text): 1062→512 (48%) and
-- 1317→579 (44%). The floor is step-config only (apply_edit whitelists
-- input_fields, so no per-item override), and loosening the FLEET step config
-- for one page would be the wrong trade. Remedy: the replacement copy gains
-- register-consistent sentences with real disclosure value, clearing the floor
-- with margin (~647 and ~817 visible chars). Items 1/3/5 completed and are not
-- re-filed.

\set ON_ERROR_STOP on

BEGIN;

INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary,
                             spec, page_id, priority, handler_agent, status, created_by, item_key)
SELECT s.id, 'operator', 'build', 'section_edit', 'high',
       'Nominet page (D16): exact disclosure copy, refile past slot floor — section ' || v.pos,
       v.spec::jsonb, p.id, 30, 'section-editor', 'triaged', 'webdesign_couk_nominet_thread',
       'nominet-exact:domains:' || v.pos || 'b'
FROM sites s
JOIN pages p ON p.site_id = s.id AND p.name = 'domains',
(VALUES
  (2, $j$
  {
    "domain": "webdesign.co.uk",
    "edit_type": "content_edit",
    "page_name": "domains",
    "page_component_id": "9343491e-d627-45a0-aaf0-f1e2eb66fe2c",
    "field_updates": {
      "heading": "Who you are dealing with",
      "content": "<p>webdesign.co.uk is the trading name under which we provide domain registration. There is no separate company name. We are the registrar of record for domains registered through this site, and Nominet lists our registrar tag DESIGNCONSULT against them.</p><ul><li>Postal address: 37 Fleetside, West Molesey, East Surrey KT8 2NF, United Kingdom</li><li>Telephone: +44 (0) 7934 524 911</li><li>Email: info@designconsultancy.co.uk</li></ul><p>You can use any of these to reach us about a domain name, whether or not you are already a customer. The details we collect to register a domain go to Nominet's register of .uk domains as the registrant record. They must be accurate, we will ask you to confirm them whenever anything changes, and we use them for nothing else.</p>"
    }
  }$j$),
  (4, $j$
  {
    "domain": "webdesign.co.uk",
    "edit_type": "content_edit",
    "page_name": "domains",
    "page_component_id": "5184e4f7-66b7-412d-aa08-f91f7b77fa3e",
    "field_updates": {
      "heading": "Renewal, expiry and our service commitments",
      "content": "<p>We send you an expiry notice no more than 30 days before your domain's renewal date. Renewal is charged at the same rate as registration, ten pounds per month, so there is never a renewal premium. You can renew at any point up to the moment Nominet would otherwise cancel and delete the domain. If you choose not to renew, the registration expires and Nominet suspends and then deletes it under its published expiry process. While a domain is suspended it stops resolving, which means the website and any email on it stop working, so renewing on schedule is worth it. We never let a registration lapse without having sent you notice first.</p><p>We acknowledge every message, including complaints and abuse reports, within five working days of receiving it, and usually within one. We aim to resolve the issues you raise within ten working days. Where something will take longer, we say so and keep you informed.</p>"
    }
  }$j$)
) AS v(pos, spec)
WHERE s.domain = 'webdesign.co.uk'
ON CONFLICT DO NOTHING;

DO $verify$
DECLARE v_items int;
BEGIN
    SELECT count(*) INTO v_items FROM site_work_items swi JOIN sites s ON s.id=swi.site_id
     WHERE s.domain='webdesign.co.uk' AND swi.item_key IN ('nominet-exact:domains:2b','nominet-exact:domains:4b')
       AND swi.status NOT IN ('complete','cancelled','rejected');
    IF v_items <> 2 THEN RAISE EXCEPTION 'expected 2 open refile items, found %', v_items; END IF;
    RAISE NOTICE 'SQL_p24b verified: sections 2b and 4b triaged';
END $verify$;

COMMIT;
