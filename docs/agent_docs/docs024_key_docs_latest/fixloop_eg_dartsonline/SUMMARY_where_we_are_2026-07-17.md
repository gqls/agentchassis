# The diagnosis→fix tool — where we are, 2026-07-17

*The state of the tool itself (not the individual bugs it has found). Successor
to `SUMMARY_where_we_are_2026-07-16.md`, which declared the triage/escalation
layer complete. This one covers the week the tool stopped running on benchmarks
and started on real code — and grew two new capabilities from what that
exposed. Companion detail: `NOTES_running_fixloop(10).md` turns 34–38.*

---

## What we're doing

A three-tier self-healing system whose top tier is a diagnosis→fix loop for
bugs whose cause is in the CODE. It takes a plain-language symptom, diagnoses it
with citations (forbidden to guess — every claim cites evidence or it abstains),
turns a CONFIRMED diagnosis into a constrained edit plan, has a reviewer council
argue that plan, implements the approved plan in a caged pod (writes only via
the git-adapter, holds no repo credential), gates it on build, and opens a PR
for a human to merge. Nothing merges itself.

The design principle that keeps recurring: **the tool is trustworthy because it
refuses.** It refuses to confirm without evidence; it refuses to confirm on code
alone without observing the mechanism occur; it refuses to bless a fix that only
patches one instance of a class. Every "failure" this week was one of those
refusals firing correctly on a defective input — and each refusal, once
understood, became the next thing we built.

## Where we are now

### The tool ran its first real cases
- **First real-case CONFIRMED (BUG A):** `GenerateText` never decodes
  `stop_reason`, so max_tokens truncation returns as silent success — the
  mechanism behind the article-body blanking, now cited platform-wide.
  Diagnosed, graded PASS against a pre-registered rubric, handed to a fixing
  thread (`/bugs_open/008`).
- **A second bug found by the tool's own failure (BUG B):** the root
  `ai_service` block SHADOWS the step-level one, so per-step `max_tokens` is dead
  config whenever a root block exists — the fixloop runbook's own gotcha was
  INVERTED. Fleet-wide (17 agents at the 2048 default). Handed off
  (`/bugs_open/009`).
- The real-case queue moved to **`/bugs_open/`** (repo root).

### Two capabilities grew from what the real cases exposed
- **Agent-state autogather (LIVE, v1.0.1130+):** the diagnosis bundle now
  auto-includes the root+step `ai_service` blocks and recent `llm_call_log` rows
  for any agent type NAMED in the symptom — so a config-shaped bug arrives with
  the state-tier evidence the two-evidence-family guard requires. Born because
  BUG B's correct static diagnosis kept being coerced to UNVERIFIABLE with no
  reachable state evidence.
- **The code-lookup verify tier — F2.3b(c) (BUILT + DEPLOYED + PROVEN,
  v1.0.1132+):** reviewers can now ask code-shaped questions (`code_checks`:
  symbol / content / ls) answered from the `code_symbols` index — a DB read,
  so it runs in-chassis with no repo token. Born because the council's first
  three-seat run escalated over a question it couldn't answer ("do other
  provider adapters exist?"). PROVEN end to end on the re-grade: the historian
  asked exactly that, the tier returned BOTH `GenerateText` implementations
  (anthropic.go + ollama.go) with locations, and the repropose widened the plan
  to cover the sibling.

### The council widened
The reviewer roster went from 2 seats to **6** this week (bug-historian, then
reuse_agent, guidelines, tooling_provenance — the concept-register stage-3
track). On its first 3-seat run the bug-historian alone caught a real scope gap
two approving seats missed, advisory-not-veto as designed. The council-gate
design (open the council as a service every thread's fixes run through) is
handed off as its own thread.

### The operating discipline hardened
Re-verify the PREMISE before every dispatch — not just the code, but the live
pod, the work-item queue, and every clause of the symptom (three runs this week
were spent on premises that shifted underneath). The repo now carries a
`CLAUDE.md` with commit-per-task and build-from-committed-HEAD rules; the 090
trigger self-checks for work-item collisions. A NEW collision surface surfaced
today — DB config re-seeds clobber concurrent config work the way `git add -A`
clobbers WIP — filed for the coordination workstream with a patch-seed
mitigation.

## Where we're going

1. **The fair-round re-grade (doing now):** the re-grade proved the tier's
   mechanics but the loop was cut short by round-count inflation (accumulated
   `council_report` artifacts on the benchmark correlation). Clearing those and
   re-firing lets the widened plan actually reach a second review round — the
   test of whether the tier turns yesterday's escalation into an approval.
2. **F2.3b residuals now mostly closed** — SQL checks (b(a)), schema hint, and
   now code lookup (b(c)) all exist. The remaining reviewer-capability gap is
   thin.
3. **Roster model policy (owner decision):** diagnose-agent runs `claude-sonnet-5`;
   the proposer and all 6 reviewers still run `claude-sonnet-4-6`.
4. **The two forward threads, both handed off:** feature-builder (multi-step
   capability construction — staged plans, a stage-looping implementer) and
   council-gate (the council as a fleet review service).
5. **The tool's own bug backlog it can now eat:** `/bugs_open/` 001–009, plus
   the three platform gaps in 003 (including the reaper's `EXECUTING_STEP` blind
   spot this thread found and filed).

## The one-line state
The loop diagnoses real code now, not benchmarks; it found a bug by failing; it
grew two verify capabilities from two honest refusals; its council is six seats
and caught a real gap on the newest one's first vote — and the tool refuses in
every direction it was built to refuse in.
