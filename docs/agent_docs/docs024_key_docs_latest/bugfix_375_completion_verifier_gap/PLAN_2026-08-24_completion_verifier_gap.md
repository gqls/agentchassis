# PLAN — `bugs_open/375`: the unguarded completion writer

**Started 2026-08-24 ~21:00Z.** First session in this lane; the lane was handed over by
`HANDOFF_2026-08-24_start_here.md` (written by the closed `bugfix_367_router_remit` lane).
The bug is CLAIMED for this directory as of commit `e0e80b65f`.

---

## 1. What the bug is, in plain terms

A **work item** is a row recording one defect on one site — "this component has empty
required fields", "this page needs a hero image". A **handler agent** picks the row up, tries to
fix it, and something then stamps the row `complete`.

A **verifier** is a per-defect-type re-check that runs *just before* that stamp. Its whole job is
to catch a handler that reported success without actually fixing anything: it re-runs the defect's
own predicate and refuses the completion if the defect is still there.

There are two pieces of code that stamp `complete`. `CompleteWorkItemAction` asks the verifier
first. `UpdateWorkItemStatusAction` does not — it has never called `GetVerifier` at all. So which
guard a completion gets depends on which of the two a workflow happened to be configured with,
and nothing anywhere says so.

## 2. What the measurement says (and why it changes the fix)

Re-run against the live DB **2026-08-24 ~21:00Z** — see `NOTES` for the queries and controls.

- **4** live agents (of **200**) reach `complete` through the unguarded writer, across **6** steps.
- Those 4 agents handle **5** item types, and **none of the five has a registered verifier**.
- So **zero verifiers are being bypassed today.** The defect is LATENT.

That means this is not a live-damage fix. It is a **trap disarmed before it is sprung**, and the
trap has a name on it: `verifier_coverage_test.go` maintains a backlog of types that *should* get
verifiers, and two of these five are on it. Whoever works that backlog will register a verifier,
watch the coverage test go green, and protect nothing.

**It is worse than the handoff said, and I measured the extra part.** The concept register's
`CQ-023` entry carries a landmine telling that person what to expect:

> *"a verifier later registered for `required_fields_missing` (RegisterVerifier) would fail-closed
> the `converted` arm's completion"*

That is **false today**, and false *because of this bug*: `close_converted` is an
`update_work_item_status` step, so no verifier would ever be consulted and nothing would
fail-close. The next person is therefore primed to expect a specific wrong outcome and will get a
different wrong outcome — a silent no-op. Correcting that entry is part of this lane's work.

## 3. The change, in three parts

**Part A — the fix (bug file candidate 1).** `UpdateWorkItemStatusAction`'s `complete` arm gains
an **opt-in step-config key, `verify_before_complete`, whose unsafe default is OFF**. When armed,
the arm runs the *same* registered-verifier gate the guarded writer runs and, on refusal, routes
through the *same* `failUnverifiedCompletion` attempt machinery. Absent — which is every live step
today — behaviour is byte-identical.

Shape per the owner's ruling of 2026-08-02 §2 (new authority on a shared seam ships as an opt-in
field, so the decision is visible to a reviewer of the CALLER). Arming is deliberately **per step**,
not per type, because `CQ-023` shows the decision is per close-path: the `converted` arm and the
`stale` arm of one router want different answers.

**Not architecture-scope, per `RFC_022`'s narrowing (owner ruling 2026-08-11).** All three
conditions hold and the third is *enumerated, not asserted* (§2 above and `NOTES`): opt-in ✓; the
unsafe side is the default ✓; **zero live consumers name it** ✓.

**Part B — what protects the next person (bug file candidate 4, sharpened).** An opt-in safety
field with an unsafe default rots unexercised unless something tells the next person to arm it.
So: when the `complete` arm runs **unarmed** and the item's type **does** have a registered
verifier, complete exactly as today but **record the bypass on the row** (`result._verification`,
status `verifier_not_consulted`) and log it.

Why this shape rather than a build-time roster: it needs no hand-maintained list, so it cannot go
stale by addition; it changes no completion outcome, so it carries no liveness risk; and it puts
the fact on a queryable surface, which is this estate's standard remedy for "the pod log does not
survive a roll".

**Part C — stop the two documents lying (bug file candidate 3).** `verifier_coverage_test.go`'s
header reads as though registering a verifier protects a type. Say that the promise is scoped to
the `complete_work_item` path, and name the arming key. Correct `CQ-023`'s landmine in place,
visibly, per CLAUDE.md's rule about stale register entries.

## 4. What this deliberately does NOT do

- **It does not unify the two writers** (candidate 2). That is the structural fix, it is squarely
  architecture-scope, and `bugs_closed/284` is the precedent for how it should be done. This change
  should make that easier, not substitute for it: after Part A the two paths share one gate
  implementation, which is the first half of unifying them.
- **It does not arm the field anywhere.** Arming is a per-type, per-arm decision with `CQ-023`
  attached, and it belongs to whoever writes the verifier.
- **It does not write any verifier**, and takes no position on whether the five types should have
  one.
- **It does not touch the `image_url_404` undispatched population** the handoff observed (§6 there).
  Different defect — routing, not guarding. Noted, not absorbed.

## 5. Decisions and their reasons

| decision | reason |
|---|---|
| opt-in, not default-on | default-on would fail-close `CQ-023`'s `converted` arm the day a verifier is registered, for consumers who never asked; and it would make the change architecture-scope |
| per STEP, not per item type | `CQ-023`: one router's arms want different answers, and the reviewer who can tell is the reviewer of the caller |
| gate 2 only (registered verifier), NOT gate 1b (no-change) | gate 1b reads the *handler's own reply payload*, which this action does not have — it has step config. Passing it something else would grade the wrong evidence, which is the exact error `complete_work_item_no_change.go`'s header records |
| reuse `failUnverifiedCompletion` rather than write a refusal | one definition of what a blocked completion does; a second copy is the drift class `bugs_closed/284` exists to stop |
| record the bypass instead of blocking it | zero liveness risk, no roster to go stale, and it is the half that protects the person the trap is set for |

## 6. Corrections to the brief I was handed

- The handoff's positive control reported "**12 of 13** types with real rows". Re-run, that is
  **12 (item_type, handler_agent) PAIRS from 10 distinct types** — three registered types
  (`orphan_element_refs`, `page_canonical_collision`, `revenue_shape_cta`) have no rows at all.
  The control still passes and the separation is still real; the number was a pair count read as a
  type count.
- The handoff (and §3a) said `complete_work_item` is named by 4 agents. A census restricted to
  `$.workflow.steps` finds only **2** — the other two name it from inside a nested loop-step config.
  **My first census had the same narrowness and I only caught it by re-running recursively.** The
  recursive scan confirms the `update_work_item_status` side is complete at 22 steps / 6 `complete`
  arms, so the bug's own number is safe; the neighbouring number was not.
