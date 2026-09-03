# NOTES — template-variable ↔ `input_fields` lint (`bugs_open/453` candidate 1)

Append-only, newest at the bottom.

---

## 2026-09-03 — picked the lane up, sized it before building

Arrived from the closed `bugfix_394_render_audit_rotation_cursor` lane. Checked ownership
first (`scripts/who-owns.py 453`): no owning workstream, and both lanes touching the file had
declined the build in writing. Took it.

### Read the mechanism at the code before trusting the bug file's summary

`bugs_open/453` §Mechanism is accurate but not complete enough to build from. Four things the
code says that the file does not:

1. **`ExtractFields` stores a DOTTED input field under its LAST segment, not its first.**
   `unified_extractor.go`: `parts := strings.Split(fieldName, "."); simpleKey := parts[len(parts)-1]`.
   So `input_fields: ["sections_for_render.sections_ready"]` makes `{{.sections_ready}}` work
   and `{{.sections_for_render.sections_ready}}` **not** work. Any lint that assumes
   "root = first segment" is wrong in the opposite direction from the bug.

2. **`ensureCoreFields` injects `domain`, `objective` and `model` unconditionally**, whatever
   `input_fields` says. Three roots that are always available and would otherwise be a
   permanent false-positive floor.

3. **The injected roots are PER ACTION.** `execute_llm_prompt` calls `injectPlatformBlocks`
   (`voice_style`, `build_standard`); `execute_vision_prompt` does **not**, and instead sets
   `templateData["vision_image_manifest"]` itself. Both actions carry `prompt_template` and both
   go through the same `extractDataForAiAgent`.

4. **When `input_data` is among `input_fields`, every key of the runtime `input_data` map is
   promoted to the ROOT** (`for k, v := range existingInputMap { result[k] = v }`). So for those
   steps the root set is genuinely undecidable from config. This is what forces the two-class split.

### MISSTEP — my first sizing model had one global injected-roots set

First run reported 12 findings including `vision_image_manifest` as missing on
`design-critique-agent critique` and `tool-acceptance-agent look`. **Both false**: those are
`execute_vision_prompt` steps and the action injects that exact key three lines above the
render call. Caught by reading `execute_vision_prompt_action.go` rather than by any test —
which is precisely why D2 in the PLAN makes the check read the injected set from the `actions`
package instead of holding its own list. **2 false positives out of 12** on a set small enough
to eyeball; on a fleet that grows, an un-sourced list is the classifier gap `bugs_open/453`
warned about in its own fix candidate.

### Second misstep, same shape — I nearly missed that a THIRD tier renders templates

`getPromptWithPriority` has three tiers. I sized only step-level `config.prompt_template` at
first. `[MEASURED 2026-09-03]` **145** live `execute_llm_prompt` steps, **14** with no
step-level template, **6** of those backed by an agent-level `default_config.prompt_template`.
Those 6 are real render sites and were invisible to the first model.

### The sizing run that justified building it `[MEASURED 2026-09-03]`

Against the live export (202 active non-snapshot agents, `default_config ? 'workflow'`),
1,474 steps walked, 139 of them rendering a prompt template, **0 template parse failures**:

| kind | count |
|---|---|
| `unreachable_root` (can never resolve) | **1** |
| `conditional_root` (`input_data` present — undecidable from config) | 16 |
| `declared_unread` (declared in `input_fields`, no template reads it) | 19 |

### THE TRUE POSITIVE, and it is on the same step `bugs_open/453` was filed about

`page-content-writer` -> `process_sections_loop.sub_workflow.generate_content`, root
**`research_result`**, and `input_data` is NOT in that step's `input_fields`, so it is
unreachable on every row.

It is not a dead template branch. In the SAME sub-workflow:

- `check_needs_research` -> `conditional`, `condition: current_section.component.needs_research == true`,
  `then_step: call_researcher`, `else_step: generate_content`
- `call_researcher` -> `call_agent` to `research-agent`, `timeout_seconds: 90`,
  **`output_field: "research_result"`**, `next_step: generate_content`

So the researcher runs, writes `research_result` into `CollectedData`, hands control to
`generate_content` — and `generate_content`'s `input_fields` does not name it, so
`{{if .research_result}}` is false and the whole `## Research Findings` block, including its
`{{range ... .sources}}` citation list, renders nothing. Meanwhile the same prompt's STRICT
RULE 7 tells the model *"Include source citations [0], [1] if research was provided."*

Note what this is NOT: it is not the `sections_for_render` mismatch that raised 453 (migration
641 fixed that one, and this check confirms it clean). It is a **different** missing root on the
**same step**, which the manual fix did not notice — which is the argument for the lint in one
line.

---

## 2026-09-03 (later) — I OVERSTATED THE TRUE POSITIVE, and the measurement that corrected it

> **CORRECTED 2026-09-03.** The section above says *"So the researcher runs, writes
> `research_result` into `CollectedData`, hands control to `generate_content`"*. **The
> researcher does NOT run.** I inferred that from the config wiring — `check_needs_research`
> branches to `call_researcher`, which has `output_field: research_result` and
> `next_step: generate_content` — and wrote the inference in the same voice as a finding,
> which is the exact failure the `[INFERRED]` marker exists to prevent. What caught it: going
> to measure the damage and finding there was none.

`[MEASURED 2026-09-03]`, live `orchestration_states`:

| | |
|---|---|
| `research-agent` orchestrations, ALL TIME | **0** |
| `page-content-writer` runs, last 30 days | **391** |
| ...whose `execution_path` mentions `call_researcher` | **0** |

Supporting, same date: `needs_research` appears in **no** table column and in **no** Go
writer (one commented-out example in `store_generated_component_action.go`), so
`current_section.component.needs_research == true` is never satisfied and the conditional
takes `else_step: generate_content` every time.

### What is actually true, and why it is a BETTER argument for the lint

**Two dead halves that hide each other.**

- The research branch never fires, so nobody notices the prompt block is dead.
- The prompt block is dead, so if the branch ever DID fire the research would be discarded
  silently — a 90-second `call_agent` per researched section, paid for, with the result
  dropped at the render and no error anywhere.

**Live damage today: ZERO, and the zero is real** — it is measured at the execution path, not
assumed from the absence of complaints. What is one word away is not zero: setting
`needs_research: true` on a single component definition is an ordinary config change, and the
lane that makes it will get silence rather than research, with the prompt still instructing the
model to *"Include source citations [0], [1] if research was provided"*.

That is precisely the class a guard beats a fixture on: **no run of the defective config can
ever reveal the defect**, because the two faults cancel. The same argument
`audit-single-owner-actions.sh` and `audit-loop-sitewide-item-keys.sh` were built on.

### A latent trap the census also settled — honestly, and it could have come out otherwise

`validateTemplateData` (`ai_actions.go`) splits an `input_fields` entry on `.` and takes
`parts[0]`, while `ExtractFields` stores under `parts[len-1]`. So a DOTTED entry that extracts
SUCCESSFULLY is logged as `TEMPLATE DATA VALIDATION FAILED — Missing fields`.

`[MEASURED 2026-09-03]` **0 of 1,474** live steps declare a dotted `input_fields` entry, so
this is **latent, not firing**. Recorded as a landmine rather than a bug: the first lane to
write `input_fields: ["a.b"]` — a natural thing to write — gets a false ERROR log for a
field that arrived fine, and a real absence would look identical.
