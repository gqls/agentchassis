-- Slice 4a rollback: the four replaces, swapped. (Belt: the .bak dump.)
UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{prompt_template}', to_jsonb(
 replace(replace(replace(replace(default_config->>'prompt_template',
 '- Inside painted sections, text consumes the chain: color: var(--section-text, var(--color-text)); headings: var(--section-heading, var(--color-heading))',
 '- Dark sections: color: var(--section-text, rgba(255,255,255,0.9))'),
 E'6. SECTION PAINTING (choose exactly ONE; appearance derives from what YOUR CSS paints):\n   (a) PAIR BAND: background: var(--color-cta-bg) (or the header/footer pair where\n       appropriate) and re-export on the root container AS REFERENCES ONLY:\n         --section-text: var(--color-cta-text);\n         --section-heading: var(--color-cta-text);\n         --section-text-muted: color-mix(in srgb, var(--color-cta-text) 70%, transparent);\n         --section-border: color-mix(in srgb, var(--color-cta-text) 25%, transparent);\n         --section-surface: color-mix(in srgb, var(--color-cta-text) 8%, transparent);\n   (b) PALETTE BAND: background: var(--color-primary) and re-export the on-colour family:\n         --section-text: var(--color-primary-text, var(--color-background));\n         (heading/link likewise; muted/border/surface via color-mix as above)\n   (c) IMAGE/LAYERED: define --hero-ink per branch and re-export --section-* from the\n       ink (see the hero component for the model)\n   (d) AMBIENT (no background of its own): declare NO --section-* variables at all\n   LITERAL COLOURS IN --section-* DECLARATIONS ARE FORBIDDEN — references only.\n   The is_dark_section flag in your output is catalogue metadata ONLY: report it\n   honestly, but nothing may style from it — the painting rules above decide.',
 E'6. DARK SECTIONS (if the section has a dark background):\n   Set on the root container:\n     --section-text: rgba(255,255,255,0.9);\n     --section-text-muted: rgba(255,255,255,0.7);\n     --section-heading: #ffffff;\n     --section-surface: rgba(255,255,255,0.05);\n     --section-border: rgba(255,255,255,0.2);'),
 E'--color-footer-bg, --color-footer-text, --color-white\n     --color-cta-bg, --color-cta-text\n     --color-surface-alt, --color-hairline, --color-code-bg\n     --color-callout-bg, --color-callout-border',
 '--color-footer-bg, --color-footer-text, --color-white'),
 E'Mark required: true only if the section genuinely cannot render without them.\n     IMAGE FIELDS RULE: any field sourced from site_assets.* MUST be\n     "required": false with "on_missing": "skip_field", and the template MUST\n     gate the image markup with a Go-template conditional on the field (the\n     same if-block form brief-explanation uses around its image wrapper), so\n     the section renders cleanly before the image exists. Imagery arrives\n     asynchronously and must never block or defer a section.',
 'Mark required: true only if the section genuinely cannot render without them.')
)),
updated_at = now()
WHERE type = 'component-creator' AND is_active = true
  AND default_config->>'prompt_template' LIKE '%SECTION PAINTING%'
RETURNING position('DARK SECTIONS (if the section' in default_config->>'prompt_template') > 0 AS restored;
