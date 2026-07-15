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
3. Handlers **can fail — in THREE very different ways** (the third added
   2026-07-14 at the owner's prompting; it is already modelled in the code):
   - **Loud failure:** the handler errors, retries, and eventually lands at
     `status='failed'` (`attempt_count >= max_attempts`). The remedy, if the
     cause is code: **fix the handler** → fix loop.
   - **Silent failure:** the handler reports success but the problem persists.
     This is the darts benchmark bug exactly — `page-build-handler` "completed"
     a page it never built. Silent failure is the dangerous one, because nothing
     in the work item looks wrong. Remedy, if code: **fix the handler** → fix
     loop.
   - **No handler yet:** the work item's type has no builder at all. This is
     NOT a bug in a handler — it is a **missing capability**; the remedy is to
     **build a new handler**, which is feature work, not repair. The platform
     already models this as a first-class signal (see below) — it must route
     to the roadmap, never to the fix loop.

**The "no handler yet" case is already first-class in the code.**
`WriteBuildItemsAction` (`load_work_item_actions.go:217-280`) keeps two maps:
`availableBuilders` (page types with a working handler) and
`unavailableBuilders` (**"Known page types whose builders don't exist yet"** —
today `tool`→`tool-builder`, `entity-directory`→`directory-builder`,
`entity-page`→`entity-page-builder`). When a page needs an unavailable builder,
it does NOT fail and does NOT silently skip — it inserts a work item of
`item_type='capability_gap'`, `status='deferred'`, `handler_agent=<the needed
builder>`, with a spec naming `builder_needed`, then `continue`s (no dispatch
item is created). So the platform already *says*, durably, "I found work I have
no handler for." The docs' standing intent (RUNBOOK/NOTES) is that these become
**builder-queue / roadmap items** — a human capability decision. Nothing
currently routes or surfaces them; triage will.

So the gap is not detection and not fixing — both exist. The gap is: **when the
fixing fails (loudly, silently, or for want of a handler at all), what turns
that into the right next action?** Today: nothing. A failed item sits at
`failed`; a silently-failed one looks done; a `capability_gap` sits `deferred`,
unseen. That triple hole is what this design fills — routing each to its correct
tier, and only the code-bug ones to the loop.

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
                                                  ├─ FAILED (attempts exhausted) ─────┐
                                                  ├─ silent-success (re-detected) ─────┤
                                                  └─ NO HANDLER (capability_gap,        │
                                                     status=deferred) ─────────────────┤
                                                                                       ▼
                                                                          ┌── diagnosis-triage ──┐
                                                                          │  filter · dedupe ·   │
                                                                          │  loop-worthiness     │
                                                                          └──────────┬───────────┘
                          capability_gap ─▶ roadmap / builder queue + digest ◀────────┤
                                    transient / retry-worthy ◀──── re-queue ◀──────────┤
                                              human ◀──────────── ambiguous/risky ─────┤
                                     (handler exists, code cause) needs_diagnosis ◀─────┘
                                                                                       │
                                                                                       ▼
                                              fix loop: diagnose ▶ plan ▶ council ▶ {PR | escalate-to-human}
                                                                                       │
                                                          fix merged + deployed ───────┘
                                                                                       ▼
                                              re-queue the original items / re-run the checker  (VERIFY)
                                                                                       │
                                        still failing ─▶ back to triage          fixed ─▶ close the escalation
```

## The new piece: `diagnosis-triage`

A thin agent (per the constitution: thin workflow, logic in one Go action) that
runs **on a schedule** and does NOT diagnose — it **routes**. Each sweep:

1. **Gather candidates** from `site_work_items`:
   - `status='failed'` with `attempt_count >= max_attempts` (loud failure), and
   - `item_type='capability_gap'` / `status='deferred'` (no-handler-yet), and
   - *(phase 2)* items re-emitted after a prior completion — the silent-failure
     signal (see "the hard case" below).
   - Exclude anything already escalated or already under an open diagnosis/PR/
     roadmap item (a marker + cool-down), so nothing is raised twice.
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
   - **→ fix loop:** handler EXISTS but fails (loud/silent) and the cause looks
     like code — write a `needs_diagnosis` work item (the loop's existing
     intake) whose symptom is the pattern ("handler X fails item_type Y on N
     sites; error: Z"), with pointers to the failed items as evidence.
   - **→ roadmap / builder queue:** `capability_gap` (no handler yet) — this is
     a **capability decision**, not a bug. Surface it (grouped by
     `builder_needed`, with the count of pages/sites waiting on it) to the human
     roadmap and to the digest. NEVER route it to the fix loop — the loop makes
     constrained repairs to existing code, not whole new handler agents. (A
     future `capability-builder` pipeline could scaffold new handlers from
     specs — the "features from mission docs" direction — but that is out of
     scope here and deliberately human-gated.)
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

**Reconciliation with the empty-sections/loop-integrity thread (2026-07-14):**
their completion-verification gate (v1.0.1116) already converts silent failure
into loud failure for `empty_section` — and for any item_type they register a
verifier for — so that slice needs no new checker; Phase-1 triage (now live)
catches it as ordinary `status='failed'`. And option (b)'s recurrence signal
already exists platform-wide as `insertWorkItem`'s two-strike rule — Phase 2
should consume that mechanism, not rebuild it. What Phase 2 still owes is the
class **no completion ever touches**: defects where no work item fails at all
(the darts guides-index class — a section-index page `active` with zero
components and no failed item anywhere). Their full reconciliation:
`empty_sections_loop_integrity/` PLAN + RUNNING_NOTES.

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
- **Exists for the third flavour too:** `capability_gap` items with
  `status='deferred'` and `builder_needed` are ALREADY emitted
  (`load_work_item_actions.go:245-280`). Triage only needs to route/surface
  them — no new detection.
- **New (this design):**
  1. `diagnosis-triage` agent — scan → filter → dedupe → route (thin), across
     all three flavours (failed / silent / capability_gap).
  2. A verification/recurrence checker for the silent-failure class (owned by
     THIS thread; highest value; can start with one high-signal check, e.g.
     section-index pages that are `active` but have zero components — the darts
     signature).
  3. The dedupe + cool-down + escalation-marker bookkeeping on `site_work_items`.
  4. A new digest section — "escalations in / capability gaps / remedies /
     verifications" — so the owner sees the whole immune system, not just the
     loop, on one page.

## Suggested phasing

1. **Loud failures + capability gaps first.** Triage on `status='failed'` items
   and `capability_gap`/`deferred` items only — both already exist in the data,
   so this is pure routing with dedupe and a hard rate cap. Smallest safe slice;
   proves the escalation channel (to the loop) AND the roadmap surface with real
   data and low risk. Manual trigger, hourly when enabled.
2. **Verification checker for the darts class.** One high-signal silent-failure
   check, feeding triage. This is where the loop starts catching the bugs it was
   actually built for. Owned by this thread.
   **IMPLEMENTED + LIVE 2026-07-14 (v1.0.1118, `diagnose_silent_check`):**
   `nav_linked_never_built` emits; `deployed_zero_components` report-only until
   the owner promotes it; proven end to end same day (notes turn 30).
3. **Feedback close-out** — re-verify after a fix deploys; open/close
   escalations honestly.
   **IMPLEMENTED + LIVE 2026-07-15 (v1.0.1122, `triageCloseResolved`):** each
   sweep closes parked escalations whose failure pattern has vanished (all-time
   check; re-escalation automatic). Re-driving originals after a fix ships
   stays a human action. Notes turn 32.
4. **Digest section** for escalations + capability gaps — fold the whole immune
   system into the owner's one awareness surface.
   **IMPLEMENTED + LIVE 2026-07-15 (v1.0.1120, digest "Escalation channel").**

> **STATUS 2026-07-15: all four phases LIVE.** Remaining fixloop work is the
> separate later track — council-widening (F2 roster) and the owner-gated
> real-case queue (`aaa_fails_to_mend/`).

Each phase is independently useful and independently reversible. Nothing here
lets the loop act without the same gates it already has — triage only decides
*what gets looked at*; the council and the human still decide *what ships*.

## Decisions (owner, 2026-07-14)

- **Cadence:** sweep **hourly for now**, slower later once it's trusted and the
  escalation volume is understood. Keep a hard cap per sweep (suggest 3) while
  confidence builds.
- **Verification checkers:** **this workstream (the fix-loop thread) owns them.**
  It writes the silent-failure / "did the remedy actually work?" checks; the
  builder/immune thread stays the owner of the underlying `site_work_items`
  machinery. Triage owns routing.
- **Enablement:** **manual for now.** Triage and its checkers ship disabled and
  are fired by hand, consistent with "more awareness before more autonomy". Auto
  cadence is a later, deliberate flip.
- **THIRD FLAVOUR added:** "no handler yet" (`capability_gap`) is a first-class
  route — to the roadmap/builder queue, never the fix loop.

**Operating-context note (2026-07-14):** this build is currently driven by the
Fable model, whose credits are running low. Keep everything **manual** (no auto
cadence that would consume model calls unattended), and keep these docs
self-sufficient so the workstream survives a model change — the whole design is
gates + deterministic routing + human decisions, so it does not depend on any
one model to stay correct.

## Where this sits in the session's arc (for continuity)

- The fix loop is COMPLETE and proven end to end: it opened, and the owner
  merged, PR #1 (a real one-file defect). See `SUMMARY_where_we_are_2026-07-13.md`.
- The awareness digest (`fixloop_digest`) is live and delivered as a committed
  file under `docs/fixloop_digests/` (owner's chosen surface). Deterministic,
  no LLM.
- This design (triage + escalation, three failure flavours) is the NEXT build,
  Phase 1 = loud failures only. Not yet implemented.
- A merge wrinkle to remember: `main` is missing PR #1's fix (PR #2 merged
  `084→main` ~5 min before PR #1 landed the fix on `084`); the fix is safe on
  `084`; bringing it to `main` is a clean one-commit merge, owner to okay the
  push.
