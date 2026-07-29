/* teaser-reveal-panel — URL addressability for the teaser-detail-deeplink shape.
 *
 * PROGRESSIVE ENHANCEMENT, deliberately. The reveal itself is native
 * <details>/<summary>: with this file blocked, missing or broken, every teaser
 * still opens and closes, and nothing renders as a dead control. That is the
 * pattern's own absence rule applied to its own JavaScript.
 *
 * What this adds, and only this:
 *   - opening a teaser pushes ?open=<key> so the open state is shareable;
 *   - a cold load carrying ?open=<key> opens that teaser and scrolls it into view;
 *   - back/forward reproduce the state the address names;
 *   - opening one teaser closes its siblings (a panel, not an accordion pile).
 *
 * The <details> body is always in the DOM. That is intentional: it is assertive
 * prose, so the claims gate and a crawler must both be able to read it. A
 * JS-populated region would hide site copy from the only checkers that read it.
 */
(function () {
  'use strict';
  var panels = document.querySelectorAll('[data-component="teaser-reveal-panel"]');
  if (!panels.length) return;

  Array.prototype.forEach.call(panels, function (panel) {
    var param = panel.getAttribute('data-trp-param') || 'open';
    var cards = panel.querySelectorAll('details.trp__card[data-trp-key]');
    if (!cards.length) return;

    function keyOf(card) { return card.getAttribute('data-trp-key'); }

    function writeAddress(key, push) {
      if (!window.history || !window.history.pushState) return;
      var url = new URL(window.location.href);
      if (key) { url.searchParams.set(param, key); } else { url.searchParams.delete(param); }
      var next = url.pathname + (url.search || '') + url.hash;
      if (push) { window.history.pushState({ trp: key || null }, '', next); }
      else { window.history.replaceState({ trp: key || null }, '', next); }
    }

    function applyFromAddress() {
      var want = new URL(window.location.href).searchParams.get(param);
      Array.prototype.forEach.call(cards, function (card) {
        card.open = (want !== null && keyOf(card) === want);
      });
      return want;
    }

    Array.prototype.forEach.call(cards, function (card) {
      card.addEventListener('toggle', function () {
        if (card.open) {
          Array.prototype.forEach.call(cards, function (other) {
            if (other !== card) { other.open = false; }
          });
          writeAddress(keyOf(card), true);
        } else if (new URL(window.location.href).searchParams.get(param) === keyOf(card)) {
          writeAddress(null, true);
        }
      });
    });

    window.addEventListener('popstate', function () { applyFromAddress(); });

    var opened = applyFromAddress();
    if (opened) {
      var target = panel.querySelector('details.trp__card[data-trp-key="' + opened.replace(/"/g, '') + '"]');
      if (target && target.scrollIntoView) {
        target.scrollIntoView({ block: 'center', behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth' });
      }
    }
  });
})();
