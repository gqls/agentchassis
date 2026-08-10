# NOTES — bugfix 234 (append-only, newest at the bottom)

## 2026-08-10 — lane opened; premises re-verified; two owner decisions

Picked up `bugs_open/234` (filed 2026-08-09 by the bugfix_136 lane, which ended with "next
action is yours"; `who-owns.py` shows no owning workstream; no live transcript is working
it beyond citations).

**Bug re-verified live before planning** (the filing was hours old, but this tree changes
under you):
- All three carriers still present, correct values, none carries
  `spec_data`/`spec_paths`/`spec_literal` (recursive all-depths walk — see RUNBOOK).
- 16/16 `improvement_rerender_*` rows `spec='{}'` (2026-08-01 → 2026-08-09), positive
  control 5,040 non-empty rows fleet-wide.
- `dedupe_rerender%` and `capability_gap_audit%`: still 0 rows each — the other two
  carriers remain unexercised, so translating them changes no live behaviour.

**Two premises of the case file's owner-decision framing are stale, and both moved the
decision** (recorded in the case file as a dated correction):
1. The case file deferred restoring the flag pending "whatever guard 226 lands". **226's
   guard is LIVE**: both chassis replicas (v1.0.1274 at check time) carry
   `emitChromeDivergenceItem` (strings-proven, negative control clean), and 226's header
   says it has already refereed a real event (the dartsonline header fixer-vs-rebuild loop).
2. "Turning on full site-component reassembly that has not run from this path in months"
   is true of THIS PATH only. Fleet-wide the flag is routine: 8 producers file
   `refresh_site_components: true`, ~5–15 rows/day, latest same-day. Restoring the
   improvement-loop path adds ~1.8 rows/day to a daily behaviour, with the divergence
   guard watching.

**Owner decisions (AskUserQuestion at plan time):** RESTORE the flag via `spec_literal`;
ship BOTH `StrictConfig: true` on `create_work_item` AND the new `RemovedConfigKeys`
opt-in field.

**StrictConfig precondition measured**: after translating the three carriers,
`create_work_item` has ZERO unknown keys fleet-wide **at all depths** (recursive walk over
every live definition; every remaining key ∈ Required ∪ Optional ∪ ConfigKeys ∪
DeprecatedConfigKeys ∪ framework set). This is the "recognised set checked against every
live step" the action's own doc comment names as the gate for strict.

**Misstep (caught in-session, cost ~1 minute):** first read of the recursive census said
2 `spec` carriers, not 3 — I counted the GROUP BY rows instead of reading the count
column (improvement-loop's two steps collapse into one group row). Re-ran showing the
rows; 3 confirmed. The check that prevents it is in the RUNBOOK. Not a WRONG_CALLS entry:
never written down as a claim, refuted by my own next query.

**Prior-art checks for the framework half** (a quiet git log is not silence — searched
the tree, not just history): no `RemovedConfigKeys`/`RetiredConfigKeys` anywhere in Go;
migration 356 (another lane, same family: dead `commit_from`/`output_format` keys) shipped
CheckConfig opt-ins only, and explicitly left its two undeletable dead keys "to be
REPORTED by the detector" — i.e. that lane also had no mechanism for a key that must
hard-fail. Both lanes' leftovers become RemovedConfigKeys candidates once the field exists.

**Fable note:** the owner asked for fable to prepare the plan; the fable Plan agent died on
the session usage limit (reset 19:40). Plan prepared by the main (Opus) session instead,
owner approved it in plan mode.
