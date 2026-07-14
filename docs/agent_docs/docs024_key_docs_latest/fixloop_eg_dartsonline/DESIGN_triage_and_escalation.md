# Design — triage and the escalation path into the fix loop

*2026-07-14. How problems found across the platform reach the fix loop — and,
just as important, how most of them should NOT. Written in response to: "we have
handlers that pick up the work items emitted from the checkers; they may fail to
fix things." That failure is the hinge this whole design turns on.*

---

## The one idea

**The fix loop should be triggered by a handler FAILING to fix something, not by
a checker DETECTING something.** Detection is cheap and common; most detected
problems are fixed by an ordinary handler re-running. Only the problems that
*survive* their normal remedy are worth the loop's expensive, code-changing,
PR-producing attention — because a problem that resists its cheap fix is the
signature of a bug in the code, which is exactly what the loop exists to repair.

## The machine we already have

The platform already runs a detect-and-heal cycle:

1. **Checkers / audit agents** detect problems and emit **work items**
   (`needs_page`, `page_rerender`, `needs_content`, …) into `site_work_items`.
2. **Handlers** pick those up and try to fix them (build the page, re-render, …),
   recording `attempt_count`, `max_attempts`, and a `status`.
3. Handlers **can fail** — in two very different ways:
   - **Loud failure:** the handler errors, retries, and eventually lands at
     `status='failed'` (`attempt_count >= max_attempts`).
   - **Silent failure:** the handler reports success but the problem persists.
     This is the darts benchmark bug exactly — `page-build-handler` "completed"
     a page it never built. Silent failure is the dangerous one, because nothing
     in the work item looks wrong.

So the gap is not detection and not fixing — both exist. The gap is: **when the
fixing fails, what turns that failure into a code fix?** Today: nothing. A
failed work item sits at `failed`; a silently-failed one looks done. That is the
hole this design fills.

## The three tiers (where each problem belongs)

| Tier | Who | Job | Output |
|---|---|---|---|
| 1. Prevention | the build/other workflows | do the work correctly | built state |
| 2. Operational healing | checkers + handlers (the immune system) | detect, and apply the **known remedy** (re-queue, re-render) | fixed state / retries |
| 3. Structural repair | **the fix loop** | for problems whose cause is in **code/workflow**, produce a reviewed PR | pull request |

Tiers 1–2 exist and work. This design is the **bridge from tier 2 to tier 3**,
and the rule that keeps tier 3 from being flooded.

## The escalation path, end to end

```
checker detects ─▶ work item ─▶ handler attempts ─┬─ success ─▶ done (no loop)
                                                  │
                                                  ├─ FAILED (attempts exhausted) ─┐
                                                  │                               │
                                                  └─ silent-success (problem       │
                                                     re-detected next sweep) ──────┤
                                                                                   ▼
                                                                        ┌── diagnosis-triage ──┐
                                                                        │  filter · dedupe ·   │
                                                                        │  loop-worthiness     │
                                                                        └──────────┬───────────┘
                                    transient / retry-worthy ◀──── re-queue ◀──────┤
                                              human ◀──────────── ambiguous/risky ─┤
                                                                                   ▼
                                                               needs_diagnosis work item
                                                                                   │
                                                                                   ▼
                                              fix loop: diagnose ▶ plan ▶ council ▶ {PR | escalate-to-human}
                                                                                   │
                                                          fix merged + deployed ───┘
                                                                                   ▼
                                              re-queue the original items / re-run the checker  (VERIFY)
                                                                                   │
                                        still failing ─▶ back to triage      fixed ─▶ close the escalation
```

## The new piece: `diagnosis-triage`

A thin agent (per the constitution: thin workflow, logic in one Go action) that
runs **on a schedule** and does NOT diagnose — it **routes**. Each sweep:

1. **Gather escalation candidates** from `site_work_items`:
   - `status='failed'` with `attempt_count >= max_attempts` (loud failure), and
   - *(phase 2)* items re-emitted after a prior completion — the silent-failure
     signal (see "the hard case" below).
   - Exclude anything already escalated or already under an open diagnosis/PR
     (a marker + cool-down), so nothing is escalated twice.
2. **Apply the loop-worthiness filter** (already written, in the runbook): only a
   symptom about system behaviour, with a plausible code cause, that a cheap
   pre-check can't resolve. A single transient failure with retries left is NOT
   loop-worthy — it goes back to re-queue.
3. **Dedupe by pattern.** Fifty pages failing the same way through the same
   handler with the same error is **one** code bug. Triage groups by
   `(item_type, handler_agent, error-signature)` and files **one** escalation for
   the pattern, naming the count — never fifty. This is the single most
   important guardrail: it is what stops the loop being buried.
4. **Route** each candidate/group:
   - **→ fix loop:** write a `needs_diagnosis` work item (the loop's existing
     intake) whose symptom is the pattern — "handler X fails item_type Y on N
     sites; error: Z" — with pointers to the failed items as evidence.
   - **→ re-queue:** transient-looking; reset for another handler attempt with
     backoff.
   - **→ human:** ambiguous or high-blast-radius; flag, don't guess.

Triage owns **routing and rate-limiting**, nothing else. It never fixes and
never diagnoses — those are tiers 2 and 3. Keeping it thin keeps it trustworthy.

## Why this shape, not the alternatives

- **Not** wiring every checker straight into the loop. That floods a slow,
  expensive, PR-producing service with cheap operational noise. Detection volume
  is huge; genuine code bugs are rare. The filter belongs between them.
- **Not** one monolith that both scans sites and fixes code. Different skills,
  different outputs (a re-queue vs a PR), different risk. Keep them separate; let
  each tier do the one thing it's good at.
- **Instead:** a federation of specialised detectors and first-line fixers (the
  build immune system; the doc-traveller keeps fixing tools in its own domain)
  plus **one generalist root-cause-to-PR service** (the loop), joined by a single
  escalation channel (`needs_diagnosis`) with triage as its rate-limiter and
  router.

## The hard case: silent failure (and why it matters most)

Loud failures are easy — the work item says `failed`. **Silent failures are the
darts bug**: the handler completes, the item looks done, but the page is still
blank. Nothing in `site_work_items` flags it. Catching it needs one of:

- **(a) A verification checker** — after a handler completes, re-check that the
  problem is actually gone (page has sections / renders / exists). If not, emit a
  "remediation-ineffective" signal that triage treats as an escalation candidate.
  This is the highest-value addition, because silent success is precisely the
  class of bug the fix loop was built to catch and a checker cannot.
- **(b) Recurrence detection** — if the same checker re-emits the same problem
  after a completed handler run, that recurrence *is* the signal.

Either way, the principle holds: **the fix loop is fed by "the remedy didn't
actually work," not by "a problem exists."** The darts benchmark is the proof
case — a verification checker noticing guides-index is still blank after a
"successful" build would escalate exactly the bug the loop already diagnoses
correctly.

## The feedback loop (closing it honestly)

An escalation is not resolved when the PR merges — it's resolved when the
**original problem is gone**. So after a fix deploys, triage (or the verification
checker) re-queues the originally-failed items / re-runs the checker:

- **Fixed** → close the escalation; record the win.
- **Still failing** → back to triage. Either the fix didn't work or the diagnosis
  was wrong — and that, too, is information worth surfacing.

This makes the whole thing a closed loop with a truthful terminal, in the same
spirit as the council's "escalate rather than pretend."

## What exists vs what's new

- **Exists:** checkers, handlers, `site_work_items` with `attempt_count`/`status`;
  the `needs_diagnosis` intake + `diagnose` pipeline (the loop's front door,
  shipped disabled); the loop-worthiness doctrine; the fix loop itself; the
  awareness digest.
- **New (this design):**
  1. `diagnosis-triage` agent — scan → filter → dedupe → route (thin).
  2. A verification/recurrence checker for the silent-failure class (highest
     value; can start with one high-signal check, e.g. section-index pages that
     are `active` but have zero components).
  3. The dedupe + cool-down + escalation-marker bookkeeping on `site_work_items`.
  4. A new digest section — "escalations in / remedies applied / verifications" —
     so the owner sees the whole immune system, not just the loop, on one page.

## Suggested phasing

1. **Loud failures first.** Triage on `status='failed'` items only, with dedupe
   and a hard rate cap. Smallest safe slice; proves the escalation channel with
   real traffic and low risk.
2. **Verification checker for the darts class.** One high-signal silent-failure
   check, feeding triage. This is where the loop starts catching the bugs it was
   actually built for.
3. **Feedback close-out** — re-verify after a fix deploys; open/close
   escalations honestly.
4. **Digest section** for escalations — fold the whole immune system into the
   owner's one awareness surface.

Each phase is independently useful and independently reversible. Nothing here
lets the loop act without the same gates it already has — triage only decides
*what gets looked at*; the council and the human still decide *what ships*.

## Open questions for the owner

- **Triage cadence & cap:** how often should it sweep, and what's the max
  escalations per sweep while we build confidence? (Suggest: hourly, cap 3.)
- **Who owns the verification checkers** — this workstream, or the builder/immune
  thread that owns `site_work_items`? (They're the natural owner of new checks;
  triage is the natural owner of routing.)
- **Auto-enable or manual?** Like the diagnose-dispatch-loop and the digest,
  ship triage disabled and turn it on once its routing is trusted — consistent
  with "more awareness before more autonomy".
