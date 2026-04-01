## Replacement section for 001 Development Guide — LLM Infrastructure / Model aliases

### Model aliases

Agent definitions use aliases like `claude-sonnet-4-6` or `claude-haiku-4-5`. These resolve to dated API strings via `model_aliases.go`. Update that file when Anthropic releases new models.

Current model strategy:

| Agent role | Model | Reasoning |
|---|---|---|
| Planning (chief-strategist, site-planner) | claude-sonnet-4-6 | High-leverage structural decisions |
| Research, reasoning, analysis | claude-sonnet-4-6 | Complex analysis, style extraction |
| Content generation (section creators, copywriter) | claude-sonnet-4-6 | Quality content matching site voice |
| Adoption (site classification, content direction) | claude-sonnet-4-6 | Reasoning about site structure and style |
| Orchestration (website-builder) | claude-haiku-4-5 | Minimal LLM use, mostly routing |
| Classification (site-classifier, future local) | claude-haiku-4-5 → ollama | Short structured output, candidate for fine-tuning |

**Rule:** `claude-sonnet-4-6` is the standard model for all LLM steps unless there is a specific reason to use something else (e.g. haiku for orchestration routing where output is trivial, or ollama for fine-tuned classification). When in doubt, use `claude-sonnet-4-6`.

**Rule:** Use short aliases (`claude-sonnet-4-6`, `claude-haiku-4-5`) not full versioned names in agent definitions. The alias map handles version resolution. Never use strings like `claude-sonnet-4-5-20250514`.

### Also update in bug #4:

**Rule:** Use short aliases (`claude-sonnet-4-6`, `claude-haiku-4-5`) not full versioned names in agent definitions. The alias map handles version resolution.
