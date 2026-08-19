# 325 — `truncated_component` items are born `needs_human_review`, a status their own completion verifier is forbidden to leave — so no item of the type has ever closed, and bugs_closed/303's residual instruction pointed at a path that cannot fire

**Filed 2026-08-19** by the session dispatched at bugs_closed/303's residual ("complete the
two false-alarm items normally").
~~**Status: FIXED AT SOURCE 2026-08-19 (`c117c1bba`), OPEN until a chassis image rolls AND the
next daily sweep runs.**~~
**Status: CLOSED 2026-08-19 — fixed AND live on `v1.0.1315`, and PROVEN BY A LIVE SWEEP RUN:
all three parked items closed `auto:revalidated` at 16:00:48Z with the verifier's evidence
sentence recorded, components untouched.** Close-out at the bottom.
Council: **APPROVED round 1, all reviewers** (14 seats, 3 abstained,
0 unreadable), correlation `70fc8ff0-b50a-41fd-a358-a94decb269e0`. The commit carries
`Council-Submitted:` and is credited automatically by the 098 report (forward-only forbids
amending the trailer to `Council-Reviewed:`).

## The one-paragraph version

`check_truncated_component` files every item at `needs_human_review` with no handler — correct,
because the remedy (restore/regenerate/remove) is unsafe to automate. bugs_closed/303 registered
a completion verifier for the type (`RegisterVerifier("truncated_component",
VerifyTruncatedComponentResolved)`) and its close-out told operators to "complete them normally;
the verifier resolves them as balanced". But that verifier runs only inside
`CompleteWorkItemAction`, whose closing UPDATE carries
`status NOT IN ('needs_human_review', …)` — a deliberate guard so handlers cannot re-stamp a
deliberately parked item. **The type is therefore born in a status its own completion path is
forbidden to leave.** The review queue's actual drain — `revalidate_review_queue`
(bugs_open/033), daily, `dry_run:false` — judges only the types in its `reviewRevalidators`
map, which had no `truncated_component` entry. Two registries that look like one:
`RegisterVerifier` guards the completion door; `reviewRevalidators` drains the parked queue; a
type born parked is only ever reachable through the second.

## Evidence `[MEASURED 2026-08-19, live DB + source]`

- **Closer census** (the bar the earlier revalidator adopters set): all **3** items ever filed
  sit at `needs_human_review`; `handler_agent` empty on every row;
  `count(DISTINCT resolution_path) = 0`. Nothing has ever closed one.
  The three: `e7a4a7dd` (tool-llm-cost-calculator, filed 07-24, component repaired since),
  `91007600` (info-card-grid, filed 07-31, 303 false alarm), `6e2c9ebf`
  (gauntlet-round-record, filed 08-03, 303 false alarm).
- **The deciding guard**, read from source (`load_work_item_actions.go:1024-1025`):
  `WHERE id = $1 AND status NOT IN ('needs_human_review','failed','unresolved','rejected',
  'wont_fix','verified','blocked')`. Even a passing verifier ends at
  `completed:false, reason:"already_flagged_or_terminal"`.
- **The birth status**, read from source (`check_truncated_component.go:217`):
  `Status: "needs_human_review"` on every WorkItemSpec the check emits; `:73` registers the
  completion verifier.
- **The drain that existed had no entry**: `reviewRevalidators` (pre-fix) held six types, none
  of them this one. Its driver is real and live: `scheduled_tasks.review-queue-revalidate-daily`
  enabled, 86400s, last completed 2026-08-19 08:45Z; step config `{"dry_run": false,
  "max_items": 1500}` on agent `diagnosis-review-queue-revalidator`.
- **All three parked components balance under the live predicate**: their stored
  `html_template`s run through `content.UnbalancedStructuralTags` at current HEAD return empty
  (measured via a replace-directive harness over the live DB values); all three are
  `is_active=true`. So all three items resolve honestly the moment anything is allowed to judge
  them.
- **Historical corroboration**: `91007600.result` still carries the 2026-08-04 revalidation
  stamp `"no revalidator registered for item_type \"truncated_component\""` — the sweep looked
  at this exact row and said so, sixteen days before 303's close-out assumed the opposite path
  existed.

## Root cause

Not the guard (correct), not the verifier (correct), not the parked status (correct — the
remedy is genuinely human). The defect is the **unwired pairing**: a completion-time verifier
was added for a type whose every instance is born in the one status completion refuses, and no
queue-drain revalidator was added alongside it. The 303 close-out then documented the
completion path as the residual's remedy without reading the completion guard's status list.

## The fix (committed `c117c1bba`, inert until a chassis roll)

`revalidateTruncatedComponent` registered in `reviewRevalidators`
(`platform/orchestration/actions/revalidate_truncated_component.go`). It **delegates to the
same `checks.VerifyTruncatedComponentResolved` the completion gate runs**, so the two doors
cannot answer the same question differently. Verdict mapping is pinned by table tests:
verifier error → `unknown` (non-terminal, stays parked — RFC_017's fail-closed direction
preserved), `Resolved` → `resolved` (closes with the verifier's evidence sentence in
`result.revalidation`, `resolution_path='auto:revalidated'`), else `still_holds` (stays parked
for a human; remedy decision untouched). Retraction, not remedy: nothing is restored,
regenerated or removed. Coverage test updated with the closer census
(`TestRevalidatorCoverageIsDeliberate`).

## How to verify (post-roll)

1. Binary probe agent-chassis for the build stamp; confirm
   `git merge-base --is-ancestor c117c1bba <stamp sha>`. (Recipe with both controls: the end of
   bugs_closed/303 — must-miss control high-entropy, never all-zeros.)
2. After the next `review-queue-revalidate-daily` run (06:50-ish UTC daily; or dispatch the
   `diagnosis-review-queue-revalidator` agent), all three items read `status='complete'`,
   `resolution_path='auto:revalidated'`, `result.revalidation.verdict='resolved'` with
   `arm='verifier_resolved'` and the balanced/deactivated detail sentence.
3. Negative control: none of the three components' `html_template` rows were modified by any of
   this — `updated_at` on `content_components` unmoved by the sweep.

## Why this was NOT put through the `090` loop (owner ruling 2026-07-31 escape hatch, stated)

Substituted equivalent first-hand verification: the deciding guard, the birth status and both
registries are quoted above from source at exact lines; the closer census and the driver config
are live-DB measurements; the balance claim was exercised through the live predicate rather
than repeated from 303. A loop run would re-read the same four functions. The fix is
additionally before the council on the correlation above.

## CLOSED 2026-08-19 — fixed AND live, proven by a live sweep run through the normal path

**The roll:** image `v1.0.1315`, pods restarted ~12:16Z. Build commit **`590ca3a20`**
(2026-08-19 12:38 BST, the newest commit before pod start with none between), confirmed by
known-value probe in `/proc/1/exe` on BOTH replicas (`bfw5n`, `nkdkl`), high-entropy must-miss
control clean. Ancestry: `git merge-base --is-ancestor c117c1bba 590ca3a20` ✓.

**The proof run** (not waited for — the scheduled task was made due by setting
`last_triggered_at` back 25h, so the kafka-scheduler dispatched it through the NORMAL path at
16:00:47Z; pods were 3h43m old, well past the 300s dispatch blackout):

- Sweep `COMPLETED`: `{scanned: 201, resolved: 12, still_holds: 91, unknown: 98, closed: 12,
  dry_run: false}` — the 12 closes are these three plus nine other covered-type items, the
  sweep's normal daily work done a cycle early.
- **All three items of this type closed within one second of dispatch**, each with the full
  audit trail: `status='complete'`, `resolution_path='auto:revalidated'`,
  `result.revalidation = {verdict: resolved, arm: verifier_resolved, reason: "html_template
  balances every paired tag"}` (completed_at 16:00:48.98 / 49.13 / 49.44Z).
- **Negative control — retraction, not remedy:** all three components untouched. `updated_at`
  on `content_components` unchanged (2026-08-10 / 08-18 / 08-10, all pre-sweep), lengths
  unchanged, all still `is_active`. Nothing was restored, regenerated or removed; 91007600's
  dangerous "restore v1" remedy text was never acted on.
- **Honestly stated:** this run exercised only the `resolved` arm live. The `still_holds` and
  `unknown` (error) arms are pinned by the table tests at the shipped code, not yet naturally
  exercised — their first live exercise will be the next genuinely truncated component the
  discovery check files.

**Residual: none on this lane.** The queue for the type is empty; the next truncated_component
item the check files will park for a human as designed and drain automatically if the
component is repaired, deactivated or falsely flagged. The dedup keys for all three components
are released, so re-detection is unobstructed.

## Related

- `bugs_closed/303` — the counter fix whose close-out this corrects; addendum there points here.
- `bugs_open/033` — the review-queue-has-no-drain umbrella; this is that bug's shape, one type
  at a time, and the fix uses 033's own drain mechanism.
- `LANDMINES.md` "`RegisterVerifier` is not a drain…" (added 2026-08-19) — the prospective form
  of this trap, for the next session that registers a completion verifier.
- `WRONG_CALLS.md` 2026-08-19 — the close-out claim, what caught it, and the cheap check.
