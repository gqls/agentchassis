
<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Mega-prompt fragility and candidate replacement patterns
- **category:** NEW:prompt-composition
- **status-signal:** aspirational
- **status-evidence:** "Treat the existing 6KB text prompt as technical debt, not a model … Not blocking imagery or anything else; just don't propagate the pattern" (undated FOCUS, ~2026-05)
- **what:** page-content-writer's single ~6KB prompt blends 11+ inputs, 16 growing STRICT RULES, and six worked output schemas; six fragility concerns (untraceable failures, monotonic rule growth, coupled component vocabulary, one blend ratio, model coupling, token waste ~160MB/build-cycle at scale). Five candidate patterns: per-component templates; structured intermediate envelope (cheap-model stage 1 → focused stage 2, cacheable, lockable); tool-calling for schema; validation-instead-of-prompt-rules; hybrid baseline+overrides. Envelope (B) flagged strongest for both text and images.
- **sources:** FOCUS_prompt_composition_pattern.md (whole)
- **relations:** image parameter shaping; validate_page_content (pattern D partially exists); LLM reliability strategy
- **verify-later:** page-content-writer default_config prompt size/shape today

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Mega-prompt fragility and candidate replacement patterns
- **category:** NEW:prompt-composition
- **status-signal:** aspirational
- **status-evidence:** "Treat the existing 6KB text prompt as technical debt, not a model … Not blocking imagery or anything else; just don't propagate the pattern" (undated FOCUS, ~2026-05)
- **what:** page-content-writer's single ~6KB prompt blends 11+ inputs, 16 growing STRICT RULES, and six worked output schemas; six fragility concerns (untraceable failures, monotonic rule growth, coupled component vocabulary, one blend ratio, model coupling, token waste ~160MB/build-cycle at scale). Five candidate patterns: per-component templates; structured intermediate envelope (cheap-model stage 1 → focused stage 2, cacheable, lockable); tool-calling for schema; validation-instead-of-prompt-rules; hybrid baseline+overrides. Envelope (B) flagged strongest for both text and images.
- **sources:** FOCUS_prompt_composition_pattern.md (whole)
- **relations:** image parameter shaping; validate_page_content (pattern D partially exists); LLM reliability strategy
- **verify-later:** page-content-writer default_config prompt size/shape today
