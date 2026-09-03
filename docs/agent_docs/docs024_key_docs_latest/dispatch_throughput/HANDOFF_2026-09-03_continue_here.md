# HANDOFF 2026-09-03 — dispatch throughput / D4 spend governor — CONTINUE HERE (supersedes HANDOFF_2026-09-02)

**Read first, in this order:** this file → `NOTES_dispatch_throughput.md` 2026-09-03 entries
(the enable, then the first post-enable watch) → `RUNBOOK_dispatch_throughput.md`
§"Spend governor (D4)" (new today — every governor query now lives there, not in prose) →
`SUMMARY_2026-09-03` (the arc read aloud) → `README_where_we_are.md` (owner's log, last two
entries). Paths relative to `docs/agent_docs/docs024_key_docs_latest/dispatch_throughput/`.

## The one-paragraph state (2026-09-03 ~10:50Z)

**D4 IS LIVE. Every precondition is met and nothing is owed by the owner.** The governor was
enabled 10:14:32Z on the owner's word; the budget is $2,000/month (L1 $1,400 · L2 $1,700 ·
L3 $1,900); the owner raised the Anthropic console cap to **$3,000** at ~10:41Z, which closes
the last dependency — the account's hard wall now sits above every governor threshold, so the
staged brake always fires first. State at close: **enabled, level 0, ~$386 MTD, heartbeat
live.** The first post-enable watch is done and clean: dispatch unchanged either side of the
switch (0 refusals, 0 withheld), wiring intact at the right object after this morning's
release, Go halves capability-probed on BOTH current pods with an absent control, and the
**shed staircase proven at all four levels without withholding any live work** (rolled-back
synthetic-level probe: 0 / 51 / 112 / 112 items withheld, class ladder in the owner's ruled
order). **The ONE thing still unproven is the Go loader/claim reading a NON-ZERO level on live
traffic** — which is also option C's gate.

## NEXT — one open question, then the queue

1. **THE OPEN QUESTION, put to the owner 2026-09-03 ~10:50Z and awaiting one word:** prove the
   last link by **inducing** a shed in a controlled ~5-minute window (drop `monthly_budget_usd`
   so MTD crosses 70%, e.g. 500; watch the 120 s task move to L1; watch `governor_withheld_now`
   fill with maintenance/llm rows; watch a loop refuse with `spend_governor_shed`; watch the
   level-change `doc_notes` row; restore 2000; watch it drain) — **or** wait for the real
   crossing, which at ~$124/day arrives ~11 September. Nothing is lost either way: withheld
   items stay `triaged`, burn no attempt, and resume on restore. **Do not induce without the
   owner's word** — it deliberately throttles the live fleet, briefly.
2. **After one real or induced shed is observed: option C unlocks** (trigger interval ≤25 s) —
   its own migration editing VERIFY 2/7's lever in lockstep, its own council round.
3. Then the standing queue, unchanged: deploy batching (D8 interim) · clients-first lane (D2) ·
   Batch API (D6) · D16 retention · per-class maintenance LLM cost (RESEARCH §6) · DNS plan B.

## Daily habits for this lane (now three)

- **584 VERIFY** (zombie NOTICEs benign, TWO reaper spellings; runtime is DB-load-bound).
- **657 VERIFY** — pins the selector md5, now `fcbe8821a2a56512911955735796460e`.
- **NEW: the governor wiring check after EVERY release**, not only after touching the governor
  (RUNBOOK §"Spend governor" — two queries). A release rewrites all 208 live
  `agent_definitions` rows in one statement ~70 s before the pods start (measured 08:56:53Z
  today, no `schema_migrations` row — it is the release's seeding step). The hand-applied
  governor clause survived it, but 674 edited the LIVE row and no repo seed carries the
  clause, so a re-seed that DID overwrite would silently remove the governor's primary gate.

## Traps (new today; the 09-02 + 08-30 + 08-26 handoffs' all still stand)

- **"The selector" names TWO live objects.** The governor clause and 657's pinned md5 live in
  `agent_definitions` (`find_dispatchable_site`); `scheduled_tasks.pre_query` is a DIFFERENT
  query (the wake-up gate, 415/688's lane). Reading the clause off `pre_query` returns a
  confident `false` and an unfamiliar md5 — which reads as *someone reverted your migration*.
- **The selector ends in `LIMIT 1`** — `count(*)` over it is 1 at every shed level, true or
  false. Strip it with an assertable arm. LANDMINES entry + WRONG_CALLS row filed today.
- **A single stale `computed_at` is not a dead heartbeat** — 120 s interval + 30 s tick makes
  ~150 s ordinary; one 211 s read self-corrected to 25 s. Two reads over ~300 s is the signal.
- **Yesterday's pod probe proves nothing after a roll** — re-probe the CURRENT pods, and run
  the absent control SEPARATELY (combined, the exec times out).
- **Handler counts across a window boundary are right-truncated** — do not grade throughput
  on them; loops spawn handlers for minutes after they start.

## Coordination state (unchanged from 09-02 unless noted)

All cross-lane threads closed (413, 414, 314/426, 384). 415's migration 688 may apply any
time — no lockstep owed. Ordering decisions ROUTE TO THIS LANE (owner 09-02); the provisional
no-reorder ruling stands, mechanical revisit trigger = the pin census age tail.

## D4 reference card

Unchanged from HANDOFF_2026-09-02 §"D4 reference card" — read it there (objects, register
AGOV-013, the STANDING GATE on a second `honour_spend_governor` consumer, rollbacks). The one
correction: the selector md5 is now `fcbe8821a2a56512911955735796460e` everywhere, and
`governor_config.enabled` is **true**, so 671's rollback will refuse until it is set false.
