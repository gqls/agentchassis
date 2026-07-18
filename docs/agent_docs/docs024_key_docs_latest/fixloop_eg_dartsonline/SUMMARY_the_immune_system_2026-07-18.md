# The self-healing immune system — the whole picture, 2026-07-18

*A standalone account of what the diagnosis→fix system IS, where it came from,
where it is now, and where it is going — written to be read on its own by
someone who has not followed the turn-by-turn. Incorporates this week's real-case
results, the code-lookup tier, the six→nine-seat council, and the two operating
policies now in `CLAUDE.md` (council-on-fix, diagnosis-before-debug). Companion
detail: `NOTES_running_fixloop(10).md` turns 34–39, and the dated
`SUMMARY_where_we_are_*` docs for each phase.*

---

## What it is — three tiers

A self-healing platform, built in three tiers that map to "do the work / catch
the problem / fix the cause":

- **Tier 1 — the build workflows** do the actual work (build sites, render
  pages, generate content).
- **Tier 2 — the immune system** (checkers + handlers): detects problems and
  applies KNOWN remedies. Two feeds into tier 3: **triage** (a deterministic
  router over every recorded failure) and **silent-check** (a verification
  checker for defects no work item ever records). Between them they watch the
  whole fleet and decide what is a genuine code bug versus operational noise
  versus a capability gap.
- **Tier 3 — the diagnosis→fix loop**: for problems whose cause is in the CODE.
  It diagnoses with citations (forbidden to guess — cite or abstain), turns a
  CONFIRMED diagnosis into a constrained edit plan, has a reviewer **council**
  argue that plan, implements the approved plan in a caged pod (writes only via
  the git-adapter, holds no repo credential), gates it on build, and opens a PR
  for a human to merge. **Nothing merges itself.**

The load-bearing principle, visible in every result this week: **the system is
trustworthy because it REFUSES.** It refuses to confirm without evidence; it
refuses to confirm on code alone without observing the mechanism occur; it
refuses to bless a fix that only patches one instance of a class. Every apparent
"failure" this week was one of those refusals firing correctly on a flawed
input — and each refusal, once understood, became the next capability built.

## Where we came from

It began as a question: could a system take a plain-English "something is wrong",
work out the real cause unattended, and fix it safely? The first three pilot
bugs **dissolved under a cheap pre-check** — schema access plus grep answered
them before any loop ran. That taught the founding lesson, which still governs
the policy choices below: on this platform, bug *mechanisms* are mostly legible,
so the loop's value is NOT discovery — it is doing the work **unattended, with
citations a human can audit, consistently across a class of bugs.** Each
dissolved pilot was promoted to a graded **benchmark** with a rubric registered
before the run.

From there the assembly line was hardened slice by slice, each proven on the
benchmark before the next: diagnosis (F0, hardened across five scored runs);
the fix-proposer + two-reviewer council (F1/F2); the caged write step (F1.1b(c),
which opened and got a human to merge **PR #1** on 2026-07-13). Then the immune
system that feeds it — triage, silent-check, the awareness digest, and
feedback close-out — went live in full (all four phases, by 2026-07-16). At that
point the *tool* was complete; what remained was to point it at real work.

## Where we are now (this week's arc: it met real code)

### It ran its first real cases — and found a bug by failing
- **BUG A (first real-case CONFIRMED):** `GenerateText` never decodes
  `stop_reason`, so max_tokens truncation returns as silent success — the
  mechanism behind the article-body blanking, now cited platform-wide. Diagnosed,
  graded PASS against a pre-registered rubric.
- **BUG B (found by the tool's own honest failure):** the root `ai_service`
  block SHADOWS the step-level one, so per-step `max_tokens` is dead config
  whenever a root block exists — the fixloop runbook's own gotcha was INVERTED.
  Fleet-wide (17 agents at the 2048 default).
- Both handed off to fixing threads in **`/bugs_open/`** (the real-case queue,
  moved to repo root this week).

### Two capabilities grew from two honest refusals
- **Agent-state autogather (LIVE):** the diagnosis bundle now auto-includes the
  root+step `ai_service` blocks and recent `llm_call_log` rows for any agent
  type NAMED in the symptom — because BUG B's correct static diagnosis kept being
  coerced to UNVERIFIABLE with no reachable state evidence.
- **The code-lookup verify tier — F2.3b(c) (LIVE + PROVEN):** reviewers can ask
  code-shaped questions (`code_checks`: symbol / content / ls) answered from the
  `code_symbols` index — a DB read, so it runs in-chassis with no repo token.
  Because the council's first multi-seat run escalated over a question it could
  not answer.

### The clean proof the code tier works (the historian result)
The bug-historian seat's whole remit is checking a fix against the platform's
recurring failure patterns. On BUG A it raised the right objection — *the fix
patches one call site of a generic mechanism; do other provider adapters exist?*

- **Before the tier (2026-07-16):** the council could not answer it; three
  rounds exhausted; escalate.
- **After the tier (2026-07-18):** the historian asked the question as a
  `code_check` → the tier answered (both `anthropic.go` and `ollama.go`
  `GenerateText`) → the repropose widened the plan to cover the second adapter →
  **the historian APPROVED**, in its own words *"it covers both provider
  implementations rather than leaving the second one open."* Decision: approved,
  round 3, all seats. It approved with a mature advisory residual (no CI guard
  yet stops a *future* third provider being added unguarded) and a validation
  baseline (23 historical truncated rows should replay `success=false`
  post-fix). **BUG A now has a council-approved fix plan**, ready for the
  implementer.

Two fixes to the tier itself were found by that same run and shipped: a
Go-receiver-aware symbol match (a reviewer's `Type.Method` query must resolve the
stored `(*Type).Method` form) and dedup of identical checks (independent
reviewers ask the same question; the cap was dropping distinct ones to make room
for repeats).

### The council widened — fast
The reviewer roster went from 2 seats to **nine** this week (bug-historian,
reuse_agent, guidelines, tooling_provenance, diagnosis_guardian, llm_reliability,
debug_historian, plus the two always-on: edit-quality and guardian). A
**relevance filter** (`select_review_panel`, keyword footprints per seat) fires
only the seats whose territory a change touches — so cost is gated and the
edit-quality + guardian seats always run. On the historian's first real vote it
alone caught a scope gap two approving seats missed, advisory-not-veto as
designed.

### Two operating policies now in `CLAUDE.md`
- **Council-on-fix (advisory gate, live 2026-07-17):** any thread can put a
  platform change through the council before committing; APPROVED commits carry a
  `Council-Reviewed:` trailer; a coverage report shows who reviewed and who
  didn't. Cost is relevance-gated. Reviews the fix you wrote.
- **Diagnosis-before-debug (opt-in by judgement, NOT a gate — added 2026-07-18):**
  the deliberate asymmetry with council-on-fix. Council-on-fix is a cheap check
  on a concrete artifact gating a rare, dangerous act (deploy). Diagnosis-first
  is expensive, slow, competes with (and usually loses to) a thread's own
  diagnosis for visible bugs, and risks diagnosing a premise about to change — so
  it is **opt-in by criteria, not a default**: reach for it only when the cause
  is non-obvious after a quick look, or you suspect it is cross-cutting / not
  where the symptom is, or you want a cited/auditable diagnosis. Crucially, the
  cross-cutting class is **already auto-covered** by triage + silent-check, so a
  manual run is often unnecessary — check the queue first. It is the one thing
  the council gate cannot do: tell you the cause isn't where you're looking.

### The discipline that made all of it hold
Re-verify the PREMISE before every dispatch — not just the code, but the live
pod, the work-item queue, and every clause of the symptom (several runs this
week were spent on premises that shifted underneath). A repo-root `CLAUDE.md`
now carries commit-per-task and build-from-committed-HEAD rules; the 090 trigger
self-checks for work-item collisions. A NEW collision surface surfaced and was
filed for the coordination workstream: **DB config re-seeds clobber concurrent
config work** the way `git add -A` clobbers WIP — fix-proposer was re-seeded by
other threads ~4× during a single grading exercise; the mitigation is
patch-style idempotent seeds, never whole-object writes on shared defs.

## Where we're going

1. **BUG A closes its own loop:** the approved plan → implementer (092) → build
   gate → PR → human merge. The historian's CI-guard residual and the ollama
   sibling both feed the fix.
2. **Two forward threads, both handed off and moving:** feature-builder
   (multi-step capability construction — staged plans, a stage-looping
   implementer) and council-gate (the council as a fleet review service, now
   live-advisory).
3. **The council roster keeps widening** via the concept-register stage-3 track
   (nine seats and counting). The standing rule: any seat change patches BOTH
   councils (fix-proposer + council-gate) in one migration, or they drift — the
   exact class the council exists to catch.
4. **The tool's own backlog it can now eat:** `/bugs_open/` 001–011, including
   the three platform gaps in 003 (the reaper's `EXECUTING_STEP` blind spot found
   this week among them).
5. **Open owner decisions:** roster-wide model policy (diagnose-agent is on
   Sonnet 5; the proposer + reviewers still on Sonnet 4.6); whether to build the
   architecture-level guards the historian keeps flagging (a CI check that no new
   provider ships without the stop-reason guard).

## The one-line state
The loop diagnoses real code now, not benchmarks; it found a bug by failing;
it grew two verify capabilities from two honest refusals; its council is nine
seats and its newest reviewer's objection was resolved to an approval by the
exact tier built for it — and the system refuses in every direction it was built
to refuse in.
