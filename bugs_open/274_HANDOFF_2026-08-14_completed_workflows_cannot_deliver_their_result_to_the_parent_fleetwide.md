# 274 — completed workflows cannot deliver their results to their parents: ~15,000 failures across 60 agent types since 2026-08-03, and the parent completes the work item without the child's payload

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
