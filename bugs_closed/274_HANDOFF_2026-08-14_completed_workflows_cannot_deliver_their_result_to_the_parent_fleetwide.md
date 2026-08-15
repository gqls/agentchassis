# 274 — a SUCCEEDED workflow is reported to its parent as FAILED: two reply headers are never set, so the reply cannot pass validation — ~15,000 times across 60 agent types since 2026-08-03

> **⚠ READ §9 FIRST. It supersedes parts of what follows, and this file would otherwise contradict itself.**
> §9 (added the same evening) LOCATES the root cause and therefore: answers §3's two `[UNVERIFIED]`
> items (the missing fields are `sender_agent_type` and `in_response_to_step_name`, and the
> unlocated wrapper is `notifyParentOfSuccess`); **corrects §1 and the original title**, which said
> the parent proceeds without the payload — it is in fact told the child FAILED; and **DOWNGRADES
> §4** from a strong candidate for `bugs_open/213` §D to one its own evidence argues against.
> §5's fix candidates are superseded by §9's. Sections 1–8 are kept as filed, unedited, because
> the route from symptom to cause is the useful part — but do not act on them alone.

**Filed 2026-08-14** by the `bugfix_213_verifier_producer_join` lane. Found while chasing a
different question — bug 213's §D, "why do 10 of 14 completed items carry a payload that is not
their handler's" — via a `090` diagnosis run whose verdict was **UNVERIFIABLE for that question**
but whose runtime citations surfaced this.

**This symptom is already NAMED and explicitly UNFILED.** `bugs_open/216` (line ~89) recorded it
as *"an unexplained sibling symptom — completed workflows whose results fail 'message validation
failed' on `complete_workflow` delivery to the parent (correlation `aee5853d`, several `*/complete`
steps) — **unfiled at the time of writing, worth its own look**"*. It is **not** what 216 fixed
(216 fixed the recoverable-response replay, and is FIXED + LIVE + PROVEN on `v1.0.1266`). This is
that look, with a week more evidence.

---

## 1. The shape

A child workflow finishes its work, then **fails to deliver its result to the parent**, logging:

```
<agent>/<step> (complete_workflow) fatal: workflow completed but its result could not be
delivered to the parent (failed_transient): message validation failed
```

The child's work is done and often persisted. The **parent never receives the payload**, so
whatever the parent does next proceeds without it.

## 2. Scale [MEASURED 2026-08-14 17:15Z]

```sql
SELECT agent_type, count(*), min(occurred_at)::date AS first, max(occurred_at) AS latest
FROM agent_error_log WHERE error_message LIKE '%message validation failed%'
GROUP BY 1 ORDER BY 2 DESC;
```

**60 agent types, ~15,000 rows, continuous since 2026-08-03, still firing** (latest
`build-dispatch-loop` 17:13:39Z on the day of filing). Largest:

| agent_type | rows |
|---|---|
| `page-rerender` | 4,794 |
| `build-dispatch-loop` | 2,495 |
| `feed-ingester` | 1,140 |
| `internal-link-resolver` | 959 |
| `page-build-handler` | 948 |
| `page-content-writer` | 885 |
| `asset-deployer` | 439 |
| …54 more, incl. `color-variable-fixer` (43), `diagnose-agent` (43), `landmine-verifier` (63) |

This is not a corner: it spans content, rerender, imagery, deploy, discovery, diagnosis and the
meta-checks. **`page-rerender` at 4,794 is the fleet's highest-volume repair path.**

## 3. The mechanism, as far as it is READ (and where it stops)

**`"message validation failed"` exists at exactly ONE site** and its reachability is decisive
(`platform/kafka/producer.go:111-131`):

```go
func (p *KafkaProducer) ProduceWithValidation(...) error {
    if p.validator == nil { return p.Produce(...) }              // no validator → no error
    if !p.validator.ValidateOutgoingMessage(headers) {
        if headers["is_error"] == "true" { … send anyway … }     // errors exempt
        return fmt.Errorf("message validation failed")           // ← the only source
    }
    return p.Produce(...)
}
```

So every one of those ~15,000 failures means: **a validator IS injected, the message is NOT an
error message, and `ValidateOutgoingMessage` rejected it.** (This settles a doubt the `090`
verdict raised — it suspected the `ProduceWithValidation` hit might be a coincidental name match
against the `Produce` that `CompleteWorkflowAction` calls. The string's uniqueness makes the path
certain, whatever wrapper logs it.)

**What the validator actually checks — headers only** (`platform/validation/validator.go:42-67`):
it returns `false` when ANY of these is empty —

`client_id` · `correlation_id` · `orchestration_id` · `sender_agent_type` ·
`step_name` (falling back to `in_response_to_step_name`)

**So the failing responses are missing one of five envelope headers.** The validator logs an
`Error` naming each value, so **a pod log at the moment of failure identifies the missing field
directly** — that is the cheapest next step and it needs no new code.

**Where the read stops, stated rather than papered over:** the wrapper that catches the returned
error and writes the `agent_error_log` template *"workflow completed but its result could not be
delivered to the parent (%s): %s"* is **not located**. `CompleteWorkflowAction` wraps producer
errors as `"failed to send response: %w"`, which does not match. The `090` verdict named this as
its first outstanding item and it is still open. **`[UNVERIFIED]`** which of the five headers is
missing, and **`[UNVERIFIED]`** whether it is the same one for all 60 agent types.

**Refuted along the way:** the `090` hypothesised that `agent_error_log.context` would name the
failing validator or field. It does not — it carries `failed_step` (e.g.
`process_item_iter_0_call_handler`, `call_content_writer`) and sometimes `item_type`/`page_name`.
So the context column cannot answer this and the pod log must.

## 4. Why it matters beyond the log noise — the candidate consequence

**[CANDIDATE MECHANISM, NOT ESTABLISHED]** This is a strong candidate for `bugs_open/213` §D:
of 14 completed `dark_section_audit` items, only 4 carry their handler's response envelope; the
other 10 carry a payload that is not that handler's at all (a design-system spec for 9, an
unrelated child-page triage decision for the 1). `color-variable-fixer` appears in the census
above with 43 delivery failures, and `build-dispatch-loop` — the agent that completes those work
items — with 2,495.

**If a child's result never reaches the parent, the parent completes the work item with whatever
else is in its `collected_data`.** That would produce exactly the observed 10-of-14 split. It
would also mean **any completion-time check reading the handler's report is reading someone
else's** — which is a fleet-wide integrity concern, not a cosmetic one.

It is marked CANDIDATE because the joining step is unread: nobody has yet traced what the parent
substitutes when delivery fails. **Do not cite this as 213 §D's cause until that is read.**

## 5. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Find and fix the missing header at its source.** One pod log identifies the field; the
   validator's five-field rule is the whole contract. If a single producer path omits
   `sender_agent_type` or a step name, this is a small fix with a 15,000-row blast radius.
2. **Make an undeliverable result LOUD at the parent, not just at the child.** Today the child
   logs and the parent proceeds with a payload that is not the child's. A parent that received no
   result should refuse to complete, not substitute. (Compare bug 213's gate 1b, which refuses a
   completion whose handler reports nothing — the same principle one level up.)
3. **Do not "fix" it by exempting these messages from validation.** The `is_error` exemption
   already exists and is deliberate; widening it would convert a loud failure into a silent
   wrong-payload completion, which is candidate §4 above made permanent.

## 6. Traps for whoever picks it up

- **`bugs_open/216` is FIXED and this is NOT its defect.** Do not read 216's PROVEN status as
  covering this; its own text calls this a separate, unfiled sibling.
- **Do not diagnose this from `agent_error_log.context`** — it names the failed step, not the
  validator or the field (§3).
- **Do not diagnose it from `input_tokens` or any cost surface.** Unrelated, and on 2026-08-14 the
  fleet also hit a monthly API cap, which makes any same-day cost census misleading.
- **The `090` verdict for this correlation is `UNVERIFIABLE`, and that is about the ORIGINAL
  question** (213 §D's binding site), not about this. Its citations are sound; its outcome label
  grades the hypothesis it was given. Read the citations, not the label.

## 7. Diagnosis status

**Went through the loop** (`090`, run correlation `6f158444-145d-41a4-88d9-13d812939c58`,
verdict `UNVERIFIABLE` for 213 §D's question; earlier attempt `266be67d-a6e1-4afc-8fc1-84b553b2ea82`
died at its verdict step on the API cap, leaving a 54,805-char bundle worth reading). Two of the
three `needed_evidence` items the verdict listed have since been answered first-hand in §3 — the
call-path certainty and the `context` column — and the third (the log-template wrapper) is
recorded as still open rather than guessed at.

## 8. Relations

`bugs_open/216` (named this symptom, explicitly unfiled, and is a DIFFERENT defect) ·
`bugs_open/213` §D (the candidate consequence, and how this was found) · `bugs_open/040` (kafka
dial timeouts — mentions the same table, different failure) · WII-017 (gate 1b, whose ABSTAIN arm
is currently the fleet's best instrument for observing the wrong-payload effect: 4 of 4 abstentions
recorded, each naming the payload's actual top-level keys).

---

## 9. ROOT CAUSE LOCATED, 2026-08-14 (evening) — two header fields are never set, and the reply can NEVER pass validation

The `090` verdict listed "the wrapper that logs the template" as its first unresolved item. It is
`SagaCoordinator.notifyParentOfSuccess`, and the whole chain is now read, from a **live stack
trace** in a still-running pod (`agent-build-dispatch-loop-17079920-wmrvq`, 18:17:15Z):

```
validation/validator.go:58        ValidateOutgoingMessage → false
kafka/producer.go:118             ProduceWithValidation → "message validation failed"
kafka/reply_delivery.go:139       DeliverReply → FailedTransient
orchestration/coordinator.go:3751 notifyParentOfSuccess
orchestration/coordinator.go:4024 completeWorkflow
```

**The log line names the missing fields directly** — everything else is present:

```
client_id:"system" · correlation_id:"496c9c55-…" · orchestration_id:"b782a84e-…"
sender_agent_type:""            ← EMPTY
in_response_to_step_name:""     ← EMPTY
```

**And the construction site shows why they are empty: they are never set.**
`coordinator.go:3709-3721` builds `types.ResponseHeaders` with exactly
`InResponseToRequestID`, `Status`, `IsComplete`, `MessageType`, `TimeSent`, `OrchestrationID`,
`CorrelationID`, `ClientID` — and **no `SenderAgentType`, no `InResponseToStepName`**.
`ResponseHeaders.ToMap()` (`types/context.go:877-883`) emits `in_response_to_step_name` straight
from that unset field. `ValidateOutgoingMessage` (`validator.go:52-56`) requires **both**.

> **So this is DETERMINISTIC, not intermittent.** The same message cannot pass validation on any
> retry, on any broker, at any time. It is not a "transient" anything. That matches the data
> exactly — ~15,000 rows, 60 agent types, continuous since 08-03 with no quiet periods.

### The consequence is worse than a lost payload — the parent is told the child FAILED

`notifyParentOfSuccess` handles undeliverability by design (`bugs_open/158` item 1, owner ruling
2026-08-03), and its chosen answer is *"TELL THE PARENT IT FAILED"*: when `DeliverReply` does not
report `Answered()`, it logs *"Could not notify parent of success — notifying parent of FAILURE
instead"* and calls `notifyParentOfFailure`.

That reasoning is sound for a genuinely undeliverable reply. Here the reply is not undeliverable —
**it is malformed by construction**, so a workflow that **succeeded** is reported to its parent as
**failed**, every time this path runs.

### A second, smaller defect at the same seam

`DeliverReply` classifies a **validation refusal** as `FailedTransient`, in the same branch as
"broker unreachable" and "context cancelled" (`reply_delivery.go:139-145`, comment included). A
validation refusal is **permanent for that message** — nothing about waiting or retrying changes
it. Grouping the two means a deterministic defect is reported to operators, and to
`agent_error_log`, under a label that says "try again".

### Fix candidates, ordered by what makes the bad state unrepresentable

1. **Set the two fields at the construction site.** `SenderAgentType` and `InResponseToStepName`
   are both available on the coordinator's state at that point. Smallest possible change, ~15,000
   rows/11 days of blast radius. ⚠ **Check every other `ResponseHeaders` construction site in the
   same pass** — this one is a literal struct with named fields, so a sibling that omits the same
   two will fail identically and silently.
2. **Make the validator's refusal impossible to construct**: a constructor for `ResponseHeaders`
   that requires the five validated fields, so omission is a compile error rather than a runtime
   rejection 15,000 times. This is the version that closes the door.
3. **Re-classify a validation refusal as permanent** in `DeliverReply`, so it stops being reported
   as transient.
4. **Do NOT widen the `is_error` exemption** to let these through — see §5.3.

### ⚠ This WEAKENS my own §4 candidate, and that must not be glossed over

§4 proposed this as the mechanism behind `bugs_open/213` §D (items completing with another
agent's payload). **The located cause makes that link LESS likely, not more.** If the parent is
notified of *failure*, the expected outcome is an errored/needs-review item — **not** a `complete`
item carrying a foreign but well-formed payload, which is what §D actually shows. Either the
parent's failure handling completes the item anyway, or §D has a different cause.

**§4 is therefore downgraded from "strong candidate" to "open, and now doubted by its own
evidence".** It stays recorded because the two may still meet further downstream, but nobody
should carry it forward as the explanation. `bugs_open/213` §D is updated to say the same.

---

## 10. CONTRIBUTION 2026-08-15 from the `bugfix_213` lane (the filing lane, now handing over) — the mechanism COMPLETES, and it is worse than §9

**I am not working this bug — the owner has it with another thread.** This is the rest of the
read, handed over rather than acted on. Three additions, each read at source.

### 10.1 The parent ALWAYS hears failure and NEVER hears success — the validator's own exemption is why

`ResponseHeaders.ToMap()` (`types/context.go`) emits `headers["sender_agent_type"] =
rh.Sender.AgentType` and `headers["is_error"] = fmt.Sprintf("%v", rh.IsError)`.

- **Success path** — `notifyParentOfSuccess` (`coordinator.go:3709`) never sets `Sender` and never
  sets `InResponseToStepName`, so both required headers are empty and `is_error` is `"false"`.
  The validator refuses it. **Dropped.**
- **Failure path** — `notifyParentOfFailure` (`coordinator.go:3954`) builds headers with the **same
  two fields missing**, but sets `IsError: true`. `ProduceWithValidation` has an explicit
  exemption: *"Check if it's an error message - those we always send"*. **Delivered.**

> **So the defect is not "a reply sometimes fails to send". It is that on this path the parent
> receives the bad news and never the good news, by construction** — the same malformed envelope
> is refused when it says "success" and waved through when it says "failure". That asymmetry is
> what makes ~15,000 events invisible rather than merely noisy.

### 10.2 The false failure is classified RECOVERABLE, so the parent REPLAYS work that already succeeded

`notifyParentOfFailure` decides its status by prose (`coordinator.go:3943-3947`):
`perrors.RetryDisposition(errors.New(errorMsg))` → `error_recoverable` when the text reads
transient. The `errorMsg` on this path is *"workflow completed but its result could not be
delivered to the parent **(failed_transient)**: message validation failed"* — which contains the
word the classifier is looking for, **because `DeliverReply` labelled a deterministic validation
refusal "transient" (§9's second defect)**.

So the two defects compound: a permanent envelope fault is labelled transient, that label is then
read by a prose classifier, and the parent is told to retry. **The child's successful work is
executed again.**

### 10.3 This is where 274 meets `bugs_open/216`, and it reframes both

`bugs_open/216` fixed the arm that **refused** a recoverable replay (FIXED + LIVE `v1.0.1266`,
2026-08-08, proven by induction — a correct fix, not in question).

**Before that fix, 274's manufactured recoverable failures went nowhere — the replay was refused,
so the duplicate work never happened.** After it, the replay reaches the wire. So:

> **216's correct fix converted 274's silent losses into real duplicate execution.** Neither bug
> is wrong about itself; the interaction belongs to neither file, which is why nobody has owned
> it. 216's fix is also load-bearing far more often than its own file suggests — a large share of
> the recoverable responses it now replays are 274's fictions.

**[UNQUANTIFIED]** how much duplicate execution this actually causes: I have not measured how many
of the ~15,000 events produced a replay that re-ran real work, and the honest instrument is the
parent side (`awaited_requests.retry_version` / replayed offsets), not this log. **That
measurement is the one I would do first**, because it converts 274 from "noisy" to a costed defect
— and on `page-rerender` (4,794 events) duplicate execution is page rebuilds.

### 10.4 What this does NOT establish

It still does not explain `bugs_open/213` §D (items completing with a foreign but well-formed
payload). §9 downgraded that link and 10.1–10.3 do not restore it: a delivered *failure* still
predicts errored / needs-review items, not `complete` ones. **§4 remains doubted.** Treat 213 §D
as an open question with no current candidate.

---

## 10. FIX BUILT, 2026-08-15 — §9's candidates 1 + 3, plus the failure arm's envelope. OPEN until it rolls.

Picked up by a fresh lane (session `ae908c77`, "bugfix 274"). Everything in §9 re-verified at
HEAD first-hand before editing: the literal, the validator's five-field rule, `ToMap` emitting
no `step_name` key, and the onset — `git log -L :notifyParentOfSuccess:` shows the literal
byte-for-byte unchanged from `11394b220` (2026-01-11) through `2d976c026` (2026-08-03), which
swapped plain `Produce` → validated `kafka.DeliverReply`. **The headers were always empty; what
changed on 08-03 is that something started checking.** Still firing at pickup: 2,610 rows since
§9 was written, latest 07:59Z.

**What shipped (one commit, `Council-Submitted: 573526cd-7a4a-4f91-abf9-b44f866759d6`):**

- `notifyParentOfSuccess` sets `Sender` (new `ownIdentity` helper — the same five-field
  `AgentIdentity` shape the coordinator already stamps at four other sites, from
  `state.OwnerAgentType/ID/Role` + `s.podName`; OwnerAgentType is the RSH-009-resolved type) and
  `InResponseToStepName` (new `parentReplyStepName` helper — recovered from
  `__execution_context__`/`__work_request__`, both written once at state creation by the same
  `BuildCollectedData` that writes the two `__parent_*` keys this function already trusts;
  handles the post-DB map shape and the pre-persist typed shape). Empty recovery → an Error log
  naming this bug, then exactly today's behaviour (validation refuses, failure arm fires).
- `notifyParentOfFailure` gains the same two fields **plus `ClientID`, which it never set** —
  its reply passed the parent's incoming validation only via the `is_error` bypass (found by
  this lane's parent-side sweep; the produce itself is unvalidated so this is truthfulness,
  not deliverability).
- `platform/kafka`: `ProduceWithValidation` now returns a typed exported
  `ErrMessageValidationFailed` (text unchanged), and `DeliverReply` classifies it
  **`FailedUndeliverable`** — §9's "second, smaller defect": a deterministic refusal was
  labelled `failed_transient` ("try again") for 12 days. Degrader does not run (refused for
  headers, not size). ADP-017's register entry updated in the same commit.
- Tests: `platform/orchestration/reply_headers_validation_test.go` runs the **real**
  `validation.NewValidator` over the headers the coordinator **really** produced (the 158
  tests' double deliberately skips validation — which is exactly how this bug stayed invisible
  to them), with a mutation control proving the refusal is reachable (blanked identity + removed
  step-name sources must FAIL the real validator); plus a `parentReplyStepName` table test, the
  failure-arm envelope test, and `TestDeliverReplyValidationRefusalIsUndeliverable` in
  `platform/kafka`.

**Why §9's candidate 2 (a constructor making the five fields compulsory) was NOT taken:**
15 `ResponseHeaders{}` literal sites share that type across `platform/` and `internal/`; making
construction compulsory is an architecture-scope refactor of a shared seam for which this bug
is thin justification — and the census (this lane, 2026-08-15) found `coordinator.go:3710` is
the **only** site that is validated + non-error + missing fields, because the validator is
injected in exactly one place (`agentbase/agent.go:311`). The validated-headers test closes the
same door at the seam that actually has a validator. Two latent near-misses recorded rather
than fixed: `agentbase/agent.go:1868` (`SendInitializationResponse` — validated, non-error,
fields set, but a runtime-empty `spawnRequest.Headers.StepName` would reproduce this exact
refusal) and `internal/adapters/imagegenerator/dynamic_adapter.go:742` (`sendErrorResponse`
sets `Status:"error"` but not `IsError`, so it is error-path by intent and same-bug-shaped by
field values if a validator is ever injected there).

**Parent-side blast radius of populating the headers: none** (this lane's sweep). Replies are
matched by `in_response_to_request_id` alone (claim SQL `WHERE request_id=$1 AND
status='waiting'`); the two headers are unread on the response path except log strings, dead
code, and `persistReportedConditions` — which reads the JSON envelope's `Sender.AgentType`
only when a body carries `reported_conditions`, i.e. a designed, sanctioned-gated mechanism
that was dark only because the reply never arrived.

**Parent-side damage, measured at fix time** (adds to §2's child-side census): since 08-03,
16,869 child-side "could not be delivered" rows; parent-side `CHILD_ORCHESTRATION_FAILED`
788 at severity `error` (= `routeToErrorStep`) + 34 `fatal` (= cascaded `failWorkflow`, the
fabricated failure climbing a level) + unrecorded `continue_on_error` loop-skips. The
fabricated message contains "validation" → the permanent-needle classifies it
`error_unrecoverable`, so **no parent ever retried**. And per 213 §D's live repro (2026-08-14
20:15Z), a parent completing a work item after this failure substitutes a foreign payload —
so this defect is also the *trigger* for that class of false completions (the substitution
itself is 213's lane, not fixed here).

**Verify after the next roll (fixed-AND-LIVE is the closing bar):**
1. Provenance, per service: `kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300
   | grep -m1 'build provenance'` → `git merge-base --is-ancestor <this fix's commit> <stamp>`.
2. Effect: `SELECT count(*) FROM agent_error_log WHERE error_message LIKE '%message validation
   failed%' AND occurred_at > '<roll time>';` → ~0. **With a demand control** (a post-fix zero
   needs demand): completed child orchestrations in the same window > 0, e.g.
   `SELECT count(*) FROM orchestration_states WHERE status='COMPLETED' AND
   parent_orchestration_id <> '' AND updated_at > '<roll time>';` and the child-side Info line
   `Successfully notified parent of workflow completion` present in chassis logs.
3. Any residual rows post-roll now name their field: this fix's empty-step-name Error log cites
   bugs_open/274 directly.

**Unrelated but observed at fix time, so my commit is not blamed for it:**
`TestDefaultBeatsTheRecursiveSearch` (`platform/orchestration/datahelpers`) fails at clean
HEAD (verified via `git archive HEAD` in a scratch dir) — pre-existing, referenced by the
209/231 lane's notes, not touched by this fix.

> **CORRECTION 2026-08-15 (minutes later): "one commit" above became TWO, by a same-file
> passenger event.** While this lane's coordinator edits were dirty, the RFC_012 park lane
> committed `coordinator.go` by pathspec and took the envelope half of this fix with it —
> `3ba384c63` carries `ownIdentity`, `parentReplyStepName` and both `notifyParent*` literals;
> `919cc6976` carries the platform/kafka classification, the tests and the docs, and its
> message states the split. Forward-only: nothing lost, HEAD verified green with both halves
> (`git archive HEAD` build + both test packages pass). For "did my fix ship?" use
> `git merge-base --is-ancestor 919cc6976 <stamp>` — it is the LATER commit, so it implies both.

### Council verdict, 2026-08-15: APPROVED round 1 — 9 reviewers, 8 abstained, advisory objections dispositioned

Correlation `573526cd-7a4a-4f91-abf9-b44f866759d6`, seat `review_architecture` fired (kafka
core). Every objection advisory, none high-severity. Dispositions, with the queries the
reviewers asked for:

- **editquality: "the CollectedData keys are inference, not a symbol read."** They were read
  before the plan and re-proven after: `data_helpers.go:833` writes `__execution_context__`
  (serialised via the struct's `step_name` JSON tag) and `:975` writes
  `"step_name": execCtx.StepName` inside `__work_request__`; the committed table test
  (`TestParentReplyStepNameRecovery`) exercises the post-DB map shape and passes. The
  Owner* fields are `state.go:117-119` DB columns — the committed code compiles against them.
- **reuse: "why a new type-switch instead of extending `extractReplyToMetadata`?"** It is an
  UNEXPORTED symbol in the sibling `actions` package; the coordinator calling it needs an
  export plus a new orchestration→actions dependency edge, inverting this tree's layering
  (actions depends on orchestration's types, not vice versa — the same cycle constraint that
  moved the 217 classifier into `platform/errors`). The 15-line local helper is the smaller
  footprint. **Named follow-up A:** consolidate the two type-switch readers if a shared home
  below both packages appears.
- **reuse: "a fifth AgentIdentity shape floating beside four unconsolidated literals."** Fair,
  and deliberate: the approved plan scoped the four existing literal sites OUT to keep a core-
  plumbing diff reviewable. **Named follow-up B:** switch the three identical literals
  (coordinator.go ~1507/1590/3414) to `ownIdentity`; the fourth (~2921) omits Role and needs a
  one-line behaviour decision first.
- **guardian: "webscrape-unreachability is a claim I want checked."** Query, not assertion:
  `grep -rn NewProducerWithValidator --include=*.go platform/ internal/ cmd/` → exactly one
  caller, `agentbase/agent.go:311`; webscrape builds via `kafka.NewProducer`
  (`adapter.go:79`) = nil validator = `ProduceWithValidation` is a passthrough there.
- **guardian: typed-error surface widening / third core-file edit this week** — recorded;
  contained by `errors.Is`, and the gate round itself is the precedent check.
- **debug_historian: "name the post-deploy check, not just unit-green."** It is §10's verify
  block above: build-provenance stamp + `git merge-base --is-ancestor 919cc6976 <stamp>` per
  service, the ~0-row query with its demand control, and the named Error log for any residual.
  (The pod-grep-for-symbol lore the seat quotes was retired 2026-08-11 — the binary now states
  its commit; ask it, don't grep for markers.)

`919cc6976` carries `Council-Submitted:`; 098 credits it automatically now the verdict is
approved. No REVISE required; follow-ups A and B are recorded here rather than smuggled into
the approved diff.

---

## 11. LIVE + PROVEN, 2026-08-15 ~10:45Z — CLOSED

Fleet roll at **10:14Z** (chassis pods 10:14:10/10:14:31Z). All four §10 verification arms ran:

1. **Provenance:** the running chassis states `build provenance git_commit=0115f2b4528b0063…`
   (startup line, pod `agent-chassis-7779f5d998-96lpf`, 10:14:35Z), and
   `git merge-base --is-ancestor` confirms BOTH halves are ancestors: `919cc6976` ✓ and the
   passenger-carried `3ba384c63` ✓ (the stamp is 27 commits ahead of the fix).
2. **Effect, demand-controlled:** pre-roll baseline **293** failures in the two hours before
   the roll (~2.4/min). Post-roll: **8** rows, all inside 10:17:58–10:23:11 — and every one
   still labelled `failed_transient`, the label this fix RENAMED, so they are self-datingly
   pre-fix binaries draining out (short-lived spawned pods created before the roll). Since
   10:24: **zero for 21+ minutes** (~50 expected at the old rate), while **23 child workflows
   with parents COMPLETED** in that same window — across exactly the former top-failing types
   (internal-link-resolver 5, build-dispatch-loop 4, page-content-writer 3, page-build-handler
   3, page-rerender 2…). The zero has demand.
3. **Positive signal:** `Successfully notified parent of workflow completion` — the log line
   that only prints when `DeliverReply` reports `Answered()`, unreachable on this path for any
   child-with-parent since 08-03 — present in 4 live spawned pods
   (`agent-build-dispatch-loop-483fe901-4vts8` et al.).
4. **Parent side:** `CHILD_ORCHESTRATION_FAILED` + "could not be delivered" rows post-roll:
   **0**.

One measurement lesson while verifying: the first positive-signal grep ran against
`-l app=agent-chassis` and returned 0 — the completing children run in their own spawned
`agent-<type>-*` pods, not the chassis deployment (the standing wrong-service landmine). The
line was found once the right pods were asked.

**Closing bar met: fixed AND live AND proven with demand.** File moves to `bugs_closed/`.
Residual watch (non-blocking): §10's named follow-ups A (shared step-name reader) and B
(identity-literal consolidation), and the two latent near-misses recorded in §10.
