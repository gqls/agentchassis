# 453 — a template variable that a step's `input_fields` never names renders EMPTY with no error, and nothing lints the pair

**Filed:** 2026-09-03, apis_uk_bees_homepage lane, from the `bug_historian` seat's MEDIUM advisory
on council corr `6c92d154` (seed 641). **Class report, not a live-damage report** — the instance
that raised it is already fixed-in-seed with a fail-loud verify; this file exists because the
CLASS has now recurred at least four times and each catch was luck-shaped.

## Mechanism (established, not hypothesised)

`ExecuteLLMPromptAction` renders a step's `prompt_template` against
`ExtractFields(CollectedData, input_fields)` — a SUBSET holding only the keys the step's
`input_fields` array names (plus the speciallyHandled set: `input_data` promotion,
`current_page`, `current_section`, `reviewed_brief`, `site_record` —
`platform/orchestration/datahelpers/unified_extractor.go`). A template variable whose root is
NOT in that subset is simply absent at render: under Go templates a guarded/ranged absent key
renders NOTHING and an unguarded one renders `<no value>` — the ranged form is the killer,
because the output is a plausible page with a silently empty list. No error, no log line, no
verdict.

## Evidence for THIS recurrence (2026-09-02/03, seed 641)

- Fixture D of `finetuning_uk_service/render_test_641/`: the owner-approved block rendered
  against the writer's LIVE `input_fields` produced "also covers, each in its own section:"
  followed by NOTHING — no error (`OUTPUT.txt`).
- Live row measured 2026-09-03: `generate_content.config.input_fields` = 10 keys, no
  `sections_for_render`, while `iterate_over` = `sections_for_render.sections_ready` — the
  loop's own source was invisible to the loop body's template.
- The fix's verify RAISEs unless BOTH template and `input_fields` carry the key, and the check
  was induced-failure-proven (strip the append → RAISE), under BEGIN/ROLLBACK on the live row.

## Priors (the recurrence record — all found by fixture/manual testing, none by a guard)

- `bugs_closed/085` (slug `render_data_advertises_current_page_and_always_supplies_empty`)
- `bugs_closed/039` (slug `section_naming_a_missing_component_renders_an_empty_stub`)
- `bugs_closed/054` (slug `unguarded_range_items_in_list_templates_no_empty_state`) — resolve
  all three by SLUG, numbers collide on this estate.
- LANDMINES "A step with NO `input_fields` resolves its inputs by RANDOMISED recursive search"
  (~line 7777) is the SIBLING arm of the same seam: no `input_fields` at all → random sibling
  wins; a key missing FROM `input_fields` → silent empty. Two failure shapes, one seam, zero lints.

## Fix candidates, ordered by what closes the door

1. **A lint that cross-references every step's template variables against its `input_fields`**
   (the bug_historian ask): for each `agent_definitions` step carrying `prompt_template`,
   extract the `{{...}}` root identifiers and diff against `input_fields` ∪ speciallyHandled ∪
   template-local range/with variables. Report a variable with no source, and (cheaper still)
   an `input_fields` entry no template reads. Natural homes: `cmd/config-key-audit` (which
   already sweeps `agent_definitions` config) or a `scripts/check-*` advisory; the extractor's
   speciallyHandled set must be read from ONE place or the lint inherits the classifier-gap
   problem. Runs against live rows, so it also catches config-only drift no commit hook sees.
2. Per-migration fail-loud verifies (what 641 does) — right for each instance, cannot see the
   next one.
3. Documentation only — rejected: this file would be the fourth document, and the fourth catch
   still came from a fixture someone happened to write.

## Why no 090 run (owner ruling 2026-07-31 — stated substitution)

No new root-cause assertion is being made: the mechanism is directly evidenced in this file
(fixture render + live-row query + induced failure), was independently named by a council seat,
and matches three closed cases. What is OPEN is a build decision (candidate 1), which is a
human's call, not a diagnosis.

## How to verify a fix

Run the lint against the estate as of today expecting AT LEAST one true positive it can
regression-pin (pre-641-apply, the writer's `sections_for_render` mismatch IS that positive;
post-apply, induce one on a scratch row). A lint that reports zero on first run against 50+
agents should be suspected of reading the wrong path (LANDMINES ~563: a wrong jsonb path is
uniformly NULL and reads as a clean fleet).
