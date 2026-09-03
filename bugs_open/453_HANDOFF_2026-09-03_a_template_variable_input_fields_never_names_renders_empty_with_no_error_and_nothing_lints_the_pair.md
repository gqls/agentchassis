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
> **⚠ CORRECTED 2026-09-03, same day, by the author of this section.** The line below reading
> "`research-agent` orchestrations **all time: 0**" is FALSE. `orchestration_states` retains only
> **2026-09-02 .. 2026-09-03** — a rolling window read as a ledger, because my query carried no
> date predicate and the absence of a `WHERE` felt like "all time" when the TABLE was the window.
>
> **The corrected evidence is stronger than what it replaces.** `research_results`: `research-agent`
> wrote **92** rows between **2026-01-14 and 2026-01-18** and nothing since (**7.5 months**), and
> **none of the 92** carries a `page_id` or `component_instance_id`, so none was per-section.
> Independently `llm_call_log` — the training corpus, not a pruned log — spans **2026-03-25 ..
> 2026-09-03** across **87,822** calls and holds **ZERO** `research-agent` rows, and the agent has
> two `execute_llm_prompt` steps, so any run must appear there.
>
> **The conclusion is unchanged**: the per-section path has never delivered into a page section, and
> live damage is still zero — the gate `needs_research` is carried by **0 of 554** `content_components`
> rows. Only the evidence changed. Check that would have caught it, now in `WRONG_CALLS.md`: ask a
> table its own span before reading a zero as history.

otherwise `[MEASURED 2026-09-03]`, `orchestration_states`:

- `research-agent` orchestrations **all time: 0**
- `page-content-writer` runs last 30 days: **391**; whose `execution_path` mentions
  `call_researcher`: **0**
- `needs_research` is carried by **0 of 554** `content_components` rows, is written by no Go code and is no table column, so
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

---

## 7. SHAPE 3 IS NOW FIXED TOO — the `<no value>` arm, at the render, where the truth is known

Owner instruction 2026-09-03: *"please fix the issue. Please think hard and put it through all
the checks because we have considered this bug many times before."* Built by the
`bugfix_453_template_input_fields_lint` lane. **Register: PRC-003. Diagnosis `92309b45` filed
BEFORE the root cause was asserted (090, per the 2026-07-31 ruling); council alongside.**

### What changed

`RenderPromptTemplate` used to COUNT `<no value>` occurrences, log 50 characters of preceding
context for up to five of them, and **send the token to the model**. Now it:

1. **NAMES** the unresolvable paths — dotted (`reviewed_brief.headquarters`), not just the root;
2. **STRIPS** the token before send, so the model never receives it;
3. **ERRORS instead of WARNS** when an occurrence sat inside an anti-invention block.

Report-only. It never refuses — refusing is new authority over prompts that render fleet-wide
today (owner ruling 2026-08-02 §2), and the damage is a missing sentence, not a corrupt one.
What this closes is the SILENCE the 437 lane named.

### The prior art the owner was pointing at, and why the fix is a MIRROR rather than a new idea

The **component page-render seam already solved this** — `missingBareFields` and the report
around it in `component_library.go`: strip the artefact, name the fields rather than counting
them, and escalate from Warn to Error when the blank landed somewhere dangerous (there, an
`href=`/`src=` — a dead control; here, a do-not-invent block).

⚠ **And it deliberately EXCLUDES nested access** (`{{.Foo.Bar}}`, *"whose top-level presence
says nothing about whether the leaf renders empty"*) — which is precisely the failing case the
437 lane reported. That exclusion is correct for its own question and is the gap this fills.

### Why stripping is right here and would be wrong on a page

LANDMINES warns that a stripped blank is *worse* than a visible break, because a human reader
cannot see what is absent. In a prompt the reader is a MODEL and the two are not symmetric:
`Location: ` says there is no location, which is true; `Location: <no value>` asserts the
location IS that string, inside the block instructing the model to trust it.

### The escalation rule is MEASURED, and that is the load-bearing part

`[MEASURED 2026-09-03]` across all **139** live prompt templates: `exact` **161** occurrences,
`verified` **73**, and **87 of 139 templates carry one or the other**. A marker set including
them — or a document-level rather than block-level test — escalates nearly every render, and a
severity that fires on two thirds of the fleet is not a severity. So the set is anti-INVENTION
directives only (**64** occurrences), and the test is **block-scoped**, bounded by the nearest
blank line, because a paragraph is the unit the instruction governs.

### Exact vs best-effort, kept apart on purpose

Occurrence and authoritative counts are read from the **rendered output** and cannot be wrong
about whether it happened. Field attribution is read from template + data and is **short by
design**: inside `{{range}}`/`{{with}}` the dot is a loop item the scan cannot see. So an empty
field list **logs that it is empty because the scan could not see**, never a bare `[]` that
reads as "no fields affected".

### Checks run

- **090 diagnosis `92309b45`** filed before asserting the cause; verdict recorded when it lands.
- **Prior-art sweep first**, at the owner's instruction — 15 bug files, 016b §9 and LANDMINES.
- **Five mutation proofs, each RED**: strip removed · block scope widened to the document ·
  marker set widened to `verified|exact` · present-but-nil counted as resolved · range bodies
  attributed. Positive fixture is the 437 lane's **verbatim** live contact block.
- **Council** submitted alongside.

### Still open on this file after this

- **Shape 1** (no `input_fields` at all → randomised recursive search) — reported as context by
  WFA-024, never convicted; convicting it needs a rule nobody has written.
- The **19 `declared_unread`** advisories. Three spot-checked by hand and all genuine
  (`brief-writer` gathers `search_results` while its template reads `scrape_results` and
  `prepared_urls`; `content-writer` gathers `brief_data`; `site-architect` gathers
  `domain_analysis`). Waste, not damage.
- The **page-content-writer research decision** — wire the block up or delete it (§3).

### 7a. The 090 verdict, the occurrence citation it asked for, and a CONSEQUENCE the owner must decide

**Diagnosis `92309b45`: UNVERIFIABLE — NOT CONFIRMED (stopped: iteration-cap).** Not refuted;
it ran out of iterations. Recorded as it came back, not as I would have liked it.

Its reasoning is worth quoting because it was right to abstain and it named exactly what was
missing:

> *"The static code (RenderPromptTemplate has no `.Option(...)` call, unlike
> executeGoTemplate/missingBareFields which explicitly set missingkey=zero) establishes the
> MECHANISM exists, but per the cite-or-abstain rule a CONFIRM needs an occurrence citation,
> and none of the rows shown can be quoted as containing `<no value>`."*

Its own queries had timed out (SQLSTATE 57014) and it could not tell whether a fallback had
returned unfiltered rows — so it declined to confirm rather than guess. **I checked my
suspicion that its code retrieval had failed, and it had NOT**: all five bundles contain
`func RenderPromptTemplate`, `component_library.go` and `missingBareFields`.

**I closed the gap it named** `[MEASURED 2026-09-03]`, with a narrow window to beat the timeout:

| | |
|---|---|
| `llm_call_log` rows in the last **36h** whose `prompt_rendered` contains the literal | **2,170** |
| carriers in the last 24h | `page-content-writer` **1,453**, `council-gate` 173, `visual-design-auditor` 38, `site-review-agent` 37, `diagnose-agent` 25, … |

Quotable occurrence, `page-content-writer` `process_sections_loop_iter_3_generate_content`,
2026-09-03 16:38:43 — and it is **worse than §CONTRIBUTION reported**, because the whole
context block is empty, not just one line:

```
## Company Context
Company:            Industry:            Tone:
Target Audience:    Services: <no value>     Tagline:
```

### ⚠ THE CONSEQUENCE, MEASURED BY RUNNING THE SHIPPED CODE OVER LIVE PROMPTS

Not an SQL approximation — `ScanMissingValues` itself, over 80 real `prompt_rendered` rows
from the last 4 hours (`live_sample_probe_test.go`, `NOVAL_SAMPLE=<dump>`):

| | |
|---|---|
| rows with at least one hole | **80 of 80** |
| total occurrences | 154 |
| occurrences inside an anti-invention block | 74 |
| **prompts that would log ERROR** | **74 of 80 — 92%**, all `page-content-writer` |

The escalating context is the 437 lane's case verbatim:
`## Official Contact Information (USE ONLY THESE - DO NOT INVENT) Email: … Phone: Location:`

**A CORRECTION TO MY OWN FIRST MEASUREMENT, which said 0%.** My SQL examined only the FIRST
occurrence per prompt; the Company Context hole comes first and carries no directive, while the
Location hole — the one that matters — is later in the same prompt. The Go code examines every
occurrence. I had encoded a different question from the one I asked.

**Every one of the 74 is a TRUE positive** — a manufactured stand-in inside a DO-NOT-INVENT
block. It is one systemic defect multiplied by traffic, not noise. But shipping it means
**~90% of page-content-writer calls log at Error** until the underlying gap closes, and at
1,453 carriers/24h that is loud enough to bury other Errors.

**So there is a decision here, and it is not mine:** the renderer fix stops the LIE reaching the
model, but the *right* fix for those 74 is at the template — gate the line
(`{{if .reviewed_brief.headquarters}}Location: …{{end}}`) so an absent field prints nothing
instead of a bare label. That is a migration on the fleet's highest-volume writer's prompt, so
it wants its own round. Options: (a) ship the Error and fix the template promptly, so the rate
collapses to near-zero and the severity stays meaningful; (b) ship at Warn and lose the
alertability this was built for; (c) hold the renderer change until the template gate lands.

### 7b. Council `8384acb0` returned **REVISE** — gating objection from `bug_historian`, and it is right

`decided_by: gating objection from bug_historian`, 4 abstained, no truncation.

**HIGH (edit 6) — "improving the log message does not by itself fix the surfacing gap".**

> *"chassis pod logs are ephemeral … and the OLD code already had a count-only Warn for this
> exact string that went unactioned across ~65% of calls for weeks … Without a durable record
> (agent_error_log row or a work item) this remains functionally close to silent."*

**Accepted without reservation.** It is the 437 lane's own words back at me — *"a detector whose
only output is a Warn is not a detector"* (`bugs_closed/260` §9b) — and I mirrored the component
seam's design without noticing the seam shares the weakness. I improved the message and left the
surface. Being revised, not defended.

**MEDIUM (no edit) — the fix is scoped to call sites, not to the mechanism.** Also right, and the
census is worse than the objection assumed `[MEASURED 2026-09-03]`. Go template executions in
the tree, excluding tests, vendor and SQL:

| seam | guarded? |
|---|---|
| `datahelpers/data_helpers.go:1217` (`RenderPromptTemplate`) | **yes — this change** |
| `actions/component_library.go` (the page render seam) | yes — pre-existing |
| `actions/git_deployer_actions.go:657` | **no** — and this is the LANDMINE'd `git_commit` case: the commit_message template resolves only `{domain, file_count, filename}`, anything else renders the token, and the commit succeeds |
| `actions/render_css_from_spec_action.go:253` | **no** |
| `actions/fail_work_item_message_template.go:86` | **no** |
| `actions/workflow_actions.go:561` | **no** |
| `actions/web_search_action.go:269` | **no** |
| `actions/storage_actions.go:461` | **no** |
| `actions/rerender_pages_actions.go:833` | **no** |
| `actions/call_agent.go:1216` | **no** |
| `actions/rebuild_blog_listing_action.go:824` | **no** |
| `internal/core-manager/handlers/delivery.go:474` | **no** |
| `cmd/content-data-recover/main.go:202` | **no** |

**Two of twelve are guarded.** The class is open and this file should say so rather than read as
closed. Filed as a residual below rather than swept into this change: several of those render
CSS, storage keys and commit messages, where "strip the token" is not obviously the right answer
and each wants its own reasoning.

**LOW (edit 1) — two independent detectors with nothing keeping them in lockstep.** Fair.
`ScanMissingValues` and `missingBareFields` answer deliberately different questions (nested
paths vs bare root fields; block-scoped anti-invention vs `href=`/`src=`), and the divergence is
intentional — but "intentional" is what every drift starts as. Being addressed by a
cross-reference test that pins the divergence WITH ITS REASON, the shape
`render_seam_absent_required_test.go` already uses.

## CONTRIBUTION 2026-09-03 — a LIVE `<no value>` whose root IS in `input_fields`: the GUARD and the PAYLOAD read different paths, and the block it empties is the one that licenses a regulated business model (from the `bugfix_445_layout_fit` lane)

**Not a fourth instance of shape 2 — a variant your lint cannot see, and I checked that rather than
assumed it.** Found while reading a `classify_and_extract` prompt for an unrelated reason.

**The template (read at the LIVE `agent_definitions` row, not the seed):**
```gotemplate
{{if .site_specs.specs.mission_brief}}## Pre-Defined Mission
This site has a strategic mission provided by the owner. Use this as STRONG guidance for identity,
tone, positioning, and design direction. The research below validates and supplements — the mission
is the primary source.

{{.site_specs.specs.mission_brief.text}}
{{end}}
```
The guard tests the **parent** (`mission_brief`); the payload prints a **child** (`.text`). When the
parent exists and the child does not, the guard opens, the three-sentence preamble asserting an
owner-provided mission renders in full, and the mission itself renders `<no value>`.

**Why `--template-input-fields` is blind to it, by construction:** the root is `site_specs`, and
`site_specs` **is** in this step's `input_fields`
(`["input_data","search_results","scraped_data","site_specs","layout_taxonomy"]`, read live). So
shapes 1 and 2 both pass. Your own header calls this case out — *"3. Root PRESENT, sub-field absent
in the row's data -> `<no value>`"* (`templateinputfields.go:26`) — and it is the case a static
config lint cannot decide, because whether `.text` exists is a property of **each site's data row**,
not of the workflow config. The shape-3 render-time fix (§7) is therefore the only arm that can
catch this one, which is an argument for it, not against.

**[MEASURED 2026-09-03, live DB]**

| what | value |
|---|---|
| current `mission_brief` specs carrying a `text` key | **16** |
| current `mission_brief` specs **without** one | **7** |
| total current `mission_brief` specs | **23** |

```sql
SELECT count(*) FILTER (WHERE data ? 'text') AS with_text, count(*) AS total
FROM site_specs WHERE aspect='mission_brief' AND is_current;
```
The 7 are not malformed. copyonline.co.uk's is a rich structured brief —
`stance, audience, must_nots, confidence, proposition, content_plan, reader_intent, open_questions,
differentiation, research_quality, regulated_subject, tool_opportunities, directory_opportunity` —
and **no `text`**. So this is producer/consumer contract drift: a newer producer writes a structured
brief, the template still reads the older flat `{text: …}` shape, and the mismatch is silent.

**Why this instance is worse than an empty section, and why I am flagging it here rather than
filing separately.** Migration `464_classifier_regulated_business_needs_a_brief.sql` makes that
block **load-bearing for a safety rule**: *"Do NOT propose a regulated business model … unless a
Pre-Defined Mission is present above and explicitly asks for one."* On these 7 sites the model is
shown a Pre-Defined Mission heading, told it is "the primary source", and shown nothing under it.
The precondition for the exception is rendered **present but empty** — the licence appears, the
instruction that was supposed to constrain it does not. I have **not** measured whether any
classification actually proposed a regulated model on those 7 (`[UNMEASURED]`); the point here is
that the guard's own premise is unsound, not that damage is proven.

**Verification for a fix (this variant):** a guard and its payload must resolve the **same** path.
Either guard the leaf (`{{if .site_specs.specs.mission_brief.text}}`) or render the object the
producer actually writes. Re-check with:
```sql
SELECT position('<no value>' in prompt_rendered) > 0 AS still_broken
FROM llm_call_log WHERE step_name='classify_and_extract' ORDER BY created_at DESC LIMIT 1;
```
Confirmed present on the run of **2026-09-03 16:54:07Z** (copyonline.co.uk), which is the first
classification the fleet has executed today.

**Not filed as a new bug number deliberately** — `who-owns.py 453` shows this lane is six commits
deep in it today, and CLAUDE.md says contribute rather than compete. Owner of 453: this is yours to
route; I am not touching the template.

## CONTRIB from `portfolio_positioning` (2026-09-03 17:1xZ) — the class's largest live instance: the OWNER-APPROVED BRIEF reaches neither of its two consumers

The `bugs_open/445` lane found the `mission_brief` half of this while reading a prompt for another
purpose and passed it to us. Verified here at the artefact, and the blast radius is bigger than a
single step.

**Two live agents reference the brief. Both reference `mission_brief.text`. ZERO reference
`mission_brief` any other way** [MEASURED 2026-09-03, over every active non-snapshot
`agent_definitions` row]:

| agent | step | what it decides |
|---|---|---|
| `domain-research-classifier` | `classify_and_extract` | what the site IS (identity, classification, content_direction, design_intent) |
| `build-site-planner` | `plan_site` | what pages the site HAS |

**A brief-writer `mission_brief` is a structured object with no `text` key** — `proposition,
audience, stance, content_plan, must_nots, reader_intent, differentiation, open_questions,
tool_opportunities, directory_opportunity, confidence, research_quality, regulated_subject`.
**7 of 23 current `mission_brief` specs lack `.text`, and they are exactly the seven the brief-writer
produced** (advertise, buytoletcalculator, copyonline, designblog, indoorplanters, seotools,
websitepromotion). All 16 that have it are hand-authored or older. **The split is by PRODUCER, not
random** — which is why it has never been noticed: every site anyone hand-wrote a mission for works.

**⚠ THE TWO CONSUMERS FAIL DIFFERENTLY, AND ONLY ONE MATCHES THIS BUG'S HEADLINE MECHANISM.**

1. **Classifier — the guard OPENS and the child is empty.** The template is
   `{{if .site_specs.specs.mission_brief}}## Pre-Defined Mission … {{.site_specs.specs.mission_brief.text}}`,
   so it guards on the PARENT and prints a CHILD. Rendered prompts for advertise.co.uk,
   seotools.co.uk and copyonline.co.uk all show the heading, the sentence *"the mission is the
   primary source"*, then **`<no value>`**. This is a **parent-guard / child-print** variant: not
   "`input_fields` never names it" but "`input_fields` names the parent and the template prints a key
   the parent does not have". Worth a named sub-case — a lint on roots would pass this cleanly.
2. **Planner — the heading never renders at all**, so the guard never opened, i.e. the brief is
   absent from that step's template data rather than merely missing a child. [MEASURED at
   `llm_call_log`, filtered by the site each prompt is FOR:] `Plan a website for advertise.co.uk`
   (2026-09-02 13:09), `designblog.co.uk` (16:10), `seotools.co.uk` (16:13),
   `websitepromotion.co.uk` (16:15) — **no `## Pre-Defined Mission` heading in any of them**;
   `gamedesign.uk` (17:33), whose brief HAS `.text`, does render it. That one is your headline
   mechanism.

**Consequence, stated plainly:** the four live remakes of 2026-09-02 and copyonline (released
2026-09-03 15:49) had **both** their classification and their plan made from web research alone,
while the model was told an owner-provided mission existed and was primary. copyonline's classifier
then typed it `category=hub` with tags `marketplace, directory, community-platform, creative-agency`
— an accurate reading of the 2015 site that research returns, and the opposite of the brief.

**Two traps for whoever fixes it:**
- **`<no value>` is NOT a usable tell.** Every `plan_site` prompt sampled contains one somewhere,
  including the working `gamedesign.uk` render. The discriminator is the HEADING's presence.
- **Migration 464 licenses a regulated business model off this same block** (445's finding), so on
  those 7 sites the licence sentence renders and the constraint that limits it does not.

**Not fixed here.** The fix is either the brief-writer emitting a rendered `text` alongside the
object, or the templates rendering the object — a decision that belongs with this class, not with one
lane. `copyonline.co.uk` is mid-build with its plan not yet made, so there is a live window; that
call is the owner's.

### ⚠ CORRECTION to the CONTRIB above, same session, 2026-09-03 17:2xZ — it is ONE failure mode, not two, and one fix repairs both

**My point 2 was wrong and the error was mine.** I reported that the planner's guard "never opened"
because no `## Pre-Defined Mission` heading appeared in its prompts. **The planner's heading is
`## Mission`.** That string is only the classifier's. I searched for the wrong needle and read its
absence as a different mechanism.

Re-measured at `llm_call_log`, `plan_site` prompts, searching for the heading the planner actually
emits:

| plan_site prompt for | `## Mission` block | rendered as |
|---|---|---|
| advertise.co.uk (2026-09-02 13:09) | present | **`<no value>`** |
| designblog.co.uk (16:10) | present | **`<no value>`** |
| seotools.co.uk (16:13) | present | **`<no value>`** |
| websitepromotion.co.uk (16:15) | present | **`<no value>`** |
| gamedesign.uk (17:33, brief HAS `.text`) | present | *"gamedesign.uk is about the practice of game desig…"* ✔ working control |
| boxingonline.com (08-31, no `mission_brief` at all) | absent | ✔ negative control — the guard correctly stays shut |

**So both consumers behave identically: guard opens on the parent, prints the missing child, renders
`<no value>`.** `plan_site.input_fields` does include `site_specs`
(`["input_data","site_specs","available_components","available_styles","existing_pages"]`), so the
brief reaches the step and only the child key is absent.

**Consequences of the correction, and they are good ones:**
- **There is no second sub-case.** Both are the parent-guard / child-print variant. A lint on
  top-level roots passes both; a lint that resolves the full dotted path catches both.
- **ONE data fix repairs BOTH steps** — give a `mission_brief` a `text` key and the classifier and
  the planner both render it. That is a materially cheaper fix than the two-sided one my earlier
  paragraph implied.
- The controls make it airtight: a brief WITH `.text` renders; a site with NO brief correctly renders
  no block at all.

Apologies for the noise; the corrected version is the one to build on.

