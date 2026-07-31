/* provocation-card-loader
 * Fetches /data/provocations.json and fills the provocation-card shell.
 * Fails gracefully: if fetch fails or data is malformed, the shell is left
 * as-is (no thrown errors, no broken layout).
 *
 * THE SEAL (owner ruling 2026-07-31) — WHY THIS PAINTS A PAST PROVOCATION
 * Today's provocation is readable in the Gauntlet, after entry, and nowhere else.
 * This card used to paint `today.headline` + `today.body`, which is exactly the
 * provocation the Gauntlet page is built to conceal — so every visitor arriving
 * at "/" had read the argument before reaching the sealed door.
 *
 * So the card now shows `data.sample`: a PAST provocation, in full, which is safe
 * because it has already been argued. The visitor learns what a provocation reads
 * like, then goes to argue today's without knowing it. The CTA carries the seal.
 *
 * `today.headline` and `today.body` DO NOT EXIST in the feed any more, and the
 * generator refuses to emit them (build_provocations.py `guard`). This file never
 * reads them — do not "restore" them here to fill an empty slot; an empty slot
 * would mean the feed is malformed, and the fallback below handles that honestly.
 *
 * Verify by RENDERING, never by grepping the HTML — the shell is served empty and
 * filled here, so a curl grep says "absent" on the page that shows it:
 *   ~/.venvs/vonc_pw/bin/python scripts/provocation_leak_sweep.py
 */
(function () {
  "use strict";

  function fillProvocationCard(data) {
    var section = document.querySelector('[data-component="provocation-card"]');
    if (!section || !data) return;

    var t = data.today || {};
    var sample = data.sample;

    var eyebrow = section.querySelector(".pc-eyebrow");
    var headline = section.querySelector(".pc-headline");
    var body = section.querySelector(".pc-body");

    /* The card's subject is the SAMPLE (a past provocation). If the feed carries no
     * sample, fall back to the seal statement rather than leaving three empty
     * elements — an empty headline is the defect this lane already fixed once. What
     * we must never do is fall back to today's provocation: it is not in the feed. */
    if (sample && sample.headline) {
      if (eyebrow) {
        eyebrow.textContent = sample.date
          ? (sample.eyebrow || "A past provocation") + " · " + sample.date
          : (sample.eyebrow || "A past provocation");
      }
      // headline may contain <em> emphasis — innerHTML, but only ever with our own
      // authored JSON (no model output, no visitor input), so this stays safe.
      if (headline) headline.innerHTML = sample.headline;
      if (body && sample.body) body.textContent = sample.body;
    } else {
      if (eyebrow && t.eyebrow) eyebrow.textContent = t.eyebrow;
      if (headline && t.seal_headline) headline.innerHTML = t.seal_headline;
      if (body && t.seal_body) body.textContent = t.seal_body;
    }

    /* --- primary CTA: into today's SEALED round (anchor keeps its inline SVG) ---
     * This is the one place today's provocation is referred to, and it names the
     * seal rather than the question. */
    var primary = section.querySelector(".pc-btn-primary");
    if (primary && t.primary_cta) {
      if (t.primary_cta.url) primary.setAttribute("href", t.primary_cta.url);
      if (t.primary_cta.label) setButtonLabel(primary, t.primary_cta.label);
    }

    /* --- secondary CTA: read the sample's full case if we showed a sample, else
     * the archive. Both are real destinations; neither reveals today's. --- */
    var secondary = section.querySelector(".pc-btn-secondary");
    if (secondary) {
      if (sample && sample.url && sample.cta_label) {
        secondary.setAttribute("href", sample.url);
        setButtonLabel(secondary, sample.cta_label);
      } else if (t.secondary_cta) {
        if (t.secondary_cta.url) secondary.setAttribute("href", t.secondary_cta.url);
        if (t.secondary_cta.label) setButtonLabel(secondary, t.secondary_cta.label);
      }
    }

    // --- stat strip (3 stats) — facts true by construction of the Gauntlet ---
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
