(function() {
  var section = document.querySelector('[data-component="gauntlet-interface"]');
  if (!section) return;

  // --- Objective checkboxes & progress ---
  var objectives = section.querySelectorAll('[data-gi-obj]');
  var fill = section.querySelector('[data-gi-fill]');
  var pct = section.querySelector('[data-gi-pct]');
  var track = section.querySelector('[data-gi-track]');

  function updateProgress() {
    var total = objectives.length;
    if (!total) return;
    var done = section.querySelectorAll('[data-gi-obj].is-complete').length;
    var p = Math.round((done / total) * 100);
    if (fill) fill.style.width = p + '%';
    if (pct) pct.textContent = p + '%';
    if (track) track.setAttribute('aria-valuenow', p);
  }

  objectives.forEach(function(obj) {
    function toggle() {
      var isComplete = obj.classList.toggle('is-complete');
      obj.setAttribute('aria-pressed', isComplete ? 'true' : 'false');
      updateProgress();
    }
    obj.addEventListener('click', toggle);
    obj.addEventListener('keydown', function(e) {
      if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); toggle(); }
    });
  });

  updateProgress();

  // --- Countdown timer ---
  var timerDisplay = section.querySelector('[data-gi-timer]');
  var startBtn = section.querySelector('[data-gi-timer-start]');
  var resetBtn = section.querySelector('[data-gi-timer-reset]');
  var TIMER_DURATION = 20 * 60; // 20 minutes in seconds
  var remaining = TIMER_DURATION;
  var interval = null;
  var running = false;

  function formatTime(s) {
    var m = Math.floor(s / 60);
    var sec = s % 60;
    return (m < 10 ? '0' : '') + m + ':' + (sec < 10 ? '0' : '') + sec;
  }

  function renderTimer() {
    if (timerDisplay) {
      timerDisplay.textContent = formatTime(remaining);
      if (remaining <= 60) {
        timerDisplay.classList.add('is-urgent');
      } else {
        timerDisplay.classList.remove('is-urgent');
      }
    }
  }

  function startTimer() {
    if (running) return;
    running = true;
    if (startBtn) startBtn.classList.add('is-active');
    interval = setInterval(function() {
      if (remaining > 0) {
        remaining--;
        renderTimer();
      } else {
        clearInterval(interval);
        running = false;
        if (startBtn) startBtn.classList.remove('is-active');
      }
    }, 1000);
  }

  function resetTimer() {
    clearInterval(interval);
    running = false;
    remaining = TIMER_DURATION;
    renderTimer();
    if (startBtn) startBtn.classList.remove('is-active');
  }

  if (startBtn) startBtn.addEventListener('click', function() {
    if (running) {
      clearInterval(interval);
      running = false;
      startBtn.classList.remove('is-active');
    } else {
      startTimer();
    }
  });

  if (resetBtn) resetBtn.addEventListener('click', resetTimer);

  renderTimer();

  // --- Primary CTA: begin the challenge (start the clock, bring the task into view) ---
  var enterBtn = section.querySelector('[data-gi-enter-btn]');
  if (enterBtn) enterBtn.addEventListener('click', function() {
    startTimer();
    var panel = section.querySelector('.gi-challenge-panel');
    if (panel && panel.scrollIntoView) {
      panel.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
    var firstObj = section.querySelector('[data-gi-obj]');
    if (firstObj && firstObj.focus) firstObj.focus();
  });

  // --- Secondary CTA: reveal the rules (the #gi-rules anchor works without JS;
  //     this only adds a smooth scroll and a brief highlight) ---
  var rulesBtn = section.querySelector('[data-gi-rules-btn]');
  var rulesCard = section.querySelector('#gi-rules');
  if (rulesBtn && rulesCard) rulesBtn.addEventListener('click', function(e) {
    if (rulesCard.scrollIntoView) {
      e.preventDefault();
      rulesCard.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
    rulesCard.style.transition = 'box-shadow 0.3s';
    rulesCard.style.boxShadow = '0 0 0 2px var(--color-accent, #fbbf24)';
    setTimeout(function() { rulesCard.style.boxShadow = ''; }, 1600);
  });

})();
