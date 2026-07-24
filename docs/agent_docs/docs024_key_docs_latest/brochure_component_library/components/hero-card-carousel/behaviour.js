/* hero-card-carousel — overlaid prev/next arrows + optional accessible
   auto-advance. Auto-advance is OFF by default (data-hcc-autoplay="false"); when
   on it is pausable, stops on hover/keyboard focus, and respects
   prefers-reduced-motion. The arrows always work. Multi-instance, no dependencies. */
(function () {
  "use strict";
  var ROTATE_MS = 6000;
  var reduceMotion = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  function initCarousel(root) {
    var track = root.querySelector("[data-hcc-track]");
    if (!track) return;
    var slides = Array.prototype.slice.call(root.querySelectorAll("[data-hcc-slide]"));
    var prevBtn = root.querySelector("[data-hcc-prev]");
    var nextBtn = root.querySelector("[data-hcc-next]");
    var pauseBtn = root.querySelector("[data-hcc-pause]");
    var pauseIcon = root.querySelector("[data-hcc-pause-icon]");
    var live = root.querySelector("[data-hcc-live]");
    var autoplay = root.getAttribute("data-hcc-autoplay") === "true";

    if (slides.length < 2) {
      [prevBtn, nextBtn, pauseBtn].forEach(function (b) { if (b) b.style.display = "none"; });
      return;
    }

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
      if (!autoplay || reduceMotion || paused || suspended) return;
      timer = setInterval(function () { goTo(current + 1, false); }, ROTATE_MS);
    }

    // Overlaid arrows — always active. Left = previous, right = next.
    if (prevBtn) prevBtn.addEventListener("click", function () { goTo(current - 1, true); });
    if (nextBtn) nextBtn.addEventListener("click", function () { goTo(current + 1, true); });

    // Pause control only exists when autoplay is on.
    if (autoplay && pauseBtn) {
      pauseBtn.addEventListener("click", function () {
        paused = !paused;
        pauseBtn.setAttribute("aria-label", paused ? "Start automatic rotation" : "Pause automatic rotation");
        if (pauseIcon) pauseIcon.innerHTML = paused ? "&#9654;" : "&#10073;&#10073;";
        startTimer();
      });
    }

    // Auto-advance courtesy behaviours (no-ops when autoplay is off).
    root.addEventListener("pointerenter", function () { suspended = true; stopTimer(); });
    root.addEventListener("pointerleave", function () { suspended = false; startTimer(); });
    root.addEventListener("focusin", function () { suspended = true; stopTimer(); });
    root.addEventListener("focusout", function () {
      if (!root.contains(document.activeElement)) { suspended = false; startTimer(); }
    });

    root.addEventListener("keydown", function (e) {
      if (e.key === "ArrowRight") { e.preventDefault(); goTo(current + 1, true); }
      else if (e.key === "ArrowLeft") { e.preventDefault(); goTo(current - 1, true); }
    });

    var scrollT = null;
    track.addEventListener("scroll", function () {
      if (scrollT) clearTimeout(scrollT);
      scrollT = setTimeout(function () { current = nearestIndex(); }, 120);
    }, { passive: true });

    if (autoplay) {
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
  }

  function initAll() {
    Array.prototype.slice.call(document.querySelectorAll(".hero-card-carousel[data-component='hero-card-carousel']")).forEach(initCarousel);
  }
  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", initAll);
  else initAll();
})();
