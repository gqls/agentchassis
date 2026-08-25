# HANDOFF — the two questions that outlive bug 327

**Written 2026-08-25, as `bugs_open/327` closes.** The lane's work is finished; these two are
judgement calls that nothing depends on. **Neither is blocking anything, and neither decays** —
they can sit for weeks. This file exists so they are not silently forgotten, which is the usual
fate of a question left in a closed bug.

Background, if you have none: `docs/agent_docs/docs024_key_docs_latest/bugfix_327_silent_publish_drop/`
— `SUMMARY_*` for the story, `HANDOFF_2026-08-24_continue_here.md` for the state.

---

## Decision A — run a council review on this work?

### What the thing is

The **council gate** is an advisory panel of reviewer agents. You submit a change with its
rationale; **17 seats** exist, of which two always run and the rest fire only when your edited
paths match their footprint. It records a verdict; it cannot block anything. `[MEASURED
2026-08-25]` **56 completed rounds average 9.2 minutes**, though dispatch queues behind the fleet
so budget ~30 minutes wall-clock.

### The rule that governs it

Platform-code changes are expected to go through it. Scope is `platform/`, `internal/`, `pkg/`,
`cmd/config-key-audit/`, **`scripts/pattern-check.py`** (added 2026-08-24), plus appliable
migrations. Anything else is refused client-side and spends nothing.

### How this case measures against it

**On 2026-08-23 the submission was REFUSED (exit 2)** — every edit was under `scripts/`. **On
2026-08-24 the same file was ADMITTED (exit 0)** because you widened scope to include
`pattern-check.py`. So this is no longer a question of overriding a refusal; it is simply whether
the credits are worth it.

**The catch, and it is the whole decision.** Of the three edits in the submission, **only one is in
scope** — the detector (`check_kcat_stdin_race` in `pattern-check.py`). The two that carry the
risk, `kafka-publish-lib.sh` and the migrated trigger, are still out of scope. **A round would
therefore review the detector, not the publisher.** The seats would see the whole plan, but the
thing that justifies the spend is the smallest and least dangerous part of it.

### ⚠ You can review the publisher ON DEMAND — the scope gate is not the only door

**Owner question, 2026-08-25: "is there a way we can only call on the publisher if we think we
need it?" Yes — `FORCE=1`.** It overrides the scope refusal for one submission and nothing else:

```bash
FORCE=1 ./…/097_TRIGGER_council_review_v1.sh <submission.json>
```

So the publisher does **not** need permanent scope to be reviewable. It can be brought to the
council the day there is a reason — say, if a drop is ever observed through it (which is also
Decision B's trigger). **That is the better shape than widening scope**: a permanent widening
taxes every future `scripts/` change, while `FORCE=1` costs exactly one round, on the day you
want it. Decision recorded 2026-08-25: **run without the publisher for now, revisit if the current
setup proves insufficient.**

### What the scope gate is actually conserving — three different things, and only one is big

Asked directly, so answered precisely rather than by feel.

1. **CREDITS — the primary reason, and the one the gate states.** `council-scope.sh` refuses
   out-of-scope submissions *client-side*: they "never spend credits". Cost is then
   **relevance-gated** a second time — two seats always run, and the rest fire only when your
   edited paths match their footprint, so an admitted submission does not automatically pay for
   all **17**.
2. **SIGNAL QUALITY — smaller, but the reason it is a scope rule and not a budget.** `RFC_005`
   rejected putting prose through the gate partly because it would "dilute the architecture
   seat's signal for the platform-code case that most needs it". A seat that fires constantly on
   things it cannot judge becomes a seat people stop reading. **This is about a seat's signal
   across many runs, not about noise inside one round.**
3. **TIME — real but minor.** `[MEASURED 2026-08-25]` 56 completed rounds average **9.2 minutes**;
   dispatch queues behind the fleet, so budget ~30 minutes wall-clock.

**What it is NOT conserving: "too many members arguing".** That was a reasonable guess and the
mechanism does not work that way — irrelevant seats do not fire at all, so an extra submission
does not add voices to a debate. The decision rule also weighs objection **severity** (fixed
2026-07-22), which is what made APPROVED reachable at all (~5% → ~80%). More seats firing means
more coverage, not more argument.

**So the honest one-line answer: you are mostly saving credits, secondarily protecting a seat's
signal over time, and barely saving time at all.**

### What I would say

**No recommendation — this one is genuinely balanced**, and I would rather say so than manufacture
a preference.

- **For:** the detector is the piece that decides whether every session's commit draws a warning,
  and a false positive there is, in the pre-commit hook's own words, a fleet-wide commit outage.
  That is worth a second opinion, and it is exactly the class you widened scope to catch.
- **Against:** the detector was already measured against 300 real commits before being wired in
  (1.7% fire rate, 5/5 true positives, controls both ways). A round would mostly re-derive that.
  And the artefact a reviewer would most want to see is the one they cannot.

**To run it — everything is ready, nothing to prepare:**

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/bugfix_327_silent_publish_drop/COUNCIL_SUBMISSION_2026-08-23_publish_receipt.json
```

Save the printed `SUBMISSION_CORR`. `DRY_RUN=1` in front validates and spends nothing.
Verdicts: `SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts WHERE
correlation_id='<corr>' AND kind='council_report' ORDER BY created_at;`

> ⚠ If you do run it, note the trigger itself was migrated by this lane — it now **asserts its
> own publish**, so a dispatch that silently vanishes is no longer possible. That was not true
> when the submission was first written.

---

## Decision B — build the in-cluster Go submit path?

> **DECIDED 2026-08-25: wait**, on the trigger below. Recorded so nobody re-opens it without one.

### What the thing is

Today a script publishes by starting a throwaway `kcat` container through `kubectl`, and proves
it worked by **scraping a marker out of the pod's output**. That is a receipt, and it works — but
it is a receipt read over `kubectl attach`, not an acknowledgement from the message broker itself.

The alternative is a small in-cluster publisher in Go. `platform/kafka/producer.go` is **already**
`RequiredAcks: kafka.RequireAll, Async: false` (`:63-65`) — the broker's own acknowledgement *is*
the success path there. Nothing to invent; it exists and the services use it.

### Why it is better, precisely

Two things the current fix cannot do:

1. **It makes a silent drop UNREPRESENTABLE rather than detected.** The receipt tells you a publish
   failed. A broker ack means the publish cannot succeed-without-sending in the first place.
2. **It could write a durable submission record at publish time**, which closes a real hole:
   `[MEASURED 2026-08-23]` `orchestration_states` retains **~2 days** while `agent_error_log`
   retains ~30. So today, 48 hours after the fact you can still learn whether a message was
   *refused* but never whether it *arrived*. A dropped submission becomes permanently
   undiagnosable almost immediately.

### What it costs

Go code, a council round (it would be **in scope**, unlike the shell library), an image build and
a fleet roll — plus, per the 2026-07-29 ruling, registration as a shared seam and telling the
other consumers. It also cannot be reached from a shell script without something to call, so it
implies a small client or an HTTP surface on an existing service.

### What I would say

**Not now — and there is a specific trigger to revisit, which matters more than the
recommendation.**

The receipt closes the live exposure: no live script can now exit 0 having sent nothing.
Layer 3 buys elegance and forensics on top of that, and costs a roll.

**Revisit the moment a message is observed lost THROUGH the library** — i.e. a
`kafka_publish_checked` that returned 0 with no orchestration row and no `agent_error_log` entry.
That would be evidence the `kubectl attach` receipt is not sufficient, which is the one thing
that would make this urgent rather than nice. **Nothing else should trigger it** — not tidiness,
and not the fact that it is the "proper" design.

Filed as proposed-and-unbuilt in the concept register under **OPP-009**'s `verify-later`, so it
is discoverable by anyone searching for the mechanism rather than only by reading this file.

---

## If you do neither

Nothing degrades. The detector keeps the class from growing, the library is adopted by 21 callers
(two of them lanes nobody asked), and the dormant scripts are caught at commit time if anyone
picks one up. **Both of these are improvements to an already-closed problem, not loose ends.**
