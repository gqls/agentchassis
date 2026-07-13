-- ============================================================
-- provocations-archive-list, Step 4: install the archive loader
-- ============================================================
-- NOTE: there is NO marker part this time — data-runtime-fill was emitted AT
-- GENERATION in the component template (the payoff of the generation-time guard).
-- Run order:
--   1. THIS FILE: insert the js_snippet.
--   2. trigger-asset-renderer-vonc.sh  (bundles /assets/js/snippets.js).
--   3. Commit the v3 provocations.json (now carrying `archive` alongside today/
--      lobby/arena) to the sites repo at vonc.com/data/provocations.json.
--   4. Browser-verify /provocations/index.html: header copy (build-time), archive
--      rows filled by the loader, empty state hidden, CTA back to today's provocation.
-- \d js_snippets first if in doubt (schema previously confirmed: name, description,
-- js_content, semantic_tags, applies_to, dependencies, is_active).

INSERT INTO js_snippets (name, description, applies_to, js_content, is_active)
VALUES (
  'provocations-archive-loader',
  'Fills the Provocations Archive list from /data/provocations.json data.archive (clone-template rows; hides the empty state). Header/CTA are build-time content. Fails gracefully.',
  '["provocations-archive-list"]'::jsonb,
  $js$/* provocations-archive-loader
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
$js$,
  true
);

-- verify
SELECT name, LENGTH(js_content) AS js_len, applies_to, is_active
FROM js_snippets WHERE name = 'provocations-archive-loader';
