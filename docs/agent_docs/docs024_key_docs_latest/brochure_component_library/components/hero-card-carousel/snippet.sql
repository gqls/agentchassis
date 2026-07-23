\set ON_ERROR_STOP on
BEGIN;
-- correct JS-delivery lane: js_snippets → bundled into the loaded /assets/js/snippets.js
INSERT INTO js_snippets (id, name, description, js_content, applies_to, semantic_tags, is_active)
VALUES (gen_random_uuid(), 'hero-card-carousel',
  'Accessible auto-advance for the hero-card-carousel component (pausable, keyboard-safe, reduced-motion aware).',
  $JS$
/* hero-card-carousel — accessible auto-advancing card carousel.
   WCAG 2.2.2: auto-rotation is pausable, stops on hover and keyboard focus,
   respects prefers-reduced-motion, and every control is a real <button>.
   Self-contained, supports multiple instances, no dependencies. */
(function () {
  "use strict";
  var ROTATE_MS = 6000;
  var reduceMotion = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  function initCarousel(root) {
    var track = root.querySelector("[data-hcc-track]");
    if (!track) return;
    var slides = Array.prototype.slice.call(root.querySelectorAll("[data-hcc-slide]"));
    if (slides.length < 2) {
      var c = root.querySelector("[data-hcc-controls]");
      if (c) c.style.display = "none";
      return;
    }
    var pauseBtn = root.querySelector("[data-hcc-pause]");
    var pauseIcon = root.querySelector("[data-hcc-pause-icon]");
    var prevBtn = root.querySelector("[data-hcc-prev]");
    var nextBtn = root.querySelector("[data-hcc-next]");
    var live = root.querySelector("[data-hcc-live]");

    var current = 0;
    var paused = false;      // user pressed pause
    var suspended = false;   // hover / focus (temporary)
    var timer = null;

    function behavior() { return reduceMotion ? "auto" : "smooth"; }

    function goTo(i, announce) {
      current = (i + slides.length) % slides.length;
      var trackRect = track.getBoundingClientRect();
      var slideRect = slides[current].getBoundingClientRect();
      track.scrollBy({ left: slideRect.left - trackRect.left, behavior: behavior() });
      if (announce && live) live.textContent = "Card " + (current + 1) + " of " + slides.length;
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

    function stopTimer() { if (timer) { clearInterval(timer); timer = null; } }
    function startTimer() {
      stopTimer();
      if (reduceMotion || paused || suspended) return;
      timer = setInterval(function () { goTo(current + 1, false); }, ROTATE_MS);
    }

    function setPaused(p) {
      paused = p;
      if (pauseBtn) pauseBtn.setAttribute("aria-label", p ? "Start automatic rotation" : "Pause automatic rotation");
      if (pauseIcon) pauseIcon.innerHTML = p ? "&#9654;" : "&#10073;&#10073;";
      startTimer();
    }

    if (pauseBtn) pauseBtn.addEventListener("click", function () { setPaused(!paused); });
    if (prevBtn) prevBtn.addEventListener("click", function () { goTo(current - 1, true); });
    if (nextBtn) nextBtn.addEventListener("click", function () { goTo(current + 1, true); });

    // Pause on hover and on keyboard focus entering the carousel; resume on leave.
    root.addEventListener("pointerenter", function () { suspended = true; stopTimer(); });
    root.addEventListener("pointerleave", function () { suspended = false; startTimer(); });
    root.addEventListener("focusin", function () { suspended = true; stopTimer(); });
    root.addEventListener("focusout", function () {
      if (!root.contains(document.activeElement)) { suspended = false; startTimer(); }
    });

    // Keyboard arrows when the track (or a control) has focus.
    root.addEventListener("keydown", function (e) {
      if (e.key === "ArrowRight") { e.preventDefault(); goTo(current + 1, true); }
      else if (e.key === "ArrowLeft") { e.preventDefault(); goTo(current - 1, true); }
    });

    // Keep `current` in sync when the user swipes/scrolls manually.
    var scrollT = null;
    track.addEventListener("scroll", function () {
      if (scrollT) clearTimeout(scrollT);
      scrollT = setTimeout(function () { current = nearestIndex(); }, 120);
    }, { passive: true });

    // Pause when the carousel scrolls out of view (don't rotate off-screen).
    if ("IntersectionObserver" in window) {
      new IntersectionObserver(function (entries) {
        entries.forEach(function (en) {
          if (en.isIntersecting) { suspended = false; startTimer(); }
          else { suspended = true; stopTimer(); }
        });
      }, { threshold: 0.25 }).observe(root);
    } else {
      startTimer();
    }
  }

  function initAll() {
    Array.prototype.slice.call(document.querySelectorAll(".hero-card-carousel[data-component='hero-card-carousel']")).forEach(initCarousel);
  }
  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", initAll);
  else initAll();
})();
$JS$,
  '["hero-card-carousel"]'::jsonb, '["carousel","interactive","accessibility"]'::jsonb, true)
RETURNING name, applies_to, is_active, length(js_content) AS js_len;
-- the /tools/assets lane doesn't inject a <script> tag (bugs_open/041 class); drop it to keep one source of truth
UPDATE content_components SET js_content = NULL WHERE function='hero-card-carousel' AND is_active
RETURNING function, (js_content IS NULL) AS js_content_cleared;
COMMIT;
