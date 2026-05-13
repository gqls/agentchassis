# FOCUS — Prompt Composition Pattern

For future discussion. Not a plan. The decision to defer this surfaced during Phase 2G step 5 design when comparing how `page-content-writer` composes text prompts vs how `generate_image_actions.go` composes image prompts.

The text pattern is the natural reference for "how should imagery do this?" but on closer inspection the text pattern itself has problems worth surfacing before we copy it.

---

## The text pattern today

`page-content-writer`'s `generate_content` step uses a single LLM call with a ~6KB Go-template prompt that blends, in one mega-prompt:

- Company context (name, industry, tone, audience, services, tagline)
- Official contact info (with "USE ONLY THESE — DO NOT INVENT" framing)
- Link context (internal linking constraints)
- `site_specs.content_direction.formatted`
- `site_specs.identity.target_audience` and `key_differentiators`
- `site_specs.design_intent.imagery_direction` (read by the text writer too)
- Section's `content_brief` (admin overrides)
- Component `input_schema` (data contract)
- Research findings with source citations
- Existing content (for adoption rewrites)
- Rewrite guidance (for refinement passes)

Then 16 STRICT RULES at the end (negative reinforcement: "NEVER invent…", "NEVER fabricate…"), then six worked JSON output schemas (hero, feature, CTA, text, contact, testimonial, case study) — the template knows how to format every component type.

It works. Output quality on robot-hands.com and earlier sites is broadly fine. But it works in spite of, not because of, the structure.

---

## Why this pattern is fragile

Six concerns, ordered by how much they'd hurt at scale:

### 1. Untraceable failure modes

When a section ships bad content, the prompt has 11+ inputs blended into it. There's no way to attribute the failure: was it the brief? The identity? The research? The rewrite guidance overriding the brief? They all flowed into the same LLM call. Debugging means reading 6KB of rendered prompt and guessing. At 100 sites this is annoying; at 10,000 it becomes a real cost.

### 2. Monotonic growth

The "16 strict rules" section grows every time a failure pattern is observed. Rule 11 is two sentences forbidding fake testimonials. Rule 14 is three sentences specifying what to write instead. Rules 12 and 13 forbid statistics and named businesses. Each addition is a small patch; the aggregate is a prompt where ~30% of the bytes are negative shouting at the model. Models trained on later instruction-tuning data respond worse to this kind of structure, not better.

### 3. Coupled component vocabulary

The template knows about hero, feature, CTA, text, contact, testimonial, and case-study sections — each with a worked JSON example. Adding a new component type requires editing this template. Removing a component type leaves dead schema in the prompt unless someone notices. The template is acting as both a generator and a schema dispatcher; those are two responsibilities that ought to be separable.

### 4. One blend ratio for every section

A hero section wants strong tagline / identity emphasis and minimal research. A FAQ wants the brief's exact wording and heavy research weighting. A "Contact" section wants almost nothing except the contact data. One template can't optimise for all three — it just dumps everything and trusts the model to weigh things appropriately. The model usually does, but not reliably.

### 5. Model coupling

A prompt this size has subtle behavioural signatures that aren't portable. The rules ordering matters. The output-schema examples teach a default format. When we swap models — Sonnet to Opus, Anthropic to local Llama 70B per doc 009's roadmap — the prompt needs revalidation. The longer the prompt, the more we're tuning to a specific model's quirks.

### 6. Token waste at scale

Most of the 6KB is invariant: the rules, the schemas, the boilerplate context labels. They render in every single section's prompt. At ~20 sections/site × 2000 sites = 40K sections in steady state, each carrying ~4KB of repeated boilerplate, that's ~160MB of identical input tokens billed across the fleet per build cycle. Not catastrophic, but real money for a structural reason.

---

## What better patterns might look like

Five candidate replacements, listed for later discussion. Not recommendations yet.

### A. Component-specific templates

Instead of one mega-template with N output schemas, N small templates each ~1KB. Dispatch happens before the LLM call based on `current_section.category`. Each template carries only the schema and context relevant to its category.

- **Pro:** Smaller, focused, easier to tune per category.
- **Pro:** Adding a new component type means a new file, not editing a central one.
- **Con:** Some duplication of common context across templates. Could be addressed with a base template that gets composed.
- **Con:** Loses the implicit cross-component consistency the mega-template provides.

### B. Structured intermediate envelope

Two-stage. Stage 1: a cheap LLM call (Haiku or Mistral-Small) consumes site_specs + brief + research and produces a small structured "envelope" — JSON with `voice`, `key_points`, `must_mention`, `must_avoid`, etc. Stage 2: the main generator (Sonnet or Opus) receives only the envelope + the component schema. Much smaller prompt, much more focused output.

- **Pro:** Envelope is cacheable. One envelope per page; reused across sections.
- **Pro:** Envelope is HITL-lockable like any other structured artefact.
- **Pro:** Failures separate cleanly: envelope failures are diagnosable independently from generation failures.
- **Con:** Adds an LLM call (cost and latency).
- **Con:** New abstraction layer to maintain.

### C. Tool calling for schema

Move the JSON output schemas out of the prompt and into the API's tool-calling contract. Each component category becomes a tool the model can call. The schema is enforced at the API layer, not coaxed by example.

- **Pro:** Eliminates 1-2KB of the prompt immediately.
- **Pro:** Schema validation is structural rather than instructional — no more "use these EXACT field names" warnings.
- **Pro:** Output is parsed not extracted — no JSON-from-prose regex failures.
- **Con:** Locks in to Anthropic-tool-call shape. Llama 3.3 tool calling is workable but different. Local-model swap path gets harder.
- **Con:** Some loss of expressivity for components whose output shape is genuinely freeform.

### D. Validation rules instead of prompt rules

The "16 STRICT RULES" become post-generation validation rather than pre-generation instruction. Generate, validate against the rules, regenerate failing sections with the failure as targeted feedback.

- **Pro:** Smaller prompts.
- **Pro:** Rules become testable code, not LLM-interpreted prose.
- **Pro:** Each rule has a defined check and a defined fix prompt; failures are diagnosable per rule.
- **Con:** Adds a verification step in the workflow.
- **Con:** Some rules don't reduce to checks easily ("professional but engaging tone").

### E. Hybrid baseline + per-category overrides

A thin baseline prompt (~500 tokens) carrying the common context. Per-category override blocks added only where they meaningfully change behaviour. Most sections use mostly the baseline; specific categories layer specific tuning.

- **Pro:** Captures the cross-section consistency benefit of today's pattern.
- **Pro:** Avoids the worst duplication of A.
- **Con:** Templates that compose are harder to read than templates that don't. The Go-template syntax doesn't help.

These aren't mutually exclusive. B + C is plausible. A + D is plausible. E alone is conservative.

---

## What this means for images

The image cascade question (Phase 2G step 5's open follow-up) was originally framed as "should images match text's composition pattern?" The honest answer is "the text pattern isn't a great target to match." But the underlying need — composing multiple inputs into the generator call — is real.

For images specifically the constraints are different from text in ways that change the answer materially:

### Images have a tiny prompt budget

SDXL's CLIP text encoder caps at 77 tokens. Everything beyond that is silently truncated. The "blend N sources into one prompt" approach text uses has no analogue here — the prompt CAN'T be 6KB because the encoder won't see most of it.

### Images respond to parameters, not negative prose

Negative reinforcement in image prompts ("no people, no logos, no text") works against SDXL because the model has no "don't" understanding. The right way to forbid people is `negative_prompt: "people, faces, humans"` as a separate parameter, not "no people" in the subject prompt.

This means the image cascade is mostly about parameter shaping, not prompt shaping:

- `subject` (the brief) — short, positive, concrete
- `negative_prompt` — derived from kind (logo gets "people, faces"; icon gets "complex backgrounds"; etc.)
- `style_preset` or LoRA — derived from imagery_direction
- `reference_image_uri` (img2img) — derived from adopted images (post-Phase-3.1) or per-kind brand exemplars
- `aspect_ratio` — derived from kind and style_hints.aspect
- `cfg_scale`, `steps` — per-kind tuning

This is what Phase 2H ("image generator request shape") is supposed to deliver. The image cascade probably lives in Phase 2H more than it lives in a step-5-style prompt extension.

### A composer step might still be useful

There's room for one cheap composition step BEFORE the generator: takes the imagery row + design_intent.imagery_direction + identity + style_hints + constraints, outputs a parameter envelope. Then `image-generator` receives concrete parameters rather than synthesising them from scattered inputs.

Possibly a small action (`compose_image_request`) sitting between `spawn_image_gen_imagery` and `call_imagery_gen`. Possibly an LLM call (cheap model) that produces the envelope. Possibly pure Go logic with per-kind rules. The pattern B "envelope" idea applies cleanly here.

---

## Recommendation for Phase 2G

Don't try to bring the text cascade into images during step 5. The step 5 migration as drafted is correct: `kind`, `style_hints`, `constraints` pass through to image-generator inertly. The composer step is its own work.

When we do come back to it, my opinion is:

1. **Don't replicate the text pattern.** It has the problems above.
2. **Use Phase 2H's request-shape work as the place to land image cascade.** Parameter-shaping fits naturally in the work that adds negative_prompt, style_preset, reference_image_uri etc. to the image-generator request.
3. **Consider pattern B (envelope) for both text and images.** The same shape works for both. Stage 1 produces a small structured envelope of generation parameters; stage 2 (text or image) consumes it. If we ever revisit the text cascade, the envelope pattern is the strongest candidate.
4. **Treat the existing 6KB text prompt as technical debt, not a model.** Worth a dedicated phase to revisit when there's space for it. Not blocking imagery or anything else; just don't propagate the pattern.

---

## Pointer for later

When the time comes:

- Sibling FOCUS docs to read alongside: `FOCUS_imagery_assessment.md` (image-side context), parts of doc 009 (model swap implications), `PLAN_product_illustration.md` (per-product prompt composition was lightly sketched there too).
- Files to inspect: `platform/orchestration/actions/generate_image_actions.go` (image side), `agent_definitions.type='page-content-writer'.default_config` (text side, mega-prompt lives here), `platform/aiservice/anthropic.go` (tool-calling capability check for pattern C).
- Sites with rich enough imagery to be useful as test beds for parameter-shaping: robot-hands.com once Phase 2G completes (~14 imagery rows across 5 kinds).
