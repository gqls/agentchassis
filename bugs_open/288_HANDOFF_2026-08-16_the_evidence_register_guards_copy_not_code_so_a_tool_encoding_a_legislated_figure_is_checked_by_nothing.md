# 288 — the evidence register guards COPY, not CODE: a figure a calculator ENCODES is checked by nothing, and the one mechanism that could was unreachable

**Filed 2026-08-16** as the CLASS behind `bugs_closed/225` (the SDLT calculator that
ran an expired £625,000 first-time-buyer cap for sixteen months). 225 itself is fixed
and live; this file is what 225's own section *"Why no existing check could ever have
caught this"* was pointing at, written down as a tracked defect instead of a remark
inside a closed case.

**Status: PARTIALLY FIXED 2026-08-16** — Pieces 2+3 of the existing plan are built
and committed (`989addb1c`, council `cff364b8`), **inert until the next chassis roll
and until a fence declares**. The residual is named in §5 and is why this stays open.

## Diagnosis provenance (2026-07-31 owner ruling)

**Not run through 090 — stated substitution, and the substitution is stronger than
usual here because there is no inference step to check.** Every mechanism below is a
single visible clause in code I read directly and quote inline with its file, plus
five counts I measured myself against the live database **with a positive control in
the same breath** (the `?` operator that returned 0 for `artifact_check` returned 61
for `citation`, so the zero is an absence and not a broken query). Same ground as
`bugs_open/281:8-12`. What a diagnosis loop would add — an independent re-read of the
same functions — is exactly what the quotes below already are.

## The class, in one sentence

**A fact in the evidence register constrains what a page may SAY. Nothing connected a
fact to the constants inside a calculator's JavaScript** — so a tool can compute a
number that contradicts the very register the site is governed by, and every check we
own passes it.

## Why every existing check is blind, with the clause that makes it so

Three of these are deliberate design decisions, correct on their own terms. That is
what makes the gap hard to see: nothing is broken.

| # | mechanism | the clause | consequence |
|---|---|---|---|
| 1 | every claims scan | `datahelpers/claims.go` `nonAssertionElements` — `"script": true` | a figure inside `<script>` is never text, so it is never an assertion, so it is never scanned |
| 2 | the numeric arm | `claims.go` `editorialPageTypes` — `"tool": true` | `ScanUnregisteredNumbers` does not run on a tool page's prose at all |
| 3 | prose currency | `claims.go` `isExcludedNumber` drops `£`/`$`/`€` amounts | even `£625,000` stated in FINANCE PROSE is scanned by nothing |
| 4 | the golden | `loanandmortgagecalculator_couk/golden_compare_post.py` | proves CONSISTENCY; a baseline recorded from a never-correct tool certifies the defect for ever (016b §9) |
| 5 | Tier-2 acceptance | `check_tool_acceptance.go` header: static checks *"CONFIRM, never refute"* | structurally cannot fail a page for what it computes |

Nothing in that list is asking whether the answer is RIGHT. The only thing that ever
has is `SQAM-003`, a Python oracle in a docs directory with no scheduler, no work
item and no Go caller — its own register entry says so, re-confirmed 2026-08-15.

## The mechanism that COULD have, and the three reasons it could not

RFC_025's `artifact_check` (`refresh_evidence_base_action.go`) re-proves a fact
against stored `rendered_html` — raw bytes, scripts included. It is the one reader in
the estate on the right surface. It was unreachable for this class:

1. **Zero consumers.** `[MEASURED 2026-08-16, with a positive control]` — of **143**
   current facts across 12 registers, **0** carry `source.artifact_check`; 28 carry a
   plain `artifact` source, which is checked for presence in the register and never
   re-proved. RFC_025 §10 predicted this ("still short of IMPLEMENTED… no real fact
   anywhere has actually been retyped").
2. **Unreachable for the facts that hold legislated figures.** The per-fact loop
   handles `source.citation` and `continue`s **before** the `artifact_check` test. Every
   SDLT fact on mortgagecalculator.co.uk is a citation fact, so an `artifact_check`
   added beside one would never be evaluated. `[read directly]`
3. **Addressed by a component id that dies.** `artifact_check.component_id` names a
   `page_components` row. **225's own component (`55682bc8-…`) no longer exists** — the
   page was decomposed into `prose-0` / `tool-1` / `prose-2`. A binding written when
   the bug was filed would today resolve to nothing.

## What was built 2026-08-16 (Pieces 2+3), and what it does NOT do

Commit `989addb1c`, council-submitted `cff364b8`. Concept register **CLM-022** (+
**TL-045**). This is not a new design: it is Pieces 2 and 3 of
`mortgagecalculator_couk_adoption/PLAN_2026-08-09_facts_into_tool_acceptance.md`,
whose Piece 1 has been live since migration 366 (CLM-021).

- **Piece 2** — a tool's criteria fence may declare `"facts": ["<fact_id>", …]`: ids
  only, never values, resolved against the SWEPT SITE's register (doc_plans has no
  `site_id`). Validator rule P11 refuses a malformed declaration where it is written.
- **Piece 3** — the daily evidence sweep resolves declaring PLANs to pages and files
  one item per (fact, tool): `improve_tool` only where a fixer may safely act (a
  tool-level component owns the code AND the fence has no `no_auto_fix`), the new
  handler-less `fact_drift_review` otherwise — including for **every** evidence-only
  drift — and nothing at all on a fetch error.

**Read the honest limits before quoting this as closed:**

- It answers *"did the register's figure MOVE, and who encodes it"*. It does **not**
  answer *"is the register's figure RIGHT"* — that is Piece 4 (an oracle computing
  expectations from the register), which stays behind its RFC. **A tool and a register
  that are both wrong in the same direction agree, and this mechanism is silent.**
- It needs a human to author every declaration. Today **0 of 90** current tool PLANs
  declare anything `[MEASURED 2026-08-16]`.
- It cannot see a page with no `pages` row. mortgagecalculator's repo-only
  `stamp-duty.html` twin carried the identical defect and stays invisible.
- Latency is up to 24 hours (the daily sweep), and a fence carrying `facts` asserts
  **nothing** at acceptance time — both tiers ignore unknown fence keys, by design.

## §5 — the residual, which is why this file stays OPEN

1. **Not live — MEASURED, not assumed.** `[2026-08-16, both replicas, with a positive
   control in the same exec]`:

   | replica | `fact_drift_review` | `stale_attestation` (control) |
   |---|---|---|
   | `agent-chassis-…-48lv6` | **0** | 5 |
   | `agent-chassis-…-vtfdx` | **0** | 5 |

   The control is what makes the zero mean something: the probe CAN find a symbol in
   this binary, so `0` is an absence rather than a broken grep. Re-run exactly this
   pair after the next roll and expect the first column to become non-zero. Do not
   verify at the tag, and never by the absence of errors.
2. **No live declaration.** The first is mortgagecalculator's `stamp-duty` fence,
   handed to that lane as a CONTRIB. Until it lands, this mechanism has never fired on
   real data and its green means nothing (`a-pass-from-a-blind-check-outlives-the-blindness`).
3. **The `improve_tool` arm is unproven in production and may be unreachable.** Both
   live SDLT fences carry `no_auto_fix` and neither page is a fork, so on today's
   fleet every route is human. Proven by unit test only — say so, do not report the
   auto-fix path as exercised.
4. **The prose half of the class is untouched.** Blind spots 2 and 3 above (tool-page
   prose, and currency anywhere) are still blind. A retired-figure scan over raw
   deployed HTML in the post-deploy claims audit would close it, and was deliberately
   NOT built here: it is a second scanner over the same components, which is
   `bugs_open/093`'s shape, and it deserves its own measurement of the false-positive
   rate first (the plan's §3 rejects the blanket form for exactly that reason).
5. **RFC_025 stage 2b is still owed** — `artifact_check` reachable for every source
   kind and addressable by `page_name` rather than a component id that dies. Small,
   same seam, would give RFC_025 its first consumer; not bundled here because it
   changes an existing mechanism's semantics and deserves its own round.

## The council round, and the defect it caught (2026-08-16)

Round 1 came back **REVISE**, gated by the `editquality` seat, and the gating objection
was **right in a way that would have shipped a mechanism blind to its own motivating
bug**. Recorded here because it is the most useful thing in this file:

> the baseline is defined purely from PRIOR REGISTER state … a tool that is wrong from
> the day it opts in, against a register fact whose value has been stable since
> (exactly bug 225's shape: correct register, stale code, no subsequent legislative
> change), produces baseline == current and … emits nothing.

**That is 225 exactly.** The fix: a declaration with no baseline now files one
`unreconciled_declaration` review item per (fact, tool) asking a human to confirm the
tool encodes the registered value. It is self-quieting — the item carries the value,
which becomes the next pass's baseline — and it is banded low, because 13 declared
facts means 13 one-time items and that is a backlog, not 13 alerts.

Two further findings acted on, and three answered by measurement rather than by change:

| seat | finding | outcome |
|---|---|---|
| `compliance` (med) | every human-routed case sat at medium/35, ranking a confirmed move on a `no_auto_fix` tax calculator BELOW an auto-fixable one | **fixed** — three bands; a confirmed move is high/30 whichever route it takes |
| `debug_historian` (high) | `p.status='active'` is the liveness predicate and is wrong for an AUDIT — and the equivalent check had been run for the eligibility predicate but not this join | **fixed** — now uses the shared `NeverDeployedPagePredicateFor`. Measured: both motivating pages are `active`/`deployed`, so it would not have shipped inert; 198 active vs 6 archived tool pages fleet-wide |
| `bug_historian` (high) | is the fact id in the `item_key`, or will a second drifted fact on the same tool be swallowed (bugs_open/091)? | **already correct** — `fact_drift:<fact_id>:<subject_key>:<site_id>`; the sketch had hidden it |
| `bug_historian` (med) | a persistently-403 source means permanent silence | **answered** — `refreshCitationFact` marks a fact drifted once it passes `staleness_days` whether or not the fetch succeeded; that arrives here as evidence_drift. Time-based, not attempt-based |
| `guardian` (med) | how many callers share the modified action? | **measured: one** — `evidence-freshness` |
| `architecture` (med) | does P11 cross the WFA-013 optional-key budget (N=10)? | **measured: no** — `audit-optional-key-budget.sh` reports 0 shared actions over budget |

## How to verify (after the roll)

```bash
# 1. the code is actually running (never infer from the tag)
kubectl -n ai-persona-system exec <chassis-pod> -- grep -ac fact_drift_review /proc/1/exe   # expect >0
kubectl -n ai-persona-system exec <chassis-pod> -- grep -ac stale_attestation /proc/1/exe   # positive control
```
Then seed the mcalc `stamp-duty` fence and **induce**: supersede
`sdlt-ftb-relief-cap` 500000 → 550000, fire a `refresh_evidence_base` dry run, and
assert it names `stamp-duty` with `kind:value_drift, route:fact_drift_review,
reason:no_auto_fix`; restore. **A dry run that reports nothing after a real change is
the failure**, and it is the only result that distinguishes this from an inert check.

## Related

- `bugs_closed/225` — the case. `bugs_open/224` — the sibling defect family found the
  same way (a 0% rate breaking six of seven calculators).
- `PLAN_2026-08-09_facts_into_tool_acceptance.md` — F1/F2/F3, and Piece 4.
- `bugs_open/161` — the register ratifying a false claim: the same register, the
  opposite failure (there the register was wrong and vouched for the copy; here the
  register was right and could not reach the code).
- Register: **CLM-022**, **TL-045**, CLM-021 (the builder half), SQAM-003 (the oracle
  that actually found 225, still outside the platform).
