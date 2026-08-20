# NOTES — `bugfix_305_negation_gate` (append-only, newest at the bottom)

## 2026-08-19/20, session 1 — research, measurement, and three refuted sub-designs

### Ownership and validity, checked before anything else

- `scripts/who-owns.py 305` → `copy_quality_two_stage` (ACTIVE, 81 commits/14d), 16 mentions across
  their handoffs and NOTES. Read `HANDOFF_2026-08-19_continue_here.md`: **no 305 fix in flight**;
  item 6 parks the brief detector's scheduling as an owner/architecture call; item 3 leaves the briefs
  to the site lanes. Writer-side half open.
- `site_work_items` where `item_type='needs_diagnosis'` and status not terminal: 4 open, none on this
  symptom. No open work item mentions negation/contrast copy.
- Bug still valid `[MEASURED 2026-08-19 ~21:30Z]`:
  - brief still supplies the tagline — `SELECT position('in days, not months' in data->>'formatted')>0`
    on `ai-agent-orchestration.com` `content_direction` `is_current` → **true**, 3,558 visible chars,
    row created 2026-07-24 and never updated;
  - the three pages still serve the quoted copy — `page_components.content_data`, all nine components
    `locked_at IS NULL`, `updated_at 2026-08-17` (a **rerender**, per the bug's §3);
  - the writer has not stopped — same site, 2026-08-19 18:26–18:32Z: *"not a catalogue built to look
    busy"*, *"not from provider marketing pages"*, *"not staging load"*.

### ⚠ MISSTEP 1 (mine, caught by a positive control) — `\b` in Postgres regex

My first distribution census reported `not X, but Y` = **0** and `rather than` = **0** across 1,503
writer calls. Both are false. I had pasted Go-shaped patterns (`\bnot …\b`) into psql, and **Postgres
ARE has no `\b` word boundary — there `\b` is a backspace character** (`\y` is the boundary;
`LANDMINES.md:4219`, and `WRONG_CALLS.md:17787` records another session making the identical mistake).
Caught by running the pattern against a sentence I *knew* matched and getting 0.

**Cheap check, now used for every pattern in this lane: assert the regex against a known-positive and
a known-negative string in the same query before quoting any count from it.** Logged in
`WRONG_CALLS.md`.

Corrected figures (`llm_call_log`, `agent_type='page-content-writer'`, `success`, 2026-08-13..19,
**1,503 calls ≈ sections**):

| shape | sections with ≥1 |
|---|---|
| `x_not_y` (`[a-z)"'],\s+(not\|never)\s+…`) | 631 (42%) — ≥2: 208 (14%) |
| `rather than` | 646 (43%) |
| `not X, but Y` (the only shape the Go detector has) | **23 (1.5%)** |
| negative reveal (`. It doesn't …`) | 168 (11%) |
| headline-class JSON field carrying `x_not_y` | 209 (14%) |

### What the existing machinery is and is not

- `platform/orchestration/datahelpers/voicetells.go` — `ScanVoice` is **site opt-in**
  (`ParseVoiceGate` returns nil without `voice_gate.enabled`): **9 of 43 sites**. `strawmanCommaRe`
  (:151) needs a trailing `, but`. So the estate's only wired detector is blind to both sentences the
  owner quoted. `FindStringIndex`, so at most one strawman finding per block.
- Callers: `discovery_checks/check_voice_tells.go` (post-deploy, files `voice_tells` at
  `needs_human_review`, **no handler** — 45 parked, 1 ever closed), `revalidate_voice_tells.go`
  (retractor), `save_page_meta_description_action.go:296-330` (**the only hard gate**, one sentence,
  and its header explains why it reuses the text-level entry point rather than the page-level check),
  `cmd/voicescan`.
- `page-content-writer` is effectively the whole writing population: **1,516 of 1,519** voice-carrier
  LLM calls in the last 7 days (`copy-editor` 3).
- `execute_llm_prompt`: 66 carriers, **no `ActionInputSpec`**, so `scripts/audit-optional-key-budget.sh`
  lists it under *"NOT COUNTED — the optional surface is UNKNOWABLE, not zero"*. A design that added a
  key there would be adding to the widest shared action in the estate, invisibly to the RFC_022 budget.

### The three refuted sub-designs

1. **Whole-section re-ask + "keep the lower score"** → adopts displacement. Neighbour baselines in the
   same corpus: `instead of` 5.9%, `isn't just/a` 6.4%, `more than (just)` 10.8%, `unlike` 0.3%,
   `without the/a` 4.5%, em dash 0.5%.
2. **"Verbatim in the rendered prompt" exemption** → `rather than` is in every rendered writer prompt
   (house voice ×6 + STRICT RULE 19), silently exempting the 43% arm.
3. **Quoting the house-voice rule in the repair prompt** → that rule's text carries the construction
   and a worked example of it.

Cost, measured rather than assumed: mean writer call **11,009 in / 2,126 out** tokens, **0 cached
prompts** (the template has no `<!--CACHE_BREAKPOINT-->`, and caching is opt-in by marker in
`platform/aiservice/anthropic.go`). Whole-section repair ≈ $0.072/call ≈ $200/month at 215
sections/day; the patch shape ≈ $0.0135.

### Decisions taken

- Scanner in `datahelpers` (pure), annotation default-ON in `render_component`, repair in its **own
  action** with its own input spec, page budget in `CollectedData`, migration held until the image is
  live. Full reasoning in `PLAN_2026-08-20_negation_gate.md`.
- Three lanes told before any code: `copy_quality_two_stage`, `site_ai_agent_orchestration`,
  `portfolio_positioning` (CONTRIB files dated 2026-08-20 in each lane's own directory).
