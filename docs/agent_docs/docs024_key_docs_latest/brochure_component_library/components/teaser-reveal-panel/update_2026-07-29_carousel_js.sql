\set ON_ERROR_STOP on
-- Re-apply teaser-reveal-panel's behaviour.js to the live js_snippets row
-- (carousel prev/next navigation added; deep-link open now also scrolls
-- horizontally, inline:'center', since a card can be off-screen to the side
-- in a carousel, not just below the fold).
BEGIN;
UPDATE js_snippets
   SET js_content = $JS$/* teaser-reveal-panel — URL addressability + carousel navigation for the
 * teaser-detail-deeplink shape.
 *
 * PROGRESSIVE ENHANCEMENT, deliberately. The reveal itself is native
 * <details>/<summary>: with this file blocked, missing or broken, every teaser
 * still opens and closes (via native disclosure) and the track still scrolls
 * (via native scroll-snap + touch/trackpad), and nothing renders as a dead
 * control. The prev/next arrows are the one thing that stops working — that is
 * an acceptable degradation, not a broken one, because scroll-snap covers the
 * same ground without them.
 *
 * What this adds, and only this:
 *   - opening a teaser pushes ?open=<key> so the open state is shareable;
 *   - a cold load carrying ?open=<key> opens that teaser and scrolls it into
 *     view, horizontally AND vertically (the panel is a carousel: the target
 *     card may be off-screen to the side, not just below the fold);
 *   - back/forward reproduce the state the address names;
 *   - opening one teaser closes its siblings (a panel, not an accordion pile);
 *   - prev/next arrow buttons scroll the track by one card, reusing the exact
 *     goTo/nearestIndex approach already proven on hero-card-carousel (same
 *     component family, same site) rather than inventing a second one;
 *   - the arrows carry an aria-live announcement ("Card 3 of 6"), same pattern.
 *
 * The <details> body is always in the DOM. That is intentional: it is
 * assertive prose, so the claims gate and a crawler must both be able to read
 * it. A JS-populated region would hide site copy from the only checkers that
 * read it.
 */
(function () {
  'use strict';
  var panels = document.querySelectorAll('[data-component="teaser-reveal-panel"]');
  if (!panels.length) return;
  var reduceMotion = window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches;

  Array.prototype.forEach.call(panels, function (panel) {
    var param = panel.getAttribute('data-trp-param') || 'open';
    var track = panel.querySelector('[data-trp-track]');
    var cards = panel.querySelectorAll('details.trp__card[data-trp-key]');
    var slides = Array.prototype.slice.call(panel.querySelectorAll('[data-trp-slide]'));
    var prevBtn = panel.querySelector('[data-trp-prev]');
    var nextBtn = panel.querySelector('[data-trp-next]');
    var live = panel.querySelector('[data-trp-live]');

    function keyOf(card) { return card.getAttribute('data-trp-key'); }
    function behavior() { return reduceMotion ? 'auto' : 'smooth'; }

    // --- Deep-link addressability (unchanged in spirit from the first build) ---
    if (cards.length) {
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
            // Bring the opened card fully into view within the horizontally
            // scrolling track, not just vertically on the page -- a card can
            // be the partially-visible "next" one at the carousel's edge.
            card.scrollIntoView({ block: 'nearest', inline: 'center', behavior: behavior() });
          } else if (new URL(window.location.href).searchParams.get(param) === keyOf(card)) {
            writeAddress(null, true);
          }
        });
      });

      window.addEventListener('popstate', function () { applyFromAddress(); });

      var opened = applyFromAddress();
      if (opened) {
        var target = panel.querySelector('details.trp__card[data-trp-key="' + opened.replace(/"/g, '') + '"]');
        if (target) {
          target.scrollIntoView({ block: 'center', inline: 'center', behavior: behavior() });
        }
      }
    }

    // --- Carousel navigation: same goTo/nearestIndex shape as hero-card-carousel ---
    if (track && slides.length > 1 && (prevBtn || nextBtn)) {
      var current = 0;

      function goTo(i, announce) {
        current = (i + slides.length) % slides.length;
        var trackRect = track.getBoundingClientRect();
        var slideRect = slides[current].getBoundingClientRect();
        track.scrollBy({ left: slideRect.left - trackRect.left, behavior: behavior() });
        if (announce && live) live.textContent = 'Card ' + (current + 1) + ' of ' + slides.length;
      }

      function nearestIndex() {
        var trackLeft = track.getBoundingClientRect().left;
        var best = 0, bestDist = Infinity;
        for (var i = 0; i < slides.length; i++) {
          var d = Math.abs(slides[i].getBoundingClientRect().left - trackLeft);
          if (d < bestDist) { bestDist = d; best = i; }
        }
        return best;
      }

      if (prevBtn) prevBtn.addEventListener('click', function () { goTo(current - 1, true); });
      if (nextBtn) nextBtn.addEventListener('click', function () { goTo(current + 1, true); });

      panel.addEventListener('keydown', function (e) {
        if (e.key === 'ArrowRight') { goTo(current + 1, true); }
        else if (e.key === 'ArrowLeft') { goTo(current - 1, true); }
      });

      var scrollT = null;
      track.addEventListener('scroll', function () {
        if (scrollT) clearTimeout(scrollT);
        scrollT = setTimeout(function () { current = nearestIndex(); }, 120);
      }, { passive: true });
    } else {
      // Fewer than 2 cards, or no track/arrows found: nothing to navigate.
      if (prevBtn) prevBtn.style.display = 'none';
      if (nextBtn) nextBtn.style.display = 'none';
    }
  });
})();
$JS$
 WHERE name = 'teaser-reveal-panel';
COMMIT;
SELECT name, is_active, length(js_content) AS bytes,
       js_content LIKE '%data-trp-prev%' AS has_carousel_nav
  FROM js_snippets WHERE name = 'teaser-reveal-panel';
