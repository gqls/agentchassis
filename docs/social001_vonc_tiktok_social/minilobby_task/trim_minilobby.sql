-- trim_minilobby.sql — provocation-card mini-lobby trim (vonc.com / Spark)
-- Generated 2026-07-09. Atomic + self-verifying: the DO block raises on any
-- mismatch, so psql -v ON_ERROR_STOP=1 never reaches COMMIT and the whole
-- transaction rolls back.
--
--   1. Dated backups of the three artefacts.
--   2. component_versions snapshot of the template (a direct UPDATE bypasses
--      store_generated_component's snapshotting; repair_template_slots snapshots
--      by hand for the same reason).
--   3. UPDATE content_components.html_template   10300 -> 6618 chars
--   4. UPDATE js_snippets.js_content              4879 -> 3365 chars
--      (js_snippets has NO updated_at column — do not add one to the SET list.)
--
-- It does NOT touch page_components.rendered_html. That is regenerated from the
-- template by the section-editor (086_section_edit_provocation-card_vonc.sh).
--
-- Verified premise: for this component rendered_html is EXACTLY the template with
-- every literal '<no value>' removed (10300 - 26*10 = 10040 = current rendered_len).
-- After the trim: 6618 - 13*10 = 6488 expected rendered_len.

\set ON_ERROR_STOP on
BEGIN;

-- ── 1. Dated backups (never reuse a name: CREATE TABLE IF NOT EXISTS no-ops silently) ──
CREATE TABLE _vonc_pc_backup_20260709 AS
  SELECT * FROM page_components WHERE page_id = 'b4d24f8e-fccd-49df-9dad-aa56a0b20a68'::uuid;
CREATE TABLE _vonc_cc_pcard_backup_20260709 AS
  SELECT * FROM content_components WHERE id = '6163ff14-9f94-4962-aa19-d2718eabdeb1'::uuid;
CREATE TABLE _vonc_snippet_backup_20260709 AS
  SELECT * FROM js_snippets WHERE name = 'provocation-card-loader';

-- ── 2. component_versions snapshot of the CURRENT template ──
INSERT INTO component_versions (component_id, version_number, html_template, input_schema, change_source)
SELECT '6163ff14-9f94-4962-aa19-d2718eabdeb1'::uuid,
       COALESCE((SELECT MAX(version_number) FROM component_versions WHERE component_id = '6163ff14-9f94-4962-aa19-d2718eabdeb1'::uuid), 0) + 1,
       html_template, input_schema, 'minilobby_trim_20260709'
FROM content_components WHERE id = '6163ff14-9f94-4962-aa19-d2718eabdeb1'::uuid;

-- ── 3. Template: drop div.pc-card-grid, all .pc-card* CSS, the '1fr 1fr' media rule,
--       the .pc-card-grid mobile rule, the reduced-motion .pc-card selector, and the
--       now-dead inline card-activation <script>. ──
UPDATE content_components
SET html_template = $tpl$<style>
  .provocation-card-section {
    --section-text: rgba(255,255,255,0.9);
    --section-text-muted: rgba(255,255,255,0.7);
    --section-heading: #ffffff;
    --section-surface: rgba(255,255,255,0.05);
    --section-border: rgba(255,255,255,0.2);
    background: var(--color-primary, #1a1a2e);
    color: var(--section-text);
    padding: var(--spacing-section, 5rem 2rem);
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 60vh;
    position: relative;
    overflow: hidden;
  }

  .provocation-card-section::before {
    content: '';
    position: absolute;
    inset: 0;
    background: radial-gradient(ellipse at 30% 50%, var(--section-surface) 0%, transparent 70%);
    pointer-events: none;
  }

  .provocation-card-section .pc-container {
    max-width: var(--container-max-width, 1200px);
    width: 100%;
    margin: 0 auto;
    display: grid;
    grid-template-columns: 1fr;
    gap: 3rem;
    align-items: center;
    position: relative;
    z-index: 1;
  }

  .provocation-card-section .pc-eyebrow {
    display: inline-block;
    font-size: 0.75rem;
    font-weight: 700;
    letter-spacing: 0.15em;
    text-transform: uppercase;
    color: var(--color-accent, var(--color-secondary));
    margin-bottom: 1rem;
    padding: 0.35rem 0.85rem;
    border: 1px solid var(--color-accent, var(--section-border));
    border-radius: calc(var(--border-radius, 0.5rem) * 2);
  }

  .provocation-card-section .pc-headline {
    font-size: clamp(2rem, 5vw, 3.5rem);
    font-weight: 800;
    line-height: 1.1;
    color: var(--section-heading);
    margin: 0 0 1.5rem;
    letter-spacing: -0.02em;
  }

  .provocation-card-section .pc-headline em {
    font-style: normal;
    color: var(--color-accent, var(--color-secondary));
  }

  .provocation-card-section .pc-body {
    font-size: 1.125rem;
    line-height: 1.7;
    color: var(--section-text-muted);
    max-width: 52ch;
    margin: 0 0 2rem;
  }

  .provocation-card-section .pc-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    align-items: center;
  }

  .provocation-card-section .pc-btn-primary {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.875rem 2rem;
    background: var(--color-accent, var(--color-primary));
    color: var(--color-primary-text, #fff);
    font-size: 1rem;
    font-weight: 700;
    border: none;
    border-radius: var(--border-radius, 0.5rem);
    cursor: pointer;
    text-decoration: none;
    transition: transform 0.2s ease, box-shadow 0.2s ease, opacity 0.2s ease;
    min-height: 44px;
  }

  .provocation-card-section .pc-btn-primary:hover,
  .provocation-card-section .pc-btn-primary:focus-visible {
    transform: translateY(-2px);
    box-shadow: 0 8px 24px rgba(0,0,0,0.3);
    opacity: 0.92;
  }

  .provocation-card-section .pc-btn-primary:focus-visible {
    outline: 3px solid var(--color-accent, #fff);
    outline-offset: 3px;
  }

  .provocation-card-section .pc-btn-secondary {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.875rem 1.75rem;
    background: transparent;
    color: var(--section-text);
    font-size: 1rem;
    font-weight: 600;
    border: 1px solid var(--section-border);
    border-radius: var(--border-radius, 0.5rem);
    cursor: pointer;
    text-decoration: none;
    transition: background 0.2s ease, border-color 0.2s ease;
    min-height: 44px;
  }

  .provocation-card-section .pc-btn-secondary:hover,
  .provocation-card-section .pc-btn-secondary:focus-visible {
    background: var(--section-surface);
    border-color: var(--section-text);
  }

  .provocation-card-section .pc-btn-secondary:focus-visible {
    outline: 3px solid var(--section-text);
    outline-offset: 3px;
  }

  .provocation-card-section .pc-stat-strip {
    display: flex;
    flex-wrap: wrap;
    gap: 2rem;
    padding-top: 2rem;
    border-top: 1px solid var(--section-border);
    margin-top: 1rem;
  }

  .provocation-card-section .pc-stat {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .provocation-card-section .pc-stat-value {
    font-size: 1.75rem;
    font-weight: 800;
    color: var(--color-accent, var(--section-heading));
    line-height: 1;
  }

  .provocation-card-section .pc-stat-label {
    font-size: 0.8rem;
    color: var(--section-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  @media (max-width: 768px) {
    .provocation-card-section {
      min-height: auto;
    }

    .provocation-card-section .pc-headline {
      font-size: clamp(1.75rem, 7vw, 2.5rem);
    }

    .provocation-card-section .pc-body {
      font-size: 1rem;
    }

    .provocation-card-section .pc-actions {
      flex-direction: column;
      align-items: stretch;
    }

    .provocation-card-section .pc-btn-primary,
    .provocation-card-section .pc-btn-secondary {
      justify-content: center;
      width: 100%;
    }

    .provocation-card-section .pc-stat-strip {
      gap: 1.25rem;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .provocation-card-section .pc-btn-primary,
    .provocation-card-section .pc-btn-secondary {
      transition: none;
    }
  }
</style>

<section class="provocation-card-section" data-component="provocation-card" data-runtime-fill="true" aria-labelledby="pc-headline">
  <div class="pc-container">
    <div class="pc-content">
      <span class="pc-eyebrow"><no value></span>
      <h2 class="pc-headline" id="pc-headline"><no value></h2>
      <p class="pc-body"><no value></p>
      <div class="pc-actions">
        <a href="<no value>" class="pc-btn-primary" role="button">
          <no value>
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true" focusable="false">
            <path d="M3 8h10M9 4l4 4-4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </a>
        <a href="<no value>" class="pc-btn-secondary" role="button">
          <no value>
        </a>
      </div>
      <div class="pc-stat-strip" aria-label="Key statistics">
        <div class="pc-stat">
          <span class="pc-stat-value"><no value></span>
          <span class="pc-stat-label"><no value></span>
        </div>
        <div class="pc-stat">
          <span class="pc-stat-value"><no value></span>
          <span class="pc-stat-label"><no value></span>
        </div>
        <div class="pc-stat">
          <span class="pc-stat-value"><no value></span>
          <span class="pc-stat-label"><no value></span>
        </div>
      </div>
    </div>
  </div>
</section>$tpl$,
    updated_at = now()
WHERE id = '6163ff14-9f94-4962-aa19-d2718eabdeb1'::uuid;

-- ── 4. Loader: drop the data.lobby fill block. ──
UPDATE js_snippets
SET js_content = $js$/* provocation-card-loader
 * Fetches /data/provocations.json and fills the provocation-card shell.
 * Fails gracefully: if fetch fails or data is malformed, the shell is left
 * as-is (no thrown errors, no broken layout).
 */
(function () {
  "use strict";

  function fillProvocationCard(data) {
    var section = document.querySelector('[data-component="provocation-card"]');
    if (!section || !data || !data.today) return;

    var t = data.today;

    // --- main provocation ---
    var eyebrow = section.querySelector(".pc-eyebrow");
    if (eyebrow && t.eyebrow) eyebrow.textContent = t.eyebrow;

    // headline may contain <em> emphasis — use innerHTML, but only with our
    // own JSON (not user input), so this is safe in this context.
    var headline = section.querySelector(".pc-headline");
    if (headline && t.headline) headline.innerHTML = t.headline;

    var body = section.querySelector(".pc-body");
    if (body && t.body) body.textContent = t.body;

    // --- primary CTA (anchor: set href + label, preserve the inline SVG) ---
    var primary = section.querySelector(".pc-btn-primary");
    if (primary && t.primary_cta) {
      if (t.primary_cta.url) primary.setAttribute("href", t.primary_cta.url);
      if (t.primary_cta.label) setButtonLabel(primary, t.primary_cta.label);
    }

    // --- secondary CTA ---
    var secondary = section.querySelector(".pc-btn-secondary");
    if (secondary && t.secondary_cta) {
      if (t.secondary_cta.url) secondary.setAttribute("href", t.secondary_cta.url);
      if (t.secondary_cta.label) setButtonLabel(secondary, t.secondary_cta.label);
    }

    // --- stat strip (3 stats) ---
    if (Array.isArray(t.stats)) {
      var values = section.querySelectorAll(".pc-stat-value");
      var labels = section.querySelectorAll(".pc-stat-label");
      for (var i = 0; i < values.length; i++) {
        if (t.stats[i]) {
          if (t.stats[i].value != null) values[i].textContent = t.stats[i].value;
          if (labels[i] && t.stats[i].label != null) labels[i].textContent = t.stats[i].label;
        }
      }
    }
  }

  // Replace an anchor's text label without destroying child elements (e.g. SVG).
  // Sets the text of the first text node, or prepends one if none exists.
  function setButtonLabel(anchor, label) {
    var textNode = null;
    for (var k = 0; k < anchor.childNodes.length; k++) {
      if (anchor.childNodes[k].nodeType === 3 && anchor.childNodes[k].textContent.trim() !== "") {
        textNode = anchor.childNodes[k];
        break;
      }
    }
    if (textNode) {
      textNode.textContent = label + " ";
    } else {
      anchor.insertBefore(document.createTextNode(label + " "), anchor.firstChild);
    }
  }

  function init() {
    if (!document.querySelector('[data-component="provocation-card"]')) return;
    fetch("/data/provocations.json", { cache: "no-cache" })
      .then(function (r) {
        if (!r.ok) throw new Error("provocations.json HTTP " + r.status);
        return r.json();
      })
      .then(fillProvocationCard)
      .catch(function (err) {
        if (window.console && console.warn) {
          console.warn("provocation-card-loader: could not load provocations", err);
        }
      });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
$js$
WHERE name = 'provocation-card-loader';

-- ── 5. Verify or abort ──
DO $guard$
DECLARE t RECORD; j RECORD; n_pc INT; n_cc INT; n_sn INT; n_ver INT;
BEGIN
  SELECT length(html_template) AS len,
         (html_template LIKE '%pc-card%')           AS has_card,
         (html_template LIKE '%<script%')           AS has_script,
         (html_template LIKE '%data-runtime-fill%') AS has_marker,
         (html_template LIKE '%1fr 1fr%')           AS has_twocol,
         (length(html_template) - length(replace(html_template,'<no value>','')))/10 AS nv
    INTO t FROM content_components WHERE id = '6163ff14-9f94-4962-aa19-d2718eabdeb1'::uuid;

  SELECT length(js_content) AS len,
         (js_content LIKE '%data.lobby%')    AS has_lobby,
         (js_content LIKE '%pc-card%')       AS has_card,
         (js_content LIKE '%pc-stat-value%') AS has_stats,
         is_active                           AS active
    INTO j FROM js_snippets WHERE name = 'provocation-card-loader';

  SELECT COUNT(*) INTO n_pc FROM _vonc_pc_backup_20260709;
  SELECT COUNT(*) INTO n_cc FROM _vonc_cc_pcard_backup_20260709;
  SELECT COUNT(*) INTO n_sn FROM _vonc_snippet_backup_20260709;
  SELECT COUNT(*) INTO n_ver FROM component_versions
    WHERE component_id = '6163ff14-9f94-4962-aa19-d2718eabdeb1'::uuid AND change_source = 'minilobby_trim_20260709';

  IF n_pc <> 6      THEN RAISE EXCEPTION 'backup page_components: % rows, expected 6', n_pc; END IF;
  IF n_cc <> 1      THEN RAISE EXCEPTION 'backup content_components: % rows, expected 1', n_cc; END IF;
  IF n_sn <> 1      THEN RAISE EXCEPTION 'backup js_snippets: % rows, expected 1', n_sn; END IF;
  IF n_ver <> 1     THEN RAISE EXCEPTION 'component_versions snapshot: % rows, expected 1', n_ver; END IF;

  IF t.len <> 6618  THEN RAISE EXCEPTION 'template length %, expected 6618', t.len; END IF;
  IF t.has_card     THEN RAISE EXCEPTION 'template still contains pc-card'; END IF;
  IF t.has_script   THEN RAISE EXCEPTION 'template still contains an inline script'; END IF;
  IF NOT t.has_marker THEN RAISE EXCEPTION 'template LOST data-runtime-fill'; END IF;
  IF t.has_twocol   THEN RAISE EXCEPTION 'template still has the 1fr 1fr media rule'; END IF;
  IF t.nv <> 13   THEN RAISE EXCEPTION 'template <no value> count %, expected 13', t.nv; END IF;

  IF j.len <> 3365  THEN RAISE EXCEPTION 'loader length %, expected 3365', j.len; END IF;
  IF j.has_lobby    THEN RAISE EXCEPTION 'loader still fills data.lobby'; END IF;
  IF j.has_card     THEN RAISE EXCEPTION 'loader still references pc-card'; END IF;
  IF NOT j.has_stats THEN RAISE EXCEPTION 'loader LOST the stat-strip fill'; END IF;
  IF NOT j.active   THEN RAISE EXCEPTION 'loader is not active'; END IF;

  RAISE NOTICE 'OK  template % chars (<no value> x %)  |  loader % chars  |  backups 6/1/1  |  snapshot 1',
    t.len, t.nv, j.len;
  RAISE NOTICE 'Expected page_components.rendered_html after section-edit: 6488 chars';
END
$guard$;

COMMIT;
