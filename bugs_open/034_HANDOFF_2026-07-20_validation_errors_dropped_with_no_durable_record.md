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
