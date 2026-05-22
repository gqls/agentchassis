# Planner Prompt Fixes — Defect 1 (Duplicate Content Surfaces)

Before/after edits against the two planner `agent_definitions` pulled from
the DB, plus the `applyNewPage` default-sections change. Pairs with the
`validate_components` Go implementation (separate doc) which handles
Defect 2 (display-name leak).

Goal: stop the planners emitting a `generic-text-block` paired with a
structured component (`faq`, `pricing`) covering the same content — the
plan shape that emptied the gaswholesalers FAQ.

A note that applies throughout: `content_components.function` is
hyphen-only (`chk_function_kebab_case` rejects underscores). So
`call_to_action` is NOT a valid function — canonical is `call-to-action`.
The current prompts use `call_to_action` in examples, which only works
because downstream normalisation is lenient. These fixes also correct the
examples to hyphens so the prompts teach the DB-valid form.

---

## Fix 1 — content-gap-planner prompt

Agent: `content-gap-planner` (id `637b750c-...`). The `new_page` example
JSON in its `prompt_template` hardcodes the bad shape and the
`add_to_page` example uses a non-canonical name.

### 1a. Fix the `new_page` example (removes the hardcoded generic-text-block)

**Before:**
```json
  "new_page": {
    "name": "kebab-case-name",
    "title": "Page Title | Company Name",
    "page_type": "content",
    "purpose": "What this page covers and why",
    "sections": ["hero", "generic-text-block", "call-to-action"],
    "nav_label": "Nav Label",
    "in_header": true,
    "in_footer": true
  },
```

**After:**
```json
  "new_page": {
    "name": "kebab-case-name",
    "title": "Page Title | Company Name",
    "page_type": "content",
    "purpose": "What this page covers and why",
    "sections": ["hero", "<choose content sections by purpose>", "call-to-action"],
    "nav_label": "Nav Label",
    "in_header": true,
    "in_footer": true
  },
```

The placeholder stops the LLM anchoring on `generic-text-block` as a
default. For a FAQ page it will then choose `faq`; for a pricing page,
`pricing`; etc.

### 1b. Fix the `add_to_page` example (canonical name)

**Before:**
```json
  "add_to_page": {
    "page_name": "existing page name",
    "add_sections": ["faq-section", "call-to-action"],
    "content_guidance": "What the new sections should cover"
  },
```

**After:**
```json
  "add_to_page": {
    "page_name": "existing page name",
    "add_sections": ["faq", "call-to-action"],
    "content_guidance": "What the new sections should cover"
  },
```

`faq-section` is not the component function (`faq`). Using the wrong name
in the example teaches the LLM the wrong token.

### 1c. Add the no-pairing rule

Insert into the `## Your Task` section, after the approach descriptions
(A/B/C/D) and before the `Return ONLY valid JSON` line:

```
### Section selection rules
- Structured components such as `faq` and `pricing` are COMPLETE content
  surfaces — they hold their own content. Do NOT also add a
  `generic-text-block` covering the same material on the same page. Use a
  `generic-text-block` only for narrative content that a structured
  component does not already cover (e.g. a short intro that is clearly
  distinct from the FAQ items themselves).
- Use ONLY component function names from the Available Section Components
  list. Use the function name (the value in parentheses), never a display
  name.
```

---

## Fix 2 — site-planner prompt

Agent: `site-planner` (id `f7c8bee1-...`). Two gaps: `faq` is missing from
the standard mappings, and the component list interpolation shows three
names (`name`, `display_name`, `function`) without saying which to use in
`sections` — the source of the `"FAQ Section"` display-name leak.

### 2a. Add faq/pricing to the standard mappings

In the `Use these standard mappings:` list, add:

**Before** (end of the mappings list):
```
   - For about content: use "about-content"
   - For differentiators/why-us: use "differentiators-section"
```

**After:**
```
   - For about content: use "about-content"
   - For differentiators/why-us: use "differentiators-section"
   - For FAQ / question-and-answer content: use "faq"
   - For pricing tables/tiers: use "pricing"
```

### 2b. Clarify which name to use, in the component list instruction

**Before:**
```
## Available Section Components
The following components are available in our component library. You MUST use ONLY these exact component names in the "sections" arrays:

{{range .available_components}}
- {{.name}} ({{.display_name}}): {{.function}} - {{.description}}
{{end}}
```

**After:**
```
## Available Section Components
The following components are available. In the "sections" arrays you MUST
use ONLY the function name — shown in [brackets] below — never the display
name or title.

{{range .available_components}}
- [{{.function}}] {{.display_name}}: {{.description}}
{{end}}
```

Putting the `function` first and in brackets, and dropping the bare
`{{.name}}` lead-in, removes the ambiguity that let the LLM emit
`"FAQ Section"`. The display name and description remain for the LLM to
understand what the component is, but the token it must use is
unmistakable.

### 2c. Add the no-pairing rule + fix the underscore example

In `STRICT RULES`, the home example uses `call_to_action` (underscore).
Change rule 6's neighbours and append a rule:

**Before** (rule 5–6 area and the example array):
```
      "sections": ["hero", "features", "testimonials", "call_to_action"]
...
5. Keep header navigation to 5-8 items maximum
6. Always include: index (home) and contact pages
```

**After:**
```
      "sections": ["hero", "features", "testimonials", "call-to-action"]
...
5. Keep header navigation to 5-8 items maximum
6. Always include: index (home) and contact pages
7. Structured components (faq, pricing, data-driven features) are complete
   content surfaces. Do not pair them with a generic-text-block covering
   the same content on the same page.
```

(Also update the standard-mapping line `For calls to action: use
"call_to_action"` → `use "call-to-action"`, and any other `call_to_action`
occurrences in the prompt, to the hyphen form.)

---

## Fix 3 — applyNewPage default (Go, backstop)

`apply_gap_plan_action.go`, `applyNewPage`. The hardcoded default fires
only when the LLM omits sections, but it is the last line of defence.

**Before:**
```go
	sections := []string{"hero", "generic-text-block", "call-to-action"}
	if sectionsRaw, ok := newPlan["sections"].([]interface{}); ok && len(sectionsRaw) > 0 {
		// ... use LLM-provided sections ...
	}
```

**After:**
```go
	// Archetype-aware default: a recognised page type gets its archetype
	// shape rather than a generic text block that competes with structured
	// content. Unknown types keep the generic default.
	sections := defaultSectionsForPage(pageName, pageType)
	if sectionsRaw, ok := newPlan["sections"].([]interface{}); ok && len(sectionsRaw) > 0 {
		// ... use LLM-provided sections (now also run through the resolver,
		//     per the validate_components implementation doc) ...
	}
```

```go
// defaultSectionsForPage returns archetype-appropriate default sections.
// Falls back to a generic content shape for unrecognised pages.
func defaultSectionsForPage(pageName, pageType string) []string {
	key := strings.ToLower(strings.TrimSpace(pageName))
	switch {
	case key == "faq" || strings.Contains(key, "faq"):
		return []string{"hero", "faq", "call-to-action"}
	case key == "contact":
		return []string{"contact-hero", "contact-form", "contact-info"}
	case key == "pricing" || strings.Contains(key, "pricing"):
		return []string{"hero", "pricing", "faq", "call-to-action"}
	case key == "about":
		return []string{"hero-about", "about-content", "call-to-action"}
	default:
		return []string{"hero", "generic-text-block", "call-to-action"}
	}
}
```

Keep the map small and obvious; it is only a fallback for when the LLM
gives nothing. The prompt fixes (1, 2) are the primary control.

---

## SQL to apply the prompt edits

Prompt edits live in `agent_definitions.default_config -> workflow ->
steps -> <step> -> config -> prompt_template`. They are large strings;
editing in place with `jsonb_set` is error-prone for multi-point edits.
Safer pattern: read the current template, edit the full string in an
editor, write it back as a parameter.

```sql
-- 1. Dump the current template to inspect/edit (content-gap-planner)
SELECT default_config #>> '{workflow,steps,plan_gaps,config,prompt_template}'
FROM agent_definitions
WHERE type = 'content-gap-planner';
```

Edit that text (apply 1a/1b/1c), then write it back. Using psql, set it
from a file to avoid quoting issues:

```sql
-- 2. Write the edited template back (content-gap-planner)
--    Load the edited prompt from a file into :newprompt first, e.g.:
--      \set newprompt `cat gap_planner_prompt_edited.txt`
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,plan_gaps,config,prompt_template}',
      to_jsonb(:'newprompt'::text),
      false
    ),
    updated_at = NOW()
WHERE type = 'content-gap-planner'
RETURNING type, version, updated_at;
```

```sql
-- 3. Same pattern for site-planner (step is plan_site)
SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}'
FROM agent_definitions
WHERE type = 'site-planner';

-- ... edit (2a/2b/2c) into site_planner_prompt_edited.txt ...
-- \set newprompt `cat site_planner_prompt_edited.txt`
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,plan_site,config,prompt_template}',
      to_jsonb(:'newprompt'::text),
      false
    ),
    updated_at = NOW()
WHERE type = 'site-planner'
RETURNING type, version, updated_at;
```

`loadAgentDefinition` reads per-spawn with no cache, so edits take effect
on the next planner run. No chassis rebuild needed for the prompt fixes
(Fix 1, 2). Fix 3 and the `validate_components` implementation are Go and
DO need a chassis rebuild + redeploy.

## Verification

After the prompt edits, run a planner against a test brief that would
previously produce a FAQ page, and confirm:
- the faq page's sections are `["hero","faq","call-to-action"]` (no
  `generic-text-block`),
- no section name is a display name (`"FAQ Section"`),
- `call-to-action` uses the hyphen form.

```sql
-- Inspect a freshly planned site's faq page sections
SELECT jsonb_pretty(section)
FROM site_specs ss, jsonb_array_elements(ss.data #> '{pages}') section
WHERE ss.site_id = '<new-test-site>'
  AND ss.aspect = 'site_plan'
  AND section->>'name' = 'faq';
```

The same isolated-build harness used for the writer test confirms the
end-to-end result: a planned-then-built faq page comes out with a
populated, linked accordion.

## Summary of the full prevention set

| Fix | Type | Defect | Rebuild? |
|---|---|---|---|
| 1. content-gap-planner prompt | prompt (SQL) | duplicate surface | no |
| 2. site-planner prompt | prompt (SQL) | duplicate surface + display-name | no |
| 3. applyNewPage default | Go | duplicate surface (backstop) | yes |
| validate_components impl | Go | display-name leak / orphan | yes |
| (later) per-section briefs | prompt + loader | disambiguate legit pairings | no |
| (later) post-build structured-field check | Go | empty structured component | yes |
