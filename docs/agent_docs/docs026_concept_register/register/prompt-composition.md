# Register — prompt-composition

1 concept, consolidated from 1 raw extraction across unit U02. No duplicates
found within this category's raw material (this new category surfaced a
single raw block from the whole cluster).

### PRC-001 — Mega-prompt fragility and candidate replacement patterns
- **status:** aspirational
- **status-evidence:** "Treat the existing 6KB text prompt as technical debt, not a model … Not blocking imagery or anything else; just don't propagate the pattern" (undated FOCUS doc, ~2026-05).
- **what:** page-content-writer's single ~6KB prompt blends 11+ inputs, 16 growing STRICT RULES, and six worked output schemas. Six fragility concerns were identified: untraceable failures, monotonic rule growth, coupled component vocabulary, a single blend ratio, model coupling, and token waste at scale (roughly 160MB/build-cycle). Five candidate replacement patterns were proposed: per-component templates; a structured intermediate envelope (a cheap-model stage 1 producing a cacheable/lockable structure consumed by a focused stage 2); tool-calling to enforce schema; validation-instead-of-prompt-rules; and a hybrid baseline+overrides approach. The structured-envelope pattern is flagged as strongest for both text and image generation, but none of the five had been implemented at time of writing — this is a design proposal, not a shipped change.
- **sources:** FOCUS_prompt_composition_pattern.md (whole)
- **relations:** image parameter shaping; validate_page_content (a partial existing instance of the validation-instead-of-rules pattern); array-item field contract for the page-content-writer (development-guide DEV-021, a concrete symptom of this same mega-prompt fragility); LLM reliability strategy
- **verify-later:** page-content-writer default_config prompt size/shape today — has any of the five candidate patterns since been adopted?
