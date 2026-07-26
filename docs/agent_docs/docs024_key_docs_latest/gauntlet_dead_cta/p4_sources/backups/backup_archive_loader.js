/* provocations-archive-loader
 * Fills the Provocations Archive list from /data/provocations.json data.archive.
 * The section's header + CTA are BUILD-TIME content (page-content-writer fills
 * them) — this loader touches ONLY the list: it clones the hidden template item
 * per entry and hides the empty state once anything renders.
 * Fails gracefully: on missing feed / malformed data / absent section, the shell
 * and its empty-state line are left exactly as built.
 *
 * DOM contract (generated template, component 70d6662a — confirmed from the dump):
 *   [data-component="provocations-archive-list"]
 *     .provocations-archive__list
 *       a.provocations-archive__item[data-archive-template][hidden]   ← clone per entry
 *         time.provocations-archive__item-date        textContent
 *         .provocations-archive__item-title           textContent
 *         .provocations-archive__item-teaser          textContent
 *         .provocations-archive__item-stat-value      textContent (dot span preserved by cloning)
 *         href                                        entry.url
 *     .provocations-archive__empty                    hidden once >=1 entry rendered
 *
 * Data contract (provocations.json):
 *   archive: { entries: [ { date, title, teaser, stat, url } ... ] }   newest first, cap 24
 */
(function () {
  "use strict";

  function fillArchive(data) {
    var section = document.querySelector('[data-component="provocations-archive-list"]');
    if (!section || !data || !data.archive || !Array.isArray(data.archive.entries)) return;

    var list = section.querySelector(".provocations-archive__list");
    var tmpl = section.querySelector("[data-archive-template]");
    if (!list || !tmpl) return;

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
      setText(node, ".provocations-archive__item-stat-value", e.stat);
      if (e.url) node.setAttribute("href", e.url);
      list.appendChild(node);
      rendered++;
    }

    if (rendered > 0) {
      var empty = section.querySelector(".provocations-archive__empty");
      if (empty) empty.hidden = true;
    }
  }

  function setText(root, sel, val) {
    if (!val) return;
    var el = root.querySelector(sel);
    if (el) el.textContent = val;
  }

  function init() {
    if (!document.querySelector('[data-component="provocations-archive-list"]')) return;
    fetch("/data/provocations.json", { cache: "no-store" })
      .then(function (res) {
        if (!res.ok) throw new Error("provocations.json " + res.status);
        return res.json();
      })
      .then(fillArchive)
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

