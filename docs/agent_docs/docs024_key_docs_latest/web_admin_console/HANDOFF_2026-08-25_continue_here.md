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

1. **After the next core-manager roll, close this out** (three steps, in NOTES 2026-08-25):
   ancestry against the pod's provenance stamp with a reversed control; then **probe the
   capability from outside**, including the `POST /c/<token>/confirm` control that must 404 at
   the box; then tell the webdesign lane the email is unblocked.
2. **RFC_054 still awaits a ruling** (unchanged from 08-24c): Q1 two-door pattern · Q2
   delivery-only listener · Q3 what makes door three automatic.
3. **`WriteSiteSpecAction` deep-merge follow-up** (unchanged from 08-24c item 4): census the
   legitimate-shrink history first.
4. **Owner eyeball still outstanding** from 08-24c item 1: admin.apis.uk → a site card →
   Builds, and press Terminate on a months-old EXECUTING_STEP orchestration. sqlmock proved the
   statement; only a real click proves the endpoint.

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
