/* stat-band — accessible count-up for code-rendered stats.
   Animates the numeric part of a stat into view, then restores the exact
   authored value. The real value is always in the DOM + on aria-label, so
   screen readers read the true figure; the count-up is purely visual and is
   skipped under prefers-reduced-motion. Never invents digits — it only counts
   up to whatever real value was rendered. Idempotent, no dependencies. */
(function () {
  "use strict";
  if (window.__statBandInit) return;
  window.__statBandInit = true;
  var reduce = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  function animate(el) {
    var original = el.getAttribute("data-final");
    if (original === null) { original = el.textContent.trim(); el.setAttribute("data-final", original); }
    // Match: optional prefix, one number (with thousands/decimals), optional suffix.
    var m = original.match(/^(\D*)(\d[\d,]*(?:\.\d+)?)(\D*)$/);
    if (!m || reduce) { el.textContent = original; return; }
    var prefix = m[1], numStr = m[2].replace(/,/g, ""), suffix = m[3];
    var target = parseFloat(numStr);
    var decimals = (numStr.split(".")[1] || "").length;
    var dur = 1100, start = null;
    function fmt(v) {
      return prefix + v.toFixed(decimals).replace(/\B(?=(\d{3})+(?!\d))/g, ",") + suffix;
    }
    function step(ts) {
      if (!start) start = ts;
      var p = Math.min((ts - start) / dur, 1);
      var eased = 1 - Math.pow(1 - p, 3);
      if (p < 1) { el.textContent = fmt(target * eased); requestAnimationFrame(step); }
      else { el.textContent = original; } // restore exact authored value
    }
    el.textContent = fmt(0);
    requestAnimationFrame(step);
  }

  function init() {
    var els = Array.prototype.slice.call(document.querySelectorAll(".stat-band [data-countup]"));
    if (!els.length) return;
    if (!("IntersectionObserver" in window)) { els.forEach(animate); return; }
    var io = new IntersectionObserver(function (entries) {
      entries.forEach(function (en) {
        if (en.isIntersecting) { animate(en.target); io.unobserve(en.target); }
      });
    }, { threshold: 0.4 });
    els.forEach(function (el) { io.observe(el); });
  }

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", init);
  else init();
})();
