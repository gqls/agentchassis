# NOTES — register↔tool fact drift (Phase B)

Append-only, newest at the bottom. Missteps are the point.

## 2026-08-16 — session 1: from "take the next bug" to a class fix

**Started at `bugs_closed/225`** on the owner's instruction. First finding, before any
design: **225 was already fixed and live** (2026-08-09), and had been kept in
`bugs_open/` by the owner's 08-06 direction — which the owner **superseded on 08-12**
("if it is fixed and live it should be moved"). So the case itself was paperwork, not
work. Re-verified at the wire rather than trusting the file: `grep -c 625000` = 0 on
both live pages, `FTB_RELIEF_CAP = 500000` present as the positive control.

**Trap found while re-verifying, worth its own line:** the component id the bug file
names in its "Fix landed" section (`55682bc8-…`) **no longer exists**. The LMC lane's
B2 work decomposed that page into `prose-0` / `tool-1` / `prose-2` after the bug was
written. A durable file's own evidence pointer rotted in a week. This is not a footnote
— it is the reason the class fix addresses artefacts by page name and not by component
id, and it is recorded in `bugs_open/288` §"the mechanism that COULD have".

**The real work was the class**, which 225's own section "Why no existing check could
ever have caught this" had already identified and nobody had filed.

### Misstep 1 — I nearly built a second mechanism beside an existing plan

My first design sketch was a register-side `retired_values` + `artifact_check`
extension: teach the daily sweep to assert that an expired figure is ABSENT from a
site's artefacts. It is a reasonable mechanism. It was also **the wrong move**, and
what caught it was reading `mortgagecalculator_couk_adoption/` before writing code:
`PLAN_2026-08-09_facts_into_tool_acceptance.md` had already designed this class fix in
four pieces, owner-visible, with Piece 1 live since 08-10 and Pieces 2+3 **designed,
unclaimed, and sitting last on that lane's list**.

Building my own version would have produced two mechanisms for one job, one of them
undocumented in the plan the owner has read. **The cheap check: `grep -rl` the lane
directories for the mechanism before designing, not just `bugs_open/`.** The plan's
own §5 landmines then bound the implementation (doc_plans has no `site_id`; never
round-trip the register through the typed struct; SKIPPED ≠ PASSED).

### Misstep 2 — a stale figure carried from a subagent into a commit message

Full row in `WRONG_CALLS.md`. Short version: I wrote "0 of 130 facts, 0 of 87 tool
PLANs" into commit `989addb1c` from a research subagent's report without grounding it.
Live numbers are **143** and **90**. The load-bearing half (zero) was right; the
denominators were ~10% stale, and forward-only means the commit message is wrong in
the archive for ever. **A subagent's report is another doc.** It has no special
standing for having been produced inside my own session — if anything it is worse,
because there is no visible seam where measuring stopped and quoting began.

### The measurement that changed the design (and would have shipped an inert check)

I was about to key the fan-out on the acceptance ladder's own eligibility predicate,
`toolEligibilityWhere` — the obvious reuse, and it would have looked like good
citizenship. Ran it live against both SDLT sites first:

> it returns **neither** of the two tools this mechanism exists for.

mortgagecalculator's `tool-stamp-duty` has two components; LMC's
`mortgages-stamp-duty` has three since B2. The predicate's sole-component clause
excludes both, deliberately and correctly for *its* purpose. A fan-out keyed on it
would have been a check that could never fire on the tools that motivated it — and it
would have reported green for ever, which is the estate's most expensive shape.

**Encoding a fact and being acceptance-eligible are different questions.** The fan-out
uses the platform's own name rule instead (the one Tier 4 already uses for a tool's
URL). This is in CLM-022 and TL-045 as a landmine, because the next person will reach
for the same reuse.

### What was built

`989addb1c` — Piece 2 (`criteria_facts.go` + validator rule P11) and Piece 3
(`refresh_evidence_fact_drift.go` + two wiring lines in the refresher). Register
CLM-022 + TL-045 in the same commit (platform-seams condition 2). Council submitted
`cff364b8` before committing, so the commit carries `Council-Submitted:` — the trailer
that asserts nothing and is credited automatically at report time.

### Verification, and what each check could have come out as

- **Six routing guards, six induced reds.** Each guard deleted in a scratch HEAD tree,
  its test run, FAIL observed, code restored, green re-observed: `no_auto_fix`, fork,
  evidence-vs-value, fetch-error, baseline-present, baseline-precedence. Plus a
  seventh on P11. Without this, thirteen passing tests would have proven only that the
  code runs.
- **The no-op case asserted, not assumed:** a site with no declaration costs exactly
  one SELECT (sqlmock `ExpectationsWereMet`) and its result JSON is byte-identical to
  before the change (golden compare in the test).
- **Built against `git archive HEAD` + my files**, because another session's untracked
  `component_name_resolver_menu.go` does not compile at the working tree. A green
  `go build` there would have been meaningless either way.

### What is deliberately NOT done, and why it is not laziness

- **Piece 4 (the oracle)** — the only piece that answers "is the figure RIGHT". Needs
  an RFC and a Go oracle library. This mechanism only answers "did it MOVE": a tool
  and a register wrong in the same direction agree, and it is silent.
- **A retired-figure scan over prose/scripts** — would close the other half of the
  class (blind spots 2 and 3), but it is a second scanner over the same components
  (`bugs_open/093`'s shape) and needs a measured false-positive rate first.
- **RFC_025 stage 2b** (`page_name` addressing, reachable for every source kind) —
  small and adjacent, but it changes an existing mechanism's semantics and deserves
  its own round rather than riding in on this one.

All three are named as residual in `bugs_open/288`, not left as folklore.

## 2026-08-16 (later) — council round 1: REVISE, and it found the defect I could not see

**The verdict was REVISE, gated by `editquality`, and it was worth every minute.** The
memory line "a REVISE round is cheaper than the defect it finds" held again.

### The gating objection, which I had reasoned my way INTO

> classifyFactDrift's baseline is defined purely from PRIOR REGISTER state … So a tool
> that is wrong from the day it opts in, against a register fact whose value has been
> stable since (exactly bug 225's shape: correct register, stale code, no subsequent
> legislative change), produces baseline == current and … emits nothing.

I had a test asserting that silence — `TestClassifyFactDrift_NoBaselineIsSilent` — with
a comment explaining that "is the current number right" is Piece 4's question. **The
reasoning was sound and the behaviour was still wrong**, which is the interesting part.
The distinction I missed: I do not need an oracle to notice that *nobody has ever
checked this pair*. That is not a claim about arithmetic; it is a claim about our own
records, and we hold those.

So a first declaration now files one `unreconciled_declaration` review item per (fact,
tool). It is self-quieting — the item carries the value, which becomes the next pass's
baseline — and the mutation test proves the point: **M8 reverts to the old behaviour and
the new test fails**, so the objection is closed by construction, not by assertion.

**The transferable lesson, which is not "listen to reviewers".** My silence had a
*justification*, and the justification was true. A defensible reason for a behaviour is
not evidence the behaviour is right — it only proves the behaviour is not accidental.
The test I wrote asserted my reasoning back to me, which is the most comfortable kind of
green there is. What the seat did that I did not: it took the mechanism's own motivating
bug and ran it through the new code. **The check I should have written first: does this
fire on 225?** It did not, and nothing in my thirteen tests asked.

### Two more real findings, and three answered by measurement

- **`compliance` (medium)** — a risk inversion I had built in without noticing: every
  human-routed finding sat at medium/35, so a *confirmed* value move on a `no_auto_fix`
  tax calculator ranked BELOW an auto-fixable drift at high/30. The harder something is
  to fix automatically, the more urgent it is, not less. Now three bands.
- **`debug_historian` (high)** — `p.status='active'` is the estate's LIVENESS predicate
  and is documented as wrong for anything choosing which pages to JUDGE. The sharp part
  of the objection was not the predicate itself but the asymmetry: *I had measured the
  eligibility predicate and not this one, in the same query*. Now uses the shared
  `NeverDeployedPagePredicateFor`. **Measured before changing it: both motivating pages
  are `active`/`deployed`, so it would NOT have shipped inert — 198 active vs 6
  archived tool pages fleet-wide. The fix closes a door rather than repairing a miss**,
  and I have said so rather than letting it read as a catch.
- Answered without code changes, each by a query or a citation: the fact id **is** in
  the `item_key` (the sketch hid it); a persistently-403 source is **not** permanent
  silence (the citation arm ages a fact past `staleness_days` regardless of fetch
  success, and that arrives here as evidence_drift); **one** agent type calls the
  action; the optional-key budget audit reports **0 shared actions over N=10**, so P11
  does not flip this to architecture scope.

### A design call the round forced into the open

Filing per (fact, tool) means seeding mortgagecalculator's 13-fact fence produces 13
reconciliation items at once. Collapsing to one item per tool would be tidier and is
**wrong**: a key coarser than the finding drops every finding after the first
(`bugs_open/091`), which the `bug_historian` seat raised in the same round. Kept
per-fact, banded low, and the burst is named in the CONTRIB so the receiving lane is
not surprised by it.
