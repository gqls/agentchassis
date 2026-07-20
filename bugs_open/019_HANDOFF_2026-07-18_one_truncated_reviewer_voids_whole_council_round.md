# 019 — One truncated reviewer voids the entire council round

> **FIX BUILT 2026-07-20 (bugfix-019 thread) — code committed `a3b606798`, config
> migration 177 committed `76ff5ed25` and APPLIED LIVE. Stays OPEN: inert until
> the next image roll after v1.0.1139.** Diagnosis, corrected mechanism and the
> full decision log live in
> `docs/agent_docs/docs024_key_docs_latest/bugfix_019_council_truncation/`.
>
> **The dominant mechanism was upstream of everything this file documents as the
> root cause.** Counted over 10 days: 9 truncation voids died at
> `execute_llm_prompt` (the provider client returns `""` + a hard error on
> `stop_reason=max_tokens`, so the partial never exists downstream and the seat's
> `error_step` routes the round to the terminal before later seats run); only 2
> died at the `json.Valid` hard error in `diagnose_council_decide` described
> under §"Root cause" below. Both mechanisms are now fixed:
>
> 1. `platform/aiservice` — `TruncatedError` carries the partial text instead of
>    discarding it (anthropic + ollama). This is what made fix-candidate 2
>    below impossible as written: there was nothing to repair.
> 2. `ExecuteLLMPromptAction` — a step with `tolerate_truncation: true` records
>    the partial (result carries `__truncated`) and SUCCEEDS, so the chain
>    continues to the seats that have not yet run — the behaviour the SECOND
>    CORRECTION below identified as "the half that matters". Armed by migration
>    177 on the `review_*` seats of council-gate/fix-proposer/feature-designer
>    (35 seats, verified; deliberately narrow — owner decision 2026-07-20 —
>    so experience-planner's `compose` still hard-fails).
> 3. `diagnose_council_decide` — salvage the verdict via `repairTruncatedJSON`
>    (marked `degraded`), else count the seat `unreadable` (distinct from
>    abstentions), surface it in report/metadata/return, and downgrade an
>    otherwise-`approve` to `revise` while any seat was unreadable. Fix
>    candidate 1 below, essentially as specified.
>
> The cap was NOT raised — reinforced by a new measurement: an
> `experience-planner/compose` call truncated at a **32,000**-token cap on
> 2026-07-19, so raising demonstrably moves the cliff rather than removing it.
>
> **How to verify after the roll** (supersedes the §"How to verify a fix" steps,
> which predate the upstream mechanism being known): pod-grep for
> `tolerate_truncation` in the chassis binary; then a cheap reproduction via a
> scratch council with one seat's `max_tokens` ~200 should produce a verdict
> whose `council_report` metadata carries `unreadable` ≥ 1 (or a `degraded`
> review) instead of `complete_invalid` — and `revise`, never `approve`, while
> any seat is unreadable. Council submission for this fix: correlation
> `2eed453a-9102-41e0-8838-7a711e99126b` — **two rounds, both REVISE, both
> COMPLETED without voiding** (10 seats round 1, 9 round 2 — themselves the
> first substantive rounds this week to survive the seat chain). Round 1's
> convergent objection earned a real second fix (`11a72dc31`: a partial that
> parses cleanly is still marked degraded via the `__truncated` marker; tolerated
> truncations legible in llm_call_log). Round 2's residue is provenance/process
> only — dispositions, the 177 rollback, and snapshot
> `bak_agentdef_councils_20260720` are in the workstream NOTES/RUNBOOK. Stopped
> at two rounds (no reviser loop exists; resubmission is not a free retry). No
> trailer on any commit — earned by APPROVED only.

*Found 2026-07-18 by the claims-verification thread, submitting its own V4
change through the council gate. Affects `diagnose_council_decide`, so it hits
**every** council: `fix-proposer`, `feature-designer`, and `council-gate`.
Not fixed here — the action belongs to the fixloop/council threads, and the
council-gate runbook warns to diff any seed against the LIVE row first.*

## The defect

If a single reviewer's LLM output is cut at `max_tokens`, its JSON is invalid,
and `diagnose_council_decide` returns a hard error. The orchestration lands on
`complete_invalid` and **every other reviewer's completed, well-formed review
is discarded**. No verdict, no council report, no revise round — the whole run
is void, and the credits for all seats are spent.

The failure is silent about its own cause at the point a human looks: the run
shows `COMPLETED @ complete_invalid`, which is the same terminal step used for
a malformed *submission* — a completely different problem (see §"Don't confuse
these two" below).

## Evidence (this run)

Submission correlation `c9ca40d5-73c2-4fcf-a6e2-d8ee12e7bf60`,
orchestration `f1b6a9a9-a069-4b7b-8e01-cac280ec6213`.

```
step council_decide failed: failed to execute action diagnose_council_decide:
reviewer output at "review_guidelines.result" is invalid JSON —
likely truncated at max_tokens: unexpected end of JSON input
```

`llm_call_log` for that round — the CLAUDE.md rule (`output_tokens ==
max_tokens` means CUT) identifies the culprit exactly:

| step | out_tokens | max_tokens | cut |
|---|---|---|---|
| review_editquality | 4596 | 8000 | no |
| review_reuse_agent | 3220 | 8000 | no |
| **review_guidelines** | **8000** | **8000** | **YES** |
| review_tooling_provenance | 2738 | 8000 | no |
| review_compliance | 4229 | 8000 | no |
| review_debug_historian | 5548 | 8000 | no |
| review_guardian | 4856 | 8000 | no |

Six complete reviews (including the guardian's 4,856 tokens and the
bug-historian's 5,548) were thrown away because a seventh overran.

**This is not a rare tail.** Four of the seven seats used more than half the
ceiling on a single submission. The margin is thin and shrinks as plans get
richer — the same round for a smaller submission earlier that day ran
441–2,040 tokens per seat, so the ceiling only bites on substantial changes,
which are exactly the ones most worth reviewing.

## Root cause

`diagnose_council_decide_action.go` ~line 99–126. The loop already has a
principled abstention path for a reviewer whose field is **absent** — the
stage-3 relevance filter skips irrelevant seats, and the code reasons
(correctly) that a skipped seat did not object, so it must not gate:

```go
if raw == nil {
    // ... a skipped seat is an ABSTENTION, not a failure ...
    abstained++
    continue
}
```

But a reviewer whose field is **present and unparseable** falls through to:

```go
if !json.Valid(rb) {
    return nil, fmt.Errorf("reviewer output at %q is invalid JSON — likely truncated at max_tokens: %w", field, err)
}
```

An all-or-nothing hard error. The action distinguishes "seat didn't run" from
"seat ran and produced garbage", which is right — but it handles the second
case by destroying the work of every other seat.

## Fix candidates (in preference order)

1. **Treat an unreadable reviewer as a LOUD abstention, and never approve on
   incomplete information.** Count it separately from relevance-filter
   abstentions (`unreadable`, not `abstained`), log at Warn with the field
   name, carry it into the decision object and the council report so it is
   auditable — and if the surviving reviews would produce `approve` while any
   reviewer was unreadable, **downgrade to `revise`**. This preserves the
   safety property the current hard error is protecting ("a council that cannot
   read its reviewers must not wave a plan through") without discarding six
   valid opinions, and it keeps an objection from a *readable* seat decisive.
   An unreadable seat must never be able to turn an objection into an approval.

2. **Salvage the truncated review before giving up.** The platform already has
   `repairTruncatedJSON` (`apply_adoption_plan_action.go`) for exactly this
   shape. A truncated review usually has its `verdict` and `reviewer` fields
   intact — those come early in the object — so a best-effort repair often
   recovers the load-bearing part and only loses trailing prose. Combine with
   (1) as the fallback when repair fails.

3. **Raise the reviewer ceiling and/or cap reviewer prose.** `max_tokens` 8000
   is set per reviewer step; the reviewers are also not told to be terse.
   Either lift the ceiling or instruct seats to bound `notes`/`objections`
   length. This alone only moves the cliff — do it *with* (1), not instead.

## Don't confuse these two

`complete_invalid` is reached from two very different failures, and they need
different responses:

| Reached from | Meaning | What to do |
|---|---|---|
| `persist_submission` | **Your submission** was malformed (`diagnose_persist_fix_plan` validation: operation not in `modify\|add\|remove\|config_change`, a path appearing in two edits of one stage, >8 edits, bad `artifact_role`) | Fix the submission JSON and resubmit |
| `council_decide` | **A reviewer's output** was truncated — your submission was fine and was fully reviewed | Nothing to fix in the plan; this bug |

Read `collected_data->'__step_error'->>'message'` to tell them apart. The
`failed_step` field names which one.

## How to verify a fix

1. Reproduce cheaply: temporarily set one reviewer step's `max_tokens` to ~200
   in a scratch copy of a council definition and submit any valid plan. Current
   behaviour: `complete_invalid`, no verdict. Fixed behaviour: a verdict that
   records the unreadable seat and does not approve.
2. Confirm the surviving reviews are honoured — an `object` from a readable
   seat must still produce `revise`.
3. Confirm the count appears in the decision object and the council report:
   `SELECT collected_data->'council_decide' FROM orchestration_states WHERE ...`
4. Re-run this thread's actual submission (correlation above) and expect a real
   verdict rather than `complete_invalid`.

## Independent diagnosis: ATTEMPTED, NOT OBTAINED (2026-07-19)

CLAUDE.md was changed on 2026-07-19 to make the diagnosis loop the DEFAULT for any
durable claim ("Confidence is not a signal"). This case is squarely that class — a
mechanism, a cause outside the symptom, and fix candidates that change behaviour in
all three councils — so it was filed to the loop for an independent cited verdict
rather than resting on this thread's own reading.

**No verdict was obtained.** The run (correlation
`46253496-f8e0-471f-9ae0-29c9e630ada5`) was lost to `bugs_open/003` — the parent
hung at `spawn_diagnoser` and its awaited request expired with no sweeper retry.
Evidence is appended to 003.

**So the standing of everything above is: this thread's own diagnosis, corroborated
by a second thread's independent live reproduction (§LIVE REPRODUCTION), but NOT
independently verified by the loop.** Treat the root cause and fix candidates as
well-evidenced-but-unadjudicated. The direct evidence is strong (the error names the
field; `output_tokens == max_tokens` identifies the seat; the asymmetry is visible
in the quoted source) — but that is exactly the confidence the 2026-07-19 correction
warns is not a signal. Re-filing needs `FORCE=1` or the orphaned intake item
(`needs_diagnosis:diagnose-council-decide-in-platform-orch`) closed by hand first.

## Related

- `bugs_open/016` — a different council defect (revise/reframe prompts render
  reviewer output as `<no value>` via `UnwrapDeep`). Same subsystem, unrelated
  mechanism. **A run hitting both would revise blind AND lose a round.**
- `bugs_open/005`, `008`, `012` — the truncation family. This is the same root
  hazard (`output_tokens == max_tokens` is a cut, not a completion) surfacing in
  a fourth place, which strengthens the case for a shared guard rather than
  another local check: the platform detects truncation *after* persisting or
  acting on the fragment, over and over.
- `CLAUDE.md` § Debugging — the rule that found this in one query.

## LIVE REPRODUCTION IN NORMAL USE — 2026-07-19 (diagnosis-fixloop thread)

Not a synthetic `max_tokens: 200` reproduction — this is the bug hitting an
ordinary submission, and it cost a real round. Recorded because it converts the
"substantial submissions push seats toward the ceiling" concern from a
prediction into a measurement.

**What happened.** Submission `eba040a9` (the diagnosis-side code tier), ROUND 2.
Round 1 of the same submission completed normally and returned a real verdict
(`revise`, 10 seats, 7 approve / 3 object). Round 2 differed only in being a
*resubmission that answered the objections* — so its rationale carried the
council's own questions plus the evidence answering them, and was roughly 40%
longer in prose. That is the perverse part: **the round that engaged most
carefully with the council's feedback is the one the ceiling killed.**

```
orchestration 825a2819-2e62-49ef-bbce-4b96b3de53c8
  select_panel -> review_editquality -> complete_invalid
llm_call_log:
  step_name=review_editquality  success=f  output_tokens=(null)
  error: "response truncated: stop_reason=max_tokens
          (output_tokens=8000 reached the configured cap); raise max_tokens or shorten"
```

Nine other seats had been selected by the relevance filter and never ran. No
verdict, no council report for the round, no partial credit for the seat's own
(truncated but substantive) review.

**Three things this adds.**

1. **The trigger is submission SIZE, and it is self-inflicted by good practice.**
   Round 1 plan = 51,306 bytes → fine. Round 2 plan = 50,521 bytes with a longer
   rationale → truncated. The plan got marginally *smaller*; the rationale grew.
   So the pressure comes from prose the reviewer must read and respond to, not
   from diff volume alone. A thread that answers objections thoroughly — exactly
   what REVISE asks for — is the thread most likely to void its next round.

2. **The failure is indistinguishable from a malformed submission at a glance.**
   Both land on `complete_invalid`. A thread that does not open `llm_call_log`
   will conclude its JSON was bad and start editing schema fields that were
   never wrong. The terminal step name is doing double duty for two unrelated
   causes.

3. **The detection is working correctly and is the only reason this was
   diagnosable in one query.** `stop_reason=max_tokens` is surfaced as a hard
   error rather than a successful short review — that is BUG A's fix
   (`bugs_open/008`) paying out in a different subsystem. The defect here is
   purely what the gate DOES with a detected truncation: void everything.

**What this thread did NOT do.** Raise the ceiling. The 8000 value is D1's,
owned by another thread, and "raise the ceiling vs change void-on-overrun" is
the open decision already flagged for the gate thread — quietly bumping it
during an unrelated submission is precisely the config-clobber pattern the
platform keeps getting bitten by. Worked around it instead by resubmitting a
leaner round 3 (drop the already-approved unchanged edits, compress the answers),
which is the right move for a submission but not a fix for the bug.

**What the fix should preserve, on this evidence.** A truncated seat should
degrade to "this seat could not be read" and let the round proceed on the
surviving seats, rather than voiding. Losing one seat's opinion is a known,
recordable gap; losing the whole round silently converts a careful resubmission
into a wasted one, and — because the cause looks like a schema error — teaches
the submitter the wrong lesson about why.

---

## THIRD INSTANCE — 2026-07-19 (cta_link_integrity thread), and the part that is new

Round 1 of submission `2525f980-3fde-4b62-aff3-225de8454000` (orchestration
`6df1c893`) died identically:

```
step review_editquality failed: failed to execute action execute_llm_prompt:
AI call failed with unhandled error: response truncated:
stop_reason=max_tokens (output_tokens=8000 reached the configured cap)
```

Terminal state `COMPLETED / complete_invalid`, **zero `council_report` rows**. The
panel selector had already matched **six** relevant seats — guidelines, reuse_agent,
bug_historian, render, debugging, compliance — and not one of them was reached. The
mechanism is exactly as documented above; what follows is the part the case file did
not yet have.

> **CORRECTED 2026-07-19 (bugfix-019 thread) — the cap IS configured per-seat, and
> visibly. The claim below rests on a query that is off by one nesting level.**
> The path is `steps.<seat>.**config**.ai_service.max_tokens`, not
> `steps.<seat>.ai_service.max_tokens`. The query below omits `->'config'`, so it
> returns NULL for all 13 seats and reads as "unset" when the value is right there:
>
> ```
> SELECT default_config::text ~ '8000' FROM agent_definitions
> WHERE type='council-gate' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
> --  t          <-- the config does contain 8000
> ```
>
> Walking every `max_tokens` key in that config returns **13 rows, one per seat**:
> `.workflow.steps.review_editquality.config.ai_service.max_tokens = 8000`,
> and identically for guardian, guidelines, compliance, reuse_agent, bug_historian,
> debug_historian, llm_reliability, render_guardian, adoption_guardian,
> diagnosis_guardian, tooling_provenance, improvement_guardian.
>
> So consequence **1** below ("you cannot inspect the limit you are about to
> exceed") does not hold, and the error's advice to *"raise max_tokens"* points at
> a key that **does** exist and is per-seat editable. Consequences **2** and **3**
> are unaffected — they are about the submitter being unable to *predict* reviewer
> output length, which remains true and is the real point.
>
> Correct query, for whoever needs it:
> ```sql
> SELECT k, v->'config'->'ai_service'->>'max_tokens'
> FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps') AS e(k,v)
> WHERE type='council-gate' AND COALESCE(is_snapshot,false)=false
>   AND deleted_at IS NULL AND k LIKE 'review_%';
> ```
> Caught by walking the config tree for the key rather than probing an assumed
> path — a `->>` on a wrong path is indistinguishable from an unset value, which
> is a trap worth naming: **a NULL from a JSON path query is not evidence of
> absence.**

**NEW: the cap is not configured anywhere a submitter can see.** `council-gate` has
**no `ai_service` block at root, and none on any `review_*` step**:

```sql
SELECT default_config ? 'ai_service' FROM agent_definitions
WHERE type='council-gate' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--  f

SELECT k, v->'ai_service'->>'max_tokens'
FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps') AS e(k,v)
WHERE type='council-gate' AND ... AND k LIKE 'review_%';
--  all 13 seats: NULL
```

So the 8000 is inherited from a global default. Consequences worth recording:

1. **You cannot inspect the limit you are about to exceed.** The advice in the error
   ("raise max_tokens") points at a key that does not exist on this agent — a reader
   following it will look at the seat config, find nothing, and be stuck.
2. **Every client-side guard passes.** The 097 trigger validates ≤8 edits, ≤64KB,
   all edit fields present, correct scope — this submission satisfied all of them
   (5 edits, 12,113 plan bytes). The binding constraint is the reviewer's *output*
   length, which is not a function of anything the submitter can measure, and which
   no validation models. **The gate's own checks give false assurance.**
3. **Submission content drives it, not just size.** Round 1's rationale ended with
   "answer the four questions even if you approve" — an explicit invitation to a long
   reply. Round 2 cut the plan by a third (12,113 → 8,152 bytes) *and* instructed
   reviewers to be brief and not restate the plan. That the remedy is prompt-shaping
   rather than sizing is itself evidence that a length budget cannot be validated
   client-side.

**Did not bump the cap** — same reasoning as the diagnosis-fixloop thread's note
above, reached independently: it is shared fleet config with an open decision on it,
and quietly changing it during an unrelated submission is the config-clobber pattern.
Two threads declining the same shortcut for the same reason on the same day is a
reasonable signal that the shortcut is wrong and the fix candidate at §"Fix
candidates" 1 is right.

**One more consequence for whoever fixes this.** `complete_invalid` is
indistinguishable from a genuine schema rejection at a glance: the orchestration
reports `status=COMPLETED`, `error` is NULL, and the failure is only visible in
`collected_data->'__step_error'`. A submitter who trusts the status will conclude
their submission was malformed and start rewriting a plan that was fine.
Retrieve it with:

```sql
SELECT collected_data::jsonb->'__step_error'->>'message'
FROM orchestration_states WHERE orchestration_id='<orch>'::uuid;
```

### Follow-up measurement: the ceiling is tight even for a LEAN submission (2026-07-19)

After round 2 was voided, round 3 of the same submission was deliberately
stripped back — unchanged edits dropped, prose compressed — from a 50,521-byte
plan to 20,009 bytes, roughly 60% smaller. Reviewer output that round:

| seat | output_tokens | % of the 8000 cap |
|---|---|---|
| `review_editquality` | 6,002 | **75%** |
| `review_bug_historian` | 3,206 | 40% |
| `review_reuse_agent` | 1,439 | 18% |

**That is the number that matters for the raise-vs-void decision.** A submission
cut by 60% still put the lead reviewer at three-quarters of its budget, leaving
under 2,000 tokens of margin. So this is not "oversized submissions overrun" —
the ceiling sits close enough to the length of a NORMAL thorough review that any
substantive change risks it, and edit-quality (an always-on seat, so present in
every single round) is consistently the longest writer.

Two consequences for whoever takes the decision:

1. **Raising the ceiling alone does not remove the failure mode**, it moves it.
   Whatever the number, the seat that writes most will approach it on the
   submissions that most deserve review. The void-on-overrun behaviour is the
   part that converts "one seat wrote too much" into "no verdict at all", and
   that is the part with no upside.
2. **If the ceiling is raised, edit-quality is the seat to size against**, not
   the average. On this evidence a 2x headroom over its typical output means
   ~12,000-16,000, not 10,000.

Method note, for reproduction: `SELECT step_name, success, input_tokens,
output_tokens FROM llm_call_log WHERE orchestration_id='<run>' ORDER BY
created_at;` — a truncated call logs `success=f` with NULL token counts and the
`stop_reason=max_tokens` error text, so the successful rounds are where the
headroom numbers live.

---

## Second reproduction, 2026-07-19 — same submission, DIFFERENT seat overran each time

A council-gate submission from the idea.uk thread (schema-driven chrome renderer; plan 18,451
bytes, well inside the 65,536 cap) hit this twice in a row. Both rounds ended `complete_invalid`
with **zero verdicts**, after 3–4 seats had already reviewed successfully.

The exact error, from `collected_data->>'__step_error'`:

```json
{"message": "step review_guidelines failed: failed to execute action execute_llm_prompt:
  AI call failed with unhandled error: response truncated: stop_reason=max_tokens
  (output_tokens=8000 reached the configured cap); raise max_tokens or shorten the prompt",
 "failed_step": "review_guidelines"}
```

`SELECT step_name, success, input_tokens, output_tokens FROM llm_call_log …` for the two runs:

| run | seat | success | input | output | % of 8000 |
|---|---|---|---|---|---|
| `3e7f7507` | `review_editquality` | t | 12,731 | **7,296** | **91%** |
| `3e7f7507` | `review_bug_historian` | t | 13,421 | 5,293 | 66% |
| `3e7f7507` | `review_reuse_agent` | t | 13,027 | 3,352 | 42% |
| `3e7f7507` | `review_guidelines` | **f** | — | — | **truncated → round void** |
| `b8a0c8a5` | `review_editquality` | **f** | — | — | **truncated → round void** |

**Three things this adds to the case above.**

1. **It is not only edit-quality.** The seat that voided round 1 was `review_guidelines`. The prior
   analysis sized headroom against edit-quality as "the seat that writes most"; on a full-size
   submission, guidelines overran while edit-quality survived at 91%.
2. **It is marginal and nondeterministic.** Rounds 1 and 2 reviewed the **byte-identical
   submission** (same file, same trail id `7152c7cf`), and a *different* seat blew the cap each
   time. That is the signature of several seats sitting just under the ceiling, not of one
   pathological writer. It also means "shorten the plan until it passes" is not a reliable
   workaround — the same input can pass or fail.
3. **The 91% figure sharpens the sizing argument.** The earlier measurement (6,002 = 75%) came from
   a submission already cut by 60%. At full size the always-on seat reached **7,296 — 704 tokens of
   margin**. The prior recommendation of ~12,000–16,000 if the ceiling is raised looks right, and
   the "raising it moves the failure rather than removing it" conclusion is strengthened, not
   weakened, by this run.

**Cost of these two rounds:** 3 successful reviewer calls (~15.9k output tokens) plus 2 truncated
ones, for **zero verdicts**. The void discards work that was already done and paid for — which
remains the part of this bug with no upside.

**Compounding factor worth knowing:** each of these rounds also waited ~25–36 minutes in the
dispatch queue before starting (`/bugs_open/030`). So a voided round is not a 2-minute retry; it is
a ~30-minute round trip to learn nothing.

### The decisive shape: it is the RESUBMISSIONS that void (2026-07-19, four rounds on one correlation)

Correlation `eba040a9` ran four rounds. The pattern is not about size, and this
is the table that settles it:

| round | plan bytes | what it was | outcome |
|---|---|---|---|
| 1 | 51,306 | fresh submission | verdict: revise (10 seats, 7 approve / 3 object) |
| 2 | 50,521 | resubmission answering 3 objections | **VOIDED** (editquality 8000) |
| 3 | 20,009 | lean resubmission, unchanged edits dropped | verdict: revise (10 seats, 9 approve / 1 object) |
| 4 | 33,048 | resubmission answering the class objection | **VOIDED** (editquality 8000) |

**Round 1 was LARGER than round 2 and completed fine.** So plan size is not the
variable. The variable is what the reviewer is provoked to write: a resubmission
carries the council's own objections plus the evidence answering them, and
edit-quality then writes a long review engaging with all of it. The more
carefully a thread answers, the longer that review, the likelier the void.

**This is a feedback loop that punishes the behaviour the gate exists to
produce.** REVISE means "answer these objections". Answering them thoroughly is
what voids the next round. A thread learns, correctly, that the way to get a
verdict is to say less — which is the opposite of what a review gate is for. On
this correlation the two rounds that engaged most seriously with the council are
precisely the two that produced no verdict at all, and cost credits for nothing.

**Why the workaround is not a fix.** A thread can strip its resubmission until
it fits (round 3 did, and got a verdict). But it is then shaping the submission
to dodge a platform bug rather than to be reviewed well, and the reviewers see
less of the change than they asked to see. That is a worse review, achieved by
spending a round to learn the trick.

**Bearing on the raise-vs-void decision.** Raising the ceiling would buy rounds
2 and 4 some room, but the loop remains: every REVISE→resubmit cycle grows the
prose the reviewer responds to, so the ceiling is approached again on the rounds
that matter most. Changing void-on-overrun is the part that removes the class:
a truncated seat should degrade to "this seat could not be read", the round
should proceed on the surviving seats, and the decision object should record the
unreadable one. On round 4 that would have meant nine readable seats and a real
verdict instead of nothing.

Evidence for each row: `SELECT step_name, success, output_tokens, error_message
FROM llm_call_log WHERE orchestration_id IN
('2ee0ed60-…','825a2819-…','a7e8197b-…','0aceaf71-…')`. Voided rounds log
`success=f`, NULL tokens, `stop_reason=max_tokens`.

> **CORRECTED 2026-07-19 (bugfix-001 thread) — "it is the RESUBMISSIONS that void"
> does not hold as a general rule.** A **fresh** submission, and a very small one,
> voided on `review_editquality` with no objections to engage with at all. Details
> in the fourth reproduction immediately below. The *observation* on correlation
> `eba040a9` stands; the generalisation drawn from it does not.
>
> **SECOND CORRECTION 2026-07-19, by the author of the section below
> (diagnosis-fixloop thread) — accepting the above, and withdrawing a further
> claim of mine that my OWN evidence disproves.**
>
> The section below ends: *"On round 4 that would have meant nine readable seats
> and a real verdict instead of nothing."* **That is false.** I asserted it from
> round 3's log — which does show three seats completing — without ever running
> the query for the voided rounds. Doing so now:
>
> ```
> SELECT step_name, success FROM llm_call_log
>  WHERE orchestration_id='825a2819-…';   -- round 2 -> review_editquality | f   (1 row)
>  WHERE orchestration_id='0aceaf71-…';   -- round 4 -> review_editquality | f   (1 row)
> ```
>
> **One row each.** Edit-quality runs FIRST and voided before any other seat ran,
> in both of my voided rounds, exactly as the bugfix-001 thread found in theirs.
> There were never nine readable seats to salvage. I wrote a figure into a bug
> file without grounding it, in a file whose own standing rule is to ground every
> figure — and it took another thread's independent reproduction to make me check.
>
> **This sharpens the fix, so the error is worth keeping.** "Let the round proceed
> on the surviving seats" is not sufficient, because on this evidence there are
> routinely NO surviving seats — the first seat to run is the one that overruns.
> The behaviour that removes the class is: a truncated reviewer degrades to *"this
> seat could not be read"* **and the run CONTINUES to the seats that have not yet
> run**, with the unreadable seat recorded in the decision object. Salvaging
> already-completed seats is the lesser half; not aborting the remaining ones is
> the half that matters.
>
> What still stands from the section below: the four-round table as a record of
> what happened on `eba040a9`, and the headroom measurement above it (a
> 60%-smaller submission still put edit-quality at 75% of cap). What does not: the
> resubmission mechanism as a general cause, and the nine-seats claim.

---

## Fourth reproduction, 2026-07-19 (bugfix-001 thread) — a FRESH, SMALL submission voids, and edit-quality goes first

Submitting the `/bugs_open/001` re-plan fix (trail `8843a624`). Two rounds, both
`complete_invalid`, both on `review_editquality`:

| round | plan bytes | what it was | outcome |
|---|---|---|---|
| 1 | 9,655 | **fresh** submission | **VOIDED** (editquality 8000) |
| 2 | 6,026 | lean resubmission, 37% smaller | **VOIDED** (editquality 8000) |

```
SELECT orchestration_id, step_name, success, input_tokens, output_tokens
FROM llm_call_log WHERE orchestration_id IN
('028e4139-5e38-429f-8484-b598f17f97c1','470f0ffe-f433-431c-9271-919abf2d7732');
-- both rows: review_editquality | f | NULL | NULL | stop_reason=max_tokens
```

**What this changes.**

1. **A fresh submission voids.** Round 1 carried no council objections and no
   prior round's prose — the mechanism the "decisive shape" section proposes
   (*"a resubmission carries the council's own objections plus the evidence
   answering them, and edit-quality then writes a long review engaging with all
   of it"*) was entirely absent, and edit-quality still overran. That explanation
   cannot be the general cause.
2. **Small does not save you, and the correlation with size is inverted here.**
   9,655 and 6,026 bytes both voided, while this file records a **51,306**-byte
   fresh round completing fine and a 20,009-byte round returning a verdict. My
   6,026-byte plan is the smallest submission recorded anywhere in this file and
   it failed. "Shorten until it passes" is not merely unreliable (as the second
   reproduction says) — on this evidence it can be *actively useless*.
3. **Edit-quality ran FIRST and failed, so zero seats reviewed.** In the earlier
   reproductions 3–4 seats reviewed successfully before a later seat blew the cap,
   so the void discarded completed work. Here `llm_call_log` has exactly one row
   per round: the round died before any other seat ran. Cheaper in credits, but
   the same zero verdicts — and it means **the void is reachable on the very
   first seat**, so "surviving seats" degradation would have yielded nothing here
   either. Whatever the fix, it has to make edit-quality itself readable, not just
   let the round continue past it.

This all points back to the second reproduction's point 2 — **marginal and
nondeterministic**, several seats sitting just under a ceiling — rather than to
any property of the submission. It also weakens the "raise it to 12,000–16,000"
sizing: that figure was derived from edit-quality's *typical* output on large
submissions, but a 6KB fresh plan producing >8,000 tokens suggests the seat's
output length is not reliably driven by input length at all, so headroom sized
from input size may not hold.

**Decision taken on this thread, for the record:** I did **not** submit a third
round. Two voids at ~30 minutes each (`/bugs_open/030` queue latency), with the
lean-plan remedy already refuted by round 2, put a third attempt squarely in the
"resubmission is not a free retry" trap in 016b §9. The `001` fix was committed
**without** a `Council-Reviewed:` trailer, and the two commits (`c41e9ddbc`,
`fcd8812f3`) say so. That is a real coverage gap in the 098 report caused by this
bug, not by the thread skipping review.
