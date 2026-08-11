# HANDOFF — vigilant designer + offer analyser (2026-08-11)

**COLD-START = this file + `PLAN_2026-08-02` (programme + owner decisions) +
`features_open/030`. NOTES tail (08-11 entry) has the mutation matrix and the missteps.
This supersedes `HANDOFF_2026-08-10_continue_here.md`.** Re-run every liveness claim before
acting on it — this tree moves.

## State

**Programme B (offer track): B1+B2+B3 all LIVE, observe-only, estate swept.** Unchanged from
08-10 except for the fix below. **The two decisions owed are the OWNER's, and they are the only
things gating what comes next** (see "Owed" below).

- **The four findings are ROUTED (owner decided 2026-08-11 evening: dispatch all three
  strategies; roadmap the conversion path; B4 next). All three premises now EXIST**, verified at
  the artefact — `site_specs` current strategy row carrying `revenue_models.primary_model`:
  - `gaswholesalers.com` → **`direct_business`**, written 16:19:04Z, work item `complete`.
    **Done by the platform's own drain, not by hand** — the first B3 finding to travel detection →
    repair autonomously.
  - `loanandmortgagecalculator.co.uk` → **`affiliate`**, written 18:33:23Z by this lane's oneshot
    (`oneshot-domain-strategist-lamc-20260811`, fired 18:32:22Z, **disabled immediately after**).
    Its work item is still `detected` — the check's retraction arm closes it on the next rotation,
    which is the design. ⚠ **`affiliate` means the next rotation files a `capability_gap`** (no
    affiliate machinery on this platform — the loancalculator.co.uk outcome). Repairing a premise
    converts one finding into another; do not read that new row as a regression.
  - `loancash.co.uk` → **still owed a premise.** Its item is `triaged` and unclaimed since
    16:34Z, and that is **fleet backlog, not a wedge** (410 `page-rerender`, 110 `asset-deployer`
    also triaged-unclaimed, oldest 12:52Z, while completions landed at 18:2xZ). Left queued
    deliberately: a oneshot would race the claim and write a second superseding row.
    **First thing to check next session** — if it is still unclaimed, that is worth a look.
  - `missing_conversion_path` (mortgagecalculator.co.uk) → **`wont_fix` at 19:01Z, and that
    exposed a real defect of OURS: `bugs_open/255`.** The platform promoted it at 17:43Z (after
    the owner chose to roadmap it — see the observe-only watch-out below), `content-gap-planner`
    claimed it, and refused: *"The content gap description and original category are both blank.
    There is no gap to evaluate."* **Our spec carries neither field**, so this item type can never
    be handled by the agent it is routed at — and `wont_fix` is terminal, so `idx_swi_dedup`
    releases the slot and the detector re-files next rotation. **Diagnosis loop CONFIRMED it first
    iteration** (`64e5ab04`), and read `idx_swi_dedup`'s live predicate to prove the release.
    Fix candidates are in the bug file, ordered; **candidate 3 (give the spec what the handler
    reads) is the only one that makes the owner's "let it plan, decide before it builds" answer
    meaningful, and it must not ship un-witnessed.** Also filed as `016b` §9.
- **The `default:` arm objection is CLOSED — code committed, council APPROVED round 1.**
  Commit `0ceb27a40`, corr `a46ff9a6-fcba-4ab4-a53d-130aae39f24b`, registered **WII-014**.
  `check_revenue_shape` no longer returns an empty `CheckResult` for a revenue model it has no
  branch for; it files ONE undispatchable `capability_gap` of the new kind `GapRuleMissing`.
  **Go is INERT until the next chassis roll** — it is not in any shipped image yet.
- **Council round 3 on the B3 correlation (`5cd586c9…`) remains REVISE and should stay that
  way.** Its gating objection is unwinnable by construction (the seat cannot see
  `schema_migrations`; full write-up in `fixloop_eg_dartsonline/RUNBOOK_council_gate.md`).
  **Do not fire round 4 to argue it.** The substantive objections from that round are now all
  discharged or answered — `bug_historian` by the commit above, `debug_historian` and
  `prior_art` checked and recorded 08-10, `guardian`'s edit-5 mislabel and `constitution`'s
  style point conceded and not worth a round on their own.

## ~~Owed: two owner decisions~~ — BOTH TAKEN 2026-08-11 evening

1. **The four findings: routed.** All three strategies dispatched (state above);
   `missing_conversion_path` to be roadmapped — **but the platform has already promoted it, and
   there is no state that holds it** (watch-out below). One question still with the owner: let it
   run, or cancel and accept re-filing.
2. **Next track: B4 — the offer analyser itself.** Not A-track. B1–B3 are live, the estate is
   swept, and the inputs the analyser needs now exist on every deployed site.

## What the next session should do

1. **B4 — start here.** `features_open/030` §5.4 and `PLAN_2026-08-02` §B4 are the brief.
   **Two live cases are already waiting to be the first test of its judgement**, and both came
   out of today's dispatches: gaswholesalers.com, where the strategist classified the domain
   `generic_industry` and then chose `site_type: brochure` with a `money_flow` narrating an
   actual gas-wholesale business — the shape its own prompt warns against; and
   loanandmortgagecalculator.co.uk, now `affiliate` on a platform with no affiliate machinery.
   **Neither is a bug. Both are exactly the judgement B4 exists to make**, which makes them the
   cheapest possible acceptance fixtures — real, recent, and not composed by us (a fixture we
   write to exercise a rule will exercise it; these did not come from us).
2. **Check `loancash.co.uk` first** (one query): if its `triaged` item is still unclaimed hours
   later, the queue explanation has expired and it wants a look.
3. **After the next chassis roll, verify WII-014 at the artefact** (its `verify-later` block has
   the query). Expect exactly ONE new row on vetcomparison.uk: `gap_kind` = `rule_missing`,
   `status` = `deferred`, `handler_agent` empty. **Positive control in the same query:**
   loancalculator.co.uk's existing `affiliate` row must still read `handler_missing` — it is
   what proves the query and the roll rather than your spelling. A `rule_missing` row that is
   anything but `deferred` means remit.go's double lock was relaxed and `bugs_open/077` is back.
4. **Watch the retractions.** Two premises now exist whose work items are NOT closed
   (loanandmortgagecalculator `detected`, loancash pending). The next rotation should retract
   them by positive observation — the first live exercise of `premise_incomplete`'s retraction
   arm on this estate. If a premise exists and the item does not close, that is a real bug.

## Watch-outs

- **⚠ NEW, and it retires a phrase this lane has used in every doc since 08-09: B3 IS NOT
  "OBSERVE-ONLY", and never was.** "Items born `detected`, nothing dispatched" was true only
  while the improvement loop had not reached those sites. Its promoter's SQL
  (`triage_detect_items_action.go:161-173`) is `UPDATE site_work_items SET status='triaged' …
  WHERE site_id = $1 AND status = 'detected'` — **no type filter, no ownership filter**, and the
  file's own header says so. Consequences, all measured today:
  - **A finding cannot be parked.** Demoting `triaged` → `detected` guarantees re-promotion (that
    predicate is what it selects on); `cancelled` is terminal, so dedup releases the key and the
    check re-files on the next 7-day rotation. Only two honest options: let it run, or cancel and
    accept it returning.
  - **Any future check this lane ships will have its findings DISPATCHED**, not inspected first.
    Design them to be right, not to be reviewed. This is
    `a-complete-work-item-is-not-a-repaired-artefact` one level up: we were reasoning about what
    our code does, not about what the estate does with its output.
  - **"Wait for the loop" is not a bounded wait per site.** loanandmortgagecalculator.co.uk holds
    9 rows still `detected` from 08-10 03:20–04:26 — given the unfiltered predicate, its triage
    step has not run in ~38h. Hand-fire when timing matters; do not assume a sweep is due.
- **⚠ RETIRED — the package-test watch-out carried by the last two handoffs is STALE.**
  `TestEveryCheckProducedItemTypeIsClassified` **passes** at this tree (75 check-produced item
  types across 106 files); another lane fixed `decision_regression` between 08-10 and 08-11.
  Do not re-carry it. `go test ./platform/orchestration/actions/... -count=1` is green.
- **⚠ A rotation stamp does not mean a site was examined** (carried, unchanged). The stamp
  COMMITS inside the pre_query, before the dispatch can fail (`cmd/scheduler/main.go`
  `runPreQuery`:427 → `fireTrigger`:278 → `stampCompleted`:287), so a failure advances the
  rotation past an unexamined site for a full 7-day period. The daily staleness watchdog
  compares **fleet totals** and cannot see it (reported clean on 08-10 while 5 of 12 stamps had
  produced no run). **The check: join `site_discovery_rotation` against `orchestration_states`
  per site.** This is SCH-025's deliberate documented trade-off, owned by **bugfix_230** — their
  mechanism, their call; this lane contributed the evidence and initially mis-called it a gap.
- **Remediation vehicle, proven twice:** oneshot envelopes in `scheduled_tasks`
  (`target_agent_type='quality-discovery-agent'`, `target_topic='system.agent.scheduled.requests'`,
  `input_data` = `{domain, site_id}`, `fire_message=true`, no pre_query), **disabled immediately
  after firing**. Fire a *predicted positive* first: a silent run is ambiguous, a run that files
  the predicted item proves scheduler and detector in one shot. **Never use
  `run_improvement_sweep_once.sh` for an observe-only read** — its `triage_findings` PROMOTES on
  every path.
- **A grep proves absence only for the spelling it searches** — the 08-11 correction. Enumerating
  "who uses this seam" needs the CALLER grep, the struct-literal grep AND the variable-assignment
  grep; the constant-name grep alone silently answers a narrower question. It put a wrong number
  into a council submission and into a register entry the same day.
- **`site_work_items.created_by` reads `'generic'` on our own rows** (agentType-fallback
  landmine). BIZ-031's register entry is the only producer record; `count(DISTINCT created_by)`
  is structurally blind here.
- **Greenfield negative control [STILL NOT EXERCISED]** — no greenfield strategist run since B2's
  gate. The next real greenfield build must file `needs_briefing`. noted.co.uk may be the occasion.
- **Migration numbers 340/341/358/359/361 all resolve by SLUG** (340/341 ambiguous with
  bugfix_220's; 356 absent from the ledger — another lane's business).
- B1 truncation watch-out unchanged (capped context; `stop_reason=max_tokens` in `error_message`,
  `output_tokens` NULL on cut first attempts).
- kafka-scheduler OOM of 08-09 (128Mi limit, exit 137): limit unchanged, so treat a repeat as
  plausible.

## Who owns what nearby

portfolio_positioning owns premise→writer wiring; brochure_component_library owns 016's
first-user relationship; bugfix_149 owns checker-layer plumbing; **bugfix_230 owns SCH-025** (the
rotation + watchdog). This lane owns: the drain, the critic, the recompose handler,
anti-brochure compose-time work, the offer analyser (B track), and now **WII-014**
(`GapRuleMissing` on the shared remit.go seam).

**Also carried by this session:** `bugs_open/198` (css-patch-agent) — both fix candidates live
and pod-verified; open only for a witnessed end-to-end run, which is this lane's next css-patch
dispatch. And the fleet-wide round-trip-writer inventory, handed off separately at
`bugfix_198_roundtrip_writers/HANDOFF_2026-08-10_continue_here.md`.
