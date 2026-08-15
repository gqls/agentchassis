# PLAN — bugfix 277: the fleet-wide repair router for `required_fields_missing`

**Owner ruling 2026-08-15 (bugs_open/277):** "we should create a repair handler fleet wide."
This lane builds it, as the next concrete increment of bugs_open/033's owner-ruled reframing
(2026-07-25: the framework, not a person, resolves these classes). Session: "bugfix 033".

## Why no 090 diagnosis run (the 2026-07-31 ruling's named escape hatch, used openly)

The mechanism is not in dispute and was verified first-hand rather than through the loop:
"no handler in the fleet claims this type" is a grep (`HandlerAgent: ""` at
`check_required_fields_missing.go`, one producer, zero consumers in code or
`agent_definitions`) plus one live query (exactly one item in platform history ever terminal —
`revalidate_review_queue_action.go:64-66` records the same measurement). The change implements
an explicit owner ruling, not a diagnosed root cause; the design risk was pressure-tested by a
planning agent against the code (verified mechanics recorded in NOTES) and by the council gate
(submission `7b0e2833-715f-4a9a-897b-efd913073582`).

## Design decisions and their reasons

1. **Router, not repairer** (IMG-071 / seed 397 pattern). The population is heterogeneous and
   two classes MUST NOT be auto-repaired: blob components (content_data NULL under serving
   rendered_html — regeneration replaces served HTML, bugs_open/263) and owned/tool pages with
   no plan (the owned-page guard, itself an owner ruling — reconcile_site_plan decision 3).
   Census 2026-08-15 (n=44): no_content_data 35 / stale 6 / no_plan_generic 1 /
   no_plan_owned 1 (the gas converter) / partial 1.
2. **Park-in-place, not checkpoint_for_review, for the two human classes.** The checkpoint
   action writes no item_key (no dedup) and hardcodes `handler_agent='human-review'`
   (unregistered); completing the original releases its dedup key → producer re-raises →
   two-strike births endless `unresolved` rows. Parking at `needs_human_review` HOLDS the key:
   churn is structurally impossible, the row stays on the dashboard with the router's triage in
   the error column, and the revalidator remains a second close path. Parking one's own item is
   first-class: `complete_work_item`'s guard no-ops benignly (`load_work_item_actions.go:956-978`).
   **This was a correction to this session's own first design** (checkpoint-and-complete),
   found by the planning agent's review — recorded in NOTES.
3. **Classification key is (page_name, slot_name)**, the revalidator's own key — never
   `spec.component_id` (016b §9: 11/45 items resolved to nothing when keyed on component_id).
4. **Conversions are born `triaged`** — the `detected` promoter is disabled (bugs_open/083);
   a `detected` item is stranded. Conversion item_keys are stable (`content_rewrite:from_rfm:`
   + component_id) so the two-strike brake works.
5. **`partial` converts to `content_rewrite` with `mode='edit_live'`** (PBP-028's third
   emitter, clause updated) — the writer edits current prose rather than fabricating; bug
   238's resolver-key protections confirmed in the running binary (stamp `a2a6912…`, both fix
   commits ancestors).
6. **Producer flips to routed-from-birth** (`HandlerAgent=required-fields-missing-handler`,
   `Status='triaged'`) — Go, inert until a chassis roll; the seed + assignment carry the live
   half until then.

## Phasing

1. Census (done — output saved beside this file; the seed's exact embedded SQL re-run against
   the five canary candidates routed all five as the census predicted, before any apply).
2. Council submission `7b0e2833-715f-4a9a-897b-efd913073582` (submitted 2026-08-15 ~11:0x).
3. Commit + apply seed 410 (inert; verify block asserts 0 assigned).
4. Canary assignment (4 rows: stale 332bb3f6, partial 4fa5b019, blob e512af8a, gas converter
   483fb749) → verify each arm → fleet assignment.
5. Commit the Go producer change (rides next chassis roll; post-roll re-run the assignment
   UPDATE once for stragglers filed pre-roll).
6. 033/277 bug files updated; 277 stays OPEN until fixed-AND-live per the closing bar.

## Out of scope, recorded

Blob decomposition (staged_component_build's domain), tool-page rebuild (tool lane),
033's other remaining pieces (Retry-refuse for handler-less items; owner decisions B/D; D3
identity; revalidator v2; the other ~20 uncovered types).
