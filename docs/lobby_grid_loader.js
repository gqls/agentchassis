/* lobby-grid-loader
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
