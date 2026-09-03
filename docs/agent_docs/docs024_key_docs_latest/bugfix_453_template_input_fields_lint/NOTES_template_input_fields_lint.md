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

> **⚠ CORRECTED 2026-09-03, same day, by the author.** This section said
> `research-agent` orchestrations **ALL TIME = 0**. That is FALSE, and the error was
> reading a rolling window as a ledger: `orchestration_states` retains only
> **2026-09-02 .. 2026-09-03**. My query carried no date predicate, so the absence of a
> `WHERE` read as "all time" when the TABLE was the window.
>
> **The corrected evidence is stronger than the claim it replaces.** `research_results`
> shows `research-agent` wrote **92** rows between **2026-01-14 and 2026-01-18** and
> nothing since — **7.5 months** — and **none of the 92** carries a `page_id` or a
> `component_instance_id`, so none was per-section. Independently, `llm_call_log` spans
> **2026-03-25 .. 2026-09-03** across **87,822** calls, is the training corpus rather
> than a pruned log, and contains **ZERO** `research-agent` rows; the agent has two
> `execute_llm_prompt` steps, so any run must appear there.
>
> **The conclusion survives unchanged** — the per-section research path has never
> delivered into a page section, and live damage from the `research_result` finding is
> still zero (its gate, `needs_research`, is carried by **0 of 554** `content_components`).
> What did not survive is the evidence I first gave for it. `WRONG_CALLS.md` carries the
> check that would have caught it: ask a table its own span before reading a zero as history.

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

---

## 2026-09-03 — shipped, submitted, and the three things left open

**Commit `4aaf64aee`** — 20 files, commit-scope block confirmed no passengers from other lanes.
**Council `54abc24b-852d-4b70-9d86-d021e45a5d5a`**, round 1, submitted after the commit
(`Council-Submitted:`, not `Council-Reviewed:` — nothing has been read yet).
**Register: WFA-024.** **HEAD builds** — `verify-head-builds.sh` OK at `15cda49ba`.

### Mutation proofs — six RED, one recorded VACUOUS rather than counted

Applied to the shipped tree, one at a time, each restored after:

| # | mutation | result |
|---|---|---|
| 1 | `RangeNode` body walked with the dot still root-scoped | RED |
| 2 | `TemplateRootForInputField` returns `parts[0]` | RED |
| 3 | `vision_image_manifest` dropped from the declared contract | RED |
| 4 | `"model"` dropped from `AlwaysEnsuredTemplateRoots` | RED |
| 5 | `requireAgentPromptProjection` always returns nil | RED |
| 6 | the `input_data` promotion flag never fires | RED (both packages) |
| 7 | `kind = kindConditional` made unreachable | **BUILD BROKE — vacuous, NOT a pass** |

⚠ **A misstep worth the line: my backup script collided on two files with the SAME basename**
(`template_context_contract.go` exists in both `datahelpers` and `actions`), so
`cp $f $SP/backup/$(basename $f)` kept only the second. Restoring after mutation 2 put the
`actions` file into the `datahelpers` path — which is what M3's "BUILD BROKE" was actually
reporting, not the mutation. Caught because the build failed rather than the test; had the two
files been compatible it would have restored silently wrong. **Back up by PATH, not basename.**

### Two things deliberately not done, both recorded on `bugs_open/453` §5

- **The `page-content-writer` config is unchanged.** Adding `research_result` to its
  `input_fields` is a behaviour change to the fleet's highest-volume writer; damage is
  measured at zero today; it should be a decision, not a side effect of shipping a detector.
- **No CronJob.** `--report` is wired (one `doc_notes` row per run, clean runs included), so
  scheduling is config-only. Left visible as an open item rather than half-wired — this
  estate's own record is that detection without dispatch decays.

### Not ours, flagged so nobody inherits it as ours

`go test ./cmd/config-key-audit/` fails `TestShippedRegistryIsSelfConsistent` at committed
HEAD: `TOOL_PAGE_HELD_NO_TOOL_SOURCE`'s `why` does not name the retention window its own
disposition rule requires (`bugfix_450`'s entry). Proven not-mine on a clean extract of
`cc572ea14` via `scripts/verify-head-builds.sh --test`, with none of this lane's files present.
Scope your `-run` or you will chase it.

---

## 2026-09-03 — council `54abc24b` **APPROVED r1**, 2 objecting seats / 5 advisory objections, all answered

`approved with 2 advisory objection(s) — none high-severity`; 12 reviews, 5 abstained,
`gated_by_truncation: false`, `unreadable: null`. Read before any trailer was written.

**Two of the five changed the code.** Recording all five with what was actually done, because
an objection answered only in prose is the thing this lane's own finding is about.

### bug_historian, MEDIUM (edit 2) — ACCEPTED AND FIXED, not filed

> *"the plan surfaces a live contract mismatch … and explicitly chooses to file it only as
> prose … a generic latent mismatch measured at 0 live occurrences today is precisely the kind
> that fires silently once a future author adds one dotted `input_fields` entry"*

Right, and the "0 live occurrences" framing was mine, so the objection is against my own
reasoning rather than against the code. **Fixed rather than filed**, because the fix is one
line in a function that only logs: `validateTemplateData` now asks
`datahelpers.TemplateRootForInputField(field)` instead of `strings.Split(field, ".")[0]`.

- **Byte-identical for an undotted entry** (leaf == whole), which is every live step today —
  so no log the fleet emits changes, and the roll that carries it changes no behaviour.
- Pinned by `validate_template_data_test.go`, which drives the **real** `ExtractFields` rather
  than a hand-built map (a fixture agreeing with the old bug would have passed), asserts the
  extractor's leaf-storage as an explicit **precondition**, and asserts the other direction —
  a genuine absence must still be reported, or the fix could have been achieved by blinding
  the check.
- **Mutation M8: reverting to `parts[0]` turns it RED.**
- ⚠ **the landmine's FIRST half is untouched and still bites**: `ExtractFields` stores under the
  LAST segment, so `{{.a.b}}` is dead while `{{.b}}` works. That is the extractor's property
  and no validator fix reaches it. LANDMINES and WFA-024 both updated to say exactly which
  half moved.

### guardian, MEDIUM (edit 2) — `ExtractFields`' 23 call sites deserve a look at the DIFF, not the intent

Fair, and the honest answer is that the diff is small enough to read in full. `git diff` on the
two extractor files is **11 added / 11 removed**, and every changed line is one of three
mechanical substitutions:

| before | after |
|---|---|
| `speciallyHandled := map[string]bool{…}` (inline literal) then `speciallyHandled[fieldName]` | `IsSpeciallyHandledInputField(fieldName)` over the same map, moved to package level |
| `parts := strings.Split(fieldName, "."); simpleKey := parts[len(parts)-1]` | `simpleKey := TemplateRootForInputField(fieldName)` — the same expression, named |
| `funcMap := template.FuncMap{…}` (inline literal) | `PromptTemplateFuncs()` returning the same four entries |

No call site changed, no signature changed, no key-selection rule changed. All pre-existing
`datahelpers` and `actions` tests pass unchanged, and `TestDottedInputFieldIsStoredUnderItsLeaf`
now pins the key-selection behaviour that had no test at all before this lane.

### guardian, LOW (edit 4) — confirm the vision constant ships in the SAME build as its contract

**Confirmed, and now concretely:** both are in commit `4aaf64aee`, and the owner has a fresh
chassis building today which takes committed HEAD. There is no window in which the constant and
its declaration are in different images.

### guardian, LOW (edit 5) — the plan listed one new file twice (`add` then `modify`)

Correct; that was an authoring artifact of mine, not two files. `templateinputfields.go` is a
single new file. Worth noting because the seat is right that it means reviewers saw two sketch
fragments rather than one whole file — a resubmission would consolidate.

### bug_historian, LOW (edit 5) — nobody triages the 16 `conditional_root` findings

Accepted as a real gap, and it is the same gap as "not scheduled" rather than a second one.
**Stated disposition, now written into the RUNBOOK and `bugs_open/453` §5 rather than left
implied:** `conditional_root` is **context read at the moment you are already looking at a
step**, not a queue anybody owns. It exists so that a reader chasing a real finding is not
told a neighbouring template is clean when the check simply cannot decide. If it ever needs an
owner, the honest move is to make it decidable — which means narrowing `input_data` promotion,
a change to `ExtractFields`' contract and a different lane's work — not to open a triage rota
over an undecidable class. `bugs_open/083`/`077`, which the seat cites, are exactly what an
unowned queue becomes.

---

## 2026-09-03 — a PROCESS misstep of my own: the reviewed commit carries no trailer

`4aaf64aee` is this lane's in-scope platform commit — the 11 files the gate actually reviewed —
and it carries **neither** trailer, so `098` will list it as unreviewed for ever. The review
happened and was approved; the join cannot see it.

**Cause: I committed first and submitted second.** `Council-Submitted:` names a correlation, and
at commit time no correlation existed yet, so there was nothing to write. The two later commits
(`6c60c3bc2` `Council-Submitted:`, `71c7ce40b` `Council-Reviewed:`) carry it, but they are the
NOTES and the follow-up fix — the small ones.

**The cheap check that would have prevented it: SUBMIT BEFORE YOU COMMIT.** The gate needs only
a rationale and a plan, never a commit, so the correlation can always exist first. CLAUDE.md's
`Council-Submitted:` guidance assumes you have a correlation in hand and does not say to get one
first — which is exactly the order I got wrong. Forward-only forbids an amend, so this cannot be
repaired; it is recorded so the next lane orders it the other way round.
