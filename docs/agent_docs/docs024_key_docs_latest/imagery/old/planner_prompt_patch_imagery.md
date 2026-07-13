# Planner Prompt Patch — Imagery Block

Target: `agent_definitions` row for `build-site-planner` agent, specifically the `default_config.workflow.steps.plan_site.config.prompt_template` field, in the "## Imagery Block" section.

Goal: address items 4, 23, and 25 from `TODO_imagery_followups.md` via prompt-level changes only. No Go code changes; no schema changes. Cheaper than functional plumbing, can be hardened later.

---

## Change 1 — Entry fields table

Move the aspect ratio from `constraints` to `style_hints` so the existing Go code path actually honours it (`parseAspectRatio` reads `style_hints.aspect_ratio`).

### Before

```
| style_hints | optional | JSON object, e.g. {"medium": "line drawing", "mood": "collaborative"} |
| constraints | optional | JSON object, e.g. {"aspect": "1:1", "transparent_background": true} |
```

### After

```
| style_hints | optional | JSON object. Use `aspect_ratio` (e.g. "1:1", "16:9", "4:3") to control image proportions. Other keys (`medium`, `mood`, `palette`) are advisory and may be ignored. |
| constraints | optional | JSON object. Currently informational only — does not influence generation. Reserved for future use (anticipated: per-provider validation, content safety modes). |
```

---

## Change 2 — Add explicit decomposition guidance to "What to populate"

After the `sections` paragraph, add a new paragraph:

```
**Each entry produces exactly ONE image.** If a concept requires multiple images (e.g., six icons representing six gripper actuation types, or three illustrations representing three values), emit one entry per image with a distinct `key` for each. The image model interprets a single prompt as a single image — describing multiple images in one prompt produces a multi-panel composition that is unusable. Err toward over-decomposing rather than under-decomposing: a few unused icons are cheaper than one botched multi-panel image.
```

---

## Change 3 — Strengthen per-row prompt construction

Replace the existing bullet list under "### Per-row prompt construction" with:

```
For each entry, write a `prompt` that:
- **Describes ONE image.** Never use "set of", "multiple", "various", "a series of", or counting words ("six", "three") that imply a composition. Each entry is its own image.
- Names the subject concretely (what is in the image)
- Reflects the site's `design_intent.imagery_direction` and `colour_mood`
- For icons specifically: emphasise "single", "flat", "minimal", "line illustration", "plain background" — these style words help the model produce icon-appropriate output rather than photorealistic renders
- Avoids brand markings, logos in the subject (unless the entry IS a logo), and text-on-image unless explicit
- Is self-contained — the image generator sees only this prompt, not the surrounding site context
```

---

## Change 4 — Replace the worked example's section block

Demonstrate (a) multi-entry sections, (b) `style_hints.aspect_ratio` use, (c) icon-friendly prompt construction.

### Before

```json
"sections": {
  "index:2": [
    {
      "key": "icon_precision",
      "kind": "icon",
      "prompt": "Geometric icon representing precision engineering — single colour, sharp corners",
      "constraints": {"aspect": "1:1", "transparent_background": true}
    }
  ]
}
```

### After

```json
"sections": {
  "index:2": [
    {
      "key": "icon_precision",
      "kind": "icon",
      "prompt": "A single minimalist flat icon representing precision engineering — line illustration, geometric, sharp corners, single dark colour on plain background, no shadows, no photorealism",
      "style_hints": {"aspect_ratio": "1:1"}
    },
    {
      "key": "icon_speed",
      "kind": "icon",
      "prompt": "A single minimalist flat icon representing fast cycle speed — line illustration, dynamic geometric form, single dark colour on plain background, no shadows, no photorealism",
      "style_hints": {"aspect_ratio": "1:1"}
    },
    {
      "key": "icon_reliability",
      "kind": "icon",
      "prompt": "A single minimalist flat icon representing process reliability — line illustration, balanced geometric form, single dark colour on plain background, no shadows, no photorealism",
      "style_hints": {"aspect_ratio": "1:1"}
    }
  ]
}
```

Note for the example: three entries, each describing ONE icon. Same `aspect_ratio` hint on each. Distinct `key` per entry. Each prompt explicitly starts with "A single minimalist flat icon" — anchoring the model toward icon style from the first words.

---

## Change 5 — Add a new RULE 16

After existing rule 15, add:

```
16. Each entry in `imagery` produces exactly ONE image. NEVER describe multiple images in a single prompt. If a section conceptually needs N icons (e.g., six gripper types, three pricing tiers), emit N separate entries with distinct `key` values. Phrases like "set of", "multiple", "various", "a series of", or counting words ("six", "three") cause the image model to produce a multi-panel composition that is unusable. Over-decomposing is cheap (a few unused icons); under-decomposing is expensive (manual cleanup of botched output).
```

---

## Application

The prompt is stored in `agent_definitions.default_config` as nested JSONB. Two application paths:

**Option A — jsonb_set migration (cleaner):**

```sql
-- Load the current prompt, apply edits in your editor, write back as a single replacement.
-- Pseudocode:
BEGIN;
SELECT jsonb_pretty(default_config #> '{workflow,steps,plan_site,config,prompt_template}'::text[])
FROM agent_definitions
WHERE type = 'build-site-planner' AND is_active = true;
-- (Save the current value, edit it according to Changes 1-5, then UPDATE)
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,plan_site,config,prompt_template}',
    to_jsonb('<the new prompt text as a single string>'::text)
  )
WHERE type = 'build-site-planner' AND is_active = true;
COMMIT;
```

**Option B — admin UI / DB management tool (faster):** Open the row in your preferred tool, edit the JSONB inline, save. Particularly easier given the multi-paragraph nature of these changes.

After application, the next site that runs `build-site-planner` should produce decomposed imagery entries. Existing sites with already-generated plans are unaffected unless you re-run the planner for them (which you'd want to do for robot-hands.com to get the decomposed icon set).

---

## Verification

After change is applied AND the planner runs once on a fresh or re-planned site, query:

```sql
SELECT spi.scope, spi.scope_ref, spi.key, spi.kind,
       spi.style_hints, spi.constraints,
       LEFT(spi.prompt, 100) AS prompt_preview
FROM site_plan_imagery spi
JOIN site_plans sp ON sp.id = spi.plan_id
WHERE sp.is_current = true
  AND sp.site_id = '<site_id>'
ORDER BY spi.scope, spi.scope_ref, spi.ordering;
```

Expected:
- All `kind='icon'` rows have `style_hints` containing `aspect_ratio`
- `constraints` is NULL or empty for new rows
- No prompt contains "set of", "six", "multiple", or similar (grep the prompts)
- Section-scope groups with multiple icons show as N rows with N distinct keys
