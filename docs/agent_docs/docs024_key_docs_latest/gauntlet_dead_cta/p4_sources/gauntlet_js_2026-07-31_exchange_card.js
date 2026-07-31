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
    entry: section.querySelector("[data-gi-entry]"),
    entryStatus: section.querySelector("[data-gi-entry-status]"),
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
    verdictReasons: section.querySelector("[data-gi-verdict-reasons]"),
    shareCard: section.querySelector("[data-gi-share-card]"),
    ledger: section.querySelector("[data-gi-ledger]"),
    ledgerList: section.querySelector("[data-gi-ledger-list]"),
    ledgerCount: section.querySelector("[data-gi-ledger-count]"),
    ledgerClear: section.querySelector("[data-gi-ledger-clear]")
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
        positionText: state.positionText || "",
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
      setEntryStatus(
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
    state.positionText = d.positionText || "";
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

    reveal();
    applyStepEmphasis();
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

  // ── sealed entry (bugs_open/131 item C, owner ruling 2026-07-28) ────────
  //
  // The page opens SEALED: the provocation panel is hidden and the entry
  // block is the only visible door. reveal() is called from exactly two
  // places — /round returning 200, and resuming a stored round that is
  // genuinely live or complete on the server. A click alone never reveals:
  // the reveal is itself an API-bound state change, like every other state
  // change in this file.
  function sealed() {
    return section.classList.contains("gi-sealed");
  }
  function reveal() {
    section.classList.remove("gi-sealed");
    if (el.entryStatus) el.entryStatus.textContent = "";
  }
  // While sealed, the main status line lives inside the hidden panel, so
  // anything said there would be said to nobody. Speak at the entry instead.
  function setEntryStatus(text, kind) {
    if (!el.entryStatus) { setStatus(text, kind); return; }
    stopWaitCounter();
    el.entryStatus.textContent = text;
    el.entryStatus.className = "gi-status gi-entry-status" + (kind ? " is-" + kind : "");
  }

  // A 15-second wait with no feedback reads as a hang. Count it out loud.
  function startWaitCounter(text, target) {
    stopWaitCounter();
    var box = target || el.status;
    if (!box) return;
    var started = Date.now();
    box.className = box === el.entryStatus
      ? "gi-status gi-entry-status is-working"
      : "gi-status is-working";
    var render = function () {
      var secs = Math.round((Date.now() - started) / 1000);
      box.textContent = text + " (" + secs + "s)";
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

  // F (bugs_open/131): the steps get a visual RANKING that follows the
  // round state — current, done, future. Called only from the handlers
  // that advance the round on real API responses (and from restore/init,
  // which re-derive the same state). It ranks attention and gates
  // nothing: every control stays enabled.
  function applyStepEmphasis() {
    var steps = section.querySelectorAll(".gi-steps .gi-step");
    // 0 position · 1 opponent reply · 2 defence · 3 verdict
    var stage = state.verdictIn ? 3 : state.positionFiled ? 2 : 0;
    for (var i = 0; i < steps.length; i++) {
      steps[i].classList.remove("is-current", "is-done", "is-future");
      if (i < stage || (state.verdictIn && i !== 3)) {
        steps[i].classList.add("is-done");
      } else if (i === stage) {
        steps[i].classList.add("is-current");
      } else {
        steps[i].classList.add("is-future");
      }
    }
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
        return "This round has expired. File your position again — it starts a fresh " +
          "round, and what you typed is kept.";
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
    startWaitCounter("Drawing today's provocation and starting your clock…",
      sealed() ? el.entryStatus : el.status);
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
        reveal();
        applyStepEmphasis();
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
        // No round: no clock, no objective, no reveal, no pretending.
        if (sealed()) {
          setEntryStatus(explain(err, "starting the round"), "error");
        } else {
          setStatus(explain(err, "starting the round"), "error");
        }
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
        state.positionText = text;
        completeObjective(0);
        applyStepEmphasis();
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
        if (err && err.status === 404) { state.roundId = null; saveRound(); }
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
        : "There is no live round — file your position again to start a fresh one. " +
          "What you have typed is kept.";
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
        recordLedgerEntry(data.verdict || "");

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
        applyStepEmphasis();
        stopClock();
        setRoundState("Round closed");
        if (el.verdictBlock && el.verdictBlock.scrollIntoView) {
          el.verdictBlock.scrollIntoView({ behavior: "smooth", block: "nearest" });
        }
      })
      .catch(function (err) {
        busy(el.defenceSubmit, false);
        if (err && err.status === 404) { state.roundId = null; saveRound(); }
        setStatus(explain(err, "sending your defence"), "error");
      });
  }

  function formatLeft(secs) {
    var m = Math.floor(secs / 60);
    var s = secs % 60;
    if (m <= 0) return s + "s";
    return m + "m " + (s < 10 ? "0" : "") + s + "s";
  }

  // ── the share card: the EXCHANGE, not just the verdict ─────────────────
  //
  // Every string drawn on the card is a fact of THIS round: the provocation it
  // was started on, the challenge the engine actually put, the visitor's own
  // defence, the judge's actual ruling, the date and the page address. No
  // win-rate, no leaderboard, no count of anything — the fabrication classes
  // deleted from this site stay deleted. The button sits inside the verdict
  // step, hidden until /defend returns, so it cannot fire before a verdict.
  //
  // WHY IT CHANGED (owner ruling 2026-07-31, option "3 staged via 1"):
  // the card used to carry the provocation headline and the verdict word, and
  // the verdict word is 13 characters ("opponent wins") — so what travelled was
  // "a stranger scored badly on an argument you cannot read". Measured over the
  // 51 complete rounds stored at that date, a full round averages 3,109
  // characters, and one 1200x630 card holds ~700 legibly once a timeline has
  // downscaled it — so the whole debate provably cannot fit (it auto-fits at
  // 11px, ~4.6px in a feed). The exchange can: challenge + defence was 599
  // characters on the measured round and fits at 26px.
  //
  // Deliberately NOT on the card: the engine's counter-argument and the judge's
  // reasons — 2,285 characters between them on that round. They are the case
  // for the per-round permalink (step 2), not something to shrink onto a card.
  //
  // The card carries no per-round URL BY DESIGN: there is no per-round page
  // yet, and a link that 404s is worse than no link.

  // Measure-only wrapper: returns the lines, callers draw them. The previous
  // version wrapped and drew in one pass, which cannot be used to size a block
  // before committing to a type size.
  function wrapLines(x, text, maxWidth) {
    var words = String(text).split(/\s+/);
    var line = "", out = [];
    for (var i = 0; i < words.length; i++) {
      var probe = line ? line + " " + words[i] : words[i];
      if (x.measureText(probe).width > maxWidth && line) {
        out.push(line);
        line = words[i];
      } else {
        line = probe;
      }
    }
    if (line) out.push(line);
    return out;
  }

  function buildVerdictCard() {
    var prov = state.provocation && state.provocation.headline
      ? String(state.provocation.headline).replace(/<[^>]*>/g, "").trim() : "";
    var verdict = el.verdict ? el.verdict.textContent.trim() : "";
    var challenge = el.opponentChallenge ? el.opponentChallenge.textContent.trim() : "";
    var defence = el.defenceInput ? el.defenceInput.value.trim() : "";

    // All four are facts of the round, so all four are required. A card with an
    // empty block would assert that one half of the exchange was blank. Both
    // the fresh path and the sessionStorage resume path populate every one of
    // these (restoreRound writes draftDefence back into the input and the
    // challenge back into its element), so this refuses only when a round
    // genuinely lacks a piece.
    if (!verdict || !challenge || !defence) return null;

    var W = 1200, H = 630, L = 70, MAXW = W - 140;
    var c = document.createElement("canvas");
    c.width = W; c.height = H;
    var x = c.getContext("2d");
    if (!x) return null;

    var BLOCKS = [
      ["VONC ASKED", challenge, "system-ui, sans-serif"],
      ["I ANSWERED", defence, "Georgia, serif"]
    ];

    // Fit the type to the round rather than truncating the round to the type: a
    // real challenge has run to 305 characters and a defence to 294, and both
    // vary per round, so a fixed size either clips or wastes the card. TOP and
    // FOOT reserve the header, the ruling line and the address. Measured
    // against the DRAWN layout, not a raw character budget — labels and chrome
    // take vertical space the prose then cannot have, and a 32px draft that was
    // "inside" a 737-character budget still overlapped its own ruling line.
    var TOP = 112, FOOT = 130, USABLE = H - TOP - FOOT;
    function heightAt(f) {
      var lh = Math.round(f * 1.3), total = 0;
      for (var i = 0; i < BLOCKS.length; i++) {
        x.font = "400 " + f + "px " + BLOCKS[i][2];
        total += 34 + wrapLines(x, BLOCKS[i][1], MAXW).length * lh + 26;
      }
      return total;
    }
    var size = 34;
    while (size > 12 && heightAt(size) > USABLE) size--;

    x.fillStyle = "#6d28d9";
    x.fillRect(0, 0, W, H);
    x.fillStyle = "#f59e0b";
    x.fillRect(0, 0, 14, H);

    x.fillStyle = "rgba(255,255,255,0.8)";
    x.font = "700 22px system-ui, sans-serif";
    x.fillText(("THE GAUNTLET \u00B7 " + prov.replace(/\.$/, "")).toUpperCase(), L, 58);

    var lh = Math.round(size * 1.3), y = TOP;
    for (var b = 0; b < BLOCKS.length; b++) {
      x.fillStyle = "#fbbf24";
      x.font = "700 20px system-ui, sans-serif";
      x.fillText(BLOCKS[b][0], L, y);
      y += 34;
      x.fillStyle = "#ffffff";
      x.font = "400 " + size + "px " + BLOCKS[b][2];
      var lines = wrapLines(x, BLOCKS[b][1], MAXW);
      for (var j = 0; j < lines.length; j++) {
        x.fillText(lines[j], L, y);
        y += lh;
      }
      y += 26;
    }

    x.fillStyle = "#f59e0b";
    x.fillRect(L, H - 112, 120, 6);
    x.fillStyle = "#ffffff";
    x.font = "800 34px system-ui, sans-serif";
    x.fillText("The judge ruled: " + verdict + ".", L, H - 62);
    x.fillStyle = "rgba(255,255,255,0.7)";
    x.font = "600 22px system-ui, sans-serif";
    x.fillText("vonc.com/tools/gauntlet \u00B7 " + new Date().toLocaleDateString("en-GB"), L, H - 24);
    return c;
  }

  function shareVerdictCard() {
    var c = buildVerdictCard();
    if (!c || !c.toBlob) return;
    c.toBlob(function (blob) {
      if (!blob) return;
      var file = null;
      try { file = new File([blob], "gauntlet-verdict.png", { type: "image/png" }); } catch (e) {}
      if (file && navigator.canShare && navigator.canShare({ files: [file] })) {
        navigator.share({ files: [file], title: "The Gauntlet \u2014 my verdict" })
          .catch(function () { /* visitor dismissed the sheet; nothing owed */ });
        return;
      }
      var a = document.createElement("a");
      a.href = URL.createObjectURL(blob);
      a.download = "gauntlet-verdict.png";
      document.body.appendChild(a);
      a.click();
      a.remove();
      setTimeout(function () { URL.revokeObjectURL(a.href); }, 4000);
    }, "image/png");
  }

  // ── opinion ledger (owner direction 2026-07-29) ──────────────────────────
  //
  // "A (dated) personal history of your opinions might be a goldmine" — a
  // device-local diary of the visitor's own completed rounds. The round store
  // above is sessionStorage because a round should die with its tab; the
  // ledger is localStorage because a history's whole point is to survive it.
  // Entries are created in exactly one place — the /defend success handler —
  // so every line is a fact of a real judged round: the provocation the round
  // served, the position actually FILED (captured when /position succeeded,
  // not read back from the editable input), the judge's verdict, the date.
  // Nothing is synthesised, backfilled, or written on restore. No accounts,
  // no server copy: the record never leaves the browser, and the visitor can
  // erase it.
  var LEDGER_STORE = "vonc_gauntlet_ledger_v1";
  var LEDGER_MAX = 100;

  function readLedger() {
    try {
      var raw = localStorage.getItem(LEDGER_STORE);
      var list = raw ? JSON.parse(raw) : [];
      return Array.isArray(list) ? list : [];
    } catch (e) { return []; }
  }

  function recordLedgerEntry(verdictText) {
    if (!state.roundId || !verdictText) return;
    var entry = {
      roundId: state.roundId,
      date: new Date().toISOString(),
      provocation: state.provocation && state.provocation.headline
        ? String(state.provocation.headline).replace(/<[^>]*>/g, "") : "",
      position: state.positionText || "",
      verdict: verdictText
    };
    try {
      // One entry per round: a second verdict for the same round replaces
      // the first rather than manufacturing a second row in the diary.
      var list = readLedger().filter(function (e) {
        return e && e.roundId !== entry.roundId;
      });
      list.push(entry);
      if (list.length > LEDGER_MAX) list = list.slice(list.length - LEDGER_MAX);
      localStorage.setItem(LEDGER_STORE, JSON.stringify(list));
    } catch (e) { /* private mode: the round still happened; only the diary is unavailable */ }
    renderLedger();
  }

  function ledgerDate(iso) {
    var d = new Date(iso);
    if (isNaN(d.getTime())) return "";
    return d.toLocaleDateString("en-GB", { day: "numeric", month: "long", year: "numeric" });
  }

  function ledgerLine(tag, text, cls) {
    var p = document.createElement("p");
    p.className = cls;
    var t = document.createElement("span");
    t.className = "gi-ledger-tag";
    t.textContent = tag;
    p.appendChild(t);
    p.appendChild(document.createTextNode(text));
    return p;
  }

  function renderLedger() {
    if (!el.ledger || !el.ledgerList) return;
    var list = readLedger();
    el.ledgerList.textContent = "";
    if (!list.length) {
      el.ledger.classList.add("is-empty");
      return;
    }
    for (var i = list.length - 1; i >= 0; i--) {
      var e = list[i];
      if (!e || !e.verdict) continue;
      var li = document.createElement("li");
      li.className = "gi-ledger-entry";
      var date = document.createElement("div");
      date.className = "gi-ledger-date";
      date.textContent = ledgerDate(e.date);
      li.appendChild(date);
      if (e.provocation) {
        var q = document.createElement("blockquote");
        q.className = "gi-ledger-provocation";
        q.textContent = "\u201C" + e.provocation + "\u201D";
        li.appendChild(q);
      }
      if (e.position) li.appendChild(ledgerLine("You argued", e.position, "gi-ledger-position"));
      li.appendChild(ledgerLine("Verdict", e.verdict, "gi-ledger-verdict"));
      el.ledgerList.appendChild(li);
    }
    if (el.ledgerCount) {
      el.ledgerCount.textContent = list.length + (list.length === 1 ? " round on record" : " rounds on record");
    }
    el.ledger.classList.remove("is-empty");
  }

  // Erasing is two presses because there is no undo: the first arms and says
  // so on the button itself; the second, within 6 seconds, erases.
  var ledgerClearArm = null;

  function clearLedger() {
    if (!ledgerClearArm) {
      if (el.ledgerClear) el.ledgerClear.textContent = "Press again to erase for good";
      ledgerClearArm = setTimeout(function () {
        ledgerClearArm = null;
        if (el.ledgerClear) el.ledgerClear.textContent = "Erase this record";
      }, 6000);
      return;
    }
    clearTimeout(ledgerClearArm);
    ledgerClearArm = null;
    try { localStorage.removeItem(LEDGER_STORE); } catch (e) {}
    if (el.ledgerClear) el.ledgerClear.textContent = "Erase this record";
    renderLedger();
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
  if (el.shareCard) el.shareCard.addEventListener("click", shareVerdictCard);
  if (el.ledgerClear) el.ledgerClear.addEventListener("click", clearLedger);

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
  applyStepEmphasis();
  setRoundState("No round yet");

  // No pre-round provocation fetch (bugs_open/131 item C, owner ruling
  // 2026-07-28). The page opens SEALED — the provocation is deliberately not
  // shown until the button press starts a real round, which returns its own
  // provocation. The previous behaviour (pre-rendering today's provocation
  // from /data/provocations.json "so a visitor can read what they'd be
  // arguing before committing") was REVERSED by that ruling: the primary
  // button now reveals, so pre-showing the question would return it to a
  // button whose only visible effect is a clock starting. Do not "fix" this
  // by restoring the fetch. A resumed round still redraws its own
  // provocation inside restoreRound().
  restoreRound();
  renderLedger();
})();



