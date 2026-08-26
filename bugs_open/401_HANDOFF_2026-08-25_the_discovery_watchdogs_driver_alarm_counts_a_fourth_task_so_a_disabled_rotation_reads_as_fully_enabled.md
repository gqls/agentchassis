# 401 — the discovery watchdog's driver alarm counts a FOURTH task, so a disabled rotation reads as "3/3 enabled"

Filed 2026-08-25 ~20:10Z by the `webdesign_tool_rebuilds` platform seat (session
`webdesign-tool-rebuilds`), found while running Track 2's post-deploy demand controls.

**Verification statement (per the owner ruling of 2026-07-31):** this filing substitutes
first-hand verification for a 090 run, and says why: the mechanism is twelve lines of a
watchdog script read directly, and every claim below is a live-DB row or a dated report body
quoted from `doc_notes` — the diagnosis loop would re-read the same file and re-run the same
queries. Nothing here is inferred from grep hits over unopened functions.

## The defect, in one paragraph

`site-discovery-staleness-check` (the daily watchdog half of `bugs_open/230`,
`deployments/kustomize/services/site-discovery-staleness-check/base/check.py`) has a
purpose-built alarm for "a rotation driver is switched off" — `driver_missing`, its reason for
existing, per its own docstring: *"Catches: the tasks disabled (the exact quiet grave 230
documents)"*. That alarm compares a **count** to a **count**: rows matching
`name LIKE 'site-discovery-rotation-%'` that are enabled (`check.py:75-79`, `:132`) against
`len(DISCOVERY_AGENTS)` = 3 (`check.py:49-53`, `:133`). On 2026-08-10 the `bugs_open/236` lane
added a legitimate FOURTH task matching that pattern (`site-discovery-rotation-availability`,
migration 372). Since the day that task was enabled, **any one content rotation can be disabled
and the enabled count still reads 3 — so `driver_missing` can never fire again for a single
dead rotation**, and the report header prints `rotation tasks enabled: 3/3`, which a reader
takes as "all three content rotations are on". It is exactly one disable away from being
wrong, and it has been wrong that way for a week.

## The live case it is currently masking

- `site-discovery-rotation-design` is `enabled=false`, disabled **2026-08-11 12:43:14Z**
  (`updated_at`; last fire 12:42:39Z) [MEASURED 2026-08-25, `scheduled_tasks`]. The disable was
  deliberate — the 2026-08-11 improvement-sweep cost incident, after which the owner ruled
  "enable the discovery rotations, slowly" and migration
  `395_enable_quality_discovery_rotation_slow_ramp.sql` re-enabled **quality only**, stating
  "`design` and `completeness` stay disabled" with a one-UPDATE foot for adding them later.
  Completeness was later re-enabled (by 2026-08-13, from the watchdog series below — at its
  as-shipped 3600s, not 395's recommended 10800s). **Design never was.**
- Consequence: `design-discovery-agent` has run on **zero** sites fleet-wide for 14 days —
  newest `site_discovery_rotation` stamp for it anywhere is **2026-08-11 12:42:38Z**, and it
  appears in none of `orchestration_states`' ~24h retention window [MEASURED 2026-08-25].
  Its 22-check roster (`discovery_checks_registration_test.go:69-82`) includes `tool_health`,
  `tool_acceptance`, `tool_acceptance_due`, `palette_contrast`, `image_url_404` — none of it
  runs anywhere.
- What the watchdog said meanwhile, quoted from `doc_notes`
  (`subject_key='site-discovery-staleness'`) [MEASURED 2026-08-25]:
  - 08-11 → 08-12: `1/3`, findings: 1 (`driver_missing` — **correct**, and it early-returns, so
    it is the whole report);
  - 08-13 → 08-17: `2/3`, findings: 1 (`driver_missing`, still correct);
  - **08-18 onward: `3/3`** — the availability task's enable made the count whole — and
    `driver_missing` fell silent with design still off. Findings 08-18 → 08-23: 1, 3, 3, 3, 4, 5
    (early `site_stale` rows only);
  - 08-24: **26** and 08-25: **27** `site_stale` findings (mass crossing of the 14-day
    threshold, stamps dated 08-09/11), every one naming `design-discovery-agent`.
- So the failure the alarm exists for was reported as itself for 7 days, then **re-described**
  as 26–27 per-site staleness rows — which read as a rotation-fairness problem, not as "one
  switch is off". Nothing consumes the report either way (the standing
  detection-works/dispatch-doesn't shape), but the report a human might eventually read now
  actively says `3/3` on its second line.

## Who this bit (why it was found today)

The `webdesign_tool_rebuilds` Track 2 demand controls — `capability_gap` filing by
`tool_health` contract rules 16/17, and `tool_acceptance` no longer enumerating tombstones,
live on chassis v1.0.1339 since 2026-08-25 19:07Z — are checks **carried only by
design-discovery-agent**, so they are unreachable until the rotation is re-enabled or a run is
hand-fired. The seat's own handoff said "the sweeps are scheduled", read `3/3`-shaped evidence,
and was wrong (logged in `WRONG_CALLS.md` 2026-08-25). Also downstream of the same off switch:
`bugs_closed/302`'s dark_section retraction path (their notes already record the carrier as
dead), and every `tool_acceptance` re-audit of the sibling lane's 43 rebuilt tools.

## Fix candidates, ordered by what closes the door

1. **Compare a name SET, not a count** (makes the masked state unrepresentable): expected =
   the three `site-discovery-rotation-{quality,design,completeness}` names derived from
   `DISCOVERY_AGENTS`; a `driver_missing` finding names each expected task that is missing OR
   disabled. A fifth pattern-matching task then cannot vouch for a dead one. Replace the
   `N/3` header line with the per-task enabled map (`quality:t design:f completeness:t`),
   which cannot summarise two truths into one number.
2. Weaker but nonzero: keep the count, filter `tasks` to the three expected names. Closes this
   door, leaves the header line able to lie by omission if a task is ever renamed.
3. Not a candidate: teaching the count about availability (`4/4`). The 236 lane explicitly
   left availability watchdog coverage as the 230 lane's call; whether to watch it is a
   separate decision from not letting it mask the others.

Secondary observation, same script, separate and minor: on findings days the job exits 1
(`check.py:263`) and the Job's `backoffLimit: 1` retries it, so **every findings day writes two
identical `doc_notes` rows ~19s apart** (visible in the whole series above; clean 08-10 wrote
one). Harmless noise today; worth a `backoffLimit: 0` when the script is next touched, since a
findings exit is not a retryable failure.

## How to verify a fix

Induce the masked state in reverse: with design still disabled, the fixed check must emit
`driver_missing` naming `site-discovery-rotation-design` (not 27 `site_stale` rows as the
lead finding), and the header must show per-task states. Then flip design on and confirm a
clean driver line. The check is pure SQL + Python against live rows — runnable as a one-off
Job or locally against the DB before touching the CronJob.

## Ownership

The watchdog belongs to the `bugfix_230_discovery_driver` lane (quiet ≥14d per `who-owns`,
2026-08-25). A CONTRIB pointing here is filed in their directory. **Whether to re-enable the
design rotation itself is the OWNER's staged-ramp decision** (395's foot), not part of this
bug: this file is about the alarm that should have kept that decision visible.

---

**UPDATE 2026-08-26 ~09:25Z:** the masked CASE ended — the owner ruled and
`site-discovery-rotation-design` was re-enabled at 09:20:04Z (10800s), proven firing end-to-end
within 40 seconds (trigger 09:20:36Z, agritec.uk stamped and swept, 24 checks, 11 findings;
evidence chain in `webdesign_tool_rebuilds/NOTES_…` 2026-08-26). **This bug stays OPEN**: the
defect is the watchdog's count, which is unchanged — it now reads a truthful "3/3"-shaped state
by coincidence, and the next single-rotation disable will be masked exactly as this one was.
Today's staleness reports will also drain from 27 `site_stale` rows toward 0 over the ~3-day
ramp; treat that drain as the rotation working, not as this bug being fixed.
