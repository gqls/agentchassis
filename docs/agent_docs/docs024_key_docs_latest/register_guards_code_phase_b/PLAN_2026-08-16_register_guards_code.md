# PLAN 2026-08-16 — the register reaches the tools that encode its facts (Phase B)

**Lane created 2026-08-16** from `bugs_open/225` (SDLT calculator, expired £625k
first-time-buyer cap, sixteen months live). 225 itself was already fixed; this lane
exists for the CLASS, now tracked as `bugs_open/288`.

## The problem, in the owner's terms

The evidence register — the per-site list of facts each with a GOV.UK citation,
re-verified daily — constrains what a page may **say**. It never constrained what a
calculator **computes**. So mortgagecalculator's stamp-duty tool could run a tax rule
that expired on 2025-03-31 while the register three feet away carried the correct
figure, re-verified that very morning, and every check we own passed.

## The decision that shaped everything: this is not a new design

`mortgagecalculator_couk_adoption/PLAN_2026-08-09_facts_into_tool_acceptance.md`
already designed this, in four pieces, and the owner has seen it. Piece 1 (show the
tool BUILDER the register) went live as migration 366 / CLM-021 on 08-10. Pieces 2
and 3 were designed, unclaimed, and sitting last on that lane's list while they
worked on brand assets. **So the correct move was to implement their plan, not to
invent a second mechanism beside it.** Two rejected alternatives, and why:

- **A retired-figure scan in the shared claims engine.** Attractive — it would cover
  prose too, and every site at once. Rejected for now: it is a second scanner over
  the same components (`bugs_open/093`'s shape), it needs a measured false-positive
  rate before it can be trusted on `<script>` bytes, and the 08-09 plan's §3 already
  rejects the blanket form. Recorded as residual §4 in `bugs_open/288`.
- **Piece 4, the oracle.** The strongest piece and the only one that answers "is the
  figure RIGHT". It needs an RFC and a Go oracle library. Out of scope, said so.

## What Phase B is

**Piece 2 — the declaration.** A tool's criteria fence (the ```criteria block in
`doc_plans.body`) may carry a fence-level `"facts": ["<fact_id>", …]`.

- **Fence-level, not per-check** — a tool encodes a fact; a check does not. A
  per-check `facts` stays refused by the validator's existing P7 inert-field rule.
- **Ids only, never values.** `doc_plans` has no `site_id`, so a value pinned there
  would be a fleet-global copy of a per-site number — which is the plan's own F3
  ("a golden that pins a wrong answer and then defends it") one rung higher. Values
  resolve from the swept site's register, every time.
- **P11** refuses a malformed declaration where it is *written*, rather than silently
  ignoring it on the day a fact moves.

**Piece 3 — the fan-out.** The daily `evidence-freshness` sweep resolves each site's
declaring PLANs to its pages and files one work item per (fact, tool) that is owed
something.

## The routing split, which is the part that most needed getting right

| what happened | where it goes | why |
|---|---|---|
| the fact's VALUE moved, page has a tool-level component, fence allows auto-fix | `improve_tool` → tool-improver | a fixer can safely act here and only here |
| value moved, but non-fork page | `fact_drift_review` (human) | tool-improver rewrites the SHARED template of a non-fork — it did exactly that fleet-wide on 08-05 and again on 08-14 (`bugs_open/281`, TL-042) |
| value moved, but fence says `no_auto_fix` | `fact_drift_review` (human) | the fence author has said a human decides what may change on a page quoting tax law (TL-040) |
| the EVIDENCE drifted (citation lost, artefact changed) | `fact_drift_review`, always | a 404 at GOV.UK is not evidence a figure moved; aiming a rewriter at arithmetic on that basis is `bugs_open/126` |
| the source could not be checked at all | **nothing filed** | CLM-008, unknown is not loss; an item on every 403 day trains people to ignore the queue |

**The honest consequence, stated rather than buried:** both live SDLT fences carry
`no_auto_fix` and neither page is a fork, so on today's fleet *every* route is human
and the `improve_tool` arm is proven by unit test only. That is the intended trade.
Widening the auto-fix door to make an arm reachable would be backwards.

## The measurement that changed the design

The obvious way to find "the tools on this site" is the acceptance ladder's own
eligibility predicate, `toolEligibilityWhere`. **Run live on 2026-08-16, it returns
NEITHER tool this mechanism exists for** — mortgagecalculator's `tool-stamp-duty`
carries two components and loanandmortgagecalculator's `mortgages-stamp-duty` three
since its B2 decomposition, and the predicate's sole-component clause excludes both.

Keying on it would have produced a check that could never fire on the tools that
motivated it — and it would have reported green for ever. The fan-out uses the
platform's own name rule instead (the one Tier 4 already uses to resolve a tool's
URL). **Encoding a fact and being acceptance-eligible are different questions.**

## Constraints honoured (all pre-existing rulings)

- The RFC_025 council objection: the existing `stale_evidence` raise gate must not
  change for existing callers. The fan-out never touches `changed`, `res.Drifted`,
  `ArtifactCheckDrifted` or `shouldRaiseStaleEvidence`.
- Owner ruling 2026-08-02 §1: converging a new producer onto an existing item type
  needs no RFC *provided* the producer set and key shape are named in the concept
  register. Done, in the same commit (CLM-022).
- Platform-seams condition (2): register entry in the commit that ships the seam.
- RFC_022: opt-in, unsafe default OFF, zero live consumers enumerated → council
  scope, not architecture scope.

## Phasing

1. **Go + register + tests** — done, `989addb1c`, council-submitted `cff364b8`.
2. **Roll**, then prove at the binary (not the tag).
3. **First declaration** — mortgagecalculator's `stamp-duty` fence, via their
   `install_fences.py` (never hand-edit the row; the script rewrites whole bodies).
   Handed over as a CONTRIB — the site is that lane's.
4. **Induce** — supersede `sdlt-ftb-relief-cap` 500000→550000, dry run must name
   `stamp-duty`, restore. A dry run that reports nothing after a real change is the
   failure, and it is the only result that distinguishes this from an inert check.
5. **LMC's declaration** — inert until LMC has a register at all (another lane's
   decision, pending).
