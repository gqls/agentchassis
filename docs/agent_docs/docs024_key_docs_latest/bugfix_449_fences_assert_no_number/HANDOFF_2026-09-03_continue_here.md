# HANDOFF — `bugs_open/449`, continue here (2026-09-03, ~14:1x BST)

**Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_449_fences_assert_no_number/`
**Bug:** `bugs_open/449_HANDOFF_2026-09-02_no_fence_the_tool_generator_writes_ever_asserts_a_number_so_a_calculator_that_computes_garbage_passes_acceptance.md` — **§8 is this lane's section; read it first.**
**Read next, in this order:** the bug's §8 → `PLAN_2026-09-03_449_fences_assert_no_number.md` (the phase order and *why* it inverts) → `RUNBOOK_…` §6/§7 (the two checks that cost the most to get right) → `NOTES_…` (evidence, and the missteps).

---

## 1. Where we are in one paragraph

A generated acceptance fence never compares a **number**, so a calculator that prints a
confidently wrong figure passes Tier 4 and its record reads PASSED. Three halves of a fix are
**live**: the Tier-4 verdict now states the scope of its own claim, the write door records a fence
born blind, and a daily CronJob reports the standing stock and — the number that matters — the
**new** blind fences per window. **The cause is deliberately untouched**: neither authoring agent
knows the `computed_values` type, and teaching them today would pin garbage and, with `441` live,
produce loud *false* failures. Council **APPROVED** (round 2). **The bug stays OPEN.**

## 2. Exact state, and the distinction that matters most

| | state | evidence |
|---|---|---|
| **P1** verdict states its scope | **live in binary, NOT yet exercised** | `v1.0.1359`; `liveness_only` + `Scope of this verdict` present at `/proc/1/exe` |
| **P2** door records a blind fence | **live in binary, NOT yet exercised** | `fence_asserts_no_value` present at `/proc/1/exe` |
| **P3** daily sweep | **LIVE and PROVEN** | CronJob `fence-value-assertion-check` @ `40 7 * * *`; manual Job pod exit 0; note written (241 fences, 58 driving-and-blind, 13 new) |
| **P4** teach both prompts | **NOT STARTED — blocked, see §5** | — |
| **P5** door refuses | **deferred, trigger written down** | PLAN §P5 |
| council | **APPROVED** round 2, 13:11:41Z | `Council-Reviewed: 8745ad9e-1802-4e08-a9b0-eb493cd11243` |

⚠ **"Live in the binary" is NOT "observed in production", and this is the first thing to close.**
At the last check there were **zero** `fence_asserts_no_value` notes, and the newest
`acceptance-run` note (08:47Z, *pre-roll*) does **not** carry the scope line — because no tool PLAN
has been written and no Tier-4 run has completed since the roll. **Until a first-fire check returns
a row, P1 and P2 are unproven in production.**

## 3. FIRST-FIRE CHECKS — run these before anything else

```sql
-- P1: has any Tier-4 verdict since the roll carried the scope line?
SELECT created_at, subject_key, (body LIKE '%Scope of this verdict%') AS carries_scope
  FROM doc_notes WHERE categories ? 'acceptance-run'
 ORDER BY created_at DESC LIMIT 5;
-- Expect carries_scope = t on anything created after ~2026-09-03 13:55Z.

-- P2: has the door recorded a blind fence yet?
SELECT created_at, subject_key, left(body,160) FROM doc_notes
 WHERE categories ? 'fence_asserts_no_value' ORDER BY created_at DESC LIMIT 5;
-- 0 rows as of handoff. The FIRST row is P2's proof.
```

⚠ **A zero here has two causes and they need opposite actions:** nothing has been authored/run yet
(wait), or the mechanism is broken (investigate). **Discriminate with a demand control** — has
anything at all been written through those paths since the roll?

```sql
SELECT max(created_at) AS newest_tool_plan FROM doc_plans WHERE subject_type='tool';
SELECT max(created_at) AS newest_acceptance_run FROM doc_notes WHERE categories ? 'acceptance-run';
```
If those are older than the roll, the zero means **not yet exercised** and nothing is wrong. If a
tool PLAN *has* been written since the roll and P2 still shows zero, the rule is broken — and the
first thing to check is in `write_doc_plan_action.go`'s own warning line (`doc_notes_subject_type_check`).

## 4. What NOT to redo — this is the expensive part of the context

- **Do NOT "just teach the generator `computed_values`".** It is the obvious move, it is the bug's
  own candidate 1, and it is **unsafe as written**. `computed_values` is a *regression* check: its
  docstring says values are "CAPTURED from the tool while it is known good and then defended", and
  that "a golden captured from an already-wrong tool pins the wrong answer". At birth nothing is
  known good and the generator is handed the tool's own HTML, so the expectation shares a failure
  mode with the code it polices. `bugs_open/224` / `bugs_closed/225` is the instance: an expired
  £625k FTB SDLT cap **certified green for sixteen months**.
- **Do NOT expect P1/P2 to make anything fail.** They **grade**; they do not strengthen. Nothing
  that passes today starts failing. §6's red-induction ("break a constant, re-run, must fail")
  **cannot** be satisfied by them and is still owed — it belongs to P4.
- **Do NOT use CLAUDE.md's `grep 'build provenance'`** to check whether a Go change shipped. That
  line does not exist in the source, and on `agent-chassis` the command returns **megabytes** of
  unrelated JSON rather than nothing, because log lines are single objects hundreds of KB wide.
  Use RUNBOOK §7's capability probe **with a control on both sides**.
- **Do NOT re-derive the seam or the caller set.** `write_doc_plan_action.go` is the only production
  Go writer of a PLAN body; exactly three agents carry the step (`tool-generator/write_plan`,
  `experience-planner/persist_plan`, `experience-register-writer/write_travelling_doc`) — verified
  structurally, not by substring. Operator scripts write `doc_plans` over `psql` and **bypass it**.
- **Do NOT read a `[MEASURED]` census as current.** `tool-generator` went 170 → 186 → 187 in a day.
  **Re-run RUNBOOK §1b, or `scripts/audit-fence-value-assertions.py`.**

## 5. P4 — the next real work, and its two blockers

**What it is:** a migration in the exact shape of `732` (pre-guard counting the verbatim anchor,
`RAISE EXCEPTION` if it moved, idempotency arm, post-verify in a `DO` block that **raises** — never
bare `SELECT`s), surgically editing
`{workflow,steps,compose_plan,config,prompt_template}` on `tool-generator`, and the equivalent on
`experience-planner`. ⚠ `732` touches the same **row** at a different **path**, so they compose in
either order — *the row is not the unit, the JSON path is*. ⚠ The prompt caps the PLAN at "under
3000 characters"; the new text must fit or it trades a value assertion for a truncated document.

**The load-bearing half is the REFUSAL arm:** if the generator cannot derive an expectation
independently of the code it was shown, it must emit **no** `computed_values` check and say so in
Dependencies. A guessed expectation is worse than none.

**Blocker 1 — `bugs_open/441`,** re-framed by its owning lane (mcalc) as a **live generator of stale
fences**. `runComputedValues` **fails rather than skips** on a missing element, by design. So value
assertions authored while 441 is live fail for the wrong reason and aim `tool-improver` at
arithmetic that was never wrong. **Ask the mcalc lane whether 441 lands before the next generator
run** — asked in their CONTRIB, unanswered.

**Blocker 2 — where an expected value may legitimately come from.** `verify_criteria.py`
(mcalc lane) re-derives from a non-page source at three labelled strengths — **DEFINITION**
(published formula), **REGISTER** (the site's own cited facts), **CONVENTION** (a rule read off the
tool, explicitly weaker). That is the honest answer and it exists — but in one lane's directory, and
possibly only because mortgages have published formulae and SDLT has a legal register. **Asked of
the `loancalculator` lane** (owner of `toolgolden.py`) 2026-09-03 — **unanswered**. Three questions
were put: is the three-strength taxonomy generalisable; does the toolgolden per-page-defaults trap
reappear at a new seam; and is `--emit-criteria`'s reactivity refusal the right birth-time primitive.

⚠ **Blocker 3 to think about before shipping P4, not after:** `no_auto_fix` (`LANDMINES.md`). Tier 2
**ignores it entirely** and appends three built-in shell failures *outside* the criteria loop, so
`computed_values` beside the four existing health checks is exactly the combination that lets Tier 2
dispatch `tool-improver` at a **shared** component. The mcalc lane was asked for its judgement.

## 6. Other threads, and what is owed to whom

| lane | state | owed |
|---|---|---|
| `mortgagecalculator_couk_adoption` | ACTIVE; filed 449, owns 441/448 and the site half | **3 questions unanswered** (division of labour; 441's landing order; `no_auto_fix`) — `CONTRIB_2026-09-03_from_the_449_lane_…` in their dir |
| `loancalculator` | ACTIVE; owns `toolgolden.py` | **3 questions unanswered** (sent via SendMessage, so *not on disk* — if it matters, re-send or write a CONTRIB) |
| `458` lane | ACTIVE; shares the `tool-generator` row | Told in `bugs_open/458` §11 that `0325ddebb` left `TestStylesheetGutted_TokenSetMatchesCanonicalCSSTokens` **RED at HEAD**, and that their own §9 verification (`-run TokenAudit`) cannot select it. **Still red at last check — not mine to fix.** |
| `staged_component_build` | quiet | nothing; its 6 fences are the existence proof P4 copies |

⚠ **Ownership checks are lagging by construction.** The mcalc lane read "inactive" at 11:50 and
restarted at 11:54. **Re-run `scripts/who-owns.py 449` and check the tree at each phase boundary.**

## 7. Commits this lane made (newest last)

```
aa258a7e3  449 resumed: bug RE-VERIFIED, a LIVE INTAKE not a backlog
4129709e7  WRONG_CALLS: a cd re-points every later relative sweep
9e4c7a4e9  CONTRIB into mcalc: framework half taken, 3 questions
10309eef6  449 PLAN: honest-record phases FIRST, authoring LAST
0b9a5c9e1  449 P1+P2  ← the Go halves (Council-Submitted)
862af90e2  CONTRIB into 458: their commit left a test RED at HEAD
e27aa00bb  register TP-009 + index row (+ a duplicate WII-035 row flagged)
e9ef673a5  449 §8 into the bug file
23c8a7d71  449 P3 + WRONG_CALLS (my own demand control was name-pinned)
d58658d31  LANDMINE: a Tier-4 PASS says nothing about arithmetic
8d22e56b8  lane docs: outcome, mutation table, two missteps
b15c07b2e  the 1358 roll did NOT carry P1+P2 (measured at the binary)
e9274c1fa  §8 correction + RECONFIRM the build-provenance landmine
<the round-2 commit>  council REVISE answered; detector SCHEDULED + proven
eb0e3a88c  register: sweep LIVE, Go halves inert
304388519  docs: the REVISE round in full; RUNBOOK §6/§7
58889613e  APPROVED + LIVE in v1.0.1359 — but LIVE is not EXERCISED
```

## 8. Missteps already paid for — in `WRONG_CALLS.md`, do not repeat

1. **A `cd` in one Bash call re-points every later relative path.** The quiet form returns zero
   matches, exit 0, no message — and I had used that sweep shape to establish a **negative**. Use
   absolute paths or a subshell; `pwd` before believing a zero. *(I hit it a second time the same
   session, on `kubectl kustomize`.)*
2. **A demand control pinned to a lane NAME.** `created_by` is free text, not an identity; the mcalc
   lane renamed itself hours earlier and my control survived on luck. It **failed closed**, which is
   why I would not have questioned it. Phrase a control over the **property**, and print the
   evidence that satisfied it.
3. **`updated_at` on an agent row is not a diff.** Both authoring agents read `2026-09-03 08:56:53`;
   **208 rows share that second**. Quoting it would have been confidently wrong about the one fact
   the bug rests on.
4. **A pathspec commit still takes a same-file passenger.** Mine took the mcalc lane's uncommitted
   `WRONG_CALLS.md` entry. Disclosed to them; nothing lost.

## 9. If you only do one thing

Run §3's first-fire checks. If P1's scope line has appeared on a post-roll `acceptance-run` note,
**this lane's shipped work is proven** and the next move is P4's two blockers (§5). If it has not,
check the demand controls in §3 before concluding anything is broken.
