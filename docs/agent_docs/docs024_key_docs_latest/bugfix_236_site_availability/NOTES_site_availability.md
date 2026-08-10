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
