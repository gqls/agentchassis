# Register — language-i18n

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

1 concept, consolidated from 1 raw extraction across unit U02. No duplicates
found within this category's raw material (this new category surfaced a
single raw block from the whole cluster).

### I18N-001 — Language handling: implicit mechanism plus minimal explicit prompt support
- **status:** partial
- **status-evidence:** "After Step 3 the page-content-writer prompt has only one explicit language signal — a ## Language section"; "There is no language field on sites, pages, content_components, or site_specs today" (undated FOCUS doc).
- **what:** Content language currently rides implicitly on the brief/specs/existing-content context passed into generation prompts, with no dedicated language field anywhere in the schema. A prompt-level change made the page-content-writer prompt language-agnostic: a `## Language` section, de-Anglicised rule examples, a translate-the-intent note for otherwise-English llm_guidance text, and an any-language placeholder rule. Remaining English-hardcoded surfaces were mapped but not fixed: Tier B static fallbacks, admin briefs, strategist internal text, other agents' prompts, and a missing `<html lang>` attribute on generated pages. Deferred design ideas: a `sites.primary_language` column (to be added only once a real consumer needs it), an explicit target-language parameter, a "soft static" LLM override of Tier B labels, and adoption-time language detection.
- **sources:** FOCUS_language.md (whole)
- **relations:** tiered field classification / static-fallback problem; mega-prompt fragility concerns (prompt-composition register, same source document family)
- **verify-later:** page-content-writer prompt's `## Language` section (still present?); head template `<html lang>` attribute; whether `sites.primary_language` or any language field has since been added
