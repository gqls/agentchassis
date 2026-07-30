/* tool-ai-vendor-trust-checklist
 *
 * Client-side only: no fetch, no framework, no LLM call, no server round trip.
 * Counts the ticked items, drops the sector-certification item out of the
 * denominator when it is marked not applicable, and writes the score, the
 * denominator and the verdict tier back into the DOM.
 *
 * All visitor-facing copy lives in the TEMPLATE, including the three verdict
 * tiers, which are read from data- attributes on #vtc-verdict-box. This file
 * therefore holds logic and no prose: the copy cannot drift between two places
 * because there is only one place.
 *
 * The static markup is deliberately the same state this script computes for a
 * fresh page (0 of 12, lowest tier). If the script never runs, the panel is
 * honest rather than blank.
 *
 * DEPLOY MARKER (long on purpose, so a grep for it cannot be a false negative):
 * "vendor trust checklist recomputes the verdict tier from the live denominator"
 */
(function () {
  'use strict';

  var section = document.querySelector('[data-component="tool-ai-vendor-trust-checklist"]');
  if (!section) return;

  var boxes = Array.prototype.slice.call(
    section.querySelectorAll('input[type="checkbox"][data-vtc-item]'));

  var naBox     = section.querySelector('#vtc-na-sector');
  var resetEl   = section.querySelector('#vtc-reset');
  var countEl   = section.querySelector('#vtc-score-count');
  var totalEl   = section.querySelector('#vtc-score-total');
  var meterEl   = section.querySelector('#vtc-meter-fill');
  var boxEl     = section.querySelector('#vtc-verdict-box');
  var verdictEl = section.querySelector('#vtc-verdict');
  var detailEl  = section.querySelector('#vtc-verdict-detail');
  var gapsEl    = section.querySelector('#vtc-gaps');

  if (!boxes.length || !countEl || !totalEl || !boxEl || !verdictEl) return;

  // Thresholds are ratios, not absolute counts, because the denominator becomes
  // 11 when the sector certification is marked not applicable. 9/12 and 9/11
  // both read as strong; 4/12 and 4/11 both read as gaps.
  function tierFor(ratio) {
    if (ratio >= 0.75) return 'strong';
    if (ratio >= 0.4) return 'mid';
    return 'low';
  }

  function isSector(box) {
    return box.getAttribute('data-vtc-optional') === 'sector';
  }

  function recompute() {
    var naOn = !!(naBox && naBox.checked);
    var i;

    // The excluded item is disabled AND cleared. Clearing matters: a tick left
    // behind when the item leaves the denominator would be counted again the
    // moment it came back.
    for (i = 0; i < boxes.length; i++) {
      if (!isSector(boxes[i])) continue;
      boxes[i].disabled = naOn;
      if (naOn) boxes[i].checked = false;
      var row = boxes[i].parentNode;
      if (row && row.classList) row.classList.toggle('vtc-item--na', naOn);
    }

    var counted = 0;
    var ticked = 0;
    for (i = 0; i < boxes.length; i++) {
      if (naOn && isSector(boxes[i])) continue;
      counted++;
      if (boxes[i].checked) ticked++;
    }

    // Guard the denominator. Zero cannot happen with the shipped markup, which
    // is precisely why it is handled rather than assumed: a NaN ratio would
    // still produce a confident-looking verdict.
    var ratio = counted > 0 ? ticked / counted : 0;
    var tier = tierFor(ratio);

    countEl.textContent = String(ticked);
    totalEl.textContent = String(counted);
    if (meterEl) meterEl.style.width = Math.round(ratio * 100) + '%';

    boxEl.setAttribute('data-tier', tier);
    verdictEl.textContent = boxEl.getAttribute('data-' + tier + '-label') || '';
    if (detailEl) {
      detailEl.textContent = boxEl.getAttribute('data-' + tier + '-detail') || '';
    }

    if (gapsEl) {
      var remaining = counted - ticked;
      if (remaining === 0) {
        gapsEl.textContent = 'Nothing left unticked. That is the strongest position an outside check can reach, which is not the same as a guarantee.';
      } else if (remaining === 1) {
        gapsEl.textContent = '1 unticked item to take to the vendor as a question.';
      } else {
        gapsEl.textContent = remaining + ' unticked items to take to the vendor as questions.';
      }
    }
  }

  for (var k = 0; k < boxes.length; k++) {
    boxes[k].addEventListener('change', recompute);
  }
  if (naBox) naBox.addEventListener('change', recompute);
  if (resetEl) {
    resetEl.addEventListener('click', function () {
      for (var m = 0; m < boxes.length; m++) boxes[m].checked = false;
      if (naBox) naBox.checked = false;
      recompute();
    });
  }

  recompute();
})();
