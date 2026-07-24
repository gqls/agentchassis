-- R8 (2026-07-24) — MatchMatrix v2: score all 10 grippers across six actuation
-- technologies, with per-technology physics. Owner decision 2026-07-24 ("extend
-- the tool", consistent with the R7 expand-don't-soften precedent): ~10 prose
-- fields claim the TOOL evaluates all six technologies; v1 tested only the 5
-- parallel-jaw grippers and its scope note said "5 grippers ... this is the
-- complete index" — false since R7 expanded the index to 10.
--
-- WHAT v2 DOES (each class assessed on the criteria its manufacturer publishes,
-- never a converted/inferred figure):
--   jaw (6 grippers incl. new Festo DHPS-10-A): friction-grip calc, unchanged
--     from v1 — F = m·a·S/(mu·n); the mu-trap conflict note retained.
--   magnetic (Schmalz SGM-HP 50): direct hold F = m·a·S vs the LOWER published
--     figure (385 N, friction ring); ferromagnetic-workpiece gate (only the
--     steel surface option passes).
--   vacuum (VG10) / adhesive (Gecko SP5) / soft (SG): published payload rating
--     vs dynamics-adjusted equivalent payload m' = m·a·S/g; soft also checks
--     the 11-118 mm cup range; per-class applicability notes (seal surface /
--     clean-smooth-dry / shape-dependent).
--   Scope note updated: 10 grippers, 6 manufacturers, how each class is tested.
--
-- TESTED: 30/30 logic tests via node:20-alpine (DOM stub over the real submit
-- handler) — includes the 2FG7 523.2 N regression case, magnetic-on-aluminium
-- fail, VG10 dynamics fail (15 kg rating < 16 kg equivalent), Gecko marginal
-- band, soft cup-range both bounds, DHPS travel fail (6 mm total stroke).
-- Test harness: scratchpad test_mm_v2.js (reproduced from RUNBOOK pattern).
--
-- DEPLOYED to gqls/sites robot-hands.com/tools/matchmatrix/index.html @ 67a6c3d5
-- (v1 template replaced verbatim inside the chrome; artefact 47,040 B). This
-- file keeps the DB html_template — the durable source (r4f) — in sync.
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < THIS_FILE

BEGIN;

UPDATE content_components
   SET html_template = $mm2$<style>
  .tool-container {
    max-width: 760px;
    margin: 0 auto;
    padding: 1.5rem;
    background: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: 8px;
    font-family: inherit;
    color: var(--color-text);
  }

  .tool-container h2 {
    margin: 0 0 0.5rem 0;
    font-size: 1.4rem;
    color: var(--color-primary);
  }

  .tool-description {
    margin: 0 0 1.5rem 0;
    font-size: 0.9rem;
    color: var(--color-text-muted);
    line-height: 1.5;
  }

  .tool-container fieldset {
    border: 1px solid var(--color-border);
    border-radius: 6px;
    padding: 1rem 1rem 0.5rem 1rem;
    margin: 0 0 1.25rem 0;
  }

  .tool-container legend {
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    padding: 0 0.4rem;
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  .form-group label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-text);
  }

  .form-group .hint {
    font-size: 0.75rem;
    color: var(--color-text-muted);
  }

  .input-unit-wrap {
    display: flex;
    align-items: stretch;
    border: 1px solid var(--color-border);
    border-radius: 5px;
    overflow: hidden;
    background: var(--color-background);
  }

  .input-unit-wrap input {
    flex: 1;
    min-width: 0;
    border: none;
    background: transparent;
    color: var(--color-text);
    padding: 0.55rem 0.7rem;
    font-size: 0.95rem;
    font-family: inherit;
  }

  .input-unit-wrap input:focus {
    outline: 2px solid var(--color-primary);
    outline-offset: -2px;
  }

  .unit-label {
    display: flex;
    align-items: center;
    padding: 0 0.7rem;
    font-size: 0.8rem;
    color: var(--color-text-muted);
    background: var(--color-surface);
    border-left: 1px solid var(--color-border);
  }

  .form-group select {
    border: 1px solid var(--color-border);
    border-radius: 5px;
    background: var(--color-background);
    color: var(--color-text);
    padding: 0.55rem 0.7rem;
    font-size: 0.95rem;
    font-family: inherit;
  }

  .form-group select:focus {
    outline: 2px solid var(--color-primary);
    outline-offset: -2px;
  }

  .error-msg {
    display: none;
    font-size: 0.75rem;
    color: #f87171;
  }

  .form-group.invalid .error-msg { display: block; }

  .tool-actions {
    display: flex;
    gap: 0.75rem;
    flex-wrap: wrap;
    margin-bottom: 1.25rem;
  }

  .btn-run, .btn-reset {
    font-family: inherit;
    font-size: 0.95rem;
    font-weight: 600;
    padding: 0.65rem 1.4rem;
    border-radius: 5px;
    cursor: pointer;
    border: 1px solid var(--color-border);
  }

  .btn-run {
    background: var(--color-primary);
    color: var(--color-background);
    border-color: var(--color-primary);
  }

  .btn-reset {
    background: transparent;
    color: var(--color-text-muted);
  }

  .btn-run:focus-visible, .btn-reset:focus-visible {
    outline: 2px solid var(--color-primary);
    outline-offset: 2px;
  }

  .requirement-panel {
    display: none;
    border: 1px solid var(--color-border);
    border-radius: 6px;
    padding: 1rem;
    margin-bottom: 1.25rem;
    background: var(--color-background);
  }

  .requirement-panel.visible { display: block; }

  .requirement-headline {
    font-size: 1.1rem;
    font-weight: 600;
    color: var(--color-text);
    margin-bottom: 0.5rem;
  }

  .requirement-headline strong { color: var(--color-primary); }

  .requirement-formula {
    font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    font-size: 0.8rem;
    color: var(--color-text-muted);
    line-height: 1.7;
  }

  .results-list { display: none; flex-direction: column; gap: 0.85rem; }
  .results-list.visible { display: flex; }

  .match-card {
    border: 1px solid var(--color-border);
    border-radius: 6px;
    padding: 0.9rem 1rem;
    background: var(--color-background);
  }

  .match-card.verdict-match { border-left: 3px solid #4ade80; }
  .match-card.verdict-marginal { border-left: 3px solid #fbbf24; }
  .match-card.verdict-unknown { border-left: 3px solid var(--color-text-muted); }
  .match-card.verdict-fail { border-left: 3px solid #f87171; }

  .match-head {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    gap: 0.75rem;
    flex-wrap: wrap;
    margin-bottom: 0.15rem;
  }

  .match-name { font-size: 1rem; font-weight: 600; color: var(--color-text); }
  .match-maker { font-size: 0.78rem; color: var(--color-text-muted); margin-bottom: 0.6rem; }

  .verdict-badge {
    font-size: 0.7rem;
    font-weight: 700;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    padding: 0.2rem 0.5rem;
    border-radius: 3px;
    white-space: nowrap;
  }

  .verdict-match .verdict-badge { background: #4ade80; color: #06240f; }
  .verdict-marginal .verdict-badge { background: #fbbf24; color: #2b1d00; }
  .verdict-unknown .verdict-badge { background: var(--color-text-muted); color: var(--color-background); }
  .verdict-fail .verdict-badge { background: #f87171; color: #2d0606; }

  .criteria-table { width: 100%; border-collapse: collapse; font-size: 0.82rem; }
  .criteria-table th {
    text-align: left;
    font-weight: 500;
    color: var(--color-text-muted);
    padding: 0.3rem 0.6rem 0.3rem 0;
    width: 8.5rem;
    vertical-align: top;
  }
  .criteria-table td { padding: 0.3rem 0; color: var(--color-text); vertical-align: top; }
  .criteria-table tr + tr th, .criteria-table tr + tr td { border-top: 1px solid var(--color-border); }

  .crit-flag { font-weight: 600; margin-left: 0.4rem; font-size: 0.75rem; }
  .crit-pass { color: #4ade80; }
  .crit-fail { color: #f87171; }
  .crit-none { color: var(--color-text-muted); font-style: italic; }

  .conflict-note {
    margin: 0.7rem 0 0 0;
    padding: 0.55rem 0.7rem;
    border-radius: 4px;
    background: var(--color-surface);
    border-left: 2px solid #fbbf24;
    font-size: 0.78rem;
    line-height: 1.55;
    color: var(--color-text-muted);
  }

  .tech-note {
    margin: 0.7rem 0 0 0;
    padding: 0.55rem 0.7rem;
    border-radius: 4px;
    background: var(--color-surface);
    border-left: 2px solid var(--color-text-muted);
    font-size: 0.78rem;
    line-height: 1.55;
    color: var(--color-text-muted);
  }

  .scope-note {
    margin-top: 1.5rem;
    padding-top: 1rem;
    border-top: 1px solid var(--color-border);
    font-size: 0.78rem;
    color: var(--color-text-muted);
    line-height: 1.6;
  }

  .scope-note strong { color: var(--color-text); }

  @media (max-width: 560px) {
    .form-row { grid-template-columns: 1fr; }
    .criteria-table th { width: 7rem; }
  }
</style>

<div class="tool-container">
  <h2>MatchMatrix &mdash; Gripper Selection Matrix</h2>
  <p class="tool-description">
    Enter your workpiece, duty cycle and environment. MatchMatrix works out what your application
    requires, then tests every gripper in the index against it &mdash; across all six actuation
    technologies, each assessed on the criteria its manufacturer actually publishes: a friction-grip
    force calculation for jaw grippers, a direct holding-force comparison for the magnetic gripper,
    and a dynamics-adjusted payload comparison where payload is the published rating. It shows
    exactly which criterion passes, which fails, and where a manufacturer publishes no figure at all.
  </p>

  <form id="mmForm" novalidate>

    <fieldset>
      <legend>Workpiece</legend>
      <div class="form-row">
        <div class="form-group" id="fg-mass">
          <label for="mmMass">Workpiece Mass</label>
          <div class="input-unit-wrap">
            <input type="number" id="mmMass" min="0.001" step="any" placeholder="e.g. 2.5" required>
            <span class="unit-label">kg</span>
          </div>
          <span class="error-msg">Enter a mass greater than zero.</span>
        </div>
        <div class="form-group" id="fg-travel">
          <label for="mmTravel">Required Jaw Travel / Part Size</label>
          <div class="input-unit-wrap">
            <input type="number" id="mmTravel" min="0" step="any" placeholder="e.g. 40" required>
            <span class="unit-label">mm</span>
          </div>
          <span class="hint">Total opening needed across the part &mdash; also checked against the soft gripper&rsquo;s cup range</span>
          <span class="error-msg">Enter a travel of zero or more.</span>
        </div>
      </div>
      <div class="form-row">
        <div class="form-group">
          <label for="mmSurface">Surface Material</label>
          <select id="mmSurface">
            <option value="0.10">Glass, smooth (&mu; 0.10)</option>
            <option value="0.15" selected>Steel, dry (&mu; 0.15)</option>
            <option value="0.20">Aluminium, machined (&mu; 0.20)</option>
            <option value="0.25">Plastic / ABS (&mu; 0.25)</option>
            <option value="0.30">Cardboard (&mu; 0.30)</option>
            <option value="0.50">Rubber (&mu; 0.50)</option>
          </select>
          <span class="hint">Sets friction &mu; for jaw grippers; also gates the magnetic gripper (ferromagnetic only)</span>
        </div>
        <div class="form-group">
          <label for="mmSurfaces">Gripping Surfaces</label>
          <select id="mmSurfaces">
            <option value="2" selected>2 &mdash; parallel two-jaw</option>
            <option value="3">3 &mdash; three-jaw centric</option>
          </select>
          <span class="hint">Number of jaws in friction contact (jaw grippers only)</span>
        </div>
      </div>
    </fieldset>

    <fieldset>
      <legend>Motion &amp; Duty</legend>
      <div class="form-row">
        <div class="form-group" id="fg-accel">
          <label for="mmAccel">Peak Acceleration</label>
          <div class="input-unit-wrap">
            <input type="number" id="mmAccel" min="0" step="any" value="9.81" required>
            <span class="unit-label">m/s&sup2;</span>
          </div>
          <span class="hint">Gravity (9.81) plus robot acceleration</span>
          <span class="error-msg">Enter a value of zero or more.</span>
        </div>
        <div class="form-group" id="fg-safety">
          <label for="mmSafety">Safety Factor</label>
          <div class="input-unit-wrap">
            <input type="number" id="mmSafety" min="1" step="any" value="2" required>
            <span class="unit-label">&times;</span>
          </div>
          <span class="hint">2 static &middot; 3&ndash;4 dynamic or overhead</span>
          <span class="error-msg">Enter a factor of 1 or more.</span>
        </div>
      </div>
    </fieldset>

    <fieldset>
      <legend>Environment</legend>
      <div class="form-row">
        <div class="form-group">
          <label for="mmIp">Minimum IP Rating</label>
          <select id="mmIp">
            <option value="0" selected>No requirement</option>
            <option value="30">IP30 &mdash; tool-protected</option>
            <option value="54">IP54 &mdash; dust / splash</option>
            <option value="64">IP64 &mdash; dust-tight / splash</option>
            <option value="67">IP67 &mdash; dust-tight / immersion</option>
          </select>
          <span class="hint">Grippers with no published rating are flagged, not excluded</span>
        </div>
        <div class="form-group">
          <label for="mmPayloadCheck">Payload Rating</label>
          <select id="mmPayloadCheck">
            <option value="1" selected>Check against published payload</option>
            <option value="0">Ignore &mdash; force calculation only</option>
          </select>
          <span class="hint">6 of 10 publish a payload figure. For vacuum, adhesive and soft grippers payload is the primary published rating and is always assessed</span>
        </div>
      </div>
    </fieldset>

    <div class="tool-actions">
      <button type="submit" class="btn-run">Run MatchMatrix</button>
      <button type="button" class="btn-reset" id="mmReset">Reset</button>
    </div>
  </form>

  <div class="requirement-panel" id="mmRequirement" role="status" aria-live="polite"></div>
  <div class="results-list" id="mmResults"></div>

  <div class="scope-note">
    <p><strong>What this index covers.</strong> MatchMatrix tests your requirement against the
    <strong>10 grippers currently held in the robot-hands.com index</strong>, spanning six actuation
    technologies &mdash; electric and pneumatic parallel-jaw, vacuum, magnetic, soft-robotic and
    adhesive. This is the complete index &mdash; it is not a survey of the gripper market, and a
    &ldquo;no match&rdquo; result means nothing in this index fits, not that no such gripper exists.</p>
    <p style="margin-top:0.6rem;"><strong>How each technology is assessed.</strong> Jaw grippers are
    tested with a friction-grip force calculation (published gripping force against the force your
    part needs at your &mu;). The magnetic gripper is tested directly against its published holding
    force &mdash; using the lower, friction-ring figure &mdash; and only on ferromagnetic material.
    Vacuum, adhesive and soft grippers publish a payload rating rather than a force figure, so they
    are tested against your part&rsquo;s dynamics-adjusted equivalent payload. The criteria differ
    because the published figures differ; no figure is ever converted into one the manufacturer did
    not publish.</p>
    <p style="margin-top:0.6rem;">Specifications are reproduced as published by Schunk, OnRobot,
    Robotiq, Zimmer Group, Festo and Schmalz. Where a field is blank, that manufacturer does not
    publish the figure &mdash; MatchMatrix marks it <em>not published</em> and never infers a value.
    Always confirm against the current manufacturer datasheet before purchase.</p>
  </div>
</div>

<script>
(function () {
  'use strict';

  // The surface-material select doubles as the material gate for the magnetic
  // gripper. Keys are the option VALUES (the friction coefficient as a string).
  var MATERIALS = {
    '0.10': { name: 'Glass, smooth',       ferrous: false },
    '0.15': { name: 'Steel, dry',          ferrous: true  },
    '0.20': { name: 'Aluminium, machined', ferrous: false },
    '0.25': { name: 'Plastic / ABS',       ferrous: false },
    '0.30': { name: 'Cardboard',           ferrous: false },
    '0.50': { name: 'Rubber',              ferrous: false }
  };

  // Specifications as published by each manufacturer. `text` is the published
  // string verbatim; the numeric field is that same figure normalised for
  // comparison. A null block means the manufacturer publishes no such figure —
  // it is never inferred. Per-jaw strokes are doubled to a total opening so all
  // travel figures compare like for like.
  //
  // tech: 'jaw' (friction grip), 'vacuum', 'magnetic', 'adhesive', 'soft'.
  // Jaw grippers are assessed by the friction-grip calculation; the magnetic
  // gripper directly against its published holding force (lower, friction-ring
  // figure) on ferromagnetic material only; vacuum/adhesive/soft against their
  // published payload rating with the dynamics factor applied.
  var GRIPPERS = [
    {
      name: 'Schunk EGP 40-N-S-B', maker: 'Schunk',
      tech: 'jaw', techLabel: 'Electric parallel-jaw',
      force: { n: 30, text: '30 N' },
      stroke: { mm: 12, text: '6 mm per jaw (12 mm total)' },
      payload: { kg: 0.15, text: '0.15 kg (recommended workpiece weight)' },
      ip: { v: 30, text: 'IP30' },
      extras: [['Weight', '0.3 kg'], ['Supply', '24 V DC'], ['Interface', 'Digital I/O']]
    },
    {
      name: 'OnRobot 2FG7', maker: 'OnRobot',
      tech: 'jaw', techLabel: 'Electric parallel-jaw',
      force: { n: 140, text: '20 N to 140 N' },
      stroke: { mm: 73, text: 'up to 73 mm external grip range' },
      payload: { kg: 11, text: '11 kg (24.3 lb)' },
      ip: { v: 67, text: 'IP67' },
      extras: []
    },
    {
      name: 'Robotiq 2F-85', maker: 'Robotiq',
      tech: 'jaw', techLabel: 'Electric parallel-jaw',
      force: { n: 235, text: '20 to 235 N' },
      stroke: { mm: 85, text: '85 mm' },
      payload: { kg: 5, text: '5 kg' },
      ip: null,
      extras: [['Weight', '925 g'], ['Supply', '24 V DC ±10%']]
    },
    {
      name: 'Zimmer Group GEP5010IO-00-A', maker: 'Zimmer Group',
      tech: 'jaw', techLabel: 'Electric parallel-jaw',
      force: { n: 1520, text: '1520 N' },
      stroke: { mm: 20, text: '10 mm per jaw (20 mm total)' },
      payload: null,
      ip: { v: 64, text: 'IP64' },
      extras: [['Weight', '1.6 kg'], ['Supply', '24 V'], ['Interface', 'I/O (IO-Link option)']]
    },
    {
      name: 'Festo EHPS-20-A-LK', maker: 'Festo',
      tech: 'jaw', techLabel: 'Electric parallel-jaw',
      force: { n: 218, text: '218 N' },
      stroke: { mm: 26, text: '13 mm per jaw (26 mm total)' },
      payload: null, ip: null,
      extras: []
    },
    {
      name: 'Festo DHPS-10-A', maker: 'Festo',
      tech: 'jaw', techLabel: 'Pneumatic parallel-jaw',
      force: { n: 34.5, text: '34.5 N per jaw closing (39 N opening) at 6 bar' },
      stroke: { mm: 6, text: '3 mm per jaw (6 mm total)' },
      payload: null, ip: null,
      extras: [['Weight', '67 g'], ['Operating pressure', '2 to 8 bar'],
               ['Repeat accuracy', '0.02 mm'], ['Supply', 'Compressed air']]
    },
    {
      name: 'OnRobot VG10', maker: 'OnRobot',
      tech: 'vacuum', techLabel: 'Electric vacuum',
      force: null, stroke: null, ip: null,
      payload: { kg: 15, text: '15 kg (35 lb)' },
      note: 'Suction hold — needs a surface a vacuum cup can seal against; porous or heavily perforated surfaces reduce holding force. Built-in pump, no external air supply.',
      extras: [['Zones', 'Dual, independently switchable']]
    },
    {
      name: 'OnRobot Gecko SP5', maker: 'OnRobot',
      tech: 'adhesive', techLabel: 'Adhesive (gecko)',
      force: null, stroke: null, ip: null,
      payload: { kg: 5, text: '5 kg' },
      note: 'Van der Waals adhesion — requires clean, smooth, dry, flat surfaces; not suitable for greasy, wet or dusty parts. No air and no electricity required.',
      extras: []
    },
    {
      name: 'OnRobot Soft Gripper SG', maker: 'OnRobot',
      tech: 'soft', techLabel: 'Soft silicone',
      force: null, ip: null,
      payload: { kg: 2.2, text: '2.2 kg (depends on shape, softness and friction of the part)' },
      grip: { min: 11, max: 118, text: '11 to 118 mm (cup-dependent)' },
      note: 'Payload depends on part geometry — the rating is an upper bound, not a guarantee. Food-grade silicone cups; no external air supply.',
      extras: [['Material', 'Food-grade silicone']]
    },
    {
      name: 'Schmalz SGM-HP 50', maker: 'Schmalz',
      tech: 'magnetic', techLabel: 'Permanent magnetic',
      force: { n: 385, text: '560 N (without friction ring), 385 N (with friction ring)' },
      stroke: null, payload: null, ip: null,
      note: 'Permanent-magnet surface hold on ferromagnetic material only. Assessed against the lower published figure (385 N, with friction ring). Workpiece temperatures up to 350 °C.',
      extras: [['Diameter', '50 mm'], ['Max workpiece temp', '350 °C']]
    }
  ];

  var form = document.getElementById('mmForm');
  var reqPanel = document.getElementById('mmRequirement');
  var results = document.getElementById('mmResults');
  if (!form) { return; }

  function esc(s) {
    return String(s).replace(/[&<>"']/g, function (c) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
    });
  }

  function setInvalid(id, bad) {
    var g = document.getElementById(id);
    if (g) { g.classList.toggle('invalid', !!bad); }
  }

  function readInputs() {
    var mass = parseFloat(document.getElementById('mmMass').value);
    var travel = parseFloat(document.getElementById('mmTravel').value);
    var accel = parseFloat(document.getElementById('mmAccel').value);
    var safety = parseFloat(document.getElementById('mmSafety').value);
    var muKey = document.getElementById('mmSurface').value;

    var bad = false;
    if (!isFinite(mass) || mass <= 0) { setInvalid('fg-mass', true); bad = true; } else { setInvalid('fg-mass', false); }
    if (!isFinite(travel) || travel < 0) { setInvalid('fg-travel', true); bad = true; } else { setInvalid('fg-travel', false); }
    if (!isFinite(accel) || accel < 0) { setInvalid('fg-accel', true); bad = true; } else { setInvalid('fg-accel', false); }
    if (!isFinite(safety) || safety < 1) { setInvalid('fg-safety', true); bad = true; } else { setInvalid('fg-safety', false); }
    if (bad) { return null; }

    return {
      mass: mass,
      travel: travel,
      accel: accel,
      safety: safety,
      mu: parseFloat(muKey),
      material: MATERIALS[muKey] || { name: 'Unknown', ferrous: false },
      n: parseFloat(document.getElementById('mmSurfaces').value),
      ip: parseInt(document.getElementById('mmIp').value, 10),
      checkPayload: document.getElementById('mmPayloadCheck').value === '1'
    };
  }

  // One application, three requirement figures — because the index's
  // manufacturers publish three different kinds of capacity figure:
  //   fJaw — friction-grip force per surface:  F = (m * a * S) / (mu * n)
  //   fDir — direct normal holding force:      F =  m * a * S
  //   mEq  — dynamics-adjusted payload:        m' = (m * a * S) / g
  function requirements(v) {
    var dyn = v.mass * v.accel * v.safety;
    return {
      fJaw: dyn / (v.mu * v.n),
      fDir: dyn,
      mEq: dyn / 9.81
    };
  }

  function ipRow(g, v, state) {
    if (v.ip > 0) {
      if (g.ip) {
        var okI = g.ip.v >= v.ip;
        if (!okI) { state.failed = true; }
        return { label: 'IP rating', text: g.ip.text, flag: okI ? 'pass' : 'fail',
                 note: okI ? '' : 'needs IP' + v.ip };
      }
      state.unknown = true;
      return { label: 'IP rating', text: null, flag: 'none', note: '' };
    }
    if (g.ip) { return { label: 'IP rating', text: g.ip.text, flag: '', note: 'no requirement set' }; }
    return null;
  }

  // Payload assessed against the dynamics-adjusted equivalent mass, so a
  // rating earned at rest is not silently credited with surviving your
  // acceleration and safety factor.
  function payloadCapacityRow(g, req, state) {
    var okP = g.payload.kg >= req.mEq;
    if (!okP) { state.failed = true; }
    var margin = req.mEq > 0 ? ((g.payload.kg - req.mEq) / req.mEq) * 100 : 0;
    return {
      label: 'Rated payload', text: g.payload.text,
      flag: okP ? 'pass' : 'fail',
      note: okP ? '+' + margin.toFixed(0) + '% vs ' + req.mEq.toFixed(2) + ' kg equivalent'
                : 'needs ' + req.mEq.toFixed(2) + ' kg equivalent (mass × a × S ÷ g)'
    };
  }

  function assessJaw(g, v, req, state) {
    var rows = [];

    if (g.force) {
      var ok = g.force.n >= req.fJaw;
      if (!ok) { state.failed = true; }
      var margin = req.fJaw > 0 ? ((g.force.n - req.fJaw) / req.fJaw) * 100 : 0;
      rows.push({
        label: 'Gripping force', text: g.force.text,
        flag: ok ? 'pass' : 'fail',
        note: ok ? (margin >= 0 ? '+' + margin.toFixed(0) + '% margin' : '') : 'needs ' + req.fJaw.toFixed(1) + ' N'
      });
    } else {
      state.unknown = true;
      rows.push({ label: 'Gripping force', text: null, flag: 'none', note: '' });
    }

    if (g.stroke) {
      var okS = g.stroke.mm >= v.travel;
      if (!okS) { state.failed = true; }
      rows.push({
        label: 'Jaw travel', text: g.stroke.text,
        flag: okS ? 'pass' : 'fail',
        note: okS ? '' : 'needs ' + v.travel + ' mm'
      });
    } else {
      state.unknown = true;
      rows.push({ label: 'Jaw travel', text: null, flag: 'none', note: '' });
    }

    if (v.checkPayload) {
      if (g.payload) {
        var okP = g.payload.kg >= v.mass;
        if (!okP) { state.failed = true; }
        rows.push({
          label: 'Rated payload', text: g.payload.text,
          flag: okP ? 'pass' : 'fail',
          note: okP ? '' : 'needs ' + v.mass + ' kg'
        });
      } else {
        state.unknown = true;
        rows.push({ label: 'Rated payload', text: null, flag: 'none', note: '' });
      }
    } else if (g.payload) {
      rows.push({ label: 'Rated payload', text: g.payload.text, flag: '', note: 'not checked' });
    }

    var ip = ipRow(g, v, state);
    if (ip) { rows.push(ip); }

    state.capacity = g.force ? g.force.n : 0;
    state.need = req.fJaw;
    return rows;
  }

  function assessMagnetic(g, v, req, state) {
    var rows = [];

    var okM = v.material.ferrous;
    if (!okM) { state.failed = true; }
    rows.push({
      label: 'Workpiece material', text: v.material.name,
      flag: okM ? 'pass' : 'fail',
      note: okM ? 'ferromagnetic' : 'requires a ferromagnetic workpiece'
    });

    var ok = g.force.n >= req.fDir;
    if (!ok) { state.failed = true; }
    var margin = req.fDir > 0 ? ((g.force.n - req.fDir) / req.fDir) * 100 : 0;
    rows.push({
      label: 'Holding force', text: g.force.text,
      flag: ok ? 'pass' : 'fail',
      note: ok ? '+' + margin.toFixed(0) + '% vs ' + req.fDir.toFixed(1) + ' N direct hold'
               : 'needs ' + req.fDir.toFixed(1) + ' N direct hold'
    });

    rows.push({ label: 'Jaw travel', text: 'Not applicable — surface hold, no jaws', flag: '', note: '' });

    var ip = ipRow(g, v, state);
    if (ip) { rows.push(ip); }

    state.capacity = g.force.n;
    state.need = req.fDir;
    return rows;
  }

  function assessPayloadRated(g, v, req, state) {
    var rows = [];

    rows.push(payloadCapacityRow(g, req, state));

    if (g.tech === 'soft' && g.grip) {
      if (v.travel > 0) {
        var okG = v.travel >= g.grip.min && v.travel <= g.grip.max;
        if (!okG) { state.failed = true; }
        rows.push({
          label: 'Grip range', text: g.grip.text,
          flag: okG ? 'pass' : 'fail',
          note: okG ? '' : (v.travel < g.grip.min
            ? 'part smaller than the smallest cup (' + g.grip.min + ' mm)'
            : 'part exceeds the largest cup (' + g.grip.max + ' mm)')
        });
      } else {
        rows.push({ label: 'Grip range', text: g.grip.text, flag: '', note: 'no part size set' });
      }
    } else {
      rows.push({ label: 'Jaw travel', text: 'Not applicable — surface hold, no jaws', flag: '', note: '' });
    }

    var ip = ipRow(g, v, state);
    if (ip) { rows.push(ip); }

    state.capacity = g.payload.kg;
    state.need = req.mEq;
    return rows;
  }

  function assess(g, v, req) {
    var state = { failed: false, unknown: false, capacity: 0, need: 1 };
    var rows;

    if (g.tech === 'magnetic') {
      rows = assessMagnetic(g, v, req, state);
    } else if (g.tech === 'vacuum' || g.tech === 'adhesive' || g.tech === 'soft') {
      rows = assessPayloadRated(g, v, req, state);
    } else {
      rows = assessJaw(g, v, req, state);
    }

    // A jaw gripper can pass on its published payload rating and still fail
    // the force calculation, because headline payload figures assume the
    // manufacturer's own friction assumptions. Say so rather than leaving the
    // user to reconcile two rows that appear to contradict each other.
    var conflict = null;
    if (g.tech === 'jaw' && g.force && g.payload && g.force.n < req.fJaw && g.payload.kg >= v.mass) {
      var holds = (g.force.n * v.mu * v.n) / (v.accel * v.safety);
      var impliedMu = (g.payload.kg * v.accel * v.safety) / (g.force.n * v.n);
      conflict = 'Rated for ' + g.payload.text + ', but ' + g.force.n +
        ' N holds only ' + holds.toFixed(1) + ' kg on your surface (μ ' + v.mu +
        '). That payload rating implies μ ≈ ' + impliedMu.toFixed(2) +
        ' — high-friction or form-fit fingers, not a bare machined surface.';
    }

    var verdict, rank;
    if (state.failed) {
      verdict = 'No match'; rank = 3;
    } else if (state.unknown) {
      verdict = 'Insufficient data'; rank = 2;
    } else if (state.capacity < state.need * 1.25) {
      verdict = 'Marginal'; rank = 1;
    } else {
      verdict = 'Match'; rank = 0;
    }

    var cls = ['verdict-match', 'verdict-marginal', 'verdict-unknown', 'verdict-fail'][rank];
    var headroom = state.need > 0 ? state.capacity / state.need : 0;
    return { g: g, rows: rows, verdict: verdict, rank: rank, cls: cls,
             headroom: headroom, conflict: conflict };
  }

  function renderRow(r) {
    var value = r.text === null
      ? '<span class="crit-none">Not published by manufacturer</span>'
      : esc(r.text);
    var flag = '';
    if (r.flag === 'pass') { flag = '<span class="crit-flag crit-pass">PASS</span>'; }
    else if (r.flag === 'fail') { flag = '<span class="crit-flag crit-fail">FAIL</span>'; }
    var note = r.note ? ' <span class="crit-none">' + esc(r.note) + '</span>' : '';
    return '<tr><th scope="row">' + esc(r.label) + '</th><td>' + value + flag + note + '</td></tr>';
  }

  function renderCard(a) {
    var extras = a.g.extras && a.g.extras.length
      ? a.g.extras.map(function (e) {
          return '<tr><th scope="row">' + esc(e[0]) + '</th><td>' + esc(e[1]) + '</td></tr>';
        }).join('')
      : '';
    var conflict = a.conflict
      ? '<p class="conflict-note">' + esc(a.conflict) + '</p>'
      : '';
    var techNote = a.g.note
      ? '<p class="tech-note">' + esc(a.g.note) + '</p>'
      : '';
    return '<div class="match-card ' + a.cls + '">' +
      '<div class="match-head">' +
        '<span class="match-name">' + esc(a.g.name) + '</span>' +
        '<span class="verdict-badge">' + esc(a.verdict) + '</span>' +
      '</div>' +
      '<div class="match-maker">' + esc(a.g.maker) + ' &middot; ' + esc(a.g.techLabel) + '</div>' +
      '<table class="criteria-table"><tbody>' +
        a.rows.map(renderRow).join('') + extras +
      '</tbody></table>' + conflict + techNote +
    '</div>';
  }

  form.addEventListener('submit', function (ev) {
    ev.preventDefault();
    var v = readInputs();
    if (!v) {
      reqPanel.classList.remove('visible');
      results.classList.remove('visible');
      return;
    }

    var req = requirements(v);

    reqPanel.innerHTML =
      '<div class="requirement-headline">Application requirement, by technology class</div>' +
      '<div class="requirement-formula">' +
        'Friction grip (jaw): F = (m &times; a &times; S) &divide; (&mu; &times; n) = (' +
          v.mass + ' &times; ' + v.accel + ' &times; ' + v.safety + ') &divide; (' + v.mu +
          ' &times; ' + v.n + ') = <strong>' + req.fJaw.toFixed(1) + ' N</strong><br>' +
        'Direct hold (magnetic): F = m &times; a &times; S = <strong>' +
          req.fDir.toFixed(1) + ' N</strong><br>' +
        'Equivalent payload (vacuum / adhesive / soft): m&prime; = (m &times; a &times; S) &divide; g = <strong>' +
          req.mEq.toFixed(2) + ' kg</strong>' +
      '</div>';
    reqPanel.classList.add('visible');

    var assessed = GRIPPERS.map(function (g) { return assess(g, v, req); });
    assessed.sort(function (a, b) {
      if (a.rank !== b.rank) { return a.rank - b.rank; }
      return b.headroom - a.headroom;
    });

    var matches = assessed.filter(function (a) { return a.rank === 0 || a.rank === 1; }).length;
    var summary = matches === 0
      ? '<div class="match-maker" style="margin-bottom:0.25rem;">No gripper in this index meets the requirement. The closest are shown first.</div>'
      : '<div class="match-maker" style="margin-bottom:0.25rem;">' + matches + ' of ' + GRIPPERS.length +
        ' indexed grippers meet the requirement.</div>';

    results.innerHTML = summary + assessed.map(renderCard).join('');
    results.classList.add('visible');
  });

  document.getElementById('mmReset').addEventListener('click', function () {
    form.reset();
    ['fg-mass', 'fg-travel', 'fg-accel', 'fg-safety'].forEach(function (id) { setInvalid(id, false); });
    reqPanel.classList.remove('visible');
    results.classList.remove('visible');
    results.innerHTML = '';
  });
})();
</script>$mm2$,
       updated_at = now()
 WHERE name = 'tool-matchmatrix-robot-hands-com';

\echo '--- verify: template length + v2 markers (expect 1 row, 10-gripper scope note, DHPS present) ---'
SELECT length(html_template) AS tpl_len,
       (html_template LIKE '%10 grippers currently held%') AS has_v2_scope,
       (html_template LIKE '%Festo DHPS-10-A%') AS has_dhps,
       (html_template LIKE '%SGM-HP 50%') AS has_magnetic
FROM content_components WHERE name = 'tool-matchmatrix-robot-hands-com';

COMMIT;
