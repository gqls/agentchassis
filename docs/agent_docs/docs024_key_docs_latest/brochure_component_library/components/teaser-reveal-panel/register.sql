\set ON_ERROR_STOP on
-- teaser-reveal-panel — a panel of teasers that open in place at a shareable URL.
--
-- Source of truth is components/teaser-reveal-panel/{template.html,input_schema.json,
-- behaviour.js}; edit those and re-apply. Proven by
-- scripts/render_teaser_reveal_panel.go (14 checks, non-vacuous: two mutants fail
-- exactly the checks they should).
--
-- It implements the EXISTING experience pattern `teaser-detail-deeplink`
-- (experience_patterns, kind micro-journey) rather than naming a new shape. The
-- pattern's own absence rule is the load-bearing part: an item with no body
-- degrades to a plain statement with no control, never to a dead one.
--
-- Two deliberate departures from the vonc implementation of the same shape:
--   1. The reveal is native <details>/<summary>, so it works with zero JS. The
--      JS snippet only adds URL addressability. vonc's contract says the detail
--      region must be EMPTIED on close; that requirement exists because its
--      region is JS-populated and innerText on a display:none node falls back to
--      textContent, so a check could pass without the interaction. With
--      <details>, the `open` attribute is the honest signal and the body must
--      stay in the DOM.
--   2. The body stays in the DOM on purpose: it is assertive prose, so the
--      claims gate and crawlers must both be able to read it. Text hidden behind
--      JS leaves the verification net (the same trap as text inside <svg>).
BEGIN;
INSERT INTO content_components
  (id, name, function, display_name, description, category, semantic_tags,
   section_type, component_level, render_mode, is_dark_section, is_active,
   suitable_site_types, suitable_page_types, html_template, input_schema)
VALUES (
  gen_random_uuid(),
  'teaser-reveal-panel','teaser-reveal-panel','Teaser Reveal Panel',
  'A panel of short teasers, each a hook plus a deliberately unfinished continuation, whose full text opens in place without a page load and gains a shareable URL. Swipeable on mobile via native scroll-snap; the reveal itself is native <details> so it works with JavaScript disabled. An item with no full text renders as a plain statement with no control rather than a dead one. Implements the teaser-detail-deeplink experience pattern.',
  'content','["teaser","carousel","reveal","progressive-disclosure","brochure","micro-journey","teaser-detail-deeplink"]'::jsonb,
  'teaser-reveal-panel','section','agent',false,true,
  '["brochure","consultancy","professional-services","b2b"]'::jsonb,
  '["index","home","about","capabilities","landing","content"]'::jsonb,
  $HTML$__TEMPLATE__$HTML$,
  $SCHEMA$__SCHEMA__$SCHEMA$::jsonb
);
COMMIT;
