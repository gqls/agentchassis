/* gauntlet-interface — a real timed debate against a live AI opponent.
 *
 * Every state change below is bound to an API response. Nothing here can be
 * ticked, faked or advanced by clicking: the three objectives complete only as
 * side-effects of /round, /position and /defend returning 200, and the clock
 * only starts when a round genuinely starts. The previous version of this file
 * simulated all of it client-side.
 *
 * Endpoints (verified live 2026-07-26):
 *   POST {API}/api/v1/tools/gauntlet/round    -> {round_id, provocation{...}}
 *   POST {API}/api/v1/tools/gauntlet/position -> {counter_position, challenge}
 *   POST {API}/api/v1/tools/gauntlet/defend   -> {verdict, reasons}
 * The AI calls are SLOW — measured 8–18s each — so every wait shows a running
 * elapsed counter rather than a silent spinner.
 *
 * Degraded mode is honest: 503 is the engine-offline status (Cloudflare
 * replaces raw origin 502 bodies, so a 502-shaped assumption breaks in
 * production), and on any failure the clock does not start and no objective
 * is marked.
 */
(function () {
  "use strict";

  var API = "https://tools.apis.uk";
  var ROUND_SECONDS = 20 * 60;
  var MAX_CHARS = 2000;

  var section = document.querySelector('[data-component="gauntlet-interface"]');
  if (!section) return;

  var el = {
    status: section.querySelector("[data-gi-status]"),
    roundState: section.querySelector("[data-gi-round-state]"),
    eyebrow: section.querySelector("[data-gi-challenge-eyebrow]"),
    title: section.querySelector("[data-gi-challenge-title]"),
    body: section.querySelector("[data-gi-challenge-body]"),
    enter: section.querySelector("[data-gi-enter-btn]"),
    panel: section.querySelector(".gi-challenge-panel"),
    objectives: section.querySelectorAll("[data-gi-obj]"),
    pct: section.querySelector("[data-gi-pct]"),
    fill: section.querySelector("[data-gi-fill]"),
    track: section.querySelector("[data-gi-track]"),
    timer: section.querySelector("[data-gi-timer]"),
    positionInput: section.querySelector("[data-gi-position-input]"),
    positionSubmit: section.querySelector("[data-gi-position-submit]"),
    positionCount: section.querySelector("[data-gi-position-count]"),
    opponentBlock: section.querySelector("[data-gi-opponent-block]"),
    opponentPosition: section.querySelector("[data-gi-opponent-position]"),
    opponentChallenge: section.querySelector("[data-gi-opponent-challenge]"),
    defenceInput: section.querySelector("[data-gi-defence-input]"),
    defenceSubmit: section.querySelector("[data-gi-defence-submit]"),
    defenceCount: section.querySelector("[data-gi-defence-count]"),
    verdictBlock: section.querySelector("[data-gi-verdict-block]"),
    verdict: section.querySelector("[data-gi-verdict]"),
    verdictReasons: section.querySelector("[data-gi-verdict-reasons]")
  };

  // Round state — per-session, in memory only, never persisted or invented.
  var state = {
    roundId: null,
    positionFiled: false,
    verdictIn: false,
    busy: false,
    deadline: null,
    tick: null
  };

  // ── round persistence ────────────────────────────────────────────────────
  //
  // WHY: reported live 2026-07-28 — a visitor part-way through answering the
  // challenge lost it, then found Send Defence silently refusing. Reproduced:
  // everything lived in page memory only, so ANY reload (an accidental refresh,
  // a mobile tab evicted while switching apps, back/forward) destroyed a round
  // that was still LIVE on the server. The visitor lost what they had typed and
  // had no route back to their own round.
  //
  // This is NOT the localStorage pattern removed from the Arena. That faked a
  // submission which never went anywhere. This resumes a REAL server-side round
  // by its real id — it stores nothing that is not already true.
  //
  // sessionStorage, not localStorage: a round is 20 minutes and tab-scoped, so
  // it should not outlive the tab or leak into another one.
  var STORE = "vonc_gauntlet_round_v1";

  function saveRound() {
    try {
      if (!state.roundId) { sessionStorage.removeItem(STORE); return; }
      sessionStorage.setItem(STORE, JSON.stringify({
        roundId: state.roundId,
        deadline: state.deadline,
        positionFiled: state.positionFiled,
        verdictIn: state.verdictIn,
        provocation: state.provocation || null,
        counter: el.opponentPosition ? el.opponentPosition.textContent : "",
        challenge: el.opponentChallenge ? el.opponentChallenge.textContent : "",
        verdict: el.verdict ? el.verdict.textContent : "",
        reasons: el.verdictReasons ? el.verdictReasons.textContent : "",
        draftPosition: el.positionInput ? el.positionInput.value : "",
        draftDefence: el.defenceInput ? el.defenceInput.value : "",
        // NodeList has forEach but no map — slice it first.
        objectives: Array.prototype.slice.call(el.objectives || []).map(function (o) {
          return o.classList.contains("is-complete");
        })
      }));
    } catch (e) { /* private mode: degrade to the old behaviour, never break */ }
  }

  function restoreRound() {
    var raw;
    try { raw = sessionStorage.getItem(STORE); } catch (e) { return false; }
    if (!raw) return false;
    var d;
    try { d = JSON.parse(raw); } catch (e) { return false; }
    if (!d || !d.roundId) return false;

    // Drafts come back even when the clock has gone: they are the only thing
    // here the visitor wrote themselves, and we should never be what loses them.
    if (el.positionInput && d.draftPosition) el.positionInput.value = d.draftPosition;
    if (el.defenceInput && d.draftDefence) el.defenceInput.value = d.draftDefence;
    if (d.provocation) renderProvocation(d.provocation);

    if (!d.deadline || d.deadline <= Date.now()) {
      try { sessionStorage.removeItem(STORE); } catch (e) {}
      setStatus(
        "Your previous round's clock ran out while the page was away. What you " +
          "typed is still here — press Enter the Gauntlet to start a fresh round.",
        "error"
      );
      setRoundState("No round yet");
      return false;
    }

    state.roundId = d.roundId;
    state.deadline = d.deadline;
    state.positionFiled = !!d.positionFiled;
    state.verdictIn = !!d.verdictIn;

    if (d.counter && el.opponentPosition) el.opponentPosition.textContent = d.counter;
    if (d.challenge && el.opponentChallenge) el.opponentChallenge.textContent = d.challenge;
    if ((d.counter || d.challenge) && el.opponentBlock) el.opponentBlock.classList.remove("is-empty");
    if (d.verdict && el.verdict) el.verdict.textContent = d.verdict;
    if (d.reasons && el.verdictReasons) el.verdictReasons.textContent = d.reasons;
    if ((d.verdict || d.reasons) && el.verdictBlock) el.verdictBlock.classList.remove("is-empty");
    (d.objectives || []).forEach(function (done, i) {
      if (done && el.objectives && el.objectives[i]) el.objectives[i].classList.add("is-complete");
    });
    updateProgress();

    if (state.verdictIn) {
      setRoundState("Round closed");
      renderClock();
    } else {
      state.tick = setInterval(renderClock, 1000);
      renderClock();
      setRoundState("Round live");
      setStatus("Picked your round back up — the clock kept running while you were away.", "live");
    }
    return true;
  }

  // ── status line ──────────────────────────────────────────────────────────

  var waitTimer = null;

  function setStatus(text, kind) {
    stopWaitCounter();
    if (!el.status) return;
    el.status.textContent = text;
    el.status.className = "gi-status" + (kind ? " is-" + kind : "");
  }

  // A 15-second wait with no feedback reads as a hang. Count it out loud.
  function startWaitCounter(text) {
    stopWaitCounter();
    if (!el.status) return;
    var started = Date.now();
    el.status.className = "gi-status is-working";
    var render = function () {
      var secs = Math.round((Date.now() - started) / 1000);
      el.status.textContent = text + " (" + secs + "s)";
    };
    render();
    waitTimer = setInterval(render, 1000);
  }

  function stopWaitCounter() {
    if (waitTimer) {
      clearInterval(waitTimer);
      waitTimer = null;
    }
  }

  function setRoundState(text) {
    if (el.roundState) el.roundState.textContent = text;
  }

  // ── objectives & progress ────────────────────────────────────────────────
  // index 0 position filed · 1 defence sent · 2 verdict in before the clock ran out

  function completeObjective(i) {
    if (el.objectives && el.objectives[i]) {
      el.objectives[i].classList.add("is-complete");
    }
    updateProgress();
  }

  function updateProgress() {
    var total = el.objectives ? el.objectives.length : 0;
    if (!total) return;
    var done = section.querySelectorAll("[data-gi-obj].is-complete").length;
    var p = Math.round((done / total) * 100);
    if (el.fill) el.fill.style.width = p + "%";
    if (el.pct) el.pct.textContent = p + "% Complete";
    if (el.track) el.track.setAttribute("aria-valuenow", p);
  }

  // ── clock ────────────────────────────────────────────────────────────────

  function remainingSeconds() {
    if (!state.deadline) return null;
    return Math.max(0, Math.round((state.deadline - Date.now()) / 1000));
  }

  function renderClock() {
    var r = remainingSeconds();
    if (r === null || !el.timer) return;
    var m = Math.floor(r / 60);
    var s = r % 60;
    el.timer.textContent = m + ":" + (s < 10 ? "0" : "") + s;
    el.timer.classList.toggle("is-urgent", r > 0 && r <= 120);
    if (r === 0) {
      stopClock();
      el.timer.classList.remove("is-urgent");
      setRoundState("Time up");
      if (!state.verdictIn) {
        setStatus(
          "The clock ran out. You can still send a defence and read the verdict, " +
            "but the third objective — a verdict before time — is gone for this round.",
          "error"
        );
      }
    }
  }

  function startClock() {
    stopClock();
    state.deadline = Date.now() + ROUND_SECONDS * 1000;
    saveRound();
    renderClock();
    state.tick = setInterval(renderClock, 1000);
  }

  function stopClock() {
    if (state.tick) {
      clearInterval(state.tick);
      state.tick = null;
    }
  }

  // ── network ──────────────────────────────────────────────────────────────

  function post(path, payload) {
    return fetch(API + "/api/v1/tools/gauntlet/" + path, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload || {})
    }).then(function (res) {
      return res
        .json()
        .catch(function () {
          return {};
        })
        .then(function (data) {
          if (!res.ok) {
            var err = new Error(data && data.error ? data.error : "HTTP " + res.status);
            err.status = res.status;
            throw err;
          }
          return data;
        });
    });
  }

  // One place decides what a visitor is told, so no failure is silent and none
  // is dressed up as something it isn't.
  function explain(err, what) {
    if (!err) return "Something went wrong " + what + ". Try again.";
    switch (err.status) {
      case 503:
        return "The AI opponent is offline — try again later. Nothing was lost; your text is still here.";
      case 404:
        return "This round has expired. Press Enter the Gauntlet to start a fresh one.";
      case 429:
        return "Too many requests from this connection. Wait a minute, then try again.";
      case 413:
        return "That is longer than " + MAX_CHARS + " characters. Trim it and send again.";
      case 400:
        return "That request was rejected as malformed. Check the text and try again.";
      case 403:
        return "This page is not authorised to reach the debate engine.";
      default:
        return "Could not reach the debate engine " + what + ". Check your connection and try again.";
    }
  }

  function busy(button, on, label) {
    state.busy = on;
    if (!button) return;
    button.setAttribute("aria-busy", on ? "true" : "false");
    if (on) {
      button.dataset.giLabel = button.textContent;
      if (label) button.textContent = label;
    } else if (button.dataset.giLabel) {
      button.textContent = button.dataset.giLabel;
      delete button.dataset.giLabel;
    }
  }

  // ── provocation rendering ────────────────────────────────────────────────

  // The provocation is our own authored JSON (the /round endpoint returns the
  // `today` object verbatim), so <em> in the headline is ours. Everything the
  // AI writes goes in as text, never markup.
  function renderProvocation(p) {
    if (p) state.provocation = p;
    if (!p) return;
    if (el.eyebrow && p.eyebrow) el.eyebrow.textContent = p.eyebrow;
    if (el.title && p.headline) el.title.innerHTML = p.headline;
    if (el.body && p.body) el.body.textContent = p.body;
  }

  // ── round lifecycle ──────────────────────────────────────────────────────

  function startRound() {
    if (state.busy) return Promise.reject(new Error("busy"));
    busy(el.enter, true);
    startWaitCounter("Drawing today's provocation and starting your clock…");
    return post("round", {})
      .then(function (data) {
        busy(el.enter, false);
        if (!data || !data.round_id) {
          throw new Error("no round_id");
        }
        state.roundId = data.round_id;
        state.positionFiled = false;
        state.verdictIn = false;
        renderProvocation(data.provocation);
        startClock();
        setRoundState("Round live");
        setStatus(
          "Your round is live and the clock is running. File a position on the " +
            "provocation above — the opponent answers, then you defend it.",
          "live"
        );
        if (el.panel && el.panel.scrollIntoView) {
          el.panel.scrollIntoView({ behavior: "smooth", block: "start" });
        }
        return data;
      })
      .catch(function (err) {
        busy(el.enter, false);
        // No round: no clock, no objective, no pretending.
        setStatus(explain(err, "starting the round"), "error");
        throw err;
      });
  }

  function submitPosition() {
    if (state.busy) return;
    var text = el.positionInput ? el.positionInput.value.trim() : "";
    if (!text) {
      setStatus("Type your position first — a sentence or two is enough.", "error");
      if (el.positionInput) el.positionInput.focus();
      return;
    }

    // Filing a position is the real entry point, so it starts the round itself
    // when there isn't one rather than sitting there as a control that does
    // nothing.
    var ready = state.roundId ? Promise.resolve() : startRound();

    ready
      .then(function () {
        busy(el.positionSubmit, true, "Sending…");
        startWaitCounter("The opponent is reading your position and writing a counter");
        return post("position", { round_id: state.roundId, position_text: text });
      })
      .then(function (data) {
        busy(el.positionSubmit, false);
        if (el.opponentPosition) el.opponentPosition.textContent = data.counter_position || "";
        if (el.opponentChallenge) el.opponentChallenge.textContent = data.challenge || "";
        if (el.opponentBlock) el.opponentBlock.classList.remove("is-empty");
        state.positionFiled = true;
        completeObjective(0);
        saveRound();
        setStatus(
          "Position filed and answered. Read the challenge, then send your defence " +
            "before the clock runs out.",
          "live"
        );
        if (el.defenceInput) el.defenceInput.focus();
      })
      .catch(function (err) {
        busy(el.positionSubmit, false);
        if (err && err.message === "busy") return;
        setStatus(explain(err, "filing your position"), "error");
      });
  }

  function submitDefence() {
    if (state.busy) return;
    var text = el.defenceInput ? el.defenceInput.value.trim() : "";
    if (!text) {
      setStatus("Type your defence first.", "error");
      if (el.defenceInput) el.defenceInput.focus();
      return;
    }
    if (!state.roundId || !state.positionFiled) {
      // This refused SILENTLY from the visitor's point of view: the only
      // explanation went to the status line at the top of the section while they
      // were at the bottom looking at the button. Reported live 2026-07-28 as
      // "clicking does nothing, no JS errors" — the button was working and was
      // explaining itself, just nowhere the visitor was looking.
      var why = state.roundId
        ? "The opponent has not answered yet — file your position first."
        : "There is no live round — press Enter the Gauntlet to start one. What you have typed is kept.";
      setStatus(why, "error");
      if (el.defenceNote) {
        el.defenceNote.textContent = why;
        el.defenceNote.classList.add("is-error");
      }
      if (el.positionInput && el.positionInput.scrollIntoView) {
        el.positionInput.scrollIntoView({ behavior: "smooth", block: "center" });
      }
      if (el.positionInput) el.positionInput.focus();
      return;
    }

    busy(el.defenceSubmit, true, "Sending…");
    if (el.defenceNote) { el.defenceNote.textContent = ""; el.defenceNote.classList.remove("is-error"); }
    startWaitCounter("The judge is reading the whole round and writing a verdict");

    post("defend", { round_id: state.roundId, defence_text: text })
      .then(function (data) {
        busy(el.defenceSubmit, false);
        if (el.verdict) el.verdict.textContent = data.verdict || "";
        if (el.verdictReasons) el.verdictReasons.textContent = data.reasons || "";
        if (el.verdictBlock) el.verdictBlock.classList.remove("is-empty");
        state.verdictIn = true;
        saveRound();

        // Objective 2 is the defence landing. Objective 3 is bound to the clock
        // at the moment the verdict arrives — not to elapsed time alone, and
        // never awarded once the clock has run out.
        completeObjective(1);
        var left = remainingSeconds();
        if (left !== null && left > 0) {
          completeObjective(2);
          setStatus("Verdict in, with " + formatLeft(left) + " left on the clock.", "live");
        } else {
          setStatus(
            "Verdict in, but the clock had already run out — the third objective " +
              "stays open for this round.",
            "error"
          );
        }
        stopClock();
        setRoundState("Round closed");
        if (el.verdictBlock && el.verdictBlock.scrollIntoView) {
          el.verdictBlock.scrollIntoView({ behavior: "smooth", block: "nearest" });
        }
      })
      .catch(function (err) {
        busy(el.defenceSubmit, false);
        setStatus(explain(err, "sending your defence"), "error");
      });
  }

  function formatLeft(secs) {
    var m = Math.floor(secs / 60);
    var s = secs % 60;
    if (m <= 0) return s + "s";
    return m + "m " + (s < 10 ? "0" : "") + s + "s";
  }

  // ── character counters ───────────────────────────────────────────────────

  function wireCounter(input, out) {
    if (!input || !out) return;
    var render = function () {
      var n = input.value.length;
      out.textContent = n + " / " + MAX_CHARS;
      out.classList.toggle("is-over", n >= MAX_CHARS);
    };
    input.addEventListener("input", render);
    render();
  }

  // ── init ─────────────────────────────────────────────────────────────────

  if (el.enter) {
    el.enter.addEventListener("click", function () {
      startRound().catch(function () {
        /* the status line already carries the reason */
      });
    });
  }
  if (el.positionSubmit) el.positionSubmit.addEventListener("click", submitPosition);
  if (el.defenceSubmit) el.defenceSubmit.addEventListener("click", submitDefence);

  // The defence refusal had nowhere visible to speak. Give it one, right by
  // the button, created here so this fix ships as JS only.
  if (el.defenceSubmit && el.defenceSubmit.parentNode) {
    el.defenceNote = document.createElement("p");
    el.defenceNote.className = "gi-defence-note";
    el.defenceNote.setAttribute("role", "status");
    el.defenceNote.setAttribute("aria-live", "polite");
    el.defenceNote.style.cssText =
      "margin:0.6rem 0 0;font-size:0.9rem;line-height:1.5;color:var(--section-text-muted,#9aa0b4);";
    el.defenceSubmit.parentNode.insertBefore(el.defenceNote, el.defenceSubmit.nextSibling);
  }
  wireCounter(el.positionInput, el.positionCount);
  wireCounter(el.defenceInput, el.defenceCount);
  if (el.positionInput) el.positionInput.addEventListener("input", saveRound);
  if (el.defenceInput) el.defenceInput.addEventListener("input", saveRound);
  window.addEventListener("pagehide", saveRound);
  updateProgress();
  setRoundState("No round yet");

  // Show today's provocation before a round starts, from the same feed the
  // /round endpoint serves, so a visitor can read what they'd be arguing
  // before committing to a clock. The round itself still comes from the API.
  // Reported live 2026-07-28: "I didn't see the initial provocation when I first
  // went on to the gauntlet page". Confirmed — the panel was EMPTY until this
  // fetch resolved (~3s), because the template ships it blank and nothing filled
  // the gap. Say what is happening instead of showing nothing.
  if (el.title && !el.title.textContent.trim()) {
    el.title.textContent = "Fetching today's provocation…";
  }

  // A resumed round redraws its own provocation, so only fetch when there is
  // nothing to restore — otherwise the feed's "today" could overwrite the
  // provocation the visitor's live round was actually started on.
  if (!restoreRound()) {
    fetch("/data/provocations.json", { cache: "no-cache" })
      .then(function (res) {
        if (!res.ok) throw new Error("provocations.json " + res.status);
        return res.json();
      })
      .then(function (data) {
        renderProvocation(data && data.today);
      })
      .catch(function () {
        // Honest, and never left sitting on "Fetching…" forever.
        if (el.title) {
          el.title.textContent = "Today's provocation could not be loaded.";
        }
        if (el.body) {
          el.body.textContent =
            "Pressing Enter the Gauntlet still works — the round draws its own provocation from the engine.";
        }
      });
  }
})();
