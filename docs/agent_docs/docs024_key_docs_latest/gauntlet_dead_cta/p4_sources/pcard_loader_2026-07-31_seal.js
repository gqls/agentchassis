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
 * `today.headline` and `today.body` STILL EXIST in the feed and MUST — the engine
 * (`internal/tools-api/handlers/round.go`) fetches this file server-side and uses
 * the whole `today` object as the round's provocation. **This file must simply not
 * READ them.** That is the seal: a renderer-level invariant, not key absence.
 * (An earlier attempt removed the keys instead; that would have served every round
 * an empty question. Do not "simplify" this back to reading `today`.)
 *
 * So the display copy lives in SIBLING keys the engine never looks at: `data.seal`
 * and `data.sample`. The builder's `check_seal` refuses to emit a feed where
 * anything outside `today` names today's provocation.
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

    /* `t` is read ONLY for the CTA and the stat strip. Its headline/body are the
     * engine's copy of today's provocation and must never be painted here. */
    var t = data.today || {};
    var seal = data.seal || {};
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
      if (headline && seal.headline) headline.innerHTML = seal.headline;
      if (body && seal.body) body.textContent = seal.body;
    }

    /* --- primary CTA: into today's SEALED round (anchor keeps its inline SVG) ---
     * This is the one place today's provocation is referred to, and it names the
     * seal rather than the question. */
    var primary = section.querySelector(".pc-btn-primary");
    var route = seal.cta || t.primary_cta;   /* seal wording preferred; same target */
    if (primary && route) {
      if (route.url) primary.setAttribute("href", route.url);
      if (route.label) setButtonLabel(primary, route.label);
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
