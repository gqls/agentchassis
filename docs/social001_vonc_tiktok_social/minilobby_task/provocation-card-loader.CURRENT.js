/* provocation-card-loader
 * Fetches /data/provocations.json and fills the provocation-card shell.
 * Coexists with the template's own card-activation IIFE (different concern).
 * Fails gracefully: if fetch fails or data is malformed, the shell is left
 * as-is (no thrown errors, no broken layout).
 */
(function () {
  "use strict";

  function fillProvocationCard(data) {
    var section = document.querySelector('[data-component="provocation-card"]');
    if (!section || !data || !data.today) return;

    var t = data.today;

    // --- main provocation ---
    var eyebrow = section.querySelector(".pc-eyebrow");
    if (eyebrow && t.eyebrow) eyebrow.textContent = t.eyebrow;

    // headline may contain <em> emphasis — use innerHTML, but only with our
    // own JSON (not user input), so this is safe in this context.
    var headline = section.querySelector(".pc-headline");
    if (headline && t.headline) headline.innerHTML = t.headline;

    var body = section.querySelector(".pc-body");
    if (body && t.body) body.textContent = t.body;

    // --- primary CTA (anchor: set href + label, preserve the inline SVG) ---
    var primary = section.querySelector(".pc-btn-primary");
    if (primary && t.primary_cta) {
      if (t.primary_cta.url) primary.setAttribute("href", t.primary_cta.url);
      if (t.primary_cta.label) setButtonLabel(primary, t.primary_cta.label);
    }

    // --- secondary CTA ---
    var secondary = section.querySelector(".pc-btn-secondary");
    if (secondary && t.secondary_cta) {
      if (t.secondary_cta.url) secondary.setAttribute("href", t.secondary_cta.url);
      if (t.secondary_cta.label) setButtonLabel(secondary, t.secondary_cta.label);
    }

    // --- stat strip (3 stats) ---
    if (Array.isArray(t.stats)) {
      var values = section.querySelectorAll(".pc-stat-value");
      var labels = section.querySelectorAll(".pc-stat-label");
      for (var i = 0; i < values.length; i++) {
        if (t.stats[i]) {
          if (t.stats[i].value != null) values[i].textContent = t.stats[i].value;
          if (labels[i] && t.stats[i].label != null) labels[i].textContent = t.stats[i].label;
        }
      }
    }

    // --- mini-lobby cards (up to 4) ---
    if (Array.isArray(data.lobby)) {
      var cards = section.querySelectorAll(".pc-card");
      for (var j = 0; j < cards.length; j++) {
        var item = data.lobby[j];
        var card = cards[j];
        if (!item) {
          // no data for this card — hide it so we don't show an empty box
          card.style.display = "none";
          continue;
        }
        card.style.display = "";
        var icon = card.querySelector(".pc-card-icon");
        var title = card.querySelector(".pc-card-title");
        var desc = card.querySelector(".pc-card-desc");
        if (icon && item.icon != null) icon.textContent = item.icon;
        if (title && item.title != null) title.textContent = item.title;
        if (desc && item.desc != null) desc.textContent = item.desc;
        // make the card navigable if a url is supplied and it isn't already a link
        if (item.url && card.tagName !== "A") {
          card.style.cursor = "pointer";
          card.setAttribute("role", "link");
          card.setAttribute("tabindex", "0");
          (function (url) {
            card.addEventListener("click", function () { window.location.href = url; });
            card.addEventListener("keydown", function (e) {
              if (e.key === "Enter" || e.key === " ") { e.preventDefault(); window.location.href = url; }
            });
          })(item.url);
        }
      }
    }
  }

  // Replace an anchor's text label without destroying child elements (e.g. SVG).
  // Sets the text of the first text node, or prepends one if none exists.
  function setButtonLabel(anchor, label) {
    var textNode = null;
    for (var k = 0; k < anchor.childNodes.length; k++) {
      if (anchor.childNodes[k].nodeType === 3 && anchor.childNodes[k].textContent.trim() !== "") {
        textNode = anchor.childNodes[k];
        break;
      }
    }
    if (textNode) {
      textNode.textContent = label + " ";
    } else {
      anchor.insertBefore(document.createTextNode(label + " "), anchor.firstChild);
    }
  }

  function init() {
    if (!document.querySelector('[data-component="provocation-card"]')) return;
    fetch("/data/provocations.json", { cache: "no-cache" })
      .then(function (r) {
        if (!r.ok) throw new Error("provocations.json HTTP " + r.status);
        return r.json();
      })
      .then(fillProvocationCard)
      .catch(function (err) {
        if (window.console && console.warn) {
          console.warn("provocation-card-loader: could not load provocations", err);
        }
      });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();

