-- centre_provocation_card.sql — single-column measure fix (vonc.com / Spark), 2026-07-09
-- Follows trim_minilobby.sql. With the mini-lobby's second grid column gone,
-- .pc-container's 1200px max-width left the copy hugging the left. 820px turns it
-- into a centred single column. One declaration; nothing else changes.
--
--   html_template  6618 -> 6589 chars   (<no value> stays at 13)
--   expected page_components.rendered_html after the section edit: 6459
--
-- Snapshot first: a direct UPDATE bypasses store_generated_component's
-- component_versions snapshotting.

\set ON_ERROR_STOP on
BEGIN;

INSERT INTO component_versions (component_id, version_number, html_template, input_schema, change_source)
SELECT '6163ff14-9f94-4962-aa19-d2718eabdeb1'::uuid,
       COALESCE((SELECT MAX(version_number) FROM component_versions WHERE component_id = '6163ff14-9f94-4962-aa19-d2718eabdeb1'::uuid), 0) + 1,
       html_template, input_schema, 'pc_container_centre_20260709'
FROM content_components WHERE id = '6163ff14-9f94-4962-aa19-d2718eabdeb1'::uuid;

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
    max-width: 820px;
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

DO $g$
DECLARE t RECORD; nver INT;
BEGIN
  SELECT length(html_template) AS len,
         (html_template LIKE '%max-width: 820px%')   AS centred,
         (html_template LIKE '%container-max-width%') AS old_rule,
         (html_template LIKE '%pc-card%')             AS has_card,
         (html_template LIKE '%<script%')             AS has_script,
         (html_template LIKE '%data-runtime-fill%')   AS has_marker,
         (length(html_template) - length(replace(html_template,'<no value>','')))/10 AS nv
    INTO t FROM content_components WHERE id = '6163ff14-9f94-4962-aa19-d2718eabdeb1'::uuid;
  SELECT COUNT(*) INTO nver FROM component_versions
    WHERE component_id = '6163ff14-9f94-4962-aa19-d2718eabdeb1'::uuid AND change_source = 'pc_container_centre_20260709';

  IF nver <> 1        THEN RAISE EXCEPTION 'snapshot rows %, expected 1', nver; END IF;
  IF t.len <> 6589    THEN RAISE EXCEPTION 'template length %, expected 6589', t.len; END IF;
  IF NOT t.centred    THEN RAISE EXCEPTION 'max-width: 820px not applied'; END IF;
  IF t.old_rule       THEN RAISE EXCEPTION 'old container-max-width rule still present'; END IF;
  IF t.has_card       THEN RAISE EXCEPTION 'pc-card reappeared'; END IF;
  IF t.has_script     THEN RAISE EXCEPTION 'inline script reappeared'; END IF;
  IF NOT t.has_marker THEN RAISE EXCEPTION 'lost data-runtime-fill'; END IF;
  IF t.nv <> 13     THEN RAISE EXCEPTION '<no value> count %, expected 13', t.nv; END IF;

  RAISE NOTICE 'OK  template % chars, centred, marker kept; expect rendered_html 6459', t.len;
END $g$;

COMMIT;
