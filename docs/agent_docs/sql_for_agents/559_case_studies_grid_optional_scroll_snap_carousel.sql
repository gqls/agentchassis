-- 559_case_studies_grid_optional_scroll_snap_carousel.sql
--
-- Gives `case-studies-grid` an OPTIONAL card carousel, and turns it on for
-- ai-agent-orchestration.com only. Owner asked for carousels on this site.
--
-- ⚠ THE HANDOFF'S PLAN FOR THIS WAS WRONG AND THIS FILE IS THE CORRECTION.
-- `HANDOFF_2026-08-18_continue_here.md` §5 says the carousel work is "APPROVE +
-- BIND, not design", because two carousel contracts already exist in the
-- experience register. The contracts DO exist and this implementation follows one
-- of them — but approving and binding would not have put a carousel on the site.
-- Measured 2026-08-22: the register is a SPECIFICATION AND VERIFICATION system,
-- not a generator. Only three Go files touch it (**3** as of 2026-08-22) —
-- `write_experience_pattern_action.go` (records a contract),
-- `bind_site_experience_action.go` (records which page it applies to) and
-- `verify_site_experience_action.go`, whose own header says it "run[s] a bound
-- fork's criteria against the deployed page". NOTHING renders from
-- `site_experiences`. Binding a contract for a carousel that does not exist would
-- have produced a fork whose criteria then fail against the page.
-- And the trigger script says so in as many words: "nothing in this lane can write
-- status='approved'. Applying a verdict is a separate action that does not exist
-- yet, deliberately." So the register could not have been the delivery mechanism.
--
-- WHICH CONTRACT THIS IMPLEMENTS. `arrow-and-swipe-card-carousel`
-- (`experience_patterns`, still `draft` — **11** entries, **0** approved as of 2026-08-22, unchanged since
-- 2026-08-18). Its clauses and how each is met:
--
--   · "swipe natively, works with NO JavaScript at all" — the track is CSS
--     `overflow-x:auto` + `scroll-snap-type:x mandatory`. With JS blocked the
--     cards still scroll, snap, and are still real links. JS adds arrows ONLY.
--   · "fewer than two cards → every control is hidden" (the `no-inert-control`
--     invariant) — see the note below; this is met by a STRONGER test than the
--     contract asks for.
--   · "initialisation must be idempotent if the script is included twice" — a
--     one-time window guard plus a per-element `data-csg-ready` stamp, so both a
--     double include and a re-init on the same node are no-ops.
--   · "prefers-reduced-motion → scrolling is instant, not smooth" — honoured in
--     CSS and re-checked in JS at click time (a media query can change after load).
--   · auto-advance — the contract makes it conditional ("only if asked") and it is
--     NOT implemented. That is deliberate: it is the clause that drags in the
--     IntersectionObserver, the hover/focus pause and the re-derive-after-swipe
--     rules, and none of them can be got right unseen. Nothing here rotates, so
--     none of those failure modes exists. Adding it later is a separate change.
--
-- ⚠ THE `no-inert-control` INVARIANT IS MET BY OVERFLOW, NOT BY COUNTING CARDS,
-- and the difference is load-bearing on THIS component. `case-studies-grid` ships
-- a category filter (`/tools/assets/case-studies-grid.js`, live, HTTP 200) that
-- hides cards with `card.style.display='none'`. A card count taken at init says
-- "5 cards, show the arrows" and stays wrong after the visitor filters down to
-- one, leaving two buttons that cannot move anything — exactly what the invariant
-- forbids. So visibility is derived from `scrollWidth > clientWidth`, re-evaluated
-- on scroll, on resize, and via a MutationObserver on the cards' `style`
-- attribute, which is the filter's own mechanism. The controls therefore
-- disappear when filtering leaves nothing to scroll.
--
-- OPT-IN, DEFAULT OFF — the owner's ruling of 2026-08-02 (RFC_010 §2): new
-- authority on a SHARED seam ships as a field whose unsafe default is OFF, not as
-- a documented contract. `case-studies-grid` is placed on **4** pages across **3**
-- sites as of 2026-08-22 (ai-agent-orchestration.com ×2, finetuning.uk, leopardessconsulting.co.uk).
-- Everything below is inside `{{if .carousel_enabled}}`, so the other two sites
-- render byte-identically to today and cannot be changed by this file. This is a
-- layout change, so "measure that nothing breaks" is not enough — the other lanes
-- must be able to opt in when THEY choose.
--
-- ⚠ DURABILITY CAVEAT, STATED BECAUSE IT IS NOT SOLVED. The flag lives in
-- `page_components.content_data`. A re-render MERGES content_data and keeps it; a
-- full page REBUILD regenerates content_data from the writer, which has no reason
-- to emit `carousel_enabled`, so a rebuild of `index` would silently drop the
-- carousel (bugs_closed/238's shape: "save REPLACES, rerender MERGES"). There is
-- no durable per-site presentation flag on this seam today — `site_experiences`
-- would be the right home and nothing reads it. If the carousel vanishes after a
-- rebuild, this is why, and the fix is to re-set the key, not to re-debug the CSS.
--
-- ⚠ NO LINK IS BUILT IN JAVASCRIPT, deliberately. `RepairPageLinks` cannot tell an
-- anchor from JS that constructs one and will delete the link from the program,
-- leaving valid JS that does nothing (LANDMINES). The script here only calls
-- `scrollBy`; every card link stays a literal `<a href>` in the markup.
--
-- DOES NOT RE-RENDER. Placements keep their existing html until re-rendered; this
-- file files nothing. Propagate with a page-scoped `template_changed` rerender
-- (RUNBOOK R8) and verify at the artefact.
--
-- ROLLBACK: 559_case_studies_grid_optional_scroll_snap_carousel_ROLLBACK.sql
--   (restores the byte-exact prior template from migration_backups).

BEGIN;

-- 1. Byte-exact backup.
INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '559_case_studies_grid_optional_scroll_snap_carousel',
       'content_components', cc.id::text,
       jsonb_build_object('html_template', cc.html_template),
       'pre-559 html_template for case-studies-grid'
FROM content_components cc WHERE cc.id='3f946437-1dc7-4164-987d-620933589076';

-- 2. The template edit. Four literal anchors, each confirmed to occur EXACTLY once
--    in the stored template before this file was written.
UPDATE content_components
SET html_template = replace(replace(replace(replace(
      html_template,

      -- (a) the track: opt-in class + a hook the script can find.
      '<div class="csg-grid" role="list">',
      '<div class="csg-grid{{if .carousel_enabled}} csg-grid--carousel{{end}}" role="list"{{if .carousel_enabled}} data-csg-track tabindex="0" aria-label="{{.section_headline}}"{{end}}>'),

      -- (b) the controls, immediately after the track, before the footer.
      --     `hidden` in the markup: with no JS they never appear, which is the
      --     contract's "the arrows simply do nothing" in its strongest form.
      E'    </div>\n\n    <footer class="csg-footer">',
      E'    </div>\n\n{{if .carousel_enabled}}    <div class="csg-carousel-controls" data-csg-controls hidden>\n'
      || E'      <button class="csg-carousel-btn" type="button" data-csg-prev aria-label="Previous case studies">\n'
      || E'        <svg viewBox="0 0 16 16" fill="none" aria-hidden="true" width="18" height="18"><path d="M13 8H3M7 4L3 8l4 4" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"/></svg>\n'
      || E'      </button>\n'
      || E'      <button class="csg-carousel-btn" type="button" data-csg-next aria-label="More case studies">\n'
      || E'        <svg viewBox="0 0 16 16" fill="none" aria-hidden="true" width="18" height="18"><path d="M3 8h10M9 4l4 4-4 4" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"/></svg>\n'
      || E'      </button>\n'
      || E'    </div>\n{{end}}\n    <footer class="csg-footer">'),

      -- (c) the CSS, appended inside the existing <style>.
      '</style>',
      E'\n  /* ── optional carousel (559) — inert unless .csg-grid--carousel is set ── */\n'
      || E'  .case-studies-grid-section .csg-grid--carousel {\n'
      || E'    display: flex;\n    grid-template-columns: none;\n    overflow-x: auto;\n'
      || E'    scroll-snap-type: x mandatory;\n    scroll-behavior: smooth;\n'
      || E'    -webkit-overflow-scrolling: touch;\n    scroll-padding-left: 0.25rem;\n'
      || E'    padding-bottom: 0.75rem;\n  }\n'
      || E'  .case-studies-grid-section .csg-grid--carousel .csg-card {\n'
      || E'    flex: 0 0 clamp(260px, 32%, 380px);\n    scroll-snap-align: start;\n  }\n'
      || E'  /* the featured card spans two columns in the grid; in a track it must not. */\n'
      || E'  .case-studies-grid-section .csg-grid--carousel .csg-card-featured {\n'
      || E'    grid-column: auto;\n    grid-row: auto;\n  }\n'
      || E'  .case-studies-grid-section .csg-grid--carousel:focus-visible {\n'
      || E'    outline: 2px solid var(--color-accent, var(--color-secondary));\n    outline-offset: 4px;\n  }\n'
      || E'  .case-studies-grid-section .csg-carousel-controls {\n'
      || E'    display: flex;\n    gap: 0.5rem;\n    justify-content: flex-end;\n    margin-top: 1rem;\n  }\n'
      || E'  .case-studies-grid-section .csg-carousel-controls[hidden] { display: none; }\n'
      || E'  .case-studies-grid-section .csg-carousel-btn {\n'
      || E'    display: inline-flex;\n    align-items: center;\n    justify-content: center;\n'
      || E'    width: 2.75rem;\n    height: 2.75rem;\n    border-radius: 50%;\n'
      || E'    background: var(--section-surface);\n    color: var(--section-heading);\n'
      || E'    border: 1px solid var(--section-border);\n    cursor: pointer;\n'
      || E'    transition: background 0.2s, transform 0.2s;\n  }\n'
      || E'  .case-studies-grid-section .csg-carousel-btn:hover { background: var(--section-border); }\n'
      || E'  .case-studies-grid-section .csg-carousel-btn:focus-visible {\n'
      || E'    outline: 2px solid var(--color-accent, var(--color-secondary));\n    outline-offset: 2px;\n  }\n'
      || E'  @media (prefers-reduced-motion: reduce) {\n'
      || E'    .case-studies-grid-section .csg-grid--carousel { scroll-behavior: auto; }\n'
      || E'    .case-studies-grid-section .csg-carousel-btn { transition: none; }\n  }\n'
      || E'</style>'),

      -- (d) the behaviour. Arrows only; the track works without this file.
      '<script src="/tools/assets/case-studies-grid.js"></script>',
      E'<script src="/tools/assets/case-studies-grid.js"></script>\n'
      || E'{{if .carousel_enabled}}<script>\n'
      || E'(function () {\n'
      || E'  if (window.__csgCarouselBound) { return; }   /* included twice: one set of handlers */\n'
      || E'  window.__csgCarouselBound = true;\n'
      || E'  function reduceMotion() {\n'
      || E'    return window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;\n'
      || E'  }\n'
      || E'  function setup(track) {\n'
      || E'    if (track.getAttribute("data-csg-ready") === "1") { return; }  /* re-init on the same node */\n'
      || E'    track.setAttribute("data-csg-ready", "1");\n'
      || E'    var section = track.closest(".case-studies-grid-section");\n'
      || E'    var controls = section && section.querySelector("[data-csg-controls]");\n'
      || E'    if (!controls) { return; }\n'
      || E'    var prev = controls.querySelector("[data-csg-prev]");\n'
      || E'    var next = controls.querySelector("[data-csg-next]");\n'
      || E'    /* A control that cannot change anything must not be presented. Overflow,\n'
      || E'       not card count: the category filter hides cards with display:none. */\n'
      || E'    function update() {\n'
      || E'      var scrollable = track.scrollWidth > track.clientWidth + 1;\n'
      || E'      controls.hidden = !scrollable;\n'
      || E'      if (!scrollable) { return; }\n'
      || E'      prev.disabled = track.scrollLeft <= 1;\n'
      || E'      next.disabled = track.scrollLeft + track.clientWidth >= track.scrollWidth - 1;\n'
      || E'    }\n'
      || E'    function step(dir) {\n'
      || E'      var card = track.querySelector(".csg-card");\n'
      || E'      var by = card ? card.getBoundingClientRect().width + 24 : Math.round(track.clientWidth * 0.8);\n'
      || E'      try {\n'
      || E'        track.scrollBy({ left: dir * by, behavior: reduceMotion() ? "auto" : "smooth" });\n'
      || E'      } catch (e) { track.scrollLeft += dir * by; }   /* older engines */\n'
      || E'    }\n'
      || E'    prev.addEventListener("click", function () { step(-1); });\n'
      || E'    next.addEventListener("click", function () { step(1); });\n'
      || E'    track.addEventListener("scroll", update, { passive: true });\n'
      || E'    window.addEventListener("resize", update);\n'
      || E'    /* the filter writes style="display:none" on cards — re-decide when it does */\n'
      || E'    if (window.MutationObserver) {\n'
      || E'      new MutationObserver(update).observe(track, {\n'
      || E'        subtree: true, attributes: true, attributeFilter: ["style"]\n'
      || E'      });\n'
      || E'    }\n'
      || E'    update();\n'
      || E'  }\n'
      || E'  function init() {\n'
      || E'    var tracks = document.querySelectorAll(".csg-grid--carousel");\n'
      || E'    for (var i = 0; i < tracks.length; i++) { setup(tracks[i]); }\n'
      || E'  }\n'
      || E'  if (document.readyState === "loading") {\n'
      || E'    document.addEventListener("DOMContentLoaded", init);\n'
      || E'  } else { init(); }\n'
      || E'})();\n'
      || E'</script>{{end}}'),
    updated_at = now()
WHERE id='3f946437-1dc7-4164-987d-620933589076';

-- 3. Turn it on for ai-agent-orchestration.com ONLY.
UPDATE page_components pc
SET content_data = coalesce(pc.content_data, '{}'::jsonb) || '{"carousel_enabled": true}'::jsonb,
    updated_at = now()
FROM pages p
WHERE p.id = pc.page_id
  AND pc.component_id = '3f946437-1dc7-4164-987d-620933589076'
  AND p.site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da';

-- 4. Guards.
DO $$
DECLARE
  tpl        text;
  backed_up  int;
  enabled    int;
  others     int;
BEGIN
  SELECT count(*) INTO backed_up FROM migration_backups
   WHERE migration_name='559_case_studies_grid_optional_scroll_snap_carousel';
  IF backed_up <> 1 THEN
    RAISE EXCEPTION '559: expected 1 backup row, wrote %', backed_up;
  END IF;

  SELECT html_template INTO tpl FROM content_components
   WHERE id='3f946437-1dc7-4164-987d-620933589076';

  -- Every one of the four edits must have landed.
  IF tpl !~ 'csg-grid\{\{if \.carousel_enabled\}\} csg-grid--carousel' THEN
    RAISE EXCEPTION '559: the track class edit did not apply';
  END IF;
  IF tpl !~ 'data-csg-controls' THEN
    RAISE EXCEPTION '559: the controls markup did not apply';
  END IF;
  IF tpl !~ '\.csg-grid--carousel \{' THEN
    RAISE EXCEPTION '559: the carousel CSS did not apply';
  END IF;
  IF tpl !~ '__csgCarouselBound' THEN
    RAISE EXCEPTION '559: the behaviour script did not apply';
  END IF;

  -- THE LOAD-BEARING ONE: every added fragment must sit inside the opt-in, or a
  -- site that never asked for a carousel gets one. Three gates, three closers.
  -- FOUR, not three: the track line opens TWO gates (one for the class, one for the
  -- data/aria attributes), then the controls block and the script. The first draft of
  -- this guard said 3 and the rehearsal refused the file, which is the guard working.
  IF (SELECT count(*) FROM regexp_matches(tpl, '\{\{if \.carousel_enabled\}\}', 'g')) <> 4 THEN
    RAISE EXCEPTION '559: expected exactly 4 carousel_enabled gates, found %',
      (SELECT count(*) FROM regexp_matches(tpl, '\{\{if \.carousel_enabled\}\}', 'g'));
  END IF;

  -- The default must be OFF: no site other than this one may be switched on.
  SELECT count(*) INTO enabled FROM page_components pc JOIN pages p ON p.id=pc.page_id
   WHERE pc.component_id='3f946437-1dc7-4164-987d-620933589076'
     AND coalesce(pc.content_data->>'carousel_enabled','false')='true';
  SELECT count(*) INTO others FROM page_components pc JOIN pages p ON p.id=pc.page_id
   WHERE pc.component_id='3f946437-1dc7-4164-987d-620933589076'
     AND p.site_id <> '2a8ebf9c-20a2-4c39-b191-840b012371da'
     AND coalesce(pc.content_data->>'carousel_enabled','false')='true';
  IF others <> 0 THEN
    RAISE EXCEPTION '559: % placement(s) on OTHER sites were switched on; the default must stay OFF for them', others;
  END IF;
  IF enabled <> 2 THEN
    RAISE EXCEPTION '559: expected exactly 2 enabled placements on this site, found %', enabled;
  END IF;

  -- No link may be constructed in JavaScript (RepairPageLinks would delete it).
  IF tpl ~ 'createElement\("a"\)|innerHTML\s*=.*<a ' THEN
    RAISE EXCEPTION '559: the script builds an anchor; RepairPageLinks would silently remove it';
  END IF;

  RAISE NOTICE '559 OK: case-studies-grid carousel is opt-in (4 gates), ON for 2 placements on ai-agent-orchestration.com, OFF everywhere else. Nothing re-renders — propagate with a template_changed rerender.';
END $$;

COMMIT;
