# PLAN — bugfix 236 (522 half): a check that asks "does this site SERVE?"

**Started 2026-08-10.** Owning thread: this one. Ownership checked before claiming:
`who-owns.py` on the slug (filer = `bugfix_209_deploy_purpose_keyed_source`, whose
handoff `794308d2f` closed its brief), live-transcript grep of every session active
in the last hour (the sessions near this area are on 240, 224, 239, 228, 230, 151;
the one session given this same "pick a bug" prompt is drilling into 211/214/206),
and the diagnosis queue (0 `awaiting_diagnosis` rows).

**Bug:** `bugs_open/236_HANDOFF_2026-08-09_a_live_site_can_serve_522_to_every_visitor_and_nothing_notices.md`
**NOT this lane:** the OTHER 236 (`hero_and_logo_deployed_lose_image_url`) — ambiguous
number, unrelated case, owned elsewhere (rfc012 lane context). Refer by slug.
**Also NOT this lane:** the bug file's fix candidate 2 (zone/route conformance) — the
file itself says it is arguably `domains_cloudflare_rollout`'s, and that lane is live
in another session today. This lane builds candidate 1 only: symptom-side detection
at the served origin, which catches candidate 2's failure class *and* every other
cause of "nobody can load the site".

## Why no 090 diagnosis run (stated per the owner ruling of 2026-07-31)

The bug's root cause for the lendzy instance was established by the filer
first-hand and proven by the repair (route created → 200 within ~30s). The durable
claim this lane acts on is an ABSENCE — "nothing checks whether a site serves" —
and I re-verified it live today rather than by citation:

- `scheduled_tasks`: no availability/uptime/serve-shaped row; the only outbound
  health probe is `ai-endpoint-health-check` (AI endpoints, not sites). `[MEASURED]`
- `discovery_checks/`: `check_backend_unreachable.go` probes `/health` but **NOOPs
  unless `deploy_config.target='vm'`** — the static/worker-routed class (where
  lendzy died) is uncovered. `check_asset_reference_404.go` probes only
  subresources of an already-fetched page. `[MEASURED — read at HEAD today]`
- `site_work_items`: no open availability-shaped items. `[MEASURED]`

An absence re-verified by direct enumeration is first-hand verification; a 090 run
would re-read the same registry and the same table. (Note also the standing memory:
a 090 on the *sibling* 236 file came back UNVERIFIABLE on the harness.)

## The fix (bug file candidate 1, built on the seams that now exist)

Three pieces, smallest that closes the class:

### 1. New discovery check `site_unreachable` (Go, `discovery_checks/` package)

Per-site. Loads `sites.domain`, probes `https://<domain>/` (GET, redirects
followed, 15s timeout, ONE retry after 5s on failure — a transient blip must not
file a high-severity item). Then:

| observation | verdict | action |
|---|---|---|
| transport/TLS/DNS error (both attempts) | UNREACHABLE | file item |
| final status not 2xx (522/523/5xx/4xx at apex) | UNREACHABLE | file item |
| 2xx, but body empty or no `<html` | UNREACHABLE | file item |
| 2xx + HTML, final host ≠ probed domain | reachable (deliberate delegation) | finding only |
| 2xx + HTML, same host, stored index title NOT in body | reachable but suspect | finding only, named reason |
| 2xx + HTML, title present (or no stored title exists) | healthy | `Resolved{AllOfType}` |

- Work item mirrors `backend_unreachable` exactly: `item_type='site_unreachable'`,
  `item_key='site_unreachable:<site_id>'` (one open item per site via
  `idx_swi_dedup`), severity high, priority 30, `HandlerAgent: ""` (alert-only —
  no agent can create a Cloudflare route today; when one exists it becomes the
  handler), `Status: 'detected'`, `source='discovery'`, `pipeline='build'`.
- Self-clears through `CheckResult.Resolved` with `AllOfType: true` — the exact
  case the field's own doc comment names ("a health probe that succeeded").
- Stored-title lookup uses `datahelpers.PageHasShippedPredicateFor` — NOT a
  hand-typed `build_status='deployed'` (bugs_open/185; the commit-time pattern
  check flags the hand-typed form).
- Verifier: none; classified in `verifier_coverage_test.go` as `catMechanical`
  with the same self-clears rationale as `backend_unreachable` (IMP-021).

**Why title mismatch does NOT file (decision, with the measurement):** probed all
21 deployed sites today. 21/21 serve 200+HTML; 2/21 fail a title assert —
`webdesign.uk` (deliberate 302 → webdesign.co.uk; final-host rule handles it) and
`mortgagecalculator.co.uk` (serves "The UK's Authority on Mortgage Finance" where
the DB row says "Free Tools & Insider Guides" — a stale/divergent render, i.e. a
*different defect* with its own machinery). Filing on mismatch = 1/21 false
positive on day one; the estate's own recorded principle (links.go: "delisting a
working page is worse than the bug") says don't. **Stated limitation:** a
registrar-parked domain answering 200 (LANDMINES 2026-08-09) lands in the
"finding only" row, visible but unfiled. Accepted for v1; the finding carries
`reason: title_absent` so tightening later is a one-line policy change, and the
census measured 0 parked domains today.

### 2. Lightweight driver at outage-appropriate cadence (config only, migration 372)

The 230 lane's rotation drivers are the right pattern but the wrong clock (7-day
cooldown; an outage is not a content defect). Same migration shape as their 346:

- `agent_definitions` row `availability-discovery-agent`: 3-step clone of
  `quality-discovery-agent` (`ensure_site_record` → `run_discovery_checks` with
  `checks: ["site_unreachable"]`, `check_pipeline: "build"` → `complete`). No LLM
  steps, no spawns (so: no job topics — relevant while bugs_open/240 is live).
- `scheduled_tasks` row `site-discovery-rotation-availability`: the 230 pre_query
  verbatim with `agent_type='availability-discovery-agent'` and cooldown
  **'4 hours'** (their `'7 days'`), `interval_seconds=300`, own
  `concurrency_group='site-availability'` (own group so the `bugs_open/048`
  in-memory head-of-queue class cannot couple it to the content rotations),
  `max_concurrent=1`.
- Arithmetic: 21 eligible sites / 4h = 5.25 dispatches/h needed; tick capacity
  12/h → 2.3× headroom. Detection latency ≤ ~4–8h (was: unbounded/never).
  ~126 lightweight orchestrations/day (context: `ai-endpoint-health-check` alone
  fires 1,440/day).
- Rotation state reuses the 230 lane's `site_discovery_rotation` table — new
  `agent_type` value, no schema change. Their PLAN names the producer set as
  extensible by agent_type; still, they are consumers of the shared table and are
  being told (§Consumers below).

### 3. Ordering (the migration is a `_HOLD` until the image rolls)

The runner **hard-fails on a check name the binary does not register** (unless
`allow_unregistered_checks`, which stays untouched — it tolerates gaps silently,
which is the failure mode this estate documents). So: code commits now and rides
the next chassis build; the migration ships as
`372_site_availability_driver_HOLD.sql` — the migration-runner practice for
ordering-critical files (a banner cannot hold a file; the `_HOLD` suffix can) —
and is renamed+applied only after a pod-grep proves the check string is in the
running binary. Per the 2026-07-29 owner ruling: no ordering constraint is
*claimed* about the code half (HEAD is shared; the code is inert without the
config, so another session's roll shipping it early is harmless).

## Consumers of shared mechanisms touched (named, per the 2026-07-29 ruling §3)

- `site_discovery_rotation` (230 lane): gains rows with a fourth `agent_type`.
  Their watchdog CronJob (`site-discovery-staleness-check`) keys on the three
  content agents' stamps — mine does NOT break it; whether they extend it to the
  availability agent is their call. → note left in their NOTES file.
- `site_work_items` / `idx_swi_dedup`: one new `item_type`, single producer, key
  shape stated above. Per the 2026-08-02 ruling §1 this needs no RFC: producer
  set and key shape are named in the register entry (IMP-053).
- `verifier_coverage_test.go` sensor: new ItemType literal classified in the same
  commit, so the build-time guard stays green by decision, not by accident.

## What this does NOT do (scope, so the file can close honestly)

- Does not drain the queue: `detected` items with no handler stay visible-but-
  unclaimed (bugs_open/083's promoter question, 033's review surface — both open,
  both explicitly not this lane; same posture as `backend_unreachable`). The
  self-clear on recovery means a hand-fixed outage closes its own item.
- Does not check zone/route conformance (candidate 2 — `domains_cloudflare_rollout`).
- Does not catch a parked-200 domain (stated above, finding-only).
- Does not alert a human out-of-band. Detection converts "nothing notices" into a
  high-severity row a census CAN see; paging is a different mechanism.

## Verification protocol (before the bug file's status moves)

1. `go test ./platform/orchestration/actions/discovery_checks/` — new tests prove
   every verdict row in the table above via a swappable probe seam (the
   `probeAssetURL` pattern), plus mutate-to-verify on the filing branch.
2. Post-roll: pod-grep the binary for `site_unreachable` (positive) and for a
   string the change removed (negative control — n/a here, new file; use the
   check name in `checks.Names()` via a live run instead).
3. Apply migration 372 (rename from `_HOLD`), watch one rotation tick dispatch,
   confirm a `COMPLETED` availability orchestration and a `Resolved`-clean run on
   a healthy site.
4. **Break it on purpose** (the bug file's own protocol): delete `cookly.uk/*`'s
   worker route, confirm a `site_unreachable` item appears within ~8h (or force
   the rotation stamp to make it next), restore the route, confirm the item
   closes itself on the next pass. A checker that has never fired on a real 522
   is not evidence (LANDMINES: a gate's 0 findings has two causes).

## Decisions log

- **D1** Discovery-check seam over an endpoint-health-style bespoke action:
  filing/dedup/retraction/coverage-tests come free and reviewed; bespoke would
  re-implement all four (the drift class the council exists to catch).
- **D2** Own rotation task over joining the weekly content rotations: outage
  latency budget is hours, not weeks. Cost measured, trivial.
- **D3** Alert-only (`HandlerAgent: ""`) over inventing a route-repairing agent:
  no Cloudflare-API agent exists; a repair agent is candidate-2 territory and a
  new authority seam needing its own round.
- **D4** Title mismatch demoted to finding: measured 1/21 false-positive rate
  today; cry-wolf detectors get ignored, which re-opens the hole.
- **D5** `_HOLD` migration over `allow_unregistered_checks:true`: explicit
  ordering beats tolerated gaps.
