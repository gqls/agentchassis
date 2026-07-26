/* provocations-archive-loader
 * Fills the Provocations Archive list from /data/provocations.json data.archive
 * and opens a full case in place, at a deep-linkable URL.
 *
 * The section's header + CTA are BUILD-TIME content (page-content-writer fills
 * them) — this loader touches ONLY the list and the detail region it creates.
 * Fails gracefully: on missing feed / malformed data / absent section, the
 * shell and its empty-state line are left exactly as built.
 *
 * DOM contract (generated template, component 70d6662a):
 *   [data-component="provocations-archive-list"]
 *     .provocations-archive__list
 *       a.provocations-archive__item[data-archive-template][hidden]   ← clone per entry
 *         time.provocations-archive__item-date        textContent
 *         .provocations-archive__item-title           textContent
 *         .provocations-archive__item-teaser          textContent
 *         .provocations-archive__item-stat            hidden when the entry has no stat
 *         .provocations-archive__item-stat-value      textContent
 *         href                                        entry.url
 *     .provocations-archive__empty                    hidden once >=1 entry rendered
 *
 * Built by this loader (the template has no slot for it):
 *   .provocations-archive__detail                     the open case, deep-linkable
 *
 * Data contract (provocations.json):
 *   archive: { entries: [ { date, title, teaser, slug, detail_body?, url? } ] }
 *   newest first, cap 24.
 *
 * An entry WITH detail_body is openable: class --linked, a real href
 * (/provocations/index.html?entry=<slug>), and a click handler that opens the
 * case in place via history.pushState. An entry WITHOUT one is NOT offered as
 * openable: class --static, the template's placeholder href is REMOVED (it is
 * href="#", a dead control), no tabindex, no handler. The absence of a case is
 * a real state in the feed, not an error — some archive rows have no written
 * case yet.
 */
(function () {
  "use strict";

  var SECTION_SEL = '[data-component="provocations-archive-list"]';
  var STYLE_ID = "provocations-archive-detail-style";
  var entriesBySlug = {};

  // The component's template carries no styling for a detail region, so the
  // loader that creates the markup brings its own. Scoped to the section and
  // written only once.
  function ensureStyle() {
    if (document.getElementById(STYLE_ID)) return;
    var style = document.createElement("style");
    style.id = STYLE_ID;
    style.textContent =
      ".provocations-archive .provocations-archive__detail{" +
      "margin:1.5rem 0 0;padding:1.5rem;border:1px solid var(--section-border,rgba(0,0,0,0.12));" +
      "border-radius:var(--border-radius,6px);background:var(--section-surface,rgba(0,0,0,0.02));}" +
      ".provocations-archive .provocations-archive__detail[hidden]{display:none;}" +
      ".provocations-archive .provocations-archive__detail-head{display:flex;align-items:baseline;" +
      "justify-content:space-between;gap:1rem;flex-wrap:wrap;margin-bottom:0.75rem;}" +
      ".provocations-archive .provocations-archive__detail-date{font-size:0.7rem;font-weight:700;" +
      "letter-spacing:0.15em;text-transform:uppercase;opacity:0.7;margin:0;}" +
      ".provocations-archive .provocations-archive__detail-close{background:transparent;" +
      "border:1px solid var(--section-border,rgba(0,0,0,0.2));border-radius:4px;cursor:pointer;" +
      "font:inherit;font-size:0.7rem;font-weight:700;letter-spacing:0.1em;text-transform:uppercase;" +
      "padding:0.45rem 0.8rem;color:inherit;min-height:44px;}" +
      ".provocations-archive .provocations-archive__detail-title{font-size:clamp(1.2rem,2.4vw,1.6rem);" +
      "font-weight:800;line-height:1.25;margin:0 0 0.85rem;overflow-wrap:anywhere;}" +
      ".provocations-archive .provocations-archive__detail-body p{margin:0 0 0.85rem;line-height:1.7;" +
      "overflow-wrap:anywhere;}" +
      ".provocations-archive .provocations-archive__detail-body p:last-child{margin-bottom:0;}" +
      ".provocations-archive .provocations-archive__item--static{cursor:default;}";
    document.head.appendChild(style);
  }

  function buildDetail(section) {
    var existing = section.querySelector(".provocations-archive__detail");
    if (existing) return existing;

    var detail = document.createElement("article");
    detail.className = "provocations-archive__detail";
    detail.setAttribute("aria-live", "polite");
    detail.setAttribute("tabindex", "-1");
    detail.hidden = true;

    var list = section.querySelector(".provocations-archive__list");
    if (list && list.parentNode) {
      list.parentNode.insertBefore(detail, list.nextSibling);
    } else {
      section.appendChild(detail);
    }
    return detail;
  }

  // Built on open, torn down on close, so a closed region holds no text at
  // all. innerText on a hidden element falls back to textContent, so anything
  // left inside it would read as content even while invisible.
  function fillDetail(section, detail, entry) {
    detail.textContent = "";

    var head = document.createElement("div");
    head.className = "provocations-archive__detail-head";

    var date = document.createElement("p");
    date.className = "provocations-archive__detail-date";
    date.textContent = entry.date || "";

    var close = document.createElement("button");
    close.type = "button";
    close.className = "provocations-archive__detail-close";
    close.textContent = "Close";
    close.addEventListener("click", function () {
      closeDetail(section, true);
    });

    head.appendChild(date);
    head.appendChild(close);

    var title = document.createElement("h2");
    title.className = "provocations-archive__detail-title";
    title.textContent = entry.title || "";

    var body = document.createElement("div");
    body.className = "provocations-archive__detail-body";
    // detail_body is authored prose with blank-line paragraphs. Real <p>
    // elements built with textContent, so feed text is never parsed as markup.
    var paras = String(entry.detail_body).split(/\n\s*\n/);
    for (var i = 0; i < paras.length; i++) {
      var chunk = paras[i].trim();
      if (!chunk) continue;
      var p = document.createElement("p");
      p.textContent = chunk;
      body.appendChild(p);
    }

    detail.appendChild(head);
    detail.appendChild(title);
    detail.appendChild(body);
  }

  function openDetail(section, slug, push) {
    var entry = entriesBySlug[slug];
    if (!entry || !entry.detail_body) return false;

    var detail = buildDetail(section);
    fillDetail(section, detail, entry);
    detail.hidden = false;

    var items = section.querySelectorAll(".provocations-archive__item");
    for (var j = 0; j < items.length; j++) {
      items[j].classList.toggle("is-open", items[j].getAttribute("data-slug") === slug);
    }

    if (push && window.history && window.history.pushState) {
      window.history.pushState({ entry: slug }, "", "?entry=" + encodeURIComponent(slug));
    }
    if (detail.focus) detail.focus();
    return true;
  }

  function closeDetail(section, push) {
    var detail = section.querySelector(".provocations-archive__detail");
    if (detail) {
      detail.hidden = true;
      detail.textContent = "";
    }
    var items = section.querySelectorAll(".provocations-archive__item");
    for (var i = 0; i < items.length; i++) items[i].classList.remove("is-open");
    if (push && window.history && window.history.pushState) {
      window.history.pushState({}, "", window.location.pathname);
    }
  }

  function currentSlug() {
    var m = /[?&]entry=([^&]+)/.exec(window.location.search || "");
    if (!m) return "";
    try {
      return decodeURIComponent(m[1]);
    } catch (e) {
      return m[1];
    }
  }

  function fillArchive(section, data) {
    if (!data || !data.archive || !Array.isArray(data.archive.entries)) return;

    var list = section.querySelector(".provocations-archive__list");
    var tmpl = section.querySelector("[data-archive-template]");
    if (!list || !tmpl) return;

    tmpl.removeAttribute("href");

    var entries = data.archive.entries.slice(0, 24);
    var rendered = 0;

    for (var i = 0; i < entries.length; i++) {
      var e = entries[i];
      if (!e || !e.title) continue;

      var node = tmpl.cloneNode(true);
      node.removeAttribute("data-archive-template");
      node.removeAttribute("hidden");
      setText(node, ".provocations-archive__item-date", e.date);
      setText(node, ".provocations-archive__item-title", e.title);
      setText(node, ".provocations-archive__item-teaser", e.teaser);

      // No invented participation numbers exist any more, so the stat slot is
      // hidden rather than left as an orphan separator dot.
      var stat = node.querySelector(".provocations-archive__item-stat");
      if (e.stat) {
        setText(node, ".provocations-archive__item-stat-value", e.stat);
      } else if (stat) {
        stat.hidden = true;
      }

      if (e.slug) {
        node.setAttribute("data-slug", e.slug);
        entriesBySlug[e.slug] = e;
      }

      if (e.detail_body) {
        node.classList.add("provocations-archive__item--linked");
        node.setAttribute("href", e.url || "?entry=" + encodeURIComponent(e.slug || ""));
        bindOpen(section, node, e.slug);
      } else {
        // Not openable: strip the template's placeholder href="#" so this is
        // not a dead control, and leave it out of the tab order.
        node.classList.add("provocations-archive__item--static");
        node.removeAttribute("href");
        node.removeAttribute("tabindex");
      }

      list.appendChild(node);
      rendered++;
    }

    if (rendered > 0) {
      var empty = section.querySelector(".provocations-archive__empty");
      if (empty) empty.hidden = true;
    }

    buildDetail(section);

    // Deep link: open the requested case on first paint, no click needed.
    var slug = currentSlug();
    if (slug) openDetail(section, slug, false);
  }

  function bindOpen(section, node, slug) {
    if (!slug) return;
    node.addEventListener("click", function (ev) {
      // Leave modified clicks alone — the href is a real URL and opening it in
      // a new tab must keep working.
      if (ev.metaKey || ev.ctrlKey || ev.shiftKey || ev.altKey || ev.button !== 0) return;
      if (openDetail(section, slug, true)) ev.preventDefault();
    });
  }

  function setText(root, sel, val) {
    if (val === undefined || val === null) return;
    var el = root.querySelector(sel);
    if (el) el.textContent = val;
  }

  function init() {
    var section = document.querySelector(SECTION_SEL);
    if (!section) return;
    ensureStyle();

    window.addEventListener("popstate", function () {
      var slug = currentSlug();
      if (slug) {
        openDetail(section, slug, false);
      } else {
        closeDetail(section, false);
      }
    });

    fetch("/data/provocations.json", { cache: "no-store" })
      .then(function (res) {
        if (!res.ok) throw new Error("provocations.json " + res.status);
        return res.json();
      })
      .then(function (data) {
        fillArchive(section, data);
      })
      .catch(function () {
        /* graceful: leave the shell + empty state as built */
      });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
