# 451 — the two-strike ladder parks a RECURRING detection as `unresolved` at birth, because a COMPLETED refresh counts as a strike — so a site's chrome stops refreshing after two refreshes in a week

Filed 2026-09-03 by the `site_delivery_and_editor` lane, found while asking why boxingonline.com's
analytics tag and consent banner had not arrived 26 hours after the chrome that carries them
was expected to rerender. Diagnosis loop: `needs_diagnosis` row `0639080d-a970-4e53-8067-07222256a455`
(filed 08:31:28Z, `awaiting_diagnosis`; find the run by payload). **Per the 2026-07-31 ruling:
this file asserts a cross-cutting root cause; the loop is running, and the substitute stated here
is first-hand — the deciding branch was read at the code, the three rows on the motivating site
were read with their timestamps, and the fleet census below uses the ladder's own summary prefix
as the discriminator, so it could have come out otherwise (rows without the prefix).**

## What the owner sees

boxingonline.com (`d2aa5206-73bc-4707-a69c-2702c1eb9152`): `analytics.gtm_container_id = GTM-PQ3WCTBD`
was written into `site_specs` (aspect `site_config`, `data->'analytics'`, `is_current=true`,
`updated_at` 2026-09-02 20:10:33Z — one row; the `sites` row carries no GTM id) by the analytics lane and the
consent-banner block (STY-060) into all three head templates at 20:55:43Z. Both reach a site
only when its chrome is re-rendered. As of 2026-09-03 08:3xZ the served pages and all three
`site_components` rows (updated 16:27:55Z on 09-02, before either write) carry neither.

## Root cause — one sentence

`insertWorkItem` (`platform/orchestration/actions/load_work_item_actions.go`, the two-strike
block around the `terminalCount >= 2` test) counts prior rows with the same `site_id` +
`item_key` in **`status IN ('complete','failed')`** over the last 7 days, and when it finds two
it writes the NEW row as **`status='unresolved'`** (terminal) with summary prefixed
`[unresolved after N attempts]` — so for a key that is *meant* to recur whenever its inputs
drift, a successful refresh is a strike, and the third drift in a week is parked at birth.

## Mechanism, read at the code

- Producer: `platform/orchestration/actions/discovery_checks/check_integrity.go` (~:415–435)
  files `needs_rerender` / `item_key: "stale_chrome"` / `HandlerAgent: "rerender-pages"` /
  `Status: "detected"` whenever a chrome slot's `render_inputs` fingerprint differs from the
  stored one. It is re-filed on every drift by design — "ONE item per site … sole producer:
  this check" (its own comment at ~:365).
- Ladder: `load_work_item_actions.go` — the block quoted above. Its comment reads *"Two strikes:
  handler had 2 chances and the issue persists"*. For `stale_chrome` the premise is false: a
  `complete` means the handler DID refresh the chrome; the next detection is a **new** drift
  (here: the GTM id being added), not the same issue persisting.
- Exemption exists but is unreachable from discovery: the block is guarded by
  `if item.itemKey != "" && !item.recurrenceExpected` (cited in `nav_rebuild_request.go:65–75`).
  `recurrenceExpected: true` is set by `nav_rebuild_request.go:199`, `prune_floor.go:391`,
  `emit_imagery_items_action.go:139`, `growth_posture_door.go:90`,
  `apply_gap_plan_action.go:798` — all on the internal `workItem` struct. **`WorkItemSpec`
  (`discovery_checks/registry.go:56–71`) has no such field**, so no discovery check can set it.
- Terminal: `unresolved` is in `workItemTerminalStatuses` (`work_items_common.go`) and in the
  exclusion list of `idx_swi_dedup`, so nothing retries it and the next detection (if the
  fingerprint changes again) is evaluated against the same two strikes until they age out of
  the 7-day window.

## Evidence `[MEASURED 2026-09-03 08:1x–08:3xZ]`

Motivating site, the three `stale_chrome` rows ever filed for it:

| id | created | status | note |
|---|---|---|---|
| `1b3c2afc` | 09-01 00:34:36Z | complete (attempt 2) | refreshed 20 pages — a SUCCESS |
| `b6c4eded` | 09-01 21:30:59Z | failed ×3 | `rebuild_blog_listing: duplicate key uq_page_components_no_byte_identical_duplicate` — a different bug |
| `5b4eb7a0` | 09-02 06:19:53Z | **unresolved at birth** | summary `[unresolved after 2 attempts] Site chrome is stale — … (footer)`; `attempt_count` 0, never triaged, `created_at = updated_at` |

Fleet, `item_key='stale_chrome'`, as of 2026-09-03 08:3xZ: **76 `unresolved` / 63 `complete` /
1 `failed`**. Of the 76, **75 carry the `[unresolved after` prefix, across 12 sites**; the one
without it is the pre-existing "needs investigation" sense. Ten most recent parked sites:
farmerinsurance.uk (09-03 06:10Z), finetuning.uk, boxingonline.com, gaswholesalers.com,
noted.co.uk, lendzy.co.uk ×2, loanzy.uk, vetcomparison.uk, vonc.com. Disconfirming result:
parked rows without the prefix, or fewer parked than completed — neither.

Consumers that route through exactly this item (told, not merely measured — see Related):
register **STY-060** ("reaches each site via its stale_chrome rebuild"); `analytics_gtm/HANDOFF_2026-08-25`
("C 15 spec-only awaiting their per-site `stale_chrome` rebuild"; "`stale_chrome` dispatches
(20 ever, all complete)" — true on 08-25, stale by 09-03); `bugs_open/397`.

## Why nothing caught it

`unresolved` reads fleet-wide as "needs investigation" (016b §9, 2026-07-26 entry, which
documents the two-strike parking for OUT-OF-REMIT handlers — where the label is at least
arguably right). This is the other case: the handler is fully in remit and succeeded, and the
label blames it anyway. The analytics lane's census read the 08-25 number ("20 ever, all
complete") as a property of the mechanism; it was a property of the first week, before any
site had accumulated two terminals. And a detector whose rows are terminal at birth still
*files* a row, so every "did the check run?" probe reads green.

## Fix candidates, ordered by what closes the door

1. **The ladder counts only `failed` as a strike.** A `complete` is not "the issue persists".
   Churn from a fast-recurring key is already bounded by the anti-churn deferral in the same
   block (`bugs_open/326`, `antiChurnWindowHours`). One-line predicate change; changes the
   rule for every keyed item type, so it is a shared-mechanism guarantee change → say so in the
   council submission and name the consumers (the 2026-07-29 ruling §3).
2. **Plumb `RecurrenceExpected` through `WorkItemSpec`** and set it on `stale_chrome` (and
   audit `improvement_rerender_*`, `reconcile_rerender`, `missing_structure`, which are the
   same recurring shape under `needs_rerender`). The designed lever, but the
   "producers must remember X" shape — every future recurring detector re-discovers this file.
   Do it as well as (1), not instead.
3. **Sweep the 75 parked rows** once (1) or (2) is live: re-file per site (a hand-filed row with
   `source='operator'`, `status='triaged'` bypasses the ladder — direct SQL never runs
   `insertWorkItem`), or re-triage in place. Until then each affected site needs the operator
   recipe below, and the analytics/consent rollout is stalled on all 12.
4. **Make the parked population visible**: a daily check listing `unresolved` rows whose
   summary carries the prefix AND whose two strikes include a `complete` — the honest count of
   "parked by success". `scripts/pattern-check.py` is not the place (it reads commits); a
   `cmd/config-key-audit`-style CronJob is.

**Operator interim, per site (what this lane did for boxingonline, 2026-09-03):** file the same
item by hand, copying the last COMPLETED row's spec —
```sql
INSERT INTO site_work_items (site_id, source, item_type, item_key, severity, summary, priority,
  handler_agent, status, created_by, approval_mode, pipeline, spec)
VALUES ('<site_id>', 'operator', 'needs_rerender', 'stale_chrome', 'medium',
  'Chrome refresh: <why> (sibling <parked id> was born unresolved by the two-strike ladder — bugs_open/451)',
  8, 'rerender-pages', 'triaged', '<your session>', 'auto', 'build',
  '{"reason":"render_inputs_drift","refresh_site_components":true,"slot_names":["footer"],"original_pipeline":"build"}'::jsonb)
RETURNING id;
```
The dedup index permits it (`unresolved` is excluded). Expect `rerender-pages` →
`render_site_components` → one `page_rerender … _assemble` item per page (reason-less ⇒
`rerender_single_page`, the stored arrays re-shipped byte for byte). Verify at
`site_components.updated_at` + `rendered_html LIKE '%<gtm id>%'`, then at the served pages
after the mirror tick.

## How to verify the fix

On a site with two `complete` `stale_chrome` rows inside 7 days, change a chrome render input
(any `site_config` key the fingerprint covers) and let the daily integrity check run: the new
row must be born `detected`/`triaged`, not `unresolved`, and must complete. Then re-run the
fleet census above: the prefixed-unresolved count must stop growing, and after the sweep in
candidate 3 it must fall.

## Related

- `bugs_open/326` — the anti-churn deferral in the same block (a failed build can never be
  retried because keys dedup in any status). Same function, adjacent defect.
- `bugs_open/397` — GTM backfilled into the stored head reverts on the next chrome render; this
  bug is why that next render never comes on 12 sites.
- `bugs_closed/091` — the second drift dropped by dedup and reported as raised (the dedup half;
  this is the ladder half).
- 016b §9, "A detector must PARTITION its population by the handler's remit" (2026-07-26) — the
  out-of-remit two-strike case; this file is its in-remit sibling.
- `bugs_open/149` — the discovery checker-layer queue.
- The rebuild_blog_listing duplicate-key failure in `b6c4eded` is a separate defect (an
  idempotency gap in the rerender-pages workflow) and is NOT filed here — grep before filing.

## Diagnosis loop verdict (read 08:49Z, 2026-09-03) — UNVERIFIABLE, stopped `scope-not-narrowing`; both stated gaps answered first-hand here

Row `0639080d` ran 08:40→08:47Z: **NOT CONFIRMED (stopped: scope-not-narrowing)**, no fix
proposed, "hand to a human with the full trail". Not a refutation — the loop cited
`check_integrity.go:(*StaleSiteComponentsCheck).Run` filing `needs_rerender`/`stale_chrome` and
then could not widen its bundle to the two things it named as still needed. Both are in this
file now, read at the code and the rows rather than asserted:

1. *"cannot verify check_integrity.go uses `WorkItemSpec` rather than building `workItem`
   directly"* — the check appends a `WorkItemSpec` to `result.WorkItems`
   (`check_integrity.go` ~:420–434), and the ONLY consumer,
   `platform/orchestration/actions/discovery_checks.go:238–255`, copies it into
   `insertWorkItem(ctx, tx, workItem{siteID, pageID, source, pipeline, itemType, severity,
   summary, spec, priority, handlerAgent, status, createdBy, itemKey, batchID}, …)` — **no
   `recurrenceExpected` in the literal**, so it is the zero value for every discovery check.
   The exemption is unreachable at the call site, not merely absent from the struct.
2. *"no `site_components`/`site_specs` rows for the site in the bundle"* — `site_components`
   for `d2aa5206`: three rows (head 41,421 B / header 2,271 / footer 2,289), all `updated_at`
   2026-09-02 16:27:55–56Z, none containing `GTM-PQ3WCTBD` or `cc_v1` `[MEASURED 2026-09-03
   08:35Z]`; `site_specs`: exactly one row for the site contains the container id (aspect
   `site_config`, `data->'analytics'->>'gtm_container_id'`, `updated_at` 20:10:33Z on 09-02),
   i.e. the render input moved AFTER the last chrome render and BEFORE the parked detection.

This is the fourth `UNVERIFIABLE`/`scope-not-narrowing` on a well-posed symptom in two days
(the components lane's three on `bugs_open/425` §2 are the others) — recorded as a signal
about the loop's reach on symptoms whose evidence spans a producer, a converter, a loader and
two tables, not as doubt about the mechanism.
