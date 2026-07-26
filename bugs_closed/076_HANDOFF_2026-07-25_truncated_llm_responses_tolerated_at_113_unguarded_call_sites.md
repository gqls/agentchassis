# 076 — tolerating a truncated LLM response was opt-in but UNPOLICED: nothing checked that a consumer read the marker

> **Filed as** *"a truncated LLM response is tolerated fleet-wide; only 5 of 58
> agents check the marker"* — **and that headline was wrong**, as was the
> filename, which is kept so existing pointers still resolve. The 5-of-58 ratio
> measured the wrong property and the real exposure was 37 steps, all of them
> already guarded. The corrected statement is the heading above. See the
> CORRECTED block below; the wrong measurement is the most instructive thing in
> this file and is deliberately not edited away.

**Filed:** 2026-07-25 by the gripper-dossier thread (robot_hands_gripper_dossier),
prompted by a council objection (seat `bug_historian`, correlation
`7ed137d1-361c-4f69-9361-9e4ba1dfa6bf`, round 2) that asked the one question I
had not: *how many other call sites have this exposure?*
**Status:** **CLOSED 2026-07-26 — fixed AND live**, which is the `/bugs_closed/`
bar. Live in chassis **v1.0.1168**, verified against the running pod and then by
**inducing the failing branch**, not by a green happy path. Council **APPROVED**
(corr `470678f4-5149-4c96-b6a8-fa0185c88426`, round 1, 12 reviewers,
`unreadable: 0`); all four advisory objections are answered below, two of them in
code. Fix: `platform/orchestration/actions/truncation_guard.go` +
`diagnose_council_decide_action.go`.

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

---

# LIVE AND VERIFIED, 2026-07-26 — chassis v1.0.1168, then v1.0.1169

## It was inert for three hours, and the tag would not have told you

Pod `agent-chassis-5785dd5c85-jff28` ran **v1.0.1167, built 17:11:30Z**. The guard
committed at **17:24Z** — thirteen minutes after the image. So `4208e7f41` was in
git, in `HEAD`, and in no binary anywhere. The discriminating pod-grep is what
said so, and it is discriminating precisely because the string is NEW:

```
076 guard message:      0     <- pre-roll
positive control:       1     ("tolerate_truncation", which the old binary has)
```

Rolled **chassis-only** to v1.0.1168 (then v1.0.1169). All 183 agent definitions
run `docker.io/aqls/agent-chassis`, so that one image is the only one that has to
exist at the new tag — `make deploy-agents` would have sed'd all 14 services'
kustomizations to it and put the other 13 into ImagePullBackOff. Post-roll, in
the running pod:

```
076 guard message:      1
degradation error_code: 1     (TRUNCATION_DEGRADED_REVIEW)
positive control:       3
negative control:       0     (a string that should not exist)
```

## The failing branch, induced — with a control that discriminates

Two throwaway agents, **identical except for the one thing under test**: whether
any step in the workflow declares it can read the marker. Both opt into
`tolerate_truncation` with `max_tokens: 16`, so both are certain to be cut.

| agent | downstream step | outcome |
|---|---|---|
| `truncation-guard-probe` | nothing reads the marker | **FAILED at `truncate_me`** |
| `truncation-guard-probe-ok` | `accepts_truncated: true` | **COMPLETED at `finish`** |

The refusal, verbatim from `orchestration_states.error`:

> step truncate_me failed: failed to execute action execute_llm_prompt: step
> "truncate_me" sets tolerate_truncation but no step in this workflow consumes
> the `__truncated` marker, so a partial response would be indistinguishable from
> a complete one (bugs_open/076): raise max_tokens, drop tolerate_truncation, or
> give the consuming step accepts_truncated: true once it handles a partial:
> response truncated: stop_reason=max_tokens (output_tokens=16 reached the
> configured cap, 0 chars recovered)

**No artefact was written.** `collected_data` for the failed run has no
`truncate_me` key at all — the keys are only `action`, `config`, `input_data`,
`agent_group`, `agent_config` and the `__`-prefixed plumbing.

The control, by contrast, kept the fragment and marked it, which is the whole
point of the mechanism:

```json
"truncate_me": {"type": "text", "result": "", "__truncated": true, "__truncated_output_tokens": 16}
```

with `error` empty. **The control is what makes this a proof rather than a
coincidence**: a guard that blanket-failed every truncation would look identical
on the first row and is refuted by the second.

> **The probe agents were DELETED after the run** — see the drop statement in
> §"Re-running this". They were `is_active` agent definitions in the live
> database, and leaving them would be leaving two loaded guns in the fleet.

## The regression check: the 37 guarded steps still work

The induced fault proves the guard refuses. It says nothing about whether the
councils the mechanism exists for still function — and those are exactly the 37
steps that opt into tolerance. Two live `council-gate` rounds reached a verdict
on the new binary:

```
18:03:45 | complete_rejected | COMPLETED   (corr 88ef6d08)
18:08:39 | complete_revise   | COMPLETED   (corr 11a36278)
```

Neither is mine; both are other threads' submissions, which makes them better
evidence than a run I controlled. A guard that had broken the guarded path would
have shown up here as `complete_invalid` or a stalled review step.

## A defect the induced fault found in this very fix

`llm_call_log` recorded the REFUSED call as **`TOLERATED (step continued on the
partial):`**. The chain had not continued — the run failed.

The prefix was chosen from step config alone, before the guard ran, and since
the guard landed, opting in is no longer sufficient to make that claim true. It
is the exact inverse of the misreading the prefix was added to prevent (council
round 2eed453a: it exists so a tolerated cut is not counted as a pure failure,
and it had begun counting a hard failure as tolerated).

Fixed in `c640a8642`, live in **v1.0.1169**: the guard's verdict is resolved
before the log line and **reused** at the tolerate branch, so the log and the
behaviour are the same variable and cannot disagree. A refused truncation now
logs `REFUSED (bugs_open/076: tolerate_truncation set, but no step in this
workflow reads the __truncated marker):`.

**This is the argument for inducing the fault rather than reading a green path,
made against this case's own fix.** Every unit test passed, the council approved,
the pod-grep was clean, and the forensic row was still lying. Only running the
failing branch and then *reading what it wrote* surfaced it.

Re-induced on v1.0.1169 (run `947faed0`), and the row now matches the behaviour:

```
llm_call_log.error_message:
  REFUSED (bugs_open/076: tolerate_truncation set, but no step in this workflow
  reads the __truncated marker): response truncated: stop_reason=max_tokens ...
orchestration_states.error:
  step truncate_me failed: ... sets tolerate_truncation but no step in this
  workflow consumes the __truncated marker ...
```

## Re-running this

```sql
-- the two probes, dropped after use
DELETE FROM agent_definitions WHERE type IN ('truncation-guard-probe','truncation-guard-probe-ok');
```
Seed a `diagnose`/`coordinator` agent (the `check_ad_category` constraint allows
only strategist/executor/analyst/integrator/coordinator/specialist) with
`workflow.start_step`, `processing_mode: "orchestrator"`, one
`execute_llm_prompt` step at `max_tokens: 16` with `tolerate_truncation: true`,
and a `complete_workflow` step — with and without `accepts_truncated: true`.
Fire with the `095_TRIGGER` envelope on `system.agent.generic.requests`,
`action=orchestrate`, `config.agent_type=<probe>`.

> **A dispatch can vanish, and the exit test is what tells you.** The first two
> firings (T+252s and T+295s after the chassis restart) never became
> orchestrations. Re-firing the **byte-identical** message at T+7min worked, and
> the rows appeared ~60s later.
>
> **[UNVERIFIED] — and I got this wrong once already in this file.** I first
> wrote that the cause was a UUID `client_id` being rejected in favour of
> `demo_client`. **That is refuted**: the successful re-fire used the same UUID
> `client_id`. I then attributed it to the ~300s post-restart drop rule — but the
> third probe was fired at **T+100s** on a freshly rolled pod and landed fine, so
> a simple time rule does not explain it either. Both stories were tidy and
> neither survived. What is established is only that a dispatch can silently fail
> to become an orchestration and that re-firing is cheap.
>
> **I then guessed a third time, and was wrong again — so the honest answer is
> that I never established this, and the useful lesson is about the guessing.**
> I applied `bugs_open/052`'s exit test (*has anything newer drained past me?*),
> concluded my council resubmission had been dropped, and re-fired it twice. It
> had not been dropped. It landed **5m37s** after publication, at `18:35:57`.
>
> Two things made the test misfire in my hands, and both are worth knowing
> because the test itself is sound:
> - **The comparator must be on YOUR lane.** I first compared against
>   `endpoint-health-checker` and `build-pipeline-trigger`, which
>   `bugs_closed/030` moved onto their **own** cron topic
>   (`EXTRA_REQUEST_TOPICS`). They drain regardless of the generic lane, so they
>   can never be evidence about it.
> - **The comparator must have been PUBLISHED AFTER your message.** My second
>   attempt did filter to `council-gate`, but the run I saw advancing had started
>   *before* I fired. A queue that is FIFO makes "an older job is progressing"
>   entirely compatible with "mine is still waiting".
>
> ```sql
> -- both corrections applied
> SELECT created_at, owner_agent_type, current_step, status
> FROM orchestration_states
> WHERE created_at > '<the moment you published>'      -- published AFTER you
>   AND owner_agent_type NOT IN ('endpoint-health-checker','build-pipeline-trigger')
> ORDER BY created_at DESC;
> ```
>
> **The operational rule, with no mechanism attached: budget ~6 minutes on this
> lane before concluding anything.** Every one of my three explanations was tidy,
> confident and wrong within twenty minutes, and the cost of each was a duplicate
> dispatch. Waiting would have cost nothing.

---

# RESIDUALS CLOSED, 2026-07-26 — commit `0657143b5`

Two of the residuals below were closed in code the same day, both of them
objections this council raised. Second submission:
`SUBMISSION_CORR=1535e2ac-2d5d-4a9c-8093-38a81dbcd472`.

> **That second submission was stranded by my own image roll, and the mechanism
> is worth knowing.** It was fired at 17:57:34, reached `review_reuse_agent`, and
> its `updated_at` froze at **18:00:52** — which is the moment the v1.0.1168
> rollout terminated the pod that owned it. It did not fail, error or retry; it
> simply stopped, and a `status` of `EXECUTING_STEP` looks identical to healthy
> progress. **A chassis roll strands every in-flight orchestration at whatever
> step it was on** — so a roll and a council run are not independent, and
> "submitted" is not a state that survives a deploy. Resubmitted with
> `RESUBMIT_CORR` once the chassis was stable at v1.0.1169. The staleness test
> is the only way to see it: `EXTRACT(EPOCH FROM (now()-updated_at))` on the run,
> not `status`.

## 1. The consumer's reaction is now durable

The asymmetry was the defect: the PRODUCER half of tolerate-and-mark has been
permanent all along — `ai_actions.go:404` prefixes the `llm_call_log` row with
`TOLERATED (step continued on the partial): ` — while the CONSUMER half, what
`diagnose_council_decide` then *did* about the fragment, was a `zap.Warn` that
dies with its pod. So "has a consumer ever actually degraded?" was unanswerable
from data, which is a poor position for a mechanism whose entire purpose is to
make an invisible failure visible. This is the same argument, from the same seat
(`bug_historian`), that put `recordUnknownVerdict` into
`complete_work_item_verification.go` in July.

`recordTruncationDegradation` now writes one `agent_error_log` row per damaged
seat, `severity='warning'`, `error_code='TRUNCATION_DEGRADED_REVIEW'`, naming
which of the three branches did the damage:

| branch | what happened | loss |
|---|---|---|
| `producer_marker` | parsed cleanly, but the producing step stamped `__truncated` | trailing objections may be gone |
| `salvaged_from_invalid_json` | cut mid-structure, closed by hand | everything after the cut |
| `unsalvageable_invalid_json` | cut too early to recover | the whole seat |

Written **after** the `council_report` insert and best-effort throughout, so it
can never cost a decision that is already made and already persisted. Scope is
truncation only — the schema-slip salvage (`bugs_closed/036`) shares that loop
but is a different defect with a different remedy, and folding it in would make
the `error_code` mean two things.

Why not `orchestration.LogAgentError`, the one exported insert: package `actions`
is imported **by** `platform/orchestration`, so using it is an import cycle. The
six existing hand-rolled inserts in this package exist for that reason.

Query it (note `occurred_at`, **not** `created_at` — that column trap has cost
other lanes a round):
```sql
SELECT occurred_at, context->>'review_field', context->>'branch',
       context->>'council_decision'
FROM agent_error_log WHERE error_code = 'TRUNCATION_DEGRADED_REVIEW'
ORDER BY occurred_at DESC;
```

### It fired twice from natural traffic within 20 minutes of going live

```
18:08:27 | review_prior_art.result  | salvaged_from_invalid_json | rejected
18:15:11 | review_editquality.result| salvaged_from_invalid_json | revise
```

Two real council seats, in two different rounds, cut mid-JSON and recovered by
`salvageTruncatedReview` — each one an opinion that **gated a live verdict while
carrying only the objections that arrived before the cut**. One of those rounds
ended `rejected`, the other `revise`.

This closes the file's own `[UNVERIFIED]` residual ("no induced-truncation run
has exercised `diagnose_council_decide`'s degrade path end to end") by something
better than an induced run: **two unforced ones, in production, inside twenty
minutes.** It also sizes the thing that was invisible. At roughly six an hour,
this had been happening continuously and the only trace was a `zap.Warn` in a pod
log that is discarded on every rollout — which is exactly why "we have never seen
it happen" was never evidence that it did not.

## 2. The registry can no longer underclaim silently

`bug_historian`'s low objection was right: the forward lockstep held every
*registered* action to a real reader, but nothing forced a *new* reader to be
registered, so an unregistered consumer was invisible to the guard.
`TestEveryMarkerReaderIsRegisteredOrExempt` inverts it — every non-test file in
the package that reads the marker must implement a registered action or sit in
`truncationMarkerExemptions` **with a stated reason** (currently three: the
producer `ai_actions.go`, the doc-comment-only `types.go`, and the registry's own
file). Adding a reader now forces a decision instead of permitting an omission.

> **A trap the inverse scan surfaced, worth more than the test itself.** This
> package holds a SECOND truncation marker one underscore away:
> `workflow_actions.go`'s `truncatedResponseStub` writes `"__truncated__"` for a
> Kafka response over `max_response_bytes` — a different mechanism, a different
> remedy, nothing to do with an LLM cut at `max_tokens`. `strings.Contains`
> matches it as a substring. So a naive scan both demands an exemption for a file
> that never touches the LLM contract *and*, in the dangerous direction, would
> credit a future response-stub handler as a truncation-aware consumer.
> `readsLLMTruncationMarker` strips it; a probe with the naive form was run and
> produced exactly that false demand, so the stripping is load-bearing.

## Every check here was falsified before it was believed

The lesson this case already paid for once (the vacuous lockstep, §MISSTEP) was
applied as a rule this time. Five probes, each of which failed with the intended
message and then passed on restore:

| probe | expected failure |
|---|---|
| `error_code` changed to `SOMETHING_ELSE` | all three rows flagged |
| message stripped of its seat and branch | `error_message does not name the seat and the branch` |
| `range damage[:1]` — only the first seat recorded | sqlmock: `remaining expectation which was not matched` |
| a new unregistered marker reader dropped into the package | `zz_probe_reader.go reads the "__truncated" marker but implements no action` |
| `readsLLMTruncationMarker` reduced to a naive `Contains` | false demand against `workflow_actions.go` |

---

## Residual (carried past the close — none of these keeps 076 open)

Two residuals the earlier revision listed here are now **closed in code**: the
consumer's degradation is durably recorded, and the registry's underclaim is held
by an inverse lockstep test (see §RESIDUALS CLOSED). A third was closed by the
verification itself: `diagnose_council_decide`'s marker path has now been
exercised by an induced truncation end to end.

What remains:

- **Named next step, from the guardian's medium objection:** a static check over
  `agent_definitions.default_config` — flag any step with
  `tolerate_truncation: true` whose workflow has no truncation-aware consumer, at
  seed/registration time rather than at the moment of truncation. That catches
  the bad config *before* a run exists, and it is the layer at which
  `accepts_truncated` could also be validated. This runtime guard is the floor;
  that would be the ceiling. Not started.
- **[UNVERIFIED]** `accepts_truncated` is trusted, not checked. Nothing confirms
  a step declaring it actually reads the marker. The positive-control probe used
  it on a `complete_workflow` step that reads nothing — which is exactly the
  hole, demonstrated. It is the config hatch's nature (the config lives in the
  DB, so no Go test can falsify it); the static check above is where it would be
  validated.
- **[UNVERIFIED]** Whether the guard's floor is ever *too* low in practice — it
  asks whether the workflow has **a** reader, not whether the fragment reaches
  one. A workflow with a guarded consumer on a branch the truncated value never
  flows through would still pass. Rejected candidate 2 (reference-scanning) is
  why; 28 spurious matches in the council agents alone made a stricter gate fail
  closed on sound workflows.
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
