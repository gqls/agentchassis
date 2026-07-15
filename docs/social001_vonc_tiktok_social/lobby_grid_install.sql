-- ============================================================
-- lobby-grid build, step: install the lobby-grid loader + marker
-- ============================================================
-- Order of operations for the whole build (see PLAN_lobby-grid.md):
--   1. THIS FILE part A: insert the js_snippet (loader).
--   2. Trigger site-asset-renderer (bundles /assets/js/snippets.js) —
--      reuse trigger-asset-renderer-vonc.sh.
--   3. THIS FILE part B: add the data-runtime-fill marker to lobby-grid
--      (template + current index instance) so the assembler keeps it.
--   4. Commit /data/provocations.json WITH the new `arena` object to the
--      sites repo (interim: provocations.sample.json v2 until Phase 3 emits it).
--   5. Rerender the index (rerender-index-vonc.sh) → lobby-grid ships + fills.
--
-- FIRST verify schema before running:  \d js_snippets

-- ── Part A: the loader snippet (dollar-quoted; no escaping needed) ──
INSERT INTO js_snippets (name, description, applies_to, js_content, is_active)
VALUES (
  'lobby-grid-loader',
  'Fills the lobby-grid shell (Arena: 6 provocation cards + header + CTA) from /data/provocations.json data.arena. Fails gracefully.',
  '["lobby-grid"]'::jsonb,
  $js$/* lobby-grid-loader
 * Fetches /data/provocations.json and fills the lobby-grid shell (the Arena:
 * six enterable provocation cards + header + CTA) from data.arena.
 * Coexists with the template's own hover/entrance IIFE (different concern).
 * Fails gracefully: if fetch fails, data is malformed, or arena is absent,
 * the shell is left as-is (no thrown errors, no broken layout) — the section
 * carries data-runtime-fill, so an unfilled shell simply renders empty space
 * rather than breaking the page.
 *
 * DOM contract (lobby-grid template, component 9304f14d):
 *   [data-component="lobby-grid"]
 *     .lobby-grid-section__eyebrow            textContent
 *     .lobby-grid-section__title              innerHTML (may carry <em>)
 *     .lobby-grid-section__subtitle           textContent
 *     .lobby-grid-section__grid .lobby-grid-section__card  × 6, in DOM order:
 *       featured, standard ×4, wide (order matches data.arena.cards[0..5])
 *       .lobby-grid-section__card-icon svg    innerHTML (SVG inner markup) —
 *                                             or emoji fallback (see fillIcon)
 *       .lobby-grid-section__card-tag         textContent
 *       .lobby-grid-section__card-title       textContent
 *       .lobby-grid-section__card-desc        textContent
 *       .lobby-grid-section__card-stat span (the one WITHOUT the dot class)
 *                                             textContent (dot span preserved)
 *     .lobby-grid-section__cta-label          textContent
 *     .lobby-grid-section__cta-btn            href + trailing text node
 *                                             (inline play SVG preserved)
 *
 * Data contract (provocations.json):
 *   arena: {
 *     eyebrow, title, subtitle, cta_label,
 *     cta: { label, url },
 *     cards: [ up to 6 × { icon, tag, title, desc, stat, url? } ]
 *   }
 *   card.icon: SVG inner markup (starts with "<", e.g. "<path d=\"...\"/>")
 *              or a short emoji/text fallback.
 */
(function () {
  "use strict";

  function fillLobbyGrid(data) {
    var section = document.querySelector('[data-component="lobby-grid"]');
    if (!section || !data || !data.arena) return;

    var a = data.arena;

    // --- header ---
    var eyebrow = section.querySelector(".lobby-grid-section__eyebrow");
    if (eyebrow && a.eyebrow) eyebrow.textContent = a.eyebrow;

    // title may contain <em> emphasis — innerHTML, our own JSON only.
    var title = section.querySelector(".lobby-grid-section__title");
    if (title && a.title) title.innerHTML = a.title;

    var subtitle = section.querySelector(".lobby-grid-section__subtitle");
    if (subtitle && a.subtitle) subtitle.textContent = a.subtitle;

    // section aria-label is a template slot too — set from the title text.
    if (title && title.textContent) {
      section.setAttribute("aria-label", title.textContent.trim());
    }

    // --- the six cards (DOM order: featured, standard ×4, wide) ---
    var cards = section.querySelectorAll(
      ".lobby-grid-section__grid .lobby-grid-section__card"
    );
    var entries = Array.isArray(a.cards) ? a.cards : [];
    var n = Math.min(cards.length, entries.length);
    for (var i = 0; i < n; i++) {
      fillCard(cards[i], entries[i]);
    }

    // --- CTA ---
    var ctaLabel = section.querySelector(".lobby-grid-section__cta-label");
    if (ctaLabel && a.cta_label) ctaLabel.textContent = a.cta_label;

    var ctaBtn = section.querySelector(".lobby-grid-section__cta-btn");
    if (ctaBtn && a.cta) {
      if (a.cta.url) ctaBtn.setAttribute("href", a.cta.url);
      if (a.cta.label) setTrailingLabel(ctaBtn, a.cta.label);
    }
  }

  function fillCard(card, entry) {
    if (!card || !entry) return;

    fillIcon(card.querySelector(".lobby-grid-section__card-icon"), entry.icon);

    var tag = card.querySelector(".lobby-grid-section__card-tag");
    if (tag && entry.tag) tag.textContent = entry.tag;

    var cardTitle = card.querySelector(".lobby-grid-section__card-title");
    if (cardTitle && entry.title) cardTitle.textContent = entry.title;

    var desc = card.querySelector(".lobby-grid-section__card-desc");
    if (desc && entry.desc) desc.textContent = entry.desc;

    // stat text lives in the span WITHOUT the dot class; the dot span (and its
    // pulse animation) must be preserved.
    if (entry.stat) {
      var statSpans = card.querySelectorAll(".lobby-grid-section__card-stat span");
      for (var i = 0; i < statSpans.length; i++) {
        if (statSpans[i].className.indexOf("card-stat-dot") === -1) {
          statSpans[i].textContent = entry.stat;
          break;
        }
      }
    }

    // enterable room: navigate on click/Enter when a url is provided.
    if (entry.url) {
      card.style.cursor = "pointer";
      card.setAttribute("tabindex", "0");
      card.addEventListener("click", function () {
        window.location.href = entry.url;
      });
      card.addEventListener("keydown", function (e) {
        if (e.key === "Enter") window.location.href = entry.url;
      });
    }
  }

  // icon slot is the svg's INNER markup (viewBox fixed at 0 0 24 24). If the
  // data carries SVG inner markup ("<path .../>"), inject it into the svg;
  // if it carries a short emoji/text fallback, show that in the icon
  // container instead (replacing the svg).
  function fillIcon(iconWrap, icon) {
    if (!iconWrap || !icon) return;
    var svg = iconWrap.querySelector("svg");
    if (icon.charAt(0) === "<") {
      if (svg) svg.innerHTML = icon;
    } else {
      iconWrap.textContent = icon;
    }
  }

  // Anchor label is a text node after the inline SVG — replace the last
  // non-empty text node (or append one), preserving the SVG. Same pattern as
  // the provocation-card loader's setButtonLabel.
  function setTrailingLabel(el, label) {
    var replaced = false;
    for (var i = el.childNodes.length - 1; i >= 0; i--) {
      var node = el.childNodes[i];
      if (node.nodeType === 3 && node.textContent.replace(/\s/g, "") !== "") {
        node.textContent = " " + label;
        replaced = true;
        break;
      }
    }
    if (!replaced) el.appendChild(document.createTextNode(" " + label));
  }

  function init() {
    if (!document.querySelector('[data-component="lobby-grid"]')) return;
    fetch("/data/provocations.json", { cache: "no-store" })
      .then(function (res) {
        if (!res.ok) throw new Error("provocations.json " + res.status);
        return res.json();
      })
      .then(fillLobbyGrid)
      .catch(function () {
        /* graceful: leave the shell as-is */
      });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
$js$,
  true
);

-- verify
SELECT name, LENGTH(js_content) AS js_len, applies_to, is_active
FROM js_snippets WHERE name = 'lobby-grid-loader';

-- ── Part B: data-runtime-fill marker (AFTER the assembler patch is live — it is) ──
-- template (future renders)
UPDATE content_components
SET html_template = REPLACE(html_template,
      'data-component="lobby-grid"',
      'data-component="lobby-grid" data-runtime-fill="true"'),
    updated_at = NOW()
WHERE id = '9304f14d-e19b-4ce1-b3fd-f6a315aec6ed'
  AND html_template LIKE '%data-component="lobby-grid"%'
  AND html_template NOT LIKE '%data-runtime-fill%'
RETURNING (html_template LIKE '%data-runtime-fill%') AS template_marked;

-- current index instance (so this rerender keeps it)
UPDATE page_components
SET rendered_html = REPLACE(rendered_html,
      'data-component="lobby-grid"',
      'data-component="lobby-grid" data-runtime-fill="true"'),
    updated_at = NOW()
WHERE component_id = '9304f14d-e19b-4ce1-b3fd-f6a315aec6ed'
  AND page_id = 'b4d24f8e-fccd-49df-9dad-aa56a0b20a68'
  AND rendered_html LIKE '%data-component="lobby-grid"%'
  AND rendered_html NOT LIKE '%data-runtime-fill%'
RETURNING (rendered_html LIKE '%data-runtime-fill%') AS rendered_marked;

-- NOTE (template vs rendered literal): the TEMPLATE carries the plain
-- data-component="lobby-grid" literal (confirmed in the dump), and the current
-- rendered_html was produced from it, so the same REPLACE literal applies to
-- both. If either UPDATE returns 0 rows, check the stored text for the exact
-- literal before concluding — do not proceed on an unexplained 0.

-- NOTE (the template's own inline script queries
-- [data-component="lobby-grid"], an attribute-EXISTS+value selector — adding
-- the extra attribute does not affect it.)
