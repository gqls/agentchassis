# Where the fix loop stands, and the one decision open — 2026-07-12

*A plain-language position summary. Companion to
MILESTONE_diagnosis_fix_loop_2026-07-10.md (the story so far) and
PLAN_fixloop_pilot.md (the technical plan).*

---

## The big picture in one paragraph

We have built, in production, a system that takes a plain-English bug report,
diagnoses the true cause with citations, writes a constrained fix plan, and
puts that plan in front of a two-reviewer council that approves it, sends it
back for revision (now with the ability to *answer its own factual questions*
against the live database), or escalates it to a human with a complete
evidence package. Every piece of that sentence has been exercised live and has
a paper trail. **What the system still cannot do is act on an approved plan** —
turn it into actual code changes a human can review and merge. That final
slice — "the write step" — is what we are building now.

## What the write step is

When the council says *approved*, a new agent (the **fix-implementer**) will:

1. Load the approved plan and re-verify it really was approved (a gate, same
   pattern as "only a CONFIRMED diagnosis may seed a plan").
2. Read the current code and write the *complete new versions* of only the
   files the plan names.
3. Pass a deterministic safety check — the plan's file list is a **hard
   allowlist**: a produced file outside the plan, a file the plan asked for
   but missing, an empty or unchanged file — any of these rejects the whole
   implementation before anything touches git. *(This is built and tested —
   see below.)*
4. Create a branch (`fix/<diagnosis-id>`), commit the files to it, and open a
   **pull request** whose description carries the full story: the diagnosis,
   the plan, the council's decision. **A human reviews and merges — or
   doesn't. Nothing merges itself.**

One security decision was already made (owner, 2026-07-12): the GitHub write
credential **stays in the git-adapter** — the platform's existing, single
write surface to GitHub. The fix-implementer never holds a token; it sends a
request to the adapter, same as the site-deploy pipeline has always done. This
was chosen over the original sketch (injecting a write token into the
implementer's pod) because it reuses a proven surface and keeps write
credentials out of LLM-driven pods entirely.

## What is already done (committed, tests green)

| Piece | State |
|---|---|
| git-adapter: `create_branch`, `create_pull_request`, commit-to-branch | Code committed + unit-tested. **Needs a git-adapter image rebuild to go live.** |
| Safety core (`diagnose_prepare_fix_commit`): the hard allowlist + payload assembly | Code committed + 7-case test suite green. Rides the next **chassis** image. |
| Council / revise / verify / reframe / escalate (F2.x) | Live on v1.0.1108 + workflow v5, proven on benchmark runs. |

## What remains

1. **The build gate** — the open decision, below.
2. The fix-implementer workflow seed (the SQL that wires the steps together).
3. Two image rebuilds (chassis + git-adapter) and a smoke test that the
   adapter's token can write to the platform repo.
4. An end-to-end run — which needs an *approved* plan. Note: our benchmark
   bug keeps honestly terminating at *escalate* (it genuinely needs an
   architecture decision), so the first end-to-end write-step test will need
   either a simpler seeded bug or a hand-approved plan.

## The open decision: the build gate

**What it is.** A rule decided early (Q-C, 2026-07-07): before the platform
opens a pull request, the changed code must at minimum *format cleanly and
compile* (`gofmt` + `go build`). An LLM writing whole Go files will sometimes
produce something that doesn't compile; the gate stops that reaching a human
as a "ready" PR.

**The complication.** The agent pods deliberately run a minimal image with no
Go toolchain. So the gate has to run somewhere else. Three ways to do that:

### Option A — GitHub Actions on the PR (fast, standard)
A one-file CI workflow in the repo: every PR on a `fix/**` branch
automatically runs gofmt + build; the result shows as a green tick or red X
on the PR itself.
- **Effort:** ~an hour. No new platform machinery at all.
- **Implication:** the PR *exists before* the gate runs. A non-compiling
  implementation becomes a **visibly red PR** rather than no PR. That softens
  the original ruling: the human might open their PR list and see broken
  proposals (clearly marked as broken). In exchange: the check re-runs on
  every later push, it lives exactly where the human already reviews, and
  it's the same mechanism every human contributor is judged by.
- **Risk if chosen alone:** if the implementer turns out to produce many
  non-compiling PRs, the PR list gets noisy — the signal we'd use to justify
  building Option B.

### Option B — a spawned build Job before the PR (the ruling as written)
The platform spawns a short-lived Kubernetes Job from a stock `golang` image:
it clones the fix branch (read-only token), runs gofmt + build, and reports.
Only a green result proceeds to open the PR.
- **Effort:** roughly a day of new machinery — a "run this Job and wait"
  action (Job spec, polling, log capture, timeouts, cleanup) that the
  platform currently does not have.
- **Implication:** a broken implementation **never becomes a PR** — the run
  ends as a failed/escalated artifact instead. Strictly honours the original
  ruling. The human only ever sees compile-clean proposals.
- **Bonus:** a generic "run a container and wait" action is reusable
  (future: tests, linters, security scanners as gates).

### Option C — A now, B next (belt and braces)
Ship the CI file today so the first end-to-end run is unblocked; build the
Job gate as the immediately-next slice regardless. CI stays as a second check
on every push even after B exists.
- **Effort:** A's hour now, B's day soon.
- **Implication:** fastest to a working loop *and* ends at the strictest
  posture. The only cost is doing both.

**Recommendation:** **C** if the write step should reach end-to-end quickly
without permanently softening the ruling; **A alone** if we want to see real
implementer behaviour before investing in Job machinery; **B alone** if no
red PR must ever appear, accepting ~a day's delay before the first
end-to-end run.

---

*Once decided: seed SQL → image rebuilds → adapter token smoke test → first
end-to-end run against a hand-approved or seeded-simple plan.*
