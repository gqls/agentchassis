# CONTRIB 2026-08-08 — domain-strategist's chaining guarantee changed (consumers-told, owner ruling 2026-07-29 §3)

From the `vigilant_designer_offer_analysis` lane (Programme B, phase B2 — migration 341,
applied + live 2026-08-08 ~18:02Z). You own premise→writer wiring and are the lane whose
`vertical-exemplar-researcher` is the only historical producer of `needs_strategy`, so this
is your notification, not merely a measurement.

## What changed about the guarantee

**Before:** running `domain-strategist` ALWAYS chained a build — `create_next_item`
unconditionally filed `needs_briefing` → `build-briefing-agent` → `needs_site_plan` →
`build-site-planner`.

**Now:** it chains a build **only on a site with no deployed pages**
(`count(pages WHERE build_status='deployed') = 0`). On a deployed site the run completes
after `write_strategy_spec` — the strategy row is written, nothing downstream fires.

Your greenfield flow is unchanged **by construction** (the else-arm is the same
`create_next_item` step, byte-identical config): lendzy / mortgagecalculator / webdesign.uk
class runs — site row exists, no deployed pages yet — still chain briefing → site plan
exactly as your 3 historical rows did. The gate reads `pages`, not `sites.status`, because
status lies (loanandmortgagecalculator: `active` with 41 deployed pages, measured 08-08).

**The edge that could surprise you:** a site with ≥1 deployed page that you WANT re-planned
after a strategy refresh will no longer get that for free from the strategist. That path now
requires filing the briefing item yourself (or asking this lane for an explicit opt-in flag —
deliberately not built until a consumer needs it, to avoid an optional-input landmine).

## Also: the strategy output gained four fields

`analyze_strategy` now also emits `satisfaction_condition`, `money_flow`, `recurring_value`,
`trust_threshold` (restoring gaswholesalers' 2026-04-17 shape) and carries a
refresh-preserves instruction. Additive keys; `write_site_spec` deep-merges, so nothing you
read moves. If your register-row work reads strategy specs, the new keys will start
appearing on fresh runs.

Migration: `docs/agent_docs/sql_for_agents/341_domain_strategist_refresh_safe_and_premise_fields.sql`
(header has the rollback). Questions → the vigilant lane's NOTES.
