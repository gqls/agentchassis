-- 267_tool_guide_intro_recovery_waterfall.sql
-- First use of `tool-guide-intro` on this site: the owner's ask #3 from
-- 2026-07-27 ("more explanation, helpful guides at each point"), landing on
-- the page that most needed it - the recovery-waterfall tool served with no
-- introduction and, until now, NO H1 (its own heading is an H2; this section's
-- H1 becomes the page's only one, fixing a real heading-structure gap).
--
-- The component is render_mode='agent' (LLM-filled fields elsewhere). On THIS
-- site copy is authored, so the content_data was written by hand in the house
-- voice and the template executed offline through text/template, mirroring
-- RenderTemplateWithMap (parse, execute, strip "<no value>") - the same
-- harness as migs 252/255/266. NO figures anywhere in the copy, so nothing
-- needs the register. Step copy uses the tool's OWN control labels
-- ("enterprise value", "New money, super-senior", "Write-down of the senior
-- claim"), read from the served page, so the guide describes the real tool.
--
-- Contrast pre-measured against the palette before writing (the 253 rule):
-- eyebrow #86ADDE on #1B2A3B = 6.27, button ink #0F1820 on #86ADDE = 7.4,
-- muted text on surface = 5.12 (the pair 253 measured). Meta icons are
-- aria-hidden SVG strokes with no text, so the claims gate has nothing to
-- miss. The CTA targets /cases/thames-water.html, live and in `pages.url`.
--
-- The page is rebuild_policy='owned': save_page_sections refuses it, so after
-- this INSERT the deploy is ASSEMBLE-ONLY (049b_deploy_single_page.sh with no
-- reason argument - RUNBOOK_oufe.md 8b). Row locked permanent with
-- rendered_html written in the same statement (182 pattern).
--
-- REPLAY GUARD from the start this time (WRONG_CALLS 2026-07-29: an
-- unguarded replay of a jsonb-append migration duplicates silently).
-- Here both statements are naturally idempotent: NOT EXISTS on the slot, and
-- jsonb_agg(DISTINCT) on sections.

BEGIN;

WITH pg AS (SELECT id FROM pages WHERE site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND name='tool-recovery-waterfall'),
     comp AS (SELECT id FROM content_components WHERE name='tool-guide-intro')
INSERT INTO page_components (page_id, component_id, position, slot_name, content_data,
                             rendered_html, build_status, locked_at, locked_by, lock_type)
SELECT pg.id, comp.id, 0, 'tool-guide-intro', $gc${
  "eyebrow_label": "Tool guide",
  "headline": "Work a recovery waterfall for yourself",
  "lead_paragraph": "The tool below takes a company's enterprise value and stacks its debts in priority order, then shows how far down the stack the money reaches. Every number in it is deliberately hypothetical — the point is the mechanism, not any real company. Change one thing at a time and watch where the loss lands.",
  "skill_level_label": "Assumes",
  "skill_level_value": "No prior knowledge",
  "audience_label": "Written for",
  "audience_value": "Anyone following a live restructuring",
  "steps_list_label": "How to use the tool",
  "step_1_title": "Accept the caveat",
  "step_1_desc": "The tool opens with a plain statement that it can be wrong. Read it and accept it — everything below is a simplified model, not advice.",
  "step_2_title": "Set the value and the debt stack",
  "step_2_desc": "Choose an enterprise value, then enter the claims in priority order — new money super-senior, senior, junior, subordinated — with the sliders or typed amounts.",
  "step_3_title": "Read where the value runs out",
  "step_3_desc": "Each class is paid in order until the value is gone. Apply a write-down to the senior claim and watch who absorbs it.",
  "cta_primary_url": "/cases/thames-water.html",
  "cta_primary_label": "Read the worked case: Thames Water"
}$gc$::jsonb,
       $gh$<style>
  .tool-guide-intro-section {
    background: var(--color-surface);
    padding: var(--spacing-section, 5rem 2rem);
    color: var(--color-text);
  }

  .tool-guide-intro-section .tgi-container {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 3rem;
    align-items: center;
  }

  .tool-guide-intro-section .tgi-eyebrow {
    display: inline-block;
    font-size: 0.8rem;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--color-primary);
    background: color-mix(in srgb, var(--color-primary) 10%, transparent);
    border: 1px solid color-mix(in srgb, var(--color-primary) 25%, transparent);
    border-radius: calc(var(--border-radius, 0.5rem) * 2);
    padding: 0.25rem 0.75rem;
    margin-bottom: 1rem;
  }

  .tool-guide-intro-section .tgi-headline {
    font-size: clamp(1.75rem, 4vw, 2.75rem);
    font-weight: 800;
    line-height: 1.2;
    color: var(--color-heading);
    margin: 0 0 1rem;
  }

  .tool-guide-intro-section .tgi-lead {
    font-size: 1.1rem;
    line-height: 1.7;
    color: var(--color-text-muted);
    margin: 0 0 1.5rem;
  }

  .tool-guide-intro-section .tgi-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    margin-bottom: 2rem;
  }

  .tool-guide-intro-section .tgi-meta-item {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.875rem;
    color: var(--color-text-muted);
  }

  .tool-guide-intro-section .tgi-meta-icon {
    width: 16px;
    height: 16px;
    flex-shrink: 0;
    color: var(--color-primary);
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .tool-guide-intro-section .tgi-meta-label {
    font-weight: 600;
    color: var(--color-text);
  }

  .tool-guide-intro-section .tgi-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
    align-items: center;
  }

  .tool-guide-intro-section .tgi-btn-primary {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    background: var(--color-primary);
    color: var(--color-primary-text, var(--color-white, #fff));
    border: none;
    border-radius: var(--border-radius, 0.5rem);
    padding: 0.75rem 1.5rem;
    font-size: 1rem;
    font-weight: 700;
    text-decoration: none;
    cursor: pointer;
    min-height: 44px;
    transition: background 0.2s ease;
  }

  .tool-guide-intro-section .tgi-btn-primary:hover,
  .tool-guide-intro-section .tgi-btn-primary:focus-visible {
    background: var(--color-primary-hover, var(--color-primary));
    outline: 2px solid var(--color-primary);
    outline-offset: 2px;
  }

  .tool-guide-intro-section .tgi-btn-secondary {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    background: transparent;
    color: var(--color-primary);
    border: 2px solid var(--color-primary);
    border-radius: var(--border-radius, 0.5rem);
    padding: 0.75rem 1.5rem;
    font-size: 1rem;
    font-weight: 700;
    text-decoration: none;
    cursor: pointer;
    min-height: 44px;
    transition: background 0.2s ease, color 0.2s ease;
  }

  .tool-guide-intro-section .tgi-btn-secondary:hover,
  .tool-guide-intro-section .tgi-btn-secondary:focus-visible {
    background: color-mix(in srgb, var(--color-primary) 10%, transparent);
    outline: 2px solid var(--color-primary);
    outline-offset: 2px;
  }

  .tool-guide-intro-section .tgi-visual {
    position: relative;
  }

  .tool-guide-intro-section .tgi-image-wrap {
    border-radius: var(--border-radius, 0.5rem);
    overflow: hidden;
    box-shadow: var(--shadow, 0 4px 24px rgba(0,0,0,0.1));
    aspect-ratio: 4/3;
    background: var(--color-card-bg);
  }

  .tool-guide-intro-section .tgi-image-wrap img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  .tool-guide-intro-section .tgi-steps {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
  }

  .tool-guide-intro-section .tgi-step {
    display: flex;
    gap: 1rem;
    align-items: flex-start;
  }

  .tool-guide-intro-section .tgi-step-num {
    flex-shrink: 0;
    width: 2rem;
    height: 2rem;
    border-radius: 50%;
    background: var(--color-primary);
    color: var(--color-primary-text, var(--color-white, #fff));
    font-size: 0.85rem;
    font-weight: 800;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-top: 0.15rem;
  }

  .tool-guide-intro-section .tgi-step-title {
    font-size: 0.95rem;
    font-weight: 700;
    color: var(--color-heading);
    margin: 0 0 0.2rem;
  }

  .tool-guide-intro-section .tgi-step-desc {
    font-size: 0.875rem;
    line-height: 1.5;
    color: var(--color-text-muted);
    margin: 0;
  }

  @media (max-width: 768px) {
    .tool-guide-intro-section .tgi-container {
      grid-template-columns: 1fr;
      gap: 2rem;
    }

    .tool-guide-intro-section .tgi-visual {
      order: -1;
    }

    .tool-guide-intro-section .tgi-actions {
      flex-direction: column;
      align-items: stretch;
    }

    .tool-guide-intro-section .tgi-btn-primary,
    .tool-guide-intro-section .tgi-btn-secondary {
      justify-content: center;
      width: 100%;
    }
  }
</style>

<section class="tool-guide-intro-section" data-component="tool-guide-intro">
  <div class="tgi-container">
    <div class="tgi-content">
      <span class="tgi-eyebrow">Tool guide</span>
      <h1 class="tgi-headline">Work a recovery waterfall for yourself</h1>
      <p class="tgi-lead">The tool below takes a company's enterprise value and stacks its debts in priority order, then shows how far down the stack the money reaches. Every number in it is deliberately hypothetical — the point is the mechanism, not any real company. Change one thing at a time and watch where the loss lands.</p>

      <div class="tgi-meta">
        
        <div class="tgi-meta-item">
          <svg class="tgi-meta-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/></svg>
          <span class="tgi-meta-label">Assumes</span>
          <span>No prior knowledge</span>
        </div>
        <div class="tgi-meta-item">
          <svg class="tgi-meta-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
          <span class="tgi-meta-label">Written for</span>
          <span>Anyone following a live restructuring</span>
        </div>
      </div>

      <div class="tgi-actions">
        <a href="/cases/thames-water.html" class="tgi-btn-primary">Read the worked case: Thames Water</a>
        
      </div>
    </div>

    <div class="tgi-visual">
      

      <ol class="tgi-steps" aria-label="How to use the tool">
        <li class="tgi-step">
          <span class="tgi-step-num" aria-hidden="true">1</span>
          <div>
            <p class="tgi-step-title">Accept the caveat</p>
            <p class="tgi-step-desc">The tool opens with a plain statement that it can be wrong. Read it and accept it — everything below is a simplified model, not advice.</p>
          </div>
        </li>
        <li class="tgi-step">
          <span class="tgi-step-num" aria-hidden="true">2</span>
          <div>
            <p class="tgi-step-title">Set the value and the debt stack</p>
            <p class="tgi-step-desc">Choose an enterprise value, then enter the claims in priority order — new money super-senior, senior, junior, subordinated — with the sliders or typed amounts.</p>
          </div>
        </li>
        <li class="tgi-step">
          <span class="tgi-step-num" aria-hidden="true">3</span>
          <div>
            <p class="tgi-step-title">Read where the value runs out</p>
            <p class="tgi-step-desc">Each class is paid in order until the value is gone. Apply a write-down to the senior claim and watch who absorbs it.</p>
          </div>
        </li>
      </ol>
    </div>
  </div>
</section>
$gh$, 'deployed', now(), 'oufe-workstream', 'permanent'
FROM pg, comp
WHERE NOT EXISTS (SELECT 1 FROM page_components pc JOIN pages p ON p.id=pc.page_id
  WHERE p.site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND p.name='tool-recovery-waterfall'
    AND pc.slot_name='tool-guide-intro');

UPDATE pages SET sections=(SELECT jsonb_agg(DISTINCT x)
  FROM jsonb_array_elements(COALESCE(sections,'[]'::jsonb) || '["tool-guide-intro"]'::jsonb) x)
WHERE site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND name='tool-recovery-waterfall';

COMMIT;

-- VERIFY: two slots, guide at position 0 above the tool, locked with real HTML.
SELECT pc.position, pc.slot_name, pc.lock_type, length(pc.rendered_html) AS len
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND p.name='tool-recovery-waterfall'
ORDER BY pc.position;
