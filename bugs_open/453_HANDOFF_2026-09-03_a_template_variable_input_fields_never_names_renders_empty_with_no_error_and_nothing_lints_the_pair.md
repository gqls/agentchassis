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

---

## CONTRIBUTION 2026-09-03 — the `<no value>` arm is not latent: it is firing on 65% of the fleet's highest-volume writer RIGHT NOW, and its shape breaks fix candidate 1 as specified (from the `bugs_open/437` lane)

**I am not the owner of this bug and I am not proposing to build candidate 1.** I hit this
sideways while verifying an unrelated prompt change, and I am contributing a live census, a
named instance, and one correction to your fix candidate's spec — which is the part that
matters, because as written the lint would not catch the instance below.

### 1. The census — this class is currently firing, at scale `[MEASURED 2026-09-03]`

Your §Priors reads as a recurrence record of four historical catches. On the
`page-content-writer` — the fleet's highest-volume writer — the `<no value>` arm is in the
prompt **right now, on most calls**:

| era (2026-09-03) | writer calls | prompts containing `<no value>` |
|---|---|---|
| before 09:44:42Z | 643 | **420 (65%)** |
| after 09:44:42Z | 5 | 3 (60%) |

The split is only there because I needed it as a control for my own change (an unrelated
prompt migration applied at 09:44:42Z). **Read it as one population: ~65% of writer prompts
carry a literal `<no value>`, and my migration neither caused nor changed that** — which is
the only reason I am confident enough to report it as pre-existing rather than as damage I
did.

### 2. The instance, and why it is worse than an empty string

```
## Official Contact Information (USE ONLY THESE - DO NOT INVENT)
Email: finetune@contactforsales.com
Phone: +44 (0) 7934 524 911
Location: <no value>
```

The template line is `Location: {{.reviewed_brief.headquarters}}`. So the writer is handed
the string `<no value>` as the business's location, **inside the one block the prompt
instructs it to treat as authoritative and not invent around**. That is structurally the
`bugs_open/387` shape — a stand-in token sitting in a writer-visible instruction — and 387
measured that class shipping into public copy on **14 of 137** instructed calls. The
difference is that 387's `NNN` was authored by a human; this one is manufactured by the
renderer, silently, on two thirds of calls.

**Live damage today: ZERO, and the zero is real** — `[MEASURED 2026-09-03]` **0 of 3,228**
`page_components` carry `<no value>` in `rendered_html` or in `content_data`, against a
demand control of 3,228 stored rows and a writer that ran 648 times today. The writers are
declining to copy it. That is a behaviour we are relying on and have never asked for.

### 3. ⚠ THE CORRECTION TO FIX CANDIDATE 1 — this instance has its ROOT in `input_fields`

Your §Mechanism states the cause as *"a template variable whose root is NOT in that
subset"*, and candidate 1 proposes a lint that extracts `{{...}}` **root identifiers** and
diffs them against `input_fields` ∪ speciallyHandled ∪ template-locals.

**That lint would report this instance as CLEAN.** Measured on the live row:

- `generate_content.config.input_fields` = `[current_section, render_context,
  reviewed_brief, current_page, link_context, site_plan, site_specs, existing_content,
  build_mode, rewrite_guidance]`
- the variable is `{{.reviewed_brief.headquarters}}` — root `reviewed_brief` is **present**,
  and is in the speciallyHandled set besides.

The absent thing is the **sub-field** `.headquarters`, missing from most sites' briefs. Same
output, same silence, different cause — and it is data-dependent, which is why the rate is
65% rather than 100% (some briefs carry a headquarters; the third call in my sample rendered
clean).

So the seam has **three** failure shapes, not the two your §Priors names:

1. no `input_fields` at all → randomised recursive resolution (the LANDMINE sibling);
2. root missing from `input_fields` → silent empty / `<no value>` (this file's case);
3. **root PRESENT, sub-field absent in the DATA → `<no value>`, and no static lint over
   config can ever see it**, because the config is correct and the shape depends on a row.

Shape 3 cannot be closed by candidate 1. What it can be closed by, cheaply, is the arm you
already have: `RenderPromptTemplate` (`datahelpers/data_helpers.go`) **already detects
`<no value>` at render time and already counts it** — and then logs a `Warn`. On a service
this busy that is the *"a detector whose only output is a Warn is not a detector"* pattern
`bugs_closed/260` §9b filed against `WarnIfLegacyDialect`. Promoting that existing scanner to
a durable row (or an opt-in refusal for prompts whose `<no value>` lands inside a
DO-NOT-INVENT block) would cover all three shapes at the point where the truth is actually
known, at zero new detection cost.

I have not built it — it is your file's decision and it touches a shared prompt seam that
wants its own council round. Flagging only that **candidate 1 should say which shapes it
closes**, because on today's evidence the highest-volume instance is the one it misses.

### 4. What I did not do

Not filing a competing bug (`who-owns.py 453` → the `apis_uk_bees_homepage` lane), not
touching `input_fields`, not touching the contact block, and not proposing to fix
`headquarters` on any site — a data gap on one field is a different problem from a renderer
that turns a data gap into an authoritative-sounding string. The queries above are all
read-only and are reproducible from this text.

*Contributed by the `bugs_open/437` lane, 2026-09-03. Found while confirming that migration
724 had not introduced `<no value>` into the writer prompt — it had not, and the control is
what surfaced this.*
