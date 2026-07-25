# 076 — a truncated LLM response is tolerated fleet-wide; only 5 of 58 agents check the marker

**Filed:** 2026-07-25 by the gripper-dossier thread (robot_hands_gripper_dossier),
prompted by a council objection (seat `bug_historian`, correlation
`7ed137d1-361c-4f69-9361-9e4ba1dfa6bf`, round 2) that asked the one question I
had not: *how many other call sites have this exposure?*
**Status:** OPEN, not started. Filed as a case, deliberately NOT fixed inside
the feature batch that found it.

---

## The mechanism

`ExecuteLLMPromptAction` (`platform/orchestration/actions/ai_actions.go`) does
not fail when a model response is cut at `max_tokens`. It **tolerates** the
truncation, stamps a marker beside the result, and returns **success**:

```go
out := map[string]interface{}{"result": parsedResult, "type": "json"}
markTruncated(out, truncationTolerated, truncatedTokens)   // sets out["__truncated"] = true
return out, nil
```

`markTruncated`'s own comment states the contract plainly, and it is a sound one:

> *"The marker is the whole point of tolerating a truncation rather than
> failing: the step now SUCCEEDS, so without it a consumer cannot distinguish a
> complete answer from a fragment, and bugs_open/019 would trade a loud void for
> a silent half-answer — a strictly worse bug."*

The contract is therefore **consumer-enforced**: tolerating is safe *only if the
consumer reads the marker*. Nothing checks that consumers do.

## The measurement (live, 2026-07-25)

```sql
-- execute_llm_prompt steps across active, non-snapshot agents
SELECT count(*) AS llm_steps, count(DISTINCT d.type) AS agents
FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') AS e(k,v)
WHERE d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND v->>'action'='execute_llm_prompt';
--  118 |  58

-- agent definitions that mention the marker AT ALL
SELECT count(*) FROM agent_definitions d
WHERE d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND d.default_config::text LIKE '%__truncated%';
--  5
```

**118 LLM steps across 58 agents. 5 definitions reference the marker.** The
consumer-enforced half of the contract is unenforced at the overwhelming
majority of call sites.

Guarded call sites known today (both had to be built by hand, one incident at a
time):
- `diagnose_council_decide_action.go` — `markerFieldFor` + degraded-review
  handling, built after `bugs_closed/019`.
- `verify_report_prose_action.go` — built 2026-07-25 (`8e8b55818`) for the
  gripper dossier.

## Why this is the shape that keeps recurring

This is the **same shape as `bugs_open/021`** (durable write guard covers ONE
path; the same shape is unguarded elsewhere) and as the `missingkey=zero`
family: a generic behaviour that fails silently, patched at whichever call site
last got burned, with the root behaviour left generic and unguarded. The
council seat that raised it named that lineage explicitly.

> `bugs_open/021` is OWNED by the durable-write-guard workstream (`who-owns.py`,
> active through 2026-07-24). This case is filed as a **sibling instance**, not
> as work on 021 — do not start a competing fix there.

The failure is worse than a crash because it is **silent and plausible**: a
response cut mid-way often still parses. A truncated JSON object can close with
every required key present and only trailing content missing, so schema
validation passes, the step succeeds, and the consumer stores a confident
half-answer. `bugs_closed/019` is the recorded case of exactly this (a review
cut mid-array closed into valid JSON with a recognised verdict and silently
missing objections).

## What is NOT claimed

- **[UNMEASURED]** How many of the 113 unguarded steps have actually been
  truncated in production. The marker is stamped on the step result in
  `orchestration_states.collected_data`, so this is answerable — see below —
  but I have not run it.
- **[UNVERIFIED]** Whether any specific unguarded consumer would visibly
  misbehave on a fragment. The exposure is structural; the impact per site
  needs the query above plus a read of each consumer.

The honest claim is exposure, not observed harm. That is still worth filing,
because the whole point of the marker is that harm from it is invisible.

## The query that would size it

```sql
-- how often has a tolerated truncation actually happened, and where?
SELECT agent_type, count(*)
FROM orchestration_states,
     LATERAL jsonb_each(collected_data) AS f(step, val)
WHERE jsonb_typeof(val) = 'object' AND (val->>'__truncated')::boolean IS TRUE
GROUP BY agent_type ORDER BY 2 DESC;
```
Note the retention trap recorded in the dormant-agents workstream:
`orchestration_states` is **pruned at 24h**, so this measures the last day only
and a zero result is not evidence of never.

## Fix candidates (not yet chosen)

1. **Fail closed by default, opt in to tolerance.** Make
   `ExecuteLLMPromptAction` return an error on truncation unless the step config
   sets something like `tolerate_truncation: true`. Correct, and it inverts the
   default so a new call site is safe by omission rather than by vigilance.
   Cost: every existing step that today silently tolerates would start failing —
   which is the point, but it needs the measurement above first to know the
   blast radius.
2. **Generic consumer-side guard.** Have the orchestrator refuse to pass a
   `__truncated` result into a subsequent step unless that step declares it
   accepts one. Catches every consumer without touching the producer, but adds
   orchestrator complexity.
3. **Surface, don't block.** Write every tolerated truncation to
   `agent_error_log` (severity warning) so the exposure is queryable and
   alertable even where it is tolerated — the `recordUnknownVerdict` precedent
   in `complete_work_item_verification.go`, which exists because a `zap.Warn`
   dies with its pod. Weakest of the three, but cheap and compatible with either
   of the others.

Candidate 1 + 3 together look right: invert the default, and make the
tolerated cases visible. **Do the measurement before choosing** — the blast
radius of candidate 1 is exactly the thing this file does not yet know.

## How to verify a fix

Induce the fault, do not trust a green path (`verify-the-failing-branch`): set
`max_tokens` absurdly low on a test agent's step so the response is certainly
cut, and confirm (a) the step fails or is recorded, per the chosen candidate,
and (b) no downstream artefact is written from the fragment. A run that
succeeds proves only that the happy path still works.

## Related

- `bugs_closed/019` — one truncated reviewer voided a whole council round; the
  case that introduced tolerate-and-mark.
- `bugs_open/021` — durable write guard covers one path only (same shape, OWNED
  elsewhere).
- `bugs_open/012` — a rewrite persisted a fragment and reported success; the
  `output_tokens == max_tokens` rule in CLAUDE.md comes from it.
