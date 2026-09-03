# HANDOFF — `bugs_open/449`, continue here (2026-09-03b, ~16:2x BST)

**Supersedes `HANDOFF_2026-09-03_continue_here.md`**, which is stale in three ways: it says P1 is
unproven (it is now **PROVEN**), it says two lanes' questions are unanswered (mcalc **answered**, and
two blockers dissolved), and its roll time is wrong (**13:28Z**, not ~13:55Z).

**Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_449_fences_assert_no_number/`
**Bug:** `bugs_open/449_HANDOFF_2026-09-02_no_fence_the_tool_generator_writes_ever_asserts_a_number_so_a_calculator_that_computes_garbage_passes_acceptance.md` — **§8 is this lane's section.**
**Read next:** the bug's §8 → this file's §4 (P4's design, now settled) → `NOTES_…` last two sections
(the evidence, and the two missteps) → `PLAN_…` for phase order.

---

## 1. Where we are in one paragraph

A generated acceptance fence never compares a **number**, so a calculator that prints a confidently
wrong figure passes Tier 4 and its record reads PASSED. **The honest-record half is now DONE AND
PROVEN**: a Tier-4 verdict states the scope of its own claim, observed firing on a real page at
14:00:07Z today. The daily sweep is live and reporting. **The cause is still untouched** — the
authoring agents do not know the `computed_values` type — but **P4 is no longer blocked**, because
the two blockers this lane was holding it behind have been answered and dissolved. Council
**APPROVED** (round 2). **The bug stays OPEN.**

## 2. Exact state

| | state | evidence |
|---|---|---|
| **P1** verdict states its scope | **LIVE and PROVEN in production** | `acceptance-run` note, `tool-idea-stage-identifier`, **14:00:07.683828+00**, carries the `Scope of this verdict: ⚠ LIVENESS ONLY` line; all pre-roll notes do not |
| **P2** door records a blind fence | **live in binary, NOT yet exercised — and NOT blocked** | 0 notes; 0 post-roll tool PLANs. See §3 — the reason is upstream, benign, and measured |
| **P3** daily sweep | **LIVE and PROVEN** | CronJob `fence-value-assertion-check` @ `40 7 * * *`; note written (241 fences, 58 driving-and-blind, 13 new) |
| **P4** teach the prompt | **ROUND 2 SUBMITTED, NOT APPLIED** — migration `749`; round 1 was REVISE and changed the file | §4d |
| **P5** door refuses | **deferred, trigger written down** | PLAN §P5 |
| council | **APPROVED** round 2 | `Council-Reviewed: 8745ad9e-1802-4e08-a9b0-eb493cd11243` |

**P1 was checked for disconfirmability, not just for the string.** `criteriaAssertionPhrase` has four
outcomes and the unit tests pin the negative direction; on this subject the *sub*-branch was also
right (the fence has no `fill`/`select`, so the "page loads and responds" variant is correct). Do not
re-verify P1 — it is done.

## 3. P2: why it has not fired, and what to watch for

**Do not read the zero as a fault.** `[MEASURED 16:04Z]` 0 `fence_asserts_no_value` notes and **0**
tool PLANs written since the roll. The old handoff offered two causes; **there is a third and it is
the one we are in:**

- A generator run *did* happen post-roll (`3f5cb558-…`, 16:04:42Z) and wrote no PLAN because it died
  at **`save_tool`** — `tool birth refused (instance scope): script is not mechanically provable`.
- The live chain is `… → suggest_related_pages → save_tool → compose_plan → write_plan → …`, so
  **the refusal is UPSTREAM of the PLAN write.** No PLAN, no note, and nothing for a
  count-the-output control to see.
- ⚠ That run reads `status=COMPLETED` with `error` **NULL** — the bugfix-099 trap. The message is in
  `collected_data->>'__step_error'`.

**It is not a standing blocker, and this was measured rather than assumed:** the instance-scope gate
landed **2026-08-21/23** (RFC_032 lane, `tool_birth_instance_scope.go`), eleven days pre-roll, and
**19 of 19** generator runs in the prior 72 h cleared it. One refusal, script-specific.

**What P2's first fire needs:** the door is gated on `DrivesButAssertsNothing()`
(`write_doc_plan_action.go:219`) — the fence must **drive inputs AND** assert nothing. P1's subject
drove nothing and could never have tripped it. Qualifying rate is **55/187** of `tool-generator`'s
fences, so roughly one run in three. Just re-run §3's two queries from the old handoff; no action
needed until a post-roll PLAN exists.

## 4. P4 — the next real work, and its design is now SETTLED

**Blockers 1 and 3 are DISSOLVED** (mcalc lane's reply, appended at line 113 of
`…/mortgagecalculator_couk_adoption/CONTRIB_2026-09-03_from_the_449_lane_I_am_taking_the_FRAMEWORK_half_you_keep_the_site_half.md`):

- **Do NOT wait for `441`.** *"Not imminent and nothing is scheduled … treat '441 lands first' as
  unavailable."* A fence written **at birth** is safe — the generator emits selectors from the
  template it just wrote, and `ScopeToolBirthTemplate` guarantees the tool carries that template
  verbatim as `rendered_html`. **The 441 risk lives in BACKFILL only.**
- **So: ship for NEW fences; do NOT backfill the 55 standing blind ones.** Backfill converts silent
  blindness into loud false failures, because `runComputedValues` **fails rather than skips** on a
  missing element (`page.Count(sel) == 0 → problems`, verified in source).
- **`no_auto_fix` fear was unfounded** — verified in source, not taken on trust:
  `evaluateStaticCriteria`'s outer arm is `default: skip(ch.ID, ch.Type+" is not statically
  checkable (Tier 4)")`, and `computed_values` is not an arm, so **Tier 2 SKIPS it** and cannot
  dispatch `tool-improver` on it. The shared-component exposure is the three built-in shell checks,
  which fire regardless of fence content — pre-existing and orthogonal (`bugs_closed/285`).
  **Set `no_auto_fix: true` anyway**: an arithmetic failure means the maths or the law moved, which
  is a human's call.

**Blocker 2 is open at `loancalculator` but no longer holds P4 up.** The answer is already in
`runComputedValues`' own contract (`run_checks_action.go:790-808`), which says the type is a
**regression** check: values are *"CAPTURED from the tool while it is known good … and then
defended"*, and *"a golden captured from an already-wrong tool pins the wrong answer — the capture
script therefore refuses to emit for a tool whose outputs do not react to its inputs."*

**Therefore the load-bearing half of P4 is the REFUSAL ARM**, and it is now evidenced twice
(the docstring above; mcalc's `verify_criteria.py`, which refused them with *"NOT VERIFIED (no
independent model): fact-finder, portfolio"*). The generator must derive the expectation from
something that is **not** the tool it was just shown — a published formula (DEFINITION), a fact the
site already cites (REGISTER), or a rule read off the tool (CONVENTION, explicitly weaker) — or emit
**no** `computed_values` check and say so in Dependencies. **A guessed expectation is worse than
none.** What is genuinely still open is whether that taxonomy generalises past domains with published
formulae; that is a question, not a blocker.

**Shape of the change:** a migration in the exact shape of `732` — pre-guard counting the verbatim
anchor, `RAISE EXCEPTION` if it moved, idempotency arm, post-verify in a `DO` block that **raises**
(never bare `SELECT`s) — surgically editing `{workflow,steps,compose_plan,config,prompt_template}` on
`tool-generator`, and the equivalent on `experience-planner`. ⚠ `732` touches the same **row** at a
different **path**, so they compose in either order — *the row is not the unit, the JSON path is*.
⚠ The prompt caps the PLAN at "under 3000 characters"; new text must fit or it trades a value
assertion for a truncated document. ⚠ Council gate applies (migrations in scope since 2026-08-19).
⚠ §6's red-induction ("break a constant, re-run, must fail") is still owed and belongs to P4 — P1/P2
**grade**, they do not strengthen, and cannot satisfy it.

### 4a. SCOPE, RESIZED by measurement — do NOT ship both agents in one migration

`[MEASURED 2026-09-03 16:3xZ]`, and both facts postdate the plan that said "two prompts":

1. **`experience-planner` carries the fence in THREE steps** — `compose`, `recompose`, `reframe`
   (not one). Editing one leaves two authoring the old vocabulary.
2. **`experience-planner` has authored 3 fences, EVER.** `tool-generator` is 187 of the 241 tool
   fences and every blind driving one.

**Recommendation: ship `tool-generator` alone first; make `experience-planner`'s three paths a
separate follow-on.** Three extra verbatim anchors for ~1% of the population is a bad trade in one
migration, and `732`-shaped guards are per-anchor — each one is a `RAISE EXCEPTION` that can go stale
independently.

### 4b. The design rule to write, which is NARROWER than "teach the type"

The generator's only inputs are the spec and `{{.generated_html}}` — **the artefact whose correctness
is in question.** No register, no formula source. So "emit `computed_values`" unqualified can only
yield a value read off the tool, i.e. the pinning failure. The one independent oracle actually
present is **the model's own knowledge of a published formula**, applied to inputs it chooses:

- **Derivable (DEFINITION) → emit.** A published, checkable rule: annuity repayment, compound
  interest, VAT, BMI, unit conversion, margin arithmetic. The generator picks the inputs, does the
  arithmetic itself, and the expectation never touches the page.
- **Not derivable → REFUSE.** An arbitrary scoring heuristic where "correct" is whatever the code
  says (`tool-idea-stage-identifier`, `tool-process-automation-scorer`). No oracle exists, so a
  check could only pin the implementation.

**The instruction is conditional and its default is refusal**, and it must require the arithmetic be
shown in `## Dependencies` so a reviewer can see which case the generator believed it was in.

### 4c. The shape, pinned — do not re-derive it

From `criteriaCheck`/`criteriaStep` (`run_checks_action.go:221-246`), confirmed against a live
`operator:staged_component_build` fence:

```json
{ "id": "<kebab-id>", "type": "computed_values", "profiles": ["desktop"],
  "steps": [ { "action": "fill", "selector": "#volume", "value": "5000" } ],
  "expect_values": { ".result-card.highlight .result-value": "$2,000.00" } }
```

`expect_values` is a **map** selector → exact text after `steps` run. Actions: `fill`/`click`/
`select` (with `value`), `reload`. Comparison is `collapseSpace` both sides — **whitespace-
insensitive, everything else exact**, so currency symbol, separators and 2dp must match as rendered.
`no_auto_fix` / `no_auto_fix_reason` are **top-level fence keys**, not per-check.

**Anchor for the pre-guard** (distinctive, one occurrence, and the claim being amended):
*"No other check type exists for interactions — never emit `"type":"click"` or `"type":"fill"` as a
check type."*

**Budget:** `compose_plan` is `claude-sonnet-5`, **`max_tokens: 4000`**; the prompt is **2,766 characters / 2,782 bytes** (~~2,783 B~~ — `wc -c` counted psql's trailing newline) / 45
lines; a `computed_values` check costs ~350-450 chars of the output document. **Raise the instructed
"under 3000 characters" cap to ~3,500** — there is ample token headroom, and leaving it makes the
model trade the new assertion against prose it was also told to write.

### 4d. STATE: migration 749 is written and tested. It is NOT applied.

**Files** (both committed):
`docs/agent_docs/sql_for_agents/749_tool_generator_learns_the_value_assertion_and_when_to_refuse_it.sql`
and its `_ROLLBACK` sidecar.

**Tested against the LIVE row inside a rolled-back transaction** — recipe and its four gotchas are
`RUNBOOK` §8:

```
2766 chars --> apply --> 5046 --> apply AGAIN (UPDATE 0, "already applied") --> reverse --> 2766
```

Byte-identical round trip; nothing was committed to the database.

**Council: ROUND 1 = REVISE (16:56Z); ROUND 2 SUBMITTED on the same correlation, verdict not yet read.**

Round 1's gating objection was against my **submission**, not the file — I had put a placeholder in
the sketch. Round 2's sketch is the file itself from `BEGIN;` to `COMMIT;`, so they cannot diverge.
**Three objections changed the artefact and are already in it:**

- `snapshot_agent('tool-generator', …)` in the pre-guard, after the idempotency return (raised
  independently by `tooling_provenance` and `debug_historian`). Verified: backup rows 0 → 1 → **1**,
  pre-image reads back at 2766 chars. ⚠ It writes to **`agent_definitions_backup`**, NOT to
  `agent_definitions` with `is_snapshot=true` — a count against the latter reads 0 and looks broken.
- The prompt gained **literal input vectors** and the **value/format split**, from the
  `loancalculator` lane's answer (§§A–C of the bug).
- The containment claim is now an **enumeration of all five `improve_tool` producers**, not one grep.

**One objection is NOT closed and I have not claimed it is:** `bug_historian` — the refusal arm is
prose with no code-side check. The named follow-on is a detector comparing a new `expect_values`
against what the tool itself renders, flagging a MATCH where `## Dependencies` shows no working.


`Council-Submitted: dda64bd1-2d34-4ee5-b903-c5bb2644733a`. Budget ~30 minutes, not ~2 — the dispatch
queues behind the fleet. Read it with:

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id='dda64bd1-2d34-4ee5-b903-c5bb2644733a' AND kind='council_report' ORDER BY created_at;
SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
```

⚠ **Do NOT write `Council-Reviewed:` until you have READ an approved verdict** — `098` buckets an
unread claim as MISMATCH, which is the coverage report's dishonesty surface.

**THE APPLY IS THE LIVE STEP AND IT IS DELIBERATELY NOT DONE.** A migration is live the moment it
applies, with no image tag to roll back, and this one changes what **every tool built from now on**
puts in its PLAN. Two things should happen first: the verdict, and the owner's call. The two
questions I most want a human on are in the submission's `risks` and are worth repeating:

1. **Is a prompt-level refusal arm sufficient?** An LLM handed a working calculator may find it easy
   to believe it *derived* a figure it in fact read off the page. The mitigation is that the rule and
   the working must appear in `## Dependencies`, so a reviewer sees the claimed derivation and not
   just its output — but nothing in SQL enforces that.
2. **Should statutory rates be excluded from the licensed sources?** A model's "known" VAT band or
   threshold can be superseded, and a wrong one would pin a wrong expectation confidently — which is
   `bugs_open/224`'s failure mode arriving by another route. Formulae and conversions are safer than
   rates.

**If the verdict is REVISE**, the objections come back with the reviewers' own read-only checks
answered; revise and resubmit with `RESUBMIT_CORR=dda64bd1-2d34-4ee5-b903-c5bb2644733a` so the trail
accumulates.

⚠ **Migration numbering collided once already**: another session claimed `748`
(`748_planner_states_its_page_type_decisions_structurally_HOLD.sql`) while this was being written, so
it was renumbered to `749`. **Re-check the highest number immediately before applying** — the same
thing can happen again, and the guards are numbered in their own error strings.

## 5. The census, refreshed — quote THIS one

`[MEASURED 2026-09-03 16:04Z]` `is_current` tool fences. **Do not quote the old figures; this moved
186 → 187 in a day.** Re-run `RUNBOOK` §1b or `scripts/audit-fence-value-assertions.py`.

| created_by | fences | assert_no_value | uses_computed_values | drives_inputs | drives_but_asserts_nothing |
|---|---|---|---|---|---|
| `tool-generator` | **187** | 116 | **0** | 91 | **55** |
| `operator:bugfix224-session` | 16 | 0 | 16 | 16 | 0 |
| `webdesign_couk_thread` | 14 | 4 | 0 | 6 | 0 |
| `operator:mortgagecalculator-…-701-rekey` | 8 | 0 | 8 | 8 | 0 |
| `operator:staged_component_build` | 8 | 0 | 6 | 7 | 0 |

**The single sharpest line in this lane: `tool-generator` has authored 187 fences and
`uses_computed_values` = ZERO.** All ~~38~~ **30** value-asserting fences in the estate came from an operator or
> **CORRECTED 2026-09-03 ~16:4x — the count was 30, not 38, and it was MY ARITHMETIC, not a stale
> census.** 16 (`bugfix224-session`) + 8 (`mortgagecalculator-…-701-rekey`) + 6
> (`staged_component_build`) = **30**. I took `staged_component_build`'s *fences* figure (8) from
> the adjacent column instead of its *uses_computed_values* (6), and mis-added on top. Caught by a
> second query run for a different purpose — `count(*) FILTER (WHERE body LIKE '%expect_values%')`
> grouped by `subject_type` returned **30** and disagreed with the total I had already published in
> four places. **The conclusion is untouched: `tool-generator` accounts for ZERO of them.** The
> number was decoration on that finding and I still got it wrong. `WRONG_CALLS.md` has it.

a lane, never from the agent. `max(created_at)` = **12:35:59Z today** — a live intake, not a backlog.

## 6. Other threads, and what is owed

| lane | state | owed |
|---|---|---|
| `mortgagecalculator_couk_adoption` | ACTIVE; owns `441`/`448` + the site half | **Nothing — they answered all 3.** They accept the division and hand `449` to this lane outright. A thank-you/ack is courteous, not required |
| `loancalculator` | ACTIVE; owns `toolgolden.py` | **3 questions still unanswered** — `CONTRIB_2026-09-03_from_the_449_lane_where_may_an_expected_value_COME_FROM…` in their dir. Not blocking (§4) |
| `458` lane | ACTIVE; shares the `tool-generator` row | Told in `bugs_open/458` §11 that `0325ddebb` left `TestStylesheetGutted_TokenSetMatchesCanonicalCSSTokens` RED at HEAD. **Not mine to fix** — but P4 edits the same row, so check with them before migrating |
| RFC_032 / instance-scope lane | ACTIVE | Nothing owed. Their gate refused one generator run at 16:04 (§3). Not a defect in their work, and not this lane's to diagnose |

⚠ **Ownership checks lag by construction.** Re-run `scripts/who-owns.py 449` and check the tree at
each phase boundary.

## 7. Missteps already paid for — in `WRONG_CALLS.md`, do not repeat

1. **A demand control counting the WRITER'S OUTPUT cannot see attempts that died UPSTREAM.** Count
   attempts too, and read the step order to name what sits before your write. *(New this session —
   it was my own §3 discriminator.)*
2. **A CONTRIB you sent is a file the other lane can APPEND TO** — the reply carries no new filename
   and no notification. **Re-read the letter you sent**; `ls -lt` on their directory cannot see it.
   *(New this session — cost a handoff and an owner README both claiming a resolved blocker.)*
3. **A `cd` in one Bash call re-points every later relative path** — quiet, exit 0, zero matches.
4. **A demand control pinned to a lane NAME**; `created_by` is free text. Phrase controls over the
   **property**.
5. **`updated_at` on an agent row is not a diff** — 208 rows share that second.
6. **A pathspec commit still takes a same-file passenger.**

## 8. If you only do one thing

**Start P4** (§4). Its design is settled, its two blockers are gone, and the honest-record half it was
sequenced behind is proven live. Check with the `458` lane first, since P4 edits the
`tool-generator` row they also touch.
