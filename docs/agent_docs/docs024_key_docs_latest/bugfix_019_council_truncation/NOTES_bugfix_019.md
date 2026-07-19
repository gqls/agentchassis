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
