# HANDOFF — vigilant designer + offer analyser (2026-08-12)

**COLD-START = this file + `PLAN_2026-08-02` (programme + owner decisions) + `features_open/030`.
NOTES tail (08-12 entries) has the evidence and the missteps. This supersedes
`HANDOFF_2026-08-11_continue_here.md`.** Re-run every liveness claim before acting on it — this
tree moves, and one claim in the file this supersedes was fiction (see "What changed" below).

## State

**Programme B: B1+B2+B3 LIVE and now EXERCISED end to end. WII-014 is live and has fired.
B4 has not started — it is the whole of the next session's work.**

All four items the 08-11 handoff left open are discharged. What is left is B4.

- **All three dispatched `needs_strategy` findings have premises**, and two of the three were
  repaired by the platform's own drain with no hand-holding: gaswholesalers.com
  (`direct_business`, 08-11 16:19), loancash.co.uk (`display_advertising`, 08-11 22:37 — it
  drained itself about four hours after the last handoff said to check it, so the "fleet
  backlog, not a wedge" reading held), loanandmortgagecalculator.co.uk (`affiliate`).
- **WII-014 is LIVE and VERIFIED AT THE ARTEFACT.** Fleet rolled to `v1.0.1291` at 14:55Z; both
  chassis replicas probed individually and built from `da5a7eb8f`, which has `0ceb27a40` as an
  ancestor. First live firing 17:15:23Z. Register entry updated to LIVE with the row quoted.
- **`premise_incomplete`'s retraction arm fired for the FIRST time on this estate** (17:18Z,
  LMC): `needs_strategy` → `complete` with `resolved_by=premise_incomplete` and a stated reason.
  Both halves of RFC_010's retraction design are now live-exercised, not just unit-tested.
- **The greenfield negative control is EXERCISED and can be retired from the watch-outs.**
  noted.co.uk was built greenfield overnight and filed `needs_briefing` (complete 08-12 02:22,
  `build-briefing-agent`) alongside `needs_strategy` → `saas_tools`. The gate did what B2 said
  it would.

## What changed that you must not inherit blindly

**The 08-11 handoff's stated positive control did not exist.** It said to verify WII-014 with
*"loancalculator.co.uk's existing `affiliate` row must still read `handler_missing`"*. There was
no such row and there never had been: `check_revenue_shape` had filed **exactly one work item in
its entire life**. That site had never been examined by the check — its only quality rotation
(08-09 10:50) predates migration 361 adding the check to the agent's array (~20:05Z the same
day). Which also means **"estate sweep COMPLETE — all 21 sites examined by both offer checks"
(NOTES, 08-10) is false**, by at least that site.

The row now exists, because firing the check there produced the predicted `handler_missing` in
one second. Recorded in `WRONG_CALLS.md` and corrected in place in the register. **The lesson is
narrower than "check your claims": the sweep table HAD a "verified how" column and it was filled
in — with the reason the arm would file, not with an observation that it had.** A justification
sitting in an evidence column reads as evidence.

## What the next session should do

1. **B4 — the offer analyser. This is the work.** `features_open/030` §5.4 and `PLAN_2026-08-02`
   §B4 are the brief, and B4 now has a **named external consumer**, which it did not have on
   08-11: the `copy_quality_two_stage` lane needs a per-site ranked "what this reader wants,
   most useful first" that a rewrite pass can read. Full reply, with what already exists and
   what does not:
   `copy_quality_two_stage/CONTRIB_2026-08-12_the_ordering_input_you_want_is_already_in_site_specs.md`.
   **Read that before designing anything** — three of the four inputs are already written on
   every site and nothing reads them.
   Two live acceptance fixtures still waiting, both from 08-11 dispatches and neither composed
   by us: gaswholesalers.com (strategist classified `generic_industry`, then chose
   `site_type: brochure` with a `money_flow` narrating a real gas-wholesale business — the shape
   its own prompt warns against) and loanandmortgagecalculator.co.uk (`affiliate` on a platform
   with no affiliate machinery). Neither is a bug; both are the judgement B4 exists to make.
2. **The cheapest thing on the whole list, and it is not B4.** Compare a page brief against its
   site's stored `value_proposition` before the writer sees it. LMC's brief led with the site
   inventory while its own `value_proposition` leads with the loan↔mortgage interaction — and
   the owner's rejection was, almost word for word, "lead with the second one". One comparison,
   no new artefact, and it would have caught a real owner rejection. Agree ownership with the
   copy lane first; it is the same check from two directions.
3. **`bugs_open/255` remains open and candidate 3 is still the only meaningful fix** (give the
   `missing_conversion_path` spec the `description`/`category` fields `content-gap-planner`
   reads). Unchanged from 08-11: it must not ship un-witnessed.
4. **Do NOT re-verify WII-014.** It is done, at the artefact, with both gap kinds
   distinguishable in one query. Re-running the check will pass again for the same correct
   reason.

## Watch-outs

- **⚠ NEW — TWO schedules drive our checks and this lane had only ever reasoned about one.**
  `site_discovery_rotation` stamps for `quality-discovery-agent` have not advanced since
  08-10 16:39, and that is **arithmetic, not a wedge**: the rotation task fires every 3h with
  `LIMIT 1` and selects on `last_selected_at < now() - interval '7 days'`; all 22 sites were
  stamped 08-09/08-10, so it correctly returns zero rows until **08-16 09:49** (robot-hands, the
  oldest stamp). Meanwhile **eight quality-discovery runs carrying the 9-check config completed
  in the last 24h** — every one a CHILD of an **improvement-loop** orchestration, hand-fired by
  other sessions, which does **not** stamp the rotation table.
  **So `site_discovery_rotation` is not the meter for "when will my check next run on site X".**
  It answers only for the rotation driver, and the other driver both runs our checks and
  **triages and dispatches** what they file. This is the 08-11 "B3 is not observe-only" watch-out
  with its second half filled in: the thing promoting our findings is fired by hand, by sessions
  who are not us, on no schedule we can read.
- **⚠ NEW — `status='complete'` cannot tell a RETRACTION from a repair.** `resolveWorkItems`
  (`work_items_common.go:287-301`) writes `status='complete'` for a positive-observation
  retraction, the same value a handler's successful drain writes. The discriminator is
  `result->>'resolved_by'` (plus `reason`). All eight `needs_strategy` closures before 08-12 read
  `complete` and **none** was a retraction — which is how we know the arm had never fired.
  Query the column; do not infer from the status.
- **⚠ B3 IS NOT "OBSERVE-ONLY"** (carried, unchanged, and now doubly true given the two drivers
  above). `triage_detect_items_action.go:161-173` promotes every `detected` row on a site the
  improvement loop reaches — no type filter, no ownership filter. A finding cannot be parked:
  demoting `triaged`→`detected` guarantees re-promotion, `cancelled` is terminal so dedup
  releases the key and the check re-files next rotation. Design future checks to be right, not
  to be reviewed. **Practical consequence proven today:** a stale `detected` finding of ours on
  LMC would have dispatched a redundant strategist run over another lane's fresh strategy row;
  retracting it is what stopped that.
- **Remediation vehicle, now proven five times:** oneshot envelopes in `scheduled_tasks`
  (`target_agent_type='quality-discovery-agent'`,
  `target_topic='system.agent.scheduled.requests'`, `input_data={domain,site_id}`,
  `fire_message=true`, no pre_query), **disabled immediately after firing**. Picked up within
  20s; the run completes in ~1s. **Fire a predicted positive** — a silent run is ambiguous, a
  run that files the predicted row proves scheduler and detector in one shot. Today's three all
  filed exactly what was predicted and **nothing dispatchable**: 9 checks ran on each site and
  the other 8 were silent, because those sites' findings are already open and dedup holds the
  keys. **Never `run_improvement_sweep_once.sh` for a read** — its `triage_findings` PROMOTES on
  every path.
- **⚠ RETIRED — the greenfield negative control watch-out.** Exercised by noted.co.uk on 08-12
  (see State). Do not re-carry it.
- **⚠ A rotation stamp does not mean a site was examined** (carried). The stamp COMMITS inside
  the pre_query, before the dispatch can fail, so a failure advances the rotation past an
  unexamined site for a full 7-day period. The daily watchdog compares **fleet totals** and
  cannot see it. Check: join `site_discovery_rotation` against `orchestration_states` per site.
  SCH-025's documented trade-off, owned by **bugfix_230** — their mechanism, their call.
  ⚠ **And `orchestration_states` retention is ~24h**, so that join can only answer for the last
  day. It cannot tell you who was examined last week — a limit this session hit and worked
  around by asking the WORK ITEMS instead (an unconditional arm that filed no row has not run).
- **A grep proves absence only for the spelling it searches** (carried, 08-11). Enumerating
  "who uses this seam" needs the caller grep, the struct-literal grep AND the
  variable-assignment grep.
- **`site_work_items.created_by` reads `'generic'` on our own rows** (carried). BIZ-031's
  register entry is the only producer record; `count(DISTINCT created_by)` is structurally blind.
- **The chassis `build provenance` startup line had scrolled out of `--tail=100000` on both
  replicas 80 minutes after the roll.** An empty result there means "not in range", never
  "unstamped" — fall back to the binary probe, with a must-be-absent control (a commit made
  after pod start) and remember a binary carries only its OWN build commit, not its ancestors.
- **Migration numbers 340/341/358/359/361 all resolve by SLUG** (carried).
- B1 truncation watch-out unchanged; kafka-scheduler OOM of 08-09 (128Mi, exit 137) unchanged.

## Who owns what nearby

portfolio_positioning owns premise→writer wiring; brochure_component_library owns 016's
first-user relationship; bugfix_149 owns checker-layer plumbing; bugfix_230 owns SCH-025.
**NEW: `copy_quality_two_stage` + the loanandmortgagecalculator lane are actively working LMC**
(they wrote its current strategy row 08-12 13:55 and are running a controlled round-3/round-4
copy pair on a live page) — coordinate before firing anything at that site, and never a sweep.
This lane owns: the drain, the critic, the recompose handler, anti-brochure compose-time work,
the offer analyser (B track), and **WII-014**.

**Also carried by this session:** `bugs_open/198` (css-patch-agent) — both fix candidates live
and pod-verified; open only for a witnessed end-to-end run. And the fleet-wide
round-trip-writer inventory, handed off at
`bugfix_198_roundtrip_writers/HANDOFF_2026-08-10_continue_here.md`.
