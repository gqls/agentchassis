# The self-healing platform — where we came from, where we are, where we're going

*2026-07-16. The workstream-level state of the diagnosis→fix loop and its
triage/escalation layer. This is the technical journey doc; for the calm,
read-aloud version see `SUMMARY_where_we_are_2026-07-14.md`. Companion detail:
`DESIGN_triage_and_escalation.md` (architecture), `RUNBOOK_diagnosis_fix_loop(10).md`
(how + gotchas), `NOTES_running_fixloop(10).md` (turn-by-turn),
`HANDOFF_diagnosis_fixloop_2.md` (the cold-start resume point).*

---

## Where we came from

The workstream began as a pilot: could a system take a plain-English "something
is wrong" report, work out the real cause on its own, and fix it safely? Three
pilot bugs in a row **dissolved under a cheap pre-check** — schema access plus
grep answered them before any loop ran. That taught the first real lesson: on
this platform, bug *mechanisms* are mostly legible, so the loop's value is not
discovery. It is doing the work **unattended, with citations a human can audit,
consistently across a class of bugs**. So each dissolved pilot was promoted to a
graded **benchmark**, with a rubric pre-registered before the run and the
fixloop docs blinded out of the loop's corpus.

From there the assembly line was built and hardened, slice by slice, each proven
on the benchmark before the next:

- **Diagnosis (F0).** Reads real code + live DB, forms a theory, and is
  *forbidden to guess* — every claim cites evidence or it abstains. Hardened
  across five scored benchmark runs (anchor fixes, bundle enrichment, three
  verdict guards, cross-iteration data-request persistence, a symptom-closure
  gate).
- **Fix-proposer + council (F1/F2).** A constrained edit plan, then two reviewer
  agents argue it (right-edits vs platform-safety), a deterministic router
  combines their verdicts (approve / revise→verify→repropose /
  veto→reframe-once / exhausted→escalate), reviewers can demand a fact be
  checked and the system runs the query and feeds back the answer, and an honest
  "I need a human" escalation is treated as a *successful* terminal, not a
  failure. Live schema hints stopped reviewers hallucinating columns.
- **The write step (F1.1b(c)).** A caged implementer in its own throwaway pod
  reads via the GitHub API with a short-lived token, writes ONLY through the
  git-adapter (the AI-driven parts never hold the repo credential), touches only
  the plan's allowlisted files, is gated on `gofmt`+`go build` in a container,
  and opens a PR. **Nothing merges itself.** On 2026-07-13 the loop opened
  **PR #1** on a real one-file defect and the owner merged it — the entire line
  proven end to end, every gate on, human holding the final say.

## Where we are (2026-07-16): the tool is complete

Two things then had to be true before pointing it at real work: the owner must
always be able to *see* what it does, and the loop must be *fed* by real
problems automatically rather than by hand. Both are now built. The
**triage/escalation design** (`DESIGN_triage_and_escalation.md`) is **live in
full — all four phases**:

1. **Triage (Phase 1, v1.0.1117).** A deterministic router (no LLM) reads every
   recorded failure across the fleet and sorts it: genuine code bugs → escalate
   to the diagnosis queue (deduped by pattern, hard-capped per sweep);
   operational blips (timeouts, dead pods) → re-queue, never the loop; no error
   text → hold for a human; missing capability → roadmap, never the loop. Its
   first live run confirmed the value: ~half of all "failures" were operational
   noise it correctly kept out.
2. **Silent-check (Phase 2, v1.0.1118).** A verification checker for the class no
   work item ever records — the darts signature: a page in a site's navigation
   that was never built, with nothing anywhere flagging it. It emits inert
   findings **only for what the immune system cannot already see** (if any work
   item references the page, it stays out), grouping every affected site into
   one platform-level pattern so the cause is fixed once. It found the darts bug
   on two sites and routed it through triage into the queue.
3. **Digest escalation section (Phase 4, v1.0.1120/1121).** The awareness digest
   (deterministic, no LLM, delivered as a committed file) now carries the whole
   immune system on one page: sweep counts, the *entire* open diagnosis queue
   every digest (parked items are decisions waiting on the owner, so they never
   fade out; new ones are flagged), silent findings open **and** closed, and
   standing capability gaps.
4. **Feedback close-out (Phase 3, v1.0.1122).** Each triage sweep re-checks
   whether a parked escalation's failure pattern still exists among failed items
   (all-time, so aging out of a window is never mistaken for resolution) and
   closes the ones that have genuinely resolved — re-escalating automatically if
   they return. Proven both ways in production: a real sweep closed nothing (all
   patterns still real), and a synthetic probe closed itself while the real ones
   stayed open.

Everything is **deterministic where it routes, cited where it reasons, gated
where it acts, and manual where it dispatches.** Nothing runs on a schedule;
nothing merges itself; correctness doesn't depend on any one model. Three real
escalations sit parked in the queue right now, inert, waiting on a dispatch
decision.

### Deploy discipline earned along the way
Same-tag deploys ship stale binaries — so every feature is verified by grepping
the **running pod's** binary for its symbol, never the tag or git. Concurrent
sessions share the branch and the image tag, so builds are coordinated (bump the
tag, never rebuild an existing one) and rollouts are held behind a cluster-quiet
check when live page-builds are mid-flight (a rollout mid-orchestration silently
drops the spawn). These are written into the runbook's gotchas; they cost hours
before they were understood.

## Where we're going

Two tracks, both human-gated, both now unblocked because the tool is finished.

### 1. Real cases — the loop stops running on benchmarks
The owner opened the real-case queue (`aaa_fails_to_mend/`) on 2026-07-16 and
chose the **first case: the image-landing data-loss trap**
(`004_HANDOFF_image_landing_blanks_article_body.md`). Landing an image on a page
fires a scoped section re-render that, on any page whose article-body was never
unwrapped from its LLM JSON envelope, silently overwrites the good HTML with a
blank shell — it blanked 9 pages across 5 sites, with 4 more JSON-leaking. It's a
genuine, high-severity, already hand-diagnosed platform bug with a clean code map
— the exact shape the loop was built for, and gradable like the benchmark. It
also spans two threads (imagery found it; empty_sections owns the guard), so it
exercises the loop's cross-thread awareness.

**The case shifted on 2026-07-16, mid-handoff:** a new chassis (`v1.0.1123`)
shipped the guard, verified in the running pod — so new blanking is prevented and
the operating rule against landing images is lifted. What remains is the half a
guard cannot do: **recover the 13 still-broken live pages**, **fix
`ParseLLMJSON`'s 14 failing fixtures** (some envelopes are truncated and
unrecoverable), and the most loop-worthy piece — **the structural defect
underneath**: a schema-`required` field silently rendering empty
(`missingkey=zero`, `call_agent.go:1152`), which is the same class as the
product-page defect. Frame the intake around what's left, not the filed headline. The other queued cases: replan-
clobbers-built-pages (001), the errors-to-fix list (002), spawn-lost-child-
response (003). Dispatch order and timing are the owner's call — each run spends
credits.

### 2. Wider council — informed by the concept register (search-tab2)
The reviewer council today is two seats. Widening it (guidelines / reuse /
bug-historian reviewers, and eventually a capability-builder) is the ambitious
direction — and it is no longer a guess about *which* seats matter first,
because the **concept register** (`docs026_concept_register/`) now exists to
answer exactly that. Its stage 2 is complete: 1,627 concepts across 107
categories, each verified against live code/DB (~7.6% documentation-error rate
found and corrected). Its **stage 3 — "build council agents per concept area" —
IS this council-widening track.** `FIX-036` in the register is explicitly the
wider-council-roster vision, flagged by a consolidator as "the seam this concept
register is meant to help fill," and concepts independently rediscovered 4–6×
across documentation eras (e.g. "adoption writes first, classifier consumes";
the wrapper-orchestrator pattern) are the strongest signals for which reviewer
seat to build first. Wiring stage-3 seats into the live fix-loop workflow is a
cross-workstream production change the owner has reserved for explicit sign-off.

### Also open
- **Phase 3's re-drive half:** close-out observes resolution; *re-queuing* the
  original items after a fix ships stays a deliberate human action.
- **A silent-path hardening:** a schema-`required` field should never render
  empty (the `missingkey=zero` pattern behind the image-landing trap and the
  product-page defect) — structural fix, not per-site patch.
- **Auto-cadence:** everything is manual by design; a scheduled cadence is a
  later, deliberate flip, only after the owner is comfortable with the volume.

## The one-line state
Detect → sort → diagnose → plan → review → implement → gate → **your merge** is
built, proven, and self-reporting. The next real bug it looks at is the
image-landing trap; the next capability it grows is a wider council chosen by the
concept register. Both wait on the owner's go.
