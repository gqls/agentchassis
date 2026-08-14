# HANDOFF — vigilant designer + offer analyser (2026-08-14)

**COLD-START = this file + `PLAN_2026-08-02` (programme + the full decision log) +
`features_open/030`. NOTES tail (08-12 → 08-14) has the evidence and the missteps.
This supersedes `HANDOFF_2026-08-12_continue_here.md`.**
Re-run every liveness claim before acting on it — this tree moves, and two claims in the file
this supersedes had to be corrected within a day.

## The one-line state

**B1+B2+B3 are live, proven end to end, and the estate is prepared. B4 — the offer analyser
itself — is the next track and has not started. It is the whole of the next session's work.**

## What the owner decided (08-11 → 08-14)

1. **B4 next, not the A-track** (08-11 evening), and **B4 before affiliate** (08-14).
2. **Refresh the pre-B2 premises** (08-12) — done, 13 sites.
3. **Add the Q-fields to the two human-authored specs** (08-13) — done, with one field
   deliberately withheld (below).
4. **Affiliate: NOT yet.** *"I don't think we can claim affiliate yet unfortunately, please take
   B4 first."* Deferred, not scoped, not started. This **reverses** the 08-13 decision the day
   after it was taken; both entries stand in the PLAN log, read them together.

## Where the estate is (all measured 08-14)

- **Every one of the 22 sites has a current `strategy` spec carrying
  `revenue_models.primary_model`.** That was B3's whole point and it is done.
- **Q-fields (`satisfaction_condition`, `trust_threshold`, `recurring_value`) are on 22 of 22
  sites** — up from 7 on 08-12. **The only gap in the estate is ONE field on ONE site**:
  leopardessconsulting.co.uk has no `recurring_value`, by decision (below).
- **WII-014 is live and has fired** (v1.0.1291). Three `revenue_shape` capability gaps stand
  open and undispatchable, which is their correct resting state:
  `rule_missing` on vetcomparison.uk (`sponsored_listings`, no rule stated) and
  `handler_missing` on dartsonline.com, loancalculator.co.uk, loanandmortgagecalculator.co.uk
  (affiliate, no machinery). **Leave them open** — they are the standing record that four live
  sites are outside the offer checker's reach, and each retracts itself if its capability is built.
- `premise_incomplete`'s retraction arm has fired for real (08-12, LMC).

## What the next session should do

1. **B4 — the offer analyser. This is the work.** `features_open/030` §5.4 and
   `PLAN_2026-08-02` §B4 are the brief. Its inputs now exist on the whole estate.
   **⚠ Design it assuming its inputs are UNVERIFIED PROSE.** See item 2 — this is not a caveat,
   it is a design constraint discovered the hard way this week.
   B4 has a **named external consumer**: the `copy_quality_two_stage` lane needs a per-site
   ranked *"what is this reader trying to do, so what should this page say first?"* Read the
   reply before designing anything —
   `copy_quality_two_stage/CONTRIB_2026-08-12_the_ordering_input_you_want_is_already_in_site_specs.md`.
   Three of the four inputs it wants already exist and nothing reads them; what does **not**
   exist is any ORDERING, and anything per-PAGE. That is B4's centre.
   Two live acceptance fixtures are still waiting, neither composed by us: **gaswholesalers.com**
   (strategist classified the domain `generic_industry`, then chose `site_type: brochure` with a
   `money_flow` narrating a real gas-wholesale business — the shape its own prompt warns against)
   and **loanandmortgagecalculator.co.uk** (`affiliate` on a platform with no affiliate
   machinery). Neither is a bug; both are exactly the judgement B4 exists to make.
2. **⚠ NOTHING ON THIS ESTATE CLAIM-CHECKS A `site_specs` ROW, and B4 will grade against them.**
   The premise prose is written by an LLM and read by `site-review-agent` (B1) and, soon, B4.
   `check_unverified_claims` scans deployed HTML and stored `content_data`
   (`check_unverified_claims.go:1-36`) — **never `site_specs`** — and `evidence_base` /
   `banned_claims` do not cover specs either. **This is not hypothetical:** a donor run on
   2026-08-13 produced a `recurring_value` asserting a twice-weekly technical blog on three named
   topics, for a site with 6 posts in ~4 months on entirely different subjects. It was caught by
   reading it, and by one query against `pages`. **The 13 records refreshed on 08-12 have never
   been claim-checked** (a 3-site eyeball found nothing of that class — a sample, not a check).
   `bugs_open/161`'s shape, one layer up: a false fact in the premise causes a claim and then
   vouches for it.
3. **ONE DECISION IS OWED BY THE OWNER, and it was left open deliberately.** He asked to *"leave
   the false one in and trigger the improvement loop to fix it naturally"*. **That is not
   available, on two counts, both read in the code** — the audit cannot see `site_specs` (item 2),
   and it never repairs anything anyway (*"Truth decisions are human — auditors raise work items,
   they never rewrite content"*, `:39-41` and `:140`). Options, his call:
   **(a)** leave `recurring_value` omitted — today's state, costs one field on one site;
   **(b)** extend the claims audit to cover `site_specs` prose — the real gap, no existing bug
   covers it, and the only option that makes "let the platform handle it" true;
   **(c)** merge it knowingly. **Avoid (c)** — a known falsehood in a record B4 grades against,
   with no detector anywhere.
4. **A falsifiable prediction left open by the 08-12 refresh, cheap to check.**
   dartsonline.com's premise changed `direct_business` → `affiliate` (the only one of 13 that
   moved; the reasoning is sound — no stock, no warehouse, ships nothing). So its next
   examination should file `capability_gap:revenue_shape`, `gap_kind=handler_missing`,
   `deferred`, empty handler. **It already has**, filed 08-12. Nothing further owed unless it
   files something *else*, which would mean the switch is not reading what we think it reads.
5. **`bugs_open/255` remains open** — the `missing_conversion_path` spec carries neither
   `description` nor `category`, the two fields `content-gap-planner` reads, so that item type
   can never be handled by the agent it is routed at. Candidate 3 (give the spec what the handler
   reads) is the only one that makes the owner's "let it plan, decide before it builds" answer
   meaningful. **It must not ship un-witnessed.** Cost of leaving it: one refused LLM call per
   7-day rotation — trivial, so this is not urgent; it is filed because it is a CLASS defect
   (a handler named in a routing decision is a contract nothing checks).

## Watch-outs

- **⚠ A banned-term screen is NOT a claims check.** The regex built from leopardess's 2026-07-16
  ruling (`70+`, `N departments`, `managing agent`, least-privilege) passed prose containing a
  brand-new fabrication — no banned term, no "department", **no numerals at all**. A banned-term
  list records what was already caught; the next invention uses different words. **Invented
  specificity is the signal** — a frequency, a count, a named topic list, in a field nobody
  supplied source material for. Read it, then check the most checkable sentence against the
  database. Full entry in `LANDMINES.md`.
- **⚠ When importing generated prose into a protected record, prove the merge additive
  MECHANICALLY.** The pattern that worked twice: pin `md5(data)` before touching anything, then
  one atomic `DO` block whose guard is `md5(merged − the added keys) = the pinned md5`, with
  `RAISE EXCEPTION` on mismatch (a verify block of bare `SELECT`s cannot stop a `COMMIT`). That
  turns "I did not change your wording" from a promise into a refusal condition.
- **⚠ Do NOT re-run a donor run hoping for cleaner prose.** Same generator, same failure rate;
  cherry-picking runs until one clears the gate launders a fabrication into the record.
- **⚠ TWO schedules drive our checks.** `site_discovery_rotation` covers only the rotation
  driver (7-day period, `LIMIT 1`, fires every 3h — dormant until ~08-16). The improvement loop
  runs the same checks as a CHILD orchestration, is hand-fired by other sessions, does **not**
  stamp that table, and **triages and dispatches** what our checks file. So that table is not the
  meter for "when will my check next run", and our findings will be acted on, not inspected.
- **⚠ `status='complete'` cannot tell a RETRACTION from a repair** — `resolveWorkItems` writes
  the same value a handler writes. Read `result->>'resolved_by'`.
- **⚠ B3 IS NOT "observe-only".** `triage_detect_items_action.go:161-173` promotes every
  `detected` row on a site the improvement loop reaches — no type filter, no ownership filter.
  A finding cannot be parked. Design future checks to be right, not to be reviewed.
- **Remediation vehicle, proven ~18 times now:** oneshot rows in `scheduled_tasks`
  (`target_topic='system.agent.scheduled.requests'`, `input_data={domain,site_id}`,
  `fire_message=true`, no pre_query), **disabled immediately after firing**; picked up within
  ~20s. **Build `input_data` from a subquery against `sites`, never from a UUID you typed** — I
  pasted a wrong site_id from memory on 08-13 and caught it only because I checked before it
  fired. **Never `run_improvement_sweep_once.sh` for a read** — its triage promotes on every path.
- **⚠ A rotation stamp does not mean a site was examined** (SCH-025, owned by bugfix_230); and
  `orchestration_states` retention is ~24h, so that join cannot answer for last week. For older
  questions ask the WORK ITEMS: a check arm that files unconditionally and produced no row did
  not run.
- **`site_work_items.created_by` reads `'generic'` on our own rows** — `count(DISTINCT created_by)`
  is structurally blind here. BIZ-031's register entry is the only producer record.
- **Two claims from this lane were corrected within a day of being written** — a predicted
  outcome recorded in an evidence column (08-12), and a real measurement generalised past its
  axis (08-13, "12 of 13 kept the same revenue model **so the refresh is repeatable**" —
  classification stability is not prose accuracy). Both in `WRONG_CALLS.md`. Neither was an
  unmarked guess; both passed every marker discipline. **Before writing "so X is
  safe/proven/repeatable", name the property the number is about and check it is the same
  property as the claim.**
- B1 truncation watch-out unchanged; kafka-scheduler OOM of 08-09 (128Mi, exit 137) unchanged.

## Who owns what nearby

portfolio_positioning owns premise→writer wiring; brochure_component_library owns 016's
first-user relationship; bugfix_149 owns checker-layer plumbing; bugfix_230 owns SCH-025.
**`copy_quality_two_stage` + the loanandmortgagecalculator lane are actively working LMC** —
coordinate before firing anything at that site, and never a sweep while their controlled
round-3/round-4 copy pair is in flight. **dartsonline's own lane is being briefed separately by
the owner on affiliate partner recommendations** — that is content, not the platform capability,
which is deferred.
This lane owns: the drain, the critic, the recompose handler, anti-brochure compose-time work,
the offer analyser (B track), and **WII-014**.

**Also carried:** `bugs_open/198` (css-patch-agent) — both fix candidates live and pod-verified,
open only for a witnessed end-to-end run. And the fleet-wide round-trip-writer inventory at
`bugfix_198_roundtrip_writers/HANDOFF_2026-08-10_continue_here.md`.
