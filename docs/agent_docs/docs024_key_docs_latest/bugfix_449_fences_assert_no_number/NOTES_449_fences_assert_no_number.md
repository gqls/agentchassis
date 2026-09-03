# NOTES — bug 449, no fence the tool-generator writes ever asserts a number

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

Bug file:
`bugs_open/449_HANDOFF_2026-09-02_no_fence_the_tool_generator_writes_ever_asserts_a_number_so_a_calculator_that_computes_garbage_passes_acceptance.md`

---

## 2026-09-03 — session opens: is there an active thread, and is the bug still true?

### The thread was inactive, so this session resumed it

`scripts/who-owns.py 449` names `mortgagecalculator_couk_adoption` [ACTIVE, 32 commits/14d]
as the owner. That verdict is **lagging by construction** (it reads commits), so I checked
the clock rather than trusting the label:

- The lane's last commit of any kind: `bcef68058`, **2026-09-02 22:35:15 +0100**. The bug
  itself was filed 2 minutes earlier (`fd33fe4f9`, 22:33:22) and the lane went quiet
  immediately after.
- `git log --since='2026-09-03 00:00' -- .../mortgagecalculator_couk_adoption/` → **empty**.
- No commit today mentions 449 in its subject.
- Today's fleet restart wave is real and large — 40+ commits between 11:23 and 11:44 from a
  dozen lanes — and **the mcalc lane is not in it.**
- `git status` on `internal/adapters/browserrunner/run_checks_action.go` and on
  `platform/orchestration/actions/*criteria*` → **clean**, so no uncommitted session is
  mid-fix in the fence code either (the check `who-owns.py` cannot do).

The bug file's own status line agrees: *"Status: OPEN. Nothing changed, nothing dispatched."*

**Conclusion: inactive. Resumed here.**

### The bug is still valid — and it is getting WORSE while nobody watches

Two independent re-measurements, both first-hand today.

**(a) The cause is still in place.** Both fence-authoring agents still lack the vocabulary:

```sql
SELECT type,
       (default_config::text LIKE '%computed_values%') AS knows_computed_values,
       (default_config::text LIKE '%interaction%')     AS knows_interaction,
       updated_at
  FROM agent_definitions
 WHERE default_config::text LIKE '%```criteria%'
   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

```
        type        | knows_computed_values | knows_interaction |          updated_at
--------------------+-----------------------+-------------------+-------------------------------
 experience-planner | f                     | t                 | 2026-09-03 08:56:53.045885+00
 tool-generator     | f                     | t                 | 2026-09-03 08:56:53.045885+00
```

⚠ **`updated_at` of today, 08:56, is NOT a prompt change and must not be read as one.**
I checked before believing it: `SELECT count(*) FROM agent_definitions WHERE updated_at =
'2026-09-03 08:56:53.045885+00'` returns **208** — a bulk touch of every row at one second.
Had I quoted the timestamp as "the prompt was revised this morning" I would have been
confidently wrong about the one fact the whole bug rests on.

**(b) The census has grown by addition, exactly as the counting-date rule predicts.**
`bugs_open/449` §2 was measured 2026-09-02. Re-run today over `doc_plans` (`subject_type
='tool' AND is_current`, fence = the ` ```criteria ` block):

| author | fences | assert **no value at all** | uses `computed_values` | **drives inputs** (fill/select) | **drives inputs AND asserts nothing** |
|---|---|---|---|---|---|
| `tool-generator` | **186** (was 170) | **115** (was 107) | **0** | 91 | **55** |
| `operator:bugfix224-session` | 16 | 0 | 16 | 16 | 0 |
| `webdesign_couk_thread` | 14 | 4 | 0 | 6 | 0 |
| `operator:mortgagecalculator-lane-a4` | 8 | 0 | 8 | 8 | 0 |
| `operator:staged_component_build` | 8 | 0 | 6 | 7 | 0 |

**[MEASURED 2026-09-03.]** `+16 fences and +8 blind ones in about 24 hours`, and the newest
`tool-generator` fence carries `created_at` of **today**. The defect is not a standing
backlog to be tidied — it is a **live intake**, and every hour the fleet runs it mints more.

The two right-hand columns are mine, not the bug's, and they are the sharper cut:
**55 fences DRIVE INPUTS with `fill`/`select` and then assert no value of any kind.** The
fence itself declares "this tool takes input"; it then declines to check what came out. That
subset needs no classifier to identify — the evidence is inside the fence — which matters,
because a guarantee conditional on a classifier inherits the classifier's gaps.

### The cause, read out of the prompt rather than inferred

I dumped `agent_definitions.default_config` for both agents and read
`workflow.steps.compose_plan.config.prompt_template` (`tool-generator`, 2,766 chars).
It enumerates a **closed** vocabulary — four mandatory checks (`selector_exists` boots,
`no_console_errors`, `page_status_ok`, `no_horizontal_overflow` mobile-fit) plus "ONE
interaction check ONLY if you can copy real ids", and it says in terms:

> "No other check type exists for interactions — never emit `"type":"click"` or
> `"type":"fill"` as a check type."

`computed_values` appears nowhere in either agent's config. So the 0-of-186 is not a
modelling failure and not a hard case: **the type is never a candidate.** The prompt also
caps the whole PLAN at "under 3000 characters", which is a real constraint on any fix that
adds text to it.

> **On the diagnosis loop (090).** Not run, deliberately, and the CLAUDE.md norm of
> 2026-07-31 asks me to say why rather than omit it silently. That norm binds a session
> *filing* a cross-cutting root cause. This one is already filed by another lane, and the
> cause is not an inference from greps — it is the literal absence of a token in a prompt I
> read in full, plus a `write_doc_plan` seam I enumerated by grepping every Go writer of
> `doc_plans`. Both are self-evidencing: they could have come out otherwise and did not.
> The structural claim I am adding (single write door) is verified below, not asserted.

### The seam, enumerated rather than assumed

```
$ grep -rn "INSERT INTO doc_plans\|UPDATE doc_plans" --include='*.go' . | grep -v _test
platform/orchestration/actions/write_doc_plan_action.go:125   UPDATE doc_plans   (supersede)
platform/orchestration/actions/write_doc_plan_action.go:136   INSERT INTO doc_plans
platform/orchestration/datahelpers/travelling_docs_rekey.go:52 UPDATE doc_plans SET subject_key   (rekey only, never body)
docs/agent_docs/.../travelling_docs/write_doc_plan_action.go   (a doc artefact, not compiled)
```

So **`write_doc_plan_action.go` is the only production Go writer of a PLAN body**, and
exactly three live agents reach it:

```sql
SELECT type FROM agent_definitions WHERE default_config::text LIKE '%write_doc_plan%'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
 → experience-planner | experience-register-writer | tool-generator
```

⚠ **But operator scripts bypass it.** `install_fences.py` in the mcalc lane writes
`doc_plans` rows over `psql` directly. So the door governs every *generated* fence and no
hand-installed one — which is the right way round (the hand-installed ones are the 16+8+6
that already carry `computed_values`), but it means the door is not a total guarantee and
must not be described as one.

### The design tension that makes fix candidate 1 unsafe as written

I read `runComputedValues`'s docstring in `internal/adapters/browserrunner/run_checks_action.go`
in full. It says the values

> "are not authored by hand and are not judged for correctness here. They are CAPTURED from
> the tool while it is known good (`toolgolden.py --emit-criteria`) and then defended … a
> golden captured from an already-wrong tool pins the wrong answer."

So `computed_values` is a **regression/pinning** check by construction, not a correctness
check. At generation time the tool is newborn: there is no known-good state to capture, and
the generator is handed `{{.generated_html}}` — the tool's own code — so any expectation it
derives shares a failure mode with the implementation it is meant to police. The estate has
shipped that mistake twice already (`bugs_open/224`, `bugs_open/225`: an expired £625k FTB
SDLT cap certified green for sixteen months).

**Therefore "just teach both agents the type" is not sufficient on its own**, and the bug
says so itself in its own ⚠ ("an emitted value is not an expected value"). Whatever ships
has to say what an expectation's SOURCE is, and let the platform tell an independently
derived value from one captured off the page. The machinery for the strong version already
exists — `verify_criteria.py` re-derives from DEFINITION / REGISTER / CONVENTION at three
labelled strengths — but it lives in one lane's directory, not in the framework.

### Prior art found, so it is not rebuilt

- `platform/orchestration/actions/experience_criteria.go` — `ValidateExperienceCriteria`,
  rules P1–P11, each traceable to a live failure. **Only production caller:**
  `write_experience_pattern_action.go`, the experience-PATTERN register. Tool fences have
  never been through it.
- `write_doc_plan_action.go` already carries the precedent for validating a tool fence at
  the door: for `subjectType == "tool"` it refuses a malformed `facts` declaration, sharing
  `criteriaFactsFromValue` with P11 rather than re-spelling the rule (2026-08-24,
  `bugs_open/288` defect A). Its comment records that P11 "had never once seen a tool fence"
  — the same gap, one rule along.
- `check_tool_acceptance.go` (Tier 2) already holds the honesty discipline this bug wants
  extended: *"passes → a finding only (Tier-2 pass must never be read as 'the tool works')"*.
  It runs on a schedule over every eligible tool, so it can see the standing 115 with no
  backfill, and it has a `needs_criteria` doc_note path with a 30-day per-subject cooldown.

---
