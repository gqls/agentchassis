# HANDOFF — a whole class of message failures is dropped with no durable record, no error response, and no retry

**Created 2026-07-20** while working `bugs_open/002` error **F** (on-demand
discovery dispatch: envelope accepted, nothing runs). F's own suspects were
both wrong; this is what the code actually says. **Severity: fleet-wide,
diagnostic — it does not itself break a workflow, it makes broken workflows
invisible and un-investigable.** Same defect class as `bugs_open/017`, which
was already fixed for one narrow site (`c80fffc83`, council-reviewed).

---

## The defect

`platform/agentbase/agent.go:828-845`. After `ProcessMessage` returns an error,
the agent classifies it by **substring** and, on a match, returns early:

```go
errMsg := err.Error()
if strings.Contains(errMsg, "is required") ||
    strings.Contains(errMsg, "validation") ||
    strings.Contains(errMsg, "invalid") {

    contextLogger.Warn("Validation error in message processing - not calling handleProcessingError", ...)
    observability.MessagesDropped.WithLabelValues(a.AgentType, "validation_error").Inc()
    return   // <-- skips handleProcessingError entirely
}
```

Skipping `handleProcessingError` (`agent.go:859`) skips **three** things, and
the third is the expensive one:

1. **No retry** — deliberate, and correct for a genuine validation error.
2. **No error RESPONSE to the parent.** `handleProcessingError` is what builds
   and publishes the `ResponseMessage` carrying `InResponseToRequestID` /
   `MyOrchestrationID` back to whoever is waiting. A parent sitting in
   `AWAITING_RESPONSES` is therefore told *nothing* and waits for the reaper.
3. **No durable record.** The only trace is a `zap.Warn` on stdout plus a
   Prometheus counter. `MessagesDropped` is labelled `(agent_type, reason)`
   only — **no correlation_id, no site_id, no message_id** — so it can tell you
   *that* something was dropped, never *which* thing. There is no
   `agent_error_log` row, no `orchestration_states` row, nothing in the DB.

**Consequence.** Every thread here investigates via the database. From the DB's
point of view the message simply never existed. The pod log does carry an
`Error` line first (`agent.go:823`) — so "no error anywhere" is not strictly
true — but the chassis pod rotates ~3.6k lines/10min, so that line is gone
within minutes and is unrecoverable by the time anyone notices the absence.

## Why the substring match makes it much worse than it looks

The three needles are matched against the **whole error string**, anywhere in
it, with no anchoring to a validation origin. `"invalid"` in particular is a
substring of ordinary runtime failures:

- `unrecoverable after control-char repair: invalid character 'w' after object key:value pair`
  — a genuine **LLM output parse failure** (this exact text is a fixture in
  `platform/orchestration/actions/testdata/envelopes/`, see `bugs_open/002` B).
  Classified as a validation error, dropped silently, parent never told.
- anything wrapping `invalid character`, `invalid memory address`,
  `invalid connection`, `invalid syntax` (`strconv`), `x509: … invalid`.

So the branch does not merely swallow malformed envelopes. It swallows a
**database driver error, a nil-pointer panic recovery, a TLS failure and a
truncated-LLM-response error** with equal enthusiasm, and each becomes a
missing row rather than a reported fault.

## How this produced 002/F's symptom (zero rows, "no error anywhere")

`platform/orchestration/coordinator.go:142-144`:

```go
if execCtx.ClientID == "" {
    return fmt.Errorf("client_id is required to execute a workflow")
}
```

This returns **before** `getOrCreateState`, so **no `orchestration_states` row
is ever created**. The error text contains `"is required"`, so `agent.go:828`
swallows it. Net effect for a hand-rolled trigger that omits `client_id` from
its headers: Kafka accepts the message, the agent consumes and acks it, zero
rows appear, no DB error is recorded, and the one stdout line rotates away.
**That is exactly 002/F.2's reported symptom.** The canonical trigger
`076_improvement_loop_trigger.sh` passes `client_id` in its headers, which is
why it works.

> **Not proven for F.2's specific correlations.** I could not confirm the two
> failing envelopes (`cd2459ce…`, `199ba851…`) omitted `client_id` — the thread
> that ran them deleted its trigger (`15f612346`), and both the pod logs and
> the Prometheus counter for 2026-07-18 are long gone. **That evidential
> dead-end is itself the bug**: there is no surface on which to check. The
> mechanism above is proven from code; its application to those two runs is a
> strong hypothesis, not a verified fact.

## What 002/F got wrong (both suspects refuted)

- **(a) "`action=orchestrate` + bare `agent_type` is no longer supported."**
  **FALSE.** `orchestrate` is the *first* clause of `isOrchestrationAction`
  (`platform/messaging/processor.go:983-988`), and the `orchestrate` path reads
  `config.agent_type` (`extractGroupInfo`, `processor.go:991-1057`) and looks
  the workflow up from `agent_definitions` (`FindByType`,
  `platform/discovery/agent_discovery.go:109-125`). No inline workflow needed.
  Git history shows no removal or rename. All three discovery agents were
  checked live 2026-07-20: `is_active=true`, not deleted, `default_config->'workflow'`
  present and an object. **So the worry that "other hand-rolled triggers across
  the repo are silently no-ops" does not follow from `action=orchestrate`.**
- **(b) the missing `kcat -P -c 1`.** Not investigated here; still open, but it
  is a *publisher* fault and would not produce a clean single accepted message.

**Also worth knowing** (`processor.go:878-980`): the action is read from **Kafka
headers only** (`types/context.go:545`). `action` set in the JSON body but not
as a `-H` header yields `action == ""`, which fails `isOrchestrationAction` and
falls through to *Priority 3 — the consuming agent's own default workflow*,
with no log saying the action was unrecognised. There is no `default:` branch
and no dead-letter on that path. And the canonical trigger's `action=process`
works **not** because `process` is a valid action (it is not in the set) but
because an inline `config.workflow` is checked at **Priority 1**, before any
action test (`processor.go:893-899`).

## Fix candidates

1. **Surface it durably — the 017 fix, applied at this site.** Write an
   `agent_error_log` row (correlation_id, agent_type, message_id, the full
   error) before returning. `c80fffc83` is the template and the precedent; it
   went through council review for exactly this reasoning. *Smallest change,
   highest value — do this one first even if nothing else lands.*
2. **Stop classifying errors by substring.** Use a typed sentinel
   (`errors.Is(err, ErrValidation)`) or a dedicated error type. The current
   match is unanchored and demonstrably catches non-validation failures.
   `platform/validation/` already defines error values to build on.
3. **Still send the error response.** "Do not retry" and "do not tell the
   parent" are independent decisions that this branch conflates. A parent
   waiting on a validation-failed child should be failed fast, not left to the
   reaper. (Check against `bugs_open/003` before changing response behaviour —
   003's hangs are a Kafka network cause, NOT this; do not conflate them.)
4. **Label the metric usefully**, or drop it. `(agent_type, reason)` cannot
   identify an incident.
5. **Guard the trigger side**: `client_id` is required but nothing validates it
   at publish time. A `-H client_id=` check in the shared trigger scaffolding
   would stop the commonest instance at source.

## How to verify a fix

```sql
-- after deliberately publishing an envelope with no client_id header:
SELECT created_at, agent_type, correlation_id, left(error_message,120)
FROM agent_error_log ORDER BY created_at DESC LIMIT 5;
```
Expect a row naming `client_id is required to execute a workflow`. Today you
get nothing. Note Go changes here are inert until the chassis image is rebuilt
and rolled.

## References

- `bugs_open/002` error **F** (the symptom this explains; F's suspects refuted above).
- `bugs_open/017_…unregistered_action_workflow_invalid_marked_complete.md` +
  `c80fffc83` — the same defect class, already fixed at one narrow site.
- `bugs_open/003` — parent-hangs-at-AWAITING_RESPONSES; a *different* cause
  (Kafka broker-2 node network). Fix candidate 3 touches adjacent behaviour.

---

# WORKED 2026-07-25 — candidate 1 done at BOTH sites; one needle list

Commit `15b0a7d96`. Council submission `180d7c68-30d6-4bc9-9050-d5241cc0f3e0`.

## CORRECTION to the section above: there are TWO drop sites, not one

> **CORRECTED 2026-07-25.** This handoff located the defect at a single site,
> `platform/agentbase/agent.go`. That is where it was *found*, but it is not
> where it usually *fires*. There is a second, older copy of the same
> classifier one layer down — `MessageProcessor.handleError`
> (`platform/messaging/processor.go:541`) — and on the ordinary processing
> path the message reaches that one FIRST.

The two sites did not agree with each other, in two ways that matter:

| | `agentbase.processMessage` | `messaging.handleError` |
|---|---|---|
| needles | `is required`, `validation`, `invalid` | the same **plus `missing`** |
| after dropping | `return` (no response to parent) | `sendErrorResponse` then **`return nil`** |
| durable record | yes — added earlier (see below) | **none** |

**`handleError` returning `nil` is the load-bearing detail.** `ProcessMessage`
hands that `nil` back to `agent.go`, which logs *"Message processed
successfully"*. So for any error routed through `handleError`, the agentbase
layer never sees a failure at all — and the recorder that had been added there
cannot fire. Candidate 1 had been applied to the site that reports the drop,
not the site that performs it. **[VERIFIED from code; the two classifiers'
needle lists and the `return nil` are quoted in commit `15b0a7d96`.]**

The needle asymmetry had a real consequence: an error whose only needle was
`missing` was **dropped without retry by one layer and retried by the other**,
decided purely by which path the message took to reach the failure. This is the
dedup-index↔Go-list drift class again — two hand-maintained lists that must be
identical, and weren't.

## What was already shipped, undocumented

Fix candidate 1 (`agent_error_log` row + `matchedValidationNeedle`) was
**already in the tree and already live** when this session started. It went in
under `fe2ba5e52` ("v1.0.1146 - sweep"), an owner sweep commit — no doc, no
note here, no `Council-Reviewed` trailer. Nothing in `/bugs_open/` or the
workstream dirs mentioned it (`grep -rn recordDroppedValidationError docs/
bugs_open/ bugs_closed/ scripts/` → nothing). It is genuinely deployed:

```
kubectl exec -n ai-persona-system agent-chassis-774877f4c6-zjh4t -- \
  sh -c 'strings /app/agent-chassis | grep -c VALIDATION_ERROR_DROPPED'   → 1
  (positive control MARK_COMPLETE_FAILED → 1)
```

**[VERIFIED live 2026-07-25.]** So the lesson is not "nobody fixed it" — it is
that a fix landed at one of two sites and no surface said which.

## What this session changed

1. **`messaging.handleError` now records the drop** before returning `nil` —
   the site that had none. `platform/messaging/validation_drop.go` (new).
2. **One needle list.** `messaging.ValidationErrorNeedles` +
   `MatchedValidationNeedle` are now the single source for both layers.
   Agentbase's private three-needle copy is deleted. Pinned by a test on each
   side (`TestValidationNeedlesAreTheOnesBothLayersUsed`,
   `TestAgentbaseUsesSharedValidationNeedles`).
3. **One `agent_error_log` writer.** `orchestration.LogAgentError` is the
   exported form of the coordinator's existing `logAgentError`; the
   hand-rolled INSERT in `agent.go` is gone. It had already drifted — it
   omitted the `NULLIF(...)::uuid` casts and the `site_id`/`domain`/
   `work_item_id` columns.
4. **Both recorders detached from the request context** (5s timeout,
   `context.Background()`). Agentbase's used `a.ctx`, which is cancelled on
   shutdown — so a drop during pod drain, exactly the one nobody can otherwise
   see, would have had its record cancelled with it.

Rows land as `error_code='VALIDATION_ERROR_DROPPED'`, `severity='warning'`,
with `correlation_id`/`message_id`/`request_id`/`client_id`, `dropped_at`
(which of the two sites), `retried`, and **`matched_needle`** in `context`.

**The drop/no-retry decision itself is unchanged.** One deliberate behaviour
change: agentbase now also treats `missing` as non-retryable, adopting the
shared needle. That direction was chosen because messaging already refuses to
retry those everywhere else and its own comment says retrying them is an
infinite loop — the old agentbase behaviour was the anomaly.

## Still open (do NOT read this as closing them)

- **Candidate 2 — substring classification.** NOT done, deliberately. It needs
  every error construction site audited and would change retry behaviour
  fleet-wide. The mitigation shipped here is visibility, not correctness:
  `matched_needle` makes a misclassification a queryable row instead of an
  invisible one. The hazard is pinned as *passing* test cases — a truncated-LLM
  parse failure, `pq: invalid connection` and a recovered nil deref all still
  match on `invalid`.
- **Candidate 3 — error response to the parent.** NOT done at the agentbase
  site. (`messaging.handleError` does send one.) Check `bugs_open/003` first.
- **Candidate 5 — trigger-side `client_id` guard.** NOT done.
- **Candidate 4 — metric labels.** Effectively moot: the DB row now carries
  what the counter could not.

## Verification — NOT yet done, and this is the honest gap

Zero `VALIDATION_ERROR_DROPPED` rows exist:

```sql
SELECT occurred_at, agent_type, error_code, context->>'matched_needle'
FROM agent_error_log WHERE error_code='VALIDATION_ERROR_DROPPED'
ORDER BY occurred_at DESC LIMIT 15;   -- (0 rows), 2026-07-25
```

That is **consistent with, but not proof of, correctness** — no drop occurred
to record. Both classifier log-lines were absent from a 93-minute window on the
live pod (`grep -c "not calling handleProcessingError"` → 0;
`grep -c "NOT retrying to prevent infinite loop"` → 0), so the path simply did
not fire. A green count here proves deployment, not behaviour.

**What is still owed: an induced fault.** Publish an envelope that fails
classification and confirm a row appears. Note the original recipe at the top
of this file names `created_at`; the column is **`occurred_at`**.

> **Which site an induced fault will hit is [INFERRED from code read, not
> exercised].** With one shared list, any error passing through
> `handleError` is consumed there. The agentbase branch is reachable only on
> paths that bypass it — `processWithoutContext` (which returns `process()`'s
> error raw, `processor.go:1624-1629`) and the `NewMessageContext` failure
> return. A hand-rolled trigger with malformed headers takes the
> `processWithoutContext` route, because `types.FromHeaders` failing is what
> sends it there. Do not restate this as fact until a row proves it.

## Adjacent, UNVERIFIED, not investigated — for whoever picks this up

`sendWorkflowFailureResponse` (`processor.go:523`) routes through
`sendWorkflowResponse`, which builds every response with
`Body.Success: true` and `Body.Error: nil` — a workflow **failure** is
published in a success-shaped envelope with the error text buried in
`body["error"]`. `grep` found no reader of `Body.Success` in
`platform/orchestration/`, so this may be harmless. **[UNVERIFIED — I did not
trace what consumes these responses, and did not file it.]** It is the same
family as this bug (a failure that does not look like one) but it is a
separate claim and should be diagnosed, not assumed.

---

# UPDATE 2026-07-26 — there are FOUR drop sites, and the induced fault found them

## Council: APPROVED

`180d7c68-30d6-4bc9-9050-d5241cc0f3e0` — 7 reviewers, **`unreadable: 0`**
(the check that matters — `abstained: 9` is the relevance filter, not
truncation), 2 advisory objections, none high-severity. Answered in
`56e77a501`. The import-cycle objection was the substantive one and it
resolves cleanly: `platform/orchestration` imports neither `messaging` nor
`agentbase`, and the tree builds.

## The induced fault worked — by failing

Publishing a request with no `client_id` (correlation `034induce-1785071937`)
was *expected* to hit one of the two classifier sites. **It never got near
them.** It was rejected at `agent.go`'s `ValidateIncomingMessage` gate — the
first thing any inbound message meets — which published an error envelope and
returned. Verified:

```sql
SELECT count(*) FROM agent_error_log WHERE context::text LIKE '%034induce%';  -- 0
SELECT ... WHERE occurred_at > now() - interval '20 minutes';                 -- (0 rows)
```

**Zero rows. A live, reproducible instance of this exact bug, at a site
neither the handoff nor the first fix had touched.** A second induction
(`message_type=response` with no `in_response_to`) also missed: responses are
routed to the orchestrator at `processor.go:1498` *before* `ValidateContext`.

Then the sweep the council's `reuse_agent` seat asked for — *"make sure a
third silent duplicate isn't sitting somewhere neither layer's author
looked… the plan's own search stopped at the two sites named in the bug
file"* — turned up a **fourth**.

## The four sites

| # | site | on drop | durable record |
|---|---|---|---|
| 1 | `agent.go` `ValidateIncomingMessage` gate | error envelope, `return` — **not even a counter** | **added 2026-07-26** (`94c4ff471`) |
| 2 | `agent.go` `missing_orchestration_id` | error response + counter, `return` | **added 2026-07-26** (`56e77a501`) |
| 3 | `messaging.handleError` | `sendErrorResponse` + **`return nil`** | **added 2026-07-26** (`15b0a7d96`) |
| 4 | `agent.go` substring classifier | `return` — no response to parent | added earlier (`fe2ba5e52`), undocumented |

Sites 1 and 2 sit **ahead of both classifiers**, which is why the original
handoff never saw them: it reasoned from the classifier inward. Site 1 is the
one a hand-rolled or malformed trigger actually hits — which is how
`bugs_open/002` error F came to look like *"accepted, never executed, no error
anywhere"*.

> **CORRECTED 2026-07-26 — my own earlier claim in this file.** The section
> above says the reachable route to site 4 is `processWithoutContext`, marked
> `[INFERRED from code read, not exercised]`. The induced fault shows the
> marker was doing real work: **the message never reaches `ProcessMessage` at
> all** when a required header is missing, because site 1 rejects it first.
> The inference was not wrong about the code, it was wrong about which gate
> fires first. What caught it: publishing the envelope instead of reading
> more code.

## Disclosed behaviour change — read this before the roll

The needle unification means **agentbase now drops-without-retry on the
`missing` needle where it previously retried**. `agentbase.Agent` underlies
every agent type, so this is a fleet-wide retry-disposition change for that
error class. It was disclosed in the submission's `risks` and approved on that
basis; the guardian seat asked for it to be stated plainly rather than folded
under "the decision is unchanged", so: **it is a deliberate change to retry
semantics, not just added recording.** The argument for it is that messaging
already refused to retry those everywhere else and its own comment calls
retrying them an infinite loop — the old agentbase behaviour was the anomaly.

## Post-roll checklist (owed)

1. Pod-grep the new symbols against the running pod (not git, not the tag):
   `INCOMING_MESSAGE_REJECTED`, `MISSING_ORCHESTRATION_ID`, with a positive
   and a negative control.
2. **Re-run the induced fault** — publish to `system.agent.generic.requests`
   with `correlation_id`, `orchestration_id`, `message_type=request` and
   **no `client_id`**; expect one `INCOMING_MESSAGE_REJECTED` row naming
   `client_id` in `context->'missing_headers'`. (kcat needs `-P -c 1`.)
   Note the original recipe at the top of this file says `created_at`; the
   column is **`occurred_at`**.
3. Watch `INCOMING_MESSAGE_REJECTED` / `VALIDATION_ERROR_DROPPED` volume for
   the retry-semantics change above.

## Status

**OPEN.** Sites 3 and 4 are **LIVE in v1.0.1165** (pod-verified 2026-07-26:
`VALIDATION_ERROR_DROPPED` resolves twice, one literal per recorder, with
controls). Sites 1 and 2 (`94c4ff471`, `56e77a501`) are **committed and
inert** — they need one more chassis roll. Close after that roll plus the
induced-fault row in step 2, which is the first end-to-end proof this bug has
ever had.
