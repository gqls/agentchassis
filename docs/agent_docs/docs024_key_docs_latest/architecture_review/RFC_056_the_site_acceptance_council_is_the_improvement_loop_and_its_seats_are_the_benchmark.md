# RFC 056 — the site acceptance council IS the improvement loop, its seats are the benchmark, and an opinion must be a verdict row before it may be a rewrite

**Status: DRAFT, raised 2026-08-25 by the `loanzy_uk_example_site` lane at the owner's request.
PARTLY RETROSPECTIVE in the RFC_002 shape: the seam change in §5 (`filing_mode: record` on
`write_audit_findings`) and three of the four seats in §4 are WRITTEN and TESTED in the tree,
council submission pending, and inert until the next chassis roll. The config halves (§3 Phase 1,
§7) are written as `_HOLD` migrations and NOT applied — the owner asked for a plan.**

Owner rulings this RFC rests on, all 2026-08-25, verbatim in substance and recorded in
`loanzy_uk_example_site/REFERENCE_2026-08-25_site_acceptance_council.md` §2 and
`OWNER_REVIEW_2026-08-25_homegarden_and_what_it_says_about_every_site.md`:

- *"The checkers will check after the fact (improvement loop)"* — placement.
- *"we have other agents like the component planner and so forth too"* — routing to existing agents.
- *"implement at least 6 different structures unless the site doesn't warrant it or refuses it"* — N = 6.
- *"I like your suggestion"* — direction approved.
- *"switch off the evolutionary aspect … it should stay … switch just that bit off and turn the
  improvement loop back on"* — the seam in §5.

Every figure carries its date. A count of things is written `N as of <date>` (owner ruling
2026-08-22) because a census goes stale by addition.

---

## 1. What the thing IS, before any rule about it

**A council** on this estate is a set of independent reviewers, each with one lens, producing a
verdict that routes work. The estate has three: the code council (`council-gate`), the experience
loop, and the concept-register reviewer. None of them has ever been pointed at a **site**.

**The improvement loop** (`improvement-loop`, carrier `improvement-sweep`, 15-minute tick, one site
per firing) is a post-hoc site reviewer with seven seats, and has been since March. Three seats are
**mechanical** — `quality-discovery-agent`, `design-discovery-agent`, `completeness-discovery-agent`,
each a `run_discovery_checks` step over a named `checks` array, 9 / 23 / 45 checks as of
2026-08-25. Four are **model** seats — `design-audit-agent` (→ `visual-design-auditor` +
`content-quality-auditor`), `site-review-agent` (role `site_reviewer`), `offer-analyser`,
`brief-fidelity-auditor` — each ending in the shared action `write_audit_findings`, which routes a
finding by category to a handler. After the seats: `record_audit_pass` (fingerprint + 14-day
cooldown + a three-strikes `capability_gap` when nothing converges), `triage_findings` (promotes
every `detected` row on the site), `insert_rerender_item`, `call_dispatch`.

**It was dark from 2026-08-17 12:30Z**, hand-fired once on 08-22, and the owner's review of the
`homegarden.uk` canary (2026-08-25) found six broken routes on one site, every one an ABSENCE with
status `complete` — research never ran, the offer analyser was off, the visual designer has no
dispatch path, the news feed cannot bootstrap, a listing rendered nothing, the classifier cannot
name a directory. **That review was the estate's only closing check.** The loop is the reviewer he
asked for; it exists; it was off; and it lacks four lenses.

## 2. The rule this RFC proposes

1. **The improvement loop is the site acceptance council.** No new orchestrator. Placement is
   after the fact (owner), so a new site and one of the 31 active sites are judged by the identical
   operation, and the first fleet pass is a census.
2. **Its seat definitions are the benchmark.** A seat is either mechanical (a predicate that could
   have come out the other way, free, runs every pass) or model (relevance-gated, costs credits).
   Prefer mechanical where a query can answer — a model seat can be wrong in the confident register
   and a query cannot.
3. **Defects dispatch; opinions record.** A mechanical finding may route to a fixer as today. A
   model seat's finding is a **verdict row** — filed, deduplicated, provenance-stamped, and
   undispatchable by construction until a person or a deliberate migration releases it. This is the
   seam in §5, and it is the "evolutionary aspect" switch the owner asked for.
4. **A seat that did not run must not be recorded as having passed.** The audit stamp is withheld
   when any seat failed, and the failure leaves a durable row (§6).

## 3. How the current loop measures against the rule

| rule | today `[all MEASURED 2026-08-25]` |
|---|---|
| 1 — one council | the loop is it; off since 08-17; `improvement-sweep.enabled=false` |
| 2 — benchmark seats | 7 seats; 4 lenses missing (§4b): prerequisites, promise, structure, reader |
| 3 — opinions record | **every** model finding dispatches, by **two** doors: the loop's own `triage_findings` (site-wide, no type filter) and `detected-item-promoter` (live since 08-15, every 15 min, promotes any `detected` row whose (type, handler) pair ever completed). **26 model-seat rows were promoted 08-20 → 08-24 while the sweep was OFF.** Lifetime from the `design-audit` source alone: 976 `content_rewrite`, 399 `needs_content_page`, 964 `needs_content_planning` (live ∪ archive). `bugs_open/238` is what one rewrite did to a page that was fine |
| 4 — no false pass | 7 of 8 seat calls have `error_step == next_step`; the 8th (`call_completeness_discovery`) errors to `triage_findings`, skipping four seats. Both `enrich_*` steps carry a CONFIG-level `error_step` that continues. Fail-open surface: **9 steps**. The only record of a seat failing is `orchestration_states`, reaped in ~25h |

The evolutionary switch is therefore not "turn the sweep off" (it was off, and rewrites continued
through door two); it is rule 3, and it needs the seam in §5.

## 4. The seats

### 4a. Existing (unchanged in what they observe)

| lens | seat | kind |
|---|---|---|
| Completeness | `completeness-discovery-agent` | mechanical, 45 checks |
| Integrity / quality | `quality-discovery-agent` | mechanical, 9 checks |
| Design | `design-discovery-agent` | mechanical, 23 checks |
| Visual + content quality | `design-audit-agent` → `visual-design-auditor`, `content-quality-auditor` | model |
| Strategic alignment | `site-review-agent` (role `site_reviewer` — resolved: `spawn_site_review` fills it; 4,046 LLM calls lifetime) | model |
| Offer | `offer-analyser` | model |
| Fidelity to brief | `brief-fidelity-auditor` | model — judges against the brief, and the brief can be the problem (homegarden's was anti-commercial) |

### 4b. The four added — three WRITTEN as discovery checks (flag-only, inert until named), one specified

All three checks are registered via `init()` and **run only when named in a `checks` array**
(IMP-004's safe-rollout pattern). All file with `HandlerAgent ""` and `Status "detected"` — the
estate's flag-only idiom — so neither promoter can ever dispatch them; they are graded by their own
next pass (`CheckResult.Resolved` on a positive re-observation, never on absence). Item keys follow
`{itemType}:{target}`. Each declines to judge on a parked domain (the invented-path control 200s →
error, nothing filed). Their item types are declared as acknowledged verifier gaps
(`verifier_coverage_test.go`).

| lens | check · item type | rule (v1 — the architecture seat is asked to rule on the shapes, §9 Q2) |
|---|---|---|
| **Prerequisites** | `build_prerequisites` · `prerequisite_missing:<kind>:<site>` | four kinds, each a SQL predicate: `vertical_landscape` (current `site_specs` row), `page_research` (any `research_results` row), `evidence_base` (current, non-empty object — **bugs_open/380 D1: observe absence, never mint**), `feed_sources` (an active `content_sources` row — `bugs_open/316`: an unenrolled site cannot enrol, a route defect). Baseline of 31 active sites: 23 / 10 / 19 / 9 |
| **Promise** | `heading_promise` · `heading_promise_unmet:<page>` | a served page's own h1–h3 promises a structure the non-anchor body lacks: calendar/month-by-month → ≥6 distinct months; checklist/step-by-step → ≥3 non-anchor `<li>`; compare/vs → ≥1 table; top-N/best-N → ≥N items. Chrome-stripped (nav/header/footer/style/script removed, anchor text discounted — every homegarden page carries all 12 month names as MENU links). **Nominates, does not adjudicate** — a keyword is not a promise |
| **Structure** | `structure_floor` · `structure_floor_unmet:<site>` | fewer than **N = 6** distinct DELIVERED reader-facing structures across the served site (list, table, calendar, checklist, comparison, tool, directory, feed, guide, faq — hero/nav/footer/CTA excluded) and no recorded refusal. N overridable and a refusal recordable at `sites.settings->maintenance_profile->structure_floor` (`n`, `refusal`); a refusal is a RECORDED verdict and retracts the item. homegarden: 21 pages, 17 identical, one structure |
| **Reader** | `reader-experience-auditor` · model, `filing_mode: record` from birth (§7, not yet written) | *would a person on this kind of site want this page?* Distinct from fidelity-to-brief (§4a) and from `content-quality-auditor` (tone, not purpose). The owner asked for it three times in three words. Evidence: `about.html` 14 of 17 headings about the site's own methodology; "Get Started → contact" on a gardening site |

Freshness and depth (`REFERENCE` §4b) are deliberately deferred: a depth seat before
`research-agent` is driven files the same finding on every site, which is a census, not a review.

## 5. The seam: `filing_mode: record` on `write_audit_findings` (WRITTEN, tested, pending council)

**What it is.** An opt-in step setting, declared in the action's `ConfigKeys` (a setting, not a
data reference; the optional-key budget of RFC_022 counts `len(Optional)` and is unmoved — the
parity test passes). Absent or `dispatch` → historical behaviour, byte-identical output. `record` →
every routable finding is filed as the **same row** (type, dedup key, spec, page) but parked:
`status 'deferred'`, `handler_agent ''`, and in spec: `routed_handler`, `routed_status`,
`filing_mode`, `deferred_by`, `deferred_reason`, `release_recipe`. Any other value is an **error**
— a typo must not silently dispatch a page rewrite.

**Why `deferred` + `''` and not `detected`.** IMP-054 believed `detected` was a safe shelf ("a lone
discovery run files findings nothing can ever see"). That stopped being true on 2026-08-15 when
`detected-item-promoter` went live. `detected` is a queue. `deferred` + empty handler is the one
convention nothing promotes (`capability_gap`, `bugs_closed/077`, `remit.go`), and `bugs_open/396`
(the second of that number) is the warning against the other shape — `deferred` WITH a named
handler and no provenance, the "park with no park verb" — which is exactly why the routing moves
into spec with a stamp and a recipe. Both doors are named in the test
(`TestRecordOnlyFinding_IsUndispatchableByBothPromoters`) so a half-fix on either fails.

**RFC_022 shape check.** Opt-in; the *unsafe* side (dispatch) is the default; live consumers
naming it: **zero as of 2026-08-25** (five live `write_audit_findings` steps, all carrying exactly
`site_id / audit_source / findings_field`). By the 2026-08-11 ruling that is not architecture-scope
on its own. It is in this RFC because it is the mechanism by which rule 3 is enforced fleet-wide,
and because the consumers it will be switched on for (§7) are shared seats.

**What it does not do.** It does not change what a seat observes or how the router classifies; it
leaves already-parked rows (`capability_gap`) untouched; it releases nothing. A record row holds
the finding's dedup slot, so a later dispatch-mode filing of the same finding is a duplicate until
the row is released or closed — one row per finding, whichever mode filed it.

## 6. The fail-open surface, and durability (from the `vigilant_designer_offer_analysis` lane)

`error_step` has **two valid spellings** — step-level and `config.error_step` — and
`routeToErrorStepOrFail` (`coordinator.go:3916`) honours both; a census of one under-reports,
confidently and low. Counting both: seven `call_*` seats route failure where success routes; the
eighth skips four seats; both enrichment steps continue past their own failure. A seat's own
description says *"the child run is the record of whether it worked"* — and that record is reaped
in ~25h. Enable the sweep as it stands and the plausible outcome is a clean audit stamped on a site
nothing audited (`bugs_open/395`'s false green at fleet scale).

Proposed (config only, §7's migration): each `call_*` seat's `error_step` goes to a
`record_<seat>_failed` step that files ONE `deferred` `capability_gap` row per site per seat
(`capability: audit_seat_failed`, `recurrence_expected`, the `record_not_converging` shape) and
continues to the seat's original next step; `record_audit_pass` is skipped when any seat failed.
Continue-the-sweep is kept (one auditor must not strand it); the pass stamp is what becomes honest.

## 7. Phasing (the plan the owner asked for is `loanzy_uk_example_site/PLAN_2026-08-25_switch_off_the_evolutionary_rewrites_and_switch_the_loop_back_on.md`)

| phase | what | kind | state 2026-08-25 |
|---|---|---|---|
| 1 | `call_completeness_discovery.next_step: spawn_design_audit → record_audit_pass` — the four model seats bypassed, not deleted; then `improvement-sweep.enabled = true` | config | migration `623_…_HOLD` written + rehearsed; **owner's word** |
| 2 | `filing_mode: record`; checks `build_prerequisites`, `heading_promise`, `structure_floor` | Go | in tree, tests green; council submission with this RFC; inert until roll |
| 3 | name the three checks; create `reader-experience-auditor`; `filing_mode: record` on all five model seats' write steps; restore the Phase-1 edge; seat-failure records (§6) | config | to write as `624_…_HOLD`; apply after roll + verdict |

## 8. Verdicts and routing

Same three verdicts as the code council. After the fact, so: **APPROVED** → `record_audit_pass`;
**REVISE** → one row per seat finding — a verdict row (model seats) or a dispatchable defect
(mechanical seats), routed to the agent that owns the stage (planner: structure, freshness;
writer: reader, depth; imagery; component planner); **REJECTED** → site live but judged unfit,
escalates to the owner. The loop's three-strikes `capability_gap` stands as "the fixer cannot fix
this".

## 9. What the architecture seat is asked to rule on

1. **Rule 3 as stated** — is "defects dispatch; opinions record" the right line, and is the
   mechanical/model boundary the right proxy for it? (The alternative is IMP-006's per-site
   `approval_mode` column, proposed four times, never built.)
2. **The v1 rubrics** in §4b — the structure shapes (ten), the promise rules (four), the four
   prerequisite kinds. A count of 6 met trivially by chrome would mean the shapes are too loose.
3. **BLD-006 superseded, not joined** — the register's "coverage baseline: guides, tools, news,
   curated top-N" (status aspirational, enforcement "the strategist/planner prompts") becomes the
   structure and prerequisite seats, enforced at the artefact.
4. **`380` D1 and the taxonomy** — the prerequisite seat asks for research that *populates* a
   register and never mints one; a structure seat that cannot find a `directory` type files
   against the taxonomy (15 layout categories, none `directory` or `news`), not against the site.
5. **§6's routing question** — should a seat failure fail the sweep CLOSED (strand it) or continue
   with a durable record and no pass stamp? This RFC proposes the latter; the peer lane asked that
   it be argued here rather than settled in a config patch.
6. **The promoter as a second door** — `detected-item-promoter` voids IMP-054's premise. Should
   flag-only and record-only be *statuses the promoter reads* rather than conventions it happens to
   respect (`COALESCE(handler_agent,'') <> ''`)? RFC_039's status-vocabulary question, one level up.

## 10. What this RFC does not do

It does not fix the six broken routes (`bugs_open/376`, `/384`, `/316`, the visual designer's
path, the taxonomy gap) — it finds them every pass. It does not judge copy (the
`copy_quality_two_stage` lane). It does not touch the render-audit rotation, the OTHER live source
of bad renders (239 `contrast_failure` completions in 14 days as of 2026-08-25), which is the `390`
lane's and an owner decision.

## 11. Falsifiers

- A record-mode row dispatched → one of the two doors changed; the test names both.
- The sweep on with the seats bypassed and bad renders continuing at the same rate → the cause was
  the mechanical fixers or the render audit, and §10's last sentence is the plan.
- Refusals filed on most sites → N too high for the vocabulary, or refusing is cheaper than composing.
- The reader seat agreeing with `brief-fidelity-auditor` everywhere → one is redundant.
- A structure count of 6 met by chrome → §4b's shapes are too loose.

---

## ADDENDUM 2026-08-25 (same day) — the parked-verdict lifecycle question, raised and answered

Raised by `vigilant_designer_offer_analysis` after the Phase-3 consumer notice, verified by them in
both directions before asking; recorded here because they asked that the RFC answer it rather than
that a config patch settle it.

**Their finding.** `deferred` is not in `workItemTerminalStatuses`, so a record row HOLDS its dedup
slot — good, no duplicate pile-up, and a reader will assume the opposite, so it is now stated: **a
record row is an OPEN row; the same finding cannot be filed twice while it stands.** But `deferred`
is also not in `workItemRevalidatableStatuses` (`["needs_human_review","unresolved"]`,
`work_items_common.go:142`), so no generic revalidator ever re-checks one — and `[their MEASUREMENT
2026-08-25]` 313 rows sit at `deferred` today, oldest 2026-07-26, none draining. Is a verdict row
parked for ever even when the page changes under it?

**The answer, from the mechanism that already exists.** Verdict rows are revalidated — by the seat
that made them. `write_audit_findings` runs a **silence retraction**
(`write_audit_findings_retraction.go`): rows whose `spec.audit_source` matches the running seat and
whose status is outside closed/in-flight — the file names `deferred` explicitly as a status that
"can accrue a silence observation" — take a streak write each run the seat does NOT re-file the
finding, and are retracted at 3 consecutive silences. Every classified finding's base spec carries
`audit_source` (`write_audit_findings_action.go:424`) and the record transform copies the whole
spec, so record rows are in scope by construction. So:

- **the only machine that closes a verdict is the seat that made it, by withdrawing it** — a
  re-observation by the same lens, which is the one closure that cannot contradict the verdict;
- freshness is bounded by the audit cadence: a verdict that stops reproducing is gone within ~3
  audits of that site (fingerprint change or 14-day cooldown, ×3);
- **`workItemRevalidatableStatuses` is deliberately NOT widened.** The generic revalidation sweep
  auto-closes on a `resolved` verdict; pointing it at rows designed for a person to release would
  let a mechanical predicate overrule a judgement lens. If it were ever widened,
  `reviewRevalidators` is the seam (registering one derives `coveredItemTypes()` in the same edit)
  — their pointer, kept here so it is not re-derived.
- The 313 pre-existing `deferred` rows are a DIFFERENT population (hand parks and capability_gap
  rows, `bugs_open/396`) — most carry no `audit_source` scope, no seat re-runs over them, and this
  RFC does not claim the retraction drains them.

**Falsifier for this addendum:** a record row whose finding stopped reproducing still open after
its site's third audit at a new fingerprint — then the retraction's scope check is not matching
record rows and the claim above is wrong at the join, not the design.

---

## ADDENDUM 2 (2026-08-25, late evening) — round 1 came back REVISE, and what each objection turned out to be

Corr `d1342f2a`, gating objection from `editquality`. Triaged the way the estate's practice says to
— a revise round is cheaper than the defect it finds, and this one found four real ones. Each line
below says what was CHANGED or what the checked answer IS; nothing is defended that was not measured.

**Real defects, fixed:**
1. **Migration number collision** (guidelines): `619` was already the live `619_cta_bg_is_not_a_colour.sql`,
   and `621`/`622` were taken the same day. The HOLDs are renumbered **623** (bypass) and **624**
   (seats). Every reference in this RFC, the PLAN, the handoff and the register is updated.
2. **The seats-gate semantics are now PROVEN, not watched** (guardian HIGH, bug_historian HIGH):
   `conditional_branch_seats_gate_test.go` pins `compareValues(nil, "capability_gap") = false` (a
   clean sweep stamps the pass) and one present field tripping the bracket-free OR chain (a failed
   seat withholds it). Bonus finding while pinning the footgun: the empty-literal hole
   (`nil == "" → true`) is UNREACHABLE in string-form conditions — the evaluator trims the
   expression first, so a trailing ` == ` never parses — and the test pins that too.
3. **The withheld pass stamp interacted badly with the cooldown** (improvement_guardian): a
   persistently failing seat would have put its site on an every-sweep audit treadmill. 624 now
   routes a failed-seat sweep to `record_audit_attempt`, which stamps `last_audit.at` (the 14-day
   cooldown holds) while touching neither the fingerprint nor `passes_at_fingerprint` — no PASS is
   counted, rule 4 holds, `not_converging` still needs three real passes. 48 steps, re-rehearsed.
4. **`heading_promise` filtered pages on a raw `status = 'active'`** (debug_historian) — the exact
   spelling the landmine warns may filter on nothing. It now carries
   `datahelpers.PageWantedLivePredicateFor("p")` like its sibling, with the round credited in the
   code comment.

**Checked answers, no change:**
5. **`approval_mode` exists — why a new convention?** (reuse_agent HIGH). Measured, not argued: the
   column is read at exactly ONE door — `load_work_item_actions.go:726`, the dispatcher's loader —
   and `[MEASURED 2026-08-25]` all **11,964** rows are `'auto'`; the review path has never been
   exercised. A verdict row held by `approval_mode` would be promoted to `triaged` first (the
   promoters do not read the column) and then held **in-flight** — where the silence-retraction
   explicitly may not write (`triaged` is in `workItemInFlightStatuses`), so it would be exactly the
   stale-forever verdict ADDENDUM 1 rules out, invisible to every census as "in flight". The
   `deferred`+`''` shape is refused by BOTH doors and lives inside the retraction's scope. Stated
   rather than assumed now; and IMP-006's `approval_mode` was a per-SITE proposal in any case.
6. **`bugs_open/083` reconciliation** (bug_historian): 083 is ACCIDENTAL starvation — `detected`
   rows nothing ever promotes. Record rows are the opposite on every distinguishing axis: parked
   `deferred` (not `detected`), by DESIGN, stamped (`spec.filing_mode='record'`, `deferred_by`,
   `release_recipe`), and lifecycle-owned by the seat's silence-retraction. An on-call engineer
   tells them apart by one predicate: `spec->>'filing_mode' = 'record'`. Named in 624's doc_notes row.
7. **`''` vs NULL at the post-078 guards** (bug_historian low): both promoter doors already coalesce
   — `COALESCE(handler_agent,'') <> ''` (promoter pre_query; `workItemRoutableSQL`) — so empty
   string and NULL are one value at every door that matters.
8. **The 5-step blast radius might undercount nested steps** (guardian): re-censused by TEXT over
   the whole config, not a top-level steps walk: 7 active rows contain the string; 5 are the model
   seats' action steps; `council-gate` and `fix-proposer` carry it as roster PROSE (0 steps with
   `action = 'write_audit_findings'`). The count is 5, now measured the blind-proof way.
9. **Prior art** (prior_art_librarian): `acceptance-discovery-agent` / `reader-experience-auditor`
   checked against the live registry — the adjacent types are `experience-approval-council`,
   `experience-planner`, `experience-register-writer`, `tool-acceptance-agent`, none a site-facing
   acceptance seat (the planner has no dispatch path at all, measured this morning). `filing_mode`
   has no prior occurrence in the repo or the register; IMP-006 records the four earlier PROPOSALS,
   none built. `check_missing_structure.go`, despite its name, checks CHROME (header/footer/head
   rendered) — different domain from `structure_floor`; `check_voice_tells` scans prose register,
   not heading-vs-body structure. Both now named here so the adjacency is recorded.
10. **spec-based stamps vs `refreshOpenWorkItem`** (raised by nobody, found while answering: the
    396 lane's park verb, migration `621_park_work_items_verb.sql`, landed the same evening and
    chose `result` because `refreshOnConflict` producers REPLACE `spec` wholesale). Record rows are
    safe in `spec` for a reason that must be stated, not lucky: `write_audit_findings` inserts with
    `dropOnConflict` (never refresh), and its dedup keys are audit_source-prefixed, so the only
    producer that can ever touch that key is the same seat through the same non-refreshing path.
    If a refreshing producer is ever pointed at these keys, the stamps die — the park verb's
    `result` channel is the escape hatch, and the two mechanisms are siblings, not rivals: 621 is
    the HUMAN park (dry-run default, unskippable stamps), record mode is the MACHINE park.
11. **Gate 1c interaction** (vigilant lane, post-roll): record rows never reach `complete`, so the
    new completion gate never grades that producer while record mode is in force. Correct on both
    sides — parking is the safer state — and that producer's gate-1c census going quiet is
    EXPECTED, not evidence; stated so neither lane misreads an empty census.
    **EXTENDED 2026-08-25 (their addition, lifted here so the pair reads safely in either order —
    their register entry WII-033 carries the same list as of `6ef02131a`):** an empty
    `_verification.acceptance_predicate` census has meant THREE different things in one day, and a
    reader who inherits the wrong reading draws the opposite conclusion each time. Put the meaning
    next to the query, never leave it to be inferred:
    - **before the 19:07Z roll** — "the gate was never switched on" (Go inert);
    - **after it** — "no completed item carried a predicate" (an adoption fact; says nothing about
      refutation);
    - **under `filing_mode: record`** — "this producer's findings are parked by design" (they never
      reach `complete`, so the gate never sees them).
12. **The release path is raw SQL only** (architecture, low — accepted as a named follow-up): a
    verdict-queue listing (and eventually a release surface) is owed before record rows accumulate
    across the fleet; until then the RUNBOOK query and the per-row recipe are the interface.

**Stated gap, new (from fixing the seat arithmetic):** the seat-failure rows are at the LOOP-CALL
level; `design-audit-agent`'s own internal error routes swallow a CHILD auditor failure
(`call_visual_auditor` error → `spawn_content_auditor`; `call_content_auditor` error → `complete`)
without erroring the call, so a child-level failure leaves no row and does not withhold the pass
stamp. Closing it means editing design-audit-agent's workflow — a follow-up on this RFC, not
smuggled into 624.

**Also corrected while revising:** the deploy check in 624's header. The CLAUDE.md `build
provenance` log grep is REFUTED (`[MEASURED 2026-08-25]`, bugs_open/395 lane, confirmed
independently: the string is emitted nowhere in this repo's Go source). The working route is
`service_binary_capabilities` + `git merge-base --is-ancestor` with a present-control and an
absent-control — and by that route the fleet's 19:07Z roll to v1.0.1339 (commit `a7459a44b…`)
**already carries `c440d5c5e`**: the seam and the three seats are in the running binary, and 624's
precondition (1) is met today.
