
<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Language handling: implicit mechanism plus minimal explicit prompt support
- **category:** NEW:language-i18n
- **status-signal:** partial
- **status-evidence:** "After Step 3 the page-content-writer prompt has only one explicit language signal — a ## Language section"; "There is no language field on sites, pages, content_components, or site_specs today" (undated FOCUS)
- **what:** Content language rides implicitly on the brief/specs/existing-content context; Step 3 made the page-content-writer prompt language-agnostic (## Language section, de-Anglicised rule examples, translate-the-intent note for English llm_guidance, any-language placeholder rule). Mapped remaining English-hardcoded surfaces: Tier B static fallbacks, admin briefs, strategist internal text, other agents' prompts, missing <html lang>. Deferred designs: sites.primary_language column (add when a consumer exists), explicit target-language parameter, "soft static" LLM override of Tier B labels, adoption-time language detection.
- **sources:** FOCUS_language.md (whole)
- **relations:** tiered field classification (fallback problem); mega-prompt concerns
- **verify-later:** page-content-writer prompt ## Language section; head template lang attribute

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Language handling: implicit mechanism plus minimal explicit prompt support
- **category:** NEW:language-i18n
- **status-signal:** partial
- **status-evidence:** "After Step 3 the page-content-writer prompt has only one explicit language signal — a ## Language section"; "There is no language field on sites, pages, content_components, or site_specs today" (undated FOCUS)
- **what:** Content language rides implicitly on the brief/specs/existing-content context; Step 3 made the page-content-writer prompt language-agnostic (## Language section, de-Anglicised rule examples, translate-the-intent note for English llm_guidance, any-language placeholder rule). Mapped remaining English-hardcoded surfaces: Tier B static fallbacks, admin briefs, strategist internal text, other agents' prompts, missing <html lang>. Deferred designs: sites.primary_language column (add when a consumer exists), explicit target-language parameter, "soft static" LLM override of Tier B labels, adoption-time language detection.
- **sources:** FOCUS_language.md (whole)
- **relations:** tiered field classification (fallback problem); mega-prompt concerns
- **verify-later:** page-content-writer prompt ## Language section; head template lang attribute
