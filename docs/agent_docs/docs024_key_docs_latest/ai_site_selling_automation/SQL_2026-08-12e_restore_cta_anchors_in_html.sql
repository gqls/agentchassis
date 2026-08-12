-- FILE: SQL_2026-08-12e_restore_cta_anchors_in_html.sql
--
-- REGRESSION REPAIR, part 2 of 2. Part 1 (SQL_2026-08-12d) restored the CTA
-- destinations into `content_data`. It was necessary and NOT sufficient: a
-- `page_rerender` dispatched AFTER that repair still rendered no buttons.
--
-- WHY content_data ALONE COULD NOT WORK, read from the code and then confirmed
-- empirically. These fields are declared in `content_components.input_schema`
-- with `source: "renderer"`, and `sourceResolver.resolve` short-circuits that
-- source: `if source == "" || source == "llm" || source == "renderer" ||
-- source == "static" { return nil, true }` — value nil, **found true**. So the
-- field is never "missing", `handleMissingField` never runs, and therefore
-- `carryStored` (the bugs_open/238 carry, PBP-039) never runs either: the carry
-- guards fields that FAIL to resolve, and a renderer-sourced field always
-- "succeeds" with nothing. plan_sections then `continue`s on that branch,
-- writing only a `fallback` if the schema declares one. These do not.
--
-- The carry IS live in the running binary — checked at the artefact, not
-- inferred: agent-chassis v1.0.1291 was built from da5a7eb8f (image label
-- org.opencontainers.image.revision) and `git merge-base --is-ancestor
-- d26c26a9a da5a7eb8f` passes, with controls both ways. So this is a GAP in the
-- carry's coverage, not an unshipped fix. That claim is filed for independent
-- diagnosis rather than asserted here (see NOTES); this file only repairs.
--
-- WHAT THIS DOES. Renders the two anchors each component is missing, exactly as
-- its own template would, and splices them into `rendered_html` at the point
-- the template left empty. The labels and URLs are read from `content_data`
-- itself — nothing is retyped, so the button text cannot drift from the data.
--
-- HAND-PATCHING rendered_html IS A KNOWN COST, taken deliberately. It creates
-- the `page_divergence_overwritten` condition (bugs_open/229): the next rebuild
-- will overwrite these bytes and file a work item saying so. That is exactly
-- what happened to the hand-patched index hero this afternoon — the sibling
-- lane had restored those same two buttons by hand on 08-11 and my regeneration
-- destroyed them. The alternative is a sales site whose four selling pages have
-- no call-to-action buttons, and the durable fix is a platform change that is
-- not this lane's to make today. It is the same route this lane's sibling has
-- already taken twice on this site, for these same CTAs.
--
-- NOT TOUCHED: index/call-to-action. It carried NO links BEFORE the migration
-- either — its labels have never had URLs. Repairing it here would improve the
-- page and, in doing so, hide the boundary of what my migration actually broke.
--
-- ROLLBACK: the pre-patch bytes are in page_component_history via the artefact
-- archive trigger.

BEGIN;

-- Heroes: the template emits the buttons between the subheadline </p> and the
-- closing </div> of .hero-content, gated on {{if and .cta_text .cta_url}}.
UPDATE page_components pc SET
  rendered_html = regexp_replace(
    pc.rendered_html,
    '(</p>)(\s*)(</div>)',
    '\1' ||
    '<a href="' || (pc.content_data->>'cta_url') || '" class="btn btn-primary">' || (pc.content_data->>'cta_text') || '</a>' ||
    ' <a href="' || (pc.content_data->>'secondary_cta_url') || '" class="btn btn-secondary">' || (pc.content_data->>'secondary_cta') || '</a>' ||
    '\3'
  ),
  updated_at = now()
FROM pages p
WHERE pc.page_id = p.id
  AND p.site_id = '1fcfa4f3-ec80-4010-878b-b971cd46711f'
  AND p.name IN ('index','faq','how-it-works','what-you-get')
  AND pc.slot_name = 'hero'
  AND pc.locked_at IS NULL
  AND pc.content_data ? 'cta_url' AND pc.content_data ? 'secondary_cta_url'
  AND pc.rendered_html !~ 'class="btn btn-primary"';       -- idempotent

-- Call-to-action blocks: the buttons live inside <div class="cta-buttons">.
UPDATE page_components pc SET
  rendered_html = regexp_replace(
    pc.rendered_html,
    '(<div class="cta-buttons">)(\s*)(</div>)',
    '\1' ||
    '<a href="' || (pc.content_data->>'primary_cta_url') || '" class="cta-btn cta-btn-primary">' || (pc.content_data->>'primary_cta') || '</a>' ||
    ' <a href="' || (pc.content_data->>'secondary_cta_url') || '" class="cta-btn cta-btn-secondary">' || (pc.content_data->>'secondary_cta') || '</a>' ||
    '\3'
  ),
  updated_at = now()
FROM pages p
WHERE pc.page_id = p.id
  AND p.site_id = '1fcfa4f3-ec80-4010-878b-b971cd46711f'
  AND p.name IN ('faq','how-it-works','what-you-get')       -- index/cta deliberately excluded
  AND pc.slot_name = 'call-to-action'
  AND pc.locked_at IS NULL
  AND pc.content_data ? 'primary_cta_url' AND pc.content_data ? 'secondary_cta_url'
  AND pc.rendered_html !~ 'class="cta-btn cta-btn-primary"';

DO $$
DECLARE n_anchors int; n_empty_href int; n_components int;
BEGIN
  SELECT count(*), sum((SELECT count(*) FROM regexp_matches(pc.rendered_html,'href="','g')))
    INTO n_components, n_anchors
    FROM page_components pc JOIN pages p ON p.id=pc.page_id
   WHERE p.site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
     AND p.name IN ('index','faq','how-it-works','what-you-get')
     AND pc.slot_name IN ('hero','call-to-action');

  -- 7 repaired components x 2 anchors = 14. index/call-to-action stays at 0.
  IF n_anchors <> 14 THEN
    RAISE EXCEPTION 'expected 14 anchors across the 8 components, got %', n_anchors;
  END IF;

  -- An empty href is the exact failure this repairs; it must not survive it.
  SELECT count(*) INTO n_empty_href
    FROM page_components pc JOIN pages p ON p.id=pc.page_id
   WHERE p.site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
     AND pc.rendered_html LIKE '%href=""%';
  IF n_empty_href <> 0 THEN
    RAISE EXCEPTION 'components still carrying an empty href: %', n_empty_href;
  END IF;
END $$;

COMMIT;
