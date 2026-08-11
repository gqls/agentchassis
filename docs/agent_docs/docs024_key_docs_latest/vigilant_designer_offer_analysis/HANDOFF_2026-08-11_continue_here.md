# HANDOFF — vigilant designer + offer analyser (2026-08-11)

**COLD-START = this file + `PLAN_2026-08-02` (programme + owner decisions) +
`features_open/030`. NOTES tail (08-11 entry) has the mutation matrix and the missteps.
This supersedes `HANDOFF_2026-08-10_continue_here.md`.** Re-run every liveness claim before
acting on it — this tree moves.

## State

**Programme B (offer track): B1+B2+B3 all LIVE, observe-only, estate swept.** Unchanged from
08-10 except for the fix below. **The two decisions owed are the OWNER's, and they are the only
things gating what comes next** (see "Owed" below).

- **The four findings are still open, unrouted, all `detected`** (re-checked live 2026-08-11):
  `needs_strategy` × 3 (`strategy_loancash.co.uk`, `strategy_loanandmortgagecalculator.co.uk`,
  `strategy_gaswholesalers.com`) and `missing_conversion_path:62b5978e…`
  (mortgagecalculator.co.uk). Nothing new overnight, which is correct: the rotation is a 7-day
  cadence and the estate was fully stamped on 08-10.
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

## Owed: two owner decisions, and nothing else blocks

1. **The four findings are arguments, not orders.** The three `needs_strategy` items are
   directly actionable — one dispatch each, and B2's gate means a run against a live site no
   longer re-plans it. `missing_conversion_path` (mortgagecalculator has no contact form
   anywhere, on a `lead_generation` site) is a real site gap, not a checker bug: it needs a
   route — fix now, or roadmap.
2. **B4, or back to the A track.** B4 is the analyser itself. A-track (the vigilant designer)
   has been parked while B ran ahead. Owner's call, unchanged since 08-09.

## What the next session should do

1. **Read the rotation's next harvest** as the 7-day cadence brings sites round:
   `SELECT item_type, item_key, status, summary FROM site_work_items WHERE item_type IN
   ('needs_strategy','revenue_shape_cta','missing_conversion_path') ORDER BY created_at DESC;`
   **Verify per-site, never by stamps** — see the rotation landmine below.
2. **After the next chassis roll, verify WII-014 at the artefact** (its `verify-later` block has
   the query). Expect exactly ONE new row on vetcomparison.uk: `gap_kind` = `rule_missing`,
   `status` = `deferred`, `handler_agent` empty. **Positive control in the same query:**
   loancalculator.co.uk's existing `affiliate` row must still read `handler_missing` — it is
   what proves the query and the roll rather than your spelling. A `rule_missing` row that is
   anything but `deferred` means remit.go's double lock was relaxed and `bugs_open/077` is back.
3. Then whichever of the two decisions above has landed.

## Watch-outs

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
