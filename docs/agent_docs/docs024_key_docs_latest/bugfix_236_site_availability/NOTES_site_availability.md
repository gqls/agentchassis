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

## 2026-08-10 22:00–22:10Z — LIVE. Hold released, migration applied, first probe end to end

- **v1.0.1283 carries the check.** Pod-grep BOTH replicas, one exec each, with
  controls in the same exec: `site_unreachable` **7**, `+site_unreachable` **1**,
  invented negative `site_unreachabl3` **0**, pipeline control
  `asset_reference_404` **13**. Pods `696d88b4c7-95mgb`/`-wnbs8`, started 21:43Z;
  probed at 22:02Z, i.e. 18 min after start — clear of the ~300s dispatch-drop window.
- **Migration 372 APPLIED** (`4864a8754`/`79feb08e7`). Dry-run and apply were run
  against a **temp `MIGRATIONS_DIR` holding only my two files** — the runner takes
  EVERY pending file otherwise, and the shared dir had other lanes' pending work in
  it. Probe ok, apply clean, then the live rows read back: agent
  `checks=["site_unreachable"]`, task enabled at 300s in its own group.
- **FIRST LIVE RUN, 22:03:47Z, `4ec82e1c` COMPLETED in 1.8s** on `robot-hands.com`:
  ```
  checks_run: ["site_unreachable"]   checks_failed: []   checks_unregistered: []
  findings: null   items_inserted: 0   items_resolved: 0
  ```
- **What that run PROVES, and it is more than "it did not crash".** The
  `[ASSUMED]` from this lane's opening — that the chassis pod's egress sees the
  same public serving path as my workstation — is now **settled, and by a check
  that could have come out the other way**: if egress were blocked, both probe
  attempts would transport-error, the verdict would be UNREACHABLE, and the run
  would have **filed an item**. A blocked network produces a FALSE POSITIVE here,
  never silence. So 0 items on a site that serves rules egress failure out.
  Wiring, registration, egress and the healthy path are all live-proven.
- **What it does NOT prove, and this is the honest gap:** the FILING path and the
  self-clear have not run in production. `0 findings` is also what a blinded check
  reports (016b §9). Only the break-it-on-purpose drill closes that, and it is
  still owed — see the handoff.
- **MISSTEP — the `git mv` pathspec landmine fired on me, verbatim.** `git mv
  HOLD -> non-HOLD` plus a commit naming only the NEW path shipped a **copy**:
  both files sat at HEAD, one of them saying HELD about config that was already
  live. Harmless in effect (`SIDECAR_RE` never auto-applies a `_HOLD` file, and the
  migration is idempotent) and caught by running the entry's own check —
  `git ls-tree -r --name-only HEAD -- <dir> | grep 372` — which is the point: `ls`
  on disk showed one file and would have told me nothing. Fixed in `4864a8754`.
- **The 372 number collided a THIRD time, and this one must NOT be "fixed":**
  another lane applied `372_provocation_generator_token_budget.sql`. Both are
  recorded in `schema_migrations` **by filename**, so both applied cleanly and
  coexist. Renumbering an APPLIED migration would make it pending again and
  re-run it. Leave it. (Chain for the record: 368 → 371 → 372, and 372 was taken
  anyway. See WRONG_CALLS.)
- **Cadence, measured not assumed:** the task takes `LIMIT 1` per 300s tick, so a
  cold start drains 21 sites in ~105 minutes, then settles to the 4-hour cooldown.
  First stamp 22:03:46Z.
- **The Anthropic cap (`bugs_open/243`) does not touch this lane.** The
  availability agent has no LLM steps — it is currently one of the few discovery
  paths in the fleet that still functions. The council round is still owed.

## 2026-08-10 22:15Z — cadence CONFIRMED, and a near-miss on my own measurement

- **Three ticks observed, one site each: 22:03:46, ~22:09:4x, 22:14:46.** Exactly
  the designed `LIMIT 1` per 300s, so the ~105-minute cold-start drain for 21
  sites is measured now, not predicted. The read-only form of the pre_query's
  `due` set returns the expected queue (loancalculator.co.uk, cookly.uk, idea.uk,
  … — unstamped sites sort first). 0 `site_unreachable` items so far, correct.
- **NEAR-MISS, caught before it reached any file: I nearly recorded "the rotation
  has stalled — 1 site in 17 minutes".** It had not. I read `date -u` once at
  22:01Z, then estimated the later wall-clock from how much work I had done in
  between, and was eight minutes out; the DB's own `now()` said 22:09, i.e. the
  second tick was not yet due. **The cheap check is to put `now()` in the same
  SELECT as the timestamps you are judging** — which is what the query above does
  and what settled it. An elapsed-time claim assembled from memory is an
  `[UNMEASURED]` figure wearing a measured one's clothes, and it would have sent
  the next session hunting a scheduler bug that does not exist.

## 2026-08-11 09:26–10:30Z — the rotation has drained twice, and the cap has lifted

- **Full-fleet coverage, twice over, 0 items.** `[MEASURED 2026-08-11 09:26Z]`
  22 rows in `site_discovery_rotation` for `availability-discovery-agent`
  (21 `deployed` + 1 `active` — the census is 21/17/1/1 deployed/pool/system/active,
  so 22 IS the whole eligible set, not 21; the handoff's "expect 21" was one short).
  Every stamp is from TODAY (06:07:46Z → 08:03:16Z), i.e. this is a *second* pass,
  not the cold-start one — the 4-hour cooldown is cycling as designed.
  `site_unreachable` items fleet-wide: **0**. All 22 sites serve.
- **v1.0.1284 rolled at 09:23:20/45Z and the check survived it.** Re-greped rather
  than assumed (a roll is not evidence your fix shipped): `site_unreachable` **7**,
  invented negative `site_unreachabl3` **0**, pipeline control
  `asset_reference_404` **13** — one exec each, on BOTH running replicas
  (`7c9d5f74b9-6j5xn`, `-rvrdg`). This was worth doing: v1.0.1284 is somebody
  else's build, and nothing in it knows about this lane.
- **A third chassis pod (`-7w2ch`) was Evicted at that roll** — `Pod was rejected:
  The node had condition: [DiskPressure]`. Not this lane's, and **transient**: at
  09:27Z two of five nodes reported `DiskPressure=True`; by 10:29Z all five report
  `False`, self-cleared. Recorded because reading it at 09:27 alone would have made
  a passing node event look like a standing fleet condition — the same
  one-instant-reading error as the 22:15Z near-miss below it in this file.
- **THE ANTHROPIC CAP HAS LIFTED — the council round is unblocked.**
  `[MEASURED 2026-08-11 09:26Z]` `llm_call_log` over 12h: 117 successes at 22:00Z
  on 08-10, then 5 at 02:00Z, 1 at 03:00Z, 8 at 08:00Z, all `success=t`. The only
  failure in 14 hours is an unrelated `ollama-adapter` EOF at 09:23:32Z (the roll).
  So the cap that killed submission `7177fb02` ended ~22:00Z on 08-10, about seven
  hours after it started — **not 2026-09-01 as the API error stated.** Anyone
  carrying that date forward as fact should re-measure; it was the API's assertion
  about a billing period, never an observation.

### The drill: the handoff's "check with curl first" is answered, and it changes the shape

- **All 17 pool domains are `.internal`** (`pool-web-tech.internal`, etc.).
  `curl` → exit 6, HTTP `000`, could-not-resolve. So a pool site flipped into scope
  yields `transport_error` — a TRUE finding on a domain no visitor can ever reach.
  That is a better fixture than the handoff assumed it would have to hunt for.
- **The blast-radius risk in the handoff is MEASURED and smaller than feared.**
  Every `sites`-level reader of `status IN ('active','deployed')` in the tree, and
  what each does with a pool row `[MEASURED 2026-08-11]`:
  | reader | extra guard | pool site hits it? |
  |---|---|---|
  | `346` content rotations ×3 (quality/design/completeness) | `COALESCE(last_selected_at,'-infinity') < now() - 7 days` | **no, IF pre-stamped** — stamping is a hard WHERE exclusion for 7 days, not merely "sorts last" as the handoff put it |
  | `check_duplicate_palette.go:76` (`JOIN sites b`) | inner `JOIN style_collections` | **no** — all 17 pool sites have `style_collection_id IS NULL` |
  | `report_request_pull_action.go:127` | `deploy_config ? 'report_island'` | **no** — 0/17 |
  | `intent_collector_actions.go:114` | `deploy_config->>'target'='vm'` | **no** — 0/17, target empty |
  | `site_admin_handlers.go:51` (admin site list) | none | **yes** — one extra row in the dashboard for the drill's duration. Cosmetic |
  So with pre-stamping, the entire cost is one row in an admin list. 14 of the 17
  pool sites also have **0 pages**, so there is nothing for a content agent to act
  on even in the branch where one somehow selected it.
- **NEW: the pool variant CANNOT prove the self-clear, and the handoff did not
  notice.** `Run()` returns early on `siteStatus != active|deployed` — so the moment
  the drill reverts the site to `pool`, the check stops probing it and the
  `Resolved`/`AllOfType` path can never fire on that item. Reverting the status is
  what *prevents* the cleanup. The two halves therefore need two fixtures:
  file on a pool site, clear on a healthy real one (a synthetic item of the exact
  shape — `AllOfType` matches on `item_type`+`site_id`, so a hand-inserted row is
  indistinguishable from a check-filed one to the resolver).
- **What the cookly route-deletion drill would add over that, stated honestly:**
  very little *code* coverage. `judgeSiteProbe` sends `transport_error` and
  `http_522` down the **same** branch (`Unreachable: true` → identical filing
  code); the difference between the two drills is one string in `reason`. What it
  would add is a fact about Cloudflare — that removing a worker route really does
  produce a non-2xx rather than a 200 parking page — and proof that file-then-clear
  chains on ONE site. Cost: a genuinely live site down for the duration.
- **Cloudflare feasibility, checked read-only:** the token in `~/.cloudflare/404-token.env`
  is valid (expires 2026-09-01) and CAN read zones and worker routes — cookly.uk is
  zone `ab126cfa3debc8e1cf33fe8b741130bb`, route `1e11858e5c1146229c3238351b394146`
  = `cookly.uk/*` → `portfolio-sites-router`. It **cannot** read DNS records
  (`success:false`), so its scope is workers-only and **route DELETE permission is
  UNVERIFIED** `[UNMEASURED]` — the only way to find out is to attempt it, and a
  403 is harmless. Do not assume the route drill is executable from this
  workstation until that returns 200.
