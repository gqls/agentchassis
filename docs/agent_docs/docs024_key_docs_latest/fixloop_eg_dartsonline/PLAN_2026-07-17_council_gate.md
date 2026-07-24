# PLAN — Council Gate (the F2 council as a service)

*Started 2026-07-17. Backfilled into the standing-four format 2026-07-18 per the
owner's working-docs directive; the design it executes is
`DESIGN_feature_builder_and_council_gate.md` §2, and the decisions below were
taken as the work happened (dates on each). Companions: `RUNBOOK_council_gate.md`
(commands), `NOTES_running_council_gate.md` (running record),
`SUMMARY_council_gate_2026-07-18.md` (read-aloud).*

## The problem

The fix loop's council judges a `fix_plan` artifact by correlation_id and does
not care who authored it. Meanwhile many concurrent sessions change platform
code daily with no second pair of eyes. So: open the council as a service any
thread can submit to, without inventing parallel machinery, without becoming a
bottleneck, and without pretending it is compulsory when it is not.

## Phasing (design §2 build order — visibility BEFORE enforcement)

| # | Phase | State |
|---|---|---|
| 1 | Submission wrapper (diff + rationale → `fix_plan` artifact) — `097` | DONE 2026-07-17 |
| 2 | Trigger + orchestrator seed running only the council portion — `0NN_council_gate.sql` | DONE, APPLIED 2026-07-17 |
| 3 | Visibility: un-reviewed platform commits — `098` | DONE 2026-07-17; corrected twice on 07-18 |
| 4 | **PR-mode (real enforcement)** — platform changes ride `fix/*`, verdict attaches to the PR, owner merges only green | **NOT BUILT — owner's explicit go required** |

## Decisions, and why

| Date | Decision | Reason |
|---|---|---|
| 07-17 | Scope = `platform/`, `internal/`, `pkg/` | Docs/site content would spend credits for no safety gain; the 097 script refuses them client-side. |
| 07-17 | Advisory at launch, not PR-mode | Standing rule: visibility before enforcement. PR-mode changes how *every* thread works — owner's call, on evidence. |
| 07-17 | One council run per coherent task/commit | Matches commit-per-task. Per-iteration review is hostile at minutes-per-run. |
| 07-17 | Launch waits for more seats | Owner ruling; then satisfied same day (seats grew 3→5→7→9→13). |
| 07-17 | No repropose/reframe loop in the gate | The *author* revises — a code-capable session that reads the objections itself. This later proved to matter: the fix loop's blind reviser (bug 016) has no gate equivalent. |
| 07-17 | `code_lookup` deliberately not mirrored | It answers code questions for the blind reviser; the gate has no reviser. Divergence by design, recorded in the seed header so it is never "fixed" as drift. |
| 07-18 | Roster mirroring goes mechanical (`099_SYNC_gate_roster.py`) | Hand-mirroring two rosters that must stay identical *is* the drift class the council reviews for. Five roster changes in 18 hours settled it. |
| 07-18 | Trailer accepts a gate correlation **or** a fix-proposer run id | A fix the fix loop's own council approved is genuinely reviewed. Refusing it would push threads away from the convention. |
| 07-18 | "Evidence gone" is not "false claim" | Council reports are deletable (see corrections); accusing an honest commit is worse than reporting a gap. |

## Corrections (kept visible, not edited away)

- **`098` reported 4 of 41 commits and looked fine.** `kubectl exec -i` inside
  the read-loop consumed the loop's own stdin, so it stopped at the first
  trailered commit. Found and fixed by another thread. Rule that generalises:
  `-i` only when SQL arrives on stdin; never when it arrives via `-c`. My own
  later heredoc hit the mirror image of this — no `-i`, so psql got nothing,
  exited 0 and printed nothing. **Verify the write, not the exit code.**
- **`098` accused an honest commit.** It resolved trailers only against
  `correlation_id`, so a fix-proposer-approved commit (`f32b208e5`) showed as
  MISMATCH. Now resolves either key, by prefix.
- **The verdict it pointed at then vanished.** `091`'s documented "clear
  council_reports for a fair run" DELETE destroyed the approved report between
  two runs of the report (12:03 approved → 13:29 gone). That advice is retired:
  round counting has been orchestration-scoped in code for some time, so
  clearing buys nothing, and a `council_report` is now commit evidence.
- **Seat mirroring by name-set was too shallow.** Another thread deepened `099`
  to a step-by-step compare after the gate silently sat on an older model/token
  config while fix-proposer moved on. Same lesson as the rosters themselves.

## Open questions for the owner

1. **PR-mode?** Only on evidence from advisory mode. Adoption is the input.
   > **RESOLVED 2026-07-24 — DEFERRED; strengthen advisory instead.** The evidence
   > came in: after the severity-gate fix (`bugs_open/057`, live v1.0.1149) the
   > approval rate went from ~5% (3/53 submissions, 07-15..07-22) to **~80%** (8/10,
   > the two days after), and `098` REVIEWED went 0 → 4 in three days. So approval
   > is now reachable AND discriminating (the gate still blocked the ~20% with a
   > high-severity/veto problem) — PR-mode is now *buildable*. The owner chose NOT
   > to build it yet: **strengthen advisory, defer structural.** PR-mode collides
   > with the many-sessions/one-shared-branch model (platform fixes land directly
   > and ride the next sweep build; fix/* PRs would hold them until merge + add a
   > council round of latency) — a workflow change for every thread, not worth it
   > while advisory is delivering coverage. Delivered instead: a `commit-msg`
   > advisory nudge on an un-reviewed platform-code commit (`scripts/council-coverage-nudge.sh`),
   > the hardened norm in CLAUDE.md, and a persisted `098` coverage baseline (the
   > "loud + regular" pair). Revisit PR-mode if direct-commit coverage stalls.
2. **Repropose seat coverage (bug 016, second finding).** The fix loop's
   reviser sees 6 of 13 seats. Fixing it wants a shape decision — list all
   thirteen, or have the reviser read the `council_report` artifact once — and
   the second scales without touching the prompt on every new seat.
3. **Should approved verdicts be immutable?** Pinning or an append-only mirror
   would make a trailer permanently verifiable. Not built.
