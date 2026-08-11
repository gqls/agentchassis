# HANDOFF — bugfix 236 (522 half) — **THE WORK THIS FILE ASKED FOR IS DONE**

> **CLOSED 2026-08-11 10:20Z.** Everything this handoff was written to hand over
> has been completed: the drill ran (both halves), the council approved, and the
> rotation is steady-state. **Do not start from the task list below** — it is kept
> only so the reasoning behind the choices is readable. If you are cold-starting,
> read this page and then `NOTES_site_availability.md` entry *2026-08-11
> 09:55–10:15Z*, which is the evidence.

## What is true now

| thing | state | proof |
|---|---|---|
| `check_site_unreachable` | live in **v1.0.1284** | both replicas: `site_unreachable` 7, invented negative `site_unreachabl3` **0**, pipeline control `asset_reference_404` 13 |
| healthy path | proven at fleet scale, twice | 22/22 eligible sites (21 `deployed` + 1 `active`), 0 items, two full rotations |
| **filing** | **PROVEN IN PRODUCTION, twice, two mechanisms** | pool site → `transport_error` (DNS); **cookly.uk route deleted 90s** → item 10:09:41Z |
| **self-clear** | **PROVEN IN PRODUCTION** | route restored → probe 10:15:11Z → `items_resolved: 1`, item `detected` → `complete`, no human |
| council | **APPROVED** | `7177fb02` at 10:11:07Z, 6 advisory objections, none high-severity, 2 abstained |
| coverage | credited | 098 lists `4a5d77004` as REVIEWED, resolved from its `Council-Submitted:` trailer at report time |
| fleet | clean | 0 open `site_unreachable` items; 1 `cancelled` (drill, with provenance) + 1 `complete` (the self-clear) |

Commits: `4a5d77004` (check + tests + IMP-053), `79feb08e7` (hold released, mig
372, roster fixture), `4864a8754`, `6474f792a`, `387dad387` (fleet passes + blast
radius), `980fe3e01` (the drill, `Council-Reviewed:`).

## What is NOT closed by this, so nobody assumes it is

- **Candidate 2 (zone/route conformance) stays with `domains_cloudflare_rollout`.**
  Candidate 3 is not taken. `bugs_open/236` covers only candidate 1.
- **Two council objections stand unfixed, on purpose**, recorded in IMP-053:
  (a) this lane's commit classifies **another lane's** `decision_regression` in the
  shared `verifier_coverage_test.go` — guardian and editquality both called it
  surface bleed, they are right, and forward-only means it cannot be unbundled;
  RFC_015's owner should know it was this lane. (b) `reuse_agent` asked why
  `backend_unreachable` was not generalised into one target-aware reachability
  check instead of a second mechanism. A design call for the improvement-loop
  owner; it does not change what a shared mechanism *guarantees*, so the
  2026-07-29 ruling does not make it an RFC.
- **`availability-discovery-agent` declares no `input_contract`** (guidelines seat,
  low severity). Worth a look; not a defect today.
- **`title_absent` still files nothing.** A registrar-parked domain answering 200
  lands in findings, not the queue — measured trade (filing on it was 1/21
  false-positive on day one). Revisit when serving-mode metadata exists.
- **The alarm is a flag, not a pager.** Nothing emails anyone. Same accepted
  posture as `backend_unreachable`; the undrained-detector cost is `bugs_open/033`.

## The three things this lane learned that outlive it

1. **A missing worker route HANGS the apex; it does not fast-fail with 522.** The
   probe recorded `context deadline exceeded` at its 15s timeout. The generous
   timeout was a guess when written and is a measurement now — a 5s probe might
   have missed the class entirely. In `LANDMINES.md`.
2. **Both edges of a route change lag, in opposite directions** — 200 for ~30s
   after a successful DELETE, 522 for ~18s after a successful CREATE. One `curl`
   either side reads as "the change failed". Poll; and re-list the routes.
   In `LANDMINES.md`.
3. **A discovery work item's `created_at` is when the RUN's transaction opened,**
   not when the check filed — `run_discovery_checks` holds one transaction across
   every check and `now()` is `transaction_timestamp()`. This nearly became a
   confident false bug report that the confirm-before-filing guard was broken. The
   disconfirming measurement was orchestration wall time (7.64s unreachable vs
   1.8s healthy — the delta *is* the 5s retry). `created_by` lies the same way: it
   is the *sender*, so scheduled runs read `generic`. In `LANDMINES.md`.

## If you are here to change the check

- Code: `platform/orchestration/actions/discovery_checks/check_site_unreachable{,_test}.go`.
  The test file header carries the **8-mutation table** — each guard broken and the
  NAMED test that caught it. **Re-read it before touching any guard.**
- The drill recipes, with four gotchas, are in `RUNBOOK_site_availability.md`
  (§ "The break-it-on-purpose drill"). Both halves are reproducible from there.
- **The rotation's agent_type list must come from the DATA, not from
  `346_site_discovery_rotation.sql`** — `SELECT DISTINCT agent_type FROM
  site_discovery_rotation` returns five; the migration names three.
  `render-audit-agent` (`369_*.sql`) is the only one currently enabled.
- **Do not renumber migration 372.** Its number collides with another lane's
  applied `372_provocation_generator_token_budget.sql`. Both are recorded in
  `schema_migrations` by FILENAME and both applied cleanly; renumbering an applied
  migration makes it pending and re-runs it.
- **The bug is by SLUG, not number.** The other 236 is
  `hero_and_logo_deployed_lose_image_url`. Same trap on 243.

## The sentence this file used to end on, and why it can now be retired

It read: *"the machinery is installed and healthy-path-proven, and its ability to
raise an alarm is proven only in tests."* That is no longer the honest sentence.
The honest sentence is: **a real site was taken off the internet for 90 seconds,
the platform noticed on its own, and it put the alarm away by itself when the site
came back.** The distinction `bugs_open/236` was filed about is closed — and the
reason it took a deliberate outage to close it is that `0 findings` and a blinded
check are the same observation (016b §9), which stays true of every future clean
sweep quoted from a chassis later than v1.0.1284.
