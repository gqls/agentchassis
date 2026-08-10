# NOTES — bugfix 236 site availability (append-only, newest at the bottom)

## 2026-08-10 — lane opened

- Bug selected after a fleet-wide ownership sweep: banner-scanned all 82 files in
  `bugs_open/`, who-owns on 13 candidates, live-transcript grep on 12 active
  sessions. Most "open" candidates were fixed-and-live under the 08-06 owner
  ruling (finished bugs stay in `bugs_open/`); genuinely open + unowned:
  029 (owned in substance by dispatch-gate lane + 003), 071 (owned, fragment lane),
  153 (fix touches the makefile — dirty with another session's edits), 146-ported
  (adjacent to the live tool-sweep session), 236-522 (this one).
- **Misstep (logged, cheap):** my first stored-title query encoded
  `pages.status='deployed'` — the exact predicate trap of `bugs_open/185`
  (pages.status is active/archived; deployment lives in build_status/deployed_at).
  Caught in one query because every title came back empty — an all-empty result
  from a plausible filter is the "small world" smell. Re-ran with
  `PageHasShippedPredicateFor`'s SQL shape. Not a WRONG_CALLS row: caught before
  any claim was written down, and the check that caught it (empty result) is
  already the documented one. The lesson that IS worth carrying: the helper
  exists precisely so nobody types this predicate again — use it in the check.
- Live probe of all 21 deployed sites (from this workstation, not the pod):
  21/21 → 200 + HTML. Title assert: 19/21 match; webdesign.uk fails via its
  deliberate off-domain 302; mortgagecalculator.co.uk serves a title the DB does
  not hold (stale/divergent render). Decision D4 (PLAN) follows from this.
  `[MEASURED 2026-08-10 ~15:45Z, commands in RUNBOOK]`
- `[ASSUMED]` the chassis pod's egress sees the same public serving path as this
  workstation (no split DNS for fleet domains in-cluster). Precedent:
  check_asset_reference_404 / check_tool_acceptance already probe deployed pages
  from the runner. To be confirmed at verification step 3 (a live run).
- Rotation drivers from the 230 lane found ENABLED and ticking
  (site-discovery-rotation-{quality,design,completeness}, hourly, 7-day
  cooldown) — the bug file predates them, so the fix plugs into a seam that did
  not exist when 236 was filed. The `CheckResult.Resolved` doc comment names a
  health probe with `AllOfType` as its intended example — the framework
  anticipated this check.

## 2026-08-10 (later) — code written, and two things the tree taught me

- **Check + tests built and green.** 8 guards, each proven load-bearing by
  mutation with a NAMED failing test (table in the test file header). Ran against
  `git archive HEAD` in a scratch tree, not the working tree — the working tree
  does not compile (`datahelpers/page_canonical.go:185: undefined:
  nestedOrFlatURL`, another session's WIP). A green local build here would have
  been unobtainable and a red one would have been misattributed to me; the
  archive is the only build that answers "does MY change compile against HEAD".
- **MISSTEP — the package test was already RED at HEAD, and I nearly read it as
  mine.** `TestEveryCheckProducedItemTypeIsClassified` fails on
  `decision_regression` (check_decision_guards.go, shipped 2026-08-08 in
  `e1628f7df` without a classification). My first reading was "my new item type
  broke the sensor" — wrong: the sensor had already accepted `site_unreachable`,
  and `decision_regression` is a different lane's. **Cheap check that settled it
  in one command: run the same test on a pristine `git archive HEAD` tree.** Two
  days red, so several sessions have hit it. Classified it in passing (the
  package cannot run at all otherwise), labelled as not-my-lane, and left the
  verifier decision with RFC_015's owner.
- **MISSTEP — migration number 368 was free when I checked and TAKEN when I
  committed.** Another session landed `368_info_card_grid…` (and 369) at 15:55,
  ~7 minutes after my check. Renumbered to 371. Also visible in the tree:
  **370 is already claimed TWICE by two different uncommitted lanes** — so the
  collision I hit is not rare, it is the steady state. The check that caught it
  was re-running `ls sql_for_agents/` at commit time rather than trusting the
  number I had chosen minutes earlier. Logged in WRONG_CALLS.md.
- Register entry IMP-053 + index row added in the same commit as the code, per
  the platform-seams rule. Re-grepped the index at commit time: 1,811 rows,
  0 duplicate ids; the row/entry off-by-one is **pre-existing at HEAD** (verified
  with `git show HEAD:` — 1,810/1,811), so it travels here unfixed and is
  recorded as not-mine rather than silently inherited.

## 2026-08-10 (evening) — committed, and the council could not review it

- **Code + register + held migration committed `4a5d77004`**, trailer
  `Council-Submitted: 7177fb02-51c5-4c2a-bb02-10aa27ae85ca`. The trailer gate in
  `commit-msg` refused a placeholder value first (`pending-this-session`) and it
  was **right to** — the trailer is a join key for the 098 report, forward-only
  forbids an amend, so a non-resolving value would have been permanent. Submit
  first, then commit.
- **The council run reached NO verdict, for a reason outside this change.**
  Panel selected (10 seats matched), `fix_plan` persisted, then
  `review_editquality` died on an upstream Anthropic **400: "You have reached
  your specified API usage limits. You will regain access on 2026-09-01"**, and
  the orchestration terminated at `complete_invalid`.
  - **MISSTEP worth recording: I read `complete_invalid` as "my submission was
    rejected as invalid" and went looking for a schema error in my own JSON.**
    It is not a verdict at all. The discriminator is
    `collected_data->'__step_error'->>'failed_step'` plus the absence of any
    `review_*` key — an actual REVISE/REJECTED verdict writes a `council_report`
    artifact, and this correlation has only a `fix_plan`.
  - Fleet-wide, measured: last successful LLM call **14:51:45Z**, then 5 failures
    across four agents, 0 successes. Already diagnosed by the webdesign.uk chat
    lane (`6a4fbab21`) as an account-level spend cap; contributed my
    different-credential-path evidence into their NOTES rather than filing a
    duplicate.
  - **Consequence for this lane:** the submission is on record and 098 will credit
    it automatically *if* the correlation is ever approved — but it will not be,
    because the run is terminal. **A fresh submission is owed once the cap
    lifts**, and the trailer on `4a5d77004` should be treated as "submitted, never
    reviewed", not as review cover.
- **The migration number collided TWICE** (368 → 371 → 372) — see WRONG_CALLS.
- **Still owed, in order:** (1) resubmit to the council when LLM access returns;
  (2) chassis roll, then pod-grep `site_unreachable` on every replica; (3) rename
  `372_*_HOLD.sql`, apply, add `site_unreachable` to `liveConfiguredChecks` in
  the same commit; (4) the break-it-on-purpose drill on a sacrificial worker route.
