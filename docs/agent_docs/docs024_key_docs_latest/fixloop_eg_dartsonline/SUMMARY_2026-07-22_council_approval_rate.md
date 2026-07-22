# SUMMARY — why the council almost never approves, and the fix we chose

**2026-07-22. Milestone: the low-approval-rate diagnosis is confirmed and a fix
direction is chosen.** Written to be read aloud. New file, per the standing-five
rule — it does not overwrite the earlier council-gate summaries; the series is the
record.

## What we're trying to do

Get **every platform bugfix reviewed by the council** before it ships — and
eventually let the council actually *block* a bad one (PR-mode). Success was meant
to be measured by the coverage report: commits that carry a `Council-Reviewed`
trailer, which is earned by an **approved** verdict.

## Where we've come from

The gate has been live and advisory since 2026-07-17. Adoption became real and
unprompted — in a recent two-day window, twenty-five separate changes were put
through a council and fifty-six verdict notes were written. The hard part —
getting people to use it — was working.

## What we've done (this milestone)

1. **Measured the mission.** Grounded in the council reports: over seven days,
   ~44 distinct submissions, **123 revise / 3 rejected / 2 approved** — an approval
   rate of about **4.5%**, and the two approvals weren't even bugfixes. Threads
   iterate hard (five, six, seven rounds) and still almost never get a yes.
2. **Got the owner's steer:** diagnose *why* approval is unreachable before
   changing anything.
3. **Filed the diagnosis loop** — and it **stalled**. It got stuck at its `route`
   step (the known spawn-loss / dispatch-queue backlog, bugs_open/003 and /030)
   and was reaped after four hours having produced no verdict. That is a second,
   separate finding: the canonical diagnosis path is currently unreliable for
   code-seeded symptoms.
4. **Confirmed the diagnosis by hand** from the primary evidence the loop would
   have read — the actual objections that decided real revise rounds.

## Where we are now — the confirmed diagnosis

Approval is unreachable **because of the decision rule, not because the plans are
bad.** The code that tallies the votes (`decideCouncil` in
`diagnose_council_decide_action.go`) sends a change back to "revise" the moment
**any one** reviewer objects, and it **never looks at how severe the objection
is** — even though every objection is graded low, medium or high. In the live
data:

- Of 485 objections in revise rounds, **high severity was only 8.5%** (medium 279,
  low 165, high 41).
- **67% of revise rounds (59 of 88) contained no high-severity objection at all** —
  they were blocked entirely by low/medium nits.
- Fourteen rounds were a **single** seat blocking eight or nine approving seats,
  and thirteen of those fourteen did it on a non-high objection.

So a tiny nit stops a change exactly as hard as a serious flaw. The objections
themselves are still worth having — the fix keeps recording and returning them;
it just stops a minor one from *blocking*.

**Two honest caveats.** The loop never independently graded this (it stalled), so
it is a hand-diagnosis from primary evidence, not a council-graded verdict. And
"severity" is the reviewer's own label — if seats under-grade, some low/medium are
really blocking — so the fix must stay conservative.

## The choices we weighed, and what we chose

The sizing that made the decision sharp (88 revise rounds; a round's class is its
worst objection):

| Worst objection in the round | Rounds | "only high blocks" | "medium still blocks" |
|---|---|---|---|
| high | 29 (33%) | stays blocked | stays blocked |
| **medium** | **56 (64%)** | **flips to approve** | stays blocked |
| low | 4 (4.5%) | flips to approve | flips to approve |

Because *medium* is the worst objection in 64% of rounds, how medium is treated is
the whole decision.

**Option A — Severity gate, only high/veto blocks. ← CHOSEN.**
A high-severity objection or a veto still blocks; low and medium become
"approve-with-notes" — recorded and handed back to the proposer, but they do not
force a revise. Flips ~68% of rounds toward approval. Simplest rule; trusts the
reviewers' severity labels the most. **This is what the owner chose.**

**Option B — Consensus-medium gate.**
High/veto blocks; a *medium* blocks only if two or more seats independently raise
medium-or-higher (a lone medium becomes a note); low becomes a note. More
conservative against a single seat over-worrying, still catches a shared concern.
Not chosen — more machinery, and the "only high blocks" rule already keeps every
high-severity catch.

**Option C — Don't change the rule; fix the process and the metric instead.**
Accept that the council's value is its *objections*, not its approvals. Tell
threads to submit once or twice, take the objections, and ship; redefine the
coverage report to count *consulted* rather than *trailered-and-approved*. No
code change, and PR-mode stays off the table. Not chosen — it concedes that
approval (and therefore enforcement) is permanently out of reach.

**Option D — Independently grade the diagnosis first.**
Before touching a fleet-wide rule, get the council's own verdict on this
diagnosis — fix the route stall so the loop can run, or re-fire and accept the
four-hour stall risk. Not chosen — the hand-diagnosis rests on reproducible
primary evidence, and the fix is conservative enough (it never weakens a
high-severity or veto block) that the risk of acting now is low.

## Where we're going

- **Build the severity gate** in `decideCouncil`: high or veto blocks; explicit
  low/medium are advisory. Conservative carve-outs so a minor label can't hide a
  real problem: a *degraded* (truncated) review still blocks, and an objection
  with no recognised severity still blocks — only an *explicitly* low/medium
  objection is waved through.
- **Test it** — extend `diagnose_council_test.go` for the new behaviour.
- **Dogfood it** — put the change through the council gate once (it will probably
  come back "revise" under the *old* rule, which is the problem proving itself).
- **Ship on the next image roll** — it is one Go change and it corrects **every**
  council at once (fix-proposer, gate, experience, concept-register), because they
  all share this action. Inert until the roll.
- **Then re-measure.** If the approval rate moves as the counterfactual predicts,
  trailer-coverage becomes meaningful again and PR-mode becomes a real
  conversation. If it doesn't, that itself is the next question.
- **Separately:** the diagnosis loop's `route` stall is worth its own line back to
  the bugs_open/003 / 030 owners — the immune system's own diagnosis path is
  currently blocked by the same dispatch backlog it exists to investigate.
