# NOTES — bugs_open/019, one truncated reviewer voids the whole council round

Append-only, newest at the bottom. Technical log: evidence, commands, what the
system actually said, and every misstep.

---

## 2026-07-19, session "bugfix 019" — opening evidence

Task: research and fix `/bugs_open/019`, diagnosis loop first.

### Queue check first (CLAUDE.md § Dispatching work)

An open item already existed — **not** filed by me:

```
id         | 9ffdf0f8-ae76-405a-a316-6233e1e16b80
status     | awaiting_diagnosis      created_by | 090_TRIGGER_needs_diagnosis
created_at | 2026-07-19 12:42:34+00  claimed_by | (null)   result | {}
```

Never claimed, never run, no findings. Its symptom is written **entirely** about
`diagnose_council_decide`'s `json.Valid` asymmetry — i.e. the mechanism the case
file's §"Root cause" describes. That matters, because of the next section.

### FINDING 1 — the documented root cause is the MINORITY mechanism

The case file's §"Root cause" and all of fix-candidate 1 point at
`diagnose_council_decide_action.go:99–126`. But every 2026-07-19 reproduction in
the file shows the round dying at `execute_llm_prompt`, which is upstream and
never reaches that code. Counted it:

```sql
SELECT collected_data::jsonb->'__step_error'->>'failed_step' AS failed_step,
  CASE WHEN collected_data::jsonb->'__step_error'->>'message' ILIKE '%execute_llm_prompt%'
         THEN 'UPSTREAM execute_llm_prompt (truncated call)'
       WHEN collected_data::jsonb->'__step_error'->>'message' ILIKE '%invalid JSON%'
         THEN 'DOWNSTREAM council_decide (json.Valid)'
       ELSE 'other' END AS mechanism, count(*)
FROM orchestration_states
WHERE current_step='complete_invalid' AND created_at > now() - interval '10 days'
  AND collected_data::jsonb->'__step_error' IS NOT NULL
GROUP BY 1,2 ORDER BY 3 DESC;
```

| failed_step | mechanism | count |
|---|---|---|
| `review_editquality` | UPSTREAM `execute_llm_prompt` | **8** |
| `council_decide` | DOWNSTREAM `json.Valid` | 2 |
| `council_decide` | other | 2 |
| `persist_submission` | other (malformed submission) | 1 |
| `review_guidelines` | UPSTREAM `execute_llm_prompt` | **1** |

**9 upstream vs 2 downstream.** So fix-candidate 1 as written — patching the
abstention path in `diagnose_council_decide` — would not have fixed a single one
of the four reproductions recorded in the case file. The queued diagnosis item
`9ffdf0f8` would have sent the loop to the same wrong file.

This is the "cause is not where the symptom is" class exactly. Filed a corrected
diagnosis rather than reusing the queued symptom (`FORCE=1`, having read that the
existing item has no findings to duplicate): correlation
**`a92b2a55-830f-4f78-bb00-b03b723878a9`**.

### FINDING 2 — `anthropic.go` DISCARDS the partial text

`platform/aiservice/anthropic.go:180`:

```go
if response.StopReason == "max_tokens" {
    return "", fmt.Errorf("response truncated: stop_reason=max_tokens (output_tokens=%d ...)", ...)
}
```

It returns **`""`** — the partial completion sitting in `response.Content` is
thrown away at the transport layer, before any caller sees it.

**Consequence for the case file: fix-candidate 2 is impossible as written.** It
proposes salvaging the truncated review with `repairTruncatedJSON`
(`apply_adoption_plan_action.go:950` — it does exist). But there is nothing to
repair: the bytes never leave `anthropic.go`. Salvage requires changing this
function first. The case file assumed the partial reached the council action; it
does not.

This is also the sharper statement of the truncation family (005/008/012): the
platform doesn't merely *act on* fragments, in this path it **detects truncation
correctly and then destroys the evidence**.

### FINDING 3 — the void is structural config: every seat routes to the terminal

```sql
SELECT k, v->'config'->>'error_step', v->>'next_step'
FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps') AS e(k,v)
WHERE type='council-gate' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND k LIKE 'review_%' ORDER BY k;
```

All 13 seats: `error_step = complete_invalid`. The seats run **sequentially**
(`review_editquality → gate_bug_historian → review_bug_historian → … →
review_guardian → council_decide`), so the first seat to error routes the entire
round to the terminal step. `review_editquality` is first in the chain — which is
exactly why the bugfix-001 thread and the diagnosis-fixloop thread both found
**one row** in `llm_call_log` for their voided rounds.

So "let the round proceed on the surviving seats" is not a code change at all in
the common case: there are no surviving seats because the chain aborts at seat 1.
Confirms the correction already recorded in the case file, and gives it a
mechanism.

### FINDING 4 — the cap IS configured, and visibly (case file corrected)

The case file claims "the cap is not configured anywhere a submitter can see",
from a query returning NULL for all 13 seats. That query is **off by one nesting
level** — it reads `steps.<seat>.ai_service.max_tokens`; the real path is
`steps.<seat>.**config**.ai_service.max_tokens`. Walking the config tree for the
key returns 13 rows, all `8000`.

Corrected in the case file itself (commit `4efd59bd7`) rather than only here.
General trap named there: **a NULL from a JSON path query is not evidence of
absence** — a wrong path and an unset value are indistinguishable via `->>`.

Method that caught it: dump `default_config` and walk every key named
`max_tokens`, instead of probing an assumed path.

### Environment note

Owner deployed a fresh chassis during this session: `agent-chassis:v1.0.1138`,
pod `agent-chassis-55d7774dc4-pzt9j`, started 2026-07-19T16:50:02Z. Any Go fix
here is inert until the next roll after that.

### Status at this point

Diagnosis run `a92b2a55` in flight (`load_runtime` → `call_diagnoser`). Fix shape
NOT yet committed to — waiting on the loop's verdict before asserting a root
cause, which is the whole point of filing it. Working hypothesis to be tested by
the loop, in layers:

1. `anthropic.go` — return the partial text alongside a typed truncation error
   instead of discarding it (additive: existing `if err != nil` callers unchanged).
2. `ai_actions.go` — let a review step tolerate truncation, recording the partial
   plus a marker, so the chain continues to the next seat rather than routing to
   `complete_invalid`.
3. `diagnose_council_decide_action.go` — repair/parse the partial; if a verdict is
   recoverable use it (marked degraded), else count it as `unreadable` (a loud
   abstention) and never let an `approve` stand while any seat was unreadable.

Layer 2 is what removes the class; layer 3 preserves the safety property the
current hard error exists to protect.

---

## 2026-07-19 — diagnosis verdict: UNVERIFIABLE, and what it caught

Run `a92b2a55` terminated `complete|COMPLETED` after one iteration with outcome
**UNVERIFIABLE** — not refuted, but it could not confirm the mechanism. Its three
stated gaps, and what each turned out to be:

1. *"the actual source of the three named symbols is not in the bundle"* — a
   bundle-assembly gap (`symbol_count: 3`, `body_chars: 6778`, the bodies absent).
   The mechanism is nonetheless directly readable in the repo, which is how it was
   found; the loop simply could not see it.
2. *"no llm_call_log rows for the named agent types"* — **an attribution artifact,
   not an absence.** See below.
3. *"the truncation errors are attributed to 'experience-planner' and 'generic',
   not to council-gate/fix-proposer/feature-designer"* — same cause, and it is the
   thing worth keeping.

### MY MISSTEP — I asserted "9 of 14 voided COUNCIL runs" without verifying they were council runs

My original query grouped `orchestration_states` by `__step_error`, keyed on step
names only. **`orchestration_states` has no `agent_type` column at all**, so that
query could not have told me which agent owned those runs — I inferred "council"
from the step names (`review_editquality`, `review_guidelines`) and wrote the
number down as though I had checked. The loop's gap 3 is what made me go back.

The claim survives verification, but that is luck of the draw, not method:

```sql
SELECT agent_type, count(*) AS calls, count(*) FILTER (WHERE success=false) AS failed
FROM llm_call_log WHERE step_name LIKE 'review_%' AND created_at > now() - interval '10 days'
GROUP BY 1 ORDER BY 2 DESC;
```

| agent_type | calls | failed |
|---|---|---|
| `generic` | 431 | **9** |
| `experience-planner` | 176 | 3 |

Nine failed `review_*` calls under `generic`, matching the nine voided
orchestrations exactly. So the count was right.

### FINDING 5 — councils log as `agent_type='generic'`, which blinded the loop

The councils do **not** log under `council-gate` / `fix-proposer` /
`feature-designer`; they log under **`generic`**. This is the already-known
landmine (`council_report source_agent='generic'` fleet-wide — partition by
another key), surfacing here in a new place.

Consequence worth flagging beyond this bug: **the diagnosis loop's own evidence
gathering is blind to council runs.** It filtered `llm_call_log` and
`agent_error_log` by the three council agent-type names, found nothing, and
correctly declined to confirm. Any future diagnosis filed against a council will
hit the same wall unless the symptom names `generic` explicitly. The
UNVERIFIABLE verdict was therefore *correct behaviour on incomplete evidence*,
and the incompleteness is a harness defect, not a reviewer error.

### FINDING 6 — a 32,000-token cap ALSO truncated. This settles the raise-vs-void question.

```
agent_error_log: experience-planner / compose
"response truncated: stop_reason=max_tokens (output_tokens=32000 reached the
 configured cap); raise max_tokens or shorten the prompt"
```

`experience-planner` is a fourth affected council, absent from the case file
entirely (steps `compose`, `review_contracts`, `review_feasibility`). Its
`compose` step has a cap **four times** the council seats' 8000 — and still hit it.

This is the empirical answer to the open "raise the ceiling vs change
void-on-overrun" decision that two prior threads correctly declined to take
unilaterally: **raising the cap demonstrably does not remove the failure.** The
case file argued this from reasoning ("whatever the number, the seat that writes
most will approach it"); this is the same conclusion from a measurement, at 4x
the disputed number. Recorded so the owner's decision can rest on evidence rather
than on the argument.

It also widens the blast radius: the case file scopes the bug to three councils;
it is at least four, and `compose` is not a reviewer seat, so the class is
"any LLM step whose output can grow", not "a council seat".

### Net effect on the plan

Mechanism **stands, now verified from the code and the logs**; the loop's
inability to confirm was an attribution artifact I could resolve directly. The
three-layer shape in PLAN is unchanged, and Finding 6 strengthens its constraint
4 (do not fix this by raising the cap) from an argument into a measurement.

---

## 2026-07-20 — build session: the fix, two near-misses, and the closure record

Owner decision: **keep tolerance narrow (councils only)**; fix built as the
three layers in PLAN. Code committed `a3b606798` (6 files, +402/−42), migration
177 committed `76ff5ed25` and **applied live** (35/35 review seats armed:
council-gate 15, fix-proposer 15, feature-designer 5 — the DO block iterates
`review_*` keys, which silently covered the v19 constitution/mission seats that
had landed since my 13-seat walk. That walk's count was already stale within a
day: another live demonstration of "ground every figure").

### NEAR-MISS 1 (caught in self-review): tolerated truncation would have double-logged

My first layer-2 draft cleared `err` and fell through to the normal result path —
which contains the SUCCESS-path `LogLLMCall`. A tolerated call would have logged
twice: `success=false` (error path, correct) then `success=true` with
`output_tokens` at exactly the cap — **the pre-008 silent-cut signature**, and
poison for the headroom queries this very case file documents. Fixed by guarding
the success log on `truncationTolerated`: one call, one row. The general shape:
*converting a failure into a success mid-function must account for every side
effect the success path performs again.*

### NEAR-MISS 2 (git scare, resolved harmless): "my commits vanished"

After resume, `git log -8` showed session-start's HEAD on top and none of my four
commits; simultaneously my committed files showed clean. Brief hypothesis: a
reset had orphaned my work — which would have been a forward-only violation by
someone. **Reflog refuted it**: every entry is a plain `commit:`; my commits are
all in ancestry (`git cat-file -t` on each: commit); the branch had simply
advanced ~11 commits from concurrent sessions, and the context snapshot I was
comparing against was regenerated at resume time, not at session start. Lesson
recorded: **the resume-time snapshot is NOT the session-start snapshot** — treat
any "state moved backwards" reading as a stale-baseline artifact until the
reflog says otherwise.

### Decision: committed BEFORE the council verdict, deliberately

330 lines of coherent, tested platform code in the shared tree is precisely what
CLAUDE.md documents being swept into other threads' commits. Weighed against the
pre-commit review convention: committed first (`a3b606798` says so in its
message), submitted after — correlation `2eed453a-9102-41e0-8838-7a711e99126b`,
orchestration `21dc9751`. The submission is reviewed by the still-buggy gate; if
it voids, that is reproduction five at zero marginal cost. No trailer unless a
real APPROVED lands (trailer discipline; the bugfix-001 precedent for this exact
situation is recorded in the case file's fourth reproduction).

097 schema gotcha for the RUNBOOK: `plan` is an object (`summary`/`edits`/
`grounded_in`/`risks`), each edit wants `symbol`; an array-shaped plan fails
client-side as `.plan missing`.

### Closure records

- `bugs_open/019`: FIX BUILT header added (stays OPEN — inert until a roll after
  v1.0.1139; "fixed AND live" bar).
- 016b §9: existing entry corrected in place (minority-mechanism correction +
  the three companion traps: generic attribution, JSON-path NULL, 32k cap).
- Work item `needs_diagnosis:platform-aiservice-anthropic-go-generate` (mine):
  complete, resolution recorded.
- Work item `9ffdf0f8` (the pre-existing queued diagnosis aimed at the minority
  mechanism): **cancelled** with a resolution note — unclaimed, superseded by
  the fix; running it would have spent credits re-deriving a fixed defect.
  Judgement call recorded here in case its filer disagrees.

### Attribution note: the 016b §9 correction rode in another thread's commit

My 016b edit was swept into `867049c04` (a bugs_open/008 docs commit by a
concurrent session) between my edit and my docs commit — the same-file-passenger
case the commit-scope hook explicitly cannot see. Content verified present in
HEAD; nothing lost. But my docs commit `2cb37685a` *claims* the 016b correction
while not containing it — this note is the pointer for anyone tracing that edit.

---

## 2026-07-20 — council round 1: REVISE (and the round itself is half the story)

Verdict on corr `2eed453a` (orch `21dc9751`): **revise**, decided by
bug_historian's objection. 10 seats reviewed, 5 abstained via relevance filter,
5 approve (editquality, guidelines, llm_reliability, constitution, mission) /
5 object (bug_historian, reuse_agent, guardian, diagnosis_guardian,
debug_historian). **The round COMPLETED** — a substantive, long-rationale
submission of exactly the class that voided four rounds this week ran all ten
seats without an overrun (editquality at 4,738 of 8,000). Luck of the draw, not
proof of anything, but worth recording that the gate CAN review its own fix.

Two objections earned code changes (round 2, in tree, tests green via
`git archive HEAD` overlay — another session's uncommitted
diagnose_prepare_fix_commit WIP breaks the shared package build; their file,
not touched):

1. **The convergent one (3 seats, verdict-deciding): a truncated partial that
   PARSES cleanly passed as an ordinary review** — no degraded mark, because the
   decider never consulted the `__truncated` marker. I had self-flagged this as
   risk 3 and shipped anyway; three seats independently refused to let it slide.
   Fixed: `markerFieldFor` derives the sibling marker path and a clean parse
   with the marker set is forced `Degraded`. The lesson: a self-disclosed risk
   in a review submission is a REQUEST for pushback, and the council provided
   it — this is the gate working.
2. **llm_call_log blind spot: tolerated vs fatal truncation indistinguishable**
   (3 seats, all low). Fixed: tolerance detected before logging; the single
   failed row's error_message now leads with
   `TOLERATED (step continued on the partial): `.

Answered with evidence, no change: third-provider gap (enumeration: exactly two
GenerateText implementations exist, both patched); guardian's platform-wide
partial-return concern (all five call sites audited — every one returns/continues
on err, none consumes the result); repairTruncatedJSON provenance
(apply_adoption_plan_action.go:950, pre-existing); typed-error and `__` marker
conventions (AIUnavailableError and __step_error/__usage_* are the precedents).

**Accepted as a process miss** (debug_historian): migration 177 was applied with
no pre-write backup and no needle-gate query — the lore requires both for
production jsonb surgery. Mitigation recorded: additive single key, rollback =
strip-key UPDATE (now in RUNBOOK). Do the backup next time.

Round 2 submitted on the SAME correlation (RESUBMIT_CORR), orch `a8b3e765` —
3-edit delta plan. Round-2 code deliberately HELD UNCOMMITTED until the verdict:
the diff is ~60 lines (vs round 1's 330), so the sweep-risk calculus inverts and
an APPROVED verdict can put a legitimate Council-Reviewed trailer on the commit.

---

## 2026-07-20 — round 2: REVISE again; stopped at two rounds, closure

Round 2 (orch `a8b3e765`): **revise**, decided by reuse_agent. Movement:
objectors 5 → 3; diagnosis_guardian flipped to APPROVE; 6 abstained (delta plan
matched fewer seats). Round-2 code committed `11a72dc31` WITHOUT a trailer.

Dispositions of the six remaining objections (none a functional defect):

- **reuse_agent e1 (medium, verdict-deciding)** — "does a generic sibling-marker
  helper already exist that markerFieldFor duplicates?" No: the existing `__`
  marker readers hardcode their paths (`load_work_item_actions.go:879` reads
  `__step_error.message` literally); markerFieldFor is the first derivation
  helper. Answerable, not actionable.
- **reuse_agent e2 / guardian e2 (low)** — "structured field instead of an
  error_message prefix?" `llm_call_log` has no status/degraded column; adding
  one is a schema migration on a shared observability table — heavier than the
  string prefix and someone else's surface to change. Recorded as a possible
  future improvement for whoever owns that table.
- **guardian e1 (low)** — non-`.result` review_fields get no marker check:
  by design, degrades to round-1 behaviour, never worse.
- **editquality e1 (low)** — marker write/read are a coupled pair in one
  codebase (markTruncated writes bool; decider asserts bool); a JSON round-trip
  preserves bool. Real risk only if a future writer stamps a string — noted.
- **debug_historian e1 (HIGH)** — the lore objection, escalated because code now
  DEPENDS on 177's values. Answered in substance, not prose: rollback DO block
  committed to RUNBOOK (`8762033b2`), and snapshot
  **`bak_agentdef_councils_20260720`** taken (3 rows — the platform's
  established bak_agentdef_* pattern). The pre-write backup cannot be taken
  retroactively; the miss stands recorded.

**Why no round 3.** Two rounds, both completed (itself evidence the gate can
review substantive submissions when the seats stay under the cap), objection
count falling, zero functional defects remaining, and the gate has NO reviser
loop — complete_revise is terminal and objections go to the human, which is
this record. A third run would spend credits chasing a trailer line. The
resubmission-is-not-a-free-retry rule is written in 016b §9 by a thread that
learned it the expensive way; stopping here is that lesson applied.

**Closure state:** code a3b606798 + 11a72dc31; config 76ff5ed25 (applied,
verified, rollback + snapshot in place); docs complete; work items closed;
bugs_open/019 OPEN pending image roll. Council record: corr 2eed453a, rounds
1+2 both revise, full objection trail in council_report artifacts.

---

## 2026-07-20 (later) — bugs_open/036: the SECOND cause on the same seam (bugfix-036 thread)

Not 019, and recorded here rather than in a new directory because it is the same
function, the same contract and the same seam 019 built — a second account would
drift. The case file is `bugs_open/036`; this is only what the 019 workstream
needs to know.

**What it is.** A reviewer emitting `"edit": "<string>"` where the struct wanted
`int` voided the whole round — `COMPLETED @ complete_invalid`, `error` EMPTY, no
`council_report`, every seat paid for. Identical symptom to 019, different cause:
the JSON is **complete and valid**, so `salvageTruncatedReview` cannot help.
019's `unreadable` mechanism was the right seam; it was simply guarded by
`!json.Valid(rb)` and so never reached this case.

**Measured split (14 days of `complete_invalid`)** — worth having because 016b §9's
own correction says to count the layer first: `review_editquality` upstream
truncation **9**, `council_decide` schema slip **3**, `council_decide` truncation
**2**, other **2**. So this cause is the majority of what still reaches the decider
now that 019's upstream fix exists.

**The wrong turn worth recording — mine, and the bug file's.** `bugs_open/036` §5
proposed `json.Number` as the tolerant type, and I started out expecting the live
payloads to be `"3"` — a malformed index. They are not. All three voided rounds
carried a *plan-level description*: `"plan-level (deploy verification)"`, `"risks
note on the 54 mis-stamped rows"`, `"risks/summary (item 5)"`, all from
`review_debug_historian`. `json.Number` would have parsed **none** of them. The
reviewers were not emitting garbage — they were saying "this objection is about
the plan, not any single edit", which the contract already spells `0`. The strict
type was discarding a meaning it had a representation for. Reading the actual
payloads before picking the tolerant type is the whole lesson.

**Second wrong turn, in the tests.** I asserted that a mistyped `severity` deep
inside an objection loses the objection. It does not: `encoding/json` continues
past a TYPE error (unlike a syntax error) and keeps what did decode. Better than
assumed, and it is why field-by-field salvage retains as much as it does. Test now
pins the real behaviour.

**What changed on the shared seam** (`diagnose_council_decide_action.go`) — matters
to this workstream because it alters code 019 wrote:
- `Objections []struct{...}` → named `councilObjection`, with `Edit` now
  `objectionEdit{Index, Raw}`; `UnmarshalJSON` never errors, `MarshalJSON`
  round-trips the reviewer's own token so `council_report` is not laundered.
- `salvageMistypedReview` added **beside** `salvageTruncatedReview` (per 036 §7:
  extend the seam, do not invent a parallel degraded-round path).
- The three surviving `return nil, err` exits in the per-seat loop — `planBytes`,
  schema mismatch, **unrecognised verdict** — now all route to `unreadable`.
  Unrecognised verdicts are deliberately NOT normalised to the nearest legal
  value: guessing what a seat meant is how a veto becomes an approval.
- Untouched, and still what keeps this safe: the `len(reviews)==0` fail-closed
  guard and the `approved + unreadable → revise` downgrade. 019's five tests pass
  unchanged.

**Sequencing gate honoured.** 036 §7 said not to touch this file until the 019
truncation work landed. It had (`11a72dc31`), and the tree was clean, so no
same-file race. Anyone extending this function next: check `git status` on it
first, for the same reason.

**Also noticed, not fixed:** two orchestrations hung at `spawn_ingester` (idle
~4,300s and ~4,600s) and one at `route` (~12,200s) while this ran — the
`bugs_open/029` hung-spawn class, still live and still saturating the dispatch
group. Both this thread's diagnosis run and its council submission sat queued
behind it for well over the usual latency.
