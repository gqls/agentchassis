-- ============================================================================
-- intent-probe — section component for traffic-probe sites (P3)
-- ============================================================================
-- STEP ZERO (0a–0g) verdict, 2026-06-10: no existing active section captures
-- anonymous intent server-side. Nearest: contact-form (collects PII — exactly
-- what the probe must NOT collect; different function under the one-function-
-- one-component rule), tool-* (client-side JS calculators), call-to-action
-- (link-out, no form). Therefore: new component, function `intent-probe`.
--
-- Contract compliance (003):
--   - function/section_type kebab-case (CHECK constraints pass)
--   - data-component on the root element matches function exactly
--   - all copy comes from template variables (no hardcoded content)
--   - CSS scoped to .intent-probe-section, var(--*, fallback) theming only,
--     no global element rules, @media 768px included
--   - NO JavaScript (js_content NULL) — plain form POST + 1x1 beacon img, so
--     the JS Content Separation contract is satisfied trivially
--   - input_schema v2 (fields/type/source/required/on_missing)
--   - is_dark_section = false (no --section-* contract needed)
--
-- Engine contract (probe-go service.go): form POSTs kind+value to /intent;
-- beacon GET /api/hit counts the visit. probe_kind must be one of
-- search|categories|freetext (engine coerces unknown → freetext).
--
-- DELIBERATE v1 LIMIT: renders ONE text-input action (covers the `search` and
-- `freetext` kinds). The category-buttons variant needs {{range}} over an
-- array; deferred until the section renderer's array handling is verified
-- (the library's own quality flags show arrays are where templates fail).
--
-- Gating (capability link, layer 1): suitable_site_types restricts this to
-- probe-type sites. NOTE the planner's load_components query currently loads
-- ALL active sections — apply the opt-in gate below so restricted components
-- are reachable ONLY via roadmap section_types pinning.
-- ============================================================================

-- ---------------------------------------------------------------------------
-- PRE-FLIGHT (run before the INSERT; per guidelines, check don't assume)
-- ---------------------------------------------------------------------------
-- 1. Name/function free?
--    SELECT name, function, is_active FROM content_components
--    WHERE name = 'intent-probe' OR function = 'intent-probe';
-- 2. Does anything consume suitable_site_types at plan time? (expect: the
--    planner does NOT — hence the gate change below)
-- 3. Verify the site_specs identity path used for contact_email against a
--    real row before trusting plan_sections resolution:
--    SELECT specs->'identity' FROM site_specs WHERE site_id = '<probe-site>'
--    ORDER BY created_at DESC LIMIT 1;

INSERT INTO content_components (
    name, display_name, description,
    html_template, input_schema,
    function, category, section_type, component_level, render_mode,
    is_active, is_dark_section,
    suitable_site_types, suitable_page_types,
    visual_density, created_from, semantic_tags
) VALUES (
    'intent-probe',
    'Intent Probe',
    'Anonymous visitor-intent capture for traffic-probe sites: one invited action (a single text input) posting server-side to /intent, plus a no-cookie 1x1 visit beacon and a plain privacy line. Collects no names, no emails, no PII. Requires a VM backend (the probe capture engine) — must NOT be placed on static-only (B2/Worker) sites.',
    $tpl$<style>
  .intent-probe-section {
    padding: var(--spacing-section, 5rem 2rem);
    background-color: var(--color-background, #ffffff);
    color: var(--color-text, #333333);
  }
  .intent-probe-section .container {
    max-width: var(--container-max-width, 720px);
    margin: 0 auto;
  }
  .intent-probe-section .probe-tagline {
    font-size: 1.125rem;
    line-height: 1.6;
    color: var(--color-text-muted, #666666);
    margin: 0 0 2rem 0;
  }
  .intent-probe-section .probe-label {
    display: block;
    font-size: 1rem;
    font-weight: 600;
    color: var(--color-heading, #1a1a1a);
    margin: 0 0 0.75rem 0;
  }
  .intent-probe-section .probe-input {
    width: 100%;
    box-sizing: border-box;
    font-size: 1.0625rem;
    font-family: inherit;
    color: var(--color-text, #333333);
    background-color: var(--color-card-bg, #ffffff);
    padding: 0.875rem 1rem;
    border: 1px solid var(--color-border, #e5e5e5);
    border-radius: var(--border-radius, 8px);
  }
  .intent-probe-section .probe-submit {
    margin-top: 1rem;
    font-size: 1rem;
    font-weight: 500;
    font-family: inherit;
    color: var(--color-button-text, #ffffff);
    background-color: var(--color-primary, #0066cc);
    padding: 0.75rem 1.5rem;
    border: 0;
    border-radius: var(--border-radius, 8px);
    cursor: pointer;
  }
  .intent-probe-section .probe-privacy {
    margin-top: 2.5rem;
    padding-top: 1.25rem;
    border-top: 1px solid var(--color-border, #e5e5e5);
    font-size: 0.8125rem;
    line-height: 1.6;
    color: var(--color-text-muted, #666666);
  }
  .intent-probe-section .probe-privacy a {
    color: var(--color-text-muted, #666666);
  }
  @media (max-width: 768px) {
    .intent-probe-section {
      padding: 3rem 1.5rem;
    }
    .intent-probe-section .probe-submit {
      width: 100%;
    }
  }
</style>

<section class="intent-probe-section" data-component="intent-probe">
  <div class="container">
    {{if .tagline}}<p class="probe-tagline">{{.tagline}}</p>{{end}}
    <form method="post" action="/intent">
      <input type="hidden" name="kind" value="{{.probe_kind}}">
      <label class="probe-label" for="intent-probe-value">{{.action_label}}</label>
      <input class="probe-input" type="text" id="intent-probe-value" name="value"
             autocomplete="off" maxlength="500"{{if .placeholder}} placeholder="{{.placeholder}}"{{end}}>
      <button class="probe-submit" type="submit">{{.submit_label}}</button>
    </form>
    <p class="probe-privacy">{{.privacy_text}}{{if .contact_email}} Questions: <a href="mailto:{{.contact_email}}">{{.contact_email}}</a>.{{end}}</p>
    <img src="/api/hit" width="1" height="1" alt="" aria-hidden="true">
  </div>
</section>$tpl$,
    $schema${
  "fields": {
    "tagline": {
      "type": "text",
      "source": "llm",
      "required": false,
      "llm_guidance": "One short line that plausibly reflects what this domain used to be about (from the site identity/classification). No salesiness, no promises. Omit rather than pad."
    },
    "action_label": {
      "type": "text",
      "source": "llm",
      "required": true,
      "llm_guidance": "The single invited action, phrased as a direct question to the visitor in the site language, e.g. 'What are you looking for on this site?' or for search-style probes 'What are you searching for? (brand, model, repair)'. One sentence."
    },
    "placeholder": {
      "type": "text",
      "source": "llm",
      "required": false,
      "llm_guidance": "Optional short example text for the input, e.g. a typical query for this vertical. Omit if not obviously useful."
    },
    "submit_label": {
      "type": "text",
      "source": "llm",
      "required": true,
      "llm_guidance": "Button text: 'Search' for search-style probes, 'Send' otherwise, in the site language. One word."
    },
    "probe_kind": {
      "type": "text",
      "source": "config.intent_probe.kind",
      "required": true,
      "on_missing": "use_fallback",
      "fallback": "search",
      "llm_guidance": "Machine value, one of: search | freetext. Comes from site config; never invent."
    },
    "privacy_text": {
      "type": "text",
      "source": "config.intent_probe.privacy_text",
      "required": true,
      "on_missing": "use_fallback",
      "fallback": "This page is run by the domain owner. When you use the box above we record only what you type, a coarse country, and the site you came from - no names, no cookies, no third-party tracking, and no IP addresses. We keep it only as long as needed to understand what visitors are looking for."
    },
    "contact_email": {
      "type": "text",
      "source": "site_specs.identity.contact_email",
      "required": false,
      "on_missing": "skip_field",
      "missing_reason": "Privacy-line contact address; section renders without it."
    }
  }
}$schema$::jsonb,
    'intent-probe',
    'cta',
    'intent-probe',
    'section',
    'template',
    true,
    false,
    '["intent-probe"]'::jsonb,
    '["landing"]'::jsonb,
    'low',
    'manual',
    '["intent", "capture", "probe", "requires-backend"]'::jsonb
)
ON CONFLICT (name) DO NOTHING;

-- Verify:
-- SELECT name, function, section_type, suitable_site_types, semantic_tags,
--        template_closed, js_content IS NULL AS no_js
-- FROM content_components WHERE name = 'intent-probe';

-- ---------------------------------------------------------------------------
-- LAYER-1 GATE (separate change, build-site-planner workflow JSON, NOT here):
-- the planner's load_components step currently loads ALL active sections, so a
-- restricted component would be offered to every site. Change its query to
-- make restricted components opt-in (roadmap-pinned only):
--
--   SELECT name, display_name, "function", category, description
--   FROM content_components
--   WHERE component_level IN ('section', 'element')
--     AND is_active = true
--     AND suitable_site_types = '[]'::jsonb      -- << the one-line gate
--   ORDER BY category, name
--
-- Apply via the agent-definition editing process (default_config jsonb), not a
-- hand-written UPDATE. Probe pages then receive intent-probe via the roadmap's
-- section_types ("ROADMAP OVERRIDES THE COMPONENT LIST"), which the component
-- selector resolves downstream.
--
-- LAYER 2 (structural, later, when backend sections multiply): capability
-- match — sites.deploy_config.capabilities ⊇ component requirements (the
-- 'requires-backend' semantic tag above is the component-side marker), with
-- one audit check writing site_work_items findings on mismatch.
-- ---------------------------------------------------------------------------
