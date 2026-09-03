# PLAN — bug 449: no generated fence ever asserts a number

**Started 2026-09-03.** Design, phasing, decisions **and their reasons**. Corrections to the
originating brief live here, marked as corrections — never silently edited away.

Bug: `bugs_open/449_HANDOFF_2026-09-02_no_fence_the_tool_generator_writes_ever_asserts_a_number_so_a_calculator_that_computes_garbage_passes_acceptance.md`
Evidence and queries: `NOTES_449_fences_assert_no_number.md`, `RUNBOOK_449_fences_assert_no_number.md`

> **Note on authorship (2026-09-03).** The owner asked for this plan to be prepared *using
> fable*. The fable run was dispatched with a full briefing and **terminated on a session
> rate limit** (`claude-fable-5-1`, HTTP 429, resets 16:10 Europe/London) having produced
> nothing but its opening line. This plan is therefore written by the Opus session that did
> the research. **It has not had fable's second opinion**, and that is the one input the
> owner asked for that this document does not contain. Re-running it after 16:10 as an
> adversarial read of what follows is cheap and worth doing; if fable's plan differs
> materially, the difference is recorded here as a correction rather than by replacing this.

---

## §0 The problem in one paragraph

The platform builds a calculator and writes it a fence — the list of things that must be
true for it to count as working. The fence is written by `tool-generator`. Every fence it
writes checks the tool's *health*: the page loads, the console is clean, it fits a phone,
and something appears when you click. **None checks whether the number that appeared is
right.** A calculator that confidently prints a wrong figure passes all of them and its
record reads PASSED.

## §1 What is actually true, measured today

**[MEASURED 2026-09-03, live `doc_plans`.]** All figures re-derived first-hand; the bug's
own §2 was measured 2026-09-02 and has already moved.

| author | fences | assert no value at all | uses `computed_values` | drives inputs | drives inputs AND asserts nothing |
|---|---|---|---|---|---|
| `tool-generator` | **186** (was 170) | **115** (was 107) | **0** | 91 | **55** |
| `operator:bugfix224-session` | 16 | 0 | 16 | 16 | 0 |
| `webdesign_couk_thread` | 14 | 4 | 0 | 6 | 0 |
| `operator:mortgagecalculator-lane-a4` | 8 | 0 | 8 | 8 | 0 |
| `operator:staged_component_build` | 8 | 0 | 6 | 7 | 0 |

`max(created_at)` for `tool-generator` is **today**. **This is a live intake, not a backlog.**

**Cause**, read out of the live prompt rather than inferred:
`agent_definitions.default_config → workflow.steps.compose_plan.config.prompt_template`
enumerates a **closed** vocabulary and ends *"No other check type exists for interactions."*
`computed_values` appears in neither fence-authoring agent. The type is never a candidate.

**Seam**, enumerated rather than assumed: `write_doc_plan_action.go` is the only production
Go writer of a PLAN body; exactly three live agents reach it (`tool-generator`,
`experience-planner`, `experience-register-writer`). Operator scripts write `doc_plans` over
`psql` and **bypass it** — so a guard there governs every *generated* fence and no
hand-installed one.

## §2 The decision that shapes everything: candidate 1 is not safe as written

The bug's own fix candidate 1 is "teach both authoring agents the type". **I am not shipping
that alone, and the bug half-says so itself.**

`computed_values` is a **regression** check by construction, not a correctness check. Its own
docstring: the values *"are CAPTURED from the tool while it is known good … and then
defended"*, and *"a golden captured from an already-wrong tool pins the wrong answer"*. At
generation time there is no known-good state, and the generator is handed
`{{.generated_html}}` — the tool's own code — so any expectation it derives **shares a
failure mode with the implementation it is meant to police**. This estate has shipped that
mistake twice (`bugs_open/224`, `225`: an expired £625k FTB SDLT cap certified green for
sixteen months).

**So the design question is not "how do we get a number into the fence" but "where is a
number allowed to come from, and can the platform tell the difference".**

### The consequence for ordering — and it inverts the obvious order

The obvious plan is: teach the generator (cheap, live-on-apply, closes the intake), then
tidy up. **That is the wrong order**, for two reasons that only appear once you look:

1. It would convert *silent blindness* into *confidently defended garbage* — strictly worse,
   because a wrong pinned value is believed, whereas an absent one is merely uninformative.
2. `bugs_open/441` — the mcalc lane's, re-framed by them today as a **live generator of
   stale fences** — means fence selectors are already going stale. A value assertion on a
   stale selector fails, and `runComputedValues` fails rather than skips on a missing
   element **by design**. So teaching value assertions while 441 is live converts silent
   blindness into loud FALSE failures, which then aim `tool-improver` at the wrong thing.

**Therefore the honest-record phases ship FIRST and the authoring phase ships LAST.**

## §3 The phases, ranked by what closes the door hardest

Ranked by the standing rule: *what makes the bad state unrepresentable*, not what is easiest.
"An author must remember to add a check" is the weakest possible control and it goes last.

---

### P1 — The verdict stops overclaiming. (Code. Highest value, no author change, no backfill.)

**The bad state P1 removes:** a PASS that reads like "the calculator works" when nothing
about the arithmetic was asserted. That is the *damage*; the weak fence is only the cause.

**What changes.** The acceptance result records **what the fence actually asserted**, not
only whether it passed. Concretely: a `value_assertions` count (the number of checks that
compare a value — `computed_values` with non-empty `expect_values`, plus `interaction` with
a non-empty `expect.text_matches`) and, when that count is zero, the pass is reported as
**`passed_liveness_only`** rather than a bare `all_passed: true`.

- `platform/orchestration/actions/tool_acceptance_actions.go` — the Tier-4 pass result
  (~:810 `"all_passed": true, "passed": …, "skipped_checks": …`).
- `platform/orchestration/actions/discovery_checks/check_tool_acceptance.go` — the Tier-2
  finding, which already carries `passed`/`failed`/`skipped`.

**Why here and not elsewhere.** This is the only placement that fixes **all 186 fences
at once, today, with no backfill and no author cooperation**, and it cannot regress: the
count is derived from the fence at run time, so a fence that gets weaker gets a weaker
verdict automatically. It also extends a discipline the file already holds rather than
inventing one — `check_tool_acceptance.go`'s own header already says *"passes → a finding
only (a Tier-2 pass must never be read as 'the tool works')"*. P1 is that sentence, made
machine-readable, one tier up.

**Rejected alternative:** a doc_note per weak PASS. Rejected because a note is a side-channel
— the thing people quote is the verdict, so the verdict is where the qualification must live.

**Live on apply?** No. Go — needs an image and a roll.
**Council scope?** Yes (`platform/`).

**Red-induction.** Take `tool-overpayment-priority` (PASSING today on a value-less fence):
its result must come back `value_assertions: 0` / `passed_liveness_only`. **Demand control
in the same run:** a tool with a `computed_values` fence (`simple`, or any of
`operator:bugfix224-session`'s 16) must come back **non-zero** — if both read zero the
counter is blind and the zero grades nothing.

**What P1 deliberately does NOT do.** It does not make any fence stronger, does not fail
anything that passes today, and does not tell anyone to write a better fence. It makes the
record honest. That is all, and it is the part that stops the estate acting on a false PASS
while the rest is built.

---

### P2 — The gap is recorded where it is created. (Code, at the one door.)

**The bad state P2 removes:** a fence is born blind and nothing anywhere says so, so the only
way to know is to run the census by hand.

**What changes.** Extend the existing `subjectType == "tool"` block in
`write_doc_plan_action.go` — the one added 2026-08-24 for `bugs_open/288` defect A, which
already extracts the criteria block and refuses an unreadable `facts` declaration. A fence
that **drives inputs** (any step with action `fill` or `select`) and asserts **no value
anywhere** writes a `doc_note` categorised `fence_asserts_no_value`, naming the subject key
and the authoring agent.

**Record, not refuse — and this is a considered choice, not timidity.** Refusing would
strand the tool: a tool with *no* PLAN is inert at **both** tiers (Tier 2 writes
`needs_criteria` and stops), which is strictly worse than a weak fence. Refusal becomes
correct only once P4 has made the rule satisfiable, and is listed as P5 below.

**The trigger is read off the fence itself, never from a classifier.** "Does this tool
compute?" is a judgement; "does this fence fill in inputs and then check nothing" is a fact
in the document. A guarantee conditional on a classifier inherits the classifier's gaps.
This is why the census carries the `drives inputs` column: **55** fences meet it today.

**One helper, shared with P1**, so the two cannot disagree on what "asserts a value" means.
That is the precedent this very file sets with `criteriaFactsFromValue` ("it exists so the
two cannot disagree on what a well-formed declaration is") — this would be its second
application, not a third spelling of the rule.

**Adds ZERO optional keys** to `write_doc_plan`'s input spec. That is deliberate:
`audit-optional-key-budget.sh` puts `write_doc_plan` at **8 optional keys / 3 carriers**
against the RFC_022 budget of 10, so a design that spent two would land the action *at* the
budget and owe a review of its accumulated surface. The rule needs no configuration.

**Live on apply?** No. Go — needs a roll.
**Council scope?** Yes (`platform/`).

**Red-induction.** Write a PLAN whose fence fills an input and asserts nothing → the note
appears. **Demand control:** the same PLAN with one `expect.text_matches` added → **no**
note. Both arms required; a rule that only ever fires proves as little as one that never does.

---

### P3 — A standing report, so nobody has to do archaeology. (Script + cron.)

**What changes.** `scripts/audit-fence-value-assertions.sh`, following the established
`audit-*` pattern, reporting per author: total fences, drives-inputs, asserts-nothing, and
— **the number that matters** — the same figures **restricted to a `created_at` window**, so
the standing stock of 115 cannot mask whether the intake has stopped. Nightly CronJob.

**Why the window is the whole design.** The bug's §6 is right: *"Compare by `created_at`
window, not by total — the 115 existing ones do not change themselves."* A fix that works
shows up as a fall in the **new** rows and leaves the total nearly flat for weeks. A report
that prints only totals would read as "no improvement" for a month after a successful fix
and would be quietly abandoned.

**Live on apply?** The script, yes. The CronJob needs a kustomize apply — and per the
standing landmine, **read the artefact** (`kubectl get cronjob … -o jsonpath=…image`), never
the make target.
**Council scope?** No (`scripts/`, and not `pattern-check.py`).

**Red-induction.** Run it today: it must report **115 / 55** for `tool-generator` and **0**
for `operator:mortgagecalculator-lane-a4`. A report that returns zero for everyone is blind,
and the operator lanes are the demand control that proves it is not.

---

### P4 — Teach both authoring agents, WITH a provenance rule. (Migration; live on apply.)

**Sequenced last on purpose — see §2.** Ships after 441 is fixed, or gated behind it.

**What changes.** A migration in the exact shape of `732` (committed today by the 458 lane):
a pre-guard counting the verbatim anchor that `RAISE EXCEPTION`s if it has moved
("re-read it and re-anchor rather than overwriting a prompt this migration has not seen"),
a surgical `jsonb_set` + `replace` on
`{workflow,steps,compose_plan,config,prompt_template}`, and a post-verify **in a `DO` block
that raises** — never a block of bare `SELECT`s, which cannot stop the `COMMIT`.

⚠ **732 touches the same ROW at a different PATH** (`generate_tool_html`), so the two
compose in either order. *The row is not the unit; the JSON path is.* Checked, not assumed.

The prompt gains `computed_values` **and the rule that governs it**:

> A tool that computes anything must assert at least one worked example — and **the expected
> value may not be whatever the tool prints.** State the worked example in the Behaviour
> contract with the arithmetic that produces it, then assert that. If you cannot derive the
> expected value independently of the code above, **do not emit a `computed_values` check**;
> say in Dependencies that the tool has no independently derived worked example.

**The refusal arm is the load-bearing half.** A generator that emits a `computed_values`
check it could not derive is worse than one that emits none.

⚠ **The 3,000-character cap on the PLAN is a real constraint** and the text above must fit
inside it, or the change trades a value assertion for a truncated document.

⚠ **This is the phase that arms the Tier-2 `no_auto_fix` trap** (LANDMINES §8626): Tier 2
ignores `no_auto_fix` entirely and appends three built-in shell failures *outside* the
criteria loop, so `computed_values` beside the four existing health checks is precisely the
combination that lets Tier 2 dispatch `tool-improver` at a **shared** component. The mcalc
lane has been asked for its judgement here (CONTRIB 2026-09-03) rather than re-deriving it.

**Live on apply?** **Yes** — DB config is live immediately. So this is the phase whose blast
radius arrives without a roll, which is a second reason to put it last.
**Council scope?** Yes (appliable migration under `docs/agent_docs/sql_for_agents/`).

**Red-induction — and this is the bug's own bar.** *"Take a passing tool, change one constant
in its JS so it computes a wrong figure, re-run acceptance. Today it still passes. After the
fix it must fail."* Plus: re-run §1's census restricted to `created_at > <the apply moment>`;
`assert_no_value` for `tool-generator` must fall **for new fences**. **A fix that only adds
checks which pass is indistinguishable from no fix.**

---

### P5 — The door refuses. (Deferred, and named so it is not forgotten.)

Once P4 has made the rule satisfiable, P2's note becomes a refusal: a fence that drives
inputs and asserts nothing is rejected at the write door. **Deferred, not dropped** — an
"inert until X" line is how a correct action comes to look premature, so the trigger is
written down: **P5 opens when the `created_at`-windowed count in P3 shows the intake has
stopped.** Until then, refusing strands tools.

### NOT DOING — backfill the 115 (the bug's candidate 3)

Real work, tool by tool, and it needs P4 first or the next generated fence re-creates the
gap. More importantly **P1 removes the reason to rush it**: once a value-less PASS reports
itself as `passed_liveness_only`, the 115 are honest-but-weak rather than false-and-strong,
which is a state the estate can live in while the intake is closed. Backfill is then a
per-site decision with a per-site oracle, which is where it belongs — and it is exactly what
the mcalc and staged_component_build lanes already do well.

## §4 What is still open

1. **Where an expected value may come from, in the general case.** `verify_criteria.py`'s
   three labelled strengths (DEFINITION / REGISTER / CONVENTION) are the honest version and
   they exist — but in one lane's directory, and possibly only because mortgages have
   published formulae and SDLT has a legal register. **Asked of the `loancalculator` lane**
   (which owns `toolgolden.py` and the capture discipline) 2026-09-03. Until answered, P4's
   rule is stated as a refusal-to-guess rather than as a provenance taxonomy.
2. **Whether a provenance LABEL should be a fence field.** Attractive — but per LANDMINES
   §8989 an unknown key in a fence is dropped in silence, so it would have to be added to
   the runner's decode struct **and** `experienceCheckTypeFields` or it is decoration. Held
   until (1) is answered; a taxonomy is worth encoding, a guess is not.
3. **441's landing order** relative to P4. Asked of the mcalc lane.
4. **fable's read of this plan** (see the note at the head).
