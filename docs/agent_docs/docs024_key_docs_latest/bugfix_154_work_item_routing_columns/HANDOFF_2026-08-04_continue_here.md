# HANDOFF 2026-08-04 (rev 2, evening) — 178's fix works for ONE case, reproduces the ORIGINAL bug for another; still OPEN

**This revision replaces the morning version of this file, which contained
two claims that turned out wrong within hours — both are corrected here and
in `bugs_open/178` itself, with the checks that would have caught them named
inline. Read `bugs_open/178`'s own file in full before doing anything
further; it now carries the complete, corrected account. This file is the
short version + the open task list.**

## What actually happened today, in order

1. Implemented candidate 1 from `bugs_open/178`: opt-in `spec.mode="edit_live"`,
   new step `load_current_section_content` in `page-build-handler`, both live
   `content_rewrite` emitters updated. Committed, built, deployed
   (`v1.0.1247`), migration applied.
2. **Bug introduced**: the new step's `output_field` reused the key
   `section_plan` (deliberately, to avoid touching `call_content_writer`'s
   mapping), but the action RETURNED A WRAPPER object on every path,
   including its "pass-through" ones. Because the orchestrator stores an
   action's return value wholesale under `output_field`, this silently
   replaced the real plan with a wrapper on **every page build in every
   mode, fleet-wide** — not just `edit_live` ones — from the moment the
   migration was applied (~08:20) until it was fixed (~09:01Z).
3. **Misdiagnosed my own outage as a pre-existing, unrelated bug** — filed
   it as `bugs_open/192`, reasoned (wrongly) that a historical failure spike
   the previous evening proved the cause predated my code. It didn't: that
   evening's spike was a different, still-undiagnosed failure at a different
   step; my own two failures that morning shared only the error STRING, not
   the cause, with that spike.
4. The `bugfix_192_select_sections_wrapper` lane read `192`, correctly
   diagnosed it as this fix's wrapper bug, fixed it (live seed workaround +
   committed Go fix, shape-preserving return value now, mutation-tested),
   and notified `178` directly. **Their fix is good and this lane owes them
   nothing on that half.**
5. **Also got a second thing wrong today**, caught by the owner asking to
   check specifically: claimed `build-dispatch-loop` has no scheduler at
   all. Wrong — `scheduled_tasks` has `build-pipeline-trigger` (enabled,
   fires every 120s via the `kafka-scheduler` service), which calls
   `build-dispatch-loop` for the oldest-eligible site every cycle. Missed it
   because the search filtered on the substring `%dispatch%`, and this row's
   name doesn't contain it. No k8s CronJob involved either way — checked,
   per the ask.
6. Re-ran verification once builds worked again. **Split result**:
   - `guide-cma-compliance`: content preserved, +206 chars, fix worked
     exactly as designed.
   - `guide-independent-strategy`: **the original bug reproduced**, by a
     route this fix does not cover — `plan_sections`' component selector
     resolved the section to a different, generic fallback component than
     what was actually stored, so the exact-name join in
     `load_current_section_content` found nothing to attach, and the writer
     fabricated fresh content exactly as it did before this fix existed.

## State: `bugs_open/178` is OPEN, not closed

The fix is real and live, and demonstrably works for the case it targets
(stable component/slot identity across rebuilds). It does **not** cover a
section whose build-time resolved component differs from what's stored —
which is exactly what happened on the second test page, using the platform's
OWN generic-fallback components (`generic-text-block`, `article-body`).
Old content is recoverable either way (`page_component_history`), so nothing
irreversible has happened, but the bug's actual promise — "a link-insertion
item cannot silently gut a page's prose" — is not yet true in general.

## OPEN — in priority order

1. **Design and implement a fix for the component-identity-drift gap.**
   `bugs_open/178`'s final update names the two directions (make the match
   tolerant of a resolved-name miss by falling back to "the page's current
   prose slot", or fix the underlying instability of generic-fallback
   component assignment across rebuilds) but does not choose between them —
   that's the next real task. Measure how often this actually happens
   fleet-wide before investing heavily; this session has one observed
   instance, not a rate.
2. **Confirm the 192 fix's Go half survives the next chassis roll** (it's
   committed, `2b9d84072`, but inert until then — currently the fleet is
   running on the live seed workaround only). Pod-grep the marker their
   commit names once a roll happens.
3. **Council submission for 178 stalled**, not rejected —
   `Council-Submitted: 97ebadcf-bbe6-485f-8231-ff16fc4e679f`, reached
   `review_constitution` at 20:09:59Z on 08-03 and never advanced, no
   verdict artifact written. Advisory only, doesn't block. A fresh
   submission covering the corrected fix (once the drift gap has a design)
   would be the natural point to resubmit, not before.
4. **Watch list, not urgent**: the shrink guard did not fire on the
   `guide-independent-strategy` case, because a whole-slot RENAME (old slot
   gone, new slot appears) isn't the same-slot shrink the guard compares —
   worth someone's attention as a second, narrower gap in that guard's
   coverage, separate from the fourth-floor deferral already tracked there.

## Landmines specific to this lane (carry-forward + corrected)

- All landmines from the 2026-08-03 handoff (shrink guard, dependency
  release rules, dispatch quiet-spell reading, `orchestration_states`
  retention) still apply unchanged.
- A `content_rewrite` item created BEFORE the emitter fix has no `mode` key
  — patch `spec` with `mode:"edit_live"` before releasing any more you find
  predating `08d0515f3`, or it hits the old destructive path.
- **A step whose `output_field` reuses an upstream key must return that
  value SHAPE-PRESERVING — never a wrapper, not even on a "pass-through"
  path.** This is now `016b` §9 (`5f569bb8e`) — read it before writing
  another step like this one.
- **`build-dispatch-loop` IS scheduled** (`build-pipeline-trigger`, 120s).
  Don't manually fire it out of a belief that nothing else will.
- `bugs_open/087`'s error string is not unique to it — `192` produced the
  identical `sections_for_render.sections_ready not found` message from a
  completely different cause. Read which agent actually failed
  (`page-rebuild` vs `page-content-writer` via build-handler) before
  assuming which bug you're looking at.

## Cold-start pointers

- `bugs_open/178`'s own file is now the authoritative account — read it in
  full, including its two `> **CORRECTED**` blocks and the final "ONE page
  confirms, ONE reproduces" update.
- `NOTES_work_item_routing_columns.md`'s 2026-08-03/04 tail (five entries
  now, including two correction entries).
- `bugs_open/192` (owned by the other lane, diagnosis + fix both there).
- Register entry PBP-028 (`docs026_concept_register/register/page-build-pipeline.md`)
  still describes the mechanism correctly; its own "OPEN REVIEW QUESTION"
  about matching by slot_name has gone from theoretical to confirmed — worth
  updating when the drift-gap fix is designed.
- Commits: `08d0515f3` (178 fix), `0a2a94b89` (tag bump), `75fceb501` (192
  filing, since corrected in place), plus the 192 lane's `a0e3ecee8`/`2b9d84072`.
