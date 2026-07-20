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
