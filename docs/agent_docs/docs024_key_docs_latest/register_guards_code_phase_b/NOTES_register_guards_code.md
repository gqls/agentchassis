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
