# HANDOFF — bugs_open/210, cold-start here (2026-08-08, post-roll)

**State in one line: FIXED, council-APPROVED (round 1), LIVE on v1.0.1268, pod-verified —
~~in substance closed; what remains is a watch, one owner decision, and one optional canary~~
**UPDATED 2026-08-09: the owner decision is EXECUTED and the canary is CLOSED OUT (stood down,
not inducible — reasoning in the bug file + NOTES). Only the passive watch remains, and it is
still 0/0.** A scope boundary was found on the way out: the guard does not cover `page-rerender`
(LANDMINES + PBP-038).**

Cold-start reading order: this file → `bugs_open/210_…md` (state block at top) →
`SUMMARY_2026-08-08_…md` → `NOTES_…md` (newest at the bottom; the council-objection
dispositions and the LIVE verification evidence live there). Register entry: **PBP-038**
(`docs026_concept_register/register/page-build-pipeline.md`).

## What is done (do not redo)

| | |
|---|---|
| Fix | `2c3efc9f5` — guard widened in `UpdatePageStatusAction`; `page_build_failure_guard.go` (refusal → error-log row → needs_rebuild flip → 3-strike park on `needs_page:<name>` slot → auto-close on success); `loadOpenPageItems` +type, +cancelled-closed |
| Post-verdict extension | `8d95779a2` — park spec carries `bug: bugs_open/210` (mistyped_deployed_page drain convention) |
| Council | corr `c9647117-3a4b-48a2-b34c-1ea25f4e1f7f`, APPROVED round 1, 4 advisory objections — ALL TEN dispositioned with evidence (NOTES 2026-08-08 verdict section); commits carry `Council-Submitted:`, 098 credits automatically |
| Tests | 3 new tests, each mutation-proven (a killed guard fails its named test) |
| Live proof | v1.0.1268: 12/12 chassis containers one digest; greped replicas `1 1 0 3` (added ×2 / fabricated 0 / positive 3) |
| Registered | PBP-038 + index row (status LIVE); PBP-036's stale scope line struck through |
| Landmines | 2 entries (raw-insert-or-stillborn; parked-slot-blocks-your-insert), synced to doc_notes |
| Consumers told | mortgagecalculator lane, feature_021 lane, 208 handoff pointer |
| Missteps | WRONG_CALLS ×2: three schema-first skips in one session; the "8 vs 7" miscount |

## Still open, in priority order

1. **The watch (this is the whole remaining substance).** The bug's frequency was
   unmeasurable pre-fix (the file's own proxy was confounded). Now:
   `SELECT count(*), max(occurred_at) FROM agent_error_log WHERE error_code='DEPLOY_STAMP_REFUSED_ON_SKIP';`
   Baseline 0 at roll time (2026-08-08 ~18:00 UTC). The first non-zero IS the first real
   measurement — record it in the bug file when it appears. Parks:
   `SELECT * FROM site_work_items WHERE item_type='page_build_failed';` (also 0 at baseline).
   The architecture seat asked for park-inflow vs queue-drain sizing "shortly after rollout" —
   these two queries are that sizing.
2. **~~Owner decision, time-bounded~~ DONE 2026-08-09: mute stands, all 7 re-marked
   `wont_fix` with provenance, audit re-run clean (NOTES). Original item:** The `cancelled`
   alignment releases the 07-20 mute on brands, brands-index, grip-styles, guides,
   product-detail, shop, shop-index — they re-emit LLM builds when dartsonline is next
   REPLANNED (reconcile runs inside build-site-planner, demand-driven — so nothing happens
   until someone replans that site). If the mute should stand, re-mark those 7 items
   `wont_fix` (durable under both old and new rules). Audit query: RUNBOOK § cancel-as-mute.
   Also released, no cost: `vonc.com/provocation` (owned → one review item, no build);
   3 synthetic verify-keys (match no plan page); 14 deployed-and-current pages (skip_built).
3. **~~Optional behavioural canary~~ CLOSED OUT 2026-08-09 — authorised by the owner, then
   STOOD DOWN as not inducible. Do not re-plan it without reading why** (bug file state block;
   NOTES § "the canary was AUTHORISED and then STOOD DOWN"): all three routes into
   `assemble_page` are the same `check_review_approved` conditional, so both of the guard's
   triggers sit downstream of an LLM review gate — an induced content failure diverts there and
   never reaches `update_page_status`. Reaching this arm needs the reviewer to APPROVE a failed
   payload. A throwaway workflow row to force it was rejected (scratch config on a shared fleet,
   and no longer the real workflow). Residual gap, stated narrowly: no production run has been
   observed on the non-owned arm; 208's canary already proved the skip flag reaches this entry.
4. **The class sweep the 208 lane left** ("what else has a DB-row guard behind a git
   commit?") — not this bug, still unclaimed, see 208's handoff item 3.

## Landmines a fixing/refactoring session must not step on

- **Never route `parkPageBuildFailure` through `insertWorkItem`** — two-strike brands it
  `unresolved` at birth (terminal → holds no slot → bounds nothing). LANDMINES entry exists;
  the guidelines seat's DELETE+INSERT objection is answered by induction (NOTES).
- **`loadOpenPageItems` keeps `unresolved` BLOCKING deliberately** — "aligning" it with
  `workItemTerminalStatuses` re-opens the unbounded reconciler loop this fix closed.
- A silently no-oping emitter on a `needs_page:<name>` key may be a PARK, not dedup — check
  `item_type='page_build_failed' AND status='needs_human_review'` before diagnosing.
