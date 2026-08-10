# HANDOFF — mortgagecalculator.co.uk — cold start, read this first (2026-08-10c)

**Supersedes `HANDOFF_2026-08-10b_continue_here.md`** (read it second: its §0 owner
rulings and §1 findings all stand; its §3 action 1 — the twelve tool PLANs — is
**DONE for the eight tools that can have one**, and §3's framing of that action was
wrong in four ways, all recorded in §2 below). Site: UNLOCKED. The design this lane
executes: **`PLAN_2026-08-09_facts_into_tool_acceptance.md`**.

## 0. Owner rulings in force (unchanged)

1. Correctness beats fidelity — never copy a wrong original; improve past it.
2. The checker proves results don't differ on identical inputs, and catches wrong
   results. 3. Site stays unlocked. 4. Everything runs from the framework.
5. Both-right-differently → supply BOTH, the alternative on its own signposted page.

## 1. What landed — A4 is done, and the ladder is ON for the first time

**Eight tool PLANs are live in `doc_plans`**, keys `bridging-loan`,
`equity-release`, `fee-analyser`, `overpayment`, `rate-forecaster`, `repayment`,
`simple`, `stamp-duty`. 80 assertions, all `computed_values`,
`profiles: ["desktop"]`, `no_auto_fix: true`.

**Every pinned value was re-derived from something that is not the page's own
script** — that is the whole of the difference between this and a golden.
`acceptance/verify_criteria.py` recomputes each one and reports the strength
separately, because the three are not equivalent:

| strength | n | source |
|---|---|---|
| DEFINITION | 56 | the published formula, via the neighbouring lane's `oracles.py` (reused, not re-written) |
| **REGISTER** | **4** | **stamp-duty: the bands built from this site's 13 registered SDLT facts** — each a scalar with its own verbatim GOV.UK quote, re-verified daily |
| CONVENTION | 20 | the tool's own design choice (rate-forecaster's 24/36 phase split; fee-analyser's "total cost") — weaker, and labelled so |

**80 of 80 agree — and the control is what makes that mean anything.**
`verify_criteria.py --mutate sdlt-ftb-relief-cap=625000` (the SUPERSEDED
pre-April-2025 cap) makes the £595k FTB vector expect **£14,750: the original
tool's wrong figure, to the pound.** Put the expired law back into the register
and the expired answer comes back out. That single run establishes the register
is the *source* of the expectation rather than decoration beside it — which 80
agreements cannot.

**Anything not re-derived was DROPPED, not pinned** (41 assertions). That rule
replaces the neighbouring lane's substring container heuristic and does the job
better: containers, prose breakdowns and echoed inputs fall out of it
automatically, with no separate rule to maintain.

## 2. Four things the previous handoff had wrong — all in the same direction

Each made the job look smaller and safer than it was. Two of them would have
produced work that *looked* finished and was inert.

**(a) The subject key is NOT the page name, and getting it wrong fails silently
and permanently.** Both tiers derive it (`tool_eligibility.go`,
`toolSubjectKeyExpr`): `cc.function` when the component is
`component_level='tool'`, else `regexp_replace(p.name,'^tool-','')`. Our
recreated pages carry a SECTION component, so `tool-stamp-duty` is keyed
**`stamp-duty`**. Under the wrong key, Tier 2 goes on writing `needs_criteria`
and Tier 4 emits nothing — **no error anywhere, and indistinguishable from having
written no PLAN at all.** Corroborated by the platform's own opinion, not just by
reading Go: `doc_notes` already carried `needs_criteria` rows under `simple` and
`stamp-duty`, days old. Filed fleet-wide to `LANDMINES.md`.

**(b) Three of the "twelve recreated tools" are not ladder-eligible.**
`tool-affordability` has two components (the rule wants a SOLE component on a
`page_type='tool'` page); `game-fact-finder` is `page_type='game'`;
`investor-index` is `page_type='section-index'`. The population is **nine**.
A PLAN for the other three is a row nothing loads.

**(c) Installing a PLAN switches Tier 2 ON, and Tier 2 can fail a page for three
reasons the fence says nothing about.** `check_tool_acceptance.go:478-500` appends
`shell-doc-header`, `shell-template-residue`, `shell-dead-controls` *outside the
criteria loop* — so the neighbouring lane's stated guard ("with only
computed_values in the fence, Tier 2 finds nothing it can fail") is **incomplete**,
and `no_auto_fix` does not cover it either (that is read on the Tier-4 path). A
failure raises `improve_tool` carrying `spec.component_id` — for these pages the
shared **`hero` component: 252 pages across 18 sites**. Measured before
installing, with the platform's own functions in a scratch Go module: **all twelve
pages pass all three, and the red was induced** in the same run. `[MEASURED
2026-08-10 — a fact about today, not a guarantee. One no-op anchor left by a copy
edit is enough. Nothing structural prevents it.]` Filed fleet-wide to `LANDMINES.md`.

**(d) "Zero acceptance runs have ever happened on this site" was already false**
when 10b repeated it — two completed 08-09 20:56–21:04 for the generator-built
companions, both PASSED. Re-measure; do not carry a count from any handoff,
including this one.

## 3. Proof the fences execute — all eight driven, all eight PASSED

A PLAN with no run is a claim, not a result. Every fence was fired by hand
(the due sweep cannot see seven of them — §4) and every one came back green,
19:05–19:16Z, **4/4 checks on desktop, mobile skipped by design**, zero
`acceptance-fail` notes:

| tool | Tier-4 verdict | worth noting |
|---|---|---|
| stamp-duty | **PASSED 4/4** | the register-driven one — this site's first acceptance run on a recreated tool |
| bridging-loan | **PASSED 4/4** | |
| equity-release | **PASSED 4/4** | the neighbouring lane's equivalent FAILED at 03:28 today on state bleed (`#dispAge` read 130, expected 65). Ours drives `#erAge` absolutely in every vector, so each check sets its own state |
| fee-analyser | **PASSED 4/4** | both supply-both figures (`tcTotal` and `tcOutlay`) asserted together |
| overpayment | **PASSED 4/4** | thinnest fence — 1 assertion per vector, see §6 on `#saveTime` |
| rate-forecaster | **PASSED 4/4** | **the U+2212 assertion survived every hop** — python → JSON → psql → `doc_plans` → fence extraction → Kafka envelope → headless Chromium → exact-text compare. `computed_values` allows whitespace and nothing else, so this is the encoding proof, not just the DB round-trip |
| repayment | **PASSED 4/4** | |
| simple | **PASSED 4/4** | the only one the due sweep would have reached on its own |

**And nothing was armed by installing them:** zero `improve_tool` and zero
`acceptance_stuck` items fleet-wide in the three hours around this work, and the
only items anywhere naming the shared `hero` component (`23f95f00…`) are these
eight acceptance runs. That is the §2(c) risk measured after the fact as well as
before it.

`computed_values` is demonstrably able to fail in production — **41
`acceptance-fail` notes fleet-wide against 118 runs**, including the
`mortgages-equity-release` one above, whose detail is the exact
`computed value(s) diverge` arm. So the check type is not inert, and that did
**not** have to be established by deliberately breaking one of our own live
fences.

## 4. THE OPEN DECISION, and it is yours — 7 of the 8 fences will never fire on their own

`check_tool_acceptance_due.go` gates on `PageHasShippedPredicateFor` =
`NOT (deployed_at IS NULL AND build_status <> 'deployed')`.

| page | build_status | deployed_at | swept automatically |
|---|---|---|---|
| `tool-simple` | deployed | 2026-08-09 | **yes** |
| the other seven | `needs_rebuild` | **NULL** | **no** |

All seven **serve HTTP 200** — every one was fetched. They were built and
deployed; `deployed_at` was simply never stamped. This is already recorded
platform-side: `datahelpers/links.go:304-308` measured it on 08-09 and names
**"nine mortgagecalculator.co.uk pages, almost all `build_status =
'needs_rebuild'`"** as its worked example. Our seven are inside that nine.

**I have not stamped `deployed_at` or flipped `build_status`.** Either would
assert a deploy event nobody observed, on rows parked in a queue (`needs_rebuild`)
that MEMORY records as dead and that another lane may own. The options:

1. **Stamp `deployed_at` on the seven** — the ladder becomes continuous here.
   Cheap, and defensible given they serve; but it writes a timestamp for an event
   we are inferring.
2. **Leave it and file runs by hand** — honest, and it does not scale.
3. **Fix the class** — the predicate treats "never built" and "built, serving,
   never stamped" identically. `links.go` already carries the weaker
   `PageMayBeLinkedPredicateFor` written for exactly this distinction, so the
   shape of the fix exists. That is a platform change and wants the council gate.

## 5. Next actions, in order

1. **The §4 decision.** Until it is taken, this site's acceptance ladder covers
   one tool automatically and seven by hand.
2. **`portfolio` has no PLAN, and needs a hand-written one.** toolgolden derives
   vectors by scaling the page's own defaults; that form has none, so it drove
   `#mortgageTerm` to 1000/2000/500/450 years and the tool correctly refused all
   four. **Every emitted assertion is the validation message.** A fence built
   from that would certify an error message. Write the vectors by hand (a real
   term, a real rate), re-derive with `verify_criteria.py` (add a `m_portfolio`
   model), then install.
3. **Finish A3 — `tool-generator` and `tool-improver`.** Add the `read_site_spec`
   step plus the prompt clause, mirroring 366. Config only. **Read 10b §1(b)
   first**: these agents will also delete what the register omits.
4. **Sweep the other tools for unregistered constants BEFORE any rebuild.** The
   check is in `LANDMINES.md`. Unchanged from 10b §3.3 and still the right order.
5. **Report the stamp-duty ORIGINAL defect to the owner** — £5,000 under-quote at
   £595k FTB. Now reproducible on demand: `verify_criteria.py --mutate
   sdlt-ftb-relief-cap=625000` regenerates the original's wrong £14,750 exactly.
6. **Id-alignment on the three stragglers** (affordability, fact-finder,
   portfolio) — note §2(b): two of the three cannot enter the ladder at all until
   their page/component shape changes, so id-alignment alone will not fence them.
7. Phase B, then the Phase C RFC. **Phase C's blocker is gone twice over**: A1
   made the bands readable as scalars, and `verify_criteria.py`'s
   `sdlt_from_register()` is a working reference implementation of the oracle the
   RFC proposes — lane-side, so it changes nothing about what a green fence
   claims today.

## 5b. Fleet weather at close of session (2026-08-10 ~22:00Z)

- **Chassis rolled to `v1.0.1283`, pods started 21:43Z.** No commit in the last
  day touched the acceptance path (`tool_eligibility.go`,
  `check_tool_acceptance*.go`, `tool_acceptance_actions.go`,
  `run_checks_action.go` — checked by `git log` on those paths). Everything this
  lane shipped is **config** (`doc_plans` rows, `site_specs` facts, migration 366
  prompt text), so the roll neither delivers nor breaks any of it; all eight
  acceptance runs completed 19:05–19:16Z, before the roll.
- **A batch-8 landmine (staged_component_build lane, commit `68b7d78da`) narrows
  what our fences CLAIM: `computed_values` reads a `display:none` subtree, so it
  pins arithmetic and says nothing about visibility.** Measured against our eight
  the same evening: **none has a reveal pattern** — every results container is
  visible from load and updates in place; the only `display:none` rules on all
  eight pages are chrome (`.header-cta` under a mobile media query) and the one
  `hidden` attribute is fee-analyser's error `<p>`. So the gap is EMPTY here
  today. It stops being empty the day a rebuild introduces a
  hidden-until-submit panel — if that happens, carry the reveal in its **own**
  `interaction` check with `expect.selector` naming the revealed state, per the
  landmine, and prove it with a reveal-only mutant.

## 6. Landmines live on this work

- **All of 08-08b §4, 08-08c §3, 08-10 §5 and 10b §4** still stand.
- **NEW — a tool PLAN under the page name is never read** (§2a). Fleet-wide.
- **NEW — installing a PLAN arms Tier 2's three built-in shell checks against a
  shared component** (§2c). Fleet-wide.
- **NEW — a sign can sit OUTSIDE the currency symbol.** `<U+2212>£961.61`. A
  `re.search(r"-?\d…")` needs the sign adjacent to the first digit, so it returns
  a POSITIVE number and a fall reads as a rise. It reported a correct calculator
  wrong by £1,923.22 and was caught only because an independent model disagreed —
  the same parser on both sides would have agreed with itself. `WRONG_CALLS.md`.
- **NEW — a one-month disagreement on a duration is a CONVENTION, not a bug.**
  The pages stop amortising at `remaining > 0.005`; `oracles.amortise` at `1e-9`.
  `#saveTime` was modelled, disagreed on 3 of 4 vectors, and was **dropped rather
  than argued into agreement** — pinning either number would make one side's
  sub-penny convention into the other's law.
- **`doc_plans` has no `site_id`.** The keys we now hold — `simple`, `repayment`,
  `overpayment`, `portfolio` — are **fleet-global and extremely generic**.
  Measured today: each is claimed by this site alone. That is luck, not design;
  a second site adopting a page named `tool-simple` collides onto our PLAN. Worth
  a check before any other lane fences a ported calculator.
- **A neighbouring lane is active in the same machinery.** Re-measure `doc_plans`.

## 7. Files of record

This dir: `PLAN_2026-08-09_facts_into_tool_acceptance.md` (**the design**) ·
`NOTES` (08-10c = A4, the four corrections, and the misstep) ·
`README_where_we_are` (owner's log) · **`RUNBOOK` §14 (writing tool PLANs — the
subject-key trap, the Tier-2 check to run before `--apply`, the by-hand run
insert)**, §11–13 · **`acceptance/verify_criteria.py`** (the re-derivation, with
`--mutate` as its own red control), **`acceptance/install_fences.py`**,
`acceptance/criteria/*.criteria.json`.
Fleet: `LANDMINES.md` (+2), `WRONG_CALLS.md` (+1).
Migrations: `sql_for_agents/366_*` (+ ROLLBACK).
Bugs: `bugs_open/218`, `bugs_open/222`, `bugs_open/225`, `bugs_open/178`.
