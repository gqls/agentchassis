# 076 — a truncated LLM response is tolerated fleet-wide; only 5 of 58 agents check the marker

**Filed:** 2026-07-25 by the gripper-dossier thread (robot_hands_gripper_dossier),
prompted by a council objection (seat `bug_historian`, correlation
`7ed137d1-361c-4f69-9361-9e4ba1dfa6bf`, round 2) that asked the one question I
had not: *how many other call sites have this exposure?*
**Status:** FIXED IN CODE, **INERT until the next image roll** — so still OPEN
per the `/bugs_closed/` bar (fixed AND live). Fix: `platform/orchestration/
actions/truncation_guard.go`.

> **CORRECTED 2026-07-26 — the title and the headline measurement are both
> wrong, and the correction matters more than the fix.** There are not 113
> unguarded call sites. Those 113 steps **do not tolerate truncation at all** —
> they already fail closed, because `tolerate_truncation` is opt-in and defaults
> to false (`ai_actions.go:401`). Fix candidate 1 below was therefore **already
> built before this case was filed**. The measurement that produced "5 of 58"
> tested the wrong property (see *The measurement was of the wrong thing*).
> The real defect is narrower and still real: the opt-in was **unpoliced**.
> Caught by reading the producer before fixing it.

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

---

# FIX, 2026-07-26 — police the opt-in

## The measurement was of the wrong thing

Every figure below is live, run 2026-07-26 against `clients_db`.

**1. `tolerate_truncation` is opt-in and defaults false.** The mechanism section
above quotes `markTruncated(out, truncationTolerated, ...)` as if it always
fires. It does not: `markTruncated` returns immediately unless `tolerated` is
true, and `truncationTolerated` is only ever set inside
`if isTruncatedCall && tolerateTruncation`, where

```go
tolerateTruncation := datahelpers.GetBoolField(params.StepConfig.Config, "tolerate_truncation", false)
```

So a step that says nothing **fails closed today**. Candidate 1 was already the
behaviour. What the live config actually shows:

```sql
SELECT count(*) AS total_llm_steps,
       count(*) FILTER (WHERE COALESCE(v->'config'->>'tolerate_truncation','false')='true') AS tolerating
FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') AS e(k,v)
WHERE d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND v->>'action'='execute_llm_prompt';
--  118 | 37
```

**37 of 118, not 113 of 118.** The other 81 have no exposure to guard.

**2. "Definitions that mention `__truncated`" is a false-negative test.** The
`LIKE '%__truncated%'` query counted 5 because `council-gate` and `fix-proposer`
happen to name the marker in config text. `feature-designer` does not — and is
guarded anyway, because its guard is in **Go**: `diagnose_council_decide` derives
the marker path from the reviewer field (`markerFieldFor`:
`review_x.result` → `review_x.__truncated`). Config mention is neither necessary
nor sufficient, so the "5 of 58" ratio measures nothing about guarding.

Where the 37 tolerating steps actually live, and what consumes them:

| agent | tolerating steps | decide step's action | guarded |
|---|---|---|---|
| `council-gate` | 16 | `diagnose_council_decide` | yes |
| `fix-proposer` | 16 | `diagnose_council_decide` | yes |
| `feature-designer` | 5 | `diagnose_council_decide` | yes |

**Every tolerating step in the fleet is consumed by a guarded action.** The
observed-harm figure the file left `[UNMEASURED]` is therefore: none reaching an
unguarded consumer.

**3. The marker does survive into `collected_data`** — which the gripper thread
had flagged as unproven ("the truncation guard is a silent no-op" was the worry):

```sql
SELECT jsonb_object_keys(collected_data->'review_editquality'), count(*) ...
--  result: 25 | type: 25 | __truncated: 7 | __truncated_output_tokens: 7
```

7 live orchestrations carry the marker. The delivery half is proven; note this
proves the marker ARRIVES, not that the consumer's degrade path was exercised.

**4. Candidate 3 is already built producer-side.** `llm_call_log` rows carry a
`TOLERATED (step continued on the partial):` prefix — 29 real rows, most recent
2026-07-26 16:21. That is durable and queryable, unlike the `orchestration_states`
copy the file's sizing query uses (pruned at 24h).

## So what was actually broken

The producer half fails closed; the consumer half was enforced by **nothing**.
Both existing guards were hand-built after an incident. Any step could set
`tolerate_truncation: true` in a workflow where nothing reads the marker, and no
test, seed check or report would say so. The exposure was **latent, not active** —
which is precisely why it would have been found the expensive way.

## The fix

`platform/orchestration/actions/truncation_guard.go`. Before keeping a partial,
`ExecuteLLMPromptAction` now checks that the workflow contains a step that reads
the marker; if not it **refuses to tolerate** and returns the loud failure the
step would have had before opting in.

- `truncationAwareActions` — one registry naming each action whose code reads
  `__truncated`, with its mechanism, replacing "grep and hope".
- `accepts_truncated: true` — per-step config escape hatch for actions too
  generic for the Go registry to speak for.
- `ActionParams.WorkflowSteps` — the plan, plumbed from the coordinator (its only
  production construction site), so a producing step can see its own workflow.
- The producing step cannot certify itself.

**Candidate 2 was considered and rejected on evidence.** Gating on "does this
step reference the truncated field" needs reference-detection across prompt
templates and config paths; a textual scan against live config matched **28**
"consumers" in the council agents alone, nearly all spurious. A gate built on it
would fail closed on sound workflows. So the guard asks whether the workflow has
**a** guarded consumer, not whether the fragment reaches it: **a floor, not a
proof**, stated as such in the code.

**Blast radius: zero.** The live mirror of the guard's own predicate:

```sql
-- has_guarded_consumer, per tolerating agent
council-gate | t     feature-designer | t     fix-proposer | t
```

## MISSTEP, recorded — the lockstep test was vacuous on its first write

`TestTruncationAwareActionsReadTheMarker` scans the package for a file that both
names a registered action and reads `__truncated`. It passed a deliberate
falsification probe (a bogus `render_page_html` entry) **because
`truncation_guard.go` itself contains both** — the registry file satisfied its own
check, so any name added would have passed forever.

Caught by running the probe rather than trusting the green. Fixed by excluding
the registry's own file; re-probed, and it now fails with the right message.
This is the `WRONG_CALLS.md` "a check sharing the fix's regex cannot falsify it"
pattern, third recorded instance — **run the probe, every time.**

## How to verify after the roll

Deployment (the symbol is new, so this is discriminating — it cannot pass
pre-roll):

```bash
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "no step in this workflow consumes the __truncated marker"'
```

Correctness needs the **failing** branch induced, not a green council: seed a
throwaway agent with one `execute_llm_prompt` step carrying
`tolerate_truncation: true`, `max_tokens: 16`, and **no** guarded consumer, then
confirm the step fails with the bugs_open/076 message and writes no artefact. A
passing council proves only that the 37 guarded steps still work — which is the
regression check, not the proof.

## Council — APPROVED, round 1

`SUBMISSION_CORR=470678f4-5149-4c96-b6a8-fa0185c88426`, 2026-07-26.
**APPROVED on round 1 with 4 advisory objections, none high-severity;
`unreadable:0`, `abstained:4`** (read `unreadable` first — an approval standing
beside an unreadable seat is downgraded automatically, so `approved` here means
every relevant seat was actually read).

**No `Council-Reviewed:` trailer is possible on the code commits.** The verdict
post-dates `4208e7f41` and `37a4e7b9a`, and forward-only forbids the amend that
would add it. This is therefore a **permanent `098` false negative** — recorded
here so nobody re-investigates the gap. The correlation above is the audit link.

### The objections, and what was done with them

- **editquality (low) — "as written this doesn't compile"**: `err` is undeclared
  in the `fmt.Errorf`'s scope, only `truncErr` is. **REFUTED, and the objection
  is sound reasoning from what it could see.** `err` is declared at
  `ai_actions.go:377` (`result, err := aiClient.GenerateText(...)`) and the whole
  truncation branch sits inside the `if err != nil {` that opens on 378, so the
  `%w` arg is correct and `go build ./platform/orchestration/actions/` is clean.
  The sketch showed the branch without its enclosing scope. **The lesson is
  about submissions, not this code**: a sketch cropped above its scope invites a
  correct-looking compile objection — the same class as the quote-fidelity trap.
- **bug_historian (medium) — `accepts_truncated` is an unpoliced
  self-declaration.** ACCEPTED as a real residual (it was named in the
  submission's own risks). A wrong or copy-pasted flag re-opens the exposure for
  that workflow, and nothing catches it. The Go registry is falsifiable by test;
  the config hatch is not, because the config lives in the DB. Left in, listed
  below as residual.
- **guardian (medium) — was a higher layer ruled out?** i.e. a static check over
  `agent_definitions.default_config` at registration/deploy time instead of
  plumbing the step map through the runtime dispatch struct. **A fair hit: the
  submission never discussed this axis, only the rejected candidate 2.** It is
  the better long-term shape — it would catch the bad config before a run exists
  rather than at the moment of truncation — and it is genuinely additional work,
  not a rewrite of this. Recorded as the named next step.
- **debug_historian (medium) — deployment confirmation unspecified.** Correct
  about the submission; the *case file* already carries the pod-grep and the
  induced-fault test (see "How to verify after the roll"). Gap was in the
  submission's risks section, not the plan.
- **bug_historian (low) — the registry can underclaim.** A future action that
  reads `__truncated` but is never registered is invisible to the guard. True,
  and it fails **loud** in that case (the safe direction): the workflow refuses
  to tolerate rather than silently accepting a fragment.
- **guardian (low) — `ActionParams` is a shared contract.** The field is purely
  additive and read only by `ExecuteLLMPromptAction`; the coordinator is the sole
  production construction site, verified against every `ActionParams{}` literal
  in the repo (all others are tests or `cmd/`).

## Residual (not fixed here)

- **[UNVERIFIED]** No induced-truncation run has yet exercised
  `diagnose_council_decide`'s degrade path end to end. The marker's arrival is
  proven; the consumer's reaction is not.
- The consumer's degradation is logged with `zap.Warn`, which dies with its pod.
  `llm_call_log` records that the producer tolerated, nothing durably records
  that a consumer degraded. The `recordUnknownVerdict` precedent applies.
- **Named next step, from the guardian's medium objection:** a static check over
  `agent_definitions.default_config` — flag any step with
  `tolerate_truncation: true` whose workflow has no truncation-aware consumer, at
  seed/registration time rather than at the moment of truncation. That catches
  the bad config *before* a run exists, and it is the layer at which
  `accepts_truncated` could also be validated. This runtime guard is the floor;
  that would be the ceiling. Not started.
- **[UNVERIFIED]** `accepts_truncated` is trusted, not checked. Nothing confirms
  a step declaring it actually reads the marker.
- `platform/orchestration/orchestration_test.go:171` does not compile at HEAD
  (`NewSagaCoordinator` called with 3 args, needs 4). Pre-existing, another
  workstream's, untouched here — but it means that package's tests do not run.

## Related

- `bugs_closed/019` — one truncated reviewer voided a whole council round; the
  case that introduced tolerate-and-mark.
- `bugs_open/021` — durable write guard covers one path only (same shape, OWNED
  elsewhere).
- `bugs_open/012` — a rewrite persisted a fragment and reported success; the
  `output_tokens == max_tokens` rule in CLAUDE.md comes from it.
