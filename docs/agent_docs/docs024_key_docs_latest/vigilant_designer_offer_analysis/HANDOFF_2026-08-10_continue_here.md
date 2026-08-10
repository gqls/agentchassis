# HANDOFF — vigilant designer + offer analyser (2026-08-10)

**COLD-START = this file + `PLAN_2026-08-02` (programme + owner decisions) +
`features_open/030`. NOTES tail (08-10 entries) has the missteps and the evidence.
This supersedes `HANDOFF_2026-08-09b_continue_here.md`.** Re-run every liveness claim
before acting on it — this tree moves.

## State

**Programme B (offer track): B1+B2+B3 all LIVE. B3 has now swept the entire estate,
observe-only.**

- **21 active/deployed sites examined by both offer checks.** Four findings, all
  hand-verified truthful; every silence spot-checked (see the table in the NOTES 08-10
  late-afternoon entry). Zero `checks_failed`, zero `checks_unregistered`, every run.
- **Open findings, all born `detected`, nothing dispatched:**
  - `needs_strategy` × 3 — `strategy_loancash.co.uk`,
    `strategy_loanandmortgagecalculator.co.uk` (both: zero current strategy rows),
    `strategy_gaswholesalers.com` (pre-2026-05 shape, no `revenue_models`).
  - `missing_conversion_path:62b5978e…` — mortgagecalculator.co.uk. Verified: recorded
    `lead_generation`; `contact-index` is `planned` (never shipped) so the shipped
    contactish candidate fell back to `index` (landing), which carries no form; the only
    `<form>`s on the site are calculator inputs on tool pages.
- **`revenue_shape`'s CTA arm has never fired, and that is correct.** Word-bounded grep of
  all 12 lexicon phrases across every shipped component of all three `saas_tools` sites:
  ONE hit fleet-wide, and it is prose (`learn-operations-browser-storage`, webdesign.co.uk:
  "If you start a project on your laptop…"). The anchor/button-text-only decision is
  vindicated — a whole-HTML matcher would have opened with a false positive.
- **A silence that carries NO information:** vetcomparison.uk is `sponsored_listings`, and
  the model switch's `default` arm (`check_revenue_shape.go:242-245`) states no rule for it.
  Silent by design, not by cleanliness. Do not read that site's quiet as a clean bill.
- **Estate model census (2026-08-10):** direct_business 10, none 4, saas_tools 3,
  display_advertising 2, lead_generation 1, affiliate 1, sponsored_listings 1.
- **noted.co.uk** appeared at 16:10Z 08-10 (another lane, greenfield, 0 shipped pages).
  Not a miss: the rotation had not reached it, and the shipped-only predicate would keep
  `premise_incomplete` silent anyway. Expect its first stamp on the next rotation tick.

## OWED, and it is the only thing blocking the council: three owner commands

**Re-checked 2026-08-10 15:45Z — `schema_migrations` still returns 0 rows for
`^(358|359|361)`.** Council round 2 gates on exactly this; resubmitting before it lands
draws the same objection. The three single-line commands are verbatim in
`HANDOFF_2026-08-09b_continue_here.md` § "OWED" (paste-wrap trap applies — one line each).

Once they land, **round 3 on the SAME correlation** (`RESUBMIT_CORR=5cd586c9-c787-417a-a102-27fbddc48687`),
per the round-3 checklist in 09b: quote round 1's verdict verbatim (prior_art asked);
attach code_checks for the file:line claims; mark shipped edits as HISTORICAL RECORD, not
forward edits; cite CLAUDE.md's "OWNER RULING 2026-07-29" by path for review-after-the-fact;
answer guardian's retraction-collateral point (retraction on premise-COMPLETE is correct for
both `needs_strategy` producers — BIZ-031 records the shared-key convergence, RFC_010 §1
working as designed); note the rollback files now exist.

## Next session, in order

1. **Check the ledger rows** (above). If present → council round 3.
2. **The owner's decision on the four findings** — they are arguments, not orders. The three
   `needs_strategy` items are directly actionable (each writes a site's first recorded
   premise, and B2's gate means a run against a live site no longer re-plans it).
   `missing_conversion_path` needs a route: fix or roadmap.
3. **Read what the rotation files** as the 7-day cadence brings sites round again:
   `SELECT item_type, item_key, status, summary FROM site_work_items WHERE item_type IN
   ('needs_strategy','revenue_shape_cta','missing_conversion_path') ORDER BY created_at DESC;`
   **Verify per-site, never by stamps** — see the landmine below.
4. **Then B4** — the analyser itself, or A-track. Owner's call, unchanged.

## Watch-outs (carried + new)

- **⚠ NEW LANDMINE — a rotation stamp does not mean a site was examined, and the daily
  staleness watchdog cannot tell you either.** The stamp COMMITS inside the pre_query,
  before the dispatch can fail (`cmd/scheduler/main.go`: `runPreQuery`:427 → `fireTrigger`:278
  → `stampCompleted`:287), so any failure advances the rotation past an unexamined site for
  a full 7-day period. Measured 08-09/10: **5 of 12 quality stamps produced no run**, and
  the watchdog reported clean at 06:35Z because it compares **fleet totals** (21 stamps vs
  24 orchestrations) — and our own oneshot re-fires were 3 of the 24 propping it up.
  **The check: join `site_discovery_rotation` against `orchestration_states` per site**
  (domain match, `created_at` between stamp − 2 min and stamp + 30 min). Full evidence in
  `bugfix_230_discovery_driver/CONTRIB_2026-08-10_…` — **their mechanism, their call**; the
  stamp-before-dispatch trade-off itself is SCH-025's deliberate, documented choice
  (`bugs_open/048`'s starvation shape) and this lane initially mis-called it a gap
  (corrected in place; WRONG_CALLS logged).
- **Remediation vehicle, proven twice now:** oneshot envelopes in `scheduled_tasks`
  (`target_agent_type='quality-discovery-agent'`, `target_topic='system.agent.scheduled.requests'`,
  `input_data` = `{domain, site_id}`, `fire_message=true`, no pre_query), **disabled
  immediately after firing**. They fire within ~1 tick and leave rotation stamps untouched.
  **Never use `run_improvement_sweep_once.sh` for an observe-only read** — its
  `triage_findings` PROMOTES on every path (its own header says so).
  **Health-probe trick worth reusing:** fire a *predicted positive* first — a silent run is
  ambiguous (broken or clean), a run that files the predicted item proves scheduler and
  detector in one shot.
- **`site_work_items.created_by` reads `'generic'` on our own rows** (agentType-fallback
  landmine). BIZ-031's register entry is the only producer record; `count(DISTINCT created_by)`
  is structurally blind here. Discriminate in the spec if producer-splitting is ever needed.
- **Greenfield negative control [STILL NOT EXERCISED]** — no greenfield strategist run since
  B2's gate. The next real greenfield build must file `needs_briefing`. noted.co.uk may be
  the occasion.
- **Migration numbers 340/341/358/359/361 all resolve by SLUG** (340/341 ambiguous with
  bugfix_220's; 356 absent from the ledger too — another lane's business).
- **Pre-existing package test failure, NOT ours:** `TestEveryCheckProducedItemTypeIsClassified`
  fails on `decision_regression` (RFC_015 lane, `e1628f7df`) at clean HEAD.
- B1 truncation watch-out unchanged (capped context; `stop_reason=max_tokens` in
  `error_message`, `output_tokens` NULL on cut first attempts).
- kafka-scheduler: OOM incident of 08-09 evening (128Mi limit, exit 137) — its 090 row now
  reads `complete`; rolled to v1.0.1280, single pod healthy at 16:00Z 08-10. Limit unchanged,
  so treat a repeat as plausible.

## Who owns what nearby

portfolio_positioning owns premise→writer wiring; brochure_component_library owns 016's
first-user relationship; bugfix_149 owns checker-layer plumbing; **bugfix_230 owns SCH-025**
(the rotation + watchdog — contributed to on 08-10, see above). This lane owns: the drain,
the critic, the recompose handler, anti-brochure compose-time work, and the offer analyser
(B track, B1+B2+B3 live).

**Also carried by this session:** `bugs_open/198` (css-patch-agent) — both fix candidates
live and pod-verified; open only for a witnessed end-to-end run, which is this lane's next
css-patch dispatch. And the fleet-wide round-trip-writer inventory, handed off separately
at `bugfix_198_roundtrip_writers/HANDOFF_2026-08-10_continue_here.md`.
