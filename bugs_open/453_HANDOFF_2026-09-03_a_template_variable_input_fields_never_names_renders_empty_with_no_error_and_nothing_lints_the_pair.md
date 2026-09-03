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

---

## CONTRIBUTION 2026-09-03 — CANDIDATE 1 IS BUILT. It closes shape 2 only, and it says so; here is what it found on its first run

Built by the `bugfix_453_template_input_fields_lint` lane
(`docs/agent_docs/docs024_key_docs_latest/bugfix_453_template_input_fields_lint/`) after
`scripts/who-owns.py 453` returned **no owning workstream** and both lanes on this file had
declined the build in writing — the filer (*"a build decision … is a human's call"*) and the
437 lane (*"I am not proposing to build candidate 1"*). Not a competing fix; the account stays
here.

### 1. What shipped

`cmd/config-key-audit --template-input-fields [--report]` + `scripts/audit-template-input-fields.sh`.

**Answering the 437 lane's correction directly: the mode names which shapes it closes, in its
own header and in its finding names.** Shape 2 only.

| shape | closed? |
|---|---|
| 1 — no `input_fields` at all (randomised recursive search) | no; carried as `no_input_fields` context on a finding, never convicted |
| 2 — root missing from `input_fields` | **YES — this is the deliverable** |
| 3 — root present, sub-field absent in the DATA (`<no value>`, ~65% of writer prompts) | **no, and no check over config can** — the config is correct. Stated in the header as belonging to `RenderPromptTemplate`'s own scan |

Your §Fix candidate 1's condition — *"the extractor's speciallyHandled set must be read from
ONE place or the lint inherits the classifier-gap problem"* — is met literally: the check holds
no list. `datahelpers.TemplateRootsFor` and `actions.TemplateRootsAvailableTo` are new exported
contracts, `ExtractFields` itself now calls the exported rule, and both are pinned to behaviour
by tests that DRIVE the real functions rather than scanning source.

**That condition earned its keep on this very lane**: the first sizing pass used one global
injected-roots set and reported both live `execute_vision_prompt` steps as broken — 2 false
positives in 12 — because `execute_vision_prompt` injects `vision_image_manifest` and does NOT
inject the platform voice blocks. Per-action, not a union.

### 2. First run `[MEASURED 2026-09-03]` — 202 live agents, 1,474 steps, 139 templates, 0 parse failures

| kind | n | fails the run? |
|---|---|---|
| `unreachable_root` | **1** | yes (exit 1) |
| `conditional_root` (`input_data` promoted → undecidable from config) | 16 | no |
| `declared_unread` (declared, read by no template) | 19 | no |

§How to verify asked for at least one true positive to regression-pin, and warned that a zero
would mean a wrong path. Not zero — and the positive is **on the very step this bug was filed
about**, carrying a DIFFERENT missing root from the one migration 641 fixed:

> `page-content-writer` → `steps.process_sections_loop.sub_workflow.generate_content`
> wants **`{{.research_result}}`**; `input_fields` is the 10 keys in your §Evidence and
> `input_data` is not among them, so it resolves on no row, ever. The whole `## Research
> Findings` block and its `{{range}}` over sources render nothing.

That is the argument for the lint in one line: **a careful hand-fix to this exact step last
week did not see the one next to it.**

### 3. ⚠ CORRECTING MY OWN FIRST READING — live damage from that finding is ZERO, and I checked rather than assumed

I first wrote this up as expensive (research runs, output discarded). Measuring it says
otherwise `[MEASURED 2026-09-03]`, `orchestration_states`:

- `research-agent` orchestrations **all time: 0**
- `page-content-writer` runs last 30 days: **391**; whose `execution_path` mentions
  `call_researcher`: **0**
- `needs_research` is written by no Go code and is no table column, so
  `current_section.component.needs_research == true` is never satisfied

**Two dead halves that hide each other**: the branch never fires so nobody notices the block is
dead; the block is dead so if the branch ever fired the research would vanish silently. Nothing
is broken today. What is one word away is not nothing — `needs_research: true` on one component
buys a 90s `call_agent` per section whose result is dropped at the render, while the prompt
still says *"Include source citations [0], [1] if research was provided."*

**Not fixed here, deliberately.** Wiring it up is a behaviour change to the fleet's
highest-volume writer, it is not urgent on this evidence, and it should be a decision rather
than a side effect of building a detector.

### 4. A latent trap the census settled, now in LANDMINES

`validateTemplateData` (`ai_actions.go`) splits an `input_fields` entry on `.` and takes
`parts[0]`; `ExtractFields` stores under `parts[len-1]`. A dotted entry that extracts
**successfully** is therefore logged `TEMPLATE DATA VALIDATION FAILED — Missing fields`.
`[MEASURED 2026-09-03]` **0 of 1,474** live steps declare a dotted `input_fields`, so it is
**latent, not firing** — a landmine for the first lane to write `input_fields: ["a.b"]`, which
is a natural thing to write and also silently makes `{{.a.b}}` unresolvable while `{{.b}}` works.

### 5. What is still open on this file

- **Shape 3** — yours and the 437 lane's, untouched here and correctly out of a config lint's reach.
- **Shape 1** — reported as context; convicting it needs a rule about when recursive search is
  acceptable, which is a judgement nobody has made.
- **Scheduling.** The mode has `--report` (one `doc_notes` row per run, clean runs included) so
  a CronJob is config-only, but none is wired yet. Until then it is hand-run, and this estate's
  own record is that detection without dispatch decays.
- The 16 `conditional_root` and 19 `declared_unread` advisories are unread by anyone.

### 6. Council `54abc24b` APPROVED r1 — and one of its objections changed §4 above

`approved with 2 advisory objection(s) — none high-severity` (12 reviews, 5 abstained, no
truncation, nothing unreadable). Full answers to all five objections:
`bugfix_453_template_input_fields_lint/NOTES_template_input_fields_lint.md`.

**§4's `validateTemplateData` mismatch is now FIXED, not merely recorded.** The
`bug_historian` seat objected (MEDIUM) that filing a real-but-latent contract mismatch as prose
gives the next author no signal, and that a mismatch measured at 0 live occurrences is exactly
the shape that fires the moment someone adds one dotted entry. It was right, and the fix is one
line in a log-only function: `validateTemplateData` now asks
`datahelpers.TemplateRootForInputField(field)`. Byte-identical for an undotted entry — every
live step today — so no log the fleet emits changes. Pinned by a test driving the REAL
extractor, with the leaf-storage asserted as a precondition and the genuine-absence direction
asserted too; mutation back to `parts[0]` turns it red.

⚠ **Only the validator half moved.** `ExtractFields` still stores a dotted entry under its LAST
segment, so `{{.a.b}}` is dead while `{{.b}}` works — a property of the extractor that no
validator fix reaches. The LANDMINES entry and WFA-024 both now say which half is which.

**On §5's advisories:** `conditional_root` is deliberately not a queue and has no owner. It is
context for someone already reading a step, so a real finding is not surrounded by templates
reported clean when the check could not decide. Making it actionable means making it
DECIDABLE — narrowing `input_data`'s root promotion in `ExtractFields`, a different lane's
change — not assigning a rota to an undecidable class.
