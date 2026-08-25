# HANDOFF 2026-08-25 — second-click page BUILT and APPROVED; it is not live, and that is the whole state

**Read order, cold:** this file, then `HANDOFF_2026-08-24c_continue_here.md` (which it
supersedes for items 1 and 2 only — 3, 4 and 5 there are untouched and still owed).
Depth: `NOTES_web_admin_console.md` 2026-08-25 · `../webdesign_uk_build_service/DECISION_2026-08-24_confirmation_needs_a_second_click.md`
· `COUNCIL_SUBMISSION_2026-08-25_second_click_confirmation.json`.

## 0. State in one paragraph

**The one owed code task is done and reviewed, and it is NOT LIVE.** `/c/<token>` now splits
by method in core-manager: `GET` renders a page with one button and reaches no database on any
arm, `POST` on the same path confirms. Commits `24b63120d` (the split) and `d1a4bdcdf` (one
shared route table, answering the guardian seat). Council `ea99befa-ec62-4f61-b052-c3af3d003d55`
**APPROVED round 1, all reviewers**, in five minutes. **Nothing ships until a core-manager roll
carries `d1a4bdcdf`**, and the webdesign lane's first delivery email stays blocked until then —
that is the owner's ruling, not a preference. `customer_access_tokens` = 0 and sites
handed_over/transfer_confirmed = 0/0 [MEASURED 2026-08-25], so nothing is at risk while it
waits. The Builds screen from 08-24 is live and was re-proved this morning after the ~09:25
roll (dashboard `v1.0.1337`, core-manager `4c996e1b5`).

## 1. NEXT — in order

1. **After the next core-manager roll, close this out — and it is now a TWO-STEP sequence with
   an order, because another lane moved the routes under this change on 2026-08-25.**
   ⚠ **`/c/` is currently mounted NOWHERE and an outside probe 404s — that is EXPECTED, not a
   defect in the second-click page.** `d30917150` (RFC_054 Q2, register **SYS-095**) moved the
   delivery routes off the admin port onto a delivery-only listener on `:8090`, opt-in via
   `SERVICE_SERVER_DELIVERY_PORT` (set in core-manager's production overlay), and repointed
   `box/links.webdesign.uk.nginx` at `:8090` in the same commit. Until BOTH the roll and the box
   apply have happened, a customer link 404s at the box either way — harmless today
   (`customer_access_tokens` = 0) and **indistinguishable from my change being broken**, which
   is why it is written here.
   Close-out, in this order:
   a. **Roll lands** → ancestry against the pod's provenance stamp, with the reversed test as a
      control. Two commits must be contained: `d1a4bdcdf` (this page) and `d30917150` (the
      listener). One without the other is a half-shipped state.
   b. **Then the box config is applied** (owner action, links.webdesign.uk vhost). The file's own
      header carries the check to run first and says why applying it EARLY breaks every link
      silently. Do not apply before (a).
   c. **Then probe the capability from outside** — `GET` renders the button page, `POST` the same
      path answers, and `POST …/c/<token>/confirm` 404s at the box (the control). Add SYS-095's
      own containment check in the same pass: `core-manager:8090/api/v1/admin/work-items` → 404
      **paired with** the same path on `:8088` → not-404, or the pairing proves nothing.
   d. **Then** tell the webdesign lane the delivery email is unblocked.
2. **RFC_054 still awaits a ruling** (unchanged from 08-24c): Q1 two-door pattern · Q2
   delivery-only listener · Q3 what makes door three automatic.
3. ~~**`WriteSiteSpecAction` deep-merge follow-up**~~ **CLOSED 2026-08-25 — REFUTED by its own
   census; do NOT open a council round.** No agent writes `evidence_base` through that action
   (0 of 20 live steps, 0 across snapshots too, 0 templated — control: 157 templated steps
   elsewhere), and the scheduled refresher shrank a register **0 times in 222 writes**. The one
   emptying in all history lasted **59 seconds** and was one session's own two-part write. Full
   census in NOTES 2026-08-25. **Residual, different from the item:** 8 of 19 sites have no
   refresher coverage at all, so nothing re-derives their registers — and the only door that has
   ever emptied one is hand-written SQL. That is a review habit (`DO`/`RAISE` verify blocks on
   migrations touching `evidence_base`), not a mechanism to build.
4. **Owner eyeball: Builds screen DONE 2026-08-25 (feedback acted on, commit `8e5a35ef9`).
   The Terminate half is BLOCKED ON DATA, not on the owner.** ⚠ 08-24c's "press Terminate on a
   months-old EXECUTING_STEP orchestration" **cannot be followed: no months-old rows exist.**
   `[MEASURED 2026-08-25]` `orchestration_states` holds nothing older than **2026-08-24** (~2
   days — the table is reaped), and fleet-wide there are **7** non-terminal orchestrations, of
   which **2** are tagged to any site (`webdesign.co.uk`, started today, and that is the other
   thread's live lane — do not terminate it to test a button). So the endpoint fix stays proven
   only at the statement (sqlmock) until a genuinely stuck run appears on a site this lane owns.
   The screen now pins any running orchestration above the fold and says "Nothing running…"
   when there is none, so the next real one is findable rather than hunted for.

## 2. What is true today, with its proof

| thing | proof |
|---|---|
| Builds screen still live after the 08-25 ~09:25 roll | dashboard `v1.0.1337`; served bundle `assets/index-Bqjp4Gs8.js` (**same hash as 08-24 — content-derived, so an unchanged SPA keeps its name through a roll**) carries all five markers, control string returns 0; core-manager stamps `4c996e1b5`, ancestry to `e6350e74b` passes, reversed test fails |
| second-click split correct | two mutations fail the suite (route GET to the confirm handler; make the page handler confirm anyway), clean suite green, `verify-head-builds.sh --test --with` OK against HEAD `c05e4f90d` |
| nothing else consumes the old GET behaviour | repo-wide: every hit is inside the change. Live: `agent_definitions` naming `/c/` = 0, current `site_specs` naming `links.webdesign.uk` = 0, control 21 |
| no box change needed | `links.webdesign.uk.nginx` already has the anchored `/c/` regex with `limit_except GET POST` |

## 3. Falsifiers

- **`customer_access_tokens` non-zero before the roll** → raise it with the owner at once: a
  delivery email may be about to go out against the ruling, and the page it needs is not live.
- The bundle hash names TODAY's image; re-grep the served bundle, never cite the hash.
- Tags roll daily — re-ask each pod its provenance stamp **per service**.
- kubectl token expires on the 3-day cycle (last refreshed ~2026-08-24 21:00, so due ~08-27).
- A newer handoff here or in `../webdesign_uk_build_service/`. **That lane has an active
  thread**; a cross-lane note for them is at
  `../webdesign_uk_build_service/NOTE_2026-08-25_second_click_built_by_web_admin_console_lane.md`
  and none of their files were touched.
