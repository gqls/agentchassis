# REVIEW — council round 1 on the retry-replay contract: REJECTED on SCOPE

**Submission correlation:** `75cb2fdc-e74c-4d3d-99b7-9264548e65d6`
**Verdict:** `rejected` · `decided_by: "hard veto from guardian"` · `unreadable: 0` ·
`abstained: 6`
**Code under review:** commit `eb70c3dd3` (+ the follow-up in this workstream's
next commit, which answers two of the objections).

`unreadable: 0` — so this was **a judgement, not the harness**. `abstained: 6` is
the relevance filter skipping seats, which is normal and says nothing about
health. Read those two, never `abstained` alone.

---

## The tally

| seat | verdict |
|---|---|
| **guardian** | **VETO** — scope/venue |
| editquality | object (high: the migration was missing from the *edit list*; low: framing of edit 3) |
| bug_historian | object (medium: back-fill UPDATE ignored rows-affected; low: legacy variants unproven) |
| debug_historian | object (medium: no post-deploy pod-verification named in the plan) |
| prior_art_librarian | approve (+ low objection: legacy variants asserted, not measured) |
| guidelines | approve |
| reuse_agent | approve |
| constitution | approve |
| mission | approve |
| diagnosis_guardian | approve (out of charter) |

**Six approvals. No seat disputed the diagnosis.** The guardian says so itself:
*"This is a well-measured plan and I don't doubt the diagnosis — the parent/child
id collision is real and the blast-radius arithmetic was done before submission,
which is exactly the discipline this seat wants to see."*

## What the veto is actually about

Not the fix. The **venue**:

> *"the plan itself states the change is architecture-scope … and then proceeds to
> implement it as a single submission touching a new shared contract, a schema
> column gating five readers, two exported-signature changes in the actions
> package, and three separate edits inside coordinator.go. That is the textbook
> pattern (b) exists to catch, self-declared rather than disguised, but
> self-declaration doesn't change the venue it belongs in."*

It is right, and it is the same finding as `bugs_closed/124`'s `$ctx.` veto: a
shared mechanism arriving inside a bug patch. Declaring it does not relocate it.

**Per the owner ruling of 2026-07-28, this is NOT answered by resubmitting with
better measurements.** A SCOPE veto is a judgement about *how* a capability
reaches production. So: recorded here, routed to a human, not re-litigated with
the council.

## The seats contradict each other on the remedy — which is why a human is needed

The guardian's **safest contained alternative**:

> *"ship edit 3 alone (the MISROUTED_REQUEST guard in handleOrchestrationStatus)
> as the immediate, single-function, no-schema-change fix — it directly stops the
> reported symptom … Land that now, and bring the full replay-payload/ReplayRequest
> contract to an actual architecture review track."*

But **editquality objects to edit 3 in precisely the opposite direction** — that
it must not be allowed to read as the primary fix:

> *"its rationale should more clearly distinguish 'defence in depth after the
> sender is fixed' from 'the child-side patch the diagnosis dismissed,' since a
> careless reviewer could read it as reintroducing the rejected fix as primary."*

And **constitution approves specifically because edit 3 is not a substitute**:

> *"The added MISROUTED_REQUEST guard … is transparently framed as defence-in-depth
> on a separate (request) code path, not as a substitute for the structural fix."*

So the guardian's contained alternative is the thing two other seats approved the
plan for *not* doing. That is exactly the "seats disagree with each other" case the
owner ruling says a human must break.

**And on the merits the contained alternative does not fix the bug.** Edit 3 alone
converts a silent swallow into a loud error. The parent still times out; the work
still does not happen. It removes the *silence*, not the *failure* — and it leaves
every retry in the fleet still carrying the wrong identity, an empty body and the
wrong action. It is a good half-measure and it is a half-measure.

## What I did with the checkable objections

A SCOPE veto is not answered by measurements — but the *other* seats' objections
are checkable, and leaving them unanswered would be a separate failure.

**1. editquality (high) — "no edit creates migration 263".** Correct about the
**submission**, wrong about the change: `263_awaited_requests_request_payload.sql`
was written and committed in `eb70c3dd3` and had already been applied and recorded
before the round returned. The omission was mine, in the ≤8-edit list. Worth
recording because the reviewers cannot open files — **an edit list that omits a
file is, to them, a plan that omits it.**

**2. bug_historian (medium) — the back-fill UPDATE discards rows-affected.**
A real defect, now fixed. The guarded `UPDATE … WHERE request_payload IS NULL`
succeeds with `err == nil` when it matches nothing, so the error alone could not
tell "backed-fill" from "did nothing". It now checks rows-affected and, on zero,
distinguishes the benign case (a payload is already recorded) from a real gap
(`RETRY_PAYLOAD_BACKFILL_MISSED`). Exactly the 016b §9 pattern the seat cited.

**3. prior_art_librarian / guidelines / bug_historian (low) — the legacy spawn
variants are asserted unreachable, not measured.** Measured now, and the answer is
better than asserted:

- `SpawnAgentActionOld` and `SpawnAgentActionOld2` **are not in the action registry
  at all** — they are unreachable Go functions, dead code, not a second router
  branch. `spawn_agent_k8s` is a deprecated *alias* whose `Handler` is
  `SpawnAgentAction`, i.e. the one that was wired.
- Across **every** `agent_definitions` row — unfiltered: no `is_active`, no
  snapshot, no `deleted_at` filter — the only spawn/call actions present anywhere
  are `spawn_agent` (93) and `call_agent` (80). `spawn_group` and
  `spawn_agent_k8s` appear zero times.

**4. debug_historian (medium) — no post-deploy pod-verification named.** Fair, and
it was in `RUNBOOK_retry_replay.md` rather than in the plan. The discriminating
check is a string this change **deleted** — `is_retry`, the old stub body — with a
positive control: **v1.0.1192 has it (1) and none of the new markers (0); v1.0.1193
has the new markers (1 each) and not `is_retry` (0).** Both directions verified on
the built binaries.

## The one thing the round made me correct in my own claim

My submission said `call_agent`/`spawn_agent` coverage was **complete**. It is not,
and the seats' pressure on the coverage question is what made me re-ask it properly.

The census I ran answered *"which spawn/call actions are seeded"*. The question
that matters is *"which actions await a response"* — a different question, and my
check had silently encoded the first one.

Re-measured on the retried population (14 days), split on whether `call_agent` or
`spawn_agent` produced the request:

| coverage | retried | exhausted anyway |
|---|---|---|
| `call_agent`/`spawn_agent` — **wired** | **422** | 289 |
| adapter — re-executes the step, untouched | 0 | 0 |
| **other awaited sender — NOT wired** | **6** | 4 |

The six are `scrape_pages` (3, all 3 exhausted) and `search_web` (3, 1 exhausted).

**They are not a coverage miss; they are a different defect.**
`web_search_action.go:139` puts `params.ExecutionContext.OrchestrationID` — the
**caller's own** id — on the *original* outbound message, not just on a retry. So
there is no child identity to replay, and wiring them would simply make
`RETRY_SELF_ADDRESSED` fire on their originals. That is its own mechanism and its
own diagnosis; bundling it here would be the very thing the guardian vetoed.

**Honest statement of the behaviour change:** those six now fail fast instead of
retrying. Four of the six exhausted their budget anyway, so the practical delta is
**at most two requests per fortnight** losing a retry that might have worked, in
exchange for a named error instead of silence. Recorded as a follow-on, not fixed
here.

## Options, costed — for the owner

**1. Deploy the full change (v1.0.1193, already built and pushed), and route the
seam to architecture review in parallel.**
The bug is fixed fleet-wide. Contradicts the guardian's venue ruling, and repeats
124's precedent — the thing the 07-28 ruling exists to stop.

**2. Take the guardian's contained alternative: revert to edit 3 only.**
Cheap and no schema change. But it does not fix the bug: the parent still times
out and every retry still carries the wrong identity. It converts silence into a
loud failure, which is worth something and is not a fix. Also needs the revert of
work that six seats approved.

**3. (Recommended) Hold the deploy; the code, the migration and the register entry
stay as they are; the seam goes to architecture review as its own item.**
Nothing is live, so nothing regresses. Migration 263 is additive, nullable and
inert against the current binary — verified. The image is built, pushed and
pod-grep-discriminated, so the deploy is one command once a human rules.
Cost: the defect keeps costing ~430 poisoned retries a fortnight until then.

I have taken **option 3** and have not rolled the fleet. A chassis roll is
fleet-wide and outward-facing, and the owner ruling is explicit that a SCOPE veto
is broken by a human, not by the thread that drew it.

## Precedent worth keeping

**Write the migration into the EDIT LIST, not only into the rationale.** Reviewers
cannot open files. The single highest-severity objection in this round was raised
against a file that existed, was committed and was already applied — because it
was described in prose and absent from the eight edits. That is a submission
defect that will recur, and it is free to avoid.

---

## ADDENDUM 2026-07-28 ~22:15 — option 3 was not available, and that is a finding about the RULING

I chose option 3 above: hold the deploy, let a human break the veto. **That option
does not exist on this tree, and I did not realise it when I chose it.**

The fleet rolled to **v1.0.1194 at 20:48:11Z**, fired by another session for its own
change. `make build-<service>` builds from committed `HEAD` — deliberately, so a
build cannot bundle anyone's WIP. My commits (19:57Z, 20:16Z) were in `HEAD`. So they
went out. Verified by pod-grep on both pods, not inferred: `is_retry` **0**, all four
new markers **1**.

Nobody did anything wrong. The build rule worked exactly as designed, and the session
that rolled had no way to know it was carrying a vetoed seam — the tree gives it no
signal, and reading every commit since the last roll is not a thing anyone does eight
times a day.

### Why this matters beyond this bug

**The platform-seam ruling of 2026-07-28 assumes the committing thread controls when
its seam reaches production. On this tree it does not.** HEAD is shared, builds come
from HEAD by design, and the fleet rolled ~8 times on 2026-07-28. Therefore:

> **Committing IS shipping — on someone else's schedule.**

That makes "ship it ahead of review only under a stated ordering constraint" only
half enforceable. It governs the thread that *writes* the seam, and there is nothing
it can do to comply beyond not committing — which CLAUDE.md separately forbids,
because a long-lived dirty tree is shared mutable state another session will sweep.

The two options actually available to a thread told to hold a seam back are:

- **(a) don't commit it** — forbidden, and unsafe for the reasons CLAUDE.md gives;
- **(b) commit it behind a flag/config switch that defaults OFF** — the only one that
  works, because it decouples *in the binary* from *in effect*.

**Nothing in the current ruling asks for (b).** If the owner wants seams genuinely
held pending architecture review on a tree that rolls this often, the ruling needs a
third clause: *a seam that has not passed review ships DARK — behind a default-off
switch — or it does not get committed to a shared HEAD.* That is a change to the
ruling, not to this bug, and it is the owner's to make.

### What this does to the three options

- **Option 1 (deploy now, review in parallel)** has effectively happened, by
  accident rather than by decision. The code is live and behaving (capture proven on
  live traffic; the self-addressed invariant returns 0; payloads are ~1.1 KB).
- **Option 2 (revert to the guardian's contained alternative)** is now a *revert of
  live code*, not a decision not to ship. Materially more expensive than it was two
  hours ago, and it would restore a defect that has been measured at 430/430.
- **Option 3 (hold)** is retired — it was never available.

**My recommendation is unchanged in substance and now cheaper to act on: leave the
code live, and route the seam to architecture review on its own merits.** That is
also exactly what `bugs_closed/124`'s owner ruling did with the `$ctx.` veto — *"the
code stays and the precedent gets fixed"* — and this is the second instance in one
day of the same shape. Two instances is the point at which the precedent is the
thing worth fixing, not the individual case.
